// ABOUTME: Parser for Perl code with type annotations
// ABOUTME: Implements parsing capabilities for PSC type checking

package parser

import (
	"io"
	"os"
	"strings"

	"tamarou.com/pvm/internal/errors"
)

// PSC Error codes
const (
	ErrParseFailed     = "801" // Failed to parse Perl code
	ErrTypeSyntaxError = "802" // Invalid type annotation syntax
	ErrParserInit      = "803" // Failed to initialize parser
)

// Parser represents a parser for Perl code with type annotations
type Parser interface {
	// ParseFile parses a Perl file and returns its AST
	ParseFile(path string) (*AST, error)

	// ParseString parses a string containing Perl code and returns its AST
	ParseString(content string) (*AST, error)

	// ParseReader parses Perl code from a reader and returns its AST
	ParseReader(reader io.Reader) (*AST, error)
}

// AST represents the Abstract Syntax Tree of a parsed Perl file
type AST struct {
	// Path is the path to the parsed file
	Path string

	// Root is the root node of the AST
	Root Node

	// TypeAnnotations is a list of type annotations found in the code
	TypeAnnotations []*TypeAnnotation

	// Errors is a list of syntax errors found during parsing
	Errors []error
}

// Node represents a node in the AST
type Node interface {
	// Type returns the type of the node
	Type() string

	// Start returns the start position of the node
	Start() Position

	// End returns the end position of the node
	End() Position

	// Children returns the child nodes
	Children() []Node
}

// Position represents a position in the source code
type Position struct {
	Line   int
	Column int
	Offset int
}

// TypeAnnotation represents a type annotation found in the code
type TypeAnnotation struct {
	// AnnotatedItem is the item that has the type annotation
	AnnotatedItem string

	// TypeExpression is the type expression
	TypeExpression *TypeExpression

	// Position is the position of the type annotation
	Pos Position

	// Kind is the kind of annotation (variable, subroutine, method, attribute, etc.)
	Kind AnnotationKind
}

// AnnotationKind represents the kind of type annotation
type AnnotationKind int

const (
	// VarAnnotation is a variable type annotation (my Type $var)
	VarAnnotation AnnotationKind = iota

	// SubAnnotation is a subroutine parameter type annotation (sub foo(Type $param))
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

// NewParser returns a new Parser instance
// The implementation now uses a more robust TreeSitterParser
func NewParser() (Parser, error) {
	// Use the TreeSitterParser for enhanced type annotation support
	return NewTreeSitterParser()
}

// simpleParser is a simple implementation of the Parser interface
// This is a placeholder and would be replaced with a tree-sitter-based parser in a real implementation
type simpleParser struct {
}

// ParseFile implements the Parser interface
func (p *simpleParser) ParseFile(path string) (*AST, error) {
	// Read the file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewSystemError("001",
			"Failed to read file", err).
			WithLocation(path)
	}

	// Parse the content
	ast, err := p.ParseString(string(content))
	if err != nil {
		return nil, err
	}

	// Set the path
	ast.Path = path

	return ast, nil
}

