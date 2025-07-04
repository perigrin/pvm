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

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/astnav"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/parser/treesitter"
)

// LanguageService provides language analysis and editor features
type LanguageService struct {
	binder  binder.Binder
	checker *parser.TypeCheck
	// Note: Using parser pool instead of shared instance for thread safety

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
	b := binder.NewBinder()

	checker, err := parser.NewTypeCheck()
	if err != nil {
		return nil, err
	}

	return &LanguageService{
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

		// Use parser pool for thread safety
		ast, err = parser.PooledParserFunc(func(p parser.Parser) (*parser.AST, error) {
			return p.ParseString(text)
		})
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

		// Parse with tree-sitter for CST binding
		tsParser := sitter.NewParser()
		tsParser.SetLanguage(treesitter.Language())
		contentBytes := []byte(text)
		tree := tsParser.Parse(contentBytes, nil)
		if tree == nil {
			bindOp.CompleteWithError(fmt.Errorf("failed to parse with tree-sitter"))
			return fmt.Errorf("failed to parse with tree-sitter")
		}

		symbolTable, err = ls.binder.BindCST(tree.RootNode(), contentBytes, ast.TypeAnnotations)
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

		// Split document text into lines for position-based text extraction
		lines := strings.Split(doc.Text, "\n")

		// Find all nodes that reference this symbol, preferring leaf nodes
		allNodes := []ast.Node{}
		navigator.Walk(doc.AST.Root, func(node ast.Node) bool {
			// Extract text from source using node positions since node.Text() is empty
			nodeText := ls.extractNodeText(node, lines)

			// Check if this node contains the symbol we're looking for
			if ls.nodeMatchesSymbolByText(nodeText, node.Type(), symbol) {
				// Check if this is a leaf node or if we already have a more specific child
				isLeafOrMostSpecific := true

				// Remove any existing nodes that are ancestors of this node
				filteredNodes := []ast.Node{}
				for _, existing := range allNodes {
					existingStart := existing.Start()
					existingEnd := existing.End()
					nodeStart := node.Start()
					nodeEnd := node.End()

					// If existing node contains this node, replace it with this more specific one
					if existingStart.Line <= nodeStart.Line && existingEnd.Line >= nodeEnd.Line &&
						existingStart.Column <= nodeStart.Column && existingEnd.Column >= nodeEnd.Column {
						// Skip the existing node (this one is more specific)
						continue
					}

					// If this node contains an existing node, don't add this one
					if nodeStart.Line <= existingStart.Line && nodeEnd.Line >= existingEnd.Line &&
						nodeStart.Column <= existingStart.Column && nodeEnd.Column >= existingEnd.Column {
						isLeafOrMostSpecific = false
					}

					filteredNodes = append(filteredNodes, existing)
				}

				if isLeafOrMostSpecific {
					// Check if already in filtered list by position
					found := false
					for _, existing := range filteredNodes {
						if existing.Start().Line == node.Start().Line && existing.Start().Column == node.Start().Column {
							found = true
							break
						}
					}
					if !found {
						filteredNodes = append(filteredNodes, node)
					}
				}

				allNodes = filteredNodes
			}
			return true
		})

		// Convert nodes to locations, separating declarations from references
		var declarationLocation *Location
		seenLocations := make(map[string]bool) // Deduplicate by position string

		for _, node := range allNodes {
			nodePos := node.Start()
			// Extract the actual text to find the exact symbol position within the node
			nodeText := ls.extractNodeText(node, lines)
			symbolStart, symbolEnd := ls.findSymbolInText(nodeText, symbol.Name)

			if symbolStart >= 0 {
				startLine := nodePos.Line - 1
				startChar := nodePos.Column - 1 + symbolStart
				endChar := nodePos.Column - 1 + symbolEnd

				// Create a unique key for this location to avoid duplicates
				locationKey := fmt.Sprintf("%d:%d:%d", startLine, startChar, endChar)
				if seenLocations[locationKey] {
					continue // Skip duplicate
				}
				seenLocations[locationKey] = true

				location := Location{
					URI: uri,
					Range: Range{
						Start: Position{
							Line:      startLine,
							Character: startChar,
						},
						End: Position{
							Line:      startLine,
							Character: endChar,
						},
					},
				}

				// Check if this is the declaration by comparing positions
				if symbol.Declaration != nil {
					declPos := symbol.Declaration.Start()
					// Both nodePos and declPos should be 1-based, so compare directly
					// But allow some tolerance for exact position matching
					if nodePos.Line == declPos.Line &&
						(nodePos.Column == declPos.Column ||
							// Allow for slight column differences (e.g., sigil position vs identifier position)
							(nodePos.Column >= declPos.Column-1 && nodePos.Column <= declPos.Column+1)) {
						// This is the declaration
						if includeDeclaration {
							declarationLocation = &location
						}
						continue
					}
				}

				// This is a reference
				locations = append(locations, location)
			}
		}

		// Add declaration at the beginning if requested and found
		if declarationLocation != nil {
			locations = append([]Location{*declarationLocation}, locations...)
		}
	}

	return locations, nil
}

