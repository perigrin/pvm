# PSC LSP Server Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the JSON-RPC protocol layer connecting editors to PSC's existing type inference engine.

**Architecture:** Four new files in `internal/psc/` (transport, protocol types, handler, command replacement) plus one prerequisite method on `SymbolTable`. The handler bridges LSP requests to existing `LSPServer` methods. Transport reads/writes Content-Length-framed JSON-RPC 2.0 on stdin/stdout.

**Tech Stack:** Go stdlib (`encoding/json`, `bufio`, `io`, `sync`), existing `internal/psc`, `internal/infer`, `internal/types`, `internal/parser` packages.

**Spec:** `docs/superpowers/specs/2026-03-31-psc-lsp-server-design.md`

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `internal/infer/symbols.go` | Modify | Add `children` to Scope, add `AllSymbols()` |
| `internal/infer/symbols_test.go` | Create or modify | Test `AllSymbols()` |
| `internal/psc/lsp_protocol.go` | Create | LSP type definitions |
| `internal/psc/lsp_transport.go` | Create | JSON-RPC reader/writer |
| `internal/psc/lsp_transport_test.go` | Create | Transport tests (package psc) |
| `internal/psc/lsp_handler.go` | Create | Handler dispatch |
| `internal/psc/lsp_handler_test.go` | Create | Handler tests (package psc) |
| `internal/psc/lsp_command.go` | Modify | Replace stub with main loop |
| `internal/psc/lsp_test.go` | Modify | Remove `TestLSPCommandReturnsNotImplemented` |
| `internal/psc/lsp_integration_test.go` | Create | Full lifecycle test (package psc) |

**Test package note:** New test files in `internal/psc/` use `package psc` (not `package psc_test`) because they need access to unexported types (`transport`, `handler`, `jsonRPCMessage`, etc.).

---

### Task 1: Add AllSymbols() to SymbolTable

**Files:**
- Modify: `internal/infer/symbols.go` (add `children` field to Scope, populate in EnterScope, add AllSymbols)
- Test: `internal/infer/symbols_test.go`

**Design decision:** The `Scope` struct has no `children` field — child scopes are orphaned after `ExitScope()`. Add `children []*Scope` to `Scope` and populate it in `EnterScope`. This is a two-line structural change that enables recursive scope walking.

- [ ] **Step 1: Write the failing test**

In `internal/infer/symbols_test.go` (use `package infer`):

```go
func TestAllSymbols(t *testing.T) {
	st := NewSymbolTable()
	st.Define(Symbol{Name: "$x", Kind: SymVariable, Type: types.Scalar, StartByte: 0, EndByte: 5})
	st.EnterScope("inner")
	st.Define(Symbol{Name: "$y", Kind: SymVariable, Type: types.Int, StartByte: 10, EndByte: 15})
	st.Define(Symbol{Name: "foo", Kind: SymSubroutine, Type: types.Code, StartByte: 20, EndByte: 50})
	st.ExitScope()

	syms := st.AllSymbols()

	assert.Len(t, syms, 3)
	names := make(map[string]bool)
	for _, s := range syms {
		names[s.Name] = true
	}
	assert.True(t, names["$x"])
	assert.True(t, names["$y"])
	assert.True(t, names["foo"])
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/infer/ -run TestAllSymbols -v`
Expected: FAIL — `AllSymbols` undefined

- [ ] **Step 3: Add children field to Scope and populate in EnterScope**

In `internal/infer/symbols.go`, add `children` to the Scope struct:

```go
type Scope struct {
	Name     string
	parent   *Scope
	children []*Scope
	symbols  map[string]Symbol
}
```

Update `EnterScope` to record the child:

```go
func (st *SymbolTable) EnterScope(name string) {
	child := &Scope{
		Name:    name,
		parent:  st.current,
		symbols: make(map[string]Symbol),
	}
	st.current.children = append(st.current.children, child)
	st.current = child
}
```

