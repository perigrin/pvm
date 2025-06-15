// ABOUTME: Cross-platform system Perl detection and automated installation management
// ABOUTME: Provides unified interface for managing system Perl installations across platforms

package perl

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"tamarou.com/pvm/internal/errors"
)

// System Perl installation error codes
const (
	ErrInstallationFailed  = "004" // Perl installation failed
	ErrUnsupportedPlatform = "005" // Platform not supported for installation
	ErrInstallationTimeout = "006" // Installation timeout
	ErrDependencyMissing   = "007" // Required dependency missing
)

// PerlDistribution represents different Perl distributions
type PerlDistribution int

const (
	DistributionSystem PerlDistribution = iota
	DistributionStrawberry
	DistributionActivePerl
	DistributionHomebrew
	DistributionApt
	DistributionYum
	DistributionDnf
	DistributionPacman
	DistributionZypper
	DistributionPerlBrew
	DistributionPlenv
)

// String returns the string representation of a PerlDistribution
func (d PerlDistribution) String() string {
	switch d {
	case DistributionSystem:
		return "system"
	case DistributionStrawberry:
		return "strawberry"
	case DistributionActivePerl:
		return "activeperl"
	case DistributionHomebrew:
		return "homebrew"
	case DistributionApt:
		return "apt"
	case DistributionYum:
		return "yum"
	case DistributionDnf:
		return "dnf"
	case DistributionPacman:
		return "pacman"
	case DistributionZypper:
		return "zypper"
	case DistributionPerlBrew:
		return "perlbrew"
	case DistributionPlenv:
		return "plenv"
	default:
		return "unknown"
	}
}

// SystemPerlManager manages system Perl installations
type SystemPerlManager struct {
	// Available distributions on current platform
	availableDistributions []PerlDistribution

	// Preferred distribution order
	preferredDistributions []PerlDistribution
}

// NewSystemPerlManager creates a new SystemPerlManager
func NewSystemPerlManager() *SystemPerlManager {
	manager := &SystemPerlManager{}
	manager.detectAvailableDistributions()
	manager.setPreferredDistributions()
	return manager
}

// detectAvailableDistributions detects what installation methods are available
func (m *SystemPerlManager) detectAvailableDistributions() {
	var available []PerlDistribution

	switch runtime.GOOS {
	case "windows":
		// Check for Windows package managers
		if m.hasCommand("choco") {
			available = append(available, DistributionStrawberry)
		}
		if m.hasCommand("scoop") {
			available = append(available, DistributionStrawberry)
		}
		if m.hasCommand("winget") {
			available = append(available, DistributionStrawberry)
		}

	case "darwin":
		// Check for macOS package managers
		if m.hasCommand("brew") {
			available = append(available, DistributionHomebrew)
		}

	case "linux":
		// Check for Linux package managers
		if m.hasCommand("apt") || m.hasCommand("apt-get") {
			available = append(available, DistributionApt)
		}
		if m.hasCommand("yum") {
			available = append(available, DistributionYum)
		}
		if m.hasCommand("dnf") {
			available = append(available, DistributionDnf)
		}
		if m.hasCommand("pacman") {
			available = append(available, DistributionPacman)
		}
		if m.hasCommand("zypper") {
			available = append(available, DistributionZypper)
		}
	}

	// Cross-platform tools
	if m.hasCommand("perlbrew") {
		available = append(available, DistributionPerlBrew)
	}
	if m.hasCommand("plenv") {
		available = append(available, DistributionPlenv)
	}

	m.availableDistributions = available
}

// setPreferredDistributions sets the preferred installation order
func (m *SystemPerlManager) setPreferredDistributions() {
	switch runtime.GOOS {
	case "windows":
		m.preferredDistributions = []PerlDistribution{
			DistributionStrawberry,
		}
	case "darwin":
		m.preferredDistributions = []PerlDistribution{
			DistributionHomebrew,
			DistributionPerlBrew,
		}
	case "linux":
		m.preferredDistributions = []PerlDistribution{
			DistributionApt,
			DistributionDnf,
			DistributionYum,
			DistributionPacman,
			DistributionZypper,
			DistributionPerlBrew,
		}
	}
}

