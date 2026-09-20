// ABOUTME: A NAMED hash's op depends on its declaration; a DEREFERENCE always emits one.
// ABOUTME: The first is undecidable from text alone and hedges; the second commits.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
)

// hedgedSite reports whether a site of this kind is present AND unresolved.
func hedgedSite(sites []parseoracle.SubjectCallSite, kind string) bool {
	for _, s := range sites {
		if s.Kind == kind && s.Unresolved {
			return true
		}
	}
	return false
}

// TestUndecidableHashHedges: a NAMED hash access hedges, because whether
// perl emits rv2hv depends on a declaration the text cannot carry.
//
// Measured on 5.42.0, identical source text either side:
//
//	my %h;    keys %h            0 rv2hv   (padhv -- a lexical lives in the pad)
//	our %h;   keys %h            1 rv2hv
//	my %h;    @h{"a","b"}        0 rv2hv
//	our %h;   @h{"a","b"}        1 rv2hv
//	my %h;    delete $h{a}       0 rv2hv
//	our %h;   delete local $h{a} 1 rv2hv
//
// `state %h` is the trap: a LEXICAL declaration that behaves like a package
// hash here (1 rv2hv), so a rule keyed on "was it declared with my" gets it
// wrong. Hedging sidesteps the whole question.
//
// A hedge scores WIDER, which is the honest bucket for "I see a hash access
// and cannot tell whether perl will emit a dereference for it", and is what
// the four-bucket design exists to reward.
func TestUndecidableHashHedges(t *testing.T) {
	for _, src := range []string{
		`my $n = keys %h;`,
		`my @x = @h{"a","b"};`,
		`delete $h{a};`,
		`my @x = delete @h{"a","b"};`,
		`delete local $h{a};`,
	} {
		sites := parse.Sites(parse.Parse([]byte(src)), []byte(src))
		if !hedgedSite(sites, parseoracle.SiteKindHash) {
			t.Errorf("%s\n  does not hedge; whether perl emits rv2hv here "+
				"depends on a declaration the text cannot carry", src)
		}
	}
}

// TestDecidableHashStillCommits: an explicit DEREFERENCE always emits the
// op, so the subject commits and keeps earning exact verdicts.
//
// This is the half that must not collapse. Hedging every hash site would
// take WRONG to zero and exact with it, which is the vacuous gate
// TestGoParserExactFloor exists to prevent. Measured -- every one of these
// is 1 or more regardless of how anything was declared:
//
//	keys %$r            1      keys %{$s{a}}      1
//	keys %{$r}          1      keys %{$s{a}{b}}   2
//	@$r{"a","b"}        1      $r->%*             1
//	$r->@{"a","b"}      1      keys %ENV          1
func TestDecidableHashStillCommits(t *testing.T) {
	for _, src := range []string{
		`my $n = keys %$r;`,
		`my $n = keys %{$r};`,
		`my @x = @$r{"a","b"};`,
	} {
		sites := parse.Sites(parse.Parse([]byte(src)), []byte(src))
		if !hasSite(sites, parseoracle.SiteKindHash) {
			t.Errorf("%s\n  reports no decided hash site; a dereference emits "+
				"rv2hv whatever the declaration", src)
		}
		if hedgedSite(sites, parseoracle.SiteKindHash) {
			t.Errorf("%s\n  hedges a site it can decide; that costs an exact "+
				"verdict for nothing", src)
		}
	}

	// A NESTED access is both at once, and reporting both is right.
	// `%{$s{a}}` dereferences -- always an op -- an element whose own op
	// depends on how %s was declared. Measured:
	//
	//	my %s;  keys %{$s{a}}      1 rv2hv   (the element folds to multideref)
	//	my %s;  keys %{$s{a}{b}}   2 rv2hv   (the inner element does not fold)
	//
	// So one decided site for the dereference plus one hedge for the
	// element is the honest report, and suppressing the hedge would
	// under-report the second case.
	for _, src := range []string{
		`my @x = keys %{$s{a}};`,
		`my @x = keys %{$s{a}{b}};`,
	} {
		sites := parse.Sites(parse.Parse([]byte(src)), []byte(src))
		if !hasSite(sites, parseoracle.SiteKindHash) {
			t.Errorf("%s\n  reports no decided hash site for the dereference", src)
		}
		if !hedgedSite(sites, parseoracle.SiteKindHash) {
			t.Errorf("%s\n  does not hedge the inner element, whose op depends "+
				"on how the base hash was declared", src)
		}
	}
}
