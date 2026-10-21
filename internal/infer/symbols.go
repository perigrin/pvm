// ABOUTME: Symbol table for the PSC type inference engine.
// ABOUTME: Defines Symbol, Scope and SymbolTable, which pass 1 fills and pass 2 narrows.

package infer

import (
	"strings"

	"tamarou.com/pvm/internal/lower"
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

	// PossibleUndef carries a declared call's own possible-undef mark
	// (markPossibleUndef, calls.go) onto the variable its result was
	// assigned to, so a later read of the variable -- not only the call
	// expression itself -- still answers WarnPossibleUndef's default
	// silence (perigrin, 2026-10-20): `my $class = blessed($self); ...
	// bless({...}, $class)` silences the same way `bless({...},
	// blessed($self))` does. Set or cleared on every assignment
	// (narrow, ops.go), so a variable reassigned from an ordinary value
	// loses a stale mark from an earlier one.
	PossibleUndef bool
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
	// blockScopes is the scope pass 1 made for each block, keyed by the
	// block's id, which pass 2 re-enters as it walks the block.
	blockScopes map[lower.ID]*Scope
	// frames is a stack of open PushFrame spans. update, Delete and
	// ClearElementShadows each record, into every frame open when they
	// touch a name for the first time in that frame, how to undo the
	// touch -- so PopFrame can hand back exactly the names a span
	// mutated, without a Lookup sweep over every name it merely might
	// have touched (the root cause of I10's cubic nested ternary: a
	// preBranchTypes/joinBranches pair that scanned and re-looked-up
	// every name in a chain's remaining subtree, at a scope depth that
	// grew with the chain itself).
	frames []map[string]frameEntry
}

// frameEntry is one name a PushFrame span has touched: before is its
// value at the moment of the first touch, for a caller (walkJoined,
// flow.go) that wants to join the old and new type together, and undo
// puts that value back where it came from, for a caller (ternary, ops.go,
// and branch, flow.go) that wants to fully revert the touch instead.
type frameEntry struct {
	before Symbol
	undo   func()
}

// PushFrame starts recording every name an update, Delete or
// ClearElementShadows call touches from here on, until the matching
// PopFrame.
func (st *SymbolTable) PushFrame() {
	st.frames = append(st.frames, map[string]frameEntry{})
}

// PopFrame closes the span the matching PushFrame started and returns
// every name touched during it, each mapped to its value from just
// before that first touch and a func that puts it back.
func (st *SymbolTable) PopFrame() map[string]frameEntry {
	f := st.frames[len(st.frames)-1]
	st.frames = st.frames[:len(st.frames)-1]
	return f
}

// recordFrame notes, in every frame currently open, that name is about to
// be changed from before by undo -- unless some earlier touch in that
// frame already recorded it, since a frame's value is what name held when
// the frame started, not after whatever touched it most recently inside
// it.
func (st *SymbolTable) recordFrame(name string, before Symbol, undo func()) {
	for _, f := range st.frames {
		if _, seen := f[name]; !seen {
			f[name] = frameEntry{before: before, undo: undo}
		}
	}
}

