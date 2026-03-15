// ABOUTME: Automatic dependency detection for PVX
// ABOUTME: Extracts use/require statements using regex heuristics (AST-based detection not yet available)

package pvx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/perl"
)

// AutoDetectDependencies extracts dependencies from a Perl script using regex heuristics
func AutoDetectDependencies(scriptPath string) ([]string, error) {
	// Read the script file
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read script file: %w", err)
	}

	// Extract dependencies using regex-based parsing
	dependencies := extractDependenciesFromContent(string(content))

	if len(dependencies) > 0 {
		log.Debugf("Auto-detected %d dependencies from %s: %v", len(dependencies), scriptPath, dependencies)
	}

	return dependencies, nil
}

// AutoDetectAndStripTypedDependencies extracts dependencies; type stripping is not yet available
func AutoDetectAndStripTypedDependencies(scriptPath string, options *ExecutionOptions) ([]string, map[string]string, []func(), error) {
	// Detect dependencies without type stripping (type-system components not yet available)
	dependencies, err := AutoDetectDependencies(scriptPath)
	if err != nil {
		return nil, nil, nil, err
	}

	// Return empty strippedModulePaths since stripping is not yet available
	strippedModulePaths := make(map[string]string)
	return dependencies, strippedModulePaths, nil, nil
}

// extractDependenciesFromContent extracts module dependencies from Perl source code using regex heuristics
func extractDependenciesFromContent(content string) []string {
	var dependencies []string

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Match 'use Module' statements
		if strings.HasPrefix(line, "use ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				module := parts[1]
				// Remove trailing semicolons or other punctuation
				module = strings.TrimSuffix(module, ";")
				module = strings.TrimSuffix(module, "(")
				if module != "" && !strings.HasPrefix(module, "v") {
					dependencies = append(dependencies, module)
				}
			}
		}

		// Match require statements (with or without quotes)
		if strings.HasPrefix(line, "require ") {
			module := strings.TrimPrefix(line, "require ")
			module = strings.TrimSuffix(module, ";")
			module = strings.TrimSpace(module)
			module = strings.Trim(module, `"'`)
			if module != "" {
				dependencies = append(dependencies, normalizeRequiredModule(module))
			}
		}
	}

	return filterAndDeduplicateDependencies(dependencies)
}

// normalizeRequiredModule normalizes a required module name
func normalizeRequiredModule(module string) string {
	// Remove quotes
	module = strings.Trim(module, `"'`)

	// Convert .pm file paths to module names
	if strings.HasSuffix(module, ".pm") {
		module = strings.TrimSuffix(module, ".pm")
		module = strings.ReplaceAll(module, "/", "::")
	}

	// Skip empty modules
	if module == "" {
		return ""
	}

	return module
}

// filterAndDeduplicateDependencies removes duplicates and filters out Perl pragmas
func filterAndDeduplicateDependencies(deps []string) []string {
	seen := make(map[string]bool)
	var filtered []string

	// Use shared pragma list
	pragmas := getPragmaList()

	for _, dep := range deps {
		if dep != "" && !seen[dep] && !pragmas[dep] {
			seen[dep] = true
			filtered = append(filtered, dep)
		}
	}

	return filtered
}

// getPragmaList returns the map of Perl pragmas (shared between functions)
func getPragmaList() map[string]bool {
	return map[string]bool{
		"strict":       true,
		"warnings":     true,
		"utf8":         true,
		"feature":      true,
		"vars":         true,
		"lib":          true,
		"base":         true,
		"parent":       true,
		"constant":     true,
		"autodie":      true,
		"experimental": true,
		"bigint":       true,
		"bignum":       true,
		"bigrat":       true,
		"integer":      true,
		"bytes":        true,
		"charnames":    true,
		"diagnostics":  true,
		"encoding":     true,
		"fields":       true,
		"filetest":     true,
		"if":           true,
		"less":         true,
		"locale":       true,
		"open":         true,
		"ops":          true,
		"overload":     true,
		"re":           true,
		"sigtrap":      true,
		"sort":         true,
		"subs":         true,
		"threads":      true,
		"version":      true,
	}
}

// isPragma checks if a module name is a core Perl pragma
func isPragma(module string) bool {
	pragmas := getPragmaList()
	return pragmas[module]
}

// FilterCPANModules filters out modules that are likely core or built-in
func FilterCPANModules(dependencies []string) []string {
	var cpanModules []string

	coreModules := map[string]bool{
		// Core modules that ship with Perl
		"Carp":           true,
		"Data::Dumper":   true,
		"File::Basename": true,
		"File::Path":     true,
		"File::Spec":     true,
		"FindBin":        true,
		"Getopt::Long":   true,
		"IO::File":       true,
		"IO::Handle":     true,
		"List::Util":     true,
		"Scalar::Util":   true,
		"Time::Local":    true,
		"Time::Piece":    true,
		"POSIX":          true,
		"Storable":       true,
		"Socket":         true,
		"Fcntl":          true,
	}

	for _, dep := range dependencies {
		// Include if it's not a core module
		if !coreModules[dep] {
			cpanModules = append(cpanModules, dep)
		}
	}

	return cpanModules
}

// AutoDetectDependenciesWithOptions extracts dependencies with filtering options
func AutoDetectDependenciesWithOptions(scriptPath string, includeCoreModules bool) ([]string, error) {
	dependencies, err := AutoDetectDependencies(scriptPath)
	if err != nil {
		return nil, err
	}

	if !includeCoreModules {
		dependencies = FilterCPANModules(dependencies)
	}

	return dependencies, nil
}

// findModulePath attempts to locate a Perl module file in the filesystem
func findModulePath(moduleName string, options *ExecutionOptions) (string, error) {
	// Convert module name to file path (e.g., "My::Module" -> "My/Module.pm")
	filePath := strings.ReplaceAll(moduleName, "::", "/") + ".pm"

	// Get Perl version and resolve executable
	perlVersion := options.PerlVersion
	if perlVersion == "" {
		perlVersion = "system" // Default to system Perl
	}

	resolvedVersion, err := perl.ResolveVersion(&perl.ResolutionOptions{
		ExplicitVersion: perlVersion,
	})
	if err != nil {
		return "", fmt.Errorf("failed to resolve Perl version %s: %w", perlVersion, err)
	}

	// Get Perl's @INC directories
	incDirs, err := getPerlIncDirectories(resolvedVersion.Path)
	if err != nil {
		return "", fmt.Errorf("failed to get Perl @INC directories: %w", err)
	}

	// Add any additional module paths from options
	if len(options.AdditionalModulePaths) > 0 {
		incDirs = append(options.AdditionalModulePaths, incDirs...)
	}

	// Search for the module file in @INC directories
	for _, incDir := range incDirs {
		fullPath := filepath.Join(incDir, filePath)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("module %s not found in @INC directories", moduleName)
}

// getPerlIncDirectories gets the @INC directories from a Perl executable
func getPerlIncDirectories(perlPath string) ([]string, error) {
	// Use perl -E 'say for @INC' to get the include directories
	cmd := exec.Command(perlPath, "-E", "say for @INC")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute perl command: %w", err)
	}

	// Split output into lines and filter out empty lines
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var incDirs []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			incDirs = append(incDirs, line)
		}
	}

	return incDirs, nil
}
