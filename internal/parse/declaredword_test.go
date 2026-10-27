// ABOUTME: A word's following `<<`, `%`, `<`, `&` or `*` reads as a term when the word is a known sub.
// ABOUTME: Only a word not known to be declared is refused (ambiguous_sigil), never read as an operator.

package parse_test

import (
	"os"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

const perl544Lib = "/root/.local/share/pvm/versions/5.44.0/lib/perl5/5.44.0"

// TestDeclaredWordTakesTerm: perl 5.44.0 compiles each of these with the
// word a call and what follows its argument -- a heredoc, the hash `%h`, a
// readline, the sub `&!`, the glob `*-`. The lexer has no symbol table, so
// it reads the term; the parser, which knows declarations, keeps it.
func TestDeclaredWordTakesTerm(t *testing.T) {
	for src, callee := range map[string]string{
		"sub foo {} foo << \"H\";\nx\nH\n": "foo",
		"sub foo; foo << \"H\";\nx\nH\n":   "foo",
		"my %h; sub foo {} foo % h;":       "foo",
		"sub A {} A < 0 >;":                "A",
		"sub A {} A &!;":                   "A",
		"sub A {} A & ;":                   "A",
		"sub A {} A *-;":                   "A",
		"print STDOUT *-;":                 "print",
		"sub A {} A *$;;":                  "A",
		"sub A {} A *%;":                   "A",
	} {
		root := parse.Parse([]byte(src))
		if found := collectUnknown(root); len(found) != 0 || root.HasError {
			t.Errorf("%q: perl compiles this, but it refuses: %v", src, refusalCodes(found))
			continue
		}
		if c := findCall(root, callee); c == nil || len(c.Children) == 0 {
			t.Errorf("%q: %s takes no argument:\n%s", src, callee, dumpTree(root, 0))
		}
		if ok, why := canonMatches([]byte(src)); !ok {
			t.Errorf("%q: canon is not a fixpoint: %s", src, why)
		}
	}
}

// TestImportedWordTakesHeredoc: `croak << "HERE"` is core File/Temp.pm's
// shape, a heredoc to the croak Carp exports. With the loader the import is
// known and the heredoc stands.
func TestImportedWordTakesHeredoc(t *testing.T) {
	if _, err := os.Stat(perl544Lib + "/Carp.pm"); err != nil {
		t.Skip("perl 5.44.0's Carp.pm is not installed")
	}
	for src, callee := range map[string]string{
		"use Carp; croak << \"HERE\";\nx\nHERE\n":     "croak",
		"use Carp qw(croak); croak << 'H';\nx\nH\n":   "croak",
		"use Carp; confess << \"H\" . \"y\";\nx\nH\n": "confess",
		"use Carp; croak <<\"H\";\nx\nH\n":            "croak",
		"use Carp; croak << \"H\", 1;\nx\nH\n":        "croak",
		"use Carp; my %h; croak % h;":                 "croak",
	} {
		root := parse.ParseWithLoader([]byte(src), parse.DirLoader(perl544Lib))
		if found := collectUnknown(root); len(found) != 0 || root.HasError {
			t.Errorf("%q: perl compiles this, but it refuses: %v", src, refusalCodes(found))
			continue
		}
		if c := findCall(root, callee); c == nil || len(c.Children) == 0 {
			t.Errorf("%q: %s takes no argument:\n%s", src, callee, dumpTree(root, 0))
		}
	}
}

// TestUndeclaredWordRefuses: a word not known to be declared is a bareword
// to perl, and what follows it an operator: `croak << "H"` a shift, `A % h`
// a modulus. The parser cannot tell an unseen import from an undeclared
// word, so it refuses (ruling R-b′) rather than read either.
func TestUndeclaredWordRefuses(t *testing.T) {
	for _, src := range []string{
		"croak << \"H\";\nx\nH\n", "A % h;", "A < 0 > 0;", "A <= 0 > 0;", "A */x/;", "A *.5;", "A *%;",
	} {
		found := collectUnknown(parse.Parse([]byte(src)))
		if len(found) == 0 || found[0].Refusal != parse.AmbiguousSigil {
			t.Errorf("%q: want an ambiguous_sigil refusal, got %v", src, refusalCodes(found))
		}
	}
	// Without the loader croak is unseen: whatever the parser says, it is
	// not a shift.
	if root := parse.Parse([]byte("use Carp; croak << \"H\";\nx\nH\n")); collectUnknown(root) == nil {
		t.Errorf("`croak << \"H\"` without the loader parses:\n%s", dumpTree(root, 0))
	}
	// perl 5.44.0 rejects these after an undeclared word, so any refusal
	// will do; declared, they are TestDeclaredWordTakesTerm's.
	for _, src := range []string{"A &!;", "A *-;"} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) == 0 {
			t.Errorf("%q: perl rejects this, but it parses", src)
		}
	}
}

// TestFileTempParsesAsOnPu: core File/Temp.pm, read with its imports, has
// no refusal, as it had before the canon fixpoint finds were fixed.
func TestFileTempParsesAsOnPu(t *testing.T) {
	path := perl544Lib + "/File/Temp.pm"
	src, err := os.ReadFile(path)
	if err != nil {
		t.Skip("perl 5.44.0's File/Temp.pm is not installed")
	}
	if found := collectUnknown(parse.ParseWithLoader(src, parse.DirLoader(perl544Lib))); len(found) != 0 {
		t.Errorf("File/Temp.pm: %d refusals, want 0: %v", len(found), refusalCodes(found))
	}
}
