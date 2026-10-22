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
created: 2026-10-22T23:46:34.622478576Z
updated: 2026-10-22T23:47:16.329385688Z
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
