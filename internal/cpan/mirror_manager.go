// ABOUTME: Mirror management functionality for CPAN operations
// ABOUTME: Provides mirror selection, health checking, and failover capabilities

package cpan

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"tamarou.com/pvm/internal/errors"
)

// MirrorManager provides mirror management operations
type MirrorManager struct {
	mirrors         []string
	timeout         time.Duration
	logger          *log.Logger
	healthCache     map[string]*MirrorHealth
	healthMutex     sync.RWMutex
	client          *http.Client
	lastHealthCheck time.Time
}

// MirrorStatus represents the status of a mirror
type MirrorStatus struct {
	URL           string        `json:"url"`
	Available     bool          `json:"available"`
	ResponseTime  time.Duration `json:"response_time"`
	LastChecked   time.Time     `json:"last_checked"`
	Error         string        `json:"error,omitempty"`
	StatusCode    int           `json:"status_code,omitempty"`
	ContentLength int64         `json:"content_length,omitempty"`
}

// MirrorHealth contains health information for a mirror
type MirrorHealth struct {
	Status       *MirrorStatus
	FailureCount int
	LastSuccess  time.Time
	LastFailure  time.Time
}

// MirrorSelectionStrategy defines how mirrors are selected
type MirrorSelectionStrategy string

const (
	// StrategyFirst uses the first available mirror
	StrategyFirst MirrorSelectionStrategy = "first"
	// StrategyFastest uses the mirror with the lowest response time
	StrategyFastest MirrorSelectionStrategy = "fastest"
	// StrategyRoundRobin rotates through available mirrors
	StrategyRoundRobin MirrorSelectionStrategy = "round_robin"
	// StrategyRandom selects a random available mirror
	StrategyRandom MirrorSelectionStrategy = "random"
)

