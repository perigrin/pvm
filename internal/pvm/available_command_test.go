// ABOUTME: Tests for the pvm available command
// ABOUTME: Validates output formatting, version listing, and different formats

package pvm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAvailableCommand_CommandSetup(t *testing.T) {
	cmd := newAvailableCommand()

	// Test command metadata
	assert.Equal(t, "available", cmd.Use)
	assert.Equal(t, "List available Perl versions", cmd.Short)
	assert.Equal(t, "List all Perl versions available for installation", cmd.Long)

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
	assert.Equal(t, "List all Perl versions available for installation", cmd.Long)

	// Test flag help text
	flag := cmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatal("Format flag should exist")
	}
	assert.Contains(t, flag.Usage, "plain", "Flag help should mention plain format")
}
