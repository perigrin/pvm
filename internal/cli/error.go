// ABOUTME: Error handling for CLI commands
// ABOUTME: Provides consistent error formatting and handling

package cli

import (
	"fmt"
	"os"

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

	// Print to stderr for user visibility
	fmt.Fprintln(os.Stderr, err.Error())

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
