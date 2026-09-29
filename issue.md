---
title: "`my method priv {...}` refuses: lexical methods are not read"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T16:37:07.142653531Z
updated: 2026-09-29T16:47:47.800302475Z
sessions:
- start_sha: 27a080e5f1a9c01275940f16c05bfa354af9d0d0
  end_sha: ae4b9b0e8875942ad5c851e8579d4cc18fc2db12
  commits: 1
  started_at: 2026-09-29T16:37:07.981068923Z
  ended_at: 2026-09-29T16:47:47.800302475Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T16:37:07.981068923Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T16:47:47.800302475Z
observed_paths:
- internal/parse/decl.go
- internal/parse/mymethod_test.go
- internal/parse/t2_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/use_test.go
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
