// ABOUTME: Enhanced LSP handlers that integrate the new type-aware features (EXPERIMENTAL)
// ABOUTME: Provides improved navigation, completion, and diagnostics handling

// EXPERIMENTAL FEATURE WARNING:
// These enhanced LSP handlers depend on experimental language service features
// and may not provide reliable results until the underlying type system is complete.
// TODO: Stabilize when enhanced LS features are moved out of experimental status

package lsp

import (
	"encoding/json"
	"fmt"

	"tamarou.com/pvm/internal/ls"
)

// handleEnhancedTextDocumentDefinition handles enhanced definition requests
func (s *Server) handleEnhancedTextDocumentDefinition(msg *JSONRPCMessage) error {
	var params DefinitionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Enhanced definition request for %s at %d:%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	// Convert LSP position to language service position
	lsPos := convertLSPPosition(params.Position)

	// Get enhanced definition from language service
	definition, err := s.languageService.GetEnhancedDefinition(params.TextDocument.URI, lsPos)
	if err != nil {
		s.logger.Printf("Failed to get enhanced definition: %v", err)
		return s.sendResponse(msg.ID, nil)
	}

	if definition == nil {
		return s.sendResponse(msg.ID, nil)
	}

	// Create enhanced location response with additional metadata
	response := map[string]interface{}{
		"uri":   definition.Location.URI,
		"range": convertToLSPRange(definition.Location.Range),
		"data": map[string]interface{}{
			"typeInfo":   definition.TypeInfo,
			"docComment": definition.DocComment,
			"signature":  definition.Signature,
			"isExported": definition.IsExported,
		},
	}

	return s.sendResponse(msg.ID, response)
}

// handleEnhancedTextDocumentCompletion handles enhanced completion requests
func (s *Server) handleEnhancedTextDocumentCompletion(msg *JSONRPCMessage) error {
	// Start request-scoped pooling
	requestID := fmt.Sprintf("enhanced_completion_%v", msg.ID)
	_ = s.poolManager.StartRequest(requestID, "textDocument/completion")
	defer s.poolManager.EndRequest(requestID)

	var params CompletionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Enhanced completion request for %s at %d:%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	// Convert LSP position to language service position
	lsPos := convertLSPPosition(params.Position)

	// Get enhanced completions from language service
	completions, err := s.languageService.GetEnhancedCompletions(params.TextDocument.URI, lsPos)
	if err != nil {
		s.logger.Printf("Failed to get enhanced completions: %v", err)
		result := s.poolManager.NewCompletionList(requestID, false)
		return s.sendResponse(msg.ID, result)
	}

	// Create pooled completion list and items
	result := s.poolManager.NewCompletionList(requestID, false)

	// Convert enhanced completions to LSP completions
	for _, comp := range completions {
		item := s.poolManager.NewCompletionItem(requestID, comp.Label, comp.Detail)

		// Set completion item kind
		kind := convertCompletionKind(comp.Kind)
		item.Kind = &kind

		// Add documentation
		if comp.Documentation != "" {
			item.Documentation = &MarkupContent{
				Kind:  MarkupKindMarkdown,
				Value: comp.Documentation,
			}
		}

		// Add insert text with snippet
		if comp.PostfixSnippet != "" {
			item.InsertText = comp.PostfixSnippet
			format := InsertTextFormatSnippet
			item.InsertTextFormat = &format
		}

		// Add sort text based on score
		sortText := fmt.Sprintf("%03d_%s", 1000-comp.Score, comp.Label)
		item.SortText = sortText

		// Add additional text edits for imports if needed
		if comp.RequiredImport != "" {
			item.AdditionalTextEdits = []TextEdit{
				{
					Range: Range{
						Start: Position{Line: 0, Character: 0},
						End:   Position{Line: 0, Character: 0},
					},
					NewText: fmt.Sprintf("use %s;\n", comp.RequiredImport),
				},
			}
		}

		result.Items = append(result.Items, *item)
	}

	return s.sendResponse(msg.ID, result)
}

