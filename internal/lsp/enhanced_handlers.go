// ABOUTME: Enhanced LSP handlers for advanced IDE features
// ABOUTME: Provides workspace symbols, inlay hints, semantic highlighting, and auto-fixes

package lsp

import (
	"encoding/json"
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
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
	// Try to find a function symbol at or near the position
	symbol, err := s.languageService.FindSymbolAtPosition(uri, pos)
	if err != nil || symbol == nil {
		// Fallback to empty signature help if no symbol found
		return &SignatureHelp{
			Signatures:      []SignatureInformation{},
			ActiveSignature: nil,
			ActiveParameter: nil,
		}, nil
	}

	// Only provide signature help for functions/methods
	if symbol.Kind != binder.SymbolSubroutine && symbol.Kind != binder.SymbolMethod {
		return &SignatureHelp{
			Signatures:      []SignatureInformation{},
			ActiveSignature: nil,
			ActiveParameter: nil,
		}, nil
	}

	// Extract parameters from the function's AST declaration
	parameters := s.extractParametersFromDeclaration(symbol.Declaration)

	// Build the function signature label
	label := fmt.Sprintf("sub %s", symbol.Name)
	if len(parameters) > 0 {
		var paramLabels []string
		for _, param := range parameters {
			if param.Type != "" && param.Type != "Any" {
				paramLabels = append(paramLabels, fmt.Sprintf("%s %s", param.Type, param.Name))
			} else {
				paramLabels = append(paramLabels, param.Name)
			}
		}
		label += fmt.Sprintf("(%s)", strings.Join(paramLabels, ", "))
	} else {
		label += "()"
	}

	// Add return type if available
	if symbol.Type != "" {
		label += fmt.Sprintf(" -> %s", symbol.Type)
	}

	// Create parameter information for LSP
	var paramInfo []ParameterInformation
	for _, param := range parameters {
		info := ParameterInformation{
			Label: param.Name,
		}
		if param.Type != "" && param.Type != "Any" {
			info.Documentation = &MarkupContent{
				Kind:  MarkupKindMarkdown,
				Value: fmt.Sprintf("**Type**: %s", param.Type),
			}
		}
		paramInfo = append(paramInfo, info)
	}

	signatureInfo := &SignatureInformation{
		Label: label,
		Documentation: &MarkupContent{
			Kind:  MarkupKindMarkdown,
			Value: fmt.Sprintf("Function signature for `%s`", symbol.Name),
		},
		Parameters: paramInfo,
	}

	signatureHelp := &SignatureHelp{
		Signatures:      []SignatureInformation{*signatureInfo},
		ActiveSignature: &[]int{0}[0],
		ActiveParameter: &[]int{0}[0],
	}

	return signatureHelp, nil
}

// extractParametersFromDeclaration extracts parameter information from AST declaration
func (s *Server) extractParametersFromDeclaration(decl ast.Node) []ParameterInfo {
	if subDecl, ok := decl.(*ast.SubDecl); ok {
		params := subDecl.Parameters()
		if len(params) == 0 {
			return nil
		}

		result := make([]ParameterInfo, len(params))
		for i, param := range params {
			result[i] = ParameterInfo{
				Name: param.Name,
				Type: s.extractTypeString(param.TypeExpr),
			}
		}
		return result
	}

	return nil
}

