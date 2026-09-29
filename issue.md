---
title: "Lexical sub declarations lose their name: our sub h; canons as our sub();h();"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-28T22:41:57.10531075Z
updated: 2026-09-29T00:46:49.082416982Z
sessions:
- start_sha: 468aa62746c7cda08884e4f3e2c6c593febca14f
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T00:46:49.082416982Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T00:46:49.082416982Z
---

Lexical sub declarations -- `my sub`, `our sub`, `state sub` -- are read as a
declarator followed by an ANONYMOUS sub, so the name falls out as a call:

    our sub h;          canon  our sub();h();          Unknown=0, WRONG tree
    my sub e ($);       canon  my sub();e ($);         Unknown=1

All valid on 5.42.0 (`perl -c` says syntax OK; lexical subs are stable since
5.26 and need no feature). The declarator path hands `sub` to the anonymous-sub
term, which takes no name.

The plan named this cause ("lexical sub with a signature or prototype, 6 files,
op/lexsub.t alone 48 nodes") without filing it. The canon case shows the
unprototyped form is wrong too, at zero Unknowns, so the reach is wider than
the count.

## Where the fixtures point here

`internal/conformance/conformance_test.go`'s `TestCorpusSkipsAreDocumented` and
`TestRefusalCodeMismatchFails` need a construct that still refuses, and use
`my sub e ($);` naming THIS issue. When it lands both fail with "no longer
refuses; replace this fixture", which is the harness working.

## Acceptance criteria

- [ ] my, our and state sub declarations keep their name, prototype or signature, and body, and canon to a fixpoint (`go test ./internal/parse/ -run '^TestLexicalSubDeclaration$' -count=1 -v | grep -q '^--- PASS: TestLexicalSubDeclaration'`)
- [ ] REGRESSION GUARD: an anonymous `my $f = sub { ... }` is still an anonymous sub (`go test ./internal/parse/ -run '^TestLexicalSubDeclaration$' -count=1 -v | grep -q '^--- PASS: TestLexicalSubDeclaration'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
