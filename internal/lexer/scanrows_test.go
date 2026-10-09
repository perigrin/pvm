// ABOUTME: Five scanner rows: literal ranges, qx, <<>>, angle globs, leading ::.
// ABOUTME: Each produced plausible wrong tokens rather than an error, which is the harder kind to notice.

package lexer_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
)

// significant renders a source's tokens, skipping whitespace, as
// `Kind("text")` so a stream can be compared as one string.
func significant(src string) []string {
	var out []string
	for _, tok := range lexer.Tokenize([]byte(src)) {
		if tok.Kind == lexer.Whitespace {
			continue
		}
		out = append(out, tok.Kind.String()+"("+src[tok.Start:tok.End]+")")
	}
	return out
}

func streamIs(src string, want ...string) bool {
	got := significant(src)
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// TestNumericRangeIsNotOneToken: `1..5` is three tokens, not one number.
//
// `scanNumber` consumed `isWordByte` or `.` in a loop, so a literal range
// arrived as `Number("1..5")` and the range operator was gone before the
// parser saw it. `$a..$b` was fine, so only literal endpoints were affected.
//
// The cost is measured in §4.14.2: `my @r = (1..5)` infers `@r` as `Str`,
// because `findOperatorText` (`internal/infer/infer.go:645-659`) matches the
// anonymous `.` before ever seeing `..`, while `..`'s declaration in
// CORE.pmt, `multi sub .. :infix :tighter(=) :assoc(nonassoc) (Str $x, Str $y) List`,
// has List waiting for a token that never arrives. perl:
// `$r[0]+1` is 2, so the elements are numbers.
func TestNumericRangeIsNotOneToken(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{"1..5", []string{"Number(1)", "Operator(..)", "Number(5)"}},
		{"1...5", []string{"Number(1)", "Operator(...)", "Number(5)"}},
		{"0..9", []string{"Number(0)", "Operator(..)", "Number(9)"}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v",
				c.src, significant(c.src), c.want)
		}
	}
}

// TestNumberFormsStillLex is the negative scenario, and the one that makes
// the range fix delicate: a `.` inside a number is part of it, and so is
// every letter in `0x1f` and `1e10`.
//
// Measured on perl 5.42.0:
//
//	$ perl -MO=Deparse -e 'my @n = (1.5, 1_000, 0x1f, 1e10, 0b11, 017);'
//	my(@n) = (1.5, 1000, 31, 10000000000.0, 3, 15);
//
// Six literals, six values. A range fix that split any of these would turn a
// number into an expression.
func TestNumberFormsStillLex(t *testing.T) {
	for _, src := range []string{
		"1.5", "1_000", "0x1f", "1e10", "0b11", "017",
		"1.5e10", "0xdead_beef", "1_000.5",
	} {
		got := significant(src)
		if len(got) != 1 {
			t.Errorf("%q lexes as %v, want one token", src, got)
		}
	}
}

// TestLeadingDecimalLiteral: a numeric literal with no digit before the point
// is ONE number, not an operator and a number.
//
// Measured: `perl -e 'my $n = .5; print $n'` prints 0.5, so `.5` is one
// number. `scanNumber` required a leading DIGIT, so the dot was emitted as
// an operator first and the parser then had no term to start on.
//
// This test was parked as a skip by the range-splitting commit, which was
// right to park it: letting `scanNumber` start on a dot collides with
// concatenation and with `..`, and deciding that inside the range fix
// would have made one number answer for two questions. Issue 01a0c13f is
// that decision, and the rule it settles on is POSITION -- a `.` before a
// digit starts a number only where a TERM is expected, which is the same
// expect-state mechanism the spec uses for a leading `%`.
//
// `TestDotIsStillConcatenation` is the other half and must stay
// green: in operator position the identical bytes are concatenation.
func TestLeadingDecimalLiteral(t *testing.T) {
	for _, c := range []struct {
		src  string
		want string
	}{
		{".5", "Number(.5)"},
		{".25", "Number(.25)"},
		{".5e10", "Number(.5e10)"},
	} {
		got := significant(c.src)
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("%q lexes as %v, want [%s]", c.src, got, c.want)
		}
	}
}

