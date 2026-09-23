// ABOUTME: Checks the glossary against the code and the code against perl.
// ABOUTME: A category the corpus can name must be defined, mapped, and measured.
package conformance

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestUnknownCategoryFails keeps an undefined category loud.
//
// A token fact naming a category nothing maps must not quietly assert
// zero. Silence here is the worst outcome available: the file looks like
// it constrains the token stream and constrains nothing, which is exactly
// the coverage illusion the corpus exists to prevent.
func TestUnknownCategoryFails(t *testing.T) {
	var got recorder
	checkTokenFact(&got, `one gerund whose text is "running"`, []byte("my $x = 1;"))

	if len(got.msgs) != 1 {
		t.Fatalf("checkTokenFact reported %d errors, want 1: %q", len(got.msgs), got.msgs)
	}
	if !strings.Contains(got.msgs[0], "gerund") {
		t.Errorf("error does not name the offending category: %q", got.msgs[0])
	}
	if !strings.Contains(got.msgs[0], "GLOSSARY.md") {
		t.Errorf("error does not say where to define it: %q", got.msgs[0])
	}
}

// TestGlossaryMatchesCategories couples the prose to the code.
//
// categories.go is the corpus's only point of coupling to our lexer, and
// GLOSSARY.md is the vocabulary corpus files are written in. A category in
// one and not the other means either a corpus file can name something with
// no definition behind it, or the glossary documents a category no file
// can use. Both are drift, and neither shows up in a passing suite.
func TestGlossaryMatchesCategories(t *testing.T) {
	defined := glossaryCategories(t)

	for name := range categories {
		if !defined[name] {
			t.Errorf("categories.go maps %q but GLOSSARY.md does not define it", name)
		}
	}
	for name := range defined {
		if _, ok := categories[name]; !ok {
			t.Errorf("GLOSSARY.md defines %q but categories.go does not map it", name)
		}
	}
}

// glossaryCategories reads the `## <name>` headings that define a category.
func glossaryCategories(t *testing.T) map[string]bool {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(corpusDir, "GLOSSARY.md"))
	if err != nil {
		t.Fatalf("reading the glossary: %v", err)
	}

	out := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?m)^## (.+)$`).FindAllStringSubmatch(string(raw), -1) {
		out[strings.TrimSpace(m[1])] = true
	}
	return out
}

