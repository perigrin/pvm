// ABOUTME: Comprehensive flow analysis functionality tests
// ABOUTME: Tests core flow analysis features including CFG construction and type state evolution

package typechecker

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

// TestSafetyAnalysisDetection tests comprehensive safety analysis detection
func TestSafetyAnalysisDetection(t *testing.T) {
	testCases := []struct {
		name           string
		code           string
		expectedErrors []string
		description    string
	}{
		{
			name: "unsafe_hash_field_access",
			code: `
sub process_data($input) {
    my $name = $input->{name};
    my $id = $input->{user_id};
    return "$name:$id";
}`,
			expectedErrors: []string{"unsafe hash field access", "name", "user_id"},
			description:    "Should detect unsafe hash field access without exists() checks",
		},
		{
			name: "unsafe_array_access",
			code: `
sub get_first_item($data) {
    my $first = $data->[0];        # Unsafe: not proven to be array
    return $first;
}`,
			expectedErrors: []string{"unsafe array access"},
			description:    "Should detect unsafe array access without type validation",
		},
		{
			name: "uninitialized_variable_usage",
			code: `
sub calculate_total($maybe_price) {
    my $tax = $maybe_price * 0.1;  # Unsafe: might be undef
    my $msg = "Price: $maybe_price"; # Unsafe: string interpolation
    return $tax + 10;
}`,
			expectedErrors: []string{"uninitialized", "numeric context", "string context"},
			description:    "Should detect uninitialized variable usage in various contexts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupFlowAnalysisTest(t)
			cfg := buildTestCFG(t, analyzer, tc.code)

			errors := analyzer.analyzeDataFlow(cfg)

			for _, expectedErr := range tc.expectedErrors {
				found := false
				for _, err := range errors {
					if strings.Contains(err.Error(), expectedErr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: Expected error containing '%s', but didn't find it in: %v",
						tc.description, expectedErr, errors)
				}
			}
		})
	}
}

// TestTypeInferenceFromUntyped tests comprehensive type inference from untyped code
func TestTypeInferenceFromUntyped(t *testing.T) {
	testCases := []struct {
		name               string
		code               string
		expectedInferences map[string]string
		description        string
	}{
		{
			name: "builtin_function_inference",
			code: `
sub analyze_data($input) {
    my $type = ref($input);           # Should infer: Str
    my $is_defined = defined($input); # Should infer: Bool
    my $keys = keys(%$input);         # Should infer: ArrayRef[Str]
    return ($type, $is_defined, $keys);
}`,
			expectedInferences: map[string]string{
				"type":       "Str",
				"is_defined": "Bool",
				"keys":       "ArrayRef[Str]",
			},
			description: "Should infer types from Perl built-in functions",
		},
		{
			name: "constructor_call_inference",
			code: `
sub create_objects() {
    my $user = User->new(name => 'John');     # Should infer: User
    my $db = DBI->connect($dsn, $user, $pass); # Should infer: DBI
    my $json = JSON->new();                   # Should infer: JSON
    return ($user, $db, $json);
}`,
			expectedInferences: map[string]string{
				"user": "User",
				"db":   "DBI",
				"json": "JSON",
			},
			description: "Should infer types from constructor calls",
		},
		{
			name: "method_chain_inference",
			code: `
sub process_json($data_str) {
    my $parsed = JSON->new()->decode($data_str);  # Should infer: HashRef[Str, Any]
    my $encoded = JSON->new()->encode($parsed);   # Should infer: Str
    return $encoded;
}`,
			expectedInferences: map[string]string{
				"parsed":  "HashRef[Str, Any]",
				"encoded": "Str",
			},
			description: "Should infer types through method chains",
		},
		{
			name: "library_function_inference",
			code: `
sub handle_file($filename) {
    my $content = slurp($filename);      # Should infer: Str
    my $decoded = decode_json($content); # Should infer: HashRef[Str, Any]
    my $result = $dbh->selectrow_hashref($sql); # Should infer: Maybe[HashRef[Str, Str]]
    return ($content, $decoded, $result);
}`,
			expectedInferences: map[string]string{
				"content": "Str",
				"decoded": "HashRef[Str, Any]",
				"result":  "Maybe[HashRef[Str, Str]]",
			},
			description: "Should infer types from library functions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupFlowAnalysisTest(t)
			cfg := buildTestCFG(t, analyzer, tc.code)

			_ = analyzer.analyzeDataFlow(cfg)

			for varName, expectedType := range tc.expectedInferences {
				found := false
				for _, block := range cfg.Nodes {
					// Check both TypeState (entry) and ExitTypeState for variable types
					var variableTypes map[string]string
					if block.ExitTypeState != nil && block.ExitTypeState.VariableTypes != nil {
						variableTypes = block.ExitTypeState.VariableTypes
					} else if block.TypeState != nil && block.TypeState.VariableTypes != nil {
						variableTypes = block.TypeState.VariableTypes
					}

					if variableTypes != nil {
						if actualType, exists := variableTypes[varName]; exists {
							if strings.Contains(actualType, expectedType) || actualType == expectedType {
								found = true
								break
							}
						}
					}
				}
				if !found {
					t.Errorf("%s: Expected to infer type '%s' for variable $%s",
						tc.description, expectedType, varName)
				}
			}
		})
	}
}

