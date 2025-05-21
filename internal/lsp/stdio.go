// ABOUTME: Standard I/O connection for LSP server
// ABOUTME: Enables communication over stdin/stdout for editor integration

package lsp

import (
	"io"
	"os"
)

// StdioConnection wraps stdin/stdout for LSP communication
type StdioConnection struct {
	stdin  io.Reader
	stdout io.Writer
}

// NewStdioConnection creates a new stdio connection
func NewStdioConnection() *StdioConnection {
	return &StdioConnection{
		stdin:  os.Stdin,
		stdout: os.Stdout,
	}
}

// Read reads from stdin
func (c *StdioConnection) Read(p []byte) (n int, err error) {
	return c.stdin.Read(p)
}

// Write writes to stdout
func (c *StdioConnection) Write(p []byte) (n int, err error) {
	return c.stdout.Write(p)
}

// Close closes the connection (no-op for stdio)
func (c *StdioConnection) Close() error {
	return nil
}
