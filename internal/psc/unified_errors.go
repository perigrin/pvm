// ABOUTME: Unified error handling between PSC and PVI
// ABOUTME: Provides consistent error reporting across components

package psc

import (
	"fmt"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser"
)

// Error prefixes for components
const (
	// PrefixPSC is the error prefix for PSC
	PrefixPSC = "PSC"

	// PrefixPVI is the error prefix for PVI
	PrefixPVI = "PVI"
)

// Type error codes shared between PSC and PVI
const (
	// ErrTypeNotFound indicates a type was not found
	ErrTypeNotFound = "700"

	// ErrTypeInvalid indicates a type is invalid
	ErrTypeInvalid = "701"

	// ErrTypeConflict indicates a type conflict
	ErrTypeConflict = "702"

	// ErrTypeStorageError indicates a storage error
	ErrTypeStorageError = "703"

	// ErrTypeDefinitionError indicates an error with a type definition
	ErrTypeDefinitionError = "704"

	// ErrTypeAnnotationError indicates an error with a type annotation
	ErrTypeAnnotationError = "705"
)

// NewTypeError creates a new error related to type system
func NewTypeError(code string, message string, cause error) *errors.Error {
	return errors.NewTypeError(code, message, cause)
}

// ConvertParserErrorToTypeError converts a parser error to a type error
func ConvertParserErrorToTypeError(err error) *errors.Error {
	if e, ok := err.(*errors.Error); ok {
		return e
	}

	// Check if it's a parser.TypeCheckError
	if tcErr, ok := err.(parser.TypeCheckError); ok {
		return errors.NewTypeError(
			parser.ErrTypeValidationError,
			tcErr.Message,
			nil,
		).WithLocation(fmt.Sprintf("%s:%d:%d", tcErr.Path, tcErr.Line, tcErr.Column))
	}

	// Default case
	return errors.NewTypeError(
		parser.ErrTypeValidationError,
		err.Error(),
		nil,
	)
}

// FormatErrorsForOutput formats error messages for consistent output
func FormatErrorsForOutput(errs []error, verbose bool) []string {
	var messages []string

	for _, err := range errs {
		if e, ok := err.(*errors.Error); ok {
			if verbose {
				// Include details for verbose output
				messages = append(messages, fmt.Sprintf("[%s-%s] %s",
					e.Prefix(), e.Code(), e.Error()))

				// Include cause if available and verbose is enabled
				if e.Unwrap() != nil {
					messages = append(messages, fmt.Sprintf("  Caused by: %v", e.Unwrap()))
				}
			} else {
				messages = append(messages, e.Error())
			}
		} else {
			messages = append(messages, err.Error())
		}
	}

	return messages
}

// MapTypeCheckerErrorToPVI maps PSC type checker errors to PVI error codes
func MapTypeCheckerErrorToPVI(pscErr *errors.Error) *errors.Error {
	// Map PSC error codes to PVI error codes
	pviCode := ""

	switch pscErr.Code() {
	case parser.ErrTypeAnnotationMismatch:
		pviCode = ErrTypeInvalid
	case parser.ErrTypeInferenceError:
		pviCode = ErrTypeInvalid
	case parser.ErrTypeValidationError:
		pviCode = ErrTypeInvalid
	case parser.ErrTypeAssignmentError:
		pviCode = ErrTypeConflict
	case parser.ErrTypeFunctionError:
		pviCode = ErrTypeDefinitionError
	case parser.ErrTypeDeclarationError:
		pviCode = ErrTypeDefinitionError
	case parser.ErrTypeIncompatibleError:
		pviCode = ErrTypeConflict
	default:
		pviCode = ErrTypeDefinitionError
	}

	return errors.NewTypeError(
		pviCode,
		pscErr.Description(),
		pscErr.Unwrap(),
	).WithLocation(pscErr.Location())
}

// MapPVIErrorToPSC maps PVI errors to PSC error codes
func MapPVIErrorToPSC(pviErr *errors.Error) *errors.Error {
	// Map PVI error codes to PSC error codes
	pscCode := ""

	switch pviErr.Code() {
	case ErrTypeNotFound:
		pscCode = parser.ErrTypeValidationError
	case ErrTypeInvalid:
		pscCode = parser.ErrTypeValidationError
	case ErrTypeConflict:
		pscCode = parser.ErrTypeIncompatibleError
	case ErrTypeStorageError:
		pscCode = parser.ErrTypeValidationError
	case ErrTypeDefinitionError:
		pscCode = parser.ErrTypeDeclarationError
	case ErrTypeAnnotationError:
		pscCode = parser.ErrTypeAnnotationMismatch
	default:
		pscCode = parser.ErrTypeValidationError
	}

	return errors.NewTypeError(
		pscCode,
		pviErr.Description(),
		pviErr.Unwrap(),
	).WithLocation(pviErr.Location())
}

// GetTypeErrorsFromPSC extracts type errors from PSC check results
func GetTypeErrorsFromPSC(result *parser.TypeCheckResult) []error {
	var errs []error

	for _, errInfo := range result.Errors {
		// Create a new error with location information
		err := errors.NewTypeError(
			parser.ErrTypeValidationError,
			errInfo.Message,
			nil,
		).WithLocation(fmt.Sprintf("%s:%d:%d", errInfo.Path, errInfo.Line, errInfo.Column))

		errs = append(errs, err)
	}

	return errs
}

// Error codes are defined in the parser package and used directly
// No explicit registration is needed in the current error system
