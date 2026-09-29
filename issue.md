---
title: A signature placeholder `$` before `)` or `,` lexes as the variable `$)` or `$,`
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T08:05:59.823676087Z
updated: 2026-09-29T08:36:26.936156172Z
sessions:
- start_sha: cbb11b367c3b0b91e97a6f5765e7f983549520bf
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T08:36:26.936156172Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T08:36:26.936156172Z
---

In a signature, `$` alone is a placeholder parameter, but the lexer reads `$)` as the effective-gid variable and `$,` as the output field separator. So `sub t010 ($a, $) { ... }` lost its closing paren to a variable and the signature swallowed the body: at Unknown=0 the tree canonicalised as `sub t010 = ($a, $){$a || "z"});`. Since trailing tokens after a declaration refuse (01a0ec21-837a-73e2-9f6d-dba44e242bd3), it refuses instead.

Measured with perl 5.42, each deparses unchanged:

    $ perl -MO=Deparse -e 'use feature "signatures";
          sub t010 ($a, $) { $a } sub t011 ($, $a) { $a } sub t013 ($) { 1 }'

perl.git t/op/signatures.t:129, 143, 157, 171. toke.c reads a signature under its own lexer state (`PL_parser->sig_seen`, yyl_sigvar), where `$` before `,`, `)` or `=` is a sigil with no name.

## Acceptance criteria

- [ ] A bare `$` placeholder before `)` or `,` in a signature is a sigil, not a punctuation variable (`go test ./internal/parse/ -run '^TestSignaturePlaceholderSigil$' -count=1 -v | grep -q '^--- PASS: TestSignaturePlaceholderSigil'`)
