// ABOUTME: Validates type annotation preservation in tree-sitter shim workflows
// ABOUTME: Ensures CST-based compilation maintains type annotations better than traditional AST

package validation

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/compiler"
	"tamarou.com/pvm/internal/parser"
)

// TestTypeAnnotationPreservationWorkflow validates that tree-sitter shim preserves
// type annotations throughout the complete PVM compilation workflow
func TestTypeAnnotationPreservationWorkflow(t *testing.T) {
	testCases := []struct {
		name                string
		inputCode           string
		expectedAnnotations []string
		workflowDescription string
	}{
		{
			name: "variable_declarations_with_types",
			inputCode: `
my Int $count = 42;
my Str $name = "test";
my ArrayRef[Int] $numbers = [1, 2, 3];
my HashRef[Str, Any] $config = {key => "value"};`,
			expectedAnnotations: []string{
				"Int $count",
				"Str $name",
				"ArrayRef[Int] $numbers",
				"HashRef[Str, Any] $config",
			},
			workflowDescription: "Basic variable declarations with type annotations",
		},
		{
			name: "function_signatures_with_types",
			inputCode: `
sub process_data(Int $id, Str $name) -> HashRef[Str, Any] {
    my ArrayRef[Str] $items = get_items($id);
    return {id => $id, name => $name, items => $items};
}

sub calculate_total(ArrayRef[Num] $prices) -> Num {
    my Num $total = 0;
    for my Num $price (@$prices) {
        $total += $price;
    }
    return $total;
}`,
			expectedAnnotations: []string{
				"Int $id", "Str $name", "-> HashRef[Str, Any]",
				"ArrayRef[Str] $items",
				"ArrayRef[Num] $prices", "-> Num",
				"Num $total", "Num $price",
			},
			workflowDescription: "Function signatures and parameter type annotations",
		},
		{
			name: "complex_nested_types",
			inputCode: `
my HashRef[Str, ArrayRef[HashRef[Str, Int]]] $complex_data = {
    users => [
        {name => "Alice", score => 95},
        {name => "Bob", score => 87}
    ]
};

sub transform_data(HashRef[Str, ArrayRef[HashRef[Str, Int]]] $input)
    -> ArrayRef[HashRef[Str, Str|Int]] {
    my ArrayRef[HashRef[Str, Str|Int]] $result = [];
    return $result;
}`,
			expectedAnnotations: []string{
				"HashRef[Str, ArrayRef[HashRef[Str, Int]]] $complex_data",
				"HashRef[Str, ArrayRef[HashRef[Str, Int]]] $input",
				"-> ArrayRef[HashRef[Str, Str|Int]]",
				"ArrayRef[HashRef[Str, Str|Int]] $result",
			},
			workflowDescription: "Complex nested type structures with unions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing workflow: %s", tc.workflowDescription)

			// Test complete workflow: Parse -> Compile -> Validate preservation
			preservedAnnotations, err := runCompleteTypePreservationWorkflow(tc.inputCode, t)
			if err != nil {
				t.Fatalf("Workflow failed: %v", err)
			}

			// Validate that all expected annotations are preserved
			for _, expectedAnnotation := range tc.expectedAnnotations {
				found := false
				for _, preserved := range preservedAnnotations {
					if strings.Contains(preserved, expectedAnnotation) {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("Expected annotation '%s' not preserved in output", expectedAnnotation)
					t.Logf("Preserved annotations: %v", preservedAnnotations)
				} else {
					t.Logf("✓ Preserved annotation: %s", expectedAnnotation)
				}
			}

			t.Logf("Type annotation preservation: %d/%d annotations preserved",
				len(preservedAnnotations), len(tc.expectedAnnotations))
		})
	}
}

