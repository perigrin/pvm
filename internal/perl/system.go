// ABOUTME: System Perl detection functionality
// ABOUTME: Detects installed Perl versions on the system

package perl

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
)

// System Perl detection error codes
const (
	ErrNoSystemPerl                 = "001" // No system Perl found
	ErrVersionParseFail             = "002" // Failed to parse Perl version
	ErrPerlExecFail                 = "003" // Failed to execute Perl
	ErrSystemPerlRegistrationFailed = "004" // Failed to register system Perl
)

// SystemPerl represents information about a system-installed Perl
type SystemPerl struct {
	// Path to the Perl executable
	Path string

	// Version string (e.g., "5.34.0")
	Version string

	// Full version string including additional info
	FullVersion string

	// Architecture (e.g., "x86_64")
	Architecture string

	// Is this the primary system Perl?
	IsPrimary bool
}

// detectSystemPerlFunc is the actual function that detects the primary system Perl
func detectSystemPerlFunc() (*SystemPerl, error) {
	// Find system perl directly, bypassing any version managers
	perlPath, err := findPerlInPath()
	if err != nil {
		return nil, err
	}

	// Then, extract version info
	return extractPerlInfo(perlPath, true)
}

// DetectSystemPerl is a variable that points to detectSystemPerlFunc,
// allowing it to be replaced in tests
var DetectSystemPerl = detectSystemPerlFunc

// DetectAllSystemPerls detects all Perl installations on the system
func DetectAllSystemPerls() ([]*SystemPerl, error) {
	// If plenv is available, use plenv-aware discovery for comprehensive results
	if isPlenvAvailable() {
		perls, err := DiscoverAllPerlsWithPlenv()
		if err == nil && len(perls) > 0 {
			return perls, nil
		}
		// If plenv discovery fails, continue with fallback
	}

	// Fallback: traditional detection approach
	var perls []*SystemPerl

	// Get the primary Perl from PATH
	primaryPerl, err := DetectSystemPerl()
	if err == nil {
		perls = append(perls, primaryPerl)
	}

	// Look for other Perls in common locations
	// This is platform-specific
	additionalPerls, err := findAdditionalPerls()
	if err != nil {
		return perls, err
	}

	perls = append(perls, additionalPerls...)
	return perls, nil
}

// FindPerlByVersion finds a system Perl with the specified version
func FindPerlByVersion(version string) (*SystemPerl, error) {
	perls, err := DetectAllSystemPerls()
	if err != nil {
		return nil, err
	}

	for _, perl := range perls {
		if perl.Version == version {
			return perl, nil
		}
	}

	return nil, errors.NewVersionError(ErrNoSystemPerl,
		fmt.Sprintf("No system Perl with version %s found", version), nil)
}

// findPerlInPath finds the perl executable in system PATH, bypassing version managers
func findPerlInPath() (string, error) {
	// On Windows, we need to add .exe extension
	perlName := "perl"
	if runtime.GOOS == "windows" {
		perlName = "perl.exe"
	}

	// First try with clean system PATH to find actual system perl
	systemPaths := []string{"/usr/bin", "/usr/local/bin", "/bin"}
	if runtime.GOOS == "windows" {
		systemPaths = []string{"C:\\Perl\\bin", "C:\\Strawberry\\perl\\bin", "C:\\Windows\\System32"}
	}

	// Check system locations directly first
	for _, dir := range systemPaths {
		perlPath := filepath.Join(dir, perlName)
		if _, err := os.Stat(perlPath); err == nil {
			return perlPath, nil
		}
	}

	// Fallback: use exec.LookPath with clean PATH as a last resort
	// Save current environment
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Set clean system PATH
	cleanPath := strings.Join(systemPaths, string(os.PathListSeparator))
	os.Setenv("PATH", cleanPath)

	perlPath, err := exec.LookPath(perlName)
	if err != nil {
		return "", errors.NewVersionError(ErrNoSystemPerl,
			"No system Perl found in system locations", err)
	}

	// Get the absolute path
	perlPath, err = filepath.Abs(perlPath)
	if err != nil {
		return "", errors.NewVersionError(ErrNoSystemPerl,
			"Failed to get absolute path to Perl", err)
	}

	return perlPath, nil
}

