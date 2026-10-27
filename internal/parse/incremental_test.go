// ABOUTME: Tests for edit triage (spec 06 §6.5.1) and widening to a safe boundary (§6.5.2).
// ABOUTME: Widening is checked on the edited source: resuming at the anchor and splicing equals a fresh parse.
package parse

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"tamarou.com/pvm/internal/lexer"
)

// at is an edit replacing the first occurrence of old after the first
// occurrence of after in src with text; old == "" inserts there.
func at(src, after, old, text string) Edit {
	i := strings.Index(src, after) + len(after)
	j := i + strings.Index(src[i:], old)
	return Edit{Start: j, OldEnd: j + len(old), NewText: text}
}

func edited(src string, e Edit) string { return src[:e.Start] + e.NewText + src[e.OldEnd:] }

func triageOf(src string, e Edit) Strategy { return triage(Parse([]byte(src)), []byte(src), e) }

func TestIncrementalRepairDefaultsOff(t *testing.T) {
	if IncrementalRepair {
		t.Fatal("IncrementalRepair must stay false until the §6.5.5 flip")
	}
}

const twoSubs = "sub f {\n  my $foo = 1; # note\n  return $foo;\n}\n\nsub g { 2 }\nprint f();\n"

// The gate review's sources: a `use` added in f changes how the last line
// reads X, and a quote-like or regex losing its closer swallows sub g.
const (
	reviewUse   = "sub f {\n  1;\n}\nprint X + 1;\n"
	reviewQuote = "sub f { my $s = q|a|; }\nsub g { 'b|' }\nprint g();\n"
	reviewRegex = "sub f { my $s = m/a/; }\nsub g { 'b/' }\nprint g();\n"
	// resolvedSub calls Foo.pm's `sub bar($)` inside a sub.
	resolvedSub = "use Foo;\nsub g {\n  my @x = (bar 1, 2);\n}\nprint g();\n"
	// reviewPOSIX is the re-review's: under a Session, floor is POSIX's.
	reviewPOSIX = "use POSIX qw(floor);\nsub f { floor 1, 2 }\nprint floor 3, 4;"
)

// TestRepairModelReadsSessionTree: the re-review's edit leaks under a
// Session parse, which a model parsing with no resolver accepts and the
// resolver-aware one rejects through the parser-state check.
func TestRepairModelReadsSessionTree(t *testing.T) {
	src := []byte(reviewPOSIX)
	// `floor 1, if }`: under the Session, floor is unary and the `if` after
	// the comma reads past the sub's `}`. The re-review's `floor --, 2` no
	// longer leaks, since a catch-all's skip stops before that `}` (§6.8.3).
	e := at(reviewPOSIX, "floor ", "2", "if")
	newSrc := []byte(edited(reviewPOSIX, e))
	old := sessionParse(src)
	if got := triage(old, src, e); got != StrategyLocal {
		t.Fatalf("triage %v, want StrategyLocal", got)
	}
	_, a := widen(old, src, e)
	delta := len(e.NewText) - (e.OldEnd - e.Start)
	if leak(old, sessionParse(newSrc), a.Start, a.End, delta) == "" {
		t.Fatal("no leak under a Session parse; the case checks nothing")
	}
	if !repairAccepts(src, newSrc, a.Start, a.End, delta, func() *resolver { return nil }) {
		t.Fatal("the no-resolver model rejects it; the case no longer shows why the model needs one")
	}
	if repairAccepts(src, newSrc, a.Start, a.End, delta, incResolver) {
		t.Fatal("the Session model accepts an edit whose tail does not match the fresh Session parse")
	}
	checkNoLeak(t, reviewPOSIX, e)
}

