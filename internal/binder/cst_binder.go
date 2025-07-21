// ABOUTME: CST-based binder implementation that works directly with tree-sitter nodes
// ABOUTME: Eliminates lossy AST conversion by operating directly on the concrete syntax tree

package binder

import (
	"fmt"
	"log"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/ast"
)

// CSTBinder implements symbol binding directly on tree-sitter CST nodes
type CSTBinder struct {
	symbolTable *SymbolTable
	poolManager *SymbolPoolManager
	content     []byte       // Original source content for text extraction
	currentNode *sitter.Node // Current CST node being processed
}

// NewCSTBinder creates a new CST-based binder
func NewCSTBinder() *CSTBinder {
	return &CSTBinder{
		poolManager: DefaultSymbolPoolManager(),
	}
}

// NewCSTBinderWithPool creates a new CST binder with a specific pool manager
func NewCSTBinderWithPool(poolManager *SymbolPoolManager) *CSTBinder {
	return &CSTBinder{
		poolManager: poolManager,
	}
}

// BindCST performs symbol binding on a tree-sitter CST node
func (b *CSTBinder) BindCST(root *sitter.Node, content []byte, typeAnnotations []*ast.TypeAnnotation) (*SymbolTable, error) {
	b.symbolTable = NewSymbolTableWithPool(b.poolManager, "main")
	b.content = content

	// Store type annotations for reference during binding
	if len(typeAnnotations) > 0 {
		b.symbolTable.TypeAnnotations = typeAnnotations
	}

	// Traverse the CST and bind symbols
	return b.symbolTable, b.visitCSTNode(root)
}

// visitCSTNode dispatches CST node handling based on node type
func (b *CSTBinder) visitCSTNode(node *sitter.Node) error {
	if node == nil {
		return nil
	}

	nodeType := node.Kind()

	// Debug logging
	if DebugScoping {
		log.Printf("[DEBUG] visitCSTNode: Visiting %s", nodeType)
	}

	// Handle specific CST node types that create symbols
	switch nodeType {
	case "variable_declaration":
		return b.bindCSTVariableDeclaration(node)
	case "subroutine_declaration_statement":
		return b.bindCSTSubroutineDeclaration(node)
	case "method_declaration_statement":
		return b.bindCSTMethodDeclaration(node)
	case "package_statement":
		return b.bindCSTPackageDeclaration(node)
	case "block":
		return b.bindCSTBlockStatement(node)
	case "if_statement", "unless_statement":
		return b.bindCSTIfStatement(node)
	case "while_statement", "until_statement":
		return b.bindCSTWhileStatement(node)
	case "for_statement", "foreach_statement":
		return b.bindCSTForStatement(node)
	case "use_statement":
		return b.bindCSTUseStatement(node)
	case "return_statement":
		return b.bindCSTReturnStatement(node)
	default:
		// For other node types, recursively visit children
		return b.visitCSTChildren(node)
	}
}

// bindCSTVariableDeclaration handles variable declarations from CST
func (b *CSTBinder) bindCSTVariableDeclaration(node *sitter.Node) error {
	nodeText := b.getNodeText(node)

	if DebugScoping {
		log.Printf("[DEBUG] bindCSTVariableDeclaration: Processing '%s'", nodeText)
	}

	// Set current node for type annotation extraction
	b.currentNode = node

	// Extract declaration type and variables from CST
	declType := b.extractDeclType(node)
	variables := b.extractVariablesFromCST(node)

	if DebugScoping {
		log.Printf("[DEBUG] bindCSTVariableDeclaration: Found %d variables with declType '%s'", len(variables), declType)
	}

	// Create symbols for each variable
	for _, varInfo := range variables {
		if !b.isValidSymbolName(varInfo.Name) {
			continue
		}

		// Skip built-in type names that might be incorrectly detected
		if b.isBuiltinTypeName(varInfo.Name) {
			if DebugScoping {
				log.Printf("[DEBUG] bindCSTVariableDeclaration: Skipping builtin type '%s' incorrectly detected as variable", varInfo.Name)
			}
			continue
		}

		// Determine symbol kind from variable name sigil
		kind := b.getVariableSymbolKind(varInfo.Name)

		// Determine flags from declaration type
		flags := b.getVariableFlags(declType)

		// Look for type annotation for this variable
		typeAnnotation := b.findTypeAnnotationForVariable(varInfo.Name)
		if typeAnnotation != "" {
			flags |= SymbolFlagTypeAnnotated
		}

		// Create symbol using pool manager
		symbol := b.poolManager.NewSymbol(
			b.stripSigil(varInfo.Name),
			kind,
			flags,
			nil, // No AST node for CST-based binding
			ast.Position{Line: int(varInfo.Line), Column: int(varInfo.Column)},
		)
		symbol.Type = typeAnnotation

		// Add to symbol table
		if err := b.symbolTable.AddSymbol(symbol); err != nil {
			return err
		}

		if DebugScoping {
			log.Printf("[DEBUG] bindCSTVariableDeclaration: Created symbol '%s' (%s)", symbol.Name, symbol.Kind.String())
		}
	}

	// Visit children for initializers, etc.
	return b.visitCSTChildren(node)
}

