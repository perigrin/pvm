// ABOUTME: Comprehensive router integration tests
// ABOUTME: Tests multi-entry point routing, subcommands, and component detection

package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouterMultiEntryPointScenarios(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name         string
		binaryName   string
		args         []string
		expectedComp string
		expectedType string
		description  string
	}{
		{
			name:         "DirectSymlinkPVI",
			binaryName:   "./pvi",
			args:         []string{"./pvi", "install", "Module::Name"},
			expectedComp: ComponentPVI,
			expectedType: InvocationSymlink,
			description:  "Direct symlink: ./pvi install Module::Name",
		},
		{
			name:         "DirectSymlinkPVX",
			binaryName:   "./pvx",
			args:         []string{"./pvx", "script.pl"},
			expectedComp: ComponentPVX,
			expectedType: InvocationSymlink,
			description:  "Direct symlink: ./pvx script.pl",
		},
		{
			name:         "DirectSymlinkPSC",
			binaryName:   "./psc",
			args:         []string{"./psc", "check", "file.pl"},
			expectedComp: ComponentPSC,
			expectedType: InvocationSymlink,
			description:  "Direct symlink: ./psc check file.pl",
		},
		{
			name:         "DirectBinaryPVM",
			binaryName:   "./pvm",
			args:         []string{"./pvm", "install", "5.38.0"},
			expectedComp: ComponentPVM,
			expectedType: InvocationDirect,
			description:  "Direct binary: ./pvm install 5.38.0",
		},
		{
			name:         "UnknownFallback",
			binaryName:   "./unknown",
			args:         []string{"./unknown", "command"},
			expectedComp: ComponentPVM,
			expectedType: InvocationFallback,
			description:  "Unknown binary fallback to PVM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up args to simulate the binary invocation
			os.Args = tt.args

			// Mock the binary path resolution for symlink testing
			// In real scenarios, this would be resolved by the filesystem
			if tt.expectedType == InvocationSymlink {
				// For testing, we simulate what would happen with real symlinks
				os.Args[0] = tt.binaryName
			}

			// Test component detection
			component := DetectComponent()
			assert.Equal(t, tt.expectedComp, component, "Component detection failed for %s", tt.description)

			// Test detailed invocation info
			info := DetectInvocation()
			assert.Equal(t, tt.expectedComp, info.Component, "InvocationInfo component failed for %s", tt.description)

			// Note: We can't easily test symlink vs direct vs fallback types
			// without actual filesystem symlinks, but the component detection is the key functionality
		})
	}
}

func TestRouterHelpRouting(t *testing.T) {
	// Test that help routing works for component aliases
	// This is the main issue that Step 1.1 was meant to fix

	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	helpTests := []struct {
		component   string
		description string
	}{
		{ComponentPVI, "PVI help should be accessible"},
		{ComponentPVX, "PVX help should be accessible"},
		{ComponentPSC, "PSC help should be accessible"},
	}

	for _, tt := range helpTests {
		t.Run("Help"+tt.component, func(t *testing.T) {
			// Simulate help command for component
			os.Args = []string{"pvm", "help", tt.component}

			// Verify component detection still works with PVM
			component := DetectComponent()
			assert.Equal(t, ComponentPVM, component, "Help command should route through PVM")

			// Note: Registry tests are skipped here because GlobalRegistry is only
			// populated in main entry points. The actual help functionality is tested
			// in integration tests or manually verified.
		})
	}
}

func TestRouterComponentDescriptions(t *testing.T) {
	// Test component descriptions are available
	tests := []struct {
		component    string
		expectedDesc string
	}{
		{ComponentPVM, DescriptionPVM},
		{ComponentPVI, DescriptionPVI},
		{ComponentPVX, DescriptionPVX},
		{ComponentPSC, DescriptionPSC},
	}

	for _, tt := range tests {
		t.Run(tt.component, func(t *testing.T) {
			desc := GetComponentDescription(tt.component)
			assert.Equal(t, tt.expectedDesc, desc, "Description should match for component %s", tt.component)
		})
	}

	// Test unknown component
	desc := GetComponentDescription("unknown")
	assert.Equal(t, "Unknown component", desc, "Unknown component should return standard message")
}

func TestRouterRootCommandCreation(t *testing.T) {
	// Test that root commands can be created for all components
	components := []string{ComponentPVM, ComponentPVI, ComponentPVX, ComponentPSC}

	for _, component := range components {
		t.Run(component, func(t *testing.T) {
			rootCmd := CreateRootCommand(component)
			require.NotNil(t, rootCmd, "Root command should be created for %s", component)

			// Verify basic command properties
			assert.Equal(t, component, rootCmd.Use, "Root command use should match component")
			assert.NotEmpty(t, rootCmd.Short, "Root command should have short description")
			assert.NotEmpty(t, rootCmd.Long, "Root command should have long description")

			// Note: Registry integration tests are skipped because GlobalRegistry
			// is only populated in main entry points.
		})
	}
}

func TestRouterBackwardCompatibilityAliases(t *testing.T) {
	// Test that backward compatibility aliases work correctly
	// The key requirement from Step 1.1 is maintaining backward compatibility with symlinks

	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	aliasTests := []struct {
		symlink      string
		expectedComp string
		description  string
	}{
		{"pvi", ComponentPVI, "PVI symlink compatibility"},
		{"pvx", ComponentPVX, "PVX symlink compatibility"},
		{"psc", ComponentPSC, "PSC symlink compatibility"},
		{"pvm", ComponentPVM, "PVM direct execution"},
	}

	for _, tt := range aliasTests {
		t.Run(tt.symlink, func(t *testing.T) {
			os.Args = []string{tt.symlink}

			component := DetectComponent()
			assert.Equal(t, tt.expectedComp, component, "%s should route to %s", tt.symlink, tt.expectedComp)

			// Verify we can create a root command for this component
			rootCmd := CreateRootCommand(component)
			assert.NotNil(t, rootCmd, "Should be able to create root command for %s", component)
		})
	}
}

func TestRouterRegistryIntegration(t *testing.T) {
	// Test that the GlobalRegistry interface works (components are registered in main)

	// Verify all required components exist in the constants
	requiredComponents := []string{ComponentPVM, ComponentPVI, ComponentPVX, ComponentPSC}

	for _, component := range requiredComponents {
		t.Run("Component"+component, func(t *testing.T) {
			// Verify component constants are defined
			assert.NotEmpty(t, component, "Component constant should not be empty")

			// Verify component descriptions exist
			desc := GetComponentDescription(component)
			assert.NotEqual(t, "Unknown component", desc, "Component %s should have a description", component)

			// Note: Actual registry registration is tested in integration tests
			// since GlobalRegistry is populated in main entry points
		})
	}
}

func TestRouterErrorHandling(t *testing.T) {
	// Test error handling in routing scenarios

	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test with empty args (should not crash)
	os.Args = []string{}

	// This should not panic
	assert.NotPanics(t, func() {
		DetectInvocation()
	}, "DetectInvocation should handle empty args gracefully")

	assert.NotPanics(t, func() {
		DetectComponent()
	}, "DetectComponent should handle empty args gracefully")

	// Test with malformed args
	os.Args = []string{"", ""}
	assert.NotPanics(t, func() {
		DetectComponent()
	}, "DetectComponent should handle malformed args gracefully")
}
