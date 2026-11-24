// ABOUTME: Tests that an edit's lowering, ids, position map and syntax errors cost what the edit damaged, not the file.
// ABOUTME: Counts each phase's nodes for 1-char edits to edit5k.pl, times the position update, and checks every result against a fresh open.

package psc

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// cstNodes is the number of nodes in the tree under n, n included.
func cstNodes(n *parse.Node) int {
	k := 1
	for _, c := range n.Children {
		k += cstNodes(c)
	}
	return k
}

// loweredNodes is every node of the lowered tree under n, n included.
func loweredNodes(n lower.Node) []lower.Node {
	out := []lower.Node{n}
	for _, c := range lower.Children(n) {
		out = append(out, loweredNodes(c)...)
	}
	return out
}

// editWork is what one incremental edit cost, phase by phase, in nodes.
type editWork struct {
	lower, assignIDs, positions, analysis, diagnostics int
	// topLevel is the lowered nodes outside every sub, which analysis
	// walks on every edit (Reanalyze); reported, not bounded.
	topLevel int
	// repaired is whether Repair built the edit's tree.
	repaired bool
}

// repairedEdit is the incremental path over one edit, src to edited, as
// the owner runs it with parse.IncrementalRepair on, from root, src's
// tree, and prev, its analysis: the tree Repair builds and the damage
// read off its region, or, where Repair refuses, a full parse and a tree
// diff. It fails t unless the edit was analysed locally, and returns the
// edit's tree and analysis and what each phase cost.
func repairedEdit(t *testing.T, root *parse.Node, prev *infer.FileAnalysis, src, edited []byte) (*parse.Node, *infer.FileAnalysis, editWork) {
	t.Helper()
	var w editWork
	next, ok := parse.Repair(root, fullSyncEdit(src, edited), edited)
	if w.repaired = ok; !ok {
		next = parse.Parse(edited)
	}
	fa, d, err := analyzeEdit(context.Background(), root, prev, next, next.Damage(), infer.Options{})
	if err != nil || d.Full {
		t.Fatalf("damage %+v, err %v; want a local edit", d, err)
	}
	old := map[lower.Node]bool{}
	for _, n := range loweredNodes(prev.File) {
		old[n] = true
	}
	for _, n := range loweredNodes(fa.File) {
		if !old[n] {
			w.lower++
		}
	}
	w.assignIDs, w.positions = fa.Positions.Work()
	decls := topDecls(fa.File)
	for _, a := range fa.Walked {
		w.analysis += len(loweredNodes(decls[a]))
	}
	w.topLevel = len(loweredNodes(fa.File))
	for _, d := range decls {
		w.topLevel -= len(loweredNodes(d))
	}
	// resyntax reads only the statements the map numbered again.
	for _, i := range fa.Positions.Renumbered() {
		w.diagnostics += cstNodes(next.Children[i])
	}
	return next, fa, w
}

// checkLocal fails t for each phase of w over 5% of a file of n nodes.
func checkLocal(t *testing.T, w editWork, n int) {
	t.Helper()
	if !w.repaired {
		t.Error("Repair refused the edit")
	}
	for _, p := range []struct {
		name  string
		nodes int
	}{{"lower", w.lower}, {"AssignIDs", w.assignIDs}, {"positions", w.positions}, {"analysis", w.analysis}, {"diagnostics", w.diagnostics}} {
		if p.nodes == 0 || p.nodes*20 >= n {
			t.Errorf("%s visits %d of the file's %d nodes, want more than none and under 5%%", p.name, p.nodes, n)
		}
	}
	t.Logf("of %d nodes: %+v", n, w)
}

// TestIncrementalEditWorkIsLocal: the benchmark's 1-char edit, inside the
// middle `sub scale` of edit5k.pl, re-lowers, renumbers, re-indexes,
// re-analyzes and re-reads for syntax errors under 5% of the file's
// nodes, and equals a fresh open.
func TestIncrementalEditWorkIsLocal(t *testing.T) {
	src := readFixture(t)
	at := editOffset(t, src)
	edited := slices.Concat(src[:at], []byte("1"), src[at:])
	root := parse.Parse(src)
	_, fa, w := repairedEdit(t, root, fresh(t, src), src, edited)
	checkLocal(t, w, cstNodes(root))
	sameAnalysis(t, fa, fresh(t, edited))
}

