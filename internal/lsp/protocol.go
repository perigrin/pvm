// ABOUTME: LSP protocol types and structures
// ABOUTME: Defines JSON-RPC and LSP message formats according to specification

package lsp

import (
	"encoding/json"
)

// JSON-RPC Message types

// JSONRPCMessage represents a JSON-RPC message
type JSONRPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSONRPCNotification represents a JSON-RPC notification
type JSONRPCNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// LSP Basic Types

// Position represents a position in a text document
type Position struct {
	Line      int `json:"line"`      // 0-based line number
	Character int `json:"character"` // 0-based character offset
}

// Range represents a range in a text document
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a text document
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// TextDocumentIdentifier represents a text document identifier
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// VersionedTextDocumentIdentifier represents a versioned text document identifier
type VersionedTextDocumentIdentifier struct {
	TextDocumentIdentifier
	Version int `json:"version"`
}

// TextDocumentItem represents a text document item
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// TextDocumentContentChangeEvent represents a change to a text document
type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitempty"`       // nil for full document sync
	RangeLength *int   `json:"rangeLength,omitempty"` // deprecated
	Text        string `json:"text"`
}

// Initialize Request/Response

// InitializeParams represents the parameters for the initialize request
type InitializeParams struct {
	ProcessID             *int               `json:"processId"`
	ClientInfo            *ClientInfo        `json:"clientInfo,omitempty"`
	Locale                string             `json:"locale,omitempty"`
	RootPath              *string            `json:"rootPath,omitempty"` // deprecated
	RootURI               *string            `json:"rootUri"`
	InitializationOptions interface{}        `json:"initializationOptions,omitempty"`
	Capabilities          ClientCapabilities `json:"capabilities"`
	WorkspaceFolders      []WorkspaceFolder  `json:"workspaceFolders,omitempty"`
}

// ClientInfo represents information about the client
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// ClientCapabilities represents client capabilities
type ClientCapabilities struct {
	Workspace    WorkspaceClientCapabilities    `json:"workspace,omitempty"`
	TextDocument TextDocumentClientCapabilities `json:"textDocument,omitempty"`
	Window       WindowClientCapabilities       `json:"window,omitempty"`
	General      GeneralClientCapabilities      `json:"general,omitempty"`
}

// WorkspaceClientCapabilities represents workspace capabilities
type WorkspaceClientCapabilities struct {
	ApplyEdit              bool                               `json:"applyEdit,omitempty"`
	WorkspaceEdit          WorkspaceEditClientCapabilities    `json:"workspaceEdit,omitempty"`
	DidChangeConfiguration DidChangeConfigurationCapabilities `json:"didChangeConfiguration,omitempty"`
	DidChangeWatchedFiles  DidChangeWatchedFilesCapabilities  `json:"didChangeWatchedFiles,omitempty"`
	Symbol                 WorkspaceSymbolClientCapabilities  `json:"symbol,omitempty"`
	ExecuteCommand         ExecuteCommandClientCapabilities   `json:"executeCommand,omitempty"`
	Configuration          bool                               `json:"configuration,omitempty"`
	WorkspaceFolders       bool                               `json:"workspaceFolders,omitempty"`
}

