// ABOUTME: Parse time grows linearly with statement nesting: a statement's inner heredoc bodies are not found by
// ABOUTME: rescanning every token it spans, which made each enclosing level pay for everything inside it.

package parse_test

import (
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/parse"
)

// TestNestedBlockParseIsNotQuadratic: each statement looks for the heredoc
// bodies it passed over. Rescanning its whole token range made a statement
// at every nesting level pay for all the tokens inside it, so a deep nest
// around a large body cost depth times the body. Measured before the fix
// with psc parse: 40000 nested blocks took 32 s and 80000 took 79 s.
//
// Nesting past the parser's depth cap is refused, so the nest here stays
// below it, around a body of empty statements, cheap to parse and many
// tokens long, so the rescan dominated: 900 levels took seven times as long
// as 10. Two depths over the same body: linear work takes about the same
// time for both.
func TestNestedBlockParseIsNotQuadratic(t *testing.T) {
	body := strings.Repeat(";\n", 300000) + "$x = 1;\n"
	timed := func(depth int) time.Duration {
		src := strings.Repeat("{\n", depth) + body + strings.Repeat("}\n", depth)
		start := time.Now()
		root := parse.Parse([]byte(src))
		d := time.Since(start)
		if u := firstUnknown(root); u != nil {
			t.Fatalf("depth %d: refused as %s", depth, u.Refusal)
		}
		return d
	}
	shallow, deep := timed(10), timed(900)
	if deep > 3*shallow+time.Second {
		t.Errorf("900 levels took %v, 10 levels %v: the nesting multiplies the work", deep, shallow)
	}
}