// TestCategoryBoundaries asserts the glossary's measured claims.
//
// Every boundary case in GLOSSARY.md was decided by running perl, and each
// is a statement about where one token ends and the next begins. Prose
// alone rots: the claim stays on the page while the lexer drifts. These
// are the same claims, executable.
//
// A case our lexer gets WRONG is recorded here with its measurement rather
// than omitted, because the glossary is the record of what perl does, not
// of what we currently produce.
func TestCategoryBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name     string
		src      string
		category string
		text     string
		want     int
		known    string // non-empty: a refusal, skipped with this reason
	}{{
		// The asymmetry the glossary exists to record: a leading minus is
		// an operator, an exponent sign is part of the token.
		name: "a leading minus is not part of the literal",
		src:  "my $x = -1;", category: "numeric literal", text: "1", want: 1,
	}, {
		name: "the minus in -1 is an operator",
		src:  "my $x = -1;", category: "operator", text: "-", want: 1,
	}, {
		// No `known`: this was skipped until the lexer learned that a `.`
		// before a digit starts a number where a term is expected, under
		// M1 issue 01a0c13f-97f5. It is an ordinary passing row now.
		name: "a leading decimal point is part of the literal",
		src:  "my $x = .5;", category: "numeric literal", text: ".5", want: 1,
	}, {
		name: "a trailing decimal point is part of the literal",
		src:  "my $x = 1.;", category: "numeric literal", text: "1.", want: 1,
	}, {
		name: "an exponent sign is part of the literal",
		src:  "my $x = 5e-1;", category: "numeric literal", text: "5e-1", want: 1,
		known: "the lexer splits 5e-1 into Number(5e) Operator(-) Number(1); found by 08_signed_exponent.t",
	}, {
		name: "an unsigned exponent is one token",
		src:  "my $x = 5e1;", category: "numeric literal", text: "5e1", want: 1,
	}, {
		name: "underscores separate digits within one token",
		src:  "my $x = 4_294_967_296;", category: "numeric literal", text: "4_294_967_296", want: 1,
	}, {
		name: "hexadecimal is one token",
		src:  "my $x = 0xff;", category: "numeric literal", text: "0xff", want: 1,
	}, {
		// Two dots make it a v-string unambiguously, so this is the case
		// that would be a number if the rule were "digits and dots".
		// Measured under 5.42.0: `printf "%vd"` prints 5.42.0 and `$x+0`
		// is 0, so perl holds a string of ordinals, not the number 5.42.
		name: "a v-string is not a numeric literal",
		src:  "my $x = 5.42.0;", category: "numeric literal", text: "5.42.0", want: 0,
		known: "the lexer reads 5.42.0 as Number; perl makes it a v-string (%vd prints 5.42.0, $x+0 is 0)",
	}, {
		name: "a variable carries its sigil",
		src:  "my $x = 1;", category: "variable", text: "$x", want: 1,
	}, {
		name: "a quoted string is one token including its delimiters",
		src:  `my $x = "hi";`, category: "string literal", text: `"hi"`, want: 1,
	}, {
		// The distinction the two categories exist for: qw is a LIST
		// (measured, qw(a b c) has three elements), so a file asserting
		// a string literal must not be satisfied by one.
		name: "qw is a quote-like operator, not a string literal",
		src:  "my @w = qw(a b);", category: "quote-like operator", text: "qw(a b)", want: 1,
	}, {
		name: "qw is not counted as a string literal",
		src:  "my @w = qw(a b);", category: "string literal", text: "qw(a b)", want: 0,
	}, {
		name: "a bare quoted string is not a quote-like operator",
		src:  `my $x = "hi";`, category: "quote-like operator", text: `"hi"`, want: 0,
	}, {
		name: "a substitution is a quote-like operator",
		src:  "$x =~ s/a/b/;", category: "quote-like operator", text: "s/a/b/", want: 1,
	}, {
		// Bracketing delimiters nest: measured, q{a{b}c} is the string
		// a{b}c, so the inner brace is content and not a terminator.
		name: "a bracketing delimiter nests within one token",
		src:  "my $x = q{a{b}c};", category: "quote-like operator", text: "q{a{b}c}", want: 1,
	}, {
		// qx runs a command, like backticks. It was missing from the
		// operator list at first, which made it read as a plain string.
		name: "qx is a quote-like operator",
		src:  "my $x = qx/echo hi/;", category: "quote-like operator", text: "qx/echo hi/", want: 1,
	}, {
		// The same operation spelled with no operator name at all.
		// Measured, `qx/echo hi/` and `` `echo hi` `` both run the
		// command, so the corpus must not call one a string literal.
		name: "backticks are a quote-like operator despite having no name",
		src:  "my $x = `echo hi`;", category: "quote-like operator", text: "`echo hi`", want: 1,
	}, {
		// The text merely CONTAINS an operator name; it does not start
		// with one, so the delimiters decide and this stays a string.
		name: "a string whose content starts with an operator name is still a string",
		src:  `my $x = "qw stuff";`, category: "string literal", text: `"qw stuff"`, want: 1,
	}, {
		name: "transliteration is a quote-like operator",
		src:  "$x =~ tr/a/b/;", category: "quote-like operator", text: "tr/a/b/", want: 1,
	}, {
		name: "a compiled pattern is a quote-like operator",
		src:  "my $r = qr/pat/;", category: "quote-like operator", text: "qr/pat/", want: 1,
	}, {
		// perl accepts a word-character delimiter when whitespace
		// separates it from the name: `q xax` is the string "a". This
		// looks like it should defeat a prefix test and does not, since
		// the space after `q` is the non-word byte the check wants.
		name: "a word-character delimiter after whitespace is still an operator",
		src:  "my $x = q xax;", category: "quote-like operator", text: "q xax", want: 1,
	}, {
		name: "a heredoc opener is one token",
		src:  "my $h = <<EOT;\nbody\nEOT\n", category: "heredoc opener", text: "<<EOT", want: 1,
	}, {
		// The reordering that makes heredocs hard: the body arrives
		// after the semicolon, because the content starts on the next
		// LINE while the statement continues on the same one.
		name: "a heredoc body is one token however many lines it spans",
		src:  "my $h = <<EOT;\nfirst\nsecond\nEOT\n", category: "heredoc body", text: "first\nsecond\nEOT\n", want: 1,
	}, {
		name: "an indentation-stripping opener keeps its tilde",
		src:  "my $h = <<~EOT;\n  body\n  EOT\n", category: "heredoc opener", text: "<<~EOT", want: 1,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			match, ok := categories[tc.category]
			if !ok {
				t.Fatalf("category %q is not mapped in categories.go", tc.category)
			}

			src := []byte(tc.src)
			got := 0
			for _, tk := range significant(src) {
				tokenText := string(src[tk.Start:tk.End])
				if tokenText == tc.text && match(tk.Kind, tokenText) {
					got++
				}
			}

			// The measurement runs BEFORE the skip is honoured, so a
			// `known` that has been fixed FAILS rather than skipping on.
			//
			// This ordering is the fix for a real rot: the leading-decimal
			// row carried a `known` citing a lexer bug for a day after the
			// bug was fixed, and nothing said so -- the skip fired first
			// and the row never ran. The corpus has
			// `TestStaleRefusalMarkerFails` for exactly this failure mode
			// on `.t` files; this table was outside its reach.
			//
			// A skip that cannot go stale is worth more than a skip that
			// is merely documented, because in a summary line a skipped
			// row and a passing row look the same.
			if tc.known != "" {
				if got == tc.want {
					t.Fatalf("marked `known` -- %s -- but the lexer now "+
						"produces %d %s token(s) whose text is %q, which "+
						"is what the row wants.\n"+
						"\tRemove the `known` field: a stale marker hides "+
						"the next regression.",
						tc.known, got, tc.category, tc.text)
				}
				t.Skipf("known refusal: %s", tc.known)
			}

			if got != tc.want {
				t.Errorf("%s: %d %s tokens whose text is %q, want %d\n\tactual tokens: %s",
					tc.src, got, tc.category, tc.text, tc.want, describe(src))
			}
		})
	}
}