// TestTriageScansDeletedText: a trigger leaving the source is as structural
// as one arriving, so the removed side is scanned as well as the new.
func TestTriageScansDeletedText(t *testing.T) {
	for _, c := range []struct{ name, src, del string }{
		{"heredoc", "sub f {\n  print <<EOF;\nbody\nEOF\n  1;\n}\n", "<<EOF"},
		{"__END__", "sub f {\n  1;\n}\n__END__\ndata\n", "__END__"},
		{"=cut", "sub f { 1 }\n=pod\n\nx\n\n=cut\n\nsub g { 2 }\n", "=cut"},
		{"sub", "sub f {\n  my $x = sub { 1 };\n}\n", "sub "},
		{"use", "sub f {\n  1;\n}\nuse strict;\n", "use "},
		{"BEGIN", "sub f {\n  1;\n}\nBEGIN { 1 }\n", "BEGIN"},
		{"#", "sub f {\n  my $x = 1; # 2\n}\n", "#"},
	} {
		t.Run(c.name, func(t *testing.T) {
			e := at(c.src, "", c.del, "")
			if got := triageOf(c.src, e); got != StrategyFull {
				t.Errorf("deleting %q from %q: %v, want StrategyFull", c.del, c.src, got)
			}
		})
	}
	// The green half: a deletion holding no trigger stays local.
	if got := triageOf(twoSubs, at(twoSubs, "my $", "foo", "")); got != StrategyLocal {
		t.Errorf("deleting a variable name: %v, want StrategyLocal", got)
	}
}

// TestTriageMissedTriggers: the §6.5.1 sketch's appear-or-disappear test
// misses a prototype edit (`sub` stays in the line) and `=pod` (only `=cut`
// is listed). Both are StrategyFull, as are edits in a sub's head, a
// `package NAME;`, a `use` and a BEGIN block, and inside a block that
// declares a sub.
func TestTriageMissedTriggers(t *testing.T) {
	for _, c := range []struct {
		name, src string
		e         Edit
	}{
		{"prototype", "sub f() { 1 }\nprint f @a;\n", at("sub f() { 1 }\nprint f @a;\n", "sub f(", "", `\@`)},
		{"add prototype", "sub f { 1 }\nprint f @a;\n", at("sub f { 1 }\nprint f @a;\n", "sub f", "", "()")},
		{"rename sub", twoSubs, at(twoSubs, "sub ", "f", "h")},
		{"=pod", twoSubs, at(twoSubs, "1; # note\n", "", "=pod\n")},
		{"=head1", twoSubs, at(twoSubs, "1; # note\n", "", "=head1 X\n")},
		{"package NAME;", "package P;\nsub f { 1 }\n", at("package P;\nsub f { 1 }\n", "package ", "P", "Q")},
		{"use args", "use Foo qw(bar);\nbar 1;\n", at("use Foo qw(bar);\nbar 1;\n", "qw(", "bar", "baz")},
		{"BEGIN body", "BEGIN { 1 }\nsub f { 2 }\n", at("BEGIN { 1 }\nsub f { 2 }\n", "{ ", "1", "3")},
		{"prototype in package block", "package Q { sub x { 1 } }\nQ::x 1, 2;\n", at("package Q { sub x { 1 } }\nQ::x 1, 2;\n", "sub x", "", "($)")},
		{"prototype in bare block", "{ sub y { 1 } }\ny 1, 2;\n", at("{ sub y { 1 } }\ny 1, 2;\n", "sub y", "", "($)")},
		// Gate review of #512: spellings a substring trigger misses.
		{"use and a tab", reviewUse, at(reviewUse, "sub f {\n", "", "use\tconstant X => 1;\n")},
		{"use and a newline", reviewUse, at(reviewUse, "sub f {\n", "", "use\nconstant X => 1;\n")},
		{"require", reviewUse, at(reviewUse, "sub f {\n", "", "require\tFoo;\n")},
		{"q| closer deleted", reviewQuote, at(reviewQuote, "q|a", "|", "")},
		{"regex closer deleted", reviewRegex, at(reviewRegex, "m/a", "/", "")},
		{"}{ inserted", twoSubs, at(twoSubs, "1;", "", "}{")},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := triageOf(c.src, c.e); got != StrategyFull {
				t.Errorf("edit %+v of %q: %v, want StrategyFull", c.e, c.src, got)
			}
		})
	}
}

