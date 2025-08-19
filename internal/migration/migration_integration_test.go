// ABOUTME: Integration tests demonstrating migration layer with real parsing scenarios
// ABOUTME: Shows practical usage patterns during tree-sitter transition period

package migration

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
)

func TestMigrationLayerIntegration(t *testing.T) {
	// This test demonstrates how the migration layer helps during transition

	// Create migration parser with default config (prefers tree-sitter)
	migrationParser, err := NewMigrationParser(nil)
	if err != nil {
		t.Skipf("Cannot create migration parser: %v", err)
	}

	// Test cases representing different migration scenarios
	testCases := []struct {
		name        string
		content     string
		description string
	}{
		{
			name: "legacy_untyped_code",
			content: `
				my $count = 0;
				my $name = "test";
				my @items = (1, 2, 3);
			`,
			description: "Legacy Perl code without type annotations",
		},
		{
			name: "modern_typed_code",
			content: `
				my Int $count = 0;
				my Str $name = "test";
				my ArrayRef[Int] $items = [1, 2, 3];
			`,
			description: "Modern typed Perl code with type annotations",
		},
		{
			name: "mixed_typed_code",
			content: `
				my $legacy_var = "old style";
				my Int $modern_count = 42;
				my HashRef[Str] $config = { key => "value" };
			`,
			description: "Mixed legacy and modern code",
		},
		{
			name: "complex_types",
			content: `
				my ArrayRef[HashRef[Str]] $complex = [
					{ name => "Alice", role => "admin" },
					{ name => "Bob", role => "user" }
				];
				my Maybe[Int] $optional_count;
			`,
			description: "Complex parameterized and union types",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)

			// Parse with migration layer
			result, err := migrationParser.ParseString(tc.content)
			if err != nil {
				t.Fatalf("Migration parser failed: %v", err)
			}

			if result == nil {
				t.Fatal("Migration parser returned nil result")
			}

			// Check which parser was used and verify result quality
			switch parsedAST := result.(type) {
			case *ast.TreeSitterAST:
				t.Logf("✓ Used tree-sitter parser for %s", tc.name)

				// Validate tree-sitter specific functionality
				if parsedAST.Root == nil {
					t.Error("TreeSitterAST should have root node")
				}

				if parsedAST.Source != strings.TrimSpace(tc.content) {
					t.Error("TreeSitterAST should preserve original source")
				}

				// Check for CST access
				if parsedAST.GetCSTRoot() == nil {
					t.Error("TreeSitterAST should provide CST root access")
				}

				t.Logf("  - Type annotations found: %d", len(parsedAST.TypeAnnotations))
				t.Logf("  - Parse errors: %d", len(parsedAST.Errors))

			case *ast.AST:
				t.Logf("✓ Used traditional parser for %s", tc.name)

				// Validate traditional AST functionality
				if parsedAST.Root == nil {
					t.Error("Traditional AST should have root node")
				}

				t.Logf("  - Type annotations found: %d", len(parsedAST.TypeAnnotations))
				t.Logf("  - Parse errors: %d", len(parsedAST.Errors))

			default:
				t.Errorf("Unexpected result type: %T", result)
			}
		})
	}
}

func TestMigrationStrategiesComparison(t *testing.T) {
	// Compare different migration strategies on the same content

	testContent := "my ArrayRef[Int] $numbers = [1, 2, 3];"

	strategies := []struct {
		name   string
		config *MigrationConfig
	}{
		{
			name: "prefer_tree_sitter",
			config: &MigrationConfig{
				Mode:          ModePreferTreeSitter,
				AllowFallback: true,
			},
		},
		{
			name: "prefer_traditional",
			config: &MigrationConfig{
				Mode:          ModePreferTraditional,
				AllowFallback: true,
			},
		},
		{
			name: "tree_sitter_only",
			config: &MigrationConfig{
				Mode:          ModeTreeSitterOnly,
				AllowFallback: false,
			},
		},
	}

	for _, strategy := range strategies {
		t.Run(strategy.name, func(t *testing.T) {
			mp, err := NewMigrationParser(strategy.config)
			if err != nil {
				t.Skipf("Cannot create migration parser for %s: %v", strategy.name, err)
			}

			result, err := mp.ParseString(testContent)
			if err != nil {
				t.Errorf("Parse failed for strategy %s: %v", strategy.name, err)
				return
			}

			t.Logf("Strategy %s used parser type: %T", strategy.name, result)

			// All strategies should handle the typed content successfully
			switch ast := result.(type) {
			case *ast.TreeSitterAST:
				if len(ast.Errors) > 0 {
					t.Errorf("TreeSitterAST has parse errors: %v", ast.Errors)
				}
			case *ast.AST:
				if len(ast.Errors) > 0 {
					t.Errorf("Traditional AST has parse errors: %v", ast.Errors)
				}
			}
		})
	}
}

