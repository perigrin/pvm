// ABOUTME: Tests for Repair (spec 06 §6.5.3): a repaired tree equals a fresh parse, or Repair returns ok=false and no tree.
// ABOUTME: Also error subtrees, byte shifts, edit5k.pl locality, the old tree untouched, buffer edges and bad input.
package parse

import (
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
)

// The Session form resolves as the fuzz harness does: edit.pl's directory,
// then lib, which holds Foo.pm and its `sub bar($)`.
var (
	repPath = filepath.Join("testdata", "incremental", "edit.pl")
	repRoot = filepath.Join("testdata", "incremental", "lib")
)

func sessionTree(src []byte) *Node { return NewSession().ParseSourceFrom(repPath, src, repRoot) }

// checkRepair asserts §6.5.4's property for e of src in both forms, and
// again for the edit undoing it from the repaired tree, which resumes from
// the snapshots that tree carries. It returns the Session form's ok.
func checkRepair(t *testing.T, src string, e Edit) bool {
	t.Helper()
	newSrc := []byte(edited(src, e))
	tree, ok := NewSession().Repair(sessionTree([]byte(src)), e, repPath, newSrc, repRoot)
	checkRepaired(t, "Session", src, e, tree, ok, sessionTree(newSrc))
	plain, pok := Repair(Parse([]byte(src)), e, newSrc)
	checkRepaired(t, "plain", src, e, plain, pok, Parse(newSrc))
	if ok {
		undo := Edit{Start: e.Start, OldEnd: e.Start + len(e.NewText), NewText: src[e.Start:e.OldEnd]}
		back, bok := NewSession().Repair(tree, undo, repPath, []byte(src), repRoot)
		checkRepaired(t, "Session undo", string(newSrc), undo, back, bok, sessionTree([]byte(src)))
	}
	return ok
}

// checkRepaired is the property for one result: ok and equal to fresh, with
// every offset inside the source, or not ok and no tree.
func checkRepaired(t *testing.T, form, src string, e Edit, tree *Node, ok bool, fresh *Node) {
	t.Helper()
	if !ok {
		if tree != nil {
			t.Fatalf("%s: edit %+v of %q: ok=false with a tree", form, e, src)
		}
		return
	}
	if d := TreeDiff(tree, fresh); d != "" {
		t.Fatalf("%s: edit %+v of %q: repaired != fresh:\n%s", form, e, src, d)
	}
	if bad := badOffset(tree, fresh.End); bad != nil {
		t.Fatalf("%s: edit %+v of %q: node [%d, %d) outside [0, %d]", form, e, src, bad.Start, bad.End, fresh.End)
	}
}

// badOffset is a node under n not within 0 <= Start <= End <= size, or nil.
func badOffset(n *Node, size int) *Node {
	if n.Start < 0 || n.Start > n.End || n.End > size {
		return n
	}
	for _, c := range slices.Concat(n.Children, n.InnerHeredocBodies) {
		if b := badOffset(c, size); b != nil {
			return b
		}
	}
	return nil
}

func TestRepairNeverReusesErrorSubtree(t *testing.T) {
	for _, c := range []struct {
		name, src string
		e         Edit
	}{
		// The statement after the anchor holds an error: not reused.
		{"error after the anchor", "sub f { 1 }\nmy $x = ;\n", at("sub f { 1 }\nmy $x = ;\n", "{ ", "1", "2")},
		{"error before the anchor", "my $x = ;\nsub f { 1 }\n", at("my $x = ;\nsub f { 1 }\n", "{ ", "1", "2")},
		// The anchor gains an error, or loses one.
		{"anchor gains an error", "sub f { my $x = 1; }\nprint 1;\n", at("sub f { my $x = 1; }\nprint 1;\n", "= ", "1", "")},
		{"anchor loses an error", "sub f { my $x = ; }\nprint 1;\n", at("sub f { my $x = ; }\nprint 1;\n", "= ", "", "1")},
	} {
		t.Run(c.name, func(t *testing.T) {
			old, src := sessionTree([]byte(c.src)), []byte(c.src)
			if s := triage(old, src, c.e); s != StrategyLocal {
				t.Fatalf("triage %v, want StrategyLocal: the case must reach Repair's own rule", s)
			}
			newSrc := []byte(edited(c.src, c.e))
			if !hasError(old) && !hasError(sessionTree(newSrc)) {
				t.Fatal("neither tree holds an error; the case checks nothing")
			}
			if tree, ok := NewSession().Repair(old, c.e, repPath, newSrc, repRoot); ok || tree != nil {
				t.Fatalf("Repair ok=%v: an error subtree was reused or an error status changed", ok)
			}
		})
	}
	// The green half: the same edit with no error anywhere is repaired.
	const clean = "sub f { my $x = 1; }\nprint 1;\n"
	if !checkRepair(t, clean, at(clean, "= ", "1", "2")) {
		t.Fatal("an error-free sub body edit returned ok=false")
	}
}

