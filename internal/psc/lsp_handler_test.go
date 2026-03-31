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
	// Source should be nil after close.
	assert.Nil(t, h.server.Source(uri), "Source should be nil after close")
}

// openDoc is a shared test helper that sends a textDocument/didOpen message
// for the given URI and source text, then drains the resulting
// publishDiagnostics notification from out.
func openDoc(t *testing.T, h *handler, out *bytes.Buffer, uri, text string) {
	t.Helper()
	params := lspDidOpenParams{
		TextDocument: lspTextDocumentItem{
			URI:        uri,
			LanguageID: "perl",
			Version:    1,
			Text:       text,
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
	// Drain the publishDiagnostics notification.
	_ = readResponse(t, out)
}

// sendRequest is a shared test helper that sends a JSON-RPC request with the
// given id, method, and params to the handler and returns the response.
func sendRequest(t *testing.T, h *handler, out *bytes.Buffer, id int, method string, params interface{}) *jsonRPCMessage {
	t.Helper()
	rawParams, err := json.Marshal(params)
	require.NoError(t, err)
	rawID := json.RawMessage(fmt.Sprintf("%d", id))
	msg := &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &rawID,
		Method:  method,
		Params:  rawParams,
	}
	require.NoError(t, h.handle(msg))
	return readResponse(t, out)
}

func TestHandlerHover(t *testing.T) {
	h, out := newTestHandler(t)
	h.initialized = true

	const uri = "file:///hover.pl"
	openDoc(t, h, out, uri, "my $x = 42;\n")
	out.Reset()

	resp := sendRequest(t, h, out, 10, "textDocument/hover", map[string]interface{}{
		"textDocument": map[string]string{"uri": uri},
		"position":     map[string]int{"line": 0, "character": 3},
	})

	require.NotNil(t, resp)
	require.Nil(t, resp.Error, "hover should not return an error")
	require.NotNil(t, resp.Result, "hover result must not be nil")
	require.NotEqual(t, json.RawMessage("null"), resp.Result, "hover result must not be JSON null")

	var hover lspHover
	require.NoError(t, json.Unmarshal(resp.Result, &hover))
	assert.Equal(t, "markdown", hover.Contents.Kind)
	assert.Contains(t, hover.Contents.Value, "Type")
}

func TestHandlerHoverNoType(t *testing.T) {
	h, out := newTestHandler(t)
	h.initialized = true

	const uri = "file:///hover_no_type.pl"
	// A comment-only file produces no type annotations anywhere in the document.
	openDoc(t, h, out, uri, "# just a comment\n")
	out.Reset()

	// Position (0, 0) is the '#' — no type annotation exists in this file.
	resp := sendRequest(t, h, out, 11, "textDocument/hover", map[string]interface{}{
		"textDocument": map[string]string{"uri": uri},
		"position":     map[string]int{"line": 0, "character": 0},
	})

	require.NotNil(t, resp)
	require.Nil(t, resp.Error, "hover should not return an error")
	assert.Equal(t, json.RawMessage("null"), resp.Result, "hover at non-type position should return null")
}

func TestHandlerDefinition(t *testing.T) {
	h, out := newTestHandler(t)
	h.initialized = true

	const uri = "file:///def.pl"
	openDoc(t, h, out, uri, "my $x = 42;\nprint $x;\n")
	out.Reset()

	// Position (1, 6) is inside "$x" on line 1 (the usage "print $x;").
	resp := sendRequest(t, h, out, 20, "textDocument/definition", map[string]interface{}{
		"textDocument": map[string]string{"uri": uri},
		"position":     map[string]int{"line": 1, "character": 6},
	})

	require.NotNil(t, resp)
	require.Nil(t, resp.Error, "definition should not return an error")
	require.NotNil(t, resp.Result, "definition result must not be nil")
	require.NotEqual(t, json.RawMessage("null"), resp.Result, "definition result must not be JSON null")

	var loc lspLocation
	require.NoError(t, json.Unmarshal(resp.Result, &loc))
	assert.Equal(t, uri, loc.URI)
	assert.Equal(t, 0, loc.Range.Start.Line, "declaration should be on line 0")
}

func TestHandlerDocumentSymbol(t *testing.T) {
	h, out := newTestHandler(t)
	h.initialized = true

	const uri = "file:///syms.pl"
	openDoc(t, h, out, uri, "my $x = 1;\nsub foo { }\n")
	out.Reset()

	resp := sendRequest(t, h, out, 30, "textDocument/documentSymbol", map[string]interface{}{
		"textDocument": map[string]string{"uri": uri},
	})

	require.NotNil(t, resp)
	require.Nil(t, resp.Error, "documentSymbol should not return an error")
	require.NotNil(t, resp.Result, "documentSymbol result must not be nil")

	var syms []lspDocumentSymbol
	require.NoError(t, json.Unmarshal(resp.Result, &syms))
	assert.GreaterOrEqual(t, len(syms), 2, "should have at least 2 symbols")

	names := make(map[string]bool)
	for _, s := range syms {
		names[s.Name] = true
	}
	assert.True(t, names["$x"], "symbols should include $x")
	assert.True(t, names["foo"], "symbols should include foo")
}

func TestHandlerRequestAfterShutdown(t *testing.T) {
	h, out := newTestHandler(t)
	h.initialized = true
	h.isShutdown = true

	rawID := json.RawMessage(`99`)
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
	require.NotNil(t, resp.Error, "request after shutdown should produce an error response")
	assert.Equal(t, codeInvalidRequest, resp.Error.Code)
}

func TestHandlerNotificationAfterShutdownDropped(t *testing.T) {
	// Notifications (no ID) after shutdown are silently dropped — no response.
	h, _ := newTestHandler(t)
	h.initialized = true
	h.isShutdown = true

	msg := &jsonRPCMessage{
		JSONRPC: "2.0",
		// No ID — this is a notification.
		Method: "textDocument/didOpen",
		Params: json.RawMessage(`{"textDocument":{"uri":"file:///a.pl","languageId":"perl","version":1,"text":""}}`),
	}
	err := h.handle(msg)
	require.NoError(t, err)
}

func TestHandlerHoverMalformedParams(t *testing.T) {
	// Malformed JSON params on a request handler must produce an InvalidParams
	// error response, not crash the server.
	h, out := newTestHandler(t)
	h.initialized = true

	rawID := json.RawMessage(`55`)
	msg := &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &rawID,
		Method:  "textDocument/hover",
		Params:  json.RawMessage(`not-valid-json`),
	}
	err := h.handle(msg)
	require.NoError(t, err, "handler must not return an error for malformed params")

	resp := readResponse(t, out)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Error, "malformed params should produce an error response")
	assert.Equal(t, codeInvalidParams, resp.Error.Code, "error code should be InvalidParams")
}

