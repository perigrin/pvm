// ABOUTME: AST node factory with object pooling following Microsoft TypeScript-Go patterns
// ABOUTME: Provides pooled allocation for all major AST node types with lifecycle hooks

package ast

import (
	"sync"
	"sync/atomic"

	"tamarou.com/pvm/internal/core"
)

// NodeFactory provides pooled allocation for AST nodes following TypeScript-Go patterns
type NodeFactory struct {
	hooks NodeFactoryHooks

	// Pools for major node types
	baseNodePool       core.Pool[BaseNode]
	literalExprPool    core.Pool[LiteralExpr]
	variableExprPool   core.Pool[VariableExpr]
	binaryExprPool     core.Pool[BinaryExpr]
	callExprPool       core.Pool[CallExpr]
	unaryExprPool      core.Pool[UnaryExpr]
	assignmentExprPool core.Pool[AssignmentExpr]

	// Statement pools
	expressionStmtPool core.Pool[ExpressionStmt]
	varDeclPool        core.Pool[VarDecl]
	ifStmtPool         core.Pool[IfStmt]
	whileStmtPool      core.Pool[WhileStmt]
	forStmtPool        core.Pool[ForStmt]
	blockStmtPool      core.Pool[BlockStmt]
	subDeclPool        core.Pool[SubDecl]

	// Type expression pools
	typeExprPool         core.Pool[TypeExpression]
	unionTypePool        core.Pool[UnionType]
	intersectionTypePool core.Pool[IntersectionType]

	// Slice pools for collections
	nodeSlicePool *core.Pool[[]Node]
	varSlicePool  *core.Pool[[]*VariableExpr]
	stmtSlicePool *core.Pool[[]StatementNode]

	// Statistics
	nodeCount    int64
	poolHits     int64
	poolMisses   int64
	memoryReused int64

	mu sync.RWMutex
}

// NodeFactoryHooks provides lifecycle hooks for debugging and monitoring
type NodeFactoryHooks struct {
	OnCreate func(node Node)                // Called when a node is created
	OnUpdate func(node Node, original Node) // Called when a node is updated
	OnClone  func(node Node, original Node) // Called when a node is cloned
	OnReset  func(node Node)                // Called when a node is reset for pooling
}

// NodeFactoryCoercible allows types to provide a NodeFactory
type NodeFactoryCoercible interface {
	AsNodeFactory() *NodeFactory
}

// NewNodeFactory creates a new node factory with the given hooks
func NewNodeFactory(hooks NodeFactoryHooks) *NodeFactory {
	factory := &NodeFactory{
		hooks: hooks,
	}

	// Initialize slice pools
	factory.nodeSlicePool = &core.Pool[[]Node]{}
	factory.varSlicePool = &core.Pool[[]*VariableExpr]{}
	factory.stmtSlicePool = &core.Pool[[]StatementNode]{}

	// Register with global pool manager for monitoring
	core.RegisterGlobalPool("ast-factory", factory)

	return factory
}

// AsNodeFactory implements NodeFactoryCoercible
func (f *NodeFactory) AsNodeFactory() *NodeFactory {
	return f
}

// Stats returns pool allocation statistics
func (f *NodeFactory) Stats() core.PoolStats {
	return core.PoolStats{
		Allocations: atomic.LoadInt64(&f.nodeCount),
		Grows:       atomic.LoadInt64(&f.poolHits),
		TotalSize:   atomic.LoadInt64(&f.poolMisses),
		CurrentSize: atomic.LoadInt64(&f.memoryReused),
		Capacity:    0, // Not applicable for factory
	}
}

// NodeCount returns the total number of nodes created
func (f *NodeFactory) NodeCount() int64 {
	return atomic.LoadInt64(&f.nodeCount)
}

