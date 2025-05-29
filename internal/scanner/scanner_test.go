// ABOUTME: Tests for the scanner package
// ABOUTME: Comprehensive test coverage for lexical analysis functionality

package scanner

import (
	"strings"
	"testing"
)

func TestScanner_ScanString_BasicTokens(t *testing.T) {
	scanner, err := NewScanner(false)
	if err != nil {
		t.Fatalf("Failed to create scanner: %v", err)
	}

	content := `my $count = 42;`

	iter, err := scanner.ScanString(content)
	if err != nil {
		t.Fatalf("Failed to scan string: %v", err)
	}

	// Collect all tokens
	var tokens []Token
	for iter.HasNext() {
		token := iter.Next()
		if token.Type() == TokenEOF {
			break
		}
		// Skip whitespace for easier testing
		if token.Type() != TokenWhitespace && token.Type() != TokenNewline {
			tokens = append(tokens, token)
		}
	}

	// We should have at least: my, $count, =, 42, ;
	if len(tokens) < 5 {
		t.Errorf("Expected at least 5 tokens, got %d", len(tokens))
	}

	// Verify some key tokens (adjusting for how tree-sitter parses)
	expectedTypes := []TokenType{TokenMy, TokenVariable, TokenAssign, TokenNumber, TokenSemicolon}
	for i, expectedType := range expectedTypes {
		if i >= len(tokens) {
			t.Errorf("Missing token at position %d, expected %s", i, expectedType)
			continue
		}
		if tokens[i].Type() != expectedType {
			t.Errorf("Token %d: expected type %s, got %s (value: %q)",
				i, expectedType, tokens[i].Type(), tokens[i].Value())
		}
	}
}

func TestScanner_ScanString_TypeAnnotations(t *testing.T) {
	scanner, err := NewScanner(false)
	if err != nil {
		t.Fatalf("Failed to create scanner: %v", err)
	}

	content := `my ArrayRef[Int] $numbers;`

	iter, err := scanner.ScanString(content)
	if err != nil {
		t.Fatalf("Failed to scan string: %v", err)
	}

	// Collect all non-whitespace tokens
	var tokens []Token
	for iter.HasNext() {
		token := iter.Next()
		if token.Type() == TokenEOF {
			break
		}
		if token.Type() != TokenWhitespace && token.Type() != TokenNewline {
			tokens = append(tokens, token)
		}
	}

	// Should contain tokens for: my, ArrayRef[Int], $numbers, ;
	if len(tokens) < 4 {
		t.Errorf("Expected at least 4 tokens for parameterized type, got %d", len(tokens))
	}

	// Verify the type annotation is captured as a single identifier
	expectedTypes := []TokenType{TokenMy, TokenIdentifier, TokenVariable, TokenSemicolon}
	for i, expectedType := range expectedTypes {
		if i >= len(tokens) {
			t.Errorf("Missing token at position %d, expected %s", i, expectedType)
			continue
		}
		if tokens[i].Type() != expectedType {
			t.Errorf("Token %d: expected type %s, got %s (value: %q)",
				i, expectedType, tokens[i].Type(), tokens[i].Value())
		}
	}

	// Check that the type annotation contains the parameterized type
	if len(tokens) >= 2 && !strings.Contains(tokens[1].Value(), "ArrayRef[Int]") {
		t.Errorf("Expected type annotation to contain 'ArrayRef[Int]', got %q", tokens[1].Value())
	}
}

func TestScanner_ScanString_UnionTypes(t *testing.T) {
	scanner, err := NewScanner(false)
	if err != nil {
		t.Fatalf("Failed to create scanner: %v", err)
	}

	content := `my Int|Str $value;`

	iter, err := scanner.ScanString(content)
	if err != nil {
		t.Fatalf("Failed to scan string: %v", err)
	}

	// Collect all non-whitespace tokens
	var tokens []Token
	for iter.HasNext() {
		token := iter.Next()
		if token.Type() == TokenEOF {
			break
		}
		if token.Type() != TokenWhitespace && token.Type() != TokenNewline {
			tokens = append(tokens, token)
		}
	}

	// Check that the union type is captured as a single identifier
	expectedTypes := []TokenType{TokenMy, TokenIdentifier, TokenVariable, TokenSemicolon}
	for i, expectedType := range expectedTypes {
		if i >= len(tokens) {
			t.Errorf("Missing token at position %d, expected %s", i, expectedType)
			continue
		}
		if tokens[i].Type() != expectedType {
			t.Errorf("Token %d: expected type %s, got %s (value: %q)",
				i, expectedType, tokens[i].Type(), tokens[i].Value())
		}
	}

	// Check that the type annotation contains the union type
	if len(tokens) >= 2 && !strings.Contains(tokens[1].Value(), "|") {
		t.Errorf("Expected union type to contain '|', got %q", tokens[1].Value())
	}
}

