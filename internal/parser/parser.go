// ABOUTME: Parser for Perl code with type annotations
// ABOUTME: Implements parsing capabilities for PSC type checking

package parser

import (
	"io"
	"os"
	"strings"

	"tamarou.com/pvm/internal/errors"
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

// NewParser returns a new Parser instance
// The implementation now uses a more robust TreeSitterParser
func NewParser() (Parser, error) {
	// Use the TreeSitterParser for enhanced type annotation support
	return NewTreeSitterParser()
}

// simpleParser is a simple implementation of the Parser interface
// This is a placeholder and would be replaced with the tree-sitter-based parser in a real implementation
// It is intentionally unused as we prefer to use the TreeSitterParser for enhanced functionality
// nolint:unused
type simpleParser struct {
}

// ParseFile implements the Parser interface
// nolint:unused
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
// nolint:unused
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
// nolint:unused
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
// This is intentionally unused as we use TreeSitterNode in the actual implementation
// nolint:unused
type simpleNode struct {
	nodeType string
	position Position
	endPos   Position
	children []Node
}

// Type implements the Node interface
// nolint:unused
func (n *simpleNode) Type() string {
	return n.nodeType
}

// Start implements the Node interface
// nolint:unused
func (n *simpleNode) Start() Position {
	return n.position
}

// End implements the Node interface
// nolint:unused
func (n *simpleNode) End() Position {
	return n.endPos
}

// Children implements the Node interface
// nolint:unused
func (n *simpleNode) Children() []Node {
	return n.children
}
