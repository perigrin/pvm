---
title: The lexer splits -e into a minus and a word, and the whole file-test family with it
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-23T17:51:06.293983513Z
updated: 2026-09-23T17:51:06.293983513Z
---

`-e $f` is a file-test operator and our lexer splits it into a minus and
a word. Found by the conformance corpus while writing tier 04's unary
slice (issue 01a0cc04-341c).

## Measured

    my $f = "/etc/hostname"; print "[", (-e $f), "]\n";

      Operator(-) Word(e) Variable($f)

Perl lexes `-e` as ONE file-test operator; we produce two tokens.

## Why it matters beyond this one operator

perlop's level 10 holds the whole family -- `-e -d -f -s -z -r -w -x -M
-A -C` and more -- so this is one rule covering a dozen spellings.

The minus is the ambiguity. A single `-` has at least three readings
decided by what follows it:

    -e $f        file test
    -$x          arithmetic negation
    -bareword    the STRING "-bareword", which tier 01's
                 07_negative.t already measures as two tokens

A lexer that leaves a bare `-` in front of a word has not decided which
of those it meant, and the decision is exactly what the level exists for.

## Where it is recorded

`conformance/04_operators/31_file_test.t` carries the refusal with a
token fact asserting what a correct lexer produces: `no operator whose
text is "-"`, because the whole `-e` should be the operator's name.
