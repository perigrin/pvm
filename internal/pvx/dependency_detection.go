// ABOUTME: Automatic dependency detection using PSC parser
// ABOUTME: Leverages PSC's Perl parsing to extract use/require statements

package pvx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/perl"
)

// AutoDetectDependencies extracts dependencies from a Perl script using PSC
func AutoDetectDependencies(scriptPath string) ([]string, error) {
	// Read the script file
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read script file: %w", err)
	}

	// Extract dependencies using PSC AST-based parsing
	dependencies, err := extractDependenciesFromContent(string(content))
	if err != nil {
		log.Warnf("Failed to parse dependencies with AST, falling back to empty list: %v", err)
		return []string{}, nil // Graceful degradation
	}

	if len(dependencies) > 0 {
		log.Debugf("Auto-detected %d dependencies from %s: %v", len(dependencies), scriptPath, dependencies)
	}

	return dependencies, nil
}

// AutoDetectAndStripTypedDependencies extracts dependencies and strips type annotations from typed modules
func AutoDetectAndStripTypedDependencies(scriptPath string, options *ExecutionOptions) ([]string, map[string]string, []func(), error) {
	// First detect all dependencies
	dependencies, err := AutoDetectDependencies(scriptPath)
	if err != nil {
		return nil, nil, nil, err
	}

	// Map of original module paths to stripped module paths
	strippedModulePaths := make(map[string]string)
	var cleanupFuncs []func()

	// Find and strip typed modules
	for _, dep := range dependencies {
		modulePath, err := findModulePath(dep, options)
		if err != nil {
			log.Debugf("Could not find module path for %s: %v", dep, err)
			continue
		}

		// Check if module contains type annotations
		if ContainsTypeAnnotations(modulePath) {
			log.Debugf("Module %s contains type annotations, stripping...", dep)

			// Strip type annotations from the module
			strippedPath, cleanupFunc, err := stripTypeAnnotationsFromScript(modulePath)
			if err != nil {
				log.Warnf("Failed to strip type annotations from module %s: %v", dep, err)
				continue
			}

			// Store the mapping and cleanup function
			strippedModulePaths[modulePath] = strippedPath
			cleanupFuncs = append(cleanupFuncs, cleanupFunc)
		}
	}

	if len(strippedModulePaths) > 0 {
		log.Debugf("Stripped type annotations from %d typed modules", len(strippedModulePaths))
	}

	return dependencies, strippedModulePaths, cleanupFuncs, nil
}

// extractDependenciesFromContent extracts module dependencies from Perl source code using AST parsing
func extractDependenciesFromContent(content string) ([]string, error) {
	// Create a new parser instance
	p, err := parser.NewParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	// Parse the content into an AST
	astRoot, err := p.ParseString(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	// Extract dependencies using AST traversal with original content as fallback
	dependencies := extractDependenciesFromASTWithFallback(astRoot, content)

	return dependencies, nil
}

// extractDependenciesFromASTWithFallback combines AST-based extraction with regex fallback for edge cases
func extractDependenciesFromASTWithFallback(astRoot *ast.AST, originalContent string) []string {
	// First try AST-based extraction
	astDeps := extractDependenciesFromAST(astRoot)

	// Also try regex-based extraction as fallback for cases where AST nodes have empty text
	regexDeps := extractDependenciesFromRegex(originalContent)

	// Combine and deduplicate
	astDeps = append(astDeps, regexDeps...)
	return filterAndDeduplicateDependencies(astDeps)
}

// extractDependenciesFromRegex provides regex-based fallback extraction
func extractDependenciesFromRegex(content string) []string {
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
			// Extract the quoted module name
			module := strings.TrimPrefix(line, "require ")
			module = strings.TrimSuffix(module, ";")
			module = strings.TrimSpace(module)
			module = strings.Trim(module, `"'`)

			// Convert to normalized form
			if module != "" {
				dependencies = append(dependencies, normalizeRequiredModule(module))
			}
		}
	}

	return dependencies
}

