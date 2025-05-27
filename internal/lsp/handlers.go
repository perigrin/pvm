// ABOUTME: LSP message handlers for PSC language server
// ABOUTME: Implements request/response handling for various LSP methods

package lsp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tamarou.com/pvm/internal/ls"
	"tamarou.com/pvm/internal/parser"
)

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(msg *JSONRPCMessage) error {
	var params InitializeParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Initialize request from client: %s", getClientName(params.ClientInfo))

	result := InitializeResult{
		Capabilities: *s.capabilities,
		ServerInfo: &ServerInfo{
			Name:    "PSC Language Server",
			Version: "1.0.0",
		},
	}

	return s.sendResponse(msg.ID, result)
}

// handleInitialized handles the initialized notification
func (s *Server) handleInitialized(msg *JSONRPCMessage) error {
	s.mutex.Lock()
	s.initialized = true
	s.mutex.Unlock()

	s.logger.Println("Server initialized")
	return nil
}

// handleShutdown handles the shutdown request
func (s *Server) handleShutdown(msg *JSONRPCMessage) error {
	s.mutex.Lock()
	s.shutdown = true
	s.mutex.Unlock()

	s.logger.Println("Shutdown request received")
	return s.sendResponse(msg.ID, nil)
}

// handleExit handles the exit notification
func (s *Server) handleExit(msg *JSONRPCMessage) error {
	s.logger.Println("Exit notification received")
	s.Stop()
	return nil
}

// handleTextDocumentDidOpen handles the textDocument/didOpen notification
func (s *Server) handleTextDocumentDidOpen(msg *JSONRPCMessage) error {
	var params DidOpenTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Document opened: %s", params.TextDocument.URI)

	// Update document in language service
	err := s.updateDocumentInLanguageService(
		params.TextDocument.URI,
		params.TextDocument.Text,
		params.TextDocument.Version,
	)
	if err != nil {
		s.logger.Printf("Failed to update document in language service: %v", err)
	}

	// Publish diagnostics from language service
	return s.publishDiagnosticsFromLanguageService(params.TextDocument.URI)
}

// handleTextDocumentDidChange handles the textDocument/didChange notification
func (s *Server) handleTextDocumentDidChange(msg *JSONRPCMessage) error {
	var params DidChangeTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Document changed: %s", params.TextDocument.URI)

	// Update document content (full sync)
	if len(params.ContentChanges) > 0 {
		// Update document in language service
		err := s.updateDocumentInLanguageService(
			params.TextDocument.URI,
			params.ContentChanges[0].Text,
			params.TextDocument.Version,
		)
		if err != nil {
			s.logger.Printf("Failed to update document in language service: %v", err)
		}

		// Publish updated diagnostics from language service
		return s.publishDiagnosticsFromLanguageService(params.TextDocument.URI)
	}

	return nil
}

// handleTextDocumentDidClose handles the textDocument/didClose notification
func (s *Server) handleTextDocumentDidClose(msg *JSONRPCMessage) error {
	var params DidCloseTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Document closed: %s", params.TextDocument.URI)

	// Clear diagnostics for the closed document
	return s.publishDiagnostics(params.TextDocument.URI, []parser.TypeCheckError{})
}

// handleTextDocumentHover handles the textDocument/hover request
func (s *Server) handleTextDocumentHover(msg *JSONRPCMessage) error {
	var params HoverParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Hover request for %s at %d:%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	// Convert LSP position to language service position
	lsPos := convertLSPPosition(params.Position)

	// Get hover information from language service
	hoverInfo, err := s.languageService.GetHover(params.TextDocument.URI, lsPos)
	if err != nil {
		s.logger.Printf("Failed to get hover info: %v", err)
		return s.sendResponse(msg.ID, nil)
	}

	if hoverInfo == nil {
		return s.sendResponse(msg.ID, nil)
	}

	// Convert language service hover to LSP hover
	lspHover := convertToLSPHover(hoverInfo)
	return s.sendResponse(msg.ID, lspHover)
}

