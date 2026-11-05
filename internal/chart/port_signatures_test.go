// ABOUTME: Tests signatures and signature subs: named and anonymous, parameters, defaults, slurpies and named parameters.
// ABOUTME: Snippets under testdata/port/signatures are recognized through internal/lexer, chart/lex and the recognizer; signatures-refused ones refuse.

package chart

import (
	"os"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// signaturesLines are the perly.y alternatives this family ports, and 1089 (a parameter's name), which is
// excluded as folded into the sigil's token. 1105 and 1109 (a slurpy parameter with a default) and 1546 (an
// anonymous sub with no body) refuse unconditionally in their actions: they stay unmarked, left for the
// issue that completes the port, which marks them and writes their refusal tests.
var signaturesLines = []int{608, 768, 894, 1087, 1089, 1094, 1096, 1100, 1105, 1109, 1115, 1117, 1123, 1133, 1146, 1148,
	1153, 1155, 1163, 1164, 1165, 1171, 1176, 1177, 1181, 1236, 1237, 1241, 1543, 1546}

const feature = "use feature 'signatures';\n"

func TestPortSignatures(t *testing.T) {
	if err := grammar.CheckFamilyMarked(signaturesLines, []int{1105, 1109, 1546}); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "signatures") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

func TestPortSignaturesRefusals(t *testing.T) {
	for _, f := range snippets(t, "signatures-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	for _, src := range []string{"sub f ($x, $y) { 1 }", "sub f ($x, ) { 1 }", "sub f ($x = 1, @r) { 1 }", "my $s = sub ($x) { 1 };"} {
		if v := recognizeSource(t, []byte(feature+src)); !v.accepted {
			t.Errorf("%q is refused at %d; perl accepts it", src, v.at)
		}
	}
	// Forms perl refuses without a syntax error, and the chart refuses too: a signature after the attributes
	// is the grammar's order (perl says "Subroutine attributes must come before the signature"), and
	// perly.y:1105, 1109 and 1546 refuse in their actions, so they have no rule.
	for _, src := range []string{"sub f ($x) : lvalue { 1 }", "my $s = sub ($x) : lvalue { 1 };", "sub f (@r = ()) { 1 }",
		"sub f (@r = 1) { 1 }", "sub f (%h = ()) { 1 }", "my $s = sub;", "my $s = sub :lvalue;",
		"sub f ($x, @r = ()) { 1 }", "sub f ($x, :@n) { 1 }"} {
		if v := recognizeSource(t, []byte(feature+src)); v.accepted {
			t.Errorf("%q is accepted; perl refuses it", src)
		}
	}
}

// A known limit that is a refusal: toke.c makes BEGIN, END and their kin KW_SUB_named_sig under the feature
// too (toke.c:8315, yyl_sub), so perl accepts `END ($x) { 1 }`, and the chart gives those words no
// signature alternative (and the lexer scans neither a prototype nor a signature after them,
// 01a1a44e-8f26-706b-96d4-b3e6d7f55c3f). Fails when the limit goes; move the row to the accepted snippets then.
func TestPortSignaturesPhaserLimit(t *testing.T) {
	if v := recognizeSource(t, []byte(feature+"END ($x) { 1 }")); v.accepted {
		t.Error("END ($x) { 1 } is accepted; the known limit is gone, so move it to the accepted snippets and drop the README line")
	}
}

// TestPortSignaturesKnownLimits pins what the chart accepts and perl refuses, each in its action or in
// toke.c's sigvar state (yyl_sigvar, toke.c:5524), which the chart does not model. Each fails when the
// limit goes; move the row to the refused forms then, and drop the README line.
func TestPortSignaturesKnownLimits(t *testing.T) {
	for _, src := range []string{
		"sub f ($x = ) { 1 }",      // "Optional parameter lacks default expression": `$x` and `$` are one terminal
		"sub f (@r, $y) { 1 }",     // "Slurpy parameter not last"
		"sub f ($x = 1, $y) { 1 }", // "Mandatory parameter follows optional parameter"
		"sub f (@r, @s) { 1 }",     // "Multiple slurpy parameters not allowed"
		"sub f (:$n, $x) { 1 }",    // "Positional parameter follows named parameter"
		// The lexer has no signature-parameter state (yyl_sigvar), v0.1 01a1a903-2d8d-713e-b264-bc8119a6a737:
		"sub f ($x::y) { 1 }",   // yyl_sigvar reads a name without a package; the lexer emits one Variable `$x::y`
		"sub f ($_) { 1 }",      // "Can't use global $_ in subroutine signature"
		"sub f (@_) { 1 }",      // "Can't use global @_ in subroutine signature"
		"sub f ($x += 1) { 1 }", // "Illegal operator following parameter": yyl_sigvar reads only =, //= and ||=
	} {
		if v := recognizeSource(t, []byte(feature+src)); !v.accepted {
			t.Errorf("%q is refused at %d; the known limit is gone, so move it to the refused forms", src, v.at)
		}
	}
}
