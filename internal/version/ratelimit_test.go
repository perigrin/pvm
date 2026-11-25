package version

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestRateLimitFailsFastUnderTest guards the retry loop's test detection:
// under `go test` a rate-limited 403 must fail at once, never sleep until the
// limit resets. On Windows the old os.Args check missed "x.test.exe", and
// TestReleaseNotesCommand_ErrorHandling hung for ten minutes.
func TestRateLimitFailsFastUnderTest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Reset", "9999999999")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	}))
	defer srv.Close()

	g := NewGitHubClient()
	g.baseURL = srv.URL
	start := time.Now()
	_, err := g.GetReleases("perigrin", "pvm", true)
	if err == nil || !strings.Contains(err.Error(), "test environment") {
		t.Fatalf("want a fail-fast test-environment rate-limit error, got %v", err)
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Fatalf("rate-limited request took %v; it waited for the reset", d)
	}
}
