// ABOUTME: Integration module connecting enhanced diagnostics with PSC type checker and error formatter
// ABOUTME: Provides unified interface for symbol-aware error reporting throughout PSC

package diagnostics

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/typechecker"
	"tamarou.com/pvm/internal/typedef"
)

// EnhancedTypeChecker wraps the existing type checker with enhanced diagnostics
type EnhancedTypeChecker struct {
	typeChecker      *typechecker.TypeChecker
	binder           *binder.DefaultBinder
	diagnosticEngine *EnhancedDiagnosticEngine
	enabledChecks    CheckFlags
}

// CheckFlags controls which enhanced checks are enabled
type CheckFlags struct {
	UndefinedVariables bool
	ShadowedVariables  bool
	UnusedVariables    bool
	TypeMismatches     bool
	SymbolSuggestions  bool
}

// DefaultCheckFlags returns the default set of enabled checks
func DefaultCheckFlags() CheckFlags {
	return CheckFlags{
		UndefinedVariables: true,
		ShadowedVariables:  true,
		UnusedVariables:    true,
		TypeMismatches:     true,
		SymbolSuggestions:  true,
	}
}

// NewEnhancedTypeChecker creates a type checker with enhanced diagnostics
func NewEnhancedTypeChecker(hierarchy *typedef.TypeHierarchy, moduleName string, filePath string) *EnhancedTypeChecker {
	typeChecker := typechecker.NewTypeCheckerLegacy(hierarchy, moduleName)
	binderInstance := binder.NewBinder() // Create default binder

	return &EnhancedTypeChecker{
		typeChecker:   typeChecker,
		binder:        binderInstance,
		enabledChecks: DefaultCheckFlags(),
	}
}

// CheckASTWithEnhancedDiagnostics performs type checking with enhanced symbol-aware diagnostics
func (etc *EnhancedTypeChecker) CheckASTWithEnhancedDiagnostics(astRoot *ast.AST) (*EnhancedTypeCheckResult, error) {
	// Step 1: Perform symbol binding
	symbolTable, err := etc.binder.BindAST(astRoot)
	if err != nil {
		return nil, fmt.Errorf("symbol binding failed: %w", err)
	}

	// Step 2: Update type checker with symbol information
	enhancedTypeChecker := typechecker.NewTypeChecker(etc.typeChecker.Hierarchy, symbolTable, astRoot.Path)

	// Step 3: Perform traditional type checking
	typeErrors := enhancedTypeChecker.CheckAST(astRoot)

	// Step 4: Initialize enhanced diagnostic engine
	etc.diagnosticEngine = NewEnhancedDiagnosticEngine(symbolTable, astRoot.Path, astRoot.Source)

	// Step 5: Perform enhanced symbol analysis
	var enhancedDiagnostics []Diagnostic
	if etc.enabledChecks.UndefinedVariables || etc.enabledChecks.ShadowedVariables ||
		etc.enabledChecks.UnusedVariables || etc.enabledChecks.TypeMismatches {
		enhancedDiagnostics = etc.diagnosticEngine.AnalyzeSymbols(astRoot.Root)
	}

	// Step 6: Convert traditional type errors to enhanced diagnostics
	enhancedTypeErrors := etc.convertTypeErrorsToEnhancedDiagnostics(typeErrors, symbolTable)

	// Step 7: Combine all diagnostics
	allDiagnostics := make([]Diagnostic, 0, len(enhancedDiagnostics)+len(enhancedTypeErrors))
	allDiagnostics = append(allDiagnostics, enhancedDiagnostics...)
	allDiagnostics = append(allDiagnostics, enhancedTypeErrors...)

	return &EnhancedTypeCheckResult{
		OriginalErrors:      typeErrors,
		EnhancedDiagnostics: allDiagnostics,
		SymbolTable:         symbolTable,
		TypeAnnotations:     astRoot.TypeAnnotations,
		Path:                astRoot.Path,
	}, nil
}

// EnhancedTypeCheckResult contains both traditional and enhanced diagnostic results
type EnhancedTypeCheckResult struct {
	OriginalErrors      []error
	EnhancedDiagnostics []Diagnostic
	SymbolTable         *binder.SymbolTable
	TypeAnnotations     []*ast.TypeAnnotation
	Path                string
}

// convertTypeErrorsToEnhancedDiagnostics converts traditional type errors to enhanced diagnostics
func (etc *EnhancedTypeChecker) convertTypeErrorsToEnhancedDiagnostics(typeErrors []error, symbolTable *binder.SymbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	for _, err := range typeErrors {
		// Try to extract position and context from error
		diagnostic := etc.createEnhancedDiagnosticFromError(err, symbolTable)
		diagnostics = append(diagnostics, diagnostic)
	}

	return diagnostics
}

