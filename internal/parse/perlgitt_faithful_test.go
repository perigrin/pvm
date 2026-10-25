// ABOUTME: Ratchets canonical re-emission (Faithful) per file over perl.git t/, failing in either direction.
// ABOUTME: Its "# N/M faithful" header is the count later re-emission work asserts its gains against.
package parse_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

const perlGitTFaithfulPath = "testdata/perlgittfaithful.ratchet"

// TestPerlGitTFaithfulRatchet: whether each perl.git t/ file's canon
// emission says what its source says, 1 = differs, as t1fidelity.ratchet
// holds it for T1. It walks the files as TestPerlGitTRatchet does, through
// the loader-aware entry point, and leaves perlRejects out.
func TestPerlGitTFaithfulRatchet(t *testing.T) {
	dir, files := perlGitTFiles(t)
	roots := perlGitTRoots(dir)

	s := parse.NewSession()
	now := make(map[string]int, len(files))
	for _, rel := range files {
		if _, rejected := perlRejects[rel]; rejected {
			continue
		}
		path := filepath.Join(dir, rel)
		src, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		n, err := s.ParseFileFrom(path, roots...)
		if err != nil {
			continue
		}
		now[rel] = faithfulCount(n, src)
	}

	if *updateRatchet {
		if err := os.WriteFile(perlGitTFaithfulPath, []byte(renderFaithfulRatchet(now)), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("rewrote %s (%d files); commit it with the change that moved it", perlGitTFaithfulPath, len(now))
		return
	}
	want, err := parseRatchetFile(perlGitTFaithfulPath)
	if err != nil {
		t.Fatalf("%v\nrun with -parse.update-ratchet to create it", err)
	}
	if err := faithfulDiff(want, now); err != nil {
		t.Error(err)
	}
}

// faithfulCount is a file's row: 1 when its emission differs from its
// source, 0 when it does not. A tree holding an Unknown is 1 before Faithful
// is asked, because canon writes an Unknown as its own source text, so
// Faithful would accept a refusal as a faithful re-emission.
func faithfulCount(n *parse.Node, src []byte) int {
	if countUnknown(n) > 0 {
		return 1
	}
	if ok, _ := parse.Faithful(n, src); !ok {
		return 1
	}
	return 0
}

// faithfulDiff compares a walk against the ratchet. An empty walk is an
// error of its own: it would otherwise agree with an empty ratchet.
func faithfulDiff(want, now map[string]int) error {
	if len(now) == 0 {
		return errors.New("the walk visited no files; a missing corpus is not a match")
	}
	return ratchetDiff(want, now)
}

// renderFaithfulRatchet is renderRatchet with the "# N/M faithful" line
// first, N the rows at 0 and M all the rows.
func renderFaithfulRatchet(now map[string]int) string {
	return faithfulHeader(now) + "\n" + renderRatchet(
		"perl.git t/ files whose canon emission does NOT say what the source says (1 = differs),\n"+
			"through parse.Session.ParseFileFrom; a tree holding an Unknown counts as 1.", now)
}

func faithfulHeader(counts map[string]int) string {
	faithful := 0
	for _, n := range counts {
		if n == 0 {
			faithful++
		}
	}
	return fmt.Sprintf("# %d/%d faithful", faithful, len(counts))
}

// faithfulHeaderAgrees checks a ratchet's "# N/M faithful" line, which must
// appear exactly once, against its rows, and its renderRatchet summary too.
func faithfulHeaderAgrees(path string) error {
	counts, err := parseRatchetFile(path)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var headers []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "# ") && strings.HasSuffix(line, " faithful") {
			headers = append(headers, line)
		}
	}
	if want := faithfulHeader(counts); len(headers) != 1 || headers[0] != want {
		return fmt.Errorf("%s: faithful header %q, its rows say %q", path, headers, want)
	}
	return ratchetHeaderAgrees(path)
}

func TestPerlGitTFaithfulRatchetHeaderMatchesRows(t *testing.T) {
	if err := faithfulHeaderAgrees(perlGitTFaithfulPath); err != nil {
		t.Error(err)
	}

	good := renderFaithfulRatchet(map[string]int{"a.t": 0, "b.t": 1, "c.t": 0})
	write := func(content string) string {
		path := filepath.Join(t.TempDir(), "fixture.ratchet")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	if err := faithfulHeaderAgrees(write(good)); err != nil {
		t.Fatalf("the header renderFaithfulRatchet wrote is rejected: %v", err)
	}
	for name, bad := range map[string]string{
		"count over different rows": strings.Replace(good, "\n0 c.t\n", "\n1 c.t\n", 1),
		"stated count":              strings.Replace(good, "# 2/3 faithful", "# 3/3 faithful", 1),
		"line missing":              strings.Replace(good, "# 2/3 faithful\n", "", 1),
	} {
		if err := faithfulHeaderAgrees(write(bad)); err == nil {
			t.Errorf("%s: a wrong faithful header passed:\n%s", name, bad)
		}
	}
}

func TestFaithfulRatchetDetectsSwap(t *testing.T) {
	want := map[string]int{"a.t": 1, "b.t": 0, "c.t": 0}
	now := map[string]int{"a.t": 0, "b.t": 1, "c.t": 0}
	if faithfulHeader(want) != faithfulHeader(now) {
		t.Fatal("the fixture must keep the faithful count")
	}
	err := faithfulDiff(want, now)
	if err == nil {
		t.Fatal("a swap that keeps the count passed")
	}
	for _, f := range []string{"a.t", "b.t"} {
		if !strings.Contains(err.Error(), f) {
			t.Errorf("the failure does not name %s:\n%v", f, err)
		}
	}
}

func TestFaithfulRatchetFailsOnUnvisitedFile(t *testing.T) {
	want := map[string]int{"a.t": 0, "b.t": 1}
	err := faithfulDiff(want, map[string]int{"a.t": 0})
	if err == nil || !strings.Contains(err.Error(), "b.t") {
		t.Errorf("a ratchet row the walk did not visit passed, or went unnamed: %v", err)
	}
}

func TestFaithfulWalkCountsUnknownTreeUnfaithful(t *testing.T) {
	src := []byte("print 1;")
	tree := &parse.Node{Kind: parse.SourceFile, Start: 0, End: len(src), Children: []*parse.Node{
		{Kind: parse.Unknown, Text: string(src), Start: 0, End: len(src)},
	}}
	// The guard matters only because Faithful cannot see the refusal.
	if ok, why := parse.Faithful(tree, src); !ok {
		t.Fatalf("Faithful now rejects a verbatim Unknown (%s); this fixture no longer proves the guard", why)
	}
	if got := faithfulCount(tree, src); got != 1 {
		t.Errorf("a tree holding an Unknown counted %d, want 1 (unfaithful)", got)
	}
}

func TestFaithfulRatchetRejectsEmptyWalk(t *testing.T) {
	if err := faithfulDiff(map[string]int{}, map[string]int{}); err == nil {
		t.Error("an empty walk matched an empty ratchet")
	}
}
