// ABOUTME: Project-scoped analysis tool for MCP server
// ABOUTME: Provides whole-project type analysis and cross-file type checking

package tools

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/mcp/validation"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

// ProjectAnalyzer provides project-wide analysis capabilities
type ProjectAnalyzer struct {
	codeAnalyzer *CodeAnalyzer
	validator    *validation.Validator
	parser       parser.Parser
	typeStorage  *typedef.Storage
	logger       *log.Logger
	mu           sync.RWMutex
	projectCache map[string]*ProjectAnalysis
}

// ProjectAnalysis represents the analysis results for an entire project
type ProjectAnalysis struct {
	ProjectPath   string                      `json:"project_path"`
	Files         map[string]*FileAnalysis    `json:"files"`
	GlobalTypes   map[string]*GlobalTypeInfo  `json:"global_types"`
	Dependencies  map[string][]string         `json:"dependencies"`
	TypeConflicts []TypeConflict              `json:"type_conflicts,omitempty"`
	CrossFileRefs map[string][]CrossReference `json:"cross_file_refs"`
	AnalysisTime  time.Time                   `json:"analysis_time"`
	TotalFiles    int                         `json:"total_files"`
	TotalTypes    int                         `json:"total_types"`
	TotalErrors   int                         `json:"total_errors"`
	TotalWarnings int                         `json:"total_warnings"`
}

// FileAnalysis represents analysis results for a single file
type FileAnalysis struct {
	Path         string                `json:"path"`
	Types        map[string]TypeDetail `json:"types"`
	Imports      []ImportInfo          `json:"imports"`
	Exports      []ExportInfo          `json:"exports"`
	Errors       []ErrorDetail         `json:"errors"`
	Warnings     []WarningDetail       `json:"warnings"`
	LastModified time.Time             `json:"last_modified"`
}

// GlobalTypeInfo represents type information available globally in the project
type GlobalTypeInfo struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	DefinedIn   string   `json:"defined_in"`
	Line        int      `json:"line"`
	Column      int      `json:"column"`
	ExportedAs  string   `json:"exported_as,omitempty"`
	UsedInFiles []string `json:"used_in_files"`
	IsPublic    bool     `json:"is_public"`
}

// TypeConflict represents a type conflict across files
type TypeConflict struct {
	TypeName     string       `json:"type_name"`
	Definitions  []TypeSource `json:"definitions"`
	ConflictType string       `json:"conflict_type"` // "redefinition", "incompatible", etc.
	Description  string       `json:"description"`
}

// TypeSource represents where a type is defined
type TypeSource struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Definition string `json:"definition"`
}

// CrossReference represents a cross-file reference
type CrossReference struct {
	FromFile string `json:"from_file"`
	ToFile   string `json:"to_file"`
	TypeName string `json:"type_name"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	RefKind  string `json:"ref_kind"` // "import", "use", "extends", etc.
}

// ImportInfo represents an import statement
type ImportInfo struct {
	Module  string   `json:"module"`
	Symbols []string `json:"symbols,omitempty"`
	Line    int      `json:"line"`
	Column  int      `json:"column"`
}

// ExportInfo represents an export
type ExportInfo struct {
	Symbol string `json:"symbol"`
	Type   string `json:"type,omitempty"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// NewProjectAnalyzer creates a new project analyzer
func NewProjectAnalyzer(codeAnalyzer *CodeAnalyzer, validator *validation.Validator) (*ProjectAnalyzer, error) {
	parser, err := parser.NewParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	typeStorage, err := typedef.NewStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to create type storage: %w", err)
	}

	logger := log.NewLogger(log.LevelInfo, nil, "project-analyzer")

	return &ProjectAnalyzer{
		codeAnalyzer: codeAnalyzer,
		validator:    validator,
		parser:       parser,
		typeStorage:  typeStorage,
		logger:       logger,
		projectCache: make(map[string]*ProjectAnalysis),
	}, nil
}

