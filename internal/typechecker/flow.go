// ABOUTME: Flow-sensitive type analysis for the typechecker
// ABOUTME: Handles control flow and type refinement based on conditions

package typechecker

import (
	"fmt"
	"sort"
	"strings"

	"tamarou.com/pvm/internal/ast"
)

// ControlFlowGraph represents the control flow graph of a program
type ControlFlowGraph struct {
	// Nodes represents the basic blocks in the control flow graph
	Nodes []*BasicBlock

	// Edges represents the control flow transitions between blocks
	Edges []*FlowEdge

	// Entry is the entry point of the control flow graph
	Entry *BasicBlock

	// Exit is the exit point of the control flow graph
	Exit *BasicBlock
}

// BasicBlock represents a straight-line sequence of statements
type BasicBlock struct {
	// ID is the unique identifier for this block
	ID int

	// Statements contains the AST nodes in this block
	Statements []ast.Node

	// Predecessors are the blocks that can flow into this block
	Predecessors []*BasicBlock

	// Successors are the blocks that this block can flow into
	Successors []*BasicBlock

	// TypeState is the type state at the beginning of this block
	TypeState *TypeState

	// ExitTypeState is the type state at the end of this block
	ExitTypeState *TypeState
}

// FlowEdge represents a control flow edge between basic blocks
type FlowEdge struct {
	// From is the source block
	From *BasicBlock

	// To is the destination block
	To *BasicBlock

	// Condition is the condition that must be true for this edge to be taken
	Condition *Condition

	// EdgeType describes the type of control flow edge
	EdgeType FlowEdgeType
}

// FlowEdgeType represents the type of control flow edge
type FlowEdgeType int

const (
	// UnconditionalEdge represents an unconditional flow
	UnconditionalEdge FlowEdgeType = iota

	// ConditionalTrueEdge represents the true branch of a conditional
	ConditionalTrueEdge

	// ConditionalFalseEdge represents the false branch of a conditional
	ConditionalFalseEdge

	// LoopBackEdge represents a loop back edge
	LoopBackEdge

	// BreakEdge represents a break/last statement
	BreakEdge

	// ContinueEdge represents a continue/next statement
	ContinueEdge

	// ExceptionEdge represents an exception flow
	ExceptionEdge
)

// FlowAnalyzer performs flow-sensitive type analysis
type FlowAnalyzer struct {
	// TypeChecker is the parent type checker
	TypeChecker *TypeChecker

	// CFG is the control flow graph being analyzed
	CFG *ControlFlowGraph

	// WorkList contains blocks that need to be processed
	WorkList []*BasicBlock

	// ProcessedBlocks tracks which blocks have been processed
	ProcessedBlocks map[int]bool

	// MaxIterations prevents infinite loops in analysis
	MaxIterations int
}

// NewFlowAnalyzer creates a new flow analyzer
func NewFlowAnalyzer(tc *TypeChecker) *FlowAnalyzer {
	return &FlowAnalyzer{
		TypeChecker:     tc,
		ProcessedBlocks: make(map[int]bool),
		MaxIterations:   100, // Reasonable default to prevent infinite loops
	}
}

// performFlowSensitiveAnalysis performs flow-sensitive type analysis
func (tc *TypeChecker) performFlowSensitiveAnalysis(ast *ast.AST) []error {
	var errors []error

	// Skip if flow analysis is disabled
	if tc.TypeState == nil {
		return errors
	}

	// Create flow analyzer
	analyzer := NewFlowAnalyzer(tc)

	// Build control flow graph
	cfg, err := analyzer.buildControlFlowGraph(ast)
	if err != nil {
		errors = append(errors, err)
		return errors
	}

	// Perform data flow analysis
	flowErrors := analyzer.analyzeDataFlow(cfg)
	if len(flowErrors) > 0 {
		errors = append(errors, flowErrors...)
	}

	return errors
}

