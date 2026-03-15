// ABOUTME: Tests for the module analyzer stub
// ABOUTME: Validates that the stub correctly reports "not yet available"

package pm

import (
	"testing"
)

func TestModuleAnalyzer_Basic(t *testing.T) {
	analyzer, err := NewModuleAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	if analyzer == nil {
		t.Fatal("Analyzer should not be nil")
	}
}

func TestModuleAnalyzer_AnalyzeReturnsError(t *testing.T) {
	analyzer, err := NewModuleAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	_, err = analyzer.AnalyzeModule("/some/module.pm")
	if err == nil {
		t.Error("Expected error from stub AnalyzeModule, got nil")
	}
}
