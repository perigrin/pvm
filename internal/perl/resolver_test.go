// ABOUTME: Tests for version resolution algorithm
// ABOUTME: Ensures the version resolution precedence works correctly

package perl

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/xdg"
)

// TestResolveForkIdentifierFromPerlVersionFile verifies that a .perl-version file
// containing a fork identifier (e.g. "mycompany/myfork-5.40.2") resolves to the
// install path recorded in the registry for that fork.
func TestResolveForkIdentifierFromPerlVersionFile(t *testing.T) {
	env := setupResolverTest(t)
	defer env.cleanup_()

	const forkDisplayName = "mycompany/myfork-5.40.2"
	const forkInstallPath = "/mock/pvm/mycompany/myfork-5.40.2"

	// Seed the registry with the fork entry
	origLoadRegistry := LoadRegistry
	defer func() { LoadRegistry = origLoadRegistry }()
	LoadRegistry = func() (*VersionRegistry, error) {
		return &VersionRegistry{
			Versions: map[string]VersionInfo{
				"uuid-fork-1": {
					Version:     "myfork-5.40.2",
					InstallPath: forkInstallPath,
					Source:      "pvm",
					Remote:      "mycompany",
					ForkName:    "myfork",
					BaseVersion: "5.40.2",
				},
			},
		}, nil
	}

	// Mock FindDotPerlVersionFiles and ReadPerlVersionFile to return the fork identifier
	origFind := FindDotPerlVersionFiles
	defer func() { FindDotPerlVersionFiles = origFind }()
	versionFilePath := filepath.Join(env.versionFileDir, ".perl-version")
	FindDotPerlVersionFiles = func(startDir string) ([]string, error) {
		return []string{versionFilePath}, nil
	}

	origRead := ReadPerlVersionFile
	defer func() { ReadPerlVersionFile = origRead }()
	ReadPerlVersionFile = func(dir string) (string, error) {
		return forkDisplayName, nil
	}

	options := &ResolutionOptions{
		ProjectDir:          env.versionFileDir,
		AvailableVersions:   []string{"5.40.2"},
		SkipVersionResolved: true,
	}

	resolved, err := ResolveVersion(options)
	if err != nil {
		t.Fatalf("Failed to resolve fork identifier from .perl-version file: %v", err)
	}

	if resolved.Version != forkDisplayName {
		t.Errorf("Expected version %q, got %q", forkDisplayName, resolved.Version)
	}

	if resolved.Source != ProjectVersionFile {
		t.Errorf("Expected source ProjectVersionFile, got %s", resolved.Source)
	}

	perlBin := "perl"
	if runtime.GOOS == "windows" {
		perlBin = "perl.exe"
	}
	expectedPath := filepath.Join(forkInstallPath, "bin", perlBin)
	if resolved.Path != expectedPath {
		t.Errorf("Expected path %q, got %q", expectedPath, resolved.Path)
	}
}

// TestResolveForkIdentifierFromEnvVar verifies that PVM_PERL_VERSION containing a
// fork identifier resolves to the correct install path from the registry.
func TestResolveForkIdentifierFromEnvVar(t *testing.T) {
	env := setupResolverTest(t)
	defer env.cleanup_()

	const forkDisplayName = "mycompany/myfork-5.40.2"
	const forkInstallPath = "/mock/pvm/mycompany/myfork-5.40.2"

	// Seed the registry with the fork entry
	origLoadRegistry := LoadRegistry
	defer func() { LoadRegistry = origLoadRegistry }()
	LoadRegistry = func() (*VersionRegistry, error) {
		return &VersionRegistry{
			Versions: map[string]VersionInfo{
				"uuid-fork-2": {
					Version:     "myfork-5.40.2",
					InstallPath: forkInstallPath,
					Source:      "pvm",
					Remote:      "mycompany",
					ForkName:    "myfork",
					BaseVersion: "5.40.2",
				},
			},
		}, nil
	}

	_ = os.Setenv("PVM_PERL_VERSION", forkDisplayName)
	defer os.Unsetenv("PVM_PERL_VERSION")

	options := &ResolutionOptions{
		AvailableVersions:   []string{"5.40.2"},
		SkipLocal:           true,
		SkipVersionResolved: true,
	}

	resolved, err := ResolveVersion(options)
	if err != nil {
		t.Fatalf("Failed to resolve fork identifier from PVM_PERL_VERSION: %v", err)
	}

	if resolved.Version != forkDisplayName {
		t.Errorf("Expected version %q, got %q", forkDisplayName, resolved.Version)
	}

	if resolved.Source != EnvironmentVariable {
		t.Errorf("Expected source EnvironmentVariable, got %s", resolved.Source)
	}

	perlBin := "perl"
	if runtime.GOOS == "windows" {
		perlBin = "perl.exe"
	}
	expectedPath := filepath.Join(forkInstallPath, "bin", perlBin)
	if resolved.Path != expectedPath {
		t.Errorf("Expected path %q, got %q", expectedPath, resolved.Path)
	}
}

