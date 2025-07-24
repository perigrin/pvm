// ABOUTME: MetaCPAN-specific implementation for CPAN metadata retrieval
// ABOUTME: Implements the Provider interface using the MetaCPAN API

package cpan

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	metaCPANAPIBaseURL = "https://api.metacpan.org/v1"
	metaCPANWebBaseURL = "https://metacpan.org"
	defaultUserAgent   = "PVM/1.0.0 (github.com/tamarou/pvm)"
	defaultTimeout     = 30 // seconds
)

// MetaCPANProvider implements the Provider interface using the MetaCPAN API
type MetaCPANProvider struct {
	baseURL        string
	client         *http.Client
	userAgent      string
	disableNetwork bool
	cacheDir       string
	cacheTTL       int
	mirrors        []string
	currentMirror  int
}

// MetaCPANModuleInfo represents a module entry within the module array
type MetaCPANModuleInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Indexed bool   `json:"indexed"`
}

// MetaCPANModule represents a module in the MetaCPAN API
type MetaCPANModule struct {
	Name             string               `json:"name"` // This is the filename (e.g., "JSON.pm")
	Version          string               `json:"version"`
	AuthorID         string               `json:"author"`
	Abstract         string               `json:"abstract"`
	Description      string               `json:"description"`
	Distribution     string               `json:"distribution"`
	DistVersion      string               `json:"dist_version"`
	DownloadURL      string               `json:"download_url"`
	ReleaseDate      string               `json:"date"`
	Status           string               `json:"status"`
	Module           []MetaCPANModuleInfo `json:"module"` // Array of modules in this distribution
	MetaData         MetaCPANMeta         `json:"metadata"`
	Resources        MetaCPANResources    `json:"resources"`
	Dependencies     []MetaCPANDependency `json:"dependency"`
	StatsCPANTesters MetaCPANStats        `json:"stats"`
	StatsRecent      MetaCPANStatsRecent  `json:"stats_recent"`
}

// MetaCPANMeta represents metadata in the MetaCPAN API
type MetaCPANMeta struct {
	License []string `json:"license"`
}

// MetaCPANResources represents resources in the MetaCPAN API
type MetaCPANResources struct {
	Repository MetaCPANRepository `json:"repository"`
	Bugtracker MetaCPANBugtracker `json:"bugtracker"`
	Homepage   string             `json:"homepage"`
}

// MetaCPANRepository represents a repository in the MetaCPAN API
type MetaCPANRepository struct {
	URL string `json:"url"`
	Web string `json:"web"`
}

// MetaCPANBugtracker represents a bugtracker in the MetaCPAN API
type MetaCPANBugtracker struct {
	Web string `json:"web"`
}

// MetaCPANStats represents statistics in the MetaCPAN API
type MetaCPANStats struct {
	Pass    int `json:"pass"`
	Fail    int `json:"fail"`
	NA      int `json:"na"`
	Unknown int `json:"unknown"`
}

// MetaCPANStatsRecent represents recent statistics in the MetaCPAN API
type MetaCPANStatsRecent struct {
	Downloads int `json:"downloads_30d"`
}

// MetaCPANDependency represents a dependency in the MetaCPAN API
type MetaCPANDependency struct {
	Phase    string `json:"phase"`
	Module   string `json:"module"`
	Requires string `json:"requires"`
}

