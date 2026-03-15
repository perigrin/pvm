// ABOUTME: Benchmarks for module management operations to measure performance
// ABOUTME: Provides performance validation for installer, manager, and parallel operations
package modules

import (
	"encoding/json"
	"testing"
)

// BenchmarkModuleJSON_Marshal measures performance of module JSON serialization
func BenchmarkModuleJSON_Marshal(b *testing.B) {
	module := &Module{
		Name:        "Test::Module",
		Version:     "1.00",
		Description: "Test module for benchmarking performance",
		Author:      "TEST",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(module)
		if err != nil {
			b.Fatalf("json.Marshal failed: %v", err)
		}
	}
}

// BenchmarkModuleJSON_Unmarshal measures performance of module JSON deserialization
func BenchmarkModuleJSON_Unmarshal(b *testing.B) {
	jsonData := []byte(`{
		"name": "Test::Module",
		"version": "1.00",
		"description": "Test module for benchmarking performance",
		"author": "TEST"
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var module Module
		err := json.Unmarshal(jsonData, &module)
		if err != nil {
			b.Fatalf("json.Unmarshal failed: %v", err)
		}
	}
}
