// ABOUTME: Main setup for PVM end-to-end tests
// ABOUTME: Provides common test setup and flags for all E2E tests

package e2e

import (
	"flag"
	"os"
	"testing"

	"tamarou.com/pvm/test/e2e/helpers"
)

// Command-line flags for tests
var (
	preserveFlag = flag.Bool("preserve", false, "Preserve test environments for debugging")
)

func TestMain(m *testing.M) {
	// Parse flags
	flag.Parse()

	// Set up global test configuration
	helpers.SetPreserveEnv(*preserveFlag)

	// Run tests
	os.Exit(m.Run())
}
