---
title: The corpus format has expect parsent and the lint cannot carry such a file
state: pending
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T22:11:13.008302813Z
updated: 2026-09-21T22:11:13.008302813Z
---

The corpus format documents `--- expect parsent` for a file perl REFUSES,
`ParseFile` reads it, `File.ExpectParsent` exists, and `verdict()` handles
it. But **no corpus file uses it, and the first one to try will fail**.

MEASURED while executing 01a0c432-fbd5 (the call-form slice):

`TestCorpusLints` and every tier's `*Lint` test call `opsOf`
unconditionally, and `opsOf` runs

    perl -MO=Concise,-exec <path>

which **exits 255 on a syntax error**. A `parsent` file is by definition
one perl refuses, so `opsOf` fails on it, so the lint fails on it. This is
true for a `parsent` file in ANY tier, not just tier 07.

So the corpus can express "perl refuses this" in its format and cannot
carry such a file through its own gates.

WHY IT MATTERS RATHER THAN BEING THEORETICAL: the croak extraction issue
(01a0c35f-1685, done) measured 203 cases in perl's own `t/lib/croak/` that
fail `perl -c` -- genuine must-not-parse negatives. The spec's whole
argument for extracting them is that "croak does not mean must-not-parse"
and the 203 that DO are the real negatives. None can become a corpus file
until this is fixed.

The immediate case that found it: `f 1, 2` with `f` undeclared is a syntax
error in perl ("Number found where operator expected (Do you need to
predeclare f?)"), true with and without `use strict` and even when `sub f`
is defined later in the same file. That is a fact worth a corpus file and
it could not have one; it is asserted in Go instead, inside
`TestTierCallFormsParenlessExtent`, where no optree is requested.

THE FIX IS SMALL AND THE QUESTION IS WHERE: `opsOf` should not be called
for a file whose `ExpectParsent` is set, or should return an empty op set
rather than an error for one. Either makes the lint vacuous for such files,
which is correct -- a program perl will not compile has no optree to lint,
and the dependency check has nothing to say about it.

Note the ordering property this must not break: `verdict()` already gets
this right by consulting perl FIRST, so a file claiming `parsent` that perl
ACCEPTS is reported as a CORPUS BUG before the parser is consulted at all
(`TestPerlAdjudicatesFirst`). Whatever skips the lint must not also skip
that adjudication.
