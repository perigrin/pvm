// ABOUTME: The thin chart/lex: lexer.Tokenize up front, trivia to the next significant token, one table lookup per token.
// ABOUTME: Heredoc bodies go to their openers; a token no class admits, or an opener with no body, gets the empty set and the first such byte is kept.

// Package lex is the chart's lexer (spec §5.2, "A thin chart/lex"): the
// production lexer runs up front with its own expect machine, and each
// significant token is mapped to a terminal-set id through the compiled
// token table. Nothing grammar-specific is decided here.
package lex

import (
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/lexer"
)

// Lex is Map over lexer.Tokenize's tokens of src.
func Lex(src []byte, tt *grammar.TokenTable) (*Stream, int) {
	return Map(src, lexer.Tokenize(src), tt)
}

// Map maps raw, src's tokens, through tt. It returns the stream, which
// shares raw rather than copying it, and the byte of the first token no class
// admits, or -1. Trivia attaches to the next significant token, and each
// heredoc body to the waiting opener whose tag its terminator line spells.
func Map(src []byte, raw []lexer.Token, tt *grammar.TokenTable) (*Stream, int) {
	// text views src without copying it; nothing Map returns holds a string, so
	// the view lives only as long as this call, during which src is not written.
	text := unsafe.String(unsafe.SliceData(src), len(src))
	// Tokens is sized to the significant tokens rather than to raw, a third of which is trivia.
	sig := 0
	for _, t := range raw {
		if !lexer.IsTrivia(t.Kind) && t.Kind != lexer.HeredocBody {
			sig++
		}
	}
	s := &Stream{Raw: raw, Tokens: make([]Token, 0, sig)}
	first := -1
	// slots holds the last token and the one being interned, alternately, so
	// neither is copied (see grammar.TokenTable.Intern).
	var slots [2]grammar.Token
	pred, p2, p1 := &grammar.Boundary, tt.Edge, tt.Edge
	c := caches.Get().(*grammar.LookupCache)
	defer caches.Put(c)
	// done looks up the last token, now that its next neighbour's class is known.
	done := func(next int) {
		n := len(s.Tokens) - 1
		last, cur := &s.Tokens[n], &slots[n&1]
		last.TerminalSet = c.Lookup(tt, cur, p2, p1, next)
		if last.TerminalSet == 0 && first < 0 {
			first = raw[last.Index].Start
		}
		p2, p1 = p1, cur.Class
	}
	var opens []int // heredoc openers whose bodies have not come yet
	// on is the features in force: toke.c reads isa as a keyword only with its feature on (keywords.c:353). A use or
	// no statement sets them (featuresAfter), and a BEGIN block may (mayBeOn), for the rest of its block, and the
	// block's } restores them from scopes; each change is recorded for resolve (Stream.Features).
	var on Features
	scopes := make([]Features, 0, 16)
	var st stashes
	for i, t := range raw {
		switch {
		case lexer.IsTrivia(t.Kind):
			continue
		case t.Kind == lexer.HeredocBody && len(s.Tokens) > 0:
			// A body goes to the first waiting opener whose tag its terminator line spells. Any other body's opener is
			// inside a quote-like (`"@{[ <<EOF ]}"`), one token to the lexer.
			// shortcut: that body goes to the token before it, not to the quote-like holding its opener, which only
			// Stream.Body and the tdt pairing read; upgrade when the lexer marks the quote-likes that open heredocs
			// (v0.1 01a1af25-7fc7-7e37-9fd8-ebf632fd461c).
			o, end := len(s.Tokens)-1, terminator(text[t.Start:t.End])
			for j, w := range opens {
				if r := raw[s.Tokens[w].Index]; tag(text[r.Start:r.End]) == end {
					o, opens = w, append(opens[:j], opens[j+1:]...)
					break
				}
			}
			s.Bodies = append(s.Bodies, Body{o, t})
			continue
		}
		if st.typedClass(t, text, on) {
			continue // toke.c consumes a typed my's class name (KEY_my's PL_in_my_stash) and emits no token for it
		}
		st.note(raw, i, text, on, pred)
		plainIsa, was := false, on
		switch n := t.End - t.Start; t.Kind {
		case lexer.Operator:
			if n == 1 && text[t.Start] == '{' {
				scopes = append(scopes, on)
			}
		case lexer.CloseBracket:
			if text[t.Start] == '}' && len(scopes) > 0 {
				on, scopes = scopes[len(scopes)-1], scopes[:len(scopes)-1]
			}
		case lexer.Word:
			switch w := text[t.Start:t.End]; {
			case n == 3 && w == "use" || n == 2 && w == "no":
				on = featuresAfter(raw, i, text, on)
			case n == 5 && w == "BEGIN" && beginBlock(raw, i, text, s):
				on = on.mayBeOn() // `BEGIN { require feature; feature->import('isa') }`
			default:
				plainIsa = on.On&FeatureIsa == 0 && n == 3 && w == "isa"
			}
		}
		s.change(was, on)
		it := &slots[len(s.Tokens)&1]
		intern(tt, text, t, lineColon(raw, i, text), blanks(raw, i, text), plainIsa, pred, it)
		if len(s.Tokens) > 0 {
			done(it.Class)
		}
		if t.Kind == lexer.HeredocOpen {
			opens = append(opens, len(s.Tokens))
		}
		s.Tokens = append(s.Tokens, Token{Index: i})
		pred = &it.N
	}
	if len(s.Tokens) > 0 {
		done(tt.Edge)
	}
	// An opener whose body never came is a heredoc perl cannot find the terminator of: no class admits it. toke.c croaks
	// there and internal/lexer does not (v0.1 01a1af7c-3b8b), so its files are on lex-unmapped-known.txt.
	for _, o := range opens {
		s.Tokens[o].TerminalSet = 0
		if b := raw[s.Tokens[o].Index].Start; first < 0 || b < first {
			first = b
		}
	}
	return s, first
}

