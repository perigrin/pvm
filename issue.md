---
title: A heredoc in a multi-line string's one-line block is queued past the string
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T18:50:09.142781156Z
updated: 2026-09-30T18:50:09.142781156Z
---

51b72146 queued a heredoc opened in a string's `@{...}` block for the lines after the statement whenever the block itself had no newline. But perl's scan_heredoc looks for the newline in the enclosing construct (toke.c:11735), which is the whole string. In `"@{[ <<E1 ]}foo\nE1\n"` the body `E1\n` is inside the string: measured with perl 5.42, it prints "[foo\n]".

PerlOnJava unit/nested_heredoc.t: the file's single refusal moved from line 100 to line 10 and swallowed the rest of the file. The refusal count stayed at 1, so no ratchet noticed.

## Acceptance criteria

- [ ] A block's heredoc takes its body from the construct's next line (`go test ./internal/lexer/ -run '^TestInterpolationHeredoc$' -count=1 -v | grep -q '^--- PASS: TestInterpolationHeredoc'`)
