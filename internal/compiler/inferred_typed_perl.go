// ABOUTME: Inferred typed Perl compiler that generates type annotations from inference results
// ABOUTME: Transforms untyped Perl into typed Perl while preserving all original behavior

package compiler

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/compiler/pipeline"
	"tamarou.com/pvm/internal/current"
	"tamarou.com/pvm/internal/types"
)

// InferredTypedPerlCompiler compiles AST with inferred type annotations to typed Perl code
type InferredTypedPerlCompiler struct {
	// Formatter for generating type annotations
	formatter TypeFormatter

	// Annotation generator for creating type annotations
	annotationGenerator *AnnotationGenerator

	// Options for compilation behavior
	options InferredCompilerOptions
}

// InferredCompilerOptions controls compilation behavior
type InferredCompilerOptions struct {
	// AnnotationStyle sets the formatting style for annotations
	AnnotationStyle FormattingStyle

	// PreserveComments controls whether original comments are preserved
	PreserveComments bool

	// PreserveFormatting controls whether original formatting is preserved
	PreserveFormatting bool

	// VerboseOutput includes additional debugging information
	VerboseOutput bool

	// MinimumConfidence sets the minimum confidence level for type annotations
	MinimumConfidence float64
}

// NewInferredTypedPerlCompiler creates a new inferred typed Perl compiler
func NewInferredTypedPerlCompiler() *InferredTypedPerlCompiler {
	formatter := NewTypeFormatter()

	return &InferredTypedPerlCompiler{
		formatter:           formatter,
		annotationGenerator: NewAnnotationGenerator(formatter),
		options: InferredCompilerOptions{
			AnnotationStyle:    StyleInline,
			PreserveComments:   true,
			PreserveFormatting: true,
			VerboseOutput:      false,
		},
	}
}

// NewInferredTypedPerlCompilerWithOptions creates a compiler with custom options
func NewInferredTypedPerlCompilerWithOptions(options InferredCompilerOptions) *InferredTypedPerlCompiler {
	formatterOptions := FormatterOptions{
		IncludeConfidenceComments: options.VerboseOutput,
		UseShortTypeNames:         options.AnnotationStyle == StyleCompact,
		PreferComments:            false, // Always use inline annotations
	}

	formatter := NewTypeFormatterWithOptions(formatterOptions)

	annotationOptions := AnnotationOptions{
		AnnotateVariables: true,
		AnnotateMethods:   true,
		AnnotateFields:    true,
		AnnotateReturns:   false, // Usually too noisy
		PreferredStyle:    options.AnnotationStyle,
		ContextAware:      true,
	}

	return &InferredTypedPerlCompiler{
		formatter:           formatter,
		annotationGenerator: NewAnnotationGeneratorWithOptions(formatter, annotationOptions),
		options:             options,
	}
}

// Target returns the compilation target this compiler supports
func (itpc *InferredTypedPerlCompiler) Target() Target {
	return TargetInferredTypeAnnotations
}

// Validate checks if the AST is suitable for compilation with this compiler
//
//nolint:sloppyTypeAssert // Function intentionally uses type assertions for interface-to-concrete conversions
func (itpc *InferredTypedPerlCompiler) Validate(inputAST AST) error {
	if inputAST == nil {
		return &CompilerError{
			Code:    "INVALID_AST",
			Message: "AST cannot be nil",
		}
	}

	root, err := inputAST.GetRootNode()
	if err != nil || root == nil {
		return &CompilerError{
			Code:    "INVALID_AST",
			Message: "AST must have a root node",
		}
	}

	return nil
}

// Compile converts an AST with inferred type information to typed Perl code
func (itpc *InferredTypedPerlCompiler) Compile(inputAST AST) (string, error) {
	// Validate input
	if err := itpc.Validate(inputAST); err != nil {
		return "", err
	}

	// Use CST-based compilation directly (like CompileInferred does)
	content, err := inputAST.GetContent()
	if err != nil {
		return "", &CompilerError{
			Code:    "COMPILATION_FAILED",
			Message: fmt.Sprintf("Failed to get source content: %v", err),
		}
	}

	// Re-parse using tree-sitter to get CST that preserves type annotations
	cstAST, err := NewCSTBasedAST(inputAST.GetPath(), content)
	if err != nil {
		return "", &CompilerError{
			Code:    "COMPILATION_FAILED",
			Message: fmt.Sprintf("Failed to parse content for CST: %v", err),
		}
	}

	// Use CreateTypedPerl to preserve existing type annotations
	result, err := CreateTypedPerl(cstAST.Root, []byte(content))
	if err != nil {
		return "", &CompilerError{
			Code:    "COMPILATION_FAILED",
			Message: fmt.Sprintf("Failed to create typed Perl: %v", err),
		}
	}

	if !result.Success {
		errorMsg := "CST transformation failed"
		if result.Error != nil {
			errorMsg = result.Error.Error()
		}
		return "", &CompilerError{
			Code:    "COMPILATION_FAILED",
			Message: errorMsg,
		}
	}

	// Add version pragma
	return itpc.addPerlVersionPragma(result.TransformedCode), nil
}

// CompileInferred generates typed Perl code from an already-inferred AST with type information
func (itpc *InferredTypedPerlCompiler) CompileInferred(inferredAST ast.InferredAST) (string, error) {
	// Validate input
	if inferredAST == nil {
		return "", &CompilerError{
			Code:    "INVALID_AST",
			Message: "InferredAST cannot be nil",
		}
	}

	if !inferredAST.IsValid() {
		return "", &CompilerError{
			Code:    "INVALID_AST",
			Message: "InferredAST is not valid",
		}
	}

	// Get the original source content
	content, err := inferredAST.GetContent()
	if err != nil {
		return "", &CompilerError{
			Code:    "COMPILATION_FAILED",
			Message: fmt.Sprintf("Failed to get source content: %v", err),
		}
	}

	// Use CST-based compilation like the working TargetTypedPerl path
	// Re-parse content using tree-sitter to get CST that preserves type annotations
	cstAST, err := NewCSTBasedAST(inferredAST.GetPath(), content)
	if err != nil {
		return "", &CompilerError{
			Code:    "COMPILATION_FAILED",
			Message: fmt.Sprintf("Failed to parse content for CST: %v", err),
		}
	}

	// Use CreateTypedPerl to preserve existing type annotations (like --disable-inference)
	result, err := CreateTypedPerl(cstAST.Root, []byte(content))
	if err != nil {
		return "", &CompilerError{
			Code:    "COMPILATION_FAILED",
			Message: fmt.Sprintf("Failed to create typed Perl: %v", err),
		}
	}

	if !result.Success {
		errorMsg := "CST transformation failed"
		if result.Error != nil {
			errorMsg = result.Error.Error()
		}
		return "", &CompilerError{
			Code:    "COMPILATION_FAILED",
			Message: errorMsg,
		}
	}

	// Add version pragma
	finalResult := itpc.addPerlVersionPragma(result.TransformedCode)

	// TODO: Add inferred type annotations for variables that don't have explicit types
	// For now, we preserve existing types and this fixes the main issue

	return finalResult, nil
}

