// ABOUTME: Integration test for the PSC LSP server protocol layer.
// ABOUTME: Verifies full message lifecycle via io.Pipe with real parsing and inference.

package psc

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLSPIntegration exercises the full LSP message lifecycle by driving the
// handler directly with constructed jsonRPCMessage structs and inspecting the
// responses written to an in-memory buffer. No subprocess is spawned.
func TestLSPIntegration(t *testing.T) {
	server := NewLSPServer()
	var out bytes.Buffer
	tr := newTransport(bytes.NewReader(nil), &out)
	h := newHandler(server, tr)

	const uri = "file:///integration.pl"
	const source = "my $x = 42;\npush();\n"

	// -------------------------------------------------------------------------
	// Step 1: Initialize — verify capability advertisement.
	// -------------------------------------------------------------------------
	rawID1 := json.RawMessage(`1`)
	initMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &rawID1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"processId":0,"rootUri":""}`),
	}
	require.NoError(t, h.handle(initMsg), "initialize must not error")

	initResp := readResponse(t, &out)
	require.NotNil(t, initResp)
	require.Nil(t, initResp.Error, "initialize response must not carry an error")

	var initResult lspInitializeResult
	require.NoError(t, json.Unmarshal(initResp.Result, &initResult))
	assert.Equal(t, 1, initResult.Capabilities.TextDocumentSync, "textDocumentSync capability must be 1 (full)")
	assert.True(t, initResult.Capabilities.HoverProvider, "hoverProvider must be true")
	assert.True(t, initResult.Capabilities.DefinitionProvider, "definitionProvider must be true")
	assert.True(t, initResult.Capabilities.DocumentSymbolProvider, "documentSymbolProvider must be true")
	out.Reset()

	// -------------------------------------------------------------------------
	// Step 2: Initialized notification — verify no response is written.
	// -------------------------------------------------------------------------
	initializedMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  "initialized",
		Params:  json.RawMessage(`{}`),
	}
	require.NoError(t, h.handle(initializedMsg), "initialized notification must not error")
	assert.Equal(t, 0, out.Len(), "initialized notification must not produce any output")

	// -------------------------------------------------------------------------
	// Step 3: didOpen — verify publishDiagnostics notification with arity error.
	// -------------------------------------------------------------------------
	didOpenParams := lspDidOpenParams{
		TextDocument: lspTextDocumentItem{
			URI:        uri,
			LanguageID: "perl",
			Version:    1,
			Text:       source,
		},
	}
	rawOpen, err := json.Marshal(didOpenParams)
	require.NoError(t, err)

	didOpenMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  "textDocument/didOpen",
		Params:  rawOpen,
	}
	require.NoError(t, h.handle(didOpenMsg), "didOpen must not error")

	diagNotification := readResponse(t, &out)
	require.NotNil(t, diagNotification)
	assert.Equal(t, "textDocument/publishDiagnostics", diagNotification.Method, "didOpen must produce publishDiagnostics")

	var diagParams lspPublishDiagnosticsParams
	require.NoError(t, json.Unmarshal(diagNotification.Params, &diagParams))
	assert.Equal(t, uri, diagParams.URI)
	assert.NotEmpty(t, diagParams.Diagnostics, "push() with no args should produce at least one diagnostic")
	out.Reset()

	// -------------------------------------------------------------------------
	// Step 4: Hover at $x (line 0, character 3) — verify markdown type info.
	// -------------------------------------------------------------------------
	rawID4 := json.RawMessage(`4`)
	hoverParams := lspTextDocumentPositionParams{
		TextDocument: lspTextDocumentIdentifier{URI: uri},
		Position:     lspPosition{Line: 0, Character: 3},
	}
	rawHoverParams, err := json.Marshal(hoverParams)
	require.NoError(t, err)

	hoverMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &rawID4,
		Method:  "textDocument/hover",
		Params:  rawHoverParams,
	}
	require.NoError(t, h.handle(hoverMsg), "hover must not error")

	hoverResp := readResponse(t, &out)
	require.NotNil(t, hoverResp)
	require.Nil(t, hoverResp.Error, "hover response must not carry an error")
	require.NotEqual(t, json.RawMessage("null"), hoverResp.Result, "hover at $x must return type info, not null")

	var hover lspHover
	require.NoError(t, json.Unmarshal(hoverResp.Result, &hover))
	assert.Equal(t, "markdown", hover.Contents.Kind, "hover content must be markdown")
	assert.Contains(t, hover.Contents.Value, "Type", "hover content must mention the type")
	out.Reset()

	// -------------------------------------------------------------------------
	// Step 5: Definition at $x (line 0, character 3) — verify Location response.
	// -------------------------------------------------------------------------
	rawID5 := json.RawMessage(`5`)
	defParams := lspTextDocumentPositionParams{
		TextDocument: lspTextDocumentIdentifier{URI: uri},
		Position:     lspPosition{Line: 0, Character: 3},
	}
	rawDefParams, err := json.Marshal(defParams)
	require.NoError(t, err)

	defMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &rawID5,
		Method:  "textDocument/definition",
		Params:  rawDefParams,
	}
	require.NoError(t, h.handle(defMsg), "definition must not error")

	defResp := readResponse(t, &out)
	require.NotNil(t, defResp)
	require.Nil(t, defResp.Error, "definition response must not carry an error")
	require.NotEqual(t, json.RawMessage("null"), defResp.Result, "definition at $x must return a Location, not null")

	var loc lspLocation
	require.NoError(t, json.Unmarshal(defResp.Result, &loc))
	assert.Equal(t, uri, loc.URI, "definition location must reference the same document")
	out.Reset()

	// -------------------------------------------------------------------------
	// Step 6: DocumentSymbol — verify $x appears in the symbol list.
	// -------------------------------------------------------------------------
	rawID6 := json.RawMessage(`6`)
	symParams := lspDocumentSymbolParams{
		TextDocument: lspTextDocumentIdentifier{URI: uri},
	}
	rawSymParams, err := json.Marshal(symParams)
	require.NoError(t, err)

	symMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &rawID6,
		Method:  "textDocument/documentSymbol",
		Params:  rawSymParams,
	}
	require.NoError(t, h.handle(symMsg), "documentSymbol must not error")

	symResp := readResponse(t, &out)
	require.NotNil(t, symResp)
	require.Nil(t, symResp.Error, "documentSymbol response must not carry an error")

	var syms []lspDocumentSymbol
	require.NoError(t, json.Unmarshal(symResp.Result, &syms))
	require.NotEmpty(t, syms, "documentSymbol must return at least one symbol")

	names := make(map[string]bool, len(syms))
	for _, s := range syms {
		names[s.Name] = true
	}
	assert.True(t, names["$x"], "symbol list must include $x")
	out.Reset()

	// -------------------------------------------------------------------------
	// Step 7: Unknown method — verify MethodNotFound error response.
	// -------------------------------------------------------------------------
	rawID7 := json.RawMessage(`7`)
	unknownMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &rawID7,
		Method:  "textDocument/unknown",
	}
	require.NoError(t, h.handle(unknownMsg), "unknown method must not return a Go error")

	unknownResp := readResponse(t, &out)
	require.NotNil(t, unknownResp)
	require.NotNil(t, unknownResp.Error, "unknown method must produce a JSON-RPC error response")
	assert.Equal(t, codeMethodNotFound, unknownResp.Error.Code, "error code must be MethodNotFound")
	out.Reset()

	// -------------------------------------------------------------------------
	// Step 8: Shutdown — verify null result, no error.
	// -------------------------------------------------------------------------
	rawID8 := json.RawMessage(`8`)
	shutdownMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &rawID8,
		Method:  "shutdown",
	}
	require.NoError(t, h.handle(shutdownMsg), "shutdown must not error")

	shutdownResp := readResponse(t, &out)
	require.NotNil(t, shutdownResp)
	require.Nil(t, shutdownResp.Error, "shutdown response must not carry an error")
	assert.Equal(t, json.RawMessage("null"), shutdownResp.Result, "shutdown result must be JSON null")
	out.Reset()

	// -------------------------------------------------------------------------
	// Step 9: Exit notification — verify exitFn called with code 0.
	// -------------------------------------------------------------------------
	var capturedCode int
	capturedCall := false
	h.exitFn = func(code int) {
		capturedCode = code
		capturedCall = true
	}

	exitMsg := &jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  "exit",
	}
	require.NoError(t, h.handle(exitMsg), "exit must not return a Go error")
	assert.True(t, capturedCall, "exitFn must be called on exit notification")
	assert.Equal(t, 0, capturedCode, "exit after shutdown must call exitFn with code 0")
}
