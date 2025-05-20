// ABOUTME: Tree-sitter integration for Perl parsing
// ABOUTME: Outlines integration with tree-sitter for enhanced parsing

package parser

import (
	"io"
	"os"
	"strings"
	"sync"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// TreeSitterParser is a Parser implementation that uses tree-sitter
// This is an enhanced implementation that simulates tree-sitter functionality
// while providing realistic type annotation parsing
type TreeSitterParser struct {
	// In a real implementation, these would be actual tree-sitter structures
	// parser        *sitter.Parser
	// perlLanguage  *sitter.Language
	// typeQueries   *sitter.Query

	// Mutex for thread safety
	sync.Mutex

	// Debug mode for verbose logging
	Debug bool
}

// TreeSitterNode represents a node in the tree-sitter syntax tree
type TreeSitterNode struct {
	NodeType     string
	Text         string
	StartPos     Position
	EndPos       Position
	NodeChildren []*TreeSitterNode
}

// Type implements the Node interface
func (n *TreeSitterNode) Type() string {
	return n.NodeType
}

// Start implements the Node interface
func (n *TreeSitterNode) Start() Position {
	return n.StartPos
}

// End implements the Node interface
func (n *TreeSitterNode) End() Position {
	return n.EndPos
}

// Children implements the Node interface
func (n *TreeSitterNode) Children() []Node {
	result := make([]Node, len(n.NodeChildren))
	for i, child := range n.NodeChildren {
		result[i] = child
	}
	return result
}

// NewTreeSitterParser creates a new TreeSitterParser
func NewTreeSitterParser() (Parser, error) {
	// Create a new TreeSitterParser
	parser := &TreeSitterParser{
		Debug: false,
	}

	log.Debugf("Created new TreeSitterParser instance")
	return parser, nil
}

// ParseFile implements the Parser interface
func (p *TreeSitterParser) ParseFile(path string) (*AST, error) {
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
func (p *TreeSitterParser) ParseString(content string) (*AST, error) {
	p.Lock()
	defer p.Unlock()

	// Create a new AST
	ast := &AST{
		TypeAnnotations: []*TypeAnnotation{},
		Errors:          []error{},
	}

	// Parse the source code and build the syntax tree
	// In a real implementation, this would use tree-sitter to parse the content
	// For now, we simulate the parsing by creating a simplified syntax tree
	root, errs := p.parseSource(content)
	if len(errs) > 0 {
		ast.Errors = append(ast.Errors, errs...)
	}

	// Set the root node
	ast.Root = root

	// Extract type annotations from the syntax tree
	annotations, errs := p.extractTypeAnnotations(content, root)
	ast.TypeAnnotations = append(ast.TypeAnnotations, annotations...)
	ast.Errors = append(ast.Errors, errs...)

	if p.Debug {
		log.Debugf("Found %d type annotations", len(ast.TypeAnnotations))
		for i, ann := range ast.TypeAnnotations {
			log.Debugf("Annotation %d: %s %s at line %d",
				i+1, ann.AnnotatedItem, ann.TypeExpression.String(), ann.Pos.Line)
		}
	}

	return ast, nil
}

// ParseReader implements the Parser interface
func (p *TreeSitterParser) ParseReader(reader io.Reader) (*AST, error) {
	// Read the content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.NewSystemError("002",
			"Failed to read from reader", err)
	}

	// Parse the content
	return p.ParseString(string(content))
}

