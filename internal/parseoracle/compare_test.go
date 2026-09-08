// ABOUTME: Tests that the four buckets mean what they say, including the ones that must not fire.
// ABOUTME: Every bucket the API exposes is reachable here, so the enum is not three buckets and a lie.

package parseoracle

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/parser"
)

// compareSource runs the whole pipeline for one snippet: ask perl, parse with
// our parser, compare. Every test here needs all three, and doing it by hand
// each time is how a test ends up asserting against a tree it never built.
func compareSource(t *testing.T, src string) Verdict {
	t.Helper()

	facts, err := Ask(context.Background(), []byte(src), Options{})
	require.NoError(t, err, "the oracle itself must not fail on %q", src)

	tree, err := parser.New().Parse([]byte(src))
	require.NoError(t, err)
	require.NotNil(t, tree)

	return Compare(facts, tree, []byte(src))
}

// TestComparePrototypePair is the pair from the spec's opening paragraph. The
// two snippets have byte-identical call syntax and differ only in a prototype,
// and perl parses them differently. A comparison that cannot tell them apart
// is measuring nothing.
func TestComparePrototypePair(t *testing.T) {
	// Our parser does not track prototypes, so it cannot know that f(@a)
	// passes a reference here. It says so, via the node kind, and that is
	// the wider bucket rather than WRONG -- see TestCompareWiderIsReachable.
	withProto := compareSource(t, "sub f(\\@){} my @a; f(@a);\n")
	assert.NotEqual(t, BucketExact, withProto.Bucket,
		"a prototype-driven reference our parser cannot see must not score exact")
	assert.Equal(t, MarkerSrefgen, withProto.Marker)

	// No prototype, so nothing to miss.
	noProto := compareSource(t, "sub f{} my @a; f(@a);\n")
	assert.Equal(t, BucketExact, noProto.Bucket, noProto.Detail)
}

// TestCompareWrongIsReachable pins the WRONG bucket to a tree that COMMITS to
// a parse perl did not make. The prototype pair above cannot do it: our parser
// hedges there. `&f(@a)` is the committed form -- and note that perl does NOT
// apply the prototype to it, so the roles are reversed and this is a synthetic
// tree rather than real source.
func TestCompareWrongIsReachable(t *testing.T) {
	// Perl took a reference; our tree is a definite, non-hedging call with a
	// plain array argument. That is a commitment to a different parse.
	facts := Facts{OK: true, Srefgen: 1, Entersub: 1}

	src := "sub f{} my @a; &f(@a);\n"
	tree, err := parser.New().Parse([]byte(src))
	require.NoError(t, err)

	v := Compare(facts, tree, []byte(src))
	assert.Equal(t, BucketWrong, v.Bucket, v.Detail)
	assert.Equal(t, MarkerSrefgen, v.Marker)
}

// TestCompareExplicitRefIsExact is the trap the issue names. Perl emits
// srefgen for an explicit f(\@a) exactly as it does for a prototype-driven
// f(@a), so a bare "srefgen present" check scores every explicit reference in
// the corpus as WRONG -- a false-positive generator built into the metric.
func TestCompareExplicitRefIsExact(t *testing.T) {
	v := compareSource(t, "sub f{} my @a; f(\\@a);\n")
	assert.Equal(t, BucketExact, v.Bucket, v.Detail)
}

// TestCompareErrorNodeIsNoAnswer: an honest refusal is not a wrong answer.
// Scoring it WRONG makes the metric punish the behaviour we want.
func TestCompareErrorNodeIsNoAnswer(t *testing.T) {
	v := compareSource(t, "sub f( { }\n")
	assert.Equal(t, BucketNoAnswer, v.Bucket, v.Detail)
	assert.NotEqual(t, BucketWrong, v.Bucket)
}

// TestCompareOracleDeclinedIsNoAnswer: the refusal can come from either side.
// When perl will not compile the file there is no ground truth to compare
// against, so there is no verdict to reach -- not a WRONG one.
func TestCompareOracleDeclinedIsNoAnswer(t *testing.T) {
	facts := Facts{OK: false, Stderr: "syntax error at - line 1"}
	src := "my @a;\n"
	tree, err := parser.New().Parse([]byte(src))
	require.NoError(t, err)

	v := Compare(facts, tree, []byte(src))
	assert.Equal(t, BucketNoAnswer, v.Bucket, v.Detail)
}

