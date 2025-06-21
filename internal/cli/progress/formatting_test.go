// ABOUTME: Tests for result formatting and display logic
// ABOUTME: Validates formatters work correctly across different output formats

package progress

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestTableFormatter_FormatInstallationResults(t *testing.T) {
	formatter := NewTableFormatter()

	results := []*Result{
		{
			Operation: "install",
			Target:    "DBI",
			Success:   true,
			Duration:  time.Second * 2,
			Message:   "Successfully installed",
		},
		{
			Operation: "install",
			Target:    "Moose",
			Success:   false,
			Duration:  time.Second * 5,
			Message:   "Build failed",
			Error:     fmt.Errorf("compilation error"),
		},
	}

	lines := formatter.FormatInstallationResults(results)

	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines (header + separator + data), got %d", len(lines))
	}

	// Check that success and failure are indicated
	found := false
	for _, line := range lines {
		if strings.Contains(line, "✓ Success") || strings.Contains(line, "✗ Failed") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected success/failure indicators in output")
	}
}

func TestTableFormatter_FormatModuleList(t *testing.T) {
	formatter := NewTableFormatter()

	modules := []ModuleInfo{
		{
			Name:        "DBI",
			Version:     "1.643",
			Status:      "installed",
			Description: "Database independent interface for Perl",
		},
		{
			Name:        "Moose",
			Version:     "2.2015",
			Status:      "installed",
			Description: "A postmodern object system for Perl 5",
		},
	}

	// Test standard format
	lines := formatter.FormatModuleList(modules, "standard")
	if len(lines) == 0 {
		t.Error("Expected some output for standard format")
	}

	// Test detailed format
	detailedLines := formatter.FormatModuleList(modules, "detailed")
	if len(detailedLines) == 0 {
		t.Error("Expected some output for detailed format")
	}

	// Test compact format
	compactLines := formatter.FormatModuleList(modules, "compact")
	if len(compactLines) == 0 {
		t.Error("Expected some output for compact format")
	}

	// Detailed should have more lines than compact
	if len(detailedLines) <= len(compactLines) {
		t.Error("Expected detailed format to have more lines than compact format")
	}
}

func TestTableFormatter_FormatErrors(t *testing.T) {
	formatter := NewTableFormatter()

	errors := []ErrorInfo{
		{
			Module:  "DBI",
			Message: "Failed to compile XS extension",
			Code:    "BUILD_FAILED",
			Details: "gcc: command not found",
		},
		{
			Message: "Network timeout",
			Code:    "NETWORK_ERROR",
		},
	}

	lines := formatter.FormatErrors(errors)

	if len(lines) == 0 {
		t.Error("Expected some output for errors")
	}

	// Check that error count is shown
	found := false
	for _, line := range lines {
		if strings.Contains(line, "Errors (2)") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error count in output")
	}
}

func TestTableFormatter_FormatSummary(t *testing.T) {
	formatter := NewTableFormatter()

	summary := &SummaryInfo{
		TotalOperations:      10,
		SuccessfulOperations: 8,
		FailedOperations:     1,
		SkippedOperations:    1,
		TotalDuration:        time.Minute * 5,
		AverageDuration:      time.Second * 30,
		Warnings:             []string{"deprecated API used", "slow network"},
	}

	lines := formatter.FormatSummary(summary)

	if len(lines) == 0 {
		t.Error("Expected some output for summary")
	}

	// Check for key summary elements
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "Total Operations: 10") {
		t.Error("Expected total operations in summary")
	}
	if !strings.Contains(content, "Successful: 8") {
		t.Error("Expected successful count in summary")
	}
	if !strings.Contains(content, "Warnings (2)") {
		t.Error("Expected warnings section in summary")
	}
}

func TestTableFormatter_FormatProgress(t *testing.T) {
	formatter := NewTableFormatter()

	status := &Status{
		Operation:          "install",
		Current:            5,
		Total:              10,
		Message:            "Installing modules",
		Percentage:         50.0,
		ElapsedTime:        time.Minute * 2,
		EstimatedRemaining: time.Minute * 2,
	}

	progress := formatter.FormatProgress(status)

	if progress == "" {
		t.Error("Expected non-empty progress string")
	}

	// Check for key progress elements
	if !strings.Contains(progress, "50.0%") {
		t.Error("Expected percentage in progress")
	}
	if !strings.Contains(progress, "(5/10)") {
		t.Error("Expected current/total in progress")
	}
	if !strings.Contains(progress, "install") {
		t.Error("Expected operation name in progress")
	}
}

