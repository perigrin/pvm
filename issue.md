---
title: "Refuse constants and sub results in \\$ slots"
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4850-704d-8c14-c15add6194f2
blocks:
- 01a10177-547c-73a9-a302-f25f579a4cb3
created: 2026-10-05T20:28:48.303887564Z
updated: 2026-10-06T08:53:43.205184657Z
---

## Prerequisites

- [ ] Parse aliased parameters (\$x, \@a, \%h, \&c, \(@list)) and derive \$, \@, \% and \& from them

## Context

- paths: `internal/parse/shape.go`, `internal/parse/call.go`, `internal/parse/refusal.go`, `internal/parse/sites.go`, `internal/parse/refscalar_test.go` (new), `internal/parse/testdata/t1.ratchet`, `internal/parse/testdata/perlgitt.ratchet`, `internal/conformance/testdata/corpus.ratchet`
- commands: `go test ./internal/parse/ -count=1`
- commands: ratchets, one named test each (measured on this tree, 2026-10-05, with compile: `TestParseRatchet` 8.2s, `TestPerlGitTRatchet` 7.0s, `TestCorpusRatchet` 7.9s; `git-zhi verify` kills a command after 300s): `go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v`, `go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1 -v`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume; `perl -c -e '...'` for each case)
- entrypoints: `parseByShape` (`internal/parse/shape.go:268`), `subPrototype`/`lookupSub` (`shape.go:246`, `:162`), `parseCallArgs` (`internal/parse/call.go:699`), the refusal codes and their inventory `RefusalSites` (`internal/parse/refusal.go`)

RFC 0001 "Refusing what perl refuses" (perigrin, 2026-10-02), its first family: a `\$` slot takes a scalar lvalue, and the parser refuses what perl refuses there, as an Unknown with its own refusal code. Measured on 5.42 with `sub f {} sub sref (\$) {}`: `sref(1)` and `sref("s")` die "Type of arg 1 to main::sref must be scalar (not constant item)"; `sref(f())` "(not subroutine entry)"; `sref(@a)` "(not private array)"; `sref(%h)` "(not private hash)"; `sref()` "Not enough arguments for main::sref"; `sref($x, $y)` "Too many arguments for main::sref". `sref($x)`, `sref($h{k})`, `sref($a[0])`, `sref(f()->[0])`, `sref($x = 7)`, `sref(my $z)` and `sref(substr($x, 0, 1))` compile. The property is lvalue-ness, not "unevaluated". The rule applies to a `\$` from a prototype read from module source (`sub sref (\$)` in a `.pm`) as well as to a `.pmt` declaration's `Scalar \$x` (the aliased-parameter issue's). The other family, a call whose arity no declaration accepts (`select(1, 2)`, `grep()`, `sort()` through `= die`), is the arity refusal issue's, which follows this one and re-marks the section once both are built; this issue re-marks no section. Refusing more can move ratchet files: every file that moves is checked against `perl -c`, and the ratchet is updated only for moves that match perl.

Also from the alignment gate on 01a0fd9b-4850-704d-8c14-c15add6194f2: a .pmt may write `Scalar \$x = 1`, which derives `;\$` with a CONSTANT default -- one perl would refuse in a `\$` slot. Refuse a constant default on an aliased scalar parameter here too, with a failing test.

## Steps

- [ ] Measure each case with `perl -c`, then write failing test `TestRefScalarSlotRefusesConstants` in `internal/parse/refscalar_test.go`: after `sub sref (\$);`, `sref(1)`, `sref("s")` and `sref(f())` are refused; `sref($x)`, `sref($h{k})`, `sref($x = 7)` are not
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestRefScalarSlotRefusesConstants$' -count=1 -v`
- [ ] Implement in the call path with a new refusal code; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestRefScalarSlotRefusesAggregatesAndArity`: `sref(@a)`, `sref(%h)`, `sref()` and `sref($x, $y)` are refused; `sref(my $z)` and `sref(substr($x, 0, 1))` are not
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestRefScalarSlotRefusesAggregatesAndArity$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestRefScalarSlotFromModuleSource`: a `.pm` declaring `sub sref (\$)` and a caller passing `1` is refused exactly as the `.pmt` case is
- [ ] Implement any fix; run test to verify pass; `make test`; commit
- [ ] Ratchet audit: run `TestParseRatchet`, `TestPerlGitTRatchet` and `TestCorpusRatchet`; for every file whose count moves, run `perl -c` on it; update the ratchet only for moves that match perl, and fix the refusal for any that do not; `make test`; commit

### Negative Scenarios

- scalar lvalues perl accepts stay accepted (measured): `sref($x)`, `sref($h{k})`, `sref($a[0])`, `sref(f()->[0])`, `sref($x = 7)`, `sref(my $z)` and `sref(substr($x, 0, 1))` (`go test ./internal/parse/ -run '^TestRefScalarSlotKeepsLvalues$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotKeepsLvalues'`)
- a sub with no prototype is not refused for a constant argument: `sub g {} g(1)` keeps perl's parse (`go test ./internal/parse/ -run '^TestRefScalarSlotLeavesUnprototypedAlone$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotLeavesUnprototypedAlone'`)
- the `\$` refusal has its own code, distinct from every existing refusal site in `RefusalSites`, so a corpus file can tell it from any other refusal (`go test ./internal/parse/ -run '^TestRefScalarSlotRefusalCodeDistinct$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotRefusalCodeDistinct'`)

## Acceptance criteria

- [ ] a constant or sub result passed to a \$ slot is refused, as perl refuses it (`go test ./internal/parse/ -run '^TestRefScalarSlotRefusesConstants$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotRefusesConstants'`)
- [ ] a \$ slot refuses aggregates and wrong arity as perl does (`go test ./internal/parse/ -run '^TestRefScalarSlotRefusesAggregatesAndArity$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotRefusesAggregatesAndArity'`)
- [ ] a \$ prototype from module source is enforced as a .pmt one is (`go test ./internal/parse/ -run '^TestRefScalarSlotFromModuleSource$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotFromModuleSource'`)
- [ ] scalar lvalues perl accepts in a \$ slot are not refused (`go test ./internal/parse/ -run '^TestRefScalarSlotKeepsLvalues$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotKeepsLvalues'`)
- [ ] an unprototyped sub taking a constant keeps perl parse (`go test ./internal/parse/ -run '^TestRefScalarSlotLeavesUnprototypedAlone$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotLeavesUnprototypedAlone'`)
- [ ] the \$ slot refusal has its own code (`go test ./internal/parse/ -run '^TestRefScalarSlotRefusalCodeDistinct$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotRefusalCodeDistinct'`)
- [ ] the T1 ratchet holds after the audit (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [ ] the perl.git t/ ratchet holds after the audit (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] the conformance corpus ratchet holds after the audit (`go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1 -v | grep -q '^--- PASS: TestCorpusRatchet'`)
