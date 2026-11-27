// ABOUTME: Tests the state-free evaluator behind Constrain: its traps, its per-node memo and its per-call state.
// ABOUTME: Each compares production with ConstrainReference, or pins a property the differential checks cannot see.

package forest

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/prectab"
	"tamarou.com/pvm/internal/chart/recognize"
)

// toyForestOf is ty's forest of words, failing where the chart refuses them.
func toyForestOf(t *testing.T, ty toy, words string) *Forest {
	t.Helper()
	f, _, _ := toyForest(t, ty.tab, strings.Fields(words), recognize.Options{})
	if f == nil {
		t.Fatalf("the chart refuses %q", words)
	}
	return f
}

// perlyForest is src's forest as ParseWith passes it to the constraint, with perly.y's precedence.
func perlyForest(t *testing.T, src string) (*Forest, *prectab.Precedence) {
	t.Helper()
	tab := committedTables(t)
	prec, err := prectab.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	l := lex.Linear(s)
	c, err := recognize.Recognize(recognize.NewNoPrecedence(tab, recognize.Options{}), l)
	if err != nil {
		t.Fatal(err)
	}
	f := Build(tab, c, l)
	if f == nil {
		t.Fatalf("the chart refuses %q", src)
	}
	if LibraryOperators(tab, f, s, []byte(src), Scope{}) >= 0 {
		t.Fatalf("LibraryOperators refuses %q", src)
	}
	return f, prec
}

// entries is how many eval and mark entries node n's memo lists hold.
func (c *judgement) entries(n int32) (evals, marks int) {
	for e := c.evalAt[n]; e != 0; e = c.evals[e].next {
		evals++
	}
	for e := c.markAt[n]; e != 0; e = c.marks[e].next {
		marks++
	}
	return evals, marks
}

