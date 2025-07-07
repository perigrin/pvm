// ABOUTME: Literal type inference implementation for basic types
// ABOUTME: Handles inference for string, number, boolean, and array literals

package inference

import (
	"regexp"
	"strconv"
	"strings"

	"tamarou.com/pvm/internal/types"
)

// LiteralInferrer handles type inference for literal values
type LiteralInferrer struct {
	// Regular expressions for literal detection
	intPattern     *regexp.Regexp
	floatPattern   *regexp.Regexp
	stringPattern  *regexp.Regexp
	booleanPattern *regexp.Regexp
}

// NewLiteralInferrer creates a new literal type inferrer
func NewLiteralInferrer() *LiteralInferrer {
	return &LiteralInferrer{
		intPattern:     regexp.MustCompile(`^-?\d+$`),
		floatPattern:   regexp.MustCompile(`^-?\d*\.\d+$`),
		stringPattern:  regexp.MustCompile(`^['"].*['"]$`),
		booleanPattern: regexp.MustCompile(`^(1|0|true|false|undef)$`),
	}
}

// InferLiteralType infers the type of a literal value
func (li *LiteralInferrer) InferLiteralType(literalValue string) *types.TypeInfo {
	// Clean up the literal value
	cleaned := strings.TrimSpace(literalValue)

	// Try to infer type based on patterns
	if li.intPattern.MatchString(cleaned) {
		return li.inferIntegerType(cleaned)
	}

	if li.floatPattern.MatchString(cleaned) {
		return li.inferNumericType(cleaned)
	}

	if li.stringPattern.MatchString(cleaned) {
		return li.inferStringType(cleaned)
	}

	if li.booleanPattern.MatchString(cleaned) {
		return li.inferBooleanType(cleaned)
	}

	// Default to string type for unknown literals
	return li.inferStringType(cleaned)
}

// inferIntegerType creates type info for integer literals
func (li *LiteralInferrer) inferIntegerType(value string) *types.TypeInfo {
	intType := types.NewIntType()

	// High confidence for clear integer patterns
	confidence := 0.95

	// Reduce confidence for very large numbers that might overflow
	if len(value) > 10 {
		confidence = 0.85
	}

	return types.NewTypeInfo(intType, confidence, types.SourceLiteral)
}

// inferNumericType creates type info for floating-point literals
func (li *LiteralInferrer) inferNumericType(value string) *types.TypeInfo {
	numType := types.NewNumType()

	// High confidence for clear float patterns
	confidence := 0.95

	// Validate that it's actually a valid float
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		confidence = 0.70 // Lower confidence if parsing fails
	}

	return types.NewTypeInfo(numType, confidence, types.SourceLiteral)
}

// inferStringType creates type info for string literals
func (li *LiteralInferrer) inferStringType(value string) *types.TypeInfo {
	strType := types.NewStrType()

	var confidence float64
	// Check if it's a quoted string
	if li.stringPattern.MatchString(value) {
		confidence = 0.98 // Very high confidence for quoted strings
	} else {
		confidence = 0.80 // Lower confidence for unquoted strings
	}

	return types.NewTypeInfo(strType, confidence, types.SourceLiteral)
}

// inferBooleanType creates type info for boolean literals
func (li *LiteralInferrer) inferBooleanType(value string) *types.TypeInfo {
	boolType := types.NewBoolType()

	var confidence float64
	// Perl's truthiness is complex, so be more conservative
	switch strings.ToLower(value) {
	case "1", "true":
		confidence = 0.95
	case "0", "false", "undef":
		confidence = 0.95
	default:
		confidence = 0.75
	}

	return types.NewTypeInfo(boolType, confidence, types.SourceLiteral)
}

// inferArrayLiteralType infers type for array literals like [1, 2, 3]
func (li *LiteralInferrer) inferArrayLiteralType(elements []string) *types.TypeInfo {
	if len(elements) == 0 {
		// Empty array - use generic ArrayRef
		arrayType := types.NewArrayRefType(types.NewStrType()) // Default to Str
		return types.NewTypeInfo(arrayType, 0.60, types.SourceLiteral)
	}

	// Infer element type from first element
	firstElement := elements[0]
	elementTypeInfo := li.InferLiteralType(firstElement)

	// Check if all elements have the same type
	confidence := elementTypeInfo.Confidence
	allSameType := true

	for _, element := range elements[1:] {
		elemTypeInfo := li.InferLiteralType(element)
		if !elemTypeInfo.Type.Equals(elementTypeInfo.Type) {
			allSameType = false
			break
		}
	}

	if !allSameType {
		// Mixed types - create union type
		confidence = 0.70
		elementTypes := make([]types.Type, 0)
		seenTypes := make(map[string]bool)

		// Collect unique types from all elements
		for _, element := range elements {
			elemTypeInfo := li.InferLiteralType(element)
			typeStr := elemTypeInfo.Type.String()
			if !seenTypes[typeStr] {
				seenTypes[typeStr] = true
				elementTypes = append(elementTypes, elemTypeInfo.Type)
			}
		}

		// Create union type if we have multiple distinct types
		var unionType types.Type
		if len(elementTypes) > 1 {
			unionType = types.NewUnionType(elementTypes...)
		} else {
			unionType = elementTypes[0]
		}

		elementTypeInfo = types.NewTypeInfo(unionType, confidence, types.SourceLiteral)
	}

	arrayType := types.NewArrayRefType(elementTypeInfo.Type)
	return types.NewTypeInfo(arrayType, confidence, types.SourceLiteral)
}

// inferHashLiteralType infers type for hash literals like {key => 'value'}
func (li *LiteralInferrer) inferHashLiteralType(values []string) *types.TypeInfo {
	if len(values) == 0 {
		// Empty hash - use generic HashRef
		hashType := types.NewHashRefType(types.NewStrType()) // Default to Str
		return types.NewTypeInfo(hashType, 0.60, types.SourceLiteral)
	}

	// Infer value type from first value
	firstValue := values[0]
	valueTypeInfo := li.InferLiteralType(firstValue)

	// Check if all values have the same type
	confidence := valueTypeInfo.Confidence
	allSameType := true

	for _, value := range values[1:] {
		valTypeInfo := li.InferLiteralType(value)
		if !valTypeInfo.Type.Equals(valueTypeInfo.Type) {
			allSameType = false
			break
		}
	}

	if !allSameType {
		// Mixed types - create union type
		confidence = 0.70
		valueTypes := make([]types.Type, 0)
		seenTypes := make(map[string]bool)

		// Collect unique types from all values
		for _, value := range values {
			valTypeInfo := li.InferLiteralType(value)
			typeStr := valTypeInfo.Type.String()
			if !seenTypes[typeStr] {
				seenTypes[typeStr] = true
				valueTypes = append(valueTypes, valTypeInfo.Type)
			}
		}

		// Create union type if we have multiple distinct types
		var unionType types.Type
		if len(valueTypes) > 1 {
			unionType = types.NewUnionType(valueTypes...)
		} else {
			unionType = valueTypes[0]
		}

		valueTypeInfo = types.NewTypeInfo(unionType, confidence, types.SourceLiteral)
	}

	hashType := types.NewHashRefType(valueTypeInfo.Type)
	return types.NewTypeInfo(hashType, confidence, types.SourceLiteral)
}
