---
title: "Identifier character class: non-ASCII names under use utf8"
state: pending
urgency: normal
milestone: m0-lexer-round-trip
created: 2026-09-14T15:42:10.413129176Z
updated: 2026-09-14T15:42:30.248888514Z
---

§0.13 ranks this FIRST by a wide margin: 34 corpus files, every
`t/mro/*_utf8.t` and most of `t/uni/`. It is a lexer defect, not a grammar
one, which is why it lands in M0.

Measured failures, each verified to fail on its own:

    use utf8; package Føø::Bær;
    @ᕘ::ISA = 'x'
    *ᕘ::ᕘ_Ƒ운ℭ = sub {}
    $Àlìcè::VERSION = 1

The discriminating measurement: `sub ᕘ { 1 }` alone PARSES, but
`(shift)->SUPER::ᕘ` does not. So the bug is the identifier character class
at every position, not a special case after `package`. A fix that only
widens the post-`package` path will pass a naive test and leave 30 files
failing.

## What perl actually does

`toke.c` decides identifier characters through `isWORDCHAR_utf8_safe` and
friends when the `utf8` pragma is on, and through plain `isWORDCHAR` when it
is not. The pragma is lexically scoped, so this is lexer state, not a global
mode: the same bytes are an identifier in one block and a syntax error in
the next.

Read the real source before writing the character class. Do not infer it
from the Unicode standard — perl's rule is `XIDStart`/`XIDContinue` with
perl-specific additions, and the additions are where the corpus files live.

## Acceptance

- [ ] An ASCII identifier lexes identically with and without `use utf8` (`go test ./internal/lexer/ -run TestIdentifierASCIIUnaffected -count=1`)
- [ ] The four measured §0.13 forms lex as identifiers under `use utf8` (`go test ./internal/lexer/ -run TestIdentifierUTF8Corpus -count=1`)
- [ ] A non-ASCII identifier is a lex error WITHOUT `use utf8`, matching perl (`go test ./internal/lexer/ -run TestIdentifierRequiresPragma -count=1`)
- [ ] The pragma's lexical scope is honoured: same bytes, different verdict per block (`go test ./internal/lexer/ -run TestIdentifierPragmaIsScoped -count=1`)
- [ ] Every identifier position accepts the class, not only after package (`go test ./internal/lexer/ -run TestIdentifierEveryPosition -count=1`)

Findings: §0.13 rank 1 (34 files). Spec: ch2 lexical structure.
