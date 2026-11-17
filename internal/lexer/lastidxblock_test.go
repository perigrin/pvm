// ABOUTME: `$#{ EXPR }` is toke.c's DOLSHARP then a block: the lexer emits `$#` as a DerefSigil
// ABOUTME: and lexes the block's interior in the file's state, as it does for `${ EXPR }`.

package lexer

import (
	"strings"
	"testing"
)

// TestLastIndexBlockLexesItsInterior: toke.c's yyl_dollar emits PREREF(DOLSHARP)
// when scan_ident finds no name after `$#`, and the `{` is then lexed as a block
// (pinned perl 94e50866, -DT on `$#{ ["}"] };`: DOLSHARP, PERLY_BRACE_OPEN,
// PERLY_BRACKET_OPEN, THING, ...). The lexer counted braces byte by byte inside one
// Variable, so a quoted `}` ended it early and a heredoc lost its body.
func TestLastIndexBlockLexesItsInterior(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		// perl -e 'print $#{ ["}"] }' prints 0.
		{`$#{ ["}"] };`, `DerefSigil:$# Operator:{ Operator:[ Quote:"}" CloseBracket:] CloseBracket:} Semicolon:;`},
		{"$#{ [<<E] };\nb\nE\n", "DerefSigil:$# Operator:{ Operator:[ HeredocOpen:<<E CloseBracket:] CloseBracket:} Semicolon:; HeredocBody:b\nE\n"},
		// perl reads N's / as division after `use constant N`: the interior keeps the file's state.
		{`use constant N => 2; $#{ [ N /2/ 1 ] }`, `Word:use Word:constant Word:N Operator:=> Number:2 Semicolon:; DerefSigil:$# Operator:{ Operator:[ Word:N Operator:/ Number:2 Operator:/ Number:1 CloseBracket:] CloseBracket:}`},
		{`$#{$r}`, `DerefSigil:$# Operator:{ Variable:$r CloseBracket:}`},
		// No closing brace: `${` lexes the same way, and perl -c says "Missing right curly".
		{`$#{`, `DerefSigil:$# Operator:{`},
		// A braced NAME is still one token (toke.c's scan_ident reads it as the name),
		// and so are `$#name` and `$#$r`.
		{`$#{a}`, `Variable:$#{a}`},
		{`$#{ a }`, `Variable:$#{ a }`},
		{`$#a`, `Variable:$#a`},
		{`$#$r`, `Variable:$#$r`},
		// A package-qualified name is a name too: -DT shows DOLSHARP then the
		// identifier `@x::y` for each of these. A leading `::` opens a block.
		{`$#{x'y}`, `Variable:$#{x'y}`},
		{`$#{x::y}`, `Variable:$#{x::y}`},
		{`$#{ x::y }`, `Variable:$#{ x::y }`},
		{`$#{x::}`, `Variable:$#{x::}`},
		{`$#{::x}`, `DerefSigil:$# Operator:{ Word:::x CloseBracket:}`},
		// The other sigils read the braced name the same way (-DT: one identifier).
		{`${x'y}`, `Variable:${x'y}`},
		{`@{x'y}`, `Variable:@{x'y}`},
		{`${x::y}`, `Variable:${x::y}`},
		{`${::x}`, `DerefSigil:$ Operator:{ Word:::x CloseBracket:}`},
	} {
		var got []string
		for _, tk := range Tokenize([]byte(c.src)) {
			if tk.Kind == Whitespace {
				continue
			}
			got = append(got, tk.Kind.String()+":"+c.src[tk.Start:tk.End])
		}
		if g := strings.Join(got, " "); g != c.want {
			t.Errorf("%q:\n got %s\nwant %s", c.src, g, c.want)
		}
	}
}
