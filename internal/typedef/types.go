// ABOUTME: Type definition structures and interfaces
// ABOUTME: Defines the Perl Type Definition (ptd) file format

package typedef

import (
	"fmt"
	"time"
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
