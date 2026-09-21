---
title: "A stray --- expect output inside a heredoc is read silently as a real section"
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T18:43:10.842334149Z
updated: 2026-09-21T21:40:49.645212263Z
sessions:
- start_sha: dc1bea2c2baed5af0cf71ddf16e379ee18855d30
  end_sha: ""
  commits: 0
  started_at: 2026-09-21T21:40:49.645212263Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T21:40:49.645212263Z
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
