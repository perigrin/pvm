// ABOUTME: MCP sampling client for collaborating with LLMs
// ABOUTME: Provides sampling infrastructure for code generation and fixes

package generation

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SamplingClient provides MCP sampling capabilities
type SamplingClient struct {
	// In a real implementation, this would hold the MCP client connection
	// For now, we'll mock the responses
	enabled bool
}

// NewSamplingClient creates a new sampling client
func NewSamplingClient(enabled bool) *SamplingClient {
	return &SamplingClient{
		enabled: enabled,
	}
}

// SamplingRequest represents a request for LLM sampling
type SamplingRequest struct {
	Prompt       string            `json:"prompt"`
	MaxTokens    int               `json:"max_tokens,omitempty"`
	Temperature  float64           `json:"temperature,omitempty"`
	SystemPrompt string            `json:"system_prompt,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// SamplingResponse represents the response from LLM sampling
type SamplingResponse struct {
	Content    string    `json:"content"`
	Confidence float64   `json:"confidence"`
	Model      string    `json:"model"`
	Timestamp  time.Time `json:"timestamp"`
}

// Sample sends a sampling request to the LLM
func (c *SamplingClient) Sample(ctx context.Context, prompt string, projectPath string) (*SamplingResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("sampling is disabled")
	}

	// Create sampling request
	request := &SamplingRequest{
		Prompt:       prompt,
		MaxTokens:    1000,
		Temperature:  0.7,
		SystemPrompt: "You are a Perl programming expert helping to fix code errors and generate typed Perl code.",
		Metadata: map[string]string{
			"project_path": projectPath,
			"language":     "perl",
		},
	}

	// In a real implementation, we would:
	// 1. Create an MCP sampling request
	// 2. Send it to the connected LLM
	// 3. Wait for the response
	// 4. Parse and return the result

	// For now, return a mock response based on the prompt
	response := c.generateMockResponse(request)

	return response, nil
}

// SampleWithOptions sends a sampling request with custom options
func (c *SamplingClient) SampleWithOptions(ctx context.Context, request *SamplingRequest) (*SamplingResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("sampling is disabled")
	}

	// Validate request
	if request.Prompt == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
	}

	// Set defaults
	if request.MaxTokens == 0 {
		request.MaxTokens = 1000
	}
	if request.Temperature == 0 {
		request.Temperature = 0.7
	}

	// In real implementation, send to LLM
	response := c.generateMockResponse(request)

	return response, nil
}

// BatchSample sends multiple sampling requests in parallel
func (c *SamplingClient) BatchSample(ctx context.Context, prompts []string, projectPath string) ([]*SamplingResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("sampling is disabled")
	}

	responses := make([]*SamplingResponse, len(prompts))

	// In real implementation, these would be sent in parallel
	for i, prompt := range prompts {
		response, err := c.Sample(ctx, prompt, projectPath)
		if err != nil {
			return nil, fmt.Errorf("failed to sample prompt %d: %w", i, err)
		}
		responses[i] = response
	}

	return responses, nil
}

// generateMockResponse generates a mock response for testing
func (c *SamplingClient) generateMockResponse(request *SamplingRequest) *SamplingResponse {
	// Generate appropriate mock responses based on prompt content
	content := ""
	confidence := 0.85

	switch {
	case contains(request.Prompt, "missing sigil"):
		content = "my $variable = 42;"
		confidence = 0.95
	case contains(request.Prompt, "type mismatch"):
		content = "my Int $count = int($value);"
		confidence = 0.9
	case contains(request.Prompt, "undefined variable"):
		content = "my $undefined_var;"
		confidence = 0.8
	case contains(request.Prompt, "generate function") || contains(request.Prompt, "Generate a Perl function"):
		content = `sub add_numbers {
    my ($num1, $num2) = @_;
    return $num1 + $num2;
}`
		confidence = 0.85
	case contains(request.Prompt, "Generate a Perl test") || contains(request.Prompt, "test framework") || contains(request.Prompt, "Generate comprehensive Perl tests"):
		// Extract function name from prompt if available
		funcName := extractFunctionName(request.Prompt)
		framework := extractFramework(request.Prompt)
		if framework == "" {
			framework = "Test2::V0"
		}
		if funcName == "" {
			funcName = "test_function"
		}

		content = fmt.Sprintf(`use %s;
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
		confidence = 0.85
	case contains(request.Prompt, "Generate a Perl class") || contains(request.Prompt, "class pattern"):
		content = `package MyClass;
use v5.40;
use warnings;

sub new {
    my ($class) = @_;
    my $self = bless {}, $class;
    return $self;
}

1;`
		confidence = 0.85
	case contains(request.Prompt, "naming convention"):
		content = "snake_case"
		confidence = 0.9
	case contains(request.Prompt, "completion") || contains(request.Prompt, "Complete the following") || contains(request.Prompt, "Provide code completion suggestions"):
		// Generate structured completion suggestions
		content = `SUGGESTION: length
DESCRIPTION: Built-in function to get string length
TYPE: Str -> Int

SUGGESTION: substr
DESCRIPTION: Extract substring from string
TYPE: (Str, Int, Int?) -> Str

SUGGESTION: chomp
DESCRIPTION: Remove trailing newline characters
TYPE: Str -> Str`
		confidence = 0.8
	case contains(request.Prompt, "refactor") || contains(request.Prompt, "Refactor"):
		if contains(request.Prompt, "extract") {
			content = `sub extracted_method {
    my ($param) = @_;
    return $param + 1;
}`
		} else if contains(request.Prompt, "rename") {
			content = `my $meaningful_name = 42;
print $meaningful_name;`
		} else if contains(request.Prompt, "inline") {
			content = `my $result = 3.14159;`
		} else {
			content = `# Refactored code
my $improved_code = 1;`
		}
		confidence = 0.8
	case contains(request.Prompt, "Generate documentation") || contains(request.Prompt, "POD documentation") || contains(request.Prompt, "Generate concise Perl documentation") || contains(request.Prompt, "Generate detailed Perl documentation"):
		content = `=head1 NAME

MyModule - A sample Perl module

=head1 SYNOPSIS

    use MyModule;
    my $obj = MyModule->new();

=head1 DESCRIPTION

This module provides sample functionality.

=head1 METHODS

=head2 new

Constructor method.

=cut`
		confidence = 0.8
	default:
		// Generic response
		content = "# Generated code based on prompt"
		confidence = 0.7
	}

	return &SamplingResponse{
		Content:    content,
		Confidence: confidence,
		Model:      "mock-model-v1",
		Timestamp:  time.Now(),
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

// extractFunctionName extracts function name from test generation prompts
func extractFunctionName(prompt string) string {
	// Look for "Generate comprehensive Perl tests for function 'function_name'"
	start := strings.Index(prompt, "function '")
	if start == -1 {
		return ""
	}
	start += len("function '")

	end := strings.Index(prompt[start:], "'")
	if end == -1 {
		return ""
	}

	return prompt[start : start+end]
}

// extractFramework extracts test framework from prompts
func extractFramework(prompt string) string {
	if contains(prompt, "Test::More") {
		return "Test::More"
	}
	if contains(prompt, "Test2::V0") {
		return "Test2::V0"
	}
	return ""
}

// CreateSamplingMessage creates an MCP sampling message structure
// Note: This is a placeholder for when MCP sampling is fully implemented
func CreateSamplingMessage(prompt string, options map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"method": "sampling/createMessage",
		"params": map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"role": "user",
					"content": map[string]interface{}{
						"type": "text",
						"text": prompt,
					},
				},
			},
			"modelPreferences": map[string]interface{}{
				"hints": []map[string]string{
					{"name": "claude-3-sonnet"},
				},
			},
			"systemPrompt": "You are a Perl programming expert. Generate clean, typed Perl code following best practices.",
		},
	}
}