// TestResolveStockVersionBackwardsCompat verifies that version strings without "/"
// continue to use the normal (stock) resolution path unchanged.
func TestResolveStockVersionBackwardsCompat(t *testing.T) {
	env := setupResolverTest(t)
	defer env.cleanup_()

	options := &ResolutionOptions{
		ExplicitVersion:     "5.38.0",
		AvailableVersions:   []string{"5.38.0", "5.36.0"},
		SkipVersionResolved: true,
	}

	resolved, err := ResolveVersion(options)
	if err != nil {
		t.Fatalf("Stock version resolution failed: %v", err)
	}

	if resolved.Version != "5.38.0" {
		t.Errorf("Expected 5.38.0, got %s", resolved.Version)
	}

	if resolved.Source != ExplicitVersion {
		t.Errorf("Expected ExplicitVersion source, got %s", resolved.Source)
	}
}

// TestResolveForkIdentifierNotInRegistry verifies that requesting a fork not present
// in the registry returns an appropriate error.
func TestResolveForkIdentifierNotInRegistry(t *testing.T) {
	env := setupResolverTest(t)
	defer env.cleanup_()

	// Registry with no fork entries
	origLoadRegistry := LoadRegistry
	defer func() { LoadRegistry = origLoadRegistry }()
	LoadRegistry = func() (*VersionRegistry, error) {
		return &VersionRegistry{
			Versions: map[string]VersionInfo{},
		}, nil
	}

	_ = os.Setenv("PVM_PERL_VERSION", "mycompany/myfork-5.40.2")
	defer os.Unsetenv("PVM_PERL_VERSION")

	options := &ResolutionOptions{
		AvailableVersions:   []string{"5.40.2"},
		SkipLocal:           true,
		SkipVersionResolved: true,
		// SkipSystemPerl prevents the resolver's step-7 fallback from
		// finding a real /usr/bin/perl on the test host. Without this,
		// the env-step's "fork not found" error is swallowed and the
		// system fallback succeeds — masking what this test is meant
		// to assert (that an unsatisfied fork bubbles up). The
		// fall-through behavior in the env step is a separate issue
		// (see #456); this test is scoped to fork resolution.
		SkipSystemPerl: true,
	}

	_, err := ResolveVersion(options)
	if err == nil {
		t.Fatal("Expected error for fork not in registry, got nil")
	}
}

// Setup for testing
type resolverTestEnv struct {
	// Original environment variables
	origEnv map[string]string

	// Temporary directories
	tempDir        string
	projectDir     string
	pvmProjectDir  string
	versionFileDir string

	// Mock system Perl
	mockSystemPerl *SystemPerl

	// Cleanup functions
	cleanup []func()
}

