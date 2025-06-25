// ABOUTME: Function signature inference from usage patterns
// ABOUTME: Analyzes function calls and definitions to infer parameter and return types

package inference

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/types"
)

// FunctionSignatureInferrer handles inference of function signatures
type FunctionSignatureInferrer struct {
	// Reference to the inference engine
	engine TypeInferenceEngine

	// Map of function names to their inferred signatures
	signatures map[string]*FunctionSignature

	// Map of function names to usage patterns
	usagePatterns map[string][]*FunctionUsage
}

// FunctionSignature represents an inferred function signature
type FunctionSignature struct {
	// Function name
	Name string

	// Parameter types (in order)
	ParameterTypes []types.Type

	// Parameter names (if available)
	ParameterNames []string

	// Return type
	ReturnType types.Type

	// Confidence in this signature
	Confidence float64

	// Source of the signature (definition vs usage inference)
	Source types.TypeSource

	// Whether this is a variadic function
	IsVariadic bool

	// Context where this signature was observed
	Context string
}

// FunctionUsage represents a single usage pattern of a function
type FunctionUsage struct {
	// Arguments passed to the function
	ArgumentTypes []types.Type

	// Context where the return value is used
	ReturnContext *ReturnContext

	// Location of the usage
	Location ast.Position

	// Confidence in this usage pattern
	Confidence float64
}

// ReturnContext describes how a function's return value is used
type ReturnContext struct {
	// Expected type based on usage
	ExpectedType types.Type

	// How the return value is used (assignment, condition, etc.)
	UsageType string

	// Confidence in the expected type
	Confidence float64
}

// NewFunctionSignatureInferrer creates a new function signature inferrer
func NewFunctionSignatureInferrer(engine TypeInferenceEngine) *FunctionSignatureInferrer {
	return &FunctionSignatureInferrer{
		engine:        engine,
		signatures:    make(map[string]*FunctionSignature),
		usagePatterns: make(map[string][]*FunctionUsage),
	}
}

// AnalyzeFunctionCall analyzes a function call and records usage patterns
func (fsi *FunctionSignatureInferrer) AnalyzeFunctionCall(callNode ast.Node, inferredAST ast.InferredAST) error {
	if callNode.Type() != "function_call" {
		return fmt.Errorf("expected function_call node, got %s", callNode.Type())
	}

	children := callNode.Children()
	if len(children) == 0 {
		return fmt.Errorf("function call node has no children")
	}

	// Extract function name
	functionName := extractFunctionName(children[0])
	if functionName == "" {
		return fmt.Errorf("could not extract function name")
	}

	// Extract argument types
	argumentTypes, err := fsi.extractArgumentTypes(children[1:], inferredAST)
	if err != nil {
		return fmt.Errorf("failed to extract argument types: %w", err)
	}

	// Determine return context
	returnContext := fsi.inferReturnContext(callNode, inferredAST)

	// Create usage pattern
	usage := &FunctionUsage{
		ArgumentTypes: argumentTypes,
		ReturnContext: returnContext,
		Location:      callNode.Start(),
		Confidence:    0.80, // Good confidence for direct usage observation
	}

	// Record usage pattern
	fsi.recordUsagePattern(functionName, usage)

	// Update or create function signature
	return fsi.updateFunctionSignature(functionName)
}

// AnalyzeFunctionDefinition analyzes a function definition
func (fsi *FunctionSignatureInferrer) AnalyzeFunctionDefinition(defNode ast.Node, inferredAST ast.InferredAST) error {
	if defNode.Type() != "function_definition" && defNode.Type() != "subroutine" {
		return fmt.Errorf("expected function definition node, got %s", defNode.Type())
	}

	// Extract function name and parameters
	functionName, parameterTypes, parameterNames, err := fsi.extractDefinitionInfo(defNode, inferredAST)
	if err != nil {
		return fmt.Errorf("failed to extract definition info: %w", err)
	}

	// Analyze function body for return type
	returnType, err := fsi.analyzeReturnType(defNode, inferredAST)
	if err != nil {
		return fmt.Errorf("failed to analyze return type: %w", err)
	}

	// Create signature from definition
	signature := &FunctionSignature{
		Name:           functionName,
		ParameterTypes: parameterTypes,
		ParameterNames: parameterNames,
		ReturnType:     returnType,
		Confidence:     0.95,                 // High confidence for explicit definitions
		Source:         types.SourceVariable, // Definition source
		IsVariadic:     false,                // TODO: Detect variadic functions
		Context:        "definition",
	}

	// Store signature
	fsi.signatures[functionName] = signature

	return nil
}

// GetFunctionSignature returns the inferred signature for a function
func (fsi *FunctionSignatureInferrer) GetFunctionSignature(functionName string) *FunctionSignature {
	return fsi.signatures[functionName]
}

