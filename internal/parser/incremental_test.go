// ABOUTME: Tests for incremental parsing support in the Perl parser.
// ABOUTME: Verifies that edited trees can be re-parsed efficiently.

package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/parser"
)

func TestParseIncremental(t *testing.T) {
	p := parser.New()

	// Parse initial source
	original := []byte("my $x = 1;\n")
	tree, err := p.Parse(original)
	require.NoError(t, err)
	require.NotNil(t, tree)

	root := tree.RootNode()
	assert.False(t, root.HasError(), "original parse should have no errors")

	// Apply an edit: replace "1" with "42" (byte 8 to 9, new text is "42")
	// "my $x = 1;\n"
	//  01234567890
	//          ^-- byte 8 is "1"
	edit := parser.Edit{
		StartByte:  8,
		OldEndByte: 9,
		NewEndByte: 10,
	}

	// Edited source
	edited := []byte("my $x = 42;\n")

	newTree, err := p.ParseIncremental(edited, tree, edit)
	require.NoError(t, err)
	require.NotNil(t, newTree)

	newRoot := newTree.RootNode()
	require.NotNil(t, newRoot)
	assert.False(t, newRoot.HasError(), "incremental parse should have no errors")
	assert.Equal(t, "source_file", newRoot.Kind())
}

func TestParseIncrementalAddStatement(t *testing.T) {
	p := parser.New()

	// Parse initial source
	original := []byte("my $x = 1;\n")
	tree, err := p.Parse(original)
	require.NoError(t, err)

	// Add a second statement at the end
	// old: "my $x = 1;\n"  (12 bytes)
	// new: "my $x = 1;\nmy $y = 2;\n"
	edit := parser.Edit{
		StartByte:  uint32(len(original)),
		OldEndByte: uint32(len(original)),
		NewEndByte: uint32(len(original)) + 12,
	}

	edited := []byte("my $x = 1;\nmy $y = 2;\n")

	newTree, err := p.ParseIncremental(edited, tree, edit)
	require.NoError(t, err)
	require.NotNil(t, newTree)

	newRoot := newTree.RootNode()
	require.NotNil(t, newRoot)
	assert.False(t, newRoot.HasError(), "incremental parse with new statement should have no errors")
	assert.Greater(t, newRoot.NamedChildCount(), 1, "should have multiple statements after adding one")
}

func TestParseIncrementalWithNilOldTree(t *testing.T) {
	p := parser.New()

	edit := parser.Edit{
		StartByte:  0,
		OldEndByte: 0,
		NewEndByte: 11,
	}

	source := []byte("my $x = 1;\n")
	// Passing nil as the old tree should fall back to a full parse
	tree, err := p.ParseIncremental(source, nil, edit)
	require.NoError(t, err)
	require.NotNil(t, tree)

	root := tree.RootNode()
	require.NotNil(t, root)
	assert.Equal(t, "source_file", root.Kind())
}