func TestHandlerDidOpenMalformedParams(t *testing.T) {
	// Malformed JSON params on a notification handler must be silently dropped.
	h, _ := newTestHandler(t)
	h.initialized = true

	msg := &jsonRPCMessage{
		JSONRPC: "2.0",
		// No ID — notification.
		Method: "textDocument/didOpen",
		Params: json.RawMessage(`not-valid-json`),
	}
	err := h.handle(msg)
	require.NoError(t, err, "notification handler must not return an error for malformed params")
}

func TestPositionConversionEdgeCases(t *testing.T) {
	source := []byte("abc\ndef\n")

	// Column beyond the end of line 0 is clamped to len(source).
	offset := positionToOffset(source, 0, 100)
	assert.LessOrEqual(t, offset, uint32(len(source)), "col beyond line should be clamped to document end")

	// Line beyond the end of the document returns len(source).
	offset2 := positionToOffset(source, 99, 0)
	assert.Equal(t, uint32(len(source)), offset2, "line beyond document should return len(source)")

	// Last line (after final newline): line 2 col 0 should be len(source).
	// The source "abc\ndef\n" has lines 0="abc", 1="def", 2="" (after final '\n').
	offset3 := positionToOffset(source, 2, 0)
	assert.Equal(t, uint32(len(source)), offset3, "position at last empty line should be len(source)")

	// Negative col should be treated as 0.
	offset4 := positionToOffset(source, 0, -1)
	assert.Equal(t, uint32(0), offset4, "negative col should be clamped to 0")
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
