package ast

import (
	"sync"
	"testing"
)

func TestNodeFactory_NewLiteralExpr(t *testing.T) {
	factory := NewNodeFactory(NodeFactoryHooks{})

	// Create literal expression
	expr := factory.NewLiteralExpr("hello", StringLiteral, Position{1, 1, 0}, Position{1, 7, 6})

	if expr == nil {
		t.Fatal("Expected non-nil literal expression")
	}
	if expr.Value != "hello" {
		t.Errorf("Expected value 'hello', got '%s'", expr.Value)
	}
	if expr.Kind != StringLiteral {
		t.Errorf("Expected StringLiteral kind, got %d", expr.Kind)
	}
	if expr.Type() != "literal" {
		t.Errorf("Expected type 'literal', got '%s'", expr.Type())
	}

	// Check factory statistics
	if factory.NodeCount() != 1 {
		t.Errorf("Expected node count 1, got %d", factory.NodeCount())
	}
}

func TestNodeFactory_NewVariableExpr(t *testing.T) {
	factory := NewNodeFactory(NodeFactoryHooks{})

	// Create variable expression
	expr := factory.NewVariableExpr("count", "$", Position{1, 1, 0}, Position{1, 6, 5})

	if expr == nil {
		t.Fatal("Expected non-nil variable expression")
	}
	if expr.Name != "count" {
		t.Errorf("Expected name 'count', got '%s'", expr.Name)
	}
	if expr.Sigil != "$" {
		t.Errorf("Expected sigil '$', got '%s'", expr.Sigil)
	}
	if expr.FullName() != "$count" {
		t.Errorf("Expected full name '$count', got '%s'", expr.FullName())
	}
}

func TestNodeFactory_NewBinaryExpr(t *testing.T) {
	factory := NewNodeFactory(NodeFactoryHooks{})

	// Create operands
	left := factory.NewVariableExpr("a", "$", Position{1, 1, 0}, Position{1, 2, 1})
	right := factory.NewLiteralExpr("5", NumberLiteral, Position{1, 6, 5}, Position{1, 7, 6})

	// Create binary expression
	expr := factory.NewBinaryExpr(left, right, "+", Position{1, 1, 0}, Position{1, 7, 6})

	if expr == nil {
		t.Fatal("Expected non-nil binary expression")
	}
	if expr.Left != left {
		t.Error("Expected left operand to match")
	}
	if expr.Right != right {
		t.Error("Expected right operand to match")
	}
	if expr.Operator != "+" {
		t.Errorf("Expected operator '+', got '%s'", expr.Operator)
	}

	// Check children are properly set
	children := expr.Children()
	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

func TestNodeFactory_NewCallExpr(t *testing.T) {
	factory := NewNodeFactory(NodeFactoryHooks{})

	// Create function and arguments
	function := factory.NewVariableExpr("print", "&", Position{1, 1, 0}, Position{1, 6, 5})
	arg1 := factory.NewLiteralExpr("hello", StringLiteral, Position{1, 7, 6}, Position{1, 14, 13})
	arg2 := factory.NewVariableExpr("world", "$", Position{1, 16, 15}, Position{1, 22, 21})

	// Create call expression
	expr := factory.NewCallExpr(function, []ExpressionNode{arg1, arg2}, Position{1, 1, 0}, Position{1, 23, 22})

	if expr == nil {
		t.Fatal("Expected non-nil call expression")
	}
	if expr.Function != function {
		t.Error("Expected function to match")
	}
	if len(expr.Arguments) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(expr.Arguments))
	}

	// Check children include function and arguments
	children := expr.Children()
	if len(children) != 3 { // function + 2 arguments
		t.Errorf("Expected 3 children, got %d", len(children))
	}
}

func TestNodeFactory_NewExpressionStmt(t *testing.T) {
	factory := NewNodeFactory(NodeFactoryHooks{})

	// Create expression
	expr := factory.NewLiteralExpr("42", NumberLiteral, Position{1, 1, 0}, Position{1, 3, 2})

	// Create expression statement
	stmt := factory.NewExpressionStmt(expr, Position{1, 1, 0}, Position{1, 4, 3})

	if stmt == nil {
		t.Fatal("Expected non-nil expression statement")
	}
	if stmt.Expression != expr {
		t.Error("Expected expression to match")
	}
	if !stmt.IsStatement() {
		t.Error("Expected IsStatement() to return true")
	}

	// Check child is properly set
	children := stmt.Children()
	if len(children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(children))
	}
}

