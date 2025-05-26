// ABOUTME: Structured output formatting for LSP diagnostics and messages
// ABOUTME: Converts PSC errors to various output formats for editor integration

package lsp

import (
	"encoding/json"
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/typechecker"
)

// OutputFormat represents different output formats for diagnostics
type OutputFormat string

const (
	OutputFormatLSP        OutputFormat = "lsp"        // Language Server Protocol format
	OutputFormatJSON       OutputFormat = "json"       // JSON format
	OutputFormatText       OutputFormat = "text"       // Human-readable text
	OutputFormatSarif      OutputFormat = "sarif"      // SARIF format for static analysis
	OutputFormatCheckstyle OutputFormat = "checkstyle" // Checkstyle XML format
)

// DiagnosticFormatter provides structured formatting for type checking results
type DiagnosticFormatter struct {
	format          OutputFormat
	includeSource   bool
	includeContext  bool
	maxContextLines int
}

// NewDiagnosticFormatter creates a new diagnostic formatter
func NewDiagnosticFormatter(format OutputFormat) *DiagnosticFormatter {
	return &DiagnosticFormatter{
		format:          format,
		includeSource:   true,
		includeContext:  true,
		maxContextLines: 3,
	}
}

// SetIncludeSource controls whether to include source code context
func (f *DiagnosticFormatter) SetIncludeSource(include bool) {
	f.includeSource = include
}

// SetIncludeContext controls whether to include surrounding context
func (f *DiagnosticFormatter) SetIncludeContext(include bool) {
	f.includeContext = include
}

// SetMaxContextLines sets the maximum number of context lines to include
func (f *DiagnosticFormatter) SetMaxContextLines(lines int) {
	f.maxContextLines = lines
}

// FormatTypeCheckResult formats a type check result according to the specified format
func (f *DiagnosticFormatter) FormatTypeCheckResult(result *typechecker.TypeCheckResult) (string, error) {
	switch f.format {
	case OutputFormatLSP:
		return f.formatLSP(result)
	case OutputFormatJSON:
		return f.formatJSON(result)
	case OutputFormatText:
		return f.formatText(result)
	case OutputFormatSarif:
		return f.formatSarif(result)
	case OutputFormatCheckstyle:
		return f.formatCheckstyle(result)
	default:
		return "", fmt.Errorf("unsupported output format: %s", f.format)
	}
}

