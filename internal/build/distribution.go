// ABOUTME: Distribution build functionality for CPAN-ready packages
// ABOUTME: Creates production builds with metadata and clean Perl code

package build

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"tamarou.com/pvm/internal/project"
	"tamarou.com/pvm/internal/pvi"
)

// DistributionBuilder handles production builds for CPAN distribution
type DistributionBuilder struct {
	pscBuilder *PSCBuilder
	projectCtx *project.ProjectContext
}

// NewDistributionBuilder creates a new distribution builder instance
func NewDistributionBuilder(projectCtx *project.ProjectContext) (*DistributionBuilder, error) {
	pscBuilder, err := NewPSCBuilder(projectCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to create PSC builder: %w", err)
	}

	return &DistributionBuilder{
		pscBuilder: pscBuilder,
		projectCtx: projectCtx,
	}, nil
}

// DistributionBuildResult represents the result of a distribution build
type DistributionBuildResult struct {
	Success           bool          `json:"success"`
	BuildDir          string        `json:"build_dir"`
	FilesProcessed    int           `json:"files_processed"`
	MetadataGenerated bool          `json:"metadata_generated"`
	TypeErrors        []TypeError   `json:"type_errors"`
	Duration          time.Duration `json:"duration"`
	BuildErrors       []string      `json:"build_errors"`
	DistributionName  string        `json:"distribution_name"`
}

// Build performs a full distribution build
func (b *DistributionBuilder) Build(ctx context.Context, options *BuildOptions) (*DistributionBuildResult, error) {
	start := time.Now()
	result := &DistributionBuildResult{
		Success:        true,
		BuildErrors:    []string{},
		TypeErrors:     []TypeError{},
		FilesProcessed: 0,
	}

	if options == nil {
		options = &BuildOptions{}
	}

	// Check for context cancellation early
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Determine build directory
	buildDir := options.OutputDir
	if buildDir == "" {
		if b.projectCtx != nil && b.projectCtx.IsProject {
			buildDir = filepath.Join(b.projectCtx.RootDir, "build")
		} else {
			buildDir = "build"
		}
	}
	result.BuildDir = buildDir

	// Clean build directory if requested
	if options.CleanFirst {
		if err := os.RemoveAll(buildDir); err != nil {
			return nil, fmt.Errorf("failed to clean build directory: %w", err)
		}
	}

	// Create build directory structure
	err := b.createBuildStructure(buildDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create build structure: %w", err)
	}

	// Determine source directories
	sourceDirs := options.SourceDirs
	if len(sourceDirs) == 0 {
		if b.projectCtx != nil && b.projectCtx.IsProject {
			sourceDirs = []string{
				filepath.Join(b.projectCtx.RootDir, "lib"),
				filepath.Join(b.projectCtx.RootDir, "script"),
				filepath.Join(b.projectCtx.RootDir, "t"),
			}
		} else {
			sourceDirs = []string{"lib", "script", "t"}
		}
	}

	// Type check all source files first
	if !options.SkipTypeCheck {
		typeCheckResult, err := b.pscBuilder.TypeCheck(ctx, sourceDirs)
		if err != nil {
			return nil, fmt.Errorf("type checking failed: %w", err)
		}

		result.TypeErrors = typeCheckResult.TypeErrors
		if len(result.TypeErrors) > 0 && options.StrictTypeCheck {
			result.Success = false
			result.Duration = time.Since(start)
			return result, nil
		}
	}

	// Process source files (copy and strip types)
	filesProcessed, err := b.processSourceFiles(ctx, sourceDirs, buildDir)
	if err != nil {
		result.BuildErrors = append(result.BuildErrors, fmt.Sprintf("Failed to process source files: %v", err))
		result.Success = false
	} else {
		result.FilesProcessed = filesProcessed
	}

	// Generate metadata files if not skipped
	if !options.SkipMetadata {
		err = b.generateMetadata(buildDir)
		if err != nil {
			result.BuildErrors = append(result.BuildErrors, fmt.Sprintf("Failed to generate metadata: %v", err))
			result.Success = false
		} else {
			result.MetadataGenerated = true
		}
	}

	// Determine distribution name
	result.DistributionName, _ = b.getDistributionName()

	result.Duration = time.Since(start)
	return result, nil
}

