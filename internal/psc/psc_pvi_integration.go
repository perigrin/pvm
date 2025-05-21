// ABOUTME: Integration between PSC and PVI for type definitions
// ABOUTME: Provides functionality to share type definitions between components

package psc

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

// TypeDefinitionOptions contains configuration for type definition operations
type TypeDefinitionOptions struct {
	// ModuleName is the name of the module to work with
	ModuleName string

	// SourceFile is the source Perl file to extract type definitions from
	SourceFile string

	// Output specifies whether to save the type definition
	Save bool

	// OutputFile is the file to write the type definition to (optional)
	OutputFile string

	// Verbose enables verbose output
	Verbose bool
}

// TypeDefinitionResult contains the result of a type definition operation
type TypeDefinitionResult struct {
	// TypeDef is the generated type definition
	TypeDef *typedef.TypeDefinition

	// SavedPath is the path where the type definition was saved (if applicable)
	SavedPath string

	// Errors contains any errors encountered during the operation
	Errors []error
}

// GenerateTypeDefinition creates a type definition from a Perl file
// It extracts type annotations from the file and converts them to a type definition
func GenerateTypeDefinition(options *TypeDefinitionOptions) (*TypeDefinitionResult, error) {
	if options == nil {
		return nil, errors.NewTypeError(
			"704",
			"No options provided for type definition generation",
			nil,
		)
	}

	result := &TypeDefinitionResult{
		Errors: []error{},
	}

	// Create a new type checker
	tc, err := parser.NewTypeCheck()
	if err != nil {
		return nil, errors.NewTypeError(
			"704",
			"Failed to create type checker",
			err,
		)
	}

	// Initialize the type definition
	moduleName := options.ModuleName
	if moduleName == "" && options.SourceFile != "" {
		// Try to extract module name from the file
		moduleName = extractModuleNameFromPath(options.SourceFile)
	}

	if moduleName == "" {
		return nil, errors.NewTypeError(
			"704",
			"No module name provided or could be extracted",
			nil,
		)
	}

	// Create the type definition
	typeDef := &typedef.TypeDefinition{
		Module:     moduleName,
		Version:    "0.0.1", // To be determined from module if possible
		Generated:  time.Now(),
		Maintainer: "PSC type definition generator",
		Source:     options.SourceFile,
		Types:      []typedef.TypeInfo{},
		Packages:   []typedef.PackageInfo{},
		Subs:       []typedef.SubInfo{},
		Methods:    []typedef.MethodInfo{},
	}

	// If a source file is provided, extract type annotations from it
	if options.SourceFile != "" {
		if options.Verbose {
			fmt.Printf("Generating type definition for %s from %s\n", moduleName, options.SourceFile)
		}

		// Parse the file to extract type annotations
		checkResult, err := tc.CheckFile(options.SourceFile)
		if err != nil {
			return nil, errors.NewTypeError(
				"704",
				fmt.Sprintf("Failed to check file %s", options.SourceFile),
				err,
			)
		}

		// Build type definition from annotations
		err = populateTypeDefinition(typeDef, checkResult)
		if err != nil {
			result.Errors = append(result.Errors, err)
		}
	}

	// Save the type definition if requested
	if options.Save {
		storage, err := typedef.NewStorage()
		if err != nil {
			result.Errors = append(result.Errors, errors.NewTypeError(
				"704",
				"Failed to create type storage",
				err,
			))
		} else {
			if err := storage.Save(typeDef); err != nil {
				result.Errors = append(result.Errors, errors.NewTypeError(
					"704",
					fmt.Sprintf("Failed to save type definition for %s", moduleName),
					err,
				))
			} else if options.Verbose {
				fmt.Printf("Saved type definition for %s\n", moduleName)
				result.SavedPath = storage.GetPathForModule(moduleName)
			}
		}
	}

	// Set the result type definition
	result.TypeDef = typeDef

	return result, nil
}

