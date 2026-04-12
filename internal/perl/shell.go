// ABOUTME: Shell integration for PVM
// ABOUTME: Provides functionality for shell initialization and version switching

package perl

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/fortune"
	"tamarou.com/pvm/internal/xdg"
)

//go:embed shell_templates/pvm.bash
var bashTemplate string

//go:embed shell_templates/pvm.zsh
var zshTemplate string

//go:embed shell_templates/pvm.fish
var fishTemplate string

//go:embed shell_templates/pvm.ps1
var powershellTemplate string

//go:embed shell_templates/pvm.cmd
var cmdTemplate string

// Shell-related error codes
const (
	ErrShellDetectionFailed   = "501" // Failed to detect user shell
	ErrShellInitFailed        = "502" // Failed to create shell initialization scripts
	ErrVersionSwitchFailed    = "503" // Failed to switch Perl version
	ErrPerlVersionFileFailed  = "504" // Failed to create .perl-version file
	ErrGlobalVersionSetFailed = "505" // Failed to set global Perl version
	ErrLocalVersionSetFailed  = "506" // Failed to set local Perl version
	ErrUseVersionFailed       = "507" // Failed to use specified Perl version
	ErrShellScriptFailed      = "508" // Failed to generate shell script
)

// ShellType represents the type of shell
type ShellType string

const (
	// Supported shell types
	ShellBash       ShellType = "bash"
	ShellZsh        ShellType = "zsh"
	ShellFish       ShellType = "fish"
	ShellPowerShell ShellType = "powershell"
	ShellCmd        ShellType = "cmd"
	ShellUnknown    ShellType = "unknown"
)

// ShellScriptData contains data for shell script templates
type ShellScriptData struct {
	// The PVM executable path
	PVMPath string

	// Shims directory
	ShimsDir string

	// User configuration directory
	ConfigDir string

	// Function prefix, used to avoid naming collisions
	FunctionPrefix string

	// Whether the shell supports advanced features
	SupportsAdvanced bool

	// Conflict warnings for other version managers
	ConflictWarnings string

	// Fortune quote for initialization Easter egg
	FortuneQuote string
}

// Returns whether the script is for a Windows shell
func (s ShellScriptData) IsWindows() bool {
	return runtime.GOOS == "windows"
}

// DetectShell attempts to detect the user's shell type using environment-first logic.
// PSModulePath is checked first on any OS (PowerShell 7 runs on Linux/macOS too),
// then SHELL is checked, with an OS-based fallback last.
func DetectShell() (ShellType, error) {
	// PVM_SHELL is set by the shell integration templates, which know the
	// shell they run in. It is the most authoritative signal because $SHELL
	// reflects the user's login shell, not the shell currently running pvm.
	switch os.Getenv("PVM_SHELL") {
	case "fish":
		return ShellFish, nil
	case "zsh":
		return ShellZsh, nil
	case "bash":
		return ShellBash, nil
	case "powershell":
		return ShellPowerShell, nil
	case "cmd":
		return ShellCmd, nil
	}

	// Check PSModulePath - PowerShell 7 (pwsh) runs on Linux/macOS
	if os.Getenv("PSModulePath") != "" {
		return ShellPowerShell, nil
	}

	// Fall back to $SHELL (set by the login shell on Unix).
	shellPath := os.Getenv("SHELL")
	if shellPath != "" {
		shellName := filepath.Base(shellPath)
		switch shellName {
		case "bash":
			return ShellBash, nil
		case "zsh":
			return ShellZsh, nil
		case "fish":
			return ShellFish, nil
		}
	}

	// OS-based fallback: CMD on Windows, bash on Unix
	if runtime.GOOS == "windows" {
		return ShellCmd, nil
	}
	return ShellBash, nil
}

