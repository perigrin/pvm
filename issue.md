---
title: A quote delimiter is a character, not a byte
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T00:47:32.887135684Z
updated: 2026-09-19T00:47:47.468496818Z
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

- [ ] `qq ϟ a ϟ` and `q ϟ a ϟ` each produce one Quote token spanning the
      whole construct
- [ ] The three-part `s ϟ a ϟ b ϟ` does too
- [ ] No Quote token ever ends mid-sequence: every token boundary falls on a
      UTF-8 character boundary
- [ ] `t/uni/parser.t` returns to 0 in the lexer ratchet
- [ ] The whole suite is green