// TestTriageNewlineEndingCommentNotReuseAll: StrategyReuseAll is an edit
// that changes one comment, POD or whitespace token and no other, leaving it
// one token of the same kind. A newline in a comment ends it, and whitespace
// beside a token rather than inside one can join or split tokens.
func TestTriageNewlineEndingCommentNotReuseAll(t *testing.T) {
	const str = "sub f { my $s = \"ab\"; }\n"
	for _, c := range []struct {
		name, src string
		e         Edit
		reuse     bool
	}{
		{"newline in comment", twoSubs, at(twoSubs, "# no", "", "\n"), false},
		{"CRLF in comment", twoSubs, at(twoSubs, "# no", "", "\r\n"), false},
		{"space beside a token", twoSubs, at(twoSubs, "my", "", " "), false},
		{"space where there was none", twoSubs, at(twoSubs, "$foo", "", " "), false},
		{"space at whitespace start", twoSubs, at(twoSubs, "sub f {", "", " "), false},
		{"text in a string", str, at(str, "\"a", "", "x"), false},
		// The green half.
		{"text in comment", twoSubs, at(twoSubs, "# no", "", "x"), true},
		{"space inside whitespace", twoSubs, at(twoSubs, "f {\n ", "", " "), true},
		{"tab inside whitespace", twoSubs, at(twoSubs, "f {\n ", "", "\t"), true},
		{"newline inside whitespace", twoSubs, at(twoSubs, "f {\n ", "", "\n"), true},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := triageOf(c.src, c.e); (got == StrategyReuseAll) != c.reuse {
				t.Errorf("edit %+v: %v, want reuse-all %v", c.e, got, c.reuse)
			}
		})
	}
}

// TestTriageSizeLimitBoundaries: each limit's last byte or line is on the
// local side and the next one over is not.
func TestTriageSizeLimitBoundaries(t *testing.T) {
	big := "sub f {\n  my $x = '" + strings.Repeat("x", 70*1024) + "';\n}\n"
	in := func(text string) Edit { return at(twoSubs, "my $foo = 1;", "", text) }
	for _, c := range []struct {
		name, src string
		e         Edit
		want      Strategy
	}{
		{"insert 1024 bytes", twoSubs, in(strings.Repeat("x", maxLocalEdit)), StrategyLocal},
		{"insert 1025 bytes", twoSubs, in(strings.Repeat("x", maxLocalEdit+1)), StrategyFull},
		{"delete 1024 bytes", big, at(big, "'", strings.Repeat("x", maxLocalEdit), ""), StrategyLocal},
		{"delete 1025 bytes", big, at(big, "'", strings.Repeat("x", maxLocalEdit+1), ""), StrategyFull},
		{"10 newlines", twoSubs, in(strings.Repeat("x\n", maxLocalLines)), StrategyLocal},
		{"11 newlines", twoSubs, in(strings.Repeat("x\n", maxLocalLines+1)), StrategyFull},
		{"insert over 64 KiB", twoSubs, in(strings.Repeat("x", maxEditBytes+1)), StrategyFull},
		{"delete over 64 KiB", big, at(big, "'", strings.Repeat("x", maxEditBytes+1), ""), StrategyFull},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := triageOf(c.src, c.e); got != c.want {
				t.Errorf("%v, want %v", got, c.want)
			}
		})
	}
}

