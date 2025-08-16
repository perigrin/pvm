// ABOUTME: Data flow validation tests to trace parser to type inference pipeline
// ABOUTME: Tests each level of data transformation to maintain correct behavior

package typechecker

import (
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

// TestDataFlowFromParserToTypeInference validates VarDecl nodes at each level
func TestDataFlowFromParserToTypeInference(t *testing.T) {
	testCode := `my $type = ref($input);`

	// Level 1: Test Parser Output
	t.Run("Level1_ParserOutput", func(t *testing.T) {
		p, err := parser.NewParser()
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		result, err := p.ParseString(testCode)
		if err != nil {
			t.Fatalf("Failed to parse code: %v", err)
		}

		// Find VarDecl node in AST
		varDecl := findVarDeclInAST(result.Root)
		if varDecl == nil {
			t.Fatal("No VarDecl node found in parser output")
		}

		t.Logf("Level 1 - Parser VarDecl: Type=%s, Initializer=%v",
			varDecl.Type(), varDecl.Initializer != nil)

		if varDecl.Initializer == nil {
			t.Error("Parser VarDecl should have initializer")
		} else {
			t.Logf("Parser initializer: Type=%s, Text=%q",
				varDecl.Initializer.Type(), varDecl.Initializer.Text())
		}
	})

	// Level 2: Test FlowAnalyzer.buildControlFlowGraph
	t.Run("Level2_CFGConstruction", func(t *testing.T) {
		analyzer := setupDataFlowTestAnalyzer(t)

		// Parse the code
		p, err := parser.NewParser()
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}
		result, err := p.ParseString(testCode)
		if err != nil {
			t.Fatalf("Failed to parse code: %v", err)
		}

		// Build CFG
		cfg, err := analyzer.buildControlFlowGraph(result)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Check VarDecl nodes in CFG statements (might be wrapped in expression_statement)
		var cfgVarDecl *ast.VarDecl
		for _, block := range cfg.Nodes {
			for _, stmt := range block.Statements {
				if stmt.Type() == "var_decl" {
					if vd, ok := stmt.(*ast.VarDecl); ok {
						cfgVarDecl = vd
						break
					}
				} else if stmt.Type() == "expression_statement" {
					// Look for VarDecl inside expression_statement
					for _, child := range stmt.Children() {
						if child.Type() == "var_decl" {
							if vd, ok := child.(*ast.VarDecl); ok {
								cfgVarDecl = vd
								break
							}
						}
					}
				}
				if cfgVarDecl != nil {
					break
				}
			}
		}

		if cfgVarDecl == nil {
			t.Fatal("No VarDecl node found in CFG statements")
		}

		t.Logf("Level 2 - CFG VarDecl: Type=%s, Initializer=%v",
			cfgVarDecl.Type(), cfgVarDecl.Initializer != nil)

		if cfgVarDecl.Initializer == nil {
			t.Error("CFG VarDecl should have initializer")
		} else {
			t.Logf("CFG initializer: Type=%s, Text=%q",
				cfgVarDecl.Initializer.Type(), cfgVarDecl.Initializer.Text())
		}
	})

	// Level 3: Test FlowAnalyzer.processBlock
	t.Run("Level3_BlockProcessing", func(t *testing.T) {
		analyzer := setupDataFlowTestAnalyzer(t)

		// Parse and build CFG
		p, err := parser.NewParser()
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}
		result, err := p.ParseString(testCode)
		if err != nil {
			t.Fatalf("Failed to parse code: %v", err)
		}
		cfg, err := analyzer.buildControlFlowGraph(result)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Process the first block that contains statements
		var targetBlock *BasicBlock
		for _, block := range cfg.Nodes {
			if len(block.Statements) > 0 {
				targetBlock = block
				break
			}
		}

		if targetBlock == nil {
			t.Fatal("No block with statements found")
		}

		// Process the block and check that it contains our expected variable
		_ = analyzer.processBlock(targetBlock)

		// Check if the block contains the expected statement
		foundVarDecl := false
		for _, stmt := range targetBlock.Statements {
			if stmt.Type() == "var_decl" || stmt.Type() == "expression_statement" {
				foundVarDecl = true
				break
			}
		}

		if !foundVarDecl {
			t.Fatal("No VarDecl or expression_statement found in processed block")
		}

		t.Logf("Level 3 - Block processing completed with %d statements", len(targetBlock.Statements))
	})

	// Level 4: Test Complete Type Inference Pipeline
	t.Run("Level4_TypeInference", func(t *testing.T) {
		analyzer := setupDataFlowTestAnalyzer(t)

		// Parse, build CFG, and run type inference
		p, err := parser.NewParser()
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}
		result, err := p.ParseString(testCode)
		if err != nil {
			t.Fatalf("Failed to parse code: %v", err)
		}
		cfg, err := analyzer.buildControlFlowGraph(result)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Run data flow analysis
		_ = analyzer.analyzeDataFlow(cfg)

		// Check if 'type' variable got correct type inference
		found := false
		for _, block := range cfg.Nodes {
			if block.ExitTypeState != nil && block.ExitTypeState.VariableTypes != nil {
				if varType, exists := block.ExitTypeState.VariableTypes["type"]; exists {
					t.Logf("Level 4 - Type inference result: $type = %s", varType)
					if varType == "Str" {
						found = true
					}
				}
			}
		}

		if !found {
			t.Error("Type inference should have inferred 'type' variable as 'Str'")
		}
	})
}

