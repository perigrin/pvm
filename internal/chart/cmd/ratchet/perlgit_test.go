// ABOUTME: Tests the perl.git t/ corpus input: read from PERL5_CORPUS, refused when missing, not clean or not pinned.
// ABOUTME: Also that its superset population is exactly the files perlgit.txt and recursive descent accept, less the lexer lists.

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart"
	"tamarou.com/pvm/internal/chart/grammar"
)

// gitTree is a git tree at dir holding t/a.t, committed with plumbing so no
// hook or user config is involved; its HEAD is not perlGitPin.
func gitTree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "t"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "t", "a.t"), []byte("print 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) string {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q")
	git("add", "t/a.t")
	git("update-ref", "HEAD", git("commit-tree", git("write-tree"), "-m", "a"))
	return dir
}

// TestPerlGitCorpus: the pinned tree reads as perlgit/<path under t/>, every
// .t file once; a missing, unclean or unpinned tree is refused naming the
// tree and the reason on one line.
func TestPerlGitCorpus(t *testing.T) {
	refused := func(t *testing.T, root, why string) {
		t.Helper()
		srcs := map[string][]byte{}
		files, err := perlGit(root, srcs)
		want := "perl.git corpus " + root + ": " + why
		if err == nil || err.Error() != want {
			t.Errorf("perlGit(%s) = %d files, %v; want the error %q", root, len(files), err, want)
		}
		if len(files) > 0 || len(srcs) > 0 {
			t.Errorf("a refused tree still read %d files", len(srcs))
		}
	}
	t.Run("absent", func(t *testing.T) { refused(t, filepath.Join(t.TempDir(), "perl5"), "no such directory") })
	t.Run("wrong commit", func(t *testing.T) { refused(t, gitTree(t), "not at 94e50866") })
	t.Run("dirty", func(t *testing.T) {
		dir := gitTree(t)
		if err := os.WriteFile(filepath.Join(dir, "t", "a.t"), []byte("print 2;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		refused(t, dir, "not clean")
	})
	t.Run("untracked", func(t *testing.T) {
		dir := gitTree(t)
		if err := os.WriteFile(filepath.Join(dir, "t", "b.t"), nil, 0o644); err != nil {
			t.Fatal(err)
		}
		refused(t, dir, "not clean")
	})
	t.Run("present", func(t *testing.T) {
		root := perl5Root()
		if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
			t.Skipf("no perl5 tree at %s; set PERL5_CORPUS to the perl5 tree at 94e50866 (%v)", root, err)
		}
		srcs := map[string][]byte{}
		files, err := perlGit(root, srcs)
		if err != nil {
			t.Fatal(err)
		}
		var want []string
		err = filepath.WalkDir(filepath.Join(root, "t"), func(p string, d os.DirEntry, err error) error {
			if err == nil && strings.HasSuffix(p, ".t") && !d.IsDir() {
				rel, _ := filepath.Rel(filepath.Join(root, "t"), p)
				want = append(want, "perlgit/"+filepath.ToSlash(rel))
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		if got := paths(files); !slices.Equal(got, want) || len(srcs) != len(want) {
			t.Errorf("read %d files (%d sources), want the %d .t files under %s/t", len(got), len(srcs), len(want), root)
		}
		for _, f := range files {
			if f.Corpus != "perlgit" {
				t.Fatalf("%s is in corpus %q, want perlgit", f.Path, f.Corpus)
			}
		}
	})
}

// TestPerlGitPopulationFromLabels: the perl.git files in the superset
// population are exactly those perlgit.txt labels accept and recursive descent
// accepts, less the lexer lists, so a reader that drops files cannot pass.
func TestPerlGitPopulationFromLabels(t *testing.T) {
	t.Chdir("../../../..")
	root := perl5Root()
	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		t.Skipf("no perl5 tree at %s; set PERL5_CORPUS to the perl5 tree at 94e50866 (%v)", root, err)
	}
	srcs := map[string][]byte{}
	files, err := perlGit(root, srcs)
	if err != nil {
		t.Fatal(err)
	}
	in, err := realInputs(nil, srcs)
	if err != nil {
		t.Fatal(err)
	}
	rdOf := map[string]bool{}
	rd := in.RD
	in.RD = func(p string) (bool, error) { // each file parsed once, for both sides
		if a, ok := rdOf[p]; ok {
			return a, nil
		}
		a, err := rd(p)
		rdOf[p] = a
		return a, err
	}
	pop, err := supersetPopulation(files, in)
	if err != nil {
		t.Fatal(err)
	}
	lf, err := chart.ReadLabels(filepath.Join(testdata, "labels", "perlgit.txt"))
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, l := range lf.Labels {
		p := "perlgit/" + l.Path
		if _, ok := srcs[p]; !ok {
			t.Errorf("%s is labelled but not read", p)
			continue
		}
		if a, err := in.RD(p); err != nil {
			t.Fatal(err)
		} else if !l.Refused() && a && !in.Excluded[p] {
			want = append(want, p)
		}
	}
	got := paths(pop)
	slices.Sort(got)
	slices.Sort(want)
	if len(want) == 0 || !slices.Equal(got, want) {
		t.Errorf("perl.git population has %d files, want the labels' %d", len(got), len(want))
	}
}

// TestTDTLeavesOutUncaptured: a perl.git file perlgit-uncaptured.txt lists
// needs no fixture; any other perl.git file still does.
func TestTDTLeavesOutUncaptured(t *testing.T) {
	t.Chdir("../../../..")
	g := &grammar.Grammar{Classes: toyDecls}
	tab := &grammar.Tables{}
	for _, n := range toyNames {
		tab.Symbols = append(tab.Symbols, grammar.TableSymbol{Name: n})
	}
	old := reviewedFile // the real list names the real table's keys
	reviewedFile = filepath.Join(t.TempDir(), "reviewed")
	defer func() { reviewedFile = old }()
	if err := os.WriteFile(reviewedFile, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "perlgit"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "perlgit", "a.t.tdt"), []byte("# capture: whole\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		path string
		ok   bool
	}{{"perlgit/re/pat_psycho.t", true}, {"perlgit/re/pat.t", false}} {
		files := []File{{"perlgit/a.t", "perlgit"}, {c.path, "perlgit"}}
		srcs := map[string][]byte{files[0].Path: []byte("1;"), c.path: []byte("1;")}
		err := measureTDT(dir, g, tab, files, srcs, map[string]string{})
		if c.ok && err != nil || !c.ok && (err == nil || !strings.Contains(err.Error(), c.path)) {
			t.Errorf("%s without a fixture: %v", c.path, err)
		}
	}
}
