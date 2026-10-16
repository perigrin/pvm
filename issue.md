---
title: "Guide: writing and testing a core-module declaration"
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d980-7382-970e-067a5ea3db84
- 01a1464e-dadb-7796-a704-4007150a1bf7
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:11:35.85395462Z
updated: 2026-10-16T21:00:54.288843104Z
---

## Prerequisites

- [ ] Pin the core-module ranking in a test and record the rule in the RFC (01a1464e-d980)
- [ ] Read typed variables with :export and :export_ok in declaration files (01a1464e-dadb)
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle (01a1464e-db1c)
- [ ] Convert every .tpm to :export and :export_ok, then refuse export lists in a .tpm (01a14683-8375-7810-8a75-b772680603c6)

## Context

- paths: `docs/contributing/core-module-declarations.md` (new), `docs/contributing/coding-conventions.md`, `docs/rfcs/0001-type-libraries.md`, `internal/parse/declarations/`, `internal/parse/ranking_test.go`, `internal/parse/heldtoperl_test.go`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/`, `internal/parse/testdata/core_module_ranking.txt`, `internal/parse/testdata/core_module_not_shipped.txt`
- commands: `test -f docs/contributing/core-module-declarations.md`
- urgency: low

This milestone's procedure for a core-module `.tpm` lives in issue bodies (the ranking rule in 01a1464e-d980, the three perl-held checks in 01a1464e-db1c, typed variables in 01a1464e-dadb, the infrastructure measurement in the Exporter/XSLoader/DynaLoader issues). Once the milestone closes nobody reads those. The next contributor -- adding module 51, or re-measuring on a new perl -- needs one page in `docs/contributing/` that says how to write and test a core-module declaration. The RFC states the design; this guide states the procedure and points at the RFC for why. Do not restate the RFC.

## Steps

- [ ] Write `docs/contributing/core-module-declarations.md` with ABOUTME lines and these sections, each naming the file or test it describes:
  - **Choosing the module**: the ranking rule as `TestCoreModuleRanking` computes it, how to regenerate `core_module_ranking.txt`, and the tie-break
  - **Writing the `.tpm`**: where it goes (`internal/parse/declarations/<Module/Path>.tpm`); no export lists (the parser refuses `@EXPORT`, `@EXPORT_OK` and `%EXPORT_TAGS` in a `.tpm`); every name a `use` installs, marked `:export` or `:export_ok` with tags as attribute arguments (`:export_ok(flock)`), including names a module's own `import` installs (Locale::Maketext::Simple's `loc`); `&name` list entries are the sub `name`; prototypes, return types, `None` for subs that never return, typed variables (`our Hash[Str|Undef] %Config :export;`), and recording each measurement in a comment beside its line
  - **Holding it to perl**: the three tests a batch adds (`Test<Batch>Prototypes`, `Test<Batch>Exports`, `Test<Batch>ReturnTypes`), that exports are measured as what `use M`, `use M qw(name)` and `use M ':tag'` install in a caller, with `exists &name` because AUTOLOAD subs are not yet defined (POSIX), the oracle's call cases in `internal/infer/testdata/oracle/<Module>.pl`, and the uncallable map in `returntypes_test.go`: which subs go in it and that each entry needs a reason
  - **Modules other modules depend on**: why Exporter, XSLoader and DynaLoader are measured before shipping, that they are the only modules that may not ship and only perigrin rules one out, and `core_module_not_shipped.txt`'s `Module<TAB>full-uuid<TAB>reason` format
  - **Verifying**: the commands to run, with perl 5.42.0 from pvm, and that the perl-held tests skip without it
- [ ] Link it from `docs/contributing/development-workflow.md` or `coding-conventions.md`, whichever lists contributor guides; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] the guide exists (`test -f docs/contributing/core-module-declarations.md`)
- [ ] it covers the uncallable list (`grep -q 'returntypes_test.go' docs/contributing/core-module-declarations.md`)
- [ ] it covers exports as measured installs marked `:export_ok` (`grep -q ':export_ok' docs/contributing/core-module-declarations.md`)
- [ ] it covers the ranking (`grep -q 'core_module_ranking.txt' docs/contributing/core-module-declarations.md`)
- [ ] it covers the return-type oracle (`grep -q 'testdata/oracle' docs/contributing/core-module-declarations.md`)
- [ ] it states the not-shipped file's format (`grep -q 'core_module_not_shipped.txt' docs/contributing/core-module-declarations.md`)
