// ABOUTME: LSP server skeleton for psc, managing open Perl documents.
// ABOUTME: Provides OpenDocument/CloseDocument and type inference query methods.

package psc

import (
	"net/url"
	"path/filepath"
	"sync"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// document holds an open Perl source file and its analysis: the lowered
// tree, its position map, and what inference reported over it.
type document struct {
	uri      string
	source   []byte
	analysis *infer.FileAnalysis
}

// LSPServer maintains parsed state for open Perl documents.
// It is safe for concurrent use.
type LSPServer struct {
	session   *parse.Session
	documents map[string]*document
	mu        sync.RWMutex
}

// NewLSPServer creates a new LSPServer ready to manage Perl documents.
func NewLSPServer() *LSPServer {
	return &LSPServer{
		session:   parse.NewSession(),
		documents: make(map[string]*document),
	}
}

// OpenDocument parses the given source, runs type inference, and stores the
// result under uri. If the uri is already open, the document is replaced.
//
// The source is the editor's buffer, not the file on disk. A file: uri is
// parsed as psc check parses the file at its path: its `use` resolves
// against the path's directory and then @INC, read as text and never by
// running perl (incRoots). Any other uri has no directory, and its `use`
// resolves nothing. Analysis re-runs over the whole tree on every open.
func (s *LSPServer) OpenDocument(uri string, source []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var root *parse.Node
	if u, err := url.Parse(uri); err == nil && u.Scheme == "file" && u.Path != "" {
		path := filepath.FromSlash(u.Path)
		// ponytail: @INC is resolved on every open; cache it per directory
		// if resolution shows up in a profile.
		root = s.session.ParseSourceFrom(path, source, incRoots(path)...)
	} else {
		root = parse.Parse(source)
	}

	s.documents[uri] = &document{
		uri:      uri,
		source:   source,
		analysis: infer.AnalyzeParsed(root, nil, infer.Options{}),
	}
	return nil
}

// CloseDocument removes the document for the given uri from the server,
// releasing its lowered tree, inference annotations, diagnostics, and
// symbols. Closing a document that is not open is a no-op.
func (s *LSPServer) CloseDocument(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if doc, ok := s.documents[uri]; ok {
		doc.analysis = nil
	}
	delete(s.documents, uri)
}

// Source returns the raw source bytes for the document at uri, or nil
// if the uri is not currently open.
func (s *LSPServer) Source(uri string) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, ok := s.documents[uri]
	if !ok {
		return nil
	}
	return doc.source
}

// Annotations returns the type of each lowered node of the document at uri,
// keyed by its id, or nil if the uri is not currently open.
func (s *LSPServer) Annotations(uri string) map[lower.ID]types.Type {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, ok := s.documents[uri]
	if !ok {
		return nil
	}
	return doc.analysis.Annotations
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
	return doc.analysis.Diagnostics
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
	return doc.analysis.Symbols
}

// nodeAt returns the innermost lowered node of the document at uri whose
// bytes cover offset, through the position map: bytes the lowering erased,
// such as a paren, resolve to the node they belonged to. It returns nil if
// uri is not open or no node covers offset.
func (s *LSPServer) nodeAt(uri string, offset uint32) (*document, lower.Node) {
	doc, ok := s.documents[uri]
	if !ok {
		return nil, nil
	}
	return doc, doc.analysis.Positions.At(int(offset))
}

// TypeAtByte returns the inferred type of the innermost lowered node at the
// given byte offset in the document at uri. A Hole, source perl rejects, is
// typed Any and never inspected (§4.14.2), so it reports nothing.
// Returns (Unknown, false) if uri is not open or no node has a type there.
func (s *LSPServer) TypeAtByte(uri string, offset uint32) (types.Type, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, n := s.nodeAt(uri, offset)
	if n == nil {
		return types.Unknown, false
	}
	if _, hole := n.(*lower.Hole); hole {
		return types.Unknown, false
	}
	t, found := doc.analysis.Annotations[n.Meta().ID]
	if !found {
		return types.Unknown, false
	}
	return t, true
}

// DefinitionAtByte finds the declaration of the variable or sub named at the
// given byte offset in the document at uri: the lowered node there is a Var,
// named by its sigil and name, or a Call, named by its sub's name, and the
// name is looked up in the symbol table.
// Returns (nil, false) if uri is not open, if no symbol table exists, or if
// no symbol covers the given offset.
func (s *LSPServer) DefinitionAtByte(uri string, offset uint32) (*infer.Symbol, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, n := s.nodeAt(uri, offset)
	if n == nil || doc.analysis.Symbols == nil {
		return nil, false
	}

	var name string
	switch n := n.(type) {
	case *lower.Var:
		name = n.Sigil + n.Name
	case *lower.Call:
		name = n.Name
	}
	if name == "" {
		return nil, false
	}

	sym, found := doc.analysis.Symbols.LookupAll(name)
	if !found {
		return nil, false
	}
	result := sym
	return &result, true
}
