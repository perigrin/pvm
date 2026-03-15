// ABOUTME: TAP (Test Anything Protocol) parser for module test output
// ABOUTME: Extracts detailed test failure information for better error reporting

package modules

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"tamarou.com/pvm/internal/errors"
)

// TAPParser parses test output in TAP format
type TAPParser struct {
	verbose bool
}

// NewTAPParser creates a new TAP output parser
func NewTAPParser(verbose bool) *TAPParser {
	return &TAPParser{
		verbose: verbose,
	}
}

// ParseTestOutput parses test output and returns structured results
func (p *TAPParser) ParseTestOutput(output string) *errors.TestResults {
	results := &errors.TestResults{
		Output:      output,
		FailedTests: []errors.FailedTest{},
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	var currentFile string
	var inFailure bool
	var failureBuffer strings.Builder

	// Regular expressions for parsing TAP output
	planRe := regexp.MustCompile(`^1\.\.(\d+)`)
	okRe := regexp.MustCompile(`^ok\s+(\d+)`)
	notOkRe := regexp.MustCompile(`^not ok\s+(\d+)\s*-?\s*(.*)`)
	fileRe := regexp.MustCompile(`^#\s+(?:Running|Testing)\s+(.+\.t)`)
	diagRe := regexp.MustCompile(`^#\s+(.+)`)
	failedFileRe := regexp.MustCompile(`^(.+\.t)\s+.*\(Wstat:`)
	summaryRe := regexp.MustCompile(`^Files=(\d+),\s*Tests=(\d+)`)
	resultRe := regexp.MustCompile(`Result:\s*(\w+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for test file indicators
		if matches := fileRe.FindStringSubmatch(line); matches != nil {
			currentFile = matches[1]
			inFailure = false
			continue
		}

		// Check for failed file in summary
		if matches := failedFileRe.FindStringSubmatch(line); matches != nil {
			currentFile = matches[1]
			continue
		}

		// Check for test plan
		if matches := planRe.FindStringSubmatch(line); matches != nil {
			if total, err := strconv.Atoi(matches[1]); err == nil {
				results.Total += total
			}
			continue
		}

		// Check for successful test
		if matches := okRe.FindStringSubmatch(line); matches != nil {
			results.Passed++
			inFailure = false
			continue
		}

		// Check for failed test
		if matches := notOkRe.FindStringSubmatch(line); matches != nil {
			results.Failed++
			inFailure = true
			failureBuffer.Reset()

			testNum := matches[1]
			testName := matches[2]
			if testName == "" {
				testName = fmt.Sprintf("Test %s", testNum)
			}

			failed := errors.FailedTest{
				File:     currentFile,
				TestName: testName,
			}

			// Try to capture the next few lines for error details
			for i := 0; i < 5 && scanner.Scan(); i++ {
				nextLine := scanner.Text()
				if diagRe.MatchString(nextLine) {
					failureBuffer.WriteString(strings.TrimPrefix(nextLine, "# "))
					failureBuffer.WriteString(" ")
				} else {
					// Put the line back for processing
					failureBuffer.WriteString(nextLine)
					break
				}
			}

			failed.Error = strings.TrimSpace(failureBuffer.String())
			if failed.Error == "" {
				failed.Error = "Test failed"
			}

			results.FailedTests = append(results.FailedTests, failed)
			continue
		}

		// Check for diagnostic messages during failure
		if inFailure && diagRe.MatchString(line) {
			if len(results.FailedTests) > 0 {
				lastIdx := len(results.FailedTests) - 1
				results.FailedTests[lastIdx].Error += " " + strings.TrimPrefix(line, "# ")
			}
			continue
		}

		// Check for summary line
		if matches := summaryRe.FindStringSubmatch(line); matches != nil {
			if total, err := strconv.Atoi(matches[2]); err == nil {
				results.Total = total
			}
			continue
		}

		// Check for result line
		if matches := resultRe.FindStringSubmatch(line); matches != nil {
			if matches[1] == "PASS" {
				results.Summary = "All tests passed"
			} else {
				results.Summary = fmt.Sprintf("%d tests failed", results.Failed)
			}
			continue
		}

		// Look for common error patterns
		p.extractErrorPatterns(line, currentFile, results)
	}

	// Calculate skipped tests if totals don't match
	if results.Total > results.Passed+results.Failed {
		results.Skipped = results.Total - results.Passed - results.Failed
	}

	// Generate summary if not already set
	if results.Summary == "" {
		if results.Failed == 0 {
			results.Summary = "All tests passed"
		} else {
			results.Summary = fmt.Sprintf("%d out of %d tests failed", results.Failed, results.Total)
		}
	}

	return results
}

// extractErrorPatterns looks for common error patterns in test output
func (p *TAPParser) extractErrorPatterns(line, currentFile string, results *errors.TestResults) {
	// Common error patterns
	patterns := map[string]*regexp.Regexp{
		"Can't locate module":   regexp.MustCompile(`Can't locate (.+\.pm) in @INC`),
		"Undefined symbol":      regexp.MustCompile(`Undefined symbol:?\s*(.+)`),
		"Compilation failed":    regexp.MustCompile(`Compilation failed`),
		"Permission denied":     regexp.MustCompile(`Permission denied`),
		"No such file":          regexp.MustCompile(`No such file or directory`),
		"Syntax error":          regexp.MustCompile(`syntax error`),
		"XS compilation failed": regexp.MustCompile(`error:.*XS`),
		"Version mismatch":      regexp.MustCompile(`Version mismatch`),
	}

	for errorType, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(line); matches != nil {
			// Check if we should add this as a new failed test or update existing
			found := false
			for i := range results.FailedTests {
				if results.FailedTests[i].File == currentFile &&
					strings.Contains(results.FailedTests[i].Error, errorType) {
					found = true
					break
				}
			}

			if !found && currentFile != "" {
				failed := errors.FailedTest{
					File:     currentFile,
					TestName: "Build/Compilation",
					Error:    line,
				}
				results.FailedTests = append(results.FailedTests, failed)
				results.Failed++
				results.Total++
			}
		}
	}
}

