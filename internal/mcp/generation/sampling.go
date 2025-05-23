// ABOUTME: MCP sampling client for collaborating with LLMs
// ABOUTME: Provides sampling infrastructure for code generation and fixes

package generation

import (
	"context"
	"fmt"
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
	case contains(request.Prompt, "generate function"):
		content = `sub calculate_sum {
    my (ArrayRef[Int] $numbers) = @_;
    my Int $sum = 0;
    for my $num (@$numbers) {
        $sum += $num;
    }
    return $sum;
}`
		confidence = 0.85
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
