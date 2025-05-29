package validation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_ValidateCode(t *testing.T) {
	// Create cache
	cache, err := NewValidationCache("10MB")
	require.NoError(t, err)

	// Create validator
	validator, err := NewValidator(cache)
	require.NoError(t, err)

	tests := []struct {
		name         string
		code         string
		expectValid  bool
		expectErrors int
		expectTypes  int
	}{
		{
			name: "valid typed code",
			code: `#!/usr/bin/perl
use strict;
use warnings;

my Int $count = 42;
my Str $name = "test";
`,
			expectValid:  true,
			expectErrors: 0,
			expectTypes:  2,
		},
		{
			name: "code with type error",
			code: `#!/usr/bin/perl
my NonExistentType $var = 42;
`,
			expectValid:  true, // Step 3 doesn't validate unknown types yet
			expectErrors: 0,
			expectTypes:  1, // But it does extract the annotation
		},
		{
			name: "code with missing sigil",
			code: `#!/usr/bin/perl
my variable = 42;
`,
			expectValid:  true, // Step 3 basic parsing doesn't catch this yet
			expectErrors: 0,
			expectTypes:  0,
		},
		{
			name:         "empty code",
			code:         "",
			expectValid:  true,
			expectErrors: 0,
			expectTypes:  0,
		},
		{
			name: "complex types",
			code: `#!/usr/bin/perl
my ArrayRef[Int] $numbers = [1, 2, 3];
my HashRef[Str] $config = { name => "test" };
`,
			expectValid:  true,
			expectErrors: 0,
			expectTypes:  2, // Parser now successfully extracts complex types
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateCode(ctx, tt.code, "/test/project")
			require.NoError(t, err)

			assert.Equal(t, tt.expectValid, result.Valid)
			assert.Equal(t, tt.expectErrors, len(result.Errors))
			assert.Equal(t, tt.expectTypes, len(result.TypeInfo))
		})
	}
}

func TestValidator_Caching(t *testing.T) {
	// Create cache
	cache, err := NewValidationCache("1MB")
	require.NoError(t, err)

	// Create validator
	validator, err := NewValidator(cache)
	require.NoError(t, err)

	ctx := context.Background()
	code := `my Int $x = 42;`

	// First validation - should miss cache
	result1, err := validator.ValidateCode(ctx, code, "/test/project")
	require.NoError(t, err)
	assert.NotNil(t, result1)

	// Second validation - should hit cache
	result2, err := validator.ValidateCode(ctx, code, "/test/project")
	require.NoError(t, err)
	assert.NotNil(t, result2)

	// Results should be the same
	assert.Equal(t, result1.Valid, result2.Valid)
	assert.Equal(t, len(result1.Errors), len(result2.Errors))
}

func TestValidator_GetTypeInfo(t *testing.T) {
	// Create cache
	cache, err := NewValidationCache("1MB")
	require.NoError(t, err)

	// Create validator
	validator, err := NewValidator(cache)
	require.NoError(t, err)

	// GetTypeInfo is not fully implemented in Step 3
	// It will be implemented in Step 4 with the analysis tool
	_, err = validator.GetTypeInfo("count")
	assert.Error(t, err)
}
