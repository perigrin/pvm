// ABOUTME: Language Server Protocol implementation for PSC
// ABOUTME: Provides editor integration for type checking and code analysis

package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"path/filepath"
	"strings"
	"sync"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser"
)

// LSP error codes
const (
	ErrServerInit     = "LSP-001" // Failed to initialize server
	ErrMethodNotFound = "LSP-002" // Method not found
	ErrParseRequest   = "LSP-003" // Failed to parse request
	ErrInvalidParams  = "LSP-004" // Invalid parameters
)

// LSP Protocol constants following specification
const (
	// JSON-RPC version
	JSONRPCVersion = "2.0"

	// LSP Protocol version
	LSPVersion = "3.17"
)

// Server represents an LSP server instance
type Server struct {
	// Connection for reading/writing LSP messages
	conn io.ReadWriteCloser

	// Parser for type checking
	typeChecker *parser.TypeCheck

	// Workspace documents
	documents map[string]*Document

	// Server capabilities
	capabilities *ServerCapabilities

	// Mutex for thread safety
	mutex sync.RWMutex

	// Context for server lifecycle
	ctx    context.Context
	cancel context.CancelFunc

	// Logger for debugging
	logger *log.Logger

	// Server state
	initialized bool
	shutdown    bool
}

// Document represents a document managed by the LSP server
type Document struct {
	URI     string
	Text    string
	Version int
	AST     *parser.AST
	Errors  []parser.TypeCheckError
}

// ServerCapabilities defines what the server can do
type ServerCapabilities struct {
	TextDocumentSync    *TextDocumentSyncOptions `json:"textDocumentSync,omitempty"`
	HoverProvider       bool                     `json:"hoverProvider,omitempty"`
	CompletionProvider  *CompletionOptions       `json:"completionProvider,omitempty"`
	DiagnosticsProvider bool                     `json:"diagnosticsProvider,omitempty"`
}

// TextDocumentSyncOptions defines text synchronization capabilities
type TextDocumentSyncOptions struct {
	OpenClose bool `json:"openClose"`
	Change    int  `json:"change"` // 1 = Full, 2 = Incremental
}

// CompletionOptions defines completion capabilities
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
	ResolveProvider   bool     `json:"resolveProvider,omitempty"`
}

// NewServer creates a new LSP server
func NewServer(conn io.ReadWriteCloser) (*Server, error) {
	tc, err := parser.NewTypeCheck()
	if err != nil {
		return nil, errors.NewSystemError(
			ErrServerInit,
			"Failed to create type checker for LSP server",
			err,
		)
	}

	ctx, cancel := context.WithCancel(context.Background())

	server := &Server{
		conn:        conn,
		typeChecker: tc,
		documents:   make(map[string]*Document),
		capabilities: &ServerCapabilities{
			TextDocumentSync: &TextDocumentSyncOptions{
				OpenClose: true,
				Change:    1, // Full text sync
			},
			HoverProvider: true,
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{"$", "@", "%", ":", ".", "->"},
				ResolveProvider:   false,
			},
			DiagnosticsProvider: true,
		},
		ctx:    ctx,
		cancel: cancel,
		logger: log.New(io.Discard, "[LSP] ", log.LstdFlags),
	}

	return server, nil
}

// Start starts the LSP server
func (s *Server) Start() error {
	s.logger.Println("Starting LSP server")

	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
			if err := s.handleMessage(); err != nil {
				if err == io.EOF {
					s.logger.Println("Client disconnected")
					return nil
				}
				s.logger.Printf("Error handling message: %v", err)
				return err
			}
		}
	}
}

// Stop stops the LSP server
func (s *Server) Stop() {
	s.logger.Println("Stopping LSP server")
	s.cancel()
	if s.conn != nil {
		_ = s.conn.Close()
	}
}

// SetLogger sets the logger for the server
func (s *Server) SetLogger(logger *log.Logger) {
	s.logger = logger
}

// handleMessage handles a single LSP message
func (s *Server) handleMessage() error {
	// Read the Content-Length header
	var contentLength int
	for {
		var line string
		if _, err := fmt.Fscanln(s.conn, &line); err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break // End of headers
		}

		if strings.HasPrefix(line, "Content-Length:") {
			if _, err := fmt.Sscanf(line, "Content-Length: %d", &contentLength); err != nil {
				return errors.NewSystemError(
					ErrParseRequest,
					"Failed to parse Content-Length header",
					err,
				)
			}
		}
	}

	if contentLength == 0 {
		return errors.NewSystemError(
			ErrParseRequest,
			"No Content-Length header found",
			nil,
		)
	}

	// Read the message body
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(s.conn, body); err != nil {
		return err
	}

	// Parse the JSON-RPC message
	var msg JSONRPCMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return errors.NewSystemError(
			ErrParseRequest,
			"Failed to parse JSON-RPC message",
			err,
		)
	}

	s.logger.Printf("Received message: %s", msg.Method)

	// Handle the message
	return s.dispatchMessage(&msg)
}