// Setup test environment
func setupResolverTest(t *testing.T) *resolverTestEnv {
	env := &resolverTestEnv{
		origEnv: make(map[string]string),
		cleanup: []func(){},
	}

	// Save original environment variables
	for _, name := range []string{"PVM_PERL_VERSION", "PLENV_VERSION", "PERLBREW_PERL", "PVM_SKIP_NETWORK_CALLS"} {
		env.origEnv[name] = os.Getenv(name)
	}

	// Log environment state before cleanup (CI debugging)
	if os.Getenv("CI") != "" {
		t.Logf("CI DEBUG - Test setup environment before cleanup:")
		for _, name := range []string{"PVM_PERL_VERSION", "PLENV_VERSION", "PERLBREW_PERL", "PVM_SKIP_NETWORK_CALLS", "PATH"} {
			value := os.Getenv(name)
			if value != "" {
				t.Logf("  %s=%s", name, value)
			} else {
				t.Logf("  %s=<unset>", name)
			}
		}
	}

	// Clear environment variables to ensure test isolation
	for _, name := range []string{"PVM_PERL_VERSION", "PLENV_VERSION", "PERLBREW_PERL"} {
		_ = os.Unsetenv(name)
	}

	// Set test mode environment variables
	_ = os.Setenv("PVM_SKIP_NETWORK_CALLS", "1")

	// Log environment state after cleanup (CI debugging)
	if os.Getenv("CI") != "" {
		t.Logf("CI DEBUG - Test setup environment after cleanup:")
		for _, name := range []string{"PVM_PERL_VERSION", "PLENV_VERSION", "PERLBREW_PERL", "PVM_SKIP_NETWORK_CALLS"} {
			value := os.Getenv(name)
			if value != "" {
				t.Logf("  %s=%s", name, value)
			} else {
				t.Logf("  %s=<unset>", name)
			}
		}
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pvm-resolver-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	env.tempDir = tempDir
	env.cleanup = append(env.cleanup, func() { _ = os.RemoveAll(tempDir) })

	// Create project directories
	env.projectDir = filepath.Join(tempDir, "project")
	if err := os.MkdirAll(env.projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create .pvm directory
	env.pvmProjectDir = filepath.Join(env.projectDir, ".pvm")
	if err := os.MkdirAll(env.pvmProjectDir, 0755); err != nil {
		t.Fatalf("Failed to create .pvm directory: %v", err)
	}

	// Create directory with .perl-version file
	env.versionFileDir = filepath.Join(env.tempDir, "version-file")
	if err := os.MkdirAll(env.versionFileDir, 0755); err != nil {
		t.Fatalf("Failed to create version file directory: %v", err)
	}

	// Set up mock system Perl function
	originalDetectSystemPerl := DetectSystemPerl
	env.cleanup = append(env.cleanup, func() { DetectSystemPerl = originalDetectSystemPerl })

	// Use actual system perl for more robust testing
	// Ensure clean environment for system perl detection to avoid test interference
	origPlenvVersion := os.Getenv("PLENV_VERSION")
	origPvmPerlVersion := os.Getenv("PVM_PERL_VERSION")
	origPerlbrewPerl := os.Getenv("PERLBREW_PERL")
	_ = os.Unsetenv("PLENV_VERSION")
	_ = os.Unsetenv("PVM_PERL_VERSION")
	_ = os.Unsetenv("PERLBREW_PERL")

	// Log system Perl detection attempt (CI debugging)
	if os.Getenv("CI") != "" {
		t.Logf("CI DEBUG - Attempting system Perl detection...")
	}

	actualSystemPerl, err := originalDetectSystemPerl()

	// Log system Perl detection result (CI debugging)
	if os.Getenv("CI") != "" {
		if err != nil {
			t.Logf("CI DEBUG - System Perl detection failed: %v", err)
		} else {
			t.Logf("CI DEBUG - System Perl detected: Path=%s, Version=%s, Architecture=%s",
				actualSystemPerl.Path, actualSystemPerl.Version, actualSystemPerl.Architecture)
		}
	}

	// Restore environment variables
	if origPlenvVersion != "" {
		_ = os.Setenv("PLENV_VERSION", origPlenvVersion)
	}
	if origPvmPerlVersion != "" {
		_ = os.Setenv("PVM_PERL_VERSION", origPvmPerlVersion)
	}
	if origPerlbrewPerl != "" {
		_ = os.Setenv("PERLBREW_PERL", origPerlbrewPerl)
	}

	if err != nil {
		// If we can't detect system perl, use a sensible default
		actualSystemPerl = &SystemPerl{
			Path:         "/usr/bin/perl",
			Version:      "5.38.2",
			FullVersion:  "5.38.2",
			Architecture: "x86_64",
			IsPrimary:    true,
		}
	}

	env.mockSystemPerl = actualSystemPerl
	DetectSystemPerl = func() (*SystemPerl, error) {
		return env.mockSystemPerl, nil
	}

	return env
}

// Cleanup test environment
func (env *resolverTestEnv) cleanup_() {
	// Log cleanup start (CI debugging)
	if os.Getenv("CI") != "" {
		// Get a test instance for logging - we'll use testing.TB interface if available
		// For now, just print directly in CI mode
		fmt.Printf("CI DEBUG - Starting test cleanup, restoring environment variables\n")
		for name, value := range env.origEnv {
			if value == "" {
				fmt.Printf("  Unsetting %s\n", name)
			} else {
				fmt.Printf("  Setting %s=%s\n", name, value)
			}
		}
	}

	// Restore environment variables
	for name, value := range env.origEnv {
		if value == "" {
			_ = os.Unsetenv(name)
		} else {
			_ = os.Setenv(name, value)
		}
	}

	// Run cleanup functions in reverse order
	for i := len(env.cleanup) - 1; i >= 0; i-- {
		env.cleanup[i]()
	}
}

// Create a project configuration file
func (env *resolverTestEnv) createProjectConfig(t *testing.T, defaultPerl string, aliases map[string]string) {
	configFile := filepath.Join(env.pvmProjectDir, "pvm.toml")

	// Create simple TOML content
	content := "[pvm]\n"
	content += "default_perl = \"" + defaultPerl + "\"\n"

	if len(aliases) > 0 {
		content += "\n[pvm.version_aliases]\n"
		for alias, value := range aliases {
			content += alias + " = \"" + value + "\"\n"
		}
	}

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create project config file: %v", err)
	}
}

// Create a .perl-version file
func (env *resolverTestEnv) createPerlVersionFile(t *testing.T, version string) {
	versionFile := filepath.Join(env.versionFileDir, ".perl-version")
	if err := os.WriteFile(versionFile, []byte(version+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create .perl-version file: %v", err)
	}
}

func (env *resolverTestEnv) createUserConfig(t *testing.T, defaultPerl string, aliases map[string]string) {
	// Set XDG_CONFIG_HOME to our temp directory to ensure consistent behavior
	testConfigDir := filepath.Join(env.tempDir, "config")
	if err := os.MkdirAll(testConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}

	// Set environment variable to override XDG config location
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", testConfigDir)
	env.cleanup = append(env.cleanup, func() {
		if originalXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		}
	})

	// Get user config directory (should now use our test directory)
	dirs, err := xdg.GetDirs()
	if err != nil {
		t.Fatalf("Failed to get XDG dirs: %v", err)
	}

	// Create the user config directory if it doesn't exist
	userConfigDir := dirs.ConfigDir
	if err := os.MkdirAll(userConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create user config directory: %v", err)
	}

	// Create user config file
	userConfigPath := dirs.GetConfigFilePath()

	// Build config content
	content := fmt.Sprintf("default_perl = \"%s\"\n", defaultPerl)
	if len(aliases) > 0 {
		content += "version_aliases = { "
		first := true
		for alias, version := range aliases {
			if !first {
				content += ", "
			}
			content += fmt.Sprintf("%s = \"%s\"", alias, version)
			first = false
		}
		content += " }\n"
	}

	if err := os.WriteFile(userConfigPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create user config file: %v", err)
	}

	t.Logf("Created user config file at: %s", userConfigPath)
}

// Test explicit version resolution
func TestResolveExplicitVersion(t *testing.T) {
	env := setupResolverTest(t)
	defer env.cleanup_()

	// Test with available version
	availableVersions := []string{"5.38.0", "5.36.0", "5.34.1"}

	cfg := &config.Config{
		PVM: &config.PVMConfig{
			DefaultPerl: "5.36.0",
			VersionAliases: map[string]string{
				"latest": "5.38.0",
				"stable": "5.36.0",
			},
		},
	}

	options := &ResolutionOptions{
		ExplicitVersion:     "5.38.0",
		AvailableVersions:   availableVersions,
		Config:              cfg,
		SkipVersionResolved: true,
	}

	resolved, err := ResolveVersion(options)
	if err != nil {
		t.Fatalf("Failed to resolve explicit version: %v", err)
	}

	if resolved.Version != "5.38.0" {
		t.Errorf("Expected version 5.38.0, got %s", resolved.Version)
	}

	if resolved.Source != ExplicitVersion {
		t.Errorf("Expected source ExplicitVersion, got %s", resolved.Source)
	}

	// Test with alias
	options.ExplicitVersion = "@latest"
	resolved, err = ResolveVersion(options)
	if err != nil {
		t.Fatalf("Failed to resolve explicit version with alias: %v", err)
	}

	if resolved.Version != "5.38.0" {
		t.Errorf("Expected version 5.38.0 for @latest, got %s", resolved.Version)
	}

	// Simple direct test of the resolveExplicitVersion function
	version := "5.99.0" // A very unlikely version
	availVersions := []string{"5.38.0", "5.36.0", "5.34.1"}
	testConfig := &config.Config{
		PVM: &config.PVMConfig{
			VersionAliases: map[string]string{
				"latest": "5.38.0",
			},
		},
	}

	// Call the function directly to see if it's working properly
	resolved2, err2 := resolveExplicitVersion(version, availVersions, testConfig)
	if err2 == nil {
		t.Errorf("Expected resolveExplicitVersion to error for unavailable version %s, but got resolved: %+v",
			version, resolved2)
	}
}

// Test .perl-version file resolution
func TestResolvePerlVersionFile(t *testing.T) {
	env := setupResolverTest(t)
	defer env.cleanup_()

	availableVersions := []string{"5.38.0", "5.36.0", "5.34.1"}

	// Create .perl-version file
	env.createPerlVersionFile(t, "5.36.0")

	// Mock FindDotPerlVersionFiles to return our test file
	originalFindDotPerlVersionFiles := FindDotPerlVersionFiles
	defer func() { FindDotPerlVersionFiles = originalFindDotPerlVersionFiles }()

	versionFilePath := filepath.Join(env.versionFileDir, ".perl-version")
	FindDotPerlVersionFiles = func(startDir string) ([]string, error) {
		return []string{versionFilePath}, nil
	}

	// Mock ReadPerlVersionFile to return our test version
	originalReadPerlVersionFile := ReadPerlVersionFile
	defer func() { ReadPerlVersionFile = originalReadPerlVersionFile }()

	ReadPerlVersionFile = func(dir string) (string, error) {
		return "5.36.0", nil
	}

	options := &ResolutionOptions{
		ProjectDir:          env.versionFileDir,
		AvailableVersions:   availableVersions,
		SkipVersionResolved: true,
	}

	resolved, err := ResolveVersion(options)
	if err != nil {
		t.Fatalf("Failed to resolve from .perl-version file: %v", err)
	}

	if resolved.Version != "5.36.0" {
		t.Errorf("Expected version 5.36.0, got %s", resolved.Version)
	}

	if resolved.Source != ProjectVersionFile {
		t.Errorf("Expected source ProjectVersionFile, got %s", resolved.Source)
	}
}

// Test project configuration resolution
func TestResolveProjectConfig(t *testing.T) {
	env := setupResolverTest(t)
	defer env.cleanup_()

	availableVersions := []string{"5.38.0", "5.36.0", "5.34.1"}

	// Create project config
	env.createProjectConfig(t, "5.34.1", map[string]string{
		"old":   "5.16.3",
		"local": "5.34.1",
	})

	// Mock GetProjectRoot to return our test directory
	originalGetProjectRoot := config.GetProjectRoot
	defer func() { config.GetProjectRoot = originalGetProjectRoot }()

	config.GetProjectRoot = func() string {
		return env.projectDir
	}

	// Set options to skip .perl-version file check since we're testing project config
	options := &ResolutionOptions{
		ProjectDir:          env.projectDir,
		AvailableVersions:   availableVersions,
		SkipVersionResolved: true,
	}

	// Mock FindDotPerlVersionFiles to return no files, so we test project config
	originalFindDotPerlVersionFiles := FindDotPerlVersionFiles
	defer func() { FindDotPerlVersionFiles = originalFindDotPerlVersionFiles }()

	FindDotPerlVersionFiles = func(startDir string) ([]string, error) {
		return []string{}, nil
	}

	resolved, err := ResolveVersion(options)
	if err != nil {
		t.Fatalf("Failed to resolve from project config: %v", err)
	}

	if resolved.Version != "5.34.1" {
		t.Errorf("Expected version 5.34.1, got %s", resolved.Version)
	}

	if resolved.Source != ProjectConfig {
		t.Errorf("Expected source ProjectConfig, got %s", resolved.Source)
	}
}

// Test environment variable resolution
func TestResolveEnvironmentVariables(t *testing.T) {
	env := setupResolverTest(t)
	defer env.cleanup_()

	// Ensure environment variables are cleaned up after test
	defer os.Unsetenv("PVM_PERL_VERSION")
	defer os.Unsetenv("PLENV_VERSION")
	defer os.Unsetenv("PERLBREW_PERL")

	availableVersions := []string{"5.38.0", "5.36.0", "5.34.1"}

	options := &ResolutionOptions{
		AvailableVersions:   availableVersions,
		SkipLocal:           true, // Skip project-local checks
		SkipVersionResolved: true,
	}

	// Test PVM_PERL_VERSION (highest precedence)
	_ = os.Setenv("PVM_PERL_VERSION", "5.34.1")
	_ = os.Setenv("PLENV_VERSION", "5.38.0")
	_ = os.Setenv("PERLBREW_PERL", "perl-5.36.0")

	resolved, err := ResolveVersion(options)
	if err != nil {
		t.Fatalf("Failed to resolve from PVM_PERL_VERSION: %v", err)
	}

	if resolved.Version != "5.34.1" {
		t.Errorf("Expected version 5.34.1 from PVM_PERL_VERSION (highest precedence), got %s", resolved.Version)
	}

	if resolved.Source != EnvironmentVariable {
		t.Errorf("Expected source EnvironmentVariable, got %s", resolved.Source)
	}

	// Test PLENV_VERSION (middle precedence)
	_ = os.Unsetenv("PVM_PERL_VERSION")

	resolved, err = ResolveVersion(options)
	if err != nil {
		t.Fatalf("Failed to resolve from PLENV_VERSION: %v", err)
	}

	if resolved.Version != "5.38.0" {
		t.Errorf("Expected version 5.38.0 from PLENV_VERSION, got %s", resolved.Version)
	}

	if resolved.Source != EnvironmentVariable {
		t.Errorf("Expected source EnvironmentVariable, got %s", resolved.Source)
	}

	// Test PERLBREW_PERL (lowest precedence)
	_ = os.Unsetenv("PLENV_VERSION")

	resolved, err = ResolveVersion(options)
	if err != nil {
		t.Fatalf("Failed to resolve from PERLBREW_PERL: %v", err)
	}

	if resolved.Version != "5.36.0" {
		t.Errorf("Expected version 5.36.0 from PERLBREW_PERL, got %s", resolved.Version)
	}
}

// Test user configuration resolution
func TestResolveUserConfig(t *testing.T) {
	env := setupResolverTest(t)
	defer env.cleanup_()

	availableVersions := []string{"5.38.0", "5.36.0", "5.34.1"}

	// Create user config object directly
	aliases := map[string]string{
		"latest": "5.38.0",
		"stable": "5.36.0",
	}

	// Create user config file for the resolveFromUserConfig file existence check
	env.createUserConfig(t, "5.34.1", aliases)

	// Create config object directly to avoid config loading complexities in tests
	cfg := &config.Config{
		PVM: &config.PVMConfig{
			DefaultPerl:    "5.34.1",
			VersionAliases: aliases,
		},
	}

	options := &ResolutionOptions{
		AvailableVersions:   availableVersions,
		Config:              cfg,
		SkipLocal:           true, // Skip project-local checks
		SkipEnvVars:         true, // Skip environment variables
		SkipVersionResolved: true,
	}

	resolved, err := ResolveVersion(options)
	if err != nil {
		t.Fatalf("Failed to resolve from user config: %v", err)
	}

	if resolved.Version != "5.34.1" {
		t.Errorf("Expected version 5.34.1 from user config, got %s", resolved.Version)
	}

	if resolved.Source != UserConfig {
		t.Errorf("Expected source UserConfig, got %s", resolved.Source)
	}

	// Test with user config alias
	env.createUserConfig(t, "@latest", aliases)

	// Update config object to use alias
	cfg.PVM.DefaultPerl = "@latest"
	options.Config = cfg

	resolved, err = ResolveVersion(options)
	if err != nil {
		t.Fatalf("Failed to resolve from user config with alias: %v", err)
	}

	if resolved.Version != "5.38.0" {
		t.Errorf("Expected version 5.38.0 for @latest in user config, got %s", resolved.Version)
	}
}

// Test fallback to system Perl
func TestResolveSystemPerl(t *testing.T) {
	// Log test start (CI debugging)
	if os.Getenv("CI") != "" {
		t.Logf("CI DEBUG - TestResolveSystemPerl starting")
	}

	env := setupResolverTest(t)
	defer env.cleanup_()

	// Log mock system perl info (CI debugging)
	if os.Getenv("CI") != "" {
		t.Logf("CI DEBUG - Mock system perl: Path=%s, Version=%s", env.mockSystemPerl.Path, env.mockSystemPerl.Version)
	}

	// Clear only system perl entries to prevent conflicts, but keep other registry entries
	// This prevents interference from previous tests while preserving PVX functionality
	err := clearSystemPerlFromRegistry()
	if err != nil {
		t.Logf("Warning: Failed to clear system perl from registry: %v", err)
	}

	// Import system perl into registry for resolution to work
	// Since AutoImportSystemPerl skips if already registered, we force import
	err = ImportSystemPerl()
	if err != nil {
		t.Fatalf("Failed to import system perl: %v", err)
	}

	// Log registry state after import (CI debugging)
	if os.Getenv("CI") != "" {
		installedVersions, regErr := GetInstalledVersions()
		if regErr == nil {
			t.Logf("CI DEBUG - Registry entries after import:")
			for _, v := range installedVersions {
				if v.Source == "system" {
					t.Logf("  System entry: Version=%s, InstallPath=%s", v.Version, v.InstallPath)
				}
			}
		}
	}

	// Set options to skip all other resolution methods
	options := &ResolutionOptions{
		SkipLocal:           true, // Skip project-local checks
		SkipEnvVars:         true, // Skip environment variables
		SkipUserConfig:      true, // Skip user configuration
		SkipVersionResolved: true,
	}

	resolved, err := ResolveVersion(options)
	if err != nil {
		t.Fatalf("Failed to resolve from system Perl: %v", err)
	}

	if resolved.Version != env.mockSystemPerl.Version {
		t.Errorf("Expected version %s from system Perl, got %s", env.mockSystemPerl.Version, resolved.Version)
	}

	if resolved.Source != SystemPerlSource {
		t.Errorf("Expected source SystemPerlSource, got %s", resolved.Source)
	}

	if resolved.Path != env.mockSystemPerl.Path {
		t.Errorf("Expected path %s from system Perl, got %s", env.mockSystemPerl.Path, resolved.Path)
	}
}

// Test full resolution precedence
func TestResolutionPrecedence(t *testing.T) {
	// Log test start (CI debugging)
	if os.Getenv("CI") != "" {
		t.Logf("CI DEBUG - TestResolutionPrecedence starting")
	}

	env := setupResolverTest(t)
	defer env.cleanup_()

	availableVersions := []string{"5.38.0", "5.36.0", "5.34.1", "5.32.1", "5.30.3", env.mockSystemPerl.Version}

	// Log available versions (CI debugging)
	if os.Getenv("CI") != "" {
		t.Logf("CI DEBUG - Available versions: %v", availableVersions)
		t.Logf("CI DEBUG - Mock system perl version: %s", env.mockSystemPerl.Version)
	}

	// Create config files and set environment variables to test precedence

	// 1. System Perl (lowest precedence) - use the actual system perl version
	// env.mockSystemPerl.Version is already set to actual system perl version

	// 2. User config has 5.32.1 - create user config file and object
	env.createUserConfig(t, "5.32.1", nil)
	cfg := &config.Config{
		PVM: &config.PVMConfig{
			DefaultPerl: "5.32.1",
		},
	}

	// 3. Environment variables
	_ = os.Setenv("PERLBREW_PERL", "perl-5.34.1")

	// 4. Project config
	env.createProjectConfig(t, "5.36.0", nil)

	// Mock GetProjectRoot to return our test directory
	originalGetProjectRoot := config.GetProjectRoot
	defer func() { config.GetProjectRoot = originalGetProjectRoot }()

	config.GetProjectRoot = func() string {
		return env.projectDir
	}

	// 5. .perl-version file
	env.createPerlVersionFile(t, "5.38.0")

	// Mock FindDotPerlVersionFiles and ReadPerlVersionFile
	originalFindDotPerlVersionFiles := FindDotPerlVersionFiles
	defer func() { FindDotPerlVersionFiles = originalFindDotPerlVersionFiles }()

	versionFilePath := filepath.Join(env.versionFileDir, ".perl-version")
	FindDotPerlVersionFiles = func(startDir string) ([]string, error) {
		return []string{versionFilePath}, nil
	}

	originalReadPerlVersionFile := ReadPerlVersionFile
	defer func() { ReadPerlVersionFile = originalReadPerlVersionFile }()

	ReadPerlVersionFile = func(dir string) (string, error) {
		return "5.38.0", nil
	}

	// Test precedence
	// Reset environment variables for testing precedence
	_ = os.Unsetenv("PLENV_VERSION")
	_ = os.Setenv("PERLBREW_PERL", "perl-5.34.1")

	tests := []struct {
		name     string
		options  ResolutionOptions
		expected struct {
			version string
			source  ResolutionSource
		}
	}{
		{
			name: "Explicit version has highest precedence",
			options: ResolutionOptions{
				ExplicitVersion:     "5.26.3",
				AvailableVersions:   append(availableVersions, "5.26.3"),
				Config:              cfg,
				ProjectDir:          env.projectDir,
				SkipVersionResolved: true,
			},
			expected: struct {
				version string
				source  ResolutionSource
			}{
				version: "5.26.3",
				source:  ExplicitVersion,
			},
		},
		{
			name: "Environment variables have precedence over .perl-version file",
			options: ResolutionOptions{
				AvailableVersions:   availableVersions,
				Config:              cfg,
				ProjectDir:          env.projectDir,
				SkipVersionResolved: true,
			},
			expected: struct {
				version string
				source  ResolutionSource
			}{
				version: "5.34.1", // From PERLBREW_PERL environment variable
				source:  EnvironmentVariable,
			},
		},
		{
			name: ".perl-version file has precedence over project config",
			options: ResolutionOptions{
				AvailableVersions:   availableVersions,
				Config:              cfg,
				ProjectDir:          env.projectDir,
				SkipEnvVars:         true, // Skip environment variables to test .perl-version vs project config
				SkipVersionResolved: true,
			},
			expected: struct {
				version string
				source  ResolutionSource
			}{
				version: "5.38.0", // From .perl-version, which has higher precedence than project config
				source:  ProjectVersionFile,
			},
		},
		{
			name: "Environment variables have precedence over user config",
			options: ResolutionOptions{
				AvailableVersions:   availableVersions,
				Config:              cfg,
				SkipLocal:           true, // Skip project-local checks
				SkipVersionResolved: true,
			},
			expected: struct {
				version string
				source  ResolutionSource
			}{
				// We set PERLBREW_PERL to perl-5.34.1 above
				version: "5.34.1",
				source:  EnvironmentVariable,
			},
		},
		{
			name: "User config has precedence over system Perl",
			options: ResolutionOptions{
				AvailableVersions:   availableVersions,
				Config:              cfg,
				SkipLocal:           true, // Skip project-local checks
				SkipEnvVars:         true, // Skip environment variables
				SkipVersionResolved: true,
			},
			expected: struct {
				version string
				source  ResolutionSource
			}{
				version: "5.32.1",
				source:  UserConfig,
			},
		},
		{
			name: "System Perl is used as last resort",
			options: ResolutionOptions{
				AvailableVersions:   availableVersions,
				SkipLocal:           true, // Skip project-local checks
				SkipEnvVars:         true, // Skip environment variables
				SkipUserConfig:      true, // Skip user config
				SkipVersionResolved: true,
			},
			expected: struct {
				version string
				source  ResolutionSource
			}{
				version: env.mockSystemPerl.Version,
				source:  SystemPerlSource,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := ResolveVersion(&tt.options)
			if err != nil {
				t.Fatalf("Failed to resolve version: %v", err)
			}

			if resolved.Version != tt.expected.version {
				t.Errorf("Expected version %s, got %s", tt.expected.version, resolved.Version)
			}

			if resolved.Source != tt.expected.source {
				t.Errorf("Expected source %s, got %s", tt.expected.source, resolved.Source)
			}
		})
	}
}

// Test version resolution callback
func TestVersionResolvedCallback(t *testing.T) {
	env := setupResolverTest(t)
	defer env.cleanup_()

	availableVersions := []string{"5.38.0", "5.36.0", "5.34.1"}

	callbackCalled := false
	callbackVersion := ""

	// Set callback
	originalCallback := OnVersionResolved
	defer func() { OnVersionResolved = originalCallback }()

	OnVersionResolved = func(version *ResolvedVersion) {
		callbackCalled = true
		callbackVersion = version.Version
	}

	options := &ResolutionOptions{
		ExplicitVersion:   "5.38.0",
		AvailableVersions: availableVersions,
	}

	_, err := ResolveVersion(options)
	if err != nil {
		t.Fatalf("Failed to resolve version: %v", err)
	}

	if !callbackCalled {
		t.Errorf("Expected callback to be called")
	}

	if callbackVersion != "5.38.0" {
		t.Errorf("Expected callback version 5.38.0, got %s", callbackVersion)
	}

	// Test skipping callback
	callbackCalled = false
	options.SkipVersionResolved = true

	_, err = ResolveVersion(options)
	if err != nil {
		t.Fatalf("Failed to resolve version: %v", err)
	}

	if callbackCalled {
		t.Errorf("Expected callback to be skipped")
	}
}

// clearSystemPerlFromRegistry removes only system perl entries from the registry
// This is more targeted than clearing the entire registry and preserves other entries
func clearSystemPerlFromRegistry() error {
	// Load the current registry
	registry, err := loadRegistryFunc()
	if err != nil {
		return err
	}

	// Remove only entries with source="system"
	for id, versionInfo := range registry.Versions {
		if versionInfo.Source == "system" {
			delete(registry.Versions, id)
		}
	}

	// Save the modified registry
	return saveRegistryFunc(registry)
}
