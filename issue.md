---
title: A glob dereference *{EXPR} is a Term, though every other dereference is a Unary
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T18:04:41.827533524Z
updated: 2026-10-09T18:04:41.827533524Z
---

## Context

- paths: `internal/parse/term.go` (glob dereference), `internal/parse/deref_test.go` ("a dereference is a Unary"), `internal/parse/wrongtree_test.go` (TestWrongTreesAreGone deliberately leaves the glob node's kind unchecked)

Seen 2026-10-09 writing TestWrongTreesAreGone: `*{EXPR}`, `*$x` and `*{$name}` build a Term with Text `*` and one child, while `${EXPR}` and `@{EXPR}` build a Unary -- and deref_test.go states a dereference is a Unary. perl compiles all of them as a dereference op (rv2gv, rv2sv, rv2av). Either the glob dereference should be a Unary like the others, or the difference is deliberate and should be said where deref_test.go says the rule.

## Steps

- [ ] Decide which (read why term.go builds a Term); failing test first if it changes; canon and Sites unaffected or toward perl; `make test`; commit

## Acceptance criteria

- [ ] a glob dereference has the node kind the other dereferences have (`go test ./internal/parse/ -run '^TestGlobDereferenceIsADereference$' -count=1 -v | grep -q '^--- PASS: TestGlobDereferenceIsADereference'`)