// CreateShellInitScripts generates shell initialization scripts for all supported shells
func CreateShellInitScripts() error {
	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return errors.NewSystemError("001",
			"Failed to determine XDG directories", err)
	}

	// Ensure directories exist
	err = dirs.EnsureDirs()
	if err != nil {
		return errors.NewSystemError("002",
			"Failed to create required directories", err)
	}

	// Get path to PVM executable
	pvmPath, err := executablePath()
	if err != nil {
		return errors.NewVersionError(
			ErrShellInitFailed,
			"Failed to determine PVM executable path",
			err)
	}

	// Create the shell integration directory
	shellDir := filepath.Join(dirs.DataDir, "shell")
	err = os.MkdirAll(shellDir, 0755)
	if err != nil {
		return errors.NewVersionError(
			ErrShellInitFailed,
			"Failed to create shell integration directory",
			err)
	}

	// Bootstrap from plenv if available
	err = bootstrapFromPlenv()
	if err != nil {
		// Log but don't fail - plenv bootstrap is optional
		// TODO: Consider adding debug logging here
	}

	// Initialize template data
	data := ShellScriptData{
		PVMPath:          pvmPath,
		ShimsDir:         dirs.ShimsDir,
		ConfigDir:        dirs.ConfigDir,
		FunctionPrefix:   "pvm_",
		SupportsAdvanced: true,
		ConflictWarnings: generateConflictWarnings(),
		FortuneQuote:     fortune.GetRandomQuote(),
	}

	// Generate scripts for each supported shell
	shells := []ShellType{ShellBash, ShellZsh, ShellFish}
	if runtime.GOOS == "windows" {
		shells = append(shells, ShellPowerShell, ShellCmd)
	}

	for _, shell := range shells {
		// Generate the script for this shell
		script, err := GenerateShellScript(shell, data)
		if err != nil {
			return err
		}

		// Determine the file extension
		var extension string
		switch shell {
		case ShellBash:
			extension = ".bash"
		case ShellZsh:
			extension = ".zsh"
		case ShellFish:
			extension = ".fish"
		case ShellPowerShell:
			extension = ".ps1"
		case ShellCmd:
			extension = ".cmd"
		default:
			extension = ".sh"
		}

		// Write the script to a file
		scriptPath := filepath.Join(shellDir, "pvm"+extension)
		err = os.WriteFile(scriptPath, []byte(script), 0644)
		if err != nil {
			return errors.NewVersionError(
				ErrShellInitFailed,
				fmt.Sprintf("Failed to write shell script for %s", shell),
				err).
				WithLocation(scriptPath)
		}
	}

	return nil
}

// GenerateShellScript creates initialization script for a specific shell
func GenerateShellScript(shellType ShellType, data ShellScriptData) (string, error) {
	var tmplText string

	// Select the appropriate template based on shell type
	switch shellType {
	case ShellBash:
		tmplText = bashTemplate
	case ShellZsh:
		tmplText = zshTemplate
	case ShellFish:
		tmplText = fishTemplate
	case ShellPowerShell:
		tmplText = powershellTemplate
	case ShellCmd:
		tmplText = cmdTemplate
	default:
		return "", errors.NewVersionError(
			ErrShellScriptFailed,
			fmt.Sprintf("Unsupported shell type: %s", shellType),
			nil)
	}

	// Parse template
	tmpl, err := template.New("shell").Parse(tmplText)
	if err != nil {
		return "", errors.NewVersionError(
			ErrShellScriptFailed,
			fmt.Sprintf("Failed to parse template for %s", shellType),
			err)
	}

	// Process template
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", errors.NewVersionError(
			ErrShellScriptFailed,
			fmt.Sprintf("Failed to execute template for %s", shellType),
			err)
	}

	return buf.String(), nil
}

// GetShellInitCommand returns a command to initialize PVM in the current shell
func GetShellInitCommand(shellType ShellType) (string, error) {
	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return "", errors.NewSystemError("001",
			"Failed to determine XDG directories", err)
	}

	// Shell integration path
	shellDir := filepath.Join(dirs.DataDir, "shell")

	// Determine the file extension and init command
	var scriptPath string
	var initCommand string

	switch shellType {
	case ShellBash:
		scriptPath = filepath.Join(shellDir, "pvm.bash")
		initCommand = fmt.Sprintf("source %s", scriptPath)
	case ShellZsh:
		scriptPath = filepath.Join(shellDir, "pvm.zsh")
		initCommand = fmt.Sprintf("source %s", scriptPath)
	case ShellFish:
		scriptPath = filepath.Join(shellDir, "pvm.fish")
		initCommand = fmt.Sprintf("source %s", scriptPath)
	case ShellPowerShell:
		scriptPath = filepath.Join(shellDir, "pvm.ps1")
		initCommand = fmt.Sprintf(". %s", scriptPath)
	case ShellCmd:
		scriptPath = filepath.Join(shellDir, "pvm.cmd")
		initCommand = scriptPath
	default:
		return "", errors.NewVersionError(
			ErrShellScriptFailed,
			fmt.Sprintf("Unsupported shell type: %s", shellType),
			nil)
	}

	return initCommand, nil
}

