---
title: Parse @{ foo { ... } } as perl does after an &-prototype sub
state: done
urgency: normal
milestone: m3-oracle-at-scale
blocks:
- 01a16602-5f77-78ef-a0c7-538e0a241cdc
created: 2026-10-24T13:36:55.397928747Z
updated: 2026-10-24T15:11:53.255155208Z
sessions:
- start_sha: 1d4a6f4e304f58c864263cc69e31979df173b738
  end_sha: ee411f22474a735637b49cfcdbf98c2fd7a8ffcf
  commits: 2
  started_at: 2026-10-24T13:38:04.003147709Z
  ended_at: 2026-10-24T15:11:53.255155208Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-24T13:38:04.003147709Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-24T15:11:53.255155208Z
observed_paths:
- .github/workflows/parse-fidelity.yml
- internal/parse/arrowhashread_test.go
- internal/parse/canon.go
- internal/parse/derefbarename_test.go
- internal/parse/fidelity.go
- internal/parse/fidelity_test.go
- internal/parse/parse.go
- internal/parse/sites.go
- internal/parse/term.go
- internal/parse/testdata/m3probes/deref-bare-name.txt
---

perigrin's ruling (2026-10-24): psc misreads a `{` as a code block, so a rule in Faithful cannot catch it. The fix belongs in the parser, and this issue fixes it for the one M3-dir file that has it.

## Context

- paths: `internal/parse/` (dereference braces, and the `&`-prototype block-argument path), `internal/parse/testdata/m3probes/`
- entrypoints: parsing `@{ ... }` whose contents begin with a bareword followed by `{`

uni/variables.t contains `@{ foo { "a" } }`, where `foo` is a sub with a `(&)` prototype. psc parses this as `foo` called with a block argument. Canon then writes `@{foo {"a";}}`, which perl rejects with `syntax error ... near ";}"`. perl compiles the source.

The narrowing of the last-semicolon forgiveness (01a16c1b-a587) cannot see this, because the tree itself says Block. So uni/variables.t counts as faithful while carrying a wrong tree.

**Measure first, with perl 5.44.0:**
- Run `-MO=Concise` and `-MO=Deparse,-p` on the exact line from uni/variables.t.
- Run the same reduction with and without `sub foo (&)`, with `use strict` and without it, and with a non-`&` prototype.
- One plausible reading is the hash slice `@foo{"a"}`, where toke.c treats a bareword followed by `{` inside deref braces as a variable name. Confirm or refute that before writing code.
- Make the parser produce perl's tree, and quote Concise in the test.

## Acceptance criteria

### Positive Scenarios

- [ ] The uni/variables.t shape gives perl's tree, held to Concise (`go test ./internal/parse/ -run '^TestDerefBareNameBraceReadsAsPerlDoes$' -count=1 -v | grep -q '^--- PASS'`)
- [ ] Probe rows for the reductions, with canon compiling under perl (`go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'`)
