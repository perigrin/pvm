// ABOUTME: Integration tests for LSP language service functionality with real-world scenarios
// ABOUTME: Tests end-to-end LSP features including caching, performance, and error handling

package ls

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestFullWorkflow tests a complete LSP workflow with a realistic Perl file
func TestFullWorkflow(t *testing.T) {
	ls, err := NewLanguageService()
	if err != nil {
		t.Fatalf("Failed to create language service: %v", err)
	}

	// Create a realistic Perl file
	content := `#!/usr/bin/perl
use strict;
use warnings;

package Calculator;

# A simple calculator class
my $version = "1.0";

sub new {
    my ($class) = @_;
    my $self = {
        result => 0,
        history => [],
    };
    return bless $self, $class;
}

sub add {
    my ($self, $value) = @_;
    $self->{result} += $value;
    push @{$self->{history}}, "add $value";
    return $self->{result};
}

sub multiply {
    my ($self, $value) = @_;
    $self->{result} *= $value;
    push @{$self->{history}}, "multiply $value";
    return $self->{result};
}

sub get_result {
    my ($self) = @_;
    return $self->{result};
}

sub get_history {
    my ($self) = @_;
    return @{$self->{history}};
}

# Usage example
my $calc = Calculator->new();
$calc->add(5);
$calc->multiply(3);
print "Result: " . $calc->get_result() . "\n";
`

	uri := "file:///calculator.pl"

	// Test document update
	err = ls.UpdateDocument(uri, content, 1)
	if err != nil {
		t.Fatalf("UpdateDocument failed: %v", err)
	}

	// Test hover on various elements
	testCases := []struct {
		name        string
		line        int
		char        int
		expectHover bool
		description string
	}{
		{"Variable", 7, 5, true, "Hover on $version variable"},
		{"Subroutine", 9, 5, true, "Hover on 'new' subroutine"},
		{"Method", 18, 5, true, "Hover on 'add' method"},
		{"Parameter", 19, 10, true, "Hover on $self parameter"},
		{"Comment", 6, 5, false, "Hover on comment (should not work)"},
		{"Whitespace", 8, 0, false, "Hover on whitespace"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pos := Position{Line: tc.line, Character: tc.char}
			hover, err := ls.GetHover(uri, pos)

			if err != nil {
				t.Fatalf("GetHover failed: %v", err)
			}

			if tc.expectHover && hover == nil {
				t.Errorf("Expected hover information for %s", tc.description)
			} else if !tc.expectHover && hover != nil {
				t.Errorf("Unexpected hover information for %s", tc.description)
			}

			if hover != nil {
				t.Logf("Hover for %s: %s", tc.description, hover.Contents)
			}
		})
	}

	// Test completions at various positions
	completionTestCases := []struct {
		name        string
		line        int
		char        int
		expectItems bool
		minItems    int
		description string
	}{
		{"AfterMy", 7, 8, true, 1, "Completion after 'my'"},
		{"MethodCall", 40, 6, true, 3, "Completion for method call"},
		{"PackageName", 4, 10, true, 1, "Completion in package context"},
		{"EmptyLine", 45, 0, true, 5, "Completion on empty line"},
	}

	for _, tc := range completionTestCases {
		t.Run("Completion_"+tc.name, func(t *testing.T) {
			pos := Position{Line: tc.line, Character: tc.char}
			completions, err := ls.GetCompletions(uri, pos)

			if err != nil {
				t.Fatalf("GetCompletions failed: %v", err)
			}

			if tc.expectItems && len(completions) < tc.minItems {
				t.Errorf("Expected at least %d completion items for %s, got %d",
					tc.minItems, tc.description, len(completions))
			}

			t.Logf("Completions for %s: %d items", tc.description, len(completions))
			for i, item := range completions {
				if i < 5 { // Log first 5 items
					t.Logf("  - %s (%s)", item.Label, item.Detail)
				}
			}
		})
	}

	// Test definition lookup
	definitionTestCases := []struct {
		name        string
		line        int
		char        int
		expectDef   bool
		description string
	}{
		{"Variable", 40, 5, true, "Definition of $calc variable"},
		{"Method", 41, 7, true, "Definition of add method"},
		{"Subroutine", 9, 5, true, "Definition of new subroutine"},
	}

	for _, tc := range definitionTestCases {
		t.Run("Definition_"+tc.name, func(t *testing.T) {
			pos := Position{Line: tc.line, Character: tc.char}
			definition, err := ls.GetDefinition(uri, pos)

			if err != nil {
				t.Fatalf("GetDefinition failed: %v", err)
			}

			if tc.expectDef && definition == nil {
				t.Errorf("Expected definition for %s", tc.description)
			} else if !tc.expectDef && definition != nil {
				t.Errorf("Unexpected definition for %s", tc.description)
			}

			if definition != nil {
				t.Logf("Definition for %s: %s line %d",
					tc.description, definition.Location.URI, definition.Location.Range.Start.Line)
			}
		})
	}

	// Test document symbols
	symbols, err := ls.GetDocumentSymbols(uri)
	if err != nil {
		t.Fatalf("GetDocumentSymbols failed: %v", err)
	}

	if len(symbols) == 0 {
		t.Error("Expected document symbols, got none")
	}

	t.Logf("Document symbols: %d found", len(symbols))
	for i, symbol := range symbols {
		if i < 10 { // Log first 10 symbols
			t.Logf("  - %s (%s)", symbol.Name, symbol.Kind)
		}
	}

	// Test workspace symbol search
	workspaceSymbols, err := ls.GetWorkspaceSymbols("calc")
	if err != nil {
		t.Fatalf("GetWorkspaceSymbols failed: %v", err)
	}

	t.Logf("Workspace symbols matching 'calc': %d found", len(workspaceSymbols))
}

