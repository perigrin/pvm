// ABOUTME: Tests damageFromTreeDiff: which top-level subs an edit damages, and when it goes full (Class B, top level, errors).
// ABOUTME: Real files are edited in place, and every sub reported undamaged is checked to analyze the same before and after.

package psc

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// replaceOnce is src with the first old after the first anchor replaced by
// new, and the offset of the replacement; the test fails if either is absent.
func replaceOnce(t *testing.T, src []byte, anchor, old, new string) ([]byte, int) {
	t.Helper()
	a := bytes.Index(src, []byte(anchor))
	if a < 0 {
		t.Fatalf("anchor %q not in source", anchor)
	}
	i := bytes.Index(src[a:], []byte(old))
	if i < 0 {
		t.Fatalf("%q not after anchor %q", old, anchor)
	}
	at := a + i
	return slices.Concat(src[:at], []byte(new), src[at+len(old):]), at
}

func damageOf(old, new []byte) editDamage {
	return damageFromTreeDiff(parse.Parse(old), parse.Parse(new))
}

func readFixture(t *testing.T) []byte {
	t.Helper()
	src, err := os.ReadFile(latencyFixturePath)
	if err != nil {
		t.Fatal(err)
	}
	return src
}

// TestDamageFromTreeDiff: the benchmark's 1-char edit, before the `0` of
// `$_ > 0` in the middle `sub scale`, damages that sub and no other, and
// does not go full.
func TestDamageFromTreeDiff(t *testing.T) {
	src := readFixture(t)
	at := editOffset(t, src)
	edited := slices.Concat(src[:at], []byte("1"), src[at:])
	d := damageOf(src, edited)
	want := fmt.Sprintf("sub scale#%d", len(scaleSubs(src))/2)
	if d.Full || !slices.Equal(d.Subs, []string{want}) {
		t.Fatalf("damage = %+v, want exactly [%s]", d, want)
	}
	// Every edit of the benchmark's sequence damages exactly one sub.
	bufs := latencyEdits(src, 20)
	prev := src
	for k, buf := range bufs {
		s := k % len(scaleSubs(src))
		if d := damageOf(prev, buf); d.Full || !slices.Equal(d.Subs, []string{fmt.Sprintf("sub scale#%d", s)}) {
			t.Errorf("edit %d: damage = %+v, want [sub scale#%d]", k, d, s)
		}
		prev = buf
	}
}

// TestDamageClassBGoesFull: an edit that changes a declaration -- a sub's
// name or prototype, a package, a use or no, an our, a BEGIN block, or a
// class -- is Class B and goes full, inside a sub as at top level.
func TestDamageClassBGoesFull(t *testing.T) {
	src := readFixture(t)
	cases := []struct{ name, anchor, old, new string }{
		{"rename a sub", "sub scale {", "scale", "scaled"},
		{"add a prototype", "sub limit {", "limit {", "limit($) {"},
		{"rename a package", "package Fixture::Block3;", "Block3", "Block3x"},
		{"change a use", "use warnings;", "warnings", "warnings FATAL => 'all'"},
		{"add a use inside a sub", "sub scale {", "{", "{ use integer;"},
		{"add a no inside a sub", "sub scale {", "{", "{ no strict;"},
		{"add an our inside a sub", "sub scale {", "{", "{ our $Z;"},
		{"add a BEGIN inside a sub", "sub scale {", "{", "{ BEGIN { 1 }"},
		{"add a nested named sub", "sub scale {", "{", "{ sub inner { 1 }"},
		{"edit an our", "our $VERSION = '0.003'", "0.003", "0.004"},
		{"add a sub", "sub limit {", "sub limit {", "sub fresh { 1 }\nsub limit {"},
		{"add a package", "sub limit {", "sub limit {", "package Other;\nsub limit {"},
		{"add a class", "sub limit {", "sub limit {", "class Other;\nsub limit {"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			edited, _ := replaceOnce(t, src, c.anchor, c.old, c.new)
			if d := damageOf(src, edited); !d.Full || !d.ClassB {
				t.Errorf("damage = %+v, want Full and ClassB", d)
			}
		})
	}
	small := []byte("sub f { 1 }\nBEGIN { $x = 1 }\nsub g { f() }\n")
	edited, _ := replaceOnce(t, small, "BEGIN", "1", "2")
	if d := damageOf(small, edited); !d.Full || !d.ClassB {
		t.Errorf("editing a BEGIN block: damage = %+v, want Full and ClassB", d)
	}
}

