---
title: "RFC and tracking gaps round 3 found: the possible-mismatch sha, named parameters as read-but-unconsumed, delete_package's glob, the rejected-alternatives contradiction"
state: done
urgency: normal
milestone: core-module-types
created: 2026-10-21T20:10:35.421794374Z
updated: 2026-10-21T22:32:04.829338211Z
sessions:
- start_sha: a893c4c7dbd73e99ee996512595b346f7c0d3243
  end_sha: 7e38336b25250af70458c96be7effee85a599655
  commits: 3
  started_at: 2026-10-21T20:24:18.757906316Z
  ended_at: 2026-10-21T22:32:04.829338211Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-21T20:24:18.757906316Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-21T22:32:04.829338211Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/infer/declared_test.go
- internal/infer/dualvar_test.go
- internal/infer/flow.go
- internal/infer/ops.go
- internal/infer/precision_test.go
- internal/infer/returntypes_test.go
- internal/infer/testdata/observe_types.pl
- internal/infer/testdata/oracle/Scalar-Util.pl
- internal/parse/declarations.go
- internal/parse/declarations/Cwd.tpm
- internal/parse/declarations/DynaLoader.tpm
- internal/parse/declarations/ExtUtils/Manifest.tpm
- internal/parse/declarations/FileHandle.tpm
- internal/parse/declarations/IO/Socket.tpm
- internal/parse/declarations/IO/Socket/IP.tpm
- internal/parse/declarations/Moose.tpm
- internal/parse/declarations/Scalar/Util.tpm
- internal/parse/declarations/Socket.tpm
- internal/parse/declarations/Symbol.tpm
- internal/parse/exportattr_test.go
- internal/parse/heldtoperl_test.go
- internal/parse/issue1959_test.go
- internal/parse/pin_test.go
- internal/parse/use.go
- internal/perltest/fixture.go
- internal/perltest/staleexempt.go
- internal/perltest/staleexempt_test.go
- internal/types/declared.go
- test/e2e/psc_check_test.go
- test/e2e/testdata/check/dualvar.pl
---

crochet:review round 3 of PR #488 at 692a560c (2026-10-21). Quality report: scratchpad/review3-core-module-types-paad.md; probes scratchpad/r3probes/ (F-<lens>.md per specialist); rebuild head/r2(2db182e6)/cf(cf08c0ac)/pu psc as the report says. perigrin, 2026-10-21: round 4 runs with Opus workers, and every fix gets a mandatory independent review that runs pu, round 2 and head against perl 5.42 before it merges. A finding pu or round 2 gets right is a regression: fix it without trading it for a new one -- every probe in r2probes/ and r3probes/ that pu or r2 gets right must still be right.

Coverage round 3: F1 (RFC :982 'Possible mismatch' still cites no sha; 7fc02ddf), F2 (RFC :1113-1146 presents named-parameter consumption and dash matching as implemented; since 56ecfcff they wait on 01a15f6b -- say so; also declared.go:89-96 and quality V11), F3 (S7 Symbol::delete_package returns a glob, measured `ref(\delete_package("Foo"))` is GLOB, declared Str|Undef; S10 Rejected alternatives names only CORE.tpm; S12 cite 01a15ac0-c1d6 for typed-our consumption), F4 (01a15f6b's criteria should include TestMooseMetaNameRenamesMeta), F5 (round-1 S11 TestPSCCheckUndecidedCallReportsNothing's second half has no positive control; round-1 S13 the hand-rolled declarations swaps -- give each a destination). Quality V10 (IO/Socket/IP.tpm installs all 170 names for any list; its comment is false -- correct the comment and add the case to 01a15f6b), V12 (v-string VERSION), V15 (export-mark scan misses else/elsif and expression-level do), V16-V18 (untracked residuals) -- each fixed here if small, otherwise given a destination issue.

## Steps
- [ ] Each finding fixed or given a named destination issue; no finding left untracked

## Acceptance criteria

### Positive Scenarios
- [ ] The RFC cites the possible-mismatch sha (`grep -A2 'Possible mismatch beside a fitting member' docs/rfcs/0001-type-libraries.md | grep -q 7fc02ddf`)
- [ ] TestSymbolDeletePackageReturnsGlob (`go test ./internal/infer/ -run '^TestSymbolDeletePackageReturnsGlob$' -count=1 -v | grep -q '^--- PASS: TestSymbolDeletePackageReturnsGlob'`)
- [ ] 01a15f6b names the meta-rename test (`git zhi issue show 01a15f6b | grep -q TestMooseMetaNameRenamesMeta`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] The RFC no longer presents named-key consumption as implemented (`! grep -q 'each named parameter is consumed from the use-list wherever it falls' docs/rfcs/0001-type-libraries.md`)