// extractPerlInfo extracts version and architecture information from a Perl executable
func extractPerlInfo(perlPath string, isPrimary bool) (*SystemPerl, error) {
	// Run "perl -v" to get version info
	cmd := exec.Command(perlPath, "-v")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, errors.NewVersionError(ErrPerlExecFail,
			"Failed to execute Perl", err).
			WithDetail(stderr.String())
	}

	// Parse the output
	output := stdout.String()
	perl := &SystemPerl{
		Path:      perlPath,
		IsPrimary: isPrimary,
	}

	// Extract version information
	versionRegex := regexp.MustCompile(`This is perl (\d+), version (\d+), subversion (\d+)`)
	matches := versionRegex.FindStringSubmatch(output)
	if len(matches) >= 4 {
		perl.Version = fmt.Sprintf("%s.%s.%s", matches[1], matches[2], matches[3])
	} else {
		// Try alternate format
		altVersionRegex := regexp.MustCompile(`\(v(\d+\.\d+\.\d+)\)`)
		matches = altVersionRegex.FindStringSubmatch(output)
		if len(matches) >= 2 {
			perl.Version = matches[1]
		} else {
			return nil, errors.NewVersionError(ErrVersionParseFail,
				fmt.Sprintf("Failed to parse Perl version from output: '%s'", strings.TrimSpace(output)), nil).
				WithDetail(fmt.Sprintf("Full output: %s", output))
		}
	}

	// Extract full version string
	summaryRegex := regexp.MustCompile(`Summary of my perl5 \((.+?)\) configuration:`)
	matches = summaryRegex.FindStringSubmatch(output)
	if len(matches) >= 2 {
		perl.FullVersion = matches[1]
	} else {
		perl.FullVersion = perl.Version
	}

	// Extract architecture information
	archRegex := regexp.MustCompile(`Platform:\s+(\S+)`)
	matches = archRegex.FindStringSubmatch(output)
	if len(matches) >= 2 {
		perl.Architecture = matches[1]
	} else {
		// Try to detect platform from full version
		switch {
		case strings.Contains(perl.FullVersion, "x86_64"):
			perl.Architecture = "x86_64"
		case strings.Contains(perl.FullVersion, "amd64"):
			perl.Architecture = "amd64"
		case strings.Contains(perl.FullVersion, "i686"):
			perl.Architecture = "i686"
		case strings.Contains(perl.FullVersion, "arm64"):
			perl.Architecture = "arm64"
		default:
			// Default to runtime.GOARCH as a fallback
			perl.Architecture = runtime.GOARCH
		}
	}

	return perl, nil
}

// findAdditionalPerls finds additional Perl installations on the system
func findAdditionalPerls() ([]*SystemPerl, error) {
	var perls []*SystemPerl

	// Common paths to check based on platform
	var perlPaths []string

	switch runtime.GOOS {
	case "darwin":
		// macOS common locations
		perlPaths = []string{
			"/usr/local/bin/perl",
			"/usr/bin/perl",
			"/opt/homebrew/bin/perl",
		}
	case "linux":
		// Linux common locations
		perlPaths = []string{
			"/usr/bin/perl",
			"/usr/local/bin/perl",
			"/opt/perl/bin/perl",
		}
	case "windows":
		// Windows common locations
		perlPaths = []string{
			"C:\\Perl\\bin\\perl.exe",
			"C:\\Strawberry\\perl\\bin\\perl.exe",
			"C:\\Perl64\\bin\\perl.exe",
		}

		// Add Program Files paths
		programFiles := os.Getenv("ProgramFiles")
		if programFiles != "" {
			perlPaths = append(perlPaths,
				filepath.Join(programFiles, "Perl\\bin\\perl.exe"),
				filepath.Join(programFiles, "Strawberry\\perl\\bin\\perl.exe"))
		}

		programFilesX86 := os.Getenv("ProgramFiles(x86)")
		if programFilesX86 != "" {
			perlPaths = append(perlPaths,
				filepath.Join(programFilesX86, "Perl\\bin\\perl.exe"),
				filepath.Join(programFilesX86, "Strawberry\\perl\\bin\\perl.exe"))
		}
	}

	// Process each path
	for _, path := range perlPaths {
		// Check if the file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		// Extract perl info
		perl, err := extractPerlInfo(path, false)
		if err != nil {
			// Skip on error
			continue
		}

		// Only add if not already in the list
		duplicate := false
		for _, p := range perls {
			if p.Path == perl.Path {
				duplicate = true
				break
			}
		}

		if !duplicate {
			perls = append(perls, perl)
		}
	}

	return perls, nil
}

