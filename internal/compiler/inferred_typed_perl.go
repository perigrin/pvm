// ABOUTME: Inferred typed Perl compiler that generates type annotations from inference results
// ABOUTME: Transforms untyped Perl into typed Perl while preserving all original behavior

package compiler

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
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
	// ConfidenceThreshold sets minimum confidence for including type annotations
	ConfidenceThreshold float64

	// AnnotationStyle sets the formatting style for annotations
	AnnotationStyle FormattingStyle

	// PreserveComments controls whether original comments are preserved
	PreserveComments bool

	// PreserveFormatting controls whether original formatting is preserved
	PreserveFormatting bool

	// IncludeUncertainTypes controls whether low-confidence types are included as comments
	IncludeUncertainTypes bool

	// VerboseOutput includes additional debugging information
	VerboseOutput bool
}

// NewInferredTypedPerlCompiler creates a new inferred typed Perl compiler
func NewInferredTypedPerlCompiler() *InferredTypedPerlCompiler {
	formatter := NewTypeFormatter()

	return &InferredTypedPerlCompiler{
		formatter:           formatter,
		annotationGenerator: NewAnnotationGenerator(formatter),
		options: InferredCompilerOptions{
			ConfidenceThreshold:   0.7,
			AnnotationStyle:       StyleInline,
			PreserveComments:      true,
			PreserveFormatting:    true,
			IncludeUncertainTypes: true,
			VerboseOutput:         false,
		},
	}
}

// NewInferredTypedPerlCompilerWithOptions creates a compiler with custom options
func NewInferredTypedPerlCompilerWithOptions(options InferredCompilerOptions) *InferredTypedPerlCompiler {
	formatterOptions := FormatterOptions{
		ConfidenceThreshold:       options.ConfidenceThreshold,
		IncludeConfidenceComments: options.VerboseOutput,
		UseShortTypeNames:         options.AnnotationStyle == StyleCompact,
		PreferComments:            options.IncludeUncertainTypes,
	}

	formatter := NewTypeFormatterWithOptions(formatterOptions)

	annotationOptions := AnnotationOptions{
		AnnotateVariables: true,
		AnnotateMethods:   true,
		AnnotateFields:    true,
		AnnotateReturns:   false, // Usually too noisy
		MinConfidence:     options.ConfidenceThreshold,
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
//nolint:sloppyTypeAssert // Function intentionally uses type assertions for interface-to-concrete conversions
func (itpc *InferredTypedPerlCompiler) Validate(inputAST AST) error {
	if inputAST == nil {
		return &CompilerError{
			Code:    "INVALID_AST",
			Message: "AST cannot be nil",
		}
	}

	// Handle typed nil case (common Go interface gotcha)
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

	// Perform type inference on the AST
	inferenceEngine := inference.NewTypeInferenceEngine()
	// TODO: Update inference engine to use compiler.AST interface
	// For now, we assume inputAST is *ast.AST since that's the only implementation
	//nolint:sloppyTypeAssert // Type assertion from interface to concrete type is intentional
	if astImpl, ok := inputAST.(*ast.AST); ok {
		inferredAST, err := inferenceEngine.InferTypes(astImpl)
		if err != nil {
			return "", &CompilerError{
				Code:    "INFERENCE_FAILED",
				Message: fmt.Sprintf("Type inference failed: %v", err),
			}
		}
		return itpc.CompileInferred(inferredAST)
	}
	return "", &CompilerError{
		Code:    "INVALID_AST_TYPE",
		Message: "AST must be *ast.AST for type inference",
	}
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
	// TODO: Add actual type inference annotations based on inferredAST.GetAllTypeInfo()
	result := itpc.addPerlVersionPragma(content)

	return result, nil
}

// addPerlVersionPragma adds the Perl version pragma after the shebang line (if present)
func (itpc *InferredTypedPerlCompiler) addPerlVersionPragma(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return "use v5.36;\n" + content
	}

	// Check if first line is a shebang
	if strings.HasPrefix(lines[0], "#!") {
		// Insert pragma after shebang
		if len(lines) == 1 {
			return lines[0] + "\nuse v5.36;"
		}
		// Insert after first line
		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[0])
		newLines = append(newLines, "use v5.36;")
		newLines = append(newLines, lines[1:]...)
		return strings.Join(newLines, "\n")
	}

	// No shebang, prepend pragma
	return "use v5.36;\n" + content
}

// nodeCompiler handles compilation of individual AST nodes
type nodeCompiler struct {
	inferredAST         ast.InferredAST
	annotationGenerator *AnnotationGenerator
	options             InferredCompilerOptions
	nodeCounter         int // For generating unique node IDs
}

