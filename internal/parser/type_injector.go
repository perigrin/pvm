// ABOUTME: Type annotation injector for adding types to existing AST
// ABOUTME: Provides functions to manipulate AST nodes and add type information

package parser

import (
	"fmt"

	"tamarou.com/pvm/internal/ast"
)

// TypeInjector manages adding type annotations to AST nodes
type TypeInjector struct {
	// ast is the AST being modified
	ast *AST

	// typeMapping maps variable/function names to their types
	typeMapping map[string]string

	// inferredTypes stores automatically inferred types
	inferredTypes map[string]string
}

// NewTypeInjector creates a new type injector for an AST
func NewTypeInjector(ast *AST) *TypeInjector {
	return &TypeInjector{
		ast:           ast,
		typeMapping:   make(map[string]string),
		inferredTypes: make(map[string]string),
	}
}

// AddVariableType adds a type annotation for a variable
func (ti *TypeInjector) AddVariableType(varName, typeName string, line, column int) error {
	if varName == "" || typeName == "" {
		return fmt.Errorf("variable name and type name cannot be empty")
	}

	// Create type expression
	typeExpr, err := ParseTypeExpression(typeName, Position{
		Line:   line,
		Column: column,
	})
	if err != nil {
		return fmt.Errorf("invalid type expression '%s': %v", typeName, err)
	}

	// Create type annotation
	annotation := &TypeAnnotation{
		AnnotatedItem:  varName,
		TypeExpression: typeExpr,
		Pos: Position{
			Line:   line,
			Column: column,
		},
		Kind: VarAnnotation,
	}

	// Add to AST
	ti.ast.TypeAnnotations = append(ti.ast.TypeAnnotations, annotation)
	ti.typeMapping[varName] = typeName

	return nil
}

// AddFunctionParameterType adds a type annotation for a function parameter
func (ti *TypeInjector) AddFunctionParameterType(funcName, paramName, typeName string, line, column int) error {
	if funcName == "" || paramName == "" || typeName == "" {
		return fmt.Errorf("function name, parameter name, and type name cannot be empty")
	}

	// Create type expression
	typeExpr, err := ParseTypeExpression(typeName, Position{
		Line:   line,
		Column: column,
	})
	if err != nil {
		return fmt.Errorf("invalid type expression '%s': %v", typeName, err)
	}

	// Create type annotation
	annotation := &TypeAnnotation{
		AnnotatedItem:  fmt.Sprintf("%s_param", funcName),
		TypeExpression: typeExpr,
		Pos: Position{
			Line:   line,
			Column: column,
		},
		Kind: SubParamAnnotation,
	}

	// Add to AST
	ti.ast.TypeAnnotations = append(ti.ast.TypeAnnotations, annotation)
	ti.typeMapping[fmt.Sprintf("%s::%s", funcName, paramName)] = typeName

	return nil
}

// AddFunctionReturnType adds a type annotation for a function return type
func (ti *TypeInjector) AddFunctionReturnType(funcName, typeName string, line, column int) error {
	if funcName == "" || typeName == "" {
		return fmt.Errorf("function name and type name cannot be empty")
	}

	// Create type expression
	typeExpr, err := ParseTypeExpression(typeName, Position{
		Line:   line,
		Column: column,
	})
	if err != nil {
		return fmt.Errorf("invalid type expression '%s': %v", typeName, err)
	}

	// Create type annotation
	annotation := &TypeAnnotation{
		AnnotatedItem:  fmt.Sprintf("%s_return", funcName),
		TypeExpression: typeExpr,
		Pos: Position{
			Line:   line,
			Column: column,
		},
		Kind: SubReturnAnnotation,
	}

	// Add to AST
	ti.ast.TypeAnnotations = append(ti.ast.TypeAnnotations, annotation)
	ti.typeMapping[fmt.Sprintf("%s::return", funcName)] = typeName

	return nil
}

// AddMethodParameterType adds a type annotation for a method parameter
func (ti *TypeInjector) AddMethodParameterType(methodName, paramName, typeName string, line, column int) error {
	if methodName == "" || paramName == "" || typeName == "" {
		return fmt.Errorf("method name, parameter name, and type name cannot be empty")
	}

	// Create type expression
	typeExpr, err := ParseTypeExpression(typeName, Position{
		Line:   line,
		Column: column,
	})
	if err != nil {
		return fmt.Errorf("invalid type expression '%s': %v", typeName, err)
	}

	// Create type annotation
	annotation := &TypeAnnotation{
		AnnotatedItem:  fmt.Sprintf("%s_param", methodName),
		TypeExpression: typeExpr,
		Pos: Position{
			Line:   line,
			Column: column,
		},
		Kind: MethodParamAnnotation,
	}

	// Add to AST
	ti.ast.TypeAnnotations = append(ti.ast.TypeAnnotations, annotation)
	ti.typeMapping[fmt.Sprintf("%s::%s", methodName, paramName)] = typeName

	return nil
}

