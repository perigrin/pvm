// ABOUTME: Implementation of the dependency resolver
// ABOUTME: Core functionality for resolving module dependencies

package deps

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// defaultResolver implements the DependencyResolver interface
type defaultResolver struct {
	cache *DependencyCache
}

// NewDefaultResolver creates a new default dependency resolver
func NewDefaultResolver(cacheDir string, cacheTTL int) (DependencyResolver, error) {
	var cache *DependencyCache
	var err error

	if cacheDir != "" && cacheTTL > 0 {
		cache, err = NewDependencyCache(cacheDir, cacheTTL)
		if err != nil {
			return nil, err
		}
	}

	return &defaultResolver{
		cache: cache,
	}, nil
}

// ResolveDependencies resolves dependencies for a module
func (r *defaultResolver) ResolveDependencies(ctx context.Context, moduleName string, options *DependencyResolutionOptions) (*DependencyResolutionResult, error) {
	fmt.Printf("[TRACE] defaultResolver.ResolveDependencies called for module: %s\n", moduleName)
	// Use defaults if options not provided
	if options == nil {
		options = &DependencyResolutionOptions{
			IncludeCore:  false,
			IncludeTest:  false,
			IncludeBuild: true,
			IncludeDev:   false,
			MaxDepth:     0, // No limit
		}
	}

	// Ensure we have a provider
	if options.Provider == nil {
		return nil, errors.NewSystemError(
			ErrResolveFailure,
			"No CPAN provider specified for dependency resolution",
			nil)
	}

	// Check cache if enabled
	if r.cache != nil && options.UseCache {
		optionsKey, err := serializeOptions(options)
		if err != nil {
			log.Warnf("Failed to serialize options for cache key: %v", err)
		} else {
			if cachedResult, found := r.cache.Get(moduleName, optionsKey); found {
				return cachedResult, nil
			}
		}
	}

	// Create result structure
	result := &DependencyResolutionResult{
		Modules:            make(map[string]*DependencyNode),
		VersionConstraints: make(map[string]map[string]bool),
		Conflicts:          []*DependencyConflict{},
		Warnings:           []string{},
	}

	// Get module info for the root module
	moduleInfo, err := options.Provider.GetModuleInfo(ctx, moduleName)
	if err != nil {
		return nil, errors.NewSystemError(
			ErrModuleNotFound,
			fmt.Sprintf("Failed to get information for module %s", moduleName),
			err)
	}

	// Create the root node
	// Strip .pm suffix if present for consistency
	rootName := strings.TrimSuffix(moduleInfo.Name, ".pm")
	root := &DependencyNode{
		Name:    rootName,
		Version: moduleInfo.Version,
		IsRoot:  true,
		Parent:  nil,
		Path:    []string{rootName},
		Depth:   0,
	}

	// Add root to the result
	result.Root = root
	result.Modules[root.Name] = root

	// Recursively resolve dependencies
	err = r.resolveNodeDependencies(ctx, root, options, result)
	if err != nil {
		return result, err
	}

	// Cache result if enabled
	if r.cache != nil && options.UseCache {
		optionsKey, err := serializeOptions(options)
		if err != nil {
			log.Warnf("Failed to serialize options for cache key: %v", err)
		} else {
			if err := r.cache.Set(moduleName, optionsKey, result); err != nil {
				log.Warnf("Failed to cache dependency resolution result: %v", err)
			}
		}
	}

	return result, nil
}

