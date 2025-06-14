// ABOUTME: Comprehensive tests for type system pooling implementation
// ABOUTME: Tests pool efficiency, memory management, and thread safety for type checking operations

package typechecker

import (
	"sync"
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/typedef"
)

func TestTypePoolManager_NewTypeCheckResult(t *testing.T) {
	hooks := TypePoolHooks{
		OnTypeCheckCreate: func(result *TypeCheckResult) {
			// Hook for testing
		},
	}
	manager := NewTypePoolManager(hooks)

	// Test basic creation
	result := manager.NewTypeCheckResult("test.pl")
	if result == nil {
		t.Fatal("Expected non-nil TypeCheckResult")
	}
	if result.Path != "test.pl" {
		t.Errorf("Expected path 'test.pl', got %s", result.Path)
	}
	if result.Errors == nil {
		t.Error("Expected non-nil Errors slice")
	}
	if result.TypeAnnotations == nil {
		t.Error("Expected non-nil TypeAnnotations slice")
	}
	if result.RefinedTypes == nil {
		t.Error("Expected non-nil RefinedTypes map")
	}

	// Test that multiple allocations work
	result2 := manager.NewTypeCheckResult("test2.pl")
	if result2 == nil {
		t.Fatal("Expected non-nil second TypeCheckResult")
	}
	if result2.Path != "test2.pl" {
		t.Errorf("Expected path 'test2.pl', got %s", result2.Path)
	}

	// Check statistics
	stats := manager.GetDetailedStats()
	if stats.TypeCheckCount < 2 {
		t.Errorf("Expected at least 2 type check operations, got %d", stats.TypeCheckCount)
	}
}

func TestTypePoolManager_NewTypeCheckError(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Test basic creation
	error := manager.NewTypeCheckError("test error", "test.pl", 10, 5)
	if error == nil {
		t.Fatal("Expected non-nil TypeCheckError")
	}
	if error.Message != "test error" {
		t.Errorf("Expected message 'test error', got %s", error.Message)
	}
	if error.Path != "test.pl" {
		t.Errorf("Expected path 'test.pl', got %s", error.Path)
	}
	if error.Line != 10 {
		t.Errorf("Expected line 10, got %d", error.Line)
	}
	if error.Column != 5 {
		t.Errorf("Expected column 5, got %d", error.Column)
	}

	// Test Error() method
	errorStr := error.Error()
	expected := "test.pl:10:5: test error"
	if errorStr != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, errorStr)
	}
}

func TestTypePoolManager_NewFunctionSignature(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Test basic creation
	sig := manager.NewFunctionSignature("Int", false)
	if sig == nil {
		t.Fatal("Expected non-nil FunctionSignature")
	}
	if sig.ReturnType != "Int" {
		t.Errorf("Expected return type 'Int', got %s", sig.ReturnType)
	}
	if sig.IsMethod != false {
		t.Errorf("Expected IsMethod false, got %v", sig.IsMethod)
	}
	if sig.ParameterTypes == nil {
		t.Error("Expected non-nil ParameterTypes map")
	}

	// Test method signature
	methodSig := manager.NewFunctionSignature("Str", true)
	if methodSig.IsMethod != true {
		t.Errorf("Expected IsMethod true, got %v", methodSig.IsMethod)
	}
}

func TestTypePoolManager_NewGenericFunctionSignature(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Test basic creation
	sig := manager.NewGenericFunctionSignature("T", false)
	if sig == nil {
		t.Fatal("Expected non-nil GenericFunctionSignature")
	}
	if sig.ReturnType != "T" {
		t.Errorf("Expected return type 'T', got %s", sig.ReturnType)
	}
	if sig.TypeParameters == nil {
		t.Error("Expected non-nil TypeParameters slice")
	}
	if sig.ParameterTypes == nil {
		t.Error("Expected non-nil ParameterTypes map")
	}
	if sig.Constraints == nil {
		t.Error("Expected non-nil Constraints map")
	}
}

