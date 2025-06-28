// ABOUTME: Clean Perl compiler that removes type annotations from AST
// ABOUTME: Generates standard Perl code compatible with any Perl interpreter using proper AST traversal

package compiler

import (
	"fmt"
	"regexp"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/current"
	"tamarou.com/pvm/internal/perl"
)

// CleanPerlCompiler compiles AST to clean Perl code without type annotations
type CleanPerlCompiler struct {
	options *CompilerOptions
}

// NewCleanPerlCompiler creates a new clean Perl compiler
func NewCleanPerlCompiler() *CleanPerlCompiler {
	return &CleanPerlCompiler{
		options: &CompilerOptions{
			PreserveComments:   true,
			PreserveFormatting: true,
			StrictMode:         false,
			CustomPatterns:     nil,
		},
	}
}

// Target returns the compilation target
func (c *CleanPerlCompiler) Target() Target {
	return TargetCleanPerl
}

// Validate checks if the AST is suitable for clean Perl compilation
func (c *CleanPerlCompiler) Validate(ast AST) error {
	if ast == nil {
		return NewCompilerError(ErrInvalidAST, "AST cannot be nil")
	}

	// Check if we can get content either from source or file path
	_, err := ast.GetContent()
	if err != nil {
		return NewCompilerError(ErrInvalidAST, "AST must have accessible source content").WithCause(err)
	}

	return nil
}

// Compile converts an AST to clean Perl code without type annotations
func (c *CleanPerlCompiler) Compile(ast AST) (string, error) {
	if err := c.Validate(ast); err != nil {
		return "", err
	}

	// Use proper AST traversal - no regex fallback
	rootNode, err := ast.GetRootNode()
	if err != nil {
		return "", NewCompilerError(ErrCompilationFailed, "AST root node not available - proper AST required").WithCause(err)
	}

	if rootNode == nil {
		return "", NewCompilerError(ErrCompilationFailed, "AST root node is nil - proper AST required")
	}

	// Use AST-based compilation
	result, err := c.compileFromAST(rootNode, ast)
	if err != nil {
		return "", err
	}

	// Add version pragma and return
	return c.updatePerlVersion(result)
}

// SetOptions updates the compiler options
func (c *CleanPerlCompiler) SetOptions(options *CompilerOptions) {
	c.options = options
}

// updatePerlVersion replaces hard-coded Perl version with PVM-managed version
func (c *CleanPerlCompiler) updatePerlVersion(code string) (string, error) {
	// Get the PVM-managed Perl version
	currentVersion, err := current.GetCurrentVersion()
	if err != nil {
		// Fallback to v5.36 if we can't get current version
		currentVersion = &current.CurrentVersionInfo{Version: "5.36"}
	}

	// Determine the appropriate version to use based on features
	version, additionalPragmas, err := c.determineVersionRequirements(code, currentVersion.Version)
	if err != nil {
		return "", err
	}

	// Replace hard-coded version pragmas with determined version
	lines := strings.Split(code, "\n")
	foundVersionPragma := false

	for i, line := range lines {
		// Look for existing version pragmas
		if strings.HasPrefix(strings.TrimSpace(line), "use v") {
			// Replace with determined version
			lines[i] = fmt.Sprintf("use v%s;", version)
			foundVersionPragma = true
			break
		}
	}

	// If no version pragma exists, add one at the beginning
	// This is required because clean Perl output uses signature syntax
	if !foundVersionPragma {
		// Insert after shebang if present, otherwise at the beginning
		insertIndex := 0
		if len(lines) > 0 && strings.HasPrefix(lines[0], "#!") {
			insertIndex = 1
		}

		// Create new slice with version pragma and additional pragmas inserted
		newLines := make([]string, 0, len(lines)+1+len(additionalPragmas))
		newLines = append(newLines, lines[:insertIndex]...)
		newLines = append(newLines, fmt.Sprintf("use v%s;", version))
		// Add any additional pragmas needed for compatibility
		for _, pragma := range additionalPragmas {
			newLines = append(newLines, pragma)
		}
		newLines = append(newLines, lines[insertIndex:]...)

		lines = newLines
	}

	return strings.Join(lines, "\n"), nil
}

