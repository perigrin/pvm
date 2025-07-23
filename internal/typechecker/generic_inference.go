// ABOUTME: Integrates generic type inference from typedef package into the type checker
// ABOUTME: Handles type variable unification and constraint solving for generic functions

package typechecker

import (
	"fmt"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/typedef"
)

// GenericInferenceEngine handles generic type parameter inference during type checking
type GenericInferenceEngine struct {
	checker *typedef.GenericTypeChecker
	// Track type variables and their bindings during inference
	typeVarBindings map[string]typedef.Type
	// Track generic function/method signatures
	genericSignatures map[string]*GenericSignature
}

// GenericSignature represents a generic function or method signature
type GenericSignature struct {
	Name           string
	TypeParameters []string
	Constraints    []*typedef.TypeConstraint
	ParameterTypes []typedef.Type
	ReturnType     typedef.Type
}

// NewGenericInferenceEngine creates a new generic inference engine
func NewGenericInferenceEngine() *GenericInferenceEngine {
	return &GenericInferenceEngine{
		checker:           typedef.NewGenericTypeChecker(),
		typeVarBindings:   make(map[string]typedef.Type),
		genericSignatures: make(map[string]*GenericSignature),
	}
}

// RegisterGenericFunction registers a generic function signature for inference
func (engine *GenericInferenceEngine) RegisterGenericFunction(node *ast.SubDecl) error {
	if node == nil {
		return fmt.Errorf("cannot register nil function")
	}

	// Extract generic signature from AST node
	signature := &GenericSignature{
		Name:           node.Name,
		TypeParameters: extractTypeParameters(node),
		Constraints:    extractConstraints(node),
	}

	// Extract parameter and return types
	if node.Signature != nil {
		signature.ParameterTypes = extractParameterTypes(node.Signature)
	}
	if node.ReturnType != nil {
		signature.ReturnType = convertASTTypeToTypedef(node.ReturnType)
	}

	engine.genericSignatures[node.Name] = signature
	return nil
}

// InferCallTypes infers type arguments for a generic function call
func (engine *GenericInferenceEngine) InferCallTypes(call *ast.CallExpr) ([]typedef.Type, error) {
	if call == nil || call.Function == nil {
		return nil, fmt.Errorf("invalid call expression")
	}

	// Get function name
	funcName := extractFunctionName(call.Function)
	signature, ok := engine.genericSignatures[funcName]
	if !ok {
		// Not a generic function
		return nil, nil
	}

	// Create generic type from signature
	genericType := &typedef.GenericType{
		Name:           signature.Name,
		TypeParameters: convertToTypedefParamsValue(signature.TypeParameters),
		Constraints:    convertConstraintsValue(signature.Constraints),
	}

	// Extract argument types from call
	argTypes := extractArgumentTypes(call.Arguments)

	// Create a usage type that represents the call pattern
	usageType := createUsageType(signature, argTypes)

	// Use the typedef inference engine
	inferredTypes, err := engine.checker.InferTypeArguments(genericType, usageType)
	if err != nil {
		return nil, fmt.Errorf("failed to infer type arguments for %s: %w", funcName, err)
	}

	// Store bindings for use in type checking
	for i, param := range signature.TypeParameters {
		if i < len(inferredTypes) {
			engine.typeVarBindings[param] = inferredTypes[i]
		}
	}

	return inferredTypes, nil
}

// ResolveTypeVariable resolves a type variable to its inferred concrete type
func (engine *GenericInferenceEngine) ResolveTypeVariable(name string) (typedef.Type, bool) {
	typ, ok := engine.typeVarBindings[name]
	return typ, ok
}

// ClearBindings clears type variable bindings (call between function checks)
func (engine *GenericInferenceEngine) ClearBindings() {
	engine.typeVarBindings = make(map[string]typedef.Type)
}

// Helper functions

func extractTypeParameters(node *ast.SubDecl) []string {
	// TODO: Extract type parameters from AST node
	// For now, return empty - will implement when AST support is complete
	return []string{}
}

func extractConstraints(node *ast.SubDecl) []*typedef.TypeConstraint {
	// TODO: Extract constraints from where clauses
	return []*typedef.TypeConstraint{}
}

func extractParameterTypes(sig *ast.MethodSignature) []typedef.Type {
	// TODO: Convert AST parameter types to typedef types
	return []typedef.Type{}
}

func convertASTTypeToTypedef(astType *ast.TypeExpression) typedef.Type {
	if astType == nil {
		return &typedef.SimpleType{Name: "Any"}
	}

	// Handle different type expression kinds
	switch astType.Kind {
	case ast.SimpleTypeKind:
		return &typedef.SimpleType{Name: astType.Name}
	case ast.ParameterizedTypeKind:
		// Convert parameterized types like ArrayRef[Int]
		base := &typedef.SimpleType{Name: astType.Name}
		var args []typedef.Type
		for _, param := range astType.Parameters {
			args = append(args, convertASTTypeToTypedef(param))
		}
		return &typedef.ParameterizedType{
			BaseType:  base,
			Arguments: args,
		}
	default:
		// Fallback for unhandled types
		return &typedef.SimpleType{Name: astType.String()}
	}
}

func extractFunctionName(funcNode ast.Node) string {
	// TODO: Extract function name from various node types
	return ""
}

func convertToTypedefParamsValue(params []string) []typedef.TypeParameter {
	var result []typedef.TypeParameter
	for _, p := range params {
		result = append(result, typedef.TypeParameter{
			Name: p,
		})
	}
	return result
}

func convertConstraintsValue(constraints []*typedef.TypeConstraint) []typedef.TypeConstraint {
	var result []typedef.TypeConstraint
	for _, c := range constraints {
		if c != nil {
			result = append(result, *c)
		}
	}
	return result
}

func extractArgumentTypes(args []ast.ExpressionNode) []typedef.Type {
	// TODO: Extract types from argument expressions
	return []typedef.Type{}
}

func createUsageType(sig *GenericSignature, argTypes []typedef.Type) typedef.Type {
	// Create a synthetic function type representing the usage
	// This helps the inference engine match patterns
	return &typedef.SimpleType{Name: "Usage"}
}
