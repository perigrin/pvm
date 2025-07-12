// ABOUTME: Real MCP sampling client implementation using JSON-RPC protocol
// ABOUTME: Replaces mock sampling with actual MCP communication

package generation

import (
	"context"
	"fmt"
	"os"
	"time"

	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/mcp/client"
)

// RealSamplingClient implements the SamplingClientInterface using real MCP protocol
type RealSamplingClient struct {
	mcpClient     client.MCPClientInterface
	logger        *log.Logger
	enabled       bool
	defaultConfig SamplingConfig
}

// SamplingConfig holds configuration for sampling requests
type SamplingConfig struct {
	MaxTokens    int     `json:"max_tokens"`
	Temperature  float64 `json:"temperature"`
	TopP         float64 `json:"top_p"`
	SystemPrompt string  `json:"system_prompt"`
}

// SamplingClientInterface defines the interface for sampling clients
type SamplingClientInterface interface {
	Sample(ctx context.Context, prompt string, projectPath string) (*SamplingResponse, error)
	SampleWithOptions(ctx context.Context, request *SamplingRequest) (*SamplingResponse, error)
	BatchSample(ctx context.Context, prompts []string, projectPath string) ([]*SamplingResponse, error)
	IsEnabled() bool
	HealthCheck(ctx context.Context) error
	GetStats() map[string]interface{}
}

// NewRealSamplingClient creates a new real sampling client
func NewRealSamplingClient(mcpClient client.MCPClientInterface, enabled bool) *RealSamplingClient {
	return &RealSamplingClient{
		mcpClient: mcpClient,
		logger:    log.NewLogger(log.LevelInfo, os.Stderr, "real-sampling-client"),
		enabled:   enabled,
		defaultConfig: SamplingConfig{
			MaxTokens:    1000,
			Temperature:  0.7,
			TopP:         0.9,
			SystemPrompt: "You are a Perl programming expert helping to fix code errors and generate typed Perl code.",
		},
	}
}

// Sample sends a sampling request to the MCP server
func (r *RealSamplingClient) Sample(ctx context.Context, prompt string, projectPath string) (*SamplingResponse, error) {
	if !r.enabled {
		return nil, fmt.Errorf("sampling is disabled")
	}

	if !r.mcpClient.IsConnected() {
		if err := r.mcpClient.Connect(ctx); err != nil {
			return nil, fmt.Errorf("failed to connect to MCP server: %w", err)
		}
	}

	// Create MCP sampling parameters
	params := client.SamplingParams{
		Messages: []client.MCPMessage{
			{
				Role: "user",
				Content: client.MCPContent{
					Type: "text",
					Text: prompt,
				},
			},
		},
		SystemPrompt: r.defaultConfig.SystemPrompt,
		MaxTokens:    r.defaultConfig.MaxTokens,
		Temperature:  r.defaultConfig.Temperature,
		TopP:         r.defaultConfig.TopP,
		Metadata: map[string]interface{}{
			"project_path": projectPath,
			"language":     "perl",
			"client":       "pvm",
		},
	}

	// Send sampling request
	result, err := r.mcpClient.Sample(ctx, params)
	if err != nil {
		r.logger.Warningf("Sampling request failed: %v (prompt_length: %d)", err, len(prompt))
		return nil, fmt.Errorf("MCP sampling failed: %w", err)
	}

	// Convert MCP result to sampling response
	response := &SamplingResponse{
		Content:    result.Content,
		Confidence: r.calculateConfidence(result),
		Model:      result.Model,
		Timestamp:  time.Now(),
	}

	r.logger.Debugf("Sampling completed successfully - model: %s, content_length: %d, input_tokens: %d, output_tokens: %d",
		result.Model,
		len(result.Content),
		result.Usage.InputTokens,
		result.Usage.OutputTokens)

	return response, nil
}

// SampleWithOptions sends a sampling request with custom options
func (r *RealSamplingClient) SampleWithOptions(ctx context.Context, request *SamplingRequest) (*SamplingResponse, error) {
	if !r.enabled {
		return nil, fmt.Errorf("sampling is disabled")
	}

	if request.Prompt == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
	}

	if !r.mcpClient.IsConnected() {
		if err := r.mcpClient.Connect(ctx); err != nil {
			return nil, fmt.Errorf("failed to connect to MCP server: %w", err)
		}
	}

	// Set defaults if not provided
	maxTokens := request.MaxTokens
	if maxTokens == 0 {
		maxTokens = r.defaultConfig.MaxTokens
	}

	temperature := request.Temperature
	if temperature == 0 {
		temperature = r.defaultConfig.Temperature
	}

	systemPrompt := request.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = r.defaultConfig.SystemPrompt
	}

	// Create MCP sampling parameters
	metadata := make(map[string]interface{})
	for k, v := range request.Metadata {
		metadata[k] = v
	}

	params := client.SamplingParams{
		Messages: []client.MCPMessage{
			{
				Role: "user",
				Content: client.MCPContent{
					Type: "text",
					Text: request.Prompt,
				},
			},
		},
		SystemPrompt: systemPrompt,
		MaxTokens:    maxTokens,
		Temperature:  temperature,
		TopP:         r.defaultConfig.TopP,
		Metadata:     metadata,
	}

	// Send sampling request
	result, err := r.mcpClient.Sample(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("MCP sampling failed: %w", err)
	}

	// Convert MCP result to sampling response
	response := &SamplingResponse{
		Content:    result.Content,
		Confidence: r.calculateConfidence(result),
		Model:      result.Model,
		Timestamp:  time.Now(),
	}

	return response, nil
}