// ParseString implements the Parser interface
func (p *simpleParser) ParseString(content string) (*AST, error) {
	// This is a placeholder implementation
	// In a real implementation, we would use tree-sitter to parse the code
	// For now, we just create a simple AST with some placeholder nodes

	ast := &AST{
		Root:            &simpleNode{nodeType: "root"},
		TypeAnnotations: []*TypeAnnotation{},
		Errors:          []error{},
	}

	// Simple parsing of type annotations
	// This is just a placeholder to demonstrate the structure
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lineNum := i + 1

		// Check for variable declarations with type annotations
		// Example: my Type $var
		if strings.Contains(line, "my ") && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			parts := strings.Fields(line)
			if len(parts) >= 3 && (parts[0] == "my" || parts[0] == "our" || parts[0] == "state") {
				varName := ""
				typeName := ""

				// Simple heuristic to detect type annotations
				// In a real implementation, we would use a proper parser
				for j, part := range parts[1:] {
					if strings.HasPrefix(part, "$") || strings.HasPrefix(part, "@") || strings.HasPrefix(part, "%") {
						varName = part
						if j > 0 {
							typeName = parts[j]

							// If it's a parameterized type that contains spaces,
							// we need to reconstruct it from multiple parts
							// For example: "HashRef[Str]" might be split into multiple tokens
							if strings.Contains(typeName, "[") && !strings.Contains(typeName, "]") {
								// Look for the matching closing bracket
								for k := j + 1; k < len(parts); k++ {
									if strings.Contains(parts[k], "]") {
										// Join all parts between the opening and closing brackets
										typeName = strings.Join(parts[j:k+1], " ")
										break
									}
								}
							}

							// Parse the type expression
							typeExpr, err := ParseTypeExpression(typeName, Position{
								Line:   lineNum,
								Column: strings.Index(line, typeName) + 1,
							})

							if err != nil {
								ast.Errors = append(ast.Errors, err)
								continue
							}

							// Add a type annotation
							annotation := &TypeAnnotation{
								AnnotatedItem:  varName,
								TypeExpression: typeExpr,
								Pos: Position{
									Line:   lineNum,
									Column: strings.Index(line, typeName) + 1,
								},
								Kind: VarAnnotation,
							}

							ast.TypeAnnotations = append(ast.TypeAnnotations, annotation)
						}
						break
					}
				}
			}
		}

		// Check for subroutine or method declarations with type annotations
		// Example: sub foo(Type $param) -> ReturnType
		// Example: method foo(Type $param) -> ReturnType
		if (strings.Contains(line, "sub ") || strings.Contains(line, "method ")) && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			// Determine if this is a method or subroutine
			isMethod := strings.Contains(line, "method ")

			// Check for return type annotations
			if strings.Contains(line, "->") {
				// This likely has a return type
				parts := strings.Split(line, "->")
				if len(parts) == 2 {
					returnType := strings.TrimSpace(parts[1])
					if strings.HasSuffix(returnType, "{") {
						returnType = strings.TrimSpace(returnType[:len(returnType)-1])
					}

					// Parse the type expression
					typeExpr, err := ParseTypeExpression(returnType, Position{
						Line:   lineNum,
						Column: strings.Index(line, "->") + 3,
					})

					if err != nil {
						ast.Errors = append(ast.Errors, err)
						continue
					}

					// Add a return type annotation
					var annotationKind AnnotationKind
					if isMethod {
						annotationKind = MethodReturnAnnotation
					} else {
						annotationKind = SubReturnAnnotation
					}

					annotation := &TypeAnnotation{
						AnnotatedItem:  "return",
						TypeExpression: typeExpr,
						Pos: Position{
							Line:   lineNum,
							Column: strings.Index(line, "->") + 3,
						},
						Kind: annotationKind,
					}

					ast.TypeAnnotations = append(ast.TypeAnnotations, annotation)
				}
			}

			// Check for parameter type annotations
			if strings.Contains(line, "(") && strings.Contains(line, ")") {
				// This is a very simplified check and would need a proper parser in a real implementation
				// In a real implementation, we would parse the parameter list properly
				paramStart := strings.Index(line, "(")
				paramEnd := strings.Index(line, ")")
				if paramStart >= 0 && paramEnd > paramStart {
					params := line[paramStart+1 : paramEnd]
					// This is a very simplified parsing of parameters
					// In a real implementation, we would use a proper parser
					paramParts := strings.Split(params, ",")
					for _, param := range paramParts {
						param = strings.TrimSpace(param)
						parts := strings.Fields(param)
						if len(parts) >= 2 && (strings.HasPrefix(parts[len(parts)-1], "$") ||
							strings.HasPrefix(parts[len(parts)-1], "@") ||
							strings.HasPrefix(parts[len(parts)-1], "%")) {
							paramName := parts[len(parts)-1]
							paramType := strings.Join(parts[:len(parts)-1], " ")

							// Parse the type expression
							typeExpr, err := ParseTypeExpression(paramType, Position{
								Line:   lineNum,
								Column: strings.Index(line, paramType) + 1,
							})

							if err != nil {
								ast.Errors = append(ast.Errors, err)
								continue
							}

							// Add a parameter type annotation
							var annotationKind AnnotationKind
							if isMethod {
								annotationKind = MethodParamAnnotation
							} else {
								annotationKind = SubParamAnnotation
							}

							annotation := &TypeAnnotation{
								AnnotatedItem:  paramName,
								TypeExpression: typeExpr,
								Pos: Position{
									Line:   lineNum,
									Column: strings.Index(line, paramType) + 1,
								},
								Kind: annotationKind,
							}

							ast.TypeAnnotations = append(ast.TypeAnnotations, annotation)
						}
					}
				}
			}
		}

		// Check for attribute declarations with type annotations
		// Example: field Type $attr
		if strings.Contains(line, "field ") && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			parts := strings.Fields(line)
			if len(parts) >= 3 && parts[0] == "field" {
				attrName := ""
				typeName := ""

				// Simple heuristic to detect type annotations
				for j, part := range parts[1:] {
					if strings.HasPrefix(part, "$") {
						attrName = part
						if j > 0 {
							typeName = parts[j]

							// If it's a parameterized type that contains spaces,
							// we need to reconstruct it from multiple parts
							if strings.Contains(typeName, "[") && !strings.Contains(typeName, "]") {
								// Look for the matching closing bracket
								for k := j + 1; k < len(parts); k++ {
									if strings.Contains(parts[k], "]") {
										// Join all parts between the opening and closing brackets
										typeName = strings.Join(parts[j:k+1], " ")
										break
									}
								}
							}

							// Parse the type expression
							typeExpr, err := ParseTypeExpression(typeName, Position{
								Line:   lineNum,
								Column: strings.Index(line, typeName) + 1,
							})

							if err != nil {
								ast.Errors = append(ast.Errors, err)
								continue
							}

							// Add a type annotation
							annotation := &TypeAnnotation{
								AnnotatedItem:  attrName,
								TypeExpression: typeExpr,
								Pos: Position{
									Line:   lineNum,
									Column: strings.Index(line, typeName) + 1,
								},
								Kind: AttrAnnotation,
							}

							ast.TypeAnnotations = append(ast.TypeAnnotations, annotation)
						}
						break
					}
				}
			}
		}
	}

	return ast, nil
}

// ParseReader implements the Parser interface
func (p *simpleParser) ParseReader(reader io.Reader) (*AST, error) {
	// Read the content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.NewSystemError("002",
			"Failed to read from reader", err)
	}

	// Parse the content
	return p.ParseString(string(content))
}

// simpleNode is a simple implementation of the Node interface
type simpleNode struct {
	nodeType string
	position Position
	endPos   Position
	children []Node
}

// Type implements the Node interface
func (n *simpleNode) Type() string {
	return n.nodeType
}

// Start implements the Node interface
func (n *simpleNode) Start() Position {
	return n.position
}

// End implements the Node interface
func (n *simpleNode) End() Position {
	return n.endPos
}

// Children implements the Node interface
func (n *simpleNode) Children() []Node {
	return n.children
}
