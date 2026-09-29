---
title: "`*]` is not a glob: the `]` closes whatever bracket is open"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T12:56:42.966415699Z
updated: 2026-09-29T12:56:43.598423126Z
sessions:
- start_sha: db9808e688f56c2993d8a528c48cd171180e3067
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T12:56:43.598423126Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T12:56:43.598423126Z
---

After a glob-sigil star toke.c's scan_ident takes one punctuation byte as the name, `]` among them. The lexer emits `]` as a closer, which pops whatever bracket is open, so `tie my $var4 => 'main', *];` refuses and throws the bracket tracking off for what follows.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'my $g = *]; my @a = (1, *], 2);'
    my $g = *];
    my(@a) = (1, *], 2);

perl.git t/op/tie_fetch_count.t:189.

## Acceptance criteria

- [ ] `*]` is a glob and its canon is a fixpoint (`go test ./internal/parse/ -run '^TestGlobCloseBracket$' -count=1 -v | grep -q '^--- PASS: TestGlobCloseBracket'`)
