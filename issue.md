---
title: Commit a perl-judged probe corpus that every parser fix extends
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a1664d-69f0-715c-b060-c8f648ddfcd3
blocks:
- 01a16602-5ebe-7169-b857-996ae2953394
- 01a12059-30a8-722e-bbc3-52066b8610cb
- 01a1664d-7bc8-7498-af02-7eaf187b43ed
created: 2026-10-23T01:08:37.943092553Z
updated: 2026-10-23T01:09:11.162740379Z
---

## Context

- paths: `internal/parse/probes_test.go` (new), `internal/parse/testdata/m3probes/` (new), `internal/parse/testdata/m3probes/README.md` (new)
- commands: `PATH=$HOME/.local/perl-5.44.0/bin:$PATH go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 -v`
- entrypoints: `docs/postmortems/core-module-types.md` (action 3, ~line 48), `internal/parse/sites.go` (`Sites`)

The core-module-types postmortem's action 3 asks for the regression-probe corpus to be committed as a perl-judged table test that every later fix runs mechanically. A per-fix review that leaves only prose cannot be checked, and in that milestone, reviews that re-checked only the named finding let through regressions that the next lens found. This issue builds the table. Each probe file under `testdata/m3probes/` holds snippets, each with the facts perl reports for it: marker-site counts per marker (rv2hv, srefgen, anonhash, match, readline) and, where relevant, `Deparse,-p` grouping. TestM3ProbesHeldToPerl does two things for every snippet. It runs perl (the corpus.pin interpreter, 5.44.0) to confirm the recorded facts, and it checks that psc's parse agrees. A known disagreement is allowed only as a row marked `known-defect <issue>`. Every later parser, sites.go or canon fix in this milestone adds its own probe file, named in that issue's criteria. Tests run perl; production psc never does.

## Steps

- [ ] Create `internal/parse/probes_test.go` with two `// ABOUTME:` lines, and a README in testdata/m3probes/ describing the readable text format (snippet, perl facts, optional `known-defect` row).
- [ ] RED: write TestM3ProbesHeldToPerl. It loads every probe file and fails on an empty directory. It runs perl on each snippet and compares perl's facts with the recorded ones, then compares psc's facts with perl's. Seed one probe file, `baseline.txt`, from adversarial cases the reviews of this milestone's predecessors used (for example, `\(@a)` vs `\@a`, `my ($x) = @_`, and `keys %$r` in each context). Flip one recorded count, watch the test fail, and restore it.
- [ ] GREEN: implement the fact extraction on both sides by reusing the oracle's perl-side probe (internal/parseoracle/testdata/parse_facts.pl) rather than writing a second one. Run the test and watch it pass.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The seed probe file is committed (`test -s internal/parse/testdata/m3probes/baseline.txt`)
- [ ] Every probe's recorded facts match perl 5.44.0, and psc agrees except on rows marked known-defect (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 -v | grep -q '^--- PASS: TestM3ProbesHeldToPerl'`)