// TestRepairLastStatementEndsAtAnchorEnd: the anchor is the last sub, with
// only trivia after it, and the edit makes its re-parse read on past the
// sub's `}`. Repair must refuse. The gate review of #512 found that check 2
// has no snapshot to compare against after the last top-level statement, so
// check 1 is what catches this. Here the EOF snapshot also catches it.
func TestRepairLastStatementEndsAtAnchorEnd(t *testing.T) {
	for _, src := range []string{
		"sub g { 2 }\nsub f {\n  return 1;\n}\n",
		"sub g { 2 }\nsub f {\n  return 1;\n}\n# tail\n",
	} {
		for _, e := range []Edit{
			at(src, "return", "", "s{]"),
			at(src, "return ", "1;", "{]"),
			at(src, "return 1;", "", ":if"),
		} {
			old, newSrc := sessionTree([]byte(src)), []byte(edited(src, e))
			if s := triage(old, []byte(src), e); s != StrategyLocal {
				t.Fatalf("%q %+v: triage %v, want StrategyLocal: the case must reach Repair's checks", src, e, s)
			}
			a := anchorOf(old, e)
			if !strings.HasPrefix(src[a.Start:], "sub f") || strings.Contains(src[a.End:], "sub") {
				t.Fatalf("%q %+v: the anchor is not the last sub", src, e)
			}
			// The premise: the fresh parse's sub f really ends past the
			// anchor's shifted end.
			if f := childHolding(Parse(newSrc), Edit{Start: a.Start, OldEnd: a.Start}); f.End == a.End+len(newSrc)-len(src) {
				t.Fatalf("%q %+v: the edited sub ends where it did; the case checks nothing", src, e)
			}
			if checkRepair(t, src, e) {
				t.Fatalf("%q %+v: Repair accepted a re-parse that reads past the anchor", src, e)
			}
		}
	}
}

func hasError(root *Node) bool {
	return slices.ContainsFunc(root.Children, func(c *Node) bool { return c.HasError })
}

// TestRepairByteShiftNoUnderflow is §6.4.2's: a deletion moves every later
// node back, and each offset stays within 0 <= start <= end <= len(src).
func TestRepairByteShiftNoUnderflow(t *testing.T) {
	const src = "sub f { my $v = 1234567890; }\nsub g { 2 }\nmy @h = <<E;\nx\nE\nprint g();\n"
	for _, e := range []Edit{
		at(src, "= ", "1234567890", "1"),
		at(src, "= ", "1234567890", "$a"),
		at(src, "my $v = ", "", strings.Repeat("9", 40)),
	} {
		if !checkRepair(t, src, e) {
			t.Errorf("edit %+v: ok=false, so no node was shifted", e)
		}
	}
	// copyShifted moves InnerStart and InnerEnd only where set, as
	// TreeDiffShifted compares them.
	n := &Node{Start: 9, End: 20, Children: []*Node{{Start: 9, End: 12, InnerStart: 10, InnerEnd: 11}, {Start: 13, End: 20}}}
	if d := TreeDiffShifted(n, copyShifted(n, -9, map[*Node]*Node{}), -9); d != "" {
		t.Fatal(d)
	}
}

