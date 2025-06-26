// ABOUTME: Adapter for wrapping parser AST with type information capabilities
// ABOUTME: Provides seamless integration between parser output and type inference system

package ast

import (
	"tamarou.com/pvm/internal/types"
)

// ASTTypeAdapter wraps any AST-like structure with type information capabilities
type ASTTypeAdapter struct {
	// The underlying AST (could be from parser or other source)
	baseAST interface {
		GetPath() string
		IsValid() bool
		GetContent() (string, error)
		GetRootNode() (Node, error)
	}

	// Type information storage
	typeInfo map[string]*types.TypeInfo

	// Type constraint storage
	constraints map[string]types.TypeConstraint
}

// NewASTTypeAdapter creates a new adapter for the given AST
func NewASTTypeAdapter(baseAST *AST) *ASTTypeAdapter {
	return &ASTTypeAdapter{
		baseAST:     baseAST,
		typeInfo:    make(map[string]*types.TypeInfo),
		constraints: make(map[string]types.TypeConstraint),
	}
}

// Delegate basic AST interface methods

// GetPath returns the source file path
func (ata *ASTTypeAdapter) GetPath() string {
	return ata.baseAST.GetPath()
}

// IsValid returns true if the AST is valid for compilation
func (ata *ASTTypeAdapter) IsValid() bool {
	return ata.baseAST.IsValid()
}

// GetContent returns the original source content
func (ata *ASTTypeAdapter) GetContent() (string, error) {
	return ata.baseAST.GetContent()
}

// GetRootNode returns the root AST node
func (ata *ASTTypeAdapter) GetRootNode() (Node, error) {
	return ata.baseAST.GetRootNode()
}

// Type information methods

// GetTypeInfo retrieves type information for a specific node
func (ata *ASTTypeAdapter) GetTypeInfo(nodeID string) *types.TypeInfo {
	typeInfo, exists := ata.typeInfo[nodeID]
	if !exists {
		return nil
	}
	return typeInfo.Copy()
}

// GetAllTypeInfo returns all type information
func (ata *ASTTypeAdapter) GetAllTypeInfo() map[string]*types.TypeInfo {
	result := make(map[string]*types.TypeInfo)
	for nodeID, typeInfo := range ata.typeInfo {
		result[nodeID] = typeInfo.Copy()
	}
	return result
}

// AttachTypeInfo attaches type information to a node
func (ata *ASTTypeAdapter) AttachTypeInfo(nodeID string, typeInfo *types.TypeInfo) error {
	if typeInfo == nil {
		return NewTypeInfoError("type information cannot be nil")
	}
	ata.typeInfo[nodeID] = typeInfo.Copy()
	return nil
}

// HasTypeInfo checks if type information exists for a node
func (ata *ASTTypeAdapter) HasTypeInfo(nodeID string) bool {
	_, exists := ata.typeInfo[nodeID]
	return exists
}

// RemoveTypeInfo removes type information for a node
func (ata *ASTTypeAdapter) RemoveTypeInfo(nodeID string) bool {
	if _, exists := ata.typeInfo[nodeID]; exists {
		delete(ata.typeInfo, nodeID)
		return true
	}
	return false
}

// Type constraint methods

// AddTypeConstraint adds a type constraint
func (ata *ASTTypeAdapter) AddTypeConstraint(constraintID string, constraint types.TypeConstraint) error {
	if constraint == nil {
		return NewTypeInfoError("constraint cannot be nil")
	}
	ata.constraints[constraintID] = constraint
	return nil
}

// GetTypeConstraints returns all type constraints
func (ata *ASTTypeAdapter) GetTypeConstraints() map[string]types.TypeConstraint {
	result := make(map[string]types.TypeConstraint)
	for constraintID, constraint := range ata.constraints {
		result[constraintID] = constraint
	}
	return result
}

// ValidateTypeConstraints validates all type constraints
func (ata *ASTTypeAdapter) ValidateTypeConstraints() []error {
	var errors []error
	for _, constraint := range ata.constraints {
		if err := constraint.Validate(); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// ConvertToInferredAST converts the adapter to an InferredAST interface
func (ata *ASTTypeAdapter) ConvertToInferredAST() InferredAST {
	// Create a new InferredAST from the base AST
	if baseAST, ok := ata.baseAST.(*AST); ok {
		inferredAST := NewInferredAST(baseAST)

		// Copy all type information
		for nodeID, typeInfo := range ata.typeInfo {
			inferredAST.AttachTypeInfo(nodeID, typeInfo)
		}

		// Copy all constraints
		for constraintID, constraint := range ata.constraints {
			inferredAST.AddTypeConstraint(constraintID, constraint)
		}

		return inferredAST
	}

	// If base AST is not our AST type, we can't convert directly
	// This would need additional handling for other AST types
	return nil
}
