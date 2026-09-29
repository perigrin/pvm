---
title: "`method Pack (\"a\")` is read as a method declaration: method is a keyword only under class syntax"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T16:25:21.950693881Z
updated: 2026-09-29T16:34:44.36188626Z
sessions:
- start_sha: 238bba9e7e6e81f19c86aa070aaf3e2230bccb92
  end_sha: 27a080e5f1a9c01275940f16c05bfa354af9d0d0
  commits: 1
  started_at: 2026-09-29T16:25:23.647064322Z
  ended_at: 2026-09-29T16:34:44.36188626Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T16:25:23.647064322Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T16:34:44.36188626Z
observed_paths:
- internal/lexer/expect_test.go
- internal/lexer/ident.go
- internal/lexer/lexer.go
- internal/lexer/proto.go
- internal/parse/decl.go
- internal/parse/methodgate_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/use.go
---

The lexer and the parser both read `method` as the class feature's declarator everywhere. Without class syntax it is an ordinary name: op/method.t declares `sub Pack::method` and calls it through indirect object syntax, `method Pack ("a")`, which became a method declaration named Pack with `("a")` as its prototype.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'sub Pack::method { "m" } my $x = method Pack ("a");'
    my $x = 'Pack'->method('a');

A second gap stood behind it: `sub Pack::method` makes Pack a package, but only `package`, `class`, `use` and `require` recorded packages, so the indirect reading could not name it.

perl.git t/op/method.t:59-67.

## Acceptance criteria

- [ ] `method` declares only with class syntax in effect, and `method Pack (...)` is an indirect call (`go test ./internal/parse/ -run '^TestMethodKeywordNeedsClassFeature$' -count=1 -v | grep -q '^--- PASS: TestMethodKeywordNeedsClassFeature'`)
- [ ] The lexer agrees (`go test ./internal/lexer/ -run '^TestMethodNeedsClassSyntax$' -count=1 -v | grep -q '^--- PASS: TestMethodNeedsClassSyntax'`)