// addInferenceAnnotations adds type annotations based on inference results using the transformation pipeline
func (itpc *InferredTypedPerlCompiler) addInferenceAnnotations(inferredAST ast.InferredAST, content string) (string, error) {
	// Get all type information from the inference engine
	typeInfo := inferredAST.GetAllTypeInfo()
	if len(typeInfo) == 0 {
		// No type information available, return content unchanged
		return content, nil
	}

	// Create a node compiler to handle AST-based compilation with type annotations
	formatter := NewTypeFormatter()
	annotationOptions := AnnotationOptions{
		AnnotateVariables: true,
		AnnotateMethods:   true,
		AnnotateFields:    true,
		AnnotateReturns:   true,
		MinConfidence:     itpc.options.MinimumConfidence,
		PreferredStyle:    itpc.options.AnnotationStyle,
		ContextAware:      true,
	}

	nc := &nodeCompiler{
		inferredAST:         inferredAST,
		annotationGenerator: NewAnnotationGeneratorWithOptions(formatter, annotationOptions),
		options:             itpc.options,
		typeInfoMap:         typeInfo,
	}

	// Compile the AST with type annotations
	rootNode, err := inferredAST.GetRootNode()
	if err != nil {
		return "", fmt.Errorf("failed to get root node: %w", err)
	}

	result, err := nc.compileNode(rootNode)
	if err != nil {
		return "", fmt.Errorf("failed to compile with type annotations: %w", err)
	}

	return result, nil
}

// addInferenceAnnotationsWithPipeline uses the transformation pipeline to inject type nodes
func (itpc *InferredTypedPerlCompiler) addInferenceAnnotationsWithPipeline(inferredAST ast.InferredAST, content string, typeInfo map[string]*types.TypeInfo) (string, error) {
	// Create type injection options
	options := pipeline.TypeInjectionOptions{
		AnnotationStyle:     string(itpc.options.AnnotationStyle),
		PreserveFormatting:  itpc.options.PreserveFormatting,
		InjectVariableTypes: true,
		InjectMethodTypes:   true,
		InjectReturnTypes:   true,
	}

	// Build pipeline with type injection
	transformationPipeline := pipeline.NewPipelineBuilder().
		WithTypeInjection(typeInfo, options).
		WithWhitespaceNormalization().
		Build()

	// Parse the content to get a CST for the transformation pipeline
	// This follows the pattern used in pipeline_compiler.go
	cstAST, err := NewCSTBasedAST(inferredAST.GetPath(), content)
	if err != nil {
		return "", fmt.Errorf("failed to create CST from content: %w", err)
	}

	// Execute the transformation pipeline
	result, err := transformationPipeline.Execute(cstAST.GetCSTRoot(), []byte(content))
	if err != nil {
		return "", fmt.Errorf("transformation pipeline failed: %w", err)
	}

	return string(result.Content), nil
}

// addPerlVersionPragma adds the Perl version pragma after the shebang line (if present)
func (itpc *InferredTypedPerlCompiler) addPerlVersionPragma(content string) string {
	// Get the PVM-managed Perl version
	currentVersion, err := current.GetCurrentVersion()
	if err != nil {
		// Fallback to v5.36 if we can't get current version
		currentVersion = &current.CurrentVersionInfo{Version: "5.36"}
	}

	pragma := fmt.Sprintf("use v%s;", currentVersion.Version)

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return pragma + "\n" + content
	}

	// Check if first line is a shebang
	if strings.HasPrefix(lines[0], "#!") {
		// Insert pragma after shebang
		if len(lines) == 1 {
			return lines[0] + "\n" + pragma
		}
		// Insert after first line
		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[0])
		newLines = append(newLines, pragma)
		newLines = append(newLines, lines[1:]...)
		return strings.Join(newLines, "\n")
	}

	// No shebang, prepend pragma
	return pragma + "\n" + content
}

// nodeCompiler handles compilation of individual AST nodes
type nodeCompiler struct {
	inferredAST         ast.InferredAST
	annotationGenerator *AnnotationGenerator
	options             InferredCompilerOptions
	nodeCounter         int                        // For generating unique node IDs
	typeInfoMap         map[string]*types.TypeInfo // Map of node IDs to type information
}

// generateNodeID generates a position-based ID for a node that matches the inference engine format
func (nc *nodeCompiler) generateNodeID(node ast.Node) string {
	// Generate ID based on node type and position
	// Format: nodetype_startcol_endcol (e.g., "variable_declaration_4_10")
	startPos := node.Start()
	endPos := node.End()

	// Map AST node types to inference engine node types
	nodeType := nc.mapNodeType(node)

	return fmt.Sprintf("%s_%d_%d", nodeType, startPos.Column, endPos.Column)
}

// mapNodeType maps AST node types to the format expected by the inference engine
func (nc *nodeCompiler) mapNodeType(node ast.Node) string {
	switch n := node.(type) {
	case *ast.VarDecl:
		return "variable_declaration"
	case *ast.SubDecl:
		if n.IsMethod {
			return "method_declaration"
		}
		return "subroutine_declaration"
	case *ast.FieldDecl:
		return "field_declaration"
	default:
		// Use the node's Type() method but convert to snake_case
		return strings.ToLower(strings.ReplaceAll(node.Type(), " ", "_"))
	}
}

