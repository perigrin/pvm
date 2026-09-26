---
title: continuesVersionString has no reachable caller left
state: pending
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T06:43:08.447239008Z
updated: 2026-09-26T07:18:15.547836045Z
---

`continuesVersionString` now has no reachable caller. `scanVString` claims
a `v`-prefixed run before any dot reaches the number scanner, so the guard
in `startsLeadingDecimal` never fires.

Filed rather than deleted in the commit that stranded it, deliberately --
see below.

## Measured, 2026-09-26, at the v-string merge

Neutering the guard leaves both suites green:

    // startsLeadingDecimal: !continuesVersionString(l) -> true
    go test ./internal/lexer/   ok  2.793s
    go test ./internal/parse/   ok  13.615s

Its last remaining branch (the Word one) was already the only live half:
`a21b05de` deleted the Number branch after measuring THAT unreachable. This
is the same finding one level up -- the whole function, not one arm.

## Why it was left in place

`a21b05de`'s own lesson. That commit deleted a dead branch whose comment
promised a test, and the comment had become false precisely because an
earlier change stranded the branch without saying so. Deleting a second
dead thing in the commit that stranded it would repeat the shape: the
reader of `da4610cd` would see a deletion and not know whether it was part
of the v-string fix or incidental to it.

So the function stays, labelled at both comment sites, and this issue owns
removing it.

## AMENDED: it is not one function, it is about forty lines

The TIER_2 gate on `01a0dc26-f9d3` measured the stranded surface and this
issue named one fifth of it. Its ruling is worth recording, because the
principle held and the record did not: separating the stranding from the
deletion is right, since a history search can then answer which change
killed a thing -- but an issue naming one function when five things died
understates the debt, and the next worker will delete
`continuesVersionString`, see green, and leave the rest.

The full set, each verified dead by neutering it and running the suites:

    internal/lexer/scan.go
      continuesVersionString                  the whole function
      startsLeadingDecimal's guard on it

    internal/lexer/ident.go
      noteSignatures `case Word:` version-prefix branch
      versionPrefix
      `case Number:` pending-minor branch
      `case Operator:` dot-preserving branch
      minorAtLeast's pending caller

    internal/lexer/lexer.go
      pendingVersionMajor                     the field

    internal/parse/use.go
      isVersionPrefix                         dead at BOTH call sites
      the reassembly loop (use.go ~70-82)     unreachable

Measured: neutering the `case Word:` branch leaves `internal/lexer` and
`internal/parse` green, which is what takes the rest of the `ident.go`
chain with it -- nothing sets `pendingVersionMajor`, so nothing reads it.

All these sites now carry a comment saying they are stranded and naming
this issue, so a reader arriving at any one of them is not misled.
Deleting them is this issue's work.

## Scope

Delete `continuesVersionString` and the `startsLeadingDecimal` guard that
calls it, or establish that some input still reaches it and record that
input as a test.

Before deleting, MEASURE reachability rather than inferring it from a green
suite -- a green suite proves no TEST reaches it, not that no PROGRAM does.
`a21b05de` made that distinction for the Number branch by dumping tokens
for `foo.5.6` and finding the predecessor was a Word; do the same here.

The `startsLeadingDecimal` doc comment has already been corrected to say
one exception survives and which one went. If the guard goes, that whole
paragraph goes with it -- and per CLAUDE.md it must be rewritten to say
what the function now does, not merely trimmed.

## Context

- paths: `internal/lexer/scan.go` (`continuesVersionString`, `startsLeadingDecimal`, `scanVString`)
- commands: `go test ./internal/lexer/ -count=1`, `go test ./internal/parse/ -count=1`
- entrypoints: `startsLeadingDecimal`'s `!continuesVersionString(l)` conjunct

## Acceptance criteria

- [ ] `continuesVersionString` is gone, or a test reaches it (`! grep -q 'func continuesVersionString' internal/lexer/scan.go`)
- [ ] `startsLeadingDecimal`'s comment describes what it now does (`! grep -q 'version-string exception' internal/lexer/scan.go`)
- [ ] REGRESSION GUARD, passes today: the lexer suite holds (`go test ./internal/lexer/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse suite holds (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the conformance suite holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
