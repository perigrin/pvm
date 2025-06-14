// ABOUTME: Enhanced type-aware autocompletion for LSP integration (EXPERIMENTAL)
// ABOUTME: Provides intelligent code completion based on type information and context

// EXPERIMENTAL FEATURE WARNING:
// This enhanced completion system requires a complete type inference pipeline
// and may not function correctly until the following are implemented:
// - Complete tree-sitter-typed-perl grammar support
// - Full symbol binding and type resolution
// - Stable cross-module type analysis
// TODO: Move to internal/experimental/ when dependencies are resolved

package ls

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/astnav"
	"tamarou.com/pvm/internal/binder"
)

// EnhancedCompletionItem provides rich completion information
type EnhancedCompletionItem struct {
	CompletionItem
	Score          int             // Relevance score for ranking
	TypeMatch      bool            // Whether type matches expected
	Documentation  string          // Full documentation
	Signature      string          // Function/method signature
	ReturnType     string          // Return type for functions
	Parameters     []ParameterInfo // Parameter information
	PostfixSnippet string          // Snippet to insert after
	RequiredImport string          // Module to import if needed
	Tags           []string        // Additional metadata tags
}

// ParameterInfo describes a function parameter
type ParameterInfo struct {
	Name         string
	Type         string
	Optional     bool
	DefaultValue string
	Description  string
}

// EnhancedCompletionContext provides detailed context for completion
type EnhancedCompletionContext struct {
	*CompletionContext
	ExpectedTypes    []string         // Expected type(s) at position
	CurrentScope     *binder.Scope    // Current scope
	ParentNode       ast.Node         // Parent AST node
	AccessPath       []string         // For method chains ($obj->method->)
	InFunctionCall   bool             // Inside function arguments
	FunctionName     string           // Name of function being called
	ParameterIndex   int              // Which parameter position
	LocalVariables   []*binder.Symbol // Local variables in scope
	AvailableModules []string         // Available modules
}

// GetEnhancedCompletions provides intelligent type-aware completions
func (ls *LanguageService) GetEnhancedCompletions(uri string, pos Position) ([]EnhancedCompletionItem, error) {
	completionOp := ls.monitor.StartOperation(context.Background(), "enhanced_completion")
	defer completionOp.Complete()

	ls.mu.RLock()
	doc, exists := ls.documents[uri]
	ls.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	// Get enhanced context
	context := ls.getEnhancedCompletionContext(doc, pos)

	// Generate completions based on context
	var items []EnhancedCompletionItem

	// Add symbol completions with type awareness
	symbolItems := ls.getTypeAwareSymbolCompletions(doc, context)
	items = append(items, symbolItems...)

	// Add method completions for object types
	if len(context.AccessPath) > 0 {
		methodItems := ls.getMethodCompletions(doc, context)
		items = append(items, methodItems...)
	}

	// Add parameter-specific completions
	if context.InFunctionCall {
		paramItems := ls.getParameterCompletions(doc, context)
		items = append(items, paramItems...)
	}

	// Add smart snippets
	snippetItems := ls.getSmartSnippets(doc, context)
	items = append(items, snippetItems...)

	// Score and sort completions
	ls.scoreCompletions(items, context)
	sort.Slice(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})

	return items, nil
}

// getEnhancedCompletionContext analyzes context with type information
func (ls *LanguageService) getEnhancedCompletionContext(doc *Document, pos Position) *EnhancedCompletionContext {
	basic := ls.getCompletionContext(doc, pos)

	enhanced := &EnhancedCompletionContext{
		CompletionContext: basic,
		LocalVariables:    []*binder.Symbol{},
	}

	// Find current scope
	if doc.SymbolTable != nil {
		enhanced.CurrentScope = ls.findScopeAtPosition(doc.SymbolTable, pos)
	}

	// Analyze AST context
	if doc.AST != nil {
		ls.analyzeASTContext(doc, pos, enhanced)
	}

	// Determine expected types
	enhanced.ExpectedTypes = ls.inferExpectedTypes(doc, pos, enhanced)

	// Collect local variables
	if enhanced.CurrentScope != nil {
		enhanced.LocalVariables = ls.collectLocalVariables(enhanced.CurrentScope)
	}

	return enhanced
}

