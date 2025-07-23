// ABOUTME: Advanced symbol binding features for Perl's complex scoping scenarios.
// ABOUTME: Handles closures, dynamic scoping, module imports, and package qualification.

package binder

import (
	"log"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/ast"
)

// AdvancedBinder provides enhanced CST-based symbol binding with advanced Perl features
type AdvancedBinder struct {
	*CSTBinder
}

// NewAdvancedBinder creates a new advanced binder
func NewAdvancedBinder() *AdvancedBinder {
	return &AdvancedBinder{
		CSTBinder: NewCSTBinder(),
	}
}

// BindCST performs advanced CST-based symbol binding
func (ab *AdvancedBinder) BindCST(root *sitter.Node, content []byte, typeAnnotations []*ast.TypeAnnotation) (*SymbolTable, error) {
	// First perform standard CST binding
	symbolTable, err := ab.CSTBinder.BindCST(root, content, typeAnnotations)
	if err != nil {
		return nil, err
	}

	// Add advanced binding passes
	if err := ab.processAdvancedFeatures(root, content, symbolTable); err != nil {
		return nil, err
	}

	return symbolTable, nil
}

// processAdvancedFeatures processes advanced Perl features in the CST
func (ab *AdvancedBinder) processAdvancedFeatures(root *sitter.Node, content []byte, st *SymbolTable) error {
	// Process package qualification, closures, and local variables
	return ab.walkCSTForAdvancedFeatures(root, content, st)
}

// walkCSTForAdvancedFeatures walks CST nodes for advanced feature processing
func (ab *AdvancedBinder) walkCSTForAdvancedFeatures(node *sitter.Node, content []byte, st *SymbolTable) error {
	if node == nil {
		return nil
	}

	nodeType := node.Kind()
	switch nodeType {
	case "use_statement":
		if err := ab.processUseStatement(node, content, st); err != nil {
			return err
		}
	case "local_variable_declaration":
		if err := ab.processLocalVariable(node, content, st); err != nil {
			return err
		}
	case "package_qualified_access":
		if err := ab.processPackageQualifiedAccess(node, content, st); err != nil {
			return err
		}
	case "subroutine_declaration_statement":
		if err := ab.analyzeClosureCapture(node, content, st); err != nil {
			return err
		}
	}

	// Recursively process child nodes
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if err := ab.walkCSTForAdvancedFeatures(child, content, st); err != nil {
			return err
		}
	}

	return nil
}

// processUseStatement processes use/require statements from CST
func (ab *AdvancedBinder) processUseStatement(node *sitter.Node, content []byte, st *SymbolTable) error {
	// Extract module name from CST node
	moduleName := ab.extractModuleName(node, content)
	if moduleName == "" {
		return nil // Skip if we can't extract module name
	}

	// Create a placeholder module symbol table
	// In a real implementation, this would load the actual module
	moduleTable := NewSymbolTable()
	moduleTable.Package = moduleName

	// Import the module
	err := st.ImportModule(moduleName, moduleTable)
	if err != nil {
		return err
	}

	// Create import symbol
	pos := ab.nodeToPosition(node)
	importSymbol := &Symbol{
		Name:     moduleName,
		Kind:     SymbolImport,
		Flags:    SymbolFlagImported,
		Package:  st.Package,
		Position: pos,
	}

	return st.AddSymbol(importSymbol)
}

// extractModuleName extracts module name from use statement CST node
func (ab *AdvancedBinder) extractModuleName(node *sitter.Node, content []byte) string {
	// Look for module name in use statement
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child.Kind() == "bareword" {
			return ab.nodeText(child, content)
		}
	}
	return ""
}

