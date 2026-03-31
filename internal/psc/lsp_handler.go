// ABOUTME: LSP handler dispatch bridging JSON-RPC requests to the PSC LSPServer.
// ABOUTME: Routes lifecycle, document sync, and query methods to existing inference infrastructure.

package psc

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

	// Document sync — implemented in Task 5.
	case "textDocument/didOpen", "textDocument/didChange", "textDocument/didClose":
		return nil

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