// parseSource parses the source code and builds a syntax tree
// This is a simulation of what tree-sitter would do in a real implementation
func (p *TreeSitterParser) parseSource(content string) (Node, []error) {
	var errors []error

	// Split the content into lines for line-by-line processing
	lines := strings.Split(content, "\n")

	// Create a root node
	root := &TreeSitterNode{
		NodeType:     "program",
		Text:         content,
		StartPos:     Position{Line: 1, Column: 1, Offset: 0},
		EndPos:       Position{Line: len(lines), Column: len(lines[len(lines)-1]) + 1, Offset: len(content)},
		NodeChildren: []*TreeSitterNode{},
	}

	// Process each line to build a simplified syntax tree
	offset := 0
	for i, line := range lines {
		lineNum := i + 1

		// Process leading whitespace
		indentLen := len(line) - len(strings.TrimLeft(line, " \t"))

		// Skip empty lines or comment lines
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			offset += len(line) + 1 // +1 for the newline
			continue
		}

		// Calculate starting column (1-indexed)
		startCol := indentLen + 1

		// Create a node for this line
		node := &TreeSitterNode{
			NodeType:     p.determineNodeType(trimmed),
			Text:         trimmed,
			StartPos:     Position{Line: lineNum, Column: startCol, Offset: offset + indentLen},
			EndPos:       Position{Line: lineNum, Column: len(line) + 1, Offset: offset + len(line)},
			NodeChildren: []*TreeSitterNode{},
		}

		// Add the node to the root
		root.NodeChildren = append(root.NodeChildren, node)

		// Add children nodes for complex statements
		p.processNodeChildren(node, trimmed, lineNum, offset+indentLen)

		// Update offset for the next line
		offset += len(line) + 1 // +1 for the newline
	}

	return root, errors
}

// determineNodeType determines the AST node type based on the line content
func (p *TreeSitterParser) determineNodeType(line string) string {
	// Basic parsing to determine node types
	words := strings.Fields(line)
	if len(words) == 0 {
		return "empty_statement"
	}

	switch words[0] {
	case "package":
		return "package_declaration"
	case "use":
		return "use_statement"
	case "require":
		return "require_statement"
	case "my", "our", "state":
		return "variable_declaration"
	case "sub":
		return "subroutine_declaration"
	case "method":
		return "method_declaration"
	case "class":
		return "class_declaration"
	case "role":
		return "role_declaration"
	case "field":
		return "field_declaration"
	case "type":
		return "type_declaration"
	default:
		switch {
		case strings.HasPrefix(line, "if ") || strings.HasPrefix(line, "if("):
			return "if_statement"
		case strings.HasPrefix(line, "while ") || strings.HasPrefix(line, "while("):
			return "while_statement"
		case strings.HasPrefix(line, "for ") || strings.HasPrefix(line, "for("):
			return "for_statement"
		case strings.HasPrefix(line, "foreach ") || strings.HasPrefix(line, "foreach("):
			return "foreach_statement"
		}
	}

	return "expression_statement"
}