// AddMethodReturnType adds a type annotation for a method return type
func (ti *TypeInjector) AddMethodReturnType(methodName, typeName string, line, column int) error {
	if methodName == "" || typeName == "" {
		return fmt.Errorf("method name and type name cannot be empty")
	}

	// Create type expression
	typeExpr, err := ParseTypeExpression(typeName, Position{
		Line:   line,
		Column: column,
	})
	if err != nil {
		return fmt.Errorf("invalid type expression '%s': %v", typeName, err)
	}

	// Create type annotation
	annotation := &TypeAnnotation{
		AnnotatedItem:  fmt.Sprintf("%s_return", methodName),
		TypeExpression: typeExpr,
		Pos: Position{
			Line:   line,
			Column: column,
		},
		Kind: MethodReturnAnnotation,
	}

	// Add to AST
	ti.ast.TypeAnnotations = append(ti.ast.TypeAnnotations, annotation)
	ti.typeMapping[fmt.Sprintf("%s::return", methodName)] = typeName

	return nil
}

// AddAttributeType adds a type annotation for a class attribute
func (ti *TypeInjector) AddAttributeType(attrName, typeName string, line, column int) error {
	if attrName == "" || typeName == "" {
		return fmt.Errorf("attribute name and type name cannot be empty")
	}

	// Create type expression
	typeExpr, err := ParseTypeExpression(typeName, Position{
		Line:   line,
		Column: column,
	})
	if err != nil {
		return fmt.Errorf("invalid type expression '%s': %v", typeName, err)
	}

	// Create type annotation
	annotation := &TypeAnnotation{
		AnnotatedItem:  attrName,
		TypeExpression: typeExpr,
		Pos: Position{
			Line:   line,
			Column: column,
		},
		Kind: AttrAnnotation,
	}

	// Add to AST
	ti.ast.TypeAnnotations = append(ti.ast.TypeAnnotations, annotation)
	ti.typeMapping[attrName] = typeName

	return nil
}

// RemoveTypeAnnotation removes a type annotation by annotation item name
func (ti *TypeInjector) RemoveTypeAnnotation(annotatedItem string) bool {
	for i, annotation := range ti.ast.TypeAnnotations {
		if annotation.AnnotatedItem == annotatedItem {
			// Remove from slice
			ti.ast.TypeAnnotations = append(
				ti.ast.TypeAnnotations[:i],
				ti.ast.TypeAnnotations[i+1:]...,
			)
			// Remove from mapping
			delete(ti.typeMapping, annotatedItem)
			return true
		}
	}
	return false
}

// UpdateTypeAnnotation updates an existing type annotation
func (ti *TypeInjector) UpdateTypeAnnotation(annotatedItem, newTypeName string) error {
	for _, annotation := range ti.ast.TypeAnnotations {
		if annotation.AnnotatedItem == annotatedItem {
			// Parse new type expression
			typeExpr, err := ParseTypeExpression(newTypeName, annotation.Pos)
			if err != nil {
				return fmt.Errorf("invalid type expression '%s': %v", newTypeName, err)
			}

			// Update the annotation
			annotation.TypeExpression = typeExpr
			ti.typeMapping[annotatedItem] = newTypeName
			return nil
		}
	}
	return fmt.Errorf("type annotation for '%s' not found", annotatedItem)
}

// GetTypeAnnotations returns all type annotations in the AST
func (ti *TypeInjector) GetTypeAnnotations() []*TypeAnnotation {
	return ti.ast.TypeAnnotations
}

// GetVariableType returns the type of a variable if it has a type annotation
func (ti *TypeInjector) GetVariableType(varName string) (string, bool) {
	typeName, exists := ti.typeMapping[varName]
	return typeName, exists
}

// GetFunctionParameterType returns the type of a function parameter
func (ti *TypeInjector) GetFunctionParameterType(funcName, paramName string) (string, bool) {
	key := fmt.Sprintf("%s::%s", funcName, paramName)
	typeName, exists := ti.typeMapping[key]
	return typeName, exists
}

// GetFunctionReturnType returns the return type of a function
func (ti *TypeInjector) GetFunctionReturnType(funcName string) (string, bool) {
	key := fmt.Sprintf("%s::return", funcName)
	typeName, exists := ti.typeMapping[key]
	return typeName, exists
}

// InferTypes performs basic type inference on the AST
func (ti *TypeInjector) InferTypes() error {
	// This is a basic type inference implementation
	// In a full implementation, this would analyze the AST and infer types
	// based on usage patterns, assignments, and function calls

	// For now, we'll do basic inference based on literal values
	// This would be expanded significantly in a production implementation

	// Clear previous inferred types
	ti.inferredTypes = make(map[string]string)

	// Add basic inference logic here
	// For example:
	// - String literals → Str
	// - Numeric literals → Int or Num
	// - Array references → ArrayRef
	// - Hash references → HashRef

	return nil
}

