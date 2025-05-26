// ABOUTME: Advanced conflict resolution system for dependency resolver
// ABOUTME: Implements sophisticated algorithms for handling version conflicts

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

// Additional conflict resolution strategies extending the existing ones
const (
	// StrategyMinimalUpgrade chooses minimal version upgrades to resolve conflicts
	StrategyMinimalUpgrade ConflictResolutionStrategy = iota + 10
	// StrategyMaximalUpgrade chooses maximal compatible versions
	StrategyMaximalUpgrade
	// StrategyBacktrack uses backtracking to find optimal resolution
	StrategyBacktrack
	// StrategyInteractive prompts user for resolution choices
	StrategyInteractive
)

// ConflictResolver handles complex dependency conflicts using various strategies
type ConflictResolver struct {
	provider     cpan.Provider
	strategy     ConflictResolutionStrategy
	maxRetries   int
	maxBacktrack int
	verbose      bool

	// Pinned versions that cannot be changed
	pinnedVersions map[string]string

	// Excluded versions that should never be selected
	excludedVersions map[string]map[string]bool

	// Preferred versions to choose when multiple options exist
	preferredVersions map[string]string

	// Resolution history for learning and reporting
	resolutionHistory []ResolutionAttempt
}

// ResolutionAttempt records a single resolution attempt
type ResolutionAttempt struct {
	Timestamp      time.Time
	Module         string
	Strategy       ConflictResolutionStrategy
	Attempted      string
	Success        bool
	Reason         string
	Constraints    map[string]bool
	BacktrackDepth int
}

// ConflictExplanation provides detailed information about why a conflict occurred
type ConflictExplanation struct {
	Module              string
	ConflictingVersions map[string][]string // version -> requiring modules
	AvailableVersions   []string
	RecommendedVersion  string
	RecommendedReason   string
	AlternativeOptions  []string
	ImpactAnalysis      string
}

// ResolutionResult contains the outcome of conflict resolution
type ResolutionResult struct {
	Success            bool
	ResolvedVersions   map[string]string
	RemainingConflicts []*DependencyConflict
	Explanations       []*ConflictExplanation
	ResolutionPath     []string
	Metrics            ConflictResolutionMetrics
}

// ConflictResolutionMetrics tracks performance and effectiveness metrics for conflict resolution
type ConflictResolutionMetrics struct {
	StartTime              time.Time
	EndTime                time.Time
	ConflictsFound         int
	ConflictsResolved      int
	BacktrackAttempts      int
	StrategySwitches       int
	CacheHits              int
	TotalVersionsEvaluated int
}

// NewConflictResolver creates a new conflict resolver with specified strategy
func NewConflictResolver(provider cpan.Provider, strategy ConflictResolutionStrategy) *ConflictResolver {
	return &ConflictResolver{
		provider:          provider,
		strategy:          strategy,
		maxRetries:        3,
		maxBacktrack:      50,
		verbose:           false,
		pinnedVersions:    make(map[string]string),
		excludedVersions:  make(map[string]map[string]bool),
		preferredVersions: make(map[string]string),
		resolutionHistory: make([]ResolutionAttempt, 0),
	}
}

// SetPinnedVersions sets versions that cannot be changed during resolution
func (cr *ConflictResolver) SetPinnedVersions(pinned map[string]string) {
	cr.pinnedVersions = pinned
}

// SetExcludedVersions sets versions that should never be selected
func (cr *ConflictResolver) SetExcludedVersions(excluded map[string]map[string]bool) {
	cr.excludedVersions = excluded
}

// SetPreferredVersions sets preferred versions to choose when multiple options exist
func (cr *ConflictResolver) SetPreferredVersions(preferred map[string]string) {
	cr.preferredVersions = preferred
}

// SetVerbose enables or disables verbose logging
func (cr *ConflictResolver) SetVerbose(verbose bool) {
	cr.verbose = verbose
}

