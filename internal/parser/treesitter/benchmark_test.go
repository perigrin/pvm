// ABOUTME: Benchmarks for tree-sitter parser
// ABOUTME: Evaluates the performance of the tree-sitter parser

package treesitter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/astnav"
)

// BenchmarkParseSmallFile benchmarks parsing a small Perl file (< 100 lines)
func BenchmarkParseSmallFile(b *testing.B) {
	// Create a temporary directory for benchmark files
	tempDir, err := os.MkdirTemp("", "benchmark-parser")
	require.NoError(b, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a sample small Perl file
	smallFile := filepath.Join(tempDir, "small.pl")
	smallContent := `use v5.36;

# Small Perl file with type annotations
my Int $count = 0;
my Str $name = "Example";
my ArrayRef[Int] $numbers = [1, 2, 3];

sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}

method greet(Str $name) returns Str {
    return "Hello, $name!";
}

field Bool $flag;

type ID = Int;
type Names = ArrayRef[Str];
`

	err = os.WriteFile(smallFile, []byte(smallContent), 0644)
	require.NoError(b, err)

	// Skip the benchmark if tree-sitter-perl is not available
	b.Skip("Skipping benchmark until tree-sitter-perl is fully implemented")

	// Create a parser
	parser, err := NewParser(false)
	require.NoError(b, err)
	defer parser.Close()

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseFile(smallFile)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseMediumFile benchmarks parsing a medium-sized Perl file (100-500 lines)
func BenchmarkParseMediumFile(b *testing.B) {
	// Create a temporary directory for benchmark files
	tempDir, err := os.MkdirTemp("", "benchmark-parser")
	require.NoError(b, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a sample medium-sized Perl file
	mediumFile := filepath.Join(tempDir, "medium.pl")

	var mediumContent strings.Builder
	mediumContent.WriteString("use v5.36;\n\n")
	mediumContent.WriteString("# Medium-sized Perl file with type annotations\n\n")

	// Generate a medium-sized file with repeated patterns
	for i := 0; i < 50; i++ {
		mediumContent.WriteString(fmt.Sprintf("my Int $count%d = %d;\n", i, i))
		mediumContent.WriteString(fmt.Sprintf("my Str $name%d = \"Example%d\";\n", i, i))
		mediumContent.WriteString(fmt.Sprintf("my ArrayRef[Int] $numbers%d = [%d, %d, %d];\n\n", i, i, i+1, i+2))

		mediumContent.WriteString(fmt.Sprintf("sub add%d(Int $a, Int $b) -> Int {\n", i))
		mediumContent.WriteString("    return $a + $b;\n")
		mediumContent.WriteString("}\n\n")
	}

	err = os.WriteFile(mediumFile, []byte(mediumContent.String()), 0644)
	require.NoError(b, err)

	// Skip the benchmark if tree-sitter-perl is not available
	b.Skip("Skipping benchmark until tree-sitter-perl is fully implemented")

	// Create a parser
	parser, err := NewParser(false)
	require.NoError(b, err)
	defer parser.Close()

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseFile(mediumFile)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseLargeFile benchmarks parsing a large Perl file (> 500 lines)
func BenchmarkParseLargeFile(b *testing.B) {
	// Create a temporary directory for benchmark files
	tempDir, err := os.MkdirTemp("", "benchmark-parser")
	require.NoError(b, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a sample large Perl file
	largeFile := filepath.Join(tempDir, "large.pl")

	var largeContent strings.Builder
	largeContent.WriteString("use v5.36;\n\n")
	largeContent.WriteString("# Large Perl file with type annotations\n\n")

	// Generate a large file with repeated patterns
	for i := 0; i < 300; i++ {
		largeContent.WriteString(fmt.Sprintf("my Int $count%d = %d;\n", i, i))
		largeContent.WriteString(fmt.Sprintf("my Str $name%d = \"Example%d\";\n", i, i))
		largeContent.WriteString(fmt.Sprintf("my ArrayRef[Int] $numbers%d = [%d, %d, %d];\n\n", i, i, i+1, i+2))

		largeContent.WriteString(fmt.Sprintf("sub add%d(Int $a, Int $b) -> Int {\n", i))
		largeContent.WriteString("    return $a + $b;\n")
		largeContent.WriteString("}\n\n")
	}

	err = os.WriteFile(largeFile, []byte(largeContent.String()), 0644)
	require.NoError(b, err)

	// Skip the benchmark if tree-sitter-perl is not available
	b.Skip("Skipping benchmark until tree-sitter-perl is fully implemented")

	// Create a parser
	parser, err := NewParser(false)
	require.NoError(b, err)
	defer parser.Close()

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseFile(largeFile)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExtractTypeAnnotations benchmarks extracting type annotations from a parsed file
func BenchmarkExtractTypeAnnotations(b *testing.B) {
	// Create a temporary directory for benchmark files
	tempDir, err := os.MkdirTemp("", "benchmark-parser")
	require.NoError(b, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a sample Perl file with many type annotations
	annotatedFile := filepath.Join(tempDir, "annotated.pl")

	var annotatedContent strings.Builder
	annotatedContent.WriteString("use v5.36;\n\n")
	annotatedContent.WriteString("# Perl file with many type annotations\n\n")

	// Generate a file with many type annotations
	for i := 0; i < 100; i++ {
		annotatedContent.WriteString(fmt.Sprintf("my Int $count%d = %d;\n", i, i))
		annotatedContent.WriteString(fmt.Sprintf("my Str $name%d = \"Example%d\";\n", i, i))
		annotatedContent.WriteString(fmt.Sprintf("my ArrayRef[Int] $numbers%d = [%d, %d, %d];\n", i, i, i+1, i+2))
		annotatedContent.WriteString(fmt.Sprintf("my HashRef[Str, Int] $map%d = { a => 1, b => 2 };\n", i))
		annotatedContent.WriteString(fmt.Sprintf("my Maybe[Int] $optional%d = undef;\n", i))
		annotatedContent.WriteString(fmt.Sprintf("my Int|Str $union%d = %d;\n\n", i, i))
	}

	err = os.WriteFile(annotatedFile, []byte(annotatedContent.String()), 0644)
	require.NoError(b, err)

	// Skip the benchmark if tree-sitter-perl is not available
	b.Skip("Skipping benchmark until tree-sitter-perl is fully implemented")

	// Create a parser
	parser, err := NewParser(false)
	require.NoError(b, err)
	defer parser.Close()

	// Parse the file once
	tree, err := parser.ParseFile(annotatedFile)
	require.NoError(b, err)

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nav := astnav.NewNavigator(tree)
		_ = nav.FindTypeAnnotations()
	}
}

// BenchmarkParseTypeExpression benchmarks parsing type expressions
func BenchmarkParseTypeExpression(b *testing.B) {
	tests := []struct {
		name string
		expr string
	}{
		{"SimpleType", "Int"},
		{"ParameterizedType", "ArrayRef[Int]"},
		{"ComplexParameterizedType", "HashRef[Str, ArrayRef[Int]]"},
		{"UnionType", "Int|Str|Bool"},
		{"IntersectionType", "Serializable&Printable&Comparable"},
		{"NegationType", "!Int"},
		{"ComplexType", "Maybe[HashRef[Str, ArrayRef[Int|Float]]]"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pos := Position{Line: 1, Column: 1}
				_, err := ParseTypeExpression(tt.expr, pos)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
