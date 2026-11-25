# How many spine-top contexts meet at a node (V2 sizing)

Measured 2026-11-25 on pu c383f861, for `docs/plans/2026-11-25-chart-precedence-v2-design.md` §3.

## Question

If the recognizer carries each item's spine-top context (V2's shift side), how many items split? An item needs a second copy only where two contexts reach it and keep different alternatives.

## Method

V1's state-free judgement already evaluates each forest node under its canonical spine top: `canon` collapses any spine top that forbids no shift to -1. Its memo holds one eval entry per (node, terminal, spine top, parent item). A scratch helper ran the judgement on each file's forest after `LibraryOperators`, under the pipeline's own scope, the way `TestPrecedenceCorpusMatchesReference` builds it. For every judged node it counted two things:

- the distinct canonical spine tops;
- within each (terminal, parent item) group, the distinct kept-alternative sets.

The largest of those group counts is the number of copies the node needs if contexts split only where they change the outcome.

The population is the in-tree corpora (T1, guacamole, std5k, edit5k), with `T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit`. That gave 1,839 recognized files that were not refused before the constraint. perl.git `t/` was not included.

## Result

```
files=1839 nodes=1410324 multi-top=3425 (0.24%) outcome-differs=145 (0.0103%) avg-tops=1.0024 avg-needed=1.0001 max-tops=3
```

| Measure | Value |
|---|---|
| Judged nodes | 1,410,324 |
| Nodes under 2 or more canonical spine tops | 3,425 (0.24%) |
| Nodes whose tops keep different alternatives | 145 (0.010%) |
| Mean distinct tops per node | 1.0024 |
| Mean copies needed per node | 1.0001 |
| Most tops at one node | 3 |

The counter is not vacuous: it finds 3,425 meetings and 145 real conflicts.

## Limits

- It counts forest nodes, which are completed items that survive into the forest. It does not count every item the recognizer creates. Items that lead nowhere never become nodes, so chart growth may be somewhat higher, though it is bounded by the same three contexts.
- perl.git `t/` was not measured.

## Scratch code

This code was not committed. To reproduce, place it in a worktree at pu c383f861 and run `T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit go test ./internal/chart/cmd/ratchet/ -run '^TestZZContextSplit$' -count=1 -v`.

`internal/chart/forest/zz_contexts.go`:

```go
package forest

import (
	"fmt"
	"slices"
)

type ContextStats struct {
	Nodes, Multi, Differ, SumTops, SumNeeded, MaxTops int
}

func ContextSplit(f *Forest, p *Precedence, s *ContextStats) {
	c := newJudgement(f, p)
	c.constrain()
	for n := range f.Nodes {
		h := c.evalAt[n]
		if h == 0 {
			continue
		}
		tops := map[int32]bool{}
		groups := map[[2]int32]map[string]bool{}
		for e := h; e != 0; e = c.evals[e].next {
			x := c.evals[e]
			tops[x.top] = true
			k := [2]int32{x.t, x.pi}
			if groups[k] == nil {
				groups[k] = map[string]bool{}
			}
			o := slices.Clone(x.out)
			slices.Sort(o)
			groups[k][fmt.Sprint(o)] = true
		}
		need := 1
		for _, g := range groups {
			need = max(need, len(g))
		}
		s.Nodes++
		s.SumTops += len(tops)
		s.SumNeeded += need
		s.MaxTops = max(s.MaxTops, len(tops))
		if len(tops) > 1 {
			s.Multi++
		}
		if need > 1 {
			s.Differ++
		}
	}
}
```

`internal/chart/cmd/ratchet/zz_contexts_test.go`: the loop of `TestPrecedenceCorpusMatchesReference`. For each file it runs Lex, Linear, Recognize and Build, skips a file that `LibraryOperators` refuses, then calls `forest.ContextSplit(f, prec, &s)` and logs the totals.
