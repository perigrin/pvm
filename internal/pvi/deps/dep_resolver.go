// ABOUTME: Dependency resolver for PVI component
// ABOUTME: Defines interfaces and types for dependency resolution

package deps

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/errors"
)

// Error code constants
const (
	ErrResolveFailure        = "PVI-3001" // General dependency resolution failure
	ErrCircularDependency    = "PVI-3002" // Circular dependency detected
	ErrVersionConflict       = "PVI-3003" // Version conflict between dependencies
	ErrModuleNotFound        = "PVI-3004" // Module not found in the registry
	ErrInvalidVersionPattern = "PVI-3005" // Invalid version pattern
)

// DependencyNode represents a module with its dependencies in the dependency tree
type DependencyNode struct {
	// Name of the module
	Name string `json:"name"`

	// Version of the module (resolved version, not constraint)
	Version string `json:"version"`

	// VersionConstraint is the original version constraint string (e.g., ">= 1.0")
	VersionConstraint string `json:"version_constraint"`

	// Phase indicates when the dependency is needed (e.g., runtime, build, test)
	Phase string `json:"phase"`

	// Type indicates the type of dependency (e.g., requires, recommends)
	Type string `json:"type"`

	// IsCore indicates whether this module is part of the Perl core
	IsCore bool `json:"is_core"`

	// IsRoot indicates whether this is the root module being resolved
	IsRoot bool `json:"is_root"`

	// Parent module that requires this dependency (nil for root)
	Parent *DependencyNode `json:"-"` // Exclude from JSON to avoid cycles

	// Children are the dependencies of this module
	Children []*DependencyNode `json:"children"`

	// The path to this module in the dependency tree for cycle detection
	Path []string `json:"path"`

	// Depth in the dependency tree
	Depth int `json:"depth"`
}

// MarshalJSON provides custom JSON marshaling for DependencyNode
// that avoids circular references
func (n *DependencyNode) MarshalJSON() ([]byte, error) {
	// Create a simplified representation that doesn't include Parent
	type NodeAlias struct {
		Name              string   `json:"name"`
		Version           string   `json:"version"`
		VersionConstraint string   `json:"version_constraint,omitempty"`
		Phase             string   `json:"phase,omitempty"`
		Type              string   `json:"type,omitempty"`
		IsCore            bool     `json:"is_core,omitempty"`
		IsRoot            bool     `json:"is_root,omitempty"`
		ChildrenNames     []string `json:"children,omitempty"`
		Path              []string `json:"path,omitempty"`
		Depth             int      `json:"depth"`
	}

	// Create a simplified view with just child names
	childrenNames := make([]string, 0, len(n.Children))
	for _, child := range n.Children {
		childrenNames = append(childrenNames, child.Name)
	}

	alias := &NodeAlias{
		Name:              n.Name,
		Version:           n.Version,
		VersionConstraint: n.VersionConstraint,
		Phase:             n.Phase,
		Type:              n.Type,
		IsCore:            n.IsCore,
		IsRoot:            n.IsRoot,
		ChildrenNames:     childrenNames,
		Path:              n.Path,
		Depth:             n.Depth,
	}

	return json.Marshal(alias)
}

// DependencyResolutionOptions contains options for dependency resolution
type DependencyResolutionOptions struct {
	// Provider to use for retrieving module information
	Provider cpan.Provider

	// IncludeCore indicates whether to include core modules in the resolution
	IncludeCore bool

	// IncludeTest indicates whether to include test dependencies in the resolution
	IncludeTest bool

	// IncludeBuild indicates whether to include build dependencies in the resolution
	IncludeBuild bool

	// IncludeDev indicates whether to include development dependencies in the resolution
	IncludeDev bool

	// MaxDepth is the maximum depth to traverse in the dependency tree (0 means no limit)
	MaxDepth int

	// Verbose enables detailed logging during resolution
	Verbose bool

	// UseCache indicates whether to use cached dependencies
	UseCache bool

	// CacheTTL is the time-to-live for cached dependencies in hours
	CacheTTL int

	// CacheDir is the directory to use for caching dependencies
	CacheDir string

	// PerlVersion is the version of Perl to use for core module detection
	PerlVersion string
}