// getNodeTypeInfo gets type info for a node using position-based ID lookup
func (nc *nodeCompiler) getNodeTypeInfo(node ast.Node) *types.TypeInfo {
	// Try multiple ID formats to find type info
	// First try position-based ID
	nodeID := nc.generateNodeID(node)
	if info, exists := nc.typeInfoMap[nodeID]; exists {
		return info
	}

	// Try with node kind and byte positions (alternative format)
	startPos := node.Start()
	endPos := node.End()
	alternativeID := fmt.Sprintf("%s_%d_%d", node.Type(), startPos.Offset, endPos.Offset)
	if info, exists := nc.typeInfoMap[alternativeID]; exists {
		return info
	}

	// For special nodes like return types, try context-based IDs
	if subDecl, ok := node.(*ast.SubDecl); ok {
		returnID := fmt.Sprintf("subroutine_return_%d_%d", startPos.Column, endPos.Column)
		if info, exists := nc.typeInfoMap[returnID]; exists {
			return info
		}
		if subDecl.IsMethod {
			methodReturnID := fmt.Sprintf("method_return_%d_%d", startPos.Column, endPos.Column)
			if info, exists := nc.typeInfoMap[methodReturnID]; exists {
				return info
			}
		}
	}

	return nil
}

// compileNode compiles a single AST node to Perl code
func (nc *nodeCompiler) compileNode(node ast.Node) (string, error) {
	if node == nil {
		return "", nil
	}

	nodeType := node.Type()
	fmt.Printf("DEBUG INFERRED COMPILER: compileNode type=%s text='%s'\n", nodeType, node.Text())

	// Handle tree-sitter specific node types
	switch nodeType {
	case "source_file":
		return nc.compileSourceFile(node)
	case "expression_statement":
		return nc.compileExpressionStatement(node)
	case "variable_declaration":
		return nc.compileVariableDeclaration(node)
	case "typed_variable_declaration":
		return nc.compileTypedVariableDeclaration(node)
	case "subroutine_declaration_statement":
		return nc.compileSubroutineDeclaration(node)
	case "method_declaration_statement":
		return nc.compileMethodDeclaration(node)
	case "class_statement":
		return nc.compileClassStatement(node)
	case "role_statement":
		return nc.compileRoleStatement(node)
	case "package_statement":
		return nc.compilePackageStatement(node)
	case "block":
		return nc.compileBlock(node)
	default:
		// Fall back to existing AST node handling for backwards compatibility
		switch n := node.(type) {
		case *ast.ProgramStmt:
			return nc.compileProgramStmt(n)
		case *ast.VarDecl:
			return nc.compileVarDecl(n)
		case *ast.SubDecl:
			return nc.compileSubDecl(n)
		case *ast.MethodDecl:
			return nc.compileMethodDecl(n)
		case *ast.FieldDecl:
			return nc.compileFieldDecl(n)
		case *ast.BlockStmt:
			return nc.compileBlockStmt(n)
		case *ast.ExpressionStmt:
			return nc.compileExpressionStmt(n)
		case *ast.ReturnStmt:
			return nc.compileReturnStmt(n)
		case *ast.IfStmt:
			return nc.compileIfStmt(n)
		case *ast.WhileStmt:
			return nc.compileWhileStmt(n)
		case *ast.ForStmt:
			return nc.compileForStmt(n)
		case *ast.UseStmt:
			return nc.compileUseStmt(n)
		case *ast.PackageStmt:
			return nc.compilePackageStmt(n)
		case *ast.LiteralExpr:
			return nc.compileLiteralExpr(n)
		case *ast.VariableExpr:
			return nc.compileVariableExpr(n)
		default:
			// For other node types, compile children
			return nc.compileGenericNode(node)
		}
	}
}

// compileProgramStmt compiles a program statement
func (nc *nodeCompiler) compileProgramStmt(node *ast.ProgramStmt) (string, error) {
	var parts []string

	// Add Perl version pragma for modern features using PVM-managed version
	currentVersion, err := current.GetCurrentVersion()
	if err != nil {
		// Fallback to v5.36 if we can't get current version
		currentVersion = &current.CurrentVersionInfo{Version: "5.36"}
	}
	parts = append(parts, fmt.Sprintf("use v%s;", currentVersion.Version))

	// Compile all statements
	for _, stmt := range node.Statements() {
		compiled, err := nc.compileNode(stmt)
		if err != nil {
			return "", err
		}
		if compiled != "" {
			parts = append(parts, compiled)
		}
	}

	return strings.Join(parts, "\n"), nil
}

// compileTypeExpression compiles a type expression to its string representation
func (nc *nodeCompiler) compileTypeExpression(typeExpr *ast.TypeExpression) (string, error) {
	if typeExpr == nil {
		return "", nil
	}

	// Handle different type expression kinds
	switch typeExpr.Kind {
	case ast.SimpleTypeKind:
		return typeExpr.Name, nil
	case ast.ParameterizedTypeKind:
		// e.g., ArrayRef[Int]
		if len(typeExpr.Parameters) > 0 {
			params := make([]string, 0, len(typeExpr.Parameters))
			for _, param := range typeExpr.Parameters {
				paramStr, err := nc.compileTypeExpression(param)
				if err != nil {
					return "", err
				}
				params = append(params, paramStr)
			}
			return typeExpr.Name + "[" + strings.Join(params, ", ") + "]", nil
		}
		return typeExpr.Name, nil
	case ast.UnionTypeKind:
		// e.g., Int|Str
		parts := []string{typeExpr.Name}
		for _, param := range typeExpr.Parameters {
			paramStr, err := nc.compileTypeExpression(param)
			if err != nil {
				return "", err
			}
			parts = append(parts, paramStr)
		}
		return strings.Join(parts, "|"), nil
	case ast.IntersectionTypeKind:
		// e.g., Object&Serializable
		parts := []string{typeExpr.Name}
		for _, param := range typeExpr.Parameters {
			paramStr, err := nc.compileTypeExpression(param)
			if err != nil {
				return "", err
			}
			parts = append(parts, paramStr)
		}
		return strings.Join(parts, "&"), nil
	case ast.NegationTypeKind:
		// e.g., !Undef
		if len(typeExpr.Parameters) > 0 {
			paramStr, err := nc.compileTypeExpression(typeExpr.Parameters[0])
			if err != nil {
				return "", err
			}
			return "!" + paramStr, nil
		}
		return "!" + typeExpr.Name, nil
	default:
		// Fallback: just use the name
		return typeExpr.Name, nil
	}
}

