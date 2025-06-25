// ABOUTME: Method resolution with type information for object-oriented Perl
// ABOUTME: Handles method calls, inheritance, and object type inference

package inference

import (
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/types"
)

// MethodResolver handles method resolution for object-oriented code
type MethodResolver struct {
	// Reference to the inference engine
	engine TypeInferenceEngine

	// External hint provider for class information
	hintProvider *ExternalHintProvider

	// Map of object variables to their types
	objectTypes map[string]types.Type

	// Map of class names to their method signatures
	classMethods map[string]map[string]*MethodSignature

	// Inheritance hierarchy
	inheritance map[string][]string // class -> parent classes
}

// MethodSignature represents a method's type signature
type MethodSignature struct {
	// Method name
	Name string

	// Class that defines this method
	DefiningClass string

	// Parameter types (including self/invocant)
	ParameterTypes []types.Type

	// Return type
	ReturnType types.Type

	// Whether this is a class method (vs instance method)
	IsClassMethod bool

	// Confidence in this signature
	Confidence float64

	// Source of the signature
	Source types.TypeSource
}

// NewMethodResolver creates a new method resolver
func NewMethodResolver(engine TypeInferenceEngine, hintProvider *ExternalHintProvider) *MethodResolver {
	return &MethodResolver{
		engine:       engine,
		hintProvider: hintProvider,
		objectTypes:  make(map[string]types.Type),
		classMethods: make(map[string]map[string]*MethodSignature),
		inheritance:  make(map[string][]string),
	}
}

// AnalyzeMethodCall analyzes a method call and infers types
func (mr *MethodResolver) AnalyzeMethodCall(callNode ast.Node, inferredAST ast.InferredAST) (types.Type, error) {
	// Parse method call syntax: $object->method() or Class->method()
	children := callNode.Children()
	if len(children) < 3 {
		return types.NewStrType(), nil // Invalid method call
	}

	invocant := children[0]   // Object or class
	arrow := children[1]      // -> operator
	methodName := children[2] // Method name

	if arrow.Text() != "->" {
		return types.NewStrType(), nil // Not a method call
	}

	// Determine if this is an instance method or class method
	if invocant.Type() == "variable" {
		// Instance method call: $object->method()
		return mr.analyzeInstanceMethodCall(invocant, methodName, callNode, inferredAST)
	} else if invocant.Type() == "identifier" {
		// Class method call: Class->method()
		return mr.analyzeClassMethodCall(invocant, methodName, callNode, inferredAST)
	}

	// Unknown invocant type
	return types.NewStrType(), nil
}

// analyzeInstanceMethodCall analyzes an instance method call
func (mr *MethodResolver) analyzeInstanceMethodCall(objectVar ast.Node, methodNameNode ast.Node, callNode ast.Node, inferredAST ast.InferredAST) (types.Type, error) {
	// Get the object's type
	objectName := extractVariableName(objectVar.Text())
	objectType := mr.getObjectType(objectName)

	if objectType == nil {
		// Unknown object type - try to infer from method call
		return mr.inferObjectTypeFromMethod(objectName, methodNameNode.Text())
	}

	// Get method signature for the object's class
	className := mr.extractClassName(objectType)
	if className == "" {
		return types.NewStrType(), nil // Can't determine class
	}

	methodName := methodNameNode.Text()
	signature := mr.getMethodSignature(className, methodName)

	if signature == nil {
		// Method not found - could be inherited or dynamic
		signature = mr.findInheritedMethod(className, methodName)
	}

	if signature != nil {
		// Store that this object is of this class type
		mr.setObjectType(objectName, mr.classNameToType(signature.DefiningClass))
		return signature.ReturnType, nil
	}

	// Method not found - return generic type
	return types.NewStrType(), nil
}

// analyzeClassMethodCall analyzes a class method call
func (mr *MethodResolver) analyzeClassMethodCall(classNode ast.Node, methodNameNode ast.Node, callNode ast.Node, inferredAST ast.InferredAST) (types.Type, error) {
	className := classNode.Text()
	methodName := methodNameNode.Text()

	// Look for class method signature
	signature := mr.getClassMethodSignature(className, methodName)

	if signature == nil {
		// Check inherited class methods
		signature = mr.findInheritedClassMethod(className, methodName)
	}

	if signature != nil {
		return signature.ReturnType, nil
	}

	// Special case for constructor methods
	if mr.isConstructorMethod(methodName) {
		// Constructor returns an instance of the class
		return mr.classNameToType(className), nil
	}

	// Method not found - return generic type
	return types.NewStrType(), nil
}

