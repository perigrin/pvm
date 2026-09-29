---
title: A signature placeholder with an empty default, `($=)`, refuses
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T08:36:04.033606018Z
updated: 2026-09-29T09:17:57.109368847Z
sessions:
- start_sha: c54e1e5b90bc395cfaca3eefadfd704a3a4a6ee5
  end_sha: 6f5e809bf3b21ccb29e252f189b7d6c7edc7461e
  commits: 1
  started_at: 2026-09-29T09:01:43.467486835Z
  ended_at: 2026-09-29T09:17:57.109368847Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T09:01:43.467486835Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T09:17:57.109368847Z
observed_paths:
- internal/parse/decl.go
- internal/parse/expr.go
- internal/parse/parse.go
- internal/parse/sigemptydef_test.go
- internal/parse/testdata/perlgitt.ratchet
---

A signature placeholder may take an empty default, `($=)`: an optional parameter whose value is discarded. The parser reads the signature as a parenthesised list, where `$ =` is an assignment with no right operand, and refuses.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'use feature "signatures"; sub t027 ($ =) { 1 } sub t119 ($ =, $a = 333) { 1 }'
    sub t027 ($=) {
    sub t119 ($=, $a = 333) {

perl.git t/op/signatures.t:493 and 506. perly.y's sigscalarelem takes `optsigscalardefault`, which may be empty.

## Acceptance criteria

- [ ] `($=)` and `($ =, $a = 333)` parse as signatures (`go test ./internal/parse/ -run '^TestSignatureEmptyDefault$' -count=1 -v | grep -q '^--- PASS: TestSignatureEmptyDefault'`)
