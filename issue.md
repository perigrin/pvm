---
title: The corpus file format, decided rather than spiked
state: done
urgency: normal
milestone: m3-conformance-corpus
blocks:
- 01a0c35d-b59e-77f7-91a3-2a889ae3051a
- 01a0c35e-66f9-71b2-b5f3-3dd0ef1cba05
- 01a0c35f-a487-7aa8-934b-4e68dd0fbaa1
created: 2026-09-21T09:47:55.331597823Z
updated: 2026-09-21T22:35:24.308317593Z
sessions:
- start_sha: 7858a7677fde70f59c516c66454f5eac756aff99
  end_sha: 1e70de780155b64e97463108cd6157694d81d679
  commits: 1
  started_at: 2026-09-21T13:48:24.775365287Z
  ended_at: 2026-09-21T13:53:19.944003237Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T13:48:24.775365287Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T13:53:19.944003237Z
observed_paths:
- conformance/README.md
- internal/conformance/conformance_test.go
- internal/conformance/file.go
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

## Acceptance criteria

- [ ] The four things the format treats as errors -- a missing source, a repeated section, neither expectation, and an unknown section name -- each produce a distinct error (`go test ./internal/conformance/ -run '^TestParseFileErrors$' -count=1 -v | grep -q '^--- PASS: TestParseFileErrors'`)
- [ ] A file asserting BOTH parses and parsent is rejected rather than resolved by field order (`go test ./internal/conformance/ -run '^TestParseFileBothExpectations$' -count=1 -v | grep -q '^--- PASS: TestParseFileBothExpectations'`)
- [ ] The happy path still reads a real corpus file end to end (`go test ./internal/conformance/ -run '^TestParseFileSections$' -count=1 -v | grep -q '^--- PASS: TestParseFileSections'`)
- [ ] The three open format questions are answered in conformance/README.md, each with its reason (`grep -qi 'trailing newline' conformance/README.md`)
- [ ] The format is documented in conformance/README.md: every section name, the ordering rule, and the repeat-is-an-error rule (`go test ./internal/conformance/ -run '^TestReadmeDocumentsEverySection$' -count=1 -v | grep -q '^--- PASS: TestReadmeDocumentsEverySection'`)
### Acceptance correction, during execution

The first criterion claimed five error cases and cited `TestParseFileSections`,
which is the pre-existing happy-path test reading one real file. It passed
before any error case was written, so the command did not verify the claim --
the same defect class chain-review found across the chain.

Split into three criteria, each naming the test that actually asserts it.
`TestParseFileErrors` covers absent/duplicated/neither/unknown as subtests;
`TestParseFileBothExpectations` covers the contradiction the spike accepted;
`TestParseFileSections` keeps the happy path it always tested.

The criterion also over-claimed. "Every section absent" is not an error:
`--- expect output` and `--- expect tokens` are OPTIONAL, and a parsent file
normally has neither. Only a missing `--- source` is an error. Reworded to
name the four things the format actually rejects.

Note for later issues: `go test -run '<no match>'` prints `ok ... [no tests to
run]` and exits 0, so an acceptance command must grep a named PASS line rather
than the package result.
