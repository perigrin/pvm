// ABOUTME: LSP server for psc: open Perl documents, each held by its owner goroutine.
// ABOUTME: Provides OpenDocument/CloseDocument/Shutdown and type inference query methods.

package psc

import (
	"context"
	"net/url"
	"path/filepath"
	"sync"
	"time"
	"unicode"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// document is one parse of an open Perl source file and its analysis: the
// text and version it parsed, the lowered tree, its position map, and what
// inference reported over it. It is never changed once its owner lands it.
type document struct {
	uri      string
	version  int
	text     *Document
	analysis *infer.FileAnalysis
	// root is the parse tree analysis was made of, which the next parse's
	// damage is read against (analyzeEdit).
	root *parse.Node

	// dir is the directory of a file: uri's path, and why is incRoots'
	// note on it: why a perl its project names was not resolved.
	dir, why string
}

// LSPServer maintains parsed state for open Perl documents, each held by
// an owner goroutine of its own (owner.go). It is safe for concurrent use.
type LSPServer struct {
	// Options is how each open is checked, as psc check's flags set it.
	// Set it before the first OpenDocument.
	Options infer.Options

	// mu guards the map alone; no parse runs under it. A replaced owner is
	// stopped under it, which never waits on mu, so a reopen cannot race it.
	mu        sync.Mutex
	documents map[string]*owner

	// What each owner is made with; tests replace them before an open.
	parseDelay, diagDelay time.Duration
	clock                 clock
	compute               func(ctx context.Context, uri string, text []byte, prev *document) (*document, error)
	publish               func(doc *document, err error)
}

// NewLSPServer creates a new LSPServer ready to manage Perl documents.
func NewLSPServer() *LSPServer {
	s := &LSPServer{
		documents:  make(map[string]*owner),
		parseDelay: parseDebounce,
		diagDelay:  diagnosticsDebounce,
		clock:      realClock{},
	}
	s.compute = s.analyzeSource
	return s
}

// OpenDocument opens source under uri, replacing a document already open
// there, and returns once it is parsed and analysed.
func (s *LSPServer) OpenDocument(uri string, source []byte) error {
	s.open(uri, 0, source)
	s.snapshot(uri)
	return nil
}

// open gives uri a new owner holding text at version, which begins parsing
// it. The owner of a document already open there is stopped first, so it
// cannot publish after the new one.
func (s *LSPServer) open(uri string, version int, text []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old := s.documents[uri]; old != nil {
		old.stop()
	}
	s.documents[uri] = newOwner(s, uri, version, text)
}

// change hands text at version to uri's owner. A uri that is not open has
// no owner, and the change is dropped.
func (s *LSPServer) change(uri string, version int, text []byte) {
	if o := s.owner(uri); o != nil {
		o.edit(version, text)
	}
}

func (s *LSPServer) owner(uri string) *owner {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.documents[uri]
}

// snapshot is the tree of the latest text of the document at uri, or nil
// if uri is not open or no parse of it has succeeded.
func (s *LSPServer) snapshot(uri string) *document {
	if o := s.owner(uri); o != nil {
		return o.current()
	}
	return nil
}

// analyzeSource parses, lowers and analyses source as the document at uri.
// The source is the editor's buffer, not the file on disk. A file: uri is
// parsed as psc check parses the file at its path: its `use` resolves
// against the path's directory and then @INC, read as text and never by
// running perl (incRoots). Any other uri has no directory, and its `use`
// resolves nothing. Given prev, the last parse of the document that
// landed, it re-lowers and re-analyzes only the subs the edit from prev's
// text damaged (analyzeEdit); with none it analyzes the whole tree.
//
// Each parse runs in a Session of its own, so every module it uses is read
// from disk as it is now: one edited since the last parse is re-read, and
// one missing then is found once it exists.
//
// It polls ctx before the parse, before lowering and before analysis, and
// returns ctx's error at the first of them after ctx is cancelled.
func (s *LSPServer) analyzeSource(ctx context.Context, uri string, source []byte, prev *document) (*document, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var root *parse.Node
	doc := &document{}
	if path, ok := uriPath(uri); ok {
		// ponytail: @INC is resolved on every parse; cache it per directory
		// if resolution shows up in a profile.
		var roots []string
		roots, doc.why = incRoots(path)
		doc.dir = filepath.Dir(path)
		// ponytail: every parse re-reads the modules it uses; cache them by
		// mtime, invalidating facts and seen together, if parses get slow.
		root = parse.NewSession().ParseSourceFrom(path, source, roots...)
	} else {
		root = parse.Parse(source)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var fa *infer.FileAnalysis
	var err error
	if prev != nil && prev.root != nil {
		fa, _, err = analyzeEdit(ctx, prev.root, prev.analysis, root, s.Options)
	} else {
		fa, err = analyze(ctx, root, s.Options)
	}
	if err != nil {
		return nil, err
	}
	doc.analysis, doc.root = fa, root
	return doc, nil
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

// CloseDocument stops the owner of the document at uri and removes it,
// releasing its lowered tree, inference annotations, diagnostics, and
// symbols. Closing a document that is not open is a no-op.
func (s *LSPServer) CloseDocument(uri string) {
	s.mu.Lock()
	o := s.documents[uri]
	delete(s.documents, uri)
	s.mu.Unlock()
	if o != nil {
		o.stop()
	}
}

// Shutdown stops every owner and cancels every parse in flight. Nothing is
// published once it returns.
func (s *LSPServer) Shutdown() {
	s.mu.Lock()
	owners := s.documents
	s.documents = make(map[string]*owner)
	s.mu.Unlock()
	for _, o := range owners {
		o.stop()
	}
}

// Source returns the raw source bytes for the document at uri, or nil
// if the uri is not currently open. They have no spare capacity, so an
// append to them copies rather than writing into bytes the document shares.
func (s *LSPServer) Source(uri string) []byte {
	if d := s.Text(uri); d != nil {
		return d.Text[:len(d.Text):len(d.Text)]
	}
	return nil
}

// Text returns a snapshot of the text and line table of the document at
// uri, one later edits leave unchanged, or nil if the uri is not open. The
// owner takes it between edits (owner.view).
func (s *LSPServer) Text(uri string) *Document {
	if doc := s.snapshot(uri); doc != nil {
		return doc.text
	}
	return nil
}

// Annotations returns the type of each lowered node of the document at uri,
// keyed by its id, or nil if the uri is not currently open.
func (s *LSPServer) Annotations(uri string) map[lower.ID]types.Type {
	if doc := s.snapshot(uri); doc != nil {
		return doc.analysis.Annotations
	}
	return nil
}

// Diagnostics returns the slice of inference diagnostics for the document at
// uri, or nil if the uri is not currently open.
func (s *LSPServer) Diagnostics(uri string) []infer.Diagnostic {
	if doc := s.snapshot(uri); doc != nil {
		return doc.analysis.Diagnostics
	}
	return nil
}

// SymbolTable returns the symbol table for the document at uri, or nil if the
// uri is not currently open.
func (s *LSPServer) SymbolTable(uri string) *infer.SymbolTable {
	if doc := s.snapshot(uri); doc != nil {
		return doc.analysis.Symbols
	}
	return nil
}

// nodeAt returns the innermost lowered node of the document at uri whose
// bytes cover offset, through the position map: bytes the lowering erased,
// such as a paren, resolve to the node they belonged to. It returns nil if
// uri is not open or no node covers offset.
func (s *LSPServer) nodeAt(uri string, offset uint32) (*document, lower.Node) {
	doc := s.snapshot(uri)
	if doc == nil {
		return nil, nil
	}
	return doc, doc.analysis.Positions.At(int(offset))
}

// TypeAtByte returns the inferred type of the innermost lowered node at the
// given byte offset in the document at uri. A Hole, source perl rejects, is
// typed Any and never inspected (§4.14.2), so it reports nothing.
// Returns (Unknown, false) if uri is not open or no node has a type there.
func (s *LSPServer) TypeAtByte(uri string, offset uint32) (types.Type, bool) {
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