// handleEnhancedTextDocumentPublishDiagnostics publishes enhanced diagnostics
func (s *Server) handleEnhancedTextDocumentPublishDiagnostics(uri string) error {
	// Get enhanced diagnostics from language service
	enhancedDiags, err := s.languageService.GetEnhancedDiagnostics(uri)
	if err != nil {
		s.logger.Printf("Failed to get enhanced diagnostics: %v", err)
		return nil
	}

	// Convert enhanced diagnostics to LSP diagnostics
	var diagnostics []Diagnostic
	for _, enhDiag := range enhancedDiags {
		// Convert ls.Diagnostic to LSP Diagnostic
		diag := Diagnostic{
			Range:    convertToLSPRange(enhDiag.Diagnostic.Range),
			Message:  enhDiag.Diagnostic.Message,
			Severity: (*DiagnosticSeverity)(enhDiag.Diagnostic.Severity),
		}

		// Add code and source
		diag.Code = enhDiag.Data
		diag.Source = "psc-typecheck"

		// Add related information
		if len(enhDiag.RelatedInfo) > 0 {
			diag.RelatedInformation = make([]DiagnosticRelatedInfo, len(enhDiag.RelatedInfo))
			for i, related := range enhDiag.RelatedInfo {
				diag.RelatedInformation[i] = DiagnosticRelatedInfo{
					Location: convertToLSPLocation(related.Location),
					Message:  related.Message,
				}
			}
		}

		// Add code description
		if enhDiag.CodeDescription != nil {
			diag.CodeDescription = &CodeDescription{
				Href: enhDiag.CodeDescription.Href,
			}
		}

		// Store enhanced data for code actions
		diag.Data = map[string]interface{}{
			"suggestions": enhDiag.Suggestions,
			"typeContext": enhDiag.TypeContext,
		}

		diagnostics = append(diagnostics, diag)
	}

	// Publish diagnostics
	params := PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	}

	return s.sendNotification("textDocument/publishDiagnostics", params)
}

// handleEnhancedTextDocumentCodeAction handles enhanced code action requests
func (s *Server) handleEnhancedTextDocumentCodeAction(msg *JSONRPCMessage) error {
	var params CodeActionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Enhanced code action request for %s at %d:%d-%d:%d",
		params.TextDocument.URI,
		params.Range.Start.Line, params.Range.Start.Character,
		params.Range.End.Line, params.Range.End.Character)

	var actions []CodeAction

	// Process diagnostics with enhanced suggestions
	for _, diag := range params.Context.Diagnostics {
		// Extract enhanced data if available
		if data, ok := diag.Data.(map[string]interface{}); ok {
			if suggestions, ok := data["suggestions"].([]ls.DiagnosticSuggestion); ok {
				for _, suggestion := range suggestions {
					action := CodeAction{
						Title: suggestion.Title,
						Kind:  CodeActionKindQuickFix,
					}

					// Create workspace edit from suggestion
					if suggestion.Edit != nil {
						action.Edit = &WorkspaceEdit{
							Changes: map[string][]TextEdit{
								params.TextDocument.URI: {convertToLSPTextEdit(*suggestion.Edit)},
							},
						}
					}

					actions = append(actions, action)
				}
			}
		}
	}

	// Add refactoring actions using language service
	lsRange := ls.Range{
		Start: convertLSPPosition(params.Range.Start),
		End:   convertLSPPosition(params.Range.End),
	}

	lsContext := ls.CodeActionContext{
		Diagnostics: convertToLSDiagnostics(params.Context.Diagnostics),
	}

	lsActions, err := s.languageService.GenerateCodeActions(params.TextDocument.URI, lsRange, lsContext)
	if err == nil {
		for _, lsAction := range lsActions {
			action := convertToLSPCodeAction(lsAction)
			actions = append(actions, action)
		}
	}

	return s.sendResponse(msg.ID, actions)
}

