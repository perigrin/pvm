// ABOUTME: Enhanced diagnostics system leveraging symbol information for better error reporting
// ABOUTME: Provides symbol-aware error messages, undefined variable detection, and usage tracking

//go:generate go run ../../scripts/generate_diagnostics.go ../../scripts/diagnostic_definitions.json

package diagnostics

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
)

// DiagnosticKind represents the type of diagnostic
type DiagnosticKind int

const (
	DiagnosticError DiagnosticKind = iota
	DiagnosticWarning
	DiagnosticInfo
	DiagnosticHint
)

// String returns the string representation of diagnostic kind
func (dk DiagnosticKind) String() string {
	switch dk {
	case DiagnosticError:
		return "error"
	case DiagnosticWarning:
		return "warning"
	case DiagnosticInfo:
		return "info"
	case DiagnosticHint:
		return "hint"
	default:
		return "unknown"
	}
}

// Diagnostic represents an enhanced diagnostic with symbol context
type Diagnostic struct {
	// Basic diagnostic information
	Kind    DiagnosticKind
	Message string
	Pos     ast.Position

	// Symbol context
	Symbol     *binder.Symbol
	SymbolName string
	SymbolKind binder.SymbolKind

	// Location context
	FilePath string
	LineText string

	// Enhancement context
	Suggestion     string
	RelatedSymbols []*binder.Symbol
	DidYouMean     []string

	// Code information
	Code        string
	HelpMessage string
}

// EnhancedDiagnosticEngine generates enhanced diagnostics using symbol information
type EnhancedDiagnosticEngine struct {
	symbolTable   *binder.SymbolTable
	usageTracker  *SymbolUsageTracker
	sourceContent string
	filePath      string
}

// NewEnhancedDiagnosticEngine creates a new enhanced diagnostic engine
func NewEnhancedDiagnosticEngine(symbolTable *binder.SymbolTable, filePath, sourceContent string) *EnhancedDiagnosticEngine {
	return &EnhancedDiagnosticEngine{
		symbolTable:   symbolTable,
		usageTracker:  NewSymbolUsageTracker(),
		sourceContent: sourceContent,
		filePath:      filePath,
	}
}

// AnalyzeSymbols performs comprehensive symbol analysis and generates diagnostics
func (engine *EnhancedDiagnosticEngine) AnalyzeSymbols(astRoot ast.Node) []Diagnostic {
	var diagnostics []Diagnostic

	// Track symbol usage first
	engine.usageTracker.TrackUsage(astRoot, engine.symbolTable)

	// Check for undefined variables
	diagnostics = append(diagnostics, engine.findUndefinedVariables(astRoot)...)

	// Check for shadowed variables
	diagnostics = append(diagnostics, engine.findShadowedVariables()...)

	// Check for unused variables
	diagnostics = append(diagnostics, engine.findUnusedVariables()...)

	// Check for type mismatches with symbol context
	diagnostics = append(diagnostics, engine.findTypeMismatches(astRoot)...)

	return diagnostics
}

// findUndefinedVariables detects undefined variable references
func (engine *EnhancedDiagnosticEngine) findUndefinedVariables(astRoot ast.Node) []Diagnostic {
	var diagnostics []Diagnostic

	// Walk the AST to find variable references
	engine.walkAST(astRoot, func(node ast.Node) {
		if varExpr, ok := node.(*ast.VariableExpr); ok {
			varName := varExpr.FullName()

			// Check if variable is defined in symbol table
			symbol := engine.symbolTable.ResolveSymbol(varName, binder.SymbolKind(-1)) // -1 means any kind
			if symbol == nil {
				// Generate suggestions for similar variable names
				suggestions := engine.findSimilarVariables(varName)

				diagnostic := Diagnostic{
					Kind:        DiagnosticError,
					Message:     fmt.Sprintf("Undefined variable '%s'", varName),
					Pos:         varExpr.Start(),
					SymbolName:  varName,
					SymbolKind:  binder.SymbolScalar, // Use actual symbol kind
					FilePath:    engine.filePath,
					LineText:    engine.getLineText(varExpr.Start().Line),
					Suggestion:  engine.generateUndefinedVariableSuggestion(varName, suggestions),
					DidYouMean:  suggestions,
					Code:        "PSC-E001",
					HelpMessage: "Variables must be declared before use with 'my', 'our', or 'state'",
				}
				diagnostics = append(diagnostics, diagnostic)
			}
		}
	})

	return diagnostics
}

