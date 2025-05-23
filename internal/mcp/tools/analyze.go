// ABOUTME: Code analysis tool implementation for MCP server
// ABOUTME: Provides type extraction, error checking, and type inference

package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	Status        string                `json:"status"`
	AnalysisType  string                `json:"analysis_type"`
	TypeInfo      map[string]TypeDetail `json:"type_info,omitempty"`
	Errors        []ErrorDetail         `json:"errors,omitempty"`
	Warnings      []WarningDetail       `json:"warnings,omitempty"`
	Fixes         []FixSuggestion       `json:"fixes,omitempty"`
	InferredTypes map[string]TypeDetail `json:"inferred_types,omitempty"`
	Valid         bool                  `json:"valid"`
	Timestamp     string                `json:"timestamp"`
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
			if len(ann.TypeExpression.Params) > 0 {
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
func (a *CodeAnalyzer) extractInferredTypes(ast *parser.AST) map[string]TypeDetail {
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
func (a *CodeAnalyzer) performTypeInference(ast *parser.AST, knownTypes map[string]TypeDetail) map[string]TypeDetail {
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
func getAnnotationKind(kind parser.AnnotationKind) string {
	switch kind {
	case parser.VarAnnotation:
		return "variable"
	case parser.SubParamAnnotation:
		return "subroutine_param"
	case parser.SubReturnAnnotation:
		return "subroutine_return"
	case parser.MethodParamAnnotation:
		return "method_param"
	case parser.MethodReturnAnnotation:
		return "method_return"
	case parser.AttrAnnotation:
		return "attribute"
	case parser.TypeDeclAnnotation:
		return "type_declaration"
	default:
		return "unknown"
	}
}

// extractConstraints extracts type constraints from a type expression
func extractConstraints(typeExpr *parser.TypeExpression) []string {
	constraints := []string{}

	if typeExpr.Union {
		constraints = append(constraints, "union_type")
	}
	if typeExpr.Intersection {
		constraints = append(constraints, "intersection_type")
	}
	if typeExpr.Negation {
		constraints = append(constraints, "negation_type")
	}

	// Add parameter constraints
	for _, param := range typeExpr.Params {
		constraints = append(constraints, fmt.Sprintf("param:%s", param.String()))
	}

	return constraints
}

// generateTimestamp generates an ISO timestamp
func generateTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
