// ABOUTME: Error handling for CLI commands
// ABOUTME: Provides consistent error formatting and handling

package cli

import (
	"os"
	"strings"

	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// Re-export error constants for backward compatibility
const (
	// Component prefixes
	PrefixPVM = errors.PrefixPVM
	PrefixPVX = errors.PrefixPVX
	PrefixPVI = errors.PrefixPVI
	PrefixPSC = errors.PrefixPSC
	PrefixCFG = errors.PrefixCFG
	PrefixSYS = errors.PrefixSYS

	// Error categories
	CategoryConfig    = errors.CategoryConfig
	CategoryVersion   = errors.CategoryVersion
	CategoryModule    = errors.CategoryModule
	CategoryExecution = errors.CategoryExecution
	CategoryType      = errors.CategoryType
	CategorySystem    = errors.CategorySystem
	CategoryUserInput = errors.CategoryUserInput
)

// For backward compatibility, redefine Error as the errors.Error type
type Error = errors.Error

// NewError creates a new CLI error (wrapper around errors.New)
func NewError(prefix, category, code, message string, err error) *Error {
	return errors.New(prefix, category, code, message, err)
}

// HandleError handles a CLI error appropriately based on flags
func HandleError(err error) {
	if err == nil {
		return
	}

	// Log the error
	errors.LogError(err)

	// Format and display the error using UI framework
	displayError(err)

	// Exit with error code
	os.Exit(1)
}

// LogDebug logs an error at debug level if verbose mode is enabled
func LogDebug(format string, args ...interface{}) {
	if Verbose {
		log.Debugf(format, args...)
	}
}

// LogInfo logs an informational message
func LogInfo(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// LogWarning logs a warning message
func LogWarning(format string, args ...interface{}) {
	log.Warningf(format, args...)
}

// LogError logs an error message
func LogError(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// displayError formats and displays an error using the UI framework
func displayError(err error) {
	// Create a UI instance for error display
	uiCtx := &ui.UIContext{
		Writer:      os.Stderr,
		ErrorWriter: os.Stderr,
		ColorMode:   ui.ColorAuto,
		Quiet:       false,
		Verbose:     Verbose,
		Interactive: true,
		RawMarkdown: RawMarkdown,
	}
	output := ui.NewOutput(uiCtx)

	// Format the error based on its type
	// Handle user-friendly errors without showing error codes
	if userErr, ok := err.(*errors.UserError); ok {
		displayUserError(output, userErr)
	} else if installErr, ok := err.(*errors.InstallError); ok {
		displayInstallError(output, installErr)
	} else if configErr, ok := err.(*errors.ConfigError); ok {
		displayConfigError(output, configErr)
	} else if versionErr, ok := err.(*errors.VersionError); ok {
		displayVersionError(output, versionErr)
	} else if cmdErr, ok := err.(*errors.CommandError); ok {
		displayCommandError(output, cmdErr)
	} else if enhancedErr, ok := err.(*errors.EnhancedError); ok {
		displayEnhancedError(output, enhancedErr)
	} else if typedErr, ok := err.(*errors.Error); ok {
		displayTypedError(output, typedErr)
	} else {
		// Standard error display
		output.Error("%s", err.Error())
	}
}

// displayEnhancedError displays an enhanced error with rich formatting
func displayEnhancedError(output *ui.Output, err *errors.EnhancedError) {
	// Display the main error
	output.Error("%s-%s: %s", err.Prefix(), err.Code(), err.Message())

	// Display category if available
	if err.Category() != "" {
		output.Printf("  Category: %s\n", err.Category())
	}

	// Display severity with appropriate styling
	switch err.Severity() {
	case errors.SeverityInfo:
		output.Info("  Severity: %s", err.Severity().String())
	case errors.SeverityWarning:
		output.Warning("  Severity: %s", err.Severity().String())
	case errors.SeverityError, errors.SeverityCritical:
		output.Error("  Severity: %s", err.Severity().String())
	}

	// Display location if available
	if err.Location() != "" {
		output.Printf("  Location: %s\n", err.Location())
	}

	// Display context if available
	if context := err.Context(); len(context) > 0 {
		output.Printf("  Context:\n")
		for key, value := range context {
			output.Printf("    %s: %v\n", key, value)
		}
	}

	// Display recovery actions if available
	if actions := err.RecoveryActions(); len(actions) > 0 {
		output.Info("  Suggested actions:")
		for i, action := range actions {
			output.Printf("    %d. %s\n", i+1, action)
		}
	}

	// Display hint if available
	if err.Hint() != "" {
		output.Info("  Hint: %s", err.Hint())
	}

	// Display related errors if available
	if related := err.RelatedErrors(); len(related) > 0 {
		output.Printf("  Related errors:\n")
		for i, relErr := range related {
			output.Printf("    %d. %s-%s: %s\n", i+1, relErr.Prefix(), relErr.Code(), relErr.Message())
		}
	}
}

// displayTypedError displays a typed error with structured formatting
func displayTypedError(output *ui.Output, err *errors.Error) {
	// Display the main error
	output.Error("%s-%s: %s", err.Prefix(), err.Code(), err.Message())

	// Display category if available
	if err.Category() != "" {
		output.Printf("  Category: %s\n", err.Category())
	}

	// Display location if available
	if err.Location() != "" {
		output.Printf("  Location: %s\n", err.Location())
	}

	// Display hint if available
	if err.Hint() != "" {
		output.Info("  Hint: %s", err.Hint())
	}
}

// displayUserError displays a generic user error without error codes
func displayUserError(output *ui.Output, err *errors.UserError) {
	output.Error("%s", err.Summary)

	if err.Explanation != "" {
		output.Printf("\n%s\n", err.Explanation)
	}

	if len(err.Actions) > 0 {
		output.Info("\nWhat you can do:")
		for _, action := range err.Actions {
			output.Printf("• %s", action.Description)
			if action.Command != "" {
				output.Printf(": %s", action.Command)
			}
			if action.Risk != "" {
				output.Printf(" (%s)", action.Risk)
			}
			output.Printf("\n")
		}
	}
}

// displayInstallError displays an installation error without error codes
func displayInstallError(output *ui.Output, err *errors.InstallError) {
	// Display module info and summary
	if err.Module != "" {
		if err.Version != "" {
			output.Error("%s v%s: %s", err.Module, err.Version, err.Summary)
		} else {
			output.Error("%s: %s", err.Module, err.Summary)
		}
	} else {
		output.Error("%s", err.Summary)
	}

	// Show test failures if available
	if err.TestResults != nil && len(err.TestResults.FailedTests) > 0 {
		output.Printf("\nFailed Tests:")
		for _, test := range err.TestResults.FailedTests {
			output.Printf("  • %s", test.File)
			if test.Line > 0 {
				output.Printf(" line %d", test.Line)
			}
			output.Printf(": %s\n", test.Error)
		}
	}

	if err.Explanation != "" {
		output.Printf("\n%s\n", err.Explanation)
	}

	if len(err.Actions) > 0 {
		output.Info("\nWhat you can do:")
		for i, action := range err.Actions {
			if action.Command != "" {
				output.Printf("  %d. %s", i+1, action.Command)
				if action.Description != "" {
					output.Printf("  # %s", action.Description)
				}
			} else {
				output.Printf("  %d. %s", i+1, action.Description)
			}
			if action.Risk != "" {
				output.Printf(" (%s)", action.Risk)
			}
			output.Printf("\n")
		}
	}
}

// displayConfigError displays a configuration error without error codes
func displayConfigError(output *ui.Output, err *errors.ConfigError) {
	output.Error("%s", err.Summary)

	if err.ConfigFile != "" {
		output.Printf("\nConfiguration file: %s\n", err.ConfigFile)
	}

	if err.Explanation != "" {
		output.Printf("\n%s\n", err.Explanation)
	}

	if len(err.Actions) > 0 {
		output.Info("\nWhat you can do:")
		for _, action := range err.Actions {
			output.Printf("• %s", action.Description)
			if action.Command != "" {
				output.Printf(": %s", action.Command)
			}
			output.Printf("\n")
		}
	}
}

// displayVersionError displays a version management error without error codes
func displayVersionError(output *ui.Output, err *errors.VersionError) {
	if err.Version != "" {
		output.Error("Perl %s: %s", err.Version, err.Summary)
	} else {
		output.Error("%s", err.Summary)
	}

	if len(err.Available) > 0 {
		output.Printf("\nAvailable versions: ")
		maxShow := 5
		if len(err.Available) > maxShow {
			output.Printf("%s (and %d more)\n", strings.Join(err.Available[:maxShow], ", "), len(err.Available)-maxShow)
		} else {
			output.Printf("%s\n", strings.Join(err.Available, ", "))
		}
	}

	if err.Explanation != "" {
		output.Printf("\n%s\n", err.Explanation)
	}

	if len(err.Actions) > 0 {
		output.Info("\nWhat you can do:")
		for _, action := range err.Actions {
			output.Printf("• %s", action.Description)
			if action.Command != "" {
				output.Printf(": %s", action.Command)
			}
			output.Printf("\n")
		}
	}
}

// displayCommandError displays a command error without error codes
func displayCommandError(output *ui.Output, err *errors.CommandError) {
	output.Error("%s", err.Summary)

	if err.Explanation != "" {
		output.Printf("\n%s\n", err.Explanation)
	}

	if err.Example != "" {
		output.Printf("\nExample usage:\n  %s\n", err.Example)
	}

	if len(err.Actions) > 0 {
		output.Info("\nWhat you can do:")
		for _, action := range err.Actions {
			output.Printf("• %s", action.Description)
			if action.Command != "" {
				output.Printf(": %s", action.Command)
			}
			output.Printf("\n")
		}
	}
}

// FormatError formats an error for display without outputting it
// This is useful for components that need to format errors but handle display themselves
func FormatError(err error) string {
	// Use the errors package formatter for consistent formatting
	formatter := errors.NewErrorFormatter(true, Verbose) // Enable color and use verbose based on flag
	return formatter.Format(err)
}

// sanitizeForTerminal strips control bytes from s so attacker-influenceable
// input (e.g. PATH directory names that flow into "detected_conflicts")
// cannot inject ANSI escape sequences, overwrite output via \r, or fake
// extra lines via \n. Tabs (0x09) and newlines (0x0a) are stripped along
// with the rest because we apply this to single-line VALUES — the
// surrounding format strings provide their own line breaks. Each stripped
// byte is replaced with '?' so the user sees that something was removed
// rather than getting silently truncated content.
func sanitizeForTerminal(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return '?'
		}
		return r
	}, s)
}

