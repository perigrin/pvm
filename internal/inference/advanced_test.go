// ABOUTME: Comprehensive tests for advanced inference features
// ABOUTME: Tests control flow, function signatures, contextual analysis, and method resolution

package inference

import (
	"testing"

	"tamarou.com/pvm/internal/types"
)

// Test control flow analysis
func TestFlowAnalyzer(t *testing.T) {
	engine := NewTypeInferenceEngine()
	flowAnalyzer := NewFlowAnalyzer(engine)

	// Test entering and exiting control flow
	state1 := flowAnalyzer.EnterControlFlow("if", nil)
	if state1 == nil {
		t.Fatal("EnterControlFlow returned nil")
	}

	if state1.BranchType != "if" {
		t.Errorf("Expected branch type 'if', got '%s'", state1.BranchType)
	}

	if state1.ConfidenceModifier != 0.95 {
		t.Errorf("Expected confidence modifier 0.95 for if, got %f", state1.ConfidenceModifier)
	}

	// Test nested control flow
	state2 := flowAnalyzer.EnterControlFlow("while", nil)
	if state2.Parent != state1 {
		t.Error("Nested state should have parent reference")
	}

	if state2.ConfidenceModifier != 0.80 {
		t.Errorf("Expected confidence modifier 0.80 for while, got %f", state2.ConfidenceModifier)
	}

	// Test exiting control flow
	exitedState := flowAnalyzer.ExitControlFlow()
	if exitedState != state2 {
		t.Error("Exited state should be the most recent state")
	}

	currentState := flowAnalyzer.GetCurrentFlowState()
	if currentState != state1 {
		t.Error("Current state should be the parent after exiting")
	}
}

func TestFlowStateMerging(t *testing.T) {
	engine := NewTypeInferenceEngine()
	flowAnalyzer := NewFlowAnalyzer(engine)

	// Create multiple flow states
	state1 := NewFlowState("if", nil)
	state1.VariableTypes["x"] = types.NewTypeInfo(types.NewIntType(), 0.9, types.SourceVariable)

	state2 := NewFlowState("elsif", nil)
	state2.VariableTypes["x"] = types.NewTypeInfo(types.NewStrType(), 0.8, types.SourceVariable)

	state3 := NewFlowState("else", nil)
	state3.VariableTypes["y"] = types.NewTypeInfo(types.NewBoolType(), 0.7, types.SourceContext)

	// Test merging
	states := []*FlowState{state1, state2, state3}
	merged := flowAnalyzer.MergeFlowStates(states)

	if merged == nil {
		t.Fatal("MergeFlowStates returned nil")
	}

	// Check that variable x has merged type (should be the most confident)
	if xType := merged.VariableTypes["x"]; xType == nil {
		t.Error("Merged state should contain variable x")
	} else if !xType.Type.Equals(types.NewIntType()) {
		t.Errorf("Expected merged x type to be Int (highest confidence), got %s", xType.Type.String())
	}

	// Check that variable y is present
	if yType := merged.VariableTypes["y"]; yType == nil {
		t.Error("Merged state should contain variable y")
	}

	// Check confidence reduction
	if merged.ConfidenceModifier != 0.80 {
		t.Errorf("Expected merged confidence modifier 0.80, got %f", merged.ConfidenceModifier)
	}
}

