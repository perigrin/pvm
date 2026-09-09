// ABOUTME: Tests the content-hash cache that lets a corpus sweep gate a commit.
// ABOUTME: Every AC here is a `go test -run` invocation, named to match the issue.

package parseoracle_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// countingPerl puts a perl on PATH that records every invocation, so "spawns
// no perl" is proved rather than inferred from a stopwatch.
func countingPerl(t *testing.T) (log string) {
	t.Helper()

	realPerl, err := exec.LookPath("perl")
	if err != nil {
		t.Skipf("perl unavailable: %v", err)
	}

	bin := t.TempDir()
	log = filepath.Join(bin, "invocations")
	shim := "#!/bin/sh\necho x >> " + log + "\nexec " + realPerl + " \"$@\"\n"
	if err := os.WriteFile(filepath.Join(bin, "perl"), []byte(shim), 0o755); err != nil {
		t.Fatalf("writing counting perl: %v", err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return log
}

// perlInvocations counts what countingPerl recorded.
func perlInvocations(t *testing.T, log string) int {
	t.Helper()
	data, err := os.ReadFile(log)
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatalf("reading invocation log: %v", err)
	}
	return strings.Count(string(data), "x")
}

// TestCacheSecondSweepSpawnsNoPerl is the issue's first AC. A sweep over an
// unchanged corpus must read every answer from disk, because six minutes of
// perl is what stops the ratchet from gating a commit.
func TestCacheSecondSweepSpawnsNoPerl(t *testing.T) {
	files := fixtureCorpus(t)
	cache, err := parseoracle.OpenCache(t.TempDir())
	if err != nil {
		t.Fatalf("OpenCache: %v", err)
	}

	if _, err := parseoracle.Run(context.Background(), files,
		parseoracle.RunOptions{Cache: cache}); err != nil {
		t.Fatalf("cold Run: %v", err)
	}

	// Only the second sweep is counted: the first legitimately spawns perl.
	log := countingPerl(t)
	warm, err := parseoracle.Run(context.Background(), files,
		parseoracle.RunOptions{Cache: cache})
	if err != nil {
		t.Fatalf("warm Run: %v", err)
	}

	if n := perlInvocations(t, log); n != 0 {
		t.Errorf("a warm sweep spawned perl %d time(s); the cache is not being read", n)
	}
	if got, want := len(warm.Files), len(files); got != want {
		t.Fatalf("warm sweep covered %d files, want %d", got, want)
	}
}

// TestCacheServesIdenticalFacts guards the thing that makes a cache worth
// having at all: the cached answer must be the answer perl gave. A fast cache
// that returns something else is worse than no cache.
func TestCacheServesIdenticalFacts(t *testing.T) {
	dir := t.TempDir()
	cache, err := parseoracle.OpenCache(dir)
	if err != nil {
		t.Fatalf("OpenCache: %v", err)
	}

	src := []byte("sub f(\\@){}\nmy @a;\nf(@a);\n")
	path := filepath.Join(t.TempDir(), "probe.pl")
	if err := os.WriteFile(path, src, 0o644); err != nil {
		t.Fatalf("writing probe: %v", err)
	}

	cold, err := cache.AskFile(context.Background(), path, parseoracle.Options{})
	if err != nil {
		t.Skipf("perl unavailable: %v", err)
	}
	warm, err := cache.AskFile(context.Background(), path, parseoracle.Options{})
	if err != nil {
		t.Fatalf("warm AskFile: %v", err)
	}

	if cold.OK != warm.OK || cold.OpCount != warm.OpCount || cold.Srefgen != warm.Srefgen {
		t.Fatalf("cache returned different facts:\ncold %+v\nwarm %+v", cold, warm)
	}
	if strings.Join(cold.Ops, ",") != strings.Join(warm.Ops, ",") {
		t.Errorf("cached ops differ from perl's:\ncold %v\nwarm %v", cold.Ops, warm.Ops)
	}
	if cold.Prototypes["f"] != warm.Prototypes["f"] {
		t.Errorf("cached prototypes differ: %q vs %q", cold.Prototypes["f"], warm.Prototypes["f"])
	}
	// The prototype fact is the whole point of the oracle; a cache that
	// round-trips an empty map would pass the equality check above.
	if warm.Prototypes["f"] != `\@` {
		t.Errorf("cached prototype of f = %q, want %q", warm.Prototypes["f"], `\@`)
	}
}

// TestCacheMissesOnChangedSource is half of the issue's second AC. Editing the
// file must change the answer, or the cache is a way to keep believing a fact
// that stopped being true.
func TestCacheMissesOnChangedSource(t *testing.T) {
	cache, err := parseoracle.OpenCache(t.TempDir())
	if err != nil {
		t.Fatalf("OpenCache: %v", err)
	}

	path := filepath.Join(t.TempDir(), "probe.pl")
	if err := os.WriteFile(path, []byte("my $x = 1;\n"), 0o644); err != nil {
		t.Fatalf("writing probe: %v", err)
	}
	if _, err := cache.AskFile(context.Background(), path, parseoracle.Options{}); err != nil {
		t.Skipf("perl unavailable: %v", err)
	}

	// Same path, different bytes: the key is content, not name.
	if err := os.WriteFile(path, []byte("sub g(\\@){}\nmy @a;\ng(@a);\n"), 0o644); err != nil {
		t.Fatalf("rewriting probe: %v", err)
	}

	log := countingPerl(t)
	facts, err := cache.AskFile(context.Background(), path, parseoracle.Options{})
	if err != nil {
		t.Fatalf("AskFile after edit: %v", err)
	}
	if n := perlInvocations(t, log); n == 0 {
		t.Error("changed source served from cache; the key is not content-addressed")
	}
	if facts.Prototypes["g"] != `\@` {
		t.Errorf("stale facts after edit: prototype of g = %q, want %q", facts.Prototypes["g"], `\@`)
	}
}

// TestCacheMissesOnChangedInterpreter is the other half. Perl's answer for a
// given source depends on which perl was asked, so the interpreter version is
// in the key — otherwise a re-baseline onto a new perl silently reuses the old
// perl's opinions.
func TestCacheMissesOnChangedInterpreter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(t.TempDir(), "probe.pl")
	if err := os.WriteFile(path, []byte("my $x = 1;\n"), 0o644); err != nil {
		t.Fatalf("writing probe: %v", err)
	}

	first, err := parseoracle.OpenCacheFor(dir, parseoracle.CacheIdentity{
		Interpreter: "5.042000",
		Revision:    "deadbeef",
		Script:      "scripthash",
	})
	if err != nil {
		t.Fatalf("OpenCacheFor: %v", err)
	}
	if _, err := first.AskFile(context.Background(), path, parseoracle.Options{}); err != nil {
		t.Skipf("perl unavailable: %v", err)
	}

	second, err := parseoracle.OpenCacheFor(dir, parseoracle.CacheIdentity{
		Interpreter: "5.044000", // the only thing that moved
		Revision:    "deadbeef",
		Script:      "scripthash",
	})
	if err != nil {
		t.Fatalf("OpenCacheFor: %v", err)
	}

	log := countingPerl(t)
	if _, err := second.AskFile(context.Background(), path, parseoracle.Options{}); err != nil {
		t.Fatalf("AskFile under a new interpreter: %v", err)
	}
	if n := perlInvocations(t, log); n == 0 {
		t.Error("a new interpreter version hit the cache; the version is not in the key")
	}
}