func TestTableFormatter_FormatParallelProgress(t *testing.T) {
	formatter := NewTableFormatter()

	operations := map[string]*OperationStatus{
		"op1": {
			ID:       "op1",
			Name:     "Install DBI",
			Status:   StatusCompleted,
			Progress: 100.0,
			Message:  "Done",
		},
		"op2": {
			ID:       "op2",
			Name:     "Install Moose",
			Status:   StatusRunning,
			Progress: 75.0,
			Message:  "Building",
		},
	}

	status := &ParallelStatus{
		Operations:          operations,
		TotalOperations:     2,
		CompletedOperations: 1,
		RunningOperations:   1,
		OverallPercentage:   87.5,
		ElapsedTime:         time.Minute * 3,
	}

	lines := formatter.FormatParallelProgress(status)

	if len(lines) == 0 {
		t.Error("Expected some output for parallel progress")
	}

	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "87.5%") {
		t.Error("Expected overall percentage in parallel progress")
	}
	if !strings.Contains(content, "Install DBI") {
		t.Error("Expected operation names in parallel progress")
	}
}

func TestJSONFormatter_FormatInstallationResults(t *testing.T) {
	formatter := NewJSONFormatter()

	results := []*Result{
		{
			Operation: "install",
			Target:    "DBI",
			Success:   true,
			Duration:  time.Second * 2,
			Message:   "Successfully installed",
		},
	}

	lines := formatter.FormatInstallationResults(results)

	if len(lines) != 1 {
		t.Errorf("Expected 1 line of JSON output, got %d", len(lines))
	}

	// Should be valid JSON
	jsonStr := lines[0]
	if !strings.Contains(jsonStr, "\"results\"") {
		t.Error("Expected 'results' key in JSON output")
	}
	if !strings.Contains(jsonStr, "\"count\"") {
		t.Error("Expected 'count' key in JSON output")
	}
}

func TestJSONFormatter_FormatModuleList(t *testing.T) {
	formatter := NewJSONFormatter()

	modules := []ModuleInfo{
		{
			Name:    "DBI",
			Version: "1.643",
			Status:  "installed",
		},
	}

	lines := formatter.FormatModuleList(modules, "standard")

	if len(lines) != 1 {
		t.Errorf("Expected 1 line of JSON output, got %d", len(lines))
	}

	jsonStr := lines[0]
	if !strings.Contains(jsonStr, "\"modules\"") {
		t.Error("Expected 'modules' key in JSON output")
	}
	if !strings.Contains(jsonStr, "DBI") {
		t.Error("Expected module name in JSON output")
	}
}

func TestListFormatter_FormatInstallationResults(t *testing.T) {
	formatter := NewListFormatter()

	results := []*Result{
		{
			Operation: "install",
			Target:    "DBI",
			Success:   true,
			Duration:  time.Second * 2,
			Message:   "Successfully installed",
		},
		{
			Operation: "install",
			Target:    "Moose",
			Success:   false,
			Duration:  time.Second * 5,
			Message:   "Build failed",
		},
	}

	lines := formatter.FormatInstallationResults(results)

	if len(lines) != 2 {
		t.Errorf("Expected 2 lines of output, got %d", len(lines))
	}

	// Check for bullets and status indicators
	for _, line := range lines {
		if !strings.Contains(line, "•") {
			t.Error("Expected bullet in list format")
		}
		if !strings.Contains(line, "SUCCESS") && !strings.Contains(line, "FAILED") {
			t.Error("Expected status indicator in list format")
		}
	}
}

func TestListFormatter_FormatModuleList(t *testing.T) {
	formatter := NewListFormatter()

	modules := []ModuleInfo{
		{
			Name:        "DBI",
			Version:     "1.643",
			Description: "Database independent interface",
		},
	}

	standardLines := formatter.FormatModuleList(modules, "standard")
	detailedLines := formatter.FormatModuleList(modules, "detailed")

	if len(standardLines) == 0 {
		t.Error("Expected output for standard format")
	}
	if len(detailedLines) == 0 {
		t.Error("Expected output for detailed format")
	}

	// Detailed should include description
	detailedContent := strings.Join(detailedLines, "\n")
	if !strings.Contains(detailedContent, "Database independent interface") {
		t.Error("Expected description in detailed format")
	}
}

