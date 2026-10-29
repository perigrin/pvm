// ABOUTME: The chart ratchet: measures the chart spike and recursive descent's baselines, prints the contract lines.
// ABOUTME: Run by `make chart-ratchet`; exits non-zero only on a regression against testdata/ratchet/baseline.txt.

// Command ratchet is the chart spike's ratchet. Its output contract is in
// internal/chart/README.md. It reads only the working tree, from the
// repository root: no commit, no refs/zhi, no network.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"tamarou.com/pvm/internal/chart"
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
	defectDir    = filepath.Join(testdata, "defect-yield")
	baselineFile = filepath.Join(testdata, "ratchet", "baseline.txt")
)

// keys is the fixed print order: the milestone's contract keys, then the
// added ones. A key may be added here and is never renamed.
var keys = []string{
	"parses-scored", "parsent-scored",
	"parses-refused", "parses-wrong-sites", "yield-violations", "oracle-below-rd",
	"std5k-refused", "parsent-accepted", "rejects-disagree",
	"speed-ratio-aggregate", "speed-ratio-std5k", "set-reuse",
	"convergence-edits", "convergence-within-8",
	"defect-yield-listed", "defect-yield-sites", "defect-yield-settled", "defect-yield-rd-settled",
	"oracle-sites-decided", "expect-states-unreproduced",
	"rd-oracle-exact", "rd-oracle-coverage", "rd-parse-ms-std5k", "rd-parse-ms-corpus",
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
	ms := make([]Measure, len(keys))
	for i, k := range keys {
		v, ok := measured[k]
		if !ok {
			v = na
		}
		ms[i] = Measure{k, v}
		delete(measured, k)
	}
	for k := range measured {
		panic("chart-ratchet: measured key " + k + " has no place in the print order")
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

	// Recursive descent through the oracle: the scored parses cases, and
	// the defect-yield list's frozen cases, in one run.
	var cases []string
	for _, l := range listed {
		ref := strings.Fields(l)[0]
		c := filepath.Join(defectDir, ref+".pl")
		if _, err := os.Stat(c); err == nil {
			cases = append(cases, c)
		}
	}
	cache, err := parseoracle.OpenCache(oracleCacheDir(), "")
	if err != nil {
		return nil, err
	}
	rep, err := parseoracle.Run(ctx, append(append([]string{}, scored...), cases...), parseoracle.RunOptions{Cache: cache})
	if err != nil {
		return nil, fmt.Errorf("rd oracle: %w", err)
	}
	isCase := map[string]bool{}
	for _, c := range cases {
		isCase[c] = true
	}
	var measuredN, exact, answered, settled int
	for _, r := range rep.Files {
		if r.Err != "" {
			fmt.Fprintf(os.Stderr, "chart-ratchet: rd oracle runner error: %s: %s\n", r.Path, r.Err)
		}
		if r.Excluded || r.Err != "" {
			continue
		}
		if isCase[r.Path] {
			// Settled by recursive descent: graded exact with a marker in play.
			if r.Verdict.Bucket == parseoracle.BucketExact && sites(r.Facts) > 0 {
				settled++
			}
			continue
		}
		measuredN++
		if r.Verdict.Bucket == parseoracle.BucketExact {
			exact++
		}
		if r.Verdict.Bucket != parseoracle.BucketNoAnswer {
			answered++
		}
	}
	if measuredN > 0 {
		m["rd-oracle-exact"] = strconv.FormatFloat(float64(exact)/float64(measuredN), 'f', 4, 64)
		m["rd-oracle-coverage"] = strconv.FormatFloat(float64(answered)/float64(measuredN), 'f', 4, 64)
	}
	m["defect-yield-rd-settled"] = strconv.Itoa(settled)
	return m, nil
}

// oracleCacheDir is the ratchet's own oracle cache, never DefaultCacheDir:
// the cache key omits the environment, and these answers depend on PERL5LIB.
func oracleCacheDir() string {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		base = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return filepath.Join(base, "pvm", "chart-ratchet-oracle")
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
	s.Buffer(nil, 1<<20)
	for s.Scan() {
		if l := strings.TrimSpace(s.Text()); l != "" && !strings.HasPrefix(l, "#") {
			out = append(out, l)
		}
	}
	return out, s.Err()
}

func sites(f parseoracle.Facts) int {
	n := 0
	for _, lines := range f.Sites {
		n += len(lines)
	}
	return n
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