// TestTriageLineWindowAtBufferEdges: triage compares token streams, so the
// span it finds changed starts and ends on token boundaries, never mid-rune
// in a valid source, and an edit at either edge of a buffer does not panic.
func TestTriageLineWindowAtBufferEdges(t *testing.T) {
	const crlf = "sub f {\r\n  my $x = 1;\r\n}\r\n"
	const mb = "sub f {\n  my $s = \"é€\"; # ü\n}\n"
	e := strings.Index(mb, "é")
	for _, c := range []struct {
		name, src string
		e         Edit
	}{
		{"empty buffer", "", Edit{NewText: "my $x;"}},
		{"empty buffer, empty edit", "", Edit{}},
		{"offset 0", twoSubs, Edit{NewText: "#"}},
		{"delete from 0", twoSubs, Edit{OldEnd: 3}},
		{"at EOF", twoSubs, Edit{Start: len(twoSubs), OldEnd: len(twoSubs), NewText: "1;"}},
		{"delete to EOF", twoSubs, Edit{Start: len(twoSubs) - 3, OldEnd: len(twoSubs)}},
		{"CRLF line", crlf, at(crlf, "my $", "x", "y")},
		{"between CR and LF", crlf, at(crlf, "1;\r", "", "z")},
		{"before a multi-byte rune", mb, Edit{Start: e, OldEnd: e, NewText: "a"}},
		{"after a multi-byte rune", mb, Edit{Start: e + 2, OldEnd: e + 2, NewText: "a"}},
		{"replace a multi-byte rune", mb, Edit{Start: e, OldEnd: e + 2, NewText: "€"}},
		{"in a multi-byte comment", mb, at(mb, "# ", "ü", "ö")},
	} {
		t.Run(c.name, func(t *testing.T) {
			d := diffTokens([]byte(c.src), c.e)
			for _, w := range [][]byte{spanOf(d.oldChanged(), d.oldSrc), spanOf(d.newChanged(), d.newSrc)} {
				if !utf8.Valid(w) {
					t.Errorf("changed span %q is not valid UTF-8", w)
				}
			}
			triageOf(c.src, c.e)
		})
	}
}

// widenCase is a source and an edit triage reads as StrategyLocal, with the
// anchor widening must choose: the top-level statement starting with
// anchor, or "" for the whole file.
type widenCase struct {
	name, src string
	e         Edit
	anchor    string
}

const blocks = "package Q { sub x { 1 } }\nQ::x 1, 2;\n{ sub y { 1 } }\nif (1) { sub z { 1 } }\n" + twoSubs

var widenCases = []widenCase{
	{"in a sub body", twoSubs, at(twoSubs, "my $", "foo", "bar"), "sub f"},
	{"split a token", twoSubs, at(twoSubs, "my $f", "", "|"), "sub f"},
	{"in the second sub", twoSubs, at(twoSubs, "sub g { ", "2", "3"), "sub g"},
	{"nested sub's body", "sub f { sub n { 1 } 2 }\n", at("sub f { sub n { 1 } 2 }\n", "sub n { ", "1", "4"), "sub f"},
	// A nested resume stops at the block's `}` with no post-block state, so
	// a sub inside a package block, bare block or `if` widens past it.
	{"sub in package block", blocks, at(blocks, "sub x { ", "1", "5"), "package Q"},
	{"sub in bare block", blocks, at(blocks, "sub y { ", "1", "6"), "{ sub y"},
	{"sub in if block", blocks, at(blocks, "sub z { ", "1", "7"), "if (1)"},
	{"sub after the blocks", blocks, at(blocks, "my $", "foo", "baz"), "sub f"},
	// Edits that never reach a sub's body widen to the whole file.
	{"top-level statement", twoSubs, at(twoSubs, "print ", "f", "g"), ""},
	{"between subs", twoSubs, at(twoSubs, "}\n", "", ";"), ""},
	{"package block's own code", "package Q { my $v = 1; }\n", at("package Q { my $v = 1; }\n", "= ", "1", "2"), ""},
	{"bare block's own code", blocks, at(blocks, "Q::x ", "1", "8"), ""},
}

