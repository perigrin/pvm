// ABOUTME: InferredAST interface and implementation for type inference integration
// ABOUTME: Extends existing AST with type information access and constraint management

package ast

import (
	"tamarou.com/pvm/internal/types"
)

// InferredAST extends the basic AST interface with type inference capabilities
type InferredAST interface {
	// Embed the compiler AST interface for backward compatibility
	GetPath() string
	IsValid() bool
	GetContent() (string, error)
	GetRootNode() (Node, error)

	// Type information access methods
	GetTypeInfo(nodeID string) *types.TypeInfo
	GetAllTypeInfo() map[string]*types.TypeInfo
	AttachTypeInfo(nodeID string, typeInfo *types.TypeInfo) error

	// Type constraint management
	AddTypeConstraint(constraintID string, constraint types.TypeConstraint) error
	GetTypeConstraints() map[string]types.TypeConstraint
	ValidateTypeConstraints() []error

	// Additional utility methods
	HasTypeInfo(nodeID string) bool
	RemoveTypeInfo(nodeID string) bool
	ClearAllTypeInfo()
}

// inferredASTImpl implements the InferredAST interface
type inferredASTImpl struct {
	// Embed the original AST for delegation
	baseAST *AST

	// Type information storage
	typeInfo map[string]*types.TypeInfo

	// Type constraint storage
	constraints map[string]types.TypeConstraint
}

// NewInferredAST creates a new InferredAST wrapping the given base AST
func NewInferredAST(baseAST *AST) InferredAST {
	return &inferredASTImpl{
		baseAST:     baseAST,
		typeInfo:    make(map[string]*types.TypeInfo),
		constraints: make(map[string]types.TypeConstraint),
	}
}

// Delegate basic AST interface methods to the base AST

// GetPath returns the source file path
func (ia *inferredASTImpl) GetPath() string {
	return ia.baseAST.GetPath()
}

// IsValid returns true if the AST is valid for compilation
func (ia *inferredASTImpl) IsValid() bool {
	return ia.baseAST.IsValid()
}

// GetContent returns the original source content
func (ia *inferredASTImpl) GetContent() (string, error) {
	return ia.baseAST.GetContent()
}

// GetRootNode returns the root AST node
func (ia *inferredASTImpl) GetRootNode() (Node, error) {
	return ia.baseAST.GetRootNode()
}

// Type information access methods

// GetTypeInfo retrieves type information for a specific node
func (ia *inferredASTImpl) GetTypeInfo(nodeID string) *types.TypeInfo {
	typeInfo, exists := ia.typeInfo[nodeID]
	if !exists {
		return nil
	}
	// Return a copy to prevent external modification
	return typeInfo.Copy()
}

// GetAllTypeInfo returns all type information as a map
func (ia *inferredASTImpl) GetAllTypeInfo() map[string]*types.TypeInfo {
	result := make(map[string]*types.TypeInfo)
	for nodeID, typeInfo := range ia.typeInfo {
		result[nodeID] = typeInfo.Copy()
	}
	return result
}

// AttachTypeInfo attaches type information to a specific node
func (ia *inferredASTImpl) AttachTypeInfo(nodeID string, typeInfo *types.TypeInfo) error {
	if typeInfo == nil {
		return NewTypeInfoError("type information cannot be nil")
	}
	ia.typeInfo[nodeID] = typeInfo.Copy()
	return nil
}

// HasTypeInfo checks if type information exists for a node
func (ia *inferredASTImpl) HasTypeInfo(nodeID string) bool {
	_, exists := ia.typeInfo[nodeID]
	return exists
}

// RemoveTypeInfo removes type information for a node
func (ia *inferredASTImpl) RemoveTypeInfo(nodeID string) bool {
	if _, exists := ia.typeInfo[nodeID]; exists {
		delete(ia.typeInfo, nodeID)
		return true
	}
	return false
}

// ClearAllTypeInfo removes all type information
func (ia *inferredASTImpl) ClearAllTypeInfo() {
	ia.typeInfo = make(map[string]*types.TypeInfo)
}

// Type constraint management methods

// AddTypeConstraint adds a type constraint
func (ia *inferredASTImpl) AddTypeConstraint(constraintID string, constraint types.TypeConstraint) error {
	if constraint == nil {
		return NewTypeInfoError("constraint cannot be nil")
	}
	ia.constraints[constraintID] = constraint
	return nil
}

// GetTypeConstraints returns all type constraints
func (ia *inferredASTImpl) GetTypeConstraints() map[string]types.TypeConstraint {
	result := make(map[string]types.TypeConstraint)
	for constraintID, constraint := range ia.constraints {
		result[constraintID] = constraint
	}
	return result
}

// ValidateTypeConstraints validates all type constraints
func (ia *inferredASTImpl) ValidateTypeConstraints() []error {
	var errors []error
	for _, constraint := range ia.constraints {
		if err := constraint.Validate(); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// TypeInfoError represents an error in type information handling
type TypeInfoError struct {
	message string
}

// NewTypeInfoError creates a new TypeInfoError
func NewTypeInfoError(message string) *TypeInfoError {
	return &TypeInfoError{message: message}
}

// Error implements the error interface
func (e *TypeInfoError) Error() string {
	return "type info error: " + e.message
}