// compileVarDecl compiles a variable declaration with type annotations
func (nc *nodeCompiler) compileVarDecl(node *ast.VarDecl) (string, error) {
	var parts []string

	// Start with declaration type (my, our, state)
	parts = append(parts, node.DeclType)

	// Check if this VarDecl already has a type annotation (from typed Perl source)
	fmt.Printf("DEBUG INFERRED COMPILER: VarDecl TypeExpr: %v\n", node.TypeExpr != nil)
	if node.TypeExpr != nil {
		fmt.Printf("DEBUG INFERRED COMPILER: TypeExpr Name: %s, Kind: %v\n", node.TypeExpr.Name, node.TypeExpr.Kind)
		// Preserve the original type annotation
		typeStr, err := nc.compileTypeExpression(node.TypeExpr)
		fmt.Printf("DEBUG INFERRED COMPILER: Compiled to: %s (err: %v)\n", typeStr, err)
		if err == nil && typeStr != "" {
			parts = append(parts, typeStr)
		}
	}

	// Process variables with potential type annotations
	var varDecls []string
	for _, variable := range node.Variables() {
		// If we already have a type annotation from the source, just use the variable name
		if node.TypeExpr != nil {
			varDecls = append(varDecls, variable.FullName())
		} else {
			// Try to get inferred type info for this specific variable
			varTypeInfo := nc.getVariableTypeInfo(variable)

			if varTypeInfo != nil && varTypeInfo.Confidence >= nc.options.MinimumConfidence {
				// Add type annotation based on style
				varDecl := nc.formatVariableWithType(variable, varTypeInfo)
				varDecls = append(varDecls, varDecl)
			} else {
				// No type info or low confidence, use plain variable
				varDecls = append(varDecls, variable.FullName())
			}
		}
	}

	// Handle multiple variables
	if len(varDecls) > 1 {
		parts = append(parts, "("+strings.Join(varDecls, ", ")+")")
	} else if len(varDecls) == 1 {
		parts = append(parts, varDecls[0])
	}

	// Add initializer if present
	if node.Initializer != nil {
		parts = append(parts, "=")
		initCode, err := nc.compileNode(node.Initializer)
		if err != nil {
			return "", err
		}
		parts = append(parts, initCode)
	}

	result := strings.Join(parts, " ") + ";"

	// Add verbose comments if requested
	if nc.options.AnnotationStyle == StyleVerbose && len(varDecls) > 0 {
		for _, variable := range node.Variables() {
			if varTypeInfo := nc.getVariableTypeInfo(variable); varTypeInfo != nil {
				comment := fmt.Sprintf("# Type: %s (confidence: %.2f)",
					nc.formatType(varTypeInfo.Type), varTypeInfo.Confidence)
				result = comment + "\n" + result
			}
		}
	}

	return result, nil
}

// getVariableTypeInfo looks up type info for a specific variable
func (nc *nodeCompiler) getVariableTypeInfo(variable *ast.VariableExpr) *types.TypeInfo {
	// Generate ID for this specific variable based on its position
	startPos := variable.Start()
	endPos := variable.End()
	varID := fmt.Sprintf("variable_declaration_%d_%d", startPos.Column, endPos.Column)

	if info, exists := nc.typeInfoMap[varID]; exists {
		return info
	}

	return nil
}

// formatVariableWithType formats a variable with its type annotation
func (nc *nodeCompiler) formatVariableWithType(variable *ast.VariableExpr, typeInfo *types.TypeInfo) string {
	typeStr := nc.formatType(typeInfo.Type)

	switch nc.options.AnnotationStyle {
	case StyleInline:
		return typeStr + " " + variable.FullName()
	case StyleCompact:
		return typeStr + " " + variable.FullName()
	case StyleVerbose:
		return typeStr + " " + variable.FullName()
	case StyleCommentOnly:
		return variable.FullName() + " # Type: " + typeStr
	default:
		return typeStr + " " + variable.FullName()
	}
}

// formatType converts a types.Type to its string representation
func (nc *nodeCompiler) formatType(t types.Type) string {
	if t == nil {
		return "Any"
	}

	// Use the Type's String() method for consistent formatting
	if stringer, ok := t.(fmt.Stringer); ok {
		typeStr := stringer.String()

		// Map internal type names to Perl type annotation format
		switch typeStr {
		case "int":
			return "Int"
		case "str":
			return "Str"
		case "bool":
			return "Bool"
		case "num":
			return "Num"
		case "any":
			return "Any"
		case "scalar":
			return "Scalar"
		case "arrayref":
			return "ArrayRef"
		case "hashref":
			return "HashRef"
		default:
			// For complex types, try to format them properly
			if strings.HasPrefix(typeStr, "arrayref[") {
				return "ArrayRef" + typeStr[8:] // Replace "arrayref" with "ArrayRef"
			}
			if strings.HasPrefix(typeStr, "hashref[") {
				return "HashRef" + typeStr[7:] // Replace "hashref" with "HashRef"
			}
			if strings.Contains(typeStr, "|") {
				// Union type - ensure proper formatting
				return "(" + typeStr + ")"
			}
			// Capitalize first letter for consistency
			if len(typeStr) > 0 {
				return strings.ToUpper(typeStr[:1]) + typeStr[1:]
			}
			return typeStr
		}
	}

	return "Any"
}

// compileBasicVarDecl compiles a variable declaration without type annotations
func (nc *nodeCompiler) compileBasicVarDecl(node *ast.VarDecl) (string, error) {
	var parts []string
	parts = append(parts, node.DeclType) // my, our, state

	// Add variable names
	var varNames []string
	for _, variable := range node.Variables() {
		varNames = append(varNames, variable.FullName())
	}

	if len(varNames) > 0 {
		parts = append(parts, strings.Join(varNames, ", "))
	}

	// Add initializer if present
	if node.Initializer != nil {
		parts = append(parts, "=")
		initCode, err := nc.compileNode(node.Initializer)
		if err != nil {
			return "", err
		}
		parts = append(parts, initCode)
	}

	return strings.Join(parts, " ") + ";", nil
}

