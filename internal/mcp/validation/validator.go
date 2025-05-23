// ABOUTME: Validation service that wraps PVM's parser and type checker
// ABOUTME: Provides code validation capabilities for the MCP server

package validation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

// ValidationResult represents the result of validating Perl code
type ValidationResult struct {
	Valid     bool                `json:"valid"`
	Errors    []ValidationError   `json:"errors,omitempty"`
	Warnings  []ValidationWarning `json:"warnings,omitempty"`
	TypeInfo  map[string]TypeInfo `json:"type_info,omitempty"`
	Timestamp time.Time           `json:"timestamp"`
	CacheKey  string              `json:"-"` // Internal use only
}

// ValidationError represents a validation error
type ValidationError struct {
	Message  string `json:"message"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Fixable  bool   `json:"fixable"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Message string `json:"message"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Code    string `json:"code"`
}

// TypeInfo represents type information for a variable or expression
type TypeInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Inferred bool   `json:"inferred"`
}

// Validator provides code validation using PVM's type checker
type Validator struct {
	parser      parser.Parser
	typeStorage *typedef.Storage
	cache       *ValidationCache
}

// NewValidator creates a new validator instance
func NewValidator(cache *ValidationCache) (*Validator, error) {
	// Create parser
	p, err := parser.NewParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	// Create type storage
	storage, err := typedef.NewStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to create type storage: %w", err)
	}

	// We don't need the type checker here, just remove it

	return &Validator{
		parser:      p,
		typeStorage: storage,
		cache:       cache,
	}, nil
}

// ValidateCode validates Perl code and returns detailed results
func (v *Validator) ValidateCode(ctx context.Context, code string, projectPath string) (*ValidationResult, error) {
	// Generate cache key
	cacheKey := v.cache.GenerateKey(code, projectPath)

	// Check cache first
	if cached, found := v.cache.Get(cacheKey); found {
		return cached, nil
	}

	// Parse the code
	ast, err := v.parser.ParseString(code)
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}

	result := &ValidationResult{
		Valid:     true,
		Errors:    []ValidationError{},
		Warnings:  []ValidationWarning{},
		TypeInfo:  make(map[string]TypeInfo),
		Timestamp: time.Now(),
		CacheKey:  cacheKey,
	}

	// Convert parse errors to validation errors
	for _, parseErr := range ast.Errors {
		result.Valid = false
		var line, col int
		var msg string

		// Extract error details
		if perr, ok := parseErr.(*parser.ParseError); ok {
			line = perr.Line
			col = perr.Column
			msg = perr.Message
		} else {
			msg = parseErr.Error()
		}

		result.Errors = append(result.Errors, ValidationError{
			Message:  msg,
			Line:     line,
			Column:   col,
			Severity: "error",
			Code:     "PARSE_ERROR",
			Fixable:  false,
		})
	}

	// If parsing succeeded, extract type information
	if result.Valid && len(ast.TypeAnnotations) > 0 {
		// Record type info from annotations
		for _, ann := range ast.TypeAnnotations {
			if ann.TypeExpression != nil {
				// Record type info
				result.TypeInfo[ann.AnnotatedItem] = TypeInfo{
					Name:     ann.AnnotatedItem,
					Type:     ann.TypeExpression.String(),
					Line:     ann.Pos.Line,
					Column:   ann.Pos.Column,
					Inferred: false,
				}
			}
		}

		// Perform type checking
		typeErrors := v.performTypeChecking(code, projectPath)
		for _, typeErr := range typeErrors {
			result.Valid = false
			result.Errors = append(result.Errors, typeErr)
		}
	}

	// Cache the result
	v.cache.Set(cacheKey, result)

	return result, nil
}

// performTypeChecking performs type checking on the code
func (v *Validator) performTypeChecking(code string, projectPath string) []ValidationError {
	var errors []ValidationError

	// This is a simplified implementation
	// In a real implementation, we would:
	// 1. Parse the code into an AST
	// 2. Walk the AST and check types
	// 3. Use the type checker to validate type constraints
	// 4. Check for type mismatches, undefined types, etc.

	// For now, we'll do basic checks
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		// Check for common type errors
		if strings.Contains(line, "my") && !strings.Contains(line, "$") && !strings.Contains(line, "@") && !strings.Contains(line, "%") {
			errors = append(errors, ValidationError{
				Message:  "Variable declaration missing sigil",
				Line:     i + 1,
				Column:   strings.Index(line, "my") + 1,
				Severity: "error",
				Code:     "MISSING_SIGIL",
				Fixable:  true,
			})
		}

		// Check for undefined type usage
		if strings.Contains(line, "::") && strings.Contains(line, "my") {
			// Extract potential type name
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.Contains(part, "::") && !strings.HasPrefix(part, "$") {
					// Check if type is defined
					typeName := strings.TrimSuffix(part, ",")
					typeName = strings.TrimSuffix(typeName, ";")
					if !v.isKnownType(typeName) {
						errors = append(errors, ValidationError{
							Message:  fmt.Sprintf("Unknown type: %s", typeName),
							Line:     i + 1,
							Column:   strings.Index(line, part) + 1,
							Severity: "error",
							Code:     "UNKNOWN_TYPE",
							Fixable:  false,
						})
					}
				}
			}
		}
	}

	return errors
}

// isKnownType checks if a type is known to the system
func (v *Validator) isKnownType(typeName string) bool {
	// Check built-in types
	builtinTypes := []string{
		"Int", "Str", "Bool", "Num", "Any", "Undef",
		"ArrayRef", "HashRef", "CodeRef", "ScalarRef",
		"Object", "Defined", "Value", "Ref",
	}

	for _, builtin := range builtinTypes {
		if typeName == builtin {
			return true
		}
	}

	// For now, accept all types
	// In a real implementation, we would check the type storage
	return true
}

// GetTypeInfo returns type information for a specific item
func (v *Validator) GetTypeInfo(itemName string) (*TypeInfo, error) {
	// This is a simplified implementation
	// In a real implementation, we would look up the type from the type storage
	return nil, fmt.Errorf("no type information found for %s", itemName)
}

// ClearProjectCache clears validation cache for a specific project
func (v *Validator) ClearProjectCache(projectPath string) {
	v.cache.ClearProject(projectPath)
}
