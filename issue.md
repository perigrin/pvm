---
title: State the M3 gate's measured numbers in the plan and READMEs
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5da6-79c8-aea5-e13d8f14c679
- 01a16602-5e01-7e7a-92c5-27d7da7d6f36
- 01a16602-5f1a-7d56-a91d-5fa2c478d608
- 01a16602-5fd5-7cc4-9645-4539ca73e839
- 01a16602-6030-7faa-8408-bb3a8cb03f12
- 01a16602-608d-71ed-8ab2-66a9951d9957
- 01a16602-60e9-7343-a630-7a93da5ee389
- 01a16602-6146-76ab-bb68-f3e8b645ae79
- 01a16602-61a2-78de-a798-da4cef3208c5
created: 2026-10-22T23:46:34.622478576Z
updated: 2026-10-22T23:56:40.77601534Z
---

## Context

- paths: `docs/plans/2026-09-05-parser-conformance-plan.md` (§5 M3 at ~606, the ladder summary at ~729), `docs/specs/perl-parser/README.md` (Status at ~99, the stale `($x)` sentence at ~129), `internal/parseoracle/README.md` (~287)
- commands: the milestone resolution command (`git zhi milestone show m3-oracle-at-scale --format json`, field `resolution`)
- entrypoints: `internal/parseoracle/testdata/ratchet/sites.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`, `internal/parse/testdata/faithful_defects.txt`, `internal/parseoracle/testdata/opcount/`

State the M3 gate's measured numbers where readers look, taken from the committed ratchets rather than from memory. The plan's M3 table says "≥ 526 of 584" and its ladder row says "584 files". internal/parseoracle/README.md:287 says "584/620 (94.2%)". All three are stale: a fuller shim now compiles 593. The README's defect sentence that `($x)` loses its `Paren` node (term.go:284-293) is also stale as worded, because `finishList` now copies the node and sets `Paren`.

## Steps

- [ ] Read the final numbers from sites.txt (whole corpus and M3-dir exact/wider/no-answer/WRONG, rv2hv wider), TestParseWithoutErrorOnCompiledCorpus's log (clean of 593), perlgittfaithful.ratchet (faithful of 619, M3 dirs), faithful_defects.txt (rows by cause) and the opcount fixture count. Re-run each test to confirm the numbers.
- [ ] Plan §5 M3: change the parse row to 593 and add a `#### Measured at the M3 gate` subsection giving each metric's measured value, the commit it was measured at, and where the remaining re-emission defects are listed. State plainly that re-emission is not yet 100% if defects remain.
- [ ] Plan ladder summary: change the M3 row's 584 to 593. Confirm with `grep -nw 584` that no 584 is left anywhere in the plan.
- [ ] docs/specs/perl-parser/README.md: add a Status paragraph starting `The M3 gate` with the measured numbers, and correct the stale `($x)` sentence to describe the current cause: canon writes source parens only for a `List`, canon now writes a `Paren` node's parens once, and `Faithful` forgives only extra nesting.
- [ ] internal/parseoracle/README.md:287: replace "584/620 (94.2%)" with the current measured count, 593/620, and its percentage.
- [ ] Run the milestone resolution command and watch it pass, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The plan has a measured-at-the-gate subsection for M3 (`grep -q '^#### Measured at the M3 gate' docs/plans/2026-09-05-parser-conformance-plan.md`)
- [ ] The stale 584 is gone from the plan (`sh -c '! grep -qw 584 docs/plans/2026-09-05-parser-conformance-plan.md'`)
- [ ] The perl-parser README's Status states the M3 gate (`grep -q '^The M3 gate' docs/specs/perl-parser/README.md`)
- [ ] The parseoracle README states 593 compiled files (`grep -q '593/620' internal/parseoracle/README.md`)
- [ ] The parseoracle README no longer states 584/620 (`sh -c '! grep -q 584/620 internal/parseoracle/README.md'`)
- [ ] The perl-parser README no longer cites term.go:284-293 for the dropped Paren (`sh -c '! grep -q term.go:284-293 docs/specs/perl-parser/README.md'`)

### Negative Scenarios

- [ ] The measured subsection's numbers are the committed ones, not a heading over stale prose. It states the whole-corpus and M3-dir site totals and the rv2hv wider count exactly as sites.txt records them, so editing either side alone fails (`go test ./internal/parseoracle/ -run '^TestPlanM3SectionQuotesSitesRatchet$' -count=1 -v | grep -q '^--- PASS: TestPlanM3SectionQuotesSitesRatchet'`)
- [ ] The measured subsection states 593, so deleting the stale row cannot satisfy the "no 584" check (`sh -c 'sed -n "/^#### Measured at the M3 gate/,/^##/p" docs/plans/2026-09-05-parser-conformance-plan.md | grep -qw 593'`)
- [ ] When faithful_defects.txt still holds rows, the measured subsection points to it, so remaining re-emission defects cannot read as 100% (`sh -c 'if [ -s internal/parse/testdata/faithful_defects.txt ]; then sed -n "/^#### Measured at the M3 gate/,/^##/p" docs/plans/2026-09-05-parser-conformance-plan.md | grep -q faithful_defects.txt; fi'`)
- [ ] The README's M3 gate paragraph states 593, not only the heading words (`sh -c 'grep "^The M3 gate" docs/specs/perl-parser/README.md | grep -qw 593'`)
- [ ] The parseoracle README gives 593/620 with its correct percentage, 95.6%, not the stale 94.2% beside the new count (`sh -c 'grep -q "593/620 (95.6%)" internal/parseoracle/README.md && ! grep -q "94.2%" internal/parseoracle/README.md'`)
- [ ] The corrected `($x)` sentence describes the extra-nesting ruling, not just a removed citation (`sh -c 'grep -qi "extra nesting" docs/specs/perl-parser/README.md'`)
