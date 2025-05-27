// ABOUTME: Language service business logic for PSC type checker and analysis
// ABOUTME: Provides editor features using symbol tables and AST navigation, separated from LSP protocol

package ls

import (
	"os"
	"strings"
	"time"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/astnav"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
)

// LanguageService provides language analysis and editor features
type LanguageService struct {
	parser  parser.Parser
	binder  binder.Binder
	checker *parser.TypeCheck

	// Document cache
	documents map[string]*Document
}

// Document represents a document in the language service
type Document struct {
	URI         string
	Text        string
	Version     int
	AST         *ast.AST
	SymbolTable *binder.SymbolTable
	Errors      []parser.TypeCheckError
	LastChecked time.Time
	LastChanged time.Time
}

// Position represents a source position
type Position struct {
	Line      int
	Character int
}

// Range represents a source range
type Range struct {
	Start Position
	End   Position
}

// Location represents a source location
type Location struct {
	URI   string
	Range Range
}

// Definition represents a symbol definition
type Definition struct {
	Location Location
	Symbol   *binder.Symbol
}

// Hover represents hover information
type Hover struct {
	Contents string
	Range    *Range
}

// CompletionItem represents a completion suggestion
type CompletionItem struct {
	Label  string
	Kind   CompletionItemKind
	Detail string
}

// CompletionItemKind represents the kind of completion item
type CompletionItemKind int

const (
	CompletionItemKindVariable CompletionItemKind = iota
	CompletionItemKindFunction
	CompletionItemKindKeyword
	CompletionItemKindType
	CompletionItemKindMethod
	CompletionItemKindModule
)

// NewLanguageService creates a new language service
func NewLanguageService() (*LanguageService, error) {
	p, err := parser.NewParser()
	if err != nil {
		return nil, err
	}

	b := binder.NewBinder()

	checker, err := parser.NewTypeCheck()
	if err != nil {
		return nil, err
	}

	return &LanguageService{
		parser:    p,
		binder:    b,
		checker:   checker,
		documents: make(map[string]*Document),
	}, nil
}

// UpdateDocument updates or creates a document in the service
func (ls *LanguageService) UpdateDocument(uri, text string, version int) error {
	doc := &Document{
		URI:         uri,
		Text:        text,
		Version:     version,
		LastChanged: time.Now(),
	}

	// Parse the document
	ast, err := ls.parser.ParseString(text)
	if err != nil {
		// Store document even with parse errors
		doc.AST = ast
		ls.documents[uri] = doc
		return err
	}
	doc.AST = ast

	// Perform symbol binding
	symbolTable, err := ls.binder.BindAST(ast)
	if err != nil {
		// Store document even with binding errors
		doc.SymbolTable = symbolTable
		ls.documents[uri] = doc
		return err
	}
	doc.SymbolTable = symbolTable

	// Perform type checking
	// Create a temporary file for type checking
	tempPath, err := ls.writeToTempFile(uri, text)
	if err != nil {
		doc.SymbolTable = symbolTable
		ls.documents[uri] = doc
		return err
	}

	result, err := ls.checker.CheckFile(tempPath)
	if err != nil {
		doc.SymbolTable = symbolTable
		ls.documents[uri] = doc
		return err
	}

	if result != nil {
		doc.Errors = result.Errors
	}
	doc.LastChecked = time.Now()

	ls.documents[uri] = doc
	return nil
}

// GetDefinition finds the definition of a symbol at the given position
func (ls *LanguageService) GetDefinition(uri string, pos Position) (*Definition, error) {
	doc, exists := ls.documents[uri]
	if !exists {
		return nil, nil
	}

	// Find the symbol at the position
	symbol := ls.findSymbolAtPosition(doc, pos)
	if symbol == nil {
		return nil, nil
	}

	// Return the symbol's declaration location
	if symbol.Declaration != nil {
		location := Location{
			URI: uri, // For now, assume same file
			Range: Range{
				Start: Position{
					Line:      symbol.Position.Line - 1, // Convert to 0-based
					Character: symbol.Position.Column - 1,
				},
				End: Position{
					Line:      symbol.Position.Line - 1,
					Character: symbol.Position.Column - 1 + len(symbol.Name),
				},
			},
		}

		return &Definition{
			Location: location,
			Symbol:   symbol,
		}, nil
	}

	return nil, nil
}

