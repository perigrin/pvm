// ABOUTME: Data types for CPAN metadata
// ABOUTME: Defines structures for module metadata and search results

package cpan

import (
	"time"
)

// ModuleInfo contains metadata about a CPAN module
type ModuleInfo struct {
	// Name is the module name (e.g., "Moose")
	Name string

	// Version is the module version (e.g., "2.2014")
	Version string

	// Author is the PAUSE ID of the module author (e.g., "ETHER")
	Author string

	// AuthorName is the full name of the module author
	AuthorName string

	// Description is a short description of the module
	Description string

	// Abstract is a one-line summary of the module
	Abstract string

	// Distribution is the name of the distribution (e.g., "Moose")
	Distribution string

	// DistributionVersion is the version of the distribution
	DistributionVersion string

	// DistributionFile is the path to the distribution file on CPAN
	DistributionFile string

	// ReleaseDate is the date when the module was released
	ReleaseDate time.Time

	// Repository is the URL to the source code repository
	Repository string

	// Homepage is the URL to the module's homepage
	Homepage string

	// Bugtracker is the URL to the module's bug tracker
	Bugtracker string

	// License is the license under which the module is released
	License string

	// Dependencies is a list of module dependencies
	Dependencies []*Dependency

	// Documentation is the URL to the module's documentation
	Documentation string

	// Rating is the module's rating on CPAN
	Rating float64

	// Downloads is the number of downloads in the last month
	Downloads int

	// IsCore indicates whether the module is part of Perl core
	IsCore bool

	// HasTests indicates whether the module has tests
	HasTests bool

	// Path is the local file system path to the module (for locally installed modules)
	Path string
}

// Dependency represents a module dependency
type Dependency struct {
	// Name is the name of the required module
	Name string

	// Version is the required version constraint (e.g., ">= 2.0")
	Version string

	// Phase indicates when the dependency is needed (e.g., runtime, build, test)
	Phase string

	// Type indicates the type of dependency (e.g., requires, recommends)
	Type string

	// IsCore indicates whether the dependency is part of Perl core
	IsCore bool
}

// SearchResult represents a single result from a module search
type SearchResult struct {
	// Name is the module name (e.g., "Moose")
	Name string

	// Version is the module version (e.g., "2.2014")
	Version string

	// Distribution is the name of the distribution (e.g., "Moose")
	Distribution string

	// Author is the PAUSE ID of the module author (e.g., "ETHER")
	Author string

	// Abstract is a one-line summary of the module
	Abstract string

	// Score is the search relevance score
	Score float64

	// IsLatest indicates whether this is the latest version
	IsLatest bool

	// HasDocumentation indicates whether the module has documentation
	HasDocumentation bool

	// ReleaseDate is the date when the module was released
	ReleaseDate time.Time
}

// SearchResults represents results from a module search
type SearchResults struct {
	// Query is the search query that produced these results
	Query string

	// Total is the total number of matching results
	Total int

	// Results is the list of search results
	Results []*SearchResult

	// Source is the source of the search results (e.g., "metacpan")
	Source string
}

// ProviderError represents an error from a metadata provider
type ProviderError struct {
	// Source is the metadata source where the error occurred
	Source string

	// Code is the error code
	Code string

	// Message is the error message
	Message string

	// URL is the URL that was being accessed when the error occurred
	URL string

	// StatusCode is the HTTP status code (if applicable)
	StatusCode int

	// Err is the underlying error
	Err error
}

// Error implements the error interface
func (e *ProviderError) Error() string {
	if e.URL != "" {
		return e.Source + " error: " + e.Message + " (" + e.URL + ")"
	}
	return e.Source + " error: " + e.Message
}

// Unwrap returns the underlying error
func (e *ProviderError) Unwrap() error {
	return e.Err
}
