// ABOUTME: Advanced router features for multi-entry point CLI
// ABOUTME: Provides debug, symlink detection, and fallback handling

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/version"
)

// Invocation types
const (
	InvocationSymlink  = "symlink"
	InvocationCopy     = "copy"
	InvocationDirect   = "direct"
	InvocationFallback = "fallback"
)

// InvocationInfo stores information about how the binary was invoked
type InvocationInfo struct {
	Component   string
	Type        string
	OriginalArg string
	BinaryPath  string
	Detected    bool
}

// DetectInvocation returns detailed information about how the binary was invoked
func DetectInvocation() *InvocationInfo {
	// Handle empty args gracefully
	if len(os.Args) == 0 {
		return &InvocationInfo{
			Component:   ComponentPVM,
			Type:        InvocationFallback,
			OriginalArg: "",
			BinaryPath:  "",
			Detected:    false,
		}
	}

	// Get the executable path from args
	exePath := os.Args[0]

	// Get the real path (resolving symlinks)
	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		// If we can't resolve symlinks, just use the original path
		realPath = exePath
	}

	// Get the base names
	exeName := filepath.Base(exePath)
	realName := filepath.Base(realPath)

	// Remove extensions (relevant for Windows)
	exeName = strings.TrimSuffix(exeName, filepath.Ext(exeName))
	realName = strings.TrimSuffix(realName, filepath.Ext(realName))

	// Initialize result
	info := &InvocationInfo{
		Component:   exeName,
		Type:        InvocationDirect,
		OriginalArg: exePath,
		BinaryPath:  realPath,
		Detected:    true,
	}

	// Check if this is a symlink
	if exePath != realPath {
		info.Type = InvocationSymlink
	} else if exeName != realName {
		// This could be a copy/rename
		info.Type = InvocationCopy
	}

	// Check if this is a known component
	switch exeName {
	case ComponentPVM, ComponentPVX, ComponentPM, ComponentPSC:
		// This is a known component, keep it
	case "pvi":
		// Backward compatibility: pvi -> pm
		info.Component = ComponentPM
	default:
		// Unknown component, fall back to PVM
		info.Component = ComponentPVM
		info.Type = InvocationFallback
		info.Detected = false
	}

	return info
}

// PrintDebugInfo prints detailed information about the binary invocation
// This is useful for debugging symlink and path issues
func PrintDebugInfo() {
	info := DetectInvocation()

	// Create UI instance for debug output
	output := ui.NewDefaultOutput()

	output.Header("PVM Ecosystem Debug Information")
	output.KeyValue(map[string]string{
		"Invoked as":      info.OriginalArg,
		"Resolved path":   info.BinaryPath,
		"Component":       info.Component,
		"Invocation type": info.Type,
		"Detected":        fmt.Sprintf("%v", info.Detected),
		"Working dir":     getWorkingDir(),
		"Version":         version.GetVersion(),
	})
}

// getWorkingDir returns the current working directory or an error message
func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return dir
}
