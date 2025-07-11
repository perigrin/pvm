// ABOUTME: Comprehensive integration tests for unified compiler architecture
// ABOUTME: Tests all PSC commands, parser integration, edge cases, and stress scenarios

package compiler

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/parser"
	basetesting "tamarou.com/pvm/internal/testing"
)

// TestPSCCommandsIntegration tests all PSC commands with unified compiler
func TestPSCCommandsIntegration(t *testing.T) {
	// Test cases for different PSC command scenarios
	testCases := []struct {
		name        string
		code        string
		target      Target
		expectError bool
	}{
		{
			name: "PSC Strip Command",
			code: `my Int $count = 42;
field Str $name = "test";
print "Count: $count, Name: $name\n";`,
			target:      TargetCleanPerl,
			expectError: false,
		},
		{
			name: "PSC Run Command",
			code: `my Int $x = 10;
my Int $y = 20;
my Int $result = $x + $y;
print "Result: $result\n";`,
			target:      TargetCleanPerl,
			expectError: false,
		},
		{
			name: "PSC Check Command (Typed)",
			code: `method validate(Int $input) -> Bool {
    return $input > 0;
}
my Bool $valid = validate(42);`,
			target:      TargetTypedPerl,
			expectError: false,
		},
		{
			name: "Complex Type Transformations",
			code: `my ArrayRef[HashRef[Str]] $complex = [{name => "alice", role => "admin"}];
my Union[Int, Str, Bool] $flexible = "string_value";
my Maybe[Object] $optional = undef;`,
			target:      TargetCleanPerl,
			expectError: false,
		},
	}

	registry := NewCompilerRegistry()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary CST-based AST
			cstAST, err := NewCSTBasedAST("test.pl", tc.code)
			if err != nil {
				t.Fatalf("Failed to create CST AST: %v", err)
			}

			// Compile using registry
			result, err := registry.Compile(cstAST, tc.target)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but compilation succeeded")
				}
			} else {
				if err != nil {
					t.Fatalf("Compilation failed: %v", err)
				}

				// Validate output quality
				validateCompiledOutput(t, tc.code, result, tc.target)
			}
		})
	}
}

// TestParserIntegration validates integration with parser package
func TestParserIntegration(t *testing.T) {
	testCodes := []string{
		`my Int $simple = 42;`,
		`field Str $field_var = "test";`,
		`my ArrayRef[Int] $array = [1, 2, 3];`,
		`method process(Str $input) -> Int { return length($input); }`,
		`my $assertion = get_value() as Str;`,
	}

	// Create parser instance
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	registry := NewCompilerRegistry()

	for i, code := range testCodes {
		t.Run(fmt.Sprintf("ParserIntegration_%d", i), func(t *testing.T) {
			// Create temporary file
			tempFile, err := os.CreateTemp("", "test_*.pl")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			// Write code to file
			if _, err := tempFile.WriteString(code); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tempFile.Close()

			// Parse file
			ast, err := p.ParseFile(tempFile.Name())
			if err != nil {
				t.Fatalf("Failed to parse file: %v", err)
			}

			// Compile with unified compiler
			result, err := registry.Compile(ast, TargetCleanPerl)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			// Validate result
			if result == "" {
				t.Error("Compilation produced empty result")
			}

			// Should contain version pragma
			if !strings.Contains(result, "use v5.36;") {
				t.Error("Clean Perl output should contain version pragma")
			}
		})
	}
}