// TestDotIsStillConcatenation is the other half of the rule, and
// the half that makes it a POSITION question rather than a lookahead one.
//
// In OPERATOR position -- just after a term -- a `.` is concatenation
// even when a digit follows, and perl proves the two readings are both
// live. Measured 5.42.0, the same four bytes parse two ways depending on
// nothing but what came before:
//
//	$ perl -MO=Deparse -e 'my $a = "x"; my $r = $a .5;'
//	my $r = $a . '5';                 CONCATENATION
//
//	$ perl -MO=Deparse -e 'my $a = "x"; print $a .5'
//	print $a 0.5;                     FILEHANDLE and a NUMBER
//
// The second is why `print $a .5` prints nothing: `$a` lands in print's
// filehandle slot and `.5` is its argument. That is a runtime failure
// from a lexical decision, and it is exactly the class of bug the expect
// state exists to get right.
//
// So a lexer that scanned `.` as a number wherever a digit follows would
// break the first line, and one that never did would break `.5` alone.
// Only position separates them.
func TestDotIsStillConcatenation(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{`$a . 5`, []string{"Variable($a)", "Operator(.)", "Number(5)"}},
		{`$a .5`, []string{"Variable($a)", "Operator(.)", "Number(5)"}},
		{`$a.5`, []string{"Variable($a)", "Operator(.)", "Number(5)"}},
	} {
		got := significant(c.src)
		if len(got) != len(c.want) {
			t.Errorf("%q lexes as %v, want %v", c.src, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%q lexes as %v, want %v", c.src, got, c.want)
				break
			}
		}
	}
}

// TestLeadingDotVersionStringUnharmed pins the two regressions the
// leading-decimal rule caused, each found by a test rather than foreseen.
//
// A v-string reaches `scanNumber` looking exactly like a leading decimal:
// term expected, a dot, a digit after it. `v5` lexes as an ordinary Word,
// so nothing upstream marks the dots as part of a version.
//
// `use v5.36` is the consequential one. Merging its `.36` into a Number
// leaves `pendingVersionMajor` without the Operator it reassembles the
// version from, the signatures feature never turns on, and `sub g ($a,
// $b)` silently becomes a PROTOTYPE -- a lexical change that alters what
// the rest of the file means.
//
// `require(v5.5.630)` NEEDED a second rule while that run arrived as
// three tokens, because by its third part the pending major was already
// spent. `scanVString` now claims the run whole, so the second rule --
// `continuesVersionString`'s Number branch -- became unreachable and was
// deleted. Measured: deleting it leaves `internal/lexer` and
// `internal/parse` green, and `foo.5.6` still lexes via the Word branch.
//
// The row below therefore checks the FIRST rule and `scanVString`, not
// both rules. The signatures half of what it once asserted lives in
// `TestVStringVersionStillEnablesSignatures`.
//
// NEITHER ROW SPLITS ANY MORE, and the five-token expectation this test
// once pinned for `use v5.36;` BECAME FALSE under 01a0dc26: `scanVString`
// now claims ONE dot as well as two, because perl does -- measured,
// `length(v5.36)` is 2. So the split the comment above describes is
// history in both rows, and what survives of the claim is that no dot of
// a version ever becomes a LEADING DECIMAL: `Number(.36)` would be the
// regression, and one `Quote` is not it. The reassembly in
// `internal/parse/use.go` is likewise no longer reached by these rows --
// measured, it needed no change, because a Quote never matches its
// Word-shaped name path and falls to `parseExpr`, which yields the same
// one Term.
func TestLeadingDotVersionStringUnharmed(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{`use v5.36;`, []string{
			"Word(use)", "Quote(v5.36)", "Semicolon(;)",
		}},
		// `v5.5.630` is now ONE token, and that is the v-string rule of
		// 01a0db78 rather than a loss: two dots make a string, perl agrees
		// (`length(v5.5.630)` is 3), and `scanVString` takes it whole.
		//
		// What this row pinned was that the FIRST dot stay a separator the
		// version reassembly could read. Nothing is reassembled from one
		// token, so `noteSignatures` reads the version out of the Quote
		// directly -- and `TestVStringVersionStillEnablesSignatures` is
		// where that is checked, in both directions. The claim moved; it
		// was not dropped.
		//
		// The FIRST-dot guard this row was added for was still live for the
		// ONE-dot form while `scanVString` declined on one dot. It no longer
		// declines, so nothing in this test reaches the guard: measured, the
		// whole of `internal/lexer` and `internal/parse` stays green with
		// `continuesVersionString` removed from `startsLeadingDecimal`.
		// Deleting it belongs to its own commit, the way its Number branch
		// did.
		{`v5.5.630`, []string{
			"Quote(v5.5.630)",
		}},
	} {
		got := significant(c.src)
		if len(got) != len(c.want) {
			t.Errorf("%q lexes as %v, want %v", c.src, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%q lexes as %v, want %v", c.src, got, c.want)
				break
			}
		}
	}
}

