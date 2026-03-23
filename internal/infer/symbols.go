// ABOUTME: Symbol table for the PSC type inference engine.
// ABOUTME: Defines Symbol, Scope, SymbolTable, and the CollectDeclarations tree walk (pass 1).

package infer

import (
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/types"
)

// SymbolKind distinguishes the syntactic role of a declared name.
type SymbolKind int

const (
	SymVariable   SymbolKind = iota // Lexical variable: $x, @arr, %hash
	SymSubroutine                   // Named subroutine: sub foo { ... }
	SymPackage                      // Package namespace: package Foo;
	SymMethod                       // Method: sub in an OO context (future)
)

// Symbol records a single declared name together with its inferred type,
// syntactic role, and source location.
type Symbol struct {
	Name       string
	Type       types.Type
	ReturnType types.Type   // Inferred return type (subroutines only, zero = unknown)
	ParamTypes []types.Type // Inferred parameter types (subroutines only, positional)
	ParamNames []string     // Parameter variable names (subroutines only, positional)
	ClassType  string       // Class name for object variables (e.g. "Foo" for Foo->new())
	Kind       SymbolKind
	StartByte  uint32
	EndByte    uint32
}

// Scope is a single lexical scope that may refer back to its enclosing parent.
type Scope struct {
	Name    string
	parent  *Scope
	symbols map[string]Symbol
}

// Parent returns the enclosing scope, or nil for the root scope.
func (s *Scope) Parent() *Scope {
	return s.parent
}

// SymbolTable tracks the full lexical scope chain.  The current scope is the
// innermost active scope; its ancestors are accessible via the Parent chain.
type SymbolTable struct {
	root    *Scope
	current *Scope
}

// NewSymbolTable creates a fresh symbol table rooted at the implicit "main"
// package scope.
func NewSymbolTable() *SymbolTable {
	root := &Scope{
		Name:    "main",
		parent:  nil,
		symbols: make(map[string]Symbol),
	}
	return &SymbolTable{root: root, current: root}
}

// Define adds sym to the current (innermost) scope, overwriting any
// same-named symbol already in this scope.
func (st *SymbolTable) Define(sym Symbol) {
	st.current.symbols[sym.Name] = sym
}

// Lookup searches the current scope and each ancestor for name, returning the
// first match found.  The inner-most definition wins (shadowing).
func (st *SymbolTable) Lookup(name string) (Symbol, bool) {
	for s := st.current; s != nil; s = s.parent {
		if sym, ok := s.symbols[name]; ok {
			return sym, true
		}
	}
	return Symbol{}, false
}

// EnterScope pushes a new child scope onto the scope stack and makes it
// current.
func (st *SymbolTable) EnterScope(name string) {
	child := &Scope{
		Name:    name,
		parent:  st.current,
		symbols: make(map[string]Symbol),
	}
	st.current = child
}

// ExitScope pops the current scope back to its parent.  Calling ExitScope on
// the root scope is a no-op to avoid panics from unbalanced enter/exit pairs.
func (st *SymbolTable) ExitScope() {
	if st.current.parent != nil {
		st.current = st.current.parent
	}
}

// UpdateType walks the scope chain from the current scope to root, looking
// for the first scope that contains name.  If found, it updates the symbol's
// Type to typ and returns true.  If no scope contains name, it returns false.
func (st *SymbolTable) UpdateType(name string, typ types.Type) bool {
	for s := st.current; s != nil; s = s.parent {
		if sym, ok := s.symbols[name]; ok {
			sym.Type = typ
			s.symbols[name] = sym
			return true
		}
	}
	return false
}

// UpdateReturnType walks the scope chain from the current scope to root,
// looking for the first scope that contains name. If found, it updates the
// symbol's ReturnType to typ and returns true. Used to store inferred
// subroutine return types after analyzing the sub body.
func (st *SymbolTable) UpdateReturnType(name string, typ types.Type) bool {
	for s := st.current; s != nil; s = s.parent {
		if sym, ok := s.symbols[name]; ok {
			sym.ReturnType = typ
			s.symbols[name] = sym
			return true
		}
	}
	return false
}

// UpdateParamTypes walks the scope chain from the current scope to root,
// looking for the first scope that contains name. If found, it updates the
// symbol's ParamTypes and ParamNames and returns true. Used to store inferred
// subroutine parameter types after analyzing the sub body.
func (st *SymbolTable) UpdateParamTypes(name string, paramNames []string, paramTypes []types.Type) bool {
	for s := st.current; s != nil; s = s.parent {
		if sym, ok := s.symbols[name]; ok {
			sym.ParamNames = paramNames
			sym.ParamTypes = paramTypes
			s.symbols[name] = sym
			return true
		}
	}
	return false
}

// UpdateClassType walks the scope chain from the current scope to root,
// looking for the first scope that contains name. If found, it updates the
// symbol's ClassType to className and returns true. Used to store the class
// name inferred from constructor calls like Foo->new().
func (st *SymbolTable) UpdateClassType(name string, className string) bool {
	for s := st.current; s != nil; s = s.parent {
		if sym, ok := s.symbols[name]; ok {
			sym.ClassType = className
			s.symbols[name] = sym
			return true
		}
	}
	return false
}

