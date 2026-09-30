---
title: "`f 1, ->m` refuses: `->` does not end a list after a trailing comma"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T19:54:02.772041056Z
updated: 2026-09-30T19:54:02.772041056Z
---

`create threads sub {...}=>->join->()()` refuses. After a trailing comma or fat comma, `->` doesn't end the list, because the rule excluded every operator starting with `-`. That exclusion came from `f 1, -> 2`, which is invalid only because of the `2`.

Measured with perl 5.42: `f 1, ->m` is `f(1)->m`; `f 1 => ->m` is the same; `f 1, -> [0]` is `f(1)->[0]`. A plain `-` still begins a term: `f 1, - 2` is `f 1, -2`.

PerlOnJava unit/threads_postfix_create_and_invalid_entry.t:13 and unit/print_indirect_method_postfix.t:16.

## Acceptance criteria

- [ ] `->` after a trailing comma ends the list (`go test ./internal/parse/ -run '^TestTrailingCommaBeforeOperator$' -count=1 -v | grep -q '^--- PASS: TestTrailingCommaBeforeOperator'`)
