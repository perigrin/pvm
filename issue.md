---
title: Decide what a parenless call to an unknown callee does with its arguments
state: pending
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0c13f-6816-79d2-945a-b84a12d39e6c
created: 2026-09-20T23:55:47.309990102Z
updated: 2026-09-21T03:31:37.104043968Z
---

**Rewritten 2026-09-21. This was filed as a three-way design choice for
perigrin. It is not a choice — perl answers it, and the answer rules out two
of the three routes.**

## What perl does, measured on 5.42.0

    $ perl -MO=Deparse -e 'ok 8;'
    Number found where operator expected (Do you need to predeclare "ok"?)
    syntax error near "ok 8"

    $ perl -MO=Deparse -e 'ok 8, 9;'
    syntax error, same shape

    $ perl -MO=Deparse -e 'ok $x, 9;'
    $x->ok, '???'                    <-- INDIRECT OBJECT, not a call at all

    $ perl -MO=Deparse -e 'sub ok {} ok 8, 9;'
    sub ok { } ok 8, 9;              <-- declared first: a call

    $ perl -MO=Deparse -e 'sub f {} f 13 % 4;'
    sub f { } f 1;                   <-- arguments consumed, 13 % 4 folded

So: **perl refuses to compile a parenless call to an undeclared sub with a
numeric argument.** It is not a call perl parses and we fail to; it is a
syntax error in perl too.

Three consequences, each ruling something out:

**Route A (consume greedily) is WRONG, not merely risky.** Building a tree
for `ok 8` where perl builds none is a tree for a program perl rejects --
WRONG in the gate's sense, and M1's gate is Oracle WRONG = 0.

**The status quo is closer to correct than this issue assumed.** For a
genuinely undeclared callee, Unknown IS the honest answer. §4.8.3's
`Call{Resolved:false}` and an Unknown statement around it both say "I cannot
settle this", and perl cannot either.

**Route C is not a heuristic. It is perl's rule.** "Declared before the call
site" is exactly what `knowsShape()` models (call.go:152), what
`comp/proto.t:566` documents in the corpus -- "prototype checking done if
sub pre-declared" -- and what 01a0ad3f already built for imports.

## So what is the actual bug

Not that we decline `ok 8`. That `ok` in T2 **is** declared -- `Test::More`
exports it -- and we are not using that. The T1 measurement from 01a0ad3f
says the machinery works when the module is reachable: 460 -> 532 clean
files, 3,322 -> 2,531 Unknown nodes, when Test::More resolved.

The work is to make the same resolution reach T2, and to check the in-file
`sub NAME` case, which needs no module at all.

## Scope

1. Measure how many of T2's ~330 parenless-call Unknowns have a callee that
   IS declared -- by an import whose module is readable, or by a `sub NAME`
   earlier in the same file. That number is the size of the real fix.
2. The remainder are genuinely undeclared. Confirm perl also rejects them,
   file-by-file, and they stay Unknown as correct behaviour.

## The fragments become corpus fixtures

The probes above are a discriminating pair the corpus does not have:
`sub ok {} ok 8, 9` versus bare `ok 8, 9` is precisely what the fidelity
harness exists to score, and nothing in T2 tests the declared/undeclared
boundary directly. Land them as oracle fixtures rather than throwing them
away with the scratch file.

## Acceptance

- [ ] The declared/undeclared boundary is a fixture perl and the parser are both scored on (`go test ./internal/parseoracle/ -run '^TestParenlessCallBoundary$' -count=1 -v | grep -q '^--- PASS: TestParenlessCallBoundary'`)
- [ ] `sub NAME` earlier in the same file makes a parenless call to it parse (`go test ./internal/parse/ -run '^TestInFileSubDeclarationIsKnown$' -count=1 -v | grep -q '^--- PASS: TestInFileSubDeclarationIsKnown'`)
- [ ] A genuinely undeclared parenless call stays Unknown, because perl rejects it too (`go test ./internal/parse/ -run '^TestUndeclaredParenlessCallStaysUnknown$' -count=1 -v | grep -q '^--- PASS: TestUndeclaredParenlessCallStaysUnknown'`)
- [ ] The T2 share whose callee is declared is measured and recorded here, before any fix (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)

Measured 2026-09-21. Plan: M1. Spec: §4.8.3, §3.2.
