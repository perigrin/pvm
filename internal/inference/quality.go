// ABOUTME: Quality control and confidence scoring system for type inference
// ABOUTME: Provides sophisticated confidence algorithms and quality metrics for type annotations

package inference

import (
	"math"

	"tamarou.com/pvm/internal/types"
)

// QualityController provides sophisticated quality control and confidence scoring
type QualityController struct {
	options QualityOptions
}

// QualityOptions contains configuration for quality control
type QualityOptions struct {
	// MinConfidenceThreshold sets minimum confidence for including annotations
	MinConfidenceThreshold float64

	// FilterVagueTypes removes overly generic types like Any
	FilterVagueTypes bool

	// ConfidenceBoostFactor multiplies confidence when evidence is strong
	ConfidenceBoostFactor float64

	// ConflictPenaltyFactor reduces confidence when conflicts are detected
	ConflictPenaltyFactor float64

	// ComplexityPenaltyFactor reduces confidence in complex code contexts
	ComplexityPenaltyFactor float64

	// AgreementBoostFactor increases confidence when multiple sources agree
	AgreementBoostFactor float64
}

// QualityMetrics represents quality assessment of a type annotation
type QualityMetrics struct {
	// OverallQuality is a categorical assessment: "high", "medium", "low"
	OverallQuality string

	// Helpfulness measures how useful this annotation would be to developers
	Helpfulness float64

	// Accuracy measures confidence in the correctness of the type
	Accuracy float64

	// Clarity measures how clear and unambiguous the type is
	Clarity float64

	// ShouldInclude indicates whether this annotation should be included
	ShouldInclude bool
}

// NewQualityController creates a new quality controller with the given options
func NewQualityController(options QualityOptions) *QualityController {
	return &QualityController{
		options: options,
	}
}

// DefaultQualityOptions returns sensible default quality control options
func DefaultQualityOptions() QualityOptions {
	return QualityOptions{
		MinConfidenceThreshold:  0.6,
		FilterVagueTypes:        true,
		ConfidenceBoostFactor:   1.2,
		ConflictPenaltyFactor:   0.5,
		ComplexityPenaltyFactor: 0.8,
		AgreementBoostFactor:    1.3,
	}
}

// CalculateConfidence computes sophisticated confidence score based on source and context
func (qc *QualityController) CalculateConfidence(source types.TypeSource, context map[string]interface{}) float64 {
	// Start with base confidence from source type
	baseConfidences := map[types.TypeSource]float64{
		types.SourceLiteral:   0.95,
		types.SourceVariable:  0.85,
		types.SourceParameter: 0.70,
		types.SourceReturn:    0.75,
		types.SourceContext:   0.60,
		types.SourceExternal:  0.90,
	}

	baseConfidence, exists := baseConfidences[source]
	if !exists {
		baseConfidence = 0.50
	}

	// Apply context-based adjustments
	adjustedConfidence := qc.applyContextAdjustments(baseConfidence, context)

	// Ensure confidence stays in valid range [0.0, 1.0]
	return math.Max(0.0, math.Min(1.0, adjustedConfidence))
}

// applyContextAdjustments modifies confidence based on context information
func (qc *QualityController) applyContextAdjustments(baseConfidence float64, context map[string]interface{}) float64 {
	confidence := baseConfidence

	// Handle conflicting evidence
	if conflictCount, ok := context["conflicting_sources"].(int); ok && conflictCount > 0 {
		penaltyFactor := math.Pow(qc.options.ConflictPenaltyFactor, float64(conflictCount))
		confidence *= penaltyFactor
	}

	// Handle agreeing evidence
	if agreeCount, ok := context["agreeing_sources"].(int); ok && agreeCount > 1 {
		boostFactor := math.Pow(qc.options.AgreementBoostFactor, float64(agreeCount-1))
		confidence *= boostFactor
	}

	// Handle evidence strength
	if strength, ok := context["evidence_strength"].(float64); ok {
		// Strong evidence boosts confidence, weak evidence reduces it
		strengthMultiplier := 0.5 + (strength * 0.5) // Range: 0.5 to 1.0
		confidence *= strengthMultiplier
	}

	// Handle code complexity
	if nestingLevel, ok := context["nesting_level"].(int); ok && nestingLevel > 3 {
		complexityPenalty := math.Pow(qc.options.ComplexityPenaltyFactor, float64(nestingLevel-3))
		confidence *= complexityPenalty
	}

	if complexity, ok := context["code_complexity"].(float64); ok && complexity > 0.5 {
		complexityPenalty := 1.0 - ((complexity - 0.5) * 0.4) // Reduce up to 20%
		confidence *= complexityPenalty
	}

	// Handle ambiguous context
	if ambiguous, ok := context["ambiguous_context"].(bool); ok && ambiguous {
		confidence *= 0.7 // 30% penalty for ambiguous context
	}

	return confidence
}

