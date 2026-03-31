// ABOUTME: Tests for the JSON-RPC 2.0 transport layer.
// ABOUTME: Verifies Content-Length framing, message parsing, and error handling.

package psc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readResponse is a shared test helper that reads a single JSON-RPC message
// from the given buffer using the transport framing protocol. Other test files
// in this package may use this helper.
func readResponse(t *testing.T, buf *bytes.Buffer) *jsonRPCMessage {
	t.Helper()
	tr := newTransport(buf, io.Discard)
	msg, err := tr.readMessage()
	require.NoError(t, err)
	return msg
}

func TestTransportReadMessage(t *testing.T) {
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	frame := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)

	tr := newTransport(strings.NewReader(frame), io.Discard)
	msg, err := tr.readMessage()
	require.NoError(t, err)
	require.NotNil(t, msg)

	assert.Equal(t, "2.0", msg.JSONRPC)
	assert.Equal(t, "initialize", msg.Method)
	require.NotNil(t, msg.ID)

	var id int
	require.NoError(t, json.Unmarshal(*msg.ID, &id))
	assert.Equal(t, 1, id)
}

func TestTransportWriteFraming(t *testing.T) {
	var buf bytes.Buffer
	rawID := json.RawMessage(`1`)
	tr := newTransport(strings.NewReader(""), &buf)

	tr.sendResponse(&rawID, map[string]string{"greeting": "hello"})

	output := buf.String()
	assert.True(t, strings.HasPrefix(output, "Content-Length: "), "output must start with Content-Length header")

	// Re-parse the written frame using a fresh transport.
	msg := readResponse(t, &buf)
	require.NotNil(t, msg)
	assert.Equal(t, "2.0", msg.JSONRPC)
	require.NotNil(t, msg.ID)

	var id int
	require.NoError(t, json.Unmarshal(*msg.ID, &id))
	assert.Equal(t, 1, id)

	var result map[string]string
	require.NoError(t, json.Unmarshal(msg.Result, &result))
	assert.Equal(t, "hello", result["greeting"])
}

func TestTransportMalformedHeader(t *testing.T) {
	// A frame with no Content-Length header at all — just a bare JSON body.
	frame := `{"jsonrpc":"2.0","method":"initialize"}`
	tr := newTransport(strings.NewReader(frame), io.Discard)

	_, err := tr.readMessage()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Content-Length")
}

func TestTransportContentLengthLimit(t *testing.T) {
	// A Content-Length value exceeding the 50MB cap must be rejected before
	// any allocation attempt.
	frame := fmt.Sprintf("Content-Length: %d\r\n\r\n", maxContentLength+1)
	tr := newTransport(strings.NewReader(frame), io.Discard)

	_, err := tr.readMessage()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum")
}

func TestTransportLargeBody(t *testing.T) {
	// Build a body larger than the typical bufio default buffer (4096 bytes).
	largeValue := strings.Repeat("x", 9000)
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"ping","params":{"data":"%s"}}`, largeValue)
	frame := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)

	tr := newTransport(strings.NewReader(frame), io.Discard)
	msg, err := tr.readMessage()
	require.NoError(t, err)
	require.NotNil(t, msg)

	assert.Equal(t, "2.0", msg.JSONRPC)
	assert.Equal(t, "ping", msg.Method)

	var params struct {
		Data string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(msg.Params, &params))
	assert.Equal(t, largeValue, params.Data)
}