// CurrentPackage returns the name of the nearest ancestor scope (inclusive)
// that was created by a package declaration, or "main" if none is found.
func (st *SymbolTable) CurrentPackage() string {
	for s := st.current; s != nil; s = s.parent {
		if s.parent == nil {
			// The root is always the implicit package scope
			return s.Name
		}
	}
	return "main"
}

// CollectDeclarations performs a top-down walk of the tree-sitter CST and
// populates a SymbolTable with every variable and subroutine declaration it
// finds.  This is pass 1 of the type inference pipeline; type information is
// derived solely from sigils and syntax, not from flow analysis.
func CollectDeclarations(tree *parser.Tree, source []byte) *SymbolTable {
	st := NewSymbolTable()
	if tree == nil {
		return st
	}
	root := tree.RootNode()
	if root == nil {
		return st
	}
	walkDeclarations(root, source, st)
	return st
}

// walkDeclarations recursively visits every node in the CST, dispatching on
// node kind to handle declarations and scope boundaries.
func walkDeclarations(node *parser.Node, source []byte, st *SymbolTable) {
	if node == nil {
		return
	}

	kind := node.Kind()

	switch kind {
	case "package_statement":
		// package Foo;  — the package name is in the child with kind "package"
		collectPackageStatement(node, source, st)

	case "variable_declaration":
		// my $x / my @arr / my %hash — the typed child is scalar/array/hash
		collectVariableDeclaration(node, source, st)

	case "subroutine_declaration_statement":
		// sub foo { ... }
		collectSubroutineDeclaration(node, source, st)
		// subroutine body is walked inside collectSubroutineDeclaration
		return

	case "block":
		// Generic block scope (e.g. if-block, anonymous block)
		st.EnterScope("block")
		for i := 0; i < node.ChildCount(); i++ {
			walkDeclarations(node.Child(i), source, st)
		}
		st.ExitScope()
		return
	}

	// Default: recurse into all children
	for i := 0; i < node.ChildCount(); i++ {
		walkDeclarations(node.Child(i), source, st)
	}
}

// collectPackageStatement extracts the package name and updates the root
// scope name so that CurrentPackage() returns the right value.  A new scope
// named after the package is entered.
func collectPackageStatement(node *parser.Node, source []byte, st *SymbolTable) {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "package" {
			pkgName := child.Text(source)
			// Update the root scope name so CurrentPackage works correctly
			st.root.Name = pkgName
			st.Define(Symbol{
				Name:      pkgName,
				Type:      types.Unknown,
				Kind:      SymPackage,
				StartByte: node.StartByte(),
				EndByte:   node.EndByte(),
			})
			return
		}
	}
}

// collectVariableDeclaration extracts the variable name and derives its type
// from the sigil: $ → Scalar, @ → Array, % → Hash.
func collectVariableDeclaration(node *parser.Node, source []byte, st *SymbolTable) {
	// Check whether this declaration is part of an assignment (my $x = ...).
	// If not, the variable is uninitialized and scalars get type Undef.
	hasInitializer := false
	if parent := node.Parent(); parent != nil && parent.Kind() == "assignment_expression" {
		hasInitializer = true
	}

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "scalar":
			name := sigildName("$", child, source)
			scalarType := types.Scalar
			if !hasInitializer {
				scalarType = types.Undef
			}
			st.Define(Symbol{
				Name:      name,
				Type:      scalarType,
				Kind:      SymVariable,
				StartByte: node.StartByte(),
				EndByte:   node.EndByte(),
			})
		case "array":
			name := sigildName("@", child, source)
			st.Define(Symbol{
				Name:      name,
				Type:      types.Array,
				Kind:      SymVariable,
				StartByte: node.StartByte(),
				EndByte:   node.EndByte(),
			})
		case "hash":
			name := sigildName("%", child, source)
			st.Define(Symbol{
				Name:      name,
				Type:      types.Hash,
				Kind:      SymVariable,
				StartByte: node.StartByte(),
				EndByte:   node.EndByte(),
			})
		}
	}
}

// sigildName finds the varname child inside a scalar/array/hash node and
// returns the sigil-prefixed name (e.g. "$x", "@arr", "%hash").
func sigildName(sigil string, node *parser.Node, source []byte) string {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Kind() == "varname" {
			return sigil + child.Text(source)
		}
	}
	// Fallback: use the full text of the node
	return sigil + node.Text(source)
}

// collectSubroutineDeclaration extracts the sub name, defines it in the
// current scope, then walks the sub body in a new child scope.
func collectSubroutineDeclaration(node *parser.Node, source []byte, st *SymbolTable) {
	var subName string
	var bodyNode *parser.Node

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "bareword":
			subName = child.Text(source)
		case "block":
			bodyNode = child
		}
	}

	if subName != "" {
		st.Define(Symbol{
			Name:      subName,
			Type:      types.Code,
			Kind:      SymSubroutine,
			StartByte: node.StartByte(),
			EndByte:   node.EndByte(),
		})
	}

	// Walk the subroutine body in a dedicated scope
	if bodyNode != nil {
		st.EnterScope(subName)
		for i := 0; i < bodyNode.ChildCount(); i++ {
			walkDeclarations(bodyNode.Child(i), source, st)
		}
		st.ExitScope()
	}
}
