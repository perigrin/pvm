// ABOUTME: External library type hint integration for type inference
// ABOUTME: Provides type information from external sources like documentation and annotations

package inference

import (
	"encoding/json"
	"os"
	"path/filepath"

	"tamarou.com/pvm/internal/types"
)

// ExternalHintProvider provides type hints from external sources
type ExternalHintProvider struct {
	// Map of module names to their type definitions
	moduleTypes map[string]*ModuleTypeInfo

	// Map of function names to their signatures
	functionSignatures map[string]*types.TypeInfo

	// Configuration for hint loading
	hintPaths []string
}

// ModuleTypeInfo contains type information for a Perl module
type ModuleTypeInfo struct {
	// Module name (e.g., "List::Util")
	ModuleName string `json:"module_name"`

	// Functions exported by the module
	Functions map[string]*FunctionTypeHint `json:"functions"`

	// Classes/objects provided by the module
	Classes map[string]*ClassTypeHint `json:"classes"`

	// Constants defined by the module
	Constants map[string]*ConstantTypeHint `json:"constants"`
}

// FunctionTypeHint provides type information for a function
type FunctionTypeHint struct {
	// Function name
	Name string `json:"name"`

	// Parameter types
	ParameterTypes []string `json:"parameter_types"`

	// Return type
	ReturnType string `json:"return_type"`

	// Whether the function is context-sensitive
	ContextSensitive bool `json:"context_sensitive"`

	// Return types in different contexts
	ContextualReturnTypes map[string]string `json:"contextual_return_types,omitempty"`

	// Confidence level for this hint
	Confidence float64 `json:"confidence"`

	// Documentation/description
	Description string `json:"description,omitempty"`
}

// ClassTypeHint provides type information for a class/object
type ClassTypeHint struct {
	// Class name
	Name string `json:"name"`

	// Methods provided by the class
	Methods map[string]*FunctionTypeHint `json:"methods"`

	// Attributes/fields of the class
	Attributes map[string]*AttributeTypeHint `json:"attributes"`

	// Parent classes
	Inherits []string `json:"inherits,omitempty"`

	// Confidence level
	Confidence float64 `json:"confidence"`
}

// AttributeTypeHint provides type information for class attributes
type AttributeTypeHint struct {
	// Attribute name
	Name string `json:"name"`

	// Type of the attribute
	Type string `json:"type"`

	// Whether it's read-only
	ReadOnly bool `json:"read_only"`

	// Confidence level
	Confidence float64 `json:"confidence"`
}

// ConstantTypeHint provides type information for constants
type ConstantTypeHint struct {
	// Constant name
	Name string `json:"name"`

	// Type of the constant
	Type string `json:"type"`

	// Constant value (if known)
	Value string `json:"value,omitempty"`

	// Confidence level
	Confidence float64 `json:"confidence"`
}

// NewExternalHintProvider creates a new external hint provider
func NewExternalHintProvider(hintPaths []string) *ExternalHintProvider {
	return &ExternalHintProvider{
		moduleTypes:        make(map[string]*ModuleTypeInfo),
		functionSignatures: make(map[string]*types.TypeInfo),
		hintPaths:          hintPaths,
	}
}

// LoadHints loads type hints from external sources
func (ehp *ExternalHintProvider) LoadHints() error {
	// Load from default locations
	defaultPaths := []string{
		"./type_hints",
		"~/.pvm/type_hints",
		"/usr/share/pvm/type_hints",
	}

	allPaths := append(ehp.hintPaths, defaultPaths...)

	for _, hintPath := range allPaths {
		if err := ehp.loadHintsFromPath(hintPath); err != nil {
			// Log error but continue - type hints are optional
			continue
		}
	}

	return nil
}

// loadHintsFromPath loads type hints from a specific directory
func (ehp *ExternalHintProvider) loadHintsFromPath(hintPath string) error {
	// Check if path exists
	if _, err := os.Stat(hintPath); os.IsNotExist(err) {
		return err
	}

	// Walk through the directory looking for JSON files
	return filepath.Walk(hintPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .json files
		if filepath.Ext(path) != ".json" {
			return nil
		}

		return ehp.loadHintFile(path)
	})
}