// ResolveConflicts attempts to resolve dependency conflicts using the configured strategy
func (cr *ConflictResolver) ResolveConflicts(ctx context.Context, result *DependencyResolutionResult) (*ResolutionResult, error) {
	if len(result.Conflicts) == 0 {
		return &ResolutionResult{
			Success:            true,
			ResolvedVersions:   make(map[string]string),
			RemainingConflicts: []*DependencyConflict{},
			Explanations:       []*ConflictExplanation{},
			ResolutionPath:     []string{"No conflicts to resolve"},
		}, nil
	}

	metrics := ConflictResolutionMetrics{
		StartTime:      time.Now(),
		ConflictsFound: len(result.Conflicts),
	}

	if cr.verbose {
		log.Infof("Starting conflict resolution for %d conflicts using strategy %v", len(result.Conflicts), cr.strategy)
	}

	// Generate explanations for all conflicts
	explanations := make([]*ConflictExplanation, 0, len(result.Conflicts))
	for _, conflict := range result.Conflicts {
		explanation, err := cr.explainConflict(ctx, conflict)
		if err != nil {
			log.Warnf("Failed to generate explanation for conflict in %s: %v", conflict.Module, err)
		} else {
			explanations = append(explanations, explanation)
		}
	}

	var resResult *ResolutionResult
	var err error

	switch cr.strategy {
	case StrategyFailFast:
		resResult, err = cr.resolveFailFast(result, explanations)
	case StrategyMinimalUpgrade:
		resResult, err = cr.resolveMinimalUpgrade(ctx, result, explanations, &metrics)
	case StrategyMaximalUpgrade:
		resResult, err = cr.resolveMaximalUpgrade(ctx, result, explanations, &metrics)
	case StrategyBacktrack:
		resResult, err = cr.resolveWithBacktracking(ctx, result, explanations, &metrics)
	case StrategyInteractive:
		resResult, err = cr.resolveInteractive(ctx, result, explanations, &metrics)
	default:
		return nil, errors.NewSystemError(
			ErrResolveFailure,
			fmt.Sprintf("Unknown conflict resolution strategy: %v", cr.strategy),
			nil)
	}

	if resResult != nil {
		metrics.EndTime = time.Now()
		resResult.Metrics = metrics

		if cr.verbose {
			log.Infof("Resolution completed in %v: %d/%d conflicts resolved",
				metrics.EndTime.Sub(metrics.StartTime),
				metrics.ConflictsResolved,
				metrics.ConflictsFound)
		}
	}

	return resResult, err
}

// resolveFailFast immediately fails when conflicts are detected
func (cr *ConflictResolver) resolveFailFast(result *DependencyResolutionResult, explanations []*ConflictExplanation) (*ResolutionResult, error) {
	conflictDetails := make([]string, 0, len(result.Conflicts))
	for _, conflict := range result.Conflicts {
		details := fmt.Sprintf("Module %s has conflicting requirements:", conflict.Module)
		for version, requiring := range conflict.Requirements {
			details += fmt.Sprintf("\n  - Version %s required by: %s", version, strings.Join(requiring, ", "))
		}
		conflictDetails = append(conflictDetails, details)
	}

	return &ResolutionResult{
			Success:            false,
			ResolvedVersions:   make(map[string]string),
			RemainingConflicts: result.Conflicts,
			Explanations:       explanations,
			ResolutionPath:     []string{"Failed fast due to conflicts"},
		}, errors.NewSystemError(
			ErrVersionConflict,
			fmt.Sprintf("Conflicts detected in %d modules:\n%s", len(result.Conflicts), strings.Join(conflictDetails, "\n\n")),
			nil)
}

