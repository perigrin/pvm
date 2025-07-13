//go:build darwin || freebsd || linux

package tree_sitter_typed_perl

import "github.com/ebitengine/purego"

// loadLibraryPlatform handles Unix-like library loading
func loadLibraryPlatform(path string) (uintptr, error) {
	return purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}
