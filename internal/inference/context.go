// ABOUTME: Contextual type analysis for scalar vs list context in Perl
// ABOUTME: Handles the fundamental Perl concept where expressions behave differently in different contexts

package inference

import (
	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/types"
)

// ContextType represents the evaluation context in Perl
type ContextType int

const (
	ScalarContext ContextType = iota
	ListContext
	VoidContext
	BooleanContext
	StringContext
	NumericContext
)

// String returns the string representation of context type
func (ct ContextType) String() string {
	switch ct {
	case ScalarContext:
		return "scalar"
	case ListContext:
		return "list"
	case VoidContext:
		return "void"
	case BooleanContext:
		return "boolean"
	case StringContext:
		return "string"
	case NumericContext:
		return "numeric"
	default:
		return "unknown"
	}
}

// ContextualAnalyzer handles context-sensitive type inference
type ContextualAnalyzer struct {
	// Reference to the inference engine
	engine TypeInferenceEngine

	// Stack of contexts (nested contexts)
	contextStack []ContextType

	// Current evaluation context
	currentContext ContextType

	// Map of expressions to their context-dependent types
	contextualTypes map[string]map[ContextType]types.Type
}

// NewContextualAnalyzer creates a new contextual analyzer
func NewContextualAnalyzer(engine TypeInferenceEngine) *ContextualAnalyzer {
	return &ContextualAnalyzer{
		engine:          engine,
		contextStack:    make([]ContextType, 0),
		currentContext:  ScalarContext, // Default context
		contextualTypes: make(map[string]map[ContextType]types.Type),
	}
}

// AnalyzeInContext analyzes a node in a specific context
func (ca *ContextualAnalyzer) AnalyzeInContext(node ast.Node, context ContextType, inferredAST ast.InferredAST) (types.Type, error) {
	// Push context onto stack
	ca.pushContext(context)
	defer ca.popContext()

	// Analyze the node in the given context
	return ca.analyzeNodeInCurrentContext(node, inferredAST)
}

// DetermineContext determines the evaluation context for a node based on its usage
func (ca *ContextualAnalyzer) DetermineContext(node ast.Node, parent ast.Node) ContextType {
	if parent == nil {
		return ScalarContext // Default context
	}

	switch parent.Type() {
	case "assignment":
		return ca.determineAssignmentContext(node, parent)
	case "if_statement", "while_loop", "for_loop":
		return ca.determineConditionalContext(node, parent)
	case "function_call":
		return ca.determineFunctionCallContext(node, parent)
	case "binary_expression":
		return ca.determineBinaryExpressionContext(node, parent)
	case "array_access":
		return ca.determineArrayAccessContext(node, parent)
	case "hash_access":
		return ca.determineHashAccessContext(node, parent)
	case "return_statement":
		return ca.determineReturnContext(node, parent)
	case "print_statement":
		return ListContext // print expects list context
	default:
		return ScalarContext // Default
	}
}

// Context-specific analysis methods

// analyzeNodeInCurrentContext analyzes a node in the current context
func (ca *ContextualAnalyzer) analyzeNodeInCurrentContext(node ast.Node, inferredAST ast.InferredAST) (types.Type, error) {
	switch node.Type() {
	case "variable":
		return ca.analyzeVariableInContext(node, inferredAST)
	case "function_call":
		return ca.analyzeFunctionCallInContext(node, inferredAST)
	case "array_literal":
		return ca.analyzeArrayLiteralInContext(node, inferredAST)
	case "hash_literal":
		return ca.analyzeHashLiteralInContext(node, inferredAST)
	case "literal":
		return ca.analyzeLiteralInContext(node, inferredAST)
	case "binary_expression":
		return ca.analyzeBinaryExpressionInContext(node, inferredAST)
	default:
		// For unknown node types, use basic inference
		return types.NewStrType(), nil
	}
}

// analyzeVariableInContext analyzes a variable in the current context
func (ca *ContextualAnalyzer) analyzeVariableInContext(node ast.Node, inferredAST ast.InferredAST) (types.Type, error) {
	variableName := extractVariableName(node.Text())

	// Check if we have context-specific type information
	if contextTypes, exists := ca.contextualTypes[variableName]; exists {
		if contextType, hasContext := contextTypes[ca.currentContext]; hasContext {
			return contextType, nil
		}
	}

	// Determine type based on variable sigil and context
	sigil := ""
	if len(node.Text()) > 0 {
		sigil = string(node.Text()[0])
	}

	switch sigil {
	case "$":
		// Scalar variable - always scalar in any context
		return types.NewStrType(), nil
	case "@":
		// Array variable - context-dependent
		return ca.analyzeArrayVariableInContext(variableName, inferredAST)
	case "%":
		// Hash variable - context-dependent
		return ca.analyzeHashVariableInContext(variableName, inferredAST)
	default:
		// No sigil - assume scalar
		return types.NewStrType(), nil
	}
}

