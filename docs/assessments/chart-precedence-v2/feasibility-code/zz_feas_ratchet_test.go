package main

import (
	"path/filepath"
	"tamarou.com/pvm/internal/chart/grammar"
	"testing"

	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

func TestZZFeasCorpus(t *testing.T) {
	t.Chdir("../../../..")
	files, srcs, err := corpus()
	if err != nil {
		t.Fatal(err)
	}
	tab := committedTables(t)
	prec, err := forest.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	perl, err := labelVerdicts(filepath.Join(testdata, "labels"))
	if err != nil {
		t.Fatal(err)
	}
	scope := chartInputs(tab, srcs, nil, perl, realDirs).Scope
	r := recognize.New(tab, recognize.Options{})
	forest.ZZStaticTops = forest.ZZM3StaticTops(tab, prec)
	var m3 forest.ZZM3Stats
	var m4, m4full forest.ZZM4Stats
	nfiles, recognized, prelib, judged := 0, 0, 0, 0
	n1626, n1626done := 0, 0
	for _, file := range files {
		nfiles++
		src := srcs[file.Path]
		s, _ := lex.Lex(src, tab.Tokens)
		l := lex.Linear(s)
		c, err := recognize.Recognize(r, l)
		if err != nil {
			t.Fatalf("%s: %v", file.Path, err)
		}
		for k, tk := range s.Tokens {
			if tk.TerminalSet == 1626 {
				n1626++
				if c.Sets[k] != nil {
					for _, it := range c.Sets[k].Items() {
						for _, ci := range tab.States[it.State].Items {
							if ci == int(tab.Items[0].Rule)*0+zzIndirobDone(tab) {
								n1626done++
							}
						}
					}
				}
			}
		}
		sc := scope(file.Path)
		dead := forest.ZZDead(tab, s, src, sc)
		f := forest.Build(tab, c, l)
		if f == nil {
			// no forest: measure 4 on the chart only, no observed tops
			forest.ZZM4(tab, prec, c, s, dead, nil, &m4)
			forest.ZZM4(tab, prec, c, s, map[[2]int]bool{}, nil, &m4full)
			continue
		}
		recognized++
		if forest.LibraryOperators(tab, f, s, src, sc) >= 0 {
			prelib++
			forest.ZZM4(tab, prec, c, s, dead, nil, &m4)
			forest.ZZM4(tab, prec, c, s, map[[2]int]bool{}, nil, &m4full)
			continue
		}
		judged++
		rec := forest.ZZShifts(f, prec)
		forest.ZZM3(tab, prec, c, func(j int) int {
			if j < len(l.Edges) {
				return l.Edges[j].Token.TerminalSet
			}
			return -1
		}, rec, &m3)
		forest.ZZM4(tab, prec, c, s, dead, rec, &m4)
		forest.ZZM4(tab, prec, c, s, map[[2]int]bool{}, rec, &m4full)
	}
	t.Logf("tokens with terminal set 1626 {HASHBRACK PERLY_BRACE_OPEN}=%d; with indirob ::= scalar . complete before them=%d", n1626, n1626done)
	t.Logf("files=%d recognized(forest)=%d libops-refused=%d judged=%d", nfiles, recognized, prelib, judged)
	t.Logf("M3: shift records=%d chart items holding one=%d unmatched records=%d records in >1 state item=%d", m3.Shifts, m3.Items, m3.Unmatched, m3.Ambiguous)
	t.Logf("M3: items with >=2 canonical tops=%d (same core item >=2: %d; core items differ: %d); needing >=2 verdict vectors=%d (cross-core: %d); over-keep under one shared context=%d",
		m3.MultiTop, m3.SameCore, m3.CrossCore, m3.NeedCopies, m3.NeedCopiesCross, m3.OverKeep)
	for _, e := range forest.ZZTop(m3.Examples, 15) {
		t.Logf("M3 cross ex %s", e)
	}
	t.Logf("M3: records whose top the static over-approximation lacks=%d", m3.StaticMiss)
	for _, e := range forest.ZZTop(m3.StaticMissEx, 20) {
		t.Logf("M3 static-miss ex %s", e)
	}
	for _, e := range forest.ZZTop(m3.UnmatchedEx, 20) {
		t.Logf("M3 unmatched ex %s", e)
	}
	for _, e := range forest.ZZTop(m3.OverExamples, 15) {
		t.Logf("M3 over ex %s", e)
	}
	for name, st := range map[string]*forest.ZZM4Stats{"narrowed": &m4, "unnarrowed": &m4full} {
		t.Logf("M4 %s: tokens=%d plugin-before=%d plugin-after=%d multi-terminal(eff>=2)=%d reduce-diff=%d reduce-diff-via-slipper-;=%d shift-diff(any top)=%d shift-diff(observed top)=%d any(reduce|shift-any)=%d",
			name, st.Tokens, st.Plugin, st.PluginAfter, st.Multi, st.ReduceDiff, st.ReduceDiffSlip, st.ShiftDiffAny, st.ShiftDiffObserved, st.AnyDiff)
		for _, e := range forest.ZZTop(st.Examples, 25) {
			t.Logf("M4 %s ex %s", name, e)
		}
	}
}

func zzIndirobDone(tab *grammar.Tables) int {
	for i, it := range tab.Items {
		if it.Rule == 347 && it.Dot == len(tab.Rules[347].RHS) {
			return i
		}
	}
	return -1
}