// populateTypeDefinition populates a type definition from type check results
func populateTypeDefinition(typeDef *typedef.TypeDefinition, checkResult *parser.TypeCheckResult) error {
	// Map to track types we've already added
	addedTypes := make(map[string]bool)

	// Process each type annotation
	for _, annotation := range checkResult.TypeAnnotations {
		typeStr := annotation.TypeExpression.String()

		// Process the annotation based on its kind
		switch annotation.Kind {
		case parser.VarAnnotation:
			// For variable annotations, extract the type
			addTypeIfNew(typeDef, typeStr, addedTypes)

		case parser.SubParamAnnotation, parser.MethodParamAnnotation:
			// For parameter annotations, add to subroutine or method info
			addTypeIfNew(typeDef, typeStr, addedTypes)
			addParameterType(typeDef, annotation)

		case parser.SubReturnAnnotation, parser.MethodReturnAnnotation:
			// For return annotations, add to subroutine or method info
			addTypeIfNew(typeDef, typeStr, addedTypes)
			addReturnType(typeDef, annotation)

		case parser.AttrAnnotation:
			// For attribute annotations, add to package info
			addTypeIfNew(typeDef, typeStr, addedTypes)
			addAttributeType(typeDef, annotation)

		case parser.TypeDeclAnnotation:
			// For type declarations, add directly to the types list
			addTypeDeclaration(typeDef, annotation)
		}
	}

	return nil
}

// addTypeIfNew adds a type to the types list if it's not already there
func addTypeIfNew(typeDef *typedef.TypeDefinition, typeStr string, addedTypes map[string]bool) {
	// Skip if already added
	if addedTypes[typeStr] {
		return
	}

	// Handle parameterized types (e.g., ArrayRef[Int])
	baseType, params := parser.ExtractTypeAndParams(typeStr)

	// Add the base type if it's not a built-in type
	if !isBuiltInType(baseType) {
		typeInfo := typedef.TypeInfo{
			Name:        baseType,
			Description: fmt.Sprintf("Type %s extracted from code", baseType),
			Parameters:  []typedef.ParamInfo{},
		}

		// Add parameter information if this is a parameterized type
		for _, param := range params {
			typeInfo.Parameters = append(typeInfo.Parameters, typedef.ParamInfo{
				Name:        fmt.Sprintf("T%d", len(typeInfo.Parameters)+1),
				Description: fmt.Sprintf("Type parameter for %s", baseType),
				Type:        param,
			})

			// Recursively add the parameter type
			addTypeIfNew(typeDef, param, addedTypes)
		}

		typeDef.Types = append(typeDef.Types, typeInfo)
		addedTypes[typeStr] = true
	}

	// For parameterized types, make sure we add the parameterized version too
	if len(params) > 0 {
		addedTypes[typeStr] = true
	}
}

// addParameterType adds a parameter type to the appropriate sub or method
func addParameterType(typeDef *typedef.TypeDefinition, annotation *parser.TypeAnnotation) {
	// Extract the function name from the parameter name
	// This requires proper parsing of the annotation context
	// For now, we'll use a simple heuristic to get the function name

	paramName := annotation.AnnotatedItem
	functionName := extractFunctionNameFromParam(paramName)
	if functionName == "" {
		return
	}

	typeStr := annotation.TypeExpression.String()

	if annotation.Kind == parser.SubParamAnnotation {
		// Find or create the subroutine
		var subInfo *typedef.SubInfo
		for i := range typeDef.Subs {
			if typeDef.Subs[i].Name == functionName {
				subInfo = &typeDef.Subs[i]
				break
			}
		}

		if subInfo == nil {
			// Create a new subroutine entry
			typeDef.Subs = append(typeDef.Subs, typedef.SubInfo{
				Name:        functionName,
				Description: fmt.Sprintf("Subroutine %s", functionName),
				Parameters:  []typedef.ParamInfo{},
			})
			subInfo = &typeDef.Subs[len(typeDef.Subs)-1]
		}

		// Add the parameter
		subInfo.Parameters = append(subInfo.Parameters, typedef.ParamInfo{
			Name:        paramName,
			Description: fmt.Sprintf("Parameter for %s", functionName),
			Type:        typeStr,
		})
	} else {
		// Find or create the method
		var methodInfo *typedef.MethodInfo
		for i := range typeDef.Methods {
			if typeDef.Methods[i].Name == functionName {
				methodInfo = &typeDef.Methods[i]
				break
			}
		}

		if methodInfo == nil {
			// Create a new method entry
			typeDef.Methods = append(typeDef.Methods, typedef.MethodInfo{
				Name:        functionName,
				Description: fmt.Sprintf("Method %s", functionName),
				Parameters:  []typedef.ParamInfo{},
			})
			methodInfo = &typeDef.Methods[len(typeDef.Methods)-1]
		}

		// Add the parameter
		methodInfo.Parameters = append(methodInfo.Parameters, typedef.ParamInfo{
			Name:        paramName,
			Description: fmt.Sprintf("Parameter for %s", functionName),
			Type:        typeStr,
		})
	}
}

