// ABOUTME: Benchmarks for progress tracking operations to measure performance
// ABOUTME: Provides performance validation for tracking, formatting, and integration components
package progress

import (
	"fmt"
	"testing"
)

// BenchmarkTracker_BasicOperations measures performance of basic progress tracking
func BenchmarkTracker_BasicOperations(b *testing.B) {
	tracker := NewTracker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Start("test-operation", 100)
		for j := 0; j < 10; j++ {
			tracker.Update(j*10, fmt.Sprintf("Progress %d", j))
		}
		tracker.Finish(&Result{
			Success: true,
			Message: "Completed",
		})
	}
}

// BenchmarkStatus_Updates measures performance of status tracking
func BenchmarkStatus_Updates(b *testing.B) {
	status := &Status{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		status.Operation = fmt.Sprintf("operation-%d", i)
		status.Current = i % 100
		status.Total = 100
		status.Message = fmt.Sprintf("Processing item %d", i)
	}
}
