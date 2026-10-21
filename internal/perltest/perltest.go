// ABOUTME: Test helper that finds pvm's perl 5.42.0 and the library directories it reports.
// ABOUTME: A test that needs that perl calls Perl542, which skips the test where it is absent.

// Package perltest locates perl 5.42.0 for tests that measure psc against
// perl itself. Tests may run perl; production psc never does.
package perltest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/platform"
)

// Perl is a perl interpreter and the library directories it reports.
type Perl struct {
	// Path is the perl executable.
	Path string
	// Privlib is $Config{privlibexp}: the pure-Perl core library.
	Privlib string
	// Archlib is $Config{archlibexp}: the architecture-specific core library.
	Archlib string
	// FromPvm reports whether Path came from pvm's own perl registry, as
	// opposed to the PATH fallback Perl542 uses where pvm has none
	// installed (CI's actions-setup-perl, Windows's Strawberry perl). A
	// test pinned to pvm's own static-libperl build checks this before
	// asserting, so it skips rather than fails on the PATH fallback.
	FromPvm bool
}

// perl542Version is perl's $], the packed form of 5.42.0, exactly as it
// prints with no formatting.
const perl542Version = "5.042000"

// Perl542 returns the perl 5.42.0 pvm has installed, with its privlib and
// archlib as that perl reports them -- or, where pvm has none installed,
// the perl on PATH that reports $] as 5.042000, as CI installs it with
// shogo82148/actions-setup-perl (Linux, macOS) or already ships it
// (Windows's Strawberry perl) rather than through pvm. It skips t where
// neither is present or the one found has no privlib directory, as on a
// CI runner without perl 5.42.0 at all; PVM_REQUIRE_PERL542=1 turns that
// skip into a failure instead, for a job (the Linux ratchets job) that a
// quiet skip would turn into a gate nothing can fail.
func Perl542(t testing.TB) Perl {
	t.Helper()
	if info, err := perl.GetVersionInfo("5.42.0"); err == nil && info != nil {
		p := locate(t, perlExecutable(info.InstallPath))
		p.FromPvm = true
		return p
	}
	if path, ok := perl542OnPath(); ok {
		return locate(t, path)
	}
	skipOrRequire(t, "perl 5.42.0 is not installed in pvm and no perl on PATH reports %s", perl542Version)
	return Perl{}
}

// perlExecutable returns the perl executable for a registered version's
// InstallPath, which is either the installation root pvm's own builds
// use (bin/perl beneath it) or, for a system perl internal/perl's
// ImportSystemPerl registered from PATH, already the directory
// containing perl itself -- the directory literally named "bin" on
// Windows's Strawberry layout. Joining "bin" onto that unconditionally
// would look for perl at bin/bin/perl.
func perlExecutable(installPath string) string {
	exe := platform.ExecutableName("perl")
	if filepath.Base(installPath) == "bin" {
		return filepath.Join(installPath, exe)
	}
	return filepath.Join(installPath, "bin", exe)
}

// perl542OnPath looks for a perl on PATH and reports whether its $] is
// 5.042000, perl's packed form of 5.42.0.
func perl542OnPath() (string, bool) {
	path, err := exec.LookPath("perl")
	if err != nil {
		return "", false
	}
	v, err := PerlVersion(path)
	if err != nil || v != perl542Version {
		return "", false
	}
	return path, true
}

// PerlVersion runs the perl at path and returns its $], perl's packed
// version string ("5.042000" for 5.42.0) -- the one fact perl542OnPath
// and internal/conformance's own perl search each need from a
// candidate path, read the same way so the T1 ratchet and a held-to-perl
// test never silently disagree about which perl "5.42.0" means.
func PerlVersion(path string) (string, error) {
	out, err := exec.Command(path, "-e", "print $]").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// skipOrRequire skips t with the given reason, or -- when
// PVM_REQUIRE_PERL542 is set -- fails it instead, so a job that sets that
// variable cannot pass by having the perl-held tests quietly skip.
func skipOrRequire(t testing.TB, format string, args ...any) {
	t.Helper()
	if os.Getenv("PVM_REQUIRE_PERL542") != "" {
		t.Fatalf("PVM_REQUIRE_PERL542 is set: "+format, args...)
	}
	t.Skipf(format, args...)
}

// locate asks the perl at path for its privlib and archlib, skipping t
// (or failing it, under PVM_REQUIRE_PERL542) where there is no perl at
// path or its privlib is not a directory.
func locate(t testing.TB, path string) Perl {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		skipOrRequire(t, "perl 5.42.0 is not at %s", path)
		return Perl{}
	}
	out, err := exec.Command(path, "-MConfig", "-e", `print "$Config{privlibexp}\n$Config{archlibexp}\n"`).Output()
	if err != nil {
		t.Fatalf("%s could not report its library directories: %v", path, err)
	}
	dirs := strings.Split(strings.TrimRight(string(out), "\r\n"), "\n")
	if len(dirs) != 2 {
		t.Fatalf("%s reported %q, not a privlib and an archlib", path, out)
	}
	p := Perl{Path: path, Privlib: strings.TrimSpace(dirs[0]), Archlib: strings.TrimSpace(dirs[1])}
	if fi, err := os.Stat(p.Privlib); err != nil || !fi.IsDir() {
		skipOrRequire(t, "perl 5.42.0's privlib %s is not a directory", p.Privlib)
		return Perl{}
	}
	return p
}

// Command is a command running p with args, in the environment the test
// runs in less PERL5LIB, PERLLIB, PERL5OPT and PERL_LOCAL_LIB_ROOT, so perl
// finds a module where an installed perl does, not where the caller's
// environment points it.
func (p Perl) Command(args ...string) *exec.Cmd {
	cmd := exec.Command(p.Path, args...)
	for _, kv := range os.Environ() {
		switch strings.SplitN(kv, "=", 2)[0] {
		case "PERL5LIB", "PERLLIB", "PERL5OPT", "PERL_LOCAL_LIB_ROOT":
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	return cmd
}
