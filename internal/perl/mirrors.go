// ABOUTME: Mirror support and CDN integration for Perl binary downloads
// ABOUTME: Provides multiple download sources for reliability and performance

package perl

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// Mirror configuration error codes
const (
	ErrMirrorConfigInvalid   = "601" // Invalid mirror configuration
	ErrAllMirrorsFailed      = "602" // All mirrors failed to respond
	ErrMirrorHealthCheckFail = "603" // Mirror health check failed
	ErrMirrorNotAvailable    = "604" // Mirror temporarily unavailable
)

// Mirror types
const (
	MirrorTypeGitHubReleases = "github-releases"
	MirrorTypeJSDelivr       = "jsdelivr"
	MirrorTypeCloudflareR2   = "cloudflare-r2"
	MirrorTypeDirectURL      = "direct"
)

// Default mirror configurations
var (
	DefaultMirrors = []MirrorConfig{
		{
			Name:        "GitHub Releases (Primary)",
			Type:        MirrorTypeGitHubReleases,
			BaseURL:     "https://github.com/example/pvm/releases/download",
			Priority:    1,
			Enabled:     true,
			Timeout:     30 * time.Second,
			MaxRetries:  3,
			HealthCheck: "/perl-5.38.0/perl-5.38.0-linux-amd64.tar.gz",
		},
		{
			Name:        "jsDelivr CDN",
			Type:        MirrorTypeJSDelivr,
			BaseURL:     "https://cdn.jsdelivr.net/gh/example/pvm@releases",
			Priority:    2,
			Enabled:     true,
			Timeout:     15 * time.Second,
			MaxRetries:  2,
			HealthCheck: "/perl-5.38.0/perl-5.38.0-linux-amd64.tar.gz",
		},
		{
			Name:        "Cloudflare R2 CDN",
			Type:        MirrorTypeCloudflareR2,
			BaseURL:     "https://pvm-binaries.example.com",
			Priority:    3,
			Enabled:     false, // Disabled by default until configured
			Timeout:     20 * time.Second,
			MaxRetries:  2,
			HealthCheck: "/perl-5.38.0/perl-5.38.0-linux-amd64.tar.gz",
		},
	}
)

// MirrorConfig represents a single mirror configuration
type MirrorConfig struct {
	// Human-readable name for the mirror
	Name string `json:"name" yaml:"name"`

	// Mirror type (github-releases, jsdelivr, cloudflare-r2, direct)
	Type string `json:"type" yaml:"type"`

	// Base URL for the mirror
	BaseURL string `json:"base_url" yaml:"base_url"`

	// Priority (lower numbers = higher priority)
	Priority int `json:"priority" yaml:"priority"`

	// Whether this mirror is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Request timeout for this mirror
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// Maximum number of retries for this mirror
	MaxRetries int `json:"max_retries" yaml:"max_retries"`

	// Health check path (relative to BaseURL)
	HealthCheck string `json:"health_check" yaml:"health_check"`

	// Custom headers to send with requests
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}

// MirrorHealth contains health status for a mirror
type MirrorHealth struct {
	// Mirror configuration
	Config MirrorConfig

	// Whether the mirror is currently healthy
	Healthy bool

	// Response time for the last health check
	ResponseTime time.Duration

	// Last health check timestamp
	LastCheck time.Time

	// Last error encountered (if any)
	LastError error

	// Number of consecutive failures
	ConsecutiveFailures int
}

