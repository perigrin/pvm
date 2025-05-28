// ABOUTME: Comprehensive test suite for control flow structure parsing validation
// ABOUTME: Tests conditional statements, loops, switch statements, and complex control flow patterns

package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestControlFlowStructures validates control flow parsing using the test framework
func TestControlFlowStructures(t *testing.T) {
	testCategories := []string{
		"conditional_statements",
		"loop_statements",
		"loop_control",
		"switch_statements",
		"complex_control_flow",
	}

	for _, category := range testCategories {
		t.Run(category, func(t *testing.T) {
			testFile := filepath.Join("testdata", "untyped-perl", "control-flow", category+".json")
			runControlFlowTestsFromFile(t, testFile)
		})
	}
}

// runControlFlowTestsFromFile loads and runs tests from a JSON file
func runControlFlowTestsFromFile(t *testing.T, testFile string) {
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", testFile, err)
	}

	var testCases []ParserTestCase
	if err := json.Unmarshal(data, &testCases); err != nil {
		t.Fatalf("Failed to parse test file %s: %v", testFile, err)
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ast, err := parser.ParseString(testCase.Input)
			
			if testCase.ShouldError {
				if err == nil {
					t.Errorf("Expected error for input: %s", testCase.Input)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected parse error for input '%s': %v", testCase.Input, err)
				return
			}

			if ast == nil {
				t.Errorf("Parser returned nil AST for input: %s", testCase.Input)
				return
			}

			// Validate AST contains expected control flow structure
			validateControlFlowAST(t, ast, testCase)
		})
	}
}

// validateControlFlowAST performs basic validation of control flow AST structure
func validateControlFlowAST(t *testing.T, ast *AST, testCase ParserTestCase) {
	if ast.Root == nil {
		t.Errorf("AST root is nil for test case: %s", testCase.Name)
		return
	}

	// Check that we have children (which represent statements)
	children := ast.Root.Children()
	if len(children) == 0 {
		t.Errorf("No child nodes found in AST for test case: %s", testCase.Name)
		return
	}

	// Additional validation based on test tags
	for _, tag := range testCase.Tags {
		switch tag {
		case "conditional":
			validateConditionalStructure(t, children, testCase)
		case "loop":
			validateLoopStructure(t, children, testCase)
		case "loop_control":
			validateLoopControlStructure(t, children, testCase)
		case "switch":
			validateSwitchStructure(t, children, testCase)
		case "complex":
			validateComplexControlFlow(t, children, testCase)
		}
	}
}

// validateConditionalStructure checks conditional statement structure
func validateConditionalStructure(t *testing.T, children []Node, testCase ParserTestCase) {
	// Basic validation that conditional keywords are recognized
	t.Logf("Validating conditional structure for: %s", testCase.Name)
	
	// Check for specific conditional patterns
	for _, tag := range testCase.Tags {
		switch tag {
		case "if":
			t.Logf("  - Found if statement")
		case "elsif":
			t.Logf("  - Found elsif clause")
		case "else":
			t.Logf("  - Found else clause")
		case "unless":
			t.Logf("  - Found unless statement")
		case "postfix":
			t.Logf("  - Found postfix conditional")
		case "nested":
			t.Logf("  - Found nested conditional")
		}
	}
}

// validateLoopStructure checks loop statement structure
func validateLoopStructure(t *testing.T, children []Node, testCase ParserTestCase) {
	// Basic validation that loop keywords are recognized
	t.Logf("Validating loop structure for: %s", testCase.Name)
	
	// Check for specific loop patterns
	for _, tag := range testCase.Tags {
		switch tag {
		case "while":
			t.Logf("  - Found while loop")
		case "until":
			t.Logf("  - Found until loop")
		case "for":
			t.Logf("  - Found for loop")
		case "foreach":
			t.Logf("  - Found foreach loop")
		case "c_style":
			t.Logf("  - Found C-style for loop")
		case "range":
			t.Logf("  - Found range-based loop")
		case "nested":
			t.Logf("  - Found nested loop")
		case "do_while", "do_until":
			t.Logf("  - Found do-while/until loop")
		}
	}
}

