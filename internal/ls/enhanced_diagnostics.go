// ABOUTME: Enhanced diagnostics with type context and intelligent suggestions
// ABOUTME: Provides rich error messages and quick fixes for type-related issues

package ls

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
)

// EnhancedDiagnostic provides rich diagnostic information
type EnhancedDiagnostic struct {
	Diagnostic
	TypeContext     *TypeDiagnosticContext
	Suggestions     []DiagnosticSuggestion
	RelatedInfo     []RelatedInformation
	CodeDescription *CodeDescription
	Data            interface{} // Custom data for code actions
}

// TypeDiagnosticContext provides type-specific diagnostic context
type TypeDiagnosticContext struct {
	ExpectedType   string
	ActualType     string
	TypeMismatch   bool
	InferredTypes  []string
	TypeHierarchy  []string // Type inheritance chain
	AvailableTypes []string // Types available in scope
}

// DiagnosticSuggestion provides a suggested fix
type DiagnosticSuggestion struct {
	Title       string
	Description string
	Edit        *TextEdit
	Priority    SuggestionPriority
}

// SuggestionPriority indicates the relevance of a suggestion
type SuggestionPriority int

const (
	SuggestionPriorityHigh   SuggestionPriority = 1
	SuggestionPriorityMedium SuggestionPriority = 2
	SuggestionPriorityLow    SuggestionPriority = 3
)

// RelatedInformation provides additional context
type RelatedInformation struct {
	Location Location
	Message  string
}

// CodeDescription provides a link to documentation
type CodeDescription struct {
	Href string // URL to documentation
}

// GetEnhancedDiagnostics converts type check errors to enhanced diagnostics
func (ls *LanguageService) GetEnhancedDiagnostics(uri string) ([]EnhancedDiagnostic, error) {
	ls.mu.RLock()
	doc, exists := ls.documents[uri]
	ls.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	var diagnostics []EnhancedDiagnostic

	// Convert each type check error to enhanced diagnostic
	for _, err := range doc.Errors {
		diagnostic := ls.enhanceTypeCheckError(err, doc)
		diagnostics = append(diagnostics, diagnostic)
	}

	// Add additional semantic diagnostics
	semanticDiags := ls.performSemanticAnalysis(doc)
	diagnostics = append(diagnostics, semanticDiags...)

	return diagnostics, nil
}

// enhanceTypeCheckError converts a type check error to enhanced diagnostic
func (ls *LanguageService) enhanceTypeCheckError(err parser.TypeCheckError, doc *Document) EnhancedDiagnostic {
	diagnostic := EnhancedDiagnostic{
		Diagnostic: Diagnostic{
			Range: Range{
				Start: Position{
					Line:      err.Line - 1,
					Character: err.Column - 1,
				},
				End: Position{
					Line:      err.Line - 1,
					Character: err.Column - 1 + len(""),
				},
			},
			Message: err.Message,
		},
	}

	// Set severity based on error type
	severity := ls.mapErrorTypeToSeverity("type_error")
	diagnostic.Severity = &severity

	// Add type context
	diagnostic.TypeContext = ls.extractTypeContext(err, doc)

	// Generate suggestions based on error type
	diagnostic.Suggestions = ls.generateSuggestions(err, doc, diagnostic.TypeContext)

	// Add related information
	diagnostic.RelatedInfo = ls.findRelatedInformation(err, doc)

	// Add code description for documentation
	diagnostic.CodeDescription = ls.getErrorDocumentation("type_error")

	// Store error data for code actions
	diagnostic.Data = err

	return diagnostic
}

// mapErrorTypeToSeverity maps error types to LSP severity levels
func (ls *LanguageService) mapErrorTypeToSeverity(errorType string) DiagnosticSeverity {
	switch errorType {
	case "type_mismatch", "undefined_type", "incompatible_types":
		return DiagnosticSeverityError
	case "deprecated_syntax", "ambiguous_type":
		return DiagnosticSeverityWarning
	case "missing_type_annotation", "could_be_typed":
		return DiagnosticSeverityInformation
	case "style_suggestion":
		return DiagnosticSeverityHint
	default:
		return DiagnosticSeverityError
	}
}