// MetaCPANAuthor represents an author in the MetaCPAN API
type MetaCPANAuthor struct {
	PAUSEID      string `json:"pauseid"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Homepage     string `json:"homepage"`
	City         string `json:"city"`
	Region       string `json:"region"`
	Country      string `json:"country"`
	Metacpan_URL string `json:"metacpan_url"`
}

// MetaCPANSearchResponse represents the search response from the MetaCPAN API
type MetaCPANSearchResponse struct {
	Total int                 `json:"total"`
	Hits  []MetaCPANSearchHit `json:"hits"`
}

// MetaCPANSearchHit represents a single search result in the MetaCPAN API
type MetaCPANSearchHit struct {
	Source MetaCPANSearchSource `json:"_source"`
	Score  float64              `json:"_score"`
}

// MetaCPANSearchSource represents the source of a search result in the MetaCPAN API
type MetaCPANSearchSource struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	Distribution  string `json:"distribution"`
	Author        string `json:"author"`
	Abstract      string `json:"abstract"`
	Date          string `json:"date"`
	Status        string `json:"status"`
	Documentation string `json:"documentation"`
}

// NewMetaCPANProvider creates a new MetaCPANProvider with the given options
func NewMetaCPANProvider(options ...ProviderOption) (*MetaCPANProvider, error) {
	provider := &MetaCPANProvider{
		baseURL:   metaCPANAPIBaseURL,
		userAgent: defaultUserAgent,
		client: &http.Client{
			Timeout: time.Duration(defaultTimeout) * time.Second,
		},
		mirrors: []string{metaCPANAPIBaseURL},
	}

	// Apply options
	for _, option := range options {
		if err := option(provider); err != nil {
			return nil, err
		}
	}

	return provider, nil
}

// WithBaseURL implements the ProviderOption interface for MetaCPANProvider
func (p *MetaCPANProvider) WithBaseURL(url string) error {
	p.baseURL = url
	return nil
}

// WithUserAgent implements the ProviderOption interface for MetaCPANProvider
func (p *MetaCPANProvider) WithUserAgent(userAgent string) error {
	p.userAgent = userAgent
	return nil
}

// WithTimeout implements the ProviderOption interface for MetaCPANProvider
func (p *MetaCPANProvider) WithTimeout(timeout int) error {
	p.client.Timeout = time.Duration(timeout) * time.Second
	return nil
}

// WithCache implements the ProviderOption interface for MetaCPANProvider
func (p *MetaCPANProvider) WithCache(cacheDir string, ttl int) error {
	p.cacheDir = cacheDir
	p.cacheTTL = ttl
	return nil
}

// WithDisableNetwork implements the ProviderOption interface for MetaCPANProvider
func (p *MetaCPANProvider) WithDisableNetwork(disable bool) error {
	p.disableNetwork = disable
	return nil
}

// WithMirror implements the ProviderOption interface for MetaCPANProvider
func (p *MetaCPANProvider) WithMirror(mirror string) error {
	// Replace the default mirror with the provided one
	p.mirrors = []string{mirror}
	p.currentMirror = 0
	return nil
}

// WithAdditionalMirrors implements the ProviderOption interface for MetaCPANProvider
func (p *MetaCPANProvider) WithAdditionalMirrors(mirrors []string) error {
	// Append additional mirrors
	p.mirrors = append(p.mirrors, mirrors...)
	return nil
}

// Name returns the name of the provider
func (p *MetaCPANProvider) Name() string {
	return "metacpan"
}

// BaseURL returns the base URL of the provider's API
func (p *MetaCPANProvider) BaseURL() string {
	return p.baseURL
}

// GetModuleInfo retrieves metadata about a specific module from MetaCPAN
func (p *MetaCPANProvider) GetModuleInfo(ctx context.Context, moduleName string) (*ModuleInfo, error) {
	if p.disableNetwork {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "network_disabled",
			Message: "Network access is disabled",
		}
	}

	// Check cache first
	var cache *Cache
	var cacheKey string
	if p.cacheDir != "" {
		var err error
		cache, err = NewCache(p.cacheDir, p.cacheTTL)
		if err == nil {
			cacheKey = fmt.Sprintf("module_info_%s", moduleName)

			// Check if cache needs validation
			if !cache.NeedsValidation(cacheKey) {
				var cachedInfo ModuleInfo
				if cache.Get(cacheKey, &cachedInfo) {
					return &cachedInfo, nil
				}
			}
		}
	}

	// Prepare the URL
	// Use module name as-is for MetaCPAN API (no :: to / conversion needed)
	endpoint := fmt.Sprintf("/module/%s", url.QueryEscape(moduleName))
	requestURL := p.baseURL + endpoint

	// Make the request
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "request_creation_failed",
			Message: "Failed to create request",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Set headers
	req.Header.Set("User-Agent", p.userAgent)
	req.Header.Set("Accept", "application/json")

	// Add conditional headers if we have cached data that needs validation
	if cache != nil && cacheKey != "" {
		if conditionalHeaders := cache.GetConditionalHeaders(cacheKey); conditionalHeaders != nil {
			for key, values := range conditionalHeaders {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}
		}
	}

	// Execute the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "request_failed",
			Message: "Failed to execute request",
			URL:     requestURL,
			Err:     err,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code first
	if resp.StatusCode == http.StatusNotModified {
		// HTTP 304 Not Modified - return cached data
		if cache != nil && cacheKey != "" {
			var cachedInfo ModuleInfo
			if cache.Get(cacheKey, &cachedInfo) {
				return &cachedInfo, nil
			}
		}

		// If we can't get cached data for some reason, treat as error
		return nil, &ProviderError{
			Source:     p.Name(),
			Code:       "cache_miss_on_304",
			Message:    "Received 304 Not Modified but no cached data available",
			URL:        requestURL,
			StatusCode: resp.StatusCode,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &ProviderError{
			Source:     p.Name(),
			Code:       "http_error",
			Message:    fmt.Sprintf("HTTP error: %d", resp.StatusCode),
			URL:        requestURL,
			StatusCode: resp.StatusCode,
		}
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "read_failed",
			Message: "Failed to read response body",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Parse the JSON response
	var module MetaCPANModule
	if err := json.Unmarshal(body, &module); err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "parse_failed",
			Message: "Failed to parse JSON response",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Debug: log the parsed module info

	// Get the primary module name from the module array
	var actualModuleName string
	var actualModuleVersion string
	if len(module.Module) > 0 {
		// Use the first indexed module
		for _, mod := range module.Module {
			if mod.Indexed {
				actualModuleName = mod.Name
				actualModuleVersion = mod.Version
				break
			}
		}
		// Fallback to first module if none are indexed
		if actualModuleName == "" {
			actualModuleName = module.Module[0].Name
			actualModuleVersion = module.Module[0].Version
		}
	}

	// Fallback to distribution name if no modules found
	if actualModuleName == "" {
		actualModuleName = module.Distribution
	}

	// Use distribution version if module version is empty
	if actualModuleVersion == "" {
		actualModuleVersion = module.Version
	}

	// Convert to a generic ModuleInfo
	info := &ModuleInfo{
		Name:                actualModuleName,    // Use actual module name, not filename
		Version:             actualModuleVersion, // Use module version
		Author:              module.AuthorID,
		Description:         module.Description,
		Abstract:            module.Abstract,
		Distribution:        module.Distribution,
		DistributionVersion: module.Version, // Use top-level version for distribution
		DistributionFile:    module.DownloadURL,
		Documentation:       fmt.Sprintf("%s/pod/%s", metaCPANWebBaseURL, actualModuleName),
		Downloads:           module.StatsRecent.Downloads,
		Rating:              calculateRating(&module),
		HasTests:            module.StatsCPANTesters.Pass > 0,
	}

	// Parse release date
	if module.ReleaseDate != "" {
		releaseDate, err := time.Parse(time.RFC3339, module.ReleaseDate)
		if err == nil {
			info.ReleaseDate = releaseDate
		}
	}

	// Set URLs
	if module.Resources.Repository.URL != "" {
		info.Repository = module.Resources.Repository.URL
	} else if module.Resources.Repository.Web != "" {
		info.Repository = module.Resources.Repository.Web
	}

	if module.Resources.Homepage != "" {
		info.Homepage = module.Resources.Homepage
	}

	if module.Resources.Bugtracker.Web != "" {
		info.Bugtracker = module.Resources.Bugtracker.Web
	}

	// Set license
	if len(module.MetaData.License) > 0 {
		info.License = strings.Join(module.MetaData.License, ", ")
	}

	// Convert dependencies
	if len(module.Dependencies) > 0 {
		info.Dependencies = make([]*Dependency, 0, len(module.Dependencies))
		for _, dep := range module.Dependencies {
			if dep.Module == "" {
				continue
			}
			info.Dependencies = append(info.Dependencies, &Dependency{
				Name:    dep.Module,
				Version: dep.Requires,
				Phase:   dep.Phase,
				Type:    "requires",
			})
		}
	}

	// Save to cache with HTTP headers
	if cache != nil && cacheKey != "" {
		_ = cache.SetWithHTTPHeaders(cacheKey, info, p.Name(), resp.Header, requestURL)
	}

	return info, nil
}

// SearchModules searches for modules matching the given query using the MetaCPAN API
func (p *MetaCPANProvider) SearchModules(ctx context.Context, query string, limit int) (*SearchResults, error) {
	if p.disableNetwork {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "network_disabled",
			Message: "Network access is disabled",
		}
	}

	// Check cache first
	if p.cacheDir != "" {
		cache, err := NewCache(p.cacheDir, p.cacheTTL)
		if err == nil {
			cacheKey := fmt.Sprintf("search_%s_%d", query, limit)
			var cachedResults SearchResults
			if cache.Get(cacheKey, &cachedResults) {
				return &cachedResults, nil
			}
		}
	}

	// Set a default limit if not provided
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100 // MetaCPAN has a maximum limit of 5000, but we'll use a smaller one by default
	}

	// Prepare the URL
	endpoint := fmt.Sprintf("/search/module?q=%s&size=%d", url.QueryEscape(query), limit)
	requestURL := p.baseURL + endpoint

	// Make the request
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "request_creation_failed",
			Message: "Failed to create request",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Set headers
	req.Header.Set("User-Agent", p.userAgent)
	req.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "request_failed",
			Message: "Failed to execute request",
			URL:     requestURL,
			Err:     err,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, &ProviderError{
			Source:     p.Name(),
			Code:       "http_error",
			Message:    fmt.Sprintf("HTTP error: %d", resp.StatusCode),
			URL:        requestURL,
			StatusCode: resp.StatusCode,
		}
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "read_failed",
			Message: "Failed to read response body",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Parse the JSON response
	var searchResponse MetaCPANSearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "parse_failed",
			Message: "Failed to parse JSON response",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Convert to a generic SearchResults
	results := &SearchResults{
		Query:   query,
		Total:   searchResponse.Total,
		Results: make([]*SearchResult, 0, len(searchResponse.Hits)),
		Source:  p.Name(),
	}

	for _, hit := range searchResponse.Hits {
		source := hit.Source
		result := &SearchResult{
			Name:             source.Name,
			Version:          source.Version,
			Distribution:     source.Distribution,
			Author:           source.Author,
			Abstract:         source.Abstract,
			Score:            hit.Score,
			IsLatest:         source.Status == "latest",
			HasDocumentation: source.Documentation != "",
		}

		// Parse release date
		if source.Date != "" {
			releaseDate, err := time.Parse(time.RFC3339, source.Date)
			if err == nil {
				result.ReleaseDate = releaseDate
			}
		}

		results.Results = append(results.Results, result)
	}

	// Save to cache
	if p.cacheDir != "" {
		cache, err := NewCache(p.cacheDir, p.cacheTTL)
		if err == nil {
			cacheKey := fmt.Sprintf("search_%s_%d", query, limit)
			_ = cache.Set(cacheKey, results, p.Name())
		}
	}

	return results, nil
}

// GetDependencies retrieves the dependencies for a module using the MetaCPAN API
func (p *MetaCPANProvider) GetDependencies(ctx context.Context, moduleName string) ([]*Dependency, error) {
	// This is a shortcut method that uses GetModuleInfo internally
	info, err := p.GetModuleInfo(ctx, moduleName)
	if err != nil {
		return nil, err
	}

	return info.Dependencies, nil
}

// GetModuleVersions retrieves all available versions of a module using the MetaCPAN API
func (p *MetaCPANProvider) GetModuleVersions(ctx context.Context, moduleName string) ([]string, error) {
	if p.disableNetwork {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "network_disabled",
			Message: "Network access is disabled",
		}
	}

	// Check cache first
	if p.cacheDir != "" {
		cache, err := NewCache(p.cacheDir, p.cacheTTL)
		if err == nil {
			cacheKey := fmt.Sprintf("versions_%s", moduleName)
			var cachedVersions []string
			if cache.Get(cacheKey, &cachedVersions) {
				return cachedVersions, nil
			}
		}
	}

	// Query MetaCPAN for all releases of this module
	endpoint := fmt.Sprintf("/release/_search?q=name:%s&fields=version,status&size=100&sort=version:desc", url.QueryEscape(moduleName))
	requestURL := p.baseURL + endpoint

	// Make the request
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "request_creation_failed",
			Message: "Failed to create request",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Set headers
	req.Header.Set("User-Agent", p.userAgent)
	req.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "request_failed",
			Message: "Failed to execute request",
			URL:     requestURL,
			Err:     err,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, &ProviderError{
			Source:     p.Name(),
			Code:       "http_error",
			Message:    fmt.Sprintf("HTTP error: %d", resp.StatusCode),
			URL:        requestURL,
			StatusCode: resp.StatusCode,
		}
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "read_failed",
			Message: "Failed to read response body",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Parse the JSON response
	var searchResponse struct {
		Hits struct {
			Hits []struct {
				Source struct {
					Version string `json:"version"`
					Status  string `json:"status"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "parse_failed",
			Message: "Failed to parse JSON response",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Extract versions
	versions := make([]string, 0, len(searchResponse.Hits.Hits))
	seen := make(map[string]bool)

	for _, hit := range searchResponse.Hits.Hits {
		version := hit.Source.Version
		// Only include unique versions and skip developer versions unless explicitly requested
		if version != "" && !seen[version] {
			seen[version] = true
			versions = append(versions, version)
		}
	}

	// Save to cache
	if p.cacheDir != "" {
		cache, err := NewCache(p.cacheDir, p.cacheTTL)
		if err == nil {
			cacheKey := fmt.Sprintf("versions_%s", moduleName)
			_ = cache.Set(cacheKey, versions, p.Name())
		}
	}

	return versions, nil
}

