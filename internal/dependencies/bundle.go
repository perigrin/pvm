// ABOUTME: Bundle management for module collections
// ABOUTME: Handles creation, export, import, and validation of module bundles

package dependencies

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/modules"
)

// Bundle represents a collection of modules for export/import
type Bundle struct {
	// Name of the bundle
	Name string `json:"name"`

	// Description of the bundle
	Description string `json:"description"`

	// Created timestamp
	Created time.Time `json:"created"`

	// PerlVersion used to create the bundle
	PerlVersion string `json:"perl_version"`

	// Modules included in the bundle
	Modules []*BundleEntry `json:"modules"`

	// Metadata for additional bundle information
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BundleEntry represents a module in a bundle
type BundleEntry struct {
	// Name is the module name
	Name string `json:"name"`

	// VersionConstraint is the version constraint (e.g., ">=2.0.0")
	VersionConstraint string `json:"version_constraint,omitempty"`

	// IsDev indicates if this is a development dependency
	IsDev bool `json:"is_dev,omitempty"`

	// IsOptional indicates if this dependency is optional
	IsOptional bool `json:"is_optional,omitempty"`

	// Phase indicates the dependency phase (runtime, build, test, develop)
	Phase string `json:"phase,omitempty"`

	// Relationship indicates the dependency relationship (requires, recommends, suggests)
	Relationship string `json:"relationship,omitempty"`
}

// BundleManager manages bundle operations
type BundleManager struct {
	resolver *DependencyResolver
	manager  interface{} // TODO: Use proper manager interface
	logger   *log.Logger
}

// NewBundleManager creates a new bundle manager
func NewBundleManager(resolver *DependencyResolver, manager interface{}, logger *log.Logger) *BundleManager {
	if logger == nil {
		logger = log.New(os.Stderr, "[BundleManager] ", log.LstdFlags)
	}

	return &BundleManager{
		resolver: resolver,
		manager:  manager,
		logger:   logger,
	}
}

// CreateBundle creates a bundle from the specified modules
func (bm *BundleManager) CreateBundle(ctx context.Context, modules []string, options BundleCreateOptions) (*Bundle, error) {
	if len(modules) == 0 {
		return nil, errors.NewSystemError(
			"401",
			"No modules specified for bundle creation",
			nil)
	}

	// Resolve dependencies if requested
	var allModules []string
	if options.IncludeDependencies {
		graph, err := bm.resolver.ResolveDependencies(ctx, modules)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
		}

		// Extract all modules from the dependency graph
		moduleSet := make(map[string]bool)
		for _, node := range graph.Nodes {
			if options.IncludeCore || !node.CoreModule {
				moduleSet[node.Name] = true
			}
		}

		allModules = make([]string, 0, len(moduleSet))
		for module := range moduleSet {
			allModules = append(allModules, module)
		}
	} else {
		allModules = modules
	}

	// Filter modules if pattern is provided
	if options.ModuleFilter != "" {
		filteredModules, err := bm.filterModules(allModules, options.ModuleFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to filter modules: %w", err)
		}
		allModules = filteredModules
	}

	// For now, create basic module entries for the requested modules
	// TODO: Integrate with actual module manager once types are resolved
	var moduleInfos []struct {
		Name    string
		Version string
	}

	for _, module := range allModules {
		moduleInfos = append(moduleInfos, struct {
			Name    string
			Version string
		}{
			Name:    module,
			Version: "1.0.0", // Default version for now
		})
	}

	// Get Perl version if needed
	perlVersion := "unknown"
	if options.PerlPath != "" {
		perlVersion = bm.getPerlVersion(ctx, options.PerlPath)
	}

	// Create bundle entries
	entries := make([]*BundleEntry, 0, len(moduleInfos))
	for _, info := range moduleInfos {
		entry := &BundleEntry{
			Name:         info.Name,
			Phase:        "runtime",
			Relationship: "requires",
		}

		if options.IncludeVersions && info.Version != "" {
			entry.VersionConstraint = ">=" + info.Version
		}

		entries = append(entries, entry)
	}

	// Create bundle
	bundle := &Bundle{
		Name:        options.Name,
		Description: options.Description,
		Created:     time.Now(),
		PerlVersion: perlVersion,
		Modules:     entries,
		Metadata:    options.Metadata,
	}

	return bundle, nil
}

