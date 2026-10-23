// ABOUTME: Signature heads the lexer can know without `use feature`: `use experimental` and a class's
// ABOUTME: `method`, whose parenthesised head perl always reads as a signature.

package lexer

import "testing"

// hasPrototype reports whether src lexes any Prototype token.
func hasPrototype(src string) bool {
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind == Prototype {
			return true
		}
	}
	return false
}

// TestExperimentalTurnsSignaturesOn: `use experimental` enables what
// `use feature` would, and `no experimental` disables it. Measured on perl
// 5.44.0, -MO=Concise,-exec,f:
//
//	use experimental "signatures"; sub f (%o) { keys %o }       padhv[%o:783,784]
//	use experimental qw(signatures); sub f (%o) { keys %o }     padhv[%o:783,784]
//	use experimental "signatures"; no experimental "signatures";
//	sub f (%o) { keys %o }                                      rv2hv
func TestExperimentalTurnsSignaturesOn(t *testing.T) {
	for src, want := range map[string]bool{
		"use experimental \"signatures\";\nsub f (%o) { keys %o }\n":                                  false,
		"use experimental qw(signatures);\nsub f (%o) { keys %o }\n":                                  false,
		"use experimental \"signatures\";\nno experimental \"signatures\";\nsub f (%o) { keys %o }\n": true,
	} {
		if got := hasPrototype(src); got != want {
			t.Errorf("%q: Prototype token = %v, want %v", src, got, want)
		}
	}
}

// TestMethodHeadIsASignature: a `method`'s parenthesised head is a signature
// under the class feature, signatures feature or not. Measured on perl
// 5.44.0, -MO=Concise,-exec,C::m:
//
//	use feature "class"; class C { method m (%o) { keys %o } }       padhv[%o:5,6]
//	use experimental "class"; class C { method m (%o) { keys %o } }  padhv[%o:785,786]
//
// and `use feature "class"; class C { method m ($x) { $x } }` called with two
// arguments dies "Too many arguments for subroutine 'C::m'".
func TestMethodHeadIsASignature(t *testing.T) {
	for _, src := range []string{
		"use feature \"class\";\nclass C { method m (%o) { keys %o } }\n",
		"use experimental \"class\";\nclass C { method m (%o) { keys %o } }\n",
		"use feature \"class\";\nclass C { method m ($x) { $x } }\n",
	} {
		if hasPrototype(src) {
			t.Errorf("%q: a method's head lexed as a Prototype", src)
		}
	}
	// A sub beside it keeps its prototype: the feature is still off.
	if !hasPrototype("use feature \"class\";\nclass C { sub s ($$) { 1 } }\n") {
		t.Errorf("a sub in a class lost its prototype")
	}
}
