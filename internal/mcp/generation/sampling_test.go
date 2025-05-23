package generation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSamplingClient_Sample(t *testing.T) {
	// Create enabled client
	client := NewSamplingClient(true)

	ctx := context.Background()

	tests := []struct {
		name           string
		prompt         string
		expectedInResp string
	}{
		{
			"missing sigil fix",
			"Fix missing sigil error",
			"my $",
		},
		{
			"type mismatch fix",
			"Fix type mismatch error",
			"Int",
		},
		{
			"undefined variable fix",
			"Fix undefined variable error",
			"my $",
		},
		{
			"function generation",
			"generate function",
			"sub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Sample(ctx, tt.prompt, "/test/project")
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Contains(t, resp.Content, tt.expectedInResp)
			assert.Greater(t, resp.Confidence, 0.0)
			assert.NotEmpty(t, resp.Model)
		})
	}
}

func TestSamplingClient_Disabled(t *testing.T) {
	// Create disabled client
	client := NewSamplingClient(false)

	ctx := context.Background()

	// Should return error when disabled
	_, err := client.Sample(ctx, "test prompt", "/test/project")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

func TestSamplingClient_SampleWithOptions(t *testing.T) {
	client := NewSamplingClient(true)
	ctx := context.Background()

	tests := []struct {
		name      string
		request   *SamplingRequest
		expectErr bool
	}{
		{
			"valid request",
			&SamplingRequest{
				Prompt:      "Generate a function",
				MaxTokens:   500,
				Temperature: 0.5,
			},
			false,
		},
		{
			"empty prompt",
			&SamplingRequest{
				Prompt: "",
			},
			true,
		},
		{
			"default values",
			&SamplingRequest{
				Prompt: "Test prompt",
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.SampleWithOptions(ctx, tt.request)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestSamplingClient_BatchSample(t *testing.T) {
	client := NewSamplingClient(true)
	ctx := context.Background()

	prompts := []string{
		"Fix missing sigil",
		"Generate function",
		"Fix type mismatch",
	}

	responses, err := client.BatchSample(ctx, prompts, "/test/project")
	require.NoError(t, err)
	assert.Len(t, responses, len(prompts))

	for i, resp := range responses {
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Content)
		assert.Greater(t, resp.Confidence, 0.0)
		t.Logf("Response %d: %s", i, resp.Content)
	}
}

func TestCreateSamplingMessage(t *testing.T) {
	msg := CreateSamplingMessage("Test prompt", nil)
	assert.NotNil(t, msg)
	assert.Equal(t, "sampling/createMessage", msg["method"])
	assert.NotNil(t, msg["params"])
}
