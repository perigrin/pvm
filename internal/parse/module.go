// ABOUTME: ParseFile and the module loader — a `use` resolved against the file's own directory.
// ABOUTME: Loading enriches; it never fails a parse, because an unresolved import is a result and not an error.

package parse

import (
	"os"
	"path/filepath"
	"strings"
)

// Loader answers "what is the source of this module?" by name.
//
// It is the seam between the parser and the filesystem. Production searches
// directories; a test hands back bytes it wrote itself, so the suite needs no
// installed module and no particular layout to be true.
//
// The bool is "found", not "succeeded". A module that cannot be located and
// one that cannot be read are the same answer here, because the consequence
// is identical: the import is not resolved and calls to it decline. Nothing a
// Loader does can fail a parse.
type Loader func(module string) ([]byte, bool)

// DirLoader searches dirs for a module, in order, using perl's own spelling:
// `My::Mod` lives at `My/Mod.pm`.
//
// The first directory is normally the using file's own, which is what makes a
// local module -- never installed anywhere -- resolvable at all.
func DirLoader(dirs ...string) Loader {
	return func(module string) ([]byte, bool) {
		rel := filepath.Join(strings.Split(module, "::")...) + ".pm"
		for _, dir := range dirs {
			src, err := os.ReadFile(filepath.Join(dir, rel))
			if err != nil {
				// Unreadable is not found. Distinguishing them would buy a
				// caller nothing: neither can resolve the import.
				continue
			}
			return src, true
		}
		return nil, false
	}
}

// ParseFile reads path and parses it, resolving `use` against its directory.
//
// The error is the read, and only the read: no such file, no permission, a
// directory. Once bytes exist a tree exists, because the parser is total --
// syntax it cannot handle becomes Unknown carrying its text, and a bareword it
// cannot identify becomes Call{Resolved:false}.
//
// Nothing found during resolution can produce an error. perl dies on
// `use Nonexistent::Mod`; this parser must not, because parsing text is not
// running it. Were an unresolvable import a failure, the 986-file T1 ratchet
// would lose every file naming an uninstalled module.
func ParseFile(path string) (*Node, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseWithLoader(src, DirLoader(filepath.Dir(path))), nil
}

// ParseWithLoader parses src, resolving `use` through the given loader.
//
// A nil loader resolves nothing, which is exactly what Parse does.
func ParseWithLoader(src []byte, load Loader) *Node {
	return parseRoot(src, &resolver{load: load, seen: map[string]bool{}})
}

// resolver carries the loader and the visited set through one parse.
//
// `use` is recursive by nature and a cycle is ordinary: A uses B uses A. The
// visited set is keyed by module name and serves both purposes -- it
// terminates the cycle, and it stops Test::More being re-parsed once per file
// that uses it.
type resolver struct {
	load   Loader
	seen   map[string]bool
	loaded []string
}

// resolve parses a module's source once, and reports whether it was reached.
//
// Enrichment of the tree from what it finds is the next two issues:
// 01a0b84d-caa8 reads exports and prototypes, 01a0b84d-ec59 derives parse
// shape. This issue proves the source is located and handed to the parser.
func (r *resolver) resolve(module string) {
	if r == nil || r.load == nil || r.seen[module] {
		return
	}
	// Marked before parsing, not after: the module's own `use` statements are
	// read during that parse, and a cycle returns here before it can finish.
	r.seen[module] = true

	src, ok := r.load(module)
	if !ok {
		return
	}
	r.loaded = append(r.loaded, module)
	parseRoot(src, r)
}

// LoadedModules names the modules reached while parsing this tree, in the
// order they were first resolved.
//
// Empty for a tree from Parse, which resolves nothing.
func LoadedModules(root *Node) []string {
	if root == nil {
		return nil
	}
	return root.loaded
}
