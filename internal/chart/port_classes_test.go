// ABOUTME: Tests the class, method, field and attribute-list rules: class declarations and blocks, fields, named and anonymous methods.
// ABOUTME: Snippets under testdata/port/classes are recognized through internal/lexer, chart/lex and the recognizer; classes-refused ones refuse.

package chart

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// classesLines are the perly.y alternatives this family ports or fits. 1553 is left for the issue that completes
// the port: its action refuses unconditionally ("Illegal declaration of anonymous subroutine"), so it is no
// accepting rule (refused-in-action.txt). 1077 and 1078 are the attrlist fit's optional form (perly.y:1070).
var classesLines = []int{278, 294, 340, 770, 880, 881, 885, 1034, 1077, 1078, 1548, 1553, 1744, 1749, 1754, 1762, 1769}

func TestPortClasses(t *testing.T) {
	if err := grammar.CheckFamilyMarked(classesLines, []int{1553}); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "classes") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

// There is no refused snippet for an attribute list after a class block (`class Foo { } :isa(Bar)`): neither tokenizer
// marks that colon as an attribute colon, so no rule shape is exercised by it.
func TestPortClassesRefusals(t *testing.T) {
	for _, f := range snippets(t, "classes-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", filepath.Base(f))
		}
	}
	// Perl refuses these other than as a syntax error (an invalid version, an illegal anonymous declaration, a signature
	// error), so they are not snippets. The anonymous-method refusals are 1553's: the rule is not ported, so they refuse.
	const feature = "use feature 'class'; no warnings 'experimental::class'; "
	for _, src := range []string{"class Foo 1.2 field", "class if Bar;", "package if Bar;", "class Foo Bar;", "class Foo Bar { }", "class Foo { method m { my $c = method; } }",
		"class Foo { method ; }", "my sub { 1 }", "class Foo { method m ($y { 1 } }"} {
		if v := recognizeSource(t, []byte(feature+src)); v.accepted {
			t.Errorf("%q is accepted; perl refuses it", src)
		}
	}
}

// TestPortClassesKnownLimits pins what the chart gets wrong against perl -c, each a row that fails once its limit moves.
// The refused rows are refused where perl accepts them; the accepted rows are accepted where perl refuses them, by an action,
// a scope or a signature check the lexer's tokens cannot tell from a legal form.
func TestPortClassesKnownLimits(t *testing.T) {
	const feature = "use feature 'class'; no warnings 'experimental::class'; "
	for src, accepted := range map[string]bool{
		"class Foo v1.2.3;":                                            false, // 01a1a481-ebef: the lexer scans no v-string after a name
		"class Foo { field $x :param //= 1; }":                         false, // 01a1aa23-746a: the lexer scans // after an attribute as a pattern
		"class Foo { method m : { } }":                                 false, // 01a1a445-61f0: the lexer does not mark the { after a bare attribute colon
		"class Bar { } class Foo :isa(Bar 1.2);":                       false, // 01a19fb5: pvm emits ( Word Number ), toke.c one ATTRLIST
		"class Foo { method m :lvalue ($$) { 1 } }":                    true,  // 01a19fb5: tokens ( Variable:$$ ), the same with or without the space
		"class Foo { method m :lvalue (1) { 1 } }":                     true,  // 01a19fb5: tokens ( Number ), as a signature perl refuses
		"class Foo { method m { my $c = method :lvalue (1) { 1 }; } }": true,  // 01a19fb5, likewise
		"field $x;":                         true, // outside a class: waypoint 2's scope refusal
		"method m { 1 }":                    true, // likewise
		"my $c = method { 1 };":             true, // likewise
		"class Foo { field $::x; }":         true, // the field is a package variable
		"class Foo { field $_; }":           true, // the field is a global
		"class Foo :isa(Foo);":              true, // recursive inheritance, found when the class is built
		"class Foo { method BEGIN { 1 } }":  true, // perl dies calling it, not parsing it
		"class Foo { field $x = 1 = 2; }":   true, // can't modify a constant, an action's
		"class Foo { field $x :Foo(bar); }": true, // an unrecognised field attribute, an action's
	} {
		if v := recognizeSource(t, []byte(feature+src)); v.accepted != accepted {
			t.Errorf("%q: the recorded limit moved (accepted=%v)", src, v.accepted)
		}
	}
}

// The word after class and method is forced to a name (toke.c:8379, yyl_sub), so it carries none of scope's call readings
// (forcedWord): a method named foo is no call to foo. Without the entries it also admits FUNC0SUB, NOAMP and UNIOPSUB.
func TestPortClassesForcedName(t *testing.T) {
	for _, tc := range []struct {
		src  string
		at   int
		want []string
	}{
		{"class foo;", 6, []string{"BAREWORD"}},
		{"class Foo { method foo { 1 } }", 19, []string{"BAREWORD", "METHCALL0"}},
		{"sub class { foo(1) }", 12, []string{"PRIVATEREF"}},
		// a sub named class or method keeps its body's readings
		{"sub class { foo 1 }", 12, []string{"BAREWORD", "FUNC0SUB", "NOAMP", "UNIOPSUB"}},
	} {
		if got := terminalsAt(t, tc.src, tc.at); !slices.Equal(got, tc.want) {
			t.Errorf("%q at %d admits %v, want only %v", tc.src, tc.at, got, tc.want)
		}
	}
}
