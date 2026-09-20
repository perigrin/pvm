---
title: "T2 core is 30.4% against a 100% target: what the 39 files actually need"
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0c13f-c7d7-7d9c-9a85-60a59aa26bc0
created: 2026-09-20T22:59:07.720149993Z
updated: 2026-09-20T23:57:02.355272123Z
---

T2 core is M1's headline metric and it is not close. The plan's target is
**100% of 56 files**; the measurement is **30.4%, 17 files**, 39 short.

Nothing on the chain owns that gap. This issue records what the gap IS, so
the work can be decomposed from a measurement rather than from a guess.

## The measurement, 2026-09-20 at b6e7c351

649 Unknown nodes across 44 of the 56 files. Bucketed by construct, not by
file:

      179   opbasic/arith.t    `%` in a PARENLESS call's arguments
      ~151  many              a parenless call whose argument is a bare number
       65   comp/parser.t     (unbucketed)
       48   base/num.t        leading-decimal literals, among others
       41   comp/proto.t      M4 prototype work, not M1's
       37   base/lex.t        heredocs, leading decimals

## Four constructs isolated by minimal reproduction

Each measured against this parser; the first two are most of the corpus.

**1. `%` as modulus in a parenless call.** 179 nodes, all of arith.t.

    tryeq $T++, 13 % 4, 1, 'x';     1 Unknown
    f 13 % 4;                       1 Unknown
    f($a, 13 % 4);                  0   <-- inside parens it works
    my $x = 13 % 4;                 0

So the lexer's expect state is wrong for an argument position after a
parenless call name: it wants a term, reads `%` as a hash sigil, and the
statement falls to Unknown. §3.2 says a leading `%` in TERM position is a
sigil and nowhere else -- the bug is that this position is not a term.

**2. A parenless call with a bare number argument.**

    ok 8;                           1 Unknown
    ok 8 - 1;                       1 Unknown

`ok`, `print` and `is` account for ~151 rows between them, and this is the
shape. Likely the same root cause as (1): the argument position is not being
entered correctly.

**3. A leading-decimal literal.**

    my $x = .5;                     1 Unknown
    my $x = int($f * 100 + .5);     2 Unknown
    my $x = 0.5;                    0

**4. Heredocs in statement position.**

    print <<'EOF';\nbody\nEOF\n    1 Unknown

## What this issue is

A MEASUREMENT, filed so the T2 gap is visible on the chain. The fixes belong
in their own issues, decomposed from the buckets above -- (1) and (2) may
well be one issue and one root cause, which is worth establishing before
either is scoped.

comp/proto.t's 41 are M4's and should be excluded from any T2 target M1 is
held to.

## Acceptance

- [ ] Each bucket above has an issue or is explicitly out of M1's scope, recorded here
- [ ] The T2 shortfall map's per-file counts are regenerated after each lands (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)

## Why this is not itself the fix

Filing 39 files' worth of work as one issue is the thing that produced a
32/35 milestone with two targets missed. The buckets are the units; this is
the survey that names them.

Measured 2026-09-20. Plan: M1. Spec: §3.2, §7.3.