// BuildOptions configures distribution build behavior
type BuildOptions struct {
	OutputDir       string   `json:"output_dir"`
	SourceDirs      []string `json:"source_dirs"`
	CleanFirst      bool     `json:"clean_first"`
	SkipTypeCheck   bool     `json:"skip_type_check"`
	StrictTypeCheck bool     `json:"strict_type_check"`
	SkipMetadata    bool     `json:"skip_metadata"`
	IncludeTests    bool     `json:"include_tests"`
	IncludeScripts  bool     `json:"include_scripts"`
}

// createBuildStructure creates the standard CPAN distribution directory structure
func (b *DistributionBuilder) createBuildStructure(buildDir string) error {
	dirs := []string{
		buildDir,
		filepath.Join(buildDir, "lib"),
		filepath.Join(buildDir, "t"),
		filepath.Join(buildDir, "script"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// processSourceFiles copies and processes source files to the build directory
func (b *DistributionBuilder) processSourceFiles(ctx context.Context, sourceDirs []string, buildDir string) (int, error) {
	filesProcessed := 0

	for _, sourceDir := range sourceDirs {
		// Check if source directory exists
		if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
			continue // Skip non-existent directories
		}

		// Determine target directory based on source directory name
		sourceName := filepath.Base(sourceDir)
		targetDir := filepath.Join(buildDir, sourceName)

		err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Check for context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Calculate relative path and target path
			relPath, err := filepath.Rel(sourceDir, path)
			if err != nil {
				return fmt.Errorf("failed to calculate relative path: %w", err)
			}

			targetPath := filepath.Join(targetDir, relPath)

			// Ensure target directory exists
			targetDirPath := filepath.Dir(targetPath)
			if err := os.MkdirAll(targetDirPath, 0755); err != nil {
				return fmt.Errorf("failed to create target directory %s: %w", targetDirPath, err)
			}

			// Process file based on type
			if strings.HasSuffix(path, ".pm") {
				// Strip type annotations from Perl modules
				err = b.pscBuilder.compileFile(path, targetPath)
				if err != nil {
					return fmt.Errorf("failed to compile %s: %w", path, err)
				}
			} else {
				// Copy other files as-is
				err = b.copyFile(path, targetPath)
				if err != nil {
					return fmt.Errorf("failed to copy %s: %w", path, err)
				}
			}

			filesProcessed++
			return nil
		})

		if err != nil {
			return filesProcessed, fmt.Errorf("failed to process source directory %s: %w", sourceDir, err)
		}
	}

	return filesProcessed, nil
}

// generateMetadata creates all CPAN metadata files
func (b *DistributionBuilder) generateMetadata(buildDir string) error {
	// Generate META.json
	if err := b.generateMETAJSON(buildDir); err != nil {
		return fmt.Errorf("failed to generate META.json: %w", err)
	}

	// Generate META.yml
	if err := b.generateMETAYML(buildDir); err != nil {
		return fmt.Errorf("failed to generate META.yml: %w", err)
	}

	// Generate Makefile.PL
	if err := b.generateMakefilePL(buildDir); err != nil {
		return fmt.Errorf("failed to generate Makefile.PL: %w", err)
	}

	// Generate MANIFEST
	if err := b.generateMANIFEST(buildDir); err != nil {
		return fmt.Errorf("failed to generate MANIFEST: %w", err)
	}

	// Copy cpanfile if it exists
	if err := b.copyCpanfile(buildDir); err != nil {
		// Not a fatal error, just log it
		fmt.Printf("Warning: failed to copy cpanfile: %v\n", err)
	}

	return nil
}

