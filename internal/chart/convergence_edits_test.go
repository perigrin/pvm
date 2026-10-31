// ABOUTME: Holds the fixed convergence-edit lists of spec §8: 200 single-statement edits to std5k.pl, and the corpus's.
// ABOUTME: Guards each edit to stay inside one recursive-descent statement and leave the file accepted.

package chart

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// convergenceEditsPath is std5k.pl's list. It is fixed once committed: a
// later issue may add a list beside it, never edit this one.
var convergenceEditsPath = filepath.Join("testdata", "convergence-edits.jsonl")

// convergenceEditCount is §8's size for the list.
const convergenceEditCount = 200

// convergenceEdit is one line of a list. Offsets are bytes into the file the
// edit applies to: Removed bytes at Offset are replaced by Inserted, and the
// edit lies inside the statement [StmtStart, StmtEnd). Case names the
// corpus case file and is empty for std5k.pl.
type convergenceEdit struct {
	Case      string `json:"case,omitempty"`
	Kind      string `json:"kind"`
	Offset    int    `json:"offset"`
	Removed   int    `json:"removed"`
	Inserted  string `json:"inserted"`
	StmtStart int    `json:"stmt_start"`
	StmtEnd   int    `json:"stmt_end"`
}

// readConvergenceEdits reads a list, one JSON object per line. Every field
// but case is required, so `{}` is an error and not an empty edit at 0.
func readConvergenceEdits(path string, wantCase bool) ([]convergenceEdit, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	required := []string{"kind", "offset", "removed", "inserted", "stmt_start", "stmt_end"}
	if wantCase {
		required = append(required, "case")
	}
	var edits []convergenceEdit
	sc := bufio.NewScanner(f)
	for line := 1; sc.Scan(); line++ {
		where := fmt.Sprintf("%s:%d", filepath.Base(path), line)
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(sc.Bytes(), &fields); err != nil {
			return nil, fmt.Errorf("%s: %w", where, err)
		}
		for _, k := range required {
			if _, ok := fields[k]; !ok {
				return nil, fmt.Errorf("%s: no %q field", where, k)
			}
		}
		var e convergenceEdit
		dec := json.NewDecoder(bytes.NewReader(sc.Bytes()))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&e); err != nil {
			return nil, fmt.Errorf("%s: %w", where, err)
		}
		edits = append(edits, e)
	}
	return edits, sc.Err()
}

// stmtSpan is a statement's byte span.
type stmtSpan struct{ Start, End int }

// editableStatements are the statements an edit may name: recursive
// descent's Statement nodes at or after from that end in `;` and hold no
// such statement inside them. So `my $n = 0;` is one, and a sub definition
// or a sort block's `$a <=> $b` is not.
func editableStatements(root *parse.Node, src []byte, from int) []*parse.Node {
	var out []*parse.Node
	var walk func(n *parse.Node) bool
	walk = func(n *parse.Node) bool {
		inner := false
		for _, c := range n.Children {
			if walk(c) {
				inner = true
			}
		}
		if n.Kind != parse.Statement || n.End == n.Start || src[n.End-1] != ';' {
			return inner
		}
		if !inner && n.Start >= from {
			out = append(out, n)
		}
		return true
	}
	walk(root)
	return out
}

// spanSet is the set of editableStatements' spans.
func spanSet(stmts []*parse.Node) map[stmtSpan]bool {
	set := map[stmtSpan]bool{}
	for _, s := range stmts {
		set[stmtSpan{s.Start, s.End}] = true
	}
	return set
}

// applyEdit returns src with e made.
func applyEdit(src []byte, e convergenceEdit) []byte {
	out := append([]byte{}, src[:e.Offset]...)
	out = append(out, e.Inserted...)
	return append(out, src[e.Offset+e.Removed:]...)
}

// checkEdit says why e is not a single-statement edit to src that recursive
// descent still accepts, or returns nil.
func checkEdit(src []byte, stmts map[stmtSpan]bool, e convergenceEdit) error {
	if e.Offset < 0 || e.Removed < 0 || e.Offset+e.Removed > len(src) {
		return fmt.Errorf("bytes [%d,%d) lie outside the file's %d bytes", e.Offset, e.Offset+e.Removed, len(src))
	}
	if !stmts[stmtSpan{e.StmtStart, e.StmtEnd}] {
		return fmt.Errorf("[%d,%d) is not one of recursive descent's statements", e.StmtStart, e.StmtEnd)
	}
	if e.Offset < e.StmtStart || e.Offset+e.Removed > e.StmtEnd {
		return fmt.Errorf("bytes [%d,%d) cross out of statement [%d,%d)", e.Offset, e.Offset+e.Removed, e.StmtStart, e.StmtEnd)
	}
	if n := refused(parse.Parse(applyEdit(src, e))); n != 0 {
		return fmt.Errorf("the edited file parses with %d refused node(s)", n)
	}
	return nil
}

