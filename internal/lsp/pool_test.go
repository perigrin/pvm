// ABOUTME: Tests for LSP object pooling functionality and request lifecycle management
// ABOUTME: Ensures efficient memory allocation and proper cleanup for LSP protocol objects

package lsp

import (
	"fmt"
	"testing"
	"time"
)

func TestLSPPoolManager_NewLSPPoolManager(t *testing.T) {
	hooks := LSPPoolHooks{
		OnPoolWarming: func(poolType string) {
			t.Logf("Pool warming: %s", poolType)
		},
	}
	manager := NewLSPPoolManager(hooks)

	if manager == nil {
		t.Fatal("Expected non-nil LSPPoolManager")
	}
	if manager.hooks.OnPoolWarming == nil {
		t.Error("Expected hooks to be set")
	}
}

func TestLSPPoolManager_RequestScopedPooling(t *testing.T) {
	manager := NewLSPPoolManager(LSPPoolHooks{})

	// Test request-scoped pooling
	requestID := "test-request"
	scope := manager.StartRequest(requestID, "textDocument/completion")
	if scope == nil {
		t.Fatal("Expected non-nil RequestScopedPool")
	}

	// Create some pooled objects
	msg := manager.NewJSONRPCMessage(requestID)
	if msg == nil {
		t.Fatal("Expected non-nil JSONRPCMessage")
	}
	if msg.JSONRPC != "2.0" {
		t.Error("Expected JSONRPC version to be set")
	}

	resp := manager.NewJSONRPCResponse(requestID)
	if resp == nil {
		t.Fatal("Expected non-nil JSONRPCResponse")
	}
	if resp.JSONRPC != "2.0" {
		t.Error("Expected JSONRPC version to be set")
	}

	pos := manager.NewPosition(requestID, 10, 5)
	if pos == nil {
		t.Fatal("Expected non-nil Position")
	}
	if pos.Line != 10 || pos.Character != 5 {
		t.Error("Expected position values to be set correctly")
	}

	// End request should clean up all objects
	manager.EndRequest(requestID)

	// Verify statistics
	if manager.RequestCount() == 0 {
		t.Error("Expected request count to be incremented")
	}
}

func TestLSPPoolManager_CompletionPooling(t *testing.T) {
	manager := NewLSPPoolManager(LSPPoolHooks{})

	requestID := "completion-test"
	_ = manager.StartRequest(requestID, "textDocument/completion")
	defer manager.EndRequest(requestID)

	// Test completion item creation
	item := manager.NewCompletionItem(requestID, "testFunction", "A test function")
	if item == nil {
		t.Fatal("Expected non-nil CompletionItem")
	}
	if item.Label != "testFunction" {
		t.Error("Expected label to be set correctly")
	}
	if item.Detail != "A test function" {
		t.Error("Expected detail to be set correctly")
	}

	// Test completion list creation
	list := manager.NewCompletionList(requestID, false)
	if list == nil {
		t.Fatal("Expected non-nil CompletionList")
	}
	if list.IsIncomplete != false {
		t.Error("Expected IsIncomplete to be set correctly")
	}
	if list.Items == nil {
		t.Error("Expected Items slice to be initialized")
	}

	// Add item to list
	list.Items = append(list.Items, *item)
	if len(list.Items) != 1 {
		t.Error("Expected one item in completion list")
	}
}

func TestLSPPoolManager_DiagnosticPooling(t *testing.T) {
	manager := NewLSPPoolManager(LSPPoolHooks{})

	requestID := "diagnostic-test"
	_ = manager.StartRequest(requestID, "textDocument/publishDiagnostics")
	defer manager.EndRequest(requestID)

	// Test range creation
	start := manager.NewPosition(requestID, 0, 0)
	end := manager.NewPosition(requestID, 0, 10)
	rng := manager.NewRange(requestID, *start, *end)
	if rng == nil {
		t.Fatal("Expected non-nil Range")
	}

	// Test diagnostic creation
	severity := DiagnosticSeverityError
	diag := manager.NewDiagnostic(requestID, *rng, "Test error", &severity)
	if diag == nil {
		t.Fatal("Expected non-nil Diagnostic")
	}
	if diag.Message != "Test error" {
		t.Error("Expected message to be set correctly")
	}
	if diag.Severity == nil || *diag.Severity != DiagnosticSeverityError {
		t.Error("Expected severity to be set correctly")
	}

	// Test publish diagnostics params creation
	params := manager.NewPublishDiagnosticsParams(requestID, "file:///test.pl")
	if params == nil {
		t.Fatal("Expected non-nil PublishDiagnosticsParams")
	}
	if params.URI != "file:///test.pl" {
		t.Error("Expected URI to be set correctly")
	}
	if params.Diagnostics == nil {
		t.Error("Expected Diagnostics slice to be initialized")
	}

	// Add diagnostic to params
	params.Diagnostics = append(params.Diagnostics, *diag)
	if len(params.Diagnostics) != 1 {
		t.Error("Expected one diagnostic in params")
	}
}

