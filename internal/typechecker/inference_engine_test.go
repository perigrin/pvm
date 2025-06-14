// ABOUTME: Comprehensive test suite for the advanced type inference engine
// ABOUTME: Tests data flow analysis, context-aware inference, and usage patterns

package typechecker

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/typedef"
)

// MockNode implements ast.Node for testing
type MockNode struct {
	NodeType  string
	NodeText  string
	StartPos  ast.Position
	Children_ []ast.Node
	Parent_   ast.Node
}

func (m *MockNode) Type() string              { return m.NodeType }
func (m *MockNode) Text() string              { return m.NodeText }
func (m *MockNode) Start() ast.Position       { return m.StartPos }
func (m *MockNode) End() ast.Position         { return m.StartPos }
func (m *MockNode) Children() []ast.Node      { return m.Children_ }
func (m *MockNode) Parent() ast.Node          { return m.Parent_ }
func (m *MockNode) Child(int) ast.Node        { return nil }
func (m *MockNode) ChildCount() int           { return len(m.Children_) }
func (m *MockNode) NamedChild(int) ast.Node   { return nil }
func (m *MockNode) NamedChildCount() int      { return 0 }
func (m *MockNode) NextSibling() ast.Node     { return nil }
func (m *MockNode) PrevSibling() ast.Node     { return nil }
func (m *MockNode) String() string            { return m.NodeText }
func (m *MockNode) SetParent(parent ast.Node) { m.Parent_ = parent }

// MockAST implements ast.AST for testing
type MockAST struct {
	Path            string
	Root            ast.Node
	TypeAnnotations []*ast.TypeAnnotation
	Errors          []error
}

func (m *MockAST) RootNode() ast.Node { return m.Root }

func TestNewInferenceEngine(t *testing.T) {
	// Removed sampling to enable test in regular runs
	storage, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(storage)
	engine := NewInferenceEngine(hierarchy, nil)

	if engine == nil {
		t.Fatal("NewInferenceEngine returned nil")
	}

	if engine.TypeHierarchy != hierarchy {
		t.Error("TypeHierarchy not set correctly")
	}

	if engine.ConfidenceThreshold != 0.7 {
		t.Errorf("Expected confidence threshold 0.7, got %f", engine.ConfidenceThreshold)
	}

	if len(engine.InferredTypes) != 0 {
		t.Error("InferredTypes should be empty initially")
	}

	// Check that all analyzers are initialized
	if engine.DataFlowAnalyzer == nil {
		t.Error("DataFlowAnalyzer not initialized")
	}
	if engine.ContextAnalyzer == nil {
		t.Error("ContextAnalyzer not initialized")
	}
	if engine.UsagePatternAnalyzer == nil {
		t.Error("UsagePatternAnalyzer not initialized")
	}
	if engine.TypePropagator == nil {
		t.Error("TypePropagator not initialized")
	}
}

func TestContextRules(t *testing.T) {
	storage, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(storage)
	engine := NewInferenceEngine(hierarchy, nil)

	// Test scalar context rule
	scalarRule, exists := engine.ContextAnalyzer.ContextRules["scalar"]
	if !exists {
		t.Fatal("Scalar context rule not found")
	}

	// Test array in scalar context becomes Int (count)
	arrayNode := &MockNode{NodeType: "array", NodeText: "@array"}
	scalarContext := PerlContext{Type: "scalar"}
	inferredType := scalarRule.InferType(arrayNode, scalarContext)
	if inferredType != "Int" {
		t.Errorf("Expected array in scalar context to infer Int, got %s", inferredType)
	}

	// Test list context rule
	listRule, exists := engine.ContextAnalyzer.ContextRules["list"]
	if !exists {
		t.Fatal("List context rule not found")
	}

	// Test scalar in list context becomes Array
	scalarNode := &MockNode{NodeType: "scalar", NodeText: "$scalar"}
	listContext := PerlContext{Type: "list"}
	inferredType = listRule.InferType(scalarNode, listContext)
	if inferredType != "Array" {
		t.Errorf("Expected scalar in list context to infer Array, got %s", inferredType)
	}

	// Test numeric context rule
	numericRule, exists := engine.ContextAnalyzer.ContextRules["numeric"]
	if !exists {
		t.Fatal("Numeric context rule not found")
	}

	numericContext := PerlContext{Type: "numeric"}
	inferredType = numericRule.InferType(scalarNode, numericContext)
	if inferredType != "Num" {
		t.Errorf("Expected numeric context to infer Num, got %s", inferredType)
	}

	// Test string context rule
	stringRule, exists := engine.ContextAnalyzer.ContextRules["string"]
	if !exists {
		t.Fatal("String context rule not found")
	}

	stringContext := PerlContext{Type: "string"}
	inferredType = stringRule.InferType(scalarNode, stringContext)
	if inferredType != "Str" {
		t.Errorf("Expected string context to infer Str, got %s", inferredType)
	}
}

