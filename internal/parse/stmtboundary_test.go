// ABOUTME: Canon writes a statement's boundaries as the source did: a deref's braces, a `;` after an
// ABOUTME: anonymous sub or a bare block's continue, a file's last `;`, and a heredoc body before its closers.

package parse_test

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/perltest"
)

// derefBraces are block dereferences written with braces around a plain
// scalar, reduced from op/avhv.t, op/gv.t, uni/gv.t, op/kvaslice.t,
// op/kvhslice.t, op/localref.t, op/postfixderef.t, op/ref.t and uni/fold.t.
// perl keeps the braces in its tree: measured on 5.44.0, `my @x = @{$r}`
// builds a nulled `scope` and `ex-nextstate` around the `padsv` that `my @x
// = @$r` builds bare, and Deparse,-p writes `@{$r;}` for one and `@$r` for
// the other. The ops that run are the same, so canon writing the source's
// spelling changes no meaning.
var derefBraces = []string{
	"my $avhv; my @x = delete @{$avhv}{'foo','pants'};\n",
	"my $a; ok(!defined ${$a});\n",
	"my $a; my $s = join(':', %{$a}[2,3,4]);\n",
	"my $h; my $s = join(':', %{$h}{'c','d','e'});\n",
	"my $x; { local ${$x}; }\n",
	"my $h; is(${$h}{ppp}, 'pp');\n",
	"my $name1; my (undef, $one) = @{$name1}[2,3];\n",
	"my @a = @{$_}[1,2];\n",
	"my $h; my $x = ${$h}{a}{b};\n",
	"my $h; my @x = @{$h}{qw(a b)};\n",
	"my $r; my @x = @{$r};\n",
	"my $r; my $n = $#{$r};\n",
	"my $c; &{$c}(1);\n",
	"my $g; *{$g} = sub { 1; };\n",
	// A scalar dereference as the operand of another, op/gv.t's and
	// uni/gv.t's `$${$_[0]}`: Deparse,-p writes it `$${$_[0];}`, bare.
	"sub FETCH { $${$_[0]}; }\n",
	"my $x; my $v = $${$x};\n",
}

// anonSubStatements are anonymous subs standing as statements, reduced from
// op/closure.t, op/const-optree.t, op/stash.t and uni/stash.t. An anonymous
// sub is an expression, so its statement takes a `;`: measured on 5.44.0,
// `sub {1} print 1` is a syntax error, and Deparse,-p writes `sub f { sub {
// 1 }; 2 }` with the `;` after the anonymous sub's brace.
var anonSubStatements = []string{
	"sub bizz { if (@_) { my $i = shift; sub { $i = shift if @_; $i; }; } }\n",
	"my $s = sub { sub () { $x; }; };\n",
	"my $sub = do { package four; sub { 1; }; };\n",
	"sub { 1; };\nprint 2;\n",
}

// gotoKeywordOperand is `goto` before a word perl reads as a keyword, reduced
// from op/goto-sub.t, op/sort.t and op/state.t, and a loop control's
// expression operand, from op/loopctl.t. toke.c's LOOPX reads the
// word after goto, last, next, redo and dump as a label only when it is not a
// keyword (force_word with check_keyword), so the keyword begins goto's
// operand, a term: measured on 5.44.0, Deparse,-p writes `goto sub {...}` as
// `(goto (sub {...}))`, `goto shift` as `(goto (shift(@ARGV)))`, `goto FOO`
// as `(goto FOO)`, and `goto state $flower = $f` as `(goto (state $flower =
// $f))`: the operand runs through an assignment, `goto $x = 1` is `(goto ($x
// = 1))`, and stops at a comma, `goto $x, 1` is `((goto $x), '???')`.
var gotoKeywordOperand = []string{
	"sub f1 { my $x; goto sub { $x = 0; ok(1); }; }\n",
	"my @output = sort { goto sub { 1; }; } 1, 2;\n",
	"use feature 'state'; foreach my $f (1) { goto state $flower = $f; }\n",
	"sub f2 { my $x; goto $x = 1; }\n",
	// last, next and redo are LOOPX too, op/loopctl.t's `last $label`:
	// Deparse,-p writes `eval { last $l }` as `eval { do { (last $l) } }`,
	// and `last shift` as `(last (shift(@ARGV)))`.
	"L: { my $l = 'L'; eval { last $l; }; }\n",
	"for my $x (1) { next $x . 'a'; redo $x if 0; last shift; }\n",
	// A modifier word begins no operand, `when` among them (op/switch.t).
	"use feature 'switch'; no warnings; for (1) { next when 2; last if 1; }\n",
}

