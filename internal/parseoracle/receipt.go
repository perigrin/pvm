// ABOUTME: The sweep's receipt: proof, checked by a later step, that TestRatchetCorpus measured the committed baseline.
// ABOUTME: A sweep that was skipped, filtered, listed, or never executed writes no receipt, so its absence is the failure.

package parseoracle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// MinCorpusRows is the floor under the baseline the gate accepts a receipt
// for. The corpus is 620 files; half of that survives a legitimate pin move
// and rejects a baseline truncated to its header, which is the perl-lsp
// shape -- a ratchet whose baseline says `0` and can never fail.
const MinCorpusRows = 310

// Receipt is what TestRatchetCorpus writes, and only after the ratchet has
// passed. Every way of making `go test` exit 0 without running the sweep --
// -run naming another test, -skip, -list, -count=0, -exec /bin/true, a
// missing -parseoracle.corpus, a skipped corpus -- ends with no receipt, and
// the workflow's next step fails on its absence. The fields are what a
// verifier can recompute from the checked-out tree, so a receipt cannot be
// satisfied by a sweep of some other baseline.
type Receipt struct {
	// Test is the test that wrote the receipt; the verifier expects the
	// corpus ratchet by name, so the fixture ratchet cannot stand in for it.
	Test string `json:"test"`
	// FilesSwept is the size of the report the ratchet checked.
	FilesSwept int `json:"files_swept"`
	// BaselineRows is the size of the baseline it checked against.
	BaselineRows int `json:"baseline_rows"`
	// BaselineSHA256 is the digest of the baseline file's bytes, so the
	// verifier can prove it was the committed one and not a substitute.
	BaselineSHA256 string `json:"baseline_sha256"`
	// Pin is the world the baseline records.
	Pin Pin `json:"pin"`
	// VerdictSHA256 digests the sorted (path, status) pairs the sweep
	// produced. Because Check passed, those are also the baseline's rows,
	// and the verifier recomputes the digest from the committed file.
	VerdictSHA256 string `json:"verdict_sha256"`
}

// NewReceipt describes a report that has just passed Check against the
// baseline at baselinePath. It does not run Check itself: the receipt is
// evidence of a measurement, and the caller decides when one has happened.
func NewReceipt(test, baselinePath string, report Report) (Receipt, error) {
	data, err := os.ReadFile(baselinePath)
	if err != nil {
		return Receipt{}, fmt.Errorf("reading baseline for the receipt: %w", err)
	}
	base, err := ParseBaseline(data)
	if err != nil {
		return Receipt{}, fmt.Errorf("%s: %w", baselinePath, err)
	}

	verdicts := make(map[string]string, len(report.Files))
	for _, f := range report.Files {
		verdicts[f.Path] = status(f)
	}
	return Receipt{
		Test:           test,
		FilesSwept:     len(report.Files),
		BaselineRows:   len(base.Rows),
		BaselineSHA256: sha256Hex(data),
		Pin:            base.Pin,
		VerdictSHA256:  verdictDigest(verdicts),
	}, nil
}

// Write records the receipt as JSON at path.
func (r Receipt) Write(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// ReadReceipt reads a receipt back. A missing file is an error, because a
// missing receipt is the whole signal.
func ReadReceipt(path string) (Receipt, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Receipt{}, fmt.Errorf("no receipt: %w", err)
	}
	var r Receipt
	if err := json.Unmarshal(data, &r); err != nil {
		return Receipt{}, fmt.Errorf("receipt %s: %w", path, err)
	}
	return r, nil
}

// Verify checks the receipt against the checked-out tree: the baseline at
// baselinePath byte for byte, the pin at pinPath, the row count, and the
// verdicts themselves. minRows is the floor on the baseline; the gate passes
// MinCorpusRows, and a fixture-scale test passes something smaller.
//
// Every failure names the field that disagreed, because "receipt invalid" is
// not an answer anyone can act on.
func (r Receipt) Verify(baselinePath, pinPath string, minRows int) error {
	if r.Test != "TestRatchetCorpus" {
		return fmt.Errorf("receipt was written by %q, want TestRatchetCorpus", r.Test)
	}
	if r.FilesSwept <= 0 {
		return fmt.Errorf("receipt reports files_swept=%d: the sweep measured nothing", r.FilesSwept)
	}
	if r.BaselineRows < minRows {
		return fmt.Errorf("receipt reports baseline_rows=%d, below the floor of %d: "+
			"a baseline this short has been truncated, and a ratchet over it proves nothing",
			r.BaselineRows, minRows)
	}
	if r.FilesSwept != r.BaselineRows {
		return fmt.Errorf("receipt reports files_swept=%d but baseline_rows=%d: "+
			"the ratchet cannot have passed over a report that differs in size from its baseline",
			r.FilesSwept, r.BaselineRows)
	}

	data, err := os.ReadFile(baselinePath)
	if err != nil {
		return fmt.Errorf("reading the checked-out baseline: %w", err)
	}
	if got := sha256Hex(data); got != r.BaselineSHA256 {
		return fmt.Errorf("receipt records baseline_sha256=%s but %s hashes to %s: "+
			"the sweep checked some other baseline", r.BaselineSHA256, baselinePath, got)
	}
	base, err := ParseBaseline(data)
	if err != nil {
		return fmt.Errorf("%s: %w", baselinePath, err)
	}
	if len(base.Rows) != r.BaselineRows {
		return fmt.Errorf("receipt records baseline_rows=%d but %s has %d rows",
			r.BaselineRows, baselinePath, len(base.Rows))
	}

	pin, err := ReadPin(pinPath)
	if err != nil {
		return err
	}
	if r.Pin != pin {
		return fmt.Errorf("receipt records pin %+v but %s records %+v: "+
			"the verdicts were measured in a different world", r.Pin, pinPath, pin)
	}

	verdicts := make(map[string]string, len(base.Rows))
	for _, row := range base.Rows {
		verdicts[row.Path] = row.Status
	}
	if got := verdictDigest(verdicts); got != r.VerdictSHA256 {
		return fmt.Errorf("receipt records verdict_sha256=%s but the checked-out baseline's "+
			"rows digest to %s: the sweep's verdicts are not this baseline's", r.VerdictSHA256, got)
	}
	return nil
}

// verdictDigest hashes (path, status) pairs in path order, so the same
// verdicts always digest the same way whichever side computed them.
func verdictDigest(verdicts map[string]string) string {
	paths := make([]string, 0, len(verdicts))
	for p := range verdicts {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var b strings.Builder
	for _, p := range paths {
		b.WriteString(p)
		b.WriteByte('\t')
		b.WriteString(verdicts[p])
		b.WriteByte('\n')
	}
	return sha256Hex([]byte(b.String()))
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