func TestUsagePatterns(t *testing.T) {
	storage, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(storage)
	engine := NewInferenceEngine(hierarchy, nil)

	tests := []struct {
		name         string
		nodeText     string
		expectedType string
		shouldMatch  bool
	}{
		// Array operations
		{"push operation", "push @array, $item", "ArrayRef", true},
		{"pop operation", "pop @array", "ArrayRef", true},
		{"array deref", "@{$arrayref}", "ArrayRef", true},
		{"shift operation", "shift @array", "ArrayRef", true},
		{"unshift operation", "unshift @array, $item", "ArrayRef", true},

		// Hash operations
		{"keys operation", "keys %hash", "HashRef", true},
		{"values operation", "values %hash", "HashRef", true},
		{"exists operation", "exists $hash{key}", "HashRef", true},
		{"hash deref", "%{$hashref}", "HashRef", true},

		// Numeric operations
		{"addition", "$a + $b", "Int", true},
		{"subtraction", "$a - $b", "Int", true},
		{"multiplication", "$a * $b", "Int", true},
		{"division", "$a / $b", "Int", true},
		{"modulo", "$a % $b", "Int", true},
		{"increment", "$a++", "Int", true},
		{"decrement", "$a--", "Int", true},
		{"power", "$a ** $b", "Int", true},
		{"comparison", "$a < $b", "Int", true},
		{"decimal operation", "$a + 3.14", "Num", true},

		// String operations
		{"regex match", "$str =~ /pattern/", "Str", true},
		{"regex no match", "$str !~ /pattern/", "Str", true},
		{"substr", "substr($str, 0, 5)", "Str", true},
		{"length", "length($str)", "Str", true},
		{"uppercase", "uc($str)", "Str", true},
		{"lowercase", "lc($str)", "Str", true},

		// Object method calls
		{"method call", "$obj->method()", "Object", true},
		{"constructor", "Class->new()", "Class", true},
		{"chained methods", "$obj->method1()->method2()", "Object", true},

		// File operations
		{"open file", "open($fh, '<', $file)", "FileHandle", true},
		{"close file", "close($fh)", "FileHandle", true},

		// Non-matching cases
		{"simple assignment", "$a = 5", "", false},
		{"hash arrow", "$hash => $value", "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := &MockNode{NodeText: test.nodeText}
			matched := false
			var inferredType string

			for _, pattern := range engine.UsagePatternAnalyzer.Patterns {
				if pattern.Matcher(node) {
					matched = true
					inferredType, _ = pattern.InferType(node)
					break
				}
			}

			if test.shouldMatch && !matched {
				t.Errorf("Expected pattern to match for %q", test.nodeText)
			}

			if !test.shouldMatch && matched {
				t.Errorf("Expected pattern not to match for %q", test.nodeText)
			}

			if test.shouldMatch && matched {
				// For numeric operations with decimals, expect Num
				if test.expectedType == "Num" && inferredType != "Num" {
					t.Errorf("Expected %s, got %s for %q", test.expectedType, inferredType, test.nodeText)
				}
				// For integer operations, expect Int or Num
				if test.expectedType == "Int" && (inferredType != "Int" && inferredType != "Num") {
					t.Errorf("Expected %s or Num, got %s for %q", test.expectedType, inferredType, test.nodeText)
				}
				// For other types, exact match
				if test.expectedType != "Int" && test.expectedType != "Num" && inferredType != test.expectedType {
					t.Errorf("Expected %s, got %s for %q", test.expectedType, inferredType, test.nodeText)
				}
			}
		})
	}
}

