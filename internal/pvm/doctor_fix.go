// ABOUTME: Auto-fix engine for PVM doctor command
// ABOUTME: Provides automatic resolution of common shell integration and configuration issues

package pvm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tamarou.com/pvm/internal/backup"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/shell"
	"tamarou.com/pvm/internal/xdg"
)

// AutoFix represents a single auto-fix operation
type AutoFix struct {
	ID              string
	Name            string
	Description     string
	CheckFunc       func(*FixContext) (bool, error)
	FixFunc         func(*FixContext) error
	RequiresRestart bool
	Priority        int // Lower numbers are higher priority
}

// FixContext provides context for auto-fix operations
type FixContext struct {
	UI        *ui.Output
	DryRun    bool
	Force     bool
	Session   *backup.BackupSession
	Shell     *shell.ShellConfig
	HomeDir   string
	XDGDirs   *xdg.Dirs
	BackupMgr *backup.Manager
	ShellMgr  *shell.ConfigManager
}

// FixResult represents the result of running auto-fixes
type FixResult struct {
	Fixed             []string
	Failed            []string
	Skipped           []string
	RequiresRestart   bool
	ManualActionItems []string
	Session           *backup.BackupSession
}

// AutoFixEngine manages and executes auto-fix operations
type AutoFixEngine struct {
	fixes []AutoFix
}

// NewAutoFixEngine creates a new auto-fix engine with all available fixes
func NewAutoFixEngine() *AutoFixEngine {
	engine := &AutoFixEngine{
		fixes: make([]AutoFix, 0),
	}

	// Register all available auto-fixes (ordered by priority)
	engine.registerFixes()

	return engine
}

// registerFixes registers all available auto-fix operations
func (afe *AutoFixEngine) registerFixes() {
	// High priority fixes (infrastructure)
	afe.fixes = append(afe.fixes, AutoFix{
		ID:          "missing-directories",
		Name:        "Missing Directories",
		Description: "Create missing PVM directories",
		CheckFunc:   afe.checkMissingDirectories,
		FixFunc:     afe.fixMissingDirectories,
		Priority:    1,
	})

	// Medium priority fixes (configuration)
	afe.fixes = append(afe.fixes, AutoFix{
		ID:              "shell-integration",
		Name:            "Shell Integration",
		Description:     "Add PVM initialization to shell configuration",
		CheckFunc:       afe.checkShellIntegration,
		FixFunc:         afe.fixShellIntegration,
		RequiresRestart: true,
		Priority:        2,
	})

	afe.fixes = append(afe.fixes, AutoFix{
		ID:          "registry-integrity",
		Name:        "Registry Integrity",
		Description: "Rebuild registry from filesystem",
		CheckFunc:   afe.checkRegistryIntegrity,
		FixFunc:     afe.fixRegistryIntegrity,
		Priority:    3,
	})

	// Lower priority fixes (optimization)
	afe.fixes = append(afe.fixes, AutoFix{
		ID:          "shell-script-permissions",
		Name:        "Shell Script Permissions",
		Description: "Fix permissions on shell integration scripts",
		CheckFunc:   afe.checkScriptPermissions,
		FixFunc:     afe.fixScriptPermissions,
		Priority:    4,
	})
}

// HasApplicableFixes reports whether any registered auto-fix would actually
// run against the current environment. Used by the doctor summary to decide
// whether suggesting `--fix` will resolve the user's warnings or just be
// noise. CheckFunc errors are treated as "needs investigation" (returns true)
// so we err on the side of pointing the user at --fix when uncertain.
func (afe *AutoFixEngine) HasApplicableFixes(ctx *FixContext) bool {
	for _, fix := range afe.fixes {
		needsFix, err := fix.CheckFunc(ctx)
		if err != nil {
			return true
		}
		if needsFix {
			return true
		}
	}
	return false
}

