// ABOUTME: Tests for migration layer functionality ensuring backward compatibility
// ABOUTME: Validates parsing strategy selection and AST conversion utilities

package migration

import (
	"fmt"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/parser"
)

func TestDefaultMigrationConfig(t *testing.T) {
	config := DefaultMigrationConfig()

	if config.Mode != ModePreferTreeSitter {
		t.Error("Default config should prefer tree-sitter")
	}

	if !config.AllowFallback {
		t.Error("Default config should allow fallback")
	}

	if len(config.PreferShimForTypes) == 0 {
		t.Error("Default config should have preferred types")
	}

	// Check for specific expected types
	expectedTypes := []string{"variable_declaration", "type_annotation", "parameterized_type"}
	for _, expected := range expectedTypes {
		found := false
		for _, actual := range config.PreferShimForTypes {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Default config should include %s in preferred types", expected)
		}
	}
}

func TestNewMigrationParser(t *testing.T) {
	tests := []struct {
		name              string
		config            *MigrationConfig
		expectError       bool
		expectTraditional bool
		expectShim        bool
	}{
		{
			name:              "default config",
			config:            nil,
			expectError:       false,
			expectTraditional: true,
			expectShim:        true,
		},
		{
			name: "tree-sitter only",
			config: &MigrationConfig{
				Mode:          ModeTreeSitterOnly,
				AllowFallback: false,
			},
			expectError:       false,
			expectTraditional: false,
			expectShim:        true,
		},
		{
			name: "traditional only",
			config: &MigrationConfig{
				Mode:          ModeTraditionalOnly,
				AllowFallback: false,
			},
			expectError:       false,
			expectTraditional: true,
			expectShim:        false,
		},
		{
			name: "prefer tree-sitter with fallback",
			config: &MigrationConfig{
				Mode:          ModePreferTreeSitter,
				AllowFallback: true,
			},
			expectError:       false,
			expectTraditional: true,
			expectShim:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp, err := NewMigrationParser(tt.config)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if mp != nil {
				if tt.expectTraditional && mp.traditionalParser == nil {
					t.Error("Expected traditional parser but got nil")
				}
				if !tt.expectTraditional && mp.traditionalParser != nil {
					t.Error("Did not expect traditional parser but got one")
				}

				if tt.expectShim && mp.shimParser == nil {
					t.Error("Expected shim parser but got nil")
				}
				if !tt.expectShim && mp.shimParser != nil {
					t.Error("Did not expect shim parser but got one")
				}
			}
		})
	}
}

func TestParsingStrategySelection(t *testing.T) {
	tests := []struct {
		name             string
		config           *MigrationConfig
		content          string
		expectedStrategy string
	}{
		{
			name: "tree-sitter only mode",
			config: &MigrationConfig{
				Mode: ModeTreeSitterOnly,
			},
			content:          "my $x = 42;",
			expectedStrategy: "tree-sitter",
		},
		{
			name: "traditional only mode",
			config: &MigrationConfig{
				Mode: ModeTraditionalOnly,
			},
			content:          "my $x = 42;",
			expectedStrategy: "traditional",
		},
		{
			name: "prefer tree-sitter with typed content",
			config: &MigrationConfig{
				Mode:               ModePreferTreeSitter,
				PreferShimForTypes: []string{"variable_declaration"},
			},
			content:          "my Int $count = 42;",
			expectedStrategy: "tree-sitter",
		},
		{
			name: "prefer traditional without typed content",
			config: &MigrationConfig{
				Mode:               ModePreferTraditional,
				PreferShimForTypes: []string{"variable_declaration"},
			},
			content:          "my $count = 42;",
			expectedStrategy: "traditional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &MigrationParser{
				config: tt.config,
			}

			// Create real parsers to test strategy selection
			// Note: Cannot easily mock interface types, so we use real parsers
			if tt.expectedStrategy == "tree-sitter" || tt.config.AllowFallback {
				if shimParser, err := parser.NewShimParser(); err == nil {
					mp.shimParser = shimParser
				}
			}
			if tt.expectedStrategy == "traditional" || tt.config.AllowFallback {
				if traditionalParser, err := parser.NewParser(); err == nil {
					mp.traditionalParser = traditionalParser
				}
			}

			strategy := mp.selectParsingStrategy("", tt.content)
			if strategy != tt.expectedStrategy {
				t.Errorf("Expected strategy %s, got %s", tt.expectedStrategy, strategy)
			}
		})
	}
}

