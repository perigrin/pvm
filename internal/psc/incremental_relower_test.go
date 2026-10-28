// ABOUTME: Tests analyzeEdit: re-lowering and re-analyzing only an edit's damaged subs equals a fresh open of the same bytes.
// ABOUTME: Lowered tree, ids, annotations, diagnostics and symbols are compared after every edit, through the owner and on real files.

package psc

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// fresh is analyze's analysis of src, parsed afresh.
func fresh(t *testing.T, src []byte) *infer.FileAnalysis {
	t.Helper()
	fa, err := analyze(context.Background(), parse.Parse(src), infer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	return fa
}

// loweredShape is f as Sexpr prints it, then every node's id in preorder:
// two trees equal under it have the same kinds, fields, children and ids.
func loweredShape(f *lower.File) string {
	var b strings.Builder
	b.WriteString(lower.Sexpr(f))
	var walk func(n lower.Node)
	walk = func(n lower.Node) {
		fmt.Fprintf(&b, " %v", n.Meta().ID)
		for _, c := range lower.Children(n) {
			walk(c)
		}
	}
	walk(f)
	return b.String()
}

// symbolView is what a reader of fa's symbol table can see: every symbol,
// what LookupAll finds for each name, what Lookup finds from the file
// scope, and the package.
func symbolView(fa *infer.FileAnalysis) []string {
	out := []string{"package " + fa.Package}
	seen := map[string]bool{}
	for _, s := range fa.Symbols.AllSymbols() {
		out = append(out, fmt.Sprintf("symbol %+v", s))
		if !seen[s.Name] {
			seen[s.Name] = true
			all, _ := fa.Symbols.LookupAll(s.Name)
			root, _ := fa.Symbols.Lookup(s.Name)
			out = append(out, fmt.Sprintf("all %+v", all), fmt.Sprintf("root %+v", root))
		}
	}
	slices.Sort(out)
	return out
}

// sameAnalysis fails t where got differs from want, a fresh analysis of
// the same bytes.
func sameAnalysis(t *testing.T, got, want *infer.FileAnalysis) {
	t.Helper()
	if g, w := loweredShape(got.File), loweredShape(want.File); g != w {
		i := 0
		for i < len(g) && i < len(w) && g[i] == w[i] {
			i++
		}
		t.Errorf("lowered tree differs at byte %d: %.120q, want %.120q", i, g[i:], w[i:])
	}
	if !maps.Equal(got.Annotations, want.Annotations) {
		n := 0
		for id, typ := range want.Annotations {
			if g, ok := got.Annotations[id]; (!ok || g != typ) && n < 5 {
				t.Errorf("annotation %v = %v (present %v), want %v", id, g, ok, typ)
				n++
			}
		}
		for id, typ := range got.Annotations {
			if _, ok := want.Annotations[id]; !ok && n < 10 {
				t.Errorf("annotation %v = %v, want none", id, typ)
				n++
			}
		}
	}
	if !slices.Equal(got.Diagnostics, want.Diagnostics) {
		t.Errorf("diagnostics:\n%+v\nwant\n%+v", got.Diagnostics, want.Diagnostics)
	}
	if g, w := symbolView(got), symbolView(want); !slices.Equal(g, w) {
		for _, s := range g {
			if !slices.Contains(w, s) {
				t.Errorf("symbols: extra %s", s)
			}
		}
		for _, s := range w {
			if !slices.Contains(g, s) {
				t.Errorf("symbols: missing %s", s)
			}
		}
	}
}

// liveDoc is one document open in an LSPServer with no debounce, edited
// a buffer at a time, each analysed by its owner from the last.
type liveDoc struct {
	s       *LSPServer
	version int
}

const liveURI = "untitled:edit.pl"

func openLive(t *testing.T, src []byte) *liveDoc {
	t.Helper()
	return openLiveWith(t, src, NewLSPServer())
}

// openLiveWith is openLive in s, which an owner takes its compute from.
func openLiveWith(t *testing.T, src []byte, s *LSPServer) *liveDoc {
	t.Helper()
	s.parseDelay, s.diagDelay = 0, 0
	t.Cleanup(s.Shutdown)
	if err := s.OpenDocument(liveURI, src); err != nil {
		t.Fatal(err)
	}
	return &liveDoc{s: s}
}

// edit hands the owner buf and returns its analysis once it lands.
func (d *liveDoc) edit(t *testing.T, buf []byte) *infer.FileAnalysis {
	t.Helper()
	d.version++
	d.s.change(liveURI, d.version, buf)
	doc := d.s.snapshot(liveURI)
	if doc == nil || !bytes.Equal(doc.text.Text, buf) {
		t.Fatalf("the owner did not land version %d", d.version)
	}
	return doc.analysis
}

// The §7.7.3 rows (chapter 7), as internal/parse's seed table holds them:
// each a source and the edit applied to it.
var seventhRows = []struct{ name, src, old, new string }{
	{"insert-quote", seedSubSrc, "= 1", `= "1`},
	{"insert-heredoc", seedSubSrc, "1;\n", "1;\n  print <<EOF;\n"},
	{"delete-heredoc-terminator", heredocSubSrc, "EOF\n", ""},
	{"pod", seedSubSrc, "1;\n", "1;\n=pod\n"},
	{"end", seedSubSrc, "1;\n", "1;\n__END__\n"},
	{"prototype", protoSubSrc, "sub f(", `sub f(\@`},
	{"insert-brace", seedSubSrc, "1;\n", "1;\n{\n"},
	{"split-token", seedSubSrc, "my $foo", "my $f|oo"},
}

const (
	seedSubSrc    = "sub f {\n  my $foo = 1;\n  return $foo;\n}\nprint f();\n"
	heredocSubSrc = "sub f {\n  print <<EOF;\nbody\nEOF\n  return 1;\n}\nprint f();\n"
	protoSubSrc   = "sub f() { 1 }\nmy @a = (1, 2);\nprint f @a;\n"
)

// TestIncrementalAnalyzeMatchesFresh: a sequence of edits through the
// owner -- each §7.7.3 row and back, then a typing sequence of 1-char
// edits inside edit5k.pl's subs -- leaves, after every edit, the lowered
// tree, annotations, diagnostics and symbols a fresh open of the same
// bytes has; and the edit5k.pl edits are local, not full re-analyses.
func TestIncrementalAnalyzeMatchesFresh(t *testing.T) {
	for _, r := range seventhRows {
		t.Run(r.name, func(t *testing.T) {
			src := []byte(r.src)
			edited := []byte(strings.Replace(r.src, r.old, r.new, 1))
			if bytes.Equal(src, edited) {
				t.Fatal("the edit changes nothing")
			}
			d := openLive(t, src)
			for _, buf := range [][]byte{edited, src, edited} {
				sameAnalysis(t, d.edit(t, buf), fresh(t, buf))
			}
		})
	}
	t.Run("edit5k.pl", func(t *testing.T) {
		src := readFixture(t)
		d := openLive(t, src)
		local := 0
		for _, buf := range latencyEdits(src, 30) {
			got := d.edit(t, buf)
			sameAnalysis(t, got, fresh(t, buf))
			if len(got.Walked) == 1 {
				local++
			}
		}
		if local != 30 {
			t.Errorf("%d of 30 edits re-analyzed one sub, want all", local)
		}
	})
}

// topDecls is each top-level sub's lowered Decl in f, by anchor.
func topDecls(f *lower.File) map[string]lower.Node {
	out := map[string]lower.Node{}
	var walk func(n lower.Node)
	walk = func(n lower.Node) {
		if id := n.Meta().ID; id.Anchor != "" {
			if id.Local == 0 && id.Role == 0 && !strings.Contains(id.Anchor, "/") {
				out[id.Anchor] = n
			}
			return
		}
		for _, c := range lower.Children(n) {
			walk(c)
		}
	}
	walk(f)
	return out
}

// TestIncrementalAnalyzeRelowersOnlyDamagedSubs: the benchmark's 1-char
// edit, inside the middle `sub scale`, re-lowers that sub alone -- every
// other top-level sub's lowered Decl is the old tree's own node -- and
// walks that sub alone, replaying every other, and equals a fresh open.
func TestIncrementalAnalyzeRelowersOnlyDamagedSubs(t *testing.T) {
	src := readFixture(t)
	at := editOffset(t, src)
	edited := slices.Concat(src[:at], []byte("1"), src[at:])
	prevRoot, root := parse.Parse(src), parse.Parse(edited)
	prev, err := analyze(context.Background(), prevRoot, infer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	got, d, err := analyzeEdit(context.Background(), prevRoot, src, prev, root, edited, infer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("sub scale#%d", len(scaleSubs(src))/2)
	if d.Full || !slices.Equal(d.Subs, []string{want}) {
		t.Fatalf("damage = %+v, want exactly [%s]", d, want)
	}
	if !slices.Equal(got.Walked, []string{want}) {
		t.Errorf("walked %v, want only [%s]", got.Walked, want)
	}
	before, after := topDecls(prev.File), topDecls(got.File)
	if len(after) != len(before) || len(after) < 400 {
		t.Fatalf("%d top-level subs became %d", len(before), len(after))
	}
	for anchor, n := range after {
		if shared := n == before[anchor]; shared == (anchor == want) {
			t.Errorf("%s: shared with the old tree = %v", anchor, shared)
		}
	}
	sameAnalysis(t, got, fresh(t, edited))
}

// TestIncrementalAnalyzeClassBFallsBackToFull: a Class B edit -- a sub
// renamed, a prototype added, a use changed -- takes analyze's full path,
// walking every sub, and equals a fresh open.
func TestIncrementalAnalyzeClassBFallsBackToFull(t *testing.T) {
	src := readFixture(t)
	for _, c := range []struct{ name, anchor, old, new string }{
		{"rename a sub", "sub scale {", "scale", "scaled"},
		{"add a prototype", "sub limit {", "limit {", "limit($) {"},
		{"change a use", "use warnings;", "warnings", "warnings FATAL => 'all'"},
	} {
		t.Run(c.name, func(t *testing.T) {
			edited, _ := replaceOnce(t, src, c.anchor, c.old, c.new)
			prevRoot := parse.Parse(src)
			prev := fresh(t, src)
			got, d, err := analyzeEdit(context.Background(), prevRoot, src, prev, parse.Parse(edited), edited, infer.Options{})
			if err != nil {
				t.Fatal(err)
			}
			if !d.Full || !d.ClassB || got.Walked != nil {
				t.Errorf("damage %+v, walked %v; want Full, ClassB and the full path", d, got.Walked)
			}
			sameAnalysis(t, got, fresh(t, edited))
		})
	}
}

// editAndCompare runs analyzeEdit from src's fresh analysis to edited's
// and compares it with a fresh one; it returns both, and the damage.
func editAndCompare(t *testing.T, src, edited []byte) (got, want *infer.FileAnalysis, d editDamage) {
	t.Helper()
	got, d, err := analyzeEdit(context.Background(), parse.Parse(src), src, fresh(t, src), parse.Parse(edited), edited, infer.Options{})
	if err != nil {
		t.Fatal(err)
	}
	want = fresh(t, edited)
	sameAnalysis(t, got, want)
	return got, want, d
}

// TestIncrementalAnalyzeCrossSubDependency: an edit to a that changes what
// its caller b infers damages b too, and b's annotations and diagnostics
// equal a fresh open's.
func TestIncrementalAnalyzeCrossSubDependency(t *testing.T) {
	src := []byte("sub a { return 1 }\nsub b { my $y = a(); return $y + 1 }\nsub c { return 3 }\n")
	edited, _ := replaceOnce(t, src, "sub a", "return 1", "return 'x'")
	before, after := fresh(t, src), fresh(t, edited)
	inB := func(fa *infer.FileAnalysis) string {
		var out []string
		for id, typ := range fa.Annotations {
			if id.Anchor == "sub b#0" {
				out = append(out, fmt.Sprintf("%v=%v", id, typ))
			}
		}
		slices.Sort(out)
		return strings.Join(out, " ")
	}
	if inB(before) == inB(after) && slices.Equal(before.Diagnostics, after.Diagnostics) {
		t.Fatal("the edit to a changes nothing b infers; the case no longer shows the dependency")
	}
	got, _, d := editAndCompare(t, src, edited)
	if d.Full || !slices.Equal(got.Walked, []string{"sub a#0", "sub b#0"}) {
		t.Errorf("damage %+v, walked %v; want a and b walked, c replayed", d, got.Walked)
	}
}

// TestIncrementalAnalyzeDropsDeletedNodeAnnotations: a statement deleted
// from a sub takes its annotations and its diagnostic with it; none is
// carried forward under a lower.ID the new tree does not have. The
// undamaged sub s keeps its own diagnostic, replayed and moved.
func TestIncrementalAnalyzeDropsDeletedNodeAnnotations(t *testing.T) {
	src := []byte("sub r {\n  my $p = undef;\n  my $q = $p eq 'n';\n  return 1;\n}\nsub s { my $u = undef; return $u eq 'n' }\n")
	edited, _ := replaceOnce(t, src, "sub r", "  my $q = $p eq 'n';\n", "")
	before := fresh(t, src)
	if len(before.Diagnostics) != 2 {
		t.Fatalf("r and s raise %d diagnostics, want 2", len(before.Diagnostics))
	}
	got, want, d := editAndCompare(t, src, edited)
	if d.Full || !slices.Equal(got.Walked, []string{"sub r#0"}) {
		t.Fatalf("damage %+v, walked %v; want r alone", d, got.Walked)
	}
	stale := 0
	for id := range before.Annotations {
		if _, ok := want.Annotations[id]; !ok {
			stale++
			if _, ok := got.Annotations[id]; ok {
				t.Errorf("%v is annotated, but the new tree has no such node", id)
			}
		}
	}
	if stale == 0 {
		t.Fatal("no old id is gone from the new tree; the case deletes nothing")
	}
	if len(got.Diagnostics) != 1 || got.Diagnostics[0].StartByte <= before.Diagnostics[0].StartByte {
		t.Errorf("diagnostics %+v, want s's alone", got.Diagnostics)
	}
}

// TestIncrementalAnalyzeErrorThenFix: breaking a sub's syntax, editing
// another sub while the file stays broken, then fixing it, leaves after
// each edit the annotations and gated diagnostics of a fresh open; the
// middle edit is local.
func TestIncrementalAnalyzeErrorThenFix(t *testing.T) {
	src := readFixture(t)
	broken, _ := replaceOnce(t, src, "sub scale {", "grep {", "grep ( {")
	if !parse.Parse(broken).HasError {
		t.Fatal("the break parses without an error")
	}
	brokenEdited, _ := replaceOnce(t, broken, "package Fixture::Block20;", "$_ > 0", "$_ > 10")
	fixed, _ := replaceOnce(t, brokenEdited, "sub scale {", "grep ( {", "grep {")
	d := openLive(t, src)
	for i, buf := range [][]byte{broken, brokenEdited, fixed} {
		got, want := d.edit(t, buf), fresh(t, buf)
		sameAnalysis(t, got, want)
		if i == 1 && !slices.Equal(got.Walked, []string{"sub scale#20"}) {
			t.Errorf("the edit in a broken file walked %v, want [sub scale#20]", got.Walked)
		}
		if i < 2 && len(want.Diagnostics) == 0 {
			t.Errorf("edit %d: the broken file has no diagnostic", i)
		}
	}
}

// cancelAt is a context whose Err is nil for its first n calls and
// Canceled from then on.
type cancelAt struct {
	context.Context
	mu sync.Mutex
	n  int
}

func (c *cancelAt) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.n--; c.n < 0 {
		return context.Canceled
	}
	return nil
}

// TestIncrementalAnalyzeCancelledLeavesConsistentState: an analysis that
// a newer edit cancels after re-lowering, before re-analyzing, returns the
// error and changes nothing of the analysis it started from -- through the
// owner, which keeps serving its last complete version -- and that version
// still re-analyzes the next edit as a fresh open does.
func TestIncrementalAnalyzeCancelledLeavesConsistentState(t *testing.T) {
	src := readFixture(t)
	bufs := latencyEdits(src, 2)
	prevRoot := parse.Parse(src)
	prev := fresh(t, src)
	shape, ann, diags, syms := loweredShape(prev.File), maps.Clone(prev.Annotations), slices.Clone(prev.Diagnostics), symbolView(prev)
	// analyzeEdit's first poll is after re-lowering, before re-analyzing.
	got, d, err := analyzeEdit(&cancelAt{Context: context.Background()}, prevRoot, src, prev, parse.Parse(bufs[0]), bufs[0], infer.Options{})
	if err == nil || got != nil || d.Full {
		t.Fatalf("analyzeEdit = %v, %+v, %v; want a local edit cancelled", got, d, err)
	}
	if loweredShape(prev.File) != shape || !maps.Equal(prev.Annotations, ann) || !slices.Equal(prev.Diagnostics, diags) || !slices.Equal(symbolView(prev), syms) {
		t.Error("the cancelled analysis changed the analysis it started from")
	}
	editAndCompare(t, src, bufs[0])

	// Through the owner: its parse polls before the parse and after it,
	// then analyzeEdit polls after re-lowering.
	s := NewLSPServer()
	var cancel atomic.Bool
	s.compute = func(ctx context.Context, uri string, text []byte, prev *document) (*document, error) {
		if cancel.Load() {
			ctx = &cancelAt{Context: ctx, n: 2}
		}
		return s.analyzeSource(ctx, uri, text, prev)
	}
	live := openLiveWith(t, src, s)
	last := live.s.snapshot(liveURI)
	cancel.Store(true)
	live.version++
	live.s.change(liveURI, live.version, bufs[0])
	if doc := live.s.snapshot(liveURI); doc == nil || doc.analysis != last.analysis || !bytes.Equal(doc.text.Text, src) {
		t.Fatal("after the cancelled parse the owner does not serve its last complete version")
	}
	if loweredShape(last.analysis.File) != shape || !maps.Equal(last.analysis.Annotations, ann) {
		t.Error("the cancelled parse changed the owner's last complete version")
	}
	cancel.Store(false)
	sameAnalysis(t, live.edit(t, bufs[1]), fresh(t, bufs[1]))
}

// TestIncrementalAnalyzeTopLevelEditGoesFull: a non-trivia edit at top
// level, outside every sub, takes the full path; a trivia-only one
// re-analyzes nothing, keeping the very annotations it had (fresh3 F4).
// Both equal a fresh open.
func TestIncrementalAnalyzeTopLevelEditGoesFull(t *testing.T) {
	src := readFixture(t)
	for _, c := range []struct{ name, anchor, old, new string }{
		{"a top-level hash's value", "package Fixture::Block4;", "limit   => 4", "limit   => 41"},
		{"a top-level statement added", "sub limit {", "\n\nsub describe", "\n\nmy $zz = 1;\nsub describe"},
	} {
		t.Run(c.name, func(t *testing.T) {
			edited, _ := replaceOnce(t, src, c.anchor, c.old, c.new)
			if got, _, d := editAndCompare(t, src, edited); !d.Full || got.Walked != nil {
				t.Errorf("damage %+v, walked %v; want the full path", d, got.Walked)
			}
		})
	}
	for _, c := range []struct{ name, anchor, old, new string }{
		{"whitespace between subs", "sub limit {", "\n\nsub describe", "\n \nsub describe"},
		{"a comment's interior", "# Generated", "Generated by", "Generated  by"},
	} {
		t.Run(c.name, func(t *testing.T) {
			edited, _ := replaceOnce(t, src, c.anchor, c.old, c.new)
			prev := fresh(t, src)
			got, d, err := analyzeEdit(context.Background(), parse.Parse(src), src, prev, parse.Parse(edited), edited, infer.Options{})
			if err != nil {
				t.Fatal(err)
			}
			if d.Full || len(d.Subs) != 0 || got.Walked == nil || len(got.Walked) != 0 {
				t.Errorf("damage %+v, walked %v; want nothing re-analyzed", d, got.Walked)
			}
			if reflect.ValueOf(got.Annotations).Pointer() != reflect.ValueOf(prev.Annotations).Pointer() || got.File != prev.File {
				t.Error("the trivia edit did not keep the lowered tree and annotations it had")
			}
			sameAnalysis(t, got, fresh(t, edited))
		})
	}
}

// coreLib is the .pm files directly under $PSC_CORE_LIB, a perl's core
// library, when it is set.
func coreLib() []string {
	dir := os.Getenv("PSC_CORE_LIB")
	if dir == "" {
		return nil
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.pm"))
	return files
}

// TestIncrementalAnalyzeRealEditsMatchFresh: on real files -- edit5k.pl,
// and perl5's lib and a core library when the environment names them --
// each of realEdits' kinds of edit, and then its undoing from the
// incremental result, equals a fresh open of the same bytes.
func TestIncrementalAnalyzeRealEditsMatchFresh(t *testing.T) {
	local, files := 0, append(realFiles(), coreLib()...)
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		perKind := 2
		if path == latencyFixturePath {
			perKind = 1
		}
		prevRoot := parse.Parse(src)
		prev := fresh(t, src)
		for _, e := range realEdits(prevRoot, src, perKind) {
			t.Run(fmt.Sprintf("%s/%q@%d", filepath.Base(path), e.repl, e.at), func(t *testing.T) {
				edited := slices.Concat(src[:e.at], []byte(e.repl), src[e.at+e.n:])
				root := parse.Parse(edited)
				got, d, err := analyzeEdit(context.Background(), prevRoot, src, prev, root, edited, infer.Options{})
				if err != nil {
					t.Fatal(err)
				}
				sameAnalysis(t, got, fresh(t, edited))
				if d.Full {
					return
				}
				local++
				back, _, err := analyzeEdit(context.Background(), root, edited, got, prevRoot, src, infer.Options{})
				if err != nil {
					t.Fatal(err)
				}
				sameAnalysis(t, back, prev)
			})
		}
	}
	if local == 0 {
		t.Fatal("no edit was local; the check checked nothing")
	}
	t.Logf("%d local edits checked, each and its undoing, over %d files", local, len(files))
}
