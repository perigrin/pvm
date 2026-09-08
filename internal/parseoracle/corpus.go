// ABOUTME: Builds a reproducible, version-pinned corpus out of perl's own test suite.
// ABOUTME: Shim construction, interpreter/revision pinning, and stderr classification.

package parseoracle

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DefaultCorpusRoot is where a perl5 checkout conventionally lives here. The
// corpus is referenced rather than vendored: perl's tests are Artistic/GPL and
// copying them into this repo carries obligations that a reference does not.
const DefaultCorpusRoot = "~/dev/perl5"

// CorpusRoot resolves the perl5 checkout from $PERL5_CORPUS, falling back to
// DefaultCorpusRoot. It returns an error rather than panicking when the corpus
// is absent so that callers can skip: a machine with no perl5 checkout must
// still be able to run the test suite.
func CorpusRoot() (string, error) {
	root := os.Getenv("PERL5_CORPUS")
	if root == "" {
		root = DefaultCorpusRoot
	}
	if strings.HasPrefix(root, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expanding %q: %w", root, err)
		}
		root = filepath.Join(home, root[2:])
	}

	// The corpus is only a corpus if it has the tests we measure.
	if _, err := os.Stat(filepath.Join(root, "t", "test.pl")); err != nil {
		return "", fmt.Errorf("no perl5 corpus at %s (set $PERL5_CORPUS): %w", root, err)
	}
	return root, nil
}

// BuildShim assembles a tree in which perl's tests actually compile.
//
// t/test.pl:119 does `@INC = ()` and then unshifts `../lib`, so the tests only
// compile inside a tree where `../lib` is populated. PERL5LIB cannot help —
// @INC is cleared after it is read — and 498 of the 620 corpus files use this
// convention.
//
// lib/ is populated from EVERY @INC directory that is not site_perl or
// vendor_perl, which means both the pure-perl root and the architecture-
// specific root. Config.pm and every XS module's .pm half live only in the
// second; omitting it costs ~170 files and is how an earlier measurement
// reported 66.3% where the true figure is 94.2%.
func BuildShim(corpus, dest string) error {
	if err := os.MkdirAll(filepath.Join(dest, "lib"), 0o755); err != nil {
		return fmt.Errorf("creating shim lib: %w", err)
	}
	if err := copyTree(filepath.Join(corpus, "t"), filepath.Join(dest, "t")); err != nil {
		return fmt.Errorf("copying corpus t/: %w", err)
	}

	roots, err := libraryRoots()
	if err != nil {
		return err
	}
	// Later roots must not clobber earlier ones, matching `cp -rn`: the
	// architecture-specific root is listed first and holds the authoritative
	// .pm half of every XS module.
	for _, root := range roots {
		if err := copyTree(root, filepath.Join(dest, "lib")); err != nil {
			return fmt.Errorf("populating shim lib from %s: %w", root, err)
		}
	}
	return nil
}

// libraryRoots asks the interpreter itself which @INC directories to copy,
// rather than guessing paths. site_perl and vendor_perl are excluded: they
// hold locally installed modules, which would make the corpus depend on
// whatever happens to be installed on the measuring machine.
func libraryRoots() ([]string, error) {
	const probe = `for (@INC) { print "$_\n" if -d && !m{site_perl|vendor_perl} }`

	out, err := exec.Command("perl", "-e", probe).Output()
	if err != nil {
		return nil, fmt.Errorf("asking perl for @INC: %w", err)
	}

	var roots []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			roots = append(roots, line)
		}
	}
	if len(roots) < 2 {
		return nil, fmt.Errorf("expected both a pure-perl and an architecture-specific "+
			"library root in @INC, got %d: %v — a shim built from one root is missing "+
			"Config.pm and will under-report by ~170 files", len(roots), roots)
	}
	return roots, nil
}

// copyTree copies src over dst without overwriting files that already exist,
// which is `cp -rn`. Symlinks are dereferenced, since perl's library roots use
// them and the shim must stand alone.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		info, err := os.Stat(path) // Stat, not d.Info: follow symlinks.
		if err != nil {
			return nil // A dangling symlink is not a reason to abandon the tree.
		}
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if _, err := os.Stat(target); err == nil {
			return nil // First root wins, as with `cp -rn`.
		}
		return copyFile(path, target)
	})
}

// Pin records the two things a corpus measurement depends on. Both halves
// matter: the corpus here is blead 5.45 while the interpreter is 5.42.0, and
// t/op/for-many.t uses `foreach my ( \@array ) (...)`, a blead-only form that
// 5.42 genuinely cannot parse. Pinning only one half lets a parser be marked
// wrong for agreeing with its own oracle.
type Pin struct {
	Interpreter string // perl's $], e.g. "5.042000"
	Revision    string // the perl5 checkout's git revision
}

// PinMismatchError reports a drifted pin. It names both the pinned and the
// observed value, because "does not match" cannot tell you which side moved.
type PinMismatchError struct {
	Field          string
	Pinned, Actual string
}

func (e *PinMismatchError) Error() string {
	return fmt.Sprintf("corpus pin mismatch: %s is %q but the pin records %q — "+
		"re-baseline deliberately rather than measuring across the skew",
		e.Field, e.Actual, e.Pinned)
}

// Check compares a pin against the interpreter and corpus actually present.
// Both halves are checked; a pin that verifies only one is a pin that
// silently drifts.
func (p Pin) Check(interpreter, revision string) error {
	if p.Interpreter != interpreter {
		return &PinMismatchError{Field: "interpreter version", Pinned: p.Interpreter, Actual: interpreter}
	}
	if !revisionsAgree(p.Revision, revision) {
		return &PinMismatchError{Field: "corpus revision", Pinned: p.Revision, Actual: revision}
	}
	return nil
}

// revisionsAgree compares git revisions that may be abbreviated to different
// lengths, which is how they are quoted in practice.
func revisionsAgree(pinned, actual string) bool {
	if pinned == "" || actual == "" {
		return false
	}
	if len(pinned) > len(actual) {
		pinned, actual = actual, pinned
	}
	return strings.HasPrefix(actual, pinned)
}

// ReadPin loads a pin file: `key = value` lines, `#` comments.
func ReadPin(path string) (Pin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Pin{}, fmt.Errorf("reading pin: %w", err)
	}

	var pin Pin
	for _, line := range strings.Split(string(data), "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "interpreter":
			pin.Interpreter = strings.TrimSpace(value)
		case "revision":
			pin.Revision = strings.TrimSpace(value)
		}
	}

	if pin.Interpreter == "" || pin.Revision == "" {
		return Pin{}, fmt.Errorf("pin %s must record both interpreter and revision, got %+v", path, pin)
	}
	return pin, nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return nil // Unreadable source files are skipped, not fatal.
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	// 0o644 rather than the source mode: perl's installed library is
	// read-only, and a read-only shim cannot be cleaned up by t.TempDir.
	return os.WriteFile(dst, data, 0o644)
}
