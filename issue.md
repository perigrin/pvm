---
title: The corpus has no parenless call with an argument, which is T2's largest bucket
state: cancelled
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T02:06:40.108445691Z
updated: 2026-09-26T02:49:06.527416046Z
transitions:
- state: cancelled
  actor: human:git-zhi
  timestamp: 2026-09-26T02:49:06.527416046Z
---

**CLOSED BY CHAIN-REVIEW, 2026-09-26. The work already exists; this issue
was filed on a premise that a five-minute check would have refuted.**

## What this issue claimed

"The corpus's largest hole: no case is a parenless call with an argument."

## What is actually there

`conformance/mdtest/argument-extent.md`, tier 07, five cases:

    ## A parenless call is greedy                refuses: 01a0c432-fbd5
    ## A `($)` prototype cuts the extent to one  refuses: 01a0c432-fbd5
    ## An undeclared callee is a syntax error    parses: no
    ## A builtin's extent is cut by a paren
    ## `lock` is a named unary with its own op

Every case this issue's Scope section proposed to add is present, with
token facts and a recorded refusal -- including the `parses: no`
undeclared-callee case that this issue presented as the constraint perl
forces on us. `b4180102` reached that conclusion by measurement in
September and wrote it down.

## No residue

Measured before closing, in case the arith.t shape was distinct:

    sub f { return "f[@_]" }  print f 1, 2;            1 Unknown
    sub f { return "f[@_]" }  print f 13 % 4, 2;       1 Unknown
    sub tryeq ($$$$) { }      tryeq 1, 13 % 4, 1, 'x'; 1 Unknown

Identical refusals, all `trailing_tokens`. An EXPRESSION in argument
position is not a distinct behaviour from a literal, so no new case would
measure anything the corpus does not already measure.

## What the real defect is

Both refusing cases cite `01a0c432-fbd5`, whose state is `done`. A reader
following the citation lands on finished work and concludes the corpus
says nothing about the construct. That is a CITATION repair, not new
coverage, and it is now owned by the citation issue in this milestone
together with the other thirteen.

## Why the check was not made

This issue's own evidence contained the refutation. The measurement it
quotes reported tier 07 as "10/13 clean" and never asked WHICH three
refused; two of the three were these cases. Counting by tier and naming
by construct are different acts, and only the second can notice that a
construct is already covered.
