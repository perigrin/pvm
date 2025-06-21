// ABOUTME: Dependency resolution and conflict handling during tool installation
// ABOUTME: Manages tool dependencies and prevents conflicts between global tools

package install

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"tamarou.com/pvm/internal/errors"
)

const (
	// Dependency error codes
	ErrDepsConflict   = "TOOL-DEPS-001"
	ErrDepsUnresolved = "TOOL-DEPS-002"
	ErrDepsCircular   = "TOOL-DEPS-003"
)

// DependencyInfo represents information about a tool dependency
type DependencyInfo struct {
	// Module name
	ModuleName string

	// Version constraint
	VersionConstraint string

	// Dependency type (requires, recommends, suggests, etc.)
	Type string

	// Whether this is a direct dependency
	Direct bool

	// Tools that depend on this module
	RequiredBy []string

	// Installed version (if any)
	InstalledVersion string

	// Whether the dependency is satisfied
	Satisfied bool
}

// ConflictInfo represents a dependency conflict
type ConflictInfo struct {
	// Module name with conflict
	ModuleName string

	// Conflicting version constraints
	Constraints []string

	// Tools that have conflicting requirements
	ConflictingTools []string

	// Suggested resolution
	Resolution string
}

// DependencyResolver handles dependency resolution for tools
type DependencyGraph struct {
	// All tools and their dependencies
	tools map[string]*ToolDependencies

	// Module to tools mapping
	moduleUsers map[string][]string

	// Storage for metadata access
	storage *ToolStorage
}

// ToolDependencies represents the dependencies of a single tool
type ToolDependencies struct {
	ToolName     string
	ModuleName   string
	Dependencies []*DependencyInfo
}

// NewDependencyGraph creates a new dependency resolver
func NewDependencyGraph(storage *ToolStorage) *DependencyGraph {
	return &DependencyGraph{
		tools:       make(map[string]*ToolDependencies),
		moduleUsers: make(map[string][]string),
		storage:     storage,
	}
}

// BuildGraph builds the dependency graph from installed tools
func (dg *DependencyGraph) BuildGraph(ctx context.Context) error {
	// Get all installed tools
	metadataList, err := dg.storage.ListTools()
	if err != nil {
		return errors.NewSystemError(ErrDepsUnresolved,
			"Failed to load installed tools", err)
	}

	// Build graph from installed tools
	for _, metadata := range metadataList {
		toolDeps := &ToolDependencies{
			ToolName:     metadata.ToolName,
			ModuleName:   metadata.ModuleName,
			Dependencies: []*DependencyInfo{},
		}

		// Convert dependency list to dependency info
		for _, depName := range metadata.Dependencies {
			depInfo := &DependencyInfo{
				ModuleName: depName,
				Type:       "requires",
				Direct:     true,
				RequiredBy: []string{metadata.ToolName},
				Satisfied:  true, // Assume satisfied if installed
			}
			toolDeps.Dependencies = append(toolDeps.Dependencies, depInfo)
		}

		dg.tools[metadata.ToolName] = toolDeps

		// Update module users mapping
		for _, depName := range metadata.Dependencies {
			dg.moduleUsers[depName] = append(dg.moduleUsers[depName], metadata.ToolName)
		}

		// The tool's main module is also used by this tool
		dg.moduleUsers[metadata.ModuleName] = append(dg.moduleUsers[metadata.ModuleName], metadata.ToolName)
	}

	return nil
}

