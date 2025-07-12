// ABOUTME: MCP resources protocol implementation for code context sharing
// ABOUTME: Provides project files, documentation, and code context to MCP servers

package mcp

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tamarou.com/pvm/internal/log"
)

// ResourceType represents the type of resource
type ResourceType string

const (
	ResourceTypeFile          ResourceType = "file"
	ResourceTypeDirectory     ResourceType = "directory"
	ResourceTypeProjectInfo   ResourceType = "project_info"
	ResourceTypeConfiguration ResourceType = "configuration"
	ResourceTypeDocumentation ResourceType = "documentation"
	ResourceTypeTest          ResourceType = "test"
)

// Resource represents a project resource that can be shared with MCP servers
type Resource struct {
	URI         string                 `json:"uri"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	Type        ResourceType           `json:"type"`
	Size        int64                  `json:"size,omitempty"`
	ModifiedAt  time.Time              `json:"modifiedAt,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ResourceContent represents the content of a resource
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
	Blob     []byte `json:"blob,omitempty"`
}

// ResourceManager manages MCP resources for a project
type ResourceManager struct {
	mu           sync.RWMutex
	projectPath  string
	resources    map[string]*Resource
	subscribers  map[string][]chan ResourceEvent
	logger       *log.Logger
	watchedPaths []string
	enabled      bool
}

