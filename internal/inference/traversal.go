// ABOUTME: AST traversal system for type inference
// ABOUTME: Provides visitor pattern for systematic type analysis

package inference

import (
	"fmt"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/types"
)

// ASTTraverser handles systematic traversal of AST for type inference
type ASTTraverser struct {
	// The inference engine using this traverser
	engine TypeInferenceEngine

	// Literal inferrer for handling literal types
	literalInferrer *LiteralInferrer

	// Current context during traversal
	context *TraversalContext
}

// TraversalContext holds state during AST traversal
type TraversalContext struct {
	// Current file path
	FilePath string

	// Variable scope stack
	ScopeStack []map[string]*types.TypeInfo

	// Current confidence adjustments
	ConfidenceModifiers map[string]float64
}

// NewASTTraverser creates a new AST traverser
func NewASTTraverser(engine TypeInferenceEngine) *ASTTraverser {
	return &ASTTraverser{
		engine:          engine,
		literalInferrer: NewLiteralInferrer(),
		context: &TraversalContext{
			ScopeStack:          make([]map[string]*types.TypeInfo, 0),
			ConfidenceModifiers: make(map[string]float64),
		},
	}
}

// TraverseAndInfer performs type inference by traversing the AST
func (at *ASTTraverser) TraverseAndInfer(inputAST *ast.AST, inferredAST ast.InferredAST) error {
	// Initialize context
	at.context.FilePath = inputAST.Path
	at.pushScope()

	// Get root node for traversal
	rootNode, err := inputAST.GetRootNode()
	if err != nil {
		// If no root node, just return the empty inferred AST
		return nil
	}

	// Start traversal from root
	if rootNode != nil {
		return at.visitNode(rootNode, inferredAST)
	}

	return nil
}

// visitNode visits a single AST node and performs type inference
func (at *ASTTraverser) visitNode(node ast.Node, inferredAST ast.InferredAST) error {
	if node == nil {
		return nil
	}

	// Handle different node types
	switch node.Type() {
	case "literal":
		return at.handleLiteralNode(node, inferredAST)
	case "variable":
		return at.handleVariableNode(node, inferredAST)
	case "assignment":
		return at.handleAssignmentNode(node, inferredAST)
	case "declaration":
		return at.handleDeclarationNode(node, inferredAST)
	case "function_call":
		return at.handleFunctionCallNode(node, inferredAST)
	case "block":
		return at.handleBlockNode(node, inferredAST)
	default:
		// For unknown node types, just traverse children
		return at.traverseChildren(node, inferredAST)
	}
}

// handleLiteralNode handles literal value nodes
func (at *ASTTraverser) handleLiteralNode(node ast.Node, inferredAST ast.InferredAST) error {
	// Extract literal value from node text
	literalValue := node.Text()

	// Infer type using literal inferrer
	typeInfo := at.literalInferrer.InferLiteralType(literalValue)

	// Generate unique node ID
	nodeID := fmt.Sprintf("literal_%s_%d_%d",
		sanitizeNodeID(literalValue), node.Start().Line, node.Start().Column)

	// Attach type information
	return inferredAST.AttachTypeInfo(nodeID, typeInfo)
}

// handleVariableNode handles variable reference nodes
func (at *ASTTraverser) handleVariableNode(node ast.Node, inferredAST ast.InferredAST) error {
	variableName := extractVariableName(node.Text())

	// Look up variable type in scope stack
	typeInfo := at.lookupVariableType(variableName)

	if typeInfo != nil {
		// Variable type found in scope
		nodeID := fmt.Sprintf("variable_%s_%d_%d",
			variableName, node.Start().Line, node.Start().Column)

		// Slightly reduce confidence for variable references
		adjustedTypeInfo := types.NewTypeInfo(
			typeInfo.Type,
			typeInfo.Confidence*0.95,
			types.SourceVariable)

		return inferredAST.AttachTypeInfo(nodeID, adjustedTypeInfo)
	}

	// Variable not found - add error but continue
	at.engine.AddInferenceError(NewInferenceError(
		fmt.Sprintf("variable_%s", variableName),
		fmt.Sprintf("Unknown variable: %s", variableName)))

	return nil
}

