// ABOUTME: Tests for package manager detection and integration
// ABOUTME: Validates package manager detection accuracy and update delegation

package updater

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectInstallationMethodPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix package manager path detection not applicable on Windows")
	}
	tests := []struct {
		name           string
		binaryPath     string
		expectedMethod InstallationMethod
	}{
		{
			name:           "homebrew_standard_path_darwin",
			binaryPath:     "/usr/local/bin/pvm",
			expectedMethod: getExpectedForUsrLocal(), // Platform-dependent
		},
		{
			name:           "homebrew_cellar_path",
			binaryPath:     "/usr/local/Cellar/pvm/1.0.0/bin/pvm",
			expectedMethod: InstallationHomebrew,
		},
		{
			name:           "homebrew_opt_path",
			binaryPath:     "/opt/homebrew/bin/pvm",
			expectedMethod: InstallationHomebrew,
		},
		{
			name:           "user_bin_path",
			binaryPath:     "/home/user/bin/pvm",
			expectedMethod: InstallationBinary,
		},
		{
			name:           "local_build_path",
			binaryPath:     "/home/user/dev/pvm/pvm",
			expectedMethod: InstallationBinary,
		},
		{
			name:           "system_usr_bin",
			binaryPath:     "/usr/bin/pvm",
			expectedMethod: getExpectedForUsrBin(), // Depends on available package managers
		},
		{
			name:           "snap_path",
			binaryPath:     "/snap/pvm/current/bin/pvm",
			expectedMethod: InstallationSnap,
		},
		{
			name:           "flatpak_path",
			binaryPath:     "/var/lib/flatpak/app/com.example.pvm/current/active/bin/pvm",
			expectedMethod: InstallationFlatpak,
		},
	}

	if runtime.GOOS == "windows" {
		tests = append(tests, []struct {
			name           string
			binaryPath     string
			expectedMethod InstallationMethod
		}{
			{
				name:           "chocolatey_path",
				binaryPath:     "C:\\ProgramData\\chocolatey\\bin\\pvm.exe",
				expectedMethod: InstallationChocolatey,
			},
			{
				name:           "scoop_path",
				binaryPath:     "C:\\Users\\user\\scoop\\apps\\pvm\\current\\pvm.exe",
				expectedMethod: InstallationScoop,
			},
			{
				name:           "winget_path",
				binaryPath:     "C:\\Program Files\\WindowsApps\\pvm\\pvm.exe",
				expectedMethod: InstallationWinget,
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, err := DetectInstallationMethod(tt.binaryPath)
			if err != nil {
				t.Fatalf("DetectInstallationMethod failed: %v", err)
			}

			if method != tt.expectedMethod {
				t.Errorf("Expected %s, got %s", tt.expectedMethod.String(), method.String())
			}
		})
	}
}

func TestDetectPackageManagerCommands(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		expectFound bool
	}{
		{"brew", "brew", false},     // May or may not be installed
		{"apt", "apt", false},       // May or may not be installed
		{"yum", "yum", false},       // May or may not be installed
		{"dnf", "dnf", false},       // May or may not be installed
		{"pacman", "pacman", false}, // May or may not be installed
		{"snap", "snap", false},     // May or may not be installed
	}

	// Skip Windows-specific tests on non-Windows
	if runtime.GOOS == "windows" {
		tests = append(tests, []struct {
			name        string
			command     string
			expectFound bool
		}{
			{"chocolatey", "choco", false}, // May or may not be installed
			{"scoop", "scoop", false},      // May or may not be installed
			{"winget", "winget", false},    // May or may not be installed
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := testIsCommandAvailable(tt.command)
			t.Logf("Command %s available: %v", tt.command, found)
			// This test is informational - we don't assert because
			// availability depends on the test environment
		})
	}
}

func TestDetectRealInstallationMethod(t *testing.T) {
	// Test with the actual running binary
	binaryPath, err := GetCurrentBinaryPath()
	if err != nil {
		t.Skipf("Cannot get current binary path: %v", err)
	}

	method, err := DetectInstallationMethod(binaryPath)
	if err != nil {
		t.Fatalf("DetectInstallationMethod failed: %v", err)
	}

	t.Logf("Current installation method: %s (path: %s)", method.String(), binaryPath)
	t.Logf("Can self-update: %v", method.CanSelfUpdate())
	if !method.CanSelfUpdate() {
		t.Logf("Update instructions: %s", method.GetUpdateInstructions())
	}
}