// MirrorManager handles mirror selection and failover
type MirrorManager struct {
	mirrors      []MirrorConfig
	healthStatus map[string]*MirrorHealth
	client       *http.Client
	mutex        sync.RWMutex

	// Health check interval
	healthCheckInterval time.Duration

	// Cache for health check results
	healthCacheTTL time.Duration

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// NewMirrorManager creates a new mirror manager with default configuration
func NewMirrorManager() *MirrorManager {
	ctx, cancel := context.WithCancel(context.Background())

	mm := &MirrorManager{
		mirrors:             make([]MirrorConfig, len(DefaultMirrors)),
		healthStatus:        make(map[string]*MirrorHealth),
		client:              &http.Client{Timeout: 30 * time.Second},
		healthCheckInterval: 5 * time.Minute,
		healthCacheTTL:      1 * time.Hour,
		ctx:                 ctx,
		cancel:              cancel,
	}

	// Copy default mirrors
	copy(mm.mirrors, DefaultMirrors)

	// Initialize health status
	for _, mirror := range mm.mirrors {
		mm.healthStatus[mirror.Name] = &MirrorHealth{
			Config:  mirror,
			Healthy: true, // Assume healthy until proven otherwise
		}
	}

	// Start background health checking
	go mm.healthCheckLoop()

	return mm
}

// NewMirrorManagerWithConfig creates a mirror manager with custom configuration
func NewMirrorManagerWithConfig(mirrors []MirrorConfig) *MirrorManager {
	mm := NewMirrorManager()
	mm.SetMirrors(mirrors)
	return mm
}

// SetMirrors updates the mirror configuration
func (mm *MirrorManager) SetMirrors(mirrors []MirrorConfig) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Validate mirror configurations
	for i, mirror := range mirrors {
		if err := mm.validateMirrorConfig(mirror); err != nil {
			log.Warnf("Invalid mirror configuration for %s: %v", mirror.Name, err)
			continue
		}

		// Set defaults if not specified
		if mirror.Timeout == 0 {
			mirrors[i].Timeout = 30 * time.Second
		}
		if mirror.MaxRetries == 0 {
			mirrors[i].MaxRetries = 3
		}
	}

	mm.mirrors = mirrors

	// Reset health status
	mm.healthStatus = make(map[string]*MirrorHealth)
	for _, mirror := range mm.mirrors {
		mm.healthStatus[mirror.Name] = &MirrorHealth{
			Config:  mirror,
			Healthy: true,
		}
	}
}

// GetMirrors returns the current mirror configuration
func (mm *MirrorManager) GetMirrors() []MirrorConfig {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	result := make([]MirrorConfig, len(mm.mirrors))
	copy(result, mm.mirrors)
	return result
}

// GetBestMirror returns the best available mirror based on health and performance
func (mm *MirrorManager) GetBestMirror() (*MirrorConfig, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	// Get enabled and healthy mirrors
	var candidates []*MirrorHealth
	for _, health := range mm.healthStatus {
		if health.Config.Enabled && health.Healthy {
			candidates = append(candidates, health)
		}
	}

	if len(candidates) == 0 {
		return nil, errors.NewSystemError(ErrAllMirrorsFailed, "No healthy mirrors available", nil)
	}

	// Sort by priority (lower = better), then by response time
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Config.Priority != candidates[j].Config.Priority {
			return candidates[i].Config.Priority < candidates[j].Config.Priority
		}
		return candidates[i].ResponseTime < candidates[j].ResponseTime
	})

	return &candidates[0].Config, nil
}

// GetMirrorChain returns a prioritized list of mirrors for failover
func (mm *MirrorManager) GetMirrorChain() []MirrorConfig {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	var mirrors []MirrorConfig
	for _, health := range mm.healthStatus {
		if health.Config.Enabled {
			mirrors = append(mirrors, health.Config)
		}
	}

	// Sort by priority, then by health status and response time
	sort.Slice(mirrors, func(i, j int) bool {
		healthI := mm.healthStatus[mirrors[i].Name]
		healthJ := mm.healthStatus[mirrors[j].Name]

		// Priority first
		if mirrors[i].Priority != mirrors[j].Priority {
			return mirrors[i].Priority < mirrors[j].Priority
		}

		// Then health status
		if healthI.Healthy != healthJ.Healthy {
			return healthI.Healthy
		}

		// Finally response time
		return healthI.ResponseTime < healthJ.ResponseTime
	})

	return mirrors
}

