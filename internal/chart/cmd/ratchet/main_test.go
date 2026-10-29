// ABOUTME: Tests that the ratchet prints every contract key once, in a fixed order, na where nothing measured it.
// ABOUTME: The measures themselves need perl and the corpus, so make chart-ratchet exercises them, not this test.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// contract is the milestone's key list, which the ratchet may add to and never rename.
var contract = []string{
	"parses-scored", "parsent-scored",
	"parses-refused", "parses-wrong-sites", "yield-violations", "oracle-below-rd", "std5k-refused", "parsent-accepted",
	"speed-ratio-aggregate", "speed-ratio-std5k", "set-reuse", "convergence-edits", "convergence-within-8",
	"defect-yield-listed", "defect-yield-sites", "defect-yield-settled", "oracle-sites-decided", "expect-states-unreproduced",
}

func TestReportPrintsEveryKeyOnceInOrder(t *testing.T) {
	ms := report(map[string]string{"parses-scored": "746", "rd-oracle-exact": "0.5"})
	seen := map[string]int{}
	for i, m := range ms {
		if seen[m.Key]++; seen[m.Key] > 1 {
			t.Fatalf("key %s printed twice", m.Key)
		}
		if m.Key != keys[i] {
			t.Fatalf("line %d is %s, want %s: the order is fixed", i, m.Key, keys[i])
		}
	}
	for _, k := range append(contract, "rejects-disagree", "rd-oracle-exact", "rd-oracle-coverage") {
		if seen[k] != 1 {
			t.Fatalf("contract key %s printed %d times, want once", k, seen[k])
		}
	}
	values := map[string]string{}
	for _, m := range ms {
		values[m.Key] = m.Value
	}
	if values["parses-scored"] != "746" || values["rd-oracle-exact"] != "0.5" {
		t.Fatalf("measured values not carried through: %v", values)
	}
	if values["set-reuse"] != "na" || values["std5k-refused"] != "na" {
		t.Fatalf("an unmeasured key must print na: %v", values)
	}
}

// TestTablesDrift: the committed tables are what the generator writes, and
// one byte added to one of them, in a copy, is one file of drift.
func TestTablesDrift(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	slif := filepath.Join(root, grammarSLIF)
	drift, err := tablesDrift(filepath.Join(root, tablesDir), slif)
	if err != nil {
		t.Fatal(err)
	}
	if len(drift) != 0 {
		t.Fatalf("drift on the committed tables: %v", drift)
	}

	copied := t.TempDir()
	paths, err := filepath.Glob(filepath.Join(root, tablesDir, "*.txt"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no committed tables: %v", err)
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if p == paths[0] {
			b = append(b, '\n')
		}
		if err := os.WriteFile(filepath.Join(copied, filepath.Base(p)), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	drift, err = tablesDrift(copied, slif)
	if err != nil {
		t.Fatal(err)
	}
	if len(drift) != 1 || !strings.Contains(drift[0], filepath.Base(paths[0])) {
		t.Fatalf("drift after altering %s = %v, want that one file", filepath.Base(paths[0]), drift)
	}
}

// report must not invent a key: a measure with no slot in the order is a bug in the ratchet.
func TestReportRejectsUnknownKey(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("report accepted a key that is not in the fixed order")
		}
	}()
	report(map[string]string{"not-a-key": "1"})
}
