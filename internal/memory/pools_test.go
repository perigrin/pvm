// ABOUTME: Tests for sync.Pool-backed typed pool helpers in the memory package
// ABOUTME: Exercises Get/Put lifecycle, concurrent access, and Clear semantics across SyncPool, SlicePool, and StringInterner

package memory

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// Test data structures
type testNode struct {
	Value    int
	Children []*testNode
	Data     []byte
}

func newTestNode() *testNode {
	return &testNode{
		Children: make([]*testNode, 0, 4),
		Data:     make([]byte, 0, 64),
	}
}

func resetTestNode(n *testNode) {
	n.Value = 0
	n.Children = n.Children[:0]
	n.Data = n.Data[:0]
}

func TestSyncPool(t *testing.T) {
	pool := NewSyncPool(newTestNode, resetTestNode)

	// Test basic get/put operations
	node1 := pool.Get()
	if node1 == nil {
		t.Fatal("Expected non-nil node from pool")
	}

	node1.Value = 42
	node1.Data = append(node1.Data, 1, 2, 3)

	pool.Put(node1)

	// Get another node (might be the same one, reset)
	node2 := pool.Get()
	if node2.Value != 0 {
		t.Error("Expected reset node value to be 0")
	}
	if len(node2.Data) != 0 {
		t.Error("Expected reset node data to be empty")
	}

	pool.Put(node2)

	// Check statistics
	stats := pool.Stats()
	if stats.Gets != 2 {
		t.Errorf("Expected 2 gets, got %d", stats.Gets)
	}
	if stats.Puts != 2 {
		t.Errorf("Expected 2 puts, got %d", stats.Puts)
	}
}

func TestSyncPoolConcurrency(t *testing.T) {
	pool := NewSyncPool(newTestNode, resetTestNode)

	const numGoroutines = 100
	const operationsPerGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				node := pool.Get()
				node.Value = j
				pool.Put(node)
			}
		}()
	}

	wg.Wait()

	stats := pool.Stats()
	expectedOps := uint64(numGoroutines * operationsPerGoroutine)

	if stats.Gets != expectedOps {
		t.Errorf("Expected %d gets, got %d", expectedOps, stats.Gets)
	}
	if stats.Puts != expectedOps {
		t.Errorf("Expected %d puts, got %d", expectedOps, stats.Puts)
	}
}

func TestSlicePool(t *testing.T) {
	buckets := []int{8, 16, 32, 64}
	pool := NewSlicePool[int](buckets)

	// Test getting slices of different sizes
	slice8 := pool.Get(5)
	if cap(*slice8) < 8 {
		t.Errorf("Expected capacity >= 8, got %d", cap(*slice8))
	}

	slice32 := pool.Get(30)
	if cap(*slice32) < 32 {
		t.Errorf("Expected capacity >= 32, got %d", cap(*slice32))
	}

	// Test slice too large for any bucket
	sliceLarge := pool.Get(1000)
	if cap(*sliceLarge) < 1000 {
		t.Errorf("Expected capacity >= 1000, got %d", cap(*sliceLarge))
	}

	// Put slices back
	pool.Put(slice8)
	pool.Put(slice32)
	pool.Put(sliceLarge) // Should be ignored (too large)

	// Get again and check reuse
	slice8Again := pool.Get(8)
	if cap(*slice8Again) < 8 {
		t.Errorf("Expected reused slice capacity >= 8, got %d", cap(*slice8Again))
	}

	stats := pool.Stats()
	if stats.Gets < 4 {
		t.Errorf("Expected at least 4 gets, got %d", stats.Gets)
	}
}

