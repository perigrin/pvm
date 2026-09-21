---
title: A pinned empty output is indistinguishable from no pinned output
state: done
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T13:58:25.096391623Z
updated: 2026-09-21T22:36:19.679050382Z
sessions:
- start_sha: 9102578c35f74c5afa0bb9ec81e2bb8098a3221e
  end_sha: 6036de4e5427d522f018fc1af4fd59dc3d760c06
  commits: 1
  started_at: 2026-09-21T19:27:49.087584502Z
  ended_at: 2026-09-21T19:49:18.617038354Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T19:27:49.087584502Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T19:49:18.617038354Z
observed_paths:
- conformance/README.md
- internal/conformance/conformance_test.go
- internal/conformance/file.go
- internal/conformance/lint_test.go
- internal/conformance/run.go
- internal/parse/chain.go
- internal/parse/expr.go
- internal/parse/parse.go
- internal/parse/refusal.go
- internal/parse/refusal_test.go
- internal/parse/term.go
---

The same representability bug the format issue fixed for parses/parsent,
surviving in a sibling field.

`internal/conformance/run.go:47`:

    if f.ExpectOutput != "" && output != f.ExpectOutput {
        t.Fatalf("CORPUS BUG: pinned output %q, perl prints %q", ...)
    }

A file that pins genuinely EMPTY output is indistinguishable from a file
with no `--- expect output` section at all, because both produce
`ExpectOutput == ""`. The check is skipped, so perl could print anything
and the corpus would not notice.

## Why this matters rather than being theoretical

The README states the contract as "what perl prints, byte for byte". A
corpus file whose whole point is that a construct prints NOTHING cannot
currently express that -- and would appear to pass while asserting nothing.
Tier 10_io is where such files will first be wanted, since a filehandle
case that prints nothing on success is a natural thing to write.

The parses/parsent fix rejected a contradiction rather than resolving it by
field order. This is the same shape: a state the FORMAT can express that the
STRUCT cannot represent.

## The fix is a presence bit, not a sentinel

`ExpectOutput string` needs to become distinguishable-when-empty. A
`*string`, or a paired `HasExpectOutput bool`, or a small type -- whichever
reads best next to the four other fields. Do not use a sentinel string; a
corpus file could legitimately contain it.

Found by the architecture review gate on 1e70de78.

## Acceptance criteria

- [ ] A corpus file pinning empty output has its output checked rather than skipped (`go test ./internal/conformance/ -run '^TestEmptyExpectedOutputIsChecked$' -count=1 -v | grep -q '^--- PASS: TestEmptyExpectedOutputIsChecked'`)
- [ ] A file with no expect-output section still skips the check, which is the normal parsent case (`go test ./internal/conformance/ -run '^TestAbsentExpectedOutputSkipsCheck$' -count=1 -v | grep -q '^--- PASS: TestAbsentExpectedOutputSkipsCheck'`)
- [ ] The two cases are distinguishable in the parsed File without a sentinel string (`go test ./internal/conformance/ -run '^TestParseFile' -count=1 -v | grep -q '^--- PASS: TestParseFileSections'`)
- [ ] The whole corpus still runs green (`go test ./internal/conformance/ -count=1`)