// analyzeArrayVariableInContext analyzes an array variable in context
func (ca *ContextualAnalyzer) analyzeArrayVariableInContext(variableName string, inferredAST ast.InferredAST) (types.Type, error) {
	switch ca.currentContext {
	case ScalarContext:
		// Array in scalar context returns count
		return types.NewIntType(), nil
	case ListContext:
		// Array in list context returns the array elements
		// Try to determine element type from previous analysis
		elementType := ca.getArrayElementType(variableName, inferredAST)
		return types.NewArrayRefType(elementType), nil
	case BooleanContext:
		// Array in boolean context - true if has elements
		return types.NewBoolType(), nil
	default:
		// Default to array reference
		return types.NewArrayRefType(types.NewStrType()), nil
	}
}

// analyzeHashVariableInContext analyzes a hash variable in context
func (ca *ContextualAnalyzer) analyzeHashVariableInContext(variableName string, inferredAST ast.InferredAST) (types.Type, error) {
	switch ca.currentContext {
	case ScalarContext:
		// Hash in scalar context returns info about hash (bucket info)
		return types.NewStrType(), nil
	case ListContext:
		// Hash in list context returns key-value pairs
		keyType := types.NewStrType() // Keys are always strings in Perl
		valueType := ca.getHashValueType(variableName, inferredAST)
		return types.NewArrayRefType(types.NewUnionType(keyType, valueType)), nil
	case BooleanContext:
		// Hash in boolean context - true if has keys
		return types.NewBoolType(), nil
	default:
		// Default to hash reference
		return types.NewHashRefType(types.NewStrType()), nil
	}
}

// analyzeFunctionCallInContext analyzes a function call in context
func (ca *ContextualAnalyzer) analyzeFunctionCallInContext(node ast.Node, inferredAST ast.InferredAST) (types.Type, error) {
	children := node.Children()
	if len(children) == 0 {
		return types.NewStrType(), nil
	}

	functionName := extractFunctionName(children[0])

	// Check for built-in functions with context-dependent behavior
	if contextType := ca.getBuiltinContextualType(functionName); contextType != nil {
		return contextType, nil
	}

	// For user-defined functions, use function signature analysis
	// This would integrate with the function signature inferrer
	return types.NewStrType(), nil // Default
}

// analyzeArrayLiteralInContext analyzes an array literal in context
func (ca *ContextualAnalyzer) analyzeArrayLiteralInContext(node ast.Node, inferredAST ast.InferredAST) (types.Type, error) {
	switch ca.currentContext {
	case ScalarContext:
		// Array literal in scalar context returns the last element
		// Analyze the last element
		children := node.Children()
		if len(children) > 0 {
			lastElement := children[len(children)-1]
			return ca.analyzeNodeInCurrentContext(lastElement, inferredAST)
		}
		return types.NewStrType(), nil
	case ListContext:
		// Array literal in list context returns all elements
		elementTypes := make([]types.Type, 0)
		for _, child := range node.Children() {
			if child.Type() == "," {
				continue // Skip comma separators
			}
			elementType, _ := ca.analyzeNodeInCurrentContext(child, inferredAST)
			elementTypes = append(elementTypes, elementType)
		}

		if len(elementTypes) == 0 {
			return types.NewArrayRefType(types.NewStrType()), nil
		} else if len(elementTypes) == 1 {
			return types.NewArrayRefType(elementTypes[0]), nil
		} else {
			// Multiple types - create union
			unionType := types.NewUnionType(elementTypes...)
			return types.NewArrayRefType(unionType), nil
		}
	default:
		return types.NewArrayRefType(types.NewStrType()), nil
	}
}

// analyzeHashLiteralInContext analyzes a hash literal in context
func (ca *ContextualAnalyzer) analyzeHashLiteralInContext(node ast.Node, inferredAST ast.InferredAST) (types.Type, error) {
	switch ca.currentContext {
	case ScalarContext:
		// Hash literal in scalar context returns hash reference
		valueType := ca.inferHashLiteralValueType(node, inferredAST)
		return types.NewHashRefType(valueType), nil
	case ListContext:
		// Hash literal in list context returns key-value pairs
		valueType := ca.inferHashLiteralValueType(node, inferredAST)
		keyType := types.NewStrType()
		return types.NewArrayRefType(types.NewUnionType(keyType, valueType)), nil
	default:
		valueType := ca.inferHashLiteralValueType(node, inferredAST)
		return types.NewHashRefType(valueType), nil
	}
}

