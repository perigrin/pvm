// ABOUTME: Tests for advanced LSP features
// ABOUTME: Validates go-to-definition, find references, formatting, and code actions

package lsp

import (
	"io"
	"log"
	"testing"

	"tamarou.com/pvm/internal/ls"
)

func testLogger(t *testing.T) *log.Logger {
	return log.New(io.Discard, "[TEST] ", log.LstdFlags)
}

// Helper function to convert LSP position to language service position
func testConvertPosition(lspPos Position) ls.Position {
	return ls.Position{
		Line:      lspPos.Line,
		Character: lspPos.Character,
	}
}

func TestFindDefinition(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		position Position
		expected int // number of locations expected
	}{
		{
			name: "Variable definition",
			content: `my $foo = 42;
print $foo;`,
			position: Position{Line: 1, Character: 7}, // Position on $foo in print
			expected: 1,
		},
		{
			name: "Subroutine definition",
			content: `sub hello {
    print "Hello\n";
}
hello();`,
			position: Position{Line: 3, Character: 0}, // Position on hello()
			expected: 1,
		},
		{
			name:     "No definition found",
			content:  `print $undefined;`,
			position: Position{Line: 0, Character: 7}, // Position on $undefined
			expected: 0,
		},
	}

	languageService, err := ls.NewLanguageService()
	if err != nil {
		t.Fatalf("Failed to create language service: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri := "file:///test.pl"

			// Update document in language service
			err := languageService.UpdateDocument(uri, tt.content, 1)
			if err != nil {
				t.Fatalf("Failed to update document: %v", err)
			}

			// Convert position to language service position
			lsPos := testConvertPosition(tt.position)

			// Get definition from language service
			definition, err := languageService.GetDefinition(uri, lsPos)
			if err != nil {
				t.Fatalf("Failed to get definition: %v", err)
			}

			actualCount := 0
			if definition != nil {
				actualCount = 1
			}

			if actualCount != tt.expected {
				t.Errorf("Expected %d locations, got %d", tt.expected, actualCount)
				t.Logf("Definition result: %+v", definition)
				t.Logf("Document content: %q", tt.content)
				t.Logf("Search position: Line=%d, Character=%d", tt.position.Line, tt.position.Character)
			}
		})
	}
}

func TestFindReferences(t *testing.T) {
	t.Skip("TODO: Update for language service architecture")
	tests := []struct {
		name               string
		content            string
		position           Position
		includeDeclaration bool
		expected           int // number of references expected
	}{
		{
			name: "Variable references with declaration",
			content: `my $count = 0;
$count++;
print $count;`,
			position:           Position{Line: 0, Character: 4}, // Position on $count declaration
			includeDeclaration: true,
			expected:           3,
		},
		{
			name: "Variable references without declaration",
			content: `my $count = 0;
$count++;
print $count;`,
			position:           Position{Line: 0, Character: 4}, // Position on $count declaration
			includeDeclaration: false,
			expected:           2,
		},
		{
			name: "Subroutine references",
			content: `sub greet {
    print "Hello\n";
}
greet();
greet();`,
			position:           Position{Line: 0, Character: 4}, // Position on greet in sub
			includeDeclaration: true,
			expected:           3,
		},
	}

	server := &Server{
		logger: testLogger(t),
		// documents: make(map[string]*Document), // TODO: Update for language service
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &Document{
				URI:  "file:///test.pl",
				Text: tt.content,
			}

			locations := server.findReferences(doc, tt.position, tt.includeDeclaration)
			if len(locations) != tt.expected {
				t.Errorf("Expected %d references, got %d", tt.expected, len(locations))
			}
		})
	}
}

func TestFormatDocument(t *testing.T) {
	t.Skip("TODO: Update for language service architecture")
	tests := []struct {
		name     string
		content  string
		options  FormattingOptions
		expected int // number of edits expected
	}{
		{
			name:     "Trim trailing whitespace",
			content:  "print 'hello';   \nprint 'world';  ",
			options:  FormattingOptions{TabSize: 4, InsertSpaces: true},
			expected: 2,
		},
		{
			name:     "Convert tabs to spaces",
			content:  "\tprint 'hello';\n\t\tprint 'world';",
			options:  FormattingOptions{TabSize: 4, InsertSpaces: true},
			expected: 2,
		},
		{
			name:     "No changes needed",
			content:  "print 'hello';\nprint 'world';",
			options:  FormattingOptions{TabSize: 4, InsertSpaces: false},
			expected: 0,
		},
	}

	server := &Server{
		logger: testLogger(t),
		// documents: make(map[string]*Document), // TODO: Update for language service
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &Document{
				URI:  "file:///test.pl",
				Text: tt.content,
			}

			edits := server.formatDocument(doc, tt.options)
			if len(edits) != tt.expected {
				t.Errorf("Expected %d edits, got %d", tt.expected, len(edits))
			}
		})
	}
}

func TestGenerateCodeActions(t *testing.T) {
	t.Skip("TODO: Update for language service architecture")
	tests := []struct {
		name        string
		content     string
		range_      Range
		diagnostics []Diagnostic
		expected    int // number of actions expected
	}{
		{
			name:    "Undefined variable fix",
			content: "print $foo;",
			range_: Range{
				Start: Position{Line: 0, Character: 6},
				End:   Position{Line: 0, Character: 10},
			},
			diagnostics: []Diagnostic{
				{
					Range: Range{
						Start: Position{Line: 0, Character: 6},
						End:   Position{Line: 0, Character: 10},
					},
					Severity: &[]DiagnosticSeverity{DiagnosticSeverityError}[0],
					Message:  "Variable $foo is undefined",
				},
			},
			expected: 1,
		},
		{
			name:    "Extract variable refactoring",
			content: "my $result = 2 + 3 + 4;",
			range_: Range{
				Start: Position{Line: 0, Character: 13},
				End:   Position{Line: 0, Character: 22},
			},
			diagnostics: []Diagnostic{},
			expected:    1, // Extract variable action
		},
	}

	server := &Server{
		logger: testLogger(t),
		// documents: make(map[string]*Document), // TODO: Update for language service
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &Document{
				URI:  "file:///test.pl",
				Text: tt.content,
			}

			context := CodeActionContext{
				Diagnostics: tt.diagnostics,
			}

			actions := server.generateCodeActions(doc, tt.range_, context)
			if len(actions) != tt.expected {
				t.Errorf("Expected %d actions, got %d", tt.expected, len(actions))
			}
		})
	}
}

func TestExtractSymbolAtPosition(t *testing.T) {
	t.Skip("TODO: Update for language service architecture")
	tests := []struct {
		name     string
		content  string
		position Position
		expected string
	}{
		{
			name:     "Variable with sigil",
			content:  "my $count = 42;",
			position: Position{Line: 0, Character: 5}, // on 'c' in count
			expected: "$count",
		},
		{
			name:     "Array variable",
			content:  "my @items = (1, 2, 3);",
			position: Position{Line: 0, Character: 4}, // on 'i' in items
			expected: "@items",
		},
		{
			name:     "Subroutine name",
			content:  "sub hello { }",
			position: Position{Line: 0, Character: 5}, // on 'h' in hello
			expected: "hello",
		},
		{
			name:     "No symbol",
			content:  "    ",
			position: Position{Line: 0, Character: 2},
			expected: "",
		},
	}

	server := &Server{
		logger: testLogger(t),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.extractSymbolAtPosition(tt.content, tt.position)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
