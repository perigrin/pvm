// ABOUTME: Writes a fixture perl module under a lib directory, the shape every
// ABOUTME: perl-held test needs to build a module perl can require without installing it.

package perltest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
