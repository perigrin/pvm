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
	"time"

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
	URI         string
	Text        string
	Version     int
	AST         *parser.AST
	Errors      []parser.TypeCheckError
	LastChecked time.Time        // When the document was last type-checked
	LastChanged time.Time        // When the document was last changed
	LineChanges map[int]struct{} // Set of line numbers that have changed since last check
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

	// Update AST
	ast, _ := s.typeChecker.Parser.ParseFile(filePath)
	doc.AST = ast

	return nil
}

// analyzeDocumentIncremental performs incremental type checking on only changed lines
func (s *Server) analyzeDocumentIncremental(doc *Document) error {
	if len(doc.LineChanges) == 0 {
		// No changes to analyze
		return nil
	}

	// Convert URI to file path but we'll only use the temp file path
	_ = uriToPath(doc.URI)

	// Create a temp file for analysis
	tempFilePath, err := s.writeDocumentToTempFile(doc)
	if err != nil {
		return err
	}

	// Get existing errors
	currentErrors := make(map[int][]parser.TypeCheckError)
	var otherErrors []parser.TypeCheckError

	// Group existing errors by line
	for _, err := range doc.Errors {
		if _, ok := doc.LineChanges[err.Line]; ok {
			// This line has changed, so we'll recheck it
			currentErrors[err.Line] = append(currentErrors[err.Line], err)
		} else {
			// This line hasn't changed, preserve the error
			otherErrors = append(otherErrors, err)
		}
	}

	// Get a fresh AST - can't do fully incremental parsing yet
	ast, err := s.typeChecker.Parser.ParseFile(tempFilePath)
	if err != nil {
		return err
	}
	doc.AST = ast

	// Create a simple type checker structure for our incremental analysis
	// This is a simplified version that doesn't need the full TypeChecker capabilities
	typeChecker := struct {
		VariableTypes   map[string]string
		TypeAnnotations map[string]string
	}{
		VariableTypes:   make(map[string]string),
		TypeAnnotations: make(map[string]string),
	}

	// Initialize with necessary base information
	if ast != nil {
		// Process imports
		for _, node := range ast.Root.Children() {
			if strings.Contains(node.Type(), "use_statement") {
				// Process import - in a real implementation, we'd add imported modules
				// This is simplified for the example
				typeChecker.VariableTypes["imported_module"] = "Module"
			}
		}

		// Collect all type annotations without validation
		for _, annotation := range ast.TypeAnnotations {
			// Store the annotation without validation
			if annotation.AnnotatedItem != "" {
				varName := annotation.AnnotatedItem
				typeName := annotation.TypeExpression.String()
				typeChecker.TypeAnnotations[varName] = typeName
				typeChecker.VariableTypes[varName] = typeName
			}
		}
	}

	// Analyze only the changed lines
	var newErrors []parser.TypeCheckError

	// We only need the AST structure for checking nodes at specific lines
	// The rest of the checking is simplified

	// Check each changed line
	for lineNum := range doc.LineChanges {
		// Check nodes on this line - simplified approach just checking position
		// A real implementation would track and check actual nodes at each line
		lineErrors := s.checkElementsAtLine(ast, lineNum, typeChecker)
		newErrors = append(newErrors, lineErrors...)
	}

	// Combine the errors
	var combinedErrors []parser.TypeCheckError
	combinedErrors = append(combinedErrors, otherErrors...)
	combinedErrors = append(combinedErrors, newErrors...)
	doc.Errors = combinedErrors

	// Clear line changes since we've checked them
	doc.LineChanges = make(map[int]struct{})
	doc.LastChecked = time.Now()

	return nil
}

// checkElementsAtLine checks AST elements at a specific line for type errors
func (s *Server) checkElementsAtLine(ast *parser.AST, lineNum int, typeChecker struct {
	VariableTypes   map[string]string
	TypeAnnotations map[string]string
}) []parser.TypeCheckError {
	var errors []parser.TypeCheckError

	// Find nodes at this line
	if ast == nil || ast.Root == nil {
		return errors
	}

	// This is a simplified implementation for incremental checking
	// A full implementation would check declarations, assignments, and type annotations
	// by traversing the AST and checking nodes at the specific line

	// Check for type annotations at this line
	for _, anno := range ast.TypeAnnotations {
		// Check if annotation is at this line
		if anno.Pos.Line == lineNum {
			// Basic validation
			varName := anno.AnnotatedItem
			typeName := "Unknown"
			if anno.TypeExpression != nil {
				typeName = anno.TypeExpression.String()
			}

			// Very basic validation - just check for empty type names
			if typeName == "" || typeName == "Unknown" {
				errors = append(errors, parser.TypeCheckError{
					Message: fmt.Sprintf("invalid type annotation for %s", varName),
					Path:    uriToPath(ast.Path),
					Line:    anno.Pos.Line,
					Column:  anno.Pos.Column,
				})
			}
		}
	}

	// In a real implementation, we would also check variable declarations and assignments
	// by traversing the AST and finding nodes at the specific line

	return errors
}

// Note: The collectNodeLines function is no longer needed as we've simplified the implementation

// We use the file path directly instead of extracting module names

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

// Note: Using the writeDocumentToTempFile method defined in handlers.go