// GetShellCompletionCommand returns a command to enable shell completion for PVM
func GetShellCompletionCommand(shellType ShellType) (string, error) {
	// Implement shell completion commands for different shells
	switch shellType {
	case ShellBash:
		return "complete -C 'pvm completion bash' pvm", nil
	case ShellZsh:
		return "compdef _pvm pvm", nil
	case ShellFish:
		return "complete -c pvm -f -a '(pvm completion fish)'", nil
	case ShellPowerShell:
		return "Register-ArgumentCompleter -Native -CommandName pvm -ScriptBlock (pvm completion powershell | Out-String | Invoke-Expression)", nil
	default:
		return "", errors.NewVersionError(
			ErrShellScriptFailed,
			fmt.Sprintf("Shell completion not supported for %s", shellType),
			nil)
	}
}

// SwitchVersion sets the Perl version for the given scope
func SwitchVersion(version string, scope string) error {
	// Handle different scopes
	switch scope {
	case "global":
		return SetGlobalVersion(version)
	case "local":
		return SetLocalVersion(version)
	case "shell":
		return UseVersion(version)
	default:
		return errors.NewVersionError(
			ErrVersionSwitchFailed,
			fmt.Sprintf("Invalid scope: %s (must be global, local, or shell)", scope),
			nil)
	}
}

// SetGlobalVersion sets the global default Perl version
func SetGlobalVersion(version string) error {
	// Allow "system" as a special case, otherwise validate the version
	if version != "system" {
		if err := ValidateVersion(version); err != nil {
			return err
		}
	}

	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return errors.NewSystemError("001",
			"Failed to determine XDG directories", err)
	}

	// Get configuration file path
	configPath := dirs.GetConfigFilePath()

	// Create the configuration directory if it doesn't exist
	err = os.MkdirAll(filepath.Dir(configPath), 0755)
	if err != nil {
		return errors.NewVersionError(
			ErrGlobalVersionSetFailed,
			"Failed to create configuration directory",
			err)
	}

	// Create/update configuration file with the version
	configContent := fmt.Sprintf("[pvm]\ndefault_perl = \"%s\"\n", version)

	// Write to file
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		return errors.NewVersionError(
			ErrGlobalVersionSetFailed,
			"Failed to write configuration file",
			err).
			WithLocation(configPath)
	}

	return nil
}

// UnsetGlobalVersion sets the global version to "system" to fall back to system Perl
func UnsetGlobalVersion() error {
	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return errors.NewSystemError("001",
			"Failed to determine XDG directories", err)
	}

	// Get configuration file path
	configPath := dirs.GetConfigFilePath()

	// Create the configuration directory if it doesn't exist
	err = os.MkdirAll(filepath.Dir(configPath), 0755)
	if err != nil {
		return errors.NewVersionError(
			ErrGlobalVersionSetFailed,
			"Failed to create configuration directory",
			err)
	}

	// Create/update configuration file with system setting
	configContent := "[pvm]\ndefault_perl = \"system\"\n"

	// Write to file
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		return errors.NewVersionError(
			ErrGlobalVersionSetFailed,
			"Failed to write configuration file",
			err).
			WithLocation(configPath)
	}

	return nil
}

// SetLocalVersion sets the local Perl version for a project
func SetLocalVersion(version string) error {
	// Allow "system" as a special case, otherwise validate the version
	if version != "system" {
		if err := ValidateVersion(version); err != nil {
			return err
		}
	}

	// Get current directory
	dir, err := os.Getwd()
	if err != nil {
		return errors.NewVersionError(
			ErrLocalVersionSetFailed,
			"Failed to determine current directory",
			err)
	}

	// Create .perl-version file
	versionFilePath := filepath.Join(dir, ".perl-version")
	err = os.WriteFile(versionFilePath, []byte(version), 0644)
	if err != nil {
		return errors.NewVersionError(
			ErrLocalVersionSetFailed,
			"Failed to write .perl-version file",
			err).
			WithLocation(versionFilePath)
	}

	return nil
}

