// ABOUTME: Basic scanner tests that don't require tree-sitter
// ABOUTME: Tests core token types and scanner interface

package scanner

import (
	"testing"
)

func TestTokenType_StringBasic(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  string
	}{
		{TokenEOF, "EOF"},
		{TokenError, "ERROR"},
		{TokenMy, "MY"},
		{TokenVariable, "VARIABLE"},
		{TokenPipe, "PIPE"},
		{TokenLBracket, "LBRACKET"},
		{TokenTypeKeyword, "TYPE"},
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

func TestTokenInterface(t *testing.T) {
	// Test token interface implementation
	token := &treeSitterToken{
		tokenType: TokenMy,
		value:     "my",
		position:  Position{Line: 1, Column: 1, Offset: 0},
		length:    2,
	}

	if token.Type() != TokenMy {
		t.Errorf("Expected TokenMy, got %s", token.Type())
	}

	if token.Value() != "my" {
		t.Errorf("Expected 'my', got %q", token.Value())
	}

	if token.Position().Line != 1 {
		t.Errorf("Expected line 1, got %d", token.Position().Line)
	}

	if token.Length() != 2 {
		t.Errorf("Expected length 2, got %d", token.Length())
	}
}

func TestTokenIterator_Basic(t *testing.T) {
	// Create a simple token iterator with mock tokens
	tokens := []Token{
		&treeSitterToken{tokenType: TokenMy, value: "my"},
		&treeSitterToken{tokenType: TokenVariable, value: "$x"},
		&treeSitterToken{tokenType: TokenEOF, value: ""},
	}

	iter := &tokenIterator{
		tokens: tokens,
		pos:    0,
	}

	// Test HasNext
	if !iter.HasNext() {
		t.Error("Iterator should have tokens")
	}

	// Test Peek
	first := iter.Peek()
	if first == nil {
		t.Fatal("Peek should not return nil")
	}
	if first.Type() != TokenMy {
		t.Errorf("Expected TokenMy, got %s", first.Type())
	}

	// Test that Peek doesn't advance
	second := iter.Peek()
	if first.Value() != second.Value() {
		t.Error("Peek should return the same token")
	}

	// Test Next
	third := iter.Next()
	if third.Type() != TokenMy {
		t.Errorf("Expected TokenMy, got %s", third.Type())
	}

	// Test Position
	if iter.Position() != 1 {
		t.Errorf("Expected position 1, got %d", iter.Position())
	}

	// Test Reset
	iter.Reset()
	if iter.Position() != 0 {
		t.Errorf("Reset should set position to 0, got %d", iter.Position())
	}
}
