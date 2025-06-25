// ABOUTME: Type conflict detection and resolution system
// ABOUTME: Detects competing type inferences and provides resolution strategies

package inference

import (
	"fmt"
	"sort"

	"tamarou.com/pvm/internal/types"
)

// ConflictDetector identifies and resolves type conflicts
type ConflictDetector struct {
	// Configuration for conflict detection sensitivity
	toleranceLevel float64
}

// TypeConflict represents a detected conflict between type inferences
type TypeConflict struct {
	// NodeID identifies where the conflict occurs
	NodeID string

	// ConflictingTypes are the types that conflict
	ConflictingTypes []*types.TypeInfo

	// Severity indicates how serious the conflict is
	Severity ConflictSeverity

	// Resolution suggests how to resolve the conflict
	Resolution ConflictResolution
}

// ConflictSeverity represents the seriousness of a type conflict
type ConflictSeverity string

const (
	// SeverityLow indicates minor conflicts that can be easily resolved
	SeverityLow ConflictSeverity = "low"

	// SeverityMedium indicates conflicts requiring careful consideration
	SeverityMedium ConflictSeverity = "medium"

	// SeverityHigh indicates serious conflicts that may indicate code issues
	SeverityHigh ConflictSeverity = "high"
)

// ConflictResolution suggests how to resolve a type conflict
type ConflictResolution struct {
	// Strategy describes the recommended resolution approach
	Strategy ResolutionStrategy

	// ResolvedType is the suggested type after resolution
	ResolvedType *types.TypeInfo

	// Confidence in the resolution
	Confidence float64

	// Explanation describes why this resolution was chosen
	Explanation string
}

// ResolutionStrategy represents different ways to resolve conflicts
type ResolutionStrategy string

const (
	// StrategyHighestConfidence selects the type with highest confidence
	StrategyHighestConfidence ResolutionStrategy = "highest_confidence"

	// StrategyMostSpecific selects the most specific type
	StrategyMostSpecific ResolutionStrategy = "most_specific"

	// StrategyUnion creates a union type of all conflicting types
	StrategyUnion ResolutionStrategy = "union"

	// StrategyMostReliableSource selects based on source reliability
	StrategyMostReliableSource ResolutionStrategy = "most_reliable_source"

	// StrategyNoResolution indicates the conflict cannot be automatically resolved
	StrategyNoResolution ResolutionStrategy = "no_resolution"
)

// NewConflictDetector creates a new conflict detector
func NewConflictDetector() *ConflictDetector {
	return &ConflictDetector{
		toleranceLevel: 0.1, // 10% tolerance for confidence differences
	}
}

// DetectConflicts analyzes type information to identify conflicts
func (cd *ConflictDetector) DetectConflicts(typeInfos []*types.TypeInfo) *TypeConflict {
	if len(typeInfos) < 2 {
		return nil // No conflict possible with fewer than 2 types
	}

	// Group types by compatibility
	compatibilityGroups := cd.groupByCompatibility(typeInfos)

	// If all types are in one group, no conflict
	if len(compatibilityGroups) <= 1 {
		return nil
	}

	// Create conflict with all incompatible types
	conflictingTypes := make([]*types.TypeInfo, 0)
	for _, group := range compatibilityGroups {
		// Add the highest confidence type from each incompatible group
		sort.Slice(group, func(i, j int) bool {
			return group[i].Confidence > group[j].Confidence
		})
		conflictingTypes = append(conflictingTypes, group[0])
	}

	severity := cd.calculateSeverity(conflictingTypes)
	resolution := cd.resolveConflict(conflictingTypes, severity)

	return &TypeConflict{
		ConflictingTypes: conflictingTypes,
		Severity:         severity,
		Resolution:       resolution,
	}
}