// extractTypeString converts a type expression to a string representation
func (s *Server) extractTypeString(typeExpr *ast.TypeExpression) string {
	if typeExpr == nil {
		return "Any"
	}

	// Handle different type expression kinds
	switch typeExpr.Kind {
	case ast.SimpleTypeKind:
		return typeExpr.Name
	case ast.UnionTypeKind:
		// For union types like Int|Str, join with |
		if len(typeExpr.UnionTypes) > 0 {
			var types []string
			for _, t := range typeExpr.UnionTypes {
				types = append(types, s.extractTypeString(t))
			}
			return strings.Join(types, "|")
		}
		return typeExpr.Name
	case ast.ParameterizedTypeKind:
		// For parameterized types like ArrayRef[Int]
		if len(typeExpr.Parameters) > 0 {
			var params []string
			for _, p := range typeExpr.Parameters {
				params = append(params, s.extractTypeString(p))
			}
			return fmt.Sprintf("%s[%s]", typeExpr.Name, strings.Join(params, ", "))
		}
		return typeExpr.Name
	case ast.IntersectionTypeKind:
		// For intersection types like Object&Serializable
		if len(typeExpr.IntersectionTypes) > 0 {
			var types []string
			for _, t := range typeExpr.IntersectionTypes {
				types = append(types, s.extractTypeString(t))
			}
			return strings.Join(types, "&")
		}
		return typeExpr.Name
	case ast.NegationTypeKind:
		// For negation types like !Undef
		if typeExpr.NegatedType != nil {
			return "!" + s.extractTypeString(typeExpr.NegatedType)
		}
		return "!" + typeExpr.Name
	default:
		return typeExpr.Name
	}
}

// handleInlayHint handles textDocument/inlayHint requests
func (s *Server) handleInlayHint(msg *JSONRPCMessage) error {
	var params InlayHintParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Inlay hint request for %s at %d:%d-%d:%d",
		params.TextDocument.URI,
		params.Range.Start.Line, params.Range.Start.Character,
		params.Range.End.Line, params.Range.End.Character)

	// Get inlay hints from language service
	inlayHints, err := s.generateInlayHints(params.TextDocument.URI, params.Range)
	if err != nil {
		s.logger.Printf("Failed to generate inlay hints: %v", err)
		return s.sendResponse(msg.ID, []InlayHint{})
	}

	return s.sendResponse(msg.ID, inlayHints)
}

// generateInlayHints generates inlay hints for a document range
func (s *Server) generateInlayHints(uri string, docRange Range) ([]InlayHint, error) {
	var hints []InlayHint

	// Get document symbols to find variables and functions
	symbols, err := s.languageService.GetDocumentSymbols(uri)
	if err != nil {
		return hints, err
	}

	// Generate type hints for variables
	for _, symbol := range symbols {
		hint := s.createInlayHintFromSymbol(symbol, docRange)
		if hint != nil {
			hints = append(hints, *hint)
		}
	}

	return hints, nil
}

// createInlayHintFromSymbol creates an inlay hint from a symbol
func (s *Server) createInlayHintFromSymbol(symbol *binder.Symbol, docRange Range) *InlayHint {
	if symbol.Declaration == nil {
		return nil
	}

	pos := symbol.Declaration.Start()

	// Check if symbol position is within the requested range
	if !s.positionInRange(pos, docRange) {
		return nil
	}

	// Create type hint for variables with inferred types
	if s.isVariableSymbol(symbol) {
		return s.createTypeInlayHint(symbol, pos)
	}

	// Create parameter hints for functions
	if symbol.Kind == binder.SymbolSubroutine && symbol.Type != "" {
		return s.createParameterInlayHint(symbol, pos)
	}

	return nil
}

// createTypeInlayHint creates a type inlay hint for a variable
func (s *Server) createTypeInlayHint(symbol *binder.Symbol, pos ast.Position) *InlayHint {
	var typeLabel string

	if symbol.Type != "" {
		typeLabel = fmt.Sprintf(": %s", symbol.Type)
	} else {
		// Infer simple type from symbol kind
		switch symbol.Kind {
		case binder.SymbolScalar:
			typeLabel = ": Scalar"
		case binder.SymbolArray:
			typeLabel = ": Array"
		case binder.SymbolHash:
			typeLabel = ": Hash"
		default:
			return nil // No type info available
		}
	}

	return &InlayHint{
		Position: Position{
			Line:      pos.Line - 1,
			Character: pos.Column - 1 + len(symbol.Name),
		},
		Label:        typeLabel,
		Kind:         InlayHintKindType,
		PaddingLeft:  false,
		PaddingRight: true,
		Tooltip: &MarkupContent{
			Kind:  MarkupKindMarkdown,
			Value: fmt.Sprintf("**Type**: %s", strings.TrimPrefix(typeLabel, ": ")),
		},
	}
}

