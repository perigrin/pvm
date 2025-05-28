// ABOUTME: Compatibility adapters for migrating from old AST types to new consolidated types
// ABOUTME: Provides seamless migration path while maintaining backward compatibility

package ast

import (
	"io"
)

// ParserCompatibilityAdapter provides compatibility with the old parser.Parser interface
// This allows existing code to continue working while we migrate to consolidated AST types
type ParserCompatibilityAdapter struct {
	underlyingParser Parser
}

// Parser interface for compatibility with existing parser package
type Parser interface {
	ParseFile(path string) (*AST, error)
	ParseString(content string) (*AST, error)
	ParseReader(reader io.Reader) (*AST, error)
}

// NewParserAdapter creates a parser compatibility adapter
func NewParserAdapter(parser Parser) *ParserCompatibilityAdapter {
	return &ParserCompatibilityAdapter{
		underlyingParser: parser,
	}
}

// ParseFile implements the parser interface
func (pca *ParserCompatibilityAdapter) ParseFile(path string) (*AST, error) {
	return pca.underlyingParser.ParseFile(path)
}

// ParseString implements the parser interface
func (pca *ParserCompatibilityAdapter) ParseString(content string) (*AST, error) {
	return pca.underlyingParser.ParseString(content)
}

// ParseReader implements the parser interface
func (pca *ParserCompatibilityAdapter) ParseReader(reader io.Reader) (*AST, error) {
	return pca.underlyingParser.ParseReader(reader)
}

// LegacyNodeAdapter wraps old Node implementations to work with new interface
type LegacyNodeAdapter struct {
	node   interface{} // Old node type
	parent Node
}

// NewLegacyNodeAdapter creates an adapter for old node types
func NewLegacyNodeAdapter(oldNode interface{}) *LegacyNodeAdapter {
	return &LegacyNodeAdapter{
		node: oldNode,
	}
}

// Type implements Node interface by trying to call Type() on the wrapped node
func (lna *LegacyNodeAdapter) Type() string {
	if typed, ok := lna.node.(interface{ Type() string }); ok {
		return typed.Type()
	}
	return "unknown"
}

// Start implements Node interface
func (lna *LegacyNodeAdapter) Start() Position {
	if hasStart, ok := lna.node.(interface{ Start() interface{} }); ok {
		if start := hasStart.Start(); start != nil {
			// Try to convert old Position to new Position
			if oldPos, ok := start.(struct {
				Line   int
				Column int
				Offset int
			}); ok {
				return Position{
					Line:   oldPos.Line,
					Column: oldPos.Column,
					Offset: oldPos.Offset,
				}
			}
		}
	}
	return Position{}
}

// End implements Node interface
func (lna *LegacyNodeAdapter) End() Position {
	if hasEnd, ok := lna.node.(interface{ End() interface{} }); ok {
		if end := hasEnd.End(); end != nil {
			// Try to convert old Position to new Position
			if oldPos, ok := end.(struct {
				Line   int
				Column int
				Offset int
			}); ok {
				return Position{
					Line:   oldPos.Line,
					Column: oldPos.Column,
					Offset: oldPos.Offset,
				}
			}
		}
	}
	return Position{}
}

// Children implements Node interface
func (lna *LegacyNodeAdapter) Children() []Node {
	if hasChildren, ok := lna.node.(interface{ Children() []interface{} }); ok {
		oldChildren := hasChildren.Children()
		children := make([]Node, len(oldChildren))
		for i, child := range oldChildren {
			children[i] = NewLegacyNodeAdapter(child)
		}
		return children
	}
	return []Node{}
}

// Text implements Node interface
func (lna *LegacyNodeAdapter) Text() string {
	if hasText, ok := lna.node.(interface{ Text() string }); ok {
		return hasText.Text()
	}
	return ""
}

// Parent implements Node interface
func (lna *LegacyNodeAdapter) Parent() Node {
	return lna.parent
}

// SetParent implements Node interface
func (lna *LegacyNodeAdapter) SetParent(parent Node) {
	lna.parent = parent
}

// ConvertLegacyAST converts old AST types to new consolidated AST types
func ConvertLegacyAST(oldAST interface{}) *AST {
	// Extract common fields from old AST
	var path string
	var root Node
	var annotations []*TypeAnnotation
	var errors []error
	var source string

	// Use reflection-like approach to extract fields
	if astWithPath, ok := oldAST.(interface{ GetPath() string }); ok {
		path = astWithPath.GetPath()
	}

	if astWithRoot, ok := oldAST.(interface{ GetRoot() interface{} }); ok {
		if rootNode := astWithRoot.GetRoot(); rootNode != nil {
			root = NewLegacyNodeAdapter(rootNode)
		}
	}

	if astWithAnnotations, ok := oldAST.(interface{ GetTypeAnnotations() []interface{} }); ok {
		oldAnnotations := astWithAnnotations.GetTypeAnnotations()
		annotations = make([]*TypeAnnotation, len(oldAnnotations))
		for i, oldAnnot := range oldAnnotations {
			annotations[i] = ConvertLegacyTypeAnnotation(oldAnnot)
		}
	}

	if astWithErrors, ok := oldAST.(interface{ GetErrors() []error }); ok {
		errors = astWithErrors.GetErrors()
	}

	if astWithSource, ok := oldAST.(interface{ GetSource() string }); ok {
		source = astWithSource.GetSource()
	}

	return &AST{
		Path:            path,
		Root:            root,
		TypeAnnotations: annotations,
		Errors:          errors,
		Source:          source,
	}
}