// GetAllSignatures returns all inferred function signatures
func (fsi *FunctionSignatureInferrer) GetAllSignatures() map[string]*FunctionSignature {
	// Return a copy to prevent external modification
	result := make(map[string]*FunctionSignature)
	for name, sig := range fsi.signatures {
		result[name] = sig
	}
	return result
}

// Helper methods

// extractFunctionName extracts the function name from a function call node
func extractFunctionName(nameNode ast.Node) string {
	if nameNode == nil {
		return ""
	}

	// Handle different ways function names can appear
	nameText := nameNode.Text()

	// Remove package qualifiers (e.g., "Package::function" -> "function")
	parts := strings.Split(nameText, "::")
	if len(parts) > 1 {
		nameText = parts[len(parts)-1]
	}

	// Remove reference prefix if present (&function -> function)
	if strings.HasPrefix(nameText, "&") {
		nameText = nameText[1:]
	}

	return nameText
}

// extractArgumentTypes extracts types of arguments from function call children
func (fsi *FunctionSignatureInferrer) extractArgumentTypes(argNodes []ast.Node, inferredAST ast.InferredAST) ([]types.Type, error) {
	var argumentTypes []types.Type

	for i, argNode := range argNodes {
		// Skip parentheses and comma nodes
		if argNode.Type() == "(" || argNode.Type() == ")" || argNode.Type() == "," {
			continue
		}

		// Get type information for this argument
		argType := fsi.inferArgumentType(argNode, inferredAST)
		if argType != nil {
			argumentTypes = append(argumentTypes, argType)
		} else {
			// If we can't infer the type, use a generic type
			argumentTypes = append(argumentTypes, types.NewStrType()) // Default to Str

			fsi.engine.AddInferenceError(NewInferenceError(
				fmt.Sprintf("function_arg_%d", i),
				fmt.Sprintf("Could not infer type for function argument %d", i)))
		}
	}

	return argumentTypes, nil
}

// inferArgumentType infers the type of a function argument
func (fsi *FunctionSignatureInferrer) inferArgumentType(argNode ast.Node, inferredAST ast.InferredAST) types.Type {
	switch argNode.Type() {
	case "literal":
		// Use literal inference
		literalInferrer := NewLiteralInferrer()
		typeInfo := literalInferrer.InferLiteralType(argNode.Text())
		return typeInfo.Type

	case "variable":
		// Look up variable type in inferred AST
		nodeID := fmt.Sprintf("variable_%s_%d_%d",
			extractVariableName(argNode.Text()),
			argNode.Start().Line, argNode.Start().Column)

		if typeInfo := inferredAST.GetTypeInfo(nodeID); typeInfo != nil {
			return typeInfo.Type
		}

		// Default to string if we can't determine the type
		return types.NewStrType()

	case "function_call":
		// The return type of a nested function call
		functionName := ""
		children := argNode.Children()
		if len(children) > 0 {
			functionName = extractFunctionName(children[0])
		}

		// Look up the signature if we have it
		if sig := fsi.GetFunctionSignature(functionName); sig != nil {
			return sig.ReturnType
		}

		// Use common Perl function return types
		return fsi.inferBuiltinReturnType(functionName)

	case "array_ref", "hash_ref":
		// Reference types
		if argNode.Type() == "array_ref" {
			return types.NewArrayRefType(types.NewStrType()) // Default element type
		} else {
			return types.NewHashRefType(types.NewStrType()) // Default value type
		}

	default:
		// For other node types, analyze children or default to string
		return types.NewStrType()
	}
}

// inferReturnContext analyzes how a function's return value is used
func (fsi *FunctionSignatureInferrer) inferReturnContext(callNode ast.Node, inferredAST ast.InferredAST) *ReturnContext {
	// Look at the parent node to see how the return value is used
	// This is a simplified implementation - real implementation would
	// traverse up the AST to find the usage context

	return &ReturnContext{
		ExpectedType: types.NewStrType(), // Default expectation
		UsageType:    "unknown",
		Confidence:   0.50, // Low confidence for default context
	}
}

// recordUsagePattern records a usage pattern for a function
func (fsi *FunctionSignatureInferrer) recordUsagePattern(functionName string, usage *FunctionUsage) {
	if _, exists := fsi.usagePatterns[functionName]; !exists {
		fsi.usagePatterns[functionName] = make([]*FunctionUsage, 0)
	}

	fsi.usagePatterns[functionName] = append(fsi.usagePatterns[functionName], usage)
}

