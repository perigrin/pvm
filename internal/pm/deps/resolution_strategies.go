// ABOUTME: Resolution strategies for advanced dependency conflict handling
// ABOUTME: Implements different approaches to resolve version conflicts

package deps

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// ResolutionStrategy defines the interface for different resolution approaches
type ResolutionStrategy interface {
	// Name returns the strategy name
	Name() string

	// Description returns a description of the strategy
	Description() string

	// Resolve attempts to resolve conflicts using this strategy
	Resolve(ctx context.Context, conflicts []*DependencyConflict, options *StrategyOptions) (*StrategyResult, error)
}

// StrategyOptions contains configuration for resolution strategies
type StrategyOptions struct {
	Provider          cpan.Provider
	PinnedVersions    map[string]string
	ExcludedVersions  map[string]map[string]bool
	PreferredVersions map[string]string
	MaxRetries        int
	Timeout           time.Duration
	Verbose           bool
	AllowDowngrades   bool
	PreferStable      bool
}

// StrategyResult contains the outcome of a resolution strategy
type StrategyResult struct {
	Success             bool
	ResolvedVersions    map[string]string
	UnresolvedConflicts []*DependencyConflict
	Explanation         string
	Warnings            []string
	Metrics             StrategyMetrics
}

// StrategyMetrics tracks strategy performance
type StrategyMetrics struct {
	StartTime          time.Time
	EndTime            time.Time
	ConflictsProcessed int
	ConflictsResolved  int
	VersionsEvaluated  int
	BacktrackAttempts  int
}

// FailFastStrategy fails immediately on any conflict
type FailFastStrategy struct{}

// MinimalUpgradeStrategy resolves conflicts with minimal version changes
type MinimalUpgradeStrategy struct{}

// MaximalUpgradeStrategy resolves conflicts with maximal compatible versions
type MaximalUpgradeStrategy struct{}

// BacktrackStrategy uses backtracking to find optimal solutions
type BacktrackStrategy struct {
	maxDepth int
}

// HybridStrategy combines multiple strategies for best results
type HybridStrategy struct {
	strategies []ResolutionStrategy
}

// NewFailFastStrategy creates a new fail-fast strategy
func NewFailFastStrategy() ResolutionStrategy {
	return &FailFastStrategy{}
}

// NewMinimalUpgradeStrategy creates a new minimal upgrade strategy
func NewMinimalUpgradeStrategy() ResolutionStrategy {
	return &MinimalUpgradeStrategy{}
}

// NewMaximalUpgradeStrategy creates a new maximal upgrade strategy
func NewMaximalUpgradeStrategy() ResolutionStrategy {
	return &MaximalUpgradeStrategy{}
}

// NewBacktrackStrategy creates a new backtracking strategy
func NewBacktrackStrategy(maxDepth int) ResolutionStrategy {
	return &BacktrackStrategy{maxDepth: maxDepth}
}

// NewHybridStrategy creates a new hybrid strategy
func NewHybridStrategy(strategies ...ResolutionStrategy) ResolutionStrategy {
	return &HybridStrategy{strategies: strategies}
}

// FailFastStrategy implementation

func (s *FailFastStrategy) Name() string {
	return "fail-fast"
}

func (s *FailFastStrategy) Description() string {
	return "Fails immediately when conflicts are detected, providing detailed error information"
}

func (s *FailFastStrategy) Resolve(ctx context.Context, conflicts []*DependencyConflict, options *StrategyOptions) (*StrategyResult, error) {
	metrics := StrategyMetrics{
		StartTime:          time.Now(),
		ConflictsProcessed: len(conflicts),
	}

	if len(conflicts) == 0 {
		return &StrategyResult{
			Success:     true,
			Explanation: "No conflicts to resolve",
			Metrics:     metrics,
		}, nil
	}

	// Generate detailed conflict explanations
	var explanations []string
	for _, conflict := range conflicts {
		explanation := fmt.Sprintf("Module %s has conflicting version requirements:", conflict.Module)
		for version, requiring := range conflict.Requirements {
			explanation += fmt.Sprintf("\n  - Version %s required by: %s", version, strings.Join(requiring, ", "))
		}
		explanations = append(explanations, explanation)
	}

	metrics.EndTime = time.Now()

	return &StrategyResult{
			Success:             false,
			UnresolvedConflicts: conflicts,
			Explanation:         fmt.Sprintf("Conflicts detected in %d modules:\n%s", len(conflicts), strings.Join(explanations, "\n\n")),
			Metrics:             metrics,
		}, errors.NewSystemError(
			ErrVersionConflict,
			fmt.Sprintf("Fail-fast strategy detected %d conflicts", len(conflicts)),
			nil)
}

