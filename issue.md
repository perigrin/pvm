---
title: "Lexical sub declarations lose their name: our sub h; canons as our sub();h();"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-28T22:41:57.10531075Z
updated: 2026-10-11T15:17:44.565858741Z
sessions:
- start_sha: 468aa62746c7cda08884e4f3e2c6c593febca14f
  end_sha: 0e922405660ff6594b872dc84d304bd97ad26489
  commits: 1
  started_at: 2026-09-29T00:46:49.082416982Z
  ended_at: 2026-09-29T01:02:38.956824631Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T00:46:49.082416982Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T01:02:38.956824631Z
observed_paths:
- internal/conformance/conformance_test.go
- internal/parse/canon.go
- internal/parse/decl.go
- internal/parse/lexsub_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1fidelity.ratchet
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

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance criteria

- [ ] my, our and state sub declarations keep their name, prototype or signature, and body, and canon to a fixpoint (`go test ./internal/parse/ -run '^TestLexicalSubDeclaration$' -count=1 -v | grep -q '^--- PASS: TestLexicalSubDeclaration'`)
- [ ] REGRESSION GUARD: an anonymous `my $f = sub { ... }` is still an anonymous sub (`go test ./internal/parse/ -run '^TestLexicalSubDeclaration$' -count=1 -v | grep -q '^--- PASS: TestLexicalSubDeclaration'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
