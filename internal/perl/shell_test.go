package perl

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/xdg"
)

// Mock functions
var origGetDirs = xdg.GetDirs
var origExecutablePath = executablePath

// Setup/teardown helpers
func setupShellTest(t *testing.T) func() {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pvm-shell-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create subdirectories
	configDir := filepath.Join(tmpDir, "config")
	cacheDir := filepath.Join(tmpDir, "cache")
	dataDir := filepath.Join(tmpDir, "data")
	stateDir := filepath.Join(tmpDir, "state")
	shimsDir := filepath.Join(dataDir, "pvm", "shims")
	shellDir := filepath.Join(dataDir, "pvm", "shell")

	// Create necessary directories
	dirs := []string{configDir, cacheDir, dataDir, stateDir, shimsDir, shellDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Mock xdg.GetDirs function
	xdg.GetDirs = func() (*xdg.Dirs, error) {
		dirs := &xdg.Dirs{
			ConfigHome: configDir,
			CacheHome:  cacheDir,
			DataHome:   dataDir,
			StateHome:  stateDir,
			ConfigDir:  filepath.Join(configDir, "pvm"),
			CacheDir:   filepath.Join(cacheDir, "pvm"),
			DataDir:    filepath.Join(dataDir, "pvm"),
			StateDir:   filepath.Join(stateDir, "pvm"),
			ShimsDir:   shimsDir,
		}
		dirs.EnsureDirs = func() error { return nil }
		return dirs, nil
	}

	// Mock executablePath function
	executablePath = func() (string, error) {
		return filepath.Join(tmpDir, "pvm"), nil
	}

	// Return a cleanup function
	return func() {
		// Restore original functions
		xdg.GetDirs = origGetDirs
		executablePath = origExecutablePath

		// Remove temporary directory
		_ = os.RemoveAll(tmpDir)
	}
}

// TestDetectShell_PowerShellOnAnyOS verifies that PSModulePath takes priority over SHELL
func TestDetectShell_PowerShellOnAnyOS(t *testing.T) {
	origPSModulePath := os.Getenv("PSModulePath")
	origShell := os.Getenv("SHELL")
	defer func() {
		os.Setenv("PSModulePath", origPSModulePath)
		os.Setenv("SHELL", origShell)
	}()

	os.Setenv("PSModulePath", "/some/path")
	os.Setenv("SHELL", "/bin/bash")

	shell, err := DetectShell()
	if err != nil {
		t.Fatalf("DetectShell() error: %v", err)
	}
	if shell != ShellPowerShell {
		t.Errorf("expected PowerShell, got %s", shell)
	}
}

// TestDetectShell_PVMShellOverridesLoginShell verifies that PVM_SHELL
// (set by the shell integration templates) takes precedence over $SHELL.
// Without this, a fish user whose login shell is bash would get POSIX
// output for fish-invoked pvm commands.
func TestDetectShell_PVMShellOverridesLoginShell(t *testing.T) {
	origPVMShell := os.Getenv("PVM_SHELL")
	origShell := os.Getenv("SHELL")
	origPSModulePath := os.Getenv("PSModulePath")
	defer func() {
		_ = os.Setenv("PVM_SHELL", origPVMShell)
		_ = os.Setenv("SHELL", origShell)
		_ = os.Setenv("PSModulePath", origPSModulePath)
	}()
	_ = os.Setenv("PSModulePath", "")
	_ = os.Setenv("SHELL", "/bin/bash")

	cases := []struct {
		pvmShell string
		want     ShellType
	}{
		{"fish", ShellFish},
		{"zsh", ShellZsh},
		{"bash", ShellBash},
		{"powershell", ShellPowerShell},
	}
	for _, tc := range cases {
		_ = os.Setenv("PVM_SHELL", tc.pvmShell)
		got, err := DetectShell()
		if err != nil {
			t.Fatalf("DetectShell() error with PVM_SHELL=%q: %v", tc.pvmShell, err)
		}
		if got != tc.want {
			t.Errorf("PVM_SHELL=%q SHELL=/bin/bash: got %q, want %q", tc.pvmShell, got, tc.want)
		}
	}
}

// Test shell detection
func TestDetectShell(t *testing.T) {
	// Save and clear PSModulePath so PowerShell doesn't override SHELL-based detection
	origPSModulePath := os.Getenv("PSModulePath")
	os.Unsetenv("PSModulePath")
	defer func() {
		if origPSModulePath != "" {
			os.Setenv("PSModulePath", origPSModulePath)
		}
	}()

	// PVM_SHELL would take precedence over SHELL; clear it so these cases
	// exercise the SHELL fallback path.
	origPVMShell := os.Getenv("PVM_SHELL")
	os.Unsetenv("PVM_SHELL")
	defer func() {
		if origPVMShell != "" {
			os.Setenv("PVM_SHELL", origPVMShell)
		}
	}()

	// Save original environment variables
	origShell := os.Getenv("SHELL")
	defer func() { _ = os.Setenv("SHELL", origShell) }()

	// Test cases
	// Default shell when SHELL env is empty depends on OS
	defaultShell := ShellBash
	if runtime.GOOS == "windows" {
		defaultShell = ShellCmd
	}

	testCases := []struct {
		shellEnv string
		expected ShellType
	}{
		{"/bin/bash", ShellBash},
		{"/usr/bin/zsh", ShellZsh},
		{"/usr/bin/fish", ShellFish},
		{"", defaultShell}, // Default case (OS-dependent)
	}

	for _, tc := range testCases {
		_ = os.Setenv("SHELL", tc.shellEnv)
		shell, err := DetectShell()
		if err != nil {
			t.Errorf("DetectShell() error = %v", err)
			continue
		}

		if shell != tc.expected {
			t.Errorf("DetectShell() got = %v, want %v", shell, tc.expected)
		}
	}
}

// Test shell script generation
func TestGenerateShellScript(t *testing.T) {
	cleanup := setupShellTest(t)
	defer cleanup()

	// Create shell script data
	data := ShellScriptData{
		PVMPath:          "/usr/local/bin/pvm",
		ShimsDir:         "/home/user/.local/share/pvm/shims",
		ConfigDir:        "/home/user/.config/pvm",
		FunctionPrefix:   "pvm_",
		SupportsAdvanced: true,
	}

	// Test generating script for each shell type
	shellTypes := []ShellType{
		ShellBash,
		ShellZsh,
		ShellFish,
		ShellPowerShell,
		ShellCmd,
	}

	for _, shellType := range shellTypes {
		script, err := GenerateShellScript(shellType, data)
		if err != nil {
			t.Errorf("GenerateShellScript(%v) error = %v", shellType, err)
			continue
		}

		// Check for expected content based on shell type
		switch shellType {
		case ShellBash, ShellZsh:
			// These should have content and proper structure
			if script == "" {
				t.Errorf("GenerateShellScript(%v) returned empty script", shellType)
				continue
			}
			if !strings.Contains(script, "#!/usr/bin/env sh") {
				t.Errorf("Bash/Zsh script missing shebang")
			}
			// For bash/zsh, PVMPath should appear as the fallback path in the template
			if !strings.Contains(script, data.PVMPath) {
				t.Errorf("Bash/Zsh script (%v) does not contain PVM path", shellType)
			}
		case ShellFish:
			// Fish should have content and proper structure
			if script == "" {
				t.Errorf("GenerateShellScript(%v) returned empty script", shellType)
				continue
			}
			if !strings.Contains(script, "#!/usr/bin/env fish") {
				t.Errorf("Fish script missing shebang")
			}
			// For fish, PVMPath should appear as the fallback path in the template
			if !strings.Contains(script, data.PVMPath) {
				t.Errorf("Fish script does not contain PVM path")
			}
			// Fish template uses XDG_BIN_HOME for tool shims (no longer uses deprecated ShimsDir)
			if !strings.Contains(script, "XDG_BIN_HOME") {
				t.Errorf("Fish script does not contain XDG_BIN_HOME integration")
			}
		case ShellPowerShell:
			// PowerShell template is implemented - check for proper structure
			if script == "" {
				t.Errorf("GenerateShellScript(%v) returned empty script", shellType)
				continue
			}
			if !strings.Contains(script, "ABOUTME: PVM Shell Integration for PowerShell") {
				t.Errorf("PowerShell script missing ABOUTME header")
			}
			if !strings.Contains(script, data.PVMPath) {
				t.Errorf("PowerShell script does not contain PVM path")
			}
			if !strings.Contains(script, "{{.FortuneQuote}}") {
				// After template execution, the fortune quote placeholder should be replaced
				// (empty string in test data means it stays empty or is substituted)
			}
		case ShellCmd:
			if script == "" {
				t.Errorf("GenerateShellScript(%v) returned empty script", shellType)
				continue
			}
			if !strings.Contains(script, "ABOUTME: PVM Shell Integration for CMD") {
				t.Errorf("CMD script missing ABOUTME header")
			}
			if !strings.Contains(script, "@echo off") {
				t.Errorf("CMD script missing @echo off")
			}
			if !strings.Contains(script, data.PVMPath) {
				t.Errorf("CMD script does not contain PVM path")
			}
		}
	}
}

// Test creating shell initialization scripts
func TestCreateShellInitScripts(t *testing.T) {
	cleanup := setupShellTest(t)
	defer cleanup()

	// Create shell initialization scripts
	err := CreateShellInitScripts()
	if err != nil {
		t.Fatalf("CreateShellInitScripts() error = %v", err)
	}

	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		t.Fatalf("GetDirs() error = %v", err)
	}

	// Check that scripts were created
	shellDir := filepath.Join(dirs.DataDir, "shell")
	extensions := []string{".bash", ".zsh", ".fish"}
	if runtime.GOOS == "windows" {
		extensions = append(extensions, ".ps1", ".cmd")
	}

	for _, ext := range extensions {
		scriptPath := filepath.Join(shellDir, "pvm"+ext)
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			t.Errorf("Script %s was not created", scriptPath)
		}
	}
}