// MinimalUpgradeStrategy implementation

func (s *MinimalUpgradeStrategy) Name() string {
	return "minimal-upgrade"
}

func (s *MinimalUpgradeStrategy) Description() string {
	return "Resolves conflicts by choosing the minimal version that satisfies all constraints"
}

func (s *MinimalUpgradeStrategy) Resolve(ctx context.Context, conflicts []*DependencyConflict, options *StrategyOptions) (*StrategyResult, error) {
	metrics := StrategyMetrics{
		StartTime:          time.Now(),
		ConflictsProcessed: len(conflicts),
	}

	resolvedVersions := make(map[string]string)
	var warnings []string
	var unresolvedConflicts []*DependencyConflict

	for _, conflict := range conflicts {
		// Check for pinned version first
		if pinnedVersion, isPinned := options.PinnedVersions[conflict.Module]; isPinned {
			if satisfiesAllConstraints(pinnedVersion, conflict, options) {
				resolvedVersions[conflict.Module] = pinnedVersion
				metrics.ConflictsResolved++
				continue
			} else {
				warning := fmt.Sprintf("Pinned version %s for %s does not satisfy all constraints", pinnedVersion, conflict.Module)
				warnings = append(warnings, warning)
				unresolvedConflicts = append(unresolvedConflicts, conflict)
				continue
			}
		}

		// Get available versions
		versions, err := options.Provider.GetModuleVersions(ctx, conflict.Module)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Failed to get versions for %s: %v", conflict.Module, err))
			unresolvedConflicts = append(unresolvedConflicts, conflict)
			continue
		}

		metrics.VersionsEvaluated += len(versions)

		// Filter excluded versions
		filteredVersions := filterExcludedVersions(conflict.Module, versions, options.ExcludedVersions)

		// Sort versions in ascending order (minimal first)
		sort.Sort(VersionSlice(filteredVersions))

		// Find minimal compatible version
		var resolvedVersion string
		for _, version := range filteredVersions {
			if satisfiesAllConstraints(version, conflict, options) {
				resolvedVersion = version
				break
			}
		}

		if resolvedVersion != "" {
			resolvedVersions[conflict.Module] = resolvedVersion
			metrics.ConflictsResolved++
		} else {
			unresolvedConflicts = append(unresolvedConflicts, conflict)
		}
	}

	metrics.EndTime = time.Now()

	success := len(unresolvedConflicts) == 0
	explanation := fmt.Sprintf("Minimal upgrade strategy resolved %d/%d conflicts", metrics.ConflictsResolved, metrics.ConflictsProcessed)

	if options.Verbose {
		log.Infof("Minimal upgrade strategy: %s", explanation)
	}

	return &StrategyResult{
		Success:             success,
		ResolvedVersions:    resolvedVersions,
		UnresolvedConflicts: unresolvedConflicts,
		Explanation:         explanation,
		Warnings:            warnings,
		Metrics:             metrics,
	}, nil
}

// MaximalUpgradeStrategy implementation

func (s *MaximalUpgradeStrategy) Name() string {
	return "maximal-upgrade"
}

func (s *MaximalUpgradeStrategy) Description() string {
	return "Resolves conflicts by choosing the latest version that satisfies all constraints"
}

