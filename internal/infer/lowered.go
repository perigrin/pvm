// ABOUTME: AnalyzeLowered: PSC type inference over the lowered tree, keyed by node id rather than byte offset.
// ABOUTME: It runs alongside the legacy tree-sitter Analyze; the ports move one rule slice at a time onto it.

package infer

import (
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/types"
)

// AnalyzeLowered runs inference over f, which lower.Lower returned, and
// returns each node's type keyed by its id, the diagnostics, and the symbol
// table. pos is f's position map: diagnostics and symbols carry byte spans,
// and the lowered tree has none of its own.
func AnalyzeLowered(f *lower.File, pos *lower.Positions) (map[lower.ID]types.Type, []Diagnostic, *SymbolTable) {
	return map[lower.ID]types.Type{}, nil, NewSymbolTable()
}
