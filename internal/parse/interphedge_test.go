// ABOUTME: Code inside an interpolated string, regex or heredoc is never parsed, so it hedges
// ABOUTME: every marker; a plain `"$x"` or `"@a"` holds no code and stays decided.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// TestInterpolationHedgesItsSites: perl compiles the code inside `${\ ...}`,
// `@{[ ... ]}`, a subscript or a postfix dereference in an interpolating
// string, regex or heredoc into ordinary ops, and the oracle counts their
// markers. internal/parse keeps the whole literal as one leaf, so silence
// there is the claim that nothing is inside -- WRONG wherever perl found
// something. The plan's M1 rule: what is not understood is reported as
// unknown. Measured on 5.42.0 with -MO=Concise,-exec:
//
//	qr/x ${\f($o)} y/                        refgen     op/coreamp.t:178
//	"@{[ g(\&g) ]}"                          srefgen    op/sub_lval.t:861
//	/b${\do { qr|a| }}/                      refgen     re/reg_eval_scope.t:99
//	"$_->@{foo}" under postderef_qq          rv2hv      op/postfixderef.t:340
//	qr/${\ <<END ... END }/                  srefgen    comp/parser.t:468
//	"@{[ {a=>1} ]}"                          anonhash
//	"@{[ <STDIN> ]}"                         readline
//	"@{[ m/x/ ]}"                            match
//	"@h{qw(a b)}" of `our %h`                rv2hv
//	<<"E" body `${\ \@a}`                    srefgen
//
// The real parse of the embedded code is 01a12059-3011; until then these hedge.
func TestInterpolationHedgesItsSites(t *testing.T) {
	for _, c := range []struct {
		src  string
		kind string
	}{
		{`sub f {} my $o; my $r = qr/x ${\f($o)} y/;`, parseoracle.SiteKindReference},
		{`sub g {} my $s = "@{[ g(\&g) ]}";`, parseoracle.SiteKindReference},
		{`"ba" =~ /b${\do { qr|a| }}/;`, parseoracle.SiteKindReference},
		{`use feature "postderef_qq"; my $s = "$_->@{foo}";`, parseoracle.SiteKindHash},
		{"my $r = qr/${\\ <<END\nblah\nEND\n }/;\n", parseoracle.SiteKindReference},
		{`my $s = "@{[ {a=>1} ]}";`, parseoracle.SiteKindAnonhash},
		{`my $s = "@{[ <STDIN> ]}";`, parseoracle.SiteKindReadline},
		{`my $s = "@{[ m/x/ ]}";`, parseoracle.SiteKindMatch},
		{`our %h; my $s = "@h{qw(a b)}";`, parseoracle.SiteKindHash},
		{`my $s = qq{a ${\ \@ARGV} b};`, parseoracle.SiteKindReference},
		{"my @a; my $s = <<\"E\";\n${\\ \\@a}\nE\n", parseoracle.SiteKindReference},
		{"my @a; my $s = <<E;\n${\\ \\@a}\nE\n", parseoracle.SiteKindReference},
		// A body inside its statement, which continues past the opener's line.
		{"my @a; my $s = join \"\", <<\"E\",\n${\\ \\@a}\nE\n  \"x\";\n", parseoracle.SiteKindReference},
	} {
		if !has(hedgedKinds(c.src), c.kind) {
			t.Errorf("%q: perl reports a %s site inside the interpolation; the subject "+
				"did not hedge one (hedged %v, decided %v)",
				c.src, c.kind, hedgedKinds(c.src), decidedKinds(c.src))
		}
	}
}

// TestInterpolationKeepsThePatternsOwnMatch: hedging what a pattern's
// interpolation holds must not take back the match the pattern itself is.
// perl emits refgen AND match for `"ba" =~ /b${\do { qr|a| }}/`.
func TestInterpolationKeepsThePatternsOwnMatch(t *testing.T) {
	src := `"ba" =~ /b${\do { qr|a| }}/;`
	if !has(decidedKinds(src), parseoracle.SiteKindMatch) {
		t.Errorf("%q: the match is decided by the binding; got decided %v", src, decidedKinds(src))
	}
}

// TestPlainInterpolationDoesNotHedge is the negative half: a plain variable
// holds no code, and perl reports none of the five markers for it --
// measured. Nor does a literal that does not interpolate, or an escaped
// sigil, or a `->` that is not a dereference (perl interpolates `$x` and
// leaves `->method` as text).
func TestPlainInterpolationDoesNotHedge(t *testing.T) {
	for _, src := range []string{
		`my $x; my $s = "$x";`,
		`my @a; my $s = "@a";`,
		`our $x; my $s = "${x}";`,
		`our @a; my $s = "@{a} $::x ${^WARNING_BITS}";`,
		`my $x; my $s = "$x->method";`,
		`my $s = q{${\ f()}};`,
		`my $s = '@{[ f() ]}';`,
		`my $s = "\${\\ f()} \@{[ 1 ]}";`,
		`my $s = "mail@example.com costs $5.00";`,
		`my $x; my $r = $x =~ /foo$/;`,
		`my $x; my $r = $x =~ m'${\ f()}';`,
	} {
		if got := hedgedKinds(src); len(got) != 0 {
			t.Errorf("%q: no code is interpolated here, want no hedge, got %v", src, got)
		}
	}
}
