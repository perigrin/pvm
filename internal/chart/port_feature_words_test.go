// ABOUTME: Tests the feature-gated keywords, which toke.c reads as plain words with their feature off.
// ABOUTME: Recognition cannot see a feature, so it gives both readings; resolve keeps one by feature (forest.TestResolveFeatures).

package chart

import "testing"

// perl -c at 94e50866 accepts each: with its feature off keywords.c does not return the KEY_, and toke.c reads the
// word as yyl_just_a_word. The first three are perl.git t/ forms (base/term.t and run/script.t, opbasic/arith.t,
// op/args.t); the catch rows are input 01a1c678.
func TestFeatureWordsReadAsWords(t *testing.T) {
	for _, src := range []string{
		`open(try, "/dev/null");`,
		`my $T = 1; sub try ($$$) {} try $T++, 13 % 4, 1;`,
		`sub try { eval { method("foo", "bar"); }; } for (1..5) { try() }`,
		`sub defer {} defer(1);`,
		`foo(catch, 1);`, `foo(catch, $x);`, `print catch . $_;`,
		`my $x = method;`, `my $x = field;`, `my $x = state;`, `my $x = finally;`,
		`sub given {} given 1;`, `when(1);`, `sub default {} default 1;`,
		`ADJUST(1);`, `class(1);`, `sub all {} my @x = all 1, 2;`, `@y = sort try @x;`,
		`my $x = __CLASS__;`, `my $x = __SUB__;`, `my $x = break;`, `my $x = fc;`, `my $x = evalbytes;`, `my $x = any;`,
	} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d, expecting %v; perl accepts it", src, v.at, v.expected)
		}
	}
}

// perl -c at 94e50866 accepts each with the feature on.
func TestFeatureWordsKeepKeywordReadings(t *testing.T) {
	for _, src := range []string{
		`use feature "try"; no warnings; try { 1 } catch ($e) { 2 }`,
		`use feature "defer"; no warnings; defer { 1 }`,
		`use v5.38; use experimental "class"; class C { field $x; method m { 1 } ADJUST { 1 } }`,
		`use feature "keyword_any"; no warnings; my @a; my $x = any { $_ } @a;`,
	} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d, expecting %v; perl accepts it", src, v.at, v.expected)
		}
	}
}

// perl -c at 94e50866 refuses each: no feature gates these keywords, so toke.c never reads them as words.
func TestFeatureWordsLeaveCoreKeywords(t *testing.T) {
	for _, src := range []string{`foo(if, 1);`, `foo(while, 1);`, `foo(my, 1);`, `my $x = CORE::try;`} {
		if recognizeSource(t, []byte(src)).accepted {
			t.Errorf("%q is accepted; perl refuses it", src)
		}
	}
}
