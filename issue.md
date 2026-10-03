---
title: "Refuse at parse time what perl refuses at compile time: constants in \\$ slots and arities no declaration accepts"
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-03T11:12:40.060535979Z
updated: 2026-10-03T11:13:48.13053555Z
---

## Prerequisites

- [ ] Derive prototypes from typed signatures and typed signatures from prototypes
- [ ] Parse multi sub declarations and select a candidate by arity and argument types
- [ ] Declare grep, map and sort as multis in CORE.pmt, with sort's invocant slot

## Context

- paths: `internal/parse/shape.go`, `internal/parse/call.go`, `internal/parse/refusal.go`, `internal/parse/sites.go`, `internal/parse/arityrefusal_test.go` (new), `internal/parse/testdata/t1.ratchet`, `internal/parse/testdata/perlgitt.ratchet`
- commands: `go test ./internal/parse/ -count=1`, `go test ./internal/conformance/ -count=1`
- commands: ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet)$' -count=1 -timeout 600s -v`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume; `perl -c -e '...'` for each case)
- entrypoints: `parseByShape` (`internal/parse/shape.go:268`), `subPrototype`/`lookupSub` (`shape.go:246`, `:162`), `parseCallArgs` (`internal/parse/call.go:699`), the refusal codes and their inventory `RefusalSites` (`internal/parse/refusal.go`), `internal/types/select.go` for candidate arity

RFC 0001 "The scalar container" and "Multi declarations" (perigrin, 2026-10-02): where perl refuses a call at compile time, the parser refuses it too, as an Unknown with its own refusal code. Two families, measured on 5.42 with `sub sref (\$) {}`:

- **A `\$` slot takes a scalar lvalue.** `sref(1)` and `sref("s")` die "Type of arg 1 to main::sref must be scalar (not constant item)"; `sref(f())` "(not subroutine entry)"; `sref(@a)` "(not private array)"; `sref(%h)` "(not private hash)"; `sref()` "Not enough arguments for main::sref"; `sref($x, $y)` "Too many arguments for main::sref". `sref($x)`, `sref($h{k})`, `sref($x = 7)`, `sref(my $z)` and `sref(substr($x, 0, 1))` compile. The property is lvalue-ness, not "unevaluated".
- **An arity no declaration accepts.** `select(1, 2)` and `select(1, 2, 3)` die "Not enough arguments for select system call", `select(1, 2, 3, 4, 5)` "Too many arguments"; `localtime(1, 2)` and `gmtime(1, 2)` "Too many arguments"; `each()`, `grep()`, `map()` and `push()` "Not enough arguments". `chomp()`, `localtime()`, `push(@a)` and `grep { 1 } ()` compile.

Both apply to prototypes read from module source (`sub sref (\$)` in a `.pm`) as well as to `.pmt` declarations and CORE.pmt's candidates (the arity a candidate set accepts comes from `internal/types/select.go`). A call perl compiles is never refused: a sub with no prototype, or a library dispatching at run time, keeps perl's parse, and `psc check` reports what the declarations rule out (the call-site issue). Refusing more can move ratchet files; every file that moves is checked against `perl -c`, and the ratchet is updated only for moves that match perl. Two measured cases are not expressible from the declarations and need perigrin's ruling: `sort ()` and bare `sort` die "Not enough arguments for sort" while `sort(())` and `sort @empty` compile, though `sort`'s `(List @list)` candidate accepts zero arguments; and `select()` compiles though neither RFC candidate takes zero arguments. Re-marks no RFC section.

## Steps

