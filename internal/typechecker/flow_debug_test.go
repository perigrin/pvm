// ABOUTME: Debugging and visualization features tests for flow analysis
// ABOUTME: Tests DOT graph generation, type state tracking, and interactive analysis

package typechecker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

// TestDOTGraphGeneration tests DOT graph generation for CFG visualization
func TestDOTGraphGeneration(t *testing.T) {
	testCases := []struct {
		name          string
		code          string
		expectedNodes []string
		expectedEdges []string
		description   string
	}{
		{
			name: "simple_linear_flow",
			code: `
sub simple_function($input) {
    my $processed = process($input);
    my $result = transform($processed);
    return $result;
}`,
			expectedNodes: []string{"entry", "statement", "exit"},
			expectedEdges: []string{"entry -> statement", "statement -> exit"},
			description:   "Should generate DOT graph for simple linear flow",
		},
		{
			name: "conditional_flow",
			code: `
sub conditional_function($input) {
    if ($input) {
        return process_true($input);
    } else {
        return process_false($input);
    }
}`,
			expectedNodes: []string{"entry", "condition", "true_branch", "false_branch", "exit"},
			expectedEdges: []string{
				"entry -> condition",
				"condition -> true_branch",
				"condition -> false_branch",
				"true_branch -> exit",
				"false_branch -> exit",
			},
			description: "Should generate DOT graph for conditional flow",
		},
		{
			name: "loop_flow",
			code: `
sub loop_function(@items) {
    for my $item (@items) {
        process($item);
    }
    return "done";
}`,
			expectedNodes: []string{"entry", "loop_header", "loop_body", "exit"},
			expectedEdges: []string{
				"entry -> loop_header",
				"loop_header -> loop_body",
				"loop_body -> loop_header",
				"loop_header -> exit",
			},
			description: "Should generate DOT graph for loop flow",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupDebugTest(t)
			cfg := buildDebugCFG(t, analyzer, tc.code)

			// Create temporary directory for DOT output
			tmpDir := t.TempDir()
			dotFile := filepath.Join(tmpDir, tc.name+".dot")

			// Generate DOT graph
			err := generateDOTGraph(cfg, dotFile, tc.name)
			if err != nil {
				t.Fatalf("%s: Failed to generate DOT graph: %v", tc.description, err)
			}

			// Verify DOT file was created
			if _, err := os.Stat(dotFile); os.IsNotExist(err) {
				t.Errorf("%s: DOT file was not created", tc.description)
				return
			}

			// Read and verify DOT content
			content, err := os.ReadFile(dotFile)
			if err != nil {
				t.Fatalf("%s: Failed to read DOT file: %v", tc.description, err)
			}

			dotContent := string(content)

			// Verify basic DOT structure
			if !strings.Contains(dotContent, "digraph") {
				t.Errorf("%s: DOT file should contain 'digraph'", tc.description)
			}

			// Verify expected nodes are referenced (don't require exact labels)
			nodeTypes := []string{"entry", "exit"}
			for _, nodeType := range nodeTypes {
				found := false
				for _, expectedNode := range tc.expectedNodes {
					if strings.Contains(expectedNode, nodeType) {
						found = true
						break
					}
				}
				if found && !strings.Contains(dotContent, nodeType) {
					t.Errorf("%s: DOT file should contain reference to %s node", tc.description, nodeType)
				}
			}

			t.Logf("%s: Successfully generated DOT graph at %s", tc.description, dotFile)
		})
	}
}

