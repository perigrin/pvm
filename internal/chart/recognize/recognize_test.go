// ABOUTME: Tests the Earley recognizer over a toy statement grammar: derivations, interning, the cache and Expected.
// ABOUTME: Also the empty, truncated and past-the-end edge cases; the toy-lattice rejoin is in lattice_test.go.

package recognize

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// stmtGrammar is the shapes a Perl grammar gives the recognizer, in small:
// a right-recursive nullable statement sequence, nested blocks, and a
// left-recursive term.
var stmtGrammar = []string{
	"Program ::= Seq",
	"Seq ::= Stmt Seq |",
	"Stmt ::= My Var Assign Term Semi | Print Term Semi | LBrace Seq RBrace",
	"Term ::= Var | Num | Term Add Num",
}

// toyStmt is stmtGrammar's tables.
func toyStmt(tb testing.TB) *grammar.Tables { return toyTables(tb, stmtGrammar...) }

// run recognizes src, terminal names separated by spaces, with r.
func run(tb testing.TB, r *Recognizer, src string) *Chart {
	tb.Helper()
	return mustRecognize(tb, r, tokens(r.tab, strings.Fields(src)...))
}

const decl = "My Var Assign Num Semi" // a five-token statement

// TestRecognizeDerivations: spec §10's three, on the toy grammar. `Num Num`
// and `Var Var` have no derivation, nor has the unterminated `Print Num`;
// `Print Num Semi` has one.
func TestRecognizeDerivations(t *testing.T) {
	r := New(toyStmt(t), Options{})
	for src, want := range map[string]bool{"Num Num": false, "Var Var": false, "Print Num": false, "Print Num Semi": true, decl + " Print Var Semi": true} {
		if got := run(t, r, src).Accepted(); got != want {
			t.Errorf("Accepted(%q) = %v, want %v", src, got, want)
		}
	}
}

// TestRecognizeEmptyAndTruncated: no token is the empty program, which is
// accepted; one truncated mid-statement is refused, and what it expects at
// its end includes a term.
func TestRecognizeEmptyAndTruncated(t *testing.T) {
	r := New(toyStmt(t), Options{})
	if !run(t, r, "").Accepted() {
		t.Error("the empty program was refused, want accepted")
	}
	c := run(t, r, "My Var Assign")
	if c.Accepted() {
		t.Fatal("`My Var Assign` accepted, want refused")
	}
	exp, err := c.Expected(len(c.Sets) - 1)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(exp, "Var") || !slices.Contains(exp, "Num") {
		t.Errorf("Expected at the end of `My Var Assign` = %v, want a term: Var and Num among them", exp)
	}
	if slices.Contains(exp, "Semi") {
		t.Errorf("Expected at the end of `My Var Assign` holds Semi: %v", exp)
	}
}

// TestExpectedPastEnd: Expected outside the lattice's nodes is an error, and
// never a panic.
func TestExpectedPastEnd(t *testing.T) {
	r := New(toyStmt(t), Options{})
	c := run(t, r, "Print Num Semi")
	if _, err := c.Expected(len(c.Sets) - 1); err != nil {
		t.Fatalf("Expected at the last node: %v", err)
	}
	for _, pos := range []int{len(c.Sets), len(c.Sets) + 10, -1} {
		if exp, err := c.Expected(pos); err == nil && len(exp) > 0 {
			t.Errorf("Expected(%d) past the lattice = %v, nil; want an error or an empty set", pos, exp)
		}
	}
	// A refused input's dead positions have no set: empty, not a panic.
	d := run(t, r, "Num Num Num Num")
	for pos := range d.Sets {
		if _, err := d.Expected(pos); err != nil {
			t.Errorf("Expected(%d) on a refused input: %v", pos, err)
		}
	}
}

// TestInternSharesEqualSets: positions with equal content share one interned
// set (§10). A statement repeated after itself reaches the same sets.
func TestInternSharesEqualSets(t *testing.T) {
	r := New(toyStmt(t), Options{})
	// A fourth statement, so the third ends before the same lookahead.
	c := run(t, r, strings.Repeat(decl+" ", 4))
	if !c.Accepted() {
		t.Fatal("refused")
	}
	// 5 tokens a statement: node 5 ends the first, node 10 the second.
	for i := 1; i <= 5; i++ {
		if c.Sets[i] != c.Sets[i+5] || c.Sets[i+5] != c.Sets[i+10] {
			t.Errorf("node %d's set differs from nodes %d and %d", i, i+5, i+10)
		}
		if !c.Interned[i+5] || !c.Interned[i+10] {
			t.Errorf("nodes %d and %d were not interned hits", i+5, i+10)
		}
	}
	if hits, n := c.Reuse(); hits == 0 || hits >= n {
		t.Errorf("reuse %d of %d, want some but not all positions", hits, n)
	}
}

