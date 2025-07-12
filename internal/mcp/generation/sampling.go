// ABOUTME: MCP sampling client types and backward compatibility layer
// ABOUTME: Provides type definitions and legacy function compatibility

package generation

import (
	"strings"
	"time"
)

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

// Legacy NewSamplingClient function for backward compatibility
// Creates a mock-mode sampling client by default
func NewSamplingClient(enabled bool) *SamplingClient {
	config := SamplingClientConfig{
		Mode:       SamplingModeMock, // Default to mock for backward compatibility
		Enabled:    enabled,
		MaxRetries: 3,
		Timeout:    30 * time.Second,
	}

	client, err := NewSamplingClientWithConfig(config)
	if err != nil {
		// Fallback to legacy mock implementation
		return NewSamplingClientFromLegacy(enabled)
	}

	return client
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
