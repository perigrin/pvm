// ABOUTME: Type information query functionality for LSP
// ABOUTME: Provides detailed type information and symbol analysis

package lsp

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
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
	References    []Location     `json:"references,omitempty"`
}

// FunctionSig represents a function signature
type FunctionSig struct {
	Parameters []ParameterInfo `json:"parameters"`
	ReturnType string          `json:"returnType,omitempty"`
	IsMethod   bool            `json:"isMethod"`
}

// ParameterInfo represents parameter information
type ParameterInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Optional     bool   `json:"optional"`
	DefaultValue string `json:"defaultValue,omitempty"`
	Description  string `json:"description,omitempty"`
}

// PropertyInfo represents class/object property information
type PropertyInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Visibility  string `json:"visibility"` // public, private, protected
	ReadOnly    bool   `json:"readOnly"`
	Description string `json:"description,omitempty"`
}

// MethodInfo represents method information
type MethodInfo struct {
	Name        string      `json:"name"`
	Signature   FunctionSig `json:"signature"`
	Visibility  string      `json:"visibility"`
	Static      bool        `json:"static"`
	Description string      `json:"description,omitempty"`
}

// TypeQueryService provides type information queries
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
	doc, exists := s.server.getDocument(query.URI)
	if !exists {
		return nil, fmt.Errorf("document not found: %s", query.URI)
	}

	// Extract symbol at position
	symbol := s.extractSymbolAtPosition(doc, query.Position)
	if symbol == "" {
		return nil, fmt.Errorf("no symbol found at position %d:%d", query.Position.Line, query.Position.Character)
	}

	// Query type information for the symbol
	return s.getTypeInfo(doc, symbol, query.Position)
}