// bareBlockContinue is a bare block with a continue block, then the next
// statement with no `;` between, reduced from op/goto.t, op/while.t and
// op/loopctl.t. It is a loop, a block form: measured on 5.44.0, `{ 1 }
// continue { 2 } 3;` compiles, and Deparse,-p writes `{ goto A; A: $ok = 1 }
// continue { } ok($ok);` with no `;` after the continue block.
var bareBlockContinue = []string{
	"my $ok; { goto A; A: $ok = 1; } continue { } ok($ok);\n",
	"my ($got_l, $l); { my $l = 0; } continue { $got_l = $l; } is($got_l, 18);\n",
}

// fileEnds are files whose last statement has no `;`, reduced from
// lib/universal.t. The main program's last statement needs none, as a
// block's does not: measured on 5.44.0, Deparse,-p writes `print 1; print 2`
// as `print(1); print(2);`, the same as with the `;`.
// loopControlWords are the words after a loop control that its review
// (2026-10-25) found read wrongly once a keyword began the operand. A word
// operator applies to the loop control: measured on perl 5.44.0, `last eq
// 1` is `((last) eq 1)` and `last and print "x"` prints nothing. `x` names a
// label (Concise: goto("x"), op/goto.t:755), and so does a switch keyword
// with the switch feature off: `last default` is a syntax error under `use
// v5.10` and names a label under `use v5.36`.
var loopControlWords = []string{
	"for (1..2) { last and print \"x\\n\"; } print \"ok\\n\";\n",
	"for (1) { last eq 1; }\n",
	"for (1) { last or 2; }\n",
	"for (1) { redo xor 1; }\n",
	"for (1) { next lt 1; }\n",
	"sub f { goto or 1; }\n",
	"goto x; x: print \"x\\n\";\n",
	"for (1) { last default; }\n",
	"use v5.36; for (1) { last given; }\n",
	// The second review of 01a16602-6030 (2026-10-25): an operator tighter
	// than the comma after a label continues the operand, with the label
	// forced to a string: `last L ne 1` is `(last 1)` and `goto x x 1` is
	// `(goto x)`. A modifier applies over a word operator over a loop
	// control, and a later `use v5.36` turns the switch feature off.
	"X: for (1) { goto x x 1; } print \"ok\\n\";\n",
	"L: for (1) { last L x 3; }\n",
	"L: for (1) { last L ne 1; }\n",
	"for (1) { last or print \"x\\n\" if 1; } print \"ok\\n\";\n",
	"use v5.10; use v5.36; for (1) { last default; }\n",
	// The third review: `->` after a label is a method call on the string
	// (`(last 'L'->foo)`), and a modifier applies over a chain of word
	// operators over a loop control.
	"L: for (1) { last L->foo; } print \"ok\\n\"; sub L::foo { 'L'; }\n",
	"for (1) { last or print 1 or print 2 if 1; } print \"ok\\n\";\n",
}

// notDataSection are `__END__` and `__DATA__` spellings perl reads as code,
// found by the same review: measured on perl 5.44.0, `__END__::foo()`
// calls __END__::foo, `__END__ => 1` quotes the word and `__END__:` is a
// label, with or without a blank before the `:`, and a marker after a
// heredoc's opener leaves the body to the heredoc.
var notDataSection = []string{
	"print \"a\\n\"; __END__::foo(\\1) if 0; print \"b\\n\";\n",
	"print \"a\\n\"; my %h = (__END__ => 1); print \"b\\n\";\n",
	"print \"a\\n\"; __END__: print \"b\\n\";\n",
	"print \"a\\n\"; __END__ : print \"b\\n\";\n",
}

