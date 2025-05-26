// ABOUTME: Test file for enhanced module introspection capabilities
// ABOUTME: Tests OOP framework detection, dynamic method analysis, and type inference

package parser

import (
	"testing"
)

func TestEnhancedIntrospector(t *testing.T) {
	introspector, err := NewModuleIntrospector()
	if err != nil {
		t.Fatalf("Failed to create introspector: %v", err)
	}

	// Test OOP framework detection
	t.Run("OOPFrameworkDetection", func(t *testing.T) {
		testDetectOOPFrameworks(t, introspector)
	})

	// Test dynamic method detection
	t.Run("DynamicMethodDetection", func(t *testing.T) {
		testDynamicMethodDetection(t, introspector)
	})

	// Test variable analysis
	t.Run("VariableAnalysis", func(t *testing.T) {
		testVariableAnalysis(t, introspector)
	})

	// Test data structure analysis
	t.Run("DataStructureAnalysis", func(t *testing.T) {
		testDataStructureAnalysis(t, introspector)
	})
}

func testDetectOOPFrameworks(t *testing.T, introspector *ModuleIntrospector) {
	// Test Moose detection
	mooseCode := `
package MyClass;
use Moose;

has 'name' => (
    is  => 'rw',
    isa => 'Str',
    required => 1,
);

has 'age' => (
    is  => 'rw',
    isa => 'Int',
    default => 0,
);

no Moose;
1;
`

	// Create a mock AST node for testing
	mockNode := &MockNode{text: mooseCode}
	mockAST := &AST{Root: mockNode}

	result := &ModuleIntrospectionResult{
		Packages: make(map[string]*PackageInfo),
	}

	// Test framework detection
	introspector.detectOOPFrameworks(mockAST, result)

	// Verify Moose was detected
	found := false
	for _, framework := range result.DetectedFrameworks {
		if framework == "Moose" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Failed to detect Moose framework")
	}
}

func testDynamicMethodDetection(t *testing.T, introspector *ModuleIntrospector) {
	// Test Class::Accessor method generation
	accessorCode := `
package MyClass;
use base 'Class::Accessor';

__PACKAGE__->mk_accessors(qw(name age email));

1;
`

	mockNode := &MockNode{text: accessorCode}
	mockAST := &AST{Root: mockNode}

	result := &ModuleIntrospectionResult{
		Packages: make(map[string]*PackageInfo),
	}

	// Test dynamic method detection
	introspector.analyzeDynamicMethods(mockAST, result)

	// Check if accessor methods were detected
	if pkg, exists := result.Packages["main"]; exists {
		// The mk_accessors pattern should generate methods
		if len(pkg.Methods) == 0 {
			t.Error("Failed to detect dynamically generated accessor methods")
		}
	}
}

func testVariableAnalysis(t *testing.T, introspector *ModuleIntrospector) {
	// Test variable extraction
	variableCode := "my $name = 'test';"
	mockNode := &MockNode{text: variableCode}

	varInfo := introspector.extractVariableInfo(mockNode)
	if varInfo == nil {
		t.Error("Failed to extract variable information")
		return
	}

	if varInfo.Name != "$name" {
		t.Errorf("Expected variable name '$name', got '%s'", varInfo.Name)
	}

	if varInfo.Scope != "my" {
		t.Errorf("Expected scope 'my', got '%s'", varInfo.Scope)
	}

	if varInfo.InitialValue != "'test'" {
		t.Errorf("Expected initial value 'test', got '%s'", varInfo.InitialValue)
	}
}

