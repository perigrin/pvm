// ABOUTME: Holds every corpus token's terminal-set id to a reference that matches each token class directly.
// ABOUTME: Benchmarks Intern against the committed token table per significant token of std5k.pl, beside lexer.Tokenize.

package lex

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unsafe"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/lexer"
)

// committed is the token table the generator builds from perly.go, and the
// classes it was compiled from.
func committed(tb testing.TB) (*grammar.TokenTable, []grammar.TokenClass) {
	tb.Helper()
	g, ref, err := grammar.Committed()
	if err != nil {
		tb.Fatal(err)
	}
	tab, err := grammar.Generate(g, ref)
	if err != nil {
		tb.Fatal(err)
	}
	classes, err := g.ResolveClasses(tab)
	if err != nil {
		tb.Fatal(err)
	}
	return tab.Tokens, classes
}

var (
	refStrictVersion = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]+)?$`)
	refLaxVersion    = regexp.MustCompile(`^[0-9][0-9_.]*$`)
)

// refPrefix is prefix as a regular expression reads a number: strict or lax version text, and only when
// whitespace, ;, {, } or the end of the file follows it. Anything else is prefix's.
func refPrefix(k lexer.Kind, tx string, src []byte, end int) byte {
	if k != lexer.Number {
		return prefix(k, tx, 0)
	}
	if end < len(src) && !bytes.ContainsRune([]byte(" \t\n\r\f\v;{}\x00"), rune(src[end])) {
		return 0
	}
	switch {
	case refStrictVersion.MatchString(tx):
		return grammar.VersionStrict
	case refLaxVersion.MatchString(tx):
		return grammar.VersionLax
	}
	return 0
}

// refClassOf is ClassOf as it stood before Intern was made cheaper, kept
// here so the reference shares no code with what it checks.
func refClassOf(t grammar.TokenView, pred grammar.Neighbour) grammar.Neighbour {
	c := grammar.Neighbour{Kind: t.Kind, OpensBlock: t.OpensBlock, TermFollows: t.TermFollows}
	switch t.Kind {
	case lexer.Word, lexer.Operator, lexer.CloseBracket:
		c.Text = t.Text
	}
	if t.Kind == lexer.Quote && (t.Prefix >= 'a' && t.Prefix <= 'z' || t.Prefix == 'Q') {
		c.Text = grammar.IdentQuote
	}
	if t.Kind == lexer.Operator && t.Text == "{" {
		switch {
		case t.OpensBlock, pred.Kind == lexer.DerefSigil, pred.Kind == lexer.FuncSigil, pred.Kind == lexer.Operator && pred.Text == "*" && pred.Glob:
			c.Brace = grammar.BraceBlock
		case pred.Kind == lexer.Variable, pred.Kind == lexer.CloseBracket && !pred.OpensBlock,
			pred.Kind == lexer.Operator && pred.Text == "->", pred.Kind == lexer.Word && pred.Name:
			c.Brace = grammar.BraceSubscript
		default:
			c.Brace = grammar.BraceHash
		}
	}
	c.Array = t.Kind == lexer.Variable && strings.HasPrefix(t.Text, "@") || t.Kind == lexer.Variable && strings.HasPrefix(t.Text, "$#")
	c.Name = t.Kind == lexer.Word && (pred.Kind == lexer.FuncSigil ||
		pred.Kind == lexer.Operator && (pred.Text == "->" || pred.Text == "*" && pred.Glob))
	// A label is a plain word with its colon on its line at a statement start; its colon carries the flag on.
	start := pred.Kind == -1 || pred.Kind == lexer.Semicolon || pred.Kind == lexer.Operator && pred.Text == "{" && pred.OpensBlock ||
		pred.Kind == lexer.CloseBracket && pred.Text == "}" && pred.OpensBlock || pred.Label
	c.Label = t.Kind == lexer.Word && t.LineColon && !t.Global && start || t.Kind == lexer.Operator && t.Text == ":" && pred.Label
	// sort's own (, not a method's.
	c.SortParen = t.Kind == lexer.Operator && t.Text == "(" && pred.Kind == lexer.Word && pred.Text == "sort" && !pred.Name
	return c
}

var (
	plain = regexp.MustCompile(`^[A-Za-z_\x{80}-\x{10FFFF}][A-Za-z0-9_\x{80}-\x{10FFFF}]*$`)
	// sigil is a `$` and the blanks after it; braced is a name in braces, blanks inside them allowed.
	sigil    = regexp.MustCompile(`^\$(?:[ \t\r\n\f\v]|#[^\n]*\n)*`)
	braced   = regexp.MustCompile(`^\{(?:[ \t\r\n\f\v]|#[^\n]*\n)*(.*?)[ \t\r\n\f\v]*\}$`)
	colonRef = regexp.MustCompile(`^[ \t\r\f]*:`)
)