// extractDependenciesFromAST traverses the AST to find all use statements and extract module dependencies
func extractDependenciesFromAST(astRoot *ast.AST) []string {
	if astRoot == nil || astRoot.Root == nil {
		return []string{}
	}

	var dependencies []string
	visited := make(map[ast.Node]bool)

	// Traverse the AST to find UseStmt nodes
	traverseASTForDependencies(astRoot.Root, &dependencies, visited)

	// Remove duplicates and filter out pragmas
	return filterAndDeduplicateDependencies(dependencies)
}

// traverseASTForDependencies recursively traverses AST nodes to find UseStmt nodes and require expressions
func traverseASTForDependencies(node ast.Node, dependencies *[]string, visited map[ast.Node]bool) {
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

	// Check for require expressions by examining node text
	if node.Type() == "require_expression" || node.Type() == "require_version_expression" {
		if module := extractModuleFromRequireExpression(node); module != "" {
			*dependencies = append(*dependencies, module)
		}
	}

	// Check for expression statements that might contain require calls
	if node.Type() == "expression_statement" {
		text := node.Text()
		if module := extractModuleFromRequireText(text); module != "" {
			*dependencies = append(*dependencies, module)
		}
	}

	// Traverse child nodes
	for _, child := range node.Children() {
		traverseASTForDependencies(child, dependencies, visited)
	}
}

// extractModuleFromRequireExpression extracts module name from require_expression AST nodes
func extractModuleFromRequireExpression(node ast.Node) string {
	// The module name should be in the children of the require expression
	for _, child := range node.Children() {
		childType := child.Type()
		text := child.Text()

		// Look for various node types that might contain the module name
		if childType == "string" || childType == "identifier" || childType == "token" ||
			childType == "interpolated_string_literal" || childType == "literal" {

			// Skip empty text and keywords
			if text != "" && text != "require" {
				return normalizeRequiredModule(text)
			}
		}

		// For complex nodes, recursively search children
		if childType == "interpolated_string_literal" {
			for _, grandchild := range child.Children() {
				grandText := grandchild.Text()
				if grandText != "" {
					return normalizeRequiredModule(grandText)
				}
			}
		}
	}
	return ""
}

// extractModuleFromRequireText extracts module name from require statement text (fallback)
func extractModuleFromRequireText(text string) string {
	text = strings.TrimSpace(text)

	// Handle various require formats
	if strings.HasPrefix(text, "require ") {
		// Remove "require " prefix
		module := strings.TrimPrefix(text, "require ")
		// Remove trailing semicolon
		module = strings.TrimSuffix(module, ";")
		module = strings.TrimSpace(module)

		return normalizeRequiredModule(module)
	}

	return ""
}

// normalizeRequiredModule normalizes a required module name
func normalizeRequiredModule(module string) string {
	// Remove quotes
	module = strings.Trim(module, `"'`)

	// Convert .pm file paths to module names
	if strings.HasSuffix(module, ".pm") {
		module = strings.TrimSuffix(module, ".pm")
		module = strings.ReplaceAll(module, "/", "::")
	}

	// Skip empty modules
	if module == "" {
		return ""
	}

	return module
}

// filterAndDeduplicateDependencies removes duplicates and filters out Perl pragmas
func filterAndDeduplicateDependencies(deps []string) []string {
	seen := make(map[string]bool)
	var filtered []string

	// Use shared pragma list
	pragmas := getPragmaList()

	for _, dep := range deps {
		if dep != "" && !seen[dep] && !pragmas[dep] {
			seen[dep] = true
			filtered = append(filtered, dep)
		}
	}

	return filtered
}

