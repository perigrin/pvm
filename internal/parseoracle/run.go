// ABOUTME: Runs the oracle across a corpus and reports how often our parser agrees with perl.
// ABOUTME: Environmental failures leave the denominator, so the parser is never blamed for the machine.

package parseoracle

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"

	"tamarou.com/pvm/internal/parser"
)

// Result is one corpus file's outcome.
type Result struct {
	// Path is the file as the runner was given it, so a report can be
	// diffed against another run's.
	Path string
	// Facts is what perl said, empty when the oracle itself failed.
	Facts Facts
	// Failure classifies Facts.Stderr. NoFailure means the file compiled.
	Failure FailureKind
	// Excluded reports that this file is out of the denominator. Only
	// Environmental failures qualify: a missing Config.pm is a fact about
	// this machine, not about how Perl parses.
	Excluded bool
	// Verdict is the comparison, meaningless when Excluded or Err != "".
	Verdict Verdict
	// Err is a runner-level failure — perl unspawnable, file unreadable —
	// as a string rather than an error so a Report renders deterministically.
	Err string
}

// Totals counts verdicts per bucket, indexed by Bucket.
type Totals [4]int

// Sum is the denominator: every file that produced a verdict.
func (t Totals) Sum() int {
	n := 0
	for _, c := range t {
		n += c
	}
	return n
}

// Report is a whole corpus run.
type Report struct {
	// Files is every file's result, sorted by path so two runs of the same
	// corpus render identically despite the workers finishing out of order.
	Files []Result
	// Totals counts the files that stayed in the denominator.
	Totals Totals
	// Environmental counts the files excluded for being about this machine.
	// Reported separately rather than folded into no-answer, because the two
	// mean opposite things: one is our parser declining, the other is a
	// missing module.
	Environmental int
	// Errors counts files the runner could not measure at all.
	Errors int
}

// Measured is how many files produced a verdict.
func (r Report) Measured() int { return r.Totals.Sum() }

// ExactRate is the share of measured files that agree with perl exactly.
//
// This is the number a report must assert a FLOOR on, and asserting only a
// ceiling on WRONG is the trap spec §7.1.5 warns about: our grammar emits
// `ambiguous_function_call_expression`, which honestly declines to commit, so
// every prototype-driven call buckets wider rather than WRONG. A parser that
// hedged on everything would score 100% non-WRONG and mean nothing.
//
// no-answer stays in the denominator for the same reason — otherwise a parser
// that emits an error node for all but one file scores 100% exact.
func (r Report) ExactRate() float64 {
	n := r.Measured()
	if n == 0 {
		return 0
	}
	return float64(r.Totals[BucketExact]) / float64(n)
}

// Wrong names every file we committed to a parse perl did not make. The list
// is more useful than the count: a report that says "3 WRONG" sends a reader
// back to the corpus to work out which three.
func (r Report) Wrong() []string {
	var names []string
	for _, f := range r.Files {
		if !f.Excluded && f.Err == "" && f.Verdict.Bucket == BucketWrong {
			names = append(names, f.Path)
		}
	}
	return names
}

// String renders the human-readable table. It is deterministic — same corpus,
// same output bytes — because the ratchet issue commits it as a baseline.
func (r Report) String() string {
	var b strings.Builder

	n := r.Measured()
	fmt.Fprintf(&b, "measured %d file(s)\n", n)
	for _, bucket := range []Bucket{BucketExact, BucketWider, BucketWrong, BucketNoAnswer} {
		count := r.Totals[bucket]
		fmt.Fprintf(&b, "  %-10s %5d  %s\n", bucket, count, percent(count, n))
	}
	fmt.Fprintf(&b, "excluded (environmental) %d\n", r.Environmental)
	fmt.Fprintf(&b, "runner errors %d\n", r.Errors)

	// The exact rate is printed on its own line because it is the figure a
	// ratchet floors, and a ceiling on WRONG alone proves nothing.
	fmt.Fprintf(&b, "exact rate %s\n", percent(r.Totals[BucketExact], n))

	if wrong := r.Wrong(); len(wrong) > 0 {
		b.WriteString("WRONG files:\n")
		for _, f := range r.Files {
			if !f.Excluded && f.Err == "" && f.Verdict.Bucket == BucketWrong {
				fmt.Fprintf(&b, "  %s: %s\n", f.Path, f.Verdict.Detail)
			}
		}
	}
	return b.String()
}

