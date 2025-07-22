// ABOUTME: Advanced intersection type validation and contradiction detection
// ABOUTME: Implements semantic validation for intersection types with conflict resolution

package typedef

import (
	"fmt"
	"sort"
	"strings"
)

// IntersectionValidator provides advanced validation for intersection types
type IntersectionValidator struct {
	hierarchy *TypeHierarchy
}

// NewIntersectionValidator creates a new intersection validator
func NewIntersectionValidator(hierarchy *TypeHierarchy) *IntersectionValidator {
	return &IntersectionValidator{
		hierarchy: hierarchy,
	}
}

// ContradictionType represents different types of intersection contradictions
type ContradictionType int

const (
	DisjointTypes      ContradictionType = iota // Int & Str - completely separate types
	SubtypeRedundancy                           // Base & Derived where Derived <: Base
	ValueContradiction                          // value constraints that can't be satisfied
	StructuralMismatch                          // incompatible structural requirements
)

// IntersectionContradiction represents a detected contradiction in an intersection
type IntersectionContradiction struct {
	Type        ContradictionType
	Members     []string // The conflicting members
	Explanation string
	Severity    ContradictionSeverity
	Suggestion  string // How to fix the contradiction
}

// ContradictionSeverity indicates how severe a contradiction is
type ContradictionSeverity int

const (
	Warning ContradictionSeverity = iota // Possibly unintended but not necessarily wrong
	Error                                // Definitely incorrect
	Info                                 // Informational - could be simplified
)

// ValidateIntersection performs comprehensive validation of an intersection type
func (iv *IntersectionValidator) ValidateIntersection(intersectionType string) ([]IntersectionContradiction, error) {
	// Parse the intersection type
	members, err := iv.parseIntersectionMembers(intersectionType)
	if err != nil {
		return nil, err
	}

	var contradictions []IntersectionContradiction

	// Check for disjoint types
	disjointContradictions := iv.checkDisjointTypes(members)
	contradictions = append(contradictions, disjointContradictions...)

	// Check for subtype redundancies
	redundancyContradictions := iv.checkSubtypeRedundancies(members)
	contradictions = append(contradictions, redundancyContradictions...)

	// Check for structural mismatches
	structuralContradictions := iv.checkStructuralMismatches(members)
	contradictions = append(contradictions, structuralContradictions...)

	// Check for value contradictions
	valueContradictions := iv.checkValueContradictions(members)
	contradictions = append(contradictions, valueContradictions...)

	return contradictions, nil
}

// parseIntersectionMembers extracts individual type members from intersection syntax
func (iv *IntersectionValidator) parseIntersectionMembers(intersectionType string) ([]string, error) {
	// Handle both A&B&C and Intersection[A,B,C] formats
	if strings.HasPrefix(intersectionType, "Intersection[") {
		// Parse parameterized format: Intersection[A,B,C]
		if !strings.HasSuffix(intersectionType, "]") {
			return nil, fmt.Errorf("invalid intersection type format: %s", intersectionType)
		}
		paramStr := intersectionType[13 : len(intersectionType)-1]
		return iv.parseTypeList(paramStr), nil
	}

	// Parse infix format: A&B&C
	return iv.parseInfixIntersection(intersectionType), nil
}

// parseTypeList parses a comma-separated list of types
func (iv *IntersectionValidator) parseTypeList(params string) []string {
	if params == "" {
		return []string{}
	}

	var members []string
	current := ""
	depth := 0

	for _, r := range params {
		switch r {
		case '[', '<', '(':
			depth++
			current += string(r)
		case ']', '>', ')':
			depth--
			current += string(r)
		case ',':
			if depth == 0 {
				members = append(members, strings.TrimSpace(current))
				current = ""
			} else {
				current += string(r)
			}
		default:
			current += string(r)
		}
	}

	if current != "" {
		members = append(members, strings.TrimSpace(current))
	}

	return members
}

// parseInfixIntersection parses A&B&C format
func (iv *IntersectionValidator) parseInfixIntersection(intersectionType string) []string {
	var members []string
	current := ""
	depth := 0

	for _, r := range intersectionType {
		switch r {
		case '[', '<', '(':
			depth++
			current += string(r)
		case ']', '>', ')':
			depth--
			current += string(r)
		case '&':
			if depth == 0 {
				if current != "" {
					members = append(members, strings.TrimSpace(current))
				}
				current = ""
			} else {
				current += string(r)
			}
		default:
			current += string(r)
		}
	}

	if current != "" {
		members = append(members, strings.TrimSpace(current))
	}

	return members
}

