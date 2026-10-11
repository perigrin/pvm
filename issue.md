---
title: A prototype's optional slots are not counted, so ($;$) parses as unary and drops its second argument
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T02:17:43.942577971Z
updated: 2026-10-11T15:17:47.495957007Z
sessions:
- start_sha: b845d9b4e8940e6f86d5c0d3f74acc5510290083
  end_sha: f9c96510c1142c37237b7b0cf0013b9a12ac0bf1
  commits: 1
  started_at: 2026-09-29T02:18:00.221825456Z
  ended_at: 2026-09-29T02:35:12.206699859Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T02:18:00.221825456Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T02:35:12.206699859Z
observed_paths:
- internal/parse/protoarity_test.go
- internal/parse/shape.go
- internal/parse/t2_test.go
- internal/parse/testdata/perlgitt.ratchet
---

ShapeOf counts only the MANDATORY slots of a prototype, so a prototype with
one mandatory slot and optional ones after it -- or two optional ones -- is
read as a named unary, and a parenless call loses every argument past the
first:

    sub f ($;$) {} f $a, $b;       parser: (f($a), $b)     perl: f($a, $b)

perl's line is the TOTAL slot count. Measured on 5.42.0 with
`perl -MO=Deparse,-p -e 'sub f (P) {} my ($a,$b); f $a, $b;'`:

    $       (f($a), $b)       unary
    ;$      (f($a), $b)       unary
    _       (f($a), $b)       unary
    *       (&f($a), $b)      unary
    \@      (&f($a), $b)      unary
    ;\@     (&f($a), $b)      unary
    $;$     f($a, $b)         list
    ;$$     f($a, $b)         list
    \@;$    &f($a, $b)        list

Found adding perl's interpreter-defined subs (01a0eaf2), where
`Internals::SvREADONLY` is `\[$%@];$` and `Internals::SvREADONLY $x, 1` must
take both arguments.

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance criteria

- [ ] A prototype with exactly one slot, mandatory or optional, is unary; more than one is a list (`go test ./internal/parse/ -run '^TestPrototypeArity$' -count=1 -v | grep -q '^--- PASS: TestPrototypeArity'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
