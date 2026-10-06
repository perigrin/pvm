// ABOUTME: ParseFile and the loader — a `use`d module, or a `require`d .pl file, found on a search path.
// ABOUTME: Loading enriches; it never fails a parse, because an unresolved name is a result and not an error.

package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Loader answers "what is the source of this name?" -- a MODULE named as
// `My::Mod`, or a FILE named by a literal relative path as `./test.pl`.
//
// It is the seam between the parser and the filesystem. Production searches
// directories; a test hands back bytes it wrote itself, so the suite needs no
// installed module and no particular layout to be true.
//
// A third spelling, `My::Mod.xs`, asks for a module's XS source, which names
// the subs it implements in C; see XSSubs.
//
// One seam serves both spellings because the caller's question is identical --
// "bytes for this name, if you have them" -- and because the resolver's cache
// and cycle set are then shared: `./test.pl`, written 459 times across
// perl.git's `t/`, is read and parsed once per parse, the same property that
// stops Test::More being re-parsed per file that uses it.
//
// The bool is "found", not "succeeded". A name that cannot be located and one
// that cannot be read are the same answer here, because the consequence is
// identical: nothing is resolved and calls decline. Nothing a Loader does can
// fail a parse.
type Loader func(name string) ([]byte, bool)

// DirLoader searches dirs, in order, for a module or a required file.
//
// A MODULE uses perl's own spelling: `My::Mod` lives at `My/Mod.pm`. A
// required FILE is already a relative path and is joined as written --
// `./test.pl` and `./t/test.pl` are the same file reached from `t/` and from
// the tree root, which is why more than one directory is searched.
//
// The first directory is normally the using file's own, which is what makes a
// local module -- never installed anywhere -- resolvable at all.
//
// A module is also found where its source tree keeps it: `My::Mod` as
// `My-Mod/Mod.pm`, the distribution directory with the module's last part in
// it, which is where ExtUtils::MakeMaker takes a top-level .pm from --
// perl.git's ext/Devel-Peek/Peek.pm is Devel::Peek. Its XS source sits
// beside it there, `My-Mod/Mod.xs`, and is found either way.
func DirLoader(dirs ...string) Loader {
	return func(name string) ([]byte, bool) {
		rels := []string{name}
		if !isRequiredPath(name) {
			ext := ".pm"
			if module, ok := strings.CutSuffix(name, ".xs"); ok {
				name, ext = module, ".xs"
			}
			parts := strings.Split(name, "::")
			rels = []string{
				filepath.Join(parts...) + ext,
				filepath.Join(strings.Join(parts, "-"), parts[len(parts)-1]+ext),
			}
		}
		for _, dir := range dirs {
			for _, rel := range rels {
				src, err := os.ReadFile(filepath.Join(dir, rel))
				if err != nil {
					// Unreadable is not found. Distinguishing them would buy
					// a caller nothing: neither can resolve the import.
					continue
				}
				return src, true
			}
		}
		return nil, false
	}
}

// isRequiredPath reports whether a `require`'s literal argument names a FILE
// rather than a module.
//
// A `.pl` suffix is the whole rule, and `require Foo::Bar` -- the module
// spelling -- can never end in one because `.` is not legal in a bareword.
//
// Restricting to that suffix rather than accepting any path-shaped string is
// deliberate, not incidental. A `.pl` file has no `package` of its own, so it
// declares into the CALLER'S package and every `sub NAME` in it is a name the
// requiring file can call. A `.pm` reached by path is a MODULE: what it
// exports is governed by its own `@EXPORT`, and treating its every declaration
// as visible would import names perl does not. Measured over perl.git's `t/`,
// the literal requires that are NOT `.pl` are exactly that -- `bleah.pm`,
// `PACK.pm`, `unknown.pm`, `test_require.pm`, `./regen/HeaderParser.pm` -- plus
// two files that are neither, `comp/hints.aux` and `./re/regexp.t`. None of
// them is a sub-declaring helper, so the suffix is not a proxy for the rule;
// it is the rule.
func isRequiredPath(name string) bool {
	return strings.HasSuffix(name, ".pl")
}

// interpolates reports whether a quoted string's TEXT AS WRITTEN has a value
// substituted into it at runtime.
//
// Only a DOUBLE-quoted body interpolates; `'./$n.pl'` is the literal four
// characters `$n`. A `$` or `@` anywhere in a double-quoted body is treated as
// interpolation without asking what follows it, which over-refuses the
// escaped `"\$"` -- deliberately, because the direction of the error is the
// whole point: over-refusing leaves a call site Unknown exactly as it is
// today, while under-refusing resolves a path that was never named.
func interpolates(text string) bool {
	if len(text) < 2 || text[0] != '"' || text[len(text)-1] != '"' {
		return false
	}
	return strings.ContainsAny(text[1:len(text)-1], "$@")
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
	return ParseFileFrom(path, "")
}

