// ABOUTME: Tests for test execution mode helpers and environment-based configuration
// ABOUTME: Validates behavior of ShouldMockExternalAPIs and other mode detection functions

package basetesting

import (
	"os"
	"testing"
)

func TestGetTestMode(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected TestMode
	}{
		{"unit mode", "unit", ModeUnit},
		{"integration mode", "integration", ModeIntegration},
		{"performance mode", "performance", ModePerformance},
		{"stress mode", "stress", ModeStress},
		{"full mode", "full", ModeFull},
		{"uppercase unit", "UNIT", ModeUnit},
		{"mixed case integration", "Integration", ModeIntegration},
		{"empty env defaults to integration", "", ModeIntegration},
		{"invalid env defaults to integration", "invalid", ModeIntegration},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env
			original := os.Getenv("PVM_TEST_MODE")
			defer os.Setenv("PVM_TEST_MODE", original)

			// Set test env
			os.Setenv("PVM_TEST_MODE", tt.envValue)

			result := GetTestMode()
			if result != tt.expected {
				t.Errorf("GetTestMode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestShouldRunPerformanceTests(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		expected bool
	}{
		{"unit mode should not run performance", "unit", false},
		{"integration mode should not run performance", "integration", false},
		{"performance mode should run performance", "performance", true},
		{"stress mode should not run performance", "stress", false},
		{"full mode should run performance", "full", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := os.Getenv("PVM_TEST_MODE")
			defer os.Setenv("PVM_TEST_MODE", original)

			os.Setenv("PVM_TEST_MODE", tt.mode)

			result := ShouldRunPerformanceTests()
			if result != tt.expected {
				t.Errorf("ShouldRunPerformanceTests() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestShouldRunStressTests(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		expected bool
	}{
		{"unit mode should not run stress", "unit", false},
		{"integration mode should not run stress", "integration", false},
		{"performance mode should not run stress", "performance", false},
		{"stress mode should run stress", "stress", true},
		{"full mode should run stress", "full", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := os.Getenv("PVM_TEST_MODE")
			defer os.Setenv("PVM_TEST_MODE", original)

			os.Setenv("PVM_TEST_MODE", tt.mode)

			result := ShouldRunStressTests()
			if result != tt.expected {
				t.Errorf("ShouldRunStressTests() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestShouldRunIntegrationTests(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		expected bool
	}{
		{"unit mode should not run integration", "unit", false},
		{"integration mode should run integration", "integration", true},
		{"performance mode should run integration", "performance", true},
		{"stress mode should run integration", "stress", true},
		{"full mode should run integration", "full", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := os.Getenv("PVM_TEST_MODE")
			defer os.Setenv("PVM_TEST_MODE", original)

			os.Setenv("PVM_TEST_MODE", tt.mode)

			result := ShouldRunIntegrationTests()
			if result != tt.expected {
				t.Errorf("ShouldRunIntegrationTests() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestShouldRunLongRunningTests(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		expected bool
	}{
		{"unit mode should not run long running", "unit", false},
		{"integration mode should not run long running", "integration", false},
		{"performance mode should run long running", "performance", true},
		{"stress mode should run long running", "stress", true},
		{"full mode should run long running", "full", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := os.Getenv("PVM_TEST_MODE")
			defer os.Setenv("PVM_TEST_MODE", original)

			os.Setenv("PVM_TEST_MODE", tt.mode)

			result := ShouldRunLongRunningTests()
			if result != tt.expected {
				t.Errorf("ShouldRunLongRunningTests() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestShouldMockExternalAPIs(t *testing.T) {
	tests := []struct {
		name        string
		mockEnv     string
		ciEnv       string
		testMode    string
		expected    bool
		description string
	}{
		{
			name:        "explicit mock true",
			mockEnv:     "true",
			ciEnv:       "",
			testMode:    "integration",
			expected:    true,
			description: "explicit PVM_MOCK_EXTERNAL_APIS=true should enable mocking",
		},
		{
			name:        "explicit mock 1",
			mockEnv:     "1",
			ciEnv:       "",
			testMode:    "integration",
			expected:    true,
			description: "explicit PVM_MOCK_EXTERNAL_APIS=1 should enable mocking",
		},
		{
			name:        "explicit mock TRUE uppercase",
			mockEnv:     "TRUE",
			ciEnv:       "",
			testMode:    "integration",
			expected:    true,
			description: "explicit PVM_MOCK_EXTERNAL_APIS=TRUE should enable mocking",
		},
		{
			name:        "explicit mock false",
			mockEnv:     "false",
			ciEnv:       "",
			testMode:    "integration",
			expected:    false,
			description: "explicit PVM_MOCK_EXTERNAL_APIS=false should disable mocking",
		},
		{
			name:        "explicit mock 0",
			mockEnv:     "0",
			ciEnv:       "",
			testMode:    "integration",
			expected:    false,
			description: "explicit PVM_MOCK_EXTERNAL_APIS=0 should disable mocking",
		},
		{
			name:        "explicit mock invalid",
			mockEnv:     "invalid",
			ciEnv:       "",
			testMode:    "integration",
			expected:    false,
			description: "explicit PVM_MOCK_EXTERNAL_APIS=invalid should disable mocking",
		},
		{
			name:        "CI environment enables mocking",
			mockEnv:     "",
			ciEnv:       "true",
			testMode:    "integration",
			expected:    true,
			description: "CI environment should enable mocking by default",
		},
		{
			name:        "CI environment with explicit false",
			mockEnv:     "false",
			ciEnv:       "true",
			testMode:    "integration",
			expected:    false,
			description: "explicit false should override CI environment",
		},
		{
			name:        "unit mode without CI",
			mockEnv:     "",
			ciEnv:       "",
			testMode:    "unit",
			expected:    false,
			description: "unit mode should not mock external APIs",
		},
		{
			name:        "integration mode without CI",
			mockEnv:     "",
			ciEnv:       "",
			testMode:    "integration",
			expected:    true,
			description: "integration mode should mock external APIs by default",
		},
		{
			name:        "performance mode without CI",
			mockEnv:     "",
			ciEnv:       "",
			testMode:    "performance",
			expected:    true,
			description: "performance mode should mock external APIs by default",
		},
		{
			name:        "stress mode without CI",
			mockEnv:     "",
			ciEnv:       "",
			testMode:    "stress",
			expected:    true,
			description: "stress mode should mock external APIs by default",
		},
		{
			name:        "full mode without CI",
			mockEnv:     "",
			ciEnv:       "",
			testMode:    "full",
			expected:    true,
			description: "full mode should mock external APIs by default",
		},
		{
			name:        "default mode without CI",
			mockEnv:     "",
			ciEnv:       "",
			testMode:    "",
			expected:    true,
			description: "default mode (integration) should mock external APIs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env values
			originalMock := os.Getenv("PVM_MOCK_EXTERNAL_APIS")
			originalCI := os.Getenv("CI")
			originalMode := os.Getenv("PVM_TEST_MODE")

			// Restore original values after test
			defer func() {
				os.Setenv("PVM_MOCK_EXTERNAL_APIS", originalMock)
				os.Setenv("CI", originalCI)
				os.Setenv("PVM_TEST_MODE", originalMode)
			}()

			// Set test environment variables
			os.Setenv("PVM_MOCK_EXTERNAL_APIS", tt.mockEnv)
			os.Setenv("CI", tt.ciEnv)
			os.Setenv("PVM_TEST_MODE", tt.testMode)

			result := ShouldMockExternalAPIs()
			if result != tt.expected {
				t.Errorf("ShouldMockExternalAPIs() = %v, want %v\nDescription: %s", result, tt.expected, tt.description)
			}
		})
	}
}

func TestGetSampleRate(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected float64
	}{
		{"empty env defaults to 1.0", "", 1.0},
		{"valid rate 0.5", "0.5", 0.5},
		{"valid rate 0.1", "0.1", 0.1},
		{"valid rate 1.0", "1.0", 1.0},
		{"valid rate 0.0", "0.0", 0.0},
		{"invalid rate negative", "-0.5", 1.0},
		{"invalid rate greater than 1", "1.5", 1.0},
		{"invalid rate non-numeric", "invalid", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := os.Getenv("PVM_TEST_SAMPLE_RATE")
			defer os.Setenv("PVM_TEST_SAMPLE_RATE", original)

			os.Setenv("PVM_TEST_SAMPLE_RATE", tt.envValue)

			result := GetSampleRate()
			if result != tt.expected {
				t.Errorf("GetSampleRate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSkipFunctions tests the skip functions by verifying their logic through the underlying Should* functions
// Since the skip functions require *testing.T specifically, we test their behavior indirectly
func TestSkipFunctionLogic(t *testing.T) {
	// Test SkipUnlessPerformance logic
	t.Run("SkipUnlessPerformance logic", func(t *testing.T) {
		tests := []struct {
			mode       string
			shouldSkip bool
		}{
			{"unit", true},
			{"integration", true},
			{"performance", false},
			{"stress", true},
			{"full", false},
		}

		for _, tt := range tests {
			t.Run(tt.mode, func(t *testing.T) {
				original := os.Getenv("PVM_TEST_MODE")
				defer os.Setenv("PVM_TEST_MODE", original)

				os.Setenv("PVM_TEST_MODE", tt.mode)
				shouldRun := ShouldRunPerformanceTests()
				shouldSkip := !shouldRun

				if shouldSkip != tt.shouldSkip {
					t.Errorf("Performance test skip logic: shouldSkip = %v, want %v", shouldSkip, tt.shouldSkip)
				}
			})
		}
	})

	// Test SkipUnlessStress logic
	t.Run("SkipUnlessStress logic", func(t *testing.T) {
		tests := []struct {
			mode       string
			shouldSkip bool
		}{
			{"unit", true},
			{"integration", true},
			{"performance", true},
			{"stress", false},
			{"full", false},
		}

		for _, tt := range tests {
			t.Run(tt.mode, func(t *testing.T) {
				original := os.Getenv("PVM_TEST_MODE")
				defer os.Setenv("PVM_TEST_MODE", original)

				os.Setenv("PVM_TEST_MODE", tt.mode)
				shouldRun := ShouldRunStressTests()
				shouldSkip := !shouldRun

				if shouldSkip != tt.shouldSkip {
					t.Errorf("Stress test skip logic: shouldSkip = %v, want %v", shouldSkip, tt.shouldSkip)
				}
			})
		}
	})

	// Test SkipUnlessIntegration logic
	t.Run("SkipUnlessIntegration logic", func(t *testing.T) {
		tests := []struct {
			mode       string
			shouldSkip bool
		}{
			{"unit", true},
			{"integration", false},
			{"performance", false},
			{"stress", false},
			{"full", false},
		}

		for _, tt := range tests {
			t.Run(tt.mode, func(t *testing.T) {
				original := os.Getenv("PVM_TEST_MODE")
				defer os.Setenv("PVM_TEST_MODE", original)

				os.Setenv("PVM_TEST_MODE", tt.mode)
				shouldRun := ShouldRunIntegrationTests()
				shouldSkip := !shouldRun

				if shouldSkip != tt.shouldSkip {
					t.Errorf("Integration test skip logic: shouldSkip = %v, want %v", shouldSkip, tt.shouldSkip)
				}
			})
		}
	})

	// Test SkipUnlessLongRunning logic
	t.Run("SkipUnlessLongRunning logic", func(t *testing.T) {
		tests := []struct {
			mode       string
			shouldSkip bool
		}{
			{"unit", true},
			{"integration", true},
			{"performance", false},
			{"stress", false},
			{"full", false},
		}

		for _, tt := range tests {
			t.Run(tt.mode, func(t *testing.T) {
				original := os.Getenv("PVM_TEST_MODE")
				defer os.Setenv("PVM_TEST_MODE", original)

				os.Setenv("PVM_TEST_MODE", tt.mode)
				shouldRun := ShouldRunLongRunningTests()
				shouldSkip := !shouldRun

				if shouldSkip != tt.shouldSkip {
					t.Errorf("Long-running test skip logic: shouldSkip = %v, want %v", shouldSkip, tt.shouldSkip)
				}
			})
		}
	})
}
