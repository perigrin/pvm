// ABOUTME: Type definitions for the parser package
// ABOUTME: Contains all type-related constants and structures

package parser

import (
	"fmt"
	"strings"
)

// PSC Error codes
const (
	ErrParseFailed     = "801" // Failed to parse Perl code
	ErrTypeSyntaxError = "802" // Invalid type annotation syntax
	ErrParserInit      = "803" // Failed to initialize parser
)

// AnnotationKind represents the kind of type annotation
type AnnotationKind int

const (
	// VarAnnotation is a variable type annotation (my Type $var)
	VarAnnotation AnnotationKind = iota

	// SubParamAnnotation is a subroutine parameter type annotation (sub foo(Type $param))
	SubParamAnnotation

	// SubReturnAnnotation is a subroutine return type annotation (sub foo() -> Type)
	SubReturnAnnotation

	// MethodParamAnnotation is a method parameter type annotation (method foo(Type $param))
	MethodParamAnnotation

	// MethodReturnAnnotation is a method return type annotation (method foo() -> Type)
	MethodReturnAnnotation

	// AttrAnnotation is an attribute type annotation (field Type $attr)
	AttrAnnotation

	// TypeDeclAnnotation is a type declaration (type MyType = Type)
	TypeDeclAnnotation
)

// Position has been moved to parser.go

// TypeExpression represents a type expression in a type annotation
type TypeExpression struct {
	// Name is the name of the type
	Name string

	// Params are the parameters for parameterized types
	Params []*TypeExpression

	// Union is true if this is a union type (Type1|Type2)
	Union bool

	// Intersection is true if this is an intersection type (Type1&Type2)
	Intersection bool

	// Negation is true if this is a negation type (!Type)
	Negation bool

	// Position is the position of the type expression
	Pos Position
}

// String returns a string representation of a TypeExpression
func (te *TypeExpression) String() string {
	if te.Negation {
		return "!" + te.basicString()
	}
	return te.basicString()
}

// basicString returns a basic string representation of a TypeExpression
func (te *TypeExpression) basicString() string {
	if te.Union && len(te.Params) == 2 {
		return te.Params[0].String() + "|" + te.Params[1].String()
	} else if te.Union {
		// If it's marked as a union but doesn't have exactly 2 params,
		// just return the name which should include the | symbol
		return te.Name
	}

	if te.Intersection && len(te.Params) == 2 {
		return te.Params[0].String() + "&" + te.Params[1].String()
	} else if te.Intersection {
		// If it's marked as an intersection but doesn't have exactly 2 params,
		// just return the name which should include the & symbol
		return te.Name
	}

	if len(te.Params) > 0 {
		var paramStrs []string
		for _, param := range te.Params {
			paramStrs = append(paramStrs, param.String())
		}
		return te.Name + "[" + strings.Join(paramStrs, ", ") + "]"
	}

	return te.Name
}

// ParseTypeExpression parses a type expression string and returns a TypeExpression
func ParseTypeExpression(text string, pos Position) (*TypeExpression, error) {
	// Handle special cases for parameterized types that our simple parser might miss
	// For example, ArrayRef[Str] or HashRef[Str, Int]

	// First, clean up the type text by removing any extra spaces
	text = strings.TrimSpace(text)

	// Check if this is a parameterized type (containing square brackets)
	openBracket := strings.Index(text, "[")
	closeBracket := strings.LastIndex(text, "]")

	if openBracket > 0 && closeBracket > openBracket {
		typeName := strings.TrimSpace(text[:openBracket])
		paramText := text[openBracket+1 : closeBracket]

		// Parse the parameters
		var params []*TypeExpression
		if paramText != "" {
			// Split by comma for multiple parameters
			paramParts := strings.Split(paramText, ",")
			for _, part := range paramParts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}

				// Recursively parse each parameter
				paramPos := Position{
					Line:   pos.Line,
					Column: pos.Column + strings.Index(text, part),
				}
				param, err := ParseTypeExpression(part, paramPos)
				if err != nil {
					return nil, err
				}
				params = append(params, param)
			}
		}

		// Create a parameterized type
		return &TypeExpression{
			Name:   typeName,
			Params: params,
			Pos:    pos,
		}, nil
	}

	// Check for union types (Type1|Type2)
	pipeIndex := strings.Index(text, "|")
	if pipeIndex > 0 && pipeIndex < len(text)-1 {
		leftType := strings.TrimSpace(text[:pipeIndex])
		rightType := strings.TrimSpace(text[pipeIndex+1:])

		leftPos := Position{
			Line:   pos.Line,
			Column: pos.Column,
		}
		rightPos := Position{
			Line:   pos.Line,
			Column: pos.Column + pipeIndex + 1,
		}

		leftExpr, err := ParseTypeExpression(leftType, leftPos)
		if err != nil {
			return nil, err
		}

		rightExpr, err := ParseTypeExpression(rightType, rightPos)
		if err != nil {
			return nil, err
		}

		return &TypeExpression{
			Name:   text,
			Params: []*TypeExpression{leftExpr, rightExpr},
			Union:  true,
			Pos:    pos,
		}, nil
	}

	// Return a simple type
	return &TypeExpression{
		Name: text,
		Pos:  pos,
	}, nil
}

// TypeAnnotation has been moved to parser.go

// ParseError represents a parsing error
type ParseError struct {
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
func (e *ParseError) Error() string {
	return e.Message
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
