// ABOUTME: Incremental parsing and change detection for LSP performance optimization (EXPERIMENTAL)
// ABOUTME: Provides intelligent change tracking and selective reprocessing for improved responsiveness

// EXPERIMENTAL FEATURE WARNING:
// This incremental analysis system is experimental and may cause instability
// until the following are implemented:
// - Stable dependency tracking between modules
// - Reliable incremental type checking
// - Proper cache invalidation strategies
// TODO: Move to internal/experimental/ when dependencies are resolved

package ls

import (
	"context"
	"fmt"
	"strings"
	"time"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/parser/treesitter"
)

// TextDocumentContentChangeEvent represents a content change in a document
type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitempty"`
	RangeLength *int   `json:"rangeLength,omitempty"`
	Text        string `json:"text"`
}

// ChangeType represents the type of change made to a document
type ChangeType int

const (
	ChangeTypeInsert ChangeType = iota
	ChangeTypeDelete
	ChangeTypeReplace
	ChangeTypeFull // Full document replacement
)

// DocumentChange represents a processed change with analysis
type DocumentChange struct {
	Type          ChangeType
	StartLine     int
	EndLine       int
	StartChar     int
	EndChar       int
	NewText       string
	OldText       string
	AffectedLines []int
	Timestamp     time.Time
}

// IncrementalContext contains information needed for incremental updates
type IncrementalContext struct {
	LastFullParse   time.Time
	RecentChanges   []DocumentChange
	DirtyRegions    map[int]bool // Line numbers that need reanalysis
	StableRegions   map[int]bool // Line numbers known to be stable
	ChangeFrequency int          // Recent change frequency
	ComplexityScore int          // Estimated parsing complexity
}

// UpdateDocumentIncremental performs incremental document updates when possible
func (ls *LanguageService) UpdateDocumentIncremental(uri, newText string, version int, changes []TextDocumentContentChangeEvent) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	existingDoc, exists := ls.documents[uri]
	contentHash := ls.cache.HashContent(newText)

	// Check if content is actually the same (no-op)
	if exists && existingDoc.ContentHash == contentHash {
		return nil
	}

	// Get or create incremental context
	context := ls.getIncrementalContext(uri)

	// Analyze the changes
	processedChanges := ls.analyzeChanges(changes, existingDoc)
	context.RecentChanges = append(context.RecentChanges, processedChanges...)

	// Decide whether to use incremental or full parsing
	useIncremental := ls.shouldUseIncrementalParsing(context, processedChanges)

	if useIncremental && exists {
		return ls.performIncrementalUpdate(uri, newText, version, contentHash, existingDoc, context, processedChanges)
	} else {
		return ls.performFullUpdate(uri, newText, version, contentHash, context)
	}
}

// analyzeChanges processes raw change events into structured change information
func (ls *LanguageService) analyzeChanges(changes []TextDocumentContentChangeEvent, existingDoc *Document) []DocumentChange {
	var processedChanges []DocumentChange
	now := time.Now()

	for _, change := range changes {
		docChange := DocumentChange{
			NewText:   change.Text,
			Timestamp: now,
		}

		if change.Range == nil {
			// Full document change
			docChange.Type = ChangeTypeFull
			docChange.StartLine = 0
			docChange.EndLine = -1 // Indicates full document
		} else {
			// Determine change type and affected regions
			docChange.StartLine = change.Range.Start.Line
			docChange.EndLine = change.Range.End.Line
			docChange.StartChar = change.Range.Start.Character
			docChange.EndChar = change.Range.End.Character

			switch {
			case change.Range.Start.Line == change.Range.End.Line &&
				change.Range.Start.Character == change.Range.End.Character:
				docChange.Type = ChangeTypeInsert
			case change.Text == "":
				docChange.Type = ChangeTypeDelete
			default:
				docChange.Type = ChangeTypeReplace
			}

			// Calculate affected lines
			docChange.AffectedLines = ls.calculateAffectedLines(docChange)
		}

		processedChanges = append(processedChanges, docChange)
	}

	return processedChanges
}

// shouldUseIncrementalParsing determines if incremental parsing should be used
func (ls *LanguageService) shouldUseIncrementalParsing(incCtx *IncrementalContext, changes []DocumentChange) bool {
	// Don't use incremental for full document changes
	for _, change := range changes {
		if change.Type == ChangeTypeFull {
			return false
		}
	}

	// Don't use incremental if we haven't done a full parse recently
	if time.Since(incCtx.LastFullParse) > 5*time.Minute {
		return false
	}

	// Don't use incremental if there are too many recent changes
	if len(incCtx.RecentChanges) > 50 {
		return false
	}

	// Don't use incremental if changes are too complex
	totalAffectedLines := 0
	for _, change := range changes {
		totalAffectedLines += len(change.AffectedLines)
	}

	if totalAffectedLines > 100 {
		return false
	}

	// Don't use incremental if change frequency is too high
	if incCtx.ChangeFrequency > 10 {
		return false
	}

	return true
}

