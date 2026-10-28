// ABOUTME: Tests for TreeDiff and TreeDiffShifted: every Node field is compared or excluded by name,
// ABOUTME: a change to any one of them is reported, and a uniform shift between real parses is not.
package parse

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// diffBase is a three-level tree with a heredoc body on its middle node, so
// a mutation of the grandchild is a change deep in the tree.
func diffBase() *Node {
	body := &Node{Kind: Term, Start: 40, End: 44, HeredocBody: true}
	grand := &Node{Kind: Term, Start: 12, End: 14, Text: "$x"}
	child := &Node{Kind: Call, Start: 10, End: 20, Children: []*Node{grand}, InnerHeredocBodies: []*Node{body}}
	return &Node{Kind: SourceFile, Start: 0, End: 50, Children: []*Node{child}}
}

func deepChild(root *Node) *Node { return root.Children[0].Children[0] }

type diffCase struct {
	name   string
	mutate func(root *Node)
}

// unexportedMutations change the unexported fields, which reflection cannot
// set, so each is named. All but nestedParen are set on the root alone.
var unexportedMutations = map[string]func(root *Node){
	"nestedParen": func(r *Node) { deepChild(r).nestedParen = true },
	"loaded":      func(r *Node) { r.loaded = append(r.loaded, "Foo") },
	"declErrs":    func(r *Node) { r.declErrs = append(r.declErrs, errors.New("bad")) },
	"declWarns":   func(r *Node) { r.declWarns = append(r.declWarns, errors.New("odd")) },
	"imports":     func(r *Node) { r.imports = map[string]Import{"f": {Name: "f"}} },
}

// treeDiffCases builds one mutation per Node field, plus child-count and
// deep-child cases. Exported fields are mutated through reflection on the
// deep child, so a new exported field of a known kind is covered at once;
// one of a kind this cannot mutate fails the test.
func treeDiffCases(t *testing.T) []diffCase {
	t.Helper()
	var cases []diffCase
	nt := reflect.TypeFor[Node]()
	for i := range nt.NumField() {
		f := nt.Field(i)
		if _, ok := treeDiffExcluded[f.Name]; ok {
			continue
		}
		if m, ok := unexportedMutations[f.Name]; ok {
			cases = append(cases, diffCase{f.Name, m})
			continue
		}
		if !f.IsExported() {
			t.Fatalf("unexported field %s has no mutation in treeDiffCases", f.Name)
		}
		idx := i
		cases = append(cases, diffCase{f.Name, func(r *Node) {
			v := reflect.ValueOf(deepChild(r)).Elem().Field(idx)
			mutateValue(t, f.Name, v)
		}})
	}
	return append(cases,
		diffCase{"child added", func(r *Node) {
			c := r.Children[0]
			c.Children = append(c.Children, &Node{Kind: Term})
		}},
		diffCase{"child removed", func(r *Node) { r.Children = nil }},
		diffCase{"heredoc body changed", func(r *Node) { r.Children[0].InnerHeredocBodies[0].End++ }},
	)
}

func mutateValue(t *testing.T, name string, v reflect.Value) {
	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(!v.Bool())
	case reflect.Int:
		v.SetInt(v.Int() + 1)
	case reflect.String:
		v.SetString(v.String() + "x")
	case reflect.Array:
		mutateValue(t, name, v.Index(0))
	case reflect.Slice:
		v.Set(reflect.Append(v, reflect.Zero(v.Type().Elem())))
	default:
		t.Fatalf("field %s: cannot mutate a %s; add a case to mutateValue", name, v.Kind())
	}
}

func TestTreeDiffDetectsEveryField(t *testing.T) {
	if d := TreeDiff(diffBase(), diffBase()); d != "" {
		t.Fatalf("two equal trees differ: %s", d)
	}
	for _, c := range treeDiffCases(t) {
		b := diffBase()
		c.mutate(b)
		d := TreeDiff(diffBase(), b)
		if d == "" {
			t.Errorf("%s: change not reported", c.name)
			continue
		}
		if !strings.Contains(c.name, " ") && !strings.Contains(d, c.name) {
			t.Errorf("%s: report does not name the field: %s", c.name, d)
		}
	}

	// Error-valued fields compare by their text, not their identity.
	a, b := diffBase(), diffBase()
	a.declErrs, b.declErrs = []error{errors.New("bad")}, []error{errors.New("bad")}
	if d := TreeDiff(a, b); d != "" {
		t.Errorf("equal error texts reported as different: %s", d)
	}
	b.declErrs = []error{errors.New("worse")}
	if TreeDiff(a, b) == "" {
		t.Error("different error texts not reported")
	}
}

func TestTreeEqualCoversEveryNodeField(t *testing.T) {
	fields := map[string]bool{}
	nt := reflect.TypeFor[Node]()
	for i := range nt.NumField() {
		f := nt.Field(i)
		fields[f.Name] = true
		why, excluded := treeDiffExcluded[f.Name]
		_, named := treeDiffByName[f.Name]
		if excluded && why == "" {
			t.Errorf("excluded field %s carries no reason", f.Name)
		}
		if !f.IsExported() && !named && !excluded {
			t.Errorf("unexported field %s is neither compared by name nor excluded", f.Name)
		}
		if named && excluded {
			t.Errorf("field %s is both compared and excluded", f.Name)
		}
	}
	for name := range treeDiffExcluded {
		if !fields[name] {
			t.Errorf("excluded name %s is not a Node field", name)
		}
	}
	for name := range treeDiffByName {
		if !fields[name] {
			t.Errorf("by-name comparer %s is not a Node field", name)
		}
	}
	for _, name := range treeDiffOffsets {
		if !fields[name] {
			t.Errorf("offset name %s is not a Node field", name)
		}
	}
}