// dispatchMessage dispatches a message to the appropriate handler
func (s *Server) dispatchMessage(msg *JSONRPCMessage) error {
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "initialized":
		return s.handleInitialized(msg)
	case "shutdown":
		return s.handleShutdown(msg)
	case "exit":
		return s.handleExit(msg)
	case "textDocument/didOpen":
		return s.handleTextDocumentDidOpen(msg)
	case "textDocument/didChange":
		return s.handleTextDocumentDidChange(msg)
	case "textDocument/didClose":
		return s.handleTextDocumentDidClose(msg)
	case "textDocument/hover":
		return s.handleTextDocumentHover(msg)
	case "textDocument/completion":
		return s.handleTextDocumentCompletion(msg)
	default:
		return s.sendError(msg.ID, -32601, "Method not found", nil)
	}
}

// sendResponse sends a JSON-RPC response
func (s *Server) sendResponse(id interface{}, result interface{}) error {
	response := JSONRPCResponse{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  result,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return s.sendMessage(data)
}

// sendError sends a JSON-RPC error response
func (s *Server) sendError(id interface{}, code int, message string, data interface{}) error {
	response := JSONRPCResponse{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return s.sendMessage(responseData)
}

// sendNotification sends a JSON-RPC notification
func (s *Server) sendNotification(method string, params interface{}) error {
	notification := JSONRPCNotification{
		JSONRPC: JSONRPCVersion,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	return s.sendMessage(data)
}

// sendMessage sends a message using LSP protocol format
func (s *Server) sendMessage(data []byte) error {
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := s.conn.Write([]byte(header)); err != nil {
		return err
	}
	if _, err := s.conn.Write(data); err != nil {
		return err
	}
	return nil
}

// StartTCPServer starts the LSP server on a TCP connection
func StartTCPServer(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return errors.NewSystemError(
			ErrServerInit,
			fmt.Sprintf("Failed to listen on %s", address),
			err,
		)
	}
	defer func() { _ = listener.Close() }()

	log.Printf("LSP server listening on %s", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go func() {
			server, err := NewServer(conn)
			if err != nil {
				log.Printf("Failed to create server: %v", err)
				_ = conn.Close()
				return
			}

			if err := server.Start(); err != nil {
				log.Printf("Server error: %v", err)
			}
		}()
	}
}

// StartStdioServer starts the LSP server using stdin/stdout
func StartStdioServer() error {
	server, err := NewServer(NewStdioConnection())
	if err != nil {
		return err
	}

	return server.Start()
}

// getDocument retrieves a document from the server's document store
func (s *Server) getDocument(uri string) (*Document, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	doc, exists := s.documents[uri]
	return doc, exists
}

// setDocument stores a document in the server's document store
func (s *Server) setDocument(uri string, doc *Document) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.documents[uri] = doc
}

// removeDocument removes a document from the server's document store
func (s *Server) removeDocument(uri string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.documents, uri)
}

// analyzeDocument performs type checking on a document
func (s *Server) analyzeDocument(doc *Document) error {
	// Convert URI to file path
	filePath := uriToPath(doc.URI)

	// Perform type checking
	result, err := s.typeChecker.CheckFile(filePath)
	if err != nil {
		s.logger.Printf("Type checking failed for %s: %v", filePath, err)
		// Don't return error - we still want to store the document
	}

	if result != nil {
		doc.Errors = result.Errors
	} else {
		doc.Errors = []parser.TypeCheckError{}
	}

	return nil
}

// publishDiagnostics sends diagnostics to the client
func (s *Server) publishDiagnostics(uri string, errors []parser.TypeCheckError) error {
	diagnostics := make([]Diagnostic, len(errors))
	for i, err := range errors {
		diagnostics[i] = Diagnostic{
			Range: Range{
				Start: Position{Line: err.Line - 1, Character: err.Column - 1}, // LSP is 0-based
				End:   Position{Line: err.Line - 1, Character: err.Column - 1 + len(err.Message)},
			},
			Severity: &[]DiagnosticSeverity{DiagnosticSeverityError}[0],
			Code:     "type-error",
			Source:   "psc",
			Message:  err.Message,
		}
	}

	params := PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	}

	return s.sendNotification("textDocument/publishDiagnostics", params)
}

// uriToPath converts a file URI to a file path
func uriToPath(uri string) string {
	if strings.HasPrefix(uri, "file://") {
		return uri[7:] // Remove "file://" prefix
	}
	return uri
}

// pathToURI converts a file path to a file URI
func pathToURI(path string) string {
	abs, _ := filepath.Abs(path)
	return "file://" + abs
}
