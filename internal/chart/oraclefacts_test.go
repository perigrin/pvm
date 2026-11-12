// ABOUTME: Checks the committed oracle facts: header, one record per oracle-graded file, every record readable, none stray.
// ABOUTME: Also shows the reader failing on malformed records; nothing here runs perl.

package chart

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var oracleDir = filepath.Join("testdata", "oracle")

// oracleKeys is the key set each facts file must hold exactly: the files
// perl accepts by the committed labels, and for perl.git t/ every .t file,
// whose own ok says whether perl accepts it. The perl.git set is nil when the
// corpus is absent.
func oracleKeys(t *testing.T) map[string]map[string]bool {
	t.Helper()
	accepted := func(name, prefix string) map[string]bool {
		lf, err := ReadLabels(filepath.Join(labelsDir, name))
		if err != nil {
			t.Fatal(err)
		}
		keys := map[string]bool{}
		for _, l := range lf.Labels {
			if !l.Refused() {
				keys[prefix+l.Path] = true
			}
		}
		return keys
	}
	keys := map[string]map[string]bool{"in-tree.txt": accepted("in-tree.txt", ""), "t1.txt": accepted("t1.txt", "T1/")}
	root := os.Getenv("PERL5_CORPUS")
	if root == "" {
		root = "/opt/corpora/perl5"
	}
	if _, err := os.Stat(filepath.Join(root, "t", "test.pl")); err != nil {
		return keys
	}
	pg := map[string]bool{}
	filepath.WalkDir(filepath.Join(root, "t"), func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".t") {
			rel, _ := filepath.Rel(filepath.Join(root, "t"), p)
			pg["perlgit/"+filepath.ToSlash(rel)] = true
		}
		return nil
	})
	keys["perlgit.txt"] = pg
	return keys
}

func readOracle(t *testing.T, name string) *OracleFactsFile {
	t.Helper()
	of, err := ReadOracleFacts(filepath.Join(oracleDir, name))
	if err != nil {
		t.Fatal(err)
	}
	return of
}

// TestOracleFactsCommitted: each facts file names the pinned -DDEBUGGING perl,
// holds a record for every oracle-graded file, and Walked is true wherever OK is.
func TestOracleFactsCommitted(t *testing.T) {
	for _, name := range []string{"in-tree.txt", "t1.txt", "perlgit.txt"} {
		of := readOracle(t, name)
		if !strings.Contains(" "+of.Provenance["ccflags"]+" ", " -DDEBUGGING ") {
			t.Errorf("%s: ccflags %q lack -DDEBUGGING", name, of.Provenance["ccflags"])
		}
		if !strings.HasPrefix(of.Provenance["git_commit_id"], pinnedCommit) {
			t.Errorf("%s: git_commit_id %q is not %s", name, of.Provenance["git_commit_id"], pinnedCommit)
		}
		for _, k := range []string{"perl", "version", "cwd", "PERL5LIB", "timeout"} {
			if _, ok := of.Provenance[k]; !ok {
				t.Errorf("%s: header has no %s", name, k)
			}
		}
		have := map[string]bool{}
		for _, r := range of.Records {
			have[r.Key] = true
			if r.Facts.OK && !r.Facts.Walked {
				t.Errorf("%s: %s: OK but not Walked", name, r.Key)
			}
		}
		want, ok := oracleKeys(t)[name]
		if !ok {
			t.Logf("%s: no perl.git corpus; completeness not checked", name)
			continue
		}
		for k := range want {
			if !have[k] {
				t.Errorf("%s: %s: no record", name, k)
			}
		}
	}
}

// TestOracleFactsNoStrayRecords: every record names an oracle-graded file,
// and no file has two records.
func TestOracleFactsNoStrayRecords(t *testing.T) {
	keys := oracleKeys(t)
	for _, name := range []string{"in-tree.txt", "t1.txt", "perlgit.txt"} {
		seen := map[string]bool{}
		for _, r := range readOracle(t, name).Records {
			if seen[r.Key] {
				t.Errorf("%s: %s: two records", name, r.Key)
			}
			seen[r.Key] = true
			if want, ok := keys[name]; ok && !want[r.Key] {
				t.Errorf("%s: %s: not an oracle-graded file", name, r.Key)
			}
		}
	}
}

