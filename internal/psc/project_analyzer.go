// ABOUTME: Project-wide analysis for dependency graphs and type conflicts
// ABOUTME: Provides comprehensive analysis across multiple Perl modules

package psc

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typechecker"
	"tamarou.com/pvm/internal/typedef"
)

// ProjectAnalyzer performs comprehensive analysis across a Perl project
type ProjectAnalyzer struct {
	// RootDir is the project root directory
	RootDir string

	// DependencyGraph represents module dependencies
	DependencyGraph *DependencyGraph

	// TypeConflictDetector detects type conflicts across modules
	TypeConflictDetector *TypeConflictDetector

	// ModuleAnalyzer analyzes individual modules
	ModuleAnalyzer *ModuleAnalyzer

	// IncrementalCache provides incremental analysis caching
	IncrementalCache *IncrementalCache

	// TypeChecker for type checking
	TypeChecker *typechecker.TypeCheck

	// PackageManager for external dependency analysis
	PackageManager PackageManager

	// Config contains analyzer configuration
	Config *AnalyzerConfig

	// Results accumulates analysis results
	Results *ProjectAnalysisResult
}

// AnalyzerConfig contains configuration for the project analyzer
type AnalyzerConfig struct {
	// IncludePaths are additional paths to search for modules
	IncludePaths []string

	// ExcludePatterns are patterns for files/dirs to exclude
	ExcludePatterns []string

	// MaxDepth is the maximum dependency depth to analyze
	MaxDepth int

	// EnableCaching enables incremental caching
	EnableCaching bool

	// CacheDir is the directory for cache files
	CacheDir string

	// Parallel enables parallel analysis
	Parallel bool

	// MaxWorkers is the maximum number of parallel workers
	MaxWorkers int
}

// DependencyGraph represents the dependency relationships between modules
type DependencyGraph struct {
	// Nodes are the modules in the graph
	Nodes map[string]*ModuleNode

	// Edges represent dependencies
	Edges map[string][]*DependencyEdge

	// mutex for thread-safe operations
	mu sync.RWMutex
}

// ModuleNode represents a module in the dependency graph
type ModuleNode struct {
	// Name is the module name
	Name string

	// Path is the file path
	Path string

	// Version is the module version
	Version string

	// Dependencies are direct dependencies
	Dependencies []string

	// Dependents are modules that depend on this one
	Dependents []string

	// Analyzed indicates if the module has been analyzed
	Analyzed bool

	// LastModified is the last modification time
	LastModified time.Time
}

// DependencyEdge represents a dependency relationship
type DependencyEdge struct {
	// From is the depending module
	From string

	// To is the depended-upon module
	To string

	// Type is the dependency type (use, require, etc.)
	Type string

	// Version is the required version
	Version string
}

// TypeConflictDetector detects type conflicts across modules
type TypeConflictDetector struct {
	// Conflicts stores detected conflicts
	Conflicts []TypeConflict

	// TypeRegistry tracks types across modules
	TypeRegistry map[string]map[string]*typedef.TypeInfo

	// mutex for thread-safe operations
	mu sync.RWMutex
}

// TypeConflict represents a type conflict between modules
type TypeConflict struct {
	// TypeName is the conflicting type name
	TypeName string

	// Module1 is the first module
	Module1 string

	// Type1 is the type definition in module 1
	Type1 *typedef.TypeInfo

	// Module2 is the second module
	Module2 string

	// Type2 is the type definition in module 2
	Type2 *typedef.TypeInfo

	// Severity is the conflict severity
	Severity ConflictSeverity

	// Description explains the conflict
	Description string
}

// ConflictSeverity indicates how severe a type conflict is
type ConflictSeverity int

const (
	// ConflictSeverityWarning is a minor conflict
	ConflictSeverityWarning ConflictSeverity = iota
	// ConflictSeverityError is a serious conflict
	ConflictSeverityError
	// ConflictSeverityCritical is a critical conflict that breaks type safety
	ConflictSeverityCritical
)

// ModuleAnalyzer analyzes individual modules
type ModuleAnalyzer struct {
	// Parser for parsing Perl files
	Parser parser.Parser

	// Introspector for deep module analysis
	Introspector *parser.ModuleIntrospector

	// TypeExtractor extracts type information
	TypeExtractor *TypeExtractor
}

// TypeExtractor extracts type information from modules
type TypeExtractor struct {
	// ExtractedTypes stores extracted type information
	ExtractedTypes map[string]*ExtractedTypeInfo
}

// ExtractedTypeInfo contains extracted type information
type ExtractedTypeInfo struct {
	// Module is the source module
	Module string

	// Types are the types defined in the module
	Types []*typedef.TypeInfo

	// Exports are the exported symbols
	Exports []string

	// UsedTypes are types used but not defined
	UsedTypes []string
}

// IncrementalCache provides caching for incremental analysis
type IncrementalCache struct {
	// CacheDir is the cache directory
	CacheDir string

	// ModuleCache caches module analysis results
	ModuleCache map[string]*CachedModuleAnalysis

	// DependencyCache caches dependency information
	DependencyCache map[string]*CachedDependencies

	// mutex for thread-safe operations
	mu sync.RWMutex
}

