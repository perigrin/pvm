// ABOUTME: Tests the sweep receipt: it records what was measured, and Verify rejects every field that disagrees with the tree.
// ABOUTME: The floor on baseline rows is the F7 fix -- a ratchet over an empty denominator passes on nothing.

package parseoracle_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

const pinPath = "testdata/corpus.pin"

// TestReceiptRecordsTheSweep: the receipt names what was measured -- how many
// files, against which baseline bytes, in which pinned world, with which
// verdicts -- and survives the round trip through the file the workflow's
// next step reads.
func TestReceiptRecordsTheSweep(t *testing.T) {
	report := fixtureReport(t)
	base := loadFixtureBaseline(t)
	if err := base.Check(report); err != nil {
		t.Fatalf("premise: the fixture must match its baseline: %v", err)
	}

	receipt, err := parseoracle.NewReceipt("TestRatchet", fixtureBaselinePath(), report)
	if err != nil {
		t.Fatalf("NewReceipt: %v", err)
	}

	data, err := os.ReadFile(fixtureBaselinePath())
	if err != nil {
		t.Fatalf("reading the fixture baseline: %v", err)
	}
	sum := sha256.Sum256(data)
	want := parseoracle.Receipt{
		Test:           "TestRatchet",
		FilesSwept:     len(report.Files),
		BaselineRows:   len(base.Rows),
		BaselineSHA256: hex.EncodeToString(sum[:]),
		Pin:            base.Pin,
		VerdictSHA256:  receipt.VerdictSHA256,
	}
	if receipt != want {
		t.Errorf("receipt is\n%+v\nwant\n%+v", receipt, want)
	}
	if receipt.FilesSwept != 5 {
		t.Errorf("the fixture sweep must report 5 files, got %d", receipt.FilesSwept)
	}
	if len(receipt.VerdictSHA256) != sha256.Size*2 {
		t.Errorf("verdict_sha256 is %q, want a hex sha256", receipt.VerdictSHA256)
	}

	path := filepath.Join(t.TempDir(), "receipt.json")
	if err := receipt.Write(path); err != nil {
		t.Fatalf("Write: %v", err)
	}
	back, err := parseoracle.ReadReceipt(path)
	if err != nil {
		t.Fatalf("ReadReceipt: %v", err)
	}
	if back != receipt {
		t.Errorf("receipt did not round-trip:\n%+v\nvs\n%+v", back, receipt)
	}

	// The keys are the contract the verifier and any reader of the job log
	// share, so they are asserted by name rather than trusted to a struct tag.
	raw, _ := os.ReadFile(path)
	for _, key := range []string{`"test"`, `"files_swept"`, `"baseline_rows"`, `"baseline_sha256"`, `"pin"`, `"verdict_sha256"`} {
		if !strings.Contains(string(raw), key) {
			t.Errorf("receipt JSON lacks %s:\n%s", key, raw)
		}
	}
}

// TestReceiptVerify is the verifier's contract: a receipt that agrees with
// the checked-out baseline and pin passes, and a receipt that disagrees on
// ANY field fails naming the field. A verifier that checked only the sha
// would pass a sweep that read the right file and measured nothing.
func TestReceiptVerify(t *testing.T) {
	baselinePath, report := syntheticCorpus(t, parseoracle.MinCorpusRows+10)
	good, err := parseoracle.NewReceipt("TestRatchetCorpus", baselinePath, report)
	if err != nil {
		t.Fatalf("NewReceipt: %v", err)
	}
	if err := good.Verify(baselinePath, pinPath, parseoracle.MinCorpusRows); err != nil {
		t.Fatalf("a receipt from the sweep it describes must verify: %v", err)
	}

	for _, tc := range []struct {
		name   string
		mutate func(r *parseoracle.Receipt)
		want   string
	}{
		{"another test wrote it", func(r *parseoracle.Receipt) { r.Test = "TestRatchet" }, "TestRatchetCorpus"},
		{"baseline bytes differ", func(r *parseoracle.Receipt) { r.BaselineSHA256 = strings.Repeat("0", 64) }, "baseline_sha256"},
		{"row count differs", func(r *parseoracle.Receipt) { r.BaselineRows-- }, "baseline_rows"},
		{"swept fewer than baselined", func(r *parseoracle.Receipt) { r.FilesSwept-- }, "files_swept"},
		{"pin differs", func(r *parseoracle.Receipt) { r.Pin.Revision = "cafef00d" }, "pin"},
		{"verdicts differ", func(r *parseoracle.Receipt) { r.VerdictSHA256 = strings.Repeat("0", 64) }, "verdict_sha256"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := good
			tc.mutate(&r)
			err := r.Verify(baselinePath, pinPath, parseoracle.MinCorpusRows)
			if err == nil {
				t.Fatalf("a receipt whose %s must not verify", tc.name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("the failure must name %s, got: %v", tc.want, err)
			}
		})
	}
}