// compileSubDecl compiles a subroutine declaration
func (nc *nodeCompiler) compileSubDecl(node *ast.SubDecl) (string, error) {
	// For methods, delegate to method compilation
	if node.IsMethod {
		methodDecl := &ast.MethodDecl{SubDecl: node}
		return nc.compileMethodDecl(methodDecl)
	}

	var parts []string

	// Add sub keyword
	if node.IsLexical {
		parts = append(parts, "my sub")
	} else {
		parts = append(parts, "sub")
	}

	// Add name
	parts = append(parts, node.Name)

	// Add parameters with type annotations
	paramList, err := nc.compileParametersWithTypes(node.Parameters())
	if err != nil {
		return "", err
	}
	parts = append(parts, fmt.Sprintf("(%s)", paramList))

	// Add return type if we have type info for it
	returnTypeInfo := nc.getReturnTypeInfo(node)
	if returnTypeInfo != nil && returnTypeInfo.Confidence >= nc.options.MinimumConfidence {
		returnTypeStr := nc.formatType(returnTypeInfo.Type)
		parts = append(parts, ":", returnTypeStr)
	}

	// Add body
	if node.Body != nil {
		bodyCode, err := nc.compileNode(node.Body)
		if err != nil {
			return "", err
		}
		parts = append(parts, bodyCode)
	}

	return strings.Join(parts, " "), nil
}

// compileMethodDecl compiles a method declaration
func (nc *nodeCompiler) compileMethodDecl(node *ast.MethodDecl) (string, error) {
	var parts []string

	// Add method keyword
	if node.IsLexical {
		parts = append(parts, "my method")
	} else {
		parts = append(parts, "method")
	}

	// Add name
	parts = append(parts, node.Name)

	// Add parameters with type annotations
	paramList, err := nc.compileParametersWithTypes(node.Parameters())
	if err != nil {
		return "", err
	}
	parts = append(parts, fmt.Sprintf("(%s)", paramList))

	// Add return type if we have type info for it
	returnTypeInfo := nc.getReturnTypeInfo(node.SubDecl)
	if returnTypeInfo != nil && returnTypeInfo.Confidence >= nc.options.MinimumConfidence {
		returnTypeStr := nc.formatType(returnTypeInfo.Type)
		parts = append(parts, ":", returnTypeStr)
	}

	// Add body
	if node.Body != nil {
		bodyCode, err := nc.compileNode(node.Body)
		if err != nil {
			return "", err
		}
		parts = append(parts, bodyCode)
	}

	return strings.Join(parts, " "), nil
}

// compileBasicMethodDecl compiles a method declaration without type annotations
func (nc *nodeCompiler) compileBasicMethodDecl(node *ast.MethodDecl) (string, error) {
	var parts []string

	// Add method keyword
	if node.IsLexical {
		parts = append(parts, "my method")
	} else {
		parts = append(parts, "method")
	}

	// Add name and parameters
	parts = append(parts, node.Name)

	paramList, err := nc.compileParametersWithTypes(node.Parameters())
	if err != nil {
		return "", err
	}
	parts = append(parts, fmt.Sprintf("(%s)", paramList))

	// Add body
	if node.Body != nil {
		bodyCode, err := nc.compileNode(node.Body)
		if err != nil {
			return "", err
		}
		parts = append(parts, bodyCode)
	}

	return strings.Join(parts, " "), nil
}

// compileFieldDecl compiles a field declaration
func (nc *nodeCompiler) compileFieldDecl(node *ast.FieldDecl) (string, error) {
	var parts []string
	parts = append(parts, "field")

	// Get type information for this field
	var fieldTypeInfo *types.TypeInfo
	if node.Variable != nil {
		fieldTypeInfo = nc.getFieldTypeInfo(node.Variable)
	}

	// Add field with type annotation if available
	if fieldTypeInfo != nil && fieldTypeInfo.Confidence >= nc.options.MinimumConfidence {
		typeStr := nc.formatType(fieldTypeInfo.Type)
		if node.Variable != nil {
			parts = append(parts, typeStr, node.Variable.FullName())
		} else {
			parts = append(parts, typeStr, "$"+node.Name)
		}
	} else {
		// No type info or low confidence
		if node.Variable != nil {
			parts = append(parts, node.Variable.FullName())
		} else {
			parts = append(parts, "$"+node.Name)
		}
	}

	// Add initializer if present
	if node.Initializer != nil {
		parts = append(parts, "=")
		initCode, err := nc.compileNode(node.Initializer)
		if err != nil {
			return "", err
		}
		parts = append(parts, initCode)
	}

	return strings.Join(parts, " ") + ";", nil
}

// compileBasicFieldDecl compiles a field declaration without type annotations
func (nc *nodeCompiler) compileBasicFieldDecl(node *ast.FieldDecl) (string, error) {
	var parts []string
	parts = append(parts, "field")

	if node.Variable != nil {
		parts = append(parts, node.Variable.FullName())
	} else {
		parts = append(parts, "$"+node.Name)
	}

	// Add initializer if present
	if node.Initializer != nil {
		parts = append(parts, "=")
		initCode, err := nc.compileNode(node.Initializer)
		if err != nil {
			return "", err
		}
		parts = append(parts, initCode)
	}

	return strings.Join(parts, " ") + ";", nil
}

// compileParametersWithTypes compiles a list of parameters with type annotations from inference
func (nc *nodeCompiler) compileParametersWithTypes(params []*ast.Parameter) (string, error) {
	var paramStrs []string

	for _, param := range params {
		paramStr := ""

		// Try to get inferred type info for this parameter
		paramTypeInfo := nc.getParameterTypeInfo(param)

		if paramTypeInfo != nil && paramTypeInfo.Confidence >= nc.options.MinimumConfidence {
			// Add inferred type annotation
			typeStr := nc.formatType(paramTypeInfo.Type)
			paramStr += typeStr + " "
		} else if param.TypeExpr != nil {
			// Fall back to explicit type annotation if present
			paramStr += param.TypeExpr.String() + " "
		}

		// Add parameter variable
		if param.Variable != nil {
			paramStr += param.Variable.FullName()
		} else {
			paramStr += "$" + param.Name
		}

		// Add default value if present
		if param.Default != nil {
			defaultCode, err := nc.compileNode(param.Default)
			if err != nil {
				return "", err
			}
			paramStr += " = " + defaultCode
		}

		paramStrs = append(paramStrs, paramStr)
	}

	return strings.Join(paramStrs, ", "), nil
}

// getParameterTypeInfo gets type info for a parameter based on its position
func (nc *nodeCompiler) getParameterTypeInfo(param *ast.Parameter) *types.TypeInfo {
	// Generate position-based ID for parameter
	if param.Variable != nil {
		startPos := param.Variable.Start()
		endPos := param.Variable.End()
		paramID := fmt.Sprintf("parameter_%d_%d", startPos.Column, endPos.Column)

		if info, exists := nc.typeInfoMap[paramID]; exists {
			return info
		}
	}

	return nil
}

