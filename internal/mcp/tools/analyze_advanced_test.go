// ABOUTME: Tests for advanced analysis features in MCP code analyzer
// ABOUTME: Verifies flow-sensitive analysis, code quality metrics, and type compatibility

package tools

import (
	"context"
	"testing"

	"tamarou.com/pvm/internal/mcp/validation"
)

func TestCodeAnalyzer_PerformFlowAnalysis(t *testing.T) {
	// Create analyzer
	cache, err := validation.NewValidationCache("10MB")
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	validator, err := validation.NewValidator(cache)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	analyzer, err := NewCodeAnalyzer(validator, nil)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	testCases := []struct {
		name     string
		code     string
		expected struct {
			patterns       []string
			hasRefinements bool
		}
	}{
		{
			name: "defined check refinement",
			code: `#!/usr/bin/env perl
use strict;
use warnings;

my $value;
if (defined $value) {
    # $value is now !Undef
    print $value;
}`,
			expected: struct {
				patterns       []string
				hasRefinements bool
			}{
				patterns:       []string{"defined_check"},
				hasRefinements: true,
			},
		},
		{
			name: "ref check pattern",
			code: `#!/usr/bin/env perl
use strict;
use warnings;

my $obj = {};
if (ref $obj eq 'HASH') {
    # $obj is refined to HashRef
    $obj->{key} = 'value';
}`,
			expected: struct {
				patterns       []string
				hasRefinements bool
			}{
				patterns:       []string{"ref_check"},
				hasRefinements: true,
			},
		},
		{
			name: "isa check pattern",
			code: `#!/usr/bin/env perl
use strict;
use warnings;

my $obj = MyClass->new();
if ($obj->isa('MyClass')) {
    # $obj is confirmed as MyClass
    $obj->method();
}`,
			expected: struct {
				patterns       []string
				hasRefinements bool
			}{
				patterns:       []string{"isa_check"},
				hasRefinements: true,
			},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := analyzer.performFlowAnalysis(ctx, tc.code, "/tmp")
			if err != nil {
				t.Fatalf("Flow analysis failed: %v", err)
			}

			if result.FlowAnalysis == nil {
				t.Fatal("Expected flow analysis result")
			}

			// Check validation patterns
			foundPatterns := result.FlowAnalysis.ValidationPatterns
			for _, expectedPattern := range tc.expected.patterns {
				found := false
				for _, pattern := range foundPatterns {
					if pattern == expectedPattern {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected pattern %s not found", expectedPattern)
				}
			}
		})
	}
}

func TestCodeAnalyzer_AnalyzeCodeQuality(t *testing.T) {
	// Create analyzer
	cache, err := validation.NewValidationCache("10MB")
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	validator, err := validation.NewValidator(cache)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	analyzer, err := NewCodeAnalyzer(validator, nil)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	testCases := []struct {
		name     string
		code     string
		expected struct {
			minComplexity int
			hasMetrics    bool
		}
	}{
		{
			name: "simple function",
			code: `#!/usr/bin/env perl
use strict;
use warnings;

sub simple_function {
    my $x = shift;
    return $x + 1;
}`,
			expected: struct {
				minComplexity int
				hasMetrics    bool
			}{
				minComplexity: 1,
				hasMetrics:    true,
			},
		},
		{
			name: "complex function with branches",
			code: `#!/usr/bin/env perl
use strict;
use warnings;

sub complex_function {
    my $x = shift;
    if ($x > 0) {
        if ($x < 10) {
            return "small";
        } elsif ($x < 100) {
            return "medium";
        } else {
            return "large";
        }
    } else {
        return "negative" if $x < 0;
        return "zero";
    }
}`,
			expected: struct {
				minComplexity int
				hasMetrics    bool
			}{
				minComplexity: 5, // Multiple if/elsif branches
				hasMetrics:    true,
			},
		},
		{
			name: "function with loops",
			code: `#!/usr/bin/env perl
use strict;
use warnings;

sub loop_function {
    my @items = @_;
    my $sum = 0;

    for my $item (@items) {
        if ($item > 0) {
            $sum += $item;
        }
    }

    while ($sum > 100) {
        $sum = $sum / 2;
    }

    return $sum;
}`,
			expected: struct {
				minComplexity int
				hasMetrics    bool
			}{
				minComplexity: 3, // for loop + if + while
				hasMetrics:    true,
			},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := analyzer.analyzeCodeQuality(ctx, tc.code, "/tmp")
			if err != nil {
				t.Fatalf("Code quality analysis failed: %v", err)
			}

			if result.CodeQuality == nil {
				t.Fatal("Expected code quality metrics")
			}

			metrics := result.CodeQuality
			if metrics.CyclomaticComplexity < tc.expected.minComplexity {
				t.Errorf("Expected complexity >= %d, got %d",
					tc.expected.minComplexity, metrics.CyclomaticComplexity)
			}

			// Check that type safety metrics exist
			if metrics.TypeSafety.SafetyScore == 0 {
				t.Error("Expected non-zero safety score")
			}
		})
	}
}

