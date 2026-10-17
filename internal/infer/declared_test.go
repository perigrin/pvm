// ABOUTME: Tests that inference types a call to a declared sub from the signatures the parser resolved, RFC 0001 "Call sites".
// ABOUTME: The fixture modules are internal/parse/testdata/declarations/My/*.tpm, read through parse.OverlayDeclarations.

package infer_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// analyzeDeclared analyses src with the parse's fixture declarations
// beside the shipped ones, resolving `use` through load, which may be nil.
func analyzeDeclared(t *testing.T, src string, load parse.Loader) (typed, []infer.Diagnostic, *infer.SymbolTable) {
	t.Helper()
	t.Cleanup(parse.OverlayDeclarations(os.DirFS("../parse/testdata")))
	if load == nil {
		load = func(string) ([]byte, bool) { return nil, false }
	}
	r := infer.AnalyzeParsed(parse.ParseWithLoader([]byte(src), load), nil, infer.Options{})
	return typed{f: r.File, pos: r.Positions, ann: r.Annotations}, r.Diagnostics, r.Symbols
}

// declaredCallType is the type inference gives the call written as call in
// src, Unknown when it gives none, and src's diagnostics.
func declaredCallType(t *testing.T, src, call string) (types.Type, []infer.Diagnostic) {
	t.Helper()
	ty, diags, _ := analyzeDeclared(t, src, nil)
	got, _ := findNodeType(ty, []byte(src), call)
	return got, diags
}

// TestInferTypesCallFromDeclaration: My::Lib declares `sub join_all
// :export (Str $sep, List[Str] @parts) Str`, so a call to the imported
// join_all is a Str.
func TestInferTypesCallFromDeclaration(t *testing.T) {
	src := `use My::Lib; my $s = join_all(",", "a", "b");`
	got, diags := declaredCallType(t, src, `join_all(",", "a", "b")`)
	assert.Equal(t, types.Str, got)
	assert.Empty(t, diags)
}

// TestInferSelectsMultiAtCallSite: a multi's candidate is selected at the
// call by its context and its arguments' types. Measured on 5.42.0,
// `my $t = localtime(0)` is "Thu Jan  1 00:00:00 1970" and `my @l =
// localtime(0)` nine numbers; `my %h = (a => 1); my $k = each %h` is "a"
// and `my @a = (5); my $i = each @a` is 0. A call whose context the site
// cannot determine, a sub's return value, joins its candidates. My::Lib's
// `when_` is a library multi forking on context the same way.
func TestInferSelectsMultiAtCallSite(t *testing.T) {
	cases := []struct {
		src, call string
		want      types.Type
	}{
		{"my $t = localtime(0);", "localtime(0)", types.Str | types.Undef},
		{"my @l = localtime(0);", "localtime(0)", types.List},
		{"my %h = (a => 1); my $k = each %h;", "each %h", types.Str | types.Undef},
		{"my @a = (5); my $i = each @a;", "each @a", types.Int | types.Undef},
		{"sub f { return localtime($_[0]) }", "localtime($_[0])", types.Str | types.Undef | types.List},
		{"use My::Lib; my $w = when_();", "when_()", types.Str},
		{"use My::Lib; my @w = when_();", "when_()", types.List},
	}
	for _, c := range cases {
		got, diags := declaredCallType(t, c.src, c.call)
		assert.Equal(t, c.want, got, "%q", c.src)
		assert.Empty(t, diags, "%q", c.src)
	}
}

// TestInferUndeclaredImportUntyped: a sub imported from a module with no
// .tpm has no declaration to type its call from.
func TestInferUndeclaredImportUntyped(t *testing.T) {
	load := func(name string) ([]byte, bool) {
		return []byte("package Plain;\nuse Exporter 'import';\nour @EXPORT = qw(plain);\nsub plain { 1 }\n1;\n"), name == "Plain"
	}
	src := "use Plain; my $x = plain(1);"
	ty, diags, _ := analyzeDeclared(t, src, load)
	got, _ := findNodeType(ty, []byte(src), "plain(1)")
	assert.Equal(t, types.Unknown, got)
	assert.Empty(t, diags)
}

