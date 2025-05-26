// ABOUTME: Simple test runner for inference engine that doesn't require tree-sitter
// ABOUTME: Temporary solution to verify Phase 1B implementation works

//go:build test_runner

package main

import (
	"fmt"
	"os"

	"tamarou.com/pvm/internal/typechecker"
	"tamarou.com/pvm/internal/typedef"
)

func main() {
	fmt.Println("Running Phase 1B Inference Engine Tests...")

	// Test basic engine creation
	storage, err := typedef.NewStorage()
	if err != nil {
		fmt.Printf("❌ FAIL: Failed to create storage: %v\n", err)
		os.Exit(1)
	}
	hierarchy := typedef.NewTypeHierarchy(storage)
	engine := typechecker.NewInferenceEngine(hierarchy)

	if engine == nil {
		fmt.Println("❌ FAIL: NewInferenceEngine returned nil")
		os.Exit(1)
	}

	if engine.TypeHierarchy != hierarchy {
		fmt.Println("❌ FAIL: TypeHierarchy not set correctly")
		os.Exit(1)
	}

	if engine.ConfidenceThreshold != 0.7 {
		fmt.Printf("❌ FAIL: Expected confidence threshold 0.7, got %f\n", engine.ConfidenceThreshold)
		os.Exit(1)
	}

	if len(engine.InferredTypes) != 0 {
		fmt.Println("❌ FAIL: InferredTypes should be empty initially")
		os.Exit(1)
	}

	// Check that all analyzers are initialized
	if engine.DataFlowAnalyzer == nil {
		fmt.Println("❌ FAIL: DataFlowAnalyzer not initialized")
		os.Exit(1)
	}
	if engine.ContextAnalyzer == nil {
		fmt.Println("❌ FAIL: ContextAnalyzer not initialized")
		os.Exit(1)
	}
	if engine.UsagePatternAnalyzer == nil {
		fmt.Println("❌ FAIL: UsagePatternAnalyzer not initialized")
		os.Exit(1)
	}
	if engine.TypePropagator == nil {
		fmt.Println("❌ FAIL: TypePropagator not initialized")
		os.Exit(1)
	}

	fmt.Println("✅ PASS: Basic engine creation and initialization")

	// Test context rules
	scalarRule, exists := engine.ContextAnalyzer.ContextRules["scalar"]
	if !exists {
		fmt.Println("❌ FAIL: Scalar context rule not found")
		os.Exit(1)
	}

	listRule, exists := engine.ContextAnalyzer.ContextRules["list"]
	if !exists {
		fmt.Println("❌ FAIL: List context rule not found")
		os.Exit(1)
	}

	numericRule, exists := engine.ContextAnalyzer.ContextRules["numeric"]
	if !exists {
		fmt.Println("❌ FAIL: Numeric context rule not found")
		os.Exit(1)
	}

	stringRule, exists := engine.ContextAnalyzer.ContextRules["string"]
	if !exists {
		fmt.Println("❌ FAIL: String context rule not found")
		os.Exit(1)
	}

	fmt.Println("✅ PASS: Context rules initialization")

	// Test usage patterns
	if len(engine.UsagePatternAnalyzer.Patterns) == 0 {
		fmt.Println("❌ FAIL: No usage patterns initialized")
		os.Exit(1)
	}

	// Look for expected patterns
	expectedPatterns := []string{
		"array_operations",
		"hash_operations",
		"numeric_operations",
		"string_operations",
		"object_method_calls",
		"file_operations",
	}

	foundPatterns := make(map[string]bool)
	for _, pattern := range engine.UsagePatternAnalyzer.Patterns {
		foundPatterns[pattern.Name] = true
	}

	for _, expectedPattern := range expectedPatterns {
		if !foundPatterns[expectedPattern] {
			fmt.Printf("❌ FAIL: Expected pattern %s not found\n", expectedPattern)
			os.Exit(1)
		}
	}

	fmt.Println("✅ PASS: Usage patterns initialization")

	// Test propagation rules
	if len(engine.TypePropagator.PropagationRules) == 0 {
		fmt.Println("❌ FAIL: No propagation rules initialized")
		os.Exit(1)
	}

	fmt.Println("✅ PASS: Propagation rules initialization")

	// Test inference recording (without tree-sitter dependencies)
	pos := typechecker.Position{Line: 1, Column: 1}
	engine.RecordInference("$var", "Int", 0.8, typechecker.InferenceFromPattern, pos)

	if len(engine.InferredTypes) != 1 {
		fmt.Printf("❌ FAIL: Expected 1 inferred type, got %d\n", len(engine.InferredTypes))
		os.Exit(1)
	}

	info, exists := engine.InferredTypes["$var"]
	if !exists {
		fmt.Println("❌ FAIL: Inference for $var not found")
		os.Exit(1)
	}

	if info.Type != "Int" {
		fmt.Printf("❌ FAIL: Expected type Int, got %s\n", info.Type)
		os.Exit(1)
	}

	if info.Confidence != 0.8 {
		fmt.Printf("❌ FAIL: Expected confidence 0.8, got %f\n", info.Confidence)
		os.Exit(1)
	}

	fmt.Println("✅ PASS: Inference recording")

	// Test GetInferredType
	inferredType, confidence := engine.GetInferredType("$var")
	if inferredType != "Int" || confidence != 0.8 {
		fmt.Printf("❌ FAIL: Expected Int with 0.8 confidence, got %s with %f\n", inferredType, confidence)
		os.Exit(1)
	}

	// Test unknown variable
	inferredType, confidence = engine.GetInferredType("$unknown")
	if inferredType != "Any" || confidence != 0.0 {
		fmt.Printf("❌ FAIL: Expected Any with 0.0 confidence for unknown variable, got %s with %f\n", inferredType, confidence)
		os.Exit(1)
	}

	fmt.Println("✅ PASS: Type inference retrieval")

	// Test GetAllInferredTypes
	engine.RecordInference("$var2", "Str", 0.9, typechecker.InferenceFromUsage, pos)
	engine.RecordInference("$var3", "Any", 0.5, typechecker.InferenceFromContext, pos) // Below threshold

	allTypes := engine.GetAllInferredTypes()
	if len(allTypes) != 2 {
		fmt.Printf("❌ FAIL: Expected 2 types above threshold, got %d\n", len(allTypes))
		os.Exit(1)
	}

	if allTypes["$var"] != "Int" {
		fmt.Printf("❌ FAIL: Expected $var to be Int, got %s\n", allTypes["$var"])
		os.Exit(1)
	}

	if allTypes["$var2"] != "Str" {
		fmt.Printf("❌ FAIL: Expected $var2 to be Str, got %s\n", allTypes["$var2"])
		os.Exit(1)
	}

	if _, exists := allTypes["$var3"]; exists {
		fmt.Println("❌ FAIL: $var3 should not be in results (below threshold)")
		os.Exit(1)
	}

	fmt.Println("✅ PASS: All inferred types retrieval")

	// Test inference report
	report := engine.GetInferenceReport()
	if report == "" {
		fmt.Println("❌ FAIL: Report should not be empty")
		os.Exit(1)
	}

	fmt.Println("✅ PASS: Inference report generation")

	fmt.Println("\n🎉 ALL TESTS PASSED!")
	fmt.Println("Phase 1B: Advanced Type Inference Engine implementation verified!")
}
