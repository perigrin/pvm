---
title: "`CORE::my` and `CORE::our` refuse in expression position"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T10:40:48.835553402Z
updated: 2026-09-29T10:52:09.257925616Z
sessions:
- start_sha: d779cb6b4445aee68bc6a11be71d4c22f84c74c8
  end_sha: de1c517390b0bc7eda5cf4a6dfc8f33a032d576a
  commits: 1
  started_at: 2026-09-29T10:40:50.485120557Z
  ended_at: 2026-09-29T10:52:09.257925616Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T10:40:50.485120557Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T10:52:09.257925616Z
observed_paths:
- internal/parse/canon.go
- internal/parse/coredecl_test.go
- internal/parse/t2_test.go
- internal/parse/term.go
- internal/parse/testdata/perlgitt.ratchet
---

The statement path names a declarator through keywordName, so `CORE::state $x = 1;` parses, but the term path (parseTerm's declarator branch, and canon's binding power for a declaration) looked the spelling up as written. `f(CORE::my $x)` and `(CORE::our $z = 1) + 1` refused.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'use feature "state"; ok(ref(CORE::state $y = "a"));
          f(CORE::my $x); my $n = (CORE::our $z = 1) + 1;'
    ok(ref(state $y = 'a'));
    f(my $x);
    my $n = (our $z = 1) + 1;

perl.git t/opbasic/concat.t:203.

## Acceptance criteria

- [ ] CORE:: declarators work in expression position (`go test ./internal/parse/ -run '^TestCoreDeclaratorInExpression$' -count=1 -v | grep -q '^--- PASS: TestCoreDeclaratorInExpression'`)
