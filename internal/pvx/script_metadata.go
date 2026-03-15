// ABOUTME: Script metadata parsing for inline dependency declarations
// ABOUTME: Implements Perl equivalent of Python's inline script metadata

package pvx

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ScriptMetadata represents parsed metadata from a Perl script
type ScriptMetadata struct {
	Dependencies   []string `json:"dependencies"`
	PerlVersion    string   `json:"perl_version,omitempty"`
	Description    string   `json:"description,omitempty"`
	RequiresPython bool     `json:"requires_python,omitempty"`
	ExcludeNewer   string   `json:"exclude_newer,omitempty"`
}

// ParseScriptMetadata extracts metadata from a Perl script file
// Supports both POD-based and comment-based metadata formats
func ParseScriptMetadata(scriptPath string) (*ScriptMetadata, error) {
	file, err := os.Open(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open script file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	metadata := &ScriptMetadata{
		Dependencies: make([]string, 0),
	}

	// Try POD format first, then comment format
	podMetadata, err := parsePODMetadata(scanner)
	if err == nil && (len(podMetadata.Dependencies) > 0 || podMetadata.PerlVersion != "") {
		return podMetadata, nil
	}

	// Reset scanner for comment format
	file.Seek(0, 0)
	scanner = bufio.NewScanner(file)

	commentMetadata, err := parseCommentMetadata(scanner)
	if err != nil {
		return metadata, nil // Return empty metadata if parsing fails
	}

	return commentMetadata, nil
}

// parsePODMetadata parses POD-based metadata format:
// =begin pvm
// dependencies = [
//
//	"DBI",
//	"JSON::PP >= 4.0"
//
// ]
// perl_version = "5.30"
// =end pvm
func parsePODMetadata(scanner *bufio.Scanner) (*ScriptMetadata, error) {
	metadata := &ScriptMetadata{
		Dependencies: make([]string, 0),
	}

	inPVMBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "=begin pvm" {
			inPVMBlock = true
			continue
		}

		if line == "=end pvm" {
			break
		}

		if !inPVMBlock {
			continue
		}

		// Parse dependencies array
		if strings.HasPrefix(line, "dependencies = [") {
			// Multi-line dependencies
			deps, err := parseDependencyArray(scanner, line)
			if err != nil {
				return nil, err
			}
			metadata.Dependencies = append(metadata.Dependencies, deps...)
		} else if match := regexp.MustCompile(`^perl_version\s*=\s*"([^"]+)"`).FindStringSubmatch(line); match != nil {
			metadata.PerlVersion = match[1]
		} else if match := regexp.MustCompile(`^description\s*=\s*"([^"]+)"`).FindStringSubmatch(line); match != nil {
			metadata.Description = match[1]
		} else if match := regexp.MustCompile(`^exclude_newer\s*=\s*"([^"]+)"`).FindStringSubmatch(line); match != nil {
			metadata.ExcludeNewer = match[1]
		}
	}

	return metadata, scanner.Err()
}

// parseCommentMetadata parses comment-based metadata format (uv-style):
// # /// pvm
// # dependencies = [
// #   "DBI",
// #   "JSON::PP >= 4.0"
// # ]
// # ///
func parseCommentMetadata(scanner *bufio.Scanner) (*ScriptMetadata, error) {
	metadata := &ScriptMetadata{
		Dependencies: make([]string, 0),
	}

	inPVMBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "# /// pvm" {
			inPVMBlock = true
			continue
		}

		if line == "# ///" {
			break
		}

		if !inPVMBlock {
			continue
		}

		// Remove comment prefix
		if strings.HasPrefix(line, "# ") {
			line = strings.TrimSpace(line[2:])
		} else if strings.HasPrefix(line, "#") {
			line = strings.TrimSpace(line[1:])
		}

		// Parse dependencies array
		if strings.HasPrefix(line, "dependencies = [") {
			deps, err := parseDependencyArrayFromComments(scanner, line)
			if err != nil {
				return nil, err
			}
			metadata.Dependencies = append(metadata.Dependencies, deps...)
		} else if match := regexp.MustCompile(`^perl_version\s*=\s*"([^"]+)"`).FindStringSubmatch(line); match != nil {
			metadata.PerlVersion = match[1]
		} else if match := regexp.MustCompile(`^description\s*=\s*"([^"]+)"`).FindStringSubmatch(line); match != nil {
			metadata.Description = match[1]
		}
	}

	return metadata, scanner.Err()
}

