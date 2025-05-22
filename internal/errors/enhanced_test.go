// ABOUTME: Tests for enhanced error handling features
// ABOUTME: Validates error formatting, collection, and severity handling

package errors

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnhancedError(t *testing.T) {
	// Create an enhanced error
	err := NewEnhancedError(
		PrefixPVM,
		CategoryVersion,
		"TEST001",
		"Test error message",
		fmt.Errorf("underlying error"),
		SeverityError,
	)

	// Test basic error properties
	assert.Equal(t, PrefixPVM, err.Prefix())
	assert.Equal(t, CategoryVersion, err.Category())
	assert.Equal(t, "TEST001", err.Code())
	assert.Equal(t, "Test error message", err.Message())
	assert.Equal(t, SeverityError, err.Severity())

	// Test error string contains all expected parts
	errorStr := err.Error()
	assert.Contains(t, errorStr, "PVM-TEST001")
	assert.Contains(t, errorStr, "Test error message")
	assert.Contains(t, errorStr, "Version Error")
	assert.Contains(t, errorStr, "Severity: ERROR")
	assert.Contains(t, errorStr, "underlying error")
}

func TestEnhancedErrorWithContext(t *testing.T) {
	err := NewEnhancedError(
		PrefixPVX,
		CategoryExecution,
		"EXEC001",
		"Execution failed",
		nil,
		SeverityError,
	).WithContext("script", "test.pl").
		WithContext("version", "5.38.0").
		WithRecoveryAction("Check script syntax").
		WithRecoveryAction("Verify Perl version compatibility")

	// Test context
	context := err.Context()
	assert.Equal(t, "test.pl", context["script"])
	assert.Equal(t, "5.38.0", context["version"])

	// Test recovery actions
	actions := err.RecoveryActions()
	assert.Len(t, actions, 2)
	assert.Contains(t, actions, "Check script syntax")
	assert.Contains(t, actions, "Verify Perl version compatibility")

	// Test that error string includes context and actions
	errorStr := err.Error()
	assert.Contains(t, errorStr, "Context:")
	assert.Contains(t, errorStr, "script: test.pl")
	assert.Contains(t, errorStr, "Suggested Actions:")
	assert.Contains(t, errorStr, "1. Check script syntax")
}

func TestEnhancedErrorWithRelatedErrors(t *testing.T) {
	relatedErr1 := New(PrefixPVI, CategoryModule, "MOD001", "Module not found", nil)
	relatedErr2 := New(PrefixPVI, CategoryModule, "MOD002", "Dependency conflict", nil)

	mainErr := NewEnhancedError(
		PrefixPVI,
		CategoryModule,
		"MAIN001",
		"Installation failed",
		nil,
		SeverityCritical,
	).WithRelatedError(relatedErr1).
		WithRelatedError(relatedErr2)

	// Test related errors
	related := mainErr.RelatedErrors()
	assert.Len(t, related, 2)

	// Test that error string includes related errors
	errorStr := mainErr.Error()
	assert.Contains(t, errorStr, "Related Errors:")
	assert.Contains(t, errorStr, "PVI-MOD001")
	assert.Contains(t, errorStr, "PVI-MOD002")
}

func TestErrorSeverity(t *testing.T) {
	tests := []struct {
		severity ErrorSeverity
		expected string
	}{
		{SeverityInfo, "INFO"},
		{SeverityWarning, "WARNING"},
		{SeverityError, "ERROR"},
		{SeverityCritical, "CRITICAL"},
		{ErrorSeverity(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.severity.String())
		})
	}
}

func TestErrorFormatter(t *testing.T) {
	tests := []struct {
		name         string
		colorEnabled bool
		verbose      bool
		severity     ErrorSeverity
		expectColor  bool
		expectDetail bool
	}{
		{
			name:         "Color enabled, verbose",
			colorEnabled: true,
			verbose:      true,
			severity:     SeverityError,
			expectColor:  true,
			expectDetail: true,
		},
		{
			name:         "Color disabled, verbose",
			colorEnabled: false,
			verbose:      true,
			severity:     SeverityWarning,
			expectColor:  false,
			expectDetail: true,
		},
		{
			name:         "Color enabled, not verbose",
			colorEnabled: true,
			verbose:      false,
			severity:     SeverityInfo,
			expectColor:  true,
			expectDetail: false,
		},
		{
			name:         "Color disabled, not verbose",
			colorEnabled: false,
			verbose:      false,
			severity:     SeverityCritical,
			expectColor:  false,
			expectDetail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewErrorFormatter(tt.colorEnabled, tt.verbose)

			err := NewEnhancedError(
				PrefixPSC,
				CategoryType,
				"TYPE001",
				"Type error",
				nil,
				tt.severity,
			).WithContext("file", "test.pl").
				WithRecoveryAction("Fix type annotation")

			formatted := formatter.Format(err)

			if tt.expectColor {
				// Check for ANSI color codes
				assert.True(t, strings.Contains(formatted, "\033["))
			} else {
				// Should not contain color codes
				assert.False(t, strings.Contains(formatted, "\033["))
			}

			if tt.expectDetail {
				// Should contain detailed information
				assert.Contains(t, formatted, "Context:")
				assert.Contains(t, formatted, "Suggested Actions:")
			} else {
				// Should be concise
				assert.Contains(t, formatted, "PSC-TYPE001")
				assert.Contains(t, formatted, "Type error")
			}
		})
	}
}

