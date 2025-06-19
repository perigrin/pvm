// ABOUTME: Contains version information for the PVM ecosystem
// ABOUTME: Used across all components to provide consistent versioning

package version

import (
	"fmt"
	"runtime"
)

// Version is the current version of the PVM ecosystem
// This can be overridden at build time with ldflags
var Version = "0.1.0"

// BuildTime is set at build time via ldflags
var BuildTime = "unknown"

// CommitHash is set at build time via ldflags
var CommitHash = "unknown"

// GetVersion returns the current version string
func GetVersion() string {
	return Version
}

// GetFullVersion returns version with build information
func GetFullVersion() string {
	return fmt.Sprintf("%s (built %s from %s)", Version, BuildTime, CommitHash)
}

// ComponentVersion returns a formatted version string for a specific component
func ComponentVersion(component string) string {
	return component + " " + Version
}

// GetBuildInfo returns detailed build information
func GetBuildInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"buildTime":  BuildTime,
		"commitHash": CommitHash,
		"goVersion":  runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	}
}

// GetVersionDisplay returns formatted version information as a string
// This is used by CLI commands to display version info with UI styling
func GetVersionDisplay(component string) string {
	return fmt.Sprintf("%s %s\nBuild Time: %s\nCommit: %s\nGo Version: %s\nOS/Arch: %s/%s",
		component, Version, BuildTime, CommitHash, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