func testDataStructureAnalysis(t *testing.T, introspector *ModuleIntrospector) {
	// Test data structure type inference with individual patterns
	testCases := []struct {
		code         string
		expectedType string
		varName      string
	}{
		{"my $arrayref = [];", "ArrayRef", "$arrayref"},
		{"my $hashref = {};", "HashRef", "$hashref"},
		{"my $dbh = DBI->connect($dsn, $user, $pass);", "DBI::db", "$dbh"},
	}

	for _, tc := range testCases {
		// Create individual nodes for each test case
		mockNode := &MockNode{text: tc.code}
		mockAST := &AST{Root: mockNode}

		result := &ModuleIntrospectionResult{
			DataStructures: make(map[string]*DataStructureInfo),
		}

		// Test data structure analysis
		introspector.analyzeDataStructures(mockAST, result)

		// Check if the expected type was detected
		found := false
		for _, dataStruct := range result.DataStructures {
			if dataStruct.Type == tc.expectedType && dataStruct.Name == tc.varName {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Failed to detect %s data structure for code: %s", tc.expectedType, tc.code)
			// Debug: Print what was actually detected
			t.Logf("Detected data structures: %+v", result.DataStructures)
		}
	}
}

// MockNode implements the Node interface for testing
type MockNode struct {
	text     string
	children []Node
}

func (m *MockNode) Type() string {
	return "mock_node"
}

func (m *MockNode) Text() string {
	return m.text
}

func (m *MockNode) Children() []Node {
	return m.children
}

func (m *MockNode) Start() Position {
	return Position{Line: 1, Column: 1, Offset: 0}
}

func (m *MockNode) End() Position {
	return Position{Line: 1, Column: len(m.text), Offset: len(m.text)}
}

func TestFrameworkPatternMatching(t *testing.T) {
	// Test framework pattern matching
	introspector, err := NewModuleIntrospector()
	if err != nil {
		t.Fatalf("Failed to create introspector: %v", err)
	}

	testCases := []struct {
		name              string
		code              string
		expectedFramework string
	}{
		{
			name:              "Moose",
			code:              "use Moose;",
			expectedFramework: "Moose",
		},
		{
			name:              "Moo",
			code:              "use Moo;",
			expectedFramework: "Moo",
		},
		{
			name:              "Class::Tiny",
			code:              "use Class::Tiny qw(name age);",
			expectedFramework: "Class::Tiny",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			framework := introspector.detectFrameworkFromUse(tc.code)
			if framework != tc.expectedFramework {
				t.Errorf("Expected framework '%s', got '%s'", tc.expectedFramework, framework)
			}
		})
	}
}

func TestImportInfoExtraction(t *testing.T) {
	introspector, err := NewModuleIntrospector()
	if err != nil {
		t.Fatalf("Failed to create introspector: %v", err)
	}

	testCases := []struct {
		name            string
		code            string
		expectedModule  string
		expectedSymbols []string
		expectedVersion string
	}{
		{
			name:           "Simple use",
			code:           "use Data::Dumper;",
			expectedModule: "Data::Dumper",
		},
		{
			name:            "Use with version",
			code:            "use List::Util '1.33';",
			expectedModule:  "List::Util",
			expectedVersion: "1.33",
		},
		{
			name:            "Use with qw",
			code:            "use List::Util qw(first reduce);",
			expectedModule:  "List::Util",
			expectedSymbols: []string{"first", "reduce"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockNode := &MockNode{text: tc.code}
			importInfo := introspector.extractImportInfo(mockNode)

			if importInfo == nil {
				t.Error("Failed to extract import information")
				return
			}

			if importInfo.ModuleName != tc.expectedModule {
				t.Errorf("Expected module '%s', got '%s'", tc.expectedModule, importInfo.ModuleName)
			}

			if tc.expectedVersion != "" && importInfo.Version != tc.expectedVersion {
				t.Errorf("Expected version '%s', got '%s'", tc.expectedVersion, importInfo.Version)
			}

			if len(tc.expectedSymbols) > 0 {
				if len(importInfo.ImportedSymbols) != len(tc.expectedSymbols) {
					t.Errorf("Expected %d symbols, got %d", len(tc.expectedSymbols), len(importInfo.ImportedSymbols))
					return
				}

				for i, expected := range tc.expectedSymbols {
					if i >= len(importInfo.ImportedSymbols) || importInfo.ImportedSymbols[i] != expected {
						t.Errorf("Expected symbol '%s', got '%s'", expected, importInfo.ImportedSymbols[i])
					}
				}
			}
		})
	}
}

