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

// TestSignaturesBundleStartsAt535: the signatures feature is in perl's 5.35
// bundle, so `use 5.035` turns it on and `use 5.034` does not. Measured on
// perl 5.44.0, each before `sub f (%o) { keys %o }`, -MO=Concise,-exec,f:
//
//	use 5.035      padhv[%o:4,5]      use 5.034      rv2hv
//	use v5.35.0    padhv[%o:4,5]      use v5.34.0    rv2hv
//	use 5.035_001  padhv[%o:4,5]
//	use feature ":5.35"   padhv       use feature ":5.34"   rv2hv
//
// and feature.pm's own table agrees: $feature::feature_bundle{"5.35"} names
// signatures, {"5.34"} does not.
func TestSignaturesBundleStartsAt535(t *testing.T) {
	for v, want := range map[string]bool{
		"use 5.035;": true, "use v5.35.0;": true, "use 5.035_001;": true, "use feature \":5.35\";": true,
		"use 5.034;": false, "use v5.34.0;": false, "use feature \":5.34\";": false,
	} {
		src := v + "\nsub f (%o) { keys %o }\n"
		if got := !hasPrototype(src); got != want {
			t.Errorf("%q: signature = %v, want %v", src, got, want)
		}
	}
}