func TestLSPPoolManager_Statistics(t *testing.T) {
	manager := NewLSPPoolManager(LSPPoolHooks{})

	// Get baseline stats after initialization/warming
	baselineStats := manager.GetDetailedStats()

	// Create multiple requests to test statistics
	for i := 0; i < 5; i++ {
		requestID := fmt.Sprintf("request-%d", i)
		_ = manager.StartRequest(requestID, "test")

		// Create some objects
		_ = manager.NewJSONRPCMessage(requestID)
		_ = manager.NewCompletionItem(requestID, "test", "test")
		_ = manager.NewDiagnostic(requestID, Range{}, "test", nil)

		manager.EndRequest(requestID)
	}

	stats := manager.GetDetailedStats()
	expectedRequests := baselineStats.RequestCount + 5
	expectedCompletions := baselineStats.CompletionCount + 5
	expectedDiagnostics := baselineStats.DiagnosticCount + 5

	if stats.RequestCount != expectedRequests {
		t.Errorf("Expected %d requests, got %d", expectedRequests, stats.RequestCount)
	}
	if stats.CompletionCount != expectedCompletions {
		t.Errorf("Expected %d completions, got %d", expectedCompletions, stats.CompletionCount)
	}
	if stats.DiagnosticCount != expectedDiagnostics {
		t.Errorf("Expected %d diagnostics, got %d", expectedDiagnostics, stats.DiagnosticCount)
	}

	efficiency := manager.PoolEfficiency()
	if efficiency < 0 || efficiency > 100 {
		t.Errorf("Expected efficiency between 0-100, got %f", efficiency)
	}
}

func TestLSPPoolManager_Hooks(t *testing.T) {
	var warmingCalled bool
	var requestStartCalled bool
	var requestEndCalled bool
	var objectCreateCalled bool

	hooks := LSPPoolHooks{
		OnPoolWarming: func(poolType string) {
			warmingCalled = true
		},
		OnRequestStart: func(requestID, method string) {
			requestStartCalled = true
		},
		OnRequestEnd: func(requestID string, duration int64) {
			requestEndCalled = true
		},
		OnObjectCreate: func(objectType string) {
			objectCreateCalled = true
		},
	}

	manager := NewLSPPoolManager(hooks)

	// Pool warming should have been called during initialization
	if !warmingCalled {
		t.Error("Expected OnPoolWarming to be called")
	}

	// Test request hooks
	requestID := "hook-test"
	_ = manager.StartRequest(requestID, "test")
	if !requestStartCalled {
		t.Error("Expected OnRequestStart to be called")
	}

	// Create an object to trigger OnObjectCreate
	_ = manager.NewJSONRPCMessage(requestID)
	if !objectCreateCalled {
		t.Error("Expected OnObjectCreate to be called")
	}

	// Small delay to ensure duration is measurable
	time.Sleep(1 * time.Millisecond)

	manager.EndRequest(requestID)
	if !requestEndCalled {
		t.Error("Expected OnRequestEnd to be called")
	}
}

func TestLSPPoolManager_GlobalInstance(t *testing.T) {
	// Test global instance creation
	global1 := GlobalLSPPoolManager()
	global2 := GlobalLSPPoolManager()

	if global1 != global2 {
		t.Error("Expected global instances to be the same")
	}

	// Test setting global hooks
	SetGlobalLSPPoolHooks(LSPPoolHooks{
		OnPoolWarming: func(poolType string) {
			// Hook implementation for testing
		},
	})

	if global1.hooks.OnPoolWarming == nil {
		t.Error("Expected global hooks to be set")
	}
}