// TestPrecedenceStateFreeTraps pins each trap the codebase lens found in the prototypes (C33) on a toy forest
// where it shows.
func TestPrecedenceStateFreeTraps(t *testing.T) {
	// emptyred's root has two alternatives of two start rules, s: VAR opt LB NUM RB with opt empty and
	// s: VAR opt, and only the first rule's spine top forbids shifting LB after the empty opt.
	t.Run("root_spine_tops", func(t *testing.T) {
		ty := loadToy(t, "emptyred")
		for _, words := range []string{"VAR LB NUM RB", "VAR LB NUM RB LB NUM RB"} {
			for _, reverse := range []bool{false, true} {
				f := toyForestOf(t, ty, words)
				if reverse {
					reverseAlts(f)
				}
				if d, _, _ := judgeBoth(f, ty.prec); d != "" {
					t.Errorf("%s, alternatives reversed %v: %s", words, reverse, d)
				}
			}
			s := ty.sentence(t, words)
			if v, g := ty.constrained(t, (*Forest).Constrain, s.words, recognize.Options{}, true); v != s.verdict || g != s.grouping {
				t.Errorf("%s reversed: got %s %s, bison %s %s", words, v, g, s.verdict, s.grouping)
			}
		}
	})

	// A rule-position node, and a symbol node whose symbol has no pair that needs a parent item, take no parent
	// item, and a spine top that forbids no shift is no context: two encodings of each hit one entry, eval's and
	// mark's.
	t.Run("canonicalisation", func(t *testing.T) {
		ty := loadToy(t, "ladder")
		f := toyForestOf(t, ty, "NUM PLUS NUM STAR NUM")
		var idle []int32 // items that forbid no shift as spine tops
		for i, row := range ty.prec.TopAt {
			if row < 0 {
				idle = append(idle, int32(i))
			}
		}
		if len(idle) < 2 {
			t.Fatalf("%d idle spine tops; want two", len(idle))
		}
		for _, c := range []struct {
			name string
			is   func(Node) bool
		}{
			{"rule-position node", func(nd Node) bool { return nd.Sym < 0 && nd.Alt != 0 }},
			{"symbol node needing no parent item", func(nd Node) bool { return nd.Sym >= 0 && nd.Alt != 0 && !ty.prec.NeedParent[nd.Sym] }},
		} {
			n := int32(slices.IndexFunc(f.Nodes, c.is))
			if n <= 0 {
				t.Fatalf("no %s", c.name)
			}
			j := newJudgement(f, ty.prec)
			eof := ty.tab.EOF
			a := j.eval(n, eof, idle[0], ty.prec.Base[1])
			b := j.eval(n, eof, idle[1], ty.prec.Base[2])
			j.mark(n, eof, idle[0], ty.prec.Base[1], -1)
			j.mark(n, eof, idle[1], ty.prec.Base[2], -1)
			if ev, mk := j.entries(n); ev != 1 || mk != 1 || !slices.Equal(a, b) {
				t.Errorf("%s %d, two encodings of one context: %d eval and %d mark entries, results %v and %v; want one each, equal", c.name, n, ev, mk, a, b)
			}
		}
	})

	// indirob before LB is decided by its parent item, so the memo keys it: one indirob node, one terminal and one
	// spine top under a parent item that forbids the reduction and one that allows it are two entries, each what a
	// fresh judgement computes.
	t.Run("parent_item_key", func(t *testing.T) {
		ty := loadToy(t, "indirob")
		p := ty.prec
		k := slices.Index(p.ReduceAt, 2)
		if k < 0 {
			t.Fatal("no pair of the indirob toy needs its parent item")
		}
		r, x := k/p.NT, k%p.NT
		lhs := ty.tab.Rules[r].LHS
		forbid, allow := int32(-1), int32(-1) // parent items, dot before lhs, the table decides each way
		for pr, rule := range ty.tab.Rules {
			for d, s := range rule.RHS {
				if s != lhs {
					continue
				}
				if pi := p.Base[pr] + int32(d); p.ParentForbid[int(p.MixAt[k])*len(p.TopAt)+int(pi)] {
					forbid = pi
				} else {
					allow = pi
				}
			}
		}
		if forbid < 0 || allow < 0 {
			t.Fatalf("parent items forbidding %d, allowing %d; want one of each", forbid, allow)
		}
		f := toyForestOf(t, ty, "LSTOP VAR LB NUM RB")
		n := int32(slices.IndexFunc(f.Nodes, func(nd Node) bool { return nd.Sym == int32(lhs) && nd.Alt != 0 }))
		if n <= 0 {
			t.Fatalf("no %s node", ty.tab.Symbols[lhs].Name)
		}
		j := newJudgement(f, p)
		a, b := j.eval(n, x, -1, forbid), j.eval(n, x, -1, allow)
		if slices.Equal(a, b) {
			t.Errorf("under a forbidding and an allowing parent item: %v and %v; want them to differ", a, b)
		}
		for _, c := range []struct {
			pi  int32
			got []int32
		}{{forbid, a}, {allow, b}} {
			if want := newJudgement(f, p).eval(n, x, -1, c.pi); !slices.Equal(c.got, want) {
				t.Errorf("parent item %d: %v, a fresh judgement %v", c.pi, c.got, want)
			}
		}
		if ev, _ := j.entries(n); ev != 2 {
			t.Errorf("%s node %d holds %d eval entries, want 2", ty.tab.Symbols[lhs].Name, n, ev)
		}
	})

	// indirob before LB is decided by its parent item; reaching the evaluator without one is a bug, never an
	// allowance.
	t.Run("pi_minus_1", func(t *testing.T) {
		ty := loadToy(t, "indirob")
		var r, x int = -1, -1
		for k, v := range ty.prec.ReduceAt {
			if v == 2 {
				r, x = k/ty.prec.NT, k%ty.prec.NT
				break
			}
		}
		if r < 0 {
			t.Fatal("no pair of the indirob toy needs its parent item")
		}
		f := toyForestOf(t, ty, "LSTOP VAR LB NUM RB")
		lhs := int32(ty.tab.Rules[r].LHS)
		n := int32(slices.IndexFunc(f.Nodes, func(nd Node) bool { return nd.Sym == lhs && nd.Alt != 0 }))
		if n <= 0 {
			t.Fatalf("no %s node", ty.tab.Symbols[lhs].Name)
		}
		got := func() (msg string) {
			defer func() { msg = fmt.Sprint(recover()) }()
			newJudgement(f, ty.prec).eval(n, x, -1, -1)
			return "no panic"
		}()
		for _, part := range []string{ty.tab.Symbols[lhs].Name, ty.tab.Symbols[x].Name, "parent item"} {
			if !strings.Contains(got, part) {
				t.Errorf("eval with no parent item: %q, want a panic naming %q", got, part)
			}
		}
	})

	// emptyred's s: VAR opt LB NUM RB with opt empty: opt reduces before LB is read, so that shift is not
	// checked, and bison accepts VAR LB NUM RB that way.
	t.Run("empty_right_child", func(t *testing.T) {
		ty := loadToy(t, "emptyred")
		s := ty.sentence(t, "VAR LB NUM RB")
		if d, ok, _ := judgeBoth(toyForestOf(t, ty, "VAR LB NUM RB"), ty.prec); d != "" || !ok {
			t.Errorf("VAR LB NUM RB: reference verdict %v, difference %q; want accepted, none", ok, d)
		}
		if v, g := ty.constrained(t, (*Forest).Constrain, s.words, recognize.Options{}, false); v != s.verdict || g != s.grouping {
			t.Errorf("got %s %s, bison %s %s", v, g, s.verdict, s.grouping)
		}
	})

	// A cached first-terminal list is shared by every parent that reads it: no entry's list is backed by another's,
	// so none can be appended into through it, and after a whole judgement each entry still holds what a fresh
	// judgement computes for its context, sorted and without repeats.
	t.Run("shared_slices", func(t *testing.T) {
		checked := 0
		for _, name := range toyNames {
			ty := loadToy(t, name)
			for _, s := range ty.sentences {
				f, _, _ := toyForest(t, ty.tab, s.words, recognize.Options{})
				if f == nil {
					continue
				}
				c := newJudgement(f.Copy(), ty.prec)
				c.constrain()
				backing := map[*int32]int{}
				for n := range f.Nodes {
					for e := c.evalAt[n]; e != 0; e = c.evals[e].next {
						x := c.evals[e]
						if cap(x.out) > 0 {
							if m, ok := backing[&x.out[:1][0]]; ok {
								t.Fatalf("%s %s: node %d's entry %v shares its array with node %d's", name, strings.Join(s.words, " "), n, x.out, m)
							}
							backing[&x.out[:1][0]] = n
						}
						want := newJudgement(f.Copy(), ty.prec).eval(int32(n), int(x.t), x.top, x.pi)
						if !slices.Equal(x.out, want) || !slices.IsSorted(x.out) || len(slices.Compact(slices.Clone(x.out))) != len(x.out) {
							t.Fatalf("%s %s: node %d's entry %v, a fresh judgement %v", name, strings.Join(s.words, " "), n, x.out, want)
						}
						checked++
					}
				}
			}
		}
		if checked == 0 {
			t.Fatal("no entry checked")
		}
	})
}