// Test function signature inference
func TestFunctionSignatureInferrer(t *testing.T) {
	engine := NewTypeInferenceEngine()
	inferrer := NewFunctionSignatureInferrer(engine)

	// Test recording usage patterns
	usage := &FunctionUsage{
		ArgumentTypes: []types.Type{types.NewStrType(), types.NewIntType()},
		ReturnContext: &ReturnContext{
			ExpectedType: types.NewStrType(),
			UsageType:    "assignment",
			Confidence:   0.8,
		},
		Confidence: 0.85,
	}

	inferrer.recordUsagePattern("test_function", usage)

	// Test signature creation from usage
	usages := inferrer.usagePatterns["test_function"]
	if len(usages) != 1 {
		t.Fatalf("Expected 1 usage pattern, got %d", len(usages))
	}

	signature := inferrer.createSignatureFromUsage("test_function", usages)
	if signature == nil {
		t.Fatal("createSignatureFromUsage returned nil")
	}

	if signature.Name != "test_function" {
		t.Errorf("Expected function name 'test_function', got '%s'", signature.Name)
	}

	if len(signature.ParameterTypes) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(signature.ParameterTypes))
	}

	if !signature.ParameterTypes[0].Equals(types.NewStrType()) {
		t.Errorf("Expected first parameter to be Str, got %s", signature.ParameterTypes[0].String())
	}

	if !signature.ReturnType.Equals(types.NewStrType()) {
		t.Errorf("Expected return type to be Str, got %s", signature.ReturnType.String())
	}
}

func TestBuiltinReturnTypes(t *testing.T) {
	engine := NewTypeInferenceEngine()
	inferrer := NewFunctionSignatureInferrer(engine)

	testCases := []struct {
		functionName string
		expectedType types.Type
	}{
		{"length", types.NewIntType()},
		{"substr", types.NewStrType()},
		{"defined", types.NewBoolType()},
		{"split", types.NewArrayRefType(types.NewStrType())},
		{"unknown_function", types.NewStrType()}, // Default
	}

	for _, tc := range testCases {
		returnType := inferrer.inferBuiltinReturnType(tc.functionName)
		if !returnType.Equals(tc.expectedType) {
			t.Errorf("Function %s: expected %s, got %s",
				tc.functionName, tc.expectedType.String(), returnType.String())
		}
	}
}

// Test contextual analysis
func TestContextualAnalyzer(t *testing.T) {
	engine := NewTypeInferenceEngine()
	analyzer := NewContextualAnalyzer(engine)

	// Test context determination
	testCases := []struct {
		context      ContextType
		expectedType types.Type
	}{
		{ScalarContext, types.NewStrType()},
		{NumericContext, types.NewNumType()},
		{BooleanContext, types.NewBoolType()},
		{StringContext, types.NewStrType()},
	}

	for _, tc := range testCases {
		analyzer.pushContext(tc.context)

		if analyzer.GetCurrentContext() != tc.context {
			t.Errorf("Expected context %s, got %s", tc.context.String(), analyzer.GetCurrentContext().String())
		}

		analyzer.popContext()
	}
}

func TestArrayVariableContextAnalysis(t *testing.T) {
	engine := NewTypeInferenceEngine()
	analyzer := NewContextualAnalyzer(engine)

	// Test array variable in different contexts
	testCases := []struct {
		context      ContextType
		expectedType func() types.Type
		description  string
	}{
		{ScalarContext, types.NewIntType, "Array in scalar context should return count (Int)"},
		{ListContext, func() types.Type { return types.NewArrayRefType(types.NewStrType()) }, "Array in list context should return ArrayRef"},
		{BooleanContext, types.NewBoolType, "Array in boolean context should return Bool"},
	}

	for _, tc := range testCases {
		analyzer.pushContext(tc.context)

		result, err := analyzer.analyzeArrayVariableInContext("test_array", nil)
		if err != nil {
			t.Errorf("Error analyzing array in %s context: %v", tc.context.String(), err)
			continue
		}

		expected := tc.expectedType()
		if !result.Equals(expected) {
			t.Errorf("%s: expected %s, got %s", tc.description, expected.String(), result.String())
		}

		analyzer.popContext()
	}
}

