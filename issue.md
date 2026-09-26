---
title: A malformed format declaration loses canon's fixpoint, and one perl rejects parses clean
state: pending
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T08:18:30.111073325Z
updated: 2026-09-26T08:18:30.111073325Z
---

A malformed `format` declaration loses canon's fixpoint, and one spelling
perl rejects parses clean. Both reachable only from invalid perl, and both
found by the TIER_2 gate on `01a0dc82`.

## Two shapes, both perl syntax errors

Measured on 5.42.0 at `/home/perigrin/.local/bin/perl`:

    format NAME =
    no terminator here
      -> Format not terminated at line 4, at end of line
         syntax error at line 4, at EOF

    format NAME
    write;
      -> syntax error at line 2, near "format NAME\nwrite"

## Defect 1: canon stops being a fixpoint

An unterminated body used to be one `unimplemented_statement` Unknown with
`fixpoint=true`. After `8af023b6` it is `not_a_term` with
**`fixpoint=false`**:

    canon  = "format NAME = no terminator here\n"
    canon2 = "format NAME = no terminator here(); "

Root cause, and it is a sibling-function gap: `internal/lexer/pod.go:136`
emits `UnknownRest` rather than `FormatBody` when the body has no
terminating `.` line. `parseFormatDecl` consumes only `lexer.FormatBody`, so
the `UnknownRest` falls through to the expression parser and canon re-spaces
it.

Fixpoint is the property `TestCanonRatchet` and the fidelity harness rest
on: canon of canon must equal canon. Losing it on any input is worth
recording even when the input is invalid, because nothing in the suite
currently notices -- `FuzzParser`'s invariants
(`internal/parse/fuzz_test.go`) do not check fixpoint, so a fuzzer reaching
this shape would find no assertion to break.

## Defect 2: `format NAME` with no `=` parses clean

`format NAME\nwrite;` gives ZERO Unknowns; canon emits
`format NAME write();`. Both the `=` and the body are `if ... ok` in
`parseFormatDecl` with no else.

This is systemic rather than new -- the parser is deliberately permissive
and the milestone has no "reject invalid perl" gate -- but it is the same
function, so it belongs here.

## Why neither is urgent

Neither hangs nor panics (verified, 60s timeouts, both returned instantly).
Both are invalid perl, so neither can appear in T1, T2 or the corpus, all
of which are perl-validated. Nothing regressed for valid input.

## Scope

1. `parseFormatDecl` consumes an `UnknownRest` following a `format NAME =`,
   so the declaration owns its malformed body rather than leaking it, OR
   `scanFormatBody` emits `FormatBody` for the unterminated case and lets
   the parser decide. Measure which is smaller before choosing.
2. Decide whether `format NAME` with no `=` should refuse. If the answer is
   "the parser is permissive by design", write that down here and close --
   a decision recorded is worth more than a behaviour left ambiguous.
3. Consider adding canon fixpoint to `FuzzParser`'s invariants, which is
   what would have caught defect 1 without a human reading canon output.
   That may belong in its own issue; say which.

## Context

- paths: `internal/lexer/pod.go` (`scanFormatBody`, ~line 136), `internal/parse/decl.go` (`parseFormatDecl`), `internal/parse/fuzz_test.go`
- commands: `go test ./internal/parse/ -count=1`, `go test ./internal/lexer/ -count=1`
- entrypoints: `scanFormatBody`'s `UnknownRest` path and `parseFormatDecl`'s `lexer.FormatBody` check

## Acceptance criteria

- [ ] An unterminated format body keeps canon's fixpoint (`go test ./internal/parse/ -run '^TestUnterminatedFormatBodyIsAFixpoint$' -count=1 -v | grep -q '^--- PASS: TestUnterminatedFormatBodyIsAFixpoint'`)
- [ ] The decision on `format NAME` with no `=` is written down (`grep -qiE 'no .=. at all|permissive by design' internal/parse/decl.go`)
- [ ] REGRESSION GUARD, passes today: the canon ratchets hold (`go test ./internal/parse/ -run '^TestCanon' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse suite holds (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the conformance suite holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