// TestCacheMissesOnChangedScript is gap G5 from the chain review.
// parse_facts.pl is the measuring instrument. Changing what it measures
// invalidates every prior answer, and a key that omits it returns facts the
// current script would never have produced.
func TestCacheMissesOnChangedScript(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(t.TempDir(), "probe.pl")
	if err := os.WriteFile(path, []byte("my $x = 1;\n"), 0o644); err != nil {
		t.Fatalf("writing probe: %v", err)
	}

	id := parseoracle.CacheIdentity{
		Interpreter: "5.042000",
		Revision:    "deadbeef",
		Script:      "scripthash",
	}
	first, err := parseoracle.OpenCacheFor(dir, id)
	if err != nil {
		t.Fatalf("OpenCacheFor: %v", err)
	}
	if _, err := first.AskFile(context.Background(), path, parseoracle.Options{}); err != nil {
		t.Skipf("perl unavailable: %v", err)
	}

	id.Script = "a-different-parse_facts.pl"
	second, err := parseoracle.OpenCacheFor(dir, id)
	if err != nil {
		t.Fatalf("OpenCacheFor: %v", err)
	}

	log := countingPerl(t)
	if _, err := second.AskFile(context.Background(), path, parseoracle.Options{}); err != nil {
		t.Fatalf("AskFile under a new script: %v", err)
	}
	if n := perlInvocations(t, log); n == 0 {
		t.Error("a changed parse_facts.pl hit the cache; the script hash is not in the key " +
			"— every prior answer would silently survive a change to what is measured")
	}
}

