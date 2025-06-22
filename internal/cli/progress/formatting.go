// ABOUTME: Result formatting and display logic for CLI operations
// ABOUTME: Provides standardized output formatting across all PVM components

package progress

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Formatter provides interface for formatting operation results
type Formatter interface {
	FormatInstallationResults(results []*Result) []string
	FormatModuleList(modules []ModuleInfo, format string) []string
	FormatErrors(errors []ErrorInfo) []string
	FormatSummary(summary *SummaryInfo) []string
	FormatProgress(status *Status) string
	FormatParallelProgress(status *ParallelStatus) []string
}

// ModuleInfo represents module information for formatting
type ModuleInfo struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version,omitempty"`
	Description string                 `json:"description,omitempty"`
	Author      string                 `json:"author,omitempty"`
	Size        int64                  `json:"size,omitempty"`
	InstallDate time.Time              `json:"install_date,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ErrorInfo represents error information for formatting
type ErrorInfo struct {
	Module  string `json:"module,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// TableFormatter provides table-based formatting
type TableFormatter struct {
	MaxWidth     int
	ShowHeaders  bool
	ShowBorders  bool
	Padding      int
	ColorEnabled bool
}

// NewTableFormatter creates a new table formatter
func NewTableFormatter() *TableFormatter {
	return &TableFormatter{
		MaxWidth:     80,
		ShowHeaders:  true,
		ShowBorders:  true,
		Padding:      1,
		ColorEnabled: true,
	}
}

// FormatInstallationResults formats installation results as a table
func (tf *TableFormatter) FormatInstallationResults(results []*Result) []string {
	if len(results) == 0 {
		return []string{"No installation results to display"}
	}

	var lines []string
	if tf.ShowHeaders {
		lines = append(lines, tf.formatTableHeader([]string{"Module", "Status", "Duration", "Message"}))
		if tf.ShowBorders {
			lines = append(lines, tf.formatTableSeparator([]string{"Module", "Status", "Duration", "Message"}))
		}
	}

	for _, result := range results {
		status := "✓ Success"
		if !result.Success {
			status = "✗ Failed"
		}

		duration := result.Duration.Round(time.Millisecond).String()
		message := result.Message
		if len(message) > 40 {
			message = message[:37] + "..."
		}

		row := tf.formatTableRow([]string{result.Target, status, duration, message})
		lines = append(lines, row)
	}

	return lines
}

// FormatModuleList formats module list as a table
func (tf *TableFormatter) FormatModuleList(modules []ModuleInfo, format string) []string {
	if len(modules) == 0 {
		return []string{"No modules to display"}
	}

	switch format {
	case "detailed":
		return tf.formatDetailedModuleList(modules)
	case "compact":
		return tf.formatCompactModuleList(modules)
	default:
		return tf.formatStandardModuleList(modules)
	}
}

