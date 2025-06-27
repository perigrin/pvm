// ABOUTME: Clean Perl compiler that removes type annotations from AST
// ABOUTME: Generates standard Perl code compatible with any Perl interpreter using proper AST traversal

package compiler

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/current"
	"tamarou.com/pvm/internal/perl"
	"regexp"
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

	// Traverse the AST and generate code
	err := generator.generateCode(rootNode)
	if err != nil {
		return "", NewCompilerError(ErrCompilationFailed, "AST traversal failed").WithCause(err)
	}

	code := result.String()
	if code == "" {
		return "", NewCompilerError(ErrCompilationFailed, "AST compilation produced empty result")
	}

	return code, nil
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
			g.buffer.WriteString(param.Name)
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
	children := node.Children()
	for _, child := range children {
		err := g.generateCode(child)
		if err != nil {
			return err
		}
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