// QuerySymbol queries type information for a specific symbol
func (s *TypeQueryService) QuerySymbol(uri, symbol string) (*TypeInfo, error) {
	doc, exists := s.server.getDocument(uri)
	if !exists {
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	return s.getTypeInfo(doc, symbol, Position{})
}

// GetAvailableSymbols returns all available symbols in a document
func (s *TypeQueryService) GetAvailableSymbols(uri string) ([]TypeInfo, error) {
	doc, exists := s.server.getDocument(uri)
	if !exists {
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	var symbols []TypeInfo

	// Extract symbols from type annotations if available
	if doc.AST != nil && len(doc.AST.TypeAnnotations) > 0 {
		for _, annotation := range doc.AST.TypeAnnotations {
			info := s.createTypeInfoFromAnnotation(annotation)
			if info != nil {
				symbols = append(symbols, *info)
			}
		}
	}

	// Add builtin symbols
	symbols = append(symbols, s.getBuiltinSymbols()...)

	return symbols, nil
}

// extractSymbolAtPosition extracts the symbol at a given position
func (s *TypeQueryService) extractSymbolAtPosition(doc *Document, pos Position) string {
	lines := strings.Split(doc.Text, "\n")
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

	// Move start backwards to beginning of word
	for start > 0 && isSymbolChar(line[start-1]) {
		start--
	}

	// Move end forwards to end of word
	for end < len(line) && isSymbolChar(line[end]) {
		end++
	}

	if start == end {
		return ""
	}

	return line[start:end]
}

// getTypeInfo retrieves comprehensive type information for a symbol
func (s *TypeQueryService) getTypeInfo(doc *Document, symbol string, pos Position) (*TypeInfo, error) {
	info := &TypeInfo{
		Symbol: symbol,
	}

	// Check if it's a builtin type
	if builtinInfo := s.getBuiltinTypeInfo(symbol); builtinInfo != nil {
		return builtinInfo, nil
	}

	// Check if it's a builtin function
	if builtinInfo := s.getBuiltinFunctionInfo(symbol); builtinInfo != nil {
		return builtinInfo, nil
	}

	// Check if it's a Perl keyword
	if keywordInfo := s.getKeywordInfo(symbol); keywordInfo != nil {
		return keywordInfo, nil
	}

	// Check if it's defined in the document's type annotations
	if doc.AST != nil {
		for _, annotation := range doc.AST.TypeAnnotations {
			if s.symbolMatchesAnnotation(symbol, annotation) {
				return s.createDetailedTypeInfo(annotation), nil
			}
		}
	}

	// Try to infer type from usage patterns
	if inferredInfo := s.inferTypeFromUsage(doc, symbol); inferredInfo != nil {
		return inferredInfo, nil
	}

	// Default case - generic symbol info
	info.Type = "unknown"
	info.Kind = "symbol"
	info.Documentation = fmt.Sprintf("Symbol '%s' found in code", symbol)

	return info, nil
}

// getBuiltinTypeInfo returns information about builtin types
func (s *TypeQueryService) getBuiltinTypeInfo(typeName string) *TypeInfo {
	builtinTypes := map[string]TypeInfo{
		"Str": {
			Symbol:        "Str",
			Type:          "Str",
			Kind:          "type",
			Documentation: "String type - represents text values. Supports concatenation, substring operations, and pattern matching.",
			Examples:      []string{`my Str $name = "John";`, `my Str $greeting = "Hello " . $name;`},
		},
		"Int": {
			Symbol:        "Int",
			Type:          "Int",
			Kind:          "type",
			Documentation: "Integer type - represents whole numbers. Supports arithmetic operations and comparisons.",
			Examples:      []string{`my Int $count = 42;`, `my Int $result = $count + 10;`},
		},
		"Num": {
			Symbol:        "Num",
			Type:          "Num",
			Kind:          "type",
			Documentation: "Number type - represents numeric values including decimals. Supports all arithmetic operations.",
			Examples:      []string{`my Num $pi = 3.14159;`, `my Num $result = $pi * 2;`},
		},
		"Bool": {
			Symbol:        "Bool",
			Type:          "Bool",
			Kind:          "type",
			Documentation: "Boolean type - represents true/false values. Used in conditional expressions.",
			Examples:      []string{`my Bool $is_valid = 1;`, `my Bool $enabled = $count > 0;`},
		},
		"ArrayRef": {
			Symbol:        "ArrayRef",
			Type:          "ArrayRef[T]",
			Kind:          "type",
			Documentation: "Array reference type - reference to an array. Can be parameterized with element type.",
			Examples:      []string{`my ArrayRef[Str] $names = ["Alice", "Bob"];`, `my ArrayRef[Int] $numbers = [1, 2, 3];`},
		},
		"HashRef": {
			Symbol:        "HashRef",
			Type:          "HashRef[T]",
			Kind:          "type",
			Documentation: "Hash reference type - reference to a hash. Can be parameterized with value type.",
			Examples:      []string{`my HashRef[Str] $config = {name => "test"};`, `my HashRef[Int] $scores = {alice => 95};`},
		},
		"Maybe": {
			Symbol:        "Maybe",
			Type:          "Maybe[T]",
			Kind:          "type",
			Documentation: "Maybe type - optional value that can be undef. Used for nullable values.",
			Examples:      []string{`my Maybe[Str] $optional_name;`, `my Maybe[Int] $count = undef;`},
		},
		"Any": {
			Symbol:        "Any",
			Type:          "Any",
			Kind:          "type",
			Documentation: "Any type - accepts any value. Use sparingly when specific typing is not possible.",
			Examples:      []string{`my Any $data = get_data();`},
		},
	}

	if info, exists := builtinTypes[typeName]; exists {
		return &info
	}

	return nil
}

// getBuiltinFunctionInfo returns information about builtin functions
func (s *TypeQueryService) getBuiltinFunctionInfo(funcName string) *TypeInfo {
	builtinFunctions := map[string]TypeInfo{
		"print": {
			Symbol:        "print",
			Type:          "function",
			Kind:          "function",
			Documentation: "Print values to STDOUT without adding a newline.",
			Signature: &FunctionSig{
				Parameters: []ParameterInfo{
					{Name: "list", Type: "List", Optional: true, Description: "Values to print"},
				},
				ReturnType: "Int",
			},
			Examples: []string{`print "Hello";`, `print $name, " is ", $age, " years old";`},
		},
		"say": {
			Symbol:        "say",
			Type:          "function",
			Kind:          "function",
			Documentation: "Print values to STDOUT and add a newline.",
			Signature: &FunctionSig{
				Parameters: []ParameterInfo{
					{Name: "list", Type: "List", Optional: true, Description: "Values to print"},
				},
				ReturnType: "Int",
			},
			Examples: []string{`say "Hello World";`, `say "Name: $name";`},
		},
		"defined": {
			Symbol:        "defined",
			Type:          "function",
			Kind:          "function",
			Documentation: "Test whether a value is defined (not undef).",
			Signature: &FunctionSig{
				Parameters: []ParameterInfo{
					{Name: "value", Type: "Any", Description: "Value to test"},
				},
				ReturnType: "Bool",
			},
			Examples: []string{`if (defined $var) { ... }`, `my Bool $has_value = defined($data);`},
		},
		"ref": {
			Symbol:        "ref",
			Type:          "function",
			Kind:          "function",
			Documentation: "Return the reference type of a value.",
			Signature: &FunctionSig{
				Parameters: []ParameterInfo{
					{Name: "value", Type: "Any", Description: "Value to check"},
				},
				ReturnType: "Str",
			},
			Examples: []string{`my Str $type = ref($data);`, `if (ref($var) eq 'ARRAY') { ... }`},
		},
		"length": {
			Symbol:        "length",
			Type:          "function",
			Kind:          "function",
			Documentation: "Return the length of a string or array.",
			Signature: &FunctionSig{
				Parameters: []ParameterInfo{
					{Name: "value", Type: "Str|ArrayRef", Description: "String or array to measure"},
				},
				ReturnType: "Int",
			},
			Examples: []string{`my Int $len = length($string);`, `my Int $size = length(@array);`},
		},
	}

	if info, exists := builtinFunctions[funcName]; exists {
		return &info
	}

	return nil
}

// getKeywordInfo returns information about Perl keywords
func (s *TypeQueryService) getKeywordInfo(keyword string) *TypeInfo {
	keywords := map[string]TypeInfo{
		"my": {
			Symbol:        "my",
			Type:          "keyword",
			Kind:          "keyword",
			Documentation: "Declare a lexically scoped variable. The variable is only visible within the current block.",
			Examples:      []string{`my $var = "value";`, `my Int $count = 0;`, `my (@array, %hash);`},
		},
		"our": {
			Symbol:        "our",
			Type:          "keyword",
			Kind:          "keyword",
			Documentation: "Declare a package variable. Creates an alias to a package global.",
			Examples:      []string{`our $VERSION = "1.0";`, `our Int $global_counter;`},
		},
		"sub": {
			Symbol:        "sub",
			Type:          "keyword",
			Kind:          "keyword",
			Documentation: "Define a subroutine (function). Can include type annotations for parameters and return value.",
			Examples:      []string{`sub hello { say "Hello"; }`, `sub Int add(Int $a, Int $b) -> Int { return $a + $b; }`},
		},
		"if": {
			Symbol:        "if",
			Type:          "keyword",
			Kind:          "keyword",
			Documentation: "Conditional statement. Executes code block if condition is true.",
			Examples:      []string{`if ($condition) { ... }`, `if (defined $var) { print $var; }`},
		},
		"while": {
			Symbol:        "while",
			Type:          "keyword",
			Kind:          "keyword",
			Documentation: "Loop while condition is true. Executes code block repeatedly.",
			Examples:      []string{`while ($i < 10) { $i++; }`, `while (my $line = <$fh>) { ... }`},
		},
		"for": {
			Symbol:        "for",
			Type:          "keyword",
			Kind:          "keyword",
			Documentation: "C-style for loop or foreach loop over a list.",
			Examples:      []string{`for my $i (1..10) { ... }`, `for my Str $name (@names) { ... }`},
		},
		"use": {
			Symbol:        "use",
			Type:          "keyword",
			Kind:          "keyword",
			Documentation: "Load and import a module. Can also enable language features.",
			Examples:      []string{`use strict;`, `use warnings;`, `use List::Util qw(sum);`},
		},
	}

	if info, exists := keywords[keyword]; exists {
		return &info
	}

	return nil
}

// symbolMatchesAnnotation checks if a symbol matches a type annotation
func (s *TypeQueryService) symbolMatchesAnnotation(symbol string, annotation *ast.TypeAnnotation) bool {
	return strings.Contains(annotation.AnnotatedItem, symbol)
}

// createTypeInfoFromAnnotation creates type info from a parser annotation
func (s *TypeQueryService) createTypeInfoFromAnnotation(annotation *ast.TypeAnnotation) *TypeInfo {
	if annotation == nil {
		return nil
	}

	info := &TypeInfo{
		Symbol: annotation.AnnotatedItem,
		Type:   annotation.TypeExpression.String(),
	}

	switch annotation.Kind {
	case ast.VarAnnotation:
		info.Kind = "variable"
		info.Documentation = fmt.Sprintf("Variable %s of type %s", annotation.AnnotatedItem, info.Type)
	case ast.SubParamAnnotation:
		info.Kind = "parameter"
		info.Documentation = fmt.Sprintf("Parameter %s of type %s", annotation.AnnotatedItem, info.Type)
	case ast.SubReturnAnnotation:
		info.Kind = "return"
		info.Documentation = fmt.Sprintf("Return type %s", info.Type)
	case ast.MethodParamAnnotation:
		info.Kind = "parameter"
		info.Documentation = fmt.Sprintf("Method parameter %s of type %s", annotation.AnnotatedItem, info.Type)
	case ast.MethodReturnAnnotation:
		info.Kind = "return"
		info.Documentation = fmt.Sprintf("Method return type %s", info.Type)
	case ast.FieldAnnotation:
		info.Kind = "attribute"
		info.Documentation = fmt.Sprintf("Attribute %s of type %s", annotation.AnnotatedItem, info.Type)
	case ast.TypeDeclAnnotation:
		info.Kind = "type"
		info.Documentation = fmt.Sprintf("Type declaration %s", annotation.AnnotatedItem)
	}

	return info
}

// createDetailedTypeInfo creates detailed type info from an annotation
func (s *TypeQueryService) createDetailedTypeInfo(annotation *ast.TypeAnnotation) *TypeInfo {
	info := s.createTypeInfoFromAnnotation(annotation)
	if info == nil {
		return nil
	}

	// Add location information
	info.Location = &Location{
		URI: "", // Would be filled in from document context
		Range: Range{
			Start: Position{Line: 0, Character: 0}, // Would be filled from AST position info
			End:   Position{Line: 0, Character: len(annotation.AnnotatedItem)},
		},
	}

	// Add examples based on type
	info.Examples = s.generateExamplesForType(info.Type)

	return info
}

// inferTypeFromUsage attempts to infer type from code usage patterns
func (s *TypeQueryService) inferTypeFromUsage(doc *Document, symbol string) *TypeInfo {
	// Simple pattern matching for type inference
	lines := strings.Split(doc.Text, "\n")

	for _, line := range lines {
		if strings.Contains(line, symbol) {
			// Look for assignment patterns
			if strings.Contains(line, symbol+"=") {
				if strings.Contains(line, `"`) || strings.Contains(line, `'`) {
					return &TypeInfo{
						Symbol:        symbol,
						Type:          "Str",
						Kind:          "variable",
						Documentation: fmt.Sprintf("Inferred string variable %s", symbol),
					}
				}
				if strings.Contains(line, "[") && strings.Contains(line, "]") {
					return &TypeInfo{
						Symbol:        symbol,
						Type:          "ArrayRef",
						Kind:          "variable",
						Documentation: fmt.Sprintf("Inferred array reference %s", symbol),
					}
				}
				if strings.Contains(line, "{") && strings.Contains(line, "}") {
					return &TypeInfo{
						Symbol:        symbol,
						Type:          "HashRef",
						Kind:          "variable",
						Documentation: fmt.Sprintf("Inferred hash reference %s", symbol),
					}
				}
			}
		}
	}

	return nil
}

// getBuiltinSymbols returns a list of commonly used builtin symbols
func (s *TypeQueryService) getBuiltinSymbols() []TypeInfo {
	symbols := []TypeInfo{}

	// Add builtin types
	for _, typeName := range []string{"Str", "Int", "Num", "Bool", "ArrayRef", "HashRef", "Maybe", "Any"} {
		if info := s.getBuiltinTypeInfo(typeName); info != nil {
			symbols = append(symbols, *info)
		}
	}

	// Add common functions
	for _, funcName := range []string{"print", "say", "defined", "ref", "length"} {
		if info := s.getBuiltinFunctionInfo(funcName); info != nil {
			symbols = append(symbols, *info)
		}
	}

	return symbols
}

// generateExamplesForType generates usage examples for a given type
func (s *TypeQueryService) generateExamplesForType(typeName string) []string {
	examples := map[string][]string{
		"Str":      {`my Str $name = "Alice";`, `my Str $message = "Hello " . $name;`},
		"Int":      {`my Int $count = 42;`, `my Int $result = $count + 10;`},
		"Num":      {`my Num $price = 19.99;`, `my Num $tax = $price * 0.08;`},
		"Bool":     {`my Bool $enabled = 1;`, `my Bool $valid = $count > 0;`},
		"ArrayRef": {`my ArrayRef[Str] $names = ["Alice", "Bob"];`, `my ArrayRef[Int] $nums = [1, 2, 3];`},
		"HashRef":  {`my HashRef[Str] $config = {key => "value"};`, `my HashRef[Int] $scores = {alice => 95};`},
		"Maybe":    {`my Maybe[Str] $optional;`, `$optional = "value" if $condition;`},
	}

	if ex, exists := examples[typeName]; exists {
		return ex
	}

	return []string{fmt.Sprintf("my %s $variable;", typeName)}
}

// isSymbolChar determines if a character is part of a symbol
func isSymbolChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_' || c == '$' || c == '@' || c == '%' || c == ':'
}
