// ABOUTME: Tests for the pvm available command
// ABOUTME: Validates output formatting, version listing, and different formats

package pvm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAvailableCommand_CommandSetup(t *testing.T) {
	cmd := newAvailableCommand()

	// Test command metadata
	assert.Equal(t, "available", cmd.Use)
	assert.Equal(t, "List available Perl versions", cmd.Short)
	assert.Contains(t, cmd.Long, "List all Perl versions available for installation")

	// Test that RunE function is set
	assert.NotNil(t, cmd.RunE, "RunE function should be set")
}

func TestAvailableCommand_FormatFlag(t *testing.T) {
	cmd := newAvailableCommand()

	// Test that format flag exists
	flag := cmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatal("Format flag should exist")
	}

	// Test default value
	assert.Equal(t, "text", flag.DefValue, "Default format should be 'text'")

	// Test short flag
	shortFlag := cmd.Flags().ShorthandLookup("f")
	if shortFlag == nil {
		t.Fatal("Short format flag should exist")
	}
	assert.Equal(t, flag, shortFlag, "Short flag should be the same as long flag")
}

func TestAvailableCommand_HelpText(t *testing.T) {
	cmd := newAvailableCommand()

	// Test command metadata
	assert.Equal(t, "available", cmd.Use)
	assert.Equal(t, "List available Perl versions", cmd.Short)
	assert.Contains(t, cmd.Long, "List all Perl versions available for installation")

	// Test flag help text
	flag := cmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatal("Format flag should exist")
	}
	assert.Contains(t, flag.Usage, "plain", "Flag help should mention plain format")
}

func TestAvailableCommand_FormatValidation(t *testing.T) {
	cmd := newAvailableCommand()

	// Test valid formats
	validFormats := []string{"text", "json", "plain"}
	for _, format := range validFormats {
		cmd.SetArgs([]string{"--format", format})
		err := cmd.ParseFlags([]string{"--format", format})
		assert.NoError(t, err, "Valid format %s should not cause parse error", format)
	}

	// Test invalid format (we can't easily test execution without network calls,
	// but we can test that the flag accepts the value)
	cmd.SetArgs([]string{"--format", "invalid"})
	err := cmd.ParseFlags([]string{"--format", "invalid"})
	assert.NoError(t, err, "Flag parsing should succeed even for invalid format values")
}

func TestAvailableCommand_RemoteFlag_Registered(t *testing.T) {
	cmd := newAvailableCommand()

	// The --remote flag must be registered.
	flag := cmd.Flags().Lookup("remote")
	require.NotNil(t, flag, "--remote flag should be registered on the available command")

	// Default value should be empty string (no remote).
	assert.Equal(t, "", flag.DefValue, "--remote default should be empty string")

	// Flag type should be string.
	assert.Equal(t, "string", flag.Value.Type(), "--remote should be a string flag")
}

func TestAvailableCommand_RemoteFlag_ParsesWithoutError(t *testing.T) {
	cmd := newAvailableCommand()

	// Parsing --remote with a value must not return an error.
	err := cmd.ParseFlags([]string{"--remote", "mycompany"})
	assert.NoError(t, err, "--remote flag should parse without error")

	val, err := cmd.Flags().GetString("remote")
	require.NoError(t, err)
	assert.Equal(t, "mycompany", val, "--remote should hold the parsed value")
}

func TestAvailableCommand_RemoteFlag_EmptyByDefault(t *testing.T) {
	cmd := newAvailableCommand()

	// When --remote is not provided, the value must be empty.
	err := cmd.ParseFlags([]string{})
	require.NoError(t, err)

	val, err := cmd.Flags().GetString("remote")
	require.NoError(t, err)
	assert.Equal(t, "", val, "--remote should be empty when not provided")
}
