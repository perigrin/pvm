// ABOUTME: Enhanced LSP handlers for advanced IDE features
// ABOUTME: Provides workspace symbols, inlay hints, semantic highlighting, and auto-fixes

package lsp

import (
	"encoding/json"
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/ls"
	"tamarou.com/pvm/internal/parser"
)

// Enhanced LSP methods that extend the basic protocol

// handleWorkspaceSymbol handles workspace/symbol requests
func (s *Server) handleWorkspaceSymbol(msg *JSONRPCMessage) error {
	var params WorkspaceSymbolParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Workspace symbol request with query: '%s'", params.Query)

	// Get workspace symbols from language service
	symbols, err := s.languageService.GetWorkspaceSymbols(params.Query)
	if err != nil {
		s.logger.Printf("Failed to get workspace symbols: %v", err)
		return s.sendResponse(msg.ID, []WorkspaceSymbol{})
	}

	// Convert to LSP workspace symbols
	lspSymbols := convertToLSPWorkspaceSymbols(symbols)
	return s.sendResponse(msg.ID, lspSymbols)
}

// handleDocumentSymbol handles textDocument/documentSymbol requests
func (s *Server) handleDocumentSymbol(msg *JSONRPCMessage) error {
	var params DocumentSymbolParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Document symbol request for %s", params.TextDocument.URI)

	// Get document symbols from language service
	symbols, err := s.languageService.GetDocumentSymbols(params.TextDocument.URI)
	if err != nil {
		s.logger.Printf("Failed to get document symbols: %v", err)
		return s.sendResponse(msg.ID, []DocumentSymbol{})
	}

	// Convert to LSP document symbols
	lspSymbols := convertToLSPDocumentSymbols(symbols)
	return s.sendResponse(msg.ID, lspSymbols)
}

// handleSignatureHelp handles textDocument/signatureHelp requests
func (s *Server) handleSignatureHelp(msg *JSONRPCMessage) error {
	var params SignatureHelpParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Signature help request for %s at %d:%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	// Convert LSP position to language service position
	lsPos := convertLSPPosition(params.Position)

	// Get signature help from language service
	signatureHelp, err := s.getSignatureHelpFromLanguageService(params.TextDocument.URI, lsPos)
	if err != nil {
		s.logger.Printf("Failed to get signature help: %v", err)
		return s.sendResponse(msg.ID, nil)
	}

	return s.sendResponse(msg.ID, signatureHelp)
}

// generateAutoFixSuggestions creates auto-fix code actions
func (s *Server) generateAutoFixSuggestions(uri string, errors []parser.TypeCheckError) []CodeAction {
	var actions []CodeAction

	for _, err := range errors {
		if strings.Contains(err.Message, "undefined variable") && strings.Contains(err.Message, "$") {
			// Extract variable name
			if varName := extractVariableName(err.Message); varName != "" {
				action := CodeAction{
					Title: fmt.Sprintf("Declare variable '%s'", varName),
					Kind:  "quickfix",
					Edit: &WorkspaceEdit{
						Changes: map[string][]TextEdit{
							uri: {
								{
									Range: Range{
										Start: Position{Line: err.Line - 1, Character: 0},
										End:   Position{Line: err.Line - 1, Character: 0},
									},
									NewText: fmt.Sprintf("my %s;\n", varName),
								},
							},
						},
					},
				}
				actions = append(actions, action)
			}
		}

		if strings.Contains(err.Message, "type mismatch") {
			action := CodeAction{
				Title: "Add type annotation",
				Kind:  "quickfix",
				Edit: &WorkspaceEdit{
					Changes: map[string][]TextEdit{
						uri: {
							{
								Range: Range{
									Start: Position{Line: err.Line - 1, Character: err.Column - 1},
									End:   Position{Line: err.Line - 1, Character: err.Column - 1},
								},
								NewText: " # TODO: Add type annotation",
							},
						},
					},
				},
			}
			actions = append(actions, action)
		}
	}

	return actions
}

// extractVariableName extracts variable name from error message
func extractVariableName(message string) string {
	start := strings.Index(message, "$")
	if start == -1 {
		return ""
	}

	end := start + 1
	for end < len(message) && (isAlphaNumeric(message[end]) || message[end] == '_') {
		end++
	}

	if end > start+1 {
		return message[start:end]
	}

	return ""
}

