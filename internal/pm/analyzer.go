// ABOUTME: Real module analysis for accurate type definition generation
// ABOUTME: Analyzes Perl modules using AST traversal to extract type information

package pm

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

// ModuleAnalyzer performs analysis of Perl modules to extract type information
type ModuleAnalyzer struct {
	// Parser for AST generation
	parser parser.Parser

	// Module path being analyzed
	modulePath string

	// Module name (derived from path or package declaration)
	moduleName string

	// Extracted type information
	types    []typedef.TypeInfo
	packages []typedef.PackageInfo
	subs     []typedef.SubInfo
	methods  []typedef.MethodInfo

	// Symbol tracking for analysis
	currentPackage  string
	exportedSymbols map[string]bool

	// Version information extracted from module
	version string
}

// NewModuleAnalyzer creates a new module analyzer
func NewModuleAnalyzer() (*ModuleAnalyzer, error) {
	p, err := parser.NewParser()
	if err != nil {
		return nil, err
	}

	return &ModuleAnalyzer{
		parser:          p,
		exportedSymbols: make(map[string]bool),
		version:         "0.0.1", // Default version
	}, nil
}

// AnalyzeModule analyzes a Perl module and returns type information
func (ma *ModuleAnalyzer) AnalyzeModule(modulePath string) (*typedef.TypeDefinition, error) {
	ma.modulePath = modulePath
	ma.moduleName = extractModuleNameFromPath(modulePath)

	// Reset state for new analysis
	ma.types = []typedef.TypeInfo{}
	ma.packages = []typedef.PackageInfo{}
	ma.subs = []typedef.SubInfo{}
	ma.methods = []typedef.MethodInfo{}
	ma.exportedSymbols = make(map[string]bool)
	ma.currentPackage = ""

	// Check if module file exists
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		return nil, errors.NewUserInputError("PVI", "301",
			fmt.Sprintf("Module file not found: %s", modulePath), err)
	}

	// Parse the module
	ast, err := ma.parser.ParseFile(modulePath)
	if err != nil {
		return nil, errors.NewSystemError("302",
			fmt.Sprintf("Failed to parse module %s", modulePath), err)
	}

	// Extract type information from the AST
	if err := ma.extractTypeInformation(ast); err != nil {
		return nil, err
	}

	// Create the type definition
	typeDef := &typedef.TypeDefinition{
		Module:     ma.moduleName,
		Version:    ma.version,
		Generated:  time.Now(),
		Maintainer: "PVI module analyzer",
		Source:     "analyzed",
		Types:      ma.types,
		Packages:   ma.packages,
		Subs:       ma.subs,
		Methods:    ma.methods,
	}

	return typeDef, nil
}

// extractTypeInformation walks the AST and extracts type information
func (ma *ModuleAnalyzer) extractTypeInformation(ast *ast.AST) error {
	// Since AST node text extraction isn't working as expected,
	// let's parse the source text directly as a fallback
	if ast.Source != "" {
		ma.parseSourceDirectly(ast.Source)
	}

	// Process the root node and its children
	if ast.Root != nil {
		ma.processNode(ast.Root)
	}

	// Process type annotations
	for _, annotation := range ast.TypeAnnotations {
		ma.processTypeAnnotation(annotation)
	}

	// If no packages were found but we have subs/methods, create a default package
	if len(ma.packages) == 0 && (len(ma.subs) > 0 || len(ma.methods) > 0) {
		ma.packages = append(ma.packages, typedef.PackageInfo{
			Name:        ma.moduleName,
			Description: fmt.Sprintf("Main package for %s", ma.moduleName),
			Exports:     ma.buildExportList(),
		})
	}

	// Update existing packages with export information
	for i := range ma.packages {
		ma.packages[i].Exports = ma.buildExportList()
	}

	return nil
}

// processNode recursively processes AST nodes to extract information
func (ma *ModuleAnalyzer) processNode(node ast.Node) {
	if node == nil {
		return
	}

	nodeType := node.Type()
	nodeText := node.Text()

	switch nodeType {
	case "package_statement":
		ma.processPackageStatement(node)
	case "sub_decl", "subroutine_declaration", "sub":
		ma.processSubroutineDeclaration(node)
	case "class_statement":
		ma.processClassStatement(node)
	case "method_statement", "method":
		ma.processMethodStatement(node)
	case "field_statement", "field":
		ma.processFieldStatement(node)
	case "var_decl":
		ma.processVarDecl(node)
	case "version_string":
		ma.processVersionString(nodeText)
	}

	// Recursively process child nodes
	children := node.Children()
	for _, child := range children {
		if child != nil {
			ma.processNode(child)
		}
	}
}