// VariableInfo holds information about a variable extracted from CST
type VariableInfo struct {
	Name   string
	Line   uint32
	Column uint32
}

// extractVariablesFromCST extracts variable information directly from CST node
func (b *CSTBinder) extractVariablesFromCST(node *sitter.Node) []VariableInfo {
	var variables []VariableInfo

	// Look for scalar, array, hash children that represent variables
	b.findVariableNodesInCST(node, &variables)

	return variables
}

// findVariableNodesInCST recursively finds variable nodes in CST
func (b *CSTBinder) findVariableNodesInCST(node *sitter.Node, variables *[]VariableInfo) {
	if node == nil {
		return
	}

	nodeType := node.Kind()

	// Check if this node represents a variable
	if nodeType == "scalar" || nodeType == "array" || nodeType == "hash" {
		text := b.getNodeText(node)

		// Only include if it starts with a sigil and looks like a variable
		if len(text) > 1 && (text[0] == '$' || text[0] == '@' || text[0] == '%') {
			// Skip if this looks like a built-in type name
			varName := text[1:] // Remove sigil
			if !b.isBuiltinTypeName(varName) {
				*variables = append(*variables, VariableInfo{
					Name:   text,
					Line:   1, // TODO: Calculate from byte position
					Column: 1,
				})

				if DebugScoping {
					log.Printf("[DEBUG] findVariableNodesInCST: Found variable '%s'", text)
				}
			}
		}
		return // Don't recurse into variable nodes
	}

	// Skip type expression nodes to avoid confusing types with variables
	if nodeType == "type_expression" {
		return
	}

	// Recurse into children
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		b.findVariableNodesInCST(child, variables)
	}
}

// extractDeclType extracts the declaration type (my, our, state) from CST
func (b *CSTBinder) extractDeclType(node *sitter.Node) string {
	// Look for declaration keywords in the node or its children
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child.Kind() == "my" || child.Kind() == "our" || child.Kind() == "state" {
			return child.Kind()
		}
	}

	// Fallback: extract from text
	text := b.getNodeText(node)
	switch {
	case strings.HasPrefix(text, "my "):
		return "my"
	case strings.HasPrefix(text, "our "):
		return "our"
	case strings.HasPrefix(text, "state "):
		return "state"
	default:
		return "my" // Default
	}
}

// findTypeAnnotationForVariable finds the type annotation for a given variable in the current CST context
func (b *CSTBinder) findTypeAnnotationForVariable(varName string) string {
	// For CST-based binding, we need to extract type annotations directly from the current node
	// This is called during processing of a variable_declaration node
	if b.currentNode == nil {
		return ""
	}

	// Look for type expression nodes in the current variable declaration
	return b.extractTypeFromCSTNode(b.currentNode, varName)
}

// extractTypeFromCSTNode extracts type annotation for a variable from CST node
func (b *CSTBinder) extractTypeFromCSTNode(node *sitter.Node, varName string) string {
	if node == nil {
		return ""
	}

	// For variable declarations like "my Int $var = 42;", we need to find
	// the type expression that precedes the variable
	return b.findTypeExpressionInDeclaration(node, varName)
}

