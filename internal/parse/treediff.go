// ABOUTME: TreeDiff compares two trees on every Node field but pointer identity and listed caches.
// ABOUTME: TreeDiffShifted is the same with byte offsets moved by one delta, for subtrees an edit only shifted.
package parse

import (
	"fmt"
	"reflect"
	"slices"
	"unsafe"
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

// treeDiffField is how treeDiff compares one Node field, worked out once from
// the lists above. The walk reads fields at their offsets rather than through
// reflect.Value, whose per-field cost alone is most of an equal-trees diff;
// reflection is left to the deep comparisons and to formatting a difference.
type treeDiffField struct {
	index  int
	name   string
	offset uintptr
	how    treeDiffHow
	eq     func(a, b *Node) bool            // byName
	same   func(pa, pb unsafe.Pointer) bool // scalar
}

type treeDiffHow int

const (
	nodes    treeDiffHow = iota // []*Node, compared node by node
	offset                      // int shifted by delta
	inner                       // int shifted by delta only where InnerEnd != 0
	byName                      // unexported, compared by treeDiffByName
	unknown                     // unexported with no comparer: TreeDiff panics on it
	scalar                      // bool, integer or string, or an array of them: compared by same
	nillable                    // slice, map or pointer: equal when both nil, else deep
	deep                        // anything else: reflect.DeepEqual
)

var treeDiffFields = func() []treeDiffField {
	var out []treeDiffField
	t := reflect.TypeFor[Node]()
	for i := range t.NumField() {
		f := t.Field(i)
		if _, ok := treeDiffExcluded[f.Name]; ok {
			continue
		}
		d := treeDiffField{index: i, name: f.Name, offset: f.Offset, how: deep}
		switch k := f.Type.Kind(); {
		case f.Type == reflect.TypeFor[[]*Node]():
			d.how = nodes
		case slices.Contains(treeDiffOffsets, f.Name):
			if k != reflect.Int {
				panic("parse.TreeDiff: offset field " + f.Name + " is not an int")
			}
			d.how = offset
			if f.Name == "InnerStart" || f.Name == "InnerEnd" {
				d.how = inner
			}
		case !f.IsExported():
			d.how = unknown
			if eq, ok := treeDiffByName[f.Name]; ok {
				d.how, d.eq = byName, eq
			}
		case k == reflect.Slice || k == reflect.Map || k == reflect.Pointer:
			d.how = nillable
		default:
			if d.same = sameAt(f.Type, f.Offset); d.same != nil {
				d.how = scalar
			}
		}
		out = append(out, d)
	}
	return out
}()

// sameAt is == on a field of type t at offset off, for the types where ==
// agrees with reflect.DeepEqual: bool, integer and string, compared by their
// bits (a bool is always 0 or 1), and arrays of them. Floats, where the two
// disagree on NaN, get nil, as does every other type.
func sameAt(t reflect.Type, off uintptr) func(pa, pb unsafe.Pointer) bool {
	switch k := t.Kind(); {
	case k == reflect.String:
		return eqAt[string](off)
	case k == reflect.Bool || reflect.Int <= k && k <= reflect.Uintptr:
		switch t.Size() {
		case 1:
			return eqAt[uint8](off)
		case 2:
			return eqAt[uint16](off)
		case 4:
			return eqAt[uint32](off)
		case 8:
			return eqAt[uint64](off)
		}
	case k == reflect.Array:
		el, size, n := sameAt(t.Elem(), 0), t.Elem().Size(), uintptr(t.Len())
		if el == nil {
			return nil
		}
		return func(pa, pb unsafe.Pointer) bool {
			for i := range n {
				if !el(unsafe.Add(pa, off+i*size), unsafe.Add(pb, off+i*size)) {
					return false
				}
			}
			return true
		}
	}
	return nil
}

func eqAt[T comparable](off uintptr) func(pa, pb unsafe.Pointer) bool {
	return func(pa, pb unsafe.Pointer) bool { return *(*T)(unsafe.Add(pa, off)) == *(*T)(unsafe.Add(pb, off)) }
}

func fieldAt[T any](p unsafe.Pointer, off uintptr) T { return *(*T)(unsafe.Add(p, off)) }

// TreeDiff reports the first difference between a and b, or "" when they are
// equal on every field of every node except pointer identity and the cache
// fields in treeDiffExcluded. Nodes, children and InnerHeredocBodies alike,
// compare by content.
func TreeDiff(a, b *Node) string { return TreeDiffShifted(a, b, 0) }

// TreeDiffShifted reports "" when b equals a with every byte-offset field
// (treeDiffOffsets) increased by delta and every other field equal,
// recursively; otherwise the first difference. With delta 0 it is TreeDiff.
func TreeDiffShifted(a, b *Node, delta int) string {
	if d := treeDiff(a, b, delta); d != "" {
		return "node" + d
	}
	return ""
}

// treeDiff reports the first difference with its path relative to a, so the
// path is built only on the way back up from a difference, never for equal fields.
func treeDiff(a, b *Node, delta int) string {
	if a == nil || b == nil {
		if a == b {
			return ""
		}
		return ": nil != non-nil"
	}
	pa, pb := unsafe.Pointer(a), unsafe.Pointer(b)
	for _, f := range treeDiffFields {
		switch f.how {
		case nodes:
			xs, ys := fieldAt[[]*Node](pa, f.offset), fieldAt[[]*Node](pb, f.offset)
			if len(xs) != len(ys) {
				return fmt.Sprintf(".%s: %d nodes != %d", f.name, len(xs), len(ys))
			}
			for j := range xs {
				if d := treeDiff(xs[j], ys[j], delta); d != "" {
					return fmt.Sprintf(".%s[%d]%s", f.name, j, d)
				}
			}
		case offset, inner:
			shift := delta
			if f.how == inner && a.InnerEnd == 0 {
				shift = 0
			}
			if x, y := fieldAt[int](pa, f.offset), fieldAt[int](pb, f.offset); x+shift != y {
				return fmt.Sprintf(".%s: %d%+d != %d", f.name, x, shift, y)
			}
		case unknown:
			panic("parse.TreeDiff: unexported Node field " + f.name + " has no comparer")
		case byName:
			if !f.eq(a, b) {
				return f.differ(a, b)
			}
		case scalar:
			if !f.same(pa, pb) {
				return f.differ(a, b)
			}
		case nillable:
			if fieldAt[unsafe.Pointer](pa, f.offset) == nil && fieldAt[unsafe.Pointer](pb, f.offset) == nil {
				continue // most nodes leave these unset; DeepEqual agrees, only slower
			}
			fallthrough
		default:
			// Through pointers, so the values are compared without being boxed.
			if fa, fb := f.values(a, b); !reflect.DeepEqual(fa.Addr().Interface(), fb.Addr().Interface()) {
				return f.differ(a, b)
			}
		}
	}
	return ""
}

func (f treeDiffField) values(a, b *Node) (reflect.Value, reflect.Value) {
	return reflect.ValueOf(a).Elem().Field(f.index), reflect.ValueOf(b).Elem().Field(f.index)
}

func (f treeDiffField) differ(a, b *Node) string {
	fa, fb := f.values(a, b)
	return fmt.Sprintf(".%s: %v != %v", f.name, fa, fb)
}
