// ABOUTME: Enhanced module introspector with advanced type inference
// ABOUTME: Combines AST analysis, POD parsing, and runtime introspection

package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"tamarou.com/pvm/internal/typedef"
)

// EnhancedIntrospector provides comprehensive module analysis
type EnhancedIntrospector struct {
	// ModuleIntrospector for AST-based analysis
	ModuleIntrospector *ModuleIntrospector

	// PODParser for documentation analysis
	PODParser *PODParser

	// RuntimeIntrospector for Perl runtime analysis
	RuntimeIntrospector *RuntimeIntrospector

	// TypeInferenceEngine for advanced type inference
	TypeInferenceEngine *TypeInferenceEngine

	// Cache for introspection results
	cache map[string]*EnhancedIntrospectionResult
}

// RuntimeIntrospector performs runtime analysis of Perl modules
type RuntimeIntrospector struct {
	// PerlPath is the path to the perl interpreter
	PerlPath string

	// IncludePaths are additional paths to include
	IncludePaths []string
}

// TypeInferenceEngine performs advanced type inference
type TypeInferenceEngine struct {
	// TypeDatabase contains known type patterns
	TypeDatabase *TypePatternDatabase

	// InferenceRules contains type inference rules
	InferenceRules []InferenceRule
}

// TypePatternDatabase stores common type patterns
type TypePatternDatabase struct {
	// MethodPatterns maps method names to likely signatures
	MethodPatterns map[string]*MethodPattern

	// VariablePatterns maps variable naming patterns to types
	VariablePatterns map[string]string

	// ReturnPatterns maps return patterns to types
	ReturnPatterns map[string]string
}

// MethodPattern describes a common method pattern
type MethodPattern struct {
	// Pattern is the method name pattern
	Pattern *regexp.Regexp

	// LikelySignature is the likely signature
	LikelySignature *typedef.MethodInfo

	// Category is the method category (accessor, constructor, etc.)
	Category string
}

// InferenceRule represents a type inference rule
type InferenceRule struct {
	// Name is the rule name
	Name string

	// Condition checks if the rule applies
	Condition func(context *InferenceContext) bool

	// Apply applies the inference
	Apply func(context *InferenceContext) *InferredType
}

// InferenceContext provides context for type inference
type InferenceContext struct {
	// VariableName is the variable being analyzed
	VariableName string

	// Usage contains usage patterns
	Usage []string

	// Scope is the variable scope
	Scope string

	// SurroundingCode provides context
	SurroundingCode string
}

// InferredType represents an inferred type
type InferredType struct {
	// Type is the inferred type
	Type string

	// Confidence is the confidence level (0-1)
	Confidence float64

	// Reason explains the inference
	Reason string
}

// EnhancedIntrospectionResult contains comprehensive analysis results
type EnhancedIntrospectionResult struct {
	// ModuleName is the module name
	ModuleName string

	// FilePath is the module file path
	FilePath string

	// TypeDefinition is the generated type definition
	TypeDefinition *typedef.TypeDefinition

	// Confidence scores for different aspects
	Confidence struct {
		Overall    float64
		Methods    float64
		Attributes float64
		Types      float64
		Parameters float64
	}

	// Warnings contains any warnings during analysis
	Warnings []string

	// Errors contains any errors during analysis
	Errors []error
}

// NewEnhancedIntrospector creates a new enhanced introspector
func NewEnhancedIntrospector() (*EnhancedIntrospector, error) {
	moduleIntrospector, err := NewModuleIntrospector()
	if err != nil {
		return nil, err
	}

	introspector := &EnhancedIntrospector{
		ModuleIntrospector:  moduleIntrospector,
		PODParser:           NewPODParser(),
		RuntimeIntrospector: NewRuntimeIntrospector(),
		TypeInferenceEngine: NewTypeInferenceEngine(),
		cache:               make(map[string]*EnhancedIntrospectionResult),
	}

	return introspector, nil
}

// NewRuntimeIntrospector creates a new runtime introspector
func NewRuntimeIntrospector() *RuntimeIntrospector {
	return &RuntimeIntrospector{
		PerlPath:     "perl",
		IncludePaths: []string{},
	}
}