// latencyEdit5k is the latency benchmarks' 1-char edit (internal/psc
// editOffset): a digit before the `0` of `$_ > 0` in the middle `sub scale`.
func latencyEdit5k(t testing.TB) (string, Edit) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "psc", "testdata", "latency", "edit5k.pl"))
	if err != nil {
		t.Fatal(err)
	}
	src := string(b)
	var subs []int
	for i := 0; ; {
		j := strings.Index(src[i:], "\nsub scale {")
		if j < 0 {
			break
		}
		subs, i = append(subs, i+j+1), i+j+1
	}
	if len(subs) == 0 {
		t.Fatal("edit5k.pl has no `sub scale`")
	}
	sub := subs[len(subs)/2]
	i := sub + strings.Index(src[sub:], "$_ > 0") + len("$_ > ")
	return src, Edit{Start: i, OldEnd: i, NewText: "4"}
}

// TestRepairEdit5kIsLocal: the 1-char edit re-parses the enclosing sub, the
// top-level statement holding it, and nothing else, both from a parse's
// tree and from the tree that repair built, which carries its snapshots.
func TestRepairEdit5kIsLocal(t *testing.T) {
	src, e := latencyEdit5k(t)
	tree := sessionTree([]byte(src))
	for _, step := range []Edit{e, {Start: e.Start, OldEnd: e.Start, NewText: "2"}} {
		sub := childHolding(tree, step)
		newSrc := edited(src, step)
		next, n, ok := sessionRepair(tree, step, repPath, []byte(newSrc), []string{repRoot}, nil)
		if !ok {
			t.Fatalf("edit %+v: ok=false", step)
		}
		if n.Start != sub.Start || n.End != sub.End+len(step.NewText) || !strings.HasPrefix(newSrc[n.Start:], "sub scale {") {
			t.Fatalf("re-parsed [%d, %d), want the enclosing sub [%d, %d) shifted", n.Start, n.End, sub.Start, sub.End)
		}
		if d := TreeDiff(next, sessionTree([]byte(newSrc))); d != "" {
			t.Fatalf("edit %+v: repaired != fresh:\n%s", step, d)
		}
		if next.inc.tops == nil {
			t.Fatal("the repaired tree carries no snapshots for the next edit")
		}
		tree, src = next, newSrc
	}
}

// TestRepairUsesTokenCache: over a typing sequence on edit5k.pl, typed and
// erased from a parse's tree, each repair re-lexes a few hundred tokens of
// the file's thousands, and the token cache each tree carries is a fresh lex
// of its source; a refused edit falls back as ReparseIncremental does, to a
// tree carrying its own cache. Then a digit typed into each `$_ > 0` of the
// file in turn, 100 edits each repaired from the last's tree: the cache they
// carry stays a fresh lex, checkpoints and all.
func TestRepairUsesTokenCache(t *testing.T) {
	src, _ := latencyEdit5k(t)
	tree, cur := Parse([]byte(src)), src
	repaired, most := 0, 0
	step := func(e Edit, mustRepair bool) {
		t.Helper()
		next := []byte(edited(cur, e))
		got, ok := Repair(tree, e, next)
		switch {
		case ok:
			repaired, most = repaired+1, max(most, got.inc.relexed)
			if all := len(got.inc.toks.Tokens); got.inc.relexed*20 > all {
				t.Fatalf("edit %+v re-lexed %d of %d tokens", e, got.inc.relexed, all)
			}
			if d := TreeDiff(got, Parse(next)); d != "" {
				t.Fatalf("edit %+v: repaired != fresh:\n%s", e, d)
			}
		case mustRepair:
			t.Fatalf("edit %+v after %d repairs: ok=false", e, repaired)
		default:
			got = parseTracked(next, nil, "")
		}
		if !slices.Equal(got.inc.toks.Tokens, lexer.TokenizeBarewords(next, coreBarewordOperators())) {
			t.Fatalf("edit %+v after %d repairs: the cache is not a fresh lex", e, repaired)
		}
		tree, cur = got, string(next)
	}
	es := typingEdits(t, src)
	for _, e := range es {
		step(e, false)
	}
	if repaired < len(es)/2 {
		t.Fatalf("%d of %d typing edits repaired: the sequence must mostly reach the cache", repaired, len(es))
	}
	typed := repaired
	for k := range 100 {
		var gts []int
		for i := 0; ; {
			j := strings.Index(cur[i:], "$_ > ")
			if j < 0 {
				break
			}
			gts, i = append(gts, i+j+len("$_ > ")), i+j+1
		}
		at := gts[k*7%len(gts)]
		step(Edit{Start: at, OldEnd: at, NewText: "1"}, true)
	}
	c := tree.inc.toks
	for _, cp := range c.Checkpoints {
		i, _ := slices.BinarySearchFunc(c.Tokens, cp.Byte, func(tok lexer.Token, at int) int { return tok.Start - at })
		if got := lexer.Resume([]byte(cur), cp); !slices.Equal(got, c.Tokens[i:]) {
			t.Fatalf("after %d repairs, the checkpoint at %d no longer resumes to the stream's suffix", repaired-typed, cp.Byte)
		}
	}
	t.Logf("%d of %d typing edits repaired; at most %d of %d tokens re-lexed", typed, len(es), most, len(c.Tokens))
}

