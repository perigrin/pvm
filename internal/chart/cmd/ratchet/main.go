// ABOUTME: The chart ratchet: measures the chart spike and recursive descent's baselines, prints the contract lines.
// ABOUTME: Run by `make chart-ratchet`; exits non-zero only on a regression against testdata/ratchet/baseline.txt.

// Command ratchet is the chart spike's ratchet. Its output contract is in
// internal/chart/README.md. It reads only the working tree, from the
// repository root: no commit, no refs/zhi, no network.
package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"tamarou.com/pvm/internal/chart"
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
)

var (
	testdata     = filepath.Join("internal", "chart", "testdata")
	guacamole    = filepath.Join(testdata, "guacamole")
	parsesDir    = filepath.Join(guacamole, "parses")
	parsentDir   = filepath.Join(guacamole, "parsent")
	rejectsFile  = filepath.Join(guacamole, "perl-rejects.txt")
	stubsDir     = filepath.Join(guacamole, "stubs")
	std5kFile    = filepath.Join(testdata, "std5k.pl")
	editsFile    = filepath.Join(testdata, "convergence-edits.jsonl")
	defectList   = filepath.Join(testdata, "defect-yield.txt")
	baselineFile = filepath.Join(testdata, "ratchet", "baseline.txt")
	grammarSLIF  = filepath.Join("internal", "chart", "grammar", "testdata", "guacamole", "guacamole.slif")
	tablesDir    = filepath.Join("internal", "chart", "grammar", "testdata", "tables")
)

// keys is the fixed print order: the milestone's contract keys, then the
// added ones. A key may be added here and is never renamed.
var keys = []string{
	"parses-scored", "parsent-scored",
	"parses-refused", "parses-wrong-sites", "yield-violations", "oracle-below-rd",
	"std5k-refused", "parsent-accepted", "rejects-disagree",
	"speed-ratio-aggregate", "speed-ratio-std5k", "set-reuse",
	"convergence-edits", "convergence-within-8",
	"defect-yield-listed", "defect-yield-sites", "defect-yield-settled",
	"oracle-sites-decided", "expect-states-unreproduced",
	"rd-oracle-exact", "rd-oracle-coverage", "rd-parse-ms-std5k", "rd-parse-ms-corpus",
	"tables-drift", "core-tpm-differences",
}

func main() {
	measured, err := measure(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "chart-ratchet: %v\n", err)
		os.Exit(2)
	}
	ms := report(measured)
	fmt.Print(Format(ms))

	text, err := os.ReadFile(baselineFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "chart-ratchet: reading the baseline: %v\n", err)
		os.Exit(2)
	}
	baseline, err := Parse(string(text))
	if err != nil {
		fmt.Fprintf(os.Stderr, "chart-ratchet: %s: %v\n", baselineFile, err)
		os.Exit(2)
	}
	regressions := Compare(baseline, ms)
	for _, r := range regressions {
		fmt.Fprintf(os.Stderr, "chart-ratchet: %s\n", r)
	}
	if len(regressions) > 0 {
		os.Exit(1)
	}
}

// report lays measured values out in the fixed order, na for every key
// nothing measured yet. A measured key with no slot is a bug, so it panics.
func report(measured map[string]string) []Measure {
	for k := range measured {
		if !slices.Contains(keys, k) {
			panic("chart-ratchet: measured key " + k + " has no place in the print order")
		}
	}
	ms := make([]Measure, len(keys))
	for i, k := range keys {
		v, ok := measured[k]
		if !ok {
			v = na
		}
		ms[i] = Measure{k, v}
	}
	return ms
}

