// ABOUTME: A signature's bare `$`, `@` or `%` is a placeholder sigil, not the start of
// ABOUTME: the punctuation variable `$)` or `$,` that the same bytes spell elsewhere.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestSignaturePlaceholderSigil: toke.c reads a signature element with
// yyl_sigvar, whose own comment names the trap -- "the general yylex code
// would otherwise try to interpret whatever follows as a var; e.g. ($, ...)
// would be seen as the var '$,'". Measured on 5.42.0, each deparses
// unchanged:
//
//	$ perl -MO=Deparse -e 'use feature "signatures";
//	      sub t010 ($a, $) { $a } sub t011 ($, $a) { $a } sub t013 ($) { 1 }'
//
// perl.git t/op/signatures.t:129, 143, 157 and 171. Lexed as `$)`, the
// closing paren went into a variable and the signature swallowed the body.
func TestSignaturePlaceholderSigil(t *testing.T) {
	for _, sub := range []string{
		`sub t010 ($a, $) { $a || "z" }`,
		`sub t011 ($, $a) { $a || "z" }`,
		`sub t012 ($, $) { $a || "z" }`,
		`sub t013 ($) { $a || "z" }`,
		`sub t014 ($a, @) { $a }`,
		`sub t015 ($a, %) { $a }`,
		// A default is ordinary code: `$)` there is the effective gid.
		`sub t016 ($x = $)) { $x }`,
	} {
		src := "use feature 'signatures';\n" + sub + "\nmy $after = 1;\n"
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", sub, shape(root))
			continue
		}
		// The head up to the body's brace, exactly as written: a `$)` read
		// as a variable leaves the signature unclosed and canon moves the
		// paren after the body.
		head := sub[:strings.Index(sub, "{")+1]
		canon := parse.Canon(root, []byte(src))
		if !strings.Contains(canon, head) || !strings.Contains(canon, "my $after = 1;") {
			t.Errorf("%q: want %q and the next statement in canon, got %q", sub, head, canon)
		}
	}
	// Outside a signature `$)` is still the effective gid.
	root := parse.Parse([]byte("my $gid = $);\n"))
	if containsKind(root, parse.Unknown) {
		t.Errorf("`$)` outside a signature must stay a variable; got %s", shape(root))
	}
}
