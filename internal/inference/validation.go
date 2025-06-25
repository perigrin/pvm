// ABOUTME: Type annotation validation system
// ABOUTME: Validates type annotations against known patterns and best practices

package inference

import (
	"regexp"
	"strings"

	"tamarou.com/pvm/internal/types"
)

// TypeValidator validates type annotations for quality and correctness
type TypeValidator struct {
	// Known good patterns for type annotations
	knownPatterns []ValidationPattern

	// Configuration for validation rules
	validationRules ValidationRules
}

// ValidationPattern represents a known good pattern for type annotations
type ValidationPattern struct {
	// Pattern is a regex pattern for matching code constructs
	Pattern *regexp.Regexp

	// ExpectedType is the type that should be inferred for this pattern
	ExpectedType string

	// Confidence is the expected confidence level for this pattern
	Confidence float64

	// Description explains what this pattern represents
	Description string
}

// ValidationRules contains configuration for validation behavior
type ValidationRules struct {
	// RequireConsistency ensures consistent typing across similar constructs
	RequireConsistency bool

	// AllowVagueTypes permits generic types like Any in certain contexts
	AllowVagueTypes bool

	// MinConfidenceForAnnotation sets minimum confidence for including annotations
	MinConfidenceForAnnotation float64

	// MaxComplexityLevel limits type complexity for readability
	MaxComplexityLevel int
}

// ValidationResult represents the result of validating a type annotation
type ValidationResult struct {
	// IsValid indicates whether the annotation passes validation
	IsValid bool

	// Issues contains any problems found during validation
	Issues []ValidationIssue

	// Suggestions contains recommendations for improvement
	Suggestions []string

	// QualityScore is an overall quality assessment (0.0 to 1.0)
	QualityScore float64
}

// ValidationIssue represents a specific problem with a type annotation
type ValidationIssue struct {
	// Severity indicates how serious the issue is
	Severity IssueSeverity

	// Message describes the issue
	Message string

	// Category classifies the type of issue
	Category IssueCategory

	// Location indicates where the issue occurs
	Location *types.SourceLocation
}

// IssueSeverity represents how serious a validation issue is
type IssueSeverity string

const (
	// SeverityError indicates a serious problem that should block annotation
	SeverityError IssueSeverity = "error"

	// SeverityWarning indicates a potential problem worth noting
	SeverityWarning IssueSeverity = "warning"

	// SeverityInfo indicates informational feedback
	SeverityInfo IssueSeverity = "info"
)

// IssueCategory classifies different types of validation issues
type IssueCategory string

const (
	// CategoryAccuracy relates to correctness of the inferred type
	CategoryAccuracy IssueCategory = "accuracy"

	// CategoryClarity relates to readability and understandability
	CategoryClarity IssueCategory = "clarity"

	// CategoryConsistency relates to consistency with surrounding code
	CategoryConsistency IssueCategory = "consistency"

	// CategoryComplexity relates to type complexity and maintainability
	CategoryComplexity IssueCategory = "complexity"
)

// NewTypeValidator creates a new type validator with default patterns and rules
func NewTypeValidator() *TypeValidator {
	return &TypeValidator{
		knownPatterns:   createDefaultPatterns(),
		validationRules: createDefaultValidationRules(),
	}
}

// createDefaultPatterns creates standard validation patterns
func createDefaultPatterns() []ValidationPattern {
	patterns := []ValidationPattern{
		{
			Pattern:      regexp.MustCompile(`my\s+\$\w+\s*=\s*\d+`),
			ExpectedType: "Int",
			Confidence:   0.95,
			Description:  "Integer literal assignment",
		},
		{
			Pattern:      regexp.MustCompile(`my\s+\$\w+\s*=\s*"[^"]*"`),
			ExpectedType: "Str",
			Confidence:   0.95,
			Description:  "String literal assignment",
		},
		{
			Pattern:      regexp.MustCompile(`my\s+\$\w+\s*=\s*\[\]`),
			ExpectedType: "ArrayRef",
			Confidence:   0.90,
			Description:  "Empty array reference assignment",
		},
		{
			Pattern:      regexp.MustCompile(`my\s+\$\w+\s*=\s*\{\}`),
			ExpectedType: "HashRef",
			Confidence:   0.90,
			Description:  "Empty hash reference assignment",
		},
	}

	return patterns
}

// createDefaultValidationRules creates standard validation rules
func createDefaultValidationRules() ValidationRules {
	return ValidationRules{
		RequireConsistency:         true,
		AllowVagueTypes:            false,
		MinConfidenceForAnnotation: 0.6,
		MaxComplexityLevel:         3,
	}
}

// ValidateTypeAnnotation validates a type annotation against known patterns and rules
func (tv *TypeValidator) ValidateTypeAnnotation(typeInfo *types.TypeInfo, code string) ValidationResult {
	result := ValidationResult{
		IsValid:      true,
		Issues:       make([]ValidationIssue, 0),
		Suggestions:  make([]string, 0),
		QualityScore: 1.0,
	}

	// Check against known patterns
	tv.validateAgainstPatterns(typeInfo, code, &result)

	// Check validation rules
	tv.validateAgainstRules(typeInfo, &result)

	// Calculate overall quality score
	result.QualityScore = tv.calculateQualityScore(&result)

	// Determine if annotation is valid
	result.IsValid = tv.isValidAnnotation(&result)

	return result
}

