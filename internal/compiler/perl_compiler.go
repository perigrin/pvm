// ABOUTME: Unified Perl compiler that works directly with tree-sitter CST
// ABOUTME: Replaces separate CleanPerlCompiler and TypedPerlCompiler with a single implementation

package compiler

import (
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/current"
	"tamarou.com/pvm/internal/parser/treesitter"
)

// CSTBasedAST represents an AST that works directly with tree-sitter CST
type CSTBasedAST struct {
	Path    string
	Content string
	Root    *sitter.Node
}

// NewCSTBasedAST creates a new CST-based AST from source content
func NewCSTBasedAST(path string, content string) (*CSTBasedAST, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(treesitter.Language())

	contentBytes := []byte(content)
	tree := parser.Parse(contentBytes, nil)
	if tree == nil {
		return nil, fmt.Errorf("failed to parse content")
	}

	return &CSTBasedAST{
		Path:    path,
		Content: content,
		Root:    tree.RootNode(),
	}, nil
}

// GetPath returns the source file path
func (a *CSTBasedAST) GetPath() string {
	return a.Path
}

// IsValid returns true if the CST is valid for compilation
func (a *CSTBasedAST) IsValid() bool {
	return a.Root != nil
}

// GetContent returns the original source content
func (a *CSTBasedAST) GetContent() (string, error) {
	return a.Content, nil
}

// GetRootNode returns nil since we don't use the traditional AST
func (a *CSTBasedAST) GetRootNode() (ast.Node, error) {
	return nil, NewCompilerError(ErrInvalidAST, "CSTBasedAST does not support traditional AST nodes")
}

// GetCSTRoot returns the tree-sitter CST root node
func (a *CSTBasedAST) GetCSTRoot() *sitter.Node {
	return a.Root
}

// PerlCompiler is a unified compiler that works directly with tree-sitter CST
type PerlCompiler struct {
	target  Target
	options CompilerOptions
}

// NewPerlCompiler creates a new unified Perl compiler for the specified target
func NewPerlCompiler(target Target) *PerlCompiler {
	return &PerlCompiler{
		target: target,
		options: CompilerOptions{
			PreserveComments:   true,
			PreserveFormatting: true,
			StrictMode:         false,
			CustomPatterns:     nil,
		},
	}
}

// NewCleanPerlCompilerUnified creates a new unified compiler for clean Perl output
func NewCleanPerlCompilerUnified() *PerlCompiler {
	return NewPerlCompiler(TargetCleanPerl)
}

// NewTypedPerlCompilerUnified creates a new unified compiler for typed Perl output
func NewTypedPerlCompilerUnified() *PerlCompiler {
	return NewPerlCompiler(TargetTypedPerl)
}

// NewOptimizedCleanPerlCompiler creates an optimized caching compiler for clean Perl
func NewOptimizedCleanPerlCompiler() *CachingPerlCompiler {
	return NewCachingCleanPerlCompiler(500) // Default cache size of 500 entries
}

// NewOptimizedTypedPerlCompiler creates an optimized caching compiler for typed Perl
func NewOptimizedTypedPerlCompiler() *CachingPerlCompiler {
	return NewCachingTypedPerlCompiler(500) // Default cache size of 500 entries
}

// Target returns the compilation target
func (c *PerlCompiler) Target() Target {
	return c.target
}

// Validate checks if the AST is suitable for compilation
func (c *PerlCompiler) Validate(ast AST) error {
	if ast == nil {
		return NewCompilerError(ErrInvalidAST, "AST cannot be nil")
	}

	if !ast.IsValid() {
		return NewCompilerError(ErrInvalidAST, "AST is not valid")
	}

	// Check if we can get content
	_, err := ast.GetContent()
	if err != nil {
		return NewCompilerError(ErrInvalidAST, "AST must have accessible source content").WithCause(err)
	}

	return nil
}

// Compile converts an AST to source code using CST-based transformation
func (c *PerlCompiler) Compile(ast AST) (string, error) {
	// Validate the AST first
	if err := c.Validate(ast); err != nil {
		return "", err
	}

	// Get the source content
	content, err := ast.GetContent()
	if err != nil {
		return "", NewCompilerError(ErrInvalidAST, "failed to get AST content").WithCause(err)
	}

	// Check if this is already a CST-based AST
	//nolint:gocritic // Interface to concrete type assertion is needed for CST access
	if cstAST, ok := ast.(*CSTBasedAST); ok {
		return c.compileFromCST(cstAST.GetCSTRoot(), []byte(content))
	}

	// For other AST types, we need to re-parse the content to get CST
	// This is for backward compatibility with existing AST types
	cstAST, err := NewCSTBasedAST(ast.GetPath(), content)
	if err != nil {
		return "", NewCompilerError(ErrInvalidAST, "failed to create CST from content").WithCause(err)
	}

	return c.compileFromCST(cstAST.GetCSTRoot(), []byte(content))
}