// handleCallHierarchyPrepare prepares call hierarchy
func (s *Server) handleCallHierarchyPrepare(msg *JSONRPCMessage) error {
	var params CallHierarchyPrepareParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Call hierarchy prepare for %s at %d:%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	// Convert position
	lsPos := convertLSPPosition(params.Position)

	// Get call hierarchy item
	item, err := s.languageService.GetCallHierarchy(params.TextDocument.URI, lsPos)
	if err != nil || item == nil {
		return s.sendResponse(msg.ID, nil)
	}

	// Convert to LSP call hierarchy item
	lspItem := CallHierarchyItem{
		Name:           item.Name,
		Kind:           12, // SymbolKindFunction = 12
		URI:            item.URI,
		Range:          convertToLSPRange(item.Range),
		SelectionRange: convertToLSPRange(item.Range),
		Detail:         &item.Detail,
	}

	return s.sendResponse(msg.ID, []CallHierarchyItem{lspItem})
}

// handleCallHierarchyIncomingCalls handles incoming calls request
func (s *Server) handleCallHierarchyIncomingCalls(msg *JSONRPCMessage) error {
	var params CallHierarchyIncomingCallsParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	// Convert LSP item back to language service item
	lsItem := &ls.CallHierarchyItem{
		Name:  params.Item.Name,
		Kind:  "function", // Convert from symbol kind
		URI:   params.Item.URI,
		Range: convertFromLSPRange(params.Item.Range),
	}

	// Get incoming calls
	calls, err := s.languageService.GetIncomingCalls(lsItem)
	if err != nil {
		return s.sendResponse(msg.ID, []CallHierarchyIncomingCall{})
	}

	// Convert to LSP format
	var lspCalls []CallHierarchyIncomingCall
	for _, call := range calls {
		lspCall := CallHierarchyIncomingCall{
			From: CallHierarchyItem{
				Name:           call.From.Name,
				Kind:           12, // SymbolKindFunction = 12
				URI:            call.From.URI,
				Range:          convertToLSPRange(call.From.Range),
				SelectionRange: convertToLSPRange(call.From.Range),
			},
			FromRanges: convertToLSPRanges(call.FromRanges),
		}
		lspCalls = append(lspCalls, lspCall)
	}

	return s.sendResponse(msg.ID, lspCalls)
}

// handleCallHierarchyOutgoingCalls handles outgoing calls request
func (s *Server) handleCallHierarchyOutgoingCalls(msg *JSONRPCMessage) error {
	var params CallHierarchyOutgoingCallsParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	// Convert LSP item back to language service item
	lsItem := &ls.CallHierarchyItem{
		Name:  params.Item.Name,
		Kind:  "function",
		URI:   params.Item.URI,
		Range: convertFromLSPRange(params.Item.Range),
	}

	// Get outgoing calls
	calls, err := s.languageService.GetOutgoingCalls(lsItem)
	if err != nil {
		return s.sendResponse(msg.ID, []CallHierarchyOutgoingCall{})
	}

	// Convert to LSP format
	var lspCalls []CallHierarchyOutgoingCall
	for _, call := range calls {
		lspCall := CallHierarchyOutgoingCall{
			To: CallHierarchyItem{
				Name:           call.To.Name,
				Kind:           12, // SymbolKindFunction = 12
				URI:            call.To.URI,
				Range:          convertToLSPRange(call.To.Range),
				SelectionRange: convertToLSPRange(call.To.Range),
			},
			FromRanges: convertToLSPRanges(call.FromRanges),
		}
		lspCalls = append(lspCalls, lspCall)
	}

	return s.sendResponse(msg.ID, lspCalls)
}

// Helper conversion functions

func convertToLSPRange(lsRange ls.Range) Range {
	return Range{
		Start: Position{
			Line:      lsRange.Start.Line,
			Character: lsRange.Start.Character,
		},
		End: Position{
			Line:      lsRange.End.Line,
			Character: lsRange.End.Character,
		},
	}
}

