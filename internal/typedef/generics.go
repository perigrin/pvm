// ABOUTME: Implements generic type system with type parameters and constraints
// ABOUTME: Provides type parameter substitution, constraint checking, and inference capabilities

package typedef

import (
	"fmt"
	"strings"
)

// Type represents any type in the PVM type system
type Type interface {
	// GetName returns the name of the type
	GetName() string

	// IsCompatible checks if this type is compatible with another type
	IsCompatible(other Type) bool
}

// SimpleType represents a simple, non-parameterized type
type SimpleType struct {
	Name string
}

func (s *SimpleType) GetName() string {
	return s.Name
}

func (s *SimpleType) IsCompatible(other Type) bool {
	if other == nil {
		return false
	}
	// Simple compatibility check - same name
	return s.Name == other.GetName()
}

// ParameterizedType represents a parameterized type like Array[T]
type ParameterizedType struct {
	BaseType  Type
	Arguments []Type
}

func (p *ParameterizedType) GetName() string {
	if p.BaseType == nil {
		return ""
	}

	if len(p.Arguments) == 0 {
		return p.BaseType.GetName()
	}

	argNames := make([]string, len(p.Arguments))
	for i, arg := range p.Arguments {
		if arg != nil {
			argNames[i] = arg.GetName()
		}
	}

	return fmt.Sprintf("%s[%s]", p.BaseType.GetName(), strings.Join(argNames, ", "))
}

func (p *ParameterizedType) IsCompatible(other Type) bool {
	if other == nil {
		return false
	}

	// If other is also parameterized, check structural compatibility
	if otherParam, ok := other.(*ParameterizedType); ok {
		if p.BaseType == nil || otherParam.BaseType == nil {
			return false
		}

		// Check base type compatibility
		if !p.BaseType.IsCompatible(otherParam.BaseType) {
			return false
		}

		// Check argument compatibility
		if len(p.Arguments) != len(otherParam.Arguments) {
			return false
		}

		for i, arg := range p.Arguments {
			if !arg.IsCompatible(otherParam.Arguments[i]) {
				return false
			}
		}

		return true
	}

	// Check if base type is compatible
	if p.BaseType != nil {
		return p.BaseType.IsCompatible(other)
	}

	return false
}

// GenericType represents a generic type with type parameters and constraints
type GenericType struct {
	Name           string
	TypeParameters []TypeParameter
	Constraints    []TypeConstraint
	BaseType       Type
}

// TypeParameter represents a type parameter like T, U, etc.
type TypeParameter struct {
	Name string
	// Bounds represent constraints on this type parameter
	Bounds []Type
}

// TypeConstraint represents a constraint on a type parameter
type TypeConstraint struct {
	Parameter  string
	Kind       ConstraintKind
	Expression Type
}

// ConstraintKind represents different kinds of type constraints
type ConstraintKind int

const (
	// TraitConstraint represents constraints like T: Serializable
	TraitConstraint ConstraintKind = iota
	// ProtocolConstraint represents constraints like T does EventHandler
	ProtocolConstraint
	// CapabilityConstraint represents constraints like T can 'serialize'
	CapabilityConstraint
	// ValueConstraint represents constraints like $size > 0
	ValueConstraint
)

// GenericTypeChecker provides methods for working with generic types
type GenericTypeChecker struct {
	// constraintCache caches constraint validation results
	constraintCache map[string]bool
}

// NewGenericTypeChecker creates a new generic type checker
func NewGenericTypeChecker() *GenericTypeChecker {
	return &GenericTypeChecker{
		constraintCache: make(map[string]bool),
	}
}

// GetName implements the Type interface for GenericType
func (g *GenericType) GetName() string {
	if !g.IsGeneric() {
		return g.Name
	}

	params := make([]string, len(g.TypeParameters))
	for i, param := range g.TypeParameters {
		params[i] = param.Name
	}

	return fmt.Sprintf("%s<%s>", g.Name, strings.Join(params, ", "))
}

