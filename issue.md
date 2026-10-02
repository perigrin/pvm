---
title: Type CORE.pmt's builtins from signatures.go, comments included
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fa46-a8d1-77c5-97d8-66cf3b4ec55d
- 01a0fa46-be8b-7823-9c6c-5be1406b57b8
- 01a0fa46-ece2-7c78-87ed-84da0b5d65f9
- 01a0fa46-f147-75f6-bd3c-4aa5f5644ee5
- 01a0fa46-97e0-7526-bc54-398af57e9df4
created: 2026-10-02T01:42:31.374841623Z
updated: 2026-10-02T01:42:37.300021156Z
---

`internal/types/signatures.go` holds 27 typed builtins ported from Chalk's TypeLibrary, used by internal/infer. Each becomes a typed line in CORE.pmt, with the measurement comments above it (why `push` takes `List`, why `split` accepts `Regex|Str`). A test holds every signatures.go builtin equal to its CORE.pmt declaration, so the two cannot drift until infer reads CORE.pmt and signatures.go goes. Element types the paper has not defined are written as plain `List`.

## Acceptance criteria

- [ ] Every signatures.go builtin has an equal typed CORE.pmt line (`go test ./internal/parse/ -run '^TestCoreTypesMatchSignatures$' -count=1 -v | grep -q '^--- PASS: TestCoreTypesMatchSignatures'`)