// resolveMinimalUpgrade attempts to resolve conflicts with minimal version changes
func (cr *ConflictResolver) resolveMinimalUpgrade(ctx context.Context, result *DependencyResolutionResult, explanations []*ConflictExplanation, metrics *ConflictResolutionMetrics) (*ResolutionResult, error) {
	resolvedVersions := make(map[string]string)
	resolutionPath := []string{"Starting minimal upgrade resolution"}

	for _, conflict := range result.Conflicts {
		// Check if version is pinned
		if pinnedVersion, isPinned := cr.pinnedVersions[conflict.Module]; isPinned {
			// Verify pinned version satisfies all constraints
			if cr.pinnedVersionSatisfiesConstraints(pinnedVersion, conflict) {
				resolvedVersions[conflict.Module] = pinnedVersion
				resolutionPath = append(resolutionPath, fmt.Sprintf("Used pinned version %s for %s", pinnedVersion, conflict.Module))
				metrics.ConflictsResolved++
				continue
			} else {
				return &ResolutionResult{
						Success:            false,
						ResolvedVersions:   resolvedVersions,
						RemainingConflicts: []*DependencyConflict{conflict},
						Explanations:       explanations,
						ResolutionPath:     resolutionPath,
					}, errors.NewSystemError(
						ErrVersionConflict,
						fmt.Sprintf("Pinned version %s for %s does not satisfy all constraints", pinnedVersion, conflict.Module),
						nil)
			}
		}

		// Get available versions
		versions, err := cr.provider.GetModuleVersions(ctx, conflict.Module)
		if err != nil {
			log.Warnf("Failed to get versions for %s: %v", conflict.Module, err)
			continue
		}

		metrics.TotalVersionsEvaluated += len(versions)

		// Sort versions in ascending order for minimal upgrade
		sort.Sort(VersionSlice(versions))

		// Find the minimal version that satisfies all constraints
		resolvedVersion := cr.findMinimalCompatibleVersion(versions, conflict)
		if resolvedVersion != "" {
			resolvedVersions[conflict.Module] = resolvedVersion
			resolutionPath = append(resolutionPath, fmt.Sprintf("Resolved %s to minimal compatible version %s", conflict.Module, resolvedVersion))
			metrics.ConflictsResolved++
		}
	}

	// Determine remaining conflicts
	remainingConflicts := make([]*DependencyConflict, 0)
	for _, conflict := range result.Conflicts {
		if _, resolved := resolvedVersions[conflict.Module]; !resolved {
			remainingConflicts = append(remainingConflicts, conflict)
		}
	}

	return &ResolutionResult{
		Success:            len(remainingConflicts) == 0,
		ResolvedVersions:   resolvedVersions,
		RemainingConflicts: remainingConflicts,
		Explanations:       explanations,
		ResolutionPath:     resolutionPath,
	}, nil
}

// resolveMaximalUpgrade attempts to resolve conflicts with maximal compatible versions
func (cr *ConflictResolver) resolveMaximalUpgrade(ctx context.Context, result *DependencyResolutionResult, explanations []*ConflictExplanation, metrics *ConflictResolutionMetrics) (*ResolutionResult, error) {
	resolvedVersions := make(map[string]string)
	resolutionPath := []string{"Starting maximal upgrade resolution"}

	for _, conflict := range result.Conflicts {
		// Check if version is pinned
		if pinnedVersion, isPinned := cr.pinnedVersions[conflict.Module]; isPinned {
			if cr.pinnedVersionSatisfiesConstraints(pinnedVersion, conflict) {
				resolvedVersions[conflict.Module] = pinnedVersion
				resolutionPath = append(resolutionPath, fmt.Sprintf("Used pinned version %s for %s", pinnedVersion, conflict.Module))
				metrics.ConflictsResolved++
				continue
			}
		}

		// Get available versions
		versions, err := cr.provider.GetModuleVersions(ctx, conflict.Module)
		if err != nil {
			log.Warnf("Failed to get versions for %s: %v", conflict.Module, err)
			continue
		}

		metrics.TotalVersionsEvaluated += len(versions)

		// Sort versions in descending order for maximal upgrade
		sort.Sort(sort.Reverse(VersionSlice(versions)))

		// Find the maximal version that satisfies all constraints
		resolvedVersion := cr.findMaximalCompatibleVersion(versions, conflict)
		if resolvedVersion != "" {
			resolvedVersions[conflict.Module] = resolvedVersion
			resolutionPath = append(resolutionPath, fmt.Sprintf("Resolved %s to maximal compatible version %s", conflict.Module, resolvedVersion))
			metrics.ConflictsResolved++
		}
	}

	// Determine remaining conflicts
	remainingConflicts := make([]*DependencyConflict, 0)
	for _, conflict := range result.Conflicts {
		if _, resolved := resolvedVersions[conflict.Module]; !resolved {
			remainingConflicts = append(remainingConflicts, conflict)
		}
	}

	return &ResolutionResult{
		Success:            len(remainingConflicts) == 0,
		ResolvedVersions:   resolvedVersions,
		RemainingConflicts: remainingConflicts,
		Explanations:       explanations,
		ResolutionPath:     resolutionPath,
	}, nil
}

