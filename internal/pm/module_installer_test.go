// ABOUTME: Tests for PMModuleInstaller CPAN provider integration
// ABOUTME: Verifies that module installation properly initializes a CPAN provider

package pm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPMModuleInstaller_ProviderIsSet(t *testing.T) {
	// PMModuleInstaller.InstallModule must create a CPAN provider before
	// calling the underlying installer. Previously the provider was nil,
	// producing the error "No CPAN provider specified" (PVI-4303).
	installer := &PMModuleInstaller{}

	// Call with a bogus module — we expect a failure, but the failure must
	// NOT be about a missing CPAN provider.
	err := installer.InstallModule("Nonexistent::Module::That::Does::Not::Exist", false)
	require.Error(t, err, "installing a nonexistent module should fail")
	assert.False(t,
		strings.Contains(err.Error(), "No CPAN provider specified"),
		"should not fail due to missing provider; got: %v", err)
}
