// ABOUTME: Type hierarchy and type checking capabilities
// ABOUTME: Implements type relationships and compatibility checking

package typedef

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/errors"
)

// PSC Error codes
const (
	ErrTypeCheckFailed     = "704" // Type checking failed
	ErrTypeIncompatible    = "705" // Incompatible types
	ErrTypeUndefined       = "706" // Undefined type referenced
	ErrTypeInvalid         = "707" // Invalid type expression
	ErrTypeAssignmentError = "708" // Type assignment error
)

// TypeHierarchy represents the hierarchical relationship between types
// It includes built-in types and their relationships
type TypeHierarchy struct {
	// BuiltinTypes maps type names to TypeInfo
	BuiltinTypes map[string]*TypeInfo

	// TypeStore provides access to stored type definitions
	TypeStore *Storage

	// Subtype relationships (child -> parent)
	subtypeRelations map[string][]string

	// Parameterized type constructors
	parameterizedTypes map[string]func([]string) (string, error)
}

// NewTypeHierarchy creates a new TypeHierarchy with built-in types
func NewTypeHierarchy(store *Storage) *TypeHierarchy {
	hierarchy := &TypeHierarchy{
		BuiltinTypes:      make(map[string]*TypeInfo),
		TypeStore:         store,
		subtypeRelations:  make(map[string][]string),
		parameterizedTypes: make(map[string]func([]string) (string, error)),
	}

	// Initialize built-in types
	hierarchy.initializeBuiltinTypes()

	return hierarchy
}

// initializeBuiltinTypes sets up the initial type hierarchy for built-in types
func (h *TypeHierarchy) initializeBuiltinTypes() {
	// Create basic type hierarchy
	// Any is the root type
	h.addBuiltinType("Any", "Top type - all types are subtypes of Any", "scalar")

	// Basic scalar types
	h.addBuiltinType("Scalar", "Basic scalar value", "scalar")
	h.addBuiltinType("Str", "String value", "scalar")
	h.addBuiltinType("Num", "Numeric value", "scalar")
	h.addBuiltinType("Int", "Integer value", "scalar")
	h.addBuiltinType("Float", "Floating point value", "scalar")
	h.addBuiltinType("Bool", "Boolean value", "scalar")
	h.addBuiltinType("Undef", "Undefined value", "scalar")

	// Reference types
	h.addBuiltinType("Ref", "Reference", "ref")
	h.addBuiltinType("ScalarRef", "Scalar reference", "ref")
	h.addBuiltinType("ArrayRef", "Array reference", "ref")
	h.addBuiltinType("HashRef", "Hash reference", "ref")
	h.addBuiltinType("CodeRef", "Code reference", "ref")
	h.addBuiltinType("RegexpRef", "Regular expression reference", "ref")
	h.addBuiltinType("GlobRef", "Glob reference", "ref")
	h.addBuiltinType("FileHandle", "File handle", "ref")

	// Container types
	h.addBuiltinType("List", "List type", "container")
	h.addBuiltinType("Array", "Array type", "container")
	h.addBuiltinType("Hash", "Hash type", "container")
	h.addBuiltinType("Code", "Code type", "container")
	h.addBuiltinType("Glob", "Glob type", "container")

	// Type modifiers
	h.addBuiltinType("Maybe", "Optional type that can be undef", "modifier")
	h.addBuiltinType("Optional", "Optional type like Maybe but specific to key existence", "modifier")

	// Role/trait types
	h.addBuiltinType("Callable", "Can be called like a function", "role")
	h.addBuiltinType("Iterable", "Can be iterated over", "role")
	h.addBuiltinType("Positional", "Has indexed elements", "role")
	h.addBuiltinType("Associative", "Has key-value pairs", "role")

	// IO and system types
	h.addBuiltinType("IO", "Input/output type", "io")
	h.addBuiltinType("Path", "Filesystem path", "io")
	h.addBuiltinType("File", "File type", "io")
	h.addBuiltinType("Dir", "Directory type", "io")

	// Additional scalar types
	h.addBuiltinType("ClassName", "Class name string", "scalar")
	h.addBuiltinType("RoleName", "Role name string", "scalar")
	h.addBuiltinType("MethodName", "Method name string", "scalar")
	h.addBuiltinType("Byte", "Byte value", "scalar")
	h.addBuiltinType("Char", "Character value", "scalar")
	h.addBuiltinType("VarName", "Variable name", "scalar")

	// Set up subtype relationships
	h.addSubtypeRelation("Scalar", "Any")
	h.addSubtypeRelation("Ref", "Any")
	h.addSubtypeRelation("List", "Any")
	h.addSubtypeRelation("Code", "Any")
	h.addSubtypeRelation("Glob", "Any")
	h.addSubtypeRelation("IO", "Any")
	
	h.addSubtypeRelation("Str", "Scalar")
	h.addSubtypeRelation("Num", "Scalar")
	h.addSubtypeRelation("Bool", "Scalar")
	h.addSubtypeRelation("Undef", "Scalar")
	
	h.addSubtypeRelation("Int", "Num")
	h.addSubtypeRelation("Float", "Num")
	
	h.addSubtypeRelation("ClassName", "Str")
	h.addSubtypeRelation("RoleName", "Str")
	h.addSubtypeRelation("MethodName", "Str")
	h.addSubtypeRelation("Byte", "Str")
	h.addSubtypeRelation("Char", "Str")
	h.addSubtypeRelation("VarName", "Str")
	
	h.addSubtypeRelation("ScalarRef", "Ref")
	h.addSubtypeRelation("ArrayRef", "Ref")
	h.addSubtypeRelation("HashRef", "Ref")
	h.addSubtypeRelation("CodeRef", "Ref")
	h.addSubtypeRelation("RegexpRef", "Ref")
	h.addSubtypeRelation("GlobRef", "Ref")
	h.addSubtypeRelation("FileHandle", "Ref")
	
	h.addSubtypeRelation("Array", "List")
	h.addSubtypeRelation("Hash", "Associative")
	h.addSubtypeRelation("Array", "Positional")
	h.addSubtypeRelation("Hash", "Iterable")
	h.addSubtypeRelation("Array", "Iterable")
	h.addSubtypeRelation("Code", "Callable")
	h.addSubtypeRelation("CodeRef", "Callable")
	
	h.addSubtypeRelation("File", "IO")
	h.addSubtypeRelation("Dir", "IO")
	h.addSubtypeRelation("Path", "Str")

	// Set up parameterized type constructors
	h.parameterizedTypes["ArrayRef"] = func(params []string) (string, error) {
		if len(params) != 1 {
			return "", fmt.Errorf("ArrayRef requires exactly one type parameter")
		}
		return fmt.Sprintf("ArrayRef[%s]", params[0]), nil
	}

	h.parameterizedTypes["HashRef"] = func(params []string) (string, error) {
		if len(params) == 1 {
			return fmt.Sprintf("HashRef[%s]", params[0]), nil
		} else if len(params) == 2 {
			return fmt.Sprintf("HashRef[%s,%s]", params[0], params[1]), nil
		}
		return "", fmt.Errorf("HashRef requires one or two type parameters")
	}

	h.parameterizedTypes["Maybe"] = func(params []string) (string, error) {
		if len(params) != 1 {
			return "", fmt.Errorf("Maybe requires exactly one type parameter")
		}
		return fmt.Sprintf("Maybe[%s]", params[0]), nil
	}

	h.parameterizedTypes["Optional"] = func(params []string) (string, error) {
		if len(params) != 1 {
			return "", fmt.Errorf("Optional requires exactly one type parameter")
		}
		return fmt.Sprintf("Optional[%s]", params[0]), nil
	}
}

