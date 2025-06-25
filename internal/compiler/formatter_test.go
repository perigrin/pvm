// ABOUTME: Comprehensive tests for type formatting system and annotation generation
// ABOUTME: Validates formatter behavior across different styles and confidence levels

package compiler

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/types"
)

func TestTypeFormatter(t *testing.T) {
	formatter := NewTypeFormatter()
	
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{"Basic formatter creation", func(t *testing.T) {
			if formatter == nil {
				t.Error("NewTypeFormatter() returned nil")
			}
		}},
		{"Formatter interface compliance", func(t *testing.T) {
			var _ TypeFormatter = formatter
		}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

func TestBasicTypeAnnotationGeneration(t *testing.T) {
	formatter := NewTypeFormatter()
	
	tests := []struct {
		name           string
		varName        string
		varType        types.Type
		confidence     float64
		expectedStyle  FormattingStyle
		shouldContain  []string
		shouldNotContain []string
	}{
		{
			name:          "High confidence Int annotation",
			varName:       "$count",
			varType:       types.NewIntType(),
			confidence:    0.95,
			expectedStyle: StyleInline,
			shouldContain: []string{"my", "Int", "$count"},
		},
		{
			name:          "High confidence Str annotation",
			varName:       "$name",
			varType:       types.NewStrType(),
			confidence:    0.90,
			expectedStyle: StyleInline,
			shouldContain: []string{"my", "Str", "$name"},
		},
		{
			name:          "Medium confidence with ArrayRef",
			varName:       "$items",
			varType:       types.NewArrayRefType(types.NewIntType()),
			confidence:    0.75,
			expectedStyle: StyleInline,
			shouldContain: []string{"my", "ArrayRef[Int]", "$items"},
		},
		{
			name:              "Low confidence should use comments",
			varName:           "$uncertain",
			varType:           types.NewStrType(),
			confidence:        0.45,
			expectedStyle:     StyleCommentOnly,
			shouldContain:     []string{"my", "$uncertain", "#"},
			shouldNotContain:  []string{"Str $uncertain"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatVariableDeclaration(tt.varName, tt.varType, tt.confidence, tt.expectedStyle)
			
			for _, expected := range tt.shouldContain {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', got: %s", expected, result)
				}
			}
			
			for _, notExpected := range tt.shouldNotContain {
				if strings.Contains(result, notExpected) {
					t.Errorf("Expected result NOT to contain '%s', got: %s", notExpected, result)
				}
			}
		})
	}
}

func TestConfidenceBasedAnnotationDecisions(t *testing.T) {
	// Test with different confidence thresholds
	tests := []struct {
		name              string
		confidence        float64
		threshold         float64
		shouldInclude     bool
	}{
		{"High confidence above threshold", 0.9, 0.7, true},
		{"Medium confidence above threshold", 0.75, 0.7, true},
		{"Low confidence below threshold", 0.6, 0.7, false},
		{"Very low confidence", 0.3, 0.7, false},
		{"Exact threshold", 0.7, 0.7, true},
	}
	
	formatter := NewTypeFormatter()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.ShouldIncludeAnnotation(tt.confidence, tt.threshold)
			if result != tt.shouldInclude {
				t.Errorf("Expected ShouldIncludeAnnotation(%f, %f) = %v, got %v", 
					tt.confidence, tt.threshold, tt.shouldInclude, result)
			}
		})
	}
}

func TestMultipleFormattingStyles(t *testing.T) {
	formatter := NewTypeFormatter()
	typeInfo := &types.TypeInfo{
		Type:       types.NewIntType(),
		Confidence: 0.85,
		Source:     types.SourceLiteral,
	}
	
	tests := []struct {
		name                string
		style               FormattingStyle
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name:             "Inline style",
			style:            StyleInline,
			expectedContains: []string{"Int"},
		},
		{
			name:             "Verbose style",
			style:            StyleVerbose,
			expectedContains: []string{"Int"},
		},
		{
			name:                "Compact style",
			style:               StyleCompact,
			expectedContains:    []string{"Int"},
		},
		{
			name:             "Comment only style",
			style:            StyleCommentOnly,
			expectedContains: []string{"#", "type:", "Int"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatTypeAnnotation(typeInfo, tt.style)
			
			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Style %s: expected result to contain '%s', got: %s", tt.style, expected, result)
				}
			}
			
			for _, notExpected := range tt.expectedNotContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("Style %s: expected result NOT to contain '%s', got: %s", tt.style, notExpected, result)
				}
			}
		})
	}
}