func TestBuiltinContextualTypes(t *testing.T) {
	engine := NewTypeInferenceEngine()
	analyzer := NewContextualAnalyzer(engine)

	testCases := []struct {
		functionName string
		context      ContextType
		expectedType types.Type
	}{
		{"split", ScalarContext, types.NewIntType()},
		{"split", ListContext, types.NewArrayRefType(types.NewStrType())},
		{"keys", ScalarContext, types.NewIntType()},
		{"keys", ListContext, types.NewArrayRefType(types.NewStrType())},
		{"grep", ScalarContext, types.NewIntType()},
		{"grep", ListContext, types.NewArrayRefType(types.NewStrType())},
	}

	for _, tc := range testCases {
		analyzer.pushContext(tc.context)

		result := analyzer.getBuiltinContextualType(tc.functionName)
		if result == nil {
			t.Errorf("Function %s in %s context: expected type, got nil", tc.functionName, tc.context.String())
			continue
		}

		if !result.Equals(tc.expectedType) {
			t.Errorf("Function %s in %s context: expected %s, got %s",
				tc.functionName, tc.context.String(), tc.expectedType.String(), result.String())
		}

		analyzer.popContext()
	}
}

// Test external hint provider
func TestExternalHintProvider(t *testing.T) {
	provider := NewExternalHintProvider([]string{})

	// Test creating core hints
	provider.CreateCoreHints()

	// Test getting function hints
	lengthHint := provider.GetFunctionTypeHint("length")
	if lengthHint == nil {
		t.Fatal("Expected length function hint, got nil")
	}

	if !lengthHint.Type.Equals(types.NewIntType()) {
		t.Errorf("Expected length to return Int, got %s", lengthHint.Type.String())
	}

	if lengthHint.Confidence < 0.9 {
		t.Errorf("Expected high confidence for core function, got %f", lengthHint.Confidence)
	}

	// Test adding custom hints
	provider.AddCustomHint("custom_func", types.NewBoolType(), 0.85)
	customHint := provider.GetFunctionTypeHint("custom_func")
	if customHint == nil {
		t.Fatal("Expected custom function hint, got nil")
	}

	if !customHint.Type.Equals(types.NewBoolType()) {
		t.Errorf("Expected custom function to return Bool, got %s", customHint.Type.String())
	}
}

func TestStringToTypeConversion(t *testing.T) {
	provider := NewExternalHintProvider([]string{})

	testCases := []struct {
		typeName     string
		expectedType types.Type
	}{
		{"Int", types.NewIntType()},
		{"Str", types.NewStrType()},
		{"Bool", types.NewBoolType()},
		{"Num", types.NewNumType()},
		{"ArrayRef[Str]", types.NewArrayRefType(types.NewStrType())},
		{"HashRef[Int]", types.NewHashRefType(types.NewIntType())},
		{"Unknown", types.NewStrType()}, // Default
	}

	for _, tc := range testCases {
		result := provider.stringToType(tc.typeName)
		if !result.Equals(tc.expectedType) {
			t.Errorf("Type conversion %s: expected %s, got %s",
				tc.typeName, tc.expectedType.String(), result.String())
		}
	}
}

// Test method resolver
func TestMethodResolver(t *testing.T) {
	engine := NewTypeInferenceEngine()
	provider := NewExternalHintProvider([]string{})
	resolver := NewMethodResolver(engine, provider)

	// Test basic object type management
	objectType := &BasicObjectType{name: "TestClass"}
	resolver.SetObjectType("obj1", objectType)

	retrievedType := resolver.GetObjectType("obj1")
	if retrievedType == nil {
		t.Fatal("Expected to retrieve object type, got nil")
	}

	if !retrievedType.Equals(objectType) {
		t.Errorf("Expected %s, got %s", objectType.String(), retrievedType.String())
	}
}

func TestBasicObjectType(t *testing.T) {
	obj1 := &BasicObjectType{name: "Class1"}
	obj2 := &BasicObjectType{name: "Class1"}
	obj3 := &BasicObjectType{name: "Class2"}

	// Test equality
	if !obj1.Equals(obj2) {
		t.Error("Objects with same class name should be equal")
	}

	if obj1.Equals(obj3) {
		t.Error("Objects with different class names should not be equal")
	}

	// Test string representation
	if obj1.String() != "Class1" {
		t.Errorf("Expected string 'Class1', got '%s'", obj1.String())
	}

	// Test type classification
	if obj1.IsBasic() {
		t.Error("Object types should not be basic")
	}

	if !obj1.IsComplex() {
		t.Error("Object types should be complex")
	}
}

