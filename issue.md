---
title: "A stray --- expect output inside a heredoc is read silently as a real section"
state: done
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T18:43:10.842334149Z
updated: 2026-09-21T22:36:20.011639467Z
sessions:
- start_sha: dc1bea2c2baed5af0cf71ddf16e379ee18855d30
  end_sha: b41801021927f68fcbf799c60d6b947393d983de
  commits: 1
  started_at: 2026-09-21T21:40:49.645212263Z
  ended_at: 2026-09-21T22:10:48.23407394Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T21:40:49.645212263Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T22:10:48.23407394Z
observed_paths:
- conformance/02_variables/README.md
- conformance/06_control/README.md
- conformance/07_subroutines/00_adjacency.t
- conformance/07_subroutines/08_parenless_extent.t
- conformance/07_subroutines/09_prototype_extent.t
- conformance/07_subroutines/11_code_ref_call.t
- conformance/07_subroutines/README.md
- conformance/08_references/README.md
- conformance/11_oo/README.md
- conformance/13_opaque/README.md
- conformance/README.md
- internal/conformance/claimonce_test.go
- internal/conformance/conformance_test.go
- internal/conformance/file.go
- internal/conformance/hardmarkers_test.go
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier07_callforms_test.go
- internal/conformance/tier07_test.go
- internal/parse/easy_test.go
- internal/parse/t2_test.go
---

A corpus file is split into sections by lines beginning `--- `. A heredoc
body, POD block or `__DATA__` section is opaque to that split, so a line
inside one that begins `--- ` is read as a section marker.

Measured against `ParseFile` while writing tier 13 at 0ce515cb, and the
two cases behave differently:

- A heredoc body containing `--- not a marker` is REJECTED, with
  `unknown section "not a marker"`. Loud, and therefore harmless.
- A heredoc body containing `--- expect output` is accepted SILENTLY. The
  source section truncates at that line, leaving an unterminated heredoc
  opener, and the remainder of the body becomes `ExpectOutput`. No error
  is reported.

The second case is the dangerous one. It mis-reads a valid Perl program as
a differently-shaped corpus file and nothing says so.

`conformance/13_opaque/README.md` documents the limitation and its files
are written clear of it, but documentation is not a check.

Two candidate fixes, neither obviously right:

1. Reject a `--- ` line whose section name is already `seen`, which is
   nearly the existing duplicate-section error -- it would catch
   `--- expect output` appearing twice but not a first occurrence inside
   a heredoc.
2. Require section markers to be the ONLY content on their line AND
   preceded by a blank line, which a heredoc body line would usually
   violate. This is a format change and would need every existing file
   re-checked.

Worth noting the format already has a related sharp edge that tier 13's
`08_data_section.t` pins rather than dodges: a section body includes the
blank separator line before the next marker, so inside a `__DATA__`
section that blank line is DATA. The file pins three lines of output for a
two-line data section and documents why.

## What the measurement decided

Option 2, and the re-check it demanded cost nothing. Measured at
`dc1bea2c` across all 134 corpus files: **472 markers, every one of them
already preceded by a blank line or by the comment block's last blank
line. Zero exceptions, zero markers with trailing content, zero markers
directly following another marker.** So the format change is a format
change on paper only -- no corpus file needed editing.

Option 1 was rejected because it does not close the hole it names: the
first `--- expect output` in a file is not `seen`, and the measured
failure is a FIRST occurrence inside a heredoc. It would catch a
duplicate, which the existing duplicate-section error already catches.

A third option -- lexing the source section to find its heredoc openers --
was rejected because the source section is Perl and the lexer that would
read it is the one this corpus exists to test. A corpus file must be
readable while that parser is broken, or a parser bug becomes a corpus
that cannot be read to diagnose it.

The rule is deliberately WEAK: a heredoc body may contain a blank line and
then `--- expect output` and still slip through. It turns the reachable
accident loud, which is the part that was missing, and leaves the rest
where `13_opaque/README.md` already records it.

A marker directly after another marker stays legal, because a section with
an empty body is ordinary -- `--- expect parses` carries nothing at all.

## Acceptance criteria

- [ ] A `--- expect output` line inside a heredoc body is REJECTED rather than silently truncating the source and pinning the rest as output, and the legitimate marker positions -- start of input, after a comment block, directly after a bodiless marker -- still parse (`go test ./internal/conformance/ -run '^TestParseFileMarkerNeedsBlankLineBefore$' -count=1 -v | grep -q '^--- PASS: TestParseFileMarkerNeedsBlankLineBefore'`)
- [ ] `13_opaque/08_data_section.t` is unbroken: a section body still runs to the next marker INCLUDING the blank separator line, so that file's data section still ends `two\n\n` and still pins three lines of output for two lines of data (`go test ./internal/conformance/ -run '^TestParseFileBodyRunsToTheNextMarker$' -count=1 -v | grep -q '^--- PASS: TestParseFileBodyRunsToTheNextMarker'`)
- [ ] The rules the blank-line change did not touch still hold -- the one-newline strip, CRLF normalisation, the nameless marker, the both-expectations contradiction, the empty-versus-absent output pin (`go test ./internal/conformance/ -run '^TestParseFile' -count=1 | grep -q '^ok'`)
- [ ] Every existing corpus file still parses and lints under the new rule, which is what makes the format change free (`go test ./internal/conformance/ -run '^TestCorpusLints$' -count=1 -v | grep -q '^--- PASS: TestCorpusLints'`)
- [ ] `conformance/README.md` states the rule, since every well-formed example already satisfies it and an author cannot infer it from one (`go test ./internal/conformance/ -run '^TestReadmeDocumentsEverySection$' -count=1 -v | grep -q '^--- PASS: TestReadmeDocumentsEverySection'`)