// TestEdgeCasesAndErrorHandling tests edge cases and error scenarios
func TestEdgeCasesAndErrorHandling(t *testing.T) {
	testCases := []struct {
		name        string
		code        string
		expectError bool
		errorType   string
	}{
		{
			name:        "Empty Code",
			code:        "",
			expectError: false, // Empty code should compile to empty output
		},
		{
			name:        "Only Comments",
			code:        "# This is just a comment\n# Another comment",
			expectError: false,
		},
		{
			name:        "Syntax Error",
			code:        "my Int $ = ;", // Invalid variable name
			expectError: false,          // Tree-sitter may recover from errors
		},
		{
			name:        "Very Long Variable Names",
			code:        fmt.Sprintf("my Int $%s = 42;", strings.Repeat("very_long_name", 100)),
			expectError: false,
		},
		{
			name: "Unicode Content",
			code: `my Str $message = "Hello, 世界! 🌍";
print $message;`,
			expectError: false,
		},
		{
			name:        "Deeply Nested Types",
			code:        `my ArrayRef[HashRef[ArrayRef[HashRef[Str]]]] $deep = [{}];`,
			expectError: false,
		},
		{
			name: "Mixed Type Annotations",
			code: `my Int $typed = 42;
my $untyped = "string";
my Str $another_typed = "test";`,
			expectError: false,
		},
	}

	compiler := NewCleanPerlCompilerUnified()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := compiler.CompileString(tc.code)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but compilation succeeded")
				}
			} else {
				if err != nil {
					t.Fatalf("Compilation failed: %v", err)
				}

				// Basic validation for non-error cases
				if tc.code != "" && result == "" {
					t.Error("Non-empty code produced empty result")
				}
			}
		})
	}
}

// TestBackwardCompatibility verifies backward compatibility
func TestBackwardCompatibility(t *testing.T) {
	// Test that existing workflow patterns still work

	// Test legacy AST compatibility
	t.Run("LegacyASTCompatibility", func(t *testing.T) {
		code := `my Int $legacy = 42;`

		// Create a CST-based AST (new format)
		cstAST, err := NewCSTBasedAST("test.pl", code)
		if err != nil {
			t.Fatalf("Failed to create CST AST: %v", err)
		}

		// Should work with unified compiler
		compiler := NewCleanPerlCompilerUnified()
		result, err := compiler.Compile(cstAST)
		if err != nil {
			t.Fatalf("Compilation failed: %v", err)
		}

		if !strings.Contains(result, "my $legacy = 42;") {
			t.Errorf("Expected output to contain 'my $legacy = 42;', got: %s", result)
		}
	})

	// Test compiler registry compatibility
	t.Run("RegistryCompatibility", func(t *testing.T) {
		registry := NewCompilerRegistry()

		// Should have both clean and typed compilers
		cleanCompiler, exists := registry.GetCompiler(TargetCleanPerl)
		if !exists {
			t.Error("Registry should have clean Perl compiler")
		}

		typedCompiler, exists := registry.GetCompiler(TargetTypedPerl)
		if !exists {
			t.Error("Registry should have typed Perl compiler")
		}

		// Both should be unified compilers
		if cleanCompiler.Target() != TargetCleanPerl {
			t.Error("Clean compiler has wrong target")
		}

		if typedCompiler.Target() != TargetTypedPerl {
			t.Error("Typed compiler has wrong target")
		}
	})
}