func convertFromLSPRange(lspRange Range) ls.Range {
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

func convertToLSPRanges(lsRanges []ls.Range) []Range {
	ranges := make([]Range, len(lsRanges))
	for i, r := range lsRanges {
		ranges[i] = convertToLSPRange(r)
	}
	return ranges
}

func convertToLSPTextEdit(lsEdit ls.TextEdit) TextEdit {
	return TextEdit{
		Range:   convertToLSPRange(lsEdit.Range),
		NewText: lsEdit.NewText,
	}
}

func convertToLSDiagnostics(lspDiags []Diagnostic) []ls.Diagnostic {
	var diags []ls.Diagnostic
	for _, lspDiag := range lspDiags {
		diag := ls.Diagnostic{
			Range:   convertFromLSPRange(lspDiag.Range),
			Message: lspDiag.Message,
		}
		if lspDiag.Severity != nil {
			severity := ls.DiagnosticSeverity(*lspDiag.Severity)
			diag.Severity = &severity
		}
		diags = append(diags, diag)
	}
	return diags
}

func convertToLSPCodeAction(lsAction ls.CodeAction) CodeAction {
	action := CodeAction{
		Title: lsAction.Title,
		Kind:  lsAction.Kind,
	}

	if lsAction.Edit != nil {
		action.Edit = &WorkspaceEdit{
			Changes: make(map[string][]TextEdit),
		}
		for uri, edits := range lsAction.Edit.Changes {
			lspEdits := make([]TextEdit, len(edits))
			for i, edit := range edits {
				lspEdits[i] = convertToLSPTextEdit(edit)
			}
			action.Edit.Changes[uri] = lspEdits
		}
	}

	if lsAction.Command != nil {
		action.Command = &Command{
			Title:     lsAction.Command.Title,
			Command:   lsAction.Command.Command,
			Arguments: lsAction.Command.Arguments,
		}
	}

	return action
}

func convertCompletionKind(kind ls.CompletionItemKind) CompletionItemKind {
	switch kind {
	case ls.CompletionItemKindFunction:
		return CompletionItemKindFunction
	case ls.CompletionItemKindVariable:
		return CompletionItemKindVariable
	case ls.CompletionItemKindMethod:
		return CompletionItemKindMethod
	case ls.CompletionItemKindKeyword:
		return CompletionItemKindKeyword
	case ls.CompletionItemKindType:
		return CompletionItemKindClass
	case ls.CompletionItemKindModule:
		return CompletionItemKindModule
	case ls.CompletionItemKindSnippet:
		return CompletionItemKindSnippet
	default:
		return CompletionItemKindText
	}
}

// Additional LSP protocol types for enhanced features

// CallHierarchyPrepareParams for call hierarchy prepare request
type CallHierarchyPrepareParams struct {
	TextDocumentPositionParams
}

// CallHierarchyItem represents an item in call hierarchy
type CallHierarchyItem struct {
	Name           string      `json:"name"`
	Kind           int         `json:"kind"` // Using int for symbol kind
	Tags           []int       `json:"tags,omitempty"`
	Detail         *string     `json:"detail,omitempty"`
	URI            string      `json:"uri"`
	Range          Range       `json:"range"`
	SelectionRange Range       `json:"selectionRange"`
	Data           interface{} `json:"data,omitempty"`
}

// CallHierarchyIncomingCallsParams for incoming calls request
type CallHierarchyIncomingCallsParams struct {
	Item CallHierarchyItem `json:"item"`
}

// CallHierarchyIncomingCall represents an incoming call
type CallHierarchyIncomingCall struct {
	From       CallHierarchyItem `json:"from"`
	FromRanges []Range           `json:"fromRanges"`
}

// CallHierarchyOutgoingCallsParams for outgoing calls request
type CallHierarchyOutgoingCallsParams struct {
	Item CallHierarchyItem `json:"item"`
}

// CallHierarchyOutgoingCall represents an outgoing call
type CallHierarchyOutgoingCall struct {
	To         CallHierarchyItem `json:"to"`
	FromRanges []Range           `json:"fromRanges"`
}

// DiagnosticRelatedInformation for related diagnostic info
type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

// CodeActionKind constants
const (
	CodeActionKindEmpty                 = ""
	CodeActionKindQuickFix              = "quickfix"
	CodeActionKindRefactor              = "refactor"
	CodeActionKindRefactorExtract       = "refactor.extract"
	CodeActionKindRefactorInline        = "refactor.inline"
	CodeActionKindRefactorRewrite       = "refactor.rewrite"
	CodeActionKindSource                = "source"
	CodeActionKindSourceOrganizeImports = "source.organizeImports"
)