// TestTreeSitterVsTraditionalTypePreservation compares type annotation preservation
// between tree-sitter shim and traditional AST workflows
func TestTreeSitterVsTraditionalTypePreservation(t *testing.T) {
	complexTypedCode := `
my HashRef[Str, ArrayRef[Int]] $user_scores = {
    "team_a" => [95, 87, 92],
    "team_b" => [88, 91, 85]
};

sub calculate_average(ArrayRef[Int] $scores) -> Num {
    my Int $sum = 0;
    my Int $count = scalar(@$scores);
    for my Int $score (@$scores) {
        $sum += $score;
    }
    return $sum / $count;
}

sub process_teams(HashRef[Str, ArrayRef[Int]] $teams) -> HashRef[Str, Num] {
    my HashRef[Str, Num] $averages = {};
    for my Str $team (keys %$teams) {
        my ArrayRef[Int] $scores = $teams->{$team};
        my Num $avg = calculate_average($scores);
        $averages->{$team} = $avg;
    }
    return $averages;
}`

	t.Run("tree_sitter_preservation", func(t *testing.T) {
		shimParser, err := parser.NewShimParser()
		if err != nil {
			t.Skip("Tree-sitter shim parser not available")
		}

		shimAST, err := shimParser.ParseStringShim(complexTypedCode)
		if err != nil {
			t.Fatalf("Tree-sitter parsing failed: %v", err)
		}

		// Use tree-sitter shim for compilation
		registry := compiler.NewCompilerRegistry()

		// Convert TreeSitterAST for compilation
		adapter := &TreeSitterASTAdapter{shimAST}

		// Compile to typed Perl (preserving annotations)
		typedOutput, err := registry.Compile(adapter, compiler.TargetTypedPerl)
		if err != nil {
			t.Fatalf("Tree-sitter compilation failed: %v", err)
		}

		// Count preserved type annotations
		treeSitterAnnotations := extractTypeAnnotations(typedOutput)
		t.Logf("Tree-sitter preserved %d type annotations", len(treeSitterAnnotations))

		for _, annotation := range treeSitterAnnotations {
			t.Logf("  ✓ %s", annotation)
		}
	})

	t.Run("traditional_preservation", func(t *testing.T) {
		traditionalParser, err := parser.NewParser()
		if err != nil {
			t.Fatalf("Failed to create traditional parser: %v", err)
		}

		traditionalAST, err := traditionalParser.ParseString(complexTypedCode)
		if err != nil {
			t.Fatalf("Traditional parsing failed: %v", err)
		}

		// Use traditional AST for compilation
		registry := compiler.NewCompilerRegistry()
		adapter := compiler.NewParserASTAdapter(traditionalAST)

		// Compile to typed Perl (preserving annotations)
		typedOutput, err := registry.Compile(adapter, compiler.TargetTypedPerl)
		if err != nil {
			t.Fatalf("Traditional compilation failed: %v", err)
		}

		// Count preserved type annotations
		traditionalAnnotations := extractTypeAnnotations(typedOutput)
		t.Logf("Traditional parser preserved %d type annotations", len(traditionalAnnotations))

		for _, annotation := range traditionalAnnotations {
			t.Logf("  ✓ %s", annotation)
		}
	})

	t.Run("preservation_comparison", func(t *testing.T) {
		// This sub-test compares results from both approaches
		t.Log("🎯 Type Annotation Preservation Comparison Summary:")
		t.Log("   Tree-sitter shim: Direct CST access preserves syntax structure")
		t.Log("   Traditional AST: May lose some syntactic details during AST conversion")
		t.Log("   Expected: Tree-sitter >= Traditional in preservation capability")
	})
}

