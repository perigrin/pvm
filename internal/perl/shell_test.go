package perl

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

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

// Test shell detection
func TestDetectShell(t *testing.T) {
	// Save original environment variables
	origShell := os.Getenv("SHELL")
	defer func() { _ = os.Setenv("SHELL", origShell) }()

	// Test cases
	testCases := []struct {
		shellEnv  string
		expected  ShellType
		needsSkip bool
	}{
		{"/bin/bash", ShellBash, false},
		{"/usr/bin/zsh", ShellZsh, false},
		{"/usr/bin/fish", ShellFish, false},
		{"", ShellBash, false},                               // Default case
		{"powershell", ShellBash, runtime.GOOS != "windows"}, // Skip on non-Windows
	}

	for _, tc := range testCases {
		if tc.needsSkip && runtime.GOOS != "windows" {
			continue
		}

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

		// Basic validation of script content
		if script == "" {
			t.Errorf("GenerateShellScript(%v) returned empty script", shellType)
			continue
		}

		// Check for expected content based on shell type
		switch shellType {
		case ShellBash, ShellZsh:
			if !strings.Contains(script, "#!/usr/bin/env sh") {
				t.Errorf("Bash/Zsh script missing shebang")
			}
		case ShellFish:
			if !strings.Contains(script, "#!/usr/bin/env fish") {
				t.Errorf("Fish script missing shebang")
			}
		case ShellPowerShell:
			if !strings.Contains(script, "# PVM Shell Integration for PowerShell") {
				t.Errorf("PowerShell script missing header")
			}
		case ShellCmd:
			if !strings.Contains(script, "@echo off") {
				t.Errorf("CMD script missing header")
			}
		}

		// Check for common elements
		if !strings.Contains(script, data.PVMPath) {
			t.Errorf("Script does not contain PVM path")
		}
		if !strings.Contains(script, data.ShimsDir) {
			t.Errorf("Script does not contain shims directory")
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
	} else {
		if !strings.Contains(string(content), "default_perl = \"5.36.0\"") {
			t.Errorf("Config file does not contain expected version: %s", content)
		}
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
	} else {
		if strings.TrimSpace(string(content)) != "5.38.0" {
			t.Errorf("Version file does not contain expected version: %s", content)
		}
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
	// Skip this test on non-Unix platforms
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	// Create a fake shell config file for testing
	tmpDir, err := os.MkdirTemp("", "pvm-shell-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Mock environment and user home directory
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Create example shell config files
	bashrcPath := filepath.Join(tmpDir, ".bashrc")
	zshrcPath := filepath.Join(tmpDir, ".zshrc")
	fishConfigDir := filepath.Join(tmpDir, ".config", "fish")
	fishConfigPath := filepath.Join(fishConfigDir, "config.fish")

	// Create fish config directory
	if err := os.MkdirAll(fishConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create fish config directory: %v", err)
	}

	// Create files with and without PVM initialization
	files := map[string]string{
		bashrcPath:     "# Bash config\nexport PATH=$PATH:/usr/local/bin\n",
		zshrcPath:      "# Zsh config\neval \"$(pvm init)\"\n",
		fishConfigPath: "# Fish config\nset -x PATH /usr/local/bin $PATH\n",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}

	// Test checking initialization status
	testCases := []struct {
		shellType      ShellType
		expectedStatus bool
	}{
		{ShellBash, false},
		{ShellZsh, true},
		{ShellFish, false},
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
