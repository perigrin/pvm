---
title: "\\my swallows the assignment that follows it"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-23T20:36:25.500878818Z
updated: 2026-10-24T14:40:49.307497055Z
---

Found in the independent review of 01a16602-5ebe (exec/5ebe, f80cbfbf), 2026-10-23, and measured with perl 5.44.0 -MO=Concise. The defect predates 5ebe.

## Context

- paths: `internal/parse/decl.go`, `internal/lower/ops.go`
- entrypoints: the `\` prefix over a declaration

`\my` swallows the assignment that follows it:

- `\my $x = \$y`: perl builds refassign. psc lowers it as `Ref(Decl Init)`.
- `\my ($p) = \$y`: perl builds aassign over srefgen. psc lowers it as `Ref(Decl ListAssign)`.
- `\my @a = []`: the same defect as the first case.

No issue covered this before. These lines need refaliasing, which few files use.

## Acceptance criteria

### Positive Scenarios

- [ ] Probe rows for the three snippets, held to perl (`go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'`)
