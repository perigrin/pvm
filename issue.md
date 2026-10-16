---
title: A statement-modifier early exit does not narrow what follows it
state: done
urgency: normal
milestone: v0.1
created: 2026-10-14T18:41:03.482257838Z
updated: 2026-10-16T03:03:43.645389428Z
sessions:
- start_sha: 9b435686d672051efee1ba667b9380fd8f320817
  end_sha: 9b435686d672051efee1ba667b9380fd8f320817
  commits: 0
  started_at: 2026-10-16T03:03:43.528359379Z
  ended_at: 2026-10-16T03:03:43.645389428Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T03:03:43.528359379Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T03:03:43.645389428Z
---

## Context

- paths: `internal/infer/lowered_flow.go` (guard narrowing; `narrowContinuation`)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

Found 2026-10-14 by the narrowing false-positive fixes, and noted earlier by the flow port (01a12a85-3100): a statement-modifier guard does not narrow what follows it. `return 0 unless defined $x; length($x)` -- after the early return, `$x` is defined (perl 5.42 -w is silent and prints 2), but PSC keeps its pre-guard type. The block form (`unless (defined $x) { return 0 }`) already narrows the continuation; the modifier form was left un-narrowed as legacy left it.

## Steps

- [ ] Failing test first; a `return`/`die`/`next`/`last` with an `unless`/`if` modifier narrows the rest of the block as the block form does; goldens under trailers; precision must not fall; `make test`; commit

## Acceptance criteria

- [ ] a modifier early exit narrows what follows (`go test ./internal/infer/ -run '^TestModifierGuardNarrowsContinuation$' -count=1 -v | grep -q '^--- PASS: TestModifierGuardNarrowsContinuation'`)
