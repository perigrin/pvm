---
title: "[I17] Exits and guards are matched by exact name: croak, CORE::die and CORE::defined are missed, and CORE::push g..."
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.760410032Z
updated: 2026-10-16T01:16:07.760410032Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I17**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:**
  - `internal/infer/flow.go:354-371` (blockExits: only die, exit and return).
  - `flow.go:425-437` (callGuard).
  - `calls.go:38` (`CoreBuiltin(n.Name)` with the `CORE::` prefix still on).
- **Repro:**
  - `v/c6b.pl` (`unless (CORE::defined $y) { die }`) and `v/c6c.pl` (`unless (defined $y) { CORE::die }`) each give a new psc error `got Undef|Int`.
  - `v/c6a.pl` (croak, with PERL5LIB set to perl's privlib) gives the same.
  - origin/pu is silent on all three, and perl prints 4.
  - `v/eb6b.pl`: `CORE::push %h, 1` is silent in psc, but perl refuses it at compile time.
- **Fix:** strip `CORE::` (and `CORE::GLOBAL::`) before matching builtins. Treat Carp's croak and confess as exits, or declare no-return in .pmt.
- **Confidence:** High. **NEW.** **Issue:** none. **Found by:** errors-B #6.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] Exits and guards are matched by exact name: `croak`, `CORE::die` and `CORE::defined` are missed, and `CORE::push` goes unchecked is fixed (`go test ./internal/infer/ -run '^TestExitsAndGuardsByCoreAndCarp$' -count=1 -v | grep -q '^--- PASS: TestExitsAndGuardsByCoreAndCarp'`)
