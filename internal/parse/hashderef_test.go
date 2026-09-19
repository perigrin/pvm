// ABOUTME: A hash DEREFERENCE is an rv2hv, and the harness scores it like any other hash site.
// ABOUTME: Reporting nothing there is the silent wrong answer Unknown's hedge exists to prevent.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
)

// TestHashDerefIsAHashSite: `%$href` reports a hash site.
//
// isHashTerm fires on a leaf whose TEXT starts with `%`, which catches
// `%hash` and misses every dereference: `%$href` is a Unary `%` over a Term
// `$href`, so the Term case sees only `$href` and the statement reports no
// hash at all.
//
// perl emits rv2hv for both, so the subject was committing to a statement
// with no such site and no hedge -- WRONG, the one bucket that fails a
// build. Measured over T2 at ef0efc46:
//
//	cmd/subval.t  rv2hv  perl read a hash at 2 site(s) the subject did not,
//	              and at lines 182, 184 the subject committed to the
//	              statement with no such site and no hedge
//
// Both lines are `%$href`.
func TestHashDerefIsAHashSite(t *testing.T) {
	for _, src := range []string{
		`print keys %$href;`,
		`print join(':', %$href);`,
		`my @k = keys %{$h};`,
		`my @k = keys %{$r->{x}};`,
		`my %copy = %$href;`,
	} {
		if !hasSite(parse.Sites(parse.Parse([]byte(src)), []byte(src)), parseoracle.SiteKindHash) {
			t.Errorf("%s\n  reports no hash site; perl emits rv2hv here", src)
		}
	}

	// The plain form must keep working.
	for _, src := range []string{
		`my @k = keys %hash;`,
		`my %copy = %other;`,
	} {
		if !hasSite(parse.Sites(parse.Parse([]byte(src)), []byte(src)), parseoracle.SiteKindHash) {
			t.Errorf("%s\n  reports no hash site", src)
		}
	}

	// An ARRAY dereference is not a hash. `@$aref` is rv2av, a different op,
	// and claiming a hash there would be the same defect pointing the other
	// way.
	for _, src := range []string{
		`my @c = @$aref;`,
		`my @c = @{$r->{x}};`,
	} {
		if hasSite(parse.Sites(parse.Parse([]byte(src)), []byte(src)), parseoracle.SiteKindHash) {
			t.Errorf("%s\n  reports a hash site; perl emits rv2av, not rv2hv", src)
		}
	}
}

// hasSite reports whether a decided (non-hedged) site of this kind is present.
func hasSite(sites []parseoracle.SubjectCallSite, kind string) bool {
	for _, s := range sites {
		if s.Kind == kind && !s.Unresolved {
			return true
		}
	}
	return false
}
