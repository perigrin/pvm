// ABOUTME: CPAN integration for package manager support
// ABOUTME: Provides integration with cpanm, carton, and CPAN metadata with full cpanfile/snapshot parsing

package cpan

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Integration provides CPAN package manager integration
type Integration struct {
	// MetaCPANClient for accessing CPAN metadata
	MetaCPANClient *MetaCPANClient

	// LocalPackageManager detects and uses local package managers
	LocalPackageManager PackageManager

	// ModuleCache caches module information
	ModuleCache *ModuleCache

	// Config contains integration configuration
	Config *IntegrationConfig
}

// IntegrationConfig contains configuration for CPAN integration
type IntegrationConfig struct {
	// MetaCPANURL is the MetaCPAN API URL
	MetaCPANURL string

	// CacheDir is the directory for caching data
	CacheDir string

	// CacheTTL is the cache time-to-live
	CacheTTL time.Duration

	// UseLocalOnly uses only local package information
	UseLocalOnly bool

	// Timeout for HTTP requests
	Timeout time.Duration
}

// PackageManager interface for different Perl package managers
type PackageManager interface {
	// Name returns the package manager name
	Name() string

	// IsAvailable checks if the package manager is available
	IsAvailable() bool

	// GetInstalledModules returns all installed modules
	GetInstalledModules() ([]ModuleInfo, error)

	// GetModuleInfo returns information about a specific module
	GetModuleInfo(module string) (*ModuleInfo, error)

	// GetDependencies returns module dependencies
	GetDependencies(module string) ([]Dependency, error)

	// GetModulePath returns the installation path for a module
	GetModulePath(module string) (string, error)
}

// MetaCPANClient provides access to MetaCPAN API
type MetaCPANClient struct {
	// BaseURL is the MetaCPAN API base URL
	BaseURL string

	// HTTPClient is the HTTP client
	HTTPClient *http.Client

	// RateLimiter controls API request rate
	RateLimiter *RateLimiter
}

// RateLimiter controls API request rate
type RateLimiter struct {
	// RequestsPerSecond is the maximum requests per second
	RequestsPerSecond int

	// lastRequest tracks the last request time
	lastRequest time.Time
}

// ModuleCache caches module information
type ModuleCache struct {
	// CacheDir is the cache directory
	CacheDir string

	// TTL is the cache time-to-live
	TTL time.Duration

	// entries stores cached entries
	entries map[string]*CacheEntry
}

// CacheEntry represents a cached module entry
type CacheEntry struct {
	// Data is the cached data
	Data interface{}

	// CachedAt is when the entry was cached
	CachedAt time.Time

	// Key is the cache key
	Key string
}

// NewIntegration creates a new CPAN integration
func NewIntegration(config *IntegrationConfig) (*Integration, error) {
	if config == nil {
		config = &IntegrationConfig{
			MetaCPANURL:  "https://fastapi.metacpan.org/v1",
			CacheDir:     filepath.Join(os.TempDir(), "pvm-cpan-cache"),
			CacheTTL:     24 * time.Hour,
			UseLocalOnly: false,
			Timeout:      30 * time.Second,
		}
	}

	// Create cache directory
	if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	integration := &Integration{
		Config: config,
		ModuleCache: &ModuleCache{
			CacheDir: config.CacheDir,
			TTL:      config.CacheTTL,
			entries:  make(map[string]*CacheEntry),
		},
	}

	// Initialize MetaCPAN client if not local-only
	if !config.UseLocalOnly {
		integration.MetaCPANClient = &MetaCPANClient{
			BaseURL: config.MetaCPANURL,
			HTTPClient: &http.Client{
				Timeout: config.Timeout,
			},
			RateLimiter: &RateLimiter{
				RequestsPerSecond: 5, // MetaCPAN rate limit
			},
		}
	}

	// Detect and initialize local package manager
	integration.LocalPackageManager = detectPackageManager()

	return integration, nil
}

// detectPackageManager detects available package managers
func detectPackageManager() PackageManager {
	// Try cpanm first
	if cpanm := NewCPANMinus(); cpanm.IsAvailable() {
		return cpanm
	}

	// Try carton
	if carton := NewCarton(); carton.IsAvailable() {
		return carton
	}

	// Try system perl
	return NewSystemPerl()
}