// processLocalVariable processes local variable declarations from CST
func (ab *AdvancedBinder) processLocalVariable(node *sitter.Node, content []byte, st *SymbolTable) error {
	// Extract variable information from CST node
	varName, kind := ab.extractVariableInfo(node, content)
	if varName == "" {
		return nil // Skip if we can't extract variable info
	}

	pos := ab.nodeToPosition(node)
	symbol := &Symbol{
		Name:     ab.stripSigil(varName),
		Kind:     kind,
		Flags:    SymbolFlagLocal,
		Position: pos,
		Scope:    st.CurrentScope,
		Package:  st.Package,
	}

	return st.CreateLocalSymbol(symbol)
}

// extractVariableInfo extracts variable name and kind from CST node
func (ab *AdvancedBinder) extractVariableInfo(node *sitter.Node, content []byte) (string, SymbolKind) {
	// Look for variable in local declaration
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child.Kind() == "variable" {
			varText := ab.nodeText(child, content)
			return varText, ab.getSymbolKindFromSigil(varText)
		}
	}
	return "", SymbolScalar
}

// processPackageQualifiedAccess processes package qualified variable access from CST
func (ab *AdvancedBinder) processPackageQualifiedAccess(node *sitter.Node, content []byte, st *SymbolTable) error {
	// Extract qualified name from CST node
	qualifiedName := ab.extractQualifiedName(node, content)
	if qualifiedName == "" {
		return nil // Skip if we can't extract qualified name
	}

	// Resolve the qualified symbol
	kind := ab.getSymbolKindFromSigil(qualifiedName)
	symbol := st.ResolveWithPackageQualification(qualifiedName, kind)

	// If not found, create a reference symbol
	if symbol == nil {
		pos := ab.nodeToPosition(node)
		symbol = &Symbol{
			Name:          ab.stripSigil(qualifiedName),
			Kind:          kind,
			Flags:         SymbolFlagPackageQualified,
			Position:      pos,
			Scope:         st.CurrentScope,
			Package:       st.Package,
			QualifiedName: qualifiedName,
		}
		st.AddSymbol(symbol)
	}

	return nil
}

// extractQualifiedName extracts package qualified name from CST node
func (ab *AdvancedBinder) extractQualifiedName(node *sitter.Node, content []byte) string {
	// Extract text from qualified variable access node
	return ab.nodeText(node, content)
}

// analyzeClosureCapture analyzes closure variable capture for subroutines
func (ab *AdvancedBinder) analyzeClosureCapture(node *sitter.Node, content []byte, st *SymbolTable) error {
	// Find the scope for this subroutine
	if st.CurrentScope == nil {
		return nil
	}

	// Only analyze if this is a subroutine or method scope
	if st.CurrentScope.Kind != ScopeSubroutine && st.CurrentScope.Kind != ScopeMethod {
		return nil
	}

	// Analyze variable references in the subroutine body
	capturedSymbols := st.AnalyzeClosureCapture(st.CurrentScope)

	// Log captured variables if debugging
	if DebugScoping && len(capturedSymbols) > 0 {
		log.Printf("[DEBUG] Captured %d variables in %s scope", len(capturedSymbols), st.CurrentScope.Kind)
	}

	return nil
}

// Helper methods for CST processing

// nodeToPosition converts a CST node to an AST position
func (ab *AdvancedBinder) nodeToPosition(node *sitter.Node) ast.Position {
	return ast.Position{
		Line:   int(node.StartPosition().Row) + 1,
		Column: int(node.StartPosition().Column) + 1,
	}
}

// nodeText extracts text content from a CST node
func (ab *AdvancedBinder) nodeText(node *sitter.Node, content []byte) string {
	if node == nil {
		return ""
	}
	startByte := node.StartByte()
	endByte := node.EndByte()
	return string(content[startByte:endByte])
}

// getSymbolKindFromSigil determines symbol kind from Perl sigil
func (ab *AdvancedBinder) getSymbolKindFromSigil(name string) SymbolKind {
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

// stripSigil removes the sigil from a variable name
func (ab *AdvancedBinder) stripSigil(name string) string {
	if len(name) > 0 && (name[0] == '$' || name[0] == '@' || name[0] == '%' || name[0] == '*') {
		return name[1:]
	}
	return name
}