func TestInferenceRecording(t *testing.T) {
	storage, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(storage)
	engine := NewInferenceEngine(hierarchy, nil)

	// Record a simple inference
	pos := ast.Position{Line: 1, Column: 1}
	engine.recordInference("$var", "Int", 0.8, InferenceFromPattern, pos)

	// Check if inference was recorded
	if len(engine.InferredTypes) != 1 {
		t.Errorf("Expected 1 inferred type, got %d", len(engine.InferredTypes))
	}

	info, exists := engine.InferredTypes["$var"]
	if !exists {
		t.Fatal("Inference for $var not found")
	}

	if info.Type != "Int" {
		t.Errorf("Expected type Int, got %s", info.Type)
	}

	if info.Confidence != 0.8 {
		t.Errorf("Expected confidence 0.8, got %f", info.Confidence)
	}

	if len(info.Sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(info.Sources))
	}

	// Record a stronger inference for the same variable
	engine.recordInference("$var", "Str", 0.9, InferenceFromUsage, pos)

	// Check if stronger inference overwrote weaker one
	info = engine.InferredTypes["$var"]
	if info.Type != "Str" {
		t.Errorf("Expected stronger inference to overwrite, got type %s", info.Type)
	}

	if info.Confidence != 0.9 {
		t.Errorf("Expected stronger confidence, got %f", info.Confidence)
	}

	if len(info.Sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(info.Sources))
	}

	// Record a weaker inference for the same variable
	engine.recordInference("$var", "Any", 0.5, InferenceFromContext, pos)

	// Check that stronger inference is maintained
	info = engine.InferredTypes["$var"]
	if info.Type != "Str" {
		t.Errorf("Expected stronger inference to be maintained, got type %s", info.Type)
	}

	if len(info.Sources) != 3 {
		t.Errorf("Expected 3 sources, got %d", len(info.Sources))
	}
}

func TestContextDetermination(t *testing.T) {
	storage, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(storage)
	engine := NewInferenceEngine(hierarchy, nil)

	tests := []struct {
		name         string
		nodeType     string
		nodeText     string
		expectedType string
	}{
		{"explicit scalar", "scalar", "scalar @array", "scalar"},
		{"list context", "list", "@list = (1, 2, 3)", "list"},
		{"array type", "array", "@array", "list"},
		{"numeric operation", "expression", "$a + $b", "numeric"},
		{"comparison", "expression", "$a < $b", "numeric"},
		{"default context", "variable", "$var", "scalar"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := &MockNode{NodeType: test.nodeType, NodeText: test.nodeText}
			context := engine.determineContext(node)

			if context.Type != test.expectedType {
				t.Errorf("Expected context %s, got %s for %q", test.expectedType, context.Type, test.nodeText)
			}
		})
	}
}

func TestGetInferredType(t *testing.T) {
	storage, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(storage)
	engine := NewInferenceEngine(hierarchy, nil)

	// Test with no inference
	inferredType, confidence := engine.GetInferredType("$unknown")
	if inferredType != "Unknown" || confidence != 0.0 {
		t.Errorf("Expected Unknown with 0.0 confidence for unknown variable, got %s with %f", inferredType, confidence)
	}

	// Record an inference above threshold
	pos := ast.Position{Line: 1, Column: 1}
	engine.recordInference("$var1", "Int", 0.8, InferenceFromPattern, pos)

	inferredType, confidence = engine.GetInferredType("$var1")
	if inferredType != "Int" || confidence != 0.8 {
		t.Errorf("Expected Int with 0.8 confidence, got %s with %f", inferredType, confidence)
	}

	// Record an inference below threshold
	engine.recordInference("$var2", "Str", 0.5, InferenceFromPattern, pos)

	inferredType, confidence = engine.GetInferredType("$var2")
	if inferredType != "Unknown" || confidence != 0.0 {
		t.Errorf("Expected Unknown with 0.0 confidence for below-threshold inference, got %s with %f", inferredType, confidence)
	}
}

