---
title: The perl.git t/ harness does not search lib/ or the dist, cpan and ext module trees
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T17:30:08.680391488Z
updated: 2026-09-29T17:30:08.680391488Z
---

The perl.git t/ harness resolves `use` against `t/` and the tree root only. perl's suite runs with test.pl's `set_up_inc('../lib')`, and a built perl's lib/ holds the dual-life modules the build copies from dist/*/lib, cpan/*/lib and ext/*/lib. An unbuilt checkout has them only at those source paths, so `use Carp` (dist/Carp/lib) resolves nothing and every `croak "..."` is an unknown word before a string.

perl.git t/porting/podcheck.t:660-677.

## Acceptance criteria

- [ ] The perl.git t/ harness searches lib/ and the dist, cpan and ext module trees (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
