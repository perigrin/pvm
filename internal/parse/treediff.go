// ABOUTME: TreeDiff compares two trees on every Node field but pointer identity and listed caches.
// ABOUTME: TreeDiffShifted is the same with byte offsets moved by one delta, for subtrees an edit only shifted.
package parse

import (
	"fmt"
	"reflect"
	"slices"
)

// treeDiffExcluded lists the Node fields equality skips, each with why a
// consumer cannot observe it. Only caches and snapshots belong here (token
// cache, lexer and parser checkpoints, a parent back-pointer); the issue that
// adds such a field adds its entry.
var treeDiffExcluded = map[string]string{
	"inc": "Repair's cache of the source bytes and parser snapshots; no accessor reads it",
}

// treeDiffOffsets are the byte-offset fields TreeDiffShifted moves by delta.
// InnerStart and InnerEnd move only where set (InnerEnd != 0): a Paren List
// leaves them 0, and only finishList's single-item copy sets them.
var treeDiffOffsets = []string{"Start", "End", "InnerStart", "InnerEnd"}

// treeDiffByName compares the unexported fields, which reflection can list
// but not read through Interface(). All but nestedParen are set on the root
// alone. Errors compare by their text, not their identity.
var treeDiffByName = map[string]func(a, b *Node) bool{
	"loaded":      func(a, b *Node) bool { return slices.Equal(a.loaded, b.loaded) },
	"declErrs":    func(a, b *Node) bool { return slices.Equal(errorTexts(a.declErrs), errorTexts(b.declErrs)) },
	"declWarns":   func(a, b *Node) bool { return slices.Equal(errorTexts(a.declWarns), errorTexts(b.declWarns)) },
	"imports":     func(a, b *Node) bool { return reflect.DeepEqual(a.imports, b.imports) },
	"nestedParen": func(a, b *Node) bool { return a.nestedParen == b.nestedParen },
}

func errorTexts(errs []error) []string {
	out := make([]string, len(errs))
	for i, e := range errs {
		out[i] = e.Error()
	}
	return out
}

var nodeSliceType = reflect.TypeFor[[]*Node]()

// TreeDiff reports the first difference between a and b, or "" when they are
// equal on every field of every node except pointer identity and the cache
// fields in treeDiffExcluded. Nodes, children and InnerHeredocBodies alike,
// compare by content.
func TreeDiff(a, b *Node) string { return TreeDiffShifted(a, b, 0) }

// TreeDiffShifted reports "" when b equals a with every byte-offset field
// (treeDiffOffsets) increased by delta and every other field equal,
// recursively; otherwise the first difference. With delta 0 it is TreeDiff.
func TreeDiffShifted(a, b *Node, delta int) string { return treeDiff(a, b, delta, "node") }

func treeDiff(a, b *Node, delta int, path string) string {
	if a == nil || b == nil {
		if a == b {
			return ""
		}
		return fmt.Sprintf("%s: nil != non-nil", path)
	}
	va, vb := reflect.ValueOf(a).Elem(), reflect.ValueOf(b).Elem()
	t := va.Type()
	for i := range t.NumField() {
		f := t.Field(i)
		if _, ok := treeDiffExcluded[f.Name]; ok {
			continue
		}
		fa, fb := va.Field(i), vb.Field(i)
		at := path + "." + f.Name
		switch {
		case f.Type == nodeSliceType:
			if fa.Len() != fb.Len() {
				return fmt.Sprintf("%s: %d nodes != %d", at, fa.Len(), fb.Len())
			}
			for j := range fa.Len() {
				if d := treeDiff(fa.Index(j).Interface().(*Node), fb.Index(j).Interface().(*Node), delta, fmt.Sprintf("%s[%d]", at, j)); d != "" {
					return d
				}
			}
		case slices.Contains(treeDiffOffsets, f.Name):
			shift := delta
			if (f.Name == "InnerStart" || f.Name == "InnerEnd") && a.InnerEnd == 0 {
				shift = 0
			}
			if fa.Int()+int64(shift) != fb.Int() {
				return fmt.Sprintf("%s: %d%+d != %d", at, fa.Int(), shift, fb.Int())
			}
		case !f.IsExported():
			eq, ok := treeDiffByName[f.Name]
			if !ok {
				panic("parse.TreeDiff: unexported Node field " + f.Name + " has no comparer")
			}
			if !eq(a, b) {
				return fmt.Sprintf("%s: %v != %v", at, fa, fb)
			}
		default:
			if !reflect.DeepEqual(fa.Interface(), fb.Interface()) {
				return fmt.Sprintf("%s: %v != %v", at, fa, fb)
			}
		}
	}
	return ""
}