Add `AllSymbols`:

```go
// AllSymbols returns every symbol defined across all scopes in the table.
func (st *SymbolTable) AllSymbols() []Symbol {
	var result []Symbol
	var walk func(s *Scope)
	walk = func(s *Scope) {
		for _, sym := range s.symbols {
			result = append(result, sym)
		}
		for _, child := range s.children {
			walk(child)
		}
	}
	walk(st.root)
	return result
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/infer/ -run TestAllSymbols -v`
Expected: PASS

- [ ] **Step 5: Run full infer package tests for regressions**

Run: `go test ./internal/infer/ -count=1`
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add internal/infer/symbols.go internal/infer/symbols_test.go
git commit -m "feat(infer): add AllSymbols() to SymbolTable for LSP documentSymbol"
```

---

### Task 2: LSP Protocol Types

**Files:**
- Create: `internal/psc/lsp_protocol.go`

- [ ] **Step 1: Create the protocol types file**

Write `internal/psc/lsp_protocol.go` with all LSP type definitions. Use `package psc`. Include:

- JSON-RPC error code constants (`codeServerNotInitialized = -32002`, `codeMethodNotFound = -32601`)
- Initialize types: `lspInitializeParams`, `lspInitializeResult`, `lspServerCapabilities`
- Document sync types: `lspTextDocumentItem`, `lspDidOpenParams`, `lspDidChangeParams`, `lspDidCloseParams`, `lspTextDocumentContentChangeEvent`, `lspVersionedTextDocumentIdentifier`
- Position types: `lspPosition`, `lspRange`, `lspTextDocumentIdentifier`, `lspTextDocumentPositionParams`
- Diagnostic types: `lspDiagnostic`, `lspPublishDiagnosticsParams`
- Hover types: `lspMarkupContent`, `lspHover`
- Definition types: `lspLocation`
- Document symbol types: `lspDocumentSymbolParams`, `lspDocumentSymbol`
- SymbolKind constants: `symbolKindModule = 2`, `symbolKindFunction = 12`, `symbolKindVariable = 13`, `symbolKindMethod = 6`

All structs have `json` tags matching LSP spec field names. ABOUTME header required.

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/psc/`
Expected: success

- [ ] **Step 3: Commit**

```bash
git add internal/psc/lsp_protocol.go
git commit -m "feat(psc): add LSP protocol type definitions"
```

---

### Task 3: JSON-RPC Transport

**Files:**
- Create: `internal/psc/lsp_transport.go`
- Create: `internal/psc/lsp_transport_test.go` (use `package psc`)

- [ ] **Step 1: Write failing test for ReadMessage**

```go
package psc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readResponse is a test helper that reads one Content-Length-framed
// JSON-RPC message from buf.
func readResponse(t *testing.T, buf *bytes.Buffer) *jsonRPCMessage {
	t.Helper()
	tr := newTransport(buf, io.Discard)
	msg, err := tr.readMessage()
	require.NoError(t, err)
	return msg
}

func TestTransportReadMessage(t *testing.T) {
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	input := bytes.NewBufferString(header + body)
	output := &bytes.Buffer{}

	tr := newTransport(input, output)
	msg, err := tr.readMessage()

	require.NoError(t, err)
	assert.Equal(t, "2.0", msg.JSONRPC)
	assert.Equal(t, "initialize", msg.Method)
	assert.NotNil(t, msg.ID)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/psc/ -run TestTransportReadMessage -v`
Expected: FAIL — `newTransport` undefined

- [ ] **Step 3: Write transport implementation**

Create `internal/psc/lsp_transport.go` with ABOUTME header. Include:

- `jsonRPCMessage` struct
- `jsonRPCError` struct
- `transport` struct with `reader *bufio.Reader`, `writer io.Writer`, `mu sync.Mutex`
- `newTransport(r io.Reader, w io.Writer) *transport`
- `readMessage() (*jsonRPCMessage, error)` — reads Content-Length headers, then `io.ReadFull` for the body
- `sendResponse(id *json.RawMessage, result interface{})`
- `sendError(id *json.RawMessage, code int, message string)`
- `sendNotification(method string, params interface{})`
- `send(msg *jsonRPCMessage)` — internal, writes Content-Length header + JSON body
- `mustMarshal(v interface{}) json.RawMessage` — helper

Use `io.ReadFull` (not buffered read) for the body to handle messages larger than `bufio.Reader`'s default buffer.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/psc/ -run TestTransportReadMessage -v`
Expected: PASS

- [ ] **Step 5: Add write framing test**

```go
func TestTransportWriteFraming(t *testing.T) {
	output := &bytes.Buffer{}
	tr := newTransport(strings.NewReader(""), output)

	id := json.RawMessage(`1`)
	tr.sendResponse(&id, map[string]string{"key": "value"})

	raw := output.String()
	assert.True(t, strings.HasPrefix(raw, "Content-Length: "))
	assert.Contains(t, raw, "\r\n\r\n")

	// Parse back through transport to verify round-trip
	resp := readResponse(t, output)
	// output was consumed by sendResponse write, re-parse from raw
	reparseBuf := bytes.NewBufferString(raw)
	resp = readResponse(t, reparseBuf)
	assert.NotNil(t, resp.Result)
}
```

- [ ] **Step 6: Add malformed header test**

```go
func TestTransportMalformedHeader(t *testing.T) {
	input := bytes.NewBufferString("Bad-Header: foo\r\n\r\n")
	tr := newTransport(input, io.Discard)
	_, err := tr.readMessage()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Content-Length")
}
```

- [ ] **Step 7: Add large body test**