func TestCodeAnalyzer_CheckTypeCompatibility(t *testing.T) {
	// Create analyzer
	cache, err := validation.NewValidationCache("10MB")
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	validator, err := validation.NewValidator(cache)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	analyzer, err := NewCodeAnalyzer(validator, nil)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	// Test basic type compatibility
	testCases := []struct {
		type1      string
		type2      string
		compatible bool
		reason     string
	}{
		{
			type1:      "Int",
			type2:      "Int",
			compatible: true,
			reason:     "Types are identical",
		},
		{
			type1:      "Int",
			type2:      "Num",
			compatible: true,
			reason:     "Int is a subtype of Num",
		},
		{
			type1:      "Str",
			type2:      "Any",
			compatible: true,
			reason:     "Str is a subtype of Any",
		},
		{
			type1:      "Str",
			type2:      "Int",
			compatible: false,
			reason:     "Str and Int are incompatible types",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.type1+"_vs_"+tc.type2, func(t *testing.T) {
			compatible := analyzer.areTypesCompatible(tc.type1, tc.type2)
			if compatible != tc.compatible {
				t.Errorf("Expected compatibility %v, got %v", tc.compatible, compatible)
			}

			reason := analyzer.getCompatibilityReason(tc.type1, tc.type2)
			if reason != tc.reason {
				t.Errorf("Expected reason '%s', got '%s'", tc.reason, reason)
			}
		})
	}
}

func TestCodeAnalyzer_AnalyzeTypeAnnotations(t *testing.T) {
	// Create analyzer
	cache, err := validation.NewValidationCache("10MB")
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	validator, err := validation.NewValidator(cache)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	analyzer, err := NewCodeAnalyzer(validator, nil)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	testCases := []struct {
		name     string
		code     string
		expected struct {
			totalAnnotations int
			hasSuggestions   bool
		}
	}{
		{
			name: "fully annotated code",
			code: `#!/usr/bin/env perl
use strict;
use warnings;

my Int $count = 0;
my Str $name = "test";

sub add_numbers {
    my Int $a = shift;
    my Int $b = shift;
    return $a + $b;
}`,
			expected: struct {
				totalAnnotations int
				hasSuggestions   bool
			}{
				totalAnnotations: 4, // $count, $name, $a, $b
				hasSuggestions:   false,
			},
		},
		{
			name: "partially annotated code",
			code: `#!/usr/bin/env perl
use strict;
use warnings;

my Int $count = 0;
my $name = "test";  # Missing annotation

sub process {
    my $input = shift;  # Missing annotation
    return length($input);
}`,
			expected: struct {
				totalAnnotations int
				hasSuggestions   bool
			}{
				totalAnnotations: 1, // Only $count
				hasSuggestions:   true,
			},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := analyzer.analyzeTypeAnnotations(ctx, tc.code, "/tmp")
			if err != nil {
				t.Fatalf("Type annotation analysis failed: %v", err)
			}

			if result.TypeAnnotationAnalysis == nil {
				t.Fatal("Expected type annotation analysis")
			}

			analysis := result.TypeAnnotationAnalysis
			if analysis.TotalAnnotations != tc.expected.totalAnnotations {
				t.Errorf("Expected %d annotations, got %d",
					tc.expected.totalAnnotations, analysis.TotalAnnotations)
			}
		})
	}
}

func TestCodeAnalyzer_PerformFullAnalysis(t *testing.T) {
	// Create analyzer
	cache, err := validation.NewValidationCache("10MB")
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	validator, err := validation.NewValidator(cache)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	analyzer, err := NewCodeAnalyzer(validator, nil)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	code := `#!/usr/bin/env perl
use strict;
use warnings;

type UserID = Int;

my UserID $id = 123;
my Str $name = "Alice";

sub get_user {
    my UserID $user_id = shift;

    if (defined $user_id && $user_id > 0) {
        return {
            id => $user_id,
            name => $name
        };
    }

    return undef;
}

my $user = get_user($id);
if (ref $user eq 'HASH') {
    print "User: $user->{name}\n";
}`

	ctx := context.Background()
	result, err := analyzer.performFullAnalysis(ctx, code, "/tmp", false)
	if err != nil {
		t.Fatalf("Full analysis failed: %v", err)
	}

	// Verify all analysis components are present
	if result.AnalysisType != "full_analysis" {
		t.Errorf("Expected analysis type 'full_analysis', got '%s'", result.AnalysisType)
	}

	if result.FlowAnalysis == nil {
		t.Error("Expected flow analysis in full analysis")
	}

	if result.CodeQuality == nil {
		t.Error("Expected code quality metrics in full analysis")
	}

	if result.TypeCompatibility == nil {
		t.Error("Expected type compatibility in full analysis")
	}

	if result.TypeAnnotationAnalysis == nil {
		t.Error("Expected type annotation analysis in full analysis")
	}

	// Check specific results
	if result.FlowAnalysis != nil {
		patterns := result.FlowAnalysis.ValidationPatterns
		expectedPatterns := []string{"defined_check", "ref_check"}
		for _, expected := range expectedPatterns {
			found := false
			for _, pattern := range patterns {
				if pattern == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected pattern '%s' not found", expected)
			}
		}
	}

	if result.CodeQuality != nil && result.CodeQuality.CyclomaticComplexity < 3 {
		t.Error("Expected higher complexity for code with conditionals")
	}
}
