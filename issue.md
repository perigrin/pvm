---
title: The corpus format has expect parsent and the lint cannot carry such a file
state: pending
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T22:11:13.008302813Z
updated: 2026-09-21T22:27:16.582086952Z
---

The corpus format documents `--- expect parsent` for a file perl REFUSES,
`ParseFile` reads it, `File.ExpectParsent` exists, and `verdict()` handles
it. But **no corpus file uses it, and the first one to try will fail the
dependency lint**.

MEASURED while executing 01a0c432-fbd5 (the call-form slice).

## Where the failure actually is

Not in perl, and not in `B::Concise`. All three invocations exit 255 on a
syntax error, measured against a file containing `f 1, 2;` with `f`
undeclared:

    perl -c file.pl              exit 255
    perl file.pl                 exit 255
    perl -MO=Concise,-exec ...   exit 255

The difference is entirely in how the two Go callers treat that exit.

`askPerl` in `run.go` treats it as an ANSWER:

    if err := exec.Command(perl, "-c", path).Run(); err != nil {
        return false, ""
    }

so a refusing file yields `compiles=false` and `verdict()` proceeds
correctly. The runner is already right.

`opsOf` in `lint.go` treats it as an ERROR:

    out, err := exec.Command(perl, "-MO=Concise,-exec", path).Output()
    if err != nil {
        return nil, fmt.Errorf("perl -MO=Concise refused the source: %w", err)
    }

and `TestCorpusLints` plus every tier's `*Lint` test call `opsOf`
unconditionally. So the lint fails on any `parsent` file, in any tier.

## Why it matters rather than being theoretical

The croak extraction (01a0c35f-1685, done) measured **203 cases** in perl's
own `t/lib/croak/` that fail `perl -c` -- the genuine must-not-parse
negatives, and the spec's whole argument for extracting them. None can
become a corpus file until this is fixed.

The case that found it: `f 1, 2` with `f` undeclared is a syntax error,
true with and without `use strict` and even when `sub f` is defined later
in the same file. That fact deserves a corpus file and could not have one;
it is asserted in Go instead, inside `TestTierCallFormsParenlessExtent`.

## The fix

`opsOf` should not be called for a file whose `ExpectParsent` is set, or
should return an empty op set rather than an error for one. Either makes
the lint vacuous for such files, which is correct: a program perl will not
compile has no optree, and the dependency check has nothing to say about
it. Prefer the caller-side skip if the callers can see the `*File` -- an
`opsOf` that silently returns empty for a genuinely broken file would hide
a real error in a `parses` file.

## What the fix must not break

`verdict()` consults perl FIRST, so a file claiming `parsent` that perl
ACCEPTS is reported as a CORPUS BUG before the parser is consulted at all
(`TestPerlAdjudicatesFirst` pins this). Whatever skips the lint must not
also skip that adjudication, or a wrong `parsent` claim becomes
unfalsifiable.