// TestRepairDoesNotMutateOldTree: whatever Repair returns, the old tree
// still equals an independent parse of its source, and its cache holds what
// it held.
func TestRepairDoesNotMutateOldTree(t *testing.T) {
	src5k, e5k := latencyEdit5k(t)
	cases := []struct {
		src string
		e   Edit
	}{
		{src5k, e5k},
		{twoSubs, at(twoSubs, "my $", "foo", "bar")},
		{twoSubs, at(twoSubs, "1;", "", "}{")},
		{"sub f { 1 }\nmy $x = ;\n", at("sub f { 1 }\nmy $x = ;\n", "{ ", "1", "2")},
		{resolvedSub, at(resolvedSub, "bar ", "1", "42")},
	}
	for _, s := range incrementalSeedSources {
		cases = append(cases, struct {
			src string
			e   Edit
		}{s, Edit{Start: len(s) / 2, OldEnd: len(s)/2 + 1, NewText: "x"}})
	}
	for _, c := range cases {
		src, newSrc := []byte(c.src), []byte(edited(c.src, c.e))
		old := sessionTree(src)
		tree, ok := NewSession().Repair(old, c.e, repPath, newSrc, repRoot)
		if d := TreeDiff(old, sessionTree(src)); d != "" || old.inc.tops != nil || string(old.inc.src) != c.src {
			t.Fatalf("edit %+v (ok=%v): the old tree changed: %s", c.e, ok, d)
		}
		if !ok {
			continue
		}
		// And a repaired tree, repaired again, is left as it was too.
		again := Edit{Start: c.e.Start, OldEnd: c.e.Start, NewText: " "}
		NewSession().Repair(tree, again, repPath, []byte(edited(string(newSrc), again)), repRoot)
		if d := TreeDiff(tree, sessionTree(newSrc)); d != "" {
			t.Fatalf("edit %+v: the repaired tree changed when repaired again: %s", again, d)
		}
		if !slices.Equal(tree.inc.toks.Tokens, lexer.TokenizeBarewords(newSrc, coreBarewordOperators())) {
			t.Fatalf("edit %+v: the repaired tree's token cache changed when repaired again", again)
		}
	}
}

