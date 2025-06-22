// ABOUTME: Benchmarks for dependency management operations to measure performance
// ABOUTME: Provides performance validation for resolution, cpanfile, and bundle operations
package dependencies

import (
	"fmt"
	"testing"
)

// BenchmarkCPANFile_AddDependency measures performance of adding dependencies
func BenchmarkCPANFile_AddDependency(b *testing.B) {
	cpanfile := &CPANFile{
		Requirements: make([]Requirement, 0),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		moduleName := fmt.Sprintf("Test::Module%d", i)
		req := Requirement{
			Phase:        "runtime",
			Relationship: "requires",
			Module:       moduleName,
			Version:      "1.00",
		}
		cpanfile.Requirements = append(cpanfile.Requirements, req)
	}
}
