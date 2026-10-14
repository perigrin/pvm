// ABOUTME: Tests for INC, the @INC a perl builds, replicated from its Config.pm and PERL5LIB without running it.
// ABOUTME: Fixture installs under t.TempDir hold a Config.pm and a perl that is never executed.

package perl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeInstall lays out a pvm-style install under a temp prefix: bin/perl,
// which is not a program, and lib/perl5/5.42.0/<arch>/Config.pm holding
// configLines, beside the Config_heavy.pl that marks the arch directory.
func fakeInstall(t *testing.T, configLines ...string) (prefix, perlPath string) {
	t.Helper()
	prefix = t.TempDir()
	perlPath = filepath.Join(prefix, "bin", "perl")
	require.NoError(t, os.MkdirAll(filepath.Dir(perlPath), 0o755))
	require.NoError(t, os.WriteFile(perlPath, []byte("not a program\n"), 0o644))
	arch := filepath.Join(prefix, "lib", "perl5", "5.42.0", "x86_64-linux")
	require.NoError(t, os.MkdirAll(arch, 0o755))
	// Net/Config.pm is a core module named Config.pm too; it is not the one.
	require.NoError(t, os.MkdirAll(filepath.Join(prefix, "lib", "perl5", "5.42.0", "Net"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(prefix, "lib", "perl5", "5.42.0", "Net", "Config.pm"),
		[]byte("package Net::Config;\n    privlibexp => '/wrong',\n"), 0o644))
	body := "package Config;\nour %Config = (\n" + strings.Join(configLines, "\n") + "\n);\n"
	require.NoError(t, os.WriteFile(filepath.Join(arch, "Config.pm"), []byte(body), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(arch, "Config_heavy.pl"), []byte("1;\n"), 0o644))
	return prefix, perlPath
}

// TestINCFollowsPerlsOrder checks the order perl 5.42 builds @INC in, as
// `PERL5LIB=a:b:missing perl -e 'print join "\n", @INC'` prints it: each
// PERL5LIB entry preceded by its version/arch, version and arch
// subdirectories that exist, then sitearch, sitelib, vendorarch,
// vendorlib, archlib and privlib. An empty vendor entry is left out.
func TestINCFollowsPerlsOrder(t *testing.T) {
	_, perlPath := fakeInstall(t,
		"    archlibexp => '/p/lib/perl5/5.42.0/x86_64-linux',",
		"    archname => 'x86_64-linux',",
		"    privlibexp => '/p/lib/perl5/5.42.0',",
		"    sitearchexp => '/p/lib/perl5/site_perl/5.42.0/x86_64-linux',",
		"    sitelibexp => '/p/lib/perl5/site_perl/5.42.0',",
		"    vendorarchexp => '',",
		"    vendorlibexp => '/p/lib/perl5/vendor_perl/5.42.0',",
		"    version => '5.42.0',",
	)
	// An entry is kept as written, `..` and all, as perl keeps it.
	tmp := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(tmp, "x"), 0o755))
	libs := tmp + "/x/.."
	a, b, missing := libs+"/a", libs+"/b", libs+"/missing"
	for _, d := range []string{
		filepath.Join(a, "5.42.0", "x86_64-linux"), filepath.Join(a, "x86_64-linux"), b,
	} {
		require.NoError(t, os.MkdirAll(d, 0o755))
	}
	perl5lib := strings.Join([]string{a, b, missing}, string(os.PathListSeparator))

	assert.Equal(t, []string{
		a + "/5.42.0/x86_64-linux",
		a + "/5.42.0",
		a + "/x86_64-linux",
		a,
		b,
		missing,
		"/p/lib/perl5/site_perl/5.42.0/x86_64-linux",
		"/p/lib/perl5/site_perl/5.42.0",
		"/p/lib/perl5/vendor_perl/5.42.0",
		"/p/lib/perl5/5.42.0/x86_64-linux",
		"/p/lib/perl5/5.42.0",
	}, INC(perlPath, perl5lib))
}

// TestINCRelocatable checks a relocatable perl's Config.pm, whose paths
// begin `.../`, the directory the perl binary is in.
func TestINCRelocatable(t *testing.T) {
	prefix, perlPath := fakeInstall(t,
		"    archlibexp => relocate_inc('.../../lib/perl5/5.42.0/x86_64-linux'),",
		"    archname => 'x86_64-linux',",
		"    privlibexp => relocate_inc('.../../lib/perl5/5.42.0'),",
		"    sitearchexp => relocate_inc('.../../lib/perl5/site_perl/5.42.0/x86_64-linux'),",
		"    sitelibexp => relocate_inc('.../../lib/perl5/site_perl/5.42.0'),",
		"    version => '5.42.0',",
	)
	assert.Equal(t, []string{
		filepath.Join(prefix, "lib", "perl5", "site_perl", "5.42.0", "x86_64-linux"),
		filepath.Join(prefix, "lib", "perl5", "site_perl", "5.42.0"),
		filepath.Join(prefix, "lib", "perl5", "5.42.0", "x86_64-linux"),
		filepath.Join(prefix, "lib", "perl5", "5.42.0"),
	}, INC(perlPath, ""))
}

// TestINCWithoutPerl checks that with no perl, or one whose Config.pm is
// not where a pvm install keeps it, @INC is PERL5LIB alone.
func TestINCWithoutPerl(t *testing.T) {
	lib := t.TempDir()
	assert.Equal(t, []string{lib}, INC("", lib))
	assert.Empty(t, INC("", ""))
	assert.Equal(t, []string{lib}, INC(filepath.Join(t.TempDir(), "bin", "perl"), lib))
}