// GenerateMirrorURL generates a URL for a specific mirror
func (mm *MirrorManager) GenerateMirrorURL(mirror MirrorConfig, version, platform string) (string, error) {
	parsedVersion, err := ParseVersion(version)
	if err != nil {
		return "", errors.NewVersionError(ErrInvalidBinaryVersion,
			fmt.Sprintf("Invalid version format for mirror URL: %s", version), err)
	}

	baseURL := strings.TrimSuffix(mirror.BaseURL, "/")

	// Determine archive extension
	archiveExt := ".tar.gz"
	if strings.HasPrefix(platform, "windows") {
		archiveExt = ".zip"
	}

	filename := fmt.Sprintf("perl-%s-%s%s", parsedVersion.String(), platform, archiveExt)

	switch mirror.Type {
	case MirrorTypeGitHubReleases:
		return fmt.Sprintf("%s/perl-%s/%s", baseURL, parsedVersion.String(), filename), nil

	case MirrorTypeJSDelivr:
		return fmt.Sprintf("%s/perl-%s/%s", baseURL, parsedVersion.String(), filename), nil

	case MirrorTypeCloudflareR2, MirrorTypeDirectURL:
		return fmt.Sprintf("%s/perl-%s/%s", baseURL, parsedVersion.String(), filename), nil

	default:
		return "", errors.NewSystemError(ErrMirrorConfigInvalid,
			fmt.Sprintf("Unknown mirror type: %s", mirror.Type), nil)
	}
}

// CheckMirrorHealth performs a health check on a specific mirror
func (mm *MirrorManager) CheckMirrorHealth(mirror MirrorConfig) *MirrorHealth {
	health := &MirrorHealth{
		Config:    mirror,
		LastCheck: time.Now(),
	}

	// Skip health check if mirror is disabled
	if !mirror.Enabled {
		health.Healthy = false
		health.LastError = fmt.Errorf("mirror is disabled")
		return health
	}

	// Perform HTTP HEAD request to health check URL
	if mirror.HealthCheck == "" {
		health.Healthy = true // Assume healthy if no health check configured
		return health
	}

	ctx, cancel := context.WithTimeout(mm.ctx, mirror.Timeout)
	defer cancel()

	healthURL := strings.TrimSuffix(mirror.BaseURL, "/") + mirror.HealthCheck

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, healthURL, nil)
	if err != nil {
		health.Healthy = false
		health.LastError = err
		return health
	}

	// Add custom headers
	for key, value := range mirror.Headers {
		req.Header.Set(key, value)
	}

	// Use mirror-specific client with timeout
	client := &http.Client{Timeout: mirror.Timeout}
	resp, err := client.Do(req)
	health.ResponseTime = time.Since(start)

	if err != nil {
		health.Healthy = false
		health.LastError = err
		return health
	}
	defer resp.Body.Close()

	// Consider 2xx and 3xx responses as healthy
	health.Healthy = resp.StatusCode < 400
	if !health.Healthy {
		health.LastError = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return health
}

// UpdateMirrorHealth updates the health status for a mirror
func (mm *MirrorManager) UpdateMirrorHealth(name string, healthy bool, responseTime time.Duration, err error) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	health, exists := mm.healthStatus[name]
	if !exists {
		return
	}

	health.Healthy = healthy
	health.ResponseTime = responseTime
	health.LastCheck = time.Now()
	health.LastError = err

	if !healthy {
		health.ConsecutiveFailures++
	} else {
		health.ConsecutiveFailures = 0
	}
}

// GetMirrorHealth returns the health status for all mirrors
func (mm *MirrorManager) GetMirrorHealth() map[string]*MirrorHealth {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	result := make(map[string]*MirrorHealth)
	for name, health := range mm.healthStatus {
		// Create a copy to avoid race conditions
		healthCopy := *health
		result[name] = &healthCopy
	}

	return result
}

// Close stops the background health checking
func (mm *MirrorManager) Close() {
	mm.cancel()
}

// healthCheckLoop runs periodic health checks on all mirrors
func (mm *MirrorManager) healthCheckLoop() {
	ticker := time.NewTicker(mm.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mm.performHealthChecks()
		case <-mm.ctx.Done():
			return
		}
	}
}