// AnalyzeObjectCreation analyzes object creation (blessing, constructors)
func (mr *MethodResolver) AnalyzeObjectCreation(creationNode ast.Node, inferredAST ast.InferredAST) (types.Type, error) {
	// Handle different object creation patterns
	switch creationNode.Type() {
	case "function_call":
		// Could be a constructor call
		return mr.analyzeConstructorCall(creationNode, inferredAST)
	case "bless_expression":
		// Perl's bless operator
		return mr.analyzeBlessExpression(creationNode, inferredAST)
	default:
		return types.NewStrType(), nil
	}
}

// analyzeConstructorCall analyzes constructor function calls
func (mr *MethodResolver) analyzeConstructorCall(callNode ast.Node, inferredAST ast.InferredAST) (types.Type, error) {
	children := callNode.Children()
	if len(children) == 0 {
		return types.NewStrType(), nil
	}

	functionName := extractFunctionName(children[0])

	// Check if this looks like a constructor
	if mr.isConstructorName(functionName) {
		// Extract class name from constructor name
		className := mr.extractClassFromConstructor(functionName)
		return mr.classNameToType(className), nil
	}

	return types.NewStrType(), nil
}

// analyzeBlessExpression analyzes Perl's bless expression
func (mr *MethodResolver) analyzeBlessExpression(blessNode ast.Node, inferredAST ast.InferredAST) (types.Type, error) {
	children := blessNode.Children()
	if len(children) < 2 {
		return types.NewStrType(), nil
	}

	// bless $reference, $class
	classNode := children[1]
	className := mr.extractClassNameFromNode(classNode)

	if className != "" {
		return mr.classNameToType(className), nil
	}

	return types.NewStrType(), nil
}

// RegisterClass registers a class and its methods
func (mr *MethodResolver) RegisterClass(className string, methods map[string]*MethodSignature, parents []string) {
	mr.classMethods[className] = methods
	if len(parents) > 0 {
		mr.inheritance[className] = parents
	}
}

// SetObjectType sets the type of an object variable
func (mr *MethodResolver) SetObjectType(objectName string, objectType types.Type) {
	mr.setObjectType(objectName, objectType)
}

// GetObjectType gets the type of an object variable
func (mr *MethodResolver) GetObjectType(objectName string) types.Type {
	return mr.getObjectType(objectName)
}

// Helper methods

// getObjectType retrieves the type of an object variable
func (mr *MethodResolver) getObjectType(objectName string) types.Type {
	if objectType, exists := mr.objectTypes[objectName]; exists {
		return objectType
	}
	return nil
}

// setObjectType sets the type of an object variable
func (mr *MethodResolver) setObjectType(objectName string, objectType types.Type) {
	mr.objectTypes[objectName] = objectType
}

// getMethodSignature gets the signature of a method for a class
func (mr *MethodResolver) getMethodSignature(className, methodName string) *MethodSignature {
	if classMethods, exists := mr.classMethods[className]; exists {
		if signature, hasMethod := classMethods[methodName]; hasMethod {
			return signature
		}
	}

	// Check external hints
	if mr.hintProvider != nil {
		if methodHint := mr.hintProvider.GetMethodTypeHint(className, methodName); methodHint != nil {
			return mr.convertMethodHintToSignature(className, methodHint)
		}
	}

	return nil
}

// getClassMethodSignature gets the signature of a class method
func (mr *MethodResolver) getClassMethodSignature(className, methodName string) *MethodSignature {
	signature := mr.getMethodSignature(className, methodName)
	if signature != nil && signature.IsClassMethod {
		return signature
	}
	return nil
}

// findInheritedMethod finds a method in parent classes
func (mr *MethodResolver) findInheritedMethod(className, methodName string) *MethodSignature {
	visited := make(map[string]bool)
	return mr.findInheritedMethodRecursive(className, methodName, visited)
}

// findInheritedMethodRecursive recursively searches for inherited methods
func (mr *MethodResolver) findInheritedMethodRecursive(className, methodName string, visited map[string]bool) *MethodSignature {
	if visited[className] {
		return nil // Avoid infinite loops
	}
	visited[className] = true

	// Check parent classes
	if parents, hasParents := mr.inheritance[className]; hasParents {
		for _, parentClass := range parents {
			if signature := mr.getMethodSignature(parentClass, methodName); signature != nil {
				return signature
			}

			// Recursively check parent's parents
			if signature := mr.findInheritedMethodRecursive(parentClass, methodName, visited); signature != nil {
				return signature
			}
		}
	}

	return nil
}

// findInheritedClassMethod finds a class method in parent classes
func (mr *MethodResolver) findInheritedClassMethod(className, methodName string) *MethodSignature {
	signature := mr.findInheritedMethod(className, methodName)
	if signature != nil && signature.IsClassMethod {
		return signature
	}
	return nil
}

// extractClassName extracts class name from a type
func (mr *MethodResolver) extractClassName(objectType types.Type) string {
	// This is simplified - real implementation would handle object types better
	typeStr := objectType.String()

	// Handle basic class names
	if !strings.Contains(typeStr, "[") && !strings.Contains(typeStr, "|") {
		return typeStr
	}

	return ""
}