// refGlobal and refLineColon are the properties read off the source with regular expressions, not the
// token stream: a word or `$` variable that is no plain identifier (or is `$_`), and a word whose rest of line starts
// (blanks only) with a colon that is not `::`.
func refGlobal(k lexer.Kind, tx string) bool {
	switch k {
	case lexer.Word:
		return !plain.MatchString(tx)
	case lexer.Variable:
		if !sigil.MatchString(tx) {
			return false
		}
		n := sigil.ReplaceAllString(tx, "")
		if m := braced.FindStringSubmatch(n); m != nil {
			n = m[1]
		}
		return n == "_" || !plain.MatchString(n)
	}
	return false
}

func refLineColon(rest []byte) bool {
	m := colonRef.Find(rest)
	return m != nil && !(len(rest) > len(m) && rest[len(m)] == ':')
}

// refBlanks is the Blanks bit read off the source bytes, not the token gaps: a plain-name word that a closing bracket
// follows after spaces or tabs alone, with spaces or tabs alone before it and, when a minus (not a --) comes first,
// before that minus too, and the start of the file or a character that is no whitespace there.
var afterBlanks = regexp.MustCompile(`^[ \t]*[)\]}]`)

func refBlanks(src []byte, start, end int) bool {
	if !plain.MatchString(string(src[start:end])) || !afterBlanks.Match(src[end:]) {
		return false
	}
	before := bytes.TrimRight(src[:start], " \t")
	if n := len(before); n > 0 && before[n-1] == '-' && (n < 2 || before[n-2] != '-') {
		before = bytes.TrimRight(before[:n-1], " \t")
	}
	return len(before) == 0 || !bytes.ContainsRune([]byte(" \t\n\r\f\v"), rune(before[len(before)-1]))
}

