// ABOUTME: Comparison between perl's facts and a subject's facts, with no syntax tree involved.
// ABOUTME: This is the portable core; compare.go adapts our tree-sitter parser onto it.

package parseoracle

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

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

	// The walk IS the reference measurement. If perl compiled the file but
	// the probe that walks its CVs never ran, Srefgen and RefLines are absent
	// rather than zero, and reading absence as "no references" is the same
	// defect a main-program-only count had, wearing a different hat.
	if !oracle.Walked {
		return Verdict{BucketNoAnswer, MarkerNone,
			"perl compiled the file but the oracle's walk of its CVs did not run, " +
				"so the reference question is unanswered"}
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

// compareReferences is the srefgen rule, restated over facts and decided one
// statement at a time.
//
// perl emits srefgen for an explicit f(\@a) exactly as it does for a
// prototype-driven f(@a), so "srefgen present" alone would score every
// explicit reference WRONG. The discriminator is whether the SUBJECT already
// accounts for the reference: it reports took_reference at a call site when
// its own parse committed to passing a reference there.
//
// The unit of comparison is the statement, keyed by the line it starts on,
// because that is the finest attribution perl offers: every op belongs to the
// nearest preceding nextstate, and a nextstate names its statement's first
// line. Both sides report sites in that unit (see RefLines and
// treeSitterCallSites), which is what makes the two populations comparable
// at all. A whole-file total could not be: it let a surplus at one statement
// cancel a deficit at another, and let one hedge anywhere in the file excuse
// every unexplained reference in it (srefgen=5, hedged=1, committed=4 scored
// wider).
//
// At each statement:
//
//   - perl took more references than the subject did, and the subject hedged
//     a call THERE: wider. Something resolved the parse behind a static
//     parser's back -- a prototype, in practice -- and the subject said so.
//   - perl took more, and the subject committed to every call there: WRONG.
//     It committed to a parse perl did not make.
//   - the subject took as many or more: agreement. A subject surplus is a
//     backslash perl represented as something other than a srefgen/refgen
//     op -- `\1` and `\"x"` fold to a constant holding the reference -- and
//     under per-statement scoring it can mask nothing beyond its own
//     statement, so it is not the unanswerable question it was under totals.
//
// The file is WRONG if any statement is, wider if any is, exact otherwise.
//
// One ceiling remains, and it is the same statement that both bounds it and
// names it: a folded `\1` on the SAME statement as a prototype-driven
// reference the subject missed sums to zero there, exactly as the whole-file
// version did across the file. Perl attributes nothing finer than a
// statement, so closing it means the subject not reporting a reference for a
// literal operand, which is a change to the adapter and not to this rule.
//
// A hedge is deliberately consulted only where there is a deficit for it to
// explain. Our grammar marks every unqualified `f(...)` call unresolved, so
// hedges are near-universal and say nothing on their own; treating one as
// disagreement would score `sub f{} f(@a)` -- where both sides say "no
// reference" -- as wider, which is not a hedge about anything.
//
// Lines are part of the contract. A site reported without one (line 0) can
// match no statement perl attributed a reference to, so a took_reference
// there accounts for nothing and a hedge there explains nothing.
func compareReferences(oracle Facts, subject SubjectFacts) Verdict {
	took, hedged := map[int]int{}, map[int]int{}
	for _, c := range subject.CallSites {
		switch {
		case c.TookReference:
			took[c.Line]++
		case c.Unresolved:
			hedged[c.Line]++
		}
	}
	perl := map[int]int{}
	for _, line := range oracle.RefLines {
		perl[line]++
	}

	var wider, wrong []int
	unexplained := 0
	for _, line := range sortedKeys(perl) {
		deficit := perl[line] - took[line]
		if deficit <= 0 {
			continue
		}
		unexplained += deficit
		if hedged[line] > 0 {
			wider = append(wider, line)
		} else {
			wrong = append(wrong, line)
		}
	}

	switch {
	case len(wrong) > 0:
		return Verdict{BucketWrong, MarkerSrefgen,
			fmt.Sprintf("perl took %d reference(s) the subject did not, and at %s the "+
				"subject committed to its calls with no reference and no hedge",
				unexplained, lineList(wrong))}
	case len(wider) > 0:
		return Verdict{BucketWider, MarkerSrefgen,
			fmt.Sprintf("perl took %d reference(s) the subject did not; at %s it "+
				"marked call(s) unresolved rather than committing",
				unexplained, lineList(wider))}
	default:
		return Verdict{BucketExact, markerFor(oracle.Srefgen),
			fmt.Sprintf("perl took %d reference(s), all accounted for by the subject",
				oracle.Srefgen)}
	}
}

func sortedKeys(m map[int]int) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// lineList renders statement lines for a verdict's detail: "line 5" or
// "lines 2, 7, 9". The report is triaged by these, so they are named rather
// than counted.
func lineList(lines []int) string {
	parts := make([]string, len(lines))
	for i, l := range lines {
		parts[i] = strconv.Itoa(l)
	}
	return fmt.Sprintf("line%s %s", plural(len(lines)), strings.Join(parts, ", "))
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