// handleAssignmentNode handles variable assignment nodes
func (at *ASTTraverser) handleAssignmentNode(node ast.Node, inferredAST ast.InferredAST) error {
	children := node.Children()
	if len(children) < 3 {
		// Invalid assignment - just traverse children
		return at.traverseChildren(node, inferredAST)
	}

	lhs := children[0]      // Left-hand side
	operator := children[1] // Assignment operator
	rhs := children[2]      // Right-hand side

	// Handle different types of assignments
	switch operator.Text() {
	case "=":
		return at.handleSimpleAssignment(lhs, rhs, inferredAST)
	case "+=", "-=", "*=", "/=", "%=":
		return at.handleCompoundAssignment(lhs, rhs, operator.Text(), inferredAST)
	case ".=":
		return at.handleStringAssignment(lhs, rhs, inferredAST)
	default:
		// Unknown assignment operator - just traverse children
		return at.traverseChildren(node, inferredAST)
	}
}

// handleDeclarationNode handles variable declaration nodes
func (at *ASTTraverser) handleDeclarationNode(node ast.Node, inferredAST ast.InferredAST) error {
	// For now, just traverse children
	// Real implementation would add variables to scope
	return at.traverseChildren(node, inferredAST)
}

// handleFunctionCallNode handles function call nodes
func (at *ASTTraverser) handleFunctionCallNode(node ast.Node, inferredAST ast.InferredAST) error {
	// For now, just traverse children
	// Real implementation would infer return types
	return at.traverseChildren(node, inferredAST)
}

// handleBlockNode handles block statement nodes
func (at *ASTTraverser) handleBlockNode(node ast.Node, inferredAST ast.InferredAST) error {
	// Push new scope for block
	at.pushScope()
	defer at.popScope()

	// Traverse all children in the new scope
	return at.traverseChildren(node, inferredAST)
}

// traverseChildren visits all child nodes
func (at *ASTTraverser) traverseChildren(node ast.Node, inferredAST ast.InferredAST) error {
	for _, child := range node.Children() {
		if err := at.visitNode(child, inferredAST); err != nil {
			return err
		}
	}
	return nil
}

// Scope management methods

// pushScope adds a new variable scope
func (at *ASTTraverser) pushScope() {
	newScope := make(map[string]*types.TypeInfo)
	at.context.ScopeStack = append(at.context.ScopeStack, newScope)
}

// popScope removes the current variable scope
func (at *ASTTraverser) popScope() {
	if len(at.context.ScopeStack) > 0 {
		at.context.ScopeStack = at.context.ScopeStack[:len(at.context.ScopeStack)-1]
	}
}

// lookupVariableType looks up a variable type in the scope stack
func (at *ASTTraverser) lookupVariableType(variableName string) *types.TypeInfo {
	// Search from most recent scope to oldest
	for i := len(at.context.ScopeStack) - 1; i >= 0; i-- {
		scope := at.context.ScopeStack[i]
		if typeInfo, exists := scope[variableName]; exists {
			return typeInfo
		}
	}
	return nil
}

// addVariableToScope adds a variable type to the current scope
func (at *ASTTraverser) addVariableToScope(variableName string, typeInfo *types.TypeInfo) {
	if len(at.context.ScopeStack) > 0 {
		currentScope := at.context.ScopeStack[len(at.context.ScopeStack)-1]
		currentScope[variableName] = typeInfo
	}
}

// Utility functions

// sanitizeNodeID creates a safe node ID from literal values
func sanitizeNodeID(value string) string {
	// Remove quotes and special characters
	sanitized := ""
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			sanitized += string(r)
		} else {
			sanitized += "_"
		}
	}
	if len(sanitized) > 20 {
		sanitized = sanitized[:20]
	}
	return sanitized
}

// extractVariableName extracts variable name from variable node text
func extractVariableName(nodeText string) string {
	// Remove $ prefix if present
	if len(nodeText) > 0 && nodeText[0] == '$' {
		return nodeText[1:]
	}
	return nodeText
}

// Assignment handling methods

