// ABOUTME: Guard-narrowing fixtures matching the lib sites issue 01a156e4's dba9 gate review classified.
// ABOUTME: Each GUARDED fixture must draw no diagnostic, each REAL one must still draw one.

package infer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// wantSilentResolved is wantSilent over resolvedDiags (flow_test.go):
// loweredDiags's bare parse.Parse resolves no module (parse.go, "a nil
// resolver resolves nothing"), which would leave every reftype/refaddr/sum
// call here Unknown and so silent either way, proving nothing about the
// narrowing these tests exist to check. resolvedDiags's nil Loader is
// enough: these declaration files are embedded, read by path regardless
// of it, and nothing here requires a module's own source.
func wantSilentResolved(t *testing.T, src string) {
	t.Helper()
	for _, d := range resolvedDiags(src) {
		t.Errorf("%q: diagnostic %q, want none", src, d.Message)
	}
}

// wantReportedResolved is wantReported over resolvedDiags.
func wantReportedResolved(t *testing.T, src string, code string) {
	t.Helper()
	if diags := resolvedDiags(src); len(diags) != 1 || diags[0].Code != code {
		t.Errorf("%q: diagnostics %v, want one %s", src, diags, code)
	}
}

// TestNarrowGuardedLibSites reproduces, in isolation, the guard shape of
// each lib site the dba9 gate review classified GUARDED (false positive):
// after the guard, perl 5.42 -w is silent, and so must psc check be, now
// that reftype, refaddr and sum are multis selected by argument type or
// arity (Scalar/Util.tpm, List/Util.tpm) and the guard shapes below narrow.
func TestNarrowGuardedLibSites(t *testing.T) {
	for name, src := range map[string]string{
		// JSON/PP.pm:398,411 -- blessed($obj) proves $obj an Object, a Ref
		// subtype, so refaddr($obj) is an Int, and ref($result) proves
		// $result a Ref the same way.
		"blessed_then_refaddr": `
my ($obj) = @_;
if (blessed($obj)) {
    my $result = $obj->TO_JSON();
    if (defined $result and ref($result)) {
        if (refaddr($obj) eq refaddr($result)) { print "same\n" }
    }
}
`,
		// Math/BigInt.pm:3380's right operand -- `my $zs = $z->{sign}; my
		// $zv = $z->{value}; ... if refaddr($x) eq refaddr($z)`: $z
		// survives an arrow deref right before the call, which under
		// `use strict 'refs'` either autovivifies it from undef into
		// exactly a HashRef or dies on anything else (measured: a plain
		// string there dies "Can't use string ... as a HASH ref"), so
		// reaching the call at all proves $z a Ref.
		"deref_proves_ref": `
my ($z) = @_;
my $zs = $z->{sign};
my $t = refaddr($z) . "!";
print $t;
`,
		// Test2/Mock.pm:356 -- ref($arg) && reftype($arg) eq 'CODE', a
		// short-circuit `&&` whose right operand runs only where ref($arg)
		// held.
		"ref_and_reftype": `
my ($arg) = @_;
return "code" if ref($arg) && reftype($arg) eq 'CODE';
return "other";
`,
		// Test2/Mock.pm:54,57 -- the falsy branch of an if/elsif/elsif
		// chain consumes reftype's Undef: by the first elsif, $type is
		// proven defined.
		"falsy_branch_consumes_undef": `
my ($val) = @_;
my $type = reftype($val);
if (!$type) {
    return "scalar";
}
elsif ($type eq 'HASH') {
    return "hash";
}
elsif ($type eq 'ARRAY') {
    return "array";
}
return "other";
`,
		// Test2/Mock.pm:368 -- after !ref($arg)'s own elsif branch, later
		// elsifs in the same chain know $arg is a Ref.
		"elsif_after_notref": `
my ($arg) = @_;
if (!ref($arg)) {
    return "scalar";
}
elsif (reftype($arg) eq 'HASH') {
    return "hash";
}
return "other";
`,
		// Test2/Compare/Ref.pm:48, Test2/Tools/Ref.pm:54,78 -- `return ...
		// unless ref $x;` rules out non-refs before refaddr runs.
		"unless_return": `
my ($x, $y) = @_;
return undef unless ref $x;
return undef unless ref $y;
return refaddr($x) == refaddr($y) ? "same" : "diff";
`,
		// threads/shared.pm:100-177 -- $item is proven a ref by an early
		// `return $item if !ref($item) || ...`, and reftype($item) read
		// several statements later, across an if/elsif chain, stays Str.
		"early_return_then_later_reftype": `
my ($item) = @_;
return $item if (!ref($item) || $item eq '');
my $ref_type = reftype($item);
if ($ref_type eq 'ARRAY') {
    return "array";
}
elsif ($ref_type eq 'HASH') {
    return "hash";
}
elsif ($ref_type eq 'REF') {
    return "refref";
}
return "other";
`,
		// JSON/PP.pm:398's right operand -- `ref $results[0]`, a named
		// array's literal-index element, not a plain scalar variable.
		"array_element_ref_guard": `
my @results = @_;
if (@results and ref $results[0]) {
    return refaddr($results[0]);
}
return undef;
`,
		// Test2/Compare/Array.pm:35 -- $_, a grep block's topic, under
		// `ref($_) && reftype($_) eq 'CODE'`.
		"underscore_topic_ref_and_reftype": `
my @order = @_;
my @odd = grep { !(ref($_) && reftype($_) eq 'CODE') } @order;
return \@odd;
`,
		// IO/Handle.pm:515 -- `tell qualify($_[0], caller) if ref($_[0])`:
		// $_[0], the invocant, is proven a Ref by the modifier's own guard.
		// The guard mechanics are the same one isElementVarName (flow.go)
		// gives $_[0] along with any other named array's literal-index
		// element; the second half, qualify's own multi correctly
		// selecting its Ref candidate despite `caller`'s unrelated
		// Unknown-ish second argument, is select_test.go's
		// TestSelectCoercionRetryKeepsFittingArgs.
		"arg0_ref_guard": `
my $t = reftype($_[0]) if ref($_[0]);
return $t;
`,
	} {
		t.Run(name, func(t *testing.T) {
			wantSilentResolved(t, "use Scalar::Util qw(reftype refaddr blessed);\n"+src)
		})
	}
}

