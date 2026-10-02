---
title: Decide how an operator declaration states its result rule and fixity
state: in-progress
urgency: normal
milestone: type-libraries
blocks:
- 01a0fa47-146b-7047-84b6-20ac55c1b0f6
created: 2026-10-02T01:42:03.565080543Z
updated: 2026-10-02T16:04:41.338749028Z
sessions:
- start_sha: 67d3a077587718198731eaaf6b7174ae3c0f7706
  end_sha: ""
  commits: 0
  started_at: 2026-10-02T16:04:41.338749028Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-02T16:04:41.338749028Z
---

RFC 0001 agrees operators are declared as `sub + :infix (Num $x, Num $y) Num;` and says nothing past the name. Three things are open:

- **The result rule.** `+` yields a result bounded by its operands (`Int + Int` is `Int`), `/` a fixed one (`Num`). bson's TypeLibrary marks the first kind with `%RESULT_IS_JOIN`; pvm's `internal/types/signatures.go` does not distinguish them, so here `Int + Int` is `Num`. A declaration needs a way to say which.
- **Fixity and precedence.** `:infix`/`:prefix`/`:postfix`, and whether a declaration carries binding strength or the parser's precedence table stays the authority.
- **Operators that fork.** `x` repeats a list when its left side is parenthesised and a string otherwise; `..` is a range in list context and a flip-flop in scalar (`:context` fits it).

perigrin decides; the decision lands in RFC 0001 as a section `### Operator declarations (*Decided*)`.

## Acceptance criteria

- [ ] The RFC records the decision (`grep -qF '### Operator declarations (*Decided*' docs/rfcs/0001-type-libraries.md`)