// createParameterInlayHint creates a parameter inlay hint for a function
func (s *Server) createParameterInlayHint(symbol *binder.Symbol, pos ast.Position) *InlayHint {
	if symbol.Type == "" {
		return nil
	}

	return &InlayHint{
		Position: Position{
			Line:      pos.Line - 1,
			Character: pos.Column - 1,
		},
		Label:        fmt.Sprintf("-> %s", symbol.Type),
		Kind:         InlayHintKindType,
		PaddingLeft:  true,
		PaddingRight: false,
		Tooltip: &MarkupContent{
			Kind:  MarkupKindMarkdown,
			Value: fmt.Sprintf("**Return Type**: %s", symbol.Type),
		},
	}
}

// isVariableSymbol checks if a symbol is a variable
func (s *Server) isVariableSymbol(symbol *binder.Symbol) bool {
	return symbol.Kind == binder.SymbolScalar ||
		symbol.Kind == binder.SymbolArray ||
		symbol.Kind == binder.SymbolHash ||
		symbol.Kind == binder.SymbolGlob
}

// positionInRange checks if a position is within a range
func (s *Server) positionInRange(pos ast.Position, docRange Range) bool {
	line := pos.Line - 1   // Convert to 0-based
	char := pos.Column - 1 // Convert to 0-based

	// Check if position is within range
	if line < docRange.Start.Line || line > docRange.End.Line {
		return false
	}

	if line == docRange.Start.Line && char < docRange.Start.Character {
		return false
	}

	if line == docRange.End.Line && char > docRange.End.Character {
		return false
	}

	return true
}

// handleSemanticTokensFull handles textDocument/semanticTokens/full requests
func (s *Server) handleSemanticTokensFull(msg *JSONRPCMessage) error {
	var params SemanticTokensParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Printf("Semantic tokens request for %s", params.TextDocument.URI)

	// Get semantic tokens from language service
	semanticTokens, err := s.generateSemanticTokens(params.TextDocument.URI)
	if err != nil {
		s.logger.Printf("Failed to generate semantic tokens: %v", err)
		return s.sendResponse(msg.ID, SemanticTokens{Data: []uint32{}})
	}

	return s.sendResponse(msg.ID, semanticTokens)
}

// generateSemanticTokens generates semantic tokens for a document
func (s *Server) generateSemanticTokens(uri string) (*SemanticTokens, error) {
	// Get document symbols for token generation
	symbols, err := s.languageService.GetDocumentSymbols(uri)
	if err != nil {
		return nil, err
	}

	// Create semantic tokens builder
	builder := &SemanticTokensBuilder{
		tokenTypes:     getSemanticTokenTypes(),
		tokenModifiers: getSemanticTokenModifiers(),
		data:           []uint32{},
	}

	// Generate tokens from symbols
	for _, symbol := range symbols {
		builder.addTokenFromSymbol(symbol)
	}

	return &SemanticTokens{
		Data: builder.data,
	}, nil
}

// SemanticTokensBuilder helps build semantic token data
type SemanticTokensBuilder struct {
	tokenTypes     []string
	tokenModifiers []string
	data           []uint32
	lastLine       int
	lastChar       int
}

