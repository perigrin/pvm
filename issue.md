---
title: try/catch/finally is declined as a statement keyword, and op/try.t refuses every clause
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-28T22:03:29.531868436Z
updated: 2026-09-28T22:03:29.531868436Z
---

`try BLOCK catch (VAR) BLOCK [finally BLOCK]` under feature 'try' was declined
through `statementKeywords` -- op/try.t carried 59 Unknowns, one per clause.

Measured on 5.42.0, a statement form with no `;` after the last block:

    $ perl -e 'use feature "try"; try { die "x\n" } catch ($e) { print "c $e" }
          finally { print "f\n" } print "after\n"'
    c x
    f
    after

The keyword reading is taken only when `catch` is followed by `(`. Try::Tiny
spells the same words as subs taking a block, and perl itself separates the
two by that paren: under the feature `catch {` is "catch block requires a
(VAR)", and without it `catch (` is a syntax error. So Try::Tiny stays the
`WORD BLOCK` call reading `defer` gets, and `statementKeywords` empties.

## Acceptance criteria

- [ ] try/catch/finally parses as one statement and canons to a fixpoint (`go test ./internal/parse/ -run '^TestTryCatch$' -count=1 -v | grep -q '^--- PASS: TestTryCatch'`)
- [ ] NEGATIVE: Try::Tiny's `try { } catch { };` does not become a try statement (`go test ./internal/parse/ -run '^TestTryTinyStaysACall$' -count=1 -v | grep -q '^--- PASS: TestTryTinyStaysACall'`)
- [ ] op/try.t is clean in the perl.git t/ ratchet (`grep -q '^0 op/try.t$' internal/parse/testdata/perlgitt.ratchet`)
- [ ] The refusal inventory still accounts for every site, with unimplemented_statement documented as unreached (`go test ./internal/parse/ -run '^TestEveryUnknownSiteHasACode$' -count=1 -v | grep -q '^--- PASS: TestEveryUnknownSiteHasACode'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