// analyzeASTContext examines the AST to understand completion context
func (ls *LanguageService) analyzeASTContext(doc *Document, pos Position, context *EnhancedCompletionContext) {
	navigator := astnav.NewNavigator(doc.AST.Root)

	// Find nodes at position
	var nodesAtPos []ast.Node
	navigator.Walk(doc.AST.Root, func(node ast.Node) bool {
		nodeStart := node.Start()
		nodeEnd := node.End()

		// Check if position is within node
		if nodeStart.Line-1 <= pos.Line && nodeEnd.Line-1 >= pos.Line {
			if nodeStart.Line-1 == pos.Line && nodeStart.Column-1 > pos.Character {
				return true
			}
			if nodeEnd.Line-1 == pos.Line && nodeEnd.Column-1 < pos.Character {
				return true
			}
			nodesAtPos = append(nodesAtPos, node)
		}
		return true
	})

	// Find most specific node
	if len(nodesAtPos) > 0 {
		context.ParentNode = nodesAtPos[len(nodesAtPos)-1]

		// Analyze specific contexts
		for i := len(nodesAtPos) - 1; i >= 0; i-- {
			node := nodesAtPos[i]
			switch node.Type() {
			case "subroutine_call", "method_call":
				context.InFunctionCall = true
				context.FunctionName = ls.extractCallName(node, doc.Text)
				// Determine parameter position
				context.ParameterIndex = ls.calculateParameterIndex(node, pos, doc.Text)
			case "method_call_expression":
				// Extract access path for method chains
				context.AccessPath = ls.extractAccessPath(node, doc.Text)
			}
		}
	}
}

// inferExpectedTypes determines what types are expected at the position
func (ls *LanguageService) inferExpectedTypes(doc *Document, pos Position, context *EnhancedCompletionContext) []string {
	var types []string

	// If in type annotation context
	if context.InTypeAnnotation {
		return []string{"Type"} // Special marker for type names
	}

	// If in function call, get parameter type
	if context.InFunctionCall && context.FunctionName != "" {
		if funcSymbol := ls.findSymbolByName(doc, context.FunctionName); funcSymbol != nil {
			paramTypes := ls.extractParameterTypes(funcSymbol)
			if context.ParameterIndex < len(paramTypes) {
				types = append(types, paramTypes[context.ParameterIndex])
			}
		}
	}

	// If after assignment, infer from left side
	if context.ParentNode != nil && context.ParentNode.Type() == "assignment" {
		// Find the variable being assigned to
		leftSide := ls.extractAssignmentTarget(context.ParentNode, doc.Text)
		if symbol := ls.findSymbolByName(doc, leftSide); symbol != nil && symbol.Type != "" {
			types = append(types, symbol.Type)
		}
	}

	// Default to Any if no specific type found
	if len(types) == 0 {
		types = append(types, "Any")
	}

	return types
}

// getTypeAwareSymbolCompletions returns completions filtered by type
func (ls *LanguageService) getTypeAwareSymbolCompletions(doc *Document, context *EnhancedCompletionContext) []EnhancedCompletionItem {
	var items []EnhancedCompletionItem

	// Get all visible symbols
	symbols := ls.getVisibleSymbols(doc.SymbolTable, Position{
		Line:      context.ParentNode.Start().Line - 1,
		Character: context.ParentNode.Start().Column - 1,
	})

	for _, symbol := range symbols {
		// Create enhanced completion item
		item := EnhancedCompletionItem{
			CompletionItem: CompletionItem{
				Label:  ls.formatSymbolForCompletion(symbol),
				Kind:   ls.symbolKindToCompletionKind(symbol.Kind),
				Detail: ls.formatSymbolDetail(symbol),
			},
			Documentation: ls.extractDocComment(doc, symbol),
		}

		// Check type compatibility
		if symbol.Type != "" {
			item.TypeMatch = ls.isTypeCompatible(symbol.Type, context.ExpectedTypes)
			item.ReturnType = symbol.Type
		}

		// Add signature for functions
		if symbol.Kind == binder.SymbolSubroutine || symbol.Kind == binder.SymbolMethod {
			item.Signature = ls.extractFunctionSignature(doc, symbol)
			item.Parameters = ls.extractParameterInfo(symbol)

			// Add parentheses snippet
			if len(item.Parameters) > 0 {
				item.PostfixSnippet = ls.generateParameterSnippet(item.Parameters)
			} else {
				item.PostfixSnippet = "()"
			}
		}

		// Add tags for additional context
		if symbol.Flags&binder.SymbolFlagExported != 0 {
			item.Tags = append(item.Tags, "exported")
		}
		if symbol.Flags&binder.SymbolFlagTypeAnnotated != 0 {
			item.Tags = append(item.Tags, "typed")
		}

		items = append(items, item)
	}

	return items
}

