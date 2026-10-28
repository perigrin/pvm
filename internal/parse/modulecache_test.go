// ABOUTME: Tests for ModuleCache: a cached re-parse equals a fresh uncached Session parse, root lists in order.
// ABOUTME: Covers byte changes under one mtime, created/deleted/shadowed modules, transitive changes, cycles and the fuzz seeds.

package parse_test

import (
	"bufio"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"tamarou.com/pvm/internal/parse"
)

// cacheFixture is a directory of modules for one test.
type cacheFixture struct {
	t   *testing.T
	dir string
}

func newCacheFixture(t *testing.T) *cacheFixture {
	return &cacheFixture{t: t, dir: t.TempDir()}
}

// write writes body at rel under the fixture and returns its path.
func (f *cacheFixture) write(rel, body string) string {
	f.t.Helper()
	p := filepath.Join(f.dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		f.t.Fatal(err)
	}
	return p
}

// overlayCacheDeclarations installs two declarations, each with an error
// and a warning, so a test's DeclarationErrors and DeclarationWarnings are
// non-empty and their order is checked.
func overlayCacheDeclarations(t *testing.T) {
	odd, err := os.ReadFile(filepath.Join(incRoot, "declarations", "Odd.tpm"))
	if err != nil {
		t.Fatal(err)
	}
	even := strings.ReplaceAll(string(odd), "Odd", "Even")
	t.Cleanup(parse.OverlayDeclarations(fstest.MapFS{
		"declarations/Odd.tpm":  {Data: odd},
		"declarations/Even.tpm": {Data: []byte(even)},
	}))
}

// equalsFresh fails t unless got equals a fresh uncached Session parse of
// src at path with roots, its root lists compared as slices in order.
func equalsFresh(t *testing.T, what string, got *parse.Node, path, src string, roots ...string) *parse.Node {
	t.Helper()
	fresh := parse.NewSession().ParseSourceFrom(path, []byte(src), roots...)
	if diff := parse.TreeDiff(got, fresh); diff != "" {
		t.Fatalf("%s: cached != fresh:\n%s", what, diff)
	}
	for _, c := range []struct {
		name string
		of   func(*parse.Node) []string
	}{
		{"LoadedModules", parse.LoadedModules},
		{"DeclarationErrors", func(n *parse.Node) []string { return errorStrings(parse.DeclarationErrors(n)) }},
		{"DeclarationWarnings", func(n *parse.Node) []string { return errorStrings(parse.DeclarationWarnings(n)) }},
	} {
		if g, w := c.of(got), c.of(fresh); !slices.Equal(g, w) {
			t.Fatalf("%s: %s %v, fresh %v", what, c.name, g, w)
		}
	}
	return fresh
}

// TestModuleCacheReparseEqualsFresh: a document's re-parses, each in a
// fresh Session sharing its cache, equal fresh uncached parses, and once
// warm parse no module again. Both forms re-parse: a full parse, with the
// flag off, and Repair, with it on.
func TestModuleCacheReparseEqualsFresh(t *testing.T) {
	overlayCacheDeclarations(t)
	f := newCacheFixture(t)
	f.write("lib/A.pm", "package A;\nuse C;\nuse Odd;\nour @EXPORT = qw(fa);\nsub fa($) { }\n1;\n")
	f.write("lib/C.pm", "package C;\nuse Even;\nour @EXPORT = qw(fc);\nsub fc($$) { }\n1;\n")
	root := filepath.Join(f.dir, "lib")
	path := filepath.Join(f.dir, "doc.pl")
	src := "use A;\nuse C;\nsub g {\n  my @x = (fa 1, 2);\n}\nmy @y = (fc 1, 2, 3);\n"

	for _, flag := range []bool{false, true} {
		t.Run(strconv.FormatBool(flag), func(t *testing.T) {
			was := parse.IncrementalRepair
			parse.IncrementalRepair = flag
			t.Cleanup(func() { parse.IncrementalRepair = was })

			c := parse.NewModuleCache()
			tree := parse.NewSessionWithCache(c).ParseSourceFrom(path, []byte(src), root)
			fresh := equalsFresh(t, "open", tree, path, src, root)
			if len(parse.DeclarationErrors(fresh)) < 2 || len(parse.DeclarationWarnings(fresh)) < 2 {
				t.Fatalf("fresh parse has errors %v, warnings %v; the fixture checks no order", parse.DeclarationErrors(fresh), parse.DeclarationWarnings(fresh))
			}
			if c := findCall(tree, "fa"); c == nil || !c.Resolved || c.Prototype != "($)" {
				t.Fatalf("want fa resolved with ($); got %+v", c)
			}
			warm := parse.ModuleCacheParses(c)
			if warm != 4 {
				t.Fatalf("the cold parse parsed %d modules, want A, C, Odd and Even", warm)
			}
			cur := src
			for _, e := range []parse.Edit{replaceFirst(cur, "1, 2)", "7, 8)"), insertAt(cur, "1, 2, 3);\n", "print 1;\n")} {
				next := applyEdit(cur, e)
				tree = parse.NewSessionWithCache(c).ReparseIncremental(tree, e, path, []byte(next), root)
				equalsFresh(t, "re-parse", tree, path, next, root)
				cur = next
			}
			if got := parse.ModuleCacheParses(c); got != warm {
				t.Errorf("re-parses parsed %d modules again; want none", got-warm)
			}
		})
	}
}