// updateFunctionSignature updates or creates a function signature based on usage patterns
func (fsi *FunctionSignatureInferrer) updateFunctionSignature(functionName string) error {
	usages := fsi.usagePatterns[functionName]
	if len(usages) == 0 {
		return nil // No usage patterns to analyze
	}

	// Check if we already have a signature from definition
	if existingSig, exists := fsi.signatures[functionName]; exists && existingSig.Source == types.SourceVariable {
		// We have a definition-based signature, don't override with usage
		return nil
	}

	// Create signature from usage patterns
	signature := fsi.createSignatureFromUsage(functionName, usages)
	fsi.signatures[functionName] = signature

	return nil
}

// createSignatureFromUsage creates a function signature from usage patterns
func (fsi *FunctionSignatureInferrer) createSignatureFromUsage(functionName string, usages []*FunctionUsage) *FunctionSignature {
	if len(usages) == 0 {
		return nil
	}

	// Find the most common parameter count
	paramCounts := make(map[int]int)
	for _, usage := range usages {
		count := len(usage.ArgumentTypes)
		paramCounts[count]++
	}

	mostCommonCount := 0
	maxOccurrences := 0
	for count, occurrences := range paramCounts {
		if occurrences > maxOccurrences {
			maxOccurrences = occurrences
			mostCommonCount = count
		}
	}

	// Build parameter types by analyzing compatible usages
	var parameterTypes []types.Type
	for i := 0; i < mostCommonCount; i++ {
		paramType := fsi.inferParameterTypeFromUsages(i, usages)
		parameterTypes = append(parameterTypes, paramType)
	}

	// Infer return type from usage contexts
	returnType := fsi.inferReturnTypeFromUsages(usages)

	// Calculate confidence based on consistency
	confidence := fsi.calculateSignatureConfidence(usages, parameterTypes, returnType)

	return &FunctionSignature{
		Name:           functionName,
		ParameterTypes: parameterTypes,
		ParameterNames: make([]string, mostCommonCount), // Empty names from usage
		ReturnType:     returnType,
		Confidence:     confidence,
		Source:         types.SourceContext,  // Usage-based inference
		IsVariadic:     len(paramCounts) > 1, // Multiple param counts suggest variadic
		Context:        "usage_analysis",
	}
}

// inferParameterTypeFromUsages infers the type of a parameter from multiple usages
func (fsi *FunctionSignatureInferrer) inferParameterTypeFromUsages(paramIndex int, usages []*FunctionUsage) types.Type {
	var parameterTypes []types.Type

	// Collect all types seen for this parameter position
	for _, usage := range usages {
		if paramIndex < len(usage.ArgumentTypes) {
			parameterTypes = append(parameterTypes, usage.ArgumentTypes[paramIndex])
		}
	}

	if len(parameterTypes) == 0 {
		return types.NewStrType() // Default
	}

	if len(parameterTypes) == 1 {
		return parameterTypes[0] // Single type
	}

	// Multiple types - check if they're all the same
	firstType := parameterTypes[0]
	allSame := true
	for _, pType := range parameterTypes[1:] {
		if !firstType.Equals(pType) {
			allSame = false
			break
		}
	}

	if allSame {
		return firstType
	}

	// Different types - create union type
	return types.NewUnionType(parameterTypes...)
}

// inferReturnTypeFromUsages infers the return type from usage contexts
func (fsi *FunctionSignatureInferrer) inferReturnTypeFromUsages(usages []*FunctionUsage) types.Type {
	var returnTypes []types.Type

	// Collect expected return types from contexts
	for _, usage := range usages {
		if usage.ReturnContext != nil && usage.ReturnContext.ExpectedType != nil {
			returnTypes = append(returnTypes, usage.ReturnContext.ExpectedType)
		}
	}

	if len(returnTypes) == 0 {
		return types.NewStrType() // Default return type
	}

	if len(returnTypes) == 1 {
		return returnTypes[0]
	}

	// Multiple expected types - create union
	return types.NewUnionType(returnTypes...)
}

// calculateSignatureConfidence calculates confidence in a signature based on usage consistency
func (fsi *FunctionSignatureInferrer) calculateSignatureConfidence(usages []*FunctionUsage, paramTypes []types.Type, returnType types.Type) float64 {
	if len(usages) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, usage := range usages {
		score := 0.0

		// Check parameter consistency
		if len(usage.ArgumentTypes) == len(paramTypes) {
			score += 0.5 // Parameter count matches

			matchingParams := 0
			for i, argType := range usage.ArgumentTypes {
				if i < len(paramTypes) && argType.CompatibleWith(paramTypes[i]) {
					matchingParams++
				}
			}

			if len(paramTypes) > 0 {
				score += 0.4 * float64(matchingParams) / float64(len(paramTypes))
			}
		}

		// Check return type consistency
		if usage.ReturnContext != nil && usage.ReturnContext.ExpectedType != nil {
			if usage.ReturnContext.ExpectedType.CompatibleWith(returnType) {
				score += 0.1
			}
		}

		totalScore += score
	}

	avgScore := totalScore / float64(len(usages))

	// Boost confidence for multiple consistent observations
	if len(usages) > 1 {
		consistencyBonus := 0.1 * float64(len(usages)-1)
		if consistencyBonus > 0.2 {
			consistencyBonus = 0.2 // Cap the bonus
		}
		avgScore += consistencyBonus
	}

	// Ensure confidence is between 0 and 1
	if avgScore > 1.0 {
		avgScore = 1.0
	}
	if avgScore < 0.0 {
		avgScore = 0.0
	}

	return avgScore
}