// AnalyzeProject performs a complete project analysis
func (pa *ProjectAnalyzer) AnalyzeProject(ctx context.Context, projectPath string) (*ProjectAnalysis, error) {
	pa.mu.RLock()
	if cached, exists := pa.projectCache[projectPath]; exists {
		// Check if cache is still fresh (less than 5 minutes old)
		if time.Since(cached.AnalysisTime) < 5*time.Minute {
			pa.mu.RUnlock()
			return cached, nil
		}
	}
	pa.mu.RUnlock()

	pa.logger.Infof("Starting project analysis for: %s", projectPath)

	analysis := &ProjectAnalysis{
		ProjectPath:   projectPath,
		Files:         make(map[string]*FileAnalysis),
		GlobalTypes:   make(map[string]*GlobalTypeInfo),
		Dependencies:  make(map[string][]string),
		TypeConflicts: []TypeConflict{},
		CrossFileRefs: make(map[string][]CrossReference),
		AnalysisTime:  time.Now(),
	}

	// Find all Perl files in the project
	perlFiles, err := pa.findPerlFiles(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find Perl files: %w", err)
	}

	analysis.TotalFiles = len(perlFiles)
	pa.logger.Infof("Found %d Perl files to analyze", len(perlFiles))

	// Analyze each file
	var wg sync.WaitGroup
	fileChan := make(chan string, len(perlFiles))
	resultChan := make(chan *FileAnalysis, len(perlFiles))
	errorChan := make(chan error, len(perlFiles))

	// Worker pool for parallel file analysis
	numWorkers := 4
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				fileAnalysis, err := pa.analyzeFile(ctx, filePath)
				if err != nil {
					errorChan <- fmt.Errorf("failed to analyze %s: %w", filePath, err)
					continue
				}
				resultChan <- fileAnalysis
			}
		}()
	}

	// Queue files for analysis
	for _, file := range perlFiles {
		fileChan <- file
	}
	close(fileChan)

	// Wait for workers to complete
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Collect results
	for fileAnalysis := range resultChan {
		analysis.Files[fileAnalysis.Path] = fileAnalysis
		analysis.TotalErrors += len(fileAnalysis.Errors)
		analysis.TotalWarnings += len(fileAnalysis.Warnings)

		// Collect global types
		for name, typeDetail := range fileAnalysis.Types {
			globalType := &GlobalTypeInfo{
				Name:        name,
				Type:        typeDetail.Type,
				DefinedIn:   fileAnalysis.Path,
				Line:        typeDetail.Line,
				Column:      typeDetail.Column,
				UsedInFiles: []string{fileAnalysis.Path},
				IsPublic:    pa.isPublicType(name),
			}

			// Check for exports
			for _, export := range fileAnalysis.Exports {
				if export.Symbol == name {
					globalType.ExportedAs = export.Symbol
					break
				}
			}

			analysis.GlobalTypes[name] = globalType
		}
	}

	// Collect errors
	for err := range errorChan {
		pa.logger.Warningf("File analysis error: %v", err)
	}

	// Perform cross-file analysis
	pa.analyzeCrossFileReferences(analysis)
	pa.detectTypeConflicts(analysis)
	pa.buildDependencyGraph(analysis)

	analysis.TotalTypes = len(analysis.GlobalTypes)

	// Cache the results
	pa.mu.Lock()
	pa.projectCache[projectPath] = analysis
	pa.mu.Unlock()

	pa.logger.Infof("Project analysis complete: %d files, %d types, %d errors, %d warnings",
		analysis.TotalFiles, analysis.TotalTypes, analysis.TotalErrors, analysis.TotalWarnings)

	return analysis, nil
}

// analyzeFile analyzes a single Perl file
func (pa *ProjectAnalyzer) analyzeFile(ctx context.Context, filePath string) (*FileAnalysis, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Get file info for last modified time
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	fileAnalysis := &FileAnalysis{
		Path:         filePath,
		Types:        make(map[string]TypeDetail),
		Imports:      []ImportInfo{},
		Exports:      []ExportInfo{},
		Errors:       []ErrorDetail{},
		Warnings:     []WarningDetail{},
		LastModified: info.ModTime(),
	}

	// Use code analyzer to get type information
	analysisResult, err := pa.codeAnalyzer.Analyze(ctx, string(content), "get_types", filepath.Dir(filePath), false)
	if err != nil {
		return nil, err
	}

	// Copy type information
	for name, typeDetail := range analysisResult.TypeInfo {
		fileAnalysis.Types[name] = typeDetail
	}

	// Check for errors
	errorResult, err := pa.codeAnalyzer.Analyze(ctx, string(content), "check_errors", filepath.Dir(filePath), false)
	if err == nil {
		fileAnalysis.Errors = errorResult.Errors
		fileAnalysis.Warnings = errorResult.Warnings
	}

	// Extract imports and exports
	pa.extractImportsExports(string(content), fileAnalysis)

	return fileAnalysis, nil
}

// findPerlFiles finds all Perl files in a project directory
func (pa *ProjectAnalyzer) findPerlFiles(projectPath string) ([]string, error) {
	var perlFiles []string

	err := filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			// Skip common non-Perl directories
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == ".svn" ||
				name == "blib" || name == "_build" || name == "cover_db" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check for Perl file extensions
		ext := filepath.Ext(path)
		if ext == ".pl" || ext == ".pm" || ext == ".t" || ext == ".pod" {
			perlFiles = append(perlFiles, path)
		}

		return nil
	})

	return perlFiles, err
}