// resolvedCallType is resolvedDiags's own parse (a nil Loader resolves
// each declaration file by path regardless, "a nil resolver resolves
// nothing" notwithstanding -- see wantSilentResolved), paired with the
// resolved type of the call written as call in src, Unknown when it
// gives none.
func resolvedCallType(src, call string) types.Type {
	root := parse.ParseWithLoader([]byte(src), func(string) ([]byte, bool) { return nil, false })
	f := lower.Lower(root)
	pos := lower.NewPositions(root, f)
	ann, _, _ := infer.Analyze(f, pos, nil, infer.Options{})
	got, _ := findNodeType(typed{f: f, pos: pos, ann: ann}, []byte(src), call)
	return got
}

// TestNarrowGuardedLibSitesResolveSansUndef: WarnPossibleUndef's default
// silence (calls.go's markPossibleUndef) accepts ANY declared call's
// Undef the call's arguments do not prove, so a GUARDED site with no
// narrowing at all would draw no diagnostic either -- wantSilentResolved
// alone cannot tell the two apart (perigrin's 2026-10-21 review of this
// issue: TestNarrowGuardedLibSites passed even with flow.go and vars.go's
// own narrowing reverted to before this issue's merge, since
// markPossibleUndef, untouched by that revert, silenced every subtest
// regardless). This asserts what wantSilentResolved cannot: the guard
// actually narrowed the call to its Ref candidate's own return, Int, with
// no Undef bit left to be silent about at all -- an assertion a
// reverted narrowDerefBase/guardShadows fails outright, each call
// instead resolving to the Any candidate's own Int|Undef.
func TestNarrowGuardedLibSitesResolveSansUndef(t *testing.T) {
	const use = "use Scalar::Util qw(reftype refaddr blessed);\n"
	for _, c := range []struct {
		name, src, call string
	}{
		// blessed_then_refaddr: blessed($obj) proves $obj an Object, so
		// refaddr($obj) selects `(Ref $ref) Int`, not the Any candidate's
		// Int|Undef.
		{"blessed_then_refaddr", `
my ($obj) = @_;
if (blessed($obj)) {
    my $result = $obj->TO_JSON();
    if (defined $result and ref($result)) {
        if (refaddr($obj) eq refaddr($result)) { print "same\n" }
    }
}
`, "refaddr($obj)"},
		// deref_proves_ref: the arrow deref right before the call proves
		// $z a Ref (narrowDerefBase, vars.go).
		{"deref_proves_ref", `
my ($z) = @_;
my $zs = $z->{sign};
my $t = refaddr($z) . "!";
print $t;
`, "refaddr($z)"},
		// unless_return: `return ... unless ref $x;` rules out non-refs
		// before refaddr runs.
		{"unless_return", `
my ($x, $y) = @_;
return undef unless ref $x;
return undef unless ref $y;
return refaddr($x) == refaddr($y) ? "same" : "diff";
`, "refaddr($x)"},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := resolvedCallType(use+c.src, c.call)
			assert.Equal(t, types.Int, got, "%s: %s resolved %v, want Int with no Undef bit", c.name, c.call, got)
		})
	}
}

