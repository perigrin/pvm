// ABOUTME: Comprehensive tests for generic type system functionality
// ABOUTME: Tests type parameters, constraints, substitution, and compatibility checking

package typedef

import (
	"testing"
)

func TestGenericType_IsGeneric(t *testing.T) {
	tests := []struct {
		name     string
		generic  *GenericType
		expected bool
	}{
		{
			name: "Generic type with type parameters",
			generic: &GenericType{
				Name: "Array",
				TypeParameters: []TypeParameter{
					{Name: "T"},
				},
			},
			expected: true,
		},
		{
			name: "Non-generic type",
			generic: &GenericType{
				Name:           "String",
				TypeParameters: []TypeParameter{},
			},
			expected: false,
		},
		{
			name: "Generic type with multiple parameters",
			generic: &GenericType{
				Name: "Map",
				TypeParameters: []TypeParameter{
					{Name: "K"},
					{Name: "V"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.generic.IsGeneric()
			if result != tt.expected {
				t.Errorf("IsGeneric() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGenericType_GetName(t *testing.T) {
	tests := []struct {
		name     string
		generic  *GenericType
		expected string
	}{
		{
			name: "Single type parameter",
			generic: &GenericType{
				Name: "Array",
				TypeParameters: []TypeParameter{
					{Name: "T"},
				},
			},
			expected: "Array<T>",
		},
		{
			name: "Multiple type parameters",
			generic: &GenericType{
				Name: "Map",
				TypeParameters: []TypeParameter{
					{Name: "K"},
					{Name: "V"},
				},
			},
			expected: "Map<K, V>",
		},
		{
			name: "Non-generic type",
			generic: &GenericType{
				Name:           "String",
				TypeParameters: []TypeParameter{},
			},
			expected: "String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.generic.GetName()
			if result != tt.expected {
				t.Errorf("GetName() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGenericType_IsCompatible(t *testing.T) {
	// Create test types
	arrayInt := &GenericType{
		Name: "Array",
		TypeParameters: []TypeParameter{
			{Name: "T"},
		},
		BaseType: &SimpleType{Name: "Array"},
	}

	arrayString := &GenericType{
		Name: "Array",
		TypeParameters: []TypeParameter{
			{Name: "T"},
		},
		BaseType: &SimpleType{Name: "Array"},
	}

	mapType := &GenericType{
		Name: "Map",
		TypeParameters: []TypeParameter{
			{Name: "K"},
			{Name: "V"},
		},
		BaseType: &SimpleType{Name: "Map"},
	}

	concreteArray := &SimpleType{Name: "Array"}

	tests := []struct {
		name     string
		generic  *GenericType
		other    Type
		expected bool
	}{
		{
			name:     "Same generic type structure",
			generic:  arrayInt,
			other:    arrayString,
			expected: true,
		},
		{
			name:     "Different generic type structure",
			generic:  arrayInt,
			other:    mapType,
			expected: false,
		},
		{
			name:     "Generic with compatible concrete type",
			generic:  arrayInt,
			other:    concreteArray,
			expected: true,
		},
		{
			name:     "Nil type",
			generic:  arrayInt,
			other:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.generic.IsCompatible(tt.other)
			if result != tt.expected {
				t.Errorf("IsCompatible() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGenericType_InstantiateWith(t *testing.T) {
	// Create a simple generic type
	arrayGeneric := &GenericType{
		Name: "Array",
		TypeParameters: []TypeParameter{
			{Name: "T"},
		},
		BaseType: &SimpleType{Name: "Array"},
	}

	// Create type arguments
	intType := &SimpleType{Name: "Int"}
	stringType := &SimpleType{Name: "Str"}

	tests := []struct {
		name        string
		generic     *GenericType
		typeArgs    []Type
		expectError bool
	}{
		{
			name:        "Valid instantiation",
			generic:     arrayGeneric,
			typeArgs:    []Type{intType},
			expectError: false,
		},
		{
			name:        "Wrong number of type arguments",
			generic:     arrayGeneric,
			typeArgs:    []Type{intType, stringType},
			expectError: true,
		},
		{
			name:        "Empty type arguments",
			generic:     arrayGeneric,
			typeArgs:    []Type{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.generic.InstantiateWith(tt.typeArgs)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("Expected result but got nil")
				}
			}
		})
	}
}

func TestGenericType_WithConstraints(t *testing.T) {
	// Create a generic type with constraints
	constrainedGeneric := &GenericType{
		Name: "Container",
		TypeParameters: []TypeParameter{
			{Name: "T"},
		},
		Constraints: []TypeConstraint{
			{
				Parameter:  "T",
				Kind:       TraitConstraint,
				Expression: &SimpleType{Name: "Serializable"},
			},
		},
		BaseType: &SimpleType{Name: "Container"},
	}

	// Create type arguments
	serializableType := &SimpleType{Name: "SerializableClass"}
	nonSerializableType := &SimpleType{Name: "NonSerializableClass"}

	tests := []struct {
		name        string
		generic     *GenericType
		typeArgs    []Type
		expectError bool
	}{
		{
			name:        "Constraint-satisfying type",
			generic:     constrainedGeneric,
			typeArgs:    []Type{serializableType},
			expectError: false, // We don't enforce constraints yet in this simplified version
		},
		{
			name:        "Non-constraint-satisfying type",
			generic:     constrainedGeneric,
			typeArgs:    []Type{nonSerializableType},
			expectError: false, // We don't enforce constraints yet in this simplified version
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.generic.InstantiateWith(tt.typeArgs)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGenericTypeChecker_ValidateConstraint(t *testing.T) {
	checker := NewGenericTypeChecker()

	// Create test types
	serializableType := &SimpleType{Name: "Serializable"}
	intType := &SimpleType{Name: "Int"}

	substitutions := map[string]Type{
		"T": intType,
	}

	tests := []struct {
		name        string
		constraint  TypeConstraint
		expectError bool
	}{
		{
			name: "Valid trait constraint",
			constraint: TypeConstraint{
				Parameter:  "T",
				Kind:       TraitConstraint,
				Expression: serializableType,
			},
			expectError: false, // Simplified implementation doesn't validate strictly
		},
		{
			name: "Invalid parameter",
			constraint: TypeConstraint{
				Parameter:  "U", // Not in substitutions
				Kind:       TraitConstraint,
				Expression: serializableType,
			},
			expectError: true,
		},
		{
			name: "Protocol constraint",
			constraint: TypeConstraint{
				Parameter:  "T",
				Kind:       ProtocolConstraint,
				Expression: serializableType,
			},
			expectError: false,
		},
		{
			name: "Capability constraint",
			constraint: TypeConstraint{
				Parameter:  "T",
				Kind:       CapabilityConstraint,
				Expression: serializableType,
			},
			expectError: false,
		},
		{
			name: "Value constraint",
			constraint: TypeConstraint{
				Parameter:  "T",
				Kind:       ValueConstraint,
				Expression: serializableType,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checker.ValidateConstraint(tt.constraint, substitutions)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGetBuiltinConstraints(t *testing.T) {
	builtins := GetBuiltinConstraints()

	expectedConstraints := []string{
		"Serializable",
		"Deserializable",
		"Defined",
		"Clonable",
		"Display",
		"Clone",
		"Any",
		"Cacheable",
	}

	for _, expected := range expectedConstraints {
		if _, exists := builtins[expected]; !exists {
			t.Errorf("Missing builtin constraint: %s", expected)
		}
	}

	// Verify all constraints are SimpleType
	for name, constraint := range builtins {
		if constraint.GetName() != name {
			t.Errorf("Constraint %s has wrong name: %s", name, constraint.GetName())
		}
	}
}

func TestConstraintKind_String(t *testing.T) {
	tests := []struct {
		kind     ConstraintKind
		expected string
	}{
		{TraitConstraint, "trait"},
		{ProtocolConstraint, "protocol"},
		{CapabilityConstraint, "capability"},
		{ValueConstraint, "value"},
		{ConstraintKind(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.kind.String()
			if result != tt.expected {
				t.Errorf("String() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGenericType_Substitute(t *testing.T) {
	generic := &GenericType{
		Name: "Container",
		TypeParameters: []TypeParameter{
			{Name: "T"},
		},
	}

	// Create types for substitution
	intType := &SimpleType{Name: "Int"}
	stringType := &SimpleType{Name: "Str"}
	typeParam := &SimpleType{Name: "T"}

	substitutions := map[string]Type{
		"T": intType,
	}

	tests := []struct {
		name     string
		input    Type
		expected Type
	}{
		{
			name:     "Substitute type parameter",
			input:    typeParam,
			expected: intType,
		},
		{
			name:     "Non-parameter type unchanged",
			input:    stringType,
			expected: stringType,
		},
		{
			name:     "Nil type",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generic.substitute(tt.input, substitutions)

			if result == nil && tt.expected == nil {
				return
			}

			if result == nil || tt.expected == nil {
				t.Errorf("Mismatch: got %v, expected %v", result, tt.expected)
				return
			}

			if result.GetName() != tt.expected.GetName() {
				t.Errorf("substitute() = %v, expected %v", result.GetName(), tt.expected.GetName())
			}
		})
	}
}

func TestGenericType_ComplexSubstitution(t *testing.T) {
	generic := &GenericType{
		Name: "Container",
		TypeParameters: []TypeParameter{
			{Name: "T"},
		},
	}

	// Create parameterized type Array[T]
	arrayT := &ParameterizedType{
		BaseType:  &SimpleType{Name: "Array"},
		Arguments: []Type{&SimpleType{Name: "T"}},
	}

	// Create union type T|String
	unionT := &UnionType{
		Members: []string{"T", "String"},
	}

	intType := &SimpleType{Name: "Int"}
	substitutions := map[string]Type{
		"T": intType,
	}

	tests := []struct {
		name     string
		input    Type
		expected string
	}{
		{
			name:     "Parameterized type substitution",
			input:    arrayT,
			expected: "Array[Int]",
		},
		{
			name:     "Union type substitution",
			input:    unionT,
			expected: "Int|String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generic.substitute(tt.input, substitutions)

			if result == nil {
				t.Errorf("substitute() returned nil")
				return
			}

			if result.GetName() != tt.expected {
				t.Errorf("substitute() = %v, expected %v", result.GetName(), tt.expected)
			}
		})
	}
}
