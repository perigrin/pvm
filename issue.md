---
title: "`... and last BIN` refuses: a loop control in an expression takes no label"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T17:25:11.566081516Z
updated: 2026-09-30T17:25:11.566081516Z
---

`$path = $bin and last BIN if $bin;` refuses. At a statement's start, `last`/`next`/`redo` take an optional label. Inside an expression they reach the niladic-builtin path, which takes nothing, so `BIN` is stranded, and a bare one canons as `last()`. perly.y's LOOPEX takes the label in both positions.

Measured with perl 5.42 and -MO=Deparse: `$p = $b and last B if $b;` deparses as written; `$q = 1 or next B;` is `next B unless $q = 1;`; `$r ||= redo B;` is `$r ||= (redo B);`.

PerlOnJava unit/loop_label_bareword_constant.t:13. In perl.git t/op/loopctl.t, `for (...; 1; last TEST16)` canoned as `last() TEST16()`, a wrong tree.

## Acceptance criteria

- [ ] Loop controls in an expression take their label (`go test ./internal/parse/ -run '^TestLoopControlInExpression$' -count=1 -v | grep -q '^--- PASS: TestLoopControlInExpression'`)
