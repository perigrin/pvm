---
title: try/catch/finally are declined as statement forms, and need a clause chain rather than WORD BLOCK
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-26T15:09:53.411251375Z
updated: 2026-09-26T15:09:53.411251375Z
---

`try`, `catch` and `finally` sit in `statementKeywords`, so the whole
statement becomes `Unknown{unimplemented_statement}`. Measured at the commit
that removed `defer` from that list:

    try { 1 } catch ($e) { 2 }    two unimplemented_statement refusals
    continue { 1 }                one

## Why `defer` came out and these did not

Issue `01a0d087` made `WORD BLOCK` read as a block followed by a list, which
is the whole shape `defer BLOCK` has. `try` needs more than that shape:

    try BLOCK catch (VAR) BLOCK finally BLOCK

The second clause takes a PARENTHESISED VARIABLE between the keyword and its
block, which `WORD BLOCK` does not describe, and the three clauses are one
statement rather than three. So `try` is a chain like `if`/`elsif`/`else` --
`parseConditional` is the shape to follow, not `parseWordTerm`.

Measured on 5.42.0, the clauses are one statement and the variable is a
declaration in the catch block's scope:

    $ perl -MO=Deparse -e 'use feature "try"; try { 1 } catch ($e) { 2 }'
    use feature 'try';
    eval { 1; };
    if ($@) { my $e = $@; 2; }

## Scope

- Read `try BLOCK catch (VAR) BLOCK` and the optional `finally BLOCK` as ONE
  statement, following `parseConditional`'s flattened-chain shape.
- `continue` is the other entry left in `statementKeywords` and is a
  different form -- `while (...) BLOCK continue BLOCK` -- so it is not this
  issue unless the measurement says the shapes share a reader.
- Remove the entries from `statementKeywords` and update its docstring, which
  names what has gone as each form lands.

## Where the fixtures point here

`internal/conformance/conformance_test.go` carries two fixtures whose premise
is "a construct that still refuses", and both now use
`try { 1 } catch ($e) { 2 }` naming THIS issue:

- `TestCorpusSkipsAreDocumented` -- a refusing file must skip, naming its
  refusal.
- `TestRefusalCodeMismatchFails` -- a recorded code that does not match must
  fail.

Both have been re-pointed twice already, each time a refusal they depended on
was fixed. When this issue lands, both will fail with "no longer refuses with
a code; replace this fixture", which is the harness working.

## Context

- paths: `internal/parse/parse.go` (`statementKeywords`), `internal/parse/control.go`
  (`parseConditional`, the chain shape to follow)
- commands: `go test ./internal/parse/ -count=1`, `go test ./internal/conformance/ -count=1`
- precedent: `01a0d087` (`defer`, the `WORD BLOCK` shape)

## Acceptance criteria

- [ ] `try BLOCK catch (VAR) BLOCK` is one statement, not a refusal (`go test ./internal/parse/ -run '^TestTryCatchIsOneStatement$' -count=1 -v | grep -q '^--- PASS: TestTryCatchIsOneStatement'`)
- [ ] REGRESSION GUARD: the parse ratchets hold (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
