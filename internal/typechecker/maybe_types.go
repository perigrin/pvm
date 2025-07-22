// ABOUTME: Comprehensive Maybe type operations and safety analysis for optional values
// ABOUTME: Implements null safety, unwrapping operations, and integration with Perl idioms

package typechecker

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/errors"
)

// MaybeTypeHandler provides comprehensive Maybe type operations
type MaybeTypeHandler struct {
	// Checker is the parent type checker
	Checker *TypeChecker

	// UnsafeAccesses tracks potentially unsafe Maybe value accesses
	UnsafeAccesses []UnsafeAccess

	// WrapWarnings tracks potentially unnecessary Maybe wrapping
	WrapWarnings []WrapWarning

	// PerlIdiomIntegrations tracks integration with Perl idioms
	PerlIdiomIntegrations map[string]PerlIdiomType
}

// UnsafeAccess represents a potentially unsafe access to a Maybe value
type UnsafeAccess struct {
	Variable   string
	MaybeType  string
	AccessType string // "direct", "method_call", "property_access"
	Position   ast.Position
	Context    string // Additional context about the access
}

// WrapWarning represents a potentially unnecessary Maybe wrapping
type WrapWarning struct {
	Variable string
	FromType string
	ToType   string
	Position ast.Position
	Reason   string
}

// PerlIdiomType represents different Perl idioms that work with Maybe types
type PerlIdiomType int

const (
	DefinedOrOperator PerlIdiomType = iota // //
	LogicalOrAssign                        // ||=
	ExistsCheck                            // exists()
	DefinedCheck                           // defined()
	UndefCheck                             // !defined()
)

// NewMaybeTypeHandler creates a new Maybe type handler
func NewMaybeTypeHandler(checker *TypeChecker) *MaybeTypeHandler {
	return &MaybeTypeHandler{
		Checker:               checker,
		UnsafeAccesses:        []UnsafeAccess{},
		WrapWarnings:          []WrapWarning{},
		PerlIdiomIntegrations: make(map[string]PerlIdiomType),
	}
}

// IsMaybeType checks if a type is a Maybe type
func (m *MaybeTypeHandler) IsMaybeType(typeStr string) bool {
	return strings.HasPrefix(typeStr, "Maybe[") && strings.HasSuffix(typeStr, "]")
}

// ExtractMaybeParameter extracts the wrapped type from Maybe[T]
func (m *MaybeTypeHandler) ExtractMaybeParameter(maybeType string) string {
	if m.IsMaybeType(maybeType) {
		return maybeType[6 : len(maybeType)-1]
	}
	return ""
}

// CreateMaybeType creates a Maybe[T] type from a base type
func (m *MaybeTypeHandler) CreateMaybeType(baseType string) string {
	// Avoid double-wrapping
	if m.IsMaybeType(baseType) {
		return baseType
	}
	return fmt.Sprintf("Maybe[%s]", baseType)
}

// IsNullableType checks if a type can be null/undef
func (m *MaybeTypeHandler) IsNullableType(typeStr string) bool {
	return typeStr == "Undef" || m.IsMaybeType(typeStr) || m.IsUnionWithUndef(typeStr)
}

// IsUnionWithUndef checks if a type is a union that includes Undef
func (m *MaybeTypeHandler) IsUnionWithUndef(typeStr string) bool {
	if !strings.Contains(typeStr, "|") {
		return false
	}

	parts := strings.Split(typeStr, "|")
	for _, part := range parts {
		if strings.TrimSpace(part) == "Undef" {
			return true
		}
	}
	return false
}

// CheckUnsafeAccess analyzes if an access to a Maybe value is potentially unsafe
func (m *MaybeTypeHandler) CheckUnsafeAccess(variable, variableType, accessType string, pos ast.Position) error {
	if !m.IsMaybeType(variableType) {
		return nil // Not a Maybe type
	}

	// Record the unsafe access
	unsafeAccess := UnsafeAccess{
		Variable:   variable,
		MaybeType:  variableType,
		AccessType: accessType,
		Position:   pos,
		Context:    fmt.Sprintf("Accessing %s value without null check", variableType),
	}
	m.UnsafeAccesses = append(m.UnsafeAccesses, unsafeAccess)

	// Return a warning (not an error, to allow gradual adoption)
	return errors.NewTypeError(
		"WARNING_UNSAFE_MAYBE_ACCESS",
		fmt.Sprintf("Unsafe access to Maybe type '%s' at %d:%d - consider checking with defined(%s) first",
			variableType, pos.Line, pos.Column, variable),
		nil,
	)
}

// CheckSafeUnwrapping validates that Maybe values are safely unwrapped
func (m *MaybeTypeHandler) CheckSafeUnwrapping(variable, variableType string, pos ast.Position, context *MaybeFlowContext) (string, error) {
	if !m.IsMaybeType(variableType) {
		return variableType, nil
	}

	wrappedType := m.ExtractMaybeParameter(variableType)

	// Check if we're in a safe context (after a defined() check)
	if context != nil && m.isInSafeContext(variable, context) {
		// Safe to unwrap - return the wrapped type
		return wrappedType, nil
	}

	// Unsafe unwrapping
	return variableType, m.CheckUnsafeAccess(variable, variableType, "unwrap", pos)
}

