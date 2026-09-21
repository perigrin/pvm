---
title: The corpus file format, decided rather than spiked
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocks:
- 01a0c35d-b59e-77f7-91a3-2a889ae3051a
- 01a0c35e-66f9-71b2-b5f3-3dd0ef1cba05
- 01a0c35f-a487-7aa8-934b-4e68dd0fbaa1
created: 2026-09-21T09:47:55.331597823Z
updated: 2026-09-21T10:00:06.974559359Z
---

The spike settled the file format by writing three files against it, but it
was written outside the chain and nothing in it was reviewed. This issue is
where the format stops being a proposal.

The format as spiked:

    #!perl
    # TIER 01_literals
    # INTRODUCES numeric literal
    # USES my, print
    # MEASURED perl 5.42.0
    # STATUS refuses
    --- source
    my $x = .5;
    print "$x\n";
    --- expect parses
    --- expect output
    0.5
    --- expect tokens
    one numeric literal whose text is ".5"
    no operator whose text is "."

`internal/conformance/file.go` parses this: sections in any order, a repeat
is an error, exactly one of `--- expect parses` / `--- expect parsent` is
required, and `--- expect output` strips exactly one trailing newline.

## What is actually open

Three things the spike decided by writing code rather than by deciding:

1. **`--- expect output` newline handling.** Stripping exactly one trailing
   newline makes `print "$x\n"` natural and makes a file testing trailing
   whitespace impossible to write. No corpus file needs the second yet.
2. **Whether `STATUS refuses` belongs in the header or is derived.** The
   runner already knows a file refuses -- it ran the parser. A header that
   can disagree with reality is a header that will.
3. **`USES` is unchecked.** `INTRODUCES` is what the tier lint reads; `USES`
   is documentation nothing verifies, so it can rot silently.

Decide all three in this issue, with the reason, rather than deferring them
to the next file that trips over them.

## Acceptance

- [ ] Every section absent, each section duplicated, both parses/parsent, neither, and an unknown section name each produce a distinct error (`go test ./internal/conformance/ -run '^TestParseFileSections$' -count=1 -v | grep -q '^--- PASS: TestParseFileSections'`)
- [ ] The three open format questions are answered in conformance/README.md, each with its reason (`grep -qi 'trailing newline' conformance/README.md`)
- [ ] The format is documented in conformance/README.md, which is where the corpus's public interface lives (`test -s conformance/README.md`)
