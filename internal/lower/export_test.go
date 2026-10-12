// ABOUTME: Test-only entry points into the unexported expression lowering, for package lower_test.
// ABOUTME: Slice tests lower one expression node of a parsed file; the file-level entry is lower.Lower.

package lower

import "tamarou.com/pvm/internal/parse"

// LowerExpr lowers n, a node of the tree rooted at root, as an expression.
func LowerExpr(root, n *parse.Node) Node {
	return (&lowerer{ids: AssignIDs(root)}).expr(n)
}

// LowerStmt lowers n, a node of the tree rooted at root, as a statement.
func LowerStmt(root, n *parse.Node) Node {
	return (&lowerer{ids: AssignIDs(root)}).stmt(n)
}

// LowerHashKey lowers n as the subscript of an `index "{"`.
func LowerHashKey(root, n *parse.Node) Node {
	return (&lowerer{ids: AssignIDs(root)}).hashKey(n)
}