// extractTypeContext analyzes type information around the error
func (ls *LanguageService) extractTypeContext(err parser.TypeCheckError, doc *Document) *TypeDiagnosticContext {
	context := &TypeDiagnosticContext{
		AvailableTypes: ls.getAvailableTypes(doc),
	}

	// Parse error message for type information
	if strings.Contains(err.Message, "expected") && strings.Contains(err.Message, "got") {
		// Extract expected and actual types from message
		// Example: "Type mismatch: expected Int, got Str"
		parts := strings.Split(err.Message, ":")
		if len(parts) > 1 {
			typePart := strings.TrimSpace(parts[1])
			if expected, actual, ok := ls.parseTypeMismatch(typePart); ok {
				context.ExpectedType = expected
				context.ActualType = actual
				context.TypeMismatch = true
			}
		}
	}

	// Find symbol at error position to get more context
	pos := Position{
		Line:      err.Line - 1,
		Character: err.Column - 1,
	}

	if symbol := ls.findSymbolAtPosition(doc, pos); symbol != nil {
		// Get inferred types for the symbol
		context.InferredTypes = ls.inferSymbolTypes(symbol, doc)

		// Get type hierarchy if applicable
		if symbol.Type != "" {
			context.TypeHierarchy = ls.getTypeHierarchy(symbol.Type, doc)
		}
	}

	return context
}

// generateSuggestions creates intelligent fix suggestions
func (ls *LanguageService) generateSuggestions(err parser.TypeCheckError, doc *Document, typeContext *TypeDiagnosticContext) []DiagnosticSuggestion {
	var suggestions []DiagnosticSuggestion

	switch "type_error" {
	case "undefined_variable":
		suggestions = append(suggestions, ls.generateVariableDeclarationSuggestions(err, doc)...)

	case "type_mismatch":
		if typeContext.TypeMismatch {
			suggestions = append(suggestions, ls.generateTypeMismatchSuggestions(err, doc, typeContext)...)
		}

	case "undefined_type":
		suggestions = append(suggestions, ls.generateTypeImportSuggestions(err, doc)...)

	case "missing_type_annotation":
		suggestions = append(suggestions, ls.generateTypeAnnotationSuggestions(err, doc, typeContext)...)

	case "deprecated_syntax":
		suggestions = append(suggestions, ls.generateModernizationSuggestions(err, doc)...)
	}

	// Add generic suggestions
	suggestions = append(suggestions, ls.generateGenericSuggestions(err, doc)...)

	return suggestions
}

// generateVariableDeclarationSuggestions creates fixes for undefined variables
func (ls *LanguageService) generateVariableDeclarationSuggestions(err parser.TypeCheckError, doc *Document) []DiagnosticSuggestion {
	var suggestions []DiagnosticSuggestion

	// Extract variable name from error
	varName := extractVariableName(err.Message)
	if varName == "" {
		return suggestions
	}

	// Suggest declaration with inferred type
	inferredType := ls.inferVariableType(varName, err.Line, err.Column, doc)

	suggestions = append(suggestions, DiagnosticSuggestion{
		Title:       fmt.Sprintf("Declare '%s' with type %s", varName, inferredType),
		Description: "Add typed variable declaration before first use",
		Edit: &TextEdit{
			Range: Range{
				Start: Position{Line: err.Line - 1, Character: 0},
				End:   Position{Line: err.Line - 1, Character: 0},
			},
			NewText: fmt.Sprintf("my %s %s;\n", inferredType, varName),
		},
		Priority: SuggestionPriorityHigh,
	})

	// Suggest declaration without type
	suggestions = append(suggestions, DiagnosticSuggestion{
		Title:       fmt.Sprintf("Declare '%s' without type", varName),
		Description: "Add untyped variable declaration",
		Edit: &TextEdit{
			Range: Range{
				Start: Position{Line: err.Line - 1, Character: 0},
				End:   Position{Line: err.Line - 1, Character: 0},
			},
			NewText: fmt.Sprintf("my %s;\n", varName),
		},
		Priority: SuggestionPriorityMedium,
	})

	// Check for typos - suggest similar variables
	if similar := ls.findSimilarVariables(varName, err.Line, err.Column, doc); len(similar) > 0 {
		for _, sim := range similar {
			suggestions = append(suggestions, DiagnosticSuggestion{
				Title:       fmt.Sprintf("Did you mean '%s'?", sim),
				Description: "Replace with existing variable",
				Edit: &TextEdit{
					Range: Range{
						Start: Position{
							Line:      err.Line - 1,
							Character: err.Column - 1,
						},
						End: Position{
							Line:      err.Line - 1,
							Character: err.Column - 1 + len(varName),
						},
					},
					NewText: sim,
				},
				Priority: SuggestionPriorityHigh,
			})
		}
	}

	return suggestions
}