// TestStressScenarios tests with large, complex codebases
func TestStressScenarios(t *testing.T) {
	basetesting.SkipUnlessStress(t, "compiler stress test scenarios")

	// Note: LargeCodebase test removed - synthetic code generation is premature
	// TODO: Re-implement with real-world Perl files when ready

	t.Run("ConcurrentCompilation", func(t *testing.T) {
		// Load real test corpus files to test cache effectiveness
		corpusFiles := []string{
			"../../testdata/corpus/tree-sitter/highlight/variables.pm",
			"../../testdata/corpus/tree-sitter/highlight/expressions.pm",
			"../../testdata/corpus/tree-sitter/highlight/operators.pm",
		}

		var testCodes []string
		for _, file := range corpusFiles {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read corpus file %s: %v", file, err)
			}
			testCodes = append(testCodes, string(content))
		}

		compiler := NewCachingCleanPerlCompiler(100)

		// Test concurrent compilation with repeated corpus files (should generate cache hits)
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()

				for j := 0; j < 100; j++ {
					// Use modulo to cycle through corpus files, creating cache hits
					testCode := testCodes[j%len(testCodes)]
					_, err := compiler.CompileString(testCode)
					if err != nil {
						t.Errorf("Concurrent compilation failed: %v", err)
						return
					}
				}
			}(i)
		}

		// Wait for all goroutines
		timeout := time.After(30 * time.Second)
		completed := 0
		for completed < 10 {
			select {
			case <-done:
				completed++
			case <-timeout:
				t.Fatal("Concurrent compilation test timed out")
			}
		}

		// Check cache effectiveness
		stats := compiler.GetCacheStats()
		if stats.HitRatio < 0.5 { // Should have at least 50% cache hits
			t.Errorf("Poor cache performance in concurrent test: %.2f%% hit ratio", stats.HitRatio*100)
		}

		t.Logf("Concurrent compilation completed with %.2f%% cache hit ratio", stats.HitRatio*100)
	})

	t.Run("MemoryStressTest", func(t *testing.T) {
		compiler := NewCachingCleanPerlCompiler(1000)

		// Generate many unique code samples to stress memory usage
		for i := 0; i < 5000; i++ {
			code := fmt.Sprintf(`my Int $stress%d = %d;
my Str $name%d = "test%d";
field ArrayRef[Int] $array%d = [%d, %d];`, i, i, i, i, i, i, i+1)

			_, err := compiler.CompileString(code)
			if err != nil {
				t.Fatalf("Memory stress test failed at iteration %d: %v", i, err)
			}

			// Check every 1000 iterations
			if i%1000 == 0 {
				stats := compiler.GetCacheStats()
				t.Logf("Iteration %d: Cache size %d, hit ratio %.2f%%",
					i, stats.CleanSize+stats.TypedSize, stats.HitRatio*100)
			}
		}
	})
}

// TestComplexRealWorldScenarios tests with real-world-like code patterns
func TestComplexRealWorldScenarios(t *testing.T) {
	scenarios := map[string]string{
		"WebService": `package WebService;
use v5.36;

field Str $base_url;
field HashRef[Str] $headers;

method new(Str $url, HashRef[Str] $hdrs) -> Self {
    $base_url = $url;
    $headers = $hdrs;
    return bless {}, __PACKAGE__;
}

method get(Str $endpoint) -> Str {
    my Str $url = "$base_url/$endpoint";
    my HashRef[Any] $response = http_get($url, $headers);
    return $response->{body} as Str;
}

method post(Str $endpoint, HashRef[Any] $data) -> HashRef[Any] {
    my Str $url = "$base_url/$endpoint";
    my Str $json = encode_json($data);
    return http_post($url, $json, $headers);
}`,

		"DataProcessor": `package DataProcessor;

field ArrayRef[HashRef[Str]] $records;
field HashRef[Int] $statistics;

method process_data(ArrayRef[HashRef[Any]] $input) -> ArrayRef[HashRef[Str]] {
    my ArrayRef[HashRef[Str]] $processed = [];

    for my HashRef[Any] $record (@$input) {
        my HashRef[Str] $clean_record = {};

        for my Str $key (keys %$record) {
            my Any $value = $record->{$key};
            $clean_record->{$key} = normalize($value as Str);
        }

        push @$processed, $clean_record;
    }

    return $processed;
}

method calculate_stats() -> HashRef[Int] {
    my Int $total = scalar(@$records);
    my Int $valid = 0;

    for my HashRef[Str] $record (@$records) {
        $valid++ if validate_record($record);
    }

    return {
        total => $total,
        valid => $valid,
        invalid => $total - $valid,
    };
}`,

		"ConfigManager": `package ConfigManager;

field HashRef[Union[Str, Int, Bool]] $config;
field Str $config_file;

method load_config(Str $file) -> Bool {
    $config_file = $file;

    my Str $content = read_file($file);
    my HashRef[Any] $raw_config = decode_json($content);

    $config = validate_config($raw_config as HashRef[Union[Str, Int, Bool]]);

    return defined($config);
}

method get(Str $key) -> Union[Str, Int, Bool] {
    return $config->{$key} // undef;
}

method set(Str $key, Union[Str, Int, Bool] $value) -> Void {
    $config->{$key} = $value;
    save_config();
}`,
	}

	registry := NewCompilerRegistry()

	for name, code := range scenarios {
		t.Run(name, func(t *testing.T) {
			// Test clean Perl compilation
			cstAST, err := NewCSTBasedAST(fmt.Sprintf("%s.pl", name), code)
			if err != nil {
				t.Fatalf("Failed to create CST AST: %v", err)
			}

			cleanResult, err := registry.Compile(cstAST, TargetCleanPerl)
			if err != nil {
				t.Fatalf("Clean Perl compilation failed: %v", err)
			}

			typedResult, err := registry.Compile(cstAST, TargetTypedPerl)
			if err != nil {
				t.Fatalf("Typed Perl compilation failed: %v", err)
			}

			// Validate clean Perl output
			validateCleanPerlOutput(t, cleanResult, name)

			// Validate typed Perl output (should preserve types)
			validateTypedPerlOutput(t, typedResult, name)

			t.Logf("Successfully compiled %s scenario", name)
		})
	}
}

