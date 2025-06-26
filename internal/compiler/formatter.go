// ABOUTME: Type formatting system that converts type information back into Perl syntax
// ABOUTME: Handles confidence-based annotation decisions and multiple output styles

package compiler

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/types"
)

// TypeFormatter handles conversion of type information into Perl code annotations
type TypeFormatter interface {
	// FormatTypeAnnotation formats a type as a Perl type annotation
	FormatTypeAnnotation(typeInfo *types.TypeInfo, style FormattingStyle) string

	// FormatVariableDeclaration formats a variable declaration with type annotation
	FormatVariableDeclaration(varName string, varType types.Type, confidence float64, style FormattingStyle) string

	// FormatFieldDeclaration formats a field declaration with type annotation
	FormatFieldDeclaration(fieldName string, fieldType types.Type, confidence float64, style FormattingStyle) string

	// FormatConfidenceComment formats a comment indicating type inference confidence
	FormatConfidenceComment(confidence float64) string

	// ShouldIncludeAnnotation determines if annotation should be included based on confidence
	ShouldIncludeAnnotation(confidence float64, threshold float64) bool
}

// FormattingStyle represents different type annotation formatting styles
type FormattingStyle string

const (
	// StyleInline places type annotations inline with variable declarations
	StyleInline FormattingStyle = "inline"

	// StyleVerbose includes confidence information and detailed annotations
	StyleVerbose FormattingStyle = "verbose"

	// StyleCompact produces minimal, clean type annotations
	StyleCompact FormattingStyle = "compact"

	// StyleCommentOnly places type information only in comments
	StyleCommentOnly FormattingStyle = "comment_only"
)

// basicTypeFormatter implements TypeFormatter with standard formatting
type basicTypeFormatter struct {
	// Configuration options
	options FormatterOptions
}

// FormatterOptions holds configuration for the type formatter
type FormatterOptions struct {
	// ConfidenceThreshold sets minimum confidence for type annotations
	ConfidenceThreshold float64

	// IncludeConfidenceComments controls whether confidence info is included
	IncludeConfidenceComments bool

	// UseShortTypeNames uses abbreviated type names when possible
	UseShortTypeNames bool

	// PreferComments puts uncertain types in comments rather than annotations
	PreferComments bool
}

// NewTypeFormatter creates a new type formatter with default options
func NewTypeFormatter() TypeFormatter {
	return &basicTypeFormatter{
		options: FormatterOptions{
			ConfidenceThreshold:       0.7,
			IncludeConfidenceComments: false,
			UseShortTypeNames:         false,
			PreferComments:            true,
		},
	}
}

// NewTypeFormatterWithOptions creates a formatter with custom options
func NewTypeFormatterWithOptions(options FormatterOptions) TypeFormatter {
	return &basicTypeFormatter{
		options: options,
	}
}

// FormatTypeAnnotation formats a type as a Perl type annotation
func (f *basicTypeFormatter) FormatTypeAnnotation(typeInfo *types.TypeInfo, style FormattingStyle) string {
	switch style {
	case StyleVerbose:
		return f.formatVerboseAnnotation(typeInfo)
	case StyleCompact:
		return f.formatCompactAnnotation(typeInfo)
	case StyleCommentOnly:
		return f.formatCommentOnlyAnnotation(typeInfo)
	default: // StyleInline
		return f.formatInlineAnnotation(typeInfo)
	}
}

// FormatVariableDeclaration formats a variable declaration with type annotation
func (f *basicTypeFormatter) FormatVariableDeclaration(varName string, varType types.Type, confidence float64, style FormattingStyle) string {
	typeInfo := &types.TypeInfo{
		Type:       varType,
		Confidence: confidence,
		Source:     types.SourceVariable,
	}

	if !f.ShouldIncludeAnnotation(confidence, f.options.ConfidenceThreshold) {
		if f.options.PreferComments && confidence > 0.4 {
			return fmt.Sprintf("my %s; %s", varName, f.FormatConfidenceComment(confidence))
		}
		return fmt.Sprintf("my %s;", varName)
	}

	switch style {
	case StyleVerbose:
		annotation := f.formatVerboseAnnotation(typeInfo)
		comment := ""
		if f.options.IncludeConfidenceComments {
			comment = " " + f.FormatConfidenceComment(confidence)
		}
		return fmt.Sprintf("my %s %s;%s", annotation, varName, comment)

	case StyleCompact:
		annotation := f.formatCompactAnnotation(typeInfo)
		return fmt.Sprintf("my %s %s;", annotation, varName)

	case StyleCommentOnly:
		comment := f.formatCommentOnlyAnnotation(typeInfo)
		return fmt.Sprintf("my %s; %s", varName, comment)

	default: // StyleInline
		annotation := f.formatInlineAnnotation(typeInfo)
		return fmt.Sprintf("my %s %s;", annotation, varName)
	}
}

// FormatConfidenceComment formats a comment indicating type inference confidence
func (f *basicTypeFormatter) FormatConfidenceComment(confidence float64) string {
	confidencePercent := int(confidence * 100)

	switch {
	case confidence >= 0.9:
		return fmt.Sprintf("# type inferred: high confidence (%d%%)", confidencePercent)
	case confidence >= 0.7:
		return fmt.Sprintf("# type inferred: medium confidence (%d%%)", confidencePercent)
	case confidence >= 0.5:
		return fmt.Sprintf("# type inferred: low confidence (%d%%)", confidencePercent)
	default:
		return fmt.Sprintf("# type uncertain (%d%%)", confidencePercent)
	}
}

// ShouldIncludeAnnotation determines if annotation should be included based on confidence
func (f *basicTypeFormatter) ShouldIncludeAnnotation(confidence float64, threshold float64) bool {
	return confidence >= threshold
}

// formatInlineAnnotation formats type for inline style
func (f *basicTypeFormatter) formatInlineAnnotation(typeInfo *types.TypeInfo) string {
	return f.formatTypeString(typeInfo.Type)
}

// formatVerboseAnnotation formats type for verbose style
func (f *basicTypeFormatter) formatVerboseAnnotation(typeInfo *types.TypeInfo) string {
	typeStr := f.formatTypeString(typeInfo.Type)

	// Add source information in verbose mode
	if typeInfo.Source != "" {
		return fmt.Sprintf("%s", typeStr) // Keep simple for now
	}

	return typeStr
}

// formatCompactAnnotation formats type for compact style
func (f *basicTypeFormatter) formatCompactAnnotation(typeInfo *types.TypeInfo) string {
	typeStr := f.formatTypeString(typeInfo.Type)

	if f.options.UseShortTypeNames {
		// Use shorter type names in compact mode
		typeStr = strings.ReplaceAll(typeStr, "ArrayRef", "Arr")
		typeStr = strings.ReplaceAll(typeStr, "HashRef", "Hash")
	}

	return typeStr
}

// formatCommentOnlyAnnotation formats type for comment-only style
func (f *basicTypeFormatter) formatCommentOnlyAnnotation(typeInfo *types.TypeInfo) string {
	typeStr := f.formatTypeString(typeInfo.Type)
	confidenceStr := ""

	if f.options.IncludeConfidenceComments {
		confidenceStr = fmt.Sprintf(" (%.0f%%)", typeInfo.Confidence*100)
	}

	return fmt.Sprintf("# type: %s%s", typeStr, confidenceStr)
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

// formatTypeString converts a Type to its string representation
func (f *basicTypeFormatter) formatTypeString(t types.Type) string {
	return t.String()
}