// ConvertLegacyTypeAnnotation converts old TypeAnnotation to new format
func ConvertLegacyTypeAnnotation(oldAnnotation interface{}) *TypeAnnotation {
	annotation := &TypeAnnotation{}

	if hasItem, ok := oldAnnotation.(interface{ GetAnnotatedItem() string }); ok {
		annotation.AnnotatedItem = hasItem.GetAnnotatedItem()
	}

	if hasTypeExpr, ok := oldAnnotation.(interface{ GetTypeExpression() interface{} }); ok {
		if oldTypeExpr := hasTypeExpr.GetTypeExpression(); oldTypeExpr != nil {
			annotation.TypeExpression = ConvertLegacyTypeExpression(oldTypeExpr)
		}
	}

	if hasPos, ok := oldAnnotation.(interface{ GetPos() interface{} }); ok {
		if oldPos := hasPos.GetPos(); oldPos != nil {
			annotation.Pos = ConvertLegacyPosition(oldPos)
		}
	}

	if hasKind, ok := oldAnnotation.(interface{ GetKind() int }); ok {
		annotation.Kind = AnnotationKind(hasKind.GetKind())
	}

	return annotation
}

// ConvertLegacyTypeExpression converts old TypeExpression to new format
func ConvertLegacyTypeExpression(oldTypeExpr interface{}) *TypeExpression {
	// Initialize with default position, will be updated if position info is available
	expr := &TypeExpression{
		BaseNode: NewBaseNode("type_expr", Position{}, Position{}),
	}

	// Extract base type/name
	if hasName, ok := oldTypeExpr.(interface{ GetName() string }); ok {
		expr.Name = hasName.GetName()
	}
	if hasBaseType, ok := oldTypeExpr.(interface{ GetBaseType() string }); ok {
		expr.Name = hasBaseType.GetBaseType()
	}

	// Extract parameters
	if hasParams, ok := oldTypeExpr.(interface{ GetParams() []interface{} }); ok {
		oldParams := hasParams.GetParams()
		expr.Parameters = make([]*TypeExpression, len(oldParams))
		for i, param := range oldParams {
			expr.Parameters[i] = ConvertLegacyTypeExpression(param)
		}
	}

	// Extract flags
	if hasUnion, ok := oldTypeExpr.(interface{ IsUnion() bool }); ok {
		expr.IsUnion = hasUnion.IsUnion()
	}
	if hasIntersection, ok := oldTypeExpr.(interface{ IsIntersection() bool }); ok {
		expr.IsIntersection = hasIntersection.IsIntersection()
	}
	if hasNegation, ok := oldTypeExpr.(interface{ IsNegation() bool }); ok {
		expr.IsNegation = hasNegation.IsNegation()
	}

	// Extract original string
	if hasOriginal, ok := oldTypeExpr.(interface{ GetOriginalString() string }); ok {
		expr.OriginalString = hasOriginal.GetOriginalString()
	}

	// Extract position
	if hasPos, ok := oldTypeExpr.(interface{ GetPos() interface{} }); ok {
		if oldPos := hasPos.GetPos(); oldPos != nil {
			pos := ConvertLegacyPosition(oldPos)
			expr.BaseNode.start = pos
			expr.BaseNode.end = pos
		}
	}

	return expr
}

// ConvertLegacyPosition converts old Position to new format
func ConvertLegacyPosition(oldPos interface{}) Position {
	pos := Position{}

	if hasLine, ok := oldPos.(interface{ GetLine() int }); ok {
		pos.Line = hasLine.GetLine()
	}
	if hasColumn, ok := oldPos.(interface{ GetColumn() int }); ok {
		pos.Column = hasColumn.GetColumn()
	}
	if hasOffset, ok := oldPos.(interface{ GetOffset() int }); ok {
		pos.Offset = hasOffset.GetOffset()
	}

	// Handle struct-based position
	if structPos, ok := oldPos.(struct {
		Line   int
		Column int
		Offset int
	}); ok {
		pos.Line = structPos.Line
		pos.Column = structPos.Column
		pos.Offset = structPos.Offset
	}

	return pos
}

// CreateConcreteAST creates concrete AST nodes from generic interfaces
// This is used when we need to construct proper typed AST nodes from parsed content
func CreateConcreteAST(path, source string, root Node, annotations []*TypeAnnotation, errors []error) *AST {
	return &AST{
		Path:            path,
		Root:            root,
		TypeAnnotations: annotations,
		Errors:          errors,
		Source:          source,
	}
}

// WrapLegacyParser wraps an old parser to work with new AST types
func WrapLegacyParser(legacyParser interface{}) Parser {
	return &legacyParserWrapper{
		parser: legacyParser,
	}
}

type legacyParserWrapper struct {
	parser interface{}
}

func (lpw *legacyParserWrapper) ParseFile(path string) (*AST, error) {
	if hasParseFile, ok := lpw.parser.(interface {
		ParseFile(string) (interface{}, error)
	}); ok {
		result, err := hasParseFile.ParseFile(path)
		if err != nil {
			return nil, err
		}
		return ConvertLegacyAST(result), nil
	}
	return nil, nil
}

func (lpw *legacyParserWrapper) ParseString(content string) (*AST, error) {
	if hasParseString, ok := lpw.parser.(interface {
		ParseString(string) (interface{}, error)
	}); ok {
		result, err := hasParseString.ParseString(content)
		if err != nil {
			return nil, err
		}
		return ConvertLegacyAST(result), nil
	}
	return nil, nil
}

func (lpw *legacyParserWrapper) ParseReader(reader io.Reader) (*AST, error) {
	if hasParseReader, ok := lpw.parser.(interface {
		ParseReader(io.Reader) (interface{}, error)
	}); ok {
		result, err := hasParseReader.ParseReader(reader)
		if err != nil {
			return nil, err
		}
		return ConvertLegacyAST(result), nil
	}
	return nil, nil
}
