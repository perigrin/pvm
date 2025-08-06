// ABOUTME: cpanfile manipulation and management functionality
// ABOUTME: Provides functions to read, write, and modify cpanfile format

package pvi

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"tamarou.com/pvm/internal/cpan"
)

// CpanfileManager handles cpanfile operations
type CpanfileManager struct {
	Path string
}

// NewCpanfileManager creates a new cpanfile manager for the given path
func NewCpanfileManager(path string) *CpanfileManager {
	return &CpanfileManager{Path: path}
}

// AddDependency adds a new dependency to the cpanfile
func (cm *CpanfileManager) AddDependency(moduleName, versionConstraint string, isDev bool) error {
	// Read existing cpanfile if it exists
	var cpanfile *cpan.CPANFile

	if fileExists(cm.Path) {
		content, err := os.ReadFile(cm.Path)
		if err != nil {
			return fmt.Errorf("failed to read cpanfile: %w", err)
		}
		cpanfile, err = cpan.ParseCPANFile(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse cpanfile: %w", err)
		}
	} else {
		// Create new cpanfile structure
		cpanfile = &cpan.CPANFile{
			Requirements: []cpan.Requirement{},
			Features:     make(map[string][]cpan.Requirement),
			Platforms:    make(map[string][]cpan.Requirement),
		}
	}

	// Check if dependency already exists
	phase := "runtime"
	if isDev {
		phase = "develop"
	}

	// Remove existing dependency if it exists
	cm.removeDependencyFromList(&cpanfile.Requirements, moduleName)

	// Create new requirement
	req := cpan.Requirement{
		Module:       moduleName,
		Version:      versionConstraint,
		Phase:        phase,
		Relationship: "requires",
	}

	// Add to appropriate list
	if isDev {
		// For development dependencies, we need to add them to the develop phase
		// This will be handled in the write phase
		req.Phase = "develop"
	}
	cpanfile.Requirements = append(cpanfile.Requirements, req)

	// Write back to file
	return cm.writeCpanfile(cpanfile)
}

// RemoveDependency removes a dependency from the cpanfile
func (cm *CpanfileManager) RemoveDependency(moduleName string) error {
	if !fileExists(cm.Path) {
		return fmt.Errorf("cpanfile does not exist: %s", cm.Path)
	}

	content, err := os.ReadFile(cm.Path)
	if err != nil {
		return fmt.Errorf("failed to read cpanfile: %w", err)
	}

	cpanfile, err := cpan.ParseCPANFile(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse cpanfile: %w", err)
	}

	// Remove from all lists
	removed := cm.removeDependencyFromList(&cpanfile.Requirements, moduleName)

	for featureName, reqs := range cpanfile.Features {
		if cm.removeDependencyFromList(&reqs, moduleName) {
			cpanfile.Features[featureName] = reqs
			removed = true
		}
	}

	for platformName, reqs := range cpanfile.Platforms {
		if cm.removeDependencyFromList(&reqs, moduleName) {
			cpanfile.Platforms[platformName] = reqs
			removed = true
		}
	}

	if !removed {
		return fmt.Errorf("dependency %s not found in cpanfile", moduleName)
	}

	return cm.writeCpanfile(cpanfile)
}

// ListDependencies returns all dependencies from the cpanfile
func (cm *CpanfileManager) ListDependencies() (*cpan.CPANFile, error) {
	if !fileExists(cm.Path) {
		return &cpan.CPANFile{
			Requirements: []cpan.Requirement{},
			Features:     make(map[string][]cpan.Requirement),
			Platforms:    make(map[string][]cpan.Requirement),
		}, nil
	}

	content, err := os.ReadFile(cm.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read cpanfile: %w", err)
	}

	return cpan.ParseCPANFile(string(content))
}

// removeDependencyFromList removes a dependency from a list of requirements
func (cm *CpanfileManager) removeDependencyFromList(reqs *[]cpan.Requirement, moduleName string) bool {
	for i, req := range *reqs {
		if req.Module == moduleName {
			*reqs = append((*reqs)[:i], (*reqs)[i+1:]...)
			return true
		}
	}
	return false
}

