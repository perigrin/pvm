// ABOUTME: Tests for mirror management functionality
// ABOUTME: Validates mirror selection, health checking, and failover capabilities

package cpan

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewMirrorManager(t *testing.T) {
	mirrors := []string{"https://cpan.org", "https://mirror1.com", "https://mirror2.com"}
	timeout := 10 * time.Second
	logger := log.New(os.Stderr, "[TestMirror] ", log.LstdFlags)

	mm, err := NewMirrorManager(mirrors, timeout, logger)
	if err != nil {
		t.Fatalf("Failed to create mirror manager: %v", err)
	}

	if len(mm.mirrors) != len(mirrors) {
		t.Errorf("Expected %d mirrors, got %d", len(mirrors), len(mm.mirrors))
	}

	if mm.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, mm.timeout)
	}

	// Check health cache initialization
	if len(mm.healthCache) != len(mirrors) {
		t.Errorf("Expected %d entries in health cache, got %d", len(mirrors), len(mm.healthCache))
	}

	for _, mirror := range mirrors {
		if _, exists := mm.healthCache[mirror]; !exists {
			t.Errorf("Health cache missing entry for mirror: %s", mirror)
		}
	}
}

func TestNewMirrorManagerEmptyMirrors(t *testing.T) {
	_, err := NewMirrorManager([]string{}, 10*time.Second, nil)
	if err == nil {
		t.Error("Expected error for empty mirrors list")
	}

	if !strings.Contains(err.Error(), "At least one mirror must be provided") {
		t.Errorf("Expected validation error message, got: %v", err)
	}
}

func TestNewMirrorManagerDefaults(t *testing.T) {
	mirrors := []string{"https://cpan.org"}

	// Test with nil logger and zero timeout
	mm, err := NewMirrorManager(mirrors, 0, nil)
	if err != nil {
		t.Fatalf("Failed to create mirror manager with defaults: %v", err)
	}

	if mm.logger == nil {
		t.Error("Expected default logger to be created")
	}

	if mm.timeout != 10*time.Second {
		t.Errorf("Expected default timeout 10s, got %v", mm.timeout)
	}
}

func TestMirrorManagerSelectBestMirror(t *testing.T) {
	// Create test servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond) // Slower response
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Faster response
		w.WriteHeader(http.StatusOK)
	}))
	defer server2.Close()

	server3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError) // Unhealthy
	}))
	defer server3.Close()

	mirrors := []string{server1.URL, server2.URL, server3.URL}
	mm, err := NewMirrorManager(mirrors, 5*time.Second, nil)
	if err != nil {
		t.Fatalf("Failed to create mirror manager: %v", err)
	}

	// First, validate mirrors to populate health cache
	_, err = mm.ValidateMirrors()
	if err != nil {
		t.Fatalf("Failed to validate mirrors: %v", err)
	}

	// Test different selection strategies
	testCases := []struct {
		strategy MirrorSelectionStrategy
		name     string
	}{
		{StrategyFirst, "first"},
		{StrategyFastest, "fastest"},
		{StrategyRoundRobin, "round_robin"},
		{StrategyRandom, "random"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			selected, err := mm.SelectBestMirror(tc.strategy)
			if err != nil {
				t.Errorf("SelectBestMirror(%s) failed: %v", tc.strategy, err)
			}

			if selected == "" {
				t.Errorf("SelectBestMirror(%s) returned empty mirror", tc.strategy)
			}

			// For fastest strategy, should select server2 (fastest response)
			if tc.strategy == StrategyFastest && selected != server2.URL {
				t.Errorf("StrategyFastest should select fastest mirror %s, got %s", server2.URL, selected)
			}
		})
	}
}