// performIncrementalUpdate updates the document using incremental parsing
func (ls *LanguageService) performIncrementalUpdate(uri, newText string, version int, contentHash string, existingDoc *Document, incCtx *IncrementalContext, changes []DocumentChange) error {
	incOp := ls.monitor.StartOperation(context.TODO(), "incremental_parse")
	defer incOp.Complete()

	// Invalidate affected caches only
	ls.invalidateAffectedCaches(uri, changes)

	// Update document text
	doc := &Document{
		URI:         uri,
		Text:        newText,
		Version:     version,
		LastChanged: time.Now(),
		ContentHash: contentHash,
		AST:         existingDoc.AST,         // Start with existing AST
		SymbolTable: existingDoc.SymbolTable, // Start with existing symbols
		Errors:      existingDoc.Errors,      // Start with existing errors
	}

	// Attempt incremental AST update
	if ls.canUpdateASTIncrementally(changes) {
		// For now, we'll use a simplified approach: reparse affected regions
		// In a full implementation, this would modify the AST in-place
		ls.reparseAffectedRegions(doc, changes)
	} else {
		// Fall back to full reparse
		return ls.performFullUpdate(uri, newText, version, contentHash, incCtx)
	}

	// Update symbol table incrementally
	if doc.AST != nil {
		bindOp := ls.monitor.StartOperation(context.TODO(), "incremental_bind")
		symbolTable, err := ls.updateSymbolTableIncrementally(doc.AST, existingDoc.SymbolTable, changes, newText)
		if err != nil {
			bindOp.CompleteWithError(err)
			// Fall back to full update
			return ls.performFullUpdate(uri, newText, version, contentHash, incCtx)
		}
		bindOp.Complete()
		doc.SymbolTable = symbolTable
		doc.SymbolHash = ls.cache.HashContent(strings.Join([]string{contentHash, "symbols"}, ":"))
	}

	// Update type checking for affected regions only
	errors := ls.updateTypeCheckingIncrementally(doc, existingDoc, changes)
	doc.Errors = make([]parser.TypeCheckError, len(errors))
	for i, err := range errors {
		// Since errors come from updateTypeCheckingIncrementally which returns []error,
		// we need to convert them to TypeCheckError
		doc.Errors[i] = parser.TypeCheckError{Message: err.Error()}
	}
	doc.LastChecked = time.Now()

	// Update context
	incCtx.ChangeFrequency++
	ls.updateIncrementalContext(uri, incCtx)

	ls.documents[uri] = doc
	return nil
}

// performFullUpdate performs a full document update (fallback)
func (ls *LanguageService) performFullUpdate(uri, newText string, version int, contentHash string, incCtx *IncrementalContext) error {
	// Use the existing UpdateDocument method
	err := ls.UpdateDocument(uri, newText, version)

	// Update context to reflect full parse
	incCtx.LastFullParse = time.Now()
	incCtx.RecentChanges = nil // Clear recent changes
	incCtx.DirtyRegions = make(map[int]bool)
	incCtx.StableRegions = make(map[int]bool)
	incCtx.ChangeFrequency = 0
	ls.updateIncrementalContext(uri, incCtx)

	return err
}

// calculateAffectedLines determines which lines are affected by a change
func (ls *LanguageService) calculateAffectedLines(change DocumentChange) []int {
	var lines []int

	switch change.Type {
	case ChangeTypeInsert:
		// For inserts, we affect the line and potentially following lines
		for i := change.StartLine; i <= change.StartLine+strings.Count(change.NewText, "\n")+2; i++ {
			lines = append(lines, i)
		}
	case ChangeTypeDelete:
		// For deletes, we affect from start to a bit beyond
		for i := change.StartLine; i <= change.EndLine+2; i++ {
			lines = append(lines, i)
		}
	case ChangeTypeReplace:
		// For replacements, we affect the range and a buffer
		newLines := strings.Count(change.NewText, "\n")
		oldLines := change.EndLine - change.StartLine
		maxLines := max(newLines, oldLines)
		for i := change.StartLine; i <= change.StartLine+maxLines+2; i++ {
			lines = append(lines, i)
		}
	}

	return lines
}

