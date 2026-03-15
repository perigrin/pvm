// ABOUTME: Build caching system for Perl compilation optimization
// ABOUTME: Provides caching of configure results and build artifacts

package perl

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"tamarou.com/pvm/internal/xdg"
)

// BuildCache manages cached build information
type BuildCache struct {
	cacheDir string
	entries  map[string]*CacheEntry
	mu       sync.RWMutex
}

// CacheEntry represents a cached build
type CacheEntry struct {
	Version        string            `json:"version"`
	BuildHash      string            `json:"build_hash"`
	ConfigureFlags []string          `json:"configure_flags"`
	SystemInfo     SystemInfo        `json:"system_info"`
	BuildResult    *BuildResult      `json:"build_result"`
	Timestamp      time.Time         `json:"timestamp"`
	ConfigCache    map[string]string `json:"config_cache"`
}

// SystemInfo contains system-specific information
type SystemInfo struct {
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	Compiler    string `json:"compiler"`
	CompilerVer string `json:"compiler_version"`
	CPUCount    int    `json:"cpu_count"`
}

// NewBuildCache creates a new build cache
func NewBuildCache() (*BuildCache, error) {
	dirs, err := xdg.GetDirs()
	if err != nil {
		return nil, err
	}

	cacheDir := filepath.Join(dirs.CacheDir, "build-cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	cache := &BuildCache{
		cacheDir: cacheDir,
		entries:  make(map[string]*CacheEntry),
	}

	// Load existing cache entries
	if err := cache.loadCache(); err != nil {
		// Non-fatal, start with empty cache
		_ = err
	}

	return cache, nil
}

// GetCachedBuild retrieves a cached build if available
func (bc *BuildCache) GetCachedBuild(version string, options *BuildOptions) (*BuildResult, bool) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	hash := bc.calculateBuildHash(version, options)
	entry, exists := bc.entries[hash]
	if !exists {
		return nil, false
	}

	// Check if cache is still valid (7 days)
	if time.Since(entry.Timestamp) > 7*24*time.Hour {
		return nil, false
	}

	// Verify the installation still exists
	if _, err := os.Stat(entry.BuildResult.InstallPath); err != nil {
		return nil, false
	}

	return entry.BuildResult, true
}

// SaveBuild saves a build to the cache
func (bc *BuildCache) SaveBuild(version string, options *BuildOptions, result *BuildResult, configCache map[string]string) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	hash := bc.calculateBuildHash(version, options)

	entry := &CacheEntry{
		Version:        version,
		BuildHash:      hash,
		ConfigureFlags: options.ConfigureOptions,
		SystemInfo:     bc.getSystemInfo(),
		BuildResult:    result,
		Timestamp:      time.Now(),
		ConfigCache:    configCache,
	}

	bc.entries[hash] = entry

	// Save to disk
	return bc.saveEntry(hash, entry)
}

// GetConfigCache retrieves cached configure results
func (bc *BuildCache) GetConfigCache(version string) (map[string]string, bool) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Look for any recent build of this version
	for _, entry := range bc.entries {
		if entry.Version == version && time.Since(entry.Timestamp) < 30*24*time.Hour {
			return entry.ConfigCache, true
		}
	}

	return nil, false
}

// calculateBuildHash creates a unique hash for build configuration
func (bc *BuildCache) calculateBuildHash(version string, options *BuildOptions) string {
	h := sha256.New()

	// Include version
	h.Write([]byte(version))

	// Include configure options
	for _, opt := range options.ConfigureOptions {
		h.Write([]byte(opt))
	}

	// Include system info
	info := bc.getSystemInfo()
	h.Write([]byte(fmt.Sprintf("%s-%s-%s", info.OS, info.Arch, info.Compiler)))

	// Include install directory (different installs should not share cache)
	if options.InstallDir != "" {
		h.Write([]byte(options.InstallDir))
	}

	return hex.EncodeToString(h.Sum(nil))
}