// stashes are the packages a file has made so far, as far as its tokens show, and the significant word before the
// current token: what a typed my's class name needs (typedClass).
type stashes struct {
	made map[string]bool
	prev string // the last significant token's text if it was a Word, else ""
	// brace is whether the last significant token was a `{` right after a sigil or a cast (`$`, `@`, `%`, `&` or a
	// glob's `*`), which opens a name where a word and its `}` are all it holds (`@{Foo::x}`, `&{Foo::bar}`).
	brace bool
}

// startStashes are the stashes the pinned perl (94e50866) starts with, measured by `my X $f;` under perl -c.
var startStashes = map[string]bool{"main": true, "UNIVERSAL": true, "CORE": true, "CORE::GLOBAL": true,
	"Internals": true, "utf8": true, "re": true, "version": true, "PerlIO": true, "PerlIO::Layer": true, "Regexp": true,
	"DB": true, "DynaLoader": true, "Exporter": true, "constant": true, "IO": true, "IO::File": true,
	"IO::Seekable": true, "IO::Handle": true, "Tie": true, "Tie::Hash": true, "Tie::Hash::NamedCapture": true,
	"mro": true, "builtin": true}

// typedClass is whether t is a typed my's class name: a word after my, our or state (its feature maybe on) other than
// sub or method, naming a stash that exists. toke.c croaks "No such class" for any other, so that word is kept and,
// with no reading the grammar takes there, refuses.
// shortcut: a stash made other than by package, class, use, no, require or a symbol's qualified name in code before it
// (a string eval, an interpolated "$Foo::x", an s///e replacement, a print's filehandle `print Foo::x "a"`, a module
// that makes another package) is not seen, so that typed my is refused; upgrade when a corpus shows one.
func (st *stashes) typedClass(t lexer.Token, text string, on Features) bool {
	if t.Kind != lexer.Word || st.prev != "my" && st.prev != "our" && (st.prev != "state" || (on.On|on.Maybe)&FeatureState == 0) {
		return false
	}
	w := text[t.Start:t.End]
	if strings.HasPrefix(w, "::") {
		return false // perl reads `my ::Foo` as a syntax error
	}
	w = stashName(strings.TrimSuffix(w, "::")) // `my Foo::` is Foo, `my main::::` the stash named "" (main::::)
	return w != "sub" && w != "method" && (w == "__PACKAGE__" || startStashes[w] || st.made[w])
}

// versionEnds reports whether a v-string whose word ends at end is a version, as toke.c's S_force_version decides:
// it walks digits, _ and . past the word (`v5_1. 1` is the version `v5_1.`), then wants an isSPACE_A byte, ;, {, }
// or the end of the source. Any other byte (`require v5_1, 1`) leaves it a module name.
func versionEnds(text string, end int) bool {
	for end < len(text) && (text[end] >= '0' && text[end] <= '9' || text[end] == '_' || text[end] == '.') {
		end++
	}
	return end >= len(text) || strings.IndexByte(" \t\n\v\f\r;{}", text[end]) >= 0
}

