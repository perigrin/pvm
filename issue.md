---
title: Refuse at parse time the call arities perl refuses at compile time
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4a5b-7e94-ad87-756dfcfe1f4c
- 01a0fd9b-519d-7b42-886e-2222a3076986
- 01a10dc1-3530-762f-9a03-177445629c6b
- 01a10c6e-988f-7aa6-811d-0e462abc442a
blocks:
- 01a0fd9b-5700-7514-a143-2caed2d06d8b
created: 2026-10-03T11:12:40.060535979Z
updated: 2026-10-05T20:31:10.668390467Z
---

## Prerequisites

- [ ] Parse multi sub declarations and select a candidate by arity and argument types
- [ ] Declare grep, map and sort as multis in CORE.pmt, with sort's invocant slot
- [ ] Type CORE.pmt's process, time, object and miscellaneous builtins, with localtime and gmtime as :context multis
- [ ] Refuse constants and sub results in \$ slots

## Context

- paths: `internal/parse/shape.go`, `internal/parse/call.go`, `internal/parse/refusal.go`, `internal/parse/sites.go`, `internal/parse/arityrefusal_test.go` (new), `internal/parse/testdata/t1.ratchet`, `internal/parse/testdata/perlgitt.ratchet`, `internal/conformance/testdata/corpus.ratchet`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: ratchets, one named test each (measured on this tree, 2026-10-05, with compile: `TestParseRatchet` 8.2s, `TestPerlGitTRatchet` 7.0s, `TestCorpusRatchet` 7.9s; `git-zhi verify` kills a command after 300s): `go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v`, `go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1 -v`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume; `perl -c -e '...'` for each case)
- entrypoints: `parseByShape` (`internal/parse/shape.go:268`), `subPrototype`/`lookupSub` (`shape.go:246`, `:162`), `parseCallArgs` (`internal/parse/call.go:699`), the refusal codes and their inventory `RefusalSites` (`internal/parse/refusal.go`), the `\$`-slot refusal code the previous refusal issue added, `internal/types/select.go` for candidate arity

RFC 0001 "Refusing what perl refuses" (perigrin, 2026-10-02), its second family: a call whose arity no declaration accepts. Where perl refuses it at compile time the parser refuses it too, as an Unknown with its own refusal code. Measured on 5.42: `select(1, 2)` and `select(1, 2, 3)` die "Not enough arguments for select system call", `select(1, 2, 3, 4, 5)` "Too many arguments"; `localtime(1, 2)` and `gmtime(1, 2)` "Too many arguments" (their candidates are the process-and-time typing batch's, which this issue follows); `each()`, `grep()`, `map()` and `push()` "Not enough arguments". `chomp()`, `localtime()`, `select()`, `push(@a)` and `grep { 1 } ()` compile (`select()` matches `select`'s `() Str` candidate, perigrin 2026-10-05). The first family, a constant, sub result or aggregate in a `\$` slot, is the `\$`-slot refusal issue's, which this one follows. The rule applies to prototypes read from module source as well as to `.pmt` declarations and CORE.pmt's candidates (the arity a candidate set accepts comes from `internal/types/select.go`). A call perl compiles is never refused: a sub with no prototype, or a library dispatching at run time, keeps perl's parse, and `psc check` reports what the declarations rule out (the call-site issue). Refusing more can move ratchet files; every file that moves is checked against `perl -c`, and the ratchet is updated only for moves that match perl. RFC 0001 "A required argument defaults to `die`" (perigrin, 2026-10-05) settles `sort`: `sort ()` and bare `sort` die "Not enough arguments for sort" while `sort(())` and `sort @empty` compile, and CORE.pmt's plain candidate `multi sub sort (List @list = die) List;` says exactly that. The refusal counts a parameter whose default is `die` as required (recorded on `types.Param` by the typed-parameters issue), so `sort()` is refused by the same arity rule as `grep()`, with no sort special case; a `List` with no default still accepts zero arguments. With the `\$` family already built, this completes the section's rule, so it re-marks "Refusing what perl refuses".

## Steps

