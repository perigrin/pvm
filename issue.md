---
title: "Declaration reader refusals: export marks in blocks, list-form typed our, redeclaration disagreements, use vars export lists"
state: done
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.38786537Z
updated: 2026-10-20T20:58:58.035655799Z
sessions:
- start_sha: 3e1e18712a71295fc187f6cf99240b81c3f67124
  end_sha: fe62372eb0dce18610b696b53c41cde6a6086bd5
  commits: 3
  started_at: 2026-10-20T19:39:29.23442591Z
  ended_at: 2026-10-20T20:58:58.035655799Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T19:39:29.23442591Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-20T20:58:58.035655799Z
observed_paths:
- .github/workflows/ci.yml
- internal/infer/calls.go
- internal/infer/flow.go
- internal/infer/infer.go
- internal/infer/narrowing_regressions_test.go
- internal/infer/ops.go
- internal/infer/returntypes_test.go
- internal/infer/symbols.go
- internal/infer/vars.go
- internal/parse/decl.go
- internal/parse/declarations.go
- internal/parse/exportattr_test.go
- internal/parse/heldtoperl_test.go
- internal/parse/holdcoverage_test.go
- internal/parse/typedour_test.go
- internal/perltest/batches.go
- internal/perltest/fixture.go
- internal/perltest/fixture_test.go
- internal/perltest/perl542_test.go
- internal/perltest/perltest.go
- internal/psc/lib_baseline_test.go
---

crochet:review of PR #488, round 1 (2026-10-21). Quality S3 internal/parse/declarations.go:281,387 (export marks inside blocks silently dropped), S4 declarations.go:290-302 + decl.go:902-928 (list-form typed our and redeclaration mark/type disagreements pass silently), S5 declarations.go:376 (use vars '@EXPORT' escapes the export-list refusal). Quality I7 (typed-our types are never read by infer) is filed as its own follow-up outside this milestone.

Parameter rule (RFC 0001 "What a parameter type means", perigrin 2026-10-02): a parameter type is the type the operation coerces its argument to. A parameter is wrong when its type is not what the operation coerces to or preserves (a lone reference carp/warn/croak pass through; NaN/Inf a C double keeps), or it admits a value perl has no coercion for and dies on. A finding that only shows perl coercing a value (sum("abc"), encode with a ref) is withdrawn, not fixed. Every fix is measured with /home/perigrin/.local/bin/perl; never bend a type to quiet the lib baseline.

## Steps
- [ ] Each silent case becomes a load error naming the file and line, or is read correctly where perl gives it meaning

## Acceptance criteria

### Positive Scenarios
- [ ] TestListFormTypedOurIsRead (`go test ./internal/parse/ -run '^TestListFormTypedOurIsRead$' -count=1 -v | grep -q '^--- PASS: TestListFormTypedOurIsRead'`)

### Negative Scenarios
- [ ] TestExportMarkInsideBlockRefused (`go test ./internal/parse/ -run '^TestExportMarkInsideBlockRefused$' -count=1 -v | grep -q '^--- PASS: TestExportMarkInsideBlockRefused'`)
- [ ] TestRedeclarationDisagreementRefused (`go test ./internal/parse/ -run '^TestRedeclarationDisagreementRefused$' -count=1 -v | grep -q '^--- PASS: TestRedeclarationDisagreementRefused'`)
- [ ] TestUseVarsExportListRefused (`go test ./internal/parse/ -run '^TestUseVarsExportListRefused$' -count=1 -v | grep -q '^--- PASS: TestUseVarsExportListRefused'`)
