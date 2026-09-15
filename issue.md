---
title: "The M1 gate: parse ratchet, WRONG stays zero, measured as a subject"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-15T21:53:08.812269582Z
updated: 2026-09-15T21:53:34.150739998Z
---

M1's gate: the parse rate ratchets, WRONG does not move, and the subject
contract is what measures it.

## The corpus numbers need correcting first

The plan says M1's second corpus is "T1 graded, easy tier (~400 of PerlOnJava
`unit/`)" and the spec's §7.3 table says T1 is "PerlOnJava `unit/` (excl.
`regex/`), 986". Measured today against the local checkout:

    /home/perigrin/dev/PerlOnJava/src/test/resources/unit
      all .t                1508
      excluding regex/      1070
      regex/ alone           438

So 1,070 rather than 986 — an 84-file drift since the spec was written, which
is what a live upstream does. Neither the spec's number nor the plan's "~400"
should be copied into a target without saying which population it names.

This issue decides that: T1-easy is a NAMED SUBSET, chosen by measurement
rather than by estimate, and the choice is recorded with the command that
produced it. A percentage of an unstated denominator is not a target.

## What ratchets and what gates

    parse without error   RATCHET   moves in either direction, recorded
    WRONG                 GATE      any nonzero value fails

The distinction is the plan's and it is the discipline the ladder rests on. A
parse-rate that only climbs is a number nobody audits; a WRONG that is
allowed to climb makes the whole harness advisory.

Reuse `internal/lexer`'s ratchet shape: per-file counts, a committed
baseline, failure in either direction, regenerated under an explicit flag.
Not `parseoracle`'s — that one is bound to its own buckets, as M0's final
issue recorded when it declined the same reuse.

## Measuring through the subject contract

`internal/parseoracle` already measures ANY implementation through a
subprocess emitting JSON facts, and perl measured against itself through that
contract scores 100% exact — a calibration, not a demo. This parser becomes
another subject: a small command that lexes, parses, and reports the five
markers it can answer.

That is what makes "oracle exact ≥ 40% of markers attempted" checkable
without a second harness.

## Acceptance

- [ ] T1-easy is defined by a recorded command, not an estimate (`go test ./internal/ast/ -run '^TestT1EasyIsDefinedByMeasurement$' -count=1 -v | grep -q '^--- PASS: TestT1EasyIsDefinedByMeasurement'`)
- [ ] All 56 T2 core files parse without an Unknown at statement level (`go test ./internal/ast/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
- [ ] The parse-rate baseline fails on movement in either direction (`go test ./internal/ast/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [ ] Everything that parses also round-trips from the tree (`go test ./internal/ast/ -run '^TestParsedFilesRoundTrip$' -count=1 -v | grep -q '^--- PASS: TestParsedFilesRoundTrip'`)
- [ ] The parser answers the fidelity harness as a subprocess subject (`go test ./internal/parseoracle/ -run '^TestGoParserSubjectMeasures$' -count=1 -v | grep -q '^--- PASS: TestGoParserSubjectMeasures'`)
- [ ] WRONG is zero across the measured corpus, and that is a gate (`go test ./internal/parseoracle/ -run '^TestGoParserWrongIsZero$' -count=1 -v | grep -q '^--- PASS: TestGoParserWrongIsZero'`)

Plan: M1 metrics. Spec: §7.3, §7.5.
