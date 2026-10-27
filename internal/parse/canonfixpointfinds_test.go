// ABOUTME: FuzzParser's pre-existing finds: canon fixpoint failures from `no &0` on, and a nest that hung it.
// ABOUTME: Each is read as perl 5.44.0 reads it, or refused where perl -c rejects it.

package parse_test

import (
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/parse"
)

// clean parses src and fails unless it has no refusal and no error, then
// returns its canon.
func clean(t *testing.T, src string) string {
	t.Helper()
	root := parse.Parse([]byte(src))
	if found := collectUnknown(root); len(found) != 0 || root.HasError {
		t.Errorf("%q: perl compiles this, but it refuses: %v", src, refusalCodes(found))
	}
	return parse.Canon(root, []byte(src))
}

// TestStarPercentAfterWordParses: after an undeclared word, `*%x` is a
// multiplication by `%x`, not the glob `*%` followed by the repetition `x`.
// perl -c on 5.44.0 accepts `A*%x`, `A*@x`, `A*&x` and `A*%$h`, and rejects
// `A*%x*(%A` at EOF; the glob reading made the latter's canon re-read
// differently. A declared A takes the glob, `sub A {} A*%x3` being
// `A(*% x 3)`, which the parser declines rather than mis-read.
func TestStarPercentAfterWordParses(t *testing.T) {
	for src, want := range map[string]string{
		"A*%x":  "A * %x",
		"A*@x":  "A * @x",
		"A*&x":  "A * &x",
		"A*%$h": "A * %$h",
		// And a punctuation name, as perl's scan_ident takes one: `&;`
		// calls the sub `;`, so the statement still needs its terminator.
		"got*&;": "got * &;;",
		"got*%;": "got * %;",
		"A*@}":   "A * @}",
		"x*$;":   "x * $;",
		// And a name apart from its sigil: FuzzParser's `A*& ;`.
		"A*& ;":  "A * & ;;",
		"A*% x;": "A * % x;",
		// And any other punctuation the operand begins with.
		"A*/x/;":  "A * /x/;",
		"A*<FH>;": "A * <FH>;",
		"A*.5;":   "A * .5;",
	} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
	for _, src := range []string{"sub A {} A *%x3", "sub A {} A*$;", "sub A {} A *% x 3"} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) == 0 || found[0].Refusal != parse.AmbiguousSigil {
			t.Errorf("%q: want an ambiguous_sigil refusal, got %v", src, refusalCodes(found))
		}
	}
	// And FuzzParser's `(A*"=>)`, whose `"` no `=>` quotes, as it does the
	// `s` of `A*s => 1`.
	for _, src := range []string{`(A*"=>)`, "A*#=>"} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) == 0 || found[0].Refusal != parse.AmbiguousSigil {
			t.Errorf("%q: want an ambiguous_sigil refusal, got %v", src, refusalCodes(found))
		}
	}
	clean(t, "A*s => 1;")
	// `{x*$}` is `{'x' * $}` with its `}` missing, to perl 5.44.0, and
	// FuzzParser's `(A*/)` an unterminated pattern.
	for _, src := range []string{"{x*$}", "(A*/)", "A*?;", "A*|"} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) == 0 {
			t.Errorf("%q: perl rejects this, but it parses", src)
		}
	}
	if ok, why := canonMatches([]byte("A*%x*(%A")); !ok {
		t.Errorf("`A*%%x*(%%A`: canon is not a fixpoint: %s", why)
	}
}

// TestUseWithoutModuleRefuses: `use` and `no` take a module name or a
// version. perl -c on 5.44.0 rejects `no &0`, `no & ! no & 00`, `use
// 'strict'`, `use $x`, `use +strict` and `use;`; read on, `no &0` took `&`
// as an import list whose canon `&;` re-read differently. `use => 1` is the
// string "use", which perl compiles.
func TestUseWithoutModuleRefuses(t *testing.T) {
	for _, src := range []string{"no &0", "no & ! no & 00", "use 'strict';", "use $x;", "use +strict;", "use;", "no"} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) == 0 {
			t.Errorf("%q: perl rejects this, but it parses", src)
		}
	}
	for _, src := range []string{"use => 1;", "no => 1;", "no 6;", "use v5.10;", "use 5.010;", "use strict;"} {
		clean(t, src)
	}
}

