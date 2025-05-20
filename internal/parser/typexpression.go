// ABOUTME: Type expression parsing for PSC
// ABOUTME: Handles parsing of type annotations and expressions

package parser

// parseTypeExpression parses a type expression string and returns a TypeExpression
func parseTypeExpression(text string, pos Position) (*TypeExpression, error) {
	parser := &typeExpressionParser{
		text: text,
		pos:  0,
		line: pos.Line,
		col:  pos.Column,
	}

	return parser.parse()
}

// typeExpressionParser is a simple parser for type expressions
type typeExpressionParser struct {
	text string
	pos  int
	line int
	col  int
}

// parse parses a type expression
func (p *typeExpressionParser) parse() (*TypeExpression, error) {
	// Skip whitespace
	p.skipWhitespace()

	// Check for negation
	negation := false
	if p.peek() == '!' {
		negation = true
		p.consume()
		p.skipWhitespace()
	}

	// Parse the base type
	baseType, err := p.parseBaseType()
	if err != nil {
		return nil, err
	}

	// Set negation flag
	baseType.Negation = negation

	// Check for union or intersection
	p.skipWhitespace()
	if p.pos < len(p.text) {
		if p.peek() == '|' {
			// Union type (Type1|Type2)
			p.consume()
			p.skipWhitespace()

			// Parse the right side
			rightType, err := p.parse()
			if err != nil {
				return nil, err
			}

			// Create a union type
			return &TypeExpression{
				Name:  baseType.Name + "|" + rightType.Name,
				Union: true,
				Params: []*TypeExpression{
					baseType,
					rightType,
				},
				Pos: Position{
					Line:   p.line,
					Column: p.col,
				},
			}, nil
		} else if p.peek() == '&' {
			// Intersection type (Type1&Type2)
			p.consume()
			p.skipWhitespace()

			// Parse the right side
			rightType, err := p.parse()
			if err != nil {
				return nil, err
			}

			// Create an intersection type
			return &TypeExpression{
				Name:         baseType.Name + "&" + rightType.Name,
				Intersection: true,
				Params: []*TypeExpression{
					baseType,
					rightType,
				},
				Pos: Position{
					Line:   p.line,
					Column: p.col,
				},
			}, nil
		}
	}

	return baseType, nil
}

// parseBaseType parses a base type, which can be a simple type or a parameterized type
func (p *typeExpressionParser) parseBaseType() (*TypeExpression, error) {
	// Parse the type name
	typeName := p.parseTypeName()
	if typeName == "" {
		return nil, &ParseError{
			Message: "Expected type name",
			Line:    p.line,
			Column:  p.col + p.pos,
		}
	}

	// Check for parameterized type
	p.skipWhitespace()
	if p.pos < len(p.text) && p.peek() == '[' {
		// Parameterized type (Type[Param1, Param2, ...])
		p.consume() // Skip '['

		// Parse parameters
		var params []*TypeExpression

		for {
			p.skipWhitespace()

			// Check for end of parameters
			if p.pos < len(p.text) && p.peek() == ']' {
				p.consume() // Skip ']'
				break
			}

			// Parse a parameter
			param, err := p.parse()
			if err != nil {
				return nil, err
			}

			params = append(params, param)

			// Skip whitespace
			p.skipWhitespace()

			// Check for comma or end of parameters
			if p.pos < len(p.text) {
				if p.peek() == ',' {
					p.consume() // Skip ','
				} else if p.peek() == ']' {
					p.consume() // Skip ']'
					break
				} else {
					return nil, &ParseError{
						Message: "Expected ',' or ']'",
						Line:    p.line,
						Column:  p.col + p.pos,
					}
				}
			} else {
				return nil, &ParseError{
					Message: "Unterminated parameter list",
					Line:    p.line,
					Column:  p.col + p.pos,
				}
			}
		}

		// Create a parameterized type
		return &TypeExpression{
			Name:   typeName,
			Params: params,
			Pos: Position{
				Line:   p.line,
				Column: p.col,
			},
		}, nil
	}

	// Simple type
	return &TypeExpression{
		Name: typeName,
		Pos: Position{
			Line:   p.line,
			Column: p.col,
		},
	}, nil
}

// parseTypeName parses a type name
func (p *typeExpressionParser) parseTypeName() string {
	start := p.pos

	// Skip initial whitespace
	p.skipWhitespace()
	start = p.pos

	// Parse the type name
	for p.pos < len(p.text) {
		ch := p.peek()

		// Type names can contain alphanumeric characters, colons (for namespaces), and underscores
		if isAlphaNumeric(ch) || ch == ':' || ch == '_' {
			p.consume()
		} else {
			break
		}
	}

	// Return the type name
	if p.pos > start {
		return p.text[start:p.pos]
	}

	return ""
}

// skipWhitespace skips whitespace characters
func (p *typeExpressionParser) skipWhitespace() {
	for p.pos < len(p.text) && isWhitespace(p.peek()) {
		p.consume()
	}
}

// peek returns the current character without consuming it
func (p *typeExpressionParser) peek() byte {
	if p.pos < len(p.text) {
		return p.text[p.pos]
	}
	return 0
}

// consume consumes the current character and advances the position
func (p *typeExpressionParser) consume() {
	if p.pos < len(p.text) {
		if p.text[p.pos] == '\n' {
			p.line++
			p.col = 0
		} else {
			p.col++
		}
		p.pos++
	}
}

// isWhitespace returns true if the character is a whitespace character
func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// isAlphaNumeric returns true if the character is an alphanumeric character
func isAlphaNumeric(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}