// TestMultipleDocuments tests LSP functionality across multiple documents
func TestMultipleDocuments(t *testing.T) {
	ls, err := NewLanguageService()
	if err != nil {
		t.Fatalf("Failed to create language service: %v", err)
	}

	// Create multiple related Perl files
	documents := map[string]string{
		"file:///lib/Utils.pl": `
package Utils;
use strict;
use warnings;

sub format_number {
    my ($num) = @_;
    return sprintf("%.2f", $num);
}

sub debug_print {
    my ($msg) = @_;
    print STDERR "[DEBUG] $msg\n";
}

1;
`,
		"file:///main.pl": `
#!/usr/bin/perl
use strict;
use warnings;
use lib 'lib';
use Utils;

my $value = 42.12345;
my $formatted = Utils::format_number($value);
Utils::debug_print("Formatted value: $formatted");
print "$formatted\n";
`,
		"file:///test.pl": `
#!/usr/bin/perl
use strict;
use warnings;
use Test::More;

use lib 'lib';
use Utils;

# Test Utils functions
is(Utils::format_number(42.12345), "42.12", "Number formatting");
ok(1, "Basic test");

done_testing();
`,
	}

	// Update all documents
	for uri, content := range documents {
		err = ls.UpdateDocument(uri, content, 1)
		if err != nil {
			t.Fatalf("UpdateDocument failed for %s: %v", uri, err)
		}
	}

	// Test workspace-wide symbol search
	searchTerms := []string{"format_number", "debug_print", "value", "test"}

	for _, term := range searchTerms {
		symbols, err := ls.GetWorkspaceSymbols(term)
		if err != nil {
			t.Fatalf("GetWorkspaceSymbols failed for '%s': %v", term, err)
		}

		t.Logf("Workspace search for '%s': %d symbols found", term, len(symbols))
		for i, symbol := range symbols {
			if i < 3 { // Log first 3 results
				t.Logf("  - %s (%s)", symbol.Name, symbol.Kind)
			}
		}
	}

	// Test that each document has symbols
	for uri := range documents {
		symbols, err := ls.GetDocumentSymbols(uri)
		if err != nil {
			t.Fatalf("GetDocumentSymbols failed for %s: %v", uri, err)
		}

		if len(symbols) == 0 {
			t.Errorf("Expected symbols in %s, got none", uri)
		}

		t.Logf("Document %s has %d symbols", uri, len(symbols))
	}
}