// generateMETAJSON creates the META.json file
func (b *DistributionBuilder) generateMETAJSON(buildDir string) error {
	metadata, err := b.getProjectMetadata()
	if err != nil {
		return fmt.Errorf("failed to get project metadata: %w", err)
	}

	// Create META.json structure
	metaJSON := map[string]interface{}{
		"abstract":       metadata.Abstract,
		"author":         []string{metadata.Author},
		"dynamic_config": 0,
		"generated_by":   "PVM (Perl Version Manager)",
		"license":        []string{metadata.License},
		"meta-spec": map[string]interface{}{
			"version": "2",
			"url":     "https://metacpan.org/pod/CPAN::Meta::Spec",
		},
		"name":           metadata.Name,
		"release_status": "stable",
		"version":        metadata.Version,
	}

	// Add prerequisites if available
	if len(metadata.Prerequisites) > 0 {
		prereqs := map[string]interface{}{
			"runtime": map[string]interface{}{
				"requires": metadata.Prerequisites,
			},
		}
		metaJSON["prereqs"] = prereqs
	}

	// Add provides information for modules
	provides, err := b.generateProvides(buildDir)
	if err == nil && len(provides) > 0 {
		metaJSON["provides"] = provides
	}

	// Write META.json
	jsonData, err := json.MarshalIndent(metaJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal META.json: %w", err)
	}

	metaPath := filepath.Join(buildDir, "META.json")
	return os.WriteFile(metaPath, jsonData, 0644)
}

// generateMETAYML creates the META.yml file (legacy format)
func (b *DistributionBuilder) generateMETAYML(buildDir string) error {
	metadata, err := b.getProjectMetadata()
	if err != nil {
		return fmt.Errorf("failed to get project metadata: %w", err)
	}

	var yamlLines []string
	yamlLines = append(yamlLines, "---")
	yamlLines = append(yamlLines, fmt.Sprintf("abstract: '%s'", metadata.Abstract))
	yamlLines = append(yamlLines, "author:")
	yamlLines = append(yamlLines, fmt.Sprintf("  - '%s'", metadata.Author))
	yamlLines = append(yamlLines, "build_requires: {}")
	yamlLines = append(yamlLines, "configure_requires: {}")
	yamlLines = append(yamlLines, "dynamic_config: 0")
	yamlLines = append(yamlLines, "generated_by: 'PVM (Perl Version Manager)'")
	yamlLines = append(yamlLines, "license: "+metadata.License)
	yamlLines = append(yamlLines, "meta-spec:")
	yamlLines = append(yamlLines, "  url: http://module-build.sourceforge.net/META-spec-v1.4.html")
	yamlLines = append(yamlLines, "  version: '1.4'")
	yamlLines = append(yamlLines, fmt.Sprintf("name: %s", metadata.Name))
	yamlLines = append(yamlLines, "requires:")

	if len(metadata.Prerequisites) > 0 {
		for module, version := range metadata.Prerequisites {
			yamlLines = append(yamlLines, fmt.Sprintf("  %s: '%s'", module, version))
		}
	} else {
		yamlLines = append(yamlLines, "  perl: '5.008'")
	}

	yamlLines = append(yamlLines, "resources: {}")
	yamlLines = append(yamlLines, fmt.Sprintf("version: '%s'", metadata.Version))

	yamlContent := strings.Join(yamlLines, "\n") + "\n"
	metaPath := filepath.Join(buildDir, "META.yml")
	return os.WriteFile(metaPath, []byte(yamlContent), 0644)
}