// handleTextDocumentCompletion handles the textDocument/completion request
func (s *Server) handleTextDocumentCompletion(msg *JSONRPCMessage) error {
	var params CompletionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Completion request for %s at %d:%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	// Convert LSP position to language service position
	lsPos := convertLSPPosition(params.Position)

	// Get completions from language service
	completions, err := s.languageService.GetCompletions(params.TextDocument.URI, lsPos)
	if err != nil {
		s.logger.Printf("Failed to get completions: %v", err)
		return s.sendResponse(msg.ID, CompletionList{
			IsIncomplete: false,
			Items:        []CompletionItem{},
		})
	}

	// Convert language service completions to LSP completions
	lspCompletions := convertToLSPCompletions(completions)

	result := CompletionList{
		IsIncomplete: false,
		Items:        lspCompletions,
	}

	return s.sendResponse(msg.ID, result)
}

// handleTextDocumentDefinition handles the textDocument/definition request
func (s *Server) handleTextDocumentDefinition(msg *JSONRPCMessage) error {
	var params DefinitionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Definition request for %s at %d:%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	// Convert LSP position to language service position
	lsPos := convertLSPPosition(params.Position)

	// Get definition from language service
	definition, err := s.languageService.GetDefinition(params.TextDocument.URI, lsPos)
	if err != nil {
		s.logger.Printf("Failed to get definition: %v", err)
		return s.sendResponse(msg.ID, nil)
	}

	if definition == nil {
		return s.sendResponse(msg.ID, nil)
	}

	// Convert language service definition to LSP location
	lspLocation := convertToLSPLocation(definition.Location)
	return s.sendResponse(msg.ID, lspLocation)
}

// handleTextDocumentReferences handles the textDocument/references request
func (s *Server) handleTextDocumentReferences(msg *JSONRPCMessage) error {
	var params ReferenceParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("References request for %s at %d:%d (includeDecl: %v)",
		params.TextDocument.URI, params.Position.Line, params.Position.Character,
		params.Context.IncludeDeclaration)

	// Convert LSP position to language service position
	lsPos := convertLSPPosition(params.Position)

	// Find references using language service
	locations, err := s.languageService.FindReferences(params.TextDocument.URI, lsPos, params.Context.IncludeDeclaration)
	if err != nil {
		s.logger.Printf("Failed to find references: %v", err)
		return s.sendResponse(msg.ID, []Location{})
	}

	// Convert language service locations to LSP locations
	lspLocations := convertToLSPLocations(locations)
	return s.sendResponse(msg.ID, lspLocations)
}

// handleTextDocumentFormatting handles the textDocument/formatting request
func (s *Server) handleTextDocumentFormatting(msg *JSONRPCMessage) error {
	var params DocumentFormattingParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Formatting request for %s", params.TextDocument.URI)

	// Formatting not yet implemented with language service
	// Return empty edits for now
	return s.sendResponse(msg.ID, []TextEdit{})
}

// handleTextDocumentCodeAction handles the textDocument/codeAction request
func (s *Server) handleTextDocumentCodeAction(msg *JSONRPCMessage) error {
	var params CodeActionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Code action request for %s at %d:%d-%d:%d",
		params.TextDocument.URI,
		params.Range.Start.Line, params.Range.Start.Character,
		params.Range.End.Line, params.Range.End.Character)

	// Code actions not yet implemented with language service
	// Return empty actions for now
	return s.sendResponse(msg.ID, []CodeAction{})
}

// generateHoverInfo generates hover information for a position in a document
func (s *Server) generateHoverInfo(doc *Document, pos Position) *Hover {
	// Extract the word at the cursor position
	lines := strings.Split(doc.Text, "\n")
	if pos.Line >= len(lines) {
		return nil
	}

	line := lines[pos.Line]
	if pos.Character >= len(line) {
		return nil
	}

	// Find word boundaries
	start := pos.Character
	end := pos.Character

	// Move start backwards to beginning of word
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}

	// Move end forwards to end of word
	for end < len(line) && isWordChar(line[end]) {
		end++
	}

	if start == end {
		return nil
	}

	word := line[start:end]

	// Generate hover content based on the word
	var content string
	kind := MarkupKindMarkdown

	// Determine content based on word type
	switch {
	case isTypeAnnotation(word):
		content = fmt.Sprintf("**Type**: `%s`\n\n%s", word, getTypeDescription(word))
	case isBuiltinFunction(word):
		content = fmt.Sprintf("**Builtin Function**: `%s`\n\n%s", word, getBuiltinDescription(word))
	case isPerlKeyword(word):
		content = fmt.Sprintf("**Keyword**: `%s`\n\n%s", word, getKeywordDescription(word))
	default:
		// Try to find type information from the document analysis
		content = fmt.Sprintf("**Symbol**: `%s`", word)
	}

	return &Hover{
		Contents: MarkupContent{
			Kind:  kind,
			Value: content,
		},
		Range: &Range{
			Start: Position{Line: pos.Line, Character: start},
			End:   Position{Line: pos.Line, Character: end},
		},
	}
}

