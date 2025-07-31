// ABOUTME: Shell integration for PVM
// ABOUTME: Provides functionality for shell initialization and version switching

package perl

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/xdg"
)

//go:embed shell_templates/pvm.bash
var bashTemplate string

//go:embed shell_templates/pvm.zsh  
var zshTemplate string

//go:embed shell_templates/pvm.fish
var fishTemplate string

// PowerShell and CMD templates (not embedded - not currently maintained)
var powershellTemplate = ""
var cmdTemplate = ""

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
}

// Returns whether the script is for a Windows shell
func (s ShellScriptData) IsWindows() bool {
	return runtime.GOOS == "windows"
}

// DetectShell attempts to detect the user's shell type
func DetectShell() (ShellType, error) {
	// On Windows, default to PowerShell or CMD
	if runtime.GOOS == "windows" {
		// Check if we're in PowerShell
		if os.Getenv("PSModulePath") != "" {
			return ShellPowerShell, nil
		}
		return ShellCmd, nil
	}

	// On Unix-like systems, check the SHELL environment variable
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		// Fallback to bash
		return ShellBash, nil
	}

	// Extract the shell name from the path
	shellName := filepath.Base(shellPath)

	// Determine the shell type based on name
	switch shellName {
	case "bash":
		return ShellBash, nil
	case "zsh":
		return ShellZsh, nil
	case "fish":
		return ShellFish, nil
	default:
		// Default to bash for unknown shells
		return ShellBash, nil
	}
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

// GenerateShellUse outputs shell commands to set up the environment for a specific Perl version
func GenerateShellUse(version string) error {
	// Allow "system" as a special case, otherwise validate the version
	if version != "system" {
		if err := ValidateVersion(version); err != nil {
			return err
		}
	}

	// Detect shell type from environment
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	// Extract shell name from full path
	shellName := filepath.Base(shell)

	// Generate shell-specific environment commands
	switch shellName {
	case "fish":
		fmt.Printf("set -gx PVM_PERL_VERSION %s\n", version)
		fmt.Printf("echo \"Using Perl %s\"\n", version)
	default: // bash, zsh, sh
		fmt.Printf("export PVM_PERL_VERSION=%s\n", version)
		fmt.Printf("echo \"Using Perl %s\"\n", version)
	}

	return nil
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