// TestInternDistinctSets: positions whose sets differ in content never
// share an interned set.
func TestInternDistinctSets(t *testing.T) {
	r := New(toyStmt(t), Options{})
	c := run(t, r, decl+" Print Var Semi LBrace My Var Assign Var Add Num Semi RBrace")
	if !c.Accepted() {
		t.Fatal("refused")
	}
	for i, a := range c.Sets {
		for j, b := range c.Sets[i+1:] {
			if a == b && !slices.Equal(a.Items(), b.Items()) {
				t.Errorf("nodes %d and %d share a set but differ in content", i, i+1+j)
			}
			if a != b && slices.Equal(a.Items(), b.Items()) {
				t.Errorf("nodes %d and %d are equal in content but were not shared", i, i+1+j)
			}
		}
	}
}

// TestCacheHitsOnSecondRecognition: recognizing the same input again hits
// the transition cache at every position after the first.
func TestCacheHitsOnSecondRecognition(t *testing.T) {
	r := New(toyStmt(t), Options{})
	src := decl + " Print Var Semi LBrace " + decl + " RBrace"
	first, second := run(t, r, src), run(t, r, src)
	if !first.Accepted() || !second.Accepted() {
		t.Fatal("refused")
	}
	for i := 1; i < len(second.Sets); i++ {
		if !second.Cached[i] {
			t.Errorf("node %d missed the transition cache on the second recognition", i)
		}
		if second.Sets[i] != first.Sets[i] {
			t.Errorf("node %d's set differs between the two recognitions", i)
		}
	}
}

// TestCacheHitKeepsAcceptance: a cache hit never changes acceptance. Every
// input is recognized by a recognizer warmed on all the others, and by one
// with no cache.
func TestCacheHitKeepsAcceptance(t *testing.T) {
	tab := toyStmt(t)
	warm := New(tab, Options{})
	srcs := []string{
		decl, "My Var Assign Num", "My Var Assign", "Print Num Semi", "Num Num Semi", "Var Var Semi",
		"LBrace " + decl + " RBrace", "LBrace My Var Assign Num Semi", "Print Var Add Num Semi", "Print Var Add Semi",
	}
	hits := 0
	for range 2 {
		for _, src := range srcs {
			c := run(t, warm, src)
			cold := run(t, New(tab, Options{NoCache: true}), src)
			if c.Accepted() != cold.Accepted() {
				t.Errorf("%q: accepted %v with the cache, %v without", src, c.Accepted(), cold.Accepted())
			}
			for _, h := range c.Cached {
				if h {
					hits++
				}
			}
		}
	}
	if hits == 0 {
		t.Error("no position hit the cache, so nothing was tested")
	}
}

// TestLookaheadKeepsAcceptance: lookahead only prunes completions that
// cannot continue, so acceptance is the same with it off.
func TestLookaheadKeepsAcceptance(t *testing.T) {
	tab := toyStmt(t)
	for _, src := range []string{decl, "My Var Assign", "Print Num", "Num Num", "LBrace Print Num Semi RBrace Num Semi", "LBrace RBrace"} {
		la := run(t, New(tab, Options{}), src).Accepted()
		no := run(t, New(tab, Options{NoLookahead: true}), src).Accepted()
		if la != no {
			t.Errorf("%q: accepted %v with lookahead, %v without", src, la, no)
		}
	}
}

// TestExpectedAfterAssign: after `My Var Assign` a term is expected, and
// `Semi` is not.
func TestExpectedAfterAssign(t *testing.T) {
	c := run(t, New(toyStmt(t), Options{}), decl)
	exp, err := c.Expected(3)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(exp, "Var") || slices.Contains(exp, "Semi") {
		t.Errorf("Expected after `My Var Assign` = %v, want Var and no Semi", exp)
	}
}

// TestExpectedIgnoresLookahead: Expected(pos) is the union over the Earley
// set as recognition without lookahead builds it, at every node, refusal
// points included: after `My Var Assign Num` before a second `Num`,
// lookahead prunes every completion, but an operator is still expected there.
func TestExpectedIgnoresLookahead(t *testing.T) {
	tab := toyStmt(t)
	for _, src := range []string{"My Var Assign Num Num Semi", decl + " Print Var Add Num Semi LBrace " + decl + " RBrace"} {
		la := run(t, New(tab, Options{}), src)
		no := run(t, New(tab, Options{NoLookahead: true}), src)
		for pos := range la.Sets {
			a, err := la.Expected(pos)
			if err != nil {
				t.Fatal(err)
			}
			b, err := no.Expected(pos)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(a, b) {
				t.Errorf("%q node %d: Expected has %d terminals with lookahead, %d without", src, pos, len(a), len(b))
			}
		}
	}
	c := run(t, New(tab, Options{}), "My Var Assign Num Num Semi")
	if exp, _ := c.Expected(4); !slices.Contains(exp, "Add") {
		t.Errorf("Expected after `My Var Assign Num` = %v, want Add among them", exp)
	}
}

