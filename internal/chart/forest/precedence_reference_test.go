// ABOUTME: Compares production Constrain against ConstrainReference, verdict and kept set, on two copies of one forest.
// ABOUTME: Covers the toy fixture forests and decision §2's seeded generators (seed 4 operators, seed 44 chains).

package forest

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/prectab"
	"tamarou.com/pvm/internal/chart/recognize"
)

// keptAlts is node n's alternatives, in order.
func keptAlts(f *Forest, n int) []int32 {
	var out []int32
	for a := f.Nodes[n].Alt; a != 0; a = f.Alts[a].Next {
		out = append(out, a)
	}
	return out
}

// altCount is how many alternatives f's nodes hold.
func altCount(f *Forest) int {
	k := 0
	for _, nd := range f.Nodes {
		for a := nd.Alt; a != 0; a = f.Alts[a].Next {
			k++
		}
	}
	return k
}

// precedenceDiff is where production's judgement of one copy of a forest, prodOK and prod, differs from the
// reference's of another copy, refOK and ref: the verdict, or the first node whose kept alternatives differ; ""
// where the two agree.
func precedenceDiff(prodOK bool, prod *Forest, refOK bool, ref *Forest) string {
	if prodOK != refOK {
		return fmt.Sprintf("node %d (root): production verdict %v, reference %v", prod.Root, prodOK, refOK)
	}
	if len(prod.Nodes) != len(ref.Nodes) {
		return fmt.Sprintf("production has %d nodes, reference %d", len(prod.Nodes), len(ref.Nodes))
	}
	for n := range prod.Nodes {
		if p, r := keptAlts(prod, n), keptAlts(ref, n); !slices.Equal(p, r) {
			return fmt.Sprintf("node %d: production keeps alternatives %v, reference %v", n, p, r)
		}
	}
	return ""
}

// judgeBoth makes two copies of f before either is judged, runs Constrain on one and ConstrainReference on the
// other, and is where they differ ("" none), the reference's verdict, and whether it accepted with an alternative
// dropped.
func judgeBoth(f *Forest, p *prectab.Precedence) (diff string, ok, pruned bool) {
	prod, ref := f.Copy(), f.Copy()
	prodOK := productionConstrain(prod, p)
	refOK := ref.ConstrainReference(p)
	return precedenceDiff(prodOK, prod, refOK, ref), refOK, refOK && prunedReachable(f, ref)
}

// productionConstrain is the judgement judgeBoth compares with the reference; TestPrecedenceHarnessesDoNotAlias
// swaps it for one that over-prunes.
var productionConstrain = (*Forest).Constrain

// prunedReachable reports whether judged, a judged copy of f, holds fewer alternatives than f on a node reachable
// from f's root: both judgements drop every alternative of an unreachable node, which proves nothing.
func prunedReachable(f, judged *Forest) bool {
	seen := make([]bool, len(f.Nodes))
	work := []int32{f.Root}
	for len(work) > 0 {
		n := work[len(work)-1]
		work = work[:len(work)-1]
		if seen[n] {
			continue
		}
		seen[n] = true
		if len(keptAlts(judged, int(n))) < len(keptAlts(f, int(n))) {
			return true
		}
		for a := f.Nodes[n].Alt; a != 0; a = f.Alts[a].Next {
			for _, c := range []int32{f.Alts[a].Left, f.Alts[a].Right} {
				if c != 0 {
					work = append(work, c)
				}
			}
		}
	}
	return false
}

// TestPrecedenceDiffDetectsDifference: the comparison reports a node that lost one alternative on one side, a verdict
// that differs alone, and a copy Constrain judged against one it did not; and nothing where both sides agree. The
// sentence is one the reference accepts and prunes, so a broken Constrain cannot choose it.
func TestPrecedenceDiffDetectsDifference(t *testing.T) {
	ty := loadToy(t, "ladder")
	var f, judged *Forest
	for _, s := range ty.sentences {
		if g, _, _ := toyForest(t, ty.tab, s.words, recognize.Options{}); g != nil {
			if j := g.Copy(); j.ConstrainReference(ty.prec) && altCount(j) < altCount(g) {
				f, judged = g, j
				break
			}
		}
	}
	if f == nil {
		t.Fatal("no ladder sentence the reference accepts and prunes")
	}
	if d := precedenceDiff(true, judged, true, judged.Copy()); d != "" {
		t.Errorf("two equal copies: %s, want no difference", d)
	}

	dropped := judged.Copy()
	n := slices.IndexFunc(dropped.Nodes, func(nd Node) bool { return nd.Alt != 0 })
	dropped.Nodes[n].Alt = dropped.Alts[dropped.Nodes[n].Alt].Next
	if d := precedenceDiff(true, dropped, true, judged); !strings.Contains(d, fmt.Sprintf("node %d:", n)) {
		t.Errorf("one alternative dropped on node %d: %q, want that node reported", n, d)
	}

	if d := precedenceDiff(false, judged, true, judged.Copy()); !strings.Contains(d, fmt.Sprintf("node %d (root)", judged.Root)) {
		t.Errorf("verdict alone differing: %q, want the root reported", d)
	}

	prod := f.Copy()
	if d := precedenceDiff(prod.Constrain(ty.prec), prod, true, f.Copy()); d == "" {
		t.Error("a copy Constrain judged against an unjudged one: no difference reported")
	}
}

