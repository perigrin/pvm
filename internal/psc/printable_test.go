// ABOUTME: Tests for printable, which escapes what psc prints of an untrusted name so no control reaches the terminal.
// ABOUTME: It covers the bytes no filesystem fixture can hold on every OS, such as one that is not UTF-8.

package psc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPrintableEscapes checks that a control character and a byte that is
// not UTF-8 print as \xNN, and that printable UTF-8, the control, prints
// as it is. A lone 0x9b is CSI to a terminal reading 8-bit controls.
func TestPrintableEscapes(t *testing.T) {
	for in, want := range map[string]string{
		"p\x1bq/y\x9b.pl":    `p\x1bq/y\x9b.pl`,
		"x\x1b]0;PWNED\a.pl": `x\x1b]0;PWNED\x07.pl`,
		"\u009b.pl":          `\x9b.pl`,
		"plain/café.pl":      "plain/café.pl",
	} {
		assert.Equal(t, want, printable(in), "%q", in)
	}
}