func (s *MaximalUpgradeStrategy) Resolve(ctx context.Context, conflicts []*DependencyConflict, options *StrategyOptions) (*StrategyResult, error) {
	metrics := StrategyMetrics{
		StartTime:          time.Now(),
		ConflictsProcessed: len(conflicts),
	}

	resolvedVersions := make(map[string]string)
	var warnings []string
	var unresolvedConflicts []*DependencyConflict

	for _, conflict := range conflicts {
		// Check for pinned version first
		if pinnedVersion, isPinned := options.PinnedVersions[conflict.Module]; isPinned {
			if satisfiesAllConstraints(pinnedVersion, conflict, options) {
				resolvedVersions[conflict.Module] = pinnedVersion
				metrics.ConflictsResolved++
				continue
			} else {
				warning := fmt.Sprintf("Pinned version %s for %s does not satisfy all constraints", pinnedVersion, conflict.Module)
				warnings = append(warnings, warning)
				unresolvedConflicts = append(unresolvedConflicts, conflict)
				continue
			}
		}

		// Get available versions
		versions, err := options.Provider.GetModuleVersions(ctx, conflict.Module)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Failed to get versions for %s: %v", conflict.Module, err))
			unresolvedConflicts = append(unresolvedConflicts, conflict)
			continue
		}

		metrics.VersionsEvaluated += len(versions)

		// Filter excluded versions
		filteredVersions := filterExcludedVersions(conflict.Module, versions, options.ExcludedVersions)

		// Sort versions in descending order (maximal first)
		sort.Sort(sort.Reverse(VersionSlice(filteredVersions)))

		// Prefer stable versions if requested
		if options.PreferStable {
			filteredVersions = sortWithStablePreference(filteredVersions)
		}

		// Find maximal compatible version
		var resolvedVersion string
		for _, version := range filteredVersions {
			if satisfiesAllConstraints(version, conflict, options) {
				resolvedVersion = version
				break
			}
		}

		if resolvedVersion != "" {
			resolvedVersions[conflict.Module] = resolvedVersion
			metrics.ConflictsResolved++
		} else {
			unresolvedConflicts = append(unresolvedConflicts, conflict)
		}
	}

	metrics.EndTime = time.Now()

	success := len(unresolvedConflicts) == 0
	explanation := fmt.Sprintf("Maximal upgrade strategy resolved %d/%d conflicts", metrics.ConflictsResolved, metrics.ConflictsProcessed)

	if options.Verbose {
		log.Infof("Maximal upgrade strategy: %s", explanation)
	}

	return &StrategyResult{
		Success:             success,
		ResolvedVersions:    resolvedVersions,
		UnresolvedConflicts: unresolvedConflicts,
		Explanation:         explanation,
		Warnings:            warnings,
		Metrics:             metrics,
	}, nil
}

// BacktrackStrategy implementation

func (s *BacktrackStrategy) Name() string {
	return "backtrack"
}

func (s *BacktrackStrategy) Description() string {
	return "Uses backtracking search to find optimal conflict resolution"
}

