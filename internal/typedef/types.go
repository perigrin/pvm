// ABOUTME: Type definition structures and interfaces
// ABOUTME: Defines the Perl Type Definition (ptd) file format

package typedef

import (
	"fmt"
	"strings"
	"time"

	"tamarou.com/pvm/internal/traits"
)

// TypeDefinition represents a type definition for a Perl module
type TypeDefinition struct {
	// Module information
	Module     string    `json:"module"`     // Module name (e.g., "Moose", "Path::Tiny")
	Version    string    `json:"version"`    // Module version (e.g., "2.2015")
	Generated  time.Time `json:"generated"`  // When the definition was generated
	Maintainer string    `json:"maintainer"` // Who maintains this type definition
	Source     string    `json:"source"`     // Source of the type definition (e.g., "static", "dynamic", "manual")

	// Type information
	Types    []TypeInfo    `json:"types"`    // Types defined in the module
	Packages []PackageInfo `json:"packages"` // Packages defined in the module
	Subs     []SubInfo     `json:"subs"`     // Subroutines defined in the module
	Methods  []MethodInfo  `json:"methods"`  // Methods defined in the module
}

// TypeInfo represents information about a type
type TypeInfo struct {
	Name        string       `json:"name"`        // Type name
	Description string       `json:"description"` // Type description
	Kind        string       `json:"kind"`        // Type kind (e.g., "class", "role", "enum", "scalar", "union")
	Parameters  []ParamInfo  `json:"parameters"`  // Type parameters (for parameterized types)
	Properties  []PropInfo   `json:"properties"`  // Type properties
	Methods     []MethodInfo `json:"methods"`     // Type methods
	Parent      string       `json:"parent"`      // Parent type name (for inheritance)
	Roles       []string     `json:"roles"`       // Roles this type consumes
}

// PackageInfo represents information about a package
type PackageInfo struct {
	Name        string       `json:"name"`        // Package name
	Description string       `json:"description"` // Package description
	Exports     []ExportInfo `json:"exports"`     // Exported symbols
}

// SubInfo represents information about a subroutine
type SubInfo struct {
	Name        string       `json:"name"`        // Subroutine name
	Description string       `json:"description"` // Subroutine description
	Parameters  []ParamInfo  `json:"parameters"`  // Subroutine parameters
	Returns     []ReturnInfo `json:"returns"`     // Return type information
	Throws      []string     `json:"throws"`      // Exceptions this subroutine may throw
	IsMethod    bool         `json:"is_method"`   // Whether this is a method
	IsPrivate   bool         `json:"is_private"`  // Whether this is a private subroutine
}

// MethodInfo represents information about a method
type MethodInfo struct {
	Name        string       `json:"name"`        // Method name
	Description string       `json:"description"` // Method description
	Parameters  []ParamInfo  `json:"parameters"`  // Method parameters
	Returns     []ReturnInfo `json:"returns"`     // Return type information
	Throws      []string     `json:"throws"`      // Exceptions this method may throw
	IsPrivate   bool         `json:"is_private"`  // Whether this is a private method
	IsStatic    bool         `json:"is_static"`   // Whether this is a static method
}

// ParamInfo represents information about a parameter
type ParamInfo struct {
	Name        string `json:"name"`        // Parameter name
	Type        string `json:"type"`        // Parameter type
	Description string `json:"description"` // Parameter description
	Optional    bool   `json:"optional"`    // Whether this parameter is optional
	Default     string `json:"default"`     // Default value for this parameter
}

// ParameterInfo is an alias for ParamInfo for backwards compatibility
type ParameterInfo = ParamInfo

// PropInfo represents information about a property
type PropInfo struct {
	Name        string `json:"name"`        // Property name
	Type        string `json:"type"`        // Property type
	Description string `json:"description"` // Property description
	Optional    bool   `json:"optional"`    // Whether this property is optional
	Default     string `json:"default"`     // Default value for this property
	ReadOnly    bool   `json:"read_only"`   // Whether this property is read-only
}

// ReturnInfo represents information about a return type
type ReturnInfo struct {
	Type        string `json:"type"`        // Return type
	Description string `json:"description"` // Return description
}

// ExportInfo represents information about an exported symbol
type ExportInfo struct {
	Name        string `json:"name"`        // Symbol name
	Type        string `json:"type"`        // Symbol type (e.g., "sub", "const")
	Description string `json:"description"` // Symbol description
}

