// ABOUTME: Tests that IsReferenceBuild recognizes pvm's own static-libperl build
// ABOUTME: and rejects a shared-libperl one, the shape CI's installed perl reports.

package perltest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsStaticLibperl(t *testing.T) {
	assert.True(t, isStaticLibperl("false"), "useshrplib false is pvm's own static build")
	assert.True(t, isStaticLibperl("false\n"), "perl's own print leaves no trailing newline trimmed")
	assert.False(t, isStaticLibperl("true"), "useshrplib true is a shared-libperl build, not the reference one")
	assert.False(t, isStaticLibperl(""), "no answer at all is not the reference build")
}

// TestIsReferenceBuildOnPvmsOwnPerl: pvm's own perl 5.42.0 is the
// reference build TestCoreModuleRanking and TestConfigValuesFitDeclaredType
// pin their exact counts against -- confirmed by running perl -V both
// here and against CI's actions-setup-perl build (useshrplib=false here,
// useshrplib=true there, in job 39489951290's Linux log).
func TestIsReferenceBuildOnPvmsOwnPerl(t *testing.T) {
	p := Perl542(t)
	if !p.FromPvm {
		t.Skip("perl came from PATH, not pvm's own registry, so it is not the build this pins against")
	}
	assert.True(t, p.IsReferenceBuild(t))
}
