// ABOUTME: Real MCP client implementation for JSON-RPC communication with LLMs
// ABOUTME: Replaces mock sampling with actual MCP protocol communication

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"tamarou.com/pvm/internal/log"
)

// MCPClient represents a real MCP client that communicates via JSON-RPC
type MCPClient struct {
	mu             sync.RWMutex
	endpoint       string
	httpClient     *http.Client
	requestID      int64
	logger         *log.Logger
	capabilities   *MCPCapabilities
	authenticated  bool
	sessionID      string
	retryConfig    RetryConfig
	circuitBreaker *CircuitBreaker
}

// MCPCapabilities represents the capabilities negotiated with the MCP server
type MCPCapabilities struct {
	Sampling  bool     `json:"sampling"`
	Resources bool     `json:"resources"`
	Prompts   bool     `json:"prompts"`
	Tools     []string `json:"tools"`
	Models    []string `json:"models"`
	MaxTokens int      `json:"max_tokens"`
	Streaming bool     `json:"streaming"`
}

// RetryConfig configures retry behavior for MCP requests
type RetryConfig struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// CircuitBreaker implements circuit breaker pattern for MCP connections
type CircuitBreaker struct {
	mu           sync.RWMutex
	state        CircuitState
	failures     int
	lastFailTime time.Time
	threshold    int
	timeout      time.Duration
	resetTimeout time.Duration
}

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// MCPRequest represents a JSON-RPC request to an MCP server
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// MCPResponse represents a JSON-RPC response from an MCP server
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an error in an MCP response
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SamplingParams represents parameters for MCP sampling requests
type SamplingParams struct {
	Messages        []MCPMessage           `json:"messages"`
	ModelHints      []string               `json:"model_hints,omitempty"`
	SystemPrompt    string                 `json:"system_prompt,omitempty"`
	MaxTokens       int                    `json:"max_tokens,omitempty"`
	Temperature     float64                `json:"temperature,omitempty"`
	TopP            float64                `json:"top_p,omitempty"`
	StopSequences   []string               `json:"stop_sequences,omitempty"`
	IncludeThinking bool                   `json:"include_thinking,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// MCPMessage represents a message in the MCP format
type MCPMessage struct {
	Role    string     `json:"role"`
	Content MCPContent `json:"content"`
}

// MCPContent represents the content of an MCP message
type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// SamplingResult represents the result of an MCP sampling request
type SamplingResult struct {
	Content    string                 `json:"content"`
	Model      string                 `json:"model"`
	StopReason string                 `json:"stop_reason"`
	Usage      MCPUsage               `json:"usage"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// MCPUsage represents token usage information
type MCPUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// NewMCPClient creates a new MCP client instance
func NewMCPClient(endpoint string, options ...ClientOption) (*MCPClient, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("MCP endpoint cannot be empty")
	}

	client := &MCPClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: log.NewLogger(log.LevelInfo, os.Stderr, "mcp-client"),
		retryConfig: RetryConfig{
			MaxRetries:    3,
			InitialDelay:  100 * time.Millisecond,
			MaxDelay:      5 * time.Second,
			BackoffFactor: 2.0,
		},
		circuitBreaker: &CircuitBreaker{
			threshold:    5,
			timeout:      30 * time.Second,
			resetTimeout: 60 * time.Second,
		},
	}

	// Apply options
	for _, option := range options {
		if err := option(client); err != nil {
			return nil, fmt.Errorf("failed to apply client option: %w", err)
		}
	}

	return client, nil
}

// ClientOption represents a configuration option for the MCP client
type ClientOption func(*MCPClient) error

// WithTimeout sets the HTTP timeout for the client
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *MCPClient) error {
		c.httpClient.Timeout = timeout
		return nil
	}
}

// WithRetryConfig sets the retry configuration
func WithRetryConfig(config RetryConfig) ClientOption {
	return func(c *MCPClient) error {
		c.retryConfig = config
		return nil
	}
}

// WithCircuitBreaker sets the circuit breaker configuration
func WithCircuitBreaker(threshold int, timeout, resetTimeout time.Duration) ClientOption {
	return func(c *MCPClient) error {
		c.circuitBreaker = &CircuitBreaker{
			threshold:    threshold,
			timeout:      timeout,
			resetTimeout: resetTimeout,
		}
		return nil
	}
}

// Connect establishes a connection with the MCP server and negotiates capabilities
func (c *MCPClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Initialize connection with handshake
	initReq := &MCPRequest{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]interface{}{
				"name":    "pvm-mcp-client",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{
				"sampling": map[string]interface{}{},
				"roots": map[string]interface{}{
					"listChanged": true,
				},
			},
		},
	}

	resp, err := c.sendRequest(ctx, initReq)
	if err != nil {
		return fmt.Errorf("failed to initialize MCP connection: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("MCP initialization error: %s", resp.Error.Message)
	}

	// Parse server capabilities
	if err := c.parseCapabilities(resp.Result); err != nil {
		return fmt.Errorf("failed to parse server capabilities: %w", err)
	}

	// Send initialized notification
	notifyReq := &MCPRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
		Params:  map[string]interface{}{},
	}

	_, err = c.sendRequest(ctx, notifyReq)
	if err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	c.authenticated = true
	c.logger.Infof("Successfully connected to MCP server: %s", c.endpoint)

	return nil
}