// resolveNodeDependencies recursively resolves dependencies for a node
func (r *defaultResolver) resolveNodeDependencies(ctx context.Context, node *DependencyNode, options *DependencyResolutionOptions, result *DependencyResolutionResult) error {
	// Check max depth if specified
	if options.MaxDepth > 0 && node.Depth >= options.MaxDepth {
		if options.Verbose {
			log.Infof("Max depth reached for %s at depth %d", node.Name, node.Depth)
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Max depth reached for %s at depth %d", node.Name, node.Depth))
		}
		return nil
	}

	// If this is not the root node, we need to resolve its version based on constraints
	if !node.IsRoot {
		moduleInfo, err := options.Provider.GetModuleInfo(ctx, node.Name)
		if err != nil {
			if options.Verbose {
				log.Warnf("Failed to get information for module %s: %v", node.Name, err)
			}
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Failed to get information for module %s: %v", node.Name, err))
			return nil // Continue with other dependencies
		}

		// Set the actual version from the module info
		node.Version = moduleInfo.Version

		// If we have a version constraint, verify it's satisfied
		if node.VersionConstraint != "" {
			satisfied, err := r.CheckVersionConstraint(moduleInfo.Version, node.VersionConstraint)
			if err != nil {
				if options.Verbose {
					log.Warnf("Error checking version constraint for %s: %v", node.Name, err)
				}
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Error checking version constraint for %s: %v", node.Name, err))
			} else if !satisfied {
				// Record the conflict
				conflict := findOrCreateConflict(result, node.Name)
				if conflict.Requirements == nil {
					conflict.Requirements = make(map[string][]string)
				}

				var parentName string
				if node.Parent != nil {
					parentName = node.Parent.Name
				} else {
					parentName = "root"
				}

				conflict.Requirements[node.VersionConstraint] = append(
					conflict.Requirements[node.VersionConstraint], parentName)

				if options.Verbose {
					log.Warnf("Version conflict for %s: required %s, got %s",
						node.Name, node.VersionConstraint, moduleInfo.Version)
				}
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Version conflict for %s: required %s, got %s",
						node.Name, node.VersionConstraint, moduleInfo.Version))
			}
		}
	}

	// Get module info from provider (already have it for non-root nodes)
	var moduleInfo *cpan.ModuleInfo
	var err error
	if node.IsRoot {
		moduleInfo, err = options.Provider.GetModuleInfo(ctx, node.Name)
		if err != nil {
			return errors.NewSystemError(
				ErrModuleNotFound,
				fmt.Sprintf("Failed to get information for module %s", node.Name),
				err)
		}
	} else {
		// We should already have the module info from above, but let's get it again to be safe
		moduleInfo, err = options.Provider.GetModuleInfo(ctx, node.Name)
		if err != nil {
			if options.Verbose {
				log.Warnf("Failed to get information for module %s: %v", node.Name, err)
			}
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Failed to get information for module %s: %v", node.Name, err))
			return nil // Continue with other dependencies
		}
	}

	// Process dependencies
	for _, dep := range moduleInfo.Dependencies {
		// Skip based on phase if configured
		if !shouldIncludeDependency(dep, options) {
			if options.Verbose {
				log.Debugf("Skipping dependency %s for %s (phase: %s)",
					dep.Name, node.Name, dep.Phase)
			}
			continue
		}

		// Check if the dependency is already in the tree
		// Strip .pm suffix if present for consistency
		normalizedDepName := strings.TrimSuffix(dep.Name, ".pm")
		fmt.Printf("[TRACE] Processing dependency %s (normalized: %s) with constraint %s from parent %s\n", dep.Name, normalizedDepName, dep.Version, node.Name)
		existingNode, exists := result.Modules[normalizedDepName]
		fmt.Printf("[TRACE] Dependency %s exists in tree: %t\n", normalizedDepName, exists)
		if exists {
			// Check for version constraints
			if dep.Version != "" {
				// Record the version constraint
				if _, ok := result.VersionConstraints[dep.Name]; !ok {
					result.VersionConstraints[dep.Name] = make(map[string]bool)
				}
				result.VersionConstraints[dep.Name][dep.Version] = true

				// Check if the constraint is compatible with the existing version
				fmt.Printf("[TRACE] Checking version constraint: %s (existing) against %s (constraint) for module %s\n", existingNode.Version, dep.Version, dep.Name)
				isCompatible, err := r.CheckVersionConstraint(existingNode.Version, dep.Version)
				fmt.Printf("[TRACE] Version constraint check result: isCompatible=%t, err=%v\n", isCompatible, err)
				if err != nil {
					if options.Verbose {
						log.Warnf("Error checking version constraint for %s: %v", dep.Name, err)
					}
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("Error checking version constraint for %s: %v", dep.Name, err))
				} else if !isCompatible {
					fmt.Printf("[TRACE] CONFLICT DETECTED: Module %s version %s does not satisfy constraint %s required by %s\n", dep.Name, existingNode.Version, dep.Version, node.Name)
					// Add conflict
					conflict := findOrCreateConflict(result, dep.Name)
					if conflict.Requirements == nil {
						conflict.Requirements = make(map[string][]string)
					}
					conflict.Requirements[dep.Version] = append(conflict.Requirements[dep.Version], node.Name)

					if options.Verbose {
						log.Warnf("Version conflict for %s: required %s by %s, got %s",
							dep.Name, dep.Version, node.Name, existingNode.Version)
					}
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("Version conflict for %s: required %s by %s, got %s",
							dep.Name, dep.Version, node.Name, existingNode.Version))
				} else {
					fmt.Printf("[TRACE] No conflict: Module %s version %s satisfies constraint %s\n", dep.Name, existingNode.Version, dep.Version)
				}
			}

			// Add a reference from this parent to the existing node
			node.Children = append(node.Children, existingNode)

			// Skip further processing for this module as it's already in the tree
			continue
		}

		// Check if this is a core module
		isCore := false
		if options.PerlVersion != "" {
			// Try to determine if it's a core module
			isCore, err = options.Provider.IsCoreModule(ctx, dep.Name, options.PerlVersion)
			if err != nil {
				if options.Verbose {
					log.Debugf("Failed to check if %s is a core module: %v", dep.Name, err)
				}
				// Assume it's not core if we can't determine
				isCore = false
			}
		}

		// Skip core modules if configured
		if isCore && !options.IncludeCore {
			if options.Verbose {
				log.Debugf("Skipping core module %s", dep.Name)
			}
			continue
		}

		// Create a new dependency node
		// Strip .pm suffix if present for consistency
		depName := strings.TrimSuffix(dep.Name, ".pm")
		depNode := &DependencyNode{
			Name:              depName,
			Version:           "", // Will be filled in during recursive resolution
			VersionConstraint: dep.Version,
			Phase:             dep.Phase,
			Type:              dep.Type,
			IsCore:            isCore,
			Parent:            node,
			Children:          []*DependencyNode{},
			Path:              append(append([]string{}, node.Path...), depName),
			Depth:             node.Depth + 1,
		}

		// Check for circular dependencies
		if r.hasCircularDependency(depNode) {
			if options.Verbose {
				log.Warnf("Circular dependency detected: %s", strings.Join(depNode.Path, " -> "))
			}
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Circular dependency detected: %s", strings.Join(depNode.Path, " -> ")))
			continue
		}

		// Add the dependency to its parent's children
		node.Children = append(node.Children, depNode)

		// Add to the modules map
		result.Modules[depName] = depNode

		// Record the version constraint for this new module
		if dep.Version != "" {
			fmt.Printf("[TRACE] Recording constraint %s for new module %s\n", dep.Version, dep.Name)
			if _, ok := result.VersionConstraints[dep.Name]; !ok {
				result.VersionConstraints[dep.Name] = make(map[string]bool)
			}
			result.VersionConstraints[dep.Name][dep.Version] = true
		}

		// Recursively resolve this dependency's dependencies
		err = r.resolveNodeDependencies(ctx, depNode, options, result)
		if err != nil {
			return err
		}
	}

	return nil
}