// findTypeExpressionInDeclaration finds the type expression for a variable in a declaration node
func (b *CSTBinder) findTypeExpressionInDeclaration(node *sitter.Node, varName string) string {
	if node == nil {
		return ""
	}

	// Look for type_expression nodes in the current declaration
	var typeExpr string
	var currentVar string

	// Traverse children to find pattern: type_expression followed by variable
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		nodeType := child.Kind()
		nodeText := b.getNodeText(child)

		if DebugScoping {
			log.Printf("[DEBUG] findTypeExpressionInDeclaration: Examining child %d: type='%s', text='%s'", i, nodeType, nodeText)
		}

		// If we find a type_expression, remember it
		if nodeType == "type_expression" {
			typeExpr = nodeText
			if DebugScoping {
				log.Printf("[DEBUG] findTypeExpressionInDeclaration: Found type expression: '%s'", typeExpr)
			}
		}

		// If we find a variable that matches our target, return the last type expression we found
		if (nodeType == "scalar" || nodeType == "array" || nodeType == "hash") && nodeText == varName {
			currentVar = nodeText
			if DebugScoping {
				log.Printf("[DEBUG] findTypeExpressionInDeclaration: Found target variable '%s', returning type '%s'", currentVar, typeExpr)
			}
			return typeExpr
		}

		// Recurse into children if this isn't a terminal node
		if child.ChildCount() > 0 {
			result := b.findTypeExpressionInDeclaration(child, varName)
			if result != "" {
				return result
			}
		}
	}

	return ""
}

// getNodeText extracts text content from a CST node
func (b *CSTBinder) getNodeText(node *sitter.Node) string {
	if node == nil || b.content == nil {
		return ""
	}

	start := node.StartByte()
	end := node.EndByte()

	if start >= uint(len(b.content)) || end > uint(len(b.content)) {
		return ""
	}

	return string(b.content[start:end])
}

// extractSubroutineName extracts the subroutine name from a subroutine declaration node
func (b *CSTBinder) extractSubroutineName(node *sitter.Node) string {
	if node == nil {
		return ""
	}

	// Look for the name field in the subroutine declaration
	// The grammar defines: field('name', $.bareword)
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		childType := child.Kind()

		// Look for bareword nodes that contain the subroutine name
		if childType == "bareword" {
			return b.getNodeText(child)
		}

		// Alternative: look for identifier nodes
		if childType == "identifier" {
			return b.getNodeText(child)
		}
	}

	return ""
}

// extractMethodName extracts the method name from a method declaration node
func (b *CSTBinder) extractMethodName(node *sitter.Node) string {
	if node == nil {
		return ""
	}

	// Look for the name field in the method declaration
	// Similar to subroutine declarations: field('name', $.bareword)
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		childType := child.Kind()

		// Look for bareword nodes that contain the method name
		if childType == "bareword" {
			return b.getNodeText(child)
		}

		// Alternative: look for identifier nodes
		if childType == "identifier" {
			return b.getNodeText(child)
		}
	}

	return ""
}

// Placeholder implementations for other binding methods
// These will be implemented as we convert each node type

func (b *CSTBinder) bindCSTSubroutineDeclaration(node *sitter.Node) error {
	nodeText := b.getNodeText(node)

	if DebugScoping {
		log.Printf("[DEBUG] bindCSTSubroutineDeclaration: Processing '%s'", nodeText)
	}

	// Extract subroutine name from CST
	subName := b.extractSubroutineName(node)
	if subName == "" {
		// No valid subroutine name found, continue with children
		return b.visitCSTChildren(node)
	}

	if DebugScoping {
		log.Printf("[DEBUG] bindCSTSubroutineDeclaration: Found subroutine '%s'", subName)
	}

	// Get position from the node
	position := ast.Position{Line: int(node.StartPosition().Row) + 1, Column: int(node.StartPosition().Column) + 1}

	// Create symbol using pool manager
	symbol := b.poolManager.NewSymbol(
		subName,
		SymbolSubroutine,
		SymbolFlagNone,
		nil, // No AST node for CST-based binding
		position,
	)

	// Add to symbol table
	if err := b.symbolTable.AddSymbol(symbol); err != nil {
		return err
	}

	if DebugScoping {
		log.Printf("[DEBUG] bindCSTSubroutineDeclaration: Created symbol '%s' (%s)", symbol.Name, symbol.Kind.String())
	}

	// Visit children for parameters, body, etc.
	return b.visitCSTChildren(node)
}

