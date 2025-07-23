// ABOUTME: PVM doctor diagnostic functions
// ABOUTME: Implements comprehensive diagnostics for PVM installation issues

package pvm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/current"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/xdg"
)

// checkShellIntegration checks if shell integration is properly set up
func checkShellIntegration(ui *ui.Output, issues *[]string, warnings *[]string) error {
	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		*issues = append(*issues, "Failed to determine XDG directories")
		return err
	}

	// Check if shell integration files exist
	shellDir := filepath.Join(dirs.DataDir, "shell")
	shellFiles := []string{"pvm.bash", "pvm.zsh", "pvm.fish"}

	missingFiles := []string{}
	for _, file := range shellFiles {
		shellFile := filepath.Join(shellDir, file)
		if _, err := os.Stat(shellFile); os.IsNotExist(err) {
			missingFiles = append(missingFiles, file)
		}
	}

	if len(missingFiles) > 0 {
		*issues = append(*issues, fmt.Sprintf("Missing shell integration files: %v", missingFiles))
		ui.Error("Shell integration files not found. Run 'pvm shell init' to create them.")
		return nil
	}

	ui.Success("Shell integration files found")

	// Check if shell integration is loaded in shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		*warnings = append(*warnings, "SHELL environment variable not set")
		return nil
	}

	// Check shell configuration files
	homeDir, err := os.UserHomeDir()
	if err != nil {
		*warnings = append(*warnings, "Failed to get user home directory")
		return nil
	}

	shellConfigs := map[string][]string{
		"bash": {".bashrc", ".bash_profile"},
		"zsh":  {".zshrc", ".zprofile"},
		"fish": {".config/fish/config.fish"},
	}

	shellName := filepath.Base(shell)
	configs, exists := shellConfigs[shellName]
	if !exists {
		*warnings = append(*warnings, fmt.Sprintf("Unknown shell: %s", shellName))
		return nil
	}

	foundInitCall := false
	for _, config := range configs {
		configPath := filepath.Join(homeDir, config)
		if data, err := os.ReadFile(configPath); err == nil {
			if strings.Contains(string(data), "pvm init") {
				foundInitCall = true
				break
			}
		}
	}

	if !foundInitCall {
		*warnings = append(*warnings, "Shell configuration doesn't contain 'eval \"$(pvm init)\"'")
		ui.Warning("Add 'eval \"$(pvm init)\"' to your shell configuration file")
	} else {
		ui.Success("Shell integration properly configured")
	}

	return nil
}

// checkVersionManagement checks if version management is working correctly
func checkVersionManagement(ui *ui.Output, issues *[]string, warnings *[]string) error {
	// Check if any versions are installed
	installedVersions, err := perl.GetInstalledVersions()
	if err != nil {
		*issues = append(*issues, "Failed to get installed versions")
		return err
	}

	if len(installedVersions) == 0 {
		*warnings = append(*warnings, "No Perl versions installed")
		ui.Warning("No Perl versions installed. Run 'pvm install <version>' to install a version.")
		return nil
	}

	ui.Success("Found %d installed Perl version(s)", len(installedVersions))

	// Test version resolution - use current package for proper formatting
	currentInfo, err := current.GetCurrentVersion()
	if err != nil {
		*issues = append(*issues, "Failed to resolve current Perl version")
		return err
	}

	ui.Success("Current version resolves to: %s (%s)", currentInfo.Version, currentInfo.SourceDescription)

	return nil
}

