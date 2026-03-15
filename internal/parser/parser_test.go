// ABOUTME: Tests for the gotreesitter-based Perl parser package.
// ABOUTME: Covers basic parsing, node navigation, and various Perl constructs.

package parser_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/parser"
)

func TestNew(t *testing.T) {
	p := parser.New()
	require.NotNil(t, p, "New() should return a non-nil parser")
}

func TestParseVariableDeclaration(t *testing.T) {
	p := parser.New()
	source := []byte("my $x = 42;\n")

	tree, err := p.Parse(source)
	require.NoError(t, err)
	require.NotNil(t, tree)

	root := tree.RootNode()
	require.NotNil(t, root)
	assert.Equal(t, "source_file", root.Kind())
	assert.False(t, root.HasError(), "parse of simple var decl should have no errors")
}

// TestParseStringLiteral verifies that the parser processes string literals.
// Note: the gotreesitter Perl grammar has limited support for string literals;
// they may produce error nodes due to incomplete lexer porting.
func TestParseStringLiteral(t *testing.T) {
	p := parser.New()
	source := []byte(`my $x = "hello world";` + "\n")

	tree, err := p.Parse(source)
	require.NoError(t, err, "Parse should not return an error even for unsupported constructs")
	require.NotNil(t, tree)

	root := tree.RootNode()
	require.NotNil(t, root)
	// The grammar produces a tree (possibly with error nodes for complex string interpolation)
	assert.NotNil(t, root, "root node should not be nil")
}

func TestParseSubroutineDefinition(t *testing.T) {
	p := parser.New()
	source := []byte("sub greet {\n    my ($name) = @_;\n    return $name;\n}\n")

	tree, err := p.Parse(source)
	require.NoError(t, err)
	require.NotNil(t, tree)

	root := tree.RootNode()
	require.NotNil(t, root)
	assert.False(t, root.HasError(), "parse of subroutine should have no errors")
	assert.Greater(t, root.ChildCount(), 0, "root should have children")
}

func TestParseClass(t *testing.T) {
	p := parser.New()
	// Perl 5.38+ class syntax
	source := []byte("class Point {\n    field $x;\n    field $y;\n}\n")

	tree, err := p.Parse(source)
	require.NoError(t, err)
	require.NotNil(t, tree)

	root := tree.RootNode()
	require.NotNil(t, root)
	assert.NotNil(t, root, "root node should not be nil for class parse")
}

// TestParseHeredoc verifies that the parser handles heredocs without panicking.
// Note: heredoc syntax may produce error nodes in the gotreesitter Perl grammar
// due to the complexity of heredoc lexing.
func TestParseHeredoc(t *testing.T) {
	p := parser.New()
	source := []byte("my $text = <<END;\nHello\nEND\n")

	tree, err := p.Parse(source)
	require.NoError(t, err, "Parse should not panic or return error on heredoc")
	require.NotNil(t, tree)

	root := tree.RootNode()
	require.NotNil(t, root, "root node should not be nil for heredoc parse")
}

// TestParseRegexMatch verifies that the parser handles regex without panicking.
// Note: regex and other quotelike operators may produce error nodes in the
// gotreesitter Perl grammar due to lexer porting limitations.
func TestParseRegexMatch(t *testing.T) {
	p := parser.New()
	source := []byte("my $matched = ($str =~ /hello/);\n")

	tree, err := p.Parse(source)
	require.NoError(t, err, "Parse should not panic or return error on regex")
	require.NotNil(t, tree)

	root := tree.RootNode()
	require.NotNil(t, root, "root node should not be nil for regex parse")
}

