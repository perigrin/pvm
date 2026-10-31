// ABOUTME: Checks the committed perl labels: every corpus file labelled, the pinned -DDEBUGGING perl, a category per refusal.
// ABOUTME: Also shows the check failing on synthetic malformed labels; nothing here runs perl.

package chart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// inTreeFiles expands inTreeCorpus, failing on a glob that matches nothing.
func inTreeFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	for _, g := range inTreeCorpus {
		m, err := filepath.Glob(filepath.Join(repoRoot, g))
		if err != nil || len(m) == 0 {
			t.Fatalf("no corpus file matches %s (%v)", g, err)
		}
		for _, f := range m {
			rel, _ := filepath.Rel(repoRoot, f)
			files = append(files, filepath.ToSlash(rel))
		}
	}
	return files
}

// TestPerlLabels: the committed labels cover every in-tree corpus file, and
// every T1 file when T1_CORPUS is set, under the pinned perl.
func TestPerlLabels(t *testing.T) {
	check := func(name string, files []string) {
		lf, err := ReadLabels(filepath.Join(labelsDir, name))
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range CheckLabels(lf, files) {
			t.Errorf("%s: %s", name, p)
		}
	}
	check("in-tree.txt", inTreeFiles(t))

	root := os.Getenv("T1_CORPUS")
	if root == "" {
		t.Log("T1_CORPUS unset: T1's labels are not checked")
		return
	}
	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		t.Fatalf("T1_CORPUS=%s names no directory (%v)", root, err)
	}
	matches, err := filepath.Glob(filepath.Join(root, "*.t"))
	if err != nil || len(matches) == 0 {
		t.Fatalf("T1_CORPUS=%s holds no top-level *.t (%v)", root, err)
	}
	for i, m := range matches {
		matches[i] = filepath.Base(m)
	}
	check("t1.txt", matches)
}

// TestPerlLabelsRefusesMalformed: each defect in a synthetic label set is
// reported, and the well-formed set passes.
func TestPerlLabelsRefusesMalformed(t *testing.T) {
	const header = "# perl: /opt/perl/bin/perl\n" +
		"# version: 5.45.0\n" +
		"# git_commit_id: 94e5086608bcf1d01cd2d836f11ca8378a3320b6\n" +
		"# ccflags: -fwrapv -DDEBUGGING -pipe\n" +
		"# cwd: .\n# root: .\n# PERL5LIB: stubs\n# timeout: 30s\n"
	const body = "a.pl\taccept\n" +
		"b.pl\trefuse\tgrammar\tproposed\tsyntax error at b.pl line 2\n" +
		"c.pl\trefuse\tprecedence\treviewed\tsyntax error at c.pl line 1\treferee-noprec viable; referee-prec refuse@4\n"
	files := []string{"a.pl", "b.pl", "c.pl"}

	cases := []struct {
		name, text string
		files      []string
		want       string // a reported problem contains this; "" means none
	}{
		{"well formed", header + body, files, ""},
		{"corpus file missing", header + body, append(files, "d.pl"), "d.pl: not labelled"},
		{"not -DDEBUGGING", strings.Replace(header, " -DDEBUGGING", "", 1) + body, files, "-DDEBUGGING"},
		{"another commit", strings.Replace(header, "94e5086608bcf1d01cd2d836f11ca8378a3320b6", "0123456789abcdef", 1) + body, files, "94e50866"},
		{"category outside the five", header + strings.Replace(body, "\tgrammar\t", "\tlexical\t", 1), files, "lexical"},
		{"precedence without referee evidence", header + strings.Replace(body, "\treferee-noprec viable; referee-prec refuse@4", "", 1), files, "referee evidence"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "labels.txt")
			writeFile(t, path, c.text)
			lf, err := ReadLabels(path)
			if err != nil {
				t.Fatal(err)
			}
			got := CheckLabels(lf, c.files)
			if c.want == "" {
				if len(got) != 0 {
					t.Fatalf("CheckLabels = %q, want no problem", got)
				}
				return
			}
			for _, p := range got {
				if strings.Contains(p, c.want) {
					return
				}
			}
			t.Fatalf("CheckLabels = %q, want a problem containing %q", got, c.want)
		})
	}
}