// getReturnTypeInfo gets type info for a subroutine or method return type
func (nc *nodeCompiler) getReturnTypeInfo(node *ast.SubDecl) *types.TypeInfo {
	startPos := node.Start()
	endPos := node.End()

	var returnID string
	if node.IsMethod {
		returnID = fmt.Sprintf("method_return_%d_%d", startPos.Column, endPos.Column)
	} else {
		returnID = fmt.Sprintf("subroutine_return_%d_%d", startPos.Column, endPos.Column)
	}

	if info, exists := nc.typeInfoMap[returnID]; exists {
		return info
	}

	return nil
}

// getFieldTypeInfo gets type info for a field declaration
func (nc *nodeCompiler) getFieldTypeInfo(variable *ast.VariableExpr) *types.TypeInfo {
	startPos := variable.Start()
	endPos := variable.End()
	fieldID := fmt.Sprintf("field_declaration_%d_%d", startPos.Column, endPos.Column)

	if info, exists := nc.typeInfoMap[fieldID]; exists {
		return info
	}

	return nil
}

// Helper methods for other node types - delegating to existing implementations

// compileBlockStmt compiles a block statement
func (nc *nodeCompiler) compileBlockStmt(node *ast.BlockStmt) (string, error) {
	var parts []string
	parts = append(parts, "{")

	for _, stmt := range node.Statements() {
		compiled, err := nc.compileNode(stmt)
		if err != nil {
			return "", err
		}
		if compiled != "" {
			parts = append(parts, "  "+compiled)
		}
	}

	parts = append(parts, "}")
	return strings.Join(parts, "\n"), nil
}

// compileExpressionStmt compiles an expression statement
func (nc *nodeCompiler) compileExpressionStmt(node *ast.ExpressionStmt) (string, error) {
	compiled, err := nc.compileNode(node.Expression)
	if err != nil {
		return "", err
	}
	return compiled + ";", nil
}

// compileReturnStmt compiles a return statement
func (nc *nodeCompiler) compileReturnStmt(node *ast.ReturnStmt) (string, error) {
	if node.Value == nil {
		return "return;", nil
	}

	compiled, err := nc.compileNode(node.Value)
	if err != nil {
		return "", err
	}

	// Add type annotation for return value if configured
	typeInfo := nc.getNodeTypeInfo(node)
	annotation := ""
	if typeInfo != nil {
		annotation = nc.annotationGenerator.GenerateReturnAnnotation(node, typeInfo)
	}

	return "return " + compiled + ";" + annotation, nil
}

// compileIfStmt compiles an if statement
func (nc *nodeCompiler) compileIfStmt(node *ast.IfStmt) (string, error) {
	condition, err := nc.compileNode(node.Condition)
	if err != nil {
		return "", err
	}

	thenStmt, err := nc.compileNode(node.ThenStmt)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("if (%s) %s", condition, thenStmt)

	if node.ElseStmt != nil {
		elseStmt, err := nc.compileNode(node.ElseStmt)
		if err != nil {
			return "", err
		}
		result += " else " + elseStmt
	}

	return result, nil
}

// compileWhileStmt compiles a while statement
func (nc *nodeCompiler) compileWhileStmt(node *ast.WhileStmt) (string, error) {
	condition, err := nc.compileNode(node.Condition)
	if err != nil {
		return "", err
	}

	body, err := nc.compileNode(node.Body)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("while (%s) %s", condition, body), nil
}

// compileForStmt compiles a for statement
func (nc *nodeCompiler) compileForStmt(node *ast.ForStmt) (string, error) {
	iterator, err := nc.compileNode(node.Iterator)
	if err != nil {
		return "", err
	}

	iterable, err := nc.compileNode(node.Iterable)
	if err != nil {
		return "", err
	}

	body, err := nc.compileNode(node.Body)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("for %s (%s) %s", iterator, iterable, body), nil
}

// compileUseStmt compiles a use statement
func (nc *nodeCompiler) compileUseStmt(node *ast.UseStmt) (string, error) {
	result := "use " + node.Module

	if node.Version != "" {
		result += " " + node.Version
	}

	if len(node.Imports) > 0 {
		result += " qw(" + strings.Join(node.Imports, " ") + ")"
	}

	return result + ";", nil
}

// compilePackageStmt compiles a package statement
func (nc *nodeCompiler) compilePackageStmt(node *ast.PackageStmt) (string, error) {
	result := "package " + node.Name

	if node.Version != "" {
		result += " " + node.Version
	}

	return result + ";", nil
}

// compileGenericNode compiles a generic node by compiling its children
func (nc *nodeCompiler) compileGenericNode(node ast.Node) (string, error) {
	// For nodes we don't handle specifically, compile children and join with spaces
	var parts []string

	for _, child := range node.Children() {
		compiled, err := nc.compileNode(child)
		if err != nil {
			return "", err
		}
		if compiled != "" {
			parts = append(parts, compiled)
		}
	}

	return strings.Join(parts, " "), nil
}

// compileLiteralExpr compiles literal expressions
func (nc *nodeCompiler) compileLiteralExpr(node *ast.LiteralExpr) (string, error) {
	return node.Value, nil
}

// compileVariableExpr compiles variable expressions
func (nc *nodeCompiler) compileVariableExpr(node *ast.VariableExpr) (string, error) {
	return node.FullName(), nil
}

// Tree-sitter node compilation methods

// compileSourceFile compiles the root source_file node
func (nc *nodeCompiler) compileSourceFile(node ast.Node) (string, error) {
	var parts []string

	// Add Perl version pragma for modern features using PVM-managed version
	currentVersion, err := current.GetCurrentVersion()
	if err != nil {
		// Fallback to v5.36 if we can't get current version
		currentVersion = &current.CurrentVersionInfo{Version: "5.36"}
	}
	parts = append(parts, fmt.Sprintf("use v%s;", currentVersion.Version))

	// Compile all children nodes
	for _, child := range node.Children() {
		compiled, err := nc.compileNode(child)
		if err != nil {
			return "", err
		}
		if compiled != "" {
			parts = append(parts, compiled)
		}
	}

	return strings.Join(parts, "\n"), nil
}

// compileExpressionStatement compiles an expression_statement node
func (nc *nodeCompiler) compileExpressionStatement(node ast.Node) (string, error) {
	children := node.Children()
	if len(children) == 0 {
		return "", nil
	}

	// Compile the expression (first child)
	expr, err := nc.compileNode(children[0])
	if err != nil {
		return "", err
	}

	// Add semicolon if not already present
	if !strings.HasSuffix(expr, ";") {
		expr += ";"
	}

	return expr, nil
}