// buildControlFlowGraph constructs a control flow graph from the AST
func (fa *FlowAnalyzer) buildControlFlowGraph(ast *ast.AST) (*ControlFlowGraph, error) {
	if ast.Root == nil {
		return nil, fmt.Errorf("AST has no root node")
	}

	cfg := &ControlFlowGraph{
		Nodes: []*BasicBlock{},
		Edges: []*FlowEdge{},
	}

	// Create entry block
	entry := &BasicBlock{
		ID:            0,
		Statements:    nil,
		Predecessors:  []*BasicBlock{},
		Successors:    []*BasicBlock{},
		TypeState:     fa.TypeChecker.TypeState,
		ExitTypeState: fa.copyTypeState(fa.TypeChecker.TypeState),
	}
	cfg.Entry = entry
	cfg.Nodes = append(cfg.Nodes, entry)

	// Build CFG by traversing the AST
	currentBlock := entry
	blockID := 1

	for _, child := range ast.Root.Children() {
		newBlock, newID, err := fa.processNode(child, currentBlock, blockID, cfg)
		if err != nil {
			return nil, err
		}
		currentBlock = newBlock
		blockID = newID
	}

	// Create exit block if needed
	if currentBlock != nil {
		exit := &BasicBlock{
			ID:            blockID,
			Statements:    nil,
			Predecessors:  []*BasicBlock{currentBlock},
			Successors:    []*BasicBlock{},
			TypeState:     fa.copyTypeState(currentBlock.ExitTypeState),
			ExitTypeState: fa.copyTypeState(currentBlock.ExitTypeState),
		}
		cfg.Exit = exit
		cfg.Nodes = append(cfg.Nodes, exit)
		currentBlock.Successors = append(currentBlock.Successors, exit)
		cfg.Edges = append(cfg.Edges, &FlowEdge{
			From:     currentBlock,
			To:       exit,
			EdgeType: UnconditionalEdge,
		})
	}

	fa.CFG = cfg
	return cfg, nil
}

// processNode processes a single AST node and updates the control flow graph
func (fa *FlowAnalyzer) processNode(node ast.Node, currentBlock *BasicBlock, blockID int, cfg *ControlFlowGraph) (*BasicBlock, int, error) {
	if node == nil {
		return currentBlock, blockID, nil
	}

	// Handle different node types
	switch node.Type() {
	case "if_statement", "unless_statement":
		return fa.processConditional(node, currentBlock, blockID, cfg)
	case "while_statement", "until_statement", "for_statement", "foreach_statement":
		return fa.processLoop(node, currentBlock, blockID, cfg)
	case "given_statement":
		return fa.processGivenWhen(node, currentBlock, blockID, cfg)
	default:
		// Regular statement - add to current block
		currentBlock.Statements = append(currentBlock.Statements, node)
		// Update type state based on this statement
		fa.updateTypeStateForStatement(node, currentBlock)
		return currentBlock, blockID, nil
	}
}

// processConditional handles if/unless statements
func (fa *FlowAnalyzer) processConditional(node ast.Node, currentBlock *BasicBlock, blockID int, cfg *ControlFlowGraph) (*BasicBlock, int, error) {
	// Create condition evaluation block
	condBlock := &BasicBlock{
		ID:            blockID,
		Statements:    []ast.Node{node},
		Predecessors:  []*BasicBlock{currentBlock},
		Successors:    []*BasicBlock{},
		TypeState:     fa.copyTypeState(currentBlock.ExitTypeState),
		ExitTypeState: fa.copyTypeState(currentBlock.ExitTypeState),
	}
	cfg.Nodes = append(cfg.Nodes, condBlock)
	blockID++

	// Connect current block to condition block
	currentBlock.Successors = append(currentBlock.Successors, condBlock)
	cfg.Edges = append(cfg.Edges, &FlowEdge{
		From:     currentBlock,
		To:       condBlock,
		EdgeType: UnconditionalEdge,
	})

	// Create true branch block
	trueBlock := &BasicBlock{
		ID:            blockID,
		Statements:    nil,
		Predecessors:  []*BasicBlock{condBlock},
		Successors:    []*BasicBlock{},
		TypeState:     fa.copyTypeState(condBlock.ExitTypeState),
		ExitTypeState: fa.copyTypeState(condBlock.ExitTypeState),
	}
	cfg.Nodes = append(cfg.Nodes, trueBlock)
	blockID++

	// Apply type refinement for true branch
	fa.refineTypeForCondition(node, trueBlock, true)

	// Create false branch block
	falseBlock := &BasicBlock{
		ID:            blockID,
		Statements:    nil,
		Predecessors:  []*BasicBlock{condBlock},
		Successors:    []*BasicBlock{},
		TypeState:     fa.copyTypeState(condBlock.ExitTypeState),
		ExitTypeState: fa.copyTypeState(condBlock.ExitTypeState),
	}
	cfg.Nodes = append(cfg.Nodes, falseBlock)
	blockID++

	// Apply type refinement for false branch
	fa.refineTypeForCondition(node, falseBlock, false)

	// Connect condition to branches
	condBlock.Successors = append(condBlock.Successors, trueBlock, falseBlock)
	cfg.Edges = append(cfg.Edges,
		&FlowEdge{
			From:     condBlock,
			To:       trueBlock,
			EdgeType: ConditionalTrueEdge,
		},
		&FlowEdge{
			From:     condBlock,
			To:       falseBlock,
			EdgeType: ConditionalFalseEdge,
		},
	)

	// Create merge block for after conditional
	mergeBlock := &BasicBlock{
		ID:            blockID,
		Statements:    nil,
		Predecessors:  []*BasicBlock{trueBlock, falseBlock},
		Successors:    []*BasicBlock{},
		TypeState:     fa.mergeTypeStates(trueBlock.ExitTypeState, falseBlock.ExitTypeState),
		ExitTypeState: fa.mergeTypeStates(trueBlock.ExitTypeState, falseBlock.ExitTypeState),
	}
	cfg.Nodes = append(cfg.Nodes, mergeBlock)
	blockID++

	// Connect branches to merge
	trueBlock.Successors = append(trueBlock.Successors, mergeBlock)
	falseBlock.Successors = append(falseBlock.Successors, mergeBlock)
	cfg.Edges = append(cfg.Edges,
		&FlowEdge{
			From:     trueBlock,
			To:       mergeBlock,
			EdgeType: UnconditionalEdge,
		},
		&FlowEdge{
			From:     falseBlock,
			To:       mergeBlock,
			EdgeType: UnconditionalEdge,
		},
	)

	return mergeBlock, blockID, nil
}