func TestTypeInferenceRules(t *testing.T) {
	introspector, err := NewModuleIntrospector()
	if err != nil {
		t.Fatalf("Failed to create introspector: %v", err)
	}

	testCases := []struct {
		name         string
		code         string
		expectedType string
	}{
		{
			name:         "Array reference",
			code:         "my $arr = [];",
			expectedType: "ArrayRef",
		},
		{
			name:         "Hash reference",
			code:         "my $hash = {};",
			expectedType: "HashRef",
		},
		{
			name:         "DBI connection",
			code:         "my $dbh = DBI->connect($dsn, $user, $pass);",
			expectedType: "DBI::db",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Find matching type inference rule
			for _, rule := range introspector.DataStructureAnalyzer.TypeInferenceRules {
				if rule.Pattern.MatchString(tc.code) {
					matches := rule.Pattern.FindStringSubmatch(tc.code)
					inferredType := rule.InferType(matches)

					if inferredType != tc.expectedType {
						t.Errorf("Expected type '%s', got '%s'", tc.expectedType, inferredType)
					}
					return
				}
			}
			t.Errorf("No type inference rule matched for code: %s", tc.code)
		})
	}
}

func TestUtilityFunctions(t *testing.T) {
	// Test contains function
	slice := []string{"a", "b", "c"}
	if !contains(slice, "b") {
		t.Error("contains() should return true for existing item")
	}
	if contains(slice, "d") {
		t.Error("contains() should return false for non-existing item")
	}

	// Test extractVariableNameFromText
	introspector, err := NewModuleIntrospector()
	if err != nil {
		t.Fatalf("Failed to create introspector: %v", err)
	}

	testCases := []struct {
		code        string
		expectedVar string
	}{
		{"my $name = 'test';", "$name"},
		{"our $version = '1.0';", "$version"},
		{"$result = process();", "$result"},
	}

	for _, tc := range testCases {
		varName := introspector.extractVariableNameFromText(tc.code)
		if varName != tc.expectedVar {
			t.Errorf("Expected variable '%s', got '%s' for code: %s", tc.expectedVar, varName, tc.code)
		}
	}
}

func TestMooseAttributeExtraction(t *testing.T) {
	introspector, err := NewModuleIntrospector()
	if err != nil {
		t.Fatalf("Failed to create introspector: %v", err)
	}

	mooseAttrCode := `has 'name' => (
    is       => 'rw',
    isa      => 'Str',
    required => 1,
    default  => 'Unknown'
);`

	mockNode := &MockNode{text: mooseAttrCode}
	result := &ModuleIntrospectionResult{
		Packages: make(map[string]*PackageInfo),
	}

	// Find the Moose framework pattern
	pattern := introspector.OOPPatternDetector.FrameworkPatterns["Moose"]
	if pattern == nil {
		t.Fatal("Moose framework pattern not found")
	}

	introspector.extractFrameworkAttribute(mockNode, pattern, result)

	// Check if attribute was extracted
	if pkg, exists := result.Packages["main"]; exists {
		if attr, hasAttr := pkg.Attributes["name"]; hasAttr {
			if attr.Type != "Str" {
				t.Errorf("Expected attribute type 'Str', got '%s'", attr.Type)
			}
			if !attr.IsRequired {
				t.Error("Expected attribute to be required")
			}
			if !attr.HasDefault {
				t.Error("Expected attribute to have default value")
			}
		} else {
			t.Error("Failed to extract 'name' attribute")
		}
	} else {
		t.Error("No package found in result")
	}
}
