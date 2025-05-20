// ABOUTME: Integration with PSC type checking
// ABOUTME: Connects the parser with the PSC component

package parser

import (
	"fmt"
	"os"
	"strings"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/typedef"
)

// TypeCheckResult represents the result of type checking a file
type TypeCheckResult struct {
	// Path is the path to the checked file
	Path string

	// Errors is a list of type errors found during checking
	Errors []TypeCheckError

	// TypeAnnotations is a list of type annotations found in the code
	TypeAnnotations []*TypeAnnotation
}

// TypeCheckError represents a type checking error
type TypeCheckError struct {
	// Message is the error message
	Message string

	// Line is the line number where the error occurred
	Line int

	// Column is the column number where the error occurred
	Column int

	// Path is the path to the file where the error occurred
	Path string
}

// Error implements the error interface
func (e TypeCheckError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.Path, e.Line, e.Column, e.Message)
}

// TypeChecker provides type checking functionality
type TypeChecker struct {
	// Parser is the parser used for parsing Perl code
	Parser Parser

	// TypeStore is the store for type definitions
	TypeStore *typedef.Storage
}

// NewTypeChecker creates a new TypeChecker
func NewTypeChecker() (*TypeChecker, error) {
	// Create a parser
	parser, err := NewParser()
	if err != nil {
		return nil, err
	}

	// Create a type store
	typeStore, err := typedef.NewStorage()
	if err != nil {
		return nil, err
	}

	return &TypeChecker{
		Parser:    parser,
		TypeStore: typeStore,
	}, nil
}

// CheckFile performs type checking on a Perl file
func (tc *TypeChecker) CheckFile(path string) (*TypeCheckResult, error) {
	// Parse the file
	ast, err := tc.Parser.ParseFile(path)
	if err != nil {
		return nil, err
	}

	// Check for parser errors
	if len(ast.Errors) > 0 {
		result := &TypeCheckResult{
			Path:            path,
			Errors:          []TypeCheckError{},
			TypeAnnotations: ast.TypeAnnotations,
		}

		// Convert parser errors to type check errors
		for _, parseErr := range ast.Errors {
			var typErr TypeCheckError

			if perr, ok := parseErr.(*ParseError); ok {
				typErr = TypeCheckError{
					Message: perr.Message,
					Line:    perr.Line,
					Column:  perr.Column,
					Path:    path,
				}
			} else {
				typErr = TypeCheckError{
					Message: parseErr.Error(),
					Line:    0,
					Column:  0,
					Path:    path,
				}
			}

			result.Errors = append(result.Errors, typErr)
		}

		return result, nil
	}

	// Type check annotations
	return tc.checkTypeAnnotations(ast)
}

// checkTypeAnnotations checks type annotations found in the AST
func (tc *TypeChecker) checkTypeAnnotations(ast *AST) (*TypeCheckResult, error) {
	result := &TypeCheckResult{
		Path:            ast.Path,
		Errors:          []TypeCheckError{},
		TypeAnnotations: ast.TypeAnnotations,
	}

	// Build module import map
	imports := make(map[string]bool)

	// This would involve scanning the AST for use statements
	// For the simplified implementation, we'll assume no imports

	// Check each type annotation
	for _, annotation := range ast.TypeAnnotations {
		// Verify the type exists
		err := tc.verifyType(annotation.TypeExpression, imports)
		if err != nil {
			result.Errors = append(result.Errors, TypeCheckError{
				Message: err.Error(),
				Line:    annotation.Pos.Line,
				Column:  annotation.Pos.Column,
				Path:    ast.Path,
			})
		}
	}

	return result, nil
}

// verifyType checks if a type exists and is valid
func (tc *TypeChecker) verifyType(typeExpr *TypeExpression, imports map[string]bool) error {
	// Check if it's a built-in type
	if isBuiltinType(typeExpr.Name) {
		return nil
	}

	// For parameterized types, check the parameters
	for _, param := range typeExpr.Params {
		if err := tc.verifyType(param, imports); err != nil {
			return err
		}
	}

	// For union and intersection types, the type itself is valid if the components are valid
	if typeExpr.Union || typeExpr.Intersection {
		return nil
	}

	// Check if it's an imported type
	parts := strings.Split(typeExpr.Name, "::")
	if len(parts) > 1 {
		moduleName := strings.Join(parts[:len(parts)-1], "::")
		if !imports[moduleName] {
			return fmt.Errorf("type %s comes from module %s which is not imported", typeExpr.Name, moduleName)
		}

		// Check if the module has a type definition
		_, err := tc.TypeStore.Load(moduleName)
		if err != nil {
			return fmt.Errorf("no type definition found for module %s", moduleName)
		}

		// In a real implementation, we would check if the specific type exists in the module
		return nil
	}

	// If we can't verify the type, we'll assume it's valid
	// In a real implementation, we would have a more thorough check
	return nil
}

// isBuiltinType returns true if the type is a built-in type
func isBuiltinType(typeName string) bool {
	builtinTypes := map[string]bool{
		"Any":       true,
		"Scalar":    true,
		"Str":       true,
		"Num":       true,
		"Int":       true,
		"Bool":      true,
		"Undef":     true,
		"Ref":       true,
		"ScalarRef": true,
		"ArrayRef":  true,
		"HashRef":   true,
		"CodeRef":   true,
		"List":      true,
		"Array":     true,
		"Hash":      true,
		"Code":      true,
		"Glob":      true,
		"Maybe":     true,
	}

	return builtinTypes[typeName]
}

// StripAnnotations removes type annotations from Perl code
func StripAnnotations(path string) (string, error) {
	// Create a parser
	p, err := NewParser()
	if err != nil {
		return "", err
	}

	// Parse the file
	_, err = p.ParseFile(path)
	if err != nil {
		return "", err
	}

	// Read the file content
	content, err := os.ReadFile(path)
	if err != nil {
		return "", errors.NewSystemError("001",
			"Failed to read file", err).
			WithLocation(path)
	}

	// This is a simplified implementation that doesn't actually strip annotations
	// In a real implementation, we would modify the content to remove annotations

	return string(content), nil
}
