---
title: A format argument block holding a nested format ends the outer body early
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T20:55:26.202515665Z
updated: 2026-09-29T20:55:26.202515665Z
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

The body is one opaque FormatBody token today. Fixing this means either lexing argument lines as Perl or tracking their brackets. Both are design choices about how format bodies are represented, so this needs a decision before code.

## Acceptance criteria

- [ ] perl.git t/op/write.t parses clean (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1` with `0 op/write.t` in internal/parse/testdata/perlgitt.ratchet)
