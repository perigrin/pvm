// ABOUTME: Tests the traditional subs: named and anonymous, prototypes, attributes, phasers and the ->& call.
// ABOUTME: Snippets under testdata/port/subs are recognized through internal/lexer, chart/lex and the recognizer; subs-refused ones refuse.

package chart

import (
	"os"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// subsLines are the perly.y alternatives this family ports or fits, and 1566 (term: anonymous) as
// written. 1713 (term: listop) is written in the grammar for reachability and is list operators' line. 1541 (an anonymous sub with no body) refuses
// unconditionally in its action: it is left to the issue that completes the port, which marks it
// excluded refused-in-action and holds its refusal in TestLedgerRefusals. Attribute arguments and names carry the attrlist limits
// (01a19fb5-79d7-76c5-a8b1-25fd04ddfb4a): an argument is one token, so `:Foo(a, b)` is refused.
// An anonymous sub's prototype, and an attribute colon with no attribute after it, leave
// internal/lexer's `{` outside a block, so `sub ($$) { 1 }` and `sub foo : { 1 }` are refused until
// 01a1a445-61f0-73c8-81db-1cc65a3620f5: no snippet carries them.
var subsLines = []int{574, 638, 893, 895, 1029, 1045, 1051, 1052, 1057, 1058, 1065, 1218, 1219, 1224, 1300, 1306,
	1538, 1541, 1566}

func TestPortSubs(t *testing.T) {
	if err := grammar.CheckFamilyMarked(subsLines, []int{1541}); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "subs") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
	// ADJUST outside a class is perl's scope refusal (waypoint 2's), so no snippet can hold it.
	for _, src := range []string{"ADJUST { 1 }", "ADJUST;", "sub foo { 1 }", "sub foo;", "sub foo ($$) { 1 }", "sub foo :lvalue { 1 }",
		"BEGIN { 1 }", "my $s = sub { 1 };", "sub BEGIN { 1 }", "$y = $x->&foo"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d", src, v.at)
		}
	}
}

func TestPortSubsRefusals(t *testing.T) {
	for _, f := range snippets(t, "subs-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	// Known limit: the lexer scans a prototype only after sub NAME, so perl's `BEGIN ($$) { 1 }` is refused
	// until 01a1a44e-8f26-706b-96d4-b3e6d7f55c3f. This fails when the limit goes; move it to the accepts then.
	if src := "BEGIN ($$) { 1 }"; recognizeSource(t, []byte(src)).accepted {
		t.Errorf("%q is accepted; the known limit is gone, so move it to the accepted forms and drop the README line", src)
	}
	// Forms perl refuses and the chart does too, for a reason of its own: a sub's attribute argument is
	// one token until 01a19fb5-79d7-76c5-a8b1-25fd04ddfb4a, so perl's `:Foo(a, b)` is refused.
	for _, src := range []string{"sub foo :Foo(a, b) { 1 }"} {
		if v := recognizeSource(t, []byte(src)); v.accepted {
			t.Errorf("%q is accepted; the chart refuses it as a known limit", src)
		}
	}
}
