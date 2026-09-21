---
title: The parse oracle measures tree-sitter, not the parser M1 is building
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0ad3f-e929-77a8-9b00-2a80b06ec42c
blocks:
- 01a0a8cd-8fcc-72a8-9ba1-c916eee9992d
created: 2026-09-19T00:00:00Z
updated: 2026-09-21T00:00:00Z
---

The parse oracle grades the parser M1 exists to replace. `run.go:381` calls
`parser.New()` -- tree-sitter -- so every verdict in the 620-file corpus is a
statement about the OLD grammar, and M1's progress is invisible to it.

## Measured, 2026-09-19 at 528b9d83

    run.go:381                     parser.New()   (tree-sitter)
    grep -c '^WRONG' baseline.txt  2
    WRONG rows                     comp/our.t (8), op/universal.t (16)
    TestCorpusBaselineIsIntact     PASSES, with the pin

M1 built its own ratchet at `internal/parse/testdata/t1.ratchet` rather than
repointing this harness, so the two measure different parsers and neither
knows about the other.

## The seam already exists

`askSubject` (run.go:371-385) has two paths, and only the fallback is pinned
to tree-sitter:

    if opts.Subject != nil { ... }          // external subject binary
    tree, err := parser.New().Parse(src)    // <- line 381
    return TreeSitterSubject(tree), nil

The corpus, the verdict vocabulary, the buckets and the ratchet are all
parser-agnostic. What is missing is a `ParseSubject` alongside
`TreeSitterSubject` that answers `SubjectFacts` from an M1 tree.

## What SubjectFacts demands

The oracle does not ask "did it parse". It compares five markers per file
against perl's own answer, so the adapter has to answer in the same terms
(subject.go:31-64):

1. **`OK`** -- M1 reports this already.
2. **`Declined`** / **`DeclinedReason`** -- M1's `Unknown` node is the
   equivalent of tree-sitter's degenerate node. The field comment anticipates
   exactly this: "every parser can answer that question in its own terms."
3. **`CallSites`** -- each call and what it decided (`srefgen`, `rv2hv`,
   `match`, `readline`, `anonhash`), plus `SiteKindReference` for a backslash
   that is not a call. `sites_decided_test.go` is the start of this but is
   not wired to the oracle's vocabulary.
4. **`Prototypes`** -- needs cross-file resolution, which is 01a0ad3f. This
   is the blocker, not a nicety: without it every imported prototype reads as
   unresolved.
5. **`KnowsCallSites`** / **`KnowsPrototypes`** -- set honestly or an M1 that
   cannot answer is scored as AGREEING with perl. See CompareFacts. Getting
   this wrong flatters the numbers instead of failing.

## Why this is its own issue

01a0a8cd ("Remove the WRONG pin") assumed this repoint as route 1 and waits
on it. It cannot close first: both WRONG rows are still present because they
are tree-sitter defects and tree-sitter is still what gets measured. Deleting
the pin today turns a green gate red for two defects that are really there.

Bundling the repoint into a pin-removal is how that pin got stuck in the
first place -- a69d8f1c left the gate deliberately red, and a permanently-red
test is one nobody reads.

## Acceptance

- [ ] `askSubject`'s default subject is the M1 parser, not `parser.New()`
- [ ] A `ParseSubject` adapter answers all five `SubjectFacts` markers from an
      M1 tree, with `KnowsCallSites`/`KnowsPrototypes` set honestly
- [ ] The corpus is re-measured and the baseline regenerated in the same
      commit, with the bucket movement reported in BYTES as well as counts
- [ ] Whether `comp/our.t` and `op/universal.t` still score WRONG is measured
      and recorded -- findings says the M1 lexer already reads `if (/TIE/)` as
      `Quote("/TIE/")` rather than division, so comp/our.t's cause is fixed
- [ ] Findings §0.11 records which parser the harness measures

Findings: §0.11. Blocks 01a0a8cd. Blocked by 01a0ad3f (prototypes need
cross-file resolution before they can be answered at all).


## MEASURED 2026-09-19 at c3e65b4f: do not repoint yet

Both subjects run over the same 620-file shim, same corpus, same run:

                  tree-sitter    M1 subject
    exact              432          406
    wider               12          110
    WRONG                2           87
    no-answer          174           17