// TestAcceptedNeedsOriginZero: a complete start-rule item at the last node
// accepts only from node 0. The same item from node 1 does not.
func TestAcceptedNeedsOriginZero(t *testing.T) {
	tab := toyTables(t, rightRec...)
	c := run(t, New(tab, Options{}), "A")
	start := slices.IndexFunc(c.Sets[1].Items(), func(it Item) bool { return slices.Contains(c.r.g.lhs[it.State], tab.Start) })
	if !c.Accepted() || start < 0 {
		t.Fatal("A is refused, or its last set holds no complete start item")
	}
	state := c.Sets[1].Items()[start].State
	for dist, want := range map[int]bool{2: true, 1: false} {
		moved := &Chart{r: c.r, Sets: []*Set{c.Sets[0], c.Sets[1], {items: []Item{{State: state, Dist: dist}}}}}
		if got := moved.Accepted(); got != want {
			t.Errorf("a complete start item from node %d at node 2: accepted %v, want %v", 2-dist, got, want)
		}
	}
}

// TestItemIsStateDistanceR2: a merged item is (state, distance) and nothing
// else (rule 2): Item's fields are exactly State and Dist, both int, as on pu
// before the builder's buffers were reused.
func TestItemIsStateDistanceR2(t *testing.T) {
	typ := reflect.TypeFor[Item]()
	var got []string
	for i := range typ.NumField() {
		f := typ.Field(i)
		got = append(got, f.Name+" "+f.Type.String())
	}
	if want := []string{"State int", "Dist int"}; !slices.Equal(got, want) {
		t.Errorf("Item's fields are %q, want %q", got, want)
	}
}

// TestAddDedupAcrossMapThreshold: a builder finds duplicates by scanning a
// small set and by a map once the set is big. Sets of one item fewer than
// the switch, at it and one past it, with each item added again both
// while the set is small and after it is big, hold each item once; so do
// a node's origin-set dependencies, which switch the same way.
func TestAddDedupAcrossMapThreshold(t *testing.T) {
	r := New(toyStmt(t), Options{})
	state := slices.IndexFunc(r.tab.States, func(s grammar.State) bool { return s.Epsilon < 0 })
	if state < 0 {
		t.Fatal("the toy has no state without a predicted state")
	}
	for _, n := range []int{23, 24, 25} {
		b := builder{r: r, pos: n}
		for i := range n {
			b.add(state, i)
			b.add(state, i/2) // already there, whatever the set's size
			b.read(i, &Set{id: i})
			b.read(i/2, &Set{id: i / 2})
		}
		for i := range n {
			b.add(state, i)
			b.read(i, &Set{id: i})
		}
		var origins []int
		for _, it := range b.items {
			origins = append(origins, it[1])
		}
		slices.Sort(origins)
		if len(origins) != n || len(slices.Compact(origins)) != n {
			t.Errorf("%d items: the set holds origins %v", n, origins)
		}
		if len(b.deps) != n {
			t.Errorf("%d items: %d dependencies, want %d: %v", n, len(b.deps), n, b.deps)
		}
	}
}

// TestRecognizerReuseMatchesFresh: one recognizer over the corpus sample,
// in order and in reverse, gives every file the sets, Inserted and Expected
// a fresh recognizer gives it alone. Its charts are read only after the last
// recognition, so a buffer a later file reuses cannot hide in an earlier chart.
func TestRecognizerReuseMatchesFresh(t *testing.T) {
	tab, files := corpusSample(t)
	fresh := make([]chartFacts, len(files))
	for i, f := range files {
		fresh[i] = factsOf(mustRecognize(t, New(tab, Options{}), f.l))
	}
	for _, reverse := range []bool{false, true} {
		order := make([]int, len(files))
		for i := range order {
			order[i] = i
		}
		if reverse {
			slices.Reverse(order)
		}
		r := New(tab, Options{})
		charts := make([]*Chart, len(files))
		for _, i := range order {
			charts[i] = mustRecognize(t, r, files[i].l)
		}
		for _, i := range order {
			if got := factsOf(charts[i]); !reflect.DeepEqual(got, fresh[i]) {
				t.Errorf("reverse %v, %s: one recognizer over the sample builds other sets, Inserted or Expected than a fresh one", reverse, files[i].name)
			}
		}
	}
}

// TestInternedSetsSurviveBufferReuse: an interned set owns its items. Every
// set interned over half the sample keeps its items, and interning them
// again finds that set, after the recognizer has built the other half
// through the same buffers (and, under -race, while they are read).
func TestInternedSetsSurviveBufferReuse(t *testing.T) {
	tab, files := corpusSample(t)
	r := New(tab, Options{})
	half := len(files) / 2
	for _, f := range files[:half] {
		mustRecognize(t, r, f.l)
	}
	sets := make([]*Set, 0, r.sets.nsets)
	var want [][]Item
	for _, co := range r.sets.cores {
		for ; co != nil; co = co.next {
			for _, s := range co.tab {
				if s == nil {
					continue
				}
				sets, want = append(sets, s), append(want, slices.Clone(s.items))
			}
		}
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for _, f := range files[half:] {
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
		for i, s := range sets {
			if !slices.Equal(s.items, want[i]) {
				t.Fatalf("interned set %d's items changed under a later recognition", s.id)
			}
		}
	}
	for i, s := range sets {
		if got, existed := r.sets.intern(slices.Clone(want[i])); got != s || !existed {
			t.Errorf("interning set %d's items again found set %d (existed %v)", s.id, got.id, existed)
		}
	}
}
