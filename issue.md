---
title: Fourteen of seventeen refusals point at nothing actionable
state: pending
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T02:51:11.245880623Z
updated: 2026-09-26T02:51:11.245880623Z
---

Fourteen of the corpus's seventeen refusals point at nothing a reader can
act on. `refuses:` exists to name an issue so a refusal names live work;
ten say `unfiled` and four cite issues that are `done`.

**Rewritten by chain-review, 2026-09-26.** This issue was filed as "ten
refusals say unfiled". Naming all seventeen rather than counting them by
tier found four more, and found that the corpus's parenless-call cases are
among them -- which cancelled a separate issue in this milestone that
proposed to write coverage the corpus already had.

## Measured, at b9f77d9c

All seventeen still refuse; none is stale in the `staleMarker` sense.

    ten   `refuses: unfiled`
            heredocs.md                  3   trailing_tokens
            formats.md                   2   unimplemented_statement
            pod-and-data.md              2   not_a_term
            comparison.md                1   not_a_term (the xor case)
            adjacency-04_operators.md    1   5 Unknown, mixed
            adjacency-13_opaque.md       1   3 Unknown, mixed

    four  cite a CLOSED issue
            argument-extent.md           2   01a0c432-fbd5   done
            adjacency-07_subroutines.md  1   01a0c432-fbd5   done
            adjacency-11_oo.md           1   01a0c35f-a487   done
            named-operators.md           1   01a0c730-b241   done

    three cite a LIVE issue, and are correct
            compound-assignment.md           01a0ce57   pending
            ungated.md                       01a0d087   pending

(Fifteen rows for fourteen citations: `01a0c432-fbd5` is cited by three
cases across two topics.)

## The four closed citations matter more than the ten unfiled

`unfiled` at least reads as an admission. A citation naming a `done`
issue reads as tracked work, so a reader follows it, finds a finished
issue, and concludes the corpus says nothing about the construct in front
of them. That happened during this milestone's own construction:
`argument-extent.md` holds five cases on parenless argument extent and a
reader concluded the corpus had no such case.

## Scope

For each of the fourteen, establish what work the refusal is waiting on
and record it. Three routes, and which applies is a measurement:

1. **A live issue already covers it.** Cite that id. Known:
   `01a0c13f-aaf8` for the three heredocs.
2. **The gap is real and untracked.** File an issue and cite it. Known to
   need one: `format`/`write` (2 cases, `unimplemented_statement`, and
   nothing in any milestone mentions `enterwrite`); `__DATA__` read
   through `DATA` (2 cases, `not_a_term`, where the LEXER is correct --
   `DataSection(...)` is produced -- so the gap is the parser not taking
   `<DATA>` as a term).
3. **The cited issue closed but the work did not land.** File the residue
   and cite it; do NOT re-cite the closed issue. This is the likely shape
   for the three `01a0c432-fbd5` cases, whose issue closed on writing the
   corpus files rather than on making them parse.

The `xor` case (comparison.md) needs measuring before it can be routed:
`not` and `xor` are both in that source and which one trips the parser is
not established. It also fails a token fact, which the glossary issue
owns -- the PARSER half is this issue's.

## The two adjacency bodies get no citation of their own

Each mixes its whole tier into one body, so its Unknown count is the sum
of the gaps above reaching one file. They are cited against whichever
issues the per-construct cases name and clear when those clear. Recorded
so a later reader does not file an "adjacency-04 refuses" issue.

## Do not

Do not close a refusal by weakening a case. A `refuses:` record is the
corpus working -- it SKIPS, keeping the suite pristine while naming a
gap. The deliverable is a live citation, not a passing case.

## Relationship to 01a0d0c7

`01a0d0c7` in `v0.1` found the closed citations and filed the liveness
check. Its count is right on its own terms -- three closed issue ids
across five cases, and its "five" is cases -- but it does not look at
`unfiled` at all, so it sees five of the fourteen.

This issue covers all fourteen. `01a0d0c7` keeps the CHECK
(`TestRefusalCitationsAreLive`), which stays its deliverable so one
guarantee is not half-owned; this issue repairs the data the check will
then hold in place. Order matters: the repair before the check, or the
check lands red.

Its own diagnosis is worth carrying forward -- three of the five closed
because THE MILESTONE they belonged to finished, not because the parser
learned anything. That is why a citation needs re-measuring rather than
re-pointing: the issue closing says nothing about whether the gap did.

## Context

- paths: `conformance/mdtest/heredocs.md`, `formats.md`, `pod-and-data.md`, `comparison.md`, `argument-extent.md`, `named-operators.md`, `adjacency-04_operators.md`, `adjacency-07_subroutines.md`, `adjacency-11_oo.md`, `adjacency-13_opaque.md`
- commands: `go test ./internal/conformance/ -count=1`
- entrypoints: `internal/conformance/run.go:150` (the `staleMarker` case, which keeps a citation honest once it is written)

## Acceptance criteria

- [ ] No case records an unfiled refusal (`! grep -rq 'refuses: unfiled' conformance/mdtest/`)
- [ ] No case cites any of the three closed issues (`! grep -rqE 'refuses: (01a0c432|01a0c35f-a487|01a0c730)' conformance/mdtest/`)
- [ ] The heredocs cite the live heredoc issue (`grep -q '01a0c13f-aaf8' conformance/mdtest/heredocs.md`)
- [ ] The format cases cite an issue (`grep -qE 'refuses: 01a0[0-9a-f]{4}' conformance/mdtest/formats.md`)
- [ ] The data-section cases cite an issue (`grep -qE 'refuses: 01a0[0-9a-f]{4}' conformance/mdtest/pod-and-data.md`)
- [ ] The parenless-call cases cite something other than the closed slice issue (`grep -qE 'refuses: 01a0[0-9a-f]{4}' conformance/mdtest/argument-extent.md`)
- [ ] Every refusal in the corpus still refuses, so no marker went stale in the edit (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
- [ ] The suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