// refIDs is the terminal-set id of each significant token of src as the
// table encodes it: the union of the terminals of every class whose given
// parts all match the token's key, found in tt.Sets. It walks no table and
// hashes nothing, so it checks Intern and Lookup rather than restating them.
func refIDs(tb testing.TB, src []byte, tt *grammar.TokenTable, classes []grammar.TokenClass) []int {
	// A neighbour of grammar.OtherWord is a word, flags as given, that no class names.
	named := map[string]bool{}
	for _, c := range classes {
		if c.Text != nil {
			named[*c.Text] = true
		}
		for _, n := range append(append([]*grammar.Neighbour{c.Prev2, c.Prev1, c.Next}, ptrs(c.NotPrev1)...), append(ptrs(c.NotPrev2), ptrs(c.NotNext)...)...) {
			if n != nil && n.Text != grammar.NonIdentWord {
				named[n.Text] = true
			}
		}
		for _, x := range c.NotText {
			named[x] = true
		}
	}
	is := func(want *grammar.Neighbour, got grammar.Neighbour) bool {
		if want != nil && want.Text == grammar.NonIdentWord {
			return got.Kind == lexer.Word && !named[got.Text] && !(got.Text[0] == '_' || got.Text[0]|0x20 >= 'a' && got.Text[0]|0x20 <= 'z' || got.Text[0] >= 0x80)
		}
		if want != nil && want.Text == grammar.OtherWord {
			w := *want
			w.Text = got.Text
			return w == got && !named[got.Text] && (got.Kind != lexer.Word || got.Text[0] == '_' || got.Text[0]|0x20 >= 'a' && got.Text[0]|0x20 <= 'z' || got.Text[0] >= 0x80)
		}
		return want == nil || *want == got
	}
	// completes is whether a token leaves PL_expect at XOPERATOR: a name, or a token some class marks
	// Completes, other than a block's }, which ends a statement more often than it ends a term.
	completes := func(n grammar.Neighbour) bool {
		if n.Name {
			return true
		}
		for _, c := range classes {
			if !c.Completes || c.Prev1 != nil {
				continue
			}
			w := grammar.Neighbour{Kind: *c.Kind, OpensBlock: c.OpensBlock != nil && *c.OpensBlock,
				Array: *c.Kind == lexer.Variable && c.Prefix != nil && slices.Contains([]byte("@#"), *c.Prefix)}
			switch {
			case c.Text != nil:
				w.Text = *c.Text
			case *c.Kind == lexer.Word:
				w.Text = grammar.OtherWord
			case *c.Kind == lexer.Quote && c.Prefix != nil && (*c.Prefix >= 'a' && *c.Prefix <= 'z' || *c.Prefix == 'Q'):
				w.Text = grammar.IdentQuote
			}
			if w.Kind == lexer.CloseBracket && w.Text == "}" && w.OpensBlock {
				continue
			}
			if is(&w, n) {
				return true
			}
		}
		return false
	}
	type tok struct {
		v grammar.TokenView
		n grammar.Neighbour
	}
	var toks []tok
	pred := grammar.Boundary
	var opens []int // the openers still waiting for a body, by index in toks
	for _, t := range lexer.Tokenize(src) {
		switch {
		case lexer.IsTrivia(t.Kind):
			continue
		case t.Kind == lexer.HeredocBody && len(toks) > 0:
			// The body's last line names its heredoc; a body no waiting opener names is a heredoc's opened inside a quote-like.
			lines := strings.Split(strings.TrimRight(string(src[t.Start:t.End]), "\r\n"), "\n")
			end := refTag(lines[len(lines)-1])
			if j := slices.IndexFunc(opens, func(w int) bool { return refTag(toks[w].v.Text) == end }); j >= 0 {
				opens = slices.Delete(opens, j, j+1)
			}
			continue
		case t.Kind == lexer.HeredocOpen:
			opens = append(opens, len(toks))
		}
		tx := string(src[t.Start:t.End])
		v := grammar.TokenView{Kind: t.Kind, Text: tx, OpensBlock: t.OpensBlock, TermFollows: t.TermFollows, Prefix: refPrefix(t.Kind, tx, src, t.End),
			LineColon: t.Kind == lexer.Word && refLineColon(src[t.End:]), Blanks: t.Kind == lexer.Word && refBlanks(src, t.Start, t.End)}
		v.Global = (v.LineColon || t.Kind == lexer.Variable) && refGlobal(t.Kind, tx)
		prev := pred
		pred = refClassOf(v, prev)
		if t.Kind == lexer.Operator && tx == "*" {
			pred.Glob = !completes(prev)
		}
		toks = append(toks, tok{v, pred})
	}
	nb := func(i int) grammar.Neighbour {
		if i < 0 || i >= len(toks) {
			return grammar.Boundary
		}
		return toks[i].n
	}
	// ident is whether toks[i] starts with an identifier character: a word opening with a letter or _, or a quote-like opening with a letter.
	ident := func(i int) bool {
		return i >= 0 && i < len(toks) && (toks[i].v.Kind == lexer.Word && (toks[i].v.Text[0] == '_' || toks[i].v.Text[0]|0x20 >= 'a' && toks[i].v.Text[0]|0x20 <= 'z' || toks[i].v.Text[0] >= 0x80) ||
			toks[i].v.Kind == lexer.Quote && (toks[i].v.Prefix >= 'a' && toks[i].v.Prefix <= 'z' || toks[i].v.Prefix == 'Q'))
	}
	// name is whether a sub's name starts at toks[i]: an identifier start, or a word opening with :: or '.
	name := func(i int) bool { return ident(i) || i >= 0 && i < len(toks) && toks[i].v.Kind == lexer.Word }
	// lone is whether a Lone class holds at toks[i]: all of its neighbours match, and a word class other than a Lone one is cut off there.
	lone := func(i int) bool {
		return slices.ContainsFunc(classes, func(l grammar.TokenClass) bool {
			return l.Lone && is(l.Prev2, nb(i-2)) && is(l.Prev1, nb(i-1)) && is(l.Next, nb(i+1)) && (l.Blanks == nil || *l.Blanks == toks[i].v.Blanks)
		})
	}
	ids := make([]int, len(toks))
	for i, t := range toks {
		var set []int
		alone := lone(i)
		for _, c := range classes {
			if alone && !c.Lone && c.Kind != nil && *c.Kind == lexer.Word {
				continue
			}
			if (c.Kind == nil || *c.Kind == t.v.Kind) && (c.Text == nil || *c.Text == t.v.Text) &&
				(c.OpensBlock == nil || *c.OpensBlock == t.v.OpensBlock) && (c.TermFollows == nil || *c.TermFollows == t.v.TermFollows) &&
				(c.Prefix == nil || *c.Prefix == t.v.Prefix) && (c.Global == nil || *c.Global == t.v.Global) &&
				(c.LineColon == nil || *c.LineColon == t.v.LineColon) && (c.Blanks == nil || *c.Blanks == t.v.Blanks) && is(c.Prev2, nb(i-2)) && is(c.Prev1, nb(i-1)) && is(c.Next, nb(i+1)) &&
				!slices.Contains(c.NotText, t.v.Text) && !slices.ContainsFunc(c.NotPrev1, func(n grammar.Neighbour) bool { return is(&n, nb(i-1)) }) &&
				!slices.ContainsFunc(c.NotPrev2, func(n grammar.Neighbour) bool { return is(&n, nb(i-2)) }) && !(c.NotNextIdent && ident(i+1)) && !(c.NotNextName && name(i+1)) && (!c.NextName || name(i+1)) && !slices.ContainsFunc(c.NotNext, func(n grammar.Neighbour) bool { return is(&n, nb(i+1)) }) {
				set = append(set, c.Terminals...)
			}
		}
		slices.Sort(set)
		set = slices.Compact(set)
		id, ok := slices.BinarySearchFunc(tt.Sets, set, slices.Compare)
		if !ok {
			tb.Fatalf("token %d (%q): the union %v is not one of the table's sets", i, t.v.Text, set)
		}
		ids[i] = id
	}
	for _, i := range opens { // no body came, so perl finds no terminator
		ids[i] = 0
	}
	return ids
}