// TestLeadingDotRangeUnharmed keeps the range fix this file's first test
// records from being undone: `..` is the range operator wherever it
// appears, and a `.` that starts a number must not swallow the second
// dot. `1..5` and `.5..1` are both three tokens.
func TestLeadingDotRangeUnharmed(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{`1..5`, []string{"Number(1)", "Operator(..)", "Number(5)"}},
		{`$a..5`, []string{"Variable($a)", "Operator(..)", "Number(5)"}},
	} {
		got := significant(c.src)
		if len(got) != len(c.want) {
			t.Errorf("%q lexes as %v, want %v", c.src, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%q lexes as %v, want %v", c.src, got, c.want)
				break
			}
		}
	}
}

// TestQxLexes: `qx//` is a quote-like operator, and `internal/lexer/quote.go`
// had rows for q, qq, qw, qr, m, s, tr and y but not qx.
//
// §4.14.2 records that a backtick string already lexes as one Quote whose
// Text keeps its delimiters, and that B::SoN calls the form `BacktickExpr` --
// a language fact. `qx//` is the same node reached by a different spelling.
func TestQxLexes(t *testing.T) {
	for _, src := range []string{
		"qx/ls/", "qx{ls}", "qx(ls)", "qx[ls]", "qx!ls!", "qx'ls'",
	} {
		got := significant(src)
		if len(got) != 1 || !strings.HasPrefix(got[0], "Quote(") {
			t.Errorf("%q lexes as %v, want one Quote", src, got)
		}
	}
}

// TestQxIsStillAName is qx's negative: a longer word beginning with `qx` is
// an identifier, the same rule every other quote-op keyword follows.
func TestQxIsStillAName(t *testing.T) {
	for _, src := range []string{
		"$qx", "qxray", "my $qx = 1;", "f(qx => 1);",
	} {
		for _, tok := range lexer.Tokenize([]byte(src)) {
			if tok.Kind == lexer.UnknownRest || tok.Kind == lexer.Error {
				t.Errorf("%q: %v %q -- qx followed by a word character is a name",
					src, tok.Kind, src[tok.Start:tok.End])
			}
		}
	}
}

// TestDoubleDiamond: `<<>>` is the 5.22 readline that treats every argument
// as a filename, never as a command.
//
// It lexed as two shift operators, so a `while (<<>>)` loop read as an
// expression that could not parse.
func TestDoubleDiamond(t *testing.T) {
	got := significant("<<>>")
	if len(got) != 1 || !strings.HasPrefix(got[0], "Readline(") {
		t.Errorf("`<<>>` lexes as %v, want one Readline", got)
	}
}

// TestShiftStillShifts is the double diamond's negative. `<<` in OPERATOR
// position is a left shift, and `<<EOF` is a heredoc; neither may become a
// readline.
func TestShiftStillShifts(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		{"$a << $c", "Operator(<<)"},
		{"$a >> $c", "Operator(>>)"},
	} {
		var found bool
		for _, g := range significant(c.src) {
			if g == c.want {
				found = true
			}
		}
		if !found {
			t.Errorf("%q lexes as %v, want a %s", c.src, significant(c.src), c.want)
		}
	}
	// A heredoc opener must still open a heredoc.
	if got := significant("<<EOF"); len(got) == 0 ||
		!strings.HasPrefix(got[0], "HeredocOpen(") {
		t.Errorf("`<<EOF` lexes as %v, want a HeredocOpen", got)
	}
}

// TestAngleGlob: `<*.c>` is a glob, and `scanAngle` accepted only word
// bytes, `$` and `:` between the angles -- so a `*` ended the candidate and
// the whole form fell through to five separate operators.
func TestAngleGlob(t *testing.T) {
	for _, src := range []string{"<*.c>", "<*>", "<~/*.txt>"} {
		got := significant(src)
		if len(got) != 1 {
			t.Errorf("%q lexes as %v, want one token", src, got)
		}
	}
}