// loadHintFile loads type hints from a JSON file
func (ehp *ExternalHintProvider) loadHintFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var moduleInfo ModuleTypeInfo
	if err := json.Unmarshal(data, &moduleInfo); err != nil {
		return err
	}

	// Store the module information
	ehp.moduleTypes[moduleInfo.ModuleName] = &moduleInfo

	// Index functions for quick lookup
	for funcName, funcHint := range moduleInfo.Functions {
		qualifiedName := moduleInfo.ModuleName + "::" + funcName
		typeInfo := ehp.convertFunctionHintToTypeInfo(funcHint)
		ehp.functionSignatures[qualifiedName] = typeInfo

		// Also store unqualified name for built-in modules
		if ehp.isBuiltinModule(moduleInfo.ModuleName) {
			ehp.functionSignatures[funcName] = typeInfo
		}
	}

	return nil
}

// GetFunctionTypeHint retrieves type information for a function
func (ehp *ExternalHintProvider) GetFunctionTypeHint(functionName string) *types.TypeInfo {
	if typeInfo, exists := ehp.functionSignatures[functionName]; exists {
		return typeInfo
	}
	return nil
}

// GetModuleTypeInfo retrieves type information for a module
func (ehp *ExternalHintProvider) GetModuleTypeInfo(moduleName string) *ModuleTypeInfo {
	if moduleInfo, exists := ehp.moduleTypes[moduleName]; exists {
		return moduleInfo
	}
	return nil
}

// GetClassTypeHint retrieves type information for a class
func (ehp *ExternalHintProvider) GetClassTypeHint(className string) *ClassTypeHint {
	// Search through all modules for the class
	for _, moduleInfo := range ehp.moduleTypes {
		if classHint, exists := moduleInfo.Classes[className]; exists {
			return classHint
		}
	}
	return nil
}

// GetMethodTypeHint retrieves type information for a method
func (ehp *ExternalHintProvider) GetMethodTypeHint(className, methodName string) *FunctionTypeHint {
	classHint := ehp.GetClassTypeHint(className)
	if classHint == nil {
		return nil
	}

	if methodHint, exists := classHint.Methods[methodName]; exists {
		return methodHint
	}
	return nil
}

// HasHintsForModule checks if we have type hints for a module
func (ehp *ExternalHintProvider) HasHintsForModule(moduleName string) bool {
	_, exists := ehp.moduleTypes[moduleName]
	return exists
}

// ListAvailableModules returns a list of modules with type hints
func (ehp *ExternalHintProvider) ListAvailableModules() []string {
	modules := make([]string, 0, len(ehp.moduleTypes))
	for moduleName := range ehp.moduleTypes {
		modules = append(modules, moduleName)
	}
	return modules
}

// Helper methods

// convertFunctionHintToTypeInfo converts a function hint to TypeInfo
func (ehp *ExternalHintProvider) convertFunctionHintToTypeInfo(hint *FunctionTypeHint) *types.TypeInfo {
	// Convert string type names to actual types
	returnType := ehp.stringToType(hint.ReturnType)

	// Create type info with external source and given confidence
	confidence := hint.Confidence
	if confidence == 0 {
		confidence = 0.90 // High confidence for external hints
	}

	return types.NewTypeInfo(returnType, confidence, types.SourceExternal)
}

// stringToType converts a string type name to a Type instance
func (ehp *ExternalHintProvider) stringToType(typeName string) types.Type {
	switch typeName {
	case "Int", "int":
		return types.NewIntType()
	case "Str", "string":
		return types.NewStrType()
	case "Bool", "boolean":
		return types.NewBoolType()
	case "Num", "number":
		return types.NewNumType()
	case "Ref", "reference":
		return types.NewRefType()
	default:
		// Handle complex types
		if typeName == "ArrayRef[Str]" {
			return types.NewArrayRefType(types.NewStrType())
		}
		if typeName == "ArrayRef[Int]" {
			return types.NewArrayRefType(types.NewIntType())
		}
		if typeName == "HashRef[Str]" {
			return types.NewHashRefType(types.NewStrType())
		}
		if typeName == "HashRef[Int]" {
			return types.NewHashRefType(types.NewIntType())
		}

		// Default to string for unknown types
		return types.NewStrType()
	}
}