// CachedModuleAnalysis represents cached analysis for a module
type CachedModuleAnalysis struct {
	// ModuleName is the module name
	ModuleName string

	// FileHash is the hash of the file content
	FileHash string

	// LastAnalyzed is when the analysis was performed
	LastAnalyzed time.Time

	// TypeInfo contains extracted type information
	TypeInfo *ExtractedTypeInfo

	// Dependencies are the module dependencies
	Dependencies []string
}

// CachedDependencies represents cached dependency information
type CachedDependencies struct {
	// ProjectHash is a hash of all relevant files
	ProjectHash string

	// Graph is the cached dependency graph
	Graph *DependencyGraph

	// LastUpdated is when the cache was updated
	LastUpdated time.Time
}

// PackageManager interface for package manager integration
type PackageManager interface {
	// GetInstalledModules returns installed CPAN modules
	GetInstalledModules() ([]string, error)

	// GetModuleInfo returns information about a module
	GetModuleInfo(module string) (*ModuleInfo, error)

	// GetDependencies returns module dependencies
	GetDependencies(module string) ([]string, error)
}

// ModuleInfo contains information about a CPAN module
type ModuleInfo struct {
	// Name is the module name
	Name string

	// Version is the installed version
	Version string

	// Path is the installation path
	Path string

	// Dependencies are the module dependencies
	Dependencies []string
}

// ProjectAnalysisResult contains the results of project analysis
type ProjectAnalysisResult struct {
	// RootDir is the analyzed project root
	RootDir string

	// AnalyzedAt is when the analysis was performed
	AnalyzedAt time.Time

	// ModuleCount is the number of modules analyzed
	ModuleCount int

	// DependencyGraph is the project dependency graph
	DependencyGraph *DependencyGraph

	// TypeConflicts are detected type conflicts
	TypeConflicts []TypeConflict

	// TypeSummary summarizes types across the project
	TypeSummary *ProjectTypeSummary

	// Errors are errors encountered during analysis
	Errors []error

	// Warnings are warnings from analysis
	Warnings []string
}

// ProjectTypeSummary summarizes type information across the project
type ProjectTypeSummary struct {
	// TotalTypes is the total number of types defined
	TotalTypes int

	// TypesByModule maps modules to their defined types
	TypesByModule map[string][]string

	// SharedTypes are types used across multiple modules
	SharedTypes map[string][]string

	// UnresolvedTypes are types referenced but not defined
	UnresolvedTypes []string
}

// NewProjectAnalyzer creates a new project analyzer
func NewProjectAnalyzer(rootDir string, config *AnalyzerConfig) (*ProjectAnalyzer, error) {
	// Set default config if not provided
	if config == nil {
		config = &AnalyzerConfig{
			MaxDepth:      10,
			EnableCaching: true,
			CacheDir:      filepath.Join(rootDir, ".psc-cache"),
			Parallel:      true,
			MaxWorkers:    4,
		}
	}

	// Create type checker
	typeChecker, err := typechecker.NewTypeCheck()
	if err != nil {
		return nil, fmt.Errorf("failed to create type checker: %w", err)
	}

	// Create module analyzer
	moduleAnalyzer := &ModuleAnalyzer{
		TypeExtractor: &TypeExtractor{
			ExtractedTypes: make(map[string]*ExtractedTypeInfo),
		},
	}

	// Create parser
	p, err := parser.NewParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}
	moduleAnalyzer.Parser = p

	// Create introspector
	introspector, err := parser.NewModuleIntrospector()
	if err != nil {
		return nil, fmt.Errorf("failed to create introspector: %w", err)
	}
	moduleAnalyzer.Introspector = introspector

	analyzer := &ProjectAnalyzer{
		RootDir: rootDir,
		DependencyGraph: &DependencyGraph{
			Nodes: make(map[string]*ModuleNode),
			Edges: make(map[string][]*DependencyEdge),
		},
		TypeConflictDetector: &TypeConflictDetector{
			Conflicts:    []TypeConflict{},
			TypeRegistry: make(map[string]map[string]*typedef.TypeInfo),
		},
		ModuleAnalyzer: moduleAnalyzer,
		IncrementalCache: &IncrementalCache{
			CacheDir:        config.CacheDir,
			ModuleCache:     make(map[string]*CachedModuleAnalysis),
			DependencyCache: make(map[string]*CachedDependencies),
		},
		TypeChecker: typeChecker,
		Config:      config,
		Results: &ProjectAnalysisResult{
			RootDir:       rootDir,
			TypeConflicts: []TypeConflict{},
			Errors:        []error{},
			Warnings:      []string{},
		},
	}

	// Load cache if enabled
	if config.EnableCaching {
		if err := analyzer.loadCache(); err != nil {
			// Cache load failure is not fatal
			analyzer.Results.Warnings = append(analyzer.Results.Warnings,
				fmt.Sprintf("Failed to load cache: %v", err))
		}
	}

	return analyzer, nil
}

