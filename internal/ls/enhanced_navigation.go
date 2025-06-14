// ABOUTME: Enhanced symbol navigation features for LSP integration (EXPERIMENTAL)
// ABOUTME: Provides improved goto definition and find references using complete AST

// EXPERIMENTAL FEATURE WARNING:
// This enhanced navigation system requires complete symbol resolution
// and may not work correctly until the following are implemented:
// - Complete symbol binding across module boundaries
// - Stable type information extraction
// - Cross-module navigation infrastructure
// TODO: Move to internal/experimental/ when dependencies are resolved

package ls

import (
	"fmt"
	"path/filepath"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/astnav"
	"tamarou.com/pvm/internal/binder"
)

// EnhancedDefinition provides more detailed definition information
type EnhancedDefinition struct {
	Location   Location
	Symbol     *binder.Symbol
	TypeInfo   *TypeInformation
	DocComment string
	Signature  string // For functions/methods
	IsExported bool
	ModulePath string // For cross-module navigation
}

// TypeInformation provides detailed type information
type TypeInformation struct {
	BaseType          string
	TypeParameters    []string // For generic types
	Constraints       []string // Type constraints
	IsUnion           bool
	UnionTypes        []string
	IsIntersection    bool
	IntersectionTypes []string
}

// GetEnhancedDefinition finds the definition with complete type information
func (ls *LanguageService) GetEnhancedDefinition(uri string, pos Position) (*EnhancedDefinition, error) {
	ls.mu.RLock()
	doc, exists := ls.documents[uri]
	ls.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	// Find the symbol at position
	symbol := ls.findSymbolAtPosition(doc, pos)
	if symbol == nil {
		return nil, nil
	}

	// Get basic definition first
	basicDef, err := ls.GetDefinition(uri, pos)
	if err != nil || basicDef == nil {
		return nil, err
	}

	// Enhance with additional information
	enhanced := &EnhancedDefinition{
		Location:   basicDef.Location,
		Symbol:     symbol,
		IsExported: symbol.Flags&binder.SymbolFlagExported != 0,
	}

	// Extract type information
	enhanced.TypeInfo = ls.extractTypeInformation(symbol)

	// Get documentation comment if available
	enhanced.DocComment = ls.extractDocComment(doc, symbol)

	// Get signature for functions/methods
	if symbol.Kind == binder.SymbolSubroutine || symbol.Kind == binder.SymbolMethod {
		enhanced.Signature = ls.extractFunctionSignature(doc, symbol)
	}

	// Handle cross-module navigation
	if symbol.Package != "" && symbol.Package != "main" {
		enhanced.ModulePath = ls.resolveModulePath(symbol.Package)
	}

	return enhanced, nil
}

// extractTypeInformation analyzes the symbol's type annotation
func (ls *LanguageService) extractTypeInformation(symbol *binder.Symbol) *TypeInformation {
	if symbol.Type == "" {
		return nil
	}

	info := &TypeInformation{
		BaseType: symbol.Type,
	}

	// Parse complex type annotations
	typeStr := symbol.Type

	// Check for union types (Int|Str)
	if strings.Contains(typeStr, "|") {
		info.IsUnion = true
		info.UnionTypes = strings.Split(typeStr, "|")
		for i, t := range info.UnionTypes {
			info.UnionTypes[i] = strings.TrimSpace(t)
		}
	}

	// Check for intersection types (Object&Serializable)
	if strings.Contains(typeStr, "&") && !info.IsUnion {
		info.IsIntersection = true
		info.IntersectionTypes = strings.Split(typeStr, "&")
		for i, t := range info.IntersectionTypes {
			info.IntersectionTypes[i] = strings.TrimSpace(t)
		}
	}

	// Check for parameterized types (ArrayRef[Int])
	if strings.Contains(typeStr, "[") && strings.Contains(typeStr, "]") {
		start := strings.Index(typeStr, "[")
		end := strings.LastIndex(typeStr, "]")
		if start < end {
			info.BaseType = typeStr[:start]
			params := typeStr[start+1 : end]
			// Handle nested parameterized types
			info.TypeParameters = ls.parseTypeParameters(params)
		}
	}

	return info
}