// TestRepairNotOkReturnsNoTree is the sweep: random edits of the seed
// sources, each repaired or refused, never a partial tree, and both
// outcomes seen.
func TestRepairNotOkReturnsNoTree(t *testing.T) {
	srcs := []string{twoSubs, blocks, reviewUse, reviewQuote, reviewRegex, resolvedSub}
	for _, s := range incrementalSeedSources {
		srcs = append(srcs, s)
	}
	slices.Sort(srcs)
	r := rand.New(rand.NewSource(467))
	var oks, refused int
	for _, src := range srcs {
		for range 150 {
			start := r.Intn(len(src) + 1)
			end := start + min(r.Intn(4), len(src)-start)
			text := ""
			if r.Intn(4) > 0 {
				text = sweepInserts[r.Intn(len(sweepInserts))]
			}
			if checkRepair(t, src, Edit{Start: start, OldEnd: end, NewText: text}) {
				oks++
			} else {
				refused++
			}
		}
	}
	if oks == 0 || refused == 0 {
		t.Fatalf("%d repaired, %d refused: the sweep must see both", oks, refused)
	}
}

// TestRepairResyncBudgetExceeded: an unterminated string near the start of
// edit5k.pl changes tokens past its anchor, so triage refuses it and Repair
// returns ok=false having only lexed, never re-parsing what follows: it
// lexes the new source at most once and never builds a resolver, which
// every snapshot and re-parse needs. An edit changing more tokens than
// maxResyncTokens, inside one sub, is refused on the count alone.
func TestRepairResyncBudgetExceeded(t *testing.T) {
	src, _ := latencyEdit5k(t)
	e := at(src, "sub new {\n", "", `"`)
	old, newSrc := Parse([]byte(src)), []byte(edited(src, e))
	if s := triage(old, []byte(src), e); s != StrategyFull {
		t.Fatalf("triage %v, want StrategyFull", s)
	}
	d := diffTokens(tokensOf(old, []byte(src)), []byte(src), e)
	if d.relexed == 0 || d.relexed > len(d.newToks) {
		t.Fatalf("re-lexed %d tokens of %d: the new source is lexed at most once", d.relexed, len(d.newToks))
	}
	resolvers := 0
	if tree, _, ok := repair(old, e, newSrc, "", func() *resolver { resolvers++; return nil }); ok || tree != nil {
		t.Fatal("Repair accepted an unterminated string")
	}
	if resolvers != 0 {
		t.Fatalf("Repair built %d resolvers refusing on triage: it went on to snapshot or re-parse", resolvers)
	}
	// The green half of the count: an edit Repair accepts builds one.
	closed := at(src, "sub new {\n", "", `"";`)
	if _, _, accepted := repair(old, closed, []byte(edited(src, closed)), "", func() *resolver { resolvers++; return nil }); !accepted || resolvers == 0 {
		t.Fatalf("accepted=%v with %d resolvers: the count sees no work", accepted, resolvers)
	}
	// `(1, 1, ...)` becoming `q(1, 1, ...)`: balanced, inside one sub, and
	// local but for its size.
	for _, c := range []struct {
		items int
		want  Strategy
	}{{maxResyncTokens / 4, StrategyLocal}, {maxResyncTokens, StrategyFull}} {
		var b strings.Builder
		b.WriteString("sub f {\n  my @a = (")
		for range c.items {
			b.WriteString("1, ")
		}
		b.WriteString(");\n}\nprint f();\n")
		big := b.String()
		if got := triage(Parse([]byte(big)), []byte(big), at(big, "= ", "", "q")); got != c.want {
			t.Errorf("q( over %d list items: %v, want %v", c.items, got, c.want)
		}
	}
	// The green half: a closed string at the same place is local.
	if !checkRepair(t, src, at(src, "sub new {\n", "", `"";`)) {
		t.Fatal("a closed string in sub new returned ok=false")
	}
}

