// ABOUTME: LSP handler dispatch bridging JSON-RPC requests to the PSC LSPServer.
// ABOUTME: Routes lifecycle, document sync, and query methods to existing inference infrastructure.

package psc

import (
	"encoding/json"
	"fmt"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/types"
)

// handler dispatches incoming JSON-RPC messages to the appropriate LSP method
// implementation. It tracks initialization and shutdown state required by the
// LSP lifecycle.
type handler struct {
	server      *LSPServer
	transport   *transport
	initialized bool
	isShutdown  bool
	exitFn      func(int) // injectable for testing; defaults to os.Exit
}

// newHandler creates a handler wired to the given server and transport.
// exitFn defaults to a no-op; callers may replace it for testing.
func newHandler(server *LSPServer, tr *transport) *handler {
	return &handler{
		server:    server,
		transport: tr,
		exitFn:    func(int) {}, // replaced by os.Exit in production
	}
}

// handle dispatches a single JSON-RPC message to the correct handler method.
// "exit" and "initialize" are valid before the server is fully initialized.
// All other requests sent before "initialize" receive a ServerNotInitialized
// error. Notifications (no ID) before initialize are silently dropped.
func (h *handler) handle(msg *jsonRPCMessage) error {
	// "exit" is always handled — it is valid both before initialize and
	// after shutdown.
	if msg.Method == "exit" {
		if h.isShutdown {
			h.exitFn(0)
		} else {
			h.exitFn(1)
		}
		return nil
	}

	// "initialize" is the bootstrap handshake and must be handled even
	// before h.initialized is set.
	if msg.Method == "initialize" {
		return h.handleInitialize(msg)
	}

	// Reject any request (has an ID) sent after shutdown per LSP spec.
	// Notifications (no ID) are silently dropped.
	if h.isShutdown {
		if msg.ID != nil {
			h.transport.sendError(msg.ID, codeInvalidRequest, "server is shutting down")
		}
		return nil
	}

	// Reject any request (has an ID) that arrives before initialize completes.
	if !h.initialized {
		if msg.ID != nil {
			h.transport.sendError(msg.ID, codeServerNotInitialized, "server not yet initialized")
		}
		return nil
	}

	switch msg.Method {
	case "initialized":
		// Client confirmation notification — no action required.
		return nil
	case "shutdown":
		return h.handleShutdown(msg)

	case "textDocument/didOpen":
		return h.handleDidOpen(msg)
	case "textDocument/didChange":
		return h.handleDidChange(msg)
	case "textDocument/didClose":
		return h.handleDidClose(msg)

	// Query methods.
	case "textDocument/hover":
		return h.handleHover(msg)
	case "textDocument/definition":
		return h.handleDefinition(msg)
	case "textDocument/documentSymbol":
		return h.handleDocumentSymbol(msg)

	default:
		if msg.ID != nil {
			h.transport.sendError(msg.ID, codeMethodNotFound, "method not found: "+msg.Method)
		}
		return nil
	}
}

// handleInitialize responds to the LSP initialize request by advertising
// server capabilities and marking the handler as ready.
func (h *handler) handleInitialize(msg *jsonRPCMessage) error {
	h.initialized = true
	result := lspInitializeResult{
		Capabilities: lspServerCapabilities{
			TextDocumentSync:       1, // full sync
			HoverProvider:          true,
			DefinitionProvider:     true,
			DocumentSymbolProvider: true,
		},
	}
	h.transport.sendResponse(msg.ID, result)
	return nil
}

// handleShutdown responds to the LSP shutdown request. The server remains
// running until the subsequent "exit" notification is received.
func (h *handler) handleShutdown(msg *jsonRPCMessage) error {
	h.isShutdown = true
	h.transport.sendResponse(msg.ID, nil)
	return nil
}

