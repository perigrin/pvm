// ABOUTME: TypeInfo and metadata structures for tracking type inference information
// ABOUTME: Provides confidence tracking and source attribution for inferred types

package types

import "fmt"

// TypeSource represents how a type was inferred or determined
type TypeSource string

const (
	// SourceLiteral indicates the type was inferred from a literal value
	SourceLiteral TypeSource = "literal"

	// SourceVariable indicates the type was inferred from a variable declaration
	SourceVariable TypeSource = "variable"

	// SourceReturn indicates the type was inferred from a return value
	SourceReturn TypeSource = "return"

	// SourceParameter indicates the type was inferred from a function parameter
	SourceParameter TypeSource = "parameter"

	// SourceContext indicates the type was inferred from context
	SourceContext TypeSource = "context"

	// SourceExternal indicates the type came from external type definitions
	SourceExternal TypeSource = "external"
)

// String returns the string representation of the TypeSource
func (s TypeSource) String() string {
	return string(s)
}

// TypeInfo contains type information with confidence and source tracking
type TypeInfo struct {
	// Type is the inferred type
	Type Type

	// Confidence is a value between 0.0 and 1.0 indicating confidence in the type inference
	Confidence float64

	// Source indicates how this type was inferred
	Source TypeSource

	// Location contains source location information (file, line, column)
	Location *SourceLocation

	// Context contains additional context information
	Context map[string]interface{}
}

// SourceLocation represents a location in source code
type SourceLocation struct {
	File   string
	Line   int
	Column int
}

// NewTypeInfo creates a new TypeInfo with the given type, confidence, and source
func NewTypeInfo(typ Type, confidence float64, source TypeSource) *TypeInfo {
	return &TypeInfo{
		Type:       typ,
		Confidence: confidence,
		Source:     source,
		Context:    make(map[string]interface{}),
	}
}

// WithLocation adds source location information to the TypeInfo
func (ti *TypeInfo) WithLocation(file string, line, column int) *TypeInfo {
	ti.Location = &SourceLocation{
		File:   file,
		Line:   line,
		Column: column,
	}
	return ti
}

// WithContext adds context information to the TypeInfo
func (ti *TypeInfo) WithContext(key string, value interface{}) *TypeInfo {
	ti.Context[key] = value
	return ti
}

// String returns a string representation of the TypeInfo
func (ti *TypeInfo) String() string {
	return fmt.Sprintf("%s (confidence: %.2f, source: %s)",
		ti.Type.String(), ti.Confidence, ti.Source)
}

// IsHighConfidence returns true if the confidence is >= 0.8
func (ti *TypeInfo) IsHighConfidence() bool {
	return ti.Confidence >= 0.8
}

// IsMediumConfidence returns true if the confidence is >= 0.5 and < 0.8
func (ti *TypeInfo) IsMediumConfidence() bool {
	return ti.Confidence >= 0.5 && ti.Confidence < 0.8
}

// IsLowConfidence returns true if the confidence is < 0.5
func (ti *TypeInfo) IsLowConfidence() bool {
	return ti.Confidence < 0.5
}

// Merge combines this TypeInfo with another, taking the higher confidence
func (ti *TypeInfo) Merge(other *TypeInfo) *TypeInfo {
	if other.Confidence > ti.Confidence {
		return other
	}
	return ti
}

// Copy creates a copy of the TypeInfo
func (ti *TypeInfo) Copy() *TypeInfo {
	context := make(map[string]interface{})
	for k, v := range ti.Context {
		context[k] = v
	}

	var location *SourceLocation
	if ti.Location != nil {
		location = &SourceLocation{
			File:   ti.Location.File,
			Line:   ti.Location.Line,
			Column: ti.Location.Column,
		}
	}

	return &TypeInfo{
		Type:       ti.Type,
		Confidence: ti.Confidence,
		Source:     ti.Source,
		Location:   location,
		Context:    context,
	}
}