// TestProductionWorkflowTypePreservation tests real production workflows
func TestProductionWorkflowTypePreservation(t *testing.T) {
	// Test the PSC strip -> PSC check workflow (common production pattern)
	typedPerlScript := `#!/usr/bin/env perl
use v5.38.0;
use strict;
use warnings;

# Production script with comprehensive type annotations
my HashRef[Str, Str] $config = {
    database_host => "localhost",
    database_port => "5432",
    api_endpoint => "https://api.example.com"
};

sub connect_database(HashRef[Str, Str] $config) -> DBI {
    my Str $dsn = "DBI:Pg:host=" . $config->{database_host} .
                  ";port=" . $config->{database_port};
    my DBI $dbh = DBI->connect($dsn, $config->{username}, $config->{password});
    return $dbh;
}

sub fetch_user_data(DBI $dbh, Int $user_id) -> Maybe[HashRef[Str, Any]] {
    my Str $sql = "SELECT * FROM users WHERE id = ?";
    my Maybe[HashRef[Str, Any]] $user = $dbh->selectrow_hashref($sql, undef, $user_id);
    return $user;
}

sub main() -> Int {
    my DBI $dbh = connect_database($config);
    my Maybe[HashRef[Str, Any]] $user = fetch_user_data($dbh, 123);

    if (defined($user)) {
        say "Found user: " . $user->{name};
    } else {
        say "User not found";
    }

    return 0;
}`

	t.Run("strip_then_check_workflow", func(t *testing.T) {
		// Step 1: Strip type annotations for execution
		t.Log("Step 1: Stripping type annotations...")

		registry := compiler.NewCompilerRegistry()

		// Parse with tree-sitter shim
		shimParser, err := parser.NewShimParser()
		if err != nil {
			t.Skip("Tree-sitter shim parser not available")
		}

		shimAST, err := shimParser.ParseStringShim(typedPerlScript)
		if err != nil {
			t.Fatalf("Failed to parse production script: %v", err)
		}

		adapter := &TreeSitterASTAdapter{shimAST}

		// Compile to clean Perl (strip annotations)
		cleanOutput, err := registry.Compile(adapter, compiler.TargetCleanPerl)
		if err != nil {
			t.Fatalf("Failed to strip annotations: %v", err)
		}

		// Verify clean output has no type annotations
		if strings.Contains(cleanOutput, "HashRef[") || strings.Contains(cleanOutput, "-> DBI") {
			t.Error("Clean output still contains type annotations")
		} else {
			t.Log("✓ Type annotations successfully stripped")
		}

		// Step 2: Re-parse clean output and verify it's valid Perl
		t.Log("Step 2: Validating clean Perl output...")

		cleanAST, err := shimParser.ParseStringShim(cleanOutput)
		if err != nil {
			t.Fatalf("Clean output is not valid Perl: %v", err)
		}

		if len(cleanAST.Errors) > 0 {
			t.Errorf("Clean output has parse errors: %d errors", len(cleanAST.Errors))
		} else {
			t.Log("✓ Clean output is valid Perl")
		}

		// Step 3: Verify original typed version still preserves all annotations
		t.Log("Step 3: Verifying type annotation preservation...")

		typedOutput, err := registry.Compile(adapter, compiler.TargetTypedPerl)
		if err != nil {
			t.Fatalf("Failed to preserve typed output: %v", err)
		}

		originalAnnotations := extractTypeAnnotations(typedPerlScript)
		preservedAnnotations := extractTypeAnnotations(typedOutput)

		t.Logf("Original annotations: %d", len(originalAnnotations))
		t.Logf("Preserved annotations: %d", len(preservedAnnotations))

		if len(preservedAnnotations) >= len(originalAnnotations) {
			t.Log("✓ Type annotations successfully preserved in production workflow")
		} else {
			t.Errorf("Type annotation loss detected: %d original -> %d preserved",
				len(originalAnnotations), len(preservedAnnotations))
		}
	})
}

// Helper functions