// FormatErrors formats errors as a structured list
func (tf *TableFormatter) FormatErrors(errors []ErrorInfo) []string {
	if len(errors) == 0 {
		return []string{"No errors to display"}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Errors (%d):", len(errors)))
	lines = append(lines, "")

	for i, err := range errors {
		lines = append(lines, fmt.Sprintf("%d. %s", i+1, err.Message))
		if err.Module != "" {
			lines = append(lines, fmt.Sprintf("   Module: %s", err.Module))
		}
		if err.Code != "" {
			lines = append(lines, fmt.Sprintf("   Code: %s", err.Code))
		}
		if err.Details != "" {
			details := strings.ReplaceAll(err.Details, "\n", "\n   ")
			lines = append(lines, fmt.Sprintf("   Details: %s", details))
		}
		if i < len(errors)-1 {
			lines = append(lines, "")
		}
	}

	return lines
}

// FormatSummary formats operation summary
func (tf *TableFormatter) FormatSummary(summary *SummaryInfo) []string {
	var lines []string

	lines = append(lines, "Operation Summary")
	lines = append(lines, strings.Repeat("=", 40))
	lines = append(lines, fmt.Sprintf("Total Operations: %d", summary.TotalOperations))
	lines = append(lines, fmt.Sprintf("Successful: %d", summary.SuccessfulOperations))
	lines = append(lines, fmt.Sprintf("Failed: %d", summary.FailedOperations))
	lines = append(lines, fmt.Sprintf("Skipped: %d", summary.SkippedOperations))
	lines = append(lines, fmt.Sprintf("Total Duration: %s", summary.TotalDuration.Round(time.Millisecond)))

	if summary.TotalOperations > 0 {
		lines = append(lines, fmt.Sprintf("Average Duration: %s", summary.AverageDuration.Round(time.Millisecond)))
	}

	if len(summary.Warnings) > 0 {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Warnings (%d):", len(summary.Warnings)))
		for _, warning := range summary.Warnings {
			lines = append(lines, fmt.Sprintf("  • %s", warning))
		}
	}

	return lines
}

// FormatProgress formats single operation progress
func (tf *TableFormatter) FormatProgress(status *Status) string {
	if status == nil {
		return ""
	}

	var parts []string

	// Progress bar
	if status.Total > 0 {
		barWidth := 30
		filled := int(float64(barWidth) * status.Percentage / 100.0)
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		parts = append(parts, fmt.Sprintf("[%s] %.1f%%", bar, status.Percentage))
	}

	// Current/Total
	if status.Total > 0 {
		parts = append(parts, fmt.Sprintf("(%d/%d)", status.Current, status.Total))
	}

	// Operation and message
	if status.Operation != "" && status.Message != "" {
		parts = append(parts, fmt.Sprintf("%s: %s", status.Operation, status.Message))
	} else if status.Operation != "" {
		parts = append(parts, status.Operation)
	} else if status.Message != "" {
		parts = append(parts, status.Message)
	}

	// Timing information
	if status.ElapsedTime > 0 {
		parts = append(parts, fmt.Sprintf("elapsed: %s", status.ElapsedTime.Round(time.Second)))
	}
	if status.EstimatedRemaining > 0 {
		parts = append(parts, fmt.Sprintf("ETA: %s", status.EstimatedRemaining.Round(time.Second)))
	}

	return strings.Join(parts, " ")
}

// FormatParallelProgress formats parallel operation progress
func (tf *TableFormatter) FormatParallelProgress(status *ParallelStatus) []string {
	if status == nil {
		return []string{}
	}

	var lines []string

	// Overall progress
	lines = append(lines, fmt.Sprintf("Overall Progress: %.1f%% (%d/%d completed, %d failed, %d running)",
		status.OverallPercentage,
		status.CompletedOperations,
		status.TotalOperations,
		status.FailedOperations,
		status.RunningOperations))

	if status.ElapsedTime > 0 {
		lines = append(lines, fmt.Sprintf("Elapsed: %s", status.ElapsedTime.Round(time.Second)))
	}
	if status.EstimatedRemaining > 0 {
		lines = append(lines, fmt.Sprintf("ETA: %s", status.EstimatedRemaining.Round(time.Second)))
	}

	// Individual operations
	if len(status.Operations) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Operation Details:")

		// Sort operations by name for consistent display
		var sortedOps []*OperationStatus
		for _, op := range status.Operations {
			sortedOps = append(sortedOps, op)
		}
		sort.Slice(sortedOps, func(i, j int) bool {
			return sortedOps[i].Name < sortedOps[j].Name
		})

		for _, op := range sortedOps {
			statusIcon := "○"
			switch op.Status {
			case StatusRunning:
				statusIcon = "◐"
			case StatusCompleted:
				statusIcon = "●"
			case StatusFailed:
				statusIcon = "✗"
			case StatusSkipped:
				statusIcon = "⊝"
			}

			line := fmt.Sprintf("  %s %s", statusIcon, op.Name)
			if op.Progress > 0 {
				line += fmt.Sprintf(" (%.1f%%)", op.Progress)
			}
			if op.Message != "" {
				line += fmt.Sprintf(": %s", op.Message)
			}
			lines = append(lines, line)
		}
	}

	return lines
}

// formatTableHeader formats table headers
func (tf *TableFormatter) formatTableHeader(columns []string) string {
	return tf.formatTableRow(columns)
}

// formatTableSeparator formats table separator line
func (tf *TableFormatter) formatTableSeparator(columns []string) string {
	var parts []string
	for _, col := range columns {
		parts = append(parts, strings.Repeat("-", len(col)+tf.Padding*2))
	}
	return strings.Join(parts, "+")
}

// formatTableRow formats a table row
func (tf *TableFormatter) formatTableRow(columns []string) string {
	var parts []string
	for _, col := range columns {
		padded := fmt.Sprintf("%s%s%s",
			strings.Repeat(" ", tf.Padding),
			col,
			strings.Repeat(" ", tf.Padding))
		parts = append(parts, padded)
	}
	if tf.ShowBorders {
		return "|" + strings.Join(parts, "|") + "|"
	}
	return strings.Join(parts, " ")
}

