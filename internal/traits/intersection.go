// ABOUTME: Trait intersection algorithm for union type capability determination
// ABOUTME: Implements lazy computation and caching for performance

package traits

import (
	"fmt"
	"sort"
	"strings"
)

// TraitIntersector handles trait intersection computation with lazy evaluation and caching
type TraitIntersector struct {
	// cache stores computed intersections for performance
	cache map[string]*TraitSet

	// resolver provides access to type traits
	resolver *OperationResolver

	// onCompute is a callback for testing (tracks computation calls)
	onCompute func()
}

// NewTraitIntersector creates a new trait intersector with empty cache
func NewTraitIntersector() *TraitIntersector {
	return &TraitIntersector{
		cache:    make(map[string]*TraitSet),
		resolver: NewOperationResolver(),
	}
}

// SetTraitsForType sets custom traits for a type (delegates to resolver)
func (ti *TraitIntersector) SetTraitsForType(typeName string, traits *TraitSet) {
	ti.resolver.SetTraitsForType(typeName, traits)
	// Invalidate cache since type definitions changed
	ti.ClearCache()
}

// ClearCache clears the intersection cache
func (ti *TraitIntersector) ClearCache() {
	ti.cache = make(map[string]*TraitSet)
}

// IntersectTypes computes the intersection of traits across multiple types
// Uses lazy computation and caching for performance
func (ti *TraitIntersector) IntersectTypes(typeNames []string) *TraitSet {
	// Handle edge cases
	if len(typeNames) == 0 {
		return NewTraitSet()
	}

	if len(typeNames) == 1 {
		return ti.resolver.GetTraitsForType(typeNames[0]).Clone()
	}

	// Create cache key by sorting type names for consistent caching
	sortedTypes := make([]string, len(typeNames))
	copy(sortedTypes, typeNames)
	sort.Strings(sortedTypes)

	// Remove duplicates
	uniqueTypes := removeDuplicates(sortedTypes)
	cacheKey := strings.Join(uniqueTypes, "|")

	// Check cache first (lazy computation)
	if cached, exists := ti.cache[cacheKey]; exists {
		return cached.Clone()
	}

	// Compute intersection
	if ti.onCompute != nil {
		ti.onCompute()
	}

	intersection := ti.computeIntersection(uniqueTypes)

	// Cache the result
	ti.cache[cacheKey] = intersection.Clone()

	return intersection
}

// computeIntersection performs the actual intersection computation
func (ti *TraitIntersector) computeIntersection(typeNames []string) *TraitSet {
	if len(typeNames) == 0 {
		return NewTraitSet()
	}

	// Start with the first type's traits
	result := ti.resolver.GetTraitsForType(typeNames[0]).Clone()

	// Intersect with each subsequent type
	for i := 1; i < len(typeNames); i++ {
		otherTraits := ti.resolver.GetTraitsForType(typeNames[i])
		result = ti.intersectTwoTraitSets(result, otherTraits)
	}

	return result
}

// intersectTwoTraitSets computes the intersection of two trait sets
func (ti *TraitIntersector) intersectTwoTraitSets(set1, set2 *TraitSet) *TraitSet {
	intersection := NewTraitSet()

	// Check each trait in set1
	for _, trait := range set1.GetAllTraits() {
		// If set2 also has this trait with the same result type, include it
		if set2.HasTrait(trait.Operation) {
			set2ResultType := set2.GetResultType(trait.Operation)
			if set2ResultType == trait.ResultType {
				intersection.AddTrait(trait)
			}
		}
	}

	return intersection
}

// GetCommonOperations returns operations supported by all given types
func (ti *TraitIntersector) GetCommonOperations(typeNames []string) []string {
	intersection := ti.IntersectTypes(typeNames)

	var operations []string
	for _, trait := range intersection.GetAllTraits() {
		operations = append(operations, trait.Operation)
	}

	sort.Strings(operations)
	return operations
}

// IsOperationSupportedByAll checks if an operation is supported by all given types
func (ti *TraitIntersector) IsOperationSupportedByAll(typeNames []string, operation string) bool {
	for _, typeName := range typeNames {
		if !ti.resolver.IsOperationSupported(typeName, operation) {
			return false
		}
	}
	return true
}