// TestCacheKeyCoversInvocationOptions: -T changes what perl accepts, and the
// working directory changes what @INC finds. Two files with identical bytes
// asked under different options are different questions.
func TestCacheKeyCoversInvocationOptions(t *testing.T) {
	cache, err := parseoracle.OpenCache(t.TempDir())
	if err != nil {
		t.Fatalf("OpenCache: %v", err)
	}

	path := filepath.Join(t.TempDir(), "probe.pl")
	if err := os.WriteFile(path, []byte("my $x = 1;\n"), 0o644); err != nil {
		t.Fatalf("writing probe: %v", err)
	}
	if _, err := cache.AskFile(context.Background(), path, parseoracle.Options{}); err != nil {
		t.Skipf("perl unavailable: %v", err)
	}

	log := countingPerl(t)
	if _, err := cache.AskFile(context.Background(), path,
		parseoracle.Options{Shebang: true}); err != nil {
		t.Fatalf("AskFile under taint: %v", err)
	}
	if n := perlInvocations(t, log); n == 0 {
		t.Error("taint mode hit the non-taint cache entry; the options are not in the key")
	}
}

// TestCacheEntriesAreReadableDiffs is the issue's fourth AC. The cache is
// committed so a reviewer can read a change to it as a change in what perl
// said, which requires the stored form be diffable rather than a blob.
func TestCacheEntriesAreReadableDiffs(t *testing.T) {
	dir := t.TempDir()
	cache, err := parseoracle.OpenCache(dir)
	if err != nil {
		t.Fatalf("OpenCache: %v", err)
	}

	path := filepath.Join(t.TempDir(), "probe.pl")
	if err := os.WriteFile(path, []byte("sub f(\\@){}\nmy @a;\nf(@a);\n"), 0o644); err != nil {
		t.Fatalf("writing probe: %v", err)
	}
	if _, err := cache.AskFile(context.Background(), path, parseoracle.Options{}); err != nil {
		t.Skipf("perl unavailable: %v", err)
	}

	var entries []string
	err = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			entries = append(entries, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking cache: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("cache holds %d file(s) after one question, want 1: %v", len(entries), entries)
	}

	data, err := os.ReadFile(entries[0])
	if err != nil {
		t.Fatalf("reading cache entry: %v", err)
	}
	text := string(data)

	// Diffable means line-oriented. A single-line JSON blob renders as one
	// changed line no matter what moved inside it.
	if strings.Count(text, "\n") < 5 {
		t.Errorf("cache entry is %d line(s); a reviewer cannot read a one-line blob as a diff:\n%s",
			strings.Count(text, "\n")+1, text)
	}
	// The entry must say what perl said, not merely that it was asked.
	if !strings.Contains(text, "srefgen") {
		t.Errorf("cache entry does not record the ops perl reported:\n%s", text)
	}
	// And it must name its own question, or a reviewer reading the diff
	// cannot tell which file's answer changed.
	if !strings.Contains(text, "probe.pl") {
		t.Errorf("cache entry does not name the source it answers for:\n%s", text)
	}
}

// TestCacheSurvivesCorruptEntry: a truncated or hand-edited entry must fall
// back to asking perl. A cache that fails the run on bad bytes turns a
// throwaway artifact into something that can break a build.
func TestCacheSurvivesCorruptEntry(t *testing.T) {
	dir := t.TempDir()
	cache, err := parseoracle.OpenCache(dir)
	if err != nil {
		t.Fatalf("OpenCache: %v", err)
	}

	path := filepath.Join(t.TempDir(), "probe.pl")
	if err := os.WriteFile(path, []byte("my $x = 1;\n"), 0o644); err != nil {
		t.Fatalf("writing probe: %v", err)
	}
	want, err := cache.AskFile(context.Background(), path, parseoracle.Options{})
	if err != nil {
		t.Skipf("perl unavailable: %v", err)
	}

	filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			os.WriteFile(p, []byte("{ truncated"), 0o644)
		}
		return nil
	})

	got, err := cache.AskFile(context.Background(), path, parseoracle.Options{})
	if err != nil {
		t.Fatalf("a corrupt cache entry must fall back to perl, got: %v", err)
	}
	if got.OpCount != want.OpCount {
		t.Errorf("recovered facts differ: got %d ops, want %d", got.OpCount, want.OpCount)
	}
}

// TestDefaultCacheDirIsGitIgnored enforces the commit/don't-commit decision
// rather than only documenting it.
//
// The conformance plan assumed a committed cache so CI would spawn no perl.
// The measurement removed most of that argument: the oracle phase is ~35s
// cold across the whole corpus, against carrying 620 files of perl output in
// the repo forever and re-churning them on every re-baseline. So the cache is
// generated locally, and this test fails if someone later stages one.
func TestDefaultCacheDirIsGitIgnored(t *testing.T) {
	root, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Skipf("not a git checkout: %v", err)
	}
	dir := filepath.Join(strings.TrimSpace(string(root)), parseoracle.DefaultCacheDir)

	out, err := exec.Command("git", "check-ignore", dir).CombinedOutput()
	if err != nil {
		t.Errorf("%s is not gitignored (%v): a generated cache must not be "+
			"committable by accident\n%s", parseoracle.DefaultCacheDir, err, out)
	}
}