// TestWidenToSafeBoundary: widening ends on a safe lexer checkpoint at or
// before the anchor, and the anchor is the whole file or a top-level
// statement holding a parser snapshot. Resuming there over the EDITED
// source and splicing equals a fresh parse, so the anchor is one Repair can
// restart from.
func TestWidenToSafeBoundary(t *testing.T) {
	for _, c := range widenCases {
		t.Run(c.name, func(t *testing.T) {
			old, src := Parse([]byte(c.src)), []byte(c.src)
			if got := triage(old, src, c.e); got != StrategyLocal {
				t.Fatalf("triage %v, want StrategyLocal: widening applies to local edits", got)
			}
			cp, anchor := widen(old, src, c.e)
			checkAnchor(t, c, old, anchor)
			_, cps := lexer.Checkpoints(src, lexer.StartStateBarewords(coreBarewordOperators()))
			if !slices.ContainsFunc(cps, func(k lexer.Checkpoint) bool { return k.Byte == cp.Byte && k.State.Equal(cp.State) }) {
				t.Fatalf("checkpoint at %d is not one the lexer recorded", cp.Byte)
			}
			if cp.Byte > anchor.Start {
				t.Fatalf("checkpoint %d after anchor start %d", cp.Byte, anchor.Start)
			}
			newSrc := []byte(edited(c.src, c.e))
			full := lexer.TokenizeBarewords(newSrc, coreBarewordOperators())
			i := slices.IndexFunc(full, func(t lexer.Token) bool { return t.Start >= cp.Byte })
			if i < 0 {
				i = len(full)
			}
			if got := lexer.Resume(newSrc, cp); !slices.Equal(got, full[i:]) {
				t.Fatalf("resuming the lexer at %d over the edited source diverges from a full lex", cp.Byte)
			}
			if anchor == old {
				return
			}
			prefix, snaps := unfinished(src, nil)
			k := slices.IndexFunc(snaps, func(s parserSnapshot) bool { return s.at == anchor.Start && !s.nested })
			if k < 0 {
				t.Fatalf("no top-level parser snapshot at the anchor's start %d", anchor.Start)
			}
			p, tail := snaps[k].resume(newSrc, nil)
			if diff := TreeDiff(spliceTop(p, prefix, anchor.Start, tail), Parse(newSrc)); diff != "" {
				t.Fatalf("resumed at the anchor over the edited source != fresh parse:\n%s", diff)
			}
		})
	}
}

// checkAnchor asserts anchor is the case's: the root, or the top-level
// statement starting with c.anchor.
func checkAnchor(t *testing.T, c widenCase, root, anchor *Node) {
	t.Helper()
	if c.anchor == "" {
		if anchor != root {
			t.Fatalf("anchor %q, want the whole file", c.src[anchor.Start:anchor.End])
		}
		return
	}
	if !slices.Contains(root.Children, anchor) || !strings.HasPrefix(c.src[anchor.Start:], c.anchor) {
		t.Fatalf("anchor %q, want the top-level statement starting %q", c.src[anchor.Start:anchor.End], c.anchor)
	}
}

// TestWidenTopLevelEditFallsBackToSourceFile: the file top level, a `use`,
// a BEGIN block and a `package NAME;` are never anchors, even around a sub.
func TestWidenTopLevelEditFallsBackToSourceFile(t *testing.T) {
	const src = "use Foo qw(bar);\npackage P;\nBEGIN { sub b { 1 } }\nmy $t = 1;\n" + twoSubs
	for _, c := range []struct {
		src string
		e   Edit
	}{
		{src, at(src, "my $t = ", "1", "2")},
		{src, at(src, "}\n", "", "\n")},
		{src, at(src, "use ", "Foo", "Fob")},
		{src, at(src, "qw(", "bar", "baz")},
		{src, at(src, "package ", "P", "R")},
		{src, at(src, "BEGIN { ", "", "1; ")},
		{src, at(src, "sub b { ", "1", "2")},
		{twoSubs, at(twoSubs, "print ", "", "1, ")},
	} {
		root := Parse([]byte(c.src))
		cp, anchor := widen(root, []byte(c.src), c.e)
		if anchor != root || cp.Byte != 0 {
			t.Errorf("edit %+v of %q: anchor %q at checkpoint %d, want the whole file from 0", c.e, c.src, c.src[anchor.Start:anchor.End], cp.Byte)
		}
	}
	// The green half: the same file's sub body is an anchor.
	root := Parse([]byte(src))
	if _, anchor := widen(root, []byte(src), at(src, "my $", "foo", "x")); anchor == root {
		t.Error("an edit in a top-level sub body widened to the whole file")
	}
}