// TestDamageClassBOnDeletion: removing or renaming a sub, package, use, no,
// our or BEGIN is Class B, as adding one is.
func TestDamageClassBOnDeletion(t *testing.T) {
	src := []byte(`package P;
use strict;
no warnings 'once';
our $V = 1;
BEGIN { $main::x = 1 }
sub a { return 1 }
sub b { my $y = a(); return $y }
package Q;
sub c { 2 }
1;
`)
	cases := []struct{ name, old, new string }{
		{"delete a sub", "sub a { return 1 }\n", ""},
		{"rename a sub", "sub b {", "sub bb {"},
		{"delete a package", "package Q;\n", ""},
		{"rename a package", "package Q;", "package R;"},
		{"delete a use", "use strict;\n", ""},
		{"delete a no", "no warnings 'once';\n", ""},
		{"delete an our", "our $V = 1;\n", ""},
		{"rename an our", "our $V", "our $W"},
		{"delete a BEGIN", "BEGIN { $main::x = 1 }\n", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			edited, _ := replaceOnce(t, src, c.old, c.old, c.new)
			if d := damageOf(src, edited); !d.Full || !d.ClassB {
				t.Errorf("damage = %+v, want Full and ClassB", d)
			}
		})
	}
}

// TestDamageShiftedSubsNotDamaged: a length-changing edit shifts every
// later sub, which TreeDiff sees as a change and TreeDiffShifted by the
// edit's delta does not; the later subs are not damaged.
func TestDamageShiftedSubsNotDamaged(t *testing.T) {
	src := readFixture(t)
	for _, c := range []struct{ name, old, new string }{
		{"grow by 2", "$_ > 0", "$_ > 100"},
		{"shrink by 1", "$a <=> $b", "$a<=> $b"},
	} {
		t.Run(c.name, func(t *testing.T) {
			edited, _ := replaceOnce(t, src, "package Fixture::Block10;", c.old, c.new)
			oldRoot, newRoot := parse.Parse(src), parse.Parse(edited)
			d := damageFromTreeDiff(oldRoot, newRoot)
			if d.Full || !slices.Equal(d.Subs, []string{"sub scale#10"}) {
				t.Fatalf("damage = %+v, want exactly [sub scale#10]", d)
			}
			// The precondition the test exists for: a later sub differs
			// by TreeDiff and not by TreeDiffShifted.
			oldSubs, newSubs := outerSubs(oldRoot), outerSubs(newRoot)
			o, n := oldSubs[len(oldSubs)-1], newSubs[len(newSubs)-1]
			delta := len(edited) - len(src)
			if parse.TreeDiff(o.node, n.node) == "" || parse.TreeDiffShifted(o.node, n.node, delta) != "" {
				t.Fatalf("the last sub is not a shifted copy: TreeDiff %q, TreeDiffShifted %q",
					parse.TreeDiff(o.node, n.node), parse.TreeDiffShifted(o.node, n.node, delta))
			}
		})
	}
}

// TestDamageTopLevelTriviaOnlyIsNoDamage: an edit inside one top-level
// trivia token -- whitespace between two subs, a comment's interior, POD's
// interior -- damages nothing and does not go full.
func TestDamageTopLevelTriviaOnlyIsNoDamage(t *testing.T) {
	src := readFixture(t)
	for _, c := range []struct{ name, anchor, old, new string }{
		{"whitespace between subs", "sub limit {", "\n\nsub describe", "\n \nsub describe"},
		{"a blank line between subs", "sub limit {", "\n\nsub describe", "\n\n\nsub describe"},
		{"a comment's interior", "# Generated", "Generated by", "Generated  by"},
		{"POD's interior", "=head2 Block7", "C<new>", "C<new()>"},
	} {
		t.Run(c.name, func(t *testing.T) {
			edited, _ := replaceOnce(t, src, c.anchor, c.old, c.new)
			if d := damageOf(src, edited); d.Full || len(d.Subs) != 0 {
				t.Errorf("damage = %+v, want none", d)
			}
		})
	}
}