// NewMirrorManager creates a new mirror manager instance
func NewMirrorManager(mirrors []string, timeout time.Duration, logger *log.Logger) (*MirrorManager, error) {
	if len(mirrors) == 0 {
		return nil, errors.NewSystemError("301", "At least one mirror must be provided", nil)
	}

	if logger == nil {
		logger = log.New(os.Stderr, "[MirrorManager] ", log.LstdFlags)
	}

	if timeout == 0 {
		timeout = 10 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Limit redirects
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	mm := &MirrorManager{
		mirrors:     mirrors,
		timeout:     timeout,
		logger:      logger,
		healthCache: make(map[string]*MirrorHealth),
		client:      client,
	}

	// Initialize health cache
	for _, mirror := range mirrors {
		mm.healthCache[mirror] = &MirrorHealth{
			Status: &MirrorStatus{
				URL:         mirror,
				Available:   true, // Assume available until proven otherwise
				LastChecked: time.Time{},
			},
		}
	}

	return mm, nil
}

// SelectBestMirror selects the best available mirror based on the strategy
func (mm *MirrorManager) SelectBestMirror(strategy MirrorSelectionStrategy) (string, error) {
	// Ensure health check is recent
	if time.Since(mm.lastHealthCheck) > 5*time.Minute {
		mm.logger.Printf("Health check is stale, performing quick health check")
		go mm.performQuickHealthCheck()
	}

	availableMirrors := mm.getAvailableMirrors()
	if len(availableMirrors) == 0 {
		return "", errors.NewSystemError("302", "No available mirrors found", nil)
	}

	switch strategy {
	case StrategyFastest:
		return mm.selectFastestMirror(availableMirrors), nil
	case StrategyRoundRobin:
		return mm.selectRoundRobinMirror(availableMirrors), nil
	case StrategyRandom:
		return mm.selectRandomMirror(availableMirrors), nil
	case StrategyFirst:
		fallthrough
	default:
		return availableMirrors[0], nil
	}
}

// ValidateMirrors checks the health of all configured mirrors
func (mm *MirrorManager) ValidateMirrors() ([]*MirrorStatus, error) {
	mm.logger.Printf("Validating %d mirrors", len(mm.mirrors))

	results := make([]*MirrorStatus, 0, len(mm.mirrors))
	var wg sync.WaitGroup
	resultsChan := make(chan *MirrorStatus, len(mm.mirrors))

	ctx, cancel := context.WithTimeout(context.Background(), mm.timeout*2)
	defer cancel()

	// Check all mirrors concurrently
	for _, mirror := range mm.mirrors {
		wg.Add(1)
		go func(mirrorURL string) {
			defer wg.Done()
			status := mm.checkMirrorHealth(ctx, mirrorURL)
			resultsChan <- status
		}(mirror)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for status := range resultsChan {
		results = append(results, status)
		mm.updateHealthCache(status)
	}

	mm.lastHealthCheck = time.Now()
	mm.logger.Printf("Mirror validation completed")
	return results, nil
}

// GetMirrorHealth returns health status for all mirrors
func (mm *MirrorManager) GetMirrorHealth() (map[string]bool, error) {
	mm.healthMutex.RLock()
	defer mm.healthMutex.RUnlock()

	health := make(map[string]bool)
	for mirror, healthInfo := range mm.healthCache {
		health[mirror] = healthInfo.Status.Available
	}

	return health, nil
}

// GetDetailedMirrorHealth returns detailed health information for all mirrors
func (mm *MirrorManager) GetDetailedMirrorHealth() (map[string]*MirrorHealth, error) {
	mm.healthMutex.RLock()
	defer mm.healthMutex.RUnlock()

	// Create a deep copy to avoid race conditions
	health := make(map[string]*MirrorHealth)
	for mirror, healthInfo := range mm.healthCache {
		health[mirror] = &MirrorHealth{
			Status: &MirrorStatus{
				URL:           healthInfo.Status.URL,
				Available:     healthInfo.Status.Available,
				ResponseTime:  healthInfo.Status.ResponseTime,
				LastChecked:   healthInfo.Status.LastChecked,
				Error:         healthInfo.Status.Error,
				StatusCode:    healthInfo.Status.StatusCode,
				ContentLength: healthInfo.Status.ContentLength,
			},
			FailureCount: healthInfo.FailureCount,
			LastSuccess:  healthInfo.LastSuccess,
			LastFailure:  healthInfo.LastFailure,
		}
	}

	return health, nil
}

// AddMirror adds a new mirror to the manager
func (mm *MirrorManager) AddMirror(mirrorURL string) error {
	// Check if mirror already exists
	for _, existing := range mm.mirrors {
		if existing == mirrorURL {
			return errors.NewSystemError("303", "Mirror already exists", nil)
		}
	}

	mm.mirrors = append(mm.mirrors, mirrorURL)

	// Initialize health cache for new mirror
	mm.healthMutex.Lock()
	mm.healthCache[mirrorURL] = &MirrorHealth{
		Status: &MirrorStatus{
			URL:         mirrorURL,
			Available:   true,
			LastChecked: time.Time{},
		},
	}
	mm.healthMutex.Unlock()

	mm.logger.Printf("Added mirror: %s", mirrorURL)
	return nil
}

// RemoveMirror removes a mirror from the manager
func (mm *MirrorManager) RemoveMirror(mirrorURL string) error {
	if len(mm.mirrors) <= 1 {
		return errors.NewSystemError("304", "Cannot remove last mirror", nil)
	}

	// Find and remove mirror
	for i, mirror := range mm.mirrors {
		if mirror == mirrorURL {
			mm.mirrors = append(mm.mirrors[:i], mm.mirrors[i+1:]...)

			// Remove from health cache
			mm.healthMutex.Lock()
			delete(mm.healthCache, mirrorURL)
			mm.healthMutex.Unlock()

			mm.logger.Printf("Removed mirror: %s", mirrorURL)
			return nil
		}
	}

	return errors.NewSystemError("305", "Mirror not found", nil)
}

// checkMirrorHealth performs a health check on a single mirror
func (mm *MirrorManager) checkMirrorHealth(ctx context.Context, mirrorURL string) *MirrorStatus {
	status := &MirrorStatus{
		URL:         mirrorURL,
		LastChecked: time.Now(),
	}

	start := time.Now()

	// Create a simple health check request (try to get the root or a well-known endpoint)
	req, err := http.NewRequestWithContext(ctx, "HEAD", mirrorURL, nil)
	if err != nil {
		status.Error = fmt.Sprintf("Failed to create request: %v", err)
		status.Available = false
		return status
	}

	resp, err := mm.client.Do(req)
	if err != nil {
		status.Error = fmt.Sprintf("Request failed: %v", err)
		status.Available = false
		return status
	}
	defer resp.Body.Close()

	status.ResponseTime = time.Since(start)
	status.StatusCode = resp.StatusCode
	status.ContentLength = resp.ContentLength

	// Consider 2xx and 3xx status codes as healthy
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		status.Available = true
	} else {
		status.Available = false
		status.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return status
}

// updateHealthCache updates the health cache with new status information
func (mm *MirrorManager) updateHealthCache(status *MirrorStatus) {
	mm.healthMutex.Lock()
	defer mm.healthMutex.Unlock()

	health, exists := mm.healthCache[status.URL]
	if !exists {
		health = &MirrorHealth{
			Status: status,
		}
		mm.healthCache[status.URL] = health
	} else {
		health.Status = status
	}

	if status.Available {
		health.LastSuccess = status.LastChecked
	} else {
		health.FailureCount++
		health.LastFailure = status.LastChecked
	}
}

// getAvailableMirrors returns a list of currently available mirrors
func (mm *MirrorManager) getAvailableMirrors() []string {
	mm.healthMutex.RLock()
	defer mm.healthMutex.RUnlock()

	available := make([]string, 0, len(mm.mirrors))
	for _, mirror := range mm.mirrors {
		if health, exists := mm.healthCache[mirror]; exists && health.Status.Available {
			available = append(available, mirror)
		}
	}

	// If no mirrors are marked as available, return all mirrors as a fallback
	if len(available) == 0 {
		return mm.mirrors
	}

	return available
}

// selectFastestMirror selects the mirror with the lowest response time
func (mm *MirrorManager) selectFastestMirror(availableMirrors []string) string {
	if len(availableMirrors) == 1 {
		return availableMirrors[0]
	}

	type mirrorTime struct {
		url  string
		time time.Duration
	}

	times := make([]mirrorTime, 0, len(availableMirrors))

	mm.healthMutex.RLock()
	for _, mirror := range availableMirrors {
		if health, exists := mm.healthCache[mirror]; exists {
			times = append(times, mirrorTime{
				url:  mirror,
				time: health.Status.ResponseTime,
			})
		}
	}
	mm.healthMutex.RUnlock()

	if len(times) == 0 {
		return availableMirrors[0]
	}

	// Sort by response time
	sort.Slice(times, func(i, j int) bool {
		return times[i].time < times[j].time
	})

	return times[0].url
}

// selectRoundRobinMirror selects the next mirror in rotation
func (mm *MirrorManager) selectRoundRobinMirror(availableMirrors []string) string {
	// Simple round-robin based on current time
	index := int(time.Now().Unix()) % len(availableMirrors)
	return availableMirrors[index]
}

// selectRandomMirror selects a random available mirror
func (mm *MirrorManager) selectRandomMirror(availableMirrors []string) string {
	// Simple random selection based on current time
	index := int(time.Now().UnixNano()) % len(availableMirrors)
	return availableMirrors[index]
}

// performQuickHealthCheck performs a lightweight health check
func (mm *MirrorManager) performQuickHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, mirror := range mm.mirrors {
		status := mm.checkMirrorHealth(ctx, mirror)
		mm.updateHealthCache(status)
	}

	mm.lastHealthCheck = time.Now()
}