// TestWidenResolvedPrototypeCallKeepsWholeCall: on the tree a Session
// parsed with its resolver, where bar is a call Foo.pm's `sub bar($)`
// shapes, an edit in its arguments widens to a region holding the call.
func TestWidenResolvedPrototypeCallKeepsWholeCall(t *testing.T) {
	path := filepath.Join("testdata", "incremental", "edit.pl")
	lib := filepath.Join("testdata", "incremental", "lib")
	for _, src := range []string{
		"use Foo; my @x = (bar 1, 2);\n",
		"use Foo;\nsub g {\n  my @x = (bar 1, 2);\n}\n",
	} {
		old := NewSession().ParseSourceFrom(path, []byte(src), lib)
		call := findCall(old, "bar")
		if call == nil || !call.Resolved || call.Prototype != "($)" {
			t.Fatalf("%q: bar is not a Resolved call with Prototype ($) on the Session tree", src)
		}
		e := at(src, "bar ", "1", "42")
		if got := triage(old, []byte(src), e); got != StrategyLocal {
			t.Fatalf("%q: triage %v, want StrategyLocal", src, got)
		}
		_, anchor := widen(old, []byte(src), e)
		if anchor.Start > call.Start || call.End > anchor.End {
			t.Errorf("%q: anchor [%d, %d) does not hold the call [%d, %d)", src, anchor.Start, anchor.End, call.Start, call.End)
		}
	}
}

// TestWidenCheckpointNotAfterAnchorStart: a lexer checkpoint falls only
// every 512 bytes, so in a file of many small subs most anchors start after
// the last checkpoint and widening must step back to it, never forward.
func TestWidenCheckpointNotAfterAnchorStart(t *testing.T) {
	var b strings.Builder
	for i := range 200 {
		b.WriteString("sub s" + strings.Repeat("x", i%7) + " {\n  my $v = 1;\n}\n")
	}
	src := b.String()
	root := Parse([]byte(src))
	stepped := false
	for i := 0; i < len(src); i += 37 {
		e := Edit{Start: i, OldEnd: i, NewText: "1"}
		cp, anchor := widen(root, []byte(src), e)
		if cp.Byte > anchor.Start {
			t.Fatalf("edit at %d: checkpoint %d after anchor start %d", i, cp.Byte, anchor.Start)
		}
		stepped = stepped || cp.Byte > 0 && cp.Byte < anchor.Start
	}
	if !stepped {
		t.Fatal("no anchor started after a non-zero checkpoint; the test checks nothing")
	}
}

// leak reports how fresh, the parse of old's source after an edit shifting
// later bytes by delta, differs from old outside old's top-level span [lo,
// hi): each top-level statement before it must be old's, and each after it
// old's shifted by delta, or a splice of [lo, hi) alone is wrong. "" when
// nothing leaks. This is what Repair relies on triage and widen for.
func leak(old, fresh *Node, lo, hi, delta int) string {
	var ob, oa, fb, fa []*Node
	for _, c := range old.Children {
		if c.End <= lo {
			ob = append(ob, c)
		} else if c.Start >= hi {
			oa = append(oa, c)
		}
	}
	for _, c := range fresh.Children {
		if c.End <= lo {
			fb = append(fb, c)
		} else if c.Start >= hi+delta {
			fa = append(fa, c)
		}
	}
	if len(ob) != len(fb) || len(oa) != len(fa) {
		return fmt.Sprintf("statements outside the anchor: %d before and %d after, fresh has %d and %d", len(ob), len(oa), len(fb), len(fa))
	}
	for i := range ob {
		if d := TreeDiff(ob[i], fb[i]); d != "" {
			return "before the anchor: " + d
		}
	}
	for i := range oa {
		if d := TreeDiffShifted(oa[i], fa[i], delta); d != "" {
			return "after the anchor: " + d
		}
	}
	return ""
}