// corpusFiles is every waypoint-1 corpus file the ratchet reads: the
// in-tree ones always, T1's when T1_CORPUS names it.
func corpusFiles(t *testing.T) []string {
	var files []string
	globs := []string{"../testdata/guacamole/parses/*.pl", "../testdata/guacamole/parsent/*.pl", "../testdata/std5k.pl", "../../psc/testdata/latency/edit5k.pl"}
	if t1 := os.Getenv("T1_CORPUS"); t1 != "" {
		globs = append(globs, filepath.Join(t1, "*.t"))
	} else {
		t.Log("T1_CORPUS is unset: the in-tree corpus only")
	}
	for _, g := range globs {
		m, err := filepath.Glob(g)
		if err != nil || len(m) == 0 {
			t.Fatalf("no corpus file matches %s", g)
		}
		files = append(files, m...)
	}
	return files
}

// Every significant token of every corpus file gets the terminal-set id the
// reference gives it, so a cheaper Intern changes no id.
func TestInternIdentityOverCorpus(t *testing.T) {
	tt, classes := committed(t)
	files := corpusFiles(t)
	tokens := 0
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		s, _ := Lex(src, tt)
		want := refIDs(t, src, tt, classes)
		got := make([]int, len(s.Tokens))
		for i, tok := range s.Tokens {
			got[i] = tok.TerminalSet
		}
		if i := slices.Compare(got, want); i != 0 {
			for j := range min(len(got), len(want)) {
				if got[j] != want[j] {
					r := s.Raw[s.Tokens[j].Index]
					t.Errorf("%s: token %d (%q at byte %d) has set %d, the reference %d", f, j, src[r.Start:r.End], r.Start, got[j], want[j])
					break
				}
			}
			if len(got) != len(want) {
				t.Errorf("%s: %d tokens, the reference %d", f, len(got), len(want))
			}
		}
		tokens += len(want)
	}
	t.Logf("%d files, %d tokens", len(files), tokens)
}