// writeCpanfile writes a CPANFile structure back to the cpanfile
func (cm *CpanfileManager) writeCpanfile(cpanfile *cpan.CPANFile) error {
	// Ensure the directory exists
	dir := filepath.Dir(cm.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Read existing content to preserve formatting and comments if possible
	var lines []string
	var existingContent string

	if fileExists(cm.Path) {
		content, err := os.ReadFile(cm.Path)
		if err == nil {
			existingContent = string(content)
			lines = strings.Split(existingContent, "\n")
		}
	}

	// If we have existing content, try to preserve structure
	if existingContent != "" {
		return cm.updateExistingCpanfile(lines, cpanfile)
	}

	// Create new cpanfile
	return cm.createNewCpanfile(cpanfile)
}

// updateExistingCpanfile updates an existing cpanfile while preserving formatting
func (cm *CpanfileManager) updateExistingCpanfile(lines []string, cpanfile *cpan.CPANFile) error {
	var output []string
	processed := make(map[string]bool)
	inDevelopBlock := false
	developBlockIndent := ""
	newDepsInserted := false

	// Regular expressions for parsing
	requiresRe := regexp.MustCompile(`^(\s*)requires\s+'([^']+)'`)
	developRe := regexp.MustCompile(`^(\s*)on\s+'develop'\s*=>\s*sub\s*\{`)
	blockEndRe := regexp.MustCompile(`^(\s*)\}`)

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments as-is
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			output = append(output, line)
			continue
		}

		// Check for develop block start
		if developRe.MatchString(line) {
			inDevelopBlock = true
			developBlockIndent = developRe.FindStringSubmatch(line)[1]
			output = append(output, line)
			continue
		}

		// Check for block end
		if blockEndRe.MatchString(line) && inDevelopBlock {
			// Add any new develop dependencies before closing block
			for _, req := range cpanfile.Requirements {
				if req.Phase == "develop" && !processed[req.Module] {
					depLine := fmt.Sprintf("%s    requires '%s'", developBlockIndent, req.Module)
					if req.Version != "" {
						depLine += fmt.Sprintf(", '%s'", req.Version)
					}
					depLine += ";"
					output = append(output, depLine)
					processed[req.Module] = true
				}
			}
			inDevelopBlock = false
			output = append(output, line)
			continue
		}

		// Check for requires statements
		if matches := requiresRe.FindStringSubmatch(line); matches != nil {
			indent := matches[1]
			moduleName := matches[2]

			// Find if this module exists in our new requirements
			var newReq *cpan.Requirement
			for _, req := range cpanfile.Requirements {
				if req.Module == moduleName {
					// Check if it should be in develop block
					if inDevelopBlock && req.Phase != "develop" {
						// Skip this line, it shouldn't be in develop block
						continue
					}
					if !inDevelopBlock && req.Phase == "develop" {
						// Skip this line, it should be in develop block
						continue
					}
					newReq = &req
					break
				}
			}

			if newReq != nil {
				// Update the line with new version
				newLine := fmt.Sprintf("%srequires '%s'", indent, newReq.Module)
				if newReq.Version != "" {
					newLine += fmt.Sprintf(", '%s'", newReq.Version)
				}
				newLine += ";"
				output = append(output, newLine)
				processed[newReq.Module] = true
			}
			// If newReq is nil, the dependency was removed, so skip the line

			// If this was a perl version requirement and we haven't inserted new deps yet,
			// insert any new runtime dependencies right after it
			if !inDevelopBlock && moduleName == "perl" && !newDepsInserted {
				for _, req := range cpanfile.Requirements {
					if req.Phase != "develop" && !processed[req.Module] {
						depLine := fmt.Sprintf("requires '%s'", req.Module)
						if req.Version != "" {
							depLine += fmt.Sprintf(", '%s'", req.Version)
						}
						depLine += ";"
						output = append(output, depLine)
						processed[req.Module] = true
					}
				}
				newDepsInserted = true
			}
			continue
		}

		// Keep other lines as-is
		output = append(output, line)
	}

	// Add any new runtime dependencies that weren't processed (only if they weren't inserted after perl version)
	if !newDepsInserted {
		addedNewline := false
		for _, req := range cpanfile.Requirements {
			if req.Phase != "develop" && !processed[req.Module] {
				if !addedNewline {
					output = append(output, "")
					addedNewline = true
				}
				depLine := fmt.Sprintf("requires '%s'", req.Module)
				if req.Version != "" {
					depLine += fmt.Sprintf(", '%s'", req.Version)
				}
				depLine += ";"
				output = append(output, depLine)
				processed[req.Module] = true
			}
		}
	}

	// Add develop block if needed and there are develop dependencies
	hasDevDeps := false
	for _, req := range cpanfile.Requirements {
		if req.Phase == "develop" && !processed[req.Module] {
			hasDevDeps = true
			break
		}
	}

	if hasDevDeps {
		output = append(output, "")
		output = append(output, "on 'develop' => sub {")
		for _, req := range cpanfile.Requirements {
			if req.Phase == "develop" && !processed[req.Module] {
				depLine := fmt.Sprintf("    requires '%s'", req.Module)
				if req.Version != "" {
					depLine += fmt.Sprintf(", '%s'", req.Version)
				}
				depLine += ";"
				output = append(output, depLine)
				processed[req.Module] = true
			}
		}
		output = append(output, "};")
	}

	// Write the file
	content := strings.Join(output, "\n") + "\n"
	return os.WriteFile(cm.Path, []byte(content), 0644)
}