// TestDamageTopLevelEdit: an edit outside every sub that changes a
// non-trivia node goes full, and is not Class B.
func TestDamageTopLevelEdit(t *testing.T) {
	src := readFixture(t)
	for _, c := range []struct{ name, anchor, old, new string }{
		{"a top-level hash's value", "package Fixture::Block4;", "limit   => 4", "limit   => 41"},
		{"the final statement", "package main;", "1;", "2;"},
		{"a top-level statement added", "sub limit {", "\n\nsub describe", "\n\nmy $zz = 1;\nsub describe"},
		{"a comment that swallows code", "package main;", "$next->() for", "# $next->() for"},
	} {
		t.Run(c.name, func(t *testing.T) {
			edited, _ := replaceOnce(t, src, c.anchor, c.old, c.new)
			if d := damageOf(src, edited); !d.Full || d.ClassB {
				t.Errorf("damage = %+v, want Full and not ClassB", d)
			}
		})
	}
}

// TestDamageIdenticalAndEmptyTrees: equal trees give no damage, and an
// empty tree on either side, or a nil one, does not panic.
func TestDamageIdenticalAndEmptyTrees(t *testing.T) {
	src := readFixture(t)
	if d := damageOf(src, src); d.Full || len(d.Subs) != 0 {
		t.Errorf("identical trees: damage = %+v, want none", d)
	}
	if d := damageOf(nil, nil); d.Full || len(d.Subs) != 0 {
		t.Errorf("empty to empty: damage = %+v, want none", d)
	}
	if d := damageFromTreeDiff(nil, nil); d.Full || len(d.Subs) != 0 {
		t.Errorf("nil to nil: damage = %+v, want none", d)
	}
	// Emptying a file, or filling an empty one, goes full: even a comment
	// is a token added rather than an edit inside one.
	for _, s := range [][]byte{src, []byte("1;\n"), []byte("sub f { 1 }\n"), []byte("# only a comment\n")} {
		for _, d := range []editDamage{
			damageOf(nil, s), damageOf(s, nil),
			damageFromTreeDiff(nil, parse.Parse(s)), damageFromTreeDiff(parse.Parse(s), nil),
		} {
			if !d.Full || len(d.Subs) != 0 {
				t.Errorf("%.20q to or from empty: damage = %+v, want Full", s, d)
			}
		}
	}
}

// TestDamageErrorStatusChange: an edit that adds or removes a parse error
// goes full, a missing `}` that re-nests every later sub among them.
func TestDamageErrorStatusChange(t *testing.T) {
	src := readFixture(t)
	broken := map[string][]byte{}
	for _, c := range []struct{ name, anchor, old, new string }{
		{"delete a sub's closing brace", "sub scale {", "\n}\n", "\n\n"},
		{"unbalance a paren", "sub scale {", "grep {", "grep ( {"},
		{"delete a statement's semicolon", "sub classify {", "unless length $word;", "unless length $word"},
	} {
		t.Run(c.name, func(t *testing.T) {
			edited, _ := replaceOnce(t, src, c.anchor, c.old, c.new)
			if !parse.Parse(edited).HasError {
				t.Fatalf("the edit parses without an error; the test needs one")
			}
			if d := damageOf(src, edited); !d.Full {
				t.Errorf("adding the error: damage = %+v, want Full", d)
			}
			if d := damageOf(edited, src); !d.Full {
				t.Errorf("removing the error: damage = %+v, want Full", d)
			}
			broken[c.name] = edited
		})
	}
	// Editing elsewhere in a file that stays broken is ordinary damage.
	if b, ok := broken["unbalance a paren"]; ok {
		edited, _ := replaceOnce(t, b, "package Fixture::Block20;", "$_ > 0", "$_ > 10")
		if d := damageOf(b, edited); d.Full || !slices.Equal(d.Subs, []string{"sub scale#20"}) {
			t.Errorf("an edit in a file that stays broken: damage = %+v, want [sub scale#20]", d)
		}
	}
}

// TestDamageReachesCallers: a sub's analysis reads what the subs it calls
// return, so the callers of a damaged sub are damaged, transitively; a
// sub that calls none of them is not. Top-level code calling one goes
// full, as does an edit that changes which variables a sub assigns.
func TestDamageReachesCallers(t *testing.T) {
	src := []byte(`sub a { return 1 }
sub b { my $y = a(); return $y + 1 }
sub c { return &b }
sub d { return sort by_num 1, 2 }
sub by_num { $a <=> $b }
sub e { return 2 }
my $top = e();
`)
	cases := []struct {
		name, old, new string
		full           bool
		subs           []string
	}{
		{"callers transitively", "return 1 }", "return 'x' }", false, []string{"sub a#0", "sub b#0", "sub c#0"}},
		{"a sort comparator", "$a <=> $b", "$b <=> $a", false, []string{"sub d#0", "sub by_num#0"}},
		{"called from top level", "return 2", "return 'two'", true, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			edited, _ := replaceOnce(t, src, c.old, c.old, c.new)
			d := damageOf(src, edited)
			if d.Full != c.full || (!c.full && !slices.Equal(d.Subs, c.subs)) {
				t.Errorf("damage = %+v, want Full=%v Subs=%v", d, c.full, c.subs)
			}
		})
	}
}

