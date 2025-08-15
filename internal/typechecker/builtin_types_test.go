// ABOUTME: Tests for Perl built-in function type registry
// ABOUTME: Validates type definition parsing and function signature lookup

package typechecker

import (
	"testing"

	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/typedef"
)

func TestBuiltinTypeRegistry_BasicFunctions(t *testing.T) {
	registry, err := NewBuiltinTypeRegistry()
	if err != nil {
		t.Fatalf("Failed to create builtin type registry: %v", err)
	}

	testCases := []struct {
		funcName   string
		paramCount int
		expected   string
		desc       string
	}{
		{"ref", 1, "Str", "ref() should return Str"},
		{"defined", 1, "Bool", "defined() should return Bool"},
		{"exists", 2, "Bool", "exists() should return Bool"},
		{"keys", 1, "Array[Str]", "keys() should return Array[Str]"},
		{"values", 1, "Array[Any]", "values() should return Array[Any]"},
		{"length", 1, "Int", "length() should return Int"},
		{"substr", 2, "Str", "substr() with 2 params should return Str"},
		{"substr", 3, "Str", "substr() with 3 params should return Str"},
		{"int", 1, "Int", "int() should return Int"},
		{"time", 0, "Int", "time() should return Int"},
		{"split", 2, "Array[Str]", "split() should return Array[Str]"},
		{"join", 2, "Str", "join() should return Str"},
		{"sprintf", 1, "Str", "sprintf() should return Str (variadic)"},
	}

	for _, tc := range testCases {
		t.Run(tc.funcName, func(t *testing.T) {
			actual := registry.GetFunctionType(tc.funcName, tc.paramCount)
			if actual != tc.expected {
				t.Errorf("%s: expected %s, got %s", tc.desc, tc.expected, actual)
			}
		})
	}
}

func TestBuiltinTypeRegistry_IsBuiltinFunction(t *testing.T) {
	registry, err := NewBuiltinTypeRegistry()
	if err != nil {
		t.Fatalf("Failed to create builtin type registry: %v", err)
	}

	testCases := []struct {
		funcName string
		expected bool
		desc     string
	}{
		{"ref", true, "ref should be recognized as builtin"},
		{"defined", true, "defined should be recognized as builtin"},
		{"length", true, "length should be recognized as builtin"},
		{"decode_json", false, "decode_json should not be recognized as builtin (it's a library function)"},
		{"unknown_function", false, "unknown function should not be recognized"},
		{"custom_func", false, "custom function should not be recognized"},
	}

	for _, tc := range testCases {
		t.Run(tc.funcName, func(t *testing.T) {
			actual := registry.IsBuiltinFunction(tc.funcName)
			if actual != tc.expected {
				t.Errorf("%s: expected %v, got %v", tc.desc, tc.expected, actual)
			}
		})
	}
}

func TestBuiltinTypeRegistry_VariadicFunctions(t *testing.T) {
	registry, err := NewBuiltinTypeRegistry()
	if err != nil {
		t.Fatalf("Failed to create builtin type registry: %v", err)
	}

	// Test that variadic functions work with different parameter counts
	testCases := []struct {
		funcName   string
		paramCount int
		expected   string
		desc       string
	}{
		{"printf", 1, "Bool", "printf with 1 param should work"},
		{"printf", 3, "Bool", "printf with 3 params should work"},
		{"printf", 10, "Bool", "printf with many params should work"},
		{"push", 2, "Int", "push with 2 params should work"},
		{"push", 5, "Int", "push with many params should work"},
	}

	for _, tc := range testCases {
		t.Run(tc.funcName, func(t *testing.T) {
			actual := registry.GetFunctionType(tc.funcName, tc.paramCount)
			if actual != tc.expected {
				t.Errorf("%s: expected %s, got %s", tc.desc, tc.expected, actual)
			}
		})
	}
}

func TestBuiltinTypeRegistry_LibraryFunctions(t *testing.T) {
	registry, err := NewBuiltinTypeRegistry()
	if err != nil {
		t.Fatalf("Failed to create builtin type registry: %v", err)
	}

	// Test common CPAN module functions that are often treated as built-ins
	testCases := []struct {
		funcName string
		expected string
		desc     string
	}{
		{"decode_json", "", "decode_json should return empty (library function, not builtin)"},
		{"encode_json", "", "encode_json should return empty (library function, not builtin)"},
		{"selectrow_hashref", "", "selectrow_hashref should return empty (library function, not builtin)"},
		{"slurp", "", "slurp should return empty (library function, not builtin)"},
		{"first", "", "List::Util::first should return empty (library function, not builtin)"},
		{"max", "", "List::Util::max should return empty (library function, not builtin)"},
	}

	for _, tc := range testCases {
		t.Run(tc.funcName, func(t *testing.T) {
			actual := registry.GetFunctionType(tc.funcName, 1) // Using 1 param for simplicity
			if actual != tc.expected {
				t.Errorf("%s: expected %s, got %s", tc.desc, tc.expected, actual)
			}
		})
	}
}

func TestBuiltinTypeRegistry_GetSignatures(t *testing.T) {
	registry, err := NewBuiltinTypeRegistry()
	if err != nil {
		t.Fatalf("Failed to create builtin type registry: %v", err)
	}

	// Test that we can get all signatures for a function
	signatures := registry.GetFunctionSignatures("substr")
	if len(signatures) == 0 {
		t.Error("Expected to find signatures for substr(), but got none")
	}

	// Verify signature string formatting works
	for _, sig := range signatures {
		sigStr := sig.GetTypeSignatureString()
		if sigStr == "" {
			t.Error("Expected non-empty signature string")
		}
		t.Logf("substr signature: %s", sigStr)
	}
}

func TestBuiltinTypeRegistry_ListAllBuiltins(t *testing.T) {
	registry, err := NewBuiltinTypeRegistry()
	if err != nil {
		t.Fatalf("Failed to create builtin type registry: %v", err)
	}

	builtins := registry.ListAllBuiltins()
	if len(builtins) == 0 {
		t.Error("Expected to find built-in functions, but got none")
	}

	// Should include some core Perl functions
	expectedFunctions := []string{"ref", "defined", "length", "keys", "values"}
	for _, expected := range expectedFunctions {
		found := false
		for _, builtin := range builtins {
			if builtin == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find %s in builtin list, but didn't", expected)
		}
	}

	t.Logf("Found %d built-in functions", len(builtins))
}

func TestFlowAnalyzer_WithBuiltinRegistry(t *testing.T) {
	// Test that FlowAnalyzer properly uses the builtin registry
	typeStore, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create type store: %v", err)
	}

	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	tc := NewTypeChecker(hierarchy, symbolTable, "test")

	analyzer := NewFlowAnalyzer(tc)

	if analyzer.BuiltinTypes == nil {
		t.Error("Expected FlowAnalyzer to have BuiltinTypes initialized")
	}

	// Test that the analyzer can use the registry for type inference
	returnType := analyzer.inferBuiltinFunctionType("ref", nil)
	if returnType != "Str" {
		t.Errorf("Expected ref() to return Str, got %s", returnType)
	}

	returnType = analyzer.inferBuiltinFunctionType("defined", nil)
	if returnType != "Bool" {
		t.Errorf("Expected defined() to return Bool, got %s", returnType)
	}
}
