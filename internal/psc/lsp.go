// ABOUTME: LSP server skeleton for psc, managing open Perl documents.
// ABOUTME: Provides OpenDocument/CloseDocument for future LSP protocol support.

package psc

import (
	"fmt"
	"sync"

	"tamarou.com/pvm/internal/parser"
)

// document holds the parsed state of an open Perl source file.
type document struct {
	uri    string
	source []byte
	tree   *parser.Tree
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

// OpenDocument parses the given source and stores it under uri.
// If the uri is already open, the document is replaced.
func (s *LSPServer) OpenDocument(uri string, source []byte) error {
	tree, err := s.parser.Parse(source)
	if err != nil {
		return fmt.Errorf("parse %s: %w", uri, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.documents[uri] = &document{
		uri:    uri,
		source: source,
		tree:   tree,
	}
	return nil
}

// CloseDocument removes the document for the given uri from the server.
// Closing a document that is not open is a no-op.
func (s *LSPServer) CloseDocument(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.documents, uri)
}
