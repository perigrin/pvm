// ABOUTME: Measures whether our parser parses the way perl parses, not just without errors.
// ABOUTME: Ground truth comes from perl's own optree, so a disagreement is a fact, not an opinion.

package parseoracle_test

import (
	"context"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
	"tamarou.com/pvm/internal/parser"
)

// askPerl compiles src and returns perl's parse facts. Perl reports how it
// parsed something, which is what makes fidelity measurable at all.
func askPerl(t *testing.T, src string) parseoracle.Facts {
	t.Helper()

	facts, err := parseoracle.Ask(context.Background(), []byte(src), parseoracle.Options{})
	if err != nil {
		t.Skipf("perl unavailable or oracle failed: %v", err)
	}
	return facts
}

// TestOracleDiscriminatesPrototypes is the oracle's own self-test. A prototype
// changes how a call PARSES — `f(@a)` passes a reference under `sub f(\@)` and
// a flattened list without it — and perl proves it with an srefgen op that is
// simply absent in the second case. If this ever stops discriminating, every
// fidelity measurement below is meaningless.
func TestOracleDiscriminatesPrototypes(t *testing.T) {
	with := askPerl(t, "sub f(\\@){}\nmy @a;\nf(@a);\n")
	without := askPerl(t, "sub f{}\nmy @a;\nf(@a);\n")

	if !with.OK || !without.OK {
		t.Fatalf("both probes must compile: with=%v (%s) without=%v (%s)",
			with.OK, with.Stderr, without.OK, without.Stderr)
	}
	if with.Srefgen == 0 {
		t.Errorf("prototype \\@ must produce srefgen, got ops %v", with.Ops)
	}
	if without.Srefgen != 0 {
		t.Errorf("no prototype must produce no srefgen, got ops %v", without.Ops)
	}
	if got := with.Prototypes["f"]; got != `\@` {
		t.Errorf("prototype of f = %q, want %q", got, `\@`)
	}
}

// TestPrototypeParsesAreIndistinguishableToUs records the fidelity gap that
// motivates the specification: the two programs above parse DIFFERENTLY in
// perl, and our tree cannot tell them apart, because a static parser has no
// prototype table. Neither tree has an error node, so coverage reports both as
// a success.
//
// This is a characterisation test. It documents what is true today rather than
// asserting what should be true, and it is the measurement a new parser has to
// improve on.
func TestPrototypeParsesAreIndistinguishableToUs(t *testing.T) {
	const withProto = "sub f(\\@){}\nmy @a;\nf(@a);\n"
	const noProto = "sub f{}\nmy @a;\nf(@a);\n"

	p := parser.New()

	shapeOf := func(src string) string {
		tree, err := p.Parse([]byte(src))
		if err != nil {
			t.Fatalf("parse %q: %v", src, err)
		}
		root := tree.RootNode()
		if root == nil {
			t.Fatalf("no root for %q", src)
		}
		if root.HasError() {
			t.Fatalf("unexpected parse error for %q", src)
		}
		// Compare only the call statement, since the sub definition differs by
		// the prototype text itself. The question is whether the CALL differs.
		return callShape(root)
	}

	if shapeOf(withProto) != shapeOf(noProto) {
		t.Log("parser now distinguishes prototype call sites — update this test")
		return
	}

	t.Log("KNOWN GAP: `f(@a)` has an identical tree with and without " +
		"`sub f(\\@)`, but perl parses them differently (srefgen present " +
		"only with the prototype). Both trees are error-free, so coverage " +
		"scores this as a success. Fidelity is what catches it.")
}

// callShape renders the shape of the last statement, which in these probes is
// the call being measured.
func callShape(root *parser.Node) string {
	var last *parser.Node
	for i := 0; i < root.NamedChildCount(); i++ {
		last = root.NamedChild(i)
	}
	if last == nil {
		return ""
	}
	var b strings.Builder
	render(last, &b)
	return b.String()
}

func render(n *parser.Node, b *strings.Builder) {
	if n == nil {
		return
	}
	b.WriteString("(")
	b.WriteString(n.Kind())
	for i := 0; i < n.NamedChildCount(); i++ {
		b.WriteString(" ")
		render(n.NamedChild(i), b)
	}
	b.WriteString(")")
}
