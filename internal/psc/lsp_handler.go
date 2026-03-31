// ABOUTME: LSP handler dispatch bridging JSON-RPC requests to the PSC LSPServer.
// ABOUTME: Routes lifecycle, document sync, and query methods to existing inference infrastructure.

package psc

import (
	"encoding/json"

	"tamarou.com/pvm/internal/infer"
)

// handler dispatches incoming JSON-RPC messages to the appropriate LSP method
// implementation. It tracks initialization and shutdown state required by the
// LSP lifecycle.
type handler struct {
	server      *LSPServer
	transport   *transport
	sources     map[string][]byte // document source for position conversion
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
		sources:   make(map[string][]byte),
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

	// Query methods — implemented in Tasks 6-8.
	case "textDocument/hover", "textDocument/definition", "textDocument/documentSymbol":
		return nil

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

// handleDidOpen handles textDocument/didOpen notifications. It stores the
// source text, runs inference via OpenDocument, and publishes diagnostics.
// If OpenDocument fails, a single error diagnostic spanning the whole
// document is published instead.
func (h *handler) handleDidOpen(msg *jsonRPCMessage) error {
	var params lspDidOpenParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return err
	}

	uri := params.TextDocument.URI
	source := []byte(params.TextDocument.Text)
	h.sources[uri] = source

	if err := h.server.OpenDocument(uri, source); err != nil {
		// Publish a single error diagnostic covering the whole document.
		docEnd := offsetToPosition(source, uint32(len(source)))
		h.transport.sendNotification("textDocument/publishDiagnostics", lspPublishDiagnosticsParams{
			URI: uri,
			Diagnostics: []lspDiagnostic{
				{
					Range:    lspRange{Start: lspPosition{Line: 0, Character: 0}, End: docEnd},
					Severity: 1, // Error
					Source:   "psc",
					Message:  err.Error(),
				},
			},
		})
		return nil
	}

	h.publishDiagnostics(uri, source)
	return nil
}

// handleDidChange handles textDocument/didChange notifications. It takes the
// full text from the first content change event (full sync mode), updates the
// source store, re-runs inference, and publishes diagnostics.
func (h *handler) handleDidChange(msg *jsonRPCMessage) error {
	var params lspDidChangeParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return err
	}

	uri := params.TextDocument.URI
	if len(params.ContentChanges) == 0 {
		return nil
	}

	source := []byte(params.ContentChanges[0].Text)
	h.sources[uri] = source

	if err := h.server.OpenDocument(uri, source); err != nil {
		docEnd := offsetToPosition(source, uint32(len(source)))
		h.transport.sendNotification("textDocument/publishDiagnostics", lspPublishDiagnosticsParams{
			URI: uri,
			Diagnostics: []lspDiagnostic{
				{
					Range:    lspRange{Start: lspPosition{Line: 0, Character: 0}, End: docEnd},
					Severity: 1,
					Source:   "psc",
					Message:  err.Error(),
				},
			},
		})
		return nil
	}

	h.publishDiagnostics(uri, source)
	return nil
}

// handleDidClose handles textDocument/didClose notifications. It removes the
// document from both the source store and the LSP server.
func (h *handler) handleDidClose(msg *jsonRPCMessage) error {
	var params lspDidCloseParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return err
	}

	uri := params.TextDocument.URI
	h.server.CloseDocument(uri)
	delete(h.sources, uri)
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
			severity = 1
		case infer.Warning:
			severity = 2
		case infer.Info:
			severity = 3
		default:
			severity = 1
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

// positionToOffset converts a zero-based line and character position to a
// byte offset within source. Returns len(source) if line exceeds the
// document's line count.
func positionToOffset(source []byte, line, col int) uint32 {
	offset := 0
	currentLine := 0
	for i, b := range source {
		if currentLine == line {
			offset = i + col
			break
		}
		if b == '\n' {
			currentLine++
		}
	}
	if currentLine < line {
		return uint32(len(source))
	}
	return uint32(offset)
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