// GetCurrentPerlPath returns the path to the current Perl executable
func GetCurrentPerlPath() (string, error) {
	perl, err := DetectSystemPerl()
	if err != nil {
		return "", err
	}
	return perl.Path, nil
}

// GetSystemPerlVersion executes a Perl command and returns its version
func GetSystemPerlVersion(perlPath string) (string, error) {
	if perlPath == "" {
		// Find perl in PATH
		var err error
		perlPath, err = findPerlInPath()
		if err != nil {
			return "", err
		}
	}

	// Run perl -e 'print $^V' to get the version
	cmd := exec.Command(perlPath, "-e", "print $^V")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", errors.NewVersionError(ErrPerlExecFail,
			"Failed to execute Perl", err).
			WithDetail(stderr.String())
	}

	version := stdout.String()

	// If the output starts with 'v', remove it
	version = strings.TrimPrefix(version, "v")

	return version, nil
}

// resolvePerlVersionManager resolves plenv or perlbrew shims to actual Perl executable
func resolvePerlVersionManager(perlPath string) (string, error) {
	// Check if this looks like a PVM shim
	if strings.Contains(perlPath, "pvm/shims") {
		// For PVM shims, fall back to system Perl
		return resolveSystemPerl()
	}

	// Check if this looks like a plenv shim
	if strings.Contains(perlPath, ".plenv/shims") {
		// If plenv is available, use command-based resolution
		if isPlenvAvailable() {
			return resolvePlenvWithCommands()
		}
		// Fallback to manual shim resolution
		return resolvePlenvShim(perlPath)
	}

	// Check if this looks like a perlbrew perl
	if strings.Contains(perlPath, "perl5/perlbrew") {
		return resolvePerlbrewPerl(perlPath)
	}

	return "", fmt.Errorf("not a known perl version manager")
}

// resolveSystemPerl finds the actual system Perl executable, bypassing any version managers
func resolveSystemPerl() (string, error) {
	// Common system Perl locations
	systemPaths := []string{
		"/usr/bin/perl",
		"/usr/local/bin/perl",
		"/opt/local/bin/perl",    // MacPorts
		"/opt/homebrew/bin/perl", // Homebrew on Apple Silicon
		"/usr/pkg/bin/perl",      // NetBSD pkgsrc
	}

	for _, path := range systemPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no system Perl found in standard locations")
}

// resolvePlenvShim resolves a plenv shim to the actual Perl executable
func resolvePlenvShim(perlPath string) (string, error) {

	// Try to find the actual Perl that plenv would use
	// First check for plenv root and current version
	plenvRoot := os.Getenv("PLENV_ROOT")
	if plenvRoot == "" {
		// Default plenv root location
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			return "", fmt.Errorf("cannot determine plenv root")
		}
		plenvRoot = filepath.Join(homeDir, ".plenv")
	}

	// Try to determine the current plenv version
	version, err := getPlenvVersion(plenvRoot)
	if err != nil {
		// If we can't determine plenv version, try to find system perl directly
		return findSystemPerlDirectly()
	}

	// Special case for "system" version - find actual system perl
	if version == "system" {
		return findSystemPerlDirectly()
	}

	// Build path to the plenv-managed perl
	perlBin := filepath.Join(plenvRoot, "versions", version, "bin", "perl")
	if _, err := os.Stat(perlBin); err == nil {
		return perlBin, nil
	}

	// If plenv version doesn't exist, fallback to system perl
	return findSystemPerlDirectly()
}

