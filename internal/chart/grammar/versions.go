// ABOUTME: The version tokens toke.c forces into BAREWORDs after package, use and no, as classes over pvm's tokens.
// ABOUTME: Generated in init beside perly.go's literal because one version token has a class per spelling and context.

package grammar

import "tamarou.com/pvm/internal/lexer"

// init declares the BAREWORD a version is. toke.c forces it from the text after `package NAME` (8858,
// force_strict_version) and after `use`/`no` (5435, tokenize_use: force_version, then an empty BAREWORD
// if `;` or `}` follows). pvm emits it as a Number, or a Quote opening with `v` or a digit for a v-string
// (`v5.36`, `5.10.1`).
//
// The first token of `use VERSION` is a version only where perl accepts what follows: `;`, `}` or the end
// of the file, else `use 5.036 1;` would read as a module list.
// A version after `use MODULE` is a version wherever a number or v-string stands there. After
// `package NAME` only a Number is, the one spelling is_STRICT_VERSION admits that pvm lexes there: its
// v-string, `package Foo v1.2.3`, lexes as Word(v1) Operator(.) Number(2.3), a lexer defect.
//
// shortcut: three limits, each upgraded when TokenClass can say it. `use VERSION MODULE` (`use 5.036 strict`)
// is refused: toke.c lexes that module as an ordinary term, so taking it would also take `use 5.036 foo bar;`,
// which perl refuses as an indirect-object call. The version's text is not checked, so `use 0x10;` and
// `package Foo 1_0;` are taken (TokenClass tests no number text). And any word after a package name is
// taken as its version (`package Foo Bar;`, which toke.c's force_strict_version refuses): BAREWORD's generic
// class cannot yet exclude a word two tokens after `package` (NotPrev2).
func init() {
	versions := []TokenClass{{Kind: ptr(lexer.Number)}}
	for _, p := range "v0123456789" {
		versions = append(versions, TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte(p))})
	}
	notVersion := []Neighbour{{Kind: lexer.Number}, {Kind: lexer.Quote}}
	endings := []*Neighbour{{Kind: lexer.Semicolon}, {Kind: lexer.CloseBracket, Text: "}", OpensBlock: true}, &Boundary}
	add := func(c TokenClass) { perlyClasses = append(perlyClasses, ClassDecl{"BAREWORD", c}) }

	pkg := versions[0]
	pkg.Prev2, pkg.NotPrev1 = &Neighbour{Kind: lexer.Word, Text: "package"}, notVersion
	add(pkg)
	for _, kw := range []string{"use", "no"} {
		for _, v := range versions {
			after := v
			after.Prev2, after.NotPrev1 = &Neighbour{Kind: lexer.Word, Text: kw}, notVersion
			add(after)
			for _, next := range endings {
				lone := v
				lone.Prev1, lone.Next = &Neighbour{Kind: lexer.Word, Text: kw}, next
				add(lone)
			}
		}
	}
}
