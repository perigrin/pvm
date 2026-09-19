---
title: A quote delimiter is a character, not a byte
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T00:47:32.887135684Z
updated: 2026-09-19T07:11:16.029362672Z
sessions:
- start_sha: 0db016fd18cba84ae70dcfb32f3453a5771bd27d
  end_sha: 0db016fd18cba84ae70dcfb32f3453a5771bd27d
  commits: 0
  started_at: 2026-09-19T07:11:09.365240704Z
  ended_at: 2026-09-19T07:11:16.029362672Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T07:11:09.365240704Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-19T07:11:16.029362672Z
---

A quote-like operator's delimiter is matched one BYTE at a time, so a
multi-byte delimiter closes on its first byte and the remaining bytes fall out
as an Error token:

    my $x = qq ϟ a ϟ;

    Quote     "qq ϟ a \xcf"
    Error     "\x9f"

The Quote token ends in the middle of a UTF-8 sequence, which is a span that
cannot be re-lexed or shown to a user. An ASCII delimiter is unaffected:

    my $x = qq ! a !;     Quote "qq ! a !"     correct

## Where it shows

`t/uni/parser.t:8884`, which is the whole reason it was found:

    ${
    #line 57
    qq ϟϟ }
    END

That file lexed clean until 01a0ad52 split dereferences into tokens. The
construct was inside a `${...}` the lexer swallowed whole, so nothing ever
reached the delimiter. The file going 0 -> 1 is the defect becoming visible,
not a new one -- verified by lexing `qq ϟ a ϟ;` on its own, which fails the
same way with no dereference anywhere near it.

## Why it is not a UTF-8 identifier problem

`use utf8` widens the IDENTIFIER class (`scanIdentRunes`), and that part
works. A delimiter is not an identifier: any non-word character may be one,
and perl compares the whole character. scanDelimitedBody compares a byte.

## What perl actually accepts, measured on 5.42.0

Only the NON-PAIRED forms. Verified by running each:

    qq ϟ hello ϟ        works
    q ϟ raw ϟ           works
    s ϟ b ϟ X ϟ         works, three parts
    q«paired»           "Use of '«' is deprecated as a string delimiter"
                        then "Can't find string terminator"

So the paired multi-byte case is not a requirement -- perl itself rejects it.
Scoping to what perl takes, rather than to every shape the form suggests.

## Acceptance criteria

- [x] `qq ϟ a ϟ` and `q ϟ a ϟ` each produce one Quote token spanning the
      whole construct
- [x] The three-part `s ϟ a ϟ b ϟ` does too
- [x] No Quote token ever ends mid-sequence: every token boundary falls on a
      UTF-8 character boundary
- [x] `t/uni/parser.t` returns to 0 in the lexer ratchet
- [x] The whole suite is green


## Done, 2026-09-19 (0db016fd)

    uni/gv.t       nodes 5 -> 0    bytes     5 -> 0
    uni/parser.t   nodes 1 -> 0    bytes     1 -> 0
    uni/stash.t    nodes 2 -> 0    bytes 5,448 -> 0

    full suite     green, 0 FAIL across the repo

All three acceptance criteria met. The delimiter is a rune everywhere --
skipToDelimiter, scanDelimitedBody, pairedCloser -- and every advance moves a
whole character.

## Two defects, and the second was hiding behind the first

With only the rune delimiter the ratchet reported three files BETTER
(gv.t 5 -> 1) and would have been regenerated on that reading. Measured in
BYTES first, as the ratchet's own doc comment requires, and gv.t had gone
from 5 stranded bytes to ONE token swallowing 10,124.

A regression that scores better on a node count. This is exactly the blind
spot the byte rule exists for, and it is the second time this milestone that
a broad rule turned out to be papering over two separate faults.

## The second defect: a name is a character too

t/uni/gv.t:488

    my $rv = \*sምḲ;

`sምḲ` is ONE identifier under `use utf8`, so the leading `s` is not a
substitution. Verified by running it:

    use utf8; sub sምḲ { 42 } print sምḲ();    ->  42

The guard that stops a keyword being an operator tested `isWordByte`, which
is ASCII-only, so `ም` read as a delimiter and the substitution ran to EOF. It
now decodes a rune and asks identContinue -- the same predicate the
identifier scanner uses, so the two cannot disagree about where a name ends.

That defect predates this issue: a byte delimiter also mis-read the name,
just less destructively. Fixing the delimiter is what made it visible.

## Scope held to what perl accepts

Only the non-paired forms, measured on 5.42.0. pairedCloser takes a rune to
match its caller but its table stays ASCII: perl rejects a paired multi-byte
delimiter outright, so there is no pair to add.

## The property, not just the instance

TestNoTokenEndsMidSequence asserts that no token boundary falls inside a
UTF-8 sequence, so the next bug of this shape fails without being named.
TestAsciiDelimiterStillWorks holds the single-byte path.

One test row of mine was wrong and perl said so: `qq #a#` SPACED is a comment
whose delimiter is on a later line, so it correctly yields no Quote. The row
is the glued `qq#a#`, which is the form where `#` delimits.