// stashName is the stash a qualified name n names, as perl reads it: ' is ::, and one leading :: and then any number
// of main:: name nothing more (`::main::main::Foo'Bar` is Foo::Bar, `main::::Foo` is ::Foo, inside the stash named
// "", and `Foo::` is a stash inside Foo).
func stashName(n string) string {
	n = strings.TrimPrefix(strings.ReplaceAll(n, "'", "::"), "::")
	for strings.HasPrefix(n, "main::") {
		n = n[6:]
	}
	return n
}

// add records stash n and each stash it is nested in.
func (st *stashes) add(n string) {
	if st.made == nil {
		st.made = map[string]bool{}
	}
	for k := 0; k <= len(n); k++ {
		if k == len(n) || strings.HasPrefix(n[k:], "::") {
			st.made[n[:k]] = true // "" too: main:: is in no stash, so only main::::'s parts name it
		}
	}
}

// note records the stashes raw[i] makes, each with every stash it is nested in: the name after package, class (its
// feature on), use, no or require, and each
// package a qualified name passes through where perl makes a symbol of it -- a variable (`$Foo::x`, `$ Foo::x`), a
// word after a sigil (`&Foo::bar`, a glob's `*Foo::x`), a word alone in the braces after a sigil or cast
// (`@{Foo::x}`, `*{ Foo::x }`), a word called with ( (`Foo::bar()`), a word before -> that no method name follows
// (`Foo::bar->[0]`, `Foo::bar->@*`), or a sub's or format's name (`sub Foo::bar`, `format Foo::x =`). A bareword, class or method name
// (`Foo::bar;`, `Foo::Bar->x`, `Foo::bar->$m`, `$x->Foo::bar()`) makes none. Neighbours are significant tokens, so
// whitespace, comments and POD between them do not count. pred is the class of the token before raw[i].
func (st *stashes) note(raw []lexer.Token, i int, text string, on Features, pred *grammar.Neighbour) {
	t := raw[i]
	tx := text[t.Start:t.End]
	glob := pred.Kind == lexer.Operator && pred.Text == "*" && pred.Glob // the grammar's glob-sigil test
	sigil := glob || pred.Kind == lexer.FuncSigil || pred.Kind == lexer.DerefSigil
	prev, brace := st.prev, st.brace
	st.prev, st.brace = "", sigil && t.Kind == lexer.Operator && tx == "{"
	if t.Kind == lexer.Word {
		st.prev = strings.TrimPrefix(tx, "CORE::") // toke.c reads CORE::kw as kw
	}
	if t.Kind != lexer.Word && t.Kind != lexer.Variable {
		return
	}
	if t.Kind == lexer.Word && (prev == "package" || prev == "use" || prev == "no" || prev == "require" ||
		prev == "class" && on.On&FeatureClass != 0) &&
		!(prev != "package" && prev != "class" && len(tx) > 1 && tx[0] == 'v' && tx[1] >= '0' && tx[1] <= '9' &&
			strings.Trim(tx[1:], "0123456789_") == "" && versionEnds(text, t.End)) {
		// a v-string after use, no or require is a version, not a module
		st.add(stashName(tx))
	}
	if !strings.Contains(tx, "::") && !strings.Contains(tx, "'") {
		return
	}
	if t.Kind == lexer.Word && !(sigil || prev == "sub" || prev == "format") {
		n := significant(raw, i)
		switch {
		case pred.Kind == lexer.Operator && pred.Text == "->":
			return // a method's name
		case brace && tx[0] != ':' && is(raw, n, text, lexer.CloseBracket, "}"):
			// scan_ident reads `{Foo::x}` as the name; `{::Foo::x}` is a block holding a string
		case is(raw, n, text, lexer.Operator, "("):
		case is(raw, n, text, lexer.Operator, "->"):
			if m := significant(raw, n); m < len(raw) && (raw[m].Kind == lexer.Word ||
				raw[m].Kind == lexer.Variable && text[raw[m].Start] == '$' && text[raw[m].End-1] != '*') {
				return // a class name: a method (`->m`, `->$m`, `->${...}`) follows, where a deref (`->$*`) does not
			}
		default:
			return // a bareword
		}
	}
	// the symbol's stash is all of its name before the last ::
	name := strings.ReplaceAll(strings.TrimRight(strings.TrimLeft(tx, "$@%&*#{ \t\n"), "} \t\n"), "'", "::")
	if k := strings.LastIndex(name, "::"); k > 0 { // `$::x` is main's
		st.add(stashName(name[:k]))
	}
}

