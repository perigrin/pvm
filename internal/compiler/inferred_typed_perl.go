// ABOUTME: Inferred typed Perl compiler that generates type annotations from inference results
// ABOUTME: Transforms untyped Perl into typed Perl while preserving all original behavior

package compiler

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/compiler/pipeline"
	"tamarou.com/pvm/internal/current"
	"tamarou.com/pvm/internal/inference"
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

	// Handle typed nil case (common Go interface gotcha)
	//nolint:sloppyTypeAssert // Type assertion from interface to concrete type is intentional
	if concreteAST, ok := inputAST.(*ast.AST); ok && concreteAST == nil {
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

	// Extract underlying *ast.AST from adapter if needed
	var astImpl *ast.AST

	// Try direct cast first
	//nolint:sloppyTypeAssert // Type assertion from interface to concrete type is intentional
	if directAST, ok := inputAST.(*ast.AST); ok {
		astImpl = directAST
		//nolint:sloppyTypeAssert // Type assertion from interface to concrete type is intentional
	} else if adapter, ok := inputAST.(*ParserASTAdapter); ok {
		// Extract from adapter
		astImpl = adapter.ast
	} else {
		return "", &CompilerError{
			Code:    "INVALID_AST_TYPE",
			Message: "AST must be *ast.AST or ParserASTAdapter for type inference",
		}
	}

	if astImpl == nil {
		return "", &CompilerError{
			Code:    "INVALID_AST",
			Message: "Underlying AST is nil",
		}
	}

	// Perform type inference on the AST
	inferenceEngine := inference.NewTypeInferenceEngine()
	inferredAST, err := inferenceEngine.InferTypes(astImpl)
	if err != nil {
		return "", &CompilerError{
			Code:    "INFERENCE_FAILED",
			Message: fmt.Sprintf("Type inference failed: %v", err),
		}
	}
	return itpc.CompileInferred(inferredAST)
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

	// For now, use a simplified approach similar to other compilers
	// Get the original source content
	content, err := inferredAST.GetContent()
	if err != nil {
		return "", &CompilerError{
			Code:    "COMPILATION_FAILED",
			Message: fmt.Sprintf("Failed to get source content: %v", err),
		}
	}

	// For integration testing, start with adding the pragma and preserving most content
	// Add actual type inference annotations based on inferredAST.GetAllTypeInfo()
	result := itpc.addPerlVersionPragma(content)

	// Apply type annotations based on inference results
	annotatedResult, err := itpc.addInferenceAnnotations(inferredAST, result)
	if err != nil {
		return "", &CompilerError{
			Code:    "ANNOTATION_FAILED",
			Message: fmt.Sprintf("Failed to add type annotations: %v", err),
		}
	}

	return annotatedResult, nil
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

// compileVarDecl compiles a variable declaration with type annotations
func (nc *nodeCompiler) compileVarDecl(node *ast.VarDecl) (string, error) {
	var parts []string

	// Start with declaration type (my, our, state)
	parts = append(parts, node.DeclType)

	// Process variables with potential type annotations
	var varDecls []string
	for _, variable := range node.Variables() {
		// Try to get type info for this specific variable
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
	// For now, preserve the original text
	// TODO: Add type annotations for parameters and return types
	return node.Text(), nil
}

// compileMethodDeclaration compiles a method_declaration_statement node
func (nc *nodeCompiler) compileMethodDeclaration(node ast.Node) (string, error) {
	// For now, preserve the original text
	// TODO: Add type annotations for parameters and return types
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
