// ABOUTME: Flow-sensitive type analysis for the typechecker
// ABOUTME: Handles control flow and type refinement based on conditions

package typechecker

import (
	"fmt"
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
	// Extract variable name and type from declaration
	// For now, this is a placeholder
	return []error{}
}

// processAssignment processes assignment statements
func (fa *FlowAnalyzer) processAssignment(stmt ast.Node, state *TypeState) []error {
	var errors []error

	// Check type compatibility for assignments
	// This would extract the LHS and RHS types
	// For now, this is a placeholder

	return errors
}

// processFunctionCall processes function calls
func (fa *FlowAnalyzer) processFunctionCall(stmt ast.Node, state *TypeState) []error {
	var errors []error

	// Analyze function calls for type errors
	// This would check parameter types against function signatures
	// For now, this is a placeholder

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