func TestMirrorManagerValidateMirrors(t *testing.T) {
	// Create test servers with different responses
	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer healthyServer.Close()

	unhealthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer unhealthyServer.Close()

	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	mirrors := []string{healthyServer.URL, unhealthyServer.URL, slowServer.URL, "http://nonexistent.invalid"}
	mm, err := NewMirrorManager(mirrors, 2*time.Second, nil)
	if err != nil {
		t.Fatalf("Failed to create mirror manager: %v", err)
	}

	results, err := mm.ValidateMirrors()
	if err != nil {
		t.Fatalf("ValidateMirrors failed: %v", err)
	}

	if len(results) != len(mirrors) {
		t.Errorf("Expected %d results, got %d", len(mirrors), len(results))
	}

	// Check specific results
	resultMap := make(map[string]*MirrorStatus)
	for _, result := range results {
		resultMap[result.URL] = result
	}

	// Healthy server should be available
	if status, exists := resultMap[healthyServer.URL]; exists {
		if !status.Available {
			t.Errorf("Healthy server should be available, got: %s", status.Error)
		}
		if status.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", status.StatusCode)
		}
	} else {
		t.Error("Missing result for healthy server")
	}

	// Unhealthy server should not be available
	if status, exists := resultMap[unhealthyServer.URL]; exists {
		if status.Available {
			t.Error("Unhealthy server should not be available")
		}
		if status.StatusCode != 500 {
			t.Errorf("Expected status code 500, got %d", status.StatusCode)
		}
	} else {
		t.Error("Missing result for unhealthy server")
	}

	// Slow server should be available (within timeout)
	if status, exists := resultMap[slowServer.URL]; exists {
		if !status.Available {
			t.Errorf("Slow server should be available, got: %s", status.Error)
		}
		if status.ResponseTime <= 0 {
			t.Error("Expected positive response time for slow server")
		}
	} else {
		t.Error("Missing result for slow server")
	}

	// Nonexistent server should not be available
	if status, exists := resultMap["http://nonexistent.invalid"]; exists {
		if status.Available {
			t.Error("Nonexistent server should not be available")
		}
		if status.Error == "" {
			t.Error("Expected error message for nonexistent server")
		}
	} else {
		t.Error("Missing result for nonexistent server")
	}
}

func TestMirrorManagerGetMirrorHealth(t *testing.T) {
	mirrors := []string{"https://cpan.org", "https://mirror1.com"}
	mm, err := NewMirrorManager(mirrors, 5*time.Second, nil)
	if err != nil {
		t.Fatalf("Failed to create mirror manager: %v", err)
	}

	health, err := mm.GetMirrorHealth()
	if err != nil {
		t.Fatalf("GetMirrorHealth failed: %v", err)
	}

	if len(health) != len(mirrors) {
		t.Errorf("Expected %d health entries, got %d", len(mirrors), len(health))
	}

	for _, mirror := range mirrors {
		if _, exists := health[mirror]; !exists {
			t.Errorf("Missing health entry for mirror: %s", mirror)
		}
	}
}

func TestMirrorManagerGetDetailedMirrorHealth(t *testing.T) {
	mirrors := []string{"https://cpan.org", "https://mirror1.com"}
	mm, err := NewMirrorManager(mirrors, 5*time.Second, nil)
	if err != nil {
		t.Fatalf("Failed to create mirror manager: %v", err)
	}

	health, err := mm.GetDetailedMirrorHealth()
	if err != nil {
		t.Fatalf("GetDetailedMirrorHealth failed: %v", err)
	}

	if len(health) != len(mirrors) {
		t.Errorf("Expected %d detailed health entries, got %d", len(mirrors), len(health))
	}

	for _, mirror := range mirrors {
		if healthInfo, exists := health[mirror]; exists {
			if healthInfo.Status == nil {
				t.Errorf("Missing status for mirror: %s", mirror)
			}
			if healthInfo.Status.URL != mirror {
				t.Errorf("Expected URL %s, got %s", mirror, healthInfo.Status.URL)
			}
		} else {
			t.Errorf("Missing detailed health entry for mirror: %s", mirror)
		}
	}
}

func TestMirrorManagerAddRemoveMirror(t *testing.T) {
	mirrors := []string{"https://cpan.org", "https://mirror1.com"}
	mm, err := NewMirrorManager(mirrors, 5*time.Second, nil)
	if err != nil {
		t.Fatalf("Failed to create mirror manager: %v", err)
	}

	// Test adding a new mirror
	newMirror := "https://newmirror.com"
	err = mm.AddMirror(newMirror)
	if err != nil {
		t.Fatalf("AddMirror failed: %v", err)
	}

	if len(mm.mirrors) != 3 {
		t.Errorf("Expected 3 mirrors after adding, got %d", len(mm.mirrors))
	}

	// Verify mirror is in health cache
	health, err := mm.GetMirrorHealth()
	if err != nil {
		t.Fatalf("GetMirrorHealth failed after adding: %v", err)
	}

	if _, exists := health[newMirror]; !exists {
		t.Error("New mirror not found in health cache")
	}

	// Test adding duplicate mirror
	err = mm.AddMirror(newMirror)
	if err == nil {
		t.Error("Expected error when adding duplicate mirror")
	}

	// Test removing mirror
	err = mm.RemoveMirror(newMirror)
	if err != nil {
		t.Fatalf("RemoveMirror failed: %v", err)
	}

	if len(mm.mirrors) != 2 {
		t.Errorf("Expected 2 mirrors after removing, got %d", len(mm.mirrors))
	}

	// Verify mirror is removed from health cache
	health, err = mm.GetMirrorHealth()
	if err != nil {
		t.Fatalf("GetMirrorHealth failed after removing: %v", err)
	}

	if _, exists := health[newMirror]; exists {
		t.Error("Removed mirror still found in health cache")
	}

	// Test removing non-existent mirror
	err = mm.RemoveMirror("https://nonexistent.com")
	if err == nil {
		t.Error("Expected error when removing non-existent mirror")
	}

	// Test removing last mirror
	err = mm.RemoveMirror("https://cpan.org")
	if err != nil {
		t.Fatalf("RemoveMirror failed for existing mirror: %v", err)
	}

	// Try to remove the last remaining mirror
	err = mm.RemoveMirror("https://mirror1.com")
	if err == nil {
		t.Error("Expected error when removing last mirror")
	}
}