// TestIncrementalEditWorkIsLocalAtFileEdges: a 1-char edit in the first
// sub of edit5k.pl that a 1-char edit leaves local, and one in its last
// sub, visit under 5% of the file's nodes too, so no phase grows with the
// text after the edit, or before it. An edit to either of the file's
// first two subs re-analyzes the whole file -- `new` assigns its
// %REGISTRY, and top-level code reaches `limit` -- so the first is
// `describe`, the third.
func TestIncrementalEditWorkIsLocalAtFileEdges(t *testing.T) {
	src := readFixture(t)
	first := bytes.Index(src, []byte("my ($self) = @_;\n    my $text")) + len("my ")
	last := bytes.LastIndex(src, []byte("my $n = 0;")) + len("my $n = ")
	for _, c := range []struct {
		name, insert string
		at           int
	}{{"first", " ", first}, {"last", "1", last}} {
		t.Run(c.name, func(t *testing.T) {
			edited := slices.Concat(src[:c.at], []byte(c.insert), src[c.at:])
			root := parse.Parse(src)
			_, fa, w := repairedEdit(t, root, fresh(t, src), src, edited)
			checkLocal(t, w, cstNodes(root))
			sameAnalysis(t, fa, fresh(t, edited))
		})
	}
}

// TestIncrementalPositionsP99WithinBudget: over 200 edits of latencyEdits'
// typing sequence on edit5k.pl, the position update the incremental path
// makes -- lower.Relower, which renumbers and re-indexes the statements an
// edit changed beside re-lowering its sub -- takes under 1 ms at p99, the
// sub-budget of the 01a17969-45ba ruling. It times, so it runs only with
// PSC_LATENCY_BUDGET=1, on the reference 4-vCPU VM.
func TestIncrementalPositionsP99WithinBudget(t *testing.T) {
	if os.Getenv("PSC_LATENCY_BUDGET") != "1" {
		t.Skip("a timing budget; set PSC_LATENCY_BUDGET=1 to run it")
	}
	src := readFixture(t)
	root := parse.Parse(src)
	f := lower.Lower(root)
	pos := lower.NewPositions(root, f)
	var times []time.Duration
	for i, buf := range latencyEdits(src, 200) {
		next, ok := parse.Repair(root, fullSyncEdit(src, buf), buf)
		if !ok {
			t.Fatalf("edit %d: Repair refused a typing edit", i)
		}
		d := damageFromRepair(root, next, *next.Damage())
		start := time.Now()
		f, pos, ok = lower.Relower(f, pos, next, d.Subs)
		times = append(times, time.Since(start))
		if !ok {
			t.Fatalf("edit %d: Relower refused damage %+v", i, d)
		}
		root, src = next, buf
	}
	p50, p99 := percentiles(times)
	t.Logf("position update over %d edits: p50 %v, p99 %v", len(times), p50, p99)
	if p99 >= time.Millisecond {
		t.Errorf("the position update's p99 is %v, want under 1ms", p99)
	}
}

// samePositions fails t unless got, the map of root an edit made, answers
// as want, a fresh NewPositions of root, does: the span and quote of every
// lowered node, the node at every offset, and the last Match before each
// node's start and end.
func samePositions(t *testing.T, root *parse.Node, got, want *infer.FileAnalysis) {
	t.Helper()
	bad := 0
	fail := func(format string, args ...any) {
		if bad++; bad <= 10 {
			t.Errorf(format, args...)
		}
	}
	gp, wp := got.Positions, want.Positions
	isMatch := func(n lower.Node) bool { _, ok := n.(*lower.Match); return ok }
	id := func(n lower.Node) any {
		if n == nil {
			return nil
		}
		return n.Meta().ID
	}
	for _, n := range loweredNodes(want.File) {
		x := n.Meta().ID
		gs, ge, gok := gp.Span(x)
		ws, we, wok := wp.Span(x)
		if gs != ws || ge != we || gok != wok {
			fail("Span(%v) = %d, %d, %v; want %d, %d, %v", x, gs, ge, gok, ws, we, wok)
		}
		if g, w := gp.Quote(x), wp.Quote(x); g != w {
			fail("Quote(%v) = %q, want %q", x, g, w)
		}
		for _, off := range []int{ws, we} {
			if g, w := id(gp.LastBefore(off, isMatch)), id(wp.LastBefore(off, isMatch)); g != w {
				fail("LastBefore(%d) = %v, want %v", off, g, w)
			}
		}
	}
	for off := -1; off <= root.End+1; off++ {
		if g, w := id(gp.At(off)), id(wp.At(off)); g != w {
			fail("At(%d) = %v, want %v", off, g, w)
		}
	}
}

