// ABOUTME: Tests the loops: for and foreach in their C-style and list forms, iteration variables, given, when and default.
// ABOUTME: Snippets under testdata/port/loops are recognized through internal/lexer, chart/lex and the recognizer; loops-refused ones refuse.

package chart

import (
	"os"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// loopsLines are the perly.y alternatives this family ports. None is left for a completing issue:
// given, when and default recognize here; when and default outside a topicalizer are parity, as perl -c accepts both under the switch feature.
var loopsLines = []int{317, 348, 379, 392, 411, 423, 452, 474, 504, 732, 882, 886, 888, 900, 993, 994, 1015,
	1803, 1808, 1810, 1814, 1818, 1819, 1825, 1829, 1830, 1831, 1834, 1835, 1838, 1839}

func TestPortLoops(t *testing.T) {
	if err := grammar.CheckFamilyMarked(loopsLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "loops") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

func TestPortLoopsRefusals(t *testing.T) {
	for _, f := range snippets(t, "loops-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	// toke.c croaks "Missing $ on loop variable" for a parenthesised list after our or state
	// at a statement start; perl's message is not a syntax error, so these are not snippets.
	for _, src := range []string{"for our ($a, $b) (@l) { 1 }", "for state ($a) (@l) { 1 }", "foreach our ($a) (@l) { 1 }",
		"{ for our ($a) (@l) { 1 } }", "1; for our ($a) (@l) { 1 }", "if (1) {} for our ($a) (@l) { 1 }",
		"L: for our ($a) (@l) { 1 }", "L: for state ($a) (@l) { 1 }", "A: B: for our ($a) (@l) { 1 }", "{ L: foreach our ($a) (@l) { 1 } }"} {
		if v := recognizeSource(t, []byte(src)); v.accepted {
			t.Errorf("%q is accepted; perl croaks", src)
		}
	}
	for _, src := range []string{"for my $x (@l) { 1 }", "1 for our ($x);", "$i++ for our ($x);", "for our $x (@l) { 1 }",
		"for state $x (@l) { 1 }", "for my ($a) (@l) { 1 }",
		"my $y : for our ($x);", "L: for my $x (@l) { 1 }", "L: 1 for our ($x);", "L: my $y : for our ($x);", "L:\nmy ($y) : for our ($x);", "my ($y) : for our ($x);", "our $y : foreach our ($x);", "my $y : for state ($x);", "our ($x);", "state ($x) = 1;", "for (my $i = 0; $i < 3; $i++) { 1 }", "for (;;) { last }", "1 for @l;",
		"for (;;) { 1 }", "for (@l) { 1 }", "foreach $x (@l) { 1 } continue { 2 }", "$x = 1 for @l;"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d", src, v.at)
		}
	}
}
