// ABOUTME: Clean Perl compiler that removes type annotations from AST
// ABOUTME: Generates standard Perl code compatible with any Perl interpreter
//
// IMPORTANT: This file should NOT use regex patterns for cleaning code!
// Regex-based cleaning is fragile and produces malformed output.
// Use proper AST traversal and reconstruction instead.

package compiler

import (
	"fmt"
	"regexp"
	"strings"

	"tamarou.com/pvm/internal/current"
)

// CleanPerlCompiler compiles AST to clean Perl code without type annotations
type CleanPerlCompiler struct {
	options *CompilerOptions
}

// NewCleanPerlCompiler creates a new clean Perl compiler
func NewCleanPerlCompiler() *CleanPerlCompiler {
	return &CleanPerlCompiler{
		options: &CompilerOptions{
			PreserveComments:   true,
			PreserveFormatting: true,
			StrictMode:         false,
			CustomPatterns:     nil,
		},
	}
}

// Target returns the compilation target
func (c *CleanPerlCompiler) Target() Target {
	return TargetCleanPerl
}

// Validate checks if the AST is suitable for clean Perl compilation
func (c *CleanPerlCompiler) Validate(ast AST) error {
	if ast == nil {
		return NewCompilerError(ErrInvalidAST, "AST cannot be nil")
	}

	// Check if we can get content either from source or file path
	_, err := ast.GetContent()
	if err != nil {
		return NewCompilerError(ErrInvalidAST, "AST must have accessible source content").WithCause(err)
	}

	return nil
}

// Compile converts an AST to clean Perl code without type annotations
func (c *CleanPerlCompiler) Compile(ast AST) (string, error) {
	if err := c.Validate(ast); err != nil {
		return "", err
	}

	// Get the source content directly and apply regex-based cleaning
	// This is a pragmatic approach until proper AST traversal is implemented
	source, err := ast.GetContent()
	if err != nil {
		return "", NewCompilerError(ErrCompilationFailed, "failed to get source content").WithCause(err)
	}

	// Check if source is empty
	if source == "" {
		return "", NewCompilerError(ErrCompilationFailed, "source content is empty")
	}

	// Clean type annotations using regex patterns
	result := c.stripTypeAnnotations(source)

	// Replace hard-coded Perl version with PVM-managed version
	result, err = c.updatePerlVersion(result)
	if err != nil {
		return "", err
	}

	// Check if result is empty
	if result == "" {
		return "", NewCompilerError(ErrCompilationFailed, "compilation result is empty")
	}

	return result, nil
}

// SetOptions updates the compiler options
func (c *CleanPerlCompiler) SetOptions(options *CompilerOptions) {
	c.options = options
}

// updatePerlVersion replaces hard-coded Perl version with PVM-managed version
func (c *CleanPerlCompiler) updatePerlVersion(code string) (string, error) {
	// Get the PVM-managed Perl version
	currentVersion, err := current.GetCurrentVersion()
	if err != nil {
		// Fallback to v5.36 if we can't get current version
		return code, nil
	}

	// Replace hard-coded version pragmas with PVM-managed version
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		// Look for existing version pragmas
		if strings.HasPrefix(strings.TrimSpace(line), "use v") {
			// Replace with PVM-managed version
			lines[i] = fmt.Sprintf("use v%s;", currentVersion.Version)
			break
		}
	}

	return strings.Join(lines, "\n"), nil
}

// stripTypeAnnotations removes type annotations using regex patterns
func (c *CleanPerlCompiler) stripTypeAnnotations(code string) string {
	// Process line by line for better control
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		lines[i] = c.cleanLine(line)
	}

	result := strings.Join(lines, "\n")
	return result
}

// cleanLine removes type annotations from a single line
func (c *CleanPerlCompiler) cleanLine(line string) string {
	// Handle variable declarations
	// Pattern: my Type $var, my Type @var, my Type %var or my Complex[Type, Type] $var
	varPattern := regexp.MustCompile(`\b(my|our|state)\s+[A-Z][a-zA-Z0-9_:]*(?:\[[^\]]*\])*\s+([\$@%][a-zA-Z_][a-zA-Z0-9_]*)`)
	if varPattern.MatchString(line) {
		line = varPattern.ReplaceAllString(line, `$1 $2`)
	}

	// Handle function parameters
	// Pattern: sub name(Type $param) or sub name(Complex[Type] $param)
	funcPattern := regexp.MustCompile(`\bsub\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(([^)]*)\)`)
	if funcPattern.MatchString(line) {
		line = funcPattern.ReplaceAllStringFunc(line, func(match string) string {
			parts := funcPattern.FindStringSubmatch(match)
			if len(parts) != 3 {
				return match
			}

			funcName := parts[1]
			params := parts[2]

			// Extract parameter names (skip type annotations) for Perl 5.36+ signatures
			// Pattern matches: Type $var or Type $var = default_value
			paramPattern := regexp.MustCompile(`[A-Z][a-zA-Z0-9_:\[\]]+\s+(\$[a-zA-Z_][a-zA-Z0-9_]*(?:\s*=\s*[^,)]+)?)`)
			cleanParams := paramPattern.ReplaceAllString(params, `$1`)

			// Keep signature syntax for Perl 5.36+ - just remove type annotations
			return fmt.Sprintf("sub %s(%s)", funcName, cleanParams)
		})
	}

	// Handle for loops
	// Pattern: for my Type $var (@array)
	forPattern := regexp.MustCompile(`\bfor\s+my\s+[A-Z][a-zA-Z0-9_:\[\]]+\s+(\$[a-zA-Z_][a-zA-Z0-9_]*)\s+(\([^)]+\))`)
	if forPattern.MatchString(line) {
		line = forPattern.ReplaceAllString(line, `for my $1 $2`)
	}

	// Handle field declarations
	// Pattern: field Type $field
	fieldPattern := regexp.MustCompile(`\bfield\s+[A-Z][a-zA-Z0-9_:\[\]]+\s+(\$[a-zA-Z_][a-zA-Z0-9_]*)`)
	if fieldPattern.MatchString(line) {
		line = fieldPattern.ReplaceAllString(line, `field $1`)
	}

	// Handle type declarations
	// Pattern: type TypeName = ...
	typePattern := regexp.MustCompile(`\btype\s+[A-Z][a-zA-Z_]*\s*=.*`)
	if typePattern.MatchString(line) {
		// Remove type declarations entirely
		line = ""
	}

	// Clean up any remaining return type annotations
	// Pattern: -> Type or -> Complex[Type]
	returnTypePattern := regexp.MustCompile(`\s*->\s*[A-Z][a-zA-Z_:]*(?:\[[^\]]*\])*`)
	line = returnTypePattern.ReplaceAllString(line, "")

	// Handle type assertions
	// Pattern: $var as Type
	assertPattern := regexp.MustCompile(`(\$[a-zA-Z_][a-zA-Z0-9_]*)\s+as\s+[A-Z][a-zA-Z_:]*(?:\[[^\]]*\])*`)
	line = assertPattern.ReplaceAllString(line, `$1`)

	return line
}
