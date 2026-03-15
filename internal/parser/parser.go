// ABOUTME: Thin wrapper around the gotreesitter pure-Go Perl grammar.
// ABOUTME: Provides a stable API for parsing Perl source into syntax trees.

package parser

import (
	gotreesitter "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

// Parser parses Perl source using the gotreesitter pure-Go runtime.
type Parser struct {
	inner *gotreesitter.Parser
	lang  *gotreesitter.Language
}

// Tree is a parsed syntax tree. It carries the source so that nodes can
// extract their text without the caller needing to pass source separately.
type Tree struct {
	inner  *gotreesitter.Tree
	lang   *gotreesitter.Language
	source []byte
}

// Node is a single node in the syntax tree. It carries the Language so
// callers can call Kind(), SExpr(), and ChildByFieldName() without passing
// Language as a separate argument.
type Node struct {
	inner *gotreesitter.Node
	lang  *gotreesitter.Language
}

// New creates a new Perl parser backed by the gotreesitter pure-Go runtime.
func New() *Parser {
	lang := grammars.PerlLanguage()
	return &Parser{
		inner: gotreesitter.NewParser(lang),
		lang:  lang,
	}
}

// Parse parses the given Perl source and returns a Tree.
func (p *Parser) Parse(source []byte) (*Tree, error) {
	tree, err := p.inner.Parse(source)
	if err != nil {
		return nil, err
	}
	return &Tree{inner: tree, lang: p.lang, source: source}, nil
}

// RootNode returns the root node of the syntax tree.
func (t *Tree) RootNode() *Node {
	if t == nil || t.inner == nil {
		return nil
	}
	root := t.inner.RootNode()
	if root == nil {
		return nil
	}
	return &Node{inner: root, lang: t.lang}
}

// Kind returns the node type name (e.g. "source_file", "scalar_variable").
func (n *Node) Kind() string {
	if n == nil || n.inner == nil {
		return ""
	}
	return n.inner.Type(n.lang)
}

// IsNamed reports whether this is a named node (not anonymous punctuation).
func (n *Node) IsNamed() bool {
	if n == nil || n.inner == nil {
		return false
	}
	return n.inner.IsNamed()
}

// HasError reports whether this node or any descendant has a parse error.
func (n *Node) HasError() bool {
	if n == nil || n.inner == nil {
		return false
	}
	return n.inner.HasError()
}

// IsError reports whether this node is an explicit error node.
func (n *Node) IsError() bool {
	if n == nil || n.inner == nil {
		return false
	}
	return n.inner.IsError()
}

// StartByte returns the byte offset where this node begins.
func (n *Node) StartByte() uint32 {
	if n == nil || n.inner == nil {
		return 0
	}
	return n.inner.StartByte()
}

// EndByte returns the byte offset where this node ends (exclusive).
func (n *Node) EndByte() uint32 {
	if n == nil || n.inner == nil {
		return 0
	}
	return n.inner.EndByte()
}

// Text returns the source text covered by this node.
func (n *Node) Text(source []byte) string {
	if n == nil || n.inner == nil {
		return ""
	}
	return n.inner.Text(source)
}

// ChildCount returns the total number of children (named and anonymous).
func (n *Node) ChildCount() int {
	if n == nil || n.inner == nil {
		return 0
	}
	return n.inner.ChildCount()
}

// Child returns the i-th child, or nil if out of range.
func (n *Node) Child(i int) *Node {
	if n == nil || n.inner == nil {
		return nil
	}
	child := n.inner.Child(i)
	if child == nil {
		return nil
	}
	return &Node{inner: child, lang: n.lang}
}

// NamedChildCount returns the number of named children.
func (n *Node) NamedChildCount() int {
	if n == nil || n.inner == nil {
		return 0
	}
	return n.inner.NamedChildCount()
}

// NamedChild returns the i-th named child, or nil if out of range.
func (n *Node) NamedChild(i int) *Node {
	if n == nil || n.inner == nil {
		return nil
	}
	child := n.inner.NamedChild(i)
	if child == nil {
		return nil
	}
	return &Node{inner: child, lang: n.lang}
}

// ChildByFieldName returns the first child with the given field name, or nil.
func (n *Node) ChildByFieldName(name string) *Node {
	if n == nil || n.inner == nil {
		return nil
	}
	child := n.inner.ChildByFieldName(name, n.lang)
	if child == nil {
		return nil
	}
	return &Node{inner: child, lang: n.lang}
}

// Parent returns this node's parent, or nil if this is the root.
func (n *Node) Parent() *Node {
	if n == nil || n.inner == nil {
		return nil
	}
	parent := n.inner.Parent()
	if parent == nil {
		return nil
	}
	return &Node{inner: parent, lang: n.lang}
}

// SExpr returns the S-expression representation of this subtree.
func (n *Node) SExpr() string {
	if n == nil || n.inner == nil {
		return ""
	}
	return n.inner.SExpr(n.lang)
}

// Edit describes a single source-text edit for incremental parsing.
// StartByte is the byte offset where the change begins.
// OldEndByte is the byte offset of the end of the old text (exclusive).
// NewEndByte is the byte offset of the end of the new text (exclusive).
type Edit struct {
	StartByte  uint32
	OldEndByte uint32
	NewEndByte uint32
}

// ParseIncremental re-parses source after an edit was applied to oldTree,
// reusing unchanged subtrees for efficiency. If oldTree is nil, it falls
// back to a full parse.
func (p *Parser) ParseIncremental(source []byte, oldTree *Tree, edit Edit) (*Tree, error) {
	if oldTree == nil {
		return p.Parse(source)
	}

	// Record the edit on the old tree so the incremental parser knows what changed.
	oldTree.inner.Edit(gotreesitter.InputEdit{
		StartByte:  edit.StartByte,
		OldEndByte: edit.OldEndByte,
		NewEndByte: edit.NewEndByte,
	})

	tree, err := p.inner.ParseIncremental(source, oldTree.inner)
	if err != nil {
		return nil, err
	}
	return &Tree{inner: tree, lang: p.lang, source: source}, nil
}
