// ABOUTME: The forest-equality differential: a candidate's forest against V1's pipeline, accept or refuse and reachable nodes.
// ABOUTME: Nodes are matched by symbol, rule position and span; the V1 side, the candidate and ForestDiff are exported for reuse.

package forest

import (
	"cmp"
	"fmt"
	"maps"
	"slices"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// NodeKey is what a node derives, by which ForestDiff matches nodes across forests: its symbol (-1 for a rule
// position), its rule position (Rule and Dot; Rule -1 for a symbol) and its span.
type NodeKey struct{ Sym, Rule, Dot, From, To int32 }

// noNode is the key of an absent child, index 0, which no node's key equals.
var noNode = NodeKey{-1, -1, -1, -1, -1}

// altKey is one kept alternative by what it derives: its rule and its children's keys.
type altKey struct {
	Rule        int32
	Left, Right NodeKey
}

// ForestDiff is where cand's forest differs from v1's, nil for a refusal: the verdict, then a node reachable from
// the root in one only, then the first reachable node, by key, whose kept alternatives differ in content or order;
// "" where they are equal. A node unreachable from the root is never compared.
func ForestDiff(cand, v1 *Forest) string {
	switch {
	case cand == nil && v1 == nil:
		return ""
	case cand == nil:
		return "the candidate refuses, V1 accepts"
	case v1 == nil:
		return "the candidate accepts, V1 refuses"
	}
	c, v := reachableAlts(cand), reachableAlts(v1)
	keys := slices.SortedFunc(maps.Keys(c), cmpNodeKey)
	for _, k := range slices.SortedFunc(maps.Keys(v), cmpNodeKey) {
		if _, ok := c[k]; !ok {
			return fmt.Sprintf("node %+v is reachable in V1 only", k)
		}
	}
	for _, k := range keys {
		if _, ok := v[k]; !ok {
			return fmt.Sprintf("node %+v is reachable in the candidate only", k)
		}
	}
	for _, k := range keys {
		if !slices.Equal(c[k], v[k]) {
			return fmt.Sprintf("node %+v: the candidate keeps %+v, V1 %+v", k, c[k], v[k])
		}
	}
	return ""
}

// reachableAlts is each node reachable from f's root, by key, with its kept alternatives in order.
func reachableAlts(f *Forest) map[NodeKey][]altKey {
	key := func(n int32) NodeKey {
		if n == 0 {
			return noNode
		}
		nd := f.Nodes[n]
		return NodeKey{nd.Sym, nd.Rule, nd.Dot, nd.From, nd.To}
	}
	out := map[NodeKey][]altKey{}
	seen := make([]bool, len(f.Nodes))
	work := []int32{f.Root}
	for len(work) > 0 {
		n := work[len(work)-1]
		work = work[:len(work)-1]
		if seen[n] {
			continue
		}
		seen[n] = true
		alts := []altKey{}
		for a := f.Nodes[n].Alt; a != 0; a = f.Alts[a].Next {
			x := f.Alts[a]
			alts = append(alts, altKey{x.Rule, key(x.Left), key(x.Right)})
			for _, ch := range []int32{x.Left, x.Right} {
				if ch != 0 {
					work = append(work, ch)
				}
			}
		}
		out[key(n)] = alts
	}
	return out
}

func cmpNodeKey(x, y NodeKey) int {
	return cmp.Or(cmp.Compare(x.From, y.From), cmp.Compare(x.To, y.To), cmp.Compare(x.Sym, y.Sym), cmp.Compare(x.Rule, y.Rule), cmp.Compare(x.Dot, y.Dot))
}

// V1Recognizer is the V1 side's recognizer over tab, and the only place the V1 side gets one: always the
// recognizer that judges no precedence, never production's. Today that is recognize.New's.
func V1Recognizer(tab *grammar.Tables) *recognize.Recognizer {
	return recognize.New(tab, recognize.Options{})
}

// v1LibraryOperators and v1Constrain are the V1 side's own stages, variables so a test can count its entries.
var (
	v1LibraryOperators = libraryOperators
	v1Constrain        = (*Forest).Constrain
)

// V1Forest is V1's forest of src: V1Recognizer's chart, Build, LibraryOperators under scope, then Constrain under
// perly.y's precedence (the V2 design's §5). Where it refuses, the forest is nil and the Refusal names the stage
// that refused (grammar, scope or precedence) with Byte 0: the differential compares verdicts, not bytes.
func V1Forest(tab *grammar.Tables, src []byte, scope Scope) (*Forest, Refusal) {
	s, _ := lex.Lex(src, tab.Tokens)
	l := lex.Linear(s)
	c, err := recognize.Recognize(V1Recognizer(tab), l)
	if err != nil {
		return nil, Refusal{Reason: err.Error()}
	}
	if !c.Accepted() {
		return nil, Refusal{Reason: "grammar"}
	}
	f := Build(tab, c, l)
	if v1LibraryOperators(tab, f, s, src, scope) >= 0 {
		return nil, Refusal{Reason: "scope"}
	}
	prec, err := ParsePrecedence(tab)
	if err != nil {
		return nil, Refusal{Reason: err.Error()}
	}
	if !v1Constrain(f, prec) {
		return nil, Refusal{Reason: "precedence"}
	}
	return f, Refusal{}
}

// V1ToyForest is V1's forest of a toy grammar's lattice l under the toy's precedence p: V1Recognizer's chart,
// Build, then Constrain; nil where it refuses. A toy has no source, so it has no library operators.
func V1ToyForest(tab *grammar.Tables, l *lex.Lattice[lex.None], p *Precedence) (*Forest, error) {
	c, err := recognize.Recognize(V1Recognizer(tab), l)
	if err != nil || !c.Accepted() {
		return nil, err
	}
	f := Build(tab, c, l)
	if !v1Constrain(f, p) {
		return nil, nil
	}
	return f, nil
}
