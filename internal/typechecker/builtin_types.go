// ABOUTME: Parser and loader for Perl built-in function type definitions
// ABOUTME: Replaces hardcoded type mapping with declarative type definition files
//
// COVERAGE: Comprehensive support for ~97% of callable Perl built-in functions (238+ functions)
//
// INTENTIONALLY EXCLUDED (~3%):
// - File test operators (-r, -w, -x, etc.) - handled as special operators by parser
// - Quote operators (m//, s///, tr///, etc.) - handled as syntax constructs by lexer
// - Language constructs (sub, import) - compile-time directives, not runtime functions
//
// These require special parser/grammar handling rather than function type definitions.

package typechecker

import (
	"embed"
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/parser"
)

//go:embed perl_builtins.ptd
var builtinTypesFile embed.FS

// BuiltinSignature represents a function's type signature
type BuiltinSignature struct {
	Name       string
	ParamTypes []string
	ReturnType string
	Context    string // "scalar", "list", or ""
	IsVariadic bool   // true if function accepts ...Any
}

// BuiltinTypeRegistry manages built-in function type signatures
type BuiltinTypeRegistry struct {
	functions map[string][]*BuiltinSignature // function name -> possible signatures
}

// NewBuiltinTypeRegistry creates and loads the built-in type registry
func NewBuiltinTypeRegistry() (*BuiltinTypeRegistry, error) {
	registry := &BuiltinTypeRegistry{
		functions: make(map[string][]*BuiltinSignature),
	}

	// Load built-in types from embedded file
	content, err := builtinTypesFile.ReadFile("perl_builtins.ptd")
	if err != nil {
		return nil, fmt.Errorf("failed to read perl_builtins.ptd: %v", err)
	}

	err = registry.parseTypeDefinitions(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse type definitions: %v", err)
	}

	return registry, nil
}

// parseTypeDefinitions parses the .ptd file format using the existing Perl parser
func (r *BuiltinTypeRegistry) parseTypeDefinitions(content string) error {
	// Create a parser instance
	p, err := parser.NewParser()
	if err != nil {
		return fmt.Errorf("failed to create parser: %v", err)
	}

	// Parse the type definitions as if they were Perl code
	astResult, err := p.ParseString(content)
	if err != nil {
		return fmt.Errorf("failed to parse type definitions: %v", err)
	}

	// Extract function signatures from the AST
	return r.extractSignaturesFromAST(astResult)
}

// extractSignaturesFromAST extracts function signatures from the parsed AST
func (r *BuiltinTypeRegistry) extractSignaturesFromAST(astResult *ast.AST) error {
	if astResult == nil || astResult.Root == nil {
		return fmt.Errorf("empty AST")
	}

	// Walk through the AST looking for subroutine definitions
	for _, child := range astResult.Root.Children() {
		if err := r.processASTNode(child); err != nil {
			return err
		}
	}

	return nil
}

// processASTNode processes a single AST node looking for function signatures
func (r *BuiltinTypeRegistry) processASTNode(node ast.Node) error {
	if node == nil {
		return nil
	}

	// Look for subroutine definitions with type signatures
	if subDef, ok := node.(*ast.SubDecl); ok {
		return r.extractSignatureFromSubDef(subDef)
	}

	// Recursively process child nodes
	for _, child := range node.Children() {
		if err := r.processASTNode(child); err != nil {
			return err
		}
	}

	return nil
}