func TestNodeFactory_NewVarDecl(t *testing.T) {
	factory := NewNodeFactory(NodeFactoryHooks{})

	// Create variable and type
	variable := factory.NewVariableExpr("count", "$", Position{1, 5, 4}, Position{1, 11, 10})
	typeExpr := factory.NewTypeExpression("Int", nil, Position{1, 1, 0}, Position{1, 4, 3})
	init := factory.NewLiteralExpr("0", NumberLiteral, Position{1, 15, 14}, Position{1, 16, 15})

	// Create variable declaration
	decl := factory.NewVarDecl("my", []*VariableExpr{variable}, typeExpr, init, Position{1, 1, 0}, Position{1, 17, 16})

	if decl == nil {
		t.Fatal("Expected non-nil variable declaration")
	}
	if decl.DeclType != "my" {
		t.Errorf("Expected declaration type 'my', got '%s'", decl.DeclType)
	}
	if len(decl.LogicalVariables()) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(decl.LogicalVariables()))
	}
	if decl.LogicalVariables()[0] != variable {
		t.Error("Expected variable to match")
	}
	if decl.TypeExpr != typeExpr {
		t.Error("Expected type expression to match")
	}
	if decl.Initializer != init {
		t.Error("Expected initializer to match")
	}
}

func TestNodeFactory_NewBlockStmt(t *testing.T) {
	factory := NewNodeFactory(NodeFactoryHooks{})

	// Create statements
	stmt1 := factory.NewExpressionStmt(
		factory.NewLiteralExpr("hello", StringLiteral, Position{1, 1, 0}, Position{1, 7, 6}),
		Position{1, 1, 0}, Position{1, 8, 7})
	stmt2 := factory.NewExpressionStmt(
		factory.NewLiteralExpr("world", StringLiteral, Position{2, 1, 8}, Position{2, 7, 14}),
		Position{2, 1, 8}, Position{2, 8, 15})

	// Create block statement
	block := factory.NewBlockStmt([]StatementNode{stmt1, stmt2}, Position{1, 1, 0}, Position{2, 9, 16})

	if block == nil {
		t.Fatal("Expected non-nil block statement")
	}
	if len(block.LogicalStatements()) != 2 {
		t.Errorf("Expected 2 statements, got %d", len(block.LogicalStatements()))
	}
	if !block.IsStatement() {
		t.Error("Expected IsStatement() to return true")
	}

	// Check children are properly set
	children := block.Children()
	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

func TestNodeFactory_NewTypeExpression(t *testing.T) {
	factory := NewNodeFactory(NodeFactoryHooks{})

	// Create parameterized type
	param := factory.NewTypeExpression("Int", nil, Position{1, 10, 9}, Position{1, 13, 12})
	typeExpr := factory.NewTypeExpression("ArrayRef", []*TypeExpression{param}, Position{1, 1, 0}, Position{1, 14, 13})

	if typeExpr == nil {
		t.Fatal("Expected non-nil type expression")
	}
	if typeExpr.Name != "ArrayRef" {
		t.Errorf("Expected name 'ArrayRef', got '%s'", typeExpr.Name)
	}
	if len(typeExpr.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(typeExpr.Parameters))
	}
	if typeExpr.Parameters[0] != param {
		t.Error("Expected parameter to match")
	}

	// Check child is properly set
	children := typeExpr.Children()
	if len(children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(children))
	}
}

func TestNodeFactory_PoolingEfficiency(t *testing.T) {
	factory := NewNodeFactory(NodeFactoryHooks{})

	// Create many nodes to test pooling
	const numNodes = 100

	for i := 0; i < numNodes; i++ {
		expr := factory.NewLiteralExpr("test", StringLiteral, Position{1, 1, 0}, Position{1, 5, 4})
		if expr == nil {
			t.Fatalf("Failed to create node %d", i)
		}
	}

	if factory.NodeCount() != numNodes {
		t.Errorf("Expected %d nodes created, got %d", numNodes, factory.NodeCount())
	}

	// Test efficiency calculation
	efficiency := factory.PoolEfficiency()
	if efficiency < 0 || efficiency > 100 {
		t.Errorf("Expected efficiency between 0-100%%, got %f", efficiency)
	}
}

func TestNodeFactory_Hooks(t *testing.T) {
	var createdNodes []Node
	var resetNodes []Node

	hooks := NodeFactoryHooks{
		OnCreate: func(node Node) {
			createdNodes = append(createdNodes, node)
		},
		OnReset: func(node Node) {
			resetNodes = append(resetNodes, node)
		},
	}

	factory := NewNodeFactory(hooks)

	// Create a node
	expr := factory.NewLiteralExpr("test", StringLiteral, Position{1, 1, 0}, Position{1, 5, 4})

	if len(createdNodes) != 1 {
		t.Errorf("Expected 1 created node, got %d", len(createdNodes))
	}
	if createdNodes[0] != expr {
		t.Error("Expected created node to match")
	}

	// Reset factory to trigger reset hooks
	factory.Reset()

	// Create another node to trigger reuse and reset
	factory.NewLiteralExpr("test2", StringLiteral, Position{1, 1, 0}, Position{1, 6, 5})

	// Should have triggered reset hook when reusing BaseNode
	if len(resetNodes) == 0 {
		t.Error("Expected reset hook to be called during reuse")
	}
}

