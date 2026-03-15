// ABOUTME: Shell integration for PVX shim PATH management
// ABOUTME: Generates shell-specific activation scripts and manages shell configuration

package shim

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

const (
	bashActivationTemplate = `# PVX Global Tool Shims
# This section was automatically added by PVX
# Do not edit manually - use 'pvm tool shim' commands instead

# Add PVX shim directory to PATH
if [ -d "{{.ShimDir}}" ]; then
    case ":$PATH:" in
        *":{{.ShimDir}}:"*) ;;
        *) export PATH="{{.PathString}}" ;;
    esac
fi

# PVX Global Tool Shims End
`

	zshActivationTemplate = `# PVX Global Tool Shims
# This section was automatically added by PVX
# Do not edit manually - use 'pvm tool shim' commands instead

# Add PVX shim directory to PATH
if [ -d "{{.ShimDir}}" ]; then
    case ":$PATH:" in
        *":{{.ShimDir}}:"*) ;;
        *) export PATH="{{.PathString}}" ;;
    esac
fi

# PVX Global Tool Shims End
`

	fishActivationTemplate = `# PVX Global Tool Shims
# This section was automatically added by PVX
# Do not edit manually - use 'pvm tool shim' commands instead

# Add PVX shim directory to PATH
if test -d "{{.ShimDir}}"
    if not contains "{{.ShimDir}}" $PATH
        set -gx PATH {{.PathString}}
    end
end

# PVX Global Tool Shims End
`

	powershellActivationTemplate = `# PVX Global Tool Shims
# This section was automatically added by PVX
# Do not edit manually - use 'pvm tool shim' commands instead

# Add PVX shim directory to PATH
if (Test-Path "{{.ShimDir}}") {
    $shimDir = "{{.ShimDir}}"
    if ($env:PATH -notlike "*$shimDir*") {
        $env:PATH = "{{.PathString}}"
    }
}

# PVX Global Tool Shims End
`
)

// ShellIntegrator handles shell integration for shim PATH management
type ShellIntegrator struct {
	shimDir     string
	pathManager *PathManager
}

// NewShellIntegrator creates a new shell integrator
func NewShellIntegrator(shimDir string) *ShellIntegrator {
	return &ShellIntegrator{
		shimDir:     shimDir,
		pathManager: NewPathManager(shimDir),
	}
}

// ActivationData holds data for generating shell activation scripts
type ActivationData struct {
	ShimDir    string
	PathString string
}

// GenerateActivationScript generates a shell activation script
func (s *ShellIntegrator) GenerateActivationScript(shell string, position PathPosition) (string, error) {
	pathString := s.pathManager.GeneratePathString(position)
	data := ActivationData{
		ShimDir:    s.shimDir,
		PathString: pathString,
	}

	var templateStr string
	switch shell {
	case "bash":
		templateStr = bashActivationTemplate
	case "zsh":
		templateStr = zshActivationTemplate
	case "fish":
		templateStr = fishActivationTemplate
	case "pwsh":
		templateStr = powershellActivationTemplate
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}

	tmpl, err := template.New("activation").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse activation template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute activation template: %w", err)
	}

	return buf.String(), nil
}

// InstallShellIntegration installs shell integration for the given shell
func (s *ShellIntegrator) InstallShellIntegration(shell string, position PathPosition) error {
	configFile, err := GetShellConfigFile(shell)
	if err != nil {
		return fmt.Errorf("failed to get shell config file: %w", err)
	}

	// Generate activation script
	activationScript, err := s.GenerateActivationScript(shell, position)
	if err != nil {
		return fmt.Errorf("failed to generate activation script: %w", err)
	}

	// Check if integration is already installed
	if isInstalled, err := s.IsShellIntegrationInstalled(shell); err != nil {
		return fmt.Errorf("failed to check shell integration status: %w", err)
	} else if isInstalled {
		// Update existing integration
		return s.UpdateShellIntegration(shell, position)
	}

	// Create config file directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Append activation script to config file
	file, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString("\n" + activationScript); err != nil {
		return fmt.Errorf("failed to write activation script: %w", err)
	}

	return nil
}

// IsShellIntegrationInstalled checks if shell integration is already installed
func (s *ShellIntegrator) IsShellIntegrationInstalled(shell string) (bool, error) {
	configFile, err := GetShellConfigFile(shell)
	if err != nil {
		return false, fmt.Errorf("failed to get shell config file: %w", err)
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read config file: %w", err)
	}

	return strings.Contains(string(content), "PVX Global Tool Shims"), nil
}

