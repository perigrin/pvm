//go:build absencereference

// ABOUTME: The absence claims' references, which map the whole corpus once per claim: make chart-absence-reference runs them.
// ABOUTME: falseAbsences against a table per claim, and its candidate judgement against Map over every T1 file for every key.

package main

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/lexer"
)

// TestAbsenceCandidatesMatchFullMap: for every key of the committed table, compiled alone as an absence claim's is,
// and every T1 file, the tokens corpusFile.admits judges the key's table to admit are those Map over the whole file
// under that table admits. That holds the candidates and lex.Trace.Holds's state rule to Map itself. No committed key's
// admissions over T1 move with the state its table decides, so a toy key whose admissions do, and a toy file, are swept with them:
// under the toy's table a `*` after a word completes a term, so the bar after it names nothing and baz's Prev2 holds.
func TestAbsenceCandidatesMatchFullMap(t *testing.T) {
	t.Chdir("../../../..")
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	_, srcs, err := corpus()
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	var files []corpusFile
	for p, src := range srcs {
		if strings.HasPrefix(p, "T1/") {
			paths, files = append(paths, p), append(files, newCorpusFile(src))
		}
	}
	paths, files = append(paths, "toy.pl"), append(files, newCorpusFile([]byte("foo * bar; baz;")))
	word := ptr(lexer.Word)
	toyKey := grammar.ClassDecl{Terminal: "WORD", Key: grammar.TokenClass{Kind: word, Completes: true, Prev2: &grammar.Neighbour{Kind: lexer.Word, Text: grammar.OtherWord}}}
	seen := map[string]bool{}
	whole, admitted, bad := map[string]bool{}, 0, 0
	for _, d := range append(g.Classes, toyKey) {
		k := strings.TrimSpace(d.Terminal + " " + keyText(d.Key))
		if seen[k] {
			continue
		}
		seen[k] = true
		c := d.Key
		c.Terminals = []int{0}
		tt, err := grammar.CompileTokens([]grammar.TokenClass{c})
		if err != nil {
			t.Fatal(err)
		}
		for i, f := range files {
			if !f.tr.Holds(tt) {
				whole[k] = true
			}
			got, want := f.admits(c, tt), mapAdmits(f.src, f.raw, tt)
			admitted += len(want)
			if !slices.Equal(got, want) {
				if bad++; bad <= 10 {
					t.Errorf("%s over %s: judged to admit %v, Map admits %v", k, paths[i], got, want)
				}
			}
		}
	}
	if len(seen) < 1000 || len(files) < 100 || admitted == 0 {
		t.Fatalf("%d keys over %d files, %d admissions: the check proves nothing", len(seen), len(files), admitted)
	}
	t.Logf("%d keys over %d files (T1 and the toy), %d admissions, %d mismatches; %d keys' tables change Map's state on some file", len(seen), len(files), admitted, bad, len(whole))
}

// TestTDTAbsentReasonsMatchPerClaimScan: falseAbsences reports exactly the admissions a table compiled from each claim
// alone reports, over the real corpus, for the keys a shared table once masked (an OtherWord or IdentQuote neighbour,
// which another claim's named word or quote changes the meaning of) and every 8th key besides.
func TestTDTAbsentReasonsMatchPerClaimScan(t *testing.T) {
	t.Chdir("../../../..")
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	tab := committedTables(t)
	_, srcs, err := corpus()
	if err != nil {
		t.Fatal(err)
	}
	names := symbolNames(tab)
	var b strings.Builder
	b.WriteString("reason: x " + absentPhrase + "\n")
	for i, d := range g.Classes {
		k := strings.TrimSpace(d.Terminal + " " + keyText(d.Key))
		masked := strings.Contains(k, grammar.OtherWord) || strings.Contains(k, grammar.IdentQuote)
		if masked && (strings.HasPrefix(k, "METHCALL") || strings.HasPrefix(k, "MULOP")) || i%8 == 0 {
			b.WriteString(k + "\n")
		}
	}
	want := refAbsences(t, g.Classes, names, b.String(), srcs)
	got, n := falseAbsences(t, g.Classes, b.String(), srcs)
	if n < 20 || len(want) == 0 {
		t.Fatalf("checked %d claims, the reference found %d admissions: the sample proves nothing", n, len(want))
	}
	slices.Sort(want)
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("falseAbsences reports %d admissions, a table per claim reports %d; first difference:\n%s", len(got), len(want), firstDiff(got, want))
	}
}

func firstDiff(a, b []string) string {
	for i := range min(len(a), len(b)) {
		if a[i] != b[i] {
			return a[i] + "\nvs\n" + b[i]
		}
	}
	return fmt.Sprintf("one is a prefix of the other, %d vs %d", len(a), len(b))
}

// symbolNames is the grammar's symbol names, which a class's Terminals index.
func symbolNames(tab *grammar.Tables) []string {
	names := make([]string, len(tab.Symbols))
	for i, s := range tab.Symbols {
		names[i] = s.Name
	}
	return names
}