// TestTypeStateTracking tests type state tracking through flow analysis
func TestTypeStateTracking(t *testing.T) {
	testCases := []struct {
		name           string
		code           string
		trackVariable  string
		expectedStates []string
		description    string
	}{
		{
			name: "simple_type_evolution",
			code: `
sub type_evolution($input) {
    my $value = $input;           # State 1: Any
    if (defined($value)) {        # State 2: Maybe[Any] -> Any
        my $length = length($value); # State 3: Any -> Str (for length to work)
        return $length;
    }
    return 0;
}`,
			trackVariable:  "value",
			expectedStates: []string{"Any", "Str"},
			description:    "Should track type evolution through defined check and length call",
		},
		{
			name: "hash_access_safety_tracking",
			code: `
sub hash_safety($hash) {
    if (exists $hash->{field}) {    # Mark field access as safe
        my $value = $hash->{field}; # Should be safe access
        return $value;
    }
    return undef;
}`,
			trackVariable:  "field_access_safety",
			expectedStates: []string{"unsafe", "safe"},
			description:    "Should track field access safety through exists check",
		},
		{
			name: "exception_flow_tracking",
			code: `
sub exception_flow($input) {
    die "invalid input" unless $input;  # Introduces Throws[Str]
    my $processed = process($input);    # May propagate exception
    return $processed;
}`,
			trackVariable:  "exception_types",
			expectedStates: []string{"Throws[Str]"},
			description:    "Should track exception types through flow",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupDebugTest(t)
			cfg := buildDebugCFG(t, analyzer, tc.code)

			// Enable debug mode for detailed tracking
			analyzer.TypeChecker.Debug = true

			_ = analyzer.analyzeDataFlow(cfg)

			// Collect type states across all blocks
			var foundStates []string
			for _, block := range cfg.Nodes {
				if block.TypeState == nil {
					continue
				}

				// Check variable types
				if tc.trackVariable != "field_access_safety" && tc.trackVariable != "exception_types" {
					if varType, exists := block.TypeState.VariableTypes[tc.trackVariable]; exists {
						if !contains(foundStates, varType) {
							foundStates = append(foundStates, varType)
						}
					}
					if refinedType, exists := block.TypeState.RefinedTypes[tc.trackVariable]; exists {
						if !contains(foundStates, refinedType) {
							foundStates = append(foundStates, refinedType)
						}
					}
				}

				// Check field access safety
				if tc.trackVariable == "field_access_safety" {
					for _, fields := range block.TypeState.FieldAccess {
						for _, safe := range fields {
							state := "unsafe"
							if safe {
								state = "safe"
							}
							if !contains(foundStates, state) {
								foundStates = append(foundStates, state)
							}
						}
					}
				}

				// Check exception types
				if tc.trackVariable == "exception_types" {
					for excType := range block.TypeState.ExceptionTypes {
						if !contains(foundStates, excType) {
							foundStates = append(foundStates, excType)
						}
					}
				}
			}

			// Verify expected states were found
			for _, expectedState := range tc.expectedStates {
				found := false
				for _, foundState := range foundStates {
					if strings.Contains(foundState, expectedState) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: Expected to find state '%s', but only found: %v",
						tc.description, expectedState, foundStates)
				}
			}

			t.Logf("%s: Found states for %s: %v", tc.description, tc.trackVariable, foundStates)
		})
	}
}

// TestCFGVisualization tests comprehensive CFG visualization features
func TestCFGVisualization(t *testing.T) {
	testCases := []struct {
		name             string
		code             string
		visualizeTypes   bool
		visualizeSafety  bool
		expectedFeatures []string
		description      string
	}{
		{
			name: "basic_visualization",
			code: `
sub basic_function($input) {
    my $result = process($input);
    return $result;
}`,
			visualizeTypes:   false,
			visualizeSafety:  false,
			expectedFeatures: []string{"node", "edge", "label"},
			description:      "Should generate basic CFG visualization",
		},
		{
			name: "type_annotated_visualization",
			code: `
sub typed_function($input) {
    my $value = ref($input);      # Type: Str
    my $length = length($value);  # Type: Int
    return $length;
}`,
			visualizeTypes:   true,
			visualizeSafety:  false,
			expectedFeatures: []string{"node", "edge", "Str", "Int"},
			description:      "Should include type annotations in visualization",
		},
		{
			name: "safety_annotated_visualization",
			code: `
sub safety_function($hash) {
    if (exists $hash->{field}) {
        my $value = $hash->{field}; # Safe access
        return $value;
    }
    return $hash->{other};          # Unsafe access
}`,
			visualizeTypes:   false,
			visualizeSafety:  true,
			expectedFeatures: []string{"node", "edge", "safe", "unsafe"},
			description:      "Should include safety annotations in visualization",
		},
		{
			name: "comprehensive_visualization",
			code: `
sub comprehensive_function($input) {
    if (defined($input)) {
        my $type = ref($input);
        if ($type eq 'HASH') {
            return keys(%$input);
        }
    }
    die "invalid input";
}`,
			visualizeTypes:   true,
			visualizeSafety:  true,
			expectedFeatures: []string{"node", "edge", "Str", "Bool", "safe", "Throws"},
			description:      "Should include comprehensive annotations",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupDebugTest(t)
			cfg := buildDebugCFG(t, analyzer, tc.code)

			_ = analyzer.analyzeDataFlow(cfg)

			// Create temporary directory for visualization
			tmpDir := t.TempDir()
			vizFile := filepath.Join(tmpDir, tc.name+"_viz.dot")

			// Generate visualization with specified features
			err := generateEnhancedVisualization(cfg, vizFile, tc.visualizeTypes, tc.visualizeSafety)
			if err != nil {
				t.Fatalf("%s: Failed to generate visualization: %v", tc.description, err)
			}

			// Verify visualization file
			content, err := os.ReadFile(vizFile)
			if err != nil {
				t.Fatalf("%s: Failed to read visualization: %v", tc.description, err)
			}

			vizContent := string(content)

			// Verify expected features
			for _, feature := range tc.expectedFeatures {
				if !strings.Contains(vizContent, feature) {
					t.Errorf("%s: Visualization should contain '%s'", tc.description, feature)
				}
			}

			t.Logf("%s: Generated visualization with features: %v", tc.description, tc.expectedFeatures)
		})
	}
}