// generateTypeMismatchSuggestions creates fixes for type mismatches
func (ls *LanguageService) generateTypeMismatchSuggestions(err parser.TypeCheckError, doc *Document, typeContext *TypeDiagnosticContext) []DiagnosticSuggestion {
	var suggestions []DiagnosticSuggestion

	// Suggest type conversion
	if converter := ls.findTypeConverter(typeContext.ActualType, typeContext.ExpectedType); converter != "" {
		suggestions = append(suggestions, DiagnosticSuggestion{
			Title:       fmt.Sprintf("Convert to %s using %s", typeContext.ExpectedType, converter),
			Description: "Add explicit type conversion",
			Edit: &TextEdit{
				Range: Range{
					Start: Position{
						Line:      err.Line - 1,
						Character: err.Column - 1,
					},
					End: Position{
						Line:      err.Line - 1,
						Character: err.Column - 1,
					},
				},
				NewText: converter + "(",
			},
			Priority: SuggestionPriorityHigh,
		})
	}

	// Suggest changing variable type
	suggestions = append(suggestions, DiagnosticSuggestion{
		Title:       fmt.Sprintf("Change variable type to %s", typeContext.ActualType),
		Description: "Update variable declaration to match usage",
		Priority:    SuggestionPriorityMedium,
		// Edit would need to find and update the declaration
	})

	// Suggest union type if appropriate
	if ls.canUseUnionType(typeContext) {
		unionType := fmt.Sprintf("%s|%s", typeContext.ExpectedType, typeContext.ActualType)
		suggestions = append(suggestions, DiagnosticSuggestion{
			Title:       fmt.Sprintf("Use union type %s", unionType),
			Description: "Allow both types with a union",
			Priority:    SuggestionPriorityLow,
		})
	}

	return suggestions
}

// generateTypeAnnotationSuggestions creates type annotation suggestions
func (ls *LanguageService) generateTypeAnnotationSuggestions(err parser.TypeCheckError, doc *Document, typeContext *TypeDiagnosticContext) []DiagnosticSuggestion {
	var suggestions []DiagnosticSuggestion

	// Find the declaration that needs annotation
	pos := Position{
		Line:      err.Line - 1,
		Character: err.Column - 1,
	}

	symbol := ls.findSymbolAtPosition(doc, pos)
	if symbol == nil {
		return suggestions
	}

	// Suggest annotations based on inferred types
	for _, inferredType := range typeContext.InferredTypes {
		suggestions = append(suggestions, DiagnosticSuggestion{
			Title:       fmt.Sprintf("Add type annotation: %s", inferredType),
			Description: fmt.Sprintf("Annotate %s with inferred type %s", symbol.Name, inferredType),
			Priority:    SuggestionPriorityMedium,
			// Edit would insert the type annotation
		})
	}

	// Suggest common types
	commonTypes := []string{"Str", "Int", "Num", "Bool", "ArrayRef", "HashRef"}
	for _, commonType := range commonTypes {
		if !contains(typeContext.InferredTypes, commonType) {
			suggestions = append(suggestions, DiagnosticSuggestion{
				Title:       fmt.Sprintf("Add type annotation: %s", commonType),
				Description: "Add common type annotation",
				Priority:    SuggestionPriorityLow,
			})
		}
	}

	return suggestions
}

