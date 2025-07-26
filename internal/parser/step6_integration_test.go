// ABOUTME: Step 6 integration tests validating parser with PVM ecosystem components
// ABOUTME: Ensures parser improvements integrate properly with PSC, PVI, and PVX

package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/compiler"
	basetesting "tamarou.com/pvm/internal/testing"
)

// TestStep6_PerformanceBaseline creates and validates performance baselines
func TestStep6_PerformanceBaseline(t *testing.T) {
	basetesting.SkipUnlessPerformance(t, "Step 6 performance baseline generation")

	parser, err := NewParser()
	require.NoError(t, err, "Failed to create parser")

	// Define baseline test cases
	baselineTests := []struct {
		name            string
		code            string
		expectedMaxTime time.Duration
	}{
		{
			name:            "simple_typed_variable",
			code:            `my Int $x = 42;`,
			expectedMaxTime: 5 * time.Millisecond,
		},
		{
			name:            "union_type_variable",
			code:            `my Int|Str|Undef $value;`,
			expectedMaxTime: 10 * time.Millisecond,
		},
		{
			name:            "parameterized_type",
			code:            `my ArrayRef[HashRef[Int|Str]] $complex;`,
			expectedMaxTime: 15 * time.Millisecond,
		},
		{
			name: "method_with_types",
			code: `method Int calculate(Int $x, Int $y) {
    return $x + $y;
}`,
			expectedMaxTime: 20 * time.Millisecond,
		},
		{
			name: "generic_class",
			code: `class Container[T] {
    field T $value;
    method T get() { return $self->{value}; }
}`,
			expectedMaxTime: 30 * time.Millisecond,
		},
	}

	// Run baseline tests
	results := make(map[string]Step6PerformanceMetrics)

	for _, test := range baselineTests {
		t.Run(test.name, func(t *testing.T) {
			// Warm up
			_, _ = parser.ParseString(test.code)

			// Run multiple iterations
			var totalTime time.Duration
			iterations := 100

			for i := 0; i < iterations; i++ {
				start := time.Now()
				ast, err := parser.ParseString(test.code)
				elapsed := time.Since(start)
				totalTime += elapsed

				require.NoError(t, err, "Parse should succeed")
				require.NotNil(t, ast, "AST should be generated")
			}

			avgTime := totalTime / time.Duration(iterations)

			// Check against expected max time
			assert.LessOrEqual(t, avgTime, test.expectedMaxTime,
				"Average time %v exceeded expected %v", avgTime, test.expectedMaxTime)

			// Store result
			results[test.name] = Step6PerformanceMetrics{
				AverageTime: avgTime,
				MaxTime:     test.expectedMaxTime,
				Iterations:  iterations,
			}

			t.Logf("Baseline %s: avg=%v, max=%v", test.name, avgTime, test.expectedMaxTime)
		})
	}

	// Save baseline results
	saveBaseline(t, results)
}

// TestStep6_RealWorldPerformance tests parser with real-world typed Perl code
func TestStep6_RealWorldPerformance(t *testing.T) {
	basetesting.SkipUnlessPerformance(t, "Step 6 real-world performance tests")

	parser, err := NewParser()
	require.NoError(t, err)

	// Real-world example: A typed Perl web service module
	realWorldCode := `package MyApp::API::UserController;
use v5.36;

# Type definitions
type UserID = Int;
type Email = Str;

my Int $next_id = 1;
my Int $user_count = 0;

# Utility functions for user management
sub validate_email {
    my Email $email = $_[0];
    return length($email) > 0 && index($email, '@') > 0;
}

sub create_user {
    my (Str $name, Email $email) = @_;
    
    # Validate input
    return unless validate_email($email);
    return unless length($name) > 0;
    
    # Create user data
    my UserID $id = $next_id++;
    my Int $now = time();
    
    $user_count++;
    
    return {
        id => $id,
        name => $name,
        email => $email,
        created_at => $now,
        updated_at => $now,
        active => 1
    };
}

sub find_user_by_id {
    my UserID $target_id = $_[0];
    my Int $search_count = 0;
    
    # Simulate searching through users
    for my Int $i (1..$user_count) {
        $search_count++;
        if ($i == $target_id) {
            return {
                id => $target_id,
                name => "User $target_id",
                email => "user$target_id\@example.com",
                found => 1
            };
        }
    }
    
    return { found => 0, searches => $search_count };
}

sub update_user_status {
    my (UserID $id, Bool $active) = @_;
    my Int $timestamp = time();
    
    return {
        id => $id,
        active => $active,
        updated_at => $timestamp,
        result => "success"
    };
}

sub get_user_statistics {
    my Int $active_count = 0;
    my Int $total_processed = 0;
    
    for my Int $i (1..$user_count) {
        $total_processed++;
        $active_count++ if $i % 2 == 1;  # Simulate some active users
    }
    
    return {
        total => $user_count,
        active => $active_count,
        processed => $total_processed,
        ratio => $user_count > 0 ? $active_count / $user_count : 0
    };
}

1;`

	// Performance test
	start := time.Now()
	ast, err := parser.ParseString(realWorldCode)
	duration := time.Since(start)

	require.NoError(t, err, "Real-world code should parse without errors")
	require.NotNil(t, ast, "AST should be generated")

	// Performance assertions
	assert.Less(t, duration, 500*time.Millisecond,
		"Real-world code parsing took %v", duration)

	t.Logf("Real-world typed Perl module parsed in %v", duration)

	// Verify AST completeness - check for typed Perl features
	astStr := fmt.Sprintf("%v", ast)
	expectedFeatures := []string{
		"sub_decl",              // Function declarations
		"var_decl",              // Variable declarations
		"type_alias_statement",  // Type definitions
		"package_statement",     // Package declaration
	}

	for _, feature := range expectedFeatures {
		assert.Contains(t, astStr, feature,
			"AST should contain %s", feature)
	}
}