// formatStandardModuleList formats modules in standard table format
func (tf *TableFormatter) formatStandardModuleList(modules []ModuleInfo) []string {
	var lines []string

	if tf.ShowHeaders {
		lines = append(lines, tf.formatTableHeader([]string{"Name", "Version", "Status", "Description"}))
		if tf.ShowBorders {
			lines = append(lines, tf.formatTableSeparator([]string{"Name", "Version", "Status", "Description"}))
		}
	}

	for _, module := range modules {
		description := module.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}

		row := tf.formatTableRow([]string{
			module.Name,
			module.Version,
			module.Status,
			description,
		})
		lines = append(lines, row)
	}

	return lines
}

// formatDetailedModuleList formats modules with detailed information
func (tf *TableFormatter) formatDetailedModuleList(modules []ModuleInfo) []string {
	var lines []string

	for i, module := range modules {
		lines = append(lines, fmt.Sprintf("Module: %s", module.Name))
		if module.Version != "" {
			lines = append(lines, fmt.Sprintf("  Version: %s", module.Version))
		}
		if module.Author != "" {
			lines = append(lines, fmt.Sprintf("  Author: %s", module.Author))
		}
		if module.Status != "" {
			lines = append(lines, fmt.Sprintf("  Status: %s", module.Status))
		}
		if module.Size > 0 {
			lines = append(lines, fmt.Sprintf("  Size: %s", formatBytes(module.Size)))
		}
		if !module.InstallDate.IsZero() {
			lines = append(lines, fmt.Sprintf("  Installed: %s", module.InstallDate.Format("2006-01-02 15:04:05")))
		}
		if module.Description != "" {
			lines = append(lines, fmt.Sprintf("  Description: %s", module.Description))
		}

		if i < len(modules)-1 {
			lines = append(lines, "")
		}
	}

	return lines
}

// formatCompactModuleList formats modules in compact format
func (tf *TableFormatter) formatCompactModuleList(modules []ModuleInfo) []string {
	var lines []string

	for _, module := range modules {
		line := module.Name
		if module.Version != "" {
			line += fmt.Sprintf(" (%s)", module.Version)
		}
		if module.Status != "" {
			line += fmt.Sprintf(" [%s]", module.Status)
		}
		lines = append(lines, line)
	}

	return lines
}

// JSONFormatter provides JSON-based formatting
type JSONFormatter struct {
	Indent  bool
	Compact bool
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		Indent:  true,
		Compact: false,
	}
}

// FormatInstallationResults formats installation results as JSON
func (jf *JSONFormatter) FormatInstallationResults(results []*Result) []string {
	data := map[string]interface{}{
		"results": results,
		"count":   len(results),
	}
	return jf.formatJSON(data)
}

// FormatModuleList formats module list as JSON
func (jf *JSONFormatter) FormatModuleList(modules []ModuleInfo, format string) []string {
	data := map[string]interface{}{
		"modules": modules,
		"count":   len(modules),
		"format":  format,
	}
	return jf.formatJSON(data)
}

// FormatErrors formats errors as JSON
func (jf *JSONFormatter) FormatErrors(errors []ErrorInfo) []string {
	data := map[string]interface{}{
		"errors": errors,
		"count":  len(errors),
	}
	return jf.formatJSON(data)
}

// FormatSummary formats summary as JSON
func (jf *JSONFormatter) FormatSummary(summary *SummaryInfo) []string {
	return jf.formatJSON(summary)
}

// FormatProgress formats progress as JSON
func (jf *JSONFormatter) FormatProgress(status *Status) string {
	data, _ := json.Marshal(status)
	return string(data)
}

// FormatParallelProgress formats parallel progress as JSON
func (jf *JSONFormatter) FormatParallelProgress(status *ParallelStatus) []string {
	return jf.formatJSON(status)
}

// formatJSON formats data as JSON
func (jf *JSONFormatter) formatJSON(data interface{}) []string {
	var jsonData []byte
	var err error

	if jf.Indent && !jf.Compact {
		jsonData, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonData, err = json.Marshal(data)
	}

	if err != nil {
		return []string{fmt.Sprintf("Error formatting JSON: %v", err)}
	}

	return []string{string(jsonData)}
}

// ListFormatter provides simple list-based formatting
type ListFormatter struct {
	ShowBullets bool
	Indent      string
	Separator   string
}

// NewListFormatter creates a new list formatter
func NewListFormatter() *ListFormatter {
	return &ListFormatter{
		ShowBullets: true,
		Indent:      "  ",
		Separator:   "\n",
	}
}

