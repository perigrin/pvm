// ABOUTME: Carton-specific types and cpanfile/snapshot parsing
// ABOUTME: Provides parsing functionality for cpanfile and cpanfile.snapshot formats

package cpan

import (
	"bufio"
	"regexp"
	"strings"
)

// CPANFile represents a parsed cpanfile
type CPANFile struct {
	// Requirements are the module requirements
	Requirements []Requirement

	// Features are optional features
	Features map[string][]Requirement

	// Platforms are platform-specific requirements
	Platforms map[string][]Requirement
}

// Requirement represents a module requirement
type Requirement struct {
	Module       string
	Version      string
	Phase        string
	Relationship string
}

// CPANSnapshot represents a parsed cpanfile.snapshot
type CPANSnapshot struct {
	// Modules are the locked module versions
	Modules map[string]SnapshotModule
}

// SnapshotModule represents a module in the snapshot
type SnapshotModule struct {
	Version      string
	Distribution string
	Path         string
	Dependencies []Dependency
}

// ParseCPANFile parses cpanfile content
func ParseCPANFile(content string) (*CPANFile, error) {
	cpanfile := &CPANFile{
		Requirements: []Requirement{},
		Features:     make(map[string][]Requirement),
		Platforms:    make(map[string][]Requirement),
	}

	// Enhanced regex patterns for different cpanfile constructs
	requiresRe := regexp.MustCompile(`(?:^|\s+)requires\s+'([^']+)'(?:\s*,\s*['"]([^'"]+)['"])?`)
	recommendsRe := regexp.MustCompile(`(?:^|\s+)recommends\s+'([^']+)'(?:\s*,\s*['"]([^'"]+)['"])?`)
	suggestsRe := regexp.MustCompile(`(?:^|\s+)suggests\s+'([^']+)'(?:\s*,\s*['"]([^'"]+)['"])?`)
	testRequiresRe := regexp.MustCompile(`(?:^|\s+)test_requires\s+'([^']+)'(?:\s*,\s*['"]([^'"]+)['"])?`)
	buildRequiresRe := regexp.MustCompile(`(?:^|\s+)build_requires\s+'([^']+)'(?:\s*,\s*['"]([^'"]+)['"])?`)

	// Feature and platform parsing
	featureRe := regexp.MustCompile(`feature\s+'([^']+)'`)
	onRe := regexp.MustCompile(`on\s+'([^']+)'`)

	var currentFeature string
	var currentPlatform string
	var currentPhase string = "runtime"

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle feature blocks
		if matches := featureRe.FindStringSubmatch(line); matches != nil {
			currentFeature = matches[1]
			if cpanfile.Features[currentFeature] == nil {
				cpanfile.Features[currentFeature] = []Requirement{}
			}
			continue
		}

		// Handle platform/phase blocks
		if matches := onRe.FindStringSubmatch(line); matches != nil {
			target := matches[1]
			// Check if it's a phase (test, build, runtime) vs platform (MSWin32, etc.)
			switch target {
			case "test", "build", "runtime":
				currentPhase = target
			default:
				currentPlatform = target
				if cpanfile.Platforms[currentPlatform] == nil {
					cpanfile.Platforms[currentPlatform] = []Requirement{}
				}
			}
			continue
		}

		// Parse different requirement types
		patterns := []struct {
			regex        *regexp.Regexp
			relationship string
		}{
			{requiresRe, "requires"},
			{recommendsRe, "recommends"},
			{suggestsRe, "suggests"},
			{testRequiresRe, "test_requires"},
			{buildRequiresRe, "build_requires"},
		}

		for _, pattern := range patterns {
			if matches := pattern.regex.FindStringSubmatch(line); matches != nil {
				req := Requirement{
					Module:       matches[1],
					Relationship: pattern.relationship,
					Phase:        currentPhase,
				}
				if len(matches) > 2 && matches[2] != "" {
					req.Version = matches[2]
				}

				// Determine phase for test_requires and build_requires
				switch pattern.relationship {
				case "test_requires":
					req.Phase = "test"
				case "build_requires":
					req.Phase = "build"
				}

				// Add to appropriate collection
				switch {
				case currentFeature != "":
					cpanfile.Features[currentFeature] = append(cpanfile.Features[currentFeature], req)
				case currentPlatform != "":
					cpanfile.Platforms[currentPlatform] = append(cpanfile.Platforms[currentPlatform], req)
				default:
					cpanfile.Requirements = append(cpanfile.Requirements, req)
				}
				break
			}
		}

		// Reset context on block end
		if strings.Contains(line, "};") {
			currentFeature = ""
			currentPlatform = ""
			currentPhase = "runtime"
		}
	}

	return cpanfile, nil
}

// ParseCPANSnapshot parses cpanfile.snapshot content
func ParseCPANSnapshot(content string) (*CPANSnapshot, error) {
	snapshot := &CPANSnapshot{
		Modules: make(map[string]SnapshotModule),
	}

	lines := strings.Split(content, "\n")
	var currentDist string

	// State tracking for snapshot parsing
	inDistribution := false
	inProvides := false

	for _, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Distribution header: DISTRIBUTIONS
		if line == "DISTRIBUTIONS" {
			inDistribution = true
			continue
		}

		// Provides section
		if strings.HasPrefix(originalLine, "    provides:") {
			inProvides = true
			continue
		}

		// Reset provides when we hit a new section
		if inDistribution && !strings.HasPrefix(originalLine, "    ") && !strings.HasPrefix(originalLine, "  ") {
			inProvides = false
		}

		// Parse module entries in provides section
		if inProvides && strings.HasPrefix(originalLine, "      ") {
			// Module entry: "      ModuleName version"
			parts := strings.Fields(strings.TrimSpace(line))
			if len(parts) >= 2 {
				moduleName := parts[0]
				version := parts[1]

				snapshot.Modules[moduleName] = SnapshotModule{
					Version:      version,
					Distribution: currentDist,
					Dependencies: []Dependency{},
				}
			}
			continue
		}

		// Distribution name line (starts with 2 spaces)
		if inDistribution && strings.HasPrefix(originalLine, "  ") && !strings.HasPrefix(originalLine, "    ") {
			parts := strings.Fields(strings.TrimSpace(line))
			if len(parts) >= 1 {
				currentDist = parts[0]
				inProvides = false
			}
			continue
		}
	}

	return snapshot, nil
}
