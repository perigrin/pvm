---
title: A union argument with a fitting member is a possible mismatch, silent by default; restore qualify's measured Ref return
state: in-progress
urgency: normal
milestone: core-module-types
created: 2026-10-21T02:33:29.838143913Z
updated: 2026-10-21T03:02:37.784165762Z
sessions:
- start_sha: 66180e7e19fc8ad1b81e1c0fb54bcc62942ffd1c
  end_sha: ""
  commits: 0
  started_at: 2026-10-21T03:02:37.784165762Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-21T03:02:37.784165762Z
---

perigrin, 2026-10-21, asked how psc check should treat an argument whose union type only partly fits a parameter: "Possible-mismatch, silent" -- treat it like possible-undef: when some member of the argument's union fits the parameter, stay silent by default and report only under --warn-lossy-coercion (or a new flag); a union with no fitting member is still reported.

The case that raised it: IO/Handle.pm:516 `() = tell qualify($_[0], caller) if ref($_[0]);`. Symbol::qualify's Ref candidate returns its argument unchanged (measured: `ref qualify({})` is HASH), so its correct return is Ref, not GlobRef; psc then reports "tell expects Glob, got Ref" though $_[0] is always a handle there and perl -w is silent. 01a15ac0-6ccc kept qualify's old GlobRef return pending this ruling (recorded in 01a15baf).

## Steps
- [ ] Record the ruling in RFC 0001 beside "Possible undef unless proven", dated and attributed
- [ ] In the argument check, an argument whose type has at least one member that fits the parameter is a possible mismatch: silent by default, reported under the chosen flag; no fitting member is still a coercion-mismatch
- [ ] Restore Symbol::qualify's Ref candidate to return Ref, as measured, and remove its deferral from 01a15baf
- [ ] psc check over perl 5.42's lib adds no unlisted line; precision holds

## Acceptance criteria

### Positive Scenarios
- [ ] A union argument with a fitting member is silent by default (`go test ./internal/infer/ -run '^TestPossibleMismatchSilentByDefault$' -count=1 -v | grep -q '^--- PASS: TestPossibleMismatchSilentByDefault'`)
- [ ] The flag reports it (`go test ./internal/infer/ -run '^TestPossibleMismatchReportedUnderFlag$' -count=1 -v | grep -q '^--- PASS: TestPossibleMismatchReportedUnderFlag'`)
- [ ] qualify returns the reference it was given (`go test ./internal/infer/ -run '^TestSymbolQualifyReturnsItsRef$' -count=1 -v | grep -q '^--- PASS: TestSymbolQualifyReturnsItsRef'`)
- [ ] The RFC records the ruling (`grep -q 'Possible mismatch' docs/rfcs/0001-type-libraries.md`)
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] Precision holds with WRONG 0 (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] A union argument with no fitting member is still reported (`go test ./internal/infer/ -run '^TestNoFittingMemberStillReported$' -count=1 -v | grep -q '^--- PASS: TestNoFittingMemberStillReported'`)
