// ABOUTME: LSP server skeleton for psc, managing open Perl documents.
// ABOUTME: Provides OpenDocument/CloseDocument and type inference query methods.

package psc

import (
	"fmt"
	"sync"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/types"
)

// document holds the parsed state of an open Perl source file, including
// inference results from the two-pass type analysis pipeline.
type document struct {
	uri         string
	source      []byte
	tree        *parser.Tree
	annotations map[uint32]types.Type
	diagnostics []infer.Diagnostic
	symbols     *infer.SymbolTable
}

// LSPServer maintains parsed state for open Perl documents.
// It is safe for concurrent use.
type LSPServer struct {
	parser    *parser.Parser
	documents map[string]*document
	mu        sync.RWMutex
}

// NewLSPServer creates a new LSPServer ready to manage Perl documents.
func NewLSPServer() *LSPServer {
	return &LSPServer{
		parser:    parser.New(),
		documents: make(map[string]*document),
	}
}

// OpenDocument parses the given source, runs type inference, and stores the
// result under uri. If the uri is already open, the document is replaced.
func (s *LSPServer) OpenDocument(uri string, source []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tree, err := s.parser.Parse(source)
	if err != nil {
		return fmt.Errorf("parse %s: %w", uri, err)
	}

	annotations, diagnostics, symbols := infer.Analyze(tree, source)

	s.documents[uri] = &document{
		uri:         uri,
		source:      source,
		tree:        tree,
		annotations: annotations,
		diagnostics: diagnostics,
		symbols:     symbols,
	}
	return nil
}

// CloseDocument removes the document for the given uri from the server,
// releasing its parsed tree, inference annotations, diagnostics, and symbols.
// Closing a document that is not open is a no-op.
func (s *LSPServer) CloseDocument(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if doc, ok := s.documents[uri]; ok {
		doc.annotations = nil
		doc.diagnostics = nil
		doc.symbols = nil
	}
	delete(s.documents, uri)
}

// Annotations returns the type annotation map for the document at uri,
// or nil if the uri is not currently open.
func (s *LSPServer) Annotations(uri string) map[uint32]types.Type {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, ok := s.documents[uri]
	if !ok {
		return nil
	}
	return doc.annotations
}

// Diagnostics returns the slice of inference diagnostics for the document at
// uri, or nil if the uri is not currently open.
func (s *LSPServer) Diagnostics(uri string) []infer.Diagnostic {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, ok := s.documents[uri]
	if !ok {
		return nil
	}
	return doc.diagnostics
}

// SymbolTable returns the symbol table for the document at uri, or nil if the
// uri is not currently open.
func (s *LSPServer) SymbolTable(uri string) *infer.SymbolTable {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, ok := s.documents[uri]
	if !ok {
		return nil
	}
	return doc.symbols
}

// TypeAtByte finds the type annotation at the given byte offset in the
// document at uri. It walks the CST to find the deepest node that contains
// the offset, then walks up the ancestor chain until it finds a node whose
// start byte has an entry in the annotation map.
// Returns (Unknown, false) if uri is not open or no annotation is found.
func (s *LSPServer) TypeAtByte(uri string, offset uint32) (types.Type, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, ok := s.documents[uri]
	if !ok {
		return types.Unknown, false
	}

	root := doc.tree.RootNode()
	if root == nil {
		return types.Unknown, false
	}

	// Find the deepest node containing the offset.
	node := deepestNodeAt(root, offset)
	if node == nil {
		return types.Unknown, false
	}

	// Walk up from the deepest node looking for a match in annotations.
	for n := node; n != nil; n = n.Parent() {
		if t, found := doc.annotations[n.StartByte()]; found {
			return t, true
		}
	}

	return types.Unknown, false
}

// deepestNodeAt returns the deepest node in the subtree rooted at node that
// spans the given byte offset (StartByte <= offset < EndByte). It returns nil
// if no such node exists in the subtree.
func deepestNodeAt(node *parser.Node, offset uint32) *parser.Node {
	if node == nil {
		return nil
	}
	if node.StartByte() > offset || node.EndByte() <= offset {
		return nil
	}

	// Try each child first (post-order deepening).
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if result := deepestNodeAt(child, offset); result != nil {
			return result
		}
	}

	// No child spans the offset — this node is the deepest match.
	return node
}

// DefinitionAtByte finds the symbol whose declaration spans the given byte
// offset in the document at uri. It uses the CST to identify the variable or
// function at the offset and then looks it up in the symbol table.
// Returns (nil, false) if uri is not open, if no symbol table exists, or if
// no symbol covers the given offset.
func (s *LSPServer) DefinitionAtByte(uri string, offset uint32) (*infer.Symbol, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, ok := s.documents[uri]
	if !ok {
		return nil, false
	}
	if doc.symbols == nil {
		return nil, false
	}

	root := doc.tree.RootNode()
	if root == nil {
		return nil, false
	}

	// Walk to the deepest node that contains the offset.
	node := deepestNodeAt(root, offset)
	if node == nil {
		return nil, false
	}

	// Build a sigil-qualified name from the node kind and text.
	name := symbolNameFromNode(node, doc.source)
	if name == "" {
		return nil, false
	}

	sym, found := doc.symbols.Lookup(name)
	if !found {
		return nil, false
	}
	result := sym
	return &result, true
}

// symbolNameFromNode derives a sigil-qualified symbol name from a CST node.
// For scalar/array/hash nodes the appropriate sigil is prepended. For
// bareword nodes (subroutine names) the text is returned as-is. For varname
// and anonymous sigil nodes ($, @, %) the parent node is consulted.
func symbolNameFromNode(node *parser.Node, source []byte) string {
	if node == nil {
		return ""
	}
	switch node.Kind() {
	case "scalar":
		return "$" + varname(node, source)
	case "array":
		return "@" + varname(node, source)
	case "hash":
		return "%" + varname(node, source)
	case "varname", "$", "@", "%":
		// Walk up to the variable node to get the full sigil-qualified name.
		parent := node.Parent()
		if parent == nil {
			return node.Text(source)
		}
		return symbolNameFromNode(parent, source)
	case "bareword":
		return node.Text(source)
	}
	return ""
}

// varname extracts the bare identifier text from a scalar/array/hash node by
// looking for a "varname" child. If none is found, the full node text is used.
func varname(node *parser.Node, source []byte) string {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Kind() == "varname" {
			return child.Text(source)
		}
	}
	return node.Text(source)
}