func TestNodeFactory_ConcurrentAccess(t *testing.T) {
	factory := NewNodeFactory(NodeFactoryHooks{})
	const numGoroutines = 10
	const nodesPerGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < nodesPerGoroutine; j++ {
				expr := factory.NewLiteralExpr("concurrent", StringLiteral, Position{1, 1, 0}, Position{1, 11, 10})
				if expr == nil {
					t.Errorf("Goroutine %d failed to create node %d", id, j)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	expectedNodes := int64(numGoroutines * nodesPerGoroutine)
	if factory.NodeCount() != expectedNodes {
		t.Errorf("Expected %d nodes from concurrent creation, got %d", expectedNodes, factory.NodeCount())
	}
}

func TestNodeFactory_SliceAllocation(t *testing.T) {
	factory := NewNodeFactory(NodeFactoryHooks{})

	// Test node slice allocation
	nodeSlice := factory.NewNodeSlice(10)
	if cap(nodeSlice) < 10 {
		t.Errorf("Expected node slice capacity >= 10, got %d", cap(nodeSlice))
	}

	// Test variable slice allocation
	varSlice := factory.NewVariableSlice(5)
	if cap(varSlice) < 5 {
		t.Errorf("Expected variable slice capacity >= 5, got %d", cap(varSlice))
	}

	// Test statement slice allocation
	stmtSlice := factory.NewStatementSlice(8)
	if cap(stmtSlice) < 8 {
		t.Errorf("Expected statement slice capacity >= 8, got %d", cap(stmtSlice))
	}
}

func TestNodeFactory_ResetAndClear(t *testing.T) {
	factory := NewNodeFactory(NodeFactoryHooks{})

	// Create some nodes
	factory.NewLiteralExpr("test1", StringLiteral, Position{1, 1, 0}, Position{1, 6, 5})
	factory.NewVariableExpr("test", "$", Position{1, 1, 0}, Position{1, 5, 4})

	initialCount := factory.NodeCount()
	if initialCount != 2 {
		t.Errorf("Expected 2 nodes initially, got %d", initialCount)
	}

	// Reset should keep statistics but clear pools for reuse
	factory.Reset()
	if factory.NodeCount() != initialCount {
		t.Error("Reset should not change node count")
	}

	// Clear should reset everything
	factory.Clear()
	if factory.NodeCount() != 0 {
		t.Errorf("Expected 0 nodes after clear, got %d", factory.NodeCount())
	}
}

func TestDefaultFactory(t *testing.T) {
	factory := DefaultFactory()
	if factory == nil {
		t.Fatal("Expected non-nil default factory")
	}

	// Test that we can create nodes with default factory
	expr := factory.NewLiteralExpr("default", StringLiteral, Position{1, 1, 0}, Position{1, 8, 7})
	if expr == nil {
		t.Fatal("Expected non-nil expression from default factory")
	}

	// Test setting custom default factory
	customFactory := NewNodeFactory(NodeFactoryHooks{})
	SetDefaultFactory(customFactory)

	if DefaultFactory() != customFactory {
		t.Error("Expected default factory to be updated")
	}
}

func BenchmarkNodeFactory_LiteralExpr(b *testing.B) {
	factory := NewNodeFactory(NodeFactoryHooks{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expr := factory.NewLiteralExpr("benchmark", StringLiteral, Position{1, 1, 0}, Position{1, 10, 9})
		if expr == nil {
			b.Fatal("Failed to create literal expression")
		}
	}
}

func BenchmarkNodeFactory_BinaryExpr(b *testing.B) {
	factory := NewNodeFactory(NodeFactoryHooks{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		left := factory.NewVariableExpr("a", "$", Position{1, 1, 0}, Position{1, 2, 1})
		right := factory.NewLiteralExpr("5", NumberLiteral, Position{1, 6, 5}, Position{1, 7, 6})
		expr := factory.NewBinaryExpr(left, right, "+", Position{1, 1, 0}, Position{1, 7, 6})
		if expr == nil {
			b.Fatal("Failed to create binary expression")
		}
	}
}

func BenchmarkNodeFactory_Concurrent(b *testing.B) {
	factory := NewNodeFactory(NodeFactoryHooks{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			expr := factory.NewLiteralExpr("concurrent", StringLiteral, Position{1, 1, 0}, Position{1, 11, 10})
			if expr == nil {
				b.Fatal("Failed to create expression concurrently")
			}
		}
	})
}
