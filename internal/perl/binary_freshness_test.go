// ABOUTME: Tests that a cached binary is re-validated against the remote asset
// ABOUTME: size before reuse, so a republished (changed) binary is not served stale.

package perl

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// cachedBinaryIsFresh compares a cached file's size against the remote asset's
// Content-Length. A republished binary that changed size must be treated as
// stale so the cache is not served after the upstream asset changes (#470).
func TestCachedBinaryIsFresh(t *testing.T) {
	tests := []struct {
		name       string
		remoteSize int64
		localSize  int64
		wantFresh  bool
	}{
		{"same size is fresh", 1000, 1000, true},
		{"different size is stale", 1000, 800, false},
		{"republished larger is stale", 46912261, 40676216, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", strconv.FormatInt(tt.remoteSize, 10))
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			got := cachedBinaryIsFresh(srv.URL, tt.localSize)
			if got != tt.wantFresh {
				t.Errorf("cachedBinaryIsFresh(remote=%d, local=%d) = %v, want %v",
					tt.remoteSize, tt.localSize, got, tt.wantFresh)
			}
		})
	}
}

// On a HEAD failure (offline, server error) freshness cannot be determined, so
// the cache should be trusted rather than breaking an offline install.
func TestCachedBinaryIsFresh_HeadFailureTrustsCache(t *testing.T) {
	// Point at a closed server so the HEAD request fails.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	if !cachedBinaryIsFresh(url, 1234) {
		t.Error("cachedBinaryIsFresh should trust the cache when the HEAD request fails")
	}
}

// When the server responds without a usable size — a non-200 status, or a 200
// with no Content-Length — freshness cannot be determined, so the cache is
// trusted (a rate-limited/private mirror must not cause a needless re-download,
// and must not be treated as "changed").
func TestCachedBinaryIsFresh_NoUsableSizeTrustsCache(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{
			name: "non-200 status",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "rate limited", http.StatusTooManyRequests)
			},
		},
		{
			name: "200 without content-length",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Chunked response: no Content-Length, so ContentLength is -1.
				w.Header().Set("Transfer-Encoding", "chunked")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("x"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			if !cachedBinaryIsFresh(srv.URL, 1234) {
				t.Error("cachedBinaryIsFresh should trust the cache when no usable remote size is available")
			}
		})
	}
}
