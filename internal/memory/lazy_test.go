// ABOUTME: Tests for the LazyValue type in the memory package
// ABOUTME: Verifies on-demand loading, single-load guarantee under concurrency, error propagation, and cache invalidation

package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestLazyValue(t *testing.T) {
	loadCount := 0
	loadFunc := func(ctx context.Context) (string, error) {
		loadCount++
		return "loaded_value", nil
	}

	lv := NewLazyValue(loadFunc)

	// Test initial state
	if lv.IsLoaded() {
		t.Error("Expected value to not be loaded initially")
	}

	// Test first load
	value, err := lv.Load(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value != "loaded_value" {
		t.Errorf("Expected 'loaded_value', got %s", value)
	}
	if loadCount != 1 {
		t.Errorf("Expected load count 1, got %d", loadCount)
	}
	if !lv.IsLoaded() {
		t.Error("Expected value to be loaded after Load()")
	}

	// Test cached value
	cached, ok := lv.GetCached()
	if !ok {
		t.Error("Expected cached value to be available")
	}
	if cached != "loaded_value" {
		t.Errorf("Expected cached 'loaded_value', got %s", cached)
	}

	// Test second load (should use cache)
	value2, err := lv.Load(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value2 != "loaded_value" {
		t.Errorf("Expected 'loaded_value', got %s", value2)
	}
	if loadCount != 1 { // Should still be 1 (cached)
		t.Errorf("Expected load count still 1, got %d", loadCount)
	}
}

func TestLazyValueWithTTL(t *testing.T) {
	loadCount := 0
	loadFunc := func(ctx context.Context) (int, error) {
		loadCount++
		return loadCount, nil
	}

	lv := NewLazyValue(loadFunc).WithTTL(50 * time.Millisecond)

	// First load
	value1, err := lv.Load(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value1 != 1 {
		t.Errorf("Expected 1, got %d", value1)
	}

	// Second load (should use cache)
	value2, err := lv.Load(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value2 != 1 {
		t.Errorf("Expected 1, got %d", value2)
	}

	// Wait for TTL to expire
	time.Sleep(60 * time.Millisecond)

	// Third load (should reload)
	value3, err := lv.Load(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value3 != 2 {
		t.Errorf("Expected 2, got %d", value3)
	}
	if loadCount != 2 {
		t.Errorf("Expected load count 2, got %d", loadCount)
	}
}

func TestLazyValueWithValidator(t *testing.T) {
	loadCount := 0
	currentValid := true

	loadFunc := func(ctx context.Context) (string, error) {
		loadCount++
		return "value", nil
	}

	validateFunc := func(data string) bool {
		return currentValid
	}

	lv := NewLazyValue(loadFunc).WithValidator(validateFunc)

	// First load
	value1, err := lv.Load(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value1 != "value" {
		t.Errorf("Expected 'value', got %s", value1)
	}
	if loadCount != 1 {
		t.Errorf("Expected load count 1, got %d", loadCount)
	}

	// Second load with valid data (should use cache)
	value2, err := lv.Load(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value2 != "value" {
		t.Errorf("Expected 'value', got %s", value2)
	}
	if loadCount != 1 {
		t.Errorf("Expected load count still 1, got %d", loadCount)
	}

	// Invalidate data
	currentValid = false

	// Third load (should reload due to validation failure)
	value3, err := lv.Load(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value3 != "value" {
		t.Errorf("Expected 'value', got %s", value3)
	}
	if loadCount != 2 {
		t.Errorf("Expected load count 2, got %d", loadCount)
	}
}

func TestLazyValueError(t *testing.T) {
	expectedErr := errors.New("load error")
	loadFunc := func(ctx context.Context) (string, error) {
		return "", expectedErr
	}

	lv := NewLazyValue(loadFunc)

	// Test error handling
	value, err := lv.Load(context.Background())
	if err != expectedErr {
		t.Errorf("Expected load error, got %v", err)
	}
	if value != "" {
		t.Errorf("Expected empty value on error, got %s", value)
	}

	// Test that error is cached
	value2, err2 := lv.Load(context.Background())
	if err2 != expectedErr {
		t.Errorf("Expected cached error, got %v", err2)
	}
	if value2 != "" {
		t.Errorf("Expected empty value on cached error, got %s", value2)
	}
}

func TestLazyValueConcurrency(t *testing.T) {
	loadCount := 0
	var loadMu sync.Mutex

	loadFunc := func(ctx context.Context) (int, error) {
		loadMu.Lock()
		defer loadMu.Unlock()

		// Simulate slow loading
		time.Sleep(10 * time.Millisecond)
		loadCount++
		return loadCount, nil
	}

	lv := NewLazyValue(loadFunc)

	const numGoroutines = 10
	results := make([]int, numGoroutines)
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines trying to load concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			value, err := lv.Load(context.Background())
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			results[index] = value
		}(i)
	}

	wg.Wait()

	// All goroutines should get the same value
	expectedValue := results[0]
	for i, result := range results {
		if result != expectedValue {
			t.Errorf("Goroutine %d got different value: expected %d, got %d", i, expectedValue, result)
		}
	}

	// Load function should only be called once
	loadMu.Lock()
	finalLoadCount := loadCount
	loadMu.Unlock()

	if finalLoadCount != 1 {
		t.Errorf("Expected load count 1, got %d", finalLoadCount)
	}
}

func TestLazyValueInvalidate(t *testing.T) {
	loadCount := 0
	loadFunc := func(ctx context.Context) (int, error) {
		loadCount++
		return loadCount, nil
	}

	lv := NewLazyValue(loadFunc)

	// Load value
	value1, err := lv.Load(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value1 != 1 {
		t.Errorf("Expected 1, got %d", value1)
	}

	// Invalidate
	lv.Invalidate()

	if lv.IsLoaded() {
		t.Error("Expected value to not be loaded after invalidation")
	}

	// Load again (should reload)
	value2, err := lv.Load(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value2 != 2 {
		t.Errorf("Expected 2, got %d", value2)
	}
	if loadCount != 2 {
		t.Errorf("Expected load count 2, got %d", loadCount)
	}
}

func TestLazyMap(t *testing.T) {
	loadCount := make(map[string]int)
	loadFunc := func(ctx context.Context, key string) (string, error) {
		loadCount[key]++
		return "value_" + key, nil
	}

	lm := NewLazyMap(loadFunc)

	// Test loading different keys
	value1, err := lm.Get(context.Background(), "key1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value1 != "value_key1" {
		t.Errorf("Expected 'value_key1', got %s", value1)
	}

	value2, err := lm.Get(context.Background(), "key2")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value2 != "value_key2" {
		t.Errorf("Expected 'value_key2', got %s", value2)
	}

	// Test caching
	value1Again, err := lm.Get(context.Background(), "key1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value1Again != "value_key1" {
		t.Errorf("Expected 'value_key1', got %s", value1Again)
	}

	// Each key should only be loaded once
	if loadCount["key1"] != 1 {
		t.Errorf("Expected key1 load count 1, got %d", loadCount["key1"])
	}
	if loadCount["key2"] != 1 {
		t.Errorf("Expected key2 load count 1, got %d", loadCount["key2"])
	}

	// Test cached retrieval
	cached, ok := lm.GetCached("key1")
	if !ok {
		t.Error("Expected cached value for key1")
	}
	if cached != "value_key1" {
		t.Errorf("Expected 'value_key1', got %s", cached)
	}

	// Test non-existent cached key
	_, ok = lm.GetCached("nonexistent")
	if ok {
		t.Error("Expected no cached value for nonexistent key")
	}

	// Test size and keys
	if lm.Size() != 2 {
		t.Errorf("Expected size 2, got %d", lm.Size())
	}

	keys := lm.Keys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}
}

func TestLazyMapInvalidation(t *testing.T) {
	loadCount := make(map[string]int)
	loadFunc := func(ctx context.Context, key string) (int, error) {
		loadCount[key]++
		return loadCount[key], nil
	}

	lm := NewLazyMap(loadFunc)

	// Load value
	value1, err := lm.Get(context.Background(), "test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value1 != 1 {
		t.Errorf("Expected 1, got %d", value1)
	}

	// Invalidate specific key
	lm.Invalidate("test")

	// Load again (should reload)
	value2, err := lm.Get(context.Background(), "test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value2 != 2 {
		t.Errorf("Expected 2, got %d", value2)
	}

	// Load another key
	_, err = lm.Get(context.Background(), "other")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if lm.Size() != 2 {
		t.Errorf("Expected size 2, got %d", lm.Size())
	}

	// Invalidate all
	lm.InvalidateAll()

	if lm.Size() != 0 {
		t.Errorf("Expected size 0 after InvalidateAll, got %d", lm.Size())
	}
}

func TestLazySlice(t *testing.T) {
	loadCount := make([]int, 5)
	loadFunc := func(ctx context.Context, index int) (string, error) {
		loadCount[index]++
		return fmt.Sprintf("item_%d", index), nil
	}

	ls := NewLazySlice(5, loadFunc)

	if ls.Size() != 5 {
		t.Errorf("Expected size 5, got %d", ls.Size())
	}

	// Test loading specific indices
	value0, err := ls.Get(context.Background(), 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value0 != "item_0" {
		t.Errorf("Expected 'item_0', got %s", value0)
	}

	value2, err := ls.Get(context.Background(), 2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value2 != "item_2" {
		t.Errorf("Expected 'item_2', got %s", value2)
	}

	// Test caching
	value0Again, err := ls.Get(context.Background(), 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if value0Again != "item_0" {
		t.Errorf("Expected 'item_0', got %s", value0Again)
	}

	// Check load counts
	if loadCount[0] != 1 {
		t.Errorf("Expected index 0 load count 1, got %d", loadCount[0])
	}
	if loadCount[1] != 0 {
		t.Errorf("Expected index 1 load count 0, got %d", loadCount[1])
	}
	if loadCount[2] != 1 {
		t.Errorf("Expected index 2 load count 1, got %d", loadCount[2])
	}

	// Test loaded count
	if ls.LoadedCount() != 2 {
		t.Errorf("Expected loaded count 2, got %d", ls.LoadedCount())
	}

	// Test out of bounds
	_, err = ls.Get(context.Background(), -1)
	if err != ErrIndexOutOfRange {
		t.Errorf("Expected ErrIndexOutOfRange, got %v", err)
	}

	_, err = ls.Get(context.Background(), 5)
	if err != ErrIndexOutOfRange {
		t.Errorf("Expected ErrIndexOutOfRange, got %v", err)
	}
}

func TestLazyInitializer(t *testing.T) {
	initCount := 0
	initFunc := func() (string, error) {
		initCount++
		return "initialized", nil
	}

	li := NewLazyInitializer(initFunc)

	// Test multiple gets (should only initialize once)
	for i := 0; i < 3; i++ {
		value, err := li.Get()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if value != "initialized" {
			t.Errorf("Expected 'initialized', got %s", value)
		}
	}

	if initCount != 1 {
		t.Errorf("Expected init count 1, got %d", initCount)
	}
}

func TestLazyInitializerConcurrency(t *testing.T) {
	initCount := 0
	var initMu sync.Mutex

	initFunc := func() (int, error) {
		initMu.Lock()
		defer initMu.Unlock()

		// Simulate slow initialization
		time.Sleep(10 * time.Millisecond)
		initCount++
		return initCount, nil
	}

	li := NewLazyInitializer(initFunc)

	const numGoroutines = 10
	results := make([]int, numGoroutines)
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines trying to initialize concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			value, err := li.Get()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			results[index] = value
		}(i)
	}

	wg.Wait()

	// All goroutines should get the same value
	expectedValue := results[0]
	for i, result := range results {
		if result != expectedValue {
			t.Errorf("Goroutine %d got different value: expected %d, got %d", i, expectedValue, result)
		}
	}

	// Init function should only be called once
	initMu.Lock()
	finalInitCount := initCount
	initMu.Unlock()

	if finalInitCount != 1 {
		t.Errorf("Expected init count 1, got %d", finalInitCount)
	}
}

func BenchmarkLazyValueLoad(b *testing.B) {
	loadFunc := func(ctx context.Context) (string, error) {
		return "test_value", nil
	}

	lv := NewLazyValue(loadFunc)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = lv.Load(context.Background())
		}
	})
}

func BenchmarkLazyMapAccess(b *testing.B) {
	loadFunc := func(ctx context.Context, key string) (string, error) {
		return "value_" + key, nil
	}

	lm := NewLazyMap(loadFunc)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := fmt.Sprintf("key_%d", b.N%100)
			_, _ = lm.Get(context.Background(), key)
		}
	})
}