// processPackageStatement extracts package information
func (ma *ModuleAnalyzer) processPackageStatement(node ast.Node) {
	nodeText := node.Text()

	// Extract package name from "package Name;"
	packageRegex := regexp.MustCompile(`package\s+([A-Za-z_][A-Za-z0-9_:]*)\s*;`)
	matches := packageRegex.FindStringSubmatch(nodeText)

	if len(matches) > 1 {
		packageName := matches[1]
		ma.currentPackage = packageName

		// If this is the first package and module name wasn't derived correctly, use this
		if ma.moduleName == "" || strings.HasSuffix(ma.modulePath, ".pl") {
			ma.moduleName = packageName
		}

		// Add to packages list if not already present
		found := false
		for _, pkg := range ma.packages {
			if pkg.Name == packageName {
				found = true
				break
			}
		}

		if !found {
			ma.packages = append(ma.packages, typedef.PackageInfo{
				Name:        packageName,
				Description: fmt.Sprintf("Package %s", packageName),
				Exports:     []typedef.ExportInfo{}, // Will be populated later
			})
		}
	}
}

// processSubroutineDeclaration extracts subroutine information
func (ma *ModuleAnalyzer) processSubroutineDeclaration(node ast.Node) {
	nodeText := node.Text()

	// Extract subroutine name and signature
	subInfo := ma.extractSubroutineInfo(nodeText)
	if subInfo != nil {
		ma.subs = append(ma.subs, *subInfo)

		// Mark as exported if it follows common export patterns
		if ma.isLikelyExported(subInfo.Name) {
			ma.exportedSymbols[subInfo.Name] = true
		}
	}
}

// processClassStatement extracts class information
func (ma *ModuleAnalyzer) processClassStatement(node ast.Node) {
	nodeText := node.Text()

	// Extract class name from "class Name { ... }"
	classRegex := regexp.MustCompile(`class\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?:<([^>]+)>)?\s*(?::isa\(([^)]+)\))?\s*{`)
	matches := classRegex.FindStringSubmatch(nodeText)

	if len(matches) > 1 {
		className := matches[1]
		var parent string
		if len(matches) > 3 && matches[3] != "" {
			parent = strings.TrimSpace(matches[3])
		}

		classInfo := typedef.TypeInfo{
			Name:        className,
			Description: fmt.Sprintf("Class %s", className),
			Kind:        "class",
			Parameters:  []typedef.ParamInfo{},
			Properties:  []typedef.PropInfo{},
			Methods:     []typedef.MethodInfo{},
			Parent:      parent,
			Roles:       []string{},
		}

		ma.types = append(ma.types, classInfo)
		ma.currentPackage = className // Classes create their own namespace
	}
}

// processMethodStatement extracts method information
func (ma *ModuleAnalyzer) processMethodStatement(node ast.Node) {
	nodeText := node.Text()

	// Extract method information
	methodInfo := ma.extractMethodInfo(nodeText)
	if methodInfo != nil {
		ma.methods = append(ma.methods, *methodInfo)
	}
}

// processFieldStatement extracts field information
func (ma *ModuleAnalyzer) processFieldStatement(node ast.Node) {
	nodeText := node.Text()

	// Extract field information from "field Type $name;"
	fieldRegex := regexp.MustCompile(`field\s+(?:([A-Za-z_][A-Za-z0-9_:\[\]|]*)\s+)?\$([A-Za-z_][A-Za-z0-9_]*)\s*(?:=\s*([^;]+))?\s*;`)
	matches := fieldRegex.FindStringSubmatch(nodeText)

	if len(matches) > 2 {
		fieldType := "Any"
		if matches[1] != "" {
			fieldType = matches[1]
		}
		fieldName := matches[2]
		defaultValue := ""
		if len(matches) > 3 && matches[3] != "" {
			defaultValue = strings.TrimSpace(matches[3])
		}

		// Add to current class if we're in one
		if len(ma.types) > 0 {
			lastType := &ma.types[len(ma.types)-1]
			if lastType.Kind == "class" {
				lastType.Properties = append(lastType.Properties, typedef.PropInfo{
					Name:        fieldName,
					Type:        fieldType,
					Description: fmt.Sprintf("Field %s", fieldName),
					Optional:    false,
					Default:     defaultValue,
					ReadOnly:    false,
				})
			}
		}
	}
}

