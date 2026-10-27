// ABOUTME: FuzzParser's `fo*#` class: after an undeclared word perl's `*` multiplies and `#` starts a comment.
// ABOUTME: The lexer read the glob `*#`, so the parser refuses the disagreement with ambiguous_sigil.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCanonFixpointOperatorBeforeComment: `fo*#` was read as `fo() * *#`,
// so canon wrote `fo * #` and that re-read as a comment. Measured on 5.42.0, perl -c rejects every input here
// (syntax error at EOF), so each must be refused with a code -- or, should a
// later parser accept one, re-emit as a fixpoint.
func TestCanonFixpointOperatorBeforeComment(t *testing.T) {
	for _, src := range []string{"fo*#", "A*#", "A::B*#", "x*#", "fo*#\n", "$x*#\n", "fo**#"} {
		checkParseInvariants(t, src)
		root := parse.Parse([]byte(src))
		if u := firstUnknown(root); u != nil {
			if u.Refusal == "" {
				t.Errorf("%q: refused with no code", src)
			}
			continue
		}
		if ok, why := canonMatches([]byte(src)); !ok {
			t.Errorf("%q: canon is not a fixpoint: %s", src, why)
		}
	}
}

// TestUndeclaredWordStarBeforePunctuationRefuses: after an undeclared word
// perl's `*` multiplies, so in `fo*#` the `#` starts a comment, while the
// lexer, with no symbol table, read the glob `*#`. Measured on 5.42.0 with
// -MO=Deparse: `fo*#` + newline + `1;` is `'fo' * 1`, and `sub fo {} fo*#`
// + newline + `;` is `fo *#;`. The parser cannot tell which token stream
// is perl's, so it refuses the undeclared case -- even the one perl
// compiles -- and the declared one keeps its glob argument.
func TestUndeclaredWordStarBeforePunctuationRefuses(t *testing.T) {
	for _, src := range []string{"fo*#", "A*#", "A::B*#", "x*#", "fo*#\n", "fo*#\n1;"} {
		found := collectUnknown(parse.Parse([]byte(src)))
		if len(found) == 0 || found[0].Refusal != parse.AmbiguousSigil {
			t.Errorf("%q: want an ambiguous_sigil refusal, got %v", src, refusalCodes(found))
		}
	}
	// `*` glued to punctuation other than `#` multiplies, measured with
	// -MO=Deparse on 5.42.0, and only `#` has a measured fixpoint failure.
	for _, src := range []string{
		"sub fo {} fo*#\n;", "A * 2;", "A*_;",
		"foo*$;;", "foo*(1);", "foo*-1;", `foo*\$x;`, "foo*!$x;", "foo*~$x;",
		"foo*+1;", "foo*[1];", "use POSIX; DBL_MAX*-1;", "$x = $y ? FOO*(2) : 0;",
	} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) != 0 {
			t.Errorf("%q: perl compiles this, but it refuses: %v", src, refusalCodes(found))
		} else if ok, why := canonMatches([]byte(src)); !ok {
			t.Errorf("%q: canon is not a fixpoint: %s", src, why)
		}
	}
}
