---
title: "`my sub s` does not shadow `s///`: `s(1)` lexes as a substitution"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T15:35:58.091676131Z
updated: 2026-09-29T15:35:58.69100902Z
sessions:
- start_sha: 59c71e8af0366ea66f124dc3c4c574467e1e02b8
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T15:35:58.69100902Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T15:35:58.69100902Z
---

A lexical sub named like a quote operator does not shadow it: after `my sub s ($arg) { ... }`, the call `s(1);` lexes as a substitution that runs on through the file. perl looks the word up among the lexical subs in scope before taking it as a quote operator, and the shadowing ends with the block.

Measured with perl 5.42:

    $ perl -e '{ my sub s { 42 } print s(1), "\n"; } $_ = "a"; s/a/b/; print "$_\n";
          { my sub y { 43 } print y, "\n"; }'
    42
    b
    43

perl.git t/op/lvref.t:653 and t/op/lexsub.t:869.

## Acceptance criteria

- [ ] A lexical sub shadows the quote operator of its name within its block only (`go test ./internal/lexer/ -run '^TestLexicalSubShadowsQuoteOp$' -count=1 -v | grep -q '^--- PASS: TestLexicalSubShadowsQuoteOp'`)