func TestRepairEditAtOffsetZeroAndEOF(t *testing.T) {
	for _, c := range []struct {
		src string
		e   Edit
	}{
		{twoSubs, Edit{NewText: "#"}},
		{twoSubs, Edit{NewText: "sub h { 1 }\n"}},
		{twoSubs, Edit{OldEnd: 3}},
		{twoSubs, Edit{Start: len(twoSubs), OldEnd: len(twoSubs), NewText: "1;"}},
		{twoSubs, Edit{Start: len(twoSubs) - 3, OldEnd: len(twoSubs)}},
		{twoSubs, Edit{OldEnd: len(twoSubs)}},
		{"", Edit{NewText: "sub f { 1 }\n"}},
		{"", Edit{}},
		// The anchor is the last statement, ending at EOF.
		{"sub f { 1 }", at("sub f { 1 }", "{ ", "1", "22")},
		{"print 1;\nsub f { 1 }", at("print 1;\nsub f { 1 }", "{ ", "1", "")},
	} {
		checkRepair(t, c.src, c.e)
	}
	if !checkRepair(t, "sub f { 1 }", at("sub f { 1 }", "{ ", "1", "22")) {
		t.Fatal("an edit in a sub ending at EOF returned ok=false")
	}
}

func TestRepairInsideHeredocPODAndEND(t *testing.T) {
	const heredoc = "sub f {\n  print <<EOF;\nbody\nEOF\n  return 1;\n}\nprint f();\n"
	pod, end := incrementalSeedSources["pod"], incrementalSeedSources["end"]
	for _, c := range []struct {
		src string
		e   Edit
	}{
		{heredoc, at(heredoc, "bo", "d", "x")},
		{heredoc, at(heredoc, "body", "", "\nmore")},
		{heredoc, at(heredoc, "body\n", "EOF", "EOX")},
		{pod, at(pod, "te", "x", "y")},
		{pod, at(pod, "text", "", "\n\n=cut\n\n")},
		{end, at(end, "da", "t", "x")},
		{end, at(end, "data", "", "\nsub g { 1 }\n")},
	} {
		checkRepair(t, c.src, c.e)
	}
}

func TestRepairMismatchedOldTreeNeverPanics(t *testing.T) {
	other := Parse([]byte("print 1;\n"))
	src := []byte(twoSubs)
	for _, c := range []struct {
		name string
		old  *Node
		e    Edit
		src  []byte
	}{
		{"another source's tree", other, at(twoSubs, "my $", "foo", "bar"), []byte(edited(twoSubs, at(twoSubs, "my $", "foo", "bar")))},
		{"src not the edit applied", Parse(src), at(twoSubs, "my $", "foo", "bar"), src},
		{"edit past the buffer", Parse(src), Edit{Start: len(twoSubs) + 3, OldEnd: len(twoSubs) + 9, NewText: "x"}, src},
		{"negative start", Parse(src), Edit{Start: -1, OldEnd: 2}, src},
		{"start after end", Parse(src), Edit{Start: 5, OldEnd: 2}, src},
		{"nil tree", nil, Edit{}, src},
		{"tree with no source", &Node{Kind: SourceFile, End: len(twoSubs)}, Edit{}, src},
		{"loader tree", ParseWithLoader(src, nil), at(twoSubs, "my $", "foo", "bar"), []byte(edited(twoSubs, at(twoSubs, "my $", "foo", "bar")))},
		{"Session tree, plain Repair", sessionTree([]byte(resolvedSub)), at(resolvedSub, "bar ", "1", "42"), []byte(edited(resolvedSub, at(resolvedSub, "bar ", "1", "42")))},
	} {
		t.Run(c.name, func(t *testing.T) {
			tree, ok := Repair(c.old, c.e, c.src)
			checkRepaired(t, "plain", string(c.src), c.e, tree, ok, Parse(c.src))
			tree, ok = NewSession().Repair(c.old, c.e, repPath, c.src, repRoot)
			checkRepaired(t, "Session", string(c.src), c.e, tree, ok, sessionTree(c.src))
		})
	}
	// A plain tree under a Session's Repair resolves again before reuse.
	tree, ok := NewSession().Repair(Parse([]byte(resolvedSub)), at(resolvedSub, "bar ", "1", "42"), repPath, []byte(edited(resolvedSub, at(resolvedSub, "bar ", "1", "42"))), repRoot)
	checkRepaired(t, "Session over plain", resolvedSub, Edit{}, tree, ok, sessionTree([]byte(edited(resolvedSub, at(resolvedSub, "bar ", "1", "42")))))
}

