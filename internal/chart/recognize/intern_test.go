// ABOUTME: Tests that the interner finds a set with one map lookup, and that equal hashes never merge different sets.
// ABOUTME: The second forces every set to one saved hash and checks that items decide identity.

package recognize

import "testing"

// TestInternOneLookup: interning a set reads the intern table once, over
// corpusSample with one recognizer for every file.
func TestInternOneLookup(t *testing.T) {
	tab, files := corpusSample(t)
	r := New(tab, Options{})
	for _, f := range files {
		mustRecognize(t, r, f.l)
	}
	if r.sets.interns == 0 {
		t.Fatal("nothing was interned")
	}
	per := float64(r.sets.lookups) / float64(r.sets.interns)
	t.Logf("%d map lookups (and %d inserts of new cores) over %d interned sets: %.4f per set", r.sets.lookups, r.sets.inserts, r.sets.interns, per)
	if per > 1 {
		t.Errorf("%.4f map lookups per interned set, above 1", per)
	}
}

// TestInternHashCollisionKeepsSetsDistinct: with every set forced to one
// hash, sets of different items, or of one core and different distances,
// intern as different sets, and equal items still intern as one.
func TestInternHashCollisionKeepsSetsDistinct(t *testing.T) {
	in := newInterner()
	in.hashMask = 0
	a := []Item{{1, 0}, {2, 0}}
	sets := [][]Item{a, {{1, 0}, {3, 0}}, {{1, 0}, {2, 1}}, {{1, 0}}, {{2, 0}, {1, 0}}}
	var got []*Set
	for i, items := range sets {
		s, existed := in.intern(items)
		if existed {
			t.Fatalf("set %d existed on first interning", i)
		}
		got = append(got, s)
	}
	for i, items := range sets {
		if s, existed := in.intern(items); s != got[i] || !existed {
			t.Errorf("set %d interned again as set %d (existed %v)", got[i].id, s.id, existed)
		}
	}
	for d := 2; d < 40; d++ { // one core past several table growths
		sets = append(sets, []Item{{1, 0}, {2, d}})
	}
	got = nil
	for _, items := range sets {
		s, _ := in.intern(items)
		got = append(got, s)
	}
	for i, items := range sets {
		if s, existed := in.intern(items); s != got[i] || !existed {
			t.Errorf("set %d interned again as set %d (existed %v)", got[i].id, s.id, existed)
		}
	}
	if got[0].core != got[2].core {
		t.Error("sets of one core do not share it")
	}
	if got[0].core == got[1].core || got[0].core == got[3].core {
		t.Error("sets of different states share a core")
	}
}