// TestAmpBeforeTildeAfterWord: `;A&~ ~0` is `'A' & ~~0` to perl 5.44.0, and
// so is its canon `;A & ~~0`. The lexer took the `&` for a sub's sigil and
// the `~~` after it for smartmatch, so the canon refused.
func TestAmpBeforeTildeAfterWord(t *testing.T) {
	for _, src := range []string{";A&~ ~0", ";A & ~~0", "A & ~~0"} {
		clean(t, src)
		if ok, why := canonMatches([]byte(src)); !ok {
			t.Errorf("%q: canon is not a fixpoint: %s", src, why)
		}
	}
}

// TestAmpDigitsNameASub: perl 5.44.0 compiles `goto&0`, `&0` and `f(&0)`,
// each calling the sub named `0`. Lexed as a sigil and a number, `goto&0`
// was `goto &` and a second statement, and its canon re-read differently.
func TestAmpDigitsNameASub(t *testing.T) {
	// The name may stand apart from the `&`: `& 0;` deparses as `&0;`, and
	// FuzzParser's `!& 0` was `!&` and then `0`.
	for src, want := range map[string]string{"goto&0": "goto &0", "&0": "&0", "f(&0)": "f(&0)", "&12;": "&12;", "!& 0": "!& 0"} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
	// A longer name may not start with 0: "Numeric variables with more than
	// one digit may not start with '0'". FuzzParser's `no n!&00`.
	// And one across a heredoc body is declined: FuzzParser's `<<'0';&` +
	// newline + body + `A`, which perl reads as `&A` but canon wrote with
	// the body twice.
	for _, src := range []string{"&00;", "f(&01)", "no n!&00", "& 00;", "<<'0';&\n0\nA"} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) == 0 {
			t.Errorf("%q: perl rejects this, but it parses", src)
		}
	}
}

// TestAmpPunctNameEndsTerm: `&!` calls the sub named `!`, so an operator
// follows it. perl 5.44.0 compiles `A*&!*&%*!x &0` as `A * &! * &% * !x &
// 0`; with the `!` lexed as an operator, the `*` after it read as a glob.
func TestAmpPunctNameEndsTerm(t *testing.T) {
	for src, want := range map[string]string{
		"A*&!*&%*!x &0": "A * &! * &% * !x & 0",
		"&!*2;":         "&! * 2;",
		// And `:` alone: FuzzParser's `&:*$%%(%A0`.
		"&: * 2;": "&: * 2;",
		"&:*2;":   "&: * 2;",
		"&::foo;": "&::foo;",
		// And a keyword's spelling: `&gt` is a call, not the operator.
		"&gt*&;": "&gt * &;;",
		// And a glob's name: FuzzParser's `*gt*&;`.
		"*gt*&;": "*gt * &;;",
		// But no word after an undeclared word's `*`, which multiplies:
		// FuzzParser's `A*A{0}` is `'A' * A {0}`.
		"A*A{0}": "A * A {0;}",
		// Nor after a closer: FuzzParser's `{}*A{0}`, `{} * A {0}` to perl.
		"{}*A{0}": "{} * A {0;}",
		"*A{0};":  "*A{0};",
	} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
}

// TestWordOperatorIsNoTerm: perl 5.44.0 rejects `gt*&;`, `gt*2;` and
// `print -eq;`, and compiles `$h{eq}`, `{eq => 1}` and `x*&;`, the last
// `'x' * &;`. Read as a call, `gt*&;`'s canon re-read differently.
func TestWordOperatorIsNoTerm(t *testing.T) {
	// Nor is `{ge}` a key outside a subscript: perl rejects the block and
	// the anonymous hash alike, and FuzzParser's `{ge}` canon'd as `{ge;}`.
	// So is `sub`'s, FuzzParser's `{sub}`: "Illegal declaration of
	// anonymous subroutine".
	for _, src := range []string{"gt*&;", "gt*2;", "print -eq;", "CORE::eq*2", "{ge}", "my $r = {ge};", "{sub}", "do {sub}"} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) == 0 {
			t.Errorf("%q: perl rejects this, but it parses", src)
		}
	}
	for _, src := range []string{"my %h; $h{eq}", "my $r = { eq => 1 };", "f(-lt => 1)", "$_{ge}", "my $r; $r->{eq}", "my %h; @h{lt}", "my %h; $h{sub}"} {
		clean(t, src)
	}
	// And after the repetition `x`, where a term is read too: FuzzParser's
	// `1x x*$(&(())` is `1 x 'x' * $( & ()`.
	for src, want := range map[string]string{"x*&;": "x * &;;", "1x x*$(&(())": "1 x x * $( & ()", "x&!x*2;": "x & !x * 2;"} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
}

