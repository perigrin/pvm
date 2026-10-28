// ABOUTME: Node identity for internal/lower: ids that survive an edit above the node, never byte offsets.
// ABOUTME: Ids are numbered per enclosing named sub, anchored by kind, name and disambiguator as in ast_id.rs.

package lower

import (
	"strconv"

	"tamarou.com/pvm/internal/parse"
)

// ID is a node's identity: the key every side table uses (§6.1.6 item 4).
//
// It is not a byte offset, because every key after an edit would shift, and
// a parent and child that start at the same byte would collide. It is not a
// whole-file preorder index either, because one statement inserted at the
// top renumbers everything below it.
//
// Instead an ID is an Anchor plus a Local index, in the manner of
// rust-analyzer's crates/span/src/ast_id.rs. The anchor is the innermost
// enclosing named sub, spelled by kind, name and disambiguator --
// "sub b#0", or "sub a#0/sub inner#0" when one sub is nested in another --
// and "" for code outside any named sub. Local is a preorder index that
// restarts inside every anchor. An edit therefore renumbers only the anchor
// it lands in: the ids inside `sub b` do not see what happened above it.
//
// Identity is per SYNTAX SITE (§4.14.1 rule 6): one `sub { $i }` has one id
// however many closures it creates at run time. A table keyed on a captured
// value keys on (ID, capture instance); ID is comparable so that pair is a
// valid map key too.
//
// Role tells apart the lowered nodes built from one parse node. Role 0 is
// the parse node itself; a lowered node with no parse node of its own takes
// its parent's Anchor and Local and a nonzero Role.
type ID struct {
	Anchor string
	Local  int
	Role   Role
}

// Role names a lowered node synthesized beside its parent's parse node.
type Role int

const (
	// RoleSubValue is a named sub's AnonSub: `sub f {...}` is one parse
	// node, lowered to a Decl and, as its Init, the sub value itself.
	RoleSubValue Role = iota + 1
	// RoleDerefOperand is the Hole in a Deref whose operand the parser
	// never built: `$#{$x}`, `$#$x` and `&$x` are each one childless term.
	RoleDerefOperand
	// RoleAngleBody is the body of a `<...>` term, one token: Readline's
	// Handle in `<STDIN>` and `<$fh>`, Glob's Pattern in `<*.c>`. A term
	// lowers to one or the other, so one role keeps every id distinct.
	RoleAngleBody
	// RoleDeclTarget is the Decl a compound-assigned declaration lowers to
	// beneath its Assign: `my $x -= 1` is one parse node, lowered to the
	// Assign and, as its target, the Decl.
	RoleDeclTarget
)

// AssignIDs gives every node of a tree, trivia and Unknown included, an ID.
//
// The tree is routinely broken (§6.8), and an Unknown may have children
// (`if ($x) {` gives an Unknown holding a block), so nothing here assumes a
// well-formed shape: every child of every node is visited.
func AssignIDs(root *parse.Node) map[*parse.Node]ID {
	ids := map[*parse.Node]ID{}
	assignIDsFrom(ids, root, "", map[string]int{}, map[string]int{})
	return ids
}

// subIDs is what AssignIDs gives the nodes of sub, a sub not inside another
// named sub, whose anchor is anchor: an anchor's numbering, and that of the
// subs nested in it, reads nothing outside the anchor.
func subIDs(sub *parse.Node, anchor string) map[*parse.Node]ID {
	ids := map[*parse.Node]ID{sub: {Anchor: anchor}}
	next := map[string]int{anchor: 1}
	seen := map[string]int{}
	for _, c := range sub.Children {
		assignIDsFrom(ids, c, anchor, next, seen)
	}
	return ids
}

// assignIDsFrom numbers n and its subtree under anchor; next is each
// anchor's next Local, seen each anchor key's uses so far.
func assignIDsFrom(ids map[*parse.Node]ID, n *parse.Node, anchor string, next, seen map[string]int) {
	var walk func(n *parse.Node, anchor string)
	walk = func(n *parse.Node, anchor string) {
		if key, ok := AnchorKey(n); ok {
			// ponytail: the name is unqualified, so inserting a same-named
			// sub in an earlier package shifts a later one's disambiguator.
			// Qualify by package if that churn shows up in practice.
			key = anchor + "/" + key
			if anchor == "" {
				key = key[1:]
			}
			anchor = key + "#" + strconv.Itoa(seen[key])
			seen[key]++
		}
		ids[n] = ID{Anchor: anchor, Local: next[anchor]}
		next[anchor]++
		for _, c := range n.Children {
			walk(c, anchor)
		}
	}
	walk(n, anchor)
}

// AnchorKey reports whether a node opens an anchor -- a named sub or method
// -- and returns its kind and name. An anonymous sub has no name to anchor
// on, so its body is numbered within the enclosing anchor.
func AnchorKey(n *parse.Node) (string, bool) {
	if n.Kind != parse.Declaration || (n.Text != "sub" && n.Text != "method") {
		return "", false
	}
	name := subName(n)
	if name == nil {
		return "", false
	}
	return n.Text + " " + name.Text, true
}

// subName returns the name term of a sub or method declaration, or nil when
// the sub is anonymous. It is the one test of a named sub: the id layer
// anchors on it and the lowering gives the sub its Decl, so the two agree.
// A parenthesised first child is an anonymous sub's one-parameter
// signature, `sub ($x) {...}`, sitting where a named sub's name would.
func subName(n *parse.Node) *parse.Node {
	if len(n.Children) == 0 {
		return nil
	}
	c := n.Children[0]
	if c.Kind != parse.Term || c.Paren || !word(c.Text) {
		return nil
	}
	return c
}