// TextDocumentClientCapabilities represents text document capabilities
type TextDocumentClientCapabilities struct {
	Synchronization    TextDocumentSyncClientCapabilities   `json:"synchronization,omitempty"`
	Completion         CompletionClientCapabilities         `json:"completion,omitempty"`
	Hover              HoverClientCapabilities              `json:"hover,omitempty"`
	SignatureHelp      SignatureHelpClientCapabilities      `json:"signatureHelp,omitempty"`
	Declaration        DeclarationClientCapabilities        `json:"declaration,omitempty"`
	Definition         DefinitionClientCapabilities         `json:"definition,omitempty"`
	TypeDefinition     TypeDefinitionClientCapabilities     `json:"typeDefinition,omitempty"`
	Implementation     ImplementationClientCapabilities     `json:"implementation,omitempty"`
	References         ReferenceClientCapabilities          `json:"references,omitempty"`
	DocumentHighlight  DocumentHighlightClientCapabilities  `json:"documentHighlight,omitempty"`
	DocumentSymbol     DocumentSymbolClientCapabilities     `json:"documentSymbol,omitempty"`
	CodeAction         CodeActionClientCapabilities         `json:"codeAction,omitempty"`
	CodeLens           CodeLensClientCapabilities           `json:"codeLens,omitempty"`
	DocumentLink       DocumentLinkClientCapabilities       `json:"documentLink,omitempty"`
	ColorProvider      DocumentColorClientCapabilities      `json:"colorProvider,omitempty"`
	Formatting         DocumentFormattingClientCapabilities `json:"formatting,omitempty"`
	RangeFormatting    DocumentRangeFormattingCapabilities  `json:"rangeFormatting,omitempty"`
	OnTypeFormatting   DocumentOnTypeFormattingCapabilities `json:"onTypeFormatting,omitempty"`
	Rename             RenameClientCapabilities             `json:"rename,omitempty"`
	PublishDiagnostics PublishDiagnosticsClientCapabilities `json:"publishDiagnostics,omitempty"`
	FoldingRange       FoldingRangeClientCapabilities       `json:"foldingRange,omitempty"`
}

// WindowClientCapabilities represents window capabilities
type WindowClientCapabilities struct {
	WorkDoneProgress bool `json:"workDoneProgress,omitempty"`
	ShowMessage      bool `json:"showMessage,omitempty"`
	ShowDocument     bool `json:"showDocument,omitempty"`
}

// GeneralClientCapabilities represents general capabilities
type GeneralClientCapabilities struct {
	RegularExpressions MarkupContent `json:"regularExpressions,omitempty"`
	Markdown           MarkupContent `json:"markdown,omitempty"`
}

// WorkspaceFolder represents a workspace folder
type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// InitializeResult represents the result of an initialize request
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

// ServerInfo represents information about the server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// Text Document Synchronization

// DidOpenTextDocumentParams represents parameters for textDocument/didOpen
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// DidChangeTextDocumentParams represents parameters for textDocument/didChange
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// DidCloseTextDocumentParams represents parameters for textDocument/didClose
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// Diagnostics

// Diagnostic represents a diagnostic message
type Diagnostic struct {
	Range              Range                   `json:"range"`
	Severity           *DiagnosticSeverity     `json:"severity,omitempty"`
	Code               interface{}             `json:"code,omitempty"`
	CodeDescription    *CodeDescription        `json:"codeDescription,omitempty"`
	Source             string                  `json:"source,omitempty"`
	Message            string                  `json:"message"`
	Tags               []DiagnosticTag         `json:"tags,omitempty"`
	RelatedInformation []DiagnosticRelatedInfo `json:"relatedInformation,omitempty"`
	Data               interface{}             `json:"data,omitempty"`
}

// DiagnosticSeverity represents the severity of a diagnostic
type DiagnosticSeverity int

const (
	DiagnosticSeverityError       DiagnosticSeverity = 1
	DiagnosticSeverityWarning     DiagnosticSeverity = 2
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	DiagnosticSeverityHint        DiagnosticSeverity = 4
)

// DiagnosticTag represents a diagnostic tag
type DiagnosticTag int

const (
	DiagnosticTagUnnecessary DiagnosticTag = 1
	DiagnosticTagDeprecated  DiagnosticTag = 2
)

// CodeDescription represents additional information about a diagnostic code
type CodeDescription struct {
	Href string `json:"href"`
}

// DiagnosticRelatedInfo represents related information for a diagnostic
type DiagnosticRelatedInfo struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

