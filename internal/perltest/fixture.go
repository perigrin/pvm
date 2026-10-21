// ABOUTME: Writes a fixture perl module under a lib directory, the shape every
// ABOUTME: perl-held test needs to build a module perl can require without installing it.

package perltest

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// heldLibPath is internal/parse/testdata/held_lib.pl, held_to_perl.pl
// and return_oracle.pl's own shared rule for whether a call to a sub
// named by its fully-qualified name runs code -- MissingSubs and
// ConstantDifferences require it instead of each embedding their own
// copy of that rule, so there is one callable(), not three. Found
// relative to this package's own source file, not the caller's working
// directory, since a perltest helper is called from several different
// packages' test binaries, each with its own.
func heldLibPath(t testing.TB) string {
	t.Helper()
	return sourceRelative(t, "..", "parse", "testdata", "held_lib.pl")
}

// sourceRelative is the file at parts, relative to this package's own
// source directory, failing t if it is not there.
func sourceRelative(t testing.TB, parts ...string) string {
	t.Helper()
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the perltest package source")
	}
	path := filepath.Join(append([]string{filepath.Dir(self)}, parts...)...)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("%s not found at %s: %v", filepath.Base(path), path, err)
	}
	return path
}

// WriteModule writes module's perl source as a .pm file under lib, at the
// path relative to lib that perl's own @INC resolution of that name
// would use ("::" becomes "/"), creating lib and any package directory
// under it as needed.
func WriteModule(t testing.TB, lib, module, pm string) {
	t.Helper()
	path := filepath.Join(lib, filepath.FromSlash(strings.ReplaceAll(module, "::", "/"))+".pm")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(pm), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
