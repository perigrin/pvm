---
title: The perl.git t/ harness does not search lib/ or the dist, cpan and ext module trees
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T17:30:08.680391488Z
updated: 2026-09-29T17:40:00.952736361Z
sessions:
- start_sha: cce0f97fdbad6745670e0dcf0d0c57d05099ad5b
  end_sha: 4aa05bb40fab4acab33c9adf4eaeb6384ccfa24c
  commits: 1
  started_at: 2026-09-29T17:30:10.546106927Z
  ended_at: 2026-09-29T17:40:00.952736361Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T17:30:10.546106927Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T17:40:00.952736361Z
observed_paths:
- internal/parse/perlgitt_test.go
- internal/parse/testdata/perlgitt.ratchet
---

The perl.git t/ harness resolves `use` against `t/` and the tree root only. perl's suite runs with test.pl's `set_up_inc('../lib')`, and a built perl's lib/ holds the dual-life modules the build copies from dist/*/lib, cpan/*/lib and ext/*/lib. An unbuilt checkout has them only at those source paths, so `use Carp` (dist/Carp/lib) resolves nothing and every `croak "..."` is an unknown word before a string.

perl.git t/porting/podcheck.t:660-677.

## Acceptance criteria

- [ ] The perl.git t/ harness searches lib/ and the dist, cpan and ext module trees (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