// TestComparisonStillCompares is the glob's negative, and it is the sharper
// half: `<` in operator position is a comparison and must stay one.
func TestComparisonStillCompares(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		{"$a < $c", "Operator(<)"},
		{"$a > $c", "Operator(>)"},
		{"$a <= $c", "Operator(<=)"},
		{"$a <=> $c", "Operator(<=>)"},
	} {
		var found bool
		for _, g := range significant(c.src) {
			if g == c.want {
				found = true
			}
		}
		if !found {
			t.Errorf("%q lexes as %v, want a %s", c.src, significant(c.src), c.want)
		}
	}
	// A readline must still lex: the glob fix must widen, not replace.
	for _, src := range []string{"<STDIN>", "<$fh>", "<>"} {
		got := significant(src)
		if len(got) != 1 || !strings.HasPrefix(got[0], "Readline(") {
			t.Errorf("%q lexes as %v, want one Readline", src, got)
		}
	}
}

// TestLeadingPackageSeparator: `$::x` is `$main::x`.
//
// `scanVarName` reached its identifier scanner, which requires a leading
// identifier byte, and `:` is not one -- so `$::x` lexed as
// `Variable("$:")`, a punctuation variable named `:`, followed by stray
// tokens. That is a plausible wrong answer rather than an error, which is
// why nothing caught it.
//
// Measured: `our $x = 5; print $::x` prints 5.
func TestLeadingPackageSeparator(t *testing.T) {
	for _, src := range []string{
		"$::x", "@::a", "%::h", "$::Foo::x",
	} {
		got := significant(src)
		if len(got) != 1 {
			t.Errorf("%q lexes as %v, want one token", src, got)
		}
	}

	// `&::f` is NOT one token, and that is the existing design rather than a
	// gap here: the lexer emits `&` as its own FuncSigil so that `&f` and
	// `&&` stay distinguishable, and `internal/parse/term.go:97` joins the
	// two. `&foo` is two tokens for the same reason. What matters is that
	// the name after the sigil carries its package separator.
	//
	// It carries it now. This assertion read `len(got) != 3` and a last
	// token of `(f)`, which pinned the sigil's separateness AND the leading
	// `::` splitting off as its own Operator -- and the second half was the
	// defect rather than the design. `scanIdentRunes` would not start a name
	// at a separator, so `&::f()` refused at the parser. The sigil is still
	// its own token, which is the rule this case exists for; the name after
	// it is one Word. See TestLeadingPackageSeparatorInWord.
	if got := significant("&::f"); len(got) != 2 ||
		!strings.HasSuffix(got[len(got)-1], "(::f)") {
		t.Errorf("`&::f` lexes as %v; the sigil is its own token by design, "+
			"and the name after it carries its separator", got)
	}

	// The sigil stays distinguishable from the operators it shares a byte
	// with, which is the reason it is a token of its own.
	for _, c := range []struct {
		src  string
		want []string
	}{
		{"&f", []string{"FuncSigil(&)", "Word(f)"}},
		{"$a && $b", []string{"Variable($a)", "Operator(&&)", "Variable($b)"}},
		{"$a & $b", []string{"Variable($a)", "Operator(&)", "Variable($b)"}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v", c.src, significant(c.src), c.want)
		}
	}
}

// TestPunctuationVariablesStillLexHere is the leading-`::` negative. `$:` is
// a real punctuation variable -- the format line-break characters -- and
// widening the name scanner must not swallow what follows it.
func TestPunctuationVariablesStillLexHere(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{"$: = 1", []string{"Variable($:)", "Operator(=)", "Number(1)"}},
		{"$:", []string{"Variable($:)"}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v",
				c.src, significant(c.src), c.want)
		}
	}
}

