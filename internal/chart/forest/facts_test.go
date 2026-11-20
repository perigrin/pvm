// ABOUTME: Tests SubjectFacts, the chart as a parseoracle subject, on committed samples decoded by parseoracle itself.
// ABOUTME: Each sample states its facts; the pinned perl -c accepts every sample (testdata/subject).

package forest

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// TestOracleSubject computes each sample's facts, marshals them, and has parseoracle decode them as it decodes any
// subject's, through a subject command that prints them; the decoded facts must be the ones the sample states.
func TestOracleSubject(t *testing.T) {
	tab := committedTables(t)
	samples := []struct{ file, want string }{
		// a call, under a prototype or none: f's \@ slot takes @a by reference, fitted slot by slot (protoRefs); g's
		// prototype resolve cannot see, so it is Unresolved; h, declared with a $ slot, takes none
		{"call.pl", `{"ok":true,"prototypes":{"f":"\\@","h":"$"},"call_sites":[
			{"line":3,"end_line":3,"name":"f","took_reference":false},
			{"kind":"reference","line":3,"end_line":3,"took_reference":true},
			{"line":4,"end_line":4,"name":"g","took_reference":false,"unresolved":true},
			{"line":6,"end_line":6,"name":"h","took_reference":false}]}`},
		// a call to a word nothing declares, where a module resolve cannot see may give it a \ prototype: perl
		// takes @array by reference here, so the call is Unresolved, not "took no reference"
		{"undeclared.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"line":3,"end_line":3,"name":"is_shared","took_reference":false,"unresolved":true}]}`},
		// a \ the source wrote, on its statement's first and last lines; code interpolated in a string or a pattern
		// is one token the forest does not read, so its statement hedges every marker (parse.InterpolatesCode),
		// while the pattern's own match stays decided
		{"reference.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"reference","line":2,"end_line":2,"took_reference":true},
			{"kind":"reference","line":3,"end_line":4,"took_reference":true},
			{"kind":"match","line":6,"end_line":6},
			{"kind":"reference","line":5,"end_line":5,"took_reference":false,"unresolved":true},
			{"kind":"hash","line":5,"end_line":5,"unresolved":true},
			{"kind":"match","line":5,"end_line":5,"unresolved":true},
			{"kind":"readline","line":5,"end_line":5,"unresolved":true},
			{"kind":"anonhash","line":5,"end_line":5,"unresolved":true},
			{"kind":"reference","line":6,"end_line":6,"took_reference":false,"unresolved":true},
			{"kind":"hash","line":6,"end_line":6,"unresolved":true},
			{"kind":"match","line":6,"end_line":6,"unresolved":true},
			{"kind":"readline","line":6,"end_line":6,"unresolved":true},
			{"kind":"anonhash","line":6,"end_line":6,"unresolved":true}]}`},
		// a hash read as perl -MO=Concise builds it: our %h and \%h are rv2hv; $h{a} and $r->{b}{c} fold into
		// multideref; %$r and @h{'a'} are rv2hv (TestOracleSubjectHashSites)
		{"hash.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"hash","line":1,"end_line":1},
			{"kind":"reference","line":2,"end_line":2,"took_reference":true},
			{"kind":"hash","line":2,"end_line":2},
			{"kind":"hash","line":3,"end_line":3},{"kind":"hash","line":3,"end_line":3}]}`},
		// a match: =~ against a pattern or an expression, and a bare pattern; not s///, not division
		{"match.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"match","line":2,"end_line":2},{"kind":"match","line":3,"end_line":3},{"kind":"match","line":5,"end_line":5}]}`},
		// a readline: <STDIN> and readline(...); <*.c> is a glob
		{"readline.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"readline","line":1,"end_line":1},{"kind":"readline","line":3,"end_line":3}]}`},
		// an anonymous hash, braced and after +
		{"anonhash.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"anonhash","line":1,"end_line":1},{"kind":"anonhash","line":2,"end_line":2}]}`},
		// Foo::bar's shape is unknown, so its { a => 1 } is an anonymous hash in some readings and a block in
		// others: Unresolved, as the call is; the \@a every reading holds is decided
		{"ambiguous.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"line":2,"end_line":2,"name":"Foo::bar","took_reference":false,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":2,"unresolved":true},
			{"kind":"reference","line":3,"end_line":3,"took_reference":true}]}`},
		// six ambiguous statements make the block 4^6 readings, above 2^10, so it collapses; a collapsed statement
		// hedges only the sites its readings disagree on (perigrin, 2026-11-16, "Narrow it"): each Foo::bar and its
		// { a => 1 } stay Unresolved, and the \@a every reading holds is decided, as is the one after the block
		{"collapsed.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"line":2,"end_line":6,"name":"Foo::bar","took_reference":false,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":6,"unresolved":true},
			{"line":2,"end_line":6,"name":"Foo::bar","took_reference":false,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":6,"unresolved":true},
			{"line":2,"end_line":6,"name":"Foo::bar","took_reference":false,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":6,"unresolved":true},
			{"line":2,"end_line":6,"name":"Foo::bar","took_reference":false,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":6,"unresolved":true},
			{"line":2,"end_line":6,"name":"Foo::bar","took_reference":false,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":6,"unresolved":true},
			{"line":2,"end_line":6,"name":"Foo::bar","took_reference":false,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":6,"unresolved":true},
			{"kind":"reference","line":2,"end_line":6,"took_reference":true},
			{"kind":"reference","line":7,"end_line":7,"took_reference":true}]}`},
		// under use utf8 a handle's name is runes, not bytes: perl reads <FÉ> as readline, not a glob
		{"utf8readline.pl", `{"ok":true,"prototypes":{},"call_sites":[{"kind":"readline","line":3,"end_line":3}]}`},
		// without use utf8 in force, perl reads <FÉ> as a glob: none at all, one a block ended, one no utf8 turned off
		{"noutf8readline.pl", `{"ok":true,"prototypes":{},"call_sites":[]}`},
		// known limit (facts.go utf8At): a utf8 a BEGIN block imports is not seen, so <FÉ> reports nothing where
		// perl reads it as readline on line 2
		{"beginutf8readline.pl", `{"ok":true,"prototypes":{},"call_sites":[]}`},
		// an empty import list calls neither import nor unimport: use utf8 () leaves <FÉ> a glob, and no utf8 ()
		// leaves use utf8 in force, so perl reads its <FÉ> as readline
		{"emptyutf8readline.pl", `{"ok":true,"prototypes":{},"call_sites":[]}`},
		{"noemptyutf8readline.pl", `{"ok":true,"prototypes":{},"call_sites":[{"kind":"readline","line":3,"end_line":3}]}`},
		// known limit (facts.go utf8At): a utf8 use if imports is not seen, so <FÉ> reports nothing where perl
		// reads it as readline on line 2
		{"ifutf8readline.pl", `{"ok":true,"prototypes":{},"call_sites":[]}`},
		// perl reads <<>> as readline, though parse.IsReadlineTerm calls it a glob
		{"doublediamondreadline.pl", `{"ok":true,"prototypes":{},"call_sites":[{"kind":"readline","line":1,"end_line":1}]}`},
		// a name may open with a digit: perl reads <1> and <$1> as readline
		{"digitreadline.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"readline","line":1,"end_line":1},{"kind":"readline","line":2,"end_line":2}]}`},
		// known limit (facts.go): the file's own r is read before the later import of R's r, so its call is decided
		// on the file's declaration; perl applies the import's \@, reports prototypes {"r":"\\@"} and a \ on line 4
		{"own_then_import.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"line":4,"end_line":4,"name":"r","took_reference":false}]}`},
		// code interpolated in a heredoc body hedges every marker over its statement and the body's lines: perl
		// reports the \ on line 3
		{"heredoc.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"reference","line":2,"end_line":4,"took_reference":false,"unresolved":true},
			{"kind":"hash","line":2,"end_line":4,"unresolved":true},
			{"kind":"match","line":2,"end_line":4,"unresolved":true},
			{"kind":"readline","line":2,"end_line":4,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":4,"unresolved":true}]}`},
		// known limit (facts.go, the interpolated-code shortcut): parse.InterpolatesCode is asked of every quote and
		// body, so a single-quoted string or a <<'E' body holding code hedges every marker where perl -MO=Concise
		// builds no refgen at all
		{"singlequote.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"reference","line":2,"end_line":2,"took_reference":false,"unresolved":true},
			{"kind":"hash","line":2,"end_line":2,"unresolved":true},
			{"kind":"match","line":2,"end_line":2,"unresolved":true},
			{"kind":"readline","line":2,"end_line":2,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":2,"unresolved":true}]}`},
		{"quotedheredoc.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"reference","line":2,"end_line":4,"took_reference":false,"unresolved":true},
			{"kind":"hash","line":2,"end_line":4,"unresolved":true},
			{"kind":"match","line":2,"end_line":4,"unresolved":true},
			{"kind":"readline","line":2,"end_line":4,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":4,"unresolved":true}]}`},
		// #line renumbers what follows it, and a name other than the file's moves it out of the file's sites
		{"line.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"reference","line":2,"end_line":2,"took_reference":true},
			{"kind":"reference","line":10,"end_line":10,"took_reference":true}]}`},
	}
	dir := t.TempDir()
	t.Setenv(subjectEchoEnv, "1")
	subject := parseoracle.Subject{Command: []string{os.Args[0], "-test.run=^TestSubjectEcho$"}}
	for _, s := range samples {
		src, err := os.ReadFile(filepath.Join("testdata", "subject", s.file))
		if err != nil {
			t.Fatal(err)
		}
		out, err := json.Marshal(SubjectFacts(tab, src, s.file, scopeOf(string(src))))
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, s.file+".json")
		if err := os.WriteFile(path, out, 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := subject.Parse(context.Background(), path)
		if err != nil {
			t.Fatalf("%s: parseoracle refuses the facts: %v\n%s", s.file, err, out)
		}
		var want parseoracle.SubjectFacts
		if err := json.Unmarshal([]byte(s.want), &want); err != nil {
			t.Fatalf("%s: the stated facts: %v", s.file, err)
		}
		want.KnowsCallSites, want.KnowsPrototypes = true, true
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: facts\n%s\nwant\n%s", s.file, out, s.want)
		}
	}
}

// TestOracleSubjectHashSites holds the chart's hash sites to the rv2hv ops perl -MO=Concise shows on the pinned perl:
// decided where perl builds one, none where it builds padhv or folds the element into multideref, and a hedge where
// the count turns on a context the chart does not read.
func TestOracleSubjectHashSites(t *testing.T) {
	tab := committedTables(t)
	for _, c := range []struct {
		src             string
		decided, hedged int
	}{
		{`my %h; my $x = $h{a};`, 0, 0},                     // padhv; multideref($h{"a"})
		{`our %h; my $x = $h{a};`, 1, 0},                    // rv2hv vK/OURINTR; multideref($h{"a"})
		{`my %foo; delete @foo{4,5};`, 0, 0},                // padhv twice, under ex-hslice
		{`my $r = {}; my $x = $r->{a}{b}{c};`, 0, 0},        // multideref($r->{"a"}{"b"}{"c"})
		{`my $x = $Foo::h{a}; my @k = keys %Foo::h;`, 1, 0}, // multideref($Foo::h{"a"}); rv2hv over gv(*Foo::h)
		{`%Foo::h = ();`, 1, 0},
		{`use feature 'state'; state %h; my @k = keys %h;`, 0, 0},
		{`my %h; { { my @k = keys %h; } }`, 0, 0},
		{`{ my %h; } my @k = keys %h;`, 1, 0},                            // the my ends with its block
		{`my %h; { our %h; my @k = keys %h; }`, 2, 0},                    // our's alias is the package hash
		{`if ((my %h = ())) { my @k = keys %h } my @j = keys %h;`, 1, 0}, // the my ends with the if
		{`my %h; my $c = sub { keys %h };`, 0, 0},                        // a closure's padhv
		{`use feature 'signatures'; sub f ($x, %h) { return keys %h } my @k = keys %h;`, 1, 0},
		// a %name is a parameter only after ( or , at the signature's own depth; in a default it is the package hash,
		// and so is the body's (Concise: rv2hv in the default and under keys)
		{`use feature "signatures"; no warnings; our %h; my $f = sub ($x = \%h) { my @k = keys %h; };`, 3, 0},
		{`use feature "signatures"; no warnings; our %h; sub f ($x = \%h) { my @k = keys %h; }`, 3, 0},
		{`use feature "signatures"; no warnings; our %h; sub f ($x = scalar(%h)) { my @k = keys %h; }`, 3, 0},
		{`use feature "signatures"; no warnings; our %h; sub f ($x = {%h}) { my @k = keys %h; }`, 3, 0},
		{`use feature "signatures"; no warnings; our %w; our $n = 1; sub f ($x = $n % w) { my @k = keys %w; }`, 2, 0}, // modulus
		{`use feature "signatures"; no warnings; our %h; sub f ($x = (1, %h)) { my @k = keys %h; }`, 3, 0},
		{`use feature "signatures"; no warnings; our %h; sub f ($x = "a,%h") { my @k = keys %h; }`, 2, 0},
		{`use feature "signatures"; no warnings; our %h; sub f ($x = sub (%h) {1}) { my @k = keys %h; }`, 2, 0},
		{`use feature "signatures"; no warnings; our %h; sub f ($x = 1, % h) { my @k = keys %h; }`, 1, 0}, // padhv
		{`use feature "signatures"; no warnings; our %h; sub f ($, %h) { my @k = keys %h; }`, 1, 0},       // a placeholder $
		{`use feature "signatures"; no warnings; our %h; sub f ($x, $, %h) { my @k = keys %h; }`, 1, 0},
		// a placeholder is a bare sigil at a parameter's position only; elsewhere $) and $, are variables (Concise)
		{`use feature "signatures"; no warnings; our %h; sub f ($x = $), %h) { my @k = keys %h; }`, 1, 0},
		{`use feature "signatures"; no warnings; our %h; sub f ($ , %h) { my @k = keys %h; }`, 1, 0},
		{`use feature "signatures"; no warnings; our %h; sub f ($x = $, % h) { my @k = keys %h; }`, 2, 0}, // modulus
		{`our %h; my @v = @h{qw(a b)}; my %w = %h{qw(a b)};`, 3, 0},
		{`our %h; my @x = %h[0];`, 1, 0},                         // %h[...] slices the array @h
		{`our %h; $h{'a', 'b'} = 1; my $x = $h{qw(a b)};`, 3, 0}, // join($;, ...): rv2hv over gv, gvsv(*;)
		// known limit (facts.go elementSite): perl stops folding the rest of the main program into multideref after a
		// range whose constant fold it abandons, so $h{a} keeps an rv2hv (op/range.t:301); the chart reports none
		{`my @x = ("*x".."az"); our %h; my $y = $h{a};`, 1, 0},
		{`my $r = {}; my @v = @$r{qw(a b)}; my %w = %$r; my @k = keys $r->%*;`, 3, 0},
		{`my $r = {}; my $k; my $x = $r->{"x$k"}{a};`, 1, 0},    // rv2hv under helem sKM/DREFHV
		{`our %h; my $k; my $x = $h{a}{"x$k"};`, 1, 0},          // multideref($h{"a"}->) takes the next rv2hv
		{`my $r = {}; my $i; my $x = $r->[$i+1]{"x$i"};`, 1, 0}, // the computed index ends the run
		{`my $r = {}; foo($r->{a}); $r->{a}++;`, 0, 0},
		{`my $r = {}; my $x = \$r->{b};`, 0, 0},                                                  // multideref($r->{"b"}) under srefgen; a refassign is unmeasured, a hedge
		{`my $r = {}; foreach my $x (@{$r->{a}}) {}`, 1, 0},                                      // rv2hv under helem sKM/DREFAV
		{`my $r = {}; my @x = @{$r->{a}};`, 0, 0},                                                // multideref($r->{"a"})
		{`my $r = {}; delete local $r->{a};`, 1, 0},                                              // delete local leaves the helem out of multideref
		{`my $r = {}; push @{$r->{a}}, 1; my @k = keys %{$r->{b}};`, 3, 0},                       // helem sKM/DREFAV; %{} and helem sKM/DREFHV
		{`my $r = {}; my $c = {}; is(scalar(@{$c->{a}}), 1); my $n = scalar(@{$r->{a}});`, 1, 0}, // op_lvalue OP_SCALAR
		{`my $r = {}; my @x = $r->{a}->@{qw(x y)}; my @y = @{$r->{b}}[0, 1];`, 1, 0},             // multideref($r->{"a"}->)
		{`my $r = {}; ${$r->{a}} = 1; ${$r->{a}{b}} = 1;`, 1, 0},                                 // OPpDEREF_SV keeps a lone level
		// a sort block holding <=> or cmp hedges whatever its operands: perl drops { $a <=> $b } and keeps
		// { lc $a cmp lc $b } (Concise), and the chart does not read operands
		{`my $r = {}; my @s = sort { $a <=> $b } @{$r->{a}}; my @t = sort { lc $a cmp lc $b } @{$r->{b}};`, 0, 2},
		{`my $r = {}; my @s = sort { $b cmp $a } @{$r->{a}};`, 0, 1},
		// a refalias of a package hash builds no rv2hv (Concise: refassign over gv(*h), no rv2hv); our's own is one
		{`use feature "refaliasing"; no warnings; our %h; \%h = {}; (\%h) = ({}); for \%h ({}) {} \local %h = {};`, 1, 0},
		{`use feature "refaliasing"; no warnings; our %h; \(my $x, %h) = (\1, {});`, 1, 0},
		{`our %h; my $r = \%h;`, 2, 0},                      // srefgen over rv2hv
		{`our %h; my $x; my @y = ($h{+$x}, $h{$$});`, 1, 0}, // multideref($h{$x}) and multideref($h{$$})
		// a string key is computed where it interpolates: a $ before anything, an @ before a name or one of :'{$+-, and
		// \c takes the next character (toke.c scan_const; Concise: helem over rv2hv); any other quote is an op
		{`our %h; my $x = $h{"$$"};`, 2, 0},
		{`our %h; my $x = $h{"@-"};`, 2, 0},
		{`our %h; my $x = $h{"$&"};`, 2, 0},
		{`our %h; my $x = $h{"$-[0]"}; my $y = $h{"\x{100}$$"}; my $z = $h{qq'$$'};`, 4, 0},
		{`our %h; my $x = $h{"$ a"}; my $y = $h{"$#a"}; my $z = $h{"@'a"}; my $w = $h{"\\c$$"};`, 5, 0},
		{`our %h; my $x = $h{"@!"}; my $y = $h{"@ x"}; my $z = $h{"\c$x"}; my $w = $h{"\$x"};`, 1, 0}, // multideref each
		{"our %h; my @x = ($h{`a`}, $h{qr/a/}, $h{m/a/}, $h{s/a/b/r}, $h{tr/a/b/});", 6, 0},
		{`my $r = {}; my $x = $r->{"$$"}; my $y = $r->{"@-"}; my $z = $r->{"\x{100}$$"};`, 3, 0},
		{`my $r = {}; my $x = $r->{qw($x)};`, 0, 0},            // qw interpolates nothing: multideref($r->{"$x"})
		{`our %h; my $x = $h{qw(@a)};`, 1, 0},                  // multideref($h{"@a"})
		{`our %h; my $x = $h{v1.2}; my $y = $h{1.2.3};`, 1, 0}, // a v-string is a constant: multideref
		{`my $r = {}; my $x = $r->{v1.2};`, 0, 0},
		{`our %h; my $x = $h{qw()};`, 2, 0}, // an empty qw: helem over rv2hv
		// scan_str drops the \ before a $ or @ delimiter, so qq@a\@b@ interpolates @b: such a qq hedges
		{`our %h; my $x = $h{qq@a\@b@}; my $y = $h{qq$a\$b$};`, 1, 2},
		// after whitespace a # starts a comment, not a delimiter, so the quote's body is after it: such a qw or qq hedges
		{"our %h; my $x = $h{qw #c d\n(a)};", 1, 1},
		{"our %h; my $x = $h{qq #$c\n(a)};", 1, 1},
		{"our %h; my $x = $h{qw\f#c d\n(a)};", 1, 1}, // a form feed and a vertical tab are isSPACE too (handy.h)
		{"our %h; my $x = $h{qq\v#$c\n(a)};", 1, 1},
		{"our %h; my $x = $h{qw#a#}; my $y = $h{qq#a#};", 1, 0}, // no whitespace: # is the delimiter
		// +word calls the constant, whose undef peep.c's OP_CONST case will not hold (perl.git perf/opcount.t:218)
		{`use constant u => undef; my $r = {}; my $x = $r->{+u};`, 0, 1},
		// a hash a refalias declares is lexical, and a refalias in a list or under local builds no rv2hv (Concise:
		// padhv, refassign), perl.git op/lvref.t:483
		{`use feature qw(refaliasing declared_refs); no warnings; our @for; for \my %a({5,6},{7,8}) { push @for, %a; }`, 0, 0},
		{`use feature qw(refaliasing declared_refs); no warnings; for my \%a ({}) { my @k = keys %a }`, 0, 0},
		{`use feature qw(refaliasing declared_refs); no warnings; my \%a = {}; my @k = keys %a;`, 0, 0},
		{`use feature qw(refaliasing declared_refs); no warnings; our \%h = {}; \our %g = {};`, 0, 0},
		{`use feature qw(refaliasing declared_refs); no warnings; our \%h = {}; my @k = keys %h; my $r = \our %g;`, 2, 0}, // keys reads the package hash; \our %g is rv2hv
		{`use feature qw(refaliasing declared_refs); no warnings; our %h; our %g; my $x; (\%h, $x) = ({}, 1); ($x, \%g) = (1, {});`, 2, 0},
		{`use feature qw(refaliasing declared_refs); no warnings; our %h; our %g; (\%h, \%g) = ({}, {});`, 2, 0},
		{`use feature qw(refaliasing declared_refs); no warnings; our %h; local \%h = {};`, 1, 0},
		// a trailing ;, parentheses, $::a in main and $Foo::a in Foo hedge too
		{`my $r = {}; my @s = sort { $a <=> $b; } @{$r->{a}}; my @t = sort { ($a <=> $b) } @{$r->{b}};`, 0, 2},
		{`my $r = {}; my @s = sort { $::a <=> $::b } @{$r->{a}}; my @t = sort { $main::b cmp $main::a } @{$r->{b}};`, 0, 2},
		// any block holding <=> or cmp hedges, since constant folding can leave it the one op perl drops
		{`my $r = {}; my $c; my @s = sort { $a <=> $b or $c } @{$r->{a}}; my @t = sort { $a <=> $a } @{$r->{b}};`, 0, 2},
		{`my $r = {}; my @s = sort { +($a <=> $b) } @{$r->{a}};`, 0, 1},
		{`my $r = {}; my @s = sort { 1 ? $a <=> $b : 0 } @{$r->{a}};`, 0, 1},
		{`my $r = {}; my @s = sort { 0 || $a <=> $b } @{$r->{a}};`, 0, 1},
		{`my $r = {}; my @s = sort { $a <=> $b, } @{$r->{a}};`, 0, 1},
		{`my $r = {}; my @s = sort { $a - $b } @{$r->{a}};`, 1, 0}, // no comparison: perl keeps the block (Concise)
		{`my $r = {}; my @s = sort { $b->{x} <=> $a->{x} } @{$r->{a}}; my @t = sort { -$a <=> $b } @{$r->{b}};`, 0, 2},
		{`my $y = [ (my %h = ()) ], my @k = keys %h;`, 1, 0}, // introduced at the statement end, past the brackets
		// other spellings of one comparison of $a and $b hedge too (Concise: each builds 0), as does $::a outside main,
		// which perl keeps (GvSTASH(gv) != PL_curstash)
		{`my $r = {}; my @s = sort { ($a) <=> ($b) } @{$r->{a}};`, 0, 1},
		{`my $r = {}; my @s = sort { ;$a <=> $b } @{$r->{a}};`, 0, 1},
		{`my $r = {}; my @s = sort { ${a} <=> ${b} } @{$r->{a}};`, 0, 1},
		{`my $r = {}; my @s = sort { +$a <=> $b } @{$r->{a}};`, 0, 1},
		{`package Foo; my $r = {}; my @s = sort { $Foo::a <=> $Foo::b } @{$r->{a}};`, 0, 1},
		{`package Foo; my $r = {}; my @s = sort { $::a <=> $::b } @{$r->{a}};`, 0, 1}, // perl keeps it, the chart hedges
		// a lexical $a or $b perl reads as a padsv comparison, which it keeps (Concise: 1 rv2hv each): the hedge covers it
		{`my ($x, ($a)); my $r = {}; my @s = sort { $a <=> $b } @{$r->{a}};`, 0, 1},
		{`CORE::my $a; my $r = {}; my @s = sort { $a <=> $b } @{$r->{a}};`, 0, 1},
		{`package Foo; my $a; my $r = {}; my @s = sort { $::a <=> $::b } @{$r->{a}};`, 0, 1},
		// operands perl drops though their spelling hides it: a newline or comment after the sigil, and $::a in a
		// package whose stash is main's (Concise: 0 rv2hv each)
		{"my $r = {}; my @s = sort { $\na <=> $b } @{$r->{a}};", 0, 1},
		{"my $r = {}; my @s = sort { $ # c\na <=> $b } @{$r->{a}};", 0, 1},
		{`package main::main; my $r = {}; my @s = sort { $::a <=> $::b } @{$r->{a}};`, 0, 1},
		// any other spelling of $a hedges, ' and a doubled main:: included (Concise: each is dropped, 0 rv2hv)
		{`my $r = {}; my @s = sort { $main'a <=> $main'b } @{$r->{a}};`, 0, 1},
		{`package Foo; my $r = {}; my @s = sort { $Foo'a <=> $Foo'b } @{$r->{a}};`, 0, 1},
		{`package Foo; my $r = {}; my @s = sort { $::Foo::a <=> $::Foo::b } @{$r->{a}};`, 0, 1},
		{`my $r = {}; my @s = sort { $::main::a <=> $b } @{$r->{a}};`, 0, 1},
		// \ items of an our (...) list are refaliases; the alias of a foreach is in scope over its body and continue
		// block, not over its own list
		{`use feature qw(refaliasing declared_refs); no warnings; our (\%h) = {}; my @k = keys %h; our (\%g, \%f) = ({}, {});`, 1, 0},
		{`use feature qw(refaliasing declared_refs); no warnings; for my \%a ({}) { 1 } continue { my @k = keys %a }`, 0, 0},
		{`use feature qw(refaliasing declared_refs); no warnings; our %a; for my \%a (\%a) { 1 }`, 2, 0},
		{`use feature qw(refaliasing declared_refs); no warnings; our %a; for \my %a ({%a}) { 1 }`, 2, 0},
		// shortcut row: a file calling export_lexically may have made %h lexical (Concise: padhv), so its reads hedge
		{`BEGIN { builtin::export_lexically("%h" => {}) } my @k = keys %h;`, 0, 1},
		{`(my %h = ()), my @k = keys %h;`, 1, 0}, // the my is introduced at the statement's end: keys reads the package's
	} {
		got := SubjectFacts(tab, []byte(c.src), "h.pl", Scope{})
		if !got.OK {
			t.Errorf("%s: refused", c.src)
			continue
		}
		var decided, hedged int
		for _, s := range got.CallSites {
			switch {
			case s.Kind != parseoracle.SiteKindHash:
			case s.Unresolved:
				hedged++
			default:
				decided++
			}
		}
		if decided != c.decided || hedged != c.hedged {
			t.Errorf("%s: %d decided and %d hedged hash sites, want %d and %d", c.src, decided, hedged, c.decided, c.hedged)
		}
	}
}

// subjectEchoEnv makes TestSubjectEcho the subject TestOracleSubject runs: this test binary, printing a file.
const subjectEchoEnv = "CHART_SUBJECT_ECHO"

// TestSubjectEcho is no test of its own: run as TestOracleSubject's subject command, it prints the file named last
// on its command line, the facts TestOracleSubject wrote, and exits before the testing package prints anything.
func TestSubjectEcho(t *testing.T) {
	if os.Getenv(subjectEchoEnv) == "" {
		t.Skip("TestOracleSubject's subject command")
	}
	b, err := os.ReadFile(os.Args[len(os.Args)-1])
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
	os.Stdout.Write(b)
	os.Exit(0)
}

// TestSubjectFactsRefusedAndEmpty: a file the chart refuses answers ok false, an empty one ok true, each with both
// lists empty and never null, which parseoracle would read as a question not asked.
func TestSubjectFactsRefusedAndEmpty(t *testing.T) {
	tab := committedTables(t)
	for _, c := range []struct{ src, want string }{
		{"1 +;", `{"ok":false,"prototypes":{},"call_sites":[]}`}, // perl -c: syntax error
		{"", `{"ok":true,"prototypes":{},"call_sites":[]}`},
	} {
		out, err := json.Marshal(SubjectFacts(tab, []byte(c.src), "f.pl", Scope{}))
		if err != nil {
			t.Fatal(err)
		}
		if string(out) != c.want {
			t.Errorf("%q: %s, want %s", c.src, out, c.want)
		}
	}
}
