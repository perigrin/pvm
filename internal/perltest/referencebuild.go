// ABOUTME: Recognizes the specific perl 5.42.0 build a measured count was pinned against.
// ABOUTME: An exact count (an import total, a %Config undef count) holds only on that build.

package perltest

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The reference build is identified by what its exact counts depend on,
// not by who built it. Measured 2026-10-26 on the build they were pinned
// against, pvm's own 5.42.0 on shin: 172 undefined %Config keys, among them
// every DBM header (i_db, i_gdbm, i_ndbm and the rest; the machine had
// none), and 766 .pm files in privlib. A pvm build on a machine with other
// C libraries differs in both -- measured in a cloud session: 163 undefined
// keys, the DBM extensions built, and 1437 core imports where the pin says
// 1427. So the build is matched by these two digests, each a sha256 of a
// newline-joined list sorted in byte order.
const (
	// referenceUndefKeysDigest hashes the sorted names of the %Config keys
	// whose value is undef, joined by "\n" with no trailing newline.
	referenceUndefKeysDigest = "da722577974b3ee90bbbf674505ea4e63a8ad7a26b5879c0a49df07fff953752"
	// referencePrivlibDigest hashes the sorted paths of privlib's .pm
	// files, relative to privlib with `/` separators, each followed by "\n".
	referencePrivlibDigest = "84f358007abade8bfa0e9e246ca8413eeb6bcca9992ad23370978e2b561a0624"
)

// IsReferenceBuild reports whether p is the perl 5.42.0 build the exact
// counts were measured on: pvm's static-libperl build ($Config{useshrplib}
// false), whose undefined %Config keys and privlib .pm files are the ones
// that build has. A perl CI installs links libperl shared (measured:
// shogo82148/actions-setup-perl's Linux and macOS builds, and Windows's
// Strawberry perl, all report useshrplib true), and a pvm build on another
// machine ships other files and detects other C headers.
//
// A test whose expected number depends on exactly which files privlib
// ships -- TestCoreModuleRanking's import counts, TestConfigValuesFitDeclaredType's
// undef count -- asserts that exact number only when IsReferenceBuild is
// true, and asserts a build-invariant property (set membership, "some
// values are undef") on every perl 5.42.0, whichever build produced it.
func (p Perl) IsReferenceBuild(t testing.TB) bool {
	t.Helper()
	out, err := p.Command("-MConfig", "-e", `print $Config{useshrplib}`).Output()
	if err != nil {
		t.Fatalf("%s could not report $Config{useshrplib}: %v", p.Path, err)
	}
	if !isStaticLibperl(string(out)) {
		return false
	}
	out, err = p.Command("-MConfig", "-e", `print "$_\n" for grep { !defined $Config{$_} } keys %Config`).Output()
	if err != nil {
		t.Fatalf("%s could not list its undefined %%Config keys: %v", p.Path, err)
	}
	if undefKeysDigest(strings.Fields(string(out))) != referenceUndefKeysDigest {
		t.Logf("%s is not the reference build: its undefined %%Config keys differ", p.Path)
		return false
	}
	paths, err := privlibModules(p.Privlib)
	if err != nil {
		t.Fatalf("listing %s's privlib %s: %v", p.Path, p.Privlib, err)
	}
	if privlibDigest(paths) != referencePrivlibDigest {
		t.Logf("%s is not the reference build: its privlib ships other .pm files", p.Path)
		return false
	}
	return true
}

// isStaticLibperl is the first decision IsReferenceBuild makes, from
// $Config{useshrplib}'s text, pulled out so it is testable without
// running perl.
func isStaticLibperl(useshrplib string) bool {
	return strings.TrimSpace(useshrplib) == "false"
}

// undefKeysDigest is the sha256 of keys sorted in byte order and joined by
// "\n", as perl's `sort` and `join "\n"` write them.
func undefKeysDigest(keys []string) string {
	sorted := append([]string(nil), keys...)
	sort.Strings(sorted)
	sum := sha256.Sum256([]byte(strings.Join(sorted, "\n")))
	return hex.EncodeToString(sum[:])
}

// privlibDigest is the sha256 of paths sorted in byte order, each followed
// by "\n", as `find . -name '*.pm' | sed 's|^\./||' | LC_ALL=C sort` writes
// them.
func privlibDigest(paths []string) string {
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)
	h := sha256.New()
	for _, p := range sorted {
		h.Write([]byte(p + "\n"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// privlibModules lists the .pm files under privlib, relative to it with
// `/` separators. Like find, it names a symlink without following it.
func privlibModules(privlib string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(privlib, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(d.Name(), ".pm") {
			rel, err := filepath.Rel(privlib, path)
			if err != nil {
				return err
			}
			paths = append(paths, filepath.ToSlash(rel))
		}
		return nil
	})
	return paths, err
}