// Error definition for type operations
type TypeDefError string

// Error implements the error interface
func (e TypeDefError) Error() string {
	return string(e)
}

// String returns a string representation of a TypeDefinition
func (td *TypeDefinition) String() string {
	return fmt.Sprintf("TypeDefinition for %s v%s", td.Module, td.Version)
}

// String returns a string representation of a TypeInfo
func (ti *TypeInfo) String() string {
	return fmt.Sprintf("Type %s (%s)", ti.Name, ti.Kind)
}

// UnionType represents a union type that can be one of multiple member types
type UnionType struct {
	// Members are the types that make up this union
	Members []string

	// cachedTraits stores the computed trait intersection (lazy loading)
	cachedTraits *traits.TraitSet

	// intersector for computing trait intersections
	intersector *traits.TraitIntersector
}

// NewUnionType creates a new union type with the given member types
func NewUnionType(members []string) *UnionType {
	if len(members) < 2 {
		panic("Union type must have at least two members")
	}

	// Remove duplicates and maintain order
	uniqueMembers := make([]string, 0, len(members))
	seen := make(map[string]bool)
	for _, member := range members {
		if !seen[member] {
			uniqueMembers = append(uniqueMembers, member)
			seen[member] = true
		}
	}

	return &UnionType{
		Members:     uniqueMembers,
		intersector: traits.NewTraitIntersector(),
	}
}

// String returns a string representation of the union type
func (ut *UnionType) String() string {
	return strings.Join(ut.Members, "|")
}

// TypeName returns the full type name for this union
func (ut *UnionType) TypeName() string {
	return fmt.Sprintf("Union[%s]", strings.Join(ut.Members, ", "))
}

// GetTraits returns the trait intersection for this union type (lazy computed)
func (ut *UnionType) GetTraits() *traits.TraitSet {
	if ut.cachedTraits == nil {
		ut.cachedTraits = ut.intersector.IntersectTypes(ut.Members)
	}
	return ut.cachedTraits
}

// SupportsOperation checks if this union type supports the given operation
func (ut *UnionType) SupportsOperation(operation string) bool {
	traits := ut.GetTraits()
	return traits.HasTrait(operation)
}

// GetOperationResultType returns the result type for the given operation
func (ut *UnionType) GetOperationResultType(operation string) (string, error) {
	traits := ut.GetTraits()
	if !traits.HasTrait(operation) {
		return "", fmt.Errorf("operation '%s' not supported by union type %s", operation, ut.String())
	}

	resultType := traits.GetResultType(operation)
	if resultType == "" {
		return "", fmt.Errorf("no result type for operation '%s'", operation)
	}

	return resultType, nil
}

// GetMembers returns a copy of the member types
func (ut *UnionType) GetMembers() []string {
	result := make([]string, len(ut.Members))
	copy(result, ut.Members)
	return result
}

// ContainsMember checks if the union contains the given type as a member
func (ut *UnionType) ContainsMember(typeName string) bool {
	for _, member := range ut.Members {
		if member == typeName {
			return true
		}
	}
	return false
}

// IsCompatibleWith checks if this union type is compatible with another type
func (ut *UnionType) IsCompatibleWith(targetType string, hierarchy *TypeHierarchy) bool {
	// Union[A, B] is compatible with T if any member is compatible with T
	for _, member := range ut.Members {
		if err := hierarchy.CheckTypeCompatibility(member, targetType); err == nil {
			return true
		}
	}
	return false
}

// CanAssignFrom checks if a source type can be assigned to this union
func (ut *UnionType) CanAssignFrom(sourceType string, hierarchy *TypeHierarchy) bool {
	// T can be assigned to Union[A, B] if T is compatible with any member
	for _, member := range ut.Members {
		if err := hierarchy.CheckTypeCompatibility(sourceType, member); err == nil {
			return true
		}
	}
	return false
}

// ClearTraitCache clears the cached traits, forcing recomputation next time
func (ut *UnionType) ClearTraitCache() {
	ut.cachedTraits = nil
	if ut.intersector != nil {
		ut.intersector.ClearCache()
	}
}