// UpdateShellIntegration updates existing shell integration
func (s *ShellIntegrator) UpdateShellIntegration(shell string, position PathPosition) error {
	configFile, err := GetShellConfigFile(shell)
	if err != nil {
		return fmt.Errorf("failed to get shell config file: %w", err)
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Generate new activation script
	activationScript, err := s.GenerateActivationScript(shell, position)
	if err != nil {
		return fmt.Errorf("failed to generate activation script: %w", err)
	}

	// Replace existing integration
	contentStr := string(content)
	startMarker := "# PVX Global Tool Shims"
	endMarker := "# PVX Global Tool Shims End"

	startIndex := strings.Index(contentStr, startMarker)
	if startIndex == -1 {
		// No existing integration found, append new one
		return s.InstallShellIntegration(shell, position)
	}

	endIndex := strings.Index(contentStr[startIndex:], endMarker)
	if endIndex == -1 {
		return fmt.Errorf("malformed shell integration: missing end marker")
	}
	endIndex += startIndex + len(endMarker)

	// Replace the section
	newContent := contentStr[:startIndex] + activationScript + contentStr[endIndex:]

	// Write updated content
	if err := os.WriteFile(configFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated config file: %w", err)
	}

	return nil
}

// RemoveShellIntegration removes shell integration
func (s *ShellIntegrator) RemoveShellIntegration(shell string) error {
	configFile, err := GetShellConfigFile(shell)
	if err != nil {
		return fmt.Errorf("failed to get shell config file: %w", err)
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to remove
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	contentStr := string(content)
	startMarker := "# PVX Global Tool Shims"
	endMarker := "# PVX Global Tool Shims End"

	startIndex := strings.Index(contentStr, startMarker)
	if startIndex == -1 {
		return nil // No integration found
	}

	endIndex := strings.Index(contentStr[startIndex:], endMarker)
	if endIndex == -1 {
		return fmt.Errorf("malformed shell integration: missing end marker")
	}
	endIndex += startIndex + len(endMarker)

	// Remove the section (including trailing newline if present)
	newContent := contentStr[:startIndex]
	if endIndex < len(contentStr) && contentStr[endIndex] == '\n' {
		endIndex++
	}
	newContent += contentStr[endIndex:]

	// Clean up multiple consecutive newlines
	newContent = strings.ReplaceAll(newContent, "\n\n\n", "\n\n")

	// Write updated content
	if err := os.WriteFile(configFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated config file: %w", err)
	}

	return nil
}

// GetShellInstructions returns manual installation instructions for the shell
func (s *ShellIntegrator) GetShellInstructions(shell string, position PathPosition) (string, error) {
	configFile, err := GetShellConfigFile(shell)
	if err != nil {
		return "", fmt.Errorf("failed to get shell config file: %w", err)
	}

	activationScript, err := s.GenerateActivationScript(shell, position)
	if err != nil {
		return "", fmt.Errorf("failed to generate activation script: %w", err)
	}

	instructions := fmt.Sprintf(`Manual Shell Integration Instructions for %s:

1. Open your shell configuration file:
   %s

2. Add the following lines to the file:

%s

3. Restart your shell or run:
   source %s

4. Verify the integration by running:
   echo $PATH

The shim directory (%s) should be included in your PATH.
`, shell, configFile, activationScript, configFile, s.shimDir)

	return instructions, nil
}

// DetectAvailableShells detects shells available on the system
func DetectAvailableShells() []string {
	var shells []string

	// Common shell locations
	shellPaths := []string{
		"/bin/bash", "/usr/bin/bash",
		"/bin/zsh", "/usr/bin/zsh", "/usr/local/bin/zsh",
		"/usr/local/bin/fish", "/opt/homebrew/bin/fish",
	}

	// Windows shells
	if runtime.GOOS == "windows" {
		shellPaths = append(shellPaths,
			"C:\\Windows\\System32\\cmd.exe",
			"C:\\Program Files\\PowerShell\\7\\pwsh.exe",
			"C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
		)
	}

	for _, shellPath := range shellPaths {
		if _, err := os.Stat(shellPath); err == nil {
			shellName := filepath.Base(shellPath)
			// Normalize shell names
			switch shellName {
			case "bash":
				shells = append(shells, "bash")
			case "zsh":
				shells = append(shells, "zsh")
			case "fish":
				shells = append(shells, "fish")
			case "cmd.exe":
				shells = append(shells, "cmd")
			case "pwsh.exe", "powershell.exe":
				shells = append(shells, "pwsh")
			}
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueShells []string
	for _, shell := range shells {
		if !seen[shell] {
			seen[shell] = true
			uniqueShells = append(uniqueShells, shell)
		}
	}

	return uniqueShells
}
