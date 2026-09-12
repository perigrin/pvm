// ABOUTME: Freezes per-file verdicts as a committed baseline and fails when a file moves in either direction.
// ABOUTME: Both directions fail deliberately, so the baseline can never silently diverge from reality.

package parseoracle

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Category is the construct that explains a file's first parse error.
//
// A raw count says "203 files are no-answer". A taxonomy says "48 of 203 are
// QuoteLike", and only the second is a work plan. The set is fixed rather than
// free-form so that a baseline diff shows work moving between named buckets
// instead of between ad-hoc strings.
type Category string

const (
	// CategoryModernFeature is P1: `class`/`field`/`method`, signatures,
	// `try`/`catch` — the post-5.36 surface, and the highest priority
	// because it is the Perl people are writing now.
	CategoryModernFeature Category = "ModernFeature"

	// CategoryQuoteLike is P2: heredocs, `qw`/`q`/`qq`/`tr`/`y` and their
	// arbitrary delimiters. The largest P2 population, because a
	// quote-like operator's body is not lexable without knowing the
	// operator.
	CategoryQuoteLike Category = "QuoteLike"

	// CategoryRegex is P2: match, substitution, and the modifier soup that
	// changes how the pattern body itself is read.
	CategoryRegex Category = "Regex"

	// CategoryControlFlow is P2: statement modifiers, loop labels, `do`
	// blocks in expression position.
	CategoryControlFlow Category = "ControlFlow"

	// CategoryDereference is P2: sigil chains, `->@*` postfix deref, and
	// block-form `${...}` dereferences.
	CategoryDereference Category = "Dereference"

	// CategorySubroutine is P2: prototypes, attributes, signatures on the
	// declaration side, and `&`-sigil calls.
	CategorySubroutine Category = "Subroutine"

	// CategoryGeneral is P3: everything the rules above do not claim. It is
	// the honest answer for a file whose first error is not attributable,
	// and a baseline that is ALL General is a taxonomy that was never
	// populated — TestBaselineCategoryFollowsConstruct exists to catch that.
	CategoryGeneral Category = "General"

	// CategoryNone is the category of a file with no first error at all: it
	// parsed, so there is no construct to name. Distinct from General,
	// which means "it failed and we cannot say why".
	CategoryNone Category = "-"
)

// Categories is the fixed taxonomy, in priority order.
var Categories = []Category{
	CategoryModernFeature,
	CategoryQuoteLike,
	CategoryRegex,
	CategoryControlFlow,
	CategoryDereference,
	CategorySubroutine,
	CategoryGeneral,
	CategoryNone,
}

// ValidCategory reports whether c is in the fixed taxonomy.
func ValidCategory(c Category) bool {
	for _, known := range Categories {
		if c == known {
			return true
		}
	}
	return false
}

// Row is one file's frozen verdict.
//
// The columns are `status metric category path`, in that order, with path last
// because it is the only field that can contain no spaces but is unbounded in
// length — putting it last makes the file parseable by splitting on the first
// three fields.
type Row struct {
	// Status is the frozen verdict: a bucket name, or "excluded" for an
	// environmental failure, or "error" for a file the runner could not
	// measure. All three are recorded, because a file silently leaving the
	// denominator is exactly the drift a baseline exists to catch.
	Status string
	// Metric is the parse fact the verdict turned on: the number of
	// references perl took. It is recorded but NOT ratcheted — a metric
	// that moved while the status held is information, not a regression,
	// and failing on it would make the baseline unmaintainable.
	Metric int
	// Category is the taxonomy entry for this file's first error.
	Category Category
	// Path is the file, as the runner was given it.
	Path string
}

// statusExcluded and statusError are the two non-bucket statuses.
const (
	statusExcluded = "excluded"
	statusError    = "error"
)

// rank orders statuses from best to worst, so a move can be called an
// improvement or a regression rather than merely a change.
//
// excluded and error sit outside the ordering: a file that left the
// denominator has not got better or worse, it has stopped being measured, and
// calling that an improvement would let a broken shim read as progress. Any
// move into or out of them is reported as a regression, because it means the
// measurement changed shape and a human should look.
func rank(status string) (int, bool) {
	switch status {
	case BucketExact.String():
		return 0, true
	case BucketWider.String():
		return 1, true
	case BucketNoAnswer.String():
		return 2, true
	case BucketWrong.String():
		return 3, true
	}
	return 0, false
}

// Pin is recorded in the baseline header; see corpus.go for its semantics.

// Baseline is a frozen corpus measurement.
type Baseline struct {
	// Pin is the world the rows were measured in. Comparing verdicts across
	// a moved pin measures the skew, not the parser.
	Pin Pin
	// Rows is one row per file, sorted by path so a baseline diff shows
	// only what actually moved.
	Rows []Row
}

const doNotEdit = "DO NOT EDIT BY HAND"