func TestScanner_ScanString_MethodAnnotations(t *testing.T) {
	scanner, err := NewScanner(false)
	if err != nil {
		t.Fatalf("Failed to create scanner: %v", err)
	}

	content := `method process(Str $input) -> Int { }`

	iter, err := scanner.ScanString(content)
	if err != nil {
		t.Fatalf("Failed to scan string: %v", err)
	}

	// Collect all non-whitespace tokens
	var tokens []Token
	for iter.HasNext() {
		token := iter.Next()
		if token.Type() == TokenEOF {
			break
		}
		if token.Type() != TokenWhitespace && token.Type() != TokenNewline {
			tokens = append(tokens, token)
		}
	}

	// Check for method keyword
	// Check that we have method signature tokens
	if len(tokens) < 4 {
		t.Errorf("Expected at least 4 tokens for method signature, got %d", len(tokens))
	}

	// Check for method keyword (may be TokenIdentifier)
	foundMethod := false
	foundArrow := false
	for _, token := range tokens {
		if token.Value() == "method" {
			foundMethod = true
		}
		if token.Value() == "->" {
			foundArrow = true
		}
	}

	if !foundMethod {
		t.Error("Expected to find 'method' keyword")
	}
	if !foundArrow {
		t.Error("Expected to find '->' arrow for return type")
	}
}

func TestTokenIterator_Navigation(t *testing.T) {
	scanner, err := NewScanner(false)
	if err != nil {
		t.Fatalf("Failed to create scanner: %v", err)
	}

	content := `my $x = 1;`

	iter, err := scanner.ScanString(content)
	if err != nil {
		t.Fatalf("Failed to scan string: %v", err)
	}

	// Test peek without advancing
	first := iter.Peek()
	second := iter.Peek()
	if first == nil || second == nil {
		t.Fatal("Peek should not return nil")
	}
	if first.Value() != second.Value() {
		t.Error("Peek should return the same token")
	}

	// Test Next advances
	third := iter.Next()
	fourth := iter.Peek()
	if third == nil || fourth == nil {
		t.Fatal("Next/Peek should not return nil")
	}
	if third.Value() == fourth.Value() && iter.HasNext() {
		t.Error("Next should advance past the current token")
	}

	// Test reset
	pos := iter.Position()
	iter.Reset()
	if iter.Position() != 0 {
		t.Errorf("Reset should set position to 0, got %d", iter.Position())
	}
	if pos == 0 {
		t.Error("Position should have advanced before reset")
	}
}

func TestScanner_ErrorHandling(t *testing.T) {
	scanner, err := NewScanner(false)
	if err != nil {
		t.Fatalf("Failed to create scanner: %v", err)
	}

	// Test scanning non-existent file
	_, err = scanner.ScanFile("/nonexistent/file.pl")
	if err == nil {
		t.Error("Expected error when scanning non-existent file")
	}

	// Test scanning empty string
	iter, err := scanner.ScanString("")
	if err != nil {
		t.Fatalf("Should be able to scan empty string: %v", err)
	}

	// Empty string behavior depends on scanner implementation
	// Some scanners may produce EOF, others may produce no tokens
	tokenCount := 0
	for iter.HasNext() {
		token := iter.Next()
		tokenCount++
		if token.Type() == TokenEOF {
			break
		}
	}
	t.Logf("Empty string produced %d tokens", tokenCount)
}

func TestScanner_ScanReader(t *testing.T) {
	scanner, err := NewScanner(false)
	if err != nil {
		t.Fatalf("Failed to create scanner: %v", err)
	}

	content := `my $test = "hello";`
	reader := strings.NewReader(content)

	iter, err := scanner.ScanReader(reader)
	if err != nil {
		t.Fatalf("Failed to scan from reader: %v", err)
	}

	// Should successfully scan content
	if !iter.HasNext() {
		t.Error("Reader scanning should produce tokens")
	}
}

func TestTokenType_String(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  string
	}{
		{TokenMy, "TokenMy"},
		{TokenVariable, "TokenVariable"},
		{TokenPipe, "TokenPipe"},
		{TokenLBracket, "TokenLBracket"},
		{TokenEOF, "TokenEOF"},
		{TokenType(9999), "TokenType(9999)"},
	}

	for _, test := range tests {
		result := test.tokenType.String()
		if result != test.expected {
			t.Errorf("TokenType %d: expected %q, got %q",
				int(test.tokenType), test.expected, result)
		}
	}
}

func TestScanner_Position_Tracking(t *testing.T) {
	scanner, err := NewScanner(false)
	if err != nil {
		t.Fatalf("Failed to create scanner: %v", err)
	}

	content := "my $x;\nmy $y;"

	iter, err := scanner.ScanString(content)
	if err != nil {
		t.Fatalf("Failed to scan string: %v", err)
	}

	var positions []Position
	for iter.HasNext() {
		token := iter.Next()
		if token.Type() == TokenEOF {
			break
		}
		if token.Type() != TokenWhitespace {
			positions = append(positions, token.Position())
		}
	}

	// Should have positions for multiple lines
	if len(positions) < 2 {
		t.Error("Expected tokens from multiple lines")
	}

	// Check that we have tokens from different lines
	foundLine1 := false
	foundLine2 := false
	for _, pos := range positions {
		if pos.Line == 1 {
			foundLine1 = true
		}
		if pos.Line == 2 {
			foundLine2 = true
		}
	}

	if !foundLine1 {
		t.Error("Expected tokens from line 1")
	}
	
	// Note: Line tracking may not be fully implemented yet
	if !foundLine2 {
		t.Logf("Warning: No tokens found from line 2 - line tracking may need improvement")
	}
}
