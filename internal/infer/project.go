// ABOUTME: Cross-file analysis index for the PSC type inference engine.
// ABOUTME: Manages module resolution, per-file analysis caching, and cross-file symbol lookup.

package infer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/types"
)

// ProjectIndex manages cross-file analysis for a Perl project rooted at a
// single directory. It resolves module names to file paths, caches per-file
// analysis results, and exposes cross-file symbol lookup.
type ProjectIndex struct {
	root    string
	libDirs []string
	cache   map[string]*FileAnalysis
	mu      sync.RWMutex
}

// FileAnalysis holds the results of analysing a single .pm file.
type FileAnalysis struct {
	Annotations map[uint32]types.Type
	Diagnostics []Diagnostic
	Symbols     *SymbolTable
	Package     string
}

// NewProjectIndex creates a ProjectIndex rooted at root with the default
// library search path of ["lib"].
func NewProjectIndex(root string) *ProjectIndex {
	return &ProjectIndex{
		root:    root,
		libDirs: []string{"lib"},
		cache:   make(map[string]*FileAnalysis),
	}
}

// ResolveModule converts a Perl module name (e.g. "Foo::Bar") to an absolute
// file path by searching each entry in libDirs for a matching .pm file.
// Returns an error if no matching file is found in any lib directory.
func (idx *ProjectIndex) ResolveModule(name string) (string, error) {
	rel := strings.ReplaceAll(name, "::", string(filepath.Separator)) + ".pm"
	for _, dir := range idx.libDirs {
		candidate := filepath.Join(idx.root, dir, rel)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("module %q not found in %v", name, idx.libDirs)
}

// AnalyzeFile parses and analyses the .pm file at path, caches the result,
// and returns it. Subsequent calls with the same path return the cached
// *FileAnalysis without re-parsing.
func (idx *ProjectIndex) AnalyzeFile(path string) (*FileAnalysis, error) {
	// Fast path: check the cache under a read lock.
	idx.mu.RLock()
	if cached, ok := idx.cache[path]; ok {
		idx.mu.RUnlock()
		return cached, nil
	}
	idx.mu.RUnlock()

	// Read and parse the file.
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("AnalyzeFile: read %q: %w", path, err)
	}

	p := parser.New()
	tree, err := p.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("AnalyzeFile: parse %q: %w", path, err)
	}

	annotations, diags, st := Analyze(tree, source)

	result := &FileAnalysis{
		Annotations: annotations,
		Diagnostics: diags,
		Symbols:     st,
		Package:     st.CurrentPackage(),
	}

	// Store in cache under a write lock.
	idx.mu.Lock()
	idx.cache[path] = result
	idx.mu.Unlock()

	return result, nil
}

// LookupSymbol resolves the module for pkg, analyses the corresponding file,
// and searches its symbol table for name. Returns the Symbol and true if
// found, or the zero Symbol and false otherwise.
func (idx *ProjectIndex) LookupSymbol(pkg, name string) (Symbol, bool) {
	path, err := idx.ResolveModule(pkg)
	if err != nil {
		return Symbol{}, false
	}

	result, err := idx.AnalyzeFile(path)
	if err != nil {
		return Symbol{}, false
	}

	return result.Symbols.Lookup(name)
}

// Prefetch walks every lib directory recursively, analyses all .pm files it
// finds, and populates the cache. This allows later LookupSymbol calls to
// return cached results without any I/O.
func (idx *ProjectIndex) Prefetch() {
	for _, dir := range idx.libDirs {
		root := filepath.Join(idx.root, dir)
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".pm") {
				// Errors are intentionally swallowed; Prefetch is best-effort.
				_, _ = idx.AnalyzeFile(path)
			}
			return nil
		})
	}
}