// getPragmaList returns the map of Perl pragmas (shared between functions)
func getPragmaList() map[string]bool {
	return map[string]bool{
		"strict":       true,
		"warnings":     true,
		"utf8":         true,
		"feature":      true,
		"vars":         true,
		"lib":          true,
		"base":         true,
		"parent":       true,
		"constant":     true,
		"autodie":      true,
		"experimental": true,
		"bigint":       true,
		"bignum":       true,
		"bigrat":       true,
		"integer":      true,
		"bytes":        true,
		"charnames":    true,
		"diagnostics":  true,
		"encoding":     true,
		"fields":       true,
		"filetest":     true,
		"if":           true,
		"less":         true,
		"locale":       true,
		"open":         true,
		"ops":          true,
		"overload":     true,
		"re":           true,
		"sigtrap":      true,
		"sort":         true,
		"subs":         true,
		"threads":      true,
		"version":      true,
	}
}

// isPragma checks if a module name is a core Perl pragma
func isPragma(module string) bool {
	pragmas := getPragmaList()
	return pragmas[module]
}

// FilterCPANModules filters out modules that are likely core or built-in
func FilterCPANModules(dependencies []string) []string {
	var cpanModules []string

	coreModules := map[string]bool{
		// Core modules that ship with Perl
		"Carp":           true,
		"Data::Dumper":   true,
		"File::Basename": true,
		"File::Path":     true,
		"File::Spec":     true,
		"FindBin":        true,
		"Getopt::Long":   true,
		"IO::File":       true,
		"IO::Handle":     true,
		"List::Util":     true,
		"Scalar::Util":   true,
		"Time::Local":    true,
		"Time::Piece":    true,
		"POSIX":          true,
		"Storable":       true,
		"Socket":         true,
		"Fcntl":          true,
	}

	for _, dep := range dependencies {
		// Include if it's not a core module
		if !coreModules[dep] {
			cpanModules = append(cpanModules, dep)
		}
	}

	return cpanModules
}

// AutoDetectDependenciesWithOptions extracts dependencies with filtering options
func AutoDetectDependenciesWithOptions(scriptPath string, includeCoreModules bool) ([]string, error) {
	dependencies, err := AutoDetectDependencies(scriptPath)
	if err != nil {
		return nil, err
	}

	if !includeCoreModules {
		dependencies = FilterCPANModules(dependencies)
	}

	return dependencies, nil
}

// findModulePath attempts to locate a Perl module file in the filesystem
func findModulePath(moduleName string, options *ExecutionOptions) (string, error) {
	// Convert module name to file path (e.g., "My::Module" -> "My/Module.pm")
	filePath := strings.ReplaceAll(moduleName, "::", "/") + ".pm"

	// Get Perl version and resolve executable
	perlVersion := options.PerlVersion
	if perlVersion == "" {
		perlVersion = "system" // Default to system Perl
	}

	resolvedVersion, err := perl.ResolveVersion(&perl.ResolutionOptions{
		ExplicitVersion: perlVersion,
	})
	if err != nil {
		return "", fmt.Errorf("failed to resolve Perl version %s: %w", perlVersion, err)
	}

	// Get Perl's @INC directories
	incDirs, err := getPerlIncDirectories(resolvedVersion.Path)
	if err != nil {
		return "", fmt.Errorf("failed to get Perl @INC directories: %w", err)
	}

	// Add any additional module paths from options
	if len(options.AdditionalModulePaths) > 0 {
		incDirs = append(options.AdditionalModulePaths, incDirs...)
	}

	// Search for the module file in @INC directories
	for _, incDir := range incDirs {
		fullPath := filepath.Join(incDir, filePath)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("module %s not found in @INC directories", moduleName)
}

// getPerlIncDirectories gets the @INC directories from a Perl executable
func getPerlIncDirectories(perlPath string) ([]string, error) {
	// Use perl -E 'say for @INC' to get the include directories
	cmd := exec.Command(perlPath, "-E", "say for @INC")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute perl command: %w", err)
	}

	// Split output into lines and filter out empty lines
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var incDirs []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			incDirs = append(incDirs, line)
		}
	}

	return incDirs, nil
}