func TestErrorFormatterWithBasicError(t *testing.T) {
	formatter := NewErrorFormatter(true, true)

	// Test with basic Error (not EnhancedError)
	basicErr := New(PrefixSYS, CategorySystem, "SYS001", "System error", nil)
	formatted := formatter.Format(basicErr)

	// Should contain color codes for basic errors
	assert.Contains(t, formatted, "\033[31m") // Red color
	assert.Contains(t, formatted, "SYS-SYS001")
	assert.Contains(t, formatted, "System error")
}

func TestErrorCollector(t *testing.T) {
	collector := NewErrorCollector()

	// Add different types of errors
	infoErr := NewEnhancedError(PrefixPVM, CategoryVersion, "INFO001", "Info message", nil, SeverityInfo)
	warningErr := NewEnhancedError(PrefixPVX, CategoryExecution, "WARN001", "Warning message", nil, SeverityWarning)
	errorErr := NewEnhancedError(PrefixPVI, CategoryModule, "ERR001", "Error message", nil, SeverityError)
	criticalErr := NewEnhancedError(PrefixPSC, CategoryType, "CRIT001", "Critical message", nil, SeverityCritical)

	collector.Add(infoErr)
	collector.Add(warningErr)
	collector.Add(errorErr)
	collector.Add(criticalErr)

	// Test counts
	assert.True(t, collector.HasErrors())
	assert.True(t, collector.HasWarnings())

	// Test categorization
	errors := collector.Errors()
	warnings := collector.Warnings()
	infos := collector.Infos()

	assert.Len(t, errors, 2) // error and critical both go to errors
	assert.Len(t, warnings, 1)
	assert.Len(t, infos, 1)

	// Test all errors
	all := collector.All()
	assert.Len(t, all, 4)
}

func TestErrorCollectorWithBasicErrors(t *testing.T) {
	collector := NewErrorCollector()

	// Add basic error (not enhanced)
	basicErr := New(PrefixCFG, CategoryConfig, "CFG001", "Config error", nil)
	collector.Add(basicErr)

	// Should be categorized as error by default
	assert.True(t, collector.HasErrors())
	assert.False(t, collector.HasWarnings())

	errors := collector.Errors()
	assert.Len(t, errors, 1)
}

func TestErrorCollectorWithNil(t *testing.T) {
	collector := NewErrorCollector()

	// Adding nil should not panic or add anything
	collector.Add(nil)

	assert.False(t, collector.HasErrors())
	assert.False(t, collector.HasWarnings())
	assert.Empty(t, collector.All())
}

func TestEnhancedErrorChaining(t *testing.T) {
	// Test method chaining
	err := NewEnhancedError(
		PrefixPVM,
		CategoryVersion,
		"CHAIN001",
		"Chaining test",
		nil,
		SeverityWarning,
	).WithSeverity(SeverityError).
		WithContext("test", "value").
		WithRecoveryAction("First action").
		WithRecoveryAction("Second action")

	// Test that chaining worked
	assert.Equal(t, SeverityError, err.Severity())
	assert.Equal(t, "value", err.Context()["test"])
	assert.Len(t, err.RecoveryActions(), 2)
}

func TestEnhancedErrorInterface(t *testing.T) {
	err := NewEnhancedError(
		PrefixPVX,
		CategoryExecution,
		"IFACE001",
		"Interface test",
		nil,
		SeverityError,
	)

	// Test that it implements TypedError interface
	var typedErr TypedError = err

	assert.Equal(t, PrefixPVX, typedErr.Prefix())
	assert.Equal(t, CategoryExecution, typedErr.Category())
	assert.Equal(t, "IFACE001", typedErr.Code())
	assert.Equal(t, "Interface test", typedErr.Message())
	assert.Equal(t, "Interface test", typedErr.Description()) // Backward compatibility
	assert.Equal(t, "", typedErr.Location())                  // Not set
	assert.Equal(t, "", typedErr.Hint())                      // Not set
}

func TestErrorFormatterColorCodes(t *testing.T) {
	formatter := NewErrorFormatter(true, false)

	tests := []struct {
		severity    ErrorSeverity
		expectColor string
	}{
		{SeverityInfo, "\033[36m"},     // Cyan
		{SeverityWarning, "\033[33m"},  // Yellow
		{SeverityError, "\033[31m"},    // Red
		{SeverityCritical, "\033[35m"}, // Magenta
	}

	for _, tt := range tests {
		t.Run(tt.severity.String(), func(t *testing.T) {
			err := NewEnhancedError(
				PrefixPVM,
				CategoryVersion,
				"COLOR001",
				"Color test",
				nil,
				tt.severity,
			)

			formatted := formatter.Format(err)
			assert.Contains(t, formatted, tt.expectColor)
			assert.Contains(t, formatted, "\033[0m") // Reset code
		})
	}
}
