// ABOUTME: Tests for the parser implementation
// ABOUTME: Verifies parsing of Perl code with type annotations

package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSimpleTypeAnnotations(t *testing.T) {
	// Create a parser
	p, err := NewParser()
	require.NoError(t, err)
	require.NotNil(t, p)

	// Sample Perl code with type annotations
	code := `
		# Variable declarations with type annotations
		my Str $name = "John";
		our Int @ages = (25, 30, 35);
		state HashRef[Str] $cache;

		# Subroutine declarations with type annotations
		sub add(Int $a, Int $b) -> Int {
			return $a + $b;
		}

		# Subroutine declaration with parameters but no return type
		sub process(Str $input, ArrayRef[Str] $options) {
			# Implementation
		}

		# Subroutine declaration with return type but no parameters
		sub get_config() -> HashRef {
			# Implementation
		}
	`

	// Parse the code
	ast, err := p.ParseString(code)
	require.NoError(t, err)
	require.NotNil(t, ast)

	// Check the type annotations
	assert.NotEmpty(t, ast.TypeAnnotations)

	// Check for variable type annotations
	foundStrVar := false
	foundIntArray := false
	foundHashRefMap := false

	// Check for subroutine parameter type annotations
	foundAddParamA := false
	foundAddParamB := false
	foundProcessParamInput := false
	foundProcessParamOptions := false

	// Check for subroutine return type annotations
	foundAddReturnType := false
	foundGetConfigReturnType := false

	for _, annotation := range ast.TypeAnnotations {
		// Variable type annotations
		if annotation.Kind == VarAnnotation {
			if annotation.AnnotatedItem == "$name" && annotation.TypeExpression.Name == "Str" {
				foundStrVar = true
			} else if annotation.AnnotatedItem == "@ages" && annotation.TypeExpression.Name == "Int" {
				foundIntArray = true
			} else if annotation.AnnotatedItem == "$cache" && annotation.TypeExpression.Name == "HashRef" {
				foundHashRefMap = true
				// Check that it's a parameterized type
				assert.Equal(t, 1, len(annotation.TypeExpression.Params))
				assert.Equal(t, "Str", annotation.TypeExpression.Params[0].Name)
			}
		}

		// Subroutine parameter type annotations
		if annotation.Kind == SubParamAnnotation {
			if annotation.AnnotatedItem == "$a" && annotation.TypeExpression.Name == "Int" {
				foundAddParamA = true
			} else if annotation.AnnotatedItem == "$b" && annotation.TypeExpression.Name == "Int" {
				foundAddParamB = true
			} else if annotation.AnnotatedItem == "$input" && annotation.TypeExpression.Name == "Str" {
				foundProcessParamInput = true
			} else if annotation.AnnotatedItem == "$options" && annotation.TypeExpression.Name == "ArrayRef" {
				foundProcessParamOptions = true
				// Check that it's a parameterized type
				assert.Equal(t, 1, len(annotation.TypeExpression.Params))
				assert.Equal(t, "Str", annotation.TypeExpression.Params[0].Name)
			}
		}

		// Subroutine return type annotations
		if annotation.Kind == SubReturnAnnotation {
			if annotation.TypeExpression.Name == "Int" {
				foundAddReturnType = true
			} else if annotation.TypeExpression.Name == "HashRef" {
				foundGetConfigReturnType = true
			}
		}
	}

	// Basic checks for variable type annotations
	assert.True(t, foundStrVar, "Should find scalar variable with Str type")
	assert.True(t, foundIntArray, "Should find array variable with Int type")
	assert.True(t, foundHashRefMap, "Should find scalar variable with HashRef[Str] type")

	// Basic checks for subroutine parameter type annotations
	assert.True(t, foundAddParamA, "Should find parameter $a with Int type")
	assert.True(t, foundAddParamB, "Should find parameter $b with Int type")
	assert.True(t, foundProcessParamInput, "Should find parameter $input with Str type")
	assert.True(t, foundProcessParamOptions, "Should find parameter $options with ArrayRef[Str] type")

	// Basic checks for subroutine return type annotations
	assert.True(t, foundAddReturnType, "Should find return type Int for add()")
	assert.True(t, foundGetConfigReturnType, "Should find return type HashRef for get_config()")
}

