// ABOUTME: Tests for enhanced AST functionality with type inference information
// ABOUTME: Validates InferredAST interface and type attachment system

package ast

import (
	"testing"

	"tamarou.com/pvm/internal/types"
)

func TestInferredASTInterface(t *testing.T) {
	// Create basic AST
	ast := &AST{
		Path:   "test.pl",
		Source: "my $x = 42;",
	}

	// Create enhanced AST with type info
	inferredAST := NewInferredAST(ast)

	t.Run("Basic interface compliance", func(t *testing.T) {
		if inferredAST.GetPath() != "test.pl" {
			t.Errorf("Expected path 'test.pl', got %s", inferredAST.GetPath())
		}

		if !inferredAST.IsValid() {
			t.Error("Expected AST to be valid")
		}

		content, err := inferredAST.GetContent()
		if err != nil {
			t.Errorf("Expected no error getting content, got %v", err)
		}
		if content != "my $x = 42;" {
			t.Errorf("Expected content 'my $x = 42;', got %s", content)
		}
	})

	t.Run("Type information access", func(t *testing.T) {
		// Should be able to get type information for nodes
		typeInfo := inferredAST.GetTypeInfo("variable_x")
		if typeInfo != nil {
			t.Error("Expected no type info initially")
		}

		// Should be able to get all inferred types
		allTypes := inferredAST.GetAllTypeInfo()
		if len(allTypes) != 0 {
			t.Error("Expected no type info initially")
		}
	})
}

func TestTypeAttachment(t *testing.T) {
	ast := &AST{
		Path:   "test.pl",
		Source: "my $x = 42;",
	}

	inferredAST := NewInferredAST(ast)

	t.Run("Attach type to node", func(t *testing.T) {
		// Create type info
		intType := types.NewIntType()
		typeInfo := types.NewTypeInfo(intType, 0.95, types.SourceLiteral)

		// Attach type to node
		err := inferredAST.AttachTypeInfo("variable_x", typeInfo)
		if err != nil {
			t.Errorf("Expected no error attaching type, got %v", err)
		}

		// Retrieve type info
		retrievedInfo := inferredAST.GetTypeInfo("variable_x")
		if retrievedInfo == nil {
			t.Error("Expected to retrieve type info")
		}

		if !retrievedInfo.Type.Equals(intType) {
			t.Errorf("Expected Int type, got %s", retrievedInfo.Type.String())
		}

		if retrievedInfo.Confidence != 0.95 {
			t.Errorf("Expected confidence 0.95, got %f", retrievedInfo.Confidence)
		}
	})

	t.Run("Get all type info", func(t *testing.T) {
		allTypes := inferredAST.GetAllTypeInfo()
		if len(allTypes) != 1 {
			t.Errorf("Expected 1 type info, got %d", len(allTypes))
		}

		if _, exists := allTypes["variable_x"]; !exists {
			t.Error("Expected variable_x in all types")
		}
	})
}

func TestASTTypeAdapter(t *testing.T) {
	// Create basic AST
	baseAST := &AST{
		Path:   "adapter_test.pl",
		Source: "my $count = 10;",
	}

	t.Run("Adapter wraps AST correctly", func(t *testing.T) {
		adapter := NewASTTypeAdapter(baseAST)

		if adapter.GetPath() != "adapter_test.pl" {
			t.Errorf("Expected path 'adapter_test.pl', got %s", adapter.GetPath())
		}

		if !adapter.IsValid() {
			t.Error("Expected adapter to be valid")
		}

		content, err := adapter.GetContent()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if content != "my $count = 10;" {
			t.Errorf("Expected content 'my $count = 10;', got %s", content)
		}
	})

	t.Run("Adapter provides type information", func(t *testing.T) {
		adapter := NewASTTypeAdapter(baseAST)

		// Add type information
		intType := types.NewIntType()
		typeInfo := types.NewTypeInfo(intType, 0.90, types.SourceLiteral)

		err := adapter.AttachTypeInfo("count_var", typeInfo)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Retrieve type information
		retrieved := adapter.GetTypeInfo("count_var")
		if retrieved == nil {
			t.Error("Expected to retrieve type info from adapter")
		}

		if !retrieved.Type.Equals(intType) {
			t.Errorf("Expected Int type, got %s", retrieved.Type.String())
		}
	})
}

func TestTypeConstraintPropagation(t *testing.T) {
	ast := &AST{
		Path:   "constraint_test.pl",
		Source: "my $x = 42; my $y = $x;",
	}

	inferredAST := NewInferredAST(ast)

	t.Run("Add type constraints", func(t *testing.T) {
		// Create constraint that $y should have same type as $x
		intType := types.NewIntType()
		constraint := types.NewAssignmentConstraint(intType, intType)

		err := inferredAST.AddTypeConstraint("x_to_y", constraint)
		if err != nil {
			t.Errorf("Expected no error adding constraint, got %v", err)
		}

		// Should be able to retrieve constraints
		constraints := inferredAST.GetTypeConstraints()
		if len(constraints) != 1 {
			t.Errorf("Expected 1 constraint, got %d", len(constraints))
		}
	})

	t.Run("Validate constraints", func(t *testing.T) {
		errors := inferredAST.ValidateTypeConstraints()
		if len(errors) != 0 {
			t.Errorf("Expected no constraint errors, got %d", len(errors))
		}
	})
}

func TestBackwardCompatibility(t *testing.T) {
	t.Run("Original AST functionality preserved", func(t *testing.T) {
		ast := &AST{
			Path:   "compat_test.pl",
			Source: "print 'hello';",
		}

		// These should work exactly as before
		if ast.GetPath() != "compat_test.pl" {
			t.Error("Original AST functionality broken")
		}

		if !ast.IsValid() {
			t.Error("Original AST validity check broken")
		}

		content, err := ast.GetContent()
		if err != nil || content != "print 'hello';" {
			t.Error("Original AST content access broken")
		}
	})

	t.Run("InferredAST doesn't break original functionality", func(t *testing.T) {
		ast := &AST{
			Path:   "compat_test2.pl",
			Source: "my $var = 'test';",
		}

		inferredAST := NewInferredAST(ast)

		// All original functionality should still work
		if inferredAST.GetPath() != "compat_test2.pl" {
			t.Error("InferredAST breaks original path access")
		}

		if !inferredAST.IsValid() {
			t.Error("InferredAST breaks original validity check")
		}

		content, err := inferredAST.GetContent()
		if err != nil || content != "my $var = 'test';" {
			t.Error("InferredAST breaks original content access")
		}
	})
}