// significant is the index of the first significant token after raw[i], len(raw) if none.
func significant(raw []lexer.Token, i int) int {
	for i++; i < len(raw) && (lexer.IsTrivia(raw[i].Kind) || raw[i].Kind == lexer.HeredocBody); i++ {
	}
	return i
}

// is is whether raw[n] exists and is a token of kind k spelling s.
func is(raw []lexer.Token, n int, text string, k lexer.Kind, s string) bool {
	return n < len(raw) && raw[n].Kind == k && text[raw[n].Start:raw[n].End] == s
}

// caches lends each Map call its own LookupCache, so concurrent calls never share one, and a cache keeps its
// answers from one file to the next.
var caches = sync.Pool{New: func() any { return new(grammar.LookupCache) }}

// tag is a heredoc opener's terminator without an indent: `<<"EOF"`, `<<~ EOF`, `<<\EOF` and `<<'  EOF'` are all EOF.
func tag(opener string) string {
	return strings.TrimLeft(strings.Trim(strings.TrimLeft(opener, "<~ \t\\"), "\"'`"), " \t")
}

// terminator is a heredoc body's last line, the one that ended it, without its line break or an indent.
func terminator(body string) string {
	body = strings.TrimSuffix(strings.TrimSuffix(body, "\n"), "\r")
	return strings.TrimLeft(body[strings.LastIndexByte(body, '\n')+1:], " \t")
}

// intern interns t's key in tt into it, given its predecessor's class; plainIsa says t is isa with its feature off.
func intern(tt *grammar.TokenTable, text string, t lexer.Token, lc, bl, plainIsa bool, pred *grammar.Neighbour, it *grammar.Token) {
	tx := text[t.Start:t.End]
	if plainIsa {
		tx = grammar.PlainIsa
	}
	if t.LexSub && lexer.IsKeyword(tx) {
		tx = grammar.PlainWord
	}
	var next byte // 0 at the end of the file, which ends a version as toke.c's NUL does
	if t.Kind == lexer.Number && t.End < len(text) {
		next = text[t.End]
	}
	// v is written field by field: a literal is built aside and copied with wide loads, which stall on its narrow stores.
	p, g := prefix(t.Kind, tx, next), (lc || t.Kind == lexer.Variable) && global(t.Kind, tx)
	if t.Kind == lexer.Number && pred.Kind == lexer.Operator && pred.Text == "*" && globDigits(tx) {
		p = grammar.GlobDigits // a glob's name if the * is a glob's sigil, which the table decides
	}
	if p == grammar.CoreQualified {
		tx = tx[6:] // toke.c's yyl_key_core reads CORE::kw as kw; decided before v is written, which reading v back would stall
	}
	var v grammar.TokenView
	v.Kind, v.Text, v.OpensBlock, v.TermFollows, v.Prefix = t.Kind, tx, t.OpensBlock, t.TermFollows, p
	v.Global, v.LineColon, v.Blanks = g, lc, bl
	tt.Intern(&v, pred, it)
}

// Feature is a set of the features feature.pm names that the chart reads, one bit each.
type Feature uint16

// The features the chart reads: isa here, the rest in resolve.
const (
	FeatureIsa Feature = 1 << iota
	FeatureSignatures
	FeatureTry
	FeatureDefer
	FeatureClass
	FeatureState
	FeatureSwitch
	FeatureCurrentSub
	FeatureFc
	FeatureEvalbytes
	FeatureKeywordAny
	FeatureKeywordAll
	featureAll = FeatureKeywordAll<<1 - 1
)

// Features is the features known on at a byte, and those an import or a BEGIN block may have turned on there,
// which no statement the chart reads has turned on or off since.
type Features struct{ On, Maybe Feature }

// mayBeOn is f after an import the chart does not run: each feature known off may be on, as internal/parse's
// featuresMayBeOn reads it (`use if 1, feature => 'try'`, `use Mojo::Base -signatures`).
func (f Features) mayBeOn() Features {
	f.Maybe |= featureAll &^ f.On
	return f
}

