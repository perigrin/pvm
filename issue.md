---
title: "perl's own XSUBs are undeclared: utf8::encode $x refuses though perl defines it at startup"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T02:16:07.059364701Z
updated: 2026-09-29T02:47:23.42073157Z
sessions:
- start_sha: f9c96510c1142c37237b7b0cf0013b9a12ac0bf1
  end_sha: ae2fb4786754099203add719feedfd96fedd7208
  commits: 1
  started_at: 2026-09-29T02:35:24.624046671Z
  ended_at: 2026-09-29T02:47:23.42073157Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T02:35:24.624046671Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T02:47:23.42073157Z
observed_paths:
- internal/parse/coresubs_test.go
- internal/parse/shape.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
---

perl defines a table of subs itself, at interpreter start, before any file is
compiled -- universal.c's `these_details[]`: `utf8::encode` and the rest of
`utf8::`, `Internals::SvREADONLY`, `PerlIO::get_layers`, `re::regexp_pattern`,
`UNIVERSAL::isa`, `constant::_make_const`, `Tie::Hash::NamedCapture::*`. A
parenless call to one resolves with nothing loaded. Measured on 5.42.0:

    $ perl -e 'my $s = "\x{100}"; utf8::encode $s; print length $s'
    2

The parser declares none of them, so a parenless call with an argument refuses:

    utf8::encode $x;               Unknown=1

12 perl.git t/ files first refuse at one (`utf8::encode`, `utf8::upgrade`,
`Internals::SvREADONLY`, `PerlIO::get_layers`).

The prototypes come from the same table: NULL is no prototype, a list
operator, and `Internals::SvREADONLY` is `\[$%@];$`. `mro::get_isarev` is NOT
in it -- it is ext/mro's XS, loaded by `use mro` -- and stays out.

## Acceptance criteria

- [ ] A parenless call to an interpreter-defined sub resolves with universal.c's prototype, and canons as a call (`go test ./internal/parse/ -run '^TestInterpreterDefinedSubs$' -count=1 -v | grep -q '^--- PASS: TestInterpreterDefinedSubs'`)
- [ ] NEGATIVE: an arbitrary qualified name is still unknown (`go test ./internal/parse/ -run '^TestInterpreterDefinedSubs$' -count=1 -v | grep -q '^--- PASS: TestInterpreterDefinedSubs'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
