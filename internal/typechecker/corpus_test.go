// ABOUTME: Corpus-based tests for TypeChecker using intentional test cases
// ABOUTME: Tests type checker against hand-written expected outputs in corpus files

package typechecker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/parser/treesitter"
	"tamarou.com/pvm/internal/typedef"
)

func TestTypeChecker_Corpus(t *testing.T) {
	// Load typechecker corpus files
	corpusDir := "../../testdata/corpus/typechecker"
	files, err := filepath.Glob(filepath.Join(corpusDir, "*.md"))
	if err != nil {
		t.Fatalf("Failed to find corpus files: %v", err)
	}

	if len(files) == 0 {
		t.Skip("No typechecker corpus files found")
	}

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".md")
		t.Run(name, func(t *testing.T) {
			runTypeCheckerCorpusTest(t, file)
		})
	}
}

func runTypeCheckerCorpusTest(t *testing.T, filepath string) {
	t.Helper()

	// Read the corpus file
	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read corpus file: %v", err)
	}

	// Extract input code and expected output
	inputCode := extractCodeBlock(string(content), "perl")
	expectedOutput := extractExpectedSymbolTable(string(content))

	if inputCode == "" {
		t.Fatalf("No perl code block found in corpus file")
	}
	if expectedOutput == "" {
		t.Fatalf("No expected symbol table found in corpus file")
	}

	// Parse the input code with both parsers
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString(inputCode)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	// Parse with tree-sitter for CST
	tsParser := sitter.NewParser()
	tsParser.SetLanguage(treesitter.Language())
	contentBytes := []byte(inputCode)
	tree := tsParser.Parse(contentBytes, nil)
	if tree == nil {
		t.Fatalf("Failed to parse with tree-sitter")
	}

	// Bind symbols using CST
	b := binder.NewBinder()
	symbolTable, err := b.BindCST(tree.RootNode(), contentBytes, ast.TypeAnnotations)
	if err != nil {
		t.Fatalf("Failed to bind symbols: %v", err)
	}

	// Type check
	store, _ := typedef.NewStorage()
	hierarchy := typedef.NewTypeHierarchy(store)
	checker := NewTypeChecker(hierarchy, symbolTable, "test_module")
	errors := checker.CheckAST(ast)

	// Format actual results
	var result strings.Builder
	result.WriteString("=== SYMBOLS ===\n")
	symbols := symbolTable.GetVisibleSymbols()
	for _, symbol := range symbols {
		result.WriteString(symbol.String() + "\n")
	}

	result.WriteString("=== TYPE ERRORS ===\n")
	if len(errors) == 0 {
		result.WriteString("No type errors\n")
	} else {
		for _, err := range errors {
			result.WriteString(err.Error() + "\n")
		}
	}

	actualOutput := strings.TrimSpace(result.String())

	// Compare expected vs actual
	if actualOutput != expectedOutput {
		t.Errorf("Symbol table mismatch:\n--- Expected ---\n%s\n--- Actual ---\n%s\n--- End ---",
			expectedOutput, actualOutput)
	}
}

// extractExpectedSymbolTable extracts the expected symbol table from corpus markdown
func extractExpectedSymbolTable(content string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder
	inSymbolSection := false
	inCodeBlock := false

	for _, line := range lines {
		// Look for the "Expected Symbol Table" section
		if strings.Contains(line, "# Expected Symbol Table") {
			inSymbolSection = true
			continue
		}

		// Stop at next section
		if inSymbolSection && strings.HasPrefix(line, "# ") &&
			!strings.Contains(line, "Expected Symbol Table") {
			break
		}

		if inSymbolSection {
			// Start of code block
			if strings.HasPrefix(line, "```") && !inCodeBlock {
				inCodeBlock = true
				continue
			}
			// End of code block
			if strings.HasPrefix(line, "```") && inCodeBlock {
				break
			}
			// Content inside code block
			if inCodeBlock {
				result.WriteString(line + "\n")
			}
		}
	}

	return strings.TrimSpace(result.String())
}

// extractCodeBlock extracts code from the first code block with specified language
func extractCodeBlock(content, language string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder
	inCodeBlock := false
	targetBlock := "```" + language

	for _, line := range lines {
		// Start of target code block
		if strings.HasPrefix(line, targetBlock) {
			inCodeBlock = true
			continue
		}
		// End of any code block
		if strings.HasPrefix(line, "```") && inCodeBlock {
			break
		}
		// Content inside target code block
		if inCodeBlock {
			result.WriteString(line + "\n")
		}
	}

	return strings.TrimSpace(result.String())
}