// quietPragmas are the core pragmas whose import turns no feature on: internal/parse's list of them, and constant
// and subs, which it leaves out only because they declare subs.
var quietPragmas = map[string]bool{"strict": true, "warnings": true, "utf8": true, "open": true, "vars": true,
	"lib": true, "integer": true, "bytes": true, "less": true, "sort": true, "overload": true, "parent": true,
	"base": true, "mro": true, "locale": true, "constant": true, "subs": true}

// featureNames are the features by the names feature.pm and experimental.pm take.
var featureNames = map[string]Feature{"isa": FeatureIsa, "signatures": FeatureSignatures, "try": FeatureTry,
	"defer": FeatureDefer, "class": FeatureClass, "state": FeatureState, "switch": FeatureSwitch,
	"current_sub": FeatureCurrentSub, "fc": FeatureFc, "evalbytes": FeatureEvalbytes, "keyword_any": FeatureKeywordAny,
	"keyword_all": FeatureKeywordAll, ":all": featureAll}

// bundle is the features of feature.pm's bundle for version v (minor*1000+patch, as version reads it), as use VERSION
// and :5.m name it; the :5.10 bundle applies from 5.9.5, as feature.pm's `$feature_bundle{"5.9.5"}` does.
func bundle(v int) Feature {
	m := v / 1000
	var f Feature
	if v >= 9005 {
		f |= FeatureState
	}
	if v >= 9005 && m < 35 {
		f |= FeatureSwitch
	}
	if m >= 15 {
		f |= FeatureCurrentSub | FeatureFc | FeatureEvalbytes
	}
	if m >= 35 {
		f |= FeatureIsa | FeatureSignatures
	}
	if m >= 39 {
		f |= FeatureTry
	}
	return f
}

// featuresAfter is the features in force after the use or no statement at raw[i], given on before it, as feature.pm
// and experimental.pm set them: a feature or experimental statement turns on (use) or off (no) each feature it names,
// and each of a bundle's (:all, :5.36); a bare `no feature` returns to the default bundle, which holds none of them;
// and use VERSION replaces them with its version's bundle. A list not all quoted names, or a use of a module that is
// no quiet pragma, may turn any on (mayBeOn).
// shortcut: a feature turned on in a string eval is not seen; upgrade if a corpus file needs one.
func featuresAfter(raw []lexer.Token, i int, text string, on Features) Features {
	use := text[raw[i].Start] == 'u'
	j := i + 1
	for j < len(raw) && lexer.IsTrivia(raw[j].Kind) {
		j++
	}
	if j == len(raw) {
		return on
	}
	w := text[raw[j].Start:raw[j].End]
	switch k := raw[j].Kind; {
	case k == lexer.Word && (w == "feature" || w == "experimental"):
		named := false
		for j++; j < len(raw) && raw[j].Kind != lexer.Semicolon; j++ {
			switch k := raw[j].Kind; {
			case lexer.IsTrivia(k) || k == lexer.Operator && (text[raw[j].Start:raw[j].End] == "," || text[raw[j].Start:raw[j].End] == "=>"):
				continue
			case k != lexer.Quote:
				return on.mayBeOn() // `use feature @f`
			}
			for _, f := range strings.FieldsFunc(quoted(text[raw[j].Start:raw[j].End]), notName) {
				named = true
				set := featureNames[f]
				if strings.HasPrefix(f, ":5.") {
					set = bundle(version("v" + f[1:]))
				}
				if use {
					on.On |= set
				} else {
					on.On &^= set
				}
				on.Maybe &^= set
			}
		}
		if !named && !use && w == "feature" {
			return Features{} // `no feature` returns to the default bundle
		}
	case use && (k == lexer.Number || k == lexer.Quote && (w[0] == 'v' || '0' <= w[0] && w[0] <= '9')):
		on = Features{On: bundle(version(w))}
	case use && k == lexer.Word && !quietPragmas[w]:
		on = on.mayBeOn()
	}
	return on
}

// beginBlock is whether the BEGIN at raw[i] opens a block: a { follows it, and no -> is before it (s's last token).
func beginBlock(raw []lexer.Token, i int, text string, s *Stream) bool {
	if n := len(s.Tokens); n > 0 {
		if p := raw[s.Tokens[n-1].Index]; text[p.Start:p.End] == "->" {
			return false
		}
	}
	for j := i + 1; j < len(raw); j++ {
		if !lexer.IsTrivia(raw[j].Kind) {
			return text[raw[j].Start:raw[j].End] == "{"
		}
	}
	return false
}