// addTokenFromSymbol adds a semantic token from a symbol
func (builder *SemanticTokensBuilder) addTokenFromSymbol(symbol *binder.Symbol) {
	if symbol.Declaration == nil {
		return
	}

	pos := symbol.Declaration.Start()
	tokenType := builder.getTokenTypeFromSymbol(symbol)
	tokenTypeIndex := builder.findTokenTypeIndex(tokenType)

	if tokenTypeIndex == -1 {
		return // Unknown token type
	}

	length := len(symbol.Name)
	modifiers := builder.getTokenModifiersFromSymbol(symbol)

	// Convert to relative position format required by LSP
	deltaLine := pos.Line - 1 - builder.lastLine
	var deltaChar int
	if deltaLine == 0 {
		deltaChar = pos.Column - 1 - builder.lastChar
	} else {
		deltaChar = pos.Column - 1
	}

	// Add token: [deltaLine, deltaChar, length, tokenType, tokenModifiers]
	builder.data = append(builder.data,
		uint32(deltaLine),
		uint32(deltaChar),
		uint32(length),
		uint32(tokenTypeIndex),
		uint32(modifiers),
	)

	// Update position tracking
	builder.lastLine = pos.Line - 1
	builder.lastChar = pos.Column - 1
}

// getTokenTypeFromSymbol maps symbol kind to semantic token type
func (builder *SemanticTokensBuilder) getTokenTypeFromSymbol(symbol *binder.Symbol) string {
	switch symbol.Kind {
	case binder.SymbolPackage:
		return SemanticTokenTypeNamespace
	case binder.SymbolType:
		return SemanticTokenTypeType
	case binder.SymbolSubroutine:
		return SemanticTokenTypeFunction
	case binder.SymbolMethod:
		return SemanticTokenTypeMethod
	case binder.SymbolScalar, binder.SymbolArray, binder.SymbolHash, binder.SymbolGlob:
		return SemanticTokenTypeVariable
	default:
		return SemanticTokenTypeVariable
	}
}

// getTokenModifiersFromSymbol determines token modifiers for a symbol
func (builder *SemanticTokensBuilder) getTokenModifiersFromSymbol(symbol *binder.Symbol) uint32 {
	var modifiers uint32

	// Check if it's a declaration
	if symbol.Declaration != nil {
		modifiers |= 1 << builder.findTokenModifierIndex(SemanticTokenModifierDeclaration)
	}

	// Check if it's exported (package symbol)
	if symbol.Flags&binder.SymbolFlagExported != 0 {
		modifiers |= 1 << builder.findTokenModifierIndex(SemanticTokenModifierStatic)
	}

	return modifiers
}

// findTokenTypeIndex finds the index of a token type
func (builder *SemanticTokensBuilder) findTokenTypeIndex(tokenType string) int {
	for i, t := range builder.tokenTypes {
		if t == tokenType {
			return i
		}
	}
	return -1
}

// findTokenModifierIndex finds the index of a token modifier
func (builder *SemanticTokensBuilder) findTokenModifierIndex(modifier string) int {
	for i, m := range builder.tokenModifiers {
		if m == modifier {
			return i
		}
	}
	return -1
}

// getSemanticTokenTypes returns the list of supported semantic token types
func getSemanticTokenTypes() []string {
	return []string{
		SemanticTokenTypeNamespace,
		SemanticTokenTypeType,
		SemanticTokenTypeClass,
		SemanticTokenTypeFunction,
		SemanticTokenTypeMethod,
		SemanticTokenTypeVariable,
		SemanticTokenTypeProperty,
		SemanticTokenTypeParameter,
		SemanticTokenTypeKeyword,
		SemanticTokenTypeModifier,
		SemanticTokenTypeComment,
		SemanticTokenTypeString,
		SemanticTokenTypeNumber,
		SemanticTokenTypeOperator,
	}
}

// getSemanticTokenModifiers returns the list of supported semantic token modifiers
func getSemanticTokenModifiers() []string {
	return []string{
		SemanticTokenModifierDeclaration,
		SemanticTokenModifierDefinition,
		SemanticTokenModifierReadonly,
		SemanticTokenModifierStatic,
		SemanticTokenModifierDeprecated,
		SemanticTokenModifierModification,
		SemanticTokenModifierDocumentation,
	}
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
