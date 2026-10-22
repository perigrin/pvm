---
title: Report a ->{...} hash read's rv2hv exactly when perl keeps it out of multideref
state: pending
urgency: normal
milestone: m3-oracle-at-scale
created: 2026-10-09T11:07:58.504119457Z
updated: 2026-10-22T23:46:47.208621818Z
---

## Context

- paths: `internal/parse/sites.go`, `internal/parseoracle` (rv2hv counting), `docs/specs/perl-parser/00-findings.md` §0.11
- commands: `make test`; the oracle sweep at `PARSEORACLE_TIMEOUT=8m` with `-parseoracle.corpus`

The real fix the m1 hedge for arrow hash reads defers to (filed 2026-10-09). Whether perl compiles a `->{...}` hash read as a visible `rv2hv` or folds it into `multideref` depends on the shape of the chain and its keys (a computed key like `->{shift()}`, a read inside a block dereference `@{$r->{k}}`, `$#{ $x->{k} }`). internal/parse commits to one answer; m1 hedges the rv2hv marker for `->{...}` reads it cannot predict (the oracle's `wider`). This issue models perl's folding rule and reports rv2hv exactly. Seen in japh/abigail.t, op/gmagic.t, op/ref.t, porting/libperl.t.

## Steps

- [ ] Measure with Concise which `->{...}` chains perl folds into multideref and which keep rv2hv; failing test first; model the rule; drop the hedge; re-baseline; `make test`; commit

## Acceptance criteria

- [ ] an arrow hash read reports rv2hv exactly when perl keeps it (`go test ./internal/parse/ -run '^TestArrowHashReadFollowsMultideref$' -count=1 -v | grep -q '^--- PASS: TestArrowHashReadFollowsMultideref'`)