// CheckConflicts checks for dependency conflicts when installing a new tool
func (dg *DependencyGraph) CheckConflicts(toolName string, dependencies []*DependencyInfo) ([]*ConflictInfo, error) {
	var conflicts []*ConflictInfo

	// Check each dependency for conflicts
	for _, dep := range dependencies {
		conflict := dg.checkModuleConflict(toolName, dep)
		if conflict != nil {
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts, nil
}

// checkModuleConflict checks if a module dependency conflicts with existing installations
func (dg *DependencyGraph) checkModuleConflict(toolName string, dep *DependencyInfo) *ConflictInfo {
	// Get tools that currently use this module
	users := dg.moduleUsers[dep.ModuleName]
	if len(users) == 0 {
		return nil // No conflict if no one uses it
	}

	// Check if any existing tools have conflicting version requirements
	var conflictingConstraints []string
	var conflictingTools []string

	for _, user := range users {
		if user == toolName {
			continue // Skip self
		}

		toolDeps := dg.tools[user]
		if toolDeps == nil {
			continue
		}

		// Find the dependency in this tool
		for _, existingDep := range toolDeps.Dependencies {
			if existingDep.ModuleName == dep.ModuleName {
				// Check for version constraint conflicts
				if existingDep.VersionConstraint != "" && dep.VersionConstraint != "" {
					if !dg.versionsCompatible(existingDep.VersionConstraint, dep.VersionConstraint) {
						conflictingConstraints = append(conflictingConstraints,
							fmt.Sprintf("%s (required by %s)", existingDep.VersionConstraint, user))
						conflictingTools = append(conflictingTools, user)
					}
				}
			}
		}
	}

	if len(conflictingConstraints) > 0 {
		return &ConflictInfo{
			ModuleName:       dep.ModuleName,
			Constraints:      append(conflictingConstraints, fmt.Sprintf("%s (required by %s)", dep.VersionConstraint, toolName)),
			ConflictingTools: append(conflictingTools, toolName),
			Resolution:       dg.suggestResolution(dep.ModuleName, conflictingConstraints),
		}
	}

	return nil
}

// versionsCompatible checks if two version constraints are compatible
func (dg *DependencyGraph) versionsCompatible(constraint1, constraint2 string) bool {
	// Simple compatibility check - in practice this would need more sophisticated logic
	// For now, we consider constraints compatible if they're the same or one is empty
	if constraint1 == "" || constraint2 == "" {
		return true
	}
	return constraint1 == constraint2
}

// suggestResolution suggests a resolution for dependency conflicts
func (dg *DependencyGraph) suggestResolution(moduleName string, conflictingConstraints []string) string {
	// Simple resolution suggestion - in practice this would be more sophisticated
	return fmt.Sprintf("Consider using a compatible version of %s that satisfies all constraints: %s",
		moduleName, strings.Join(conflictingConstraints, ", "))
}

// GetDependents returns tools that depend on a specific module
func (dg *DependencyGraph) GetDependents(moduleName string) []string {
	return dg.moduleUsers[moduleName]
}

// CanRemoveTool checks if a tool can be safely removed
func (dg *DependencyGraph) CanRemoveTool(toolName string) (bool, []string, error) {
	toolDeps := dg.tools[toolName]
	if toolDeps == nil {
		return true, nil, nil // Tool not found, can "remove"
	}

	var blockers []string

	// Check if other tools depend on this tool's module
	dependents := dg.moduleUsers[toolDeps.ModuleName]
	for _, dependent := range dependents {
		if dependent != toolName {
			blockers = append(blockers, dependent)
		}
	}

	canRemove := len(blockers) == 0
	return canRemove, blockers, nil
}

// GetOrphanedModules returns modules that are no longer needed by any tool
func (dg *DependencyGraph) GetOrphanedModules() []string {
	var orphaned []string

	// Check each module to see if it's still needed
	allModules := make(map[string]bool)

	// Collect all modules from all tools
	for _, toolDeps := range dg.tools {
		allModules[toolDeps.ModuleName] = true
		for _, dep := range toolDeps.Dependencies {
			allModules[dep.ModuleName] = true
		}
	}

	// Check which modules have no users
	for module := range allModules {
		users := dg.moduleUsers[module]
		if len(users) == 0 {
			orphaned = append(orphaned, module)
		}
	}

	sort.Strings(orphaned)
	return orphaned
}

// RemoveToolFromGraph removes a tool from the dependency graph
func (dg *DependencyGraph) RemoveToolFromGraph(toolName string) {
	toolDeps := dg.tools[toolName]
	if toolDeps == nil {
		return
	}

	// Remove from module users mapping
	dg.removeFromModuleUsers(toolDeps.ModuleName, toolName)
	for _, dep := range toolDeps.Dependencies {
		dg.removeFromModuleUsers(dep.ModuleName, toolName)
	}

	// Remove from tools map
	delete(dg.tools, toolName)
}

// removeFromModuleUsers removes a tool from the module users list
func (dg *DependencyGraph) removeFromModuleUsers(moduleName, toolName string) {
	users := dg.moduleUsers[moduleName]
	var newUsers []string
	for _, user := range users {
		if user != toolName {
			newUsers = append(newUsers, user)
		}
	}

	if len(newUsers) == 0 {
		delete(dg.moduleUsers, moduleName)
	} else {
		dg.moduleUsers[moduleName] = newUsers
	}
}

// AddToolToGraph adds a tool to the dependency graph
func (dg *DependencyGraph) AddToolToGraph(toolName, moduleName string, dependencies []string) {
	toolDeps := &ToolDependencies{
		ToolName:     toolName,
		ModuleName:   moduleName,
		Dependencies: []*DependencyInfo{},
	}

	// Convert dependencies to dependency info
	for _, depName := range dependencies {
		depInfo := &DependencyInfo{
			ModuleName: depName,
			Type:       "requires",
			Direct:     true,
			RequiredBy: []string{toolName},
			Satisfied:  true,
		}
		toolDeps.Dependencies = append(toolDeps.Dependencies, depInfo)
	}

	dg.tools[toolName] = toolDeps

	// Update module users mapping
	dg.moduleUsers[moduleName] = append(dg.moduleUsers[moduleName], toolName)
	for _, depName := range dependencies {
		dg.moduleUsers[depName] = append(dg.moduleUsers[depName], toolName)
	}
}

// ValidateGraph performs sanity checks on the dependency graph
func (dg *DependencyGraph) ValidateGraph() error {
	// Check for circular dependencies (simplified check)
	for toolName := range dg.tools {
		if dg.hasCircularDependency(toolName, make(map[string]bool)) {
			return errors.NewSystemError(ErrDepsCircular,
				fmt.Sprintf("Circular dependency detected involving tool %s", toolName), nil)
		}
	}

	return nil
}

// hasCircularDependency checks for circular dependencies (simplified implementation)
func (dg *DependencyGraph) hasCircularDependency(toolName string, visited map[string]bool) bool {
	if visited[toolName] {
		return true
	}

	visited[toolName] = true
	defer delete(visited, toolName)

	toolDeps := dg.tools[toolName]
	if toolDeps == nil {
		return false
	}

	// Check dependencies (simplified - would need more sophisticated logic for real circular detection)
	for _, dep := range toolDeps.Dependencies {
		dependents := dg.moduleUsers[dep.ModuleName]
		for _, dependent := range dependents {
			if dependent != toolName && dg.hasCircularDependency(dependent, visited) {
				return true
			}
		}
	}

	return false
}

// GenerateReport generates a dependency report
func (dg *DependencyGraph) GenerateReport() string {
	var report strings.Builder

	report.WriteString("Tool Dependency Report\n")
	report.WriteString("=====================\n\n")

	// List all tools and their dependencies
	report.WriteString("Installed Tools:\n")
	for toolName, toolDeps := range dg.tools {
		report.WriteString(fmt.Sprintf("- %s (module: %s)\n", toolName, toolDeps.ModuleName))
		if len(toolDeps.Dependencies) > 0 {
			report.WriteString("  Dependencies:\n")
			for _, dep := range toolDeps.Dependencies {
				report.WriteString(fmt.Sprintf("    - %s\n", dep.ModuleName))
			}
		}
		report.WriteString("\n")
	}

	// List orphaned modules
	orphaned := dg.GetOrphanedModules()
	if len(orphaned) > 0 {
		report.WriteString("Orphaned Modules:\n")
		for _, module := range orphaned {
			report.WriteString(fmt.Sprintf("- %s\n", module))
		}
		report.WriteString("\n")
	}

	return report.String()
}
