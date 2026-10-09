// ABOUTME: An opener still open at end of file is refused, on the line perl reports.
// ABOUTME: Every opener kind is measured against perl 5.42, and each closed form parses clean.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestUnclosedOpenerAtEOFRefused: a `(`, `[` or `{` that never closes is
// refused, and the same source closed is not. Measured on 5.42.0 with
// `perl -c`, every refused case below is a compile error, reported on the
// line of the last token perl read:
//
//	0(                  syntax error at m.pl line 1, near "0("
//	for(0               syntax error at m.pl line 1, at EOF
//	{                   Missing right curly or square bracket at m.pl line 1
//	sub f {             Missing right curly or square bracket at m.pl line 1
//	my $x = (1          syntax error at m.pl line 1, at EOF
//	my $h = {           Missing right curly or square bracket at m.pl line 1
//	$h{  $a[  $x->{     syntax error at m.pl line 1, at EOF
//	for(  if (          syntax error at m.pl line 1, at EOF
//	foo(\n1,\n2         syntax error at m.pl line 3, at EOF
//	sub f {\n1;\n       Missing right curly or square bracket at m.pl line 2
//	sub f {\n if (1) {\n }\n   ... at m.pl line 3
//
// The line is the last token's, not the opener's: perl reads on to end of
// file looking for the closer. A quote-like with no closing delimiter is
// the lexer's UnknownRest, which perl reports on the line the quote starts:
// `q(abc\ndef` is "Can't find string terminator ")" anywhere before EOF at
// m.pl line 1". A dereference block left open, `@{`, is refused by the
// deref's own site before end of file is reached, also on perl's line 1:
// "Missing right curly or square bracket at m.pl line 1".
//
// `for(0A0` is not here: perl's error there is the missing operator, "Bareword
// found where operator expected ... near "0A0"", and the parser refuses it as
// trailing_tokens -- see TestForHeadWithTwoExpressionsRefuses.
//
// `0()`, the closed form of `0(`, is no negative: perl refuses it too, as
// "syntax error ... near "0("", so it is not listed below.
func TestUnclosedOpenerAtEOFRefused(t *testing.T) {
	cases := []struct {
		src  string
		line int
	}{
		{"0(", 1},
		{"for(0", 1},
		{"(", 1},
		{"[", 1},
		{"{", 1},
		{"sub f {", 1},
		{"my $x = (1", 1},
		{"my $h = {", 1},
		{"$h{", 1},
		{"$a[", 1},
		{"$x->{", 1},
		{"$x->[", 1},
		{"for(", 1},
		{"if (", 1},
		{"if (1) {", 1},
		{"foo(\n1,\n2", 3},
		{"sub f {\n1;\n", 2},
		{"sub f {\n if (1) {\n }\n", 3},
	}
	for _, c := range cases {
		n := parse.Parse([]byte(c.src))
		u := firstUnknown(n)
		if u == nil {
			t.Errorf("%q: accepted, perl refuses it", c.src)
			continue
		}
		if got := firstUnknownLine(n, []byte(c.src)); got != c.line || u.Refusal != parse.UnclosedOpener {
			t.Errorf("%q: refused %s on line %d, want %s on line %d",
				c.src, u.Refusal, got, parse.UnclosedOpener, c.line)
		}
	}

	elsewhere := []string{"@{", "q(abc", "q(abc\ndef", `"abc`, "s{a}{b", "qw(a b"}
	for _, src := range elsewhere {
		n := parse.Parse([]byte(src))
		if firstUnknown(n) == nil {
			t.Errorf("%q: accepted, perl refuses it", src)
			continue
		}
		if got := firstUnknownLine(n, []byte(src)); got != 1 {
			t.Errorf("%q: refused on line %d, perl refuses it on line 1", src, got)
		}
	}

	closed := []string{
		"()", "[]", "{}", "sub f {}", "my $x = (1)", "my $h = {}",
		"$h{a}", "$a[0]", "@{[]}", "$x->{a}", "$x->[0]", "for(;;){}",
		"if (1) {}", "foo(\n1,\n2)", "sub f {\n1;\n}", "for(0){A0}", "q(abc)",
	}
	for _, src := range closed {
		if u := firstUnknown(parse.Parse([]byte(src))); u != nil {
			t.Errorf("%q: refused %s, perl accepts it", src, u.Refusal)
		}
	}
}

// TestUnterminatedAttributeParameterRefused is the parser fuzzer's
// `my$0:A(}`. perl scans an attribute's argument as raw text to the `)`
// that balances its `(`, so a `}` inside does not close it. Measured on
// 5.42.0, `my $x :A(}` is "Unterminated attribute parameter in attribute
// list", while `my $x :A(})` compiles. Counted as one bracket by the
// end-of-file check, the `}` closed the `(`, and canon's `;` after it
// re-read as part of the argument.
func TestUnterminatedAttributeParameterRefused(t *testing.T) {
	for _, src := range []string{"my$0:A(}", "my $x :A(}", "sub f :A(} {1}"} {
		found := collectUnknown(parse.Parse([]byte(src)))
		if len(found) == 0 || found[0].Refusal != parse.UnclosedOpener {
			t.Errorf("%q: want an unclosed_opener refusal, got %v", src, refusalCodes(found))
		}
	}
	for _, src := range []string{"my $x :A(});", "my $x :A(x);", "my $x :A((x));"} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) != 0 {
			t.Errorf("%q: perl compiles this, but it refuses: %v", src, refusalCodes(found))
		}
	}
}
