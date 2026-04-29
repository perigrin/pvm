// ABOUTME: User-friendly error types without error codes for better UX
// ABOUTME: Provides clear summaries, explanations, and actionable recovery steps

package errors

import (
	"fmt"
	"strings"
)

// ActionOption represents a recovery action the user can take
type ActionOption struct {
	Description string // What this action does
	Command     string // The actual command to run
	Risk        string // Risk level or considerations
}

// UserError represents a user-friendly error without error codes
type UserError struct {
	Summary     string            // One-line summary of what went wrong
	Explanation string            // Optional detailed explanation
	Actions     []ActionOption    // Concrete recovery actions
	Context     map[string]string // Additional context
}

// Error implements the error interface
func (e *UserError) Error() string {
	var msg strings.Builder

	// Clear summary
	msg.WriteString(e.Summary)

	// Explanation if provided
	if e.Explanation != "" {
		msg.WriteString("\n\n")
		msg.WriteString(e.Explanation)
	}

	// Actionable next steps
	if len(e.Actions) > 0 {
		msg.WriteString("\n\nWhat you can do:")
		for _, action := range e.Actions {
			msg.WriteString("\n• ")
			msg.WriteString(action.Description)
			if action.Command != "" {
				msg.WriteString(": ")
				msg.WriteString(action.Command)
			}
			if action.Risk != "" {
				msg.WriteString(" (")
				msg.WriteString(action.Risk)
				msg.WriteString(")")
			}
		}
	}

	return msg.String()
}

// InstallError represents a module or perl installation failure
type InstallError struct {
	Module      string
	Version     string
	Summary     string
	Explanation string
	Actions     []ActionOption
	TestResults *TestResults
}

// Error implements the error interface
func (e *InstallError) Error() string {
	var msg strings.Builder

	// Module/version info if available
	if e.Module != "" {
		msg.WriteString(e.Module)
		if e.Version != "" {
			msg.WriteString(" v")
			msg.WriteString(e.Version)
		}
		msg.WriteString(": ")
	}

	msg.WriteString(e.Summary)

	// Test results if available
	if e.TestResults != nil && len(e.TestResults.FailedTests) > 0 {
		msg.WriteString("\n\nFailed Tests:")
		for _, test := range e.TestResults.FailedTests {
			msg.WriteString("\n  • ")
			msg.WriteString(test.File)
			if test.Line > 0 {
				msg.WriteString(fmt.Sprintf(" line %d", test.Line))
			}
			msg.WriteString(": ")
			msg.WriteString(test.Error)
		}
	}

	if e.Explanation != "" {
		msg.WriteString("\n\n")
		msg.WriteString(e.Explanation)
	}

	if len(e.Actions) > 0 {
		msg.WriteString("\n\nWhat you can do:")
		for i, action := range e.Actions {
			msg.WriteString(fmt.Sprintf("\n  %d. ", i+1))
			if action.Command != "" {
				msg.WriteString(action.Command)
				if action.Description != "" {
					msg.WriteString("  # ")
					msg.WriteString(action.Description)
				}
			} else {
				msg.WriteString(action.Description)
			}
			if action.Risk != "" {
				msg.WriteString(" (")
				msg.WriteString(action.Risk)
				msg.WriteString(")")
			}
		}
	}

	return msg.String()
}

// TestResults contains parsed test output
type TestResults struct {
	Total       int
	Passed      int
	Failed      int
	Skipped     int
	FailedTests []FailedTest
	Output      string
	Summary     string
}

// FailedTest represents a single test failure
type FailedTest struct {
	File     string
	TestName string
	Error    string
	Line     int
}

// ConfigError represents a configuration issue
type ConfigError struct {
	Summary     string
	Explanation string
	Actions     []ActionOption
	ConfigFile  string
}

// Error implements the error interface
func (e *ConfigError) Error() string {
	var msg strings.Builder

	msg.WriteString(e.Summary)

	if e.ConfigFile != "" {
		msg.WriteString("\n\nConfiguration file: ")
		msg.WriteString(e.ConfigFile)
	}

	if e.Explanation != "" {
		msg.WriteString("\n\n")
		msg.WriteString(e.Explanation)
	}

	if len(e.Actions) > 0 {
		msg.WriteString("\n\nWhat you can do:")
		for _, action := range e.Actions {
			msg.WriteString("\n• ")
			msg.WriteString(action.Description)
			if action.Command != "" {
				msg.WriteString(": ")
				msg.WriteString(action.Command)
			}
		}
	}

	return msg.String()
}

// VersionError represents a version management issue
type VersionError struct {
	Version     string
	Summary     string
	Explanation string
	Actions     []ActionOption
	Available   []string // Available versions if relevant
}