- [ ] Measure each case with `perl -c`, then write failing test `TestRefScalarSlotRefusesConstants` in `internal/parse/arityrefusal_test.go`: after `sub sref (\$);`, `sref(1)`, `sref("s")` and `sref(f())` are refused; `sref($x)`, `sref($h{k})`, `sref($x = 7)` are not
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestRefScalarSlotRefusesConstants$' -count=1 -v`
- [ ] Implement in the call path with a new refusal code; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestRefScalarSlotRefusesAggregatesAndArity`: `sref(@a)`, `sref(%h)`, `sref()` and `sref($x, $y)` are refused; `sref(my $z)` and `sref(substr($x, 0, 1))` are not
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestBuiltinArityPerlRefusesIsRefused`: `select(1, 2)`, `select(1, 2, 3, 4, 5)`, `localtime(1, 2)`, `gmtime(1, 2)`, `each()`, `grep()`, `map()` and `push()` are refused
- [ ] Implement from the candidates' arity; run test to verify pass; `make test`; commit
- [ ] Ratchet audit: run `TestParseRatchet`, `TestPerlGitTRatchet` and `go test ./internal/conformance/ -count=1`; for every file whose count moves, run `perl -c` on it; update the ratchet only for moves that match perl, and fix the refusal for any that do not; `make test`; commit

### Negative Scenarios

- calls perl compiles stay accepted (measured): `chomp()`, `localtime()`, `push(@a)`, `grep { 1 } ()`, `sort(())`, `my @e; sort @e`, and `sref(f()->[0])` (`go test ./internal/parse/ -run '^TestArityRefusalKeepsWhatPerlAccepts$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalKeepsWhatPerlAccepts'`)
- a library sub with no prototype, declared in a `.pmt` as a multi whose candidates do not accept the call, is not refused by the parser: perl compiles it, so the parse stays perl's (`go test ./internal/parse/ -run '^TestArityRefusalLeavesUnprototypedCallsAlone$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalLeavesUnprototypedCallsAlone'`)
- the same `\$` refusal applies to a prototype read from module source: a `.pm` declaring `sub sref (\$)` and a caller passing `1` is refused exactly as the `.pmt` case is (`go test ./internal/parse/ -run '^TestRefScalarSlotFromModuleSource$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotFromModuleSource'`)
- each refusal names its own code, distinct from every other refusal site, so a corpus file can tell a constant in a `\$` slot from an arity refusal (`go test ./internal/parse/ -run '^TestArityRefusalCodesDistinct$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalCodesDistinct'`)
- `sort ()` and bare `sort`, which perl refuses though `(List @list)` accepts zero arguments, and `select()`, which perl compiles though no RFC candidate takes zero arguments: needs perigrin's ruling before it becomes a criterion (`go test ./internal/parse/ -run '^TestSortEmptyAndSelectEmptyFollowPerl$' -count=1 -v | grep -q '^--- PASS: TestSortEmptyAndSelectEmptyFollowPerl'`)

## Acceptance criteria

- [ ] a constant or sub result passed to a \$ slot is refused, as perl refuses it (`go test ./internal/parse/ -run '^TestRefScalarSlotRefusesConstants$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotRefusesConstants'`)
- [ ] a \$ slot refuses aggregates and wrong arity as perl does (`go test ./internal/parse/ -run '^TestRefScalarSlotRefusesAggregatesAndArity$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotRefusesAggregatesAndArity'`)
- [ ] a builtin call whose arity perl refuses is refused (`go test ./internal/parse/ -run '^TestBuiltinArityPerlRefusesIsRefused$' -count=1 -v | grep -q '^--- PASS: TestBuiltinArityPerlRefusesIsRefused'`)
- [ ] calls perl compiles are not refused (`go test ./internal/parse/ -run '^TestArityRefusalKeepsWhatPerlAccepts$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalKeepsWhatPerlAccepts'`)
- [ ] unprototyped library calls keep the parse perl gives them (`go test ./internal/parse/ -run '^TestArityRefusalLeavesUnprototypedCallsAlone$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalLeavesUnprototypedCallsAlone'`)
- [ ] a \$ prototype from module source is enforced as a .pmt one is (`go test ./internal/parse/ -run '^TestRefScalarSlotFromModuleSource$' -count=1 -v | grep -q '^--- PASS: TestRefScalarSlotFromModuleSource'`)
- [ ] each refusal has its own code (`go test ./internal/parse/ -run '^TestArityRefusalCodesDistinct$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalCodesDistinct'`)
- [ ] the T1 ratchet holds after the audit (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -timeout 600s -v | grep -q '^--- PASS: TestParseRatchet'`)
- [ ] the perl.git t/ ratchet holds after the audit (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -timeout 600s -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] the conformance corpus passes (`go test ./internal/conformance/ -count=1`)
