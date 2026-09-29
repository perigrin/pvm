---
title: Blead named signature parameters `(:$x)` refuse
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T23:33:05.880220592Z
updated: 2026-09-29T23:33:05.880220592Z
---

Blead's named signature parameters refuse: `sub tnamed01 (:$alpha, :$beta)`, with defaults `:$x = "A"`, `:$y //= "B"`, after positionals and before a slurpy. The grammar is blead's perly.y, `sigscalarelem: optcolon PERLY_DOLLAR sigvar [ASSIGNOP optsigscalardefault]`, and perlsub's "Named parameters" (since 5.43.5).

No released perl has this syntax, and 5.42 rejects `sub f (:$x) {}`. So accepting it changes the parse of no program 5.42 accepts. perl.git's t/ is blead's own suite, and blead is the perl it is written for.

perl.git t/op/signatures.t:927-1005.

## Acceptance criteria

- [ ] `:$name` parses as a named parameter in a signature and nowhere else (`go test ./internal/parse/ -run '^TestNamedParameter$' -count=1 -v | grep -q '^--- PASS: TestNamedParameter'`)