// DependencyResolutionResult contains the result of dependency resolution
type DependencyResolutionResult struct {
	// Root node of the dependency tree
	Root *DependencyNode

	// All modules in the dependency tree (flattened, unique by name)
	Modules map[string]*DependencyNode

	// Map of module names to sets of version constraints required by different dependencies
	VersionConstraints map[string]map[string]bool

	// Conflicts encountered during resolution (if any)
	Conflicts []*DependencyConflict

	// Warnings encountered during resolution (if any)
	Warnings []string
}

// DependencyConflict represents a conflict between different version requirements
type DependencyConflict struct {
	// Module name that has conflicting requirements
	Module string

	// Requirements is a map of version constraints to requiring modules
	Requirements map[string][]string
}

// DependencyResolver interfaces with CPAN to resolve module dependencies
type DependencyResolver interface {
	// ResolveDependencies resolves dependencies for a module
	ResolveDependencies(ctx context.Context, moduleName string, options *DependencyResolutionOptions) (*DependencyResolutionResult, error)

	// CheckVersionConstraint checks if a version satisfies a constraint
	CheckVersionConstraint(version, constraint string) (bool, error)

	// GetFlattenedDependencies returns a flattened list of dependencies
	GetFlattenedDependencies(result *DependencyResolutionResult) []*DependencyNode

	// PrintDependencyTree prints a visual representation of the dependency tree
	PrintDependencyTree(node *DependencyNode) string
}

// NewDependencyResolver creates a new dependency resolver
func NewDependencyResolver() DependencyResolver {
	return &dependencyResolver{}
}

// dependencyResolver implements the DependencyResolver interface
type dependencyResolver struct {
	// Add fields as needed for implementation
}

// ResolveDependencies resolves dependencies for a module
func (r *dependencyResolver) ResolveDependencies(ctx context.Context, moduleName string, options *DependencyResolutionOptions) (*DependencyResolutionResult, error) {
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
	root := &DependencyNode{
		Name:    moduleInfo.Name,
		Version: moduleInfo.Version,
		IsRoot:  true,
		Parent:  nil,
		Path:    []string{moduleInfo.Name},
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

	return result, nil
}

// resolveNodeDependencies recursively resolves dependencies for a node
func (r *dependencyResolver) resolveNodeDependencies(ctx context.Context, node *DependencyNode, options *DependencyResolutionOptions, result *DependencyResolutionResult) error {
	// Check max depth if specified
	if options.MaxDepth > 0 && node.Depth >= options.MaxDepth {
		if options.Verbose {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Max depth reached for %s at depth %d", node.Name, node.Depth))
		}
		return nil
	}

	// Get module info from provider
	moduleInfo, err := options.Provider.GetModuleInfo(ctx, node.Name)
	if err != nil {
		return errors.NewSystemError(
			ErrModuleNotFound,
			fmt.Sprintf("Failed to get information for module %s", node.Name),
			err)
	}

	// Process dependencies
	for _, dep := range moduleInfo.Dependencies {
		// Skip based on phase if configured
		if !shouldIncludeDependency(dep, options) {
			continue
		}

		// Check if the dependency is already in the tree
		existingNode, exists := result.Modules[dep.Name]
		if exists {
			// Check for version constraints
			if dep.Version != "" {
				// Record the version constraint
				if _, ok := result.VersionConstraints[dep.Name]; !ok {
					result.VersionConstraints[dep.Name] = make(map[string]bool)
				}
				result.VersionConstraints[dep.Name][dep.Version] = true

				// Check if the constraint is compatible with the existing version
				isCompatible, err := r.CheckVersionConstraint(existingNode.Version, dep.Version)
				if err != nil {
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("Failed to check version constraint for %s: %v", dep.Name, err))
				} else if !isCompatible {
					// Add conflict
					conflict := findOrCreateConflict(result, dep.Name)
					if conflict.Requirements == nil {
						conflict.Requirements = make(map[string][]string)
					}
					conflict.Requirements[dep.Version] = append(conflict.Requirements[dep.Version], node.Name)
				}
			}

			// Skip further processing for this module as it's already in the tree
			continue
		}

		// Create a new dependency node
		depNode := &DependencyNode{
			Name:              dep.Name,
			Version:           "", // Will be filled in later
			VersionConstraint: dep.Version,
			Phase:             dep.Phase,
			Type:              dep.Type,
			IsCore:            dep.IsCore,
			Parent:            node,
			Children:          []*DependencyNode{},
			Path:              append(append([]string{}, node.Path...), dep.Name),
			Depth:             node.Depth + 1,
		}

		// Check for circular dependencies
		if r.hasCircularDependency(depNode) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Circular dependency detected: %s", strings.Join(depNode.Path, " -> ")))
			continue
		}

		// Add the dependency to its parent's children
		node.Children = append(node.Children, depNode)

		// Add to the modules map
		result.Modules[dep.Name] = depNode

		// Recursively resolve this dependency's dependencies
		err = r.resolveNodeDependencies(ctx, depNode, options, result)
		if err != nil {
			return err
		}
	}

	return nil
}