// FormatInstallationResults formats installation results as a list
func (lf *ListFormatter) FormatInstallationResults(results []*Result) []string {
	var lines []string

	for _, result := range results {
		bullet := "•"
		if !lf.ShowBullets {
			bullet = ""
		}

		status := "SUCCESS"
		if !result.Success {
			status = "FAILED"
		}

		line := fmt.Sprintf("%s %s [%s] %s (%s)",
			bullet,
			result.Target,
			status,
			result.Message,
			result.Duration.Round(time.Millisecond))

		lines = append(lines, line)
	}

	return lines
}

// FormatModuleList formats module list as a list
func (lf *ListFormatter) FormatModuleList(modules []ModuleInfo, format string) []string {
	var lines []string

	for _, module := range modules {
		bullet := "•"
		if !lf.ShowBullets {
			bullet = ""
		}

		line := fmt.Sprintf("%s %s", bullet, module.Name)
		if module.Version != "" {
			line += fmt.Sprintf(" (%s)", module.Version)
		}
		if module.Description != "" && format == "detailed" {
			line += fmt.Sprintf(" - %s", module.Description)
		}

		lines = append(lines, line)
	}

	return lines
}

// FormatErrors formats errors as a list
func (lf *ListFormatter) FormatErrors(errors []ErrorInfo) []string {
	var lines []string

	for _, err := range errors {
		bullet := "•"
		if !lf.ShowBullets {
			bullet = ""
		}

		line := fmt.Sprintf("%s %s", bullet, err.Message)
		if err.Module != "" {
			line += fmt.Sprintf(" (Module: %s)", err.Module)
		}

		lines = append(lines, line)
	}

	return lines
}

// FormatSummary formats summary as a list
func (lf *ListFormatter) FormatSummary(summary *SummaryInfo) []string {
	var lines []string

	lines = append(lines, "Operation Summary:")
	lines = append(lines, fmt.Sprintf("%sTotal: %d", lf.Indent, summary.TotalOperations))
	lines = append(lines, fmt.Sprintf("%sSuccessful: %d", lf.Indent, summary.SuccessfulOperations))
	lines = append(lines, fmt.Sprintf("%sFailed: %d", lf.Indent, summary.FailedOperations))
	lines = append(lines, fmt.Sprintf("%sSkipped: %d", lf.Indent, summary.SkippedOperations))
	lines = append(lines, fmt.Sprintf("%sDuration: %s", lf.Indent, summary.TotalDuration.Round(time.Millisecond)))

	return lines
}

// FormatProgress formats progress as a simple line
func (lf *ListFormatter) FormatProgress(status *Status) string {
	if status == nil {
		return ""
	}

	return fmt.Sprintf("%s: %.1f%% (%d/%d) - %s",
		status.Operation,
		status.Percentage,
		status.Current,
		status.Total,
		status.Message)
}

// FormatParallelProgress formats parallel progress as a list
func (lf *ListFormatter) FormatParallelProgress(status *ParallelStatus) []string {
	if status == nil {
		return []string{}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Progress: %.1f%% (%d/%d operations)",
		status.OverallPercentage,
		status.CompletedOperations,
		status.TotalOperations))

	for _, op := range status.Operations {
		bullet := "•"
		if !lf.ShowBullets {
			bullet = ""
		}

		line := fmt.Sprintf("%s %s [%s] %.1f%%",
			bullet,
			op.Name,
			op.Status.String(),
			op.Progress)

		if op.Message != "" {
			line += fmt.Sprintf(" - %s", op.Message)
		}

		lines = append(lines, line)
	}

	return lines
}

// GetFormatter returns the appropriate formatter based on format type
func GetFormatter(formatType string) Formatter {
	switch strings.ToLower(formatType) {
	case "json":
		return NewJSONFormatter()
	case "list":
		return NewListFormatter()
	case "table", "":
		return NewTableFormatter()
	default:
		return NewTableFormatter()
	}
}

// formatBytes formats byte count as human readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration formats duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	} else if d < time.Minute {
		return d.Round(time.Second).String()
	} else if d < time.Hour {
		return d.Round(time.Second).String()
	} else {
		return d.Round(time.Minute).String()
	}
}

// FormatTimestamp formats timestamp in a consistent way
func FormatTimestamp(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format("2006-01-02 15:04:05")
}

// FormatPercentage formats percentage with appropriate precision
func FormatPercentage(percentage float64) string {
	if percentage == 0 {
		return "0%"
	} else if percentage == 100 {
		return "100%"
	} else if percentage < 1 {
		return fmt.Sprintf("%.2f%%", percentage)
	} else {
		return fmt.Sprintf("%.1f%%", percentage)
	}
}