// Render writes the baseline in its committed form.
func (b Baseline) Render() string {
	var out strings.Builder
	out.WriteString("# ABOUTME: Frozen per-file parse verdicts; a file that moves in either direction fails the ratchet.\n")
	out.WriteString("# ABOUTME: Regenerated by `go test ./internal/parseoracle/ -run TestRatchet -parseoracle.update`.\n")
	out.WriteString("#\n")
	out.WriteString("# " + doNotEdit + ". An improvement fails the build too, on purpose:\n")
	out.WriteString("# the baseline must never silently diverge from reality in either\n")
	out.WriteString("# direction, so re-running with -parseoracle.update in the same commit\n")
	out.WriteString("# is how a gain is recorded.\n")
	out.WriteString("#\n")
	out.WriteString("# Columns: status metric category path\n")
	fmt.Fprintf(&out, "interpreter = %s\n", b.Pin.Interpreter)
	fmt.Fprintf(&out, "revision    = %s\n", b.Pin.Revision)

	rows := append([]Row(nil), b.Rows...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].Path < rows[j].Path })
	for _, r := range rows {
		fmt.Fprintf(&out, "%-9s %5d %-14s %s\n", r.Status, r.Metric, r.Category, r.Path)
	}
	return out.String()
}

// NewBaseline freezes a report against a pin.
//
// dir is the working directory the report was measured from — RunOptions.Dir.
// It is needed because the runner records corpus paths relative to the shim's
// t/ while the taxonomy has to re-read each file to categorise it. Pass "" for
// a report whose paths resolve from the process working directory.
//
// Categorisation runs in parallel because it re-parses every no-answer file,
// and there are 203 of those in the corpus. Measured serially that pass costs
// more than the perl sweep it follows, which would make a re-baseline
// something nobody runs — and an unrunnable re-baseline is how the ratchet
// ends up lagging reality, which is the exact failure this whole issue exists
// to prevent.
func NewBaseline(pin Pin, report Report, dir string) Baseline {
	b := Baseline{Pin: pin, Rows: make([]Row, len(report.Files))}

	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup
	for i, f := range report.Files {
		wg.Add(1)
		go func(i int, f Result) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			b.Rows[i] = Row{
				Status:   status(f),
				Metric:   f.Facts.Srefgen,
				Category: Categorise(f, dir),
				Path:     f.Path,
			}
		}(i, f)
	}
	wg.Wait()

	sort.Slice(b.Rows, func(i, j int) bool { return b.Rows[i].Path < b.Rows[j].Path })
	return b
}

// status names a result's frozen verdict.
func status(r Result) string {
	switch {
	case r.Err != "":
		return statusError
	case r.Excluded:
		return statusExcluded
	default:
		return r.Verdict.Bucket.String()
	}
}

// ParseBaseline reads the committed form back.
func ParseBaseline(data []byte) (Baseline, error) {
	var b Baseline
	for n, line := range strings.Split(string(data), "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Header lines are `key = value`; rows never contain `=` before
		// their path, and a path that does is still unambiguous because
		// the key would have to be one of the two names below.
		if key, value, ok := strings.Cut(line, "="); ok {
			switch strings.TrimSpace(key) {
			case "interpreter":
				b.Pin.Interpreter = strings.TrimSpace(value)
				continue
			case "revision":
				b.Pin.Revision = strings.TrimSpace(value)
				continue
			}
		}

		fields := strings.Fields(line)
		if len(fields) != 4 {
			return Baseline{}, fmt.Errorf("baseline line %d: want `status metric category path`, got %q", n+1, line)
		}
		metric, err := strconv.Atoi(fields[1])
		if err != nil {
			return Baseline{}, fmt.Errorf("baseline line %d: metric %q is not a number", n+1, fields[1])
		}
		category := Category(fields[2])
		if !ValidCategory(category) {
			return Baseline{}, fmt.Errorf("baseline line %d: category %q is not in the fixed taxonomy %v",
				n+1, category, Categories)
		}
		b.Rows = append(b.Rows, Row{Status: fields[0], Metric: metric, Category: category, Path: fields[3]})
	}

	if b.Pin.Interpreter == "" || b.Pin.Revision == "" {
		return Baseline{}, fmt.Errorf("baseline header must record both interpreter and revision, got %+v", b.Pin)
	}
	sort.Slice(b.Rows, func(i, j int) bool { return b.Rows[i].Path < b.Rows[j].Path })
	return b, nil
}

// LoadBaseline reads a baseline file.
func LoadBaseline(path string) (Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Baseline{}, fmt.Errorf("reading baseline: %w", err)
	}
	b, err := ParseBaseline(data)
	if err != nil {
		return Baseline{}, fmt.Errorf("%s: %w", path, err)
	}
	return b, nil
}

// WriteBaseline is the ONLY path that rewrites a baseline, and it is reached
// only from -parseoracle.update. A check that rewrote what it was checking
// against could never fail, which is the shape of every ratchet that turns
// into decoration.
func WriteBaseline(path string, b Baseline) error {
	return os.WriteFile(path, []byte(b.Render()), 0o644)
}