func TestConstructorRecognition(t *testing.T) {
	engine := NewTypeInferenceEngine()
	provider := NewExternalHintProvider([]string{})
	resolver := NewMethodResolver(engine, provider)

	testCases := []struct {
		methodName string
		expected   bool
	}{
		{"new", true},
		{"create", true},
		{"build", true},
		{"init", true},
		{"make", true},
		{"destroy", false},
		{"get", false},
		{"set", false},
	}

	for _, tc := range testCases {
		result := resolver.isConstructorMethod(tc.methodName)
		if result != tc.expected {
			t.Errorf("Method %s: expected constructor recognition %v, got %v",
				tc.methodName, tc.expected, result)
		}
	}
}

func TestClassExtractionFromConstructor(t *testing.T) {
	engine := NewTypeInferenceEngine()
	provider := NewExternalHintProvider([]string{})
	resolver := NewMethodResolver(engine, provider)

	testCases := []struct {
		constructorName string
		expectedClass   string
	}{
		{"MyClass::new", "MyClass"},
		{"Package::Module::new", "Package::Module"},
		{"new MyClass", "MyClass"},
		{"create", "create"},
	}

	for _, tc := range testCases {
		result := resolver.extractClassFromConstructor(tc.constructorName)
		if result != tc.expectedClass {
			t.Errorf("Constructor %s: expected class '%s', got '%s'",
				tc.constructorName, tc.expectedClass, result)
		}
	}
}

// Integration test for the full inference engine
func TestAdvancedInferenceIntegration(t *testing.T) {
	engine := NewTypeInferenceEngine()

	// Test that all components can be created and work together
	flowAnalyzer := NewFlowAnalyzer(engine)
	functionInferrer := NewFunctionSignatureInferrer(engine)
	contextAnalyzer := NewContextualAnalyzer(engine)
	hintProvider := NewExternalHintProvider([]string{})
	methodResolver := NewMethodResolver(engine, hintProvider)

	// Initialize hints
	hintProvider.CreateCoreHints()

	// Test that components don't interfere with each other
	if flowAnalyzer == nil || functionInferrer == nil || contextAnalyzer == nil ||
		hintProvider == nil || methodResolver == nil {
		t.Fatal("Failed to create inference components")
	}

	// Test basic functionality integration
	flowState := flowAnalyzer.EnterControlFlow("if", nil)
	if flowState == nil {
		t.Error("Flow analyzer should work in integration")
	}

	signature := functionInferrer.createSignatureFromUsage("test", []*FunctionUsage{})
	if signature != nil {
		t.Error("Empty usage should return nil signature")
	}

	contextAnalyzer.pushContext(ScalarContext)
	if contextAnalyzer.GetCurrentContext() != ScalarContext {
		t.Error("Context analyzer should work in integration")
	}

	lengthHint := hintProvider.GetFunctionTypeHint("length")
	if lengthHint == nil {
		t.Error("Hint provider should have core hints in integration")
	}

	objectType := &BasicObjectType{name: "TestClass"}
	methodResolver.SetObjectType("test_obj", objectType)
	retrieved := methodResolver.GetObjectType("test_obj")
	if retrieved == nil {
		t.Error("Method resolver should work in integration")
	}

	t.Log("Advanced inference integration test completed successfully")
}

// Benchmark tests for performance
func BenchmarkFlowStateCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewFlowState("if", nil)
	}
}

func BenchmarkContextPushPop(b *testing.B) {
	engine := NewTypeInferenceEngine()
	analyzer := NewContextualAnalyzer(engine)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.pushContext(ScalarContext)
		analyzer.popContext()
	}
}

func BenchmarkTypeConversion(b *testing.B) {
	provider := NewExternalHintProvider([]string{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.stringToType("ArrayRef[Str]")
	}
}
