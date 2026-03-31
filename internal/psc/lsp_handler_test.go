// ABOUTME: Tests for the LSP handler dispatch and lifecycle management.
// ABOUTME: Verifies initialize, shutdown, exit, and error handling behavior.

package psc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeRequest builds a Content-Length-framed JSON-RPC request message and
// writes it into buf so a transport can read it back.
func makeRequest(t *testing.T, buf *bytes.Buffer, id int, method string, params interface{}) {
	t.Helper()
	rawParams, err := json.Marshal(params)
	require.NoError(t, err)

	rawID := json.RawMessage(fmt.Sprintf("%d", id))
	msg := jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &rawID,
		Method:  method,
		Params:  rawParams,
	}
	body, err := json.Marshal(msg)
	require.NoError(t, err)

	frame := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), string(body))
	buf.WriteString(frame)
}

// makeNotification builds a Content-Length-framed JSON-RPC notification (no
// ID) and writes it into buf.
func makeNotification(t *testing.T, buf *bytes.Buffer, method string, params interface{}) {
	t.Helper()
	rawParams, err := json.Marshal(params)
	require.NoError(t, err)

	msg := jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  method,
		Params:  rawParams,
	}
	body, err := json.Marshal(msg)
	require.NoError(t, err)

	frame := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), string(body))
	buf.WriteString(frame)
}

// newTestHandler creates a handler wired to a pair of in-memory buffers and
// returns the handler plus the output buffer so tests can inspect responses.
func newTestHandler(t *testing.T) (*handler, *bytes.Buffer) {
	t.Helper()
	server := NewLSPServer()
	var out bytes.Buffer
	tr := newTransport(bytes.NewReader(nil), &out)
	h := newHandler(server, tr)
	return h, &out
}

func TestHandlerInitializeShutdown(t *testing.T) {
	h, out := newTestHandler(t)

	// Send initialize request.
	rawID := json.RawMessage(`1`)
	initMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &rawID,
		Method:  "initialize",
		Params:  json.RawMessage(`{"processId":0,"rootUri":""}`),
	}
	err := h.handle(initMsg)
	require.NoError(t, err)
	assert.True(t, h.initialized, "handler should be initialized after initialize request")

	// Read and validate the initialize response.
	resp := readResponse(t, out)
	require.NotNil(t, resp)
	assert.NotNil(t, resp.ID)
	require.Nil(t, resp.Error, "initialize should not return an error")

	var result lspInitializeResult
	require.NoError(t, json.Unmarshal(resp.Result, &result))
	assert.Equal(t, 1, result.Capabilities.TextDocumentSync)
	assert.True(t, result.Capabilities.HoverProvider)
	assert.True(t, result.Capabilities.DefinitionProvider)
	assert.True(t, result.Capabilities.DocumentSymbolProvider)

	// Send shutdown request.
	rawID2 := json.RawMessage(`2`)
	shutdownMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &rawID2,
		Method:  "shutdown",
	}
	err = h.handle(shutdownMsg)
	require.NoError(t, err)
	assert.True(t, h.isShutdown, "handler should be shut down after shutdown request")

	// Read and validate the shutdown response (result should be null).
	resp2 := readResponse(t, out)
	require.NotNil(t, resp2)
	require.Nil(t, resp2.Error, "shutdown should not return an error")
	assert.Equal(t, json.RawMessage("null"), resp2.Result)
}

func TestHandlerUnknownMethod(t *testing.T) {
	h, out := newTestHandler(t)
	h.initialized = true

	rawID := json.RawMessage(`42`)
	msg := &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &rawID,
		Method:  "textDocument/unknownThing",
	}
	err := h.handle(msg)
	require.NoError(t, err)

	resp := readResponse(t, out)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Error, "unknown method with ID should produce an error response")
	assert.Equal(t, codeMethodNotFound, resp.Error.Code)
}

func TestHandlerRequestBeforeInitialize(t *testing.T) {
	h, out := newTestHandler(t)
	// Do NOT call initialize — handler.initialized remains false.

	rawID := json.RawMessage(`7`)
	msg := &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &rawID,
		Method:  "textDocument/hover",
		Params:  json.RawMessage(`{"textDocument":{"uri":"file:///a.pl"},"position":{"line":0,"character":0}}`),
	}
	err := h.handle(msg)
	require.NoError(t, err)

	resp := readResponse(t, out)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Error, "request before initialize should produce an error response")
	assert.Equal(t, codeServerNotInitialized, resp.Error.Code)
}

func TestHandlerExitWithoutShutdown(t *testing.T) {
	h, _ := newTestHandler(t)
	h.initialized = true

	var capturedCode int
	h.exitFn = func(code int) { capturedCode = code }

	msg := &jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  "exit",
	}
	_ = h.handle(msg)

	assert.Equal(t, 1, capturedCode, "exit without prior shutdown should call exitFn with code 1")
}

func TestHandlerExitAfterShutdown(t *testing.T) {
	h, _ := newTestHandler(t)
	h.initialized = true
	h.isShutdown = true

	var capturedCode int
	h.exitFn = func(code int) { capturedCode = code }

	msg := &jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  "exit",
	}
	_ = h.handle(msg)

	assert.Equal(t, 0, capturedCode, "exit after shutdown should call exitFn with code 0")
}

