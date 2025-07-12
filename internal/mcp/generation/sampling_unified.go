// ABOUTME: Unified sampling client that supports both real MCP and mock implementations
// ABOUTME: Provides factory methods and interface-based switching between implementations

package generation

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/mcp/client"
)

// SamplingClient provides MCP sampling capabilities with pluggable backends
type SamplingClient struct {
	implementation SamplingClientInterface
	logger         *log.Logger
	enabled        bool
	useMockClient  bool
}

// SamplingMode represents the mode of operation for the sampling client
type SamplingMode string

const (
	SamplingModeReal SamplingMode = "real"
	SamplingModeMock SamplingMode = "mock"
)

// SamplingClientConfig holds configuration for creating sampling clients
type SamplingClientConfig struct {
	Mode          SamplingMode      `json:"mode"`
	Enabled       bool              `json:"enabled"`
	MCPEndpoint   string            `json:"mcp_endpoint"`
	MaxRetries    int               `json:"max_retries"`
	Timeout       time.Duration     `json:"timeout"`
	MockResponses map[string]string `json:"mock_responses,omitempty"`
}

// NewSamplingClientWithConfig creates a new sampling client with the specified configuration
func NewSamplingClientWithConfig(config SamplingClientConfig) (*SamplingClient, error) {
	logger := log.NewLogger(log.LevelInfo, os.Stderr, "sampling-client")

	var implementation SamplingClientInterface

	switch config.Mode {
	case SamplingModeReal:
		if config.MCPEndpoint == "" {
			return nil, fmt.Errorf("MCP endpoint is required for real mode")
		}

		// Create real MCP client
		mcpClient, err := client.NewMCPClient(config.MCPEndpoint,
			client.WithTimeout(config.Timeout),
			client.WithRetryConfig(client.RetryConfig{
				MaxRetries:    config.MaxRetries,
				InitialDelay:  100 * time.Millisecond,
				MaxDelay:      5 * time.Second,
				BackoffFactor: 2.0,
			}))
		if err != nil {
			return nil, fmt.Errorf("failed to create MCP client: %w", err)
		}

		implementation = NewRealSamplingClient(mcpClient, config.Enabled)
		logger.Infof("Created real MCP sampling client: %s", config.MCPEndpoint)

	case SamplingModeMock:
		mockClient := client.NewMockMCPClient()

		// Set custom responses if provided
		for prompt, response := range config.MockResponses {
			mockClient.SetMockResponse(prompt, &client.SamplingResult{
				Content: response,
				Model:   "mock-model-v1",
			})
		}

		implementation = NewMockSamplingClient(mockClient, config.Enabled)
		logger.Infof("Created mock MCP sampling client")

	default:
		return nil, fmt.Errorf("unsupported sampling mode: %s", config.Mode)
	}

	return &SamplingClient{
		implementation: implementation,
		logger:         logger,
		enabled:        config.Enabled,
		useMockClient:  config.Mode == SamplingModeMock,
	}, nil
}

// NewSamplingClientFromLegacy creates a sampling client from the legacy boolean parameter
// This maintains backward compatibility with existing code
func NewSamplingClientFromLegacy(enabled bool) *SamplingClient {
	config := SamplingClientConfig{
		Mode:       SamplingModeMock, // Default to mock for backward compatibility
		Enabled:    enabled,
		MaxRetries: 3,
		Timeout:    30 * time.Second,
	}

	samplingClient, err := NewSamplingClientWithConfig(config)
	if err != nil {
		// Fallback to a basic mock implementation
		mockClient := client.NewMockMCPClient()
		return &SamplingClient{
			implementation: NewMockSamplingClient(mockClient, enabled),
			logger:         log.NewLogger(log.LevelInfo, os.Stderr, "sampling-client"),
			enabled:        enabled,
			useMockClient:  true,
		}
	}

	return samplingClient
}

// Sample sends a sampling request
func (c *SamplingClient) Sample(ctx context.Context, prompt string, projectPath string) (*SamplingResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("sampling is disabled")
	}

	start := time.Now()
	response, err := c.implementation.Sample(ctx, prompt, projectPath)
	duration := time.Since(start)

	if err != nil {
		c.logger.Warningf("Sampling failed: %v (prompt_length: %d, duration: %v, mock_mode: %v)",
			err,
			len(prompt),
			duration,
			c.useMockClient)
		return nil, err
	}

	c.logger.Debugf("Sampling completed - duration: %v, content_length: %d, confidence: %f, mock_mode: %v",
		duration,
		len(response.Content),
		response.Confidence,
		c.useMockClient)

	return response, nil
}

// SampleWithOptions sends a sampling request with custom options
func (c *SamplingClient) SampleWithOptions(ctx context.Context, request *SamplingRequest) (*SamplingResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("sampling is disabled")
	}

	if request == nil {
		return nil, fmt.Errorf("sampling request cannot be nil")
	}

	return c.implementation.SampleWithOptions(ctx, request)
}

// BatchSample sends multiple sampling requests in parallel
func (c *SamplingClient) BatchSample(ctx context.Context, prompts []string, projectPath string) ([]*SamplingResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("sampling is disabled")
	}

	if len(prompts) == 0 {
		return []*SamplingResponse{}, nil
	}

	start := time.Now()
	responses, err := c.implementation.BatchSample(ctx, prompts, projectPath)
	duration := time.Since(start)

	c.logger.Debugf("Batch sampling completed - prompt_count: %d, duration: %v, success: %v, mock_mode: %v",
		len(prompts),
		duration,
		err == nil,
		c.useMockClient)

	return responses, err
}

