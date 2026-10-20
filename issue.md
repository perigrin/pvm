---
title: "Declaration reader refusals: export marks in blocks, list-form typed our, redeclaration disagreements, use vars export lists"
state: pending
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.38786537Z
updated: 2026-10-20T19:18:42.38786537Z
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