// processVarDecl extracts variable declarations including our variables
func (ma *ModuleAnalyzer) processVarDecl(node ast.Node) {
	nodeText := node.Text()

	// Look for "our @EXPORT = ..." or similar export declarations
	if strings.Contains(nodeText, "@EXPORT") {
		ma.extractExportList(nodeText)
	}

	// Look for version declarations
	if strings.Contains(nodeText, "$VERSION") {
		ma.processVersionString(nodeText)
	}
}

// processVersionString extracts version information
func (ma *ModuleAnalyzer) processVersionString(nodeText string) {
	// Extract version from various patterns
	versionRegex := regexp.MustCompile(`(?:our\s+)?\$VERSION\s*=\s*['"]([\d.]+)['"]`)
	matches := versionRegex.FindStringSubmatch(nodeText)

	if len(matches) > 1 {
		ma.version = matches[1]
	}
}

// processTypeAnnotation processes type annotations to extract additional type information
func (ma *ModuleAnalyzer) processTypeAnnotation(annotation *ast.TypeAnnotation) {
	if annotation == nil {
		return
	}

	// Extract additional type information from annotations
	// This could include complex type definitions, constraints, etc.

	switch annotation.Kind {
	case ast.VarAnnotation:
		// Variable type annotations
		ma.processVariableTypeAnnotation(annotation)
	case ast.SubReturnAnnotation:
		// Subroutine return type annotations
		ma.processSubReturnAnnotation(annotation)
	case ast.MethodReturnAnnotation:
		// Method return type annotations
		ma.processMethodReturnAnnotation(annotation)
	case ast.TypeDeclAnnotation:
		// Type declaration annotations
		ma.processTypeDeclaration(annotation)
	}
}

// extractSubroutineInfo extracts subroutine information from text
func (ma *ModuleAnalyzer) extractSubroutineInfo(text string) *typedef.SubInfo {
	// Match various subroutine patterns
	patterns := []string{
		`sub\s+([A-Za-z_][A-Za-z0-9_:\[\]|]*)\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]*)\)\s*{`, // New prefix syntax: sub ReturnType name(params)
		`sub\s+([A-Za-z_][A-Za-z0-9_]*)\s*{`,                                                // Simple: sub name
		`sub\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]*)\)\s*{`,                                  // Typed params: sub name(params)
	}

	for _, pattern := range patterns {
		subRegex := regexp.MustCompile(pattern)
		matches := subRegex.FindStringSubmatch(text)

		if len(matches) > 1 {
			// Handle different patterns based on the matched regex
			var subName string
			parameters := []typedef.ParamInfo{}
			returns := []typedef.ReturnInfo{}

			if pattern == patterns[0] { // New prefix syntax: sub ReturnType name(params)
				// Group 1: return type, Group 2: name, Group 3: parameters
				if len(matches) > 2 {
					subName = matches[2]
				}
				if len(matches) > 3 && matches[3] != "" {
					parameters = ma.parseParameters(matches[3])
				}
				if len(matches) > 1 && matches[1] != "" {
					returns = append(returns, typedef.ReturnInfo{
						Type:        matches[1],
						Description: fmt.Sprintf("Return value of %s", subName),
					})
				}
			} else {
				// Other patterns: Group 1: name, Group 2: parameters (if present)
				subName = matches[1]
				if len(matches) > 2 && matches[2] != "" {
					parameters = ma.parseParameters(matches[2])
				}
			}

			return &typedef.SubInfo{
				Name:        subName,
				Description: fmt.Sprintf("Subroutine %s", subName),
				Parameters:  parameters,
				Returns:     returns,
				Throws:      []string{},
				IsMethod:    false,
				IsPrivate:   ma.isPrivateSymbol(subName),
			}
		}
	}

	return nil
}

