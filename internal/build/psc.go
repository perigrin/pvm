// ABOUTME: PSC integration infrastructure for build system
// ABOUTME: Provides type checking and compilation services

package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tamarou.com/pvm/internal/compiler"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/project"
	"tamarou.com/pvm/internal/typechecker"
)

// TypeError represents a structured type error from PSC
type TypeError struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Level   string `json:"level"` // "error", "warning", "info"
}

// PSCResult represents the result of PSC operations
type PSCResult struct {
	Success      bool          `json:"success"`
	TypeErrors   []TypeError   `json:"type_errors"`
	FilesChecked int           `json:"files_checked"`
	Duration     time.Duration `json:"duration"`
}

// PSCBuilder provides PSC integration for the build system
type PSCBuilder struct {
	projectContext *project.ProjectContext
	registry       *compiler.CompilerRegistry
	typeChecker    *typechecker.TypeCheck
}

// NewPSCBuilder creates a new PSC builder instance
func NewPSCBuilder(projectCtx *project.ProjectContext) (*PSCBuilder, error) {
	// Create type checker
	tc, err := typechecker.NewTypeCheck()
	if err != nil {
		return nil, fmt.Errorf("failed to create type checker: %w", err)
	}

	// Create compiler registry
	registry := compiler.NewCompilerRegistry()

	return &PSCBuilder{
		projectContext: projectCtx,
		registry:       registry,
		typeChecker:    tc,
	}, nil
}

// TypeCheck performs type checking on the specified files or directories
func (b *PSCBuilder) TypeCheck(ctx context.Context, paths []string) (*PSCResult, error) {
	start := time.Now()
	result := &PSCResult{
		Success:      true,
		TypeErrors:   []TypeError{},
		FilesChecked: 0,
		Duration:     0,
	}

	// Discover files to check
	files, err := b.discoverFiles(paths)
	if err != nil {
		return nil, fmt.Errorf("failed to discover files: %w", err)
	}

	// Check each file
	for _, file := range files {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		errors, err := b.typeCheckFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to type check %s: %w", file, err)
		}

		result.FilesChecked++
		result.TypeErrors = append(result.TypeErrors, errors...)
	}

	result.Duration = time.Since(start)
	result.Success = len(result.TypeErrors) == 0

	return result, nil
}

// Compile strips type annotations from input files and writes to output directory
func (b *PSCBuilder) Compile(ctx context.Context, inputDir, outputDir string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Discover Perl files in input directory
	files, err := b.discoverFiles([]string{inputDir})
	if err != nil {
		return fmt.Errorf("failed to discover files: %w", err)
	}

	// Process each file
	for _, file := range files {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Calculate relative path and output path
		relPath, err := filepath.Rel(inputDir, file)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path for %s: %w", file, err)
		}

		outputFile := filepath.Join(outputDir, relPath)
		outputFileDir := filepath.Dir(outputFile)

		// Ensure output directory exists
		if err := os.MkdirAll(outputFileDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory %s: %w", outputFileDir, err)
		}

		// Compile the file
		if err := b.compileFile(file, outputFile); err != nil {
			return fmt.Errorf("failed to compile %s: %w", file, err)
		}
	}

	return nil
}

// Watch monitors directories for changes and triggers callbacks
func (b *PSCBuilder) Watch(ctx context.Context, dirs []string, callback func(string, error)) error {
	// Note: This is a basic implementation. A production implementation would use
	// a proper file system watcher like fsnotify

	// For now, we'll implement a simple polling-based watcher
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Track file modification times
	fileModTimes := make(map[string]time.Time)

	// Initialize with current file states
	for _, dir := range dirs {
		files, err := b.discoverFiles([]string{dir})
		if err != nil {
			callback("", fmt.Errorf("failed to discover files in %s: %w", dir, err))
			continue
		}

		for _, file := range files {
			info, err := os.Stat(file)
			if err != nil {
				continue
			}
			fileModTimes[file] = info.ModTime()
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Check for file changes
			for _, dir := range dirs {
				files, err := b.discoverFiles([]string{dir})
				if err != nil {
					callback("", fmt.Errorf("failed to discover files in %s: %w", dir, err))
					continue
				}

				for _, file := range files {
					info, err := os.Stat(file)
					if err != nil {
						continue
					}

					lastModTime, exists := fileModTimes[file]
					if !exists || info.ModTime().After(lastModTime) {
						fileModTimes[file] = info.ModTime()
						callback(file, nil)
					}
				}
			}
		}
	}
}

// GetTypeErrors returns the last set of type errors (for integration with other tools)
func (b *PSCBuilder) GetTypeErrors() []TypeError {
	// This would typically store the last set of errors from TypeCheck
	// For now, return empty slice as this is a placeholder
	return []TypeError{}
}

// discoverFiles finds all Perl files in the given paths
func (b *PSCBuilder) discoverFiles(paths []string) ([]string, error) {
	var files []string

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to stat %s: %w", path, err)
		}

		if info.IsDir() {
			err := filepath.Walk(path, func(walkPath string, walkInfo os.FileInfo, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}

				if !walkInfo.IsDir() && b.isPerlFile(walkPath) {
					files = append(files, walkPath)
				}

				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("failed to walk directory %s: %w", path, err)
			}
		} else if b.isPerlFile(path) {
			files = append(files, path)
		}
	}

	return files, nil
}

// isPerlFile checks if a file is a Perl file based on extension
func (b *PSCBuilder) isPerlFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".pl" || ext == ".pm" || ext == ".t"
}

// typeCheckFile performs type checking on a single file
func (b *PSCBuilder) typeCheckFile(filePath string) ([]TypeError, error) {
	result, err := b.typeChecker.CheckFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("type checking failed: %w", err)
	}

	var typeErrors []TypeError
	for _, err := range result.Errors {
		typeErrors = append(typeErrors, TypeError{
			File:    filePath,
			Line:    err.Line,
			Column:  err.Column,
			Message: err.Message,
			Code:    "type_error", // Default code since TypeCheckError doesn't have Code field
			Level:   "error",      // Default to error level
		})
	}

	return typeErrors, nil
}

// compileFile strips type annotations from a file and writes the result
func (b *PSCBuilder) compileFile(inputFile, outputFile string) error {
	// Parse the input file
	standardParser, err := parser.NewParser()
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	ast, err := standardParser.ParseFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Get clean Perl compiler
	compiler, exists := b.registry.GetCompiler(compiler.TargetCleanPerl)
	if !exists {
		return fmt.Errorf("clean Perl compiler not available")
	}

	// Compile to clean Perl
	result, err := compiler.Compile(ast)
	if err != nil {
		return fmt.Errorf("failed to compile: %w", err)
	}

	// Write the result
	if err := os.WriteFile(outputFile, []byte(result), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