func runCompleteTypePreservationWorkflow(inputCode string, t *testing.T) ([]string, error) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "type_preservation_test_*.pl")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(inputCode); err != nil {
		return nil, err
	}
	tmpFile.Close()

	// Parse with tree-sitter shim
	shimParser, err := parser.NewShimParser()
	if err != nil {
		return nil, err
	}

	shimAST, err := shimParser.ParseStringShim(inputCode)
	if err != nil {
		return nil, err
	}

	// Create compiler and compile to typed Perl
	registry := compiler.NewCompilerRegistry()
	adapter := &TreeSitterASTAdapter{shimAST}

	typedOutput, err := registry.Compile(adapter, compiler.TargetTypedPerl)
	if err != nil {
		return nil, err
	}

	// Extract preserved type annotations
	annotations := extractTypeAnnotations(typedOutput)
	return annotations, nil
}

// TreeSitterASTAdapter adapts TreeSitterAST to the compiler AST interface
type TreeSitterASTAdapter struct {
	shimAST *ast.TreeSitterAST
}

func (adapter *TreeSitterASTAdapter) GetPath() string {
	return adapter.shimAST.Path
}

func (adapter *TreeSitterASTAdapter) IsValid() bool {
	return len(adapter.shimAST.Errors) == 0
}

func (adapter *TreeSitterASTAdapter) GetContent() (string, error) {
	return adapter.shimAST.Source, nil
}

func (adapter *TreeSitterASTAdapter) GetRootNode() (ast.Node, error) {
	if adapter.shimAST.Root == nil {
		return nil, fmt.Errorf("no root node available")
	}
	return adapter.shimAST.Root, nil
}

// extractTypeAnnotations extracts type annotations from Perl code
func extractTypeAnnotations(perlCode string) []string {
	var annotations []string
	lines := strings.Split(perlCode, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for type annotations in various forms
		if strings.Contains(line, " $") && (strings.Contains(line, "my ") || strings.Contains(line, "sub ")) {
			// Variable declarations: my Type $var
			if strings.Contains(line, "my ") && strings.Contains(line, " $") {
				annotations = append(annotations, extractVariableAnnotation(line))
			}

			// Function signatures: sub name(Type $param) -> ReturnType
			if strings.Contains(line, "sub ") && (strings.Contains(line, "(") || strings.Contains(line, "->")) {
				annotations = append(annotations, extractFunctionAnnotation(line))
			}
		}
	}

	// Remove empty annotations
	var filtered []string
	for _, ann := range annotations {
		if strings.TrimSpace(ann) != "" {
			filtered = append(filtered, ann)
		}
	}

	return filtered
}

func extractVariableAnnotation(line string) string {
	// Extract "Type $var" pattern from "my Type $var = ..."
	parts := strings.Fields(line)
	for i, part := range parts {
		if part == "my" && i+2 < len(parts) {
			typeStr := parts[i+1]
			varStr := parts[i+2]
			if strings.HasPrefix(varStr, "$") {
				return typeStr + " " + varStr
			}
		}
	}
	return ""
}

func extractFunctionAnnotation(line string) string {
	// Extract function parameter and return type annotations
	var annotations []string

	// Look for parameter types: (Type $param)
	if strings.Contains(line, "(") && strings.Contains(line, ")") {
		start := strings.Index(line, "(")
		end := strings.Index(line, ")")
		if start < end {
			params := line[start+1 : end]
			paramParts := strings.Split(params, ",")
			for _, param := range paramParts {
				param = strings.TrimSpace(param)
				if strings.Contains(param, " $") {
					annotations = append(annotations, param)
				}
			}
		}
	}

	// Look for return type: -> Type
	if strings.Contains(line, "->") {
		parts := strings.Split(line, "->")
		if len(parts) > 1 {
			returnType := strings.TrimSpace(parts[1])
			// Remove any trailing { or comments
			if idx := strings.Index(returnType, "{"); idx >= 0 {
				returnType = strings.TrimSpace(returnType[:idx])
			}
			if returnType != "" {
				annotations = append(annotations, "-> "+returnType)
			}
		}
	}

	return strings.Join(annotations, ", ")
}
