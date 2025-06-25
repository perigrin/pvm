// ABOUTME: Complex type implementations including unions, intersections, and enhanced parameterized types
// ABOUTME: Provides sophisticated type representations for advanced Perl type scenarios

package types

import (
	"fmt"
	"sort"
	"strings"
)

// UnionType represents a type that can be one of several possible types (Int|Str)
type UnionType struct {
	types []Type
}

// NewUnionType creates a new union type from the given types
func NewUnionType(types ...Type) Type {
	if len(types) == 0 {
		panic("Union type must have at least one type")
	}
	if len(types) == 1 {
		return types[0] // Single type union is just the type itself
	}

	// Flatten nested unions and remove duplicates
	var flatTypes []Type
	for _, t := range types {
		if union, ok := t.(*UnionType); ok {
			// Flatten nested union
			flatTypes = append(flatTypes, union.types...)
		} else {
			flatTypes = append(flatTypes, t)
		}
	}

	// Remove duplicates
	uniqueTypes := make([]Type, 0, len(flatTypes))
	seen := make(map[string]bool)
	for _, t := range flatTypes {
		typeStr := t.String()
		if !seen[typeStr] {
			seen[typeStr] = true
			uniqueTypes = append(uniqueTypes, t)
		}
	}

	// Sort types for consistent ordering
	sort.Slice(uniqueTypes, func(i, j int) bool {
		return uniqueTypes[i].String() < uniqueTypes[j].String()
	})

	return &UnionType{types: uniqueTypes}
}

// String returns the string representation of the union type
func (u *UnionType) String() string {
	var parts []string
	for _, t := range u.types {
		parts = append(parts, t.String())
	}
	return strings.Join(parts, "|")
}

// Equals checks if this union type equals another type
func (u *UnionType) Equals(other Type) bool {
	if otherUnion, ok := other.(*UnionType); ok {
		if len(u.types) != len(otherUnion.types) {
			return false
		}
		// Since types are sorted, we can compare directly
		for i, t := range u.types {
			if !t.Equals(otherUnion.types[i]) {
				return false
			}
		}
		return true
	}
	return false
}

// CompatibleWith checks if this union type is compatible with another type
func (u *UnionType) CompatibleWith(other Type) bool {
	// A union is compatible with another union if all of our types are contained in the other union
	if otherUnion, ok := other.(*UnionType); ok {
		// All of our types must be contained in the other union
		for _, ourType := range u.types {
			if !otherUnion.ContainsType(ourType) {
				return false
			}
		}
		return true
	}
	
	// A union is compatible with a single type if any of its types are compatible
	for _, t := range u.types {
		if t.CompatibleWith(other) {
			return true
		}
	}
	return false
}

// ContainsType checks if the union contains a specific type
func (u *UnionType) ContainsType(t Type) bool {
	for _, unionType := range u.types {
		if unionType.Equals(t) {
			return true
		}
	}
	return false
}

// IsBasic returns false for union types
func (u *UnionType) IsBasic() bool {
	return false
}

// IsComplex returns true for union types
func (u *UnionType) IsComplex() bool {
	return true
}

// IntersectionType represents a type that must satisfy all of several requirements (Object&Serializable)
type IntersectionType struct {
	types []Type
}

// NewIntersectionType creates a new intersection type from the given types
func NewIntersectionType(types ...Type) Type {
	if len(types) == 0 {
		panic("Intersection type must have at least one type")
	}
	if len(types) == 1 {
		return types[0] // Single type intersection is just the type itself
	}

	// Flatten nested intersections and remove duplicates
	var flatTypes []Type
	for _, t := range types {
		if intersection, ok := t.(*IntersectionType); ok {
			// Flatten nested intersection
			flatTypes = append(flatTypes, intersection.types...)
		} else {
			flatTypes = append(flatTypes, t)
		}
	}

	// Remove duplicates
	uniqueTypes := make([]Type, 0, len(flatTypes))
	seen := make(map[string]bool)
	for _, t := range flatTypes {
		typeStr := t.String()
		if !seen[typeStr] {
			seen[typeStr] = true
			uniqueTypes = append(uniqueTypes, t)
		}
	}

	// Sort types for consistent ordering
	sort.Slice(uniqueTypes, func(i, j int) bool {
		return uniqueTypes[i].String() < uniqueTypes[j].String()
	})

	return &IntersectionType{types: uniqueTypes}
}

// String returns the string representation of the intersection type
func (i *IntersectionType) String() string {
	var parts []string
	for _, t := range i.types {
		parts = append(parts, t.String())
	}
	return strings.Join(parts, "&")
}

// Equals checks if this intersection type equals another type
func (i *IntersectionType) Equals(other Type) bool {
	if otherIntersection, ok := other.(*IntersectionType); ok {
		if len(i.types) != len(otherIntersection.types) {
			return false
		}
		// Since types are sorted, we can compare directly
		for idx, t := range i.types {
			if !t.Equals(otherIntersection.types[idx]) {
				return false
			}
		}
		return true
	}
	return false
}

// CompatibleWith checks if this intersection type is compatible with another type
func (i *IntersectionType) CompatibleWith(other Type) bool {
	// An intersection is compatible with another type if that type satisfies all requirements
	// For basic implementation, we're conservative - the other type must be an intersection
	// that contains all our types or more
	if otherIntersection, ok := other.(*IntersectionType); ok {
		// All our types must be present in the other intersection
		for _, ourType := range i.types {
			found := false
			for _, otherType := range otherIntersection.types {
				if ourType.Equals(otherType) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
	return false
}

// IsBasic returns false for intersection types
func (i *IntersectionType) IsBasic() bool {
	return false
}

// IsComplex returns true for intersection types
func (i *IntersectionType) IsComplex() bool {
	return true
}

// MultiParameterizedType represents a type with multiple type parameters (Map[Str, Int])
type MultiParameterizedType struct {
	name       string
	parameters []Type
}

// NewMapType creates a new Map type with key and value types
func NewMapType(keyType, valueType Type) Type {
	return &MultiParameterizedType{
		name:       "Map",
		parameters: []Type{keyType, valueType},
	}
}

// String returns the string representation of the multi-parameterized type
func (m *MultiParameterizedType) String() string {
	var params []string
	for _, param := range m.parameters {
		params = append(params, param.String())
	}
	return fmt.Sprintf("%s[%s]", m.name, strings.Join(params, ", "))
}

// Equals checks if this multi-parameterized type equals another type
func (m *MultiParameterizedType) Equals(other Type) bool {
	if otherMulti, ok := other.(*MultiParameterizedType); ok {
		if m.name != otherMulti.name || len(m.parameters) != len(otherMulti.parameters) {
			return false
		}
		for i, param := range m.parameters {
			if !param.Equals(otherMulti.parameters[i]) {
				return false
			}
		}
		return true
	}
	return false
}

// CompatibleWith checks if this multi-parameterized type is compatible with another type
func (m *MultiParameterizedType) CompatibleWith(other Type) bool {
	if otherMulti, ok := other.(*MultiParameterizedType); ok {
		if m.name != otherMulti.name || len(m.parameters) != len(otherMulti.parameters) {
			return false
		}
		for i, param := range m.parameters {
			if !param.CompatibleWith(otherMulti.parameters[i]) {
				return false
			}
		}
		return true
	}
	return false
}

// IsBasic returns false for multi-parameterized types
func (m *MultiParameterizedType) IsBasic() bool {
	return false
}

// IsComplex returns true for multi-parameterized types
func (m *MultiParameterizedType) IsComplex() bool {
	return true
}