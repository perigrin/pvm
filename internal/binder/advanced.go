// ABOUTME: Advanced symbol binding features for Perl's complex scoping scenarios.
// ABOUTME: Handles closures, dynamic scoping, module imports, and package qualification.

package binder

import (
	"log"
	"strings"

	"tamarou.com/pvm/internal/ast"
)

// AdvancedBinder extends DefaultBinder with advanced Perl scoping features
type AdvancedBinder struct {
	*DefaultBinder
}

// NewAdvancedBinder creates a new advanced binder
func NewAdvancedBinder() *AdvancedBinder {
	return &AdvancedBinder{
		DefaultBinder: NewBinder(),
	}
}

// ResolvePackageSymbol resolves package-qualified symbols ($Package::var)
func (st *SymbolTable) ResolvePackageSymbol(packageName, symbolName string, kind SymbolKind) *Symbol {
	// First check if we have a package scope for this package
	if packageScope, exists := st.PackageSymbols[packageName]; exists {
		if symbol, found := packageScope.Symbols[symbolName]; found {
			if kind == SymbolKind(-1) || symbol.Kind == kind {
				return symbol
			}
		}
	}

	// Check in module symbol tables
	if moduleTable, exists := st.ModuleSymbols[packageName]; exists {
		return moduleTable.ResolveSymbol(symbolName, kind)
	}

	return nil
}

// ResolveImportedSymbol resolves symbols from imported modules
func (st *SymbolTable) ResolveImportedSymbol(moduleName, symbolName string, kind SymbolKind) *Symbol {
	// Check if module is imported in current scope or parent scopes
	current := st.CurrentScope
	for current != nil {
		if _, imported := current.ImportedModules[moduleName]; imported {
			if moduleTable, exists := st.ModuleSymbols[moduleName]; exists {
				return moduleTable.ResolveSymbol(symbolName, kind)
			}
		}
		current = current.Parent
	}

	return nil
}

// CreateAlias creates a symbol alias
func (st *SymbolTable) CreateAlias(aliasName string, originalSymbol *Symbol) *Symbol {
	alias := &Symbol{
		Name:           aliasName,
		Kind:           originalSymbol.Kind,
		Flags:          originalSymbol.Flags | SymbolFlagAlias,
		Declaration:    originalSymbol.Declaration,
		Type:           originalSymbol.Type,
		Scope:          st.CurrentScope,
		Package:        st.Package,
		Position:       originalSymbol.Position,
		OriginalSymbol: originalSymbol,
		QualifiedName:  originalSymbol.QualifiedName,
	}

	st.AddSymbol(alias)
	return alias
}

// CaptureSymbol marks a symbol as captured by a closure
func (st *SymbolTable) CaptureSymbol(symbol *Symbol, capturingScope *Scope) error {
	// Mark symbol as captured
	symbol.Flags |= SymbolFlagClosure
	symbol.CapturedBy = append(symbol.CapturedBy, capturingScope)

	// Add to capturing scope's captured symbols
	capturingScope.CapturedSymbols = append(capturingScope.CapturedSymbols, symbol)

	// If this is a closure scope, mark symbol as upvalue
	if capturingScope.Kind == ScopeSubroutine || capturingScope.Kind == ScopeMethod {
		symbol.Flags |= SymbolFlagUpvalue
	}

	return nil
}

// ImportModule imports symbols from another module
func (st *SymbolTable) ImportModule(moduleName string, moduleTable *SymbolTable) error {
	// Register the module
	st.ModuleSymbols[moduleName] = moduleTable

	// Add to current scope's imports
	if st.CurrentScope != nil {
		st.CurrentScope.ImportedModules[moduleName] = moduleName
	}

	return nil
}

// ExportSymbol marks a symbol as exported
func (st *SymbolTable) ExportSymbol(symbol *Symbol) error {
	symbol.Flags |= SymbolFlagExported
	st.ExportedSymbols[symbol.Name] = symbol
	return nil
}