// Test shell initialization command
func TestGetShellInitCommand(t *testing.T) {
	cleanup := setupShellTest(t)
	defer cleanup()

	// Test for each shell type
	shellTypes := []ShellType{
		ShellBash,
		ShellZsh,
		ShellFish,
		ShellPowerShell,
		ShellCmd,
	}

	for _, shellType := range shellTypes {
		cmd, err := GetShellInitCommand(shellType)
		if err != nil {
			t.Errorf("GetShellInitCommand(%v) error = %v", shellType, err)
			continue
		}

		// Check that command is not empty
		if cmd == "" {
			t.Errorf("GetShellInitCommand(%v) returned empty command", shellType)
			continue
		}

		// Check command format based on shell type
		switch shellType {
		case ShellBash, ShellZsh:
			if !strings.Contains(cmd, "source") {
				t.Errorf("Bash/Zsh init command should contain 'source'")
			}
		case ShellFish:
			if !strings.Contains(cmd, "source") {
				t.Errorf("Fish init command should contain 'source'")
			}
		case ShellPowerShell:
			if !strings.Contains(cmd, ".") {
				t.Errorf("PowerShell init command should contain '.'")
			}
		}
	}
}

// Test shell completion command
func TestGetShellCompletionCommand(t *testing.T) {
	// Test for each shell type
	shellTypes := []ShellType{
		ShellBash,
		ShellZsh,
		ShellFish,
		ShellPowerShell,
	}

	for _, shellType := range shellTypes {
		cmd, err := GetShellCompletionCommand(shellType)
		if err != nil {
			t.Errorf("GetShellCompletionCommand(%v) error = %v", shellType, err)
			continue
		}

		// Check that command is not empty
		if cmd == "" {
			t.Errorf("GetShellCompletionCommand(%v) returned empty command", shellType)
			continue
		}

		// Check command format based on shell type
		switch shellType {
		case ShellBash:
			if !strings.Contains(cmd, "complete") {
				t.Errorf("Bash completion command should contain 'complete'")
			}
		case ShellZsh:
			if !strings.Contains(cmd, "compdef") {
				t.Errorf("Zsh completion command should contain 'compdef'")
			}
		case ShellFish:
			if !strings.Contains(cmd, "complete") {
				t.Errorf("Fish completion command should contain 'complete'")
			}
		case ShellPowerShell:
			if !strings.Contains(cmd, "Register-ArgumentCompleter") {
				t.Errorf("PowerShell completion command should contain 'Register-ArgumentCompleter'")
			}
		}
	}
}

