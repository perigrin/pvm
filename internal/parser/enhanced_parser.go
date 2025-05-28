// ABOUTME: Enhanced parser with improved error recovery and position tracking
// ABOUTME: Integrates type error recovery with tree-sitter parsing

package parser

import (
	"fmt"
	"io"
	"os"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/log"
)

// EnhancedParser provides improved error handling and position tracking
type EnhancedParser struct {
	baseParser    Parser
	errorRecovery *TypeErrorRecovery
	debug         bool
}

// NewEnhancedParser creates a new parser with enhanced error recovery
func NewEnhancedParser(debug bool) (*EnhancedParser, error) {
	baseParser, err := NewTreeSitterParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create base parser: %v", err)
	}

	return &EnhancedParser{
		baseParser:    baseParser,
		errorRecovery: NewTypeErrorRecovery(),
		debug:         debug,
	}, nil
}

// ParseFile parses a file with enhanced error recovery
func (ep *EnhancedParser) ParseFile(path string) (*ast.AST, error) {
	// Read file content for error recovery context
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", path, err)
	}

	return ep.parseWithErrorRecovery(string(content), path)
}

// ParseString parses a string with enhanced error recovery
func (ep *EnhancedParser) ParseString(content string) (*ast.AST, error) {
	return ep.parseWithErrorRecovery(content, "<string>")
}

// ParseReader parses content from a reader with enhanced error recovery
func (ep *EnhancedParser) ParseReader(reader io.Reader) (*ast.AST, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %v", err)
	}

	return ep.parseWithErrorRecovery(string(content), "<reader>")
}

// parseWithErrorRecovery performs parsing with comprehensive error recovery
func (ep *EnhancedParser) parseWithErrorRecovery(content, source string) (*ast.AST, error) {
	// First attempt: normal parsing
	result, err := ep.baseParser.ParseString(content)
	if err == nil {
		// Success - enhance with position validation
		return ep.enhanceAST(result, content, source), nil
	}

	if ep.debug {
		log.Debugf("Initial parse failed: %v, attempting error recovery", err)
	}

	// Parse failed - attempt error recovery
	recoveredAST, recoveryErrors := ep.attemptErrorRecovery(content, source, err)
	if recoveredAST != nil {
		// Add recovery errors as warnings
		for _, recErr := range recoveryErrors {
			recoveredAST.Errors = append(recoveredAST.Errors, recErr)
		}
		return recoveredAST, nil
	}

	// Recovery failed - return original error with enhanced information
	return nil, ep.enhanceParseError(err, content, source)
}

// attemptErrorRecovery tries to recover from parsing errors
func (ep *EnhancedParser) attemptErrorRecovery(content, source string, originalErr error) (*ast.AST, []*TypeError) {
	var recoveryErrors []*TypeError
	
	// Try to identify and fix common type expression errors
	fixedContent, fixes := ep.fixCommonTypeErrors(content)
	recoveryErrors = append(recoveryErrors, fixes...)

	if fixedContent != content {
		if ep.debug {
			log.Debugf("Applied %d fixes, retrying parse", len(fixes))
		}

		// Try parsing the fixed content
		result, err := ep.baseParser.ParseString(fixedContent)
		if err == nil {
			// Success with fixes
			enhanced := ep.enhanceAST(result, fixedContent, source)
			return enhanced, recoveryErrors
		}

		if ep.debug {
			log.Debugf("Parse still failed after fixes: %v", err)
		}
	}

	// Attempt partial parsing by breaking content into segments
	partialAST, partialErrors := ep.attemptPartialParsing(content, source)
	recoveryErrors = append(recoveryErrors, partialErrors...)

	return partialAST, recoveryErrors
}