// addBuiltinType adds a built-in type to the hierarchy
func (h *TypeHierarchy) addBuiltinType(name, description, kind string) {
	h.BuiltinTypes[name] = &TypeInfo{
		Name:        name,
		Description: description,
		Kind:        kind,
	}
}

// addSubtypeRelation adds a subtype relationship (child is a subtype of parent)
func (h *TypeHierarchy) addSubtypeRelation(child, parent string) {
	h.subtypeRelations[child] = append(h.subtypeRelations[child], parent)
}

// IsBuiltinType checks if a type is a built-in type
func (h *TypeHierarchy) IsBuiltinType(typeName string) bool {
	// Check for parameterized types like ArrayRef[Int]
	if idx := strings.Index(typeName, "["); idx > 0 {
		baseType := typeName[:idx]
		_, isParamType := h.parameterizedTypes[baseType]
		return isParamType
	}
	
	_, exists := h.BuiltinTypes[typeName]
	return exists
}

// IsSubtypeOf checks if childType is a subtype of parentType
func (h *TypeHierarchy) IsSubtypeOf(childType, parentType string) bool {
	// Same type is always compatible
	if childType == parentType {
		return true
	}

	// Special case: Any is the top type
	if parentType == "Any" {
		return true
	}

	// Handle parameterized types
	if strings.Contains(childType, "[") && strings.Contains(parentType, "[") {
		return h.checkParameterizedSubtype(childType, parentType)
	} else if strings.Contains(childType, "[") {
		// Parameterized child type, non-parameterized parent type
		// Extract base type of child
		baseChild := childType
		if idx := strings.Index(childType, "["); idx > 0 {
			baseChild = childType[:idx]
		}
		return h.isSubtypeOfBase(baseChild, parentType)
	}

	// Regular subtype check
	return h.isSubtypeOfBase(childType, parentType)
}

// isSubtypeOfBase checks if a base type is a subtype of another base type
func (h *TypeHierarchy) isSubtypeOfBase(childType, parentType string) bool {
	// Same type
	if childType == parentType {
		return true
	}

	// Check direct relationships
	parents, ok := h.subtypeRelations[childType]
	if !ok {
		return false
	}

	// Check if parent is directly in the list
	for _, parent := range parents {
		if parent == parentType {
			return true
		}
		// Recursively check parents of parents
		if h.isSubtypeOfBase(parent, parentType) {
			return true
		}
	}

	return false
}

