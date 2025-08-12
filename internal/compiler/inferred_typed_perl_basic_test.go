// ABOUTME: Basic functionality tests for the inferred typed Perl compiler implementation
// ABOUTME: Tests core type annotation generation without complex AST construction

package compiler

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/types"
)

func TestBasicInferredTypedPerlCompiler(t *testing.T) {
	t.Run("Compiler creation and options", func(t *testing.T) {
		compiler := NewInferredTypedPerlCompiler()
		if compiler == nil {
			t.Fatal("NewInferredTypedPerlCompiler() returned nil")
		}

		// Test with options
		options := InferredCompilerOptions{
			AnnotationStyle:   StyleInline,
			MinimumConfidence: 0.8,
			PreserveComments:  true,
		}
		compilerWithOptions := NewInferredTypedPerlCompilerWithOptions(options)
		if compilerWithOptions == nil {
			t.Fatal("NewInferredTypedPerlCompilerWithOptions() returned nil")
		}

		if compilerWithOptions.options.MinimumConfidence != 0.8 {
			t.Errorf("Expected MinimumConfidence 0.8, got %f", compilerWithOptions.options.MinimumConfidence)
		}
	})

	t.Run("Target identification", func(t *testing.T) {
		compiler := NewInferredTypedPerlCompiler()
		if compiler.Target() != TargetInferredTypeAnnotations {
			t.Errorf("Expected target %s, got %s", TargetInferredTypeAnnotations, compiler.Target())
		}
	})
}

func TestTypeFormatting(t *testing.T) {
	options := InferredCompilerOptions{
		AnnotationStyle:   StyleInline,
		MinimumConfidence: 0.7,
	}

	nc := &nodeCompiler{
		options:     options,
		typeInfoMap: make(map[string]*types.TypeInfo),
	}

	tests := []struct {
		name            string
		typeConstructor func() types.Type
		expected        string
	}{
		{"Int type", types.NewIntType, "Int"},
		{"Str type", types.NewStrType, "Str"},
		{"Bool type", types.NewBoolType, "Bool"},
		{"Num type", types.NewNumType, "Num"},
		{"Any type", types.NewAnyType, "Any"},
		{"ArrayRef type", types.NewArrayRefAnyType, "ArrayRef"},
		{"HashRef type", types.NewHashRefAnyType, "HashRef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeInstance := tt.typeConstructor()
			result := nc.formatType(typeInstance)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTypeInfoCreation(t *testing.T) {
	t.Run("Basic TypeInfo creation", func(t *testing.T) {
		typeInfo := types.NewTypeInfo(types.NewIntType(), 0.9, types.SourceLiteral)
		if typeInfo == nil {
			t.Fatal("NewTypeInfo returned nil")
		}

		if typeInfo.Confidence != 0.9 {
			t.Errorf("Expected confidence 0.9, got %f", typeInfo.Confidence)
		}

		if typeInfo.Source != types.SourceLiteral {
			t.Errorf("Expected source %s, got %s", types.SourceLiteral, typeInfo.Source)
		}
	})

	t.Run("TypeInfo with location", func(t *testing.T) {
		typeInfo := types.NewTypeInfo(types.NewStrType(), 0.8, types.SourceVariable)
		typeInfo.WithLocation("test.pl", 1, 10)

		if typeInfo.Location == nil {
			t.Fatal("Location was not set")
		}

		if typeInfo.Location.File != "test.pl" {
			t.Errorf("Expected file 'test.pl', got %s", typeInfo.Location.File)
		}
	})
}

func TestInferredCompilerRegistryIntegration(t *testing.T) {
	registry := NewCompilerRegistry()

	t.Run("Inferred compiler is registered", func(t *testing.T) {
		compiler, exists := registry.GetCompiler(TargetInferredTypeAnnotations)
		if !exists {
			t.Error("TargetInferredTypeAnnotations not registered")
		}
		if compiler == nil {
			t.Error("Compiler is nil")
		}
	})

	t.Run("Target appears in available targets", func(t *testing.T) {
		targets := registry.AvailableTargets()
		found := false
		for _, target := range targets {
			if target == TargetInferredTypeAnnotations {
				found = true
				break
			}
		}
		if !found {
			t.Error("TargetInferredTypeAnnotations not found in available targets")
		}
	})
}

func TestAddPerlVersionPragma(t *testing.T) {
	compiler := NewInferredTypedPerlCompiler()

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "Simple code without shebang",
			input:    "print \"hello\";",
			contains: "use v",
		},
		{
			name:     "Code with shebang",
			input:    "#!/usr/bin/perl\nprint \"hello\";",
			contains: "use v",
		},
		{
			name:     "Empty input",
			input:    "",
			contains: "use v",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compiler.addPerlVersionPragma(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain '%s', got: %s", tt.contains, result)
			}

			// Should preserve original content
			if tt.input != "" {
				// For shebang case, check that both shebang and content are preserved
				if strings.HasPrefix(tt.input, "#!") {
					lines := strings.Split(tt.input, "\n")
					for _, line := range lines {
						if strings.TrimSpace(line) != "" && !strings.Contains(result, strings.TrimSpace(line)) {
							t.Errorf("Line '%s' not preserved in result: %s", line, result)
						}
					}
				} else if !strings.Contains(result, strings.TrimSpace(tt.input)) {
					t.Errorf("Original content not preserved in result: %s", result)
				}
			}
		})
	}
}
