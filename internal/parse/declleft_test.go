// ABOUTME: A declaration with no initialiser is the left operand of a following infix
// ABOUTME: operator: `my $s =~ /x/` is `(my $s) =~ /x/`, not a declaration and a stray `=~`.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestDeclarationAsLeftOperand holds perl's reading of a bare declaration
// followed by an operator. Measured on 5.42.0, each deparses unchanged:
//
//	my $s =~ /x/;
//	my $t . 'a';
//	our @a == 1;
//
// parseVarDeclNoSemi took the target and an assignment-level initialiser, and
// handed back only for the word operators below the comma, so every other
// infix operator was left unconsumed and the statement refused as
// trailing_tokens -- `my $s =~ /x/;` canon'd as `my $s;=~ /x/;`.
func TestDeclarationAsLeftOperand(t *testing.T) {
	for _, src := range []string{
		`my $s =~ /x/;`,
		`my $t . "a";`,
		`our @a == 1;`,
	} {
		n := parse.Parse([]byte(src))
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q) has %d Unknown, want 0", src, got)
		}
		got := strings.TrimSpace(parse.Canon(n, []byte(src)))
		if got != src {
			t.Errorf("Canon(%q) = %s, want it unchanged", src, got)
		}
		again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
		if again != got {
			t.Errorf("canon is not a fixpoint:\n  once  %s\n  twice %s", got, again)
		}
	}
}