// TestDamageNewAssignmentGoesFull: a sub that starts assigning a variable
// changes how every other sub reading it is analyzed -- here r's `$p eq
// 'n'` stops being an Undef coercion, as perl 5.42 -w is silent on it once
// w can set $p -- so the edit goes full though only w changed.
func TestDamageNewAssignmentGoesFull(t *testing.T) {
	old := []byte("my $p = undef; sub r { return $p eq 'n' } sub w { return 1 }\n")
	edited, _ := replaceOnce(t, old, "sub w", "{ ", "{ $p = 'n'; ")
	diags := func(src []byte) int {
		return len(infer.AnalyzeParsed(parse.Parse(src), nil, infer.Options{}).Diagnostics)
	}
	if diags(old) != 1 || diags(edited) != 0 {
		t.Fatalf("diagnostics %d -> %d, want 1 -> 0: the case no longer shows the dependency", diags(old), diags(edited))
	}
	if d := damageOf(old, edited); !d.Full {
		t.Errorf("damage = %+v, want Full", d)
	}
	same := []byte("my $p = undef; sub r { return $p eq 'n' } sub w { my $z; $z = 'n'; return 1 }\n")
	edited, _ = replaceOnce(t, same, "sub w", "return 1", "return 2")
	if d := damageOf(same, edited); d.Full || !slices.Equal(d.Subs, []string{"sub w#0"}) {
		t.Errorf("an edit keeping w's assignments, all to its own variable: damage = %+v, want [sub w#0]", d)
	}
}

// TestDamageAssignmentInDeclaration: an assignment inside a declaration's
// initialiser writes its variable as any other does, so a sub that starts
// writing a file-scoped variable that way goes full (review finding 1).
func TestDamageAssignmentInDeclaration(t *testing.T) {
	src := []byte("my $p = 5;\nsub w { my $q = 1 }\nsub r { return $p + 1 }\n")
	for _, repl := range []string{
		"my $q = $p = undef", "state $q = $p = undef", "my $q = ($p = undef)", "my @q = ($p = undef)",
		"my $q = ($p ||= undef)", "local $_ = $p = undef", "my ($q) = ($p) = (undef)", "my $q = [$p = undef]",
	} {
		t.Run(repl, func(t *testing.T) {
			if d := checkEdit(t, src, bytes.Index(src, []byte("my $q = 1")), len("my $q = 1"), repl); !d.Full {
				t.Errorf("damage = %+v, want Full", d)
			}
		})
	}
}

// TestDamageAggregateAssignmentGoesFull: what a sub assigns a file-scoped
// array or hash is the element type every later reader sees, so an edit to
// a sub that assigns one goes full (review finding 2).
func TestDamageAggregateAssignmentGoesFull(t *testing.T) {
	for _, c := range []struct{ src, old, new string }{
		{"my %h;\nsub w { %h = () }\nsub r { return $h{a} }\n", "()", "undef"},
		{"my @c;\nsub w { @c = caller }\nsub r { return $c[0] }\n", "caller", "undef"},
		{"my @c;\n{ sub w { @c = caller } }\nmy $z = $c[0];\n", "caller", "undef"},
	} {
		t.Run(c.src, func(t *testing.T) {
			src := []byte(c.src)
			if d := checkEdit(t, src, bytes.Index(src, []byte(c.old)), len(c.old), c.new); !d.Full {
				t.Errorf("damage = %+v, want Full", d)
			}
		})
	}
	// A file-scoped scalar goes full too: its class and possible-undef
	// mark are set on the file's symbol the same way.
	src := []byte("my $s;\nsub w { $s = 1 }\nsub r { return $s + 1 }\n")
	if d := checkEdit(t, src, bytes.Index(src, []byte("= 1")), len("= 1"), "= undef"); !d.Full {
		t.Errorf("a file-scoped scalar: damage = %+v, want Full", d)
	}
	// A sub's own hash is its own: that edit stays local.
	src = []byte("my %h;\nsub w { my %g; %g = () }\nsub r { return $h{a} }\n")
	if d := checkEdit(t, src, bytes.Index(src, []byte("()")), len("()"), "undef"); d.Full || !slices.Equal(d.Subs, []string{"sub w#0"}) {
		t.Errorf("a sub's own hash: damage = %+v, want [sub w#0]", d)
	}
}