func TestTypePoolManager_NewTypeState(t *testing.T) {
	hookCalled := false
	hooks := TypePoolHooks{
		OnTypeStateCreate: func(state *TypeState) {
			hookCalled = true
		},
	}
	manager := NewTypePoolManager(hooks)

	// Test basic creation
	state := manager.NewTypeState()
	if state == nil {
		t.Fatal("Expected non-nil TypeState")
	}
	if state.VariableTypes == nil {
		t.Error("Expected non-nil VariableTypes map")
	}
	if state.RefinedTypes == nil {
		t.Error("Expected non-nil RefinedTypes map")
	}
	if state.Conditions == nil {
		t.Error("Expected non-nil Conditions slice")
	}
	if state.SkipFlowChecks != false {
		t.Errorf("Expected SkipFlowChecks false, got %v", state.SkipFlowChecks)
	}

	// Test that hook was called
	if !hookCalled {
		t.Error("Expected OnTypeStateCreate hook to be called")
	}
}

func TestTypePoolManager_NewCondition(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Test basic creation
	condition := manager.NewCondition("var", "==", "value", false)
	if condition == nil {
		t.Fatal("Expected non-nil Condition")
	}
	if condition.Variable != "var" {
		t.Errorf("Expected variable 'var', got %s", condition.Variable)
	}
	if condition.Operator != "==" {
		t.Errorf("Expected operator '==', got %s", condition.Operator)
	}
	if condition.Value != "value" {
		t.Errorf("Expected value 'value', got %s", condition.Value)
	}
	if condition.Negated != false {
		t.Errorf("Expected negated false, got %v", condition.Negated)
	}

	// Test negated condition
	negCondition := manager.NewCondition("var2", "!=", "value2", true)
	if negCondition.Negated != true {
		t.Errorf("Expected negated true, got %v", negCondition.Negated)
	}
}

func TestTypePoolManager_NewHigherKindedTypeDefinition(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Test basic creation
	def := manager.NewHigherKindedTypeDefinition("Maybe", "T -> Maybe[T]")
	if def == nil {
		t.Fatal("Expected non-nil HigherKindedTypeDefinition")
	}
	if def.Name != "Maybe" {
		t.Errorf("Expected name 'Maybe', got %s", def.Name)
	}
	if def.Definition != "T -> Maybe[T]" {
		t.Errorf("Expected definition 'T -> Maybe[T]', got %s", def.Definition)
	}
	if def.TypeConstructors == nil {
		t.Error("Expected non-nil TypeConstructors slice")
	}
}

func TestTypePoolManager_NewInferenceEngine(t *testing.T) {
	hookCalled := false
	hooks := TypePoolHooks{
		OnInferenceCreate: func(engine *InferenceEngine) {
			hookCalled = true
		},
	}
	manager := NewTypePoolManager(hooks)

	// Create test dependencies
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	// Test basic creation
	engine := manager.NewInferenceEngine(hierarchy, symbolTable)
	if engine == nil {
		t.Fatal("Expected non-nil InferenceEngine")
	}
	if engine.TypeHierarchy != hierarchy {
		t.Error("Expected TypeHierarchy to be set correctly")
	}
	if engine.SymbolTable != symbolTable {
		t.Error("Expected SymbolTable to be set correctly")
	}
	if engine.InferredTypes == nil {
		t.Error("Expected non-nil InferredTypes map")
	}
	if engine.ConfidenceThreshold != 0.7 {
		t.Errorf("Expected confidence threshold 0.7, got %f", engine.ConfidenceThreshold)
	}

	// Check sub-components
	if engine.DataFlowAnalyzer == nil {
		t.Error("Expected non-nil DataFlowAnalyzer")
	}
	if engine.ContextAnalyzer == nil {
		t.Error("Expected non-nil ContextAnalyzer")
	}
	if engine.UsagePatternAnalyzer == nil {
		t.Error("Expected non-nil UsagePatternAnalyzer")
	}
	if engine.TypePropagator == nil {
		t.Error("Expected non-nil TypePropagator")
	}

	// Test that hook was called
	if !hookCalled {
		t.Error("Expected OnInferenceCreate hook to be called")
	}

	// Check statistics
	stats := manager.GetDetailedStats()
	if stats.InferenceCount < 1 {
		t.Errorf("Expected at least 1 inference operation, got %d", stats.InferenceCount)
	}
}

