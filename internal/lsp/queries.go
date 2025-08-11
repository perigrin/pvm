// ABOUTME: Type information query functionality for LSP
// ABOUTME: Provides detailed type information and symbol analysis

package lsp

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/ls"
)

// TypeQuery represents a query for type information
type TypeQuery struct {
	URI      string   `json:"uri"`
	Position Position `json:"position"`
	Symbol   string   `json:"symbol,omitempty"`
}

// TypeInfo represents detailed type information
type TypeInfo struct {
	Symbol        string         `json:"symbol"`
	Type          string         `json:"type"`
	Kind          string         `json:"kind"` // variable, function, class, etc.
	Documentation string         `json:"documentation,omitempty"`
	Location      *Location      `json:"location,omitempty"`
	Signature     *FunctionSig   `json:"signature,omitempty"`
	Properties    []PropertyInfo `json:"properties,omitempty"`
	Methods       []MethodInfo   `json:"methods,omitempty"`
	Examples      []string       `json:"examples,omitempty"`
}

// FunctionSig represents a function signature
type FunctionSig struct {
	Parameters []ParameterInfo `json:"parameters"`
	ReturnType string          `json:"returnType,omitempty"`
}

// ParameterInfo represents function parameter information
type ParameterInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// PropertyInfo represents object property information
type PropertyInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// MethodInfo represents method information
type MethodInfo struct {
	Name      string       `json:"name"`
	Signature *FunctionSig `json:"signature,omitempty"`
}

// TypeQueryService provides type query functionality
type TypeQueryService struct {
	server *Server
}

// NewTypeQueryService creates a new type query service
func NewTypeQueryService(server *Server) *TypeQueryService {
	return &TypeQueryService{
		server: server,
	}
}

// QueryTypeAtPosition queries type information at a specific position
func (s *TypeQueryService) QueryTypeAtPosition(query TypeQuery) (*TypeInfo, error) {
	// Convert LSP position to language service position
	lsPos := convertLSPPositionForQuery(query.Position)

	// Find symbol at position
	symbol, err := s.server.languageService.FindSymbolAtPosition(query.URI, lsPos)
	if err != nil {
		return nil, fmt.Errorf("failed to find symbol: %w", err)
	}
	if symbol == nil {
		return nil, nil // No symbol found, not an error
	}

	// Get hover information for additional context
	hover, err := s.server.languageService.GetHover(query.URI, lsPos)
	if err != nil {
		return nil, fmt.Errorf("failed to get hover info: %w", err)
	}

	// Get definition location
	definition, err := s.server.languageService.GetDefinition(query.URI, lsPos)
	if err != nil {
		return nil, fmt.Errorf("failed to get definition: %w", err)
	}

	// Build type information
	typeInfo := &TypeInfo{
		Symbol:        symbol.Name,
		Type:          symbol.Type,
		Kind:          s.symbolKindToString(symbol.Kind),
		Documentation: "",
	}

	// Add hover content as documentation if available
	if hover != nil {
		typeInfo.Documentation = hover.Contents
	}

	// Add location information if available
	if definition != nil {
		typeInfo.Location = &Location{
			URI: definition.Location.URI,
			Range: Range{
				Start: Position{
					Line:      definition.Location.Range.Start.Line,
					Character: definition.Location.Range.Start.Character,
				},
				End: Position{
					Line:      definition.Location.Range.End.Line,
					Character: definition.Location.Range.End.Character,
				},
			},
		}
	}

	// Add signature information for functions/methods
	if symbol.Kind == binder.SymbolSubroutine || symbol.Kind == binder.SymbolMethod {
		parameters := extractParametersFromAST(symbol.Declaration)
		typeInfo.Signature = &FunctionSig{
			Parameters: parameters,
			ReturnType: symbol.Type,
		}
	}

	// Add examples for common types
	typeInfo.Examples = s.generateExamplesForSymbol(symbol)

	return typeInfo, nil
}

