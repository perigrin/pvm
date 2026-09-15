---
title: "Identifier character class: non-ASCII names under use utf8"
state: in-progress
urgency: normal
milestone: m0-lexer-round-trip
created: 2026-09-14T15:42:10.413129176Z
updated: 2026-09-15T04:48:38.257754802Z
sessions:
- start_sha: 912cd549cc31d6a08ec0a4d6dd2eca07cedc1bad
  end_sha: ""
  commits: 0
  started_at: 2026-09-15T04:48:38.257754802Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-15T04:48:38.257754802Z
---

§0.13 ranks this first: 34 corpus files, concentrated in `t/mro/*_utf8.t`
and `t/uni/`. It is a lexer defect, not a grammar one, which is why it is
M0.

Measured failures, each verified to fail on its own:

    use utf8; package Føø::Bær;
    @ᕘ::ISA = 'x'
    *ᕘ::ᕘ_Ƒ운ℭ = sub {}
    $Àlìcè::VERSION = 1

The discriminating measurement: `sub ᕘ { 1 }` alone PARSES, but
`(shift)->SUPER::ᕘ` does not. So the bug is the identifier character class
at every position, not a special case after `package`. A fix that only
widens the post-`package` path passes a naive test and leaves 30 files
failing.

## The rule, from perl's own documentation

`perldata.pod:179-180` gives it as

    (\p{Word} & \p{XID_Start}) + [_]      then      (\p{Word} & \p{XID_Continue})*

An INTERSECTION plus underscore — a restriction of XID, not an extension of
it. All four measured forms above are plain XID letters. Do not go looking
for perl-specific additions; there are none to find, and an earlier draft of
this issue said otherwise.

## Scope: file-level pragma, deliberately

`use utf8` is lexically scoped in perl, but spec §2.3.1 says "a file-level
flag is a reasonable v1 with a recorded limitation". Block scoping needs a
brace stack (the expect-state issue) and the ability to tell a real `{` from
one inside a string (the quote-scanner issue) — both later in this chain.
Implement the file-level flag, record the limitation, and leave scoping to
M1.

## The `'` package separator belongs here

Spec §2.6.3 and §2.15 item 4: BOTH separators, `::` and `'`. §0.13 rank 10.
Two T2 files use it, and it changes delimitation rather than merely naming:

    comp/package.t:17    $main'a = $'b;
    comp/parser.t:397    sub CORE'print'foo { 43 }

A lexer unaware of the separator opens a string at `'a = $'`. The same line
also contains `$'`, the postmatch variable, so the rule that a `'` is a
separator only between identifier characters is exercised by the corpus
rather than by a fixture.

## On "lex error without the pragma"

A byte outside `isWORDCHAR_L1` is a lex error without `use utf8` — verified,
`\x95` gives "Unrecognized character". A Latin-1 letter byte is NOT: `\xE1`
lexes fine as `$á` and fails later in the compiler. The test must pick a
byte outside Latin-1 word characters, or it asserts something perl does not
do.

## Acceptance

- [ ] An ASCII identifier lexes identically with and without use utf8 (`go test ./internal/lexer/ -run '^TestIdentifierASCIIUnaffected$' -count=1 -v | grep -q '^--- PASS: TestIdentifierASCIIUnaffected'`)
- [ ] The four measured §0.13 forms lex as identifiers under use utf8 (`go test ./internal/lexer/ -run '^TestIdentifierUTF8Corpus$' -count=1 -v | grep -q '^--- PASS: TestIdentifierUTF8Corpus'`)
- [ ] A byte outside Latin-1 word chars is an Error token without the pragma (`go test ./internal/lexer/ -run '^TestIdentifierRequiresPragma$' -count=1 -v | grep -q '^--- PASS: TestIdentifierRequiresPragma'`)
- [ ] The class is accepted at every identifier position, not only after package (`go test ./internal/lexer/ -run '^TestIdentifierEveryPosition$' -count=1 -v | grep -q '^--- PASS: TestIdentifierEveryPosition'`)
- [ ] Both :: and ' separate package names, and $' stays the postmatch variable (`go test ./internal/lexer/ -run '^TestIdentifierPackageSeparators$' -count=1 -v | grep -q '^--- PASS: TestIdentifierPackageSeparators'`)

Findings: §0.13 ranks 1 (34 files) and 10 (2 files). Spec: §2.3.1, §2.6.3, §2.15 item 4.