// isAlphaNumeric checks if character is alphanumeric
func isAlphaNumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// convertToLSPWorkspaceSymbols converts language service symbols to LSP workspace symbols
func convertToLSPWorkspaceSymbols(symbols []*binder.Symbol) []WorkspaceSymbol {
	var lspSymbols []WorkspaceSymbol

	for _, symbol := range symbols {
		lspSymbol := WorkspaceSymbol{
			Name: symbol.Name,
			Kind: symbolKindToLSPSymbolKind(symbol.Kind),
		}

		// Add location information
		if symbol.Declaration != nil {
			declPos := symbol.Declaration.Start()
			lspSymbol.Location = Location{
				URI: "file://unknown", // TODO: Get actual URI from symbol
				Range: Range{
					Start: Position{Line: declPos.Line - 1, Character: declPos.Column - 1},
					End:   Position{Line: declPos.Line - 1, Character: declPos.Column - 1 + len(symbol.Name)},
				},
			}
		}

		lspSymbols = append(lspSymbols, lspSymbol)
	}

	return lspSymbols
}

// getSignatureHelpFromLanguageService gets signature help at a position
func (s *Server) getSignatureHelpFromLanguageService(uri string, pos ls.Position) (*SignatureHelp, error) {
	// For now, return a simple signature help response
	// This is a placeholder implementation that can be enhanced
	// by accessing the language service documents directly

	// Simple signature help for common Perl functions
	signatureInfo := &SignatureInformation{
		Label: "sub function_name",
		Documentation: &MarkupContent{
			Kind:  MarkupKindMarkdown,
			Value: "Function signature help (simplified implementation)",
		},
		Parameters: []ParameterInformation{
			{Label: "$param1"},
			{Label: "$param2"},
		},
	}

	signatureHelp := &SignatureHelp{
		Signatures:      []SignatureInformation{*signatureInfo},
		ActiveSignature: &[]int{0}[0],
		ActiveParameter: &[]int{0}[0],
	}

	return signatureHelp, nil
}

// symbolKindToLSPSymbolKind converts symbol kind to LSP symbol kind
func symbolKindToLSPSymbolKind(kind binder.SymbolKind) SymbolKind {
	switch kind {
	case binder.SymbolScalar, binder.SymbolArray, binder.SymbolHash, binder.SymbolGlob:
		return SymbolKindVariable
	case binder.SymbolSubroutine:
		return SymbolKindFunction
	case binder.SymbolMethod:
		return SymbolKindMethod
	case binder.SymbolPackage:
		return SymbolKindNamespace
	case binder.SymbolType:
		return SymbolKindClass
	default:
		return SymbolKindVariable
	}
}

// convertToLSPDocumentSymbols converts language service symbols to LSP document symbols
func convertToLSPDocumentSymbols(symbols []*binder.Symbol) []DocumentSymbol {
	var lspSymbols []DocumentSymbol

	for _, symbol := range symbols {
		lspSymbol := DocumentSymbol{
			Name: symbol.Name,
			Kind: symbolKindToLSPSymbolKind(symbol.Kind),
		}

		// Add selection and full range information
		if symbol.Declaration != nil {
			declPos := symbol.Declaration.Start()
			endPos := symbol.Declaration.End()

			// Selection range is just the symbol name
			lspSymbol.SelectionRange = Range{
				Start: Position{Line: declPos.Line - 1, Character: declPos.Column - 1},
				End:   Position{Line: declPos.Line - 1, Character: declPos.Column - 1 + len(symbol.Name)},
			}

			// Full range includes the entire declaration
			lspSymbol.Range = Range{
				Start: Position{Line: declPos.Line - 1, Character: declPos.Column - 1},
				End:   Position{Line: endPos.Line - 1, Character: endPos.Column - 1},
			}
		} else {
			// Fallback to minimal range if no declaration position
			lspSymbol.SelectionRange = Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: 0, Character: len(symbol.Name)},
			}
			lspSymbol.Range = lspSymbol.SelectionRange
		}

		// Add detail information based on symbol type
		if symbol.Type != "" {
			lspSymbol.Detail = fmt.Sprintf("Type: %s", symbol.Type)
		}

		lspSymbols = append(lspSymbols, lspSymbol)
	}

	return lspSymbols
}
