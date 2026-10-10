// ABOUTME: The parser's own parse of print, exec and their kin, held to golden trees.
// ABOUTME: Declaring them in CORE.pmt with an invocant colon must leave how they parse unchanged.
package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestInvocantBuiltinParseUnchanged: RFC 0001 "Builtins that keep their
// own parse". CORE.pmt types print, printf, say, exec and system, and the
// parser still reads their leading slot as it always has: a handle or a
// block with no comma after it, or nothing at all. The trees are golden
// values held here, as the parser gave them before the declarations.
func TestInvocantBuiltinParseUnchanged(t *testing.T) {
	for src, want := range map[string]string{
		`print STDERR "a", "b";`:  "source_file\n  statement\n    call print\n      term STDERR\n      binary ,\n        term \"a\"\n        term \"b\"\n",
		`print {$fh} "a";`:        "source_file\n  statement\n    call print\n      block\n        statement\n          term $fh\n      term \"a\"\n",
		`print;`:                  "source_file\n  statement\n    call print\n",
		`exec {"/bin/sh"} @args;`: "source_file\n  statement\n    call exec\n      block\n        statement\n          term \"/bin/sh\"\n      term @args\n",
	} {
		if got := dumpTree(parse.Parse([]byte(src)), 0); got != want {
			t.Errorf("%s:\n got\n%s\nwant\n%s", src, got, want)
		}
	}
}