```go
func TestTransportLargeBody(t *testing.T) {
	// Body larger than bufio.Reader's 4096 default buffer
	bigParams := strings.Repeat("x", 8192)
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"test","params":{"data":"%s"}}`, bigParams)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	input := bytes.NewBufferString(header + body)

	tr := newTransport(input, io.Discard)
	msg, err := tr.readMessage()

	require.NoError(t, err)
	assert.Equal(t, "test", msg.Method)
}
```

- [ ] **Step 8: Run all transport tests**

Run: `go test ./internal/psc/ -run TestTransport -v`
Expected: all PASS

- [ ] **Step 9: Commit**

```bash
git add internal/psc/lsp_transport.go internal/psc/lsp_transport_test.go
git commit -m "feat(psc): add JSON-RPC 2.0 transport for LSP server"
```

---

### Task 4: Handler — Lifecycle + Structure

**Files:**
- Create: `internal/psc/lsp_handler.go`
- Create: `internal/psc/lsp_handler_test.go` (use `package psc`)

The handler struct includes a `sources map[string][]byte` field from the start (needed by later tasks for position conversion).

- [ ] **Step 1: Write failing test for initialize lifecycle**

```go
package psc

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerInitializeShutdown(t *testing.T) {
	out := &bytes.Buffer{}
	tr := newTransport(bytes.NewReader(nil), out)
	h := newHandler(NewLSPServer(), tr)

	// Initialize
	initID := json.RawMessage(`1`)
	initParams, _ := json.Marshal(lspInitializeParams{})
	err := h.handle(&jsonRPCMessage{JSONRPC: "2.0", ID: &initID, Method: "initialize", Params: initParams})
	require.NoError(t, err)

	resp := readResponse(t, out)
	assert.NotNil(t, resp.Result)
	assert.Nil(t, resp.Error)

	// Shutdown
	shutdownID := json.RawMessage(`2`)
	err = h.handle(&jsonRPCMessage{JSONRPC: "2.0", ID: &shutdownID, Method: "shutdown"})
	require.NoError(t, err)
	assert.True(t, h.isShutdown)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/psc/ -run TestHandlerInitializeShutdown -v`
Expected: FAIL — `newHandler` undefined

- [ ] **Step 3: Write handler implementation**

Create `internal/psc/lsp_handler.go` with ABOUTME header. Include:

```go
type handler struct {
	server      *LSPServer
	transport   *transport
	sources     map[string][]byte // document source for position conversion
	initialized bool
	isShutdown  bool
	exitFn      func(int)
}

func newHandler(server *LSPServer, tr *transport) *handler {
	return &handler{
		server:    server,
		transport: tr,
		sources:   make(map[string][]byte),
		exitFn:    func(code int) {},
	}
}
```

Implement `handle()` dispatch, `handleInitialize`, `handleShutdown`, and exit handling. Stub remaining handlers as `return nil`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/psc/ -run TestHandlerInitializeShutdown -v`
Expected: PASS

- [ ] **Step 5: Add test for unknown method**

```go
func TestHandlerUnknownMethod(t *testing.T) {
	out := &bytes.Buffer{}
	tr := newTransport(bytes.NewReader(nil), out)
	h := newHandler(NewLSPServer(), tr)
	h.initialized = true

	id := json.RawMessage(`1`)
	_ = h.handle(&jsonRPCMessage{JSONRPC: "2.0", ID: &id, Method: "textDocument/unknown"})

	resp := readResponse(t, out)
	require.NotNil(t, resp.Error)
	assert.Equal(t, codeMethodNotFound, resp.Error.Code)
}
```

- [ ] **Step 6: Add test for request before initialize**

```go
func TestHandlerRequestBeforeInitialize(t *testing.T) {
	out := &bytes.Buffer{}
	tr := newTransport(bytes.NewReader(nil), out)
	h := newHandler(NewLSPServer(), tr)
	// h.initialized is false

	id := json.RawMessage(`1`)
	_ = h.handle(&jsonRPCMessage{JSONRPC: "2.0", ID: &id, Method: "textDocument/hover"})

	resp := readResponse(t, out)
	require.NotNil(t, resp.Error)
	assert.Equal(t, codeServerNotInitialized, resp.Error.Code)
}
```

- [ ] **Step 7: Add test for exit without shutdown**

```go
func TestHandlerExitWithoutShutdown(t *testing.T) {
	tr := newTransport(bytes.NewReader(nil), io.Discard)
	h := newHandler(NewLSPServer(), tr)
	h.initialized = true

	var exitCode int
	h.exitFn = func(code int) { exitCode = code }

	_ = h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "exit"})
	assert.Equal(t, 1, exitCode)
}

func TestHandlerExitAfterShutdown(t *testing.T) {
	tr := newTransport(bytes.NewReader(nil), io.Discard)
	h := newHandler(NewLSPServer(), tr)
	h.initialized = true
	h.isShutdown = true

	var exitCode int
	h.exitFn = func(code int) { exitCode = code }

	_ = h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "exit"})
	assert.Equal(t, 0, exitCode)
}
```

- [ ] **Step 8: Run all handler tests**

Run: `go test ./internal/psc/ -run TestHandler -v`
Expected: all PASS

- [ ] **Step 9: Commit**

```bash
git add internal/psc/lsp_handler.go internal/psc/lsp_handler_test.go
git commit -m "feat(psc): add LSP handler with lifecycle dispatch"
```

---

### Task 5: Handler — Document Sync + Publish Diagnostics

**Files:**
- Modify: `internal/psc/lsp_handler.go`
- Modify: `internal/psc/lsp_handler_test.go`

- [ ] **Step 1: Write failing test for didOpen publishing diagnostics**

```go
func TestHandlerDidOpenPublishesDiagnostics(t *testing.T) {
	out := &bytes.Buffer{}
	tr := newTransport(bytes.NewReader(nil), out)
	h := newHandler(NewLSPServer(), tr)
	h.initialized = true

	params, _ := json.Marshal(lspDidOpenParams{
		TextDocument: lspTextDocumentItem{
			URI:  "file:///test.pl",
			Text: "push();\n",
		},
	})
	err := h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didOpen", Params: params})
	require.NoError(t, err)

	resp := readResponse(t, out)
	assert.Equal(t, "textDocument/publishDiagnostics", resp.Method)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/psc/ -run TestHandlerDidOpen -v`
Expected: FAIL

- [ ] **Step 3: Implement position conversion helpers**

Add `positionToOffset` and `offsetToPosition` to `lsp_handler.go`.

- [ ] **Step 4: Implement handleDidOpen, handleDidChange, handleDidClose, publishDiagnostics**

`handleDidOpen`: parse params, store source in `h.sources[uri]`, call `OpenDocument`, call `publishDiagnostics`. On `OpenDocument` error, publish a single error diagnostic spanning the full document.

`handleDidChange`: get full text from first content change, store in sources, call `OpenDocument` (re-open), publish diagnostics.

`handleDidClose`: call `CloseDocument`, delete from `h.sources`.

`publishDiagnostics`: convert each `infer.Diagnostic` to `lspDiagnostic` with severity mapping and position conversion. Append suggestion text to message if non-empty.

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/psc/ -run TestHandlerDidOpen -v`
Expected: PASS

- [ ] **Step 6: Add position conversion round-trip test**

```go
func TestPositionConversion(t *testing.T) {
	source := []byte("my $x = 42;\nprint $x;\n")

	// Line 0, col 3 -> byte 3 ($x)
	offset := positionToOffset(source, 0, 3)
	assert.Equal(t, uint32(3), offset)

	pos := offsetToPosition(source, 3)
	assert.Equal(t, 0, pos.Line)
	assert.Equal(t, 3, pos.Character)

	// Line 1, col 0 -> byte 12 (start of "print")
	offset = positionToOffset(source, 1, 0)
	assert.Equal(t, uint32(12), offset)

	pos = offsetToPosition(source, 12)
	assert.Equal(t, 1, pos.Line)
	assert.Equal(t, 0, pos.Character)
}
```

- [ ] **Step 7: Add didChange test**

```go
func TestHandlerDidChange(t *testing.T) {
	out := &bytes.Buffer{}
	tr := newTransport(bytes.NewReader(nil), out)
	h := newHandler(NewLSPServer(), tr)
	h.initialized = true

	// Open
	openParams, _ := json.Marshal(lspDidOpenParams{
		TextDocument: lspTextDocumentItem{URI: "file:///test.pl", Text: "my $x = 1;\n"},
	})
	_ = h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didOpen", Params: openParams})
	out.Reset()

	// Change
	changeParams, _ := json.Marshal(lspDidChangeParams{
		TextDocument:   lspVersionedTextDocumentIdentifier{URI: "file:///test.pl", Version: 2},
		ContentChanges: []lspTextDocumentContentChangeEvent{{Text: "push();\n"}},
	})
	err := h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didChange", Params: changeParams})
	require.NoError(t, err)

	resp := readResponse(t, out)
	assert.Equal(t, "textDocument/publishDiagnostics", resp.Method)
}
```

- [ ] **Step 8: Add didClose test**

```go
func TestHandlerDidClose(t *testing.T) {
	out := &bytes.Buffer{}
	tr := newTransport(bytes.NewReader(nil), out)
	h := newHandler(NewLSPServer(), tr)
	h.initialized = true

	// Open then close
	openParams, _ := json.Marshal(lspDidOpenParams{
		TextDocument: lspTextDocumentItem{URI: "file:///test.pl", Text: "my $x = 1;\n"},
	})
	_ = h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didOpen", Params: openParams})

	closeParams, _ := json.Marshal(lspDidCloseParams{
		TextDocument: lspTextDocumentIdentifier{URI: "file:///test.pl"},
	})
	err := h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didClose", Params: closeParams})
	require.NoError(t, err)

	// Verify document is gone
	assert.Nil(t, h.server.Diagnostics("file:///test.pl"))
}
```

- [ ] **Step 9: Run all handler tests**

Run: `go test ./internal/psc/ -run TestHandler -v`
Expected: all PASS

- [ ] **Step 10: Commit**

```bash
git add internal/psc/lsp_handler.go internal/psc/lsp_handler_test.go
git commit -m "feat(psc): add document sync and diagnostic publishing to LSP handler"
```

---

### Task 6: Handler — Hover

**Files:**
- Modify: `internal/psc/lsp_handler.go`
- Modify: `internal/psc/lsp_handler_test.go`

- [ ] **Step 1: Write failing test for hover with type info**

```go
func TestHandlerHover(t *testing.T) {
	out := &bytes.Buffer{}
	tr := newTransport(bytes.NewReader(nil), out)
	h := newHandler(NewLSPServer(), tr)
	h.initialized = true

	openParams, _ := json.Marshal(lspDidOpenParams{
		TextDocument: lspTextDocumentItem{URI: "file:///test.pl", Text: "my $x = 42;\n"},
	})
	_ = h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didOpen", Params: openParams})
	out.Reset()

	hoverParams, _ := json.Marshal(lspTextDocumentPositionParams{
		TextDocument: lspTextDocumentIdentifier{URI: "file:///test.pl"},
		Position:     lspPosition{Line: 0, Character: 3},
	})
	hoverID := json.RawMessage(`2`)
	_ = h.handle(&jsonRPCMessage{JSONRPC: "2.0", ID: &hoverID, Method: "textDocument/hover", Params: hoverParams})

	resp := readResponse(t, out)
	assert.NotNil(t, resp.Result)
	// Should contain type information
	var hover lspHover
	_ = json.Unmarshal(resp.Result, &hover)
	assert.Equal(t, "markdown", hover.Contents.Kind)
	assert.Contains(t, hover.Contents.Value, "Type")
}
```

- [ ] **Step 2: Run test to verify it fails**

- [ ] **Step 3: Implement handleHover**

Reads source from `h.sources[uri]`, converts position to offset, calls `TypeAtByte`. Returns `null` when type is Unknown or not found. Otherwise returns `lspHover` with Markdown content showing the type name via `typ.String()`.

- [ ] **Step 4: Run test to verify it passes**

- [ ] **Step 5: Add test for hover on whitespace returning null**

```go
func TestHandlerHoverNoType(t *testing.T) {
	out := &bytes.Buffer{}
	tr := newTransport(bytes.NewReader(nil), out)
	h := newHandler(NewLSPServer(), tr)
	h.initialized = true

	openParams, _ := json.Marshal(lspDidOpenParams{
		TextDocument: lspTextDocumentItem{URI: "file:///test.pl", Text: "my $x = 42;\n"},
	})
	_ = h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didOpen", Params: openParams})
	out.Reset()

	// Hover on whitespace (after semicolon)
	hoverParams, _ := json.Marshal(lspTextDocumentPositionParams{
		TextDocument: lspTextDocumentIdentifier{URI: "file:///test.pl"},
		Position:     lspPosition{Line: 0, Character: 11},
	})
	hoverID := json.RawMessage(`3`)
	_ = h.handle(&jsonRPCMessage{JSONRPC: "2.0", ID: &hoverID, Method: "textDocument/hover", Params: hoverParams})

	resp := readResponse(t, out)
	// Result should be JSON null
	assert.Equal(t, "null", string(resp.Result))
}
```

- [ ] **Step 6: Run all hover tests**

- [ ] **Step 7: Commit**

```bash
git add internal/psc/lsp_handler.go internal/psc/lsp_handler_test.go
git commit -m "feat(psc): add hover handler to LSP server"
```

---

### Task 7: Handler — Definition

**Files:**
- Modify: `internal/psc/lsp_handler.go`
- Modify: `internal/psc/lsp_handler_test.go`

- [ ] **Step 1: Write failing test for definition**

```go
func TestHandlerDefinition(t *testing.T) {
	out := &bytes.Buffer{}
	tr := newTransport(bytes.NewReader(nil), out)
	h := newHandler(NewLSPServer(), tr)
	h.initialized = true

	source := "my $x = 42;\nprint $x;\n"
	openParams, _ := json.Marshal(lspDidOpenParams{
		TextDocument: lspTextDocumentItem{URI: "file:///test.pl", Text: source},
	})
	_ = h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didOpen", Params: openParams})
	out.Reset()

	// Request definition at $x on line 1 (the usage site)
	defParams, _ := json.Marshal(lspTextDocumentPositionParams{
		TextDocument: lspTextDocumentIdentifier{URI: "file:///test.pl"},
		Position:     lspPosition{Line: 1, Character: 6}, // $x in "print $x"
	})
	defID := json.RawMessage(`2`)
	_ = h.handle(&jsonRPCMessage{JSONRPC: "2.0", ID: &defID, Method: "textDocument/definition", Params: defParams})

	resp := readResponse(t, out)
	assert.NotNil(t, resp.Result)
	// Should point back to the declaration on line 0
	var loc lspLocation
	_ = json.Unmarshal(resp.Result, &loc)
	assert.Equal(t, "file:///test.pl", loc.URI)
	assert.Equal(t, 0, loc.Range.Start.Line)
}
```

- [ ] **Step 2: Run test to verify it fails**

- [ ] **Step 3: Implement handleDefinition**

Reads source from `h.sources[uri]`, converts position to offset, calls `DefinitionAtByte`. Returns `null` when symbol not found. Otherwise returns `lspLocation` with URI set to the request document URI (single-file only) and range from the symbol's StartByte/EndByte.

- [ ] **Step 4: Run test to verify it passes**

- [ ] **Step 5: Commit**

```bash
git add internal/psc/lsp_handler.go internal/psc/lsp_handler_test.go
git commit -m "feat(psc): add definition handler to LSP server"
```

---

### Task 8: Handler — Document Symbols

**Files:**
- Modify: `internal/psc/lsp_handler.go`
- Modify: `internal/psc/lsp_handler_test.go`

Depends on Task 1 (`AllSymbols()`).

- [ ] **Step 1: Write failing test for documentSymbol**

```go
func TestHandlerDocumentSymbol(t *testing.T) {
	out := &bytes.Buffer{}
	tr := newTransport(bytes.NewReader(nil), out)
	h := newHandler(NewLSPServer(), tr)
	h.initialized = true

	source := "my $x = 1;\nsub foo { }\n"
	openParams, _ := json.Marshal(lspDidOpenParams{
		TextDocument: lspTextDocumentItem{URI: "file:///test.pl", Text: source},
	})
	_ = h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didOpen", Params: openParams})
	out.Reset()

	symParams, _ := json.Marshal(lspDocumentSymbolParams{
		TextDocument: lspTextDocumentIdentifier{URI: "file:///test.pl"},
	})
	symID := json.RawMessage(`2`)
	_ = h.handle(&jsonRPCMessage{JSONRPC: "2.0", ID: &symID, Method: "textDocument/documentSymbol", Params: symParams})

	resp := readResponse(t, out)
	assert.NotNil(t, resp.Result)

	var symbols []lspDocumentSymbol
	_ = json.Unmarshal(resp.Result, &symbols)
	assert.GreaterOrEqual(t, len(symbols), 2, "should have at least $x and foo")

	names := make(map[string]bool)
	for _, s := range symbols {
		names[s.Name] = true
	}
	assert.True(t, names["$x"], "should contain $x")
	assert.True(t, names["foo"], "should contain foo")
}
```

- [ ] **Step 2: Run test to verify it fails**

- [ ] **Step 3: Implement handleDocumentSymbol**

Calls `SymbolTable(uri)`, then `st.AllSymbols()`. Maps each symbol to `lspDocumentSymbol` with `symbolKindForInfer()` helper. Sets `selectionRange == range` (Phase 1 limitation).

- [ ] **Step 4: Run test to verify it passes**

- [ ] **Step 5: Commit**

```bash
git add internal/psc/lsp_handler.go internal/psc/lsp_handler_test.go
git commit -m "feat(psc): add documentSymbol handler to LSP server"
```

---

### Task 9: Command Entry Point

**Files:**
- Modify: `internal/psc/lsp_command.go`
- Modify: `internal/psc/lsp_test.go` (remove `TestLSPCommandReturnsNotImplemented`)

- [ ] **Step 1: Replace the stub in lsp_command.go**

Update ABOUTME to remove "stub" / "not yet implemented" language. Replace `RunE` with the main loop: create LSPServer, transport (stdin/stdout), handler with `os.Exit` as exitFn, loop on `readMessage` → `handle`.

- [ ] **Step 2: Update lsp_test.go**

Remove `TestLSPCommandReturnsNotImplemented` (lines 66-77). The command-exists test (`TestLSPCommandExists`) still covers discoverability. The working server is tested via the integration test.

- [ ] **Step 3: Verify it compiles**

Run: `go build ./internal/psc/`
Expected: success

- [ ] **Step 4: Verify all existing tests still pass**

Run: `go test ./internal/psc/ -count=1`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/psc/lsp_command.go internal/psc/lsp_test.go
git commit -m "feat(psc): replace LSP command stub with working server entry point"
```

---

### Task 10: Integration Test

**Files:**
- Create: `internal/psc/lsp_integration_test.go` (use `package psc`)

- [ ] **Step 1: Write integration test**

Full lifecycle using `io.Pipe`: create connected reader/writer pairs, wrap in transport, create handler.

Sequence: initialize → didOpen (with `"my $x = 42;\npush();\n"`) → hover (on $x) → definition (on $x) → documentSymbol → shutdown → exit.

Verify:
- Initialize response has capabilities
- didOpen triggers publishDiagnostics with at least one diagnostic (push arity)
- Hover returns type info
- Definition returns a location
- documentSymbol returns symbols
- Shutdown acknowledges
- Exit calls exitFn with 0

Also test: unknown method returns MethodNotFound.

- [ ] **Step 2: Run integration test**

Run: `go test ./internal/psc/ -run TestLSPIntegration -v`
Expected: PASS

- [ ] **Step 3: Run full test suite**

Run: `make test`
Expected: all packages PASS

- [ ] **Step 4: Commit**

```bash
git add internal/psc/lsp_integration_test.go
git commit -m "test(psc): add LSP integration test with full message lifecycle"
```

---

### Task 11: Update Website Documentation

**Files:**
- Modify: `guides/editor-setup.html` (in gh-pages worktree)
- Modify: `reference/psc.html` (in gh-pages worktree)

- [ ] **Step 1: Verify gh-pages worktree exists**

Run: `ls /home/perigrin/dev/pvm/.claude/worktrees/gh-pages-cookbook/ 2>&1`

If not found, check `git worktree list` for the correct path.

- [ ] **Step 2: Update editor-setup.html**

Replace "LSP server (coming soon)" section with actual LSP configurations for each editor. Now that `psc lsp` works, editors can use it directly:

- Neovim: `lspconfig` entry pointing to `psc lsp`
- VS Code: direct LSP client config or extension
- Emacs: `lsp-mode` / `eglot` configuration
- Helix: direct `psc lsp` in `languages.toml`

Keep the `psc check` linter configs as a fallback option.

- [ ] **Step 3: Update psc.html reference**

Remove "(not yet implemented)" from the lsp command description.

- [ ] **Step 4: Commit and push**

```bash
git add guides/editor-setup.html reference/psc.html
git commit -m "docs: update editor setup with working LSP configurations"
git push origin gh-pages
```
