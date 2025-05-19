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

	"tamarou.com/pvm/internal/errors"
)

// System Perl detection error codes
const (
	ErrNoSystemPerl     = "001" // No system Perl found
	ErrVersionParseFail = "002" // Failed to parse Perl version
	ErrPerlExecFail     = "003" // Failed to execute Perl
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
	// First, try to find perl in PATH
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

// findPerlInPath finds the perl executable in PATH
func findPerlInPath() (string, error) {
	// On Windows, we need to add .exe extension
	perlName := "perl"
	if runtime.GOOS == "windows" {
		perlName = "perl.exe"
	}

	perlPath, err := exec.LookPath(perlName)
	if err != nil {
		return "", errors.NewVersionError(ErrNoSystemPerl,
			"No system Perl found in PATH", err)
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
				"Failed to parse Perl version", nil).
				WithDetail(output)
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
