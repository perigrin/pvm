---
title: "Four wrong trees that round-trip: anon sub, statement-start brace, glob deref, signatures"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
created: 2026-09-17T14:24:26.929009826Z
updated: 2026-09-18T15:08:01.209065906Z
sessions:
- start_sha: e0d124d1932fa7a5443eb3f763146c1a6171decc
  end_sha: ""
  commits: 0
  started_at: 2026-09-18T14:54:54.022377723Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-18T14:54:54.022377723Z
---

Four constructs produce a tree that is not a parse of its source, and every one of them round-trips. That is the failure `Unknown` exists to prevent (`internal/parse/parse.go:14-26`): a refusal is never a wrong answer, and these are wrong answers.

Measured at 3c997a6b:

    sub { 1 };                  [statement declaration] [statement anon_hash term]
    {a=>1};                     [statement block [statement binary call term]] [unknown]
    *{"foo"};                   [statement term [anon_hash term]]
    use v5.36;
    sub f ($x, $y) { 1 }        [statement declaration term]
                                [statement index [list term term] term]

No test we have can catch these. Round-trip passes on all four, because
round-trip proves no byte was lost, not that the parse was right (chapter 7
§7.2(c)). Canonical re-emission (01a0a8d8) would catch the grouping errors
and is itself blocked behind the M1 gate.

## Each one

**`sub { 1 };` in statement position.** An anonymous sub read as a
declaration plus an anonymous hash — two statements where there is one
expression. `statement()` reaches `parseDeclaration` before `parseTerm`'s
anon-sub check (`internal/parse/term.go:126-132` has the check; the statement
path never gets there), and `finishBodyOrSemicolon` then sees a `{` the lexer
classified as a term. In expression position it is correct:
`my $f = sub { 1 }` gives `[declaration "my" [term "$f"] [declaration "sub"
[block ...]]]`.

**`{a=>1};` at statement start.** Parsed as a block containing a `=>`
statement, then an `Unknown` for the `;`. perl disagrees — `B::Deparse` gives
`+{'a', 1};` — because `toke.c:6698-6842`'s heuristic fires at statement
start too. This is `intuit_curly`'s subject and overlaps 01a0ac52; take it
there or here, but not in both.

**`*{"foo"}`.** A glob dereference read as `term "*"` over an `anon_hash`.
The braces group an expression; they do not construct a hash. Same shape as
the `${EXPR}` problem in 01a0ad52 and probably the same fix.

**Signatures under `use v5.36`.** The worst of the four, because turning the
feature ON makes the parse worse. Without the pragma the signature is read as
a prototype and the body as a block — which is what perl does when signatures
are off. With it, `internal/lexer/proto.go` returns false when `l.signatures`
is set and nothing downstream picks the signature up, so the body is read as
a hash slice OF the signature and the sub body is lost. `use feature
'signatures'` does the same.

## Why they are one issue

Three of the four are the same class of bug: a `{` classified by position
when the classification needs a lookahead. The fourth (signatures) is a
handoff nobody completes. They are grouped because fixing them separately
means four passes over the same statement-dispatch path, and because the
measured lesson of this session is that changes to brace classification
interact — two locally correct fixes each made T1 worse.

Split if the signature work turns out to be independent, which it may be.

## Scope

Not in scope: executing any of it, prototype resolution (M4), or
`MethodCall.Indirect` — `new Foo(1)` is also `Unknown` but indirect-object
syntax is a scope decision recorded as open in §4.14.6, not a defect.

## Acceptance

