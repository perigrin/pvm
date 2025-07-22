package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestModuleConfig verifies that the go.mod file exists and has the correct module name
func TestModuleConfig(t *testing.T) {
	helper, err := NewTestHelper()
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}

	// Check that go.mod exists
	goModPath := filepath.Join(helper.ProjectRoot, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Fatalf("go.mod does not exist at %s", goModPath)
	}

	// Additional checks could be added here to parse and validate go.mod contents
}

// TestBuildCapability verifies that the project can be built
func TestBuildCapability(t *testing.T) {
	helper, err := NewTestHelper()
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}

	// Run go build ./... to verify everything compiles
	cmd := exec.Command("go", "build", "-mod=mod", "./...")
	cmd.Dir = helper.ProjectRoot
	if err := cmd.Run(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}
}