// performHealthChecks runs health checks on all mirrors
func (mm *MirrorManager) performHealthChecks() {
	mirrors := mm.GetMirrors()

	// Check mirrors in parallel
	var wg sync.WaitGroup
	for _, mirror := range mirrors {
		if !mirror.Enabled {
			continue
		}

		wg.Add(1)
		go func(m MirrorConfig) {
			defer wg.Done()

			health := mm.CheckMirrorHealth(m)
			mm.UpdateMirrorHealth(m.Name, health.Healthy, health.ResponseTime, health.LastError)

			if !health.Healthy {
				log.Warnf("Mirror %s health check failed: %v", m.Name, health.LastError)
			} else {
				log.Debugf("Mirror %s health check passed (%v)", m.Name, health.ResponseTime)
			}
		}(mirror)
	}

	wg.Wait()
}

// validateMirrorConfig validates a mirror configuration
func (mm *MirrorManager) validateMirrorConfig(mirror MirrorConfig) error {
	if mirror.Name == "" {
		return fmt.Errorf("mirror name cannot be empty")
	}

	if mirror.BaseURL == "" {
		return fmt.Errorf("mirror base URL cannot be empty")
	}

	if !strings.HasPrefix(mirror.BaseURL, "http://") && !strings.HasPrefix(mirror.BaseURL, "https://") {
		return fmt.Errorf("mirror base URL must start with http:// or https://")
	}

	validTypes := []string{MirrorTypeGitHubReleases, MirrorTypeJSDelivr, MirrorTypeCloudflareR2, MirrorTypeDirectURL}
	isValidType := false
	for _, vt := range validTypes {
		if mirror.Type == vt {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("invalid mirror type: %s", mirror.Type)
	}

	if mirror.Priority < 0 {
		return fmt.Errorf("mirror priority cannot be negative")
	}

	return nil
}

// DownloadWithMirrorFailover attempts to download using the mirror chain with automatic failover
func (mm *MirrorManager) DownloadWithMirrorFailover(version, platform string, options *BinaryDownloadOptions) (*BinaryDownloadResult, error) {
	mirrors := mm.GetMirrorChain()
	if len(mirrors) == 0 {
		return nil, errors.NewSystemError(ErrAllMirrorsFailed, "No mirrors configured", nil)
	}

	var lastErr error

	for _, mirror := range mirrors {
		log.Infof("Attempting download from mirror: %s", mirror.Name)

		// Generate URL for this mirror
		url, err := mm.GenerateMirrorURL(mirror, version, platform)
		if err != nil {
			log.Warnf("Failed to generate URL for mirror %s: %v", mirror.Name, err)
			lastErr = err
			continue
		}

		// Create a copy of options with mirror-specific settings
		mirrorOptions := *options
		mirrorOptions.RepoURL = url
		if mirrorOptions.MaxRetries == 0 {
			mirrorOptions.MaxRetries = mirror.MaxRetries
		}

		// Create context with mirror timeout
		ctx := options.Context
		if ctx == nil {
			ctx = context.Background()
		}

		ctx, cancel := context.WithTimeout(ctx, mirror.Timeout)
		mirrorOptions.Context = ctx

		// Attempt download
		start := time.Now()
		result, err := DownloadPerlBinary(&mirrorOptions)
		responseTime := time.Since(start)

		cancel() // Clean up context

		// Update mirror health based on result
		if err != nil {
			mm.UpdateMirrorHealth(mirror.Name, false, responseTime, err)
			log.Warnf("Download failed from mirror %s: %v", mirror.Name, err)
			lastErr = err
			continue
		}

		// Success
		mm.UpdateMirrorHealth(mirror.Name, true, responseTime, nil)
		log.Infof("Download succeeded from mirror: %s", mirror.Name)
		return result, nil
	}

	// All mirrors failed
	return nil, errors.NewSystemError(ErrAllMirrorsFailed,
		fmt.Sprintf("All mirrors failed, last error: %v", lastErr), lastErr)
}
