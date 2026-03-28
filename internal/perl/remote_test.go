// ABOUTME: Tests for fork remote configuration types and operations
// ABOUTME: Verifies remote CRUD, validation, and merge behavior for custom Perl fork sources

package perl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteValidation(t *testing.T) {
	tests := []struct {
		name    string
		remote  Remote
		wantErr bool
	}{
		{
			name:    "valid remote",
			remote:  Remote{Name: "mycompany", URL: "git@github.com:mycompany/perl-fork.git", Type: "git"},
			wantErr: false,
		},
		{
			name:    "type defaults to git when empty",
			remote:  Remote{Name: "mycompany", URL: "https://example.com/perl.git"},
			wantErr: false,
		},
		{
			name:    "empty name",
			remote:  Remote{Name: "", URL: "https://example.com/perl.git"},
			wantErr: true,
		},
		{
			name:    "invalid name with uppercase",
			remote:  Remote{Name: "MyCompany", URL: "https://example.com/perl.git"},
			wantErr: true,
		},
		{
			name:    "invalid name starting with hyphen",
			remote:  Remote{Name: "-company", URL: "https://example.com/perl.git"},
			wantErr: true,
		},
		{
			name:    "name origin is reserved",
			remote:  Remote{Name: "origin", URL: "https://example.com/perl.git"},
			wantErr: true,
		},
		{
			name:    "empty URL",
			remote:  Remote{Name: "mycompany", URL: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.remote.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRemoteListAdd(t *testing.T) {
	rl := NewRemoteList()

	// Add a valid remote
	err := rl.Add(Remote{Name: "mycompany", URL: "git@github.com:mycompany/perl.git"})
	require.NoError(t, err)

	// Verify it was added
	r, ok := rl.Find("mycompany")
	assert.True(t, ok)
	assert.Equal(t, "mycompany", r.Name)
	assert.Equal(t, "git@github.com:mycompany/perl.git", r.URL)
	assert.Equal(t, "git", r.Type, "type should default to git")

	// Adding duplicate should fail
	err = rl.Add(Remote{Name: "mycompany", URL: "https://other.com/perl.git"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Adding origin should fail
	err = rl.Add(Remote{Name: "origin", URL: "https://example.com/perl.git"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reserved")
}

func TestRemoteListRemove(t *testing.T) {
	rl := NewRemoteList()

	// Add then remove
	err := rl.Add(Remote{Name: "mycompany", URL: "git@github.com:mycompany/perl.git"})
	require.NoError(t, err)

	err = rl.Remove("mycompany")
	assert.NoError(t, err)

	_, ok := rl.Find("mycompany")
	assert.False(t, ok)

	// Removing nonexistent should fail
	err = rl.Remove("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Removing origin should fail
	err = rl.Remove("origin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reserved")
}

func TestRemoteListFind(t *testing.T) {
	rl := NewRemoteList()

	_ = rl.Add(Remote{Name: "company-a", URL: "https://a.com/perl.git"})
	_ = rl.Add(Remote{Name: "company-b", URL: "https://b.com/perl.git"})

	r, ok := rl.Find("company-a")
	assert.True(t, ok)
	assert.Equal(t, "https://a.com/perl.git", r.URL)

	r, ok = rl.Find("company-b")
	assert.True(t, ok)
	assert.Equal(t, "https://b.com/perl.git", r.URL)

	_, ok = rl.Find("nonexistent")
	assert.False(t, ok)
}

func TestRemoteListAll(t *testing.T) {
	rl := NewRemoteList()

	_ = rl.Add(Remote{Name: "company-a", URL: "https://a.com/perl.git"})
	_ = rl.Add(Remote{Name: "company-b", URL: "https://b.com/perl.git"})

	all := rl.All()
	assert.Len(t, all, 2)
}

func TestMergeRemoteLists(t *testing.T) {
	base := NewRemoteList()
	_ = base.Add(Remote{Name: "company-a", URL: "https://a.com/perl.git"})
	_ = base.Add(Remote{Name: "shared", URL: "https://shared-base.com/perl.git"})

	override := NewRemoteList()
	_ = override.Add(Remote{Name: "company-b", URL: "https://b.com/perl.git"})
	_ = override.Add(Remote{Name: "shared", URL: "https://shared-override.com/perl.git"})

	merged := MergeRemoteLists(base, override)

	// company-a from base should be present
	r, ok := merged.Find("company-a")
	assert.True(t, ok)
	assert.Equal(t, "https://a.com/perl.git", r.URL)

	// company-b from override should be present
	r, ok = merged.Find("company-b")
	assert.True(t, ok)
	assert.Equal(t, "https://b.com/perl.git", r.URL)

	// shared should use override's value
	r, ok = merged.Find("shared")
	assert.True(t, ok)
	assert.Equal(t, "https://shared-override.com/perl.git", r.URL)

	assert.Len(t, merged.All(), 3)
}
