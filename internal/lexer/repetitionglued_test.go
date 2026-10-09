// ABOUTME: Binary `x` glued to its count -- `(1)x3` -- is `x` then a Number, in operator position only.
// ABOUTME: Also holds expect.go's docstring examples to the token streams they claim.

package lexer_test

import (
	"os"
	"strings"
	"testing"
)

// TestRepetitionGluedToNumber: in operator position `x` followed by a digit
// is the repeat operator and a count, whatever follows the `x`. Measured on
// 5.42.0:
//
//	$ perl -MO=Deparse -e 'my @a = (1)x3;'       my(@a) = (1) x 3;
//	$ perl -MO=Deparse -e 'my $s="a"; $s x3;'    $s x 3;
//	$ perl -MO=Deparse -e 'my $s; $s x3_0;'      $s x 30;
//	$ perl -e 'my $s; $s x3abc;'   Missing operator before "abc"
//
// The last is toke.c's `isDIGIT(s[1])` test showing through: only the byte
// after the `x` is examined, so `3abc` is a number and a stray bareword, not
// part of the operator.
//
// `x` is the ONLY word-shaped operator perl splits this way. Measured, each
// of these is `Bareword found where operator expected`:
//
//	$ perl -e 'my $r = 1 cmp3;'
//	$ perl -e 'my $r = 1 lt3;'
//	$ perl -e 'my $r = 1 and3;'
//
// and in term position, or as a name, `x3` is an identifier:
//
//	$ perl -MO=Deparse -e 'my %h = (x3 => 1);'   my(%h) = ('x3', 1);
//	$ perl -MO=Deparse -e 'my $o; $o->x3;'       $o->x3;
//	$ perl -MO=Deparse -e 'my $x3 = 1;'          my $x3 = 1;
func TestRepetitionGluedToNumber(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{`(1)x3`, []string{"Operator(()", "Number(1)", "CloseBracket())",
			"Word(x)", "Number(3)"}},
		{`$s x3`, []string{"Variable($s)", "Word(x)", "Number(3)"}},
		{`"a"x3`, []string{"Quote(\"a\")", "Word(x)", "Number(3)"}},
		{`5x3`, []string{"Number(5)", "Word(x)", "Number(3)"}},

		// Perl's other word-shaped operators are not split: one Word each.
		{`1 cmp3`, []string{"Number(1)", "Word(cmp3)"}},
		{`1 lt3`, []string{"Number(1)", "Word(lt3)"}},
		{`1 and3`, []string{"Number(1)", "Word(and3)"}},

		// Term position, a method name, and a variable's name: identifiers.
		{`(x3 => 1)`, []string{"Operator(()", "Word(x3)", "Operator(=>)",
			"Number(1)", "CloseBracket())"}},
		{`$o->x3`, []string{"Variable($o)", "Operator(->)", "Word(x3)"}},
		{`$x3 = 1`, []string{"Variable($x3)", "Operator(=)", "Number(1)"}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v",
				c.src, significant(c.src), c.want)
		}
	}
}

// TestExpectDocstringExamples holds every example in the Expect type's
// docstring to the reading it names. The table below is keyed by the
// example's source text, and every example line in expect.go must have a
// row, so an example added to the docstring without a check fails here.
func TestExpectDocstringExamples(t *testing.T) {
	rows := map[string][]string{
		`@a / 2`:    {"Variable(@a)", "Operator(/)", "Number(2)"},
		`split /,/`: {"Word(split)", "Quote(/,/)"},
		`(1)x3`: {"Operator(()", "Number(1)", "CloseBracket())", "Word(x)",
			"Number(3)"},
		`x3 => 1`: {"Word(x3)", "Operator(=>)", "Number(1)"},
	}

	src, err := os.ReadFile("expect.go")
	if err != nil {
		t.Fatal(err)
	}
	// The examples are the tab-indented lines of the form
	// `//	<source>   <reading>   -- <why>`.
	seen := 0
	for _, line := range strings.Split(string(src), "\n") {
		body, ok := strings.CutPrefix(line, "//\t")
		if !ok || !strings.Contains(body, " -- ") {
			continue
		}
		example, _, _ := strings.Cut(body, "  ")
		want, ok := rows[example]
		if !ok {
			t.Errorf("expect.go cites %q, which has no row here", example)
			continue
		}
		seen++
		if !streamIs(example, want...) {
			t.Errorf("expect.go cites %q; it lexes as %v, want %v",
				example, significant(example), want)
		}
	}
	if seen != len(rows) {
		t.Errorf("found %d of %d examples in expect.go", seen, len(rows))
	}
}