// GetInstalledModules returns all installed CPAN modules
func (i *Integration) GetInstalledModules() ([]ModuleInfo, error) {
	if i.LocalPackageManager != nil {
		return i.LocalPackageManager.GetInstalledModules()
	}
	return nil, fmt.Errorf("no package manager available")
}

// GetModuleInfo returns information about a specific module
func (i *Integration) GetModuleInfo(module string) (*ModuleInfo, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("module:%s", module)
	if cached := i.ModuleCache.Get(cacheKey); cached != nil {
		if info, ok := cached.(*ModuleInfo); ok {
			return info, nil
		}
	}

	// Try local package manager first
	if i.LocalPackageManager != nil {
		info, err := i.LocalPackageManager.GetModuleInfo(module)
		if err == nil && info != nil {
			i.ModuleCache.Set(cacheKey, info)
			return info, nil
		}
	}

	// Try MetaCPAN if enabled
	if !i.Config.UseLocalOnly && i.MetaCPANClient != nil {
		info, err := i.MetaCPANClient.GetModuleInfo(module)
		if err == nil && info != nil {
			i.ModuleCache.Set(cacheKey, info)
			return info, nil
		}
	}

	return nil, fmt.Errorf("module %s not found", module)
}

// GetDependencies returns module dependencies
func (i *Integration) GetDependencies(module string) ([]Dependency, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("deps:%s", module)
	if cached := i.ModuleCache.Get(cacheKey); cached != nil {
		if deps, ok := cached.([]Dependency); ok {
			return deps, nil
		}
	}

	// Try local package manager
	if i.LocalPackageManager != nil {
		deps, err := i.LocalPackageManager.GetDependencies(module)
		if err == nil {
			i.ModuleCache.Set(cacheKey, deps)
			return deps, nil
		}
	}

	// Try MetaCPAN
	if !i.Config.UseLocalOnly && i.MetaCPANClient != nil {
		deps, err := i.MetaCPANClient.GetDependencies(module)
		if err == nil {
			i.ModuleCache.Set(cacheKey, deps)
			return deps, nil
		}
	}

	return nil, fmt.Errorf("dependencies for %s not found", module)
}

// GetModulePath returns the installation path for a module
func (i *Integration) GetModulePath(module string) (string, error) {
	if i.LocalPackageManager != nil {
		return i.LocalPackageManager.GetModulePath(module)
	}
	return "", fmt.Errorf("no package manager available")
}

// SearchModules searches for modules matching a query
func (i *Integration) SearchModules(query string) ([]ModuleInfo, error) {
	if !i.Config.UseLocalOnly && i.MetaCPANClient != nil {
		return i.MetaCPANClient.SearchModules(query)
	}
	return nil, fmt.Errorf("search requires MetaCPAN access")
}

// CPANMinus implements PackageManager for cpanm
type CPANMinus struct {
	// cpanmPath is the path to cpanm executable
	cpanmPath string
}

// NewCPANMinus creates a new cpanm package manager
func NewCPANMinus() *CPANMinus {
	cpanmPath, _ := exec.LookPath("cpanm")
	return &CPANMinus{
		cpanmPath: cpanmPath,
	}
}

// Name returns the package manager name
func (c *CPANMinus) Name() string {
	return "cpanm"
}

// IsAvailable checks if cpanm is available
func (c *CPANMinus) IsAvailable() bool {
	return c.cpanmPath != ""
}

// GetInstalledModules returns all installed modules
func (c *CPANMinus) GetInstalledModules() ([]ModuleInfo, error) {
	// Use perl to list installed modules
	cmd := exec.Command("perl", "-e", `
		use ExtUtils::Installed;
		use JSON;
		my $inst = ExtUtils::Installed->new();
		my @modules = $inst->modules();
		my @info;
		foreach my $module (@modules) {
			eval {
				my $version = $inst->version($module) || "0";
				push @info, {
					name => $module,
					version => "$version",
				};
			};
		}
		print encode_json(\@info);
	`)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list modules: %w", err)
	}

	var modules []ModuleInfo
	if err := json.Unmarshal(output, &modules); err != nil {
		return nil, fmt.Errorf("failed to parse module list: %w", err)
	}

	return modules, nil
}