// midLineDataSection is a data section whose marker does not begin its
// line. toke.c ends the program at the keyword wherever a statement starts
// with it: measured on perl 5.44.0, `print "a\n"; __END__ not code` prints
// "a", and Deparse shows the section.
var midLineDataSection = []string{
	"print \"a\\n\"; __END__\nnot code\n",
	"print \"a\\n\"; __DATA__\nd\n",
}

// TestLoopControlWordLabels: a word that names a label after a loop
// control is a Label, not a call -- Faithful cannot see the difference, the
// tree can. Under the switch feature a switch keyword is the operand.
func TestLoopControlWordLabels(t *testing.T) {
	for src, want := range map[string]parse.Kind{
		"goto x; x: 1;\n":                       parse.Label,
		"for (1) { last default; }\n":           parse.Label,
		"for (1) { last given; }\n":             parse.Label,
		"use v5.10; for (1) { last given; }\n":  parse.Call,
		"use feature 'switch'; last default;\n": parse.Call,
	} {
		lc := firstOfKind(parse.Parse([]byte(src)), parse.LoopControl)
		if lc == nil || len(lc.Children) != 1 || lc.Children[0].Kind != want {
			t.Errorf("%q: want the loop control's one child a %v, got %+v", src, want, lc)
		}
	}
}

var fileEnds = []string{
	"print 1;\nprint 2\n",
	"print 1;\nprint 2 # a comment\n",
	"print <<A\nbody\nA\n",
}

// heredocClosers are heredoc bodies the source put before closers that end
// the statement's call, 5fd5's review follow-up F3. perl reads the body from
// the line after its opener's and then resumes, so the closers follow it:
// measured on 5.44.0, `print f(<<A\nx\nA\n), "\n";` is `print(f("x\n"),
// "\n");` under Deparse,-p.
var heredocClosers = []string{
	"print f(<<A\nx\nA\n), \"\\n\";\n",
	"print(f(g(<<A\nx\nA\n)), 1);\n",
	"my $x = [<<A\nx\nA\n];\n",
}

// TestFaithfulStatementBoundaries: each reduced statement-boundary case
// re-emits faithfully with nothing forgiven, and perl reads canon as it reads
// the source.
func TestFaithfulStatementBoundaries(t *testing.T) {
	var all []string
	for _, group := range [][]string{derefBraces, anonSubStatements, gotoKeywordOperand, bareBlockContinue, fileEnds, heredocClosers, loopControlWords, midLineDataSection, notDataSection} {
		all = append(all, group...)
	}
	for _, src := range all {
		faithfulUnforgiven(t, src)
	}
	t.Run("perl", func(t *testing.T) {
		deparsesAlike(t, perltest.PinnedPerl(t), all)
	})
}

// derefBraceShapes are the brace-form dereferences whose flag must change
// nothing Sites reads: plain operands, a non-plain one whose braces were
// always written, `@{$r->{k}}` in the contexts that keep rv2hv, and the
// subscripted forms 30a8's review asked for.
var derefBraceShapes = []string{
	"my $x; my @y = @{$x};\n",
	"my $a; my @k = keys %{$a};\n",
	"my $a; my $v = ${$a->[0]};\n",
	"my $r; push @{$r->{k}}, 1;\n",
	"my $r; foo(@{$r->{k}});\n",
	"my $r; my @x = @{$r->{k}};\n",
	"my $h; my $x = ${$h}{a}{b};\n",
	"my $h; my @x = @{$h}{qw(a b)};\n",
	"my $h; my $x = ${$h}{a};\n",
	"our %h; push @{$h{a}}, 1;\n",
	"my $a; my %c = %{$a}{'c','d'};\n",
}

