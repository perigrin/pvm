---
title: "A reference in a Str slot is a valid coercion: drop the warning, add an opt-in lossy-coercion warning"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-13T23:22:39.523151838Z
updated: 2026-10-13T23:22:39.523151838Z
---

## Context

- paths: `internal/types/CORE.pmt` (Str slots: print, length, ...), `internal/infer/lowered_calls.go` (argument checks), `internal/infer/infer.go` (legacy), `internal/infer/options.go` (`Options`)
- commands: `go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$'`

perigrin ruled 2026-10-13: "Perl coerces the hashref to a string … so we should drop the warning on the input and add an optional warning on a lossy coercion from hashref to string." Today PSC reports `call to print: argument 1 expects Str, got HashRef` for `print $h` (and, since the calls port 01a12a85-32cf, for `print STDOUT $h`; golden `calls/print_ref_arg`). perl 5.42 -w is silent and prints `HASH(0x…)`: a reference is a valid Str by coercion. The coercion is lossy (the referent is not recoverable from the string), which is what an opt-in warning should name.

Related: `$h + 1` (golden `ops/paren_operand`) reports `left operand of "+": expected Num, got HashRef`; perigrin kept that one on 2026-10-13. Whether ref->Num joins this opt-in lossy-coercion class is a question to raise when this lands, not to decide here.

Do after the cutover (01a0c115-df67), on the one remaining path.

## Steps

- [ ] Failing tests first: a ref in a Str slot is no diagnostic by default; under the new option it is a `lossy-coercion` warning naming the ref type
- [ ] Option in `Options` (e.g. `WarnLossyCoercion`), off by default; psc flag if psc exposes Options
- [ ] `calls/print_ref_arg` golden edited under a `Diagnostics-changed:` trailer; precision must not fall; `make test`; commit

## Acceptance criteria

- [ ] a ref in a Str slot is silent by default (`go test ./internal/infer/ -run '^TestRefInStrSlotIsSilent$' -count=1 -v | grep -q '^--- PASS: TestRefInStrSlotIsSilent'`)
- [ ] the opt-in warns on a lossy ref-to-string coercion (`go test ./internal/infer/ -run '^TestLossyCoercionWarningIsOptIn$' -count=1 -v | grep -q '^--- PASS: TestLossyCoercionWarningIsOptIn'`)