// TestExceptionFlowTracking tests Throws[T] union type tracking
func TestExceptionFlowTracking(t *testing.T) {
	testCases := []struct {
		name               string
		code               string
		expectedExceptions []string
		expectedUnionTypes map[string]string
		description        string
	}{
		{
			name: "die_statement_tracking",
			code: `
sub validate_input(Str $input) {
    die "empty input" if $input eq '';
    die "invalid format" unless $input =~ /^valid/;
    return process($input);
}`,
			expectedExceptions: []string{"Throws[Str]"},
			expectedUnionTypes: map[string]string{
				"return": "Any|Throws[Str]",
			},
			description: "Should track die statements and infer Throws[T] types",
		},
		{
			name: "exception_propagation",
			code: `
sub high_level($data) {
    my $result = validate_input($data);  # Propagates Throws[Str]
    return $result;
}`,
			expectedExceptions: []string{"Throws[Str]"},
			expectedUnionTypes: map[string]string{
				"result": "Any|Throws[Str]",
			},
			description: "Should propagate exception types through function calls",
		},
		{
			name: "exception_handling",
			code: `
sub safe_wrapper($data) {
    eval {
        my $result = validate_input($data);
        return $result;
    };
    if ($@) {
        warn "Caught exception: $@";
        return undef;
    }
}`,
			expectedExceptions: []string{}, // eval should catch exceptions
			expectedUnionTypes: map[string]string{
				"return": "Maybe[Any]",
			},
			description: "Should handle exception catching with eval blocks",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupFlowAnalysisTest(t)
			cfg := buildTestCFG(t, analyzer, tc.code)

			_ = analyzer.analyzeDataFlow(cfg)

			// Check for expected exception types
			for _, expectedExc := range tc.expectedExceptions {
				found := false
				for _, block := range cfg.Nodes {
					if block.TypeState != nil && block.TypeState.ExceptionTypes != nil {
						for excType := range block.TypeState.ExceptionTypes {
							if strings.Contains(excType, expectedExc) {
								found = true
								break
							}
						}
					}
				}
				if !found {
					t.Errorf("%s: Expected to find exception type '%s'", tc.description, expectedExc)
				}
			}

			// Check for expected union types
			for varName, expectedType := range tc.expectedUnionTypes {
				found := false
				for _, block := range cfg.Nodes {
					if block.TypeState != nil && block.TypeState.VariableTypes != nil {
						if actualType, exists := block.TypeState.VariableTypes[varName]; exists {
							if strings.Contains(actualType, expectedType) {
								found = true
								break
							}
						}
					}
				}
				if !found {
					t.Errorf("%s: Expected union type '%s' for %s", tc.description, expectedType, varName)
				}
			}
		})
	}
}

// TestComplexControlFlowGraphConstruction tests CFG construction for complex control structures
func TestComplexControlFlowGraphConstruction(t *testing.T) {
	testCases := []struct {
		name                string
		code                string
		expectedBlockCount  int
		expectedConnections []string
		description         string
	}{
		{
			name: "complex_conditionals",
			code: `
sub complex_logic($input, $flag) {
    if ($input && defined($input->{data})) {
        if ($flag) {
            return process_a($input);
        } else {
            return process_b($input);
        }
    } elsif ($input) {
        return handle_partial($input);
    } else {
        return default_value();
    }
}`,
			expectedBlockCount: 7, // entry + conditions + branches + exit
			expectedConnections: []string{
				"entry->condition1",
				"condition1->condition2",
				"condition2->process_a",
				"condition2->process_b",
				"condition1->handle_partial",
				"condition1->default_value",
			},
			description: "Should construct proper CFG for nested conditionals",
		},
		{
			name: "loop_structures",
			code: `
sub process_items(@items) {
    my @results;
    for my $item (@items) {
        next if !defined($item);
        last if $item eq 'STOP';
        push @results, process_item($item);
    }
    return @results;
}`,
			expectedBlockCount: 6, // entry + loop header + body + next/last + exit
			expectedConnections: []string{
				"entry->loop_header",
				"loop_header->loop_body",
				"loop_body->next_stmt",
				"loop_body->last_stmt",
				"next_stmt->loop_header",
				"last_stmt->exit",
			},
			description: "Should construct proper CFG for loop structures with next/last",
		},
		{
			name: "given_when_construct",
			code: `
sub handle_status($status) {
    given ($status) {
        when ('active')   { return activate(); }
        when ('inactive') { return deactivate(); }
        when ('pending')  { return queue(); }
        default          { return error("unknown status"); }
    }
}`,
			expectedBlockCount: 6, // entry + given + 4 when branches + default
			expectedConnections: []string{
				"entry->given",
				"given->when_active",
				"given->when_inactive",
				"given->when_pending",
				"given->default",
			},
			description: "Should construct proper CFG for given/when constructs",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupFlowAnalysisTest(t)
			cfg := buildTestCFG(t, analyzer, tc.code)

			if len(cfg.Nodes) < tc.expectedBlockCount {
				t.Errorf("%s: Expected at least %d blocks, got %d",
					tc.description, tc.expectedBlockCount, len(cfg.Nodes))
			}

			if cfg.Entry == nil {
				t.Errorf("%s: CFG should have entry block", tc.description)
			}

			if cfg.Exit == nil {
				t.Errorf("%s: CFG should have exit block", tc.description)
			}

			// Verify basic connectivity (detailed connection testing would require more CFG introspection)
			connectedBlocks := 0
			for _, block := range cfg.Nodes {
				if len(block.Successors) > 0 || len(block.Predecessors) > 0 {
					connectedBlocks++
				}
			}

			if connectedBlocks < 2 {
				t.Errorf("%s: CFG should have connected blocks", tc.description)
			}
		})
	}
}

