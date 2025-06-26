// ABOUTME: Type annotation generation logic for converting inferred types to Perl syntax
// ABOUTME: Handles complex annotation patterns and context-aware type formatting

package compiler

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/types"
)

// AnnotationGenerator handles generation of type annotations from AST nodes
type AnnotationGenerator struct {
	formatter TypeFormatter
	options   AnnotationOptions
}

// AnnotationOptions controls annotation generation behavior
type AnnotationOptions struct {
	// AnnotateVariables controls whether variable declarations get annotations
	AnnotateVariables bool

	// AnnotateMethods controls whether method signatures get annotations
	AnnotateMethods bool

	// AnnotateFields controls whether field declarations get annotations
	AnnotateFields bool

	// AnnotateReturns controls whether return statements get type comments
	AnnotateReturns bool

	// MinConfidence sets minimum confidence for any annotation
	MinConfidence float64

	// PreferredStyle sets the default formatting style
	PreferredStyle FormattingStyle

	// ContextAware enables context-sensitive annotation decisions
	ContextAware bool
}

// NewAnnotationGenerator creates a new annotation generator
func NewAnnotationGenerator(formatter TypeFormatter) *AnnotationGenerator {
	return &AnnotationGenerator{
		formatter: formatter,
		options: AnnotationOptions{
			AnnotateVariables: true,
			AnnotateMethods:   true,
			AnnotateFields:    true,
			AnnotateReturns:   false, // Usually too noisy
			MinConfidence:     0.7,
			PreferredStyle:    StyleInline,
			ContextAware:      true,
		},
	}
}

// NewAnnotationGeneratorWithOptions creates a generator with custom options
func NewAnnotationGeneratorWithOptions(formatter TypeFormatter, options AnnotationOptions) *AnnotationGenerator {
	return &AnnotationGenerator{
		formatter: formatter,
		options:   options,
	}
}

// GenerateVariableAnnotation generates annotation for a variable declaration
func (ag *AnnotationGenerator) GenerateVariableAnnotation(node *ast.VarDecl, typeInfo *types.TypeInfo) string {
	if !ag.options.AnnotateVariables || typeInfo.Confidence < ag.options.MinConfidence {
		// Get first variable name from VarDecl
		varName := ""
		if len(node.Variables()) > 0 {
			varName = node.Variables()[0].FullName()
		}
		return ag.generateFallbackAnnotation(varName, typeInfo)
	}

	style := ag.determineStyleForContext("variable", typeInfo.Confidence)
	// Get first variable name from VarDecl
	varName := ""
	if len(node.Variables()) > 0 {
		varName = node.Variables()[0].FullName()
	}
	return ag.formatter.FormatVariableDeclaration(varName, typeInfo.Type, typeInfo.Confidence, style)
}

// GenerateMethodAnnotation generates annotation for a method signature
func (ag *AnnotationGenerator) GenerateMethodAnnotation(node *ast.MethodDecl, signature *MethodSignatureInfo) string {
	if !ag.options.AnnotateMethods {
		return ag.generateBasicMethodSignature(node.Name, signature.Parameters)
	}

	style := ag.determineStyleForContext("method", signature.OverallConfidence)

	// Format parameters with type information
	var paramStrs []string
	for _, param := range signature.Parameters {
		if ag.formatter.ShouldIncludeAnnotation(param.Confidence, ag.options.MinConfidence) {
			typeStr := param.Type.String()
			paramStrs = append(paramStrs, fmt.Sprintf("%s %s", typeStr, param.Name))
		} else {
			paramStrs = append(paramStrs, param.Name)
		}
	}

	paramList := strings.Join(paramStrs, ", ")

	// Add return type if confident enough
	if signature.ReturnType != nil && signature.ReturnConfidence >= ag.options.MinConfidence {
		returnTypeStr := signature.ReturnType.String()
		switch style {
		case StyleVerbose:
			return fmt.Sprintf("sub %s(%s) -> %s", node.Name, paramList, returnTypeStr)
		case StyleCompact:
			return fmt.Sprintf("sub %s(%s):%s", node.Name, paramList, returnTypeStr)
		case StyleCommentOnly:
			return fmt.Sprintf("sub %s(%s) # returns %s", node.Name, paramList, returnTypeStr)
		default:
			return fmt.Sprintf("sub %s(%s) # -> %s", node.Name, paramList, returnTypeStr)
		}
	}

	return fmt.Sprintf("sub %s(%s)", node.Name, paramList)
}

