---
title: Type CORE.pmt's system, user, network, IPC, socket and control-flow builtins, and hold every prototyped line typed
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a10177-4c46-7807-b01c-b90e4c645a5b
- 01a10177-4e7b-7d2c-8d2a-2033647ffc4a
- 01a10177-5057-7f6f-bfa1-363a4568ae80
- 01a10c6e-988f-7aa6-811d-0e462abc442a
created: 2026-10-05T14:18:57.557243435Z
updated: 2026-10-06T17:06:23.488051434Z
sessions:
- start_sha: bbcf54860c14ed10b826ded5ec4531d0116a534f
  end_sha: bbcf54860c14ed10b826ded5ec4531d0116a534f
  commits: 0
  started_at: 2026-10-06T17:06:23.348846391Z
  ended_at: 2026-10-06T17:06:23.488051434Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-06T17:06:23.348846391Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-06T17:06:23.488051434Z
---

## Prerequisites

- [ ] Type CORE.pmt's string, numeric and regex builtins
- [ ] Type CORE.pmt's array, list and hash builtins
- [ ] Type CORE.pmt's I/O, file and directory builtins
- [ ] Type CORE.pmt's process, time, object and miscellaneous builtins, with localtime and gmtime as :context multis

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/coretypes_test.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `TestCoreDerivedPrototypesArePerls` (`internal/parse/coretypes_test.go`, the harness); `perl -MPod::Functions -e 'print "@{$Kinds{User}} @{$Kinds{Network}} @{$Kinds{Socket}} @{$Kinds{SysV}} @{$Kinds{Flow}}"'`

Last batch, by perlfunc's categories "Fetching user and group info", "Fetching network info", "Low-level socket functions", "System V interprocess communication functions" and "Keywords related to the control flow of your Perl program" (Pod::Functions kinds `User`, `Network`, `Socket`, `SysV`, `Flow`), intersected with CORE.pmt's prototyped lines and without names earlier batches hold. Measured on 5.42 that is 65 lines: `endgrent`, `endhostent`, `endnetent`, `endpwent`, `getgrent`, `getgrgid`, `getgrnam`, `getlogin`, `getpwent`, `getpwnam`, `getpwuid`, `setgrent`, `setpwent`; `endprotoent`, `endservent`, `gethostbyaddr`, `gethostbyname`, `gethostent`, `getnetbyaddr`, `getnetbyname`, `getnetent`, `getprotobyname`, `getprotobynumber`, `getprotoent`, `getservbyname`, `getservbyport`, `getservent`, `sethostent`, `setnetent`, `setprotoent`, `setservent`; `accept`, `bind`, `connect`, `getpeername`, `getsockname`, `getsockopt`, `listen`, `recv`, `send`, `setsockopt`, `shutdown`, `socket`, `socketpair`; `msgctl`, `msgget`, `msgrcv`, `msgsnd`, `semctl`, `semget`, `semop`, `shmctl`, `shmget`, `shmread`, `shmwrite`; `__FILE__`, `__LINE__`, `__PACKAGE__`, `__SUB__`, `break`, `continue`, `dump`, `evalbytes`, `exit`, `wantarray`. Every type is measured and recorded as a comment. With this batch every one of the 188 prototyped lines is typed except the six that hold a glob slot -- `lock`, `undef`, `tie`, `tied`, `untie` and `pos` -- which stay prototype-only until RFC open question 11 is settled (perigrin, 2026-10-03); the every-line check names exactly those six as its exemptions, so it can pass now and fails if any other line is left untyped or any of the six is typed by a guess. The same check enforces the directive of RFC "The scalar container" that those six builtins keep prototype-only lines, untyped; the aliased-parameter issue re-marks that section earlier, since the six lines are prototype-only already, and this check is what holds them so. This finishes the RFC's decided content for the 188 (the derivation, the aliased rows including `Code \&c`, the final-`List` rule and the multi-candidate rule came first; the glob slot is an open question, not part of the section), so it completes RFC "A typed signature and a prototype say the same thing" and re-marks it.

## Steps

- [ ] Measure the batch's names; write failing test `TestCoreSystemFlowBuiltinsTyped`: every name in the batch has a typed CORE.pmt line
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreSystemFlowBuiltinsTyped$' -count=1 -v`
- [ ] Type the lines, category by category, with measurement comments; run the harness after each category; `make test`; commit per category
- [ ] Write failing test `TestCoreEveryPrototypedLineIsTyped`: every one of the 188 prototyped lines carries a typed signature except exactly `lock`, `undef`, `tie`, `tied`, `untie` and `pos`, named in the test, which carry none; and the file reads with no declaration errors (unknown names, bare slurpies)
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreEveryPrototypedLineIsTyped$' -count=1 -v`
- [ ] Type whatever the check finds untyped, measuring each; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC heading "A typed signature and a prototype say the same thing" `(*Implemented*, <sha>)`; commit

### Negative Scenarios

- no CORE.pmt line is left half-done: every prototyped line but the six named glob-slot lines carries a typed signature, and the file reads with no declaration errors (`go test ./internal/parse/ -run '^TestCoreEveryPrototypedLineIsTyped$' -count=1 -v | grep -q '^--- PASS: TestCoreEveryPrototypedLineIsTyped'`)
- the exemption list is exact, not a loophole: the same check over a copy of CORE.pmt with one non-glob line left untyped (`sub socket :prototype(*$$$);`) reports it, and over a copy with `pos` typed reports that too (`go test ./internal/parse/ -run '^TestCoreEveryLineCheckExemptsOnlyGlobSlots$' -count=1 -v | grep -q '^--- PASS: TestCoreEveryLineCheckExemptsOnlyGlobSlots'`)
- a niladic control-flow keyword (`__FILE__`, `wantarray`, `break`: prototype `()`) is typed with zero parameters, so its derived prototype stays `()` (`go test ./internal/parse/ -run '^TestCoreNiladicFlowKeywordsTypedEmpty$' -count=1 -v | grep -q '^--- PASS: TestCoreNiladicFlowKeywordsTypedEmpty'`)

## Acceptance criteria

- [ ] CORE.pmt types every user, network, socket, IPC and control-flow builtin (`go test ./internal/parse/ -run '^TestCoreSystemFlowBuiltinsTyped$' -count=1 -v | grep -q '^--- PASS: TestCoreSystemFlowBuiltinsTyped'`)
- [ ] every prototyped CORE.pmt line but the six glob-slot lines is typed and the file has no declaration errors (`go test ./internal/parse/ -run '^TestCoreEveryPrototypedLineIsTyped$' -count=1 -v | grep -q '^--- PASS: TestCoreEveryPrototypedLineIsTyped'`)
- [ ] every prototype derived from CORE.pmt types is perl's (`go test ./internal/parse/ -run '^TestCoreDerivedPrototypesArePerls$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedPrototypesArePerls'`)
- [ ] the every-line check exempts exactly the glob-slot lines (`go test ./internal/parse/ -run '^TestCoreEveryLineCheckExemptsOnlyGlobSlots$' -count=1 -v | grep -q '^--- PASS: TestCoreEveryLineCheckExemptsOnlyGlobSlots'`)
- [ ] niladic control-flow keywords derive an empty prototype (`go test ./internal/parse/ -run '^TestCoreNiladicFlowKeywordsTypedEmpty$' -count=1 -v | grep -q '^--- PASS: TestCoreNiladicFlowKeywordsTypedEmpty'`)
- [ ] the RFC section on typed signatures and prototypes is implemented (`grep -q '^### A typed signature and a prototype.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