// validateCompiledOutput validates the quality of compiled output
func validateCompiledOutput(t *testing.T, input, output string, target Target) {
	if output == "" {
		t.Error("Compilation produced empty output")
		return
	}

	switch target {
	case TargetCleanPerl:
		// Clean Perl should not contain type annotations
		if strings.Contains(output, " Int ") || strings.Contains(output, " Str ") {
			t.Error("Clean Perl output should not contain type annotations")
		}

		// Should contain version pragma
		if !strings.Contains(output, "use v5.36;") {
			t.Error("Clean Perl output should contain version pragma")
		}

	case TargetTypedPerl:
		// Typed Perl should preserve type annotations
		if strings.Contains(input, " Int ") && !strings.Contains(output, " Int ") {
			t.Error("Typed Perl output should preserve Int type annotations")
		}
	}
}

// validateCleanPerlOutput validates clean Perl compilation results
func validateCleanPerlOutput(t *testing.T, output, scenario string) {
	// Should not contain type annotations
	typeKeywords := []string{" Int ", " Str ", " Bool ", " ArrayRef[", " HashRef[", " Union["}
	for _, keyword := range typeKeywords {
		if strings.Contains(output, keyword) {
			t.Errorf("Clean Perl output for %s contains type annotation: %s", scenario, keyword)
		}
	}

	// Should contain version pragma for compatibility
	if !strings.Contains(output, "use v5.36;") {
		t.Errorf("Clean Perl output for %s should contain version pragma", scenario)
	}

	// Should preserve variable names based on scenario
	var hasExpectedVars bool
	switch scenario {
	case "WebService":
		hasExpectedVars = strings.Contains(output, "$base_url") || strings.Contains(output, "$headers")
	case "DataProcessor":
		hasExpectedVars = strings.Contains(output, "$records") || strings.Contains(output, "$statistics")
	case "ConfigManager":
		hasExpectedVars = strings.Contains(output, "$config") || strings.Contains(output, "$config_file")
	default:
		hasExpectedVars = true // Allow for new scenarios
	}

	if !hasExpectedVars && strings.Contains(output, "field ") {
		t.Errorf("Clean Perl output for %s should have variable names, not just field keyword", scenario)
	}
}

// validateTypedPerlOutput validates typed Perl compilation results
func validateTypedPerlOutput(t *testing.T, output, scenario string) {
	// Should preserve type annotations
	if strings.Contains(output, "field Str") || strings.Contains(output, "HashRef[") {
		// Good - type annotations preserved
	} else {
		t.Errorf("Typed Perl output for %s should preserve type annotations", scenario)
	}
}

// TestErrorRecoveryAndDiagnostics tests error handling and diagnostics
func TestErrorRecoveryAndDiagnostics(t *testing.T) {
	errorCases := []struct {
		name     string
		code     string
		checkMsg func(error) bool
	}{
		{
			name: "Invalid AST",
			code: "", // Will create invalid AST
			checkMsg: func(err error) bool {
				return strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "nil")
			},
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test with nil AST
			compiler := NewCleanPerlCompilerUnified()
			_, err := compiler.Compile(nil)

			if err == nil {
				t.Error("Expected error with nil AST")
				return
			}

			if !tc.checkMsg(err) {
				t.Errorf("Error message doesn't match expected pattern: %v", err)
			}
		})
	}
}
