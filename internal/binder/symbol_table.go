// ABOUTME: SymbolTable implementation for managing symbols and scopes in Perl code analysis.
// ABOUTME: Provides scope chain management and symbol resolution following Perl's lexical scoping rules.

package binder

import (
	"fmt"
	"log"
	"strings"

	"tamarou.com/pvm/internal/ast"
)

// Debug flag for scope tracking
var DebugScoping = false

// Global scope counter for unique IDs
var scopeCounter int

// NewSymbolTable creates a new symbol table with global scope
func NewSymbolTable() *SymbolTable {
	return NewSymbolTableWithPool(DefaultSymbolPoolManager(), "main")
}

// NewSymbolTableWithPool creates a new symbol table with global scope using a pool manager
func NewSymbolTableWithPool(poolManager *SymbolPoolManager, packageName string) *SymbolTable {
	// Create symbol table using pool manager
	table := poolManager.NewSymbolTable(packageName)

	// Create global scope using pool manager
	globalScope := poolManager.NewScope(
		ScopeGlobal,
		nil,
		nil,
		ast.Position{Line: 1, Column: 1},
	)
	globalScope.Package = packageName

	// Set up symbol table
	table.GlobalScope = globalScope
	table.CurrentScope = globalScope
	table.PoolManager = poolManager

	return table
}

// EnterScope creates and enters a new scope
func (st *SymbolTable) EnterScope(kind ScopeKind, node ast.Node) *Scope {
	// Assign unique scope ID
	scopeCounter++
	currentScopeID := scopeCounter

	// Use pool manager to create scope if available
	var scope *Scope
	if st.PoolManager != nil {
		pos := ast.Position{Line: 1, Column: 1}
		if node != nil {
			pos = node.Start()
		}
		scope = st.PoolManager.NewScope(kind, st.CurrentScope, node, pos)
		scope.ScopeID = currentScopeID
	} else {
		// Fallback to direct allocation
		scope = &Scope{
			Kind:     kind,
			Parent:   st.CurrentScope,
			Children: []*Scope{},
			Symbols:  make(map[string]*Symbol),
			Node:     node,
			ScopeID:  currentScopeID,

			// Initialize advanced fields
			LocalSymbols:    make(map[string]*Symbol),
			SavedValues:     make(map[string]*Symbol),
			ImportedModules: make(map[string]string),
			CapturedSymbols: []*Symbol{},
		}

		// Add to parent's children
		if st.CurrentScope != nil {
			st.CurrentScope.Children = append(st.CurrentScope.Children, scope)
		}
	}

	scope.Package = st.Package

	// Debug logging
	if DebugScoping {
		parentID := -1
		if st.CurrentScope != nil {
			parentID = st.CurrentScope.ScopeID
		}
		log.Printf("[DEBUG] EnterScope: Created %s scope ID=%d (parent ID=%d)",
			kind.String(), currentScopeID, parentID)
	}

	// Map AST node to scope
	if node != nil {
		st.Scopes[node] = scope
	}

	// Update current scope
	st.CurrentScope = scope

	return scope
}

// ExitScope returns to parent scope
func (st *SymbolTable) ExitScope() *Scope {
	if st.CurrentScope != nil && st.CurrentScope.Parent != nil {
		// Debug logging
		if DebugScoping {
			log.Printf("[DEBUG] ExitScope: Leaving %s scope ID=%d, returning to parent ID=%d",
				st.CurrentScope.Kind.String(), st.CurrentScope.ScopeID, st.CurrentScope.Parent.ScopeID)
		}
		st.CurrentScope = st.CurrentScope.Parent
	}
	return st.CurrentScope
}

// GetCurrentScope returns the active scope
func (st *SymbolTable) GetCurrentScope() *Scope {
	return st.CurrentScope
}

// FindScope finds scope containing the given AST node
func (st *SymbolTable) FindScope(node ast.Node) *Scope {
	return st.Scopes[node]
}

// AddSymbol adds a symbol to the current scope
func (st *SymbolTable) AddSymbol(symbol *Symbol) error {
	if st.CurrentScope == nil {
		return NewBindingError("no current scope for symbol", symbol.Name, symbol.Kind.String(), symbol.Position)
	}

	// Debug logging
	if DebugScoping {
		log.Printf("[DEBUG] AddSymbol: Adding %s '%s' to %s scope ID=%d",
			symbol.Kind.String(), symbol.Name, st.CurrentScope.Kind.String(), st.CurrentScope.ScopeID)
	}

	// Check for redeclaration in same scope
	if existing, exists := st.CurrentScope.Symbols[symbol.Name]; exists {
		// Debug logging for conflict
		if DebugScoping {
			log.Printf("[DEBUG] AddSymbol: CONFLICT! Symbol '%s' already exists in %s scope ID=%d",
				symbol.Name, st.CurrentScope.Kind.String(), st.CurrentScope.ScopeID)
		}

		// Allow redeclaration in some cases (like our variables)
		if !st.canRedeclare(existing, symbol) {
			return NewBindingError(
				fmt.Sprintf("symbol '%s' already declared in this scope", symbol.Name),
				symbol.Name,
				symbol.Kind.String(),
				symbol.Position,
			)
		}
	}

	// Set symbol's scope
	symbol.Scope = st.CurrentScope
	symbol.Package = st.Package

	// Add to current scope
	st.CurrentScope.Symbols[symbol.Name] = symbol

	// Add to global symbol index
	st.Symbols[symbol.Name] = append(st.Symbols[symbol.Name], symbol)

	return nil
}