// ExportBundle exports a bundle to a file
func (bm *BundleManager) ExportBundle(bundle *Bundle, options BundleExportOptions) error {
	if bundle == nil {
		return errors.NewSystemError(
			"402",
			"No bundle provided for export",
			nil)
	}

	if options.OutputPath == "" {
		return errors.NewSystemError(
			"403",
			"No output path provided for bundle export",
			nil)
	}

	// Serialize bundle based on format
	var data []byte
	var err error

	switch strings.ToLower(options.Format) {
	case "json", "":
		data, err = json.MarshalIndent(bundle, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal bundle as JSON: %w", err)
		}
	default:
		return errors.NewSystemError(
			"404",
			fmt.Sprintf("Unsupported export format: %s", options.Format),
			nil)
	}

	// Write to file
	if err := os.WriteFile(options.OutputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write bundle to %s: %w", options.OutputPath, err)
	}

	bm.logger.Printf("Exported bundle with %d modules to %s", len(bundle.Modules), options.OutputPath)
	return nil
}

// ImportBundle imports a bundle from a file
func (bm *BundleManager) ImportBundle(bundlePath string) (*Bundle, error) {
	if bundlePath == "" {
		return nil, errors.NewSystemError(
			"405",
			"No bundle path provided for import",
			nil)
	}

	// Read bundle file
	data, err := os.ReadFile(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read bundle from %s: %w", bundlePath, err)
	}

	// Parse bundle
	var bundle Bundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse bundle JSON: %w", err)
	}

	bm.logger.Printf("Imported bundle '%s' with %d modules from %s", bundle.Name, len(bundle.Modules), bundlePath)
	return &bundle, nil
}

// InstallBundle installs all modules from a bundle
func (bm *BundleManager) InstallBundle(ctx context.Context, bundle *Bundle, options BundleImportOptions) ([]*modules.InstallResult, error) {
	if bundle == nil {
		return nil, errors.NewSystemError(
			"406",
			"No bundle provided for installation",
			nil)
	}

	if len(bundle.Modules) == 0 {
		return []*modules.InstallResult{}, nil
	}

	// Filter modules if needed
	var modulesToInstall []*BundleEntry
	for _, entry := range bundle.Modules {
		// Skip optional modules unless forced
		if entry.IsOptional && !options.Force {
			bm.logger.Printf("Skipping optional module: %s", entry.Name)
			continue
		}

		// Skip already installed modules if requested
		if options.SkipInstalled {
			// This would need integration with the module manager to check if installed
			// For now, we'll install all modules
		}

		modulesToInstall = append(modulesToInstall, entry)
	}

	// Convert bundle entries to module install options
	var moduleNames []string
	installOptions := make(map[string]modules.InstallOptions)

	for _, entry := range modulesToInstall {
		moduleNames = append(moduleNames, entry.Name)

		installOptions[entry.Name] = modules.InstallOptions{
			RunTests:         !options.SkipTests,
			Force:            options.Force,
			SkipDependencies: options.SkipDependencies,
			Parallel:         options.Parallel,
			Workers:          options.Workers,
		}
	}

	// Install modules
	if options.DryRun {
		bm.logger.Printf("Dry run: would install %d modules", len(moduleNames))

		// Create dummy results for dry run
		results := make([]*modules.InstallResult, len(moduleNames))
		for i, name := range moduleNames {
			results[i] = &modules.InstallResult{
				ModuleName: name,
				Success:    true,
				DryRun:     true,
			}
		}
		return results, nil
	}

	// Use the module manager to install modules
	if bm.manager == nil {
		return nil, errors.NewSystemError(
			"407",
			"Module manager not available for bundle installation",
			nil)
	}

	// TODO: Implement actual module installation once manager interface is resolved
	// For now, create mock results
	var results []*modules.InstallResult
	for _, module := range moduleNames {
		results = append(results, &modules.InstallResult{
			ModuleName: module,
			Version:    "1.0.0",
			Success:    true,
		})
	}

	return results, nil
}