// NewTypeInferenceEngine creates a new type inference engine
func NewTypeInferenceEngine() *TypeInferenceEngine {
	engine := &TypeInferenceEngine{
		TypeDatabase:   NewTypePatternDatabase(),
		InferenceRules: []InferenceRule{},
	}

	// Add inference rules
	engine.addInferenceRules()

	return engine
}

// NewTypePatternDatabase creates a new type pattern database
func NewTypePatternDatabase() *TypePatternDatabase {
	db := &TypePatternDatabase{
		MethodPatterns:   make(map[string]*MethodPattern),
		VariablePatterns: make(map[string]string),
		ReturnPatterns:   make(map[string]string),
	}

	// Add common patterns
	db.addCommonPatterns()

	return db
}

// addCommonPatterns adds common type patterns
func (db *TypePatternDatabase) addCommonPatterns() {
	// Common accessor patterns
	db.MethodPatterns["get_*"] = &MethodPattern{
		Pattern: regexp.MustCompile(`^get_(\w+)$`),
		LikelySignature: &typedef.MethodInfo{
			Name:        "get_*",
			Description: "Getter method",
			Parameters: []typedef.ParamInfo{
				{Name: "$self", Type: "Object"},
			},
			Returns: []typedef.ReturnInfo{
				{Type: "Any", Description: "Property value"},
			},
		},
		Category: "accessor",
	}

	db.MethodPatterns["set_*"] = &MethodPattern{
		Pattern: regexp.MustCompile(`^set_(\w+)$`),
		LikelySignature: &typedef.MethodInfo{
			Name:        "set_*",
			Description: "Setter method",
			Parameters: []typedef.ParamInfo{
				{Name: "$self", Type: "Object"},
				{Name: "$value", Type: "Any"},
			},
			Returns: []typedef.ReturnInfo{
				{Type: "Object", Description: "Self for chaining"},
			},
		},
		Category: "accessor",
	}

	// Common variable patterns
	db.VariablePatterns[`\$\w+_ref$`] = "Ref"
	db.VariablePatterns[`\$\w+_arrayref$`] = "ArrayRef"
	db.VariablePatterns[`\$\w+_hashref$`] = "HashRef"
	db.VariablePatterns[`\$\w+_coderef$`] = "CodeRef"
	db.VariablePatterns[`\$\w+_count$`] = "Int"
	db.VariablePatterns[`\$\w+_size$`] = "Int"
	db.VariablePatterns[`\$\w+_length$`] = "Int"
	db.VariablePatterns[`\$\w+_flag$`] = "Bool"
	db.VariablePatterns[`\$is_\w+$`] = "Bool"
	db.VariablePatterns[`\$has_\w+$`] = "Bool"
	db.VariablePatterns[`\$\w+_name$`] = "Str"
	db.VariablePatterns[`\$\w+_path$`] = "Str"
	db.VariablePatterns[`\$\w+_url$`] = "Str"

	// Common return patterns
	db.ReturnPatterns[`return 1;`] = "Bool"
	db.ReturnPatterns[`return 0;`] = "Bool"
	db.ReturnPatterns[`return undef;`] = "Undef"
	db.ReturnPatterns[`return \$self;`] = "Object"
	db.ReturnPatterns[`return \[\];`] = "ArrayRef"
	db.ReturnPatterns[`return \{\};`] = "HashRef"
}

// addInferenceRules adds type inference rules
func (e *TypeInferenceEngine) addInferenceRules() {
	// Rule: DBI handle inference
	e.InferenceRules = append(e.InferenceRules, InferenceRule{
		Name: "dbi_handle",
		Condition: func(ctx *InferenceContext) bool {
			return strings.Contains(ctx.SurroundingCode, "DBI->connect")
		},
		Apply: func(ctx *InferenceContext) *InferredType {
			return &InferredType{
				Type:       "DBI::db",
				Confidence: 0.95,
				Reason:     "DBI->connect returns a database handle",
			}
		},
	})

	// Rule: File handle inference
	e.InferenceRules = append(e.InferenceRules, InferenceRule{
		Name: "file_handle",
		Condition: func(ctx *InferenceContext) bool {
			return strings.Contains(ctx.SurroundingCode, "open") &&
				strings.Contains(ctx.VariableName, "fh")
		},
		Apply: func(ctx *InferenceContext) *InferredType {
			return &InferredType{
				Type:       "FileHandle",
				Confidence: 0.90,
				Reason:     "Variable used with open() and named like a file handle",
			}
		},
	})

	// Rule: Array operations inference
	e.InferenceRules = append(e.InferenceRules, InferenceRule{
		Name: "array_operations",
		Condition: func(ctx *InferenceContext) bool {
			for _, usage := range ctx.Usage {
				if strings.Contains(usage, "push") || strings.Contains(usage, "pop") ||
					strings.Contains(usage, "shift") || strings.Contains(usage, "unshift") {
					return true
				}
			}
			return false
		},
		Apply: func(ctx *InferenceContext) *InferredType {
			return &InferredType{
				Type:       "ArrayRef",
				Confidence: 0.85,
				Reason:     "Variable used with array operations",
			}
		},
	})
}

