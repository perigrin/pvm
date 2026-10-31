// ABOUTME: Reads the committed perl labels (testdata/labels/) and checks them against a corpus file list.
// ABOUTME: Labels come from offline/label.sh under the pinned -DDEBUGGING perl; nothing here runs perl.

// Package chart is the Earley chart-parser spike (docs/plans/2026-10-30-chart-parser-v2-design.md).
package chart

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// labelsDir holds the committed label files, one per corpus root.
var labelsDir = filepath.Join("testdata", "labels")

// pinnedCommit is the perl blead commit every label must come from.
const pinnedCommit = "94e50866"

// categories are the five refusal categories of spec §9.1.
var categories = map[string]bool{"grammar": true, "precedence": true, "scope": true, "semantic": true, "environment": true}

// refereeEvidence is the shape a precedence label's evidence takes: the
// file's -DT prefix is viable without precedence and refused with it.
var refereeEvidence = regexp.MustCompile(`^referee-noprec viable; referee-prec refuse@[0-9]+$`)

// Label is one corpus file's perl verdict. A refusal carries a category,
// its status (proposed by label.sh, or reviewed), perl's first error and,
// for precedence, the referee evidence.
type Label struct {
	Path, Verdict, Category, Status, Evidence, Referee string
}

// LabelFile is a label file: its provenance header and its labels.
type LabelFile struct {
	Provenance map[string]string
	Labels     []Label
}

// Refused reports whether perl refuses the file.
func (l Label) Refused() bool { return l.Verdict == "refuse" }

// ReadLabels reads a label file: "# key: value" provenance lines, then one
// tab-separated line per file, "path accept" or
// "path refuse category status evidence [referee]".
func ReadLabels(path string) (*LabelFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	lf := &LabelFile{Provenance: map[string]string{}}
	sc := bufio.NewScanner(f)
	sc.Buffer(nil, 1<<20)
	for n := 1; sc.Scan(); n++ {
		line := sc.Text()
		if rest, ok := strings.CutPrefix(line, "# "); ok {
			k, v, _ := strings.Cut(rest, ": ")
			lf.Provenance[k] = v
			continue
		}
		c := strings.Split(line, "\t")
		switch {
		case len(c) == 2 && c[1] == "accept":
			lf.Labels = append(lf.Labels, Label{Path: c[0], Verdict: "accept"})
		case (len(c) == 5 || len(c) == 6) && c[1] == "refuse":
			l := Label{Path: c[0], Verdict: "refuse", Category: c[2], Status: c[3], Evidence: c[4]}
			if len(c) == 6 {
				l.Referee = c[5]
			}
			lf.Labels = append(lf.Labels, l)
		default:
			return nil, fmt.Errorf("%s:%d: malformed label %q", path, n, line)
		}
	}
	return lf, sc.Err()
}

// CheckLabels returns one problem per defect: a provenance not naming a
// -DDEBUGGING perl at the pinned commit, a corpus file unlabelled or labelled
// twice, a label naming no corpus file, a refusal whose category, status or
// evidence is wrong, or a precedence refusal without referee evidence.
func CheckLabels(lf *LabelFile, files []string) []string {
	var out []string
	if !strings.Contains(" "+lf.Provenance["ccflags"]+" ", " -DDEBUGGING ") {
		out = append(out, fmt.Sprintf("provenance ccflags %q lack -DDEBUGGING", lf.Provenance["ccflags"]))
	}
	if !strings.HasPrefix(lf.Provenance["git_commit_id"], pinnedCommit) {
		out = append(out, fmt.Sprintf("provenance git_commit_id %q is not %s", lf.Provenance["git_commit_id"], pinnedCommit))
	}
	for _, k := range []string{"perl", "version", "cwd", "root", "PERL5LIB", "timeout"} {
		if _, ok := lf.Provenance[k]; !ok {
			out = append(out, "provenance has no "+k)
		}
	}
	want := map[string]bool{}
	for _, f := range files {
		want[f] = true
	}
	seen := map[string]bool{}
	for _, l := range lf.Labels {
		switch {
		case seen[l.Path]:
			out = append(out, l.Path+": labelled twice")
		case !want[l.Path]:
			out = append(out, l.Path+": labelled but not a corpus file")
		}
		seen[l.Path] = true
		if !l.Refused() {
			continue
		}
		if !categories[l.Category] {
			out = append(out, fmt.Sprintf("%s: category %q is not one of the five", l.Path, l.Category))
		}
		// shortcut: accepts an unreviewed (proposed) category, until LABELS-review makes reviewed the only status.
		if l.Status != "proposed" && l.Status != "reviewed" {
			out = append(out, fmt.Sprintf("%s: status %q is neither proposed nor reviewed", l.Path, l.Status))
		}
		if l.Evidence == "" {
			out = append(out, l.Path+": refusal without perl's first error")
		}
		if l.Category == "precedence" && !refereeEvidence.MatchString(l.Referee) {
			out = append(out, fmt.Sprintf("%s: precedence without referee evidence (%q)", l.Path, l.Referee))
		}
	}
	for _, f := range files {
		if !seen[f] {
			out = append(out, f+": not labelled")
		}
	}
	return out
}
