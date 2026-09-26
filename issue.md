---
title: Attribute the 1,118 bare-brace Unknowns using matched braces, not the nearest one
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T16:26:49.795028845Z
updated: 2026-09-26T16:26:49.795028845Z
---

1,118 of perl.git t/'s 7,745 Unknown nodes are a bare `}` -- eight times the
next bucket. The probe that found that has a KNOWN ARTIFACT, so the share a
labelled block actually owns is unmeasured. MEASUREMENT ONLY -- no fix.

## The bucket

Measured 2026-09-26 at `59029a2c`, Unknown spans by leading text:

    1118  "}"          <- this issue
     164  ")"
      52  ":"
      38  "is $s, join('', map chr,"
      35  "like $@, _create_mismatc"
      33  ","
      31  "->"

A bare `}` reported as an Unknown span means a block did not close and the brace
was left over. The construct that failed is whatever OPENED that block, not the
brace that reports it.

## The artifact, stated plainly

The probe walked back from the Unknown to the NEAREST preceding `{` and took the
words before it. When blocks nest, the nearest `{` is an INNER brace, so its top
three answers were:

      56  "}"           a brace, not a construct
      27  ")"           ditto
      19  ""            nothing at all

Those are the probe failing, not findings. Below them the real constructs
appear -- `default` (24), `eval` (27), `else` (17), `SKIP:` (9), `when` -- but
their counts are all suspect for the same reason: a hit attributed to the inner
brace is a hit missing from the outer one.

## What to produce

The same table, built by walking to the MATCHING open brace rather than the
nearest one. Then the share each construct owns, which is the number issue
`<labelled-block>` needs to know its own reach.

Bracket matching has to skip braces inside strings, regexes, heredoc bodies and
comments, or it will mismatch on exactly the files this goal cares about.
`internal/lexer.Tokenize` already does that work -- use the token stream rather
than scanning bytes, which is also how `fidelity.go` and `canon_test.go` walk
source.

## What is already known, so it need not be re-derived

Measured in isolation, one construct per program:

    SKIP: { print "in" }          Unknown=1     REFUSES
    SKIP: { last SKIP; }          Unknown=2     REFUSES, second node for `last`
    given/when/default            Unknown=2     REFUSES
    eval { ... }                  Unknown=0     parses
    eval { ... }; if ($@) {...}   Unknown=0     parses
    plain bare block              Unknown=0     parses
    if/else                       Unknown=0     parses

So `eval` and `else` appearing in the brace-context table are artifacts of the
nearest-brace walk; they cannot be causes. A labelled bare block appears in 148
of the 620 files and 83 of `op/`'s 227, which makes it the only measured cause
whose reach is the right order of magnitude for a 1,118-node bucket -- but
"right order of magnitude" is not a measurement, and that is what this issue is
for.

## Scope

1. Rebuild the brace-context table using matched braces from
   `lexer.Tokenize`'s stream. Report the top 30 openers.
2. Report what SHARE of the 1,118 each construct owns, and what share is
   attributable to nothing the table names.
3. Append the corrected table to
   `docs/plans/2026-09-26-perl-git-t-parses.md`, replacing the bullet in "What
   is NOT yet measured" that describes this artifact.
4. If the share owned by a labelled bare block is small, say so loudly -- that
   would mean the chain's ordering is wrong and issue `<labelled-block>` is not
   the first fix.

## Context

- paths: `docs/plans/2026-09-26-perl-git-t-parses.md`,
  `internal/lexer/lexer.go` (`Tokenize`, `Kind`, `OpenBracket`-equivalents),
  `internal/lexer/scan.go` (`IsTrivia`, if skipping trivia helps)
- commands: `go test ./internal/parse/ -run <your probe> -count=1 -v`
- corpus: `/home/perigrin/dev/perl5/t`, 620 `.t` files walked recursively
- NOTE: `lexer` has a bracket stack already; read `internal/lexer/brackets.go`
  before writing a matcher of your own.

## Acceptance criteria

- [ ] The plan holds a brace-context table built from MATCHED braces (`grep -qi 'matched brace\|matching open brace' docs/plans/2026-09-26-perl-git-t-parses.md`)
- [ ] The share of the 1,118 owned by a labelled bare block is recorded as a number (`grep -qE 'labelled.*[0-9]+ of 1,?118|[0-9]+ of the 1,?118' docs/plans/2026-09-26-perl-git-t-parses.md`)
- [ ] No probe file is left behind (`test -z "$(ls internal/parse/zz*_test.go 2>/dev/null)"`)
- [ ] REGRESSION GUARD, passes today: the parse suite holds (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
