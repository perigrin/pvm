// ABOUTME: Aggregates a sweep's per-statement findings into a ranked table of where WRONG comes from.
// ABOUTME: A file count is undifferentiated work; a construct count says what to fix first.

package parseoracle

import (
	"fmt"
	"sort"
	"strings"
)

// WrongSites counts the sites the subject got WRONG, by marker and by file.
//
// Only BucketWrong findings count. A wider finding is a site the subject
// hedged -- it said "I cannot settle this", which is the behaviour the
// four-bucket design exists to reward -- so counting it as work to do would
// invert the whole metric.
//
// Excluded and errored files are skipped, matching Wrong(): a file the runner
// could not measure has no findings to attribute, and folding it in would
// make an environmental failure look like a parser defect.
func (r Report) WrongSites() (byMarker map[Marker]int, byFile map[string]int) {
	byMarker, byFile = map[Marker]int{}, map[string]int{}
	for _, f := range r.Files {
		if f.Excluded || f.Err != "" {
			continue
		}
		for _, finding := range f.Verdict.Findings {
			if finding.Bucket != BucketWrong {
				continue
			}
			byMarker[finding.Marker]++
			byFile[f.Path]++
		}
	}
	return byMarker, byFile
}

// WhereWrongComesFrom renders the ranked table.
//
// The point is the RANKING. A bucket count says 72 files are wrong, which is
// 72 units of undifferentiated work; the findings say those hold 249 sites
// across five markers, and that the ten worst files hold 145 of them -- 58%
// of the total in 1.6% of the corpus. That is not a corpus-wide problem, it
// is a handful of constructs, and this is the order they are worth fixing in.
//
// Deterministic: ties break on name, so the same corpus renders the same
// bytes and a report can be committed as a baseline.
func (r Report) WhereWrongComesFrom() string {
	byMarker, byFile := r.WrongSites()
	if len(byFile) == 0 {
		return "no WRONG sites\n"
	}

	var b strings.Builder
	total := 0
	for _, n := range byMarker {
		total += n
	}
	fmt.Fprintf(&b, "%d WRONG site(s) in %d file(s)\n", total, len(byFile))

	b.WriteString("by marker:\n")
	for _, m := range Markers { // declaration order, so the table is stable
		if n := byMarker[m]; n > 0 {
			fmt.Fprintf(&b, "  %-10s %5d\n", m, n)
		}
	}

	type row struct {
		path string
		n    int
	}
	rows := make([]row, 0, len(byFile))
	for path, n := range byFile {
		rows = append(rows, row{path, n})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].n != rows[j].n {
			return rows[i].n > rows[j].n
		}
		return rows[i].path < rows[j].path
	})

	b.WriteString("by file:\n")
	for i, r := range rows {
		if i >= 25 {
			fmt.Fprintf(&b, "  ... and %d more file(s)\n", len(rows)-25)
			break
		}
		fmt.Fprintf(&b, "  %5d  %s\n", r.n, r.path)
	}
	return b.String()
}