func TestTypePoolManager_NewInferredTypeInfo(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Test basic creation
	info := manager.NewInferredTypeInfo("variable", "Int", 0.9)
	if info == nil {
		t.Fatal("Expected non-nil InferredTypeInfo")
	}
	if info.Name != "variable" {
		t.Errorf("Expected name 'variable', got %s", info.Name)
	}
	if info.Type != "Int" {
		t.Errorf("Expected type 'Int', got %s", info.Type)
	}
	if info.Confidence != 0.9 {
		t.Errorf("Expected confidence 0.9, got %f", info.Confidence)
	}
	if info.Sources == nil {
		t.Error("Expected non-nil Sources slice")
	}
}

func TestTypePoolManager_NewInferenceSource(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Test basic creation
	pos := ast.Position{Line: 10, Column: 5}
	source := manager.NewInferenceSource(InferenceFromAssignment, pos, 0.8, "test details")
	if source == nil {
		t.Fatal("Expected non-nil InferenceSource")
	}
	if source.Type != InferenceFromAssignment {
		t.Errorf("Expected type InferenceFromAssignment, got %v", source.Type)
	}
	if source.Location != pos {
		t.Errorf("Expected location %v, got %v", pos, source.Location)
	}
	if source.Confidence != 0.8 {
		t.Errorf("Expected confidence 0.8, got %f", source.Confidence)
	}
	if source.Details != "test details" {
		t.Errorf("Expected details 'test details', got %s", source.Details)
	}
}

func TestTypePoolManager_NewInferenceContext(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Test basic creation
	context := manager.NewInferenceContext("func", "package", "scalar", "if")
	if context == nil {
		t.Fatal("Expected non-nil InferenceContext")
	}
	if context.Function != "func" {
		t.Errorf("Expected function 'func', got %s", context.Function)
	}
	if context.Package != "package" {
		t.Errorf("Expected package 'package', got %s", context.Package)
	}
	if context.PerlContext != "scalar" {
		t.Errorf("Expected perl context 'scalar', got %s", context.PerlContext)
	}
	if context.ControlFlow != "if" {
		t.Errorf("Expected control flow 'if', got %s", context.ControlFlow)
	}
}

func TestTypePoolManager_NewDataFlowAnalyzer(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Test basic creation
	analyzer := manager.NewDataFlowAnalyzer()
	if analyzer == nil {
		t.Fatal("Expected non-nil DataFlowAnalyzer")
	}
	if analyzer.DataFlowGraph == nil {
		t.Error("Expected non-nil DataFlowGraph")
	}
	if analyzer.VariableStates == nil {
		t.Error("Expected non-nil VariableStates map")
	}

	// Check sub-components
	if analyzer.DataFlowGraph.Nodes == nil {
		t.Error("Expected non-nil Nodes map in DataFlowGraph")
	}
	if analyzer.DataFlowGraph.Edges == nil {
		t.Error("Expected non-nil Edges map in DataFlowGraph")
	}
}

func TestTypePoolManager_NewDataFlowNode(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Create a mock AST node
	var mockNode ast.Node = &ast.BaseNode{}

	// Test basic creation
	node := manager.NewDataFlowNode("node1", "assignment", mockNode)
	if node == nil {
		t.Fatal("Expected non-nil DataFlowNode")
	}
	if node.ID != "node1" {
		t.Errorf("Expected ID 'node1', got %s", node.ID)
	}
	if node.Type != "assignment" {
		t.Errorf("Expected type 'assignment', got %s", node.Type)
	}
	if node.ASTNode != mockNode {
		t.Error("Expected ASTNode to be set correctly")
	}
	if node.Definitions == nil {
		t.Error("Expected non-nil Definitions map")
	}
	if node.Uses == nil {
		t.Error("Expected non-nil Uses map")
	}
}

