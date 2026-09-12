// ABOUTME: Tests that a deref block swallowing an explicit reference scores no-answer, never WRONG.
// ABOUTME: WRONG is the only bucket that fails a build, so a false positive there makes the gate untrustworthy.

package parseoracle

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/parser"
)

// TestCompareDerefOfExplicitRefIsNotWrong is the defect the issue reports.
//
// `my @a = (1,2); my @b = @{ \@a };` compiles under perl with one srefgen in
// its optree. Our grammar collapses the deref block into a `varname` holding
// the text `\@a`, so nothing in the tree accounts for the reference and the
// comparison scored the file WRONG.
//
// WRONG is the only bucket that fails a build. A false positive there makes
// the gate untrustworthy, and an untrustworthy gate gets turned off -- which
// is why this is the critical half of the fix and no-answer is only the
// second-best outcome rather than a consolation.
func TestCompareDerefOfExplicitRefIsNotWrong(t *testing.T) {
	// One srefgen is what perl really reports for each of these. Measured
	// against perl 5.42.0 with -MO=Concise,-exec.
	sources := []string{
		"my @a = (1,2); my @b = @{ \\@a };\n",
		"my %h = (a=>1); my %g = %{ \\%h };\n",
		"my $x = 1; my $y = ${ \\$x };\n",
		"sub f {} my $z = &{ \\&f };\n",
	}

	for _, src := range sources {
		t.Run(src, func(t *testing.T) {
			tree, err := parser.New().Parse([]byte(src))
			require.NoError(t, err)

			v := Compare(Facts{OK: true, Srefgen: 1}, tree, []byte(src))
			assert.NotEqual(t, BucketWrong, v.Bucket,
				"a reference our tree threw away is not a parse we got wrong (%s)", v.Detail)
			assert.Equal(t, BucketNoAnswer, v.Bucket,
				"the honest verdict is a refusal: we cannot compare a tree that is not a parse (%s)",
				v.Detail)
		})
	}
}

// TestCompareDerefDetailNamesTheCollapse keeps the report triageable. A
// no-answer reading only "declined" cannot be acted on, so the detail has to
// name the node kind that contradicted itself -- the same contract the
// hidden-rule leak already meets.
func TestCompareDerefDetailNamesTheCollapse(t *testing.T) {
	src := "my @a = (1,2); my @b = @{ \\@a };\n"
	tree, err := parser.New().Parse([]byte(src))
	require.NoError(t, err)

	v := Compare(Facts{OK: true, Srefgen: 1}, tree, []byte(src))
	require.Equal(t, BucketNoAnswer, v.Bucket)
	assert.Contains(t, v.Detail, "varname",
		"detail must name what collapsed so the report can be triaged")

	// And it must not mis-describe what happened. Nothing was dropped here --
	// every byte of the source is still in the tree, under a node that denies
	// its own rule -- and `varname` is not a hidden rule. A detail carrying
	// the _term family's wording sends a reader looking for the wrong defect.
	assert.NotContains(t, v.Detail, "dropped source",
		"nothing was dropped: the source is all present, the structure is wrong")
	assert.NotContains(t, v.Detail, "hidden grammar rule",
		"varname is a real rule, not a hidden one")
}

// TestCompareDerefGateStaysNarrow is the guard against paying for this fix
// with the metric itself. Declining on every dereference would score 100%
// non-WRONG and measure nothing, which is the failure mode the Bucket doc
// comment warns about.
//
// Every row here is a dereference the grammar parses correctly, including the
// near misses that differ from the broken shape by one character. All must
// still reach a verdict.
func TestCompareDerefGateStaysNarrow(t *testing.T) {
	sources := []string{
		"my $r = [1]; my @b = @{ $r };\n",
		"my $r = [1]; my @b = @$r;\n",
		"my $r = [1]; my @b = $r->@*;\n",
		"my $r = [1]; my $v = $$r[0];\n",
		"my @a = (1,2); my @b = @{ \\@a, };\n",
		"my @a = (1,2); my @b = @{ +\\@a };\n",
		"my @a = (1,2); my @b = @{ \\ @a };\n",
	}

	for _, src := range sources {
		t.Run(src, func(t *testing.T) {
			tree, err := parser.New().Parse([]byte(src))
			require.NoError(t, err)

			v := Compare(Facts{OK: true}, tree, []byte(src))
			assert.NotEqual(t, BucketNoAnswer, v.Bucket,
				"a dereference we parse correctly must still be scored (%s)", v.Detail)
		})
	}
}
