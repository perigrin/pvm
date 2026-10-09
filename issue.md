---
title: The subject pools every site under its top-level statement, so a hedge can excuse a site elsewhere
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T05:00:43.559152981Z
updated: 2026-10-09T05:00:43.559152981Z
---

## Context

- paths: `internal/parse/sites.go`, `internal/parse/cmd/subject`, `internal/parseoracle/compare_facts.go` (`innermost`, how sites are spent against perl's)
- commands: `make test`; the oracle's corpus ratchet `go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -count=1 -v -parseoracle.corpus`

Found by the WRONG triage of 2026-10-09. The subject reports every site at its top-level statement's start line, so a whole SKIP block or package block pools into one statement, and decided sites are spent in sorted line order: a finding can name a later sibling of the line actually missed, and a hedge anywhere in the pool excuses a WRONG elsewhere in it (measured: with sites spanning to statement end, op/sprintf2.t:452 went WRONG -> wider through such a hedge). Reporting nested statements at their own spans needs care -- an outer statement's moved line can land on an inner statement's start line, which `innermost` would then pick -- and may need an oracle rule change. A design question for perigrin: should the subject report sites per nested statement, and what does the oracle then match them against?

## Steps

- [ ] Measure how many verdicts the pooling decides (findings whose named line is not the line missed; hedges excusing a site in another statement)
- [ ] Bring perigrin the options; record the ruling in findings §0.11; implement with a failing test first

## Acceptance criteria

- [ ] sites are reported per statement as ruled, and no hedge excuses a site outside its own statement (`go test ./internal/parse/cmd/subject/ -run '^TestSitesArePerStatement$' -count=1 -v | grep -q '^--- PASS: TestSitesArePerStatement'`)