// generateCompletions generates completion items for a position in a document
func (s *Server) generateCompletions(doc *Document, pos Position, context *CompletionContext) []CompletionItem {
	var items []CompletionItem

	// Get the current line
	lines := strings.Split(doc.Text, "\n")
	if pos.Line >= len(lines) {
		return items
	}

	line := lines[pos.Line]
	prefix := ""
	if pos.Character <= len(line) {
		prefix = line[:pos.Character]
	}

	// Determine what kind of completions to offer based on context
	triggerChar := ""
	if context != nil && context.TriggerKind == CompletionTriggerKindTriggerCharacter {
		triggerChar = context.TriggerCharacter
	}

	switch triggerChar {
	case "$":
		// Variable completions
		items = append(items, s.getVariableCompletions()...)
	case "@":
		// Array completions
		items = append(items, s.getArrayCompletions()...)
	case "%":
		// Hash completions
		items = append(items, s.getHashCompletions()...)
	case ":":
		if strings.HasSuffix(prefix, "::") {
			// Module method completions
			items = append(items, s.getModuleCompletions(prefix)...)
		}
	case ".":
		// Object method completions
		items = append(items, s.getMethodCompletions()...)
	case "-":
		if strings.HasSuffix(prefix, "->") {
			// Dereference completions
			items = append(items, s.getDereferenceCompletions()...)
		}
	default:
		// General completions
		items = append(items, s.getKeywordCompletions()...)
		items = append(items, s.getTypeCompletions()...)
		items = append(items, s.getBuiltinCompletions()...)
	}

	return items
}

// Helper functions for hover and completion

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_' || c == ':'
}

func isTypeAnnotation(word string) bool {
	types := []string{"Str", "Int", "Num", "Bool", "ArrayRef", "HashRef", "CodeRef", "Any", "Undef", "Maybe"}
	for _, t := range types {
		if word == t {
			return true
		}
	}
	return false
}

func isBuiltinFunction(word string) bool {
	builtins := []string{"print", "say", "defined", "ref", "substr", "length", "chomp", "split", "join", "grep", "map", "sort"}
	for _, b := range builtins {
		if word == b {
			return true
		}
	}
	return false
}

func isPerlKeyword(word string) bool {
	keywords := []string{"my", "our", "local", "sub", "if", "elsif", "else", "while", "for", "foreach", "use", "package"}
	for _, k := range keywords {
		if word == k {
			return true
		}
	}
	return false
}

func getTypeDescription(typeName string) string {
	descriptions := map[string]string{
		"Str":      "String type - represents text values",
		"Int":      "Integer type - represents whole numbers",
		"Num":      "Number type - represents numeric values",
		"Bool":     "Boolean type - represents true/false values",
		"ArrayRef": "Array reference type - reference to an array",
		"HashRef":  "Hash reference type - reference to a hash",
		"CodeRef":  "Code reference type - reference to a subroutine",
		"Any":      "Any type - accepts any value",
		"Undef":    "Undefined type - represents undefined values",
		"Maybe":    "Maybe type - optional value that can be undef",
	}

	if desc, ok := descriptions[typeName]; ok {
		return desc
	}
	return "User-defined type"
}

func getBuiltinDescription(funcName string) string {
	descriptions := map[string]string{
		"print":   "Print values to STDOUT",
		"say":     "Print values to STDOUT with newline",
		"defined": "Test whether a value is defined",
		"ref":     "Return reference type of a value",
		"substr":  "Extract substring from a string",
		"length":  "Return length of a string or array",
		"chomp":   "Remove trailing newline",
		"split":   "Split string into array",
		"join":    "Join array elements into string",
		"grep":    "Filter array elements",
		"map":     "Transform array elements",
		"sort":    "Sort array elements",
	}

	if desc, ok := descriptions[funcName]; ok {
		return desc
	}
	return "Perl builtin function"
}