// AnalyzeModule performs comprehensive module analysis
func (e *EnhancedIntrospector) AnalyzeModule(moduleName string) (*EnhancedIntrospectionResult, error) {
	// Check cache
	if cached, exists := e.cache[moduleName]; exists {
		return cached, nil
	}

	result := &EnhancedIntrospectionResult{
		ModuleName: moduleName,
		Warnings:   []string{},
		Errors:     []error{},
	}

	// Find module file
	modulePath, err := e.findModuleFile(moduleName)
	if err != nil {
		// Try runtime introspection without file
		return e.analyzeWithRuntimeOnly(moduleName)
	}
	result.FilePath = modulePath

	// Phase 1: AST-based introspection
	astResult, err := e.ModuleIntrospector.IntrospectModule(modulePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("AST introspection failed: %w", err))
	}

	// Phase 2: POD documentation analysis
	podDoc, err := e.PODParser.ParsePODFromFile(modulePath)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("POD parsing failed: %v", err))
	}

	// Phase 3: Runtime introspection
	runtimeResult, err := e.RuntimeIntrospector.IntrospectModule(moduleName)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Runtime introspection failed: %v", err))
	}

	// Phase 4: Merge all results
	typeDef := e.mergeResults(moduleName, astResult, podDoc, runtimeResult)

	// Phase 5: Apply type inference
	e.applyTypeInference(typeDef, astResult)

	// Phase 6: Calculate confidence scores
	e.calculateConfidence(result, typeDef)

	result.TypeDefinition = typeDef

	// Cache the result
	e.cache[moduleName] = result

	return result, nil
}

// findModuleFile finds the file path for a module
func (e *EnhancedIntrospector) findModuleFile(moduleName string) (string, error) {
	// Convert module name to file path
	moduleFile := strings.ReplaceAll(moduleName, "::", "/") + ".pm"

	// Use perl to find the module
	cmd := exec.Command("perl", "-e", fmt.Sprintf(`
		foreach my $inc (@INC) {
			my $file = "$inc/%s";
			if (-f $file) {
				print $file;
				exit 0;
			}
		}
		exit 1;
	`, moduleFile))

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("module not found: %s", moduleName)
	}

	return strings.TrimSpace(string(output)), nil
}

// analyzeWithRuntimeOnly performs runtime-only analysis
func (e *EnhancedIntrospector) analyzeWithRuntimeOnly(moduleName string) (*EnhancedIntrospectionResult, error) {
	result := &EnhancedIntrospectionResult{
		ModuleName: moduleName,
		Warnings:   []string{"No source file found, using runtime introspection only"},
		Errors:     []error{},
	}

	// Perform runtime introspection
	runtimeResult, err := e.RuntimeIntrospector.IntrospectModule(moduleName)
	if err != nil {
		return nil, err
	}

	// Convert to TypeDefinition
	typeDef := e.runtimeResultToTypeDef(moduleName, runtimeResult)
	result.TypeDefinition = typeDef

	// Calculate confidence (lower for runtime-only)
	result.Confidence.Overall = 0.6
	result.Confidence.Methods = 0.5
	result.Confidence.Attributes = 0.4
	result.Confidence.Types = 0.5
	result.Confidence.Parameters = 0.3

	return result, nil
}

