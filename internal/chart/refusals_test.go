// ABOUTME: Tests spec §12's refusals: each alternative the ledger marks refused in action or by precedence is refused.
// ABOUTME: Each has testdata/refusals/<line>/refused.pl, which the chart refuses, beside accepted.pl, which it accepts.

package chart

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// ledgerRefusals checks dir against the ledger's refused alternatives: <line>/refused.pl must be refused and
// <line>/accepted.pl, its nearby form perl accepts, accepted; dir holds no other line.
func ledgerRefusals(t *testing.T, ledger []grammar.Entry, dir string) []string {
	t.Helper()
	var errs []string
	n := 0
	for _, e := range ledger {
		if e.Mark != "excluded refused-in-action" && e.Mark != "refused-by-precedence" {
			continue
		}
		n++
		for name, want := range map[string]bool{"refused.pl": false, "accepted.pl": true} {
			src, err := os.ReadFile(filepath.Join(dir, strconv.Itoa(e.Line), name))
			if err != nil {
				errs = append(errs, err.Error())
			} else if v := recognizeSource(t, src); v.accepted != want {
				errs = append(errs, fmt.Sprintf("perly.y:%d (%s): %s is accepted=%v at byte %d, want accepted=%v", e.Line, e.Mark, name, v.accepted, v.at, want))
			}
		}
	}
	if ds, _ := filepath.Glob(filepath.Join(dir, "*")); n == 0 || len(ds) != n {
		errs = append(errs, fmt.Sprintf("the ledger marks %d refused alternatives and %s holds %d", n, dir, len(ds)))
	}
	return errs
}

func TestLedgerRefusals(t *testing.T) {
	text, err := os.ReadFile(filepath.Join("grammar", "testdata", "perly", "ledger.txt"))
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := grammar.ParseEntries(string(text), true)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range ledgerRefusals(t, ledger, filepath.Join("testdata", "refusals")) {
		t.Error(e)
	}
}

// TestLedgerRefusalsDetects shows the check failing both ways on a toy ledger, and passing on a sound pair.
func TestLedgerRefusalsDetects(t *testing.T) {
	const refused, accepted = "$y = $a <=> $b <=>;", "$y = $a <=> $b;"
	ledger := []grammar.Entry{{Line: 1, Text: "toy: toy", Mark: "excluded refused-in-action"}}
	for name, c := range map[string]struct {
		refused, accepted string
		fails             bool
	}{
		"sound pair":                 {refused, accepted, false},
		"refused sentence accepted":  {accepted, accepted, true},
		"accepted neighbour refused": {refused, refused, true},
	} {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "1"), 0o755); err != nil {
			t.Fatal(err)
		}
		for f, src := range map[string]string{"refused.pl": c.refused, "accepted.pl": c.accepted} {
			if err := os.WriteFile(filepath.Join(dir, "1", f), []byte(src), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if errs := ledgerRefusals(t, ledger, dir); (len(errs) > 0) != c.fails {
			t.Errorf("%s: errors %q, want failure=%v", name, errs, c.fails)
		}
	}
}