// getMethodCompletions returns method completions for object types
func (ls *LanguageService) getMethodCompletions(doc *Document, context *EnhancedCompletionContext) []EnhancedCompletionItem {
	var items []EnhancedCompletionItem

	// Determine the type of the object in the access path
	objectType := ls.resolveAccessPathType(doc, context.AccessPath)
	if objectType == "" {
		return items
	}

	// Find methods for this type
	methods := ls.findMethodsForType(doc, objectType)

	for _, method := range methods {
		item := EnhancedCompletionItem{
			CompletionItem: CompletionItem{
				Label:  method.Name,
				Kind:   CompletionItemKindMethod,
				Detail: fmt.Sprintf("method of %s", objectType),
			},
			Documentation: ls.extractDocComment(doc, method),
			Signature:     ls.extractFunctionSignature(doc, method),
			Parameters:    ls.extractParameterInfo(method),
			ReturnType:    method.Type,
		}

		// Generate parameter snippet
		if len(item.Parameters) > 0 {
			item.PostfixSnippet = ls.generateParameterSnippet(item.Parameters)
		} else {
			item.PostfixSnippet = "()"
		}

		items = append(items, item)
	}

	return items
}

// getParameterCompletions returns completions specific to function parameters
func (ls *LanguageService) getParameterCompletions(doc *Document, context *EnhancedCompletionContext) []EnhancedCompletionItem {
	var items []EnhancedCompletionItem

	// Find the function being called
	funcSymbol := ls.findSymbolByName(doc, context.FunctionName)
	if funcSymbol == nil {
		return items
	}

	// Get parameter info
	params := ls.extractParameterInfo(funcSymbol)
	if context.ParameterIndex >= len(params) {
		return items
	}

	param := params[context.ParameterIndex]

	// Find symbols matching the parameter type
	symbols := ls.getVisibleSymbols(doc.SymbolTable, Position{
		Line:      context.ParentNode.Start().Line - 1,
		Character: context.ParentNode.Start().Column - 1,
	})

	for _, symbol := range symbols {
		if ls.isTypeCompatible(symbol.Type, []string{param.Type}) {
			item := EnhancedCompletionItem{
				CompletionItem: CompletionItem{
					Label:  ls.formatSymbolForCompletion(symbol),
					Kind:   ls.symbolKindToCompletionKind(symbol.Kind),
					Detail: fmt.Sprintf("Parameter %d: %s", context.ParameterIndex+1, param.Type),
				},
				TypeMatch:     true,
				Documentation: fmt.Sprintf("Matches parameter '%s' of type %s", param.Name, param.Type),
			}

			// Boost score for exact type matches
			item.Score = 100

			items = append(items, item)
		}
	}

	return items
}

