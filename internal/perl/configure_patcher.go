// ABOUTME: Configure script patching for macOS compatibility
// ABOUTME: Implements various strategies to patch Perl Configure scripts for newer macOS versions

package perl

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"tamarou.com/pvm/internal/errors"
)

// ConfigurePatcher handles patching of Perl Configure scripts for macOS compatibility
type ConfigurePatcher struct {
	sourceDir    string
	perlVersion  string
	macOSVersion *MacOSVersion
	strategy     ConfigurePatchStrategy
	verbose      bool
}

// validateFilePath ensures the target path is within the expected source directory
func (cp *ConfigurePatcher) validateFilePath(path string) error {
	cleanPath := filepath.Clean(path)
	cleanSourceDir := filepath.Clean(cp.sourceDir)

	// Convert to absolute paths for proper comparison
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path for %s: %w", path, err)
	}

	absSourceDir, err := filepath.Abs(cleanSourceDir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path for source directory: %w", err)
	}

	// Check if path is within source directory
	if !strings.HasPrefix(absPath, absSourceDir) {
		return fmt.Errorf("path %s is outside expected source directory %s", path, absSourceDir)
	}

	return nil
}

// NewConfigurePatcher creates a new Configure patcher
func NewConfigurePatcher(sourceDir, perlVersion string, verbose bool) (*ConfigurePatcher, error) {
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("Configure patching only supported on macOS")
	}

	macOSVer, err := GetMacOSVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to detect macOS version: %w", err)
	}

	strategy := GetConfigurePatchStrategy(perlVersion)

	return &ConfigurePatcher{
		sourceDir:    sourceDir,
		perlVersion:  perlVersion,
		macOSVersion: macOSVer,
		strategy:     strategy,
		verbose:      verbose,
	}, nil
}

// ApplyPatches applies the appropriate patches based on the strategy
func (cp *ConfigurePatcher) ApplyPatches() error {
	if cp.strategy == PatchStrategyNone {
		if cp.verbose {
			fmt.Printf("No Configure patching needed for Perl %s on macOS %s\n",
				cp.perlVersion, cp.macOSVersion.String())
		}
		return nil
	}

	if cp.verbose {
		fmt.Printf("Applying Configure patches for Perl %s on macOS %s using strategy: %s\n",
			cp.perlVersion, cp.macOSVersion.String(), cp.getStrategyName())
	}

	switch cp.strategy {
	case PatchStrategyDarwinHints:
		return cp.patchDarwinHints()
	case PatchStrategyConfigureScript:
		return cp.patchConfigureScript()
	case PatchStrategyEnvironmentOverride:
		return cp.applyEnvironmentOverride()
	default:
		return fmt.Errorf("unknown patch strategy: %d", cp.strategy)
	}
}

// getStrategyName returns a human-readable name for the strategy
func (cp *ConfigurePatcher) getStrategyName() string {
	switch cp.strategy {
	case PatchStrategyDarwinHints:
		return "Darwin hints"
	case PatchStrategyConfigureScript:
		return "Configure script"
	case PatchStrategyEnvironmentOverride:
		return "Environment override"
	default:
		return "Unknown"
	}
}

// patchDarwinHints patches the hints/darwin.sh file to handle newer macOS versions
func (cp *ConfigurePatcher) patchDarwinHints() error {
	hintsFile := filepath.Join(cp.sourceDir, "hints", "darwin.sh")

	// Validate file path to prevent directory traversal
	if err := cp.validateFilePath(hintsFile); err != nil {
		return fmt.Errorf("invalid hints file path: %w", err)
	}

	// Check if hints file exists
	if _, err := os.Stat(hintsFile); os.IsNotExist(err) {
		// Fall back to Configure script patching if no hints file
		if cp.verbose {
			fmt.Printf("Darwin hints file not found, falling back to Configure script patching\n")
		}
		return cp.patchConfigureScript()
	}

	// Read hints file
	content, err := os.ReadFile(hintsFile)
	if err != nil {
		return fmt.Errorf("failed to read darwin.sh: %w", err)
	}

	originalContent := string(content)
	patchedContent := originalContent

	// Apply patches based on what's needed
	patches := cp.getDarwinHintsPatches()

	for _, patch := range patches {
		patchedContent = patch.Apply(patchedContent)
	}

	// Only write if content changed
	if patchedContent != originalContent {
		// Create backup
		backupFile := hintsFile + ".pvm-backup"
		if err := os.WriteFile(backupFile, content, 0644); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		// Write patched content
		if err := os.WriteFile(hintsFile, []byte(patchedContent), 0644); err != nil {
			return fmt.Errorf("failed to write patched darwin.sh: %w", err)
		}

		if cp.verbose {
			fmt.Printf("Patched darwin.sh (backup created at %s)\n", backupFile)
		}
	}

	return nil
}