// isInSafeContext checks if a variable is in a context where it's safe to unwrap
func (m *MaybeTypeHandler) isInSafeContext(variable string, context *MaybeFlowContext) bool {
	if context == nil {
		return false
	}

	// Check if we've recently seen a defined() check for this variable
	for _, condition := range context.Conditions {
		if m.isDefiningCondition(condition, variable) {
			return true
		}
	}

	// Check refined types
	if refinedType, exists := context.RefinedTypes[variable]; exists {
		// If the refined type is not a Maybe type, it's safe
		return !m.IsMaybeType(refinedType)
	}

	return false
}

// isDefiningCondition checks if a condition establishes that a variable is defined
func (m *MaybeTypeHandler) isDefiningCondition(condition MaybeCondition, variable string) bool {
	// Handle different types of defining conditions
	switch condition.Type {
	case "defined":
		return condition.Variable == variable && condition.Positive
	case "not_undef":
		return condition.Variable == variable && condition.Positive
	case "exists":
		return condition.Variable == variable && condition.Positive
	default:
		return false
	}
}

// InferMaybeType infers when a type should be Maybe based on context
func (m *MaybeTypeHandler) InferMaybeType(expression string, context *MaybeInferenceContext) string {
	// Check for patterns that commonly return undef
	if m.isUndefReturningPattern(expression) {
		baseType := m.inferBaseType(expression)
		if baseType != "Unknown" && baseType != "Undef" {
			return m.CreateMaybeType(baseType)
		}
	}

	// Check for uninitialized variables
	if m.isUninitializedVariable(expression, context) {
		// Uninitialized variables should be Maybe types
		return m.CreateMaybeType("Any")
	}

	return ""
}

// isUndefReturningPattern checks if an expression commonly returns undef
func (m *MaybeTypeHandler) isUndefReturningPattern(expression string) bool {
	// Common Perl patterns that can return undef
	patterns := []string{
		"shift",    // shift can return undef
		"pop",      // pop can return undef
		"shift @_", // parameter extraction
		"$_[0]",    // array access can be undef
		"split",    // split can return empty
		"readline", // file operations can fail
		"<",        // file read can return undef
		"getline",  // getline can fail
	}

	for _, pattern := range patterns {
		if strings.Contains(expression, pattern) {
			return true
		}
	}

	// Check for hash/array access patterns
	if strings.Contains(expression, "[") || strings.Contains(expression, "{") {
		return true // Array/hash access can return undef
	}

	return false
}

// isUninitializedVariable checks if a variable appears to be uninitialized
func (m *MaybeTypeHandler) isUninitializedVariable(expression string, context *MaybeInferenceContext) bool {
	// This would check if we're looking at a variable that hasn't been initialized
	// For now, simplified implementation
	return false
}

// inferBaseType infers the base type for Maybe wrapping
func (m *MaybeTypeHandler) inferBaseType(expression string) string {
	// Simple heuristics for common patterns
	if strings.Contains(expression, "shift") || strings.Contains(expression, "pop") {
		return "Scalar" // Array operations typically return scalars
	}

	if strings.Contains(expression, "split") {
		return "ArrayRef" // split returns array reference
	}

	if strings.Contains(expression, "readline") || strings.Contains(expression, "getline") || strings.Contains(expression, "<") {
		return "Scalar" // File operations return scalars
	}

	if strings.Contains(expression, "{") {
		return "Scalar" // Hash access returns scalar
	}

	if strings.Contains(expression, "[") {
		return "Scalar" // Array access returns scalar
	}

	return "Unknown"
}

// CheckPerlIdiomIntegration validates integration with Perl idioms
func (m *MaybeTypeHandler) CheckPerlIdiomIntegration(expression string, pos ast.Position) error {
	// Check for defined-or operator: //
	if strings.Contains(expression, "//") {
		return m.handleDefinedOrOperator(expression, pos)
	}

	// Check for logical-or assignment: ||=
	if strings.Contains(expression, "||=") {
		return m.handleLogicalOrAssign(expression, pos)
	}

	// Check for exists() usage
	if strings.Contains(expression, "exists(") {
		return m.handleExistsCheck(expression, pos)
	}

	// Check for defined() usage
	if strings.Contains(expression, "defined(") {
		return m.handleDefinedCheck(expression, pos)
	}

	return nil
}

// handleDefinedOrOperator handles the // operator with Maybe types
func (m *MaybeTypeHandler) handleDefinedOrOperator(expression string, pos ast.Position) error {
	// The // operator is perfect for Maybe types
	// $result = $maybe_value // $default;
	// This pattern should be encouraged for Maybe types

	parts := strings.Split(expression, "//")
	if len(parts) >= 2 {
		leftSide := strings.TrimSpace(parts[0])

		// Check if left side is a Maybe type
		if varType, exists := m.Checker.VariableTypes[leftSide]; exists && m.IsMaybeType(varType) {
			// This is good usage - no warning needed
			m.PerlIdiomIntegrations[expression] = DefinedOrOperator
			return nil
		}
	}

	return nil
}

