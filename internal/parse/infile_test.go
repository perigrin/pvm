// ABOUTME: A `sub NAME` earlier in the same file is a declaration the call site can see.
// ABOUTME: Its pair: a callee with no declaration above it stays Unknown, because perl rejects it too.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestInFileSubDeclarationIsKnown asserts that a sub declared EARLIER in the
// same file makes a parenless call to it parse, with no module involved.
//
// This is perl's own rule and the direction of reading is the whole of it.
// Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'sub ok {} ok 8, 9;'
//	sub ok { } ok 8, 9;              <-- a call
//	$ perl -MO=Deparse -e 'ok 8, 9;'
//	syntax error near "ok 8"         <-- no call at all
//
// The prototype decides the EXTENT, which is why the cases below span the
// three shapes `ShapeOf` distinguishes rather than testing one.
func TestInFileSubDeclarationIsKnown(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{
			// No prototype: a list operator, so `8, 9` is all one call's.
			name: "no prototype takes the list",
			src:  "sub ok { 1 }\nok 8, 9;\n",
		},
		{
			// `($;$)` is one mandatory slot: unary at the call site. The
			// comma then belongs to the enclosing list, and the statement
			// still parses.
			name: "unary prototype",
			src:  "sub ok ($;$) { 1 }\nok 8, 9;\n",
		},
		{
			// `($$$$)` is four mandatory slots, which ShapeOf calls a list
			// operator. This is `opbasic/arith.t`'s `tryeq`, 147 of T2's
			// 316 in-file-declared parenless refusals.
			name: "four slots is a list operator",
			src:  "sub tryeq ($$$$) { 1 }\ntryeq 1, 2, 3, 4;\n",
		},
		{
			// A bodiless forward declaration is still a declaration: perl
			// takes `sub f;` as predeclaring `f`, measured --
			// `perl -e 'sub f; f 1, 2;'` is `syntax OK`.
			name: "a forward declaration counts",
			src:  "sub f;\nf 1, 2;\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if n := countUnknown(parse.Parse([]byte(c.src))); n != 0 {
				t.Errorf("want no Unknown nodes, got %d in:\n%s", n, c.src)
			}
		})
	}
}

// TestUndeclaredParenlessCallStaysUnknown is the other half, and it is the
// half that keeps the fix honest.
//
// perl refuses each of these, so a tree for them would be WRONG in the
// oracle's sense rather than merely wide -- and M1's gate is Oracle WRONG = 0.
// Declining is the correct answer, not a gap.
//
// The second case is the one a whole-file pre-pass would get wrong: the
// declaration is BELOW the call, and perl reads top to bottom, so
// `conformance/mdtest/argument-extent.md`'s third case pins it as a syntax
// error. Measured on 5.42.0:
//
//	$ perl -e 'f 1, 2; sub f { }'
//	Number found where operator expected (Do you need to predeclare "f"?)
func TestUndeclaredParenlessCallStaysUnknown(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{
			name: "no declaration anywhere",
			src:  "ok 8, 9;\n",
		},
		{
			name: "the declaration is below the call site",
			src:  "f 1, 2;\nsub f { 1 }\n",
		},
		{
			// A different name is not a declaration of this one.
			name: "a different sub is declared",
			src:  "sub g { 1 }\nf 1, 2;\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if n := countUnknown(parse.Parse([]byte(c.src))); n == 0 {
				t.Errorf("want the statement to stay Unknown, got a clean "+
					"parse of source perl rejects:\n%s", c.src)
			}
		})
	}
}