// QuerySymbol queries type information for a specific symbol
func (s *TypeQueryService) QuerySymbol(uri, symbolName string) (*TypeInfo, error) {
	// Get document from language service
	doc, exists := s.server.languageService.GetDocumentForDebug(uri)
	if !exists {
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	if doc.SymbolTable == nil {
		return nil, fmt.Errorf("no symbol table available for document: %s", uri)
	}

	// Search for the symbol in the symbol table
	symbol := s.findSymbolByName(doc.SymbolTable, symbolName)
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	// Build type information
	typeInfo := &TypeInfo{
		Symbol: symbol.Name,
		Type:   symbol.Type,
		Kind:   s.symbolKindToString(symbol.Kind),
	}

	// Add location information from symbol declaration
	if symbol.Declaration != nil {
		declPos := symbol.Declaration.Start()
		typeInfo.Location = &Location{
			URI: uri,
			Range: Range{
				Start: Position{
					Line:      declPos.Line - 1, // Convert to 0-based
					Character: declPos.Column - 1,
				},
				End: Position{
					Line:      declPos.Line - 1,
					Character: declPos.Column - 1 + len(symbol.Name),
				},
			},
		}
	} else if symbol.Position.Line > 0 {
		// Fallback to position from symbol table
		typeInfo.Location = &Location{
			URI: uri,
			Range: Range{
				Start: Position{
					Line:      symbol.Position.Line - 1,
					Character: symbol.Position.Column - 1,
				},
				End: Position{
					Line:      symbol.Position.Line - 1,
					Character: symbol.Position.Column - 1 + len(symbol.Name),
				},
			},
		}
	}

	// Add signature information for functions/methods
	if symbol.Kind == binder.SymbolSubroutine || symbol.Kind == binder.SymbolMethod {
		parameters := extractParametersFromAST(symbol.Declaration)
		typeInfo.Signature = &FunctionSig{
			Parameters: parameters,
			ReturnType: symbol.Type,
		}
	}

	// Add documentation from hover if available
	if typeInfo.Location != nil {
		pos := ls.Position{
			Line:      typeInfo.Location.Range.Start.Line,
			Character: typeInfo.Location.Range.Start.Character,
		}
		hover, err := s.server.languageService.GetHover(uri, pos)
		if err == nil && hover != nil {
			typeInfo.Documentation = hover.Contents
		}
	}

	// Add examples for the symbol
	typeInfo.Examples = s.generateExamplesForSymbol(symbol)

	return typeInfo, nil
}

// GetAvailableSymbols returns all available symbols in a document
func (s *TypeQueryService) GetAvailableSymbols(uri string) ([]TypeInfo, error) {
	// Get document symbols from language service
	symbols, err := s.server.languageService.GetDocumentSymbols(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to get document symbols: %w", err)
	}

	if symbols == nil {
		return []TypeInfo{}, nil
	}

	var typeInfos []TypeInfo
	for _, symbol := range symbols {
		typeInfo := TypeInfo{
			Symbol: symbol.Name,
			Type:   symbol.Type,
			Kind:   s.symbolKindToString(symbol.Kind),
		}

		// Add location information
		if symbol.Declaration != nil {
			declPos := symbol.Declaration.Start()
			typeInfo.Location = &Location{
				URI: uri,
				Range: Range{
					Start: Position{
						Line:      declPos.Line - 1,
						Character: declPos.Column - 1,
					},
					End: Position{
						Line:      declPos.Line - 1,
						Character: declPos.Column - 1 + len(symbol.Name),
					},
				},
			}
		} else if symbol.Position.Line > 0 {
			typeInfo.Location = &Location{
				URI: uri,
				Range: Range{
					Start: Position{
						Line:      symbol.Position.Line - 1,
						Character: symbol.Position.Column - 1,
					},
					End: Position{
						Line:      symbol.Position.Line - 1,
						Character: symbol.Position.Column - 1 + len(symbol.Name),
					},
				},
			}
		}

		// Add signature for functions/methods
		if symbol.Kind == binder.SymbolSubroutine || symbol.Kind == binder.SymbolMethod {
			parameters := extractParametersFromAST(symbol.Declaration)
			typeInfo.Signature = &FunctionSig{
				Parameters: parameters,
				ReturnType: symbol.Type,
			}
		}

		// Add examples
		typeInfo.Examples = s.generateExamplesForSymbol(symbol)

		typeInfos = append(typeInfos, typeInfo)
	}

	return typeInfos, nil
}

// Helper functions

// convertLSPPositionForQuery converts LSP position to language service position
func convertLSPPositionForQuery(pos Position) ls.Position {
	return ls.Position{
		Line:      pos.Line,
		Character: pos.Character,
	}
}

// findSymbolByName searches for a symbol by name in the symbol table
func (s *TypeQueryService) findSymbolByName(symbolTable *binder.SymbolTable, name string) *binder.Symbol {
	if symbolTable == nil || symbolTable.GlobalScope == nil {
		return nil
	}

	return s.findSymbolInScope(symbolTable.GlobalScope, name)
}

// findSymbolInScope recursively searches for a symbol in scopes
func (s *TypeQueryService) findSymbolInScope(scope *binder.Scope, name string) *binder.Symbol {
	if scope == nil {
		return nil
	}

	// Check current scope
	if symbol, exists := scope.Symbols[name]; exists {
		return symbol
	}

	// Check child scopes
	for _, child := range scope.Children {
		if symbol := s.findSymbolInScope(child, name); symbol != nil {
			return symbol
		}
	}

	return nil
}

// symbolKindToString converts symbol kind to string representation
func (s *TypeQueryService) symbolKindToString(kind binder.SymbolKind) string {
	switch kind {
	case binder.SymbolScalar:
		return "scalar"
	case binder.SymbolArray:
		return "array"
	case binder.SymbolHash:
		return "hash"
	case binder.SymbolGlob:
		return "glob"
	case binder.SymbolSubroutine:
		return "function"
	case binder.SymbolMethod:
		return "method"
	case binder.SymbolPackage:
		return "package"
	case binder.SymbolImport:
		return "import"
	case binder.SymbolType:
		return "type"
	default:
		return "unknown"
	}
}

// generateExamplesForSymbol generates usage examples for a symbol
func (s *TypeQueryService) generateExamplesForSymbol(symbol *binder.Symbol) []string {
	var examples []string

	switch symbol.Kind {
	case binder.SymbolScalar:
		examples = append(examples, fmt.Sprintf("my $%s = \"value\";", symbol.Name))
		if symbol.Type != "" && symbol.Type != "Scalar" {
			examples = append(examples, fmt.Sprintf("my %s $%s = ...;", symbol.Type, symbol.Name))
		}
	case binder.SymbolArray:
		examples = append(examples, fmt.Sprintf("my @%s = (1, 2, 3);", symbol.Name))
		examples = append(examples, fmt.Sprintf("push @%s, $item;", symbol.Name))
	case binder.SymbolHash:
		examples = append(examples, fmt.Sprintf("my %%%s = (key => 'value');", symbol.Name))
		examples = append(examples, fmt.Sprintf("$%s{key} = 'value';", symbol.Name))
	case binder.SymbolSubroutine, binder.SymbolMethod:
		if symbol.Type != "" {
			examples = append(examples, fmt.Sprintf("%s(); # returns %s", symbol.Name, symbol.Type))
		} else {
			examples = append(examples, fmt.Sprintf("%s();", symbol.Name))
		}
	case binder.SymbolPackage:
		examples = append(examples, fmt.Sprintf("use %s;", symbol.Name))
		examples = append(examples, fmt.Sprintf("%s::function();", symbol.Name))
	case binder.SymbolType:
		examples = append(examples, fmt.Sprintf("my %s $var;", symbol.Name))
	}

	return examples
}

// extractParametersFromAST extracts function parameters from AST node
func extractParametersFromAST(decl ast.Node) []ParameterInfo {
	// Check if this is a SubDecl (subroutine declaration)
	if subDecl, ok := decl.(*ast.SubDecl); ok {
		params := subDecl.Parameters()
		if len(params) == 0 {
			return nil
		}

		result := make([]ParameterInfo, len(params))
		for i, param := range params {
			result[i] = ParameterInfo{
				Name: param.Name,
				Type: extractTypeString(param.TypeExpr),
			}
		}
		return result
	}

	return nil
}

// extractTypeString converts a type expression to a string representation
func extractTypeString(typeExpr *ast.TypeExpression) string {
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
				types = append(types, extractTypeString(t))
			}
			return strings.Join(types, "|")
		}
		return typeExpr.Name
	case ast.ParameterizedTypeKind:
		// For parameterized types like ArrayRef[Int]
		if len(typeExpr.Parameters) > 0 {
			var params []string
			for _, p := range typeExpr.Parameters {
				params = append(params, extractTypeString(p))
			}
			return fmt.Sprintf("%s[%s]", typeExpr.Name, strings.Join(params, ", "))
		}
		return typeExpr.Name
	case ast.IntersectionTypeKind:
		// For intersection types like Object&Serializable
		if len(typeExpr.IntersectionTypes) > 0 {
			var types []string
			for _, t := range typeExpr.IntersectionTypes {
				types = append(types, extractTypeString(t))
			}
			return strings.Join(types, "&")
		}
		return typeExpr.Name
	case ast.NegationTypeKind:
		// For negation types like !Undef
		if typeExpr.NegatedType != nil {
			return "!" + extractTypeString(typeExpr.NegatedType)
		}
		return "!" + typeExpr.Name
	default:
		return typeExpr.Name
	}
}
