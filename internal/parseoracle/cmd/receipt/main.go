// ABOUTME: Verifies the receipt TestRatchetCorpus wrote against the checked-out baseline and pin.
// ABOUTME: The fidelity workflow runs it from the repository root; a missing or disagreeing receipt exits non-zero.

// Usage: go run ./internal/parseoracle/cmd/receipt [RECEIPT]
//
// RECEIPT defaults to $PARSEORACLE_RECEIPT. The baseline and pin are the
// committed ones, deliberately not configurable: a verifier that could be
// pointed at another baseline would verify whatever it was pointed at.
package main

import (
	"fmt"
	"os"

	"tamarou.com/pvm/internal/parseoracle"
)

const (
	baseline = "internal/parseoracle/testdata/ratchet/baseline.txt"
	pin      = "internal/parseoracle/testdata/corpus.pin"
)

func main() {
	path := os.Getenv("PARSEORACLE_RECEIPT")
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	if path == "" {
		fail("no receipt path: pass one or set $PARSEORACLE_RECEIPT")
	}

	receipt, err := parseoracle.ReadReceipt(path)
	if err != nil {
		fail("%v -- the sweep did not run TestRatchetCorpus to completion", err)
	}
	if err := receipt.Verify(baseline, pin, parseoracle.MinCorpusRows); err != nil {
		fail("%v", err)
	}
	fmt.Printf("receipt: %s swept %d files against the %d-row baseline (sha256 %s, pin %s@%s)\n",
		receipt.Test, receipt.FilesSwept, receipt.BaselineRows, receipt.BaselineSHA256,
		receipt.Pin.Interpreter, receipt.Pin.Revision)
}

// fail prints in the form GitHub surfaces as an annotation, then exits 1.
func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "::error::receipt: "+format+"\n", args...)
	os.Exit(1)
}