// processLoop handles loop statements
func (fa *FlowAnalyzer) processLoop(node ast.Node, currentBlock *BasicBlock, blockID int, cfg *ControlFlowGraph) (*BasicBlock, int, error) {
	// Create loop header block
	headerBlock := &BasicBlock{
		ID:            blockID,
		Statements:    []ast.Node{node},
		Predecessors:  []*BasicBlock{currentBlock},
		Successors:    []*BasicBlock{},
		TypeState:     fa.copyTypeState(currentBlock.ExitTypeState),
		ExitTypeState: fa.copyTypeState(currentBlock.ExitTypeState),
	}
	cfg.Nodes = append(cfg.Nodes, headerBlock)
	blockID++

	// Connect current to header
	currentBlock.Successors = append(currentBlock.Successors, headerBlock)
	cfg.Edges = append(cfg.Edges, &FlowEdge{
		From:     currentBlock,
		To:       headerBlock,
		EdgeType: UnconditionalEdge,
	})

	// Create loop body block
	bodyBlock := &BasicBlock{
		ID:            blockID,
		Statements:    nil,
		Predecessors:  []*BasicBlock{headerBlock},
		Successors:    []*BasicBlock{},
		TypeState:     fa.copyTypeState(headerBlock.ExitTypeState),
		ExitTypeState: fa.copyTypeState(headerBlock.ExitTypeState),
	}
	cfg.Nodes = append(cfg.Nodes, bodyBlock)
	blockID++

	// Create loop exit block
	exitBlock := &BasicBlock{
		ID:            blockID,
		Statements:    nil,
		Predecessors:  []*BasicBlock{headerBlock},
		Successors:    []*BasicBlock{},
		TypeState:     fa.copyTypeState(headerBlock.ExitTypeState),
		ExitTypeState: fa.copyTypeState(headerBlock.ExitTypeState),
	}
	cfg.Nodes = append(cfg.Nodes, exitBlock)
	blockID++

	// Connect header to body and exit
	headerBlock.Successors = append(headerBlock.Successors, bodyBlock, exitBlock)
	cfg.Edges = append(cfg.Edges,
		&FlowEdge{
			From:     headerBlock,
			To:       bodyBlock,
			EdgeType: ConditionalTrueEdge,
		},
		&FlowEdge{
			From:     headerBlock,
			To:       exitBlock,
			EdgeType: ConditionalFalseEdge,
		},
	)

	// Connect body back to header (loop back edge)
	bodyBlock.Successors = append(bodyBlock.Successors, headerBlock)
	headerBlock.Predecessors = append(headerBlock.Predecessors, bodyBlock)
	cfg.Edges = append(cfg.Edges, &FlowEdge{
		From:     bodyBlock,
		To:       headerBlock,
		EdgeType: LoopBackEdge,
	})

	return exitBlock, blockID, nil
}

// processGivenWhen handles given/when statements
func (fa *FlowAnalyzer) processGivenWhen(node ast.Node, currentBlock *BasicBlock, blockID int, cfg *ControlFlowGraph) (*BasicBlock, int, error) {
	// For now, treat given/when like a simple conditional
	// In a full implementation, we'd handle multiple when clauses
	return fa.processConditional(node, currentBlock, blockID, cfg)
}

// analyzeDataFlow performs data flow analysis on the control flow graph
func (fa *FlowAnalyzer) analyzeDataFlow(cfg *ControlFlowGraph) []error {
	var errors []error

	// Initialize worklist with entry block
	fa.WorkList = []*BasicBlock{cfg.Entry}
	iterations := 0

	// Process blocks until fixpoint is reached
	for len(fa.WorkList) > 0 && iterations < fa.MaxIterations {
		block := fa.WorkList[0]
		fa.WorkList = fa.WorkList[1:]

		// Skip if already processed in this iteration
		if fa.ProcessedBlocks[block.ID] {
			continue
		}

		// Process the block
		blockErrors := fa.processBlock(block)
		if len(blockErrors) > 0 {
			errors = append(errors, blockErrors...)
		}

		// Mark as processed
		fa.ProcessedBlocks[block.ID] = true

		// Add successors to worklist if their input changed
		for _, successor := range block.Successors {
			if fa.shouldReprocess(successor) {
				fa.WorkList = append(fa.WorkList, successor)
				delete(fa.ProcessedBlocks, successor.ID)
			}
		}

		iterations++
	}

	if iterations >= fa.MaxIterations {
		errors = append(errors, fmt.Errorf("flow analysis reached maximum iterations, possible infinite loop"))
	}

	return errors
}