// realEdit replaces src[at:at+n] with repl.
type realEdit struct {
	at, n int
	repl  string
}

// realEdits is a sample of up to perKind edits of each kind to src, inside
// its subs: a digit inserted in a number, a space in whitespace, a letter
// after a `$`, which renames a variable; a statement deleted, a literal set
// to undef and to 's', a call renamed, and an assignment's right side set
// to undef.
func realEdits(root *parse.Node, src []byte, perKind int) []realEdit {
	kinds := make([][]realEdit, 8)
	var walk func(n *parse.Node, in bool)
	walk = func(n *parse.Node, in bool) {
		if in {
			switch {
			case n.Kind == parse.Statement:
				kinds[0] = append(kinds[0], realEdit{n.Start, n.End - n.Start, ""})
			case n.Kind == parse.Term && len(n.Children) == 0 && n.End > n.Start && strings.ContainsRune("0123456789'\"", rune(src[n.Start])):
				kinds[1] = append(kinds[1], realEdit{n.Start, n.End - n.Start, "undef"})
				kinds[2] = append(kinds[2], realEdit{n.Start, n.End - n.Start, "'s'"})
				if src[n.Start] >= '0' && src[n.Start] <= '9' {
					kinds[3] = append(kinds[3], realEdit{n.Start + 1, 0, "7"})
				}
			case n.Kind == parse.Term && len(n.Text) > 1 && n.Text[0] == '$' && src[n.Start] == '$':
				kinds[4] = append(kinds[4], realEdit{n.Start + 1, 0, "q"})
			case n.Kind == parse.Trivia && src[n.Start] == ' ':
				kinds[5] = append(kinds[5], realEdit{n.Start, 0, " "})
			case n.Kind == parse.Call && n.Text != "":
				if i := bytes.Index(src[n.Start:n.End], []byte(n.Text)); i >= 0 {
					kinds[6] = append(kinds[6], realEdit{n.Start + i + len(n.Text), 0, "zz"})
				}
			case n.Kind == parse.Binary && n.Text == "=" && len(n.Children) > 1:
				r := n.Children[len(n.Children)-1]
				kinds[7] = append(kinds[7], realEdit{r.Start, r.End - r.Start, "undef"})
			}
		}
		_, sub := lower.AnchorKey(n)
		for _, c := range n.Children {
			walk(c, in || sub)
		}
	}
	walk(root, false)
	var out []realEdit
	for _, k := range kinds {
		m := min(perKind, len(k))
		for j := range m {
			out = append(out, k[j*len(k)/m])
		}
	}
	return out
}

// realFiles is edit5k.pl, and with $PERL5_CORPUS set perl5's lib/*.pm
// and the two files review finding 2 was seen in.
func realFiles() []string {
	files := []string{latencyFixturePath}
	if dir := os.Getenv("PERL5_CORPUS"); dir != "" {
		more, _ := filepath.Glob(filepath.Join(dir, "lib", "*.pm"))
		files = append(files, more...)
		for _, f := range []string{"cpan/Locale-Maketext-Simple/lib/Locale/Maketext/Simple.pm", "cpan/Test-Simple/t/Legacy/use_ok.t"} {
			if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
				files = append(files, filepath.Join(dir, f))
			}
		}
	}
	return files
}

// TestDamageUndamagedSubsAnalyzeUnchanged: on real files, edited in place,
// every node outside the damaged subs has the same type before and after
// the edit, and every diagnostic and symbol outside them is the same,
// shifted; an edit that goes full is not checked. This is the definition's
// soundness check: a sub reported undamaged must be one whose analysis the
// edit cannot change.
func TestDamageUndamagedSubsAnalyzeUnchanged(t *testing.T) {
	checked, files := 0, realFiles()
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		perKind := 4
		if path == latencyFixturePath {
			perKind = 1
		}
		for _, e := range realEdits(parse.Parse(src), src, perKind) {
			t.Run(fmt.Sprintf("%s/%q@%d", filepath.Base(path), e.repl, e.at), func(t *testing.T) {
				if !checkEdit(t, src, e.at, e.n, e.repl).Full {
					checked++
				}
			})
		}
	}
	if checked == 0 {
		t.Fatal("no edit was local; the check checked nothing")
	}
	t.Logf("%d local edits checked over %d files", checked, len(files))
}

