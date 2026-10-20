// ABOUTME: Tests a .tpm's typed `our`: its type recorded, its :export or :export_ok deciding what `use` imports.
// ABOUTME: The fixture is testdata/declarations/My/Vars.tpm; a typed `our` elsewhere is refused, as perl refuses it.
package parse

import (
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/types"
)

// TestTypedOurExport: RFC 0001 "Typed variables". A bare `use` imports
// `%Config :export`, and the module's facts hold its type.
func TestTypedOurExport(t *testing.T) {
	wantImports(t, map[string][]string{"use My::Vars;\n": {"%Config"}})
	facts := resolveLibrary(t, "My::Vars")
	got := facts.variables["%Config"]
	if got.Type != types.Hash || got.Element != types.Str|types.Undef {
		t.Errorf("%%Config is %v of %v, want Hash[Str|Undef]", got.Type, got.Element)
	}
}

// TestTypedOurExportOk: `$CRLF :export_ok` is imported only by name.
func TestTypedOurExportOk(t *testing.T) {
	wantImports(t, map[string][]string{"use My::Vars qw($CRLF);\n": {"$CRLF"}})
	if got := importedFrom(t, "use My::Vars;\n"); slices.Contains(got, "$CRLF") {
		t.Errorf("a bare use imports %v; $CRLF is :export_ok", got)
	}
}

// TestTypedOurWithoutExportNotImported: `$plain` carries no mark, so its
// type is recorded and nothing imports it.
func TestTypedOurWithoutExportNotImported(t *testing.T) {
	wantImports(t, map[string][]string{"use My::Vars qw($plain);\n": nil})
	if got := importedFrom(t, "use My::Vars;\n"); slices.Contains(got, "$plain") {
		t.Errorf("a bare use imports %v; $plain carries no export mark", got)
	}
	if got := resolveLibrary(t, "My::Vars").variables["$plain"]; got.Type != types.Str {
		t.Errorf("$plain is %v, want Str", got.Type)
	}
}

// TestTypedOurEmptyImportListImportsNothing: `use M ()` loads and imports
// nothing, an `:export` variable included.
func TestTypedOurEmptyImportListImportsNothing(t *testing.T) {
	wantImports(t, map[string][]string{"use My::Vars ();\n": nil})
}

// TestTypedOurOnlyInDeclarations: a typed `our` is typed Perl, legal only
// in a `.tpm`. Measured on 5.42.0, perl refuses it in a program at compile
// time: `No such class Hash ... syntax error`.
func TestTypedOurOnlyInDeclarations(t *testing.T) {
	if got := refusedAs("our Hash[Str|Undef] %Config :export;\n"); len(got) == 0 {
		t.Errorf("a typed our in a program parses; perl refuses it")
	}
}

// wantDeclarationError wants src, read as a declaration file, to be in
// error with a message holding want, and to record and export no variable.
func wantDeclarationError(t *testing.T, src, want string) {
	t.Helper()
	facts := readDeclaration([]byte(src), nil)
	if len(facts.errs) == 0 || !strings.Contains(facts.errs[0].Error(), want) {
		t.Errorf("%q: errors %v, want one holding %q", src, facts.errs, want)
	}
	if len(facts.variables) > 0 {
		t.Errorf("%q records %v; a declaration in error records nothing", src, facts.variables)
	}
	if len(facts.exports) > 0 {
		t.Errorf("%q exports %v; a declaration in error exports nothing", src, facts.exports)
	}
}

// TestTypedOurUnknownAttributeRejected: a typed `our` takes `:export` or
// `:export_ok`, and a misspelling is an error rather than an unexported
// variable.
func TestTypedOurUnknownAttributeRejected(t *testing.T) {
	wantDeclarationError(t, "our Str $x :exported;\n", ":exported")
}

// TestLexicalExportMarkRejected: only `our` declares a package variable
// Exporter can install -- a `my` or `state` lexical, or a `local`'ed
// global, is never in the symbol table Exporter installs from -- so
// `:export` on any of them is refused on that ground alone, whatever perl
// itself does with the attribute. Measured on 5.42.0, perl's own verdict
// differs by declarator: `my $x :export;` and `state $x :export;` (the
// latter needs `use feature 'state'`) both pass `perl -c` ("syntax OK")
// and die only when run ("Invalid SCALAR attribute: export"); `local $x
// :export;` is a compile-time syntax error, since `local` takes no
// attributes at all; only `our $x :export;` is perl's own compile-time
// refusal ("Invalid SCALAR attribute: export" from `BEGIN`).
func TestLexicalExportMarkRejected(t *testing.T) {
	wantDeclarationError(t, "my $x :export;\n", ":export")
	wantDeclarationError(t, "local $x :export;\n", ":export")
	wantDeclarationError(t, "state $x :export;\n", ":export")
}