// processNodeChildren adds child nodes to a parent node based on its type
func (p *TreeSitterParser) processNodeChildren(node *TreeSitterNode, content string, lineNum int, offset int) {
	// In a real tree-sitter implementation, this would be done by the parser
	// We're simulating it with a simple approach

	switch node.NodeType {
	case "variable_declaration":
		// Look for variable name and type annotations
		parts := strings.Fields(content)
		if len(parts) >= 3 {
			// Extract type annotations and variable name
			varDeclNode := &TreeSitterNode{
				NodeType:     "variable_declarator",
				Text:         parts[0],
				StartPos:     Position{Line: lineNum, Column: offset + 1, Offset: offset},
				EndPos:       Position{Line: lineNum, Column: offset + len(parts[0]) + 1, Offset: offset + len(parts[0])},
				NodeChildren: []*TreeSitterNode{},
			}
			node.NodeChildren = append(node.NodeChildren, varDeclNode)

			// The second part might be a type annotation
			typePos := strings.Index(content, parts[1])
			if typePos >= 0 {
				typeNode := &TreeSitterNode{
					NodeType:     "type_annotation",
					Text:         parts[1],
					StartPos:     Position{Line: lineNum, Column: offset + typePos + 1, Offset: offset + typePos},
					EndPos:       Position{Line: lineNum, Column: offset + typePos + len(parts[1]) + 1, Offset: offset + typePos + len(parts[1])},
					NodeChildren: []*TreeSitterNode{},
				}
				node.NodeChildren = append(node.NodeChildren, typeNode)
			}

			// The third part is usually the variable name
			varPos := strings.Index(content, parts[2])
			if varPos >= 0 {
				varNode := &TreeSitterNode{
					NodeType:     "identifier",
					Text:         parts[2],
					StartPos:     Position{Line: lineNum, Column: offset + varPos + 1, Offset: offset + varPos},
					EndPos:       Position{Line: lineNum, Column: offset + varPos + len(parts[2]) + 1, Offset: offset + varPos + len(parts[2])},
					NodeChildren: []*TreeSitterNode{},
				}
				node.NodeChildren = append(node.NodeChildren, varNode)
			}
		}

	case "subroutine_declaration", "method_declaration":
		// Parse subroutine/method name, parameters, and return type
		parts := strings.Fields(content)
		if len(parts) >= 2 {
			// Extract subroutine name
			subNamePos := strings.Index(content, parts[1])
			if subNamePos >= 0 {
				nameNode := &TreeSitterNode{
					NodeType:     "identifier",
					Text:         parts[1],
					StartPos:     Position{Line: lineNum, Column: offset + subNamePos + 1, Offset: offset + subNamePos},
					EndPos:       Position{Line: lineNum, Column: offset + subNamePos + len(parts[1]) + 1, Offset: offset + subNamePos + len(parts[1])},
					NodeChildren: []*TreeSitterNode{},
				}
				node.NodeChildren = append(node.NodeChildren, nameNode)
			}

			// Extract parameter list
			paramStart := strings.Index(content, "(")
			paramEnd := strings.Index(content, ")")
			if paramStart >= 0 && paramEnd > paramStart {
				paramText := content[paramStart : paramEnd+1]
				paramNode := &TreeSitterNode{
					NodeType:     "parameter_list",
					Text:         paramText,
					StartPos:     Position{Line: lineNum, Column: offset + paramStart + 1, Offset: offset + paramStart},
					EndPos:       Position{Line: lineNum, Column: offset + paramEnd + 2, Offset: offset + paramEnd + 1},
					NodeChildren: []*TreeSitterNode{},
				}
				node.NodeChildren = append(node.NodeChildren, paramNode)

				// Process each parameter
				params := content[paramStart+1 : paramEnd]
				if len(params) > 0 {
					paramsList := strings.Split(params, ",")
					paramOffset := offset + paramStart + 1

					for _, param := range paramsList {
						param = strings.TrimSpace(param)
						if param == "" {
							continue
						}

						paramPos := strings.Index(content[paramStart+1:], param) + paramStart + 1
						if paramPos >= paramStart {
							// Check if the parameter has a type annotation
							paramParts := strings.Fields(param)
							if len(paramParts) >= 2 {
								// First part might be a type annotation
								typePos := paramOffset + strings.Index(param, paramParts[0])
								if typePos >= 0 {
									typeNode := &TreeSitterNode{
										NodeType:     "type_annotation",
										Text:         paramParts[0],
										StartPos:     Position{Line: lineNum, Column: typePos + 1, Offset: typePos},
										EndPos:       Position{Line: lineNum, Column: typePos + len(paramParts[0]) + 1, Offset: typePos + len(paramParts[0])},
										NodeChildren: []*TreeSitterNode{},
									}
									paramNode.NodeChildren = append(paramNode.NodeChildren, typeNode)
								}

								// Second part is usually the parameter name
								varPos := paramOffset + strings.Index(param, paramParts[len(paramParts)-1])
								if varPos >= 0 {
									varNode := &TreeSitterNode{
										NodeType:     "identifier",
										Text:         paramParts[len(paramParts)-1],
										StartPos:     Position{Line: lineNum, Column: varPos + 1, Offset: varPos},
										EndPos:       Position{Line: lineNum, Column: varPos + len(paramParts[len(paramParts)-1]) + 1, Offset: varPos + len(paramParts[len(paramParts)-1])},
										NodeChildren: []*TreeSitterNode{},
									}
									paramNode.NodeChildren = append(paramNode.NodeChildren, varNode)
								}
							}
						}

						paramOffset += len(param) + 1 // +1 for the comma
					}
				}
			}

			// Extract return type
			returnArrow := strings.Index(content, "->")
			if returnArrow >= 0 {
				returnTypeText := content[returnArrow+2:]
				returnTypeText = strings.TrimSpace(returnTypeText)

				// Remove trailing curly brace or semicolon if present
				if idx := strings.Index(returnTypeText, "{"); idx != -1 {
					returnTypeText = strings.TrimSpace(returnTypeText[:idx])
				}
				if idx := strings.Index(returnTypeText, ";"); idx != -1 {
					returnTypeText = strings.TrimSpace(returnTypeText[:idx])
				}

				returnNode := &TreeSitterNode{
					NodeType:     "return_type_annotation",
					Text:         returnTypeText,
					StartPos:     Position{Line: lineNum, Column: offset + returnArrow + 3, Offset: offset + returnArrow + 2},
					EndPos:       Position{Line: lineNum, Column: offset + returnArrow + 3 + len(returnTypeText), Offset: offset + returnArrow + 2 + len(returnTypeText)},
					NodeChildren: []*TreeSitterNode{},
				}
				node.NodeChildren = append(node.NodeChildren, returnNode)
			}
		}

	case "field_declaration":
		// Parse field name and type annotation
		parts := strings.Fields(content)
		if len(parts) >= 3 {
			// Extract type annotation
			typePos := strings.Index(content, parts[1])
			if typePos >= 0 {
				typeNode := &TreeSitterNode{
					NodeType:     "type_annotation",
					Text:         parts[1],
					StartPos:     Position{Line: lineNum, Column: offset + typePos + 1, Offset: offset + typePos},
					EndPos:       Position{Line: lineNum, Column: offset + typePos + len(parts[1]) + 1, Offset: offset + typePos + len(parts[1])},
					NodeChildren: []*TreeSitterNode{},
				}
				node.NodeChildren = append(node.NodeChildren, typeNode)
			}

			// Extract field name
			fieldPos := strings.Index(content, parts[2])
			if fieldPos >= 0 {
				fieldNode := &TreeSitterNode{
					NodeType:     "identifier",
					Text:         parts[2],
					StartPos:     Position{Line: lineNum, Column: offset + fieldPos + 1, Offset: offset + fieldPos},
					EndPos:       Position{Line: lineNum, Column: offset + fieldPos + len(parts[2]) + 1, Offset: offset + fieldPos + len(parts[2])},
					NodeChildren: []*TreeSitterNode{},
				}
				node.NodeChildren = append(node.NodeChildren, fieldNode)
			}
		}

	case "type_declaration":
		// Parse type name and definition
		parts := strings.Fields(content)
		if len(parts) >= 4 && parts[2] == "=" {
			// Extract type name
			typeNamePos := strings.Index(content, parts[1])
			if typeNamePos >= 0 {
				typeNameNode := &TreeSitterNode{
					NodeType:     "identifier",
					Text:         parts[1],
					StartPos:     Position{Line: lineNum, Column: offset + typeNamePos + 1, Offset: offset + typeNamePos},
					EndPos:       Position{Line: lineNum, Column: offset + typeNamePos + len(parts[1]) + 1, Offset: offset + typeNamePos + len(parts[1])},
					NodeChildren: []*TreeSitterNode{},
				}
				node.NodeChildren = append(node.NodeChildren, typeNameNode)
			}

			// Extract type definition
			typeDefPos := strings.Index(content, parts[3])
			if typeDefPos >= 0 {
				typeDefText := content[typeDefPos:]
				// Remove trailing semicolon if present
				if idx := strings.Index(typeDefText, ";"); idx != -1 {
					typeDefText = typeDefText[:idx]
				}

				typeDefNode := &TreeSitterNode{
					NodeType:     "type_definition",
					Text:         typeDefText,
					StartPos:     Position{Line: lineNum, Column: offset + typeDefPos + 1, Offset: offset + typeDefPos},
					EndPos:       Position{Line: lineNum, Column: offset + typeDefPos + len(typeDefText) + 1, Offset: offset + typeDefPos + len(typeDefText)},
					NodeChildren: []*TreeSitterNode{},
				}
				node.NodeChildren = append(node.NodeChildren, typeDefNode)
			}
		}
	}
}

