// ABOUTME: Tests for quality control and confidence scoring system
// ABOUTME: Validates sophisticated confidence scoring algorithms and quality metrics

package inference

import (
	"testing"

	"tamarou.com/pvm/internal/types"
)

func TestQualityController_CalculateConfidence(t *testing.T) {
	tests := []struct {
		name          string
		source        types.TypeSource
		context       map[string]interface{}
		expectedRange [2]float64 // [min, max]
		description   string
	}{
		{
			name:          "literal inference high confidence",
			source:        types.SourceLiteral,
			context:       map[string]interface{}{},
			expectedRange: [2]float64{0.9, 1.0},
			description:   "Literal types should have very high confidence",
		},
		{
			name:   "conflicting evidence reduces confidence",
			source: types.SourceVariable,
			context: map[string]interface{}{
				"conflicting_sources": 2,
				"evidence_strength":   0.3,
			},
			expectedRange: [2]float64{0.1, 0.4},
			description:   "Conflicting type evidence should reduce confidence",
		},
		{
			name:   "multiple agreeing sources increase confidence",
			source: types.SourceParameter,
			context: map[string]interface{}{
				"agreeing_sources":  3,
				"evidence_strength": 0.9,
			},
			expectedRange: [2]float64{0.8, 1.0},
			description:   "Multiple sources agreeing should increase confidence",
		},
		{
			name:   "complex code reduces confidence",
			source: types.SourceContext,
			context: map[string]interface{}{
				"nesting_level":     5,
				"code_complexity":   0.8,
				"ambiguous_context": true,
			},
			expectedRange: [2]float64{0.2, 0.5},
			description:   "Complex code context should reduce confidence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qc := NewQualityController(DefaultQualityOptions())
			confidence := qc.CalculateConfidence(tt.source, tt.context)

			if confidence < tt.expectedRange[0] || confidence > tt.expectedRange[1] {
				t.Errorf("CalculateConfidence() = %v, expected range [%v, %v]",
					confidence, tt.expectedRange[0], tt.expectedRange[1])
			}
		})
	}
}

func TestQualityController_EvaluateTypeQuality(t *testing.T) {
	qc := NewQualityController(DefaultQualityOptions())

	tests := []struct {
		name     string
		typeInfo *types.TypeInfo
		expected QualityMetrics
	}{
		{
			name: "high quality type annotation",
			typeInfo: types.NewTypeInfo(
				types.NewIntType(),
				0.95,
				types.SourceLiteral,
			),
			expected: QualityMetrics{
				OverallQuality: "high",
				Helpfulness:    0.9,
				Accuracy:       0.95,
				Clarity:        0.9,
			},
		},
		{
			name: "low quality type annotation",
			typeInfo: types.NewTypeInfo(
				types.NewAnyType(),
				0.3,
				types.SourceContext,
			),
			expected: QualityMetrics{
				OverallQuality: "low",
				Helpfulness:    0.2,
				Accuracy:       0.3,
				Clarity:        0.1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := qc.EvaluateTypeQuality(tt.typeInfo)

			if metrics.OverallQuality != tt.expected.OverallQuality {
				t.Errorf("EvaluateTypeQuality() OverallQuality = %v, expected %v",
					metrics.OverallQuality, tt.expected.OverallQuality)
			}

			tolerance := 0.1
			if abs(metrics.Helpfulness-tt.expected.Helpfulness) > tolerance {
				t.Errorf("EvaluateTypeQuality() Helpfulness = %v, expected %v",
					metrics.Helpfulness, tt.expected.Helpfulness)
			}
		})
	}
}

func TestQualityController_ShouldIncludeAnnotation(t *testing.T) {
	tests := []struct {
		name          string
		options       QualityOptions
		typeInfo      *types.TypeInfo
		shouldInclude bool
	}{
		{
			name:    "high confidence above threshold",
			options: QualityOptions{MinConfidenceThreshold: 0.7},
			typeInfo: types.NewTypeInfo(
				types.NewIntType(),
				0.9,
				types.SourceLiteral,
			),
			shouldInclude: true,
		},
		{
			name:    "low confidence below threshold",
			options: QualityOptions{MinConfidenceThreshold: 0.7},
			typeInfo: types.NewTypeInfo(
				types.NewAnyType(),
				0.5,
				types.SourceContext,
			),
			shouldInclude: false,
		},
		{
			name:    "any type filtered out",
			options: QualityOptions{FilterVagueTypes: true},
			typeInfo: types.NewTypeInfo(
				types.NewAnyType(),
				0.9,
				types.SourceLiteral,
			),
			shouldInclude: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qc := NewQualityController(tt.options)
			result := qc.ShouldIncludeAnnotation(tt.typeInfo)

			if result != tt.shouldInclude {
				t.Errorf("ShouldIncludeAnnotation() = %v, expected %v",
					result, tt.shouldInclude)
			}
		})
	}
}

func TestConflictDetector_DetectConflicts(t *testing.T) {
	detector := NewConflictDetector()

	tests := []struct {
		name        string
		typeInfos   []*types.TypeInfo
		hasConflict bool
		description string
	}{
		{
			name: "no conflicts with same types",
			typeInfos: []*types.TypeInfo{
				types.NewTypeInfo(types.NewIntType(), 0.9, types.SourceLiteral),
				types.NewTypeInfo(types.NewIntType(), 0.8, types.SourceVariable),
			},
			hasConflict: false,
			description: "Same types should not conflict",
		},
		{
			name: "conflict with incompatible types",
			typeInfos: []*types.TypeInfo{
				types.NewTypeInfo(types.NewIntType(), 0.9, types.SourceLiteral),
				types.NewTypeInfo(types.NewStrType(), 0.8, types.SourceVariable),
			},
			hasConflict: true,
			description: "Incompatible types should conflict",
		},
		{
			name: "no conflict with compatible types",
			typeInfos: []*types.TypeInfo{
				types.NewTypeInfo(types.NewIntType(), 0.9, types.SourceLiteral),
				types.NewTypeInfo(types.NewAnyType(), 0.3, types.SourceContext),
			},
			hasConflict: false,
			description: "Any type is compatible with all types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflict := detector.DetectConflicts(tt.typeInfos)
			hasConflict := conflict != nil

			if hasConflict != tt.hasConflict {
				t.Errorf("DetectConflicts() hasConflict = %v, expected %v",
					hasConflict, tt.hasConflict)
			}
		})
	}
}

// Helper function for absolute value of float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