- [x] `sub { 1 };` is one statement holding one anonymous sub (`go test ./internal/parse/ -run '^TestAnonSubInStatementPosition$' -count=1 -v | grep -q '^--- PASS: TestAnonSubInStatementPosition'`)
- [x] `{a=>1};` parses as an anonymous hash, matching `B::Deparse` (`go test ./internal/parse/ -run '^TestStatementStartBrace$' -count=1 -v | grep -q '^--- PASS: TestStatementStartBrace'`)
- [x] A bare block `{ my $x = 1; }` is still a block (`go test ./internal/parse/ -run '^TestBareBlockStillParses$' -count=1 -v | grep -q '^--- PASS: TestBareBlockStillParses'`)
- [x] `*{"foo"}` is a glob dereference, not a hash constructor (`go test ./internal/parse/ -run '^TestGlobDerefIsNotAConstructor$' -count=1 -v | grep -q '^--- PASS: TestGlobDerefIsNotAConstructor'`)
- [x] `sub f ($x, $y) { 1 }` keeps its body under `use v5.36` and under `use feature 'signatures'` (`go test ./internal/parse/ -run '^TestSignaturesKeepTheirBody$' -count=1 -v | grep -q '^--- PASS: TestSignaturesKeepTheirBody'`)
- [x] The signature's parameters are reachable, not swallowed into one span (`go test ./internal/parse/ -run '^TestSignatureParametersAreNodes$' -count=1 -v | grep -q '^--- PASS: TestSignatureParametersAreNodes'`)
- [x] Without any signature pragma, `sub f ($x) { 1 }` still reads `($x)` as a prototype, as perl does (`go test ./internal/parse/ -run '^TestPrototypeWithoutSignatures$' -count=1 -v | grep -q '^--- PASS: TestPrototypeWithoutSignatures'`)
- [x] Each of the four produces a tree whose shape a test asserts, not merely a round trip (`go test ./internal/parse/ -run '^TestWrongTreesAreGone$' -count=1 -v | grep -q '^--- PASS: TestWrongTreesAreGone'`)
- [x] T1 clean does not fall below 408 and Unknown does not rise above 4,944 (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [x] The whole suite is green (`go test ./internal/... 2>&1 | grep -qv FAIL`)

The T1 AC uses `TestParseRatchet`, committed at 4bc971ec. For this issue the ratchet matters more than for the others: three
of these four touch brace classification, and this session measured two
locally correct brace fixes that each made T1 worse.

Spec: §4.14.2 (`AnonSub`, `AnonList`/`AnonHash`, `Deref{Sigil,Expr}`,
signatures). Plan: M1.


## Done, 2026-09-18 (5b415101)

    T1-easy   50.6% -> 50.8%
    T1        7 files better, none worse
    T2        16 of 56 clean; class/method.t 12 -> 9
    oracle    40 exact, 1 WRONG (the M4 pin), unmoved

All four fixed. The statement-start brace rule turned out NARROWER than this
issue assumed -- measured on perl 5.42.0:

    {a=>1};      +{'a', 1}    a hash: BAREWORD then `=>`
    {"a", 1};    +{'a', 1}    a hash: STRING then `,`
    {a, 1};      a block      a bareword and a comma is NOT enough
    {$k=>1};     a block      a variable is not a key here
    {};          a block      empty is a block, not an empty hash

This issue's own AC said "`{};` parses as an anonymous hash". perl says
block, so that expectation was wrong and the test records the measurement
instead.

## An existing test asserted the opposite and was wrong

`TestBlockVersusHashref` held that `{ a => 1 }` must be a Block, "because
the lexer read the brace in XState" -- a description of the implementation
rather than of Perl, which is the kind of test that freezes a defect. perl
deparses it `+{'a', 1}`. Its sibling `if ($c) { a => 1 }` IS a block and
stays, because a block is required there.

## One T2 file rose and it is not a regression

class/gh22169.t went 3 -> 5 because reaching further into the file exposed
`ADJUST { ... }` -- a class phaser in no table -- twice more. Verified by
stashing the change and counting: 3 Unknowns at HEAD, 5 after, every new one
an ADJUST block. Nothing in the chain owns ADJUST.

## Two cases found and deliberately left

Both named in the tests rather than quietly skipped:

  - `LOOP: { last LOOP; }` -- a labelled bare block. Valid Perl, declined
    before and after this change. A label before a BLOCK rather than before
    a loop or statement.
  - `{};` -- the block parses, the trailing `;` becomes an Unknown. An
    empty-statement defect, not a brace-classification one.

Neither is this issue's subject and neither is tracked.