// addReturnType adds a return type to the appropriate sub or method
func addReturnType(typeDef *typedef.TypeDefinition, annotation *parser.TypeAnnotation) {
	// Extract the function name from the return annotation
	// This requires proper parsing of the annotation context
	functionName := extractFunctionNameFromReturn(annotation.AnnotatedItem)
	if functionName == "" {
		return
	}

	typeStr := annotation.TypeExpression.String()

	if annotation.Kind == parser.SubReturnAnnotation {
		// Find or create the subroutine
		var subInfo *typedef.SubInfo
		for i := range typeDef.Subs {
			if typeDef.Subs[i].Name == functionName {
				subInfo = &typeDef.Subs[i]
				break
			}
		}

		if subInfo == nil {
			// Create a new subroutine entry
			typeDef.Subs = append(typeDef.Subs, typedef.SubInfo{
				Name:        functionName,
				Description: fmt.Sprintf("Subroutine %s", functionName),
				Parameters:  []typedef.ParamInfo{},
			})
			subInfo = &typeDef.Subs[len(typeDef.Subs)-1]
		}

		// Set the return type
		subInfo.Returns = append(subInfo.Returns, typedef.ReturnInfo{
			Type:        typeStr,
			Description: fmt.Sprintf("Return value for %s", functionName),
		})
	} else {
		// Find or create the method
		var methodInfo *typedef.MethodInfo
		for i := range typeDef.Methods {
			if typeDef.Methods[i].Name == functionName {
				methodInfo = &typeDef.Methods[i]
				break
			}
		}

		if methodInfo == nil {
			// Create a new method entry
			typeDef.Methods = append(typeDef.Methods, typedef.MethodInfo{
				Name:        functionName,
				Description: fmt.Sprintf("Method %s", functionName),
				Parameters:  []typedef.ParamInfo{},
			})
			methodInfo = &typeDef.Methods[len(typeDef.Methods)-1]
		}

		// Set the return type
		methodInfo.Returns = append(methodInfo.Returns, typedef.ReturnInfo{
			Type:        typeStr,
			Description: fmt.Sprintf("Return value for %s", functionName),
		})
	}
}

// addAttributeType adds an attribute type to the type info
func addAttributeType(typeDef *typedef.TypeDefinition, annotation *parser.TypeAnnotation) {
	typeName := typeDef.Module // Default to the module name

	// Find or create the type info
	var typeInfo *typedef.TypeInfo
	for i := range typeDef.Types {
		if typeDef.Types[i].Name == typeName {
			typeInfo = &typeDef.Types[i]
			break
		}
	}

	if typeInfo == nil {
		// Create a new type entry
		typeDef.Types = append(typeDef.Types, typedef.TypeInfo{
			Name:        typeName,
			Description: fmt.Sprintf("Type %s", typeName),
			Kind:        "class",
			Properties:  []typedef.PropInfo{},
		})
		typeInfo = &typeDef.Types[len(typeDef.Types)-1]
	}

	// Add the attribute as a property
	attributeName := annotation.AnnotatedItem
	typeStr := annotation.TypeExpression.String()

	typeInfo.Properties = append(typeInfo.Properties, typedef.PropInfo{
		Name:        attributeName,
		Description: fmt.Sprintf("Attribute %s", attributeName),
		Type:        typeStr,
	})
}

// addTypeDeclaration adds a type declaration to the types list
func addTypeDeclaration(typeDef *typedef.TypeDefinition, annotation *parser.TypeAnnotation) {
	typeName := annotation.AnnotatedItem
	typeStr := annotation.TypeExpression.String()

	// Check if the type already exists
	for i := range typeDef.Types {
		if typeDef.Types[i].Name == typeName {
			// Update the existing type
			typeDef.Types[i].Description = fmt.Sprintf("Type %s declared in code", typeName)
			return
		}
	}

	// Create a new type
	typeInfo := typedef.TypeInfo{
		Name:        typeName,
		Description: fmt.Sprintf("Type %s declared in code", typeName),
		Kind:        "type",
		Parent:      typeStr, // The base type of this declaration
	}

	// Add parameter information if this is a parameterized type
	baseType, params := parser.ExtractTypeAndParams(typeName)
	if baseType != typeName {
		for _, param := range params {
			typeInfo.Parameters = append(typeInfo.Parameters, typedef.ParamInfo{
				Name:        fmt.Sprintf("T%d", len(typeInfo.Parameters)+1),
				Description: fmt.Sprintf("Type parameter for %s", baseType),
				Type:        param,
			})
		}
	}

	typeDef.Types = append(typeDef.Types, typeInfo)
}

