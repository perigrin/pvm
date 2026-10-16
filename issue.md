---
title: substr's result is Str, but its 3-argument form returns undef past the end
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-08T04:57:29.01334186Z
updated: 2026-10-16T18:48:34.410711237Z
---

## Context

- paths: `internal/parse/declarations/CORE.tpm` (`substr`), `internal/parse/coretypes_test.go` (substr golden row)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found reviewing 01a1193c (2026-10-08). CORE.tpm types substr's result `Str`, but measured on 5.42 the 3-argument form past the end returns undef with the warning "substr outside of string" (`substr("abc", 5, 1)` is undef), while the 4-argument form past the end dies. So the 3-argument result is `Maybe[Str]` (Undef|Str). Whether that needs a multi (4-argument form never returns undef) or one signature with `Maybe[Str]` is a selection question: measure, and prefer one signature unless the result type genuinely differs by arity.

## Steps

- [ ] Measure substr's result in every arity, in and past the string, and as an lvalue
- [ ] Write failing test `TestCoreSubstrMayReturnUndef` stating the measured result types; change CORE.tpm; `make test`; commit

## Acceptance criteria

- [ ] substr's declared result admits undef where perl returns it (`go test ./internal/parse/ -run '^TestCoreSubstrMayReturnUndef$' -count=1 -v | grep -q '^--- PASS: TestCoreSubstrMayReturnUndef'`)