// extractImportsExports extracts import and export information from Perl code
func (pa *ProjectAnalyzer) extractImportsExports(code string, fileAnalysis *FileAnalysis) {
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for use statements
		if strings.HasPrefix(trimmed, "use ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				module := strings.TrimSuffix(parts[1], ";")
				importInfo := ImportInfo{
					Module: module,
					Line:   i + 1,
					Column: strings.Index(line, "use") + 1,
				}

				// Check for imported symbols
				if strings.Contains(line, "qw(") || strings.Contains(line, "qw/") {
					// Extract symbols from qw() or qw//
					start := strings.Index(line, "qw")
					if start != -1 {
						symbolsStr := line[start:]
						// Simple extraction - could be improved
						if qwStart := strings.IndexAny(symbolsStr, "(/"); qwStart != -1 {
							if qwEnd := strings.IndexAny(symbolsStr[qwStart+1:], ")/"); qwEnd != -1 {
								symbols := strings.Fields(symbolsStr[qwStart+1 : qwStart+1+qwEnd])
								importInfo.Symbols = symbols
							}
						}
					}
				}

				fileAnalysis.Imports = append(fileAnalysis.Imports, importInfo)
			}
		}

		// Check for @EXPORT or @EXPORT_OK
		if strings.Contains(trimmed, "@EXPORT") || strings.Contains(trimmed, "@EXPORT_OK") {
			// Extract exported symbols
			if strings.Contains(line, "qw(") || strings.Contains(line, "qw/") {
				start := strings.Index(line, "qw")
				if start != -1 {
					symbolsStr := line[start:]
					if qwStart := strings.IndexAny(symbolsStr, "(/"); qwStart != -1 {
						if qwEnd := strings.IndexAny(symbolsStr[qwStart+1:], ")/"); qwEnd != -1 {
							symbols := strings.Fields(symbolsStr[qwStart+1 : qwStart+1+qwEnd])
							for _, symbol := range symbols {
								fileAnalysis.Exports = append(fileAnalysis.Exports, ExportInfo{
									Symbol: symbol,
									Line:   i + 1,
									Column: strings.Index(line, symbol) + 1,
								})
							}
						}
					}
				}
			}
		}
	}
}

// analyzeCrossFileReferences analyzes references between files
func (pa *ProjectAnalyzer) analyzeCrossFileReferences(analysis *ProjectAnalysis) {
	for filePath, fileAnalysis := range analysis.Files {
		for _, imp := range fileAnalysis.Imports {
			// Find which file exports the imported module/symbols
			for otherPath, otherAnalysis := range analysis.Files {
				if filePath == otherPath {
					continue
				}

				// Check if the other file exports any of the imported symbols
				for _, symbol := range imp.Symbols {
					for _, export := range otherAnalysis.Exports {
						if export.Symbol == symbol {
							ref := CrossReference{
								FromFile: filePath,
								ToFile:   otherPath,
								TypeName: symbol,
								Line:     imp.Line,
								Column:   imp.Column,
								RefKind:  "import",
							}
							analysis.CrossFileRefs[filePath] = append(analysis.CrossFileRefs[filePath], ref)

							// Update global type usage
							if globalType, exists := analysis.GlobalTypes[symbol]; exists {
								globalType.UsedInFiles = append(globalType.UsedInFiles, filePath)
							}
						}
					}
				}
			}
		}
	}
}

// detectTypeConflicts detects type conflicts across files
func (pa *ProjectAnalyzer) detectTypeConflicts(analysis *ProjectAnalysis) {
	typeDefinitions := make(map[string][]TypeSource)

	// Collect all type definitions
	for filePath, fileAnalysis := range analysis.Files {
		for typeName, typeDetail := range fileAnalysis.Types {
			source := TypeSource{
				File:       filePath,
				Line:       typeDetail.Line,
				Column:     typeDetail.Column,
				Definition: typeDetail.Type,
			}
			typeDefinitions[typeName] = append(typeDefinitions[typeName], source)
		}
	}

	// Check for conflicts
	for typeName, sources := range typeDefinitions {
		if len(sources) > 1 {
			// Check if definitions are compatible
			firstDef := sources[0].Definition
			conflicting := false

			for i := 1; i < len(sources); i++ {
				if sources[i].Definition != firstDef {
					conflicting = true
					break
				}
			}

			if conflicting {
				conflict := TypeConflict{
					TypeName:     typeName,
					Definitions:  sources,
					ConflictType: "incompatible",
					Description:  fmt.Sprintf("Type '%s' has incompatible definitions across files", typeName),
				}
				analysis.TypeConflicts = append(analysis.TypeConflicts, conflict)
			} else if len(sources) > 1 {
				// Same definition in multiple files - might be intentional but worth noting
				conflict := TypeConflict{
					TypeName:     typeName,
					Definitions:  sources,
					ConflictType: "redefinition",
					Description:  fmt.Sprintf("Type '%s' is defined in multiple files with the same definition", typeName),
				}
				analysis.TypeConflicts = append(analysis.TypeConflicts, conflict)
			}
		}
	}
}

