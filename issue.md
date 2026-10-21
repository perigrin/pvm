---
title: "Test-harness cuts the round-2 review left: one exemption map pair, Signature.ParamType, one oracle-run helper"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-21T17:04:58.388510343Z
updated: 2026-10-21T17:04:58.388510343Z
---

Deferred by 01a15ce6-1980 (crochet:review round 2 ponytail lens, PR #488): four test-harness cuts it did not attempt, each needing its own look at five-plus call sites under acceptance-criterion constraints.

- pendingAudit and parameterExempt are two mechanisms for one job: two flat map[string]string (return misfits, parameter misfits) would drop pendingMisfit, the nested loop, splitModuleSub and parameterExemptMatch (internal/infer/returntypes_test.go:217-257, parametertypes_test.go).
- paramTypeAt copies types.Signature.paramType: export ParamType (internal/infer/parametertypes_test.go:25-41, internal/types/select.go).
- One moduleOracleRun helper for the five copies of read-tpm, find-cases, build-exempt (parametertypes_test.go, returntypes_test.go holdReturnTypes, moduleFailures, TestOOModulesReturnTypes).
- TestFileStatTakesHandles duplicates TestOracleCasesFitDeclaredParameters/File::stat.
- FromPvm: kept; TestIsReferenceBuildOnPvmsOwnPerl cannot use IsReferenceBuild as its own precondition (01a15ce6-1980's reason).

## Steps
- [ ] Make each cut, one per commit, keeping every acceptance-criterion test name and meaning

## Acceptance criteria

### Positive Scenarios
- [ ] Signature exports ParamType (`grep -q 'func (s Signature) ParamType' internal/types/select.go`)

### Negative Scenarios
- [ ] paramTypeAt is gone (`! grep -q 'func paramTypeAt' internal/infer/parametertypes_test.go`)