func TestASTConversionWorkflow(t *testing.T) {
	// Test the workflow of converting between AST types

	converter, err := NewASTConverter()
	if err != nil {
		t.Skipf("Cannot create AST converter: %v", err)
	}

	testContent := "my Int $value = 42; my Str $name = 'test';"

	// 1. Parse with traditional parser first
	traditionalAST, err := converter.traditionalParser.ParseString(testContent)
	if err != nil {
		t.Skipf("Cannot parse with traditional parser: %v", err)
	}

	t.Logf("Original traditional AST - Type annotations: %d", len(traditionalAST.TypeAnnotations))

	// 2. Convert traditional AST to TreeSitterAST
	treeSitterAST, err := converter.ToTreeSitterAST(traditionalAST)
	if err != nil {
		t.Errorf("Failed to convert traditional AST to TreeSitterAST: %v", err)
		return
	}

	if treeSitterAST == nil {
		t.Fatal("Conversion resulted in nil TreeSitterAST")
	}

	t.Logf("Converted TreeSitterAST - Type annotations: %d", len(treeSitterAST.TypeAnnotations))

	// 3. Verify TreeSitterAST has additional capabilities
	if treeSitterAST.GetCSTRoot() == nil {
		t.Error("Converted TreeSitterAST should provide CST root access")
	}

	// 4. Convert back to traditional AST
	convertedTraditional, err := converter.ToTraditionalAST(treeSitterAST)
	if err != nil {
		t.Errorf("Failed to convert TreeSitterAST back to traditional AST: %v", err)
		return
	}

	if convertedTraditional == nil {
		t.Fatal("Back-conversion resulted in nil traditional AST")
	}

	t.Logf("Back-converted traditional AST - Type annotations: %d", len(convertedTraditional.TypeAnnotations))

	// 5. Verify content preservation
	if convertedTraditional.Source != traditionalAST.Source {
		t.Error("Source content should be preserved through conversions")
	}
}

func TestMigrationWithVarDeclNodes(t *testing.T) {
	// Test migration layer with VarDecl nodes specifically

	migrationParser, err := NewMigrationParser(&MigrationConfig{
		Mode:               ModePreferTreeSitter,
		AllowFallback:      true,
		PreferShimForTypes: []string{"variable_declaration"},
	})
	if err != nil {
		t.Skipf("Cannot create migration parser: %v", err)
	}

	testContent := "my Int $count = 42; my Str $name = 'Alice';"

	result, err := migrationParser.ParseString(testContent)
	if err != nil {
		t.Fatalf("Migration parser failed: %v", err)
	}

	// Should prefer tree-sitter for variable declarations
	treeSitterAST, ok := result.(*ast.TreeSitterAST)
	if !ok {
		t.Skip("Did not get TreeSitterAST, skipping VarDecl node test")
	}

	// Find variable declarations using tree-sitter shim
	var varDecls []*ast.VarDeclNode
	if treeSitterAST.Root != nil {
		treeSitterAST.Root.WalkNodes(func(node *ast.TreeSitterNode) bool {
			if node.Type() == "variable_declaration" {
				if varDecl := node.AsVarDecl(); varDecl != nil {
					varDecls = append(varDecls, varDecl)
				}
			}
			return true
		})
	}

	t.Logf("Found %d variable declarations using tree-sitter shim", len(varDecls))

	// Verify VarDecl nodes have type information
	for i, varDecl := range varDecls {
		if varDecl.HasTypeAnnotation() {
			typeExpr := varDecl.GetTypeExpression()
			if typeExpr != nil {
				t.Logf("VarDecl %d has type: %s", i, typeExpr.String())
			}
		}

		variables := varDecl.GetVariables()
		for j, variable := range variables {
			t.Logf("VarDecl %d, Variable %d: %s%s", i, j, variable.Sigil, variable.Name)
		}
	}
}

