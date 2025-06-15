// ABOUTME: LSP message handlers for PSC language server
// ABOUTME: Implements request/response handling for various LSP methods

package lsp

import (
	"encoding/json"
	"fmt"

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
	// Start request-scoped pooling
	requestID := fmt.Sprintf("completion_%v", msg.ID)
	_ = s.poolManager.StartRequest(requestID, "textDocument/completion")
	defer s.poolManager.EndRequest(requestID)

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

		// Use pooled completion list for empty response
		result := s.poolManager.NewCompletionList(requestID, false)
		return s.sendResponse(msg.ID, result)
	}

	// Create pooled completion list and items
	result := s.poolManager.NewCompletionList(requestID, false)

	// Convert language service completions to pooled LSP completions
	for _, comp := range completions {
		item := s.poolManager.NewCompletionItem(requestID, comp.Label, comp.Detail)

		// Convert LS completion item kind to LSP completion item kind
		var lspKind CompletionItemKind
		switch comp.Kind {
		case ls.CompletionItemKindFunction:
			lspKind = CompletionItemKindFunction
		case ls.CompletionItemKindVariable:
			lspKind = CompletionItemKindVariable
		case ls.CompletionItemKindMethod:
			lspKind = CompletionItemKindMethod
		case ls.CompletionItemKindKeyword:
			lspKind = CompletionItemKindKeyword
		case ls.CompletionItemKindType:
			lspKind = CompletionItemKindClass
		default:
			lspKind = CompletionItemKindText
		}
		item.Kind = &lspKind

		result.Items = append(result.Items, *item)
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

	// Convert LSP formatting options to language service options
	lsOptions := convertLSPFormattingOptions(params.Options)

	// Get formatting edits from language service
	edits, err := s.languageService.FormatDocument(params.TextDocument.URI, lsOptions)
	if err != nil {
		s.logger.Printf("Failed to format document: %v", err)
		return s.sendResponse(msg.ID, []TextEdit{})
	}

	// Convert language service text edits to LSP text edits
	lspEdits := convertToLSPTextEdits(edits)
	return s.sendResponse(msg.ID, lspEdits)
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

	// Convert LSP range and context to language service types
	lsRange := convertLSPRangeForCodeActions(params.Range)
	lsContext := convertLSPCodeActionContext(params.Context)

	// Get code actions from language service
	actions, err := s.languageService.GenerateCodeActions(params.TextDocument.URI, lsRange, lsContext)
	if err != nil {
		s.logger.Printf("Failed to generate code actions: %v", err)
		return s.sendResponse(msg.ID, []CodeAction{})
	}

	// Convert language service code actions to LSP code actions
	lspActions := convertToLSPCodeActions(actions)
	return s.sendResponse(msg.ID, lspActions)
}

// TODO: Legacy hover generation - replaced by language service

// Helper functions for hover and completion

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_' || c == ':'
}

// Completion generators

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

// Conversion functions for code actions and formatting

// convertLSPRangeForCodeActions converts LSP range to language service range
func convertLSPRangeForCodeActions(lspRange Range) ls.Range {
	return ls.Range{
		Start: ls.Position{
			Line:      lspRange.Start.Line,
			Character: lspRange.Start.Character,
		},
		End: ls.Position{
			Line:      lspRange.End.Line,
			Character: lspRange.End.Character,
		},
	}
}

// convertLSPCodeActionContext converts LSP context to language service context
func convertLSPCodeActionContext(lspContext CodeActionContext) ls.CodeActionContext {
	var lsDiagnostics []ls.Diagnostic
	for _, d := range lspContext.Diagnostics {
		lsDiag := ls.Diagnostic{
			Range: ls.Range{
				Start: ls.Position{Line: d.Range.Start.Line, Character: d.Range.Start.Character},
				End:   ls.Position{Line: d.Range.End.Line, Character: d.Range.End.Character},
			},
			Message: d.Message,
		}
		if d.Severity != nil {
			severity := ls.DiagnosticSeverity(*d.Severity)
			lsDiag.Severity = &severity
		}
		lsDiagnostics = append(lsDiagnostics, lsDiag)
	}

	return ls.CodeActionContext{
		Diagnostics: lsDiagnostics,
	}
}

// convertToLSPCodeActions converts language service code actions to LSP code actions
func convertToLSPCodeActions(lsActions []ls.CodeAction) []CodeAction {
	var lspActions []CodeAction
	for _, action := range lsActions {
		lspAction := CodeAction{
			Title: action.Title,
			Kind:  action.Kind,
		}

		if action.Edit != nil {
			lspAction.Edit = &WorkspaceEdit{
				Changes: make(map[string][]TextEdit),
			}
			for uri, edits := range action.Edit.Changes {
				var lspEdits []TextEdit
				for _, edit := range edits {
					lspEdits = append(lspEdits, TextEdit{
						Range: Range{
							Start: Position{Line: edit.Range.Start.Line, Character: edit.Range.Start.Character},
							End:   Position{Line: edit.Range.End.Line, Character: edit.Range.End.Character},
						},
						NewText: edit.NewText,
					})
				}
				lspAction.Edit.Changes[uri] = lspEdits
			}
		}

		if action.Command != nil {
			lspAction.Command = &Command{
				Title:     action.Command.Title,
				Command:   action.Command.Command,
				Arguments: action.Command.Arguments,
			}
		}

		lspActions = append(lspActions, lspAction)
	}
	return lspActions
}

// convertLSPFormattingOptions converts LSP formatting options to language service options
func convertLSPFormattingOptions(lspOptions FormattingOptions) ls.FormattingOptions {
	return ls.FormattingOptions{
		TabSize:      lspOptions.TabSize,
		InsertSpaces: lspOptions.InsertSpaces,
	}
}

// convertToLSPTextEdits converts language service text edits to LSP text edits
func convertToLSPTextEdits(lsEdits []ls.TextEdit) []TextEdit {
	var lspEdits []TextEdit
	for _, edit := range lsEdits {
		lspEdits = append(lspEdits, TextEdit{
			Range: Range{
				Start: Position{Line: edit.Range.Start.Line, Character: edit.Range.Start.Character},
				End:   Position{Line: edit.Range.End.Line, Character: edit.Range.End.Character},
			},
			NewText: edit.NewText,
		})
	}
	return lspEdits
}
