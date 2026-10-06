---
title: "Parse Fidelity ratchet: 71 perl.git t/ files regressed against the baseline, unseen since September"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T04:20:27.880192443Z
updated: 2026-10-06T04:20:27.880192443Z
---

## Context

- paths: `internal/parseoracle` (`TestRatchetCorpus`, its committed baseline), `.github/workflows/parse-fidelity.yml`
- commands: the Parse Fidelity workflow; locally `PERL5_CORPUS=<perl5 checkout at the pin> go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -count=1` (about 6 minutes on 24 workers)

The Parse Fidelity ratchet fails on pu (cdce17ba, 2026-09-29) and identically on PR #483: 71 perl.git t/ files regressed against the committed baseline and 148 improved without the baseline recording it. The regressed and improved sets are the same file-for-file on both runs, so the drift predates PR #483. Most regressions are exact -> wider ("perl read a hash at N site(s) the subject did not ... marked the statement unresolved"), and a few are -> WRONG (comp/proto.t, comp/parser.t, lib/universal.t, lib/deprecate.t, op/coreamp.t, io/scalar_ungetc.t). The workflow does not gate, so this went unseen.

The baseline must not simply be re-recorded: the WRONG files are real fidelity losses and each needs a cause.

## Steps

- [ ] Run the sweep locally at the pin; confirm the 71/148 split
- [ ] For each -> WRONG file, find the commit that moved it (bisect over origin/pu's history) and file or fix the cause
- [ ] For the exact -> wider group, find the shared cause (likely one hash-read site kind)
- [ ] Once every regression is fixed or individually justified, re-record the baseline with the improvements; commit

## Acceptance criteria

- [ ] the fidelity ratchet passes on pu (`sh -c 'gh run list --branch pu --workflow "Parse Fidelity" --limit 1 --json conclusion --jq ".[0].conclusion" | grep -qx success'`)
