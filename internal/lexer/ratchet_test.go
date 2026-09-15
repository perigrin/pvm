// ABOUTME: The M0 metrics: 100% round-trip over the whole corpus, and a per-file error-token ratchet.
// ABOUTME: Round-trip is a hard gate; the error counts are a ratchet that must move deliberately.

package lexer

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var updateRatchet = flag.Bool("lexer.update-ratchet", false,
	"rewrite the committed per-file error-token baseline")

// corpusFiles walks the perl5 t/ directory, or skips.
func corpusFiles(t *testing.T) (string, []string) {
	t.Helper()
	root := os.Getenv("PERL5_CORPUS")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("no PERL5_CORPUS and no home directory: %v", err)
		}
		root = filepath.Join(home, "dev", "perl5")
	}
	tDir := filepath.Join(root, "t")
	if _, err := os.Stat(tDir); err != nil {
		t.Skipf("perl5 corpus not present at %s: %v\n"+
			"set PERL5_CORPUS to a checkout to run the M0 metrics", tDir, err)
	}

	var files []string
	err := filepath.Walk(tDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".t") {
			rel, err := filepath.Rel(tDir, path)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		t.Skipf("walking the corpus: %v", err)
	}
	sort.Strings(files)
	return tDir, files
}

// t2Dirs are the M0 corpus: spec §7.3's tier-2 core.
var t2Dirs = map[string]bool{
	"base": true, "cmd": true, "comp": true, "opbasic": true, "class": true,
}

// TestT2CoreRoundTrips is M0's headline metric, and it is a HARD gate rather
// than a ratchet: invariant 3 must never fail, on any file, ever.
//
// Asserted over the whole corpus rather than only the 56 T2 files. Lexing 620
// files costs milliseconds, and the wider denominator already paid for
// itself: the survey that preceded this test found comp/parser.t dropping 141
// bytes at EOF -- a heredoc body that ran to the end without its terminator
// and emitted no token for what it had consumed. A 56-file check would have
// caught that one too (comp/ is T2), but only because the bug happened to
// land there.
func TestT2CoreRoundTrips(t *testing.T) {
	tDir, files := corpusFiles(t)

	var lossy []string
	var t2Total, t2Clean int
	for _, rel := range files {
		src, err := os.ReadFile(filepath.Join(tDir, rel))
		if err != nil {
			continue
		}
		var sb strings.Builder
		for _, tok := range Tokenize(src) {
			sb.Write(src[tok.Start:tok.End])
		}
		if sb.String() != string(src) {
			lossy = append(lossy, rel)
		}
		if t2Dirs[strings.SplitN(rel, string(filepath.Separator), 2)[0]] {
			t2Total++
			if sb.String() == string(src) {
				t2Clean++
			}
		}
	}

	if len(lossy) > 0 {
		t.Errorf("%d file(s) do not round-trip; invariant 3 admits no exceptions:\n  %s",
			len(lossy), strings.Join(lossy, "\n  "))
	}
	if t2Total == 0 {
		t.Fatal("no T2 core files found: the corpus layout changed")
	}
	if t2Clean != t2Total {
		t.Errorf("T2 core round-trips %d/%d; M0 requires 100%%", t2Clean, t2Total)
	}

	// The fixtures the plan also names, which need no corpus.
	for _, src := range fuzzSeeds {
		var sb strings.Builder
		for _, tok := range Tokenize([]byte(src)) {
			sb.WriteString(src[tok.Start:tok.End])
		}
		if sb.String() != src {
			t.Errorf("fixture does not round-trip: %q", src)
		}
	}
}

