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

	// unionTypes stores created union type instances
	unionTypes map[string]*UnionType

	// intersectionTypes stores created intersection type instances
	intersectionTypes map[string]*IntersectionType
}

// NewTypeHierarchy creates a new TypeHierarchy with built-in types
func NewTypeHierarchy(store *Storage) *TypeHierarchy {
	hierarchy := &TypeHierarchy{
		BuiltinTypes:       make(map[string]*TypeInfo),
		TypeStore:          store,
		subtypeRelations:   make(map[string][]string),
		parameterizedTypes: make(map[string]func([]string) (string, error)),
		unionTypes:         make(map[string]*UnionType),
		intersectionTypes:  make(map[string]*IntersectionType),
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
	h.addBuiltinType("Union", "Union type that can be one of multiple types", "modifier")
	h.addBuiltinType("Intersection", "Intersection type that must satisfy all types", "modifier")

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
	h.addSubtypeRelation("Num", "Str") // Numbers can be automatically stringified in Perl
	h.addSubtypeRelation("Undef", "Scalar")

	h.addSubtypeRelation("Int", "Num")
	h.addSubtypeRelation("Float", "Num")
	h.addSubtypeRelation("Bool", "Int") // Bool is a subtype of Int per specification

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
			return "", fmt.Errorf("maybe requires exactly one type parameter")
		}
		return fmt.Sprintf("Maybe[%s]", params[0]), nil
	}

	h.parameterizedTypes["Optional"] = func(params []string) (string, error) {
		if len(params) != 1 {
			return "", fmt.Errorf("optional requires exactly one type parameter")
		}
		return fmt.Sprintf("Optional[%s]", params[0]), nil
	}

	h.parameterizedTypes["List"] = func(params []string) (string, error) {
		if len(params) != 1 {
			return "", fmt.Errorf("List requires exactly one type parameter")
		}
		return fmt.Sprintf("List[%s]", params[0]), nil
	}

	h.parameterizedTypes["Union"] = func(params []string) (string, error) {
		if len(params) < 2 {
			return "", fmt.Errorf("union requires at least two type parameters")
		}
		return fmt.Sprintf("Union[%s]", strings.Join(params, ", ")), nil
	}

	h.parameterizedTypes["Intersection"] = func(params []string) (string, error) {
		if len(params) < 2 {
			return "", fmt.Errorf("intersection requires at least two type parameters")
		}
		return fmt.Sprintf("Intersection[%s]", strings.Join(params, ", ")), nil
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
		if err := h.CheckTypeCompatibility(childParam, parentParams[i]); err != nil {
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

	// Check if the closing bracket exists
	if !strings.HasSuffix(paramType, "]") || len(paramType) <= idx+1 {
		return paramType, nil // Malformed, return as simple type
	}

	baseType := paramType[:idx]
	paramStr := paramType[idx+1 : len(paramType)-1] // Remove outer brackets

	// Split parameters by comma, handling nested brackets
	var params []string
	bracketCount := 0
	start := 0

	for i, c := range paramStr {
		switch c {
		case '[':
			bracketCount++
		case ']':
			bracketCount--
		case ',':
			if bracketCount == 0 {
				params = append(params, strings.TrimSpace(paramStr[start:i]))
				start = i + 1
			}
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

	// Special case: T and Undef can be assigned to Maybe[T]
	if strings.HasPrefix(targetType, "Maybe[") {
		_, params := extractTypeAndParams(targetType)
		if len(params) > 0 {
			// T -> Maybe[T] or Undef -> Maybe[T]
			if sourceType == "Undef" || h.IsSubtypeOf(sourceType, params[0]) {
				return nil
			}
		}
	}

	// Enhanced Union type handling using trait intersection
	if h.IsUnionType(sourceType) || h.IsUnionType(targetType) {
		return h.CheckUnionTypeCompatibility(sourceType, targetType)
	}

	// Enhanced Intersection type handling using trait intersection
	if h.IsIntersectionType(sourceType) || h.IsIntersectionType(targetType) {
		return h.CheckIntersectionTypeCompatibility(sourceType, targetType)
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
	// Handle empty type names
	if typeName == "" {
		return errors.NewTypeError(
			ErrTypeInvalid,
			"Type name cannot be empty",
			nil,
		)
	}

	// Check for parameterized types
	if idx := strings.Index(typeName, "["); idx > 0 {
		return h.validateParameterizedType(typeName)
	}

	// Check built-in types
	if _, exists := h.BuiltinTypes[typeName]; exists {
		return nil
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

// validateParameterizedType validates a parameterized type and its parameters
func (h *TypeHierarchy) validateParameterizedType(typeName string) error {
	baseType, params := extractTypeAndParams(typeName)

	// Check if the base type supports parameterization
	constructor, exists := h.parameterizedTypes[baseType]
	if !exists {
		return errors.NewTypeError(
			ErrTypeInvalid,
			fmt.Sprintf("Type '%s' is not parameterizable", baseType),
			nil,
		)
	}

	// Check if parameters are empty
	if len(params) == 0 {
		return errors.NewTypeError(
			ErrTypeInvalid,
			fmt.Sprintf("Type '%s' requires parameters", baseType),
			nil,
		)
	}

	// Validate each parameter recursively
	for i, param := range params {
		if err := h.ValidateType(param); err != nil {
			return errors.NewTypeError(
				ErrTypeInvalid,
				fmt.Sprintf("Invalid parameter %d ('%s') in type '%s': %s",
					i+1, param, typeName, err.Error()),
				err,
			)
		}
	}

	// Use the constructor to validate parameter count and compatibility
	_, err := constructor(params)
	if err != nil {
		return errors.NewTypeError(
			ErrTypeInvalid,
			fmt.Sprintf("Invalid parameterized type '%s': %s", typeName, err.Error()),
			err,
		)
	}

	return nil
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

// CreateUnionType creates and stores a union type instance
func (h *TypeHierarchy) CreateUnionType(members []string) *UnionType {
	unionType := NewUnionType(members)
	typeName := unionType.TypeName()
	h.unionTypes[typeName] = unionType
	return unionType
}

// GetUnionType retrieves a stored union type by its type name
func (h *TypeHierarchy) GetUnionType(typeName string) *UnionType {
	return h.unionTypes[typeName]
}

// IsUnionType checks if a type name represents a union type
func (h *TypeHierarchy) IsUnionType(typeName string) bool {
	// Check if it's a stored union type
	if _, exists := h.unionTypes[typeName]; exists {
		return true
	}

	// Check if it's a union type format like "Union[A, B]" or "A|B"
	if strings.HasPrefix(typeName, "Union[") {
		return true
	}

	// Check for pipe-separated format, but only at top level (not inside brackets)
	return h.containsTopLevelPipe(typeName)
}

// containsTopLevelPipe checks if a string contains "|" at the top level (not inside brackets)
func (h *TypeHierarchy) containsTopLevelPipe(typeName string) bool {
	bracketCount := 0
	for _, c := range typeName {
		switch c {
		case '[':
			bracketCount++
		case ']':
			bracketCount--
		case '|':
			if bracketCount == 0 {
				return true
			}
		}
	}
	return false
}

// ParseUnionType parses a union type string and creates a UnionType instance
func (h *TypeHierarchy) ParseUnionType(typeName string) *UnionType {
	// Check if it's already stored
	if unionType, exists := h.unionTypes[typeName]; exists {
		return unionType
	}

	var members []string

	// Handle different union type formats
	switch {
	case strings.HasPrefix(typeName, "Union["):
		// Handle "Union[A, B, C]" format
		_, params := extractTypeAndParams(typeName)
		members = params
	case strings.Contains(typeName, "|"):
		// Handle "A|B|C" format
		parts := strings.Split(typeName, "|")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" { // Only add non-empty parts
				members = append(members, part)
			}
		}
	default:
		// Not a union type
		return nil
	}

	if len(members) < 2 {
		return nil
	}

	// Create and store the union type
	unionType := NewUnionType(members)
	h.unionTypes[unionType.TypeName()] = unionType
	return unionType
}

// CheckUnionTypeCompatibility checks union type compatibility using trait intersection
func (h *TypeHierarchy) CheckUnionTypeCompatibility(sourceType, targetType string) error {
	// Handle union to non-union
	if h.IsUnionType(sourceType) && !h.IsUnionType(targetType) {
		unionType := h.ParseUnionType(sourceType)
		if unionType == nil {
			return fmt.Errorf("invalid union type: %s", sourceType)
		}

		if unionType.IsCompatibleWith(targetType, h) {
			return nil
		}

		return fmt.Errorf("union type %s is not compatible with %s", sourceType, targetType)
	}

	// Handle non-union to union
	if !h.IsUnionType(sourceType) && h.IsUnionType(targetType) {
		unionType := h.ParseUnionType(targetType)
		if unionType == nil {
			return fmt.Errorf("invalid union type: %s", targetType)
		}

		if unionType.CanAssignFrom(sourceType, h) {
			return nil
		}

		return fmt.Errorf("type %s cannot be assigned to union type %s", sourceType, targetType)
	}

	// Handle union to union
	if h.IsUnionType(sourceType) && h.IsUnionType(targetType) {
		sourceUnion := h.ParseUnionType(sourceType)
		targetUnion := h.ParseUnionType(targetType)

		if sourceUnion == nil || targetUnion == nil {
			return fmt.Errorf("invalid union types: %s, %s", sourceType, targetType)
		}

		// Union[A, B] is compatible with Union[C, D] if every member of source
		// is compatible with at least one member of target
		for _, sourceMember := range sourceUnion.GetMembers() {
			if !targetUnion.CanAssignFrom(sourceMember, h) {
				return fmt.Errorf("union type %s is not compatible with union type %s", sourceType, targetType)
			}
		}

		return nil
	}

	// Neither is union, shouldn't reach here in normal flow
	return fmt.Errorf("internal error: non-union types in union compatibility check")
}

// CreateIntersectionType creates and stores an intersection type instance
func (h *TypeHierarchy) CreateIntersectionType(members []string) *IntersectionType {
	intersectionType := NewIntersectionType(members)
	typeName := intersectionType.TypeName()
	h.intersectionTypes[typeName] = intersectionType
	return intersectionType
}

// GetIntersectionType retrieves a stored intersection type by its type name
func (h *TypeHierarchy) GetIntersectionType(typeName string) *IntersectionType {
	return h.intersectionTypes[typeName]
}

// IsIntersectionType checks if a type name represents an intersection type
func (h *TypeHierarchy) IsIntersectionType(typeName string) bool {
	// Check if it's a stored intersection type
	if _, exists := h.intersectionTypes[typeName]; exists {
		return true
	}

	// Check if it's an intersection type format like "Intersection[A, B]" or "A&B"
	if strings.HasPrefix(typeName, "Intersection[") {
		return true
	}

	// Check for ampersand-separated format
	return strings.Contains(typeName, "&")
}

// ParseIntersectionType parses an intersection type string and creates an IntersectionType instance
func (h *TypeHierarchy) ParseIntersectionType(typeName string) *IntersectionType {
	// Check if it's already stored
	if intersectionType, exists := h.intersectionTypes[typeName]; exists {
		return intersectionType
	}

	var members []string

	// Handle different intersection type formats
	switch {
	case strings.HasPrefix(typeName, "Intersection["):
		// Handle "Intersection[A, B, C]" format
		_, params := extractTypeAndParams(typeName)
		members = params
	case strings.Contains(typeName, "&"):
		// Handle "A&B&C" format
		parts := strings.Split(typeName, "&")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" { // Only add non-empty parts
				members = append(members, part)
			}
		}
	default:
		// Not an intersection type
		return nil
	}

	if len(members) < 2 {
		return nil
	}

	// Create and store the intersection type
	intersectionType := NewIntersectionType(members)
	h.intersectionTypes[intersectionType.TypeName()] = intersectionType
	return intersectionType
}

// CheckIntersectionTypeCompatibility checks intersection type compatibility using trait intersection
func (h *TypeHierarchy) CheckIntersectionTypeCompatibility(sourceType, targetType string) error {
	// Handle intersection to non-intersection
	if h.IsIntersectionType(sourceType) && !h.IsIntersectionType(targetType) {
		intersectionType := h.ParseIntersectionType(sourceType)
		if intersectionType == nil {
			return fmt.Errorf("invalid intersection type: %s", sourceType)
		}

		if intersectionType.IsCompatibleWith(targetType, h) {
			return nil
		}

		return fmt.Errorf("intersection type %s is not compatible with %s", sourceType, targetType)
	}

	// Handle non-intersection to intersection
	if !h.IsIntersectionType(sourceType) && h.IsIntersectionType(targetType) {
		intersectionType := h.ParseIntersectionType(targetType)
		if intersectionType == nil {
			return fmt.Errorf("invalid intersection type: %s", targetType)
		}

		if intersectionType.CanAssignFrom(sourceType, h) {
			return nil
		}

		return fmt.Errorf("type %s cannot be assigned to intersection type %s", sourceType, targetType)
	}

	// Handle intersection to intersection
	if h.IsIntersectionType(sourceType) && h.IsIntersectionType(targetType) {
		sourceIntersection := h.ParseIntersectionType(sourceType)
		targetIntersection := h.ParseIntersectionType(targetType)

		if sourceIntersection == nil || targetIntersection == nil {
			return fmt.Errorf("invalid intersection types: %s, %s", sourceType, targetType)
		}

		// Intersection[A, B] is compatible with Intersection[C, D] if every member of target
		// is compatible with at least one member of source (reverse of union logic)
		for _, targetMember := range targetIntersection.GetMembers() {
			compatible := false
			for _, sourceMember := range sourceIntersection.GetMembers() {
				if err := h.CheckTypeCompatibility(sourceMember, targetMember); err == nil {
					compatible = true
					break
				}
			}
			if !compatible {
				return fmt.Errorf("intersection type %s is not compatible with intersection type %s", sourceType, targetType)
			}
		}

		return nil
	}

	// Neither is intersection, shouldn't reach here in normal flow
	return fmt.Errorf("internal error: non-intersection types in intersection compatibility check")
}
