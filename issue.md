---
title: Statement modifiers and the block-versus-hashref decision
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a75b-5e51-79f0-8af1-95c5aa378706
- 01a0a75b-7c7b-78ca-b47f-40cd34c5c963
blocks:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
created: 2026-09-15T23:16:18.129308407Z
updated: 2026-09-15T23:17:26.04689228Z
---

Statement modifiers, and the `{` that is either a block or a hash
constructor.

Last of the statement seams, and the one that depends on the keyword table:
deciding what a `{` opens sometimes needs to know what precedes it is `map`.

## Modifiers invert the tree

`print if $x` is a conditional whose body is a print. The modifier binds
looser than anything in the expression grammar — looser than `or`, which is
itself below the `and or not xor` cliff of §4.1.1.

## The `{` decision

Spec §4.9.2 is titled "The `{` decision" and gives it for `sort`/`map`/`grep`
specifically:

    map { $_ => 1 } @a      # hashref or block? perl guesses, and documents that it guesses
    map { ; $_ => 1 } @a    # the leading ; forces block

M0's `expect.go` declares XBlock and XRef for this and **never assigns
either** — `after()` returns only XOperator, XTerm and XState. So the state
this issue was originally written to "just read" does not exist yet. The
lexer-feedback issue adds the brace stack; this issue consumes it.

perl's own heuristic (`toke.c`, `intuit_curly`) peeks past the `{` at the
first token: a bareword or string followed by `,` or `=>` means hashref.
Match that heuristic, and where it is genuinely ambiguous emit Unknown rather
than pick — the plan's rule applies to this as much as to anything.

## Acceptance

- [ ] Modifier forms parse: if, unless, while, until, for, foreach (`go test ./internal/parse/ -run '^TestStatementModifierForms$' -count=1 -v | grep -q '^--- PASS: TestStatementModifierForms'`)
- [ ] A modifier binds looser than every expression operator, `or` included (`go test ./internal/parse/ -run '^TestModifierBindsLoosest$' -count=1 -v | grep -q '^--- PASS: TestModifierBindsLoosest'`)
- [ ] `map { $_ => 1 } @a` and `map { ; $_ => 1 } @a` parse differently (`go test ./internal/parse/ -run '^TestBlockVersusHashref$' -count=1 -v | grep -q '^--- PASS: TestBlockVersusHashref'`)
- [ ] The decision reads the lexer's XBlock/XRef state, not a local re-derivation (`go test ./internal/parse/ -run '^TestBraceDecisionUsesExpectState$' -count=1 -v | grep -q '^--- PASS: TestBraceDecisionUsesExpectState'`)

Spec: ch5, ch4 §4.9.2, §4.1.1. toke.c `intuit_curly`.