func TestParseComplexTypeExpressions(t *testing.T) {
	// Test parsing complex type expressions
	testCases := []struct {
		name           string
		typeExpression string
		validate       func(t *testing.T, expr *TypeExpression)
	}{
		{
			name:           "Simple type",
			typeExpression: "Int",
			validate: func(t *testing.T, expr *TypeExpression) {
				assert.Equal(t, "Int", expr.Name)
				assert.Empty(t, expr.Params)
				assert.False(t, expr.Union)
				assert.False(t, expr.Intersection)
				assert.False(t, expr.Negation)
			},
		},
		{
			name:           "Parameterized type",
			typeExpression: "ArrayRef[Int]",
			validate: func(t *testing.T, expr *TypeExpression) {
				assert.Equal(t, "ArrayRef", expr.Name)
				assert.Equal(t, 1, len(expr.Params))
				assert.Equal(t, "Int", expr.Params[0].Name)
				assert.False(t, expr.Union)
				assert.False(t, expr.Intersection)
				assert.False(t, expr.Negation)
			},
		},
		{
			name:           "Multi-parameter type",
			typeExpression: "HashRef[Str, Int]",
			validate: func(t *testing.T, expr *TypeExpression) {
				assert.Equal(t, "HashRef", expr.Name)
				assert.Equal(t, 2, len(expr.Params))
				assert.Equal(t, "Str", expr.Params[0].Name)
				assert.Equal(t, "Int", expr.Params[1].Name)
				assert.False(t, expr.Union)
				assert.False(t, expr.Intersection)
				assert.False(t, expr.Negation)
			},
		},
		{
			name:           "Nested parameterized type",
			typeExpression: "ArrayRef[HashRef[Str, Int]]",
			validate: func(t *testing.T, expr *TypeExpression) {
				assert.Equal(t, "ArrayRef", expr.Name)
				assert.Equal(t, 1, len(expr.Params))
				assert.Equal(t, "HashRef", expr.Params[0].Name)
				assert.Equal(t, 2, len(expr.Params[0].Params))
				assert.Equal(t, "Str", expr.Params[0].Params[0].Name)
				assert.Equal(t, "Int", expr.Params[0].Params[1].Name)
				assert.False(t, expr.Union)
				assert.False(t, expr.Intersection)
				assert.False(t, expr.Negation)
			},
		},
		{
			name:           "Union type",
			typeExpression: "Str|Undef",
			validate: func(t *testing.T, expr *TypeExpression) {
				assert.Equal(t, "Str|Undef", expr.Name)
				assert.Equal(t, 2, len(expr.Params))
				assert.Equal(t, "Str", expr.Params[0].Name)
				assert.Equal(t, "Undef", expr.Params[1].Name)
				assert.True(t, expr.Union)
				assert.False(t, expr.Intersection)
				assert.False(t, expr.Negation)
			},
		},
		{
			name:           "Intersection type",
			typeExpression: "HasField&Serializable",
			validate: func(t *testing.T, expr *TypeExpression) {
				assert.Equal(t, "HasField&Serializable", expr.Name)
				assert.Equal(t, 2, len(expr.Params))
				assert.Equal(t, "HasField", expr.Params[0].Name)
				assert.Equal(t, "Serializable", expr.Params[1].Name)
				assert.False(t, expr.Union)
				assert.True(t, expr.Intersection)
				assert.False(t, expr.Negation)
			},
		},
		{
			name:           "Negation type",
			typeExpression: "!Undef",
			validate: func(t *testing.T, expr *TypeExpression) {
				assert.Equal(t, "Undef", expr.Name)
				assert.Empty(t, expr.Params)
				assert.False(t, expr.Union)
				assert.False(t, expr.Intersection)
				assert.True(t, expr.Negation)
			},
		},
		{
			name:           "Complex type with union and parameterized types",
			typeExpression: "ArrayRef[Int]|HashRef[Str, Int]",
			validate: func(t *testing.T, expr *TypeExpression) {
				assert.True(t, expr.Union)
				assert.Equal(t, 2, len(expr.Params))

				// First parameter should be ArrayRef[Int]
				assert.Equal(t, "ArrayRef", expr.Params[0].Name)
				assert.Equal(t, 1, len(expr.Params[0].Params))
				assert.Equal(t, "Int", expr.Params[0].Params[0].Name)

				// Second parameter should be HashRef[Str, Int]
				assert.Equal(t, "HashRef", expr.Params[1].Name)
				assert.Equal(t, 2, len(expr.Params[1].Params))
				assert.Equal(t, "Str", expr.Params[1].Params[0].Name)
				assert.Equal(t, "Int", expr.Params[1].Params[1].Name)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := ParseTypeExpression(tc.typeExpression, Position{Line: 1, Column: 1})
			require.NoError(t, err)
			require.NotNil(t, expr)

			tc.validate(t, expr)

			// Test that the string representation matches the original (ignoring whitespace)
			strRepr := expr.String()
			assert.Equal(t,
				strings.ReplaceAll(tc.typeExpression, " ", ""),
				strings.ReplaceAll(strRepr, " ", ""))
		})
	}
}