// TestInferScalarSortHasNoType: a sort in scalar context is not the List
// its list-context candidates return, and the call site gives it no type.
// Measured on 5.42.0, `my @x = (3, 1); my $n = sort @x` leaves $n undef,
// warning "Useless use of sort in scalar context".
func TestInferScalarSortHasNoType(t *testing.T) {
	got, _ := declaredCallType(t, "my @x = (3, 1); my $n = sort @x;", "sort @x")
	assert.NotEqual(t, types.List, got)
	assert.Equal(t, types.Unknown, got)
}

// TestInferNoCandidateArityIsError: My::Multi's two takes one argument or
// two, and has no prototype, so perl compiles `two(1, 2, 3)` and the
// declaration rules it out: an error at the call, which has no type. A
// call passing an array has no count the declaration can rule out.
func TestInferNoCandidateArityIsError(t *testing.T) {
	got, diags := declaredCallType(t, "use My::Multi; my $r = two(1, 2, 3);", "two(1, 2, 3)")
	assert.Equal(t, types.Unknown, got)
	require.Len(t, diags, 1)
	assert.Equal(t, infer.Error, diags[0].Severity)
	assert.Equal(t, infer.CodeArityMismatch, diags[0].Code, diags[0].Message)

	// An array flattens into the call's arguments, however many it holds.
	_, diags = declaredCallType(t, "use My::Multi; my @a = (1, 2, 3); my $r = two(@a, 4, 5);", "two(@a, 4, 5)")
	assert.Empty(t, diags)
}

// TestInferArityChoosesCandidate: CORE.tpm's select is a Str with one
// argument and a List with four. Measured on 5.42.0, `select(STDERR)` is
// "main::STDOUT" and `select(undef, undef, undef, 0)` is (0, 0). My::Multi's
// two is chosen by its arity the same way.
func TestInferArityChoosesCandidate(t *testing.T) {
	cases := []struct {
		src, call string
		want      types.Type
	}{
		{"open(my $fh, '<', '/dev/null'); my $h = select($fh);", "select($fh)", types.Str},
		{"my @r = select(undef, undef, undef, 0.25);", "select(undef, undef, undef, 0.25)", types.List},
		{"use My::Multi; my $r = two(1);", "two(1)", types.Int},
		{"use My::Multi; my $r = two(1, 2);", "two(1, 2)", types.Str},
	}
	for _, c := range cases {
		got, diags := declaredCallType(t, c.src, c.call)
		assert.Equal(t, c.want, got, "%q", c.src)
		assert.Empty(t, diags, "%q", c.src)
	}
}

// TestInferArgTypesChooseSameArityCandidate: My::Multi's which takes an
// Int, returning an Int, or a CodeRef, returning a Str.
func TestInferArgTypesChooseSameArityCandidate(t *testing.T) {
	cases := []struct {
		src, call string
		want      types.Type
	}{
		{"use My::Multi; my $r = which(1);", "which(1)", types.Int},
		{"use My::Multi; my $r = which(sub { 1 });", "which(sub { 1 })", types.Str},
	}
	for _, c := range cases {
		got, diags := declaredCallType(t, c.src, c.call)
		assert.Equal(t, c.want, got, "%q", c.src)
		assert.Empty(t, diags, "%q", c.src)
	}
}

// TestInferUnimportedExportOkUntyped: My::OkOnly marks f `:export_ok`, so
// a bare `use My::OkOnly` does not import it and the call is to no
// declared sub; naming f in the use list does import it.
func TestInferUnimportedExportOkUntyped(t *testing.T) {
	got, _ := declaredCallType(t, "use My::OkOnly; my $y = f(1);", "f(1)")
	assert.Equal(t, types.Unknown, got)
	got, _ = declaredCallType(t, "use My::OkOnly qw(f); my $y = f(1);", "f(1)")
	assert.Equal(t, types.Int, got, "the fixture declares f, imported when asked for")
}

// TestInferLocalSubShadowsDeclaration: a file's own sub lone, defined
// before the call, is the one perl calls, not My::Multi's declared one.
// Measured on 5.42.0 with a module exporting `pick`, a local `sub pick`
// makes a later `pick(1)` call the local sub ("Subroutine pick
// redefined").
func TestInferLocalSubShadowsDeclaration(t *testing.T) {
	got, diags := declaredCallType(t, `use My::Multi; sub lone { return "s" } my $r = lone([]);`, "lone([])")
	assert.Equal(t, types.Str, got)
	assert.Empty(t, diags)
}
