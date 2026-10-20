// ABOUTME: Recognizes the specific perl 5.42.0 build a measured count was pinned against.
// ABOUTME: An exact count (an import total, a %Config undef count) holds only on that build.

package perltest

import (
	"strings"
	"testing"
)

// IsReferenceBuild reports whether p is the perl 5.42.0 pvm's own
// perl-build process produces, measured by $Config{useshrplib}: pvm
// builds it with a static libperl.a (useshrplib false), where a perl CI
// installs instead links libperl as a shared library (measured:
// shogo82148/actions-setup-perl's Linux and macOS builds, and Windows's
// Strawberry perl, all report useshrplib true).
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
	return isStaticLibperl(string(out))
}

// isStaticLibperl is the decision IsReferenceBuild makes from
// $Config{useshrplib}'s text, pulled out so it is testable without
// running perl.
func isStaticLibperl(useshrplib string) bool {
	return strings.TrimSpace(useshrplib) == "false"
}