// PublishDiagnosticsParams represents parameters for textDocument/publishDiagnostics
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Version     *int         `json:"version,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Hover

// HoverParams represents parameters for textDocument/hover
type HoverParams struct {
	TextDocumentPositionParams
}

// TextDocumentPositionParams represents common parameters for text document position requests
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// Hover represents a hover response
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// MarkupContent represents markup content
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// MarkupKind represents markup kinds
const (
	MarkupKindPlainText = "plaintext"
	MarkupKindMarkdown  = "markdown"
)

// Completion

// CompletionParams represents parameters for textDocument/completion
type CompletionParams struct {
	TextDocumentPositionParams
	Context *CompletionContext `json:"context,omitempty"`
}

// CompletionContext represents completion context
type CompletionContext struct {
	TriggerKind      CompletionTriggerKind `json:"triggerKind"`
	TriggerCharacter string                `json:"triggerCharacter,omitempty"`
}

// CompletionTriggerKind represents completion trigger kinds
type CompletionTriggerKind int

const (
	CompletionTriggerKindInvoked                        CompletionTriggerKind = 1
	CompletionTriggerKindTriggerCharacter               CompletionTriggerKind = 2
	CompletionTriggerKindTriggerForIncompleteCompletion CompletionTriggerKind = 3
)

// CompletionList represents a completion response
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// CompletionItem represents a completion item
type CompletionItem struct {
	Label               string              `json:"label"`
	Kind                *CompletionItemKind `json:"kind,omitempty"`
	Tags                []CompletionItemTag `json:"tags,omitempty"`
	Detail              string              `json:"detail,omitempty"`
	Documentation       *MarkupContent      `json:"documentation,omitempty"`
	Deprecated          bool                `json:"deprecated,omitempty"`
	Preselect           bool                `json:"preselect,omitempty"`
	SortText            string              `json:"sortText,omitempty"`
	FilterText          string              `json:"filterText,omitempty"`
	InsertText          string              `json:"insertText,omitempty"`
	InsertTextFormat    *InsertTextFormat   `json:"insertTextFormat,omitempty"`
	InsertTextMode      *InsertTextMode     `json:"insertTextMode,omitempty"`
	TextEdit            *TextEdit           `json:"textEdit,omitempty"`
	AdditionalTextEdits []TextEdit          `json:"additionalTextEdits,omitempty"`
	CommitCharacters    []string            `json:"commitCharacters,omitempty"`
	Command             *Command            `json:"command,omitempty"`
	Data                interface{}         `json:"data,omitempty"`
}

// CompletionItemKind represents completion item kinds
type CompletionItemKind int

const (
	CompletionItemKindText          CompletionItemKind = 1
	CompletionItemKindMethod        CompletionItemKind = 2
	CompletionItemKindFunction      CompletionItemKind = 3
	CompletionItemKindConstructor   CompletionItemKind = 4
	CompletionItemKindField         CompletionItemKind = 5
	CompletionItemKindVariable      CompletionItemKind = 6
	CompletionItemKindClass         CompletionItemKind = 7
	CompletionItemKindInterface     CompletionItemKind = 8
	CompletionItemKindModule        CompletionItemKind = 9
	CompletionItemKindProperty      CompletionItemKind = 10
	CompletionItemKindUnit          CompletionItemKind = 11
	CompletionItemKindValue         CompletionItemKind = 12
	CompletionItemKindEnum          CompletionItemKind = 13
	CompletionItemKindKeyword       CompletionItemKind = 14
	CompletionItemKindSnippet       CompletionItemKind = 15
	CompletionItemKindColor         CompletionItemKind = 16
	CompletionItemKindFile          CompletionItemKind = 17
	CompletionItemKindReference     CompletionItemKind = 18
	CompletionItemKindFolder        CompletionItemKind = 19
	CompletionItemKindEnumMember    CompletionItemKind = 20
	CompletionItemKindConstant      CompletionItemKind = 21
	CompletionItemKindStruct        CompletionItemKind = 22
	CompletionItemKindEvent         CompletionItemKind = 23
	CompletionItemKindOperator      CompletionItemKind = 24
	CompletionItemKindTypeParameter CompletionItemKind = 25
)

// CompletionItemTag represents completion item tags
type CompletionItemTag int

const (
	CompletionItemTagDeprecated CompletionItemTag = 1
)

// InsertTextFormat represents insert text formats
type InsertTextFormat int

const (
	InsertTextFormatPlainText InsertTextFormat = 1
	InsertTextFormatSnippet   InsertTextFormat = 2
)

// InsertTextMode represents insert text modes
type InsertTextMode int

const (
	InsertTextModeAsIs              InsertTextMode = 1
	InsertTextModeAdjustIndentation InsertTextMode = 2
)

// TextEdit represents a text edit
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// Command represents a command
type Command struct {
	Title     string        `json:"title"`
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments,omitempty"`
}