// Test version switching
func TestVersionSwitching(t *testing.T) {
	cleanup := setupShellTest(t)
	defer cleanup()

	// Use a function variable for mocking
	origValidateVersionFunc := ValidateVersion
	// Create a simple mock that just returns nil
	ValidateVersion = func(version string) error {
		return nil
	}
	// Restore original function when done
	defer func() {
		ValidateVersion = origValidateVersionFunc
	}()

	// Test global version setting
	err := SetGlobalVersion("5.36.0")
	if err != nil {
		t.Errorf("SetGlobalVersion() error = %v", err)
	}

	// Check that config file was created with correct version
	dirs, _ := xdg.GetDirs()
	configPath := dirs.GetConfigFilePath()
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Errorf("Failed to read config file: %v", err)
	} else if !strings.Contains(string(content), "default_perl = \"5.36.0\"") {
		t.Errorf("Config file does not contain expected version: %s", content)
	}

	// Test local version setting
	// Create a temporary directory for local version test
	tmpDir, err := os.MkdirTemp("", "pvm-local-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Change to the temporary directory
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tmpDir)

	// Set local version
	err = SetLocalVersion("5.38.0")
	if err != nil {
		t.Errorf("SetLocalVersion() error = %v", err)
	}

	// Check that .perl-version file was created with correct version
	versionFilePath := filepath.Join(tmpDir, ".perl-version")
	content, err = os.ReadFile(versionFilePath)
	if err != nil {
		t.Errorf("Failed to read .perl-version file: %v", err)
	} else if strings.TrimSpace(string(content)) != "5.38.0" {
		t.Errorf("Version file does not contain expected version: %s", content)
	}
}

