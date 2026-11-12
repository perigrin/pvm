// ABOUTME: The perl.git t/ corpus (spec §9.1): every .t file under $PERL5_CORPUS/t, named perlgit/<path under t/>.
// ABOUTME: Refuses a tree that is missing, not clean, or not at 94e50866 (perigrin, 2026-11-10: "Accept it, pinned").

package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// perlGitPin is the perl5 commit the labels and the -DT captures were taken at.
const perlGitPin = "94e50866"

// perl5Root is PERL5_CORPUS, the variable parseoracle reads, else the
// development machine's perl5 tree.
func perl5Root() string {
	if r := os.Getenv("PERL5_CORPUS"); r != "" {
		return r
	}
	return "/opt/corpora/perl5"
}

// perlGit adds every .t file under root's t/ to srcs as perlgit/<path under
// t/> and returns them. It refuses, naming root, a tree that is no directory,
// has uncommitted or untracked changes, or is at another commit.
func perlGit(root string, srcs map[string][]byte) ([]File, error) {
	refuse := func(why string) error { return fmt.Errorf("perl.git corpus %s: %s", root, why) }
	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		return nil, refuse("no such directory")
	}
	git := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", root}, args...)...).Output()
		return strings.TrimSpace(string(out)), err
	}
	// Clean before pinned, so an edited tree reads as not clean whatever its commit.
	if st, err := git("status", "--porcelain"); err != nil || st != "" {
		return nil, refuse("not clean")
	}
	if head, err := git("rev-parse", "HEAD"); err != nil || !strings.HasPrefix(head, perlGitPin) {
		return nil, refuse("not at " + perlGitPin)
	}
	t := filepath.Join(root, "t")
	var files []File
	err := filepath.WalkDir(t, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".t") {
			return err
		}
		src, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(t, p)
		name := "perlgit/" + filepath.ToSlash(rel)
		files, srcs[name] = append(files, File{name, "perlgit"}), src
		return nil
	})
	if err == nil && len(files) == 0 {
		return nil, refuse("no .t file under t/")
	}
	return files, err
}
