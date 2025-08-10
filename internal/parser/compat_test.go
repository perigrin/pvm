// ABOUTME: Compatibility tests for parser interface
// ABOUTME: Ensures backward compatibility across different parser creation methods

package parser

import (
	"strings"
	"testing"
)

func TestNewParserWithOptions_Fallback(t *testing.T) {
	// Test that NewParserWithOptions with useScanner=false works like NewParser
	parser, err := NewParserWithOptions(false)
	if err != nil {
		t.Fatalf("Failed to create parser with fallback: %v", err)
	}

	// Test basic parsing functionality
	content := `my $x = 42;`
	ast, err := parser.ParseString(content)
	if err != nil {
		t.Fatalf("Failed to parse simple content: %v", err)
	}

	if ast == nil {
		t.Fatal("AST should not be nil")
	}

	if ast.Root == nil {
		t.Fatal("Root node should not be nil")
	}
}

func TestBackwardCompatibility_NewParser(t *testing.T) {
	// Ensure NewParser still works exactly as before
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser failed: %v", err)
	}

	content := `my Int $count = 42;`
	ast, err := parser.ParseString(content)
	if err != nil {
		t.Fatalf("Failed to parse typed content: %v", err)
	}

	if ast == nil {
		t.Fatal("AST should not be nil")
	}

	// The parser should still find type annotations
	if len(ast.TypeAnnotations) == 0 {
		t.Error("Expected to find type annotations")
	}
}

func TestNewParserWithOptions_UseScanner(t *testing.T) {
	// Test that NewParserWithOptions(true) works (now uses tree-sitter since scanner removed)
	parser, err := NewParserWithOptions(true)
	if err != nil {
		t.Fatalf("Failed to create parser with useScanner=true: %v", err)
	}

	content := `my $x = 42;`
	ast, err := parser.ParseString(content)
	if err != nil {
		t.Fatalf("Failed to parse content: %v", err)
	}

	if ast == nil {
		t.Fatal("AST should not be nil")
	}
}

func TestParserInterface_Consistency(t *testing.T) {
	// Test that both parser types implement the same interface consistently
	parsers := []struct {
		name   string
		parser func() (Parser, error)
	}{
		{"traditional", func() (Parser, error) { return NewParser() }},
		{"fallback", func() (Parser, error) { return NewParserWithOptions(false) }},
	}

	content := `my $test = "hello";`

	for _, tc := range parsers {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := tc.parser()
			if err != nil {
				t.Fatalf("Failed to create %s parser: %v", tc.name, err)
			}

			// Test ParseString
			ast1, err := parser.ParseString(content)
			if err != nil {
				t.Fatalf("ParseString failed for %s: %v", tc.name, err)
			}
			if ast1 == nil {
				t.Fatalf("ParseString returned nil AST for %s", tc.name)
			}

			// Test ParseReader
			reader := strings.NewReader(content)
			ast2, err := parser.ParseReader(reader)
			if err != nil {
				t.Fatalf("ParseReader failed for %s: %v", tc.name, err)
			}
			if ast2 == nil {
				t.Fatalf("ParseReader returned nil AST for %s", tc.name)
			}

			// Both should parse the same content successfully
			if ast1.Root == nil || ast2.Root == nil {
				t.Errorf("Root nodes should not be nil for %s", tc.name)
			}
		})
	}
}

func TestNodeInterface_Compatibility(t *testing.T) {
	// Test that Node interface works consistently
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	content := `my $x = 1;`
	ast, err := parser.ParseString(content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if ast.Root == nil {
		t.Fatal("Root should not be nil")
	}

	// Test Node interface methods
	rootType := ast.Root.Type()
	if rootType == "" {
		t.Error("Node Type() should not return empty string")
	}

	startPos := ast.Root.Start()
	endPos := ast.Root.End()
	if startPos.Line < 1 {
		t.Error("Start position should be valid")
	}
	if endPos.Offset < startPos.Offset {
		t.Error("End position should be after start position")
	}

	children := ast.Root.Children()
	// Children should be a valid slice (may be empty)
	if children == nil {
		t.Error("Children() should return a slice, not nil")
	}

	text := ast.Root.Text()
	// Text should contain the source (may be empty for abstract nodes)
	_ = text // No specific assertion needed, just ensure it doesn't panic
}

func TestErrorHandling_Consistency(t *testing.T) {
	// Test that error handling is consistent across parser types
	parsers := []func() (Parser, error){
		func() (Parser, error) { return NewParser() },
		func() (Parser, error) { return NewParserWithOptions(false) },
	}

	for i, createParser := range parsers {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			parser, err := createParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			// Test parsing non-existent file
			_, err = parser.ParseFile("/non/existent/file.pl")
			if err == nil {
				t.Error("Expected error when parsing non-existent file")
			}

			// Test parsing empty content (should succeed)
			ast, err := parser.ParseString("")
			if err != nil {
				t.Errorf("Parsing empty string should succeed: %v", err)
			}
			if ast == nil {
				t.Error("Empty string should still return an AST")
			}
		})
	}
}
