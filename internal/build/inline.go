// ABOUTME: Inline build functionality for development mode
// ABOUTME: Generates .pmc files alongside .pm files for fast development

package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tamarou.com/pvm/internal/project"
)

// InlineBuilder handles development builds that generate .pmc files
type InlineBuilder struct {
	pscBuilder *PSCBuilder
	projectCtx *project.ProjectContext
}

// NewInlineBuilder creates a new inline builder instance
func NewInlineBuilder(projectCtx *project.ProjectContext) (*InlineBuilder, error) {
	pscBuilder, err := NewPSCBuilder(projectCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to create PSC builder: %w", err)
	}

	return &InlineBuilder{
		pscBuilder: pscBuilder,
		projectCtx: projectCtx,
	}, nil
}

// InlineBuildResult represents the result of an inline build
type InlineBuildResult struct {
	Success        bool          `json:"success"`
	FilesProcessed int           `json:"files_processed"`
	PmcGenerated   int           `json:"pmc_generated"`
	TypeErrors     []TypeError   `json:"type_errors"`
	Duration       time.Duration `json:"duration"`
	BuildErrors    []string      `json:"build_errors"`
}

// Build performs an inline build, type checking files and generating .pmc files
func (b *InlineBuilder) Build(ctx context.Context, targetDirs []string) (*InlineBuildResult, error) {
	start := time.Now()
	result := &InlineBuildResult{
		Success:        true,
		FilesProcessed: 0,
		PmcGenerated:   0,
		TypeErrors:     []TypeError{},
		Duration:       0,
		BuildErrors:    []string{},
	}

	// If no target directories specified, use project lib directory
	if len(targetDirs) == 0 {
		if b.projectCtx != nil && b.projectCtx.IsProject {
			targetDirs = []string{filepath.Join(b.projectCtx.RootDir, "lib")}
		} else {
			targetDirs = []string{"lib"}
		}
	}

	// First, perform type checking on all target directories
	typeCheckResult, err := b.pscBuilder.TypeCheck(ctx, targetDirs)
	if err != nil {
		return nil, fmt.Errorf("type checking failed: %w", err)
	}

	result.TypeErrors = typeCheckResult.TypeErrors
	result.FilesProcessed = typeCheckResult.FilesChecked

	// If there are type errors, stop the build (unless in permissive mode)
	if len(result.TypeErrors) > 0 {
		result.Success = false
		result.Duration = time.Since(start)
		return result, nil
	}

	// Discover all .pm files in target directories
	pmFiles, err := b.discoverPmFiles(targetDirs)
	if err != nil {
		return nil, fmt.Errorf("failed to discover .pm files: %w", err)
	}

	// Process each .pm file to generate corresponding .pmc file
	for _, pmFile := range pmFiles {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if err := b.generatePmcFile(pmFile); err != nil {
			result.BuildErrors = append(result.BuildErrors, fmt.Sprintf("Failed to generate .pmc for %s: %v", pmFile, err))
			result.Success = false
		} else {
			result.PmcGenerated++
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// Clean removes all generated .pmc files
func (b *InlineBuilder) Clean(ctx context.Context, targetDirs []string) error {
	// If no target directories specified, use project lib directory
	if len(targetDirs) == 0 {
		if b.projectCtx != nil && b.projectCtx.IsProject {
			targetDirs = []string{filepath.Join(b.projectCtx.RootDir, "lib")}
		} else {
			targetDirs = []string{"lib"}
		}
	}

	// Discover all .pmc files in target directories
	pmcFiles, err := b.discoverPmcFiles(targetDirs)
	if err != nil {
		return fmt.Errorf("failed to discover .pmc files: %w", err)
	}

	// Remove each .pmc file
	for _, pmcFile := range pmcFiles {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := os.Remove(pmcFile); err != nil {
			// Log warning but continue with other files
			fmt.Printf("Warning: failed to remove %s: %v\n", pmcFile, err)
		}
	}

	return nil
}

// IsStale checks if a .pmc file is stale compared to its corresponding .pm file
func (b *InlineBuilder) IsStale(pmFile string) (bool, error) {
	pmcFile := b.getPmcPath(pmFile)

	// If .pmc doesn't exist, it's stale
	pmcInfo, err := os.Stat(pmcFile)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to stat .pmc file: %w", err)
	}

	// Check if .pm file is newer than .pmc file
	pmInfo, err := os.Stat(pmFile)
	if err != nil {
		return false, fmt.Errorf("failed to stat .pm file: %w", err)
	}

	return pmInfo.ModTime().After(pmcInfo.ModTime()), nil
}

// discoverPmFiles finds all .pm files in the given directories
func (b *InlineBuilder) discoverPmFiles(dirs []string) ([]string, error) {
	var pmFiles []string

	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				// Skip directories that don't exist
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}

			if !info.IsDir() && strings.HasSuffix(path, ".pm") {
				pmFiles = append(pmFiles, path)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
		}
	}

	return pmFiles, nil
}

// discoverPmcFiles finds all .pmc files in the given directories
func (b *InlineBuilder) discoverPmcFiles(dirs []string) ([]string, error) {
	var pmcFiles []string

	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				// Skip directories that don't exist
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}

			if !info.IsDir() && strings.HasSuffix(path, ".pmc") {
				pmcFiles = append(pmcFiles, path)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
		}
	}

	return pmcFiles, nil
}

// generatePmcFile creates a .pmc file from a .pm file by stripping type annotations
func (b *InlineBuilder) generatePmcFile(pmFile string) error {
	pmcFile := b.getPmcPath(pmFile)

	// Ensure the directory exists
	pmcDir := filepath.Dir(pmcFile)
	if err := os.MkdirAll(pmcDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", pmcDir, err)
	}

	// Use PSC builder to compile the file (strip type annotations)
	if err := b.pscBuilder.compileFile(pmFile, pmcFile); err != nil {
		return fmt.Errorf("failed to compile %s to %s: %w", pmFile, pmcFile, err)
	}

	return nil
}

// getPmcPath returns the .pmc file path for a given .pm file
func (b *InlineBuilder) getPmcPath(pmFile string) string {
	return strings.TrimSuffix(pmFile, ".pm") + ".pmc"
}