// ABOUTME: Symbol table for the PSC type inference engine.
// ABOUTME: Defines Symbol, Scope and SymbolTable, which pass 1 fills and pass 2 narrows.

package infer

import (
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

	// ElemType is the type of an ELEMENT of an aggregate — the join of the
	// values it was seen to hold. Reading `$n[0]` yields this rather than the
	// sigil default, which is Scalar and says almost nothing.
	//
	// It is the join because an index selects one of the elements and PSC
	// does not evaluate indices: `my @m = (1, "s")` holds an Int and a Str,
	// and a read is whichever the index picks. Zero means no element type is
	// known, and a read then falls back to Scalar rather than guessing.
	ElemType types.Type

	// KeyTypes is, for a hash filled from key/value literals, the type of
	// the value under each constant key: `{count => 1, values => [2]}`
	// holds an Int under count. A read under one of these keys yields its
	// type rather than ElemType, the join of every value. Nil when no
	// literal filled the hash.
	KeyTypes  map[string]types.Type
	Kind      SymbolKind
	StartByte uint32
	EndByte   uint32
}

// Scope is a single lexical scope that may refer back to its enclosing parent.
// children holds every scope created by EnterScope while this scope was current,
// enabling AllSymbols to traverse the full scope tree after analysis is complete.
type Scope struct {
	Name     string
	parent   *Scope
	children []*Scope
	symbols  map[string]Symbol
}

// SymbolTable tracks the full lexical scope chain.  The current scope is the
// innermost active scope; its ancestors are reached through each scope's parent.
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
// current.  The child is appended to the parent's children slice so that
// AllSymbols can traverse the full scope tree later.
func (st *SymbolTable) EnterScope(name string) {
	child := &Scope{
		Name:    name,
		parent:  st.current,
		symbols: make(map[string]Symbol),
	}
	st.current.children = append(st.current.children, child)
	st.current = child
}

// ExitScope pops the current scope back to its parent.  Calling ExitScope on
// the root scope is a no-op to avoid panics from unbalanced enter/exit pairs.
func (st *SymbolTable) ExitScope() {
	if st.current.parent != nil {
		st.current = st.current.parent
	}
}

// AllSymbols returns every symbol defined across all scopes in the symbol
// table by recursively walking the scope tree from the root.  The order of
// symbols within a scope is not guaranteed (map iteration), and scopes are
// visited depth-first (parent before children).
func (st *SymbolTable) AllSymbols() []Symbol {
	var result []Symbol
	var walk func(s *Scope)
	walk = func(s *Scope) {
		for _, sym := range s.symbols {
			result = append(result, sym)
		}
		for _, child := range s.children {
			walk(child)
		}
	}
	walk(st.root)
	return result
}

// LookupAll searches every scope in the symbol table tree for name, returning
// the first match found via a depth-first walk from the root. Unlike Lookup,
// which only searches the current scope chain upward, LookupAll also finds
// symbols defined in child scopes that are no longer current — for example,
// symbols inside subroutine bodies after analysis has completed and current
// has returned to root.
func (st *SymbolTable) LookupAll(name string) (Symbol, bool) {
	var found Symbol
	var ok bool
	var walk func(s *Scope)
	walk = func(s *Scope) {
		if ok {
			return
		}
		if sym, exists := s.symbols[name]; exists {
			found = sym
			ok = true
			return
		}
		for _, child := range s.children {
			walk(child)
		}
	}
	walk(st.root)
	return found, ok
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

// UpdateElemType records the element type of an aggregate, joining it with
// whatever was already known.
//
// Joining rather than replacing is what makes a container filled from several
// places come out right: each assignment contributes an arm, and the element
// type is the merge of all of them.
func (st *SymbolTable) UpdateElemType(name string, typ types.Type) bool {
	for s := st.current; s != nil; s = s.parent {
		if sym, ok := s.symbols[name]; ok {
			sym.ElemType = types.Join(sym.ElemType, typ)
			s.symbols[name] = sym
			return true
		}
	}
	return false
}

// UpdateKeyTypes joins keys, a key/value literal's value type under each
// constant key, into the KeyTypes of the first scope's name, as
// UpdateElemType joins an element type. A key only one literal gave keeps
// that literal's type.
func (st *SymbolTable) UpdateKeyTypes(name string, keys map[string]types.Type) bool {
	for s := st.current; s != nil; s = s.parent {
		if sym, ok := s.symbols[name]; ok {
			joined := make(map[string]types.Type, len(sym.KeyTypes)+len(keys))
			for k, t := range sym.KeyTypes {
				joined[k] = t
			}
			for k, t := range keys {
				joined[k] = types.Join(joined[k], t)
			}
			sym.KeyTypes = joined
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