// TestModuleCacheInvalidatesOnByteChange: a module rewritten with its mtime
// put back, one created in an earlier root over a later one's, and that one
// deleted again, are each seen by the next parse, prototype included.
func TestModuleCacheInvalidatesOnByteChange(t *testing.T) {
	f := newCacheFixture(t)
	early, late := filepath.Join(f.dir, "early"), filepath.Join(f.dir, "late")
	lateFoo := f.write("late/Foo.pm", "package Foo;\nour @EXPORT = qw(bar);\nsub bar($) { }\n1;\n")
	if err := os.MkdirAll(early, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(f.dir, "doc.pl")
	src := "use Foo;\nmy @x = (bar 1, 2);\n"
	c := parse.NewModuleCache()
	parseNow := func(what, want string) {
		t.Helper()
		tree := parse.NewSessionWithCache(c).ParseSourceFrom(path, []byte(src), early, late)
		equalsFresh(t, what, tree, path, src, early, late)
		call := findCall(tree, "bar")
		switch {
		case want == "" && call != nil && call.Resolved:
			t.Fatalf("%s: bar resolved with %q; want unresolved", what, call.Prototype)
		case want != "" && (call == nil || !call.Resolved || call.Prototype != want):
			t.Fatalf("%s: want bar resolved with %s; got %+v", what, want, call)
		}
	}
	parseNow("first", "($)")
	parseNow("unchanged", "($)")

	info, err := os.Stat(lateFoo)
	if err != nil {
		t.Fatal(err)
	}
	f.write("late/Foo.pm", "package Foo;\nour @EXPORT = qw(bar);\nsub bar(@) { }\n1;\n")
	if err := os.Chtimes(lateFoo, time.Time{}, info.ModTime()); err != nil {
		t.Fatal(err)
	}
	if again, err := os.Stat(lateFoo); err != nil || !again.ModTime().Equal(info.ModTime()) || again.Size() != info.Size() {
		t.Fatalf("the rewrite changed mtime or size (%v); the test checks nothing", err)
	}
	parseNow("rewritten, same mtime and size", "(@)")

	earlyFoo := f.write("early/Foo.pm", "package Foo;\nour @EXPORT = qw(bar);\nsub bar(\\@) { }\n1;\n")
	parseNow("shadowed in an earlier root", `(\@)`)
	if err := os.Remove(earlyFoo); err != nil {
		t.Fatal(err)
	}
	parseNow("shadow deleted", "(@)")
	if err := os.Remove(lateFoo); err != nil {
		t.Fatal(err)
	}
	parseNow("deleted", "")
}

// TestModuleCacheInvalidatesOnTransitiveDependencyChange: A's bytes do not
// change, but C, which A uses, stops using D. The cached A is invalidated:
// the next parse reads A and C again and loads no D, as a fresh one does.
func TestModuleCacheInvalidatesOnTransitiveDependencyChange(t *testing.T) {
	overlayCacheDeclarations(t)
	f := newCacheFixture(t)
	f.write("lib/A.pm", "package A;\nuse C;\nour @EXPORT = qw(fa);\nsub fa($) { }\n1;\n")
	f.write("lib/C.pm", "package C;\nuse D;\n1;\n")
	f.write("lib/D.pm", "package D;\nuse Odd;\n1;\n")
	root := filepath.Join(f.dir, "lib")
	path := filepath.Join(f.dir, "doc.pl")
	src := "use A;\nmy @x = (fa 1, 2);\n"
	c := parse.NewModuleCache()

	tree := parse.NewSessionWithCache(c).ParseSourceFrom(path, []byte(src), root)
	equalsFresh(t, "first", tree, path, src, root)
	if got := parse.LoadedModules(tree); !slices.Equal(got, []string{"A", "C", "D", "Odd"}) {
		t.Fatalf("loaded %v; the fixture is not what this test assumes", got)
	}
	before := parse.ModuleCacheParses(c)

	f.write("lib/C.pm", "package C;\n1;\n")
	tree = parse.NewSessionWithCache(c).ParseSourceFrom(path, []byte(src), root)
	equalsFresh(t, "after C changed", tree, path, src, root)
	if got := parse.LoadedModules(tree); !slices.Equal(got, []string{"A", "C"}) {
		t.Errorf("loaded %v, want [A C]", got)
	}
	if got := parse.ModuleCacheParses(c) - before; got != 2 {
		t.Errorf("parsed %d modules, want A and C", got)
	}
}

// fuzzSeed is one committed corpus file's arguments.
func fuzzSeed(t *testing.T, path string) []string {
	t.Helper()
	fh, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	var args []string
	sc := bufio.NewScanner(fh)
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		line := sc.Text()
		if line == "go test fuzz v1" || line == "" {
			continue
		}
		open, close := strings.IndexByte(line, '('), strings.LastIndexByte(line, ')')
		if open < 0 || close < open {
			t.Fatalf("%s: unreadable line %q", path, line)
		}
		v := line[open+1 : close]
		if strings.HasPrefix(v, `"`) {
			if v, err = strconv.Unquote(v); err != nil {
				t.Fatalf("%s: %q: %v", path, line, err)
			}
		}
		args = append(args, v)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return args
}

// seedEdits is every committed FuzzIncremental and FuzzIncrementalSequence
// seed, each as its source and edits.
func seedEdits(t *testing.T) map[string]struct {
	src   string
	edits []parse.Edit
} {
	out := map[string]struct {
		src   string
		edits []parse.Edit
	}{}
	for _, dir := range []string{"FuzzIncremental", "FuzzIncrementalSequence"} {
		files, err := filepath.Glob(filepath.Join("testdata", "fuzz", dir, "*"))
		if err != nil || len(files) == 0 {
			t.Fatalf("%s: no seeds (%v)", dir, err)
		}
		for _, file := range files {
			args := fuzzSeed(t, file)
			var edits []parse.Edit
			switch {
			case dir == "FuzzIncremental" && len(args) == 4:
				off, err1 := strconv.Atoi(args[1])
				del, err2 := strconv.Atoi(args[2])
				if err1 != nil || err2 != nil {
					t.Fatalf("%s: %v", file, args)
				}
				edits = []parse.Edit{clampEdit(len(args[0]), off, del, args[3])}
			case dir == "FuzzIncrementalSequence" && len(args) == 2:
				edits = decodeEdits([]byte(args[1]), len(args[0]))
			default:
				t.Fatalf("%s: unexpected arguments %q", file, args)
			}
			out[dir+"/"+filepath.Base(file)] = struct {
				src   string
				edits []parse.Edit
			}{args[0], edits}
		}
	}
	return out
}

// TestIncrementalSeedsEqualFreshWithModuleCacheWarmAndCold: the property
// FuzzIncremental and FuzzIncrementalSequence hold, with a module cache in
// the re-parse, cold and then warm from the whole replay. Then the orders
// a replayed module must keep: A and B each use C, so `use A; use B;`
// loads [A C B] and `use B;` [B C]; and P and Q use each other, so `use P;`
// loads [P Q] and `use Q;` [Q P]. Each is re-parsed warm from the others.
func TestIncrementalSeedsEqualFreshWithModuleCacheWarmAndCold(t *testing.T) {
	repairOn(t)
	t.Cleanup(parse.OverlayDeclarations(os.DirFS(incRoot)))
	seeds := seedEdits(t)
	names := slices.Sorted(func(yield func(string) bool) {
		for k := range seeds {
			if !yield(k) {
				return
			}
		}
	})
	replay := func(what string, c *parse.ModuleCache, names ...string) {
		for _, name := range names {
			s := seeds[name]
			tree := parse.NewSessionWithCache(c).ParseSourceFrom(incPath, []byte(s.src), incRoot)
			equalsFresh(t, what+" "+name+" open", tree, incPath, s.src, incRoot)
			cur := s.src
			for _, e := range s.edits {
				cur = applyEdit(cur, e)
				tree = parse.NewSessionWithCache(c).ReparseIncremental(tree, e, incPath, []byte(cur), incRoot)
				equalsFresh(t, what+" "+name, tree, incPath, cur, incRoot)
			}
		}
	}
	for _, name := range names {
		replay("cold", parse.NewModuleCache(), name)
	}
	warm := parse.NewModuleCache()
	replay("warming", warm, names...)
	parses := parse.ModuleCacheParses(warm)
	replay("warm", warm, names...)
	if got := parse.ModuleCacheParses(warm); got != parses {
		t.Errorf("the warm replay parsed %d modules again", got-parses)
	}

	overlayCacheDeclarations(t)
	f := newCacheFixture(t)
	f.write("lib/A.pm", "package A;\nuse C;\nuse Odd;\nour @EXPORT = qw(fa);\nsub fa($) { }\n1;\n")
	f.write("lib/B.pm", "package B;\nuse C;\nuse Even;\nour @EXPORT = qw(fb);\nsub fb($) { }\n1;\n")
	f.write("lib/C.pm", "package C;\nuse Odd;\nour @EXPORT = qw(fc);\nsub fc($$) { }\n1;\n")
	f.write("lib/P.pm", "package P;\nuse Q;\nuse Odd;\nour @EXPORT = qw(fp);\nsub fp($) { }\n1;\n")
	f.write("lib/Q.pm", "package Q;\nuse P;\nuse Even;\nour @EXPORT = qw(fq);\nsub fq($) { }\n1;\n")
	root := filepath.Join(f.dir, "lib")
	path := filepath.Join(f.dir, "doc.pl")
	orders := []struct{ src, loaded string }{
		{"use A; use B;\nmy @x = (fa 1, 2);\n", "A C Odd B Even"},
		{"use B; use A;\nmy @x = (fb 1, 2);\n", "B C Odd Even A"},
		{"use B;\nmy @x = (fb 1, 2);\n", "B C Odd Even"},
		{"use P;\nmy @x = (fp 1, 2);\n", "P Q Even Odd"},
		{"use Q;\nmy @x = (fq 1, 2);\n", "Q P Odd Even"},
	}
	c := parse.NewModuleCache()
	for round := range 2 {
		for _, o := range orders {
			for _, e := range []parse.Edit{{}, replaceFirst(o.src, "1, 2", "3, 4")} {
				var tree *parse.Node
				if e == (parse.Edit{}) {
					tree = parse.NewSessionWithCache(c).ParseSourceFrom(path, []byte(o.src), root)
				} else {
					old := parse.NewSessionWithCache(c).ParseSourceFrom(path, []byte(o.src), root)
					tree = parse.NewSessionWithCache(c).ReparseIncremental(old, e, path, []byte(applyEdit(o.src, e)), root)
				}
				src := applyEdit(o.src, e)
				equalsFresh(t, "round "+strconv.Itoa(round)+" "+o.src, tree, path, src, root)
				if got := strings.Join(parse.LoadedModules(tree), " "); got != o.loaded {
					t.Errorf("%q: loaded %s, want %s", src, got, o.loaded)
				}
			}
		}
	}
	// What a module on a cycle says depends on which of it was entered
	// first, so neither is kept; the modules off it are.
	for name, want := range map[string]bool{"A": true, "B": true, "C": true, "Odd": true, "P": false, "Q": false} {
		if got := parse.ModuleCacheHolds(c, name); got != want {
			t.Errorf("cache holds %s: %v, want %v", name, got, want)
		}
	}
}
