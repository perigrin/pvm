// ABOUTME: Shell integration specific error types and context-aware error messages
// ABOUTME: Provides detailed diagnostics and remediation suggestions for shell integration issues

package errors

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Shell integration error codes
const (
	ErrShellIntegrationMissing     = "601" // Shell integration not detected
	ErrShellConfigMissing          = "602" // Shell config file missing pvm init
	ErrShellEnvironmentIncorrect   = "603" // Shell environment not properly set up
	ErrShellVersionManagerConflict = "604" // Other version managers detected
	ErrShellDirectoryMissing       = "605" // Required directories missing
	ErrShellPermissionError        = "606" // Permission issues with shell files
)

// ShellIntegrationError provides context-aware shell integration error messages
type ShellIntegrationError struct {
	*EnhancedError
	shellType         string
	configFiles       []string
	detectedConflicts []string
	currentPath       string
	homeDir           string
}

// NewShellIntegrationError creates a shell integration error with context
func NewShellIntegrationError(code, message string, inner error, shell string) *ShellIntegrationError {
	base := NewEnhancedError(PrefixPVM, CategorySystem, code, message, inner, SeverityError)

	homeDir, _ := os.UserHomeDir()
	currentPath := os.Getenv("PATH")

	sie := &ShellIntegrationError{
		EnhancedError:     base,
		shellType:         shell,
		configFiles:       getShellConfigFiles(shell, homeDir),
		detectedConflicts: detectVersionManagerConflicts(currentPath),
		currentPath:       currentPath,
		homeDir:           homeDir,
	}

	sie.generateContextualGuidance()
	return sie
}

// generateContextualGuidance creates shell-specific remediation suggestions
func (sie *ShellIntegrationError) generateContextualGuidance() {
	switch sie.Code() {
	case ErrShellIntegrationMissing:
		sie.addShellIntegrationGuidance()
	case ErrShellConfigMissing:
		sie.addShellConfigGuidance()
	case ErrShellEnvironmentIncorrect:
		sie.addEnvironmentGuidance()
	case ErrShellVersionManagerConflict:
		sie.addConflictGuidance()
	case ErrShellDirectoryMissing:
		sie.addDirectoryGuidance()
	}
}

// addShellIntegrationGuidance provides guidance for missing shell integration
func (sie *ShellIntegrationError) addShellIntegrationGuidance() {
	sie.WithSeverity(SeverityError)

	// Add shell-specific initialization instructions
	switch sie.shellType {
	case "bash":
		sie.WithRecoveryAction("Add 'eval \"$(pvm init)\"' to ~/.bashrc")
		sie.WithRecoveryAction("Run: echo 'eval \"$(pvm init)\"' >> ~/.bashrc")
		sie.WithRecoveryAction("Reload shell: source ~/.bashrc")
		if sie.fileExists(filepath.Join(sie.homeDir, ".bash_profile")) {
			sie.WithContext("alternative_config", "You may also need to add to ~/.bash_profile for login shells")
		}
	case "zsh":
		sie.WithRecoveryAction("Add 'eval \"$(pvm init)\"' to ~/.zshrc")
		sie.WithRecoveryAction("Run: echo 'eval \"$(pvm init)\"' >> ~/.zshrc")
		sie.WithRecoveryAction("Reload shell: source ~/.zshrc")
	case "fish":
		sie.WithRecoveryAction("Add 'pvm init | source' to ~/.config/fish/config.fish")
		sie.WithRecoveryAction("Run: echo 'pvm init | source' >> ~/.config/fish/config.fish")
		sie.WithRecoveryAction("Reload shell: source ~/.config/fish/config.fish")
	default:
		sie.WithRecoveryAction("Add shell integration to your shell's configuration file")
		sie.WithRecoveryAction("Run: eval \"$(pvm init)\" (for POSIX shells)")
	}

	// Add verification step
	sie.WithRecoveryAction("Verify installation: pvm self doctor")
	sie.WithRecoveryAction("Quick fix: pvm self doctor --fix")

	// Add context about detected configuration
	sie.WithContext("detected_shell", sie.shellType)
	if len(sie.configFiles) > 0 {
		sie.WithContext("config_files", strings.Join(sie.configFiles, ", "))
	}
}

// addShellConfigGuidance provides guidance for shell configuration issues
func (sie *ShellIntegrationError) addShellConfigGuidance() {
	sie.WithSeverity(SeverityWarning)

	configFile := sie.getPrimaryConfigFile()
	if configFile != "" {
		sie.WithRecoveryAction(fmt.Sprintf("Add 'eval \"$(pvm init)\"' to %s", configFile))
		sie.WithContext("primary_config_file", configFile)

		// Check if config file exists
		fullPath := filepath.Join(sie.homeDir, configFile)
		if !sie.fileExists(fullPath) {
			sie.WithRecoveryAction(fmt.Sprintf("Create config file: touch %s", fullPath))
			sie.WithContext("config_file_missing", true)
		}
	}

	sie.WithRecoveryAction("Restart your shell or run 'source <config_file>'")
	sie.WithRecoveryAction("Auto-fix: pvm self doctor --fix")
}