// colonForms are the : after a declarator and a variable with each kind of token after it, and the : after other things.
const colonForms = "$c ? my $x : shared; $c ? my $x :shared; $c ? our @x : if; $c ? state $x : qw(a); $c ? my $x : q(a); " +
	"$c ? my $x : my $z; $c ? my $x : 1; $c ? my $x : $z; $c ? my $x : -1; $c ? my $x : \"s\"; $c ? my $x : \\$z; " +
	"$c ? my $x : 'a'; $c ? my $x : s/a/b/; $c ? my $x : tr/a/b/; $c ? my $x : y/a/b/; $c ? my $x : m/a/; $c ? my $x : qq(a); " +
	"$c ? my $x : ::foo; use utf8; $c ? my $x : é; my $x :Foo é bar; my $x : Foo ::bar baz; $c ? (my $x) : shared; $c ? local $x : shared; $c ? $x : shared; $c ? my ($x) : shared; $c ? my $x :\n"

// loopForms start the source, for NotPrev2's start of file: for our and state with a parenthesis at each kind of statement start,
// and where it is not one.
const loopForms = "for our ($a) (1) {} 1; for state ($a) (1) {} { foreach our ($a) (1) {} } sub f {} for our ($a) (1) {} " +
	"1 for our ($x); do {1} for state ($x); $h{a} for our ($x); for our $x (1) {} $c ? 1 : 2 for our ($x); L: for our ($x) (1) {} "

// loneForms are a word alone in a subscript's braces, with a minus before it or not, and the words that are not alone there: a
// keyword, a name and x among them, in a hash subscript, an arrow's, a slice's and an anonymous hash's, which keeps its call readings.
const loneForms = "$h{time}; $h->{shift}; $h{ if }; $h{x}; $h{-if}; $h{ - while }; $h{a}{print}; @h{time}; $h{time + 1}; $h{shift @a}; $h{-time + 1}; " +
	"+{ time }; +{ shift }; $h{time, 1}; $h{$a}; $h{Foo::time}; $h{ q }; $a[time]; $r->[0]{y}; f()->{s}; " +
	"$h{if\n}; $h{\nif}; $h{if # c\n}; $h{-if\n}; $h{q\n}; $h{time\n}; $h{\tif}; $h{ -\nif}; $h{\n-if}; $h{Foo'bar}; $h{::foo}; $h{foo::}; $h{Foo::bar}; $a\n-if}; $a --if}; $a\n--if}; "