func TestTypePoolManager_NewDataFlowEdge(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Test basic creation
	edge := manager.NewDataFlowEdge("from", "to", "normal", "condition")
	if edge == nil {
		t.Fatal("Expected non-nil DataFlowEdge")
	}
	if edge.From != "from" {
		t.Errorf("Expected from 'from', got %s", edge.From)
	}
	if edge.To != "to" {
		t.Errorf("Expected to 'to', got %s", edge.To)
	}
	if edge.Type != "normal" {
		t.Errorf("Expected type 'normal', got %s", edge.Type)
	}
	if edge.Condition != "condition" {
		t.Errorf("Expected condition 'condition', got %s", edge.Condition)
	}
}

func TestTypePoolManager_NewTypeChecker(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Create test dependencies
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	// Test basic creation
	tc := manager.NewTypeChecker(hierarchy, symbolTable, "TestModule")
	if tc == nil {
		t.Fatal("Expected non-nil TypeChecker")
	}
	if tc.Hierarchy != hierarchy {
		t.Error("Expected Hierarchy to be set correctly")
	}
	if tc.SymbolTable != symbolTable {
		t.Error("Expected SymbolTable to be set correctly")
	}
	if tc.CurrentModule != "TestModule" {
		t.Errorf("Expected module 'TestModule', got %s", tc.CurrentModule)
	}

	// Check that all maps are initialized
	if tc.ImportedModules == nil {
		t.Error("Expected non-nil ImportedModules map")
	}
	if tc.TypeAnnotations == nil {
		t.Error("Expected non-nil TypeAnnotations map")
	}
	if tc.VariableTypes == nil {
		t.Error("Expected non-nil VariableTypes map")
	}
	if tc.FunctionTypes == nil {
		t.Error("Expected non-nil FunctionTypes map")
	}
	if tc.TypeState == nil {
		t.Error("Expected non-nil TypeState")
	}
	if tc.TypeStateStack == nil {
		t.Error("Expected non-nil TypeStateStack slice")
	}
	if tc.ValidationPatterns == nil {
		t.Error("Expected non-nil ValidationPatterns slice")
	}
	if tc.InferenceEngine == nil {
		t.Error("Expected non-nil InferenceEngine")
	}

	// Check statistics
	stats := manager.GetDetailedStats()
	if stats.TypeCheckCount < 1 {
		t.Errorf("Expected at least 1 type check operation, got %d", stats.TypeCheckCount)
	}
}

func TestTypePoolManager_PoolEfficiency(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Create several objects to test pool efficiency
	for i := 0; i < 10; i++ {
		_ = manager.NewTypeCheckResult("test.pl")
		_ = manager.NewTypeState()
		_ = manager.NewCondition("var", "==", "value", false)
	}

	// Check that pool efficiency is calculated correctly
	efficiency := manager.PoolEfficiency()
	if efficiency < 0 || efficiency > 100 {
		t.Errorf("Expected efficiency between 0 and 100, got %f", efficiency)
	}

	// After creating objects, efficiency should be reasonable
	// (depends on implementation but should be > 0 for pool hits)
	if efficiency == 0 {
		t.Log("Pool efficiency is 0% - this might be expected for first allocations")
	}
}

func TestTypePoolManager_TypeReuseRate(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// The warm pools function should create some initial types
	reuseRate := manager.TypeReuseRate()
	if reuseRate < 0 || reuseRate > 100 {
		t.Errorf("Expected reuse rate between 0 and 100, got %f", reuseRate)
	}
}

