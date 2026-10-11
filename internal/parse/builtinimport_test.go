// ABOUTME: The builtin:: functions are the interpreter's own, and `use builtin LIST`
// ABOUTME: imports them with builtin.c's prototypes, so `blessed $obj` is a call.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestBuiltinImportsHaveShapes: builtin.c's boot_core_builtin registers every
// builtin:: function at interpreter start, with a prototype chosen by its
// call checker -- "" for the constants, "$" for the one-argument functions,
// "@" for indexed. `use builtin LIST` imports them by name. Unknown, a word
// before `$obj` is an indirect method call; known, it is a named unary.
// Measured on 5.42.0:
//
//	$ perl -MO=Deparse,-p -e 'use builtin qw(blessed reftype);
//	      is(blessed $obj, "T", "x"); my $w = builtin::reftype $r, 1;
//	      use builtin "true"; my @a = (true, 1);'
//	is(scalar(builtin::blessed($obj)), 'T', 'x');
//	((my($w) = builtin::reftype($r)), '???');
//	(my(@a) = ((!0), 1));
//
// class/construct.t:29 has `is(blessed $obj, "Testcase1", ...)` after
// `use builtin qw( blessed reftype );`.
func TestBuiltinImportsHaveShapes(t *testing.T) {
	for _, tc := range []struct {
		src, name string
		args      int
	}{
		{`use builtin qw( blessed reftype ); is(blessed $obj, "T", "x");`, "blessed", 1},
		{`my $w = builtin::reftype $r, 1;`, "builtin::reftype", 1},
		{`use builtin 'true'; my @a = (true, 1);`, "true", 0},
		// A list operator: it takes everything, so only resolution is asserted.
		{`use builtin qw(indexed); my @p = indexed @a, 1;`, "indexed", -1},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", tc.src, shape(root))
			continue
		}
		call := findCall(root, tc.name)
		if call == nil {
			t.Errorf("%q: no call to %s; got %s", tc.src, tc.name, shape(root))
			continue
		}
		if !call.Resolved {
			t.Errorf("%q: %s is the interpreter's and must resolve", tc.src, tc.name)
		}
		if tc.args >= 0 && len(call.Children) != tc.args {
			t.Errorf("%q: %s takes %d argument(s), got %d; tree %s",
				tc.src, tc.name, tc.args, len(call.Children), shape(root))
		}
	}
}

// TestBuiltinBundlesImportTheirNames: a version bundle, by `use builtin
// ':5.39'` or by `use v5.39` and later, imports every builtin that is not
// experimental -- the 13 names builtin.pm's :5.40 table lists. The NO_BUNDLE
// ones (inf, nan, is_bool, stringify, created_as_*, load_module,
// export_lexically) come in only by name, and `use v5.38` imports no
// builtins. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'use strict; my $x; use v5.40; my $r = blessed $x, 1;'
//	my $r = builtin::blessed($x), '???';
//	$ perl -c -e 'use strict; my $x; use v5.38; my $r = blessed $x, 1;'
//	Scalar found where operator expected (Do you need to predeclare "blessed"?)
//	$ perl -MO=Deparse -e 'use strict; my $x; use builtin ":5.42"; my $r = is_bool $x;'
//	my $r = $x->is_bool;
//
// ':5.39' through ':5.42' are accepted; ':5.38' is "not supported" and
// ':5.40.1' and ':5.039' are "Invalid version bundle".
func TestBuiltinBundlesImportTheirNames(t *testing.T) {
	bundle := []string{"true", "false", "weaken", "unweaken", "is_weak", "blessed",
		"refaddr", "reftype", "ceil", "floor", "is_tainted", "trim", "indexed"}
	noBundle := []string{"inf", "nan", "is_bool", "stringify", "created_as_string",
		"created_as_number", "load_module", "export_lexically"}
	call := func(name string) string {
		switch name {
		case "true", "false", "inf", "nan":
			return "my $r = " + name + ";"
		case "indexed":
			return "my @r = indexed @a;"
		}
		return "my $r = " + name + " $x;"
	}
	resolves := func(src, name string) bool {
		c := findCall(parse.Parse([]byte(src)), name)
		return c != nil && c.Resolved
	}
	for _, header := range []string{"use builtin ':5.39';", "use builtin ':5.40';",
		"use builtin ':5.42';", "use v5.39;", "use v5.40;", "use v5.42;", "use 5.040;"} {
		for _, name := range bundle {
			if src := header + " " + call(name); !resolves(src, name) {
				t.Errorf("%q: the bundle imports %s; got %s", src, name, shape(parse.Parse([]byte(src))))
			}
		}
		for _, name := range noBundle {
			if src := header + " " + call(name); resolves(src, name) {
				t.Errorf("%q: %s is NO_BUNDLE and must not be imported", src, name)
			}
		}
	}
	for _, header := range []string{"use v5.36;", "use v5.38;", "use builtin ':5.38';",
		"use builtin ':5.40.1';", "use builtin ':5.039';"} {
		for _, name := range append(bundle, noBundle...) {
			if src := header + " " + call(name); resolves(src, name) {
				t.Errorf("%q: imports no builtins, yet %s resolved", src, name)
			}
		}
	}
	// A bundle's names keep their prototypes: blessed is a named unary.
	src := `use v5.40; my $r = blessed $x, 1;`
	if c := findCall(parse.Parse([]byte(src)), "blessed"); c == nil || len(c.Children) != 1 {
		t.Errorf("%q: blessed takes one argument; got %s", src, shape(parse.Parse([]byte(src))))
	}
}
