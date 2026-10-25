// ABOUTME: The per-site ratchet: sites.txt rows per file, marker and bucket, plus per-scope totals, compared both ways.
// ABOUTME: SiteFloor is M3's oracle condition: the M3 dirs' sites at least 80% exact, and no WRONG site anywhere.

package parseoracle

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// M3Dirs are the corpus directories, under t/, that M3's floor measures.
var M3Dirs = []string{"op", "mro", "uni", "lib", "test_pl"}

// InM3Dir reports whether a corpus path's first segment is one of M3Dirs.
// A whole segment, so opbasic/ is not op/.
func InM3Dir(path string) bool {
	dir, _, ok := strings.Cut(filepath.ToSlash(path), "/")
	return ok && slices.Contains(M3Dirs, dir)
}

// SiteRows renders the report's per-site counts as sites.txt rows:
//
//	site <path> <marker> <bucket> <count>
//	total <scope> <marker> <bucket> <count>
//
// with scope `all` or `m3`. Only nonzero rows are written, sorted by kind,
// path or scope, marker and bucket.
func SiteRows(report Report) ([]string, error) {
	counts := map[string]int{}
	for _, f := range report.Files {
		buckets, err := f.SiteBuckets()
		if err != nil {
			return nil, err
		}
		for m, totals := range buckets {
			for b, n := range totals {
				if n == 0 {
					continue
				}
				tail := fmt.Sprintf("%s %s", m, Bucket(b))
				counts["site "+f.Path+" "+tail] += n
				counts["total all "+tail] += n
				if InM3Dir(f.Path) {
					counts["total m3 "+tail] += n
				}
			}
		}
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rows := make([]string, len(keys))
	for i, k := range keys {
		rows[i] = k + " " + strconv.Itoa(counts[k])
	}
	return rows, nil
}

// SiteFloor checks M3's oracle condition against the total rows: the M3
// dirs' exact sites are at least 80% of all their sites, wider and
// no-answer included, and no site in the whole corpus is WRONG.
func SiteFloor(rows []string) error {
	var m3Exact, m3Sites, wrong int
	for _, row := range rows {
		f := strings.Fields(row)
		if len(f) != 5 || f[0] != "total" {
			continue
		}
		n, err := strconv.Atoi(f[4])
		if err != nil {
			return fmt.Errorf("sites row %q: %w", row, err)
		}
		switch f[1] {
		case "all":
			if f[3] == BucketWrong.String() {
				wrong += n
			}
		case "m3":
			m3Sites += n
			if f[3] == BucketExact.String() {
				m3Exact += n
			}
		}
	}

	var problems []string
	if wrong != 0 {
		problems = append(problems, fmt.Sprintf("%d WRONG site(s): WRONG is a gate, and it is 0 everywhere", wrong))
	}
	switch {
	case m3Sites == 0:
		problems = append(problems, "the M3 dirs have 0 sites: 0 of 0 is not 80% exact")
	case m3Exact*5 < m3Sites*4:
		problems = append(problems, fmt.Sprintf("the M3 dirs have %d of %d sites exact (%.1f%%), below the 80%% floor",
			m3Exact, m3Sites, 100*float64(m3Exact)/float64(m3Sites)))
	}
	if len(problems) != 0 {
		return fmt.Errorf("the M3 site floor fails: %s", strings.Join(problems, "; "))
	}
	return nil
}

// CheckSites compares the report's site rows to want, failing on any
// difference in either direction. It also fails when the report is empty,
// has no sites, or lacks a file the per-file baseline files lists: a sweep
// that lost files has shrunk the denominator, not improved.
func CheckSites(want []string, report Report, files Baseline) error {
	if len(report.Files) == 0 {
		return fmt.Errorf("the report has 0 files: a corpus that was never walked matches nothing")
	}
	got, err := SiteRows(report)
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(got, func(r string) bool { return strings.HasPrefix(r, "site ") }) {
		return fmt.Errorf("the report's %d files have 0 sites: there is nothing to compare", len(report.Files))
	}

	var problems []string
	swept := make(map[string]bool, len(report.Files))
	for _, f := range report.Files {
		swept[f.Path] = true
	}
	for _, row := range files.Rows {
		if !swept[row.Path] {
			problems = append(problems, "dropped: "+row.Path+" is in the per-file baseline but not in this sweep")
		}
	}

	in := func(rows []string) map[string]bool {
		set := make(map[string]bool, len(rows))
		for _, r := range rows {
			set[r] = true
		}
		return set
	}
	gotSet, wantSet := in(got), in(want)
	for _, r := range want {
		if !gotSet[r] {
			problems = append(problems, "- "+r)
		}
	}
	for _, r := range got {
		if !wantSet[r] {
			problems = append(problems, "+ "+r)
		}
	}
	if len(problems) != 0 {
		return fmt.Errorf("the per-site counts no longer match the committed rows "+
			"(- committed, + measured); a deliberate change re-runs with -parseoracle.update "+
			"and commits sites.txt in the same commit:\n  %s", strings.Join(problems, "\n  "))
	}
	return nil
}

// sitesHeader opens a written sites.txt.
const sitesHeader = `# ABOUTME: Frozen per-site oracle verdicts: per file, marker and bucket, and per-scope totals (all, m3).
# ABOUTME: Regenerated under -parseoracle.update: sites.txt by TestSiteRatchetCorpus, fixture-sites.txt by TestRatchet.
#
# DO NOT EDIT BY HAND. A move in either direction fails the site ratchet.
#
# site <path> <marker> <bucket> <count>
# total <scope> <marker> <bucket> <count>
`

// WriteSiteRows writes rows to path under the sites.txt header.
func WriteSiteRows(path string, rows []string) error {
	return os.WriteFile(path, []byte(sitesHeader+strings.Join(rows, "\n")+"\n"), 0o644)
}

// LoadSiteRows reads a sites.txt's rows. A missing file, or one with no
// rows, is an error: a ratchet with nothing to compare cannot fail.
func LoadSiteRows(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rows []string
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			rows = append(rows, line)
		}
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("%s has no rows: a site ratchet over nothing agrees with anything", path)
	}
	return rows, nil
}