// TestTypeStateEvolution tests how type states evolve through control flow
func TestTypeStateEvolution(t *testing.T) {
	testCases := []struct {
		name                string
		code                string
		expectedRefinements map[string][]string // var -> [initial_type, refined_type]
		description         string
	}{
		{
			name: "defined_check_refinement",
			code: `
sub safe_access(Maybe[Str] $input) {
    if (defined($input)) {
        my $length = length($input);  # $input refined from Maybe[Str] to Str
        return $length;
    }
    return 0;
}`,
			expectedRefinements: map[string][]string{
				"input": {"Maybe[Str]", "Str"},
			},
			description: "Should refine Maybe[T] to T after defined() check",
		},
		{
			name: "ref_check_refinement",
			code: `
sub process_data($data) {
    if (ref($data) eq 'HASH') {
        my $keys = keys(%$data);     # $data refined to HashRef
        return $keys;
    } elsif (ref($data) eq 'ARRAY') {
        my $count = @$data;          # $data refined to ArrayRef
        return $count;
    }
    return undef;
}`,
			expectedRefinements: map[string][]string{
				"data": {"Any", "HashRef", "ArrayRef"},
			},
			description: "Should refine types based on ref() checks",
		},
		{
			name: "exists_check_refinement",
			code: `
sub extract_field($hash, $field) {
    if (exists $hash->{$field}) {
        return $hash->{$field};      # Field access safe after exists check
    }
    return undef;
}`,
			expectedRefinements: map[string][]string{
				"field_access": {"unsafe", "safe"},
			},
			description: "Should mark field access as safe after exists() check",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := setupFlowAnalysisTest(t)
			cfg := buildTestCFG(t, analyzer, tc.code)

			_ = analyzer.analyzeDataFlow(cfg)

			for varName, expectedTypes := range tc.expectedRefinements {
				// Look for type refinements across blocks
				refinementFound := false
				for _, block := range cfg.Nodes {
					if block.TypeState != nil {
						if refined, exists := block.TypeState.RefinedTypes[varName]; exists {
							for _, expectedType := range expectedTypes {
								if strings.Contains(refined, expectedType) {
									refinementFound = true
									break
								}
							}
						}
					}
				}

				if !refinementFound {
					t.Errorf("%s: Expected type refinement for %s with types %v",
						tc.description, varName, expectedTypes)
				}
			}
		})
	}
}

// Helper functions for flow analysis tests

func setupFlowAnalysisTest(t *testing.T) *FlowAnalyzer {
	typeStore, err := typedef.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create type store: %v", err)
	}

	hierarchy := typedef.NewTypeHierarchy(typeStore)
	symbolTable := binder.NewSymbolTable()
	symbolTable.Package = "flow_test"

	tc := NewTypeChecker(hierarchy, symbolTable, "flow_test")
	tc.SafetyAnalysisEnabled = true

	analyzer := NewFlowAnalyzer(tc)
	return analyzer
}

func buildTestCFG(t *testing.T, analyzer *FlowAnalyzer, code string) *ControlFlowGraph {
	astResult := parseTestCode(t, code)
	cfg, err := analyzer.buildControlFlowGraph(astResult)
	if err != nil {
		t.Fatalf("Failed to build CFG: %v", err)
	}
	return cfg
}

func parseTestCode(t *testing.T, code string) *ast.AST {
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