// AnalyzeProject performs comprehensive analysis on the project
func (pa *ProjectAnalyzer) AnalyzeProject() (*ProjectAnalysisResult, error) {
	pa.Results.AnalyzedAt = time.Now()

	// Phase 1: Discover all Perl modules
	modules, err := pa.discoverModules()
	if err != nil {
		return nil, fmt.Errorf("failed to discover modules: %w", err)
	}
	pa.Results.ModuleCount = len(modules)

	// Phase 2: Build dependency graph
	if err := pa.buildDependencyGraph(modules); err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %w", err)
	}
	pa.Results.DependencyGraph = pa.DependencyGraph

	// Phase 3: Analyze types in each module
	if pa.Config.Parallel {
		err = pa.analyzeModulesParallel(modules)
	} else {
		err = pa.analyzeModulesSequential(modules)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to analyze modules: %w", err)
	}

	// Phase 4: Detect type conflicts
	pa.detectTypeConflicts()
	pa.Results.TypeConflicts = pa.TypeConflictDetector.Conflicts

	// Phase 5: Generate type summary
	pa.Results.TypeSummary = pa.generateTypeSummary()

	// Phase 6: Save cache if enabled
	if pa.Config.EnableCaching {
		if err := pa.saveCache(); err != nil {
			pa.Results.Warnings = append(pa.Results.Warnings,
				fmt.Sprintf("Failed to save cache: %v", err))
		}
	}

	return pa.Results, nil
}

// discoverModules finds all Perl modules in the project
func (pa *ProjectAnalyzer) discoverModules() ([]string, error) {
	var modules []string

	err := filepath.Walk(pa.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded patterns
		for _, pattern := range pa.Config.ExcludePatterns {
			matched, _ := filepath.Match(pattern, path)
			if matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Check if it's a Perl file
		if !info.IsDir() && (strings.HasSuffix(path, ".pm") || strings.HasSuffix(path, ".pl")) {
			modules = append(modules, path)
		}

		return nil
	})

	return modules, err
}

// buildDependencyGraph builds the module dependency graph
func (pa *ProjectAnalyzer) buildDependencyGraph(modules []string) error {
	for _, module := range modules {
		if err := pa.analyzeModuleDependencies(module); err != nil {
			pa.Results.Errors = append(pa.Results.Errors, err)
		}
	}
	return nil
}

// analyzeModuleDependencies analyzes dependencies for a single module
func (pa *ProjectAnalyzer) analyzeModuleDependencies(modulePath string) error {
	// Check cache first
	if pa.Config.EnableCaching {
		if cached := pa.getCachedModuleAnalysis(modulePath); cached != nil {
			// Use cached dependencies
			pa.updateDependencyGraph(modulePath, cached.Dependencies)
			return nil
		}
	}

	// Parse the module to find dependencies
	content, err := os.ReadFile(modulePath)
	if err != nil {
		return err
	}

	dependencies, err := pa.extractDependencies(string(content))
	if err != nil {
		// Log the error but continue with empty dependencies for graceful degradation
		fmt.Fprintf(os.Stderr, "Warning: Failed to extract dependencies from %s: %v\n", modulePath, err)
		dependencies = []string{}
	}
	pa.updateDependencyGraph(modulePath, dependencies)

	return nil
}