// quoted is a quote-like's contents: its text without the operator's letters and the delimiters.
func quoted(s string) string {
	s = strings.TrimLeft(s, "qw")
	if len(s) < 2 {
		return ""
	}
	return s[1 : len(s)-1]
}

func notName(r rune) bool {
	return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '(' || r == ')'
}

// version is a Perl 5 version (`v5.36`, `5.36.0`, `5.036`, `5.036_000`) as minor*1000+patch, or -1; a bundle's (`:5.36`) reads as `v5.36`.
// A decimal version's minor is the first three digits after its point and its patch the next three; a dotted one's, the
// second and third parts.
func version(v string) int {
	dotted := strings.HasPrefix(v, "v") || strings.Count(v, ".") > 1
	parts := strings.Split(strings.ReplaceAll(strings.TrimPrefix(v, "v"), "_", ""), ".")
	if len(parts) < 2 || parts[0] != "5" {
		return -1
	}
	m, p := parts[1], "0"
	if dotted && len(parts) > 2 {
		p = parts[2]
	} else if !dotted {
		m += "000000"
		m, p = m[:3], m[3:6]
	}
	n, err := strconv.Atoi(m)
	q, perr := strconv.Atoi(p)
	if err != nil || perr != nil {
		return -1
	}
	return n*1000 + q
}

// keywords are the words CORE:: may qualify, looked up only for a word that opens with CORE::.
var keywords = func() map[string]bool {
	m := map[string]bool{}
	for _, k := range lexer.Keywords() {
		m[k] = true
	}
	return m
}()

// global is whether a word, or a `$` variable, is no plain name (a variable's name as varName gives it): it has a package separator or another
// character an identifier lacks (`$1`, `$^W`, `$'`), or it is the variable `$_`. A `my` declares only plain names.
func global(k lexer.Kind, tx string) bool {
	switch k {
	case lexer.Word:
		return !plainName(tx)
	case lexer.Variable:
		if len(tx) < 2 || tx[0] != '$' {
			return false
		}
		n := varName(tx[1:])
		return n == "_" || !plainName(n)
	}
	return false
}

// varName is a variable's name without the whitespace and comments after its sigil and without braces, with the
// whitespace inside them: the lexer keeps `${ e }` and `$ e` as one token, which perl reads as `$e`.
func varName(s string) string {
	const blank = " \t\n\r\f\v"
	s = skipBlank(s)
	if len(s) > 1 && s[0] == '{' && s[len(s)-1] == '}' {
		s = strings.TrimRight(skipBlank(s[1:len(s)-1]), blank)
	}
	return s
}

// skipBlank is s without the blanks and the `#...` comments, each to its newline, that open it.
func skipBlank(s string) string {
	for {
		for s != "" && (s[0] == ' ' || s[0]-'\t' <= '\r'-'\t') { // strings.TrimLeft's cutset scan cost more than the name
			s = s[1:]
		}
		i := strings.IndexByte(s, '\n')
		if s == "" || s[0] != '#' || i < 0 {
			return s
		}
		s = s[i+1:]
	}
}

// plainName is whether s is a letter or _ (or a UTF-8 byte) and then those and digits: an identifier.
func plainName(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b != '_' && b|0x20-'a' >= 26 && b < 0x80 && (i == 0 || b-'0' >= 10) {
			return false
		}
	}
	return s != ""
}

// lineColon is whether raw[i], a word, is followed by a one-byte colon with at most one whitespace token,
// holding no newline, between: toke.c labels a word only when the colon is on its line.
func lineColon(raw []lexer.Token, i int, src string) bool {
	if raw[i].Kind != lexer.Word {
		return false
	}
	j := i + 1
	if j < len(raw) && raw[j].Kind == lexer.Whitespace && !strings.Contains(src[raw[j].Start:raw[j].End], "\n") {
		j++
	}
	return j < len(raw) && raw[j].Kind == lexer.Operator && raw[j].End-raw[j].Start == 1 && src[raw[j].Start] == ':'
}