func TestGetAllInferredTypes(t *testing.T) {
	storage, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(storage)
	engine := NewInferenceEngine(hierarchy, nil)

	pos := ast.Position{Line: 1, Column: 1}

	// Record some inferences
	engine.recordInference("$var1", "Int", 0.8, InferenceFromPattern, pos) // Above threshold
	engine.recordInference("$var2", "Str", 0.9, InferenceFromUsage, pos)   // Above threshold
	engine.recordInference("$var3", "Any", 0.5, InferenceFromContext, pos) // Below threshold

	allTypes := engine.GetAllInferredTypes()

	if len(allTypes) != 2 {
		t.Errorf("Expected 2 types above threshold, got %d", len(allTypes))
	}

	if allTypes["$var1"] != "Int" {
		t.Errorf("Expected $var1 to be Int, got %s", allTypes["$var1"])
	}

	if allTypes["$var2"] != "Str" {
		t.Errorf("Expected $var2 to be Str, got %s", allTypes["$var2"])
	}

	if _, exists := allTypes["$var3"]; exists {
		t.Error("$var3 should not be in results (below threshold)")
	}
}

func TestAnalyzeUsagePatterns(t *testing.T) {
	storage, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(storage)
	engine := NewInferenceEngine(hierarchy, nil)

	// Create a simple AST with array operations
	root := &MockNode{
		NodeType: "program",
		NodeText: "push @array, $item",
		Children_: []ast.Node{
			&MockNode{
				NodeType: "expression",
				NodeText: "push @array, $item",
				StartPos: ast.Position{Line: 1, Column: 1},
			},
		},
	}

	// Analyze patterns
	engine.analyzeUsagePatterns(root)

	// Check if inference was recorded
	if len(engine.InferredTypes) == 0 {
		t.Error("Expected at least one inference from pattern analysis")
	}

	// Look for array-related inference
	found := false
	for _, info := range engine.InferredTypes {
		if info.Type == "ArrayRef" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected ArrayRef inference from push operation")
	}
}

func TestInferenceReport(t *testing.T) {
	storage, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(storage)
	engine := NewInferenceEngine(hierarchy, nil)

	pos := ast.Position{Line: 1, Column: 1}

	// Record some inferences
	engine.recordInference("$var1", "Int", 0.8, InferenceFromPattern, pos)
	engine.recordInference("$var2", "Str", 0.9, InferenceFromUsage, pos)

	report := engine.GetInferenceReport()

	if report == "" {
		t.Error("Report should not be empty")
	}

	// Check that report contains expected content
	expectedContent := []string{
		"Type Inference Report",
		"Variable: $var1",
		"Inferred Type: Int",
		"confidence: 0.80",
		"Pattern",
		"Variable: $var2",
		"Inferred Type: Str",
		"confidence: 0.90",
		"Usage",
	}

	for _, content := range expectedContent {
		if !containsString(report, content) {
			t.Errorf("Report missing expected content: %s", content)
		}
	}
}

func TestInferTypesIntegration(t *testing.T) {
	storage, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(storage)
	engine := NewInferenceEngine(hierarchy, nil)

	// Create a mock AST
	mockAST := &MockAST{
		Root: &MockNode{
			NodeType: "program",
			NodeText: "my $count = @array; $obj->method();",
			Children_: []ast.Node{
				&MockNode{
					NodeType: "assignment",
					NodeText: "my $count = @array",
					StartPos: ast.Position{Line: 1, Column: 1},
				},
				&MockNode{
					NodeType: "method_call",
					NodeText: "$obj->method()",
					StartPos: ast.Position{Line: 1, Column: 20},
				},
			},
		},
	}

	// Convert MockAST to ast.AST
	realAST := &ast.AST{
		Path:            mockAST.Path,
		Root:            mockAST.Root,
		TypeAnnotations: mockAST.TypeAnnotations,
		Errors:          mockAST.Errors,
	}

	// Run full inference
	err = engine.InferTypes(realAST)
	if err != nil {
		t.Errorf("InferTypes failed: %v", err)
	}

	// Check that some inferences were made
	if len(engine.InferredTypes) == 0 {
		t.Error("Expected some type inferences to be made")
	}
}

