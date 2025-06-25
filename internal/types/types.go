// ABOUTME: Core type interfaces and basic implementations for type inference
// ABOUTME: Defines the foundational type system used throughout type inference and compilation

package types

import "fmt"

// Type represents a type in the type system
type Type interface {
	// String returns the string representation of the type
	String() string

	// Equals checks if this type is equal to another type
	Equals(other Type) bool

	// CompatibleWith checks if this type is compatible with another type
	CompatibleWith(other Type) bool

	// IsBasic returns true if this is a basic type (Int, Str, Bool)
	IsBasic() bool

	// IsComplex returns true if this is a complex type (ArrayRef, HashRef, etc.)
	IsComplex() bool
}

// BasicType represents a basic scalar type
type BasicType struct {
	name string
}

// NewIntType creates a new Int type
func NewIntType() Type {
	return &BasicType{name: "Int"}
}

// NewStrType creates a new Str type
func NewStrType() Type {
	return &BasicType{name: "Str"}
}

// NewBoolType creates a new Bool type
func NewBoolType() Type {
	return &BasicType{name: "Bool"}
}

// NewNumType creates a new Num type for floating-point numbers
func NewNumType() Type {
	return &BasicType{name: "Num"}
}

// String returns the string representation of the basic type
func (b *BasicType) String() string {
	return b.name
}

// Equals checks if this basic type equals another type
func (b *BasicType) Equals(other Type) bool {
	if otherBasic, ok := other.(*BasicType); ok {
		return b.name == otherBasic.name
	}
	return false
}

// CompatibleWith checks if this basic type is compatible with another type
func (b *BasicType) CompatibleWith(other Type) bool {
	return b.Equals(other)
}

// IsBasic returns true for basic types
func (b *BasicType) IsBasic() bool {
	return true
}

// IsComplex returns false for basic types
func (b *BasicType) IsComplex() bool {
	return false
}

// ParameterizedType represents a type with type parameters (e.g., ArrayRef[Int])
type ParameterizedType struct {
	name      string
	parameter Type
}

// NewArrayRefType creates a new ArrayRef type with the given element type
func NewArrayRefType(elementType Type) Type {
	return &ParameterizedType{
		name:      "ArrayRef",
		parameter: elementType,
	}
}

// NewHashRefType creates a new HashRef type with the given value type
func NewHashRefType(valueType Type) Type {
	return &ParameterizedType{
		name:      "HashRef",
		parameter: valueType,
	}
}

// String returns the string representation of the parameterized type
func (p *ParameterizedType) String() string {
	return fmt.Sprintf("%s[%s]", p.name, p.parameter.String())
}

// Equals checks if this parameterized type equals another type
func (p *ParameterizedType) Equals(other Type) bool {
	if otherParam, ok := other.(*ParameterizedType); ok {
		return p.name == otherParam.name && p.parameter.Equals(otherParam.parameter)
	}
	return false
}

// CompatibleWith checks if this parameterized type is compatible with another type
func (p *ParameterizedType) CompatibleWith(other Type) bool {
	return p.Equals(other)
}

// IsBasic returns false for parameterized types
func (p *ParameterizedType) IsBasic() bool {
	return false
}

// IsComplex returns true for parameterized types
func (p *ParameterizedType) IsComplex() bool {
	return true
}