// mergeResults merges results from all analysis phases
func (e *EnhancedIntrospector) mergeResults(
	moduleName string,
	astResult *ModuleIntrospectionResult,
	podDoc *PODDocument,
	runtimeResult *RuntimeIntrospectionResult,
) *typedef.TypeDefinition {
	typeDef := &typedef.TypeDefinition{
		Module:     moduleName,
		Version:    "0.0.1",
		Maintainer: "Enhanced PSC type generator",
		Source:     "comprehensive-introspection",
		Types:      []typedef.TypeInfo{},
		Packages:   []typedef.PackageInfo{},
		Subs:       []typedef.SubInfo{},
		Methods:    []typedef.MethodInfo{},
	}

	// Start with runtime result if available
	if runtimeResult != nil {
		// Extract version
		if runtimeResult.Version != "" {
			typeDef.Version = runtimeResult.Version
		}

		// Add packages
		for _, pkg := range runtimeResult.Packages {
			typeDef.Packages = append(typeDef.Packages, typedef.PackageInfo{
				Name:        pkg.Name,
				Description: fmt.Sprintf("Package %s", pkg.Name),
				Exports:     []typedef.ExportInfo{},
			})
		}
	}

	// Merge AST results
	if astResult != nil {
		e.mergeASTResults(typeDef, astResult)
	}

	// Merge POD documentation
	if podDoc != nil {
		e.mergePODResults(typeDef, podDoc)
	}

	// Merge runtime-specific information
	if runtimeResult != nil {
		e.mergeRuntimeSpecifics(typeDef, runtimeResult)
	}

	return typeDef
}

// mergeASTResults merges AST analysis results
func (e *EnhancedIntrospector) mergeASTResults(typeDef *typedef.TypeDefinition, astResult *ModuleIntrospectionResult) {
	for pkgName, pkgInfo := range astResult.Packages {
		// Find or create type info
		var typeInfo *typedef.TypeInfo
		for i := range typeDef.Types {
			if typeDef.Types[i].Name == pkgName {
				typeInfo = &typeDef.Types[i]
				break
			}
		}
		if typeInfo == nil {
			typeInfo = &typedef.TypeInfo{
				Name:        pkgName,
				Description: fmt.Sprintf("Type information for %s", pkgName),
				Kind:        "class",
				Methods:     []typedef.MethodInfo{},
				Properties:  []typedef.PropInfo{},
			}
			typeDef.Types = append(typeDef.Types, *typeInfo)
		}

		// Add methods
		for _, methodSig := range pkgInfo.Methods {
			methodInfo := e.convertMethodSignature(methodSig)
			typeInfo.Methods = append(typeInfo.Methods, methodInfo)
			typeDef.Methods = append(typeDef.Methods, methodInfo)
		}

		// Add attributes
		for attrName, attrInfo := range pkgInfo.Attributes {
			typeInfo.Properties = append(typeInfo.Properties, typedef.PropInfo{
				Name:        attrName,
				Type:        attrInfo.Type,
				Description: attrInfo.Documentation,
			})
		}
	}
}

// mergePODResults merges POD documentation
func (e *EnhancedIntrospector) mergePODResults(typeDef *typedef.TypeDefinition, podDoc *PODDocument) {
	// Update method documentation
	for i := range typeDef.Methods {
		method := &typeDef.Methods[i]
		if podMethod, exists := podDoc.Methods[method.Name]; exists {
			method.Description = podMethod.Description
			if podMethod.ReturnType != "" && len(method.Returns) > 0 {
				method.Returns[0].Type = podMethod.ReturnType
			}
		}
	}

	// Update type documentation
	for i := range typeDef.Types {
		typeInfo := &typeDef.Types[i]
		if podType, exists := podDoc.Types[typeInfo.Name]; exists {
			typeInfo.Description = podType.Definition
		}
	}
}

// mergeRuntimeSpecifics merges runtime-specific information
func (e *EnhancedIntrospector) mergeRuntimeSpecifics(typeDef *typedef.TypeDefinition, runtimeResult *RuntimeIntrospectionResult) {
	// Add dynamically detected methods
	for _, dynMethod := range runtimeResult.DynamicMethods {
		found := false
		for _, method := range typeDef.Methods {
			if method.Name == dynMethod.Name {
				found = true
				break
			}
		}
		if !found {
			typeDef.Methods = append(typeDef.Methods, typedef.MethodInfo{
				Name:        dynMethod.Name,
				Description: fmt.Sprintf("Dynamically generated method %s", dynMethod.Name),
				Parameters:  []typedef.ParamInfo{{Name: "$self", Type: "Object"}},
				Returns:     []typedef.ReturnInfo{{Type: "Any"}},
			})
		}
	}
}