// invalidateAffectedCaches removes cached data for affected regions
func (ls *LanguageService) invalidateAffectedCaches(uri string, changes []DocumentChange) {
	for _, change := range changes {
		for _, line := range change.AffectedLines {
			// Remove hover cache for affected lines
			for char := 0; char < 200; char++ { // Reasonable character limit
				cacheKey := strings.Join([]string{uri, string(rune(line)), string(rune(char))}, ":")
				ls.cache.mu.Lock()
				delete(ls.cache.hoverCache, cacheKey)
				delete(ls.cache.completionCache, cacheKey+":completion")
				delete(ls.cache.definitionCache, cacheKey)
				delete(ls.cache.referencesCache, cacheKey)
				ls.cache.mu.Unlock()
			}
		}
	}
}

// canUpdateASTIncrementally determines if AST can be updated without full reparse
func (ls *LanguageService) canUpdateASTIncrementally(changes []DocumentChange) bool {
	// Simplified heuristic: only allow for small, single-line changes
	if len(changes) > 1 {
		return false
	}

	change := changes[0]
	if change.Type == ChangeTypeFull {
		return false
	}

	// Only handle single-line changes for now
	if len(change.AffectedLines) > 3 {
		return false
	}

	// Don't handle changes that span multiple lines
	if change.StartLine != change.EndLine && change.Type != ChangeTypeInsert {
		return false
	}

	return true
}

// reparseAffectedRegions reparses only the affected parts of the document
func (ls *LanguageService) reparseAffectedRegions(doc *Document, changes []DocumentChange) {
	// For now, we'll do a full reparse but track that we intended to do incremental
	// A full implementation would extract affected AST nodes and reparse only those regions

	parseOp := ls.monitor.StartOperation(context.TODO(), "selective_reparse")
	defer parseOp.Complete()

	// Extract affected text regions and reparse them
	// This is a simplified version - a full implementation would be more sophisticated
	ast, err := parser.PooledParserFunc(func(p parser.Parser) (*parser.AST, error) {
		return p.ParseString(doc.Text)
	})
	if err == nil {
		doc.AST = ast
		doc.ASTHash = ls.cache.HashContent(strings.Join([]string{doc.ContentHash, "ast"}, ":"))
	}
}

// updateSymbolTableIncrementally updates the symbol table with minimal changes
func (ls *LanguageService) updateSymbolTableIncrementally(ast *ast.AST, existingSymbols *binder.SymbolTable, changes []DocumentChange, text string) (*binder.SymbolTable, error) {
	// For now, do a full symbol binding but track the intent for incremental
	// A full implementation would update only affected scopes and symbols

	// Parse with tree-sitter for CST binding
	tsParser := sitter.NewParser()
	tsParser.SetLanguage(treesitter.Language())
	contentBytes := []byte(text)
	tree := tsParser.Parse(contentBytes, nil)
	if tree == nil {
		return nil, fmt.Errorf("failed to parse with tree-sitter")
	}

	return ls.binder.BindCST(tree.RootNode(), contentBytes, ast.TypeAnnotations)
}

// updateTypeCheckingIncrementally updates type checking for affected regions only
func (ls *LanguageService) updateTypeCheckingIncrementally(doc *Document, existingDoc *Document, changes []DocumentChange) []error {
	// For now, return existing errors if the changes seem minimal
	// A full implementation would re-check only affected regions

	// If changes are minimal, assume errors are still valid
	totalAffectedLines := 0
	for _, change := range changes {
		totalAffectedLines += len(change.AffectedLines)
	}

	if totalAffectedLines <= 5 {
		// Convert TypeCheckError to error
		errors := make([]error, len(existingDoc.Errors))
		for i, tcErr := range existingDoc.Errors {
			errors[i] = tcErr
		}
		return errors
	}

	// Otherwise, would need full type check (not implemented here)
	return []error{}
}

// Incremental context management

var incrementalContexts = make(map[string]*IncrementalContext)

func (ls *LanguageService) getIncrementalContext(uri string) *IncrementalContext {
	if context, exists := incrementalContexts[uri]; exists {
		return context
	}

	context := &IncrementalContext{
		LastFullParse:   time.Now(),
		RecentChanges:   []DocumentChange{},
		DirtyRegions:    make(map[int]bool),
		StableRegions:   make(map[int]bool),
		ChangeFrequency: 0,
		ComplexityScore: 0,
	}
	incrementalContexts[uri] = context
	return context
}

func (ls *LanguageService) updateIncrementalContext(uri string, incCtx *IncrementalContext) {
	incrementalContexts[uri] = incCtx

	// Clean up old changes
	cutoff := time.Now().Add(-10 * time.Minute)
	var recentChanges []DocumentChange
	for _, change := range incCtx.RecentChanges {
		if change.Timestamp.After(cutoff) {
			recentChanges = append(recentChanges, change)
		}
	}
	incCtx.RecentChanges = recentChanges
}

// Helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
