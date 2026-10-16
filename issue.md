---
title: "Guide: writing and testing a core-module declaration"
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d980-7382-970e-067a5ea3db84
- 01a1464e-dadb-7796-a704-4007150a1bf7
created: 2026-10-16T20:11:35.85395462Z
updated: 2026-10-16T20:11:51.639208723Z
---

## Prerequisites

- [ ] Pin the core-module ranking in a test and record the rule in the RFC (01a1464e-d980)
- [ ] Read typed variables with :export and :export_ok in declaration files (01a1464e-dadb)
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle (01a1464e-db1c)

## Context

- paths: `docs/contributing/core-module-declarations.md` (new), `docs/contributing/coding-conventions.md`, `docs/rfcs/0001-type-libraries.md`, `internal/parse/declarations/`, `internal/parse/ranking_test.go`, `internal/parse/heldtoperl_test.go`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/`, `internal/parse/testdata/core_module_ranking.txt`, `internal/parse/testdata/core_module_not_shipped.txt`
- commands: `test -f docs/contributing/core-module-declarations.md`
- urgency: low

This milestone's procedure for a core-module `.tpm` lives in issue bodies (the ranking rule in 01a1464e-d980, the three perl-held checks in 01a1464e-db1c, typed variables in 01a1464e-dadb, the infrastructure measurement in the Exporter/XSLoader/DynaLoader issues). Once the milestone closes nobody reads those. The next contributor -- adding module 51, or re-measuring on a new perl -- needs one page in `docs/contributing/` that says how to write and test a core-module declaration. The RFC states the design; this guide states the procedure and points at the RFC for why. Do not restate the RFC.

## Steps

- [ ] Write `docs/contributing/core-module-declarations.md` with ABOUTME lines and these sections, each naming the file or test it describes:
  - **Choosing the module**: the ranking rule as `TestCoreModuleRanking` computes it, how to regenerate `core_module_ranking.txt`, and the tie-break
  - **Writing the `.tpm`**: where it goes (`internal/parse/declarations/<Module/Path>.tpm`), every name in `@EXPORT`, `@EXPORT_OK` and `%EXPORT_TAGS`, prototypes, return types, `None` for subs that never return, typed variables with `:export`/`:export_ok`, and recording each measurement in a comment beside its line
  - **Holding it to perl**: the three tests a batch adds (`Test<Batch>Prototypes`, `Test<Batch>Exports`, `Test<Batch>ReturnTypes`), that export lists are read after `M->import` (POSIX builds them there), the oracle's call cases in `internal/infer/testdata/oracle/<Module>.pl`, and the uncallable map in `returntypes_test.go`: which subs go in it and that each entry needs a reason
  - **Modules other modules depend on**: why Exporter, XSLoader and DynaLoader are measured before shipping, and `core_module_not_shipped.txt` for a module that is not shipped
  - **Verifying**: the commands to run, with perl 5.42.0 from pvm, and that the perl-held tests skip without it
- [ ] Link it from `docs/contributing/development-workflow.md` or `coding-conventions.md`, whichever lists contributor guides; commit

## Acceptance Criteria

- [ ] the guide exists (`test -f docs/contributing/core-module-declarations.md`)
- [ ] it covers the uncallable list (`grep -q 'returntypes_test.go' docs/contributing/core-module-declarations.md`)
- [ ] it covers export coverage read after import (`grep -q 'EXPORT_TAGS' docs/contributing/core-module-declarations.md`)
- [ ] it covers the ranking (`grep -q 'core_module_ranking.txt' docs/contributing/core-module-declarations.md`)
- [ ] it covers the return-type oracle (`grep -q 'testdata/oracle' docs/contributing/core-module-declarations.md`)
