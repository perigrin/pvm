// ABOUTME: Type constraint system for complex type relationships
// ABOUTME: Defines constraints used in type checking and inference

package types

import "fmt"

// TypeConstraint represents a constraint between types
type TypeConstraint interface {
	// Validate checks if the constraint is satisfied
	Validate() error

	// String returns a string representation of the constraint
	String() string

	// GetSourceType returns the source type in the constraint
	GetSourceType() Type

	// GetTargetType returns the target type in the constraint
	GetTargetType() Type
}

// AssignmentConstraint represents a constraint for variable assignment
type AssignmentConstraint struct {
	sourceType Type
	targetType Type
}

// NewAssignmentConstraint creates a new assignment constraint
func NewAssignmentConstraint(sourceType, targetType Type) TypeConstraint {
	return &AssignmentConstraint{
		sourceType: sourceType,
		targetType: targetType,
	}
}

// Validate checks if the assignment is valid
func (ac *AssignmentConstraint) Validate() error {
	if !ac.sourceType.CompatibleWith(ac.targetType) {
		return fmt.Errorf("type mismatch: cannot assign %s to %s",
			ac.sourceType.String(), ac.targetType.String())
	}
	return nil
}

// String returns a string representation of the assignment constraint
func (ac *AssignmentConstraint) String() string {
	return fmt.Sprintf("%s := %s", ac.targetType.String(), ac.sourceType.String())
}

// GetSourceType returns the source type
func (ac *AssignmentConstraint) GetSourceType() Type {
	return ac.sourceType
}

// GetTargetType returns the target type
func (ac *AssignmentConstraint) GetTargetType() Type {
	return ac.targetType
}

// FunctionCallConstraint represents a constraint for function calls
type FunctionCallConstraint struct {
	argumentTypes  []Type
	parameterTypes []Type
	returnType     Type
}

// NewFunctionCallConstraint creates a new function call constraint
func NewFunctionCallConstraint(argumentTypes, parameterTypes []Type, returnType Type) TypeConstraint {
	return &FunctionCallConstraint{
		argumentTypes:  argumentTypes,
		parameterTypes: parameterTypes,
		returnType:     returnType,
	}
}

// Validate checks if the function call is valid
func (fcc *FunctionCallConstraint) Validate() error {
	if len(fcc.argumentTypes) != len(fcc.parameterTypes) {
		return fmt.Errorf("argument count mismatch: expected %d, got %d",
			len(fcc.parameterTypes), len(fcc.argumentTypes))
	}

	for i, argType := range fcc.argumentTypes {
		paramType := fcc.parameterTypes[i]
		if !argType.CompatibleWith(paramType) {
			return fmt.Errorf("argument type mismatch at position %d: expected %s, got %s",
				i, paramType.String(), argType.String())
		}
	}

	return nil
}

// String returns a string representation of the function call constraint
func (fcc *FunctionCallConstraint) String() string {
	args := ""
	for i, argType := range fcc.argumentTypes {
		if i > 0 {
			args += ", "
		}
		args += argType.String()
	}
	return fmt.Sprintf("call(%s) -> %s", args, fcc.returnType.String())
}

// GetSourceType returns the first argument type (if any)
func (fcc *FunctionCallConstraint) GetSourceType() Type {
	if len(fcc.argumentTypes) > 0 {
		return fcc.argumentTypes[0]
	}
	return nil
}

// GetTargetType returns the return type
func (fcc *FunctionCallConstraint) GetTargetType() Type {
	return fcc.returnType
}

// ConstraintSet manages a collection of type constraints
type ConstraintSet struct {
	constraints []TypeConstraint
}

// NewConstraintSet creates a new constraint set
func NewConstraintSet() *ConstraintSet {
	return &ConstraintSet{
		constraints: make([]TypeConstraint, 0),
	}
}

// Add adds a constraint to the set
func (cs *ConstraintSet) Add(constraint TypeConstraint) {
	cs.constraints = append(cs.constraints, constraint)
}

// ValidateAll validates all constraints in the set
func (cs *ConstraintSet) ValidateAll() []error {
	var errors []error
	for _, constraint := range cs.constraints {
		if err := constraint.Validate(); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// GetConstraints returns all constraints in the set
func (cs *ConstraintSet) GetConstraints() []TypeConstraint {
	return cs.constraints
}

// Size returns the number of constraints in the set
func (cs *ConstraintSet) Size() int {
	return len(cs.constraints)
}
