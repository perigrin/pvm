// ABOUTME: Shell integration for PVM
// ABOUTME: Provides functionality for shell initialization and version switching

package perl

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/xdg"
)

// Shell-related error codes
const (
	ErrShellDetectionFailed    = "501" // Failed to detect user shell
	ErrShellInitFailed         = "502" // Failed to create shell initialization scripts
	ErrVersionSwitchFailed     = "503" // Failed to switch Perl version
	ErrPerlVersionFileFailed   = "504" // Failed to create .perl-version file
	ErrGlobalVersionSetFailed  = "505" // Failed to set global Perl version
	ErrLocalVersionSetFailed   = "506" // Failed to set local Perl version
	ErrUseVersionFailed        = "507" // Failed to use specified Perl version
	ErrShellScriptFailed       = "508" // Failed to generate shell script
)

// ShellType represents the type of shell
type ShellType string

const (
	// Supported shell types
	ShellBash      ShellType = "bash"
	ShellZsh       ShellType = "zsh"
	ShellFish      ShellType = "fish"
	ShellPowerShell ShellType = "powershell"
	ShellCmd       ShellType = "cmd"
	ShellUnknown   ShellType = "unknown"
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

	// Initialize template data
	data := ShellScriptData{
		PVMPath:          pvmPath,
		ShimsDir:         dirs.ShimsDir,
		ConfigDir:        dirs.ConfigDir,
		FunctionPrefix:   "pvm_",
		SupportsAdvanced: true,
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
	case ShellBash, ShellZsh:
		tmplText = bashZshTemplate
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
	// Check if the version is valid
	if err := ValidateVersion(version); err != nil {
		return err
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

// SetLocalVersion sets the local Perl version for a project
func SetLocalVersion(version string) error {
	// Check if the version is valid
	if err := ValidateVersion(version); err != nil {
		return err
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

// UseVersion sets the Perl version for the current shell session
func UseVersion(version string) error {
	// Check if the version is valid
	if err := ValidateVersion(version); err != nil {
		return err
	}

	// Since we can't directly affect the parent shell's environment,
	// we need to output shell commands that will be evaluated by the parent shell.
	// This is typically handled by the shell integration script.

	// For shell script integration, we'll set an environment variable that the shell
	// integration can use to determine the current version
	fmt.Printf("export PVM_CURRENT_PERL=%s\n", version)

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

// Shell script templates for different shell types

// Bash/Zsh shell script template
const bashZshTemplate = `#!/usr/bin/env sh
# PVM Shell Integration for Bash/Zsh
# Generated by PVM - DO NOT EDIT

# Path to PVM executable
PVM_EXEC="{{ .PVMPath }}"

# Shims directory
PVM_SHIMS_DIR="{{ .ShimsDir }}"

# Function to initialize PVM
{{ .FunctionPrefix }}init() {
    # Add shims directory to PATH if not already there
    case ":${PATH}:" in
        *":${PVM_SHIMS_DIR}:"*) ;;
        *) export PATH="${PVM_SHIMS_DIR}:${PATH}" ;;
    esac

    # Define shell functions for version switching
    {{ .FunctionPrefix }}use() {
        local version="$1"
        if [ -z "$version" ]; then
            echo "Usage: pvm use <version>"
            return 1
        fi

        # Set version for current shell
        export PVM_PERL_VERSION="$version"
        echo "Using Perl $version"
    }

    {{ .FunctionPrefix }}local() {
        "$PVM_EXEC" local "$@"
    }

    {{ .FunctionPrefix }}global() {
        "$PVM_EXEC" global "$@"
    }

    # Create shell aliases
    if [ -n "$ZSH_VERSION" ]; then
        # Zsh aliases
        alias pvm-use="{{ .FunctionPrefix }}use"
        alias pvm-local="{{ .FunctionPrefix }}local"
        alias pvm-global="{{ .FunctionPrefix }}global"
    else
        # Bash aliases
        alias pvm-use="{{ .FunctionPrefix }}use"
        alias pvm-local="{{ .FunctionPrefix }}local"
        alias pvm-global="{{ .FunctionPrefix }}global"
    fi

    # Function to run custom commands upon directory change
    {{ .FunctionPrefix }}cd() {
        \cd "$@" || return $?
        
        # Check for .perl-version file in current directory
        if [ -f .perl-version ]; then
            local version=$(cat .perl-version)
            [ -n "$version" ] && {{ .FunctionPrefix }}use "$version"
        fi
    }

    # Set up cd override (if supported)
    if [ -n "$ZSH_VERSION" ]; then
        # For Zsh, use chpwd hook
        {{ .FunctionPrefix }}chpwd() {
            if [ -f .perl-version ]; then
                local version=$(cat .perl-version)
                [ -n "$version" ] && {{ .FunctionPrefix }}use "$version"
            fi
        }
        
        autoload -Uz add-zsh-hook
        add-zsh-hook chpwd {{ .FunctionPrefix }}chpwd
    else
        # For Bash, override cd
        alias cd="{{ .FunctionPrefix }}cd"
    fi
}

# Initialize PVM
{{ .FunctionPrefix }}init

# Output message
echo "PVM environment initialized"
`

// Fish shell script template
const fishTemplate = `#!/usr/bin/env fish
# PVM Shell Integration for Fish
# Generated by PVM - DO NOT EDIT

# Path to PVM executable
set PVM_EXEC "{{ .PVMPath }}"

# Shims directory
set PVM_SHIMS_DIR "{{ .ShimsDir }}"

# Function to initialize PVM
function {{ .FunctionPrefix }}init
    # Add shims directory to PATH if not already there
    if not contains $PVM_SHIMS_DIR $PATH
        set -gx PATH $PVM_SHIMS_DIR $PATH
    end

    # Define shell functions for version switching
    function {{ .FunctionPrefix }}use
        set version $argv[1]
        if test -z "$version"
            echo "Usage: pvm use <version>"
            return 1
        end

        # Set version for current shell
        set -gx PVM_PERL_VERSION $version
        echo "Using Perl $version"
    end

    # Create shell aliases
    alias pvm-use="{{ .FunctionPrefix }}use"
    alias pvm-local="$PVM_EXEC local"
    alias pvm-global="$PVM_EXEC global"

    # Function to run when directory changes
    function {{ .FunctionPrefix }}on_pwd_change --on-variable PWD
        if test -f "$PWD/.perl-version"
            set version (cat "$PWD/.perl-version")
            if test -n "$version"
                {{ .FunctionPrefix }}use $version
            end
        end
    end

    # Check current directory immediately
    {{ .FunctionPrefix }}on_pwd_change
end

# Initialize PVM
{{ .FunctionPrefix }}init

# Output message
echo "PVM environment initialized"
`

// PowerShell script template
const powershellTemplate = `# PVM Shell Integration for PowerShell
# Generated by PVM - DO NOT EDIT

# Path to PVM executable
$PVM_EXEC = "{{ .PVMPath }}"

# Shims directory
$PVM_SHIMS_DIR = "{{ .ShimsDir }}"

# Function to initialize PVM
function {{ .FunctionPrefix }}Init {
    # Add shims directory to PATH if not already there
    if ($env:PATH -split ';' -notcontains $PVM_SHIMS_DIR) {
        $env:PATH = "$PVM_SHIMS_DIR;$env:PATH"
    }

    # Define shell functions for version switching
    function global:Use-PerlVersion {
        param(
            [Parameter(Mandatory=$true)]
            [string]$Version
        )

        # Set version for current shell
        $env:PVM_PERL_VERSION = $Version
        Write-Host "Using Perl $Version"
    }

    # Create shell aliases
    Set-Alias -Name pvm-use -Value Use-PerlVersion -Scope Global
    Set-Alias -Name pvm-local -Value { & $PVM_EXEC local $args } -Scope Global
    Set-Alias -Name pvm-global -Value { & $PVM_EXEC global $args } -Scope Global

    # Function to check for .perl-version when directory changes
    function global:{{ .FunctionPrefix }}OnLocationChanged {
        if (Test-Path .perl-version) {
            $version = Get-Content .perl-version -Raw
            $version = $version.Trim()
            if ($version) {
                Use-PerlVersion $version
            }
        }
    }

    # Set up directory change hook
    $ExecutionContext.InvokeCommand.LocationChangedAction = { {{ .FunctionPrefix }}OnLocationChanged }

    # Check current directory immediately
    {{ .FunctionPrefix }}OnLocationChanged
}

# Initialize PVM
{{ .FunctionPrefix }}Init

# Output message
Write-Host "PVM environment initialized"
`

// CMD script template
const cmdTemplate = `@echo off
:: PVM Shell Integration for CMD
:: Generated by PVM - DO NOT EDIT

:: Path to PVM executable
set PVM_EXEC={{ .PVMPath }}

:: Shims directory
set PVM_SHIMS_DIR={{ .ShimsDir }}

:: Add shims directory to PATH if not already there
set PATH=%PVM_SHIMS_DIR%;%PATH%

:: Message
echo PVM environment initialized
`