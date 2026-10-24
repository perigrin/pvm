---
title: A trailing comma before /= re-emits as ;;
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-24T02:36:35.542533867Z
updated: 2026-10-24T14:40:49.713725037Z
---

Found by the independent review of NL-b (01a1692f-65de, exec/nlb 3630cbed) on 2026-10-24. The defect predates NL-b.

## Context

- paths: `internal/parse/canon.go`, `internal/parse/call.go`
- entrypoints: a list op whose trailing comma comes before an assignment operator that does not end the list

`substr $x, 0, 1, /= "d"` re-emits with `;;`, and `(foo 1, 2, /= 3)` re-emits as `);;`. NL-a measured that perl does not end the list before `/=`. It reports a syntax error or reads a term there. Find out whether psc refuses these sources, mis-trees them, or only miswrites them. Then make psc agree with perl: refuse what perl refuses, and keep the source's meaning otherwise.

## Acceptance criteria

### Positive Scenarios

- [ ] Probe rows for `substr $x, 0, 1, /= "d"` and `(foo 1, 2, /= 3)`, held to perl's verdict (`go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'`)
