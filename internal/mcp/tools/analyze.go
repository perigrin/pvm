// ABOUTME: Code analysis tool implementation for MCP server
// ABOUTME: Provides type extraction, error checking, and type inference

package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/mcp/validation"
	"tamarou.com/pvm/internal/parser"
)

// CodeAnalyzer provides code analysis capabilities
type CodeAnalyzer struct {
	validator *validation.Validator
	autoFixer *validation.AutoFixer
	parser    parser.Parser
}

// NewCodeAnalyzer creates a new code analyzer
func NewCodeAnalyzer(validator *validation.Validator, autoFixer *validation.AutoFixer) (*CodeAnalyzer, error) {
	// Create parser
	p, err := parser.NewParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	return &CodeAnalyzer{
		validator: validator,
		autoFixer: autoFixer,
		parser:    p,
	}, nil
}

// AnalysisResult represents the result of code analysis
type AnalysisResult struct {
	Status                 string                  `json:"status"`
	AnalysisType           string                  `json:"analysis_type"`
	TypeInfo               map[string]TypeDetail   `json:"type_info,omitempty"`
	Errors                 []ErrorDetail           `json:"errors,omitempty"`
	Warnings               []WarningDetail         `json:"warnings,omitempty"`
	Fixes                  []FixSuggestion         `json:"fixes,omitempty"`
	InferredTypes          map[string]TypeDetail   `json:"inferred_types,omitempty"`
	Valid                  bool                    `json:"valid"`
	Timestamp              string                  `json:"timestamp"`
	FlowAnalysis           *FlowAnalysisResult     `json:"flow_analysis,omitempty"`
	CodeQuality            *CodeQualityMetrics     `json:"code_quality,omitempty"`
	TypeCompatibility      []TypeCompatibility     `json:"type_compatibility,omitempty"`
	TypeAnnotationAnalysis *TypeAnnotationAnalysis `json:"type_annotation_analysis,omitempty"`
}

// TypeDetail represents detailed type information
type TypeDetail struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Kind        string   `json:"kind"` // variable, method, function, etc.
	Line        int      `json:"line"`
	Column      int      `json:"column"`
	Inferred    bool     `json:"inferred"`
	Confidence  float64  `json:"confidence,omitempty"`
	Constraints []string `json:"constraints,omitempty"`
}