// GenerateFieldAnnotation generates annotation for a field declaration
func (ag *AnnotationGenerator) GenerateFieldAnnotation(node *ast.FieldDecl, typeInfo *types.TypeInfo) string {
	if !ag.options.AnnotateFields || typeInfo.Confidence < ag.options.MinConfidence {
		return ag.generateFallbackFieldAnnotation(node.Name, typeInfo)
	}

	style := ag.determineStyleForContext("field", typeInfo.Confidence)

	// Use the formatter's field declaration method
	return ag.formatter.FormatFieldDeclaration(node.Name, typeInfo.Type, typeInfo.Confidence, style)
}

// GenerateReturnAnnotation generates annotation for a return statement
func (ag *AnnotationGenerator) GenerateReturnAnnotation(node *ast.ReturnStmt, typeInfo *types.TypeInfo) string {
	if !ag.options.AnnotateReturns || typeInfo.Confidence < ag.options.MinConfidence {
		return "" // No annotation for returns by default
	}

	style := ag.determineStyleForContext("return", typeInfo.Confidence)

	switch style {
	case StyleVerbose:
		return fmt.Sprintf(" # returns %s (confidence: %.0f%%)", typeInfo.Type.String(), typeInfo.Confidence*100)
	case StyleCompact:
		return fmt.Sprintf(" # -> %s", typeInfo.Type.String())
	case StyleCommentOnly:
		return fmt.Sprintf(" # type: %s", typeInfo.Type.String())
	default:
		return fmt.Sprintf(" # -> %s", typeInfo.Type.String())
	}
}

// determineStyleForContext chooses appropriate style based on context and confidence
func (ag *AnnotationGenerator) determineStyleForContext(context string, confidence float64) FormattingStyle {
	if !ag.options.ContextAware {
		return ag.options.PreferredStyle
	}

	// High confidence: use preferred style
	if confidence >= 0.9 {
		return ag.options.PreferredStyle
	}

	// Medium confidence: use more verbose style for clarity
	if confidence >= 0.7 {
		if ag.options.PreferredStyle == StyleCompact {
			return StyleInline
		}
		return ag.options.PreferredStyle
	}

	// Low confidence: prefer comments
	return StyleCommentOnly
}

// generateFallbackAnnotation generates a fallback annotation for low-confidence inferences
func (ag *AnnotationGenerator) generateFallbackAnnotation(varName string, typeInfo *types.TypeInfo) string {
	if typeInfo.Confidence > 0.4 {
		comment := ag.formatter.FormatConfidenceComment(typeInfo.Confidence)
		return fmt.Sprintf("my %s; %s", varName, comment)
	}
	return fmt.Sprintf("my %s;", varName)
}

// generateFallbackFieldAnnotation generates a fallback annotation for field declarations
func (ag *AnnotationGenerator) generateFallbackFieldAnnotation(fieldName string, typeInfo *types.TypeInfo) string {
	if typeInfo.Confidence > 0.4 {
		comment := ag.formatter.FormatConfidenceComment(typeInfo.Confidence)
		return fmt.Sprintf("field %s; %s", fieldName, comment)
	}
	return fmt.Sprintf("field %s;", fieldName)
}

// generateBasicMethodSignature generates a basic method signature without type annotations
func (ag *AnnotationGenerator) generateBasicMethodSignature(methodName string, params []ParameterInfo) string {
	var paramNames []string
	for _, param := range params {
		paramNames = append(paramNames, param.Name)
	}
	return fmt.Sprintf("sub %s(%s)", methodName, strings.Join(paramNames, ", "))
}

// MethodSignatureInfo holds comprehensive information about a method signature
type MethodSignatureInfo struct {
	Parameters        []ParameterInfo
	ReturnType        types.Type
	ReturnConfidence  float64
	OverallConfidence float64 // Average confidence across all type information
}

// AnnotationContext provides context information for annotation decisions
type AnnotationContext struct {
	// NodeType indicates the type of AST node being annotated
	NodeType string

	// Depth indicates nesting depth in the AST
	Depth int

	// ParentContext provides information about parent nodes
	ParentContext map[string]interface{}

	// LocalScope indicates if this is in a local scope
	LocalScope bool

	// IsPublic indicates if this declaration is publicly visible
	IsPublic bool
}