// buildDependencyGraph builds a dependency graph for the project
func (pa *ProjectAnalyzer) buildDependencyGraph(analysis *ProjectAnalysis) {
	for filePath, fileAnalysis := range analysis.Files {
		dependencies := make(map[string]bool)

		// Add dependencies from imports
		for _, imp := range fileAnalysis.Imports {
			// Try to resolve module to file
			for otherPath := range analysis.Files {
				if filePath != otherPath && pa.moduleMatchesFile(imp.Module, otherPath) {
					dependencies[otherPath] = true
				}
			}
		}

		// Add dependencies from cross-file references
		if refs, exists := analysis.CrossFileRefs[filePath]; exists {
			for _, ref := range refs {
				dependencies[ref.ToFile] = true
			}
		}

		// Convert to slice
		var deps []string
		for dep := range dependencies {
			deps = append(deps, dep)
		}

		analysis.Dependencies[filePath] = deps
	}
}

// moduleMatchesFile checks if a module name matches a file path
func (pa *ProjectAnalyzer) moduleMatchesFile(module string, filePath string) bool {
	// Convert module name to potential file path
	// e.g., "Foo::Bar" -> "Foo/Bar.pm"
	potentialPath := strings.ReplaceAll(module, "::", "/") + ".pm"

	// Check if the file path ends with the potential path
	return strings.HasSuffix(filePath, potentialPath)
}

// isPublicType checks if a type name represents a public type
func (pa *ProjectAnalyzer) isPublicType(typeName string) bool {
	// In Perl, types starting with uppercase are typically public
	if len(typeName) > 0 && typeName[0] >= 'A' && typeName[0] <= 'Z' {
		return true
	}
	return false
}

// GetProjectSummary returns a summary of the project analysis
func (pa *ProjectAnalyzer) GetProjectSummary(projectPath string) (*ProjectSummary, error) {
	pa.mu.RLock()
	analysis, exists := pa.projectCache[projectPath]
	pa.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no analysis found for project: %s", projectPath)
	}

	summary := &ProjectSummary{
		ProjectPath:     analysis.ProjectPath,
		TotalFiles:      analysis.TotalFiles,
		TotalTypes:      analysis.TotalTypes,
		TotalErrors:     analysis.TotalErrors,
		TotalWarnings:   analysis.TotalWarnings,
		TypeConflicts:   len(analysis.TypeConflicts),
		AnalysisTime:    analysis.AnalysisTime,
		FilesWithErrors: pa.countFilesWithErrors(analysis),
		PublicTypes:     pa.countPublicTypes(analysis),
		CrossFileRefs:   pa.countCrossFileRefs(analysis),
	}

	return summary, nil
}

// ProjectSummary provides a high-level summary of project analysis
type ProjectSummary struct {
	ProjectPath     string    `json:"project_path"`
	TotalFiles      int       `json:"total_files"`
	TotalTypes      int       `json:"total_types"`
	TotalErrors     int       `json:"total_errors"`
	TotalWarnings   int       `json:"total_warnings"`
	TypeConflicts   int       `json:"type_conflicts"`
	FilesWithErrors int       `json:"files_with_errors"`
	PublicTypes     int       `json:"public_types"`
	CrossFileRefs   int       `json:"cross_file_refs"`
	AnalysisTime    time.Time `json:"analysis_time"`
}

// Helper methods for summary

func (pa *ProjectAnalyzer) countFilesWithErrors(analysis *ProjectAnalysis) int {
	count := 0
	for _, fileAnalysis := range analysis.Files {
		if len(fileAnalysis.Errors) > 0 {
			count++
		}
	}
	return count
}

func (pa *ProjectAnalyzer) countPublicTypes(analysis *ProjectAnalysis) int {
	count := 0
	for _, globalType := range analysis.GlobalTypes {
		if globalType.IsPublic {
			count++
		}
	}
	return count
}

func (pa *ProjectAnalyzer) countCrossFileRefs(analysis *ProjectAnalysis) int {
	count := 0
	for _, refs := range analysis.CrossFileRefs {
		count += len(refs)
	}
	return count
}
