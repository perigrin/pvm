// ABOUTME: Tests MissingSubs reports exactly the names a module does not define as a callable sub.
// ABOUTME: A name the module does define, even a bareword constant installed only by name, is not reported.

package perltest

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMissingSubsReportsOnlyUndefinedNames(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	WriteModule(t, lib, "My::Subs", "package My::Subs;\nsub one { 1 }\nsub two { 2 }\n1;\n")

	missing := MissingSubs(t, p, lib, "My::Subs", []string{"one", "two", "three"})

	assert.Equal(t, []string{"three"}, slices.Sorted(maps.Keys(missing)))
	assert.NotEmpty(t, missing["three"], "a missing name's reason should not be empty")
}

// TestMissingSubsFindsANameInstalledOnlyByRequest: Exporter-based modules
// (Errno, Fcntl) generate some constant subs only once something asks
// for them by name -- measured, the same reason return_oracle.pl primes
// a name individually when a bare `use` does not materialize it (see its
// own comment). MissingSubs must prime the same way, or it would wrongly
// report such a name missing.
func TestMissingSubsFindsANameInstalledOnlyByRequest(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	WriteModule(t, lib, "My::ByRequest", `package My::ByRequest;
use Exporter 'import';
our @EXPORT_OK = qw(available);
our %constants = (available => 1);
sub import {
    my $class = shift;
    for my $name (@_) {
        if (exists $constants{$name}) {
            no strict 'refs';
            my $v = $constants{$name};
            *{caller() . "::$name"} = sub () { $v };
        }
    }
    local @_ = ($class);
    goto &Exporter::import;
}
1;
`)

	missing := MissingSubs(t, p, lib, "My::ByRequest", []string{"available"})

	assert.Empty(t, missing, "available is installed once asked for by name, not missing")
}
