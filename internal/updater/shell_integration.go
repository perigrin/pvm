// ABOUTME: Shell integration updates for PVM binary replacement
// ABOUTME: Ensures shell configuration remains functional after updates

package updater

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// ShellIntegrationManager handles shell configuration updates
type ShellIntegrationManager struct {
	detectedShells []ShellInfo
}

// NewShellIntegrationManager creates a new shell integration manager
func NewShellIntegrationManager() *ShellIntegrationManager {
	return &ShellIntegrationManager{}
}

// ShellInfo contains information about a detected shell
type ShellInfo struct {
	Type       ShellType
	ConfigPath string
	Version    string
	IsActive   bool
}

// ShellType represents different shell types
type ShellType int

const (
	ShellUnknown ShellType = iota
	ShellBash
	ShellZsh
	ShellFish
	ShellPowerShell
	ShellCmd
)

// String returns the string representation of a shell type
func (st ShellType) String() string {
	switch st {
	case ShellBash:
		return "bash"
	case ShellZsh:
		return "zsh"
	case ShellFish:
		return "fish"
	case ShellPowerShell:
		return "powershell"
	case ShellCmd:
		return "cmd"
	default:
		return "unknown"
	}
}

// IntegrationResult contains the result of shell integration updates
type IntegrationResult struct {
	UpdatedShells   []ShellInfo
	SkippedShells   []ShellInfo
	FailedShells    []ShellInfo
	RequiresRestart bool
	Instructions    []string
}

// UpdateShellIntegration updates shell configurations after binary replacement
func (sim *ShellIntegrationManager) UpdateShellIntegration(oldPath, newPath string) (*IntegrationResult, error) {
	result := &IntegrationResult{}

	// Detect available shells
	shells, err := sim.detectShells()
	if err != nil {
		return result, fmt.Errorf("detecting shells: %w", err)
	}

	// Update each shell configuration
	for _, shell := range shells {
		if err := sim.updateShellConfig(shell, oldPath, newPath); err != nil {
			result.FailedShells = append(result.FailedShells, shell)
			fmt.Printf("Failed to update %s configuration: %v\n", shell.Type.String(), err)
		} else {
			result.UpdatedShells = append(result.UpdatedShells, shell)
		}
	}

	// Generate user instructions
	result.Instructions = sim.generateInstructions(result)
	result.RequiresRestart = len(result.UpdatedShells) > 0

	return result, nil
}

// detectShells finds available shells and their configuration files
func (sim *ShellIntegrationManager) detectShells() ([]ShellInfo, error) {
	var shells []ShellInfo
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return shells, fmt.Errorf("getting home directory: %w", err)
	}

	// Detect shells based on platform
	if runtime.GOOS == "windows" {
		shells = append(shells, sim.detectWindowsShells(homeDir)...)
	} else {
		shells = append(shells, sim.detectUnixShells(homeDir)...)
	}

	// Filter out shells without config files
	var validShells []ShellInfo
	for _, shell := range shells {
		if _, err := os.Stat(shell.ConfigPath); err == nil {
			validShells = append(validShells, shell)
		}
	}

	return validShells, nil
}

// detectUnixShells detects Unix-like shell configurations
func (sim *ShellIntegrationManager) detectUnixShells(homeDir string) []ShellInfo {
	var shells []ShellInfo

	// Bash configurations
	bashConfigs := []string{
		".bashrc",
		".bash_profile",
		".profile",
	}

	for _, config := range bashConfigs {
		configPath := filepath.Join(homeDir, config)
		if sim.containsPVMReference(configPath) {
			shells = append(shells, ShellInfo{
				Type:       ShellBash,
				ConfigPath: configPath,
				IsActive:   sim.isActiveShell("bash"),
			})
		}
	}

	// Zsh configurations
	zshConfigs := []string{
		".zshrc",
		".zprofile",
	}

	for _, config := range zshConfigs {
		configPath := filepath.Join(homeDir, config)
		if sim.containsPVMReference(configPath) {
			shells = append(shells, ShellInfo{
				Type:       ShellZsh,
				ConfigPath: configPath,
				IsActive:   sim.isActiveShell("zsh"),
			})
		}
	}

	// Fish configuration
	fishConfigPath := filepath.Join(homeDir, ".config", "fish", "config.fish")
	if sim.containsPVMReference(fishConfigPath) {
		shells = append(shells, ShellInfo{
			Type:       ShellFish,
			ConfigPath: fishConfigPath,
			IsActive:   sim.isActiveShell("fish"),
		})
	}

	return shells
}