// FindReferences finds all references to a symbol at the given position
func (ls *LanguageService) FindReferences(uri string, pos Position, includeDeclaration bool) ([]Location, error) {
	doc, exists := ls.documents[uri]
	if !exists {
		return nil, nil
	}

	// Find the symbol at the position
	symbol := ls.findSymbolAtPosition(doc, pos)
	if symbol == nil {
		return nil, nil
	}

	var locations []Location

	// Use AST navigation to find all references
	if doc.AST != nil && doc.AST.Root != nil {
		navigator := astnav.NewNavigator(doc.AST.Root)

		// Find all variable expressions that match this symbol
		variableNodes := navigator.FindNodesByType("variable_expression")
		for _, node := range variableNodes {
			if ls.nodeMatchesSymbol(node, symbol) {
				nodePos := node.Start()
				location := Location{
					URI: uri,
					Range: Range{
						Start: Position{
							Line:      nodePos.Line - 1,
							Character: nodePos.Column - 1,
						},
						End: Position{
							Line:      nodePos.Line - 1,
							Character: nodePos.Column - 1 + len(symbol.Name),
						},
					},
				}
				locations = append(locations, location)
			}
		}
	}

	// Include declaration if requested
	if includeDeclaration && symbol.Declaration != nil {
		location := Location{
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
		locations = append(locations, location)
	}

	return locations, nil
}

// GetHover provides hover information for a symbol at the given position
func (ls *LanguageService) GetHover(uri string, pos Position) (*Hover, error) {
	doc, exists := ls.documents[uri]
	if !exists {
		return nil, nil
	}

	// Find the symbol at the position
	symbol := ls.findSymbolAtPosition(doc, pos)
	if symbol == nil {
		return nil, nil
	}

	// Generate hover content based on symbol information
	var content strings.Builder

	// Add symbol kind and name
	content.WriteString("**")
	content.WriteString(ls.symbolKindToString(symbol.Kind))
	content.WriteString("**: `")
	content.WriteString(symbol.Name)
	content.WriteString("`\n\n")

	// Add type information if available
	if symbol.Type != "" {
		content.WriteString("**Type**: `")
		content.WriteString(symbol.Type)
		content.WriteString("`\n\n")
	}

	// Add scope information
	if symbol.Scope != nil {
		content.WriteString("**Scope**: ")
		content.WriteString(ls.scopeKindToString(symbol.Scope.Kind))
		content.WriteString("\n\n")
	}

	// Add package information
	if symbol.Package != "" {
		content.WriteString("**Package**: `")
		content.WriteString(symbol.Package)
		content.WriteString("`\n\n")
	}

	// Add flags information
	if symbol.Flags != binder.SymbolFlagNone {
		content.WriteString("**Flags**: ")
		content.WriteString(ls.symbolFlagsToString(symbol.Flags))
		content.WriteString("\n\n")
	}

	return &Hover{
		Contents: content.String(),
		Range: &Range{
			Start: pos,
			End:   Position{Line: pos.Line, Character: pos.Character + len(symbol.Name)},
		},
	}, nil
}

// GetCompletions provides completion suggestions at the given position
func (ls *LanguageService) GetCompletions(uri string, pos Position) ([]CompletionItem, error) {
	doc, exists := ls.documents[uri]
	if !exists {
		return nil, nil
	}

	var items []CompletionItem

	// Get all visible symbols from the symbol table
	if doc.SymbolTable != nil {
		visibleSymbols := ls.getVisibleSymbols(doc.SymbolTable, pos)

		for _, symbol := range visibleSymbols {
			item := CompletionItem{
				Label:  symbol.Name,
				Kind:   ls.symbolKindToCompletionKind(symbol.Kind),
				Detail: symbol.Type,
			}
			items = append(items, item)
		}
	}

	// Add keywords
	keywords := []string{"my", "our", "state", "sub", "method", "if", "elsif", "else", "while", "for", "foreach", "use", "package", "return"}
	for _, keyword := range keywords {
		items = append(items, CompletionItem{
			Label:  keyword,
			Kind:   CompletionItemKindKeyword,
			Detail: "Perl keyword",
		})
	}

	// Add types
	types := []string{"Str", "Int", "Num", "Bool", "ArrayRef", "HashRef", "CodeRef", "Any", "Undef", "Maybe"}
	for _, typeName := range types {
		items = append(items, CompletionItem{
			Label:  typeName,
			Kind:   CompletionItemKindType,
			Detail: "Type annotation",
		})
	}

	return items, nil
}

// GetDocumentSymbols returns all symbols in the document for outline view
func (ls *LanguageService) GetDocumentSymbols(uri string) ([]*binder.Symbol, error) {
	doc, exists := ls.documents[uri]
	if !exists {
		return nil, nil
	}

	if doc.SymbolTable == nil {
		return nil, nil
	}

	var symbols []*binder.Symbol

	// Collect all symbols from all scopes
	ls.collectSymbolsFromScope(doc.SymbolTable.GlobalScope, &symbols)

	return symbols, nil
}

// Helper methods

func (ls *LanguageService) findSymbolAtPosition(doc *Document, pos Position) *binder.Symbol {
	if doc.SymbolTable == nil {
		return nil
	}

	// Find the scope at the given position
	scope := ls.findScopeAtPosition(doc.SymbolTable, pos)
	if scope == nil {
		scope = doc.SymbolTable.GlobalScope
	}

	// Extract the word at the position
	word := ls.extractWordAtPosition(doc.Text, pos)
	if word == "" {
		return nil
	}

	// Strip sigil for lookup
	symbolName := word
	if len(word) > 0 && strings.ContainsRune("$@%*", rune(word[0])) {
		symbolName = word[1:]
	}

	// Search for the symbol in the scope chain
	return ls.resolveSymbolInScope(scope, symbolName)
}

func (ls *LanguageService) findScopeAtPosition(symbolTable *binder.SymbolTable, pos Position) *binder.Scope {
	// This is a simplified implementation
	// In a full implementation, you would traverse the AST to find the scope containing the position
	return symbolTable.CurrentScope
}

func (ls *LanguageService) extractWordAtPosition(text string, pos Position) string {
	lines := strings.Split(text, "\n")
	if pos.Line >= len(lines) {
		return ""
	}

	line := lines[pos.Line]
	if pos.Character >= len(line) {
		return ""
	}

	// Find word boundaries
	start := pos.Character
	end := pos.Character

	// Include sigil for variables
	if start > 0 && strings.ContainsRune("$@%*", rune(line[start-1])) {
		start--
	}

	// Move start backwards to beginning of word
	for start > 0 && ls.isWordChar(line[start-1]) {
		start--
	}

	// Include sigil if at the start
	if start > 0 && strings.ContainsRune("$@%*", rune(line[start-1])) {
		start--
	}

	// Move end forwards to end of word
	for end < len(line) && ls.isWordChar(line[end]) {
		end++
	}

	if start == end {
		return ""
	}

	return line[start:end]
}

func (ls *LanguageService) isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_' || c == ':'
}