// IsCompatible implements the Type interface for GenericType
func (g *GenericType) IsCompatible(other Type) bool {
	if other == nil {
		return false
	}

	// If other is also generic, check structural compatibility
	if otherGeneric, ok := other.(*GenericType); ok {
		return g.isStructurallyCompatible(otherGeneric)
	}

	// If other is concrete, check if it can be an instantiation of this generic
	return g.canInstantiate(other)
}

// IsGeneric returns true if the type is generic
func (g *GenericType) IsGeneric() bool {
	return len(g.TypeParameters) > 0
}

// isCompatibleOld checks if this generic type is compatible with another type (old implementation)
func (g *GenericType) isCompatibleOld(other Type) bool {
	if other == nil {
		return false
	}

	// If other is also generic, check structural compatibility
	if otherGeneric, ok := other.(*GenericType); ok {
		return g.isStructurallyCompatible(otherGeneric)
	}

	// If other is concrete, check if it can be an instantiation of this generic
	return g.canInstantiate(other)
}

// isStructurallyCompatible checks if two generic types have the same structure
func (g *GenericType) isStructurallyCompatible(other *GenericType) bool {
	// Same name and same number of type parameters
	if g.Name != other.Name || len(g.TypeParameters) != len(other.TypeParameters) {
		return false
	}

	// Check type parameter bounds compatibility
	for i, param := range g.TypeParameters {
		otherParam := other.TypeParameters[i]
		if !g.areParameterBoundsCompatible(param, otherParam) {
			return false
		}
	}

	return true
}

// areParameterBoundsCompatible checks if type parameter bounds are compatible
func (g *GenericType) areParameterBoundsCompatible(param1, param2 TypeParameter) bool {
	// For now, we require exact match of bounds
	// This could be relaxed to allow contravariant/covariant relationships
	if len(param1.Bounds) != len(param2.Bounds) {
		return false
	}

	for i, bound1 := range param1.Bounds {
		bound2 := param2.Bounds[i]
		if !bound1.IsCompatible(bound2) {
			return false
		}
	}

	return true
}

// canInstantiate checks if a concrete type can be an instantiation of this generic
func (g *GenericType) canInstantiate(concrete Type) bool {
	// This is a simplified check - in a full implementation,
	// we would need to perform type inference and constraint checking
	return g.BaseType != nil && g.BaseType.IsCompatible(concrete)
}

// InstantiateWith creates a concrete type by substituting type parameters
func (g *GenericType) InstantiateWith(typeArgs []Type) (Type, error) {
	if len(typeArgs) != len(g.TypeParameters) {
		return nil, fmt.Errorf("wrong number of type arguments: expected %d, got %d",
			len(g.TypeParameters), len(typeArgs))
	}

	// Create type substitution map
	substitutions := make(map[string]Type)
	for i, param := range g.TypeParameters {
		substitutions[param.Name] = typeArgs[i]
	}

	// Validate constraints
	checker := NewGenericTypeChecker()
	for _, constraint := range g.Constraints {
		if err := checker.ValidateConstraint(constraint, substitutions); err != nil {
			return nil, fmt.Errorf("constraint violation: %v", err)
		}
	}

	// Perform type substitution
	return g.substitute(g.BaseType, substitutions), nil
}

// substitute performs type parameter substitution
func (g *GenericType) substitute(t Type, substitutions map[string]Type) Type {
	if t == nil {
		return nil
	}

	// If this is a type parameter, substitute it
	if simple, ok := t.(*SimpleType); ok {
		if replacement, exists := substitutions[simple.Name]; exists {
			return replacement
		}
		return t
	}

	// For parameterized types, recursively substitute
	if param, ok := t.(*ParameterizedType); ok {
		newArgs := make([]Type, len(param.Arguments))
		for i, arg := range param.Arguments {
			newArgs[i] = g.substitute(arg, substitutions)
		}
		return &ParameterizedType{
			BaseType:  param.BaseType,
			Arguments: newArgs,
		}
	}

	// For union types, we would need to substitute all members
	// This requires the union type to support Type interface members
	// For now, return the original type
	if _, ok := t.(*UnionType); ok {
		// Union types with Type members are not yet supported in substitution
		return t
	}

	return t
}

