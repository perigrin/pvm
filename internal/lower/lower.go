// ABOUTME: The entry point of internal/lower: Lower turns an internal/parse tree into the lowered File.
// ABOUTME: It assigns every CST node its id and lowers the root's statements in source order.

package lower

import "tamarou.com/pvm/internal/parse"

// Lower lowers the tree root, as parse.Parse returns it, to a File. Every
// lowered node carries the id AssignIDs gives the CST node it came from;
// trivia lowers to nothing.
func Lower(root *parse.Node) *File {
	l := &lowerer{ids: AssignIDs(root)}
	out := &File{}
	for _, c := range root.Children {
		if s := l.stmt(c); s != nil {
			out.Stmts = append(out.Stmts, s)
		}
	}
	l.stamp(root, out)
	return out
}
