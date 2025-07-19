// ABOUTME: Performance tests for Fang UI integration to ensure no regression
// ABOUTME: Tests UI rendering speed, memory usage, and scalability

package e2e

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"tamarou.com/pvm/internal/cli/ui"
	basetesting "tamarou.com/pvm/internal/testing"
	"tamarou.com/pvm/test/e2e/helpers"
)

// TestUIPerformance_BasicOperations tests performance of basic UI operations
func TestUIPerformance_BasicOperations(t *testing.T) {
	basetesting.SkipUnlessPerformance(t, "basic UI operations performance test")

	var buf bytes.Buffer
	ctx := &ui.UIContext{
		Writer:    &buf,
		ColorMode: ui.ColorNever,
		Quiet:     false,
		Verbose:   true,
	}
	output := ui.NewOutput(ctx)

	// Performance benchmarks for basic operations
	performanceTests := []struct {
		name       string
		operation  func()
		maxTime    time.Duration
		iterations int
	}{
		{
			name: "Simple message output",
			operation: func() {
				output.Info("Test message")
			},
			maxTime:    50 * time.Millisecond,
			iterations: 1000,
		},
		{
			name: "Error message output",
			operation: func() {
				output.Error("Test error")
			},
			maxTime:    50 * time.Millisecond,
			iterations: 1000,
		},
		{
			name: "Success message output",
			operation: func() {
				output.Success("Test success")
			},
			maxTime:    50 * time.Millisecond,
			iterations: 1000,
		},
		{
			name: "Warning message output",
			operation: func() {
				output.Warning("Test warning")
			},
			maxTime:    50 * time.Millisecond,
			iterations: 1000,
		},
	}

	for _, test := range performanceTests {
		t.Run(test.name, func(t *testing.T) {
			// Warm up
			for i := 0; i < 10; i++ {
				test.operation()
			}
			buf.Reset()

			// Measure performance
			start := time.Now()
			for i := 0; i < test.iterations; i++ {
				test.operation()
			}
			duration := time.Since(start)

			avgTime := duration / time.Duration(test.iterations)

			t.Logf("%s: %d operations in %v (avg: %v per operation)",
				test.name, test.iterations, duration, avgTime)

			// Check that average time per operation is reasonable
			if avgTime > test.maxTime {
				t.Errorf("Performance regression: %s took %v per operation, expected < %v",
					test.name, avgTime, test.maxTime)
			}

			// Verify output was generated
			result := buf.String()
			assert.NotEmpty(t, result, "Operations should produce output")
		})
	}
}

// TestUIPerformance_StructuredOutput tests performance of complex structured output
func TestUIPerformance_StructuredOutput(t *testing.T) {
	basetesting.SkipUnlessPerformance(t, "basic UI operations performance test")

	var buf bytes.Buffer
	ctx := &ui.UIContext{
		Writer:    &buf,
		ColorMode: ui.ColorNever,
		Quiet:     false,
	}
	output := ui.NewOutput(ctx)

	t.Run("Large table performance", func(t *testing.T) {
		// Create large table data
		headers := []string{"ID", "Name", "Status", "Version", "Path", "Description"}
		rows := make([][]string, 100)
		for i := 0; i < 100; i++ {
			rows[i] = []string{
				fmt.Sprintf("item-%d", i),
				fmt.Sprintf("Component %d", i),
				"Active",
				"1.0.0",
				fmt.Sprintf("/path/to/component/%d", i),
				fmt.Sprintf("Description for component %d with more details", i),
			}
		}

		start := time.Now()
		output.Table(headers, rows)
		duration := time.Since(start)

		t.Logf("Large table (100 rows, 6 columns) rendered in %v", duration)

		// Should complete within reasonable time
		maxTime := 500 * time.Millisecond
		if duration > maxTime {
			t.Errorf("Large table performance regression: took %v, expected < %v",
				duration, maxTime)
		}

		// Verify output
		result := buf.String()
		assert.Contains(t, result, "Component 0", "Table should contain first row")
		assert.Contains(t, result, "Component 99", "Table should contain last row")
	})

	t.Run("Large list performance", func(t *testing.T) {
		// Create large list
		items := make([]string, 500)
		for i := 0; i < 500; i++ {
			items[i] = fmt.Sprintf("List item %d with some additional content", i)
		}

		buf.Reset()
		start := time.Now()
		output.List(items)
		duration := time.Since(start)

		t.Logf("Large list (500 items) rendered in %v", duration)

		// Should complete within reasonable time
		maxTime := 200 * time.Millisecond
		if duration > maxTime {
			t.Errorf("Large list performance regression: took %v, expected < %v",
				duration, maxTime)
		}

		// Verify output
		result := buf.String()
		assert.Contains(t, result, "List item 0", "List should contain first item")
		assert.Contains(t, result, "List item 499", "List should contain last item")
	})
}