// GetModuleSymbolTable returns the symbol table for a module
func (st *SymbolTable) GetModuleSymbolTable(moduleName string) *SymbolTable {
	return st.ModuleSymbols[moduleName]
}

// RegisterModule registers a module's symbol table
func (st *SymbolTable) RegisterModule(moduleName string, symbolTable *SymbolTable) error {
	st.ModuleSymbols[moduleName] = symbolTable
	return nil
}

// CreateLocalSymbol creates a local dynamic symbol (local $var)
func (st *SymbolTable) CreateLocalSymbol(symbol *Symbol) error {
	if st.CurrentScope == nil {
		return NewBindingError("no current scope for local symbol", symbol.Name, symbol.Kind.String(), symbol.Position)
	}

	// Save original value if it exists
	if existing := st.ResolveSymbol(symbol.Name, symbol.Kind); existing != nil {
		st.CurrentScope.SavedValues[symbol.Name] = existing
	}

	// Mark as local and add to current scope
	symbol.Flags |= SymbolFlagLocal
	st.CurrentScope.LocalSymbols[symbol.Name] = symbol

	return st.AddSymbol(symbol)
}

// RestoreLocalSymbols restores local symbols when exiting a scope
func (st *SymbolTable) RestoreLocalSymbols(scope *Scope) {
	// Restore saved values for local symbols
	for name, savedSymbol := range scope.SavedValues {
		if localSymbol, exists := scope.LocalSymbols[name]; exists {
			// Remove local symbol
			delete(scope.Symbols, name)
			delete(scope.LocalSymbols, name)

			// Restore saved symbol if it existed
			if savedSymbol != nil {
				scope.Symbols[name] = savedSymbol
			}

			// Remove from global symbol index
			if symbols, exists := st.Symbols[name]; exists {
				filtered := make([]*Symbol, 0, len(symbols))
				for _, sym := range symbols {
					if sym != localSymbol {
						filtered = append(filtered, sym)
					}
				}
				if len(filtered) > 0 {
					st.Symbols[name] = filtered
				} else {
					delete(st.Symbols, name)
				}
			}
		}
	}
}

// CreatePackageQualifiedSymbol creates a package-qualified symbol
func (st *SymbolTable) CreatePackageQualifiedSymbol(packageName, symbolName string, kind SymbolKind, flags SymbolFlags) *Symbol {
	qualifiedName := packageName + "::" + symbolName

	symbol := &Symbol{
		Name:          symbolName,
		Kind:          kind,
		Flags:         flags | SymbolFlagPackageQualified,
		Package:       packageName,
		QualifiedName: qualifiedName,
	}

	// Ensure package scope exists
	if _, exists := st.PackageSymbols[packageName]; !exists {
		packageScope := &Scope{
			Kind:            ScopePackage,
			Parent:          nil, // Package scopes are top-level
			Children:        []*Scope{},
			Symbols:         make(map[string]*Symbol),
			Package:         packageName,
			LocalSymbols:    make(map[string]*Symbol),
			SavedValues:     make(map[string]*Symbol),
			ImportedModules: make(map[string]string),
			CapturedSymbols: []*Symbol{},
		}
		st.PackageSymbols[packageName] = packageScope
	}

	// Add to package scope
	packageScope := st.PackageSymbols[packageName]
	packageScope.Symbols[symbolName] = symbol
	symbol.Scope = packageScope

	// Add to global symbol index
	st.Symbols[symbolName] = append(st.Symbols[symbolName], symbol)

	return symbol
}

// ResolveWithPackageQualification resolves symbols with package qualification
func (st *SymbolTable) ResolveWithPackageQualification(name string, kind SymbolKind) *Symbol {
	// Check if name contains package qualification
	if strings.Contains(name, "::") {
		parts := strings.SplitN(name, "::", 2)
		if len(parts) == 2 {
			packageName := parts[0]
			symbolName := parts[1]
			return st.ResolvePackageSymbol(packageName, symbolName, kind)
		}
	}

	// Regular resolution
	return st.ResolveSymbol(name, kind)
}

