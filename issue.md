---
title: exec's and system's program slot is optional in perl, but CORE.pmt declares it required
state: done
urgency: normal
milestone: type-libraries
created: 2026-10-06T08:40:10.661906883Z
updated: 2026-10-07T21:31:44.556572411Z
sessions:
- start_sha: 951d595808be5e3c54d345b83ee638a24fd588cf
  end_sha: 951d595808be5e3c54d345b83ee638a24fd588cf
  commits: 0
  started_at: 2026-10-07T21:31:44.437146462Z
  ended_at: 2026-10-07T21:31:44.556572411Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-07T21:31:44.437146462Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-07T21:31:44.556572411Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt` (exec, system), `internal/types/declared.go` (`Signature.Invocant`), `docs/rfcs/0001-type-libraries.md` ("A leading slot with no comma is written with the invocant colon")
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

From 01a0fd9b-49b2-7ea9-bfa1-ba398baadc3f. CORE.pmt declares exec and system as the RFC writes them, `(Str $program: List[Str] @args)`, which makes the program slot REQUIRED. Measured on 5.42, it is optional: `exec "ls";`, a bare `exec;` and `system "ls";` all compile, with the program taken from the list; only `exec {"/bin/ls"} "ls";` fills the slot. print's slot is optional through its default (`FileHandle $fh = select()`), but exec's has no default value to write, and typed Perl has no spelling for "optional, no default". Nothing reads `Required` on an invocant yet, so nothing is wrong today.

The same gap for an ordinary parameter, from 01a10177-4c46-7807-b01c-b90e4c645a5b: perl chooses srand's seed itself when it gets none, and no expression states that -- `undef` gives "Integer overflow in srand" and seed UV_MAX, `time()` is not what perl uses. CORE.pmt has `sub srand (Int $seed = srand()) Str;`, a circular default naming the call that makes the choice. One spelling for "optional, no default" should serve both an invocant slot (exec, system) and a parameter (srand).

## Steps

- [ ] perigrin rules on the spelling of an optional slot with no default (e.g. `Str $program = undef:`, or the RFC states an invocant slot is always optional)
- [ ] Record it in the RFC; write a failing test that exec's and system's invocant is optional; implement; `make test`; commit

## Acceptance criteria

- [ ] exec's and system's program slot is optional (`go test ./internal/parse/ -run '^TestCoreExecSlotIsOptional$' -count=1 -v | grep -q '^--- PASS: TestCoreExecSlotIsOptional'`)
