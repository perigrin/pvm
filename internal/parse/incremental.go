// ABOUTME: Edit and ReparseIncremental, the incremental re-parse entry points (spec §6.5, §7.7).
// ABOUTME: For now each is a full re-parse, so §7.7's property holds before Repair exists.
package parse

// Edit replaces the old source's bytes [Start, OldEnd) with NewText.
type Edit struct {
	Start, OldEnd int
	NewText       string
}

// ReparseIncremental returns the tree of src, the old source with e applied,
// reusing old where it can. It equals Parse(src) under TreeDiff (§7.7.1).
// Until Repair lands it reuses nothing.
func ReparseIncremental(old *Node, e Edit, src []byte) *Node {
	return Parse(src)
}

// ReparseIncremental is the resolver form: its result equals a fresh
// NewSession().ParseSourceFrom(path, src, roots...). It parses in a fresh
// Session, never s's module facts, because a Session's second parse reports
// none of the LoadedModules, DeclarationErrors or DeclarationWarnings its
// first already read, and a module edited on disk must be read again.
func (s *Session) ReparseIncremental(old *Node, e Edit, path string, src []byte, roots ...string) *Node {
	return NewSession().ParseSourceFrom(path, src, roots...)
}
