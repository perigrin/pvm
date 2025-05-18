// ABOUTME: Contains version information for the PVM ecosystem
// ABOUTME: Used across all components to provide consistent versioning

package version

// Version is the current version of the PVM ecosystem
const Version = "0.1.0"

// GetVersion returns the current version string
func GetVersion() string {
	return Version
}

// ComponentVersion returns a formatted version string for a specific component
func ComponentVersion(component string) string {
	return component + " " + Version
}
