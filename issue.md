---
title: A format argument block holding a nested format ends the outer body early
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T20:55:26.202515665Z
updated: 2026-09-29T23:28:51.166983486Z
sessions:
- start_sha: 85a1b24708885d3b391109f081d755998f949c7c
  end_sha: 3b2e9d4a691b7ab3d1a26ce791dbc5e36b85748d
  commits: 1
  started_at: 2026-09-29T23:19:24.790034874Z
  ended_at: 2026-09-29T23:28:51.166983486Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T23:19:24.790034874Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T23:28:51.166983486Z
observed_paths:
- internal/lexer/formatargblock_test.go
- internal/lexer/pod.go
- internal/parse/testdata/perlgitt.ratchet
---

A format argument line may open a `{` block that spans lines, and that block may itself declare a format. perl.git t/op/write.t:1938-1950:

    format NEST =
    @<<<
    {
        my $birds = "birds";
        local *NEST = *BIRDS{FORMAT};
        write NEST;
        format BIRDS =
    @<<<<<
    $birds;
    .
        "nest"
    }
    .

scanFormatBody ends the outer body at the first lone `.`, which belongs to the nested BIRDS format, so `"nest" } .` is left as code and the `}` refuses. In LEX_FORMLINE perl lexes argument lines as Perl, so a brace there opens a real block, and the body ends only at a `.` outside it.

perlform settles the representation: "the expressions may be spread out to more than one line if enclosed in braces. If so, the opening brace must be the first token on the first line." So the body stays one opaque FormatBody token, and only its end moves. A braced argument line is lexed by a sub-lexer up to its matching `}`, which handles the nested format as ordinary Perl, and the body resumes after it. Measured with perl 5.42: the NEST/BIRDS shape writes "birds" then "nest" and runs the statement after the outer `.`.

## Acceptance criteria

- [ ] A braced argument block holding a nested format stays in the body (`go test ./internal/lexer/ -run '^TestFormatArgumentBlock$' -count=1 -v | grep -q '^--- PASS: TestFormatArgumentBlock'`)
