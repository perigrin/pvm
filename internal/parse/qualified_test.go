// ABOUTME: Tests that a qualified call, Carp::confess(...), carries the signatures its module's .tpm declares.
// ABOUTME: A name the .tpm does not declare, or a module with no .tpm, carries none.
package parse

import (
	"reflect"
	"testing"
	"testing/fstest"
)

// qualifiedCall parses src, loading no module source, and returns the
// call to name in it.
func qualifiedCall(t *testing.T, src, name string) *Node {
	t.Helper()
	call := findCallNamed(ParseWithLoader([]byte(src), noModules), name)
	if call == nil {
		t.Fatalf("%s: no call to %s", src, name)
	}
	return call
}

// TestQualifiedCallCarriesSignatures: loading a module makes its subs
// callable by their qualified names, and such a call carries the
// signatures the module's .tpm declares, as an imported call does.
func TestQualifiedCallCarriesSignatures(t *testing.T) {
	carp := resolveLibrary(t, "Carp")
	call := qualifiedCall(t, "use Carp (); Carp::confess(\"x\");\n", "Carp::confess")
	if want := carp.signatures["confess"]; len(want) == 0 || !reflect.DeepEqual(call.Signatures, want) {
		t.Errorf("Carp::confess: signatures %+v, want %+v", call.Signatures, want)
	}

	// A module loaded with no import list at all: its .tpm here is a
	// test's own, standing in until Scalar/Util.tpm ships.
	t.Cleanup(OverlayDeclarations(fstest.MapFS{
		"declarations/Scalar/Util.tpm": {Data: []byte("package Scalar::Util;\nsub blessed :export_ok (Any $x) Str;\n")},
	}))
	util := resolveLibrary(t, "Scalar::Util")
	call = qualifiedCall(t, "use Scalar::Util (); my $x; Scalar::Util::blessed($x);\n", "Scalar::Util::blessed")
	if want := util.signatures["blessed"]; len(want) == 0 || !reflect.DeepEqual(call.Signatures, want) {
		t.Errorf("Scalar::Util::blessed: signatures %+v, want %+v", call.Signatures, want)
	}
}

// TestQualifiedCallUndeclaredNameHasNoSignatures: Carp.tpm declares no
// nosuch, so a call to it carries nothing.
func TestQualifiedCallUndeclaredNameHasNoSignatures(t *testing.T) {
	call := qualifiedCall(t, "use Carp (); Carp::nosuch(1);\n", "Carp::nosuch")
	if call.Signatures != nil {
		t.Errorf("Carp::nosuch: signatures %+v, want none", call.Signatures)
	}
}

// TestQualifiedCallUndeclaredModuleHasNoSignatures: a module with no .tpm
// declares nothing a call could carry.
func TestQualifiedCallUndeclaredModuleHasNoSignatures(t *testing.T) {
	load := func(m string) ([]byte, bool) {
		if m == "No::Tpm" {
			return []byte("package No::Tpm; sub f { 1 } 1;\n"), true
		}
		return nil, false
	}
	call := findCallNamed(ParseWithLoader([]byte("use No::Tpm (); No::Tpm::f(1);\n"), load), "No::Tpm::f")
	if call == nil {
		t.Fatal("no call to No::Tpm::f")
	}
	if call.Signatures != nil {
		t.Errorf("No::Tpm::f: signatures %+v, want none", call.Signatures)
	}
}
