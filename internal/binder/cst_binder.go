// ABOUTME: CST-based binder implementation that works directly with tree-sitter nodes
// ABOUTME: Eliminates lossy AST conversion by operating directly on the concrete syntax tree

package binder

import (
	"log"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/ast"
)

// CSTBinder implements symbol binding directly on tree-sitter CST nodes
type CSTBinder struct {
	symbolTable *SymbolTable
	poolManager *SymbolPoolManager
	content     []byte // Original source content for text extraction
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

// findTypeAnnotationForVariable finds the type annotation for a given variable
func (b *CSTBinder) findTypeAnnotationForVariable(varName string) string {
	if b.symbolTable.TypeAnnotations == nil {
		return ""
	}

	// Look through type annotations to find one matching this variable
	for _, annotation := range b.symbolTable.TypeAnnotations {
		if annotation.AnnotatedItem == varName && annotation.Kind == ast.VarAnnotation {
			if annotation.TypeExpression != nil {
				return annotation.TypeExpression.String()
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
	// TODO: Implement CST-based package binding
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
	// TODO: Implement CST-based if statement binding
	return b.visitCSTChildren(node)
}

func (b *CSTBinder) bindCSTWhileStatement(node *sitter.Node) error {
	// TODO: Implement CST-based while statement binding
	return b.visitCSTChildren(node)
}

func (b *CSTBinder) bindCSTForStatement(node *sitter.Node) error {
	// TODO: Implement CST-based for statement binding
	return b.visitCSTChildren(node)
}

func (b *CSTBinder) bindCSTUseStatement(node *sitter.Node) error {
	// TODO: Implement CST-based use statement binding
	return b.visitCSTChildren(node)
}

func (b *CSTBinder) bindCSTReturnStatement(node *sitter.Node) error {
	// TODO: Implement CST-based return statement binding
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