// TestLexerRatchet freezes the per-file count of Error and UnknownRest
// tokens, and fails when any file moves in EITHER direction.
//
// Error COUNTS rather than round-trip status, and that choice is the point.
// A round-trip baseline would be 620 identical `pass` rows the day this
// milestone closes: nothing could ever move and Check would never fire, which
// is a ratchet over a constant. perl-lsp shipped one of those, with a baseline
// containing the line `0`.
//
// An improvement fails too, deliberately. A baseline that silently absorbs
// gains is a baseline nobody reads, and the gain goes unrecorded.
func TestLexerRatchet(t *testing.T) {
	tDir, files := corpusFiles(t)

	now := make(map[string]int, len(files))
	for _, rel := range files {
		src, err := os.ReadFile(filepath.Join(tDir, rel))
		if err != nil {
			continue
		}
		var bad int
		for _, tok := range Tokenize(src) {
			if tok.Kind == Error || tok.Kind == UnknownRest {
				bad++
			}
		}
		now[rel] = bad
	}

	path := filepath.Join("testdata", "corpus.ratchet")
	if *updateRatchet {
		if err := os.WriteFile(path, []byte(renderRatchet(now)), 0o644); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
		t.Logf("rewrote %s (%d files); commit it with the change that moved it",
			path, len(now))
		return
	}

	want, err := parseRatchet(path)
	if err != nil {
		t.Fatalf("%v\nrun with -lexer.update-ratchet to create it", err)
	}

	var regressed, improved, added, removed []string
	for rel, n := range now {
		was, ok := want[rel]
		if !ok {
			added = append(added, rel)
			continue
		}
		switch {
		case n > was:
			regressed = append(regressed, fmt.Sprintf("%s: %d -> %d", rel, was, n))
		case n < was:
			improved = append(improved, fmt.Sprintf("%s: %d -> %d", rel, was, n))
		}
	}
	for rel := range want {
		if _, ok := now[rel]; !ok {
			removed = append(removed, rel)
		}
	}
	sort.Strings(regressed)
	sort.Strings(improved)
	sort.Strings(added)
	sort.Strings(removed)

	if len(regressed) > 0 {
		t.Errorf("%d file(s) lex worse than the baseline:\n  %s",
			len(regressed), strings.Join(regressed, "\n  "))
	}
	if len(improved) > 0 {
		t.Errorf("%d file(s) lex BETTER than the baseline:\n  %s\n\n"+
			"This is good news and still fails: re-run with -lexer.update-ratchet "+
			"and commit the baseline with the change that earned it.",
			len(improved), strings.Join(improved, "\n  "))
	}
	if len(added) > 0 || len(removed) > 0 {
		t.Errorf("the corpus changed shape: %d added, %d removed\n  added: %v\n  removed: %v",
			len(added), len(removed), added, removed)
	}
}

// TestSkipsWithoutCorpus: the corpus-dependent tests must SKIP where perl5 is
// absent, not fail. ci.yml runs `go test ./...` with no checkout, and a
// missing corpus reading as a lexer defect is the environmental-fact-as-code-
// fact class this project keeps finding.
func TestSkipsWithoutCorpus(t *testing.T) {
	t.Setenv("PERL5_CORPUS", filepath.Join(t.TempDir(), "absent"))
	var skipped bool
	t.Run("resolve", func(st *testing.T) {
		defer func() { skipped = st.Skipped() }()
		corpusFiles(st)
	})
	if !skipped {
		t.Error("corpusFiles did not skip with the corpus absent")
	}
}

// renderRatchet formats the baseline: one `count path` line per file, sorted.
func renderRatchet(counts map[string]int) string {
	var sb strings.Builder
	sb.WriteString("# ABOUTME: Frozen per-file count of Error and UnknownRest tokens over the perl5 corpus.\n")
	sb.WriteString("# ABOUTME: Regenerated by `go test ./internal/lexer/ -run TestLexerRatchet -lexer.update-ratchet`.\n")
	sb.WriteString("#\n")
	sb.WriteString("# DO NOT EDIT BY HAND. A file that lexes BETTER fails the build too,\n")
	sb.WriteString("# on purpose: the baseline must never silently diverge from reality in\n")
	sb.WriteString("# either direction, so re-running with the flag in the same commit is\n")
	sb.WriteString("# how a gain is recorded.\n")
	paths := make([]string, 0, len(counts))
	for p := range counts {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		fmt.Fprintf(&sb, "%5d %s\n", counts[p], p)
	}
	return sb.String()
}

// parseRatchet reads the committed baseline.
func parseRatchet(path string) (map[string]int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	out := map[string]int{}
	for n, line := range strings.Split(string(data), "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		count, rest, ok := strings.Cut(line, " ")
		if !ok {
			return nil, fmt.Errorf("%s line %d: want `count path`, got %q", path, n+1, line)
		}
		c, err := strconv.Atoi(count)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %q is not a count", path, n+1, count)
		}
		out[strings.TrimSpace(rest)] = c
	}
	return out, nil
}