// resolveWithBacktracking uses backtracking algorithm to find optimal resolution
func (cr *ConflictResolver) resolveWithBacktracking(ctx context.Context, result *DependencyResolutionResult, explanations []*ConflictExplanation, metrics *ConflictResolutionMetrics) (*ResolutionResult, error) {
	resolutionPath := []string{"Starting backtracking resolution"}

	// Create a search state for backtracking
	state := &BacktrackState{
		conflicts:         result.Conflicts,
		resolvedVersions:  make(map[string]string),
		candidateVersions: make(map[string][]string),
		conflictIndex:     0,
		depth:             0,
	}

	// Populate candidate versions for each conflict
	for _, conflict := range result.Conflicts {
		if pinnedVersion, isPinned := cr.pinnedVersions[conflict.Module]; isPinned {
			state.candidateVersions[conflict.Module] = []string{pinnedVersion}
		} else {
			versions, err := cr.provider.GetModuleVersions(ctx, conflict.Module)
			if err != nil {
				log.Warnf("Failed to get versions for %s: %v", conflict.Module, err)
				continue
			}

			// Filter out excluded versions
			filteredVersions := cr.filterExcludedVersions(conflict.Module, versions)

			// Sort based on preference (preferred versions first, then by version)
			sortedVersions := cr.sortVersionsByPreference(conflict.Module, filteredVersions)

			state.candidateVersions[conflict.Module] = sortedVersions
			metrics.TotalVersionsEvaluated += len(versions)
		}
	}

	// Perform backtracking search
	success := cr.backtrackSearch(ctx, state, metrics)

	resolutionPath = append(resolutionPath, fmt.Sprintf("Backtracking completed after %d attempts", metrics.BacktrackAttempts))

	// Determine remaining conflicts
	remainingConflicts := make([]*DependencyConflict, 0)
	if success {
		// Verify all conflicts are resolved
		for _, conflict := range result.Conflicts {
			if _, resolved := state.resolvedVersions[conflict.Module]; !resolved {
				remainingConflicts = append(remainingConflicts, conflict)
			}
		}
		metrics.ConflictsResolved = len(result.Conflicts) - len(remainingConflicts)
	} else {
		remainingConflicts = result.Conflicts
	}

	return &ResolutionResult{
		Success:            success && len(remainingConflicts) == 0,
		ResolvedVersions:   state.resolvedVersions,
		RemainingConflicts: remainingConflicts,
		Explanations:       explanations,
		ResolutionPath:     resolutionPath,
	}, nil
}

// BacktrackState represents the state during backtracking search
type BacktrackState struct {
	conflicts         []*DependencyConflict
	resolvedVersions  map[string]string
	candidateVersions map[string][]string
	conflictIndex     int
	depth             int
}

