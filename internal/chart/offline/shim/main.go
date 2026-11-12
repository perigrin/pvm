// ABOUTME: Builds the shim perl.git t/ compiles in, with parseoracle.BuildShim, for label.sh and tdt-align.sh.
// ABOUTME: Offline only: usage: shim <perl> <perl5-tree> <dest>. Refuses a perl5 tree not at 94e50866 or not clean.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"tamarou.com/pvm/internal/parseoracle"
)

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "shim: "+format+"\n", a...)
	os.Exit(1)
}

func git(tree string, args ...string) string {
	b, err := exec.Command("git", append([]string{"-C", tree}, args...)...).Output()
	if err != nil {
		die("%s is not a git tree: %v", tree, err)
	}
	return strings.TrimSpace(string(b))
}

func main() {
	if len(os.Args) != 4 {
		die("usage: shim <perl> <perl5-tree> <dest>")
	}
	perl, tree, dest := os.Args[1], os.Args[2], os.Args[3]
	head := git(tree, "rev-parse", "HEAD")
	if !strings.HasPrefix(head, "94e50866") {
		die("%s is at %s, not 94e50866", tree, head)
	}
	if git(tree, "status", "--porcelain") != "" {
		die("%s has uncommitted changes; the pinned tree must be clean", tree)
	}
	if err := parseoracle.BuildShim(perl, tree, dest, head); err != nil {
		die("building the shim: %v", err)
	}
}