// extractDefinitionInfo extracts information from a function definition node
func (fsi *FunctionSignatureInferrer) extractDefinitionInfo(defNode ast.Node, inferredAST ast.InferredAST) (string, []types.Type, []string, error) {
	// This is a simplified implementation
	// Real implementation would parse the function definition syntax

	functionName := "unknown_function"
	var parameterTypes []types.Type
	var parameterNames []string

	// Try to extract function name from definition
	children := defNode.Children()
	for _, child := range children {
		if child.Type() == "identifier" {
			functionName = child.Text()
			break
		}
	}

	// TODO: Parse parameter list from definition
	// For now, return empty parameters

	return functionName, parameterTypes, parameterNames, nil
}

// analyzeReturnType analyzes a function body to infer return type
func (fsi *FunctionSignatureInferrer) analyzeReturnType(defNode ast.Node, inferredAST ast.InferredAST) (types.Type, error) {
	// This is a simplified implementation
	// Real implementation would analyze return statements in the function body

	// Look for explicit return statements
	returnTypes := fsi.findReturnStatements(defNode, inferredAST)

	if len(returnTypes) == 0 {
		// No explicit returns found - Perl functions return the last expression
		lastExprType := fsi.findLastExpressionType(defNode, inferredAST)
		if lastExprType != nil {
			return lastExprType, nil
		}
		return types.NewStrType(), nil // Default
	}

	if len(returnTypes) == 1 {
		return returnTypes[0], nil
	}

	// Multiple return types - create union
	return types.NewUnionType(returnTypes...), nil
}

// findReturnStatements finds return statements in a function and infers their types
func (fsi *FunctionSignatureInferrer) findReturnStatements(node ast.Node, inferredAST ast.InferredAST) []types.Type {
	var returnTypes []types.Type

	// Recursively search for return statements
	if node.Type() == "return_statement" {
		// Analyze the return expression
		children := node.Children()
		if len(children) > 1 {
			returnExpr := children[1] // Skip "return" keyword
			returnType := fsi.inferArgumentType(returnExpr, inferredAST)
			returnTypes = append(returnTypes, returnType)
		}
	}

	// Recursively search children
	for _, child := range node.Children() {
		childReturns := fsi.findReturnStatements(child, inferredAST)
		returnTypes = append(returnTypes, childReturns...)
	}

	return returnTypes
}

// findLastExpressionType finds the type of the last expression in a function
func (fsi *FunctionSignatureInferrer) findLastExpressionType(defNode ast.Node, inferredAST ast.InferredAST) types.Type {
	// This is a simplified implementation
	// Real implementation would find the actual last expression in the function body
	return types.NewStrType() // Default
}

// inferBuiltinReturnType provides known return types for built-in Perl functions
func (fsi *FunctionSignatureInferrer) inferBuiltinReturnType(functionName string) types.Type {
	builtinTypes := map[string]types.Type{
		"length":    types.NewIntType(),
		"substr":    types.NewStrType(),
		"index":     types.NewIntType(),
		"rindex":    types.NewIntType(),
		"sprintf":   types.NewStrType(),
		"printf":    types.NewIntType(),
		"chomp":     types.NewIntType(),
		"chop":      types.NewStrType(),
		"split":     types.NewArrayRefType(types.NewStrType()),
		"join":      types.NewStrType(),
		"grep":      types.NewArrayRefType(types.NewStrType()),
		"map":       types.NewArrayRefType(types.NewStrType()),
		"sort":      types.NewArrayRefType(types.NewStrType()),
		"reverse":   types.NewArrayRefType(types.NewStrType()),
		"keys":      types.NewArrayRefType(types.NewStrType()),
		"values":    types.NewArrayRefType(types.NewStrType()),
		"defined":   types.NewBoolType(),
		"exists":    types.NewBoolType(),
		"ref":       types.NewStrType(),
		"scalar":    types.NewStrType(),
		"wantarray": types.NewBoolType(),
	}

	if returnType, exists := builtinTypes[functionName]; exists {
		return returnType
	}

	// Default return type for unknown functions
	return types.NewStrType()
}
