// ABOUTME: damageFromTreeDiff: the top-level subs an edit damages, from the old and new full-parse trees (§6.10.5 item 4).
// ABOUTME: Class B declaration edits, non-trivia top-level edits and parse-error changes go full; a damaged sub's callers are damaged.

package psc

import (
	"fmt"
	"maps"
	"strings"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// editDamage is what an edit invalidates of a file's lowering and analysis.
//
// Full means re-lower and re-analyze everything; Subs is then nil. Otherwise
// Subs is the anchors, as lower.ID.Anchor spells them in the new tree, of
// the top-level subs to re-lower and re-analyze, in source order, and
// everything else -- the other subs and all code outside any sub -- keeps
// its lowering and analysis, shifted. No Subs and not Full is an edit that
// changed nothing the analysis reads, such as one inside top-level trivia.
type editDamage struct {
	Full bool
	// ClassB is set with Full when a declaration changed (§6.1.2 Class B).
	ClassB bool
	// Reason says why Full, for a test or a log to read.
	Reason string
	Subs   []string
}

// damageFromTreeDiff is the damage an edit did, read off the full parses of
// the buffer before it (oldRoot) and after it (newRoot); nil is an empty file.
//
// A top-level sub is one named sub not inside another, which is what opens
// an anchor at the top of lower.AssignIDs. It is damaged when anything its
// lowering or analysis reads changed:
//
//   - its own subtree, every field of every node compared, with its offsets
//     moved by however far the edit shifted it (parse.TreeDiffShifted);
//   - a sub it calls, by name, transitively: a caller's analysis reads its
//     callee's return and parameter types and, for a guard, its body.
//
// The edit goes full instead when what it changed reaches beyond the subs:
//
//   - a declaration (Class B): a `sub` or `method`'s name, prototype,
//     attributes or signature, a `package` or `class`, a `field`, an `our`,
//     a `use` or `no`, or a BEGIN block, added, removed or edited, anywhere;
//   - whether the file has a parse error, which gates what is published and
//     is how a missing `}` that re-nests later subs shows;
//   - the set of top-level subs, or any non-trivia node outside them;
//   - which variables a damaged sub assigns, which widens their type for
//     every other sub that reads them (infer's assignedNames);
//   - top-level code calling a damaged sub.
//
// An edit inside top-level trivia, whitespace or a comment between two subs,
// leaves every sub a shifted copy and every node outside them equal, so it
// damages nothing.
//
// shortcut: a method call is not a dependency, because infer types one only
// through the project index, which psc does not pass; when infer resolves a
// method to a sub of the file, add `->name` to callees.
func damageFromTreeDiff(oldRoot, newRoot *parse.Node) editDamage {
	oldRoot, newRoot = orEmpty(oldRoot), orEmpty(newRoot)
	full := func(classB bool, format string, args ...any) editDamage {
		return editDamage{Full: true, ClassB: classB, Reason: fmt.Sprintf(format, args...)}
	}
	if oldRoot.HasError != newRoot.HasError {
		return full(false, "parse error status changed")
	}
	if d := declarationDiff(oldRoot, newRoot); d != "" {
		return full(true, "a declaration changed: %s", d)
	}
	oldSubs, newSubs := outerSubs(oldRoot), outerSubs(newRoot)
	if len(oldSubs) != len(newSubs) {
		return full(false, "%d top-level subs became %d", len(oldSubs), len(newSubs))
	}
	oldSet, newSet := map[*parse.Node]bool{}, map[*parse.Node]bool{}
	for i := range oldSubs {
		if oldSubs[i].anchor != newSubs[i].anchor {
			return full(false, "top-level sub %s became %s", oldSubs[i].anchor, newSubs[i].anchor)
		}
		oldSet[oldSubs[i].node], newSet[newSubs[i].node] = true, true
	}
	if d := skeletonDiff(oldRoot, newRoot, newRoot.End-oldRoot.End, oldSet, newSet); d != "" {
		return full(false, "top-level edit: %s", d)
	}

	damaged := make([]bool, len(newSubs))
	names := map[string]bool{}
	for i := range newSubs {
		o, n := oldSubs[i].node, newSubs[i].node
		if parse.TreeDiffShifted(o, n, n.Start-o.Start) == "" {
			continue
		}
		if !maps.Equal(writes(o), writes(n)) {
			return full(false, "%s changes which variables it assigns", newSubs[i].anchor)
		}
		damaged[i], names[newSubs[i].name] = true, true
	}
	calls := make([]map[string]bool, len(newSubs))
	for i, s := range newSubs {
		calls[i] = callees(s.node, nil)
	}
	for grew := len(names) > 0; grew; {
		grew = false
		for i, s := range newSubs {
			if !damaged[i] && intersects(calls[i], names) {
				damaged[i], names[s.name], grew = true, true, true
			}
		}
	}
	if len(names) > 0 && intersects(callees(newRoot, newSet), names) {
		return full(false, "top-level code calls a damaged sub")
	}
	var out editDamage
	for i, s := range newSubs {
		if damaged[i] {
			out.Subs = append(out.Subs, s.anchor)
		}
	}
	return out
}

func orEmpty(root *parse.Node) *parse.Node {
	if root == nil {
		return &parse.Node{Kind: parse.SourceFile}
	}
	return root
}

// outerSub is a top-level sub: its declaration node, its name as a call
// names it (unqualified), and its anchor in lower.AssignIDs' spelling.
type outerSub struct {
	node   *parse.Node
	name   string
	anchor string
}

// outerSubs is root's top-level subs in source order, numbered as
// lower.AssignIDs numbers same-named anchors.
func outerSubs(root *parse.Node) []outerSub {
	var out []outerSub
	seen := map[string]int{}
	var walk func(n *parse.Node)
	walk = func(n *parse.Node) {
		if key, ok := lower.AnchorKey(n); ok {
			out = append(out, outerSub{node: n, name: baseName(n.Children[0].Text), anchor: fmt.Sprintf("%s#%d", key, seen[key])})
			seen[key]++
			return
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(root)
	return out
}

// declarationDiff is the first Class B declaration that differs between
// the two trees, or "".
func declarationDiff(oldRoot, newRoot *parse.Node) string {
	a, b := declarations(oldRoot), declarations(newRoot)
	if len(a) != len(b) {
		return fmt.Sprintf("%d declarations became %d", len(a), len(b))
	}
	for i := range a {
		if d := parse.TreeDiffShifted(a[i], b[i], b[i].Start-a[i].Start); d != "" {
			return fmt.Sprintf("%v %q: %s", a[i].Kind, a[i].Text, d)
		}
	}
	return ""
}

// declarations is every Class B declaration under root in preorder: a
// `use` or `no` and a BEGIN block whole, the others without their body.
func declarations(root *parse.Node) []*parse.Node {
	var out []*parse.Node
	var walk func(n *parse.Node)
	walk = func(n *parse.Node) {
		switch {
		case n.Kind == parse.Use && n.Text != "require", n.Kind == parse.Phaser && n.Text == "BEGIN":
			out = append(out, n)
		case n.Kind == parse.Declaration:
			switch n.Text {
			case "sub", "method", "package", "class", "field", "our":
				out = append(out, head(n))
			}
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(root)
	return out
}

// head is a copy of a declaration without its trailing block, ending
// where the block starts: `sub f($) :lvalue` of `sub f($) :lvalue {...}`.
func head(n *parse.Node) *parse.Node {
	h := *n
	if k := len(n.Children) - 1; k >= 0 && n.Children[k].Kind == parse.Block {
		h.Children, h.End, h.HasError = n.Children[:k], n.Children[k].Start, false
	}
	return &h
}

// skeletonDiff compares the two trees outside their top-level subs (the
// nodes in oldSubs and newSubs), every field of every node, where an
// offset is equal either as it was or moved by delta, the edit's length
// change: a node before the edit keeps its offsets, one after it moves by
// delta, and one around it, the file or a package block, keeps its start.
func skeletonDiff(o, n *parse.Node, delta int, oldSubs, newSubs map[*parse.Node]bool) string {
	if oldSubs[o] || newSubs[n] {
		if oldSubs[o] && newSubs[n] {
			return ""
		}
		return fmt.Sprintf("a sub and a %v at %d", n.Kind, n.Start)
	}
	oc, nc := *o, *n
	oc.Children, nc.Children = nil, nil
	for _, f := range []struct {
		old int
		new *int
	}{{o.Start, &nc.Start}, {o.End, &nc.End}, {o.InnerStart, &nc.InnerStart}, {o.InnerEnd, &nc.InnerEnd}} {
		if *f.new == f.old+delta {
			*f.new = f.old
		}
	}
	if d := parse.TreeDiff(&oc, &nc); d != "" {
		return fmt.Sprintf("%v at %d: %s", n.Kind, n.Start, d)
	}
	if len(o.Children) != len(n.Children) {
		return fmt.Sprintf("%v at %d: %d children became %d", n.Kind, n.Start, len(o.Children), len(n.Children))
	}
	for i := range o.Children {
		if d := skeletonDiff(o.Children[i], n.Children[i], delta, oldSubs, newSubs); d != "" {
			return d
		}
	}
	return ""
}

// assignOps are the operators infer's assignedNames reads as an
// assignment once lowered: `=` and every compound form.
var assignOps = map[string]bool{
	"=": true, "+=": true, "-=": true, "*=": true, "/=": true, ".=": true, "%=": true,
	"x=": true, "**=": true, "&=": true, "|=": true, "^=": true, "<<=": true, ">>=": true,
	"&&=": true, "||=": true, "//=": true, "&.=": true, "|.=": true, "^.=": true,
}

// writes is the name of every variable an assignment under n writes, as
// infer's assignedNames reads the lowered tree: the variable, or each of a
// list's, left of an assignment operator that is not a declaration's
// initialiser.
func writes(n *parse.Node) map[string]bool {
	out := map[string]bool{}
	var walk func(n *parse.Node, decl bool)
	walk = func(n *parse.Node, decl bool) {
		if n.Kind == parse.Binary && assignOps[n.Text] && !decl && len(n.Children) > 0 {
			targets := []*parse.Node{n.Children[0]}
			if l := n.Children[0]; l.Kind == parse.List {
				targets = l.Children
			}
			for _, t := range targets {
				if t.Kind == parse.Term && t.Text != "" && strings.ContainsRune("$@%", rune(t.Text[0])) {
					out[t.Text] = true
				}
			}
		}
		for _, c := range n.Children {
			walk(c, n.Kind == parse.Declaration)
		}
	}
	walk(n, false)
	return out
}

// callees is the unqualified name of every sub called under n, outside
// the nodes in skip: a call, an `&name`, and a sort comparator; not a
// method call (see damageFromTreeDiff's shortcut).
func callees(n *parse.Node, skip map[*parse.Node]bool) map[string]bool {
	out := map[string]bool{}
	var walk func(n *parse.Node, method bool)
	walk = func(n *parse.Node, method bool) {
		if skip[n] {
			return
		}
		switch {
		case n.Kind == parse.Call && !method:
			out[baseName(n.Text)] = true
		case n.Kind == parse.Term && (strings.HasPrefix(n.Text, "&") || n.Comparator):
			out[baseName(n.Text)] = true
		}
		for i, c := range n.Children {
			// `$obj->name(...)` is a Call right of `->`: a method call.
			walk(c, n.Kind == parse.Binary && n.Text == "->" && i == 1)
		}
	}
	walk(n, false)
	return out
}

// baseName is a sub's name without its `&` or package: f of `&Foo::f`.
func baseName(name string) string {
	name = strings.TrimPrefix(name, "&")
	if i := strings.LastIndex(name, "::"); i >= 0 {
		name = name[i+2:]
	}
	return name
}

func intersects(a, b map[string]bool) bool {
	for k := range a {
		if b[k] {
			return true
		}
	}
	return false
}