// TestOracleFactsMalformedRecordFails: a truncated or malformed record is an
// error rather than skipped, and a file with no facts reads as not accepted.
func TestOracleFactsMalformedRecordFails(t *testing.T) {
	good := `a.t	facts	{"ok":true,"walked":true,"op_count":3,"sites":{"srefgen":[1]},"prototypes":{}}`
	for name, rec := range map[string]string{
		"truncated json": good[:len(good)-6],
		"missing fields": "a.t\tfacts",
		"unknown kind":   "a.t\tmaybe\tx",
		"unknown field":  `a.t	facts	{"ok":true,"extra":1}`,
		"trailing junk":  good + ` {}`,
		"empty reason":   "a.t\tnone\t",
		"extra column":   good + "\tx",
	} {
		p := filepath.Join(t.TempDir(), "f.txt")
		if err := os.WriteFile(p, []byte("# git_commit_id: x\n"+rec+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadOracleFacts(p); err == nil {
			t.Errorf("%s: read without error", name)
		}
	}
	p := filepath.Join(t.TempDir(), "f.txt")
	os.WriteFile(p, []byte(good+"\nb.t\tnone\ttimeout after 180s\n"), 0o644)
	of, err := ReadOracleFacts(p)
	if err != nil || len(of.Records) != 2 {
		t.Fatalf("well-formed set: %v, %d records", err, len(of.Records))
	}
	if r := of.Records[0]; !r.Facts.OK || r.Facts.Sites["srefgen"][0] != 1 {
		t.Errorf("good record misread: %+v", r)
	}
	if r := of.Records[1]; r.Facts.OK || r.Missing == "" {
		t.Errorf("a file with no facts reads as accepted: %+v", r)
	}
}

// TestOracleFactsScriptRefusesAMissingPerl pins the refusal's wording: the
// script's name, the missing perl and the tree it was asked to read, exit 1.
func TestOracleFactsScriptRefusesAMissingPerl(t *testing.T) {
	tree := t.TempDir()
	cmd := exec.Command("sh", filepath.Join("offline", "oracle-facts.sh"), "/nonexistent/perl", tree, t.TempDir())
	out, err := cmd.CombinedOutput()
	if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 1 {
		t.Fatalf("exit = %v, want 1; output: %s", err, out)
	}
	want := "oracle-facts.sh: perl /nonexistent/perl is missing or not executable; refusing to read the tree " + tree + "\n"
	if string(out) != want {
		t.Errorf("stderr = %q, want %q", out, want)
	}
}

// TestOracleFactsLabelAcceptedWithoutOKFacts pins the known limits: files
// the labels accept whose facts say ok false. T1/cpan_tooling_regressions.t
// compiles under perl -c but B::Concise dies on it (Can't locate object method
// NAME via package B::NULL, Concise.pm:925). A regeneration that adds another
// is a change to review, not to commit. perl.git t/'s facts cover every .t
// file, so only those perlgit.txt labels accept are in scope:
// perlgit/re/pat_psycho.t has none (its BEGIN forks a watchdog that holds
// perl's stderr past the 180 s timeout); its 26 other ok:false records are
// files the labels refuse.
func TestOracleFactsLabelAcceptedWithoutOKFacts(t *testing.T) {
	keys := oracleKeys(t)
	lf, err := ReadLabels(filepath.Join(labelsDir, "perlgit.txt"))
	if err != nil {
		t.Fatal(err)
	}
	keys["perlgit.txt"] = map[string]bool{}
	for _, l := range lf.Labels {
		keys["perlgit.txt"]["perlgit/"+l.Path] = !l.Refused()
	}
	for name, want := range map[string]map[string]bool{
		"in-tree.txt": {},
		"t1.txt":      {"T1/cpan_tooling_regressions.t": true},
		"perlgit.txt": {"perlgit/re/pat_psycho.t": true},
	} {
		got := map[string]bool{}
		for _, r := range readOracle(t, name).Records {
			if keys[name][r.Key] && !r.Facts.OK {
				got[r.Key] = true
			}
		}
		for k := range got {
			if !want[k] {
				t.Errorf("%s: %s: label-accepted but its facts are not ok; a new limit", name, k)
			}
		}
		for k := range want {
			if !got[k] {
				t.Errorf("%s: %s: the pinned limit is gone; update this test and the README", name, k)
			}
		}
	}
}