// TestPrecedenceMemoManyEntriesOneNode forces several memo entries onto one node, eval's for many terminals after
// it, then judges the whole forest on that memo, and requires the reference's verdict and kept set.
func TestPrecedenceMemoManyEntriesOneNode(t *testing.T) {
	t.Run("more_than_24", func(t *testing.T) {
		f, p := perlyForest(t, "$x = $y + 1 * $z;\n")
		n := int32(slices.IndexFunc(f.Nodes, func(nd Node) bool { return nd.Sym >= 0 && nd.Alt != 0 && !p.NeedParent[nd.Sym] }))
		if n <= 0 {
			t.Fatal("no symbol node to load")
		}
		prod, ref := f.Copy(), f.Copy()
		c := newJudgement(prod, p)
		for x := range 40 {
			c.eval(n, x, -1, -1)
		}
		if ev, _ := c.entries(n); ev <= 24 {
			t.Fatalf("node %d holds %d eval entries, want more than 24", n, ev)
		}
		if d := precedenceDiff(c.constrain(), prod, ref.ConstrainReference(p), ref); d != "" {
			t.Error(d)
		}
	})
}

// TestPrecedenceConcurrentCallsShareNothing judges many toy and perly.y forests at once, each Precedence shared by
// every goroutine, and requires each result to equal the same forest judged alone.
func TestPrecedenceConcurrentCallsShareNothing(t *testing.T) {
	type job struct {
		f    *Forest
		p    *prectab.Precedence
		ok   bool
		kept *Forest
	}
	var jobs []*job
	add := func(f *Forest, p *prectab.Precedence) {
		k := f.Copy()
		jobs = append(jobs, &job{f: f, p: p, ok: k.Constrain(p), kept: k})
	}
	for _, name := range toyNames {
		ty := loadToy(t, name)
		for _, s := range ty.sentences {
			if f, _, _ := toyForest(t, ty.tab, s.words, recognize.Options{}); f != nil {
				add(f, ty.prec)
			}
		}
	}
	tab := committedTables(t)
	prec, err := prectab.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	for _, src := range operatorCrafted {
		s, _ := lex.Lex([]byte(src), tab.Tokens)
		l := lex.Linear(s)
		c, err := recognize.Recognize(recognize.NewNoPrecedence(tab, recognize.Options{}), l)
		if err != nil || !c.Accepted() {
			continue
		}
		if f := Build(tab, c, l); LibraryOperators(tab, f, s, []byte(src), Scope{}) < 0 {
			add(f, prec)
		}
	}
	var wg sync.WaitGroup
	diffs := make([]string, 8*len(jobs))
	for g := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				j := jobs[(i+g*len(jobs)/8)%len(jobs)]
				f := j.f.Copy()
				diffs[g*len(jobs)+i] = precedenceDiff(f.Constrain(j.p), f, j.ok, j.kept)
			}
		}()
	}
	wg.Wait()
	bad := 0
	for _, d := range diffs {
		if d != "" {
			if bad++; bad <= 10 {
				t.Error(d)
			}
		}
	}
	t.Logf("forests=%d goroutines=8", len(jobs))
}

