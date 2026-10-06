---
title: ShapeOf reads a ($;) prototype as named unary; perl makes a user sub with it a list operator
state: in-progress
urgency: normal
milestone: v0.1
created: 2026-10-06T07:22:17.984195552Z
updated: 2026-10-06T20:29:35.68573568Z
sessions:
- start_sha: 99fed916148c08d826249cc82a99de5080a45746
  end_sha: ""
  commits: 0
  started_at: 2026-10-06T20:29:35.68573568Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-06T20:29:35.68573568Z
---

## Context

- paths: `internal/parse/shape.go` (`ShapeOf`), the parse ratchets
- commands: `make test`; `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet)$' -count=1`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found by the alignment gate on 01a0fd9b-46c4-7a8c-b8d6-498c5182a47f. `ShapeOf` reads a prototype of `$;` as a named unary, but perl 5.42 makes a USER sub with prototype `($;)` a list operator: `sub f($;) {...} (f 1, 2)` dies "Too many arguments for main::f", while `sub g($) {...} (g 1, 2)` reads `g(1), 2`. (Builtins with `$;`, such as `not`, may differ -- measure them separately.) So pvm parses `f 1, 2` under `($;)` as `f(1), 2`: a wrong tree for ordinary .pm/.pl source.

## Steps

- [ ] Measure user subs with `($;)`, `(;$)`, `($;$)`, `(;)` and builtins with `$;` (`not`, `getprotobynumber`) under perl 5.42 (`-MO=Deparse,-p` and the "Too many arguments" check)
- [ ] Write a failing parse test `TestShapeOfTrailingSemicolonIsListOp` for the user-sub case; fix ShapeOf; ratchets must not regress (report movement); `make test`; commit

## Acceptance criteria

- [ ] a user sub prototyped ($;) parses as a list operator (`go test ./internal/parse/ -run '^TestShapeOfTrailingSemicolonIsListOp$' -count=1 -v | grep -q '^--- PASS: TestShapeOfTrailingSemicolonIsListOp'`)
