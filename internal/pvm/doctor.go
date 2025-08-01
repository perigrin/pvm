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
	ui.Success("Shell integration available via: eval \"$(pvm init)\"")

	// Check if shell integration is loaded in shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		*warnings = append(*warnings, "SHELL environment variable not set")
		return nil
	}

	shellName := filepath.Base(shell)
	ui.Info("Detected shell: %s", shellName)

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

	configs, exists := shellConfigs[shellName]
	if !exists {
		*warnings = append(*warnings, fmt.Sprintf("Unknown shell: %s", shellName))
		return nil
	}

	foundInitCall := false
	for _, config := range configs {
		configPath := filepath.Join(homeDir, config)
		if data, err := os.ReadFile(configPath); err == nil {
			content := string(data)
			// Check for various pvm init patterns
			if strings.Contains(content, "pvm init") ||
				strings.Contains(content, "pvm_path init") ||
				strings.Contains(content, "$pvm_path init") {
				foundInitCall = true
				ui.Success("Found pvm init in %s", config)
				break
			}
		}
	}

	if !foundInitCall {
		*warnings = append(*warnings, "Shell configuration doesn't contain 'eval \"$(pvm init)\"'")
		ui.Warning("Add 'eval \"$(pvm init)\"' to your shell configuration file")
	}

	// Check if shell integration is actually active in current session
	err = checkActiveShellIntegration(ui, issues, warnings, shellName)
	if err != nil {
		return err
	}

	return nil
}