// UnsetLocalVersion removes the local Perl version for the current directory
func UnsetLocalVersion() error {
	// Get current directory
	dir, err := os.Getwd()
	if err != nil {
		return errors.NewVersionError(
			ErrLocalVersionSetFailed,
			"Failed to get current directory",
			err)
	}

	// Define the .perl-version file path
	versionFile := filepath.Join(dir, ".perl-version")

	// Check if the file exists
	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		// File doesn't exist, nothing to unset
		return nil
	}

	// Remove the .perl-version file
	err = os.Remove(versionFile)
	if err != nil {
		return errors.NewVersionError(
			ErrLocalVersionSetFailed,
			"Failed to remove .perl-version file",
			err).
			WithLocation(versionFile)
	}

	return nil
}

// UseVersion sets the Perl version for the current shell session
func UseVersion(version string) error {
	// Allow "system" as a special case, otherwise validate the version
	if version != "system" {
		if err := ValidateVersion(version); err != nil {
			return err
		}
	}

	// Since we can't directly affect the parent shell's environment,
	// we need to output shell commands that will be evaluated by the parent shell.
	// This is typically handled by the shell integration script.

	// For shell script integration, we'll set an environment variable that the shell
	// integration can use to determine the current version
	fmt.Printf("export PVM_CURRENT_PERL=%s\n", version)

	return nil
}

// sanitizeLibraryName validates that library names are safe and don't contain malicious content
func sanitizeLibraryName(library string) error {
	if library == "" {
		return nil // Empty library is valid (uses default)
	}

	// Reject names that are only whitespace
	if strings.TrimSpace(library) == "" {
		return errors.NewVersionError(
			ErrVersionSwitchFailed,
			"library name cannot be empty or whitespace-only",
			nil)
	}

	// Check for path traversal attempts
	if strings.Contains(library, "..") || strings.Contains(library, "/") || strings.Contains(library, "\\") {
		return errors.NewVersionError(
			ErrVersionSwitchFailed,
			"library name contains invalid path characters",
			nil)
	}

	// Prevent absolute paths
	if filepath.IsAbs(library) {
		return errors.NewVersionError(
			ErrVersionSwitchFailed,
			"library name cannot be an absolute path",
			nil)
	}

	// Limit length to prevent DoS attacks
	if len(library) > 64 {
		return errors.NewVersionError(
			ErrVersionSwitchFailed,
			"library name is too long (maximum 64 characters)",
			nil)
	}

	// Only allow alphanumeric characters, dash, and underscore
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", library)
	if !matched {
		return errors.NewVersionError(
			ErrVersionSwitchFailed,
			"library name can only contain alphanumeric characters, dash, and underscore",
			nil)
	}

	return nil
}

// escapeShellArg properly escapes a string for safe use in shell commands
func escapeShellArg(arg string) string {
	// Use single quotes and escape any single quotes within
	return "'" + strings.ReplaceAll(arg, "'", "'\"'\"'") + "'"
}

// validateLibraryEnvironmentImpl is the actual implementation
func validateLibraryEnvironmentImpl(library string) error {
	if library == "" {
		return nil // Empty library is valid (uses default)
	}

	// First sanitize the library name
	if err := sanitizeLibraryName(library); err != nil {
		return err
	}

	dirs, err := xdg.GetDirs()
	if err != nil {
		return errors.NewVersionError(
			ErrVersionSwitchFailed,
			"Failed to determine XDG directories",
			err)
	}

	// Build the environment directory path and verify it's within bounds
	envDir := filepath.Clean(filepath.Join(dirs.DataDir, "environments", library))
	expectedPrefix := filepath.Clean(filepath.Join(dirs.DataDir, "environments")) + string(os.PathSeparator)

	// Double-check that the resolved path is within the expected directory
	if !strings.HasPrefix(envDir+string(os.PathSeparator), expectedPrefix) {
		return errors.NewVersionError(
			ErrVersionSwitchFailed,
			"library path resolves outside allowed directory",
			nil)
	}

	// Check if library environment exists
	if _, err := os.Stat(envDir); os.IsNotExist(err) {
		return errors.NewVersionError(
			ErrVersionSwitchFailed,
			fmt.Sprintf("Library environment '%s' does not exist. Library environments provide isolated module installations. Create with: pvm pvx --name %s --isolation local", library, library),
			nil)
	}

	return nil
}