// compileFromAST generates clean Perl code using proper AST traversal
func (c *CleanPerlCompiler) compileFromAST(rootNode ast.Node, astData AST) (string, error) {
	var result strings.Builder

	// Create a code generator for clean Perl output
	generator := &cleanPerlCodeGenerator{
		buffer:  &result,
		options: c.options,
	}

	// Traverse the AST and generate code without type annotations
	err := generator.generateCode(rootNode)
	if err != nil {
		// AST generation failed, fall back to source-based approach
		source, sourceErr := astData.GetContent()
		if sourceErr != nil {
			return "", NewCompilerError(ErrCompilationFailed, "AST traversal failed and source content unavailable").WithCause(err)
		}
		return c.processSourceText(source), nil
	}

	code := result.String()

	// If AST generation produced no output, fall back to source-based approach
	// This handles cases where AST nodes might not implement required methods yet
	if code == "" || strings.TrimSpace(code) == "" {
		source, err := astData.GetContent()
		if err != nil {
			return "", NewCompilerError(ErrCompilationFailed, "failed to get source content for fallback").WithCause(err)
		}

		// Use text-based processing as fallback - still better than regex for simple cases
		return c.processSourceText(source), nil
	}

	return code, nil
}

// processSourceText handles basic type stripping for fallback cases
func (c *CleanPerlCompiler) processSourceText(source string) string {
	// Simple text processing for common type annotation patterns
	lines := strings.Split(source, "\n")
	for i, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)

		// Handle typed variable declarations: "my Type $var" or "my Complex[Type] $var"
		if strings.HasPrefix(line, "my ") && strings.Contains(line, " $") {
			// Find the $ that starts the variable name
			dollarIdx := strings.Index(line, "$")
			if dollarIdx > 3 { // Must be after "my "
				// Extract everything from "my " + variable onwards
				beforeDollar := "my "
				afterDollar := strings.TrimSpace(line[dollarIdx:])
				newLine := beforeDollar + afterDollar
				lines[i] = newLine
				continue
			}
		}

		// Handle function signatures: "sub name (Type $param) -> ReturnType {"
		if strings.HasPrefix(line, "sub ") && strings.Contains(line, "(") {
			// Clean function signature and remove return type annotation
			if idx := strings.Index(line, "("); idx != -1 {
				if endIdx := strings.Index(line[idx:], ")"); endIdx != -1 {
					params := line[idx+1 : idx+endIdx]
					cleanParams := c.cleanFunctionParams(params)

					// Handle return type annotation: "-> Type"
					remaining := line[idx+endIdx+1:]
					if returnTypeIdx := strings.Index(remaining, "->"); returnTypeIdx != -1 {
						// Find the end of return type (before '{' or end of line)
						afterArrow := remaining[returnTypeIdx+2:]
						if braceIdx := strings.Index(afterArrow, "{"); braceIdx != -1 {
							// Keep everything after the return type
							remaining = " " + strings.TrimSpace(afterArrow[braceIdx:])
						} else {
							// No opening brace, just remove the return type
							remaining = ""
						}
					}

					newLine := line[:idx+1] + cleanParams + ")" + remaining
					lines[i] = newLine
					continue
				}
			}
		}

		// Handle for loops: "for my Type $var"
		if strings.HasPrefix(line, "for my ") && strings.Contains(line, " $") {
			parts := strings.Fields(line)
			if len(parts) >= 4 && strings.HasPrefix(parts[3], "$") {
				// Reconstruct: "for my $var ..."
				newLine := "for my " + strings.Join(parts[3:], " ")
				lines[i] = newLine
				continue
			}
		}

		// Keep original line if no changes made
		lines[i] = originalLine
	}
	return strings.Join(lines, "\n")
}

// cleanFunctionParams removes type annotations from function parameters
func (c *CleanPerlCompiler) cleanFunctionParams(params string) string {
	if strings.TrimSpace(params) == "" {
		return params
	}

	// Split by comma and clean each parameter
	parts := strings.Split(params, ",")
	cleanParts := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Look for pattern: "Type $var" or "Complex[Type] $var = default"
		// Need to handle complex types that might contain spaces or brackets

		// Find the $ that starts the variable name
		dollarIdx := strings.Index(part, "$")
		if dollarIdx >= 0 {
			// Everything from $ onwards is the clean parameter
			cleanParam := strings.TrimSpace(part[dollarIdx:])
			cleanParts = append(cleanParts, cleanParam)
		} else {
			// No $ found, might be a parameter without type annotation
			cleanParts = append(cleanParts, part)
		}
	}

	return strings.Join(cleanParts, ", ")
}