// IsEnabled returns whether sampling is enabled
func (c *SamplingClient) IsEnabled() bool {
	return c.enabled
}

// SetEnabled enables or disables sampling
func (c *SamplingClient) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// GetMode returns the current sampling mode
func (c *SamplingClient) GetMode() SamplingMode {
	if c.useMockClient {
		return SamplingModeMock
	}
	return SamplingModeReal
}

// HealthCheck performs a health check on the sampling client
func (c *SamplingClient) HealthCheck(ctx context.Context) error {
	if !c.enabled {
		return fmt.Errorf("sampling is disabled")
	}

	return c.implementation.HealthCheck(ctx)
}

// GetStats returns statistics about the sampling client
func (c *SamplingClient) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled":   c.enabled,
		"mode":      string(c.GetMode()),
		"mock_mode": c.useMockClient,
	}

	// Add implementation-specific stats
	implStats := c.implementation.GetStats()
	for k, v := range implStats {
		stats[k] = v
	}

	return stats
}

// MockSamplingClient wraps the mock client to implement SamplingClientInterface
type MockSamplingClient struct {
	mockClient client.MCPClientInterface
	enabled    bool
}

// NewMockSamplingClient creates a new mock sampling client
func NewMockSamplingClient(mockClient client.MCPClientInterface, enabled bool) *MockSamplingClient {
	return &MockSamplingClient{
		mockClient: mockClient,
		enabled:    enabled,
	}
}

// Sample implements the SamplingClientInterface for mock client
func (m *MockSamplingClient) Sample(ctx context.Context, prompt string, projectPath string) (*SamplingResponse, error) {
	if !m.enabled {
		return nil, fmt.Errorf("sampling is disabled")
	}

	if prompt == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
	}

	// Ensure the mock client is connected
	if !m.mockClient.IsConnected() {
		if err := m.mockClient.Connect(ctx); err != nil {
			return nil, fmt.Errorf("failed to connect mock client: %w", err)
		}
	}

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
		Metadata: map[string]interface{}{
			"project_path": projectPath,
			"language":     "perl",
		},
	}

	result, err := m.mockClient.Sample(ctx, params)
	if err != nil {
		return nil, err
	}

	return &SamplingResponse{
		Content:    result.Content,
		Confidence: m.calculateMockConfidence(result.Content),
		Model:      result.Model,
		Timestamp:  time.Now(),
	}, nil
}

// SampleWithOptions implements the SamplingClientInterface for mock client
func (m *MockSamplingClient) SampleWithOptions(ctx context.Context, request *SamplingRequest) (*SamplingResponse, error) {
	if !m.enabled {
		return nil, fmt.Errorf("sampling is disabled")
	}

	if request.Prompt == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
	}

	// Ensure the mock client is connected
	if !m.mockClient.IsConnected() {
		if err := m.mockClient.Connect(ctx); err != nil {
			return nil, fmt.Errorf("failed to connect mock client: %w", err)
		}
	}

	projectPath := ""
	if path, ok := request.Metadata["project_path"]; ok {
		projectPath = path
	}
	return m.Sample(ctx, request.Prompt, projectPath)
}

// BatchSample implements the SamplingClientInterface for mock client
func (m *MockSamplingClient) BatchSample(ctx context.Context, prompts []string, projectPath string) ([]*SamplingResponse, error) {
	if !m.enabled {
		return nil, fmt.Errorf("sampling is disabled")
	}

	responses := make([]*SamplingResponse, len(prompts))
	for i, prompt := range prompts {
		response, err := m.Sample(ctx, prompt, projectPath)
		if err != nil {
			return nil, fmt.Errorf("failed to sample prompt %d: %w", i, err)
		}
		responses[i] = response
	}

	return responses, nil
}

// IsEnabled returns whether mock sampling is enabled
func (m *MockSamplingClient) IsEnabled() bool {
	return m.enabled
}

// HealthCheck implements the SamplingClientInterface for mock client
func (m *MockSamplingClient) HealthCheck(ctx context.Context) error {
	if !m.enabled {
		return fmt.Errorf("sampling is disabled")
	}

	// Ensure the mock client is connected
	if !m.mockClient.IsConnected() {
		if err := m.mockClient.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect mock client: %w", err)
		}
	}

	return m.mockClient.HealthCheck(ctx)
}

// GetStats implements the SamplingClientInterface for mock client
func (m *MockSamplingClient) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled": m.enabled,
		"mode":    "mock",
	}
}

// calculateMockConfidence calculates confidence for mock responses
func (m *MockSamplingClient) calculateMockConfidence(content string) float64 {
	// Simple confidence calculation based on content characteristics
	switch {
	case strings.Contains(content, "missing sigil"), strings.Contains(content, "type mismatch"):
		return 0.95
	case strings.Contains(content, "generate function"), strings.Contains(content, "sub "):
		return 0.85
	case strings.Contains(content, "test"), strings.Contains(content, "ok("):
		return 0.85
	case strings.Contains(content, "class"), strings.Contains(content, "package"):
		return 0.85
	case strings.Contains(content, "refactor"):
		return 0.8
	case strings.Contains(content, "documentation"), strings.Contains(content, "=head"):
		return 0.8
	default:
		return 0.7
	}
}