// getPlenvVersion determines the current plenv version
func getPlenvVersion(plenvRoot string) (string, error) {
	// Check PLENV_VERSION environment variable first
	if version := os.Getenv("PLENV_VERSION"); version != "" {
		return version, nil
	}

	// Check for .perl-version file in current directory
	if version, err := readVersionFile(".perl-version"); err == nil {
		return version, nil
	}

	// Check plenv global version
	globalVersionFile := filepath.Join(plenvRoot, "version")
	if version, err := readVersionFile(globalVersionFile); err == nil {
		return version, nil
	}

	return "", fmt.Errorf("no plenv version configured")
}

// readVersionFile reads a version from a file
func readVersionFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	version := strings.TrimSpace(string(content))
	if version == "" {
		return "", fmt.Errorf("empty version file")
	}
	return version, nil
}

// findSystemPerlDirectly finds the system perl bypassing plenv
func findSystemPerlDirectly() (string, error) {
	// Common system perl locations
	systemPaths := []string{
		"/usr/bin/perl",
		"/usr/local/bin/perl",
		"/opt/perl/bin/perl",
	}

	for _, path := range systemPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no system perl found")
}

// resolvePerlbrewPerl resolves a perlbrew perl to handle perlbrew version management
func resolvePerlbrewPerl(perlPath string) (string, error) {
	// If it's already a direct perlbrew perl path, return it
	if strings.Contains(perlPath, "/bin/perl") {
		return perlPath, nil
	}

	// For perlbrew, the perl path is usually correct, but we might need to handle
	// the case where PERLBREW_PERL is set to something unavailable
	if perlbrewPerl := os.Getenv("PERLBREW_PERL"); perlbrewPerl != "" {
		// Try to find the specific perlbrew perl
		perlbrewRoot := os.Getenv("PERLBREW_ROOT")
		if perlbrewRoot == "" {
			homeDir := os.Getenv("HOME")
			if homeDir != "" {
				perlbrewRoot = filepath.Join(homeDir, "perl5", "perlbrew")
			}
		}

		if perlbrewRoot != "" {
			perlbrewBin := filepath.Join(perlbrewRoot, "perls", perlbrewPerl, "bin", "perl")
			if _, err := os.Stat(perlbrewBin); err == nil {
				return perlbrewBin, nil
			}
		}
	}

	// If we can't resolve perlbrew properly, fallback to system perl
	return findSystemPerlDirectly()
}

// AutoImportSystemPerl automatically imports system Perl into the registry
// This is used during first-run scenarios when no Perl versions are available
func AutoImportSystemPerl() error {
	// Detect system Perl
	systemPerl, err := DetectSystemPerl()
	if err != nil {
		return errors.NewVersionError(
			ErrNoSystemPerl,
			"Failed to detect system Perl for auto-import",
			err)
	}

	// Check if this version is already registered
	isRegistered, err := IsVersionInstalled(systemPerl.Version)
	if err != nil {
		return err
	}
	if isRegistered {
		// Already registered, nothing to do
		return nil
	}

	// Calculate proper install path - handle both /usr/bin/perl and /opt/perl/bin/perl cases
	installPath := filepath.Dir(systemPerl.Path)
	if filepath.Base(installPath) == "bin" {
		installPath = filepath.Dir(installPath)
	}

	// Create version info for registry
	versionInfo := VersionInfo{
		Version:     systemPerl.Version,
		InstallPath: installPath,
		Source:      "system",
		InstallTime: time.Now(),
	}

	// Register the version
	err = RegisterVersion(versionInfo)
	if err != nil {
		return errors.NewVersionError(
			ErrSystemPerlRegistrationFailed,
			fmt.Sprintf("Failed to register auto-imported system Perl %s", systemPerl.Version),
			err)
	}

	return nil
}