// detectWindowsShells detects Windows shell configurations
func (sim *ShellIntegrationManager) detectWindowsShells(homeDir string) []ShellInfo {
	var shells []ShellInfo

	// PowerShell profile
	psProfile := filepath.Join(homeDir, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	if sim.containsPVMReference(psProfile) {
		shells = append(shells, ShellInfo{
			Type:       ShellPowerShell,
			ConfigPath: psProfile,
			IsActive:   true, // Assume PowerShell is available on Windows
		})
	}

	// Windows PowerShell (5.x) profile
	winPsProfile := filepath.Join(homeDir, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	if sim.containsPVMReference(winPsProfile) {
		shells = append(shells, ShellInfo{
			Type:       ShellPowerShell,
			ConfigPath: winPsProfile,
			IsActive:   true,
		})
	}

	return shells
}

// containsPVMReference checks if a file contains PVM-related configurations
func (sim *ShellIntegrationManager) containsPVMReference(configPath string) bool {
	file, err := os.Open(configPath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	pvmPattern := regexp.MustCompile(`(?i)(pvm|\.pvm|pvm_|PVM_)`)

	for scanner.Scan() {
		line := scanner.Text()
		if pvmPattern.MatchString(line) {
			return true
		}
	}

	return false
}

// isActiveShell checks if a shell is currently active/available
func (sim *ShellIntegrationManager) isActiveShell(shellName string) bool {
	// Check if shell is available in PATH
	_, err := os.Stat(fmt.Sprintf("/bin/%s", shellName))
	if err == nil {
		return true
	}

	_, err = os.Stat(fmt.Sprintf("/usr/bin/%s", shellName))
	return err == nil
}

// updateShellConfig updates a shell configuration to use the new binary path
func (sim *ShellIntegrationManager) updateShellConfig(shell ShellInfo, oldPath, newPath string) error {
	// Read the configuration file
	content, err := os.ReadFile(shell.ConfigPath)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	// Create backup of original config
	backupPath := shell.ConfigPath + ".pvm-backup"
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}

	// Update content based on shell type
	updatedContent := sim.updateConfigContent(string(content), shell.Type, oldPath, newPath)

	// Write updated configuration
	if err := os.WriteFile(shell.ConfigPath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("writing updated config: %w", err)
	}

	return nil
}

// updateConfigContent updates the configuration content for a specific shell
func (sim *ShellIntegrationManager) updateConfigContent(content string, shellType ShellType, oldPath, newPath string) string {
	lines := strings.Split(content, "\n")
	var updatedLines []string

	for _, line := range lines {
		updatedLine := sim.updateConfigLine(line, shellType, oldPath, newPath)
		updatedLines = append(updatedLines, updatedLine)
	}

	return strings.Join(updatedLines, "\n")
}

// updateConfigLine updates a single configuration line
func (sim *ShellIntegrationManager) updateConfigLine(line string, shellType ShellType, oldPath, newPath string) string {
	// Patterns to match and update
	patterns := []struct {
		regex       *regexp.Regexp
		replacement string
	}{
		// Direct path references
		{
			regex:       regexp.MustCompile(regexp.QuoteMeta(oldPath)),
			replacement: newPath,
		},
		// Directory references (update to new directory)
		{
			regex:       regexp.MustCompile(regexp.QuoteMeta(filepath.Dir(oldPath))),
			replacement: filepath.Dir(newPath),
		},
	}

	updatedLine := line
	for _, pattern := range patterns {
		updatedLine = pattern.regex.ReplaceAllString(updatedLine, pattern.replacement)
	}

	return updatedLine
}

// generateInstructions creates user instructions for completing the integration
func (sim *ShellIntegrationManager) generateInstructions(result *IntegrationResult) []string {
	var instructions []string

	if len(result.UpdatedShells) > 0 {
		instructions = append(instructions, "Shell configurations have been updated.")
		instructions = append(instructions, "")
		instructions = append(instructions, "To activate the changes:")

		for _, shell := range result.UpdatedShells {
			switch shell.Type {
			case ShellBash:
				instructions = append(instructions, fmt.Sprintf("  • For Bash: source %s", shell.ConfigPath))
			case ShellZsh:
				instructions = append(instructions, fmt.Sprintf("  • For Zsh: source %s", shell.ConfigPath))
			case ShellFish:
				instructions = append(instructions, "  • For Fish: restart your terminal or run 'exec fish'")
			case ShellPowerShell:
				instructions = append(instructions, "  • For PowerShell: restart your PowerShell session")
			}
		}

		instructions = append(instructions, "")
		instructions = append(instructions, "Alternatively, restart your terminal to apply all changes.")
	}

	if len(result.FailedShells) > 0 {
		instructions = append(instructions, "")
		instructions = append(instructions, "Manual updates may be required for:")
		for _, shell := range result.FailedShells {
			instructions = append(instructions, fmt.Sprintf("  • %s configuration: %s", shell.Type.String(), shell.ConfigPath))
		}
	}

	return instructions
}

// ValidateShellIntegration checks if shell integration is working correctly
func (sim *ShellIntegrationManager) ValidateShellIntegration(binaryPath string) error {
	shells, err := sim.detectShells()
	if err != nil {
		return fmt.Errorf("detecting shells for validation: %w", err)
	}

	for _, shell := range shells {
		if err := sim.validateShellConfig(shell, binaryPath); err != nil {
			return fmt.Errorf("validating %s config: %w", shell.Type.String(), err)
		}
	}

	return nil
}

// validateShellConfig validates that a shell configuration is correct
func (sim *ShellIntegrationManager) validateShellConfig(shell ShellInfo, binaryPath string) error {
	content, err := os.ReadFile(shell.ConfigPath)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	// Check if the binary path is referenced correctly
	configText := string(content)
	binaryDir := filepath.Dir(binaryPath)

	if !strings.Contains(configText, binaryPath) && !strings.Contains(configText, binaryDir) {
		return fmt.Errorf("binary path not found in configuration")
	}

	return nil
}

// RestoreShellBackups restores shell configurations from backups
func (sim *ShellIntegrationManager) RestoreShellBackups() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	// Find backup files
	pattern := filepath.Join(homeDir, "**", "*.pvm-backup")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("finding backup files: %w", err)
	}

	for _, backupPath := range matches {
		originalPath := strings.TrimSuffix(backupPath, ".pvm-backup")

		// Read backup content
		content, err := os.ReadFile(backupPath)
		if err != nil {
			fmt.Printf("Failed to read backup %s: %v\n", backupPath, err)
			continue
		}

		// Restore original file
		if err := os.WriteFile(originalPath, content, 0644); err != nil {
			fmt.Printf("Failed to restore %s: %v\n", originalPath, err)
			continue
		}

		// Remove backup file
		os.Remove(backupPath)
		fmt.Printf("Restored shell configuration: %s\n", originalPath)
	}

	return nil
}

// GetShellStatus returns information about current shell integration status
func (sim *ShellIntegrationManager) GetShellStatus() (*ShellStatus, error) {
	shells, err := sim.detectShells()
	if err != nil {
		return nil, fmt.Errorf("detecting shells: %w", err)
	}

	status := &ShellStatus{
		DetectedShells: len(shells),
		ActiveShells:   0,
	}

	for _, shell := range shells {
		if shell.IsActive {
			status.ActiveShells++
		}
		status.Shells = append(status.Shells, shell)
	}

	return status, nil
}

// ShellStatus contains information about shell integration status
type ShellStatus struct {
	DetectedShells int         `json:"detected_shells"`
	ActiveShells   int         `json:"active_shells"`
	Shells         []ShellInfo `json:"shells"`
}