// findShadowedVariables detects variable shadowing across scopes
func (engine *EnhancedDiagnosticEngine) findShadowedVariables() []Diagnostic {
	var diagnostics []Diagnostic

	// Get all symbols from all scopes
	allSymbols := engine.symbolTable.GetVisibleSymbols()

	// Check for shadowing by comparing symbols with same name in different scopes
	symbolNames := make(map[string][]*binder.Symbol)
	for _, symbol := range allSymbols {
		if symbol.Kind == binder.SymbolScalar || symbol.Kind == binder.SymbolArray || symbol.Kind == binder.SymbolHash {
			symbolNames[symbol.Name] = append(symbolNames[symbol.Name], symbol)
		}
	}

	for name, symbols := range symbolNames {
		if len(symbols) > 1 {
			// Sort by scope depth to find inner/outer relationships
			for i, innerSymbol := range symbols {
				for j, outerSymbol := range symbols {
					if i != j && engine.isShadowing(innerSymbol, outerSymbol) {
						diagnostic := Diagnostic{
							Kind:           DiagnosticWarning,
							Message:        fmt.Sprintf("Variable '%s' shadows outer scope variable", name),
							Pos:            innerSymbol.Position,
							Symbol:         innerSymbol,
							SymbolName:     name,
							SymbolKind:     binder.SymbolScalar, // Use actual symbol kind
							FilePath:       engine.filePath,
							LineText:       engine.getLineText(innerSymbol.Position.Line),
							Suggestion:     "Consider using a different name or accessing outer variable as needed",
							RelatedSymbols: []*binder.Symbol{outerSymbol},
							Code:           "PSC-W001",
							HelpMessage:    fmt.Sprintf("Outer variable '%s' declared at line %d", name, outerSymbol.Position.Line),
						}
						diagnostics = append(diagnostics, diagnostic)
					}
				}
			}
		}
	}

	return diagnostics
}

// findUnusedVariables detects variables that are declared but never used
func (engine *EnhancedDiagnosticEngine) findUnusedVariables() []Diagnostic {
	var diagnostics []Diagnostic

	unusedSymbols := engine.usageTracker.GetUnusedSymbols()

	for _, symbol := range unusedSymbols {
		if symbol.Kind == binder.SymbolScalar || symbol.Kind == binder.SymbolArray || symbol.Kind == binder.SymbolHash {
			diagnostic := Diagnostic{
				Kind:        DiagnosticWarning,
				Message:     fmt.Sprintf("Variable '%s' declared but never used", symbol.Name),
				Pos:         symbol.Position,
				Symbol:      symbol,
				SymbolName:  symbol.Name,
				SymbolKind:  symbol.Kind,
				FilePath:    engine.filePath,
				LineText:    engine.getLineText(symbol.Position.Line),
				Suggestion:  "Remove the unused variable or prefix with underscore if intentional",
				Code:        "PSC-W002",
				HelpMessage: "Unused variables may indicate incomplete code or typos",
			}
			diagnostics = append(diagnostics, diagnostic)
		}
	}

	return diagnostics
}

