---
title: "An anonymous sub's prototype or signature is not read: `sub ($$) { 1 }` is a call with a subscript"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T10:58:01.699040331Z
updated: 2026-09-29T10:58:03.141465097Z
sessions:
- start_sha: de1c517390b0bc7eda5cf4a6dfc8f33a032d576a
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T10:58:03.141465097Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T10:58:03.141465097Z
---

An anonymous sub's prototype or signature is not read. The lexer scans a prototype only after a sub's NAME, and parseAnonSub reads no signature, so the parens after `sub` stay code: `my $g = sub ($$) { 1 };` parses as a call to `sub` with a hash subscript, at Unknown=0, and `sub () { $f++ if 0; 1 }` splits at the modifier.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'my $f2 = sub () { $f++ if 0; 1 }; my $h = sub (&) { my $c = shift; };
          use feature "signatures"; my $s = sub ($x, $) { $x };'
    my $f2 = sub () { 1 }
    my $h = sub (&) {
    my $s = sub ($x, $) {

perl.git t/op/hash.t:79, t/op/current_sub.t:83, t/op/override.t, t/perf/opcount.t and others.

## Acceptance criteria

- [ ] An anonymous sub takes a prototype or signature, and `->sub(...)` and a sub named `method` stay calls (`go test ./internal/parse/ -run '^TestAnonymousSubHead$' -count=1 -v | grep -q '^--- PASS: TestAnonymousSubHead'`)
- [ ] REGRESSION GUARD: comp/proto.t's prototype count holds (`go test ./internal/lexer/ -run '^TestProtoCorpusLexes$' -count=1 -v | grep -q '^--- PASS: TestProtoCorpusLexes'`)
