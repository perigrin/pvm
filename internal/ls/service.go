// ABOUTME: Language service business logic for PSC type checker and analysis
// ABOUTME: Provides editor features using symbol tables and AST navigation, separated from LSP protocol

package ls

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
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

	// Performance optimizations
	cache   *DocumentCache
	monitor *PerformanceMonitor

	mu sync.RWMutex
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

	// Performance tracking
	ContentHash string
	ASTHash     string
	SymbolHash  string
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
		cache:     NewDocumentCache(),
		monitor:   NewPerformanceMonitor(),
	}, nil
}

// UpdateDocument updates or creates a document in the service
func (ls *LanguageService) UpdateDocument(uri, text string, version int) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	contentHash := ls.cache.HashContent(text)

	// Check if we have a cached version that's still valid
	if existingDoc, exists := ls.documents[uri]; exists {
		if existingDoc.ContentHash == contentHash && existingDoc.Version == version {
			// No changes, return existing document
			return nil
		}
	}

	// Invalidate caches for this document
	ls.cache.InvalidateDocument(uri)

	doc := &Document{
		URI:         uri,
		Text:        text,
		Version:     version,
		LastChanged: time.Now(),
		ContentHash: contentHash,
	}

	// Parse with performance monitoring
	parseOp := ls.monitor.StartOperation(context.Background(), "parse")
	ast := ls.cache.GetAST(uri, contentHash)
	if ast == nil {
		ls.monitor.RecordCacheMiss()
		var err error
		parseStart := time.Now()
		ast, err = ls.parser.ParseString(text)
		parseDuration := time.Since(parseStart)

		if err != nil {
			parseOp.CompleteWithError(err)
			doc.AST = ast
			ls.documents[uri] = doc
			return err
		}

		// Cache the AST
		ls.cache.SetAST(uri, contentHash, ast, parseDuration, []string{})
	} else {
		ls.monitor.RecordCacheHit()
	}
	parseOp.Complete()
	doc.AST = ast
	doc.ASTHash = ls.cache.HashContent(fmt.Sprintf("%p", ast)) // Simple AST hash

	// Symbol binding with performance monitoring
	bindOp := ls.monitor.StartOperation(context.Background(), "bind")
	symbolTable := ls.cache.GetSymbolTable(uri, contentHash, doc.ASTHash)
	if symbolTable == nil {
		ls.monitor.RecordCacheMiss()
		var err error
		bindStart := time.Now()
		symbolTable, err = ls.binder.BindAST(ast)
		bindDuration := time.Since(bindStart)

		if err != nil {
			bindOp.CompleteWithError(err)
			doc.SymbolTable = symbolTable
			ls.documents[uri] = doc
			return err
		}

		// Cache the symbol table
		ls.cache.SetSymbolTable(uri, contentHash, doc.ASTHash, symbolTable, bindDuration)
	} else {
		ls.monitor.RecordCacheHit()
	}
	bindOp.Complete()
	doc.SymbolTable = symbolTable
	doc.SymbolHash = ls.cache.HashContent(fmt.Sprintf("%p", symbolTable)) // Simple symbol table hash

	// Type checking with performance monitoring
	typeCheckOp := ls.monitor.StartOperation(context.Background(), "typecheck")
	errors := ls.cache.GetTypeCheckResult(uri, contentHash, doc.SymbolHash)
	if errors == nil {
		ls.monitor.RecordCacheMiss()
		tempPath, err := ls.writeToTempFile(uri, text)
		if err != nil {
			typeCheckOp.CompleteWithError(err)
			doc.SymbolTable = symbolTable
			ls.documents[uri] = doc
			return err
		}

		typeCheckStart := time.Now()
		result, err := ls.checker.CheckFile(tempPath)
		typeCheckDuration := time.Since(typeCheckStart)

		if err != nil {
			typeCheckOp.CompleteWithError(err)
			doc.SymbolTable = symbolTable
			ls.documents[uri] = doc
			return err
		}

		if result != nil {
			errors = result.Errors
		} else {
			errors = []parser.TypeCheckError{}
		}

		// Cache the type check results
		ls.cache.SetTypeCheckResult(uri, contentHash, doc.SymbolHash, errors, typeCheckDuration)
	} else {
		ls.monitor.RecordCacheHit()
	}
	typeCheckOp.Complete()
	doc.Errors = errors
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
		// Use the declaration's actual position for accurate navigation
		declPos := symbol.Declaration.Start()
		location := Location{
			URI: uri, // For now, assume same file
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

		return &Definition{
			Location: location,
			Symbol:   symbol,
		}, nil
	}

	// If no declaration node, fall back to position from symbol table
	if symbol.Position.Line > 0 {
		location := Location{
			URI: uri,
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

		// Find all nodes that reference this symbol
		allNodes := []ast.Node{}
		navigator.Walk(doc.AST.Root, func(node ast.Node) bool {
			// Check various node types that might reference symbols
			switch node.Type() {
			case "variable_expression", "subroutine_call", "method_call", "identifier":
				if ls.nodeMatchesSymbol(node, symbol) {
					allNodes = append(allNodes, node)
				}
			}
			return true
		})

		// Convert nodes to locations
		for _, node := range allNodes {
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

	// Include declaration if requested
	if includeDeclaration && symbol.Declaration != nil {
		declPos := symbol.Declaration.Start()
		location := Location{
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
		locations = append(locations, location)
	}

	return locations, nil
}

// GetHover provides hover information for a symbol at the given position
func (ls *LanguageService) GetHover(uri string, pos Position) (*Hover, error) {
	hoverOp := ls.monitor.StartOperation(context.Background(), "hover")
	defer hoverOp.Complete()

	ls.mu.RLock()
	doc, exists := ls.documents[uri]
	ls.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("%s:%d:%d", uri, pos.Line, pos.Character)

	// Check cache first
	if cached := ls.cache.GetHover(cacheKey); cached != nil {
		ls.monitor.RecordCacheHit()
		return cached, nil
	}
	ls.monitor.RecordCacheMiss()

	// Find the symbol at the position
	symbol := ls.findSymbolAtPosition(doc, pos)
	if symbol == nil {
		return nil, nil
	}

	// Generate hover content based on symbol information
	var content strings.Builder

	// Add symbol kind and name with sigil for variables
	content.WriteString("**")
	content.WriteString(ls.symbolKindToString(symbol.Kind))
	content.WriteString("**: `")
	if ls.isVariableSymbol(symbol.Kind) {
		content.WriteString(ls.getSigilForSymbol(symbol.Kind))
	}
	content.WriteString(symbol.Name)
	content.WriteString("`\n\n")

	// Add type information if available
	if symbol.Type != "" {
		content.WriteString("**Type**: `")
		content.WriteString(symbol.Type)
		content.WriteString("`\n\n")
	}

	// Add declaration location if available
	if symbol.Declaration != nil {
		declPos := symbol.Declaration.Start()
		content.WriteString("**Declared at**: Line ")
		content.WriteString(fmt.Sprintf("%d", declPos.Line))
		content.WriteString(", Column ")
		content.WriteString(fmt.Sprintf("%d", declPos.Column))
		content.WriteString("\n\n")
	}

	// Add scope information with additional context
	if symbol.Scope != nil {
		content.WriteString("**Scope**: ")
		content.WriteString(ls.scopeKindToString(symbol.Scope.Kind))
		if symbol.Scope.Parent != nil {
			content.WriteString(" (nested in ")
			content.WriteString(ls.scopeKindToString(symbol.Scope.Parent.Kind))
			content.WriteString(")")
		}
		content.WriteString("\n\n")
	}

	// Add package information
	if symbol.Package != "" {
		content.WriteString("**Package**: `")
		content.WriteString(symbol.Package)
		content.WriteString("`\n\n")
	}

	// Add flags information with descriptions
	if symbol.Flags != binder.SymbolFlagNone {
		content.WriteString("**Properties**: ")
		content.WriteString(ls.symbolFlagsToString(symbol.Flags))
		content.WriteString("\n\n")
	}

	// Add additional context for specific symbol types
	switch symbol.Kind {
	case binder.SymbolSubroutine, binder.SymbolMethod:
		content.WriteString("**Type**: Function/Method")
		// Add more function-specific info if available from the AST
		if symbol.Declaration != nil {
			content.WriteString("\n\n**Declaration**: Available in AST")
		}
		content.WriteString("\n\n")
	}

	// Remove trailing newlines
	finalContent := strings.TrimSuffix(content.String(), "\n\n")

	hover := &Hover{
		Contents: finalContent,
		Range: &Range{
			Start: pos,
			End:   Position{Line: pos.Line, Character: pos.Character + len(symbol.Name)},
		},
	}

	// Cache the result
	ls.cache.SetHover(cacheKey, doc.ContentHash, hover)

	return hover, nil
}

// GetCompletions provides completion suggestions at the given position
func (ls *LanguageService) GetCompletions(uri string, pos Position) ([]CompletionItem, error) {
	completionOp := ls.monitor.StartOperation(context.Background(), "completion")
	defer completionOp.Complete()

	ls.mu.RLock()
	doc, exists := ls.documents[uri]
	ls.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	// Generate cache key based on position and context
	cacheKey := fmt.Sprintf("%s:%d:%d:completion", uri, pos.Line, pos.Character)

	// Check cache first
	if cached := ls.cache.GetCompletions(cacheKey); cached != nil {
		ls.monitor.RecordCacheHit()
		return cached, nil
	}
	ls.monitor.RecordCacheMiss()

	// Use memory pool for completion items
	itemsPtr := ls.cache.GetCompletionItems(32)
	defer ls.cache.PutCompletionItems(itemsPtr)
	items := *itemsPtr

	// Get context information for better completions
	context := ls.getCompletionContext(doc, pos)

	// Get all visible symbols from the symbol table
	if doc.SymbolTable != nil {
		visibleSymbols := ls.getVisibleSymbols(doc.SymbolTable, pos)

		for _, symbol := range visibleSymbols {
			// Filter by context when appropriate
			if ls.shouldIncludeSymbolInCompletion(symbol, context) {
				item := CompletionItem{
					Label:  ls.formatSymbolForCompletion(symbol),
					Kind:   ls.symbolKindToCompletionKind(symbol.Kind),
					Detail: ls.formatSymbolDetail(symbol),
				}
				items = append(items, item)
			}
		}
	}

	// Add context-appropriate keywords
	keywords := ls.getContextualKeywords(context)
	for _, keyword := range keywords {
		items = append(items, CompletionItem{
			Label:  keyword,
			Kind:   CompletionItemKindKeyword,
			Detail: "Perl keyword",
		})
	}

	// Add types when in type context
	if context.ExpectedType || context.InTypeAnnotation {
		types := []string{"Str", "Int", "Num", "Bool", "ArrayRef", "HashRef", "CodeRef", "Any", "Undef", "Maybe"}
		for _, typeName := range types {
			items = append(items, CompletionItem{
				Label:  typeName,
				Kind:   CompletionItemKindType,
				Detail: "Type annotation",
			})
		}
	}

	// Add built-in functions when appropriate
	if context.ExpectedFunction {
		builtins := []string{"print", "say", "defined", "ref", "substr", "length", "chomp", "split", "join", "grep", "map", "sort", "push", "pop", "shift", "unshift"}
		for _, builtin := range builtins {
			items = append(items, CompletionItem{
				Label:  builtin,
				Kind:   CompletionItemKindFunction,
				Detail: "Perl builtin function",
			})
		}
	}

	// Make a copy for caching (since we're returning the pooled slice)
	result := make([]CompletionItem, len(items))
	copy(result, items)

	// Cache the result
	ls.cache.SetCompletions(cacheKey, doc.ContentHash, result)

	return result, nil
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

// TextEdit represents a text edit for rename operations
type TextEdit struct {
	Range   Range
	NewText string
}

// RenameSymbol provides rename functionality for symbols
func (ls *LanguageService) RenameSymbol(uri string, pos Position, newName string) ([]TextEdit, error) {
	doc, exists := ls.documents[uri]
	if !exists {
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	// Find the symbol at the position
	symbol := ls.findSymbolAtPosition(doc, pos)
	if symbol == nil {
		return nil, fmt.Errorf("no symbol found at position")
	}

	// Check if the new name is valid
	if !ls.isValidSymbolName(newName) {
		return nil, fmt.Errorf("invalid symbol name: %s", newName)
	}

	var edits []TextEdit

	// Find all references to this symbol
	locations, err := ls.FindReferences(uri, pos, true) // Include declaration
	if err != nil {
		return nil, fmt.Errorf("failed to find references: %w", err)
	}

	// Create text edits for all references
	for _, location := range locations {
		// Only handle same-file renames for now
		if location.URI == uri {
			edit := TextEdit{
				Range:   location.Range,
				NewText: newName,
			}
			edits = append(edits, edit)
		}
	}

	return edits, nil
}

// GetWorkspaceSymbols searches for symbols across all documents in the workspace
func (ls *LanguageService) GetWorkspaceSymbols(query string) ([]*binder.Symbol, error) {
	var allSymbols []*binder.Symbol

	// Search through all documents
	for _, doc := range ls.documents {
		if doc.SymbolTable == nil {
			continue
		}

		var docSymbols []*binder.Symbol
		ls.collectSymbolsFromScope(doc.SymbolTable.GlobalScope, &docSymbols)

		// Filter symbols by query
		for _, symbol := range docSymbols {
			if query == "" || strings.Contains(strings.ToLower(symbol.Name), strings.ToLower(query)) {
				allSymbols = append(allSymbols, symbol)
			}
		}
	}

	return allSymbols, nil
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

// Helper methods for enhanced hover information

func (ls *LanguageService) isVariableSymbol(kind binder.SymbolKind) bool {
	switch kind {
	case binder.SymbolScalar, binder.SymbolArray, binder.SymbolHash, binder.SymbolGlob:
		return true
	default:
		return false
	}
}

func (ls *LanguageService) getSigilForSymbol(kind binder.SymbolKind) string {
	switch kind {
	case binder.SymbolScalar:
		return "$"
	case binder.SymbolArray:
		return "@"
	case binder.SymbolHash:
		return "%"
	case binder.SymbolGlob:
		return "*"
	default:
		return ""
	}
}

func (ls *LanguageService) isValidSymbolName(name string) bool {
	if name == "" {
		return false
	}

	// Check if first character is valid (letter or underscore)
	first := name[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Check remaining characters (letters, digits, underscores, colons)
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == ':') {
			return false
		}
	}

	return true
}

// CompletionContext represents the context for code completion
type CompletionContext struct {
	ExpectedType     bool   // True if we're in a type annotation context
	InTypeAnnotation bool   // True if cursor is in a type annotation
	ExpectedFunction bool   // True if we expect a function/method call
	TriggerCharacter string // Character that triggered completion
	PreviousWord     string // Word before cursor
}

// getCompletionContext analyzes the text around the cursor to determine completion context
func (ls *LanguageService) getCompletionContext(doc *Document, pos Position) *CompletionContext {
	lines := strings.Split(doc.Text, "\n")
	if pos.Line >= len(lines) {
		return &CompletionContext{}
	}

	line := lines[pos.Line]
	prefix := ""
	if pos.Character <= len(line) {
		prefix = line[:pos.Character]
	}

	context := &CompletionContext{}

	// Check for type annotation context
	if strings.Contains(prefix, "my ") && !strings.Contains(prefix, ";") {
		// Likely in variable declaration with potential type
		context.ExpectedType = true
	}

	// Check for explicit type annotation
	if strings.Contains(prefix, " ") {
		words := strings.Fields(prefix)
		if len(words) >= 2 && (words[len(words)-2] == "my" || words[len(words)-2] == "our" || words[len(words)-2] == "state") {
			context.InTypeAnnotation = true
		}
	}

	// Check for function call context
	if strings.HasSuffix(prefix, "(") || strings.HasSuffix(prefix, " ") {
		context.ExpectedFunction = true
	}

	// Extract previous word
	words := strings.Fields(prefix)
	if len(words) > 0 {
		context.PreviousWord = words[len(words)-1]
	}

	return context
}

// shouldIncludeSymbolInCompletion determines if a symbol should be included based on context
func (ls *LanguageService) shouldIncludeSymbolInCompletion(symbol *binder.Symbol, context *CompletionContext) bool {
	// If expecting a function, prefer functions and methods
	if context.ExpectedFunction {
		switch symbol.Kind {
		case binder.SymbolSubroutine, binder.SymbolMethod:
			return true
		default:
			// Still include variables but with lower priority
			return true
		}
	}

	// If in type annotation, exclude variables
	if context.InTypeAnnotation {
		return symbol.Kind == binder.SymbolType
	}

	// Default: include all symbols
	return true
}

// formatSymbolForCompletion formats a symbol name for completion display
func (ls *LanguageService) formatSymbolForCompletion(symbol *binder.Symbol) string {
	if ls.isVariableSymbol(symbol.Kind) {
		return ls.getSigilForSymbol(symbol.Kind) + symbol.Name
	}
	return symbol.Name
}

// formatSymbolDetail formats detailed information about a symbol for completion
func (ls *LanguageService) formatSymbolDetail(symbol *binder.Symbol) string {
	var detail strings.Builder

	if symbol.Type != "" {
		detail.WriteString(symbol.Type)
	} else {
		detail.WriteString(ls.symbolKindToString(symbol.Kind))
	}

	if symbol.Package != "" && symbol.Package != "main" {
		detail.WriteString(" (")
		detail.WriteString(symbol.Package)
		detail.WriteString(")")
	}

	return detail.String()
}

// getContextualKeywords returns keywords appropriate for the current context
func (ls *LanguageService) getContextualKeywords(context *CompletionContext) []string {
	baseKeywords := []string{"my", "our", "state", "sub", "method", "if", "elsif", "else", "while", "for", "foreach", "use", "package", "return"}

	// Add control flow keywords based on context
	if context.ExpectedFunction {
		return append(baseKeywords, "last", "next", "redo", "goto")
	}

	return baseKeywords
}

// GetCacheStats returns cache performance statistics
func (ls *LanguageService) GetCacheStats() CacheStats {
	return ls.cache.GetStats()
}

// GetPerformanceStats returns performance monitoring statistics
func (ls *LanguageService) GetPerformanceStats() PerformanceStats {
	return ls.monitor.GetStats()
}

// ClearCache clears all cached data
func (ls *LanguageService) ClearCache() {
	ls.cache.Clear()
}

// ResetPerformanceMetrics resets all performance counters
func (ls *LanguageService) ResetPerformanceMetrics() {
	ls.monitor.Reset()
}