// checkDisjointTypes identifies types that cannot possibly intersect
func (iv *IntersectionValidator) checkDisjointTypes(members []string) []IntersectionContradiction {
	var contradictions []IntersectionContradiction

	// Known disjoint type sets
	disjointSets := [][]string{
		{"Int", "Str", "Bool", "Undef"},           // Basic scalar types
		{"ArrayRef", "HashRef", "CodeRef", "Ref"}, // Reference types
		{"Scalar", "Array", "Hash", "Code"},       // Perl built-in types
	}

	// Check each disjoint set for base type conflicts
	for _, disjointSet := range disjointSets {
		foundMembers := []string{}
		baseTypes := make(map[string]bool)

		for _, member := range members {
			baseType := iv.extractBaseType(member)
			for _, disjointType := range disjointSet {
				if baseType == disjointType {
					foundMembers = append(foundMembers, member)
					baseTypes[baseType] = true
					break
				}
			}
		}

		// Only report disjoint if we have DIFFERENT base types from the same disjoint set
		// Don't report disjoint for different parameterizations of the same base type
		if len(foundMembers) > 1 && len(baseTypes) > 1 {
			contradiction := IntersectionContradiction{
				Type:        DisjointTypes,
				Members:     foundMembers,
				Explanation: fmt.Sprintf("Types %s are mutually exclusive and cannot be intersected", strings.Join(foundMembers, " and ")),
				Severity:    Error,
				Suggestion:  fmt.Sprintf("Consider using a union type (%s) instead of intersection", strings.Join(foundMembers, "|")),
			}
			contradictions = append(contradictions, contradiction)
		}
	}

	return contradictions
}

// checkSubtypeRedundancies identifies redundant types in intersections
func (iv *IntersectionValidator) checkSubtypeRedundancies(members []string) []IntersectionContradiction {
	var contradictions []IntersectionContradiction

	// First, check if any members are disjoint or have structural mismatches
	// If so, skip redundancy checks because those take precedence over subtype relationships
	disjointContradictions := iv.checkDisjointTypes(members)
	if len(disjointContradictions) > 0 {
		// If we have disjoint types, don't check for redundancy
		// because the disjoint relationship is more important
		return contradictions
	}

	structuralContradictions := iv.checkStructuralMismatches(members)
	if len(structuralContradictions) > 0 {
		// If we have structural mismatches, don't check for redundancy
		// because the structural issue is more important
		return contradictions
	}

	// Check each pair of members for subtype relationships
	for i, member1 := range members {
		for j, member2 := range members {
			if i >= j {
				continue // Avoid duplicates and self-comparison
			}

			// Check if member1 is a subtype of member2
			if iv.isSubtype(member1, member2) {
				contradiction := IntersectionContradiction{
					Type:        SubtypeRedundancy,
					Members:     []string{member1, member2},
					Explanation: fmt.Sprintf("Type %s is a subtype of %s, making the intersection redundant", member1, member2),
					Severity:    Info,
					Suggestion:  fmt.Sprintf("Use %s alone (the more specific type)", member1),
				}
				contradictions = append(contradictions, contradiction)
			} else if iv.isSubtype(member2, member1) {
				contradiction := IntersectionContradiction{
					Type:        SubtypeRedundancy,
					Members:     []string{member2, member1},
					Explanation: fmt.Sprintf("Type %s is a subtype of %s, making the intersection redundant", member2, member1),
					Severity:    Info,
					Suggestion:  fmt.Sprintf("Use %s alone (the more specific type)", member2),
				}
				contradictions = append(contradictions, contradiction)
			}
		}
	}

	return contradictions
}

// checkStructuralMismatches identifies structural incompatibilities
func (iv *IntersectionValidator) checkStructuralMismatches(members []string) []IntersectionContradiction {
	var contradictions []IntersectionContradiction

	// This would check for structural type conflicts
	// For now, we'll implement basic parameterized type checking
	parameterizedTypes := make(map[string][]string)

	for _, member := range members {
		if strings.Contains(member, "[") {
			baseType, _ := extractTypeAndParams(member)
			if baseType != "" {
				parameterizedTypes[baseType] = append(parameterizedTypes[baseType], member)
			}
		}
	}

	// Check for conflicting parameterizations of the same type
	for baseType, typeInstances := range parameterizedTypes {
		if len(typeInstances) > 1 {
			contradiction := IntersectionContradiction{
				Type:        StructuralMismatch,
				Members:     typeInstances,
				Explanation: fmt.Sprintf("Multiple conflicting parameterizations of %s type: %s", baseType, strings.Join(typeInstances, ", ")),
				Severity:    Error,
				Suggestion:  fmt.Sprintf("Use a single parameterization or consider if intersection is the right approach"),
			}
			contradictions = append(contradictions, contradiction)
		}
	}

	return contradictions
}