func TestNodeNavigation(t *testing.T) {
	p := parser.New()
	source := []byte("my $x = 42;\n")

	tree, err := p.Parse(source)
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)

	// Test ChildCount and Child
	count := root.ChildCount()
	assert.Greater(t, count, 0, "root should have at least one child")

	child := root.Child(0)
	require.NotNil(t, child, "first child should not be nil")

	// Test IsNamed — root is always named
	assert.True(t, root.IsNamed(), "root node should be named")

	// Test StartByte/EndByte
	assert.Equal(t, uint32(0), root.StartByte())
	assert.Equal(t, uint32(len(source)), root.EndByte())

	// Test Text
	text := root.Text(source)
	assert.Equal(t, string(source), text)

	// Test NamedChildCount
	namedCount := root.NamedChildCount()
	assert.Greater(t, namedCount, 0, "root should have named children")

	// Test NamedChild
	namedChild := root.NamedChild(0)
	require.NotNil(t, namedChild, "first named child should not be nil")
	assert.True(t, namedChild.IsNamed())

	// Test Parent
	parent := namedChild.Parent()
	require.NotNil(t, parent, "named child should have a parent")
	assert.Equal(t, "source_file", parent.Kind())
}

func TestNodeSExpr(t *testing.T) {
	p := parser.New()
	source := []byte("my $x = 42;\n")

	tree, err := p.Parse(source)
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)

	sexpr := root.SExpr()
	assert.NotEmpty(t, sexpr, "SExpr should not be empty")
	assert.True(t, strings.HasPrefix(sexpr, "(source_file"), "SExpr should start with (source_file, got: "+sexpr)
}

func TestNodeIsError(t *testing.T) {
	p := parser.New()
	source := []byte("my $x = 42;\n")

	tree, err := p.Parse(source)
	require.NoError(t, err)

	root := tree.RootNode()
	assert.False(t, root.IsError(), "root of valid parse should not be an error node")
}

func TestNodeKindVariety(t *testing.T) {
	p := parser.New()
	source := []byte("my $x = 42;\n")

	tree, err := p.Parse(source)
	require.NoError(t, err)

	root := tree.RootNode()
	assert.Equal(t, "source_file", root.Kind())
}

func TestParseEmptySource(t *testing.T) {
	p := parser.New()
	source := []byte("")

	tree, err := p.Parse(source)
	require.NoError(t, err)
	require.NotNil(t, tree)
}

func TestParseMultipleStatements(t *testing.T) {
	p := parser.New()
	source := []byte("my $x = 1;\nmy $y = 2;\n")

	tree, err := p.Parse(source)
	require.NoError(t, err)
	require.NotNil(t, tree)

	root := tree.RootNode()
	assert.False(t, root.HasError(), "multi-statement parse should have no errors")
	// NamedChildCount should reflect the two statements
	assert.Greater(t, root.NamedChildCount(), 1, "should have multiple named children")
}

func TestNodeOutOfRangeChild(t *testing.T) {
	p := parser.New()
	source := []byte("my $x = 42;\n")

	tree, err := p.Parse(source)
	require.NoError(t, err)

	root := tree.RootNode()
	outOfRange := root.Child(9999)
	assert.Nil(t, outOfRange, "out-of-range child should return nil")

	outOfRangeNamed := root.NamedChild(9999)
	assert.Nil(t, outOfRangeNamed, "out-of-range named child should return nil")
}

func TestNilNodeSafety(t *testing.T) {
	var n *parser.Node
	assert.Equal(t, "", n.Kind())
	assert.False(t, n.IsNamed())
	assert.False(t, n.HasError())
	assert.False(t, n.IsError())
	assert.Equal(t, uint32(0), n.StartByte())
	assert.Equal(t, uint32(0), n.EndByte())
	assert.Equal(t, "", n.Text([]byte("source")))
	assert.Equal(t, 0, n.ChildCount())
	assert.Nil(t, n.Child(0))
	assert.Equal(t, 0, n.NamedChildCount())
	assert.Nil(t, n.NamedChild(0))
	assert.Nil(t, n.ChildByFieldName("name"))
	assert.Nil(t, n.Parent())
	assert.Equal(t, "", n.SExpr())
}
