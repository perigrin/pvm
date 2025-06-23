// ABOUTME: Environment setup for PVM end-to-end tests
// ABOUTME: Provides isolated test environment with custom directories and PVM binary

package helpers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/psc"
	"tamarou.com/pvm/internal/pvi"
	"tamarou.com/pvm/internal/pvm"
	"tamarou.com/pvm/internal/pvx"
)

// Flag to preserve test environments for debugging
var preserveEnv = false

// TestEnv represents an isolated test environment
type TestEnv struct {
	// Root directory for all test files
	RootDir string

	// HOME directory for the test
	HomeDir string

	// XDG directories
	ConfigHome string
	DataHome   string
	CacheHome  string
	StateHome  string

	// PVM-specific directories
	PVMConfigDir string
	PVMDataDir   string
	PVMCacheDir  string
	PVMStateDir  string
	PVMBinDir    string
	PVMShimsDir  string

	// Path to the PVM binary
	PVMBinary string

	// Original environment variables
	OriginalEnv map[string]string

	// Testing object
	T *testing.T

	// Standard output capture
	Stdout bytes.Buffer
	Stderr bytes.Buffer
}

// NewTestEnv creates a new isolated test environment
func NewTestEnv(t *testing.T) *TestEnv {
	// Create root directory for test with test name
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	rootDir, err := os.MkdirTemp("", "pvm-e2e-test-"+testName+"-")
	if err != nil {
		t.Fatalf("Failed to create test root directory: %v", err)
	}

	// Create environment structure
	env := &TestEnv{
		RootDir:      rootDir,
		HomeDir:      filepath.Join(rootDir, "home"),
		ConfigHome:   filepath.Join(rootDir, "config"),
		DataHome:     filepath.Join(rootDir, "data"),
		CacheHome:    filepath.Join(rootDir, "cache"),
		StateHome:    filepath.Join(rootDir, "state"),
		PVMConfigDir: filepath.Join(rootDir, "config", "pvm"),
		PVMDataDir:   filepath.Join(rootDir, "data", "pvm"),
		PVMCacheDir:  filepath.Join(rootDir, "cache", "pvm"),
		PVMStateDir:  filepath.Join(rootDir, "state", "pvm"),
		PVMBinDir:    filepath.Join(rootDir, "bin"),
		PVMShimsDir:  filepath.Join(rootDir, "data", "pvm", "shims"),
		T:            t,
	}

	// Create all directories
	dirs := []string{
		env.HomeDir,
		env.ConfigHome,
		env.DataHome,
		env.CacheHome,
		env.StateHome,
		env.PVMConfigDir,
		env.PVMDataDir,
		env.PVMCacheDir,
		env.PVMStateDir,
		env.PVMBinDir,
		env.PVMShimsDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Build PVM binary
	env.PVMBinary = filepath.Join(env.PVMBinDir, "pvm")
	if err := env.buildPVM(); err != nil {
		t.Fatalf("Failed to build PVM binary: %v", err)
	}

	// Save original environment
	env.saveEnvironment()

	// Set up test environment
	env.setEnvironment()

	// Import system Perl to ensure it's available for version resolution
	// This ensures tests have a working Perl environment
	env.importSystemPerl()

	return env
}

// buildPVM builds the PVM binary in the test environment
func (e *TestEnv) buildPVM() error {
	// Get project root directory
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// First run make to ensure tree-sitter parser is generated
	makeCmd := exec.Command("make", "pvm")
	makeCmd.Dir = projectRoot
	makeCmd.Stdout = &e.Stdout
	makeCmd.Stderr = &e.Stderr

	if err := makeCmd.Run(); err != nil {
		return fmt.Errorf("failed to build PVM binary with make: %w\nOutput: %s\nError: %s",
			err, e.Stdout.String(), e.Stderr.String())
	}

	// Set the binary path to the built location
	e.PVMBinary = filepath.Join(projectRoot, "build", "pvm")

	return nil
}

// saveEnvironment saves the original environment variables
func (e *TestEnv) saveEnvironment() {
	e.OriginalEnv = make(map[string]string)

	// Save important environment variables
	vars := []string{
		"HOME",
		"XDG_CONFIG_HOME",
		"XDG_DATA_HOME",
		"XDG_CACHE_HOME",
		"XDG_STATE_HOME",
		"PATH",
		"PVM_HOME",
		"PERL_VERSION",
		"DYLD_LIBRARY_PATH",
		"LD_LIBRARY_PATH",
	}

	for _, v := range vars {
		e.OriginalEnv[v] = os.Getenv(v)
	}
}

// setEnvironment sets up the test environment variables
func (e *TestEnv) setEnvironment() {
	// Set environment variables
	_ = os.Setenv("HOME", e.HomeDir)
	_ = os.Setenv("XDG_CONFIG_HOME", e.ConfigHome)
	_ = os.Setenv("XDG_DATA_HOME", e.DataHome)
	_ = os.Setenv("XDG_CACHE_HOME", e.CacheHome)
	_ = os.Setenv("XDG_STATE_HOME", e.StateHome)

	// Preserve plenv location if it exists in the original HOME
	if originalHome := e.OriginalEnv["HOME"]; originalHome != "" {
		plenvRoot := filepath.Join(originalHome, ".plenv")
		if _, err := os.Stat(plenvRoot); err == nil {
			_ = os.Setenv("PLENV_ROOT", plenvRoot)
		}
	}

	// Add PVM bin and shims directories to PATH
	path := fmt.Sprintf("%s:%s:%s", e.PVMBinDir, e.PVMShimsDir, os.Getenv("PATH"))
	_ = os.Setenv("PATH", path)

	// Set PVM_HOME
	_ = os.Setenv("PVM_HOME", e.PVMDataDir)

	// Set up library paths for tree-sitter
	projectRoot, _ := findProjectRoot()
	if projectRoot != "" {
		libPaths := []string{
			filepath.Join(projectRoot, "lib"),
			filepath.Join(projectRoot, "internal", "parser"),
			filepath.Join(projectRoot, "vendor", "tree-sitter-perl"),
		}

		// Set DYLD_LIBRARY_PATH for macOS
		currentDyldPath := os.Getenv("DYLD_LIBRARY_PATH")
		allPaths := make([]string, len(libPaths))
		copy(allPaths, libPaths)
		allPaths = append(allPaths, currentDyldPath)
		_ = os.Setenv("DYLD_LIBRARY_PATH", strings.Join(allPaths, ":"))

		// Set LD_LIBRARY_PATH for Linux
		currentLdPath := os.Getenv("LD_LIBRARY_PATH")
		allLdPaths := make([]string, len(libPaths))
		copy(allLdPaths, libPaths)
		allLdPaths = append(allLdPaths, currentLdPath)
		_ = os.Setenv("LD_LIBRARY_PATH", strings.Join(allLdPaths, ":"))
	}

	// Unset PERL_VERSION to start clean
	_ = os.Unsetenv("PERL_VERSION")

	// Ensure plenv uses system version for consistent test environment
	_ = os.Setenv("PLENV_VERSION", "system")
}

// Cleanup removes the test environment
func (e *TestEnv) Cleanup() {
	// Don't clean up if preserve flag is set
	if preserveEnv {
		e.T.Logf("Preserving test environment at: %s", e.RootDir)
		return
	}

	// Restore original environment
	for k, v := range e.OriginalEnv {
		if v == "" {
			_ = os.Unsetenv(k)
		} else {
			_ = os.Setenv(k, v)
		}
	}

	// Remove test directory
	_ = os.RemoveAll(e.RootDir)
}

// RunPVM runs the PVM binary with the given arguments
func (e *TestEnv) RunPVM(args ...string) (string, string, error) {
	cmd := exec.Command(e.PVMBinary, args...)

	// Reset output buffers
	e.Stdout.Reset()
	e.Stderr.Reset()

	cmd.Stdout = &e.Stdout
	cmd.Stderr = &e.Stderr

	// Inherit environment to ensure system Perl detection works
	cmd.Env = os.Environ()

	err := cmd.Run()

	return e.Stdout.String(), e.Stderr.String(), err
}

// RunCommand runs a system command with the test environment
func (e *TestEnv) RunCommand(name string, args ...string) (string, string, error) {
	cmd := exec.Command(name, args...)

	// Reset output buffers
	e.Stdout.Reset()
	e.Stderr.Reset()

	cmd.Stdout = &e.Stdout
	cmd.Stderr = &e.Stderr

	// Set PATH to include our test directories
	cmd.Env = os.Environ()

	err := cmd.Run()

	return e.Stdout.String(), e.Stderr.String(), err
}

// CreateFile creates a file in the test environment with the given content
func (e *TestEnv) CreateFile(path string, content string) error {
	// If path is not absolute, make it relative to root dir
	if !filepath.IsAbs(path) {
		path = filepath.Join(e.RootDir, path)
	}

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// ReadFile reads a file in the test environment
func (e *TestEnv) ReadFile(path string) (string, error) {
	// If path is not absolute, make it relative to root dir
	if !filepath.IsAbs(path) {
		path = filepath.Join(e.RootDir, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// FileExists checks if a file exists in the test environment
func (e *TestEnv) FileExists(path string) bool {
	// If path is not absolute, make it relative to root dir
	if !filepath.IsAbs(path) {
		path = filepath.Join(e.RootDir, path)
	}

	_, err := os.Stat(path)
	return err == nil
}

// findProjectRoot finds the root directory of the project
func findProjectRoot() (string, error) {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Look for go.mod file to identify project root
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		// Go up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding go.mod
			return "", fmt.Errorf("could not find project root (no go.mod file found)")
		}
		dir = parent
	}
}

// SetPreserveEnv sets the flag to preserve test environments
func SetPreserveEnv(preserve bool) {
	preserveEnv = preserve
}

// CopyFile copies a file from source to destination
func (e *TestEnv) CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }()

	// If dst is not absolute, make it relative to root dir
	if !filepath.IsAbs(dst) {
		dst = filepath.Join(e.RootDir, dst)
	}

	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

// RunPVMCommand runs a PVM command using the internal CLI framework instead of the binary
func (e *TestEnv) RunPVMCommand(args ...string) (string, string, error) {
	// Reset global state to prevent test interference
	cli.ResetGlobalState()

	// Set up component registry (use global registry)
	cli.GlobalRegistry.Register(cli.ComponentPVM, pvm.NewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPSC, psc.NewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPVI, pvi.NewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPVX, pvx.NewCommand)

	// Create root command
	rootCmd := cli.CreateRootCommand(cli.ComponentPVM)

	// Capture output
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	// Set arguments
	rootCmd.SetArgs(args)

	// Execute command
	err := rootCmd.Execute()

	return stdout.String(), stderr.String(), err
}

// importSystemPerl imports system Perl during test environment setup
func (e *TestEnv) importSystemPerl() {
	// Reset global state to prevent test interference
	cli.ResetGlobalState()

	// Register components in global registry
	cli.GlobalRegistry.Register(cli.ComponentPVM, pvm.NewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPSC, psc.NewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPVI, pvi.NewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPVX, pvx.NewCommand)

	// Run import-system command
	rootCmd := cli.CreateRootCommand(cli.ComponentPVM)
	rootCmd.SetArgs([]string{"import-system"})

	// Capture output but ignore errors for test setup
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	_ = rootCmd.Execute()
}

// RunPVXWithCleanIsolation runs PVX with clean isolation for reliable test environments
// This provides a consistent, clean module environment that's perfect for e2e tests
func (e *TestEnv) RunPVXWithCleanIsolation(args ...string) (string, string, error) {
	// Prepend clean isolation flags to the arguments
	pvxArgs := []string{"pvx", "--isolation", "clean", "--verbose"}
	pvxArgs = append(pvxArgs, args...)
	return e.RunPVM(pvxArgs...)
}

// RunPVXScriptWithCleanIsolation runs a Perl script using PVX with clean isolation
// This ensures the script runs in a clean module environment without system interference
func (e *TestEnv) RunPVXScriptWithCleanIsolation(scriptPath string, scriptArgs ...string) (string, string, error) {
	args := []string{scriptPath}
	args = append(args, scriptArgs...)
	return e.RunPVXWithCleanIsolation(args...)
}

// RunPVXToolWithCleanIsolation runs a Perl tool using PVX with clean isolation
// Uses the -- separator to properly separate PVX flags from tool arguments
func (e *TestEnv) RunPVXToolWithCleanIsolation(toolName string, toolArgs ...string) (string, string, error) {
	args := []string{"--", toolName}
	args = append(args, toolArgs...)
	return e.RunPVXWithCleanIsolation(args...)
}

// RunPVXInlineWithCleanIsolation runs inline Perl code using PVX with clean isolation
func (e *TestEnv) RunPVXInlineWithCleanIsolation(code string) (string, string, error) {
	return e.RunPVXWithCleanIsolation("-e", code)
}
