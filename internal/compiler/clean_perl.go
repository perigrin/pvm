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

	// Get the original content from AST
	content, err := ast.GetContent()
	if err != nil {
		return "", NewCompilerError(ErrCompilationFailed,
			fmt.Sprintf("failed to get source content: %v", err)).
			WithLocation(ast.GetPath(), 0, 0).
			WithCause(err)
	}

	// Strip type annotations using regex patterns
	cleanCode := c.stripTypeAnnotations(content)

	return cleanCode, nil
}

// SetOptions updates the compiler options
func (c *CleanPerlCompiler) SetOptions(options *CompilerOptions) {
	c.options = options
}

// stripTypeAnnotations removes type annotations using regex patterns
func (c *CleanPerlCompiler) stripTypeAnnotations(code string) string {
	// Process line by line for better control
	lines := strings.Split(code, "\n")
	hasSignatures := false

	for i, line := range lines {
		// Check if this line contains a function signature
		if strings.Contains(line, "sub ") && strings.Contains(line, "(") && strings.Contains(line, ")") {
			funcPattern := regexp.MustCompile(`\bsub\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(([^)]*)\)`)
			if funcPattern.MatchString(line) {
				hasSignatures = true
			}
		}
		lines[i] = c.cleanLine(line)
	}

	result := strings.Join(lines, "\n")

	// If we found signatures, prepend v5.36 which includes signatures
	if hasSignatures {
		result = "use v5.36;\n" + result
	}

	return result
}

// cleanLine removes type annotations from a single line
func (c *CleanPerlCompiler) cleanLine(line string) string {
	// Handle variable declarations
	// Pattern: my Type $var or my Complex[Type, Type] $var
	// Match type names with optional parameterized types (including commas and spaces)
	varPattern := regexp.MustCompile(`\b(my|our|state)\s+[A-Z][a-zA-Z0-9_:]*(?:\[[^\]]*\])*\s+(\$[a-zA-Z_][a-zA-Z0-9_]*)`)
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