// generateMakefilePL creates the Makefile.PL installer
func (b *DistributionBuilder) generateMakefilePL(buildDir string) error {
	metadata, err := b.getProjectMetadata()
	if err != nil {
		return fmt.Errorf("failed to get project metadata: %w", err)
	}

	var lines []string
	lines = append(lines, "use strict;")
	lines = append(lines, "use warnings;")
	lines = append(lines, "use ExtUtils::MakeMaker;")
	lines = append(lines, "")
	lines = append(lines, "WriteMakefile(")
	lines = append(lines, fmt.Sprintf("    NAME         => '%s',", metadata.Name))
	lines = append(lines, fmt.Sprintf("    VERSION      => '%s',", metadata.Version))
	lines = append(lines, fmt.Sprintf("    ABSTRACT     => '%s',", metadata.Abstract))
	lines = append(lines, fmt.Sprintf("    AUTHOR       => '%s',", metadata.Author))
	lines = append(lines, fmt.Sprintf("    LICENSE      => '%s',", metadata.License))

	// Add prerequisites
	if len(metadata.Prerequisites) > 0 {
		lines = append(lines, "    PREREQ_PM    => {")
		for module, version := range metadata.Prerequisites {
			lines = append(lines, fmt.Sprintf("        '%s' => '%s',", module, version))
		}
		lines = append(lines, "    },")
	}

	lines = append(lines, "    MIN_PERL_VERSION => '5.008',")
	lines = append(lines, ");")

	content := strings.Join(lines, "\n") + "\n"
	makefilePath := filepath.Join(buildDir, "Makefile.PL")
	return os.WriteFile(makefilePath, []byte(content), 0644)
}

// generateMANIFEST creates the MANIFEST file listing all files
func (b *DistributionBuilder) generateMANIFEST(buildDir string) error {
	var files []string

	err := filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, err := filepath.Rel(buildDir, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk build directory: %w", err)
	}

	// Sort files for consistent output
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i] > files[j] {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Add MANIFEST itself to the list
	files = append(files, "MANIFEST")

	// Sort files again after adding MANIFEST
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i] > files[j] {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	content := strings.Join(files, "\n") + "\n"
	manifestPath := filepath.Join(buildDir, "MANIFEST")
	return os.WriteFile(manifestPath, []byte(content), 0644)
}

// ProjectMetadata holds metadata about the project
type ProjectMetadata struct {
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	Abstract      string            `json:"abstract"`
	Author        string            `json:"author"`
	License       string            `json:"license"`
	Prerequisites map[string]string `json:"prerequisites"`
}

// getProjectMetadata extracts metadata from the project
func (b *DistributionBuilder) getProjectMetadata() (*ProjectMetadata, error) {
	metadata := &ProjectMetadata{
		Name:          "My-Perl-Module",
		Version:       "0.01",
		Abstract:      "Perl module",
		Author:        "Author <author@example.com>",
		License:       "perl_5",
		Prerequisites: make(map[string]string),
	}

	// Try to extract name from main module
	if b.projectCtx != nil && b.projectCtx.IsProject {
		name, err := b.extractModuleName()
		if err == nil && name != "" {
			metadata.Name = name
		}

		// Try to load prerequisites from cpanfile
		cpanfilePath := filepath.Join(b.projectCtx.RootDir, "cpanfile")
		if _, err := os.Stat(cpanfilePath); err == nil {
			manager := pvi.NewCpanfileManager(cpanfilePath)
			cpanfile, err := manager.ListDependencies()
			if err == nil {
				for _, req := range cpanfile.Requirements {
					if req.Phase == "runtime" && req.Relationship == "requires" {
						version := req.Version
						if version == "" {
							version = "0"
						}
						metadata.Prerequisites[req.Module] = version
					}
				}
			}
		}
	}

	return metadata, nil
}

// extractModuleName tries to extract the main module name from the project
func (b *DistributionBuilder) extractModuleName() (string, error) {
	if b.projectCtx == nil || !b.projectCtx.IsProject {
		return "", fmt.Errorf("not in a project")
	}

	libDir := filepath.Join(b.projectCtx.RootDir, "lib")
	var moduleName string

	err := filepath.Walk(libDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".pm") && moduleName == "" {
			relPath, err := filepath.Rel(libDir, path)
			if err != nil {
				return err
			}

			// Convert file path to module name
			name := strings.TrimSuffix(relPath, ".pm")
			name = strings.ReplaceAll(name, string(filepath.Separator), "::")
			moduleName = name
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// Convert :: to - for distribution name
	distName := strings.ReplaceAll(moduleName, "::", "-")
	return distName, nil
}

// generateProvides creates the provides section for META files
func (b *DistributionBuilder) generateProvides(buildDir string) (map[string]interface{}, error) {
	provides := make(map[string]interface{})
	libDir := filepath.Join(buildDir, "lib")

	if _, err := os.Stat(libDir); os.IsNotExist(err) {
		return provides, nil
	}

	err := filepath.Walk(libDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".pm") {
			relPath, err := filepath.Rel(libDir, path)
			if err != nil {
				return err
			}

			// Convert file path to module name
			moduleName := strings.TrimSuffix(relPath, ".pm")
			moduleName = strings.ReplaceAll(moduleName, string(filepath.Separator), "::")

			// Try to extract version from the module
			version, err := b.extractModuleVersion(path)
			if err != nil {
				version = "undef"
			}

			provides[moduleName] = map[string]interface{}{
				"file":    filepath.Join("lib", relPath),
				"version": version,
			}
		}

		return nil
	})

	return provides, err
}