// hasCircularDependency checks if a node creates a circular dependency
func (r *defaultResolver) hasCircularDependency(node *DependencyNode) bool {
	// Check if the module name appears more than once in the path
	nameCount := make(map[string]int)
	for _, name := range node.Path {
		nameCount[name]++
		if nameCount[name] > 1 {
			return true
		}
	}
	return false
}

// CheckVersionConstraint checks if a version satisfies a constraint
func (r *defaultResolver) CheckVersionConstraint(version, constraint string) (bool, error) {
	return CheckVersionConstraint(version, constraint)
}

// GetFlattenedDependencies returns a flattened list of dependencies
func (r *defaultResolver) GetFlattenedDependencies(result *DependencyResolutionResult) []*DependencyNode {
	// Convert map to slice
	deps := make([]*DependencyNode, 0, len(result.Modules))
	for _, node := range result.Modules {
		deps = append(deps, node)
	}
	return deps
}

// PrintDependencyTree returns a formatted string representation of the dependency tree
func (r *defaultResolver) PrintDependencyTree(node *DependencyNode) string {
	if node == nil {
		return ""
	}

	builder := strings.Builder{}
	r.printNode(&builder, node, "", true)
	return builder.String()
}

// printNode prints a node and its children to the string builder
func (r *defaultResolver) printNode(builder *strings.Builder, node *DependencyNode, prefix string, isLast bool) {
	// Add the current node
	if node.IsRoot {
		// Root node gets special treatment
		fmt.Fprintf(builder, "%s\n", node.Name)
		prefix = ""
	} else {
		branch := "├── "
		if isLast {
			branch = "└── "
		}

		// Print the node with its version constraint if available
		nodeName := node.Name
		if node.VersionConstraint != "" {
			nodeName = fmt.Sprintf("%s (%s)", node.Name, node.VersionConstraint)
		}
		fmt.Fprintf(builder, "%s%s%s\n", prefix, branch, nodeName)

		// Update the prefix for children
		if isLast {
			prefix += "    "
		} else {
			prefix += "│   "
		}
	}

	// Print children
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		r.printNode(builder, child, prefix, isLastChild)
	}
}

// serializeOptions serializes options to a string for cache key generation
func serializeOptions(options *DependencyResolutionOptions) (string, error) {
	// Create a simplified options object for serialization
	optStruct := struct {
		IncludeCore  bool   `json:"include_core"`
		IncludeTest  bool   `json:"include_test"`
		IncludeBuild bool   `json:"include_build"`
		IncludeDev   bool   `json:"include_dev"`
		MaxDepth     int    `json:"max_depth"`
		PerlVersion  string `json:"perl_version"`
		ProviderName string `json:"provider_name"`
	}{
		IncludeCore:  options.IncludeCore,
		IncludeTest:  options.IncludeTest,
		IncludeBuild: options.IncludeBuild,
		IncludeDev:   options.IncludeDev,
		MaxDepth:     options.MaxDepth,
		PerlVersion:  options.PerlVersion,
	}

	// Add provider name if available
	if options.Provider != nil {
		optStruct.ProviderName = options.Provider.Name()
	}

	// Serialize to JSON
	data, err := json.Marshal(optStruct)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