// blanks is whether raw[i], a word, is a plain name that a closing bracket follows, and that only spaces or tabs
// separate from the token on either side and, when that is a minus, from the one before it: where toke.c's
// yyl_leftcurly (toke.c:6656-6679) forces a word to a string. It reads the gaps between tokens, not the bytes.
func blanks(raw []lexer.Token, i int, src string) bool {
	if raw[i].Kind != lexer.Word {
		return false
	}
	// Most words fail here, before any gap is read.
	k := i + 1
	if k < len(raw) && raw[k].Kind == lexer.Whitespace {
		k++
	}
	if k >= len(raw) || raw[k].Kind != lexer.CloseBracket || skipBlanks(raw, i+1, 1, src) < 0 {
		return false
	}
	j := skipBlanks(raw, i-1, -1, src)
	if j >= 0 && raw[j].Kind == lexer.Operator && src[raw[j].Start:raw[j].End] == "-" {
		j = skipBlanks(raw, j-1, -1, src)
	}
	return j >= -1 && plainName(src[raw[i].Start:raw[i].End])
}

// skipBlanks is the index of the token past raw[j], a whitespace token of spaces and tabs, stepping by step, or j
// itself when raw[j] is not one; -2 when the token reached is other trivia (a newline, a comment), and -1 when the file's edge is.
func skipBlanks(raw []lexer.Token, j, step int, src string) int {
	if j >= 0 && j < len(raw) && raw[j].Kind == lexer.Whitespace && strings.Trim(src[raw[j].Start:raw[j].End], " \t") == "" {
		j += step
	}
	if j >= 0 && j < len(raw) && lexer.IsTrivia(raw[j].Kind) {
		return -2
	}
	return j
}

// globDigits is whether tx is what scan_ident reads as a glob's name after `*`: decimal digits, no leading 0 but 0.
func globDigits(tx string) bool {
	if tx == "" || len(tx) > 1 && tx[0] == '0' {
		return false
	}
	for i := 0; i < len(tx); i++ {
		if tx[i] < '0' || tx[i] > '9' {
			return false
		}
	}
	return true
}

// versionShape is a Number's version class for the key, which drops its text: grammar.VersionStrict
// for one toke.c's force_strict_version reads as a version (0 or a digit run without a leading zero,
// then an optional . and digits), grammar.VersionLax for the rest of what force_version reads (a
// digit, then digits, _ and .), and 0 for any other. Both need the byte after the number, next, to be
// whitespace, ;, {, } or the end of the file: a number that something else follows is a term to toke.c,
// whatever its digits (`use lib 1.2, 3`).
func versionShape(tx string, next byte) byte {
	if next != 0 && next != ';' && next != '{' && next != '}' && next != ' ' && (next < '\t' || next > '\r') {
		return 0
	}
	if tx == "" || tx[0] < '0' || tx[0] > '9' {
		return 0
	}
	strict, dots := true, 0
	for i := 0; i < len(tx); i++ {
		switch c := tx[i]; {
		case c >= '0' && c <= '9':
		case c == '.':
			dots++
			strict = strict && i+1 < len(tx) && dots == 1
		case c == '_':
			strict = false
		default:
			return 0
		}
	}
	// a leading zero is strict only alone before the point
	if strict && tx[0] == '0' && len(tx) > 1 && tx[1] != '.' {
		strict = false
	}
	if strict {
		return grammar.VersionStrict
	}
	return grammar.VersionLax
}

// prefix is a variable's sigil (`#` for `$#`), a quote's class (its opening
// quote, or its operator: `Q` for qq, `y` for tr), a number's versionShape or grammar.PackageName for a word
// ending in ::; 0 for anything else.
func prefix(k lexer.Kind, tx string, next byte) byte {
	switch {
	case k == lexer.Number:
		return versionShape(tx, next)
	case k == lexer.Variable && len(tx) > 2 && tx[:2] == "$#":
		return '#'
	case k == lexer.Variable && tx != "":
		return tx[0]
	case k == lexer.Word && len(tx) > 2 && tx[len(tx)-1] == ':' && tx[len(tx)-2] == ':':
		return grammar.PackageName
	case k == lexer.Word && tx[0] == 'C' && len(tx) > 6 && tx[:6] == "CORE::" && keywords[tx[6:]]:
		return grammar.CoreQualified
	case k != lexer.Quote || tx == "":
		return 0
	}
	n := 0
	for n < len(tx) && 'a' <= tx[n] && tx[n] <= 'z' {
		n++
	}
	switch tx[:n] {
	case "":
		return tx[0]
	case "qq":
		return 'Q'
	case "qw":
		return 'w'
	case "qr":
		return 'r'
	case "qx":
		return 'x'
	case "tr":
		return 'y'
	}
	return tx[0] // q, m, s, y
}