// checkActiveShellIntegration checks if shell integration is active in the current session
func checkActiveShellIntegration(ui *ui.Output, issues *[]string, warnings *[]string, shellName string) error {
	// Test if PVM shell functions are defined by running shell commands
	switch shellName {
	case "bash":
		// Check if cd is aliased to pvm_cd
		cmd := exec.Command("bash", "-c", "type cd")
		output, err := cmd.Output()
		if err == nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "pvm_cd") {
				ui.Success("Shell integration active: cd is aliased to pvm_cd")
			} else {
				*warnings = append(*warnings, "Shell integration not active: cd is not aliased")
				ui.Warning("Shell integration not loaded in current session")
				ui.Info("Run 'eval \"$(pvm init)\"' or restart your shell")
			}
		}

	case "zsh":
		// Check if pvm_chpwd function exists and chpwd hook is registered
		cmd := exec.Command("zsh", "-c", "type pvm_chpwd 2>/dev/null")
		err := cmd.Run()
		if err == nil {
			ui.Success("Shell integration active: pvm_chpwd function exists")

			// Check if chpwd hook is registered
			hookCmd := exec.Command("zsh", "-c", "echo $chpwd_functions | grep -q pvm_chpwd")
			hookErr := hookCmd.Run()
			if hookErr == nil {
				ui.Success("chpwd hook properly registered")
			} else {
				*warnings = append(*warnings, "pvm_chpwd function exists but hook not registered")
				ui.Warning("chpwd hook not registered properly")
			}
		} else {
			*warnings = append(*warnings, "Shell integration not active: pvm_chpwd function not found")
			ui.Warning("Shell integration not loaded in current session")
			ui.Info("Run 'eval \"$(pvm init)\"' or restart your shell")
		}

	case "fish":
		// Check if fish functions exist
		cmd := exec.Command("fish", "-c", "functions -q pvm")
		err := cmd.Run()
		if err == nil {
			ui.Success("Shell integration active: pvm function exists")
		} else {
			*warnings = append(*warnings, "Shell integration not active: pvm function not found")
			ui.Warning("Shell integration not loaded in current session")
		}
	}

	// Check if the pvm executable is accessible from shell integration
	cmd := exec.Command(shellName, "-c", "command -v pvm")
	output, err := cmd.Output()
	if err == nil {
		pvmPath := strings.TrimSpace(string(output))
		ui.Success("PVM executable accessible at: %s", pvmPath)

		// Check if it's the same binary we're running from
		currentExec, _ := os.Executable()
		if currentExec != "" {
			// Resolve symlinks for accurate comparison
			currentPath, _ := filepath.EvalSymlinks(currentExec)
			shellPath, _ := filepath.EvalSymlinks(pvmPath)

			// Convert to absolute paths for comparison
			currentPath, _ = filepath.Abs(currentPath)
			shellPath, _ = filepath.Abs(shellPath)

			if currentPath != shellPath {
				*warnings = append(*warnings, "Shell integration using different PVM binary")
				ui.Warning("Shell integration using: %s (resolves to %s)", pvmPath, shellPath)
				ui.Warning("Current doctor running from: %s", currentPath)
				ui.Info("This may cause inconsistent behavior")
			} else {
				ui.Success("Shell integration using same PVM binary")
				if pvmPath != currentPath {
					ui.Info("Via symlink: %s -> %s", pvmPath, currentPath)
				}
			}
		}
	} else {
		*issues = append(*issues, "PVM executable not accessible from shell")
		ui.Error("Shell cannot find pvm executable")
		ui.Info("Ensure PVM is in your PATH or shell integration is properly loaded")
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

// checkRegistryIntegrity checks if the version registry is intact
func checkRegistryIntegrity(ui *ui.Output, issues *[]string, warnings *[]string) error {
	// Get XDG directories for paths
	dirs, err := xdg.GetDirs()
	if err != nil {
		*issues = append(*issues, "Failed to determine XDG directories")
		return err
	}

	registryPath := filepath.Join(dirs.DataDir, "registry.json")
	versionsDir := filepath.Join(dirs.DataDir, "versions")

	// Check if registry file exists
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		// Check if versions directory exists with installations
		if entries, err := os.ReadDir(versionsDir); err == nil && len(entries) > 0 {
			*issues = append(*issues, "Registry missing but versions directory contains installations")
			ui.Error("Registry file missing: %s", registryPath)
			return nil
		} else {
			*warnings = append(*warnings, "Registry file doesn't exist (no versions installed)")
			ui.Warning("No registry file found, but no versions are installed")
			return nil
		}
	}

	// Load and validate registry
	registry, err := perl.LoadRegistry()
	if err != nil {
		*issues = append(*issues, "Failed to load registry file")
		ui.Error("Registry file corrupted: %v", err)
		return nil
	}

	// Check if registry is empty but versions exist
	if len(registry.Versions) == 0 {
		if entries, err := os.ReadDir(versionsDir); err == nil && len(entries) > 0 {
			*issues = append(*issues, "Registry is empty but versions directory contains installations")
			ui.Error("Registry file is empty but versions exist")
			return nil
		}
	}

	// Check for registry-filesystem mismatch
	registryVersions := make(map[string]bool)
	for version := range registry.Versions {
		registryVersions[version] = true
	}

	// Check filesystem versions
	if entries, err := os.ReadDir(versionsDir); err == nil {
		filesystemVersions := []string{}
		missingFromRegistry := []string{}

		for _, entry := range entries {
			if entry.IsDir() || entry.Type() == os.ModeSymlink {
				// Validate it's a real Perl installation
				versionPath := filepath.Join(versionsDir, entry.Name())
				perlBinary := filepath.Join(versionPath, "bin", "perl")
				if _, err := os.Stat(perlBinary); err == nil {
					filesystemVersions = append(filesystemVersions, entry.Name())
					if !registryVersions[entry.Name()] {
						missingFromRegistry = append(missingFromRegistry, entry.Name())
						*warnings = append(*warnings, fmt.Sprintf("Version %s exists on filesystem but not in registry", entry.Name()))
					}
				}
			}
		}

		// If we have versions missing from registry, mark this as an issue that can be auto-fixed
		if len(missingFromRegistry) > 0 {
			*issues = append(*issues, fmt.Sprintf("Registry missing %d version(s) that exist on filesystem", len(missingFromRegistry)))
		}

		// Check for registry entries without filesystem presence
		orphanedInRegistry := []string{}
		for version, versionInfo := range registry.Versions {
			found := false

			// Special handling for system version - check the actual install path
			if version == "system" {
				perlBinary := filepath.Join(versionInfo.InstallPath, "perl")
				if _, err := os.Stat(perlBinary); err == nil {
					found = true
				}
			} else {
				// Regular version - check in filesystem versions
				for _, fsVersion := range filesystemVersions {
					if fsVersion == version {
						found = true
						break
					}
				}
			}

			if !found {
				orphanedInRegistry = append(orphanedInRegistry, version)
				*warnings = append(*warnings, fmt.Sprintf("Version %s in registry but not on filesystem", version))
			}
		}

		// If we have orphaned registry entries, mark this as an issue that can be auto-fixed
		if len(orphanedInRegistry) > 0 {
			*issues = append(*issues, fmt.Sprintf("Registry contains %d version(s) that don't exist on filesystem", len(orphanedInRegistry)))
		}
	}

	ui.Success("Registry integrity check passed")
	return nil
}