func TestDataFlowComponents(t *testing.T) {
	analyzer := NewDataFlowAnalyzer()

	if analyzer == nil {
		t.Fatal("NewDataFlowAnalyzer returned nil")
	}

	if analyzer.DataFlowGraph == nil {
		t.Error("DataFlowGraph not initialized")
	}

	if analyzer.VariableStates == nil {
		t.Error("VariableStates not initialized")
	}
}

func TestContextAnalyzerComponents(t *testing.T) {
	analyzer := NewContextAnalyzer()

	if analyzer == nil {
		t.Fatal("NewContextAnalyzer returned nil")
	}

	if analyzer.ContextStack == nil {
		t.Error("ContextStack not initialized")
	}

	if analyzer.ContextRules == nil {
		t.Error("ContextRules not initialized")
	}
}

func TestUsagePatternAnalyzerComponents(t *testing.T) {
	analyzer := NewUsagePatternAnalyzer()

	if analyzer == nil {
		t.Fatal("NewUsagePatternAnalyzer returned nil")
	}

	if analyzer.Patterns == nil {
		t.Error("Patterns not initialized")
	}

	if analyzer.PatternCache == nil {
		t.Error("PatternCache not initialized")
	}
}

func TestTypePropagatorComponents(t *testing.T) {
	propagator := NewTypePropagator()

	if propagator == nil {
		t.Fatal("NewTypePropagator returned nil")
	}

	if propagator.PropagationRules == nil {
		t.Error("PropagationRules not initialized")
	}

	if propagator.TypeConstraints == nil {
		t.Error("TypeConstraints not initialized")
	}

	if propagator.SolvedTypes == nil {
		t.Error("SolvedTypes not initialized")
	}
}

func TestSourceTypeNames(t *testing.T) {
	tests := []struct {
		sourceType   InferenceSourceType
		expectedName string
	}{
		{InferenceFromAssignment, "Assignment"},
		{InferenceFromUsage, "Usage"},
		{InferenceFromMethodCall, "Method Call"},
		{InferenceFromOperator, "Operator"},
		{InferenceFromContext, "Context"},
		{InferenceFromPattern, "Pattern"},
		{InferenceFromDataFlow, "Data Flow"},
	}

	for _, test := range tests {
		name := getSourceTypeName(test.sourceType)
		if name != test.expectedName {
			t.Errorf("Expected %s, got %s for source type %v", test.expectedName, name, test.sourceType)
		}
	}

	// Test unknown type
	unknownType := InferenceSourceType(999)
	name := getSourceTypeName(unknownType)
	if name != "Unknown" {
		t.Errorf("Expected Unknown for unknown source type, got %s", name)
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests for performance verification
func BenchmarkInferenceEngine(b *testing.B) {
	storage, err := typedef.NewStorage()
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(storage)
	engine := NewInferenceEngine(hierarchy, nil)

	// Create a large mock AST
	children := make([]ast.Node, 100)
	for i := 0; i < 100; i++ {
		children[i] = &MockNode{
			NodeType: "expression",
			NodeText: "push @array, $item",
			StartPos: ast.Position{Line: i + 1, Column: 1},
		}
	}

	root := &MockNode{
		NodeType:  "program",
		NodeText:  "large program",
		Children_: children,
	}

	mockAST := &MockAST{Root: root}

	// Convert to ast.AST
	realAST := &ast.AST{
		Path:            mockAST.Path,
		Root:            mockAST.Root,
		TypeAnnotations: mockAST.TypeAnnotations,
		Errors:          mockAST.Errors,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.InferredTypes = make(map[string]*InferredTypeInfo) // Reset state
		engine.InferTypes(realAST)
	}
}

func BenchmarkPatternMatching(b *testing.B) {
	storage, err := typedef.NewStorage()
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	hierarchy := typedef.NewTypeHierarchy(storage)
	engine := NewInferenceEngine(hierarchy, nil)

	node := &MockNode{
		NodeType: "expression",
		NodeText: "push @array, $item",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.analyzeUsagePatterns(node)
	}
}