// isBuiltinModule checks if a module is a built-in Perl module
func (ehp *ExternalHintProvider) isBuiltinModule(moduleName string) bool {
	builtinModules := map[string]bool{
		"CORE":         true,
		"UNIVERSAL":    true,
		"List::Util":   true,
		"Scalar::Util": true,
		"Data::Dumper": true,
		"File::Spec":   true,
		"File::Path":   true,
		"JSON":         true,
		"YAML":         true,
		"DBI":          true,
		"LWP":          true,
		"DateTime":     true,
		"Moose":        true,
		"Moo":          true,
	}

	return builtinModules[moduleName]
}

// CreateCoreHints creates type hints for Perl core functions
func (ehp *ExternalHintProvider) CreateCoreHints() {
	coreModule := &ModuleTypeInfo{
		ModuleName: "CORE",
		Functions: map[string]*FunctionTypeHint{
			"length": {
				Name:           "length",
				ParameterTypes: []string{"Str"},
				ReturnType:     "Int",
				Confidence:     0.99,
				Description:    "Returns the length of a string",
			},
			"substr": {
				Name:           "substr",
				ParameterTypes: []string{"Str", "Int", "Int"},
				ReturnType:     "Str",
				Confidence:     0.99,
				Description:    "Extracts a substring",
			},
			"index": {
				Name:           "index",
				ParameterTypes: []string{"Str", "Str"},
				ReturnType:     "Int",
				Confidence:     0.99,
				Description:    "Finds the position of a substring",
			},
			"split": {
				Name:             "split",
				ParameterTypes:   []string{"Str", "Str"},
				ReturnType:       "ArrayRef[Str]",
				ContextSensitive: true,
				ContextualReturnTypes: map[string]string{
					"scalar": "Int",
					"list":   "ArrayRef[Str]",
				},
				Confidence:  0.99,
				Description: "Splits a string into an array",
			},
			"join": {
				Name:           "join",
				ParameterTypes: []string{"Str", "ArrayRef[Str]"},
				ReturnType:     "Str",
				Confidence:     0.99,
				Description:    "Joins array elements into a string",
			},
			"defined": {
				Name:           "defined",
				ParameterTypes: []string{"Any"},
				ReturnType:     "Bool",
				Confidence:     0.99,
				Description:    "Tests whether a value is defined",
			},
			"exists": {
				Name:           "exists",
				ParameterTypes: []string{"Any"},
				ReturnType:     "Bool",
				Confidence:     0.99,
				Description:    "Tests whether a hash key or array element exists",
			},
		},
		Classes:   make(map[string]*ClassTypeHint),
		Constants: make(map[string]*ConstantTypeHint),
	}

	ehp.moduleTypes["CORE"] = coreModule

	// Index the core functions
	for funcName, funcHint := range coreModule.Functions {
		typeInfo := ehp.convertFunctionHintToTypeInfo(funcHint)
		ehp.functionSignatures[funcName] = typeInfo
	}
}

// AddCustomHint allows adding a custom type hint programmatically
func (ehp *ExternalHintProvider) AddCustomHint(functionName string, returnType types.Type, confidence float64) {
	typeInfo := types.NewTypeInfo(returnType, confidence, types.SourceExternal)
	ehp.functionSignatures[functionName] = typeInfo
}

// ExportHints exports current hints to a JSON file
func (ehp *ExternalHintProvider) ExportHints(filePath string) error {
	// Create a summary of all hints
	summary := make(map[string]*ModuleTypeInfo)
	for moduleName, moduleInfo := range ehp.moduleTypes {
		summary[moduleName] = moduleInfo
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}