// TestNarrowKeepsRealLibSites reproduces the guard shape -- or absence of
// one -- at a lib site the dba9 gate review classified REAL: perl 5.42 -w
// does warn there, so psc check must still report it. reftype_unguarded
// and realpath_unguarded moved out to TestReftypeUnguardedIsPossibleUndef
// and TestRealpathUnguardedIsPossibleUndef (perigrin, 2026-10-20): an
// argument that was never proven to rule out the Undef branch, reaching
// a declared return that states it as one of several outcomes -- a
// multi's catch-all candidate or a single declaration alike -- is
// possible, not certain, unlike sum()'s own `()` candidate below, the
// only one zero arguments admit, where the call's own arguments, not
// just the declaration, already answer it.
func TestNarrowKeepsRealLibSites(t *testing.T) {
	for name, src := range map[string]string{
		// List::Util::sum with no arguments is undef; this reaches an
		// operator with no guard on it at all, and sum()'s `()` candidate
		// is the only one zero arguments admit -- there is no narrower
		// sibling to weigh against, since none takes zero arguments either.
		"sum_no_args_unguarded": `
use List::Util qw(sum);
my $s = sum();
return $s + 1;
`,
	} {
		t.Run(name, func(t *testing.T) {
			wantReportedResolved(t, "use Scalar::Util qw(reftype);\n"+src, infer.CodeCoercionMismatch)
		})
	}
}

// TestRealpathUnguardedIsPossibleUndef: CPAN/Distribution.pm:848's own
// shape -- Cwd::realpath($dir) with nothing ruling out an unresolvable
// path -- selects its one declared candidate, `(Str $path) Str|Undef`,
// whatever $dir is, so its Undef is one of two outcomes the declaration
// states, not one the call's argument proved: WarnPossibleUndef's
// default silence (perigrin, 2026-10-20), carried from the call onto
// $a, the variable its result is assigned to (Symbol.PossibleUndef,
// ops.go's narrow), through to the "eq" the value actually reaches.
func TestRealpathUnguardedIsPossibleUndef(t *testing.T) {
	const src = `use Cwd;
my ($dir) = @_;
my $a = Cwd::realpath($dir);
return $a eq "/wanted";
`
	assert.Empty(t, diagsWithOptions(src, infer.Options{}))
	diags := diagsWithOptions(src, infer.Options{WarnPossibleUndef: true})
	if assert.Len(t, diags, 1) {
		assert.Equal(t, infer.CodeCoercionMismatch, diags[0].Code)
	}
}

// TestReftypeOfRefIsStr: reftype selects its Ref candidate, Str, not the
// Any candidate's Str|Undef, once a ref() guard proves its argument a
// Ref -- Scalar/Util.tpm's `multi sub reftype`, RFC 0001 "Multi
// declarations". perl 5.42 -w is silent concatenating reftype($x) once
// ref($x) held, as it is before any such guard too -- reftype's own
// catch-all candidate, selected over an argument never proven a Ref, is
// WarnPossibleUndef's default silence (TestReftypeUnguardedIsPossibleUndef
// covers that state directly; perigrin, 2026-10-20).
func TestReftypeOfRefIsStr(t *testing.T) {
	wantSilentResolved(t, `use Scalar::Util qw(reftype);
my ($x) = @_;
if (ref($x)) {
    my $t = reftype($x) . "!";
    print $t;
}
`)
	wantSilentResolved(t, `use Scalar::Util qw(reftype);
my ($x) = @_;
my $t = reftype($x) . "!";
print $t;
`)
}

