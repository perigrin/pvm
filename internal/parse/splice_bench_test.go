// ABOUTME: BenchmarkRepairSplice times Repair's splice stage alone over latencyEdits on edit5k.pl (01a17a40-4915 stage 1).
// ABOUTME: copy-shift is the splice as landed; in-place-shift a prototype shifting the old tree's reused nodes with no copy.
package parse

import (
	"bytes"
	"slices"
	"testing"
	"time"
)

// spliceEdits is internal/psc's latencyEdits as Edits: edit k inserts
// 1 + (s mod 9) before the `0` of `$_ > 0` in the s-th `sub scale` body,
// s = k mod (number of such subs), each applying to the buffer the one
// before it left. Duplicated because internal/parse cannot import
// internal/psc's test helper; keep the two rules the same.
func spliceEdits(tb testing.TB, src []byte, n int) []Edit {
	tb.Helper()
	var zeros []int
	for i := 0; ; {
		j := bytes.Index(src[i:], []byte("\nsub scale {"))
		if j < 0 {
			break
		}
		i += j + 1
		zeros = append(zeros, i+bytes.Index(src[i:], []byte("$_ > 0"))+len("$_ > "))
	}
	if len(zeros) == 0 {
		tb.Fatal("edit5k.pl has no `sub scale`")
	}
	es := make([]Edit, n)
	for k := range n {
		s := k % len(zeros)
		es[k] = Edit{Start: zeros[s], OldEnd: zeros[s], NewText: string(rune('1' + s%9))}
		for r := s; r < len(zeros); r++ {
			zeros[r]++
		}
	}
	return es
}

// spliceInPlace is the in-place prototype: old's root with a replaced by n
// and every child past a shifted by delta where it stands. It consumes old.
func spliceInPlace(old, a, n *Node, delta int) (*Node, int) {
	root := &Node{Kind: SourceFile, Start: 0, End: old.End + delta, Children: slices.Clone(old.Children)}
	written := 0
	var shift func(*Node)
	shift = func(c *Node) {
		written++
		c.Start += delta
		c.End += delta
		if c.InnerEnd != 0 {
			c.InnerStart += delta
			c.InnerEnd += delta
		}
		for _, k := range c.Children {
			shift(k)
		}
		for _, k := range c.InnerHeredocBodies {
			shift(k)
		}
	}
	for i, c := range root.Children {
		switch {
		case c == a:
			root.Children[i] = n
		case c.Start > a.Start:
			shift(c)
		}
	}
	return root, written
}

// anchorReplaced is the old child that n, a child of tree, replaced.
func anchorReplaced(old, tree, n *Node) *Node {
	return old.Children[slices.Index(tree.Children, n)]
}

// BenchmarkRepairSplice runs one latencyEdits edit per iteration through
// Repair, carrying the tree on, and times only the splice of the reused
// nodes around the statement Repair re-parsed. nodes/op is the nodes the
// splice allocated, copied or wrote.
func BenchmarkRepairSplice(b *testing.B) {
	src, _ := latencyEdit5k(b)
	for _, mode := range []string{"copy-shift", "in-place-shift"} {
		b.Run(mode, func(b *testing.B) {
			es := spliceEdits(b, []byte(src), b.N)
			tree, cur := parseTracked([]byte(src), nil, ""), []byte(src)
			var times []time.Duration
			nodes := 0
			b.ResetTimer()
			for _, e := range es {
				next := []byte(edited(string(cur), e))
				t1, n, ok := repair(tree, e, next, "", func() *resolver { return nil })
				if !ok {
					b.Fatalf("Repair refused %+v", e)
				}
				a, delta := anchorReplaced(tree, t1, n), len(next)-len(cur)
				var k int
				t0 := time.Now()
				if mode == "copy-shift" {
					_, k = splice(tree, a, n, delta)
				} else {
					_, k = spliceInPlace(tree, a, n, delta) // tree is consumed: t1 carries on
				}
				times = append(times, time.Since(t0))
				nodes += k
				tree, cur = t1, next
			}
			b.StopTimer()
			slices.Sort(times)
			at := func(p int) float64 {
				return float64(times[max((p*len(times)+99)/100-1, 0)]) / float64(time.Millisecond)
			}
			b.ReportMetric(at(50), "ms/p50")
			b.ReportMetric(at(99), "ms/p99")
			b.ReportMetric(float64(nodes)/float64(len(es)), "nodes/op")
		})
	}
}

// TestSpliceInPlaceEqualsFresh: the prototype BenchmarkRepairSplice times
// builds the tree a fresh parse does, for edits in the first, middle and
// last `sub scale`, so its timing is of a correct splice.
func TestSpliceInPlaceEqualsFresh(t *testing.T) {
	src, _ := latencyEdit5k(t)
	es := spliceEdits(t, []byte(src), 140)
	for _, k := range []int{0, 35, 69} {
		cur := []byte(src)
		for _, e := range es[:k] {
			cur = []byte(edited(string(cur), e))
		}
		tree, e := parseTracked(cur, nil, ""), es[k]
		next := []byte(edited(string(cur), e))
		t1, n, ok := repair(tree, e, next, "", func() *resolver { return nil })
		if !ok {
			t.Fatalf("edit %d: Repair refused", k)
		}
		got, _ := spliceInPlace(tree, anchorReplaced(tree, t1, n), n, len(next)-len(cur))
		if d := TreeDiff(got, Parse(next)); d != "" {
			t.Errorf("edit %d: in-place splice differs from a fresh parse:\n%s", k, d)
		}
	}
}
