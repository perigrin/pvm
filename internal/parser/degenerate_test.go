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

// TestDocumentedProbeTableIsAccurate pins the table in
// docs/specs/perl-parser/00-findings.md §0.4.1 to the parser's real behaviour,
// so the documented limitation cannot drift from what the code does. Every row
// asserts both halves the doc claims: whether the grammar flags an error, and
// whether the wrapper reports the tree as degenerate.
//
// The `perl -c` column is not asserted here -- shelling out to perl belongs to
// internal/parseoracle, and this package must not depend on an interpreter
// being installed. Those verdicts were measured on perl 5.42.0 and are recorded
// in the doc.
func TestDocumentedProbeTableIsAccurate(t *testing.T) {
	rows := []struct {
		src        string
		hasError   bool
		degenerate bool
		perlOK     bool // recorded from perl 5.42.0, documented not asserted
	}{
		{"my $x = ;", false, true, false},
		{"my @a = ;", false, true, false},
		{"my %h = ;", false, true, false},
		{"$x = ;", false, true, false},
		{"my ($a) = ;", false, true, false},
		{"1 +;", false, true, false},
		{"my $y = 1 +;", false, true, false},
		// The near miss: the grammar CAN flag this family, it just does not
		// for the assignment rows above.
		{"f( ;", true, false, false},
		// The trap: looks like the family, is legal Perl.
		{"return ;", false, false, true},
	}

	p := parser.New()
	for _, row := range rows {
		t.Run(row.src, func(t *testing.T) {
			tree, err := p.Parse([]byte(row.src))
			require.NoError(t, err)
			root := tree.RootNode()
			require.NotNil(t, root)

			assert.Equal(t, row.hasError, root.HasError(),
				"HasError() column of the documented table")
			assert.Equal(t, row.degenerate, tree.IsDegenerate(),
				"IsDegenerate() column of the documented table")

			// The two signals together must cover everything perl rejects.
			// If both are false for source perl rejects, the limitation is
			// undocumented again, which is the whole thing this closes.
			if !row.perlOK {
				assert.True(t, root.HasError() || tree.IsDegenerate(),
					"source perl rejects must be detectable by one signal or the other")
			}
		})
	}
}

// TestSplitStatementLossIsTheWorstCase records the most damaging row
// concretely: `my $y = 1 +;` does not merely drop a token, it turns one
// expression into two sibling statements with the operator gone. A consumer
// reading that tree sees two unrelated terms, which is why "no error node" is
// not good enough on its own.
func TestSplitStatementLossIsTheWorstCase(t *testing.T) {
	p := parser.New()
	tree, err := p.Parse([]byte("my $y = 1 +;"))
	require.NoError(t, err)
	root := tree.RootNode()
	require.NotNil(t, root)

	assert.Equal(t, 2, root.NamedChildCount(),
		"one expression became two sibling statements")
	assert.NotContains(t, root.SExpr(), "binary",
		"the + operator is absent from the tree entirely")
	assert.True(t, tree.IsDegenerate(), "and nothing but this flag marks the loss")
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