func (ls *LanguageService) nodeMatchesSymbol(node ast.Node, symbol *binder.Symbol) bool {
	// Extract variable name from node and compare with symbol
	if node.Type() == "variable_expression" {
		text := node.Text()
		if len(text) > 0 && strings.ContainsRune("$@%*", rune(text[0])) {
			varName := text[1:]
			return varName == symbol.Name
		}
		return text == symbol.Name
	}
	return false
}

func (ls *LanguageService) resolveSymbolInScope(scope *binder.Scope, name string) *binder.Symbol {
	if scope == nil {
		return nil
	}

	// Check current scope
	if symbol, exists := scope.Symbols[name]; exists {
		return symbol
	}

	// Check parent scope
	return ls.resolveSymbolInScope(scope.Parent, name)
}

func (ls *LanguageService) getVisibleSymbols(symbolTable *binder.SymbolTable, pos Position) []*binder.Symbol {
	var symbols []*binder.Symbol

	// Find the scope at the position
	scope := ls.findScopeAtPosition(symbolTable, pos)
	if scope == nil {
		scope = symbolTable.GlobalScope
	}

	// Collect symbols from current scope and all parent scopes
	currentScope := scope
	for currentScope != nil {
		for _, symbol := range currentScope.Symbols {
			symbols = append(symbols, symbol)
		}
		currentScope = currentScope.Parent
	}

	return symbols
}