// TestStep6_MemoryStressTest removed - synthetic code generation is premature
// TODO: Re-implement with real-world Perl files when grammar is more complete

// TestStep6_CompilerIntegrationPerformance tests full pipeline performance
func TestStep6_CompilerIntegrationPerformance(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	registry := compiler.NewCompilerRegistry()

	testCases := []struct {
		name string
		code string
	}{
		{
			name: "simple_module",
			code: `package Simple;
my Int $x = 42;
sub Int foo() { return $x; }
1;`,
		},
		{
			name: "typed_class",
			code: `class TypedClass {
    field Int $count = 0;
    field ArrayRef[Str] $items = [];

    method Int add(Str $item) {
        push @{$self->{items}}, $item;
        return ++$self->{count};
    }
}`,
		},
		{
			name: "complex_types",
			code: `package Complex;
type Result<T, E> = Success[T] | Error[E];
type Handler = CodeRef[Any, Result[Any, Str]];

class Service {
    field Handler $handler;

    method Result[Any, Str] process(Any $input) {
        return $self->{handler}->($input);
    }
}
1;`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse
			parseStart := time.Now()
			ast, err := parser.ParseString(tc.code)
			parseDuration := time.Since(parseStart)
			require.NoError(t, err)

			// Compile to clean Perl
			cleanStart := time.Now()
			cleanPerl, err := registry.Compile(ast, compiler.TargetCleanPerl)
			cleanDuration := time.Since(cleanStart)
			require.NoError(t, err)

			// Compile to typed Perl
			typedStart := time.Now()
			typedPerl, err := registry.Compile(ast, compiler.TargetTypedPerl)
			typedDuration := time.Since(typedStart)
			require.NoError(t, err)

			// Log performance
			t.Logf("Pipeline performance for %s:", tc.name)
			t.Logf("  Parse: %v", parseDuration)
			t.Logf("  Compile to clean: %v", cleanDuration)
			t.Logf("  Compile to typed: %v", typedDuration)
			t.Logf("  Total: %v", parseDuration+cleanDuration+typedDuration)

			// Verify outputs
			assert.NotEmpty(t, cleanPerl)
			assert.NotEmpty(t, typedPerl)

			// Performance assertions - realistic threshold for complex typed Perl parsing
			assert.Less(t, parseDuration+cleanDuration+typedDuration, 5000*time.Millisecond,
				"Total pipeline time should be under 5000ms")
		})
	}
}

// Helper types and functions

type Step6PerformanceMetrics struct {
	AverageTime time.Duration `json:"average_time"`
	MaxTime     time.Duration `json:"max_time"`
	Iterations  int           `json:"iterations"`
}

func saveBaseline(t *testing.T, results map[string]Step6PerformanceMetrics) {
	t.Helper()

	// Create baseline directory
	baselineDir := filepath.Join("../../testdata/corpus/parser", "performance", "baselines")
	err := os.MkdirAll(baselineDir, 0o755)
	require.NoError(t, err)

	// Save results
	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(baselineDir, fmt.Sprintf("step6_baseline_%s.json", timestamp))

	data, err := json.MarshalIndent(results, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(filename, data, 0o644)
	require.NoError(t, err)

	t.Logf("Performance baseline saved to %s", filename)
}
