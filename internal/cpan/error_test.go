// ABOUTME: Tests for CPAN error handling
// ABOUTME: Validates error generation and handling for CPAN operations

package cpan

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProviderErrorImplementsError tests that ProviderError implements the error interface
func TestProviderErrorImplementsError(t *testing.T) {
	// Create a provider error
	err := &ProviderError{
		Source:  "test",
		Code:    "101",
		Message: "Test error",
		URL:     "https://example.com",
		Err:     fmt.Errorf("underlying error"),
	}

	// Check that it implements the error interface
	var _ error = err // This will fail to compile if ProviderError doesn't implement error

	// Check the error message
	assert.Equal(t, "test error: Test error (https://example.com)", err.Error(), "Error message should be formatted correctly")

	// Check the unwrap functionality
	unwrapped := err.Unwrap()
	assert.EqualError(t, unwrapped, "underlying error", "Unwrapped error should match the underlying error")
}

// TestProviderErrorWithoutURL tests that ProviderError works without a URL
func TestProviderErrorWithoutURL(t *testing.T) {
	// Create a provider error without a URL
	err := &ProviderError{
		Source:  "test",
		Code:    "101",
		Message: "Test error",
		Err:     fmt.Errorf("underlying error"),
	}

	// Check the error message
	assert.Equal(t, "test error: Test error", err.Error(), "Error message should be formatted correctly")
}

// TestNetworkDisabledError tests that operations fail with the correct error when network is disabled
func TestNetworkDisabledError(t *testing.T) {
	// Create a provider with network disabled
	provider, err := NewMetaCPANProvider(WithDisableNetwork(true))
	require.NoError(t, err, "NewMetaCPANProvider should not return an error")
	require.NotNil(t, provider, "Provider should not be nil")

	// Try to get module info
	_, err = provider.GetModuleInfo(context.Background(), "Test::Module")
	require.Error(t, err, "GetModuleInfo should return an error when network is disabled")

	// Check that the error is a ProviderError
	var providerErr *ProviderError
	switch e := err.(type) {
	case *ProviderError:
		providerErr = e
	default:
		t.Fatalf("Error should be a ProviderError, got %T", err)
	}
	assert.Equal(t, "metacpan", providerErr.Source, "Error source should be 'metacpan'")
	assert.Equal(t, "network_disabled", providerErr.Code, "Error code should be 'network_disabled'")
	assert.Equal(t, "Network access is disabled", providerErr.Message, "Error message should be correct")
}

// TestHTTPErrorHandling tests that HTTP errors are handled correctly
func TestHTTPErrorHandling(t *testing.T) {
	// Create a test server that returns a 404 error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a 404 Not Found error
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	// Create a provider that uses the test server
	provider, err := NewMetaCPANProvider()
	require.NoError(t, err, "NewMetaCPANProvider should not return an error")
	require.NotNil(t, provider, "Provider should not be nil")

	// Directly set the baseURL on the provider to ensure it's used
	provider.baseURL = server.URL

	// Try to get module info
	_, err = provider.GetModuleInfo(context.Background(), "Test::Module")
	require.Error(t, err, "GetModuleInfo should return an error when the server returns an error")

	// Check that the error is a ProviderError
	var providerErr *ProviderError
	switch e := err.(type) {
	case *ProviderError:
		providerErr = e
	default:
		t.Fatalf("Error should be a ProviderError, got %T", err)
	}
	assert.Equal(t, "metacpan", providerErr.Source, "Error source should be 'metacpan'")
	assert.Equal(t, "http_error", providerErr.Code, "Error code should be 'http_error'")
	assert.Equal(t, fmt.Sprintf("HTTP error: %d", http.StatusNotFound), providerErr.Message, "Error message should be correct")
	assert.Equal(t, http.StatusNotFound, providerErr.StatusCode, "Error status code should be correct")
	assert.Contains(t, providerErr.URL, server.URL, "Error URL should contain the server URL")
}

// TestRequestFailureHandling tests that request failures are handled correctly
func TestRequestFailureHandling(t *testing.T) {
	// Create a provider with an invalid URL
	provider, err := NewMetaCPANProvider(
		WithBaseURL("http://invalid.example.com:9999"),
		WithDisableNetwork(false),
	)
	require.NoError(t, err, "NewMetaCPANProvider should not return an error")
	require.NotNil(t, provider, "Provider should not be nil")

	// Try to get module info
	_, err = provider.GetModuleInfo(context.Background(), "Test::Module")
	require.Error(t, err, "GetModuleInfo should return an error when the request fails")

	// Check that the error is a ProviderError
	var providerErr *ProviderError
	switch e := err.(type) {
	case *ProviderError:
		providerErr = e
	default:
		t.Fatalf("Error should be a ProviderError, got %T", err)
	}
	assert.Equal(t, "metacpan", providerErr.Source, "Error source should be 'metacpan'")
	assert.Equal(t, "request_failed", providerErr.Code, "Error code should be 'request_failed'")
	assert.Contains(t, providerErr.Message, "Failed to execute request", "Error message should be correct")
	assert.Contains(t, providerErr.URL, "invalid.example.com", "Error URL should contain the invalid URL")
	assert.NotNil(t, providerErr.Err, "Error should have an underlying error")
}

// TestParseFailureHandling tests that JSON parsing failures are handled correctly
func TestParseFailureHandling(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return invalid JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("This is not valid JSON"))
	}))
	defer server.Close()

	// Create a provider that uses the test server
	provider, err := NewMetaCPANProvider()
	require.NoError(t, err, "NewMetaCPANProvider should not return an error")
	require.NotNil(t, provider, "Provider should not be nil")

	// Directly set the baseURL on the provider to ensure it's used
	provider.baseURL = server.URL

	// Try to get module info
	_, err = provider.GetModuleInfo(context.Background(), "Test::Module")
	require.Error(t, err, "GetModuleInfo should return an error when the response is invalid JSON")

	// Check that the error is a ProviderError
	var providerErr *ProviderError
	switch e := err.(type) {
	case *ProviderError:
		providerErr = e
	default:
		t.Fatalf("Error should be a ProviderError, got %T", err)
	}
	assert.Equal(t, "metacpan", providerErr.Source, "Error source should be 'metacpan'")
	assert.Equal(t, "parse_failed", providerErr.Code, "Error code should be 'parse_failed'")
	assert.Contains(t, providerErr.Message, "Failed to parse JSON response", "Error message should be correct")
	assert.Contains(t, providerErr.URL, server.URL, "Error URL should contain the server URL")
	assert.NotNil(t, providerErr.Err, "Error should have an underlying error")
}
