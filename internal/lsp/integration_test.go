// ABOUTME: Integration tests for LSP server
// ABOUTME: Tests end-to-end functionality of the language server

package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"
)

// mockConn implements io.ReadWriteCloser for testing
type mockConn struct {
	input  *bytes.Buffer
	output *bytes.Buffer
	closed bool
}

func newMockConn() *mockConn {
	return &mockConn{
		input:  &bytes.Buffer{},
		output: &bytes.Buffer{},
	}
}

func (m *mockConn) Read(p []byte) (n int, err error) {
	if m.closed {
		return 0, io.EOF
	}

	// If buffer is empty, wait for data instead of returning EOF immediately
	for m.input.Len() == 0 && !m.closed {
		time.Sleep(10 * time.Millisecond)
	}

	if m.closed {
		return 0, io.EOF
	}

	return m.input.Read(p)
}

func (m *mockConn) Write(p []byte) (n int, err error) {
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	return m.output.Write(p)
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

// writeMessage writes a JSON-RPC message to the connection
func (m *mockConn) writeMessage(id interface{}, method string, params interface{}) error {
	msg := JSONRPCMessage{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Method:  method,
		Params:  nil,
	}

	if params != nil {
		paramBytes, err := json.Marshal(params)
		if err != nil {
			return err
		}
		msg.Params = json.RawMessage(paramBytes)
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(msgBytes))
	m.input.WriteString(header)
	m.input.Write(msgBytes)

	return nil
}

// readResponse reads a JSON-RPC response from the connection
func (m *mockConn) readResponse() (*JSONRPCResponse, error) {
	// Wait for response data to be available
	timeout := time.After(1 * time.Second)
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for response")
		default:
			if m.output.Len() > 0 {
				goto parseResponse
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

parseResponse:
	// Use buffered reader for proper parsing
	reader := bufio.NewReader(m.output)

	// Parse headers
	var contentLength int
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = line[:len(line)-1] // Remove \n
		if line == "\r" || line == "" {
			break
		}

		if _, err := fmt.Sscanf(line, "Content-Length: %d\r", &contentLength); err == nil {
			// Found content length
		}
	}

	// Read body
	body := make([]byte, contentLength)
	n, err := io.ReadFull(reader, body)
	if err != nil || n != contentLength {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func TestLSPServerIntegration(t *testing.T) {
	conn := newMockConn()
	server, err := NewServer(conn)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server in background
	serverDone := make(chan error)
	go func() {
		serverDone <- server.Start()
	}()

	// Test initialize request
	t.Run("Initialize", func(t *testing.T) {
		processID := 12345
		params := InitializeParams{
			ProcessID: &processID,
			ClientInfo: &ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
			Capabilities: ClientCapabilities{},
		}

		if err := conn.writeMessage(1, "initialize", params); err != nil {
			t.Fatalf("Failed to write initialize request: %v", err)
		}

		// Give server time to process
		// In real implementation, we'd wait for response

		// Read response
		resp, err := conn.readResponse()
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		if resp.Error != nil {
			t.Fatalf("Initialize failed: %v", resp.Error.Message)
		}

		// Verify capabilities
		var result InitializeResult
		resultBytes, err := json.Marshal(resp.Result)
		if err != nil {
			t.Fatalf("Failed to marshal result: %v", err)
		}
		if err := json.Unmarshal(resultBytes, &result); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if !result.Capabilities.HoverProvider {
			t.Error("Expected hover provider to be enabled")
		}

		if !result.Capabilities.DefinitionProvider {
			t.Error("Expected definition provider to be enabled")
		}

		if !result.Capabilities.ReferencesProvider {
			t.Error("Expected references provider to be enabled")
		}

		if !result.Capabilities.DocumentFormattingProvider {
			t.Error("Expected formatting provider to be enabled")
		}

		if !result.Capabilities.CodeActionProvider {
			t.Error("Expected code action provider to be enabled")
		}
	})

	// Send shutdown
	conn.writeMessage(2, "shutdown", nil)

	// Send exit
	conn.writeMessage(nil, "exit", nil)

	// Wait for server to finish
	select {
	case err := <-serverDone:
		if err != nil {
			t.Errorf("Server exited with error: %v", err)
		}
	default:
		// Server still running, force close
		conn.Close()
	}
}