func TestHandlerDidOpenPublishesDiagnostics(t *testing.T) {
	h, out := newTestHandler(t)
	h.initialized = true

	// "push();\n" is known to produce an arity-mismatch diagnostic.
	params := lspDidOpenParams{
		TextDocument: lspTextDocumentItem{
			URI:        "file:///test.pl",
			LanguageID: "perl",
			Version:    1,
			Text:       "push();\n",
		},
	}
	rawParams, err := json.Marshal(params)
	require.NoError(t, err)

	msg := &jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  "textDocument/didOpen",
		Params:  rawParams,
	}
	require.NoError(t, h.handle(msg))

	// The handler must send a textDocument/publishDiagnostics notification.
	notification := readResponse(t, out)
	require.NotNil(t, notification)
	assert.Equal(t, "textDocument/publishDiagnostics", notification.Method)

	var diagParams lspPublishDiagnosticsParams
	require.NoError(t, json.Unmarshal(notification.Params, &diagParams))
	assert.Equal(t, "file:///test.pl", diagParams.URI)
	assert.NotEmpty(t, diagParams.Diagnostics, "push() with no args should produce diagnostics")
}

func TestHandlerDidChange(t *testing.T) {
	h, out := newTestHandler(t)
	h.initialized = true

	// Open a clean file first.
	openParams := lspDidOpenParams{
		TextDocument: lspTextDocumentItem{
			URI:        "file:///test.pl",
			LanguageID: "perl",
			Version:    1,
			Text:       "my $x = 42;\n",
		},
	}
	rawOpen, err := json.Marshal(openParams)
	require.NoError(t, err)

	openMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  "textDocument/didOpen",
		Params:  rawOpen,
	}
	require.NoError(t, h.handle(openMsg))
	// Drain the publishDiagnostics notification from didOpen.
	_ = readResponse(t, out)
	out.Reset()

	// Now send didChange with bad source.
	changeParams := lspDidChangeParams{
		TextDocument: lspVersionedTextDocumentIdentifier{
			URI:     "file:///test.pl",
			Version: 2,
		},
		ContentChanges: []lspTextDocumentContentChangeEvent{
			{Text: "push();\n"},
		},
	}
	rawChange, err := json.Marshal(changeParams)
	require.NoError(t, err)

	changeMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  "textDocument/didChange",
		Params:  rawChange,
	}
	require.NoError(t, h.handle(changeMsg))

	notification := readResponse(t, out)
	require.NotNil(t, notification)
	assert.Equal(t, "textDocument/publishDiagnostics", notification.Method)

	var diagParams lspPublishDiagnosticsParams
	require.NoError(t, json.Unmarshal(notification.Params, &diagParams))
	assert.Equal(t, "file:///test.pl", diagParams.URI)
	assert.NotEmpty(t, diagParams.Diagnostics)
}

func TestHandlerDidClose(t *testing.T) {
	h, out := newTestHandler(t)
	h.initialized = true

	uri := "file:///test.pl"

	// Open first.
	openParams := lspDidOpenParams{
		TextDocument: lspTextDocumentItem{
			URI:        uri,
			LanguageID: "perl",
			Version:    1,
			Text:       "my $x = 42;\n",
		},
	}
	rawOpen, err := json.Marshal(openParams)
	require.NoError(t, err)

	openMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  "textDocument/didOpen",
		Params:  rawOpen,
	}
	require.NoError(t, h.handle(openMsg))
	// Drain the publishDiagnostics notification.
	_ = readResponse(t, out)

	// Now close.
	closeParams := lspDidCloseParams{
		TextDocument: lspTextDocumentIdentifier{URI: uri},
	}
	rawClose, err := json.Marshal(closeParams)
	require.NoError(t, err)

	closeMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  "textDocument/didClose",
		Params:  rawClose,
	}
	require.NoError(t, h.handle(closeMsg))

	// Diagnostics should be nil after close.
	assert.Nil(t, h.server.Diagnostics(uri), "Diagnostics should be nil after close")
	// sources map should not contain the uri.
	_, inSources := h.sources[uri]
	assert.False(t, inSources, "sources map should not contain uri after close")
}

func TestPositionConversion(t *testing.T) {
	source := []byte("my $x = 42;\nprint $x;\n")

	// (line=0, col=3) should be offset 3.
	offset := positionToOffset(source, 0, 3)
	assert.Equal(t, uint32(3), offset)

	// offset 3 should round-trip back to (line=0, col=3).
	pos := offsetToPosition(source, 3)
	assert.Equal(t, 0, pos.Line)
	assert.Equal(t, 3, pos.Character)

	// (line=1, col=0) should be offset 12 (length of "my $x = 42;\n").
	offset2 := positionToOffset(source, 1, 0)
	assert.Equal(t, uint32(12), offset2)

	// offset 12 should round-trip back to (line=1, col=0).
	pos2 := offsetToPosition(source, 12)
	assert.Equal(t, 1, pos2.Line)
	assert.Equal(t, 0, pos2.Character)
}