// compileVariableDeclaration compiles a variable_declaration node with optional type inference
func (nc *nodeCompiler) compileVariableDeclaration(node ast.Node) (string, error) {
	// Get type information for this variable
	typeInfo := nc.getNodeTypeInfo(node)

	if typeInfo != nil {
		// Generate annotated variable declaration
		return nc.generateTypedVariableFromTreeSitter(node, typeInfo), nil
	}

	// Fall back to basic variable declaration
	return nc.generateBasicVariableFromTreeSitter(node)
}

// compileTypedVariableDeclaration compiles a typed_variable_declaration node
func (nc *nodeCompiler) compileTypedVariableDeclaration(node ast.Node) (string, error) {
	// Typed variable declarations already have types, just preserve them
	return node.Text(), nil
}

// generateTypedVariableFromTreeSitter generates a typed variable declaration from tree-sitter node
func (nc *nodeCompiler) generateTypedVariableFromTreeSitter(node ast.Node, typeInfo *types.TypeInfo) string {
	nodeText := node.Text()

	// Extract variable components from the original text
	parts := strings.Fields(nodeText)
	if len(parts) < 2 {
		return nodeText // Fallback to original
	}

	declType := parts[0] // my, our, state
	varName := parts[1]  // $var, @array, %hash

	// Format the type annotation
	typeStr := typeInfo.Type.String()

	// Generate the typed declaration
	return fmt.Sprintf("%s %s %s", declType, typeStr, varName)
}

// generateBasicVariableFromTreeSitter generates an untyped variable declaration from tree-sitter node
func (nc *nodeCompiler) generateBasicVariableFromTreeSitter(node ast.Node) (string, error) {
	return node.Text(), nil
}

// compileSubroutineDeclaration compiles a subroutine_declaration_statement node
func (nc *nodeCompiler) compileSubroutineDeclaration(node ast.Node) (string, error) {
	// Extract subroutine information from tree-sitter node
	subInfo, err := nc.extractSubroutineInfo(node, false)
	if err != nil {
		// Fall back to original text if extraction fails
		return node.Text(), nil
	}

	// Generate typed signature if we have sufficient type information
	if nc.hasTypeInfoForSubroutine(subInfo) {
		return nc.generateTypedSubroutineSignature(subInfo)
	}

	// Fall back to original text if no type information
	return node.Text(), nil
}

// compileMethodDeclaration compiles a method_declaration_statement node
func (nc *nodeCompiler) compileMethodDeclaration(node ast.Node) (string, error) {
	// Extract method information from tree-sitter node
	subInfo, err := nc.extractSubroutineInfo(node, true)
	if err != nil {
		// Fall back to original text if extraction fails
		return node.Text(), nil
	}

	// Generate typed signature if we have sufficient type information
	if nc.hasTypeInfoForSubroutine(subInfo) {
		return nc.generateTypedSubroutineSignature(subInfo)
	}

	// Fall back to original text if no type information
	return node.Text(), nil
}

// compileClassStatement compiles a class_statement node
func (nc *nodeCompiler) compileClassStatement(node ast.Node) (string, error) {
	// Preserve class statements as-is
	return node.Text(), nil
}

// compileRoleStatement compiles a role_statement node
func (nc *nodeCompiler) compileRoleStatement(node ast.Node) (string, error) {
	// Preserve role statements as-is
	return node.Text(), nil
}

// compilePackageStatement compiles a package_statement node
func (nc *nodeCompiler) compilePackageStatement(node ast.Node) (string, error) {
	// Preserve package statements as-is
	return node.Text(), nil
}

// compileBlock compiles a block node
func (nc *nodeCompiler) compileBlock(node ast.Node) (string, error) {
	var parts []string

	// Compile all statements in the block
	for _, child := range node.Children() {
		compiled, err := nc.compileNode(child)
		if err != nil {
			return "", err
		}
		if compiled != "" {
			parts = append(parts, compiled)
		}
	}

	return "{\n" + strings.Join(parts, "\n") + "\n}", nil
}

// SubroutineInfo holds extracted information about a subroutine or method from tree-sitter
type SubroutineInfo struct {
	Name         string
	IsMethod     bool
	IsLexical    bool
	Parameters   []ParsedParameterInfo
	Body         string
	StartPos     ast.Position
	EndPos       ast.Position
	OriginalText string
}

// ParsedParameterInfo holds information about a parameter extracted from tree-sitter
type ParsedParameterInfo struct {
	Name     string
	Variable string // Full variable name like $self, $input
	StartPos ast.Position
	EndPos   ast.Position
}

// extractSubroutineInfo extracts subroutine or method information from a tree-sitter node
func (nc *nodeCompiler) extractSubroutineInfo(node ast.Node, isMethod bool) (*SubroutineInfo, error) {
	info := &SubroutineInfo{
		IsMethod:     isMethod,
		StartPos:     node.Start(),
		EndPos:       node.End(),
		OriginalText: node.Text(),
	}

	// Parse the original text to extract components
	text := node.Text()
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty subroutine text")
	}

	// Parse the first line for declaration
	declaration := lines[0]
	parts := strings.Fields(declaration)

	// Handle different declaration patterns
	var nameIndex int
	if len(parts) >= 2 {
		if parts[0] == "my" && (parts[1] == "sub" || parts[1] == "method") {
			info.IsLexical = true
			nameIndex = 2
		} else if parts[0] == "sub" || parts[0] == "method" {
			nameIndex = 1
		}
	}

	if nameIndex < len(parts) {
		// Extract name, removing any signature part
		name := parts[nameIndex]
		if parenIndex := strings.Index(name, "("); parenIndex >= 0 {
			info.Name = name[:parenIndex]
		} else {
			info.Name = name
		}
	}

	// Extract parameters from signature if present
	if sigStart := strings.Index(declaration, "("); sigStart >= 0 {
		if sigEnd := strings.Index(declaration[sigStart:], ")"); sigEnd >= 0 {
			signature := declaration[sigStart+1 : sigStart+sigEnd]
			info.Parameters = nc.parseParametersFromSignature(signature, info.StartPos)
		}
	}

	// Extract body (everything after the first line)
	if len(lines) > 1 {
		info.Body = strings.Join(lines[1:], "\n")
	}

	return info, nil
}

