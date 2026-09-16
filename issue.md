---
title: "The M1 gate: parse ratchet, WRONG stays zero, measured as a subject"
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a75b-5e51-79f1-8575-3f30abd05ac9
- 01a0a70f-3c3d-7ff7-aa8b-e59b91d899ac
- 01a0a75b-95b2-730a-9303-b6aba9adb26c
blocks:
- 01a0a8cd-8fcc-72a8-9ba1-c916eee9992d
- 01a0a8d8-e100-79ca-8680-811c86efd39d
created: 2026-09-15T21:53:08.812269582Z
updated: 2026-09-16T06:13:36.535503896Z
---

M1's gate: the parse rate ratchets, WRONG does not move, and every row of the
plan's metric table has an owner.

## Naming the corpora, because two numbers were being confused

The spec's §7.3.5 table is right and unambiguous:

| Location | Files |
|---|---:|
| `src/test/resources/unit/` **(top level)** | 986 |
| `unit/regex/` | 438 |
| `unit/refcount/`, `overload/`, `pack/`, `string/`, `runtime/`, `lib/`, `fixtures/` | the rest |

Measured today:

    ls  .../unit/*.t                              986   <- the spec's number
    find .../unit -name '*.t'                    1508
    find .../unit -name '*.t' -not -path '*/regex/*'  1070

986 is exactly the top level. An earlier draft of this issue claimed an
"84-file drift" by comparing the recursive count against the top-level one —
there is no drift, and that draft has been corrected. T1 is the 986.

**T1-easy** is a named subset of those 986, chosen by measurement and
recorded with the command that produced it. The plan's "~400" is an estimate,
not a definition; a percentage of an unstated denominator is not a target.

## T2 = 100% is the hardest row, and the scope has to admit what it needs

All 56 T2 files must parse. Measured against what M1's issues own:

    class / field / method      12 files
    use / no                    18 files
    BEGIN                       31 files
    prototype declarations       4 files (grep for prototype chars)
                                11 files (grep for any paren after subname)
                                46 prototype decls in comp/proto.t alone
    sort / map / grep            9 files
    method calls (->)           24 files

`use`, `no` and `BEGIN` need only to PARSE here, not to execute — a `use`
statement is a declaration with a module name and an optional list, and this
milestone owns its syntax. Prototype RECOGNITION is in the declarations
issue. Prototype RESOLUTION is M4's and `Call{Resolved:false}` covers the gap
without a WRONG.

`class`/`field`/`method` and `sort`/`map`/`grep` blocks are NOT owned by any
M1 issue. Either an issue is added for them or this row is stated as a
measured number rather than 100%. Do not carry an unreachable 100% into a
gate: that is how a gate becomes advisory.

## Every metric row gets an owner

The plan's table, and where each row lands:

| Row | Owner |
|---|---|
| 100% of T2 (56) | this issue |
| ≥ 70% of T1-easy | this issue — needs T1-easy defined first |
| Round-trip 100% of what parses | this issue |
| Oracle WRONG = 0 | this issue |
| Oracle exact ≥ 40% of markers attempted | this issue |

The last row is not decoration. WRONG = 0 is trivially satisfiable by a
parser that answers nothing — every M0 lesson about guards that cannot fail
applies. The exact floor is what makes WRONG = 0 mean something.

## What ratchets and what gates

    parse without error   RATCHET   moves in either direction, recorded
    WRONG                 GATE      any nonzero value fails

Follow `internal/lexer`'s ratchet shape: per-file counts, committed baseline,
failure in either direction, regenerated under an explicit flag
(`-lexer.update-ratchet`). `parseoracle`'s ratchet has the same
fail-both-directions discipline but is keyed on its four verdict buckets, so
its table shape does not transfer; the discipline does.

This would be the third ratchet in the tree (`internal/lexer`,
`internal/parseoracle`, this one). Decide explicitly rather than by default:
either extract the shared shape, or record in the test file why a third copy
is still cheaper than the abstraction. Three copies is where "two is a
coincidence" stops being true.

## Measuring through the subject contract

`internal/parseoracle` measures any implementation through a subprocess
emitting JSON facts, and perl measured against itself scores 100% exact — a
calibration. This parser becomes another subject.

The sweep skips silently without a corpus flag, so the AC commands carry
`-parseoracle.corpus` explicitly rather than relying on a default.

## Acceptance

- [ ] T1-easy is defined by a recorded command over the 986 top-level files (`go test ./internal/parse/ -run '^TestT1EasyIsDefinedByMeasurement$' -count=1 -v | grep -q '^--- PASS: TestT1EasyIsDefinedByMeasurement'`)
- [ ] All 56 T2 core files parse, or the shortfall is named per file (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
- [ ] At least 70% of T1-easy parses without error (`go test ./internal/parse/ -run '^TestT1EasyParseRate$' -count=1 -v | grep -q '^--- PASS: TestT1EasyParseRate'`)
- [ ] The parse-rate baseline fails on movement in either direction (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [ ] Everything that parses also round-trips from the tree (`go test ./internal/parse/ -run '^TestParsedFilesRoundTrip$' -count=1 -v | grep -q '^--- PASS: TestParsedFilesRoundTrip'`)
- [ ] The parser answers the fidelity harness as a subprocess subject (`go test ./internal/parseoracle/ -run '^TestGoParserSubjectMeasures$' -count=1 -v -parseoracle.corpus=t2 | grep -q '^--- PASS: TestGoParserSubjectMeasures'`)
- [ ] WRONG is zero across the measured corpus, and that is a gate (`go test ./internal/parseoracle/ -run '^TestGoParserWrongIsZero$' -count=1 -v -parseoracle.corpus=t2 | grep -q '^--- PASS: TestGoParserWrongIsZero'`)
- [ ] Exact is at least 40% of markers attempted, so WRONG=0 is not vacuous (`go test ./internal/parseoracle/ -run '^TestGoParserExactFloor$' -count=1 -v -parseoracle.corpus=t2 | grep -q '^--- PASS: TestGoParserExactFloor'`)

Plan: M1 metrics. Spec: §7.3.5, §7.5.
