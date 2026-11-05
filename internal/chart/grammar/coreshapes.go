// ABOUTME: Compares the port's builtin token classes with parse.CoreShapes() and lists every difference.
// ABOUTME: Spec §5.3 and §3 rule 9: a fact held twice is checked by a test that lists each disagreement for review.

package grammar

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"tamarou.com/pvm/internal/parse"
)

// builtinTerminals are the terminals a builtin's name is classed under, the ones whose classes are compared.
var builtinTerminals = []string{"FUNC0", "FUNC0OP", "UNIOP", "FUNC1", "LSTOP", "BLKLSTOP"}

// shapeTerminals is the correspondence of parse shapes to token types: niladic to FUNC0, unary to UNIOP or
// FUNC1 (toke.c's FUN1 or UNI, decided by the next token), list to LSTOP. It is a hypothesis, amended once:
// FUNC0OP is FUNC0 for a word that yields an op (toke.c:263 FUN0OP differs from :262 FUN0 only in the token
// and the lvalue it returns; both leave PL_expect at XOPERATOR), so __FILE__ and its kin are niladic. A
// block shape corresponds to nothing, and BLKLSTOP is deliberately not LSTOP: `any {...} @l` is the list
// form with a block first, which a list shape does not say.
var shapeTerminals = map[parse.Shape][]string{
	parse.ShapeNiladic: {"FUNC0", "FUNC0OP"},
	parse.ShapeUnary:   {"UNIOP", "FUNC1"},
	parse.ShapeList:    {"LSTOP"},
}

// CoreShapeDifferences lists every builtin the shapes and the classes disagree on, one line each naming both
// sides: `name<TAB>core: <shape or none><TAB>port: <terminals or none>`, sorted by name. A builtin differs when
// it has a shape and no builtin class, a class and no shape, or any class its shape does not correspond to.
func CoreShapeDifferences(shapes map[string]parse.Shape, classes []ClassDecl) []string {
	port := map[string][]string{}
	for _, d := range classes {
		if d.Key.Text != nil && slices.Contains(builtinTerminals, d.Terminal) && !slices.Contains(port[*d.Key.Text], d.Terminal) {
			port[*d.Key.Text] = append(port[*d.Key.Text], d.Terminal)
		}
	}
	names := map[string]bool{}
	for n := range shapes {
		names[n] = true
	}
	for n := range port {
		names[n] = true
	}
	var diffs []string
	for _, n := range slices.Sorted(maps.Keys(names)) {
		shape, hasShape := shapes[n]
		ts := slices.Sorted(slices.Values(port[n]))
		want := shapeTerminals[shape]
		if hasShape && len(ts) > 0 && !slices.ContainsFunc(ts, func(t string) bool { return !slices.Contains(want, t) }) {
			continue
		}
		core, got := "none", "none"
		if hasShape {
			core = shape.String()
		}
		if len(ts) > 0 {
			got = strings.Join(ts, "/")
		}
		diffs = append(diffs, fmt.Sprintf("%s\tcore: %s\tport: %s", n, core, got))
	}
	return diffs
}
