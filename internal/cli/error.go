// ABOUTME: Error handling for CLI commands
// ABOUTME: Provides consistent error formatting and handling

package cli

import (
	"os"

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
	}
	output := ui.NewOutput(uiCtx)

	// Format the error based on its type
	if enhancedErr, ok := err.(*errors.EnhancedError); ok {
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

// FormatError formats an error for display without outputting it
// This is useful for components that need to format errors but handle display themselves
func FormatError(err error) string {
	// Use the errors package formatter for consistent formatting
	formatter := errors.NewErrorFormatter(true, Verbose) // Enable color and use verbose based on flag
	return formatter.Format(err)
}
