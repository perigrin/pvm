---
title: "`map ({...} LIST)` with a space before the paren reads a hash, not a block"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T13:23:31.555153953Z
updated: 2026-09-29T13:31:58.867483899Z
sessions:
- start_sha: d38a84e916690dea34c5f5d3a8816b336d4690c7
  end_sha: 526d95e0d62a9397cf0ccf3b36b0c00e1d193265
  commits: 1
  started_at: 2026-09-29T13:23:32.293909833Z
  ended_at: 2026-09-29T13:31:58.867483899Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T13:23:32.293909833Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T13:31:58.867483899Z
observed_paths:
- internal/lexer/lexer.go
- internal/parse/listopspacedparen_test.go
- internal/parse/t2_test.go
- internal/parse/testdata/perlgitt.ratchet
---

The lexer carries a block-taking word's brace lookahead into its `(` -- `map({...} @a)` -- through a one-token flag, and the whitespace token between the word and a spaced paren reset it. So `map ({"$_\n"} @x)` read its braces as an anonymous hash and refused.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'my @e = map ({"$_\n"} @x); my @g = grep ({ $_ } @y);'
    my(@e) = map({"$_\n";} @x);
    my(@g) = grep({$_;} @y);

perl.git t/comp/retainedlines.t:55.

## Acceptance criteria

- [ ] A spaced paren after map, grep or sort keeps its block (`go test ./internal/parse/ -run '^TestListOpSpacedParenBlock$' -count=1 -v | grep -q '^--- PASS: TestListOpSpacedParenBlock'`)