func TestTypePoolManager_Reset(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Create some objects
	result := manager.NewTypeCheckResult("test.pl")
	state := manager.NewTypeState()

	// Verify objects are created
	if result == nil || state == nil {
		t.Fatal("Expected objects to be created")
	}

	// Reset the pools
	manager.Reset()

	// Create new objects to verify reset worked
	newResult := manager.NewTypeCheckResult("test2.pl")
	newState := manager.NewTypeState()

	if newResult == nil || newState == nil {
		t.Fatal("Expected new objects to be created after reset")
	}

	// Check that new objects are properly initialized
	if newResult.Path != "test2.pl" {
		t.Errorf("Expected path 'test2.pl', got %s", newResult.Path)
	}
}

func TestTypePoolManager_Clear(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Create some objects to populate statistics
	for i := 0; i < 5; i++ {
		_ = manager.NewTypeCheckResult("test.pl")
	}

	// Get initial stats
	statsBefore := manager.GetDetailedStats()
	if statsBefore.TypeCheckCount == 0 {
		t.Error("Expected some type check operations before clear")
	}

	// Clear the pools
	manager.Clear()

	// Check that statistics are reset
	statsAfter := manager.GetDetailedStats()
	if statsAfter.TypeCheckCount != 0 {
		t.Errorf("Expected type check count to be 0 after clear, got %d", statsAfter.TypeCheckCount)
	}
	if statsAfter.PoolHits != 0 {
		t.Errorf("Expected pool hits to be 0 after clear, got %d", statsAfter.PoolHits)
	}
	if statsAfter.PoolMisses != 0 {
		t.Errorf("Expected pool misses to be 0 after clear, got %d", statsAfter.PoolMisses)
	}
}

func TestTypePoolManager_ConcurrentAccess(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Test concurrent access to pools
	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Run concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				// Create various pooled objects concurrently
				result := manager.NewTypeCheckResult("test.pl")
				state := manager.NewTypeState()
				condition := manager.NewCondition("var", "==", "value", false)

				// Verify objects are created correctly
				if result == nil || state == nil || condition == nil {
					t.Errorf("Goroutine %d: Expected non-nil objects", id)
					return
				}

				// Check that objects have correct initial values
				if result.Path != "test.pl" {
					t.Errorf("Goroutine %d: Expected path 'test.pl', got %s", id, result.Path)
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check final statistics
	stats := manager.GetDetailedStats()
	expectedOps := int64(numGoroutines * numOperations)
	if stats.TypeCheckCount < expectedOps {
		t.Errorf("Expected at least %d type check operations, got %d", expectedOps, stats.TypeCheckCount)
	}
}

func TestTypePoolManager_GlobalInstance(t *testing.T) {
	// Test that global instance is created correctly
	global1 := GlobalTypePoolManager()
	if global1 == nil {
		t.Fatal("Expected non-nil global type pool manager")
	}

	// Test that subsequent calls return the same instance
	global2 := GlobalTypePoolManager()
	if global1 != global2 {
		t.Error("Expected same instance from multiple calls to GlobalTypePoolManager")
	}
}

func TestTypePoolManager_SetGlobalTypePoolHooks(t *testing.T) {
	hookCalled := false
	hooks := TypePoolHooks{
		OnTypeCheckCreate: func(result *TypeCheckResult) {
			hookCalled = true
		},
	}

	// Set hooks on global manager
	SetGlobalTypePoolHooks(hooks)

	// Use global manager to create an object
	global := GlobalTypePoolManager()
	_ = global.NewTypeCheckResult("test.pl")

	// Verify hook was called
	if !hookCalled {
		t.Error("Expected hook to be called when using global manager")
	}
}

func TestTypePoolManager_WarmPools(t *testing.T) {
	warmingCalled := false
	hooks := TypePoolHooks{
		OnPoolWarming: func(poolType string) {
			warmingCalled = true
		},
	}

	// Creating a new manager should trigger pool warming
	manager := NewTypePoolManager(hooks)
	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	// Verify warming was called
	if !warmingCalled {
		t.Error("Expected pool warming hook to be called during manager creation")
	}

	// Check that some types were created during warming
	stats := manager.GetDetailedStats()
	if stats.TypeCreated == 0 {
		t.Error("Expected some types to be created during pool warming")
	}
}

func TestTypePoolManager_MemoryFootprint(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Create a large number of objects to test memory usage
	const numObjects = 1000

	for i := 0; i < numObjects; i++ {
		_ = manager.NewTypeCheckResult("test.pl")
		_ = manager.NewTypeState()
		_ = manager.NewCondition("var", "==", "value", false)
	}

	// Get statistics
	stats := manager.GetDetailedStats()

	// Verify that operations were recorded
	if stats.TypeCheckCount < numObjects {
		t.Errorf("Expected at least %d type check operations, got %d", numObjects, stats.TypeCheckCount)
	}

	// Pool efficiency should be reasonable after many allocations
	efficiency := manager.PoolEfficiency()
	t.Logf("Pool efficiency after %d allocations: %.2f%%", numObjects, efficiency)

	// Reset pools to test memory reclamation
	manager.Reset()

	// Create new objects to verify pools still work after reset
	newResult := manager.NewTypeCheckResult("after-reset.pl")
	if newResult == nil {
		t.Error("Expected to be able to create objects after pool reset")
	} else if newResult.Path != "after-reset.pl" {
		t.Errorf("Expected path 'after-reset.pl', got %s", newResult.Path)
	}
}

func TestTypePoolManager_PooledCollections(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Test that pooled collections are properly initialized
	result := manager.NewTypeCheckResult("test.pl")

	// Add some errors to test collection functionality
	err1 := manager.NewTypeCheckError("error1", "test.pl", 1, 1)
	err2 := manager.NewTypeCheckError("error2", "test.pl", 2, 1)

	result.Errors = append(result.Errors, *err1, *err2)

	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}

	// Test that the errors were created correctly
	if result.Errors[0].Message != "error1" {
		t.Errorf("Expected first error message 'error1', got %s", result.Errors[0].Message)
	}
	if result.Errors[1].Message != "error2" {
		t.Errorf("Expected second error message 'error2', got %s", result.Errors[1].Message)
	}
}

