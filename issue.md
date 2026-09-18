---
title: "Four keywords missing from the table: delete, exists, goto, do FILE"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
created: 2026-09-17T14:23:31.374573065Z
updated: 2026-09-18T09:48:31.315942628Z
sessions:
- start_sha: 6393fdaa0bd1a57acd1acc0fbf10810c76893536
  end_sha: ""
  commits: 0
  started_at: 2026-09-18T09:35:10.078210719Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-18T09:35:10.078210719Z
---

Four constructs are `Unknown` because a word is missing from a table, not because anything is hard. Measured at 3c997a6b, each round-tripping:

    delete $h{k};    [source_file unknown unknown]
    exists $h{k};    [source_file unknown unknown]
    goto &foo;       [source_file unknown]
    do 'f.pl';       [source_file unknown]

`chomp $x;` parses, which is the contrast: `chomp` is in
`internal/parse/keyword.go` and these are not.

## What each needs

**`delete` and `exists`** are named unary operators taking one argument.
§4.14.2 records that B::SoN gives each its own IR kind and explains why —
`exists` is membership, not definedness. Measured on perl 5.42.0:

    my %h = (u => undef);
    exists $h{u}   # true
    defined $h{u}  # false

Here they are `Call{Name}` and the name carries the distinction, so this is
a keyword-table row each, nothing more.

**`goto`** takes a label, an expression, or `&NAME`. §4.14.2 puts it under
`LoopEx{Op,Label}` alongside `last`/`next`/`redo`, which already parse:
`last FOO;` gives `[loop_control "last" [label "FOO"]]`. The `&foo` form is
the one in the corpus and the one that does not parse.

**`do FILE`** is `do` with an expression where `do BLOCK` has a block.
`internal/parse/term.go` already splits these two — `parseBlockOperator`
takes the block form and the expression form goes through `parseWordTerm` as
an ordinary named unary — so `do 'f.pl'` should already work and does not.
Find out why before adding anything; the fix may be that `do` is on a list it
should not be on. §4.14.2 marks `DoFile` SPECULATIVE for its consumer: no
inference rule reads it. Parse it anyway — an `Unknown` in the middle of a
file costs more than a node nobody reads.

## Scope

Table rows and whatever minimal parse path each needs. Not in scope: what
`delete` returns, whether `goto &f` unwinds, or executing `do FILE`. This
milestone is about producing a tree, not about semantics.

## Acceptance

- [x] `delete $h{k}` and `delete $h->{k}` parse as calls with one argument (`go test ./internal/parse/ -run '^TestDeleteParses$' -count=1 -v | grep -q '^--- PASS: TestDeleteParses'`)
- [x] `exists $h{k}` parses, and is a distinct name from `defined` (`go test ./internal/parse/ -run '^TestExistsParses$' -count=1 -v | grep -q '^--- PASS: TestExistsParses'`)
- [x] `goto &foo`, `goto LABEL` and `goto $where` each parse (`go test ./internal/parse/ -run '^TestGotoForms$' -count=1 -v | grep -q '^--- PASS: TestGotoForms'`)
- [x] `do 'f.pl'` parses, and `do { 1 }` still takes the block path (`go test ./internal/parse/ -run '^TestDoFileAndDoBlock$' -count=1 -v | grep -q '^--- PASS: TestDoFileAndDoBlock'`)
- [x] None of the four is `Unknown`, and none swallowed the statement after it (`go test ./internal/parse/ -run '^TestKeywordGapsConsumeExactlyTheirArguments$' -count=1 -v | grep -q '^--- PASS: TestKeywordGapsConsumeExactlyTheirArguments'`)
- [x] `delete` as a hash key, `$h{delete}`, still parses as a key (`go test ./internal/parse/ -run '^TestKeywordsAreStillIdentifiers$' -count=1 -v | grep -q '^--- PASS: TestKeywordsAreStillIdentifiers'`)
- [x] T1 clean does not fall below 408 and Unknown does not rise above 4,944 (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [x] The whole suite is green (`go test ./internal/... 2>&1 | grep -qv FAIL`)

The negative criterion is the one that matters: `internal/parse/termform_test.go`
records that treating every declarator as a declaration made `$h{field}` parse
its KEY as one. A keyword added without the identifier case repeats it.

The T1 AC uses `TestParseRatchet`, committed at 4bc971ec.

Spec: §4.14.2 (`LoopEx`, `DoFile`, and the named-builtins note). Plan: M1.


## Done, 2026-09-18 (f905b4e9)

    T1 clean     408 -> 434 (44.0%)
    T1 Unknown   4,944 -> 4,627
    T2 clean     15 -> 16 of 56
    t/class      48 -> 47

73 T1 files improved; array_delete_exists.t went 49 -> 0.

All four classified by keyword.go's own stated method -- deparse `KW $x, $y`
and see where the comma lands:

    (delete $h{'a'}, $y)   outside  -> named unary
    (exists $h{'a'}, $y)   outside  -> named unary
    ((goto $x), $y)        outside  -> named unary
    ((do $x), $y)          outside  -> named unary

GOTO NEEDED MORE THAN A TABLE ROW. `perl -e 'sub FOO { 42 } goto FOO;'` says
"Can't find label FOO" -- it looks for a LABEL, not the sub. So parseGoto
joins last/next/redo in parseControlFlow, per §4.14.2's `LoopEx{Op,Label}`
grouping, and `goto` came off statementKeywords where it sat unimplemented.

TWO FILES GOT WORSE AND IT IS NOT THIS CHANGE. `defined &$r` already split
into two statements at HEAD -- the `&` code-ref sigil after a named unary is
unhandled. Adding `exists` made a second keyword reach the same pre-existing
bug. Verified by stashing and re-measuring. Named in the baseline rather than
absorbed; nothing in the chain owns it.

An off-by-one was found while counting: the table's comment said 77 entries
and it held 76, before these four. Now 80, with both numbers recorded.