// extractMethodInfo extracts method information from text
func (ma *ModuleAnalyzer) extractMethodInfo(text string) *typedef.MethodInfo {
	// Match method patterns - try new prefix syntax first, then old syntax
	patterns := []string{
		`method\s+([A-Za-z_][A-Za-z0-9_:\[\]|]*)\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]*)\)\s*{`,                 // New prefix syntax: method ReturnType name(params)
		`method\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]*)\)\s*(?:returns?\s+([A-Za-z_][A-Za-z0-9_:\[\]|]*))?\s*{`, // Old syntax: method name(params) returns ReturnType
		`method\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]*)\)\s*{`,                                                  // Simple: method name(params)
	}

	for _, pattern := range patterns {
		methodRegex := regexp.MustCompile(pattern)
		matches := methodRegex.FindStringSubmatch(text)

		if len(matches) > 1 {
			var methodName string
			parameters := []typedef.ParamInfo{}
			returns := []typedef.ReturnInfo{}

			if pattern == patterns[0] { // New prefix syntax: method ReturnType name(params)
				// Group 1: return type, Group 2: name, Group 3: parameters
				if len(matches) > 2 {
					methodName = matches[2]
				}
				if len(matches) > 3 && matches[3] != "" {
					parameters = ma.parseParameters(matches[3])
				}
				if len(matches) > 1 && matches[1] != "" {
					returns = append(returns, typedef.ReturnInfo{
						Type:        matches[1],
						Description: fmt.Sprintf("Return value of %s", methodName),
					})
				}
			} else {
				// Old syntax patterns: Group 1: name, Group 2: parameters, Group 3: return type (if present)
				methodName = matches[1]
				if len(matches) > 2 && matches[2] != "" {
					parameters = ma.parseParameters(matches[2])
				}
				if len(matches) > 3 && matches[3] != "" {
					returns = append(returns, typedef.ReturnInfo{
						Type:        matches[3],
						Description: fmt.Sprintf("Return value of %s", methodName),
					})
				}
			}

			return &typedef.MethodInfo{
				Name:        methodName,
				Description: fmt.Sprintf("Method %s", methodName),
				Parameters:  parameters,
				Returns:     returns,
				Throws:      []string{},
				IsPrivate:   ma.isPrivateSymbol(methodName),
				IsStatic:    false,
			}
		}
	}

	return nil
}

// parseParameters parses parameter list from method/subroutine signature
func (ma *ModuleAnalyzer) parseParameters(paramText string) []typedef.ParamInfo {
	parameters := []typedef.ParamInfo{}

	// Split by comma and process each parameter
	paramParts := strings.Split(paramText, ",")
	for _, paramPart := range paramParts {
		paramPart = strings.TrimSpace(paramPart)
		if paramPart == "" {
			continue
		}

		// Match typed parameter: "Type $name" or just "$name"
		paramRegex := regexp.MustCompile(`(?:([A-Za-z_][A-Za-z0-9_:\[\]|]*)\s+)?\$([A-Za-z_][A-Za-z0-9_]*)\s*(?:=\s*([^,]+))?`)
		matches := paramRegex.FindStringSubmatch(paramPart)

		if len(matches) > 2 {
			paramType := "Any"
			if matches[1] != "" {
				paramType = matches[1]
			}
			paramName := matches[2]
			defaultValue := ""
			optional := false

			if len(matches) > 3 && matches[3] != "" {
				defaultValue = strings.TrimSpace(matches[3])
				optional = true
			}

			parameters = append(parameters, typedef.ParamInfo{
				Name:        paramName,
				Type:        paramType,
				Description: fmt.Sprintf("Parameter %s", paramName),
				Optional:    optional,
				Default:     defaultValue,
			})
		}
	}

	return parameters
}

// Helper functions for analysis

func (ma *ModuleAnalyzer) isPrivateSymbol(name string) bool {
	return strings.HasPrefix(name, "_")
}

func (ma *ModuleAnalyzer) isLikelyExported(name string) bool {
	// Common patterns for exported symbols
	return !ma.isPrivateSymbol(name) &&
		!strings.Contains(name, "_internal") &&
		!strings.Contains(name, "_private")
}

func (ma *ModuleAnalyzer) extractExportList(text string) {
	// Extract symbols from @EXPORT declarations
	exportRegex := regexp.MustCompile(`@EXPORT\s*=\s*\(\s*([^)]+)\s*\)`)
	matches := exportRegex.FindStringSubmatch(text)

	if len(matches) > 1 {
		exportList := matches[1]
		// Split by comma and extract symbol names
		symbols := strings.Split(exportList, ",")
		for _, symbol := range symbols {
			symbol = strings.TrimSpace(symbol)
			symbol = strings.Trim(symbol, `'"`)
			if symbol != "" {
				ma.exportedSymbols[symbol] = true
			}
		}
	}
}

func (ma *ModuleAnalyzer) buildExportList() []typedef.ExportInfo {
	exports := []typedef.ExportInfo{}

	for symbol := range ma.exportedSymbols {
		exports = append(exports, typedef.ExportInfo{
			Name:        symbol,
			Type:        "subroutine", // Default type
			Description: fmt.Sprintf("Exported symbol %s", symbol),
		})
	}

	return exports
}