// createNewCpanfile creates a new cpanfile from scratch
func (cm *CpanfileManager) createNewCpanfile(cpanfile *cpan.CPANFile) error {
	var lines []string

	// Add header comment
	lines = append(lines, "# cpanfile")
	lines = append(lines, "# Generated by PVM on "+time.Now().Format("2006-01-02 15:04:05"))
	lines = append(lines, "")

	// Add runtime requirements
	for _, req := range cpanfile.Requirements {
		if req.Phase != "develop" {
			line := fmt.Sprintf("requires '%s'", req.Module)
			if req.Version != "" {
				line += fmt.Sprintf(", '%s'", req.Version)
			}
			line += ";"
			lines = append(lines, line)
		}
	}

	// Add develop requirements if any
	var devReqs []cpan.Requirement
	for _, req := range cpanfile.Requirements {
		if req.Phase == "develop" {
			devReqs = append(devReqs, req)
		}
	}

	if len(devReqs) > 0 {
		lines = append(lines, "")
		lines = append(lines, "on 'develop' => sub {")
		for _, req := range devReqs {
			line := fmt.Sprintf("    requires '%s'", req.Module)
			if req.Version != "" {
				line += fmt.Sprintf(", '%s'", req.Version)
			}
			line += ";"
			lines = append(lines, line)
		}
		lines = append(lines, "};")
	}

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(cm.Path, []byte(content), 0644)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// SnapshotEntry represents a single entry in the cpanfile.snapshot
type SnapshotEntry struct {
	Distribution string            `json:"distribution"`
	Pathname     string            `json:"pathname"`
	Provides     map[string]string `json:"provides"`
	Requirements map[string]string `json:"requirements"`
	Checksum     string            `json:"checksum,omitempty"`
}

// Snapshot represents the complete cpanfile.snapshot structure
type Snapshot struct {
	Version       string                   `json:"version"`
	Distributions map[string]SnapshotEntry `json:"distributions"`
	GeneratedAt   time.Time                `json:"generated_at"`
	PerlVersion   string                   `json:"perl_version"`
}

// GenerateSnapshot creates a cpanfile.snapshot from currently installed modules
func (cm *CpanfileManager) GenerateSnapshot() (*Snapshot, error) {
	// Read cpanfile to get the dependency list
	cpanfile, err := cm.ListDependencies()
	if err != nil {
		return nil, fmt.Errorf("failed to read cpanfile: %w", err)
	}

	// Create snapshot structure
	snapshot := &Snapshot{
		Version:       "1.0",
		Distributions: make(map[string]SnapshotEntry),
		GeneratedAt:   time.Now(),
	}

	// Get current Perl version
	if currentPerl, err := getPerlVersion(); err == nil {
		snapshot.PerlVersion = currentPerl
	}

	// For each dependency in cpanfile, find the installed version
	for _, req := range cpanfile.Requirements {
		entry, err := cm.getInstalledModuleInfo(req.Module)
		if err != nil {
			// Module not installed, skip (could warn here)
			continue
		}
		snapshot.Distributions[entry.Distribution] = entry
	}

	return snapshot, nil
}

// WriteSnapshot writes a snapshot to cpanfile.snapshot
func (cm *CpanfileManager) WriteSnapshot(snapshot *Snapshot) error {
	snapshotPath := strings.TrimSuffix(cm.Path, "cpanfile") + "cpanfile.snapshot"

	// Create snapshot content in carton format
	var lines []string
	lines = append(lines, "# carton snapshot format v1.0")
	lines = append(lines, fmt.Sprintf("# Generated on %s", snapshot.GeneratedAt.Format("2006-01-02 15:04:05")))
	if snapshot.PerlVersion != "" {
		lines = append(lines, fmt.Sprintf("# Perl version: %s", snapshot.PerlVersion))
	}
	lines = append(lines, "DISTRIBUTIONS")

	for distName, entry := range snapshot.Distributions {
		lines = append(lines, fmt.Sprintf("  %s", distName))
		lines = append(lines, fmt.Sprintf("    pathname: %s", entry.Pathname))

		if len(entry.Provides) > 0 {
			lines = append(lines, "    provides:")
			for module, version := range entry.Provides {
				lines = append(lines, fmt.Sprintf("      %s %s", module, version))
			}
		}

		if len(entry.Requirements) > 0 {
			lines = append(lines, "    requirements:")
			for module, version := range entry.Requirements {
				lines = append(lines, fmt.Sprintf("      %s %s", module, version))
			}
		}

		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return os.WriteFile(snapshotPath, []byte(content), 0644)
}

// ReadSnapshot reads a cpanfile.snapshot file
func (cm *CpanfileManager) ReadSnapshot() (*Snapshot, error) {
	snapshotPath := strings.TrimSuffix(cm.Path, "cpanfile") + "cpanfile.snapshot"

	if !fileExists(snapshotPath) {
		return nil, fmt.Errorf("cpanfile.snapshot not found at %s", snapshotPath)
	}

	content, err := os.ReadFile(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cpanfile.snapshot: %w", err)
	}

	return cm.parseSnapshot(string(content))
}

// parseSnapshot parses carton snapshot format
func (cm *CpanfileManager) parseSnapshot(content string) (*Snapshot, error) {
	snapshot := &Snapshot{
		Version:       "1.0",
		Distributions: make(map[string]SnapshotEntry),
		GeneratedAt:   time.Now(),
	}

	lines := strings.Split(content, "\n")
	var currentDist string
	var currentEntry SnapshotEntry
	inProvides := false
	inRequirements := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			if strings.Contains(line, "Perl version:") {
				parts := strings.Split(line, "Perl version:")
				if len(parts) > 1 {
					snapshot.PerlVersion = strings.TrimSpace(parts[1])
				}
			}
			continue
		}

		if trimmed == "DISTRIBUTIONS" {
			continue
		}

		// Check indentation to determine structure
		indent := len(line) - len(strings.TrimLeft(line, " "))

		switch indent {
		case 2:
			if !strings.HasPrefix(trimmed, "pathname:") && !strings.HasPrefix(trimmed, "provides:") && !strings.HasPrefix(trimmed, "requirements:") {
				// New distribution
				if currentDist != "" {
					snapshot.Distributions[currentDist] = currentEntry
				}
				currentDist = trimmed
				currentEntry = SnapshotEntry{
					Distribution: currentDist,
					Provides:     make(map[string]string),
					Requirements: make(map[string]string),
				}
				inProvides = false
				inRequirements = false
			}
		case 4:
			switch {
			case strings.HasPrefix(trimmed, "pathname:"):
				currentEntry.Pathname = strings.TrimSpace(strings.TrimPrefix(trimmed, "pathname:"))
			case trimmed == "provides:":
				inProvides = true
				inRequirements = false
			case trimmed == "requirements:":
				inProvides = false
				inRequirements = true
			}
		case 6:
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				module := parts[0]
				version := parts[1]
				switch {
				case inProvides:
					currentEntry.Provides[module] = version
				case inRequirements:
					currentEntry.Requirements[module] = version
				}
			}
		}
	}

	// Add the last distribution
	if currentDist != "" {
		snapshot.Distributions[currentDist] = currentEntry
	}

	return snapshot, nil
}

// getInstalledModuleInfo gets information about an installed module
func (cm *CpanfileManager) getInstalledModuleInfo(moduleName string) (SnapshotEntry, error) {
	// This is a simplified implementation
	// In a real implementation, you would query the installed modules
	// For now, we'll create a placeholder entry

	entry := SnapshotEntry{
		Distribution: fmt.Sprintf("%s-1.00", moduleName),
		Pathname:     fmt.Sprintf("A/AU/AUTHOR/%s-1.00.tar.gz", moduleName),
		Provides:     map[string]string{moduleName: "1.00"},
		Requirements: make(map[string]string),
	}

	return entry, nil
}

// getPerlVersion returns the current Perl version
func getPerlVersion() (string, error) {
	// This would typically call perl -v and parse the output
	// For now, return a placeholder
	return "5.38.0", nil
}