// Equals checks if this union type is equal to another union type
func (ut *UnionType) Equals(other *UnionType) bool {
	if len(ut.Members) != len(other.Members) {
		return false
	}

	// Check if all members match (order independent)
	for _, member := range ut.Members {
		if !other.ContainsMember(member) {
			return false
		}
	}
	return true
}

// IntersectionType represents an intersection type that must satisfy all member types
type IntersectionType struct {
	// Members are the types that make up this intersection
	Members []string

	// cachedTraits stores the computed trait intersection (lazy loading)
	cachedTraits *traits.TraitSet

	// intersector for computing trait intersections
	intersector *traits.TraitIntersector
}

// NewIntersectionType creates a new intersection type with the given member types
func NewIntersectionType(members []string) *IntersectionType {
	if len(members) < 2 {
		panic("Intersection type must have at least two members")
	}

	// Remove duplicates and maintain order
	uniqueMembers := make([]string, 0, len(members))
	seen := make(map[string]bool)
	for _, member := range members {
		if !seen[member] {
			uniqueMembers = append(uniqueMembers, member)
			seen[member] = true
		}
	}

	return &IntersectionType{
		Members:     uniqueMembers,
		intersector: traits.NewTraitIntersector(),
	}
}

// String returns a string representation of the intersection type
func (it *IntersectionType) String() string {
	return strings.Join(it.Members, "&")
}

// TypeName returns the full type name for this intersection
func (it *IntersectionType) TypeName() string {
	return fmt.Sprintf("Intersection[%s]", strings.Join(it.Members, ", "))
}

// GetTraits returns the trait intersection for this intersection type (lazy computed)
func (it *IntersectionType) GetTraits() *traits.TraitSet {
	if it.cachedTraits == nil {
		it.cachedTraits = it.intersector.IntersectTypes(it.Members)
	}
	return it.cachedTraits
}

// SupportsOperation checks if this intersection type supports the given operation
func (it *IntersectionType) SupportsOperation(operation string) bool {
	traits := it.GetTraits()
	return traits.HasTrait(operation)
}

// GetOperationResultType returns the result type for the given operation
func (it *IntersectionType) GetOperationResultType(operation string) (string, error) {
	traits := it.GetTraits()
	if !traits.HasTrait(operation) {
		return "", fmt.Errorf("operation '%s' not supported by intersection type %s", operation, it.String())
	}

	resultType := traits.GetResultType(operation)
	if resultType == "" {
		return "", fmt.Errorf("no result type for operation '%s'", operation)
	}

	return resultType, nil
}

// GetMembers returns a copy of the member types
func (it *IntersectionType) GetMembers() []string {
	result := make([]string, len(it.Members))
	copy(result, it.Members)
	return result
}

// ContainsMember checks if the intersection contains the given type as a member
func (it *IntersectionType) ContainsMember(typeName string) bool {
	for _, member := range it.Members {
		if member == typeName {
			return true
		}
	}
	return false
}

// IsCompatibleWith checks if this intersection type is compatible with another type
func (it *IntersectionType) IsCompatibleWith(targetType string, hierarchy *TypeHierarchy) bool {
	// Intersection[A, B] is compatible with T if the intersection can be assigned to T
	// Since intersection types represent values that satisfy ALL members,
	// the intersection is compatible with T if at least one member is compatible with T
	for _, member := range it.Members {
		if err := hierarchy.CheckTypeCompatibility(member, targetType); err == nil {
			return true
		}
	}
	return false
}

// CanAssignFrom checks if a source type can be assigned to this intersection
func (it *IntersectionType) CanAssignFrom(sourceType string, hierarchy *TypeHierarchy) bool {
	// T can be assigned to Intersection[A, B] if T is compatible with all members
	for _, member := range it.Members {
		if err := hierarchy.CheckTypeCompatibility(sourceType, member); err != nil {
			return false
		}
	}
	return true
}

// ClearTraitCache clears the cached traits, forcing recomputation next time
func (it *IntersectionType) ClearTraitCache() {
	it.cachedTraits = nil
	if it.intersector != nil {
		it.intersector.ClearCache()
	}
}

// Equals checks if this intersection type is equal to another intersection type
func (it *IntersectionType) Equals(other *IntersectionType) bool {
	if len(it.Members) != len(other.Members) {
		return false
	}

	// Check if all members match (order independent)
	for _, member := range it.Members {
		if !other.ContainsMember(member) {
			return false
		}
	}
	return true
}
