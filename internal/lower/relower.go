// ABOUTME: Relower: an edited file's lowering from the one before it, re-lowering only the top-level subs an edit damaged.
// ABOUTME: The rest of the old tree is shared; only the path from the File down to each replaced Decl is copied.

package lower

import (
	"reflect"
	"strconv"

	"tamarou.com/pvm/internal/parse"
)

// Relower is Lower(root) for root, the tree of an edited buffer, given f,
// the lowering of the tree before the edit, and anchors, the top-level subs
// (named subs not inside another, as AssignIDs spells their anchors) the
// edit damaged. Every node outside those subs must lower as it did in f, as
// psc's damage says, and is shared with f; f itself is not changed.
//
// It reports false, and returns nil, when an anchor names no sub of root
// or of f.
func Relower(f *File, root *parse.Node, anchors []string) (*File, bool) {
	want := map[string]bool{}
	for _, a := range anchors {
		want[a] = true
	}
	repl := map[ID]Node{}
	seen := map[string]int{}
	var find func(n *parse.Node)
	find = func(n *parse.Node) {
		if key, ok := AnchorKey(n); ok {
			anchor := key + "#" + strconv.Itoa(seen[key])
			seen[key]++
			if want[anchor] {
				l := &lowerer{ids: subIDs(n, anchor)}
				repl[ID{Anchor: anchor}] = l.expr(n)
			}
			return
		}
		for _, c := range n.Children {
			find(c)
		}
	}
	find(root)
	if len(repl) != len(want) {
		return nil, false
	}
	out, ok := splice(f, repl).(*File)
	if !ok || len(repl) != 0 {
		return nil, false
	}
	return out, true
}

// splice is n with each node whose id repl holds replaced by repl's node,
// deleting each from repl as it is placed; a node on the path to one is a
// shallow copy, everything else is n's own. A top-level sub is never inside
// a named sub, so a node of some anchor's is not looked inside.
func splice(n Node, repl map[ID]Node) Node {
	id := n.Meta().ID
	if r, ok := repl[id]; ok {
		delete(repl, id)
		return r
	}
	if id.Anchor != "" || len(repl) == 0 {
		return n
	}
	v := reflect.ValueOf(n).Elem()
	var out reflect.Value // n's copy, made at its first changed child
	set := func(i int, x reflect.Value) {
		if !out.IsValid() {
			out = reflect.New(v.Type()).Elem()
			out.Set(v)
		}
		out.Field(i).Set(x)
	}
	for i := 0; i < v.NumField(); i++ {
		f, ft := v.Field(i), v.Type().Field(i).Type
		switch {
		case ft == nodeType && !f.IsNil():
			c := f.Interface().(Node)
			if r := splice(c, repl); r != c {
				set(i, reflect.ValueOf(r))
			}
		case ft.Kind() == reflect.Slice && ft.Elem() == nodeType:
			var kids []Node
			for j := 0; j < f.Len(); j++ {
				c, _ := f.Index(j).Interface().(Node)
				if c == nil {
					continue
				}
				if r := splice(c, repl); r != c {
					if kids == nil {
						kids = append([]Node(nil), f.Interface().([]Node)...)
					}
					kids[j] = r
				}
			}
			if kids != nil {
				set(i, reflect.ValueOf(kids))
			}
		}
	}
	if !out.IsValid() {
		return n
	}
	return out.Addr().Interface().(Node)
}