// ParseFileFrom is ParseFile with more search directories: the roots the
// parse is rooted at, searched in order after the file's own.
//
// A required helper is named RELATIVE TO A WORKING DIRECTORY, not to the file
// that names it, and perl.git's `t/` disagrees about which directory that is.
// 459 occurrences write `require './test.pl'` and 464 of the 503 files holding
// a literal require run `chdir 't' if -d 't'` first, so the cwd is `t/` and
// `./test.pl` means `t/test.pl` even from `t/op/`. 17 write
// `require './t/test.pl'` instead, naming the same file from the tree root.
//
// So the root is what resolves the common case and the file's own directory is
// the fallback, not the other way round -- but the file's own is searched FIRST
// because a helper sitting beside its caller is unambiguous where a root-
// relative guess is not. Searching only the file's own directory moves 0 files
// and 0 nodes over the 620-file corpus; adding `t/` as the root moves 93 files
// and 3,267 nodes.
//
// MORE THAN ONE ROOT, because the suite itself runs from two working
// directories and each spelling works only from its own: the 17 files that
// write `./t/test.pl` reach the top of the tree first -- `use TestInit
// qw(T)`, or `chdir '..' if -f 'test.pl'` -- and a parse rooted at `t/` alone
// found nothing for them. The perl.git harness passes `t/` and its parent.
//
// An empty root is skipped, and with none the file's own directory is the
// only one searched, which is ParseFile.
func ParseFileFrom(path string, roots ...string) (*Node, error) {
	return ParseFileAssumingUseIf(path, false, roots...)
}

// ParseFileAssumingUseIf is ParseFileFrom with the answer to every `use if`
// condition given. `use if COND, MODULE, LIST` imports only when COND holds
// at compile time, which a parse cannot evaluate, so a file using it has two
// readings: passes, and it is `use MODULE LIST`; fails, and it defines
// nothing. ParseFileFrom takes the failing one.
func ParseFileAssumingUseIf(path string, passes bool, roots ...string) (*Node, error) {
	return parseFile(path, passes, roots, nil)
}

// Session parses many files and reads each module once: what a module says
// about itself is shared by every parse that searches the same directories.
// A corpus run parses hundreds of files using the same modules, and reading
// Test::More and everything it uses again for each file is what made the T1
// ratchet outlast go test's ten-minute timeout.
//
// Sharing is safe because a module's facts depend only on its source and the
// directories its own imports are found in. Files searching different ones --
// the file's own directory comes first, and may hold its own Foo.pm -- share
// nothing. The parses of one Session behave as one parse of all its files
// would: within a parse, a module is already read once however many uses
// name it.
//
// LoadedModules of a Session parse names only the modules that parse read,
// not those an earlier one already had. A Session is not for concurrent use.
type Session struct {
	shared map[string]*resolver
}

// NewSession returns a Session that has read nothing yet.
func NewSession() *Session {
	return &Session{shared: map[string]*resolver{}}
}

// ParseFileFrom is the package's ParseFileFrom, sharing module facts with this
// Session's other parses.
func (s *Session) ParseFileFrom(path string, roots ...string) (*Node, error) {
	return parseFile(path, false, roots, s)
}