// extractDependencies extracts module dependencies from content using AST parsing
func (pa *ProjectAnalyzer) extractDependencies(content string) ([]string, error) {
	// Use the same AST-based dependency extraction as PVX
	parser, err := parser.NewParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	astRoot, err := parser.ParseString(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	// Extract dependencies using the same logic as PVX
	dependencies := pa.extractDependenciesFromASTWithFallback(astRoot, content)

	return dependencies, nil
}

// extractDependenciesFromASTWithFallback mirrors PVX's hybrid AST + regex extraction
func (pa *ProjectAnalyzer) extractDependenciesFromASTWithFallback(astRoot *ast.AST, originalContent string) []string {
	// First try AST-based extraction
	astDeps := pa.extractDependenciesFromAST(astRoot)

	// Also try regex-based extraction as fallback for edge cases
	regexDeps := pa.extractDependenciesFromRegex(originalContent)

	// Combine and deduplicate
	astDeps = append(astDeps, regexDeps...)
	return pa.filterAndDeduplicateDependencies(astDeps)
}

// extractDependenciesFromAST traverses AST to find UseStmt nodes and require expressions
func (pa *ProjectAnalyzer) extractDependenciesFromAST(astRoot *ast.AST) []string {
	if astRoot == nil || astRoot.Root == nil {
		return []string{}
	}

	var dependencies []string
	visited := make(map[ast.Node]bool)

	// Traverse the AST to find UseStmt nodes and require expressions
	pa.traverseASTForDependencies(astRoot.Root, &dependencies, visited)

	return dependencies
}

// traverseASTForDependencies recursively traverses AST nodes
func (pa *ProjectAnalyzer) traverseASTForDependencies(node ast.Node, dependencies *[]string, visited map[ast.Node]bool) {
	if node == nil || visited[node] {
		return
	}
	visited[node] = true

	// Check if this node is a UseStmt
	if useStmt, ok := node.(*ast.UseStmt); ok {
		if useStmt.Module != "" {
			*dependencies = append(*dependencies, useStmt.Module)
		}
	}

	// Check for require expressions
	if node.Type() == "require_expression" || node.Type() == "require_version_expression" {
		if module := pa.extractModuleFromRequireExpression(node); module != "" {
			*dependencies = append(*dependencies, module)
		}
	}

	// Check for expression statements that might contain require calls
	if node.Type() == "expression_statement" {
		text := node.Text()
		if module := pa.extractModuleFromRequireText(text); module != "" {
			*dependencies = append(*dependencies, module)
		}
	}

	// Traverse child nodes
	for _, child := range node.Children() {
		pa.traverseASTForDependencies(child, dependencies, visited)
	}
}

// extractModuleFromRequireExpression extracts module name from require_expression AST nodes
func (pa *ProjectAnalyzer) extractModuleFromRequireExpression(node ast.Node) string {
	for _, child := range node.Children() {
		childType := child.Type()
		text := child.Text()

		if childType == "string" || childType == "identifier" || childType == "token" ||
			childType == "interpolated_string_literal" || childType == "literal" {

			if text != "" && text != "require" {
				return pa.normalizeRequiredModule(text)
			}
		}

		if childType == "interpolated_string_literal" {
			for _, grandchild := range child.Children() {
				grandText := grandchild.Text()
				if grandText != "" {
					return pa.normalizeRequiredModule(grandText)
				}
			}
		}
	}
	return ""
}

// extractModuleFromRequireText extracts module name from require statement text (fallback)
func (pa *ProjectAnalyzer) extractModuleFromRequireText(text string) string {
	text = strings.TrimSpace(text)

	if strings.HasPrefix(text, "require ") {
		module := strings.TrimPrefix(text, "require ")
		module = strings.TrimSuffix(module, ";")
		module = strings.TrimSpace(module)

		return pa.normalizeRequiredModule(module)
	}

	return ""
}

// normalizeRequiredModule normalizes a required module name
func (pa *ProjectAnalyzer) normalizeRequiredModule(module string) string {
	// Remove quotes
	module = strings.Trim(module, `"'`)

	// Convert .pm file paths to module names
	if strings.HasSuffix(module, ".pm") {
		module = strings.TrimSuffix(module, ".pm")
		module = strings.ReplaceAll(module, "/", "::")
	}

	if module == "" {
		return ""
	}

	return module
}

// extractDependenciesFromRegex provides regex-based fallback extraction
func (pa *ProjectAnalyzer) extractDependenciesFromRegex(content string) []string {
	var dependencies []string

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Match require statements with quoted strings
		if strings.HasPrefix(line, "require ") && (strings.Contains(line, `"`) || strings.Contains(line, `'`)) {
			module := strings.TrimPrefix(line, "require ")
			module = strings.TrimSuffix(module, ";")
			module = strings.TrimSpace(module)
			module = strings.Trim(module, `"'`)

			if module != "" {
				dependencies = append(dependencies, pa.normalizeRequiredModule(module))
			}
		}
	}

	return dependencies
}

// filterAndDeduplicateDependencies removes duplicates and filters out Perl pragmas
func (pa *ProjectAnalyzer) filterAndDeduplicateDependencies(deps []string) []string {
	seen := make(map[string]bool)
	var filtered []string

	// List of common Perl pragmas to filter out
	pragmas := map[string]bool{
		"strict": true, "warnings": true, "utf8": true, "feature": true,
		"vars": true, "lib": true, "base": true, "parent": true,
		"constant": true, "autodie": true, "experimental": true,
		"bigint": true, "bignum": true, "bigrat": true, "integer": true,
		"bytes": true, "charnames": true, "diagnostics": true,
		"encoding": true, "fields": true, "filetest": true, "if": true,
		"less": true, "locale": true, "open": true, "ops": true,
		"overload": true, "re": true, "sigtrap": true, "sort": true,
		"subs": true, "threads": true, "version": true,
	}

	for _, dep := range deps {
		if dep != "" && !seen[dep] && !pragmas[dep] {
			seen[dep] = true
			filtered = append(filtered, dep)
		}
	}

	return filtered
}

// updateDependencyGraph updates the dependency graph with module dependencies
func (pa *ProjectAnalyzer) updateDependencyGraph(modulePath string, dependencies []string) {
	pa.DependencyGraph.mu.Lock()
	defer pa.DependencyGraph.mu.Unlock()

	moduleName := pa.pathToModuleName(modulePath)

	// Add or update node
	if _, exists := pa.DependencyGraph.Nodes[moduleName]; !exists {
		pa.DependencyGraph.Nodes[moduleName] = &ModuleNode{
			Name:         moduleName,
			Path:         modulePath,
			Dependencies: dependencies,
			Dependents:   []string{},
		}
	} else {
		pa.DependencyGraph.Nodes[moduleName].Dependencies = dependencies
	}

	// Add edges
	for _, dep := range dependencies {
		edge := &DependencyEdge{
			From: moduleName,
			To:   dep,
			Type: "use",
		}
		pa.DependencyGraph.Edges[moduleName] = append(pa.DependencyGraph.Edges[moduleName], edge)

		// Update dependents
		if depNode, exists := pa.DependencyGraph.Nodes[dep]; exists {
			depNode.Dependents = append(depNode.Dependents, moduleName)
		}
	}
}

