---
title: A signature sub canonicalises as `sub f = ($a) {...}`, which perl rejects
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T08:19:53.771628631Z
updated: 2026-09-29T08:23:34.931305919Z
sessions:
- start_sha: 908e6f5973a65e853f2dce0be7bafa7f7495192e
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T08:19:54.4680783Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T08:19:54.4680783Z
---

A signature is a parenthesised list child of the sub's Declaration, and canon joins it to the name with ` = `, the separator for a variable's initialiser. The tree is right and the emission is not Perl:

    use feature "signatures"; sub f ($a) { $a }     canon: sub f = ($a) {$a;}

Measured with perl 5.42:

    $ perl -e 'use feature "signatures"; sub f = ($a) { $a }'
    Illegal declaration of subroutine main::f at -e line 1.

Deparse writes `sub f ($a) {`.

A one-element signature with a default is worse: the element is its own node with Paren set, not a List, and canon rebuilt it without the parentheses:

    use feature "signatures"; sub d ($x = 1) { $x }  canon: sub d = $x = 1 {$x;} The count cannot see it, since Unknown=0 throughout.

## Acceptance criteria

- [ ] A signature follows the sub name with a space in canon (`go test ./internal/parse/ -run '^TestSignatureCanonHasNoAssignment$' -count=1 -v | grep -q '^--- PASS: TestSignatureCanonHasNoAssignment'`)
- [ ] REGRESSION GUARD: the fidelity ratchet holds (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