// handleSimpleAssignment handles simple assignment (=)
func (at *ASTTraverser) handleSimpleAssignment(lhs, rhs ast.Node, inferredAST ast.InferredAST) error {
	// First, analyze the right-hand side to determine its type
	if err := at.visitNode(rhs, inferredAST); err != nil {
		return err
	}

	// Infer the type of the RHS
	rhsType := at.inferNodeType(rhs, inferredAST)

	// Handle different types of left-hand side
	switch lhs.Type() {
	case "variable":
		return at.assignToVariable(lhs, rhsType, inferredAST)
	case "array_access":
		return at.assignToArrayElement(lhs, rhsType, inferredAST)
	case "hash_access":
		return at.assignToHashElement(lhs, rhsType, inferredAST)
	default:
		// For unknown LHS types, just traverse
		return at.visitNode(lhs, inferredAST)
	}
}

// handleCompoundAssignment handles compound assignments (+=, -=, etc.)
func (at *ASTTraverser) handleCompoundAssignment(lhs, rhs ast.Node, operator string, inferredAST ast.InferredAST) error {
	// Analyze both sides
	if err := at.visitNode(lhs, inferredAST); err != nil {
		return err
	}
	if err := at.visitNode(rhs, inferredAST); err != nil {
		return err
	}

	// Determine result type based on operator
	lhsType := at.inferNodeType(lhs, inferredAST)

	var resultType types.Type
	switch operator {
	case "+=", "-=", "*=", "/=", "%=":
		// Arithmetic operations result in numeric type
		resultType = types.NewNumType()
	default:
		// For unknown operators, use LHS type
		resultType = lhsType
	}

	// Assign the result type back to LHS
	if lhs.Type() == "variable" {
		return at.assignToVariable(lhs, resultType, inferredAST)
	}

	return nil
}

// handleStringAssignment handles string concatenation assignment (.=)
func (at *ASTTraverser) handleStringAssignment(lhs, rhs ast.Node, inferredAST ast.InferredAST) error {
	// Analyze both sides
	if err := at.visitNode(lhs, inferredAST); err != nil {
		return err
	}
	if err := at.visitNode(rhs, inferredAST); err != nil {
		return err
	}

	// String concatenation always results in string type
	resultType := types.NewStrType()

	// Assign string type to LHS
	if lhs.Type() == "variable" {
		return at.assignToVariable(lhs, resultType, inferredAST)
	}

	return nil
}

// assignToVariable assigns a type to a variable
func (at *ASTTraverser) assignToVariable(varNode ast.Node, varType types.Type, inferredAST ast.InferredAST) error {
	variableName := extractVariableName(varNode.Text())

	// Create type info with good confidence for assignments
	typeInfo := types.NewTypeInfo(varType, 0.90, types.SourceVariable)

	// Add to current scope
	at.addVariableToScope(variableName, typeInfo)

	// Create node ID and attach type info
	nodeID := fmt.Sprintf("variable_%s_%d_%d",
		variableName, varNode.Start().Line, varNode.Start().Column)

	return inferredAST.AttachTypeInfo(nodeID, typeInfo)
}

// assignToArrayElement assigns a type to an array element
func (at *ASTTraverser) assignToArrayElement(arrayAccess ast.Node, elementType types.Type, inferredAST ast.InferredAST) error {
	// This would update the array's element type information
	// For now, just traverse the array access
	return at.visitNode(arrayAccess, inferredAST)
}

// assignToHashElement assigns a type to a hash element
func (at *ASTTraverser) assignToHashElement(hashAccess ast.Node, valueType types.Type, inferredAST ast.InferredAST) error {
	// This would update the hash's value type information
	// For now, just traverse the hash access
	return at.visitNode(hashAccess, inferredAST)
}