// Type annotation processing implementations
func (ma *ModuleAnalyzer) processVariableTypeAnnotation(annotation *ast.TypeAnnotation) {
	if annotation == nil || annotation.TypeExpression == nil {
		return
	}

	// Extract variable type information and validate it
	varName := annotation.AnnotatedItem
	if varName == "" {
		return
	}

	// Store variable type information for enhanced analysis and validation
	// This can be used for type checking or IDE features in the future
	// For now, we validate that the type expression is well-formed
	if err := ma.validateTypeExpression(annotation.TypeExpression); err != nil {
		// Type expression is malformed, but we continue processing
		return
	}

	// Variable type annotations enhance our understanding of module structure
	// though they don't directly create type definitions in the current implementation
}

func (ma *ModuleAnalyzer) processSubReturnAnnotation(annotation *ast.TypeAnnotation) {
	if annotation == nil || annotation.TypeExpression == nil {
		return
	}

	subName := annotation.AnnotatedItem
	if subName == "" {
		return
	}

	returnType := ma.extractTypeString(annotation.TypeExpression)

	// Use helper method to update return type information
	ma.updateSubReturnType(subName, returnType)
}

func (ma *ModuleAnalyzer) processMethodReturnAnnotation(annotation *ast.TypeAnnotation) {
	if annotation == nil || annotation.TypeExpression == nil {
		return
	}

	methodName := annotation.AnnotatedItem
	if methodName == "" {
		return
	}

	returnType := ma.extractTypeString(annotation.TypeExpression)

	// Use helper method to update return type information
	ma.updateMethodReturnType(methodName, returnType)
}

func (ma *ModuleAnalyzer) processTypeDeclaration(annotation *ast.TypeAnnotation) {
	if annotation == nil || annotation.TypeExpression == nil {
		return
	}

	typeName := annotation.AnnotatedItem
	if typeName == "" {
		return
	}

	// Validate the type expression before processing
	if err := ma.validateTypeExpression(annotation.TypeExpression); err != nil {
		// Type expression is malformed, skip processing
		return
	}

	// Check if type already exists
	if ma.typeExists(typeName) {
		return
	}

	// Create a new type definition based on the annotation
	typeInfo := typedef.TypeInfo{
		Name:        typeName,
		Description: fmt.Sprintf("Type %s", typeName),
		Kind:        ma.determineTypeKind(annotation.TypeExpression),
		Parameters:  make([]typedef.ParamInfo, 0),
		Properties:  make([]typedef.PropInfo, 0),
		Methods:     make([]typedef.MethodInfo, 0),
		Parent:      "",
		Roles:       make([]string, 0),
	}

	// Handle different types of type declarations with proper validation
	switch {
	case annotation.TypeExpression.IsUnion:
		typeInfo.Kind = "union"
		typeInfo.Description = fmt.Sprintf("Union type %s", typeName)
	case annotation.TypeExpression.IsIntersection:
		typeInfo.Kind = "intersection"
		typeInfo.Description = fmt.Sprintf("Intersection type %s", typeName)
	case len(annotation.TypeExpression.Parameters) > 0:
		typeInfo.Kind = "parameterized"
		typeInfo.Description = fmt.Sprintf("Parameterized type %s", typeName)

		// Extract parameter information safely
		typeInfo.Parameters = make([]typedef.ParamInfo, 0, len(annotation.TypeExpression.Parameters))
		for _, param := range annotation.TypeExpression.Parameters {
			if param != nil && param.Name != "" {
				typeInfo.Parameters = append(typeInfo.Parameters, typedef.ParamInfo{
					Name:        param.Name,
					Type:        ma.extractTypeString(param),
					Description: fmt.Sprintf("Type parameter %s", param.Name),
					Optional:    false,
					Default:     "",
				})
			}
		}
	}

	ma.types = append(ma.types, typeInfo)
}

// extractTypeString converts a TypeExpression to a string representation
func (ma *ModuleAnalyzer) extractTypeString(typeExpr *ast.TypeExpression) string {
	if typeExpr == nil {
		return "Any"
	}

	// Use the built-in String() method which handles all complex type cases
	return typeExpr.String()
}

// determineTypeKind determines the kind of type based on the TypeExpression
func (ma *ModuleAnalyzer) determineTypeKind(typeExpr *ast.TypeExpression) string {
	if typeExpr == nil {
		return "scalar"
	}

	if typeExpr.IsUnion {
		return "union"
	}
	if typeExpr.IsIntersection {
		return "intersection"
	}
	if typeExpr.IsNegation {
		return "negation"
	}
	if len(typeExpr.Parameters) > 0 {
		return "parameterized"
	}

	// Default to scalar for simple types
	return "scalar"
}

