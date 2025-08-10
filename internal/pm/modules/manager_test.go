// ABOUTME: Tests for the module manager
// ABOUTME: Tests the module management functionality

package modules

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareVersions(t *testing.T) {
	testCases := []struct {
		v1       string
		v2       string
		expected int
	}{
		// Equal versions
		{"1.0.0", "1.0.0", 0},
		{"v1.0.0", "1.0.0", 0},
		{"1.0.0", "v1.0.0", 0},

		// Greater than
		{"1.0.1", "1.0.0", 1},
		{"1.1.0", "1.0.0", 1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "0.9.9", 1},
		{"1.0.0", "0.1.0", 1},

		// Less than
		{"1.0.0", "1.0.1", -1},
		{"1.0.0", "1.1.0", -1},
		{"1.0.0", "2.0.0", -1},
		{"0.9.9", "1.0.0", -1},
		{"0.1.0", "1.0.0", -1},

		// Different length versions
		{"1.0", "1.0.0", 0},
		{"1", "1.0.0", 0},
		{"1.0.0.0", "1.0.0", 0},
		{"1.0.1", "1.0", 1},
		{"1.0", "1.0.1", -1},

		// Dev versions (ignoring suffix for now)
		{"1.0.0_01", "1.0.0", 0},
		{"1.0.0_01", "1.0.0_02", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.v1+"_vs_"+tc.v2, func(t *testing.T) {
			result := compareVersions(tc.v1, tc.v2)
			assert.Equal(t, tc.expected, result, "Expected compareVersions(%s, %s) to be %d, got %d",
				tc.v1, tc.v2, tc.expected, result)
		})
	}
}

func TestModuleBundleInfo(t *testing.T) {
	// Test serialization and deserialization of bundle info
	bundle := &ModuleBundleInfo{
		Name:        "Test Bundle",
		Description: "Test bundle for testing",
		PerlVersion: "5.36.0",
		Modules: []*ModuleBundleEntry{
			{
				Name:              "Test::Module1",
				VersionConstraint: ">=1.0.0",
			},
			{
				Name:              "Test::Module2",
				VersionConstraint: ">=2.0.0",
				IsDev:             true,
			},
			{
				Name:              "Test::Module3",
				VersionConstraint: ">=3.0.0",
				IsOptional:        true,
			},
		},
	}

	// Convert to JSON
	data, err := json.Marshal(bundle)
	assert.NoError(t, err, "Should marshal to JSON without error")
	assert.NotEmpty(t, data, "JSON data should not be empty")

	// Convert back from JSON
	var newBundle ModuleBundleInfo
	err = json.Unmarshal(data, &newBundle)
	assert.NoError(t, err, "Should unmarshal from JSON without error")

	// Check equal values
	assert.Equal(t, bundle.Name, newBundle.Name, "Bundle name should match")
	assert.Equal(t, bundle.Description, newBundle.Description, "Bundle description should match")
	assert.Equal(t, bundle.PerlVersion, newBundle.PerlVersion, "Perl version should match")
	assert.Equal(t, len(bundle.Modules), len(newBundle.Modules), "Number of modules should match")

	// Check modules
	for i, mod := range bundle.Modules {
		assert.Equal(t, mod.Name, newBundle.Modules[i].Name, "Module name should match")
		assert.Equal(t, mod.VersionConstraint, newBundle.Modules[i].VersionConstraint, "Version constraint should match")
		assert.Equal(t, mod.IsDev, newBundle.Modules[i].IsDev, "IsDev flag should match")
		assert.Equal(t, mod.IsOptional, newBundle.Modules[i].IsOptional, "IsOptional flag should match")
	}
}
