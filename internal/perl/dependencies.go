// ABOUTME: System dependency checking for Perl builds
// ABOUTME: Ensures required build tools and libraries are available

package perl

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// DependencyInfo contains information about system dependencies
type DependencyInfo struct {
	Required    []string          // Required dependencies that are present
	Missing     []string          // Required dependencies that are missing
	Optional    []string          // Optional dependencies that are missing
	InstallHint map[string]string // Platform-specific install hints
}

// DependencyChecker checks for system dependencies
type DependencyChecker struct {
	cache *DependencyInfo
}

// NewDependencyChecker creates a new dependency checker
func NewDependencyChecker() (*DependencyChecker, error) {
	return &DependencyChecker{}, nil
}

// CheckBuildDependencies checks for required build dependencies
func (dc *DependencyChecker) CheckBuildDependencies() (*DependencyInfo, error) {
	// Use cached result if available
	if dc.cache != nil {
		return dc.cache, nil
	}

	info := &DependencyInfo{
		Required:    []string{},
		Missing:     []string{},
		Optional:    []string{},
		InstallHint: make(map[string]string),
	}

	// Check required dependencies
	required := dc.getRequiredDependencies()
	for _, dep := range required {
		if dc.checkCommand(dep.command) {
			info.Required = append(info.Required, dep.name)
		} else {
			info.Missing = append(info.Missing, dep.name)
			if hint := dc.getInstallHint(dep.name); hint != "" {
				info.InstallHint[dep.name] = hint
			}
		}
	}

	// Check optional dependencies
	optional := dc.getOptionalDependencies()
	for _, dep := range optional {
		if !dc.checkCommand(dep.command) {
			info.Optional = append(info.Optional, dep.name)
			if hint := dc.getInstallHint(dep.name); hint != "" {
				info.InstallHint[dep.name] = hint
			}
		}
	}

	// Check for specific libraries
	if err := dc.checkLibraries(info); err != nil {
		return nil, err
	}

	dc.cache = info
	return info, nil
}

// dependency represents a system dependency
type dependency struct {
	name    string
	command string
}

// getRequiredDependencies returns required build dependencies
func (dc *DependencyChecker) getRequiredDependencies() []dependency {
	deps := []dependency{
		{name: "make", command: "make"},
		{name: "gcc", command: "gcc"},
		{name: "tar", command: "tar"},
		{name: "gzip", command: "gzip"},
	}

	// Platform-specific dependencies
	switch runtime.GOOS {
	case "darwin":
		// macOS uses clang instead of gcc
		deps[1] = dependency{name: "clang", command: "clang"}
		deps = append(deps, dependency{name: "xcode-select", command: "xcode-select"})
	case "linux":
		deps = append(deps,
			dependency{name: "patch", command: "patch"},
			dependency{name: "bzip2", command: "bzip2"},
		)
	case "windows":
		// Windows has different requirements
		deps = []dependency{
			{name: "nmake", command: "nmake"},
			{name: "cl", command: "cl"},
		}
	}

	return deps
}

// getOptionalDependencies returns optional build dependencies
func (dc *DependencyChecker) getOptionalDependencies() []dependency {
	deps := []dependency{
		{name: "git", command: "git"},
		{name: "wget", command: "wget"},
		{name: "curl", command: "curl"},
		{name: "xz", command: "xz"},
	}

	// Development tools for optional features
	if runtime.GOOS != "windows" {
		deps = append(deps,
			dependency{name: "valgrind", command: "valgrind"},
			dependency{name: "gdb", command: "gdb"},
		)
	}

	return deps
}

