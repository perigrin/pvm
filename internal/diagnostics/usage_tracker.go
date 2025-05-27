// ABOUTME: Symbol usage tracking for detecting unused variables and patterns
// ABOUTME: Tracks symbol declarations, references, and usage patterns across AST

package diagnostics

import (
	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
)

// SymbolUsage tracks how a symbol is used throughout the code
type SymbolUsage struct {
	Symbol      *binder.Symbol
	Declared    bool
	References  []ast.Position
	Assignments []ast.Position
	LastUsed    ast.Position
}

// SymbolUsageTracker tracks symbol usage patterns
type SymbolUsageTracker struct {
	usageMap map[string]*SymbolUsage
}

// NewSymbolUsageTracker creates a new symbol usage tracker
func NewSymbolUsageTracker() *SymbolUsageTracker {
	return &SymbolUsageTracker{
		usageMap: make(map[string]*SymbolUsage),
	}
}

// TrackUsage analyzes the AST and tracks symbol usage patterns
func (tracker *SymbolUsageTracker) TrackUsage(astRoot ast.Node, symbolTable *binder.SymbolTable) {
	// Initialize usage tracking for all declared symbols
	allSymbols := symbolTable.GetVisibleSymbols()
	for _, symbol := range allSymbols {
		tracker.usageMap[symbol.Name] = &SymbolUsage{
			Symbol:      symbol,
			Declared:    true,
			References:  []ast.Position{},
			Assignments: []ast.Position{},
			LastUsed:    ast.Position{},
		}
	}

	// Walk AST to track usage
	tracker.walkASTForUsage(astRoot)
}

// walkASTForUsage walks the AST and records symbol usage
func (tracker *SymbolUsageTracker) walkASTForUsage(node ast.Node) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.VariableExpr:
		// Track variable reference
		tracker.trackReference(n.FullName(), n.Start())

	case *ast.BinaryExpr:
		// Track assignment (assignment is typically a binary operation in Perl)
		if n.Operator == "=" || n.Operator == "+=" || n.Operator == "-=" || n.Operator == "*=" || n.Operator == "/=" {
			if varExpr, ok := n.Left.(*ast.VariableExpr); ok {
				tracker.trackAssignment(varExpr.FullName(), n.Start())
			}
		}

	case *ast.VarDecl:
		// Track variable declarations
		for _, variable := range n.Variables {
			tracker.trackDeclaration(variable.FullName(), n.Start())
		}

	case *ast.CallExpr:
		// Track function/method calls
		if varExpr, ok := n.Function.(*ast.VariableExpr); ok {
			tracker.trackReference(varExpr.FullName(), n.Start())
		}
	}

	// Recursively process children
	for _, child := range node.Children() {
		tracker.walkASTForUsage(child)
	}
}

// trackReference records a reference to a symbol
func (tracker *SymbolUsageTracker) trackReference(symbolName string, pos ast.Position) {
	if usage, exists := tracker.usageMap[symbolName]; exists {
		usage.References = append(usage.References, pos)
		usage.LastUsed = pos
	}
}

// trackAssignment records an assignment to a symbol
func (tracker *SymbolUsageTracker) trackAssignment(symbolName string, pos ast.Position) {
	if usage, exists := tracker.usageMap[symbolName]; exists {
		usage.Assignments = append(usage.Assignments, pos)
		usage.LastUsed = pos
	}
}

// trackDeclaration records a declaration of a symbol
func (tracker *SymbolUsageTracker) trackDeclaration(symbolName string, pos ast.Position) {
	if usage, exists := tracker.usageMap[symbolName]; exists {
		usage.Declared = true
	} else {
		// Create new usage entry for undeclared symbols found in AST
		tracker.usageMap[symbolName] = &SymbolUsage{
			Symbol:      nil, // No symbol table entry
			Declared:    false,
			References:  []ast.Position{},
			Assignments: []ast.Position{},
			LastUsed:    ast.Position{},
		}
	}
}

// GetUnusedSymbols returns symbols that are declared but never used
func (tracker *SymbolUsageTracker) GetUnusedSymbols() []*binder.Symbol {
	var unused []*binder.Symbol

	for _, usage := range tracker.usageMap {
		if usage.Symbol != nil && usage.Declared && len(usage.References) == 0 && len(usage.Assignments) == 0 {
			// Symbol is declared but never referenced or assigned
			unused = append(unused, usage.Symbol)
		}
	}

	return unused
}

// GetUsageStats returns usage statistics for a symbol
func (tracker *SymbolUsageTracker) GetUsageStats(symbolName string) *SymbolUsage {
	return tracker.usageMap[symbolName]
}

// IsSymbolUsed checks if a symbol is used (referenced or assigned)
func (tracker *SymbolUsageTracker) IsSymbolUsed(symbolName string) bool {
	if usage, exists := tracker.usageMap[symbolName]; exists {
		return len(usage.References) > 0 || len(usage.Assignments) > 0
	}
	return false
}

// GetSymbolReferences returns all reference positions for a symbol
func (tracker *SymbolUsageTracker) GetSymbolReferences(symbolName string) []ast.Position {
	if usage, exists := tracker.usageMap[symbolName]; exists {
		return usage.References
	}
	return []ast.Position{}
}

// GetSymbolAssignments returns all assignment positions for a symbol
func (tracker *SymbolUsageTracker) GetSymbolAssignments(symbolName string) []ast.Position {
	if usage, exists := tracker.usageMap[symbolName]; exists {
		return usage.Assignments
	}
	return []ast.Position{}
}

// GetReadOnlySymbols returns symbols that are only read (never assigned after declaration)
func (tracker *SymbolUsageTracker) GetReadOnlySymbols() []*binder.Symbol {
	var readOnly []*binder.Symbol

	for _, usage := range tracker.usageMap {
		if usage.Symbol != nil && len(usage.References) > 0 && len(usage.Assignments) == 0 {
			readOnly = append(readOnly, usage.Symbol)
		}
	}

	return readOnly
}

// GetWriteOnlySymbols returns symbols that are only written (never read)
func (tracker *SymbolUsageTracker) GetWriteOnlySymbols() []*binder.Symbol {
	var writeOnly []*binder.Symbol

	for _, usage := range tracker.usageMap {
		if usage.Symbol != nil && len(usage.References) == 0 && len(usage.Assignments) > 0 {
			writeOnly = append(writeOnly, usage.Symbol)
		}
	}

	return writeOnly
}

// GetAllUsageInfo returns usage information for all tracked symbols
func (tracker *SymbolUsageTracker) GetAllUsageInfo() map[string]*SymbolUsage {
	return tracker.usageMap
}
