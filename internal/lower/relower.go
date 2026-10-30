// ABOUTME: Relower: an edited file's lowering and position map from the ones before it, redoing only the statements an edit changed.
// ABOUTME: The rest of the old tree and map is shared; only the path from the File down to each replaced Decl is copied.

package lower

import (
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"

	"tamarou.com/pvm/internal/parse"
)

// Relower is Lower(root) and its NewPositions for root, the tree of an
// edited buffer, given f and pos, the lowering and position map of the
// tree before the edit, and anchors, the top-level subs (named subs not
// inside another, as AssignIDs spells their anchors) the edit damaged.
//
// Every top-level statement but those holding a damaged sub or the edit
// itself must be one of the old tree's moved whole, and every node outside
// the damaged subs must lower as it did in f, as psc's damage says. Those
// statements keep their lowering and their chunk of pos; only the others
// are numbered again, and only the damaged subs lowered again. Neither f
// nor pos is changed.
//
// It reports false, and returns nil, when an anchor names no sub of root
// or of f, or root's top-level statements do not number as the old
// tree's did; the caller lowers afresh.
//
// shortcut: a statement is renumbered whole, so a damaged sub inside a
// `package NAME {...}` block renumbers the block; chunk below the
// statement if a file of package blocks shows up in a profile.
func Relower(f *File, pos *Positions, root *parse.Node, anchors []string) (*File, *Positions, bool) {
	if pos == nil || len(root.Children) != len(pos.chunks) {
		return nil, nil, false
	}
	redo := map[int][]string{} // each chunk to number again, and its damaged subs
	for _, a := range anchors {
		i, ok := pos.subs[a]
		if !ok {
			return nil, nil, false
		}
		redo[i] = append(redo[i], a)
	}
	if i, ok := edited(pos.root.Children, root.Children); ok {
		if _, in := redo[i]; !in {
			redo[i] = nil
		}
	}
	out := f
	if len(anchors) != 0 {
		out = &File{nodeBase: f.nodeBase, Stmts: slices.Clone(f.Stmts)}
	}
	q := &Positions{root: root, file: out, chunks: slices.Clone(pos.chunks), his: pos.his, subs: pos.subs}
	for i, subs := range redo {
		c, ok := q.renumber(pos, i)
		if !ok {
			return nil, nil, false
		}
		stmts := out.Stmts[sort.Search(len(out.Stmts), func(k int) bool { return q.chunkOf(out.Stmts[k].Meta().ID) >= i }):]
		stmts = stmts[:sort.Search(len(stmts), func(k int) bool { return q.chunkOf(stmts[k].Meta().ID) > i })]
		if len(subs) != 0 {
			l := &lowerer{ids: c.ids}
			repl := map[ID]Node{}
			for _, a := range subs {
				n := c.cst[ID{Anchor: a}]
				if n == nil {
					return nil, nil, false
				}
				repl[ID{Anchor: a}] = l.expr(n)
			}
			for k, s := range stmts {
				stmts[k] = splice(s, repl)
			}
			if len(repl) != 0 {
				return nil, nil, false
			}
		}
		for _, s := range stmts {
			q.index(c, s)
		}
		c.sort()
		q.chunks[i] = c
		q.renumbered = append(q.renumbered, i)
	}
	slices.Sort(q.renumbered)
	return out, q, true
}

// renumber numbers the top-level statement root.Children[i] afresh, as
// pos numbered the one before the edit, and reports false when it numbers
// other ids than that one did.
func (q *Positions) renumber(pos *Positions, i int) (*chunk, bool) {
	next, seen := map[string]int{"": 1}, map[string]int{}
	if i > 0 {
		next[""] = pos.his[i-1]
	}
	// A sub's disambiguator counts the same-named subs before it.
	for _, a := range pos.chunks[i].subs {
		j := strings.LastIndexByte(a, '#')
		k, _ := strconv.Atoi(a[j+1:])
		if v, ok := seen[a[:j]]; !ok || k < v {
			seen[a[:j]] = k
		}
	}
	c := q.number(q.root.Children[i], next, seen)
	return c, next[""] == pos.his[i] && slices.Equal(c.subs, pos.chunks[i].subs)
}

// edited is the index of the top-level statement an edit grew or shrank,
// old's in the tree before it and now's after: the first whose end moved,
// if its start did not. Every statement before it ends where it did and
// every one after it moved by the same length, so a binary search finds it.
func edited(old, now []*parse.Node) (int, bool) {
	i := sort.Search(len(now), func(i int) bool { return now[i].End != old[i].End })
	return i, i < len(now) && now[i].Start == old[i].Start
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