// extractNodeText extracts text from source using node positions
func (ls *LanguageService) extractNodeText(node ast.Node, lines []string) string {
	start := node.Start()
	end := node.End()

	if start.Line <= 0 || start.Line > len(lines) {
		return ""
	}

	line := lines[start.Line-1] // Convert 1-based to 0-based

	// Handle single line extraction
	if start.Line == end.Line {
		if start.Column > 0 && end.Column > 0 &&
			start.Column-1 < len(line) && end.Column-1 <= len(line) {
			return line[start.Column-1 : end.Column-1]
		}
	} else {
		// For multi-line nodes, just return the first line for now
		if start.Column > 0 && start.Column-1 < len(line) {
			return line[start.Column-1:]
		}
	}

	return ""
}

// nodeMatchesSymbolByText checks if extracted text contains the symbol
func (ls *LanguageService) nodeMatchesSymbolByText(text, nodeType string, symbol *binder.Symbol) bool {
	if text == "" {
		return false
	}

	// Look for the symbol name in the text
	symbolName := symbol.Name

	// Only match specific node types to avoid duplicates from parent nodes
	switch nodeType {
	case "scalar_variable", "array_variable", "hash_variable", "variable_expression", "identifier", "scalar", "array", "hash":
		// For variable nodes, match exact variable references
		trimmedText := strings.TrimSpace(text)
		if trimmedText == "$"+symbolName || trimmedText == "@"+symbolName ||
			trimmedText == "%"+symbolName || trimmedText == "*"+symbolName {
			return true
		}
		// Also match bare identifiers for subroutines
		if trimmedText == symbolName {
			return true
		}
	case "var_decl", "variable":
		// For variable declaration nodes, check if they contain the variable
		if strings.Contains(text, "$"+symbolName) || strings.Contains(text, "@"+symbolName) ||
			strings.Contains(text, "%"+symbolName) || strings.Contains(text, "*"+symbolName) {
			return true
		}
	case "sub_decl", "method_decl":
		// For subroutine/method declarations, check if they contain the subroutine name
		// The text might be "sub greet {" so we need to extract the name
		if strings.Contains(text, "sub "+symbolName) || strings.Contains(text, "method "+symbolName) {
			return true
		}
	case "subroutine_call", "method_call":
		// For subroutine calls, match the call name
		trimmedText := strings.TrimSpace(text)
		// Remove parentheses if present
		callName := strings.TrimSuffix(trimmedText, "()")
		callName = strings.TrimSuffix(callName, "(")
		callName = strings.TrimSpace(callName)
		if callName == symbolName {
			return true
		}
	default:
		// For other node types, be more restrictive to avoid false matches
		// Only match if the text is exactly the symbol (no extra content)
		trimmedText := strings.TrimSpace(text)
		if len(trimmedText) <= len(symbolName)+1 { // Allow for sigil
			if trimmedText == "$"+symbolName || trimmedText == "@"+symbolName ||
				trimmedText == "%"+symbolName || trimmedText == "*"+symbolName ||
				trimmedText == symbolName {
				return true
			}
		}
	}

	return false
}

