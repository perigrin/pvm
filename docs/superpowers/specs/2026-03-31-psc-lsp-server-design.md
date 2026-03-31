# PSC LSP Server

Date: 2026-03-31

## Problem

PSC can type-check Perl code and report diagnostics, but only as a batch
command (`psc check`). Editors that support the Language Server Protocol
can deliver diagnostics, hover tooltips, go-to-definition, and symbol
outlines in real time — but PSC's built-in LSP server is a stub that
returns "not yet implemented."

The inference engine, type system, parser, and LSPServer document
management are all complete. What is missing is the JSON-RPC protocol
layer that connects an editor to the existing LSPServer methods.

## Scope

### In scope (Phase 1)

- JSON-RPC 2.0 transport over stdin/stdout
- LSP lifecycle (initialize, initialized, shutdown, exit)
- Full document sync (textDocumentSyncKind = 1)
- textDocument/publishDiagnostics (on open and change)
- textDocument/hover (inferred type at cursor)
- textDocument/definition (go-to-definition for declarations)
- textDocument/documentSymbol (outline view)

### Out of scope

- Incremental document sync
- textDocument/completion
- textDocument/references
- textDocument/rename
- workspace/symbol
- Cross-file project indexing via LSP (single-file inference only)

## Architecture

Four new files, one modified:

```
internal/psc/
  lsp_transport.go    — JSON-RPC 2.0 reader/writer (stdin/stdout framing)
  lsp_protocol.go     — LSP type definitions (structs with json tags)
  lsp_handler.go      — Handler dispatch: bridges LSP requests to LSPServer
  lsp_command.go      — Modified: stub replaced with main loop
```

### Message Flow

```
Editor → stdin → lsp_transport (parse) → lsp_handler (dispatch)
       → LSPServer methods → lsp_handler (format) → lsp_transport (write)
       → stdout → Editor
```

### Server Capabilities

The initialize response advertises:

- `textDocumentSync`: Full (kind 1)
- `hoverProvider`: true
- `definitionProvider`: true
- `documentSymbolProvider`: true

## Transport Layer (`lsp_transport.go`)

Reads and writes Content-Length-framed JSON-RPC 2.0 messages.

### Message Type

```go
type jsonRPCMessage struct {
    JSONRPC string           `json:"jsonrpc"`
    ID      *json.RawMessage `json:"id,omitempty"`
    Method  string           `json:"method,omitempty"`
    Params  json.RawMessage  `json:"params,omitempty"`
    Result  json.RawMessage  `json:"result,omitempty"`
    Error   *jsonRPCError    `json:"error,omitempty"`
}
```

Messages with a non-nil `ID` are requests (need a response). Messages
with a nil `ID` are notifications.

### Reader

`bufio.Reader` on stdin. Reads headers line by line until `\r\n`,
extracts `Content-Length`, reads that many bytes as the JSON body.

### Writer

`sync.Mutex`-protected writes to stdout. Marshals response to JSON,
prepends `Content-Length: N\r\n\r\n`, writes atomically.

### Public API

- `ReadMessage() (*jsonRPCMessage, error)` — blocks until next message
- `SendResponse(id *json.RawMessage, result interface{})` — send request response
- `SendNotification(method string, params interface{})` — send server-initiated notification
- `SendError(id *json.RawMessage, code int, message string)` — send error response

## Handler Dispatch (`lsp_handler.go`)

A `Handler` struct holds the `LSPServer` and transport. A method dispatch
map routes incoming method names to handler functions.

### Lifecycle

| Method | Type | Action |
|--------|------|--------|
| `initialize` | request | Respond with server capabilities |
| `initialized` | notification | No-op |
| `shutdown` | request | Acknowledge, set shutdown flag |
| `exit` | notification | `os.Exit(0)` if shutdown called, `os.Exit(1)` if not |

### Document Sync

| Method | Type | Action |
|--------|------|--------|
| `textDocument/didOpen` | notification | `OpenDocument(uri, source)`, publish diagnostics |
| `textDocument/didChange` | notification | Re-open with full text (full sync), publish diagnostics |
| `textDocument/didClose` | notification | `CloseDocument(uri)` |

### Queries

| Method | Type | Action |
|--------|------|--------|
| `textDocument/hover` | request | Convert position to byte offset, call `TypeAtByte`, return type as Markdown |
| `textDocument/definition` | request | Convert position to byte offset, call `DefinitionAtByte`, return Location |
| `textDocument/documentSymbol` | request | Call `SymbolTable(uri)`, walk scopes, return DocumentSymbol array |

### Position Conversion

Two helper functions convert between LSP's 0-based line/character
positions and byte offsets:

- `positionToOffset(source []byte, line, col int) uint32`
- `offsetToPosition(source []byte, offset uint32) (line, col int)`

Both scan source bytes counting `\n` characters. These are the only
position-aware code in the handler layer.

### Publishing Diagnostics

After `didOpen` and `didChange`, the handler calls
`LSPServer.Diagnostics(uri)`, converts each `infer.Diagnostic` to LSP
format (severity mapping, byte offset to line/col, diagnostic code),
and sends a `textDocument/publishDiagnostics` notification.

Severity mapping:
- `infer.Error` → LSP DiagnosticSeverity 1 (Error)
- `infer.Warning` → LSP DiagnosticSeverity 2 (Warning)
- `infer.Info` → LSP DiagnosticSeverity 3 (Information)

Each diagnostic includes `source: "psc"` and `code` from the inference
diagnostic code field.

When a diagnostic has a non-empty `Suggestion` field (guard suggestions
like "Add guard: if (defined($x)) { ... }"), the suggestion text is
appended to the diagnostic message with a newline separator. This
matches the `psc check` CLI output format where hint lines follow the
main diagnostic.

## LSP Types (`lsp_protocol.go`)

Minimal set of structs matching the LSP specification. Only the types
needed for the four features.

### Initialize

- `InitializeParams` — rootUri, capabilities
- `InitializeResult` — capabilities struct with feature flags
- `ServerCapabilities` — textDocumentSync, hoverProvider, definitionProvider, documentSymbolProvider

### Document Sync

- `TextDocumentItem` — uri, languageId, version, text
- `DidOpenTextDocumentParams` — textDocument (TextDocumentItem)
- `DidChangeTextDocumentParams` — textDocument (VersionedTextDocumentIdentifier), contentChanges
- `TextDocumentContentChangeEvent` — text (full content, no range)
- `DidCloseTextDocumentParams` — textDocument

### Diagnostics

- `PublishDiagnosticsParams` — uri, diagnostics
- `Diagnostic` — range, severity, message, code, source

### Hover

- `HoverParams` — textDocument, position
- `Hover` — contents (MarkupContent), range

### Definition

- `DefinitionParams` — textDocument, position
- `Location` — uri, range

### Document Symbols

- `DocumentSymbolParams` — textDocument
- `DocumentSymbol` — name, kind, range, selectionRange, children

### Shared

- `Position` — line, character (0-based)
- `Range` — start, end
- `TextDocumentIdentifier` — uri
- `VersionedTextDocumentIdentifier` — uri, version
- `TextDocumentPositionParams` — textDocument, position
- `MarkupContent` — kind ("markdown"), value

## Command Entry Point (`lsp_command.go`)

Replace the stub with:

1. Create `LSPServer`
2. Create transport (stdin reader, stdout writer)
3. Create handler with server and transport
4. Enter read loop: `ReadMessage()` → dispatch → respond
5. Exit on `exit` notification

The loop runs until the client sends `shutdown` followed by `exit`.
Unknown methods receive a MethodNotFound (-32601) error response.

## Testing

### Transport tests (`lsp_transport_test.go`)

- Read a well-formed Content-Length message from bytes.Buffer
- Write a response and verify framing
- Handle malformed headers (missing Content-Length, bad number)

### Handler tests (`lsp_handler_test.go`)

- Initialize/shutdown lifecycle round-trip
- didOpen triggers publishDiagnostics notification
- Hover returns type info for a known position
- Definition returns location for a variable declaration
- documentSymbol returns symbols from a Perl file
- Position/offset conversion round-trips correctly

### Integration test (`lsp_integration_test.go`)

- Pipe a sequence of JSON-RPC messages through the handler using
  `io.Pipe` (initialize → didOpen → hover → shutdown → exit)
- Verify each response matches expectations
- Uses real parsing and inference on small Perl snippets

No mock mode. Tests use the real inference engine on small Perl code,
following the pattern established in `lsp_inference_test.go`.

## Existing Infrastructure Used

All of the following are complete and tested:

- `LSPServer.OpenDocument/CloseDocument` — document management
- `LSPServer.TypeAtByte` — type lookup at byte offset
- `LSPServer.DefinitionAtByte` — definition lookup at byte offset
- `LSPServer.SymbolTable` — symbol table for document
- `LSPServer.Diagnostics` — diagnostic list for document
- `infer.Analyze` — two-pass type inference engine
- `types.Type` — uint32 bitset type system
- `parser.Parser` — gotreesitter pure-Go Perl parser

## Prerequisites

### Add `AllSymbols()` to SymbolTable

The `SymbolTable` struct in `internal/infer/symbols.go` only exposes
point lookups (`Lookup(name)`). The `documentSymbol` handler needs to
enumerate all symbols across all scopes. Add:

```go
func (st *SymbolTable) AllSymbols() []Symbol
```

This walks the scope tree and returns every declared symbol. Implement
TDD in `internal/infer/` before any handler work begins.

## Error Handling

### Parse errors on `OpenDocument`

When `LSPServer.OpenDocument` returns an error (e.g., tree-sitter
rejects the source), the handler publishes a single Error-severity
diagnostic spanning the entire document with the parse error message.
The publish step always fires — it is never skipped.

### Requests before `initialize`

Return JSON-RPC error code -32002 (ServerNotInitialized).

### Unknown methods

Return JSON-RPC error code -32601 (MethodNotFound).

## Handler Details

### Hover null response

When `TypeAtByte` returns `(Unknown, false)` or `(Unknown, true)`,
respond with JSON `null`. This is standard LSP behavior for "no hover
at this position."

### Definition URI

The definition `uri` is always the same as the request document URI.
Phase 1 is single-file only.

### SymbolKind mapping

| `infer.SymbolKind` | LSP `SymbolKind` | Value |
|--------------------|------------------|-------|
| `SymVariable` | Variable | 13 |
| `SymSubroutine` | Function | 12 |
| `SymPackage` | Module | 2 |
| `SymMethod` | Method | 6 |

### selectionRange

Phase 1 sets `selectionRange` equal to `range`. The `Symbol` struct
does not store the name token's byte range separately.

### Main loop

The main loop is synchronous — one message at a time. The transport
writer mutex exists for correctness but is not exercised by concurrent
writes in Phase 1.

## Known Limitations

- Position conversion assumes ASCII source. Non-ASCII characters in
  string literals and comments may cause incorrect hover/definition
  positions. UTF-16 code unit conversion is deferred to a follow-up.
- `selectionRange` for document symbols equals `range` (full
  declaration span, not just the name token).
- Single-file inference only. Cross-file resolution via LSP is
  deferred.

## Implementation Sequence

Each step has a clear test-first target. Steps are ordered by
dependency.

1. **Prerequisite:** Add `AllSymbols()` to `SymbolTable` with unit test
2. **Transport:** `lsp_transport.go` — reader/writer, transport tests
3. **Protocol types:** `lsp_protocol.go` — struct definitions (no logic)
4. **Lifecycle handlers:** initialize, shutdown, exit in `lsp_handler.go`
5. **Document sync:** didOpen, didChange, didClose + publish diagnostics
6. **Hover:** textDocument/hover handler
7. **Definition:** textDocument/definition handler
8. **Document symbols:** textDocument/documentSymbol handler
9. **Command entry point:** Replace stub in `lsp_command.go`
10. **Integration test:** Full message sequence via io.Pipe

## Testing

### Transport tests (`lsp_transport_test.go`)

- Read a well-formed Content-Length message from bytes.Buffer
- Write a response and verify framing
- Handle malformed headers (missing Content-Length, bad number)
- Read a message whose body exceeds bufio.Reader default buffer size

### Handler tests (`lsp_handler_test.go`)

- Initialize/shutdown lifecycle round-trip
- Exit without prior shutdown returns exit code 1
- Request before initialize returns ServerNotInitialized
- Unknown method returns MethodNotFound
- didOpen triggers publishDiagnostics notification
- didOpen with parse error publishes error diagnostic
- Hover returns type info for a known position
- Hover on whitespace returns null
- Definition returns location for a variable declaration
- documentSymbol returns symbols from a Perl file
- Position/offset conversion round-trips correctly

### Integration test (`lsp_integration_test.go`)

- Pipe a sequence of JSON-RPC messages through the handler using
  `io.Pipe` (initialize → didOpen → hover → definition →
  documentSymbol → shutdown → exit)
- Unknown method returns MethodNotFound error
- Verify each response matches expectations
- Uses real parsing and inference on small Perl snippets

No mock mode. Tests use the real inference engine on small Perl code,
following the pattern established in `lsp_inference_test.go`.

## Existing Infrastructure Used

All of the following are complete and tested:

- `LSPServer.OpenDocument/CloseDocument` — document management
- `LSPServer.TypeAtByte` — type lookup at byte offset
- `LSPServer.DefinitionAtByte` — definition lookup at byte offset
- `LSPServer.SymbolTable` — symbol table for document
- `LSPServer.Diagnostics` — diagnostic list for document
- `infer.Analyze` — two-pass type inference engine
- `types.Type` — uint32 bitset type system
- `parser.Parser` — gotreesitter pure-Go Perl parser

## What This Design Does NOT Change

- The inference engine (except adding `AllSymbols()` to SymbolTable)
- The type system
- The parser
- The existing LSPServer methods
- The `psc check` command
- Any other PVM component