// checkCommand checks if a command is available in PATH
func (dc *DependencyChecker) checkCommand(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// checkLibraries checks for required libraries
func (dc *DependencyChecker) checkLibraries(info *DependencyInfo) error {
	switch runtime.GOOS {
	case "linux":
		// Check for development headers
		libraries := []struct {
			name  string
			files []string
		}{
			{
				name:  "libc-dev",
				files: []string{"/usr/include/stdio.h", "/usr/include/stdlib.h"},
			},
			{
				name:  "zlib-dev",
				files: []string{"/usr/include/zlib.h", "/usr/include/x86_64-linux-gnu/zlib.h"},
			},
		}

		for _, lib := range libraries {
			found := false
			for _, file := range lib.files {
				if _, err := os.Stat(file); err == nil {
					found = true
					break
				}
			}
			if !found {
				info.Missing = append(info.Missing, lib.name)
				if hint := dc.getInstallHint(lib.name); hint != "" {
					info.InstallHint[lib.name] = hint
				}
			}
		}

	case "darwin":
		// Check for Xcode command line tools
		cmd := exec.Command("xcode-select", "-p")
		if err := cmd.Run(); err != nil {
			info.Missing = append(info.Missing, "xcode-command-line-tools")
			info.InstallHint["xcode-command-line-tools"] = "Run: xcode-select --install"
		}
	}

	return nil
}

// getInstallHint returns platform-specific installation hints
func (dc *DependencyChecker) getInstallHint(dep string) string {
	hints := map[string]map[string]string{
		"darwin": {
			"xcode-command-line-tools": "xcode-select --install",
			"make":                     "Install Xcode Command Line Tools",
			"clang":                    "Install Xcode Command Line Tools",
			"git":                      "brew install git",
			"wget":                     "brew install wget",
			"xz":                       "brew install xz",
		},
		"linux": {
			"gcc":      "apt-get install build-essential (Debian/Ubuntu) or yum install gcc (RHEL/CentOS)",
			"make":     "apt-get install make (Debian/Ubuntu) or yum install make (RHEL/CentOS)",
			"libc-dev": "apt-get install libc6-dev (Debian/Ubuntu) or yum install glibc-devel (RHEL/CentOS)",
			"zlib-dev": "apt-get install zlib1g-dev (Debian/Ubuntu) or yum install zlib-devel (RHEL/CentOS)",
			"patch":    "apt-get install patch (Debian/Ubuntu) or yum install patch (RHEL/CentOS)",
			"git":      "apt-get install git (Debian/Ubuntu) or yum install git (RHEL/CentOS)",
			"wget":     "apt-get install wget (Debian/Ubuntu) or yum install wget (RHEL/CentOS)",
			"xz":       "apt-get install xz-utils (Debian/Ubuntu) or yum install xz (RHEL/CentOS)",
		},
		"windows": {
			"nmake": "Install Visual Studio or Build Tools for Visual Studio",
			"cl":    "Install Visual Studio or Build Tools for Visual Studio",
			"git":   "Download from https://git-scm.com/download/win",
			"wget":  "Install from chocolatey: choco install wget",
		},
	}

	if platformHints, ok := hints[runtime.GOOS]; ok {
		if hint, ok := platformHints[dep]; ok {
			return hint
		}
	}

	// Check for common package managers
	if runtime.GOOS == "linux" {
		// Detect distribution
		if _, err := os.Stat("/etc/debian_version"); err == nil {
			return fmt.Sprintf("Try: apt-get install %s", dep)
		} else if _, err := os.Stat("/etc/redhat-release"); err == nil {
			return fmt.Sprintf("Try: yum install %s", dep)
		} else if _, err := os.Stat("/etc/arch-release"); err == nil {
			return fmt.Sprintf("Try: pacman -S %s", dep)
		}
	}

	return ""
}

// CheckModuleDependencies checks for Perl module dependencies
func (dc *DependencyChecker) CheckModuleDependencies(modules []string) (*ModuleDependencyInfo, error) {
	info := &ModuleDependencyInfo{
		Installed: []string{},
		Missing:   []string{},
		Outdated:  []string{},
	}

	// This would check for required Perl modules
	// For now, return empty info
	return info, nil
}

// ModuleDependencyInfo contains information about Perl module dependencies
type ModuleDependencyInfo struct {
	Installed []string
	Missing   []string
	Outdated  []string
}

// GetPlatformSpecificNotes returns platform-specific build notes
func GetPlatformSpecificNotes() string {
	switch runtime.GOOS {
	case "darwin":
		return strings.Join([]string{
			"macOS Build Notes:",
			"- Ensure Xcode Command Line Tools are installed: xcode-select --install",
			"- For Apple Silicon (M1/M2), builds will use arm64 architecture",
			"- Consider using Homebrew for additional dependencies",
		}, "\n")

	case "linux":
		return strings.Join([]string{
			"Linux Build Notes:",
			"- Ensure development headers are installed (build-essential on Debian/Ubuntu)",
			"- SELinux may require additional configuration",
			"- Consider installing optional dependencies for full functionality",
		}, "\n")

	case "windows":
		return strings.Join([]string{
			"Windows Build Notes:",
			"- Visual Studio or Build Tools for Visual Studio required",
			"- Consider using WSL2 for a more Unix-like build environment",
			"- Native Windows builds may have limited functionality",
		}, "\n")

	default:
		return "Platform-specific build notes not available"
	}
}