func TestRepairNonASCIIEdit(t *testing.T) {
	const mb = "sub f {\n  my $s = \"é€\"; # ü\n}\nsub g { 'ß' }\nprint f();\n"
	i := strings.Index(mb, "é")
	for _, e := range []Edit{
		{Start: i, OldEnd: i, NewText: "€"},
		{Start: i, OldEnd: i + len("é"), NewText: ""},
		{Start: i + 1, OldEnd: i + 1, NewText: "x"},
		{Start: i + 1, OldEnd: i + 3, NewText: ""},
		{Start: i, OldEnd: i + 1, NewText: "\xe2"},
		at(mb, "# ", "ü", "ö"),
		at(mb, "'", "ß", "日本"),
	} {
		checkRepair(t, mb, e)
	}
	if !checkRepair(t, mb, Edit{Start: i, OldEnd: i, NewText: "€"}) {
		t.Fatal("inserting a multi-byte character in a string returned ok=false")
	}
}

// TestRepairEveryWideningAnchorHasSnapshot is alignment O1: on edit5k.pl
// and the §7.7.3 sources, every anchor widening returns other than the
// whole file starts at a top-level parser snapshot, a sub nested in a
// package or bare block included, or Repair falls back to the whole file.
func TestRepairEveryWideningAnchorHasSnapshot(t *testing.T) {
	src5k, _ := latencyEdit5k(t)
	srcs := map[string]string{"edit5k.pl": src5k, "blocks": blocks}
	for name, s := range incrementalSeedSources {
		srcs[name] = s
	}
	nested := 0
	for name, src := range srcs {
		old := sessionTree([]byte(src))
		dirs := searchDirs(repPath, []string{repRoot})
		tops, _ := snapshotsOf(old, newResolver(dirs, false))
		if tops == nil {
			t.Fatalf("%s: no snapshots", name)
		}
		step := max(1, len(src)/400)
		for i := 0; i < len(src); i += step {
			e := Edit{Start: i, OldEnd: i, NewText: "1"}
			if triage(old, []byte(src), e) != StrategyLocal {
				continue
			}
			a := anchorOf(old, e)
			if a == old {
				continue
			}
			if slices.ContainsFunc(tops, func(s parserSnapshot) bool { return s.at == a.Start }) {
				if strings.HasPrefix(src[a.Start:], "package Q {") || strings.HasPrefix(src[a.Start:], "{ sub y") {
					nested++
				}
				continue
			}
			if _, ok := NewSession().Repair(old, e, repPath, []byte(edited(src, e)), repRoot); ok {
				t.Fatalf("%s: edit at %d anchors at %d, which has no snapshot, and Repair accepted it", name, i, a.Start)
			}
		}
	}
	if nested == 0 {
		t.Fatal("no edit anchored at a package or bare block holding a sub; the nested case is unchecked")
	}
}

// TestRepairFeatureInsideSub: a `no feature` in a sub body ends at the
// body's `}`, so the parser state after the sub matches the snapshot there
// and an edit in the body repairs, equal to a fresh parse. Were the feature
// file-level, the `isa` after the sub would be no operator and an error.
func TestRepairFeatureInsideSub(t *testing.T) {
	for _, src := range []string{
		"use v5.36;\nsub f {\n  no feature 'isa';\n  my $n = 1;\n}\nmy $r = $main::x isa Y;\n",
		"use v5.36;\nsub f {\n  no feature ':all';\n  my $n = 1;\n}\nmy $r = $main::x isa Y;\n",
		"use v5.10;\nsub f {\n  no feature 'state';\n  my $n = 1;\n}\nstate $s;\n",
	} {
		e := at(src, "my $n = ", "", "2 + ")
		if hasError(Parse([]byte(edited(src, e)))) {
			t.Fatalf("%q: the edited source holds an error; the case checks nothing", src)
		}
		if !checkRepair(t, src, e) {
			t.Errorf("%q: an edit in a sub body scoping its own feature returned ok=false", src)
		}
	}
}