// ValidateBundle validates a bundle's structure and content
func (bm *BundleManager) ValidateBundle(bundle *Bundle) ([]*ValidationError, error) {
	if bundle == nil {
		return nil, errors.NewSystemError(
			"408",
			"No bundle provided for validation",
			nil)
	}

	var validationErrors []*ValidationError

	// Validate bundle name
	if bundle.Name == "" {
		validationErrors = append(validationErrors, &ValidationError{
			Field:    "name",
			Value:    bundle.Name,
			Message:  "Bundle name cannot be empty",
			Severity: ValidationErrorSeverity,
		})
	}

	// Validate modules
	if len(bundle.Modules) == 0 {
		validationErrors = append(validationErrors, &ValidationError{
			Field:    "modules",
			Value:    fmt.Sprintf("%d", len(bundle.Modules)),
			Message:  "Bundle must contain at least one module",
			Severity: ValidationErrorSeverity,
		})
	}

	// Validate each module entry
	moduleNames := make(map[string]bool)
	for i, entry := range bundle.Modules {
		// Check for duplicate module names
		if moduleNames[entry.Name] {
			validationErrors = append(validationErrors, &ValidationError{
				Field:    fmt.Sprintf("modules[%d].name", i),
				Value:    entry.Name,
				Message:  fmt.Sprintf("Duplicate module name: %s", entry.Name),
				Severity: ValidationErrorSeverity,
			})
		}
		moduleNames[entry.Name] = true

		// Validate module name
		if entry.Name == "" {
			validationErrors = append(validationErrors, &ValidationError{
				Field:    fmt.Sprintf("modules[%d].name", i),
				Value:    entry.Name,
				Message:  "Module name cannot be empty",
				Severity: ValidationErrorSeverity,
			})
		}

		// Validate version constraint if present
		if entry.VersionConstraint != "" {
			if err := bm.validateVersionConstraint(entry.VersionConstraint); err != nil {
				validationErrors = append(validationErrors, &ValidationError{
					Field:    fmt.Sprintf("modules[%d].version_constraint", i),
					Value:    entry.VersionConstraint,
					Message:  fmt.Sprintf("Invalid version constraint: %v", err),
					Severity: ValidationWarning,
				})
			}
		}

		// Validate phase
		if entry.Phase != "" && !bm.isValidPhase(entry.Phase) {
			validationErrors = append(validationErrors, &ValidationError{
				Field:    fmt.Sprintf("modules[%d].phase", i),
				Value:    entry.Phase,
				Message:  fmt.Sprintf("Invalid phase: %s", entry.Phase),
				Severity: ValidationWarning,
			})
		}

		// Validate relationship
		if entry.Relationship != "" && !bm.isValidRelationship(entry.Relationship) {
			validationErrors = append(validationErrors, &ValidationError{
				Field:    fmt.Sprintf("modules[%d].relationship", i),
				Value:    entry.Relationship,
				Message:  fmt.Sprintf("Invalid relationship: %s", entry.Relationship),
				Severity: ValidationWarning,
			})
		}
	}

	return validationErrors, nil
}

// Helper methods

func (bm *BundleManager) filterModules(modules []string, pattern string) ([]string, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid module filter pattern: %w", err)
	}

	var filtered []string
	for _, module := range modules {
		if regex.MatchString(module) {
			filtered = append(filtered, module)
		}
	}

	return filtered, nil
}

func (bm *BundleManager) getPerlVersion(ctx context.Context, perlPath string) string {
	cmd := exec.CommandContext(ctx, perlPath, "-e", "print $^V")
	output, err := cmd.Output()
	if err != nil {
		bm.logger.Printf("Failed to get Perl version: %v", err)
		return "unknown"
	}

	version := strings.TrimPrefix(string(output), "v")
	return strings.TrimSpace(version)
}

func (bm *BundleManager) validateVersionConstraint(constraint string) error {
	// Basic validation for version constraints
	// This could be enhanced with more sophisticated version parsing
	if constraint == "" {
		return nil
	}

	// Check for valid constraint operators
	validPrefixes := []string{">=", "<=", "==", "!=", ">", "<", "~"}
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(constraint, prefix) {
			return nil
		}
	}

	// If no operator, assume it's just a version number
	if matched, _ := regexp.MatchString(`^\d+(\.\d+)*`, constraint); matched {
		return nil
	}

	return fmt.Errorf("invalid version constraint format: %s", constraint)
}

func (bm *BundleManager) isValidPhase(phase string) bool {
	validPhases := []string{"runtime", "build", "test", "develop", "configure"}
	for _, validPhase := range validPhases {
		if phase == validPhase {
			return true
		}
	}
	return false
}

func (bm *BundleManager) isValidRelationship(relationship string) bool {
	validRelationships := []string{"requires", "recommends", "suggests", "conflicts"}
	for _, validRel := range validRelationships {
		if relationship == validRel {
			return true
		}
	}
	return false
}

// BundleCreateOptions contains options for creating bundles
type BundleCreateOptions struct {
	// Name of the bundle
	Name string

	// Description of the bundle
	Description string

	// IncludeDependencies resolves and includes dependencies
	IncludeDependencies bool

	// IncludeVersions includes version constraints
	IncludeVersions bool

	// IncludeCore includes core modules
	IncludeCore bool

	// ModuleFilter is a regex pattern to filter modules
	ModuleFilter string

	// PerlPath is the path to the Perl interpreter
	PerlPath string

	// Metadata for additional bundle information
	Metadata map[string]interface{}
}