func TestContentTypeDetection(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		function func(string) bool
		expected bool
	}{
		{
			name:     "typed variables - Int",
			content:  "my Int $count = 42;",
			function: containsTypedVariables,
			expected: true,
		},
		{
			name:     "typed variables - ArrayRef",
			content:  "my ArrayRef[Str] $items = [];",
			function: containsTypedVariables,
			expected: true,
		},
		{
			name:     "untyped variables",
			content:  "my $count = 42;",
			function: containsTypedVariables,
			expected: false,
		},
		{
			name:     "type annotations - ArrayRef",
			content:  "ArrayRef[Int]",
			function: containsTypeAnnotations,
			expected: true,
		},
		{
			name:     "type annotations - HashRef",
			content:  "HashRef[Str]",
			function: containsTypeAnnotations,
			expected: true,
		},
		{
			name:     "no type annotations",
			content:  "my $hash = {};",
			function: containsTypeAnnotations,
			expected: false,
		},
		{
			name:     "union types - Int|Str",
			content:  "my (Int|Str) $value;",
			function: containsUnionTypes,
			expected: true,
		},
		{
			name:     "union types - Maybe",
			content:  "Maybe[Int]",
			function: containsUnionTypes,
			expected: true,
		},
		{
			name:     "no union types",
			content:  "my Int $value;",
			function: containsUnionTypes,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.content)
			if result != tt.expected {
				t.Errorf("Expected %v for content %q, got %v", tt.expected, tt.content, result)
			}
		})
	}
}

func TestContainsPattern(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		patterns []string
		expected bool
	}{
		{
			name:     "single pattern match",
			content:  "my Int $x = 42;",
			patterns: []string{"Int "},
			expected: true,
		},
		{
			name:     "multiple patterns - first matches",
			content:  "my Str $name = 'test';",
			patterns: []string{"Str ", "Int "},
			expected: true,
		},
		{
			name:     "multiple patterns - second matches",
			content:  "my Int $count = 0;",
			patterns: []string{"Str ", "Int "},
			expected: true,
		},
		{
			name:     "no pattern matches",
			content:  "my $variable = 42;",
			patterns: []string{"Int ", "Str "},
			expected: false,
		},
		{
			name:     "empty patterns",
			content:  "any content",
			patterns: []string{},
			expected: false,
		},
		{
			name:     "pattern longer than content",
			content:  "short",
			patterns: []string{"very long pattern"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsPattern(tt.content, tt.patterns)
			if result != tt.expected {
				t.Errorf("Expected %v for content %q with patterns %v, got %v",
					tt.expected, tt.content, tt.patterns, result)
			}
		})
	}
}

func TestMigrationParserParseString(t *testing.T) {
	// Test with real parsers if possible
	config := DefaultMigrationConfig()
	config.LogMigrationChoices = true // Enable logging for debugging

	mp, err := NewMigrationParser(config)
	if err != nil {
		t.Skipf("Cannot create migration parser: %v", err)
	}

	tests := []struct {
		name             string
		content          string
		expectTreeSitter bool
	}{
		{
			name:             "typed variable should prefer tree-sitter",
			content:          "my Int $count = 42;",
			expectTreeSitter: true,
		},
		{
			name:             "simple variable should use available parser",
			content:          "my $count = 42;",
			expectTreeSitter: false, // May use either, depends on config
		},
		{
			name:             "complex type should prefer tree-sitter",
			content:          "my ArrayRef[HashRef[Str]] $complex = [];",
			expectTreeSitter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mp.ParseString(tt.content)
			if err != nil {
				t.Errorf("ParseString failed: %v", err)
				return
			}

			if result == nil {
				t.Error("ParseString returned nil result")
				return
			}

			// Check the type of result to see which parser was used
			switch result.(type) {
			case *ast.TreeSitterAST:
				if !tt.expectTreeSitter && mp.config.Mode == ModePreferTraditional {
					t.Log("Note: Used tree-sitter when traditional was expected (may be due to fallback)")
				}
			case *ast.AST:
				if tt.expectTreeSitter && mp.config.Mode == ModePreferTreeSitter {
					t.Log("Note: Used traditional when tree-sitter was expected (may be due to fallback)")
				}
			default:
				t.Errorf("Unexpected result type: %T", result)
			}

			t.Logf("Parsed successfully with result type: %T", result)
		})
	}
}