// clearDerefBraces clears every DerefBraces flag under n.
func clearDerefBraces(n *parse.Node) {
	n.DerefBraces = false
	for _, c := range n.Children {
		clearDerefBraces(c)
	}
}

// countDerefBraces counts the nodes under n with DerefBraces set.
func countDerefBraces(n *parse.Node) int {
	count := 0
	if n.DerefBraces {
		count++
	}
	for _, c := range n.Children {
		count += countDerefBraces(c)
	}
	return count
}

// TestDerefBraceFlagLeavesSitesUnchanged: sites.go decides rv2hv from node
// Kind and Text, so the deref-brace flag must not move a site. Over every
// probe snippet and the brace-form shapes, Sites on the parsed tree equals
// Sites on the same tree with the flag cleared.
func TestDerefBraceFlagLeavesSitesUnchanged(t *testing.T) {
	probes, err := loadProbes(m3ProbeDir)
	if err != nil {
		t.Fatal(err)
	}
	srcs := append([]string(nil), derefBraceShapes...)
	for _, p := range probes {
		srcs = append(srcs, string(p.src))
	}
	flagged := 0
	for _, src := range srcs {
		root := parse.Parse([]byte(src))
		with := parse.Sites(root, []byte(src))
		flagged += countDerefBraces(root)
		clearDerefBraces(root)
		if without := parse.Sites(root, []byte(src)); !reflect.DeepEqual(with, without) {
			t.Errorf("%q: Sites with the deref-brace flag\n%+v\nwithout it\n%+v", src, with, without)
		}
	}
	// The flag is set somewhere, or the comparison above compares nothing.
	if flagged < len(derefBraceShapes) {
		t.Errorf("the deref-brace flag is set on %d node(s) across %d snippets, want at least %d", flagged, len(srcs), len(derefBraceShapes))
	}
}

// bracelessTree is n with every span cleared, so trees of two spellings
// compare by shape and flags alone.
func bracelessTree(n *parse.Node) *parse.Node {
	cp := *n
	cp.Start, cp.End, cp.InnerStart, cp.InnerEnd = 0, 0, 0, 0
	cp.Children = nil
	for _, c := range n.Children {
		cp.Children = append(cp.Children, bracelessTree(c))
	}
	return &cp
}

// TestDerefBraceFlagClearedEqualsBracelessTree: the flag is only a flag.
// With it cleared, the tree of each brace form equals the tree of its
// braceless spelling, spans aside, so the flag cannot hide a shape change.
// With it set, the two differ, so the comparison can see a difference.
func TestDerefBraceFlagClearedEqualsBracelessTree(t *testing.T) {
	for _, pair := range [][2]string{
		{"my $x; my @y = @{$x};\n", "my $x; my @y = @$x;\n"},
		{"my $a; my @k = keys %{$a};\n", "my $a; my @k = keys %$a;\n"},
		{"my $a; my $v = ${$a};\n", "my $a; my $v = $$a;\n"},
		{"my $h; my $x = ${$h}{a}{b};\n", "my $h; my $x = $$h{a}{b};\n"},
		{"my $h; my @x = @{$h}{qw(a b)};\n", "my $h; my @x = @$h{qw(a b)};\n"},
		{"my $a; my %c = %{$a}[2,3];\n", "my $a; my %c = %$a[2,3];\n"},
		{"my $x; { local ${$x}; }\n", "my $x; { local $$x; }\n"},
	} {
		braced := parse.Parse([]byte(pair[0]))
		bare := bracelessTree(parse.Parse([]byte(pair[1])))
		if reflect.DeepEqual(bracelessTree(braced), bare) {
			t.Errorf("%q: the brace form's tree, flag set, equals %q's", pair[0], pair[1])
		}
		clearDerefBraces(braced)
		if !reflect.DeepEqual(bracelessTree(braced), bare) {
			t.Errorf("%q with the flag cleared is not %q's tree", pair[0], pair[1])
		}
	}
}