// processBlock processes a single basic block for data flow analysis
func (fa *FlowAnalyzer) processBlock(block *BasicBlock) []error {
	var errors []error

	// Merge input from predecessors
	if len(block.Predecessors) > 1 {
		var predecessorStates []*TypeState
		for _, pred := range block.Predecessors {
			predecessorStates = append(predecessorStates, pred.ExitTypeState)
		}
		block.TypeState = fa.mergeMultipleTypeStates(predecessorStates)
	}

	// Process statements in the block
	currentState := fa.copyTypeState(block.TypeState)
	for _, stmt := range block.Statements {
		stmtErrors := fa.processStatement(stmt, currentState)
		if len(stmtErrors) > 0 {
			errors = append(errors, stmtErrors...)
		}
	}

	// Update exit state
	block.ExitTypeState = currentState

	return errors
}

// processStatement processes a single statement and updates the type state
func (fa *FlowAnalyzer) processStatement(stmt ast.Node, state *TypeState) []error {
	var errors []error

	// Handle different statement types
	switch stmt.Type() {
	case "variable_declaration":
		errors = append(errors, fa.processVariableDeclaration(stmt, state)...)
	case "assignment":
		errors = append(errors, fa.processAssignment(stmt, state)...)
	case "function_call":
		errors = append(errors, fa.processFunctionCall(stmt, state)...)
	default:
		// For unknown statement types, we don't change the type state
		// In a full implementation, we'd handle more statement types
	}

	return errors
}

// Helper methods for type state management

// copyTypeState creates a deep copy of a type state
func (fa *FlowAnalyzer) copyTypeState(state *TypeState) *TypeState {
	if state == nil {
		return &TypeState{
			VariableTypes: make(map[string]string),
			RefinedTypes:  make(map[string]string),
			Conditions:    []Condition{},
		}
	}

	newState := &TypeState{
		VariableTypes:  make(map[string]string),
		RefinedTypes:   make(map[string]string),
		Conditions:     make([]Condition, len(state.Conditions)),
		SkipFlowChecks: state.SkipFlowChecks,
	}

	// Copy variable types
	for k, v := range state.VariableTypes {
		newState.VariableTypes[k] = v
	}

	// Copy refined types
	for k, v := range state.RefinedTypes {
		newState.RefinedTypes[k] = v
	}

	// Copy conditions
	copy(newState.Conditions, state.Conditions)

	return newState
}

// mergeTypeStates merges two type states (for control flow join points)
func (fa *FlowAnalyzer) mergeTypeStates(state1, state2 *TypeState) *TypeState {
	if state1 == nil {
		return fa.copyTypeState(state2)
	}
	if state2 == nil {
		return fa.copyTypeState(state1)
	}

	merged := &TypeState{
		VariableTypes: make(map[string]string),
		RefinedTypes:  make(map[string]string),
		Conditions:    []Condition{},
	}

	// Merge variable types - use union types if they differ
	for varName := range state1.VariableTypes {
		type1 := state1.VariableTypes[varName]
		type2, exists := state2.VariableTypes[varName]
		if exists {
			if type1 == type2 {
				merged.VariableTypes[varName] = type1
			} else {
				// Create union type
				merged.VariableTypes[varName] = fa.createUnionType(type1, type2)
			}
		} else {
			// Variable only exists in state1, include it
			merged.VariableTypes[varName] = type1
		}
	}

	// Add variables that only exist in state2
	for varName, varType := range state2.VariableTypes {
		if _, exists := state1.VariableTypes[varName]; !exists {
			merged.VariableTypes[varName] = varType
		}
	}

	return merged
}

// mergeMultipleTypeStates merges multiple type states
func (fa *FlowAnalyzer) mergeMultipleTypeStates(states []*TypeState) *TypeState {
	if len(states) == 0 {
		return &TypeState{
			VariableTypes: make(map[string]string),
			RefinedTypes:  make(map[string]string),
			Conditions:    []Condition{},
		}
	}
	if len(states) == 1 {
		return fa.copyTypeState(states[0])
	}

	result := fa.copyTypeState(states[0])
	for i := 1; i < len(states); i++ {
		result = fa.mergeTypeStates(result, states[i])
	}
	return result
}