// pathToModuleName converts a file path to a module name
func (pa *ProjectAnalyzer) pathToModuleName(path string) string {
	// Remove root directory
	rel, _ := filepath.Rel(pa.RootDir, path)

	// Convert to module name
	module := strings.TrimSuffix(rel, ".pm")
	module = strings.TrimSuffix(module, ".pl")
	module = strings.ReplaceAll(module, string(filepath.Separator), "::")

	// Remove lib/ prefix if present
	module = strings.TrimPrefix(module, "lib::")

	return module
}

// analyzeModulesParallel analyzes modules in parallel
func (pa *ProjectAnalyzer) analyzeModulesParallel(modules []string) error {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, pa.Config.MaxWorkers)
	errorChan := make(chan error, len(modules))

	for _, module := range modules {
		wg.Add(1)
		go func(modulePath string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := pa.analyzeModule(modulePath); err != nil {
				errorChan <- err
			}
		}(module)
	}

	wg.Wait()
	close(errorChan)

	// Collect errors
	for err := range errorChan {
		pa.Results.Errors = append(pa.Results.Errors, err)
	}

	return nil
}

// analyzeModulesSequential analyzes modules sequentially
func (pa *ProjectAnalyzer) analyzeModulesSequential(modules []string) error {
	for _, module := range modules {
		if err := pa.analyzeModule(module); err != nil {
			pa.Results.Errors = append(pa.Results.Errors, err)
		}
	}
	return nil
}

// analyzeModule performs type analysis on a single module
func (pa *ProjectAnalyzer) analyzeModule(modulePath string) error {
	// Type check the module
	result, err := pa.TypeChecker.CheckFile(modulePath)
	if err != nil {
		return fmt.Errorf("failed to type check %s: %w", modulePath, err)
	}

	// Extract type information
	typeInfo := pa.extractTypeInfo(modulePath, result)

	// Store in type registry
	pa.registerTypes(modulePath, typeInfo)

	// Cache the results if enabled
	if pa.Config.EnableCaching {
		pa.cacheModuleAnalysis(modulePath, typeInfo)
	}

	return nil
}

// extractTypeInfo extracts type information from type check results
func (pa *ProjectAnalyzer) extractTypeInfo(modulePath string, result *typechecker.TypeCheckResult) *ExtractedTypeInfo {
	moduleName := pa.pathToModuleName(modulePath)

	typeInfo := &ExtractedTypeInfo{
		Module:    moduleName,
		Types:     []*typedef.TypeInfo{},
		Exports:   []string{},
		UsedTypes: []string{},
	}

	// Extract types from annotations
	for _, annotation := range result.TypeAnnotations {
		// This is simplified - would need more sophisticated extraction
		if annotation.TypeExpression != nil {
			typeInfo.UsedTypes = append(typeInfo.UsedTypes, annotation.TypeExpression.Name)
		}
	}

	pa.ModuleAnalyzer.TypeExtractor.ExtractedTypes[moduleName] = typeInfo
	return typeInfo
}

// registerTypes registers types in the type registry
func (pa *ProjectAnalyzer) registerTypes(modulePath string, typeInfo *ExtractedTypeInfo) {
	pa.TypeConflictDetector.mu.Lock()
	defer pa.TypeConflictDetector.mu.Unlock()

	moduleName := pa.pathToModuleName(modulePath)

	if _, exists := pa.TypeConflictDetector.TypeRegistry[moduleName]; !exists {
		pa.TypeConflictDetector.TypeRegistry[moduleName] = make(map[string]*typedef.TypeInfo)
	}

	for _, typeInfo := range typeInfo.Types {
		pa.TypeConflictDetector.TypeRegistry[moduleName][typeInfo.Name] = typeInfo
	}
}

// detectTypeConflicts detects type conflicts across modules
func (pa *ProjectAnalyzer) detectTypeConflicts() {
	pa.TypeConflictDetector.mu.Lock()
	defer pa.TypeConflictDetector.mu.Unlock()

	// Map to track type definitions across modules
	typeModules := make(map[string][]string)

	// Collect all type definitions
	for module, types := range pa.TypeConflictDetector.TypeRegistry {
		for typeName := range types {
			typeModules[typeName] = append(typeModules[typeName], module)
		}
	}

	// Check for conflicts
	for typeName, modules := range typeModules {
		if len(modules) > 1 {
			// Type defined in multiple modules - check if they're compatible
			for i := 0; i < len(modules)-1; i++ {
				for j := i + 1; j < len(modules); j++ {
					type1 := pa.TypeConflictDetector.TypeRegistry[modules[i]][typeName]
					type2 := pa.TypeConflictDetector.TypeRegistry[modules[j]][typeName]

					if !pa.areTypesCompatible(type1, type2) {
						conflict := TypeConflict{
							TypeName:    typeName,
							Module1:     modules[i],
							Type1:       type1,
							Module2:     modules[j],
							Type2:       type2,
							Severity:    pa.determineConflictSeverity(type1, type2),
							Description: pa.describeConflict(type1, type2),
						}
						pa.TypeConflictDetector.Conflicts = append(pa.TypeConflictDetector.Conflicts, conflict)
					}
				}
			}
		}
	}
}