// TestUIPerformance_MemoryUsage tests memory efficiency of UI operations
func TestUIPerformance_MemoryUsage(t *testing.T) {
	basetesting.SkipUnlessPerformance(t, "basic UI operations performance test")

	t.Run("Memory usage during operations", func(t *testing.T) {
		var buf bytes.Buffer
		ctx := &ui.UIContext{
			Writer:    &buf,
			ColorMode: ui.ColorNever,
			Quiet:     false,
		}
		output := ui.NewOutput(ctx)

		// Force garbage collection and get baseline
		runtime.GC()
		runtime.GC() // Call twice to ensure cleanup
		var beforeStats runtime.MemStats
		runtime.ReadMemStats(&beforeStats)

		// Perform many UI operations
		for i := 0; i < 1000; i++ {
			output.Info("Test message %d", i)
			output.Success("Success %d", i)
			output.Warning("Warning %d", i)
			output.Error("Error %d", i)

			if i%100 == 0 {
				buf.Reset() // Clear buffer periodically
			}
		}

		// Force garbage collection and measure
		runtime.GC()
		runtime.GC()
		var afterStats runtime.MemStats
		runtime.ReadMemStats(&afterStats)

		// Calculate memory usage
		allocatedBytes := afterStats.TotalAlloc - beforeStats.TotalAlloc
		t.Logf("Memory allocated during 4000 UI operations: %d bytes", allocatedBytes)

		// Memory usage should be reasonable (less than 10MB for this test)
		maxMemory := uint64(10 * 1024 * 1024) // 10MB
		if allocatedBytes > maxMemory {
			t.Errorf("Memory usage too high: %d bytes allocated, expected < %d",
				allocatedBytes, maxMemory)
		}
	})

	t.Run("Memory cleanup", func(t *testing.T) {
		// Test that UI operations don't leak memory
		var buf bytes.Buffer
		ctx := &ui.UIContext{
			Writer:    &buf,
			ColorMode: ui.ColorNever,
			Quiet:     false,
		}

		runtime.GC()
		runtime.GC()
		var beforeStats runtime.MemStats
		runtime.ReadMemStats(&beforeStats)

		// Create and destroy many UI outputs
		for i := 0; i < 100; i++ {
			output := ui.NewOutput(ctx)
			output.Info("Test %d", i)
			output.Success("Success %d", i)
			buf.Reset()
			// Let output go out of scope
		}

		runtime.GC()
		runtime.GC()
		var afterStats runtime.MemStats
		runtime.ReadMemStats(&afterStats)

		// Check for memory leaks
		heapIncrease := afterStats.HeapInuse - beforeStats.HeapInuse
		t.Logf("Heap increase after creating/destroying 100 UI outputs: %d bytes", heapIncrease)

		// Heap should not grow significantly (some growth is expected)
		maxHeapIncrease := uint64(1024 * 1024) // 1MB
		if heapIncrease > maxHeapIncrease {
			t.Logf("Warning: Heap grew by %d bytes, may indicate memory leak", heapIncrease)
		}
	})
}

// TestUIPerformance_CLICommandSpeed tests real CLI command performance
func TestUIPerformance_CLICommandSpeed(t *testing.T) {
	basetesting.SkipUnlessPerformance(t, "basic UI operations performance test")

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test performance of common CLI commands
	commandTests := []struct {
		name    string
		command []string
		maxTime time.Duration
	}{
		{
			name:    "PVM help command",
			command: []string{"--help"},
			maxTime: 3 * time.Second,
		},
		{
			name:    "PVX help command",
			command: []string{"pvx", "--help"},
			maxTime: 3 * time.Second,
		},
		{
			name:    "PVI help command",
			command: []string{"pvi", "--help"},
			maxTime: 3 * time.Second,
		},
		{
			name:    "PSC help command",
			command: []string{"psc", "--help"},
			maxTime: 3 * time.Second,
		},
		{
			name:    "PVM version command",
			command: []string{"version"},
			maxTime: 2 * time.Second,
		},
	}

	for _, test := range commandTests {
		t.Run(test.name, func(t *testing.T) {
			// Warm up (first run might be slower due to initialization)
			_, _, _ = env.RunPVM(test.command...)

			// Measure actual performance
			start := time.Now()
			stdout := helpers.AssertPVMSucceeds(t, env, test.command, test.name)
			duration := time.Since(start)

			t.Logf("%s completed in %v", test.name, duration)

			// Check performance
			if duration > test.maxTime {
				t.Errorf("Performance regression: %s took %v, expected < %v",
					test.name, duration, test.maxTime)
			}

			// Verify output quality
			assert.NotEmpty(t, stdout, "Command should produce output")
		})
	}
}