// groupByCompatibility groups types that are compatible with each other
func (cd *ConflictDetector) groupByCompatibility(typeInfos []*types.TypeInfo) [][]*types.TypeInfo {
	groups := make([][]*types.TypeInfo, 0)

	for _, typeInfo := range typeInfos {
		placed := false

		// Try to place in existing group
		for i, group := range groups {
			if cd.isCompatibleWithGroup(typeInfo, group) {
				groups[i] = append(groups[i], typeInfo)
				placed = true
				break
			}
		}

		// Create new group if not placed
		if !placed {
			groups = append(groups, []*types.TypeInfo{typeInfo})
		}
	}

	return groups
}

// isCompatibleWithGroup checks if a type is compatible with all types in a group
func (cd *ConflictDetector) isCompatibleWithGroup(typeInfo *types.TypeInfo, group []*types.TypeInfo) bool {
	for _, groupType := range group {
		if !cd.areTypesCompatible(typeInfo.Type, groupType.Type) {
			return false
		}
	}
	return true
}

// areTypesCompatible determines if two types are compatible
func (cd *ConflictDetector) areTypesCompatible(type1, type2 types.Type) bool {
	// Same types are always compatible
	if type1.String() == type2.String() {
		return true
	}

	// Any type is compatible with everything
	if type1.String() == "Any" || type2.String() == "Any" {
		return true
	}

	// Check for subtype relationships
	if cd.isSubtype(type1, type2) || cd.isSubtype(type2, type1) {
		return true
	}

	// Otherwise, types are incompatible
	return false
}

// isSubtype checks if type1 is a subtype of type2
func (cd *ConflictDetector) isSubtype(type1, type2 types.Type) bool {
	// Simple subtype relationships
	subtypeMap := map[string][]string{
		"Scalar": {"Int", "Str", "Bool"},
		"Item":   {"Int", "Str", "Bool", "ArrayRef", "HashRef"},
		"Ref":    {"ArrayRef", "HashRef"},
	}

	if supertypes, exists := subtypeMap[type2.String()]; exists {
		for _, subtype := range supertypes {
			if type1.String() == subtype {
				return true
			}
		}
	}

	return false
}

// calculateSeverity determines how serious a conflict is
func (cd *ConflictDetector) calculateSeverity(conflictingTypes []*types.TypeInfo) ConflictSeverity {
	if len(conflictingTypes) < 2 {
		return SeverityLow
	}

	// Calculate confidence spread
	minConfidence := conflictingTypes[0].Confidence
	maxConfidence := conflictingTypes[0].Confidence

	for _, typeInfo := range conflictingTypes {
		if typeInfo.Confidence < minConfidence {
			minConfidence = typeInfo.Confidence
		}
		if typeInfo.Confidence > maxConfidence {
			maxConfidence = typeInfo.Confidence
		}
	}

	confidenceSpread := maxConfidence - minConfidence

	// Determine severity based on confidence spread and type specificity
	switch {
	case confidenceSpread > 0.5:
		return SeverityHigh // Large confidence differences suggest serious conflict
	case confidenceSpread > 0.2:
		return SeverityMedium // Moderate confidence differences
	default:
		return SeverityLow // Small confidence differences
	}
}