// hasCircularDependency checks if a node creates a circular dependency
func (r *dependencyResolver) hasCircularDependency(node *DependencyNode) bool {
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

// shouldIncludeDependency determines if a dependency should be included based on options
func shouldIncludeDependency(dep *cpan.Dependency, options *DependencyResolutionOptions) bool {
	// Skip core modules if configured
	if dep.IsCore && !options.IncludeCore {
		return false
	}

	// Filter based on phase
	switch dep.Phase {
	case "runtime":
		return true // Always include runtime dependencies
	case "test":
		return options.IncludeTest
	case "build":
		return options.IncludeBuild
	case "develop":
		return options.IncludeDev
	default:
		return true // Include unknown phases by default
	}
}

// findOrCreateConflict finds an existing conflict or creates a new one
func findOrCreateConflict(result *DependencyResolutionResult, moduleName string) *DependencyConflict {
	for _, conflict := range result.Conflicts {
		if conflict.Module == moduleName {
			return conflict
		}
	}

	// Create a new conflict
	conflict := &DependencyConflict{
		Module:       moduleName,
		Requirements: make(map[string][]string),
	}
	result.Conflicts = append(result.Conflicts, conflict)
	return conflict
}

// CheckVersionConstraint checks if a version satisfies a constraint
func (r *dependencyResolver) CheckVersionConstraint(version, constraint string) (bool, error) {
	// This is a placeholder - a real implementation would parse and evaluate the constraint
	// For example: ">= 1.0", "< 2.0", "== 1.2.3", etc.
	return true, nil // Always return true for now
}

// GetFlattenedDependencies returns a flattened list of dependencies
func (r *dependencyResolver) GetFlattenedDependencies(result *DependencyResolutionResult) []*DependencyNode {
	// Convert map to slice
	deps := make([]*DependencyNode, 0, len(result.Modules))
	for _, node := range result.Modules {
		deps = append(deps, node)
	}
	return deps
}

// PrintDependencyTree returns a formatted string representation of the dependency tree
func (r *dependencyResolver) PrintDependencyTree(node *DependencyNode) string {
	if node == nil {
		return ""
	}

	builder := strings.Builder{}
	r.printNode(&builder, node, "", true)
	return builder.String()
}

// printNode prints a node and its children to the string builder
func (r *dependencyResolver) printNode(builder *strings.Builder, node *DependencyNode, prefix string, isLast bool) {
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