// TestFaithfulBlockDerefBracesKeptWhenGroupingChanges: canon writes a block
// dereference's braces where the source had them, and their absence is
// never forgiven. Measured on perl 5.44.0, Deparse,-p writes `${$a->[0]}` as
// `${$a->[0];}` and `$$a->[0]` as `$$a->[0]`, a different tree: the first
// dereferences the element, the second indexes the array $$a refers to.
func TestFaithfulBlockDerefBracesKeptWhenGroupingChanges(t *testing.T) {
	for _, src := range []string{"my $a; my $v = ${$a->[0]};\n", "my $x; my @y = @{$x};\n"} {
		faithfulUnforgiven(t, src)
	}
	rejects(t, "my $a; my $v = ${$a->[0]};", "my $a; my $v = $$a->[0];")
	// A plain operand's braces too: the flag is how canon writes them, so
	// a canon without them is a difference.
	if ok, _, _ := parse.FaithfulTokens(sigTokens("my @y = @{$x};"), asCanon(sigTokens("my @y = @$x;"))); ok {
		t.Errorf("source `@{$x}` against canon `@$x` was forgiven")
	}
}

// TestFaithfulDoesNotForgiveMissingInnerSemicolon: a `;` between two
// statements is compared like any token. Last-semicolon forgives only a
// marked `;` right before a `}` the source went straight to, so a `;`
// missing between statements is a difference even with every `;` marked.
func TestFaithfulDoesNotForgiveMissingInnerSemicolon(t *testing.T) {
	canon := asCanon(sigTokens("{$a = 1 $b = 2;}"))
	for i := range canon {
		canon[i].BlockSemi = canon[i].Text == ";"
	}
	if ok, forgiven, _ := parse.FaithfulTokens(sigTokens("{ $a = 1; $b = 2 }"), canon); ok {
		t.Errorf("a missing inner `;` was forgiven by %v", ruleNames(forgiven))
	}
	// The other way: canon writes a `;` the source has not, inside the block.
	canon = asCanon(sigTokens("{$a = 1;; $b = 2;}"))
	for i := range canon {
		canon[i].BlockSemi = canon[i].Text == ";"
	}
	if ok, forgiven, _ := parse.FaithfulTokens(sigTokens("{ $a = 1; $b = 2 }"), canon); ok {
		t.Errorf("an extra inner `;` was forgiven by %v", ruleNames(forgiven))
	}
}

// TestFaithfulDetectsStatementMovedAcrossBrace: the same tokens with a `}`
// moved, so a statement goes into or out of a block, is a difference.
// Measured on perl 5.44.0, `{ my $a = 1; } print $a;` prints nothing and
// `{ my $a = 1; print $a; }` prints 1.
func TestFaithfulDetectsStatementMovedAcrossBrace(t *testing.T) {
	rejects(t, "{ $a = 1; } $b = 2;", "{$a = 1;$b = 2;}")
	rejects(t, "{$a = 1;$b = 2;}", "{$a = 1;} $b = 2;")
}

// TestFaithfulTokenCountMismatchIsADifference: a canon that ends early, or
// runs on, is a difference; the longer side's tail is never skipped.
// lib/universal.t's 318 against 369 was the file's real length against a
// stale count: the file is 75 lines, and its last statement has no `;`.
func TestFaithfulTokenCountMismatchIsADifference(t *testing.T) {
	for _, c := range []struct{ src, canon string }{
		{"f(1); g(2);", "f(1);"},
		{"f(1);", "f(1); g(2);"},
		{"is $], 7, 'x'", "is $], 7, 'x';"},
		{"is $], 7, 'x';", "is $], 7, 'x'"},
	} {
		canon := asCanon(sigTokens(c.canon))
		// A marked `;` at the end, with no `}` after it, is still extra.
		for i := range canon {
			canon[i].BlockSemi = canon[i].Text == ";"
		}
		if ok, forgiven, _ := parse.FaithfulTokens(sigTokens(c.src), canon); ok {
			t.Errorf("%q against canon %q: forgiven by %v", c.src, c.canon, ruleNames(forgiven))
		}
	}
}