// TestNodeIdentityPreservation tests that VarDecl node identity is preserved
func TestNodeIdentityPreservation(t *testing.T) {
	testCode := `my $test = defined($value);`

	// Track the same VarDecl node through the pipeline
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	result, err := p.ParseString(testCode)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Get original VarDecl from parser
	originalVarDecl := findVarDeclInAST(result.Root)
	if originalVarDecl == nil {
		t.Fatal("No VarDecl found in original AST")
	}

	analyzer := setupDataFlowTestAnalyzer(t)
	cfg, err := analyzer.buildControlFlowGraph(result)
	if err != nil {
		t.Fatalf("Failed to build CFG: %v", err)
	}

	// Find VarDecl in CFG
	var cfgVarDecl *ast.VarDecl
	for _, block := range cfg.Nodes {
		for _, stmt := range block.Statements {
			if vd, ok := stmt.(*ast.VarDecl); ok {
				cfgVarDecl = vd
				break
			}
		}
	}

	if cfgVarDecl == nil {
		t.Fatal("No VarDecl found in CFG")
	}

	// Test if they are the same object
	if originalVarDecl != cfgVarDecl {
		t.Error("VarDecl node identity changed between parser and CFG construction")
		t.Logf("Original: %p, CFG: %p", originalVarDecl, cfgVarDecl)
		t.Logf("Original initializer: %v, CFG initializer: %v",
			originalVarDecl.Initializer != nil, cfgVarDecl.Initializer != nil)
	} else {
		t.Log("VarDecl node identity preserved correctly")
	}
}

// Helper functions

func setupDataFlowTestAnalyzer(t *testing.T) *FlowAnalyzer {
	typeStore, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create type store: %v", err)
	}

	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	symbolTable.Package = "test"

	tc := NewTypeChecker(hierarchy, symbolTable, "test")
	tc.SafetyAnalysisEnabled = true

	analyzer := NewFlowAnalyzer(tc)
	return analyzer
}

func findVarDeclInAST(node ast.Node) *ast.VarDecl {
	if node == nil {
		return nil
	}

	// Check if this node is a VarDecl
	if vd, ok := node.(*ast.VarDecl); ok {
		return vd
	}

	// Search children recursively
	for _, child := range node.Children() {
		if result := findVarDeclInAST(child); result != nil {
			return result
		}
	}

	return nil
}
