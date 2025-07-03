//go:build ignore
// +build ignore

package binder

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
)

func TestLexicalFunctionScoping(t *testing.T) {
	// Create pool manager and symbol table
	poolManager := NewSymbolPoolManager(SymbolPoolHooks{})
	symbolTable := NewSymbolTableWithPool(poolManager, "TestPackage")

	// Create binder
	binder := NewBinderWithPool(poolManager)
	binder.symbolTable = symbolTable

	t.Run("Package subroutines conflict detection", func(t *testing.T) {
		// Create package subroutine 1
		sub1 := ast.NewSubDecl("func1", nil, nil, nil, false, ast.Position{}, ast.Position{})
		err := binder.bindSubroutineDeclaration(sub1)
		if err != nil {
			t.Fatalf("Error binding sub1: %v", err)
		}

		// Create package subroutine 2 with same name - should conflict
		sub2 := ast.NewSubDecl("func1", nil, nil, nil, false, ast.Position{}, ast.Position{})
		err = binder.bindSubroutineDeclaration(sub2)
		if err == nil {
			t.Fatal("Expected conflict for package subroutine func1, but got none")
		}
		if !strings.Contains(err.Error(), "already declared") {
			t.Fatalf("Expected 'already declared' error, got: %v", err)
		}
		t.Logf("✓ Package subroutine conflict detected: %v", err)
	})

	t.Run("Lexical subroutines in different scopes", func(t *testing.T) {
		// Create block 1 and enter its scope
		block1 := ast.NewBlockStmt(nil, ast.Position{}, ast.Position{})
		symbolTable.EnterScope(ScopeBlock, block1)

		// Create lexical subroutine 1
		lexSub1 := ast.NewLexicalSubDecl("lexical_func", nil, nil, nil, false, ast.Position{}, ast.Position{})
		err := binder.bindSubroutineDeclaration(lexSub1)
		if err != nil {
			t.Fatalf("Error binding lexical_func in block1: %v", err)
		}
		t.Log("✓ Lexical subroutine lexical_func bound in block1")

		// Exit block 1 scope
		symbolTable.ExitScope()

		// Create block 2 and enter its scope
		block2 := ast.NewBlockStmt(nil, ast.Position{}, ast.Position{})
		symbolTable.EnterScope(ScopeBlock, block2)

		// Create lexical subroutine 2 with same name - should NOT conflict
		lexSub2 := ast.NewLexicalSubDecl("lexical_func", nil, nil, nil, false, ast.Position{}, ast.Position{})
		err = binder.bindSubroutineDeclaration(lexSub2)
		if err != nil {
			t.Fatalf("Unexpected conflict for lexical subroutine lexical_func: %v", err)
		}
		t.Log("✓ Lexical subroutine lexical_func bound in block2 (no conflict)")

		// Exit block 2 scope
		symbolTable.ExitScope()
	})

	t.Run("Lexical methods in different scopes", func(t *testing.T) {
		// Create block 3 and enter its scope
		block3 := ast.NewBlockStmt(nil, ast.Position{}, ast.Position{})
		symbolTable.EnterScope(ScopeBlock, block3)

		// Create lexical method 1
		lexMethod1 := ast.NewLexicalMethodDecl("lexical_method", nil, nil, nil, ast.Position{}, ast.Position{})
		err := binder.bindMethodDeclaration(lexMethod1)
		if err != nil {
			t.Fatalf("Error binding lexical_method in block3: %v", err)
		}
		t.Log("✓ Lexical method lexical_method bound in block3")

		// Exit block 3 scope
		symbolTable.ExitScope()

		// Create block 4 and enter its scope
		block4 := ast.NewBlockStmt(nil, ast.Position{}, ast.Position{})
		symbolTable.EnterScope(ScopeBlock, block4)

		// Create lexical method 2 with same name - should NOT conflict
		lexMethod2 := ast.NewLexicalMethodDecl("lexical_method", nil, nil, nil, ast.Position{}, ast.Position{})
		err = binder.bindMethodDeclaration(lexMethod2)
		if err != nil {
			t.Fatalf("Unexpected conflict for lexical method lexical_method: %v", err)
		}
		t.Log("✓ Lexical method lexical_method bound in block4 (no conflict)")

		// Exit block 4 scope
		symbolTable.ExitScope()
	})

	t.Run("Mixed lexical and package functions", func(t *testing.T) {
		// Create package function
		packageFunc := ast.NewSubDecl("mixed_func", nil, nil, nil, false, ast.Position{}, ast.Position{})
		err := binder.bindSubroutineDeclaration(packageFunc)
		if err != nil {
			t.Fatalf("Error binding package mixed_func: %v", err)
		}

		// Create block and enter its scope
		block := ast.NewBlockStmt(nil, ast.Position{}, ast.Position{})
		symbolTable.EnterScope(ScopeBlock, block)

		// Create lexical function with same name - should NOT conflict (different scopes)
		lexicalFunc := ast.NewLexicalSubDecl("mixed_func", nil, nil, nil, false, ast.Position{}, ast.Position{})
		err = binder.bindSubroutineDeclaration(lexicalFunc)
		if err != nil {
			t.Fatalf("Unexpected conflict between package and lexical mixed_func: %v", err)
		}
		t.Log("✓ Lexical function mixed_func coexists with package function")

		// Exit block scope
		symbolTable.ExitScope()
	})
}
