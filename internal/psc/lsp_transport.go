// ABOUTME: JSON-RPC 2.0 transport for the PSC language server.
// ABOUTME: Reads and writes Content-Length-framed messages on stdin/stdout.

package psc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// jsonRPCMessage is the envelope for all JSON-RPC 2.0 messages: requests,
// responses, and notifications.
type jsonRPCMessage struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  json.RawMessage  `json:"params,omitempty"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *jsonRPCError    `json:"error,omitempty"`
}

// jsonRPCError carries the error code and message for a JSON-RPC error response.
type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// transport handles Content-Length-framed JSON-RPC 2.0 I/O over an arbitrary
// reader/writer pair (typically stdin/stdout of the language server process).
type transport struct {
	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex
}

// newTransport wraps r and w in a transport ready for LSP message exchange.
func newTransport(r io.Reader, w io.Writer) *transport {
	return &transport{
		reader: bufio.NewReader(r),
		writer: w,
	}
}

// readMessage reads one Content-Length-framed JSON-RPC message from the
// transport's reader. It parses all header lines up to the blank line
// separator, then reads exactly Content-Length bytes using io.ReadFull so
// that messages larger than the bufio buffer are handled correctly.
func (t *transport) readMessage() (*jsonRPCMessage, error) {
	contentLength := -1

	// Read header lines until we hit the blank line that separates
	// headers from the body.
	for {
		line, err := t.reader.ReadString('\n')
		if err != nil {
			// If we hit EOF before finding a blank separator line and
			// Content-Length was never provided, surface that as the root cause.
			if contentLength < 0 {
				return nil, fmt.Errorf("Content-Length header missing (got EOF before header block)")
			}
			return nil, fmt.Errorf("reading header: %w", err)
		}
		// Trim CRLF / LF.
		line = strings.TrimRight(line, "\r\n")

		// Blank line signals end of headers.
		if line == "" {
			break
		}

		if strings.HasPrefix(line, "Content-Length: ") {
			value := strings.TrimPrefix(line, "Content-Length: ")
			n, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length value %q: %w", value, err)
			}
			contentLength = n
		}
		// Unknown headers (e.g. Content-Type) are silently ignored.
	}

	if contentLength < 0 {
		return nil, fmt.Errorf("Content-Length header missing or invalid")
	}

	// Use io.ReadFull so we read exactly contentLength bytes even when the
	// body exceeds the underlying bufio buffer capacity.
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(t.reader, body); err != nil {
		return nil, fmt.Errorf("reading body (%d bytes): %w", contentLength, err)
	}

	var msg jsonRPCMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("unmarshalling JSON-RPC message: %w", err)
	}
	return &msg, nil
}

// sendResponse writes a JSON-RPC success response with the given id and result.
func (t *transport) sendResponse(id *json.RawMessage, result interface{}) {
	t.send(&jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Result:  mustMarshal(result),
	})
}

// sendError writes a JSON-RPC error response with the given id, code, and message.
func (t *transport) sendError(id *json.RawMessage, code int, message string) {
	t.send(&jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &jsonRPCError{Code: code, Message: message},
	})
}

// sendNotification writes a JSON-RPC notification (no id) with the given
// method and optional params.
func (t *transport) sendNotification(method string, params interface{}) {
	t.send(&jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  method,
		Params:  mustMarshal(params),
	})
}

// send marshals msg, then writes a Content-Length header followed by the body.
// The write is protected by a mutex so multiple goroutines can call send
// concurrently without interleaving output.
func (t *transport) send(msg *jsonRPCMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		// Marshalling our own struct should never fail; log and return.
		return
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))

	t.mu.Lock()
	defer t.mu.Unlock()

	_, _ = io.WriteString(t.writer, header)
	_, _ = t.writer.Write(data)
}

// mustMarshal marshals v to a json.RawMessage. If v is nil or marshalling
// fails, it returns the JSON null literal.
func mustMarshal(v interface{}) json.RawMessage {
	if v == nil {
		return json.RawMessage("null")
	}
	data, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("null")
	}
	return data
}