// fixCommonTypeErrors identifies and fixes common type expression errors
func (ep *EnhancedParser) fixCommonTypeErrors(content string) (string, []*TypeError) {
	var fixes []*TypeError
	lines := strings.Split(content, "\n")
	
	for lineNum, line := range lines {
		position := ast.Position{Line: lineNum + 1, Column: 1}
		
		// Fix missing closing brackets
		if strings.Contains(line, "ArrayRef[") && !strings.Contains(line, "]") {
			if strings.Contains(line, ";") {
				// Insert ] before semicolon
				fixed := strings.Replace(line, ";", "];", 1)
				lines[lineNum] = fixed
				
				fixes = append(fixes, &TypeError{
					Message:    "Added missing closing bracket",
					Position:   position,
					Suggestion: "Check bracket matching in type expressions",
					Context:    "auto-fix",
					ErrorCode:  MissingClosingBracketError,
				})
			}
		}
		
		// Fix double union operators
		if strings.Contains(line, "||") && containsTypeLikePattern(line) {
			fixed := strings.ReplaceAll(line, "||", "|")
			lines[lineNum] = fixed
			
			fixes = append(fixes, &TypeError{
				Message:    "Fixed double union operator",
				Position:   position,
				Suggestion: "Use single '|' for union types",
				Context:    "auto-fix",
				ErrorCode:  InvalidUnionSyntaxError,
			})
		}
		
		// Fix double intersection operators
		if strings.Contains(line, "&&") && containsTypeLikePattern(line) {
			fixed := strings.ReplaceAll(line, "&&", "&")
			lines[lineNum] = fixed
			
			fixes = append(fixes, &TypeError{
				Message:    "Fixed double intersection operator",
				Position:   position,
				Suggestion: "Use single '&' for intersection types",
				Context:    "auto-fix",
				ErrorCode:  InvalidIntersectionSyntaxError,
			})
		}
		
		// Fix double negation operators
		if strings.Contains(line, "!!") && containsTypeLikePattern(line) {
			fixed := strings.ReplaceAll(line, "!!", "!")
			lines[lineNum] = fixed
			
			fixes = append(fixes, &TypeError{
				Message:    "Fixed double negation operator",
				Position:   position,
				Suggestion: "Use single '!' for negation types",
				Context:    "auto-fix",
				ErrorCode:  InvalidNegationSyntaxError,
			})
		}
	}
	
	return strings.Join(lines, "\n"), fixes
}

// attemptPartialParsing tries to parse content in segments
func (ep *EnhancedParser) attemptPartialParsing(content, source string) (*ast.AST, []*TypeError) {
	var errors []*TypeError
	
	// Split content by statements/blocks and try to parse each
	segments := ep.splitIntoSegments(content)
	
	var validNodes []ast.Node
	var typeAnnotations []*ast.TypeAnnotation
	
	for i, segment := range segments {
		segmentAST, err := ep.baseParser.ParseString(segment)
		if err == nil && segmentAST != nil {
			// Successfully parsed segment
			if segmentAST.Root != nil {
				validNodes = append(validNodes, segmentAST.Root)
			}
			typeAnnotations = append(typeAnnotations, segmentAST.TypeAnnotations...)
		} else {
			// Failed to parse segment - record error
			position := ep.calculateSegmentPosition(content, segments, i)
			typeErr := ep.errorRecovery.RecoverFromTypeError(segment, position, "partial parsing")
			errors = append(errors, typeErr)
		}
	}
	
	// Construct a partial AST from valid segments
	if len(validNodes) > 0 {
		// Create a root node containing all valid parsed segments
		rootNode := ast.NewBaseNode("partial_root", ast.Position{Line: 1, Column: 1}, ast.Position{Line: 1, Column: 1})
		
		// Add all valid nodes as children
		for _, node := range validNodes {
			rootNode.AddChild(node)
		}
		
		return &ast.AST{
			Path:            source,
			Root:            rootNode,
			TypeAnnotations: typeAnnotations,
			Errors:          make([]error, 0), // Will be added by caller
		}, errors
	}
	
	return nil, errors
}

// splitIntoSegments splits content into parseable segments
func (ep *EnhancedParser) splitIntoSegments(content string) []string {
	var segments []string
	lines := strings.Split(content, "\n")
	
	currentSegment := ""
	braceLevel := 0
	
	for _, line := range lines {
		currentSegment += line + "\n"
		
		// Track brace nesting
		for _, char := range line {
			switch char {
			case '{':
				braceLevel++
			case '}':
				braceLevel--
				if braceLevel == 0 && strings.TrimSpace(currentSegment) != "" {
					// End of balanced block
					segments = append(segments, strings.TrimSpace(currentSegment))
					currentSegment = ""
				}
			}
		}
		
		// Also split on statement boundaries when not in braces
		if braceLevel == 0 && strings.HasSuffix(strings.TrimSpace(line), ";") {
			if strings.TrimSpace(currentSegment) != "" {
				segments = append(segments, strings.TrimSpace(currentSegment))
				currentSegment = ""
			}
		}
	}
	
	// Add remaining content as final segment
	if strings.TrimSpace(currentSegment) != "" {
		segments = append(segments, strings.TrimSpace(currentSegment))
	}
	
	return segments
}

// calculateSegmentPosition calculates the position of a segment in the original content
func (ep *EnhancedParser) calculateSegmentPosition(content string, segments []string, segmentIndex int) ast.Position {
	// Find where this segment appears in the original content
	offset := 0
	for i := 0; i < segmentIndex; i++ {
		offset += len(segments[i]) + 1 // +1 for separator
	}
	
	// Convert offset to line/column
	lines := strings.Split(content[:offset], "\n")
	line := len(lines)
	column := 1
	if len(lines) > 0 {
		column = len(lines[len(lines)-1]) + 1
	}
	
	return ast.Position{
		Line:   line,
		Column: column,
		Offset: offset,
	}
}

