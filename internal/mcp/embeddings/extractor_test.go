// ABOUTME: Tests for code block extraction functionality
// ABOUTME: Verifies extraction of functions, methods, and classes from Perl code

package embeddings

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractor_ExtractFromFile(t *testing.T) {
	// Create test files
	testFiles := map[string]string{
		"simple_sub.pl": `#!/usr/bin/perl
use strict;
use warnings;

sub hello {
    my $name = shift;
    print "Hello, $name!\n";
}

sub goodbye {
    print "Goodbye!\n";
}
`,
		"typed_sub.pl": `#!/usr/bin/perl
use strict;
use warnings;

sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}

sub concat(Str $a, Str $b) -> Str {
    return $a . $b;
}
`,
		"class_example.pl": `#!/usr/bin/perl
use v5.40;
use experimental 'class';

class Point {
    field Num $x;
    field Num $y;

    method distance(Point $other) -> Num {
        my $dx = $self->x - $other->x;
        my $dy = $self->y - $other->y;
        return sqrt($dx * $dx + $dy * $dy);
    }
}
`,
		"package_example.pl": `package MyModule;
use strict;
use warnings;
use List::Util qw(sum);
use Data::Dumper;

sub process_data {
    my @data = @_;
    return sum(@data);
}

1;
`,
	}

	for filename, content := range testFiles {
		t.Run(filename, func(t *testing.T) {
			// Write test file
			tmpFile := t.TempDir() + "/" + filename
			err := writeTestFile(tmpFile, content)
			require.NoError(t, err)

			// Create extractor
			extractor, err := NewExtractor()
			require.NoError(t, err)

			// Extract blocks
			blocks, err := extractor.ExtractFromFile("test-project", tmpFile)
			require.NoError(t, err)

			// Verify we got blocks
			assert.NotEmpty(t, blocks)

			// Debug output for failing tests
			if len(blocks) == 0 {
				t.Logf("DEBUG: No blocks extracted for %s", filename)
				t.Logf("DEBUG: File content:\n%s", content)
			} else {
				t.Logf("DEBUG: Found %d blocks for %s", len(blocks), filename)
				for i, block := range blocks {
					t.Logf("DEBUG: Block %d: Type=%s, Name=%s, Content=%q", i, block.Type, block.Name, truncateString(block.Content, 50))
				}
			}

			// Check specific expectations based on file
			switch filename {
			case "simple_sub.pl":
				// Should find 2 subroutines
				funcBlocks := filterBlocksByType(blocks, "function")
				assert.Len(t, funcBlocks, 2)

				// Check names
				names := extractNames(funcBlocks)
				assert.Contains(t, names, "hello")
				assert.Contains(t, names, "goodbye")

			case "typed_sub.pl":
				// Should find 2 typed subroutines
				funcBlocks := filterBlocksByType(blocks, "function")
				if len(funcBlocks) == 0 {
					t.Logf("DEBUG: No function blocks found in typed_sub.pl")
					t.Logf("DEBUG: All blocks: %+v", blocks)
				}
				assert.Len(t, funcBlocks, 2)

				// Check type information
				for _, block := range funcBlocks {
					if block.Name == "add" {
						assert.Contains(t, block.TypeInfo, "return")
						assert.Equal(t, "Int", block.TypeInfo["return"])
					}
				}

			case "class_example.pl":
				// Should find class and method
				classBlocks := filterBlocksByType(blocks, "class")
				methodBlocks := filterBlocksByType(blocks, "method")

				assert.Len(t, classBlocks, 1)
				assert.Len(t, methodBlocks, 1)

				if len(classBlocks) > 0 {
					assert.Equal(t, "Point", classBlocks[0].Name)
				}
				if len(methodBlocks) > 0 {
					assert.Equal(t, "distance", methodBlocks[0].Name)
				}

			case "package_example.pl":
				// Should extract package context and imports
				allBlocks := blocks
				assert.NotEmpty(t, allBlocks)

				// Check context
				for _, block := range allBlocks {
					assert.Equal(t, "MyModule", block.Context)
					assert.Contains(t, block.Imports, "List::Util")
					assert.Contains(t, block.Imports, "Data::Dumper")
				}
			}
		})
	}
}