// PoolEfficiency returns the pool hit rate as a percentage
func (f *NodeFactory) PoolEfficiency() float64 {
	hits := atomic.LoadInt64(&f.poolHits)
	misses := atomic.LoadInt64(&f.poolMisses)
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

// Expression creation methods

// NewLiteralExpr creates a pooled literal expression
func (f *NodeFactory) NewLiteralExpr(value string, kind LiteralKind, start, end Position) *LiteralExpr {
	expr := f.literalExprPool.New()

	// Reset/initialize the pooled object
	if expr.BaseNode == nil {
		expr.BaseNode = f.baseNodePool.New()
		atomic.AddInt64(&f.poolMisses, 1)
	} else {
		f.resetBaseNode(expr.BaseNode)
		atomic.AddInt64(&f.poolHits, 1)
	}

	// Initialize fields
	expr.BaseNode.nodeType = "literal"
	expr.BaseNode.start = start
	expr.BaseNode.end = end
	expr.Value = value
	expr.Kind = kind

	atomic.AddInt64(&f.nodeCount, 1)

	if f.hooks.OnCreate != nil {
		f.hooks.OnCreate(expr)
	}

	return expr
}

// NewVariableExpr creates a pooled variable expression
func (f *NodeFactory) NewVariableExpr(name, sigil string, start, end Position) *VariableExpr {
	expr := f.variableExprPool.New()

	if expr.BaseNode == nil {
		expr.BaseNode = f.baseNodePool.New()
		atomic.AddInt64(&f.poolMisses, 1)
	} else {
		f.resetBaseNode(expr.BaseNode)
		atomic.AddInt64(&f.poolHits, 1)
	}

	expr.BaseNode.nodeType = "variable"
	expr.BaseNode.start = start
	expr.BaseNode.end = end
	expr.Name = name
	expr.Sigil = sigil

	atomic.AddInt64(&f.nodeCount, 1)

	if f.hooks.OnCreate != nil {
		f.hooks.OnCreate(expr)
	}

	return expr
}

// NewBinaryExpr creates a pooled binary expression
func (f *NodeFactory) NewBinaryExpr(left, right ExpressionNode, operator string, start, end Position) *BinaryExpr {
	expr := f.binaryExprPool.New()

	if expr.BaseNode == nil {
		expr.BaseNode = f.baseNodePool.New()
		atomic.AddInt64(&f.poolMisses, 1)
	} else {
		f.resetBaseNode(expr.BaseNode)
		atomic.AddInt64(&f.poolHits, 1)
	}

	expr.BaseNode.nodeType = "binary_expr"
	expr.BaseNode.start = start
	expr.BaseNode.end = end
	expr.Left = left
	expr.Right = right
	expr.Operator = operator

	// Add children
	if left != nil {
		expr.AddChild(left)
	}
	if right != nil {
		expr.AddChild(right)
	}

	atomic.AddInt64(&f.nodeCount, 1)

	if f.hooks.OnCreate != nil {
		f.hooks.OnCreate(expr)
	}

	return expr
}

// NewCallExpr creates a pooled call expression
func (f *NodeFactory) NewCallExpr(function ExpressionNode, args []ExpressionNode, start, end Position) *CallExpr {
	expr := f.callExprPool.New()

	if expr.BaseNode == nil {
		expr.BaseNode = f.baseNodePool.New()
		atomic.AddInt64(&f.poolMisses, 1)
	} else {
		f.resetBaseNode(expr.BaseNode)
		atomic.AddInt64(&f.poolHits, 1)
	}

	expr.BaseNode.nodeType = "call_expr"
	expr.BaseNode.start = start
	expr.BaseNode.end = end
	expr.Function = function
	expr.Arguments = args

	// Add children
	if function != nil {
		expr.AddChild(function)
	}
	for _, arg := range args {
		if arg != nil {
			expr.AddChild(arg)
		}
	}

	atomic.AddInt64(&f.nodeCount, 1)

	if f.hooks.OnCreate != nil {
		f.hooks.OnCreate(expr)
	}

	return expr
}

// Statement creation methods

// NewExpressionStmt creates a pooled expression statement
func (f *NodeFactory) NewExpressionStmt(expr ExpressionNode, start, end Position) *ExpressionStmt {
	stmt := f.expressionStmtPool.New()

	if stmt.BaseNode == nil {
		stmt.BaseNode = f.baseNodePool.New()
		atomic.AddInt64(&f.poolMisses, 1)
	} else {
		f.resetBaseNode(stmt.BaseNode)
		atomic.AddInt64(&f.poolHits, 1)
	}

	stmt.BaseNode.nodeType = "expression_stmt"
	stmt.BaseNode.start = start
	stmt.BaseNode.end = end
	stmt.Expression = expr

	if expr != nil {
		stmt.AddChild(expr)
	}

	atomic.AddInt64(&f.nodeCount, 1)

	if f.hooks.OnCreate != nil {
		f.hooks.OnCreate(stmt)
	}

	return stmt
}

// NewVarDecl creates a pooled variable declaration
func (f *NodeFactory) NewVarDecl(declType string, vars []*VariableExpr, typeExpr *TypeExpression, init ExpressionNode, start, end Position) *VarDecl {
	decl := f.varDeclPool.New()

	if decl.BaseNode == nil {
		decl.BaseNode = f.baseNodePool.New()
		atomic.AddInt64(&f.poolMisses, 1)
	} else {
		f.resetBaseNode(decl.BaseNode)
		atomic.AddInt64(&f.poolHits, 1)
	}

	decl.BaseNode.nodeType = "var_decl"
	decl.BaseNode.start = start
	decl.BaseNode.end = end
	decl.DeclType = declType
	decl.Variables = vars
	decl.TypeExpr = typeExpr
	decl.Initializer = init

	// Add children
	for _, v := range vars {
		if v != nil {
			decl.AddChild(v)
		}
	}
	if typeExpr != nil {
		decl.AddChild(typeExpr)
	}
	if init != nil {
		decl.AddChild(init)
	}

	atomic.AddInt64(&f.nodeCount, 1)

	if f.hooks.OnCreate != nil {
		f.hooks.OnCreate(decl)
	}

	return decl
}

// NewBlockStmt creates a pooled block statement
func (f *NodeFactory) NewBlockStmt(statements []StatementNode, start, end Position) *BlockStmt {
	stmt := f.blockStmtPool.New()

	if stmt.BaseNode == nil {
		stmt.BaseNode = f.baseNodePool.New()
		atomic.AddInt64(&f.poolMisses, 1)
	} else {
		f.resetBaseNode(stmt.BaseNode)
		atomic.AddInt64(&f.poolHits, 1)
	}

	stmt.BaseNode.nodeType = "block_stmt"
	stmt.BaseNode.start = start
	stmt.BaseNode.end = end
	stmt.Statements = statements

	// Add children
	for _, s := range statements {
		if s != nil {
			stmt.AddChild(s)
		}
	}

	atomic.AddInt64(&f.nodeCount, 1)

	if f.hooks.OnCreate != nil {
		f.hooks.OnCreate(stmt)
	}

	return stmt
}

// Type expression creation methods

// NewTypeExpression creates a pooled type expression
func (f *NodeFactory) NewTypeExpression(name string, params []*TypeExpression, start, end Position) *TypeExpression {
	expr := f.typeExprPool.New()

	if expr.BaseNode == nil {
		expr.BaseNode = f.baseNodePool.New()
		atomic.AddInt64(&f.poolMisses, 1)
	} else {
		f.resetBaseNode(expr.BaseNode)
		atomic.AddInt64(&f.poolHits, 1)
	}

	expr.BaseNode.nodeType = "type_expr"
	expr.BaseNode.start = start
	expr.BaseNode.end = end
	expr.Name = name
	expr.Parameters = params

	// Add children
	for _, param := range params {
		if param != nil {
			expr.AddChild(param)
		}
	}

	atomic.AddInt64(&f.nodeCount, 1)

	if f.hooks.OnCreate != nil {
		f.hooks.OnCreate(expr)
	}

	return expr
}

// Slice allocation methods

// NewNodeSlice creates a pooled slice of nodes
func (f *NodeFactory) NewNodeSlice(capacity int) []Node {
	if capacity <= 0 {
		capacity = 8 // Default capacity
	}

	slice := make([]Node, 0, capacity)
	atomic.AddInt64(&f.memoryReused, int64(cap(slice)))

	return slice
}

// NewVariableSlice creates a pooled slice of variable expressions
func (f *NodeFactory) NewVariableSlice(capacity int) []*VariableExpr {
	if capacity <= 0 {
		capacity = 4 // Default capacity for variable lists
	}

	slice := make([]*VariableExpr, 0, capacity)
	atomic.AddInt64(&f.memoryReused, int64(cap(slice)))

	return slice
}

// NewStatementSlice creates a pooled slice of statements
func (f *NodeFactory) NewStatementSlice(capacity int) []StatementNode {
	if capacity <= 0 {
		capacity = 8 // Default capacity for statement lists
	}

	slice := make([]StatementNode, 0, capacity)
	atomic.AddInt64(&f.memoryReused, int64(cap(slice)))

	return slice
}

// Utility methods

// resetBaseNode resets a BaseNode for reuse
func (f *NodeFactory) resetBaseNode(node *BaseNode) {
	// Clear children slice but keep capacity
	node.children = node.children[:0]
	node.parent = nil
	node.text = ""

	if f.hooks.OnReset != nil {
		f.hooks.OnReset(node)
	}
}

// Reset clears all pools for reuse
func (f *NodeFactory) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Reset all pools
	f.baseNodePool.Reset()
	f.literalExprPool.Reset()
	f.variableExprPool.Reset()
	f.binaryExprPool.Reset()
	f.callExprPool.Reset()
	f.unaryExprPool.Reset()
	f.assignmentExprPool.Reset()
	f.expressionStmtPool.Reset()
	f.varDeclPool.Reset()
	f.ifStmtPool.Reset()
	f.whileStmtPool.Reset()
	f.forStmtPool.Reset()
	f.blockStmtPool.Reset()
	f.subDeclPool.Reset()
	f.typeExprPool.Reset()
	f.unionTypePool.Reset()
	f.intersectionTypePool.Reset()

	if f.nodeSlicePool != nil {
		f.nodeSlicePool.Reset()
	}
	if f.varSlicePool != nil {
		f.varSlicePool.Reset()
	}
	if f.stmtSlicePool != nil {
		f.stmtSlicePool.Reset()
	}
}