// sanitizeStrings applies sanitizeForTerminal to every element of in,
// returning a new slice (in is not mutated).
func sanitizeStrings(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = sanitizeForTerminal(s)
	}
	return out
}

// DisplayEnhancedError displays an enhanced error using the provided UI output
// This is used by commands that want to show detailed error information
func DisplayEnhancedError(output *ui.Output, err *errors.EnhancedError) {
	// Display the main error with appropriate styling
	switch err.Severity() {
	case errors.SeverityInfo:
		output.Info("%s-%s: %s", err.Prefix(), err.Code(), err.Message())
	case errors.SeverityWarning:
		output.Warning("%s-%s: %s", err.Prefix(), err.Code(), err.Message())
	case errors.SeverityError, errors.SeverityCritical:
		output.Error("%s-%s: %s", err.Prefix(), err.Code(), err.Message())
	default:
		output.Error("%s-%s: %s", err.Prefix(), err.Code(), err.Message())
	}

	output.Println()

	// Display context information if available. Printf does not append a
	// trailing newline, so each line-oriented field needs an explicit \n;
	// otherwise they collapse into a single line of output.
	//
	// Every value sourced from context is sanitized before printing —
	// some context values (notably "detected_conflicts", which is built
	// from PATH directory names) carry attacker-influenceable bytes.
	// Without sanitization a malicious PATH entry could inject ANSI
	// escape codes, overwrite output via \r, or fake fresh lines via \n.
	if context := err.Context(); len(context) > 0 {
		if cmdName, ok := context["command"].(string); ok {
			output.Printf("Command: %s\n", sanitizeForTerminal(cmdName))
		}
		if desc, ok := context["description"].(string); ok {
			output.Printf("Issue: %s\n", sanitizeForTerminal(desc))
		}
		if shell, ok := context["detected_shell"].(string); ok {
			output.Printf("Detected shell: %s\n", sanitizeForTerminal(shell))
		}
		if conflicts, ok := context["detected_conflicts"].([]string); ok && len(conflicts) > 0 {
			output.Printf("Conflicting version managers: %v\n", sanitizeStrings(conflicts))
		}
		output.Println()
	}

	// Display recovery actions with clear formatting. Recovery actions can
	// interpolate paths derived from $HOME (via WithRecoveryAction +
	// fmt.Sprintf), so the same sanitation applies.
	if actions := err.RecoveryActions(); len(actions) > 0 {
		output.Info("To fix this issue:")
		for i, action := range actions {
			action = sanitizeForTerminal(action)
			// Skip numbered actions for examples
			if strings.HasPrefix(action, "Examples") {
				output.Println()
				output.Info("%s", action)
				continue
			}
			// Format action items nicely
			if strings.HasPrefix(action, "  ") {
				// This is an example or sub-item, show it indented
				output.Printf("%s\n", action)
			} else {
				// This is a main action item, number it
				output.Printf("%d. %s\n", i+1, action)
			}
		}
		output.Println()
	}

	// Display hint if available
	if hint := err.Hint(); hint != "" {
		output.Info("Hint: %s", sanitizeForTerminal(hint))
		output.Println()
	}
}