// enhanceAST adds improved position tracking and validation to parsed AST
func (ep *EnhancedParser) enhanceAST(ast *ast.AST, content, source string) *ast.AST {
	if ast == nil {
		return ast
	}
	
	// Validate type expressions in the AST
	var validationErrors []error
	for _, ta := range ast.TypeAnnotations {
		if ta.TypeExpression != nil {
			typeErrors := ep.errorRecovery.ValidateTypeExpression(ta.TypeExpression, content)
			for _, te := range typeErrors {
				validationErrors = append(validationErrors, te)
			}
		}
	}
	
	// Add validation errors to AST
	ast.Errors = append(ast.Errors, validationErrors...)
	
	// Enhance position information
	ep.enhancePositionTracking(ast, content)
	
	return ast
}

// enhancePositionTracking improves position accuracy for AST nodes
func (ep *EnhancedParser) enhancePositionTracking(ast *ast.AST, content string) {
	if ast == nil || ast.Root == nil {
		return
	}
	
	// Traverse AST and validate/enhance position information
	ep.visitNodeForPositionEnhancement(ast.Root, content)
	
	// Enhance type annotation positions
	for _, ta := range ast.TypeAnnotations {
		if ta.TypeExpression != nil {
			ep.enhanceTypeExpressionPosition(ta.TypeExpression, content)
		}
	}
}

// visitNodeForPositionEnhancement recursively enhances node positions
func (ep *EnhancedParser) visitNodeForPositionEnhancement(node ast.Node, content string) {
	if node == nil {
		return
	}
	
	// Validate position bounds
	pos := node.Start()
	lines := strings.Split(content, "\n")
	
	if pos.Line > len(lines) || pos.Line < 1 {
		if ep.debug {
			log.Debugf("Invalid line number %d for node %s", pos.Line, node.Type())
		}
	}
	
	if pos.Line <= len(lines) {
		line := lines[pos.Line-1]
		if pos.Column > len(line) || pos.Column < 1 {
			if ep.debug {
				log.Debugf("Invalid column %d for node %s on line %d", pos.Column, node.Type(), pos.Line)
			}
		}
	}
	
	// Recursively enhance child nodes
	for _, child := range node.Children() {
		ep.visitNodeForPositionEnhancement(child, content)
	}
}

// enhanceTypeExpressionPosition improves position tracking for type expressions
func (ep *EnhancedParser) enhanceTypeExpressionPosition(expr *ast.TypeExpression, content string) {
	if expr == nil {
		return
	}
	
	// Validate position
	pos := expr.Start()
	lines := strings.Split(content, "\n")
	
	if pos.Line > 0 && pos.Line <= len(lines) {
		line := lines[pos.Line-1]
		if pos.Column > 0 && pos.Column <= len(line) {
			// Position looks valid
			if ep.debug && expr.Name != "" {
				expectedText := line[pos.Column-1:]
				if !strings.HasPrefix(expectedText, expr.Name) {
					log.Debugf("Position mismatch for type %s at %d:%d", expr.Name, pos.Line, pos.Column)
				}
			}
		}
	}
	
	// Recursively enhance parameter positions
	for _, param := range expr.Parameters {
		ep.enhanceTypeExpressionPosition(param, content)
	}
	
	// Enhance union type positions
	for _, unionType := range expr.UnionTypes {
		ep.enhanceTypeExpressionPosition(unionType, content)
	}
	
	// Enhance intersection type positions
	for _, intersectionType := range expr.IntersectionTypes {
		ep.enhanceTypeExpressionPosition(intersectionType, content)
	}
}

// enhanceParseError creates an enhanced parse error with better context
func (ep *EnhancedParser) enhanceParseError(originalErr error, content, source string) error {
	// Try to extract position information from the original error
	errorStr := originalErr.Error()
	
	// Look for common error patterns and enhance them
	if strings.Contains(errorStr, "syntax error") {
		// Find approximate position of syntax error
		position := ep.findErrorPosition(content, errorStr)
		
		typeErr := ep.errorRecovery.RecoverFromTypeError(content, position, "parse error")
		return fmt.Errorf("enhanced parse error: %v (original: %v)", typeErr, originalErr)
	}
	
	return originalErr
}

// findErrorPosition attempts to find the position of a parsing error
func (ep *EnhancedParser) findErrorPosition(content, errorStr string) ast.Position {
	// Default position
	position := ast.Position{Line: 1, Column: 1}
	
	// Try to extract line/column from error message
	// This is a simple heuristic - real implementation would be more sophisticated
	lines := strings.Split(content, "\n")
	
	// Look for lines that might contain syntax errors
	for lineNum, line := range lines {
		if strings.Contains(line, "ArrayRef[") && !strings.Contains(line, "]") {
			position.Line = lineNum + 1
			position.Column = strings.Index(line, "ArrayRef[") + 1
			break
		}
		if strings.Contains(line, "||") && containsTypeLikePattern(line) {
			position.Line = lineNum + 1
			position.Column = strings.Index(line, "||") + 1
			break
		}
	}
	
	return position
}