// Patch represents a single patch operation
type Patch struct {
	Name        string
	Description string
	Pattern     *regexp.Regexp
	Replacement string
}

// Apply applies the patch to the content
func (p *Patch) Apply(content string) string {
	if p.Pattern.MatchString(content) {
		return p.Pattern.ReplaceAllString(content, p.Replacement)
	}
	return content
}

// getDarwinHintsPatches returns patches for darwin.sh
func (cp *ConfigurePatcher) getDarwinHintsPatches() []Patch {
	patches := []Patch{
		{
			Name:        "macOS version detection",
			Description: "Update version detection to handle macOS 11+ versions",
			Pattern:     regexp.MustCompile(`case "\$osvers" in\s*\n\s*10\.\d+.*?\n\s*;;`),
			Replacement: `case "$osvers" in
	[1-9][0-9]*.*|10.*)
		;;`,
		},
		{
			Name:        "version validation",
			Description: "Remove strict version validation that rejects macOS 11+",
			Pattern:     regexp.MustCompile(`\*\) echo "Unsupported Darwin version: \$osvers" >&2; exit 1 ;;\s*esac`),
			Replacement: `*) ;;
esac`,
		},
		{
			Name:        "macOS Big Sur compatibility",
			Description: "Add explicit support for macOS 11+ versions",
			Pattern:     regexp.MustCompile(`(osvers=.*\n.*sw_vers.*\n)`),
			Replacement: `$1
# Handle macOS Big Sur (11.x) and later
case "$osvers" in
	11.*|12.*|13.*|14.*|15.*|1[6-9].*|[2-9][0-9].*)
		# Use compatible settings for newer macOS versions
		osvers="$osvers"
		;;
esac
`,
		},
	}

	return patches
}

// patchConfigureScript directly patches the Configure script
func (cp *ConfigurePatcher) patchConfigureScript() error {
	configureFile := filepath.Join(cp.sourceDir, "Configure")

	// Validate file path to prevent directory traversal
	if err := cp.validateFilePath(configureFile); err != nil {
		return fmt.Errorf("invalid Configure script path: %w", err)
	}

	// Check if Configure script exists
	if _, err := os.Stat(configureFile); os.IsNotExist(err) {
		return fmt.Errorf("Configure script not found at %s", configureFile)
	}

	// Read Configure script
	content, err := os.ReadFile(configureFile)
	if err != nil {
		return fmt.Errorf("failed to read Configure script: %w", err)
	}

	originalContent := string(content)
	patchedContent := originalContent

	// Apply Configure script patches
	patches := cp.getConfigureScriptPatches()

	for _, patch := range patches {
		patchedContent = patch.Apply(patchedContent)
	}

	// Only write if content changed
	if patchedContent != originalContent {
		// Create backup
		backupFile := configureFile + ".pvm-backup"
		if err := os.WriteFile(backupFile, content, 0755); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		// Write patched content
		if err := os.WriteFile(configureFile, []byte(patchedContent), 0755); err != nil {
			return fmt.Errorf("failed to write patched Configure script: %w", err)
		}

		if cp.verbose {
			fmt.Printf("Patched Configure script (backup created at %s)\n", backupFile)
		}
	}

	return nil
}

