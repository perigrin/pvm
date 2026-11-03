// ABOUTME: Tests that leo-disagree runs H2c's Leo on/off comparison over the files it is given and prints a count.
// ABOUTME: Over the committed grammar's tables, so the count is the one the ratchet prints.

package main

import (
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

func TestLeoDisagreeMeasured(t *testing.T) {
	g, ref, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	tab, err := grammar.Generate(g, ref)
	if err != nil || tab == nil {
		t.Fatalf("the committed grammar generates no tables: %v", err)
	}
	srcs := map[string][]byte{"a.pl": []byte("use strict;\n{ $x = 1 }\n"), "b.pl": []byte("$x = = 1;")}
	files := []File{{"a.pl", "t"}, {"b.pl", "t"}}
	measured := map[string]string{}
	if err := measureLeo(tab, files, srcs, measured); err != nil {
		t.Fatal(err)
	}
	if got := measured["leo-disagree"]; got != "0" {
		t.Errorf("leo-disagree %q, want 0", got)
	}
}
