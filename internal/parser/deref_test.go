// ABOUTME: Tests the third silent-loss family: a deref block holding one explicit reference.
// ABOUTME: `@{ \@a }` compiles, emits no error node, and collapses into a varname that is not a name.

package parser_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/parser"
)

// TestDerefOfExplicitRefCollapses records the bug exactly as measured, and is
// the reason a third detector is needed at all.
//
// `my @b = @{ \@a };` compiles under perl and its optree holds one srefgen. The
// grammar accepts it, reports no error node, and leaks no hidden rule -- so
// neither HasError nor the _term detector fires. What it produces instead is an
// `array` node whose `varname` child spans the literal text `\@a`: the braces
// became anonymous tokens of the array and the block was never built.
//
// That is not a dropped token, it is a node contradicting its own rule. A
// consumer reading `varname` expects an identifier and gets a backslash and a
// sigil. The reference perl took is nowhere in the tree as a reference, which
// is what made the fidelity harness score the file WRONG.
//
// This test asserts the STATUS QUO of the two existing signals, so a grammar
// fix that starts flagging these shows up here rather than silently changing
// what the new detector compensates for.
func TestDerefOfExplicitRefCollapses(t *testing.T) {
	// Every source here is accepted by `perl -c` and its optree contains
	// exactly one srefgen. Measured against perl 5.42.0.
	sources := []string{
		"my @a = (1,2); my @b = @{ \\@a };",
		"my %h = (a=>1); my %g = %{ \\%h };",
		"my $x = 1; my $y = ${ \\$x };",
		"sub f {} my $z = &{ \\&f };",
		"my @b = @{\\@a};",
	}

	p := parser.New()
	for _, src := range sources {
		t.Run(src, func(t *testing.T) {
			tree, err := p.Parse([]byte(src))
			require.NoError(t, err)
			root := tree.RootNode()
			require.NotNil(t, root)

			assert.False(t, root.HasError(),
				"documented grammar gap: no error node for a collapsed deref block")

			// And no hidden rule leaks, so the _term detector could never
			// have seen this. A `_`-prefixed kind is what that signal
			// reports; its absence is why a second one had to exist.
			for _, k := range tree.DegenerateKinds() {
				assert.False(t, strings.HasPrefix(k, "_"),
					"no hidden rule leaks here, so the _term detector cannot see it (%s)", k)
			}
		})
	}
}

// TestIsDegenerateFlagsCollapsedDeref is the fix. The grammar will not tell us
// the tree is wrong, so the wrapper must: a `varname` node whose text is not a
// name is a node contradicting its own rule, exactly as a leaked `_term` is.
func TestIsDegenerateFlagsCollapsedDeref(t *testing.T) {
	sources := []string{
		"my @a = (1,2); my @b = @{ \\@a };",
		"my %h = (a=>1); my %g = %{ \\%h };",
		"my $x = 1; my $y = ${ \\$x };",
		"sub f {} my $z = &{ \\&f };",
		"my @b = @{\\@a};",
		"my @b = @{ \\$r };",
		"print ${ \\$x };",
		"foo(@{ \\@a });",
	}

	p := parser.New()
	for _, src := range sources {
		t.Run(src, func(t *testing.T) {
			tree, err := p.Parse([]byte(src))
			require.NoError(t, err)
			assert.True(t, tree.IsDegenerate(),
				"a collapsed deref block must be detectable as degenerate")
		})
	}
}

// TestCollapsedDerefDetectorIsNarrow is the half that makes the signal worth
// having. A detector that fires on every deref would cost the sweep every
// `@$r` and `@{ $r }` in the corpus, which is most of them.
//
// The rows here are the near misses, and they are near misses by one character:
// the grammar builds a correct block for `@{ \@a, }` and `@{ +\@a }` and even
// `@{ \ @a }`. Only a lone backslash tight against a sigilled variable
// collapses, so only that must be flagged.
func TestCollapsedDerefDetectorIsNarrow(t *testing.T) {
	sources := []string{
		"my @b = @{ $r };",
		"my @b = @$r;",
		"my @b = $r->@*;",
		"my $v = $$r[0];",
		"my $r = \\@a;",
		"my @b = @a;",
		"my $x = $Foo::Bar::baz;",
		"my $x = ${^GLOBAL_PHASE};",
		"my @b = @{$r->{k}};",
		"my @b = @{ [1,2] };",
		"my @b = @{ f() };",
		"my @b = @{ \\@a, };",
		"my @b = @{ +\\@a };",
		"my @b = @{ \\ @a };",
		"my $s = \"@{[ \\@a ]}\";",

		// `$\` is the output record separator, and its varname really is a
		// lone backslash -- a correct parse of valid Perl that a naive
		// "varname starts with a backslash" test flags. Found by the corpus
		// sweep, in t/op/tiehandle.t and t/uni/lex_utf8.t, which is the only
		// reason this row exists rather than the collapse family's shape.
		"my $ors = $\\;",
		"local $\\ = 'x';",
		"local $\\;",
		"print $\\;",
	}

	p := parser.New()
	for _, src := range sources {
		t.Run(src, func(t *testing.T) {
			tree, err := p.Parse([]byte(src))
			require.NoError(t, err)
			assert.False(t, tree.IsDegenerate(),
				"a deref the grammar parses correctly must not be flagged")
		})
	}
}

// TestDegenerateKindsNamesTheCollapsedDeref keeps the report triageable. A
// no-answer that says only "declined" cannot be acted on, so the detector must
// name what it saw the same way a leaked hidden rule is named.
func TestDegenerateKindsNamesTheCollapsedDeref(t *testing.T) {
	p := parser.New()

	tree, err := p.Parse([]byte("my @b = @{ \\@a };"))
	require.NoError(t, err)
	assert.Contains(t, tree.DegenerateKinds(), "varname",
		"the detector must name the node kind that contradicted itself")

	clean, err := p.Parse([]byte("my @b = @{ $r };"))
	require.NoError(t, err)
	assert.Empty(t, clean.DegenerateKinds(), "a clean parse names nothing")
}
