---
title: A declarator followed by a subscript loses the subscript to a second statement
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T11:53:47.469850114Z
updated: 2026-09-19T12:44:24.41492651Z
sessions:
- start_sha: 5c9be6bd44b1282e2477531f28d641caad4eb9a0
  end_sha: 5c9be6bd44b1282e2477531f28d641caad4eb9a0
  commits: 0
  started_at: 2026-09-19T12:11:22.218620805Z
  ended_at: 2026-09-19T12:12:26.849568149Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T12:11:22.218620805Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T12:12:26.849568149Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-19T12:44:24.41492651Z
observed_paths:
- internal/parse/canon.go
- internal/parse/canon_test.go
- internal/parse/decl.go
- internal/parse/declsubscript_test.go
- internal/parse/easy_test.go
- internal/parse/expr.go
- internal/parse/keyquote_test.go
- internal/parse/precedence.go
- internal/parse/t2_test.go
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1canon.ratchet
- internal/parse/use_test.go
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

- [x] `local $SIG{__WARN__} = sub {}` is ONE statement holding one Index (`go test ./internal/parse/ -run '^TestDeclaredSubscriptIsOneStatement$' -count=1 -v | grep -q '^--- PASS: TestDeclaredSubscriptIsOneStatement'`)
- [x] The same holds for `my`, `our` and `state`, and for `[` as well as `{` (`go test ./internal/parse/ -run '^TestDeclaredSubscriptIsOneStatement$' -count=1 -v | grep -q '^--- PASS: TestDeclaredSubscriptIsOneStatement'`)
- [x] The T1 canon fixpoint improves, measured in files and bytes (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)

Plan: M1. Spec: §4.12.


## Done, 2026-09-19 (657d0caf)

    T1 parse        460 -> 462 clean, 3322 -> 3317 Unknown
    T1 canon        566 -> 611 clean (57.4% -> 62.0%)
    T1-easy         53.0% -> 53.2%
    t/class         45 -> 43 Unknown
    T2 class/field  9 -> 7
    full suite green, 0 FAIL across the repo

parseVarDeclNoSemi took the variable as a bare Term and jumped straight to
the `=` check, never entering the postfix chain. The subscript was then
re-read in statement position, where a brace is an anonymous hash.

peekSubscript is shared with parseExpr rather than duplicated: `$h{k}` and
`$h->{k}` are one operation and one flag apart (§4.14), so both walks ask the
same question in one place.

## A second defect, found by measuring rather than by reading

Attaching the subscript made it visible: with the target complete, a
COMPOUND assignment fell out as an Unknown beside it, because the initialiser
check accepted only `=`.

    concat_assignment_env_taint.t   5 nodes 268 bytes -> 6 nodes 286 bytes

A rise in NODES and in BYTES, so not the ratchet's known blind spot but a
real regression. The source is `local $ENV{PATH} .= $tainted_empty;`, which
perl accepts. Level 9 is the whole assignment class in one token group
(toke.c:250), so the test is the level rather than nineteen spellings.

With that fixed the file returns to 5 nodes and no T1 file regressed --
2 better, 0 worse. Without the byte measurement this would have shipped as
"1 worse, 45 better, net win", and the 18 bytes would have stayed.

Acceptance criterion 3 asked for the canon ratchet to improve, and it did:
45 files moved 1 -> 0.