// backtrackSearch performs recursive backtracking to find a valid resolution
func (cr *ConflictResolver) backtrackSearch(ctx context.Context, state *BacktrackState, metrics *ConflictResolutionMetrics) bool {
	metrics.BacktrackAttempts++

	// Check if we've exceeded maximum backtrack depth
	if state.depth >= cr.maxBacktrack {
		return false
	}

	// Base case: all conflicts resolved
	if state.conflictIndex >= len(state.conflicts) {
		// Verify the solution is valid
		return cr.validateSolution(state.resolvedVersions, state.conflicts)
	}

	conflict := state.conflicts[state.conflictIndex]
	candidates := state.candidateVersions[conflict.Module]

	// Try each candidate version
	for _, version := range candidates {
		// Skip excluded versions
		if cr.isVersionExcluded(conflict.Module, version) {
			continue
		}

		// Check if this version satisfies the current conflict
		if !cr.versionSatisfiesConflict(version, conflict) {
			continue
		}

		// Make move
		state.resolvedVersions[conflict.Module] = version
		state.conflictIndex++
		state.depth++

		// Check for consistency with other resolved versions
		if cr.isConsistentWithOtherResolutions(ctx, conflict.Module, version, state.resolvedVersions) {
			// Recursively solve remaining conflicts
			if cr.backtrackSearch(ctx, state, metrics) {
				return true
			}
		}

		// Backtrack
		delete(state.resolvedVersions, conflict.Module)
		state.conflictIndex--
		state.depth--
	}

	return false
}

// Helper methods for conflict resolution

// explainConflict generates a detailed explanation for a conflict
func (cr *ConflictResolver) explainConflict(ctx context.Context, conflict *DependencyConflict) (*ConflictExplanation, error) {
	// Get available versions
	versions, err := cr.provider.GetModuleVersions(ctx, conflict.Module)
	if err != nil {
		return nil, err
	}

	// Find recommended version
	recommendedVersion := cr.findRecommendedVersion(versions, conflict)
	recommendedReason := "No compatible version found"
	if recommendedVersion != "" {
		recommendedReason = "Latest version that satisfies all constraints"
	}

	// Generate alternative options
	alternatives := cr.findAlternativeVersions(versions, conflict, 3)

	// Create impact analysis
	impact := cr.analyzeConflictImpact(conflict)

	return &ConflictExplanation{
		Module:              conflict.Module,
		ConflictingVersions: conflict.Requirements,
		AvailableVersions:   versions,
		RecommendedVersion:  recommendedVersion,
		RecommendedReason:   recommendedReason,
		AlternativeOptions:  alternatives,
		ImpactAnalysis:      impact,
	}, nil
}

// findMinimalCompatibleVersion finds the minimal version that satisfies all constraints
func (cr *ConflictResolver) findMinimalCompatibleVersion(versions []string, conflict *DependencyConflict) string {
	for _, version := range versions {
		if cr.isVersionExcluded(conflict.Module, version) {
			continue
		}
		if cr.versionSatisfiesConflict(version, conflict) {
			return version
		}
	}
	return ""
}

// findMaximalCompatibleVersion finds the maximal version that satisfies all constraints
func (cr *ConflictResolver) findMaximalCompatibleVersion(versions []string, conflict *DependencyConflict) string {
	for _, version := range versions {
		if cr.isVersionExcluded(conflict.Module, version) {
			continue
		}
		if cr.versionSatisfiesConflict(version, conflict) {
			return version
		}
	}
	return ""
}

// pinnedVersionSatisfiesConstraints checks if a pinned version satisfies all constraints
func (cr *ConflictResolver) pinnedVersionSatisfiesConstraints(version string, conflict *DependencyConflict) bool {
	return cr.versionSatisfiesConflict(version, conflict)
}

// versionSatisfiesConflict checks if a version satisfies all requirements in a conflict
func (cr *ConflictResolver) versionSatisfiesConflict(version string, conflict *DependencyConflict) bool {
	for constraint := range conflict.Requirements {
		satisfied, err := CheckVersionConstraint(version, constraint)
		if err != nil || !satisfied {
			return false
		}
	}
	return true
}

// isVersionExcluded checks if a version is in the excluded list
func (cr *ConflictResolver) isVersionExcluded(module, version string) bool {
	if excluded, ok := cr.excludedVersions[module]; ok {
		return excluded[version]
	}
	return false
}