// Move is one file whose verdict changed.
type Move struct {
	Path     string
	From, To string
	Category Category
	Detail   string
}

func (m Move) String() string {
	s := fmt.Sprintf("%s: %s -> %s [%s]", m.Path, m.From, m.To, m.Category)
	if m.Detail != "" {
		s += " (" + m.Detail + ")"
	}
	return s
}

// RatchetError is a baseline that no longer describes reality.
//
// Regressions and Improvements are separate fields rather than one list of
// changes, because the two require different actions and a message that
// conflates them trains readers to ignore both.
type RatchetError struct {
	Regressions  []Move
	Improvements []Move
	// Missing are baseline rows with no corresponding file in the report,
	// and Unexpected are report files with no row. Both mean the corpus
	// itself moved, which is a re-baseline rather than a parser change.
	Missing    []string
	Unexpected []string
}

func (e *RatchetError) Error() string {
	var out strings.Builder

	if len(e.Regressions) > 0 {
		fmt.Fprintf(&out, "%d file(s) regressed:\n", len(e.Regressions))
		for _, m := range e.Regressions {
			fmt.Fprintf(&out, "  %s\n", m)
		}
	}
	if len(e.Improvements) > 0 {
		fmt.Fprintf(&out, "%d file(s) improved and the baseline does not record it:\n",
			len(e.Improvements))
		for _, m := range e.Improvements {
			fmt.Fprintf(&out, "  %s\n", m)
		}
		out.WriteString("  re-run with -parseoracle.update and commit the new baseline " +
			"in the same commit, so it never lags reality\n")
	}
	if len(e.Missing) > 0 {
		fmt.Fprintf(&out, "%d baseline row(s) have no file in this run: %s\n",
			len(e.Missing), strings.Join(e.Missing, ", "))
	}
	if len(e.Unexpected) > 0 {
		fmt.Fprintf(&out, "%d file(s) in this run have no baseline row: %s\n",
			len(e.Unexpected), strings.Join(e.Unexpected, ", "))
	}
	return strings.TrimRight(out.String(), "\n")
}

// Check compares a fresh report against the frozen baseline.
//
// Both directions fail. Plan §4.1: "Any file that regresses fails the build.
// Any file that improves requires a baseline update in the same commit." The
// property protected is not "the parser never gets worse" but "the baseline
// never diverges from reality", and a baseline that quietly absorbs
// improvements decays into a file nobody trusts — which is how perl-lsp
// arrived at a `ci/parse_errors_baseline.txt` containing the single line `0`.
func (b Baseline) Check(report Report) error {
	frozen := make(map[string]Row, len(b.Rows))
	for _, r := range b.Rows {
		frozen[r.Path] = r
	}

	var diff RatchetError
	seen := make(map[string]struct{}, len(report.Files))
	for _, f := range report.Files {
		seen[f.Path] = struct{}{}
		was, ok := frozen[f.Path]
		if !ok {
			diff.Unexpected = append(diff.Unexpected, f.Path)
			continue
		}
		now := status(f)
		if now == was.Status {
			continue
		}

		move := Move{Path: f.Path, From: was.Status, To: now,
			Category: was.Category, Detail: f.Verdict.Detail}

		// A file entering or leaving the measured population has not got
		// better or worse; it has changed shape. That always wants a
		// human, so it is reported with the regressions.
		wasRank, wasMeasured := rank(was.Status)
		nowRank, nowMeasured := rank(now)
		switch {
		case !wasMeasured || !nowMeasured:
			diff.Regressions = append(diff.Regressions, move)
		case nowRank > wasRank:
			diff.Regressions = append(diff.Regressions, move)
		default:
			diff.Improvements = append(diff.Improvements, move)
		}
	}

	for _, r := range b.Rows {
		if _, ok := seen[r.Path]; !ok {
			diff.Missing = append(diff.Missing, r.Path)
		}
	}

	if len(diff.Regressions) == 0 && len(diff.Improvements) == 0 &&
		len(diff.Missing) == 0 && len(diff.Unexpected) == 0 {
		return nil
	}
	return &diff
}

// SkewError reports that the baseline was measured in a different world.
//
// It is deliberately NOT a RatchetError: skew is not a defect. When the
// interpreter or the corpus moves, every verdict below the header was measured
// against something else, so reporting "203 files regressed" would be a
// confident wrong explanation of a version bump.
type SkewError struct{ Cause error }

func (e *SkewError) Error() string {
	return fmt.Sprintf("re-baseline needed: %v", e.Cause)
}

func (e *SkewError) Unwrap() error { return e.Cause }

// CheckPin compares the baseline's recorded world against the one present.
func (b Baseline) CheckPin(actual Pin) error {
	if err := b.Pin.Check(actual.Interpreter, actual.Revision); err != nil {
		return &SkewError{Cause: err}
	}
	return nil
}