// extractTypeAnnotations extracts type annotations from the syntax tree
func (p *TreeSitterParser) extractTypeAnnotations(content string, root Node) ([]*TypeAnnotation, []error) {
	var annotations []*TypeAnnotation
	var errors []error

	// Split content into lines for easier parsing
	lines := strings.Split(content, "\n")

	// Process the root node and its children
	p.processNode(content, lines, root, &annotations, &errors)

	return annotations, errors
}

// processNode processes a node and its children to extract type annotations
func (p *TreeSitterParser) processNode(content string, lines []string, node Node, annotations *[]*TypeAnnotation, errors *[]error) {
	// Process based on node type
	switch node.Type() {
	case "variable_declaration":
		p.processVariableDeclaration(content, lines, node, annotations, errors)
	case "subroutine_declaration", "method_declaration":
		p.processSubroutineDeclaration(content, lines, node, annotations, errors)
	case "field_declaration":
		p.processFieldDeclaration(content, lines, node, annotations, errors)
	case "type_declaration":
		p.processTypeDeclaration(content, lines, node, annotations, errors)
	}

	// Process children recursively
	for _, child := range node.Children() {
		p.processNode(content, lines, child, annotations, errors)
	}
}

// processVariableDeclaration extracts type annotations from variable declarations
func (p *TreeSitterParser) processVariableDeclaration(content string, lines []string, node Node, annotations *[]*TypeAnnotation, errors *[]error) {
	// Get the line containing the variable declaration
	lineNum := node.Start().Line
	if lineNum <= 0 || lineNum > len(lines) {
		return
	}
	line := lines[lineNum-1]

	// Parse the variable declaration to find type annotations
	// Example: my Type $var
	parts := strings.Fields(line)
	if len(parts) >= 3 && (parts[0] == "my" || parts[0] == "our" || parts[0] == "state") {
		varName := ""
		typeName := ""

		for j, part := range parts[1:] {
			if strings.HasPrefix(part, "$") || strings.HasPrefix(part, "@") || strings.HasPrefix(part, "%") {
				varName = part
				if j > 0 {
					typeName = parts[j]

					// Handle parameterized types that might be split across multiple tokens
					if strings.Contains(typeName, "[") && !strings.Contains(typeName, "]") {
						for k := j + 1; k < len(parts); k++ {
							if strings.Contains(parts[k], "]") {
								typeName = strings.Join(parts[j:k+1], " ")
								break
							}
						}
					}

					// Parse the type expression
					pos := Position{
						Line:   lineNum,
						Column: strings.Index(line, typeName) + 1,
					}
					typeExpr, err := ParseTypeExpression(typeName, pos)
					if err != nil {
						*errors = append(*errors, err)
						continue
					}

					// Create a type annotation
					*annotations = append(*annotations, &TypeAnnotation{
						AnnotatedItem:  varName,
						TypeExpression: typeExpr,
						Pos:            pos,
						Kind:           VarAnnotation,
					})
				}
				break
			}
		}
	}
}

