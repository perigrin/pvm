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
	// A quote is the same disagreement as `#`: measured on 5.42.0, `A*"3"`,
	// `A*'3'` and A*`ls` are `'A' * 3` and `'A' * \`ls\``, a string to perl
	// and the glob `*"` to the lexer. FuzzParser's `A{} (A*" )` canon'd as
	// `A {}(A * ")`, which re-read as an unterminated string.
	for _, src := range []string{"fo*#", "A*#", "A::B*#", "x*#", "fo*#\n", "fo*#\n1;",
		`A{} (A*" )`, `A*"3";`, "A*'3';", "A*`ls`;",
		// And the glob `**`: `x***x` is `x ** *x` to perl 5.42.0 ("Operator
		// or semicolon missing before *x"), `x * ** * x` to the parser.
		// FuzzParser's `*A*x***x****$0` canon'd past a fixpoint.
		"*A*x***x****$0", "x***x;",
		// And the globs `*;` and `*}`, which end the operand perl reads
		// instead: `A*;` and `{A*}` are syntax errors on 5.42.0. The
		// merge-commit campaign's `A*}x0` canon'd as `A * } x 0`, where
		// the `}` ends the operator's missing operand (R-g).
		"A*}x0", "A*;", "{A*}"} {
		found := collectUnknown(parse.Parse([]byte(src)))
		if len(found) == 0 || found[0].Refusal != parse.AmbiguousSigil {
			t.Errorf("%q: want an ambiguous_sigil refusal, got %v", src, refusalCodes(found))
		}
	}
	// `*` glued to punctuation other than `#` or a quote multiplies,
	// measured with -MO=Deparse on 5.42.0, and only those have a measured
	// fixpoint failure.
	for _, src := range []string{
		"sub fo {} fo*#\n;", "A * 2;", "A*_;",
		"foo*$;;", "foo*(1);", "foo*-1;", `foo*\$x;`, "foo*!$x;", "foo*~$x;",
		"foo*+1;", "foo*[1];", "foo**2;", "A ** 2;", "sub A {} A*;", "use POSIX; DBL_MAX*-1;", "$x = $y ? FOO*(2) : 0;",
	} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) != 0 {
			t.Errorf("%q: perl compiles this, but it refuses: %v", src, refusalCodes(found))
		} else if ok, why := canonMatches([]byte(src)); !ok {
			t.Errorf("%q: canon is not a fixpoint: %s", src, why)
		}
	}
}