// TestOurMarkTrailingCommaRejected: a comma after a marked `our`'s
// attributes makes the declaration one operand of a bigger list --
// `our Int $x :export, $y;` is `(our Int $x :export), $y` -- which this
// parser reads at parse time (the type, recorded before the error is
// found) but never scans for its mark (the per-statement export pass sees
// only a declaration ending its own statement). Read correctly it would
// export nothing silently; refused, the trailing comma is the fix.
func TestOurMarkTrailingCommaRejected(t *testing.T) {
	facts := readDeclaration([]byte("our Int $x :export, $y;\n"), nil)
	if len(facts.errs) == 0 || !strings.Contains(facts.errs[0].Error(), "is not read here") {
		t.Errorf("errors %v, want one holding %q", facts.errs, "is not read here")
	}
	if len(facts.exports) > 0 {
		t.Errorf("exports %v; a declaration in error exports nothing", facts.exports)
	}
}

// TestOurMarkInLongCommaChainRejected: perl's `,` is left-associative, so
// three comma-joined declarations parse as `Binary(",", Binary(",", A,
// B), C)`, A and B sharing one inner node rather than three operands
// sitting beside each other. A mark in the first position -- the one
// buried two levels down -- is read the same as one in the last.
func TestOurMarkInLongCommaChainRejected(t *testing.T) {
	facts := readDeclaration([]byte("our $a :export = 1, our $b = 2, our $c = 3;\n"), nil)
	if len(facts.errs) == 0 || !strings.Contains(facts.errs[0].Error(), "is not read here") {
		t.Errorf("errors %v, want one holding %q", facts.errs, "is not read here")
	}
	if len(facts.exports) > 0 {
		t.Errorf("exports %v; a declaration in error exports nothing", facts.exports)
	}
}

// TestTypedOurUnknownTypeRejected: a type name the lattice does not hold
// is an error, as it is in a signature.
func TestTypedOurUnknownTypeRejected(t *testing.T) {
	wantDeclarationError(t, "our Nope $x :export;\n", `unknown type name "Nope"`)
	wantDeclarationError(t, "our Hash[Nope] %h :export;\n", `unknown type name "Nope"`)
}

// TestTypedOurAggregateWithBareTypeRejected: a hash or array is typed by
// its container, as a slurpy is (RFC 0001, "A slurpy takes no bare element
// type"): `Str %h` leaves the container unsaid.
func TestTypedOurAggregateWithBareTypeRejected(t *testing.T) {
	wantDeclarationError(t, "our Str %h :export;\n", "Hash[Str] %h")
	wantDeclarationError(t, "our Str @a :export;\n", "Array[Str] @a")
}

// TestTypedOurListRejected: a typed or exported `our` declares one
// variable, so a list of them is an error rather than a statement read as
// nothing. Measured on 5.42.0, perl dies on `our Str ($a, $b);` with "No
// such class Str".
func TestTypedOurListRejected(t *testing.T) {
	wantDeclarationError(t, "our Str ($ta, $tb) :export;\n", "declare one variable per statement")
	wantDeclarationError(t, "our Str ($ta, $tb);\n", "declare one variable per statement")
	wantDeclarationError(t, "our ($pa, $pb) :export;\n", "declare one variable per statement")
	wantDeclarationError(t, "our ($pa, $pb) :export_ok(t);\n", "declare one variable per statement")
}

// TestListFormTypedOurIsRead: `our TYPE (VAR) :export;` holds one variable
// in a parenthesised list -- grouping parens that don't multiply it, as
// `my ($x) = @_` is one variable too, forcing list context on its init.
// Its type and export mark are read exactly as the unparenthesised form's
// are; perl drops the class for a list-form `our` (measured on 5.42.0,
// `our Foo ($x) = 1;` deparses identically to `our ($x) = 1;`), but typed
// Perl is this parser's own convention, not perl's grammar, and a single
// name in parens says which variable the type is for.
func TestListFormTypedOurIsRead(t *testing.T) {
	facts := readDeclaration([]byte("package M;\nour Hash[Str] (%a) :export;\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errs = %v, want none", facts.errs)
	}
	got := facts.variables["%a"]
	if got.Type != types.Hash || got.Element != types.Str {
		t.Errorf("%%a is %v of %v, want Hash[Str]", got.Type, got.Element)
	}
	if !slices.Contains(facts.exports, "%a") {
		t.Errorf("exports %v, want %%a", facts.exports)
	}
}

// TestOurUntypedExport: an untyped `our` exports by its mark alone, and
// records no type.
func TestOurUntypedExport(t *testing.T) {
	wantImports(t, map[string][]string{
		"use My::UntypedVars;\n":        {"$x"},
		"use My::UntypedVars qw(@y);\n": {"@y"},
		"use My::UntypedVars ':t';\n":   {"@y"},
		"use My::UntypedVars qw($x);\n": {"$x"},
	})
	if got := resolveLibrary(t, "My::UntypedVars").variables; len(got) > 0 {
		t.Errorf("My::UntypedVars records %v; an untyped our records no type", got)
	}
}