// updateDocument calls OpenDocument on the server for the given uri and
// source. On success it publishes diagnostics. On failure it publishes a
// single error diagnostic spanning the whole document.
func (h *handler) updateDocument(uri string, source []byte) {
	if err := h.server.OpenDocument(uri, source); err != nil {
		// Publish a single error diagnostic covering the whole document.
		docEnd := offsetToPosition(source, uint32(len(source)))
		h.transport.sendNotification("textDocument/publishDiagnostics", lspPublishDiagnosticsParams{
			URI: uri,
			Diagnostics: []lspDiagnostic{{
				Range:    lspRange{Start: lspPosition{}, End: docEnd},
				Severity: lspSeverityError,
				Source:   "psc",
				Message:  err.Error(),
			}},
		})
		return
	}
	h.publishDiagnostics(uri, source)
}

// handleDidOpen handles textDocument/didOpen notifications. It runs inference
// via OpenDocument and publishes diagnostics. If OpenDocument fails, a single
// error diagnostic spanning the whole document is published instead.
func (h *handler) handleDidOpen(msg *jsonRPCMessage) error {
	var params lspDidOpenParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil
	}

	h.updateDocument(params.TextDocument.URI, []byte(params.TextDocument.Text))
	return nil
}

// handleDidChange handles textDocument/didChange notifications. It takes the
// full text from the first content change event (full sync mode), re-runs
// inference, and publishes diagnostics.
func (h *handler) handleDidChange(msg *jsonRPCMessage) error {
	var params lspDidChangeParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil
	}

	if len(params.ContentChanges) == 0 {
		return nil
	}

	h.updateDocument(params.TextDocument.URI, []byte(params.ContentChanges[0].Text))
	return nil
}

// handleDidClose handles textDocument/didClose notifications. It removes the
// document from the LSP server.
func (h *handler) handleDidClose(msg *jsonRPCMessage) error {
	var params lspDidCloseParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil
	}

	h.server.CloseDocument(params.TextDocument.URI)
	return nil
}

// publishDiagnostics retrieves inference diagnostics for uri from the server,
// converts them to LSP diagnostic format, and sends a
// textDocument/publishDiagnostics notification to the client.
func (h *handler) publishDiagnostics(uri string, source []byte) {
	diags := h.server.Diagnostics(uri)

	lspDiags := make([]lspDiagnostic, 0, len(diags))
	for _, d := range diags {
		var severity int
		switch d.Severity {
		case infer.Error:
			severity = lspSeverityError
		case infer.Warning:
			severity = lspSeverityWarning
		case infer.Info:
			severity = lspSeverityInformation
		default:
			severity = lspSeverityError
		}

		message := d.Message
		if d.Suggestion != "" {
			message += "\nhint: " + d.Suggestion
		}

		lspDiags = append(lspDiags, lspDiagnostic{
			Range: lspRange{
				Start: offsetToPosition(source, d.StartByte),
				End:   offsetToPosition(source, d.EndByte),
			},
			Severity: severity,
			Source:   "psc",
			Message:  message,
			Code:     d.Code,
		})
	}

	h.transport.sendNotification("textDocument/publishDiagnostics", lspPublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: lspDiags,
	})
}

// handleHover handles textDocument/hover requests. It resolves the inferred
// type at the cursor position and returns it as a markdown hover card. If no
// type annotation is found, it returns a null result per the LSP specification.
func (h *handler) handleHover(msg *jsonRPCMessage) error {
	var params lspTextDocumentPositionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		h.transport.sendError(msg.ID, codeInvalidParams, "invalid params: "+err.Error())
		return nil
	}
	uri := params.TextDocument.URI
	source := h.server.Source(uri)
	if source == nil {
		h.transport.sendResponse(msg.ID, nil)
		return nil
	}

	offset := positionToOffset(source, params.Position.Line, params.Position.Character)
	typ, ok := h.server.TypeAtByte(uri, offset)
	if !ok || typ == types.Unknown {
		h.transport.sendResponse(msg.ID, nil)
		return nil
	}

	h.transport.sendResponse(msg.ID, lspHover{
		Contents: lspMarkupContent{
			Kind:  "markdown",
			Value: fmt.Sprintf("**Type:** `%s`", typ.String()),
		},
	})
	return nil
}

