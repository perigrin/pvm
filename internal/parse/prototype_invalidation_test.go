// ABOUTME: §7.7.3's prototype row: an edit that can change a prototype a later call reads makes Repair return ok=false.
// ABOUTME: Covers prototype, BEGIN, use and module-byte edits, predeclarations, inert prototype text, and an edit toggled back.

package parse_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// protoCase is one edit that changes, or could change, the prototype a
// later call to f reads. Each is a committed FuzzIncremental seed named
// "prototype-" + name.
type protoCase struct {
	name string
	src  string
	edit parse.Edit
}

const (
	protoBeginSrc = "BEGIN { my $n = 1; }\nsub f() { 1 }\nmy @a = (1, 2);\nprint f @a;\n"
	protoUseSrc   = "use Foo;\nsub f() { 1 }\nmy @a = (1, 2);\nprint f @a;\n"
	// f's prototype is read by a later call in another sub, outside any
	// region a local repair would re-parse, and by one at file level.
	protoLaterSrc   = "sub f($) { 1 }\nsub g { my @x = (f 1, 2); }\nsub h { return f 3, 4; }\nmy @y = (f 5, 6);\n"
	protoNoneSrc    = "sub f { 1 }\nsub g { my @x = (f 1, 2); }\nsub h { return f 3, 4; }\nmy @y = (f 5, 6);\n"
	protoPredeclSrc = "sub f($);\nsub g { my @x = (f 1, 2); }\nmy @y = (f 5, 6);\nsub f { 1 }\n"
	protoDefSrc     = "sub f($) { 1 }\nsub g { my @x = (f 1, 2); }\nmy @y = (f 5, 6);\n"
)

// protoCases is the prototype row (the first four) and the other ways a
// prototype can reach a later call, each a path Repair might let past.
var protoCases = []protoCase{
	{"edit", protoSrc, insertAt(protoSrc, "sub f(", `\@`)},
	{"begin", protoBeginSrc, replaceFirst(protoBeginSrc, "1", "2")},
	{"use", protoUseSrc, replaceFirst(protoUseSrc, "use Foo;", "use Foo ();")},
	protoRow("use-constant", "use constant f => 1;\nmy @a = (1, 2);\nprint f @a;\n", "f =>", "g =>"),
	protoRow("attribute", "sub f :prototype($) { 1 }\nmy @x = (f 1, 2);\n", "($)", "(@)"),
	protoRow("declared-twice", "sub f($);\nsub f($) { 1 }\nmy @x = (f 1, 2);\n", "($) {", "(@) {"),
	protoRow("nested-sub", "sub g {\n  sub f($) { 1 }\n}\nmy @x = (f 1, 2);\n", "($)", "(@)"),
	protoRow("lexical-sub", "my sub f($) { 1 }\nmy @x = (f 1, 2);\n", "($)", "(@)"),
	protoRow("glob-in-begin", "BEGIN { *f = sub ($) { 1 } }\nmy @x = (f 1, 2);\n", "($)", "(@)"),
	protoRow("glob-in-sub", "sub g {\n  *f = sub ($) { 1 };\n}\nmy @x = (f 1, 2);\n", "($)", "(@)"),
	{"remove", protoLaterSrc, replaceFirst(protoLaterSrc, "sub f($)", "sub f")},
	{"add", protoNoneSrc, replaceFirst(protoNoneSrc, "sub f ", "sub f($) ")},
	{"predeclaration", protoPredeclSrc, replaceFirst(protoPredeclSrc, "($)", "(@)")},
	{"definition", protoDefSrc, replaceFirst(protoDefSrc, "($)", "(@)")},
}

// protoRow is the protoCase replacing the first old in src with text.
func protoRow(name, src, old, text string) protoCase {
	return protoCase{name, src, replaceFirst(src, old, text)}
}

// protoCaseNamed is the protoCases entry called name.
func protoCaseNamed(t *testing.T, name string) protoCase {
	t.Helper()
	for _, c := range protoCases {
		if c.name == name {
			return c
		}
	}
	t.Fatalf("no protoCase %q", name)
	return protoCase{}
}

