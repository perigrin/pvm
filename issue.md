---
title: Statements, blocks and declarations
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-15T21:52:57.320252116Z
updated: 2026-09-15T21:53:22.438464197Z
---

Statements, blocks, and the declarations the T2 corpus is made of: `my`/`our`/
`local`, `sub`, `if`/`elsif`/`else`, `unless`, `while`, `until`, `for`,
`foreach`, statement modifiers, labels, and `package`.

Spec ch5 is the reference. This is the largest issue in the milestone by
volume and the least surprising by content — most of it is ordinary recursive
descent over the expression parser the previous issue built.

## The parts that are not ordinary

**A `{` is a block or a hash constructor**, and the lexer's expect state
already carries the distinction (XBLOCK versus XREF, M0's expect.go). Use it
rather than re-deriving: `map { $_ => 1 } @a` and `map { ; $_ => 1 } @a` mean
different things and perl decides with the same state this parser can read.

**Statement modifiers invert the tree.** `print if $x` is a conditional whose
body is a print, and the modifier binds looser than anything in the
expression grammar.

**A label is not an expression.** `LOOP: while (...)` needs the XSTATE
recognition M0 already does for POD, which is the same "could a statement
start here" question.

## Unknown is the fallback, everywhere

Any statement form this issue does not implement emits an Unknown node
spanning the whole statement, and parsing continues at the next `;` or `}`.
Recovery that resynchronises on a statement boundary is what keeps one
unfamiliar construct from cascading, and the cascade is exactly what
`op/filetest.t` showed the tree-sitter grammar doing: one bad line at 96
derailed a reference at 131.

## Acceptance

- [ ] The declaration forms parse: my, our, local, sub, package (`go test ./internal/ast/ -run '^TestDeclarations$' -count=1 -v | grep -q '^--- PASS: TestDeclarations'`)
- [ ] Conditionals and loops parse, including elsif chains (`go test ./internal/ast/ -run '^TestControlFlow$' -count=1 -v | grep -q '^--- PASS: TestControlFlow'`)
- [ ] Statement modifiers bind looser than any expression operator (`go test ./internal/ast/ -run '^TestStatementModifiers$' -count=1 -v | grep -q '^--- PASS: TestStatementModifiers'`)
- [ ] A `{` is a block or a hashref per the lexer's expect state (`go test ./internal/ast/ -run '^TestBlockVersusHashref$' -count=1 -v | grep -q '^--- PASS: TestBlockVersusHashref'`)
- [ ] A label before a loop parses as a label, not an expression (`go test ./internal/ast/ -run '^TestLoopLabels$' -count=1 -v | grep -q '^--- PASS: TestLoopLabels'`)
- [ ] An unknown statement becomes Unknown and parsing resumes at the next boundary (`go test ./internal/ast/ -run '^TestStatementRecovery$' -count=1 -v | grep -q '^--- PASS: TestStatementRecovery'`)

Plan: M1. Spec: ch5.