// BatchSample sends multiple sampling requests in parallel
func (r *RealSamplingClient) BatchSample(ctx context.Context, prompts []string, projectPath string) ([]*SamplingResponse, error) {
	if !r.enabled {
		return nil, fmt.Errorf("sampling is disabled")
	}

	if len(prompts) == 0 {
		return []*SamplingResponse{}, nil
	}

	if !r.mcpClient.IsConnected() {
		if err := r.mcpClient.Connect(ctx); err != nil {
			return nil, fmt.Errorf("failed to connect to MCP server: %w", err)
		}
	}

	responses := make([]*SamplingResponse, len(prompts))
	errors := make([]error, len(prompts))

	// Use a worker pool for parallel processing
	type result struct {
		index    int
		response *SamplingResponse
		err      error
	}

	resultChan := make(chan result, len(prompts))
	semaphore := make(chan struct{}, 5) // Limit concurrent requests

	// Start workers
	for i, prompt := range prompts {
		go func(index int, p string) {
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			resp, err := r.Sample(ctx, p, projectPath)
			resultChan <- result{index: index, response: resp, err: err}
		}(i, prompt)
	}

	// Collect results
	for i := 0; i < len(prompts); i++ {
		select {
		case res := <-resultChan:
			responses[res.index] = res.response
			errors[res.index] = res.err
		case <-ctx.Done():
			return nil, fmt.Errorf("batch sampling cancelled: %w", ctx.Err())
		}
	}

	// Check for errors
	for i, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("failed to sample prompt %d: %w", i, err)
		}
	}

	return responses, nil
}

// IsEnabled returns whether sampling is enabled
func (r *RealSamplingClient) IsEnabled() bool {
	return r.enabled
}

// HealthCheck performs a health check on the sampling client
func (r *RealSamplingClient) HealthCheck(ctx context.Context) error {
	if !r.enabled {
		return fmt.Errorf("sampling is disabled")
	}

	if !r.mcpClient.IsConnected() {
		return fmt.Errorf("MCP client is not connected")
	}

	// Use the MCP client's health check
	if err := r.mcpClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("MCP client health check failed: %w", err)
	}

	// Optionally, perform a lightweight sampling test
	testParams := client.SamplingParams{
		Messages: []client.MCPMessage{
			{
				Role: "user",
				Content: client.MCPContent{
					Type: "text",
					Text: "health check",
				},
			},
		},
		MaxTokens:   10,
		Temperature: 0.1,
	}

	_, err := r.mcpClient.Sample(ctx, testParams)
	if err != nil {
		return fmt.Errorf("sampling health check failed: %w", err)
	}

	return nil
}

// calculateConfidence estimates confidence based on MCP result
func (r *RealSamplingClient) calculateConfidence(result *client.SamplingResult) float64 {
	// Base confidence
	confidence := 0.7

	// Adjust based on stop reason
	switch result.StopReason {
	case "end_turn":
		confidence += 0.2
	case "max_tokens":
		confidence -= 0.1
	case "stop_sequence":
		confidence += 0.1
	}

	// Adjust based on content length (reasonable content gets higher confidence)
	contentLength := len(result.Content)
	if contentLength > 10 && contentLength < 1000 {
		confidence += 0.1
	} else if contentLength > 1000 {
		confidence -= 0.1
	}

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	} else if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// UpdateDefaultConfig updates the default configuration for sampling
func (r *RealSamplingClient) UpdateDefaultConfig(config SamplingConfig) {
	r.defaultConfig = config
}

// GetDefaultConfig returns the current default configuration
func (r *RealSamplingClient) GetDefaultConfig() SamplingConfig {
	return r.defaultConfig
}

// GetStats returns statistics about the real sampling client
func (r *RealSamplingClient) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":       r.enabled,
		"mode":          "real",
		"max_tokens":    r.defaultConfig.MaxTokens,
		"temperature":   r.defaultConfig.Temperature,
		"top_p":         r.defaultConfig.TopP,
		"system_prompt": r.defaultConfig.SystemPrompt,
	}
}
