---
title: A declaration on the right of a tighter operator swallows what follows it
state: done
urgency: normal
milestone: v0.1
created: 2026-10-23T20:36:25.460408142Z
updated: 2026-10-26T00:27:49.405317723Z
sessions:
- start_sha: 3b77ae24bd593b2eb20e829fdc36fd5135584b9a
  end_sha: 3b77ae24bd593b2eb20e829fdc36fd5135584b9a
  commits: 0
  started_at: 2026-10-26T00:16:33.734740562Z
  ended_at: 2026-10-26T00:27:49.405317723Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-26T00:16:33.734740562Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-26T00:27:49.405317723Z
---

Found in the independent review of 01a16602-5ebe (exec/5ebe, f80cbfbf), 2026-10-23, and measured with perl 5.44.0 -MO=Deparse,-p.

## Context

- paths: `internal/parse/decl.go`, `internal/parse/testdata/m3probes/dropped-paren-context.txt`
- entrypoints: `decl.go`'s declaration parse, which ends `return p.parseInfix(n, infix[","].BP)`

A declaration on the right of a tighter operator swallows what follows it. The declaration parse resumes infix parsing at the comma's binding power and ignores the caller's.

- `my $y = 2 * my ($x) + 1;` Perl reads `((2 * my $x) + 1)`. Before 5ebe psc read `2*my(($x)+1)`; after it, psc reads `2*(my($x)+1)`. Both trees are wrong, and the second is also unfaithful.
- `2 * my $x + 1` has the same defect, on the scalar path, and it predates 5ebe.

No perl.git t/ or T1 file is known to hit this. 5ebe's probe file carries the first snippet as a known-defect row naming this issue.

## Acceptance criteria

### Positive Scenarios

- [ ] This issue's probe rows lose their known-defect mark (`go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'`)