// GetAuthorInfo retrieves information about a CPAN author using the MetaCPAN API
func (p *MetaCPANProvider) GetAuthorInfo(ctx context.Context, authorID string) (map[string]interface{}, error) {
	if p.disableNetwork {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "network_disabled",
			Message: "Network access is disabled",
		}
	}

	// Check cache first
	if p.cacheDir != "" {
		cache, err := NewCache(p.cacheDir, p.cacheTTL)
		if err == nil {
			cacheKey := fmt.Sprintf("author_%s", authorID)
			var cachedAuthor map[string]interface{}
			if cache.Get(cacheKey, &cachedAuthor) {
				return cachedAuthor, nil
			}
		}
	}

	// Prepare the URL
	endpoint := fmt.Sprintf("/author/%s", url.PathEscape(authorID))
	requestURL := p.baseURL + endpoint

	// Make the request
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "request_creation_failed",
			Message: "Failed to create request",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Set headers
	req.Header.Set("User-Agent", p.userAgent)
	req.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "request_failed",
			Message: "Failed to execute request",
			URL:     requestURL,
			Err:     err,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, &ProviderError{
			Source:     p.Name(),
			Code:       "http_error",
			Message:    fmt.Sprintf("HTTP error: %d", resp.StatusCode),
			URL:        requestURL,
			StatusCode: resp.StatusCode,
		}
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "read_failed",
			Message: "Failed to read response body",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Parse the JSON response
	var author MetaCPANAuthor
	if err := json.Unmarshal(body, &author); err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "parse_failed",
			Message: "Failed to parse JSON response",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Convert to a map for the interface
	authorInfo := map[string]interface{}{
		"pauseid": author.PAUSEID,
		"name":    author.Name,
		"email":   author.Email,
	}

	if author.Homepage != "" {
		authorInfo["homepage"] = author.Homepage
	}

	if author.City != "" {
		authorInfo["city"] = author.City
	}

	if author.Region != "" {
		authorInfo["region"] = author.Region
	}

	if author.Country != "" {
		authorInfo["country"] = author.Country
	}

	if author.Metacpan_URL != "" {
		authorInfo["metacpan_url"] = author.Metacpan_URL
	}

	// Save to cache
	if p.cacheDir != "" {
		cache, err := NewCache(p.cacheDir, p.cacheTTL)
		if err == nil {
			cacheKey := fmt.Sprintf("author_%s", authorID)
			_ = cache.Set(cacheKey, authorInfo, p.Name())
		}
	}

	return authorInfo, nil
}