// areTypesCompatible checks if two type definitions are compatible
func (pa *ProjectAnalyzer) areTypesCompatible(type1, type2 *typedef.TypeInfo) bool {
	// Simple compatibility check - would need more sophisticated comparison
	if type1 == nil || type2 == nil {
		return true
	}

	// Check if they have the same kind
	if type1.Kind != type2.Kind {
		return false
	}

	// Check if they have the same parent
	if type1.Parent != type2.Parent {
		return false
	}

	// More checks would go here...
	return true
}

// determineConflictSeverity determines how severe a type conflict is
func (pa *ProjectAnalyzer) determineConflictSeverity(type1, type2 *typedef.TypeInfo) ConflictSeverity {
	if type1 == nil || type2 == nil {
		return ConflictSeverityWarning
	}

	// Different kinds is critical
	if type1.Kind != type2.Kind {
		return ConflictSeverityCritical
	}

	// Different parents is an error
	if type1.Parent != type2.Parent {
		return ConflictSeverityError
	}

	return ConflictSeverityWarning
}

// describeConflict creates a description of the type conflict
func (pa *ProjectAnalyzer) describeConflict(type1, type2 *typedef.TypeInfo) string {
	if type1 == nil || type2 == nil {
		return "One or both type definitions are missing"
	}

	if type1.Kind != type2.Kind {
		return fmt.Sprintf("Type kind mismatch: %s vs %s", type1.Kind, type2.Kind)
	}

	if type1.Parent != type2.Parent {
		return fmt.Sprintf("Different parent types: %s vs %s", type1.Parent, type2.Parent)
	}

	return "Type definitions differ"
}

// generateTypeSummary generates a summary of types across the project
func (pa *ProjectAnalyzer) generateTypeSummary() *ProjectTypeSummary {
	summary := &ProjectTypeSummary{
		TypesByModule:   make(map[string][]string),
		SharedTypes:     make(map[string][]string),
		UnresolvedTypes: []string{},
	}

	// Count types and track where they're defined
	typeLocations := make(map[string][]string)

	pa.TypeConflictDetector.mu.RLock()
	for module, types := range pa.TypeConflictDetector.TypeRegistry {
		for typeName := range types {
			summary.TotalTypes++
			summary.TypesByModule[module] = append(summary.TypesByModule[module], typeName)
			typeLocations[typeName] = append(typeLocations[typeName], module)
		}
	}
	pa.TypeConflictDetector.mu.RUnlock()

	// Identify shared types
	for typeName, locations := range typeLocations {
		if len(locations) > 1 {
			summary.SharedTypes[typeName] = locations
		}
	}

	// Identify unresolved types
	allUsedTypes := make(map[string]bool)
	for _, typeInfo := range pa.ModuleAnalyzer.TypeExtractor.ExtractedTypes {
		for _, usedType := range typeInfo.UsedTypes {
			allUsedTypes[usedType] = true
		}
	}

	for usedType := range allUsedTypes {
		if _, defined := typeLocations[usedType]; !defined {
			// Skip built-in types
			if !isBuiltinType(usedType) {
				summary.UnresolvedTypes = append(summary.UnresolvedTypes, usedType)
			}
		}
	}

	return summary
}

// isBuiltinType checks if a type is a built-in Perl type
func isBuiltinType(typeName string) bool {
	builtins := []string{
		"Str", "Int", "Num", "Bool", "Any", "Undef",
		"ArrayRef", "HashRef", "CodeRef", "ScalarRef",
		"Object", "Maybe", "Optional",
	}

	for _, builtin := range builtins {
		if typeName == builtin || strings.HasPrefix(typeName, builtin+"[") {
			return true
		}
	}

	return false
}

// Cache management methods

// loadCache loads cached analysis results
func (pa *ProjectAnalyzer) loadCache() error {
	if !pa.Config.EnableCaching {
		return nil
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(pa.Config.CacheDir, 0755); err != nil {
		return err
	}

	// Load dependency cache
	depCachePath := filepath.Join(pa.Config.CacheDir, "dependencies.json")
	if data, err := os.ReadFile(depCachePath); err == nil {
		var cached CachedDependencies
		if err := json.Unmarshal(data, &cached); err == nil {
			// Verify cache validity
			if pa.isProjectHashValid(cached.ProjectHash) {
				pa.DependencyGraph = cached.Graph
			}
		}
	}

	// Load module caches
	moduleCacheDir := filepath.Join(pa.Config.CacheDir, "modules")
	if entries, err := os.ReadDir(moduleCacheDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				cachePath := filepath.Join(moduleCacheDir, entry.Name())
				if data, err := os.ReadFile(cachePath); err == nil {
					var cached CachedModuleAnalysis
					if err := json.Unmarshal(data, &cached); err == nil {
						pa.IncrementalCache.ModuleCache[cached.ModuleName] = &cached
					}
				}
			}
		}
	}

	return nil
}