// GetModuleInfo returns information about a specific module
func (c *CPANMinus) GetModuleInfo(module string) (*ModuleInfo, error) {
	// Use perl to get module info
	cmd := exec.Command("perl", "-e", fmt.Sprintf(`
		use %s;
		use JSON;
		my $info = {
			name => '%s',
			version => $%s::VERSION || "0",
		};
		print encode_json($info);
	`, module, module, module))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get module info: %w", err)
	}

	var info ModuleInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse module info: %w", err)
	}

	// Get module path (not stored in ModuleInfo)
	// _, _ = c.GetModulePath(module)

	return &info, nil
}

// GetDependencies returns module dependencies
func (c *CPANMinus) GetDependencies(module string) ([]Dependency, error) {
	// Use cpanm --showdeps
	cmd := exec.Command(c.cpanmPath, "--showdeps", module)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}

	var deps []Dependency
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "-->") {
			deps = append(deps, Dependency{
				Name:  line,
				Type:  "requires",
				Phase: "runtime",
			})
		}
	}

	return deps, nil
}

// GetModulePath returns the installation path for a module
func (c *CPANMinus) GetModulePath(module string) (string, error) {
	// Convert module name to file path
	modulePath := strings.ReplaceAll(module, "::", "/") + ".pm"

	// Use perl to find the module
	cmd := exec.Command("perl", "-e", fmt.Sprintf(`
		foreach my $inc (@INC) {
			my $path = "$inc/%s";
			if (-f $path) {
				print $path;
				exit 0;
			}
		}
		exit 1;
	`, modulePath))

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("module %s not found", module)
	}

	return strings.TrimSpace(string(output)), nil
}

// Carton implements PackageManager for Carton
type Carton struct {
	// CartonPath is the path to carton executable
	CartonPath string

	// ProjectRoot is the project root with cpanfile
	ProjectRoot string
}

// NewCarton creates a new Carton package manager
func NewCarton() *Carton {
	cartonPath, _ := exec.LookPath("carton")
	return &Carton{
		CartonPath:  cartonPath,
		ProjectRoot: findProjectRoot(),
	}
}