// filterExcludedVersions removes excluded versions from the list
func (cr *ConflictResolver) filterExcludedVersions(module string, versions []string) []string {
	if excluded, ok := cr.excludedVersions[module]; !ok {
		return versions
	} else {
		filtered := make([]string, 0, len(versions))
		for _, version := range versions {
			if !excluded[version] {
				filtered = append(filtered, version)
			}
		}
		return filtered
	}
}

// sortVersionsByPreference sorts versions with preferred versions first
func (cr *ConflictResolver) sortVersionsByPreference(module string, versions []string) []string {
	preferred, hasPreferred := cr.preferredVersions[module]

	if !hasPreferred {
		// No preference, just sort by version
		sort.Sort(sort.Reverse(VersionSlice(versions)))
		return versions
	}

	// Put preferred version first, then sort rest by version
	result := make([]string, 0, len(versions))
	rest := make([]string, 0, len(versions))

	for _, version := range versions {
		if version == preferred {
			result = append(result, version)
		} else {
			rest = append(rest, version)
		}
	}

	sort.Sort(sort.Reverse(VersionSlice(rest)))
	result = append(result, rest...)

	return result
}

// validateSolution checks if a solution is valid for all conflicts
func (cr *ConflictResolver) validateSolution(resolvedVersions map[string]string, conflicts []*DependencyConflict) bool {
	for _, conflict := range conflicts {
		version, resolved := resolvedVersions[conflict.Module]
		if !resolved {
			return false
		}
		if !cr.versionSatisfiesConflict(version, conflict) {
			return false
		}
	}
	return true
}

// isConsistentWithOtherResolutions checks for consistency with other resolved versions
func (cr *ConflictResolver) isConsistentWithOtherResolutions(ctx context.Context, module, version string, resolvedVersions map[string]string) bool {
	// This would check for dependency compatibility with other resolved versions
	// For now, we'll assume consistency - a more sophisticated implementation would
	// verify that the chosen version's dependencies are compatible with other choices
	return true
}

// findRecommendedVersion finds the best version to recommend for a conflict
func (cr *ConflictResolver) findRecommendedVersion(versions []string, conflict *DependencyConflict) string {
	// Sort versions in descending order (latest first)
	sort.Sort(sort.Reverse(VersionSlice(versions)))

	for _, version := range versions {
		if cr.versionSatisfiesConflict(version, conflict) {
			return version
		}
	}
	return ""
}

// findAlternativeVersions finds alternative versions that might work
func (cr *ConflictResolver) findAlternativeVersions(versions []string, conflict *DependencyConflict, limit int) []string {
	alternatives := make([]string, 0, limit)

	// Sort versions in descending order
	sort.Sort(sort.Reverse(VersionSlice(versions)))

	for _, version := range versions {
		if len(alternatives) >= limit {
			break
		}
		if cr.versionSatisfiesConflict(version, conflict) {
			alternatives = append(alternatives, version)
		}
	}

	return alternatives
}

// analyzeConflictImpact analyzes the impact of a conflict
func (cr *ConflictResolver) analyzeConflictImpact(conflict *DependencyConflict) string {
	impactLevel := "Low"
	requirerCount := 0

	for _, requirers := range conflict.Requirements {
		requirerCount += len(requirers)
	}

	if requirerCount > 10 {
		impactLevel = "High"
	} else if requirerCount > 5 {
		impactLevel = "Medium"
	}

	return fmt.Sprintf("%s impact - affects %d dependent modules", impactLevel, requirerCount)
}

// Placeholder for interactive resolution - would require user input handling
func (cr *ConflictResolver) resolveInteractive(ctx context.Context, result *DependencyResolutionResult, explanations []*ConflictExplanation, metrics *ConflictResolutionMetrics) (*ResolutionResult, error) {
	// For now, fall back to maximal upgrade strategy
	// In a real implementation, this would prompt the user for choices
	return cr.resolveMaximalUpgrade(ctx, result, explanations, metrics)
}
