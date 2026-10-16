// ABOUTME: Nesting deeper than the parser's cap refuses as nesting_too_deep instead of overflowing the Go stack.
// ABOUTME: A Go stack overflow is fatal and no recover catches it, so the LSP would die on one deep file.

package parse_test

import (
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/parse"
)

// TestDeepNestingIsBounded: perl accepts all of these -- measured on 5.42.0,
// `perl -c` says "syntax OK" to 400000 nested parens, 100000 nested bare
// blocks, and a chain of 100000 right-associative `**` -- but a recursive
// parser following them overflows its stack. Past the cap the parser
// refuses, so each is an Unknown rather than a fatal error.
//
// The control is real code's nesting, far below the cap, which parses clean.
func TestDeepNestingIsBounded(t *testing.T) {
	deep := map[string]string{
		"parens":   "my $x = " + strings.Repeat("(", 400000) + "1" + strings.Repeat(")", 400000) + ";\n",
		"brackets": "my $x = " + strings.Repeat("[", 100000) + strings.Repeat("]", 100000) + ";\n",
		"blocks":   strings.Repeat("{", 100000) + strings.Repeat("}", 100000) + "\n",
		"stmts":    strings.Repeat("{;", 100000) + strings.Repeat("}", 100000) + "\n",
		"power":    "my $x = 1" + strings.Repeat(" ** 1", 100000) + ";\n",
		"not":      "my $x = " + strings.Repeat("!", 100000) + "1;\n",
	}
	for name, src := range deep {
		start := time.Now()
		found := collectUnknown(parse.Parse([]byte(src)))
		if len(found) == 0 || found[0].Refusal != parse.NestingTooDeep {
			t.Errorf("%s: want a nesting_too_deep refusal, got %v", name, refusalCodes(found))
		}
		if d := time.Since(start); d > 60*time.Second {
			t.Errorf("%s: took %v", name, d)
		}
	}

	for _, src := range []string{
		"my $x = " + strings.Repeat("(", 100) + "1" + strings.Repeat(")", 100) + ";\n",
		strings.Repeat("{", 100) + strings.Repeat("}", 100) + "\n",
		"my $x = 1" + strings.Repeat(" ** 1", 100) + ";\n",
	} {
		if u := firstUnknown(parse.Parse([]byte(src))); u != nil {
			t.Errorf("control %.20q...: refused as %s", src, u.Refusal)
		}
	}
}