// validateTypeExpression validates that a TypeExpression is well-formed
func (ma *ModuleAnalyzer) validateTypeExpression(typeExpr *ast.TypeExpression) error {
	if typeExpr == nil {
		return fmt.Errorf("type expression is nil")
	}

	// Validate that union types have at least 2 members
	if typeExpr.IsUnion && len(typeExpr.UnionTypes) < 2 {
		return fmt.Errorf("union type must have at least 2 members")
	}

	// Validate that intersection types have at least 2 members
	if typeExpr.IsIntersection && len(typeExpr.IntersectionTypes) < 2 {
		return fmt.Errorf("intersection type must have at least 2 members")
	}

	// Validate that negation types have a target
	if typeExpr.IsNegation && typeExpr.NegatedType == nil {
		return fmt.Errorf("negation type must have a target type")
	}

	// Validate parameterized types
	if len(typeExpr.Parameters) > 0 {
		for _, param := range typeExpr.Parameters {
			if param == nil {
				return fmt.Errorf("parameterized type contains nil parameter")
			}
			if param.Name == "" {
				return fmt.Errorf("parameterized type contains parameter with empty name")
			}
		}
	}

	return nil
}

// updateSubReturnType updates the return type for a subroutine
func (ma *ModuleAnalyzer) updateSubReturnType(subName, returnType string) {
	for i := range ma.subs {
		if ma.subs[i].Name == subName {
			if len(ma.subs[i].Returns) == 0 {
				ma.subs[i].Returns = []typedef.ReturnInfo{
					{
						Type:        returnType,
						Description: fmt.Sprintf("Return value of %s", subName),
					},
				}
			} else {
				// Update existing return type
				ma.subs[i].Returns[0].Type = returnType
			}
			return
		}
	}
}

// updateMethodReturnType updates the return type for a method
func (ma *ModuleAnalyzer) updateMethodReturnType(methodName, returnType string) {
	for i := range ma.methods {
		if ma.methods[i].Name == methodName {
			if len(ma.methods[i].Returns) == 0 {
				ma.methods[i].Returns = []typedef.ReturnInfo{
					{
						Type:        returnType,
						Description: fmt.Sprintf("Return value of %s", methodName),
					},
				}
			} else {
				// Update existing return type
				ma.methods[i].Returns[0].Type = returnType
			}
			return
		}
	}
}

// typeExists checks if a type with the given name already exists
func (ma *ModuleAnalyzer) typeExists(typeName string) bool {
	for _, existingType := range ma.types {
		if existingType.Name == typeName {
			return true
		}
	}
	return false
}

// extractModuleNameFromPath extracts module name from file path
func extractModuleNameFromPath(path string) string {
	parts := strings.Split(path, string(filepath.Separator))
	if len(parts) == 0 {
		return ""
	}

	filename := parts[len(parts)-1]
	moduleName := strings.TrimSuffix(filename, ".pm")
	moduleName = strings.TrimSuffix(moduleName, ".pl")

	// Handle lib/Module/Name.pm style paths
	if len(parts) >= 3 {
		libIndex := -1
		for i, part := range parts {
			if part == "lib" {
				libIndex = i
				break
			}
		}

		if libIndex >= 0 && libIndex < len(parts)-1 {
			// Reconstruct the module name from the path components after "lib"
			moduleComponents := parts[libIndex+1 : len(parts)-1]
			moduleComponents = append(moduleComponents, moduleName)
			moduleName = strings.Join(moduleComponents, "::")
		}
	}

	return moduleName
}

// parseSourceDirectly parses source text using regex patterns as a fallback
func (ma *ModuleAnalyzer) parseSourceDirectly(source string) {
	lines := strings.Split(source, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Extract package declaration
		if strings.HasPrefix(line, "package ") {
			ma.extractPackageFromLine(line)
		}

		// Extract version
		if strings.Contains(line, "$VERSION") {
			ma.extractVersionFromLine(line)
		}

		// Extract subroutine declarations
		if strings.HasPrefix(line, "sub ") {
			ma.extractSubFromLine(line)
		}

		// Extract class declarations
		if strings.HasPrefix(line, "class ") {
			ma.extractClassFromLine(line)
		}

		// Extract method declarations
		if strings.HasPrefix(line, "method ") {
			ma.extractMethodFromLine(line)
		}

		// Extract field declarations
		if strings.HasPrefix(line, "field ") {
			ma.extractFieldFromLine(line)
		}

		// Extract export declarations
		if strings.Contains(line, "@EXPORT") {
			ma.extractExportsFromLine(line)
		}
	}
}

// Source parsing helper methods

