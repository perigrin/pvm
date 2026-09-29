---
title: "`sub 'Hello'_he_said` opens a string: a declared name's leading apostrophe"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T21:09:54.112615857Z
updated: 2026-09-29T21:09:56.186907831Z
sessions:
- start_sha: e9260f1981687ce4acae6a901b5da5d8b651d8be
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T21:09:56.186907831Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T21:09:56.186907831Z
---

`sub 'Hello'_he_said (_);` and `format 'one =` read the leading apostrophe as the start of a string. toke.c scans the name after `sub` and `format` with scan_word, and its parse_ident reads `'` before an identifier start as `::` in the first position as well as later ones.

Measured with perl 5.42:

    $ perl -MO=Deparse -e "no warnings; sub 'Hello'_he_said (_);"
    sub Hello::_he_said (_);

A file with `format 'one =`, a picture line and `.` deparses as `format one =`.

perl.git t/comp/parser.t:480 and :506.

## Acceptance criteria

- [ ] A declared name's leading apostrophe is a separator (`go test ./internal/lexer/ -run '^TestDeclNameLeadingApostrophe$' -count=1 -v | grep -q '^--- PASS: TestDeclNameLeadingApostrophe'`)
