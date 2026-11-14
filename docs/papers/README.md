<!-- ABOUTME: Papers the chart parser's design and spikes read, kept so a reader can check a citation against the text. -->
<!-- ABOUTME: Each entry gives the citation, the licence it is redistributed under, and where pvm's documents use it. -->

# Papers

## Scott (2008), SPPF-style parsing from Earley recognisers

`scott-2008-sppf-style-parsing-from-earley-recognisers.pdf`

Elizabeth Scott, "SPPF-Style Parsing From Earley Recognisers", *Electronic
Notes in Theoretical Computer Science* 203(2) (2008), 53–67,
doi:10.1016/j.entcs.2008.03.044. © 2008 Elsevier B.V., open access under
CC BY-NC-ND. The file is the publisher's PDF with its streams recompressed
losslessly (qpdf) to fit the repository's file-size hook; the content is
unchanged.

- §3 shows Earley's own pointers give spurious derivations.
- §4 decorates items with labelled predecessor and reduction pointers and
  walks the sets into a binarised SPPF.
- §5 builds the SPPF during recognition with a node in each item.

Cited by the chart-parser design (`docs/plans/2026-10-29-chart-parser-design.md`)
and its assessment's finding B7 (`docs/assessments/chart-waypoint.md`), which
rejected §5's form because a node in an item defeats set interning.