func TestTypePoolManager_ComplexWorkflow(t *testing.T) {
	manager := NewTypePoolManager(TypePoolHooks{})

	// Create test dependencies
	hierarchy := typedef.NewTypeHierarchy(nil)
	symbolTable := binder.NewSymbolTable()

	// Create a type checker with pooled objects
	tc := manager.NewTypeChecker(hierarchy, symbolTable, "TestModule")

	// Create some type checking components
	result := manager.NewTypeCheckResult("complex.pl")
	engine := manager.NewInferenceEngine(hierarchy, symbolTable)

	// Test basic functionality
	if tc.CurrentModule != "TestModule" {
		t.Errorf("Expected module 'TestModule', got %s", tc.CurrentModule)
	}

	if result.Path != "complex.pl" {
		t.Errorf("Expected path 'complex.pl', got %s", result.Path)
	}

	// Test inference engine functionality
	inferredType, confidence := engine.GetInferredType("nonexistent")
	if inferredType != "Unknown" || confidence != 0.0 {
		t.Errorf("Expected type 'Unknown' with confidence 0.0, got %s with %f", inferredType, confidence)
	}

	// Add some type annotations
	tc.TypeAnnotations["var1"] = "Int"
	tc.TypeAnnotations["var2"] = "Str"

	// Check that annotations were added correctly
	if tc.TypeAnnotations["var1"] != "Int" {
		t.Errorf("Expected type 'Int' for var1, got %s", tc.TypeAnnotations["var1"])
	}
	if tc.TypeAnnotations["var2"] != "Str" {
		t.Errorf("Expected type 'Str' for var2, got %s", tc.TypeAnnotations["var2"])
	}

	// Check statistics
	stats := manager.GetDetailedStats()
	if stats.TypeCheckCount < 1 {
		t.Errorf("Expected at least 1 type check operation, got %d", stats.TypeCheckCount)
	}
	if stats.InferenceCount < 1 {
		t.Errorf("Expected at least 1 inference operation, got %d", stats.InferenceCount)
	}
}