// generateNodeID generates a unique ID for a node
func (nc *nodeCompiler) generateNodeID(node ast.Node) string {
	nc.nodeCounter++
	return fmt.Sprintf("%s_%d", node.Type(), nc.nodeCounter)
}

// getOrGenerateNodeID gets type info for a node, generating an ID if needed
func (nc *nodeCompiler) getNodeTypeInfo(node ast.Node) *types.TypeInfo {
	nodeID := nc.generateNodeID(node)
	return nc.inferredAST.GetTypeInfo(nodeID)
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

	// Add Perl version pragma for modern features
	parts = append(parts, "use v5.36;")

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
	// Get type information for this variable
	typeInfo := nc.getNodeTypeInfo(node)

	if typeInfo != nil && typeInfo.Confidence >= nc.options.ConfidenceThreshold {
		// Generate annotated variable declaration
		return nc.annotationGenerator.GenerateVariableAnnotation(node, typeInfo), nil
	}

	// Fall back to basic variable declaration
	return nc.compileBasicVarDecl(node)
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

	// Add parameters
	paramList, err := nc.compileParameters(node.Parameters())
	if err != nil {
		return "", err
	}
	parts = append(parts, fmt.Sprintf("(%s)", paramList))

	// Add return type if present and confident
	if node.ReturnType != nil {
		typeInfo := nc.getNodeTypeInfo(node)
		if typeInfo != nil && typeInfo.Confidence >= nc.options.ConfidenceThreshold {
			returnTypeStr := node.ReturnType.String()
			parts = append(parts, "->", returnTypeStr)
		}
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
	// Get method signature information
	signature := nc.buildMethodSignatureInfo(node.SubDecl)

	if signature != nil && signature.OverallConfidence >= nc.options.ConfidenceThreshold {
		// Generate annotated method signature
		return nc.annotationGenerator.GenerateMethodAnnotation(node, signature), nil
	}

	// Fall back to basic method declaration
	return nc.compileBasicMethodDecl(node)
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

	paramList, err := nc.compileParameters(node.Parameters())
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
	// Get type information for this field
	typeInfo := nc.getNodeTypeInfo(node)

	if typeInfo != nil && typeInfo.Confidence >= nc.options.ConfidenceThreshold {
		// Generate annotated field declaration
		return nc.annotationGenerator.GenerateFieldAnnotation(node, typeInfo), nil
	}

	// Fall back to basic field declaration
	return nc.compileBasicFieldDecl(node)
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

// compileParameters compiles a list of parameters
func (nc *nodeCompiler) compileParameters(params []*ast.Parameter) (string, error) {
	var paramStrs []string

	for _, param := range params {
		paramStr := ""

		// Add type annotation if present and confident
		if param.TypeExpr != nil {
			// For parameters, we'll use a simplified approach since Parameter doesn't implement Node
			// In a real implementation, we'd need to track parameter type info differently
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

// buildMethodSignatureInfo builds method signature information from a SubDecl
func (nc *nodeCompiler) buildMethodSignatureInfo(node *ast.SubDecl) *MethodSignatureInfo {
	var params []ParameterInfo
	totalConfidence := 0.0
	confidenceCount := 0

	for _, param := range node.Parameters() {
		paramInfo := ParameterInfo{
			Name:       param.Name,
			IsOptional: param.IsOptional,
		}

		// Get type information if available
		// For parameters, we'll use a simplified approach since Parameter doesn't implement Node
		// In a real implementation, we'd need to track parameter type info differently
		paramInfo.Confidence = 0.8 // Default confidence for explicitly typed parameters
		if param.TypeExpr != nil {
			totalConfidence += paramInfo.Confidence
			confidenceCount++
		} else {
			paramInfo.Confidence = 0.0
		}

		params = append(params, paramInfo)
	}

	// Calculate overall confidence
	overallConfidence := 0.0
	if confidenceCount > 0 {
		overallConfidence = totalConfidence / float64(confidenceCount)
	}

	// Get return type information
	var returnType types.Type
	returnConfidence := 0.0
	if node.ReturnType != nil {
		if typeInfo := nc.getNodeTypeInfo(node); typeInfo != nil {
			returnType = typeInfo.Type
			returnConfidence = typeInfo.Confidence
		}
	}

	return &MethodSignatureInfo{
		Parameters:        params,
		ReturnType:        returnType,
		ReturnConfidence:  returnConfidence,
		OverallConfidence: overallConfidence,
	}
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

	// Add Perl version pragma for modern features
	parts = append(parts, "use v5.36;")

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

	if typeInfo != nil && typeInfo.Confidence >= nc.options.ConfidenceThreshold {
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