// performSemanticAnalysis performs additional semantic checks
func (ls *LanguageService) performSemanticAnalysis(doc *Document) []EnhancedDiagnostic {
	var diagnostics []EnhancedDiagnostic

	if doc.AST == nil || doc.SymbolTable == nil {
		return diagnostics
	}

	// Check for unused variables
	unusedDiags := ls.checkUnusedVariables(doc)
	diagnostics = append(diagnostics, unusedDiags...)

	// Check for type safety improvements
	typeSafetyDiags := ls.checkTypeSafetyImprovements(doc)
	diagnostics = append(diagnostics, typeSafetyDiags...)

	// Check for deprecated patterns
	deprecationDiags := ls.checkDeprecatedPatterns(doc)
	diagnostics = append(diagnostics, deprecationDiags...)

	return diagnostics
}

// checkUnusedVariables finds variables that are declared but never used
func (ls *LanguageService) checkUnusedVariables(doc *Document) []EnhancedDiagnostic {
	var diagnostics []EnhancedDiagnostic

	// Track variable usage
	usage := make(map[*binder.Symbol]int)

	// Walk through all symbols
	ls.collectSymbolsFromScope(doc.SymbolTable.GlobalScope, &[]*binder.Symbol{})

	// Count references (simplified - real implementation would use AST)
	// ...

	// Report unused variables
	for symbol, count := range usage {
		if count == 0 && symbol.Kind == binder.SymbolScalar {
			diagnostic := EnhancedDiagnostic{
				Diagnostic: Diagnostic{
					Range:   ls.getSymbolRange(symbol),
					Message: fmt.Sprintf("Variable '%s' is declared but never used", symbol.Name),
				},
			}

			severity := DiagnosticSeverityWarning
			diagnostic.Severity = &severity

			// Add suggestion to remove
			diagnostic.Suggestions = []DiagnosticSuggestion{
				{
					Title:       "Remove unused variable",
					Description: "Delete the variable declaration",
					Priority:    SuggestionPriorityMedium,
				},
			}

			diagnostics = append(diagnostics, diagnostic)
		}
	}

	return diagnostics
}

// checkTypeSafetyImprovements suggests where types could improve safety
func (ls *LanguageService) checkTypeSafetyImprovements(doc *Document) []EnhancedDiagnostic {
	var diagnostics []EnhancedDiagnostic

	// Find untyped variables that could benefit from types
	var symbols []*binder.Symbol
	ls.collectSymbolsFromScope(doc.SymbolTable.GlobalScope, &symbols)

	for _, symbol := range symbols {
		if symbol.Type == "" && ls.couldBenefitFromType(symbol, doc) {
			diagnostic := EnhancedDiagnostic{
				Diagnostic: Diagnostic{
					Range:   ls.getSymbolRange(symbol),
					Message: fmt.Sprintf("Variable '%s' could benefit from type annotation", symbol.Name),
				},
			}

			severity := DiagnosticSeverityInformation
			diagnostic.Severity = &severity

			// Infer and suggest types
			inferredTypes := ls.inferSymbolTypes(symbol, doc)
			for _, inferredType := range inferredTypes {
				diagnostic.Suggestions = append(diagnostic.Suggestions, DiagnosticSuggestion{
					Title:       fmt.Sprintf("Add type: %s", inferredType),
					Description: "Type annotation improves safety and documentation",
					Priority:    SuggestionPriorityLow,
				})
			}

			diagnostics = append(diagnostics, diagnostic)
		}
	}

	return diagnostics
}

// Helper methods