// handleDefinition handles textDocument/definition requests. It resolves the
// symbol at the cursor position and returns the declaration location. If no
// symbol is found, it returns a null result per the LSP specification.
func (h *handler) handleDefinition(msg *jsonRPCMessage) error {
	var params lspTextDocumentPositionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		h.transport.sendError(msg.ID, codeInvalidParams, "invalid params: "+err.Error())
		return nil
	}
	uri := params.TextDocument.URI
	source := h.server.Source(uri)
	if source == nil {
		h.transport.sendResponse(msg.ID, nil)
		return nil
	}

	offset := positionToOffset(source, params.Position.Line, params.Position.Character)
	sym, ok := h.server.DefinitionAtByte(uri, offset)
	if !ok {
		h.transport.sendResponse(msg.ID, nil)
		return nil
	}

	h.transport.sendResponse(msg.ID, lspLocation{
		URI: uri, // same document — single-file only
		Range: lspRange{
			Start: offsetToPosition(source, sym.StartByte),
			End:   offsetToPosition(source, sym.EndByte),
		},
	})
	return nil
}

// handleDocumentSymbol handles textDocument/documentSymbol requests. It
// returns all symbols from the document's symbol table converted to LSP
// DocumentSymbol format. An empty array is returned if the document has no
// symbol table or is not open.
func (h *handler) handleDocumentSymbol(msg *jsonRPCMessage) error {
	var params lspDocumentSymbolParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		h.transport.sendError(msg.ID, codeInvalidParams, "invalid params: "+err.Error())
		return nil
	}
	uri := params.TextDocument.URI
	source := h.server.Source(uri)
	st := h.server.SymbolTable(uri)
	if st == nil || source == nil {
		h.transport.sendResponse(msg.ID, []lspDocumentSymbol{})
		return nil
	}

	allSyms := st.AllSymbols()
	result := make([]lspDocumentSymbol, 0, len(allSyms))
	for _, sym := range allSyms {
		r := lspRange{
			Start: offsetToPosition(source, sym.StartByte),
			End:   offsetToPosition(source, sym.EndByte),
		}
		result = append(result, lspDocumentSymbol{
			Name:           sym.Name,
			Kind:           symbolKindForInfer(sym.Kind),
			Range:          r,
			SelectionRange: r, // Phase 1: selectionRange == range
		})
	}
	h.transport.sendResponse(msg.ID, result)
	return nil
}

// symbolKindForInfer maps an infer.SymbolKind to the corresponding LSP symbol
// kind constant. Unknown kinds fall back to symbolKindVariable.
func symbolKindForInfer(kind infer.SymbolKind) int {
	switch kind {
	case infer.SymVariable:
		return symbolKindVariable
	case infer.SymSubroutine:
		return symbolKindFunction
	case infer.SymPackage:
		return symbolKindModule
	case infer.SymMethod:
		return symbolKindMethod
	default:
		return symbolKindVariable
	}
}

// positionToOffset converts a zero-based line and character position to a
// byte offset within source. Negative col values are treated as 0. If the
// computed offset would exceed len(source) it is clamped to len(source).
// Returns len(source) if line exceeds the document's line count.
func positionToOffset(source []byte, line, col int) uint32 {
	if col < 0 {
		col = 0
	}
	currentLine := 0
	for i, b := range source {
		if currentLine == line {
			offset := i + col
			if offset > len(source) {
				return uint32(len(source))
			}
			return uint32(offset)
		}
		if b == '\n' {
			currentLine++
		}
	}
	// Target line starts at end of source (e.g. after the final newline).
	if currentLine == line {
		offset := len(source) + col
		if offset > len(source) {
			return uint32(len(source))
		}
		return uint32(offset)
	}
	return uint32(len(source))
}

// offsetToPosition converts a byte offset within source to a zero-based
// line/character LSP position.
func offsetToPosition(source []byte, offset uint32) lspPosition {
	line, col := 0, 0
	for i := uint32(0); i < offset && int(i) < len(source); i++ {
		if source[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return lspPosition{Line: line, Character: col}
}