// TestUIPerformance_ConcurrentAccess tests UI framework under concurrent access
func TestUIPerformance_ConcurrentAccess(t *testing.T) {
	basetesting.SkipUnlessPerformance(t, "basic UI operations performance test")

	t.Run("Concurrent UI operations", func(t *testing.T) {
		var buf bytes.Buffer
		var mu sync.Mutex // Protect concurrent access to shared buffer

		// Create a synchronized writer wrapper
		syncWriter := &struct {
			io.Writer
			mu *sync.Mutex
		}{&buf, &mu}

		ctx := &ui.UIContext{
			Writer:    syncWriter,
			ColorMode: ui.ColorNever,
			Quiet:     false,
		}
		output := ui.NewOutput(ctx)

		// Test concurrent access with proper synchronization
		numGoroutines := 10
		operationsPerGoroutine := 100

		start := time.Now()

		// Channel to coordinate goroutines
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				for j := 0; j < operationsPerGoroutine; j++ {
					mu.Lock()
					output.Info("Goroutine %d, operation %d", id, j)
					output.Success("Success from goroutine %d", id)
					output.Warning("Warning from goroutine %d", id)
					mu.Unlock()
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		duration := time.Since(start)
		totalOperations := numGoroutines * operationsPerGoroutine * 3 // 3 operations per iteration

		t.Logf("Concurrent operations: %d operations from %d goroutines in %v",
			totalOperations, numGoroutines, duration)

		// Should complete within reasonable time
		maxTime := 5 * time.Second
		if duration > maxTime {
			t.Errorf("Concurrent operations too slow: took %v, expected < %v",
				duration, maxTime)
		}

		// Verify some output was generated (exact count is hard to verify due to concurrency)
		result := buf.String()
		assert.NotEmpty(t, result, "Concurrent operations should produce output")
		assert.Contains(t, result, "Goroutine", "Output should contain goroutine messages")
	})
}

// TestUIPerformance_LargeOutput tests handling of very large output
func TestUIPerformance_LargeOutput(t *testing.T) {
	basetesting.SkipUnlessPerformance(t, "basic UI operations performance test")

	var buf bytes.Buffer
	ctx := &ui.UIContext{
		Writer:    &buf,
		ColorMode: ui.ColorNever,
		Quiet:     false,
	}
	output := ui.NewOutput(ctx)

	t.Run("Very long messages", func(t *testing.T) {
		// Create very long message
		longMessage := strings.Repeat("This is a very long message that tests UI performance with large text. ", 1000)

		start := time.Now()
		output.Info("%s", longMessage)
		duration := time.Since(start)

		t.Logf("Very long message (%d chars) rendered in %v", len(longMessage), duration)

		// Should handle large messages efficiently
		maxTime := 100 * time.Millisecond
		if duration > maxTime {
			t.Errorf("Large message performance regression: took %v, expected < %v",
				duration, maxTime)
		}

		result := buf.String()
		assert.Contains(t, result, "This is a very long message", "Long message should be rendered")
	})

	t.Run("Many small messages", func(t *testing.T) {
		buf.Reset()
		numMessages := 5000

		start := time.Now()
		for i := 0; i < numMessages; i++ {
			output.Info("Message %d", i)
		}
		duration := time.Since(start)

		t.Logf("%d small messages rendered in %v (avg: %v per message)",
			numMessages, duration, duration/time.Duration(numMessages))

		// Should handle many messages efficiently
		maxTime := 2 * time.Second
		if duration > maxTime {
			t.Errorf("Many messages performance regression: took %v, expected < %v",
				duration, maxTime)
		}

		result := buf.String()
		assert.Contains(t, result, "Message 0", "Should contain first message")
		assert.Contains(t, result, fmt.Sprintf("Message %d", numMessages-1), "Should contain last message")
	})
}

// TestUIPerformance_ColorModeImpact tests performance impact of different color modes
func TestUIPerformance_ColorModeImpact(t *testing.T) {
	basetesting.SkipUnlessPerformance(t, "basic UI operations performance test")

	colorModes := []struct {
		name string
		mode ui.ColorMode
	}{
		{"ColorNever", ui.ColorNever},
		{"ColorAlways", ui.ColorAlways},
		{"ColorAuto", ui.ColorAuto},
	}

	for _, colorMode := range colorModes {
		t.Run(colorMode.name, func(t *testing.T) {
			var buf bytes.Buffer
			ctx := &ui.UIContext{
				Writer:    &buf,
				ColorMode: colorMode.mode,
				Quiet:     false,
			}
			output := ui.NewOutput(ctx)

			numOperations := 1000
			start := time.Now()

			for i := 0; i < numOperations; i++ {
				output.Info("Test message %d", i)
				output.Success("Success %d", i)
				output.Error("Error %d", i)
				output.Warning("Warning %d", i)
			}

			duration := time.Since(start)
			avgTime := duration / time.Duration(numOperations*4) // 4 operations per iteration

			t.Logf("%s: %d operations in %v (avg: %v per operation)",
				colorMode.name, numOperations*4, duration, avgTime)

			// Performance should be reasonable regardless of color mode
			maxAvgTime := 100 * time.Microsecond
			if avgTime > maxAvgTime {
				t.Logf("Warning: %s performance may need optimization: %v per operation",
					colorMode.name, avgTime)
			}

			// Verify output
			result := buf.String()
			assert.NotEmpty(t, result, "Should produce output")
		})
	}
}
