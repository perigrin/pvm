// ABOUTME: Test suite for flow-sensitive type analysis implementation
// ABOUTME: Validates control flow graph construction and type state tracking

package typechecker

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/typedef"
)

// TestFlowAnalyzerCreation tests basic flow analyzer creation
func TestFlowAnalyzerCreation(t *testing.T) {
	// Create a type hierarchy and symbol table
	typeStore, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create type store: %v", err)
	}

	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	symbolTable.Package = "test"

	// Create a type checker
	tc := NewTypeChecker(hierarchy, symbolTable, "test")

	// Create flow analyzer
	analyzer := NewFlowAnalyzer(tc)

	if analyzer == nil {
		t.Fatal("Failed to create flow analyzer")
	}

	if analyzer.TypeChecker != tc {
		t.Error("Flow analyzer type checker not set correctly")
	}

	if analyzer.MaxIterations != 100 {
		t.Errorf("Expected MaxIterations to be 100, got %d", analyzer.MaxIterations)
	}

	if analyzer.ProcessedBlocks == nil {
		t.Error("ProcessedBlocks map not initialized")
	}
}

// TestControlFlowGraphConstruction tests basic CFG construction
func TestControlFlowGraphConstruction(t *testing.T) {
	// Create a simple AST for testing
	testAST := &ast.AST{
		Path: "test.pl",
		Root: &ast.BaseNode{},
	}

	// Create type checker
	typeStore, _ := typedef.NewStorage()
	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	tc := NewTypeChecker(hierarchy, symbolTable, "test")

	// Create flow analyzer
	analyzer := NewFlowAnalyzer(tc)

	// Build control flow graph
	cfg, err := analyzer.buildControlFlowGraph(testAST)
	if err != nil {
		t.Fatalf("Failed to build CFG: %v", err)
	}

	if cfg == nil {
		t.Fatal("CFG is nil")
	}

	if cfg.Entry == nil {
		t.Error("CFG entry block is nil")
	}

	if cfg.Exit == nil {
		t.Error("CFG exit block is nil")
	}

	if len(cfg.Nodes) < 2 {
		t.Errorf("Expected at least 2 nodes (entry and exit), got %d", len(cfg.Nodes))
	}
}

// TestBasicBlockTypeStateCopy tests type state copying
func TestBasicBlockTypeStateCopy(t *testing.T) {
	// Create type checker
	typeStore, _ := typedef.NewStorage()
	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	tc := NewTypeChecker(hierarchy, symbolTable, "test")

	// Create flow analyzer
	analyzer := NewFlowAnalyzer(tc)

	// Create a type state
	originalState := &TypeState{
		VariableTypes: map[string]string{
			"var1": "Int",
			"var2": "Str",
		},
		RefinedTypes: map[string]string{
			"var1": "Int|Undef",
		},
		Conditions: []Condition{
			{Variable: "var1", Operator: "==", Value: "42"},
		},
	}

	// Copy the state
	copiedState := analyzer.copyTypeState(originalState)

	if copiedState == nil {
		t.Fatal("Copied state is nil")
	}

	// Verify that it's a deep copy
	if copiedState == originalState {
		t.Error("copyTypeState returned the same reference, not a copy")
	}

	// Verify variable types are copied
	if len(copiedState.VariableTypes) != len(originalState.VariableTypes) {
		t.Error("Variable types not copied correctly")
	}

	for k, v := range originalState.VariableTypes {
		if copiedState.VariableTypes[k] != v {
			t.Errorf("Variable type %s not copied correctly", k)
		}
	}

	// Verify refined types are copied
	if len(copiedState.RefinedTypes) != len(originalState.RefinedTypes) {
		t.Error("Refined types not copied correctly")
	}

	// Verify conditions are copied
	if len(copiedState.Conditions) != len(originalState.Conditions) {
		t.Error("Conditions not copied correctly")
	}

	// Modify original to ensure it's a deep copy
	originalState.VariableTypes["var3"] = "Float"
	if _, exists := copiedState.VariableTypes["var3"]; exists {
		t.Error("copyTypeState did not create a deep copy")
	}
}