// NewSymbolTable creates a fresh symbol table rooted at the implicit "main"
// package scope.
func NewSymbolTable() *SymbolTable {
	root := &Scope{
		Name:    "main",
		parent:  nil,
		symbols: make(map[string]Symbol),
	}
	return &SymbolTable{root: root, current: root, blockScopes: map[lower.ID]*Scope{}}
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

// Reenter makes s, a scope pass 1 made, current again, as a child of the
// scope current now: a lookup in s then reaches what pass 2 defined
// around it, a guard's shadows or a loop variable, before the scopes that
// enclose it. ExitScope leaves it.
func (st *SymbolTable) Reenter(s *Scope) {
	s.parent = st.current
	st.current = s
}

// scopeOf is the innermost scope on the current chain that holds name,
// or nil.
func (st *SymbolTable) scopeOf(name string) *Scope {
	for s := st.current; s != nil; s = s.parent {
		if _, ok := s.symbols[name]; ok {
			return s
		}
	}
	return nil
}

// inScope reports whether s is on the current chain: the current scope or
// one that encloses it.
func (st *SymbolTable) inScope(s *Scope) bool {
	for c := st.current; c != nil; c = c.parent {
		if c == s {
			return true
		}
	}
	return false
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

// update walks the scope chain from the current scope to root, looking for
// the first scope that contains name. If found, it runs f on the symbol,
// writes it back, and returns true; every Update* method below is this walk
// with f setting the one field it owns. If no scope contains name, it
// returns false and f never runs.
func (st *SymbolTable) update(name string, f func(*Symbol)) bool {
	for s := st.current; s != nil; s = s.parent {
		if sym, ok := s.symbols[name]; ok {
			scope, before := s, sym
			st.recordFrame(name, before, func() { scope.symbols[name] = before })
			f(&sym)
			s.symbols[name] = sym
			return true
		}
	}
	return false
}

// UpdateType walks the scope chain from the current scope to root, looking
// for the first scope that contains name.  If found, it updates the symbol's
// Type to typ and returns true.  If no scope contains name, it returns false.
func (st *SymbolTable) UpdateType(name string, typ types.Type) bool {
	return st.update(name, func(sym *Symbol) { sym.Type = typ })
}

// UpdateElemType records the element type of an aggregate, joining it with
// whatever was already known.
//
// Joining rather than replacing is what makes a container filled from several
// places come out right: each assignment contributes an arm, and the element
// type is the merge of all of them.
func (st *SymbolTable) UpdateElemType(name string, typ types.Type) bool {
	return st.update(name, func(sym *Symbol) { sym.ElemType = types.Join(sym.ElemType, typ) })
}

// UpdateKeyTypes joins keys, a key/value literal's value type under each
// constant key, into the KeyTypes of the first scope's name, as
// UpdateElemType joins an element type. A key only one literal gave keeps
// that literal's type.
func (st *SymbolTable) UpdateKeyTypes(name string, keys map[string]types.Type) bool {
	return st.update(name, func(sym *Symbol) {
		joined := make(map[string]types.Type, len(sym.KeyTypes)+len(keys))
		for k, t := range sym.KeyTypes {
			joined[k] = t
		}
		for k, t := range keys {
			joined[k] = types.Join(joined[k], t)
		}
		sym.KeyTypes = joined
	})
}

// UpdateReturnType walks the scope chain from the current scope to root,
// looking for the first scope that contains name. If found, it updates the
// symbol's ReturnType to typ and returns true. Used to store inferred
// subroutine return types after analyzing the sub body.
func (st *SymbolTable) UpdateReturnType(name string, typ types.Type) bool {
	return st.update(name, func(sym *Symbol) { sym.ReturnType = typ })
}

// UpdateParamTypes walks the scope chain from the current scope to root,
// looking for the first scope that contains name. If found, it updates the
// symbol's ParamTypes and ParamNames and returns true. Used to store inferred
// subroutine parameter types after analyzing the sub body.
func (st *SymbolTable) UpdateParamTypes(name string, paramNames []string, paramTypes []types.Type) bool {
	return st.update(name, func(sym *Symbol) {
		sym.ParamNames = paramNames
		sym.ParamTypes = paramTypes
	})
}

// UpdateClassType walks the scope chain from the current scope to root,
// looking for the first scope that contains name. If found, it updates the
// symbol's ClassType to className and returns true. Used to store the class
// name inferred from constructor calls like Foo->new().
func (st *SymbolTable) UpdateClassType(name string, className string) bool {
	return st.update(name, func(sym *Symbol) { sym.ClassType = className })
}

// UpdatePossibleUndef walks the scope chain from the current scope to root,
// looking for the first scope that contains name. If found, it updates the
// symbol's PossibleUndef to v and returns true.
func (st *SymbolTable) UpdatePossibleUndef(name string, v bool) bool {
	return st.update(name, func(sym *Symbol) { sym.PossibleUndef = v })
}

// Delete removes name from the first scope on the current chain that holds
// it. A name no scope on the chain holds is a no-op.
func (st *SymbolTable) Delete(name string) {
	for s := st.current; s != nil; s = s.parent {
		if sym, ok := s.symbols[name]; ok {
			scope, before := s, sym
			st.recordFrame(name, before, func() { scope.symbols[name] = before })
			delete(s.symbols, name)
			return
		}
	}
}

// ClearElementShadows drops every $name[N] shadow guardShadows defined for
// the array named name (flow.go's guardVarName), from every scope on the
// current chain: a shift, a splice, an element write or a whole-array
// write moves or replaces what sits at each index, so a guard's shadow
// from before it no longer answers for what a read finds there after it.
func (st *SymbolTable) ClearElementShadows(name string) {
	prefix := "$" + name + "["
	for s := st.current; s != nil; s = s.parent {
		for k, sym := range s.symbols {
			if strings.HasPrefix(k, prefix) {
				scope, key, before := s, k, sym
				st.recordFrame(key, before, func() { scope.symbols[key] = before })
				delete(s.symbols, k)
			}
		}
	}
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