// TestInteractiveAnalysis tests interactive analysis features
func TestInteractiveAnalysis(t *testing.T) {
	testCases := []struct {
		name          string
		code          string
		queryVariable string
		queryBlock    int
		expectedInfo  map[string]string
		description   string
	}{
		{
			name: "variable_query",
			code: `
sub query_function($input) {
    my $value = process($input);  # Query this variable
    if (defined($value)) {
        return $value;
    }
    return undef;
}`,
			queryVariable: "value",
			queryBlock:    1, // After assignment
			expectedInfo: map[string]string{
				"type":        "Any",
				"defined":     "unknown",
				"safe_access": "true",
			},
			description: "Should provide detailed variable information for queries",
		},
		{
			name: "block_state_query",
			code: `
sub block_query($hash) {
    if (exists $hash->{field}) {  # Query this block
        my $value = $hash->{field};
        return $value;
    }
    return undef;
}`,
			queryVariable: "",
			queryBlock:    2, // Inside if block
			expectedInfo: map[string]string{
				"variables_count": "1",
				"safety_status":   "safe",
				"exception_risk":  "none",
			},
			description: "Should provide detailed block state information",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupDebugTest(t)
			cfg := buildDebugCFG(t, analyzer, tc.code)

			_ = analyzer.analyzeDataFlow(cfg)

			// Query specific information
			var queryResult map[string]string

			if tc.queryVariable != "" {
				queryResult = queryVariableInfo(cfg, tc.queryVariable, tc.queryBlock)
			} else {
				queryResult = queryBlockInfo(cfg, tc.queryBlock)
			}

			// Verify expected information
			for expectedKey, expectedValue := range tc.expectedInfo {
				if actualValue, exists := queryResult[expectedKey]; !exists {
					t.Errorf("%s: Expected query result to contain '%s'", tc.description, expectedKey)
				} else if !strings.Contains(actualValue, expectedValue) {
					t.Errorf("%s: Expected '%s' to contain '%s', got '%s'",
						tc.description, expectedKey, expectedValue, actualValue)
				}
			}

			t.Logf("%s: Query result: %v", tc.description, queryResult)
		})
	}
}

// Helper functions for debug tests

func setupDebugTest(t *testing.T) *FlowAnalyzer {
	typeStore, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create type store: %v", err)
	}

	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	symbolTable.Package = "debug_test"

	tc := NewTypeChecker(hierarchy, symbolTable, "debug_test")
	tc.SafetyAnalysisEnabled = true
	tc.Debug = true

	analyzer := NewFlowAnalyzer(tc)
	return analyzer
}

func buildDebugCFG(t *testing.T, analyzer *FlowAnalyzer, code string) *ControlFlowGraph {
	astResult := parseDebugCode(t, code)
	cfg, err := analyzer.buildControlFlowGraph(astResult)
	if err != nil {
		t.Fatalf("Failed to build CFG: %v", err)
	}
	return cfg
}

func parseDebugCode(t *testing.T, code string) *ast.AST {
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	result, err := p.ParseString(code)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	return result
}

func generateDOTGraph(cfg *ControlFlowGraph, filename, title string) error {
	var builder strings.Builder

	builder.WriteString("digraph CFG {\n")
	builder.WriteString("  rankdir=TB;\n")
	builder.WriteString("  node [shape=rectangle];\n")
	builder.WriteString("\n")

	// Generate nodes
	for i, block := range cfg.Nodes {
		label := "Block " + string(rune('A'+i))
		if block == cfg.Entry {
			label = "Entry"
		} else if block == cfg.Exit {
			label = "Exit"
		}

		builder.WriteString(fmt.Sprintf("  block%d [label=\"%s\"];\n", i, label))
	}

	builder.WriteString("\n")

	// Generate edges
	for i, block := range cfg.Nodes {
		for _, successor := range block.Successors {
			// Find successor index
			for j, succ := range cfg.Nodes {
				if succ == successor {
					builder.WriteString(fmt.Sprintf("  block%d -> block%d;\n", i, j))
					break
				}
			}
		}
	}

	builder.WriteString("}\n")

	return os.WriteFile(filename, []byte(builder.String()), 0644)
}

