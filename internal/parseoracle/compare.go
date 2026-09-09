// ABOUTME: Buckets our parse against perl's as exact, wider, WRONG, or no-answer.
// ABOUTME: The asymmetry is the point: declining to answer is correct behaviour, committing to a wrong parse is not.

package parseoracle

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/parser"
)

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
)

// Verdict is one comparison result.
type Verdict struct {
	// Bucket is the verdict.
	Bucket Bucket
	// Marker is the parse fact that decided it, empty when nothing was in play.
	Marker Marker
	// Detail is a human-readable reason, for a report and for test failures.
	Detail string
}

// Compare buckets our parse of src against perl's facts for the same source.
//
// It compares one marker: srefgen, perl taking a reference at a call site.
// That is deliberate rather than incomplete -- srefgen is where a prototype
// changes the parse invisibly, which is the measured blind spot this harness
// exists to close. Additional markers (rv2hv, match, readline, anonhash) go in
// as separate rules with their own tests; a whole-tree diff never does, because
// the optree is a lossy, post-optimisation witness of the parse.
//
// src is carried for the markers that need the literal text -- a heredoc body
// or a quote-like operator's delimiter cannot be settled from node kinds alone.
// The srefgen rule deliberately does not use it: see compareSrefgen.
func Compare(facts Facts, tree *parser.Tree, _ []byte) Verdict {
	root := treeRoot(tree)

	// Either side may decline, and a refusal is not a wrong answer.
	if root == nil {
		return Verdict{BucketNoAnswer, MarkerNone, "our parser produced no tree"}
	}
	if root.HasError() {
		return Verdict{BucketNoAnswer, MarkerNone, "our parse contains an error node"}
	}
	// A missing error node is not the same as a parse. The grammar accepts
	// `my $x = ;` -- which perl rejects -- by dropping the right-hand side
	// and reporting no error, so HasError alone lets a tree that is not a
	// parse of this source reach a verdict and score exact. Comparing
	// against source we threw away measures nothing, so decline: no-answer
	// is what "we declined" means, and it keeps the limitation visible in
	// the coverage number instead of inflating the fidelity one.
	if kinds := tree.DegenerateKinds(); len(kinds) > 0 {
		return Verdict{BucketNoAnswer, MarkerNone,
			fmt.Sprintf("our parse dropped source without an error node "+
				"(hidden grammar rule%s surfaced: %s)",
				plural(len(kinds)), strings.Join(kinds, ", "))}
	}
	if !facts.OK {
		// No ground truth means no verdict to reach. Scoring this WRONG
		// would blame our parser for perl's refusal -- and the corpus is
		// full of files that fail for environmental reasons (see Classify).
		return Verdict{BucketNoAnswer, MarkerNone, "perl declined to compile the file"}
	}

	return compareSrefgen(facts, root)
}

// compareSrefgen decides the srefgen marker.
//
// srefgen presence ALONE is not a sufficient marker, and getting this wrong is
// a false-positive generator built into the metric. Perl emits srefgen for both
// of these:
//
//	sub f(\@){} my @a; f(@a)    -- the prototype took the reference
//	sub f{}    my @a; f(\@a)    -- the source took the reference
//
// So a bare "srefgen present but our tree has no reference" check scores every
// explicit f(\@a) in the corpus as WRONG. The discriminator is whether OUR tree
// already accounts for the srefgen: an explicit `\` shows up as a refgen node,
// which we find structurally rather than by scanning the text for a backslash.
func compareSrefgen(facts Facts, root *parser.Node) Verdict {
	explicit, hedged, committed := 0, 0, 0
	walk(root, func(n *parser.Node) {
		switch {
		case isRefgen(n):
			explicit++
		case isHedgedCall(n):
			hedged++
		case isCommittedCall(n):
			committed++
		}
	})

	// Every reference perl took is one our source wrote explicitly. Nothing
	// was resolved behind our back.
	if facts.Srefgen <= explicit {
		return Verdict{BucketExact, markerFor(facts.Srefgen),
			fmt.Sprintf("perl took %d reference(s), all explicit in the source", facts.Srefgen)}
	}

	// Perl took a reference we did not write. Something resolved it -- a
	// prototype, in practice. Whether that is wider or WRONG turns entirely
	// on whether our tree admits it does not know.
	unexplained := facts.Srefgen - explicit
	if hedged > 0 {
		return Verdict{BucketWider, MarkerSrefgen,
			fmt.Sprintf("perl took %d reference(s) the source did not write; "+
				"we emitted %d call(s) marked unresolved rather than committing",
				unexplained, hedged)}
	}
	return Verdict{BucketWrong, MarkerSrefgen,
		fmt.Sprintf("perl took %d reference(s) the source did not write, and our tree "+
			"committed to %d call(s) with no reference and no hedge", unexplained, committed)}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func markerFor(srefgen int) Marker {
	if srefgen > 0 {
		return MarkerSrefgen
	}
	return MarkerNone
}

// isRefgen reports whether a node is the source taking a reference with `\`.
// Matching on the node kind rather than searching src for a backslash is what
// keeps a `\` inside a string or a comment from counting.
func isRefgen(n *parser.Node) bool {
	return n.Kind() == "refgen_expression"
}

// isHedgedCall reports whether a call node names its own uncertainty. The
// grammar emits `ambiguous_function_call_expression` for `f(...)`, whose parse
// genuinely cannot be settled without knowing f's prototype -- the tree is
// saying "unresolved", which is what makes the wider bucket reachable.
func isHedgedCall(n *parser.Node) bool {
	return n.Kind() == "ambiguous_function_call_expression"
}

// isCommittedCall reports whether a call node commits to a definite parse.
// `&f(@a)` is the committed form: it is unambiguously a call with the argument
// list written, so a disagreement here is a disagreement about the parse.
func isCommittedCall(n *parser.Node) bool {
	return n.Kind() == "function_call_expression" || n.Kind() == "method_call_expression"
}

func walk(n *parser.Node, visit func(*parser.Node)) {
	if n == nil {
		return
	}
	visit(n)
	for i := 0; i < n.NamedChildCount(); i++ {
		walk(n.NamedChild(i), visit)
	}
}

func treeRoot(tree *parser.Tree) *parser.Node {
	if tree == nil {
		return nil
	}
	return tree.RootNode()
}
