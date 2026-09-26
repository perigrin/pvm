---
title: Bare v5 is a v-string in perl, and a lexer cannot tell it from a bareword
state: pending
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T06:41:03.496757916Z
updated: 2026-09-26T06:41:03.496757916Z
---

Bare `v5` is a v-string in perl and an identifier to us. It is the third
and last of the v-string forms; the two-dot and one-dot forms are fixed.

Filed by the worker that fixed the one-dot form, which measured this and
correctly declined it as a distinct decision rather than a continuation.

## Measured, 2026-09-26

    $ perl -e 'print length(v5)'
    1

One character, so perl holds a v-string of one ordinal. We lex `v5` as
`Word(v5)`.

The other two forms now work:

    v5.42.0   one Quote   (01a0db78)
    v5.36     one Quote   (01a0dc26-f9d3)
    v5        Word        THIS

## Why it is a decision and not a continuation

`scanVString` currently requires at least one dot. Dropping that
requirement claims EVERY `v`-plus-digits bareword as a v-string, and the
worker's note is the whole argument: that is a much wider net than the
dotted forms, which are unambiguous.

Consider what stops being an identifier:

    v5        a v-string in perl, yes
    v1        likewise
    sub v2 {} a SUB NAME that happens to match the shape

So the fix needs a rule for when `v\d+` is a version and when it is a
name, and perl's own answer is context-dependent in a way the dotted forms
are not. That is the measurement to make first, not a smaller version of
the last two issues.

## Do not assume it is cheap because the last two were

`01a0dc26-f9d3` found its stated cost was ZERO -- `use.go` needed no
change, because a `Quote` falls through to `parseExpr` and yields the same
version Term. That will not transfer: `v5` as a Quote would change what a
bareword in sub-name position lexes as, which is a different blast radius
from a dotted literal nobody writes as a name.

Measured already, and the ambiguity is real:

    $ perl -e 'sub v2 { 7 } print v2();'          ->  7
    $ perl -e 'my %h = (v5 => 1); print keys %h;' ->  v5

So perl reads the SAME `v5` three ways by context: a v-string as a term, a
sub name after `sub`, and the autoquoted string `"v5"` before a fat comma.
A lexer cannot settle that from the bytes alone, which is why this is the
form left over after the two unambiguous ones.

That may mean the honest answer is to leave `v5` a Word and record WHY,
rather than to claim it. Deciding that is this issue; `v5->foo` and
sub-name position are the cases to measure next.

## Where it is recorded now

`conformance/mdtest/vstrings.md` records it in prose, and the
`scanVString` comment names it. Neither is a tracker, which is why this
exists.

## Context

- paths: `internal/lexer/scan.go` (`scanVString`), `internal/lexer/ident.go` (`scanWord`, `versionPrefix`), `conformance/mdtest/vstrings.md`
- commands: `go test ./internal/lexer/ -count=1`, `go test ./internal/parse/ -count=1`
- entrypoints: `scanVString`'s dot-count requirement, and `scanWord`, which currently claims `v5`

## Acceptance criteria

- [ ] The decision is recorded with the bareword measurements behind it (`grep -qE 'sub v2|bareword' conformance/mdtest/vstrings.md`)
- [ ] Whichever way it went, `v5` is no longer an unrecorded gap (`! grep -qi 'STILL OPEN' conformance/mdtest/vstrings.md`)
- [ ] REGRESSION GUARD, passes today: the lexer suite holds (`go test ./internal/lexer/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse suite holds (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
