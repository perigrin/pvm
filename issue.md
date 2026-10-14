---
title: A Boolean in a numeric slot draws a coercion warning, though perl numifies it silently
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T05:30:01.696995046Z
updated: 2026-10-14T13:00:59.945603284Z
---

## Context

- paths: `internal/types/` (Boolean's coercions), `internal/infer/lowered_ops.go` (`operandMismatch`)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-14 by cutover 1 (01a0c115-a44b), which surfaced 31 new warnings once `=~` and comparisons parse with perl's precedence: a Boolean in a numeric slot draws `coercion-mismatch ... expected Num, got Boolean`. perl 5.42 -w is silent: a Boolean numifies to 1 or 0 (`""` numifies to 0 without warning for perl's own booleans), and the idiom is common in perl's own lib: B/Deparse.pm:3429 `'()' x (7 < $cx)`, :1828 `(@kids != 1)/2`. `my $s = "-" x ($n > 2); my $q = ($n > 2) / 1` prints `[-] 1` with no warning. perigrin ruled 2026-10-14 to file it and fix after the cutover (01a0c115-df67).

**Rescheduled 2026-10-14:** perigrin ruled that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 (01a0c115-a44b) merges, so psc check output only improves at the cutover. Fix it on the lowered path (`internal/infer/lowered_*.go`); the legacy path is deleted by 01a0c115-df67.

## Steps

- [ ] Failing test first; Boolean is accepted in Num and Int slots (it coerces losslessly to 0/1); goldens under trailers; precision must not fall; `make test`; commit

## Acceptance criteria

- [ ] a Boolean in a numeric slot is no mismatch (`go test ./internal/infer/ -run '^TestBooleanInNumericSlotIsSilent$' -count=1 -v | grep -q '^--- PASS: TestBooleanInNumericSlotIsSilent'`)
