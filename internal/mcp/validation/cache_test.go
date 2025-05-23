package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationCache_NewCache(t *testing.T) {
	tests := []struct {
		name      string
		sizeStr   string
		expectErr bool
	}{
		{"valid MB", "50MB", false},
		{"valid KB", "512KB", false},
		{"valid GB", "1GB", false},
		{"valid bytes", "1024", false},
		{"invalid format", "50XB", true},
		{"invalid number", "abcMB", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewValidationCache(tt.sizeStr)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, cache)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cache)
			}
		})
	}
}

func TestValidationCache_SetGet(t *testing.T) {
	cache, err := NewValidationCache("1MB")
	require.NoError(t, err)

	// Create test result
	result := &ValidationResult{
		Valid: true,
		TypeInfo: map[string]TypeInfo{
			"test": {Name: "test", Type: "Int"},
		},
	}

	// Generate key
	key := cache.GenerateKey("test code", "/project")

	// Set in cache
	cache.Set(key, result)

	// Get from cache
	cached, found := cache.Get(key)
	assert.True(t, found)
	assert.NotNil(t, cached)
	assert.Equal(t, result.Valid, cached.Valid)
	assert.Equal(t, len(result.TypeInfo), len(cached.TypeInfo))

	// Try non-existent key
	_, found = cache.Get("non-existent")
	assert.False(t, found)
}

func TestValidationCache_Eviction(t *testing.T) {
	// Create small cache
	cache, err := NewValidationCache("1KB")
	require.NoError(t, err)

	// Add multiple entries
	for i := 0; i < 100; i++ {
		result := &ValidationResult{
			Valid:  true,
			Errors: make([]ValidationError, 10), // Make it large
		}
		key := cache.GenerateKey(string(rune(i)), "/project")
		cache.Set(key, result)
	}

	// Cache should have evicted some entries
	stats := cache.GetStats()
	assert.True(t, stats.EntryCount < 100)
	assert.True(t, stats.CurrentSize <= stats.MaxSize)
}

func TestValidationCache_ProjectClear(t *testing.T) {
	cache, err := NewValidationCache("1MB")
	require.NoError(t, err)

	// Add entries for different projects
	projects := []string{"/project1", "/project2", "/project3"}
	for _, project := range projects {
		for i := 0; i < 5; i++ {
			result := &ValidationResult{Valid: true}
			key := cache.GenerateKey(string(rune(i)), project)
			cache.Set(key, result)

			// Manually update project index for testing
			if cache.projectIndex[project] == nil {
				cache.projectIndex[project] = make(map[string]bool)
			}
			cache.projectIndex[project][key] = true
		}
	}

	// Clear one project
	cache.ClearProject("/project1")

	// Check that project1 entries are gone
	stats := cache.GetStats()
	assert.Equal(t, 10, stats.EntryCount) // Only project2 and project3 remain
}

func TestValidationCache_Clear(t *testing.T) {
	cache, err := NewValidationCache("1MB")
	require.NoError(t, err)

	// Add some entries
	for i := 0; i < 10; i++ {
		result := &ValidationResult{Valid: true}
		key := cache.GenerateKey(string(rune(i)), "/project")
		cache.Set(key, result)
	}

	// Clear cache
	cache.Clear()

	// Check that cache is empty
	stats := cache.GetStats()
	assert.Equal(t, 0, stats.EntryCount)
	assert.Equal(t, int64(0), stats.CurrentSize)
}
