// ABOUTME: Symbol table for the PSC type inference engine.
// ABOUTME: Defines Symbol, Scope and SymbolTable, which pass 1 fills and pass 2 narrows.

package infer

import (
	"fmt"
	"maps"
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
	// open is whether the scope is on the current chain: entered or
	// re-entered, and not yet exited; depth is its distance from the root
	// there.
	open  bool
	depth int
	// unreached is whether a statement in the scope always exits, so the
	// code past it never runs: Unreached opened a frame there, which
	// ExitScope undoes.
	unreached bool
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
	// ClearElementShadows each record, in the innermost open frame, the
	// first touch of each name -- so PopFrame can hand back exactly the
	// names a span mutated, without a Lookup sweep over every name it
	// merely might have touched (the root cause of I10's cubic nested
	// ternary). PopFrame passes what it held on to the frame around it,
	// so a touch costs the same however many frames are open (V5).
	frames []frame
	// holders is, for each name, the open scopes that hold it, outermost
	// first, so the innermost is found without walking the chain: a chain
	// of guards nests a scope at every level, and a walk per lookup per
	// level made a long one quadratic (V5).
	holders map[string][]*Scope
}

// put writes sym as n's value in the scope n names, adding the scope to
// the name's holders when it did not hold the name. It records no touch:
// with a frame's before value, it is how a caller (UndoFrame, and branch,
// flow.go) fully reverts a touch rather than joining the old and new
// values the way walkJoined and loop (flow.go) do.
func (st *SymbolTable) put(n scopedName, sym Symbol) {
	if _, held := n.scope.symbols[n.name]; !held && n.scope.open {
		h := st.holders[n.name]
		i := len(h)
		for i > 0 && h[i-1].depth > n.scope.depth {
			i--
		}
		st.holders[n.name] = append(h[:i], append([]*Scope{n.scope}, h[i:]...)...)
	}
	n.scope.symbols[n.name] = sym
}

// drop deletes n from the scope n names, and the scope from the name's
// holders.
func (st *SymbolTable) drop(n scopedName) {
	delete(n.scope.symbols, n.name)
	h := st.holders[n.name]
	for i := len(h) - 1; i >= 0; i-- {
		if h[i] == n.scope {
			st.holders[n.name] = append(h[:i], h[i+1:]...)
			return
		}
	}
}

// frame is one open PushFrame span's touches: each name's value from
// just before its first touch in the span, keyed by the scope that held
// it. Element guard shadows ("$a[0]", flow.go's guardVarName) are kept
// apart from every other name: callers drop them or put them back but
// never join them, and a write that clears an array's shadows touches
// one per guard around it, which a caller joining the rest should not
// have to step over at every level of a chain (V5).
type frame struct {
	vars, elems map[scopedName]Symbol
}

// PushFrame starts recording every name an update, Delete or
// ClearElementShadows call touches from here on, until the matching
// PopFrame.
func (st *SymbolTable) PushFrame() {
	st.frames = append(st.frames, frame{map[scopedName]Symbol{}, map[scopedName]Symbol{}})
}

// Touched is every name but an element guard shadow the innermost open
// span has touched so far, keyed by the scope that held it, each mapped
// to its value from just before that first touch. A caller reads it
// before PopFrame, which may hand the map on to the span around it.
func (st *SymbolTable) Touched() map[scopedName]Symbol {
	return st.frames[len(st.frames)-1].vars
}

// TouchedElems is every element guard shadow the innermost open span has
// touched so far, as Touched says.
func (st *SymbolTable) TouchedElems() map[scopedName]Symbol {
	return st.frames[len(st.frames)-1].elems
}

// PopFrame closes the span the matching PushFrame started. Each name it
// touched is also a touch of the span around it, as merged says.
func (st *SymbolTable) PopFrame() {
	f := st.frames[len(st.frames)-1]
	st.frames = st.frames[:len(st.frames)-1]
	if len(st.frames) == 0 {
		return
	}
	parent := &st.frames[len(st.frames)-1]
	parent.vars = merged(f.vars, parent.vars)
	parent.elems = merged(f.elems, parent.elems)
}

// merged is parent with child's touches added, keeping parent's own
// earlier first touch of a name where it has one: nothing ran in it while
// the child span was open. The smaller of the two maps is merged into the
// larger, so a touch deep in a long chain of nested spans is copied a
// logarithmic number of times on its way out, not once per span (V5).
func merged(child, parent map[scopedName]Symbol) map[scopedName]Symbol {
	if len(child) > len(parent) {
		maps.Copy(child, parent)
		return child
	}
	for n, before := range child {
		if _, seen := parent[n]; !seen {
			parent[n] = before
		}
	}
	return parent
}

// UndoFrame closes the span the matching PushFrame started and puts back
// every name touched during it. A span undone this way changed nothing,
// so the span around it is not told of its touches: passing them on
// would only grow every enclosing frame by touches that no longer matter
// (V5).
func (st *SymbolTable) UndoFrame() {
	f := st.frames[len(st.frames)-1]
	st.frames = st.frames[:len(st.frames)-1]
	for _, touched := range []map[scopedName]Symbol{f.vars, f.elems} {
		for n, before := range touched {
			st.put(n, before)
		}
	}
}