func (s *BacktrackStrategy) Resolve(ctx context.Context, conflicts []*DependencyConflict, options *StrategyOptions) (*StrategyResult, error) {
	metrics := StrategyMetrics{
		StartTime:          time.Now(),
		ConflictsProcessed: len(conflicts),
	}

	// Use the version solver for backtracking
	solver := NewVersionSolver(options.Provider)
	solver.SetVerbose(options.Verbose)

	// Collect all modules mentioned in conflicts (both targets and sources)
	allModules := make(map[string]bool)
	for _, conflict := range conflicts {
		allModules[conflict.Module] = true
		for _, requiring := range conflict.Requirements {
			for _, requirer := range requiring {
				allModules[requirer] = true
			}
		}
	}

	// Add variables for all modules
	for module := range allModules {
		// Get available versions
		versions, err := options.Provider.GetModuleVersions(ctx, module)
		if err != nil {
			if options.Verbose {
				log.Warnf("Skipping module %s: %v", module, err)
			}
			continue
		}

		if len(versions) == 0 {
			if options.Verbose {
				log.Warnf("Skipping module %s: no versions available", module)
			}
			continue
		}

		metrics.VersionsEvaluated += len(versions)

		// Filter excluded versions
		filteredVersions := filterExcludedVersions(module, versions, options.ExcludedVersions)

		// Check if module is pinned
		isPinned := false
		pinnedValue := ""
		if pinned, exists := options.PinnedVersions[module]; exists {
			isPinned = true
			pinnedValue = pinned
		}

		err = solver.AddVariable(module, filteredVersions, isPinned, pinnedValue)
		if err != nil {
			if options.Verbose {
				log.Warnf("Failed to add variable %s: %v", module, err)
			}
			continue
		}
	}

	// Add constraints from conflicts
	constraintCount := 0
	for _, conflict := range conflicts {
		for constraintStr, requiring := range conflict.Requirements {
			for _, requirer := range requiring {
				// Priority based on number of requirers (more requirers = higher priority)
				priority := len(requiring)
				err := solver.AddConstraint(requirer, conflict.Module, constraintStr, priority)
				if err != nil && options.Verbose {
					log.Warnf("Failed to add constraint %s -> %s %s: %v", requirer, conflict.Module, constraintStr, err)
				} else {
					constraintCount++
				}
			}
		}
	}

	if options.Verbose {
		log.Infof("Backtrack strategy: solving with %d variables and %d constraints", len(conflicts), constraintCount)
	}

	// Solve using CSP
	result, err := solver.Solve(ctx)
	if err != nil {
		return nil, err
	}

	// Convert solver metrics to strategy metrics
	metrics.EndTime = result.Metrics.EndTime
	metrics.BacktrackAttempts = result.Metrics.BacktrackCount
	metrics.VersionsEvaluated += result.Metrics.ConstraintChecks

	var unresolvedConflicts []*DependencyConflict
	if !result.Success {
		unresolvedConflicts = conflicts
	} else {
		metrics.ConflictsResolved = len(conflicts)
	}

	explanation := result.Explanation
	if options.Verbose && result.Success {
		explanation += fmt.Sprintf(" (backtrack depth: %d)", result.Metrics.BacktrackCount)
	}

	return &StrategyResult{
		Success:             result.Success,
		ResolvedVersions:    result.Assignments,
		UnresolvedConflicts: unresolvedConflicts,
		Explanation:         explanation,
		Metrics:             metrics,
	}, nil
}

// HybridStrategy implementation

func (s *HybridStrategy) Name() string {
	return "hybrid"
}

func (s *HybridStrategy) Description() string {
	return "Combines multiple strategies, falling back to more sophisticated approaches as needed"
}

func (s *HybridStrategy) Resolve(ctx context.Context, conflicts []*DependencyConflict, options *StrategyOptions) (*StrategyResult, error) {
	metrics := StrategyMetrics{
		StartTime:          time.Now(),
		ConflictsProcessed: len(conflicts),
	}

	var lastResult *StrategyResult
	var lastErr error

	for i, strategy := range s.strategies {
		if options.Verbose {
			log.Infof("Hybrid strategy: trying %s (attempt %d/%d)", strategy.Name(), i+1, len(s.strategies))
		}

		result, err := strategy.Resolve(ctx, conflicts, options)
		if err != nil {
			lastErr = err
			if options.Verbose {
				log.Warnf("Strategy %s failed: %v", strategy.Name(), err)
			}
			continue
		}

		lastResult = result

		// Update hybrid metrics
		metrics.ConflictsResolved += result.Metrics.ConflictsResolved
		metrics.VersionsEvaluated += result.Metrics.VersionsEvaluated
		metrics.BacktrackAttempts += result.Metrics.BacktrackAttempts

		if result.Success {
			metrics.EndTime = time.Now()
			result.Explanation = fmt.Sprintf("Hybrid strategy succeeded with %s: %s", strategy.Name(), result.Explanation)
			result.Metrics = metrics

			if options.Verbose {
				log.Infof("Hybrid strategy: %s succeeded", strategy.Name())
			}

			return result, nil
		}

		// Update conflicts for next strategy (only unresolved ones)
		if len(result.UnresolvedConflicts) < len(conflicts) {
			conflicts = result.UnresolvedConflicts
			if options.Verbose {
				log.Infof("Strategy %s partially resolved conflicts, %d remaining", strategy.Name(), len(conflicts))
			}
		}
	}

	metrics.EndTime = time.Now()

	// All strategies failed
	if lastResult != nil {
		lastResult.Explanation = fmt.Sprintf("Hybrid strategy: all %d strategies failed, last result: %s", len(s.strategies), lastResult.Explanation)
		lastResult.Metrics = metrics
		return lastResult, lastErr
	}

	return &StrategyResult{
		Success:             false,
		UnresolvedConflicts: conflicts,
		Explanation:         fmt.Sprintf("Hybrid strategy: all %d strategies failed", len(s.strategies)),
		Metrics:             metrics,
	}, lastErr
}