// processSubroutineDeclaration extracts type annotations from subroutine declarations
func (p *TreeSitterParser) processSubroutineDeclaration(content string, lines []string, node Node, annotations *[]*TypeAnnotation, errors *[]error) {
	// Get the line containing the subroutine declaration
	lineNum := node.Start().Line
	if lineNum <= 0 || lineNum > len(lines) {
		return
	}
	line := lines[lineNum-1]

	// Determine if this is a method or subroutine
	isMethod := strings.HasPrefix(strings.TrimSpace(line), "method ")

	// Parse return type annotations
	// Example: sub foo() -> ReturnType
	if strings.Contains(line, "->") {
		parts := strings.Split(line, "->")
		if len(parts) == 2 {
			returnType := strings.TrimSpace(parts[1])

			// Remove trailing block or semicolon if present
			if strings.Contains(returnType, "{") {
				returnType = strings.TrimSpace(returnType[:strings.Index(returnType, "{")])
			}
			if strings.Contains(returnType, ";") {
				returnType = strings.TrimSpace(returnType[:strings.Index(returnType, ";")])
			}

			// Parse the type expression
			pos := Position{
				Line:   lineNum,
				Column: strings.Index(line, "->") + 3,
			}
			typeExpr, err := ParseTypeExpression(returnType, pos)
			if err != nil {
				*errors = append(*errors, err)
			} else {
				// Create a return type annotation
				var kind AnnotationKind
				if isMethod {
					kind = MethodReturnAnnotation
				} else {
					kind = SubReturnAnnotation
				}

				*annotations = append(*annotations, &TypeAnnotation{
					AnnotatedItem:  "return",
					TypeExpression: typeExpr,
					Pos:            pos,
					Kind:           kind,
				})
			}
		}
	}

	// Parse parameter type annotations
	// Example: sub foo(Type $param, AnotherType @array)
	if strings.Contains(line, "(") && strings.Contains(line, ")") {
		paramStart := strings.Index(line, "(")
		paramEnd := strings.Index(line, ")")
		if paramStart >= 0 && paramEnd > paramStart {
			params := line[paramStart+1 : paramEnd]

			// Split parameters by comma
			paramList := strings.Split(params, ",")
			for _, param := range paramList {
				param = strings.TrimSpace(param)
				if param == "" {
					continue
				}

				// Parse parameter for type annotations
				parts := strings.Fields(param)
				if len(parts) >= 2 && (strings.HasPrefix(parts[len(parts)-1], "$") ||
					strings.HasPrefix(parts[len(parts)-1], "@") ||
					strings.HasPrefix(parts[len(parts)-1], "%")) {

					paramName := parts[len(parts)-1]
					paramType := strings.Join(parts[:len(parts)-1], " ")

					// Parse the type expression
					paramPos := strings.Index(params, param) + paramStart + 1
					typePos := strings.Index(param, parts[0]) + paramPos

					pos := Position{
						Line:   lineNum,
						Column: typePos + 1,
					}
					typeExpr, err := ParseTypeExpression(paramType, pos)
					if err != nil {
						*errors = append(*errors, err)
						continue
					}

					// Create a parameter type annotation
					var kind AnnotationKind
					if isMethod {
						kind = MethodParamAnnotation
					} else {
						kind = SubParamAnnotation
					}

					*annotations = append(*annotations, &TypeAnnotation{
						AnnotatedItem:  paramName,
						TypeExpression: typeExpr,
						Pos:            pos,
						Kind:           kind,
					})
				}
			}
		}
	}
}