// formatLSP formats errors in LSP diagnostic format
func (f *DiagnosticFormatter) formatLSP(result *typechecker.TypeCheckResult) (string, error) {
	diagnostics := make([]Diagnostic, len(result.Errors))

	for i, err := range result.Errors {
		severity := DiagnosticSeverityError

		// Determine severity based on error type
		if strings.Contains(err.Message, "warning") {
			severity = DiagnosticSeverityWarning
		} else if strings.Contains(err.Message, "hint") {
			severity = DiagnosticSeverityHint
		}

		diagnostics[i] = Diagnostic{
			Range: Range{
				Start: Position{Line: err.Line - 1, Character: err.Column - 1},  // LSP is 0-based
				End:   Position{Line: err.Line - 1, Character: err.Column + 10}, // Estimate end position
			},
			Severity: &severity,
			Code:     "PSC-" + extractErrorCode(err.Message),
			Source:   "psc",
			Message:  err.Message,
		}

		// Add code description if available
		if codeDesc := getCodeDescription(err.Message); codeDesc != "" {
			diagnostics[i].CodeDescription = &CodeDescription{
				Href: codeDesc,
			}
		}
	}

	params := PublishDiagnosticsParams{
		URI:         pathToURI(result.Path),
		Diagnostics: diagnostics,
	}

	data, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// formatJSON formats errors in structured JSON format
func (f *DiagnosticFormatter) formatJSON(result *typechecker.TypeCheckResult) (string, error) {
	type JSONError struct {
		File     string `json:"file"`
		Line     int    `json:"line"`
		Column   int    `json:"column"`
		Severity string `json:"severity"`
		Code     string `json:"code"`
		Message  string `json:"message"`
		Source   string `json:"source,omitempty"`
		Context  string `json:"context,omitempty"`
	}

	type JSONResult struct {
		Path            string      `json:"path"`
		ErrorCount      int         `json:"errorCount"`
		WarningCount    int         `json:"warningCount"`
		Errors          []JSONError `json:"errors"`
		TypeCheckPassed bool        `json:"typeCheckPassed"`
		FlowSensitive   bool        `json:"flowSensitive"`
	}

	jsonErrors := make([]JSONError, len(result.Errors))
	errorCount := 0
	warningCount := 0

	for i, err := range result.Errors {
		severity := "error"
		if strings.Contains(err.Message, "warning") {
			severity = "warning"
			warningCount++
		} else {
			errorCount++
		}

		jsonErrors[i] = JSONError{
			File:     err.Path,
			Line:     err.Line,
			Column:   err.Column,
			Severity: severity,
			Code:     extractErrorCode(err.Message),
			Message:  err.Message,
			Source:   "psc-type-checker",
		}

		// Add context if requested
		if f.includeContext {
			jsonErrors[i].Context = f.getErrorContext(err)
		}
	}

	jsonResult := JSONResult{
		Path:            result.Path,
		ErrorCount:      errorCount,
		WarningCount:    warningCount,
		Errors:          jsonErrors,
		TypeCheckPassed: len(result.Errors) == 0,
		FlowSensitive:   result.FlowSensitiveEnabled,
	}

	data, err := json.MarshalIndent(jsonResult, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// formatText formats errors in human-readable text format
func (f *DiagnosticFormatter) formatText(result *typechecker.TypeCheckResult) (string, error) {
	var builder strings.Builder

	// Header
	builder.WriteString(fmt.Sprintf("Type checking results for: %s\n", result.Path))
	builder.WriteString(strings.Repeat("=", 50) + "\n\n")

	if len(result.Errors) == 0 {
		builder.WriteString("✓ No type errors found\n")
		if result.FlowSensitiveEnabled {
			builder.WriteString("✓ Flow-sensitive analysis enabled\n")
		}
		return builder.String(), nil
	}

	// Error summary
	errorCount := 0
	warningCount := 0
	for _, err := range result.Errors {
		if strings.Contains(err.Message, "warning") {
			warningCount++
		} else {
			errorCount++
		}
	}

	builder.WriteString(fmt.Sprintf("Found %d error(s) and %d warning(s)\n\n", errorCount, warningCount))

	// Individual errors
	for i, err := range result.Errors {
		severity := "ERROR"
		if strings.Contains(err.Message, "warning") {
			severity = "WARNING"
		}

		builder.WriteString(fmt.Sprintf("%d. [%s] %s:%d:%d\n",
			i+1, severity, err.Path, err.Line, err.Column))
		builder.WriteString(fmt.Sprintf("   %s\n", err.Message))

		// Add context if requested
		if f.includeContext {
			if context := f.getErrorContext(err); context != "" {
				builder.WriteString(fmt.Sprintf("   Context: %s\n", context))
			}
		}

		builder.WriteString("\n")
	}

	// Footer
	if result.FlowSensitiveEnabled {
		builder.WriteString("Note: Flow-sensitive analysis was enabled\n")
	}

	return builder.String(), nil
}

// formatSarif formats errors in SARIF format for static analysis tools
func (f *DiagnosticFormatter) formatSarif(result *typechecker.TypeCheckResult) (string, error) {
	type SarifLocation struct {
		PhysicalLocation struct {
			ArtifactLocation struct {
				URI string `json:"uri"`
			} `json:"artifactLocation"`
			Region struct {
				StartLine   int `json:"startLine"`
				StartColumn int `json:"startColumn"`
				EndLine     int `json:"endLine,omitempty"`
				EndColumn   int `json:"endColumn,omitempty"`
			} `json:"region"`
		} `json:"physicalLocation"`
	}

	type SarifResult struct {
		RuleID  string `json:"ruleId"`
		Level   string `json:"level"`
		Message struct {
			Text string `json:"text"`
		} `json:"message"`
		Locations []SarifLocation `json:"locations"`
	}

	type SarifRun struct {
		Tool struct {
			Driver struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"driver"`
		} `json:"tool"`
		Results []SarifResult `json:"results"`
	}

	type SarifReport struct {
		Schema  string     `json:"$schema"`
		Version string     `json:"version"`
		Runs    []SarifRun `json:"runs"`
	}

	results := make([]SarifResult, len(result.Errors))

	for i, err := range result.Errors {
		level := "error"
		if strings.Contains(err.Message, "warning") {
			level = "warning"
		}

		results[i] = SarifResult{
			RuleID: "PSC-" + extractErrorCode(err.Message),
			Level:  level,
			Message: struct {
				Text string `json:"text"`
			}{
				Text: err.Message,
			},
			Locations: []SarifLocation{
				{
					PhysicalLocation: struct {
						ArtifactLocation struct {
							URI string `json:"uri"`
						} `json:"artifactLocation"`
						Region struct {
							StartLine   int `json:"startLine"`
							StartColumn int `json:"startColumn"`
							EndLine     int `json:"endLine,omitempty"`
							EndColumn   int `json:"endColumn,omitempty"`
						} `json:"region"`
					}{
						ArtifactLocation: struct {
							URI string `json:"uri"`
						}{
							URI: pathToURI(err.Path),
						},
						Region: struct {
							StartLine   int `json:"startLine"`
							StartColumn int `json:"startColumn"`
							EndLine     int `json:"endLine,omitempty"`
							EndColumn   int `json:"endColumn,omitempty"`
						}{
							StartLine:   err.Line,
							StartColumn: err.Column,
						},
					},
				},
			},
		}
	}

	report := SarifReport{
		Schema:  "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0.json",
		Version: "2.1.0",
		Runs: []SarifRun{
			{
				Tool: struct {
					Driver struct {
						Name    string `json:"name"`
						Version string `json:"version"`
					} `json:"driver"`
				}{
					Driver: struct {
						Name    string `json:"name"`
						Version string `json:"version"`
					}{
						Name:    "PSC Type Checker",
						Version: "1.0.0",
					},
				},
				Results: results,
			},
		},
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// formatCheckstyle formats errors in Checkstyle XML format
func (f *DiagnosticFormatter) formatCheckstyle(result *typechecker.TypeCheckResult) (string, error) {
	var builder strings.Builder

	builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	builder.WriteString(`<checkstyle version="10.0">` + "\n")

	if len(result.Errors) > 0 {
		builder.WriteString(fmt.Sprintf(`  <file name="%s">`, escapeXML(result.Path)) + "\n")

		for _, err := range result.Errors {
			severity := "error"
			if strings.Contains(err.Message, "warning") {
				severity = "warning"
			}

			builder.WriteString(fmt.Sprintf(
				`    <error line="%d" column="%d" severity="%s" message="%s" source="psc.%s"/>`,
				err.Line, err.Column, severity, escapeXML(err.Message), extractErrorCode(err.Message),
			) + "\n")
		}

		builder.WriteString("  </file>\n")
	}

	builder.WriteString("</checkstyle>\n")

	return builder.String(), nil
}

// extractErrorCode extracts error code from error message
func extractErrorCode(message string) string {
	// Look for patterns like "PSC-801" or "801"
	parts := strings.Fields(message)
	for _, part := range parts {
		if strings.HasPrefix(part, "PSC-") {
			return strings.TrimPrefix(part, "PSC-")
		}
		if len(part) == 3 && isNumeric(part) {
			return part
		}
	}
	return "unknown"
}

// getCodeDescription returns a URL with more information about the error code
func getCodeDescription(message string) string {
	code := extractErrorCode(message)
	if code != "unknown" {
		return fmt.Sprintf("https://docs.pvm.dev/errors/PSC-%s", code)
	}
	return ""
}

// getErrorContext retrieves context around an error (placeholder implementation)
func (f *DiagnosticFormatter) getErrorContext(err typechecker.TypeCheckError) string {
	// In a real implementation, this would read the source file
	// and extract lines around the error location
	return fmt.Sprintf("Error at line %d, column %d", err.Line, err.Column)
}

// escapeXML escapes characters for XML output
func escapeXML(s string) string {
	replacements := map[string]string{
		"&":  "&amp;",
		"<":  "&lt;",
		">":  "&gt;",
		"\"": "&quot;",
		"'":  "&apos;",
	}

	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}

	return s
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// FormatErrorsForEditor formats errors specifically for editor display
func FormatErrorsForEditor(errors []typechecker.TypeCheckError, format OutputFormat) (string, error) {
	result := &typechecker.TypeCheckResult{
		Path:   "",
		Errors: errors,
	}

	if len(errors) > 0 {
		result.Path = errors[0].Path
	}

	formatter := NewDiagnosticFormatter(format)
	return formatter.FormatTypeCheckResult(result)
}

// FormatSingleError formats a single error for quick display
func FormatSingleError(err typechecker.TypeCheckError, includeLocation bool) string {
	if includeLocation {
		return fmt.Sprintf("%s:%d:%d: %s", err.Path, err.Line, err.Column, err.Message)
	}
	return err.Message
}
