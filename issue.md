---
title: A signature placeholder with an empty default, `($=)`, refuses
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T08:36:04.033606018Z
updated: 2026-09-29T08:36:04.033606018Z
---

A signature placeholder may take an empty default, `($=)`: an optional parameter whose value is discarded. The parser reads the signature as a parenthesised list, where `$ =` is an assignment with no right operand, and refuses.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'use feature "signatures"; sub t027 ($ =) { 1 } sub t119 ($ =, $a = 333) { 1 }'
    sub t027 ($=) {
    sub t119 ($=, $a = 333) {

perl.git t/op/signatures.t:493 and 506. perly.y's sigscalarelem takes `optsigscalardefault`, which may be empty.

## Acceptance criteria

- [ ] `($=)` and `($ =, $a = 333)` parse as signatures (`go test ./internal/parse/ -run '^TestSignatureEmptyDefault$' -count=1 -v | grep -q '^--- PASS: TestSignatureEmptyDefault'`)