// TestPrecedenceToyForestsMatchReference requires production and the reference to keep the same alternatives, and
// reach the same verdict, on every toy fixture sentence's forest.
func TestPrecedenceToyForestsMatchReference(t *testing.T) {
	for _, name := range toyNames {
		ty := loadToy(t, name)
		bad := 0
		for _, s := range ty.sentences {
			f, _, _ := toyForest(t, ty.tab, s.words, recognize.Options{})
			if f == nil {
				continue
			}
			if diff, _, _ := judgeBoth(f, ty.prec); diff != "" {
				if bad++; bad <= 10 {
					t.Errorf("%s %s: %s", name, strings.Join(s.words, " "), diff)
				}
			}
		}
	}
}

// sentencesMatchReference runs each source as ParseWith does up to the constraint (lex, recognition, Build,
// LibraryOperators under the empty scope) and compares production and the reference on that forest. It logs the
// counts and requires a sentence compared, one refused and one pruned, so a generator cannot agree vacuously.
func sentencesMatchReference(t *testing.T, srcs []string) {
	tab := committedTables(t)
	prec, err := prectab.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	r := recognize.New(tab, recognize.Options{})
	compared, refused, pruned, bad := 0, 0, 0, 0
	for _, src := range srcs {
		s, _ := lex.Lex([]byte(src), tab.Tokens)
		l := lex.Linear(s)
		c, err := recognize.Recognize(r, l)
		if err != nil {
			t.Fatalf("%q: %v", src, err)
		}
		if !c.Accepted() {
			continue
		}
		f := Build(tab, c, l)
		if LibraryOperators(tab, f, s, []byte(src), Scope{}) >= 0 {
			continue
		}
		compared++
		diff, ok, p := judgeBoth(f, prec)
		if diff != "" {
			if bad++; bad <= 10 {
				t.Errorf("%q: %s", src, diff)
			}
		}
		if !ok {
			refused++
		}
		if p {
			pruned++
		}
	}
	t.Logf("sentences=%d compared=%d refused=%d pruned=%d", len(srcs), compared, refused, pruned)
	if bad > 0 {
		t.Errorf("%d of %d compared sentences differ from the reference", bad, compared)
	}
	if compared == 0 || refused == 0 || pruned == 0 {
		t.Errorf("compared %d, refused %d, pruned %d; want each above 0", compared, refused, pruned)
	}
}

// operatorCrafted are the 83 handcrafted sentences of the seed-4 harness (the assessment's harness/fresh4).
var operatorCrafted = []string{
	"1 == 2 == 3;", "1 < 2 < 3;", "1 .. 2 .. 3;", "-e -f $x;", "my $x [0];", "my $x {1};", "my ($x) [0];",
	"print $fh [1,2];", "print {$fh} [1];", "print $fh {1};", "print STDOUT [1];", "sort $f [1];", "sort $f {1};",
	"local $x [0];", "our @a [0];", "my %h {1};", "$x = my $y [0];", "f(my $x [0]);", "{ my $x }", "{ my $x [0] }",
	"{ print $fh [1] }", "{ $a ** $b }", "{ $a ** $b ** $c }", "{ 1 == 2 }", "{ 1 == 2 == 3 }", "{ -$x ** 2 }",
	"{ !$x }", "{ $x = $y ? $z : $w }", "{ not $x }", "{ defined $x }", "{ ref $x }", "{ $x++ }", "{ my $x {1} }",
	"$x = { my $y };", "$x = sub { $a <=> $b };", "sort { $a <=> $b } @x", "$x ? $y : $z ? $w : $v;", "$x = $y = $z;",
	"$x->[0]{1}[2];", "$x->${\\ 'y'};", "defined $x ? 1 : 0;", "ref $x eq 'HASH';", "-$x ** 2;", "!$x == 1;",
	"\\$x->[0];", "$x x 3 x 4;", "$x =~ s/a/b/r =~ tr/a/b/r;", "print { $fh } $x, $y;", "print $fh $x, $y;",
	"return my $x [0];", "my $x = 1 if $y;", "do { 1 } while $x;", "$x = do { 1 } + 2;", "$x = (1) [0];",
	"$x = (1,2) [0];", "$x = (my $y) [0];", "@x = (sort $f) [0];", "$x = { 1 => 2 } -> {1};", "$x lt $y lt $z;",
	"$x <=> $y <=> $z;", "$x =~ $y =~ $z;", "-f -d $x;", "stat _ -f;", "$x->@*;", "my $x; my $y", "my $x\n}",
	"exit 1 }", "$x **", "", ";", ";;;", "{}", "{;}", "sub f { }", "sub f { my $x }", "sub f { $a ** $b }",
	"sub f { return $a ** $b }", "if ($x) { $a == $b }", "if ($x) { $a == $b == $c }", "$x = [ my $y [0] ];",
	"$x = { my $y {0} };", "@x = map { my $y [0] } @z;", "@x = grep { $_ == 1 == 2 } @z;",
}