// AnalyzeFailureCause analyzes test results to determine likely failure cause
func (p *TAPParser) AnalyzeFailureCause(results *errors.TestResults) string {
	if results == nil || len(results.FailedTests) == 0 {
		return ""
	}

	// Count different types of failures
	missingModules := 0
	xsFailures := 0
	compilationFailures := 0
	permissionErrors := 0

	for _, test := range results.FailedTests {
		errorLower := strings.ToLower(test.Error)

		if strings.Contains(errorLower, "can't locate") ||
			strings.Contains(errorLower, "@inc") {
			missingModules++
		}
		if strings.Contains(errorLower, "xs") ||
			strings.Contains(errorLower, "undefined symbol") {
			xsFailures++
		}
		if strings.Contains(errorLower, "compilation failed") ||
			strings.Contains(errorLower, "syntax error") {
			compilationFailures++
		}
		if strings.Contains(errorLower, "permission denied") {
			permissionErrors++
		}
	}

	// Determine the most likely cause
	if missingModules > 0 {
		return "Missing dependencies - some required modules are not installed"
	}
	if xsFailures > 0 {
		return "XS compilation issue - C extensions failed to build properly"
	}
	if compilationFailures > 0 {
		return "Build failure - module failed to compile"
	}
	if permissionErrors > 0 {
		return "Permission issue - insufficient permissions for testing"
	}

	// Default cause
	if results.Failed > results.Total/2 {
		return "Multiple test failures - likely environment or dependency issue"
	}

	return "Some tests failed - module may still work correctly"
}

// GetRecoveryActions suggests recovery actions based on failure analysis
func (p *TAPParser) GetRecoveryActions(module string, results *errors.TestResults) []errors.ActionOption {
	actions := []errors.ActionOption{
		{
			Description: "Skip tests",
			Command:     fmt.Sprintf("pvm module install --notest %s", module),
			Risk:        "low risk for most modules",
		},
		{
			Description: "See full test output",
			Command:     fmt.Sprintf("pvm module install --verbose %s", module),
		},
	}

	// Add specific actions based on failure cause
	cause := p.AnalyzeFailureCause(results)

	if strings.Contains(cause, "Missing dependencies") {
		actions = append(actions, errors.ActionOption{
			Description: "Install dependencies first",
			Command:     fmt.Sprintf("pvm module deps %s | xargs pvm module install", module),
		})
	}

	if strings.Contains(cause, "XS compilation") {
		actions = append(actions, errors.ActionOption{
			Description: "Check build tools",
			Command:     "pvm doctor --build-tools",
		})
	}

	if strings.Contains(cause, "Permission") {
		actions = append(actions, errors.ActionOption{
			Description: "Try with elevated permissions",
			Command:     fmt.Sprintf("sudo pvm module install %s", module),
		})
	}

	return actions
}
