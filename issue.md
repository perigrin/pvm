---
title: A heredoc body in the middle of a running string is read as string text
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T19:04:21.072317235Z
updated: 2026-09-30T19:04:22.473357643Z
sessions:
- start_sha: 8e6f0c76e7101ddf0d436fa5d512d7175fa93c7c
  end_sha: ""
  commits: 0
  started_at: 2026-09-30T19:04:22.473357643Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T19:04:22.473357643Z
---

A heredoc queued on a line whose string runs on to later lines must take its body from the next line, out of the middle of that string. perl's scan_heredoc removes the body from the input as soon as it lexes the opener, and the string resumes after it. The lexer scanned the string straight through, so the body became string text and the heredoc never ended.

Measured with perl 5.42: `<<E21 . 'single\nE21 content\nE21\nquoted'` is "E21 content\n" . "single\nquoted", and `<<E13 . "@{[ <<E14 ]}text\nouter 1\nE13\ninner\nE14\nafter"` prints "[outer 1\ninner\ntext\nafter]".

Canon has the matching bug. It writes bodies after the statement's terminator, and perl misreads that when a quote in the statement runs across a newline after the opener. The bodies have to go at that newline. This showed up in PerlOnJava unit/heredoc_infix_continuation.t and perl.git t/run/dtrace.t.

PerlOnJava unit/nested_heredoc.t:94-100 and :148.

## Acceptance criteria

- [ ] A queued heredoc's body is taken out of a string that runs on (`go test ./internal/lexer/ -run '^TestHeredocInsideRunningString$' -count=1 -v | grep -q '^--- PASS: TestHeredocInsideRunningString'`)
