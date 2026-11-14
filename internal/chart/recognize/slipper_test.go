// ABOUTME: Tests the ruby slipper: the zero-width ; toke.c inserts before } and at end of input, scanned where expected.
// ABOUTME: A toy grammar with ;-terminated statements, blocks and a hash whose } takes no ;; the reference must agree.

package recognize

import (
	"cmp"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// slipperGrammar has statements ending in ;, blocks, and a hash H whose }
// takes no ;. The statement sequence is right-recursive, so Leo applies.
var slipperGrammar = []string{
	"Prog ::= Seq",
	"Seq ::= Stmt Seq |",
	"Stmt ::= X PERLY_SEMICOLON | PERLY_BRACE_OPEN Seq PERLY_BRACE_CLOSE | H PERLY_SEMICOLON",
	"H ::= HASH X PERLY_BRACE_CLOSE",
}

// slip spells the toy's terminals short, for the test tables.
var slip = strings.NewReplacer(";", "PERLY_SEMICOLON", "{", "PERLY_BRACE_OPEN", "}", "PERLY_BRACE_CLOSE")

// TestSlipperToy: a ; is inserted before } and at end of input only where
// the set expects one, Inserted lists those nodes, Expected stays what the
// source may hold, a missing ; before another token is still refused, and a
// } that takes no ; gets no insertion. Under every Options combination.
func TestSlipperToy(t *testing.T) {
	tab := toyTables(t, slipperGrammar...)
	for _, tc := range []struct {
		src      string
		accepted bool
		inserted []int
	}{
		{"X", true, []int{1}},
		{"{ X }", true, []int{2}},
		{"X ; { X }", true, []int{4}},
		{"{ X ; }", true, nil},
		{"HASH X } ;", true, nil},        // the hash's } takes no ;
		{"HASH X }", true, []int{3}},     // only the end of input
		{"{ HASH X } }", true, []int{4}}, // before the block's }, not the hash's
		{"X X", false, nil},              // no ; before another token
		{"{ X X }", false, nil},          // nor inside a block
		{"HASH X ; }", false, nil},       // a written ; where the } takes none
		{"X }", false, []int{1}},         // inserted, but the } closes nothing
		{"", true, nil},                  // nothing expects a ;
		{"{ X } X", true, []int{2, 4}},   // end of input ends the last statement too
		{"{ X } X ;", true, []int{2}},
	} {
		src := slip.Replace(tc.src)
		for _, opts := range optionCombos {
			c := run(t, New(tab, opts), src)
			if c.Accepted() != tc.accepted {
				t.Errorf("%+v: Accepted(%q) = %v, want %v", opts, tc.src, c.Accepted(), tc.accepted)
			}
			if !slices.Equal(c.Inserted(), tc.inserted) {
				t.Errorf("%+v: Inserted(%q) = %v, want %v", opts, tc.src, c.Inserted(), tc.inserted)
			}
		}
	}
	c := run(t, New(tab, Options{}), slip.Replace("{ X }"))
	if exp := expectedOf(t, c, 2); !slices.Equal(exp, []string{"PERLY_SEMICOLON"}) {
		t.Errorf("Expected before the } of `{ X }` = %v, want only the ; the source may hold", exp)
	}
}

// slipperEmptyGrammar has a ; that is a whole statement, and blocks that
// need one, so `{ }` is accepted only by an inserted ; that completes from
// the node it was inserted at.
var slipperEmptyGrammar = []string{
	"Prog ::= Seq |",
	"Seq ::= Stmt Seq | Stmt",
	"Stmt ::= X PERLY_SEMICOLON | PERLY_SEMICOLON | PERLY_BRACE_OPEN Seq PERLY_BRACE_CLOSE",
}

// TestSlipperReferenceAgrees: the recognizer and the reference agree on
// acceptance and Expected at every node, under every Options combination
// (Leo on and off), over every input of up to five of each toy's tokens,
// each combination's recognizer warmed across them.
func TestSlipperReferenceAgrees(t *testing.T) {
	agree(t, toyTables(t, slipperGrammar...), "X", ";", "{", "}", "HASH")
	agree(t, toyTables(t, slipperEmptyGrammar...), "X", ";", "{", "}")
	if !refAccepted(toyTables(t, slipperEmptyGrammar...), slip.Replace("{ }")) {
		t.Error("the reference refuses `{ }`, which only an inserted ; accepts")
	}
}

// agree is TestSlipperReferenceAgrees over one toy grammar and its words.
func agree(t *testing.T, tab *grammar.Tables, words ...string) {
	t.Helper()
	warm := map[Options]*Recognizer{}
	build := func(o Options) *Recognizer {
		if warm[o] == nil {
			warm[o] = New(tab, o)
		}
		return warm[o]
	}
	inputs := [][]string{nil}
	for n := 0; n < len(inputs); n++ {
		if len(inputs[n]) < 5 {
			for _, w := range words {
				inputs = append(inputs, append(slices.Clone(inputs[n]), slip.Replace(w)))
			}
		}
	}
	accepted, inserted := 0, 0
	for _, in := range inputs {
		l := tokens(tab, in...)
		if d := oracleDisagreement(tab, build, l); d != "" {
			t.Fatalf("%q: %s", in, d)
		}
		c := mustRecognize(t, build(Options{}), l)
		if c.Accepted() {
			accepted++
		}
		if len(c.Inserted()) > 0 {
			inserted++
		}
	}
	if accepted == 0 || inserted == 0 {
		t.Errorf("of %d inputs %d were accepted and %d had an insertion, want some of each", len(inputs), accepted, inserted)
	}
}

// TestSlippedOwnsItsItems: the items Slipped returns are the caller's. A
// recognition on the same recognizer afterwards, run while they are read
// (under -race), writes none of them: they share no buffer it reuses.
func TestSlippedOwnsItsItems(t *testing.T) {
	tab, files := corpusSample(t)
	r := New(tab, Options{})
	c := mustRecognize(t, r, files[0].l)
	var got, want [][]Item
	for _, p := range c.Inserted() {
		next := -1
		if p < len(files[0].l.Edges) {
			next = files[0].l.Edges[p].Token.TerminalSet
		}
		if s := c.Slipped(p, next); s != nil {
			got, want = append(got, s), append(want, slices.Clone(s))
		}
	}
	if len(got) == 0 {
		t.Fatalf("%s: no node's inserted ; advances an item", files[0].name)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for _, f := range files[1:] {
			if _, err := Recognize(r, f.l); err != nil {
				t.Errorf("%s: %v", f.name, err) // Fatal may not leave a goroutine the test did not start
			}
		}
	}()
	for reading := true; reading; {
		select {
		case <-done:
			reading = false
		default:
		}
		for i := range got {
			if !slices.Equal(got[i], want[i]) {
				t.Fatalf("Slipped's items at the %dth inserted ; changed under a later recognition", i)
			}
		}
	}
}

// TestSlipperClosuresAgree: at every node of the corpus sample where a ; is
// inserted, the completed (symbol, origin, rule) facts the forest reads
// there are the same whether the slipper's items are recognition's closure
// (semiAt, Leo on, as the chart keeps it) or Slipped's recomputation (Leo
// off), Leo links expanded as the forest expands them: a chain is walked
// where its top is among the node's facts.
// No corpus-sample closure takes a Leo memo, so the slipper toys, whose
// statement sequence is right-recursive, are checked too, and some of their
// closures must differ in items for the check to mean anything.
func TestSlipperClosuresAgree(t *testing.T) {
	tab, files := corpusSample(t)
	r := New(tab, Options{})
	nodes, leo := 0, 0
	for _, f := range files {
		n, d := closuresAgree(t, f.name, mustRecognize(t, r, f.l), f.l)
		nodes, leo = nodes+n, leo+d
	}
	if nodes == 0 {
		t.Fatal("no inserted ; in the corpus sample")
	}
	t.Logf("corpus sample: %d inserted-; nodes, %d whose closures differ in items", nodes, leo)
	nodes, leo = 0, 0
	for _, g := range [][]string{slipperGrammar, slipperEmptyGrammar} {
		tab := toyTables(t, g...)
		r := New(tab, Options{})
		inputs := [][]string{nil}
		for n := 0; n < len(inputs); n++ {
			if len(inputs[n]) < 5 {
				for _, w := range []string{"X", ";", "{", "}", "HASH"} {
					inputs = append(inputs, append(slices.Clone(inputs[n]), slip.Replace(w)))
				}
			}
		}
		for _, in := range inputs {
			if slices.Contains(in, "HASH") && !slices.Equal(g, slipperGrammar) {
				continue
			}
			l := tokens(tab, in...)
			n, d := closuresAgree(t, strings.Join(in, " "), mustRecognize(t, r, l), l)
			nodes, leo = nodes+n, leo+d
		}
	}
	if leo == 0 {
		t.Errorf("of the toys' %d inserted-; nodes none took a Leo memo: the check is vacuous", nodes)
	}
}

// closuresAgree checks c's inserted-; nodes, as TestSlipperClosuresAgree
// says, and is how many there are and how many closures differ in items.
func closuresAgree(t *testing.T, name string, c *Chart, l *lex.Lattice[lex.None]) (nodes, differ int) {
	t.Helper()
	r := c.r
	for _, p := range c.Inserted() {
		next := -1
		if p < len(l.Edges) {
			next = l.Edges[p].Token.TerminalSet
		}
		la := r.g.eofBits
		if next >= 0 {
			la = r.cls.bits[next]
		}
		on, kept := c.slips[[2]int{p, next}] // recognition's closure, where it built the node after p
		if !kept {                           // a cache-served node: what recognition would have kept
			if v := r.semiAt(c, p, la); v != nil {
				on = slices.Clone(v.relative())
			}
		}
		off := c.slippedLeoOff(p, next)
		nodes++
		if !slices.Equal(on, off) {
			differ++
		}
		if got, want := slipperFacts(c, p, on), slipperFacts(c, p, off); !slices.Equal(got, want) {
			t.Errorf("%s node %d: the Leo-on closure gives facts %v, Slipped's %v", name, p, got, want)
		}
	}
	return nodes, differ
}

// slipperFacts is the sorted (symbol, origin, rule) facts at node p of the
// set there and the slipper's items, each Leo chain whose top is among them
// expanded link by link, as the forest reads them (forest.leoFacts).
func slipperFacts(c *Chart, p int, slipped []Item) [][3]int {
	tab := c.r.tab
	var facts [][3]int
	for _, items := range [2][]Item{c.Sets[p].Items(), slipped} {
		for _, it := range items {
			for _, rule := range tab.States[it.State].Reduces {
				facts = append(facts, [3]int{tab.Rules[rule].LHS, p - it.Dist, rule})
			}
		}
	}
	has := func(f [3]int) bool { return slices.Contains(facts, f) }
	for _, f := range slices.Clone(facts) {
		var chain [][3]int
		for at, sym := f[1], f[0]; ; {
			rule, origin, more, ok := c.LeoLink(at, sym)
			if !ok {
				break
			}
			chain = append(chain, [3]int{tab.Rules[rule].LHS, origin, rule})
			if !more {
				break
			}
			at, sym = origin, tab.Rules[rule].LHS
		}
		if len(chain) > 0 && has(chain[len(chain)-1]) {
			facts = append(facts, chain...)
		}
	}
	slices.SortFunc(facts, func(x, y [3]int) int {
		return cmp.Or(cmp.Compare(x[0], y[0]), cmp.Compare(x[1], y[1]), cmp.Compare(x[2], y[2]))
	})
	return slices.Compact(facts)
}

// TestBuiltSlipperClosureOwnedByChart: the slipper closure the chart keeps
// for a node recognition built is what Slipped returns there, and is the
// chart's own: a recognition on the same recognizer afterwards, run while it
// is read (under -race), writes none of it.
func TestBuiltSlipperClosureOwnedByChart(t *testing.T) {
	tab, files := corpusSample(t)
	r := New(tab, Options{})
	c := mustRecognize(t, r, files[0].l)
	var got, want [][]Item
	for k, items := range c.slips {
		if s := c.Slipped(k[0], k[1]); len(s) == 0 || &s[0] != &items[0] {
			t.Errorf("%s: Slipped(%d, %d) is not the closure the chart kept", files[0].name, k[0], k[1])
		}
		got, want = append(got, items), append(want, slices.Clone(items))
	}
	if len(got) == 0 {
		t.Fatalf("%s: the chart kept no slipper closure", files[0].name)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for _, f := range files[1:] {
			if _, err := Recognize(r, f.l); err != nil {
				t.Errorf("%s: %v", f.name, err) // Fatal may not leave a goroutine the test did not start
			}
		}
	}()
	for reading := true; reading; {
		select {
		case <-done:
			reading = false
		default:
		}
		for i := range got {
			if !slices.Equal(got[i], want[i]) {
				t.Fatalf("a kept slipper closure changed under a later recognition")
			}
		}
	}
}