func (ma *ModuleAnalyzer) extractPackageFromLine(line string) {
	packageRegex := regexp.MustCompile(`package\s+([A-Za-z_][A-Za-z0-9_:]*)\s*;`)
	matches := packageRegex.FindStringSubmatch(line)

	if len(matches) > 1 {
		packageName := matches[1]
		ma.currentPackage = packageName

		if ma.moduleName == "" || strings.HasSuffix(ma.modulePath, ".pl") {
			ma.moduleName = packageName
		}

		found := false
		for _, pkg := range ma.packages {
			if pkg.Name == packageName {
				found = true
				break
			}
		}

		if !found {
			ma.packages = append(ma.packages, typedef.PackageInfo{
				Name:        packageName,
				Description: fmt.Sprintf("Package %s", packageName),
				Exports:     []typedef.ExportInfo{},
			})
		}
	}
}

func (ma *ModuleAnalyzer) extractVersionFromLine(line string) {
	versionRegex := regexp.MustCompile(`(?:our\s+)?\$VERSION\s*=\s*['"]([\d.]+)['"]`)
	matches := versionRegex.FindStringSubmatch(line)

	if len(matches) > 1 {
		ma.version = matches[1]
	}
}

func (ma *ModuleAnalyzer) extractSubFromLine(line string) {
	// Try new prefix syntax first: sub ReturnType name(params)
	prefixRegex := regexp.MustCompile(`sub\s+([A-Za-z_][A-Za-z0-9_:\[\]|]*)\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?:\([^)]*\))?\s*{?`)
	matches := prefixRegex.FindStringSubmatch(line)

	var subName string
	var returns []typedef.ReturnInfo

	if len(matches) > 2 {
		// New prefix syntax found
		returnType := matches[1]
		subName = matches[2]
		returns = append(returns, typedef.ReturnInfo{
			Type:        returnType,
			Description: fmt.Sprintf("Return value of %s", subName),
		})
	} else {
		// Fall back to old syntax: sub name(params)
		oldRegex := regexp.MustCompile(`sub\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?:\([^)]*\))?\s*{?`)
		oldMatches := oldRegex.FindStringSubmatch(line)
		if len(oldMatches) > 1 {
			subName = oldMatches[1]
		}
	}

	if subName != "" {
		subInfo := typedef.SubInfo{
			Name:        subName,
			Description: fmt.Sprintf("Subroutine %s", subName),
			Parameters:  []typedef.ParamInfo{},
			Returns:     returns,
			Throws:      []string{},
			IsMethod:    false,
			IsPrivate:   ma.isPrivateSymbol(subName),
		}

		ma.subs = append(ma.subs, subInfo)

		if ma.isLikelyExported(subName) {
			ma.exportedSymbols[subName] = true
		}
	}
}

func (ma *ModuleAnalyzer) extractClassFromLine(line string) {
	classRegex := regexp.MustCompile(`class\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?:<([^>]+)>)?\s*(?::isa\(([^)]+)\))?\s*{?`)
	matches := classRegex.FindStringSubmatch(line)

	if len(matches) > 1 {
		className := matches[1]
		var parent string
		if len(matches) > 3 && matches[3] != "" {
			parent = strings.TrimSpace(matches[3])
		}

		classInfo := typedef.TypeInfo{
			Name:        className,
			Description: fmt.Sprintf("Class %s", className),
			Kind:        "class",
			Parameters:  []typedef.ParamInfo{},
			Properties:  []typedef.PropInfo{},
			Methods:     []typedef.MethodInfo{},
			Parent:      parent,
			Roles:       []string{},
		}

		ma.types = append(ma.types, classInfo)
		ma.currentPackage = className
	}
}

func (ma *ModuleAnalyzer) extractMethodFromLine(line string) {
	// Try new prefix syntax first: method ReturnType name(params)
	prefixRegex := regexp.MustCompile(`method\s+([A-Za-z_][A-Za-z0-9_:\[\]|]*)\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]*)\)\s*{?`)
	matches := prefixRegex.FindStringSubmatch(line)

	var methodName string
	var parameters []typedef.ParamInfo
	var returns []typedef.ReturnInfo

	if len(matches) > 2 {
		// New prefix syntax found
		returnType := matches[1]
		methodName = matches[2]
		if len(matches) > 3 && matches[3] != "" {
			parameters = ma.parseParametersFromString(matches[3])
		}
		returns = append(returns, typedef.ReturnInfo{
			Type:        returnType,
			Description: fmt.Sprintf("Return value of %s", methodName),
		})
	} else {
		// Fall back to old syntax: method name(params) returns ReturnType
		oldRegex := regexp.MustCompile(`method\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]*)\)\s*(?:returns?\s+([A-Za-z_][A-Za-z0-9_:\[\]|]*))?\s*{?`)
		oldMatches := oldRegex.FindStringSubmatch(line)

		if len(oldMatches) > 1 {
			methodName = oldMatches[1]
			if len(oldMatches) > 2 && oldMatches[2] != "" {
				parameters = ma.parseParametersFromString(oldMatches[2])
			}
			if len(oldMatches) > 3 && oldMatches[3] != "" {
				returns = append(returns, typedef.ReturnInfo{
					Type:        oldMatches[3],
					Description: fmt.Sprintf("Return value of %s", methodName),
				})
			}
		}
	}

	if methodName != "" {
		methodInfo := typedef.MethodInfo{
			Name:        methodName,
			Description: fmt.Sprintf("Method %s", methodName),
			Parameters:  parameters,
			Returns:     returns,
			Throws:      []string{},
			IsPrivate:   ma.isPrivateSymbol(methodName),
			IsStatic:    false,
		}

		ma.methods = append(ma.methods, methodInfo)
	}
}