// GetInferredType returns an inferred type for a variable
func (ti *TypeInjector) GetInferredType(varName string) (string, bool) {
	typeName, exists := ti.inferredTypes[varName]
	return typeName, exists
}

// ApplyTypeMapping applies a mapping of variables to types
func (ti *TypeInjector) ApplyTypeMapping(mapping map[string]string) error {
	for varName, typeName := range mapping {
		// Try to determine the appropriate line/column
		// In a real implementation, this would use the AST to find the actual position
		line, column := ti.findVariablePosition(varName)

		err := ti.AddVariableType(varName, typeName, line, column)
		if err != nil {
			return fmt.Errorf("failed to add type for variable '%s': %v", varName, err)
		}
	}
	return nil
}

// findVariablePosition finds the position of a variable in the AST
func (ti *TypeInjector) findVariablePosition(varName string) (int, int) {
	// This is a placeholder implementation
	// In a real implementation, this would traverse the AST to find the variable
	// and return its actual position

	// For now, return a default position
	return 1, 1
}

// ValidateTypeAnnotations validates all type annotations in the AST
func (ti *TypeInjector) ValidateTypeAnnotations() []error {
	var errors []error

	for _, annotation := range ti.ast.TypeAnnotations {
		if annotation.TypeExpression == nil {
			errors = append(errors, fmt.Errorf("nil type expression for %s", annotation.AnnotatedItem))
			continue
		}

		if annotation.TypeExpression.Name == "" {
			errors = append(errors, fmt.Errorf("empty type name for %s", annotation.AnnotatedItem))
			continue
		}

		// Validate type name format
		if !ti.isValidTypeName(annotation.TypeExpression.Name) {
			errors = append(errors, fmt.Errorf("invalid type name '%s' for %s",
				annotation.TypeExpression.Name, annotation.AnnotatedItem))
		}
	}

	return errors
}

// isValidTypeName checks if a type name is valid
func (ti *TypeInjector) isValidTypeName(typeName string) bool {
	if typeName == "" {
		return false
	}

	// Basic validation - starts with letter, contains only alphanumeric chars and underscore
	if (typeName[0] < 'A' || typeName[0] > 'Z') && (typeName[0] < 'a' || typeName[0] > 'z') {
		return false
	}

	for _, char := range typeName[1:] {
		if (char < 'A' || char > 'Z') && (char < 'a' || char > 'z') &&
			(char < '0' || char > '9') && char != '_' && char != '[' && char != ']' && char != '|' && char != '&' {
			return false
		}
	}

	return true
}

// GetAST returns the modified AST
func (ti *TypeInjector) GetAST() *AST {
	return ti.ast
}

// GetTypeMapping returns the current type mapping
func (ti *TypeInjector) GetTypeMapping() map[string]string {
	// Return a copy to prevent external modification
	result := make(map[string]string)
	for k, v := range ti.typeMapping {
		result[k] = v
	}
	return result
}

// ClearTypeAnnotations removes all type annotations from the AST
func (ti *TypeInjector) ClearTypeAnnotations() {
	ti.ast.TypeAnnotations = nil
	ti.typeMapping = make(map[string]string)
	ti.inferredTypes = make(map[string]string)
}

// Clone creates a deep copy of the AST with type annotations
func (ti *TypeInjector) Clone() (*AST, error) {
	newAST := &AST{
		Path:   ti.ast.Path,
		Root:   ti.ast.Root, // Note: This is a shallow copy of the root node
		Errors: make([]error, len(ti.ast.Errors)),
	}

	// Copy errors
	copy(newAST.Errors, ti.ast.Errors)

	// Deep copy type annotations
	newAST.TypeAnnotations = make([]*TypeAnnotation, len(ti.ast.TypeAnnotations))
	for i, annotation := range ti.ast.TypeAnnotations {
		newAST.TypeAnnotations[i] = &TypeAnnotation{
			AnnotatedItem: annotation.AnnotatedItem,
			TypeExpression: func() *TypeExpression {
				te := ast.NewTypeExpression(
					annotation.TypeExpression.Name,
					annotation.TypeExpression.Parameters,
					annotation.TypeExpression.Start(),
					annotation.TypeExpression.End(),
				)
				te.IsUnion = annotation.TypeExpression.IsUnion
				te.IsIntersection = annotation.TypeExpression.IsIntersection
				te.IsNegation = annotation.TypeExpression.IsNegation
				te.UnionTypes = annotation.TypeExpression.UnionTypes
				te.IntersectionTypes = annotation.TypeExpression.IntersectionTypes
				te.NegatedType = annotation.TypeExpression.NegatedType
				te.OriginalString = annotation.TypeExpression.OriginalString
				return te
			}(),
			Pos:  annotation.Pos,
			Kind: annotation.Kind,
		}
	}

	return newAST, nil
}
