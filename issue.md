---
title: Type CORE.pmt's builtins from signatures.go, comments included
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-02T01:42:31.374841623Z
updated: 2026-10-02T01:42:31.374841623Z
---

`internal/types/signatures.go` holds 27 typed builtins ported from Chalk's TypeLibrary, used by internal/infer. Each becomes a typed line in CORE.pmt, with the measurement comments above it (why `push` takes `List`, why `split` accepts `Regex|Str`). A test holds every signatures.go builtin equal to its CORE.pmt declaration, so the two cannot drift until infer reads CORE.pmt and signatures.go goes. Element types the paper has not defined are written as plain `List`.

## Acceptance criteria

- [ ] Every signatures.go builtin has an equal typed CORE.pmt line (`go test ./internal/parse/ -run '^TestCoreTypesMatchSignatures$' -count=1 -v | grep -q '^--- PASS: TestCoreTypesMatchSignatures'`)