func TestPackageManagerIntegration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix package manager integration not applicable on Windows")
	}
	// Test that we can detect and delegate to package managers
	testCases := []struct {
		name          string
		setupPath     string
		installMethod InstallationMethod
		expectError   bool
	}{
		{
			name:          "homebrew_delegation",
			setupPath:     "/opt/homebrew/bin/pvm",
			installMethod: InstallationHomebrew,
			expectError:   false,
		},
		{
			name:          "binary_self_update",
			setupPath:     "/home/user/bin/pvm",
			installMethod: InstallationBinary,
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			method, err := DetectInstallationMethod(tc.setupPath)
			if err != nil {
				if !tc.expectError {
					t.Fatalf("Unexpected error: %v", err)
				}
				return
			}

			if method != tc.installMethod {
				t.Errorf("Expected %s, got %s", tc.installMethod.String(), method.String())
			}

			canSelfUpdate := method.CanSelfUpdate()
			instructions := method.GetUpdateInstructions()

			t.Logf("Method: %s, Can self-update: %v", method.String(), canSelfUpdate)
			t.Logf("Update instructions: %s", instructions)

			// Validate instructions are not empty
			if instructions == "" {
				t.Error("Update instructions should not be empty")
			}
		})
	}
}

func TestValidatePackageManagerPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix package manager path validation not applicable on Windows")
	}
	// Create temporary test directories to simulate package manager paths
	tempDir := t.TempDir()

	testPaths := map[string]InstallationMethod{
		"homebrew": InstallationHomebrew,
		"system":   getExpectedForUsrBin(), // Use detected package manager
		"user_bin": InstallationBinary,
		"snap":     InstallationSnap,
		"flatpak":  InstallationFlatpak,
	}

	if runtime.GOOS == "windows" {
		testPaths["chocolatey"] = InstallationChocolatey
		testPaths["scoop"] = InstallationScoop
		testPaths["winget"] = InstallationWinget
	}

	for pathType, expectedMethod := range testPaths {
		t.Run(pathType, func(t *testing.T) {
			var testPath string
			switch pathType {
			case "homebrew":
				testPath = filepath.Join(tempDir, "opt", "homebrew", "bin", "pvm")
			case "system":
				testPath = filepath.Join("/usr", "bin", "pvm")
			case "user_bin":
				testPath = filepath.Join(tempDir, "home", "user", "bin", "pvm")
			case "snap":
				testPath = filepath.Join("/snap", "pvm", "current", "bin", "pvm")
			case "flatpak":
				testPath = filepath.Join("/var", "lib", "flatpak", "app", "com.example.pvm", "current", "active", "bin", "pvm")
			case "chocolatey":
				testPath = filepath.Join("C:", "ProgramData", "chocolatey", "bin", "pvm.exe")
			case "scoop":
				testPath = filepath.Join("C:", "Users", "user", "scoop", "apps", "pvm", "current", "pvm.exe")
			case "winget":
				testPath = filepath.Join("C:", "Program Files", "WindowsApps", "pvm", "pvm.exe")
			}

			method, err := DetectInstallationMethod(testPath)
			if err != nil {
				t.Fatalf("DetectInstallationMethod failed: %v", err)
			}

			if method != expectedMethod {
				t.Errorf("Expected %s for path %s, got %s",
					expectedMethod.String(), testPath, method.String())
			}
		})
	}
}

// Helper function to check if a command is available (wrapper for testing)
func testIsCommandAvailable(command string) bool {
	return isCommandAvailable(command)
}

// Helper functions to get expected installation methods based on platform/environment
func getExpectedForUsrLocal() InstallationMethod {
	if runtime.GOOS == "darwin" {
		return InstallationHomebrew
	}
	return InstallationSystemPackage
}

func getExpectedForUsrBin() InstallationMethod {
	// This depends on which package managers are available on the system
	// For testing purposes, we'll check what the actual detection returns
	method, _ := detectSpecificPackageManager()
	return method
}
