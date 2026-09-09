// ABOUTME: Tests that a tree the grammar built by dropping source cannot score as a clean parse.
// ABOUTME: Closes the hole where malformed Perl inside an otherwise-compiling file scored exact.

package parseoracle

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/parser"
)

// TestCompareDeclinesOnDegenerateTree is the fidelity hole from the issue.
//
// `my $x = ;` is rejected by perl and accepted by our grammar with no error
// node and the right-hand side dropped. Today the harness escapes scoring it
// only by accident: perl refuses the whole file, so Facts.OK is false and
// comparison never reaches a verdict. Hand Compare facts that DID compile --
// which is exactly what happens when this construct sits inside a file that
// otherwise compiles, or when the oracle's ground truth comes from a different
// region of the file -- and the degenerate tree scores exact.
//
// A parse built by throwing source away is not a parse we can compare, so the
// honest verdict is no-answer: we declined, which is what the bucket means.
func TestCompareDeclinesOnDegenerateTree(t *testing.T) {
	p := parser.New()

	sources := []string{
		"my $x = ;",
		"my @a = ;",
		"my %h = ;",
		"$x = ;",
		"1 +;",
		"my $y = 1 +;",
		"my ($a) = ;",
	}

	for _, src := range sources {
		t.Run(src, func(t *testing.T) {
			tree, err := p.Parse([]byte(src))
			require.NoError(t, err)

			// Sanity: the grammar really does not flag these, so the
			// existing HasError gate cannot be what saves us.
			require.False(t, tree.RootNode().HasError(),
				"precondition: grammar emits no error node here")

			// Ground truth that compiled. Compare must still decline,
			// because our tree is not a parse of this source.
			v := Compare(Facts{OK: true}, tree, []byte(src))
			assert.Equal(t, BucketNoAnswer, v.Bucket,
				"a tree built by dropping source must not score a verdict (%s)", v.Detail)
			assert.NotEqual(t, BucketExact, v.Bucket,
				"scoring this exact is the fidelity hole this test exists to close")
		})
	}
}

// TestCompareStillScoresValidPerl is the guard that keeps the new gate from
// eating the metric. A gate that declines on everything scores 100% non-WRONG
// and measures nothing, which is the failure mode the Bucket doc comment warns
// about. `return ;` is here because it looks like the family above and compiles.
func TestCompareStillScoresValidPerl(t *testing.T) {
	p := parser.New()

	sources := []string{
		"sub f{} my @a; f(@a);\n",
		"my $x = 1;",
		"return ;",
		"my @a = (1, 2, 3);",
	}

	for _, src := range sources {
		t.Run(src, func(t *testing.T) {
			tree, err := p.Parse([]byte(src))
			require.NoError(t, err)
			v := Compare(Facts{OK: true}, tree, []byte(src))
			assert.NotEqual(t, BucketNoAnswer, v.Bucket,
				"valid perl must still reach a verdict (%s)", v.Detail)
		})
	}
}

// TestCompareDegenerateDetailNamesTheLeak keeps the report actionable: a
// no-answer that says only "declined" cannot be triaged, so the detail must
// name the hidden rule that surfaced.
func TestCompareDegenerateDetailNamesTheLeak(t *testing.T) {
	p := parser.New()
	tree, err := p.Parse([]byte("my $x = ;"))
	require.NoError(t, err)

	v := Compare(Facts{OK: true}, tree, []byte("my $x = ;"))
	require.Equal(t, BucketNoAnswer, v.Bucket)
	assert.Contains(t, v.Detail, "_term",
		"detail must name the hidden rule so the report can be triaged")
}