// extractModuleVersion tries to extract version from a Perl module file
func (b *DistributionBuilder) extractModuleVersion(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Look for version declarations
	versionRe := regexp.MustCompile(`(?m)^\s*(?:our\s+)?\$VERSION\s*=\s*['"]([^'"]+)['"]`)
	matches := versionRe.FindStringSubmatch(string(content))
	if len(matches) > 1 {
		return matches[1], nil
	}

	// Look for version in use statements
	useVersionRe := regexp.MustCompile(`(?m)^\s*use\s+version\s*;\s*our\s+\$VERSION\s*=\s*version->declare\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	matches = useVersionRe.FindStringSubmatch(string(content))
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "undef", nil
}

// copyCpanfile copies the cpanfile to the build directory if it exists
func (b *DistributionBuilder) copyCpanfile(buildDir string) error {
	if b.projectCtx == nil || !b.projectCtx.IsProject {
		return fmt.Errorf("not in a project")
	}

	sourcePath := filepath.Join(b.projectCtx.RootDir, "cpanfile")
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("cpanfile does not exist")
	}

	targetPath := filepath.Join(buildDir, "cpanfile")
	return b.copyFile(sourcePath, targetPath)
}

// copyFile copies a file from source to destination
func (b *DistributionBuilder) copyFile(src, dst string) error {
	sourceData, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	return os.WriteFile(dst, sourceData, 0644)
}

// getDistributionName returns the name of the distribution
func (b *DistributionBuilder) getDistributionName() (string, error) {
	name, err := b.extractModuleName()
	if err != nil || name == "" {
		return "My-Perl-Distribution", nil
	}
	return name, nil
}

// Clean removes the build directory
func (b *DistributionBuilder) Clean(buildDir string) error {
	if buildDir == "" {
		if b.projectCtx != nil && b.projectCtx.IsProject {
			buildDir = filepath.Join(b.projectCtx.RootDir, "build")
		} else {
			buildDir = "build"
		}
	}

	return os.RemoveAll(buildDir)
}

// Validate checks if the distribution structure is valid
func (b *DistributionBuilder) Validate(buildDir string) error {
	if buildDir == "" {
		if b.projectCtx != nil && b.projectCtx.IsProject {
			buildDir = filepath.Join(b.projectCtx.RootDir, "build")
		} else {
			buildDir = "build"
		}
	}

	// Check if build directory exists
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		return fmt.Errorf("build directory does not exist: %s", buildDir)
	}

	// Check for required files
	requiredFiles := []string{
		"META.json",
		"META.yml",
		"Makefile.PL",
		"MANIFEST",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(buildDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("required file missing: %s", file)
		}
	}

	// Check if lib directory has content
	libDir := filepath.Join(buildDir, "lib")
	if _, err := os.Stat(libDir); os.IsNotExist(err) {
		return fmt.Errorf("lib directory is missing")
	}

	// Check if there are any .pm files in lib
	hasModules := false
	err := filepath.Walk(libDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".pm") {
			hasModules = true
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to check lib directory: %w", err)
	}

	if !hasModules {
		return fmt.Errorf("no Perl modules found in lib directory")
	}

	return nil
}