func (ma *ModuleAnalyzer) extractFieldFromLine(line string) {
	fieldRegex := regexp.MustCompile(`field\s+(?:([A-Za-z_][A-Za-z0-9_:\[\]|]*)\s+)?\$([A-Za-z_][A-Za-z0-9_]*)\s*(?:=\s*([^;]+))?\s*;?`)
	matches := fieldRegex.FindStringSubmatch(line)

	if len(matches) > 2 {
		fieldType := "Any"
		if matches[1] != "" {
			fieldType = matches[1]
		}
		fieldName := matches[2]
		defaultValue := ""
		if len(matches) > 3 && matches[3] != "" {
			defaultValue = strings.TrimSpace(matches[3])
		}

		if len(ma.types) > 0 {
			lastType := &ma.types[len(ma.types)-1]
			if lastType.Kind == "class" {
				lastType.Properties = append(lastType.Properties, typedef.PropInfo{
					Name:        fieldName,
					Type:        fieldType,
					Description: fmt.Sprintf("Field %s", fieldName),
					Optional:    false,
					Default:     defaultValue,
					ReadOnly:    false,
				})
			}
		}
	}
}

func (ma *ModuleAnalyzer) extractExportsFromLine(line string) {
	// Look for @EXPORT declarations with qw()
	exportRegex := regexp.MustCompile(`@EXPORT\s*=\s*qw\(([^)]+)\)`)
	matches := exportRegex.FindStringSubmatch(line)

	if len(matches) > 1 {
		exportList := matches[1]
		symbols := strings.Fields(exportList)
		for _, symbol := range symbols {
			symbol = strings.TrimSpace(symbol)
			if symbol != "" {
				ma.exportedSymbols[symbol] = true
			}
		}
		return
	}

	// Also look for regular array syntax: @EXPORT = ('sub1', 'sub2');
	exportRegex2 := regexp.MustCompile(`@EXPORT\s*=\s*\(\s*([^)]+)\s*\)`)
	matches2 := exportRegex2.FindStringSubmatch(line)

	if len(matches2) > 1 {
		exportList := matches2[1]
		symbols := strings.Split(exportList, ",")
		for _, symbol := range symbols {
			symbol = strings.TrimSpace(symbol)
			symbol = strings.Trim(symbol, `'"`)
			if symbol != "" {
				ma.exportedSymbols[symbol] = true
			}
		}
	}
}

func (ma *ModuleAnalyzer) parseParametersFromString(paramText string) []typedef.ParamInfo {
	parameters := []typedef.ParamInfo{}

	paramParts := strings.Split(paramText, ",")
	for _, paramPart := range paramParts {
		paramPart = strings.TrimSpace(paramPart)
		if paramPart == "" {
			continue
		}

		paramRegex := regexp.MustCompile(`(?:([A-Za-z_][A-Za-z0-9_:\[\]|]*)\s+)?\$([A-Za-z_][A-Za-z0-9_]*)\s*(?:=\s*([^,]+))?`)
		matches := paramRegex.FindStringSubmatch(paramPart)

		if len(matches) > 2 {
			paramType := "Any"
			if matches[1] != "" {
				paramType = matches[1]
			}
			paramName := matches[2]
			defaultValue := ""
			optional := false

			if len(matches) > 3 && matches[3] != "" {
				defaultValue = strings.TrimSpace(matches[3])
				optional = true
			}

			parameters = append(parameters, typedef.ParamInfo{
				Name:        paramName,
				Type:        paramType,
				Description: fmt.Sprintf("Parameter %s", paramName),
				Optional:    optional,
				Default:     defaultValue,
			})
		}
	}

	return parameters
}