// createEnhancedDiagnosticFromError creates an enhanced diagnostic from a traditional error
func (etc *EnhancedTypeChecker) createEnhancedDiagnosticFromError(err error, symbolTable *binder.SymbolTable) Diagnostic {
	errStr := err.Error()

	// Parse error for position information
	pos := ast.Position{Line: 1, Column: 1} // Default position
	if typedErr, ok := err.(interface{ Location() string }); ok {
		pos = etc.parseLocationFromError(typedErr.Location())
	}

	// Extract variable name if this is a variable-related error
	symbolName := etc.extractSymbolNameFromError(errStr)
	var symbol *binder.Symbol
	if symbolName != "" {
		symbol = symbolTable.ResolveSymbol(symbolName, binder.SymbolKind(-1))
	}

	diagnostic := Diagnostic{
		Kind:        DiagnosticError,
		Message:     errStr,
		Pos:         pos,
		Symbol:      symbol,
		SymbolName:  symbolName,
		FilePath:    etc.diagnosticEngine.filePath,
		LineText:    etc.diagnosticEngine.getLineText(pos.Line),
		Suggestion:  etc.generateSuggestionFromError(errStr, symbol),
		Code:        etc.generateErrorCode(errStr),
		HelpMessage: etc.generateHelpMessage(errStr),
	}

	// Add symbol context if available
	if symbol != nil {
		diagnostic.SymbolKind = symbol.Kind
		diagnostic.HelpMessage = fmt.Sprintf("Symbol '%s' declared at line %d", symbolName, symbol.Position.Line)
	}

	return diagnostic
}

// parseLocationFromError extracts position from error location string
func (etc *EnhancedTypeChecker) parseLocationFromError(location string) ast.Position {
	// Parse location strings like "file.pl:10:5"
	parts := strings.Split(location, ":")
	if len(parts) >= 3 {
		var line, col int
		fmt.Sscanf(parts[1], "%d", &line)
		fmt.Sscanf(parts[2], "%d", &col)
		return ast.Position{Line: line, Column: col}
	}
	return ast.Position{Line: 1, Column: 1}
}

// extractSymbolNameFromError attempts to extract a symbol name from error message
func (etc *EnhancedTypeChecker) extractSymbolNameFromError(errStr string) string {
	// Look for variable patterns in error messages
	if strings.Contains(errStr, "$") {
		words := strings.Fields(errStr)
		for _, word := range words {
			if strings.HasPrefix(word, "$") {
				// Clean up the variable name (remove punctuation)
				return strings.Trim(word, ".,;:!?'\"")
			}
		}
	}
	return ""
}

// generateSuggestionFromError creates helpful suggestions based on error content
func (etc *EnhancedTypeChecker) generateSuggestionFromError(errStr string, symbol *binder.Symbol) string {
	errLower := strings.ToLower(errStr)

	switch {
	case strings.Contains(errLower, "type mismatch"):
		return "Check that the assigned value matches the declared type"
	case strings.Contains(errLower, "undefined"):
		if symbol != nil {
			return fmt.Sprintf("Variable '%s' was declared at line %d", symbol.Name, symbol.Position.Line)
		}
		return "Make sure the variable is declared before use"
	case strings.Contains(errLower, "incompatible"):
		return "Review type compatibility rules or add explicit type conversion"
	default:
		return "Review the type annotation and value assignment"
	}
}

// generateErrorCode assigns error codes based on error content
func (etc *EnhancedTypeChecker) generateErrorCode(errStr string) string {
	errLower := strings.ToLower(errStr)

	switch {
	case strings.Contains(errLower, "undefined"):
		return "PSC-E001"
	case strings.Contains(errLower, "type mismatch"):
		return "PSC-E002"
	case strings.Contains(errLower, "incompatible"):
		return "PSC-E003"
	case strings.Contains(errLower, "annotation"):
		return "PSC-E004"
	default:
		return "PSC-E000"
	}
}

// generateHelpMessage creates contextual help messages
func (etc *EnhancedTypeChecker) generateHelpMessage(errStr string) string {
	errLower := strings.ToLower(errStr)

	switch {
	case strings.Contains(errLower, "type mismatch"):
		return "Type annotations in Perl specify the expected type of values"
	case strings.Contains(errLower, "undefined"):
		return "All variables must be declared with 'my', 'our', or 'state'"
	case strings.Contains(errLower, "annotation"):
		return "Type annotation syntax: my TypeName $variable = value;"
	default:
		return "See Typed Perl documentation for more information"
	}
}

// SetEnabledChecks configures which enhanced checks to perform
func (etc *EnhancedTypeChecker) SetEnabledChecks(flags CheckFlags) {
	etc.enabledChecks = flags
}

// GetSymbolTable returns the symbol table from the last analysis
func (etc *EnhancedTypeChecker) GetSymbolTable() *binder.SymbolTable {
	if etc.diagnosticEngine != nil {
		return etc.diagnosticEngine.symbolTable
	}
	return nil
}

// FormatDiagnostics formats all diagnostics for display
func (etc *EnhancedTypeChecker) FormatDiagnostics(result *EnhancedTypeCheckResult, colorEnabled bool) string {
	var output strings.Builder

	for i, diagnostic := range result.EnhancedDiagnostics {
		if i > 0 {
			output.WriteString("\n")
		}
		output.WriteString(diagnostic.FormatDiagnostic(colorEnabled))
	}

	return output.String()
}

// Helper function to create a type checker with symbols
// This would normally be in the typechecker package, but we'll add it here for integration
func NewTypeCheckerWithSymbols(hierarchy *typedef.TypeHierarchy, symbolTable *binder.SymbolTable, moduleName string) *typechecker.TypeChecker {
	// For now, create a regular type checker
	// In a full implementation, this would be enhanced to use symbol information
	return typechecker.NewTypeChecker(hierarchy, symbolTable, moduleName)
}
