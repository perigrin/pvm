// ABOUTME: plenv integration for robust Perl version management
// ABOUTME: Uses plenv commands directly instead of manual shim resolution

package perl

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"tamarou.com/pvm/internal/errors"
)

// plenv integration error codes
const (
	ErrPlenvNotAvailable  = "401" // plenv command not available
	ErrPlenvCommandFailed = "402" // plenv command execution failed
	ErrPlenvNoVersions    = "403" // no plenv versions available
)

// PlenvVersion represents a Perl version managed by plenv
type PlenvVersion struct {
	// Version string (e.g., "5.40.2" or "system")
	Version string

	// Path to the Perl executable for this version
	Path string

	// Is this the currently active version?
	IsActive bool

	// Is this the system Perl?
	IsSystem bool
}

// isPlenvAvailable checks if plenv command is available in PATH
func isPlenvAvailable() bool {
	_, err := exec.LookPath("plenv")
	return err == nil
}

// getPlenvVersions returns all Perl versions available through plenv
func getPlenvVersions() ([]PlenvVersion, error) {
	if !isPlenvAvailable() {
		return nil, errors.NewVersionError(
			ErrPlenvNotAvailable,
			"plenv command not found in PATH",
			nil)
	}

	// Run 'plenv versions' to get available versions
	cmd := exec.Command("plenv", "versions")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, errors.NewVersionError(
			ErrPlenvCommandFailed,
			"Failed to execute 'plenv versions'",
			err).WithDetail(stderr.String())
	}

	// Parse the output
	output := stdout.String()
	lines := strings.Split(output, "\n")
	var versions []PlenvVersion

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var version PlenvVersion

		// Check if this is the active version (marked with *)
		if strings.HasPrefix(line, "*") {
			version.IsActive = true
			line = strings.TrimSpace(strings.TrimPrefix(line, "*"))
		} else {
			// Remove leading whitespace
			line = strings.TrimSpace(line)
		}

		version.Version = line
		version.IsSystem = (line == "system")

		// Get the path to this version's perl executable
		perlPath, err := getPlenvPerlPath(version.Version)
		if err != nil {
			// Skip versions we can't resolve
			continue
		}
		version.Path = perlPath

		versions = append(versions, version)
	}

	if len(versions) == 0 {
		return nil, errors.NewVersionError(
			ErrPlenvNoVersions,
			"No valid Perl versions found in plenv",
			nil)
	}

	return versions, nil
}

// getPlenvPerlPath gets the path to perl for a specific plenv version
func getPlenvPerlPath(version string) (string, error) {
	if !isPlenvAvailable() {
		return "", errors.NewVersionError(
			ErrPlenvNotAvailable,
			"plenv command not found in PATH",
			nil)
	}

	// For system version, use 'plenv which perl'
	if version == "system" {
		return getPlenvSystemPerlPath()
	}

	// For other versions, construct path using plenv root
	plenvRoot := getPlenvRoot()
	if plenvRoot == "" {
		return "", fmt.Errorf("cannot determine plenv root")
	}

	perlPath := filepath.Join(plenvRoot, "versions", version, "bin", "perl")

	// Verify the perl executable exists
	if _, err := os.Stat(perlPath); os.IsNotExist(err) {
		return "", fmt.Errorf("perl executable not found for version %s: %s", version, perlPath)
	}

	return perlPath, nil
}

// getPlenvSystemPerlPath uses plenv to find the system perl path
func getPlenvSystemPerlPath() (string, error) {
	// Use 'plenv which perl' to get the actual perl path for system version
	cmd := exec.Command("plenv", "which", "perl")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set PLENV_VERSION=system to ensure we get system perl
	cmd.Env = append(os.Environ(), "PLENV_VERSION=system")

	err := cmd.Run()
	if err != nil {
		// If plenv which fails, fallback to direct system perl detection
		return findSystemPerlDirectly()
	}

	perlPath := strings.TrimSpace(stdout.String())
	if perlPath == "" {
		return findSystemPerlDirectly()
	}

	return perlPath, nil
}