// parseFile reads and parses path, searching its own directory and then
// roots. With a Session, the resolver's module facts and visited set are the
// ones every parse searching the same directories shares; its loaded list is
// this parse's own.
func parseFile(path string, passes bool, roots []string, s *Session) (*Node, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dirs := []string{filepath.Dir(path)}
	for _, root := range roots {
		if root != "" {
			dirs = append(dirs, root)
		}
	}
	r := &resolver{load: DirLoader(dirs...), seen: map[string]bool{}, useIfPasses: passes}
	if s != nil {
		key := strings.Join(dirs, "\x00")
		if shared, ok := s.shared[key]; ok {
			r.seen, r.facts = shared.seen, shared.facts
		} else {
			r.facts = map[string]moduleFacts{}
			s.shared[key] = r
		}
	}
	return parseRoot(src, r), nil
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
//
// A required `.pl` FILE shares the same set and cache, keyed by its path. The
// keys cannot collide: a module name is `::`-spelled and a required path ends
// in `.pl`, and `.` is not legal in a bareword. Sharing is the point --
// `./test.pl` is written 459 times across perl.git's `t/` files.
type resolver struct {
	load   Loader
	seen   map[string]bool
	loaded []string

	// declErrs are the errors in the declaration files this parse read, each
	// prefixed by its module's name.
	declErrs []error

	// facts caches what each module's source said about itself, so a module
	// used twice is read once as well as parsed once.
	facts map[string]moduleFacts

	// inRequiredFile marks a parse that IS a required `.pl` helper, so its
	// own `require` of another one is not followed. Rule 4, one level, and it
	// lives here because this is the only place that can see the depth.
	inRequiredFile bool

	// useIfPasses answers every `use if` condition true; see
	// ParseFileAssumingUseIf.
	useIfPasses bool
}

// resolve parses a module's source once and returns what it says about
// itself. The bool is whether the module was reached at all.
//
// A module already being resolved -- a cycle -- returns its facts so far,
// which is what a partially-read module can honestly offer.
func (r *resolver) resolve(module string) (moduleFacts, bool) {
	if r == nil || r.load == nil {
		return moduleFacts{}, false
	}
	if facts, ok := r.facts[module]; ok {
		return facts, true
	}
	if r.seen[module] {
		// In progress: this is a cycle. Returning "not reached" keeps the
		// recursion finite without claiming the module does not exist.
		return moduleFacts{}, false
	}
	// Marked before parsing, not after: the module's own `use` statements are
	// read during that parse, and a cycle returns here before it can finish.
	r.seen[module] = true

	// A declaration states what the module defines and is read in its
	// place, installed or not: it exists because the module's source cannot
	// show it.
	src, declared := declaration(module)
	ok := declared
	if !declared {
		src, ok = r.load(module)
	}
	if !ok {
		return moduleFacts{}, false
	}
	r.loaded = append(r.loaded, module)

	var facts moduleFacts
	switch {
	case !declared:
		facts = readModule(parseRoot(src, r))
	// CORE.pmt is the interpreter's declaration file, not a library's.
	case module == "CORE":
		facts = readDeclaration(src, r)
	default:
		facts = readLibraryDeclaration(src, r)
	}
	for _, err := range facts.errs {
		r.declErrs = append(r.declErrs, fmt.Errorf("%s: %w", module, err))
	}
	// Its XS subs are as real as the ones it declares, once it has loaded.
	// A sub the module declares in Perl keeps that declaration's prototype.
	if xs, ok := r.load(module + ".xs"); ok {
		for name, proto := range XSSubs(xs) {
			if i := strings.LastIndex(name, "::"); i > 0 && name[:i] != module &&
				!slices.Contains(facts.packages, name[:i]) {
				facts.packages = append(facts.packages, name[:i])
			}
			name = strings.TrimPrefix(name, module+"::")
			if _, declared := facts.protos[name]; !declared {
				facts.protos[name] = proto
			}
		}
	}
	if r.facts == nil {
		r.facts = map[string]moduleFacts{}
	}
	r.facts[module] = facts
	return facts, true
}

// resolveFile parses a required `.pl` helper once and returns what it
// declares. The bool is whether the file was reached at all.
//
// ONE LEVEL, which is rule 4: a helper reached from another helper returns
// "not found" rather than being read. It COSTS NOTHING, measured rather than
// assumed -- one helper does nest, `thread_it.pl` requiring `./test.pl` at its
// line 10, and all 13 `*_thr.t` files that reach it that way were already
// clean before this change and stayed clean after. They are wrappers that set
// up a thread and `require` the real test, with no parenless call site of
// their own, so a second level would buy zero files and zero nodes.
//
// The guard is here and not at the call site because this is the only place
// that knows which parse it is in.
//
// A file that cannot be read returns false and the caller records nothing.
// That is rule 5, and it is the whole reason a missing helper cannot fail a
// parse: `not found` is already the answer a missing module gets, and it has
// never been an error.
func (r *resolver) resolveFile(path string) (moduleFacts, bool) {
	if r == nil || r.load == nil || r.inRequiredFile {
		return moduleFacts{}, false
	}
	if facts, ok := r.facts[path]; ok {
		return facts, true
	}
	if r.seen[path] {
		// In progress: a cycle, which is as ordinary between two helpers as
		// between two modules. Finite recursion without claiming the file is
		// absent.
		return moduleFacts{}, false
	}
	r.seen[path] = true

	src, ok := r.load(path)
	if !ok {
		return moduleFacts{}, false
	}
	r.loaded = append(r.loaded, path)

	// The helper is parsed with a resolver that shares the cache and the
	// visited set -- so its own `use` statements still resolve, and are
	// resolved once across the whole parse -- and differs only in carrying the
	// one-level flag. The maps are allocated first so the copy shares them
	// rather than each growing one of its own.
	if r.facts == nil {
		r.facts = map[string]moduleFacts{}
	}
	inner := *r
	inner.inRequiredFile = true
	facts := readModule(parseRoot(src, &inner))

	// `loaded` and `declErrs` are SLICES, so the inner parse's appends are
	// not visible on the outer resolver and must be taken back explicitly.
	// The maps need no such handling.
	r.loaded, r.declErrs = inner.loaded, inner.declErrs

	r.facts[path] = facts
	return facts, true
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

// DeclarationErrors are the errors in the declaration files read while
// parsing this tree, each naming its module. A shipped declaration in error
// is a fault in pvm, not in the parsed source, so it is reported here and
// does not fail the parse. As with LoadedModules, a Session parse reports
// only the declarations that parse read.
//
// Empty for a tree from Parse, which reads none.
func DeclarationErrors(root *Node) []error {
	if root == nil {
		return nil
	}
	return root.declErrs
}