// addEnvironmentGuidance provides guidance for environment setup issues
func (sie *ShellIntegrationError) addEnvironmentGuidance() {
	sie.WithSeverity(SeverityError)

	// Check what's wrong with the environment
	xdgBinHome := os.Getenv("XDG_BIN_HOME")
	if xdgBinHome == "" {
		xdgBinHome = filepath.Join(sie.homeDir, ".local", "bin")
	}

	if !strings.Contains(sie.currentPath, xdgBinHome) {
		sie.WithRecoveryAction("XDG_BIN_HOME is not in PATH")
		sie.WithRecoveryAction("This usually means shell integration is not active")
		sie.WithRecoveryAction("Run: eval \"$(pvm init)\" to activate in current session")
		sie.WithContext("xdg_bin_home", xdgBinHome)
		sie.WithContext("path_contains_xdg", false)
	}

	// Check for missing directories
	if !sie.dirExists(xdgBinHome) {
		sie.WithRecoveryAction(fmt.Sprintf("Create directory: mkdir -p %s", xdgBinHome))
		sie.WithContext("xdg_bin_home_missing", true)
	}
}

// addConflictGuidance provides guidance for version manager conflicts
func (sie *ShellIntegrationError) addConflictGuidance() {
	sie.WithSeverity(SeverityWarning)

	if len(sie.detectedConflicts) > 0 {
		sie.WithRecoveryAction("Other Perl version managers detected")
		sie.WithContext("detected_conflicts", sie.detectedConflicts)

		for _, conflict := range sie.detectedConflicts {
			if strings.Contains(conflict, "plenv") {
				sie.WithRecoveryAction("Consider removing plenv from PATH or disabling it")
				sie.WithRecoveryAction("Plenv shims may interfere with PVM")
			}
			if strings.Contains(conflict, "perlbrew") {
				sie.WithRecoveryAction("Consider deactivating perlbrew or removing from PATH")
				sie.WithRecoveryAction("Perlbrew may interfere with PVM version switching")
			}
		}

		sie.WithRecoveryAction("Suppress warnings: export PVM_SUPPRESS_WARNINGS=1")
	}
}

// addDirectoryGuidance provides guidance for missing directory issues
func (sie *ShellIntegrationError) addDirectoryGuidance() {
	sie.WithSeverity(SeverityError)

	sie.WithRecoveryAction("Required PVM directories are missing")
	sie.WithRecoveryAction("Run: pvm init (creates required directories)")
	sie.WithRecoveryAction("Auto-fix: pvm self doctor --fix")

	// Add specific directory information
	xdgBinHome := os.Getenv("XDG_BIN_HOME")
	if xdgBinHome == "" {
		xdgBinHome = filepath.Join(sie.homeDir, ".local", "bin")
	}
	sie.WithContext("required_xdg_bin_home", xdgBinHome)
}

// Helper methods

// getShellConfigFiles returns the configuration files for the given shell
func getShellConfigFiles(shell, homeDir string) []string {
	switch shell {
	case "bash":
		files := []string{".bashrc"}
		if _, err := os.Stat(filepath.Join(homeDir, ".bash_profile")); err == nil {
			files = append(files, ".bash_profile")
		}
		return files
	case "zsh":
		files := []string{".zshrc"}
		if _, err := os.Stat(filepath.Join(homeDir, ".zprofile")); err == nil {
			files = append(files, ".zprofile")
		}
		return files
	case "fish":
		return []string{".config/fish/config.fish"}
	default:
		return []string{}
	}
}

// getPrimaryConfigFile returns the primary configuration file for the shell
func (sie *ShellIntegrationError) getPrimaryConfigFile() string {
	switch sie.shellType {
	case "bash":
		return ".bashrc"
	case "zsh":
		return ".zshrc"
	case "fish":
		return ".config/fish/config.fish"
	default:
		return ""
	}
}

// detectVersionManagerConflicts detects other version managers in PATH
func detectVersionManagerConflicts(path string) []string {
	if path == "" {
		return []string{}
	}

	var conflicts []string
	pathDirs := strings.Split(path, string(os.PathListSeparator))

	for _, dir := range pathDirs {
		if strings.Contains(dir, "plenv") && strings.Contains(dir, "shims") {
			conflicts = append(conflicts, "plenv ("+dir+")")
		}
		if strings.Contains(dir, "perlbrew") && (strings.Contains(dir, "bin") || strings.Contains(dir, "perl")) {
			conflicts = append(conflicts, "perlbrew ("+dir+")")
		}
	}

	return conflicts
}

// fileExists checks if a file exists
func (sie *ShellIntegrationError) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// dirExists checks if a directory exists
func (sie *ShellIntegrationError) dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// Convenience constructors for common shell integration errors

// NewMissingShellIntegrationError creates an error for missing shell integration
func NewMissingShellIntegrationError(shell string) *ShellIntegrationError {
	return NewShellIntegrationError(
		ErrShellIntegrationMissing,
		"Shell integration not detected",
		nil,
		shell,
	)
}

// NewShellConfigMissingError creates an error for missing shell configuration
func NewShellConfigMissingError(shell string) *ShellIntegrationError {
	return NewShellIntegrationError(
		ErrShellConfigMissing,
		"Shell configuration doesn't contain PVM initialization",
		nil,
		shell,
	)
}

// NewShellEnvironmentError creates an error for incorrect shell environment
func NewShellEnvironmentError(shell string, issue string) *ShellIntegrationError {
	return NewShellIntegrationError(
		ErrShellEnvironmentIncorrect,
		fmt.Sprintf("Shell environment issue: %s", issue),
		nil,
		shell,
	)
}

// NewVersionManagerConflictError creates an error for version manager conflicts
func NewVersionManagerConflictError(conflicts []string) *ShellIntegrationError {
	return NewShellIntegrationError(
		ErrShellVersionManagerConflict,
		fmt.Sprintf("Version manager conflicts detected: %v", conflicts),
		nil,
		"", // Shell type not needed for this error
	)
}