// createUnionType creates a union type from two types
func (fa *FlowAnalyzer) createUnionType(type1, type2 string) string {
	// Simple union type creation - in a full implementation,
	// this would use the type hierarchy to create proper union types
	if type1 == type2 {
		return type1
	}
	return type1 + "|" + type2
}

// shouldReprocess determines if a block should be reprocessed
func (fa *FlowAnalyzer) shouldReprocess(block *BasicBlock) bool {
	// Simple heuristic - in a full implementation, we'd compare
	// the input type state to see if it changed
	return !fa.ProcessedBlocks[block.ID]
}

// Type refinement methods

// refineTypeForCondition refines types based on conditional expressions
func (fa *FlowAnalyzer) refineTypeForCondition(condNode ast.Node, block *BasicBlock, positive bool) {
	// Extract condition information and apply type refinement
	// This is a simplified implementation - a full version would
	// parse the condition and extract the variable and operation

	// For now, we'll implement basic 'defined' checking
	if fa.isDefinedCheck(condNode) {
		varName := fa.extractVariableFromCondition(condNode)
		if varName != "" {
			fa.TypeChecker.refineTypeAfterCondition(varName, "defined", positive)
			// Update the block's type state
			if currentType, exists := block.TypeState.VariableTypes[varName]; exists {
				if positive {
					// After defined($var), exclude Undef
					refinedType := fa.TypeChecker.excludeTypeFromUnion(currentType, "Undef")
					block.TypeState.RefinedTypes[varName] = refinedType
				} else {
					// After !defined($var), variable must be Undef
					block.TypeState.RefinedTypes[varName] = "Undef"
				}
			}
		}
	}
}

// updateTypeStateForStatement updates type state based on a statement
func (fa *FlowAnalyzer) updateTypeStateForStatement(stmt ast.Node, block *BasicBlock) {
	// Update type state based on the statement
	// This would analyze assignments, declarations, etc.
	// For now, this is a placeholder
}

// Helper methods for condition analysis

// isDefinedCheck checks if a condition is a 'defined' check
func (fa *FlowAnalyzer) isDefinedCheck(node ast.Node) bool {
	// Simple heuristic - check if the node text contains 'defined'
	return strings.Contains(node.Text(), "defined")
}

// extractVariableFromCondition extracts the variable name from a condition
func (fa *FlowAnalyzer) extractVariableFromCondition(node ast.Node) string {
	// Simple extraction - in a full implementation, this would
	// parse the condition expression properly
	text := node.Text()
	if strings.Contains(text, "defined(") {
		// Extract variable name from defined($var)
		start := strings.Index(text, "defined(") + 8
		end := strings.Index(text[start:], ")")
		if end > 0 {
			varText := strings.TrimSpace(text[start : start+end])
			// Remove $ prefix if present
			if strings.HasPrefix(varText, "$") {
				return varText[1:]
			}
			return varText
		}
	}
	return ""
}

// Statement processing methods

// processVariableDeclaration processes variable declarations
func (fa *FlowAnalyzer) processVariableDeclaration(stmt ast.Node, state *TypeState) []error {
	var errors []error

	// Cast to VarDecl node - check by node type first
	if stmt.Type() != "var_decl" {
		return errors // Not a variable declaration, skip
	}

	varDecl, ok := stmt.(*ast.VarDecl)
	if !ok {
		return errors // Type assertion failed, skip
	}

	// Initialize type maps if needed
	if state.VariableTypes == nil {
		state.VariableTypes = make(map[string]string)
	}

	// Process each variable in the declaration
	variables := varDecl.Variables()
	for _, variable := range variables {
		if variable == nil {
			continue
		}

		varName := variable.Name
		if varName == "" {
			continue
		}

		// Extract type from type annotation or infer from initializer
		var varType string

		// First, check for explicit type annotation on the declaration
		if varDecl.TypeExpr != nil {
			varType = fa.extractTypeFromTypeExpression(varDecl.TypeExpr)
		} else if varDecl.Initializer != nil {
			// If no type annotation, try to infer from initializer
			varType = fa.inferTypeFromExpression(varDecl.Initializer)
		}

		// Fall back to "Any" if we can't determine the type
		if varType == "" {
			varType = "Any"
		}

		// Update type state with the variable type
		state.VariableTypes[varName] = varType

		// Validate the type annotation if provided
		if varDecl.TypeExpr != nil && varType != "" && varType != "Any" {
			if err := fa.validateTypeAnnotation(varType, varDecl.Start()); err != nil {
				errors = append(errors, err)
			}
		}

		// If there's an initializer, check type compatibility
		if varDecl.Initializer != nil && varType != "Any" {
			initType := fa.inferTypeFromExpression(varDecl.Initializer)
			if initType != "" && initType != "Any" && initType != varType {
				if err := fa.checkTypeCompatibility(initType, varType, varDecl.Start()); err != nil {
					errors = append(errors, err)
				}
			}
		}
	}

	return errors
}