// ValidateConstraint validates a type constraint against type substitutions
func (checker *GenericTypeChecker) ValidateConstraint(constraint TypeConstraint, substitutions map[string]Type) error {
	cacheKey := fmt.Sprintf("%s:%d:%s", constraint.Parameter, constraint.Kind, constraint.Expression.GetName())

	// Check cache first
	if result, exists := checker.constraintCache[cacheKey]; exists {
		if !result {
			return fmt.Errorf("constraint %s failed", cacheKey)
		}
		return nil
	}

	// Get the actual type for the parameter
	actualType, exists := substitutions[constraint.Parameter]
	if !exists {
		return fmt.Errorf("no type substitution for parameter %s", constraint.Parameter)
	}

	var result bool
	switch constraint.Kind {
	case TraitConstraint:
		result = checker.validateTraitConstraint(actualType, constraint.Expression)
	case ProtocolConstraint:
		result = checker.validateProtocolConstraint(actualType, constraint.Expression)
	case CapabilityConstraint:
		result = checker.validateCapabilityConstraint(actualType, constraint.Expression)
	case ValueConstraint:
		result = checker.validateValueConstraint(actualType, constraint.Expression)
	default:
		return fmt.Errorf("unknown constraint kind: %d", constraint.Kind)
	}

	// Cache the result
	checker.constraintCache[cacheKey] = result

	if !result {
		return fmt.Errorf("constraint %s failed for type %s", cacheKey, actualType.GetName())
	}

	return nil
}

// validateTraitConstraint validates that a type implements a trait
func (checker *GenericTypeChecker) validateTraitConstraint(actualType, traitType Type) bool {
	// For now, use simple name matching
	// In a full implementation, this would check method signatures, etc.
	return actualType.IsCompatible(traitType)
}

// validateProtocolConstraint validates that a type implements a protocol
func (checker *GenericTypeChecker) validateProtocolConstraint(actualType, protocolType Type) bool {
	// Protocol constraints are similar to trait constraints but with different semantics
	return actualType.IsCompatible(protocolType)
}

// validateCapabilityConstraint validates that a type has a specific capability
func (checker *GenericTypeChecker) validateCapabilityConstraint(actualType, capabilityType Type) bool {
	// Capability constraints check if a type can perform certain operations
	// This is a simplified implementation
	return true // For now, assume all types have all capabilities
}

// validateValueConstraint validates value-based constraints
func (checker *GenericTypeChecker) validateValueConstraint(actualType, constraintExpr Type) bool {
	// Value constraints are runtime checks, so we can't validate them at compile time
	// For now, we assume they're always valid
	return true
}

// InferTypeArguments attempts to infer type arguments from usage context
func (checker *GenericTypeChecker) InferTypeArguments(generic *GenericType, usage Type) ([]Type, error) {
	// Type inference is complex and would require analysis of the usage context
	// For now, we return an error indicating inference is not implemented
	return nil, fmt.Errorf("type inference not yet implemented")
}

// GetBuiltinConstraints returns the built-in constraint types
func GetBuiltinConstraints() map[string]Type {
	return map[string]Type{
		"Serializable":   &SimpleType{Name: "Serializable"},
		"Deserializable": &SimpleType{Name: "Deserializable"},
		"Defined":        &SimpleType{Name: "Defined"},
		"Clonable":       &SimpleType{Name: "Clonable"},
		"Display":        &SimpleType{Name: "Display"},
		"Clone":          &SimpleType{Name: "Clone"},
		"Any":            &SimpleType{Name: "Any"},
		"Cacheable":      &SimpleType{Name: "Cacheable"},
	}
}

// String returns a string representation of the constraint kind
func (ck ConstraintKind) String() string {
	switch ck {
	case TraitConstraint:
		return "trait"
	case ProtocolConstraint:
		return "protocol"
	case CapabilityConstraint:
		return "capability"
	case ValueConstraint:
		return "value"
	default:
		return "unknown"
	}
}
