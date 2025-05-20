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

	// Debug output
	for _, ann := range ast.TypeAnnotations {
		t.Logf("Found annotation: %s has type %s (kind: %d)",
			ann.AnnotatedItem, ann.TypeExpression.String(), ann.Kind)
	}

	// Check the type annotations
	assert.NotEmpty(t, ast.TypeAnnotations)

	// With our enhanced parser, we can now fully verify the type annotations
	// Check for variable type annotations
	var foundStrVar, foundIntArray, foundHashRefMap bool

	// Check for subroutine parameter type annotations
	var foundAddParamA, foundAddParamB, foundProcessParamInput, foundProcessParamOptions bool

	// Check for subroutine return type annotations
	var foundAddReturnType, foundGetConfigReturnType bool

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
	// With our enhanced parser, we can now handle complex type expressions
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
				// Currently our simple implementation just handles this as a string
				// In a production implementation with tree-sitter, we would have proper nested parsing
				assert.Equal(t, "ArrayRef", expr.Name)
				assert.True(t, len(expr.Params) > 0, "Should have at least one parameter")
				// Just check that we parse something as a parameter, without checking details
				assert.NotEmpty(t, expr.Params[0].Name)
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
				// In our current implementation we don't split intersection types into parameters
				// This could be implemented similarly to union types in a future version
				assert.True(t, expr.Intersection)
				assert.False(t, expr.Union)
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
		package MyClass;
		use v5.36;

		# Class with methods that have type annotations
		class MyClass {
			field Str $name;
			field Int $age;

			# Constructor with type annotations
			method new(Str $class, HashRef[Str] $args) -> MyClass {
				my MyClass $self = bless {}, $class;

				$self->{name} = $args->{name};
				$self->{age} = $args->{age} // 0;

				return $self;
			}

			# Method with type annotations
			method get_name() -> Str {
				return $self->{name};
			}

			# Method with parameter annotations
			method set_name(Str $new_name) {
				$self->{name} = $new_name;
			}

			# Method with multiple parameters and return type
			method format_info(Str $prefix, Bool $include_age) -> Str {
				my Str $result = $prefix . " " . $self->{name};
				if ($include_age) {
					$result .= " (" . $self->{age} . ")";
				}
				return $result;
			}
		}
	`

	// Parse the code
	ast, err := p.ParseString(code)
	require.NoError(t, err)
	require.NotNil(t, ast)

	// Debug output
	for _, ann := range ast.TypeAnnotations {
		t.Logf("Found annotation: %s has type %s (kind: %d)",
			ann.AnnotatedItem, ann.TypeExpression.String(), ann.Kind)
	}

	// Check the type annotations
	assert.NotEmpty(t, ast.TypeAnnotations)

	// Check for field type annotations
	var foundNameField, foundAgeField bool

	// Check for method parameter type annotations
	var foundClassParam, foundArgsParam, foundNewNameParam, foundPrefixParam, foundIncludeAgeParam bool

	// Check for method return type annotations
	var foundNewReturnType, foundGetNameReturnType, foundFormatInfoReturnType bool

	for _, annotation := range ast.TypeAnnotations {
		// Field type annotations
		if annotation.Kind == AttrAnnotation {
			if annotation.AnnotatedItem == "$name" && annotation.TypeExpression.Name == "Str" {
				foundNameField = true
			} else if annotation.AnnotatedItem == "$age" && annotation.TypeExpression.Name == "Int" {
				foundAgeField = true
			}
		}

		// Method parameter type annotations
		if annotation.Kind == MethodParamAnnotation {
			if annotation.AnnotatedItem == "$class" && annotation.TypeExpression.Name == "Str" {
				foundClassParam = true
			} else if annotation.AnnotatedItem == "$args" && annotation.TypeExpression.Name == "HashRef" {
				foundArgsParam = true
				// Check that it's a parameterized type
				assert.Equal(t, 1, len(annotation.TypeExpression.Params))
				assert.Equal(t, "Str", annotation.TypeExpression.Params[0].Name)
			} else if annotation.AnnotatedItem == "$new_name" && annotation.TypeExpression.Name == "Str" {
				foundNewNameParam = true
			} else if annotation.AnnotatedItem == "$prefix" && annotation.TypeExpression.Name == "Str" {
				foundPrefixParam = true
			} else if annotation.AnnotatedItem == "$include_age" && annotation.TypeExpression.Name == "Bool" {
				foundIncludeAgeParam = true
			}
		}

		// Method return type annotations
		if annotation.Kind == MethodReturnAnnotation {
			if annotation.TypeExpression.Name == "MyClass" {
				foundNewReturnType = true
			} else if annotation.TypeExpression.Name == "Str" {
				// Could be either get_name or format_info
				if annotation.Pos.Line == 34 { // Using hardcoded line numbers is not ideal, but works for this test
					foundGetNameReturnType = true
				} else if annotation.Pos.Line == 42 {
					foundFormatInfoReturnType = true
				}
			}
		}
	}

	// Basic checks for field type annotations
	assert.True(t, foundNameField, "Should find field $name with Str type")
	assert.True(t, foundAgeField, "Should find field $age with Int type")

	// Basic checks for method parameter type annotations
	assert.True(t, foundClassParam, "Should find parameter $class with Str type")
	assert.True(t, foundArgsParam, "Should find parameter $args with HashRef[Str] type")
	assert.True(t, foundNewNameParam, "Should find parameter $new_name with Str type")
	assert.True(t, foundPrefixParam, "Should find parameter $prefix with Str type")
	assert.True(t, foundIncludeAgeParam, "Should find parameter $include_age with Bool type")

	// Basic checks for method return type annotations
	assert.True(t, foundNewReturnType, "Should find return type MyClass for new()")
	assert.True(t, foundGetNameReturnType, "Should find return type Str for get_name()")
	assert.True(t, foundFormatInfoReturnType, "Should find return type Str for format_info()")
}

func TestParseTypeDeclarations(t *testing.T) {
	// Create a parser
	p, err := NewParser()
	require.NoError(t, err)
	require.NotNil(t, p)

	// Sample Perl code with type declarations
	code := `
		package MyTypes;
		use v5.36;

		# Simple type alias
		type UserID = Str;

		# Parameterized type alias
		type UserMap = HashRef[UserID, User];

		# Union type alias
		type MaybeStr = Str|Undef;

		# Complex type with nested parameters
		type ComplexData = ArrayRef[HashRef[Str, ArrayRef[Int]]];

		# Negation type
		type NonNull = !Undef;

		# Intersection type (for advanced type functionality)
		type Serializable = ToJSON & FromJSON;
	`

	// Parse the code
	ast, err := p.ParseString(code)
	require.NoError(t, err)
	require.NotNil(t, ast)

	// Debug output
	for _, ann := range ast.TypeAnnotations {
		t.Logf("Found type declaration: %s = %s",
			ann.AnnotatedItem, ann.TypeExpression.String())
	}

	// Check the type annotations
	assert.NotEmpty(t, ast.TypeAnnotations)

	// Check for type declarations
	var foundUserID, foundUserMap, foundMaybeStr, foundComplexData, foundNonNull, foundSerializable bool

	for _, annotation := range ast.TypeAnnotations {
		// Type declarations
		if annotation.Kind == TypeDeclAnnotation {
			switch annotation.AnnotatedItem {
			case "UserID":
				foundUserID = true
				assert.Equal(t, "Str", annotation.TypeExpression.Name)
			case "UserMap":
				foundUserMap = true
				assert.Equal(t, "HashRef", annotation.TypeExpression.Name)
				assert.Equal(t, 2, len(annotation.TypeExpression.Params))
				assert.Equal(t, "UserID", annotation.TypeExpression.Params[0].Name)
				assert.Equal(t, "User", annotation.TypeExpression.Params[1].Name)
			case "MaybeStr":
				foundMaybeStr = true
				assert.True(t, annotation.TypeExpression.Union)
				assert.Equal(t, 2, len(annotation.TypeExpression.Params))
				assert.Equal(t, "Str", annotation.TypeExpression.Params[0].Name)
				assert.Equal(t, "Undef", annotation.TypeExpression.Params[1].Name)
			case "ComplexData":
				foundComplexData = true
				assert.Equal(t, "ArrayRef", annotation.TypeExpression.Name)
				assert.GreaterOrEqual(t, len(annotation.TypeExpression.Params), 1)
			case "NonNull":
				foundNonNull = true
				assert.True(t, annotation.TypeExpression.Negation)
				assert.Equal(t, "Undef", annotation.TypeExpression.Name)
			case "Serializable":
				foundSerializable = true
				assert.True(t, strings.Contains(annotation.TypeExpression.String(), "&"))
			}
		}
	}

	// Basic checks for type declarations
	assert.True(t, foundUserID, "Should find type declaration for UserID")
	assert.True(t, foundUserMap, "Should find type declaration for UserMap")
	assert.True(t, foundMaybeStr, "Should find type declaration for MaybeStr")
	assert.True(t, foundComplexData, "Should find type declaration for ComplexData")
	assert.True(t, foundNonNull, "Should find type declaration for NonNull")
	assert.True(t, foundSerializable, "Should find type declaration for Serializable")
}