// measure takes every measurement the skeleton has a component for.
func measure(ctx context.Context) (map[string]string, error) {
	m := map[string]string{}

	// The stubs reach perl, and every compile perl starts, through PERL5LIB.
	// Set here rather than by the caller: the oracle cache key omits the
	// environment, so a run without the stubs would poison the cache.
	stubs, err := filepath.Abs(stubsDir)
	if err != nil {
		return nil, err
	}
	os.Setenv("PERL5LIB", stubs)

	parses, err := filepath.Glob(filepath.Join(parsesDir, "*.pl"))
	if err != nil {
		return nil, err
	}
	parsent, err := filepath.Glob(filepath.Join(parsentDir, "*.pl"))
	if err != nil {
		return nil, err
	}
	rejected, err := lines(rejectsFile)
	if err != nil {
		return nil, err
	}
	isRejected := map[string]bool{}
	for _, r := range rejected {
		isRejected[r] = true
	}
	var scored []string
	for _, p := range parses {
		if !isRejected[filepath.Base(p)] {
			scored = append(scored, p)
		}
	}
	m["parses-scored"] = strconv.Itoa(len(scored))
	m["parsent-scored"] = strconv.Itoa(len(parsent))

	disagree, err := chart.RejectsDisagree("perl", parsesDir, rejectsFile, stubs)
	if err != nil {
		return nil, fmt.Errorf("rejects-disagree: %w", err)
	}
	for _, d := range disagree {
		fmt.Fprintf(os.Stderr, "chart-ratchet: rejects-disagree: %s\n", d)
	}
	m["rejects-disagree"] = strconv.Itoa(len(disagree))

	edits, err := lines(editsFile)
	if err != nil {
		return nil, fmt.Errorf("convergence-edits: %w", err)
	}
	m["convergence-edits"] = strconv.Itoa(len(edits))

	listed, err := lines(defectList)
	if err != nil {
		return nil, fmt.Errorf("defect-yield-listed: %w", err)
	}
	m["defect-yield-listed"] = strconv.Itoa(len(listed))

	drift, err := tablesDrift(tablesDir, grammarSLIF)
	if err != nil {
		return nil, fmt.Errorf("tables-drift: %w", err)
	}
	for _, d := range drift {
		fmt.Fprintf(os.Stderr, "chart-ratchet: tables-drift: %s\n", d)
	}
	m["tables-drift"] = strconv.Itoa(len(drift))

	diffs, err := grammar.GuacamoleCoreTpmDifferences(grammarSLIF)
	if err != nil {
		return nil, fmt.Errorf("core-tpm-differences: %w", err)
	}
	m["core-tpm-differences"] = strconv.Itoa(len(diffs))

	std5k, err := os.ReadFile(std5kFile)
	if err != nil {
		return nil, fmt.Errorf("std5k: %w", err)
	}
	m["rd-parse-ms-std5k"] = millis(bestOf(5, func() { parse.Parse(std5k) }))
	srcs := make([][]byte, len(scored))
	for i, p := range scored {
		if srcs[i], err = os.ReadFile(p); err != nil {
			return nil, err
		}
	}
	m["rd-parse-ms-corpus"] = millis(bestOf(3, func() {
		for _, s := range srcs {
			parse.Parse(s)
		}
	}))

	// Recursive descent through the oracle, its default subject.
	dir, err := oracleCacheDir(stubs)
	if err != nil {
		return nil, err
	}
	cache, err := parseoracle.OpenCache(dir, "")
	if err != nil {
		return nil, err
	}
	rep, err := parseoracle.Run(ctx, scored, parseoracle.RunOptions{Cache: cache})
	if err != nil {
		return nil, fmt.Errorf("rd oracle: %w", err)
	}
	for _, r := range rep.Files {
		if r.Err != "" {
			fmt.Fprintf(os.Stderr, "chart-ratchet: rd oracle runner error: %s: %s\n", r.Path, r.Err)
		}
	}
	if n := rep.Measured(); n > 0 {
		m["rd-oracle-exact"] = rate(rep.Totals[parseoracle.BucketExact], n)
		m["rd-oracle-coverage"] = rate(n-rep.Totals[parseoracle.BucketNoAnswer], n)
	}
	return m, nil
}

// tablesDrift regenerates the tables from slif into a temporary directory
// and names each file that differs from the committed ones in dir.
func tablesDrift(dir, slif string) ([]string, error) {
	files, err := grammar.GenerateTables(slif)
	if err != nil {
		return nil, err
	}
	tmp, err := os.MkdirTemp("", "chart-tables-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	if err := grammar.WriteTables(tmp, files); err != nil {
		return nil, err
	}
	return grammar.TablesDrift(dir, tmp)
}

func rate(k, n int) string {
	return strconv.FormatFloat(float64(k)/float64(n), 'f', 4, 64)
}

// oracleCacheDir is the ratchet's own oracle cache, never DefaultCacheDir:
// the cache key omits the environment, and these answers depend on PERL5LIB,
// so the directory is keyed on a hash of the stubs it names.
func oracleCacheDir(stubs string) (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("locating the oracle cache: %w", err)
	}
	h := sha256.New()
	err = filepath.WalkDir(stubs, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(stubs, path)
		fmt.Fprintf(h, "%s\x00%d\x00", filepath.ToSlash(rel), len(b))
		h.Write(b)
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("hashing the stubs: %w", err)
	}
	return filepath.Join(base, "pvm", "chart-ratchet-oracle", hex.EncodeToString(h.Sum(nil))[:16]), nil
}

// lines returns a file's non-empty lines that are not # comments.
func lines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		if l := strings.TrimSpace(s.Text()); l != "" && !strings.HasPrefix(l, "#") {
			out = append(out, l)
		}
	}
	return out, s.Err()
}

func bestOf(n int, f func()) time.Duration {
	best := time.Duration(1<<63 - 1)
	for i := 0; i < n; i++ {
		start := time.Now()
		f()
		if d := time.Since(start); d < best {
			best = d
		}
	}
	return best
}

func millis(d time.Duration) string {
	return strconv.FormatFloat(float64(d.Microseconds())/1000, 'f', 3, 64)
}
