// ABOUTME: Tests for the pvm remote subcommands (add, list, remove)
// ABOUTME: Validates command structure, argument handling, and error cases

package pvm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteCommand_Structure(t *testing.T) {
	cmd := newRemoteCommand()

	assert.Equal(t, "remote", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	// Verify subcommands are present
	subNames := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subNames[sub.Name()] = true
	}
	assert.True(t, subNames["add"], "remote should have 'add' subcommand")
	assert.True(t, subNames["list"], "remote should have 'list' subcommand")
	assert.True(t, subNames["remove"], "remote should have 'remove' subcommand")
}

func TestRemoteAddCommand_Structure(t *testing.T) {
	cmd := newRemoteAddCommand()

	assert.Equal(t, "add <name> <url>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestRemoteAddCommand_RequiresTwoArgs(t *testing.T) {
	cmd := newRemoteAddCommand()

	// No args
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "add should require exactly 2 args")

	// One arg
	err = cmd.Args(cmd, []string{"myremote"})
	assert.Error(t, err, "add should require exactly 2 args")

	// Two args — ok
	err = cmd.Args(cmd, []string{"myremote", "https://example.com/perl.git"})
	assert.NoError(t, err, "add should accept exactly 2 args")

	// Three args — too many
	err = cmd.Args(cmd, []string{"myremote", "https://example.com/perl.git", "extra"})
	assert.Error(t, err, "add should not accept 3 args")
}

func TestRemoteListCommand_Structure(t *testing.T) {
	cmd := newRemoteListCommand()

	assert.Equal(t, "list", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestRemoteRemoveCommand_Structure(t *testing.T) {
	cmd := newRemoteRemoveCommand()

	assert.Equal(t, "remove <name>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestRemoteRemoveCommand_RequiresOneArg(t *testing.T) {
	cmd := newRemoteRemoveCommand()

	// No args
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "remove should require exactly 1 arg")

	// One arg — ok
	err = cmd.Args(cmd, []string{"myremote"})
	assert.NoError(t, err)

	// Two args — too many
	err = cmd.Args(cmd, []string{"myremote", "extra"})
	assert.Error(t, err, "remove should not accept 2 args")
}

func TestRemoteCommand_IsWiredIntoRootCommand(t *testing.T) {
	root := NewCommand()

	subNames := make(map[string]bool)
	for _, sub := range root.Commands() {
		subNames[sub.Name()] = true
	}
	require.True(t, subNames["remote"], "root pvm command should include 'remote' subcommand")
}
