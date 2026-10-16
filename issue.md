---
title: "[R1] The review fixes add 75 lines to psc check over perl's lib"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T12:09:13.711338306Z
updated: 2026-10-16T12:09:17.185041672Z
sessions:
- start_sha: 77107dfbeb77e7b3f1d94437583eae6d28eb41c3
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T12:09:17.185041672Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T12:09:17.185041672Z
---

## Context

- found by: re-measuring psc check over perl 5.42's lib after the PR #486 review fixes merged (77107dfb), 2026-10-16
- evidence: `scratchpad/sweep/{old,new,added,removed}.diag` -- old = 8a0c85c6 (before the review fixes), new = 77107dfb
- commands: `PVM_PERL_VERSION=5.42.0 psc check <perl 5.42 privlib>`; perl 5.42 `-w`

perigrin's bar for this PR is that psc check output only improves. Before the review fixes, perl 5.42's lib drew 8 lines; after them, 79: 75 added, 4 removed. The added lines cluster: `ref` argument 1 given a near-total union (7), `waitpid`/`read`/`syswrite` given Undef|Int, `eq`/`==`/`+`/`>` operands, `chomp` given List, `open`/`fileno`/`close` given Str, `chdir`, `lc`, `substr`. One removed line, JSON/PP.pm:918, was a true positive (perl warns on truncated input).

## Steps

- [ ] Bucket every added line and the removed JSON/PP.pm:918 by root cause, with a minimal repro and perl -w evidence each; attribute each to the review merge that introduced it
- [ ] Fix each false-positive bucket test-first, keeping a control; restore the JSON/PP true positive or show perl does not warn there
- [ ] Re-sweep: no added line remains that is not a true positive perl warns on

## Acceptance criteria

- [ ] psc check over perl 5.42's lib adds nothing over the pre-review baseline but true positives (`go test ./internal/psc/ -run '^TestReviewFixesAddNoFalsePositives$' -count=1 -v | grep -q '^--- PASS: TestReviewFixesAddNoFalsePositives'`)
