// ABOUTME: Operation resolution system for trait-based type checking
// ABOUTME: Determines operation support and result types using trait system

package traits

import (
	"fmt"
)

// OperationResolver handles operation validation and result type resolution
type OperationResolver struct {
	// typeTraits maps type names to their trait sets
	typeTraits map[string]*TraitSet
}

// NewOperationResolver creates a new operation resolver with default type traits
func NewOperationResolver() *OperationResolver {
	resolver := &OperationResolver{
		typeTraits: make(map[string]*TraitSet),
	}

	// Initialize with default traits for basic types
	basicTypes := []string{"Int", "Str", "Num", "Bool", "ArrayRef", "HashRef"}
	for _, typeName := range basicTypes {
		resolver.typeTraits[typeName] = GetDefaultTraitsForType(typeName)
	}

	return resolver
}

// SetTraitsForType sets the traits for a specific type
func (r *OperationResolver) SetTraitsForType(typeName string, traits *TraitSet) {
	r.typeTraits[typeName] = traits
}

// GetTraitsForType returns the traits for a type, or default traits if unknown
func (r *OperationResolver) GetTraitsForType(typeName string) *TraitSet {
	if traits, exists := r.typeTraits[typeName]; exists {
		return traits
	}
	// Return default traits for unknown types (basic conversions only)
	return GetDefaultTraitsForType(typeName)
}

// IsOperationSupported checks if a type supports a given operation
func (r *OperationResolver) IsOperationSupported(typeName, operation string) bool {
	traits := r.GetTraitsForType(typeName)
	return traits.HasTrait(operation)
}

// GetResultType returns the result type for an operation on a type
// Returns empty string if the operation is not supported
func (r *OperationResolver) GetResultType(typeName, operation string) string {
	traits := r.GetTraitsForType(typeName)
	return traits.GetResultType(operation)
}

// ResolveOperation resolves a binary operation between two types
// For now, this focuses on the left operand's capabilities
// Returns the result type or an error if the operation is not supported
func (r *OperationResolver) ResolveOperation(leftType, operation, rightType string) (string, error) {
	// Check if the left type supports the operation
	if !r.IsOperationSupported(leftType, operation) {
		return "", fmt.Errorf("operation '%s' not supported on type '%s'", operation, leftType)
	}

	// Get the result type from the left operand's traits
	resultType := r.GetResultType(leftType, operation)
	if resultType == "" {
		return "", fmt.Errorf("operation '%s' on type '%s' has no defined result type", operation, leftType)
	}

	return resultType, nil
}

// ValidateUnaryOperation validates a unary operation on a single type
func (r *OperationResolver) ValidateUnaryOperation(typeName, operation string) error {
	if !r.IsOperationSupported(typeName, operation) {
		return fmt.Errorf("unary operation '%s' not supported on type '%s'", operation, typeName)
	}
	return nil
}

// GetSupportedOperations returns all operations supported by a type
func (r *OperationResolver) GetSupportedOperations(typeName string) []string {
	traits := r.GetTraitsForType(typeName)
	allTraits := traits.GetAllTraits()

	var operations []string
	for _, trait := range allTraits {
		operations = append(operations, trait.Operation)
	}

	return operations
}

// GetOperationInfo returns detailed information about an operation on a type
type OperationInfo struct {
	Operation   string
	Supported   bool
	ResultType  string
	ErrorReason string
}

// GetOperationInfo provides detailed information about an operation
func (r *OperationResolver) GetOperationInfo(typeName, operation string) OperationInfo {
	info := OperationInfo{
		Operation: operation,
		Supported: r.IsOperationSupported(typeName, operation),
	}

	if info.Supported {
		info.ResultType = r.GetResultType(typeName, operation)
	} else {
		info.ErrorReason = fmt.Sprintf("Type '%s' does not support operation '%s'", typeName, operation)
	}

	return info
}

// ValidateOperationSequence validates a sequence of operations
// This is useful for checking complex expressions
func (r *OperationResolver) ValidateOperationSequence(startType string, operations []string) (string, error) {
	currentType := startType

	for i, operation := range operations {
		if !r.IsOperationSupported(currentType, operation) {
			return "", fmt.Errorf("operation %d: '%s' not supported on type '%s'", i+1, operation, currentType)
		}

		// Get the result type for the next iteration
		resultType := r.GetResultType(currentType, operation)
		if resultType == "" {
			return "", fmt.Errorf("operation %d: '%s' on type '%s' has no defined result type", i+1, operation, currentType)
		}

		currentType = resultType
	}

	return currentType, nil
}

// CompareOperationSupport compares operation support between two types
type OperationComparison struct {
	Operation      string
	LeftSupported  bool
	RightSupported bool
	LeftResult     string
	RightResult    string
}

// CompareTypes compares the operation support between two types
func (r *OperationResolver) CompareTypes(leftType, rightType string, operations []string) []OperationComparison {
	var comparisons []OperationComparison

	for _, operation := range operations {
		comparison := OperationComparison{
			Operation:      operation,
			LeftSupported:  r.IsOperationSupported(leftType, operation),
			RightSupported: r.IsOperationSupported(rightType, operation),
		}

		if comparison.LeftSupported {
			comparison.LeftResult = r.GetResultType(leftType, operation)
		}

		if comparison.RightSupported {
			comparison.RightResult = r.GetResultType(rightType, operation)
		}

		comparisons = append(comparisons, comparison)
	}

	return comparisons
}
