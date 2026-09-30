---
title: A heredoc in a multi-line string's one-line block is queued past the string
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T18:50:09.142781156Z
updated: 2026-09-30T18:57:03.110125923Z
sessions:
- start_sha: 51b72146eadb4ff692d269aa362884d9b3c8c7c2
  end_sha: 8e6f0c76e7101ddf0d436fa5d512d7175fa93c7c
  commits: 1
  started_at: 2026-09-30T18:50:10.704200711Z
  ended_at: 2026-09-30T18:57:03.110125923Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T18:50:10.704200711Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-30T18:57:03.110125923Z
observed_paths:
- internal/lexer/interpheredoc_test.go
- internal/lexer/quote.go
---

51b72146 queued a heredoc opened in a string's `@{...}` block for the lines after the statement whenever the block itself had no newline. But perl's scan_heredoc looks for the newline in the enclosing construct (toke.c:11735), which is the whole string. In `"@{[ <<E1 ]}foo\nE1\n"` the body `E1\n` is inside the string: measured with perl 5.42, it prints "[foo\n]".

PerlOnJava unit/nested_heredoc.t: the file's single refusal moved from line 100 to line 10 and swallowed the rest of the file. The refusal count stayed at 1, so no ratchet noticed.

## Acceptance criteria

- [ ] A block's heredoc takes its body from the construct's next line (`go test ./internal/lexer/ -run '^TestInterpolationHeredoc$' -count=1 -v | grep -q '^--- PASS: TestInterpolationHeredoc'`)
