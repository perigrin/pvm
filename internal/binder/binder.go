// ABOUTME: Main binder implementation that traverses AST and performs symbol binding.
// ABOUTME: Handles Perl's lexical scoping rules and variable/subroutine declarations following TypeScript-Go architecture.

package binder

import (
	"strings"

	"tamarou.com/pvm/internal/ast"
)

// DefaultBinder implements the Binder interface
type DefaultBinder struct {
	symbolTable *SymbolTable
}

// NewBinder creates a new default binder
func NewBinder() *DefaultBinder {
	return &DefaultBinder{}
}

// Bind performs symbol binding on an AST node
func (b *DefaultBinder) Bind(node ast.Node) (*SymbolTable, error) {
	b.symbolTable = NewSymbolTable()

	// Traverse the AST and bind symbols
	return b.symbolTable, b.visitNode(node)
}

// BindAST performs symbol binding on a parsed AST
func (b *DefaultBinder) BindAST(astTree *ast.AST) (*SymbolTable, error) {
	b.symbolTable = NewSymbolTable()

	// Traverse the root node
	if astTree.Root != nil {
		if err := b.visitNode(astTree.Root); err != nil {
			return nil, err
		}
	}

	return b.symbolTable, nil
}

// visitNode dispatches to specific node handlers
func (b *DefaultBinder) visitNode(node ast.Node) error {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.VarDecl:
		return b.bindVariableDeclaration(n)
	case *ast.SubDecl:
		return b.bindSubroutineDeclaration(n)
	case *ast.MethodDecl:
		return b.bindMethodDeclaration(n)
	case *ast.PackageStmt:
		return b.bindPackageDeclaration(n)
	case *ast.BlockStmt:
		return b.bindBlockStatement(n)
	case *ast.IfStmt:
		return b.bindIfStatement(n)
	case *ast.WhileStmt:
		return b.bindWhileStatement(n)
	case *ast.ForStmt:
		return b.bindForStatement(n)
	case *ast.UseStmt:
		return b.bindUseStatement(n)
	case *ast.ReturnStmt:
		return b.bindReturnStatement(n)
	case ast.ExpressionNode:
		return b.bindExpression(n)
	default:
		// For other node types, recursively visit children
		return b.visitChildren(node)
	}
}

// bindVariableDeclaration handles variable declarations (my, our, state)
func (b *DefaultBinder) bindVariableDeclaration(node *ast.VarDecl) error {
	// Process each variable in the declaration
	for _, variable := range node.Variables {
		// Determine symbol kind from variable type
		kind := b.getVariableSymbolKind(variable.Name)

		// Determine flags from declaration type
		flags := b.getVariableFlags(node.DeclType)

		// Extract type annotation if present
		typeAnnotation := ""
		if node.TypeExpr != nil {
			typeAnnotation = node.TypeExpr.String()
			flags |= SymbolFlagTypeAnnotated
		}

		// Create symbol
		symbol := &Symbol{
			Name:        b.stripSigil(variable.Name),
			Kind:        kind,
			Flags:       flags,
			Declaration: node,
			Type:        typeAnnotation,
			Position:    variable.Start(),
		}

		// Add to symbol table
		if err := b.symbolTable.AddSymbol(symbol); err != nil {
			return err
		}
	}

	// Visit initializer if present
	if node.Initializer != nil {
		return b.visitNode(node.Initializer)
	}

	return nil
}

// bindSubroutineDeclaration handles subroutine declarations
func (b *DefaultBinder) bindSubroutineDeclaration(node *ast.SubDecl) error {
	// Create symbol for the subroutine
	symbol := &Symbol{
		Name:        node.Name,
		Kind:        SymbolSubroutine,
		Flags:       SymbolFlagNone,
		Declaration: node,
		Position:    node.Start(),
	}

	// Add return type if present
	if node.ReturnType != nil {
		symbol.Type = node.ReturnType.String()
		symbol.Flags |= SymbolFlagTypeAnnotated
	}

	// Add to current scope
	if err := b.symbolTable.AddSymbol(symbol); err != nil {
		return err
	}

	// Enter subroutine scope
	b.symbolTable.EnterScope(ScopeSubroutine, node)

	// Bind parameters
	for _, param := range node.Parameters {
		if err := b.bindParameter(param); err != nil {
			return err
		}
	}

	// Bind body
	if node.Body != nil {
		if err := b.visitNode(node.Body); err != nil {
			return err
		}
	}

	// Exit subroutine scope with advanced features
	b.symbolTable.ExitScopeAdvanced()

	return nil
}

// bindMethodDeclaration handles method declarations
func (b *DefaultBinder) bindMethodDeclaration(node *ast.MethodDecl) error {
	// Create symbol for the method
	symbol := &Symbol{
		Name:        node.Name,
		Kind:        SymbolMethod,
		Flags:       SymbolFlagMethod,
		Declaration: node,
		Position:    node.Start(),
	}

	// Add return type if present
	if node.ReturnType != nil {
		symbol.Type = node.ReturnType.String()
		symbol.Flags |= SymbolFlagTypeAnnotated
	}

	// Add to current scope
	if err := b.symbolTable.AddSymbol(symbol); err != nil {
		return err
	}

	// Enter method scope
	b.symbolTable.EnterScope(ScopeMethod, node)

	// Bind parameters (including implicit $self)
	for _, param := range node.Parameters {
		if err := b.bindParameter(param); err != nil {
			return err
		}
	}

	// Bind body
	if node.Body != nil {
		if err := b.visitNode(node.Body); err != nil {
			return err
		}
	}

	// Exit method scope with advanced features
	b.symbolTable.ExitScopeAdvanced()

	return nil
}