// checkFilesystemLocations shows where PVM stores its files
func checkFilesystemLocations(ui *ui.Output, issues *[]string, warnings *[]string) error {
	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		*issues = append(*issues, "Failed to determine XDG directories")
		return err
	}

	// Get XDG_BIN_HOME for current shim location
	xdgBinHome := os.Getenv("XDG_BIN_HOME")
	if xdgBinHome == "" {
		homeDir, _ := os.UserHomeDir()
		xdgBinHome = filepath.Join(homeDir, ".local", "bin")
	}

	ui.Success("PVM filesystem locations:")
	ui.Info("  Data directory: %s", dirs.DataDir)
	ui.Info("  Registry file: %s", filepath.Join(dirs.DataDir, "registry.json"))
	ui.Info("  Versions directory: %s", filepath.Join(dirs.DataDir, "versions"))
	ui.Info("  XDG_BIN_HOME (current shims): %s", xdgBinHome)
	ui.Info("  Legacy shims directory: %s (deprecated)", filepath.Join(dirs.DataDir, "shims"))
	ui.Info("  Shell integration: Generated dynamically (no directory needed)")
	ui.Info("  Type definitions: %s", filepath.Join(dirs.DataDir, "type_definitions"))

	// Check if directories exist and show their status
	locations := map[string]string{
		"Data directory": dirs.DataDir,
		"Versions":       filepath.Join(dirs.DataDir, "versions"),
		"XDG_BIN_HOME":   xdgBinHome,
	}

	allExist := true
	for name, path := range locations {
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				ui.Success("  ✓ %s directory exists", name)
			} else {
				*warnings = append(*warnings, fmt.Sprintf("%s path exists but is not a directory", name))
				ui.Warning("  ! %s path exists but is not a directory", name)
			}
		} else if os.IsNotExist(err) {
			*warnings = append(*warnings, fmt.Sprintf("%s directory missing", name))
			ui.Warning("  ! %s directory missing", name)
			allExist = false
		}
	}

	if allExist {
		ui.Success("All required directories exist")
	}

	return nil
}

