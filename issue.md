---
title: "Four keywords missing from the table: delete, exists, goto, do FILE"
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
created: 2026-09-17T14:23:31.374573065Z
updated: 2026-09-17T14:24:46.542745387Z
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

- [ ] `delete $h{k}` and `delete $h->{k}` parse as calls with one argument (`go test ./internal/parse/ -run '^TestDeleteParses$' -count=1 -v | grep -q '^--- PASS: TestDeleteParses'`)
- [ ] `exists $h{k}` parses, and is a distinct name from `defined` (`go test ./internal/parse/ -run '^TestExistsParses$' -count=1 -v | grep -q '^--- PASS: TestExistsParses'`)
- [ ] `goto &foo`, `goto LABEL` and `goto $where` each parse (`go test ./internal/parse/ -run '^TestGotoForms$' -count=1 -v | grep -q '^--- PASS: TestGotoForms'`)
- [ ] `do 'f.pl'` parses, and `do { 1 }` still takes the block path (`go test ./internal/parse/ -run '^TestDoFileAndDoBlock$' -count=1 -v | grep -q '^--- PASS: TestDoFileAndDoBlock'`)
- [ ] None of the four is `Unknown`, and none swallowed the statement after it (`go test ./internal/parse/ -run '^TestKeywordGapsConsumeExactlyTheirArguments$' -count=1 -v | grep -q '^--- PASS: TestKeywordGapsConsumeExactlyTheirArguments'`)
- [ ] `delete` as a hash key, `$h{delete}`, still parses as a key (`go test ./internal/parse/ -run '^TestKeywordsAreStillIdentifiers$' -count=1 -v | grep -q '^--- PASS: TestKeywordsAreStillIdentifiers'`)
- [ ] T1 clean does not fall below 408 and Unknown does not rise above 4,944 (`go test ./internal/parse/ -run '^TestT1Ratchet$' -count=1 -v | grep -q '^--- PASS: TestT1Ratchet'`)
- [ ] The whole suite is green (`go test ./internal/... 2>&1 | grep -qv FAIL`)

The negative criterion is the one that matters: `internal/parse/termform_test.go`
records that treating every declarator as a declaration made `$h{field}` parse
its KEY as one. A keyword added without the identifier case repeats it.

The T1 AC depends on `TestT1Ratchet`, owned by the M1 gate and not yet written.

Spec: §4.14.2 (`LoopEx`, `DoFile`, and the named-builtins note). Plan: M1.