// processAssignment processes assignment statements
func (fa *FlowAnalyzer) processAssignment(stmt ast.Node, state *TypeState) []error {
	var errors []error

	// Cast to AssignmentExpr node - check by node type first
	if stmt.Type() != "assignment_expr" {
		return errors // Not an assignment expression, skip
	}

	assign, ok := stmt.(*ast.AssignmentExpr)
	if !ok {
		return errors // Type assertion failed, skip
	}

	// Initialize type maps if needed
	if state.VariableTypes == nil {
		state.VariableTypes = make(map[string]string)
	}

	// Extract target variable name from LHS
	targetVar := fa.extractVariableFromExpression(assign.Left)
	if targetVar == "" {
		// Not assigning to a simple variable (e.g., array element, method call result)
		// Still validate RHS for any type errors
		fa.inferTypeFromExpression(assign.Right)
		return errors
	}

	// Get current type of target variable from type state
	targetType := fa.getVariableTypeFromState(targetVar, state)

	// Infer type of RHS expression
	sourceType := fa.inferTypeFromExpression(assign.Right)
	if sourceType == "" {
		sourceType = "Any"
	}

	// For compound assignments (+=, -=, etc.), need special handling
	if assign.Operator != "=" {
		// For compound assignments, both operands should be compatible with the operation
		if targetType != "" && targetType != "Any" && sourceType != "Any" {
			if err := fa.checkCompoundAssignmentCompatibility(targetType, sourceType, assign.Operator, assign.Start()); err != nil {
				errors = append(errors, err)
			}
		}
		// Result type remains the same for compound assignments
		return errors
	}

	// For simple assignment (=), check type compatibility
	if targetType != "" && targetType != "Any" && sourceType != "Any" {
		if err := fa.checkTypeCompatibility(sourceType, targetType, assign.Start()); err != nil {
			errors = append(errors, err)
		}
	} else if targetType == "" || targetType == "Any" {
		// Variable not previously declared or has Any type - refine to source type
		if sourceType != "Any" {
			state.VariableTypes[targetVar] = sourceType
		} else {
			state.VariableTypes[targetVar] = "Any"
		}
	}

	return errors
}

// processFunctionCall processes function calls
func (fa *FlowAnalyzer) processFunctionCall(stmt ast.Node, state *TypeState) []error {
	var errors []error

	// Cast to CallExpr node - check by node type first
	if stmt.Type() != "call_expr" {
		return errors // Not a function call expression, skip
	}

	call, ok := stmt.(*ast.CallExpr)
	if !ok {
		return errors // Type assertion failed, skip
	}

	// Extract function name from the call expression
	functionName := fa.extractFunctionNameFromCall(call)
	if functionName == "" {
		// Can't determine function name (e.g., complex expression)
		// Still validate arguments for any type errors
		for _, arg := range call.Arguments {
			fa.inferTypeFromExpression(arg)
		}
		return errors
	}

	// Look up function signature in the type checker's function types
	signature, exists := fa.TypeChecker.FunctionTypes[functionName]
	if !exists {
		// Function signature not found - might be built-in, external, or undeclared
		// Still validate arguments for any type errors
		for _, arg := range call.Arguments {
			fa.inferTypeFromExpression(arg)
		}
		return errors
	}

	// Validate parameter count
	expectedParamCount := len(signature.ParameterTypes)
	actualParamCount := len(call.Arguments)

	if actualParamCount != expectedParamCount {
		err := &TypeCheckError{
			Message: fmt.Sprintf("Function '%s' expects %d parameters, got %d", functionName, expectedParamCount, actualParamCount),
			Path:    "",
			Line:    call.Start().Line,
			Column:  call.Start().Column,
		}
		errors = append(errors, err)
		return errors // Can't validate individual parameters if count is wrong
	}

	// Validate each parameter type
	paramNames := fa.getParameterNamesFromSignature(signature)
	for i, arg := range call.Arguments {
		if i >= len(paramNames) {
			break // Safety check
		}

		paramName := paramNames[i]
		expectedType, exists := signature.ParameterTypes[paramName]
		if !exists {
			continue // Parameter not found in signature, skip
		}

		actualType := fa.inferTypeFromExpression(arg)
		if actualType == "" || actualType == "Any" {
			continue // Can't validate unknown or Any type
		}

		// Check type compatibility
		if expectedType != "Any" && actualType != expectedType {
			if err := fa.checkTypeCompatibility(actualType, expectedType, call.Start()); err != nil {
				// Customize error message for function parameter
				paramErr := &TypeCheckError{
					Message: fmt.Sprintf("Parameter %d of function '%s' expects type '%s', got '%s'", i+1, functionName, expectedType, actualType),
					Path:    err.(*TypeCheckError).Path,
					Line:    call.Start().Line,
					Column:  call.Start().Column,
				}
				errors = append(errors, paramErr)
			}
		}
	}

	return errors
}