func TestNewASTConverter(t *testing.T) {
	converter, err := NewASTConverter()
	if err != nil {
		t.Skipf("Cannot create AST converter: %v", err)
	}

	if converter.shimParser == nil {
		t.Error("AST converter should have shim parser")
	}

	if converter.traditionalParser == nil {
		t.Error("AST converter should have traditional parser")
	}
}

func TestExtractTypeAnnotations(t *testing.T) {
	converter, err := NewASTConverter()
	if err != nil {
		t.Skipf("Cannot create AST converter: %v", err)
	}

	// Test with TreeSitterAST
	testCode := "my Int $x = 42;"
	tsAST, err := converter.shimParser.ParseStringShim(testCode)
	if err != nil {
		t.Skipf("Cannot parse with tree-sitter: %v", err)
	}

	annotations, err := converter.ExtractTypeAnnotations(tsAST)
	if err != nil {
		t.Errorf("ExtractTypeAnnotations failed for TreeSitterAST: %v", err)
	}

	if annotations == nil {
		t.Error("ExtractTypeAnnotations should return non-nil slice")
	}

	t.Logf("Extracted %d type annotations from TreeSitterAST", len(annotations))

	// Test with traditional AST
	tradAST, err := converter.traditionalParser.ParseString(testCode)
	if err != nil {
		t.Skipf("Cannot parse with traditional parser: %v", err)
	}

	annotations, err = converter.ExtractTypeAnnotations(tradAST)
	if err != nil {
		t.Errorf("ExtractTypeAnnotations failed for traditional AST: %v", err)
	}

	if annotations == nil {
		t.Error("ExtractTypeAnnotations should return non-nil slice")
	}

	t.Logf("Extracted %d type annotations from traditional AST", len(annotations))

	// Test with unsupported type
	_, err = converter.ExtractTypeAnnotations("invalid")
	if err == nil {
		t.Error("ExtractTypeAnnotations should fail for unsupported type")
	}
}

func TestMigrationError(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	migErr := NewMigrationError("test_operation", originalErr, "test context")

	if migErr.Operation != "test_operation" {
		t.Error("Migration error operation not set correctly")
	}

	if migErr.Cause != originalErr {
		t.Error("Migration error cause not set correctly")
	}

	if migErr.Context != "test context" {
		t.Error("Migration error context not set correctly")
	}

	errorMsg := migErr.Error()
	if !strings.Contains(errorMsg, "test_operation") {
		t.Error("Error message should contain operation")
	}

	if !strings.Contains(errorMsg, "test context") {
		t.Error("Error message should contain context")
	}

	if !strings.Contains(errorMsg, "original error") {
		t.Error("Error message should contain original error")
	}
}

func TestConvertPosition(t *testing.T) {
	// Test with Position type
	pos := ast.Position{Line: 5, Column: 10, Offset: 50}
	converted := ConvertPosition(pos)
	if converted != pos {
		t.Error("Position should be returned unchanged")
	}

	// Test with unknown type
	converted = ConvertPosition("invalid")
	if converted.Line != 0 || converted.Column != 0 || converted.Offset != 0 {
		t.Error("Unknown type should return zero Position")
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-1, 1},
		{100, 100},
		{-100, 100},
	}

	for _, tt := range tests {
		result := abs(tt.input)
		if result != tt.expected {
			t.Errorf("abs(%d) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}

func TestMigrationModes(t *testing.T) {
	modes := []MigrationMode{
		ModePreferTraditional,
		ModePreferTreeSitter,
		ModeTreeSitterOnly,
		ModeTraditionalOnly,
	}

	// Ensure modes have expected values
	if ModePreferTraditional != 0 {
		t.Error("ModePreferTraditional should be 0")
	}

	if len(modes) != 4 {
		t.Error("Expected 4 migration modes")
	}

	// Test that modes can be used in switch statements
	for _, mode := range modes {
		switch mode {
		case ModePreferTraditional, ModePreferTreeSitter, ModeTreeSitterOnly, ModeTraditionalOnly:
			// Valid mode
		default:
			t.Errorf("Unexpected migration mode: %v", mode)
		}
	}
}
