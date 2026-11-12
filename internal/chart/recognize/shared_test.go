// ABOUTME: Tests the grammar-only tables New builds once per tables value and shares between its recognizers.
// ABOUTME: Shared read-only and race-free, extended for sets added later, and never shared across tables values.

package recognize

import "testing"

// BenchmarkNew is the allocations of one New over the committed tables,
// after one warm-up New.
func BenchmarkNew(b *testing.B) {
	tab := committedTables(b)
	New(tab, Options{})
	b.ReportAllocs()
	for b.Loop() {
		New(tab, Options{})
	}
}