// reuseSpan is the span of old a splice re-parses for e, and whether triage
// lets any of old be reused: widen's anchor for StrategyLocal, the
// top-level statement holding the edit for StrategyReuseAll.
func reuseSpan(old *Node, s Strategy, e Edit) (lo, hi int, reused bool) {
	switch s {
	case StrategyLocal:
		a := anchorOf(old, e)
		return a.Start, a.End, a != old
	case StrategyReuseAll:
		if c := childHolding(old, e); c != nil {
			return c.Start, c.End, true
		}
	}
	return 0, 0, false
}

// incResolver is a fresh resolver as NewSession().ParseSourceFrom(path,
// src, lib) builds one, for the fuzz harness's path and root: the model
// parses as the Session triage's contract reads, not as a bare Parse.
func incResolver() *resolver {
	dir := filepath.Join("testdata", "incremental")
	return &resolver{load: DirLoader(dir, filepath.Join(dir, "lib")), seen: map[string]bool{}, facts: map[string]moduleFacts{}}
}

// sessionParse is the Session tree of src under incResolver.
func sessionParse(src []byte) *Node { return parseRoot(src, incResolver()) }

// repairAccepts models the check §6.5.2 leaves to Repair for a local edit:
// one statement re-parsed at the anchor [lo, hi) over the edited source
// leaves the next unconsumed token starting at the anchor's shifted end (or
// at EOF), and leaves the parser in the state the old parse had after the
// anchor (its token index moved by the token count's change). Both parses
// use res, as Repair would its Session's. A parse error can swallow the
// anchor's `}` with every token unchanged, which no token comparison sees.
func repairAccepts(src, newSrc []byte, lo, hi, delta int, res func() *resolver) bool {
	_, snaps := unfinished(src, res())
	var top []parserSnapshot
	for _, s := range snaps {
		if !s.nested {
			top = append(top, s)
		}
	}
	k := slices.IndexFunc(top, func(s parserSnapshot) bool { return s.at == lo })
	if k < 0 {
		return false
	}
	p := top[k].restore(newSrc, res())
	p.statement()
	end := len(newSrc)
	if p.pos < len(p.toks) {
		end = p.toks[p.pos].Start
	}
	if end != hi+delta {
		return false
	}
	if k+1 == len(top) {
		return true
	}
	want := top[k+1].p.clone()
	want.pos += len(p.toks) - len(top[k].p.toks)
	return sameState(p, &want)
}

// checkNoLeak asserts that whatever triage lets a splice reuse, once Repair's
// own check passes, equals the fresh parse of the edited source.
func checkNoLeak(t *testing.T, src string, e Edit) {
	t.Helper()
	old := sessionParse([]byte(src))
	s := triage(old, []byte(src), e)
	lo, hi, reused := reuseSpan(old, s, e)
	if !reused {
		return
	}
	delta := len(e.NewText) - (e.OldEnd - e.Start)
	newSrc := []byte(edited(src, e))
	if s == StrategyLocal && !repairAccepts([]byte(src), newSrc, lo, hi, delta, incResolver) {
		return // Repair returns ok=false and parses the whole file
	}
	if d := leak(old, sessionParse(newSrc), lo, hi, delta); d != "" {
		t.Fatalf("%v edit %+v of %q leaks past [%d, %d): %s", s, e, src, lo, hi, d)
	}
}

