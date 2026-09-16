---
title: Declarations, and the prototype the lexer must scan
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a70f-0fe8-7d8e-9c2d-d9aa50bef115
- 01a0a75b-7c7b-78ca-b47f-40cd34c5c963
blocks:
- 01a0a75b-5e51-79f0-8af1-95c5aa378706
created: 2026-09-15T23:16:18.129308407Z
updated: 2026-09-16T03:13:21.044099804Z
sessions:
- start_sha: 5b936a5d9c9c023339bf9efbde2faf62e31f3c78
  end_sha: ""
  commits: 0
  started_at: 2026-09-16T03:13:21.044099804Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-16T03:13:21.044099804Z
---

Declarations: `my`/`our`/`local`/`state`, `sub`, `package` in both forms, and
the prototype the lexer must learn to scan.

## Prototype RECOGNITION is M1's, not M4's

M4 owns prototype resolution — deciding what a prototype does to a call. This
issue owns recognising one at a declaration, which is a lexer job perl does
with `scan_str` (spec §5.5.2: "the lexer produces the PROTOTYPE token by
calling scan_str... it scans a balanced `(...)` as an opaque string").

`internal/lexer` has none: `grep -i proto internal/lexer/*.go` is empty, so
`sub f ($$)` lexes `$$` as a variable today. Measured in the T2 corpus:
`comp/proto.t` alone has 22 prototype declarations across 11 T2 files.

`TestT2CoreParses` cannot reach 100% without this.

## The bodiless form

§0.13 rank 4, 7 corpus files: `sub NAME;` and `sub NAME (proto);`. §5.5.2
gives the grammar as `( Block | ";" )`. For a fresh parser this is just the
`;` alternative — the "statement-level leak" §0.13 describes was a
tree-sitter recovery artifact, not a language subtlety.

## Acceptance

- [ ] my, our, local and state parse with single and list targets (`go test ./internal/parse/ -run '^TestVariableDeclarations$' -count=1 -v | grep -q '^--- PASS: TestVariableDeclarations'`)
- [ ] A sub declaration parses with a block and with a bare `;` (`go test ./internal/parse/ -run '^TestSubDeclarationForms$' -count=1 -v | grep -q '^--- PASS: TestSubDeclarationForms'`)
- [ ] The lexer scans a prototype as one opaque token, so `$$` is not a variable (`go test ./internal/lexer/ -run '^TestPrototypeIsOpaque$' -count=1 -v | grep -q '^--- PASS: TestPrototypeIsOpaque'`)
- [ ] `package NAME;` and `package NAME { ... }` both parse (`go test ./internal/parse/ -run '^TestPackageForms$' -count=1 -v | grep -q '^--- PASS: TestPackageForms'`)
- [ ] All 22 prototype declarations in comp/proto.t lex without an error token (`go test ./internal/lexer/ -run '^TestProtoCorpusLexes$' -count=1 -v | grep -q '^--- PASS: TestProtoCorpusLexes'`)

Findings: §0.13 rank 4 (7 files). Spec: ch5 §5.5.2.