// AddSymbolToPackageScope adds a symbol to the global/package scope
func (st *SymbolTable) AddSymbolToPackageScope(symbol *Symbol) error {
	if st.GlobalScope == nil {
		return NewBindingError("no global scope available", symbol.Name, symbol.Kind.String(), symbol.Position)
	}

	// Check for redeclaration in package scope
	if existing, exists := st.GlobalScope.Symbols[symbol.Name]; exists {
		// Allow redeclaration in some cases (like our variables)
		if !st.canRedeclare(existing, symbol) {
			return NewBindingError(
				fmt.Sprintf("symbol '%s' already declared in package scope", symbol.Name),
				symbol.Name,
				symbol.Kind.String(),
				symbol.Position,
			)
		}
	}

	// Set symbol's scope to global scope
	symbol.Scope = st.GlobalScope
	symbol.Package = st.Package

	// Add to global scope
	st.GlobalScope.Symbols[symbol.Name] = symbol

	// Add to global symbol index
	st.Symbols[symbol.Name] = append(st.Symbols[symbol.Name], symbol)

	return nil
}

// canRedeclare determines if a symbol can be redeclare
func (st *SymbolTable) canRedeclare(existing, new *Symbol) bool {
	// Special handling for functions and methods - don't allow redeclaration if both are package symbols
	if (existing.Kind == SymbolSubroutine || existing.Kind == SymbolMethod) &&
		(new.Kind == SymbolSubroutine || new.Kind == SymbolMethod) {
		if existing.Flags&SymbolFlagPackage != 0 && new.Flags&SymbolFlagPackage != 0 {
			// Package functions/methods cannot be redeclared
			return false
		}
	}

	// Allow our variables to be redeclared
	if existing.Flags&SymbolFlagPackage != 0 && new.Flags&SymbolFlagPackage != 0 {
		return true
	}

	// Allow different kinds in different scopes
	if existing.Kind != new.Kind {
		return true
	}

	return false
}

// ResolveSymbol finds a symbol by name in current scope chain
func (st *SymbolTable) ResolveSymbol(name string, kind SymbolKind) *Symbol {
	return st.ResolveInScope(st.CurrentScope, name, kind)
}

// ResolveInScope finds a symbol within a specific scope and its parents
func (st *SymbolTable) ResolveInScope(scope *Scope, name string, kind SymbolKind) *Symbol {
	current := scope

	for current != nil {
		if symbol, exists := current.Symbols[name]; exists {
			// Check if kind matches (or if we're looking for any kind)
			if kind == SymbolKind(-1) || symbol.Kind == kind {
				return symbol
			}
		}
		current = current.Parent
	}

	return nil
}

// GetVisibleSymbols returns all symbols visible from current scope
func (st *SymbolTable) GetVisibleSymbols() []*Symbol {
	var symbols []*Symbol
	current := st.CurrentScope

	for current != nil {
		for _, symbol := range current.Symbols {
			symbols = append(symbols, symbol)
		}
		current = current.Parent
	}

	return symbols
}

// GetSymbolsByName returns all symbols with the given name
func (st *SymbolTable) GetSymbolsByName(name string) []*Symbol {
	return st.Symbols[name]
}

// GetScopeDepth returns the depth of current scope (0 = global)
func (st *SymbolTable) GetScopeDepth() int {
	depth := 0
	current := st.CurrentScope

	for current != nil && current.Parent != nil {
		depth++
		current = current.Parent
	}

	return depth
}

// SetPackage updates the current package name
func (st *SymbolTable) SetPackage(packageName string) {
	st.Package = packageName
	if st.CurrentScope != nil {
		st.CurrentScope.Package = packageName
	}
}

// String methods for debugging
func (kind SymbolKind) String() string {
	switch kind {
	case SymbolScalar:
		return "scalar"
	case SymbolArray:
		return "array"
	case SymbolHash:
		return "hash"
	case SymbolGlob:
		return "glob"
	case SymbolSubroutine:
		return "subroutine"
	case SymbolMethod:
		return "method"
	case SymbolPackage:
		return "package"
	case SymbolImport:
		return "import"
	case SymbolType:
		return "type"
	default:
		return "unknown"
	}
}

func (flags SymbolFlags) String() string {
	var parts []string

	if flags&SymbolFlagLexical != 0 {
		parts = append(parts, "lexical")
	}
	if flags&SymbolFlagPackage != 0 {
		parts = append(parts, "package")
	}
	if flags&SymbolFlagExported != 0 {
		parts = append(parts, "exported")
	}
	if flags&SymbolFlagImported != 0 {
		parts = append(parts, "imported")
	}
	if flags&SymbolFlagTypeAnnotated != 0 {
		parts = append(parts, "typed")
	}
	if flags&SymbolFlagMethod != 0 {
		parts = append(parts, "method")
	}
	if flags&SymbolFlagClosure != 0 {
		parts = append(parts, "closure")
	}
	if flags&SymbolFlagLocal != 0 {
		parts = append(parts, "local")
	}
	if flags&SymbolFlagPackageQualified != 0 {
		parts = append(parts, "qualified")
	}
	if flags&SymbolFlagAlias != 0 {
		parts = append(parts, "alias")
	}
	if flags&SymbolFlagUpvalue != 0 {
		parts = append(parts, "upvalue")
	}
	if flags&SymbolFlagDynamic != 0 {
		parts = append(parts, "dynamic")
	}

	if len(parts) == 0 {
		return "none"
	}

	return strings.Join(parts, "|")
}