// classNameToType converts a class name to a type
func (mr *MethodResolver) classNameToType(className string) types.Type {
	// For now, create a basic type with the class name
	// Real implementation would have proper object types
	return &BasicObjectType{name: className}
}

// BasicObjectType represents a basic object type
type BasicObjectType struct {
	name string
}

func (bot *BasicObjectType) String() string {
	return bot.name
}

func (bot *BasicObjectType) Equals(other types.Type) bool {
	if otherObj, ok := other.(*BasicObjectType); ok {
		return bot.name == otherObj.name
	}
	return false
}

func (bot *BasicObjectType) CompatibleWith(other types.Type) bool {
	return bot.Equals(other)
}

func (bot *BasicObjectType) IsBasic() bool {
	return false
}

func (bot *BasicObjectType) IsComplex() bool {
	return true
}

// isConstructorMethod checks if a method name is a constructor
func (mr *MethodResolver) isConstructorMethod(methodName string) bool {
	constructorNames := []string{"new", "create", "build", "make", "init"}
	for _, name := range constructorNames {
		if methodName == name {
			return true
		}
	}
	return false
}

// isConstructorName checks if a function name looks like a constructor
func (mr *MethodResolver) isConstructorName(functionName string) bool {
	// Check for patterns like "Class::new" or "new Class"
	return strings.Contains(functionName, "::new") ||
		strings.HasPrefix(functionName, "new ") ||
		mr.isConstructorMethod(functionName)
}

// extractClassFromConstructor extracts class name from constructor function name
func (mr *MethodResolver) extractClassFromConstructor(constructorName string) string {
	if strings.Contains(constructorName, "::") {
		parts := strings.Split(constructorName, "::")
		if len(parts) >= 2 && parts[len(parts)-1] == "new" {
			// Remove the "new" part and rejoin the class parts
			return strings.Join(parts[:len(parts)-1], "::")
		}
		return parts[0] // Return the first part if not ending with "new"
	}

	if strings.HasPrefix(constructorName, "new ") {
		return constructorName[4:] // Remove "new " prefix
	}

	return constructorName // Assume the whole name is the class
}

// extractClassNameFromNode extracts class name from an AST node
func (mr *MethodResolver) extractClassNameFromNode(node ast.Node) string {
	switch node.Type() {
	case "literal":
		// String literal with class name
		text := node.Text()
		if strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"") {
			return text[1 : len(text)-1] // Remove quotes
		}
		if strings.HasPrefix(text, "'") && strings.HasSuffix(text, "'") {
			return text[1 : len(text)-1] // Remove quotes
		}
		return text
	case "identifier":
		return node.Text()
	case "variable":
		// Variable containing class name - would need runtime analysis
		return ""
	default:
		return ""
	}
}

// inferObjectTypeFromMethod infers object type based on method being called
func (mr *MethodResolver) inferObjectTypeFromMethod(objectName, methodName string) (types.Type, error) {
	// Search through known classes to find which ones have this method
	var possibleClasses []string

	for className := range mr.classMethods {
		if mr.getMethodSignature(className, methodName) != nil {
			possibleClasses = append(possibleClasses, className)
		}
	}

	if len(possibleClasses) == 1 {
		// Only one possible class - use it
		classType := mr.classNameToType(possibleClasses[0])
		mr.setObjectType(objectName, classType)
		return classType, nil
	}

	if len(possibleClasses) > 1 {
		// Multiple possible classes - create union type
		var classTypes []types.Type
		for _, className := range possibleClasses {
			classTypes = append(classTypes, mr.classNameToType(className))
		}
		unionType := types.NewUnionType(classTypes...)
		mr.setObjectType(objectName, unionType)
		return unionType, nil
	}

	// No known classes with this method
	return types.NewStrType(), nil
}

// convertMethodHintToSignature converts external method hint to signature
func (mr *MethodResolver) convertMethodHintToSignature(className string, hint *FunctionTypeHint) *MethodSignature {
	// Convert parameter types
	var paramTypes []types.Type
	for _, paramStr := range hint.ParameterTypes {
		paramTypes = append(paramTypes, mr.stringToType(paramStr))
	}

	// Convert return type
	returnType := mr.stringToType(hint.ReturnType)

	return &MethodSignature{
		Name:           hint.Name,
		DefiningClass:  className,
		ParameterTypes: paramTypes,
		ReturnType:     returnType,
		IsClassMethod:  false, // Assume instance method by default
		Confidence:     hint.Confidence,
		Source:         types.SourceExternal,
	}
}

// stringToType converts string type name to Type (delegate to hint provider)
func (mr *MethodResolver) stringToType(typeName string) types.Type {
	if mr.hintProvider != nil {
		return mr.hintProvider.stringToType(typeName)
	}
	return types.NewStrType() // Default
}
