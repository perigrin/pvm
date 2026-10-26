---
title: Canon wraps an arrow call that is followed by a subscript
state: in-progress
urgency: normal
milestone: v0.1
created: 2026-10-24T00:31:36.873798557Z
updated: 2026-10-26T00:16:33.810326439Z
sessions:
- start_sha: 3b77ae24bd593b2eb20e829fdc36fd5135584b9a
  end_sha: ""
  commits: 0
  started_at: 2026-10-26T00:16:33.810326439Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-26T00:16:33.810326439Z
---

Found by the independent review of NL-a (01a1690b-4745, exec/nla 43ddc197) on 2026-10-24. The defect predates NL-a.

## Context

- paths: `internal/parse/canon.go`
- entrypoints: canon's handling of `->(...)` followed by a subscript

Canon wraps an arrow call that has a subscript after it: `$f->(...)->[0]` comes out as `($f -> (...))->[0]`. The same happens to `$f->(1)->[0]` and `$f->($x)->{a}`. The meaning doesn't change, but Faithful reports these lines false. If any M3-dir file has this shape, 01a16602-5f77 lists it as a located defect.

## Acceptance criteria

### Positive Scenarios

- [ ] Faithful probe rows for `$f->(1)->[0]`, `$f->($x)->{a}` and `$f->(not $x)->[0]` (`go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'`)
