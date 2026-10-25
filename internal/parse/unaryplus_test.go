// ABOUTME: Canon writes a unary operator back as its source token, never its perl op name.
// ABOUTME: `+$x` emitted as `pos $x` is perl's pos builtin, a different program.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCanonUnaryPlusIsPlus: the tree names unary `+` by its operation, "pos",
// and canon must spell that back as `+`.
//
// Measured on 5.42.0 with -MO=Deparse: `+$x` deparses to `$x` -- unary plus
// is a no-op -- while `pos $x` deparses to `pos $x`, the match-position
// builtin. Writing the op name is a different program, and a re-parse reads
// it as a call, so canon is not a fixpoint either.
//
// `-` and `\` are here because they share the mapping: their op names "neg"
// and "ref" also re-parse silently (`neg $x` deparses to `$x->neg`).
func TestCanonUnaryPlusIsPlus(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"my $x; +$x;", "my $x;+$x;"},
		{"my $x; -$x;", "my $x;-$x;"},
		{`my $x; \$x;`, `my $x;\$x;`},
	} {
		got := strings.TrimSpace(parse.Canon(parse.Parse([]byte(tc.src)), []byte(tc.src)))
		if got != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, got, tc.want)
		}
		again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
		if again != got {
			t.Errorf("canon is not a fixpoint on %q:\n  once  %q\n  twice %q", tc.src, got, again)
		}
	}
}

// TestCanonNestedMinusIsNotDecrement is the parser fuzzer's `- -!00`. Two
// prefix minuses written with nothing between them are `--`, the
// predecrement: measured on 5.42.0, `print --1` is "Can't modify constant
// item in predecrement (--)". perl's Deparse writes `- -$x` as `-(-$x)` and
// `- --$x` as `---$x`, so canon parenthesises an operand whose first byte
// would join the operator.
func TestCanonNestedMinusIsNotDecrement(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"my $x; - -$x;", "my $x;-(-$x);"},
		{"my $x; - --$x;", "my $x;-(--$x);"},
		{"my $x; + +$x;", "my $x;+(+$x);"},
		{"my $x; - +$x;", "my $x;-+$x;"},
		{"- -!0;", "-(-!0);"},
	} {
		got := strings.TrimSpace(parse.Canon(parse.Parse([]byte(tc.src)), []byte(tc.src)))
		if got != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, got, tc.want)
		}
		again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
		if again != got {
			t.Errorf("canon is not a fixpoint on %q:\n  once  %q\n  twice %q", tc.src, got, again)
		}
	}
}

// TestCanonNotOfComplementIsNotNegatedMatch is the parser fuzzer's `! ~&A`.
// A `!` written against a `~` is `!~`, the negated binding operator:
// measured on 5.42.0, `! ~$x` compiles as not-of-complement and `!~$x` is
// "syntax error ... near "!~"". perl's own Deparse writes `!~$x` there,
// which does not re-read, so canon parenthesises the operand instead, as it
// does for `- -$x`.
func TestCanonNotOfComplementIsNotNegatedMatch(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"! ~&A", "!(~&A)"},
		{"my $x; ! ~$x;", "my $x;!(~$x);"},
		{"my $x; ! ~~$x;", "my $x;!(~~$x);"},
		{"my $x; !!$x;", "my $x;!!$x;"},
	} {
		got := strings.TrimSpace(parse.Canon(parse.Parse([]byte(tc.src)), []byte(tc.src)))
		if got != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, got, tc.want)
		}
		again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
		if again != got {
			t.Errorf("canon is not a fixpoint on %q:\n  once  %q\n  twice %q", tc.src, got, again)
		}
	}
}

// TestCanonMinusBeforeAFiletestLetter is the parser fuzzer's `--- A`. A `-`
// written against a one-letter name is a filetest: measured on 5.42.0 with
// -MO=Deparse,-p, `print - A()` is `print((-(A())))` but `print -A()` is
// `print((-A $_))`. Deparse writes the negation as `-(A())`, and so does
// canon. A longer name joins nothing: `print -foo()` is `print((-foo()))`.
func TestCanonMinusBeforeAFiletestLetter(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"- A();", "-(A());"},
		{"- f(1);", "-(f(1));"},
		{"- foo();", "-foo();"},
	} {
		got := strings.TrimSpace(parse.Canon(parse.Parse([]byte(tc.src)), []byte(tc.src)))
		if got != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, got, tc.want)
		}
		again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
		if again != got {
			t.Errorf("canon is not a fixpoint on %q:\n  once  %q\n  twice %q", tc.src, got, again)
		}
	}
}
