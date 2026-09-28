---
title: "A heredoc body inside its statement stops it: fresh_perl_is(<<~EOF, ...) loses every argument after the body"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-28T22:03:29.769297285Z
updated: 2026-09-28T22:54:45.938771322Z
sessions:
- start_sha: 102569cdaab8154c20c34fde9f4e09e2bf2d4a65
  end_sha: ""
  commits: 0
  started_at: 2026-09-28T22:54:45.938771322Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-28T22:54:45.938771322Z
---

A heredoc body that falls INSIDE its statement stops the statement. perl reads
a body from the line after its opener and then resumes the statement where the
line broke off, so a call whose arguments continue past the body is ordinary
perl -- and it is the house style of perl's own fresh-perl helpers:

    fresh_perl_is(<<~'EOF',
        ...program...
        EOF
        "expected", "name");

Measured on 5.42.0:

    $ printf 'sub f { print join("|", @_), "\n" }\nf(<<EOF,\nbody\nEOF\n 1);\n' | perl
    body
    |1

The parser took bodies only after a statement's terminator, so the argument
list met the body token mid-expression and refused. run/todo.t carried a dozen.

The fix passes over a body while peeking and records the bodies a statement
passed over on the Statement itself, not as children: SourceText walks
children in position order and those bytes are already inside the call's span.

## Acceptance criteria

- [ ] A body inside its statement parses, round-trips, and canons to a fixpoint -- including one nested in an inner statement, which the outer must not claim twice (`go test ./internal/parse/ -run '^TestHeredocBodyInsideItsStatement$' -count=1 -v | grep -q '^--- PASS: TestHeredocBodyInsideItsStatement'`)
- [ ] REGRESSION GUARD: the existing heredoc placements still hold (`go test ./internal/parse/ -run '^TestHeredocInStatementPosition$' -count=1 -v | grep -q '^--- PASS: TestHeredocInStatementPosition'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
