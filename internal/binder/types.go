// ABOUTME: Core types and interfaces for symbol binding in Perl code analysis.
// ABOUTME: Defines Symbol, Scope, and SymbolTable structures following TypeScript-Go binder architecture.

//go:generate moq -out binder_mock.go . Binder SymbolResolver ScopeManager

package binder

import (
	"fmt"

	"tamarou.com/pvm/internal/ast"
)

// SymbolKind represents the different types of symbols in Perl
type SymbolKind int

const (
	// Variable symbols
	SymbolScalar SymbolKind = iota // $var
	SymbolArray                    // @array
	SymbolHash                     // %hash
	SymbolGlob                     // *glob

	// Callable symbols
	SymbolSubroutine // sub name
	SymbolMethod     // method name

	// Package and namespace symbols
	SymbolPackage // package declarations
	SymbolImport  // use/require imports

	// Type symbols (for typed Perl)
	SymbolType // type declarations
)

// SymbolFlags provides additional metadata about symbols
type SymbolFlags int

const (
	SymbolFlagNone             SymbolFlags = 0
	SymbolFlagLexical          SymbolFlags = 1 << iota // my, state variables
	SymbolFlagPackage                                  // our variables
	SymbolFlagExported                                 // symbols visible outside package
	SymbolFlagImported                                 // symbols from other modules
	SymbolFlagTypeAnnotated                            // has type annotation
	SymbolFlagMethod                                   // method symbol
	SymbolFlagClosure                                  // captured in closure
	SymbolFlagLocal                                    // local dynamic scoping
	SymbolFlagPackageQualified                         // package qualified ($Package::var)
	SymbolFlagAlias                                    // symbol alias or reference
	SymbolFlagUpvalue                                  // closure upvalue
	SymbolFlagDynamic                                  // dynamically created symbol
)

// Symbol represents a bound symbol with its metadata
type Symbol struct {
	Name        string       // Symbol name (without sigil for variables)
	Kind        SymbolKind   // Type of symbol
	Flags       SymbolFlags  // Additional metadata
	Declaration ast.Node     // AST node where symbol is declared
	Type        string       // Type annotation if present
	Scope       *Scope       // Scope where symbol is defined
	Package     string       // Package name where symbol is defined
	Position    ast.Position // Source position of declaration

	// Advanced scoping fields
	OriginalSymbol *Symbol   // For aliases, points to original symbol
	CapturedBy     []*Scope  // Scopes that capture this symbol (for closures)
	Upvalues       []*Symbol // Symbols captured from outer scopes
	QualifiedName  string    // Full package-qualified name
}

// String returns a string representation of the symbol for debugging and baseline testing
func (s *Symbol) String() string {
	result := s.Kind.String() + " " + s.Name

	if s.Type != "" {
		result += " :: " + s.Type
	}

	if s.Package != "" && s.Package != "main" {
		result += " @ " + s.Package
	}

	if s.Flags != SymbolFlagNone {
		result += " [" + s.Flags.String() + "]"
	}

	return fmt.Sprintf("%s at %d:%d", result, s.Position.Line, s.Position.Column)
}

// ScopeKind represents different types of scopes in Perl
type ScopeKind int

const (
	ScopeGlobal     ScopeKind = iota // Global/main scope
	ScopePackage                     // Package scope
	ScopeSubroutine                  // Subroutine scope
	ScopeMethod                      // Method scope
	ScopeBlock                       // Block scope (if/while/for/etc)
	ScopeEval                        // eval scope
)

// Scope represents a lexical scope with symbol resolution
type Scope struct {
	Kind     ScopeKind          // Type of scope
	Parent   *Scope             // Parent scope for scope chain
	Children []*Scope           // Child scopes
	Symbols  map[string]*Symbol // Symbols defined in this scope
	Package  string             // Current package name
	Node     ast.Node           // AST node that created this scope
	Position ast.Position       // Source position of scope start

	// Advanced scoping fields
	LocalSymbols    map[string]*Symbol // Local dynamic symbols (restored on scope exit)
	SavedValues     map[string]*Symbol // Saved symbol values for local restoration
	ImportedModules map[string]string  // Module imports in this scope
	CapturedSymbols []*Symbol          // Symbols captured by this scope (for closures)
}

