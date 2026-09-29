---
title: "$$ swallows the byte after it: f($$, 1) lexes the comma into the variable"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T04:51:36.703917912Z
updated: 2026-09-29T04:51:37.177887211Z
sessions:
- start_sha: 7e3713d302ede624a4828ec22a3ed861f5ffc651
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T04:51:37.177887211Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T04:51:37.177887211Z
---

`$$` -- the process id -- followed by a comma, a terminator, a closer or even
a space lexes as ONE variable with that byte swallowed:

    f($$, 1);          tokens  Variable "$$,"  ...        Unknown=1
    kill 9, $$;        tokens  ... Variable "$$;"         the `;` is gone
    print $$ . "\n";   tokens  Variable "$$ "

scanVarName treated every second `$` as a dereference and recursed, and the
recursion took the next byte as a one-byte punctuation name. toke.c's
scan_ident takes the dereference only when a name, digit, `$`, `{` or `::`
follows:

    if (*s == '$' && s[1]
        && (   isIDFIRST_lazy_if_safe(s+1, PL_bufend, is_utf8)
            || isDIGIT_A((U8)s[1]) || s[1] == '$' || s[1] == '{'
            || memBEGINs(s+1, ..., "::")) )

Measured on 5.42.0, `my @a = ($$, 1)` has two elements.

Reach: op/sigdispatch.t, op/magic.t, op/taint.t, op/getppid.t,
op/signame_canonical.t, lib/cygwin.t and more -- every test that passes `$$`
to kill or compares it.

## Acceptance criteria

- [ ] `$$` before a comma, terminator, closer or space is the pid variable alone, and `$$ref`, `$$1`, `$$::x` still dereference (`go test ./internal/lexer/ -run '^TestPidVariableEndsAtItsSecondSigil$' -count=1 -v | grep -q '^--- PASS: TestPidVariableEndsAtItsSecondSigil'`)
- [ ] REGRESSION GUARD: the lexer's golden token stream is unchanged (`go test ./internal/lexer/ -count=1`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