// RunAutoFixes executes auto-fixes based on the provided context
func (afe *AutoFixEngine) RunAutoFixes(ctx *FixContext) (*FixResult, error) {
	result := &FixResult{
		Fixed:             make([]string, 0),
		Failed:            make([]string, 0),
		Skipped:           make([]string, 0),
		ManualActionItems: make([]string, 0),
		Session:           ctx.Session,
	}

	ctx.UI.Info("Scanning for fixable issues...")
	ctx.UI.Printf("")

	// Check and fix each issue in priority order
	for _, fix := range afe.fixes {
		ctx.UI.Printf("Checking: %s", fix.Name)

		needsFix, err := fix.CheckFunc(ctx)
		if err != nil {
			ctx.UI.Warning("  Failed to check: %s", err.Error())
			result.Failed = append(result.Failed, fix.Name)
			continue
		}

		if !needsFix {
			ctx.UI.Success("  ✓ Already configured correctly")
			result.Skipped = append(result.Skipped, fix.Name)
			continue
		}

		if ctx.DryRun {
			ctx.UI.Info("  → Would fix: %s", fix.Description)
			continue
		}

		// Attempt the fix
		ctx.UI.Info("  → Fixing: %s", fix.Description)
		err = fix.FixFunc(ctx)
		if err != nil {
			ctx.UI.Error("  ✗ Fix failed: %s", err.Error())
			result.Failed = append(result.Failed, fix.Name)
			continue
		}

		ctx.UI.Success("  ✓ Fixed: %s", fix.Name)
		result.Fixed = append(result.Fixed, fix.Name)

		if fix.RequiresRestart {
			result.RequiresRestart = true
		}
	}

	// Add manual action items if shell restart is required
	if result.RequiresRestart {
		shellName := string(ctx.Shell.Shell)
		switch ctx.Shell.Shell {
		case perl.ShellBash:
			result.ManualActionItems = append(result.ManualActionItems, "Restart your shell or run: source ~/.bashrc")
		case perl.ShellZsh:
			result.ManualActionItems = append(result.ManualActionItems, "Restart your shell or run: source ~/.zshrc")
		case perl.ShellFish:
			result.ManualActionItems = append(result.ManualActionItems, "Restart your shell or run: source ~/.config/fish/config.fish")
		default:
			result.ManualActionItems = append(result.ManualActionItems, fmt.Sprintf("Restart your %s shell", shellName))
		}
	}

	return result, nil
}

// Individual auto-fix implementations

// checkMissingDirectories checks if required PVM directories exist
func (afe *AutoFixEngine) checkMissingDirectories(ctx *FixContext) (bool, error) {
	xdgBinHome := os.Getenv("XDG_BIN_HOME")
	if xdgBinHome == "" {
		xdgBinHome = filepath.Join(ctx.HomeDir, ".local", "bin")
	}

	requiredDirs := []string{
		ctx.XDGDirs.DataDir,
		filepath.Join(ctx.XDGDirs.DataDir, "versions"),
		xdgBinHome,
	}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return true, nil // Needs fix
		}
	}

	return false, nil // All directories exist
}

// fixMissingDirectories creates missing PVM directories
func (afe *AutoFixEngine) fixMissingDirectories(ctx *FixContext) error {
	xdgBinHome := os.Getenv("XDG_BIN_HOME")
	if xdgBinHome == "" {
		xdgBinHome = filepath.Join(ctx.HomeDir, ".local", "bin")
	}

	requiredDirs := []string{
		ctx.XDGDirs.DataDir,
		filepath.Join(ctx.XDGDirs.DataDir, "versions"),
		xdgBinHome,
	}

	for _, dir := range requiredDirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return errors.NewSystemError("018", "Failed to create directory", err).
				WithLocation(dir)
		}
	}

	return nil
}

// checkShellIntegration checks if shell integration is properly configured
func (afe *AutoFixEngine) checkShellIntegration(ctx *FixContext) (bool, error) {
	hasIntegration, _, err := ctx.ShellMgr.CheckShellIntegration(ctx.Shell)
	return !hasIntegration, err // Needs fix if integration is missing
}