// parseParametersFromSignature parses parameter information from a signature string
func (nc *nodeCompiler) parseParametersFromSignature(signature string, basePos ast.Position) []ParsedParameterInfo {
	var params []ParsedParameterInfo
	if strings.TrimSpace(signature) == "" {
		return params
	}

	// Split by comma, but be careful of nested structures
	paramStrs := nc.splitParameters(signature)

	for i, paramStr := range paramStrs {
		paramStr = strings.TrimSpace(paramStr)
		if paramStr == "" {
			continue
		}

		param := ParsedParameterInfo{
			StartPos: ast.Position{
				Line:   basePos.Line,
				Column: basePos.Column + i*10, // Approximate position
				Offset: basePos.Offset + i*10,
			},
			EndPos: ast.Position{
				Line:   basePos.Line,
				Column: basePos.Column + (i+1)*10,
				Offset: basePos.Offset + (i+1)*10,
			},
		}

		// Extract variable name (look for $, @, % prefixed names)
		fields := strings.Fields(paramStr)
		for _, field := range fields {
			if strings.HasPrefix(field, "$") || strings.HasPrefix(field, "@") || strings.HasPrefix(field, "%") {
				param.Variable = field
				// Extract base name without sigil
				param.Name = field[1:]
				break
			}
		}

		// If no variable found, try to extract from the parameter string
		if param.Variable == "" && len(fields) > 0 {
			// Assume the last field is the variable
			lastField := fields[len(fields)-1]
			if strings.Contains(lastField, "$") {
				param.Variable = lastField
				param.Name = strings.TrimPrefix(lastField, "$")
			}
		}

		params = append(params, param)
	}

	return params
}

// splitParameters splits a parameter string by commas, handling nested structures
func (nc *nodeCompiler) splitParameters(signature string) []string {
	var params []string
	var current strings.Builder
	depth := 0

	for _, char := range signature {
		switch char {
		case '(':
			depth++
			current.WriteRune(char)
		case ')':
			depth--
			current.WriteRune(char)
		case ',':
			if depth == 0 {
				params = append(params, current.String())
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		params = append(params, current.String())
	}

	return params
}

// hasTypeInfoForSubroutine checks if we have sufficient type information for a subroutine
func (nc *nodeCompiler) hasTypeInfoForSubroutine(info *SubroutineInfo) bool {
	// Check if we have type info for any parameters
	for _, param := range info.Parameters {
		if typeInfo := nc.getParameterTypeInfoByPosition(param); typeInfo != nil {
			if typeInfo.Confidence >= nc.options.MinimumConfidence {
				return true
			}
		}
	}

	// Check if we have return type info
	if returnInfo := nc.getReturnTypeInfoByPosition(info); returnInfo != nil {
		if returnInfo.Confidence >= nc.options.MinimumConfidence {
			return true
		}
	}

	return false
}

// getParameterTypeInfoByPosition gets type info for a parameter by position
func (nc *nodeCompiler) getParameterTypeInfoByPosition(param ParsedParameterInfo) *types.TypeInfo {
	// Try different ID formats for parameter lookup
	paramID := fmt.Sprintf("parameter_%d_%d", param.StartPos.Column, param.EndPos.Column)
	if info, exists := nc.typeInfoMap[paramID]; exists {
		return info
	}

	// Try variable-based lookup
	if param.Variable != "" {
		varID := fmt.Sprintf("variable_declaration_%d_%d", param.StartPos.Column, param.EndPos.Column)
		if info, exists := nc.typeInfoMap[varID]; exists {
			return info
		}
	}

	return nil
}

// getReturnTypeInfoByPosition gets return type info for a subroutine by position
func (nc *nodeCompiler) getReturnTypeInfoByPosition(info *SubroutineInfo) *types.TypeInfo {
	var returnID string
	if info.IsMethod {
		returnID = fmt.Sprintf("method_return_%d_%d", info.StartPos.Column, info.EndPos.Column)
	} else {
		returnID = fmt.Sprintf("subroutine_return_%d_%d", info.StartPos.Column, info.EndPos.Column)
	}

	if typeInfo, exists := nc.typeInfoMap[returnID]; exists {
		return typeInfo
	}

	return nil
}

// generateTypedSubroutineSignature generates a typed signature for a subroutine or method
func (nc *nodeCompiler) generateTypedSubroutineSignature(info *SubroutineInfo) (string, error) {
	var parts []string

	// Add declaration keyword
	if info.IsLexical {
		if info.IsMethod {
			parts = append(parts, "my method")
		} else {
			parts = append(parts, "my sub")
		}
	} else {
		if info.IsMethod {
			parts = append(parts, "method")
		} else {
			parts = append(parts, "sub")
		}
	}

	// Add return type if available
	if returnInfo := nc.getReturnTypeInfoByPosition(info); returnInfo != nil {
		if returnInfo.Confidence >= nc.options.MinimumConfidence {
			returnType := nc.formatType(returnInfo.Type)
			parts = append(parts, returnType)
		}
	}

	// Add name
	parts = append(parts, info.Name)

	// Generate typed parameters
	paramList := nc.generateTypedParameterList(info)
	parts = append(parts, fmt.Sprintf("(%s)", paramList))

	// Add body
	if info.Body != "" {
		result := strings.Join(parts, " ") + " " + info.Body
		return result, nil
	}

	return strings.Join(parts, " ") + " { ... }", nil
}

// generateTypedParameterList generates a typed parameter list for a subroutine
func (nc *nodeCompiler) generateTypedParameterList(info *SubroutineInfo) string {
	var paramStrs []string

	for _, param := range info.Parameters {
		paramStr := ""

		// Try to get inferred type info for this parameter
		if paramTypeInfo := nc.getParameterTypeInfoByPosition(param); paramTypeInfo != nil {
			if paramTypeInfo.Confidence >= nc.options.MinimumConfidence {
				typeStr := nc.formatType(paramTypeInfo.Type)
				paramStr = typeStr + " " + param.Variable
			} else {
				paramStr = param.Variable
			}
		} else {
			// For methods, add Object type for $self if it's the first parameter
			if info.IsMethod && len(paramStrs) == 0 && param.Variable == "$self" {
				paramStr = "Object " + param.Variable
			} else {
				paramStr = param.Variable
			}
		}

		if paramStr != "" {
			paramStrs = append(paramStrs, paramStr)
		}
	}

	return strings.Join(paramStrs, ", ")
}