func (b *CSTBinder) bindCSTMethodDeclaration(node *sitter.Node) error {
	nodeText := b.getNodeText(node)

	if DebugScoping {
		log.Printf("[DEBUG] bindCSTMethodDeclaration: Processing '%s'", nodeText)
	}

	// Extract method name from CST
	methodName := b.extractMethodName(node)
	if methodName == "" {
		// No valid method name found, continue with children
		return b.visitCSTChildren(node)
	}

	if DebugScoping {
		log.Printf("[DEBUG] bindCSTMethodDeclaration: Found method '%s'", methodName)
	}

	// Get position from the node
	position := ast.Position{Line: int(node.StartPosition().Row) + 1, Column: int(node.StartPosition().Column) + 1}

	// Create symbol using pool manager
	symbol := b.poolManager.NewSymbol(
		methodName,
		SymbolMethod,
		SymbolFlagNone,
		nil, // No AST node for CST-based binding
		position,
	)

	// Add to symbol table
	if err := b.symbolTable.AddSymbol(symbol); err != nil {
		return err
	}

	if DebugScoping {
		log.Printf("[DEBUG] bindCSTMethodDeclaration: Created symbol '%s' (%s)", symbol.Name, symbol.Kind.String())
	}

	// Visit children for parameters, body, etc.
	return b.visitCSTChildren(node)
}

func (b *CSTBinder) bindCSTPackageDeclaration(node *sitter.Node) error {
	if DebugScoping {
		fmt.Printf("Binding package declaration at position %+v\n", b.getNodePosition(node))
	}

	// Extract package name from the CST node
	packageName := b.extractPackageName(node)
	if packageName == "" {
		if DebugScoping {
			fmt.Printf("Warning: Could not extract package name from package declaration\n")
		}
		return b.visitCSTChildren(node)
	}

	if DebugScoping {
		fmt.Printf("Found package declaration: %s\n", packageName)
	}

	// Create package symbol
	symbol := b.poolManager.NewSymbol(
		packageName,
		SymbolPackage,
		SymbolFlagPackage,
		nil, // No AST node for CST binding
		b.getNodePosition(node),
	)

	// Add to symbol table
	if err := b.symbolTable.AddSymbol(symbol); err != nil {
		return fmt.Errorf("failed to add package symbol '%s': %w", packageName, err)
	}

	// Update package context in symbol table
	b.symbolTable.SetPackage(packageName)

	if DebugScoping {
		fmt.Printf("Successfully bound package '%s'\n", packageName)
	}

	// Continue binding children
	return b.visitCSTChildren(node)
}

func (b *CSTBinder) bindCSTBlockStatement(node *sitter.Node) error {
	// Enter block scope
	b.symbolTable.EnterScope(ScopeBlock, nil)

	// Bind all children in block
	err := b.visitCSTChildren(node)

	// Exit block scope
	b.symbolTable.ExitScope()

	return err
}

func (b *CSTBinder) bindCSTIfStatement(node *sitter.Node) error {
	if DebugScoping {
		fmt.Printf("Binding if statement at position %+v\n", b.getNodePosition(node))
	}

	// Enter block scope for the if statement
	b.symbolTable.EnterScope(ScopeBlock, nil)

	if DebugScoping {
		fmt.Printf("Entered block scope for if statement\n")
	}

	// Bind condition and body - this will handle any variable declarations
	// or references within the if statement
	err := b.visitCSTChildren(node)

	// Exit block scope
	b.symbolTable.ExitScope()

	if DebugScoping {
		fmt.Printf("Exited block scope for if statement\n")
	}

	return err
}

func (b *CSTBinder) bindCSTWhileStatement(node *sitter.Node) error {
	if DebugScoping {
		fmt.Printf("Binding while statement at position %+v\n", b.getNodePosition(node))
	}

	// Enter block scope for the while loop
	b.symbolTable.EnterScope(ScopeBlock, nil)

	if DebugScoping {
		fmt.Printf("Entered block scope for while statement\n")
	}

	// Bind condition and body - this will handle any variable declarations
	// or references within the while statement
	err := b.visitCSTChildren(node)

	// Exit block scope
	b.symbolTable.ExitScope()

	if DebugScoping {
		fmt.Printf("Exited block scope for while statement\n")
	}

	return err
}