// shiftSrc holds a two-item list, an empty list and a single-item list, so
// InnerStart/InnerEnd are set on some Paren nodes and zero on others.
const shiftSrc = "\nmy @a = (1, 2);\nmy @e = ();\nmy $s = (3);\nsub f { my ($x) = @_; return $x + 1 }\n"

func TestTreeDiffShiftedAcceptsUniformShift(t *testing.T) {
	a, b := Parse([]byte("#"+shiftSrc)), Parse([]byte("#xy"+shiftSrc))
	if len(a.Children) == 0 || len(a.Children) != len(b.Children) {
		t.Fatalf("parses have %d and %d statements", len(a.Children), len(b.Children))
	}
	var inner, parenNoInner bool
	var walk func(*Node)
	walk = func(n *Node) {
		inner = inner || n.InnerEnd != 0
		parenNoInner = parenNoInner || n.Paren && n.InnerEnd == 0
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(a)
	if !inner || !parenNoInner {
		t.Fatalf("source does not exercise both inner-span cases: set=%v paren-unset=%v", inner, parenNoInner)
	}
	// Children[0] is the leading trivia the edit grew; every node after it
	// only moved.
	for i := 1; i < len(a.Children); i++ {
		if d := TreeDiffShifted(a.Children[i], b.Children[i], 2); d != "" {
			t.Errorf("statement %d, delta 2: %s", i, d)
		}
		if d := TreeDiffShifted(b.Children[i], a.Children[i], -2); d != "" {
			t.Errorf("statement %d, delta -2: %s", i, d)
		}
		if TreeDiff(a.Children[i], b.Children[i]) == "" {
			t.Errorf("statement %d: TreeDiff ignores a shift", i)
		}
	}
}

func TestTreeDiffShiftedRejectsNonUniformShift(t *testing.T) {
	// The last statement is compared, not the root: the root starts at 0
	// in both parses, so it differs by more than the shift.
	stmt := func(n *Node) *Node {
		for i := len(n.Children) - 1; ; i-- {
			if len(n.Children[i].Children) > 0 {
				return n.Children[i]
			}
		}
	}
	leaf := func(n *Node) *Node {
		for len(n.Children) > 0 {
			n = n.Children[len(n.Children)-1]
		}
		return n
	}
	for name, mutate := range map[string]func(s *Node){
		"offset shifted by delta+1": func(s *Node) { leaf(s).End++ },
		"non-offset field changed":  func(s *Node) { leaf(s).Text += "x" },
		"child added":               func(s *Node) { s.Children = append(s.Children, &Node{Kind: Term}) },
		"child removed":             func(s *Node) { s.Children = s.Children[:len(s.Children)-1] },
	} {
		a, b := stmt(Parse([]byte("#"+shiftSrc))), stmt(Parse([]byte("#xy"+shiftSrc)))
		if d := TreeDiffShifted(a, b, 2); d != "" {
			t.Fatalf("%s: unmutated statements differ: %s", name, d)
		}
		mutate(b)
		if TreeDiffShifted(a, b, 2) == "" {
			t.Errorf("%s: not reported", name)
		}
	}
	for _, c := range treeDiffCases(t) {
		b := diffBase()
		c.mutate(b)
		if got, want := TreeDiffShifted(diffBase(), b, 0), TreeDiff(diffBase(), b); got != want {
			t.Errorf("%s: delta 0 gives %q, TreeDiff %q", c.name, got, want)
		}
	}
}

// TestTreeDiffPaths pins the exact report text, path and all, for a
// difference at each kind of field and depth TreeDiff walks.
func TestTreeDiffPaths(t *testing.T) {
	for _, c := range []struct {
		name   string
		delta  int
		mutate func(r *Node)
		want   string
	}{
		{"root field", 0, func(r *Node) { r.Text = "x" }, "node.Text:  != x"},
		{"deep exported", 0, func(r *Node) { deepChild(r).Text = "$y" }, "node.Children[0].Children[0].Text: $x != $y"},
		{"deep offset", 0, func(r *Node) { deepChild(r).End = 15 }, "node.Children[0].Children[0].End: 14+0 != 15"},
		{"shifted offset", 3, func(r *Node) {}, "node.Start: 0+3 != 0"},
		{"child count", 0, func(r *Node) { r.Children[0].Children = nil }, "node.Children[0].Children: 1 nodes != 0"},
		{"nil child", 0, func(r *Node) { r.Children[0].Children[0] = nil }, "node.Children[0].Children[0]: nil != non-nil"},
		{"heredoc body", 0, func(r *Node) { r.Children[0].InnerHeredocBodies[0].End = 45 }, "node.Children[0].InnerHeredocBodies[0].End: 44+0 != 45"},
		{"unexported", 0, func(r *Node) { deepChild(r).nestedParen = true }, "node.Children[0].Children[0].nestedParen: false != true"},
	} {
		b := diffBase()
		c.mutate(b)
		if got := TreeDiffShifted(diffBase(), b, c.delta); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
	if got, want := TreeDiff(nil, diffBase()), "node: nil != non-nil"; got != want {
		t.Errorf("nil root: got %q, want %q", got, want)
	}
}

// BenchmarkTreeDiffEdit5k diffs two separately parsed, equal edit5k.pl trees,
// the case every incremental test and fuzz target pays for most.
func BenchmarkTreeDiffEdit5k(b *testing.B) {
	src, err := os.ReadFile(filepath.Join("..", "psc", "testdata", "latency", "edit5k.pl"))
	if err != nil {
		b.Fatal(err)
	}
	x, y := Parse(src), Parse(src)
	b.ResetTimer()
	for range b.N {
		if d := TreeDiff(x, y); d != "" {
			b.Fatal(d)
		}
	}
}