// TestListSeparatorAfterInnerTrailingComma: perl 5.44.0 compiles
// `((A,),%y)`. The comma inside `(A,)` is that element's, and canon took it
// for the list's separator, writing `((A,) %y)`, a modulus.
func TestListSeparatorAfterInnerTrailingComma(t *testing.T) {
	for src, want := range map[string]string{"((A,),%y)": "((A,), %y)", "((1,),2)": "((1,), 2)", "((A),,%y)": "((A), , %y)", "(0,A,,0)": "(0, A, , 0)"} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
}

// TestListSeparatorInOuterPair: perl 5.44.0 compiles `((1,2),) x 2` and
// FuzzParser's `((),)x*!`. Canon wrote the separator after the list's own
// pair and dropped the one around it, `(1, 2), x 2`, so the list ended
// before the `x`.
func TestListSeparatorInOuterPair(t *testing.T) {
	for src, want := range map[string]string{
		"my @a = ((1,2),) x 2;": "my @a = ((1, 2),) x 2;",
		"my @a = ((),) x 2;":    "my @a = ((),) x 2;",
		"((),)x*!":              "((),) x *!",
		"f((1,2),)":             "f((1, 2),)",
		// And one after the outer pair too: FuzzParser's `((),),`.
		"((),),":    "((),),",
		"((1,2),),": "((1, 2),),",
		// And the list's separator after that pair: FuzzParser's
		// `(((),),,0)`, written `(((),) 0)`, a call of `()`.
		"(((),),0)":  "(((),), 0)",
		"(((),),,0)": "(((),), , 0)",
		// The parser keeps the one inside ahead of these: FuzzParser's
		// `(((),),,,0)` lost it, and canon then misplaced one.
		"(((),),,,0)": "(((),), , , 0)",
		// And more than one inside: FuzzParser's `A:(((),,)%$!)`.
		"A:(((),,)%$!)": "A: (((), ,) % $!)",
		// And a pair around that pair: FuzzParser's `A:(((),))x$0`.
		"A:(((),))x$0": "A: ((),) x $0",
	} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
}

// TestKeptSeparatorWrittenOnce: a node that keeps its trailing separator has
// emit write it, so the text written from source for it leaves it out.
// FuzzParser's `((&0()),)` canon'd as `((&0()),,)`.
func TestKeptSeparatorWrittenOnce(t *testing.T) {
	// And all of them: FuzzParser's `((&0()),,)`.
	for src, want := range map[string]string{"((&0()),)": "((&0()),)", "((&f()),)": "((&f()),)", "(&f(),)": "(&f(),)", "((1),)": "((1),)", "((&0()),,)": "((&0()), ,)"} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
}

// TestSeparatorAfterHeredocBody: perl 5.44.0 compiles `<<""` + newline +
// an empty line + `,,0`, the separators after the body. Canon wrote them
// before it, and the re-read placed the body after them.
func TestSeparatorAfterHeredocBody(t *testing.T) {
	for src, want := range map[string]string{"<<\"\"\n\n,,0": "<<\"\"\n\n,, 0", "print <<\"\",1\n\n": "print <<\"\", 1\n\n"} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
}

// TestFiletestTakesGlob: perl 5.44.0 deparses `(-A*y)` as `-A *y`, the
// filetest of a glob. Read as `-A() * y`, FuzzParser's find canon'd with
// `y)` a transliteration.
func TestFiletestTakesGlob(t *testing.T) {
	for src, want := range map[string]string{"(-A*y)": "(-A *y)", "-e *STDIN;": "-e *STDIN;", "-e;": "-e;"} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
}

// TestPostfixIncrementEndsTerm: perl 5.44.0 reads `A++*&;` as `A++ * &;`.
// The lexer took a term to follow the `++`, read the glob `*&`, and the
// re-read of canon's `A++ * &;` differed.
func TestPostfixIncrementEndsTerm(t *testing.T) {
	for src, want := range map[string]string{"A++*&;": "A++ * &;;", "my $x; $x++ / 2;": "my $x;$x++ / 2;", "my $x; $x++*2;": "my $x;$x++ * 2;"} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
}

