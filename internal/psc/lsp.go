// ABOUTME: LSP server skeleton for psc, managing open Perl documents.
// ABOUTME: Provides OpenDocument/CloseDocument and type inference query methods.

package psc

import (
	"net/url"
	"path/filepath"
	"sync"
	"unicode"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// document holds an open Perl source file and its analysis: the lowered
// tree, its position map, and what inference reported over it.
type document struct {
	uri      string
	text     *Document
	analysis *infer.FileAnalysis

	// dir is the directory of a file: uri's path, and why is incRoots'
	// note on it: why a perl its project names was not resolved.
	dir, why string
}

// LSPServer maintains parsed state for open Perl documents.
// It is safe for concurrent use.
type LSPServer struct {
	// Options is how each open is checked, as psc check's flags set it.
	// Set it before the first OpenDocument.
	Options infer.Options

	documents map[string]*document
	mu        sync.RWMutex
}

// NewLSPServer creates a new LSPServer ready to manage Perl documents.
func NewLSPServer() *LSPServer {
	return &LSPServer{
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
//
// Each open parses in a Session of its own, so every module it uses is read
// from disk as it is now: one edited since the last open is re-read, and one
// missing then is found once it exists.
func (s *LSPServer) OpenDocument(uri string, source []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var root *parse.Node
	var dir, why string
	if path, ok := uriPath(uri); ok {
		// ponytail: @INC is resolved on every open; cache it per directory
		// if resolution shows up in a profile.
		var roots []string
		roots, why = incRoots(path)
		dir = filepath.Dir(path)
		// ponytail: every open re-reads the modules it uses; cache them by
		// mtime, invalidating facts and seen together, if opens get slow.
		root = parse.NewSession().ParseSourceFrom(path, source, roots...)
	} else {
		root = parse.Parse(source)
	}

	s.documents[uri] = &document{
		uri:      uri,
		text:     NewDocument(source),
		analysis: analyze(root, s.Options),
		dir:      dir,
		why:      why,
	}
	return nil
}

// uriPath is the path a file: uri names, and false for a uri of any other
// scheme or with no path. A Windows uri spells its drive after the root
// slash, file:///C:/x.pl, and names C:\x.pl.
func uriPath(uri string) (string, bool) {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "file" || u.Path == "" {
		return "", false
	}
	p := u.Path
	// ponytail: read as a drive on every OS, so a Unix directory named
	// like one, /C:/x.pl, is misread; check runtime.GOOS if one appears.
	if len(p) >= 3 && p[0] == '/' && p[2] == ':' && unicode.IsLetter(rune(p[1])) {
		p = p[1:]
	}
	return filepath.FromSlash(p), true
}

// incNote returns the directory of the document at uri and incRoots' note
// on it, why a perl its project names was not resolved: "" when one was,
// when the project names none, or when uri is not an open file: uri.
func (s *LSPServer) incNote(uri string) (dir, why string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if doc, ok := s.documents[uri]; ok {
		return doc.dir, doc.why
	}
	return "", ""
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
	if d := s.Text(uri); d != nil {
		return d.Text
	}
	return nil
}

// Text returns the text and line table of the document at uri, or nil if
// the uri is not currently open.
func (s *LSPServer) Text(uri string) *Document {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, ok := s.documents[uri]
	if !ok {
		return nil
	}
	return doc.text
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
// named by the sigil of the variable it reads (aggregateSigil) and its name, or a Call, named by its sub's name, and the
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
		name = aggregateSigil(doc.analysis.Positions, n) + n.Name
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

// aggregateSigil is the sigil of the variable v reads. The base of a
// subscript without an arrow reads the aggregate, not the variable its
// sigil spells: $x[0] and @x[0,1] read @x, $h{k} and @h{k} read %h. Any
// other Var, $r->[0]'s base among them, reads the variable it spells.
func aggregateSigil(positions *lower.Positions, v *lower.Var) string {
	start, _, _ := positions.Span(v.ID)
	// A subscript starts where its base does.
	sub := positions.LastBefore(start+1, func(n lower.Node) bool {
		switch n := n.(type) {
		case *lower.Index:
			return n.Base == lower.Node(v) && !n.Arrow
		case *lower.Key:
			return n.Base == lower.Node(v) && !n.Arrow
		case *lower.Slice:
			return n.Base == lower.Node(v)
		}
		return false
	})
	switch n := sub.(type) {
	case *lower.Index:
		return "@"
	case *lower.Key:
		return "%"
	case *lower.Slice:
		if n.Kind == lower.SliceHash || n.Kind == lower.SliceKVHash {
			return "%"
		}
		return "@"
	}
	return v.Sigil
}