Repointing today takes WRONG from 2 to 87. WRONG is the gate bucket -- a file
the parser gets positively WRONG rather than declines -- so this would be the
largest regression in the project's history, not the win this issue assumed.

## Two things this issue got wrong

**1. The adapter already exists.** `internal/parse/cmd/subject/main.go` builds
SubjectFacts from an M1 tree, and `parse.Sites(root, src)` already returns
`[]parseoracle.SubjectCallSite` with all five markers. TestGoParserSubjectMeasures
has been measuring it over T2 since the M1 gate landed. The work I scoped as
"write a ParseSubject answering five markers" was done before this issue was
filed.

**2. Prototypes were never the blocker.** This issue was filed blocked on
01a0ad3f for cross-file resolution. That chain is now complete (f066f2ab,
7034e553, 9cf2ee6d) and the numbers above are measured WITH it. Prototypes are
not why the repoint fails.

## What actually blocks it

The hedge. An Unknown statement reports five Unresolved sites spanning it --
one per marker -- which `internal/parse/sites.go` documents as the whole
reason Unknown exists: "emit Unknown and stay quiet is not a refusal. It is
the claim that nothing is there."

    $ go run ./internal/parse/cmd/subject t/op/bop.t
    {"ok":true,...,"call_sites":[
      {"kind":"reference","line":22,...,"unresolved":true},
      {"kind":"hash","line":22,...,"unresolved":true},
      {"kind":"match","line":22,...,"unresolved":true},
      {"kind":"readline","line":22,...,"unresolved":true},
      {"kind":"anonhash","line":22,...,"unresolved":true}, ...

Correct, and expensive: every hedged site where perl found nothing scores
against the file. The tree-sitter path avoids this by declining the whole
FILE instead (174 no-answer against 17), so those files are never scored at
all.

So the two subjects are not answering the same question. One declines by the
file, the other by the statement, and the harness's buckets price those very
differently. That is the real work here, and it is a design question about
what a hedge costs -- not a repoint.

## Recommendation

Split. This issue should become:

1. **Price the hedge.** Decide whether five Unresolved sites per Unknown is
   the right report, or whether a statement-level decline should map to
   no-answer for that statement rather than WRONG for the file. That is a
   change to CompareFacts or to Sites, and it needs its own measurement.
2. **Then repoint**, when the M1 subject's WRONG count is at or below
   tree-sitter's 2.

01a0a8cd (remove the WRONG pin) stays blocked on the repoint either way, and
its own analysis is unaffected: both pinned files are tree-sitter defects.

Paused rather than closed: the issue is real and its goal is right, but the
route it describes is not available and the next step is a decision about
hedge pricing that belongs to perigrin.


## 2026-09-19: the earlier diagnosis was partly wrong

I paused this issue saying the blocker was hedge PRICING -- that the M1
subject reports five Unresolved sites per Unknown while tree-sitter declines
whole files, so the two answer different questions. That is still true as far
as it goes, and it is NOT the whole story.

Running the T2 gate (which needs PARSEORACLE_SHIM and had not been run in
this session) surfaced a plain missing-site bug:

    cmd/subval.t  rv2hv  perl read a hash at 2 site(s) the subject did not,
                  and at lines 182, 184 the subject committed to the
                  statement with no such site and no hedge

Both lines are `%$href`. `isHashTerm` tests a LEAF's text for a leading `%`,
so it catches `%hash` and misses every DEREFERENCE -- `%$href` is a Unary
`%` over a Term `$href`, and the Term case sees only `$href`. perl emits
rv2hv for both.

Fixed in 461ea788:

    T2 subject   2 WRONG -> 1, and the 1 is the pinned comp/proto.t M4 TODO
    T2 exact     41 -> 42 of 56 measured (73.2% -> 75.0%)

So some part of the 87 WRONG this issue measured was never about hedging at
all -- it was a site kind the subject did not know how to report. How much is
being measured now over the full corpus.

## What this means for the repoint

The 620-file numbers in this issue were taken before that fix AND before
today's parser work (01a0b983, 01a0b99b, which between them removed 1,025
false unresolved calls from T1). They are stale and should not be used to
decide anything.

The decision this issue was paused on -- how a hedge should be priced --
remains open, but it should be made against a re-measurement, and against the
per-statement findings 01a0ba79 adds. A ranked table of where WRONG comes
from turns "87 versus 2" into a list of things to fix.
