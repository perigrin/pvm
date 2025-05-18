// ABOUTME: Integration between errors and logging
// ABOUTME: Provides helper functions for logging errors

package errors

import (
	"tamarou.com/pvm/internal/log"
)

// LogError logs an error at the appropriate level
func LogError(err error) {
	if err == nil {
		return
	}

	// Check if it's our structured error type
	if e, ok := err.(*Error); ok {
		// Choose log level based on category
		switch e.Category {
		case CategoryUserInput:
			// User input errors are usually not critical
			log.Warningf("User input error: %v", err)
		case CategorySystem:
			// System errors are usually critical
			log.Errorf("System error: %v", err)
		case CategoryExecution:
			// Execution errors can be varied in severity
			log.Errorf("Execution error: %v", err)
		case CategoryConfig:
			// Configuration errors are usually user fixable
			log.Warningf("Configuration error: %v", err)
		default:
			// Default to error level
			log.Errorf("Error: %v", err)
		}
	} else {
		// For standard errors, just log at error level
		log.Errorf("Error: %v", err)
	}
}

// LogDebug logs an error at debug level
func LogDebug(err error) {
	if err == nil {
		return
	}

	log.Debugf("Error: %v", err)
}

// LogFatal logs an error at fatal level and exits
func LogFatal(err error) {
	if err == nil {
		return
	}

	log.Fatalf("Fatal error: %v", err)
}

// LogErrorWithLocation logs an error with location information
func LogErrorWithLocation(err error, location string) {
	if err == nil {
		return
	}

	// Check if it's our structured error type
	if e, ok := err.(*Error); ok {
		// Add location if not already set
		if e.Location == "" {
			e.Location = location
		}
	}

	// Log the error
	LogError(err)
}