// TestPrecedenceOperatorSentencesMatchReference is decision §2's seed-4 generator, production against the reference
// on each sentence.
func TestPrecedenceOperatorSentencesMatchReference(t *testing.T) {
	sentencesMatchReference(t, operatorSentences())
}

// operatorSentences is decision §2's seed-4 generator: the 83 handcrafted sentences and 20,000 random operator
// sentences.
func operatorSentences() []string {
	terms := []string{"$x", "$y", "1", "@a", "%h", "f()", "(1)", "[1]", "{1}", "\"s\"", "$h{1}", "$a[0]", "$r->[0]", "my $m", "my ($p)", "local $l", "our $o", "_", "STDOUT", "$fh", "{$fh}"}
	pre := []string{"-", "!", "\\", "~", "+", "not ", "defined ", "ref ", "-e ", "-f ", "++", "--", "scalar ", "print ", "print $fh ", "print {$fh} ", "sort ", "sort $s ", "return ", "my ", "lc ", "exists ", "delete ", "rand ", "-M "}
	bin := []string{"+", "-", "*", "/", "**", "==", "!=", "<=>", "<", ">", "lt", "eq", "cmp", "&&", "||", "//", "and", "or", "xor", "=", "+=", "**=", "..", "...", "=~", "!~", "x", ".", ",", "=>", "?", ":", "->", "&", "|", "^", "<<", ">>", "isa"}
	post := []string{"++", "--", "[0]", "{1}", "->[0]", "->{1}", "->()", "(1)", " [0]", " {1}"}
	wraps := []string{"%s;", "{ %s }", "$z = %s;", "f(%s);", "if (%s) { 1 }", "sub { %s }", "[ %s ]", "%s"}
	rng := rand.New(rand.NewSource(4))
	var gen func(d int) string
	gen = func(d int) string {
		var b strings.Builder
		for i := rng.Intn(2); i > 0; i-- {
			b.WriteString(pre[rng.Intn(len(pre))])
		}
		if d > 0 && rng.Intn(4) == 0 {
			b.WriteString("(" + gen(d-1) + ")")
		} else {
			b.WriteString(terms[rng.Intn(len(terms))])
		}
		if rng.Intn(3) == 0 {
			b.WriteString(post[rng.Intn(len(post))])
		}
		if d > 0 && rng.Intn(3) > 0 {
			b.WriteString(" " + bin[rng.Intn(len(bin))] + " " + gen(d-1))
		}
		return b.String()
	}
	srcs := slices.Clone(operatorCrafted)
	for range 20000 {
		srcs = append(srcs, fmt.Sprintf(wraps[rng.Intn(len(wraps))], gen(3)))
	}
	return srcs
}

// TestPrecedenceChainSentencesMatchReference is decision §2's seed-44 generator, production against the reference on
// each sentence.
func TestPrecedenceChainSentencesMatchReference(t *testing.T) {
	sentencesMatchReference(t, chainSentences())
}

// chainSentences is decision §2's seed-44 generator: 15,000 Leo chains of 1 to 25 links, some wrapped in a nullable
// context.
func chainSentences() []string {
	links := []string{"$a = ", "$a ||= ", "$a **= ", "not ", "!", "-", "\\", "~", "defined ", "ref ", "print ", "return ", "$a ? $b : ", "$a ** ", "$a .. ", "$a x ", "$a // ", "$a || ", "$a or ", "$a && ", "$a and ", "$a <=> ", "$a == ", "$a . ", "$a + ", "my $m = ", "local $l = ", "sort ", "lc ", "-e ", "f ", "$a, ", "$a => ", "($a) x ", "scalar "}
	tails := []string{"$z", "$z ** $y", "$z + $y * $w", "$z ? $y : $w", "$z = $y", "$z->[0]", "$z++", "--$z", "@z", "()", "[]", "{}", "f()", "$z <=> $y", "$z .. $y", "$z [0]", "$fh [1]", "$fh []", "{$fh} []", "$z or $y", "1"}
	nulls := []string{"for (;;) { %s }", "%s for ();", "sub f :lvalue { %s }", "my $x = () = (%s);", "@x = (%s,);", "do {} until %s;", "{ ; %s ; }", "return;", "print;", "%s"}
	rng := rand.New(rand.NewSource(44))
	var srcs []string
	for range 15000 {
		var b strings.Builder
		for k := 1 + rng.Intn(25); k > 0; k-- {
			b.WriteString(links[rng.Intn(len(links))])
		}
		b.WriteString(tails[rng.Intn(len(tails))])
		s := b.String()
		if w := nulls[rng.Intn(len(nulls))]; strings.Contains(w, "%s") {
			s = fmt.Sprintf(w, s)
		}
		if rng.Intn(2) == 0 {
			s += ";"
		}
		srcs = append(srcs, s)
	}
	return srcs
}
