// ABOUTME: MCP resources protocol handler for JSON-RPC resource requests
// ABOUTME: Implements resources/list, resources/read, and subscription methods

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"tamarou.com/pvm/internal/log"
)

// ResourcesHandler handles MCP resources protocol requests
type ResourcesHandler struct {
	manager *ResourceManager
	logger  *log.Logger
}

// ResourceListRequest represents a resources/list request
type ResourceListRequest struct {
	Cursor string          `json:"cursor,omitempty"`
	Filter *ResourceFilter `json:"filter,omitempty"`
}

// ResourceListResponse represents a resources/list response
type ResourceListResponse struct {
	Resources  []*Resource `json:"resources"`
	NextCursor string      `json:"nextCursor,omitempty"`
}

// ResourceReadRequest represents a resources/read request
type ResourceReadRequest struct {
	URI string `json:"uri"`
}

// ResourceReadResponse represents a resources/read response
type ResourceReadResponse struct {
	Contents []*ResourceContent `json:"contents"`
}

// ResourceSubscribeRequest represents a resources/subscribe request
type ResourceSubscribeRequest struct {
	URI string `json:"uri"`
}

// ResourceUnsubscribeRequest represents a resources/unsubscribe request
type ResourceUnsubscribeRequest struct {
	URI string `json:"uri"`
}

// ResourceUpdateNotification represents a resources/updated notification
type ResourceUpdateNotification struct {
	URI      string    `json:"uri"`
	Resource *Resource `json:"resource,omitempty"`
}

// NewResourcesHandler creates a new resources handler
func NewResourcesHandler(manager *ResourceManager) *ResourcesHandler {
	return &ResourcesHandler{
		manager: manager,
		logger:  log.NewLogger(log.LevelInfo, os.Stderr, "resources-handler"),
	}
}

// HandleListResources handles resources/list requests
func (h *ResourcesHandler) HandleListResources(ctx context.Context, params json.RawMessage) (interface{}, error) {
	if !h.manager.IsEnabled() {
		return nil, fmt.Errorf("resources are disabled")
	}

	var req ResourceListRequest
	if len(params) > 0 {
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid list resources request: %w", err)
		}
	}

	h.logger.Debugf("Handling resources/list request with filter: %v", req.Filter)

	resources, err := h.manager.ListResources(ctx, req.Filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	response := &ResourceListResponse{
		Resources: resources,
		// Note: Cursor-based pagination could be implemented here if needed
	}

	h.logger.Debugf("Resources/list completed with %d resources", len(resources))

	return response, nil
}

// HandleReadResource handles resources/read requests
func (h *ResourcesHandler) HandleReadResource(ctx context.Context, params json.RawMessage) (interface{}, error) {
	if !h.manager.IsEnabled() {
		return nil, fmt.Errorf("resources are disabled")
	}

	var req ResourceReadRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid read resource request: %w", err)
	}

	if req.URI == "" {
		return nil, fmt.Errorf("URI is required for read resource request")
	}

	h.logger.Debugf("Handling resources/read request for URI: %s", req.URI)

	content, err := h.manager.ReadResource(ctx, req.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource: %w", err)
	}

	response := &ResourceReadResponse{
		Contents: []*ResourceContent{content},
	}

	h.logger.Debugf("Resources/read completed for URI: %s, content_size: %d", req.URI, len(content.Text)+len(content.Blob))

	return response, nil
}

// HandleSubscribeToResource handles resources/subscribe requests
func (h *ResourcesHandler) HandleSubscribeToResource(ctx context.Context, params json.RawMessage) (interface{}, error) {
	if !h.manager.IsEnabled() {
		return nil, fmt.Errorf("resources are disabled")
	}

	var req ResourceSubscribeRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid subscribe request: %w", err)
	}

	if req.URI == "" {
		return nil, fmt.Errorf("URI is required for subscribe request")
	}

	h.logger.Debugf("Handling resources/subscribe request for URI: %s", req.URI)

	// Subscribe to resource changes
	eventChan, err := h.manager.SubscribeToChanges(ctx, req.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to resource: %w", err)
	}

	// Start listening for events and sending notifications
	// Note: In a real implementation, this would be handled by the MCP server's
	// notification system. For now, we just establish the subscription.
	go h.handleResourceEvents(req.URI, eventChan)

	h.logger.Debugf("Resources/subscribe completed for URI: %s", req.URI)

	// MCP subscribe returns an empty response
	return map[string]interface{}{}, nil
}