// Test .perl-version file handling
func TestPerlVersionFile(t *testing.T) {
	// Create a test directory structure with .perl-version files
	rootDir, err := os.MkdirTemp("", "pvm-version-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(rootDir) }()

	// Create nested directories
	subDir1 := filepath.Join(rootDir, "dir1")
	subDir2 := filepath.Join(subDir1, "dir2")
	subDir3 := filepath.Join(subDir2, "dir3")

	// Create directories
	for _, dir := range []string{subDir1, subDir2, subDir3} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create .perl-version files
	versionFiles := map[string]string{
		rootDir: "5.30.0",
		subDir1: "5.32.0",
		subDir3: "5.38.0",
	}

	for dir, version := range versionFiles {
		versionFile := filepath.Join(dir, ".perl-version")
		if err := os.WriteFile(versionFile, []byte(version), 0644); err != nil {
			t.Fatalf("Failed to write version file %s: %v", versionFile, err)
		}
	}

	// Test FindDotPerlVersionFiles from different starting points
	testCases := []struct {
		startDir      string
		expectedCount int
		expectedFirst string
	}{
		{subDir3, 3, filepath.Join(subDir3, ".perl-version")},
		{subDir2, 2, filepath.Join(subDir1, ".perl-version")},
		{subDir1, 2, filepath.Join(subDir1, ".perl-version")},
		{rootDir, 1, filepath.Join(rootDir, ".perl-version")},
	}

	for _, tc := range testCases {
		files, err := FindDotPerlVersionFiles(tc.startDir)
		if err != nil {
			t.Errorf("FindDotPerlVersionFiles(%s) error = %v", tc.startDir, err)
			continue
		}

		if len(files) != tc.expectedCount {
			t.Errorf("FindDotPerlVersionFiles(%s) got %d files, want %d", tc.startDir, len(files), tc.expectedCount)
			continue
		}

		if len(files) > 0 && files[0] != tc.expectedFirst {
			t.Errorf("FindDotPerlVersionFiles(%s) first file = %s, want %s", tc.startDir, files[0], tc.expectedFirst)
		}
	}

	// Test ReadPerlVersionFile
	for dir, expectedVersion := range versionFiles {
		version, err := ReadPerlVersionFile(dir)
		if err != nil {
			t.Errorf("ReadPerlVersionFile(%s) error = %v", dir, err)
			continue
		}

		if version != expectedVersion {
			t.Errorf("ReadPerlVersionFile(%s) = %s, want %s", dir, version, expectedVersion)
		}
	}
}

// Test shell init check
func TestCheckShellInit(t *testing.T) {
	// This test works on all platforms with appropriate shell configurations

	// Create a fake shell config file for testing
	tmpDir, err := os.MkdirTemp("", "pvm-shell-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Mock environment and user home directory (platform-specific)
	var origHome string
	var homeEnvVar string
	if runtime.GOOS == "windows" {
		homeEnvVar = "USERPROFILE"
		origHome = os.Getenv("USERPROFILE")
		_ = os.Setenv("USERPROFILE", tmpDir)
	} else {
		homeEnvVar = "HOME"
		origHome = os.Getenv("HOME")
		_ = os.Setenv("HOME", tmpDir)
	}
	defer func() { _ = os.Setenv(homeEnvVar, origHome) }()

	// Create platform-appropriate shell config files
	var files map[string]string

	if runtime.GOOS == "windows" {
		// Windows shell configurations
		powershellPath := filepath.Join(tmpDir, "Microsoft.PowerShell_profile.ps1")
		cmdPath := filepath.Join(tmpDir, "pvm_init.bat")
		files = map[string]string{
			powershellPath: "# PowerShell config\n$env:PATH = \"$env:PATH;C:\\tools\\bin\"\n",
			cmdPath:        "# CMD config\nREM PVM initialization would go here\n",
		}
	} else {
		// Unix shell configurations
		bashrcPath := filepath.Join(tmpDir, ".bashrc")
		zshrcPath := filepath.Join(tmpDir, ".zshrc")
		fishConfigDir := filepath.Join(tmpDir, ".config", "fish")
		fishConfigPath := filepath.Join(fishConfigDir, "config.fish")

		// Create fish config directory
		if err := os.MkdirAll(fishConfigDir, 0755); err != nil {
			t.Fatalf("Failed to create fish config directory: %v", err)
		}

		files = map[string]string{
			bashrcPath:     "# Bash config\nexport PATH=$PATH:/usr/local/bin\n",
			zshrcPath:      "# Zsh config\neval \"$(pvm init)\"\n",
			fishConfigPath: "# Fish config\nset -x PATH /usr/local/bin $PATH\n",
		}
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}

	// Test checking initialization status (platform-specific shells)
	var testCases []struct {
		shellType      ShellType
		expectedStatus bool
	}

	if runtime.GOOS == "windows" {
		testCases = []struct {
			shellType      ShellType
			expectedStatus bool
		}{
			{ShellPowerShell, false},
			{ShellCmd, false},
		}
	} else {
		testCases = []struct {
			shellType      ShellType
			expectedStatus bool
		}{
			{ShellBash, false},
			{ShellZsh, true},
			{ShellFish, false},
		}
	}

	for _, tc := range testCases {
		initialized, _, err := CheckShellInit(tc.shellType)
		if err != nil {
			t.Errorf("CheckShellInit(%v) error = %v", tc.shellType, err)
			continue
		}

		if initialized != tc.expectedStatus {
			t.Errorf("CheckShellInit(%v) = %v, want %v", tc.shellType, initialized, tc.expectedStatus)
		}
	}
}

// Test shell initialization instructions
func TestGetShellInitInstructions(t *testing.T) {
	shellTypes := []ShellType{
		ShellBash,
		ShellZsh,
		ShellFish,
		ShellPowerShell,
		ShellCmd,
		ShellUnknown,
	}

	for _, shellType := range shellTypes {
		instructions := GetShellInitInstructions(shellType)
		if instructions == "" {
			t.Errorf("GetShellInitInstructions(%v) returned empty instructions", shellType)
			continue
		}

		// Check that instructions contain relevant commands based on shell type
		switch shellType {
		case ShellBash:
			if !strings.Contains(instructions, "~/.bashrc") {
				t.Errorf("Bash instructions should mention ~/.bashrc")
			}
		case ShellZsh:
			if !strings.Contains(instructions, "~/.zshrc") {
				t.Errorf("Zsh instructions should mention ~/.zshrc")
			}
		case ShellFish:
			if !strings.Contains(instructions, "config.fish") {
				t.Errorf("Fish instructions should mention config.fish")
			}
		case ShellPowerShell:
			if !strings.Contains(instructions, "PowerShell profile") {
				t.Errorf("PowerShell instructions should mention PowerShell profile")
			}
		case ShellCmd:
			if !strings.Contains(instructions, "CMD") {
				t.Errorf("CMD instructions should mention CMD")
			}
		}
	}
}

// TestGenerateShellUse tests the GenerateShellUse function with library support
func TestGenerateShellUse(t *testing.T) {
	cleanup := setupShellTest(t)
	defer cleanup()

	// Mock validation functions
	origValidateVersionFunc := ValidateVersion

	ValidateVersion = func(version string) error {
		if version == "5.38.0" || version == "5.42.0" {
			return nil
		}
		return errors.NewVersionError(ErrUnsatisfiedVersion, "Version not available", nil)
	}

	defer func() {
		ValidateVersion = origValidateVersionFunc
	}()

	// Create test library environments
	dirs, err := xdg.GetDirs()
	if err != nil {
		t.Fatalf("Failed to get XDG dirs: %v", err)
	}

	// Create test library environments
	testLibraries := []string{"testlib", "tools"}
	for _, lib := range testLibraries {
		envDir := filepath.Join(dirs.DataDir, "environments", lib)
		if err := os.MkdirAll(envDir, 0755); err != nil {
			t.Fatalf("Failed to create test library environment %s: %v", lib, err)
		}
	}

	// Save and clear PSModulePath so DetectShell doesn't pick up PowerShell
	// when testing bash/fish cases
	origPSModulePath := os.Getenv("PSModulePath")
	_ = os.Setenv("PSModulePath", "")
	defer func() { _ = os.Setenv("PSModulePath", origPSModulePath) }()

	// Save original SHELL environment variable
	origShell := os.Getenv("SHELL")
	defer func() { _ = os.Setenv("SHELL", origShell) }()

	testCases := []struct {
		name            string
		version         string
		library         string
		shell           string
		psModulePath    string // non-empty triggers PowerShell detection
		wantError       bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:      "ValidVersionWithLibrary",
			version:   "5.38.0",
			library:   "testlib",
			shell:     "/bin/bash",
			wantError: false,
			wantContains: []string{
				"export PVM_PERL_VERSION='5.38.0'",
				"export PVM_PERL_LIBRARY='testlib'",
				"export PVM_PERL_VERSION_FULL='5.38.0@testlib'",
				"Using Perl 5.38.0@testlib",
			},
		},
		{
			name:      "ValidVersionWithoutLibrary",
			version:   "5.42.0",
			library:   "",
			shell:     "/bin/bash",
			wantError: false,
			wantContains: []string{
				"export PVM_PERL_VERSION='5.42.0'",
				"unset PVM_PERL_LIBRARY",
				"export PVM_PERL_VERSION_FULL='5.42.0'",
				"Using Perl 5.42.0",
			},
		},
		{
			name:      "ValidVersionWithLibraryFish",
			version:   "5.38.0",
			library:   "tools",
			shell:     "/usr/bin/fish",
			wantError: false,
			wantContains: []string{
				"set -gx PVM_PERL_VERSION '5.38.0'",
				"set -gx PVM_PERL_LIBRARY 'tools'",
				"set -gx PVM_PERL_VERSION_FULL '5.38.0@tools'",
				"Using Perl 5.38.0@tools",
			},
		},
		{
			name:      "ValidVersionWithoutLibraryFish",
			version:   "5.42.0",
			library:   "",
			shell:     "/usr/bin/fish",
			wantError: false,
			wantContains: []string{
				"set -gx PVM_PERL_VERSION '5.42.0'",
				"set -e PVM_PERL_LIBRARY 2>/dev/null; or true",
				"set -gx PVM_PERL_VERSION_FULL '5.42.0'",
				"Using Perl 5.42.0",
			},
		},
		{
			name:         "ValidVersionWithLibraryPowerShell",
			version:      "5.38.0",
			library:      "testlib",
			shell:        "",
			psModulePath: "/some/powershell/path",
			wantError:    false,
			wantContains: []string{
				"$env:PVM_PERL_VERSION = '5.38.0'",
				"$env:PVM_PERL_LIBRARY = 'testlib'",
				"$env:PVM_PERL_VERSION_FULL = '5.38.0@testlib'",
				"Write-Host 'Using Perl 5.38.0@testlib'",
			},
		},
		{
			name:         "ValidVersionWithoutLibraryPowerShell",
			version:      "5.42.0",
			library:      "",
			shell:        "",
			psModulePath: "/some/powershell/path",
			wantError:    false,
			wantContains: []string{
				"$env:PVM_PERL_VERSION = '5.42.0'",
				"Remove-Item Env:PVM_PERL_LIBRARY -ErrorAction SilentlyContinue",
				"$env:PVM_PERL_VERSION_FULL = '5.42.0'",
				"Write-Host 'Using Perl 5.42.0'",
			},
		},
		{
			name:      "InvalidVersion",
			version:   "999.999.999",
			library:   "",
			shell:     "/bin/bash",
			wantError: true,
		},
		{
			name:      "InvalidLibrary",
			version:   "5.38.0",
			library:   "nonexistent",
			shell:     "/bin/bash",
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set SHELL environment variable
			_ = os.Setenv("SHELL", tc.shell)
			// Set PSModulePath for PowerShell detection (clear for non-PS cases)
			_ = os.Setenv("PSModulePath", tc.psModulePath)

			// Capture output by temporarily redirecting stdout
			origStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run the function
			err := GenerateShellUse(tc.version, tc.library)

			// Restore stdout and read output
			w.Close()
			os.Stdout = origStdout

			output := make([]byte, 1024)
			n, _ := r.Read(output)
			outputStr := string(output[:n])

			// Check error expectation
			if tc.wantError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.wantError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check expected content
			for _, want := range tc.wantContains {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Output missing expected string: %s\nOutput: %s", want, outputStr)
				}
			}

			// Check unwanted content
			for _, unwant := range tc.wantNotContains {
				if strings.Contains(outputStr, unwant) {
					t.Errorf("Output contains unwanted string: %s\nOutput: %s", unwant, outputStr)
				}
			}
		})
	}
}