// findProjectRoot finds the project root with cpanfile
func findProjectRoot() string {
	dir, _ := os.Getwd()
	for dir != "/" && dir != "" {
		if _, err := os.Stat(filepath.Join(dir, "cpanfile")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}

// Name returns the package manager name
func (c *Carton) Name() string {
	return "carton"
}

// IsAvailable checks if carton is available
func (c *Carton) IsAvailable() bool {
	return c.CartonPath != "" && c.ProjectRoot != ""
}

// GetInstalledModules returns all installed modules
func (c *Carton) GetInstalledModules() ([]ModuleInfo, error) {
	// Parse cpanfile.snapshot
	snapshotPath := filepath.Join(c.ProjectRoot, "cpanfile.snapshot")
	if _, err := os.Stat(snapshotPath); err != nil {
		return nil, fmt.Errorf("cpanfile.snapshot not found")
	}

	snapshot, err := c.ParseSnapshot(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot: %w", err)
	}

	var modules []ModuleInfo
	for name, info := range snapshot.Modules {
		modules = append(modules, ModuleInfo{
			Name:         name,
			Version:      info.Version,
			Distribution: info.Distribution,
			Path:         info.Path,
		})
	}

	return modules, nil
}

// GetModuleInfo returns information about a specific module
func (c *Carton) GetModuleInfo(module string) (*ModuleInfo, error) {
	// First check snapshot for exact version
	snapshotPath := filepath.Join(c.ProjectRoot, "cpanfile.snapshot")
	if _, err := os.Stat(snapshotPath); err == nil {
		snapshot, err := c.ParseSnapshot(snapshotPath)
		if err == nil {
			if info, exists := snapshot.Modules[module]; exists {
				return &ModuleInfo{
					Name:         module,
					Version:      info.Version,
					Distribution: info.Distribution,
					Path:         info.Path,
				}, nil
			}
		}
	}

	// Fallback to local lib search
	path, err := c.GetModulePath(module)
	if err != nil {
		return nil, fmt.Errorf("module %s not found in carton environment", module)
	}

	return &ModuleInfo{
		Name: module,
		Path: path,
	}, nil
}

// GetDependencies returns module dependencies
func (c *Carton) GetDependencies(module string) ([]Dependency, error) {
	// Parse cpanfile for declared dependencies
	cpanfile, err := c.ParseCPANFile()
	if err != nil {
		return nil, fmt.Errorf("failed to parse cpanfile: %w", err)
	}

	var deps []Dependency
	for _, req := range cpanfile.Requirements {
		if req.Module == module || module == "" {
			deps = append(deps, Dependency{
				Name:    req.Module,
				Version: req.Version,
				Type:    req.Relationship,
				Phase:   req.Phase,
			})
		}
	}

	// If no module specified, return all dependencies
	if module == "" {
		return deps, nil
	}

	// For specific module, also check snapshot for transitive deps
	snapshotPath := filepath.Join(c.ProjectRoot, "cpanfile.snapshot")
	if _, err := os.Stat(snapshotPath); err == nil {
		snapshot, err := c.ParseSnapshot(snapshotPath)
		if err == nil {
			if info, exists := snapshot.Modules[module]; exists {
				deps = append(deps, info.Dependencies...)
			}
		}
	}

	return deps, nil
}

// GetModulePath returns the installation path for a module
func (c *Carton) GetModulePath(module string) (string, error) {
	localLib := filepath.Join(c.ProjectRoot, "local", "lib", "perl5")
	if _, err := os.Stat(localLib); err != nil {
		return "", fmt.Errorf("carton local lib not found")
	}

	// Convert module name to file path
	modulePath := strings.ReplaceAll(module, "::", "/") + ".pm"

	// Search in carton's local lib directories
	searchPaths := []string{
		localLib,
		filepath.Join(localLib, "auto"),
	}

	// Also check architecture-specific paths
	archDirs, _ := filepath.Glob(filepath.Join(localLib, "*-*-*"))
	searchPaths = append(searchPaths, archDirs...)

	for _, searchPath := range searchPaths {
		fullPath := filepath.Join(searchPath, modulePath)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("module %s not found in carton local lib", module)
}

// ParseCPANFile parses the project's cpanfile
func (c *Carton) ParseCPANFile() (*CPANFile, error) {
	cpanfilePath := filepath.Join(c.ProjectRoot, "cpanfile")
	if _, err := os.Stat(cpanfilePath); err != nil {
		return nil, fmt.Errorf("cpanfile not found")
	}

	content, err := os.ReadFile(cpanfilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cpanfile: %w", err)
	}

	return ParseCPANFile(string(content))
}

// ParseSnapshot parses a cpanfile.snapshot file
func (c *Carton) ParseSnapshot(snapshotPath string) (*CPANSnapshot, error) {
	content, err := os.ReadFile(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot: %w", err)
	}

	return ParseCPANSnapshot(string(content))
}

// CPANSnapshot represents a parsed cpanfile.snapshot
type CPANSnapshot struct {
	// Modules are the locked module versions
	Modules map[string]SnapshotModule
}

// SnapshotModule represents a module in the snapshot
type SnapshotModule struct {
	Version      string
	Distribution string
	Path         string
	Dependencies []Dependency
}

// ParseCPANSnapshot parses cpanfile.snapshot content
func ParseCPANSnapshot(content string) (*CPANSnapshot, error) {
	snapshot := &CPANSnapshot{
		Modules: make(map[string]SnapshotModule),
	}

	lines := strings.Split(content, "\n")
	var currentDist string

	// State tracking for snapshot parsing
	inDistribution := false
	inProvides := false

	for _, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Distribution header: DISTRIBUTIONS
		if line == "DISTRIBUTIONS" {
			inDistribution = true
			continue
		}

		// Provides section
		if strings.HasPrefix(originalLine, "    provides:") {
			inProvides = true
			continue
		}

		// Reset provides when we hit a new section
		if inDistribution && !strings.HasPrefix(originalLine, "    ") && !strings.HasPrefix(originalLine, "  ") {
			inProvides = false
		}

		// Parse module entries in provides section
		if inProvides && strings.HasPrefix(originalLine, "      ") {
			// Module entry: "      ModuleName version"
			parts := strings.Fields(strings.TrimSpace(line))
			if len(parts) >= 2 {
				moduleName := parts[0]
				version := parts[1]

				snapshot.Modules[moduleName] = SnapshotModule{
					Version:      version,
					Distribution: currentDist,
					Dependencies: []Dependency{},
				}
			}
			continue
		}

		// Distribution name line (starts with 2 spaces)
		if inDistribution && strings.HasPrefix(originalLine, "  ") && !strings.HasPrefix(originalLine, "    ") {
			parts := strings.Fields(strings.TrimSpace(line))
			if len(parts) >= 1 {
				currentDist = parts[0]
				inProvides = false
			}
			continue
		}
	}

	return snapshot, nil
}

// SystemPerl implements PackageManager using system perl
type SystemPerl struct{}

// NewSystemPerl creates a new system perl package manager
func NewSystemPerl() *SystemPerl {
	return &SystemPerl{}
}

// Name returns the package manager name
func (s *SystemPerl) Name() string {
	return "system-perl"
}

// IsAvailable checks if system perl is available
func (s *SystemPerl) IsAvailable() bool {
	_, err := exec.LookPath("perl")
	return err == nil
}

// GetInstalledModules returns all installed modules
func (s *SystemPerl) GetInstalledModules() ([]ModuleInfo, error) {
	// Same as CPANMinus implementation
	cpanm := NewCPANMinus()
	return cpanm.GetInstalledModules()
}

// GetModuleInfo returns information about a specific module
func (s *SystemPerl) GetModuleInfo(module string) (*ModuleInfo, error) {
	cpanm := NewCPANMinus()
	return cpanm.GetModuleInfo(module)
}

// GetDependencies returns module dependencies
func (s *SystemPerl) GetDependencies(module string) ([]Dependency, error) {
	// Basic implementation - would need more sophisticated parsing
	return []Dependency{}, nil
}

// GetModulePath returns the installation path for a module
func (s *SystemPerl) GetModulePath(module string) (string, error) {
	cpanm := NewCPANMinus()
	return cpanm.GetModulePath(module)
}

// MetaCPAN API methods

// GetModuleInfo retrieves module information from MetaCPAN
func (m *MetaCPANClient) GetModuleInfo(module string) (*ModuleInfo, error) {
	// Rate limiting
	m.RateLimiter.Wait()

	url := fmt.Sprintf("%s/module/%s", m.BaseURL, module)
	resp, err := m.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MetaCPAN returned status %d", resp.StatusCode)
	}

	var result struct {
		Name         string   `json:"name"`
		Version      string   `json:"version"`
		Distribution string   `json:"distribution"`
		Author       string   `json:"author"`
		Abstract     string   `json:"abstract"`
		License      []string `json:"license"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	info := &ModuleInfo{
		Name:         result.Name,
		Version:      result.Version,
		Distribution: result.Distribution,
		Author:       result.Author,
		Abstract:     result.Abstract,
	}

	if len(result.License) > 0 {
		info.License = result.License[0]
	}

	return info, nil
}

// GetDependencies retrieves module dependencies from MetaCPAN
func (m *MetaCPANClient) GetDependencies(module string) ([]Dependency, error) {
	// Rate limiting
	m.RateLimiter.Wait()

	url := fmt.Sprintf("%s/release/_search", m.BaseURL)
	query := fmt.Sprintf(`{
		"query": {
			"match": {
				"provides": "%s"
			}
		},
		"fields": ["dependency"]
	}`, module)

	resp, err := m.HTTPClient.Post(url, "application/json", strings.NewReader(query))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response - simplified
	return []Dependency{}, nil
}

// SearchModules searches for modules on MetaCPAN
func (m *MetaCPANClient) SearchModules(query string) ([]ModuleInfo, error) {
	// Rate limiting
	m.RateLimiter.Wait()

	url := fmt.Sprintf("%s/module/_search?q=%s", m.BaseURL, query)
	resp, err := m.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse search results - simplified
	return []ModuleInfo{}, nil
}

// Wait implements rate limiting
func (r *RateLimiter) Wait() {
	if r.RequestsPerSecond <= 0 {
		return
	}

	minInterval := time.Second / time.Duration(r.RequestsPerSecond)
	elapsed := time.Since(r.lastRequest)

	if elapsed < minInterval {
		time.Sleep(minInterval - elapsed)
	}

	r.lastRequest = time.Now()
}

// Cache methods

// Get retrieves a cached entry
func (c *ModuleCache) Get(key string) interface{} {
	if entry, exists := c.entries[key]; exists {
		if time.Since(entry.CachedAt) < c.TTL {
			return entry.Data
		}
		// Expired
		delete(c.entries, key)
	}
	return nil
}

// Set stores an entry in the cache
func (c *ModuleCache) Set(key string, data interface{}) {
	c.entries[key] = &CacheEntry{
		Data:     data,
		CachedAt: time.Now(),
		Key:      key,
	}

	// Also persist to disk
	c.persistEntry(key, data)
}

// persistEntry saves cache entry to disk
func (c *ModuleCache) persistEntry(key string, data interface{}) error {
	// Sanitize key for filename
	filename := strings.ReplaceAll(key, ":", "_") + ".json"
	filepath := filepath.Join(c.CacheDir, filename)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, jsonData, 0644)
}

// LoadFromDisk loads cache entries from disk
func (c *ModuleCache) LoadFromDisk() error {
	entries, err := os.ReadDir(c.CacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			filepath := filepath.Join(c.CacheDir, entry.Name())
			data, err := os.ReadFile(filepath)
			if err != nil {
				continue
			}

			// Determine type from filename
			key := strings.TrimSuffix(entry.Name(), ".json")
			key = strings.ReplaceAll(key, "_", ":")

			if strings.HasPrefix(key, "module:") {
				var info ModuleInfo
				if err := json.Unmarshal(data, &info); err == nil {
					c.entries[key] = &CacheEntry{
						Data:     &info,
						CachedAt: time.Now(), // Reset TTL on load
						Key:      key,
					}
				}
			} else if strings.HasPrefix(key, "deps:") {
				var deps []Dependency
				if err := json.Unmarshal(data, &deps); err == nil {
					c.entries[key] = &CacheEntry{
						Data:     deps,
						CachedAt: time.Now(),
						Key:      key,
					}
				}
			}
		}
	}

	return nil
}

// Clear removes all cache entries
func (c *ModuleCache) Clear() error {
	c.entries = make(map[string]*CacheEntry)

	// Clear disk cache
	entries, err := os.ReadDir(c.CacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			os.Remove(filepath.Join(c.CacheDir, entry.Name()))
		}
	}

	return nil
}

// AnalyzeDependencyTree builds a complete dependency tree for a module
func (i *Integration) AnalyzeDependencyTree(module string, maxDepth int) (*DependencyTree, error) {
	tree := &DependencyTree{
		Module:   module,
		Children: []*DependencyTree{},
	}

	visited := make(map[string]bool)
	i.buildDependencyTree(tree, module, 0, maxDepth, visited)

	return tree, nil
}

// DependencyTree represents a module dependency tree
type DependencyTree struct {
	Module   string
	Version  string
	Children []*DependencyTree
}

// buildDependencyTree recursively builds the dependency tree
func (i *Integration) buildDependencyTree(tree *DependencyTree, module string, depth, maxDepth int, visited map[string]bool) error {
	if depth >= maxDepth || visited[module] {
		return nil
	}
	visited[module] = true

	// Get module info
	info, err := i.GetModuleInfo(module)
	if err == nil && info != nil {
		tree.Version = info.Version
	}

	// Get dependencies
	deps, err := i.GetDependencies(module)
	if err != nil {
		return err
	}

	for _, dep := range deps {
		if dep.Type == "requires" && dep.Phase == "runtime" {
			child := &DependencyTree{
				Module:   dep.Name,
				Version:  dep.Version,
				Children: []*DependencyTree{},
			}
			tree.Children = append(tree.Children, child)
			i.buildDependencyTree(child, dep.Name, depth+1, maxDepth, visited)
		}
	}

	return nil
}

// GetCPANFile parses a cpanfile if it exists
func GetCPANFile(projectRoot string) (*CPANFile, error) {
	cpanfilePath := filepath.Join(projectRoot, "cpanfile")
	if _, err := os.Stat(cpanfilePath); err != nil {
		return nil, fmt.Errorf("cpanfile not found")
	}

	content, err := os.ReadFile(cpanfilePath)
	if err != nil {
		return nil, err
	}

	return ParseCPANFile(string(content))
}

// CPANFile represents a parsed cpanfile
type CPANFile struct {
	// Requirements are the module requirements
	Requirements []Requirement

	// Features are optional features
	Features map[string][]Requirement

	// Platforms are platform-specific requirements
	Platforms map[string][]Requirement
}

// Requirement represents a module requirement
type Requirement struct {
	Module       string
	Version      string
	Phase        string
	Relationship string
}

// ParseCPANFile parses cpanfile content
func ParseCPANFile(content string) (*CPANFile, error) {
	cpanfile := &CPANFile{
		Requirements: []Requirement{},
		Features:     make(map[string][]Requirement),
		Platforms:    make(map[string][]Requirement),
	}

	// Enhanced regex patterns for different cpanfile constructs
	requiresRe := regexp.MustCompile(`(?:^|\s+)requires\s+'([^']+)'(?:\s*,\s*['"]([^'"]+)['"])?`)
	recommendsRe := regexp.MustCompile(`(?:^|\s+)recommends\s+'([^']+)'(?:\s*,\s*['"]([^'"]+)['"])?`)
	suggestsRe := regexp.MustCompile(`(?:^|\s+)suggests\s+'([^']+)'(?:\s*,\s*['"]([^'"]+)['"])?`)
	testRequiresRe := regexp.MustCompile(`(?:^|\s+)test_requires\s+'([^']+)'(?:\s*,\s*['"]([^'"]+)['"])?`)
	buildRequiresRe := regexp.MustCompile(`(?:^|\s+)build_requires\s+'([^']+)'(?:\s*,\s*['"]([^'"]+)['"])?`)

	// Feature and platform parsing
	featureRe := regexp.MustCompile(`feature\s+'([^']+)'`)
	onRe := regexp.MustCompile(`on\s+'([^']+)'`)

	var currentFeature string
	var currentPlatform string
	var currentPhase string = "runtime"

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle feature blocks
		if matches := featureRe.FindStringSubmatch(line); matches != nil {
			currentFeature = matches[1]
			if cpanfile.Features[currentFeature] == nil {
				cpanfile.Features[currentFeature] = []Requirement{}
			}
			continue
		}

		// Handle platform/phase blocks
		if matches := onRe.FindStringSubmatch(line); matches != nil {
			target := matches[1]
			// Check if it's a phase (test, build, runtime) vs platform (MSWin32, etc.)
			switch target {
			case "test", "build", "runtime":
				currentPhase = target
			default:
				currentPlatform = target
				if cpanfile.Platforms[currentPlatform] == nil {
					cpanfile.Platforms[currentPlatform] = []Requirement{}
				}
			}
			continue
		}

		// Parse different requirement types
		patterns := []struct {
			regex        *regexp.Regexp
			relationship string
		}{
			{requiresRe, "requires"},
			{recommendsRe, "recommends"},
			{suggestsRe, "suggests"},
			{testRequiresRe, "test_requires"},
			{buildRequiresRe, "build_requires"},
		}

		for _, pattern := range patterns {
			if matches := pattern.regex.FindStringSubmatch(line); matches != nil {
				req := Requirement{
					Module:       matches[1],
					Relationship: pattern.relationship,
					Phase:        currentPhase,
				}
				if len(matches) > 2 && matches[2] != "" {
					req.Version = matches[2]
				}

				// Determine phase for test_requires and build_requires
				switch pattern.relationship {
				case "test_requires":
					req.Phase = "test"
				case "build_requires":
					req.Phase = "build"
				}

				// Add to appropriate collection
				switch {
				case currentFeature != "":
					cpanfile.Features[currentFeature] = append(cpanfile.Features[currentFeature], req)
				case currentPlatform != "":
					cpanfile.Platforms[currentPlatform] = append(cpanfile.Platforms[currentPlatform], req)
				default:
					cpanfile.Requirements = append(cpanfile.Requirements, req)
				}
				break
			}
		}

		// Reset context on block end
		if strings.Contains(line, "};") {
			currentFeature = ""
			currentPlatform = ""
			currentPhase = "runtime"
		}
	}

	return cpanfile, nil
}
