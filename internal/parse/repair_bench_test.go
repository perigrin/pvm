// ABOUTME: Times Repair over a typing sequence on edit5k.pl: a statement typed one byte at a time into a sub, then erased.
// ABOUTME: Reports p50 and p99 per edit in ms and tokens re-lexed, ReparseIncremental's full parse on a refusal included.
package parse

import (
	"slices"
	"strings"
	"testing"
	"time"
)

// typingEdits is a statement typed into the middle `sub scale` of src one
// byte at a time and then erased one byte at a time: each edit applies to
// the source the one before it left, and the last leaves src again.
func typingEdits(t testing.TB, src string) []Edit {
	t.Helper()
	const typed = "\n  my $zz = $_ * 2;"
	mid := strings.LastIndex(src[:len(src)/2], "\nsub scale {")
	if mid < 0 {
		t.Fatal("edit5k.pl has no `sub scale`")
	}
	at := mid + len("\nsub scale {")
	var es []Edit
	for k := range len(typed) {
		es = append(es, Edit{Start: at + k, OldEnd: at + k, NewText: typed[k : k+1]})
	}
	for k := len(typed); k > 0; k-- {
		es = append(es, Edit{Start: at + k - 1, OldEnd: at + k})
	}
	return es
}

// BenchmarkRepairTyping is one typing-sequence edit per iteration, carrying
// the tree from one to the next as an editor would. ms/* times every edit,
// repair-ms/* only those Repair accepted.
func BenchmarkRepairTyping(b *testing.B) {
	src, _ := latencyEdit5k(b)
	es := typingEdits(b, src)
	tree, cur := parseTracked([]byte(src), nil, ""), src
	times := make([]time.Duration, b.N)
	var repaired []time.Duration
	relexed := 0
	b.ResetTimer()
	for i := range b.N {
		e := es[i%len(es)]
		next := []byte(edited(cur, e))
		t0 := time.Now()
		t1, ok := Repair(tree, e, next)
		if ok {
			repaired = append(repaired, time.Since(t0))
			relexed += t1.inc.relexed
		} else {
			t1 = parseTracked(next, nil, "") // ReparseIncremental's fallback
		}
		times[i] = time.Since(t0)
		tree, cur = t1, string(next)
	}
	b.StopTimer()
	ms := func(ts []time.Duration, p int) float64 {
		if len(ts) == 0 {
			return 0
		}
		slices.Sort(ts)
		return float64(ts[max((p*len(ts)+99)/100-1, 0)]) / float64(time.Millisecond)
	}
	b.ReportMetric(ms(times, 50), "ms/p50")
	b.ReportMetric(ms(times, 99), "ms/p99")
	b.ReportMetric(ms(repaired, 50), "repair-ms/p50")
	b.ReportMetric(ms(repaired, 99), "repair-ms/p99")
	b.ReportMetric(1-float64(len(repaired))/float64(b.N), "refused/op")
	b.ReportMetric(float64(relexed)/float64(max(len(repaired), 1)), "relexed/repair")
}
