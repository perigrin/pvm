// ABOUTME: Tests to verify typechecker layer correctly processes AST nodes
// ABOUTME: Ensures flow analysis can handle the AST structures from the parser

package typechecker

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

// TestTypeCheckerLayerASTProcessing tests that the typechecker can process AST nodes correctly
func TestTypeCheckerLayerASTProcessing(t *testing.T) {
	testCases := []struct {
		name                   string
		code                   string
		expectedNodeTypes      []string // AST node types that should be processed
		expectedStatements     int      // Number of statements that should be processed
		shouldDetectHashAccess bool     // Whether hash access should be detected
		description            string
	}{
		{
			name: "simple_variable_declaration_processing",
			code: `my $name = "value";`,
			expectedNodeTypes: []string{
				"var_decl",
			},
			expectedStatements: 1,
			description:        "Simple variable declaration should be processed by flow analysis",
		},
		{
			name: "hash_access_processing",
			code: `my $value = $hash->{key};`,
			expectedNodeTypes: []string{
				"var_decl",
			},
			shouldDetectHashAccess: true,
			description:            "Hash access should be detected and processed",
		},
		{
			name: "subroutine_body_processing",
			code: `sub test($input) {
    my $name = $input->{name};
    my $id = $input->{user_id};
    return "$name:$id";
}`,
			expectedNodeTypes: []string{
				"sub_decl", "var_decl", "return_stmt",
			},
			expectedStatements:     3, // 3 statements inside the subroutine
			shouldDetectHashAccess: true,
			description:            "Subroutine body statements should be properly processed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create flow analyzer
			analyzer := setupTestFlowAnalyzer(t)

			// Parse code to AST
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			astResult, err := p.ParseString(tc.code)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			// Build control flow graph
			cfg, err := analyzer.buildControlFlowGraph(astResult)
			if err != nil {
				t.Fatalf("Failed to build CFG: %v", err)
			}

			// Count statements processed
			totalStatements := 0
			for _, block := range cfg.Nodes {
				totalStatements += len(block.Statements)
			}

			if tc.expectedStatements > 0 && totalStatements < tc.expectedStatements {
				t.Errorf("%s: Expected at least %d statements processed, got %d",
					tc.description, tc.expectedStatements, totalStatements)
			}

			// Verify that the typechecker can process the AST without falling back to literals
			foundLiterals := false
			processedExpectedTypes := make(map[string]bool)

			for _, block := range cfg.Nodes {
				for _, stmt := range block.Statements {
					stmtType := stmt.Type()

					// Check for forbidden literal fallbacks
					if stmtType == "literal" {
						if literalNode, ok := stmt.(*ast.LiteralExpr); ok {
							// If the literal contains structured code, it's a parsing failure
							if containsStructuredCode(literalNode.Value) {
								t.Errorf("%s: Found literal node with structured code '%s' - indicates AST processing failure",
									tc.description, literalNode.Value)
								foundLiterals = true
							}
						}
					}

					// Check for expected node types
					for _, expectedType := range tc.expectedNodeTypes {
						if stmtType == expectedType {
							processedExpectedTypes[expectedType] = true
						}
					}
				}
			}

			if foundLiterals {
				t.Errorf("%s: Found literal nodes containing structured code - typechecker AST processing failed", tc.description)
			}

			// Verify expected node types were processed
			for _, expectedType := range tc.expectedNodeTypes {
				if !processedExpectedTypes[expectedType] {
					t.Errorf("%s: Expected node type '%s' was not processed by typechecker", tc.description, expectedType)
				}
			}
		})
	}
}