// checkProtoRefused asserts Repair's ok=false for c in both forms, that
// ReparseIncremental with the flag on equals a fresh parse in both, and
// that c's committed seed is its edit.
func checkProtoRefused(t *testing.T, c protoCase) {
	t.Helper()
	edited := applyEdit(c.src, c.edit)
	old := parse.NewSession().ParseSourceFrom(incPath, []byte(c.src), incRoot)
	if tree, ok := parse.NewSession().Repair(old, c.edit, incPath, []byte(edited), incRoot); ok || tree != nil {
		t.Errorf("Session: Repair ok=%v, want ok=false and no tree", ok)
	}
	if tree, ok := parse.Repair(parse.Parse([]byte(c.src)), c.edit, []byte(edited)); ok || tree != nil {
		t.Errorf("plain: Repair ok=%v, want ok=false and no tree", ok)
	}
	checkIncremental(t, c.src, c.edit)
	got, err := os.ReadFile(filepath.Join("testdata", "fuzz", "FuzzIncremental", "prototype-"+c.name))
	if want := fuzzIncrementalCorpus(c.src, c.edit); err != nil || string(got) != want {
		t.Errorf("seed prototype-%s: want\n%s\ngot (err %v)\n%s", c.name, want, err, got)
	}
}

// laterCallsChange asserts that the fresh parses of src before and after e
// differ in the first call to f, so the edit is one a prototype decides. A
// call under `sub f()` taking an argument is no call at all.
func laterCallsChange(t *testing.T, src string, e parse.Edit) {
	t.Helper()
	before, after := findCall(freshParse(src), "f"), findCall(freshParse(applyEdit(src, e)), "f")
	if before == nil && after == nil {
		t.Fatal("no call to f before or after the edit")
	}
	if before != nil && after != nil && parse.TreeDiffShifted(before, after, after.Start-before.Start) == "" {
		t.Fatal("the call to f parses the same before and after; the edit checks nothing")
	}
}

// TestPrototypeEditInvalidation is §7.7.3's prototype row: a prototype edit,
// an edit inside BEGIN, an edit to a `use` line and a change to a `use`d
// module's bytes each make Repair return ok=false, and a later call to f
// parses as a fresh parse does. The other rows of protoCases are the paths
// a prototype could take past Repair.
func TestPrototypeEditInvalidation(t *testing.T) {
	for _, c := range protoCases {
		t.Run(c.name, func(t *testing.T) { checkProtoRefused(t, c) })
	}
	t.Run("edit changes the call", func(t *testing.T) { laterCallsChange(t, protoSrc, protoCases[0].edit) })
	t.Run("module-bytes", testModuleBytesInvalidate)
	t.Run("module-comment", testModuleCommentKeepsProperty)
	t.Run("module-prototype-change", testModulePrototypeChangeKeepsProperty)
}

// testModuleBytesInvalidate: a module on disk installs f's prototype, and
// changing its bytes refuses a sub-body edit Repair would otherwise
// accept, both from a fresh parse's tree and from a repaired tree that
// carries its module reads.
// modSrc uses Mod, which exports f, and has a sub body to edit locally.
const modSrc = "use Mod;\nsub g {\n  my $y = 1;\n}\nmy @x = (f 1, 2);\n"