// GenerateShellUse outputs shell commands to set up the environment for a specific
// Perl version and optional library. When version is "system", PVM_PERL_VERSION
// (and PVM_PERL_LIBRARY when no library is given) are unset so the resolver
// falls back to the system Perl.
//
// The output is eval'd by the shell integration templates. Callers are
// responsible for refreshing PATH afterward (the templates call
// _pvm_update_perl_path) — this function does not touch PATH.
func GenerateShellUse(version string, library string) error {
	isSystem := version == "system"

	if !isSystem {
		if err := ValidateVersion(version); err != nil {
			return err
		}
	}

	if err := ValidateLibraryEnvironment(library); err != nil {
		return err
	}

	shellType, err := DetectShell()
	if err != nil {
		shellType = ShellBash
	}

	displayVersion := version
	if library != "" {
		displayVersion = fmt.Sprintf("%s@%s", version, library)
	}

	// For the "system" fallback we always clear both env vars: the library
	// name in version@library is accepted (and validated) only so the
	// confirmation message can reference it, not to activate the library.
	exportLibrary := !isSystem && library != ""

	switch shellType {
	case ShellFish:
		if isSystem {
			fmt.Println("set -e PVM_PERL_VERSION 2>/dev/null; or true")
		} else {
			fmt.Printf("set -gx PVM_PERL_VERSION %s\n", escapeShellArg(version))
		}
		if exportLibrary {
			fmt.Printf("set -gx PVM_PERL_LIBRARY %s\n", escapeShellArg(library))
			fmt.Printf("set -gx PVM_PERL_VERSION_FULL %s\n", escapeShellArg(displayVersion))
		} else {
			fmt.Println("set -e PVM_PERL_LIBRARY 2>/dev/null; or true")
			if isSystem {
				fmt.Println("set -e PVM_PERL_VERSION_FULL 2>/dev/null; or true")
			} else {
				fmt.Printf("set -gx PVM_PERL_VERSION_FULL %s\n", escapeShellArg(version))
			}
		}
		fmt.Printf("echo %s\n", escapeShellArg(useMessage(isSystem, library, displayVersion)))
	case ShellPowerShell:
		if isSystem {
			fmt.Println("Remove-Item Env:PVM_PERL_VERSION -ErrorAction SilentlyContinue")
		} else {
			fmt.Printf("$env:PVM_PERL_VERSION = %s\n", escapePowerShellArg(version))
		}
		if exportLibrary {
			fmt.Printf("$env:PVM_PERL_LIBRARY = %s\n", escapePowerShellArg(library))
			fmt.Printf("$env:PVM_PERL_VERSION_FULL = %s\n", escapePowerShellArg(displayVersion))
		} else {
			fmt.Println("Remove-Item Env:PVM_PERL_LIBRARY -ErrorAction SilentlyContinue")
			if isSystem {
				fmt.Println("Remove-Item Env:PVM_PERL_VERSION_FULL -ErrorAction SilentlyContinue")
			} else {
				fmt.Printf("$env:PVM_PERL_VERSION_FULL = %s\n", escapePowerShellArg(version))
			}
		}
		fmt.Printf("Write-Host %s\n", escapePowerShellArg(useMessage(isSystem, library, displayVersion)))
	default: // bash, zsh, sh, cmd
		if isSystem {
			fmt.Println("unset PVM_PERL_VERSION")
		} else {
			fmt.Printf("export PVM_PERL_VERSION=%s\n", escapeShellArg(version))
		}
		if exportLibrary {
			fmt.Printf("export PVM_PERL_LIBRARY=%s\n", escapeShellArg(library))
			fmt.Printf("export PVM_PERL_VERSION_FULL=%s\n", escapeShellArg(displayVersion))
		} else {
			fmt.Println("unset PVM_PERL_LIBRARY")
			if isSystem {
				fmt.Println("unset PVM_PERL_VERSION_FULL")
			} else {
				fmt.Printf("export PVM_PERL_VERSION_FULL=%s\n", escapeShellArg(version))
			}
		}
		fmt.Printf("echo %s\n", escapeShellArg(useMessage(isSystem, library, displayVersion)))
	}

	return nil
}

// useMessage returns the confirmation text printed after switching.
func useMessage(isSystem bool, library, displayVersion string) string {
	if isSystem {
		if library != "" {
			return fmt.Sprintf("Using system Perl with library '%s'", library)
		}
		return "Using system Perl"
	}
	return "Using Perl " + displayVersion
}

