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

// SiteBuckets counts the file's marker sites per marker and bucket, so the
// metric grades each op perl built rather than each file.
//
// A graded file's sites are exact unless a finding names them: each finding
// is one site that fell short, in its own bucket. A whole-file no-answer --
// a verdict that names no marker, reached before any marker was compared,
// whichever side declined -- carries no findings, and every one of its sites
// is no-answer; subtracting nothing from them would score them exact.
//
// A file that never reached a verdict, excluded or an oracle-side error,
// contributes nothing: its zero-value Verdict reads as exact.
//
// More findings than sites for a marker is an error rather than a negative
// exact count: the verdict and the facts no longer describe the same file.
func (r Result) SiteBuckets() (map[Marker]Totals, error) {
	counts := map[Marker]Totals{}
	if !r.measured() {
		return counts, nil
	}
	whole := BucketExact
	if r.Verdict.Bucket == BucketNoAnswer && r.Verdict.Marker == MarkerNone {
		whole = BucketNoAnswer
	}
	for m, lines := range r.Facts.Sites {
		if len(lines) > 0 {
			var t Totals
			t[whole] = len(lines)
			counts[m] = t
		}
	}
	if whole == BucketNoAnswer {
		return counts, nil
	}
	for _, f := range r.Verdict.Findings {
		t := counts[f.Marker]
		if t[BucketExact] == 0 {
			return nil, fmt.Errorf("%s has more %s findings than the %d site(s) perl reported",
				r.Path, f.Marker, len(r.Facts.Sites[f.Marker]))
		}
		t[BucketExact]--
		t[f.Bucket]++
		counts[f.Marker] = t
	}
	return counts, nil
}

// SiteBuckets sums Result.SiteBuckets over every file in the report.
func (r Report) SiteBuckets() (map[Marker]Totals, error) {
	sum := map[Marker]Totals{}
	for _, f := range r.Files {
		counts, err := f.SiteBuckets()
		if err != nil {
			return nil, err
		}
		for m, t := range counts {
			s := sum[m]
			for b, n := range t {
				s[b] += n
			}
			sum[m] = s
		}
	}
	return sum, nil
}