// getConfigureScriptPatches returns patches for the Configure script
func (cp *ConfigurePatcher) getConfigureScriptPatches() []Patch {
	patches := []Patch{
		{
			Name:        "macOS version error bypass",
			Description: "Remove or bypass macOS version validation errors",
			Pattern:     regexp.MustCompile(`\*\*\* Unexpected product version.*?\n\*\*\*.*?\n.*?exit 1`),
			Replacement: `# PVM: Bypassed macOS version check for compatibility
echo "Note: macOS version compatibility enabled by PVM"`,
		},
		{
			Name:        "sw_vers product version handling",
			Description: "Handle newer macOS versions in sw_vers output processing",
			Pattern:     regexp.MustCompile(`(\$sw_vers -productVersion)`),
			Replacement: `$1 | sed 's/^\([0-9]*\)\.\([0-9]*\).*/\1.\2/'`,
		},
		{
			Name:        "Darwin version flexibility",
			Description: "Make Darwin version detection more flexible",
			Pattern:     regexp.MustCompile(`case.*\$osvers.*in.*\n.*10\..*\n.*\)\)`),
			Replacement: `case "$osvers" in
	[1-9][0-9]*.*|10.*)
		;;
	*)
		# Accept any version for macOS compatibility
		;;`,
		},
	}

	return patches
}

// applyEnvironmentOverride sets environment variables to override version detection
func (cp *ConfigurePatcher) applyEnvironmentOverride() error {
	// This strategy sets environment variables that Configure script will use
	// instead of detecting the actual macOS version

	// Set a compatible macOS version (10.15 Catalina is widely supported)
	os.Setenv("MACOSX_DEPLOYMENT_TARGET", "10.15")

	// Override version detection if sw_vers command is used
	os.Setenv("PERL_DARWIN_VERSION", "10.15.7")

	if cp.verbose {
		fmt.Printf("Set macOS compatibility environment variables:\n")
		fmt.Printf("  MACOSX_DEPLOYMENT_TARGET=10.15\n")
		fmt.Printf("  PERL_DARWIN_VERSION=10.15.7\n")
	}

	return nil
}

// RestoreBackups restores any backed up files
func (cp *ConfigurePatcher) RestoreBackups() error {
	var restoredFiles []string

	// Look for backup files
	backupPatterns := []string{
		filepath.Join(cp.sourceDir, "Configure.pvm-backup"),
		filepath.Join(cp.sourceDir, "hints", "darwin.sh.pvm-backup"),
	}

	for _, backupFile := range backupPatterns {
		if _, err := os.Stat(backupFile); err == nil {
			// Backup exists, restore it
			originalFile := strings.TrimSuffix(backupFile, ".pvm-backup")

			// Read backup content
			content, err := os.ReadFile(backupFile)
			if err != nil {
				return fmt.Errorf("failed to read backup %s: %w", backupFile, err)
			}

			// Restore original
			var perm fs.FileMode = 0644
			if strings.Contains(originalFile, "Configure") {
				perm = 0755
			}

			if err := os.WriteFile(originalFile, content, perm); err != nil {
				return fmt.Errorf("failed to restore %s: %w", originalFile, err)
			}

			// Remove backup
			if err := os.Remove(backupFile); err != nil {
				return fmt.Errorf("failed to remove backup %s: %w", backupFile, err)
			}

			restoredFiles = append(restoredFiles, originalFile)
		}
	}

	if len(restoredFiles) > 0 && cp.verbose {
		fmt.Printf("Restored backup files: %s\n", strings.Join(restoredFiles, ", "))
	}

	return nil
}

// ApplyMacOSConfigurePatches is a convenience function that creates a patcher and applies patches
func ApplyMacOSConfigurePatches(sourceDir, perlVersion string, verbose bool) error {
	if runtime.GOOS != "darwin" {
		return nil // No-op on non-macOS systems
	}

	patcher, err := NewConfigurePatcher(sourceDir, perlVersion, verbose)
	if err != nil {
		return errors.NewVersionError(
			ErrConfigureFailed,
			"Failed to create Configure patcher",
			err,
		)
	}

	err = patcher.ApplyPatches()
	if err != nil {
		return errors.NewVersionError(
			ErrConfigureFailed,
			"Failed to apply Configure patches",
			err,
		)
	}

	return nil
}
