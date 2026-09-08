// ABOUTME: Tests that malformed Perl the grammar accepts without an error node is still detectable.
// ABOUTME: Covers the dropped-right-hand-side family: my $x = ; and its relatives.

package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/parser"
)

// TestDroppedRHSFamilyHasNoErrorNode records the bug exactly as filed: perl
// rejects every one of these, and the grammar reports HasError()==false. This
// test asserts the STATUS QUO so that a future grammar fix which starts
// producing error nodes shows up as a failure here rather than silently
// changing what IsDegenerate is compensating for.
func TestDroppedRHSFamilyHasNoErrorNode(t *testing.T) {
	// Every source here is rejected by `perl -c`. Verified against perl
	// 5.42.0: each reports `syntax error ... near "= ;"` or `near "+;"`.
	sources := []string{
		"my $x = ;",
		"my @a = ;",
		"my %h = ;",
		"$x = ;",
		"1 +;",
		"my $y = 1 +;",
		"my ($a) = ;",
	}

	p := parser.New()
	for _, src := range sources {
		t.Run(src, func(t *testing.T) {
			tree, err := p.Parse([]byte(src))
			require.NoError(t, err)
			root := tree.RootNode()
			require.NotNil(t, root)
			assert.False(t, root.HasError(),
				"documented grammar gap: no error node for source perl rejects")
		})
	}
}

// TestIsDegenerateFlagsDroppedRHS is the actual fix. The grammar will not tell
// us these are broken, so the wrapper must: each of these parses into a bare
// hidden supertype node (_term), which is an internal grammar rule name that
// only reaches the tree when error recovery has dropped real source.
func TestIsDegenerateFlagsDroppedRHS(t *testing.T) {
	sources := []string{
		"my $x = ;",
		"my @a = ;",
		"my %h = ;",
		"$x = ;",
		"1 +;",
		"my $y = 1 +;",
		"my ($a) = ;",
		"my $x = ; 1;",
	}

	p := parser.New()
	for _, src := range sources {
		t.Run(src, func(t *testing.T) {
			tree, err := p.Parse([]byte(src))
			require.NoError(t, err)
			assert.True(t, tree.IsDegenerate(),
				"source perl rejects must be detectable as degenerate")
		})
	}
}

// TestIsDegenerateAcceptsValidPerl is the other half, and the half that makes
// the signal worth anything: a check that fires on everything is not a check.
// `return ;` is in here deliberately -- it looks like the family but perl
// compiles it, so flagging it would be a false positive.
func TestIsDegenerateAcceptsValidPerl(t *testing.T) {
	sources := []string{
		"my $x = 1;",
		"my $x = 42;\n",
		"my @a = (1, 2, 3);",
		"my %h = (a => 1);",
		"$x = 1;",
		"1 + 2;",
		"return ;",
		"sub f { return ; }",
		"sub f { my $x = shift; return $x + 1; }",
		"my $s = \"hello\";",
		"for my $i (1 .. 10) { print $i; }",
		"package Foo; use strict; use warnings; 1;",
	}

	p := parser.New()
	for _, src := range sources {
		t.Run(src, func(t *testing.T) {
			tree, err := p.Parse([]byte(src))
			require.NoError(t, err)
			assert.False(t, tree.IsDegenerate(),
				"valid perl must not be flagged degenerate")
		})
	}
}

// TestDegenerateKindsNamesTheLeak reports WHICH hidden rule leaked, so a
// report can say what it saw rather than only that something was wrong.
func TestDegenerateKindsNamesTheLeak(t *testing.T) {
	p := parser.New()

	tree, err := p.Parse([]byte("my $x = ;"))
	require.NoError(t, err)
	kinds := tree.DegenerateKinds()
	require.NotEmpty(t, kinds, "a degenerate tree must name what leaked")
	assert.Contains(t, kinds, "_term")

	clean, err := p.Parse([]byte("my $x = 1;"))
	require.NoError(t, err)
	assert.Empty(t, clean.DegenerateKinds(), "a clean parse names nothing")
}