func (ls *LanguageService) collectSymbolsFromScope(scope *binder.Scope, symbols *[]*binder.Symbol) {
	if scope == nil {
		return
	}

	// Add symbols from this scope
	for _, symbol := range scope.Symbols {
		*symbols = append(*symbols, symbol)
	}

	// Recursively collect from child scopes
	for _, child := range scope.Children {
		ls.collectSymbolsFromScope(child, symbols)
	}
}

// Conversion helper methods

func (ls *LanguageService) symbolKindToString(kind binder.SymbolKind) string {
	switch kind {
	case binder.SymbolScalar:
		return "Scalar Variable"
	case binder.SymbolArray:
		return "Array Variable"
	case binder.SymbolHash:
		return "Hash Variable"
	case binder.SymbolGlob:
		return "Glob"
	case binder.SymbolSubroutine:
		return "Subroutine"
	case binder.SymbolMethod:
		return "Method"
	case binder.SymbolPackage:
		return "Package"
	case binder.SymbolImport:
		return "Import"
	case binder.SymbolType:
		return "Type"
	default:
		return "Symbol"
	}
}

func (ls *LanguageService) scopeKindToString(kind binder.ScopeKind) string {
	switch kind {
	case binder.ScopeGlobal:
		return "Global"
	case binder.ScopePackage:
		return "Package"
	case binder.ScopeSubroutine:
		return "Subroutine"
	case binder.ScopeMethod:
		return "Method"
	case binder.ScopeBlock:
		return "Block"
	case binder.ScopeEval:
		return "Eval"
	default:
		return "Unknown"
	}
}

func (ls *LanguageService) symbolFlagsToString(flags binder.SymbolFlags) string {
	var parts []string

	if flags&binder.SymbolFlagLexical != 0 {
		parts = append(parts, "lexical")
	}
	if flags&binder.SymbolFlagPackage != 0 {
		parts = append(parts, "package")
	}
	if flags&binder.SymbolFlagExported != 0 {
		parts = append(parts, "exported")
	}
	if flags&binder.SymbolFlagImported != 0 {
		parts = append(parts, "imported")
	}
	if flags&binder.SymbolFlagTypeAnnotated != 0 {
		parts = append(parts, "typed")
	}
	if flags&binder.SymbolFlagMethod != 0 {
		parts = append(parts, "method")
	}

	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

func (ls *LanguageService) symbolKindToCompletionKind(kind binder.SymbolKind) CompletionItemKind {
	switch kind {
	case binder.SymbolScalar, binder.SymbolArray, binder.SymbolHash, binder.SymbolGlob:
		return CompletionItemKindVariable
	case binder.SymbolSubroutine:
		return CompletionItemKindFunction
	case binder.SymbolMethod:
		return CompletionItemKindMethod
	case binder.SymbolPackage:
		return CompletionItemKindModule
	case binder.SymbolType:
		return CompletionItemKindType
	default:
		return CompletionItemKindVariable
	}
}

func (ls *LanguageService) writeToTempFile(uri, text string) (string, error) {
	// Create a temporary file with Perl extension
	tempFile, err := os.CreateTemp("", "lsp_*.pl")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Write content to the temporary file
	_, err = tempFile.WriteString(text)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}
