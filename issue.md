---
title: unfiled is a legal refusal citation and nothing rejects it
state: pending
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T02:09:54.934577593Z
updated: 2026-09-26T02:52:33.82661721Z
---

**CANCELLED BY CHAIN-REVIEW, 2026-09-26. The premise is false: `unfiled`
is a deliberate sentinel with a documented rationale and a passing test
asserting it must NOT be looked up.**

## What this issue claimed

"`unfiled` is a legal value of `refuses:` and nothing rejects it. Sixteen
cases reached the corpus with it."

## What `unfiled` actually is

`internal/conformance/file.go:275`:

    // unfiledRefusal is the same claim in the topic format's vocabulary.
    //
    // `--- ` files spelled it `STATUS refuses as of this file`, and "this
    // file" WAS the address: one file, one case. A topic holds several,
    // so the phrase stopped pointing anywhere and the port spells it
    // `refuses: unfiled` instead -- the same claim, that the refusal is
    // recorded by the case that carries it and by no issue.
    const unfiledRefusal = "unfiled"

It is the topic-format spelling of `selfRecorded` ("this file"), and it
means THE CASE IS ITS OWN RECORD -- the case body states the measured
behaviour and the tokens we produce instead, so an issue would be a second
copy free to go stale.

`TestRefusalCitationMustResolve/unfiled_is_not_looked_up` asserts this and
PASSES. Its comment rejects this issue's framing in advance:

    // Sixteen cases carry it, and every one was a `.t` file that said
    // `refuses as of this file` -- so a resolver that knew only the old
    // phrase would have reported all sixteen as citations pointing at
    // nothing, which is a claim about the corpus that is not true.

Implementing this issue would require deleting that test and the
`unfiledRefusal` const -- a decision about what `refuses:` MEANS, which is
nowhere in this chain and is not a validator gap.

## What survives

Two narrower things, neither of which needs this issue:

1. **The four citations naming CLOSED issues.** That is the real defect and
   `01a0d0c7` owns it.
2. **`refuses:` accepts an arbitrary string.** `yes`, `TODO`, `see above`
   would all pass, because only `unfiled` and `this file` are special-cased
   and everything else is assumed to be an id. A real but minor shape
   check, deferred rather than bundled here.

## Why the check was not made

`grep -rn unfiledRefusal internal/` would have found the const, its
rationale, and the test, in one command. I read the corpus files and never
read the reader.
