---
title: "Four trees that disagree with perl, found building the arity refusal: pipe my(...) or die, indirect new/is, f { } for (&@), canon (&$)"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T17:47:42.315506839Z
updated: 2026-10-06T17:47:42.315506839Z
---

## Context

- paths: `internal/parse` (declaration/list precedence, indirect method calls, `(&@)` block calls), `internal/parse/canon*` (canon emission)
- commands: `make test`; ratchets; perl 5.42 `/home/perigrin/.local/bin/perl -MO=Deparse,-p`

Found while building 01a10177-547c-73a9-a302-f25f579a4cb3 (parse-time arity refusal); each is a tree that disagrees with perl, none caused by that issue:

1. `pipe my ($r, $w) or die` parses with the `or` inside the `my` declaration; perl: `(pipe(my($r, $w)) or die)`.
2. `new IO::File` under a declared `new`, and `is Time::Moment->...` with Test::More loaded, give trees that disagree with perl's Deparse (perl reads both as indirect method calls here).
3. `sub f (&@) {} f { 1 };` is refused as `not_a_term`; perl compiles it (a `&@` sub called with a block and an empty list).
4. Canon emits `f({$_;} g())` for a `(&$)` block call, which `perl -c` rejects; the canon ratchet's closure.t moved 1 -> 0 only because the re-parse now refuses it as perl does -- the emission itself is still wrong.

## Steps

- [ ] For each, measure with Deparse,-p; write a failing parse (or canon) test; fix; ratchets must not regress (report movement, checked against perl); `make test`; commit

## Acceptance criteria

- [ ] the four shapes match perl (`go test ./internal/parse/ -run '^TestArityFoundShapesMatchPerl$' -count=1 -v | grep -q '^--- PASS: TestArityFoundShapesMatchPerl'`)