// processFieldDeclaration extracts type annotations from field declarations
func (p *TreeSitterParser) processFieldDeclaration(content string, lines []string, node Node, annotations *[]*TypeAnnotation, errors *[]error) {
	// Get the line containing the field declaration
	lineNum := node.Start().Line
	if lineNum <= 0 || lineNum > len(lines) {
		return
	}
	line := lines[lineNum-1]

	// Parse field declarations with type annotations
	// Example: field Type $attr
	parts := strings.Fields(line)
	if len(parts) >= 3 && parts[0] == "field" {
		attrName := ""
		typeName := ""

		for j, part := range parts[1:] {
			if strings.HasPrefix(part, "$") {
				attrName = part
				if j > 0 {
					typeName = parts[j]

					// Handle parameterized types that might be split across multiple tokens
					if strings.Contains(typeName, "[") && !strings.Contains(typeName, "]") {
						for k := j + 1; k < len(parts); k++ {
							if strings.Contains(parts[k], "]") {
								typeName = strings.Join(parts[j:k+1], " ")
								break
							}
						}
					}

					// Parse the type expression
					pos := Position{
						Line:   lineNum,
						Column: strings.Index(line, typeName) + 1,
					}
					typeExpr, err := ParseTypeExpression(typeName, pos)
					if err != nil {
						*errors = append(*errors, err)
						continue
					}

					// Create a type annotation
					*annotations = append(*annotations, &TypeAnnotation{
						AnnotatedItem:  attrName,
						TypeExpression: typeExpr,
						Pos:            pos,
						Kind:           AttrAnnotation,
					})
				}
				break
			}
		}
	}
}

