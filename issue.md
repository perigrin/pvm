---
title: "Test & Build (Windows) fails: CRLF corpus files and quoted temp paths"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T04:03:32.575397266Z
updated: 2026-10-16T18:48:32.666257344Z
---

## Context

- paths: `.gitattributes` (absent?), `internal/conformance`, `internal/parse` (`cvwalk_test.go`, `parenless_test.go`, `ratchet_test.go`), `internal/parseoracle` (`scratch_test.go`, `run_test.go`), `internal/infer`
- commands: CI `Test & Build (Windows)`; locally only by reasoning about CRLF and path quoting

`Test & Build (Windows)` has failed on pu since at least 2026-09-04 (102 failures on cdce17ba) and on PR #483 (96 on 98c8a364, run 37411297109). Linux and macOS pass on PR #483 once the pinned perl is installed. The Windows failures are portability bugs, grouped by the log:

1. CRLF in checked-out corpus files: `parses must be yes or no, got "yes\r"` across internal/conformance and its readers (numbering, vacuous, topic, ratchet tests). Likely fix: a `.gitattributes` forcing `eol=lf` for `*.md`/`*.pl`/`*.tpm` testdata, or the corpus reader trimming `\r`.
2. Temp-path quoting when handing perl a script: `cvwalk_test.go` "probe must compile, stderr: Can't open perl script "'C:\Users\RUNNER~1\..."` (7) -- a path wrapped in single quotes, which Windows' shell does not strip.
3. `parenless_test.go`: "perl compiled = false, want true" (3) and "want a syntax error, got stderr:" (2) -- probably the same quoting.
4. `parseoracle`: a timeout fixture that does not time out (`scratch_test.go`), and `wider.pl bucketed exact` (`run_test.go`).
5. `internal/infer` failures not yet read.

TestCoreDeclarationsArePerls had the CRLF form on stdout and is fixed by 98c8a364 (`binmode STDOUT`).

## Steps

- [ ] Fetch the Windows job log; list every failing test by root cause (the groups above, plus infer)
- [ ] Fix CRLF at the source (.gitattributes for testdata), then re-run Windows CI
- [ ] Fix path quoting for perl invocations; re-run; fix what remains
- [ ] Windows CI green; commit

## Acceptance criteria

- [ ] the Windows CI job passes on pu (`sh -c 'gh run list --branch pu --workflow CI --limit 1 --json databaseId --jq ".[0].databaseId" | xargs -I{} gh run view {} --json jobs --jq ".jobs[] | select(.name==\"Test & Build (Windows)\") | .conclusion" | grep -qx success'`)
