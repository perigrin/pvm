---
title: A declarator followed by a subscript loses the subscript to a second statement
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T11:53:47.469850114Z
updated: 2026-09-19T11:53:47.469850114Z
---

A declarator followed by a SUBSCRIPTED variable stops after the variable and
leaves the subscript as a second statement. Found by canonical re-emission
(01a0a8d8) over T1, where it is the largest remaining class.

    $ local $SIG{__WARN__} = sub { 1 };

    statement            span="local $SIG"
      declaration local
        term $SIG
    statement            span="{__WARN__} = sub { 1 };"
      binary =
        anon_hash                       <-- the subscript, orphaned
          call __WARN__
        declaration sub

Correct for comparison, the same expression without the declarator:

    $ $SIG{__WARN__} = 1;

    statement
      binary =
        index {                         <-- one Index, as it should be
          term $SIG
          term 1

## This is the bug decl.go:34-37 already describes

That comment records the identical failure mode for `sub {`: "a Declaration
spanning only `sub`, then a SECOND statement holding an AnonHash for the
body. A tree that is not a parse of its source, and it round-trips."

`parseVarDecl` has it for the subscript case. The declarator parses its
variable and returns without continuing into the postfix chain, so `{...}`
is re-read in statement position as an anonymous hash.

## Why round-trip did not catch it

Same leaves, same order, every byte back -- `TestParsedFilesRoundTrip` is
green on these files. Canonical re-emission catches it because the emission
re-parses to a different tree, which is the disjointness chapter 7 §7.2(c)
claims and this is a measured instance of it.

## Scope

`local`, `my`, `our` and `state` all route through parseVarDecl, so fix it
once there rather than per declarator. Both subscript forms:

    local $h{k} = 1;
    local $a[0] = 1;

and the arrow form for completeness.

## Acceptance

- [ ] `local $SIG{__WARN__} = sub {}` is ONE statement holding one Index (`go test ./internal/parse/ -run '^TestDeclaredSubscriptIsOneStatement$' -count=1 -v | grep -q '^--- PASS: TestDeclaredSubscriptIsOneStatement'`)
- [ ] The same holds for `my`, `our` and `state`, and for `[` as well as `{` (`go test ./internal/parse/ -run '^TestDeclaredSubscriptIsOneStatement$' -count=1 -v | grep -q '^--- PASS: TestDeclaredSubscriptIsOneStatement'`)
- [ ] The T1 canon fixpoint improves, measured in files and bytes (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)

Plan: M1. Spec: §4.12.