// analyzeLiteralInContext analyzes a literal in context
func (ca *ContextualAnalyzer) analyzeLiteralInContext(node ast.Node, inferredAST ast.InferredAST) (types.Type, error) {
	// Literals are generally context-independent, but some operations can affect them
	literalInferrer := NewLiteralInferrer()
	typeInfo := literalInferrer.InferLiteralType(node.Text())

	// Apply context-specific modifications
	switch ca.currentContext {
	case NumericContext:
		// Force numeric interpretation
		if typeInfo.Type.Equals(types.NewStrType()) {
			return types.NewNumType(), nil
		}
	case StringContext:
		// Force string interpretation
		return types.NewStrType(), nil
	case BooleanContext:
		// Force boolean interpretation
		return types.NewBoolType(), nil
	}

	return typeInfo.Type, nil
}

// analyzeBinaryExpressionInContext analyzes a binary expression in context
func (ca *ContextualAnalyzer) analyzeBinaryExpressionInContext(node ast.Node, inferredAST ast.InferredAST) (types.Type, error) {
	children := node.Children()
	if len(children) < 3 {
		return types.NewStrType(), nil
	}

	operator := children[1].Text()

	// Determine result type based on operator and context
	switch operator {
	case "+", "-", "*", "/", "%", "**":
		// Arithmetic operators always return numeric values
		return types.NewNumType(), nil
	case ".", "x":
		// String operators return strings
		return types.NewStrType(), nil
	case "==", "!=", "<", ">", "<=", ">=":
		// Numeric comparison operators return boolean
		return types.NewBoolType(), nil
	case "eq", "ne", "lt", "gt", "le", "ge":
		// String comparison operators return boolean
		return types.NewBoolType(), nil
	case "&&", "||", "and", "or":
		// Logical operators return the value of one of their operands
		// This is complex - for now, return a union of both operand types
		leftType, _ := ca.analyzeNodeInCurrentContext(children[0], inferredAST)
		rightType, _ := ca.analyzeNodeInCurrentContext(children[2], inferredAST)
		return types.NewUnionType(leftType, rightType), nil
	case "=~", "!~":
		// Regex operators return boolean
		return types.NewBoolType(), nil
	default:
		return types.NewStrType(), nil
	}
}

// Context determination methods

// determineAssignmentContext determines context for assignment operations
func (ca *ContextualAnalyzer) determineAssignmentContext(node ast.Node, parent ast.Node) ContextType {
	// Look at the left-hand side of the assignment
	children := parent.Children()
	if len(children) < 3 {
		return ScalarContext
	}

	lhs := children[0]
	operator := children[1].Text()

	// Check assignment operator
	if operator == "=" {
		// Simple assignment - context depends on LHS
		return ca.determineLHSContext(lhs)
	}

	// Compound assignment operators
	return ScalarContext // Most compound assignments are scalar
}

// determineLHSContext determines context based on left-hand side of assignment
func (ca *ContextualAnalyzer) determineLHSContext(lhs ast.Node) ContextType {
	switch lhs.Type() {
	case "variable":
		sigil := ""
		if len(lhs.Text()) > 0 {
			sigil = string(lhs.Text()[0])
		}
		switch sigil {
		case "$":
			return ScalarContext
		case "@":
			return ListContext
		case "%":
			return ListContext // Hash assignment expects list of key-value pairs
		}
	case "array_access":
		return ScalarContext // Array element assignment
	case "hash_access":
		return ScalarContext // Hash element assignment
	}
	return ScalarContext
}

// determineConditionalContext determines context for conditional expressions
func (ca *ContextualAnalyzer) determineConditionalContext(node ast.Node, parent ast.Node) ContextType {
	// Conditional expressions are evaluated in boolean context
	return BooleanContext
}

// determineFunctionCallContext determines context for function call arguments
func (ca *ContextualAnalyzer) determineFunctionCallContext(node ast.Node, parent ast.Node) ContextType {
	// This would need to look up function signatures to determine parameter contexts
	// For now, default to scalar context
	return ScalarContext
}

// determineBinaryExpressionContext determines context for binary expression operands
func (ca *ContextualAnalyzer) determineBinaryExpressionContext(node ast.Node, parent ast.Node) ContextType {
	children := parent.Children()
	if len(children) < 3 {
		return ScalarContext
	}

	operator := children[1].Text()

	switch operator {
	case "+", "-", "*", "/", "%", "**":
		return NumericContext
	case ".", "x":
		return StringContext
	case "==", "!=", "<", ">", "<=", ">=":
		return NumericContext
	case "eq", "ne", "lt", "gt", "le", "ge":
		return StringContext
	case "&&", "||", "and", "or":
		return BooleanContext
	case "=~", "!~":
		// Left operand is string, right operand is regex
		if node == children[0] {
			return StringContext
		}
		return ScalarContext
	default:
		return ScalarContext
	}
}