// TestTypeStateMerging tests merging of type states
func TestTypeStateMerging(t *testing.T) {
	// Create type checker
	typeStore, _ := typedef.NewStorage()
	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	tc := NewTypeChecker(hierarchy, symbolTable, "test")

	// Create flow analyzer
	analyzer := NewFlowAnalyzer(tc)

	// Create two different type states
	state1 := &TypeState{
		VariableTypes: map[string]string{
			"var1": "Int",
			"var2": "Str",
		},
		RefinedTypes: map[string]string{},
	}

	state2 := &TypeState{
		VariableTypes: map[string]string{
			"var1": "Float", // Different type for var1
			"var3": "Bool",  // New variable
		},
		RefinedTypes: map[string]string{},
	}

	// Merge the states
	merged := analyzer.mergeTypeStates(state1, state2)

	if merged == nil {
		t.Fatal("Merged state is nil")
	}

	// Check that var1 becomes a union type
	if merged.VariableTypes["var1"] != "Int|Float" {
		t.Errorf("Expected var1 to be 'Int|Float', got '%s'", merged.VariableTypes["var1"])
	}

	// Check that var2 is preserved
	if merged.VariableTypes["var2"] != "Str" {
		t.Errorf("Expected var2 to be 'Str', got '%s'", merged.VariableTypes["var2"])
	}

	// Check that var3 is included
	if merged.VariableTypes["var3"] != "Bool" {
		t.Errorf("Expected var3 to be 'Bool', got '%s'", merged.VariableTypes["var3"])
	}
}

// TestPerformFlowSensitiveAnalysisWithNilTypeState tests that analysis skips when disabled
func TestPerformFlowSensitiveAnalysisWithNilTypeState(t *testing.T) {
	// Create type checker with nil type state (disabled)
	typeStore, _ := typedef.NewStorage()
	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	tc := NewTypeChecker(hierarchy, symbolTable, "test")
	tc.TypeState = nil // Disable flow analysis

	// Create simple AST
	testAST := &ast.AST{
		Path: "test.pl",
		Root: &ast.BaseNode{},
	}

	// Perform analysis
	errors := tc.performFlowSensitiveAnalysis(testAST)

	// Should return no errors and not crash
	if len(errors) != 0 {
		t.Errorf("Expected no errors when flow analysis is disabled, got %d", len(errors))
	}
}

// TestAddFlowPatterns tests adding custom flow patterns
func TestAddFlowPatterns(t *testing.T) {
	// Create type checker
	typeStore, _ := typedef.NewStorage()
	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	tc := NewTypeChecker(hierarchy, symbolTable, "test")

	// Add flow patterns
	patterns := []string{"defined_check", "null_check", "custom_pattern"}
	tc.AddFlowPatterns(patterns)

	// Verify patterns were added
	if len(tc.ValidationPatterns) != len(patterns) {
		t.Errorf("Expected %d validation patterns, got %d", len(patterns), len(tc.ValidationPatterns))
	}

	// Verify pattern names
	for i, pattern := range patterns {
		if tc.ValidationPatterns[i].Name != pattern {
			t.Errorf("Expected pattern name '%s', got '%s'", pattern, tc.ValidationPatterns[i].Name)
		}
	}
}

// TestDefinedCheckDetection tests detection of 'defined' checks
func TestDefinedCheckDetection(t *testing.T) {
	// Create type checker and analyzer
	typeStore, _ := typedef.NewStorage()
	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	tc := NewTypeChecker(hierarchy, symbolTable, "test")
	analyzer := NewFlowAnalyzer(tc)

	// Create mock nodes with different text content
	definedNode := &ast.BaseNode{}
	definedNode.SetText("defined($var)")

	notDefinedNode := &ast.BaseNode{}
	notDefinedNode.SetText("$var == 42")

	// Test detection
	if !analyzer.isDefinedCheck(definedNode) {
		t.Error("Failed to detect defined check")
	}

	if analyzer.isDefinedCheck(notDefinedNode) {
		t.Error("False positive for defined check detection")
	}
}

// TestVariableExtractionFromCondition tests extracting variable names from conditions
func TestVariableExtractionFromCondition(t *testing.T) {
	// Create type checker and analyzer
	typeStore, _ := typedef.NewStorage()
	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	tc := NewTypeChecker(hierarchy, symbolTable, "test")
	analyzer := NewFlowAnalyzer(tc)

	// Test cases
	testCases := []struct {
		nodeText    string
		expectedVar string
		description string
	}{
		{"defined($var)", "var", "simple variable"},
		{"defined($my_var)", "my_var", "underscore variable"},
		{"defined($x)", "x", "single character variable"},
		{"if (defined($test))", "test", "variable in if statement"},
		{"$var == 42", "", "non-defined check"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			node := &ast.BaseNode{}
			node.SetText(testCase.nodeText)

			result := analyzer.extractVariableFromCondition(node)
			if result != testCase.expectedVar {
				t.Errorf("Expected variable '%s', got '%s' for text '%s'",
					testCase.expectedVar, result, testCase.nodeText)
			}
		})
	}
}