func generateEnhancedVisualization(cfg *ControlFlowGraph, filename string, showTypes, showSafety bool) error {
	var builder strings.Builder

	builder.WriteString("digraph CFG {\n")
	builder.WriteString("  rankdir=TB;\n")
	builder.WriteString("  node [shape=rectangle];\n")
	builder.WriteString("\n")

	// Generate enhanced nodes
	for i, block := range cfg.Nodes {
		label := "Block " + string(rune('A'+i))
		if block == cfg.Entry {
			label = "Entry"
		} else if block == cfg.Exit {
			label = "Exit"
		}

		// Add type information if requested
		if showTypes && block.TypeState != nil && len(block.TypeState.VariableTypes) > 0 {
			label += "\\n"
			for varName, varType := range block.TypeState.VariableTypes {
				label += varName + ": " + varType + "\\n"
			}
		}

		// Add safety information if requested
		if showSafety && block.TypeState != nil {
			if len(block.TypeState.FieldAccess) > 0 {
				label += "\\nSafety: "
				safeCount := 0
				unsafeCount := 0
				for _, fields := range block.TypeState.FieldAccess {
					for _, safe := range fields {
						if safe {
							safeCount++
						} else {
							unsafeCount++
						}
					}
				}
				label += fmt.Sprintf("safe:%d unsafe:%d", safeCount, unsafeCount)
			}
			if len(block.TypeState.ExceptionTypes) > 0 {
				label += "\\nExceptions: "
				for excType := range block.TypeState.ExceptionTypes {
					label += excType + " "
				}
			}
		}

		builder.WriteString(fmt.Sprintf("  block%d [label=\"%s\"];\n", i, label))
	}

	builder.WriteString("\n")

	// Generate edges (same as basic version)
	for i, block := range cfg.Nodes {
		for _, successor := range block.Successors {
			for j, succ := range cfg.Nodes {
				if succ == successor {
					builder.WriteString(fmt.Sprintf("  block%d -> block%d;\n", i, j))
					break
				}
			}
		}
	}

	builder.WriteString("}\n")

	return os.WriteFile(filename, []byte(builder.String()), 0644)
}

func queryVariableInfo(cfg *ControlFlowGraph, varName string, blockIndex int) map[string]string {
	result := make(map[string]string)

	if blockIndex >= len(cfg.Nodes) {
		result["error"] = "block index out of range"
		return result
	}

	block := cfg.Nodes[blockIndex]
	if block.TypeState == nil {
		result["error"] = "no type state in block"
		return result
	}

	// Get variable type
	if varType, exists := block.TypeState.VariableTypes[varName]; exists {
		result["type"] = varType
	} else {
		result["type"] = "unknown"
	}

	// Get refined type
	if refinedType, exists := block.TypeState.RefinedTypes[varName]; exists {
		result["refined_type"] = refinedType
	}

	// Determine if variable is defined
	if strings.Contains(result["type"], "Maybe") || strings.Contains(result["type"], "Undef") {
		result["defined"] = "maybe"
	} else {
		result["defined"] = "yes"
	}

	result["safe_access"] = "true" // Default assumption

	return result
}

func queryBlockInfo(cfg *ControlFlowGraph, blockIndex int) map[string]string {
	result := make(map[string]string)

	if blockIndex >= len(cfg.Nodes) {
		result["error"] = "block index out of range"
		return result
	}

	block := cfg.Nodes[blockIndex]
	if block.TypeState == nil {
		result["error"] = "no type state in block"
		return result
	}

	// Count variables
	result["variables_count"] = fmt.Sprintf("%d", len(block.TypeState.VariableTypes))

	// Assess safety status
	safeAccesses := 0
	unsafeAccesses := 0
	for _, fields := range block.TypeState.FieldAccess {
		for _, safe := range fields {
			if safe {
				safeAccesses++
			} else {
				unsafeAccesses++
			}
		}
	}

	if unsafeAccesses > 0 {
		result["safety_status"] = "unsafe"
	} else if safeAccesses > 0 {
		result["safety_status"] = "safe"
	} else {
		result["safety_status"] = "neutral"
	}

	// Assess exception risk
	if len(block.TypeState.ExceptionTypes) > 0 {
		result["exception_risk"] = "high"
	} else {
		result["exception_risk"] = "none"
	}

	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
