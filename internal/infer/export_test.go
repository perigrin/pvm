// ABOUTME: Exposes the lowered path's unexported helpers to the infer_test package's tests.
// ABOUTME: It is compiled only with the tests, so the package's API does not grow.

package infer

import (
	"slices"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// The lowered pass 1 and the helpers the guard and parameter rules read
// a sub with, under the names the tests call them by.
var (
	CollectDeclarations = collectDeclarations
	FileSubs            = fileSubs
	SubParam            = subParam
	SubParams           = subParams
	SubReturnExpr       = subReturnExpr
	ArgVarName          = argVarName
	Finite              = finite
	SpaceshipResult     = spaceshipResult
	Numified            = numified
)

// SortSelection is the selection the call site makes for the first sort of
// src, analysed as Analyze analyses it.
func SortSelection(src string) types.Selection {
	root := parse.Parse([]byte(src))
	f := lower.Lower(root)
	a := newAnalysis(f, lower.NewPositions(root, f), nil, Options{})
	a.walk(f)
	var sel types.Selection
	var find func(n lower.Node) bool
	find = func(n lower.Node) bool {
		if s, ok := n.(*lower.MapGrepSort); ok && s.Op == "sort" {
			sel = a.sortSelection(s)
			return true
		}
		return slices.ContainsFunc(lower.Children(n), find)
	}
	find(f)
	return sel
}

// SetFlowTouches is how many names setFlow records as touched when it
// sets name in st to typ, unmarked: what restoring or joining a variable
// costs the frames open around it.
func SetFlowTouches(st *SymbolTable, name string, typ types.Type) int {
	a := &analysis{st: st}
	st.PushFrame()
	a.setFlow(name, flowVar{typ, false})
	n := len(st.Touched())
	st.PopFrame()
	return n
}