// findTypeMismatches detects type mismatches with enhanced symbol context
func (engine *EnhancedDiagnosticEngine) findTypeMismatches(astRoot ast.Node) []Diagnostic {
	var diagnostics []Diagnostic

	engine.walkAST(astRoot, func(node ast.Node) {
		if assignment, ok := node.(*ast.BinaryExpr); ok {
			// Check if this is an assignment operation
			if assignment.Operator == "=" || assignment.Operator == "+=" || assignment.Operator == "-=" || assignment.Operator == "*=" || assignment.Operator == "/=" {
				if varExpr, ok := assignment.Left.(*ast.VariableExpr); ok {
					varName := varExpr.FullName()
					symbol := engine.symbolTable.ResolveSymbol(varName, binder.SymbolKind(-1))

					if symbol != nil && symbol.Type != "" {
						// Check if assignment type matches declared type
						if !engine.isTypeCompatible(symbol.Type, assignment.Right) {
							diagnostic := Diagnostic{
								Kind:        DiagnosticError,
								Message:     fmt.Sprintf("Variable '%s' declared as %s but assigned incompatible value", varName, symbol.Type),
								Pos:         assignment.Start(),
								Symbol:      symbol,
								SymbolName:  varName,
								SymbolKind:  symbol.Kind,
								FilePath:    engine.filePath,
								LineText:    engine.getLineText(assignment.Start().Line),
								Suggestion:  engine.generateTypeMismatchSuggestion(symbol.Type, assignment.Right),
								Code:        "PSC-E002",
								HelpMessage: fmt.Sprintf("Variable '%s' declared at line %d", varName, symbol.Position.Line),
							}
							diagnostics = append(diagnostics, diagnostic)
						}
					}
				}
			}
		}
	})

	return diagnostics
}

// Helper methods

// findSimilarVariables finds variables with similar names using edit distance
func (engine *EnhancedDiagnosticEngine) findSimilarVariables(varName string) []string {
	var suggestions []string
	allSymbols := engine.symbolTable.GetVisibleSymbols()

	for _, symbol := range allSymbols {
		if (symbol.Kind == binder.SymbolScalar || symbol.Kind == binder.SymbolArray || symbol.Kind == binder.SymbolHash) && engine.editDistance(varName, symbol.Name) <= 2 {
			suggestions = append(suggestions, symbol.Name)
		}
	}

	// Limit suggestions to most relevant
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}

	return suggestions
}

// generateUndefinedVariableSuggestion creates helpful suggestions for undefined variables
func (engine *EnhancedDiagnosticEngine) generateUndefinedVariableSuggestion(varName string, suggestions []string) string {
	if len(suggestions) > 0 {
		if len(suggestions) == 1 {
			return fmt.Sprintf("Did you mean '%s'?", suggestions[0])
		}
		return fmt.Sprintf("Did you mean one of: %s?", strings.Join(suggestions, ", "))
	}

	// Check if it looks like a common variable pattern
	if strings.HasPrefix(varName, "$") {
		return "Make sure to declare the variable with 'my', 'our', or 'state'"
	}

	return "Declare the variable before use or check for typos"
}

// generateTypeMismatchSuggestion creates helpful suggestions for type mismatches
func (engine *EnhancedDiagnosticEngine) generateTypeMismatchSuggestion(expectedType string, actualExpr ast.Node) string {
	switch expectedType {
	case "Int":
		if strings.Contains(actualExpr.Text(), "\"") || strings.Contains(actualExpr.Text(), "'") {
			return "Convert string to integer: int($value) or use 0 + $value"
		}
		return "Ensure the value is an integer"

	case "Str":
		return "Convert to string: \"$value\" or use string interpolation"

	case "Bool":
		return "Use a boolean expression or explicit true/false"

	default:
		return fmt.Sprintf("Ensure the value is compatible with type %s", expectedType)
	}
}

// isShadowing determines if one symbol shadows another
func (engine *EnhancedDiagnosticEngine) isShadowing(inner, outer *binder.Symbol) bool {
	// Simple check: if inner symbol is declared after outer and has different scope
	return inner.Position.Line > outer.Position.Line && inner.Scope != outer.Scope
}