// checkEdit replaces src[at:at+n] with repl and returns the damage; when it
// is not full, t fails for every annotation, diagnostic or symbol outside
// the damaged subs that differs between the analyses before and after.
func checkEdit(t *testing.T, src []byte, at, n int, repl string) editDamage {
	t.Helper()
	if at < 0 {
		t.Fatal("the edit's text is not in the source")
	}
	edited := slices.Concat(src[:at], []byte(repl), src[at+n:])
	oldRoot, newRoot := parse.Parse(src), parse.Parse(edited)
	d := damageFromTreeDiff(oldRoot, newRoot)
	if d.Full {
		return d
	}
	before := infer.AnalyzeParsed(oldRoot, nil, infer.Options{})
	after := infer.AnalyzeParsed(newRoot, nil, infer.Options{})
	damaged := map[string]bool{}
	for _, s := range d.Subs {
		damaged[s] = true
	}
	outside := func(id lower.ID) bool {
		top, _, _ := strings.Cut(id.Anchor, "/")
		return !damaged[top]
	}
	for id, typ := range before.Annotations {
		if outside(id) && after.Annotations[id] != typ {
			t.Errorf("damage %v, but %v's type moved %v -> %v", d.Subs, id, typ, after.Annotations[id])
		}
	}
	for id, typ := range after.Annotations {
		if _, ok := before.Annotations[id]; outside(id) && !ok {
			t.Errorf("damage %v, but %v gained type %v", d.Subs, id, typ)
		}
	}
	// keep is what of a side's diagnostics and symbols lies outside its
	// damaged subs, the old side's moved to where the edit put it.
	keep := func(root *parse.Node, starts []uint32, show func(i int, shift int) string, old bool) []string {
		var spans [][2]int
		for _, s := range outerSubs(root) {
			if damaged[s.anchor] {
				spans = append(spans, [2]int{s.node.Start, s.node.End})
			}
		}
		var out []string
	next:
		for i, start := range starts {
			for _, s := range spans {
				if int(start) >= s[0] && int(start) < s[1] {
					continue next
				}
			}
			shift := 0
			if old && int(start) >= at+n {
				shift = len(repl) - n
			}
			out = append(out, show(i, shift))
		}
		return out
	}
	diags := func(fa *infer.FileAnalysis, root *parse.Node, old bool) []string {
		var starts []uint32
		for _, g := range fa.Diagnostics {
			starts = append(starts, g.StartByte)
		}
		return keep(root, starts, func(i, shift int) string {
			g := fa.Diagnostics[i]
			g.StartByte, g.EndByte = uint32(int(g.StartByte)+shift), uint32(int(g.EndByte)+shift)
			return fmt.Sprintf("%+v", g)
		}, old)
	}
	// symbols is the file scope's: after the walk the table's current
	// scope is the root, so Lookup finds what a sub left a file-scoped
	// variable as. A sub's own scope holds shadows carrying the outer
	// declaration's span, which no span test could tell apart.
	symbols := func(fa *infer.FileAnalysis, root *parse.Node, old bool) []string {
		var syms []infer.Symbol
		var starts []uint32
		seen := map[string]bool{}
		for _, s := range fa.Symbols.AllSymbols() {
			if sym, ok := fa.Symbols.Lookup(s.Name); ok && !seen[s.Name] {
				seen[s.Name] = true
				syms, starts = append(syms, sym), append(starts, sym.StartByte)
			}
		}
		out := keep(root, starts, func(i, shift int) string {
			s := syms[i]
			s.StartByte, s.EndByte = uint32(int(s.StartByte)+shift), uint32(int(s.EndByte)+shift)
			return fmt.Sprintf("%+v", s)
		}, old)
		slices.Sort(out)
		return out
	}
	if b, a := diags(before, oldRoot, true), diags(after, newRoot, false); !slices.Equal(b, a) {
		t.Errorf("damage %v, but the diagnostics outside it moved:\n%v\n->\n%v", d.Subs, b, a)
	}
	if b, a := symbols(before, oldRoot, true), symbols(after, newRoot, false); !slices.Equal(b, a) {
		t.Errorf("damage %v, but the symbols outside it moved:\n%v\n->\n%v", d.Subs, b, a)
	}
	return d
}
