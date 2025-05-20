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
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
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

// MetaCPANModule represents a module in the MetaCPAN API
type MetaCPANModule struct {
	Name             string               `json:"name"`
	Version          string               `json:"version"`
	AuthorID         string               `json:"author"`
	Abstract         string               `json:"abstract"`
	Description      string               `json:"description"`
	Distribution     string               `json:"distribution"`
	DistVersion      string               `json:"dist_version"`
	DownloadURL      string               `json:"download_url"`
	ReleaseDate      string               `json:"date"`
	Status           string               `json:"status"`
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

	// Check cache first (not implemented yet)

	// Prepare the URL
	endpoint := fmt.Sprintf("/module/%s", url.PathEscape(moduleName))
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

	// Convert to a generic ModuleInfo
	info := &ModuleInfo{
		Name:                module.Name,
		Version:             module.Version,
		Author:              module.AuthorID,
		Description:         module.Description,
		Abstract:            module.Abstract,
		Distribution:        module.Distribution,
		DistributionVersion: module.DistVersion,
		Documentation:       fmt.Sprintf("%s/pod/%s", metaCPANWebBaseURL, module.Name),
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

	// Save to cache (not implemented yet)

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

	// Check cache first (not implemented yet)

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

	// Save to cache (not implemented yet)

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

	// This is a simplified implementation that would be expanded in a full implementation
	// to actually query the API for all versions
	return []string{}, errors.NewSystemError("101", "GetModuleVersions is not fully implemented yet", nil)
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

	// Check cache first (not implemented yet)

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

	// Save to cache (not implemented yet)

	return authorInfo, nil
}

// IsCoreModule checks if a module is part of the Perl core
func (p *MetaCPANProvider) IsCoreModule(ctx context.Context, moduleName string, perlVersion string) (bool, error) {
	// This is a simplified implementation that would be expanded in a full implementation
	// to actually check if the module is part of the core for the given Perl version
	return false, errors.NewSystemError("101", "IsCoreModule is not fully implemented yet", nil)
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
