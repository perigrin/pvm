---
title: A bare block in a class body is read as an anon hash, losing the declaration after it
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-26T12:24:40.484457586Z
updated: 2026-09-28T23:26:16.981058783Z
sessions:
- start_sha: c6fb51f36af20e7039f81cb35302be95b0e07005
  end_sha: d8029d82fa384bc5d781d771297fa742dfb4b01e
  commits: 1
  started_at: 2026-09-28T23:13:26.573168861Z
  ended_at: 2026-09-28T23:26:16.981058783Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-28T23:13:26.573168861Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-28T23:26:16.981058783Z
observed_paths:
- internal/lexer/expect.go
- internal/lexer/lexer.go
- internal/lexer/openblock_test.go
- internal/parse/nestedblock_test.go
- internal/parse/t2_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/use_test.go
---

A bare block inside a `class` body is read as an anonymous hash, so the
declaration after it is lost. Valid perl, and a divergence on valid input.
Found by the worker on `01a0dd71` and reported rather than fixed, correctly --
it is a separate cause with no ADJUST involved.

## Measured

    class C { { 1 } method m { 1 } }        1 Unknown

Perl accepts it. Verified on 5.42.0 at `/home/perigrin/.local/bin/perl`:

    perl -e 'use feature "class"; no warnings "experimental::class";
             class C { { 1 } method m { 2 } } print C->new->m;'
      -> 2

So the method is reached and callable; we lose it.

## Root cause, as reported

A bare `{ 1 }` is a `Block` at top level and an `anon_hash` inside a class
body. The class-body parse has no rule that reads a brace group in statement
position as a block, so it falls to the term parser, which takes it as a
hash and then wants a `;`.

That is the SAME MECHANISM `01a0dd71` fixed for `ADJUST`: there, `ADJUST`
was in neither phaser table so `ADJUST { 1 }` parsed as an index call -- a
bareword subscripted by an anon hash. Two spellings of one gap: a brace
group in a class body is not being read as a block.

The fix may therefore be smaller than it looks, or may be the same rule
generalised. Measure before choosing.

## Why it is not 01a0dd71's

No `ADJUST` appears. The worker measured the two independently and fixed
only what its issue named, which is why this is a separate record rather
than scope creep on a closed issue.

## Scope

1. Read a brace group in class-body statement position as a `Block`, the way
   top level does. Measure whether the top-level rule can be reached
   directly or whether the class-body parse needs its own.
2. Check the neighbours before choosing: a bare block after a `field`, after
   a `method`, two bare blocks in a row, and a bare block as the LAST member
   of the body. Each may or may not share the cause.
3. Check whether perl actually runs the block at class-construction time or
   treats it as a no-op -- it affects nothing about the parse, but the corpus
   case that pins this will need the right `output`.

## Context

- paths: whatever parses a `class` body (see `01a0dd71`'s fix for where the
  phaser tables live), `internal/parse/term.go` for the anon-hash reading
- commands: `go test ./internal/parse/ -count=1`,
  `go test ./internal/conformance/ -count=1`
- entrypoint: the class-body member parse, statement position

## Acceptance criteria

- [ ] A bare block before a method in a class body parses (`go test ./internal/parse/ -run '^TestBareBlockInClassBody$' -count=1 -v | grep -q '^--- PASS: TestBareBlockInClassBody'`)
- [ ] REGRESSION GUARD, passes today: ADJUST followed by a method still parses (`go test ./internal/parse/ -run '^TestAdjustThenMethod$' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse ratchets hold (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