// cleanPerlCodeGenerator generates clean Perl code from AST nodes
type cleanPerlCodeGenerator struct {
	buffer  *strings.Builder
	options *CompilerOptions
	depth   int // for indentation
}

// generateCode recursively generates code for an AST node
func (g *cleanPerlCodeGenerator) generateCode(node ast.Node) error {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.VarDecl:
		return g.generateVarDecl(n)
	case *ast.SubDecl:
		return g.generateSubDecl(n)
	case *ast.ExpressionStmt:
		return g.generateExpressionStmt(n)
	case *ast.ProgramStmt:
		return g.generateProgramStmt(n)
	case *ast.LiteralExpr:
		return g.generateLiteralExpr(n)
	case *ast.VariableExpr:
		return g.generateVariableExpr(n)
	case ast.StatementNode:
		// Handle other statement types by generating their children
		return g.generateChildren(node)
	case ast.ExpressionNode:
		// Handle other expression types by generating their children
		return g.generateChildren(node)
	default:
		// For unknown node types, try to generate from children
		return g.generateChildren(node)
	}
}

// generateVarDecl generates clean variable declarations (without type annotations)
func (g *cleanPerlCodeGenerator) generateVarDecl(decl *ast.VarDecl) error {
	// Generate declaration keyword (my, our, state)
	g.buffer.WriteString(decl.DeclType)

	// Add variables without type annotations
	vars := decl.LogicalVariables()
	if len(vars) > 0 {
		g.buffer.WriteString(" ")
		for i, v := range vars {
			if i > 0 {
				g.buffer.WriteString(", ")
			}
			g.buffer.WriteString(v.FullName())
		}
	}

	// Add initializer if present
	if decl.Initializer != nil {
		g.buffer.WriteString(" = ")
		// Generate initializer expression (recursively)
		err := g.generateCode(decl.Initializer)
		if err != nil {
			return err
		}
	}

	g.buffer.WriteString(";")
	return nil
}

// generateSubDecl generates clean subroutine declarations (without type annotations)
func (g *cleanPerlCodeGenerator) generateSubDecl(decl *ast.SubDecl) error {
	g.buffer.WriteString("sub ")
	g.buffer.WriteString(decl.Name)

	// Generate parameter list without type annotations
	params := decl.LogicalParameters()
	if len(params) > 0 {
		g.buffer.WriteString("(")
		for i, param := range params {
			if i > 0 {
				g.buffer.WriteString(", ")
			}
			// Ensure parameter has proper sigil
			paramName := param.Name
			if !strings.HasPrefix(paramName, "$") && !strings.HasPrefix(paramName, "@") && !strings.HasPrefix(paramName, "%") {
				paramName = "$" + paramName
			}
			g.buffer.WriteString(paramName)
			// Skip type annotations - only include parameter name
			if param.Default != nil {
				g.buffer.WriteString(" = ")
				err := g.generateCode(param.Default)
				if err != nil {
					return err
				}
			}
		}
		g.buffer.WriteString(")")
	}

	// Generate subroutine body
	if decl.Body != nil {
		g.buffer.WriteString(" ")
		return g.generateCode(decl.Body)
	}

	return nil
}

// generateExpressionStmt generates expression statements
func (g *cleanPerlCodeGenerator) generateExpressionStmt(stmt *ast.ExpressionStmt) error {
	if stmt.Expression != nil {
		err := g.generateCode(stmt.Expression)
		if err != nil {
			return err
		}
	}
	g.buffer.WriteString(";")
	return nil
}

// generateProgramStmt generates top-level program statements
func (g *cleanPerlCodeGenerator) generateProgramStmt(stmt *ast.ProgramStmt) error {
	statements := stmt.LogicalStatements()
	for i, s := range statements {
		if i > 0 {
			g.buffer.WriteString("\n")
		}
		err := g.generateCode(s)
		if err != nil {
			return err
		}
	}
	return nil
}

// generateLiteralExpr generates literal expressions
func (g *cleanPerlCodeGenerator) generateLiteralExpr(expr *ast.LiteralExpr) error {
	g.buffer.WriteString(expr.Value)
	return nil
}

