// ABOUTME: Different formatting style implementations for type annotations
// ABOUTME: Provides specialized formatters for various output preferences and use cases

package compiler

import (
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
