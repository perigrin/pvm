// ABOUTME: Holds set identity to a committed golden: a digest of every set, Inserted and sampled Expected over the corpus sample.
// ABOUTME: One golden line per option set; only -recognize.update-set-digest rewrites it, and moving it means set identity changed.

package recognize

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var updateSetDigest = flag.Bool("recognize.update-set-digest", false,
	"rewrite testdata/set-digest.golden from this tree's recognizer")

var setDigestPath = filepath.Join("testdata", "set-digest.golden")

// digestOptions is the five option sets the golden holds, by the name its lines carry.
var digestOptions = []struct {
	name string
	o    Options
}{
	{"default", Options{}},
	{"NoCache", Options{NoCache: true}},
	{"NoLeo", Options{NoLeo: true}},
	{"NoCache+NoLeo", Options{NoCache: true, NoLeo: true}},
	{"NoLookahead", Options{NoLookahead: true}},
}

// expectedStride samples Expected at every 37th node: every node costs too much.
const expectedStride = 37

// chartFacts is what the digest reads of one chart.
type chartFacts struct {
	items    [][]Item
	inserted []int
	expected [][]string // at nodes 0, 37, 74, …
}

func factsOf(c *Chart) chartFacts {
	f := chartFacts{inserted: c.Inserted()}
	for i, s := range c.Sets {
		f.items = append(f.items, s.Items())
		if i%expectedStride == 0 {
			e, _ := c.Expected(i)
			f.expected = append(f.expected, e)
		}
	}
	return f
}

func (f chartFacts) write(h hash.Hash) {
	for i, items := range f.items {
		fmt.Fprint(h, i, items, ";")
	}
	fmt.Fprint(h, f.inserted)
	for _, e := range f.expected {
		fmt.Fprint(h, e)
	}
}

type digestLine struct {
	name             string
	accepted, cached int
	sum              []byte
}

func (d digestLine) String() string {
	return fmt.Sprintf("%s accepted %d cached %d digest %x", d.name, d.accepted, d.cached, d.sum)
}

// setDigests recognizes every sample file, whole, with one recognizer per option set.
func setDigests(t *testing.T) []digestLine {
	t.Helper()
	tab, files := corpusSample(t)
	var out []digestLine
	for _, do := range digestOptions {
		d := digestLine{name: do.name}
		h := sha256.New()
		r := New(tab, do.o)
		for _, f := range files {
			c, err := Recognize(r, f.l)
			if err != nil {
				t.Fatalf("%s %s: %v", do.name, f.name, err)
			}
			if c.Accepted() {
				d.accepted++
			}
			for _, cached := range c.Cached {
				if cached {
					d.cached++
				}
			}
			factsOf(c).write(h)
		}
		d.sum = h.Sum(nil)
		out = append(out, d)
	}
	return out
}

// TestSetDigestGolden: set identity over corpusSample under the five option
// sets equals testdata/set-digest.golden. Never regenerate it to make a change
// pass: a change that moves it has changed which sets the recognizer builds.
func TestSetDigestGolden(t *testing.T) {
	var got strings.Builder
	for _, d := range setDigests(t) {
		fmt.Fprintln(&got, d)
	}
	if *updateSetDigest {
		if err := os.WriteFile(setDigestPath, []byte(got.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(setDigestPath)
	if err != nil {
		t.Fatalf("%v (regenerate with -recognize.update-set-digest)", err)
	}
	if got.String() != string(want) {
		t.Errorf("set identity moved from %s:\ngot:\n%swant:\n%s", setDigestPath, got.String(), want)
	}
}

// TestSetDigestDetectsChange: the digest cannot pass vacuously. Every option
// set accepts files, the cache is used by default and never under NoCache, and
// one changed item, Inserted entry or sampled Expected each changes the digest.
func TestSetDigestDetectsChange(t *testing.T) {
	for _, d := range setDigests(t) {
		if d.accepted == 0 {
			t.Errorf("%s accepts no sample file", d.name)
		}
		if strings.Contains(d.name, "NoCache") != (d.cached == 0) {
			t.Errorf("%s: cached count %d", d.name, d.cached)
		}
	}

	tab, files := corpusSample(t)
	r := New(tab, Options{})
	var base chartFacts
	for _, f := range files {
		c, err := Recognize(r, f.l)
		if err != nil {
			t.Fatal(err)
		}
		if base = factsOf(c); len(base.inserted) > 0 {
			break
		}
	}
	if len(base.inserted) == 0 || len(base.items) == 0 || len(base.items[0]) == 0 || len(base.expected[0]) == 0 {
		t.Fatal("no sample file has an inserted ;, a non-empty first set and a non-empty Expected at node 0")
	}
	sum := func(f chartFacts) string {
		h := sha256.New()
		f.write(h)
		return fmt.Sprintf("%x", h.Sum(nil))
	}
	want := sum(base)
	// Each copy shares base's slices but the one it changes.
	item, ins, exp := base, base, base
	item.items = slices.Clone(base.items)
	item.items[0] = slices.Clone(base.items[0])
	item.items[0][0].Dist++
	ins.inserted = slices.Clone(base.inserted)
	ins.inserted[0]++
	exp.expected = slices.Clone(base.expected)
	exp.expected[0] = slices.Clone(base.expected[0])
	exp.expected[0][0] += "x"
	for name, f := range map[string]chartFacts{"item": item, "Inserted": ins, "Expected": exp} {
		if sum(f) == want {
			t.Errorf("changing one %s left the digest unchanged", name)
		}
	}
	if sum(base) != want {
		t.Error("changing a copy changed base's digest")
	}
}
