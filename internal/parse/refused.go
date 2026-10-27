// ABOUTME: Refused, the one reader of "is this node a refusal": an Unknown, or the Error or Missing node that replaced one.
// ABOUTME: Every Kind comparison with Unknown outside Kind.String goes through it, so no reader misses a recovered node.

package parse

// Refused reports whether n is a refusal: an Unknown, or an Error or Missing
// node recovery built where an Unknown stood, which carries that Unknown's
// code (recovery.go). A Missing node recovery synthesized for a token -- a
// closer, a `;`, an identifier -- carries no code and is not one.
//
// The one place the question is asked. A reader that compared Kind with
// Unknown would count a catch-all Error node as nothing, and every ratchet
// built on it would read the replacement as an improvement.
func (n *Node) Refused() bool {
	return n.Kind == Unknown || (n.Kind == Error || n.Kind == Missing) && n.Refusal != ""
}