func percent(count, total int) string {
	if total == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.1f%%", 100*float64(count)/float64(total))
}

// RunOptions controls a corpus run.
type RunOptions struct {
	// Dir is the working directory every file is compiled from, and for
	// perl's own test suite it is mandatory: t/test.pl clears @INC and
	// unshifts ../lib, so a corpus file compiles only from inside a
	// populated shim t/. Measured on op/sub.t, the difference is ok:0 ops:0
	// without it and ok:1 ops:1055 with it, across 498 of 620 files.
	Dir string
	// Workers bounds concurrent perl invocations. Zero means NumCPU.
	Workers int
}

// Run compiles every file in the corpus through perl, parses it with our
// parser, and buckets the two against each other.
//
// Concurrency is a bounded channel semaphore rather than errgroup: go.mod has
// no golang.org/x/sync and this needs no new dependency.
func Run(ctx context.Context, files []string, opts RunOptions) (Report, error) {
	workers := opts.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	results := make([]Result, len(files))
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	for i, path := range files {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i] = measure(ctx, path, opts)
		}(i, path)
	}
	wg.Wait()

	if err := ctx.Err(); err != nil {
		return Report{}, err
	}

	report := Report{Files: results}
	// Sort by path: the workers finish out of order, and a report that
	// reorders between runs cannot be committed as a baseline.
	sort.Slice(report.Files, func(a, b int) bool {
		return report.Files[a].Path < report.Files[b].Path
	})

	for _, r := range report.Files {
		switch {
		case r.Err != "":
			report.Errors++
		case r.Excluded:
			report.Environmental++
		default:
			report.Totals[r.Verdict.Bucket]++
		}
	}
	return report, nil
}

// measure is one file: ask perl, parse ourselves, compare.
//
// Every parser.Parser is created here rather than shared, because a
// tree-sitter parser holds mutable scan state and this runs concurrently.
func measure(ctx context.Context, path string, opts RunOptions) Result {
	r := Result{Path: path}

	src, err := os.ReadFile(resolve(path, opts.Dir))
	if err != nil {
		r.Err = fmt.Sprintf("reading %s: %v", path, err)
		return r
	}

	facts, err := AskFile(ctx, path, Options{Dir: opts.Dir, Shebang: wantsTaint(src)})
	if err != nil {
		r.Err = err.Error()
		return r
	}
	r.Facts = facts
	r.Failure = Classify(facts.Stderr)

	// A missing module says nothing about the parser, so it leaves the
	// denominator. VersionSkew and SyntaxError stay: they are real facts
	// about what this interpreter parses, and Compare buckets them
	// no-answer because there is no ground truth to compare against.
	if r.Failure == Environmental {
		r.Excluded = true
		return r
	}

	tree, err := parser.New().Parse(src)
	if err != nil {
		r.Err = fmt.Sprintf("our parser failed on %s: %v", path, err)
		return r
	}
	r.Verdict = Compare(facts, tree, src)
	return r
}

// resolve mirrors AskFile: a relative path is relative to Dir, which is how
// perl's own tests refer to each other.
func resolve(path, dir string) string {
	if dir == "" || strings.HasPrefix(path, "/") {
		return path
	}
	return dir + "/" + path
}

// wantsTaint reports whether a file's shebang asks for taint mode. Perl
// refuses to compile such a file unless -T is also on the command line, and
// that refusal looks like a syntax error while being nothing of the kind.
func wantsTaint(src []byte) bool {
	line, _, _ := strings.Cut(string(src), "\n")
	if !strings.HasPrefix(line, "#!") {
		return false
	}
	// Match the switch, not the letter: a path containing a T is not a
	// request for taint mode.
	for _, field := range strings.Fields(line) {
		if strings.HasPrefix(field, "-") && strings.ContainsRune(field, 'T') {
			return true
		}
	}
	return false
}