func (b *CSTBinder) bindCSTForStatement(node *sitter.Node) error {
	if DebugScoping {
		fmt.Printf("Binding for statement at position %+v\n", b.getNodePosition(node))
	}

	// Enter block scope for the for loop
	b.symbolTable.EnterScope(ScopeBlock, nil)

	if DebugScoping {
		fmt.Printf("Entered block scope for for statement\n")
	}

	// Handle loop variable binding - for statements may introduce
	// new variables (e.g., for my $var (@array))
	if err := b.bindLoopVariable(node); err != nil {
		b.symbolTable.ExitScope()
		return fmt.Errorf("failed to bind loop variable in for statement: %w", err)
	}

	// Bind the rest of the statement (condition, body, etc.)
	err := b.visitCSTChildren(node)

	// Exit block scope
	b.symbolTable.ExitScope()

	if DebugScoping {
		fmt.Printf("Exited block scope for for statement\n")
	}

	return err
}

func (b *CSTBinder) bindCSTUseStatement(node *sitter.Node) error {
	if DebugScoping {
		fmt.Printf("Binding use statement at position %+v\n", b.getNodePosition(node))
	}

	// Extract module name from the CST node
	moduleName := b.extractModuleName(node)
	if moduleName == "" {
		if DebugScoping {
			fmt.Printf("Warning: Could not extract module name from use statement\n")
		}
		return b.visitCSTChildren(node)
	}

	if DebugScoping {
		fmt.Printf("Found use statement: %s\n", moduleName)
	}

	// Create import symbol
	symbol := b.poolManager.NewSymbol(
		moduleName,
		SymbolImport,
		SymbolFlagImported,
		nil, // No AST node for CST binding
		b.getNodePosition(node),
	)

	// Add to symbol table
	if err := b.symbolTable.AddSymbol(symbol); err != nil {
		return fmt.Errorf("failed to add import symbol '%s': %w", moduleName, err)
	}

	// Add the import to the current scope's import tracking
	if b.symbolTable.CurrentScope.ImportedModules == nil {
		b.symbolTable.CurrentScope.ImportedModules = make(map[string]string)
	}
	b.symbolTable.CurrentScope.ImportedModules[moduleName] = moduleName

	if DebugScoping {
		fmt.Printf("Successfully bound use statement for module '%s'\n", moduleName)
	}

	// Continue binding children (for version specs, import lists, etc.)
	return b.visitCSTChildren(node)
}

func (b *CSTBinder) bindCSTReturnStatement(node *sitter.Node) error {
	if DebugScoping {
		fmt.Printf("Binding return statement at position %+v\n", b.getNodePosition(node))
	}

	// Return statements don't create new scopes, just process their children
	// This will bind any variables referenced in the return expression
	return b.visitCSTChildren(node)
}

// visitCSTChildren recursively visits all child nodes
func (b *CSTBinder) visitCSTChildren(node *sitter.Node) error {
	if node == nil {
		return nil
	}

	// Visit all children
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if err := b.visitCSTNode(child); err != nil {
			return err
		}
	}

	return nil
}

// Helper methods (reused from original binder)

// isBuiltinTypeName checks if a name is a built-in type that shouldn't be treated as a variable
func (b *CSTBinder) isBuiltinTypeName(name string) bool {
	builtinTypes := map[string]bool{
		"Int":       true,
		"Str":       true,
		"Num":       true,
		"Bool":      true,
		"ArrayRef":  true,
		"HashRef":   true,
		"CodeRef":   true,
		"ScalarRef": true,
		"Optional":  true,
		"Void":      true,
		"Any":       true,
		"Undef":     true,
		"Object":    true,
		"DateTime":  true,
		"Map":       true,
		"Array":     true,
		"Hash":      true,
		"Scalar":    true,
		"Ref":       true,
		"Tuple":     true,
		"Union":     true,
		"Result":    true,
		"Maybe":     true,
		"Either":    true,
		"List":      true,
		"Set":       true,
		"IO":        true,
		"File":      true,
	}
	return builtinTypes[name]
}