// TestWordShapedCompoundAssignment: `x=` is ONE token, and the only compound
// assignment perl spells with a letter.
//
// The other twelve are punctuation, so a scanner that forms compound
// assignment by scanning punctuation runs never reaches this one. Measured
// before the fix, `$t x= 2` lexed as `Word(x) Operator(=)` -- a Word where an
// operator belongs, which the parser reported as trailing_tokens.
//
// The token is a Word, not an Operator, because that is the kind this lexer
// gives every word-shaped operator (`x`, `cmp`, `and`); the glossary's
// `word-shaped operator` category is `Word` plus `parse.IsWordShapedOperator`,
// which reads `x=` out of the precedence table at level 9.
//
// OPERATOR POSITION ONLY, which is perl's own rule and the reason the fat
// comma survives. Measured 5.42.0:
//
//	$ perl -MO=Deparse -e 'my %h = (x=>1);'          my(%h) = ('x', 1);
//	$ perl -MO=Deparse -e 'my $t; my @a=($t x=> 2);' syntax error near "$t x"
//
// In term position `x=>1` is the bareword `x` and a fat comma. In operator
// position perl has already formed `x=` and the `>` is then junk, which is
// what the error says.
func TestWordShapedCompoundAssignment(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		// The operator itself, spaced and unspaced on the right.
		{`$t x= 2`, []string{"Variable($t)", "Word(x=)", "Number(2)"}},
		{`$t x=2`, []string{"Variable($t)", "Word(x=)", "Number(2)"}},

		// Binary `x` stays TWO tokens. Without this the fix could not be
		// told from one that eats any `x`.
		//
		// SPACED only. Unspaced `(1)x3` lexes as `Word(x3)` here, which is a
		// SEPARATE gap -- the identifier scanner takes the digits, and that
		// is true of `x3` with or without this row -- so it is not asserted
		// against a fix for `x=`.
		{`$t x 2`, []string{"Variable($t)", "Word(x)", "Number(2)"}},

		// Greedy by exactly one `=`. Measured, `x==` is a syntax error and
		// `x=~ 3` is `x=` applied to `~3`:
		//
		//	$ perl -e 'my $s="ab"; $s x=~ 3;'  ->  $s x= 18446744073709551612
		//
		// so in both cases perl formed `x=` and read the rest separately.
		{`$t x== 2`, []string{"Variable($t)", "Word(x=)", "Operator(=)",
			"Number(2)"}},
		{`$t x=~ 2`, []string{"Variable($t)", "Word(x=)", "Operator(~)",
			"Number(2)"}},

		// `xx` is not `x`, and the scanner must not form `x=` from the tail
		// of a longer identifier.
		{`$t xx= 2`, []string{"Variable($t)", "Word(xx)", "Operator(=)",
			"Number(2)"}},

		// The name is the variable's, not the operator's: `$tx= 2` is a
		// variable called `$tx`. scanVariable claims it first, and this pins
		// that the fix did not reach back into it.
		{`$tx= 2`, []string{"Variable($tx)", "Operator(=)", "Number(2)"}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v",
				c.src, significant(c.src), c.want)
		}
	}
}

// TestWordShapedCompoundAssignmentTermPosition is the negative: in TERM
// position an `x` followed by `=` is a bareword, never the operator.
//
// This is where the fat comma lives, and it is the case that makes the
// expect-state guard load-bearing rather than decorative.
func TestWordShapedCompoundAssignmentTermPosition(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		// `x=>1` is the autoquoted bareword and a fat comma. A scanner that
		// formed `x=` here would leave `Word(x=) Operator(>)`.
		{`(x=>1)`, []string{"Operator(()", "Word(x)", "Operator(=>)",
			"Number(1)", "CloseBracket())"}},
		{`(x => 1)`, []string{"Operator(()", "Word(x)", "Operator(=>)",
			"Number(1)", "CloseBracket())"}},

		// A hash key and a sub call, both term position.
		{`$h{x}`, []string{"Variable($h)", "Operator({)", "Word(x)",
			"CloseBracket(})"}},
		{`x(1)`, []string{"Word(x)", "Operator(()", "Number(1)",
			"CloseBracket())"}},

		// A variable NAMED x, which the operator must never touch.
		{`$x = 1`, []string{"Variable($x)", "Operator(=)", "Number(1)"}},
		{`$x =~ s/a/b/`, []string{"Variable($x)", "Operator(=~)",
			"Quote(s/a/b/)"}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v",
				c.src, significant(c.src), c.want)
		}
	}
}