// checkParameterizedSubtype checks if a parameterized type is a subtype of another
func (h *TypeHierarchy) checkParameterizedSubtype(childType, parentType string) bool {
	// Extract base types and parameters
	childBase, childParams := extractTypeAndParams(childType)
	parentBase, parentParams := extractTypeAndParams(parentType)

	// Base types must be compatible
	if !h.isSubtypeOfBase(childBase, parentBase) {
		return false
	}

	// If parent base type is compatible but doesn't have parameters,
	// then we don't need to check parameters
	if len(parentParams) == 0 {
		return true
	}

	// Parameter count must match
	if len(childParams) != len(parentParams) {
		return false
	}

	// Check each parameter
	for i, childParam := range childParams {
		if !h.IsSubtypeOf(childParam, parentParams[i]) {
			return false
		}
	}

	return true
}

// extractTypeAndParams extracts the base type and parameters from a parameterized type
// e.g., "ArrayRef[Int]" -> "ArrayRef", ["Int"]
func extractTypeAndParams(paramType string) (string, []string) {
	idx := strings.Index(paramType, "[")
	if idx < 0 {
		return paramType, nil
	}

	baseType := paramType[:idx]
	paramStr := paramType[idx+1 : len(paramType)-1] // Remove outer brackets

	// Split parameters by comma, handling nested brackets
	var params []string
	bracketCount := 0
	start := 0

	for i, c := range paramStr {
		if c == '[' {
			bracketCount++
		} else if c == ']' {
			bracketCount--
		} else if c == ',' && bracketCount == 0 {
			params = append(params, strings.TrimSpace(paramStr[start:i]))
			start = i + 1
		}
	}

	// Add the last parameter
	if start < len(paramStr) {
		params = append(params, strings.TrimSpace(paramStr[start:]))
	}

	return baseType, params
}

// CheckTypeCompatibility checks if two types are compatible (for assignment, etc.)
func (h *TypeHierarchy) CheckTypeCompatibility(sourceType, targetType string) error {
	// Check if source is a subtype of target
	if h.IsSubtypeOf(sourceType, targetType) {
		return nil
	}

	// Special case: Maybe[T] can be assigned to T (but not the other way around)
	if strings.HasPrefix(sourceType, "Maybe[") {
		_, params := extractTypeAndParams(sourceType)
		if len(params) > 0 && h.IsSubtypeOf(params[0], targetType) {
			return nil
		}
	}

	// Types are incompatible
	return errors.NewTypeError(
		ErrTypeIncompatible,
		fmt.Sprintf("Type '%s' is not compatible with '%s'", sourceType, targetType),
		nil,
	)
}

// GetBaseType returns the base type and parameters for a parameterized type
func (h *TypeHierarchy) GetBaseType(typeName string) string {
	baseType, _ := extractTypeAndParams(typeName)
	return baseType
}

// ValidateType checks if a type is valid
func (h *TypeHierarchy) ValidateType(typeName string) error {
	// Check built-in types
	if h.IsBuiltinType(typeName) {
		return nil
	}

	// Check parameterized types
	if idx := strings.Index(typeName, "["); idx > 0 {
		baseType := typeName[:idx]
		_, isParamType := h.parameterizedTypes[baseType]
		if isParamType {
			// TODO: Validate parameters as well
			return nil
		}
	}

	// Check user-defined types from type definitions
	// This would require looking up type definitions
	// For now, we'll just accept any valid type name pattern
	if isValidTypeName(typeName) {
		return nil
	}

	return errors.NewTypeError(
		ErrTypeUndefined,
		fmt.Sprintf("Undefined type: %s", typeName),
		nil,
	)
}

// isValidTypeName checks if a string is a valid type name
func isValidTypeName(name string) bool {
	// Simple validation: must start with capital letter
	// and can contain alphanumeric characters and ::
	if len(name) == 0 {
		return false
	}

	// Handle parameterized types
	if strings.Contains(name, "[") {
		baseName, _ := extractTypeAndParams(name)
		return isValidTypeName(baseName)
	}

	if name[0] < 'A' || name[0] > 'Z' {
		return false
	}

	for _, part := range strings.Split(name, "::") {
		if len(part) == 0 || (part[0] < 'A' || part[0] > 'Z') {
			return false
		}
	}

	return true
}

// CreateParameterizedType creates a parameterized type from a base type and parameters
func (h *TypeHierarchy) CreateParameterizedType(baseType string, params []string) (string, error) {
	constructor, ok := h.parameterizedTypes[baseType]
	if !ok {
		return "", errors.NewTypeError(
			ErrTypeInvalid,
			fmt.Sprintf("Not a parameterized type: %s", baseType),
			nil,
		)
	}

	return constructor(params)
}