// AnalyzeClosureCapture analyzes closure variable capture
func (st *SymbolTable) AnalyzeClosureCapture(closureScope *Scope) []*Symbol {
	var capturedSymbols []*Symbol

	// For each symbol referenced in the closure scope
	for _, symbol := range closureScope.Symbols {
		// Check if symbol is defined in an outer scope
		if symbol.Scope != closureScope && symbol.Scope != st.GlobalScope {
			// This is potentially a captured variable
			if err := st.CaptureSymbol(symbol, closureScope); err == nil {
				capturedSymbols = append(capturedSymbols, symbol)
			}
		}
	}

	// Recursively analyze child scopes
	for _, child := range closureScope.Children {
		childCaptured := st.AnalyzeClosureCapture(child)
		capturedSymbols = append(capturedSymbols, childCaptured...)
	}

	return capturedSymbols
}

// CreateDynamicSymbol creates a dynamically created symbol
func (st *SymbolTable) CreateDynamicSymbol(name string, kind SymbolKind, flags SymbolFlags) *Symbol {
	symbol := &Symbol{
		Name:    name,
		Kind:    kind,
		Flags:   flags | SymbolFlagDynamic,
		Scope:   st.CurrentScope,
		Package: st.Package,
	}

	st.DynamicSymbols[name] = symbol
	st.AddSymbol(symbol)

	return symbol
}

// ProcessUseStatement processes use/require statements
func (ab *AdvancedBinder) ProcessUseStatement(useStmt *ast.UseStmt) error {
	moduleName := useStmt.Module

	// Create a placeholder module symbol table
	// In a real implementation, this would load the actual module
	moduleTable := NewSymbolTable()
	moduleTable.Package = moduleName

	// Import the module
	err := ab.symbolTable.ImportModule(moduleName, moduleTable)
	if err != nil {
		return err
	}

	// Create import symbol
	importSymbol := &Symbol{
		Name:     moduleName,
		Kind:     SymbolImport,
		Flags:    SymbolFlagImported,
		Package:  ab.symbolTable.Package,
		Position: useStmt.Start(),
	}

	return ab.symbolTable.AddSymbol(importSymbol)
}

// ProcessLocalVariable processes local variable declarations
func (ab *AdvancedBinder) ProcessLocalVariable(varName string, kind SymbolKind, pos ast.Position) error {
	symbol := &Symbol{
		Name:     ab.stripSigil(varName),
		Kind:     kind,
		Flags:    SymbolFlagLocal,
		Position: pos,
	}

	return ab.symbolTable.CreateLocalSymbol(symbol)
}

// ProcessPackageQualifiedAccess processes package qualified variable access
func (ab *AdvancedBinder) ProcessPackageQualifiedAccess(qualifiedName string, kind SymbolKind, pos ast.Position) *Symbol {
	return ab.symbolTable.ResolveWithPackageQualification(qualifiedName, kind)
}

// Enhanced ExitScope that handles local variable restoration
func (st *SymbolTable) ExitScopeAdvanced() *Scope {
	if st.CurrentScope != nil {
		// Debug logging
		if DebugScoping {
			parentID := -1
			if st.CurrentScope.Parent != nil {
				parentID = st.CurrentScope.Parent.ScopeID
			}
			log.Printf("[DEBUG] ExitScopeAdvanced: Leaving %s scope ID=%d, returning to parent ID=%d",
				st.CurrentScope.Kind.String(), st.CurrentScope.ScopeID, parentID)
		}

		// Restore local symbols before exiting
		st.RestoreLocalSymbols(st.CurrentScope)

		// Analyze closure capture if this is a subroutine/method scope
		if st.CurrentScope.Kind == ScopeSubroutine || st.CurrentScope.Kind == ScopeMethod {
			st.AnalyzeClosureCapture(st.CurrentScope)
		}

		// Return to parent scope
		if st.CurrentScope.Parent != nil {
			st.CurrentScope = st.CurrentScope.Parent
		}
	}
	return st.CurrentScope
}