// IsCoreModule checks if a module is part of the Perl core
func (p *MetaCPANProvider) IsCoreModule(ctx context.Context, moduleName string, perlVersion string) (bool, error) {
	// Basic implementation that checks common core modules
	// A full implementation would query the MetaCPAN API or use Module::CoreList
	coreModules := map[string]bool{
		"strict":          true,
		"warnings":        true,
		"Carp":            true,
		"File::Spec":      true,
		"Data::Dumper":    true,
		"Getopt::Long":    true,
		"Pod::Usage":      true,
		"Scalar::Util":    true,
		"List::Util":      true,
		"Time::Local":     true,
		"File::Basename":  true,
		"File::Path":      true,
		"IO::File":        true,
		"IO::Handle":      true,
		"Exporter":        true,
		"Test::More":      true,
		"Test::Simple":    true,
		"Digest::MD5":     true,
		"MIME::Base64":    true,
		"Storable":        true,
		"Socket":          true,
		"Fcntl":           true,
		"POSIX":           true,
		"Errno":           true,
		"constant":        true,
		"vars":            true,
		"base":            true,
		"parent":          true,
		"lib":             true,
		"utf8":            true,
		"feature":         true,
		"Config":          true,
		"DynaLoader":      true,
		"XSLoader":        true,
		"AutoLoader":      true,
		"SelfLoader":      true,
		"Benchmark":       true,
		"File::Find":      true,
		"File::Copy":      true,
		"File::Temp":      true,
		"File::Glob":      true,
		"Cwd":             true,
		"FindBin":         true,
		"Term::ReadLine":  true,
		"B":               true,
		"O":               true,
		"re":              true,
		"overload":        true,
		"attributes":      true,
		"fields":          true,
		"bytes":           true,
		"charnames":       true,
		"integer":         true,
		"less":            true,
		"locale":          true,
		"open":            true,
		"ops":             true,
		"sigtrap":         true,
		"sort":            true,
		"subs":            true,
		"threads":         true,
		"threads::shared": true,
		"version":         true,
		"vmsish":          true,
	}

	// Check if the module is in our basic list of core modules
	isCore, exists := coreModules[moduleName]
	if exists {
		return isCore, nil
	}

	// For modules not in our basic list, we can't determine core status
	// without additional API calls or Module::CoreList integration
	return false, nil
}

