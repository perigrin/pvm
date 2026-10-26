---
title: Loop-control, declaration and canon defects found by the M3 per-fix reviews
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-26T00:16:46.116463775Z
updated: 2026-10-26T00:16:46.116463775Z
---

Found by the per-fix reviews of the last M3 issues (6030, 608d, 60e9, 75e8, 61fe; PR #493) in the cloud session that closed M3, 2026-10-25. Every defect below predates those issues and shows up on snippets perl 5.44 accepts.

## Context

- paths: `internal/parse/` (loop control, declarations, the feature tracker), `internal/parse/canon.go`

- **Loop control:** `last when`, `last L, 1`, `next, print`, and `return last or 1`.
- **A last statement with no `;` before `use`:** `print "a\n"; use strict` re-emits wrongly when the file's last `;` is missing.
- **Declarations with attributes or types:** `\my $x :shared` and `\my Foo $x`.
- **A filehandle followed by parens:** `print STDOUT ("x\n")` re-emits as `STDOUT(`, which perl reads as a call to a sub named STDOUT. Faithful passes it because whitespace is not a compared token. See also the print-FH issue in v0.1.
- **Switch features:** psc does not scope them lexically.
- **`sub x {} goto x`:** psc misreads the goto operand.
- **Two heredocs before `__DATA__`:** their bodies and the data section interleave wrongly.

Measure each one with perl 5.44 (Concise and Deparse,-p) before fixing it.

## Acceptance criteria

### Positive Scenarios

- [ ] Each listed shape gives perl's tree and a faithful canon that compiles (`go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'`)