// processTypeDeclaration extracts type declarations
func (p *TreeSitterParser) processTypeDeclaration(content string, lines []string, node Node, annotations *[]*TypeAnnotation, errors *[]error) {
	// Get the line containing the type declaration
	lineNum := node.Start().Line
	if lineNum <= 0 || lineNum > len(lines) {
		return
	}
	line := lines[lineNum-1]

	// Parse type declarations
	// Example: type MyType = Type
	if strings.HasPrefix(strings.TrimSpace(line), "type ") {
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			// Extract the type name
			typeParts := strings.Fields(parts[0])
			if len(typeParts) >= 2 && typeParts[0] == "type" {
				typeName := typeParts[1]

				// Extract the type definition
				typeDef := strings.TrimSpace(parts[1])
				if strings.Contains(typeDef, ";") {
					typeDef = strings.TrimSpace(typeDef[:strings.Index(typeDef, ";")])
				}

				// Parse the type expression
				pos := Position{
					Line:   lineNum,
					Column: strings.Index(line, "=") + 2,
				}
				typeExpr, err := ParseTypeExpression(typeDef, pos)
				if err != nil {
					*errors = append(*errors, err)
					return
				}

				// Create a type declaration annotation
				*annotations = append(*annotations, &TypeAnnotation{
					AnnotatedItem:  typeName,
					TypeExpression: typeExpr,
					Pos:            pos,
					Kind:           TypeDeclAnnotation,
				})
			}
		}
	}
}

/*
 * Tree-sitter Grammar Extensions for Perl Type Annotations
 *
 * The following are the grammar extensions that would be needed for tree-sitter:
 *
 * 1. Variable Declarations:
 *    - Scalar variables: `my Type $name`
 *    - Array variables: `my Type @array`
 *    - Hash variables: `my Type %hash`
 *    - With assignments: `my Type $var = value`
 *
 * 2. Subroutine Declarations:
 *    - Parameter types: `sub name(Type $param, AnotherType @array)`
 *    - Return types: `sub name() -> ReturnType`
 *    - Combined: `sub name(Type $param) -> ReturnType`
 *
 * 3. Method Declarations:
 *    - In regular packages: `sub method(Type $self, Type $param) -> ReturnType`
 *    - In class syntax: `method name(Type $param) -> ReturnType`
 *
 * 4. Attribute Declarations:
 *    - In class syntax: `field Type $attribute`
 *    - With default values: `field Type $attribute = default_value`
 *
 * 5. Type Expressions:
 *    - Simple types: `Int`, `Str`, `Bool`
 *    - Parameterized types: `ArrayRef[Type]`, `HashRef[KeyType, ValueType]`
 *    - Union types: `Type1|Type2`
 *    - Intersection types: `Type1&Type2`
 *    - Negation types: `!Type`
 *
 * 6. Package-level Type Declarations:
 *    - Type aliases: `type TypeName = Type`
 *    - Class declarations: `class ClassName { ... }`
 *
 * 7. Typecast Expressions:
 *    - Type assertion: `$var as Type`
 */