// checkValueContradictions identifies value-level contradictions
func (iv *IntersectionValidator) checkValueContradictions(members []string) []IntersectionContradiction {
	var contradictions []IntersectionContradiction

	// Check for literal value types that can't coexist
	literalValues := []string{}
	for _, member := range members {
		if iv.isLiteralType(member) {
			literalValues = append(literalValues, member)
		}
	}

	if len(literalValues) > 1 {
		contradiction := IntersectionContradiction{
			Type:        ValueContradiction,
			Members:     literalValues,
			Explanation: fmt.Sprintf("Literal types %s cannot be satisfied simultaneously", strings.Join(literalValues, " and ")),
			Severity:    Error,
			Suggestion:  "Use a union type if you want to accept any of these literal values",
		}
		contradictions = append(contradictions, contradiction)
	}

	return contradictions
}

// extractBaseType extracts the base type name from a potentially parameterized type
func (iv *IntersectionValidator) extractBaseType(typeStr string) string {
	if idx := strings.Index(typeStr, "["); idx != -1 {
		return typeStr[:idx]
	}
	return typeStr
}

// isSubtype checks if type1 is a subtype of type2
func (iv *IntersectionValidator) isSubtype(type1, type2 string) bool {
	if iv.hierarchy == nil {
		return false
	}
	return iv.hierarchy.IsSubtypeOf(type1, type2)
}

// isLiteralType checks if a type represents a literal value
func (iv *IntersectionValidator) isLiteralType(typeStr string) bool {
	// This would be expanded to check for actual literal types
	// For now, just check some common patterns
	return strings.HasPrefix(typeStr, "'") || // String literals
		strings.HasPrefix(typeStr, "\"") || // String literals
		(len(typeStr) > 0 && typeStr[0] >= '0' && typeStr[0] <= '9') // Numeric literals
}

// SimplifyIntersection removes redundant members and resolves simple contradictions
func (iv *IntersectionValidator) SimplifyIntersection(intersectionType string) (string, []IntersectionContradiction, error) {
	members, err := iv.parseIntersectionMembers(intersectionType)
	if err != nil {
		return intersectionType, nil, err
	}

	contradictions, err := iv.ValidateIntersection(intersectionType)
	if err != nil {
		return intersectionType, nil, err
	}

	// Remove redundant members based on detected contradictions
	simplifiedMembers := make([]string, 0, len(members))
	memberSet := make(map[string]bool)

	for _, member := range members {
		// Skip if we've already added this member
		if memberSet[member] {
			continue
		}

		// Check if this member is made redundant by a subtype relationship
		isRedundant := false
		for _, contradiction := range contradictions {
			if contradiction.Type == SubtypeRedundancy {
				// If this member is the more general type in a subtype relationship,
				// mark it as redundant (we keep the more specific type)
				if len(contradiction.Members) == 2 {
					if contradiction.Members[1] == member && iv.contains(members, contradiction.Members[0]) {
						isRedundant = true
						break
					}
				}
			}
		}

		if !isRedundant {
			simplifiedMembers = append(simplifiedMembers, member)
			memberSet[member] = true
		}
	}

	// Rebuild the intersection string
	switch len(simplifiedMembers) {
	case 0:
		return "Never", contradictions, nil // Empty intersection
	case 1:
		return simplifiedMembers[0], contradictions, nil // Single type
	default:
		// Sort for consistent output
		sort.Strings(simplifiedMembers)
		return strings.Join(simplifiedMembers, "&"), contradictions, nil
	}
}

// contains checks if a slice contains a string
func (iv *IntersectionValidator) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetContradictionsByType filters contradictions by type
func (iv *IntersectionValidator) GetContradictionsByType(contradictions []IntersectionContradiction, contradictionType ContradictionType) []IntersectionContradiction {
	var filtered []IntersectionContradiction
	for _, contradiction := range contradictions {
		if contradiction.Type == contradictionType {
			filtered = append(filtered, contradiction)
		}
	}
	return filtered
}

// GetContradictionsBySeverity filters contradictions by severity
func (iv *IntersectionValidator) GetContradictionsBySeverity(contradictions []IntersectionContradiction, severity ContradictionSeverity) []IntersectionContradiction {
	var filtered []IntersectionContradiction
	for _, contradiction := range contradictions {
		if contradiction.Severity == severity {
			filtered = append(filtered, contradiction)
		}
	}
	return filtered
}

// FormatContradictions formats contradictions for display to users
func (iv *IntersectionValidator) FormatContradictions(contradictions []IntersectionContradiction) []string {
	var formatted []string
	for _, contradiction := range contradictions {
		severityStr := ""
		switch contradiction.Severity {
		case Error:
			severityStr = "ERROR"
		case Warning:
			severityStr = "WARNING"
		case Info:
			severityStr = "INFO"
		}

		message := fmt.Sprintf("[%s] %s", severityStr, contradiction.Explanation)
		if contradiction.Suggestion != "" {
			message += fmt.Sprintf(" Suggestion: %s", contradiction.Suggestion)
		}
		formatted = append(formatted, message)
	}
	return formatted
}
