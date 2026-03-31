// ABOUTME: LSP protocol type definitions for the PSC language server.
// ABOUTME: Minimal set of structs matching the LSP specification for diagnostics, hover, definition, and symbols.

package psc

// JSON-RPC and LSP error codes.
const (
	codeServerNotInitialized = -32002
	codeMethodNotFound       = -32601
)

// LSP symbol kind constants (subset used by PSC).
const (
	symbolKindModule   = 2
	symbolKindMethod   = 6
	symbolKindFunction = 12
	symbolKindVariable = 13
)

// lspInitializeParams carries the client capabilities sent in the initialize request.
type lspInitializeParams struct {
	ProcessID int    `json:"processId"`
	RootURI   string `json:"rootUri"`
}

// lspServerCapabilities declares what LSP features the server supports.
type lspServerCapabilities struct {
	TextDocumentSync       int  `json:"textDocumentSync"`
	HoverProvider          bool `json:"hoverProvider"`
	DefinitionProvider     bool `json:"definitionProvider"`
	DocumentSymbolProvider bool `json:"documentSymbolProvider"`
}

// lspInitializeResult is the response body for the initialize request.
type lspInitializeResult struct {
	Capabilities lspServerCapabilities `json:"capabilities"`
}

// lspTextDocumentItem represents a text document sent by the client on open.
type lspTextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// lspDidOpenParams is the params payload for textDocument/didOpen notifications.
type lspDidOpenParams struct {
	TextDocument lspTextDocumentItem `json:"textDocument"`
}

// lspVersionedTextDocumentIdentifier identifies a document along with its version.
type lspVersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

// lspTextDocumentContentChangeEvent carries a full-document content replacement.
// PSC uses full sync (TextDocumentSync = 1), so range fields are omitted.
type lspTextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

// lspDidChangeParams is the params payload for textDocument/didChange notifications.
type lspDidChangeParams struct {
	TextDocument   lspVersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []lspTextDocumentContentChangeEvent `json:"contentChanges"`
}

// lspTextDocumentIdentifier identifies a text document by URI.
type lspTextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// lspDidCloseParams is the params payload for textDocument/didClose notifications.
type lspDidCloseParams struct {
	TextDocument lspTextDocumentIdentifier `json:"textDocument"`
}

// lspPosition is a zero-based line/character position within a text document.
type lspPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// lspRange is a span within a text document defined by start and end positions.
type lspRange struct {
	Start lspPosition `json:"start"`
	End   lspPosition `json:"end"`
}

// lspDiagnostic represents a compiler or type-inference diagnostic message.
type lspDiagnostic struct {
	Range    lspRange `json:"range"`
	Severity int      `json:"severity"`
	Source   string   `json:"source"`
	Message  string   `json:"message"`
	Code     string   `json:"code,omitempty"`
}

// lspPublishDiagnosticsParams is the params payload for textDocument/publishDiagnostics notifications.
type lspPublishDiagnosticsParams struct {
	URI         string          `json:"uri"`
	Diagnostics []lspDiagnostic `json:"diagnostics"`
}

// lspTextDocumentPositionParams identifies a position within a specific document.
type lspTextDocumentPositionParams struct {
	TextDocument lspTextDocumentIdentifier `json:"textDocument"`
	Position     lspPosition               `json:"position"`
}

// lspMarkupContent holds a string value in a given markup kind (e.g. "markdown").
type lspMarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// lspHover is the response body for textDocument/hover requests.
type lspHover struct {
	Contents lspMarkupContent `json:"contents"`
}

// lspLocation is a source location (URI + range) used in definition responses.
type lspLocation struct {
	URI   string   `json:"uri"`
	Range lspRange `json:"range"`
}

// lspDocumentSymbolParams is the params payload for textDocument/documentSymbol requests.
type lspDocumentSymbolParams struct {
	TextDocument lspTextDocumentIdentifier `json:"textDocument"`
}

// lspDocumentSymbol represents a symbol (function, variable, class, etc.) in a document.
type lspDocumentSymbol struct {
	Name           string              `json:"name"`
	Kind           int                 `json:"kind"`
	Range          lspRange            `json:"range"`
	SelectionRange lspRange            `json:"selectionRange"`
	Children       []lspDocumentSymbol `json:"children,omitempty"`
}