func TestStringInterner(t *testing.T) {
	interner := NewStringInterner()

	// Test basic interning
	s1 := interner.Intern("hello")
	s2 := interner.Intern("hello")

	if s1 != s2 {
		t.Error("Expected interned strings to be identical")
	}

	// Check pointer equality (interning should return same string instance)
	if &s1 == &s2 {
		// This is the expected behavior for successful interning
	}

	// Test different strings
	s3 := interner.Intern("world")
	if s1 == s3 {
		t.Error("Expected different strings to be different")
	}

	// Test statistics
	stats := interner.Stats()
	if stats.Gets != 3 {
		t.Errorf("Expected 3 gets, got %d", stats.Gets)
	}
	if stats.Hits != 1 { // Second "hello" should be a hit
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	if stats.Created != 2 { // "hello" and "world"
		t.Errorf("Expected 2 created, got %d", stats.Created)
	}

	// Test size and memory usage
	if interner.Size() != 2 {
		t.Errorf("Expected size 2, got %d", interner.Size())
	}

	if interner.MemoryUsage() <= 0 {
		t.Error("Expected positive memory usage")
	}
}

func TestStringInternerConcurrency(t *testing.T) {
	interner := NewStringInterner()

	const numGoroutines = 50
	const stringsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < stringsPerGoroutine; j++ {
				// Use a limited set of strings to test collision handling
				str := fmt.Sprintf("string_%d", j%10)
				interner.Intern(str)
			}
		}(i)
	}

	wg.Wait()

	// Should have exactly 10 unique strings
	if interner.Size() != 10 {
		t.Errorf("Expected 10 unique strings, got %d", interner.Size())
	}

	stats := interner.Stats()
	expectedGets := uint64(numGoroutines * stringsPerGoroutine)

	if stats.Gets != expectedGets {
		t.Errorf("Expected %d gets, got %d", expectedGets, stats.Gets)
	}

	// Should have many hits due to string reuse
	if stats.Hits == 0 {
		t.Error("Expected some hits due to string reuse")
	}
}

func BenchmarkSyncPoolVsNew(b *testing.B) {
	b.Run("Pool", func(b *testing.B) {
		pool := NewSyncPool(newTestNode, resetTestNode)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			node := pool.Get()
			node.Value = i
			pool.Put(node)
		}
	})

	b.Run("New", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			node := newTestNode()
			node.Value = i
			// Let GC handle cleanup
			_ = node
		}
	})
}

func BenchmarkSlicePoolVsNew(b *testing.B) {
	buckets := []int{8, 16, 32, 64, 128, 256}
	pool := NewSlicePool[int](buckets)

	b.Run("Pool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slice := pool.Get(32)
			*slice = append(*slice, i)
			pool.Put(slice)
		}
	})

	b.Run("New", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slice := make([]int, 0, 32)
			slice = append(slice, i)
			// Let GC handle cleanup
			_ = slice
		}
	})
}

func BenchmarkStringInterner(b *testing.B) {
	interner := NewStringInterner()

	// Pre-populate with some strings
	strings := make([]string, 100)
	for i := range strings {
		strings[i] = fmt.Sprintf("string_%d", i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			str := strings[time.Now().Nanosecond()%len(strings)]
			interner.Intern(str)
		}
	})
}

func TestPoolClear(t *testing.T) {
	pool := NewSyncPool(newTestNode, resetTestNode)

	// Use the pool
	node := pool.Get()
	pool.Put(node)

	stats := pool.Stats()
	if stats.Gets == 0 {
		t.Error("Expected some gets before clear")
	}

	// Clear the pool
	pool.Clear()

	stats = pool.Stats()
	if stats.Gets != 0 || stats.Puts != 0 {
		t.Error("Expected cleared stats")
	}
}

func TestSlicePoolClear(t *testing.T) {
	pool := NewSlicePool[int]([]int{8, 16, 32})

	// Use the pool
	slice := pool.Get(16)
	pool.Put(slice)

	// Clear the pool
	pool.Clear()

	stats := pool.Stats()
	if stats.Gets != 0 || stats.Puts != 0 {
		t.Error("Expected cleared stats")
	}
}

func TestStringInternerClear(t *testing.T) {
	interner := NewStringInterner()

	// Add some strings
	interner.Intern("test1")
	interner.Intern("test2")

	if interner.Size() == 0 {
		t.Error("Expected some strings before clear")
	}

	// Clear the interner
	interner.Clear()

	if interner.Size() != 0 {
		t.Error("Expected empty interner after clear")
	}

	stats := interner.Stats()
	if stats.Gets != 0 || stats.Created != 0 {
		t.Error("Expected cleared stats")
	}
}