// TestReftypeUnguardedIsPossibleUndef: Safe.pm:424's own shape --
// reftype($sub) with no ref() or defined() proof -- selects reftype's
// catch-all `(Any $ref) Str|Undef` candidate over a Scalar argument
// never proven a Ref, so its Undef is WarnPossibleUndef's default
// silence, not a certain mismatch (perigrin, 2026-10-20): $sub's own
// type, unlike a plain non-ref literal, still overlaps Ref, so psc
// cannot rule out that the call could have reached reftype's narrower,
// Undef-free `(Ref $ref) Str` candidate instead.
func TestReftypeUnguardedIsPossibleUndef(t *testing.T) {
	const src = `use Scalar::Util qw(reftype);
my ($sub) = @_;
if (reftype($sub) ne 'CODE') { die "not a code ref" }
`
	assert.Empty(t, diagsWithOptions(src, infer.Options{}))
	diags := diagsWithOptions(src, infer.Options{WarnPossibleUndef: true})
	if assert.Len(t, diags, 1) {
		assert.Equal(t, infer.CodeCoercionMismatch, diags[0].Code)
	}
}

// TestEncodeTextIsStr: encode selects its Str candidate, Str, not the
// Str|Undef candidate's join, once its $string argument is itself
// proven Str -- Encode.tpm's `multi sub encode`, the same shape as
// Scalar/Util.tpm's reftype. perl 5.42 -w is silent concatenating
// encode("UTF-8", $text, 0) once $text is join of @_, a Str, join
// never undef. Pod/Text.pm's own call has this proof on $text but not
// on $encoding (a hash element, Unknown); it still selects Str, once
// Select's Unknown bypass only counts an Unknown at a position the
// candidates actually differ on
// (select_test.go's TestSelectUnknownAtNonDiscriminatingPositionDecides).
//
// Without that proof, $text joins encode's two candidates to Str|Undef
// (types.Select's own coercion retry, not a genuinely Unknown
// argument): this is one of two outcomes the declaration states, not
// one the call's own argument proved, so it is WarnPossibleUndef's
// default silence (perigrin, 2026-10-20), reported only under the flag.
func TestEncodeTextIsStr(t *testing.T) {
	wantSilentResolved(t, `use Encode qw(encode);
my (@parts) = @_;
my $text = join('', @parts);
my $s = encode("UTF-8", $text, 0) . "!";
print $s;
`)
	const unguarded = `use Encode qw(encode);
my ($text) = @_;
my $s = encode("UTF-8", $text, 0) . "!";
print $s;
`
	assert.Empty(t, diagsWithOptions(unguarded, infer.Options{}))
	diags := diagsWithOptions(unguarded, infer.Options{WarnPossibleUndef: true})
	if assert.Len(t, diags, 1) {
		assert.Equal(t, infer.CodeCoercionMismatch, diags[0].Code)
	}
}

// Symbol.tpm (issue db4c) ships from here:
// declaring `qualify` a multi by argument type (Ref returns that Ref,
// Str returns Str) measures correctly standing alone, and IO/Handle.pm:515
// calls it with a second argument, `caller`, whose own type -- CORE.tpm's
// `caller (Optional[Int] $level) List` -- answers as List in the
// UnknownCtx a `$`-shaped declared-call slot leaves a scalar argument in
// (calls.go's slotContext, deliberately, so an aggregate there reaches
// checkArg as itself); neither multi candidate's `Optional[Str] $pkg`
// fits a List argument, so this reaches types.Select's coercion retry
// (select.go), fixed to keep $_[0]'s own proven Ref rather than zero it
// along with $pkg's non-fitting one (TestSelectCoercionRetryKeepsFittingArgs,
// select_test.go). The pureExporters batch (heldtoperl_test.go) holds it
// to perl along with the rest.
//
// TestSumWithArgsIsNum: List::Util::sum is declared by arity (RFC 0001
// "Multi declarations") -- `multi sub sum (Num $x, List @rest) Num`
// beside `multi sub sum () Undef`, both stating sum's one real
// prototype, `@`, which admits either's arity (parse/derive.go's
// prototypeAdmitsArity; TestDeriveMultiCandidatesAdmittedByArity,
// parse/derive_test.go, pins the declaration itself). perl 5.42 -w is
// silent concatenating sum(1,2,3); it is not silent, and psc check
// must still report it, for sum() with no arguments.
func TestSumWithArgsIsNum(t *testing.T) {
	wantSilentResolved(t, `use List::Util qw(sum);
my $s = sum(1, 2, 3) . "!";
print $s;
`)
	wantReportedResolved(t, `use List::Util qw(sum);
my $s = sum() . "!";
print $s;
`, infer.CodeCoercionMismatch)
}
