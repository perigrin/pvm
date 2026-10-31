// ABOUTME: Tests FIRST and FOLLOW on the tables, against sets worked out by hand on toy grammars.
// ABOUTME: Nullable symbols, left and right recursion, and the end of input are each pinned.

package grammar

import (
	"slices"
	"strings"
	"testing"
)

// setNames is the sorted names of the symbol ids in set, the end of input as "$".
func setNames(tab *Tables, set []int) string {
	var names []string
	for _, id := range set {
		if id == tab.EOF {
			names = append(names, "$")
		} else {
			names = append(names, tab.Symbols[id].Name)
		}
	}
	slices.Sort(names)
	return strings.Join(names, " ")
}

func checkSets(t *testing.T, tab *Tables, kind string, sets [][]int, want map[string]string) {
	t.Helper()
	for nt, w := range want {
		id := slices.IndexFunc(tab.Symbols, func(s TableSymbol) bool { return s.Name == nt })
		if got := setNames(tab, sets[id]); got != w {
			t.Errorf("%s(%s) = {%s}, want {%s}", kind, nt, got, w)
		}
	}
}

// TestFirstFollowOnTables: S is left recursive and nullable through B and C,
// C is right recursive, B is nullable.
//
//	S ::= S a | B C    B ::= b | (empty)    C ::= c C | (empty)
func TestFirstFollowOnTables(t *testing.T) {
	tab := build(t, toy([]string{"a", "b", "c"}, alts("S", "S a", "B C"), alts("B", "b", ""), alts("C", "c C", "")))
	checkSets(t, tab, "FIRST", tab.First, map[string]string{"S": "a b c", "B": "b", "C": "c", "a": "a"})
	checkSets(t, tab, "FOLLOW", tab.Follow, map[string]string{"S": "$ a", "B": "$ a c", "C": "$ a"})
}

// TestFirstFollowNullableSuffix: X is followed only by a nullable N, so it
// inherits FOLLOW(S); a nullable start symbol is followed by the end of input.
func TestFirstFollowNullableSuffix(t *testing.T) {
	tab := build(t, toy([]string{"x"}, alts("S", "X N"), alts("X", "x"), alts("N", "")))
	checkSets(t, tab, "FOLLOW", tab.Follow, map[string]string{"S": "$", "X": "$", "N": "$"})
	tab = build(t, toy(nil, alts("S", "N"), alts("N", "")))
	checkSets(t, tab, "FOLLOW", tab.Follow, map[string]string{"S": "$", "N": "$"})
}
