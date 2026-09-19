// ABOUTME: Sites sharing a statement's first line are ONE statement, however far each node ends.
// ABOUTME: Keying a statement by its end line shattered one into many, each holding one reference.

package parseoracle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStatementSpanCoversItsOps: many references in one multi-line statement
// all count.
//
// `Sites` reports each site from the STATEMENT's first line to its OWN
// node's last line, and both ends are load-bearing -- sites.go:83-101 cites
// the corpus cases for each. So a statement holding several references
// emits several sites that share a start and differ in end:
//
//	for (                      # line 214, op/ref.t
//	    [ 'undef', \undef  ],  # 215
//	    [ 'IV',    \1      ],  # 216
//	    ...
//
//	line=214 end=215  took_ref
//	line=214 end=216  took_ref
//	... fifteen of them
//
// groupByStatement keyed on [start, end], so those became FIFTEEN statement
// records of one reference each. `innermost` then picks a single winner --
// the one ending earliest -- whose one reference is spent on perl's first
// site, and the remaining fourteen find it empty and score WRONG.
//
// Measured at afdd06c9: op/ref.t:214 alone produced 15 of the 88 WRONG
// srefgen sites, and the same shape appears in op/pack.t, op/lvref.t,
// re/reg_mesg.t, re/reg_pmod.t, run/switches.t and op/sprintf2.t across
// every marker -- about 70 of the 292 WRONG sites in the corpus.
func TestStatementSpanCoversItsOps(t *testing.T) {
	// Three references in one statement beginning at line 214, each node
	// ending on its own line.
	subject := subjectWith(
		SubjectCallSite{Line: 214, EndLine: 215, TookReference: true},
		SubjectCallSite{Line: 214, EndLine: 216, TookReference: true},
		SubjectCallSite{Line: 214, EndLine: 217, TookReference: true},
	)

	// Perl attributes all three to the statement's nextstate line.
	v := CompareFacts(withRefsAt(214, 214, 214), subject)

	assert.Equal(t, BucketExact, v.Bucket,
		"three references in one statement should account for perl's three: %s", v.Detail)
	assert.Empty(t, v.Findings, "nothing should be left unaccounted for")
}

// TestNestedStatementStillOwnsItsLines: the fix must not break the case the
// end line exists for.
//
// A `\` inside a sub body is reported at the body's line, because perl's
// probe walks the CVs and attributes it there. That site must stay separate
// from the enclosing statement's -- pooling it under `sub foo {`'s whole
// span is what sites.go:83-95 says must not happen.
func TestNestedStatementStillOwnsItsLines(t *testing.T) {
	outer := SubjectCallSite{Line: 100, EndLine: 112, TookReference: true}
	inner := SubjectCallSite{Line: 105, EndLine: 105, Unresolved: true}

	// Perl's site at 105 belongs to the INNER statement, which hedged.
	v := CompareFacts(withRefsAt(105), subjectWith(outer, inner))
	assert.Equal(t, BucketWider, v.Bucket,
		"the inner statement hedged line 105 and owns it: %s", v.Detail)

	// And the outer statement's own reference still accounts for 112.
	v = CompareFacts(withRefsAt(105, 112), subjectWith(outer, inner))
	assert.Equal(t, BucketWider, v.Bucket, v.Detail)
	assert.Contains(t, v.Detail, "105")
	assert.NotContains(t, v.Detail, "112")
}

// TestOneReferenceStillAccountsForOne: merging must not hand a statement
// more references than it reported.
//
// Two of perl's sites against one backslash is still a deficit; the merge
// changes which RECORD a site lands in, never how many a record holds.
func TestOneReferenceStillAccountsForOne(t *testing.T) {
	subject := subjectWith(
		SubjectCallSite{Line: 100, EndLine: 112, TookReference: true},
		SubjectCallSite{Line: 100, EndLine: 112}, // a committed call, no reference
	)
	v := CompareFacts(withRefsAt(100, 112), subject)
	assert.Equal(t, BucketWrong, v.Bucket,
		"one backslash cannot account for two references: %s", v.Detail)
}

// TestSitesSpanMultilineStatement records what the cause turned out to be,
// because the issue proposed two and it was neither.
//
// The issue asked whether (a) the subject's span fails to contain perl's
// line, or (b) the subject emits many sites where one is meant. Dumping
// op/ref.t showed the subject reporting exactly what it should:
//
//	kind=reference line=214 end=215 took_ref=true
//	kind=reference line=214 end=216 took_ref=true
//	... fifteen, one per backslash, every Line matching perl's
//
// Both hypotheses were wrong. The sites are right and the GROUPING was
// wrong: keying a statement record on [start, end] made each of those a
// statement of its own, because each node ends on its own line.
//
// Asserted as the invariant rather than against one file: sites sharing a
// start line land in one record, and that record's end covers the furthest
// of them.
func TestSitesSpanMultilineStatement(t *testing.T) {
	stmts := groupByStatement([]SubjectCallSite{
		{Line: 214, EndLine: 215, TookReference: true},
		{Line: 214, EndLine: 230, TookReference: true},
		{Line: 214, EndLine: 220, TookReference: true},
	})

	if assert.Len(t, stmts, 1, "three sites at line 214 are ONE statement") {
		assert.Equal(t, 214, stmts[0].start)
		assert.Equal(t, 230, stmts[0].end, "the record covers the furthest node")
		assert.Equal(t, 3, stmts[0].decided[MarkerSrefgen],
			"all three references belong to it")
	}

	// A different start line is still a different statement.
	stmts = groupByStatement([]SubjectCallSite{
		{Line: 214, EndLine: 220, TookReference: true},
		{Line: 216, EndLine: 216, TookReference: true},
	})
	assert.Len(t, stmts, 2, "different start lines are different statements")
}