// ResourceEvent represents a resource change event
type ResourceEvent struct {
	Type      string    `json:"type"` // created, updated, deleted
	URI       string    `json:"uri"`
	Resource  *Resource `json:"resource,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ResourceFilter represents criteria for filtering resources
type ResourceFilter struct {
	Types      []ResourceType `json:"types,omitempty"`
	Extensions []string       `json:"extensions,omitempty"`
	Patterns   []string       `json:"patterns,omitempty"`
	MaxSize    int64          `json:"maxSize,omitempty"`
}

// NewResourceManager creates a new resource manager
func NewResourceManager(projectPath string, enabled bool) *ResourceManager {
	return &ResourceManager{
		projectPath: projectPath,
		resources:   make(map[string]*Resource),
		subscribers: make(map[string][]chan ResourceEvent),
		logger:      log.NewLogger(log.LevelInfo, os.Stderr, "resource-manager"),
		enabled:     enabled,
	}
}

// Initialize scans the project and builds the resource index
func (rm *ResourceManager) Initialize(ctx context.Context) error {
	if !rm.enabled {
		return nil
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.logger.Infof("Initializing resource manager: %s", rm.projectPath)

	// Clear existing resources
	rm.resources = make(map[string]*Resource)

	// Scan project directory
	err := rm.scanDirectory(ctx, rm.projectPath)
	if err != nil {
		return fmt.Errorf("failed to scan project directory: %w", err)
	}

	rm.logger.Infof("Resource manager initialized - project_path: %s, resource_count: %d",
		rm.projectPath,
		len(rm.resources))

	return nil
}

// ListResources returns a list of available resources
func (rm *ResourceManager) ListResources(ctx context.Context, filter *ResourceFilter) ([]*Resource, error) {
	if !rm.enabled {
		return []*Resource{}, nil
	}

	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var result []*Resource
	for _, resource := range rm.resources {
		if rm.matchesFilter(resource, filter) {
			result = append(result, resource)
		}
	}

	return result, nil
}

// ReadResource returns the content of a specific resource
func (rm *ResourceManager) ReadResource(ctx context.Context, uri string) (*ResourceContent, error) {
	if !rm.enabled {
		return nil, fmt.Errorf("resource manager is disabled")
	}

	rm.mu.RLock()
	_, exists := rm.resources[uri]
	rm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("resource not found: %s", uri)
	}

	// Convert URI back to file path
	filePath := rm.uriToPath(uri)

	// Check if file still exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("resource file does not exist: %s", filePath)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource: %w", err)
	}

	// Determine if content is text or binary
	mimeType := rm.getMimeType(filePath)
	isText := rm.isTextFile(mimeType, content)

	resourceContent := &ResourceContent{
		URI:      uri,
		MimeType: mimeType,
	}

	if isText {
		resourceContent.Text = string(content)
	} else {
		resourceContent.Blob = content
	}

	return resourceContent, nil
}

// SubscribeToChanges subscribes to resource change events
func (rm *ResourceManager) SubscribeToChanges(ctx context.Context, uri string) (<-chan ResourceEvent, error) {
	if !rm.enabled {
		return nil, fmt.Errorf("resource manager is disabled")
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	eventChan := make(chan ResourceEvent, 10)
	rm.subscribers[uri] = append(rm.subscribers[uri], eventChan)

	rm.logger.Debugf("Subscribed to resource changes: %s", uri)

	// Return read-only channel
	return eventChan, nil
}

// UnsubscribeFromChanges unsubscribes from resource change events
func (rm *ResourceManager) UnsubscribeFromChanges(uri string, eventChan <-chan ResourceEvent) error {
	if !rm.enabled {
		return nil
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	subscribers := rm.subscribers[uri]
	for i, ch := range subscribers {
		if ch == eventChan {
			// Remove from slice
			rm.subscribers[uri] = append(subscribers[:i], subscribers[i+1:]...)
			close(ch)
			rm.logger.Debugf("Unsubscribed from resource changes: %s", uri)
			return nil
		}
	}

	return fmt.Errorf("subscription not found")
}

// GetResourceInfo returns information about a specific resource
func (rm *ResourceManager) GetResourceInfo(uri string) (*Resource, error) {
	if !rm.enabled {
		return nil, fmt.Errorf("resource manager is disabled")
	}

	rm.mu.RLock()
	defer rm.mu.RUnlock()

	resource, exists := rm.resources[uri]
	if !exists {
		return nil, fmt.Errorf("resource not found: %s", uri)
	}

	// Return a copy to avoid race conditions
	resourceCopy := *resource
	return &resourceCopy, nil
}

// RefreshResources rescans the project for changes
func (rm *ResourceManager) RefreshResources(ctx context.Context) error {
	if !rm.enabled {
		return nil
	}

	return rm.Initialize(ctx)
}

// IsEnabled returns whether the resource manager is enabled
func (rm *ResourceManager) IsEnabled() bool {
	return rm.enabled
}

// SetEnabled enables or disables the resource manager
func (rm *ResourceManager) SetEnabled(enabled bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.enabled = enabled
	if !enabled {
		// Clear resources and close all subscriptions
		rm.resources = make(map[string]*Resource)
		for uri, channels := range rm.subscribers {
			for _, ch := range channels {
				close(ch)
			}
			delete(rm.subscribers, uri)
		}
	}
}

// Private methods

func (rm *ResourceManager) scanDirectory(ctx context.Context, dirPath string) error {
	return filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			rm.logger.Warningf("Error accessing path during scan - path: %s, error: %v", path, err)
			return nil // Continue scanning
		}

		// Skip hidden directories and files
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip common ignore patterns
		if rm.shouldIgnore(path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Create resource entry
		resource, err := rm.createResource(path, d)
		if err != nil {
			rm.logger.Warningf("Failed to create resource - path: %s, error: %v", path, err)
			return nil
		}

		if resource != nil {
			uri := rm.pathToURI(path)
			rm.resources[uri] = resource
		}

		return nil
	})
}

func (rm *ResourceManager) createResource(path string, d fs.DirEntry) (*Resource, error) {
	info, err := d.Info()
	if err != nil {
		return nil, err
	}

	relPath, err := filepath.Rel(rm.projectPath, path)
	if err != nil {
		return nil, err
	}

	uri := rm.pathToURI(path)
	resourceType := rm.determineResourceType(path, d)

	resource := &Resource{
		URI:         uri,
		Name:        d.Name(),
		Description: rm.generateDescription(path, resourceType),
		MimeType:    rm.getMimeType(path),
		Type:        resourceType,
		Size:        info.Size(),
		ModifiedAt:  info.ModTime(),
		Metadata: map[string]interface{}{
			"relative_path": relPath,
			"absolute_path": path,
			"is_directory":  d.IsDir(),
		},
	}

	return resource, nil
}

func (rm *ResourceManager) pathToURI(path string) string {
	relPath, err := filepath.Rel(rm.projectPath, path)
	if err != nil {
		// Fallback to absolute path
		return fmt.Sprintf("file://%s", filepath.ToSlash(path))
	}
	return fmt.Sprintf("file://%s", filepath.ToSlash(relPath))
}

func (rm *ResourceManager) uriToPath(uri string) string {
	// Remove file:// prefix
	path := strings.TrimPrefix(uri, "file://")

	// Convert to absolute path if relative
	if !filepath.IsAbs(path) {
		return filepath.Join(rm.projectPath, path)
	}

	return path
}

func (rm *ResourceManager) determineResourceType(path string, d fs.DirEntry) ResourceType {
	if d.IsDir() {
		return ResourceTypeDirectory
	}

	ext := strings.ToLower(filepath.Ext(path))
	base := strings.ToLower(filepath.Base(path))

	switch {
	case ext == ".pl" || ext == ".pm" || ext == ".t":
		if strings.Contains(path, "t/") || strings.Contains(path, "test") {
			return ResourceTypeTest
		}
		return ResourceTypeFile
	case base == "cpanfile" || base == "meta.json" || base == "meta.yml" || ext == ".yml" || ext == ".yaml" || ext == ".toml":
		return ResourceTypeConfiguration
	case ext == ".md" || ext == ".pod" || base == "readme" || base == "changes" || base == "changelog":
		return ResourceTypeDocumentation
	case base == "project.toml" || base == "pvm.toml":
		return ResourceTypeProjectInfo
	default:
		return ResourceTypeFile
	}
}

func (rm *ResourceManager) generateDescription(path string, resourceType ResourceType) string {
	base := filepath.Base(path)

	switch resourceType {
	case ResourceTypeProjectInfo:
		return fmt.Sprintf("Project configuration file: %s", base)
	case ResourceTypeConfiguration:
		return fmt.Sprintf("Configuration file: %s", base)
	case ResourceTypeDocumentation:
		return fmt.Sprintf("Documentation file: %s", base)
	case ResourceTypeTest:
		return fmt.Sprintf("Test file: %s", base)
	case ResourceTypeDirectory:
		return fmt.Sprintf("Directory: %s", base)
	default:
		return fmt.Sprintf("Source file: %s", base)
	}
}

func (rm *ResourceManager) getMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	base := strings.ToLower(filepath.Base(path))

	switch {
	case ext == ".pl" || ext == ".pm" || ext == ".t":
		return "text/x-perl"
	case ext == ".md":
		return "text/markdown"
	case ext == ".pod":
		return "text/x-pod"
	case ext == ".json":
		return "application/json"
	case ext == ".yml" || ext == ".yaml":
		return "text/yaml"
	case ext == ".toml":
		return "text/toml"
	case base == "cpanfile":
		return "text/x-perl"
	case ext == ".txt" || strings.HasPrefix(base, "readme") || base == "changes" || base == "changelog":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

func (rm *ResourceManager) isTextFile(mimeType string, content []byte) bool {
	if strings.HasPrefix(mimeType, "text/") || strings.HasSuffix(mimeType, "/json") {
		return true
	}

	// Check if content appears to be text (simple heuristic)
	if len(content) == 0 {
		return true
	}

	// Check for null bytes (common in binary files)
	for i := 0; i < len(content) && i < 512; i++ {
		if content[i] == 0 {
			return false
		}
	}

	return true
}

func (rm *ResourceManager) shouldIgnore(path string) bool {
	base := filepath.Base(path)

	// Common ignore patterns
	ignorePatterns := []string{
		"node_modules",
		".git",
		".svn",
		".hg",
		"_build",
		"blib",
		".build",
		"cover_db",
		"nytprof",
		"nytprof.out",
		"MYMETA.json",
		"MYMETA.yml",
		".prove",
	}

	for _, pattern := range ignorePatterns {
		if base == pattern {
			return true
		}
	}

	return false
}

func (rm *ResourceManager) matchesFilter(resource *Resource, filter *ResourceFilter) bool {
	if filter == nil {
		return true
	}

	// Check types
	if len(filter.Types) > 0 {
		found := false
		for _, t := range filter.Types {
			if resource.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check extensions
	if len(filter.Extensions) > 0 {
		ext := strings.ToLower(filepath.Ext(resource.Name))
		found := false
		for _, e := range filter.Extensions {
			if ext == strings.ToLower(e) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check patterns (simple substring matching)
	if len(filter.Patterns) > 0 {
		found := false
		for _, pattern := range filter.Patterns {
			if strings.Contains(strings.ToLower(resource.Name), strings.ToLower(pattern)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check max size
	if filter.MaxSize > 0 && resource.Size > filter.MaxSize {
		return false
	}

	return true
}