// fixShellIntegration adds shell integration to configuration
func (afe *AutoFixEngine) fixShellIntegration(ctx *FixContext) error {
	return ctx.ShellMgr.AddShellIntegration(ctx.Session, ctx.Shell, ctx.Force)
}

// checkRegistryIntegrity checks if the version registry needs rebuilding
func (afe *AutoFixEngine) checkRegistryIntegrity(ctx *FixContext) (bool, error) {
	// Use the same logic as needsRegistryRebuild from command_init.go
	registry, err := perl.LoadRegistry()
	if err != nil {
		return true, nil // Registry corrupted, needs rebuild
	}

	versionsDir := filepath.Join(ctx.XDGDirs.DataDir, "versions")

	// If registry is empty, check if versions exist on filesystem
	if len(registry.Versions) == 0 {
		if entries, err := os.ReadDir(versionsDir); err == nil && len(entries) > 0 {
			return true, nil // Registry empty but versions exist
		}
		return false, nil // Both empty
	}

	// Check for corrupted entries (test paths)
	validEntries := 0
	for _, versionInfo := range registry.Versions {
		if versionInfo.Source == "system" {
			validEntries++
			continue
		}

		// Skip corrupted test entries
		if strings.Contains(versionInfo.InstallPath, "/tmp/") ||
			strings.Contains(versionInfo.InstallPath, "pvm-shim-test") ||
			strings.Contains(versionInfo.InstallPath, "/var/folders/") {
			continue
		}

		if _, err := os.Stat(versionInfo.InstallPath); err == nil {
			validEntries++
		}
	}

	// If no valid entries but filesystem has installations, needs rebuild
	if validEntries == 0 {
		if entries, err := os.ReadDir(versionsDir); err == nil && len(entries) > 0 {
			return true, nil
		}
	}

	return false, nil // Registry is valid
}

// fixRegistryIntegrity rebuilds the version registry from filesystem
func (afe *AutoFixEngine) fixRegistryIntegrity(ctx *FixContext) error {
	return perl.RebuildRegistry()
}

// checkScriptPermissions checks if shell scripts have correct permissions
func (afe *AutoFixEngine) checkScriptPermissions(ctx *FixContext) (bool, error) {
	xdgBinHome := os.Getenv("XDG_BIN_HOME")
	if xdgBinHome == "" {
		xdgBinHome = filepath.Join(ctx.HomeDir, ".local", "bin")
	}

	// Check if XDG_BIN_HOME directory exists
	if _, err := os.Stat(xdgBinHome); os.IsNotExist(err) {
		return false, nil // Directory doesn't exist, skip permission check
	}

	// Check for any executable files that might need permission fixes
	files, err := os.ReadDir(xdgBinHome)
	if err != nil {
		return false, err
	}

	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(xdgBinHome, file.Name())
			info, err := os.Stat(filePath)
			if err != nil {
				continue
			}

			// Check if file should be executable but isn't
			if !strings.HasSuffix(file.Name(), ".txt") &&
				!strings.HasSuffix(file.Name(), ".md") &&
				info.Mode()&0111 == 0 { // Not executable
				return true, nil // Needs permission fix
			}
		}
	}

	return false, nil // Permissions are correct
}

// fixScriptPermissions fixes permissions on shell scripts
func (afe *AutoFixEngine) fixScriptPermissions(ctx *FixContext) error {
	xdgBinHome := os.Getenv("XDG_BIN_HOME")
	if xdgBinHome == "" {
		xdgBinHome = filepath.Join(ctx.HomeDir, ".local", "bin")
	}

	files, err := os.ReadDir(xdgBinHome)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(xdgBinHome, file.Name())

			// Make shell scripts and executable files executable
			if !strings.HasSuffix(file.Name(), ".txt") &&
				!strings.HasSuffix(file.Name(), ".md") {
				err := os.Chmod(filePath, 0755)
				if err != nil {
					return errors.NewSystemError("019", "Failed to fix file permissions", err).
						WithLocation(filePath)
				}
			}
		}
	}

	return nil
}