// Placeholder types for client capabilities (simplified for this implementation)
type WorkspaceEditClientCapabilities struct{}
type DidChangeConfigurationCapabilities struct{}
type DidChangeWatchedFilesCapabilities struct{}
type WorkspaceSymbolClientCapabilities struct{}
type ExecuteCommandClientCapabilities struct{}
type TextDocumentSyncClientCapabilities struct{}
type CompletionClientCapabilities struct{}
type HoverClientCapabilities struct{}
type SignatureHelpClientCapabilities struct{}
type DeclarationClientCapabilities struct{}
type DefinitionClientCapabilities struct{}
type TypeDefinitionClientCapabilities struct{}
type ImplementationClientCapabilities struct{}
type ReferenceClientCapabilities struct{}
type DocumentHighlightClientCapabilities struct{}
type DocumentSymbolClientCapabilities struct{}
type CodeActionClientCapabilities struct{}
type CodeLensClientCapabilities struct{}
type DocumentLinkClientCapabilities struct{}
type DocumentColorClientCapabilities struct{}
type DocumentFormattingClientCapabilities struct{}
type DocumentRangeFormattingCapabilities struct{}
type DocumentOnTypeFormattingCapabilities struct{}
type RenameClientCapabilities struct{}
type PublishDiagnosticsClientCapabilities struct{}
type FoldingRangeClientCapabilities struct{}

// DefinitionParams represents parameters for textDocument/definition
type DefinitionParams struct {
	TextDocumentPositionParams
}

// ReferenceParams represents parameters for textDocument/references
type ReferenceParams struct {
	TextDocumentPositionParams
	Context ReferenceContext `json:"context"`
}

// ReferenceContext provides additional context for references
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// DocumentFormattingParams represents parameters for textDocument/formatting
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

// FormattingOptions represents formatting options
type FormattingOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}

// CodeActionParams represents parameters for textDocument/codeAction
type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

// CodeActionContext provides context for code actions
type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// CodeAction represents a code action
type CodeAction struct {
	Title       string         `json:"title"`
	Kind        string         `json:"kind,omitempty"`
	Diagnostics []Diagnostic   `json:"diagnostics,omitempty"`
	Edit        *WorkspaceEdit `json:"edit,omitempty"`
	Command     *Command       `json:"command,omitempty"`
}

// WorkspaceEdit represents changes to multiple text documents
type WorkspaceEdit struct {
	Changes map[string][]TextEdit `json:"changes"`
}

// Workspace Symbol

// WorkspaceSymbolParams represents parameters for workspace/symbol
type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}

// WorkspaceSymbol represents a workspace symbol
type WorkspaceSymbol struct {
	Name     string      `json:"name"`
	Kind     SymbolKind  `json:"kind"`
	Tags     []SymbolTag `json:"tags,omitempty"`
	Location Location    `json:"location"`
	Data     interface{} `json:"data,omitempty"`
}

// SymbolKind represents symbol kinds
type SymbolKind int

const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

// SymbolTag represents symbol tags
type SymbolTag int

const (
	SymbolTagDeprecated SymbolTag = 1
)

// Inlay Hints

// InlayHintParams represents parameters for textDocument/inlayHint
type InlayHintParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
}

// InlayHint represents an inlay hint
type InlayHint struct {
	Position     Position       `json:"position"`
	Label        string         `json:"label"`
	Kind         InlayHintKind  `json:"kind,omitempty"`
	TextEdits    []TextEdit     `json:"textEdits,omitempty"`
	Tooltip      *MarkupContent `json:"tooltip,omitempty"`
	PaddingLeft  bool           `json:"paddingLeft,omitempty"`
	PaddingRight bool           `json:"paddingRight,omitempty"`
	Data         interface{}    `json:"data,omitempty"`
}

// InlayHintKind represents inlay hint kinds
type InlayHintKind int

const (
	InlayHintKindType      InlayHintKind = 1
	InlayHintKindParameter InlayHintKind = 2
)

// Semantic Tokens

// SemanticTokensParams represents parameters for textDocument/semanticTokens/full
type SemanticTokensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SemanticTokens represents semantic tokens
type SemanticTokens struct {
	ResultID string   `json:"resultId,omitempty"`
	Data     []uint32 `json:"data"`
}
