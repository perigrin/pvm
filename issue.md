---
title: Deref braces around a quote word, or a subscript that stops short, read wrongly
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-24T14:27:59.984199171Z
updated: 2026-10-24T14:27:59.984199171Z
---

Found by the independent review of 01a16e20-f266 (exec/univar, 5fd34251) on 2026-10-24. These defects predate that branch. They are rv2hv WRONG verdicts outside the corpus.

## Context

- paths: `internal/parse/term.go` (bracedSubscriptedName), `internal/lexer/` (quote words inside deref braces)
- entrypoints: `${`, `@{`, `%{` or `*{` followed by a word, then `{` or `[`

01a16e20 taught the parser that perl reads `@{ foo {"a"} }` as the hash slice `@foo{"a"}`. Three related shapes are still wrong, and each one makes psc decide an rv2hv count that perl does not.

1. **Quote words.** perl reads `@{ q {a} }` as `@q{'a'}`, and the same holds for `qq`, `qw`, `m`, `s`, `y` and `tr`. psc lexes these as quotes. CompareFacts scores WRONG for q, qq, qw and m; for s, y and tr psc's tree is Unknown.
2. **A subscript that does not run to the closing brace.** perl reads `@{ foo {"a"}; 1 }` as `@foo{'a'}; 1;`, and `@{ foo {"a"} + 1 }` as `@foo{'a'} + 1`. psc keeps the deref. Both are rv2hv WRONG, and Faithful says true for both.
3. **`Foo'bar`, the old package separator.** It is WRONG at base too.

Before writing any code, measure each one with perl 5.44 using Concise and Deparse,-p.

## Acceptance criteria

### Positive Scenarios

- `go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'` (probe rows for `@{ q {a} }`, `@{ qw {a} }`, `@{ m {a} }`, `@{ foo {"a"}; 1 }`, `@{ foo {"a"} + 1 }` and a `Foo'bar` form, each held to perl's hslice and rv2hv count)