// determineArrayAccessContext determines context for array access
func (ca *ContextualAnalyzer) determineArrayAccessContext(node ast.Node, parent ast.Node) ContextType {
	// Array access index is always evaluated in scalar context
	return ScalarContext
}

// determineHashAccessContext determines context for hash access
func (ca *ContextualAnalyzer) determineHashAccessContext(node ast.Node, parent ast.Node) ContextType {
	// Hash access key is always evaluated in scalar context
	return ScalarContext
}

// determineReturnContext determines context for return statements
func (ca *ContextualAnalyzer) determineReturnContext(node ast.Node, parent ast.Node) ContextType {
	// Return context depends on how the function is called
	// For now, default to scalar context
	return ScalarContext
}

// Helper methods

// pushContext pushes a new context onto the stack
func (ca *ContextualAnalyzer) pushContext(context ContextType) {
	ca.contextStack = append(ca.contextStack, ca.currentContext)
	ca.currentContext = context
}

// popContext pops the current context from the stack
func (ca *ContextualAnalyzer) popContext() {
	if len(ca.contextStack) > 0 {
		ca.currentContext = ca.contextStack[len(ca.contextStack)-1]
		ca.contextStack = ca.contextStack[:len(ca.contextStack)-1]
	} else {
		ca.currentContext = ScalarContext // Default
	}
}

// getArrayElementType tries to determine the element type of an array
func (ca *ContextualAnalyzer) getArrayElementType(arrayName string, inferredAST ast.InferredAST) types.Type {
	// This would look at array assignments and usage to infer element types
	// For now, return a default type
	return types.NewStrType()
}

// getHashValueType tries to determine the value type of a hash
func (ca *ContextualAnalyzer) getHashValueType(hashName string, inferredAST ast.InferredAST) types.Type {
	// This would look at hash assignments and usage to infer value types
	// For now, return a default type
	return types.NewStrType()
}

// getBuiltinContextualType returns context-dependent types for built-in functions
func (ca *ContextualAnalyzer) getBuiltinContextualType(functionName string) types.Type {
	// Built-in functions with context-dependent behavior
	contextualBuiltins := map[string]map[ContextType]types.Type{
		"split": {
			ScalarContext: types.NewIntType(), // Number of elements
			ListContext:   types.NewArrayRefType(types.NewStrType()),
		},
		"keys": {
			ScalarContext: types.NewIntType(), // Number of keys
			ListContext:   types.NewArrayRefType(types.NewStrType()),
		},
		"values": {
			ScalarContext: types.NewIntType(), // Number of values
			ListContext:   types.NewArrayRefType(types.NewStrType()),
		},
		"each": {
			ScalarContext: types.NewStrType(),                        // Next key
			ListContext:   types.NewArrayRefType(types.NewStrType()), // Key-value pair
		},
		"grep": {
			ScalarContext: types.NewIntType(), // Number of matches
			ListContext:   types.NewArrayRefType(types.NewStrType()),
		},
		"map": {
			ScalarContext: types.NewIntType(), // Number of results
			ListContext:   types.NewArrayRefType(types.NewStrType()),
		},
		"sort": {
			ScalarContext: types.NewIntType(), // Number of elements
			ListContext:   types.NewArrayRefType(types.NewStrType()),
		},
		"reverse": {
			ScalarContext: types.NewStrType(),                        // Reversed string
			ListContext:   types.NewArrayRefType(types.NewStrType()), // Reversed list
		},
	}

	if contexts, exists := contextualBuiltins[functionName]; exists {
		if contextType, hasContext := contexts[ca.currentContext]; hasContext {
			return contextType
		}
	}

	return nil // Not a contextual built-in
}

// inferHashLiteralValueType infers the value type from a hash literal
func (ca *ContextualAnalyzer) inferHashLiteralValueType(node ast.Node, inferredAST ast.InferredAST) types.Type {
	// This would analyze the hash literal structure to determine value types
	// For now, return a default type
	return types.NewStrType()
}

// StoreContextualType stores type information for a variable in a specific context
func (ca *ContextualAnalyzer) StoreContextualType(variableName string, context ContextType, varType types.Type) {
	if _, exists := ca.contextualTypes[variableName]; !exists {
		ca.contextualTypes[variableName] = make(map[ContextType]types.Type)
	}
	ca.contextualTypes[variableName][context] = varType
}

// GetContextualType retrieves type information for a variable in a specific context
func (ca *ContextualAnalyzer) GetContextualType(variableName string, context ContextType) types.Type {
	if contexts, exists := ca.contextualTypes[variableName]; exists {
		if varType, hasContext := contexts[context]; hasContext {
			return varType
		}
	}
	return nil
}

// GetCurrentContext returns the current evaluation context
func (ca *ContextualAnalyzer) GetCurrentContext() ContextType {
	return ca.currentContext
}