func (ls *LanguageService) parseTypeMismatch(text string) (expected, actual string, ok bool) {
	// Parse "expected Int, got Str" format
	if strings.Contains(text, "expected") && strings.Contains(text, "got") {
		parts := strings.Split(text, ",")
		if len(parts) == 2 {
			expected = strings.TrimSpace(strings.TrimPrefix(parts[0], "expected"))
			actual = strings.TrimSpace(strings.TrimPrefix(parts[1], "got"))
			ok = true
		}
	}
	return
}

func (ls *LanguageService) inferVariableType(varName string, line, column int, doc *Document) string {
	// Analyze usage context to infer type
	// In a real implementation, this would analyze AST context

	// Default to common type based on variable name patterns
	if strings.HasSuffix(varName, "_count") || strings.HasSuffix(varName, "_id") {
		return "Int"
	}
	if strings.HasSuffix(varName, "_name") || strings.HasSuffix(varName, "_str") {
		return "Str"
	}
	if strings.HasSuffix(varName, "_list") || strings.HasSuffix(varName, "_array") {
		return "ArrayRef"
	}
	if strings.HasSuffix(varName, "_map") || strings.HasSuffix(varName, "_hash") {
		return "HashRef"
	}

	return "Any"
}

func (ls *LanguageService) findSimilarVariables(varName string, line, column int, doc *Document) []string {
	var similar []string

	// Get all visible variables
	scope := ls.findScopeAtPosition(doc.SymbolTable, Position{
		Line:      line - 1,
		Character: column - 1,
	})

	if scope != nil {
		for _, symbol := range scope.Symbols {
			if ls.isVariableSymbol(symbol.Kind) {
				similarity := ls.calculateSimilarity(varName, symbol.Name)
				if similarity > 0.7 { // 70% similarity threshold
					similar = append(similar, ls.formatSymbolForCompletion(symbol))
				}
			}
		}
	}

	return similar
}

func (ls *LanguageService) calculateSimilarity(s1, s2 string) float64 {
	// Simple Levenshtein distance-based similarity
	// In a real implementation, use proper string similarity algorithm
	if s1 == s2 {
		return 1.0
	}

	// Very simple approximation
	common := 0
	for i := 0; i < len(s1) && i < len(s2); i++ {
		if s1[i] == s2[i] {
			common++
		}
	}

	return float64(common) / float64(max(len(s1), len(s2)))
}

func (ls *LanguageService) findTypeConverter(from, to string) string {
	// Map of type conversions
	converters := map[string]map[string]string{
		"Str": {
			"Int": "int",
			"Num": "0+",
		},
		"Int": {
			"Str": "\"\" .",
			"Num": "0+",
		},
		"Num": {
			"Int": "int",
			"Str": "\"\" .",
		},
	}

	if fromMap, ok := converters[from]; ok {
		if converter, ok := fromMap[to]; ok {
			return converter
		}
	}

	return ""
}

func (ls *LanguageService) canUseUnionType(context *TypeDiagnosticContext) bool {
	// Check if union type makes sense
	// Avoid suggesting unions for completely unrelated types
	unrelatedPairs := [][2]string{
		{"Int", "HashRef"},
		{"Str", "ArrayRef"},
		{"Bool", "CodeRef"},
	}

	for _, pair := range unrelatedPairs {
		if (context.ExpectedType == pair[0] && context.ActualType == pair[1]) ||
			(context.ExpectedType == pair[1] && context.ActualType == pair[0]) {
			return false
		}
	}

	return true
}

func (ls *LanguageService) inferSymbolTypes(symbol *binder.Symbol, doc *Document) []string {
	// Analyze symbol usage to infer possible types
	// In a real implementation, this would use data flow analysis

	var types []string

	// Basic inference based on symbol kind
	switch symbol.Kind {
	case binder.SymbolScalar:
		types = append(types, "Str", "Int", "Num")
	case binder.SymbolArray:
		types = append(types, "ArrayRef")
	case binder.SymbolHash:
		types = append(types, "HashRef")
	}

	return types
}

