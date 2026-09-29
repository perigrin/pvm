---
title: A signature sub canonicalises as `sub f = ($a) {...}`, which perl rejects
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T08:19:53.771628631Z
updated: 2026-09-29T08:19:53.771628631Z
---

A signature is a parenthesised list child of the sub's Declaration, and canon joins it to the name with ` = `, the separator for a variable's initialiser. The tree is right and the emission is not Perl:

    use feature "signatures"; sub f ($a) { $a }     canon: sub f = ($a) {$a;}

Measured with perl 5.42:

    $ perl -e 'use feature "signatures"; sub f = ($a) { $a }'
    Illegal declaration of subroutine main::f at -e line 1.

Deparse writes `sub f ($a) {`. The count cannot see it, since Unknown=0 throughout.

## Acceptance criteria

- [ ] A signature follows the sub name with a space in canon (`go test ./internal/parse/ -run '^TestSignatureCanonHasNoAssignment$' -count=1 -v | grep -q '^--- PASS: TestSignatureCanonHasNoAssignment'`)
- [ ] REGRESSION GUARD: the fidelity ratchet holds (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