// getSmartSnippets generates context-aware code snippets
func (ls *LanguageService) getSmartSnippets(doc *Document, context *EnhancedCompletionContext) []EnhancedCompletionItem {
	var items []EnhancedCompletionItem

	// Common Perl idioms based on context
	if context.ParentNode != nil {
		switch context.ParentNode.Type() {
		case "block", "subroutine_body":
			// Control flow snippets
			items = append(items, EnhancedCompletionItem{
				CompletionItem: CompletionItem{
					Label:  "foreach",
					Kind:   CompletionItemKindKeyword,
					Detail: "Iterate over array",
				},
				PostfixSnippet: " my $${1:item} (@${2:array}) {\n\t${3}\n}",
				Documentation:  "Iterate over each element in an array",
				Tags:           []string{"snippet", "control-flow"},
			})

			items = append(items, EnhancedCompletionItem{
				CompletionItem: CompletionItem{
					Label:  "if",
					Kind:   CompletionItemKindKeyword,
					Detail: "Conditional statement",
				},
				PostfixSnippet: " (${1:condition}) {\n\t${2}\n}",
				Documentation:  "Execute code block if condition is true",
				Tags:           []string{"snippet", "control-flow"},
			})

		case "class_body":
			// Class-specific snippets
			items = append(items, EnhancedCompletionItem{
				CompletionItem: CompletionItem{
					Label:  "field",
					Kind:   CompletionItemKindKeyword,
					Detail: "Declare class field",
				},
				PostfixSnippet: " ${1:Type} $${2:name};",
				Documentation:  "Declare a typed field in the class",
				Tags:           []string{"snippet", "class"},
			})

			items = append(items, EnhancedCompletionItem{
				CompletionItem: CompletionItem{
					Label:  "method",
					Kind:   CompletionItemKindKeyword,
					Detail: "Define class method",
				},
				PostfixSnippet: " ${1:name}(${2:params}) -> ${3:ReturnType} {\n\t${4}\n}",
				Documentation:  "Define a new method with type annotations",
				Tags:           []string{"snippet", "class"},
			})
		}
	}

	// Type-specific snippets
	for _, expectedType := range context.ExpectedTypes {
		switch expectedType {
		case "ArrayRef", "ArrayRef[Any]":
			items = append(items, EnhancedCompletionItem{
				CompletionItem: CompletionItem{
					Label:  "[]",
					Kind:   CompletionItemKindSnippet,
					Detail: "Empty array reference",
				},
				PostfixSnippet: "[${1}]",
				Documentation:  "Create a new array reference",
				TypeMatch:      true,
				Tags:           []string{"snippet", "literal"},
			})

		case "HashRef", "HashRef[Any]":
			items = append(items, EnhancedCompletionItem{
				CompletionItem: CompletionItem{
					Label:  "{}",
					Kind:   CompletionItemKindSnippet,
					Detail: "Empty hash reference",
				},
				PostfixSnippet: "{${1}}",
				Documentation:  "Create a new hash reference",
				TypeMatch:      true,
				Tags:           []string{"snippet", "literal"},
			})
		}
	}

	return items
}

// Helper methods for enhanced completion

func (ls *LanguageService) isTypeCompatible(symbolType string, expectedTypes []string) bool {
	// Special case for "Any" type
	for _, expected := range expectedTypes {
		if expected == "Any" || symbolType == "Any" {
			return true
		}
	}

	// Check direct matches
	for _, expected := range expectedTypes {
		if symbolType == expected {
			return true
		}

		// Check if symbol type is a subtype
		if ls.isSubtypeOf(symbolType, expected) {
			return true
		}
	}

	return false
}

func (ls *LanguageService) isSubtypeOf(symbolType, expectedType string) bool {
	// Simple subtype checking
	// In a real implementation, this would use the type hierarchy

	// Handle parameterized types
	if strings.HasPrefix(expectedType, "ArrayRef") && strings.HasPrefix(symbolType, "ArrayRef") {
		return true
	}
	if strings.HasPrefix(expectedType, "HashRef") && strings.HasPrefix(symbolType, "HashRef") {
		return true
	}

	// Handle union types
	if strings.Contains(expectedType, "|") {
		types := strings.Split(expectedType, "|")
		for _, t := range types {
			if strings.TrimSpace(t) == symbolType {
				return true
			}
		}
	}

	return false
}

func (ls *LanguageService) extractParameterTypes(funcSymbol *binder.Symbol) []string {
	// In a real implementation, this would parse the function signature
	// to extract parameter types
	return []string{}
}

func (ls *LanguageService) extractParameterInfo(funcSymbol *binder.Symbol) []ParameterInfo {
	// In a real implementation, this would parse the function signature
	// to extract detailed parameter information
	return []ParameterInfo{}
}

func (ls *LanguageService) generateParameterSnippet(params []ParameterInfo) string {
	if len(params) == 0 {
		return "()"
	}

	var snippetParts []string
	for i, param := range params {
		placeholder := fmt.Sprintf("${%d:%s}", i+1, param.Name)
		if param.Optional {
			placeholder = fmt.Sprintf("${%d:%s = %s}", i+1, param.Name, param.DefaultValue)
		}
		snippetParts = append(snippetParts, placeholder)
	}

	return "(" + strings.Join(snippetParts, ", ") + ")"
}

