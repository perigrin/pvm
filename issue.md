---
title: "[R2] psc check ignores no warnings scopes, reporting what perl has been told not to warn about"
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T15:49:27.346112746Z
updated: 2026-10-16T15:49:27.346112746Z
---

## Context

- found by: R1's re-sweep of perl 5.42's lib (2026-10-16): Test/Builder.pm:683 is under `no warnings qw/uninitialized numeric/`, and psc reports Undef in a Str slot there
- paths: `internal/parse/` (pragmas: `use warnings` / `no warnings` with category lists), `internal/lower/` (UseStmt), `internal/infer/` (operand and argument checks)
- commands: `go test ./internal/infer/`; perl 5.42 `-w`

perigrin ruled 2026-10-16: honour `no warnings` now, on PR #486. A lexical `no warnings` (all categories, or a list) turns off, for the rest of its enclosing block, the warnings perl would emit there; psc must skip the diagnostics that correspond to those categories:
- `uninitialized`: an Undef-including operand or argument (coercion/type mismatch whose only failing part is Undef)
- `numeric`: a Str (non-numeric) in a numeric slot
- `no warnings;` with no list: all of them
`use warnings` (or a nested block) restores them. perl 5.42: `my $u; { no warnings "uninitialized"; my $s = "x" . $u; } my $t = "y" . $u;` warns only on the second concatenation. A diagnostic perl reports as an error regardless of warnings (e.g. a ref used as a filehandle) is not affected.

## Steps

- [ ] Measure the category for each psc diagnostic kind against perl; failing test first; scope by lexical block; `make test`; commit

## Acceptance criteria

- [ ] a no warnings scope silences the diagnostics of its categories (`go test ./internal/infer/ -run '^TestNoWarningsScopeSilencesItsCategories$' -count=1 -v | grep -q '^--- PASS: TestNoWarningsScopeSilencesItsCategories'`)