// parseDependencyArray parses a multi-line dependency array from POD format
func parseDependencyArray(scanner *bufio.Scanner, firstLine string) ([]string, error) {
	dependencies := make([]string, 0)

	// Check if it's a single-line array
	if strings.Contains(firstLine, "]") {
		return parseSingleLineDependencies(firstLine)
	}

	// Multi-line array
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.Contains(line, "]") {
			// End of array, parse final deps and break
			if dep := extractDependencyFromLine(line); dep != "" {
				dependencies = append(dependencies, dep)
			}
			break
		}

		if dep := extractDependencyFromLine(line); dep != "" {
			dependencies = append(dependencies, dep)
		}
	}

	return dependencies, nil
}

// parseDependencyArrayFromComments parses dependencies from comment format
func parseDependencyArrayFromComments(scanner *bufio.Scanner, firstLine string) ([]string, error) {
	dependencies := make([]string, 0)

	// Check if it's a single-line array
	if strings.Contains(firstLine, "]") {
		return parseSingleLineDependencies(firstLine)
	}

	// Multi-line array
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Remove comment prefix
		if strings.HasPrefix(line, "# ") {
			line = strings.TrimSpace(line[2:])
		} else if strings.HasPrefix(line, "#") {
			line = strings.TrimSpace(line[1:])
		}

		if strings.Contains(line, "]") {
			// End of array
			if dep := extractDependencyFromLine(line); dep != "" {
				dependencies = append(dependencies, dep)
			}
			break
		}

		if dep := extractDependencyFromLine(line); dep != "" {
			dependencies = append(dependencies, dep)
		}
	}

	return dependencies, nil
}

// parseSingleLineDependencies parses dependencies from a single line
func parseSingleLineDependencies(line string) ([]string, error) {
	dependencies := make([]string, 0)

	// Extract content between [ and ]
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return dependencies, nil
	}

	content := matches[1]

	// Split by comma and extract quoted strings
	parts := strings.Split(content, ",")
	for _, part := range parts {
		if dep := extractDependencyFromLine(part); dep != "" {
			dependencies = append(dependencies, dep)
		}
	}

	return dependencies, nil
}

// extractDependencyFromLine extracts a dependency from a line
func extractDependencyFromLine(line string) string {
	line = strings.TrimSpace(line)

	// Remove trailing comma
	line = strings.TrimSuffix(line, ",")

	// Extract quoted string
	re := regexp.MustCompile(`"([^"]+)"`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}

	// Try single quotes
	re = regexp.MustCompile(`'([^']+)'`)
	matches = re.FindStringSubmatch(line)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// HasMetadata checks if a script file contains PVM metadata
func HasMetadata(scriptPath string) bool {
	metadata, err := ParseScriptMetadata(scriptPath)
	if err != nil {
		return false
	}

	return len(metadata.Dependencies) > 0 || metadata.PerlVersion != ""
}

// FormatMetadataAsPOD formats metadata as POD for writing to scripts
func FormatMetadataAsPOD(metadata *ScriptMetadata) string {
	var builder strings.Builder

	builder.WriteString("=begin pvm\n")

	if len(metadata.Dependencies) > 0 {
		builder.WriteString("dependencies = [\n")
		for _, dep := range metadata.Dependencies {
			builder.WriteString(fmt.Sprintf("  \"%s\",\n", dep))
		}
		builder.WriteString("]\n")
	}

	if metadata.PerlVersion != "" {
		builder.WriteString(fmt.Sprintf("perl_version = \"%s\"\n", metadata.PerlVersion))
	}

	if metadata.Description != "" {
		builder.WriteString(fmt.Sprintf("description = \"%s\"\n", metadata.Description))
	}

	if metadata.ExcludeNewer != "" {
		builder.WriteString(fmt.Sprintf("exclude_newer = \"%s\"\n", metadata.ExcludeNewer))
	}

	builder.WriteString("=end pvm\n")

	return builder.String()
}

// FormatMetadataAsComments formats metadata as comments for writing to scripts
func FormatMetadataAsComments(metadata *ScriptMetadata) string {
	var builder strings.Builder

	builder.WriteString("# /// pvm\n")

	if len(metadata.Dependencies) > 0 {
		builder.WriteString("# dependencies = [\n")
		for _, dep := range metadata.Dependencies {
			builder.WriteString(fmt.Sprintf("#   \"%s\",\n", dep))
		}
		builder.WriteString("# ]\n")
	}

	if metadata.PerlVersion != "" {
		builder.WriteString(fmt.Sprintf("# perl_version = \"%s\"\n", metadata.PerlVersion))
	}

	if metadata.Description != "" {
		builder.WriteString(fmt.Sprintf("# description = \"%s\"\n", metadata.Description))
	}

	builder.WriteString("# ///\n")

	return builder.String()
}