func TestConfidenceCommentFormatting(t *testing.T) {
	formatter := NewTypeFormatter()
	
	tests := []struct {
		name              string
		confidence        float64
		expectedContains  []string
	}{
		{
			name:             "High confidence comment",
			confidence:       0.95,
			expectedContains: []string{"#", "high confidence", "95%"},
		},
		{
			name:             "Medium confidence comment",
			confidence:       0.75,
			expectedContains: []string{"#", "medium confidence", "75%"},
		},
		{
			name:             "Low confidence comment",
			confidence:       0.55,
			expectedContains: []string{"#", "low confidence", "55%"},
		},
		{
			name:             "Very low confidence comment",
			confidence:       0.35,
			expectedContains: []string{"#", "uncertain", "35%"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatConfidenceComment(tt.confidence)
			
			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected confidence comment to contain '%s', got: %s", expected, result)
				}
			}
		})
	}
}

func TestComplexTypeFormatting(t *testing.T) {
	formatter := NewTypeFormatter()
	
	tests := []struct {
		name         string
		varType      types.Type
		expectedType string
	}{
		{
			name:         "Nested ArrayRef",
			varType:      types.NewArrayRefType(types.NewArrayRefType(types.NewIntType())),
			expectedType: "ArrayRef[ArrayRef[Int]]",
		},
		{
			name:         "HashRef with complex value",
			varType:      types.NewHashRefType(types.NewArrayRefType(types.NewStrType())),
			expectedType: "HashRef[ArrayRef[Str]]",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatVariableDeclaration("$var", tt.varType, 0.9, StyleInline)
			
			if !strings.Contains(result, tt.expectedType) {
				t.Errorf("Expected complex type formatting to contain '%s', got: %s", tt.expectedType, result)
			}
		})
	}
}

func TestFormatterOptions(t *testing.T) {
	tests := []struct {
		name    string
		options FormatterOptions
		testFunc func(t *testing.T, formatter TypeFormatter)
	}{
		{
			name: "High confidence threshold",
			options: FormatterOptions{
				ConfidenceThreshold: 0.9,
				PreferComments:      false,
			},
			testFunc: func(t *testing.T, formatter TypeFormatter) {
				result := formatter.FormatVariableDeclaration("$var", types.NewIntType(), 0.8, StyleInline)
				// Should not include type annotation due to high threshold
				if strings.Contains(result, "Int $var") {
					t.Errorf("Expected no type annotation with high threshold, got: %s", result)
				}
			},
		},
		{
			name: "Prefer comments option",
			options: FormatterOptions{
				ConfidenceThreshold: 0.7,
				PreferComments:      true,
			},
			testFunc: func(t *testing.T, formatter TypeFormatter) {
				result := formatter.FormatVariableDeclaration("$var", types.NewIntType(), 0.6, StyleInline)
				// Should include comment due to prefer comments option
				if !strings.Contains(result, "#") {
					t.Errorf("Expected comment with PreferComments=true, got: %s", result)
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewTypeFormatterWithOptions(tt.options)
			tt.testFunc(t, formatter)
		})
	}
}

func TestAnnotationGenerator(t *testing.T) {
	formatter := NewTypeFormatter()
	generator := NewAnnotationGenerator(formatter)
	
	t.Run("Generator creation", func(t *testing.T) {
		if generator == nil {
			t.Error("NewAnnotationGenerator() returned nil")
		}
	})
	
	// Note: More comprehensive annotation generator tests would require
	// actual AST nodes, which would be integration tests
}