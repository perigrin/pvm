// ABOUTME: Tests for Step 1.1 subcommand routing requirements
// ABOUTME: Verifies that pvm subcommands work properly with all components

package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStep1_1_SubcommandRouting(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test cases as specified in Step 1.1 requirements
	testCases := []struct {
		name         string
		args         []string
		expectedComp string
		description  string
	}{
		{
			name:         "DirectSymlinkPVI",
			args:         []string{"./pvi", "install", "Module::Name"},
			expectedComp: ComponentPVI,
			description:  "Direct symlink: ./pvi install Module::Name",
		},
		{
			name:         "SubcommandPVI",
			args:         []string{"./pvm", "pvi", "install", "Module::Name"},
			expectedComp: ComponentPVM, // PVM acts as router for subcommands
			description:  "Subcommand: ./pvm pvi install Module::Name",
		},
		{
			name:         "DirectSymlinkPVX",
			args:         []string{"./pvx", "script.pl"},
			expectedComp: ComponentPVX,
			description:  "Direct symlink: ./pvx script.pl",
		},
		{
			name:         "SubcommandPVX",
			args:         []string{"./pvm", "pvx", "script.pl"},
			expectedComp: ComponentPVM, // PVM acts as router for subcommands
			description:  "Subcommand: ./pvm pvx script.pl",
		},
		{
			name:         "DirectSymlinkPSC",
			args:         []string{"./psc", "check", "file.pl"},
			expectedComp: ComponentPSC,
			description:  "Direct symlink: ./psc check file.pl",
		},
		{
			name:         "SubcommandPSC",
			args:         []string{"./pvm", "psc", "check", "file.pl"},
			expectedComp: ComponentPVM, // PVM acts as router for subcommands
			description:  "Subcommand: ./pvm psc check file.pl",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Args = tc.args

			component := DetectComponent()
			assert.Equal(t, tc.expectedComp, component, "Component detection failed for %s", tc.description)

			// Verify we can create a root command for this component
			rootCmd := CreateRootCommand(component)
			require.NotNil(t, rootCmd, "Should be able to create root command for %s", component)
		})
	}
}

func TestStep1_1_HelpRouting(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test help routing as specified in Step 1.1
	helpTests := []struct {
		component   string
		args        []string
		description string
	}{
		{
			component:   ComponentPVI,
			args:        []string{"./pvm", "help", "pvi"},
			description: "Help: ./pvm help pvi (should show PVI help)",
		},
		{
			component:   ComponentPVX,
			args:        []string{"./pvm", "help", "pvx"},
			description: "Help: ./pvm help pvx (should show PVX help)",
		},
		{
			component:   ComponentPSC,
			args:        []string{"./pvm", "help", "psc"},
			description: "Help: ./pvm help psc (should show PSC help)",
		},
	}

	for _, tt := range helpTests {
		t.Run("Help"+tt.component, func(t *testing.T) {
			os.Args = tt.args

			// For help commands, PVM should be the detected component
			// as it acts as the router
			component := DetectComponent()
			assert.Equal(t, ComponentPVM, component, "%s should route through PVM", tt.description)

			// Verify we can create the root command
			rootCmd := CreateRootCommand(component)
			require.NotNil(t, rootCmd, "Should be able to create root command for %s", tt.description)
		})
	}
}

func TestStep1_1_VersionRouting(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test version routing as specified in Step 1.1
	versionTests := []struct {
		component   string
		args        []string
		description string
	}{
		{
			component:   ComponentPVI,
			args:        []string{"./pvm", "pvi", "version"},
			description: "Version: ./pvm pvi version (should show PVI version)",
		},
		{
			component:   ComponentPVX,
			args:        []string{"./pvm", "pvx", "version"},
			description: "Version: ./pvm pvx version (should show PVX version)",
		},
		{
			component:   ComponentPSC,
			args:        []string{"./pvm", "psc", "version"},
			description: "Version: ./pvm psc version (should show PSC version)",
		},
	}

	for _, tt := range versionTests {
		t.Run("Version"+tt.component, func(t *testing.T) {
			os.Args = tt.args

			// For version subcommands, PVM should be the detected component
			// as it acts as the router
			component := DetectComponent()
			assert.Equal(t, ComponentPVM, component, "%s should route through PVM", tt.description)

			// Verify we can create the root command
			rootCmd := CreateRootCommand(component)
			require.NotNil(t, rootCmd, "Should be able to create root command for %s", tt.description)
		})
	}
}

func TestStep1_1_BackwardCompatibility(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test that backward compatibility is maintained as specified in Step 1.1
	compatibilityTests := []struct {
		name         string
		args         []string
		expectedComp string
		description  string
	}{
		{
			name:         "PVISymlink",
			args:         []string{"pvi", "install", "Module::Name"},
			expectedComp: ComponentPVI,
			description:  "Backward compatibility: pvi symlink should work",
		},
		{
			name:         "PVXSymlink",
			args:         []string{"pvx", "script.pl"},
			expectedComp: ComponentPVX,
			description:  "Backward compatibility: pvx symlink should work",
		},
		{
			name:         "PSCSymlink",
			args:         []string{"psc", "check", "file.pl"},
			expectedComp: ComponentPSC,
			description:  "Backward compatibility: psc symlink should work",
		},
		{
			name:         "PVMDirect",
			args:         []string{"pvm", "install", "5.38.0"},
			expectedComp: ComponentPVM,
			description:  "Direct execution: pvm should work",
		},
	}

	for _, tc := range compatibilityTests {
		t.Run(tc.name, func(t *testing.T) {
			os.Args = tc.args

			component := DetectComponent()
			assert.Equal(t, tc.expectedComp, component, "Component detection failed for %s", tc.description)

			// Verify component description exists
			desc := GetComponentDescription(component)
			assert.NotEqual(t, "Unknown component", desc, "Component %s should have a description", component)

			// Verify we can create a root command
			rootCmd := CreateRootCommand(component)
			require.NotNil(t, rootCmd, "Should be able to create root command for %s", component)
		})
	}
}
