---
title: "`require mro, diag \"x\"` refuses: require is read only as a statement"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T19:16:18.729037698Z
updated: 2026-09-29T19:16:20.512994638Z
sessions:
- start_sha: 59ea632c8762f2c3df4f7f0b435dd2eb915e6fa1
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T19:16:20.512994638Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T19:16:20.512994638Z
---

`require MODULE` inside an expression refuses. The parser reads `require BAREWORD` only as a statement form, where anything after the name is an import list -- which `require` does not have -- so `... or require mro, diag "x"` and `(require Foo, 2)` fail. perl makes `require` a named unary whose operand is the bareword module name.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'sub diag {} ok($x) or require mro, diag "x"; my @a = (require Foo, 2);'
    require mro, diag('x') unless ok($x);
    my(@a) = (require Foo, 2);

perl.git t/op/universal.t:311.

## Acceptance criteria

- [ ] `require MODULE` parses inside an expression and canons to a fixpoint (`go test ./internal/parse/ -run '^TestRequireInExpression$' -count=1 -v | grep -q '^--- PASS: TestRequireInExpression'`)