// TestPrecedenceConstrainCallsIndependent judges forest A, then B, then a fresh copy of A, and requires A's kept
// set both times: nothing from one call reaches the next. B has A's node and alternative counts and the reference
// keeps a different set of it, so a memo keyed by node index that outlived A would judge B wrongly.
func TestPrecedenceConstrainCallsIndependent(t *testing.T) {
	ty := loadToy(t, "ladder")
	a := toyForestOf(t, ty, "NUM PLUS NUM STAR NUM")
	aRef := a.Copy()
	aRef.ConstrainReference(ty.prec)
	var b *Forest
	for _, s := range ty.sentences {
		g, _, _ := toyForest(t, ty.tab, s.words, recognize.Options{})
		if g == nil || len(g.Nodes) != len(a.Nodes) || len(g.Alts) != len(a.Alts) {
			continue
		}
		if r := g.Copy(); r.ConstrainReference(ty.prec) && precedenceDiff(true, r, true, aRef) != "" {
			b = g
			break
		}
	}
	if b == nil {
		t.Fatal("no ladder sentence the size of A that the reference judges differently")
	}
	first := a.Copy()
	ok := first.Constrain(ty.prec)
	bProd, bRef := b.Copy(), b.Copy()
	if d := precedenceDiff(bProd.Constrain(ty.prec), bProd, bRef.ConstrainReference(ty.prec), bRef); d != "" {
		t.Errorf("B after A: %s", d)
	}
	again := a.Copy()
	if d := precedenceDiff(again.Constrain(ty.prec), again, ok, first); d != "" {
		t.Errorf("A after B: %s", d)
	}
	if d, _, _ := judgeBoth(a, ty.prec); d != "" {
		t.Errorf("A against the reference: %s", d)
	}
}

// TestPrecedenceStateFreeEdgeForests requires the reference's verdict and kept set, and no panic, on a forest the
// constraint refuses, a one-token sentence, a forest with no ambiguity and a root that is only nullable.
func TestPrecedenceStateFreeEdgeForests(t *testing.T) {
	ladder, nonassoc := loadToy(t, "ladder"), loadToy(t, "nonassoc")
	empty, emptyP := perlyForest(t, "")
	for _, c := range []struct {
		name string
		f    *Forest
		p    *prectab.Precedence
		ok   bool
	}{
		{"refused", toyForestOf(t, nonassoc, "NUM DOTDOT NUM DOTDOT NUM"), nonassoc.prec, false},
		{"one_token", toyForestOf(t, ladder, "NUM"), ladder.prec, true},
		{"no_ambiguity", toyForestOf(t, ladder, "NUM PLUS NUM"), ladder.prec, true},
		{"root_only_nullable", empty, emptyP, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			if c.name == "root_only_nullable" && c.f.Nodes[c.f.Root].From != c.f.Nodes[c.f.Root].To {
				t.Fatal("the root spans a token")
			}
			d, ok, _ := judgeBoth(c.f, c.p)
			if d != "" {
				t.Error(d)
			}
			if ok != c.ok {
				t.Errorf("reference verdict %v, want %v", ok, c.ok)
			}
		})
	}
}

// TestPrecedenceJudgementHoldsNoMap requires the judgement's memo to hold no map, built-in, in any field: a map
// there would put a hash lookup on the evaluation path. A hand-written hash table over a slice is the review's.
func TestPrecedenceJudgementHoldsNoMap(t *testing.T) {
	for _, ty := range []reflect.Type{reflect.TypeFor[judgement](), reflect.TypeFor[evalEntry](), reflect.TypeFor[markEntry]()} {
		for i := range ty.NumField() {
			f := ty.Field(i)
			elem := f.Type
			for elem.Kind() == reflect.Slice {
				elem = elem.Elem()
			}
			if elem.Kind() == reflect.Map {
				t.Errorf("%s.%s is %s; want no map", ty.Name(), f.Name, f.Type)
			}
		}
	}
}

// TestPrecedenceShortcutStated requires the shortcut the state-free evaluator keeps, that an alternative kept in
// one context is kept in all (decision §6), still stated on Constrain and in the README's precedence section.
func TestPrecedenceShortcutStated(t *testing.T) {
	src, err := os.ReadFile("precedence.go")
	if err != nil {
		t.Fatal(err)
	}
	_, doc, _ := strings.Cut(string(src), "// Constrain applies p to f in place")
	doc, _, _ = strings.Cut(doc, "func (f *Forest) Constrain(")
	if !strings.Contains(doc, "// shortcut: an alternative kept in one context is kept in all") {
		t.Error("Constrain's doc comment no longer states its shortcut")
	}
	readme, err := os.ReadFile(filepath.Join("..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, sec, _ := strings.Cut(string(readme), "\n## The precedence constraint")
	sec, _, _ = strings.Cut(sec, "\n## ")
	if !strings.Contains(strings.Join(strings.Fields(sec), " "), "alternative kept in one context is kept in all") {
		t.Error("the README's precedence-constraint section no longer states the shortcut")
	}
}
