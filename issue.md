---
title: "Read the :lvalue parameter attribute and derive the \\$ prototype from it"
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-02T17:13:27.376660947Z
updated: 2026-10-02T17:13:27.376660947Z
---

## Prerequisites

- [ ] Enforce the slurpy rule and caller-view sigils in .pmt signatures
- [ ] Derive prototypes from typed signatures and typed signatures from prototypes

## Context

- paths: `internal/parse/decl.go`, `internal/parse/derive.go`, `internal/parse/derive_test.go`, `internal/parse/pmt_test.go`, `internal/parse/call.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `internal/parse/derive.go`; the param reader in `internal/parse/decl.go`; `parseAttributes` (`internal/parse/use.go:815`) for how attributes are read today

RFC 0001 "The scalar container" (perigrin, 2026-10-02): a parameter that must be an assignable place, whose callee receives the place, carries `:lvalue` after the variable, as the class feature's `field $x :param` does. `sub sref (Scalar $x :lvalue);` derives `\$`, the table row the derivation issue left out. On the parse side the same attribute states `chomp`'s and `chop`'s in-place list, `List[Str] @args :lvalue`; what an lvalue list is as a type stays the paper's question. Measured on 5.42, `\$` accepts any scalar lvalue (`$x`, `$h{k}`, `$a[0]`, `f()->[0]`, `$x = 7`, `substr($x,0,1)`), and a constant or a sub's result dies at compile time: "must be scalar (not constant item)", "(not subroutine entry)". The parser accepts those calls today for any `\$` slot; refuse them as perl does, measuring each case first. `Array @a` and `Hash %h` need no `:lvalue`.

## Steps

- [ ] Write failing test `TestLvalueParamDerivesRefPrototype`: `sub sref (Scalar $x :lvalue);` derives `\$`; `sub sref :prototype(\$);` derives `(Scalar $ :lvalue)`; the round trip holds
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestLvalueParamDerivesRefPrototype$' -count=1 -v`
- [ ] Implement the param attribute and the derivation row; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPmtLvalueSlurpyParses`: `sub chomp (List[Str] @args :lvalue);` parses with the attribute on the slurpy
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Measure with perl, then write failing test `TestRefScalarSlotRefusesConstants`: after `sub sref (\$);`, `sref(1)`, `sref("s")` and `sref(f())` are refused; `sref($x)`, `sref($h{k})`, `sref($x = 7)` are not
- [ ] Implement in the call path; run test to verify pass; `make test` (watch the T1/T2 ratchets); commit
- [ ] Re-mark the RFC heading "The scalar container" `(*Implemented*, <sha>)`; commit

## Acceptance criteria

- [ ] an :lvalue scalar parameter derives the \$ prototype (`go test ./internal/parse/ -run '^TestLvalueParamDerivesRefPrototype$' -count=1 -v | grep -q '^--- PASS: TestLvalueParamDerivesRefPrototype'`)
- [ ] an :lvalue slurpy such as chomp's parses (`go test ./internal/parse/ -run '^TestPmtLvalueSlurpyParses$' -count=1 -v | grep -q '^--- PASS: TestPmtLvalueSlurpyParses'`)
- [ ] a constant or sub result passed to a \$ slot is refused, as perl refuses it (`go test ./internal/parse/ -run '^TestRefScalarSlotRefusesConstants$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotRefusesConstants'`)
- [ ] the RFC section on the scalar container is implemented (`grep -q '^### The scalar container.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
