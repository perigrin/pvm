// ABOUTME: Main setup for PVM end-to-end tests
// ABOUTME: Provides common test setup and flags for all E2E tests

package e2e

import (
	"flag"
	"fmt"
	"os"
	"runtime"
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

	// E2E tests depend on Unix shell integration and path conventions.
	// Skip the entire suite on Windows until platform-specific support is added.
	if runtime.GOOS == "windows" {
		fmt.Println("SKIP: E2E tests not yet supported on Windows")
		os.Exit(0)
	}

	// Set up global test configuration
	helpers.SetPreserveEnv(*preserveFlag)

	// Run tests
	os.Exit(m.Run())
}