// TestTriageLocalReusesShiftedTail: for every local case above, and local
// edits of the gate review's sources, the statements outside the anchor are
// reusable, and Repair's check accepts them, so the reuse is exercised
// rather than skipped.
func TestTriageLocalReusesShiftedTail(t *testing.T) {
	cases := slices.Clone(widenCases)
	for _, c := range []widenCase{
		{"quote-like body", reviewQuote, at(reviewQuote, "q|", "a", "ab"), "sub f"},
		{"regex body", reviewRegex, at(reviewRegex, "m/", "a", "b+"), "sub f"},
		{"use source's sub", reviewUse, at(reviewUse, "  ", "1", "2"), "sub f"},
		{"resolved call in a sub", resolvedSub, at(resolvedSub, "bar ", "1", "42"), "sub g"},
	} {
		cases = append(cases, c)
	}
	for _, c := range cases {
		checkNoLeak(t, c.src, c.e)
		old, delta := sessionParse([]byte(c.src)), len(c.e.NewText)-(c.e.OldEnd-c.e.Start)
		if got := triage(old, []byte(c.src), c.e); got != StrategyLocal {
			t.Errorf("%s: triage %v, want StrategyLocal", c.name, got)
		}
		if _, a := widen(old, []byte(c.src), c.e); a != old && !repairAccepts([]byte(c.src), []byte(edited(c.src, c.e)), a.Start, a.End, delta, incResolver) {
			t.Errorf("%s: Repair's check rejects a plain local edit", c.name)
		}
	}
	// The green half: leak sees the gate review's leak when handed it.
	e := at(reviewQuote, "q|a", "|", "")
	old, fresh := Parse([]byte(reviewQuote)), Parse([]byte(edited(reviewQuote, e)))
	if leak(old, fresh, old.Children[0].Start, old.Children[0].End, -1) == "" {
		t.Fatal("leak reports nothing for q| losing its closer; the check checks nothing")
	}
}

// sweepInserts are the texts the sweep inserts: quote and bracket halves,
// heredoc, POD, comment and declaration openers, and harmless text.
var sweepInserts = []string{
	`"`, `'`, "`", "|", "/", "q|", "s{", "m/", "{", "}", "}{", "(", ")", "[", "]",
	"#", "\n", "<<E", "<<~E\n", "=pod\n", "=cut\n", "__END__\n", "sub ", "sub h;",
	"use\tconstant X => 1;", "no\nstrict;", "package\tP;", "BEGIN{", "format=\n",
	"x", " ", ";", "$", "@", "%", "&", "->", "1", "my $v;", "our @w;", ",",
}

// TestTriageSweepNoLeak is the gate reviewer's random sweep: random edits
// of the seed sources, none of which may leak past what triage reuses.
func TestTriageSweepNoLeak(t *testing.T) {
	srcs := []string{twoSubs, blocks, reviewUse, reviewQuote, reviewRegex,
		"sub f {\n  print <<EOF;\nbody\nEOF\n  return 1;\n}\nsub g { 2 }\nprint f();\n"}
	for _, s := range incrementalSeedSources {
		srcs = append(srcs, s)
	}
	slices.Sort(srcs)
	r := rand.New(rand.NewSource(512))
	for _, src := range srcs {
		for range 400 {
			start := r.Intn(len(src) + 1)
			end := start + min(r.Intn(4), len(src)-start)
			text := ""
			if r.Intn(4) > 0 {
				text = sweepInserts[r.Intn(len(sweepInserts))]
			}
			checkNoLeak(t, src, Edit{Start: start, OldEnd: end, NewText: text})
		}
	}
}

// FuzzTriage holds the sweep's property for any source and edit:
//
//	go test ./internal/parse/ -run '^$' -fuzz '^FuzzTriage$' -fuzzminimizetime 5s
//
// Each input lexes and parses its source several times, so the fuzzer
// minimizing a new input at the default 60s of minimization reports 0
// execs/s while it does. A short -fuzzminimizetime, and sources capped at
// 1 KiB, keep it exploring.
func FuzzTriage(f *testing.F) {
	f.Add(reviewQuote, 18, 1, "")
	f.Add(reviewUse, 8, 0, "use\tconstant X => 1;\n")
	f.Add(twoSubs, 20, 0, "}{")
	f.Fuzz(func(t *testing.T, src string, offset, delLen int, insert string) {
		if len(src) > 1024 || len(insert) > 64 {
			t.Skip()
		}
		start := min(max(offset, 0), len(src))
		checkNoLeak(t, src, Edit{Start: start, OldEnd: start + min(max(delLen, 0), len(src)-start), NewText: insert})
	})
}