// handleLogicalOrAssign handles the ||= operator with Maybe types
func (m *MaybeTypeHandler) handleLogicalOrAssign(expression string, pos ast.Position) error {
	// $var ||= $default; pattern
	parts := strings.Split(expression, "||=")
	if len(parts) == 2 {
		varName := strings.TrimSpace(parts[0])

		if varType, exists := m.Checker.VariableTypes[varName]; exists && m.IsMaybeType(varType) {
			// This is good usage
			m.PerlIdiomIntegrations[expression] = LogicalOrAssign
			return nil
		}
	}

	return nil
}

// handleExistsCheck handles exists() checks with Maybe types
func (m *MaybeTypeHandler) handleExistsCheck(expression string, pos ast.Position) error {
	// exists($hash{$key}) with Maybe types
	m.PerlIdiomIntegrations[expression] = ExistsCheck
	return nil
}

// handleDefinedCheck handles defined() checks with Maybe types
func (m *MaybeTypeHandler) handleDefinedCheck(expression string, pos ast.Position) error {
	// defined($var) with Maybe types - this should refine the type
	m.PerlIdiomIntegrations[expression] = DefinedCheck
	return nil
}

// GetUnsafeAccesses returns all detected unsafe accesses
func (m *MaybeTypeHandler) GetUnsafeAccesses() []UnsafeAccess {
	return m.UnsafeAccesses
}

// GetWrapWarnings returns all wrap warnings
func (m *MaybeTypeHandler) GetWrapWarnings() []WrapWarning {
	return m.WrapWarnings
}

// GenerateSafetyReport generates a comprehensive safety report
func (m *MaybeTypeHandler) GenerateSafetyReport() *MaybeSafetyReport {
	return &MaybeSafetyReport{
		UnsafeAccesses:      len(m.UnsafeAccesses),
		WrapWarnings:        len(m.WrapWarnings),
		PerlIdiomUsage:      len(m.PerlIdiomIntegrations),
		SafetyScore:         m.calculateSafetyScore(),
		Recommendations:     m.generateRecommendations(),
		UnsafeAccessDetails: m.UnsafeAccesses,
		WrapWarningDetails:  m.WrapWarnings,
	}
}

// MaybeSafetyReport provides comprehensive analysis of Maybe type usage
type MaybeSafetyReport struct {
	UnsafeAccesses      int
	WrapWarnings        int
	PerlIdiomUsage      int
	SafetyScore         float64
	Recommendations     []string
	UnsafeAccessDetails []UnsafeAccess
	WrapWarningDetails  []WrapWarning
}

// calculateSafetyScore calculates a safety score (0-100)
func (m *MaybeTypeHandler) calculateSafetyScore() float64 {
	totalIssues := len(m.UnsafeAccesses) + len(m.WrapWarnings)
	goodPractices := len(m.PerlIdiomIntegrations)

	if totalIssues == 0 && goodPractices > 0 {
		return 100.0
	}

	if totalIssues == 0 {
		return 90.0
	}

	// Calculate score based on ratio of good practices to issues
	ratio := float64(goodPractices) / float64(totalIssues+goodPractices)
	return ratio * 100.0
}

// generateRecommendations generates safety recommendations
func (m *MaybeTypeHandler) generateRecommendations() []string {
	var recommendations []string

	if len(m.UnsafeAccesses) > 0 {
		recommendations = append(recommendations,
			"Consider adding defined() checks before accessing Maybe values")
		recommendations = append(recommendations,
			"Use the // operator for safe default values with Maybe types")
	}

	if len(m.WrapWarnings) > 0 {
		recommendations = append(recommendations,
			"Review unnecessary Maybe wrapping - some values may not need to be optional")
	}

	if len(m.PerlIdiomIntegrations) == 0 && (len(m.UnsafeAccesses) > 0 || len(m.WrapWarnings) > 0) {
		recommendations = append(recommendations,
			"Consider using Perl idioms like // and ||= for safer Maybe type handling")
	}

	return recommendations
}

// MaybeFlowContext represents the current flow analysis context for Maybe types
type MaybeFlowContext struct {
	Conditions    []MaybeCondition
	RefinedTypes  map[string]string
	SafeVariables map[string]bool
}

// MaybeInferenceContext provides context for Maybe type inference
type MaybeInferenceContext struct {
	VariableTypes   map[string]string
	FunctionReturns map[string]string
	CurrentScope    string
}

// MaybeCondition represents a condition that affects Maybe type refinement
type MaybeCondition struct {
	Type       string // "defined", "not_undef", "exists", etc.
	Variable   string
	Positive   bool   // true for positive condition, false for negative
	Expression string // the full expression
}