// The negated parts (NotText, NotPrev1, NotPrev2, NotNext, NotNextIdent) need keys the corpus may not hold: x after a unary
// operator, a keyword as a word, a unary operator before a parenthesis.
func TestInternIdentityNegatedParts(t *testing.T) {
	tt, classes := committed(t)
	src := []byte(loopForms + "shift x 3; $a x 3; ref($a) x 3; ref x 3; last x; not (x); if; foo; time x 3; L: 1;\nL :1; L\n: 1; L #c\n: 1; L::M: 1; L'M: 1; $::e $a'b $' @Foo::x; L :: 1; catch ($\ne) catch ($ #c\ne) catch ($ #c\n e) catch (${e}) catch (${ e }) catch ($ e) catch (${_}) catch (${1}) catch ($ _) catch (${ ^E }) catch (${a}) catch (${__}) catch ($ ARGV); use utf8; é: 1; catch ($é); no utf8; L: for our ($x) (1) {} L: for state ($x) (1) {} my $y : for our ($x); A: B: for our ($x) (1) {} $c ? L : for our ($x); {L: for our ($x) (1) {}} ; M: our $z; catch ($::e) catch ($Foo'e) catch ($') catch ($e) catch ($_) catch ($1) catch ($^W) catch ($_e); catch ($::e) catch ($Foo'e) catch ($') catch ($e) catch ($_) catch ($1) catch ($^W) catch ($_e); " + colonForms +
		"$y = @a * ref x 3; $y = $#a * ref x 3; $y = @a * 2; @a x 3; $#a x 3; my @b = @a x 3; $c ? my @z : shared; $y = qw(a) * ref x 3; $y = q(a) * ref x 3; $y = qq(a) * ref x 3; foo * ref x 3; $y = foo * ref x 3; bar foo * ref x 3; $y = *print * ref x 3; $y = &print * ref x 3; $y = $o->print * ref x 3; $y = 2 * ref x 3; *print x 3; $y = $a * shift x 3; $y = *print x 3; $y = 2 * *print x 3; $y = f() * ref x 3; " +
		"format print =\n.\n" +
		"sub ::foo { 1 } sub 'foo; BEGIN ::foo { 1 } END ::x; sub open { 1 } sub join { 1 } sub foo { 1 } " + loneForms +
		"print FH x 3; print x x 3; sort foo x 3; join foo x 3; new Foo x 3; new Foo(1); foo $x; foo { 1 } 2; &foo x 3; open FH; " +
		"sort(foo x 3); sort (foo @a); $o->sort(foo x 3); sort(foo) x 3; foo { a => 1 }; &$foo; my $x :sort(1); my $x :sort($y); " +
		"\n\"a\"(1); \"a$x\"(1); qq{a}(1); `ls`(1); qx(ls)(1); <STDIN>(1); <*.c>(1); <<A(1);\na\nA\n<<'  E' . <<~F;\n  E\n  F\n" +
		"$y = $$x; ${ $x }; @{$x}; %$x; @$x; $r->$*; $r->@*; $r->%*; $r->&*; $r->**; $r->$#*; $r->@[0]; $r->%{a}; $r->%[0]; $r->$m; $r->@foo; $r->$#foo; " +
		"*g{CODE}; *g{if}; *{$x}; *{$x}{CODE}; $r->*{CODE}; *$x; $r->$* x 3; $r->@* x 3; $y = @*; $y = $*; &{$x}(1); \\&{$x}; " +
		"print \"@{[ <<B ]}\" . <<A;\nb\nB\n")
	// The last line's <<A has no body (the one after it is the string's <<B), so it is refused: pairing by tag, not by order.
	s, _ := Lex(src, tt)
	want := refIDs(t, src, tt, classes)
	for i, tok := range s.Tokens {
		if tok.TerminalSet != want[i] {
			t.Errorf("token %d has set %d, the reference %d", i, tok.TerminalSet, want[i])
		}
	}
	if len(s.Tokens) != len(want) {
		t.Errorf("%d tokens, the reference %d", len(s.Tokens), len(want))
	}
}

// The version shapes (chart/lex's Prefix for a Number) need numbers the corpus may not hold after a package name, and
// the byte after each: a strict, a lax and a plain number, each delimited by every byte toke.c ends a version at and by
// ones it does not.
func TestInternIdentityVersionShapes(t *testing.T) {
	tt, classes := committed(t)
	var src []byte
	for _, n := range []string{"1", "1.23", "0.1", "01", "1_0", "1.", "1.2.3", "0x10", "1e3", ".5", "10.01"} {
		for _, d := range []string{";", " ;", "\t;", "\n;", "{ }", "}", ",", "#c\n;", "x;", ")", "'a';", ""} {
			src = append(src, "package Foo "+n+d+"\nuse "+n+d+"\nuse lib "+n+d+"\n$y = "+n+d+"\n"...)
		}
	}
	// What follows a version or a word after use MODULE: each operator a NotNext list names, the ones it leaves, a term.
	for _, n := range []string{"1.2", "v1.2", "foo", "x", "Foo::Bar"} {
		for _, f := range []string{"=>", ".", "x", "?", "==", "!=", "||", "//", "&&", "..", "...", "=~", "!~", "**", "<=>", "eq", "ne", "lt", "gt", "cmp",
			"=", "+=", "->", "|", "^", ">>", ">", ">=", ",", "-", "+", "*", "%", "&", "<", "/", "<<", "1", "'a'", `"a"`, "qw(a)", "[1]", `\1`, "$x", "bar", "(3)", "{ }", "-1", ";",
			"le", "ge", "**=", "//=", "@x", "$#x", "@$x", "!", "~", "sub", "do", "my", "print", "__FILE__", "::bar", "%h", "&f", "*g", "__END__", "shift", "if", "eq", "isa"} {
			src = append(src, "use lib "+n+" "+f+" 3;\nno lib "+n+" "+f+" 3;\n"...)
		}
	}
	s, _ := Lex(src, tt)
	want := refIDs(t, src, tt, classes)
	for i, tok := range s.Tokens {
		if tok.TerminalSet != want[i] {
			r := s.Raw[tok.Index]
			t.Errorf("token %d (%q at byte %d) has set %d, the reference %d", i, src[r.Start:r.End], r.Start, tok.TerminalSet, want[i])
		}
	}
	if len(s.Tokens) != len(want) {
		t.Errorf("%d tokens, the reference %d", len(s.Tokens), len(want))
	}
}

