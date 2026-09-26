---
title: A __DATA__ section has no statement form, and it is the section not the handle
state: in-progress
urgency: normal
milestone: corpus-covers-t2
blocked_by:
- 01a0dc82-28e1-72d3-82cd-88ed8479f327
created: 2026-09-26T07:00:39.376307576Z
updated: 2026-09-26T08:50:45.925189552Z
sessions:
- start_sha: d1537f9631443c467a73f3e56cf382da4eeab2ae
  end_sha: d1537f9631443c467a73f3e56cf382da4eeab2ae
  commits: 0
  started_at: 2026-09-26T07:02:01.342336494Z
  ended_at: 2026-09-26T07:02:45.841983723Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T07:02:01.342336494Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T07:02:45.841983723Z
---

A `__DATA__` or `__END__` section refuses with `not_a_term`. The lexer
reads it correctly; the parser has no statement form for the token.

Two corpus cases, tier 13's own subject, and nothing tracked them.

## The obvious diagnosis is wrong, and that matters

Both failing cases read `<DATA>`, so the natural reading -- and the one
the earlier record carried -- is "the parser will not take `<DATA>` as a
term". Measured, that is FALSE:

    <DATA> with section      Unknown=1
    <DATA> no section        Unknown=0     <- the readline alone is fine
    <FOO> with section       Unknown=1
    <STDIN> with section     Unknown=1
    print 1; __DATA__ x      Unknown=1     <- NO readline at all
    <DATA> after __END__     Unknown=1

The discriminator is the SECTION, not the handle. `<DATA>`, `<FOO>` and
`<STDIN>` all parse without one and all refuse with one, and a program
with a section and no readline refuses too.

`<DATA>` appeared in every failing case because a program that reads a data
section usually has one, which is exactly the coincidence that makes a
count-then-attribute reading unsafe.

## What the lexer already does

    my @lines = <DATA>;
    print @lines;
    __DATA__
    one
    two

lexes to, among others:

    Readline("<DATA>")
    DataSection("__DATA__\none\ntwo\n")

Both correct. `DataSection` is a lexer Kind, the glossary has a
`data section` category, and tier 13 exists for constructs the optree
cannot see. So the work is a statement form, not lexing.

## A second, separate failure on the same case

The case also fails a token fact:

    one data section whose text is "__DATA__\none\ntwo\n\n"

Note the TRAILING BLANK LINE the fact wants; the lexer emits
`"__DATA__\none\ntwo\n"` without it. Whether the fact or the lexer is
right is a measurement nobody has made -- perl's `DATA` filehandle yields
`one\n` and `two\n`, and whether the section token should carry a final
newline beyond that is a boundary question the glossary's `data section`
entry does not settle.

Decide that BEFORE the statement form, because a statement form that makes
the parse clean leaves the case still failing on the fact, and a reader
would then think the fix did not work.

## Sequencing, and a state the tool would not clear

BLOCKED BY `01a0dc82` (format/write), because both need a statement form
and both would edit `internal/parse/stmt.go`. Running them concurrently
races on one file for no gain.

This issue was `--state start`ed and then paused when that was noticed.
`--state pause` recorded the transition but left the state
`in-progress` -- measured, both in the display and in
`issue show --format json`. So the state is wrong and the EDGE is what
actually holds the work back: nothing should pick this up until
`01a0dc82` is done, regardless of what the state says.

Recorded rather than papered over, because an `in-progress` issue with
zero commits is exactly the shape that reads as abandoned work later.

## Closed with --force, and this is why

The done transition was REFUSED: "cannot mark done: session has 0 commits".
Correct as the tool sees it, and wrong about the work.

The commits exist -- `47e5b1ff`, `99b0314e` on the worker's branch, merged
as `da0d3dfe` -- but they were made by a DISPATCHED WORKER under
`agent:x7k2-6` in its own worktree, and the merge was made by the
orchestrator. Neither the worker's session nor the orchestrator's holds
both, so the "commits in this session" test finds none.

That is the same actor-split the orchestration was designed around: each
worker gets its own identity so mutual exclusion works, which is exactly
what makes a per-session commit count the wrong test for whether an issue's
work landed.

Forced rather than worked around, with the evidence here instead of a
silent override:

  - six of six acceptance criteria exit 0, run verbatim
  - corpus ratchet 197 -> 199 clean, 15 -> 13 refusing
  - four other ratchets moved, every mover verified to be a file with a
    trailing `__END__`/`__DATA__`
  - `internal/parse`, `internal/lexer`, `internal/conformance` all green
  - the PAAD gate HAS now run: TIER_2 on ten files, verdict IN-SCOPE
    FINDINGS, and all four were fixed in `2963c47d` before this close

### What the gate found, since the fix itself was sound

It confirmed the model (trivia is right), the trailing-newline ruling
(verified against perl), all six criteria, and all five ratchet claims --
including the one mover whose filename does not advertise a trailing
marker, which it went looking for as a counterexample and found compliant.

Every finding was in the RECORD rather than the code:

  - "canon protects the section's bytes" was FALSE, and it was the stated
    reason the sibling `format` issue's canon hazard could not recur here.
    Canon is blind twice: `canon.go` drops Trivia unemitted and
    `canon_test.go`'s `significant()` skips `DataSection`. The real guard
    is `TestDataSectionIsTrivia`'s `last.End != len(src)`.
  - the leftover refusal was attributed to `write` and a bare block. Both
    parse clean; there is no bare block in that program at all; the cause
    is the heredoc BODY. That attribution was the stated reason a sentinel
    stayed `unfiled`, so it now cites `01a0c13f-aaf8`, which owns the gap.
  - the same passage said "Three Unknowns" where there is one.
  - `use_test.go:215` was the fourth copy of the trivia kind set and the
    only one not updated.

## Scope

1. Establish whether the section token should include a trailing newline.
   Measure against perl, and record the answer in the glossary's
   `data section` entry, which is silent on it.
2. A statement form for `__DATA__`/`__END__` so the section terminates the
   program rather than refusing. The body stays opaque.

## Context

- paths: `internal/parse/stmt.go`, `internal/lexer/scan.go` (`scanDataSection`, already present), `conformance/mdtest/pod-and-data.md`, `conformance/GLOSSARY.md`
- commands: `go test ./internal/parse/ -count=1`, `go test ./internal/conformance/ -count=1`
- entrypoints: whatever emits `not_a_term` when a `DataSection` token reaches statement position

## Acceptance criteria

- [ ] The data-section case parses (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -qE 'PASS: TestCorpus/pod-and-data.md/A_data_section_read_back_through_.DATA.'`)
- [ ] The not-a-program case parses (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q 'PASS: TestCorpus/pod-and-data.md/A_data_section_holding_bytes_that_are_not_a_program'`)
- [ ] Neither case leaves the sentinel behind (`! grep -q 'refuses: unfiled' conformance/mdtest/pod-and-data.md`)
- [ ] The glossary settles the trailing-newline boundary (`grep -qiE 'trailing newline|final newline' conformance/GLOSSARY.md`)
- [ ] REGRESSION GUARD, passes today: the parse suite holds (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the conformance suite holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