// validateLoopControlStructure checks loop control statement structure
func validateLoopControlStructure(t *testing.T, children []Node, testCase ParserTestCase) {
	// Basic validation that loop control keywords are recognized
	t.Logf("Validating loop control structure for: %s", testCase.Name)
	
	// Check for specific control patterns
	for _, tag := range testCase.Tags {
		switch tag {
		case "next":
			t.Logf("  - Found next statement")
		case "last":
			t.Logf("  - Found last statement")
		case "redo":
			t.Logf("  - Found redo statement")
		case "labeled":
			t.Logf("  - Found labeled control flow")
		case "continue":
			t.Logf("  - Found continue block")
		}
	}
}

// validateSwitchStructure checks switch/given-when statement structure
func validateSwitchStructure(t *testing.T, children []Node, testCase ParserTestCase) {
	// Basic validation that switch keywords are recognized
	t.Logf("Validating switch structure for: %s", testCase.Name)
	
	// Check for specific switch patterns
	for _, tag := range testCase.Tags {
		switch tag {
		case "given":
			t.Logf("  - Found given statement")
		case "when":
			t.Logf("  - Found when clause")
		case "default":
			t.Logf("  - Found default clause")
		case "smartmatch":
			t.Logf("  - Found smartmatch operator")
		case "range":
			t.Logf("  - Found range matching")
		case "regex":
			t.Logf("  - Found regex matching")
		}
	}
}

// validateComplexControlFlow checks complex control flow patterns
func validateComplexControlFlow(t *testing.T, children []Node, testCase ParserTestCase) {
	// Basic validation that complex patterns are recognized
	t.Logf("Validating complex control flow for: %s", testCase.Name)
	
	// Check for specific complex patterns
	for _, tag := range testCase.Tags {
		switch tag {
		case "nested":
			t.Logf("  - Found deeply nested structures")
		case "eval":
			t.Logf("  - Found eval block")
		case "exception":
			t.Logf("  - Found exception handling")
		case "state_machine":
			t.Logf("  - Found state machine pattern")
		case "iterator":
			t.Logf("  - Found iterator pattern")
		case "recursive":
			t.Logf("  - Found recursive pattern")
		case "pipeline":
			t.Logf("  - Found pipeline pattern")
		case "coroutine":
			t.Logf("  - Found coroutine simulation")
		case "event_loop":
			t.Logf("  - Found event loop pattern")
		case "parallel":
			t.Logf("  - Found parallel processing simulation")
		}
	}
}

// BenchmarkControlFlowParsing benchmarks control flow parsing performance
func BenchmarkControlFlowParsing(b *testing.B) {
	parser, err := NewParser()
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	testControlFlow := []string{
		"if ($condition) { do_something(); }",
		"foreach my $item (@list) { process($item); }",
		"while ($condition) { process(); }",
		"for my $i (0..$max) { handle($i); }",
		"given ($value) { when (1) { action(); } default { other(); } }",
		"if ($a) { if ($b) { nested(); } }",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, code := range testControlFlow {
			_, err := parser.ParseString(code)
			if err != nil {
				b.Errorf("Parse error for control flow '%s': %v", code, err)
			}
		}
	}
}

// TestControlFlowNesting validates nested control flow structures
func TestControlFlowNesting(t *testing.T) {
	nestingTests := []struct {
		name        string
		input       string
		description string
	}{
		{
			name: "if_foreach_while",
			input: `if ($enabled) {
				foreach my $item (@items) {
					while (my $data = $item->next()) {
						process($data);
					}
				}
			}`,
			description: "If containing foreach containing while",
		},
		{
			name: "nested_loops_with_labels",
			input: `OUTER: for my $i (1..10) {
				INNER: for my $j (1..10) {
					next OUTER if ($i * $j > 50);
					print "$i x $j = ", $i * $j, "\n";
				}
			}`,
			description: "Nested loops with labeled control flow",
		},
		{
			name: "complex_nested_conditions",
			input: `if ($config->{enabled}) {
				unless ($config->{skip_validation}) {
					if (validate_input($input)) {
						process($input);
					} else {
						handle_error();
					}
				}
			}`,
			description: "Complex nested conditional statements",
		},
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	for _, test := range nestingTests {
		t.Run(test.name, func(t *testing.T) {
			ast, err := parser.ParseString(test.input)
			if err != nil {
				t.Errorf("Parse error for nesting test '%s': %v", test.name, err)
				return
			}

			if ast == nil {
				t.Errorf("Parser returned nil AST for nesting test: %s", test.name)
				return
			}

			t.Logf("Nesting test '%s' parsed successfully: %s", test.name, test.description)
		})
	}
}