// AddFlowPatterns adds custom flow patterns for validation
func (tc *TypeChecker) AddFlowPatterns(patterns []string) {
	// Convert string patterns to ValidationPattern objects
	for _, pattern := range patterns {
		// Parse the pattern and create a ValidationPattern
		// For now, we'll just store the patterns as names
		validationPattern := ValidationPattern{
			Name:    pattern,
			Pattern: pattern,
			RefinementFunc: func(variable string, currentType string) string {
				// Default refinement - no change
				return currentType
			},
			Checker: func(node ast.Node) (string, bool) {
				// Default checker - matches nothing
				return "", false
			},
		}
		tc.ValidationPatterns = append(tc.ValidationPatterns, validationPattern)
	}
}

// Helper methods for flow control analysis

// extractTypeFromTypeExpression extracts a type string from a TypeExpression AST node
func (fa *FlowAnalyzer) extractTypeFromTypeExpression(typeExpr *ast.TypeExpression) string {
	if typeExpr == nil {
		return ""
	}

	// Use the existing String() method on TypeExpression
	return typeExpr.String()
}

// inferTypeFromExpression performs basic type inference on expressions
func (fa *FlowAnalyzer) inferTypeFromExpression(expr ast.ExpressionNode) string {
	if expr == nil {
		return "Any"
	}

	// Use the existing inference engine if available
	if fa.TypeChecker.InferenceEngine != nil {
		// Try to get inferred type from the inference engine
		inferredTypes := fa.TypeChecker.InferenceEngine.GetAllInferredTypes()

		// For variable expressions, check if we have an inferred type
		if varExpr, ok := expr.(*ast.VariableExpr); ok {
			if inferredType, exists := inferredTypes[varExpr.Name]; exists {
				return inferredType
			}
		}
	}

	// Basic type inference based on expression type
	switch expr.Type() {
	case "literal":
		if literal, ok := expr.(*ast.LiteralExpr); ok {
			return fa.inferTypeFromLiteral(literal)
		}
	case "variable":
		if varExpr, ok := expr.(*ast.VariableExpr); ok {
			// Check current type state first
			if fa.TypeChecker.TypeState != nil {
				if varType, exists := fa.TypeChecker.TypeState.VariableTypes[varExpr.Name]; exists {
					return varType
				}
			}
			// Fall back to type annotations
			if varType, exists := fa.TypeChecker.TypeAnnotations[varExpr.Name]; exists {
				return varType
			}
		}
	case "call_expr":
		if callExpr, ok := expr.(*ast.CallExpr); ok {
			return fa.inferReturnTypeFromCall(callExpr)
		}
	}

	return "Any" // Default fallback
}

// inferTypeFromLiteral infers type from literal expressions
func (fa *FlowAnalyzer) inferTypeFromLiteral(literal *ast.LiteralExpr) string {
	if literal == nil {
		return "Any"
	}

	switch literal.Kind {
	case ast.NumberLiteral:
		return "Num" // Could be more specific with Int vs Num analysis
	case ast.StringLiteral:
		return "Str"
	case ast.BooleanLiteral:
		return "Bool"
	case ast.UndefLiteral:
		return "Undef"
	case ast.RegexLiteral:
		return "Regex"
	default:
		return "Any"
	}
}

// inferReturnTypeFromCall infers the return type of a function call
func (fa *FlowAnalyzer) inferReturnTypeFromCall(call *ast.CallExpr) string {
	functionName := fa.extractFunctionNameFromCall(call)
	if functionName == "" {
		return "Any"
	}

	if signature, exists := fa.TypeChecker.FunctionTypes[functionName]; exists {
		return signature.ReturnType
	}

	return "Any"
}

// extractVariableFromExpression extracts variable name from an expression (for LHS of assignments)
func (fa *FlowAnalyzer) extractVariableFromExpression(expr ast.ExpressionNode) string {
	if expr == nil {
		return ""
	}

	// Handle simple variable expressions
	if varExpr, ok := expr.(*ast.VariableExpr); ok {
		return varExpr.Name
	}

	// For more complex expressions (array/hash refs, etc.), we don't extract the variable
	// This is intentional - we only want simple variable assignments
	return ""
}

// getVariableTypeFromState retrieves the current type of a variable from the type state
func (fa *FlowAnalyzer) getVariableTypeFromState(varName string, state *TypeState) string {
	if state == nil || varName == "" {
		return ""
	}

	// Check refined types first (more specific)
	if state.RefinedTypes != nil {
		if refinedType, exists := state.RefinedTypes[varName]; exists {
			return refinedType
		}
	}

	// Check base variable types
	if state.VariableTypes != nil {
		if varType, exists := state.VariableTypes[varName]; exists {
			return varType
		}
	}

	// Check type checker's type annotations as fallback
	if varType, exists := fa.TypeChecker.TypeAnnotations[varName]; exists {
		return varType
	}

	return ""
}

