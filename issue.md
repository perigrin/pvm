---
title: "A call before -> takes the arrow as its argument: shift->m canons as shift(->);m()"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T00:32:51.563768272Z
updated: 2026-10-11T15:17:46.237522164Z
sessions:
- start_sha: 0c70adb2c77f4e06e9adf757e2199a2580c48a09
  end_sha: 468aa62746c7cda08884e4f3e2c6c593febca14f
  commits: 1
  started_at: 2026-09-29T00:33:08.179846967Z
  ended_at: 2026-09-29T00:44:47.790498519Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T00:33:08.179846967Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T00:44:47.790498519Z
observed_paths:
- internal/parse/arrowargs_test.go
- internal/parse/call.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1fidelity.ratchet
---

A call with no parens followed by `->` takes the arrow as its argument:

    my $x = shift->m;              canon  my $x = shift(->);m();        Unknown=1
    sub t { 1 } my $x = t->m(1);   canon  ... my $x = t(->);m(1);       Unknown=1

perl calls the sub with no arguments and invokes the method on the result.
Measured on 5.42.0, with and without a prototype:

    $ perl -MO=Deparse -e 'sub t (;$) { 1 } my $x = t->m(1);'
    my $x = t()->m(1);

`endsArgumentList` lists the infix operators that cannot start a term, so a
call before one takes no argument; `->` cannot start a term either and is
missing. `shift->method` is the idiom every OO module's accessors use, and 8
perl.git t/ files spell it.

Found when lexical subs (01a0ea2e-9771) started declaring `my sub t (;$)` and
the T1 file lexical_sub_prototype_method_invocant.t, which calls `t->...`, went
from 1 to 2 Unknowns: the declaration was right, and this bug was waiting
behind it.

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance criteria

- [ ] A builtin or declared sub before `->` takes no argument, and the arrow applies to its result (`go test ./internal/parse/ -run '^TestArrowEndsArguments$' -count=1 -v | grep -q '^--- PASS: TestArrowEndsArguments'`)
- [ ] REGRESSION GUARD: a class name before `->` is still a class name (`go test ./internal/parse/ -run '^TestArrowEndsArguments$' -count=1 -v | grep -q '^--- PASS: TestArrowEndsArguments'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
