// ABOUTME: Tests for the dev command stub
// ABOUTME: Validates that the stub correctly reports "not yet available"

package pvm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDevCommand(t *testing.T) {
	cmd := newDevCommand()

	assert.Equal(t, "dev", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	// Command should fail with "not yet available" message
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not yet available")
}