// Helper functions

// satisfiesAllConstraints checks if a version satisfies all constraints in a conflict
func satisfiesAllConstraints(version string, conflict *DependencyConflict, options *StrategyOptions) bool {
	for constraint := range conflict.Requirements {
		satisfied, err := CheckVersionConstraint(version, constraint)
		if err != nil || !satisfied {
			return false
		}
	}
	return true
}

// filterExcludedVersions removes excluded versions from the list
func filterExcludedVersions(module string, versions []string, excludedVersions map[string]map[string]bool) []string {
	excluded, hasExclusions := excludedVersions[module]
	if !hasExclusions {
		return versions
	}

	filtered := make([]string, 0, len(versions))
	for _, version := range versions {
		if !excluded[version] {
			filtered = append(filtered, version)
		}
	}
	return filtered
}

// sortWithStablePreference sorts versions preferring stable releases
func sortWithStablePreference(versions []string) []string {
	stable := make([]string, 0)
	unstable := make([]string, 0)

	for _, version := range versions {
		if isStableVersion(version) {
			stable = append(stable, version)
		} else {
			unstable = append(unstable, version)
		}
	}

	// Sort each group
	sort.Sort(sort.Reverse(VersionSlice(stable)))
	sort.Sort(sort.Reverse(VersionSlice(unstable)))

	// Combine with stable first
	result := make([]string, 0, len(versions))
	result = append(result, stable...)
	result = append(result, unstable...)

	return result
}

// isStableVersion determines if a version is considered stable
func isStableVersion(version string) bool {
	// Simple heuristic: versions with _XX, -TRIAL, -RC, etc. are considered unstable
	unstableMarkers := []string{"_", "-TRIAL", "-RC", "-ALPHA", "-BETA", "-DEV"}

	versionUpper := strings.ToUpper(version)
	for _, marker := range unstableMarkers {
		if strings.Contains(versionUpper, marker) {
			return false
		}
	}

	return true
}

// GetAvailableStrategies returns a list of all available resolution strategies
func GetAvailableStrategies() []ResolutionStrategy {
	return []ResolutionStrategy{
		NewFailFastStrategy(),
		NewMinimalUpgradeStrategy(),
		NewMaximalUpgradeStrategy(),
		NewBacktrackStrategy(50),
		NewHybridStrategy(
			NewMinimalUpgradeStrategy(),
			NewMaximalUpgradeStrategy(),
			NewBacktrackStrategy(50),
		),
	}
}

// GetStrategyByName returns a strategy by its name
func GetStrategyByName(name string) ResolutionStrategy {
	switch strings.ToLower(name) {
	case "fail-fast", "failfast":
		return NewFailFastStrategy()
	case "minimal-upgrade", "minimal", "min":
		return NewMinimalUpgradeStrategy()
	case "maximal-upgrade", "maximal", "max":
		return NewMaximalUpgradeStrategy()
	case "backtrack", "bt":
		return NewBacktrackStrategy(50)
	case "hybrid":
		return NewHybridStrategy(
			NewMinimalUpgradeStrategy(),
			NewMaximalUpgradeStrategy(),
			NewBacktrackStrategy(50),
		)
	default:
		return NewMaximalUpgradeStrategy() // Default strategy
	}
}