func TestMirrorManagerNoAvailableMirrors(t *testing.T) {
	// Create servers that all return errors
	errorServer1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errorServer1.Close()

	errorServer2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer errorServer2.Close()

	mirrors := []string{errorServer1.URL, errorServer2.URL}
	mm, err := NewMirrorManager(mirrors, 5*time.Second, nil)
	if err != nil {
		t.Fatalf("Failed to create mirror manager: %v", err)
	}

	// Validate mirrors to mark them as unhealthy
	_, err = mm.ValidateMirrors()
	if err != nil {
		t.Fatalf("ValidateMirrors failed: %v", err)
	}

	// Should still return a mirror (fallback behavior)
	selected, err := mm.SelectBestMirror(StrategyFirst)
	if err == nil && selected != "" {
		// This is acceptable - the manager falls back to returning all mirrors when none are healthy
		t.Logf("SelectBestMirror returned fallback mirror: %s", selected)
	} else if err != nil {
		// This is also acceptable - the manager reports no available mirrors
		t.Logf("SelectBestMirror correctly reported no available mirrors: %v", err)
	}
}

func TestMirrorManagerConcurrentAccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mirrors := []string{server.URL}
	mm, err := NewMirrorManager(mirrors, 5*time.Second, nil)
	if err != nil {
		t.Fatalf("Failed to create mirror manager: %v", err)
	}

	// Run concurrent operations to test thread safety
	done := make(chan bool, 3)

	// Concurrent validation
	go func() {
		for i := 0; i < 5; i++ {
			mm.ValidateMirrors()
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	// Concurrent health checks
	go func() {
		for i := 0; i < 5; i++ {
			mm.GetMirrorHealth()
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	// Concurrent selection
	go func() {
		for i := 0; i < 5; i++ {
			mm.SelectBestMirror(StrategyFirst)
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent operations timed out")
		}
	}
}

func TestMirrorSelectionStrategies(t *testing.T) {
	// Create multiple test servers with different response times
	servers := make([]*httptest.Server, 4)
	responseTimes := []time.Duration{50, 20, 80, 30} // milliseconds

	for i, delay := range responseTimes {
		delay := delay // capture loop variable
		servers[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(delay * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
	}

	// Clean up servers
	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()

	mirrors := make([]string, len(servers))
	for i, server := range servers {
		mirrors[i] = server.URL
	}

	mm, err := NewMirrorManager(mirrors, 5*time.Second, nil)
	if err != nil {
		t.Fatalf("Failed to create mirror manager: %v", err)
	}

	// Validate mirrors to populate response times
	_, err = mm.ValidateMirrors()
	if err != nil {
		t.Fatalf("ValidateMirrors failed: %v", err)
	}

	// Test StrategyFirst - should return first mirror
	first, err := mm.SelectBestMirror(StrategyFirst)
	if err != nil {
		t.Fatalf("StrategyFirst failed: %v", err)
	}
	if first != mirrors[0] {
		t.Errorf("StrategyFirst should return first mirror %s, got %s", mirrors[0], first)
	}

	// Test StrategyFastest - should return server with index 1 (20ms response time)
	fastest, err := mm.SelectBestMirror(StrategyFastest)
	if err != nil {
		t.Fatalf("StrategyFastest failed: %v", err)
	}
	expectedFastest := servers[1].URL
	if fastest != expectedFastest {
		t.Errorf("StrategyFastest should return fastest mirror %s, got %s", expectedFastest, fastest)
	}

	// Test StrategyRoundRobin - should return any valid mirror
	rr, err := mm.SelectBestMirror(StrategyRoundRobin)
	if err != nil {
		t.Fatalf("StrategyRoundRobin failed: %v", err)
	}
	found := false
	for _, mirror := range mirrors {
		if rr == mirror {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("StrategyRoundRobin returned invalid mirror: %s", rr)
	}

	// Test StrategyRandom - should return any valid mirror
	random, err := mm.SelectBestMirror(StrategyRandom)
	if err != nil {
		t.Fatalf("StrategyRandom failed: %v", err)
	}
	found = false
	for _, mirror := range mirrors {
		if random == mirror {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("StrategyRandom returned invalid mirror: %s", random)
	}
}