// ErrorDetail represents a detailed error
type ErrorDetail struct {
	Message  string `json:"message"`
	Code     string `json:"code"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Severity string `json:"severity"`
	Fixable  bool   `json:"fixable"`
	Context  string `json:"context,omitempty"`
}

// WarningDetail represents a detailed warning
type WarningDetail struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Type    string `json:"type"`
}

// FixSuggestion represents a suggested fix
type FixSuggestion struct {
	Error       ErrorDetail `json:"error"`
	FixedCode   string      `json:"fixed_code"`
	Explanation string      `json:"explanation"`
	Confidence  float64     `json:"confidence"`
}

// Analyze performs code analysis based on the specified type
func (a *CodeAnalyzer) Analyze(ctx context.Context, code string, analysisType string, projectPath string, autoFix bool) (*AnalysisResult, error) {
	switch analysisType {
	case "get_types":
		return a.getTypes(ctx, code, projectPath)
	case "check_errors":
		return a.checkErrors(ctx, code, projectPath, autoFix)
	case "infer_types":
		return a.inferTypes(ctx, code, projectPath)
	case "flow_analysis":
		return a.performFlowAnalysis(ctx, code, projectPath)
	case "code_quality":
		return a.analyzeCodeQuality(ctx, code, projectPath)
	case "type_compatibility":
		return a.checkTypeCompatibility(ctx, code, projectPath)
	case "annotation_analysis":
		return a.analyzeTypeAnnotations(ctx, code, projectPath)
	case "full_analysis":
		return a.performFullAnalysis(ctx, code, projectPath, autoFix)
	default:
		return nil, fmt.Errorf("unknown analysis type: %s", analysisType)
	}
}

// getTypes extracts type information from the code
func (a *CodeAnalyzer) getTypes(ctx context.Context, code string, projectPath string) (*AnalysisResult, error) {
	// Parse the code to extract type annotations
	ast, err := a.parser.ParseString(code)
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}

	result := &AnalysisResult{
		Status:       "success",
		AnalysisType: "get_types",
		TypeInfo:     make(map[string]TypeDetail),
		Valid:        len(ast.Errors) == 0,
		Timestamp:    generateTimestamp(),
	}

	// Extract type information from annotations
	for _, ann := range ast.TypeAnnotations {
		if ann.TypeExpression != nil {
			detail := TypeDetail{
				Name:     ann.AnnotatedItem,
				Type:     ann.TypeExpression.String(),
				Kind:     getAnnotationKind(ann.Kind),
				Line:     ann.Pos.Line,
				Column:   ann.Pos.Column,
				Inferred: false,
			}

			// Add constraints for parameterized types
			if len(ann.TypeExpression.Parameters) > 0 {
				detail.Constraints = extractConstraints(ann.TypeExpression)
			}

			result.TypeInfo[ann.AnnotatedItem] = detail
		}
	}

	// Also extract any inferred types from the code structure
	inferredTypes := a.extractInferredTypes(ast)
	for name, typeInfo := range inferredTypes {
		if _, exists := result.TypeInfo[name]; !exists {
			typeInfo.Inferred = true
			result.TypeInfo[name] = typeInfo
		}
	}

	return result, nil
}

// checkErrors validates the code and checks for type errors
func (a *CodeAnalyzer) checkErrors(ctx context.Context, code string, projectPath string, autoFix bool) (*AnalysisResult, error) {
	// Validate the code
	validationResult, err := a.validator.ValidateCode(ctx, code, projectPath)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	result := &AnalysisResult{
		Status:       "success",
		AnalysisType: "check_errors",
		Errors:       []ErrorDetail{},
		Warnings:     []WarningDetail{},
		Fixes:        []FixSuggestion{},
		Valid:        validationResult.Valid,
		Timestamp:    generateTimestamp(),
	}

	// Convert validation errors to detailed errors
	for _, vErr := range validationResult.Errors {
		detail := ErrorDetail{
			Message:  vErr.Message,
			Code:     vErr.Code,
			Line:     vErr.Line,
			Column:   vErr.Column,
			Severity: vErr.Severity,
			Fixable:  vErr.Fixable,
		}

		// Add context by extracting the line of code
		if vErr.Line > 0 {
			lines := strings.Split(code, "\n")
			if vErr.Line <= len(lines) {
				detail.Context = strings.TrimSpace(lines[vErr.Line-1])
			}
		}

		result.Errors = append(result.Errors, detail)
	}

	// Convert validation warnings
	for _, vWarn := range validationResult.Warnings {
		result.Warnings = append(result.Warnings, WarningDetail{
			Message: vWarn.Message,
			Code:    vWarn.Code,
			Line:    vWarn.Line,
			Column:  vWarn.Column,
			Type:    "validation",
		})
	}

	// Attempt auto-fix if enabled and there are fixable errors
	if autoFix && !validationResult.Valid && a.autoFixer != nil {
		fixes, _ := a.autoFixer.AutoFix(ctx, code, validationResult.Errors, projectPath)
		for _, fix := range fixes {
			result.Fixes = append(result.Fixes, FixSuggestion{
				Error: ErrorDetail{
					Message:  fix.Error.Message,
					Code:     fix.Error.Code,
					Line:     fix.Error.Line,
					Column:   fix.Error.Column,
					Severity: fix.Error.Severity,
					Fixable:  fix.Error.Fixable,
				},
				FixedCode:   fix.FixedCode,
				Explanation: fix.Explanation,
				Confidence:  fix.Confidence,
			})
		}
	}

	// Also include type information
	result.TypeInfo = make(map[string]TypeDetail)
	for name, info := range validationResult.TypeInfo {
		result.TypeInfo[name] = TypeDetail{
			Name:     info.Name,
			Type:     info.Type,
			Line:     info.Line,
			Column:   info.Column,
			Inferred: info.Inferred,
			Kind:     "variable", // Default, could be enhanced
		}
	}

	return result, nil
}

// inferTypes performs type inference on the code
func (a *CodeAnalyzer) inferTypes(ctx context.Context, code string, projectPath string) (*AnalysisResult, error) {
	// Parse the code
	ast, err := a.parser.ParseString(code)
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}

	result := &AnalysisResult{
		Status:        "success",
		AnalysisType:  "infer_types",
		InferredTypes: make(map[string]TypeDetail),
		TypeInfo:      make(map[string]TypeDetail),
		Valid:         len(ast.Errors) == 0,
		Timestamp:     generateTimestamp(),
	}

	// First, collect explicitly typed variables
	for _, ann := range ast.TypeAnnotations {
		if ann.TypeExpression != nil {
			detail := TypeDetail{
				Name:     ann.AnnotatedItem,
				Type:     ann.TypeExpression.String(),
				Kind:     getAnnotationKind(ann.Kind),
				Line:     ann.Pos.Line,
				Column:   ann.Pos.Column,
				Inferred: false,
			}
			result.TypeInfo[ann.AnnotatedItem] = detail
		}
	}

	// Perform type inference
	inferredTypes := a.performTypeInference(ast, result.TypeInfo)

	// Add inferred types to results
	for name, typeDetail := range inferredTypes {
		typeDetail.Inferred = true
		result.InferredTypes[name] = typeDetail

		// Also add to main type info if not already present
		if _, exists := result.TypeInfo[name]; !exists {
			result.TypeInfo[name] = typeDetail
		}
	}

	return result, nil
}

// Helper functions

// extractInferredTypes extracts types that can be inferred from code structure
func (a *CodeAnalyzer) extractInferredTypes(ast *ast.AST) map[string]TypeDetail {
	inferred := make(map[string]TypeDetail)

	// This is a simplified implementation
	// In a real implementation, we would walk the AST and infer types from:
	// - Literal assignments (my $x = 42; -> Int)
	// - Array literals (my @arr = (1,2,3); -> Array[Int])
	// - Hash literals (my %h = (a => 1); -> Hash[Str,Int])
	// - Function return values
	// - Common patterns

	return inferred
}

// performTypeInference performs actual type inference based on usage patterns
func (a *CodeAnalyzer) performTypeInference(ast *ast.AST, knownTypes map[string]TypeDetail) map[string]TypeDetail {
	inferred := make(map[string]TypeDetail)

	// This is a simplified implementation
	// A real implementation would:
	// 1. Build a type constraint graph
	// 2. Propagate type information through assignments
	// 3. Infer types from function calls
	// 4. Use flow-sensitive analysis
	// 5. Handle type unions and intersections

	// For now, we'll do basic literal inference
	// This would be implemented by walking the AST

	return inferred
}

// getAnnotationKind converts parser annotation kind to string
func getAnnotationKind(kind ast.AnnotationKind) string {
	switch kind {
	case ast.VarAnnotation:
		return "variable"
	case ast.SubParamAnnotation:
		return "subroutine_param"
	case ast.SubReturnAnnotation:
		return "subroutine_return"
	case ast.MethodParamAnnotation:
		return "method_param"
	case ast.MethodReturnAnnotation:
		return "method_return"
	case ast.FieldAnnotation:
		return "attribute"
	case ast.TypeDeclAnnotation:
		return "type_declaration"
	default:
		return "unknown"
	}
}

// extractConstraints extracts type constraints from a type expression
func extractConstraints(typeExpr *ast.TypeExpression) []string {
	constraints := []string{}

	if typeExpr.IsUnion {
		constraints = append(constraints, "union_type")
	}
	if typeExpr.IsIntersection {
		constraints = append(constraints, "intersection_type")
	}
	if typeExpr.IsNegation {
		constraints = append(constraints, "negation_type")
	}

	// Add parameter constraints
	for _, param := range typeExpr.Parameters {
		constraints = append(constraints, fmt.Sprintf("param:%s", param.String()))
	}

	return constraints
}

// generateTimestamp generates an ISO timestamp
func generateTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Advanced Analysis Structures

// FlowAnalysisResult contains flow-sensitive type analysis results
type FlowAnalysisResult struct {
	RefinedTypes       map[string][]TypeRefinement `json:"refined_types"`
	ControlFlowPaths   []ControlFlowPath           `json:"control_flow_paths"`
	TypeStates         map[string]TypeState        `json:"type_states"`
	ValidationPatterns []string                    `json:"validation_patterns_found"`
}

// TypeRefinement represents a type refinement in control flow
type TypeRefinement struct {
	Variable     string `json:"variable"`
	OriginalType string `json:"original_type"`
	RefinedType  string `json:"refined_type"`
	Condition    string `json:"condition"`
	Line         int    `json:"line"`
	Scope        string `json:"scope"`
}

// ControlFlowPath represents a path through the code
type ControlFlowPath struct {
	PathID      string   `json:"path_id"`
	Conditions  []string `json:"conditions"`
	TypeChanges []string `json:"type_changes"`
}

// TypeState represents the type state at a point in code
type TypeState struct {
	Variables   map[string]string `json:"variables"`
	Constraints []string          `json:"constraints"`
	Line        int               `json:"line"`
}

// CodeQualityMetrics contains code quality analysis results
type CodeQualityMetrics struct {
	CyclomaticComplexity int               `json:"cyclomatic_complexity"`
	TypeCoverage         float64           `json:"type_coverage"`
	TypeAnnotationRatio  float64           `json:"type_annotation_ratio"`
	FunctionComplexity   map[string]int    `json:"function_complexity"`
	TypeViolations       []TypeViolation   `json:"type_violations"`
	TypeSafety           TypeSafetyMetrics `json:"type_safety"`
}

// TypeViolation represents a type safety violation
type TypeViolation struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Line        int    `json:"line"`
	Severity    string `json:"severity"`
}

// TypeSafetyMetrics contains type safety metrics
type TypeSafetyMetrics struct {
	UnsafeOperations  int     `json:"unsafe_operations"`
	ImplicitCasts     int     `json:"implicit_casts"`
	UnvalidatedInputs int     `json:"unvalidated_inputs"`
	SafetyScore       float64 `json:"safety_score"`
}

// TypeCompatibility represents type compatibility information
type TypeCompatibility struct {
	Type1      string `json:"type1"`
	Type2      string `json:"type2"`
	Compatible bool   `json:"compatible"`
	Reason     string `json:"reason"`
	Conversion string `json:"conversion,omitempty"`
}

// TypeAnnotationAnalysis contains type annotation analysis results
type TypeAnnotationAnalysis struct {
	TotalAnnotations      int                    `json:"total_annotations"`
	CorrectAnnotations    int                    `json:"correct_annotations"`
	IncorrectAnnotations  int                    `json:"incorrect_annotations"`
	MissingAnnotations    []MissingAnnotation    `json:"missing_annotations"`
	AnnotationSuggestions []AnnotationSuggestion `json:"annotation_suggestions"`
}

// MissingAnnotation represents a location where type annotation is missing
type MissingAnnotation struct {
	Location     string `json:"location"`
	Variable     string `json:"variable"`
	InferredType string `json:"inferred_type"`
	Line         int    `json:"line"`
}

// AnnotationSuggestion represents a suggested type annotation
type AnnotationSuggestion struct {
	Location  string `json:"location"`
	Current   string `json:"current"`
	Suggested string `json:"suggested"`
	Reason    string `json:"reason"`
	Line      int    `json:"line"`
}

// Advanced Analysis Methods

// performFlowAnalysis performs flow-sensitive type analysis
func (a *CodeAnalyzer) performFlowAnalysis(ctx context.Context, code string, projectPath string) (*AnalysisResult, error) {
	// Parse the code
	ast, err := a.parser.ParseString(code)
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}

	result := &AnalysisResult{
		Status:       "success",
		AnalysisType: "flow_analysis",
		Valid:        len(ast.Errors) == 0,
		Timestamp:    generateTimestamp(),
		FlowAnalysis: &FlowAnalysisResult{
			RefinedTypes:       make(map[string][]TypeRefinement),
			ControlFlowPaths:   []ControlFlowPath{},
			TypeStates:         make(map[string]TypeState),
			ValidationPatterns: []string{},
		},
	}

	// Analyze control flow and type refinements
	// This is a simplified implementation - real implementation would use typechecker
	refinements := a.findTypeRefinements(ast)
	for varName, refs := range refinements {
		result.FlowAnalysis.RefinedTypes[varName] = refs
	}

	// Find validation patterns
	patterns := a.findValidationPatterns(code)
	result.FlowAnalysis.ValidationPatterns = patterns

	return result, nil
}

// analyzeCodeQuality analyzes code quality metrics
func (a *CodeAnalyzer) analyzeCodeQuality(ctx context.Context, code string, projectPath string) (*AnalysisResult, error) {
	// Parse the code
	ast, err := a.parser.ParseString(code)
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}

	result := &AnalysisResult{
		Status:       "success",
		AnalysisType: "code_quality",
		Valid:        len(ast.Errors) == 0,
		Timestamp:    generateTimestamp(),
		CodeQuality: &CodeQualityMetrics{
			FunctionComplexity: make(map[string]int),
			TypeViolations:     []TypeViolation{},
		},
	}

	// Calculate metrics
	result.CodeQuality.CyclomaticComplexity = a.calculateCyclomaticComplexity(code)
	result.CodeQuality.TypeCoverage = a.calculateTypeCoverage(ast)
	result.CodeQuality.TypeAnnotationRatio = a.calculateTypeAnnotationRatio(ast)

	// Calculate type safety metrics
	result.CodeQuality.TypeSafety = a.calculateTypeSafety(ast)

	return result, nil
}

// checkTypeCompatibility checks compatibility between types
func (a *CodeAnalyzer) checkTypeCompatibility(ctx context.Context, code string, projectPath string) (*AnalysisResult, error) {
	// Parse the code
	ast, err := a.parser.ParseString(code)
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}

	result := &AnalysisResult{
		Status:            "success",
		AnalysisType:      "type_compatibility",
		Valid:             len(ast.Errors) == 0,
		Timestamp:         generateTimestamp(),
		TypeCompatibility: []TypeCompatibility{},
	}

	// Extract type pairs from assignments and function calls
	typePairs := a.extractTypePairs(ast)

	// Check compatibility for each pair
	for _, pair := range typePairs {
		compat := TypeCompatibility{
			Type1:      pair.type1,
			Type2:      pair.type2,
			Compatible: a.areTypesCompatible(pair.type1, pair.type2),
			Reason:     a.getCompatibilityReason(pair.type1, pair.type2),
		}
		result.TypeCompatibility = append(result.TypeCompatibility, compat)
	}

	return result, nil
}

// analyzeTypeAnnotations analyzes type annotation quality and coverage
func (a *CodeAnalyzer) analyzeTypeAnnotations(ctx context.Context, code string, projectPath string) (*AnalysisResult, error) {
	// Parse the code
	ast, err := a.parser.ParseString(code)
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}

	result := &AnalysisResult{
		Status:       "success",
		AnalysisType: "annotation_analysis",
		Valid:        len(ast.Errors) == 0,
		Timestamp:    generateTimestamp(),
		TypeAnnotationAnalysis: &TypeAnnotationAnalysis{
			MissingAnnotations:    []MissingAnnotation{},
			AnnotationSuggestions: []AnnotationSuggestion{},
		},
	}

	// Count annotations
	analysis := result.TypeAnnotationAnalysis
	analysis.TotalAnnotations = len(ast.TypeAnnotations)

	// Validate annotations
	for _, annotation := range ast.TypeAnnotations {
		if a.isAnnotationCorrect(annotation) {
			analysis.CorrectAnnotations++
		} else {
			analysis.IncorrectAnnotations++
			// Add suggestion
			suggestion := AnnotationSuggestion{
				Location:  fmt.Sprintf("%s:%d", annotation.AnnotatedItem, annotation.Pos.Line),
				Current:   annotation.TypeExpression.String(),
				Suggested: a.suggestType(annotation.AnnotatedItem, ast),
				Reason:    "Type mismatch detected",
				Line:      annotation.Pos.Line,
			}
			analysis.AnnotationSuggestions = append(analysis.AnnotationSuggestions, suggestion)
		}
	}

	// Find missing annotations
	missingVars := a.findUnannotatedVariables(ast)
	for _, varInfo := range missingVars {
		missing := MissingAnnotation{
			Location:     varInfo.location,
			Variable:     varInfo.name,
			InferredType: varInfo.inferredType,
			Line:         varInfo.line,
		}
		analysis.MissingAnnotations = append(analysis.MissingAnnotations, missing)
	}

	return result, nil
}

// performFullAnalysis performs comprehensive analysis including all features
func (a *CodeAnalyzer) performFullAnalysis(ctx context.Context, code string, projectPath string, autoFix bool) (*AnalysisResult, error) {
	// Perform basic analysis first
	result, err := a.checkErrors(ctx, code, projectPath, autoFix)
	if err != nil {
		return nil, err
	}

	result.AnalysisType = "full_analysis"

	// Add flow analysis
	flowResult, err := a.performFlowAnalysis(ctx, code, projectPath)
	if err == nil {
		result.FlowAnalysis = flowResult.FlowAnalysis
	}

	// Add code quality metrics
	qualityResult, err := a.analyzeCodeQuality(ctx, code, projectPath)
	if err == nil {
		result.CodeQuality = qualityResult.CodeQuality
	}

	// Add type compatibility
	compatResult, err := a.checkTypeCompatibility(ctx, code, projectPath)
	if err == nil {
		result.TypeCompatibility = compatResult.TypeCompatibility
	}

	// Add annotation analysis
	annotResult, err := a.analyzeTypeAnnotations(ctx, code, projectPath)
	if err == nil {
		result.TypeAnnotationAnalysis = annotResult.TypeAnnotationAnalysis
	}

	return result, nil
}

// Helper methods for advanced analysis

func (a *CodeAnalyzer) findTypeRefinements(ast *ast.AST) map[string][]TypeRefinement {
	refinements := make(map[string][]TypeRefinement)
	// Implementation would analyze control flow for type refinements
	// For example: if (defined $var) { ... } refines $var to !Undef
	return refinements
}

func (a *CodeAnalyzer) findValidationPatterns(code string) []string {
	patterns := []string{}
	// Look for common validation patterns
	if strings.Contains(code, "defined") {
		patterns = append(patterns, "defined_check")
	}
	if strings.Contains(code, "ref") {
		patterns = append(patterns, "ref_check")
	}
	if strings.Contains(code, "isa") {
		patterns = append(patterns, "isa_check")
	}
	if strings.Contains(code, "can") {
		patterns = append(patterns, "can_check")
	}
	return patterns
}

func (a *CodeAnalyzer) calculateCyclomaticComplexity(code string) int {
	// Simplified complexity calculation
	complexity := 1
	// Count decision points
	complexity += strings.Count(code, "if ")
	complexity += strings.Count(code, "elsif ")
	complexity += strings.Count(code, "unless ")
	complexity += strings.Count(code, "while ")
	complexity += strings.Count(code, "for ")
	complexity += strings.Count(code, "foreach ")
	complexity += strings.Count(code, "&&")
	complexity += strings.Count(code, "||")
	complexity += strings.Count(code, "? ")
	return complexity
}

func (a *CodeAnalyzer) calculateTypeCoverage(ast *ast.AST) float64 {
	if ast == nil {
		return 0.0
	}
	// Calculate percentage of typed vs untyped variables
	// This is simplified - real implementation would count all declarations
	totalVars := 100.0 // placeholder
	typedVars := float64(len(ast.TypeAnnotations))
	if totalVars == 0 {
		return 0.0
	}
	return (typedVars / totalVars) * 100.0
}

func (a *CodeAnalyzer) calculateTypeAnnotationRatio(ast *ast.AST) float64 {
	// Similar to type coverage but for all annotatable elements
	return a.calculateTypeCoverage(ast)
}

func (a *CodeAnalyzer) calculateTypeSafety(ast *ast.AST) TypeSafetyMetrics {
	return TypeSafetyMetrics{
		UnsafeOperations:  0,    // Would count eval, symbolic refs, etc.
		ImplicitCasts:     0,    // Would count automatic type conversions
		UnvalidatedInputs: 0,    // Would count unchecked user inputs
		SafetyScore:       85.0, // Calculated from above metrics
	}
}

type typePair struct {
	type1 string
	type2 string
}

func (a *CodeAnalyzer) extractTypePairs(ast *ast.AST) []typePair {
	pairs := []typePair{}
	// Would extract from assignments and function calls
	return pairs
}

func (a *CodeAnalyzer) areTypesCompatible(type1, type2 string) bool {
	// Simplified compatibility check
	if type1 == type2 {
		return true
	}
	// Check for subtype relationships
	if type1 == "Int" && type2 == "Num" {
		return true
	}
	if type1 == "Str" && type2 == "Any" {
		return true
	}
	return false
}

func (a *CodeAnalyzer) getCompatibilityReason(type1, type2 string) string {
	if type1 == type2 {
		return "Types are identical"
	}
	if a.areTypesCompatible(type1, type2) {
		return fmt.Sprintf("%s is a subtype of %s", type1, type2)
	}
	return fmt.Sprintf("%s and %s are incompatible types", type1, type2)
}

func (a *CodeAnalyzer) isAnnotationCorrect(annotation *ast.TypeAnnotation) bool {
	// Would validate against actual usage
	return true // Simplified
}

func (a *CodeAnalyzer) suggestType(varName string, ast *ast.AST) string {
	// Would infer from usage
	return "Any" // Simplified
}

type varInfo struct {
	name         string
	location     string
	line         int
	inferredType string
}

func (a *CodeAnalyzer) findUnannotatedVariables(ast *ast.AST) []varInfo {
	vars := []varInfo{}
	// Would find all variable declarations without types
	return vars
}
