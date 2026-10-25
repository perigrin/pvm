// ABOUTME: Canon writes `__END__` or `__DATA__` and everything after it back verbatim, and Faithful
// ABOUTME: compares that tail as code: what DATA yields is the program's input, so dropping it changes the run.

package parse_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/perltest"
)

// dataSections are files with a data section, each read back through DATA
// or followed by POD, in a package other than main, and with CRLF line ends.
// perigrin's ruling (2026-10-25): canon writes the section verbatim and
// Faithful compares it exactly; it is not a dropped token kind. Measured on
// perl 5.44.0, the first prints "line one\nline two\n", and Deparse,-p
// writes each section after `__DATA__` as the source has it.
var dataSections = []string{
	"print while <DATA>;\n__DATA__\nline one\nline two\n",
	"print \"code\\n\";\n__END__\n\n=head1 NAME\n\npod here\n\n=cut\n",
	"package Foo;\nprint while <DATA>;\n__DATA__\nfoo data\n",
	"print while <DATA>;\r\n__END__\r\ncrlf data\r\n",
	"print while <DATA>;\n__END__",
	"sub f { 1; }\n__DATA__\n) not ( perl $$$ {\n",
	// The last statement keeps the source's missing `;` before the section.
	"print 1\n__END__\nx\n",
}

// TestCanonKeepsDataSection: canon ends with the source's data section,
// byte for byte, the file is faithful with nothing forgiven, canon is a
// fixpoint, and perl runs canon as it runs the source.
func TestCanonKeepsDataSection(t *testing.T) {
	for _, src := range dataSections {
		faithfulUnforgiven(t, src)
		at := strings.Index(src, "__END__")
		if d := strings.Index(src, "__DATA__"); at < 0 || d >= 0 && d < at {
			at = d
		}
		canon := parse.Canon(parse.Parse([]byte(src)), []byte(src))
		if !strings.HasSuffix(canon, src[at:]) {
			t.Errorf("canon of %q is %q, without the data section %q", src, canon, src[at:])
		}
		if again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon)); again != canon {
			t.Errorf("canon of %q is %q, and canon of that is %q", src, canon, again)
		}
	}
	t.Run("perl", func(t *testing.T) {
		perl := perltest.PinnedPerl(t)
		dir := t.TempDir()
		run := func(name, text string) string {
			path := filepath.Join(dir, name)
			if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
				t.Fatal(err)
			}
			out, err := perl.Command(path).CombinedOutput()
			if err != nil {
				t.Errorf("perl %q: %v\n%s", text, err, out)
			}
			return string(out)
		}
		for _, src := range dataSections {
			canon := parse.Canon(parse.Parse([]byte(src)), []byte(src))
			if a, b := run("src.pl", src), run("canon.pl", canon); a != b {
				t.Errorf("perl prints %q for %q but %q for its canon %q", a, src, b, canon)
			}
		}
	})
}

// TestFaithfulDataSectionChangeIsADifference: the data section is one token,
// compared like any other, so a canon that drops it or alters a byte of it
// is unfaithful.
func TestFaithfulDataSectionChangeIsADifference(t *testing.T) {
	src := "print while <DATA>;\n__DATA__\nline one\n"
	for _, canon := range []string{
		"print while <DATA>;",
		"print while <DATA>;\n__DATA__\nline 1\n",
		"print while <DATA>;\n__DATA__\nline one\r\n",
		"print while <DATA>;\n__END__\nline one\n",
	} {
		if ok, forgiven, _ := parse.FaithfulTokens(sigTokens(src), asCanon(sigTokens(canon))); ok {
			t.Errorf("%q against canon %q: forgiven by %v", src, canon, ruleNames(forgiven))
		}
	}
}