// checkEdits reports, naming each edit's line in path, every edit that fails
// checkEdit or repeats an earlier one. srcs maps an edit's Case to its file.
func checkEdits(t *testing.T, path string, edits []convergenceEdit, srcs map[string][]byte, stmts map[string]map[stmtSpan]bool) {
	t.Helper()
	seen := map[convergenceEdit]int{}
	for i, e := range edits {
		where := fmt.Sprintf("%s:%d", filepath.Base(path), i+1)
		if first, ok := seen[e]; ok {
			t.Errorf("%s: repeats the edit on line %d", where, first)
			continue
		}
		seen[e] = i + 1
		src, ok := srcs[e.Case]
		if !ok {
			t.Errorf("%s: names case %q, which is not in the list's file set", where, e.Case)
			continue
		}
		if err := checkEdit(src, stmts[e.Case], e); err != nil {
			t.Errorf("%s: %v", where, err)
		}
	}
}

var updateConvergenceEdits = flag.Bool("chart.update-convergence-edits", false,
	"draw the convergence-edit lists afresh; only ever for a list not yet committed")

// convergenceSeed fixes the draw. PCG's output for a seed is specified by
// math/rand/v2, so the same seed draws the same list on any Go release.
const convergenceSeed = 20261029

// editKinds are the four edits §8's list covers, taken in turn.
var editKinds = [...]string{"insert", "delete", "replace", "token"}

// insertedStatement is the statement insert adds and replace substitutes.
const insertedStatement = "my $edit = 1;"

// caseStmt is an editable statement of the file named by Case.
type caseStmt struct {
	Case string
	Node *parse.Node
}

// drawEdits draws up to n edits. Edit i takes kind editKinds[i%4] and a
// random statement from the i-th of n equal runs of stmts, so the list
// spreads across the files in order. Where that statement admits no new
// edit of the kind, the next statement that does is taken.
func drawEdits(stmts []caseStmt, srcs map[string][]byte, n int) []convergenceEdit {
	rng := rand.New(rand.NewPCG(convergenceSeed, 0))
	seen := map[convergenceEdit]bool{}
	var out []convergenceEdit
	for i := 0; i < n; i++ {
		lo := i * len(stmts) / n
		hi := max(lo+1, (i+1)*len(stmts)/n)
		first := lo + rng.IntN(hi-lo)
		for j := range stmts {
			s := stmts[(first+j)%len(stmts)]
			e, ok := makeEdit(srcs[s.Case], s, editKinds[i%len(editKinds)], rng)
			if ok && !seen[e] {
				seen[e] = true
				out = append(out, e)
				break
			}
		}
	}
	return out
}

// makeEdit builds one edit of kind to s. A token edit changes one variable,
// integer or single-quoted string in the statement, picked by rng.
func makeEdit(src []byte, s caseStmt, kind string, rng *rand.Rand) (convergenceEdit, bool) {
	n := s.Node
	e := convergenceEdit{Case: s.Case, Kind: kind, StmtStart: n.Start, StmtEnd: n.End}
	switch kind {
	case "insert":
		e.Offset, e.Inserted = n.End, " "+insertedStatement
	case "delete":
		e.Offset, e.Removed = n.Start, n.End-n.Start
	case "replace":
		if string(src[n.Start:n.End]) == insertedStatement {
			return e, false
		}
		e.Offset, e.Removed, e.Inserted = n.Start, n.End-n.Start, insertedStatement
	case "token":
		var terms []*parse.Node
		var walk func(m *parse.Node)
		walk = func(m *parse.Node) {
			if m.Kind == parse.Term && tokenReplacement(string(src[m.Start:m.End])) != "" {
				terms = append(terms, m)
			}
			for _, c := range m.Children {
				walk(c)
			}
		}
		walk(n)
		if len(terms) == 0 {
			return e, false
		}
		m := terms[rng.IntN(len(terms))]
		e.Offset, e.Removed = m.Start, m.End-m.Start
		e.Inserted = tokenReplacement(string(src[m.Start:m.End]))
	}
	return e, true
}

