---
title: "`\"x @{[ <<EOT ]} x\"` lexes its heredoc body as code"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T18:39:35.353388254Z
updated: 2026-09-30T18:39:36.594742915Z
sessions:
- start_sha: fed0bab5ed153c134139df0f67090f36146d3ffc
  end_sha: ""
  commits: 0
  started_at: 2026-09-30T18:39:36.594742915Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T18:39:36.594742915Z
---

`my $text = "x @{[ <<'EOT' ]} x";` lexes its heredoc body (`HERE\nEOT\n`) as code. perl lexes a double-quoted string's `@{...}` and `${...}` blocks as Perl, and a heredoc opened there on a single line takes its body from the lines after the statement, like one in a pattern's `(?{...})`, which is already handled.

Measured with perl 5.42: the example prints "[x HERE\n x]", and `qq{a ${\ <<E2} b}` prints "[a body\n b]".

PerlOnJava unit/string_interpolation.t:350.

## Acceptance criteria

- [ ] A heredoc in a string's interpolation block takes the following lines (`go test ./internal/lexer/ -run '^TestInterpolationHeredoc$' -count=1 -v | grep -q '^--- PASS: TestInterpolationHeredoc'`)