// createTypeAnnotationQueries creates queries for extracting type annotations from a tree-sitter syntax tree
// This function would be used in a real implementation with tree-sitter
func (p *TreeSitterParser) createTypeAnnotationQueries() string {
	// Example of what tree-sitter queries might look like
	// These would be used with the tree-sitter query API
	return `
	;; Variable declarations with type annotations
	(variable_declaration
		declarator: (variable_declarator)
		type: (identifier) @type
		name: (identifier) @name) @variable_declaration

	;; Subroutine parameter type annotations
	(subroutine_declaration
		name: (identifier) @sub_name
		parameters: (parameter_list
			(parameter
				type: (identifier) @param_type
				name: (identifier) @param_name))) @sub_declaration

	;; Subroutine return type annotations
	(subroutine_declaration
		name: (identifier) @sub_name
		return_type: (return_type_annotation
			type: (identifier) @return_type)) @sub_declaration

	;; Method parameter type annotations
	(method_declaration
		name: (identifier) @method_name
		parameters: (parameter_list
			(parameter
				type: (identifier) @param_type
				name: (identifier) @param_name))) @method_declaration

	;; Method return type annotations
	(method_declaration
		name: (identifier) @method_name
		return_type: (return_type_annotation
			type: (identifier) @return_type)) @method_declaration

	;; Field declarations with type annotations
	(field_declaration
		type: (identifier) @field_type
		name: (identifier) @field_name) @field_declaration

	;; Type declarations
	(type_declaration
		name: (identifier) @type_name
		definition: (type_definition) @type_definition) @type_declaration
	`
}

/*
 * Tree-sitter Grammar Extensions for Perl Type Annotations
 *
 * The following are the grammar extensions that would be needed for tree-sitter:
 *
 * 1. Variable Declarations:
 *    - Scalar variables: `my Type $name`
 *    - Array variables: `my Type @array`
 *    - Hash variables: `my Type %hash`
 *    - With assignments: `my Type $var = value`
 *
 * 2. Subroutine Declarations:
 *    - Parameter types: `sub name(Type $param, AnotherType @array)`
 *    - Return types: `sub name() -> ReturnType`
 *    - Combined: `sub name(Type $param) -> ReturnType`
 *
 * 3. Method Declarations:
 *    - In regular packages: `sub method(Type $self, Type $param) -> ReturnType`
 *    - In class syntax: `method name(Type $param) -> ReturnType`
 *
 * 4. Attribute Declarations:
 *    - In class syntax: `field Type $attribute`
 *    - With default values: `field Type $attribute = default_value`
 *
 * 5. Type Expressions:
 *    - Simple types: `Int`, `Str`, `Bool`
 *    - Parameterized types: `ArrayRef[Type]`, `HashRef[KeyType, ValueType]`
 *    - Union types: `Type1|Type2`
 *    - Intersection types: `Type1&Type2`
 *    - Negation types: `!Type`
 *
 * 6. Package-level Type Declarations:
 *    - Type aliases: `type TypeName = Type`
 *    - Class declarations: `class ClassName { ... }`
 *
 * 7. Typecast Expressions:
 *    - Type assertion: `$var as Type`
 */

// createTypeAnnotationQueries creates tree-sitter queries for extracting type annotations
// This is a placeholder for future tree-sitter implementation
// Currently unused as we're using a simpler parser implementation
//
// TODO: Remove this comment and fully implement tree-sitter integration
// in the future when we move to a more sophisticated parser