// getManualFixInstructions returns help text for issues that can't be automatically fixed
func getManualFixInstructions(issue string) string {
	switch {
	case strings.Contains(issue, "PATH environment variable not set"):
		return `PATH environment variable is not set. This is unusual and may indicate a severe system configuration issue.
Try running: export PATH="/usr/local/bin:/usr/bin:/bin"`

	case strings.Contains(issue, "XDG_BIN_HOME not in PATH"):
		return `XDG_BIN_HOME is not in PATH. This usually means shell integration isn't active.
Make sure your shell configuration file contains:
  eval "$(pvm init)"

Add this line to:
  - Bash: ~/.bashrc or ~/.bash_profile
  - Zsh: ~/.zshrc
  - Fish: ~/.config/fish/config.fish

Do NOT manually add XDG_BIN_HOME to PATH - let 'pvm init' handle it automatically.`

	case strings.Contains(issue, "PVM executable not found in system PATH"):
		return `PVM is not in your system PATH. Add the PVM binary location to PATH:
  export PATH="/path/to/pvm/bin:$PATH"

Or create a symlink: ln -s /path/to/pvm/binary /usr/local/bin/pvm`

	case strings.Contains(issue, "Shell integration files not found"):
		return `Shell integration is available without files. Use:
  eval "$(pvm init)"

Add this to your shell configuration file.`

	case strings.Contains(issue, "Shell configuration doesn't contain"):
		return `Your shell configuration needs to initialize PVM. Add this line to your shell config:
  eval "$(pvm init)"

For bash: add to ~/.bashrc or ~/.bash_profile
For zsh: add to ~/.zshrc
For fish: add to ~/.config/fish/config.fish`

	case strings.Contains(issue, "directory missing"):
		return `Missing PVM directories. These should be created automatically.
Try running: pvm init
If the issue persists, reinstall PVM.`

	case strings.Contains(issue, "Failed to determine XDG directories"):
		return `XDG directory detection failed. This may be a system configuration issue.
Ensure your HOME environment variable is set:
  echo $HOME

If unset, try: export HOME="/path/to/your/home/directory"`

	case strings.Contains(issue, "No Perl versions installed"):
		return `No Perl versions are currently installed. Install a Perl version:
  pvm install latest       # Install latest stable version
  pvm install 5.40.0      # Install specific version
  pvm install -B 5.40.0   # Install from pre-compiled binary (faster)`

	case strings.Contains(issue, "Version") && strings.Contains(issue, "exists on filesystem but not in registry"):
		return `Some versions exist on filesystem but are missing from registry.
Run: pvm init
This will scan and register all existing installations.`

	case strings.Contains(issue, "Version") && strings.Contains(issue, "in registry but not on filesystem"):
		return `Registry contains versions that don't exist on filesystem.
This usually happens after manual deletion. Run:
  pvm init
This will rebuild the registry from actual installations.`

	default:
		return ""
	}
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

	// Check if XDG_BIN_HOME is in PATH (replaces old shims directory check)
	// XDG_BIN_HOME defaults to $HOME/.local/bin
	xdgBinHome := os.Getenv("XDG_BIN_HOME")
	if xdgBinHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			*warnings = append(*warnings, "Failed to determine home directory for XDG_BIN_HOME check")
			return nil
		}
		xdgBinHome = filepath.Join(homeDir, ".local", "bin")
	}

	pathDirs := strings.Split(path, ":")
	xdgBinInPath := false
	for _, dir := range pathDirs {
		if dir == xdgBinHome {
			xdgBinInPath = true
			break
		}
	}

	if !xdgBinInPath {
		*issues = append(*issues, "XDG_BIN_HOME not in PATH")
		ui.Error("XDG_BIN_HOME (%s) not in PATH. Shell integration is not active.", xdgBinHome)
		ui.Info("This usually means 'eval \"$(pvm init)\"' is not in your shell config.")
	} else {
		ui.Success("XDG_BIN_HOME (%s) found in PATH", xdgBinHome)
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

// checkShimsDirectory checks if legacy shims directory exists (informational only)
func checkShimsDirectory(ui *ui.Output, issues *[]string, warnings *[]string) error {
	dirs, err := xdg.GetDirs()
	if err != nil {
		*issues = append(*issues, "Failed to determine XDG directories")
		return err
	}

	shimsDir := filepath.Join(dirs.DataDir, "shims")

	// Check if legacy shims directory exists
	if _, err := os.Stat(shimsDir); os.IsNotExist(err) {
		ui.Info("Legacy shims directory not found (expected - PVM now uses XDG_BIN_HOME)")
		return nil
	}

	// Check if legacy shims directory contains files
	files, err := os.ReadDir(shimsDir)
	if err != nil {
		*warnings = append(*warnings, "Failed to read legacy shims directory")
		return nil
	}

	foundShims := []string{}
	for _, file := range files {
		if !file.IsDir() {
			foundShims = append(foundShims, file.Name())
		}
	}

	if len(foundShims) == 0 {
		ui.Info("Legacy shims directory exists but is empty (can be safely removed)")
	} else {
		ui.Warning("Legacy shims directory contains %d file(s): %v", len(foundShims), foundShims)
		ui.Info("PVM now uses XDG_BIN_HOME instead of legacy shims - these files can be safely removed")
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