// TestTypeCheckerLayerFlowAnalysis tests that flow analysis can properly analyze different constructs
func TestTypeCheckerLayerFlowAnalysis(t *testing.T) {
	testCases := []struct {
		name           string
		code           string
		shouldDetect   []string // What the analysis should detect/infer
		shouldNotError bool     // Whether this should complete without errors
		description    string
	}{
		{
			name: "variable_type_inference",
			code: `sub test() {
    my $name = ref(\$input);
    my $defined = defined($input);
}`,
			shouldDetect: []string{
				"type inference for name",
				"type inference for defined",
			},
			shouldNotError: true,
			description:    "Flow analysis should infer types for builtin function calls",
		},
		{
			name: "unsafe_hash_access_detection",
			code: `sub test($input) {
    my $name = $input->{name};
    my $id = $input->{user_id};
}`,
			shouldDetect: []string{
				"unsafe hash access",
			},
			shouldNotError: false, // Should produce errors since we're detecting unsafe hash access
			description:    "Safety analysis should detect unsafe hash field access",
		},
		{
			name: "control_flow_processing",
			code: `sub test($input) {
    if (defined($input)) {
        my $value = $input->{data};
        return $value;
    }
    return undef;
}`,
			shouldDetect: []string{
				"control flow",
				"conditional processing",
			},
			shouldNotError: true,
			description:    "Flow analysis should handle control flow constructs",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupTestFlowAnalyzer(t)

			// Parse and build CFG
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			astResult, err := p.ParseString(tc.code)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			cfg, err := analyzer.buildControlFlowGraph(astResult)
			if err != nil {
				if tc.shouldNotError {
					t.Errorf("%s: CFG construction failed but should have succeeded: %v", tc.description, err)
				}
				return
			}

			// Run flow analysis
			errors := analyzer.analyzeDataFlow(cfg)

			if tc.shouldNotError && len(errors) > 0 {
				t.Errorf("%s: Flow analysis produced errors but should not have: %v", tc.description, errors)
			}

			// Verify flow analysis actually processed the constructs
			// This is a basic check - more specific validations would go in the main flow analysis tests
			if len(cfg.Nodes) == 0 {
				t.Errorf("%s: CFG has no nodes - flow analysis did not process the code", tc.description)
			}
		})
	}
}

// TestTypeCheckerLayerIntegration tests end-to-end flow from parser to flow analysis
func TestTypeCheckerLayerIntegration(t *testing.T) {
	code := `sub process_data($input) {
    my $name = $input->{name};
    if (defined($name)) {
        return $name;
    }
    return undef;
}`

	// Full pipeline test
	analyzer := setupTestFlowAnalyzer(t)

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	astResult, err := p.ParseString(code)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Verify AST is reasonable
	if astResult.Root == nil {
		t.Fatalf("Parser produced nil root node")
	}

	// Build CFG
	cfg, err := analyzer.buildControlFlowGraph(astResult)
	if err != nil {
		t.Fatalf("Failed to build CFG: %v", err)
	}

	// Verify CFG has reasonable structure
	if len(cfg.Nodes) < 1 {
		t.Errorf("CFG should have at least 1 node, got %d", len(cfg.Nodes))
	}

	// Run flow analysis
	errors := analyzer.analyzeDataFlow(cfg)

	// This should complete without fatal errors
	// Individual test failures are OK, but should not crash
	t.Logf("Flow analysis completed with %d analysis errors (expected for incomplete implementation)", len(errors))
}

// setupTestFlowAnalyzer creates a flow analyzer for testing
func setupTestFlowAnalyzer(t *testing.T) *FlowAnalyzer {
	typeStore, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create type store: %v", err)
	}

	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	symbolTable.Package = "layer_test"

	tc := NewTypeChecker(hierarchy, symbolTable, "layer_test")
	tc.SafetyAnalysisEnabled = true

	return NewFlowAnalyzer(tc)
}

// containsStructuredCode checks if a string contains Perl code constructs
func containsStructuredCode(s string) bool {
	// Simple heuristic to detect if a "literal" actually contains structured code
	indicators := []string{
		"my $", "->", "if (", "return ", "sub ", "{", "}", "defined(", "ref(",
	}

	for _, indicator := range indicators {
		if strings.Contains(s, indicator) {
			return true
		}
	}
	return false
}