// TestSpacedPercentAfterWordIsModulus: perl 5.44.0 reads `A% #` + newline
// + `<<0` as `'A' % <<0`. Canon wrote `A % <<0`, where `% <` lexed as a
// hash named `<`.
func TestSpacedPercentAfterWordIsModulus(t *testing.T) {
	for src, want := range map[string]string{"A% #\n<<0;;0\n0": "A % <<0;;0\n0", "A % 2;": "A % 2;"} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
}

// TestUnclosedDerefNestIsNotExponential: derefHoldsStatements trial-parses
// a deref's braces, and the trial held the nested deref's own trial, so
// each unclosed `${` doubled the work: FuzzParser's
// `${${${...${&{zx` (19 levels) took 4.4 s and hung its worker. 14 levels
// took 147 ms, so 40 would take days; held, they take milliseconds.
func TestUnclosedDerefNestIsNotExponential(t *testing.T) {
	src := strings.Repeat("${", 40) + "&{zx"
	start := time.Now()
	parse.Parse([]byte(src))
	if d := time.Since(start); d > 5*time.Second {
		t.Errorf("40 unclosed `${` took %v: each level multiplies the work", d)
	}
}

// TestSpacedAngleAfterWordCompares: perl 5.44.0 reads `A:A<` + newline +
// `0>0` as `A: 'A' < 0 > 0`; canon wrote it on one line, where `< 0 >`
// lexed as a readline. A declared sub's `<STDIN>` still reads a line.
func TestSpacedAngleAfterWordCompares(t *testing.T) {
	// And a shift, `A:A<<` + newline + `0>0`.
	for src, want := range map[string]string{"A:A<\n0>0": "A: A < 0 > 0", "A < 0 > 0": "A < 0 > 0", "sub f {} f <STDIN>;": "sub f {} f <STDIN>;", "A:A<<\n0>0": "A: A << 0 > 0", "A:A<=\n0>0": "A: A <= 0 > 0"} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
}

// TestFormFeedBeforeBrace: a form feed is whitespace to perl's lookahead
// at a `{` as a blank is, so `A\f{0}` is `A {0}`, a block. Read as a
// subscript, its canon `A{0}` re-read as the block.
func TestFormFeedBeforeBrace(t *testing.T) {
	for src, want := range map[string]string{"A\f{0}": "A {0;}", "A\v{0}": "A {0;}"} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
}

// TestIncrementOverPrefixCanon: canon wrote `+++y=>` as `++(+y)=>`, and
// `y)=>` re-lexed as a transliteration. A blank keeps `++` and the prefix
// operator apart without a paren the source lacks; perl 5.44.0 compiles
// `my $y; ++ +$y;`.
func TestIncrementOverPrefixCanon(t *testing.T) {
	for src, want := range map[string]string{"my $y; ++ +$y;": "my $y;++ +$y;", "my $y; --+$y;": "my $y;-- +$y;"} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
	if ok, why := canonMatches([]byte("+++y=>")); !ok {
		t.Errorf("`+++y=>`: canon is not a fixpoint: %s", why)
	}
}

// TestNumberTakesNoSubscript: perl 5.44.0 rejects `A{},(0{},)` near "0{".
// Read as a hash subscript on 0, its canon doubled the trailing comma and
// re-read differently. A list slice, `(1)[0]`, is no number.
func TestNumberTakesNoSubscript(t *testing.T) {
	// Nor is `&{}` a call: FuzzParser's `(&{},)`, perl's syntax error near
	// "&{", read as a subscript on a bare `&`.
	// Nor is an empty subscript one: FuzzParser's `($0{},)`.
	for _, src := range []string{"A{},(0{},)", "(0{},)", "(0[1])", "(&{},)", "&{};", "($0{},)", "my %h; $h{}", "my @a; $a[]", "my $r; $r->{}"} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) == 0 {
			t.Errorf("%q: perl rejects this, but it parses", src)
		}
	}
	for _, src := range []string{"my @a = (1)[0];", "my %x; ($x{a},)", "my @x; ($x[0],)", "my $x; (&{$x},)"} {
		clean(t, src)
	}
}
