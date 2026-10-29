// ABOUTME: Tests that the ratchet prints every contract key once, in a fixed order, na where nothing measured it.
// ABOUTME: The measures themselves need perl and the corpus, so make chart-ratchet exercises them, not this test.

package main

import (
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

// report must not invent a key: a measure with no slot in the order is a bug in the ratchet.
func TestReportRejectsUnknownKey(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("report accepted a key that is not in the fixed order")
		}
	}()
	report(map[string]string{"not-a-key": "1"})
}