// TestCompareIgnoresConstantFolding: the optree is post-peephole. `my $x = 1+2`
// arrives as const[IV 3] s/FOLD with no add op, so a comparison that expected
// the parser's ops to survive would report the optimiser's work as our bug.
func TestCompareIgnoresConstantFolding(t *testing.T) {
	facts, err := Ask(context.Background(), []byte("my $x = 1+2;\n"), Options{})
	require.NoError(t, err)
	// Guard the premise: if perl ever stops folding this, the test below
	// proves nothing and should be rewritten rather than silently pass.
	assert.NotContains(t, facts.Ops, "add",
		"premise of this test is that folding removed the add op")

	v := compareSource(t, "my $x = 1+2;\n")
	assert.Equal(t, BucketExact, v.Bucket, v.Detail)
}

// TestCompareIgnoresOpCounts asserts the comparison is blind to op counts and
// to op order. Feeding it wildly different Ops and OpCount for the same tree
// must not move the verdict: only marker presence may.
func TestCompareIgnoresOpCounts(t *testing.T) {
	src := "sub f{} my @a; f(@a);\n"
	tree, err := parser.New().Parse([]byte(src))
	require.NoError(t, err)

	base := Facts{OK: true, Entersub: 1, Ops: []string{"enter", "nextstate", "padav", "gv", "entersub", "leave"}}

	// Same markers, absurd op sequence and count.
	noisy := base
	noisy.Ops = []string{"leave", "entersub", "padav", "aelem", "helem", "add", "concat", "enter"}
	noisy.OpCount = 512

	// Same markers, no ops reported at all.
	empty := base
	empty.Ops = nil
	empty.OpCount = 0

	want := Compare(base, tree, []byte(src))
	for name, facts := range map[string]Facts{"noisy": noisy, "empty": empty} {
		got := Compare(facts, tree, []byte(src))
		assert.Equal(t, want.Bucket, got.Bucket,
			"%s: op counts and sequence must not change the verdict (%s)", name, got.Detail)
	}
}

// TestCompareWiderIsReachable pins the bucket the issue told us to define or
// delete. It is reachable because our grammar has a node kind that names its
// own uncertainty: `ambiguous_function_call_expression`. Perl resolved the
// prototype; we emitted a call marked unresolved. Less specific than perl, and
// not a commitment to anything else -- which is exactly what wider means.
func TestCompareWiderIsReachable(t *testing.T) {
	v := compareSource(t, "sub f(\\@){} my @a; f(@a);\n")
	assert.Equal(t, BucketWider, v.Bucket, v.Detail)

	// The distinction from WRONG is the hedging node kind, nothing else. Show
	// that by pointing the same facts at the committed form.
	require.NotEqual(t, BucketWider, compareSource(t, "sub f{} my @a; &f(@a);\n").Bucket,
		"the committed call form must not also land in wider")
}

// TestCompareAllBucketsReachable is the AC that keeps the enum honest: a
// four-bucket API that can only produce three buckets is a lie in the type
// system. Each bucket here is produced by a distinct input.
func TestCompareAllBucketsReachable(t *testing.T) {
	seen := map[Bucket]string{}

	// exact: no prototype, nothing for us to miss.
	seen[compareSource(t, "sub f{} my @a; f(@a);\n").Bucket] = "no-prototype call"

	// wider: perl resolved a prototype, we emitted a hedging call node.
	seen[compareSource(t, "sub f(\\@){} my @a; f(@a);\n").Bucket] = "prototype-driven reference"

	// no-answer: our parser produced an error node.
	seen[compareSource(t, "sub f( { }\n").Bucket] = "error node"

	// WRONG: perl took a reference, our tree committed to a plain call.
	src := "sub f{} my @a; &f(@a);\n"
	tree, err := parser.New().Parse([]byte(src))
	require.NoError(t, err)
	seen[Compare(Facts{OK: true, Srefgen: 1, Entersub: 1}, tree, []byte(src)).Bucket] = "committed non-ref call"

	for _, b := range []Bucket{BucketExact, BucketWider, BucketWrong, BucketNoAnswer} {
		assert.Contains(t, seen, b, "bucket %s is defined but no input reaches it", b)
	}
}

// TestCompareBucketStrings keeps the bucket names stable, since the report the
// next issue writes quotes them and a renamed bucket silently changes a metric.
func TestCompareBucketStrings(t *testing.T) {
	assert.Equal(t, "exact", BucketExact.String())
	assert.Equal(t, "wider", BucketWider.String())
	assert.Equal(t, "WRONG", BucketWrong.String())
	assert.Equal(t, "no-answer", BucketNoAnswer.String())
}