// TestLargeFilePerformance tests performance with large files
func TestLargeFilePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}

	ls, err := NewLanguageService()
	if err != nil {
		t.Fatalf("Failed to create language service: %v", err)
	}

	// Generate a large Perl file (10,000 lines)
	var builder strings.Builder
	builder.WriteString("#!/usr/bin/perl\nuse strict;\nuse warnings;\n\n")

	// Generate many packages and subroutines
	for pkg := 0; pkg < 100; pkg++ {
		builder.WriteString(fmt.Sprintf("package TestPackage%d;\n\n", pkg))

		for sub := 0; sub < 100; sub++ {
			builder.WriteString(fmt.Sprintf("sub test_function_%d_%d {\n", pkg, sub))
			builder.WriteString("    my ($param1, $param2) = @_;\n")
			builder.WriteString(fmt.Sprintf("    my $result = $param1 + $param2 + %d;\n", sub))
			builder.WriteString("    return $result;\n")
			builder.WriteString("}\n\n")
		}
	}

	largeContent := builder.String()
	uri := "file:///large_file.pl"

	// Test document update performance
	start := time.Now()
	err = ls.UpdateDocument(uri, largeContent, 1)
	if err != nil {
		t.Fatalf("UpdateDocument failed: %v", err)
	}
	updateDuration := time.Since(start)

	t.Logf("Large file update took: %v", updateDuration)

	// Performance target: should complete within reasonable time
	if updateDuration > 10*time.Second {
		t.Errorf("Large file update too slow: %v", updateDuration)
	}

	// Test hover performance on large file
	pos := Position{Line: 500, Character: 10}
	start = time.Now()
	hover, err := ls.GetHover(uri, pos)
	if err != nil {
		t.Fatalf("GetHover failed: %v", err)
	}
	hoverDuration := time.Since(start)

	t.Logf("Hover on large file took: %v", hoverDuration)

	if hoverDuration > 100*time.Millisecond {
		t.Errorf("Hover on large file too slow: %v", hoverDuration)
	}

	if hover != nil {
		t.Logf("Hover result: %s", hover.Contents[:min(50, len(hover.Contents))])
	}

	// Test completion performance
	start = time.Now()
	completions, err := ls.GetCompletions(uri, pos)
	if err != nil {
		t.Fatalf("GetCompletions failed: %v", err)
	}
	completionDuration := time.Since(start)

	t.Logf("Completion on large file took: %v (found %d items)",
		completionDuration, len(completions))

	if completionDuration > 200*time.Millisecond {
		t.Errorf("Completion on large file too slow: %v", completionDuration)
	}

	// Test document symbols performance
	start = time.Now()
	symbols, err := ls.GetDocumentSymbols(uri)
	if err != nil {
		t.Fatalf("GetDocumentSymbols failed: %v", err)
	}
	symbolsDuration := time.Since(start)

	t.Logf("Document symbols on large file took: %v (found %d symbols)",
		symbolsDuration, len(symbols))

	if symbolsDuration > 500*time.Millisecond {
		t.Errorf("Document symbols on large file too slow: %v", symbolsDuration)
	}
}

// TestConcurrentAccess tests concurrent access to the language service
func TestConcurrentAccess(t *testing.T) {
	ls, err := NewLanguageService()
	if err != nil {
		t.Fatalf("Failed to create language service: %v", err)
	}

	content := generatePerlCode(200)
	uri := "file:///concurrent_test.pl"

	err = ls.UpdateDocument(uri, content, 1)
	if err != nil {
		t.Fatalf("UpdateDocument failed: %v", err)
	}

	// Launch multiple goroutines to access LSP features concurrently
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			pos := Position{Line: id % 20, Character: 5}

			// Test hover
			hover, err := ls.GetHover(uri, pos)
			if err != nil {
				t.Errorf("Concurrent GetHover failed (goroutine %d): %v", id, err)
				return
			}

			// Test completions
			completions, err := ls.GetCompletions(uri, pos)
			if err != nil {
				t.Errorf("Concurrent GetCompletions failed (goroutine %d): %v", id, err)
				return
			}

			// Test definition
			definition, err := ls.GetDefinition(uri, pos)
			if err != nil {
				t.Errorf("Concurrent GetDefinition failed (goroutine %d): %v", id, err)
				return
			}

			t.Logf("Goroutine %d: hover=%v, completions=%d, definition=%v",
				id, hover != nil, len(completions), definition != nil)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestErrorHandling tests error handling in various scenarios
func TestErrorHandling(t *testing.T) {
	ls, err := NewLanguageService()
	if err != nil {
		t.Fatalf("Failed to create language service: %v", err)
	}

	// Test with invalid Perl syntax
	invalidContent := `#!/usr/bin/perl
use strict;
use warnings;

# Invalid syntax
my $var = ;  # Missing value
sub incomplete_sub   # Missing body
print "unclosed string
`

	uri := "file:///invalid.pl"

	// UpdateDocument should handle parse errors gracefully
	err = ls.UpdateDocument(uri, invalidContent, 1)
	// We expect this might fail, but shouldn't crash
	t.Logf("UpdateDocument with invalid syntax: error=%v", err)

	// Test operations on non-existent document
	nonExistentURI := "file:///does_not_exist.pl"
	pos := Position{Line: 0, Character: 0}

	hover, err := ls.GetHover(nonExistentURI, pos)
	if err != nil {
		t.Fatalf("GetHover on non-existent document failed: %v", err)
	}
	if hover != nil {
		t.Error("Expected nil hover for non-existent document")
	}

	completions, err := ls.GetCompletions(nonExistentURI, pos)
	if err != nil {
		t.Fatalf("GetCompletions on non-existent document failed: %v", err)
	}
	if completions != nil {
		t.Error("Expected nil completions for non-existent document")
	}

	// Test with extreme positions
	extremePos := Position{Line: 999999, Character: 999999}

	hover, err = ls.GetHover(uri, extremePos)
	if err != nil {
		t.Fatalf("GetHover with extreme position failed: %v", err)
	}
	if hover != nil {
		t.Log("Hover with extreme position returned result (unexpected but not error)")
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