// LoadTypeDefinition loads a type definition from the type store
func LoadTypeDefinition(moduleName string) (*typedef.TypeDefinition, error) {
	if moduleName == "" {
		return nil, errors.NewTypeError(
			"704",
			"No module name provided",
			nil,
		)
	}

	// Create a new storage instance
	storage, err := typedef.NewStorage()
	if err != nil {
		return nil, err
	}

	// Load the type definition
	return storage.Load(moduleName)
}

// GetTypeHierarchy loads a type hierarchy with all available type definitions
func GetTypeHierarchy() (*typedef.TypeHierarchy, error) {
	// Create a new storage instance
	storage, err := typedef.NewStorage()
	if err != nil {
		return nil, err
	}

	// Create a type hierarchy
	hierarchy := typedef.NewTypeHierarchy(storage)

	return hierarchy, nil
}

// ExtractTypeDefinitionsFromFile analyzes a Perl file and extracts its type annotations
// into a type definition which can be stored and used for type checking
func ExtractTypeDefinitionsFromFile(filePath string, moduleName string) (*typedef.TypeDefinition, error) {
	options := &TypeDefinitionOptions{
		ModuleName: moduleName,
		SourceFile: filePath,
		Verbose:    false,
		Save:       false,
	}

	result, err := GenerateTypeDefinition(options)
	if err != nil {
		return nil, err
	}

	if len(result.Errors) > 0 {
		// Return the first error
		return result.TypeDef, result.Errors[0]
	}

	return result.TypeDef, nil
}

// Helper functions

// extractModuleNameFromPath extracts a module name from a file path
func extractModuleNameFromPath(path string) string {
	// Simple implementation for now - this can be enhanced later
	filename := filepath.Base(path)
	moduleName := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Handle lib/Module/Name.pm style paths
	dir := filepath.Dir(path)
	if strings.Contains(dir, "lib") {
		// Try to reconstruct module name from path
		parts := strings.Split(dir, "/")
		libIdx := -1
		for i, part := range parts {
			if part == "lib" {
				libIdx = i
				break
			}
		}

		if libIdx >= 0 && libIdx < len(parts)-1 {
			// Build module name from path components after lib
			moduleComponents := parts[libIdx+1:]
			moduleComponents = append(moduleComponents, moduleName)
			moduleName = strings.Join(moduleComponents, "::")
		}
	}

	return moduleName
}

// extractFunctionNameFromParam extracts a function name from a parameter name
func extractFunctionNameFromParam(paramName string) string {
	// Simple implementation - in a real system, we would use the AST
	// to precisely determine which function this parameter belongs to

	// For now, assume parameters start with $ and the function name
	// is everything before the first _ or numeric character

	if !strings.HasPrefix(paramName, "$") {
		return ""
	}

	// Remove the $ prefix
	varName := strings.TrimPrefix(paramName, "$")

	// Find the first _ or numeric character
	for i, c := range varName {
		if c == '_' || (c >= '0' && c <= '9') {
			return varName[:i]
		}
	}

	// If no separator found, use the whole name
	return varName
}

// extractFunctionNameFromReturn extracts a function name from a return annotation
func extractFunctionNameFromReturn(returnAnnotation string) string {
	// Simple implementation - in a real system, we would use the AST
	// to precisely determine which function this return belongs to

	// For the simple case, assume the return annotation is the function name
	// In a real implementation, this would be much more complex
	return returnAnnotation
}

// isBuiltInType checks if a type is a built-in type
func isBuiltInType(typeName string) bool {
	builtInTypes := map[string]bool{
		"Any":      true,
		"Bool":     true,
		"Int":      true,
		"Num":      true,
		"Str":      true,
		"Undef":    true,
		"ArrayRef": true,
		"HashRef":  true,
		"CodeRef":  true,
		"Object":   true,
		"Maybe":    true,
	}

	return builtInTypes[typeName]
}