// getSystemInfo collects current system information
func (bc *BuildCache) getSystemInfo() SystemInfo {
	info := SystemInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		CPUCount: runtime.NumCPU(),
	}

	// Detect compiler
	if compiler, version := detectCompiler(); compiler != "" {
		info.Compiler = compiler
		info.CompilerVer = version
	}

	return info
}

// detectCompiler detects the system compiler
func detectCompiler() (string, string) {
	// Try gcc first
	if out, err := exec.Command("gcc", "--version").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 0 {
			return "gcc", extractVersion(lines[0])
		}
	}

	// Try clang
	if out, err := exec.Command("clang", "--version").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 0 {
			return "clang", extractVersion(lines[0])
		}
	}

	// Windows cl.exe
	if runtime.GOOS == "windows" {
		if out, err := exec.Command("cl", "/?").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				return "cl", extractVersion(lines[0])
			}
		}
	}

	return "", ""
}

// extractVersion extracts version from compiler output
func extractVersion(line string) string {
	// Simple version extraction, could be improved
	parts := strings.Fields(line)
	for _, part := range parts {
		if strings.Contains(part, ".") && len(part) > 2 {
			// Likely a version number
			return part
		}
	}
	return "unknown"
}

// loadCache loads cache entries from disk
func (bc *BuildCache) loadCache() error {
	entries, err := os.ReadDir(bc.cacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		hash := strings.TrimSuffix(entry.Name(), ".json")
		cachePath := filepath.Join(bc.cacheDir, entry.Name())

		data, err := os.ReadFile(cachePath)
		if err != nil {
			continue
		}

		var cacheEntry CacheEntry
		if err := json.Unmarshal(data, &cacheEntry); err != nil {
			continue
		}

		bc.entries[hash] = &cacheEntry
	}

	return nil
}

// saveEntry saves a cache entry to disk
func (bc *BuildCache) saveEntry(hash string, entry *CacheEntry) error {
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	cachePath := filepath.Join(bc.cacheDir, hash+".json")
	return os.WriteFile(cachePath, data, 0644)
}

// CleanOldEntries removes cache entries older than specified duration
func (bc *BuildCache) CleanOldEntries(maxAge time.Duration) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	now := time.Now()
	var toDelete []string

	for hash, entry := range bc.entries {
		if now.Sub(entry.Timestamp) > maxAge {
			toDelete = append(toDelete, hash)
		}
	}

	for _, hash := range toDelete {
		delete(bc.entries, hash)
		cachePath := filepath.Join(bc.cacheDir, hash+".json")
		os.Remove(cachePath)
	}

	return nil
}

// BuildArtifactCache manages compiled object files for incremental builds
type BuildArtifactCache struct {
	baseDir string
	mu      sync.Mutex
}

// NewBuildArtifactCache creates a new artifact cache
func NewBuildArtifactCache(baseDir string) (*BuildArtifactCache, error) {
	artifactDir := filepath.Join(baseDir, ".build-cache")
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return nil, err
	}

	return &BuildArtifactCache{
		baseDir: artifactDir,
	}, nil
}

// SaveArtifact saves a build artifact
func (bac *BuildArtifactCache) SaveArtifact(relPath string, data []byte) error {
	bac.mu.Lock()
	defer bac.mu.Unlock()

	artifactPath := filepath.Join(bac.baseDir, relPath)
	if err := os.MkdirAll(filepath.Dir(artifactPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(artifactPath, data, 0644)
}

// GetArtifact retrieves a build artifact
func (bac *BuildArtifactCache) GetArtifact(relPath string) ([]byte, error) {
	bac.mu.Lock()
	defer bac.mu.Unlock()

	artifactPath := filepath.Join(bac.baseDir, relPath)
	return os.ReadFile(artifactPath)
}

// HasArtifact checks if an artifact exists
func (bac *BuildArtifactCache) HasArtifact(relPath string) bool {
	bac.mu.Lock()
	defer bac.mu.Unlock()

	artifactPath := filepath.Join(bac.baseDir, relPath)
	_, err := os.Stat(artifactPath)
	return err == nil
}