// Sample sends a sampling request to the MCP server
func (c *MCPClient) Sample(ctx context.Context, params SamplingParams) (*SamplingResult, error) {
	if !c.isConnected() {
		return nil, fmt.Errorf("not connected to MCP server")
	}

	if !c.capabilities.Sampling {
		return nil, fmt.Errorf("server does not support sampling")
	}

	// Check circuit breaker
	if !c.circuitBreaker.Allow() {
		return nil, fmt.Errorf("circuit breaker is open")
	}

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "sampling/createMessage",
		Params:  params,
	}

	resp, err := c.sendRequestWithRetry(ctx, req)
	if err != nil {
		c.circuitBreaker.RecordFailure()
		return nil, fmt.Errorf("sampling request failed: %w", err)
	}

	if resp.Error != nil {
		c.circuitBreaker.RecordFailure()
		return nil, fmt.Errorf("sampling error: %s", resp.Error.Message)
	}

	c.circuitBreaker.RecordSuccess()

	// Parse sampling result
	result, err := c.parseSamplingResult(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sampling result: %w", err)
	}

	return result, nil
}

// Disconnect closes the connection to the MCP server
func (c *MCPClient) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.authenticated {
		return nil
	}

	c.authenticated = false
	c.sessionID = ""
	c.capabilities = nil

	c.logger.Infof("Disconnected from MCP server")
	return nil
}

// IsConnected returns true if the client is connected to the MCP server
func (c *MCPClient) IsConnected() bool {
	return c.isConnected()
}

// GetCapabilities returns the negotiated capabilities
func (c *MCPClient) GetCapabilities() *MCPCapabilities {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.capabilities == nil {
		return nil
	}

	// Return a copy to avoid race conditions
	caps := *c.capabilities
	return &caps
}

// HealthCheck performs a health check on the MCP connection
func (c *MCPClient) HealthCheck(ctx context.Context) error {
	if !c.isConnected() {
		return fmt.Errorf("not connected to MCP server")
	}

	// Check circuit breaker state
	if !c.circuitBreaker.Allow() {
		return fmt.Errorf("circuit breaker is open")
	}

	// Send a lightweight ping request
	pingReq := &MCPRequest{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "ping",
		Params:  map[string]interface{}{},
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.sendRequest(ctx, pingReq)
	if err != nil {
		c.circuitBreaker.RecordFailure()
		return fmt.Errorf("health check ping failed: %w", err)
	}

	if resp.Error != nil {
		c.circuitBreaker.RecordFailure()
		return fmt.Errorf("health check ping error: %s", resp.Error.Message)
	}

	c.circuitBreaker.RecordSuccess()
	return nil
}

// Private methods

func (c *MCPClient) isConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.authenticated
}

func (c *MCPClient) nextRequestID() int64 {
	c.requestID++
	return c.requestID
}

func (c *MCPClient) sendRequest(ctx context.Context, req *MCPRequest) (*MCPResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "pvm-mcp-client/1.0.0")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", httpResp.StatusCode, httpResp.Status)
	}

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var resp MCPResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

func (c *MCPClient) sendRequestWithRetry(ctx context.Context, req *MCPRequest) (*MCPResponse, error) {
	var lastErr error
	delay := c.retryConfig.InitialDelay

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}

			// Exponential backoff
			delay = time.Duration(float64(delay) * c.retryConfig.BackoffFactor)
			if delay > c.retryConfig.MaxDelay {
				delay = c.retryConfig.MaxDelay
			}
		}

		resp, err := c.sendRequest(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		c.logger.Warningf("MCP request failed, retrying attempt %d: %v", attempt+1, err)
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.retryConfig.MaxRetries+1, lastErr)
}

func (c *MCPClient) parseCapabilities(result interface{}) error {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid capabilities format")
	}

	serverInfo, ok := resultMap["serverInfo"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing server info")
	}

	capabilities, ok := resultMap["capabilities"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing capabilities")
	}

	c.capabilities = &MCPCapabilities{}

	// Parse sampling capability
	if _, exists := capabilities["sampling"]; exists {
		c.capabilities.Sampling = true
	}

	// Parse resources capability
	if _, exists := capabilities["resources"]; exists {
		c.capabilities.Resources = true
	}

	// Parse prompts capability
	if _, exists := capabilities["prompts"]; exists {
		c.capabilities.Prompts = true
	}

	// Parse other capabilities as needed
	c.logger.Infof("Parsed MCP server capabilities - server: %v, version: %v, sampling: %v, resources: %v, prompts: %v",
		serverInfo["name"],
		serverInfo["version"],
		c.capabilities.Sampling,
		c.capabilities.Resources,
		c.capabilities.Prompts)

	return nil
}

func (c *MCPClient) parseSamplingResult(result interface{}) (*SamplingResult, error) {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid sampling result format")
	}

	samplingResult := &SamplingResult{}

	// Parse content
	if content, exists := resultMap["content"]; exists {
		if contentStr, ok := content.(string); ok {
			samplingResult.Content = contentStr
		}
	}

	// Parse model
	if model, exists := resultMap["model"]; exists {
		if modelStr, ok := model.(string); ok {
			samplingResult.Model = modelStr
		}
	}

	// Parse stop reason
	if stopReason, exists := resultMap["stopReason"]; exists {
		if stopReasonStr, ok := stopReason.(string); ok {
			samplingResult.StopReason = stopReasonStr
		}
	}

	// Parse usage if available
	if usage, exists := resultMap["usage"].(map[string]interface{}); exists {
		samplingResult.Usage = MCPUsage{}
		if inputTokens, ok := usage["inputTokens"].(float64); ok {
			samplingResult.Usage.InputTokens = int(inputTokens)
		}
		if outputTokens, ok := usage["outputTokens"].(float64); ok {
			samplingResult.Usage.OutputTokens = int(outputTokens)
		}
		samplingResult.Usage.TotalTokens = samplingResult.Usage.InputTokens + samplingResult.Usage.OutputTokens
	}

	return samplingResult, nil
}

// Circuit breaker methods

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = CircuitClosed
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.threshold {
		cb.state = CircuitOpen
	}
}
