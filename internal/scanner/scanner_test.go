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

	content := `my Int $count = 42;`

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

	// We should have at least: my, Int, $count, =, 42, ;
	if len(tokens) < 6 {
		t.Errorf("Expected at least 6 tokens, got %d", len(tokens))
	}

	// Verify some key tokens
	expectedTypes := []TokenType{TokenMy, TokenIdentifier, TokenVariable, TokenAssign, TokenNumber, TokenSemicolon}
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

	// Should contain tokens for: my, ArrayRef, [, Int, ], $numbers, ;
	if len(tokens) < 7 {
		t.Errorf("Expected at least 7 tokens for parameterized type, got %d", len(tokens))
	}

	// Check for bracket tokens
	foundLBracket := false
	foundRBracket := false
	for _, token := range tokens {
		if token.Type() == TokenLBracket {
			foundLBracket = true
		}
		if token.Type() == TokenRBracket {
			foundRBracket = true
		}
	}

	if !foundLBracket {
		t.Error("Expected to find left bracket token")
	}
	if !foundRBracket {
		t.Error("Expected to find right bracket token")
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

	// Check for pipe token
	foundPipe := false
	for _, token := range tokens {
		if token.Type() == TokenPipe && token.Value() == "|" {
			foundPipe = true
		}
	}

	if !foundPipe {
		t.Error("Expected to find pipe token for union type")
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
	foundMethod := false
	foundArrow := false
	for _, token := range tokens {
		if token.Type() == TokenMethod {
			foundMethod = true
		}
		if token.Type() == TokenArrow {
			foundArrow = true
		}
	}

	if !foundMethod {
		t.Error("Expected to find method token")
	}
	if !foundArrow {
		t.Error("Expected to find arrow token for return type")
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

	// Should have at least EOF token
	if !iter.HasNext() {
		t.Error("Empty string should still produce EOF token")
	}

	token := iter.Next()
	if token == nil {
		t.Error("Expected at least EOF token")
	}
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
		{TokenMy, "MY"},
		{TokenVariable, "VARIABLE"},
		{TokenPipe, "PIPE"},
		{TokenLBracket, "LBRACKET"},
		{TokenEOF, "EOF"},
		{TokenType(9999), "UNKNOWN"},
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

	if !foundLine1 || !foundLine2 {
		t.Error("Expected tokens from both line 1 and line 2")
	}
}