// HandleUnsubscribeFromResource handles resources/unsubscribe requests
func (h *ResourcesHandler) HandleUnsubscribeFromResource(ctx context.Context, params json.RawMessage) (interface{}, error) {
	if !h.manager.IsEnabled() {
		return nil, fmt.Errorf("resources are disabled")
	}

	var req ResourceUnsubscribeRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid unsubscribe request: %w", err)
	}

	if req.URI == "" {
		return nil, fmt.Errorf("URI is required for unsubscribe request")
	}

	h.logger.Debugf("Handling resources/unsubscribe request for URI: %s", req.URI)

	// Note: In a full implementation, we would need to track the specific
	// subscription channel and close it. For now, we'll just log the request.

	h.logger.Debugf("Resources/unsubscribe completed for URI: %s", req.URI)

	// MCP unsubscribe returns an empty response
	return map[string]interface{}{}, nil
}

// HandleRefreshResources handles a custom refresh resources request
func (h *ResourcesHandler) HandleRefreshResources(ctx context.Context) (interface{}, error) {
	if !h.manager.IsEnabled() {
		return nil, fmt.Errorf("resources are disabled")
	}

	h.logger.Debugf("Handling resources refresh request")

	err := h.manager.RefreshResources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh resources: %w", err)
	}

	// Get updated resource count
	resources, err := h.manager.ListResources(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource count after refresh: %w", err)
	}

	response := map[string]interface{}{
		"refreshed":      true,
		"resource_count": len(resources),
	}

	h.logger.Debugf("Resources refresh completed with %d resources", len(resources))

	return response, nil
}

// GetResourceStats returns statistics about the resources
func (h *ResourcesHandler) GetResourceStats(ctx context.Context) (map[string]interface{}, error) {
	if !h.manager.IsEnabled() {
		return map[string]interface{}{
			"enabled":        false,
			"resource_count": 0,
		}, nil
	}

	resources, err := h.manager.ListResources(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources for stats: %w", err)
	}

	// Calculate statistics
	stats := map[string]interface{}{
		"enabled":        true,
		"resource_count": len(resources),
	}

	// Count by type
	typeCounts := make(map[ResourceType]int)
	var totalSize int64

	for _, resource := range resources {
		typeCounts[resource.Type]++
		totalSize += resource.Size
	}

	stats["total_size_bytes"] = totalSize
	stats["types"] = map[string]interface{}{
		"files":          typeCounts[ResourceTypeFile],
		"directories":    typeCounts[ResourceTypeDirectory],
		"configurations": typeCounts[ResourceTypeConfiguration],
		"documentation":  typeCounts[ResourceTypeDocumentation],
		"tests":          typeCounts[ResourceTypeTest],
		"project_info":   typeCounts[ResourceTypeProjectInfo],
	}

	return stats, nil
}

// handleResourceEvents processes resource change events
func (h *ResourcesHandler) handleResourceEvents(uri string, eventChan <-chan ResourceEvent) {
	for event := range eventChan {
		h.logger.Debugf("Resource event received - uri: %s, type: %s, timestamp: %v",
			uri,
			event.Type,
			event.Timestamp)

		// Note: In a full MCP implementation, this would send a notification
		// to all connected MCP clients. For now, we just log the event.

		// The notification would be sent as:
		// {
		//   "jsonrpc": "2.0",
		//   "method": "notifications/resources/updated",
		//   "params": {
		//     "uri": event.URI,
		//     "resource": event.Resource
		//   }
		// }
	}
}

// IsEnabled returns whether the resources handler is enabled
func (h *ResourcesHandler) IsEnabled() bool {
	return h.manager.IsEnabled()
}

// GetSupportedMethods returns the list of MCP methods supported by this handler
func (h *ResourcesHandler) GetSupportedMethods() []string {
	return []string{
		"resources/list",
		"resources/read",
		"resources/subscribe",
		"resources/unsubscribe",
	}
}

// ValidateResourceURI validates that a resource URI is properly formatted
func (h *ResourcesHandler) ValidateResourceURI(uri string) error {
	if uri == "" {
		return fmt.Errorf("URI cannot be empty")
	}

	if !strings.HasPrefix(uri, "file://") {
		return fmt.Errorf("URI must start with file://")
	}

	return nil
}