// EvaluateTypeQuality assesses the overall quality of a type annotation
func (qc *QualityController) EvaluateTypeQuality(typeInfo *types.TypeInfo) QualityMetrics {
	accuracy := typeInfo.Confidence
	helpfulness := qc.calculateHelpfulness(typeInfo)
	clarity := qc.calculateClarity(typeInfo)

	// Overall quality is the minimum of the three metrics
	overallScore := math.Min(accuracy, math.Min(helpfulness, clarity))

	var overallQuality string
	switch {
	case overallScore >= 0.8:
		overallQuality = "high"
	case overallScore >= 0.5:
		overallQuality = "medium"
	default:
		overallQuality = "low"
	}

	shouldInclude := qc.ShouldIncludeAnnotation(typeInfo)

	return QualityMetrics{
		OverallQuality: overallQuality,
		Helpfulness:    helpfulness,
		Accuracy:       accuracy,
		Clarity:        clarity,
		ShouldInclude:  shouldInclude,
	}
}

// calculateHelpfulness determines how helpful a type annotation would be
func (qc *QualityController) calculateHelpfulness(typeInfo *types.TypeInfo) float64 {
	// More specific types are more helpful
	switch typeInfo.Type.String() {
	case "Any":
		return 0.1 // Any type provides little value
	case "Int", "Str", "Bool":
		return 0.9 // Basic specific types are very helpful
	case "ArrayRef", "HashRef":
		return 0.8 // Container types are helpful
	default:
		// Complex types are moderately helpful
		return 0.7
	}
}

// calculateClarity determines how clear and unambiguous a type is
func (qc *QualityController) calculateClarity(typeInfo *types.TypeInfo) float64 {
	typeStr := typeInfo.Type.String()

	// Simple types are clearer
	switch typeStr {
	case "Any":
		return 0.1 // Any is not clear
	case "Int", "Str", "Bool":
		return 0.9 // Basic types are very clear
	case "ArrayRef", "HashRef":
		return 0.8 // Simple container types are clear
	default:
		// Complex types may be less clear
		complexity := float64(len(typeStr)) / 50.0 // Rough complexity measure
		return math.Max(0.3, 1.0-complexity)
	}
}

// ShouldIncludeAnnotation determines whether to include a type annotation
func (qc *QualityController) ShouldIncludeAnnotation(typeInfo *types.TypeInfo) bool {
	// Check confidence threshold
	if typeInfo.Confidence < qc.options.MinConfidenceThreshold {
		return false
	}

	// Filter vague types if enabled
	if qc.options.FilterVagueTypes && qc.isVagueType(typeInfo.Type) {
		return false
	}

	// Check overall quality without calling back to ShouldIncludeAnnotation
	// Calculate quality directly to avoid infinite recursion
	accuracy := typeInfo.Confidence
	helpfulness := qc.calculateHelpfulness(typeInfo)
	clarity := qc.calculateClarity(typeInfo)
	overallScore := math.Min(accuracy, math.Min(helpfulness, clarity))

	return overallScore >= 0.5
}

// isVagueType determines if a type is too vague to be useful
func (qc *QualityController) isVagueType(typ types.Type) bool {
	switch typ.String() {
	case "Any", "Scalar", "Item":
		return true
	default:
		return false
	}
}

// GenerateUncertaintyComment creates a comment for uncertain type annotations
func (qc *QualityController) GenerateUncertaintyComment(typeInfo *types.TypeInfo) string {
	confidence := typeInfo.Confidence

	switch {
	case confidence >= 0.8:
		return "" // No comment needed for high confidence
	case confidence >= 0.6:
		return "# Type inferred with medium confidence"
	case confidence >= 0.4:
		return "# Type inferred with low confidence - may need verification"
	default:
		return "# Type uncertain - manual verification recommended"
	}
}
