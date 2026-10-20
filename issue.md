---
title: Give a declared typed our its type at use sites
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-20T19:19:04.406296656Z
updated: 2026-10-20T19:19:04.406296656Z
---

Quality lens I7, crochet:review of PR #488 (2026-10-21): internal/parse/declarations.go:82 and :534 read typed `our` variables from a .tpm (01a1464e-dadb), but internal/infer never consumes their types, so `$Config{...}`, `$Getopt::Long::...`, `$Text::Wrap::columns` and the rest stay Unknown at use sites. dadb deferred the consumer and no issue tracked it.

## Steps
- [ ] Lower an imported or qualified reference to a declared `our` with its declared type
- [ ] psc check over perl 5.42's lib adds no unlisted line

## Acceptance criteria

### Positive Scenarios
- [ ] An imported typed our carries its declared type (`go test ./internal/infer/ -run '^TestDeclaredOurTypesUseSites$' -count=1 -v | grep -q '^--- PASS: TestDeclaredOurTypesUseSites'`)
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