// GetPerlCoreVersionsWithDev retrieves available Perl core versions from MetaCPAN with optional development version support
func (p *MetaCPANProvider) GetPerlCoreVersionsWithDev(ctx context.Context, includeDev bool) ([]string, error) {
	if p.disableNetwork {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "network_disabled",
			Message: "Network access is disabled",
		}
	}

	// Check cache first
	if p.cacheDir != "" {
		cache, err := NewCache(p.cacheDir, p.cacheTTL)
		if err == nil {
			cacheKey := fmt.Sprintf("perl_core_versions_dev_%t", includeDev)
			var cachedVersions []string
			if cache.Get(cacheKey, &cachedVersions) {
				return cachedVersions, nil
			}
		}
	}

	// Query MetaCPAN for Perl core releases
	// Use the release endpoint to search for distributions named "perl"
	// Include maturity and authorized fields to filter out RCs and unauthorized releases
	endpoint := "/release/_search?q=distribution:perl&fields=version,date,maturity,authorized&size=100&sort=date:desc"
	requestURL := p.baseURL + endpoint

	// Make the request
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "request_creation_failed",
			Message: "Failed to create request",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Set headers
	req.Header.Set("User-Agent", p.userAgent)
	req.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "request_failed",
			Message: "Failed to execute request",
			URL:     requestURL,
			Err:     err,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, &ProviderError{
			Source:     p.Name(),
			Code:       "http_error",
			Message:    fmt.Sprintf("HTTP error: %d", resp.StatusCode),
			URL:        requestURL,
			StatusCode: resp.StatusCode,
		}
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "read_failed",
			Message: "Failed to read response body",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Parse the JSON response
	var searchResponse struct {
		Hits struct {
			Hits []struct {
				Fields struct {
					Version    string `json:"version"`
					Date       string `json:"date"`
					Maturity   string `json:"maturity"`
					Authorized bool   `json:"authorized"`
				} `json:"fields"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "parse_failed",
			Message: "Failed to parse JSON response",
			URL:     requestURL,
			Err:     err,
		}
	}

	// Extract versions and filter for stable releases
	versions := make([]string, 0)
	seen := make(map[string]bool)

	for _, hit := range searchResponse.Hits.Hits {
		version := hit.Fields.Version
		maturity := hit.Fields.Maturity
		authorized := hit.Fields.Authorized

		// Skip unauthorized releases entirely
		if !authorized {
			continue
		}

		// Skip release candidates and other developer releases
		// Only include releases with maturity "released" when not including dev versions
		if !includeDev && maturity != "released" {
			continue
		}

		// Convert MetaCPAN version format (5.042000) to standard format (5.42.0)
		standardVersion := convertMetaCPANVersion(version)
		// Filter versions based on includeDev parameter
		if standardVersion != "" && !seen[standardVersion] {
			// Include stable versions always, dev versions only if includeDev is true
			if includeDev || isStablePerlVersion(standardVersion) {
				seen[standardVersion] = true
				versions = append(versions, standardVersion)
			}
		}
	}

	// Save to cache
	if p.cacheDir != "" {
		cache, err := NewCache(p.cacheDir, p.cacheTTL)
		if err == nil {
			cacheKey := fmt.Sprintf("perl_core_versions_dev_%t", includeDev)
			_ = cache.Set(cacheKey, versions, p.Name())
		}
	}

	return versions, nil
}

// GetPerlCoreVersions retrieves available Perl core versions from MetaCPAN (stable versions only)
func (p *MetaCPANProvider) GetPerlCoreVersions(ctx context.Context) ([]string, error) {
	return p.GetPerlCoreVersionsWithDev(ctx, false)
}

// convertMetaCPANVersion converts MetaCPAN version format to standard format
// Examples: 5.042000 -> 5.42.0, 5.040001 -> 5.40.1, 5.038003 -> 5.38.3
func convertMetaCPANVersion(metacpanVersion string) string {
	// MetaCPAN uses format like "5.042000" for Perl versions
	// We need to convert this to "5.42.0" format

	// Match MetaCPAN version format: major.minorrevision where revision is 3 digits
	re := regexp.MustCompile(`^(\d+)\.(\d{3})(\d{3})$`)
	matches := re.FindStringSubmatch(metacpanVersion)

	if len(matches) != 4 {
		// If it doesn't match MetaCPAN format, return as-is
		return metacpanVersion
	}

	major := matches[1]
	minorStr := matches[2]
	revisionStr := matches[3]

	// Convert minor and revision to integers to remove leading zeros
	minor := 0
	revision := 0

	if _, err := fmt.Sscanf(minorStr, "%d", &minor); err != nil {
		return metacpanVersion
	}

	if _, err := fmt.Sscanf(revisionStr, "%d", &revision); err != nil {
		return metacpanVersion
	}

	// Return in standard format
	return fmt.Sprintf("%s.%d.%d", major, minor, revision)
}

// isStablePerlVersion determines if a version string represents a stable Perl release
func isStablePerlVersion(version string) bool {
	// Parse version string to check if it's a stable release
	// Stable versions have even minor version numbers (5.38.x, 5.36.x, etc.)
	// Development versions have odd minor version numbers (5.39.x, 5.37.x, etc.)

	// Simple regex to match version pattern
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
	matches := re.FindStringSubmatch(version)

	if len(matches) != 4 {
		return false
	}

	// Convert minor version to int
	minor := 0
	if _, err := fmt.Sscanf(matches[2], "%d", &minor); err != nil {
		return false
	}

	// Stable versions have even minor version numbers
	return minor%2 == 0
}

// calculateRating calculates a rating for a module based on CPAN Testers results
// This is a simplified rating calculation for demonstration purposes
func calculateRating(module *MetaCPANModule) float64 {
	total := module.StatsCPANTesters.Pass + module.StatsCPANTesters.Fail
	if total == 0 {
		return 0.0
	}
	return float64(module.StatsCPANTesters.Pass) / float64(total) * 5.0
}