// saveCache saves analysis results to cache
func (pa *ProjectAnalyzer) saveCache() error {
	if !pa.Config.EnableCaching {
		return nil
	}

	// Save dependency cache
	depCache := &CachedDependencies{
		ProjectHash: pa.calculateProjectHash(),
		Graph:       pa.DependencyGraph,
		LastUpdated: time.Now(),
	}

	depCachePath := filepath.Join(pa.Config.CacheDir, "dependencies.json")
	if data, err := json.MarshalIndent(depCache, "", "  "); err == nil {
		if err := os.WriteFile(depCachePath, data, 0644); err != nil {
			return err
		}
	}

	// Save module caches
	moduleCacheDir := filepath.Join(pa.Config.CacheDir, "modules")
	if err := os.MkdirAll(moduleCacheDir, 0755); err != nil {
		return err
	}

	pa.IncrementalCache.mu.RLock()
	for _, cached := range pa.IncrementalCache.ModuleCache {
		cachePath := filepath.Join(moduleCacheDir, cached.ModuleName+".json")
		if data, err := json.MarshalIndent(cached, "", "  "); err == nil {
			os.WriteFile(cachePath, data, 0644)
		}
	}
	pa.IncrementalCache.mu.RUnlock()

	return nil
}

// getCachedModuleAnalysis retrieves cached analysis for a module
func (pa *ProjectAnalyzer) getCachedModuleAnalysis(modulePath string) *CachedModuleAnalysis {
	moduleName := pa.pathToModuleName(modulePath)

	pa.IncrementalCache.mu.RLock()
	cached, exists := pa.IncrementalCache.ModuleCache[moduleName]
	pa.IncrementalCache.mu.RUnlock()

	if !exists {
		return nil
	}

	// Verify cache validity
	currentHash := pa.calculateFileHash(modulePath)
	if currentHash != cached.FileHash {
		return nil
	}

	return cached
}

// cacheModuleAnalysis caches module analysis results
func (pa *ProjectAnalyzer) cacheModuleAnalysis(modulePath string, typeInfo *ExtractedTypeInfo) {
	moduleName := pa.pathToModuleName(modulePath)

	cached := &CachedModuleAnalysis{
		ModuleName:   moduleName,
		FileHash:     pa.calculateFileHash(modulePath),
		LastAnalyzed: time.Now(),
		TypeInfo:     typeInfo,
	}

	// Extract dependencies for caching
	if node, exists := pa.DependencyGraph.Nodes[moduleName]; exists {
		cached.Dependencies = node.Dependencies
	}

	pa.IncrementalCache.mu.Lock()
	pa.IncrementalCache.ModuleCache[moduleName] = cached
	pa.IncrementalCache.mu.Unlock()
}

// calculateFileHash calculates the hash of a file
func (pa *ProjectAnalyzer) calculateFileHash(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash)
}

// calculateProjectHash calculates a hash representing the project state
func (pa *ProjectAnalyzer) calculateProjectHash() string {
	h := md5.New()

	// Walk through all Perl files and add their modification times
	filepath.Walk(pa.RootDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() &&
			(strings.HasSuffix(path, ".pm") || strings.HasSuffix(path, ".pl")) {
			io.WriteString(h, fmt.Sprintf("%s:%d", path, info.ModTime().Unix()))
		}
		return nil
	})

	return fmt.Sprintf("%x", h.Sum(nil))
}

// isProjectHashValid checks if a project hash is still valid
func (pa *ProjectAnalyzer) isProjectHashValid(hash string) bool {
	return hash == pa.calculateProjectHash()
}

// Utility methods

// GetDependencyTree returns a dependency tree for a module
func (pa *ProjectAnalyzer) GetDependencyTree(moduleName string, maxDepth int) (*DependencyTree, error) {
	tree := &DependencyTree{
		Root:     moduleName,
		Children: []*DependencyTree{},
	}

	pa.buildDependencyTree(tree, moduleName, 0, maxDepth, make(map[string]bool))
	return tree, nil
}

// DependencyTree represents a tree view of dependencies
type DependencyTree struct {
	Root     string
	Children []*DependencyTree
}

// buildDependencyTree recursively builds a dependency tree
func (pa *ProjectAnalyzer) buildDependencyTree(tree *DependencyTree, module string, depth, maxDepth int, visited map[string]bool) {
	if depth >= maxDepth || visited[module] {
		return
	}
	visited[module] = true

	pa.DependencyGraph.mu.RLock()
	node, exists := pa.DependencyGraph.Nodes[module]
	pa.DependencyGraph.mu.RUnlock()

	if !exists {
		return
	}

	for _, dep := range node.Dependencies {
		child := &DependencyTree{
			Root:     dep,
			Children: []*DependencyTree{},
		}
		tree.Children = append(tree.Children, child)
		pa.buildDependencyTree(child, dep, depth+1, maxDepth, visited)
	}
}

// GetCircularDependencies detects circular dependencies
func (pa *ProjectAnalyzer) GetCircularDependencies() [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	pa.DependencyGraph.mu.RLock()
	defer pa.DependencyGraph.mu.RUnlock()

	for module := range pa.DependencyGraph.Nodes {
		if !visited[module] {
			pa.detectCycles(module, visited, recStack, &path, &cycles)
		}
	}

	return cycles
}

