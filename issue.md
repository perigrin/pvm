---
title: A bare `&` with no name is accepted as a term
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T22:20:52.910326593Z
updated: 2026-09-29T22:21:15.136567513Z
---

A lone `&` in term position with no name after it parses as a term. `f 1, &= 2` (perl: "Number found where operator expected") canons as `f(1 , & = 2)`: the FuncSigil branch of parseTerm returns `Term{Text: "&"}` when ampTakes says no. perl has no reading of a bare `&` followed by `=` as a term.

Found while measuring the trailing-comma rule in parseInfix.

## Acceptance criteria

- [ ] `sub f {} f 1, &= 2;` refuses (a parse test asserting an Unknown, named in the fix)
