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
// Deprecated: Use NewCleanPerlCompilerUnified() from perl_compiler.go instead.
// This legacy compiler uses the old AST-based approach and will be removed in a future version.
type CleanPerlCompiler struct {
	options *CompilerOptions
}

// NewCleanPerlCompiler creates a new clean Perl compiler
// Deprecated: Use NewCleanPerlCompilerUnified() instead for better CST-based compilation.
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

	// Get source content for text extraction
	source, err := astData.GetContent()
	if err != nil {
		return "", NewCompilerError(ErrCompilationFailed, "failed to get source content for AST traversal").WithCause(err)
	}

	// Create a code generator for clean Perl output
	generator := &cleanPerlCodeGenerator{
		buffer:  &result,
		options: c.options,
		source:  source,
	}

	// Traverse the AST and generate code without type annotations
	genErr := generator.generateCode(rootNode)
	if genErr != nil {
		// AST generation failed, fall back to source-based approach
		source, sourceErr := astData.GetContent()
		if sourceErr != nil {
			return "", NewCompilerError(ErrCompilationFailed, "AST traversal failed and source content unavailable").WithCause(genErr)
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
	source  string // source text for node text extraction
	depth   int    // for indentation
}

// generateCode recursively generates code for an AST node
func (g *cleanPerlCodeGenerator) generateCode(node ast.Node) error {
	if node == nil {
		return nil
	}

	nodeType := node.Type()

	// Skip type annotation nodes entirely (like ASTCompiler does)
	if g.isTypeAnnotationNode(nodeType) {
		return nil
	}

	// Handle token nodes - preserve their text directly
	if tokenNode, ok := node.(*ast.TokenNode); ok {
		return g.handleTokenNode(tokenNode)
	}

	// Use semantic node handling like ASTCompiler
	return g.handleNodeSemantics(node)
}

// isTypeAnnotationNode checks if a node represents a type annotation
func (g *cleanPerlCodeGenerator) isTypeAnnotationNode(nodeType string) bool {
	typeNodes := map[string]bool{
		// Basic type annotations
		"type_expression": true,
		"type_annotation": true,
		"scalar_type":     true,
		"array_type":      true,
		"hash_type":       true,

		// Method/function types
		"method_return_type":     true,
		"typed_method_parameter": true,
		"return_type":            true,
		"parameter_type":         true,

		// Complex type constructs
		"type_assertion":     true,
		"type_declaration":   true,
		"union_type":         true,
		"intersection_type":  true,
		"negation_type":      true,
		"parameterized_type": true,

		// Named types that start with capital letters (heuristic)
		// These might be parsed as identifiers but represent types
		"Int":      true,
		"Str":      true,
		"Bool":     true,
		"Num":      true,
		"ArrayRef": true,
		"HashRef":  true,
		"CodeRef":  true,
		"Any":      true,
		"Undef":    true,
		"Maybe":    true,
		"Union":    true,
	}
	return typeNodes[nodeType]
}

// handleTokenNode processes token nodes
func (g *cleanPerlCodeGenerator) handleTokenNode(tokenNode *ast.TokenNode) error {
	// Extract text directly for token nodes
	text := tokenNode.Text()
	g.buffer.WriteString(text)
	return nil
}

// handleVariableDeclaration processes variable declarations
func (g *cleanPerlCodeGenerator) handleVariableDeclaration(node ast.Node) error {
	// Check if it's a typed VarDecl we can handle directly
	if varDecl, ok := node.(*ast.VarDecl); ok {
		return g.generateVarDecl(varDecl)
	}

	// Otherwise use generic var_decl handling for parsed ASTs
	return g.generateGenericVarDecl(node)
}

// handleSubroutineDeclaration processes subroutine declarations
func (g *cleanPerlCodeGenerator) handleSubroutineDeclaration(node ast.Node) error {
	// Check if it's a typed SubDecl we can handle directly
	if subDecl, ok := node.(*ast.SubDecl); ok {
		return g.generateSubDecl(subDecl)
	}

	// Otherwise use generic sub_decl handling
	return g.generateGenericSubDecl(node)
}

// handleFunctionCall processes function calls
func (g *cleanPerlCodeGenerator) handleFunctionCall(node ast.Node) error {
	// For function calls, just walk children normally
	return g.walkChildren(node)
}

// handleStringLiteral processes string literals
func (g *cleanPerlCodeGenerator) handleStringLiteral(node ast.Node) error {
	// For string literals, preserve exact text
	text := g.extractNodeText(node)
	if text != "" {
		g.buffer.WriteString(text)
		return nil
	}
	return g.walkChildren(node)
}

// handleLiteral processes literal expressions
func (g *cleanPerlCodeGenerator) handleLiteral(node ast.Node) error {
	// Check if it's a typed LiteralExpr we can handle directly
	if literalExpr, ok := node.(*ast.LiteralExpr); ok {
		g.buffer.WriteString(literalExpr.Value)
		return nil
	}

	// For other literal nodes, try to extract text
	text := g.extractNodeText(node)
	if text != "" {
		g.buffer.WriteString(text)
		return nil
	}
	return g.walkChildren(node)
}

// walkChildren walks all children of a node
func (g *cleanPerlCodeGenerator) walkChildren(node ast.Node) error {
	children := node.Children()

	// Collect valid children (non-type-annotation children)
	var validChildren []ast.Node
	for _, child := range children {
		childType := child.Type()
		// Skip type annotations
		if g.isTypeAnnotationNode(childType) {
			continue
		}
		validChildren = append(validChildren, child)
	}

	// Process valid children with appropriate spacing
	for i, child := range validChildren {
		if i > 0 {
			// Add space between valid children
			g.buffer.WriteString(" ")
		}

		err := g.generateCode(child)
		if err != nil {
			return err
		}
	}
	return nil
}

// cleanVariableDeclarationText removes type annotations from variable declaration text
func (g *cleanPerlCodeGenerator) cleanVariableDeclarationText(text string) string {
	// Handle patterns like "my Int $x = 42" -> "my $x = 42"
	words := strings.Fields(text)
	if len(words) < 3 {
		return text
	}

	// Look for pattern: "my" + TypeName + "$variable" + optional rest
	if words[0] == "my" && len(words) >= 3 {
		// Check if second word looks like a type name and third word is a variable
		if g.looksLikeTypeName(words[1]) && strings.HasPrefix(words[2], "$") {
			// Remove the type name (second word)
			result := words[0] // "my"
			for i := 2; i < len(words); i++ {
				result += " " + words[i]
			}
			return result
		}
	}

	return text
}

// looksLikeTypeName checks if a string looks like a type name
func (g *cleanPerlCodeGenerator) looksLikeTypeName(identifier string) bool {
	if len(identifier) == 0 {
		return false
	}
	firstChar := identifier[0]
	return firstChar >= 'A' && firstChar <= 'Z'
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

// extractNodeText extracts the source text covered by a node (like ASTCompiler does)
func (g *cleanPerlCodeGenerator) extractNodeText(node ast.Node) string {
	// First try to use the node's Text() method directly
	nodeText := node.Text()
	if nodeText != "" {
		return nodeText
	}

	// Fallback to position-based extraction if we have valid offsets
	start := node.Start()
	end := node.End()

	if g.source != "" && start.Offset >= 0 && end.Offset > start.Offset && end.Offset <= len(g.source) {
		// Extract text using byte offsets
		return g.source[start.Offset:end.Offset]
	}

	// If all else fails, return empty string
	return ""
}

// handleNodeSemantics processes nodes based on their semantic meaning (like ASTCompiler)
func (g *cleanPerlCodeGenerator) handleNodeSemantics(node ast.Node) error {
	nodeType := node.Type()

	// Handle based on semantic type, not text manipulation
	switch nodeType {
	case "var_decl", "variable_declaration":
		return g.handleVariableDeclaration(node)
	case "subroutine_definition", "sub_decl":
		return g.handleSubroutineDeclaration(node)
	case "ambiguous_function_call_expression":
		return g.handleFunctionCall(node)
	case "interpolated_string_literal", "string_literal":
		return g.handleStringLiteral(node)
	case "literal":
		return g.handleLiteral(node)
	case "source_file":
		return g.generateSourceFile(node)
	case "expression_statement":
		return g.generateExpressionStatement(node)
	case "token":
		return g.generateToken(node)
	default:
		// For unknown types, walk children
		return g.walkChildren(node)
	}
}

// generateGenericVarDecl handles variable declarations when not a specific *ast.VarDecl type
func (g *cleanPerlCodeGenerator) generateGenericVarDecl(node ast.Node) error {
	// Use AST-based processing as primary approach to handle all type annotation scenarios properly
	children := node.Children()

	// Collect valid children (non-type-annotation children)
	var validChildren []ast.Node
	for _, child := range children {
		childType := child.Type()
		// Skip type annotations
		if g.isTypeAnnotationNode(childType) {
			continue
		}
		validChildren = append(validChildren, child)
	}

	// Process valid children with appropriate spacing
	for i, child := range validChildren {
		if i > 0 {
			g.buffer.WriteString(" ")
		}

		err := g.generateCode(child)
		if err != nil {
			return err
		}
	}

	// Add semicolon if we generated content
	if len(validChildren) > 0 {
		g.buffer.WriteString(";")
	}
	return nil
}

// generateGenericSubDecl handles subroutine declarations when not a specific *ast.SubDecl type
func (g *cleanPerlCodeGenerator) generateGenericSubDecl(node ast.Node) error {
	// For generic sub_decl nodes, process children while stripping type annotations
	children := node.Children()
	for i, child := range children {
		if i > 0 {
			g.buffer.WriteString(" ")
		}

		childType := child.Type()
		// Skip type annotations and return types
		if childType == "type_expression" || childType == "type_annotation" ||
			childType == "method_return_type" || childType == "return_type" {
			continue
		}

		err := g.generateCode(child)
		if err != nil {
			return err
		}
	}
	return nil
}

// generateExpressionStatement handles expression_statement nodes
func (g *cleanPerlCodeGenerator) generateExpressionStatement(node ast.Node) error {
	// Process all children of the expression statement
	children := node.Children()
	for _, child := range children {
		err := g.generateCode(child)
		if err != nil {
			return err
		}
	}
	return nil
}

// generateSourceFile handles source_file root nodes
func (g *cleanPerlCodeGenerator) generateSourceFile(node ast.Node) error {
	// Process all top-level statements
	children := node.Children()
	for i, child := range children {
		if i > 0 && child.Type() != "token" {
			g.buffer.WriteString("\n")
		}
		err := g.generateCode(child)
		if err != nil {
			return err
		}
	}
	return nil
}

// generateToken handles token nodes
func (g *cleanPerlCodeGenerator) generateToken(node ast.Node) error {
	// For token nodes, include their text unless they're type-related
	text := node.Text()
	// Skip certain tokens that are just whitespace or type-related
	if text == "\n" || text == ";" {
		g.buffer.WriteString(text)
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
