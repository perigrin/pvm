---
title: Control flow, and the label that is not an expression
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a75b-5e51-79ef-a59a-3a997f738cca
blocks:
- 01a0a75b-5e51-79f1-8575-3f30abd05ac9
created: 2026-09-15T23:16:18.129308407Z
updated: 2026-09-15T23:16:18.129308407Z
---

Control flow: `if`/`elsif`/`else`, `unless`, `while`, `until`, `for`,
`foreach` in both forms, and labels.

Ordinary recursive descent over the expression parser, with one non-obvious
piece.

## A label is not an expression

`LOOP: while (...)` — the `LOOP:` is a label, but `$h{LOOP}` and `LOOP ? 1 : 0`
put the same bareword-then-colon shape in expression position. The
distinction is "could a statement start here", which is the XState question
M0 already answers for POD recognition. Read that state rather than
re-deriving it.

## The C-style and list forms

`for (my $i = 0; $i < 10; $i++)` and `for my $x (@list)` share a keyword and
nothing else structurally. Both are in §0.13's corpus.

## Acceptance

- [ ] if/elsif/else chains parse, including an elsif with no else (`go test ./internal/parse/ -run '^TestConditionalChains$' -count=1 -v | grep -q '^--- PASS: TestConditionalChains'`)
- [ ] unless, while and until parse with the same block shape (`go test ./internal/parse/ -run '^TestNegatedAndLoopForms$' -count=1 -v | grep -q '^--- PASS: TestNegatedAndLoopForms'`)
- [ ] Both for forms parse: C-style three-part and list iteration (`go test ./internal/parse/ -run '^TestForLoopForms$' -count=1 -v | grep -q '^--- PASS: TestForLoopForms'`)
- [ ] A label before a loop is a label; `$h{LOOP}` and `LOOP ? 1 : 0` are not (`go test ./internal/parse/ -run '^TestLoopLabels$' -count=1 -v | grep -q '^--- PASS: TestLoopLabels'`)

Spec: ch5.