// applyTypeInference applies advanced type inference
func (e *EnhancedIntrospector) applyTypeInference(typeDef *typedef.TypeDefinition, astResult *ModuleIntrospectionResult) {
	// Apply inference to methods
	for i := range typeDef.Methods {
		method := &typeDef.Methods[i]

		// Check method patterns
		for _, pattern := range e.TypeInferenceEngine.TypeDatabase.MethodPatterns {
			if pattern.Pattern.MatchString(method.Name) {
				// Apply pattern-based inference
				if pattern.LikelySignature != nil {
					if len(method.Parameters) == 0 {
						method.Parameters = pattern.LikelySignature.Parameters
					}
					if len(method.Returns) == 0 {
						method.Returns = pattern.LikelySignature.Returns
					}
				}
			}
		}
	}
}

// calculateConfidence calculates confidence scores
func (e *EnhancedIntrospector) calculateConfidence(result *EnhancedIntrospectionResult, typeDef *typedef.TypeDefinition) {
	// Method confidence
	methodsWithTypes := 0
	for _, method := range typeDef.Methods {
		hasTypes := false
		for _, param := range method.Parameters {
			if param.Type != "" && param.Type != "Any" {
				hasTypes = true
				break
			}
		}
		if hasTypes {
			methodsWithTypes++
		}
	}
	if len(typeDef.Methods) > 0 {
		result.Confidence.Methods = float64(methodsWithTypes) / float64(len(typeDef.Methods))
	}

	// Attribute confidence
	attrsWithTypes := 0
	totalAttrs := 0
	for _, typeInfo := range typeDef.Types {
		for _, prop := range typeInfo.Properties {
			totalAttrs++
			if prop.Type != "" && prop.Type != "Any" {
				attrsWithTypes++
			}
		}
	}
	if totalAttrs > 0 {
		result.Confidence.Attributes = float64(attrsWithTypes) / float64(totalAttrs)
	}

	// Overall confidence
	result.Confidence.Overall = (result.Confidence.Methods + result.Confidence.Attributes) / 2
}

// convertMethodSignature converts internal method signature to typedef format
func (e *EnhancedIntrospector) convertMethodSignature(sig *MethodSignature) typedef.MethodInfo {
	// Convert parameters
	params := []typedef.ParamInfo{}
	for _, param := range sig.Parameters {
		params = append(params, typedef.ParamInfo{
			Name:        param.Name,
			Type:        param.Type,
			Description: param.Documentation,
			Optional:    param.IsOptional,
			Default:     param.DefaultValue,
		})
	}

	// Convert return type
	returns := []typedef.ReturnInfo{}
	if sig.ReturnType != "" && sig.ReturnType != "Any" {
		returns = append(returns, typedef.ReturnInfo{
			Type:        sig.ReturnType,
			Description: "Return value",
		})
	}

	return typedef.MethodInfo{
		Name:        sig.Name,
		Description: sig.Documentation,
		Parameters:  params,
		Returns:     returns,
	}
}

// runtimeResultToTypeDef converts runtime result to TypeDefinition
func (e *EnhancedIntrospector) runtimeResultToTypeDef(moduleName string, result *RuntimeIntrospectionResult) *typedef.TypeDefinition {
	typeDef := &typedef.TypeDefinition{
		Module:     moduleName,
		Version:    result.Version,
		Maintainer: "PSC runtime introspector",
		Source:     "runtime-introspection",
		Types:      []typedef.TypeInfo{},
		Packages:   []typedef.PackageInfo{},
		Subs:       []typedef.SubInfo{},
		Methods:    []typedef.MethodInfo{},
	}

	// Convert packages
	for _, pkg := range result.Packages {
		typeDef.Packages = append(typeDef.Packages, typedef.PackageInfo{
			Name:        pkg.Name,
			Description: fmt.Sprintf("Package %s", pkg.Name),
			Exports:     []typedef.ExportInfo{},
		})

		// Create type info
		typeInfo := typedef.TypeInfo{
			Name:        pkg.Name,
			Description: fmt.Sprintf("Runtime-detected type for %s", pkg.Name),
			Kind:        "class",
			Methods:     []typedef.MethodInfo{},
			Properties:  []typedef.PropInfo{},
		}

		// Add methods
		for _, method := range pkg.Methods {
			methodInfo := typedef.MethodInfo{
				Name:        method.Name,
				Description: fmt.Sprintf("Method %s", method.Name),
				Parameters:  []typedef.ParamInfo{{Name: "$self", Type: "Object"}},
				Returns:     []typedef.ReturnInfo{{Type: "Any"}},
			}
			typeInfo.Methods = append(typeInfo.Methods, methodInfo)
			typeDef.Methods = append(typeDef.Methods, methodInfo)
		}

		typeDef.Types = append(typeDef.Types, typeInfo)
	}

	return typeDef
}