// recordFrame notes, in the innermost open frame, that name in scope is
// about to be changed from before -- unless some earlier touch in that
// frame already recorded it, since a frame's value is what name held when
// the frame started, not after whatever touched it most recently inside
// it.
func (st *SymbolTable) recordFrame(scope *Scope, name string, before Symbol) {
	if len(st.frames) == 0 {
		return
	}
	f := st.frames[len(st.frames)-1].vars
	if strings.Contains(name, "[") {
		f = st.frames[len(st.frames)-1].elems
	}
	n := scopedName{scope, name}
	if _, seen := f[n]; !seen {
		f[n] = before
	}
}

// NewSymbolTable creates a fresh symbol table rooted at the implicit "main"
// package scope.
func NewSymbolTable() *SymbolTable {
	root := &Scope{
		Name:    "main",
		parent:  nil,
		symbols: make(map[string]Symbol),
		open:    true,
	}
	return &SymbolTable{root: root, current: root, blockScopes: map[lower.ID]*Scope{}, holders: map[string][]*Scope{}}
}

// Define adds sym to the current (innermost) scope, overwriting any
// same-named symbol already in this scope.
func (st *SymbolTable) Define(sym Symbol) {
	st.put(scopedName{st.current, sym.Name}, sym)
}

// Lookup searches the current scope and each ancestor for name, returning the
// first match found.  The inner-most definition wins (shadowing).
func (st *SymbolTable) Lookup(name string) (Symbol, bool) {
	if s := st.scopeOf(name); s != nil {
		return s.symbols[name], true
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
		open:    true,
		depth:   st.current.depth + 1,
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
	s.open = true
	s.depth = st.current.depth + 1
	st.current = s
	for name := range s.symbols {
		st.holders[name] = append(st.holders[name], s)
	}
}

// scopeOf is the innermost scope on the current chain that holds name,
// or nil.
//
// An element guard shadow, `$r[0]` (flow.go's guardVarName), speaks for
// the @r its scope sees: a `my @r` in a scope nested inside it declares
// another array, whose elements the shadow says nothing of, so there it
// is hidden. perl 5.42 dies "Can't bless non-reference value" on `my @r
// = ({}); if (ref $r[0]) { my @r = (undef); bless $r[0], "Foo" }`.
func (st *SymbolTable) scopeOf(name string) *Scope {
	h := st.holders[name]
	if len(h) == 0 {
		return nil
	}
	s := h[len(h)-1]
	if array, _, ok := strings.Cut(name, "["); ok && strings.HasPrefix(array, "$") {
		if a := st.holders["@"+array[1:]]; len(a) > 0 && a[len(a)-1].depth > s.depth {
			return nil
		}
	}
	return s
}

// inScope reports whether s is on the current chain: the current scope or
// one that encloses it.
func (st *SymbolTable) inScope(s *Scope) bool {
	return s.open
}

// Unreached marks the rest of the current scope as code that never runs,
// past a statement that always exits: what it changes is put back when
// the scope is exited, for it never happens on a path that leaves the
// scope. perl 5.42 dies "Can't bless non-reference value" on `my $y =
// undef; eval { die 'x'; my $q = $y->{k} }; bless $y, 'Foo'`. Nothing
// leaves the root scope, which is never exited, so it is not marked.
func (st *SymbolTable) Unreached() {
	if st.current.parent != nil && !st.current.unreached {
		st.current.unreached = true
		st.PushFrame()
	}
}

// ExitScope pops the current scope back to its parent, undoing what the
// scope changed past a statement that always exits, as Unreached says.
// Calling ExitScope on the root scope is a no-op to avoid panics from
// unbalanced enter/exit pairs.
func (st *SymbolTable) ExitScope() {
	if st.current.parent != nil {
		if st.current.unreached {
			st.current.unreached = false
			st.UndoFrame()
		}
		for name := range st.current.symbols {
			h := st.holders[name]
			if len(h) == 0 || h[len(h)-1] != st.current {
				panic(fmt.Sprintf("symbol table: the scope being exited holds %s, but is not the innermost scope its holders list", name))
			}
			st.holders[name] = h[:len(h)-1]
		}
		st.current.open = false
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
	s := st.scopeOf(name)
	if s == nil {
		return false
	}
	sym := s.symbols[name]
	f(&sym)
	st.set(scopedName{s, name}, sym)
	return true
}

// set writes sym as n's value in the scope n names, recording the touch
// as update does: for a caller that already holds the scope, and so need
// not walk the chain to find it again.
func (st *SymbolTable) set(n scopedName, sym Symbol) {
	st.recordFrame(n.scope, n.name, n.scope.symbols[n.name])
	st.put(n, sym)
}

// remove deletes n from the scope n names, recording the touch as Delete
// does, for a caller that already holds the scope.
func (st *SymbolTable) remove(n scopedName) {
	if sym, ok := n.scope.symbols[n.name]; ok {
		st.recordFrame(n.scope, n.name, sym)
		st.drop(n)
	}
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
	if s := st.scopeOf(name); s != nil {
		st.remove(scopedName{s, name})
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
				st.recordFrame(s, k, sym)
				st.drop(scopedName{s, k})
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
