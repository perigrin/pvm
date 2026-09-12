// ABOUTME: Comparison between perl's facts and a subject's facts, with no syntax tree involved.
// ABOUTME: This is the portable core; compare.go adapts our tree-sitter parser onto it.

package parseoracle

import "fmt"

// CompareFacts reaches a verdict from two sets of facts about the same file:
// perl's, and the subject's. Neither side is a syntax tree, which is what lets
// a subject written in any language be measured.
//
// The bucketing rules are the ones compare.go established and are deliberately
// unchanged here; only the source of the subject's side is different. A
// refactor that moved the numbers would be a bug, not a refactor.
func CompareFacts(oracle Facts, subject SubjectFacts) Verdict {
	// Either side may decline, and a refusal is never a wrong answer. This is
	// the asymmetry the whole four-bucket design exists for: a parser that
	// says "I don't know" is behaving correctly, and scoring it WRONG would
	// make the metric punish the behaviour we want.
	if subject.Declined {
		detail := subject.DeclinedReason
		if detail == "" {
			detail = "the subject declined to produce a parse"
		}
		return Verdict{BucketNoAnswer, MarkerNone, detail}
	}
	if !subject.OK {
		return Verdict{BucketNoAnswer, MarkerNone, "the subject rejected the file"}
	}
	if !oracle.OK {
		// No ground truth means no verdict to reach. Scoring this WRONG would
		// blame the subject for perl's refusal, and the corpus is full of
		// files that fail for environmental reasons (see Classify).
		return Verdict{BucketNoAnswer, MarkerNone, "perl declined to compile the file"}
	}

	// Success with no optree is not a parse. A corpus file that calls
	// skip_all inside a BEGIN block exits DURING compilation, so perl reports
	// ok:1 having built nothing:
	//
	//	uni/greek.t  ->  ok: 1   op_count: 0   srefgen: 0
	//
	// Our parser meanwhile produces a whole tree for the same file, finds no
	// references in it, and the two zero totals match trivially -- so the file
	// scored exact against a parse that never happened. Ten corpus files
	// report this shape and six of them were being counted as agreement.
	//
	// A genuinely empty Perl file has no optree either, and both sides would
	// honestly agree there is nothing there. It is NOT distinguished here, and
	// that is deliberate: the corpus contains no such file (measured: all ten
	// no-optree files have source, the smallest 647 bytes), so a rule to
	// separate them would be untested code guarding a case that does not
	// arise. Source length is also the wrong discriminator -- it cannot tell
	// an empty file from one that is entirely comments, and both are the same
	// honest zero. Declining on a truly empty file is a small, honest loss;
	// scoring a skipped file is a lie.
	//
	// This is NOT a comparison of op counts, which compare.go forbids for good
	// reason: the optree is post-peephole, so `my $x = 1+2` loses its add op
	// and any count-based rule reports the optimiser's work as the parser's.
	// The question here is whether an optree exists at all. Zero is not a
	// point on that spectrum -- it is perl telling us it never got far enough
	// to build one.
	if noOptree(oracle) {
		return Verdict{BucketNoAnswer, MarkerNone,
			"perl reported success but produced no optree, so it never finished " +
				"parsing the file and there is nothing to compare against"}
	}

	// An unanswered question is not agreement. A subject that omits call_sites
	// has told us nothing about references, and treating silence as "no
	// references taken" would let it score exact by staying quiet.
	if !subject.KnowsCallSites {
		return Verdict{BucketNoAnswer, MarkerNone,
			"the subject did not report call sites, so the reference question is unanswered"}
	}

	return compareReferences(oracle, subject)
}

// compareReferences is the srefgen rule, restated over facts.
//
// perl emits srefgen for an explicit f(\@a) exactly as it does for a
// prototype-driven f(@a), so "srefgen present" alone would score every
// explicit reference WRONG. The discriminator is whether the SUBJECT already
// accounts for the reference: it reports took_reference at a call site when
// its own parse committed to passing a reference there.
func compareReferences(oracle Facts, subject SubjectFacts) Verdict {
	accounted, hedged := 0, 0
	for _, c := range subject.CallSites {
		if c.TookReference {
			accounted++
		} else if c.Unresolved {
			hedged++
		}
	}

	// The two sides do not count the same population, and when the subject's
	// total runs ahead of perl's that difference is not agreement -- it is a
	// measurement we cannot make. Measured, the split is by syntactic form:
	//
	//	\@a  \%h  \&f  \$x  \(&f)   -> srefgen
	//	\(@a)  \(@a,@b)  \($x,$y)   -> refgen
	//
	// so a list-form `\( ... )` contributes to `accounted` and not to
	// Srefgen. Counting refgen too would not fix it, because the units differ
	// as well as the population: `\(@a,@b,@c)` is ONE refgen for three
	// references, while a `f(\@\@)` prototype is TWO srefgen for no backslash
	// at all. Only a per-call-site attribution could reconcile them, and the
	// oracle does not attribute srefgen to call sites.
	//
	// So a surplus means the question is unanswerable rather than answered
	// yes, and the surplus is large enough to swallow a prototype-driven
	// reference whole. Scoring it exact is how a real disagreement hides.
	if accounted > oracle.Srefgen {
		return Verdict{BucketNoAnswer, MarkerNone,
			fmt.Sprintf("the subject took %d reference(s) to perl's %d srefgen; "+
				"the two counts are of different populations, so the comparison "+
				"is not meaningful", accounted, oracle.Srefgen)}
	}

	// Every reference perl took is one the subject also took. Nothing was
	// resolved behind its back.
	//
	// A hedge is deliberately NOT consulted here, and the reason is a property
	// of the subject rather than of the rule. Our grammar marks every
	// unqualified `f(...)` call unresolved, so `hedged > 0` holds for almost
	// every file in the corpus including ones where perl took no reference at
	// all. Treating that as disagreement would score `sub f{} f(@a)` -- where
	// both sides say "no reference" -- as wider, which is not a hedge about
	// anything. A hedge is only evidence when there is an unexplained
	// reference for it to explain, which is the branch below.
	if oracle.Srefgen == accounted {
		return Verdict{BucketExact, markerFor(oracle.Srefgen),
			fmt.Sprintf("perl took %d reference(s), all accounted for by the subject",
				oracle.Srefgen)}
	}

	// Perl took a reference the subject did not. Something resolved it -- a
	// prototype, in practice. Whether that is wider or WRONG turns entirely on
	// whether the subject admits it does not know.
	unexplained := oracle.Srefgen - accounted
	if hedged > 0 {
		return Verdict{BucketWider, MarkerSrefgen,
			fmt.Sprintf("perl took %d reference(s) the subject did not; "+
				"it marked %d call(s) unresolved rather than committing",
				unexplained, hedged)}
	}
	return Verdict{BucketWrong, MarkerSrefgen,
		fmt.Sprintf("perl took %d reference(s) the subject did not, and the subject "+
			"committed to its calls with no reference and no hedge", unexplained)}
}

// noOptree reports that perl compiled successfully without building an optree,
// which is how a file that exits during compilation looks from out here.
//
// Both signals are required, and requiring both is what keeps this from
// becoming the op-count comparison compare.go forbids. perl reports the op
// list and its length together; a file that really was parsed has ops. A
// caller that supplies ops while claiming a zero count is describing an
// interpreter that does not exist, and is trusted rather than declined.
func noOptree(oracle Facts) bool {
	return oracle.OpCount == 0 && len(oracle.Ops) == 0
}
