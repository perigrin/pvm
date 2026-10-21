// ABOUTME: Tests oracleStderr, which trims a failed oracle run's stderr to
// ABOUTME: perl's own messages and the last call the oracle started.

package infer_test

import "testing"

func TestOracleStderrKeepsTheLastCall(t *testing.T) {
	in := "oracle: calling fileno:scalar:sub:0\r\n" +
		"oracle: calling ioctl:scalar:sub:2\r\n" +
		"readline() on closed filehandle Handle at x line 130.\r\n"
	want := "readline() on closed filehandle Handle at x line 130.\n" +
		"last call started: ioctl:scalar:sub:2"
	if got := oracleStderr(in); got != want {
		t.Errorf("oracleStderr = %q, want %q", got, want)
	}
	if got := oracleStderr("died here\n"); got != "died here" {
		t.Errorf("with no trace, oracleStderr = %q, want %q", got, "died here")
	}
}