func getKeywordDescription(keyword string) string {
	descriptions := map[string]string{
		"my":      "Declare lexical variable",
		"our":     "Declare package variable",
		"local":   "Temporarily localize variable",
		"sub":     "Define subroutine",
		"if":      "Conditional statement",
		"elsif":   "Additional condition",
		"else":    "Default condition",
		"while":   "Loop while condition is true",
		"for":     "C-style for loop",
		"foreach": "Loop over list",
		"use":     "Load and import module",
		"package": "Declare package namespace",
	}

	if desc, ok := descriptions[keyword]; ok {
		return desc
	}
	return "Perl keyword"
}

// Completion generators

func (s *Server) getVariableCompletions() []CompletionItem {
	return []CompletionItem{
		{Label: "$_", Kind: &[]CompletionItemKind{CompletionItemKindVariable}[0], Detail: "Default variable"},
		{Label: "$@", Kind: &[]CompletionItemKind{CompletionItemKindVariable}[0], Detail: "Error variable"},
		{Label: "$!", Kind: &[]CompletionItemKind{CompletionItemKindVariable}[0], Detail: "System error variable"},
		{Label: "$$", Kind: &[]CompletionItemKind{CompletionItemKindVariable}[0], Detail: "Process ID"},
	}
}

func (s *Server) getArrayCompletions() []CompletionItem {
	return []CompletionItem{
		{Label: "@_", Kind: &[]CompletionItemKind{CompletionItemKindVariable}[0], Detail: "Subroutine arguments"},
		{Label: "@ARGV", Kind: &[]CompletionItemKind{CompletionItemKindVariable}[0], Detail: "Command line arguments"},
	}
}

func (s *Server) getHashCompletions() []CompletionItem {
	return []CompletionItem{
		{Label: "%ENV", Kind: &[]CompletionItemKind{CompletionItemKindVariable}[0], Detail: "Environment variables"},
		{Label: "%SIG", Kind: &[]CompletionItemKind{CompletionItemKindVariable}[0], Detail: "Signal handlers"},
	}
}

func (s *Server) getModuleCompletions(prefix string) []CompletionItem {
	// Extract module name from prefix
	parts := strings.Split(prefix, "::")
	if len(parts) < 2 {
		return []CompletionItem{}
	}

	// Common module completions
	return []CompletionItem{
		{Label: "new", Kind: &[]CompletionItemKind{CompletionItemKindConstructor}[0], Detail: "Constructor method"},
		{Label: "DESTROY", Kind: &[]CompletionItemKind{CompletionItemKindMethod}[0], Detail: "Destructor method"},
	}
}

func (s *Server) getMethodCompletions() []CompletionItem {
	return []CompletionItem{
		{Label: "can", Kind: &[]CompletionItemKind{CompletionItemKindMethod}[0], Detail: "Check if method exists"},
		{Label: "isa", Kind: &[]CompletionItemKind{CompletionItemKindMethod}[0], Detail: "Check object type"},
		{Label: "DOES", Kind: &[]CompletionItemKind{CompletionItemKindMethod}[0], Detail: "Check role implementation"},
	}
}

func (s *Server) getDereferenceCompletions() []CompletionItem {
	return []CompletionItem{
		{Label: "[]", Kind: &[]CompletionItemKind{CompletionItemKindOperator}[0], Detail: "Array dereference"},
		{Label: "{}", Kind: &[]CompletionItemKind{CompletionItemKindOperator}[0], Detail: "Hash dereference"},
		{Label: "()", Kind: &[]CompletionItemKind{CompletionItemKindOperator}[0], Detail: "Code dereference"},
	}
}

func (s *Server) getKeywordCompletions() []CompletionItem {
	keywords := []string{"my", "our", "local", "sub", "if", "elsif", "else", "while", "for", "foreach", "use", "package", "return", "last", "next", "redo"}
	items := make([]CompletionItem, len(keywords))

	for i, keyword := range keywords {
		items[i] = CompletionItem{
			Label:  keyword,
			Kind:   &[]CompletionItemKind{CompletionItemKindKeyword}[0],
			Detail: getKeywordDescription(keyword),
		}
	}

	return items
}

func (s *Server) getTypeCompletions() []CompletionItem {
	types := []string{"Str", "Int", "Num", "Bool", "ArrayRef", "HashRef", "CodeRef", "Any", "Undef", "Maybe"}
	items := make([]CompletionItem, len(types))

	for i, typeName := range types {
		items[i] = CompletionItem{
			Label:  typeName,
			Kind:   &[]CompletionItemKind{CompletionItemKindClass}[0],
			Detail: getTypeDescription(typeName),
		}
	}

	return items
}