// TestFileTestIsOneOperator: `-e` is ONE token, not a minus and a word.
//
// perlop's named-unary level holds the whole family, 27 letters of it, and
// the lexer emitted `Operator(-) Word(e)` for every one. `internal/parse`
// carried a `fileTests` table to glue the pair back together at parse time,
// which built the right tree and left the token stream wrong -- and the token
// stream is what `conformance/mdtest/unary.md`'s "The file-test operators"
// pins with `no operator whose text is "-"`.
//
// THE EXPECT STATE IS NOT THE GUARD HERE, which is where this departs from
// `x=`. Measured 5.42.0, a filetest in what looks like operator position is
// not subtraction, it is a syntax error:
//
//	$ perl -e 'my $x = 1 -e "/etc";'   syntax error near "1 -e "
//	$ perl -e 'my ($a,$b); $a-e$b;'    syntax error near "$a-e"
//
// perl has already formed `-e` in both, and then has nowhere to put it. The
// one spelling that DOES compile only compiles because `print` takes an
// indirect filehandle:
//
//	$ perl -MO=Concise -e 'my ($a,$b); print $a -e $b;'
//	  rv2gv <- $a        (the filehandle)
//	  ftis  <- $b        (the filetest, still one operator)
//
// So `-e` forms wherever it appears and the state is not consulted.
func TestFileTestIsOneOperator(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		// The two spellings the corpus case writes.
		{`-e $f`, []string{"Operator(-e)", "Variable($f)"}},
		{`-e($f)`, []string{"Operator(-e)", "Operator(()", "Variable($f)",
			"CloseBracket())"}},

		// No operand: perl defaults it to `$_`, and the operator is still
		// whole. Measured, `$_="/etc"; print -e;` deparses as `print -e $_`.
		{`-e`, []string{"Operator(-e)"}},

		// Stacked filetests are legal since 5.10 and both are operators.
		// Measured, `-e -f $f` deparses back as itself.
		{`-e -f $f`, []string{"Operator(-e)", "Operator(-f)",
			"Variable($f)"}},

		// A letter from each end of the family, so the table is exercised
		// and not just its first row. All 27 were measured from the optree:
		// a letter is a filetest iff `-L $f` compiles to an `ft*` op.
		{`-M $f`, []string{"Operator(-M)", "Variable($f)"}},
		{`-o $f`, []string{"Operator(-o)", "Variable($f)"}},
		{`-X $f`, []string{"Operator(-X)", "Variable($f)"}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v",
				c.src, significant(c.src), c.want)
		}
	}
}

// TestMinusStillMeansMinus is the filetest's negative, and it carries the
// three readings of `-` that the issue names.
//
// A lexer that takes any letter after a `-` breaks arithmetic; one that takes
// any single letter breaks the method call; one that ignores the fat comma
// breaks the option-name idiom. Each row below is a measurement.
func TestMinusStillMeansMinus(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		// Negation and subtraction, which have no letter to confuse.
		{`-$x`, []string{"Operator(-)", "Variable($x)"}},
		{`-5`, []string{"Operator(-)", "Number(5)"}},
		{`$a - $b`, []string{"Variable($a)", "Operator(-)", "Variable($b)"}},

		// A LONGER word is not a filetest. Measured, `-e1 $f` deparses as
		// `-$f->e1` and `-exists` is the builtin `exists`, negated -- so the
		// letter must not be followed by an identifier character.
		{`-e1 $f`, []string{"Operator(-)", "Word(e1)", "Variable($f)"}},
		{`-exists $h`, []string{"Operator(-)", "Word(exists)",
			"Variable($h)"}},

		// A letter that is NOT in the family. Measured, `-n $f` deparses as
		// `-$f->n`: a method call on the negated scalar, never a filetest.
		{`-n $f`, []string{"Operator(-)", "Word(n)", "Variable($f)"}},

		// WHITESPACE KILLS IT. This is the sharpest boundary, because the
		// bytes `-`, `e`, `$f` are all present. Measured:
		//
		//	$ perl -MO=Deparse -e 'my $f="/etc"; print - e $f;'
		//	print -$f->e;
		//
		// so `- e` is negate(method call) and `-e` is the filetest. The two
		// letters must be adjacent.
		{`- e $f`, []string{"Operator(-)", "Word(e)", "Variable($f)"}},

		// THE FAT COMMA. `-bareword` is the string, and that survives even
		// when the bareword is a filetest letter. Measured:
		//
		//	$ perl -MO=Deparse -e 'my $x = { -e => 1 };'
		//	my $x = {'-e', 1};
		//
		// A `=>` after the letter autoquotes the whole `-e` as a string, so
		// the filetest must decline in front of one.
		{`{ -e => 1 }`, []string{"Operator({)", "Operator(-)", "Word(e)",
			"Operator(=>)", "Number(1)", "CloseBracket(})"}},
		{`-foo => 1`, []string{"Operator(-)", "Word(foo)", "Operator(=>)",
			"Number(1)"}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v",
				c.src, significant(c.src), c.want)
		}
	}
}