func (ls *LanguageService) calculateParameterIndex(node ast.Node, pos Position, text string) int {
	// Simple comma counting within the function call
	lines := strings.Split(text, "\n")
	if pos.Line >= len(lines) {
		return 0
	}

	line := lines[pos.Line]
	nodeStart := node.Start()

	// Extract text from function start to cursor position
	startCol := nodeStart.Column - 1
	if startCol >= len(line) || pos.Character >= len(line) {
		return 0
	}

	relevantText := line[startCol:pos.Character]

	// Count commas not inside nested structures
	commaCount := 0
	depth := 0
	for _, ch := range relevantText {
		switch ch {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case ',':
			if depth == 1 { // Only count commas at the parameter level
				commaCount++
			}
		}
	}

	return commaCount
}

func (ls *LanguageService) extractAccessPath(node ast.Node, text string) []string {
	// Extract the chain of method calls
	// $obj->method1->method2
	// Returns ["$obj", "method1", "method2"]
	lines := strings.Split(text, "\n")
	nodeText := ls.extractNodeText(node, lines)

	parts := strings.Split(nodeText, "->")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}

	return parts
}

func (ls *LanguageService) resolveAccessPathType(doc *Document, accessPath []string) string {
	if len(accessPath) == 0 {
		return ""
	}

	// Start with the first element (usually a variable)
	currentType := ""

	// Remove sigil from variable name
	varName := accessPath[0]
	if len(varName) > 1 && strings.ContainsRune("$@%", rune(varName[0])) {
		varName = varName[1:]
	}

	// Find the variable's type
	if symbol := ls.findSymbolByName(doc, varName); symbol != nil {
		currentType = symbol.Type
	}

	// For each method in the chain, resolve its return type
	for i := 1; i < len(accessPath)-1; i++ {
		methodName := accessPath[i]
		// Find the method's return type
		if method := ls.findMethodForType(doc, currentType, methodName); method != nil {
			currentType = method.Type // Return type
		} else {
			break
		}
	}

	return currentType
}

func (ls *LanguageService) findMethodsForType(doc *Document, typeName string) []*binder.Symbol {
	var methods []*binder.Symbol

	// In a real implementation, this would:
	// 1. Find the class/role definition for the type
	// 2. Collect all methods from that class and its parents
	// 3. Include methods from composed roles

	return methods
}

func (ls *LanguageService) findMethodForType(doc *Document, typeName, methodName string) *binder.Symbol {
	methods := ls.findMethodsForType(doc, typeName)
	for _, method := range methods {
		if method.Name == methodName {
			return method
		}
	}
	return nil
}

func (ls *LanguageService) extractAssignmentTarget(node ast.Node, text string) string {
	// Extract the variable name from assignment node
	lines := strings.Split(text, "\n")
	nodeText := ls.extractNodeText(node, lines)

	// Split by = and get the left side
	parts := strings.Split(nodeText, "=")
	if len(parts) > 0 {
		target := strings.TrimSpace(parts[0])
		// Remove "my", "our", etc.
		words := strings.Fields(target)
		if len(words) > 0 {
			lastWord := words[len(words)-1]
			// Remove sigil to get variable name
			if len(lastWord) > 1 && strings.ContainsRune("$@%", rune(lastWord[0])) {
				return lastWord[1:]
			}
			return lastWord
		}
	}

	return ""
}

func (ls *LanguageService) collectLocalVariables(scope *binder.Scope) []*binder.Symbol {
	var locals []*binder.Symbol

	current := scope
	for current != nil {
		for _, symbol := range current.Symbols {
			if symbol.Flags&binder.SymbolFlagLexical != 0 {
				locals = append(locals, symbol)
			}
		}
		current = current.Parent
	}

	return locals
}

func (ls *LanguageService) scoreCompletions(items []EnhancedCompletionItem, context *EnhancedCompletionContext) {
	for i := range items {
		item := &items[i]

		// Base score
		score := 50

		// Type match bonus
		if item.TypeMatch {
			score += 30
		}

		// Local variable bonus
		for _, local := range context.LocalVariables {
			if item.Label == ls.formatSymbolForCompletion(local) {
				score += 20
				break
			}
		}

		// Prefix match bonus
		if context.PreviousWord != "" && strings.HasPrefix(item.Label, context.PreviousWord) {
			score += 25
		}

		// Recently used bonus (would need usage tracking)
		// score += getUsageScore(item.Label)

		// Snippet penalty (less relevant than actual symbols)
		for _, tag := range item.Tags {
			if tag == "snippet" {
				score -= 10
				break
			}
		}

		item.Score = score
	}
}

// CompletionItemKind additions for enhanced completions
const (
	CompletionItemKindSnippet CompletionItemKind = 15
)
