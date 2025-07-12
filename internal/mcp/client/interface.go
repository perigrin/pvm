// ABOUTME: Interface definitions for MCP client implementations
// ABOUTME: Provides abstractions for both real and mock MCP clients

package client

import (
	"context"
	"fmt"
	"strings"
)

// MCPClientInterface defines the interface for MCP client implementations
type MCPClientInterface interface {
	// Connect establishes a connection with the MCP server
	Connect(ctx context.Context) error

	// Sample sends a sampling request to the MCP server
	Sample(ctx context.Context, params SamplingParams) (*SamplingResult, error)

	// Disconnect closes the connection to the MCP server
	Disconnect(ctx context.Context) error

	// IsConnected returns true if the client is connected
	IsConnected() bool

	// GetCapabilities returns the negotiated capabilities
	GetCapabilities() *MCPCapabilities

	// HealthCheck performs a health check on the MCP connection
	HealthCheck(ctx context.Context) error
}

// MockMCPClient provides a mock implementation for testing
type MockMCPClient struct {
	connected    bool
	capabilities *MCPCapabilities
	responses    map[string]*SamplingResult
}

// NewMockMCPClient creates a new mock MCP client
func NewMockMCPClient() *MockMCPClient {
	return &MockMCPClient{
		capabilities: &MCPCapabilities{
			Sampling:  true,
			Resources: true,
			Prompts:   true,
			MaxTokens: 1000,
			Models:    []string{"mock-model-v1"},
		},
		responses: make(map[string]*SamplingResult),
	}
}

// Connect simulates connecting to an MCP server
func (m *MockMCPClient) Connect(ctx context.Context) error {
	m.connected = true
	return nil
}

// Sample returns a mock sampling response
func (m *MockMCPClient) Sample(ctx context.Context, params SamplingParams) (*SamplingResult, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected to MCP server")
	}

	// Generate mock response based on the prompt
	content := m.generateMockContent(params)

	return &SamplingResult{
		Content:    content,
		Model:      "mock-model-v1",
		StopReason: "end_turn",
		Usage: MCPUsage{
			InputTokens:  100,
			OutputTokens: 200,
			TotalTokens:  300,
		},
	}, nil
}

// Disconnect simulates disconnecting from the MCP server
func (m *MockMCPClient) Disconnect(ctx context.Context) error {
	m.connected = false
	return nil
}

// IsConnected returns the connection status
func (m *MockMCPClient) IsConnected() bool {
	return m.connected
}

// GetCapabilities returns the mock capabilities
func (m *MockMCPClient) GetCapabilities() *MCPCapabilities {
	if m.capabilities == nil {
		return nil
	}
	caps := *m.capabilities
	return &caps
}

// HealthCheck simulates a health check
func (m *MockMCPClient) HealthCheck(ctx context.Context) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

// SetMockResponse allows setting custom responses for testing
func (m *MockMCPClient) SetMockResponse(prompt string, result *SamplingResult) {
	m.responses[prompt] = result
}

// generateMockContent generates mock content based on the sampling parameters
func (m *MockMCPClient) generateMockContent(params SamplingParams) string {
	if len(params.Messages) == 0 {
		return "# Generated mock response"
	}

	userMessage := params.Messages[len(params.Messages)-1].Content.Text

	// Check for custom responses first
	if result, exists := m.responses[userMessage]; exists {
		return result.Content
	}

	// Generate appropriate mock responses based on message content
	switch {
	case contains(userMessage, "missing sigil"):
		return "my $variable = 42;"
	case contains(userMessage, "type mismatch"):
		return "my Int $count = int($value);"
	case contains(userMessage, "undefined variable"):
		return "my $undefined_var;"
	case contains(userMessage, "generate function") || contains(userMessage, "Generate a Perl function"):
		return `sub add_numbers {
    my ($num1, $num2) = @_;
    return $num1 + $num2;
}`
	case contains(userMessage, "Generate a Perl test") || contains(userMessage, "test framework") || contains(userMessage, "Generate comprehensive Perl tests"):
		// Extract function name and framework from the prompt
		funcName := extractFunctionNameFromPrompt(userMessage)
		framework := extractFrameworkFromPrompt(userMessage)
		if framework == "" {
			framework = "Test2::V0"
		}
		if funcName == "" {
			funcName = "test_function"
		}

		return fmt.Sprintf(`use %s;
use strict;
use warnings;

# Test %s function
ok(1, "valid input");
ok(1, "invalid input");
ok(1, "edge cases");
ok(1, "type constraints");
ok(1, "return type correctness");

# Test %s with specific cases
is(%s(42), "42", "%s converts integer to string");

done_testing();`, framework, funcName, funcName, funcName, funcName)
	case contains(userMessage, "Generate a Perl class"):
		return `package MyClass;
use v5.40;
use warnings;

sub new {
    my ($class) = @_;
    my $self = bless {}, $class;
    return $self;
}

1;`
	case contains(userMessage, "refactor"):
		switch {
		case contains(userMessage, "extract"):
			return `sub extracted_method {
    my ($param) = @_;
    return $param + 1;
}`
		case contains(userMessage, "rename"):
			return `my $meaningful_name = 42;
print $meaningful_name;`
		case contains(userMessage, "inline"):
			return `my $result = 3.14159;`
		default:
			return `# Refactored code
my $improved_code = 1;`
		}
	case contains(userMessage, "Generate documentation") || contains(userMessage, "POD documentation") || contains(userMessage, "pod") || contains(userMessage, "=head") || contains(userMessage, "documentation"):
		return `=head1 NAME

MyModule - A sample Perl module

=head1 SYNOPSIS

    use MyModule;
    my $obj = MyModule->new();

=head1 DESCRIPTION

This module provides sample functionality.

=cut`
	case contains(userMessage, "completion") || contains(userMessage, "Complete the following") || contains(userMessage, "Provide code completion suggestions"):
		// Generate structured completion suggestions
		return `SUGGESTION: length
DESCRIPTION: Built-in function to get string length
TYPE: Str -> Int

SUGGESTION: substr
DESCRIPTION: Extract substring from string
TYPE: (Str, Int, Int?) -> Str

SUGGESTION: chomp
DESCRIPTION: Remove trailing newline characters
TYPE: Str -> Str`
	default:
		return "# Generated code based on prompt"
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// extractFunctionNameFromPrompt extracts function name from test generation prompts
func extractFunctionNameFromPrompt(prompt string) string {
	// Look for various patterns to extract function names
	if strings.Contains(prompt, "FunctionName:") {
		start := strings.Index(prompt, "FunctionName:") + len("FunctionName:")
		end := strings.Index(prompt[start:], "\n")
		if end == -1 {
			end = len(prompt) - start
		}
		return strings.TrimSpace(prompt[start : start+end])
	}

	// Look for "function 'function_name'"
	if strings.Contains(prompt, "function '") {
		start := strings.Index(prompt, "function '") + len("function '")
		end := strings.Index(prompt[start:], "'")
		if end == -1 {
			return ""
		}
		return prompt[start : start+end]
	}

	// Look for function names in context
	if strings.Contains(prompt, "int_to_string") {
		return "int_to_string"
	}
	if strings.Contains(prompt, "sum_array") {
		return "sum_array"
	}

	return ""
}

// extractFrameworkFromPrompt extracts test framework from prompts
func extractFrameworkFromPrompt(prompt string) string {
	if contains(prompt, "Test::More") {
		return "Test::More"
	}
	if contains(prompt, "Test2::V0") {
		return "Test2::V0"
	}
	return ""
}