// inferNodeType infers the type of a node based on previous analysis
func (at *ASTTraverser) inferNodeType(node ast.Node, inferredAST ast.InferredAST) types.Type {
	switch node.Type() {
	case "literal":
		// Use literal inference
		literalInferrer := NewLiteralInferrer()
		typeInfo := literalInferrer.InferLiteralType(node.Text())
		return typeInfo.Type

	case "variable":
		// Look up in scope stack
		variableName := extractVariableName(node.Text())
		if typeInfo := at.lookupVariableType(variableName); typeInfo != nil {
			return typeInfo.Type
		}
		// Default to string if unknown
		return types.NewStrType()

	case "function_call":
		// Use function signature inference if available
		children := node.Children()
		if len(children) > 0 {
			functionName := extractFunctionName(children[0])
			return at.inferFunctionReturnType(functionName)
		}
		return types.NewStrType()

	case "binary_expression":
		// Infer based on operator
		return at.inferBinaryExpressionType(node, inferredAST)

	case "array_literal":
		// Array literals become array references
		elementType := at.inferArrayElementType(node, inferredAST)
		return types.NewArrayRefType(elementType)

	case "hash_literal":
		// Hash literals become hash references
		valueType := at.inferHashValueType(node, inferredAST)
		return types.NewHashRefType(valueType)

	default:
		// Default to string for unknown node types
		return types.NewStrType()
	}
}

// inferFunctionReturnType infers the return type of a function call
func (at *ASTTraverser) inferFunctionReturnType(functionName string) types.Type {
	// Built-in function return types
	builtinTypes := map[string]types.Type{
		"length":  types.NewIntType(),
		"substr":  types.NewStrType(),
		"index":   types.NewIntType(),
		"sprintf": types.NewStrType(),
		"printf":  types.NewIntType(),
		"chomp":   types.NewIntType(),
		"split":   types.NewArrayRefType(types.NewStrType()),
		"join":    types.NewStrType(),
		"defined": types.NewBoolType(),
		"exists":  types.NewBoolType(),
		"ref":     types.NewStrType(),
	}

	if returnType, exists := builtinTypes[functionName]; exists {
		return returnType
	}

	// Default return type for unknown functions
	return types.NewStrType()
}

// inferBinaryExpressionType infers the type of a binary expression
func (at *ASTTraverser) inferBinaryExpressionType(node ast.Node, inferredAST ast.InferredAST) types.Type {
	children := node.Children()
	if len(children) < 3 {
		return types.NewStrType()
	}

	operator := children[1].Text()

	switch operator {
	case "+", "-", "*", "/", "%", "**":
		// Arithmetic operators return numeric values
		return types.NewNumType()
	case ".", "x":
		// String operators return strings
		return types.NewStrType()
	case "==", "!=", "<", ">", "<=", ">=", "eq", "ne", "lt", "gt", "le", "ge":
		// Comparison operators return boolean
		return types.NewBoolType()
	case "&&", "||", "and", "or":
		// Logical operators return one of their operands
		// For simplicity, analyze both and create union
		leftType := at.inferNodeType(children[0], inferredAST)
		rightType := at.inferNodeType(children[2], inferredAST)
		return types.NewUnionType(leftType, rightType)
	case "=~", "!~":
		// Regex operators return boolean
		return types.NewBoolType()
	default:
		return types.NewStrType()
	}
}

// inferArrayElementType infers the element type of an array literal
func (at *ASTTraverser) inferArrayElementType(arrayNode ast.Node, inferredAST ast.InferredAST) types.Type {
	var elementTypes []types.Type

	// Analyze each element
	for _, child := range arrayNode.Children() {
		if child.Type() == "," {
			continue // Skip comma separators
		}
		elementType := at.inferNodeType(child, inferredAST)
		elementTypes = append(elementTypes, elementType)
	}

	if len(elementTypes) == 0 {
		return types.NewStrType() // Default element type
	}

	if len(elementTypes) == 1 {
		return elementTypes[0]
	}

	// Multiple types - create union
	return types.NewUnionType(elementTypes...)
}

// inferHashValueType infers the value type of a hash literal
func (at *ASTTraverser) inferHashValueType(hashNode ast.Node, inferredAST ast.InferredAST) types.Type {
	var valueTypes []types.Type

	// Hash literals have key => value pairs
	children := hashNode.Children()
	for i := 0; i < len(children); i += 3 { // key, =>, value pattern
		if i+2 < len(children) {
			valueNode := children[i+2]
			valueType := at.inferNodeType(valueNode, inferredAST)
			valueTypes = append(valueTypes, valueType)
		}
	}

	if len(valueTypes) == 0 {
		return types.NewStrType() // Default value type
	}

	if len(valueTypes) == 1 {
		return valueTypes[0]
	}

	// Multiple types - create union
	return types.NewUnionType(valueTypes...)
}
