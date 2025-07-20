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

	// Try string inference - but may return nil
	return li.inferStringType(cleaned)
}

// inferIntegerType creates type info for integer literals
func (li *LiteralInferrer) inferIntegerType(value string) *types.TypeInfo {
	intType := types.NewIntType()
	return types.NewTypeInfo(intType, 1.0, types.SourceLiteral)
}

// inferNumericType creates type info for floating-point literals
func (li *LiteralInferrer) inferNumericType(value string) *types.TypeInfo {
	numType := types.NewNumType()

	// Validate that it's actually a valid float - if not, don't infer
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		return nil // Failed to parse - don't infer
	}

	return types.NewTypeInfo(numType, 1.0, types.SourceLiteral)
}

// inferStringType creates type info for string literals
func (li *LiteralInferrer) inferStringType(value string) *types.TypeInfo {
	strType := types.NewStrType()

	// Only infer for clearly quoted strings
	if li.stringPattern.MatchString(value) {
		return types.NewTypeInfo(strType, 1.0, types.SourceLiteral)
	}
	
	// For unquoted strings, don't infer - too ambiguous in Perl
	return nil
}

// inferBooleanType creates type info for boolean literals
func (li *LiteralInferrer) inferBooleanType(value string) *types.TypeInfo {
	boolType := types.NewBoolType()

	// Only infer for clear boolean values
	switch strings.ToLower(value) {
	case "1", "0", "true", "false", "undef":
		return types.NewTypeInfo(boolType, 1.0, types.SourceLiteral)
	default:
		// Ambiguous value - don't infer
		return nil
	}
}

// inferArrayLiteralType infers type for array literals like [1, 2, 3]
func (li *LiteralInferrer) inferArrayLiteralType(elements []string) *types.TypeInfo {
	if len(elements) == 0 {
		// Empty array - don't infer type, too ambiguous
		return nil
	}

	// Infer element type from first element
	firstElement := elements[0]
	elementTypeInfo := li.InferLiteralType(firstElement)
	if elementTypeInfo == nil {
		// First element couldn't be inferred
		return nil
	}

	// Check if all elements have the same type
	allSameType := true
	for _, element := range elements[1:] {
		elemTypeInfo := li.InferLiteralType(element)
		if elemTypeInfo == nil || !elemTypeInfo.Type.Equals(elementTypeInfo.Type) {
			allSameType = false
			break
		}
	}

	var elementType types.Type
	if allSameType {
		// All elements have the same type
		elementType = elementTypeInfo.Type
	} else {
		// Mixed types - create union type
		elementTypes := make([]types.Type, 0)
		seenTypes := make(map[string]bool)

		// Collect unique types from all elements
		for _, element := range elements {
			elemTypeInfo := li.InferLiteralType(element)
			if elemTypeInfo != nil {
				typeStr := elemTypeInfo.Type.String()
				if !seenTypes[typeStr] {
					seenTypes[typeStr] = true
					elementTypes = append(elementTypes, elemTypeInfo.Type)
				}
			}
		}

		if len(elementTypes) == 0 {
			// No types could be inferred
			return nil
		} else if len(elementTypes) == 1 {
			elementType = elementTypes[0]
		} else {
			// Create union type
			elementType = types.NewUnionType(elementTypes...)
		}
	}

	arrayType := types.NewArrayRefType(elementType)
	return types.NewTypeInfo(arrayType, 1.0, types.SourceLiteral)
}

// inferHashLiteralType infers type for hash literals like {key => 'value'}
func (li *LiteralInferrer) inferHashLiteralType(values []string) *types.TypeInfo {
	if len(values) == 0 {
		// Empty hash - don't infer type, too ambiguous
		return nil
	}

	// Infer value type from first value
	firstValue := values[0]
	valueTypeInfo := li.InferLiteralType(firstValue)
	if valueTypeInfo == nil {
		// First value couldn't be inferred
		return nil
	}

	// Check if all values have the same type
	allSameType := true
	for _, value := range values[1:] {
		valTypeInfo := li.InferLiteralType(value)
		if valTypeInfo == nil || !valTypeInfo.Type.Equals(valueTypeInfo.Type) {
			allSameType = false
			break
		}
	}

	var valueType types.Type
	if allSameType {
		// All values have the same type
		valueType = valueTypeInfo.Type
	} else {
		// Mixed types - create union type
		valueTypes := make([]types.Type, 0)
		seenTypes := make(map[string]bool)

		// Collect unique types from all values
		for _, value := range values {
			valTypeInfo := li.InferLiteralType(value)
			if valTypeInfo != nil {
				typeStr := valTypeInfo.Type.String()
				if !seenTypes[typeStr] {
					seenTypes[typeStr] = true
					valueTypes = append(valueTypes, valTypeInfo.Type)
				}
			}
		}

		if len(valueTypes) == 0 {
			// No types could be inferred
			return nil
		} else if len(valueTypes) == 1 {
			valueType = valueTypes[0]
		} else {
			// Create union type
			valueType = types.NewUnionType(valueTypes...)
		}
	}

	hashType := types.NewHashRefType(valueType)
	return types.NewTypeInfo(hashType, 1.0, types.SourceLiteral)
}
