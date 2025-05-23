// ABOUTME: Auto-fix functionality for validation errors
// ABOUTME: Uses MCP sampling to collaborate with LLM on fixes

package validation

import (
	"context"
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/mcp/generation"
)

// AutoFixer provides automatic fixing of validation errors
type AutoFixer struct {
	validator      *Validator
	samplingClient *generation.SamplingClient
	enabled        bool
}

// NewAutoFixer creates a new auto-fixer instance
func NewAutoFixer(validator *Validator, samplingClient *generation.SamplingClient, enabled bool) *AutoFixer {
	return &AutoFixer{
		validator:      validator,
		samplingClient: samplingClient,
		enabled:        enabled,
	}
}

// FixError represents a suggested fix for a validation error
type FixError struct {
	Error       ValidationError `json:"error"`
	FixedCode   string          `json:"fixed_code"`
	Explanation string          `json:"explanation"`
	Confidence  float64         `json:"confidence"`
}

// AutoFix attempts to fix validation errors in the code
func (f *AutoFixer) AutoFix(ctx context.Context, code string, errors []ValidationError, projectPath string) ([]FixError, error) {
	if !f.enabled {
		return nil, nil
	}

	var fixes []FixError

	// Process fixable errors
	for _, err := range errors {
		if !err.Fixable {
			continue
		}

		fix, fixErr := f.fixError(ctx, code, err, projectPath)
		if fixErr != nil {
			// Log error but continue with other fixes
			continue
		}

		if fix != nil {
			fixes = append(fixes, *fix)
		}
	}

	return fixes, nil
}

// fixError attempts to fix a single validation error
func (f *AutoFixer) fixError(ctx context.Context, code string, err ValidationError, projectPath string) (*FixError, error) {
	switch err.Code {
	case "MISSING_SIGIL":
		return f.fixMissingSigil(ctx, code, err)
	case "TYPE_MISMATCH":
		return f.fixTypeMismatch(ctx, code, err, projectPath)
	case "UNDEFINED_VARIABLE":
		return f.fixUndefinedVariable(ctx, code, err)
	default:
		// For other errors, use sampling to get fix suggestions
		return f.fixWithSampling(ctx, code, err, projectPath)
	}
}

// fixMissingSigil fixes missing sigil errors
func (f *AutoFixer) fixMissingSigil(ctx context.Context, code string, err ValidationError) (*FixError, error) {
	lines := strings.Split(code, "\n")
	if err.Line <= 0 || err.Line > len(lines) {
		return nil, fmt.Errorf("invalid line number")
	}

	line := lines[err.Line-1]

	// Simple heuristic: if it's a scalar variable, add $
	if strings.Contains(line, "my") && !strings.Contains(line, "$") {
		// Find the variable name after "my"
		parts := strings.Fields(line)
		for i, part := range parts {
			if part == "my" && i+1 < len(parts) {
				varName := parts[i+1]
				// Add $ sigil
				fixedLine := strings.Replace(line, "my "+varName, "my $"+varName, 1)
				lines[err.Line-1] = fixedLine

				return &FixError{
					Error:       err,
					FixedCode:   strings.Join(lines, "\n"),
					Explanation: fmt.Sprintf("Added $ sigil to scalar variable %s", varName),
					Confidence:  0.9,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("could not determine fix for missing sigil")
}

// fixTypeMismatch attempts to fix type mismatch errors
func (f *AutoFixer) fixTypeMismatch(ctx context.Context, code string, err ValidationError, projectPath string) (*FixError, error) {
	// Use sampling to get type conversion suggestions
	prompt := fmt.Sprintf(
		"Fix the type mismatch error on line %d: %s\nCode:\n%s\n\nProvide the fixed line of code.",
		err.Line,
		err.Message,
		code,
	)

	response, samplingErr := f.samplingClient.Sample(ctx, prompt, projectPath)
	if samplingErr != nil {
		return nil, samplingErr
	}

	// Parse the response and create fix
	lines := strings.Split(code, "\n")
	if err.Line > 0 && err.Line <= len(lines) {
		lines[err.Line-1] = response.Content

		return &FixError{
			Error:       err,
			FixedCode:   strings.Join(lines, "\n"),
			Explanation: "Fixed type mismatch using type conversion",
			Confidence:  response.Confidence,
		}, nil
	}

	return nil, fmt.Errorf("could not apply type mismatch fix")
}

// fixUndefinedVariable attempts to fix undefined variable errors
func (f *AutoFixer) fixUndefinedVariable(ctx context.Context, code string, err ValidationError) (*FixError, error) {
	lines := strings.Split(code, "\n")
	if err.Line <= 0 || err.Line > len(lines) {
		return nil, fmt.Errorf("invalid line number")
	}

	// Extract variable name from error message
	varName := extractVariableName(err.Message)
	if varName == "" {
		return nil, fmt.Errorf("could not extract variable name")
	}

	// Add variable declaration before first use
	declarationLine := fmt.Sprintf("my %s;  # Auto-generated declaration", varName)

	// Insert declaration at the beginning of the scope
	insertLine := findScopeStart(lines, err.Line)
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertLine]...)
	newLines = append(newLines, declarationLine)
	newLines = append(newLines, lines[insertLine:]...)

	return &FixError{
		Error:       err,
		FixedCode:   strings.Join(newLines, "\n"),
		Explanation: fmt.Sprintf("Added declaration for undefined variable %s", varName),
		Confidence:  0.8,
	}, nil
}

// fixWithSampling uses MCP sampling to get fix suggestions from the LLM
func (f *AutoFixer) fixWithSampling(ctx context.Context, code string, err ValidationError, projectPath string) (*FixError, error) {
	// Create a prompt for the LLM
	prompt := fmt.Sprintf(
		"Fix the following Perl code error:\nError: %s (Line %d, Column %d)\nCode:\n%s\n\nProvide only the fixed code, no explanation.",
		err.Message,
		err.Line,
		err.Column,
		code,
	)

	// Use sampling to get fix suggestion
	response, samplingErr := f.samplingClient.Sample(ctx, prompt, projectPath)
	if samplingErr != nil {
		return nil, samplingErr
	}

	// Validate the fixed code
	validationResult, valErr := f.validator.ValidateCode(ctx, response.Content, projectPath)
	if valErr != nil {
		return nil, valErr
	}

	// Only return the fix if it actually fixes the error
	if validationResult.Valid || !containsError(validationResult.Errors, err.Code) {
		return &FixError{
			Error:       err,
			FixedCode:   response.Content,
			Explanation: "Fixed using LLM suggestion",
			Confidence:  response.Confidence,
		}, nil
	}

	return nil, fmt.Errorf("suggested fix did not resolve the error")
}

// Helper functions

func extractVariableName(errorMsg string) string {
	// Simple extraction - in real implementation would be more robust
	if strings.Contains(errorMsg, "$") {
		start := strings.Index(errorMsg, "$")
		end := start + 1
		for end < len(errorMsg) && isIdentChar(errorMsg[end]) {
			end++
		}
		return errorMsg[start:end]
	}
	return ""
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_'
}

func findScopeStart(lines []string, errorLine int) int {
	// Find the start of the current scope (simplified)
	for i := errorLine - 2; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasSuffix(line, "{") || i == 0 {
			return i + 1
		}
	}
	return 0
}

func containsError(errors []ValidationError, code string) bool {
	for _, err := range errors {
		if err.Code == code {
			return true
		}
	}
	return false
}