// TestLibraryNameSecurity tests security validation for library names
func TestLibraryNameSecurity(t *testing.T) {
	cleanup := setupShellTest(t)
	defer cleanup()

	testCases := []struct {
		name          string
		libraryName   string
		expectError   bool
		errorContains string
	}{
		{
			name:        "ValidLibraryName",
			libraryName: "tools",
			expectError: false,
		},
		{
			name:        "ValidLibraryNameWithDash",
			libraryName: "my-tools",
			expectError: false,
		},
		{
			name:        "ValidLibraryNameWithUnderscore",
			libraryName: "test_env",
			expectError: false,
		},
		{
			name:          "PathTraversalAttack",
			libraryName:   "../../../etc",
			expectError:   true,
			errorContains: "invalid path characters",
		},
		{
			name:          "RelativePathAttack",
			libraryName:   "../../.ssh",
			expectError:   true,
			errorContains: "invalid path characters",
		},
		{
			name:          "AbsolutePathAttack",
			libraryName:   "/etc/passwd",
			expectError:   true,
			errorContains: "invalid path characters",
		},
		{
			name:          "WindowsPathAttack",
			libraryName:   "..\\..\\windows",
			expectError:   true,
			errorContains: "invalid path characters",
		},
		{
			name:          "ShellInjectionSingleQuote",
			libraryName:   "lib'; rm -rf /; echo 'hack",
			expectError:   true,
			errorContains: "invalid path characters",
		},
		{
			name:          "ShellInjectionBacktick",
			libraryName:   "lib`id`",
			expectError:   true,
			errorContains: "alphanumeric characters",
		},
		{
			name:          "ShellInjectionDollarSign",
			libraryName:   "lib$(whoami)",
			expectError:   true,
			errorContains: "alphanumeric characters",
		},
		{
			name:        "EmptyLibraryName",
			libraryName: "",
			expectError: false, // Empty is allowed
		},
		{
			name:          "WhitespaceOnlyLibrary",
			libraryName:   "   ",
			expectError:   true,
			errorContains: "whitespace-only",
		},
		{
			name:          "TooLongLibraryName",
			libraryName:   strings.Repeat("a", 65),
			expectError:   true,
			errorContains: "too long",
		},
		{
			name:          "SpecialCharacters",
			libraryName:   "lib@#$%",
			expectError:   true,
			errorContains: "alphanumeric characters",
		},
		{
			name:          "SpacesInName",
			libraryName:   "my tools",
			expectError:   true,
			errorContains: "alphanumeric characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := sanitizeLibraryName(tc.libraryName)

			if tc.expectError && err == nil {
				t.Fatalf("Expected error for library name '%s' but got none", tc.libraryName)
			}

			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error for library name '%s': %v", tc.libraryName, err)
			}

			if tc.expectError && tc.errorContains != "" {
				if !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Error message should contain '%s' but got: %s", tc.errorContains, err.Error())
				}
			}
		})
	}
}