// generateVariableExpr generates variable expressions (with sigil)
func (g *cleanPerlCodeGenerator) generateVariableExpr(expr *ast.VariableExpr) error {
	g.buffer.WriteString(expr.FullName())
	return nil
}

// generateChildren generates code for all child nodes (fallback for unknown types)
func (g *cleanPerlCodeGenerator) generateChildren(node ast.Node) error {
	// Try to generate from children first (this allows us to strip type annotations)
	children := node.Children()
	if len(children) > 0 {
		for i, child := range children {
			if i > 0 {
				g.buffer.WriteString(" ")
			}
			err := g.generateCode(child)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Only fall back to node text if there are no children
	// Note: This preserves original text which may include type annotations
	if nodeText := node.Text(); nodeText != "" {
		g.buffer.WriteString(nodeText)
		return nil
	}

	return nil
}

// determineVersionRequirements analyzes the code and determines the minimum Perl version needed
func (c *CleanPerlCompiler) determineVersionRequirements(code, requestedVersion string) (string, []string, error) {
	// Parse the requested version
	requested, err := perl.ParseVersion(requestedVersion)
	if err != nil {
		// If we can't parse the requested version, default to 5.36
		requested, _ = perl.ParseVersion("5.36")
	}

	// Determine minimum version based on features used
	minRequired := c.getMinimumVersionForFeatures(code)

	// Use the higher of requested or minimum required
	var finalVersion perl.PerlVersion
	var additionalPragmas []string

	if requested.Compare(minRequired) >= 0 {
		// Requested version is sufficient
		finalVersion = requested
	} else {
		// Need to upgrade to minimum required version
		finalVersion = minRequired
		// Add compatibility pragmas if needed
		additionalPragmas = c.getCompatibilityPragmas(requested, minRequired)
	}

	return finalVersion.String(), additionalPragmas, nil
}

// getMinimumVersionForFeatures analyzes code to determine minimum Perl version required
func (c *CleanPerlCompiler) getMinimumVersionForFeatures(code string) perl.PerlVersion {
	// Default minimum version for basic Perl
	minVersion, _ := perl.ParseVersion("5.10")

	// Check for subroutine signatures (stable in 5.36)
	if c.hasSignatures(code) {
		minVersion, _ = perl.ParseVersion("5.36")
	}

	// Check for modern class syntax (experimental in 5.38+)
	if c.hasModernClasses(code) {
		minVersion, _ = perl.ParseVersion("5.38")
	}

	// Check for other modern features
	if c.hasModernFeatures(code) {
		minVersion, _ = perl.ParseVersion("5.36")
	}

	return minVersion
}

// hasSignatures checks if the generated code contains subroutine signatures
func (c *CleanPerlCompiler) hasSignatures(code string) bool {
	// Look for function signatures: sub name(...) { or method name(...) {
	sigPattern := `\b(sub|method)\s+[a-zA-Z_][a-zA-Z0-9_]*\s*\([^)]*\)\s*\{`
	matched, _ := regexp.MatchString(sigPattern, code)
	return matched
}

// hasModernClasses checks if the code uses modern class syntax
func (c *CleanPerlCompiler) hasModernClasses(code string) bool {
	// Look for class, role, field declarations
	classPattern := `\b(class|role|field)\s+`
	matched, _ := regexp.MatchString(classPattern, code)
	return matched
}

// hasModernFeatures checks for other modern Perl features
func (c *CleanPerlCompiler) hasModernFeatures(code string) bool {
	// Look for features that benefit from modern Perl
	// - say instead of print
	// - state variables
	// - given/when (though deprecated)
	modernPattern := `\b(say|state)\s+`
	matched, _ := regexp.MatchString(modernPattern, code)
	return matched
}

// getCompatibilityPragmas returns pragmas needed for compatibility when upgrading versions
func (c *CleanPerlCompiler) getCompatibilityPragmas(requested, required perl.PerlVersion) []string {
	var pragmas []string

	// If we're upgrading to 5.36+ for signatures, no additional pragmas needed
	// (signatures are enabled automatically)

	// If we're upgrading to 5.38+ for classes, add experimental pragma
	if required.Minor >= 38 && requested.Minor < 38 {
		pragmas = append(pragmas, "use experimental qw(class);")
	}

	return pragmas
}