// std5k is std5k.pl and its significant tokens.
func std5k(b *testing.B) ([]byte, []lexer.Token) {
	src, err := os.ReadFile("../testdata/std5k.pl")
	if err != nil {
		b.Fatal(err)
	}
	var sig []lexer.Token
	for _, t := range lexer.Tokenize(src) {
		if !lexer.IsTrivia(t.Kind) && t.Kind != lexer.HeredocBody {
			sig = append(sig, t)
		}
	}
	return src, sig
}

func perToken(b *testing.B, n int) {
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(n), "ns/token")
}

// BenchmarkIntern is Intern alone, each token given its predecessor's class
// as Map gives it, reported per significant token.
func BenchmarkIntern(b *testing.B) {
	tt, _ := committed(b)
	src, sig := std5k(b)
	text := unsafe.String(unsafe.SliceData(src), len(src))
	b.ResetTimer()
	for range b.N {
		var slots [2]grammar.Token
		pred := &grammar.Boundary
		for i, t := range sig {
			intern(tt, text, t, false, false, pred, &slots[i&1])
			pred = &slots[i&1].N
		}
	}
	perToken(b, len(sig))
}

// BenchmarkTokenize is the reference for Intern's share: lexer.Tokenize per
// significant token of the same file.
func BenchmarkTokenize(b *testing.B) {
	src, sig := std5k(b)
	b.ResetTimer()
	for range b.N {
		lexer.Tokenize(src)
	}
	perToken(b, len(sig))
}

func ptrs(ns []grammar.Neighbour) []*grammar.Neighbour {
	ps := make([]*grammar.Neighbour, len(ns))
	for i := range ns {
		ps[i] = &ns[i]
	}
	return ps
}

// The Blanks bit chart/lex reads off the token gaps is the one the reference reads off the bytes, on every word of the lone
// forms, and it is the expected bit on the rows that name one: a newline, a comment or a package part clears it.
func TestBlanksAgreeWithReference(t *testing.T) {
	src := []byte(loneForms)
	raw := lexer.Tokenize(src)
	for i, r := range raw {
		if r.Kind == lexer.Word {
			if got, want := blanks(raw, i, string(src)), refBlanks(src, r.Start, r.End); got != want {
				t.Errorf("%q at byte %d: chart/lex says %v, the reference %v", src[r.Start:r.End], r.Start, got, want)
			}
		}
	}
	for _, c := range []struct {
		src  string
		want bool
	}{
		{"$h{if}", true}, {"$h{ if }", true}, {"$h{\tif}", true}, {"$h{ - if }", true}, {"$h{-if}", true},
		{"$h{if\n}", false}, {"$h{\nif}", false}, {"$h{if # c\n}", false}, {"$h{-if\n}", false}, {"$h{ -\nif}", false}, {"$h{\n-if}", false},
		{"$h{Foo::bar}", false}, {"$h{Foo'bar}", false}, {"$h{::foo}", false}, {"$h{foo::}", false},
		{"$h{if 1}", false}, {"if}", true},
	} {
		raw := lexer.Tokenize([]byte(c.src))
		got := false
		for i, r := range raw {
			if r.Kind == lexer.Word {
				got = blanks(raw, i, c.src)
			}
		}
		if got != c.want {
			t.Errorf("%q: Blanks is %v, want %v", c.src, got, c.want)
		}
	}
}

// refTag is the word a heredoc opener waits for, or a terminator line names: what is left once <, ~, blanks, backslashes and quotes are gone.
func refTag(opener string) string {
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune("<~ \t\\'\"`", r) {
			return -1
		}
		return r
	}, opener)
}