// TestShellEscaping tests the shell escaping function
func TestShellEscaping(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"simple", "'simple'"},
		{"with space", "'with space'"},
		{"with'quote", "'with'\"'\"'quote'"},
		{"multiple'quotes'here", "'multiple'\"'\"'quotes'\"'\"'here'"},
		{"", "''"},
		{"$injection", "'$injection'"},
		{"`backtick`", "'`backtick`'"},
		{"$(command)", "'$(command)'"},
	}

	for _, tc := range testCases {
		t.Run("Escape_"+tc.input, func(t *testing.T) {
			result := escapeShellArg(tc.input)
			if result != tc.expected {
				t.Errorf("escapeShellArg(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestGenerateShellUseSystem verifies the "system" version case: it must
// detect the shell (fish uses `set -e`, PowerShell uses `Remove-Item`, POSIX
// uses `unset`) and must reject unsafe library names the same way the
// non-system path does, since the output is eval'd by the shell integration.
func TestGenerateShellUseSystem(t *testing.T) {
	cleanup := setupShellTest(t)
	defer cleanup()

	// Create a real library environment so ValidateLibraryEnvironment passes
	// for the "valid library" cases.
	dirs, err := xdg.GetDirs()
	if err != nil {
		t.Fatalf("Failed to get XDG dirs: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dirs.DataDir, "environments", "tools"), 0755); err != nil {
		t.Fatalf("Failed to create test library: %v", err)
	}

	origPSModulePath := os.Getenv("PSModulePath")
	_ = os.Setenv("PSModulePath", "")
	defer func() { _ = os.Setenv("PSModulePath", origPSModulePath) }()
	origShell := os.Getenv("SHELL")
	defer func() { _ = os.Setenv("SHELL", origShell) }()

	cases := []struct {
		name            string
		library         string
		shell           string
		psModulePath    string
		wantError       bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:            "bash system no library emits POSIX unset",
			shell:           "/bin/bash",
			wantContains:    []string{"unset PVM_PERL_VERSION", "unset PVM_PERL_LIBRARY"},
			wantNotContains: []string{"set -e PVM_PERL_VERSION", "Remove-Item"},
		},
		{
			name:            "fish system no library emits fish set -e",
			shell:           "/usr/bin/fish",
			wantContains:    []string{"set -e PVM_PERL_VERSION"},
			wantNotContains: []string{"unset PVM_PERL_VERSION"},
		},
		{
			name:            "powershell system no library emits Remove-Item",
			shell:           "/bin/bash",
			psModulePath:    "C:\\Program Files\\PowerShell\\Modules",
			wantContains:    []string{"Remove-Item Env:PVM_PERL_VERSION"},
			wantNotContains: []string{"unset PVM_PERL_VERSION"},
		},
		{
			name:    "bash system@validlib unsets both and references library in message",
			library: "tools",
			shell:   "/bin/bash",
			wantContains: []string{
				"unset PVM_PERL_VERSION",
				"unset PVM_PERL_LIBRARY",
				"Using system Perl with library",
				"tools",
			},
			wantNotContains: []string{
				"export PVM_PERL_LIBRARY",
			},
		},
		{
			name:      "system@<injection> is rejected",
			library:   "x';id #",
			shell:     "/bin/bash",
			wantError: true,
		},
		{
			name:      "system@<path traversal> is rejected",
			library:   "../../etc/passwd",
			shell:     "/bin/bash",
			wantError: true,
		},
		{
			name:      "system@<unknown> is rejected",
			library:   "nonexistent",
			shell:     "/bin/bash",
			wantError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Setenv("SHELL", tc.shell)
			_ = os.Setenv("PSModulePath", tc.psModulePath)

			origStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := GenerateShellUse("system", tc.library)

			w.Close()
			os.Stdout = origStdout
			output := make([]byte, 4096)
			n, _ := r.Read(output)
			outputStr := string(output[:n])

			if tc.wantError && err == nil {
				t.Fatalf("expected error but got none; output: %s", outputStr)
			}
			if !tc.wantError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, want := range tc.wantContains {
				if !strings.Contains(outputStr, want) {
					t.Errorf("output missing %q\nfull output: %s", want, outputStr)
				}
			}
			for _, unwant := range tc.wantNotContains {
				if strings.Contains(outputStr, unwant) {
					t.Errorf("output unexpectedly contains %q\nfull output: %s", unwant, outputStr)
				}
			}
		})
	}
}

// TestShellTemplatesExportPVMShell verifies each shell template exports
// PVM_SHELL so pvm subprocesses can detect the invoking shell reliably,
// regardless of the user's login $SHELL.
func TestShellTemplatesExportPVMShell(t *testing.T) {
	cases := []struct {
		name     string
		template string
		want     string
	}{
		{"bash", bashTemplate, "export PVM_SHELL=bash"},
		{"zsh", zshTemplate, "export PVM_SHELL=zsh"},
		{"fish", fishTemplate, "set -gx PVM_SHELL fish"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(tc.template, tc.want) {
				t.Errorf("%s template missing %q", tc.name, tc.want)
			}
		})
	}
}

// TestShellTemplatesRefreshPathAfterShUse verifies each shell integration
// template calls _pvm_update_perl_path after the sh-use eval. sh-use only
// exports PVM_PERL_VERSION — it does not rewrite PATH — so without the
// refresh, `perl` keeps resolving to the previously-active version's bin.
func TestShellTemplatesRefreshPathAfterShUse(t *testing.T) {
	cases := []struct {
		name     string
		template string
		refresh  string // the PATH-refresh call expected to follow sh-use
	}{
		{"bash", bashTemplate, "_pvm_update_perl_path"},
		{"zsh", zshTemplate, "_pvm_update_perl_path"},
		{"fish", fishTemplate, "_pvm_update_perl_path"},
	}

	// The refresh must appear between the sh-use eval and the next control-flow
	// boundary that ends the `use` branch (elif / else / end); anywhere later
	// in the template doesn't help the `use` path.
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			shUseIdx := strings.Index(tc.template, "sh-use")
			if shUseIdx < 0 {
				t.Fatalf("%s template does not mention sh-use", tc.name)
			}

			tail := tc.template[shUseIdx:]
			branchEnd := len(tail)
			for _, marker := range []string{"\n        elif ", "\n        else if ", "\n        else\n", "\n        end\n"} {
				if i := strings.Index(tail, marker); i >= 0 && i < branchEnd {
					branchEnd = i
				}
			}
			branchBody := tail[:branchEnd]

			if !strings.Contains(branchBody, tc.refresh) {
				t.Errorf("%s template: expected %q inside the `use` branch after the sh-use eval (to rewrite PATH for the new version), but the branch body is:\n---\n%s\n---", tc.name, tc.refresh, branchBody)
			}
		})
	}
}
