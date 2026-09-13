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
// The unit of comparison is the statement, because that is the finest
// attribution perl offers: every op belongs to the nearest preceding
// nextstate. The subject reports each site with its statement's span
// (Line..EndLine); perl reports each reference with the one line its
// nextstate recorded, and which line of the statement that is depends on
// what the statement contains -- the first line for a plain call, the last
// when a block-bearing term precedes the reference (op/qr.t line 112,
// measured) -- so a perl site belongs to the innermost subject statement
// whose span contains its line. A whole-file total could not do this: it let
// a surplus at one statement cancel a deficit at another, and let one hedge
// anywhere in the file excuse every unexplained reference in it (srefgen=5,
// hedged=1, committed=4 scored wider).
//
// For each reference perl took, in the statement that owns it:
//
//   - the subject has a backslash there not yet matched: agreement, and
//     that backslash is spent.
//   - otherwise the subject hedged a call there: wider. Something resolved
//     the parse behind a static parser's back -- a prototype, in practice --
//     and the subject said so.
//   - otherwise: WRONG. The subject committed to a parse perl did not make,
//     or -- the case op/filetest.t and mro/package_aliases.t showed -- its
//     tree stopped being a parse of the source and no error node said so,
//     which is the same thing from where this rule stands.
//
// A subject surplus -- a backslash perl represented as something other than
// a srefgen/refgen op, such as `\1` folded to a constant -- is never
// consulted, and under per-statement scoring it can mask nothing beyond its
// own statement, so it no longer forces no-answer as it did under totals.
//
// The file is WRONG if any reference is, wider if any is, exact otherwise.
//
// One ceiling remains, and it is the statement itself: a folded `\1` in the
// SAME statement as a prototype-driven reference the subject missed spends
// itself on that reference, exactly as the whole-file version did across
// the file. Perl attributes nothing finer than a statement, so closing it
// means the subject not reporting a reference for a literal operand, which
// is a change to the adapter and not to this rule.
//
// A hedge is deliberately consulted only where there is a deficit for it to
// explain. Our grammar marks every unqualified `f(...)` call unresolved, so
// hedges are near-universal and say nothing on their own; treating one as
// disagreement would score `sub f{} f(@a)` -- where both sides say "no
// reference" -- as wider, which is not a hedge about anything.
//
// Lines are part of the contract. A site reported without one (line 0)
// spans no line perl attributed a reference to, so a took_reference there
// accounts for nothing and a hedge there explains nothing.
func compareReferences(oracle Facts, subject SubjectFacts) Verdict {
	stmts := groupByStatement(subject.CallSites)

	var wider, wrong, unowned []int
	refs := append([]int(nil), oracle.RefLines...)
	sort.Ints(refs)
	for _, line := range refs {
		owner := innermost(stmts, line)
		switch {
		case owner != nil && owner.took > 0:
			owner.took--
		case owner != nil && owner.hedged > 0:
			wider = append(wider, line)
		case len(subject.CallSites) == 0:
			// The subject reported no sites AT ALL, so it has not committed to
			// a parse anywhere -- it has nothing here to be wrong about.
			// `s/x/\@a/e` gives perl an srefgen and a static parser no call
			// site to attach it to; a faithful external subject reporting an
			// empty call_sites list for a statement with no call is in the
			// same position. Silence is a gap in what the subject can see, so
			// it scores no-answer. Only a claim can be WRONG.
			//
			// Reporting sites but none covering this line is NOT silence: the
			// subject described this file and its description omits a
			// reference perl took, which is a claim about the parse. That
			// stays WRONG, including a site that names no statement at all --
			// a site without a line spans nothing and explains nothing.
			unowned = append(unowned, line)
		default:
			wrong = append(wrong, line)
		}
	}
	unexplained := len(wider) + len(wrong) + len(unowned)

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
	case len(unowned) > 0:
		return Verdict{BucketNoAnswer, MarkerSrefgen,
			fmt.Sprintf("perl took %d reference(s) the subject did not, and at %s it "+
				"reported no call the reference could belong to",
				unexplained, lineList(unowned))}
	default:
		return Verdict{BucketExact, markerFor(oracle.Srefgen),
			fmt.Sprintf("perl took %d reference(s), all accounted for by the subject",
				oracle.Srefgen)}
	}
}

// statement is one span of the subject's source and what it reported there.
type statement struct {
	start, end   int
	took, hedged int
}

// groupByStatement pools the subject's sites by the statement they sit in.
// Two sites with the same span are the same statement, so a backslash and a
// call on one line share a pool, which is what lets the backslash account
// for the call's reference.
func groupByStatement(sites []SubjectCallSite) []*statement {
	var stmts []*statement
	byKey := map[[2]int]*statement{}
	for _, c := range sites {
		end := c.EndLine
		if end < c.Line {
			end = c.Line
		}
		key := [2]int{c.Line, end}
		s := byKey[key]
		if s == nil {
			s = &statement{start: c.Line, end: end}
			byKey[key] = s
			stmts = append(stmts, s)
		}
		switch {
		case c.TookReference:
			s.took++
		case c.Unresolved:
			s.hedged++
		}
	}
	return stmts
}

// innermost picks the statement that owns a line: the one containing it
// that starts last and, among those, ends first. A statement nested inside
// a multi-line one owns its own lines; the outer statement owns the rest.
func innermost(stmts []*statement, line int) *statement {
	var best *statement
	for _, s := range stmts {
		if line < s.start || line > s.end {
			continue
		}
		if best == nil || s.start > best.start || (s.start == best.start && s.end < best.end) {
			best = s
		}
	}
	return best
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
