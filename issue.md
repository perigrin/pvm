---
title: The corpus has no parenless call with an argument, which is T2's largest bucket
state: pending
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T02:06:40.108445691Z
updated: 2026-09-26T02:06:40.108445691Z
---

The corpus's largest hole: no case is a parenless call with an argument,
which is T2 core's single largest failure bucket.

## Measured, at b0f1a9e9

Declaration form and prototype are IRRELEVANT. Argument count is
everything:

    sub answer { 42 }   print answer, "\n";   clean   <- corpus has this
    sub answer { 42 }   answer;               clean
    sub f { }           f(8);                 clean
    sub f { }           f 8;                  UNKNOWN
    sub f;              f 8;                  UNKNOWN
    sub f ($) { }       f 8;                  UNKNOWN
    sub tryeq ($$$$){}  tryeq 1, 13 % 4, 1;   UNKNOWN  <- arith.t, 179

The corpus's own call-form slice (`subroutines.md`, "Four spellings of
one call") passes because all four spellings call with NO arguments. The
fifth spelling is the one T2 needs.

`%` is not implicated: `f 13 % 4` and `f 8` fail identically, so
`01a0c10b`'s "`%` read as a hash sigil" is a symptom of argument
position, not a modulus bug.

## What perl says, which bounds what may be asserted

Measured on 5.42.0:

    ok 8;                 SYNTAX ERROR -- "Number found where operator
                          expected (Do you need to predeclare ok?)"
    ok $x, 9;             syntax OK -- $x->ok, '???'  INDIRECT OBJECT
    sub ok {} ok 8, 9;    syntax OK -- a call
    sub f {} f 13 % 4;    syntax OK -- f 1, constant folded

So TWO cases, not one. The undeclared form is `parses: no`, and asserting
otherwise would have the corpus demand a tree for a program perl rejects
-- WRONG in the gate's sense, and M1's gate is Oracle WRONG = 0.

## Tier placement

Tier 07. `entersub`/`leavesub` are tier 07's INTRODUCES, so a case
needing a `sub` declaration cannot sit in tier 04 -- the dependency lint
would reject it. This is the "call-form slice from tier 07" the plan
already names as part of the smallest useful corpus.

## Context

- paths: `conformance/mdtest/subroutines.md`, `conformance/07_subroutines/README.md`
- commands: `go test ./internal/conformance/ -count=1`
- entrypoints: `conformance/mdtest/subroutines.md` ("Four spellings of one call" is the existing case to sit beside)

## Scope

Cases to add, each with its perl-verified verdict:

- a parenless call to a declared sub with ONE argument
- the same with several arguments, including a `%` expression (arith.t's shape)
- the UNDECLARED callee with a numeric argument, `parses: no`
- the undeclared callee with a SCALAR argument, which perl reads as an
  indirect-object method call and is therefore a different claim

Each records `refuses:` naming `01a0c13f-6816` where our parser declines,
so the refusal points at the live issue that fixes it rather than at
`unfiled`.

## Acceptance criteria

- [ ] A parenless call with arguments is in the corpus (`python3 tools/mdtest/tiersrc.py 07_subroutines | grep -qE '^[a-z_]+ +[0-9$]'`)
- [ ] The undeclared-callee case asserts perl refuses it (`grep -q 'parses: no' conformance/mdtest/subroutines.md`)
- [ ] Every new case agrees with perl and our refusals are recorded (`go test ./internal/conformance/ -run '^TestTierSubroutinesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesPerlValidated'`)
- [ ] The dependency lint accepts the tier placement (`go test ./internal/conformance/ -run '^TestTierSubroutinesLint$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesLint'`)
- [ ] The suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