func (ls *LanguageService) getTypeHierarchy(typeName string, doc *Document) []string {
	// Return type inheritance hierarchy
	// In a real implementation, this would use class/role analysis

	hierarchy := []string{typeName}

	// Add common parent types
	switch typeName {
	case "Int", "Num", "Str":
		hierarchy = append(hierarchy, "Value", "Any")
	case "ArrayRef", "HashRef":
		hierarchy = append(hierarchy, "Ref", "Any")
	}

	return hierarchy
}

func (ls *LanguageService) getAvailableTypes(doc *Document) []string {
	// Collect all available type names in scope
	types := []string{
		// Built-in types
		"Any", "Undef", "Str", "Int", "Num", "Bool",
		"ArrayRef", "HashRef", "CodeRef", "RegexpRef",
		"ScalarRef", "GlobRef", "Object",
	}

	// Add user-defined types from symbol table
	if doc.SymbolTable != nil {
		var symbols []*binder.Symbol
		ls.collectSymbolsFromScope(doc.SymbolTable.GlobalScope, &symbols)

		for _, symbol := range symbols {
			if symbol.Kind == binder.SymbolType {
				types = append(types, symbol.Name)
			}
		}
	}

	return types
}

func (ls *LanguageService) getErrorDocumentation(errorType string) *CodeDescription {
	// Map error types to documentation URLs
	docURLs := map[string]string{
		"type_mismatch":           "https://docs.perl.org/types#type-mismatch",
		"undefined_type":          "https://docs.perl.org/types#undefined-types",
		"missing_type_annotation": "https://docs.perl.org/types#annotations",
	}

	if url, ok := docURLs[errorType]; ok {
		return &CodeDescription{Href: url}
	}

	return nil
}

func (ls *LanguageService) couldBenefitFromType(symbol *binder.Symbol, doc *Document) bool {
	// Determine if a symbol would benefit from type annotation

	// Functions always benefit from type annotations
	if symbol.Kind == binder.SymbolSubroutine || symbol.Kind == binder.SymbolMethod {
		return true
	}

	// Public/exported symbols benefit from types
	if symbol.Flags&binder.SymbolFlagExported != 0 {
		return true
	}

	// Complex data structures benefit from types
	if symbol.Kind == binder.SymbolArray || symbol.Kind == binder.SymbolHash {
		return true
	}

	return false
}

func (ls *LanguageService) checkDeprecatedPatterns(doc *Document) []EnhancedDiagnostic {
	var diagnostics []EnhancedDiagnostic

	// Check for old-style variable declarations, bareword filehandles, etc.
	// In a real implementation, this would analyze the AST

	return diagnostics
}

func (ls *LanguageService) generateModernizationSuggestions(err parser.TypeCheckError, doc *Document) []DiagnosticSuggestion {
	var suggestions []DiagnosticSuggestion

	// Suggest modern Perl patterns
	// In a real implementation, this would provide specific modernization advice

	return suggestions
}

func (ls *LanguageService) generateTypeImportSuggestions(err parser.TypeCheckError, doc *Document) []DiagnosticSuggestion {
	var suggestions []DiagnosticSuggestion

	// Suggest importing modules that define the missing type
	// In a real implementation, this would search available modules

	return suggestions
}

func (ls *LanguageService) generateGenericSuggestions(err parser.TypeCheckError, doc *Document) []DiagnosticSuggestion {
	var suggestions []DiagnosticSuggestion

	// Add generic helpful suggestions
	suggestions = append(suggestions, DiagnosticSuggestion{
		Title:       "View documentation",
		Description: "Learn more about this error type",
		Priority:    SuggestionPriorityLow,
	})

	return suggestions
}

func (ls *LanguageService) findRelatedInformation(err parser.TypeCheckError, doc *Document) []RelatedInformation {
	var related []RelatedInformation

	// Find related declarations, usages, etc.
	// In a real implementation, this would search the AST

	return related
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