// modFixture is a script path and a lib root on disk, writeMod putting Mod.pm
// there with f's prototype proto after the lines extra, and fresh a fresh
// Session parse of src at path.
func modFixture(t *testing.T) (path, root string, writeMod func(proto, extra string), fresh func(src string) *parse.Node) {
	dir := t.TempDir()
	path, root = filepath.Join(dir, "edit.pl"), filepath.Join(dir, "lib")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	writeMod = func(proto, extra string) {
		t.Helper()
		body := "package Mod;\n" + extra + "our @EXPORT = qw(f);\nsub f" + proto + " { }\n1;\n"
		if err := os.WriteFile(filepath.Join(root, "Mod.pm"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fresh = func(src string) *parse.Node { return parse.NewSession().ParseSourceFrom(path, []byte(src), root) }
	return path, root, writeMod, fresh
}

// testModuleCommentKeepsProperty: a module change that leaves its prototype
// alone, on a first Repair, may be accepted or refused, but an accepted tree
// equals a fresh parse (§6.5.2 check 3). The edit adds a call to f inside
// g, so the module's facts reach the re-parsed statement.
func testModuleCommentKeepsProperty(t *testing.T) {
	path, root, writeMod, fresh := modFixture(t)
	e := replaceFirst(modSrc, "1", "f 7, 8")
	out := applyEdit(modSrc, e)
	writeMod("($)", "")
	old := fresh(modSrc)
	writeMod("($)", "# a comment\n")
	tree, ok := parse.NewSession().Repair(old, e, path, []byte(out), root)
	if !ok {
		if tree != nil {
			t.Fatal("ok=false with a tree")
		}
		return
	}
	if diff := parse.TreeDiff(tree, fresh(out)); diff != "" {
		t.Fatalf("ok=true but repaired != fresh:\n%s", diff)
	}
}

// testModulePrototypeChangeKeepsProperty: like testModuleCommentKeepsProperty,
// but the module change is to the prototype of Mod::k, which only the edit
// calls, so a stale read of Mod.pm would parse that call wrongly while the
// old tree, which has no call to k, re-parses equal.
func testModulePrototypeChangeKeepsProperty(t *testing.T) {
	path, root, writeMod, fresh := modFixture(t)
	const src = "use Mod;\nsub g {\n  my $y = 1;\n}\n"
	e := replaceFirst(src, "1", "Mod::k 7, 8")
	out := applyEdit(src, e)
	writeMod("($)", "sub k($) { }\n")
	old, before := fresh(src), fresh(out)
	writeMod("($)", "sub k(@) { }\n")
	if parse.TreeDiff(before, fresh(out)) == "" {
		t.Fatal("k's prototype change leaves the edited source's parse unchanged; the test checks nothing")
	}
	tree, ok := parse.NewSession().Repair(old, e, path, []byte(out), root)
	if !ok {
		if tree != nil {
			t.Fatal("ok=false with a tree")
		}
		return
	}
	if diff := parse.TreeDiff(tree, fresh(out)); diff != "" {
		t.Fatalf("ok=true but repaired != fresh:\n%s", diff)
	}
}

func testModuleBytesInvalidate(t *testing.T) {
	repairOn(t)
	path, root, writeMod, fresh := modFixture(t)
	const src = modSrc
	e1 := replaceFirst(src, "1", "2")
	src1 := applyEdit(src, e1)
	e2 := replaceFirst(src1, "2", "3")
	src2 := applyEdit(src1, e2)

	writeMod("($)", "")
	old := fresh(src)
	repaired, ok := parse.NewSession().Repair(old, e1, path, []byte(src1), root)
	if !ok {
		t.Fatal("a sub-body edit with the module unchanged returned ok=false; the module check is unexercised")
	}
	if diff := parse.TreeDiff(repaired, fresh(src1)); diff != "" {
		t.Fatalf("repaired != fresh:\n%s", diff)
	}

	writeMod("(@)", "")
	if diff := parse.TreeDiffShifted(findCall(old, "f"), findCall(fresh(src), "f"), 0); diff == "" {
		t.Fatal("the module's prototype change leaves the call to f unchanged; the test checks nothing")
	}
	for _, c := range []struct {
		name     string
		old      *parse.Node
		e        parse.Edit
		src, out string
	}{{"fresh tree", old, e1, src, src1}, {"repaired tree", repaired, e2, src1, src2}} {
		if tree, ok := parse.NewSession().Repair(c.old, c.e, path, []byte(c.out), root); ok || tree != nil {
			t.Errorf("%s: Repair ok=%v after the module changed, want ok=false", c.name, ok)
		}
		inc := parse.NewSession().ReparseIncremental(c.old, c.e, path, []byte(c.out), root)
		if diff := parse.TreeDiff(inc, fresh(c.out)); diff != "" {
			t.Errorf("%s: ReparseIncremental != fresh:\n%s", c.name, diff)
		}
		if call := findCall(inc, "f"); call == nil || call.Prototype != "(@)" {
			t.Errorf("%s: call to f %+v, want the module's new prototype (@)", c.name, call)
		}
	}
}

// TestPrototypeRemovalInvalidatesLaterCalls: removing f's prototype, or
// adding one where none was, re-parses every later call to f, the ones in
// g and h, outside any local re-parse region, included.
func TestPrototypeRemovalInvalidatesLaterCalls(t *testing.T) {
	for _, name := range []string{"remove", "add"} {
		t.Run(name, func(t *testing.T) {
			c := protoCaseNamed(t, name)
			checkProtoRefused(t, c)
			laterCallsChange(t, c.src, c.edit)
		})
	}
}

// TestPrototypePredeclarationEdit: editing the forward declaration
// `sub f($);` is refused and re-parses later calls as editing the
// definition's prototype is and does.
func TestPrototypePredeclarationEdit(t *testing.T) {
	for _, name := range []string{"predeclaration", "definition"} {
		t.Run(name, func(t *testing.T) {
			c := protoCaseNamed(t, name)
			checkProtoRefused(t, c)
			laterCallsChange(t, c.src, c.edit)
		})
	}
}

// TestPrototypeTextInStringOrPODIsInert: prototype-like text edited inside
// a string, a heredoc or POD gives a tree equal to a fresh parse, and the
// edit inside a sub body's string repairs locally.
func TestPrototypeTextInStringOrPODIsInert(t *testing.T) {
	const src = "sub f($) { 1 }\nsub g {\n  my $s = 'sub f($)';\n  return $s;\n}\n" +
		"=pod\n\nsub f($)\n\n=cut\n\nmy $h = <<'EOF';\nsub f($)\nEOF\nmy @y = (f 5, 6);\n"
	// at is the edit of the nth `sub f($)`'s `$` to `\@`, 0 being the real one.
	at := func(nth int) parse.Edit {
		i := 0
		for range nth {
			i += strings.Index(src[i+1:], "sub f($)") + 1
		}
		i += len("sub f(")
		return parse.Edit{Start: i, OldEnd: i + 1, NewText: `\@`}
	}
	for i, name := range []string{"string", "pod", "heredoc"} {
		t.Run(name, func(t *testing.T) { checkIncremental(t, src, at(i+1)) })
	}
	if !repairEqualsFresh(t, src, at(1)) {
		t.Error("an edit inside a sub body's string returned ok=false, want a local repair")
	}
}

// TestPrototypeEditToggleSequence: `sub f()` to `sub f(\@)` and back on one
// tree, through ReparseIncremental with the flag on, equals a fresh parse
// at every step, and the last is the original source's fresh parse.
func TestPrototypeEditToggleSequence(t *testing.T) {
	repairOn(t)
	there := insertAt(protoSrc, "sub f(", `\@`)
	back := parse.Edit{Start: there.Start, OldEnd: there.Start + len(there.NewText)}
	s := parse.NewSession()
	tree, plain, cur := s.ParseSourceFrom(incPath, []byte(protoSrc), incRoot), parse.Parse([]byte(protoSrc)), protoSrc
	for i, e := range []parse.Edit{there, back, there, back} {
		cur = applyEdit(cur, e)
		tree = s.ReparseIncremental(tree, e, incPath, []byte(cur), incRoot)
		plain = parse.ReparseIncremental(plain, e, []byte(cur))
		if diff := parse.TreeDiff(tree, freshParse(cur)); diff != "" {
			t.Fatalf("step %d: Session != fresh:\n%s", i, diff)
		}
		if diff := parse.TreeDiff(plain, parse.Parse([]byte(cur))); diff != "" {
			t.Fatalf("step %d: plain != fresh:\n%s", i, diff)
		}
	}
	if diff := parse.TreeDiff(tree, freshParse(protoSrc)); diff != "" {
		t.Fatalf("toggled back != the original fresh parse:\n%s", diff)
	}
	got, err := os.ReadFile(filepath.Join("testdata", "fuzz", "FuzzIncrementalSequence", "prototype-toggle"))
	if want := fuzzSequenceCorpus(protoSrc, there, back); err != nil || string(got) != want {
		t.Errorf("seed prototype-toggle: want\n%s\ngot (err %v)\n%s", want, err, got)
	}
}