// compileFromCST performs the actual compilation using tree transformation
func (c *PerlCompiler) compileFromCST(root *sitter.Node, content []byte) (string, error) {
	if root == nil {
		return "", NewCompilerError(ErrInvalidAST, "CST root node is nil")
	}

	// Use the transformation system based on target
	var result *TransformationResult
	var err error

	switch c.target {
	case TargetCleanPerl:
		result, err = CreateCleanPerl(root, content)
	case TargetTypedPerl:
		result, err = CreateTypedPerl(root, content)
	default:
		return "", NewCompilerError(ErrUnsupportedTarget, "unsupported compilation target: "+string(c.target))
	}

	if err != nil {
		return "", NewCompilerError(ErrCompilationFailed, "compilation failed").WithCause(err)
	}

	if !result.Success {
		errorMsg := "transformation failed"
		if result.Error != nil {
			errorMsg = result.Error.Error()
		}
		return "", NewCompilerError(ErrCompilationFailed, errorMsg)
	}

	// Add version pragma for clean Perl if needed
	if c.target == TargetCleanPerl {
		finalCode, err := c.addVersionPragma(result.TransformedCode)
		if err != nil {
			return "", NewCompilerError(ErrCompilationFailed, "failed to add version pragma").WithCause(err)
		}
		return finalCode, nil
	}

	return result.TransformedCode, nil
}

// CompileString compiles Perl source code directly to the target format
func (c *PerlCompiler) CompileString(content string) (string, error) {
	// Create a CST-based AST from the content
	cstAST, err := NewCSTBasedAST("", content)
	if err != nil {
		return "", NewCompilerError(ErrInvalidAST, "failed to parse content").WithCause(err)
	}

	// Compile using the standard interface
	return c.Compile(cstAST)
}

// SetOptions updates the compiler options
func (c *PerlCompiler) SetOptions(options CompilerOptions) {
	c.options = options
}

// GetOptions returns the current compiler options
func (c *PerlCompiler) GetOptions() CompilerOptions {
	return c.options
}

// SupportsTarget returns true if the compiler can compile to the specified target
func (c *PerlCompiler) SupportsTarget(target Target) bool {
	return target == TargetCleanPerl || target == TargetTypedPerl
}

// GetSupportedTargets returns all targets this compiler supports
func (c *PerlCompiler) GetSupportedTargets() []Target {
	return []Target{TargetCleanPerl, TargetTypedPerl}
}

// addVersionPragma adds the appropriate Perl version pragma to generated code
func (c *PerlCompiler) addVersionPragma(code string) (string, error) {
	// Import the necessary packages for version handling
	// This is similar to the logic in CleanPerlCompiler.updatePerlVersion

	// Check if code already has a version pragma
	lines := strings.Split(code, "\n")
	hasVersionPragma := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "use v") || strings.HasPrefix(trimmed, "use feature") {
			hasVersionPragma = true
			break
		}
	}

	// If already has version pragma, return as-is
	if hasVersionPragma {
		return code, nil
	}

	// Add version pragma for signature support using current PVM version
	versionPragma := c.getCurrentVersionPragma()

	// Check if we need signatures specifically
	if c.hasSignatureFeatures(code) {
		// v5.36 enables signatures by default
		return versionPragma + "\n" + code, nil
	}

	// Add version pragma for other modern features
	return versionPragma + "\n" + code, nil
}

// hasSignatureFeatures checks if the code uses subroutine signatures
func (c *PerlCompiler) hasSignatureFeatures(code string) bool {
	// Simple check for signature patterns
	return strings.Contains(code, "sub ") && (strings.Contains(code, "($") || strings.Contains(code, "( $"))
}

// getCurrentVersionPragma returns the version pragma for the current PVM version
func (c *PerlCompiler) getCurrentVersionPragma() string {
	// Get current version from PVM
	currentInfo, err := current.GetCurrentVersion()
	if err != nil || !currentInfo.IsAvailable {
		// Fallback to 5.36 if we can't determine current version
		return "use v5.36;"
	}

	// Format as version pragma
	return fmt.Sprintf("use v%s;", currentInfo.Version)
}

// CreateUnifiedCompilerForTarget creates a unified compiler for any supported target
func CreateUnifiedCompilerForTarget(target Target) (Compiler, error) {
	compiler := NewPerlCompiler(target)
	if !compiler.SupportsTarget(target) {
		return nil, NewCompilerError(ErrUnsupportedTarget, "unsupported target: "+string(target))
	}
	return compiler, nil
}
