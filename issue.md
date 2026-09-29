---
title: "print $x if $c takes $x as a filehandle: a statement modifier is read as a term"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T05:59:34.016442834Z
updated: 2026-09-29T05:59:35.25354043Z
sessions:
- start_sha: 9c48146ef630f3df38870b9dc5d83c4ecb775b6b
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T05:59:35.25354043Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T05:59:35.25354043Z
---

`print $x if $c` refuses: the scalar is taken as print's filehandle, and the
modifier is left with nothing to attach to.

    print $warn if length $warn;     Unknown=1     comp/redef.t:86
    print $w unless $q;              Unknown=1
    print $w foreach @a;             Unknown=1

The handle slot takes the scalar when the token after it starts a term, and
startsTerm answered yes for every word that is not an infix operator -- the six
statement modifiers included. Measured on 5.42.0 with -MO=Deparse,
`print $warn if length $warn;` is kept as written: $warn is printed.

stmtmodcanon_test.go pinned `print $_ foreach @a;` to canon verbatim, which is
the echo of an Unknown; that test asserts only the canon and so matched the
refusal.

## Acceptance criteria

- [ ] A scalar after print followed by a modifier is printed, not a handle, and a handle before a real term is kept (`go test ./internal/parse/ -run '^TestPrintScalarBeforeModifier$' -count=1 -v | grep -q '^--- PASS: TestPrintScalarBeforeModifier'`)
- [ ] REGRESSION GUARD: statement-modifier canon holds (`go test ./internal/parse/ -run '^TestStatementModifierKeepsItsBody$' -count=1 -v | grep -q '^--- PASS: TestStatementModifierKeepsItsBody'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
