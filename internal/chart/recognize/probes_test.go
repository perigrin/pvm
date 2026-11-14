// ABOUTME: Holds the postdot index to its bound: goTo binary searches per lexer token over the corpus sample.
// ABOUTME: Also checks the index against a whole-set scan at every set, and its build under two recognizers at once.

package recognize

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/chart/lex"
)

// completionProbesBound is the most goTo binary searches per lexer token
// the corpus sample may take under default options: the review prototype's
// count, rounded up, recorded with the base tree's as chart-recovery's
// "R1 bound" bullet.
const completionProbesBound = 0.61

// TestCompletionProbesBounded: scan, completion and the Leo test read the
// postdot index, so goTo's binary searches per token over corpusSample stay
// within completionProbesBound. Without the index completion searches every
// item of every parent set it reads.
func TestCompletionProbesBounded(t *testing.T) {
	tab, files := corpusSample(t)
	r := New(tab, Options{})
	tokens := 0
	for _, f := range files {
		mustRecognize(t, r, f.l)
		tokens += len(f.l.Edges)
	}
	per := float64(r.probes) / float64(tokens)
	t.Logf("%d goTo binary searches over %d tokens: %.4f per token", r.probes, tokens, per)
	if per > completionProbesBound {
		t.Errorf("%.4f goTo binary searches per token, above the bound %v", per, completionProbesBound)
	}
}

// TestPostdotIndexMatchesSetScan: the index never goes stale or aliases
// across sets sharing a core. At every set of every node of the corpus
// sample, whole and with a dead-end branch, under the five option sets, an
// interned set's index lists for each symbol the items, and their target
// states, that a goTo over the whole set finds; a branch set has no core,
// so it takes the whole-set loop.
func TestPostdotIndexMatchesSetScan(t *testing.T) {
	tab, files := corpusSample(t)
	branches := 0
	for _, do := range digestOptions {
		r := New(tab, do.o)
		checked := map[*Set]bool{}
		for _, f := range files {
			ls := []*lex.Lattice[lex.None]{f.l}
			if n := len(f.l.Edges); n >= 2 {
				ls = append(ls, withDeadEnd(f.l, n/2-1))
			}
			for _, l := range ls {
				for pos, s := range mustRecognize(t, r, l).Sets {
					if s == nil || checked[s] {
						continue
					}
					checked[s] = true
					if s.id < 0 {
						branches++
						if s.core != nil {
							t.Fatalf("%s %s node %d: a branch set has a core", do.name, f.name, pos)
						}
						continue
					}
					if s.core == nil {
						t.Fatalf("%s %s node %d: interned set %d has no core", do.name, f.name, pos, s.id)
					}
					for sym := range tab.Symbols {
						var want []coreEnt
						for i, it := range s.items {
							if k, ok := r.goTo(it.State, sym); ok {
								want = append(want, coreEnt{int32(i), int32(k)})
							}
						}
						if got := r.awaiting(s, sym); !slices.Equal(got, want) {
							t.Fatalf("%s %s node %d symbol %d: the index lists %v, the set %v", do.name, f.name, pos, sym, got, want)
						}
					}
				}
			}
		}
	}
	if branches == 0 {
		t.Fatal("no branch set was built: the dead-end paths were not inside a branch")
	}
}

// TestPostdotIndexConcurrentRecognizers: two recognizers over one tables
// value, each building its cores' indexes as it interns them, run at once without
// a race (under -race) and build the sets one recognizer builds alone.
func TestPostdotIndexConcurrentRecognizers(t *testing.T) {
	tab, files := corpusSample(t)
	sets := func(r *Recognizer) string {
		var b strings.Builder
		for _, f := range files {
			c, err := Recognize(r, f.l)
			if err != nil {
				return err.Error()
			}
			for _, s := range c.Sets {
				fmt.Fprint(&b, s.Items(), ";")
			}
		}
		return b.String()
	}
	got := make([]string, 2)
	var wg sync.WaitGroup
	for i := range got {
		r := New(tab, Options{})
		wg.Add(1)
		go func() {
			defer wg.Done()
			got[i] = sets(r)
		}()
	}
	wg.Wait()
	// Alone, after: a build run beside another must come first.
	want := sets(New(tab, Options{}))
	for i, g := range got {
		if g != want {
			t.Errorf("recognizer %d, run beside another, built other sets than one run alone", i)
		}
	}
}
