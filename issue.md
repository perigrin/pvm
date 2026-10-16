---
title: A .pmt parameter type must fit its slot, and aliased parameters print with their backslash
state: done
urgency: normal
milestone: type-libraries
created: 2026-10-06T08:53:38.742701359Z
updated: 2026-10-16T18:48:33.115231586Z
sessions:
- start_sha: 26bdde901a32213a3a9b612e036893c409c965ee
  end_sha: 26bdde901a32213a3a9b612e036893c409c965ee
  commits: 0
  started_at: 2026-10-06T13:46:23.392180657Z
  ended_at: 2026-10-06T13:46:23.612483797Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-06T13:46:23.392180657Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-06T13:46:23.612483797Z
---

## Context

- paths: `internal/parse/decl.go` (typed parameter reader, messages ~159/172), `internal/parse/derive.go`, `internal/types/select.go` (`Signature.params()`)
- commands: `make test`

Found by the alignment gate on 01a0fd9b-4850-704d-8c14-c15add6194f2. A .pmt parameter's type is not checked against its slot:

1. `Array \$x`, `Code \$x` and `Str \&c` parse; `Str \&c` derives `\&` with Type Str. Unaliased `Str &c` and `Str *g` are also accepted. `Array $x` derives `+` while `Array \$x` derives `\$`, and `:prototype(\$) (Array \$x)` passes agreement.
2. Messages drop the backslash: `Str \(@a)` suggests `List[Str] @a` (losing the alias); `Signature.params()` and the decl.go messages print `@a` for `\@a`, so the invalid spelling `Array @a` appears in error text.

## Steps

- [ ] Decide each slot's admissible types from the RFC table (`&` -> Code, `*` -> Glob, `\$` -> a Scalar type, `\@` -> Array, `\%` -> Hash); failing tests first; refuse a mismatch with a message naming slot and type; `make test`; commit
- [ ] Make the messages and `params()` print the backslash for aliased parameters; failing tests first; `make test`; commit

## Acceptance criteria

- [ ] a parameter whose type its slot cannot hold is refused (`go test ./internal/parse/ -run '^TestTpmSlotTypeMismatchRefused$' -count=1 -v | grep -q '^--- PASS: TestTpmSlotTypeMismatchRefused'`)
- [ ] aliased parameters print with their backslash (`go test ./internal/types/ -run '^TestSignatureParamsPrintAlias$' -count=1 -v | grep -q '^--- PASS: TestSignatureParamsPrintAlias'`)