// TestIncrementalPositionsMatchFresh: after edits inside edit5k.pl's subs
// that add a line, remove one, and add multi-byte characters, each made
// from the last through the incremental path, every node's position
// answers as a fresh lower.NewPositions of the same tree does.
func TestIncrementalPositionsMatchFresh(t *testing.T) {
	src := readFixture(t)
	root := parse.Parse(src)
	prev := fresh(t, src)
	for _, e := range [][2]string{
		{"my $n = 0;\n", "my $n = 0;\n\n"},
		{"    return 'empty' unless length $word;\n", "    return 'empty' unless length $word;"},
		{"return 'empty'", "return 'émpty→😀'"},
		{"    my @kept", "    #é\n    my @kept"},
		{"return $n;\n}", "return $n;}"},
	} {
		t.Run(fmt.Sprintf("%q", e[1]), func(t *testing.T) {
			at := bytes.LastIndex(src, []byte(e[0]))
			if at < 0 {
				t.Fatalf("%q is not in the source", e[0])
			}
			edited := slices.Concat(src[:at], []byte(e[1]), src[at+len(e[0]):])
			var fa *infer.FileAnalysis
			root, fa, _ = repairedEdit(t, root, prev, src, edited)
			if fa.Positions.Renumbered() == nil {
				t.Fatal("the edit's map is a fresh one, not one Relower made")
			}
			samePositions(t, root, fa, fresh(t, edited))
			prev, src = fa, edited
		})
	}
}

// TestIncrementalIDsUniqueAfterEditSequence: after 100 edits of
// latencyEdits' sequence, then one adding a line and one removing it, each
// made from the last, no two lowered nodes share an id, every annotation
// is of a node of the lowered tree, and the analysis is a fresh open's.
func TestIncrementalIDsUniqueAfterEditSequence(t *testing.T) {
	src := readFixture(t)
	root := parse.Parse(src)
	prev := fresh(t, src)
	bufs := latencyEdits(src, 100)
	at := bytes.Index(bufs[99], []byte("    my @kept"))
	line := slices.Concat(bufs[99][:at], []byte("\n"), bufs[99][at:])
	for _, buf := range append(bufs, line, bufs[99]) {
		root, prev, _ = repairedEdit(t, root, prev, src, buf)
		src = buf
	}
	ids := map[lower.ID]lower.Node{}
	for _, n := range loweredNodes(prev.File) {
		if o, dup := ids[n.Meta().ID]; dup {
			t.Errorf("%T and %T share id %v", o, n, n.Meta().ID)
		}
		ids[n.Meta().ID] = n
	}
	for id := range prev.Annotations {
		if ids[id] == nil {
			t.Errorf("annotation %v is of no node of the lowered tree", id)
		}
	}
	sameAnalysis(t, prev, fresh(t, src))
}

// TestIncrementalDiagnosticRangesMatchFresh: after an edit that adds a
// line inside one sub, through the LSP with incremental sync and
// parse.IncrementalRepair on, the published ranges of the diagnostics
// below it -- a type diagnostic in a later sub, and the syntax error of
// one after that, reused rather than found again -- are a fresh open's.
// Removing the line moves them back.
func TestIncrementalDiagnosticRangesMatchFresh(t *testing.T) {
	const uri = "untitled:ranges.pl"
	text := strings.Join([]string{
		"sub a {",
		"    my $x = 1;",
		"    return $x;",
		"}",
		"sub b {",
		`    my $s = "héllo";`,
		"    my $u = undef;",
		"    my $n = $u + 1;",
		"    return $n;",
		"}",
		"sub c {",
		"    my $y = 2 * ;",
		"    return $y;",
		"}",
		"",
	}, "\n")
	h, out, _, _ := syncHandler(t, true)
	didOpen(t, h, uri, text)
	before := published(t, out, uri)
	if len(before) < 2 {
		t.Fatalf("published %+v, want a type and a syntax diagnostic", before)
	}
	for v, e := range [][2]string{{"my $x = 1;", "my $x = 1;\n"}, {"my $x = 1;\n", "my $x = 1;"}} {
		var c lspTextDocumentContentChangeEvent
		c, text = replace(t, text, e[0], e[1])
		didChange(t, h, uri, v+2, c)
		if got := h.server.snapshot(uri).analysis; got.Walked == nil {
			t.Fatalf("edit %d was analysed in full, want incrementally", v)
		}
		matchesFreshOpen(t, h, out, uri, text)
	}
}
