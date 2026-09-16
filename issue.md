---
title: "Canonical re-emission: compare the tree written back out against its source"
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
created: 2026-09-16T06:13:00.800228773Z
updated: 2026-09-16T06:13:36.521296521Z
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

- [ ] `canon` emits Perl from a tree, with the tree's own parenthesisation (`go test ./internal/parse/ -run '^TestCanonEmitsTreeParens$' -count=1 -v | grep -q '^--- PASS: TestCanonEmitsTreeParens'`)
- [ ] A deliberately mis-grouped tree re-emits differently from its source (`go test ./internal/parse/ -run '^TestCanonCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonCatchesMisgrouping'`)
- [ ] Significant tokens of `canon(parse(S))` equal those of `S`, over the whole corpus (`go test ./internal/parse/ -run '^TestCanonTokenIdentity$' -count=1 -v | grep -q '^--- PASS: TestCanonTokenIdentity'`)
- [ ] The per-file result is ratcheted, failing in either direction (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)

Plan: M3 metrics, M6 metrics. Spec: §7.5 (c').