// bindPackageDeclaration handles package declarations
func (b *DefaultBinder) bindPackageDeclaration(node *ast.PackageStmt) error {
	// Update current package
	b.symbolTable.SetPackage(node.Name)

	// Create package symbol
	symbol := &Symbol{
		Name:        node.Name,
		Kind:        SymbolPackage,
		Flags:       SymbolFlagNone,
		Declaration: node,
		Position:    node.Start(),
	}

	return b.symbolTable.AddSymbol(symbol)
}

// bindBlockStatement handles block statements that create new scopes
func (b *DefaultBinder) bindBlockStatement(node *ast.BlockStmt) error {
	// Enter block scope
	b.symbolTable.EnterScope(ScopeBlock, node)

	// Bind all statements in block
	for _, stmt := range node.Statements {
		if err := b.visitNode(stmt); err != nil {
			return err
		}
	}

	// Exit block scope with advanced features
	b.symbolTable.ExitScopeAdvanced()

	return nil
}

// bindIfStatement handles if statements
func (b *DefaultBinder) bindIfStatement(node *ast.IfStmt) error {
	// Bind condition
	if err := b.visitNode(node.Condition); err != nil {
		return err
	}

	// Bind then branch
	if err := b.visitNode(node.ThenStmt); err != nil {
		return err
	}

	// Bind else branch if present
	if node.ElseStmt != nil {
		return b.visitNode(node.ElseStmt)
	}

	return nil
}

// bindWhileStatement handles while loops
func (b *DefaultBinder) bindWhileStatement(node *ast.WhileStmt) error {
	// Bind condition
	if err := b.visitNode(node.Condition); err != nil {
		return err
	}

	// Bind body
	return b.visitNode(node.Body)
}

// bindForStatement handles for loops
func (b *DefaultBinder) bindForStatement(node *ast.ForStmt) error {
	// Enter block scope for loop variable
	b.symbolTable.EnterScope(ScopeBlock, node)

	// Bind iterator variable
	if node.Iterator != nil {
		if err := b.visitNode(node.Iterator); err != nil {
			return err
		}
	}

	// Bind iterable
	if node.Iterable != nil {
		if err := b.visitNode(node.Iterable); err != nil {
			return err
		}
	}

	// Bind body
	if err := b.visitNode(node.Body); err != nil {
		return err
	}

	// Exit loop scope with advanced features
	b.symbolTable.ExitScopeAdvanced()

	return nil
}

// bindUseStatement handles use/require statements
func (b *DefaultBinder) bindUseStatement(node *ast.UseStmt) error {
	// Create import symbol
	symbol := &Symbol{
		Name:        node.Module,
		Kind:        SymbolImport,
		Flags:       SymbolFlagImported,
		Declaration: node,
		Position:    node.Start(),
	}

	// Add to symbol table
	if err := b.symbolTable.AddSymbol(symbol); err != nil {
		return err
	}

	// Create placeholder module symbol table
	moduleTable := NewSymbolTable()
	moduleTable.Package = node.Module

	// Import the module
	return b.symbolTable.ImportModule(node.Module, moduleTable)
}

// bindReturnStatement handles return statements
func (b *DefaultBinder) bindReturnStatement(node *ast.ReturnStmt) error {
	// Visit the return value if present
	if node.Value != nil {
		return b.visitNode(node.Value)
	}
	return nil
}

// bindExpression handles expression nodes
func (b *DefaultBinder) bindExpression(node ast.ExpressionNode) error {
	// Handle specific expression types that may introduce symbols
	switch expr := node.(type) {
	case *ast.VariableExpr:
		return b.bindVariableExpression(expr)
	default:
		// For most expressions, we just need to visit children
		return b.visitChildren(node)
	}
}

// bindVariableExpression handles variable expressions (references)
func (b *DefaultBinder) bindVariableExpression(node *ast.VariableExpr) error {
	// Check if this is a package-qualified variable
	if strings.Contains(node.Name, "::") {
		// This is a package-qualified access - we don't create new symbols,
		// just verify resolution during later phases
		return nil
	}

	// For regular variable references, we don't create new symbols,
	// just ensure they can be resolved in later phases
	return nil
}

// bindParameter handles parameter declarations
func (b *DefaultBinder) bindParameter(param *ast.Parameter) error {
	// Determine symbol kind from parameter type
	kind := b.getVariableSymbolKind(param.Name)

	// Parameters are always lexical
	flags := SymbolFlagLexical

	// Add type annotation if present
	typeAnnotation := ""
	if param.TypeExpr != nil {
		typeAnnotation = param.TypeExpr.String()
		flags |= SymbolFlagTypeAnnotated
	}

	// Create symbol
	symbol := &Symbol{
		Name:        b.stripSigil(param.Name),
		Kind:        kind,
		Flags:       flags,
		Declaration: param.Variable, // Use the Variable field which is a Node
		Type:        typeAnnotation,
		Position:    param.Pos,
	}

	return b.symbolTable.AddSymbol(symbol)
}

// visitChildren recursively visits all child nodes
func (b *DefaultBinder) visitChildren(node ast.Node) error {
	if node == nil {
		return nil
	}

	// Visit all children
	for _, child := range node.Children() {
		if err := b.visitNode(child); err != nil {
			return err
		}
	}

	return nil
}

// Helper methods

// getVariableSymbolKind determines symbol kind from variable name sigil
func (b *DefaultBinder) getVariableSymbolKind(name string) SymbolKind {
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
func (b *DefaultBinder) getVariableFlags(declType string) SymbolFlags {
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
func (b *DefaultBinder) stripSigil(name string) string {
	if len(name) > 0 && strings.ContainsRune("$@%*", rune(name[0])) {
		return name[1:]
	}
	return name
}