func TestParseMethodTypeAnnotations(t *testing.T) {
	// Create a parser
	p, err := NewParser()
	require.NoError(t, err)
	require.NotNil(t, p)

	// Sample Perl code with method type annotations
	code := `
		# Class with method declarations using the class feature
		class UserManager {
			field HashRef[Str] $config;

			method connect(Str $host, Int $port) -> Bool {
				# Implementation
			}

			method get_user(Int $id) -> User|Undef {
				# Implementation
			}
		}

		# Traditional OO style with method declarations
		package UserRepository;

		sub new(UserRepository $class, Str $dsn) -> UserRepository {
			# Implementation
		}

		sub find_by_id(UserRepository $self, Int $id) -> Maybe[User] {
			# Implementation
		}
	`

	// Parse the code
	ast, err := p.ParseString(code)
	require.NoError(t, err)
	require.NotNil(t, ast)

	// Check the type annotations
	assert.NotEmpty(t, ast.TypeAnnotations)

	// Check for field type annotations
	foundConfigField := false

	// Check for method parameter type annotations
	foundConnectHostParam := false
	foundConnectPortParam := false
	foundGetUserIdParam := false

	// Check for method return type annotations
	foundConnectReturnType := false
	foundGetUserReturnType := false

	// Check for traditional method parameter type annotations
	foundNewClassParam := false
	foundNewDsnParam := false
	foundFindByIdSelfParam := false
	foundFindByIdIdParam := false

	// Check for traditional method return type annotations
	foundNewReturnType := false
	foundFindByIdReturnType := false

	for _, annotation := range ast.TypeAnnotations {
		// Field type annotations
		if annotation.Kind == AttrAnnotation {
			if annotation.AnnotatedItem == "$config" && annotation.TypeExpression.Name == "HashRef" {
				foundConfigField = true
				// Check that it's a parameterized type
				assert.Equal(t, 1, len(annotation.TypeExpression.Params))
				assert.Equal(t, "Str", annotation.TypeExpression.Params[0].Name)
			}
		}

		// Method parameter type annotations
		if annotation.Kind == MethodParamAnnotation {
			if annotation.AnnotatedItem == "$host" && annotation.TypeExpression.Name == "Str" {
				foundConnectHostParam = true
			} else if annotation.AnnotatedItem == "$port" && annotation.TypeExpression.Name == "Int" {
				foundConnectPortParam = true
			} else if annotation.AnnotatedItem == "$id" && annotation.TypeExpression.Name == "Int" {
				foundGetUserIdParam = true
			}
		}

		// Method return type annotations
		if annotation.Kind == MethodReturnAnnotation {
			if annotation.TypeExpression.Name == "Bool" {
				foundConnectReturnType = true
			} else if annotation.TypeExpression.Name == "User|Undef" {
				foundGetUserReturnType = true
				// Check that it's a union type
				assert.True(t, annotation.TypeExpression.Union)
				assert.Equal(t, 2, len(annotation.TypeExpression.Params))
				assert.Equal(t, "User", annotation.TypeExpression.Params[0].Name)
				assert.Equal(t, "Undef", annotation.TypeExpression.Params[1].Name)
			}
		}

		// Traditional method parameter type annotations
		if annotation.Kind == SubParamAnnotation {
			if annotation.AnnotatedItem == "$class" && annotation.TypeExpression.Name == "UserRepository" {
				foundNewClassParam = true
			} else if annotation.AnnotatedItem == "$dsn" && annotation.TypeExpression.Name == "Str" {
				foundNewDsnParam = true
			} else if annotation.AnnotatedItem == "$self" && annotation.TypeExpression.Name == "UserRepository" {
				foundFindByIdSelfParam = true
			} else if annotation.AnnotatedItem == "$id" && annotation.TypeExpression.Name == "Int" {
				foundFindByIdIdParam = true
			}
		}

		// Traditional method return type annotations
		if annotation.Kind == SubReturnAnnotation {
			if annotation.TypeExpression.Name == "UserRepository" {
				foundNewReturnType = true
			} else if annotation.TypeExpression.Name == "Maybe" {
				foundFindByIdReturnType = true
				// Check that it's a parameterized type
				assert.Equal(t, 1, len(annotation.TypeExpression.Params))
				assert.Equal(t, "User", annotation.TypeExpression.Params[0].Name)
			}
		}
	}

	// Basic checks for field type annotations
	assert.True(t, foundConfigField, "Should find field $config with HashRef[Str] type")

	// Basic checks for method parameter type annotations
	assert.True(t, foundConnectHostParam, "Should find parameter $host with Str type")
	assert.True(t, foundConnectPortParam, "Should find parameter $port with Int type")
	assert.True(t, foundGetUserIdParam, "Should find parameter $id with Int type")

	// Basic checks for method return type annotations
	assert.True(t, foundConnectReturnType, "Should find return type Bool for connect()")
	assert.True(t, foundGetUserReturnType, "Should find return type User|Undef for get_user()")

	// Basic checks for traditional method parameter type annotations
	assert.True(t, foundNewClassParam, "Should find parameter $class with UserRepository type")
	assert.True(t, foundNewDsnParam, "Should find parameter $dsn with Str type")
	assert.True(t, foundFindByIdSelfParam, "Should find parameter $self with UserRepository type")
	assert.True(t, foundFindByIdIdParam, "Should find parameter $id with Int type")

	// Basic checks for traditional method return type annotations
	assert.True(t, foundNewReturnType, "Should find return type UserRepository for new()")
	assert.True(t, foundFindByIdReturnType, "Should find return type Maybe[User] for find_by_id()")
}
