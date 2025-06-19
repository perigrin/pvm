// ABOUTME: Comprehensive tests for PVM UI framework output functionality
// ABOUTME: Tests all output methods, styling, and configuration options

package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewOutput(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *UIContext
		expected bool
	}{
		{
			name:     "nil context creates default",
			ctx:      nil,
			expected: true,
		},
		{
			name: "custom context",
			ctx: &UIContext{
				Writer:    &bytes.Buffer{},
				ColorMode: ColorNever,
				Quiet:     true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewOutput(tt.ctx)
			if (output != nil) != tt.expected {
				t.Errorf("NewOutput() = %v, expected %v", output != nil, tt.expected)
			}
			if output != nil && output.context == nil {
				t.Error("NewOutput() created output with nil context")
			}
		})
	}
}

func TestNewDefaultOutput(t *testing.T) {
	output := NewDefaultOutput()
	if output == nil {
		t.Error("NewDefaultOutput() returned nil")
	}
	if output.context == nil {
		t.Error("NewDefaultOutput() created output with nil context")
	}
}

func TestBasicOutputMethods(t *testing.T) {
	var buf bytes.Buffer
	ctx := &UIContext{
		Writer:    &buf,
		ColorMode: ColorNever,
		Quiet:     false,
		Verbose:   true,
	}
	output := NewOutput(ctx)

	tests := []struct {
		name     string
		method   func()
		expected string
	}{
		{
			name: "Info",
			method: func() {
				output.Info("test info message")
			},
			expected: "test info message",
		},
		{
			name: "Success",
			method: func() {
				output.Success("test success message")
			},
			expected: "test success message",
		},
		{
			name: "Warning",
			method: func() {
				output.Warning("test warning message")
			},
			expected: "test warning message",
		},
		{
			name: "Error",
			method: func() {
				output.Error("test error message")
			},
			expected: "test error message",
		},
		{
			name: "Debug",
			method: func() {
				output.Debug("test debug message")
			},
			expected: "test debug message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.method()
			result := buf.String()
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormattedOutput(t *testing.T) {
	var buf bytes.Buffer
	ctx := &UIContext{
		Writer:    &buf,
		ColorMode: ColorNever,
		Quiet:     false,
	}
	output := NewOutput(ctx)

	tests := []struct {
		name     string
		method   func()
		expected string
	}{
		{
			name: "Info with format",
			method: func() {
				output.Info("test %s %d", "message", 42)
			},
			expected: "test message 42",
		},
		{
			name: "Success with format",
			method: func() {
				output.Success("completed %d tasks", 5)
			},
			expected: "completed 5 tasks",
		},
		{
			name: "Printf",
			method: func() {
				output.Printf("formatted %s", "output")
			},
			expected: "formatted output",
		},
		{
			name: "Println",
			method: func() {
				output.Println("line", "output")
			},
			expected: "line output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.method()
			result := buf.String()
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestQuietMode(t *testing.T) {
	var buf bytes.Buffer
	ctx := &UIContext{
		Writer:    &buf,
		ColorMode: ColorNever,
		Quiet:     true,
		Verbose:   true,
	}
	output := NewOutput(ctx)

	// Test that quiet mode suppresses most output
	output.Info("info message")
	output.Success("success message")
	output.Warning("warning message")
	output.Debug("debug message")
	output.Printf("printf message")
	output.Println("println message")

	// Error should still show in quiet mode
	output.Error("error message")

	result := buf.String()

	// Should contain error message
	if !strings.Contains(result, "error message") {
		t.Error("Error message should be shown in quiet mode")
	}

	// Should not contain other messages
	if strings.Contains(result, "info message") ||
		strings.Contains(result, "success message") ||
		strings.Contains(result, "warning message") ||
		strings.Contains(result, "debug message") ||
		strings.Contains(result, "printf message") ||
		strings.Contains(result, "println message") {
		t.Error("Non-error messages should be suppressed in quiet mode")
	}
}

func TestVerboseMode(t *testing.T) {
	var buf bytes.Buffer

	// Test with verbose disabled
	ctx := &UIContext{
		Writer:    &buf,
		ColorMode: ColorNever,
		Quiet:     false,
		Verbose:   false,
	}
	output := NewOutput(ctx)

	output.Debug("debug message")
	result := buf.String()

	if strings.Contains(result, "debug message") {
		t.Error("Debug message should not be shown when verbose is disabled")
	}

	// Test with verbose enabled
	buf.Reset()
	ctx.Verbose = true
	output = NewOutput(ctx)

	output.Debug("debug message")
	result = buf.String()

	if !strings.Contains(result, "debug message") {
		t.Error("Debug message should be shown when verbose is enabled")
	}
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	ctx := &UIContext{
		Writer:    &buf,
		ColorMode: ColorNever,
		Quiet:     false,
	}
	output := NewOutput(ctx)

	headers := []string{"Name", "Version", "Status"}
	rows := [][]string{
		{"perl", "5.38.0", "active"},
		{"cpanm", "1.7047", "installed"},
	}

	output.Table(headers, rows)
	result := buf.String()

	// Check that table contains expected content
	expectedContent := []string{"Name", "Version", "Status", "perl", "5.38.0", "active", "cpanm", "1.7047", "installed"}
	for _, content := range expectedContent {
		if !strings.Contains(result, content) {
			t.Errorf("Table output should contain %q, got %q", content, result)
		}
	}
}

func TestTableWithOptions(t *testing.T) {
	var buf bytes.Buffer
	ctx := &UIContext{
		Writer:    &buf,
		ColorMode: ColorNever,
		Quiet:     false,
	}
	output := NewOutput(ctx)

	opts := TableOptions{
		Headers: []string{"Command", "Description"},
		Rows: [][]string{
			{"pvm install", "Install Perl version"},
			{"pvm use", "Switch Perl version"},
		},
		Title:       "Available Commands",
		ShowBorders: true,
	}

	output.TableWithOptions(opts)
	result := buf.String()

	if !strings.Contains(result, "Available Commands") {
		t.Error("Table should contain title")
	}
	if !strings.Contains(result, "pvm install") {
		t.Error("Table should contain row data")
	}
}

func TestList(t *testing.T) {
	var buf bytes.Buffer
	ctx := &UIContext{
		Writer:    &buf,
		ColorMode: ColorNever,
		Quiet:     false,
	}
	output := NewOutput(ctx)

	items := []string{"First item", "Second item", "Third item"}
	output.List(items)
	result := buf.String()

	for _, item := range items {
		if !strings.Contains(result, item) {
			t.Errorf("List should contain %q, got %q", item, result)
		}
	}
}

func TestListWithOptions(t *testing.T) {
	var buf bytes.Buffer
	ctx := &UIContext{
		Writer:    &buf,
		ColorMode: ColorNever,
		Quiet:     false,
	}
	output := NewOutput(ctx)

	opts := ListOptions{
		Items:    []string{"Task one", "Task two"},
		Title:    "Todo List",
		Numbered: true,
	}

	output.ListWithOptions(opts)
	result := buf.String()

	if !strings.Contains(result, "Todo List") {
		t.Error("List should contain title")
	}
	if !strings.Contains(result, "Task one") {
		t.Error("List should contain items")
	}
}

func TestKeyValue(t *testing.T) {
	var buf bytes.Buffer
	ctx := &UIContext{
		Writer:    &buf,
		ColorMode: ColorNever,
		Quiet:     false,
	}
	output := NewOutput(ctx)

	pairs := map[string]string{
		"Version": "5.38.0",
		"Status":  "active",
		"Path":    "/usr/local/bin/perl",
	}

	output.KeyValue(pairs)
	result := buf.String()

	for key, value := range pairs {
		if !strings.Contains(result, key) {
			t.Errorf("KeyValue should contain key %q", key)
		}
		if !strings.Contains(result, value) {
			t.Errorf("KeyValue should contain value %q", value)
		}
	}
}

func TestStatusAndProgress(t *testing.T) {
	var buf bytes.Buffer
	ctx := &UIContext{
		Writer:    &buf,
		ColorMode: ColorNever,
		Quiet:     false,
	}
	output := NewOutput(ctx)

	// Test status
	output.Status("Installing dependencies")
	result := buf.String()
	if !strings.Contains(result, "Installing dependencies") {
		t.Error("Status should display message")
	}

	// Test progress
	buf.Reset()
	output.Progress(3, 10, "Processing files")
	result = buf.String()
	if !strings.Contains(result, "Processing files") {
		t.Error("Progress should display message")
	}
	if !strings.Contains(result, "3") && !strings.Contains(result, "10") {
		t.Error("Progress should display current and total")
	}
}

func TestSetterMethods(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	output := NewDefaultOutput()

	// Test SetWriter
	output.SetWriter(&buf1)
	output.Info("test message")
	if !strings.Contains(buf1.String(), "test message") {
		t.Error("SetWriter should change output destination")
	}

	// Test SetQuiet
	output.SetWriter(&buf2)
	output.SetQuiet(true)
	output.Info("quiet test")
	if strings.Contains(buf2.String(), "quiet test") {
		t.Error("SetQuiet(true) should suppress output")
	}

	// Test SetVerbose
	buf2.Reset()
	output.SetQuiet(false)
	output.SetVerbose(true)
	output.Debug("verbose test")
	if !strings.Contains(buf2.String(), "verbose test") {
		t.Error("SetVerbose(true) should enable debug output")
	}

	// Test SetColorMode
	output.SetColorMode(ColorNever)
	if output.context.ColorMode != ColorNever {
		t.Error("SetColorMode should update color mode")
	}
}

func TestHeadersAndSections(t *testing.T) {
	var buf bytes.Buffer
	ctx := &UIContext{
		Writer:    &buf,
		ColorMode: ColorNever,
		Quiet:     false,
	}
	output := NewOutput(ctx)

	// Test Header
	output.Header("Main Title")
	result := buf.String()
	if !strings.Contains(result, "Main Title") {
		t.Error("Header should display title")
	}

	// Test SubHeader
	buf.Reset()
	output.SubHeader("Subtitle")
	result = buf.String()
	if !strings.Contains(result, "Subtitle") {
		t.Error("SubHeader should display subtitle")
	}

	// Test Section
	buf.Reset()
	output.Section("Section Title", "Section content")
	result = buf.String()
	if !strings.Contains(result, "Section Title") || !strings.Contains(result, "Section content") {
		t.Error("Section should display both title and content")
	}

	// Test Box
	buf.Reset()
	output.Box("Boxed content")
	result = buf.String()
	if !strings.Contains(result, "Boxed content") {
		t.Error("Box should display content")
	}
}

func TestMarkdown(t *testing.T) {
	var buf bytes.Buffer
	ctx := &UIContext{
		Writer:    &buf,
		ColorMode: ColorNever,
		Quiet:     false,
	}
	output := NewOutput(ctx)

	markdown := `# Test Markdown

This is a **bold** test with *italic* text.

- List item 1
- List item 2

[Link](https://example.com)
`

	output.Markdown(markdown)
	result := buf.String()

	// Should contain markdown content (Fang will process it)
	if !strings.Contains(result, "Test Markdown") {
		t.Error("Markdown should render heading")
	}
}

func TestContext(t *testing.T) {
	ctx := &UIContext{
		Writer:      &bytes.Buffer{},
		ColorMode:   ColorAlways,
		Quiet:       true,
		Verbose:     false,
		Interactive: false,
	}
	output := NewOutput(ctx)

	returnedCtx := output.Context()
	if returnedCtx != ctx {
		t.Error("Context() should return the original context")
	}
	if returnedCtx.ColorMode != ColorAlways {
		t.Error("Context should preserve ColorMode")
	}
	if !returnedCtx.Quiet {
		t.Error("Context should preserve Quiet setting")
	}
}
