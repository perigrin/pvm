// ABOUTME: A braced variable name may follow its sigil after spaces: `$ {^XY}` is
// ABOUTME: `${^XY}`, and `$ { foo }` is `$foo`.

package lexer

import "testing"

// TestSpacedBracedName: scan_ident skips space after a sigil, before a brace
// as before a name. Measured on 5.42.0:
//
//	$ perl -e '$ {^XY} = 23; print $ {^XY}, " ", ${^XY}, "\n";
//	      ${ foo } = 4; print $ { foo }, "\n";'
//	23 23
//	4
//
// perl.git t/base/lex.t:129, `if ($ {^XY} != 23)`. Taken as the one-byte
// punctuation variable `$ `, the space swallowed the name.
func TestSpacedBracedName(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"my $v = $ {^XY};", "$ {^XY}"},
		{"my $v = $ { foo };", "$ { foo }"},
	} {
		toks := significant(Tokenize([]byte(tc.src)))
		if len(toks) < 4 || toks[3].Kind != Variable || tc.src[toks[3].Start:toks[3].End] != tc.want {
			t.Errorf("%q: want one Variable %q", tc.src, tc.want)
		}
	}
}