func TestExtractor_ConvertToDocuments(t *testing.T) {
	blocks := []*CodeBlock{
		{
			ID:        "project/file.pl/sub/test",
			Content:   "sub test { return 42; }",
			Type:      "function",
			Name:      "test",
			File:      "/path/to/file.pl",
			StartLine: 10,
			EndLine:   12,
			TypeInfo: map[string]string{
				"return": "Int",
			},
			Imports: []string{"strict", "warnings"},
			Context: "MyPackage",
		},
		{
			ID:        "project/file.pl/method/process",
			Content:   "method process { }",
			Type:      "method",
			Name:      "process",
			File:      "/path/to/file.pl",
			StartLine: 20,
			EndLine:   25,
			Context:   "MyClass",
		},
	}

	docs := ConvertToDocuments(blocks)
	require.Len(t, docs, 2)

	// Check first document
	doc1 := docs[0]
	assert.Equal(t, "project/file.pl/sub/test", doc1.ID)
	assert.Equal(t, "sub test { return 42; }", doc1.Content)
	assert.Equal(t, "function", doc1.Metadata["type"])
	assert.Equal(t, "test", doc1.Metadata["name"])
	assert.Equal(t, "/path/to/file.pl", doc1.Metadata["file"])
	assert.Equal(t, "10", doc1.Metadata["start_line"])
	assert.Equal(t, "12", doc1.Metadata["end_line"])
	assert.Equal(t, "MyPackage", doc1.Metadata["context"])
	assert.Equal(t, "strict,warnings", doc1.Metadata["imports"])
	assert.Equal(t, "Int", doc1.Metadata["type_return"])

	// Check second document
	doc2 := docs[1]
	assert.Equal(t, "project/file.pl/method/process", doc2.ID)
	assert.Equal(t, "method", doc2.Metadata["type"])
	assert.Equal(t, "process", doc2.Metadata["name"])
	assert.Equal(t, "MyClass", doc2.Metadata["context"])
}

func TestBatchExtractAndConvert(t *testing.T) {
	// Create test files
	testDir := t.TempDir()
	files := []string{
		testDir + "/file1.pl",
		testDir + "/file2.pl",
	}

	content1 := `sub func1 { return 1; }`
	content2 := `sub func2 { return 2; }`

	require.NoError(t, writeTestFile(files[0], content1))
	require.NoError(t, writeTestFile(files[1], content2))

	// Create extractor
	extractor, err := NewExtractor()
	require.NoError(t, err)

	// Batch extract
	docs, err := BatchExtractAndConvert(extractor, "test-project", files)
	require.NoError(t, err)

	// Should have at least 2 documents (one per file minimum)
	assert.GreaterOrEqual(t, len(docs), 2)

	// Check that we have documents from both files
	var file1Found, file2Found bool
	for _, doc := range docs {
		if strings.Contains(doc.Metadata["file"], "file1.pl") {
			file1Found = true
		}
		if strings.Contains(doc.Metadata["file"], "file2.pl") {
			file2Found = true
		}
	}
	assert.True(t, file1Found, "Should have extracted from file1.pl")
	assert.True(t, file2Found, "Should have extracted from file2.pl")
}

func TestExtractor_EdgeCases(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		expect  func(t *testing.T, blocks []*CodeBlock)
	}{
		{
			name:    "empty_file",
			content: "",
			expect: func(t *testing.T, blocks []*CodeBlock) {
				// Should create a file-level block
				assert.Len(t, blocks, 1)
				assert.Equal(t, "file", blocks[0].Type)
			},
		},
		{
			name: "comments_only",
			content: `# This is a comment
# Another comment
`,
			expect: func(t *testing.T, blocks []*CodeBlock) {
				// Should create a file-level block
				assert.Len(t, blocks, 1)
				assert.Equal(t, "file", blocks[0].Type)
			},
		},
		{
			name: "nested_subs",
			content: `sub outer {
    my $x = 10;
    my $inner = sub {
        return $x * 2;
    };
    return $inner->();
}`,
			expect: func(t *testing.T, blocks []*CodeBlock) {
				// Should find at least the outer sub
				funcBlocks := filterBlocksByType(blocks, "function")
				assert.GreaterOrEqual(t, len(funcBlocks), 1)

				hasOuter := false
				for _, block := range funcBlocks {
					if block.Name == "outer" {
						hasOuter = true
					}
				}
				assert.True(t, hasOuter)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Write test file
			tmpFile := t.TempDir() + "/test.pl"
			err := writeTestFile(tmpFile, tc.content)
			require.NoError(t, err)

			// Create extractor
			extractor, err := NewExtractor()
			require.NoError(t, err)

			// Extract blocks
			blocks, err := extractor.ExtractFromFile("test-project", tmpFile)
			require.NoError(t, err)

			// Run expectations
			tc.expect(t, blocks)
		})
	}
}

// Helper functions

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func filterBlocksByType(blocks []*CodeBlock, blockType string) []*CodeBlock {
	var filtered []*CodeBlock
	for _, block := range blocks {
		if block.Type == blockType {
			filtered = append(filtered, block)
		}
	}
	return filtered
}

func extractNames(blocks []*CodeBlock) []string {
	var names []string
	for _, block := range blocks {
		names = append(names, block.Name)
	}
	return names
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
