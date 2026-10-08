---
title: Arithmetic in CORE.pmt can yield Inf and NaN, which the lattice places outside Num
state: done
urgency: normal
milestone: type-libraries
created: 2026-10-06T08:56:07.912685152Z
updated: 2026-10-08T01:04:11.89326804Z
sessions:
- start_sha: 71a78d587b054f357c33464a4ed57869aa943b0f
  end_sha: 71a78d587b054f357c33464a4ed57869aa943b0f
  commits: 0
  started_at: 2026-10-08T01:04:11.855463565Z
  ended_at: 2026-10-08T01:04:11.89326804Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-08T01:04:11.855463565Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-08T01:04:11.89326804Z
---

## Context

- paths: `internal/types/types.go` (NaN, Inf, Num), `internal/parse/declarations/CORE.pmt` (arithmetic operators), `internal/parse/coreops_test.go` (`TestCoreOperatorTypesMatchMeasured`)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`; the paper at ~/dev/types-paper

Found by the alignment gate on 01a0fd9b-5468-7355-a22e-1b16fa5ee354. The lattice follows the paper in placing NaN and Inf under Str, not Num (`TestPaperNaNInfAreStrings`). So arithmetic declared `(Num $x, Num $y) Num` exceeds its bound: measured on 5.42, `1e308*10` is Inf and `(-1)**0.5` is NaN. And `<=>`'s `Int|Undef` result is undef only for a NaN operand, which its own `Num` parameters do not admit. TestCoreOperatorTypesMatchMeasured's comment was corrected to say so rather than claim every result is in bound.

## Steps

- [ ] perigrin decides: arithmetic results declared `Num|NaN|Inf` (honest, wider), or the declarations stay `Num` with the overflow/NaN cases documented as the paper's open question, or the paper revisits where NaN/Inf sit
- [ ] Apply the decision to CORE.pmt and the measured-types test; `make test`; commit

## Acceptance criteria

- [ ] every arithmetic operator's declared result admits what perl 5.42 measures, or the exception is recorded in RFC 0001 (`go test ./internal/parse/ -run '^TestCoreOperatorTypesMatchMeasured$' -count=1 -v | grep -q '^--- PASS: TestCoreOperatorTypesMatchMeasured'`)
