//go:build absencereference

// ABOUTME: The absence claims' candidate judgement held to lex.Map over every T1 file for every key, once per key.
// ABOUTME: It maps everything, so it runs only under the build tag absencereference.

package main

import (
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
