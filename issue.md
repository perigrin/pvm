---
title: "`blessed $obj` after `use builtin qw(blessed)` refuses: builtin:: functions are unknown to the parser"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T07:47:15.219965373Z
updated: 2026-09-29T08:00:39.962510293Z
sessions:
- start_sha: a57a36bf42137037fa1a23668d43b9d0d1777d38
  end_sha: a0dcdb8fa37d37e59dc0b241cb1a2fe1bbc97bae
  commits: 1
  started_at: 2026-09-29T07:49:21.399037887Z
  ended_at: 2026-09-29T08:00:39.962510293Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T07:49:21.399037887Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T08:00:39.962510293Z
observed_paths:
- internal/parse/builtinimport_test.go
- internal/parse/shape.go
- internal/parse/t2_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/use.go
- internal/parse/use_test.go
---

The builtin:: functions are the interpreter's own, registered by builtin.c's boot_core_builtin at startup, and `use builtin LIST` imports them by name. The parser knew neither, so an imported `blessed $obj` was `WORD $var` with an unknown WORD and refused (01a0ebb0-6329-7a13-b398-a5120ec9bbe8's deliberate half). A qualified `builtin::reftype $r, 1` took no argument at all.

Prototypes follow the call checker in boot_core_builtin: ck_builtin_const `""`, ck_builtin_func1 `"$"`, ck_builtin_funcN `"@"`, and none for export_lexically.

Measured with perl 5.42:

    $ perl -MO=Deparse,-p -e 'use builtin qw(blessed reftype);
          is(blessed $obj, "T", "x"); my $w = builtin::reftype $r, 1;
          use builtin "true"; my @a = (true, 1);'
    is(scalar(builtin::blessed($obj)), 'T', 'x');
    ((my($w) = builtin::reftype($r)), '???');
    (my(@a) = ((!0), 1));

perl.git t/class/construct.t:29-30 refuses on `blessed $obj` and `reftype $obj` after `use builtin qw( blessed reftype );`.

## Acceptance criteria

- [ ] builtin:: functions resolve with builtin.c's prototypes, qualified and imported (`go test ./internal/parse/ -run '^TestBuiltinImportsHaveShapes$' -count=1 -v | grep -q '^--- PASS: TestBuiltinImportsHaveShapes'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
