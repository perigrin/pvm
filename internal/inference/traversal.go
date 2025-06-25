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
	// For now, just traverse children
	// Real implementation would analyze assignment patterns
	return at.traverseChildren(node, inferredAST)
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
