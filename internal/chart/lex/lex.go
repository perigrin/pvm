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
	s := &Stream{Raw: raw, Tokens: make([]Token, 0, len(raw))}
	first := -1
	// slots holds the last token and the one being interned, alternately, so
	// neither is copied (see grammar.TokenTable.Intern).
	var slots [2]grammar.Token
	pred, p2, p1 := &grammar.Boundary, tt.Edge, tt.Edge
	// done looks up the last token, now that its next neighbour's class is known.
	done := func(next int) {
		n := len(s.Tokens) - 1
		last, cur := &s.Tokens[n], &slots[n&1]
		last.TerminalSet = tt.Lookup(*cur, p2, p1, next)
		if last.TerminalSet == 0 && first < 0 {
			first = raw[last.Index].Start
		}
		p2, p1 = p1, cur.Class
	}
	var opens []int // heredoc openers whose bodies have not come yet
	// isa is whether the isa feature is on: toke.c reads isa as a keyword only then (keywords.c:353). A use or no
	// statement sets it (isaAfter) for the rest of its block, and the block's } restores it from scopes.
	var isa bool
	scopes := make([]bool, 0, 16)
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
		plainIsa := false
		switch n := t.End - t.Start; t.Kind {
		case lexer.Operator:
			if n == 1 && text[t.Start] == '{' {
				scopes = append(scopes, isa)
			}
		case lexer.CloseBracket:
			if text[t.Start] == '}' && len(scopes) > 0 {
				isa, scopes = scopes[len(scopes)-1], scopes[:len(scopes)-1]
			}
		case lexer.Word:
			if w := text[t.Start:t.End]; n == 3 && w == "use" || n == 2 && w == "no" {
				isa = isaAfter(raw, i, text, isa)
			} else {
				plainIsa = !isa && n == 3 && w == "isa"
			}
		}
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
	var next byte // 0 at the end of the file, which ends a version as toke.c's NUL does
	if t.Kind == lexer.Number && t.End < len(text) {
		next = text[t.End]
	}
	// v is written field by field: a literal is built aside and copied with wide loads, which stall on its narrow stores.
	p, g := prefix(t.Kind, tx, next), (lc || t.Kind == lexer.Variable) && global(t.Kind, tx)
	if p == grammar.CoreQualified {
		tx = tx[6:] // toke.c's yyl_key_core reads CORE::kw as kw; decided before v is written, which reading v back would stall
	}
	var v grammar.TokenView
	v.Kind, v.Text, v.OpensBlock, v.TermFollows, v.Prefix = t.Kind, tx, t.OpensBlock, t.TermFollows, p
	v.Global, v.LineColon, v.Blanks = g, lc, bl
	tt.Intern(&v, pred, it)
}

// isaAfter is the isa feature's state after the use or no statement at raw[i], given on before it, as feature.pm and
// experimental.pm set it: a feature or experimental statement naming isa, :all or a bundle from :5.35 turns it on (use)
// or off (no), and a bare `no feature` turns it off; use VERSION turns it on from 5.35 and off below it, as the
// version's bundle replaces the features named before it.
// shortcut: a feature a module or a BEGIN block imports (`BEGIN { feature->import('isa') }`) is not seen, nor one
// in a string eval; upgrade if a corpus file needs one.
func isaAfter(raw []lexer.Token, i int, text string, on bool) bool {
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
			if raw[j].Kind != lexer.Quote {
				continue
			}
			for _, f := range strings.FieldsFunc(quoted(text[raw[j].Start:raw[j].End]), notName) {
				named = true
				if f == "isa" || f == ":all" || strings.HasPrefix(f, ":5.") && minor("v"+f[1:]) >= 35 {
					on = use
				}
			}
		}
		if !named && !use && w == "feature" {
			return false // `no feature` returns to the default bundle, which has no isa
		}
	case use && (k == lexer.Number || k == lexer.Quote && (w[0] == 'v' || '0' <= w[0] && w[0] <= '9')):
		if m := minor(w); m >= 35 {
			on = true
		} else {
			on = false
		}
	}
	return on
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

// minor is the minor version of a Perl 5 version (`v5.36`, `5.36.0`, `5.036`, `5.036_000`), or -1; a bundle's (`:5.36`) reads as `v5.36`.
// A decimal version's minor is the first three digits after its point; a dotted one's, the second part.
func minor(v string) int {
	dotted := strings.HasPrefix(v, "v") || strings.Count(v, ".") > 1
	parts := strings.Split(strings.ReplaceAll(strings.TrimPrefix(v, "v"), "_", ""), ".")
	if len(parts) < 2 || parts[0] != "5" {
		return -1
	}
	m := parts[1]
	if !dotted && len(m) < 3 {
		m += strings.Repeat("0", 3-len(m))
	} else if !dotted {
		m = m[:3]
	}
	n, err := strconv.Atoi(m)
	if err != nil {
		return -1
	}
	return n
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