var (
	scalarToken  = regexp.MustCompile(`^\$[A-Za-z_]\w*$`)
	integerToken = regexp.MustCompile(`^[0-9]+$`)
	stringToken  = regexp.MustCompile(`^'[^'\\]*'$`)
)

// tokenReplacement is the text a token edit puts in place of tok, or "" when
// tok is not a token the edit changes.
func tokenReplacement(tok string) string {
	switch {
	case scalarToken.MatchString(tok):
		return tok + "e"
	case integerToken.MatchString(tok) && tok != "42":
		return "42"
	case stringToken.MatchString(tok) && tok != "'edited'":
		return "'edited'"
	}
	return ""
}

// writeConvergenceEdits writes edits as a list, one JSON object per line.
func writeConvergenceEdits(path string, edits []convergenceEdit) error {
	var b bytes.Buffer
	for _, e := range edits {
		line, err := json.Marshal(e)
		if err != nil {
			return err
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

// std5kEditableStatements are std5k.pl's editable statements, as caseStmts.
func std5kEditableStatements(src []byte) []caseStmt {
	var out []caseStmt
	for _, n := range editableStatements(parse.Parse(src), src, 0) {
		out = append(out, caseStmt{Node: n})
	}
	return out
}

// TestConvergenceEditsMatchDraw: the committed list is what drawEdits draws
// from std5k.pl, so anyone can repeat the draw and get the same 200 edits.
func TestConvergenceEditsMatchDraw(t *testing.T) {
	src, err := os.ReadFile(std5kPath)
	if err != nil {
		t.Fatal(err)
	}
	want := drawEdits(std5kEditableStatements(src), map[string][]byte{"": src}, convergenceEditCount)
	if *updateConvergenceEdits {
		if err := writeConvergenceEdits(convergenceEditsPath, want); err != nil {
			t.Fatal(err)
		}
	}
	got, err := readConvergenceEdits(convergenceEditsPath, false)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("%s differs from drawEdits's draw over %s", convergenceEditsPath, std5kPath)
	}
}

// TestConvergenceEditsHold: std5k.pl's list holds exactly 200 edits, none
// twice, each inside one statement and leaving the file accepted.
func TestConvergenceEditsHold(t *testing.T) {
	src, err := os.ReadFile(std5kPath)
	if err != nil {
		t.Fatal(err)
	}
	edits, err := readConvergenceEdits(convergenceEditsPath, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(edits) != convergenceEditCount {
		t.Errorf("%s holds %d edits, want %d", convergenceEditsPath, len(edits), convergenceEditCount)
	}
	stmts := spanSet(editableStatements(parse.Parse(src), src, 0))
	checkEdits(t, convergenceEditsPath, edits,
		map[string][]byte{"": src}, map[string]map[stmtSpan]bool{"": stmts})
}

// The Guacamole corpus's parses cases and the preamble every case starts
// with.
var (
	guacamoleParsesDir    = filepath.Join("testdata", "guacamole", "parses")
	guacamolePreamblePath = filepath.Join("testdata", "guacamole", "preamble.pl")
	convergenceCorpusPath = filepath.Join("testdata", "convergence-edits-guacamole.jsonl")
)

// multiStatementCasesRecorded are the multi-statement cases, in file-name
// order, as recorded in the milestone.
var multiStatementCasesRecorded = []string{
	"Statements-Block-007.pl",
	"Statements-Ellipsis-005.pl",
	"bugs-rt_132920_regex_subset-001.pl",
}

// multiStatementCases returns the perl-accepted parses cases, by bare file
// name, that hold more than one editable statement after the preamble (the
// draw's own unit, so the count and the draw share one rule), with each
// case's bytes and tree.
func multiStatementCases(t *testing.T) (names []string, srcs map[string][]byte, roots map[string]*parse.Node, preambleLen int) {
	t.Helper()
	preamble, err := os.ReadFile(guacamolePreamblePath)
	if err != nil {
		t.Fatal(err)
	}
	labels, err := ReadLabels(filepath.Join(labelsDir, "in-tree.txt"))
	if err != nil {
		t.Fatal(err)
	}
	rejected := map[string]bool{}
	for _, l := range labels.Labels {
		rejected[l.Path] = l.Refused()
	}
	files, err := filepath.Glob(filepath.Join(guacamoleParsesDir, "*.pl"))
	if err != nil {
		t.Fatal(err)
	}
	srcs, roots = map[string][]byte{}, map[string]*parse.Node{}
	for _, f := range files {
		name := filepath.Base(f)
		if rejected[filepath.ToSlash(filepath.Join("internal", "chart", f))] {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		root := parse.Parse(src)
		if len(editableStatements(root, src, len(preamble))) > 1 {
			names = append(names, name)
			srcs[name], roots[name] = src, root
		}
	}
	return names, srcs, roots, len(preamble)
}

// TestConvergenceCorpusEditsWhereNeeded: the Guacamole corpus has a fixed
// edit list exactly when it has multi-statement cases. The list is drawn
// as std5k.pl's is, over those cases' editable statements, up to 200 edits,
// and is held to the same guard.
func TestConvergenceCorpusEditsWhereNeeded(t *testing.T) {
	names, srcs, roots, preambleLen := multiStatementCases(t)
	t.Logf("%d perl-accepted parses case(s) hold more than one editable statement: %v", len(names), names)
	// The count recorded under chart-waypoint's ## Records; a corpus or
	// parser change that moves it must be recorded there, not absorbed.
	if !slices.Equal(names, multiStatementCasesRecorded) {
		t.Errorf("multi-statement parses cases are %d %v, want the %d recorded %v",
			len(names), names, len(multiStatementCasesRecorded), multiStatementCasesRecorded)
	}
	if len(names) == 0 {
		if _, err := os.Stat(convergenceCorpusPath); !os.IsNotExist(err) {
			t.Fatalf("no parses case holds more than one editable statement, so %s should not exist (stat: %v)", convergenceCorpusPath, err)
		}
		return
	}
	var stmts []caseStmt
	spans := map[string]map[stmtSpan]bool{}
	for _, name := range names {
		nodes := editableStatements(roots[name], srcs[name], preambleLen)
		spans[name] = spanSet(nodes)
		for _, n := range nodes {
			stmts = append(stmts, caseStmt{Case: name, Node: n})
		}
	}
	want := drawEdits(stmts, srcs, min(convergenceEditCount, len(editKinds)*len(stmts)))
	if *updateConvergenceEdits {
		if err := writeConvergenceEdits(convergenceCorpusPath, want); err != nil {
			t.Fatal(err)
		}
	}
	got, err := readConvergenceEdits(convergenceCorpusPath, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 || len(got) > convergenceEditCount {
		t.Errorf("%s holds %d edits, want 1 to %d", convergenceCorpusPath, len(got), convergenceEditCount)
	}
	if !slices.Equal(got, want) {
		t.Errorf("%s differs from drawEdits's draw over the multi-statement cases", convergenceCorpusPath)
	}
	checkEdits(t, convergenceCorpusPath, got, srcs, spans)
}

// TestConvergenceEditsRejectOutOfRange: checkEdit refuses hand-built edits
// that run past the end of std5k.pl or cross from one statement into the
// next, and admits the in-statement edit beside them.
func TestConvergenceEditsRejectOutOfRange(t *testing.T) {
	src, err := os.ReadFile(std5kPath)
	if err != nil {
		t.Fatal(err)
	}
	nodes := editableStatements(parse.Parse(src), src, 0)
	stmts := spanSet(nodes)
	a, b := nodes[0], nodes[1]
	cases := []struct {
		name string
		edit convergenceEdit
		want string
	}{
		{"past the end", convergenceEdit{Kind: "delete", Offset: len(src) + 1, Removed: 1, StmtStart: a.Start, StmtEnd: a.End}, "outside the file"},
		{"ends past the end", convergenceEdit{Kind: "delete", Offset: len(src) - 1, Removed: 2, StmtStart: a.Start, StmtEnd: a.End}, "outside the file"},
		{"crosses into the next", convergenceEdit{Kind: "delete", Offset: a.End - 1, Removed: b.Start - a.End + 2, StmtStart: a.Start, StmtEnd: a.End}, "cross out of statement"},
		{"names two statements", convergenceEdit{Kind: "delete", Offset: a.Start, Removed: b.End - a.Start, StmtStart: a.Start, StmtEnd: b.End}, "not one of recursive descent's statements"},
	}
	for _, c := range cases {
		err := checkEdit(src, stmts, c.edit)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: checkEdit = %v, want an error saying %q", c.name, err, c.want)
		}
	}
	inside := convergenceEdit{Kind: "delete", Offset: a.Start, Removed: a.End - a.Start, StmtStart: a.Start, StmtEnd: a.End}
	if err := checkEdit(src, stmts, inside); err != nil {
		t.Errorf("an edit inside one statement: checkEdit = %v, want nil", err)
	}
}
