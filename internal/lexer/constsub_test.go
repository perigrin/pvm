// ABOUTME: A name declared with the empty prototype -- `use constant X => 2`,
// ABOUTME: `sub Z () { 3 }` -- has produced a value, so `/` after it divides.

package lexer

import "testing"

// TestDeclaredConstantDivides: perl's lexer finds the constant sub in its
// symbol table and expects an operator after it. Measured on 5.42.0,
//
//	$ perl -MO=Deparse -e 'sub Z () { 3 } print Z / 2, "\n"; use constant X => 2;
//	      print X / 2; use constant { A => 1 }; print A / 1;'
//	print 1.5, "\n";  print 1.0;  print 1.0;
//
// An undeclared word still leaves a term expected: `f / 2/` is a match.
// PerlOnJava unit/constant.t:60.
func TestDeclaredConstantDivides(t *testing.T) {
	for _, src := range []string{
		"use constant X => 2;\nprint X / 2, 1 / 2;\n",
		"use constant 'Y' => 4;\nprint Y / 2, 1 / 2;\n",
		"use constant { A => 1, B => 2 };\nprint B / 1, 1 / 2;\n",
		"sub Z () { 3 }\nprint Z / 2, 1 / 2;\n",
	} {
		for _, tok := range Tokenize([]byte(src)) {
			if tok.Kind == Quote && src[tok.Start] == '/' || tok.Kind == UnknownRest {
				t.Errorf("%q: %v %q; the `/` after a constant divides", src, tok.Kind, src[tok.Start:tok.End])
			}
		}
	}
	if _, ok := firstOfKind("print f / 2/;\n", Quote); !ok {
		t.Errorf("`f / 2/` after an undeclared word is a match")
	}
}