// RuntimeIntrospectionResult contains runtime introspection results
type RuntimeIntrospectionResult struct {
	// Version is the module version
	Version string

	// Packages contains package information
	Packages []*RuntimePackageInfo

	// DynamicMethods contains dynamically generated methods
	DynamicMethods []*DynamicMethodInfo

	// Exports contains exported symbols
	Exports []string
}

// RuntimePackageInfo contains runtime package information
type RuntimePackageInfo struct {
	// Name is the package name
	Name string

	// Methods contains method names
	Methods []*RuntimeMethodInfo

	// ISA contains parent classes
	ISA []string
}

// RuntimeMethodInfo contains runtime method information
type RuntimeMethodInfo struct {
	// Name is the method name
	Name string

	// Source indicates where the method comes from
	Source string
}

// DynamicMethodInfo contains information about dynamic methods
type DynamicMethodInfo struct {
	// Name is the method name
	Name string

	// Generator is what generated the method
	Generator string
}

// IntrospectModule performs runtime introspection of a module
func (r *RuntimeIntrospector) IntrospectModule(moduleName string) (*RuntimeIntrospectionResult, error) {
	// Create a Perl script for runtime introspection
	script := r.createIntrospectionScript(moduleName)

	// Execute the script
	cmd := exec.Command(r.PerlPath, "-e", script)

	// Add include paths
	for _, inc := range r.IncludePaths {
		cmd.Args = append(cmd.Args, "-I", inc)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("runtime introspection failed: %v\nstderr: %s", err, stderr.String())
	}

	// Parse the JSON output
	var result RuntimeIntrospectionResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse introspection output: %w", err)
	}

	return &result, nil
}

// createIntrospectionScript creates the Perl introspection script
func (r *RuntimeIntrospector) createIntrospectionScript(moduleName string) string {
	return fmt.Sprintf(`
use strict;
use warnings;
use JSON;
use Module::Load;

my $module = '%s';
my $result = {
    version => undef,
    packages => [],
    dynamic_methods => [],
    exports => []
};

# Try to load the module
eval { load $module };
if ($@) {
    die "Failed to load module: $@";
}

# Get version
eval { $result->{version} = $module->VERSION };

# Analyze the module namespace
analyze_namespace($module, $result);

# Output JSON
print encode_json($result);

sub analyze_namespace {
    my ($namespace, $result) = @_;

    my $pkg_info = {
        name => $namespace,
        methods => [],
        isa => []
    };

    # Get ISA
    {
        no strict 'refs';
        @{$pkg_info->{isa}} = @{"${namespace}::ISA"};
    }

    # Get methods
    {
        no strict 'refs';
        my $stash = \%%{"${namespace}::"};

        for my $symbol (keys %%$stash) {
            next if $symbol =~ /^[A-Z_]+$/; # Skip constants
            next if $symbol =~ /^(BEGIN|CHECK|INIT|END|DESTROY)$/;

            my $fullname = "${namespace}::${symbol}";
            if (defined &$fullname) {
                push @{$pkg_info->{methods}}, {
                    name => $symbol,
                    source => 'defined'
                };
            }
        }
    }

    # Check for AUTOLOAD
    {
        no strict 'refs';
        if (defined &{"${namespace}::AUTOLOAD"}) {
            push @{$result->{dynamic_methods}}, {
                name => 'AUTOLOAD',
                generator => 'AUTOLOAD'
            };
        }
    }

    # Check for common accessor generators
    if ($namespace->can('mk_accessors')) {
        push @{$result->{dynamic_methods}}, {
            name => 'accessors',
            generator => 'mk_accessors'
        };
    }

    push @{$result->{packages}}, $pkg_info;
}
`, moduleName)
}
