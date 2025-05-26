// ABOUTME: Automatic dependency detection using PSC parser
// ABOUTME: Leverages PSC's Perl parsing to extract use/require statements

package pvx

import (
	"fmt"
	"os"
	"strings"

	"tamarou.com/pvm/internal/log"
)

// AutoDetectDependencies extracts dependencies from a Perl script using PSC
func AutoDetectDependencies(scriptPath string) ([]string, error) {
	// Read the script file
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read script file: %w", err)
	}

	// Extract dependencies from the script content using our own parser
	dependencies := extractDependenciesFromContent(string(content))

	if len(dependencies) > 0 {
		log.Debugf("Auto-detected %d dependencies from %s: %v", len(dependencies), scriptPath, dependencies)
	}

	return dependencies, nil
}

// extractDependenciesFromContent extracts module dependencies from Perl source code
// This mirrors the PSC ProjectAnalyzer.extractDependencies functionality
func extractDependenciesFromContent(content string) []string {
	var dependencies []string
	seen := make(map[string]bool)

	// Look for use statements
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Match use statements
		if strings.HasPrefix(line, "use ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				module := parts[1]
				// Remove version and semicolon
				module = strings.TrimSuffix(module, ";")
				if idx := strings.Index(module, " "); idx > 0 {
					module = module[:idx]
				}

				// Skip pragmas (lowercase words without ::)
				if !strings.Contains(module, "::") && strings.ToLower(module) == module {
					continue
				}

				// Skip core pragmas specifically
				if isPragma(module) {
					continue
				}

				if !seen[module] {
					dependencies = append(dependencies, module)
					seen[module] = true
				}
			}
		}

		// Match require statements
		if strings.HasPrefix(line, "require ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				module := strings.Trim(parts[1], "\"';")
				if strings.HasSuffix(module, ".pm") {
					module = strings.TrimSuffix(module, ".pm")
					module = strings.ReplaceAll(module, "/", "::")
				}

				// For require statements, accept both Module::Name and ModuleName forms
				if !seen[module] && (strings.Contains(module, "::") || (!isPragma(module) && module != "")) {
					dependencies = append(dependencies, module)
					seen[module] = true
				}
			}
		}
	}

	return dependencies
}

// isPragma checks if a module name is a core Perl pragma
func isPragma(module string) bool {
	pragmas := map[string]bool{
		"strict":      true,
		"warnings":    true,
		"utf8":        true,
		"feature":     true,
		"autodie":     true,
		"base":        true,
		"parent":      true,
		"constant":    true,
		"vars":        true,
		"lib":         true,
		"bigint":      true,
		"bignum":      true,
		"bigrat":      true,
		"integer":     true,
		"bytes":       true,
		"charnames":   true,
		"diagnostics": true,
		"encoding":    true,
		"fields":      true,
		"filetest":    true,
		"if":          true,
		"less":        true,
		"locale":      true,
		"open":        true,
		"ops":         true,
		"overload":    true,
		"re":          true,
		"sigtrap":     true,
		"sort":        true,
		"subs":        true,
		"threads":     true,
		"version":     true,
	}

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