// isTypeCompatible checks if an expression is compatible with a declared type
func (engine *EnhancedDiagnosticEngine) isTypeCompatible(declaredType string, expr ast.Node) bool {
	// This is a simplified type compatibility check
	// In a full implementation, this would use the type checker
	exprText := strings.ToLower(expr.Text())

	switch declaredType {
	case "Int":
		// Check for numeric literals or expressions
		return strings.ContainsAny(exprText, "0123456789") && !strings.ContainsAny(exprText, "\"'")
	case "Str":
		// Check for string literals
		return strings.Contains(exprText, "\"") || strings.Contains(exprText, "'")
	case "Bool":
		// Check for boolean expressions
		return strings.Contains(exprText, "true") || strings.Contains(exprText, "false") ||
			strings.Contains(exprText, "==") || strings.Contains(exprText, "!=")
	default:
		// For complex types, assume compatible (would need full type checker)
		return true
	}
}

// getLineText retrieves the text of a specific line
func (engine *EnhancedDiagnosticEngine) getLineText(lineNum int) string {
	lines := strings.Split(engine.sourceContent, "\n")
	if lineNum > 0 && lineNum <= len(lines) {
		return lines[lineNum-1]
	}
	return ""
}

// walkAST performs a depth-first walk of the AST
func (engine *EnhancedDiagnosticEngine) walkAST(node ast.Node, visitor func(ast.Node)) {
	if node == nil {
		return
	}

	visitor(node)

	for _, child := range node.Children() {
		engine.walkAST(child, visitor)
	}
}

// editDistance calculates the Levenshtein distance between two strings
func (engine *EnhancedDiagnosticEngine) editDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}

	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// FormatDiagnostic formats a diagnostic for display
func (d *Diagnostic) FormatDiagnostic(colorEnabled bool) string {
	var builder strings.Builder

	// Error location
	builder.WriteString(fmt.Sprintf("%s:%d:%d: ", d.FilePath, d.Pos.Line, d.Pos.Column))

	// Kind with color
	kindStr := d.Kind.String() + ":"
	if colorEnabled {
		switch d.Kind {
		case DiagnosticError:
			kindStr = "\033[31m" + kindStr + "\033[0m"
		case DiagnosticWarning:
			kindStr = "\033[33m" + kindStr + "\033[0m"
		case DiagnosticInfo:
			kindStr = "\033[34m" + kindStr + "\033[0m"
		case DiagnosticHint:
			kindStr = "\033[36m" + kindStr + "\033[0m"
		}
	}
	builder.WriteString(kindStr + " ")

	// Message
	builder.WriteString(d.Message)
	if d.Code != "" {
		builder.WriteString(fmt.Sprintf(" [%s]", d.Code))
	}
	builder.WriteString("\n")

	// Line text with pointer
	if d.LineText != "" {
		builder.WriteString(fmt.Sprintf("   %d | %s\n", d.Pos.Line, d.LineText))

		// Error pointer
		pointer := strings.Repeat(" ", 6+d.Pos.Column) + "^"
		if colorEnabled {
			pointer = "\033[31m" + pointer + "\033[0m"
		}
		builder.WriteString(pointer + "\n")
	}

	// Suggestion
	if d.Suggestion != "" {
		suggestionLine := "   help: " + d.Suggestion
		if colorEnabled {
			suggestionLine = "\033[36m" + suggestionLine + "\033[0m"
		}
		builder.WriteString(suggestionLine + "\n")
	}

	// Help message
	if d.HelpMessage != "" {
		builder.WriteString("   note: " + d.HelpMessage + "\n")
	}

	// Did you mean suggestions
	if len(d.DidYouMean) > 0 {
		builder.WriteString("   note: Did you mean: " + strings.Join(d.DidYouMean, ", ") + "\n")
	}

	return builder.String()
}
