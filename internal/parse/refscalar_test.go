// ABOUTME: Tests that a \$ prototype slot refuses what perl refuses there: constants, sub results, aggregates, wrong arity.
// ABOUTME: Each case was measured with perl 5.42.0's `perl -c`; scalar lvalues perl accepts stay parsed.
package parse

import (
	"testing"
)

// refScalarPrelude declares what every case below calls: a plain sub f, the
// `\$`-prototyped sref, and the variables the cases pass.
const refScalarPrelude = "sub f {} sub sref (\\$); my ($x, $y, %h, @a, $r); "

// refusedAs returns the refusal codes of every Unknown in a parse of src.
func refusedAs(src string) []RefusalCode {
	var codes []RefusalCode
	var walk func(*Node)
	walk = func(n *Node) {
		if n.Kind == Unknown {
			codes = append(codes, n.Refusal)
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(Parse([]byte(src)))
	return codes
}

// wantRefScalarRefusal wants each call refused, once, with the `\$` slot's
// own code.
func wantRefScalarRefusal(t *testing.T, calls ...string) {
	t.Helper()
	for _, call := range calls {
		src := refScalarPrelude + call + ";\n"
		if got := refusedAs(src); len(got) != 1 || got[0] != "ref_scalar_slot" {
			t.Errorf("%s: refusals %v, want one ref_scalar_slot", call, got)
		}
	}
}

// wantNoRefusal wants each source parsed with no Unknown at all.
func wantNoRefusal(t *testing.T, srcs ...string) {
	t.Helper()
	for _, src := range srcs {
		if got := refusedAs(src); len(got) > 0 {
			t.Errorf("%q: refused %v; perl compiles it", src, got)
		}
	}
}

// TestRefScalarSlotRefusesConstants: RFC 0001 "Refusing what perl refuses".
// A `\$` slot takes a scalar lvalue. Measured on 5.42.0 with `perl -c`:
// `sref(1)` and `sref("s")` die "Type of arg 1 to main::sref must be scalar
// (not constant item)", and `sref(f())` "(not subroutine entry)", while
// `sref($x)`, `sref($h{k})` and `sref($x = 7)` compile.
func TestRefScalarSlotRefusesConstants(t *testing.T) {
	wantRefScalarRefusal(t, "sref(1)", `sref("s")`, "sref(f())")
	wantNoRefusal(t,
		refScalarPrelude+"sref($x);\n",
		refScalarPrelude+"sref($h{k});\n",
		refScalarPrelude+"sref($x = 7);\n",
	)
}

// TestRefScalarSlotRefusesAggregatesAndArity: measured on 5.42.0, `sref(@a)`
// dies "(not private array)", `sref(%h)` "(not private hash)", `sref(@$r)`
// "(not array dereference)" and `sref(my @q)` "(not private array)";
// `sref()` and a bare `sref` die "Not enough arguments for main::sref" and
// `sref($x, $y)` "Too many arguments". `sref(my $z)` and `sref(substr($x,
// 0, 1))` compile, as do a slice, `sref(@a[0])`, a parenless `sref $x, $y`
// whose comma is the enclosing list's, an optional slot left empty,
// `sopt()` under `(;\$)`, and comp/proto.t's `sreftest my $a = 'quidgley',
// $i++` under `(\$$)`.
func TestRefScalarSlotRefusesAggregatesAndArity(t *testing.T) {
	wantRefScalarRefusal(t, "sref(@a)", "sref(%h)", "sref(@$r)", "sref(%{$r})",
		"sref(my @q)", "sref(our %q)", "sref()", "sref", "sref($x, $y)")
	wantNoRefusal(t,
		refScalarPrelude+"sref(my $z);\n",
		refScalarPrelude+"sref(substr($x, 0, 1));\n",
		refScalarPrelude+"sref(@a[0]);\n",
		refScalarPrelude+"sref $x, $y;\n",
		"sub sopt (;\\$); my $x; sopt(); sopt($x);\n",
		"sub stwo (\\$$); my $x; stwo($x, 1);\n",
		// comp/proto.t:623: the comma after an initialiser is the call's.
		"sub stwo (\\$$); my $i; stwo my $a = 'quidgley', $i++;\n",
	)
}