// Error implements the error interface
func (e *VersionError) Error() string {
	var msg strings.Builder

	if e.Version != "" {
		msg.WriteString("Perl ")
		msg.WriteString(e.Version)
		msg.WriteString(": ")
	}

	msg.WriteString(e.Summary)

	if len(e.Available) > 0 {
		msg.WriteString("\n\nAvailable versions: ")
		msg.WriteString(strings.Join(e.Available[:min(5, len(e.Available))], ", "))
		if len(e.Available) > 5 {
			msg.WriteString(fmt.Sprintf(" (and %d more)", len(e.Available)-5))
		}
	}

	if e.Explanation != "" {
		msg.WriteString("\n\n")
		msg.WriteString(e.Explanation)
	}

	if len(e.Actions) > 0 {
		msg.WriteString("\n\nWhat you can do:")
		for _, action := range e.Actions {
			msg.WriteString("\n• ")
			msg.WriteString(action.Description)
			if action.Command != "" {
				msg.WriteString(": ")
				msg.WriteString(action.Command)
			}
		}
	}

	return msg.String()
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CommandError represents a command execution issue
type CommandError struct {
	Command     string
	Summary     string
	Explanation string
	Actions     []ActionOption
	Example     string
}

// Error implements the error interface
func (e *CommandError) Error() string {
	var msg strings.Builder

	msg.WriteString(e.Summary)

	if e.Explanation != "" {
		msg.WriteString("\n\n")
		msg.WriteString(e.Explanation)
	}

	if e.Example != "" {
		msg.WriteString("\n\nExample usage:\n  ")
		msg.WriteString(e.Example)
	}

	if len(e.Actions) > 0 {
		msg.WriteString("\n\nWhat you can do:")
		for _, action := range e.Actions {
			msg.WriteString("\n• ")
			msg.WriteString(action.Description)
			if action.Command != "" {
				msg.WriteString(": ")
				msg.WriteString(action.Command)
			}
		}
	}

	return msg.String()
}

// NewModuleTestFailureError creates an error for module test failures
func NewModuleTestFailureError(module string, version string, testResults *TestResults) *InstallError {
	actions := []ActionOption{
		{
			Description: "Skip tests",
			Command:     fmt.Sprintf("pvm module install --notest %s", module),
			Risk:        "low risk",
		},
		{
			Description: "See test details",
			Command:     fmt.Sprintf("pvm module install --verbose %s", module),
		},
		{
			Description: "Check dependencies",
			Command:     fmt.Sprintf("pvm module info %s", module),
		},
	}

	explanation := fmt.Sprintf("The module built successfully but %d out of %d tests failed.\nThis is often due to missing build dependencies or environment issues.",
		testResults.Failed, testResults.Total)

	return &InstallError{
		Module:      module,
		Version:     version,
		Summary:     "Module tests failed during installation",
		Explanation: explanation,
		Actions:     actions,
		TestResults: testResults,
	}
}

// NewPerlInstallFailureError creates an error for Perl installation failures
func NewPerlInstallFailureError(version string, stage string, cause error) *VersionError {
	var summary, explanation string
	var actions []ActionOption

	switch stage {
	case "configure":
		summary = fmt.Sprintf("installation failed during configuration")
		explanation = "The configure script couldn't complete. This often happens with older Perl versions on newer systems."
		actions = []ActionOption{
			{
				Description: "Try a newer Perl version",
				Command:     "pvm install 5.38.0",
			},
			{
				Description: "See full output",
				Command:     fmt.Sprintf("pvm install %s --verbose", version),
			},
		}
	case "compile":
		summary = fmt.Sprintf("installation failed during compilation")
		explanation = "The compilation process failed. This might be due to missing build tools or system libraries."
		actions = []ActionOption{
			{
				Description: "Check build dependencies",
				Command:     "pvm self doctor",
			},
			{
				Description: "See compilation errors",
				Command:     fmt.Sprintf("pvm install %s --verbose", version),
			},
		}
	case "test":
		summary = fmt.Sprintf("installation failed during tests")
		explanation = "Perl's test suite failed. This might indicate compatibility issues with your system."
		actions = []ActionOption{
			{
				Description: "Skip tests",
				Command:     fmt.Sprintf("pvm install %s --skip-tests", version),
			},
			{
				Description: "See test failures",
				Command:     fmt.Sprintf("pvm install %s --verbose", version),
			},
		}
	default:
		summary = "installation failed"
		explanation = ""
		actions = []ActionOption{
			{
				Description: "See full output",
				Command:     fmt.Sprintf("pvm install %s --verbose", version),
			},
		}
	}

	return &VersionError{
		Version:     version,
		Summary:     summary,
		Explanation: explanation,
		Actions:     actions,
	}
}

// NewVersionNotFoundError creates an error when a Perl version isn't available
func NewVersionNotFoundError(version string, available []string) *VersionError {
	return &VersionError{
		Version:   version,
		Summary:   "is not available",
		Available: available,
		Actions: []ActionOption{
			{
				Description: "See all available versions",
				Command:     "pvm list --remote",
			},
			{
				Description: "Install a similar version",
				Command:     fmt.Sprintf("pvm install %s", suggestSimilarVersion(version, available)),
			},
		},
	}
}

// suggestSimilarVersion finds a similar available version
func suggestSimilarVersion(requested string, available []string) string {
	// Simple heuristic: find the closest version
	if len(available) > 0 {
		// For now, just return the latest available
		return available[0]
	}
	return "5.38.0"
}

// NewModuleNotFoundError creates an error when a module isn't found
func NewModuleNotFoundError(module string) *InstallError {
	return &InstallError{
		Module:  module,
		Summary: "not found on CPAN",
		Actions: []ActionOption{
			{
				Description: "Search for similar modules",
				Command:     fmt.Sprintf("pvm module search %s", strings.Split(module, "::")[0]),
			},
			{
				Description: "Check spelling",
				Command:     "pvm module list",
			},
		},
	}
}

// NewPermissionError creates an error for permission issues
func NewPermissionError(path string) *UserError {
	return &UserError{
		Summary:     fmt.Sprintf("Cannot write to %s", path),
		Explanation: "This usually means PVM doesn't have permission to install system-wide.",
		Actions: []ActionOption{
			{
				Description: "Try with sudo",
				Command:     "sudo " + getCurrentCommand(),
			},
			{
				Description: "Install user-local",
				Command:     getCurrentCommand() + " --user",
			},
		},
	}
}

// getCurrentCommand is a helper to get the current command being executed
// In a real implementation, this would be passed in or retrieved from context
func getCurrentCommand() string {
	return "pvm install"
}