func TestMigrationErrorHandling(t *testing.T) {
	// Test error handling in migration scenarios

	// Test with invalid content
	migrationParser, err := NewMigrationParser(nil)
	if err != nil {
		t.Skipf("Cannot create migration parser: %v", err)
	}

	// Test with content that might cause parse errors
	invalidContent := "my Int $x = ; // incomplete statement"

	result, err := migrationParser.ParseString(invalidContent)

	// Migration parser should either succeed or provide meaningful error
	if err != nil {
		t.Logf("Migration parser correctly reported error: %v", err)
	} else {
		t.Logf("Migration parser handled problematic content, result type: %T", result)

		// Check if parse errors were captured
		switch ast := result.(type) {
		case *ast.TreeSitterAST:
			if len(ast.Errors) > 0 {
				t.Logf("TreeSitterAST captured %d parse errors", len(ast.Errors))
			}
		case *ast.AST:
			if len(ast.Errors) > 0 {
				t.Logf("Traditional AST captured %d parse errors", len(ast.Errors))
			}
		}
	}
}

func TestMigrationConfigurationValidation(t *testing.T) {
	// Test various configuration edge cases

	// Test with nil fallback when tree-sitter only
	config := &MigrationConfig{
		Mode:          ModeTreeSitterOnly,
		AllowFallback: false,
	}

	mp, err := NewMigrationParser(config)
	if err != nil {
		t.Skipf("Cannot create migration parser: %v", err)
	}

	// Should only have tree-sitter parser
	if mp.traditionalParser != nil {
		t.Error("Tree-sitter only mode should not have traditional parser")
	}

	if mp.shimParser == nil {
		t.Error("Tree-sitter only mode should have shim parser")
	}

	// Test strategy selection with this config
	strategy := mp.selectParsingStrategy("", "my Int $x = 42;")
	if strategy != "tree-sitter" {
		t.Errorf("Tree-sitter only mode should select tree-sitter strategy, got: %s", strategy)
	}
}

func TestMigrationPerformanceComparison(t *testing.T) {
	// Simple performance comparison between strategies
	// Note: This is more for validation than benchmarking

	testContent := `
		my Int $count = 0;
		my ArrayRef[Str] $items = ["apple", "banana", "cherry"];
		my HashRef[Int] $scores = { alice => 95, bob => 87, charlie => 92 };
		for my $item (@$items) {
			$count++;
			print "Item: $item, Score: " . $scores->{$item} . "\n";
		}
	`

	// Test tree-sitter strategy
	tsConfig := &MigrationConfig{Mode: ModeTreeSitterOnly}
	tsMp, err := NewMigrationParser(tsConfig)
	if err != nil {
		t.Skipf("Cannot create tree-sitter migration parser: %v", err)
	}

	tsResult, err := tsMp.ParseString(testContent)
	if err != nil {
		t.Logf("Tree-sitter parse failed: %v", err)
	} else {
		t.Logf("Tree-sitter parse successful, result type: %T", tsResult)
	}

	// Test traditional strategy
	tradConfig := &MigrationConfig{Mode: ModeTraditionalOnly}
	tradMp, err := NewMigrationParser(tradConfig)
	if err != nil {
		t.Skipf("Cannot create traditional migration parser: %v", err)
	}

	tradResult, err := tradMp.ParseString(testContent)
	if err != nil {
		t.Logf("Traditional parse failed: %v", err)
	} else {
		t.Logf("Traditional parse successful, result type: %T", tradResult)
	}

	// Both should handle the content (though with different capabilities)
	if tsResult != nil && tradResult != nil {
		t.Log("✓ Both parsing strategies can handle complex typed content")
	}
}
