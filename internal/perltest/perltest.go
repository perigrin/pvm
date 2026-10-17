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
}

// Perl542 returns the perl 5.42.0 pvm has installed, with its privlib and
// archlib as that perl reports them. It skips t where pvm has no perl
// 5.42.0 or that perl's privlib is not a directory, as on CI runners
// without it; it fails t where that perl is present but cannot answer.
func Perl542(t testing.TB) Perl {
	t.Helper()
	info, err := perl.GetVersionInfo("5.42.0")
	if err != nil || info == nil {
		t.Skipf("perl 5.42.0 is not installed in pvm: %v", err)
	}
	return locate(t, filepath.Join(info.InstallPath, "bin", platform.ExecutableName("perl")))
}

// locate asks the perl at path for its privlib and archlib, skipping t
// where there is no perl at path or its privlib is not a directory.
func locate(t testing.TB, path string) Perl {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Skipf("perl 5.42.0 is not at %s", path)
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
		t.Skipf("perl 5.42.0's privlib %s is not a directory", p.Privlib)
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
