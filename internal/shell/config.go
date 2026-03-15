// ABOUTME: Shell configuration file management for auto-fix operations
// ABOUTME: Handles safe modification of shell configuration files with backup support

package shell

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tamarou.com/pvm/internal/backup"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/perl"
)

// ConfigManager handles shell configuration file modifications
type ConfigManager struct {
	backup *backup.Manager
}

// NewConfigManager creates a new shell configuration manager
func NewConfigManager() (*ConfigManager, error) {
	backupMgr, err := backup.NewManager()
	if err != nil {
		return nil, err
	}

	return &ConfigManager{
		backup: backupMgr,
	}, nil
}

// ShellConfig represents a shell configuration
type ShellConfig struct {
	Shell       perl.ShellType
	ConfigFiles []string
	InitCommand string
	HomeDir     string
}

// DetectShellConfig detects the current shell and its configuration files
func (cm *ConfigManager) DetectShellConfig() (*ShellConfig, error) {
	shell, err := perl.DetectShell()
	if err != nil {
		return nil, errors.NewSystemError("009", "Failed to detect shell type", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.NewSystemError("010", "Failed to get home directory", err)
	}

	config := &ShellConfig{
		Shell:   shell,
		HomeDir: homeDir,
	}

	// Set shell-specific configuration
	switch shell {
	case perl.ShellBash:
		config.ConfigFiles = []string{".bashrc", ".bash_profile"}
		config.InitCommand = `eval "$(pvm init)"`
	case perl.ShellZsh:
		config.ConfigFiles = []string{".zshrc", ".zprofile"}
		config.InitCommand = `eval "$(pvm init)"`
	case perl.ShellFish:
		config.ConfigFiles = []string{".config/fish/config.fish"}
		config.InitCommand = `pvm init | source`
	case perl.ShellPowerShell:
		config.ConfigFiles = []string{
			filepath.Join("Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"),
			filepath.Join("Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"),
		}
		config.InitCommand = `pvm init | Invoke-Expression`
	default:
		return nil, errors.NewSystemError("011",
			fmt.Sprintf("Unsupported shell type: %s", shell), nil)
	}

	return config, nil
}

// CheckShellIntegration checks if shell integration is present in configuration
func (cm *ConfigManager) CheckShellIntegration(config *ShellConfig) (bool, string, error) {
	for _, configFile := range config.ConfigFiles {
		configPath := filepath.Join(config.HomeDir, configFile)

		// Check if file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue
		}

		// Read file and check for pvm init
		hasInit, err := cm.fileContainsPVMInit(configPath)
		if err != nil {
			continue // Skip files we can't read
		}

		if hasInit {
			return true, configFile, nil
		}
	}

	return false, "", nil
}

// AddShellIntegration adds PVM initialization to shell configuration
func (cm *ConfigManager) AddShellIntegration(session *backup.BackupSession, config *ShellConfig, force bool) error {
	// Find the primary config file to modify
	configFile := cm.getPrimaryConfigFile(config)
	configPath := filepath.Join(config.HomeDir, configFile)

	// Check if integration already exists (unless force is true)
	if !force {
		hasInit, existingFile, err := cm.CheckShellIntegration(config)
		if err != nil {
			return err
		}
		if hasInit {
			return fmt.Errorf("PVM integration already exists in %s", existingFile)
		}
	}

	// Create backup if file exists
	if _, err := os.Stat(configPath); err == nil {
		err = cm.backup.BackupFile(session, configPath)
		if err != nil {
			return errors.NewSystemError("012", "Failed to backup configuration file", err).
				WithLocation(configPath)
		}
	}

	// Create directory if it doesn't exist (for fish config)
	configDir := filepath.Dir(configPath)
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		return errors.NewSystemError("013", "Failed to create configuration directory", err).
			WithLocation(configDir)
	}

	// Add PVM initialization
	err = cm.appendPVMInit(configPath, config)
	if err != nil {
		return errors.NewSystemError("014", "Failed to add PVM initialization", err).
			WithLocation(configPath)
	}

	return nil
}

// RemoveShellIntegration removes PVM initialization from shell configuration
func (cm *ConfigManager) RemoveShellIntegration(session *backup.BackupSession, config *ShellConfig) error {
	var errors []error

	for _, configFile := range config.ConfigFiles {
		configPath := filepath.Join(config.HomeDir, configFile)

		// Check if file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue
		}

		// Check if file contains PVM init
		hasInit, err := cm.fileContainsPVMInit(configPath)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		if !hasInit {
			continue // Nothing to remove
		}

		// Create backup
		err = cm.backup.BackupFile(session, configPath)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		// Remove PVM init lines
		err = cm.removePVMInit(configPath)
		if err != nil {
			errors = append(errors, err)
			continue
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors removing shell integration: %v", errors)
	}

	return nil
}

// Helper methods

// getPrimaryConfigFile returns the primary configuration file for the shell
func (cm *ConfigManager) getPrimaryConfigFile(config *ShellConfig) string {
	if len(config.ConfigFiles) == 0 {
		return ""
	}

	// Return the first (primary) config file
	return config.ConfigFiles[0]
}

// fileContainsPVMInit checks if a file contains PVM initialization
func (cm *ConfigManager) fileContainsPVMInit(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "pvm init") {
			return true, nil
		}
	}

	return false, scanner.Err()
}

// appendPVMInit adds PVM initialization to a configuration file
func (cm *ConfigManager) appendPVMInit(filePath string, config *ShellConfig) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add some spacing and comments
	content := fmt.Sprintf(`

# PVM (Perl Version Manager) initialization
# Added by 'pvm doctor --fix'
%s
`, config.InitCommand)

	_, err = file.WriteString(content)
	return err
}

// removePVMInit removes PVM initialization lines from a configuration file
func (cm *ConfigManager) removePVMInit(filePath string) error {
	// Read file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	inPVMSection := false

	for scanner.Scan() {
		line := scanner.Text()

		// Check if this line starts a PVM section
		if strings.Contains(line, "PVM (Perl Version Manager)") ||
			strings.Contains(line, "Added by 'pvm doctor --fix'") {
			inPVMSection = true
			continue
		}

		// Check if this is a PVM init line
		if strings.Contains(line, "pvm init") {
			// Skip this line and any PVM section we might be in
			inPVMSection = false
			continue
		}

		// Skip empty lines in PVM section
		if inPVMSection && strings.TrimSpace(line) == "" {
			continue
		}

		// If we have content, we're out of the PVM section
		if inPVMSection && strings.TrimSpace(line) != "" {
			inPVMSection = false
		}

		// Keep non-PVM lines
		if !inPVMSection {
			lines = append(lines, line)
		}
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		return err
	}

	// Write file back
	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

// ValidateShellIntegration validates that shell integration was added correctly
func (cm *ConfigManager) ValidateShellIntegration(config *ShellConfig) error {
	hasInit, configFile, err := cm.CheckShellIntegration(config)
	if err != nil {
		return errors.NewSystemError("015", "Failed to validate shell integration", err)
	}

	if !hasInit {
		return errors.NewSystemError("016", "Shell integration validation failed", nil).
			WithHint("PVM initialization was not found in shell configuration")
	}

	// Additional validation - check if the config file is readable
	configPath := filepath.Join(config.HomeDir, configFile)
	file, err := os.Open(configPath)
	if err != nil {
		return errors.NewSystemError("017", "Configuration file is not readable", err).
			WithLocation(configPath)
	}
	file.Close()

	return nil
}