// getVariableSymbolKind determines symbol kind from variable name sigil
func (b *CSTBinder) getVariableSymbolKind(name string) SymbolKind {
	if len(name) == 0 {
		return SymbolScalar
	}

	switch name[0] {
	case '$':
		return SymbolScalar
	case '@':
		return SymbolArray
	case '%':
		return SymbolHash
	case '*':
		return SymbolGlob
	default:
		return SymbolScalar
	}
}

// getVariableFlags determines symbol flags from declaration type
func (b *CSTBinder) getVariableFlags(declType string) SymbolFlags {
	switch declType {
	case "my", "state":
		return SymbolFlagLexical
	case "our":
		return SymbolFlagPackage
	default:
		return SymbolFlagNone
	}
}

// stripSigil removes the sigil from a variable name
func (b *CSTBinder) stripSigil(name string) string {
	if len(name) > 0 && strings.ContainsRune("$@%*", rune(name[0])) {
		return name[1:]
	}
	return name
}

// isValidSymbolName checks if a symbol name is valid for Perl
func (b *CSTBinder) isValidSymbolName(name string) bool {
	if name == "" {
		return false
	}

	// Strip sigil for validation
	cleanName := b.stripSigil(name)
	if cleanName == "" {
		return false
	}

	// Check for invalid characters that suggest malformed parsing
	// Valid Perl identifiers contain only alphanumeric characters, underscores, and colons
	for _, char := range cleanName {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == ':') {
			return false
		}
	}

	// Check that first character is not a digit
	firstChar := rune(cleanName[0])
	if firstChar >= '0' && firstChar <= '9' {
		return false
	}

	return true
}

// getNodePosition creates an AST position from a tree-sitter node
func (b *CSTBinder) getNodePosition(node *sitter.Node) ast.Position {
	if node == nil {
		return ast.Position{}
	}

	startPos := node.StartPosition()
	return ast.Position{
		Line:   int(startPos.Row) + 1,    // Convert 0-based to 1-based
		Column: int(startPos.Column) + 1, // Convert 0-based to 1-based
		Offset: int(node.StartByte()),
	}
}

// extractPackageName extracts the package name from a package_statement node
func (b *CSTBinder) extractPackageName(node *sitter.Node) string {
	if node == nil {
		return ""
	}

	// Look for the name field in the package declaration
	// The grammar defines: field('name', $.package)
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		childType := child.Kind()

		// Look for package name nodes
		if childType == "package" || childType == "bareword" || childType == "identifier" {
			text := b.getNodeText(child)
			if text != "" && text != "package" {
				return text
			}
		}

		// If it's a nested structure, recurse
		if child.ChildCount() > 0 {
			if name := b.extractPackageName(child); name != "" {
				return name
			}
		}
	}

	return ""
}

// extractModuleName extracts the module name from a use_statement node
func (b *CSTBinder) extractModuleName(node *sitter.Node) string {
	if node == nil {
		return ""
	}

	// Look for the module field in the use declaration
	// The grammar defines: field('module', $.package)
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		childType := child.Kind()

		// Look for module name nodes
		if childType == "package" || childType == "bareword" || childType == "identifier" {
			text := b.getNodeText(child)
			if text != "" && text != "use" && text != ";" {
				return text
			}
		}

		// If it's a nested structure, recurse
		if child.ChildCount() > 0 {
			if name := b.extractModuleName(child); name != "" {
				return name
			}
		}
	}

	return ""
}

// bindLoopVariable handles binding of loop variables in for statements
func (b *CSTBinder) bindLoopVariable(node *sitter.Node) error {
	if node == nil {
		return nil
	}

	// Look for loop variable declarations in the for statement
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(uint(i))
		if child == nil {
			continue
		}

		childType := child.Kind()

		// Look for variable declarations or variable expressions
		if childType == "variable_declaration" {
			return b.bindCSTVariableDeclaration(child)
		}

		if strings.HasSuffix(childType, "variable") && strings.HasPrefix(b.getNodeText(child), "$") {
			// This is likely a loop variable, create a symbol for it
			varName := b.getNodeText(child)
			if b.isValidSymbolName(varName) {
				symbol := b.poolManager.NewSymbol(
					b.stripSigil(varName),
					SymbolScalar,
					SymbolFlagLexical,
					nil,
					b.getNodePosition(child),
				)

				if err := b.symbolTable.AddSymbol(symbol); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
