// ABOUTME: Tests MetaCPAN release search request compatibility.
// ABOUTME: Ensures module version lookups use supported source filtering.

package cpan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaCPANProvider_GetModuleVersions_UsesSourceFiltering(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/release/_search", r.URL.Path)
		assert.Equal(t, "name:Test::Module", r.URL.Query().Get("q"))
		assert.Equal(t, "version,status", r.URL.Query().Get("_source"))
		assert.Empty(t, r.URL.Query().Get("fields"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"hits": {
				"hits": [
					{"_source": {"version": "2.0.0", "status": "latest"}},
					{"_source": {"version": "1.0.0", "status": "cpan"}},
					{"_source": {"version": "2.0.0", "status": "latest"}}
				]
			}
		}`))
	}))
	defer server.Close()

	provider, err := NewMetaCPANProvider()
	require.NoError(t, err)
	provider.baseURL = server.URL

	versions, err := provider.GetModuleVersions(context.Background(), "Test::Module")

	require.NoError(t, err)
	assert.Equal(t, []string{"2.0.0", "1.0.0"}, versions)
}
