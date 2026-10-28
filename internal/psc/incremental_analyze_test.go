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
	same := []byte("my $p = undef; sub r { return $p eq 'n' } sub w { $p = 'n'; return 1 }\n")
	edited, _ = replaceOnce(t, same, "sub w", "return 1", "return 2")
	if d := damageOf(same, edited); d.Full || !slices.Equal(d.Subs, []string{"sub w#0"}) {
		t.Errorf("an edit keeping w's assignments: damage = %+v, want [sub w#0]", d)
	}
}

// realEdits is a sample of edits to src, each one byte inserted at an
// offset: a digit after a digit of a number, a space after a whitespace
// byte, and a letter after a `$` sigil, which renames a variable.
func realEdits(root *parse.Node, src []byte, perKind int) [][2]int {
	var digits, spaces, sigils []int
	for i := 0; i < len(src); i++ {
		leaf := leafAt(root, i)
		switch {
		case leaf.Kind == parse.Term && src[i] >= '0' && src[i] <= '9':
			digits = append(digits, i)
		case leaf.Kind == parse.Trivia && (src[i] == ' ' || src[i] == '\n'):
			spaces = append(spaces, i)
		case leaf.Kind == parse.Term && src[i] == '$' && i+1 < len(src) && src[i+1] >= 'a' && src[i+1] <= 'z':
			sigils = append(sigils, i)
		}
	}
	var edits [][2]int
	for b, at := range [][]int{digits, spaces, sigils} {
		for k := range min(perKind, len(at)) {
			edits = append(edits, [2]int{at[k*len(at)/min(perKind, len(at))] + 1, b})
		}
	}
	return edits
}

// TestDamageUndamagedSubsAnalyzeUnchanged: on real files, edited in place,
// every node outside the damaged subs has the same type before and after
// the edit and every diagnostic outside them is the same, shifted; an edit
// that goes full is not checked. This is the definition's soundness check:
// a sub reported undamaged must be one whose analysis the edit cannot change.
func TestDamageUndamagedSubsAnalyzeUnchanged(t *testing.T) {
	files := []string{latencyFixturePath}
	if dir := os.Getenv("PERL5_CORPUS"); dir != "" {
		more, _ := filepath.Glob(filepath.Join(dir, "lib", "*.pm"))
		files = append(files, more...)
	}
	insert := [][]byte{[]byte("7"), []byte(" "), []byte("q")}
	checked := 0
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		root := parse.Parse(src)
		perKind := 4
		if path == latencyFixturePath {
			perKind = 2
		}
		before := infer.AnalyzeParsed(root, nil, infer.Options{})
		for _, e := range realEdits(root, src, perKind) {
			at := e[0]
			edited := slices.Concat(src[:at], insert[e[1]], src[at:])
			newRoot := parse.Parse(edited)
			d := damageFromTreeDiff(root, newRoot)
			if d.Full {
				continue
			}
			checked++
			after := infer.AnalyzeParsed(newRoot, nil, infer.Options{})
			name := fmt.Sprintf("%s +%q@%d", filepath.Base(path), insert[e[1]], at)
			checkUndamagedUnchanged(t, name, d, root, newRoot, before, after, at)
		}
	}
	if checked == 0 {
		t.Fatal("no edit was local; the check checked nothing")
	}
	t.Logf("%d local edits checked over %d files", checked, len(files))
}

// checkUndamagedUnchanged fails t when an annotation or diagnostic outside
// d's subs differs between before and after, for a 1-byte insert at at.
func checkUndamagedUnchanged(t *testing.T, name string, d editDamage, oldRoot, newRoot *parse.Node, before, after *infer.FileAnalysis, at int) {
	t.Helper()
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
			t.Errorf("%s: damage %v, but %v's type moved %v -> %v", name, d.Subs, id, typ, after.Annotations[id])
		}
	}
	for id, typ := range after.Annotations {
		if _, ok := before.Annotations[id]; outside(id) && !ok {
			t.Errorf("%s: damage %v, but %v gained type %v", name, d.Subs, id, typ)
		}
	}
	spans := func(root *parse.Node) [][2]int {
		var out [][2]int
		for _, s := range outerSubs(root) {
			if damaged[s.anchor] {
				out = append(out, [2]int{s.node.Start, s.node.End})
			}
		}
		return out
	}
	keep := func(diags []infer.Diagnostic, in [][2]int, shift bool) []string {
		var out []string
	next:
		for _, g := range diags {
			for _, s := range in {
				if int(g.StartByte) >= s[0] && int(g.StartByte) < s[1] {
					continue next
				}
			}
			if shift && int(g.StartByte) >= at {
				g.StartByte++
				g.EndByte++
			}
			out = append(out, fmt.Sprintf("%+v", g))
		}
		return out
	}
	if b, a := keep(before.Diagnostics, spans(oldRoot), true), keep(after.Diagnostics, spans(newRoot), false); !slices.Equal(b, a) {
		t.Errorf("%s: damage %v, but the diagnostics outside it moved:\n%v\n->\n%v", name, d.Subs, b, a)
	}
}