// extractFunctionNameFromCall extracts the function name from a CallExpr
func (fa *FlowAnalyzer) extractFunctionNameFromCall(call *ast.CallExpr) string {
	if call == nil || call.Function == nil {
		return ""
	}

	// Handle simple function name (identifier)
	if varExpr, ok := call.Function.(*ast.VariableExpr); ok {
		return varExpr.Name
	}

	// For method calls or more complex expressions, we might need more sophisticated handling
	// For now, return empty string to indicate we can't handle it
	return ""
}

// getParameterNamesFromSignature extracts ordered parameter names from a function signature
func (fa *FlowAnalyzer) getParameterNamesFromSignature(signature *FunctionSignature) []string {
	if signature == nil || signature.ParameterTypes == nil {
		return []string{}
	}

	// Since Go maps don't have guaranteed order, we need to extract parameter names
	// in a consistent way. For now, we'll extract them alphabetically.
	// In a real implementation, we'd want to maintain parameter order in the signature.
	var paramNames []string
	for paramName := range signature.ParameterTypes {
		paramNames = append(paramNames, paramName)
	}

	// Sort alphabetically for consistency (this is a limitation of the current design)
	// Ideally, FunctionSignature should maintain parameter order
	sort.Strings(paramNames)

	return paramNames
}

// validateTypeAnnotation validates that a type annotation is valid
func (fa *FlowAnalyzer) validateTypeAnnotation(typeStr string, pos ast.Position) error {
	if typeStr == "" || typeStr == "Any" {
		return nil // Empty or Any type is always valid
	}

	// Use the type hierarchy to validate the type
	if fa.TypeChecker.Hierarchy != nil {
		if err := fa.TypeChecker.Hierarchy.ValidateType(typeStr); err != nil {
			return &TypeCheckError{
				Message: fmt.Sprintf("Invalid type annotation '%s'", typeStr),
				Path:    "",
				Line:    pos.Line,
				Column:  pos.Column,
			}
		}
	}

	return nil
}

// checkTypeCompatibility checks if sourceType is compatible with targetType
func (fa *FlowAnalyzer) checkTypeCompatibility(sourceType, targetType string, pos ast.Position) error {
	if sourceType == "Any" || targetType == "Any" {
		return nil // Any type is compatible with everything
	}

	// Use the type hierarchy to check compatibility
	if fa.TypeChecker.Hierarchy != nil {
		if err := fa.TypeChecker.Hierarchy.CheckTypeCompatibility(sourceType, targetType); err != nil {
			return &TypeCheckError{
				Message: fmt.Sprintf("Cannot assign type '%s' to type '%s'", sourceType, targetType),
				Path:    "",
				Line:    pos.Line,
				Column:  pos.Column,
			}
		}
	}

	return nil
}

// checkCompoundAssignmentCompatibility checks type compatibility for compound assignments (+=, -=, etc.)
func (fa *FlowAnalyzer) checkCompoundAssignmentCompatibility(targetType, sourceType, operator string, pos ast.Position) error {
	// For compound assignments, both operands should be compatible with the operation
	switch operator {
	case "+=", "-=", "*=", "/=", "%=":
		// Numeric operations - both types should be numeric or compatible
		if !fa.isNumericType(targetType) || !fa.isNumericType(sourceType) {
			return &TypeCheckError{
				Message: fmt.Sprintf("Operator '%s' requires numeric types, got '%s' %s '%s'", operator, targetType, operator, sourceType),
				Path:    "",
				Line:    pos.Line,
				Column:  pos.Column,
			}
		}
	case ".=":
		// String concatenation - both types should be string-compatible
		if !fa.isStringCompatibleType(targetType) || !fa.isStringCompatibleType(sourceType) {
			return &TypeCheckError{
				Message: fmt.Sprintf("Operator '%s' requires string-compatible types, got '%s' %s '%s'", operator, targetType, operator, sourceType),
				Path:    "",
				Line:    pos.Line,
				Column:  pos.Column,
			}
		}
	}

	return nil
}

// isNumericType checks if a type is numeric (Int, Num, or compatible)
func (fa *FlowAnalyzer) isNumericType(typeStr string) bool {
	switch typeStr {
	case "Int", "Num", "Any":
		return true
	default:
		return false
	}
}

// isStringCompatibleType checks if a type can be used in string operations
func (fa *FlowAnalyzer) isStringCompatibleType(typeStr string) bool {
	// In Perl, most types can be stringified, but we'll be more strict
	switch typeStr {
	case "Str", "Int", "Num", "Bool", "Any":
		return true
	default:
		return false
	}
}