// resolveConflict attempts to resolve a type conflict
func (cd *ConflictDetector) resolveConflict(conflictingTypes []*types.TypeInfo, severity ConflictSeverity) ConflictResolution {
	if len(conflictingTypes) < 2 {
		return ConflictResolution{
			Strategy:   StrategyNoResolution,
			Confidence: 0.0,
		}
	}

	// Sort by confidence (highest first)
	sortedTypes := make([]*types.TypeInfo, len(conflictingTypes))
	copy(sortedTypes, conflictingTypes)
	sort.Slice(sortedTypes, func(i, j int) bool {
		return sortedTypes[i].Confidence > sortedTypes[j].Confidence
	})

	highestConfidenceType := sortedTypes[0]

	// Choose resolution strategy based on severity and confidence differences
	switch severity {
	case SeverityLow:
		// For low severity, use highest confidence
		return ConflictResolution{
			Strategy:     StrategyHighestConfidence,
			ResolvedType: highestConfidenceType,
			Confidence:   highestConfidenceType.Confidence * 0.9, // Slight penalty for conflict
			Explanation:  "Selected highest confidence type due to low conflict severity",
		}

	case SeverityMedium:
		// For medium severity, try to find most specific type
		mostSpecific := cd.findMostSpecificType(sortedTypes)
		if mostSpecific != nil {
			return ConflictResolution{
				Strategy:     StrategyMostSpecific,
				ResolvedType: mostSpecific,
				Confidence:   mostSpecific.Confidence * 0.8, // Moderate penalty for conflict
				Explanation:  "Selected most specific type to resolve moderate conflict",
			}
		}
		fallthrough

	case SeverityHigh:
		// For high severity, create union type or give up
		if cd.canCreateUnion(sortedTypes) {
			unionType := cd.createUnionType(sortedTypes)
			return ConflictResolution{
				Strategy:     StrategyUnion,
				ResolvedType: unionType,
				Confidence:   0.5, // Union types have medium confidence
				Explanation:  "Created union type to capture all possibilities",
			}
		}

		// Cannot resolve automatically
		return ConflictResolution{
			Strategy:    StrategyNoResolution,
			Confidence:  0.0,
			Explanation: "Conflict too severe for automatic resolution",
		}
	}

	return ConflictResolution{
		Strategy:   StrategyNoResolution,
		Confidence: 0.0,
	}
}

// findMostSpecificType finds the most specific type among the conflicting types
func (cd *ConflictDetector) findMostSpecificType(typeInfos []*types.TypeInfo) *types.TypeInfo {
	// Specificity ranking (higher is more specific)
	specificityRanking := map[string]int{
		"Any":      1,
		"Item":     2,
		"Scalar":   3,
		"Ref":      4,
		"ArrayRef": 5,
		"HashRef":  5,
		"Int":      6,
		"Str":      6,
		"Bool":     6,
	}

	var mostSpecific *types.TypeInfo
	highestSpecificity := 0

	for _, typeInfo := range typeInfos {
		specificity, exists := specificityRanking[typeInfo.Type.String()]
		if !exists {
			specificity = 7 // Unknown types are considered highly specific
		}

		if specificity > highestSpecificity {
			highestSpecificity = specificity
			mostSpecific = typeInfo
		}
	}

	return mostSpecific
}

// canCreateUnion determines if a union type can be created from the conflicting types
func (cd *ConflictDetector) canCreateUnion(typeInfos []*types.TypeInfo) bool {
	// For now, we can create unions for basic types
	for _, typeInfo := range typeInfos {
		switch typeInfo.Type.String() {
		case "Int", "Str", "Bool", "ArrayRef", "HashRef":
			// These are good candidates for union types
		default:
			return false // Complex types may not work well in unions
		}
	}
	return len(typeInfos) <= 3 // Limit union complexity
}

// createUnionType creates a union type from conflicting types
func (cd *ConflictDetector) createUnionType(typeInfos []*types.TypeInfo) *types.TypeInfo {
	// Create a union type string
	typeStrings := make([]string, len(typeInfos))
	for i, typeInfo := range typeInfos {
		typeStrings[i] = typeInfo.Type.String()
	}

	// Sort for consistent union type representation
	sort.Strings(typeStrings)

	// Create union type string
	unionTypeStr := typeStrings[0]
	for i := 1; i < len(typeStrings); i++ {
		unionTypeStr += "|" + typeStrings[i]
	}

	// Calculate average confidence
	totalConfidence := 0.0
	for _, typeInfo := range typeInfos {
		totalConfidence += typeInfo.Confidence
	}
	avgConfidence := totalConfidence / float64(len(typeInfos))

	// Create a union type from the conflicting types
	unionTypes := make([]types.Type, len(typeInfos))
	for i, typeInfo := range typeInfos {
		unionTypes[i] = typeInfo.Type
	}
	unionType := types.NewUnionType(unionTypes...)

	return types.NewTypeInfo(unionType, avgConfidence, types.SourceContext)
}

// String returns a string representation of the conflict
func (tc *TypeConflict) String() string {
	return fmt.Sprintf("Type conflict (%s severity): %d conflicting types",
		tc.Severity, len(tc.ConflictingTypes))
}