// Clear completely empties all pools and resets statistics
func (f *NodeFactory) Clear() {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Clear all pools
	f.baseNodePool.Clear()
	f.literalExprPool.Clear()
	f.variableExprPool.Clear()
	f.binaryExprPool.Clear()
	f.callExprPool.Clear()
	f.unaryExprPool.Clear()
	f.assignmentExprPool.Clear()
	f.expressionStmtPool.Clear()
	f.varDeclPool.Clear()
	f.ifStmtPool.Clear()
	f.whileStmtPool.Clear()
	f.forStmtPool.Clear()
	f.blockStmtPool.Clear()
	f.subDeclPool.Clear()
	f.typeExprPool.Clear()
	f.unionTypePool.Clear()
	f.intersectionTypePool.Clear()

	if f.nodeSlicePool != nil {
		f.nodeSlicePool.Clear()
	}
	if f.varSlicePool != nil {
		f.varSlicePool.Clear()
	}
	if f.stmtSlicePool != nil {
		f.stmtSlicePool.Clear()
	}

	// Reset statistics
	atomic.StoreInt64(&f.nodeCount, 0)
	atomic.StoreInt64(&f.poolHits, 0)
	atomic.StoreInt64(&f.poolMisses, 0)
	atomic.StoreInt64(&f.memoryReused, 0)
}

// Global factory instance
var defaultFactory = NewNodeFactory(NodeFactoryHooks{})

// DefaultFactory returns the global default factory
func DefaultFactory() *NodeFactory {
	return defaultFactory
}

// SetDefaultFactory sets the global default factory
func SetDefaultFactory(factory *NodeFactory) {
	defaultFactory = factory
}
