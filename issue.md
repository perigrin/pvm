---
title: "Tier 12_packages: use and require, heavier than the object system"
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c362-073a-78c2-ba73-227fb48a26c4
created: 2026-09-21T09:52:58.866189072Z
updated: 2026-09-21T13:43:02.235715224Z
---

**Introduces:** package, use, require, imports

Later than the object system because these are the heavier construct: they
locate and load a FILE, then call `import`, which needs subroutines (tier
07) to mean anything. A tier list that puts them before the object system is
ordering by how a program is usually WRITTEN rather than by what the grammar
depends on.

Depends on nothing from `10_io` -- another of the independent pairs.

## The authoring constraint this tier explains

Of the 35 T1 files matching tier 04, 31 `use Test::More`. Every extracted
tier-04 file would contain this tier's construct and fail the dependency
check by construction. That is why everything is authored and T1 is a source
of CASES, cited by filename and pinned revision, never copied.

## Acceptance

- [ ] Every file in 12_packages/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierPackagesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierPackagesPerlValidated'`)
- [ ] The dependency lint passes over 12_packages/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierPackagesLint$' -count=1 -v | grep -q '^--- PASS: TestTierPackagesLint'`)
- [ ] The tier's adjacency file puts every construct 12_packages introduces in one body, each adjacent to another, pairing with the declared prerequisite rather than tier N-1 (`go test ./internal/conformance/ -run '^TestTierPackagesAdjacency$' -count=1 -v | grep -q '^--- PASS: TestTierPackagesAdjacency'`)
- [ ] 12_packages/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierPackagesRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierPackagesRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus/12_packages' -count=1 -v | grep -q '^    --- PASS: TestCorpus/12_packages'`)