// getCurrentPlenvVersion gets the currently active plenv version
func getCurrentPlenvVersion() (string, error) {
	if !isPlenvAvailable() {
		return "", errors.NewVersionError(
			ErrPlenvNotAvailable,
			"plenv command not found in PATH",
			nil)
	}

	// Check PLENV_VERSION environment variable first, but validate it
	if version := os.Getenv("PLENV_VERSION"); version != "" {
		// Validate that this version actually exists
		if isValidPlenvVersion(version) {
			return version, nil
		}
		// If invalid, ignore the env var and continue
	}

	// Run 'plenv version' to get current version
	cmd := exec.Command("plenv", "version")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Clear PLENV_VERSION temporarily to get the real default version
	env := os.Environ()
	var cleanEnv []string
	for _, e := range env {
		if !strings.HasPrefix(e, "PLENV_VERSION=") {
			cleanEnv = append(cleanEnv, e)
		}
	}
	cmd.Env = cleanEnv

	err := cmd.Run()
	if err != nil {
		return "", errors.NewVersionError(
			ErrPlenvCommandFailed,
			"Failed to execute 'plenv version'",
			err).WithDetail(stderr.String())
	}

	// Parse output like "5.40.2 (set by /home/user/.plenv/version)"
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return "", fmt.Errorf("no plenv version set")
	}

	// Extract just the version part
	parts := strings.Fields(output)
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid plenv version output: %s", output)
	}

	return parts[0], nil
}

// getPlenvRoot returns the plenv root directory
func getPlenvRoot() string {
	// Check PLENV_ROOT environment variable first
	if root := os.Getenv("PLENV_ROOT"); root != "" {
		return root
	}

	// Default to ~/.plenv
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return ""
	}

	return filepath.Join(homeDir, ".plenv")
}

// isValidPlenvVersion checks if a version is valid/available in plenv
func isValidPlenvVersion(version string) bool {
	if !isPlenvAvailable() {
		return false
	}

	// "system" is always valid
	if version == "system" {
		return true
	}

	// Check if the version directory exists
	plenvRoot := getPlenvRoot()
	if plenvRoot == "" {
		return false
	}

	versionDir := filepath.Join(plenvRoot, "versions", version)
	perlBin := filepath.Join(versionDir, "bin", "perl")

	_, err := os.Stat(perlBin)
	return err == nil
}

// resolvePlenvWithCommands uses plenv commands to resolve the current Perl executable
func resolvePlenvWithCommands() (string, error) {
	if !isPlenvAvailable() {
		return "", errors.NewVersionError(
			ErrPlenvNotAvailable,
			"plenv command not available",
			nil)
	}

	// Get current plenv version
	version, err := getCurrentPlenvVersion()
	if err != nil {
		return "", err
	}

	// Get the path for this version
	return getPlenvPerlPath(version)
}

// detectSystemPerlWithPlenv detects system Perl using plenv awareness
func detectSystemPerlWithPlenv() (*SystemPerl, error) {
	// If plenv is available, use it to find system perl
	if isPlenvAvailable() {
		perlPath, err := getPlenvSystemPerlPath()
		if err == nil {
			return extractPerlInfo(perlPath, true)
		}
	}

	// Fallback to direct system perl detection
	return detectSystemPerlDirectly()
}

// detectSystemPerlDirectly detects system perl without version managers
func detectSystemPerlDirectly() (*SystemPerl, error) {
	perlPath, err := findSystemPerlDirectly()
	if err != nil {
		return nil, err
	}

	return extractPerlInfo(perlPath, true)
}

// DiscoverAllPerlsWithPlenv discovers all available Perl installations using plenv
func DiscoverAllPerlsWithPlenv() ([]*SystemPerl, error) {
	var allPerls []*SystemPerl

	// If plenv is available, get all plenv-managed versions
	if isPlenvAvailable() {
		// Get the current active version first
		currentVersion, _ := getCurrentPlenvVersion()

		plenvVersions, err := getPlenvVersions()
		if err == nil {
			for _, pv := range plenvVersions {
				perl, err := extractPerlInfo(pv.Path, false) // Don't use pv.IsActive here
				if err != nil {
					// Skip versions that can't be processed
					continue
				}

				// Mark as primary if it's the current active plenv version
				perl.IsPrimary = (pv.Version == currentVersion)
				allPerls = append(allPerls, perl)
			}
		}
	}

	// Also find additional system perls not managed by plenv
	additionalPerls, err := findAdditionalPerls()
	if err == nil {
		// Filter out duplicates by path
		for _, additional := range additionalPerls {
			duplicate := false
			for _, existing := range allPerls {
				if existing.Path == additional.Path {
					duplicate = true
					break
				}
			}
			if !duplicate {
				allPerls = append(allPerls, additional)
			}
		}
	}

	return allPerls, nil
}