// extractSignatureFromSubDef extracts type signature from a subroutine definition
func (r *BuiltinTypeRegistry) extractSignatureFromSubDef(subDef *ast.SubDecl) error {
	if subDef == nil || subDef.Name == "" {
		return nil
	}

	funcName := subDef.Name
	var paramTypes []string
	isVariadic := false
	returnType := "Any" // Default return type

	// Extract parameter types
	for _, param := range subDef.Parameters() {
		if param.TypeExpr != nil && param.TypeExpr.Name != "" {
			typeStr := param.TypeExpr.Name
			if strings.HasPrefix(typeStr, "...") {
				isVariadic = true
				paramTypes = append(paramTypes, strings.TrimPrefix(typeStr, "..."))
			} else {
				paramTypes = append(paramTypes, typeStr)
			}
		} else {
			paramTypes = append(paramTypes, "Any")
		}
	}

	// Extract return type - Debug what we're getting
	if subDef.ReturnType != nil {
		if subDef.ReturnType.Name != "" {
			returnType = subDef.ReturnType.Name
		}
		// Also try to build full type string for complex types
		if returnType == "Any" || returnType == "" {
			returnType = r.buildTypeString(subDef.ReturnType)
		}
	}

	// Create signature
	signature := &BuiltinSignature{
		Name:       funcName,
		ParamTypes: paramTypes,
		ReturnType: returnType,
		Context:    "", // Could be enhanced to detect context from comments
		IsVariadic: isVariadic,
	}

	// Add to registry (functions can have multiple signatures)
	r.functions[funcName] = append(r.functions[funcName], signature)

	return nil
}

// buildTypeString reconstructs a type string from a TypeExpression
func (r *BuiltinTypeRegistry) buildTypeString(typeExpr *ast.TypeExpression) string {
	if typeExpr == nil {
		return "Any"
	}

	if typeExpr.Name != "" {
		// Handle parameterized types like ArrayRef[Str]
		if len(typeExpr.Parameters) > 0 {
			var params []string
			for _, param := range typeExpr.Parameters {
				params = append(params, r.buildTypeString(param))
			}
			return typeExpr.Name + "[" + strings.Join(params, ", ") + "]"
		}

		// Handle union types like Str|ArrayRef[Int]
		if typeExpr.IsUnion {
			// For union types, we may need to check AlternativeTypes or similar
			// For now, just return the Name part
			return typeExpr.Name
		}

		return typeExpr.Name
	}

	return "Any"
}

// GetFunctionType returns the return type for a built-in function
func (r *BuiltinTypeRegistry) GetFunctionType(funcName string, paramCount int) string {
	signatures, exists := r.functions[funcName]
	if !exists {
		return ""
	}

	// Find best matching signature
	for _, sig := range signatures {
		if r.signatureMatches(sig, paramCount) {
			return sig.ReturnType
		}
	}

	// If no exact match, return first signature's return type
	if len(signatures) > 0 {
		return signatures[0].ReturnType
	}

	return ""
}

// GetFunctionSignatures returns all signatures for a function
func (r *BuiltinTypeRegistry) GetFunctionSignatures(funcName string) []*BuiltinSignature {
	return r.functions[funcName]
}

// IsBuiltinFunction checks if a function is a known built-in
func (r *BuiltinTypeRegistry) IsBuiltinFunction(funcName string) bool {
	_, exists := r.functions[funcName]
	return exists
}

// signatureMatches checks if a signature matches the given parameter count
func (r *BuiltinTypeRegistry) signatureMatches(sig *BuiltinSignature, paramCount int) bool {
	if sig.IsVariadic {
		// Variadic functions can accept any number of parameters >= required params
		requiredParams := len(sig.ParamTypes) - 1 // -1 for the ...Any parameter
		return paramCount >= requiredParams
	}

	// Exact parameter count match
	return paramCount == len(sig.ParamTypes)
}

// ListAllBuiltins returns all known built-in function names
func (r *BuiltinTypeRegistry) ListAllBuiltins() []string {
	var names []string
	for name := range r.functions {
		names = append(names, name)
	}
	return names
}

// GetTypeSignatureString returns a human-readable signature for debugging
func (sig *BuiltinSignature) GetTypeSignatureString() string {
	paramStr := strings.Join(sig.ParamTypes, ", ")
	if sig.IsVariadic && len(sig.ParamTypes) > 0 {
		// Replace last param with ...param
		lastParam := sig.ParamTypes[len(sig.ParamTypes)-1]
		if len(sig.ParamTypes) > 1 {
			paramStr = strings.Join(sig.ParamTypes[:len(sig.ParamTypes)-1], ", ") + ", ..." + lastParam
		} else {
			paramStr = "..." + lastParam
		}
	}

	result := fmt.Sprintf("%s(%s) -> %s", sig.Name, paramStr, sig.ReturnType)
	if sig.Context != "" {
		result += " [" + sig.Context + " context]"
	}

	return result
}