- [ ] Measure each case with `perl -c`, then write failing test `TestBuiltinArityPerlRefusesIsRefused` in `internal/parse/arityrefusal_test.go`: `select(1, 2)`, `select(1, 2, 3, 4, 5)`, `localtime(1, 2)`, `gmtime(1, 2)`, `each()`, `grep()`, `map()` and `push()` are refused
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestBuiltinArityPerlRefusesIsRefused$' -count=1 -v`
- [ ] Implement from the candidates' arity with a new refusal code; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestArityRefusalDieDefaultIsRequired`: `sort()` and bare `sort` are refused with the same arity refusal code as `grep()`, because the plain candidate's `@list` defaults to `die`; `sort(())` and `my @e; sort @e` are not refused; the refusal path reads requiredness from the declaration and names no builtin
- [ ] Run test to verify failure, implement, run test to verify pass: `go test ./internal/parse/ -run '^TestArityRefusalDieDefaultIsRequired$' -count=1 -v`; `make test`; commit
- [ ] Ratchet audit: run `TestParseRatchet`, `TestPerlGitTRatchet` and `TestCorpusRatchet`; for every file whose count moves, run `perl -c` on it; update the ratchet only for moves that match perl, and fix the refusal for any that do not; `make test`; commit
- [ ] Re-mark the RFC heading "Refusing what perl refuses" `(*Implemented*, <sha>)`; commit

### Negative Scenarios

- calls perl compiles stay accepted (measured): `chomp()`, `localtime()`, `select()`, `push(@a)`, `grep { 1 } ()`, `sort(())` and `my @e; sort @e` (`go test ./internal/parse/ -run '^TestArityRefusalKeepsWhatPerlAccepts$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalKeepsWhatPerlAccepts'`)
- a library sub with no prototype, declared in a `.pmt` as a multi whose candidates do not accept the call, is not refused by the parser: perl compiles it, so the parse stays perl's (`go test ./internal/parse/ -run '^TestArityRefusalLeavesUnprototypedCallsAlone$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalLeavesUnprototypedCallsAlone'`)
- each refusal names its own code, distinct from every other refusal site, so a corpus file can tell a constant in a `\$` slot from an arity refusal (`go test ./internal/parse/ -run '^TestArityRefusalCodesDistinct$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalCodesDistinct'`)
- a parameter whose default is `die` is required (RFC 0001 "A required argument defaults to `die`"): `sort()` and bare `sort` are refused through `(List @list = die)`, as perl refuses them at compile time, while `sort(())` and `sort @e` are not; no sort special case (`go test ./internal/parse/ -run '^TestArityRefusalDieDefaultIsRequired$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalDieDefaultIsRequired'`)

## Acceptance criteria

- [ ] a builtin call whose arity perl refuses is refused (`go test ./internal/parse/ -run '^TestBuiltinArityPerlRefusesIsRefused$' -count=1 -v | grep -q '^--- PASS: TestBuiltinArityPerlRefusesIsRefused'`)
- [ ] calls perl compiles are not refused (`go test ./internal/parse/ -run '^TestArityRefusalKeepsWhatPerlAccepts$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalKeepsWhatPerlAccepts'`)
- [ ] unprototyped library calls keep the parse perl gives them (`go test ./internal/parse/ -run '^TestArityRefusalLeavesUnprototypedCallsAlone$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalLeavesUnprototypedCallsAlone'`)
- [ ] a parameter whose default is `die` is required, so `sort()` is refused and `sort(())` is not, with no sort special case (`go test ./internal/parse/ -run '^TestArityRefusalDieDefaultIsRequired$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalDieDefaultIsRequired'`)
- [ ] each refusal has its own code (`go test ./internal/parse/ -run '^TestArityRefusalCodesDistinct$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalCodesDistinct'`)
- [ ] the T1 ratchet holds after the audit (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [ ] the perl.git t/ ratchet holds after the audit (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] the conformance corpus ratchet holds after the audit (`go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1 -v | grep -q '^--- PASS: TestCorpusRatchet'`)
- [ ] the RFC section on refusing what perl refuses is implemented (`grep -q '^### Refusing what perl refuses (\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
