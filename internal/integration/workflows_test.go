// ABOUTME: Tests for end-to-end workflow integration stubs
// ABOUTME: Validates that stub workflows correctly return "not yet available" errors

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/pvx"
)

func TestCompleteWorkflowStub(t *testing.T) {
	options := &WorkflowOptions{
		ScriptPath:     "/some/script.pl",
		PerlVersion:    "",
		Verbose:        false,
		IsolationLevel: pvx.IsolationLocal,
	}

	result, err := CompleteWorkflow(options)
	require.Error(t, err, "CompleteWorkflow should return error in stub build")
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not yet available")
}

func TestTypeCheckWorkflowStub(t *testing.T) {
	result, err := TypeCheckWorkflow("/some/script.pl", "", false)
	require.Error(t, err, "TypeCheckWorkflow should return error in stub build")
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not yet available")
}

func TestExecutionWorkflowStub(t *testing.T) {
	result, err := ExecutionWorkflow("/some/script.pl", "", false)
	require.Error(t, err, "ExecutionWorkflow should return error in stub build")
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not yet available")
}

func TestValidationWorkflowStub(t *testing.T) {
	result, err := ValidationWorkflow("")
	require.Error(t, err, "ValidationWorkflow should return error in stub build")
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not yet available")
}