// checkPathConfiguration checks if PATH is properly configured
func checkPathConfiguration(ui *ui.Output, issues *[]string, warnings *[]string) error {
	path := os.Getenv("PATH")
	if path == "" {
		*issues = append(*issues, "PATH environment variable not set")
		return nil
	}

	// Check if PVM executable is in PATH
	if pvmPath, err := exec.LookPath("pvm"); err == nil {
		ui.Success("PVM executable found in system PATH at %s", pvmPath)
	} else {
		*warnings = append(*warnings, "PVM executable not found in system PATH")
		ui.Warning("Consider adding PVM to your system PATH for easier access")
	}

	// Check if shims directory is in PATH
	dirs, err := xdg.GetDirs()
	if err != nil {
		*issues = append(*issues, "Failed to determine XDG directories")
		return err
	}

	shimsDir := filepath.Join(dirs.DataDir, "shims")
	pathDirs := strings.Split(path, ":")

	shimsInPath := false
	for _, dir := range pathDirs {
		if dir == shimsDir {
			shimsInPath = true
			break
		}
	}

	if !shimsInPath {
		*issues = append(*issues, "Shims directory not in PATH")
		ui.Error("Shims directory not in PATH. Shell integration may not be working.")
	} else {
		ui.Success("Shims directory found in PATH")
	}

	return nil
}

// checkEnvironmentVariables checks if environment variables are properly set
func checkEnvironmentVariables(ui *ui.Output, issues *[]string, warnings *[]string) error {
	// Check PVM-specific environment variables
	pvmVersion := os.Getenv("PVM_PERL_VERSION")
	if pvmVersion != "" {
		ui.Success("PVM_PERL_VERSION set to: %s", pvmVersion)
	} else {
		ui.Info("PVM_PERL_VERSION not set (this is normal if no version is explicitly selected)")
	}

	// Check for conflicts with other version managers
	plenvVersion := os.Getenv("PLENV_VERSION")
	if plenvVersion != "" {
		*warnings = append(*warnings, fmt.Sprintf("PLENV_VERSION is set: %s", plenvVersion))
	}

	perlbrewPerl := os.Getenv("PERLBREW_PERL")
	if perlbrewPerl != "" {
		*warnings = append(*warnings, fmt.Sprintf("PERLBREW_PERL is set: %s", perlbrewPerl))
	}

	return nil
}

// checkShimsDirectory checks if shims directory is properly set up
func checkShimsDirectory(ui *ui.Output, issues *[]string, warnings *[]string) error {
	dirs, err := xdg.GetDirs()
	if err != nil {
		*issues = append(*issues, "Failed to determine XDG directories")
		return err
	}

	shimsDir := filepath.Join(dirs.DataDir, "shims")

	// Check if shims directory exists
	if _, err := os.Stat(shimsDir); os.IsNotExist(err) {
		*issues = append(*issues, "Shims directory does not exist")
		ui.Error("Shims directory not found. Run 'pvm rehash' to create it.")
		return nil
	}

	// Check if shims directory contains expected files
	files, err := os.ReadDir(shimsDir)
	if err != nil {
		*issues = append(*issues, "Failed to read shims directory")
		return err
	}

	foundShims := []string{}

	for _, file := range files {
		if !file.IsDir() {
			foundShims = append(foundShims, file.Name())
		}
	}

	if len(foundShims) == 0 {
		*warnings = append(*warnings, "No shims found in shims directory")
		ui.Warning("No shims found. Run 'pvm rehash' to create shims.")
	} else {
		ui.Success("Found %d shim(s): %v", len(foundShims), foundShims)
	}

	return nil
}

// checkVersionManagerConflicts checks for conflicts with other version managers
func checkVersionManagerConflicts(ui *ui.Output, issues *[]string, warnings *[]string) error {
	path := os.Getenv("PATH")
	if path == "" {
		return nil
	}

	pathDirs := strings.Split(path, ":")
	conflicts := []string{}

	// Check for plenv
	for _, dir := range pathDirs {
		if strings.Contains(dir, "plenv") && strings.Contains(dir, "shims") {
			conflicts = append(conflicts, fmt.Sprintf("plenv (%s)", dir))
		}
		if strings.Contains(dir, "perlbrew") {
			conflicts = append(conflicts, fmt.Sprintf("perlbrew (%s)", dir))
		}
	}

	if len(conflicts) > 0 {
		*warnings = append(*warnings, fmt.Sprintf("Detected other Perl version managers in PATH: %v", conflicts))
		ui.Warning("Other Perl version managers detected. This may cause conflicts.")
		ui.Info("Consider removing them from PATH or setting PVM_SUPPRESS_WARNINGS=1")
	} else {
		ui.Success("No conflicting version managers detected")
	}

	return nil
}