// GetIntersectionResultType returns the result type for an operation in the intersection
// Returns empty string if the operation is not in the intersection
func (ti *TraitIntersector) GetIntersectionResultType(typeNames []string, operation string) string {
	intersection := ti.IntersectTypes(typeNames)
	return intersection.GetResultType(operation)
}

// IntersectionInfo provides detailed information about an intersection
type IntersectionInfo struct {
	TypeNames    []string
	CommonTraits []Trait
	TotalTraits  int
	CacheHit     bool
}

// GetIntersectionInfo returns detailed information about a type intersection
func (ti *TraitIntersector) GetIntersectionInfo(typeNames []string) IntersectionInfo {
	// Check if this would be a cache hit
	sortedTypes := make([]string, len(typeNames))
	copy(sortedTypes, typeNames)
	sort.Strings(sortedTypes)
	uniqueTypes := removeDuplicates(sortedTypes)
	cacheKey := strings.Join(uniqueTypes, "|")

	_, cacheHit := ti.cache[cacheKey]

	intersection := ti.IntersectTypes(typeNames)
	commonTraits := intersection.GetAllTraits()

	return IntersectionInfo{
		TypeNames:    uniqueTypes,
		CommonTraits: commonTraits,
		TotalTraits:  len(commonTraits),
		CacheHit:     cacheHit,
	}
}

// CompareIntersections compares intersections of different type combinations
type IntersectionComparison struct {
	Types1     []string
	Types2     []string
	CommonOps1 []string
	CommonOps2 []string
	SharedOps  []string
	UniqueOps1 []string
	UniqueOps2 []string
}

// CompareIntersections compares the intersections of two type combinations
func (ti *TraitIntersector) CompareIntersections(types1, types2 []string) IntersectionComparison {
	ops1 := ti.GetCommonOperations(types1)
	ops2 := ti.GetCommonOperations(types2)

	// Find shared and unique operations
	var shared, unique1, unique2 []string

	// Operations in both
	for _, op1 := range ops1 {
		found := false
		for _, op2 := range ops2 {
			if op1 == op2 {
				shared = append(shared, op1)
				found = true
				break
			}
		}
		if !found {
			unique1 = append(unique1, op1)
		}
	}

	// Operations only in ops2
	for _, op2 := range ops2 {
		found := false
		for _, op1 := range ops1 {
			if op2 == op1 {
				found = true
				break
			}
		}
		if !found {
			unique2 = append(unique2, op2)
		}
	}

	return IntersectionComparison{
		Types1:     types1,
		Types2:     types2,
		CommonOps1: ops1,
		CommonOps2: ops2,
		SharedOps:  shared,
		UniqueOps1: unique1,
		UniqueOps2: unique2,
	}
}

// ValidateIntersection validates that an intersection is correctly computed
func (ti *TraitIntersector) ValidateIntersection(typeNames []string) error {
	intersection := ti.IntersectTypes(typeNames)

	// Check that every trait in the intersection is supported by all types
	for _, trait := range intersection.GetAllTraits() {
		for _, typeName := range typeNames {
			if !ti.resolver.IsOperationSupported(typeName, trait.Operation) {
				return fmt.Errorf("validation failed: operation '%s' not supported by type '%s'", trait.Operation, typeName)
			}

			// Check result type consistency
			expectedResult := ti.resolver.GetResultType(typeName, trait.Operation)
			if expectedResult != trait.ResultType {
				return fmt.Errorf("validation failed: operation '%s' on type '%s' should result in '%s', not '%s'",
					trait.Operation, typeName, expectedResult, trait.ResultType)
			}
		}
	}

	return nil
}

// GetCacheStats returns statistics about cache usage
type CacheStats struct {
	CacheSize int
	HitRate   float64
}

// GetCacheStats returns cache statistics (simplified version)
func (ti *TraitIntersector) GetCacheStats() CacheStats {
	return CacheStats{
		CacheSize: len(ti.cache),
		// HitRate would require tracking hits/misses over time
		HitRate: 0.0,
	}
}

// removeDuplicates removes duplicate strings from a sorted slice
func removeDuplicates(sorted []string) []string {
	if len(sorted) == 0 {
		return sorted
	}

	unique := make([]string, 0, len(sorted))
	unique = append(unique, sorted[0])

	for i := 1; i < len(sorted); i++ {
		if sorted[i] != sorted[i-1] {
			unique = append(unique, sorted[i])
		}
	}

	return unique
}
