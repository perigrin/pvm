// ABOUTME: Tests spec §12's refusals: each alternative the ledger marks refused in action or by precedence is refused.
// ABOUTME: Each has testdata/refusals/<line>/refused.pl, which the chart refuses, beside accepted.pl, which it accepts.
// ABOUTME: Also testdata/port/refused-where.txt: where, and by the token table or the grammar, each refused port snippet is refused.

package chart

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
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

// refusedWherePath is the reviewed list of where each refused port snippet
// is refused: one `<snippet> <table|grammar> <byte>` line per snippet, the
// snippet relative to testdata/port, "table" for chart/lex's lex-unmapped
// and "grammar" for the recognizer.
var refusedWherePath = filepath.Join("testdata", "port", "refused-where.txt")

// refusedWhere is each snippet under testdata/port/*-refused, relative to
// testdata/port, mapped to its list line's verdict, "<table|grammar> <byte>";
// an accepted one maps to "accepted".
func refusedWhere(t *testing.T) map[string]string {
	t.Helper()
	fs, err := filepath.Glob(filepath.Join("testdata", "port", "*-refused", "*.pl"))
	if err != nil || len(fs) == 0 {
		t.Fatalf("no refused port snippets: %v", err)
	}
	out := map[string]string{}
	for _, f := range fs {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		name, _ := filepath.Rel(filepath.Join("testdata", "port"), f)
		out[filepath.ToSlash(name)] = "accepted"
		if v := recognizeSource(t, src); !v.accepted {
			out[filepath.ToSlash(name)] = fmt.Sprintf("%s %d", v.by, v.at)
		}
	}
	return out
}

// readRefusedWhere is the list's lines as snippet to verdict, and a
// complaint per line that is malformed or names its snippet twice.
func readRefusedWhere(t *testing.T) (map[string]string, []string) {
	t.Helper()
	data, err := os.ReadFile(refusedWherePath)
	if err != nil {
		t.Fatal(err)
	}
	list, errs := map[string]string{}, []string(nil)
	for i, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
		f := strings.Fields(line)
		if len(f) != 3 || (f[1] != "table" && f[1] != "grammar") {
			errs = append(errs, fmt.Sprintf("line %d, %q: want `<snippet> <table|grammar> <byte>`", i+1, line))
			continue
		}
		if _, err := strconv.Atoi(f[2]); err != nil {
			errs = append(errs, fmt.Sprintf("line %d, %q: the byte is not a number", i+1, line))
			continue
		}
		if _, dup := list[f[0]]; dup {
			errs = append(errs, fmt.Sprintf("line %d: %s is listed twice", i+1, f[0]))
		}
		list[f[0]] = f[1] + " " + f[2]
	}
	return list, errs
}

// whereDiffs is a complaint for each snippet whose refusal, in got, is not
// the one list records, sorted: refused by the other cause or at another
// byte, or refused and not listed.
func whereDiffs(list, got map[string]string) []string {
	var out []string
	for name, v := range got {
		if v == "accepted" {
			continue // its family's port test reports it
		}
		if want, ok := list[name]; !ok {
			out = append(out, fmt.Sprintf("%s is refused and not listed: add `%s %s` once reviewed", name, name, v))
		} else if want != v {
			out = append(out, fmt.Sprintf("%s: listed as refused by %s, now refused by %s: review it, then update the list", name, want, v))
		}
	}
	slices.Sort(out)
	return out
}

// TestRefusedSnippetsRefusedWhere: every refused port snippet is refused
// where refused-where.txt says, by the token table or by the grammar.
func TestRefusedSnippetsRefusedWhere(t *testing.T) {
	list, _ := readRefusedWhere(t)
	got := refusedWhere(t)
	for _, d := range whereDiffs(list, got) {
		t.Error(d)
	}
	table := 0
	for _, v := range got {
		if strings.HasPrefix(v, "table ") {
			table++
		}
	}
	t.Logf("%d of %d refused snippets are refused by the token table", table, len(got))
}

// TestRefusedWhereListHasNoStaleEntries: each refused-where.txt line is
// well formed and names, once, a refused snippet that exists.
func TestRefusedWhereListHasNoStaleEntries(t *testing.T) {
	list, errs := readRefusedWhere(t)
	got := refusedWhere(t)
	for name := range list {
		switch got[name] {
		case "":
			errs = append(errs, fmt.Sprintf("%s is listed and is no refused snippet", name))
		case "accepted":
			errs = append(errs, fmt.Sprintf("%s is listed and is accepted", name))
		}
	}
	slices.Sort(errs)
	for _, e := range errs {
		t.Error(e)
	}
}

// TestRefusedWhereDetectsMovedSnippet: a snippet refused by the other cause,
// or at another byte, than its line says fails the check until the list is
// reviewed; the list as committed passes it.
func TestRefusedWhereDetectsMovedSnippet(t *testing.T) {
	list, _ := readRefusedWhere(t)
	got := refusedWhere(t)
	if d := whereDiffs(list, got); len(d) != 0 {
		t.Fatalf("the committed list does not hold: %v", d)
	}
	for _, by := range []string{"table", "grammar"} {
		name := ""
		for n, v := range got {
			if strings.HasPrefix(v, by+" ") && (name == "" || n < name) {
				name = n
			}
		}
		if name == "" {
			t.Fatalf("no snippet is refused by the %s", by)
		}
		other := map[string]string{"table": "grammar", "grammar": "table"}[by]
		var at int
		fmt.Sscanf(got[name], by+" %d", &at)
		for _, moved := range []string{fmt.Sprintf("%s %d", other, at), fmt.Sprintf("%s %d", by, at+1)} {
			was := got[name]
			got[name] = moved
			if d := whereDiffs(list, got); len(d) != 1 || !strings.Contains(d[0], name) {
				t.Errorf("%s moved from %s to %s: the check says %v", name, was, moved, d)
			}
			got[name] = was
		}
	}
}