// parseTypeParameters handles nested type parameters
func (ls *LanguageService) parseTypeParameters(params string) []string {
	var result []string
	var current strings.Builder
	depth := 0

	for _, ch := range params {
		switch ch {
		case '[':
			depth++
			current.WriteRune(ch)
		case ']':
			depth--
			current.WriteRune(ch)
		case ',':
			if depth == 0 {
				result = append(result, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	return result
}

// extractDocComment finds and extracts documentation comments
func (ls *LanguageService) extractDocComment(doc *Document, symbol *binder.Symbol) string {
	if symbol.Declaration == nil || doc.AST == nil {
		return ""
	}

	// Look for comments above the declaration
	declLine := symbol.Declaration.Start().Line
	lines := strings.Split(doc.Text, "\n")

	var comments []string
	for i := declLine - 2; i >= 0 && i < len(lines); i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "#") {
			// Extract comment text
			comment := strings.TrimPrefix(line, "#")
			comment = strings.TrimSpace(comment)
			comments = append([]string{comment}, comments...)
		} else if line == "" {
			// Continue through empty lines
			continue
		} else {
			// Stop at non-comment, non-empty line
			break
		}
	}

	return strings.Join(comments, "\n")
}

// extractFunctionSignature extracts the full signature of a function/method
func (ls *LanguageService) extractFunctionSignature(doc *Document, symbol *binder.Symbol) string {
	if symbol.Declaration == nil {
		return ""
	}

	// Use AST navigation to find the signature
	navigator := astnav.NewNavigator(doc.AST.Root)

	var signature string
	navigator.Walk(symbol.Declaration, func(node ast.Node) bool {
		switch node.Type() {
		case "signature", "method_signature":
			signature = ls.reconstructSignature(node, doc.Text)
			return false // Stop walking
		}
		return true
	})

	if signature == "" && symbol.Kind == binder.SymbolSubroutine {
		// Fallback for simple subroutines
		signature = fmt.Sprintf("sub %s", symbol.Name)
	}

	return signature
}

// reconstructSignature builds a readable signature from AST nodes
func (ls *LanguageService) reconstructSignature(node ast.Node, text string) string {
	lines := strings.Split(text, "\n")

	start := node.Start()
	end := node.End()

	if start.Line == end.Line {
		line := lines[start.Line-1]
		if start.Column > 0 && end.Column <= len(line) {
			return strings.TrimSpace(line[start.Column-1 : end.Column-1])
		}
	}

	// Multi-line signature
	var parts []string
	for i := start.Line; i <= end.Line && i <= len(lines); i++ {
		line := lines[i-1]
		if i == start.Line {
			if start.Column > 0 && start.Column <= len(line) {
				parts = append(parts, line[start.Column-1:])
			}
		} else if i == end.Line {
			if end.Column > 0 && end.Column <= len(line) {
				parts = append(parts, line[:end.Column-1])
			}
		} else {
			parts = append(parts, line)
		}
	}

	return strings.TrimSpace(strings.Join(parts, " "))
}

// resolveModulePath attempts to find the file path for a module
func (ls *LanguageService) resolveModulePath(packageName string) string {
	// Convert package name to file path
	// Foo::Bar::Baz -> Foo/Bar/Baz.pm
	parts := strings.Split(packageName, "::")
	relativePath := filepath.Join(parts...) + ".pm"

	// In a real implementation, this would search @INC paths
	// For now, return the relative path
	return relativePath
}

// FindEnhancedReferences finds all references with context information
func (ls *LanguageService) FindEnhancedReferences(uri string, pos Position, includeDeclaration bool) ([]EnhancedReference, error) {
	// Get basic references first
	basicRefs, err := ls.FindReferences(uri, pos, includeDeclaration)
	if err != nil {
		return nil, err
	}

	doc, exists := ls.documents[uri]
	if !exists {
		return nil, fmt.Errorf("document not found")
	}

	symbol := ls.findSymbolAtPosition(doc, pos)
	if symbol == nil {
		return nil, nil
	}

	// Enhance each reference with context
	enhanced := make([]EnhancedReference, 0, len(basicRefs))
	for _, ref := range basicRefs {
		enhancedRef := EnhancedReference{
			Location: ref,
			Context:  ls.extractReferenceContext(doc, ref, symbol),
		}
		enhanced = append(enhanced, enhancedRef)
	}

	return enhanced, nil
}

// EnhancedReference provides reference with context
type EnhancedReference struct {
	Location Location
	Context  ReferenceContext
}

// ReferenceContext describes how a symbol is used at a reference location
type ReferenceContext struct {
	Kind        ReferenceKind
	IsWrite     bool
	IsRead      bool
	InCondition bool
	InLoop      bool
	ParentNode  string // Type of parent AST node
	LineText    string // Full line text for context
}

// ReferenceKind represents the kind of reference
type ReferenceKind int

const (
	ReferenceKindDeclaration ReferenceKind = iota
	ReferenceKindAssignment
	ReferenceKindFunctionCall
	ReferenceKindMethodCall
	ReferenceKindParameterPass
	ReferenceKindReturn
	ReferenceKindCondition
	ReferenceKindOther
)

// extractReferenceContext analyzes how a symbol is used at a reference location
func (ls *LanguageService) extractReferenceContext(doc *Document, location Location, symbol *binder.Symbol) ReferenceContext {
	lines := strings.Split(doc.Text, "\n")

	context := ReferenceContext{
		Kind: ReferenceKindOther,
	}

	// Get the line text
	if location.Range.Start.Line < len(lines) {
		context.LineText = strings.TrimSpace(lines[location.Range.Start.Line])
	}

	// Find the AST node at this location
	if doc.AST != nil {
		navigator := astnav.NewNavigator(doc.AST.Root)

		// Find the most specific node at this position
		var targetNode ast.Node
		navigator.Walk(doc.AST.Root, func(node ast.Node) bool {
			nodeStart := node.Start()
			nodeEnd := node.End()

			// Check if node contains the location
			if nodeStart.Line-1 <= location.Range.Start.Line &&
				nodeEnd.Line-1 >= location.Range.End.Line {
				targetNode = node
			}
			return true
		})

		if targetNode != nil {
			context.ParentNode = targetNode.Type()

			// Analyze the context based on parent node type
			switch targetNode.Type() {
			case "assignment", "assignment_expression":
				context.Kind = ReferenceKindAssignment
				context.IsWrite = true
			case "subroutine_call", "method_call":
				if symbol.Kind == binder.SymbolSubroutine || symbol.Kind == binder.SymbolMethod {
					context.Kind = ReferenceKindFunctionCall
				} else {
					context.Kind = ReferenceKindParameterPass
				}
				context.IsRead = true
			case "if_statement", "unless_statement", "while_statement":
				context.Kind = ReferenceKindCondition
				context.InCondition = true
				context.IsRead = true
			case "for_statement", "foreach_statement":
				context.InLoop = true
				context.IsRead = true
			case "return_statement":
				context.Kind = ReferenceKindReturn
				context.IsRead = true
			case "var_decl", "variable_declaration":
				context.Kind = ReferenceKindDeclaration
				if strings.Contains(context.LineText, "=") {
					context.IsWrite = true
				}
			default:
				// Default to read for most contexts
				context.IsRead = true
			}
		}
	}

	return context
}

// GetCallHierarchy provides call hierarchy information
func (ls *LanguageService) GetCallHierarchy(uri string, pos Position) (*CallHierarchyItem, error) {
	doc, exists := ls.documents[uri]
	if !exists {
		return nil, fmt.Errorf("document not found")
	}

	symbol := ls.findSymbolAtPosition(doc, pos)
	if symbol == nil {
		return nil, nil
	}

	// Only provide call hierarchy for functions/methods
	if symbol.Kind != binder.SymbolSubroutine && symbol.Kind != binder.SymbolMethod {
		return nil, nil
	}

	item := &CallHierarchyItem{
		Name:   symbol.Name,
		Kind:   ls.symbolKindToString(symbol.Kind),
		URI:    uri,
		Range:  ls.getSymbolRange(symbol),
		Detail: ls.extractFunctionSignature(doc, symbol),
	}

	return item, nil
}

// CallHierarchyItem represents an item in the call hierarchy
type CallHierarchyItem struct {
	Name   string
	Kind   string
	URI    string
	Range  Range
	Detail string
}

// GetIncomingCalls finds all calls to a function/method
func (ls *LanguageService) GetIncomingCalls(item *CallHierarchyItem) ([]CallHierarchyIncomingCall, error) {
	var calls []CallHierarchyIncomingCall

	// Search all documents for calls to this function
	for docURI, doc := range ls.documents {
		if doc.AST == nil {
			continue
		}

		navigator := astnav.NewNavigator(doc.AST.Root)

		navigator.Walk(doc.AST.Root, func(node ast.Node) bool {
			if node.Type() == "subroutine_call" || node.Type() == "method_call" {
				// Extract the function name being called
				callName := ls.extractCallName(node, doc.Text)
				if callName == item.Name {
					// Find the containing function
					containingFunc := ls.findContainingFunction(doc, node)
					if containingFunc != nil {
						call := CallHierarchyIncomingCall{
							From: CallHierarchyItem{
								Name:  containingFunc.Name,
								Kind:  ls.symbolKindToString(containingFunc.Kind),
								URI:   docURI,
								Range: ls.getSymbolRange(containingFunc),
							},
							FromRanges: []Range{ls.nodeToRange(node)},
						}
						calls = append(calls, call)
					}
				}
			}
			return true
		})
	}

	return calls, nil
}

// CallHierarchyIncomingCall represents an incoming call
type CallHierarchyIncomingCall struct {
	From       CallHierarchyItem
	FromRanges []Range
}

// GetOutgoingCalls finds all calls made by a function/method
func (ls *LanguageService) GetOutgoingCalls(item *CallHierarchyItem) ([]CallHierarchyOutgoingCall, error) {
	doc, exists := ls.documents[item.URI]
	if !exists {
		return nil, fmt.Errorf("document not found")
	}

	var calls []CallHierarchyOutgoingCall

	// Find the function's AST node
	funcSymbol := ls.findSymbolByName(doc, item.Name)
	if funcSymbol == nil || funcSymbol.Declaration == nil {
		return calls, nil
	}

	navigator := astnav.NewNavigator(funcSymbol.Declaration)

	navigator.Walk(funcSymbol.Declaration, func(node ast.Node) bool {
		if node.Type() == "subroutine_call" || node.Type() == "method_call" {
			callName := ls.extractCallName(node, doc.Text)
			if callName != "" && callName != item.Name { // Exclude recursive calls
				// Try to find the called function
				calledSymbol := ls.findSymbolByName(doc, callName)
				if calledSymbol != nil {
					call := CallHierarchyOutgoingCall{
						To: CallHierarchyItem{
							Name:  calledSymbol.Name,
							Kind:  ls.symbolKindToString(calledSymbol.Kind),
							URI:   item.URI,
							Range: ls.getSymbolRange(calledSymbol),
						},
						FromRanges: []Range{ls.nodeToRange(node)},
					}
					calls = append(calls, call)
				}
			}
		}
		return true
	})

	return calls, nil
}

// CallHierarchyOutgoingCall represents an outgoing call
type CallHierarchyOutgoingCall struct {
	To         CallHierarchyItem
	FromRanges []Range
}

// Helper methods

func (ls *LanguageService) extractCallName(node ast.Node, text string) string {
	lines := strings.Split(text, "\n")
	nodeText := ls.extractNodeText(node, lines)

	// Remove parentheses and arguments
	callName := strings.TrimSpace(nodeText)
	if idx := strings.Index(callName, "("); idx > 0 {
		callName = callName[:idx]
	}

	// Handle method calls (remove object part)
	if strings.Contains(callName, "->") {
		parts := strings.Split(callName, "->")
		if len(parts) > 1 {
			callName = strings.TrimSpace(parts[len(parts)-1])
		}
	}

	return callName
}

func (ls *LanguageService) findContainingFunction(doc *Document, node ast.Node) *binder.Symbol {
	// In a real implementation, this would traverse up the AST
	// to find the containing function/method
	// For now, return nil
	return nil
}

func (ls *LanguageService) findSymbolByName(doc *Document, name string) *binder.Symbol {
	if doc.SymbolTable == nil {
		return nil
	}

	return ls.searchSymbolInScope(doc.SymbolTable.GlobalScope, name)
}

func (ls *LanguageService) searchSymbolInScope(scope *binder.Scope, name string) *binder.Symbol {
	if scope == nil {
		return nil
	}

	if symbol, exists := scope.Symbols[name]; exists {
		return symbol
	}

	for _, child := range scope.Children {
		if found := ls.searchSymbolInScope(child, name); found != nil {
			return found
		}
	}

	return nil
}

func (ls *LanguageService) getSymbolRange(symbol *binder.Symbol) Range {
	if symbol.Declaration != nil {
		return ls.nodeToRange(symbol.Declaration)
	}

	// Fallback to symbol position
	return Range{
		Start: Position{
			Line:      symbol.Position.Line - 1,
			Character: symbol.Position.Column - 1,
		},
		End: Position{
			Line:      symbol.Position.Line - 1,
			Character: symbol.Position.Column - 1 + len(symbol.Name),
		},
	}
}

func (ls *LanguageService) nodeToRange(node ast.Node) Range {
	start := node.Start()
	end := node.End()

	return Range{
		Start: Position{
			Line:      start.Line - 1,
			Character: start.Column - 1,
		},
		End: Position{
			Line:      end.Line - 1,
			Character: end.Column - 1,
		},
	}
}