// findSymbolInText finds the start and end position of a symbol within text
func (ls *LanguageService) findSymbolInText(text, symbolName string) (int, int) {
	// Look for $symbolName first (most common case)
	if idx := strings.Index(text, "$"+symbolName); idx >= 0 {
		return idx, idx + 1 + len(symbolName)
	}

	// Look for @symbolName
	if idx := strings.Index(text, "@"+symbolName); idx >= 0 {
		return idx, idx + 1 + len(symbolName)
	}

	// Look for %symbolName
	if idx := strings.Index(text, "%"+symbolName); idx >= 0 {
		return idx, idx + 1 + len(symbolName)
	}

	// Look for bare symbolName
	if idx := strings.Index(text, symbolName); idx >= 0 {
		return idx, idx + len(symbolName)
	}

	return -1, -1
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

// FormattingOptions represents document formatting options
type FormattingOptions struct {
	TabSize      int  // Size of a tab in spaces
	InsertSpaces bool // Prefer spaces over tabs
}

// CodeActionContext represents the context for code action requests
type CodeActionContext struct {
	Diagnostics []Diagnostic // Diagnostics that triggered the code action
}

// CodeAction represents a code action (quick fix, refactoring, etc.)
type CodeAction struct {
	Title   string         // Human-readable title
	Kind    string         // Kind of code action (quickfix, refactor, etc.)
	Edit    *WorkspaceEdit // Optional workspace edit to apply
	Command *Command       // Optional command to execute
}

// WorkspaceEdit represents changes to multiple documents
type WorkspaceEdit struct {
	Changes map[string][]TextEdit // URI -> list of edits
}

// Command represents a command that can be executed
type Command struct {
	Title     string        // Human-readable command title
	Command   string        // Command identifier
	Arguments []interface{} // Optional command arguments
}

// Diagnostic represents a diagnostic message (error, warning, etc.)
type Diagnostic struct {
	Range    Range               // Source range where diagnostic applies
	Severity *DiagnosticSeverity // Optional severity level
	Message  string              // Diagnostic message
}

// DiagnosticSeverity represents diagnostic severity levels
type DiagnosticSeverity int

const (
	DiagnosticSeverityError       DiagnosticSeverity = 1
	DiagnosticSeverityWarning     DiagnosticSeverity = 2
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	DiagnosticSeverityHint        DiagnosticSeverity = 4
)

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

// FormatDocument provides document formatting functionality
func (ls *LanguageService) FormatDocument(uri string, options FormattingOptions) ([]TextEdit, error) {
	doc, exists := ls.documents[uri]
	if !exists {
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	var edits []TextEdit
	lines := strings.Split(doc.Text, "\n")

	for lineNum, line := range lines {
		var newLine string
		changed := false

		// Rule 1: Trim trailing whitespace
		trimmed := strings.TrimRight(line, " \t")
		if trimmed != line {
			newLine = trimmed
			changed = true
		} else {
			newLine = line
		}

		// Rule 2: Convert tabs to spaces if InsertSpaces is true
		if options.InsertSpaces && strings.Contains(newLine, "\t") {
			spaces := strings.Repeat(" ", options.TabSize)
			converted := strings.ReplaceAll(newLine, "\t", spaces)
			if converted != newLine {
				newLine = converted
				changed = true
			}
		}

		// Create edit if line was changed
		if changed {
			edit := TextEdit{
				Range: Range{
					Start: Position{Line: lineNum, Character: 0},
					End:   Position{Line: lineNum, Character: len(line)},
				},
				NewText: newLine,
			}
			edits = append(edits, edit)
		}
	}

	return edits, nil
}

// GenerateCodeActions provides code action functionality (quick fixes, refactoring)
func (ls *LanguageService) GenerateCodeActions(uri string, range_ Range, context CodeActionContext) ([]CodeAction, error) {
	doc, exists := ls.documents[uri]
	if !exists {
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	var actions []CodeAction

	// Process diagnostics to generate quick fixes
	for _, diagnostic := range context.Diagnostics {
		if strings.Contains(diagnostic.Message, "undefined") && strings.Contains(diagnostic.Message, "$") {
			// Quick fix for undefined variables
			// Extract variable name from diagnostic message
			if varName := extractVariableName(diagnostic.Message); varName != "" {
				action := CodeAction{
					Title: fmt.Sprintf("Declare variable '%s'", varName),
					Kind:  "quickfix",
					Edit: &WorkspaceEdit{
						Changes: map[string][]TextEdit{
							uri: {
								{
									Range: Range{
										Start: Position{Line: diagnostic.Range.Start.Line, Character: 0},
										End:   Position{Line: diagnostic.Range.Start.Line, Character: 0},
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
	}

	// Generate refactoring actions for selected code (only if no diagnostics to fix)
	if len(context.Diagnostics) == 0 && range_.Start.Line == range_.End.Line && range_.End.Character > range_.Start.Character {
		// Extract variable refactoring
		lines := strings.Split(doc.Text, "\n")
		if range_.Start.Line < len(lines) {
			line := lines[range_.Start.Line]
			if range_.Start.Character < len(line) && range_.End.Character <= len(line) {
				selectedText := line[range_.Start.Character:range_.End.Character]
				if strings.TrimSpace(selectedText) != "" {
					action := CodeAction{
						Title: "Extract to variable",
						Kind:  "refactor.extract",
						Edit: &WorkspaceEdit{
							Changes: map[string][]TextEdit{
								uri: {
									{
										Range: Range{
											Start: Position{Line: range_.Start.Line, Character: 0},
											End:   Position{Line: range_.Start.Line, Character: 0},
										},
										NewText: "my $extracted_value = " + selectedText + ";\n",
									},
									{
										Range:   range_,
										NewText: "$extracted_value",
									},
								},
							},
						},
					}
					actions = append(actions, action)
				}
			}
		}
	}

	return actions, nil
}

// Helper function to extract variable name from diagnostic message
func extractVariableName(message string) string {
	// Look for patterns like "Variable $foo is undefined"
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

// Helper function to check if character is alphanumeric
func isAlphaNumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
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
	// For now, start from global scope to ensure we can find all symbols
	return symbolTable.GlobalScope
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

	// Handle case where we're starting on a sigil
	if strings.ContainsRune("$@%*", rune(line[pos.Character])) {
		// Move end forward to capture the variable name after the sigil
		end++
		for end < len(line) && ls.isWordChar(line[end]) {
			end++
		}
	} else {
		// Include sigil for variables if one exists before the current position
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
	text := node.Text()
	if text == "" {
		return false
	}

	switch node.Type() {
	case "variable_expression", "scalar_variable", "variable_declaration":
		// Handle variable references like $foo, @arr, %hash
		if len(text) > 0 && strings.ContainsRune("$@%*", rune(text[0])) {
			varName := text[1:]
			return varName == symbol.Name
		}
		return text == symbol.Name

	case "subroutine_call", "method_call":
		// Handle subroutine/method calls - extract name without parentheses
		callName := strings.TrimSuffix(text, "()")
		callName = strings.TrimSpace(callName)
		return callName == symbol.Name

	case "identifier":
		// Handle plain identifiers
		return text == symbol.Name

	default:
		// For any other node type, check if the text matches the symbol name
		// This is a fallback for cases we might have missed
		if len(text) > 0 && strings.ContainsRune("$@%*", rune(text[0])) {
			varName := text[1:]
			return varName == symbol.Name
		}
		return text == symbol.Name
	}
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

// GetDocumentForDebug returns document for debugging (should only be used in tests)
func (ls *LanguageService) GetDocumentForDebug(uri string) (*Document, bool) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	doc, exists := ls.documents[uri]
	return doc, exists
}

// FindSymbolAtPosition finds the symbol at a specific position in a document
func (ls *LanguageService) FindSymbolAtPosition(uri string, pos Position) (*binder.Symbol, error) {
	ls.mu.RLock()
	doc, exists := ls.documents[uri]
	ls.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	return ls.findSymbolAtPosition(doc, pos), nil
}
