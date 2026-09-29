---
title: "`my method priv {...}` refuses: lexical methods are not read"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T16:37:07.142653531Z
updated: 2026-09-29T16:37:07.981068923Z
sessions:
- start_sha: 27a080e5f1a9c01275940f16c05bfa354af9d0d0
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T16:37:07.981068923Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T16:37:07.981068923Z
---

`my method NAME {...}`, 5.42's lexical method, refuses: parseLexicalSub reads `my sub NAME` only.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'use feature "class"; no warnings;
          class X { field $f = 1; my method priv { return $f } method pub { 1 } }'
        my method priv {
            return $f;
        }

perl.git t/class/method.t:104, 120 and 153.

## Acceptance criteria

- [ ] `my method NAME` declares under class syntax (`go test ./internal/parse/ -run '^TestLexicalMethod$' -count=1 -v | grep -q '^--- PASS: TestLexicalMethod'`)