func (s *Server) getBuiltinCompletions() []CompletionItem {
	builtins := []string{"print", "say", "defined", "ref", "substr", "length", "chomp", "split", "join", "grep", "map", "sort", "push", "pop", "shift", "unshift", "splice"}
	items := make([]CompletionItem, len(builtins))

	for i, builtin := range builtins {
		items[i] = CompletionItem{
			Label:  builtin,
			Kind:   &[]CompletionItemKind{CompletionItemKindFunction}[0],
			Detail: getBuiltinDescription(builtin),
		}
	}

	return items
}

// writeDocumentToTempFile writes document content to a temporary file for analysis
// and returns the path to the temporary file
func (s *Server) writeDocumentToTempFile(doc *Document) (string, error) {
	// Create a temporary file with Perl extension
	tempDir := os.TempDir()
	fileName := fmt.Sprintf("lsp_%d_%s.pl", time.Now().Unix(), filepath.Base(uriToPath(doc.URI)))
	tempPath := filepath.Join(tempDir, fileName)

	err := os.WriteFile(tempPath, []byte(doc.Text), 0644)
	return tempPath, err
}

// Conversion helper functions between LSP and language service types

// convertLSPPosition converts LSP position to language service position
func convertLSPPosition(lspPos Position) ls.Position {
	return ls.Position{
		Line:      lspPos.Line,
		Character: lspPos.Character,
	}
}

// convertToLSPHover converts language service hover to LSP hover
func convertToLSPHover(lsHover *ls.Hover) *Hover {
	if lsHover == nil {
		return nil
	}

	hover := &Hover{
		Contents: MarkupContent{
			Kind:  MarkupKindMarkdown,
			Value: lsHover.Contents,
		},
	}

	if lsHover.Range != nil {
		hover.Range = &Range{
			Start: Position{
				Line:      lsHover.Range.Start.Line,
				Character: lsHover.Range.Start.Character,
			},
			End: Position{
				Line:      lsHover.Range.End.Line,
				Character: lsHover.Range.End.Character,
			},
		}
	}

	return hover
}

// convertToLSPCompletions converts language service completions to LSP completions
func convertToLSPCompletions(lsCompletions []ls.CompletionItem) []CompletionItem {
	items := make([]CompletionItem, len(lsCompletions))

	for i, lsItem := range lsCompletions {
		items[i] = CompletionItem{
			Label:  lsItem.Label,
			Kind:   convertCompletionItemKind(lsItem.Kind),
			Detail: lsItem.Detail,
		}
	}

	return items
}

// convertCompletionItemKind converts language service completion kind to LSP kind
func convertCompletionItemKind(lsKind ls.CompletionItemKind) *CompletionItemKind {
	var lspKind CompletionItemKind

	switch lsKind {
	case ls.CompletionItemKindVariable:
		lspKind = CompletionItemKindVariable
	case ls.CompletionItemKindFunction:
		lspKind = CompletionItemKindFunction
	case ls.CompletionItemKindKeyword:
		lspKind = CompletionItemKindKeyword
	case ls.CompletionItemKindType:
		lspKind = CompletionItemKindClass
	case ls.CompletionItemKindMethod:
		lspKind = CompletionItemKindMethod
	case ls.CompletionItemKindModule:
		lspKind = CompletionItemKindModule
	default:
		lspKind = CompletionItemKindText
	}

	return &lspKind
}

// convertToLSPLocation converts language service location to LSP location
func convertToLSPLocation(lsLocation ls.Location) Location {
	return Location{
		URI: lsLocation.URI,
		Range: Range{
			Start: Position{
				Line:      lsLocation.Range.Start.Line,
				Character: lsLocation.Range.Start.Character,
			},
			End: Position{
				Line:      lsLocation.Range.End.Line,
				Character: lsLocation.Range.End.Character,
			},
		},
	}
}

// convertToLSPLocations converts language service locations to LSP locations
func convertToLSPLocations(lsLocations []ls.Location) []Location {
	locations := make([]Location, len(lsLocations))

	for i, lsLocation := range lsLocations {
		locations[i] = convertToLSPLocation(lsLocation)
	}

	return locations
}

// getClientName extracts client name from client info
func getClientName(clientInfo *ClientInfo) string {
	if clientInfo != nil {
		if clientInfo.Version != "" {
			return fmt.Sprintf("%s %s", clientInfo.Name, clientInfo.Version)
		}
		return clientInfo.Name
	}
	return "Unknown"
}