// SymbolTable manages all symbols and scopes for a compilation unit
type SymbolTable struct {
	GlobalScope  *Scope               // Root scope
	Scopes       map[ast.Node]*Scope  // Map AST nodes to their scopes
	Symbols      map[string][]*Symbol // All symbols by name (for fast lookup)
	CurrentScope *Scope               // Current scope during binding
	Package      string               // Current package name

	// Advanced features
	ModuleSymbols   map[string]*SymbolTable // Symbol tables for imported modules
	PackageSymbols  map[string]*Scope       // Package-level symbol scopes
	ExportedSymbols map[string]*Symbol      // Symbols exported by this module
	DynamicSymbols  map[string]*Symbol      // Dynamically created symbols

	// Pool manager for efficient memory allocation
	PoolManager *SymbolPoolManager // Pool manager for symbols and scopes
}

// Binder interface defines the symbol binding operations
type Binder interface {
	// Bind performs symbol binding on an AST
	Bind(node ast.Node) (*SymbolTable, error)

	// BindAST performs symbol binding on a parsed AST
	BindAST(astTree *ast.AST) (*SymbolTable, error)
}

// SymbolResolver interface for resolving symbols in different contexts
type SymbolResolver interface {
	// ResolveSymbol finds a symbol by name in current scope chain
	ResolveSymbol(name string, kind SymbolKind) *Symbol

	// ResolveInScope finds a symbol within a specific scope
	ResolveInScope(scope *Scope, name string, kind SymbolKind) *Symbol

	// GetVisibleSymbols returns all symbols visible from current scope
	GetVisibleSymbols() []*Symbol
}

// ScopeManager interface for managing scope lifecycle
type ScopeManager interface {
	// EnterScope creates and enters a new scope
	EnterScope(kind ScopeKind, node ast.Node) *Scope

	// ExitScope returns to parent scope
	ExitScope() *Scope

	// GetCurrentScope returns the active scope
	GetCurrentScope() *Scope

	// FindScope finds scope containing the given AST node
	FindScope(node ast.Node) *Scope
}

// AdvancedSymbolResolver interface for complex symbol resolution
type AdvancedSymbolResolver interface {
	SymbolResolver

	// ResolvePackageSymbol resolves package-qualified symbols
	ResolvePackageSymbol(packageName, symbolName string, kind SymbolKind) *Symbol

	// ResolveImportedSymbol resolves symbols from imported modules
	ResolveImportedSymbol(moduleName, symbolName string, kind SymbolKind) *Symbol

	// CreateAlias creates a symbol alias
	CreateAlias(aliasName string, originalSymbol *Symbol) *Symbol

	// CaptureSymbol marks a symbol as captured by a closure
	CaptureSymbol(symbol *Symbol, capturingScope *Scope) error
}

// ModuleManager interface for handling module imports and exports
type ModuleManager interface {
	// ImportModule imports symbols from another module
	ImportModule(moduleName string, symbolTable *SymbolTable) error

	// ExportSymbol marks a symbol as exported
	ExportSymbol(symbol *Symbol) error

	// GetModuleSymbolTable returns the symbol table for a module
	GetModuleSymbolTable(moduleName string) *SymbolTable

	// RegisterModule registers a module's symbol table
	RegisterModule(moduleName string, symbolTable *SymbolTable) error
}

// BindingError represents errors during symbol binding
type BindingError struct {
	Message  string
	Position ast.Position
	Symbol   string
	Kind     string
}

func (e *BindingError) Error() string {
	return e.Message
}

// NewBindingError creates a new binding error
func NewBindingError(message, symbol, kind string, pos ast.Position) *BindingError {
	return &BindingError{
		Message:  message,
		Position: pos,
		Symbol:   symbol,
		Kind:     kind,
	}
}
