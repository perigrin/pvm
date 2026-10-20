// ABOUTME: Tests for the lowered path's variable, subscript and match rules: what a bound match, a hash element and a slot's context are typed.
// ABOUTME: Each case cites perl 5.42 -w's verdict, and each fix carries a control perl reports that must stay reported.

package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/types"
)

// A named unary operator binds looser than `=~`, so `chr $u =~ /re/` is
// chr($u =~ /re/): perl 5.42 deparses it so, and -w is silent, printing
// "ascii" for $u = 65. A `$` slot takes the match in scalar context, where
// it is a Boolean, not a List. A match in list context is still a List, and
// an aggregate in a `$` slot is still the mistake perl warns about:
// `length(1 ? @a : "x")` draws "length() used on @a". A reference there is
// a lossy coercion, reported under WarnLossyCoercion.
func TestChrOfBoundMatchIsScalar(t *testing.T) {
	for _, src := range []string{
		"my $u = 65;\nprint \"ascii\\n\" if chr $u =~ /[[:^ascii:]]/;\n",
		"my $u = 65;\nmy $c = chr($u =~ /x/);\n",
		"my $m;\nmy $c = chr($m =~ /x/);\n",
		"my $c = chr(;",
	} {
		_, diags := analyzeSource(t, []byte(src))
		for _, d := range diags {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	src := []byte("my $u = 65;\nmy @m = ($u =~ /(\\d)/);\n")
	ty, _ := analyzeSource(t, src)
	if got, ok := findNodeType(ty, src, `($u =~ /(\d)/)`); !ok || got != types.List {
		t.Errorf("%q: the match in list context is %v (found %v), want List", src, got, ok)
	}

	for _, src := range []string{
		"my @a = (1, 2);\nmy $n = length(1 ? @a : \"x\");\n",
		"my $r = {};\nmy $c = chr($r);\n",
	} {
		if _, diags, _ := analyzeSourceWithOptions(t, []byte(src), infer.Options{WarnLossyCoercion: true}); len(diags) == 0 {
			t.Errorf("%q: no diagnostic, want the argument reported", src)
		}
	}
}

// A hash literal's element under a constant key is the value given that
// key, not the union of every value: perl 5.42 -w is silent on
// `{count => 1, values => [2]}->{count} == 1`, and on `$h{count} + 1` and
// `$h{"count"} + 1` over `%h = (count => 1, values => [2])`. A key holding a
// broad value, a Scalar, is checked as that. A string value under its own
// key is still reported where perl warns: `{name => "abc"}->{name} + 1`
// draws "Argument "abc" isn't numeric".
func TestHashElementTypeIsPerKey(t *testing.T) {
	for _, src := range []string{
		"my $source = { count => 1, values => [2] };\nprint \"ok\\n\" if $source->{count} == 1;\n",
		"my %h = (count => 1, values => [2]);\nmy $y = $h{count} + 1;\nmy $z = $h{\"count\"} + 1;\n",
		"my $m;\nmy $h = { a => $m, b => [1] };\nmy $y = $h->{a} + 1;\n",
		"my $h = { a => (;",
	} {
		_, diags := analyzeSource(t, []byte(src))
		for _, d := range diags {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	src := []byte("my $source = { count => 1, values => [2] };\nmy $c = $source->{count};\n")
	ty, _ := analyzeSource(t, src)
	if got, ok := findNodeType(ty, src, `$source->{count}`); !ok || got != types.Int {
		t.Errorf("%q: $source->{count} is %v (found %v), want Int", src, got, ok)
	}

	// A Str element's content is not tracked, so it is a lossy coercion,
	// reported under WarnLossyCoercion as perigrin ruled on 2026-10-14.
	for _, src := range []string{
		"my $h = { name => \"abc\", n => [1] };\nmy $x = $h->{name} + 1;\n",
		"my %h = (name => \"abc\", n => [1]);\nmy $x = $h{name} + 1;\n",
	} {
		if diags := lossyDiags(t, src); len(diags) != 1 || diags[0].Code != infer.CodeLossyCoercion {
			t.Errorf("%q: diagnostics %v, want one lossy coercion", src, diags)
		}
	}
}

// A hash filled from a key/value list holds its values as elements, not
// its keys: perl 5.42 -w says nothing of the numbers on
// `$left{$op} == $left{$l}` over `%left = ('multiply' => 19, 'i_multiply'
// => 19)`, warning only that $op is undefined. When an element of the list
// may flatten, which are keys is not known, and no element type is guessed.
// A string value is still reported where perl warns, under
// WarnLossyCoercion: `$h{$k} + 1` over `%h = ("a" => "abc")` draws
// "Argument "abc" isn't numeric".
func TestKeyValueHashElementsKeepValueTypes(t *testing.T) {
	for _, src := range []string{
		"my (%left, $op, $l) = ((), \"multiply\", \"multiply\");\n%left = ('multiply' => 19, 'i_multiply' => 19);\nprint \"same\\n\" if $left{$op} && $left{$l} and $left{$op} == $left{$l};\n",
		"my %left = ('multiply' => 19, 'i_multiply' => 19);\nmy $k = 'multiply';\nmy $y = $left{$k} + 1;\n",
		"my %a = ('x' => 'y');\nmy %h = (%a, 'b' => 1);\nmy $k = 'b';\nmy $y = $h{$k} + 1;\n",
		"my %h = ('a' => (;",
	} {
		_, diags := analyzeSource(t, []byte(src))
		for _, d := range diags {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	// A Str element's content is not tracked, so it is a lossy coercion,
	// reported under WarnLossyCoercion as perigrin ruled on 2026-10-14.
	for _, src := range []string{
		"my %h;\n%h = (\"a\" => \"abc\", \"b\" => \"x\");\nmy $k = \"a\";\nmy $x = $h{$k} + 1;\n",
	} {
		if diags := lossyDiags(t, src); len(diags) != 1 || diags[0].Code != infer.CodeLossyCoercion {
			t.Errorf("%q: diagnostics %v, want one lossy coercion", src, diags)
		}
	}
}

// A literal range in list context is a list, not the flip-flop's Str, and
// an array it fills holds its values: perl 5.42, `my @r = (1..5)` is five
// Ints, `$r[0] + 1` is 2, and `my @s = ("a".."c")` is ("a", "b", "c"),
// each a Str. Written without spaces, `1..5` is still a range.
func TestLiteralRangeInfersAList(t *testing.T) {
	for _, c := range []struct {
		src, rng string
		elem     types.Type
	}{
		{"my @r = (1..5);\nmy $n = $r[0] + 1;\n", "(1..5)", types.Int},
		{"my @r = (\"a\"..\"c\");\nmy $n = $r[0] . \"!\";\n", "(\"a\"..\"c\")", types.Str},
	} {
		ty, diags, st := analyzeLowered([]byte(c.src), nil, infer.Options{})
		if got, _ := findNodeType(ty, []byte(c.src), c.rng); got != types.List {
			t.Errorf("%q: %s typed %v, want List", c.src, c.rng, got)
		}
		if sym, ok := st.Lookup("@r"); !ok || sym.ElemType != c.elem {
			t.Errorf("%q: @r's elements are %v, want %v", c.src, sym.ElemType, c.elem)
		}
		for _, d := range diags {
			t.Errorf("%q: diagnostic %q, want none", c.src, d.Message)
		}
	}
}

// An aggregate takes its initialiser in list context, so `my @a = (1)` is
// an array of one Int, not an Int: perl 5.42, `my @a = (1); print
// scalar(@a), ref(\@a)` prints "1ARRAY", and -w is silent on `my @a =
// ("x"); push @a, 3` and `my @n = (1); my $s = shift @n; $s + 1`. The
// element type is still recorded, so what perl warns on is still
// reported: `my $x = undef; my @u = ($x); $u[0] + 1` warns "Use of
// uninitialized value in addition (+)".
func TestParenInitialiserIsListContext(t *testing.T) {
	for _, c := range []struct {
		src, name string
		want      types.Type
	}{
		{"my @a = (1);\n", "@a", types.Array},
		{"my $x = \"s\";\nmy @a = ($x);\n", "@a", types.Array},
		{"our @o = (1);\n", "@o", types.Array},
	} {
		if got, _ := loweredSymbolType(t, c.src, c.name); got != c.want {
			t.Errorf("%q: %s is %v, want %v", c.src, c.name, got, c.want)
		}
	}
	for _, src := range []string{
		"my @a = (\"x\");\npush @a, 3;\n",
		"my @n = (1);\nmy $s = shift @n;\nprint $s + 1;\n",
	} {
		for _, d := range loweredDiags(src) {
			t.Errorf("%q: diagnostic %q, want none", src, d.Message)
		}
	}

	src := "my $x = undef;\nmy @u = ($x);\nprint $u[0] + 1;\n"
	if diags := loweredDiags(src); len(diags) != 1 {
		t.Errorf("%q: diagnostics %v, want one", src, diags)
	}
}

// Copying an array copies its elements, so the copy records their type,
// not the Array: perl 5.42, `my @b = (1, 2); my @c = @b; $c[0] + 1` is 2,
// and `my @d = ("p"); my @e = @d; $e[0] . "!"` is "p!". An undefined
// element is copied too: `my @u = (undef, undef); my @w = @u; $w[0] + 1`
// warns
// "Use of uninitialized value in addition (+)".
func TestArrayCopyKeepsElementType(t *testing.T) {
	for _, c := range []struct {
		src  string
		want types.Type
	}{
		{"my @b = (1, 2);\nmy @c = @b;\n", types.Int},
		{"my @d = (\"p\");\nmy @c = @d;\n", types.Str},
	} {
		_, _, st := analyzeLowered([]byte(c.src), nil, infer.Options{})
		if sym, ok := st.Lookup("@c"); !ok || sym.ElemType != c.want {
			t.Errorf("%q: @c's elements are %v, want %v", c.src, sym.ElemType, c.want)
		}
	}

	src := "my @u = (undef, undef);\nmy @w = @u;\nprint $w[0] + 1;\n"
	if diags := loweredDiags(src); len(diags) != 1 {
		t.Errorf("%q: diagnostics %v, want one", src, diags)
	}
}

// TestHashDerefCopyDoesNotRecordItsOwnScalarCountAsElemType: `%data =
// %$self` hands %data its referent's elements, not `%$self`'s own value
// -- a hash in scalar context is its key count, an Int. Recording that
// Int as every element's type, as an untracked copied array's source
// gives Unknown rather than its own type (TestArrayCopyKeepsElementType
// above), would make an unrelated key's read, `$data{foo}`, type as Int
// rather than Unknown and wrongly satisfy nothing it should. Measured on
// 5.42.0, `Storable::dclone($data{foo})` with `%data = %$self` and no
// "foo" key in %$self dies "Not a reference", never silently passing an
// Int where dclone wants a Ref -- the honest answer for an unlisted key
// is Unknown, which this does not check as a type mismatch either way.
func TestHashDerefCopyDoesNotRecordItsOwnScalarCountAsElemType(t *testing.T) {
	src := "my $self = {};\nmy %data = %$self;\n"
	_, _, st := analyzeLowered([]byte(src), nil, infer.Options{})
	if sym, ok := st.Lookup("%data"); ok && sym.ElemType != types.Unknown {
		t.Errorf("%q: %%data's elements are %v, want Unknown", src, sym.ElemType)
	}

	wantSilent(t, "use Storable qw/dclone/;\nmy $self = {};\nmy %data = %$self;\ndclone($data{foo});\n")
}

// A reference-aliasing declaration without parens takes its initialiser
// in scalar context, as it binds the one reference it is given: perl
// 5.42, under `use feature qw(refaliasing declared_refs)`, `my \@a = f()`
// calls f with wantarray false, and `my (\@b) = f()` with it true. So
// `localtime()` there is typed as `my $s = localtime()` types it, while
// in the parenthesised form, and for an array declared without the
// alias, it is a List.
func TestRefAliasDeclContext(t *testing.T) {
	localtime := func(src string) types.Type {
		ty, _, _ := analyzeLowered([]byte(src), nil, infer.Options{})
		got, _ := findNodeType(ty, []byte(src), "localtime()")
		return got
	}
	scalar := localtime("my $s = localtime();\n")
	if scalar == types.List {
		t.Fatalf("my $s = localtime(): localtime() typed List")
	}
	for _, c := range []struct {
		src  string
		want types.Type
	}{
		{"my \\@a = localtime();\n", scalar},
		{"my (\\@b) = localtime();\n", types.List},
		{"my @c = localtime();\n", types.List},
	} {
		if got := localtime(c.src); got != c.want {
			t.Errorf("%q: localtime() typed %v, want %v", c.src, got, c.want)
		}
	}
}