// escapePowerShellArg wraps a value in PowerShell single quotes, doubling any
// embedded single quotes per PowerShell literal-string rules.
func escapePowerShellArg(arg string) string {
	return "'" + strings.ReplaceAll(arg, "'", "''") + "'"
}

// validateVersionImpl is the actual implementation
func validateVersionImpl(version string) error {
	// If version is empty, return error
	if version == "" {
		return errors.NewVersionError(
			ErrVersionSwitchFailed,
			"Version cannot be empty",
			nil)
	}

	// Check if the version is installed
	versions, err := GetInstalledVersions()
	if err != nil {
		return err
	}

	// Check if the version is installed or a valid alias
	for _, v := range versions {
		if v.Version == version {
			return nil
		}
	}

	// If we're here, version is not installed
	return errors.NewVersionError(
		ErrVersionSwitchFailed,
		fmt.Sprintf("Version '%s' is not installed", version),
		nil)
}

// ValidateVersion is a variable pointing to validateVersionImpl for easier mocking in tests
var ValidateVersion = validateVersionImpl

// ValidateLibraryEnvironment is a variable pointing to validateLibraryEnvironmentImpl for easier mocking in tests
var ValidateLibraryEnvironment = validateLibraryEnvironmentImpl

// GetCurrentShellScript generates shell script to be evaluated for the command:
// eval "$(pvm shell)"
func GetCurrentShellScript(shellType ShellType) (string, error) {
	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return "", errors.NewSystemError("001",
			"Failed to determine XDG directories", err)
	}

	// Get path to PVM executable
	pvmPath, err := executablePath()
	if err != nil {
		return "", errors.NewVersionError(
			ErrShellInitFailed,
			"Failed to determine PVM executable path",
			err)
	}

	// Initialize template data
	data := ShellScriptData{
		PVMPath:          pvmPath,
		ShimsDir:         dirs.ShimsDir,
		ConfigDir:        dirs.ConfigDir,
		FunctionPrefix:   "pvm_",
		SupportsAdvanced: true,
		FortuneQuote:     fortune.GetRandomQuote(),
	}

	// Generate the script for the detected shell
	return GenerateShellScript(shellType, data)
}

// CheckShellInit checks if the shell is properly initialized with PVM
func CheckShellInit(shellType ShellType) (bool, string, error) {
	// Implement shell-specific checks
	switch shellType {
	case ShellBash:
		// Check for initialization in .bashrc
		home, err := os.UserHomeDir()
		if err != nil {
			return false, "", errors.NewSystemError("001",
				"Failed to determine user home directory", err)
		}

		bashrcPath := filepath.Join(home, ".bashrc")
		content, err := os.ReadFile(bashrcPath)
		if err != nil {
			return false, "", errors.NewVersionError(
				ErrShellDetectionFailed,
				"Failed to read .bashrc",
				err).
				WithLocation(bashrcPath)
		}

		return strings.Contains(string(content), "pvm init"),
			"Add 'eval \"$(pvm init)\"' to your ~/.bashrc file", nil

	case ShellZsh:
		// Check for initialization in .zshrc
		home, err := os.UserHomeDir()
		if err != nil {
			return false, "", errors.NewSystemError("001",
				"Failed to determine user home directory", err)
		}

		zshrcPath := filepath.Join(home, ".zshrc")
		content, err := os.ReadFile(zshrcPath)
		if err != nil {
			return false, "", errors.NewVersionError(
				ErrShellDetectionFailed,
				"Failed to read .zshrc",
				err).
				WithLocation(zshrcPath)
		}

		return strings.Contains(string(content), "pvm init"),
			"Add 'eval \"$(pvm init)\"' to your ~/.zshrc file", nil

	case ShellFish:
		// Check for initialization in fish config
		home, err := os.UserHomeDir()
		if err != nil {
			return false, "", errors.NewSystemError("001",
				"Failed to determine user home directory", err)
		}

		fishConfigPath := filepath.Join(home, ".config", "fish", "config.fish")
		content, err := os.ReadFile(fishConfigPath)
		if err != nil {
			return false, "", errors.NewVersionError(
				ErrShellDetectionFailed,
				"Failed to read fish config",
				err).
				WithLocation(fishConfigPath)
		}

		return strings.Contains(string(content), "pvm init"),
			"Add 'pvm init | source' to your ~/.config/fish/config.fish file", nil

	case ShellPowerShell:
		// Check for initialization in PowerShell profile
		// This is more complex on Windows, so we'll just return a recommendation
		return false, "Add 'pvm init | Invoke-Expression' to your PowerShell profile", nil

	case ShellCmd:
		// CMD has no persistent initialization file like .bashrc.
		// The standard approach is the AutoRun registry key, but checking
		// that is complex and fragile. Return a recommendation instead.
		return false, "Run 'pvm init' in each new CMD session", nil

	default:
		return false, "", errors.NewVersionError(
			ErrShellDetectionFailed,
			fmt.Sprintf("Shell detection not supported for %s", shellType),
			nil)
	}
}

