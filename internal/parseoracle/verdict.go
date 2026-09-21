// ABOUTME: The verdict vocabulary: four buckets, five markers, and the findings behind them.
// ABOUTME: Portable by construction — nothing here knows what parsed the file.

package parseoracle

import "fmt"

// Bucket is how our parse compares to perl's.
//
// The four are not symmetric, and that asymmetry is the whole reason there are
// four rather than a pass/fail. A parser that says "I don't know" is behaving
// correctly; a parser that commits to a parse perl did not make is lying to
// every downstream consumer. Collapsing those two into one failure count makes
// the metric punish the behaviour we want. This mirrors the Unknown-versus-Any
// decision already made in internal/types.
type Bucket int

const (
	// BucketExact means our tree implies the same parse perl made.
	BucketExact Bucket = iota

	// BucketWider means we were LESS specific than perl and not wrong: our
	// tree hedges where perl resolved. Tracked, never a build failure, but
	// it must not grow silently -- declaring every call unresolved scores
	// 100% non-WRONG and is useless, so a report must assert a floor on
	// exact and not only a ceiling on WRONG.
	BucketWider

	// BucketWrong means we committed to a parse perl did not make. This is
	// the only bucket that fails a build. It is worse than saying nothing,
	// because a consumer acts on it.
	BucketWrong

	// BucketNoAnswer means somebody declined: we produced an error node, or
	// perl would not compile the file so there is no ground truth to
	// compare against. This is the coverage metric, a ratchet rather than a
	// gate.
	BucketNoAnswer
)

func (b Bucket) String() string {
	switch b {
	case BucketExact:
		return "exact"
	case BucketWider:
		return "wider"
	case BucketWrong:
		return "WRONG"
	case BucketNoAnswer:
		return "no-answer"
	default:
		return fmt.Sprintf("Bucket(%d)", int(b))
	}
}

// Marker names which parse fact drove the verdict.
//
// Comparison is by MARKER PRESENCE, never by op count or op sequence. The
// optree perl hands back is post-peephole: `my $x = 1+2` arrives as
// `const[IV 3] s/FOLD` with no `add` op left, so a comparison against counts
// or sequences would report the optimiser's differences as the parser's.
type Marker string

const (
	// MarkerNone means no marker op was in play.
	MarkerNone Marker = ""

	// MarkerSrefgen is perl taking a reference at a call site. It is the
	// highest-value parse fact available, because a prototype changes the
	// parse at every call site and a static parser cannot know it without
	// having already seen the definition.
	MarkerSrefgen Marker = "srefgen"

	// MarkerRv2hv is perl reading a hash: `%$r`, `%{...}`, `->%*`, a hash
	// slice through a reference, a global `%h`, or an element with a key
	// too complex to fold. It is the "is `%` a sigil or a modulus" question
	// of spec §7.1.2. Measured: a lexical `%h` is padhv and `$h{a}` folds
	// to multideref, so perl emits FEWER of these than the source has hash
	// accesses, never more -- which is the direction presence-matching
	// tolerates.
	MarkerRv2hv Marker = "rv2hv"

	// MarkerMatch is perl matching a regex: `/.../`, `m//`, and `=~` against
	// any pattern that is not s/// or tr///. The "is `/` a match or a
	// divide" question. Measured: `split /,/` compiles to a split op with
	// no match op, and a match under `if (0)` is discarded, so again perl
	// emits at most as many as the source has.
	MarkerMatch Marker = "match"

	// MarkerReadline is perl reading a line: `<FH>`, `<$fh>`, `<>`, `<<>>`
	// and the readline builtin. The "did `<...>` read or glob" question.
	// Measured: `<*.c>` and `<${fh}>` are glob; `$x .= <FH>` is rewritten
	// to rcatline, which the oracle folds back into this marker.
	MarkerReadline Marker = "readline"

	// MarkerAnonhash is perl building an anonymous hash: `{ a => 1 }`, or
	// `{}` as emptyavhv flagged OPpEMPTYAVHV_IS_HV. The "is `{` a block or
	// a hashref" question. A block emits no marker at all.
	MarkerAnonhash Marker = "anonhash"
)

// Markers is every marker the harness decides, in the order a verdict
// names them when more than one is in play. It is the table spec §7.5.4
// sketches, with the tree-sitter predicates living in adapter.go.
var Markers = []Marker{MarkerSrefgen, MarkerRv2hv, MarkerMatch, MarkerReadline, MarkerAnonhash}

// Verdict is one comparison result.
type Verdict struct {
	// Bucket is the verdict.
	Bucket Bucket
	// Marker is the parse fact that decided it, empty when nothing was in play.
	Marker Marker
	// Detail is a human-readable reason, for a report and for test failures.
	Detail string

	// Findings are the individual sites that fell short, across EVERY marker
	// rather than only the one that decided Bucket.
	//
	// The attribution already happens per statement -- decideMarker finds the
	// innermost statement covering each of perl's sites -- and keeping it is
	// what lets a report name work. A bucket says "this file is WRONG"; a
	// finding says which line, which marker, and whether the subject hedged
	// or committed. Measured over the 620-file corpus: 72 WRONG files hold
	// 249 sites, and the ten worst files hold 145 of them.
	//
	// Empty for an exact verdict: there is nothing to name.
	//
	// This is DIAGNOSIS, not scoring. Bucket is unchanged by its presence,
	// and the gate still reads Bucket alone -- a per-statement bucket would
	// need a threshold, which is the thing ratchets exist to avoid.
	Findings []Finding
}

// Finding is one site perl reported that the subject did not account for.
type Finding struct {
	// Line is perl's line for the site, inside the statement that owns it.
	Line int
	// Marker is the parse fact perl reported there.
	Marker Marker
	// Bucket is BucketWider when the subject hedged the statement and
	// BucketWrong when it committed with no site and no hedge. Never exact:
	// an accounted-for site is not a finding.
	Bucket Bucket
}

// plural is "" for one and "s" for any other count.
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
