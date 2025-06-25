// ABOUTME: Different formatting style implementations for type annotations
// ABOUTME: Provides specialized formatters for various output preferences and use cases

package compiler

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/types"
)

// StyleProvider provides specialized formatters for different output styles
type StyleProvider struct {
	formatter TypeFormatter
	options   FormatterOptions
}

// NewStyleProvider creates a new style provider with the given formatter
func NewStyleProvider(formatter TypeFormatter, options FormatterOptions) *StyleProvider {
	return &StyleProvider{
		formatter: formatter,
		options:   options,
	}
}

// InlineStyleFormatter specializes in inline type annotations
type InlineStyleFormatter struct {
	*basicTypeFormatter
}

// NewInlineStyleFormatter creates a formatter optimized for inline annotations
func NewInlineStyleFormatter() *InlineStyleFormatter {
	return &InlineStyleFormatter{
		basicTypeFormatter: &basicTypeFormatter{
			options: FormatterOptions{
				ConfidenceThreshold:       0.8,
				IncludeConfidenceComments: false,
				UseShortTypeNames:         false,
				PreferComments:            false,
			},
		},
	}
}

// VerboseStyleFormatter specializes in verbose, detailed annotations
type VerboseStyleFormatter struct {
	*basicTypeFormatter
}

// NewVerboseStyleFormatter creates a formatter optimized for verbose output
func NewVerboseStyleFormatter() *VerboseStyleFormatter {
	return &VerboseStyleFormatter{
		basicTypeFormatter: &basicTypeFormatter{
			options: FormatterOptions{
				ConfidenceThreshold:       0.6,
				IncludeConfidenceComments: true,
				UseShortTypeNames:         false,
				PreferComments:            false,
			},
		},
	}
}

// CompactStyleFormatter specializes in compact, minimal annotations
type CompactStyleFormatter struct {
	*basicTypeFormatter
}

// NewCompactStyleFormatter creates a formatter optimized for compact output
func NewCompactStyleFormatter() *CompactStyleFormatter {
	return &CompactStyleFormatter{
		basicTypeFormatter: &basicTypeFormatter{
			options: FormatterOptions{
				ConfidenceThreshold:       0.85,
				IncludeConfidenceComments: false,
				UseShortTypeNames:         true,
				PreferComments:            false,
			},
		},
	}
}

// CommentOnlyStyleFormatter specializes in comment-based type hints
type CommentOnlyStyleFormatter struct {
	*basicTypeFormatter
}

// NewCommentOnlyStyleFormatter creates a formatter that only uses comments
func NewCommentOnlyStyleFormatter() *CommentOnlyStyleFormatter {
	return &CommentOnlyStyleFormatter{
		basicTypeFormatter: &basicTypeFormatter{
			options: FormatterOptions{
				ConfidenceThreshold:       0.5,
				IncludeConfidenceComments: true,
				UseShortTypeNames:         false,
				PreferComments:            true,
			},
		},
	}
}

// FormatMethodSignature formats a method signature with type annotations
func (f *basicTypeFormatter) FormatMethodSignature(methodName string, params []ParameterInfo, returnType types.Type, style FormattingStyle) string {
	var paramStrs []string
	
	for _, param := range params {
		if f.ShouldIncludeAnnotation(param.Confidence, f.options.ConfidenceThreshold) {
			typeStr := f.formatTypeString(param.Type)
			paramStrs = append(paramStrs, fmt.Sprintf("%s %s", typeStr, param.Name))
		} else {
			paramStrs = append(paramStrs, param.Name)
		}
	}
	
	paramList := strings.Join(paramStrs, ", ")
	
	// Format return type if confident enough
	if returnType != nil && f.ShouldIncludeAnnotation(0.8, f.options.ConfidenceThreshold) {
		returnTypeStr := f.formatTypeString(returnType)
		switch style {
		case StyleVerbose:
			return fmt.Sprintf("sub %s(%s) -> %s", methodName, paramList, returnTypeStr)
		case StyleCompact:
			return fmt.Sprintf("sub %s(%s):%s", methodName, paramList, returnTypeStr)
		default:
			return fmt.Sprintf("sub %s(%s) # returns %s", methodName, paramList, returnTypeStr)
		}
	}
	
	return fmt.Sprintf("sub %s(%s)", methodName, paramList)
}

// FormatFieldDeclaration formats a field declaration with type annotation
func (f *basicTypeFormatter) FormatFieldDeclaration(fieldName string, fieldType types.Type, confidence float64, style FormattingStyle) string {
	if !f.ShouldIncludeAnnotation(confidence, f.options.ConfidenceThreshold) {
		if f.options.PreferComments && confidence > 0.4 {
			return fmt.Sprintf("field %s; %s", fieldName, f.FormatConfidenceComment(confidence))
		}
		return fmt.Sprintf("field %s;", fieldName)
	}
	
	typeStr := f.formatTypeString(fieldType)
	
	switch style {
	case StyleVerbose:
		comment := ""
		if f.options.IncludeConfidenceComments {
			comment = " " + f.FormatConfidenceComment(confidence)
		}
		return fmt.Sprintf("field %s %s;%s", typeStr, fieldName, comment)
		
	case StyleCompact:
		return fmt.Sprintf("field %s %s;", typeStr, fieldName)
		
	case StyleCommentOnly:
		comment := fmt.Sprintf("# field type: %s", typeStr)
		if f.options.IncludeConfidenceComments {
			comment += fmt.Sprintf(" (%.0f%%)", confidence*100)
		}
		return fmt.Sprintf("field %s; %s", fieldName, comment)
		
	default: // StyleInline
		return fmt.Sprintf("field %s %s;", typeStr, fieldName)
	}
}

// ParameterInfo holds information about a method parameter
type ParameterInfo struct {
	Name       string
	Type       types.Type
	Confidence float64
	IsOptional bool
}

// StylePreferences holds user preferences for formatting styles
type StylePreferences struct {
	// DefaultStyle is the preferred style for most annotations
	DefaultStyle FormattingStyle
	
	// VariableStyle is the style for variable declarations
	VariableStyle FormattingStyle
	
	// MethodStyle is the style for method signatures
	MethodStyle FormattingStyle
	
	// FieldStyle is the style for field declarations
	FieldStyle FormattingStyle
	
	// UncertaintyStyle is how to handle low-confidence inferences
	UncertaintyStyle FormattingStyle
}