// hasCommand checks if a command is available in PATH
func (m *SystemPerlManager) hasCommand(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// DetectOrInstallPerl detects system Perl or installs it if not available
func (m *SystemPerlManager) DetectOrInstallPerl() (*SystemPerl, error) {
	// First try to detect existing system Perl
	perl, err := DetectSystemPerl()
	if err == nil {
		return perl, nil
	}

	// If no system Perl found, try to install one
	return m.InstallSystemPerl()
}

// InstallSystemPerl installs Perl using the best available method
func (m *SystemPerlManager) InstallSystemPerl() (*SystemPerl, error) {
	// Find the best available distribution
	var distribution PerlDistribution
	var found bool

	for _, preferred := range m.preferredDistributions {
		for _, available := range m.availableDistributions {
			if preferred == available {
				distribution = preferred
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return nil, errors.NewVersionError(ErrUnsupportedPlatform,
			fmt.Sprintf("No supported Perl installation method found for %s", runtime.GOOS), nil)
	}

	return m.installWithDistribution(distribution)
}

// installWithDistribution installs Perl using the specified distribution method
func (m *SystemPerlManager) installWithDistribution(dist PerlDistribution) (*SystemPerl, error) {
	switch dist {
	case DistributionHomebrew:
		return m.installWithHomebrew()
	case DistributionStrawberry:
		return m.installStrawberryPerl()
	case DistributionApt:
		return m.installWithApt()
	case DistributionDnf:
		return m.installWithDnf()
	case DistributionYum:
		return m.installWithYum()
	case DistributionPacman:
		return m.installWithPacman()
	case DistributionZypper:
		return m.installWithZypper()
	case DistributionPerlBrew:
		return m.installWithPerlBrew()
	case DistributionPlenv:
		return m.installWithPlenv()
	default:
		return nil, errors.NewVersionError(ErrUnsupportedPlatform,
			fmt.Sprintf("Installation method %s not implemented", dist.String()), nil)
	}
}

// installWithHomebrew installs Perl using Homebrew on macOS
func (m *SystemPerlManager) installWithHomebrew() (*SystemPerl, error) {
	cmd := exec.Command("brew", "install", "perl")
	err := cmd.Run()
	if err != nil {
		return nil, errors.NewVersionError(ErrInstallationFailed,
			"Failed to install Perl with Homebrew", err)
	}

	// Detect the newly installed Perl
	return DetectSystemPerl()
}

// installStrawberryPerl installs Strawberry Perl on Windows
func (m *SystemPerlManager) installStrawberryPerl() (*SystemPerl, error) {
	// Try different package managers
	if m.hasCommand("choco") {
		cmd := exec.Command("choco", "install", "strawberryperl", "-y")
		err := cmd.Run()
		if err == nil {
			return DetectSystemPerl()
		}
	}

	if m.hasCommand("scoop") {
		cmd := exec.Command("scoop", "install", "perl")
		err := cmd.Run()
		if err == nil {
			return DetectSystemPerl()
		}
	}

	if m.hasCommand("winget") {
		cmd := exec.Command("winget", "install", "StrawberryPerl.StrawberryPerl")
		err := cmd.Run()
		if err == nil {
			return DetectSystemPerl()
		}
	}

	return nil, errors.NewVersionError(ErrInstallationFailed,
		"Failed to install Strawberry Perl on Windows", nil)
}

// installWithApt installs Perl using apt on Debian/Ubuntu
func (m *SystemPerlManager) installWithApt() (*SystemPerl, error) {
	// Update package list first
	updateCmd := exec.Command("sudo", "apt", "update")
	err := updateCmd.Run()
	if err != nil {
		return nil, errors.NewVersionError(ErrInstallationFailed,
			"Failed to update package list with apt", err)
	}

	// Install perl
	cmd := exec.Command("sudo", "apt", "install", "-y", "perl")
	err = cmd.Run()
	if err != nil {
		return nil, errors.NewVersionError(ErrInstallationFailed,
			"Failed to install Perl with apt", err)
	}

	return DetectSystemPerl()
}

// installWithDnf installs Perl using dnf on Fedora/RHEL 8+
func (m *SystemPerlManager) installWithDnf() (*SystemPerl, error) {
	cmd := exec.Command("sudo", "dnf", "install", "-y", "perl")
	err := cmd.Run()
	if err != nil {
		return nil, errors.NewVersionError(ErrInstallationFailed,
			"Failed to install Perl with dnf", err)
	}

	return DetectSystemPerl()
}

// installWithYum installs Perl using yum on RHEL/CentOS
func (m *SystemPerlManager) installWithYum() (*SystemPerl, error) {
	cmd := exec.Command("sudo", "yum", "install", "-y", "perl")
	err := cmd.Run()
	if err != nil {
		return nil, errors.NewVersionError(ErrInstallationFailed,
			"Failed to install Perl with yum", err)
	}

	return DetectSystemPerl()
}

// installWithPacman installs Perl using pacman on Arch Linux
func (m *SystemPerlManager) installWithPacman() (*SystemPerl, error) {
	cmd := exec.Command("sudo", "pacman", "-S", "--noconfirm", "perl")
	err := cmd.Run()
	if err != nil {
		return nil, errors.NewVersionError(ErrInstallationFailed,
			"Failed to install Perl with pacman", err)
	}

	return DetectSystemPerl()
}

// installWithZypper installs Perl using zypper on openSUSE
func (m *SystemPerlManager) installWithZypper() (*SystemPerl, error) {
	cmd := exec.Command("sudo", "zypper", "install", "-y", "perl")
	err := cmd.Run()
	if err != nil {
		return nil, errors.NewVersionError(ErrInstallationFailed,
			"Failed to install Perl with zypper", err)
	}

	return DetectSystemPerl()
}

// installWithPerlBrew installs Perl using perlbrew
func (m *SystemPerlManager) installWithPerlBrew() (*SystemPerl, error) {
	// Install latest stable perl
	cmd := exec.Command("perlbrew", "install", "perl-stable")
	err := cmd.Run()
	if err != nil {
		return nil, errors.NewVersionError(ErrInstallationFailed,
			"Failed to install Perl with perlbrew", err)
	}

	// Switch to the newly installed perl
	switchCmd := exec.Command("perlbrew", "switch", "perl-stable")
	err = switchCmd.Run()
	if err != nil {
		return nil, errors.NewVersionError(ErrInstallationFailed,
			"Failed to switch to newly installed Perl with perlbrew", err)
	}

	return DetectSystemPerl()
}

// installWithPlenv installs Perl using plenv
func (m *SystemPerlManager) installWithPlenv() (*SystemPerl, error) {
	// Get latest stable version
	listCmd := exec.Command("plenv", "install", "--list")
	output, err := listCmd.Output()
	if err != nil {
		return nil, errors.NewVersionError(ErrInstallationFailed,
			"Failed to list available Perl versions with plenv", err)
	}

	// Parse output to find latest stable version
	lines := strings.Split(string(output), "\n")
	var latestVersion string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "5.") && !strings.Contains(line, "RC") && !strings.Contains(line, "TRIAL") {
			latestVersion = line
		}
	}

	if latestVersion == "" {
		return nil, errors.NewVersionError(ErrInstallationFailed,
			"No suitable Perl version found for plenv installation", nil)
	}

	// Install the version
	cmd := exec.Command("plenv", "install", latestVersion)
	err = cmd.Run()
	if err != nil {
		return nil, errors.NewVersionError(ErrInstallationFailed,
			fmt.Sprintf("Failed to install Perl %s with plenv", latestVersion), err)
	}

	// Set as global version
	globalCmd := exec.Command("plenv", "global", latestVersion)
	err = globalCmd.Run()
	if err != nil {
		return nil, errors.NewVersionError(ErrInstallationFailed,
			fmt.Sprintf("Failed to set Perl %s as global with plenv", latestVersion), err)
	}

	return DetectSystemPerl()
}

// GetAvailableDistributions returns the list of available distributions
func (m *SystemPerlManager) GetAvailableDistributions() []PerlDistribution {
	return m.availableDistributions
}

// GetPreferredDistributions returns the preferred distribution order
func (m *SystemPerlManager) GetPreferredDistributions() []PerlDistribution {
	return m.preferredDistributions
}

// ValidateInstallation validates that a Perl installation is working
func (m *SystemPerlManager) ValidateInstallation(perl *SystemPerl) error {
	if perl == nil {
		return errors.NewVersionError(ErrPerlExecFail,
			"Cannot validate nil Perl installation", nil)
	}

	// Check if the executable exists
	if _, err := os.Stat(perl.Path); os.IsNotExist(err) {
		return errors.NewVersionError(ErrPerlExecFail,
			fmt.Sprintf("Perl executable not found at %s", perl.Path), err)
	}

	// Try to run a simple perl command
	cmd := exec.Command(perl.Path, "-e", "print 'OK'")
	output, err := cmd.Output()
	if err != nil {
		return errors.NewVersionError(ErrPerlExecFail,
			"Perl executable failed basic test", err)
	}

	if string(output) != "OK" {
		return errors.NewVersionError(ErrPerlExecFail,
			fmt.Sprintf("Perl executable produced unexpected output: %s", string(output)), nil)
	}

	return nil
}

// CheckForUpdates checks if there are updates available for system Perl
func (m *SystemPerlManager) CheckForUpdates() (bool, error) {
	// This is a placeholder - actual implementation would depend on the distribution
	// For now, we'll return false (no updates available)
	return false, nil
}
