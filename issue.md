---
title: Ten refusals say unfiled, so ten recorded gaps point at no work
state: pending
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T02:09:22.469057177Z
updated: 2026-09-26T02:09:22.469057177Z
---

Ten corpus cases refuse and record `refuses: unfiled`. `refuses:` exists
to name an issue so a refusal points at live work; `unfiled` is the corpus
admitting it does not. This issue makes every one of the ten point
somewhere, by CITING where an issue exists and FILING where none does.

## Measured, at b0f1a9e9

All ten still refuse -- none is stale. Grouped by refusal code, which is
what says whether they are one gap or several:

    trailing_tokens             heredocs.md, 3 cases
      The interpolating heredoc
      The non-interpolating heredoc
      The indentation-stripping heredoc

    unimplemented_statement     formats.md, 2 cases
      A format declaration and the `write` that uses it
      A picture line that is not Perl

    not_a_term                  pod-and-data.md, 2 cases
      A data section read back through `DATA`
      A data section holding bytes that are not a program

    not_a_term                  comparison.md, 1 case
      The five logical operators          (also fails a `xor` token fact)

    mixed                       the two adjacency bodies
      adjacency-04_operators.md  5 Unknown: not_a_term, trailing_tokens
      adjacency-13_opaque.md     3 Unknown: unimplemented_statement,
                                 trailing_tokens, not_a_term

## What already has an issue

- **The three heredocs**: `01a0c13f-aaf8`, "A heredoc in statement
  position parses, body and all". Pending in m1-parse-the-core. Cite it.
- **The `xor` half of the logical-operators case**: the word-shaped
  operator issue in this milestone. Its `not_a_term` half is separate and
  needs measuring -- `not` and `xor` are both in that source, so which one
  the parser trips on is not yet established.

## What has no issue and needs one

- **`format` / `write`**: two cases, `unimplemented_statement`. Nothing in
  any milestone mentions `enterwrite` or a format declaration as parser
  work. Tier 13 introduces `enterwrite`, so this is the tier's own subject
  refusing.
- **`__DATA__` read back through `DATA`**: two cases, `not_a_term`. The
  lexer produces `DataSection(...)` correctly -- visible in the
  adjacency-13 token dump -- so the gap is the PARSER not accepting
  `<DATA>` as a term, which is narrower than "data sections".

## The adjacency bodies are not their own gap

Each mixes its whole tier into one body, so its Unknown count is the sum
of the gaps above reaching one file. They get no citation of their own;
they are cited against whichever issues the per-construct cases name, and
they clear when those clear. Recording that here so a later reader does
not file an "adjacency-04 refuses" issue.

## Do not

Do not close a refusal by weakening a case. A `refuses:` record is the
corpus working -- it SKIPS, keeping the suite pristine while naming a gap.
The deliverable is a live citation, not a passing case.

## The liveness check is someone else's issue

`01a0d0c7` in `v0.1` already owns "every `refuses:` citation names an OPEN
issue", having found five pointing at closed ones. Do not build that check
here -- two issues writing one test is how a check ends up half-owned.

This issue's job is narrower: no citation says `unfiled`. A citation that
is live TODAY and closed later is 01a0d0c7's problem, and the two together
cover the field.

## Context

- paths: `conformance/mdtest/heredocs.md`, `conformance/mdtest/formats.md`, `conformance/mdtest/pod-and-data.md`, `conformance/mdtest/comparison.md`, `conformance/mdtest/adjacency-04_operators.md`, `conformance/mdtest/adjacency-13_opaque.md`
- commands: `go test ./internal/conformance/ -count=1`
- entrypoints: `internal/conformance/run.go:151` (the staleMarker check that keeps a citation honest)

## Acceptance criteria

- [ ] No case in the corpus records an unfiled refusal (`! grep -rq 'refuses: unfiled' conformance/mdtest/`)
- [ ] The heredoc cases cite the existing heredoc issue (`grep -q '01a0c13f-aaf8' conformance/mdtest/heredocs.md`)
- [ ] A format/write issue exists and is cited (`grep -qE 'refuses: 01a0[0-9a-f]{4}' conformance/mdtest/formats.md`)
- [ ] The data-section cases cite an issue (`grep -qE 'refuses: 01a0[0-9a-f]{4}' conformance/mdtest/pod-and-data.md`)
- [ ] Every citation still refuses, so no marker went stale in the edit (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 | grep -q '^ok'`)
- [ ] The suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