func TestGetFormatter(t *testing.T) {
	tests := []struct {
		formatType string
		expected   string
	}{
		{"json", "*progress.JSONFormatter"},
		{"list", "*progress.ListFormatter"},
		{"table", "*progress.TableFormatter"},
		{"", "*progress.TableFormatter"},
		{"unknown", "*progress.TableFormatter"},
	}

	for _, test := range tests {
		formatter := GetFormatter(test.formatType)
		actualType := fmt.Sprintf("%T", formatter)
		if actualType != test.expected {
			t.Errorf("GetFormatter(%q) = %s, expected %s", test.formatType, actualType, test.expected)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, test := range tests {
		result := formatBytes(test.bytes)
		if result != test.expected {
			t.Errorf("formatBytes(%d) = %q, expected %q", test.bytes, result, test.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Millisecond, "500ms"},
		{2 * time.Second, "2s"},
		{90 * time.Second, "1m30s"},
		{2 * time.Hour, "2h0m0s"},
	}

	for _, test := range tests {
		result := FormatDuration(test.duration)
		if result != test.expected {
			t.Errorf("FormatDuration(%v) = %q, expected %q", test.duration, result, test.expected)
		}
	}
}

func TestFormatTimestamp(t *testing.T) {
	// Test zero time
	zeroTime := time.Time{}
	result := FormatTimestamp(zeroTime)
	if result != "N/A" {
		t.Errorf("FormatTimestamp(zero) = %q, expected %q", result, "N/A")
	}

	// Test actual time
	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
	result = FormatTimestamp(testTime)
	expected := "2023-12-25 15:30:45"
	if result != expected {
		t.Errorf("FormatTimestamp(%v) = %q, expected %q", testTime, result, expected)
	}
}

func TestFormatPercentage(t *testing.T) {
	tests := []struct {
		percentage float64
		expected   string
	}{
		{0, "0%"},
		{0.5, "0.50%"},
		{1.0, "1.0%"},
		{50.0, "50.0%"},
		{100.0, "100%"},
		{99.9, "99.9%"},
	}

	for _, test := range tests {
		result := FormatPercentage(test.percentage)
		if result != test.expected {
			t.Errorf("FormatPercentage(%f) = %q, expected %q", test.percentage, result, test.expected)
		}
	}
}

func TestTableFormatter_EmptyInputs(t *testing.T) {
	formatter := NewTableFormatter()

	// Test empty results
	lines := formatter.FormatInstallationResults([]*Result{})
	if len(lines) != 1 || !strings.Contains(lines[0], "No installation results") {
		t.Error("Expected 'No installation results' message for empty input")
	}

	// Test empty modules
	lines = formatter.FormatModuleList([]ModuleInfo{}, "standard")
	if len(lines) != 1 || !strings.Contains(lines[0], "No modules") {
		t.Error("Expected 'No modules' message for empty input")
	}

	// Test empty errors
	lines = formatter.FormatErrors([]ErrorInfo{})
	if len(lines) != 1 || !strings.Contains(lines[0], "No errors") {
		t.Error("Expected 'No errors' message for empty input")
	}
}

func TestTableFormatter_NilInputs(t *testing.T) {
	formatter := NewTableFormatter()

	// Test nil progress
	progress := formatter.FormatProgress(nil)
	if progress != "" {
		t.Error("Expected empty string for nil progress status")
	}

	// Test nil parallel progress
	lines := formatter.FormatParallelProgress(nil)
	if len(lines) != 0 {
		t.Error("Expected empty slice for nil parallel progress status")
	}
}

func TestFormatterOptions(t *testing.T) {
	// Test table formatter options
	tableFormatter := &TableFormatter{
		MaxWidth:     120,
		ShowHeaders:  false,
		ShowBorders:  false,
		Padding:      2,
		ColorEnabled: false,
	}

	results := []*Result{
		{Target: "DBI", Success: true, Duration: time.Second, Message: "OK"},
	}

	lines := tableFormatter.FormatInstallationResults(results)
	if len(lines) == 0 {
		t.Error("Expected output even with custom options")
	}

	// Since headers are disabled, should be fewer lines
	if len(lines) > 2 {
		t.Error("Expected fewer lines when headers are disabled")
	}

	// Test JSON formatter options
	jsonFormatter := &JSONFormatter{
		Indent:  false,
		Compact: true,
	}

	jsonLines := jsonFormatter.FormatInstallationResults(results)
	if len(jsonLines) != 1 {
		t.Error("Expected single line JSON output")
	}

	// Compact format should not have newlines within the JSON
	if strings.Contains(jsonLines[0], "\n") {
		t.Error("Expected compact JSON without newlines")
	}

	// Test list formatter options
	listFormatter := &ListFormatter{
		ShowBullets: false,
		Indent:      "    ",
		Separator:   " | ",
	}

	listLines := listFormatter.FormatInstallationResults(results)
	if len(listLines) == 0 {
		t.Error("Expected output from list formatter")
	}

	// Should not contain bullets
	for _, line := range listLines {
		if strings.Contains(line, "•") {
			t.Error("Expected no bullets when ShowBullets is false")
		}
	}
}

func TestLongTextTruncation(t *testing.T) {
	formatter := NewTableFormatter()

	// Test long message truncation
	results := []*Result{
		{
			Target:   "SomeModule",
			Success:  true,
			Message:  "This is a very long message that should be truncated to fit within the table format limits",
			Duration: time.Second,
		},
	}

	lines := formatter.FormatInstallationResults(results)
	found := false
	for _, line := range lines {
		if strings.Contains(line, "...") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected message truncation with '...' for long messages")
	}

	// Test long description truncation
	modules := []ModuleInfo{
		{
			Name:        "TestModule",
			Description: "This is a very long description that should be truncated when displayed in standard table format to maintain readability",
		},
	}

	moduleLines := formatter.FormatModuleList(modules, "standard")
	foundTruncation := false
	for _, line := range moduleLines {
		if strings.Contains(line, "...") {
			foundTruncation = true
			break
		}
	}
	if !foundTruncation {
		t.Error("Expected description truncation with '...' for long descriptions")
	}
}