// detectCycles uses DFS to detect circular dependencies
func (pa *ProjectAnalyzer) detectCycles(module string, visited, recStack map[string]bool, path *[]string, cycles *[][]string) {
	visited[module] = true
	recStack[module] = true
	*path = append(*path, module)

	if node, exists := pa.DependencyGraph.Nodes[module]; exists {
		for _, dep := range node.Dependencies {
			if !visited[dep] {
				pa.detectCycles(dep, visited, recStack, path, cycles)
			} else if recStack[dep] {
				// Found a cycle
				cycleStart := -1
				for i, m := range *path {
					if m == dep {
						cycleStart = i
						break
					}
				}
				if cycleStart != -1 {
					cycle := make([]string, len(*path)-cycleStart)
					copy(cycle, (*path)[cycleStart:])
					*cycles = append(*cycles, cycle)
				}
			}
		}
	}

	recStack[module] = false
	*path = (*path)[:len(*path)-1]
}

// GenerateReport generates a comprehensive analysis report
func (pa *ProjectAnalyzer) GenerateReport() string {
	var report strings.Builder

	report.WriteString("Project Analysis Report\n")
	report.WriteString("======================\n\n")

	report.WriteString(fmt.Sprintf("Project Root: %s\n", pa.Results.RootDir))
	report.WriteString(fmt.Sprintf("Analysis Date: %s\n", pa.Results.AnalyzedAt.Format("2006-01-02 15:04:05")))
	report.WriteString(fmt.Sprintf("Modules Analyzed: %d\n\n", pa.Results.ModuleCount))

	// Dependency summary
	report.WriteString("Dependency Summary\n")
	report.WriteString("------------------\n")
	report.WriteString(fmt.Sprintf("Total Dependencies: %d\n", len(pa.DependencyGraph.Edges)))

	// Check for circular dependencies
	cycles := pa.GetCircularDependencies()
	if len(cycles) > 0 {
		report.WriteString(fmt.Sprintf("Circular Dependencies Found: %d\n", len(cycles)))
		for i, cycle := range cycles {
			report.WriteString(fmt.Sprintf("  %d. %s\n", i+1, strings.Join(cycle, " -> ")))
		}
	} else {
		report.WriteString("No circular dependencies detected\n")
	}
	report.WriteString("\n")

	// Type summary
	if pa.Results.TypeSummary != nil {
		report.WriteString("Type Summary\n")
		report.WriteString("------------\n")
		report.WriteString(fmt.Sprintf("Total Types Defined: %d\n", pa.Results.TypeSummary.TotalTypes))
		report.WriteString(fmt.Sprintf("Modules with Types: %d\n", len(pa.Results.TypeSummary.TypesByModule)))
		report.WriteString(fmt.Sprintf("Shared Types: %d\n", len(pa.Results.TypeSummary.SharedTypes)))
		report.WriteString(fmt.Sprintf("Unresolved Types: %d\n", len(pa.Results.TypeSummary.UnresolvedTypes)))

		if len(pa.Results.TypeSummary.UnresolvedTypes) > 0 {
			report.WriteString("\nUnresolved Types:\n")
			for _, t := range pa.Results.TypeSummary.UnresolvedTypes {
				report.WriteString(fmt.Sprintf("  - %s\n", t))
			}
		}
		report.WriteString("\n")
	}

	// Type conflicts
	if len(pa.Results.TypeConflicts) > 0 {
		report.WriteString("Type Conflicts\n")
		report.WriteString("--------------\n")

		// Group by severity
		critical := 0
		errors := 0
		warnings := 0

		for _, conflict := range pa.Results.TypeConflicts {
			switch conflict.Severity {
			case ConflictSeverityCritical:
				critical++
			case ConflictSeverityError:
				errors++
			case ConflictSeverityWarning:
				warnings++
			}
		}

		report.WriteString(fmt.Sprintf("Critical: %d, Errors: %d, Warnings: %d\n\n", critical, errors, warnings))

		for _, conflict := range pa.Results.TypeConflicts {
			severity := "WARNING"
			switch conflict.Severity {
			case ConflictSeverityCritical:
				severity = "CRITICAL"
			case ConflictSeverityError:
				severity = "ERROR"
			}

			report.WriteString(fmt.Sprintf("[%s] Type '%s' conflict between %s and %s\n",
				severity, conflict.TypeName, conflict.Module1, conflict.Module2))
			report.WriteString(fmt.Sprintf("  %s\n", conflict.Description))
		}
		report.WriteString("\n")
	}

	// Errors and warnings
	if len(pa.Results.Errors) > 0 {
		report.WriteString("Errors\n")
		report.WriteString("------\n")
		for _, err := range pa.Results.Errors {
			report.WriteString(fmt.Sprintf("  - %v\n", err))
		}
		report.WriteString("\n")
	}

	if len(pa.Results.Warnings) > 0 {
		report.WriteString("Warnings\n")
		report.WriteString("--------\n")
		for _, warning := range pa.Results.Warnings {
			report.WriteString(fmt.Sprintf("  - %s\n", warning))
		}
		report.WriteString("\n")
	}

	return report.String()
}