// validateAgainstPatterns checks type annotation against known good patterns
func (tv *TypeValidator) validateAgainstPatterns(typeInfo *types.TypeInfo, code string, result *ValidationResult) {
	for _, pattern := range tv.knownPatterns {
		if pattern.Pattern.MatchString(code) {
			// This code matches a known pattern
			expectedType := pattern.ExpectedType
			actualType := typeInfo.Type.String()

			if actualType != expectedType {
				// Type mismatch with known pattern
				result.Issues = append(result.Issues, ValidationIssue{
					Severity: SeverityWarning,
					Message:  "Type annotation doesn't match known pattern",
					Category: CategoryAccuracy,
				})

				result.Suggestions = append(result.Suggestions,
					"Consider using type "+expectedType+" for this construct")
			}

			// Check confidence against expected
			if typeInfo.Confidence < pattern.Confidence-0.1 {
				result.Issues = append(result.Issues, ValidationIssue{
					Severity: SeverityInfo,
					Message:  "Confidence lower than expected for this pattern",
					Category: CategoryAccuracy,
				})
			}

			break // Only match against first matching pattern
		}
	}
}

// validateAgainstRules checks type annotation against validation rules
func (tv *TypeValidator) validateAgainstRules(typeInfo *types.TypeInfo, result *ValidationResult) {
	// Check minimum confidence
	if typeInfo.Confidence < tv.validationRules.MinConfidenceForAnnotation {
		result.Issues = append(result.Issues, ValidationIssue{
			Severity: SeverityWarning,
			Message:  "Type annotation confidence below minimum threshold",
			Category: CategoryAccuracy,
		})
	}

	// Check for vague types
	if !tv.validationRules.AllowVagueTypes && tv.isVagueType(typeInfo.Type) {
		result.Issues = append(result.Issues, ValidationIssue{
			Severity: SeverityWarning,
			Message:  "Type annotation is too vague to be helpful",
			Category: CategoryClarity,
		})

		result.Suggestions = append(result.Suggestions,
			"Consider using a more specific type or omitting this annotation")
	}

	// Check type complexity
	complexity := tv.calculateTypeComplexity(typeInfo.Type)
	if complexity > tv.validationRules.MaxComplexityLevel {
		result.Issues = append(result.Issues, ValidationIssue{
			Severity: SeverityInfo,
			Message:  "Type annotation is quite complex",
			Category: CategoryComplexity,
		})

		result.Suggestions = append(result.Suggestions,
			"Consider using type aliases for complex types")
	}
}

// isVagueType determines if a type is too vague to be useful
func (tv *TypeValidator) isVagueType(typ types.Type) bool {
	vagueTypes := []string{"Any", "Item", "Scalar"}
	typeStr := typ.String()

	for _, vague := range vagueTypes {
		if typeStr == vague {
			return true
		}
	}

	return false
}

// calculateTypeComplexity determines the complexity level of a type
func (tv *TypeValidator) calculateTypeComplexity(typ types.Type) int {
	typeStr := typ.String()

	// Count complexity indicators
	complexity := 0

	// Union types add complexity
	if strings.Contains(typeStr, "|") {
		complexity += strings.Count(typeStr, "|")
	}

	// Parameterized types add complexity
	if strings.Contains(typeStr, "[") {
		complexity += strings.Count(typeStr, "[")
	}

	// Intersection types add complexity
	if strings.Contains(typeStr, "&") {
		complexity += strings.Count(typeStr, "&")
	}

	// Base complexity for any type
	complexity += 1

	return complexity
}

// calculateQualityScore computes an overall quality score for the validation result
func (tv *TypeValidator) calculateQualityScore(result *ValidationResult) float64 {
	if len(result.Issues) == 0 {
		return 1.0
	}

	// Start with perfect score and deduct for issues
	score := 1.0

	for _, issue := range result.Issues {
		switch issue.Severity {
		case SeverityError:
			score -= 0.3
		case SeverityWarning:
			score -= 0.2
		case SeverityInfo:
			score -= 0.1
		}
	}

	// Ensure score doesn't go below 0
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// isValidAnnotation determines if an annotation should be considered valid
func (tv *TypeValidator) isValidAnnotation(result *ValidationResult) bool {
	// Annotations with errors are not valid
	for _, issue := range result.Issues {
		if issue.Severity == SeverityError {
			return false
		}
	}

	// Quality score must be above minimum threshold
	return result.QualityScore >= 0.5
}

// ValidateConsistency checks for consistency across multiple type annotations
func (tv *TypeValidator) ValidateConsistency(annotations []*types.TypeInfo) []ValidationIssue {
	issues := make([]ValidationIssue, 0)

	if !tv.validationRules.RequireConsistency {
		return issues
	}

	// Group annotations by similar patterns
	groups := tv.groupSimilarAnnotations(annotations)

	// Check for consistency within each group
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}

		// Check if types are consistent within the group
		firstType := group[0].Type.String()
		for i := 1; i < len(group); i++ {
			if group[i].Type.String() != firstType {
				issues = append(issues, ValidationIssue{
					Severity: SeverityWarning,
					Message:  "Inconsistent type annotations for similar constructs",
					Category: CategoryConsistency,
				})
				break
			}
		}
	}

	return issues
}

// groupSimilarAnnotations groups annotations that should have consistent types
func (tv *TypeValidator) groupSimilarAnnotations(annotations []*types.TypeInfo) [][]*types.TypeInfo {
	// Simple grouping by source type for now
	groups := make(map[types.TypeSource][]*types.TypeInfo)

	for _, annotation := range annotations {
		source := annotation.Source
		if _, exists := groups[source]; !exists {
			groups[source] = make([]*types.TypeInfo, 0)
		}
		groups[source] = append(groups[source], annotation)
	}

	// Convert map to slice
	result := make([][]*types.TypeInfo, 0, len(groups))
	for _, group := range groups {
		result = append(result, group)
	}

	return result
}