// TestReceiptRefusesAnEmptyDenominator is F7. Check is a symmetric diff, so
// a baseline truncated to its header and a shim with no .t files agree
// perfectly: zero rows, zero files, nothing moved, PASS. The receipt's floor
// is where that stops being possible inside the gate's own job.
func TestReceiptRefusesAnEmptyDenominator(t *testing.T) {
	t.Run("zero files", func(t *testing.T) {
		baselinePath, report := syntheticCorpus(t, 0)
		r, err := parseoracle.NewReceipt("TestRatchetCorpus", baselinePath, report)
		if err != nil {
			t.Fatalf("NewReceipt: %v", err)
		}
		err = r.Verify(baselinePath, pinPath, 0)
		if err == nil || !strings.Contains(err.Error(), "files_swept") {
			t.Errorf("a sweep of zero files must fail naming files_swept, got: %v", err)
		}
	})

	t.Run("below the floor", func(t *testing.T) {
		baselinePath, report := syntheticCorpus(t, parseoracle.MinCorpusRows-1)
		r, err := parseoracle.NewReceipt("TestRatchetCorpus", baselinePath, report)
		if err != nil {
			t.Fatalf("NewReceipt: %v", err)
		}
		err = r.Verify(baselinePath, pinPath, parseoracle.MinCorpusRows)
		if err == nil || !strings.Contains(err.Error(), "baseline_rows") {
			t.Errorf("%d rows must fail the %d floor naming baseline_rows, got: %v",
				parseoracle.MinCorpusRows-1, parseoracle.MinCorpusRows, err)
		}
		// The same receipt passes a floor it clears, so the floor is the
		// only thing that failed above.
		if err := r.Verify(baselinePath, pinPath, 1); err != nil {
			t.Errorf("the same receipt must pass a floor of 1: %v", err)
		}
	})
}

// syntheticCorpus writes an n-row baseline pinned to the committed pin and
// returns a report that matches it exactly, so tests can exercise the
// receipt at corpus scale without a corpus.
func syntheticCorpus(t *testing.T, n int) (string, parseoracle.Report) {
	t.Helper()

	pin, err := parseoracle.ReadPin(pinPath)
	if err != nil {
		t.Fatalf("ReadPin: %v", err)
	}
	base := parseoracle.Baseline{Pin: pin}
	var report parseoracle.Report
	for i := 0; i < n; i++ {
		path := fmt.Sprintf("op/file%04d.t", i)
		status := parseoracle.BucketExact.String()
		if i%3 == 0 {
			status = parseoracle.BucketNoAnswer.String()
		}
		base.Rows = append(base.Rows, parseoracle.Row{
			Status: status, Category: parseoracle.CategoryNone, Path: path,
		})
		report.Files = append(report.Files, reportWithStatus(t, path, status).Files...)
	}

	path := filepath.Join(t.TempDir(), "baseline.txt")
	if err := parseoracle.WriteBaseline(path, base); err != nil {
		t.Fatalf("WriteBaseline: %v", err)
	}
	return path, report
}