// GetShellInitInstructions returns instructions for initializing PVM in the given shell
func GetShellInitInstructions(shellType ShellType) string {
	switch shellType {
	case ShellBash:
		return "Add the following to your ~/.bashrc file:\n" +
			"eval \"$(pvm init)\"\n"
	case ShellZsh:
		return "Add the following to your ~/.zshrc file:\n" +
			"eval \"$(pvm init)\"\n"
	case ShellFish:
		return "Add the following to your ~/.config/fish/config.fish file:\n" +
			"pvm init | source\n"
	case ShellPowerShell:
		return "Add the following to your PowerShell profile:\n" +
			"pvm init | Invoke-Expression\n"
	case ShellCmd:
		return "There's no persistent initialization for CMD.\n" +
			"Run 'pvm init' in each new CMD session.\n"
	default:
		return "Shell initialization not supported for this shell type."
	}
}

// These functions are already defined in legacy.go
// Using them directly from there

// detectVersionManagerConflicts checks for other version managers in PATH
// that might conflict with PVM and returns a list of detected conflicts
func detectVersionManagerConflicts() []string {
	var conflicts []string

	// Split PATH into directories
	pathDirs := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))

	// Check for plenv shims directory
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

// generateConflictWarnings creates shell script code to warn about version manager conflicts
func generateConflictWarnings() string {
	conflicts := detectVersionManagerConflicts()
	if len(conflicts) == 0 {
		return ""
	}

	var warnings strings.Builder
	warnings.WriteString("    # Check for potential version manager conflicts\n")
	warnings.WriteString("    if [ \"$PVM_SUPPRESS_WARNINGS\" != \"1\" ]; then\n")
	warnings.WriteString("        echo \"⚠️  PVM detected other Perl version managers in PATH:\"\n")

	for _, conflict := range conflicts {
		warnings.WriteString(fmt.Sprintf("        echo \"   - %s\"\n", conflict))
	}

	warnings.WriteString("        echo \"   PVM shims should take precedence, but conflicts may occur.\"\n")
	warnings.WriteString("        echo \"   To suppress this warning: export PVM_SUPPRESS_WARNINGS=1\"\n")
	warnings.WriteString("        echo\n")
	warnings.WriteString("    fi\n")

	return warnings.String()
}

// bootstrapFromPlenv attempts to bootstrap PVM from existing plenv installations
func bootstrapFromPlenv() error {
	// Check if plenv is available
	if !isPlenvAvailable() {
		return nil // Not an error - plenv is optional
	}

	// Get all plenv-managed Perl versions
	plenvVersions, err := getPlenvVersions()
	if err != nil {
		return err // Return error to indicate bootstrap failed
	}

	if len(plenvVersions) == 0 {
		return nil // No versions to import
	}

	// Import each plenv version into PVM registry
	for _, pv := range plenvVersions {
		// Skip system version - it should be handled separately
		if pv.IsSystem {
			continue
		}

		// Create a SystemPerl for this plenv version
		perl, err := extractPerlInfo(pv.Path, false)
		if err != nil {
			// Skip versions that can't be processed
			continue
		}

		// Register this version in PVM
		registryEntry := VersionInfo{
			Version:     perl.Version,
			InstallPath: filepath.Dir(filepath.Dir(pv.Path)), // Go from /path/to/version/bin/perl to /path/to/version
			InstallTime: time.Now(),
			Source:      "plenv",
		}

		err = RegisterVersion(registryEntry)
		if err != nil {
			// Continue with other versions if one fails
			continue
		}
	}

	return nil
}
