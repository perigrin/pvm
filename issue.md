---
title: A value that might be Inf draws a coercion warning wherever it is used
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-13T15:58:43.197699471Z
updated: 2026-10-14T13:04:18.734313404Z
sessions:
- start_sha: 1feeef97410aa0897fde9a5c646b36484162cc2b
  end_sha: ""
  commits: 0
  started_at: 2026-10-14T13:04:18.734313404Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T13:04:18.734313404Z
---

## Context

- paths: `internal/types/CORE.pmt` (`/` returns `Num|Inf`), `internal/infer/infer.go` (`checkBinaryOperand`), `internal/infer/lowered_ops.go` (`operandMismatch`, `countType`)
- commands: `go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$'`

Found 2026-10-13 by the operator port (01a12a85-3233); perigrin ruled to file it and merge the port. Division returns `Num|Inf` because it can overflow, and Inf is not in Num, so any operand that MIGHT be Inf draws a coercion-mismatch: `my $q = $lvl / 8; $q + 1` and `"\t" x ($lvl / 8)` both warn, on both paths. perl 5.42 -w is silent on both (20/8 + 1 is 3.5); Inf arises only on real overflow (`1e308 / 1e-10` is `Inf`), and perl warns only on a count that IS non-finite ("Non-finite repeat count does nothing"). The golden `ops/repeat_div` keeps the interim warning (`expected Num, got Int|Num|Inf`) until this lands.

Proposed rule (the operator agent's option A): an operand that may be Inf from arithmetic is checked as finite; only a value KNOWN to be Inf or NaN is reported. This changes operand checking in general, so it lands after the cutover (01a0c115-df67), on the one remaining path.

**Rescheduled 2026-10-14:** perigrin ruled that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 (01a0c115-a44b) merges, so psc check output only improves at the cutover. Fix it on the lowered path (`internal/infer/lowered_*.go`); the legacy path is deleted by 01a0c115-df67. Cutover 1 measured 388 added lines.

**Ruled 2026-10-14:** perigrin chose the proposed rule: a type that only MIGHT be Inf (from arithmetic) is checked as finite; only a value KNOWN to be Inf or NaN is reported. Arithmetic results stay typed `Num|Inf`.

## Steps

- [ ] Failing test first; possible-Inf checked as finite, known Inf/NaN still reported; `ops/repeat_div` golden edited under a trailer; precision must not fall; `make test`; commit

## Acceptance criteria

- [ ] arithmetic that might overflow is no operand mismatch (`go test ./internal/infer/ -run '^TestMaybeInfIsCheckedAsFinite$' -count=1 -v | grep -q '^--- PASS: TestMaybeInfIsCheckedAsFinite'`)
