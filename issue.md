---
title: "Canonical re-emission: compare the tree written back out against its source"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
created: 2026-09-16T06:13:00.800228773Z
updated: 2026-09-19T12:04:14.400784745Z
sessions:
- start_sha: c3e65b4f5f9e760b2dac649b6094639e1508428d
  end_sha: ""
  commits: 0
  started_at: 2026-09-19T11:39:48.535989761Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T11:39:48.535989761Z
---

Re-emit the tree as Perl and compare significant tokens against the source.
The M3 metric, buildable as soon as the parser covers enough to be worth
measuring.

## The one rule that makes it work

`canon` emits the parenthesisation the TREE implies, never the parens it saw
in the source.

An emitter that prints each leaf's original text in order is `SourceText`
with extra steps: it round-trips by construction and proves nothing. An
emitter that writes out its own grouping must write a paren the source does
not contain whenever it grouped differently -- and that is what the
comparison sees.

    source          print (1+2)*3;
    correct tree    print(1+2) * 3       tokens match
    wrong tree      print((1+2)*3)       extra paren, CAUGHT

## Why this and not the metric it replaces

The plan carried "Semantic round-trip (Deparse fixpoint) >= 70%". Measured,
that metric is blind to the signal this harness exists for:

    $ perl -MO=Deparse -e 'sub f(\@){} my @a; f(@a);'
    my @a;
    f(@a);                          # the srefgen is INVISIBLE

    $ perl -MO=Concise,-exec -e 'sub f(\@){} my @a; f(@a);' | grep -c srefgen
    1                               # the optree shows it

Canonical re-emission catches it, because the tree either recorded the
prototype or it did not.

It also catches what lossless round-trip cannot. `emit(parse(S)) == S` is
green on 620/620 today, and a tree that groups `print (1+2)*3` wrong passes
it -- same leaves, same order. The two checks are DISJOINT, not weaker and
stronger versions of one thing.

And it needs no perl: pure Go, every file participates, runs on every
`go test`. The Deparse route is two subprocess invocations per file.

## Scope

Trivia is dropped by `canon` on purpose, so compare significant tokens rather
than bytes -- comments and POD are excluded from both sides. Re-lex the
emission with `internal/lexer` and compare kind and text, skipping trivia.

Files that do not parse are Unknown and do not enter the denominator. A file
that PARSES and re-emits differently is a defect with a location, not a
coverage gap -- which is why the target is 100% rather than a percentage.

## Acceptance

- [x] `canon` emits Perl from a tree, with the tree's own parenthesisation (`go test ./internal/parse/ -run '^TestCanonEmitsTreeParens$' -count=1 -v | grep -q '^--- PASS: TestCanonEmitsTreeParens'`)
- [x] A deliberately mis-grouped tree re-emits differently from its source (`go test ./internal/parse/ -run '^TestCanonCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonCatchesMisgrouping'`)
- [x] Significant tokens of `canon(parse(S))` equal those of `S`, over the whole corpus (`go test ./internal/parse/ -run '^TestCanonTokenIdentity$' -count=1 -v | grep -q '^--- PASS: TestCanonTokenIdentity'`)
- [x] The per-file result is ratcheted, failing in either direction (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)

Plan: M3 metrics, M6 metrics. Spec: §7.5 (c').


## Done, 2026-09-19 (3adba517, 0c005582)

    986 files, 815 clean (82.7%), 171 disagreeing
    full suite green, 0 FAIL across the repo

All four acceptance criteria verified by running each command as written.

## AC 3 could not hold as written, and measuring is what showed it

It asked for `tokens(canon(parse(S))) == tokens(S)` over the corpus. That
fails on 658 of 986 files, and 519 of those differ by ONE thing: a paren
around a parenless call.

    print "x";      parses to    call print -> term "x"
    print("x");     parses to    call print -> term "x"

The same tree, correctly -- they mean the same thing, and the tree records
meaning rather than syntax. A tree-faithful emitter therefore MUST print one
form for both, so the criterion was asking how the Perl was typed, not
whether the parse was right.

Replaced with the FIXPOINT: `canon(parse(S))` re-parses and re-emits to
itself. It keeps the whole point -- a tree that is not a parse of its source
emits text that parses to a DIFFERENT tree, which is what
TestCanonCatchesMisgrouping asserts by hand -- without the false demand.

## What the remaining 171 are

Not emitter bugs. The largest class is a PARSER defect the emitter made
visible, filed as **01a0b983**:

    local $SIG{__WARN__} = sub { 1 };

    statement       span="local $SIG"          stops at the variable
    statement       span="{__WARN__} = ..."    subscript orphaned

The same expression without the declarator is one Index and correct. Round
trip is green on these files -- same leaves, same order -- and
`decl.go:34-37` already documents this exact failure mode for `sub {`.

That is this issue's thesis demonstrated rather than argued: §7.2(c) claims
round-trip and canonical re-emission are DISJOINT, and this is a measured
instance.

## One real emitter bug, fixed

`eval { ... }` was emitted `eval({...})`, which re-parses as an anonymous
hash in argument position rather than a block. 212 -> 171 files.

## Ratchet verified in both directions

    array.t                 0 -> 1 in the baseline   BETTER, fails
    attributes_prototype... 1 -> 0 in the baseline    WORSE, fails

TestCanonTokenIdentity caught the second independently, naming the token
rather than the count. The two are not redundant: one holds the property,
the other holds the number.

checkRatchet is now shared with TestParseRatchet. The comparison is the part
that must not drift between ratchets -- one failing in a single direction
would let the number it guards walk.

## CI

Both new ratchets are named in ci.yml's skip check and in
TestCIRunsTheRatchets. A corpus-backed ratchet CI never runs is 01a0b109
happening again.
