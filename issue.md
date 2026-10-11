---
title: A bare `&` with no name is accepted as a term
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T22:20:52.910326593Z
updated: 2026-10-11T02:35:36.545103195Z
sessions:
- start_sha: 24bfac4a5293052e7c38dd54a56a58aef818e639
  end_sha: e18b59c82f89ca3a59df5a4be37757a6ea0d15f5
  commits: 2
  started_at: 2026-10-11T00:57:42.810624868Z
  ended_at: 2026-10-11T02:35:36.545103195Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T00:57:42.810624868Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-11T02:35:36.545103195Z
observed_paths:
- internal/lexer/aposfeature_test.go
- internal/lexer/ident.go
- internal/lexer/lexer.go
- internal/lexer/pod.go
- internal/lexer/proto_test.go
- internal/lexer/scan.go
- internal/parse/arrowcodederef_test.go
- internal/parse/bareampersand_test.go
- internal/parse/builtinimport_test.go
- internal/parse/canon.go
- internal/parse/parse.go
- internal/parse/term.go
- internal/parse/testdata/t1canon.ratchet
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/use.go
---

A lone `&` in term position with no name after it parses as a term. `f 1, &= 2` (perl: "Number found where operator expected") canons as `f(1 , & = 2)`: the FuncSigil branch of parseTerm returns `Term{Text: "&"}` when ampTakes says no. perl has no reading of a bare `&` followed by `=` as a term.

Found while measuring the trailing-comma rule in parseInfix.

### Fixed (2026-10-11)

perl has no bare `&`: its tokenizer takes the next single punctuation
character as the sub's name (as `$;` and `$=` are names), so `&= 2` is a call
to a sub named `=` followed by a term, "Number found where operator
expected". Measured on 5.42: `&;`, `&=;` and `&$x` compile; `&= 2`, `f(&)`,
`&` at EOF and `&&` are refused. Fixed in 74e5bb88 through the existing
refusal sites; the criterion had no command and now names the test.

## Acceptance criteria

- [ ] a bare `&` with no name is refused where perl refuses it, and every valid `&` form still parses (`go test ./internal/parse/ -run '^TestBareAmpersandIsRefused$' -count=1 -v | grep -q '^--- PASS: TestBareAmpersandIsRefused'`)
- [ ] no refusal code or site was added unrecorded (`go test ./internal/parse/ -run '^TestRefusalCodeInventory$' -count=1 -v | grep -q '^--- PASS: TestRefusalCodeInventory'`)
