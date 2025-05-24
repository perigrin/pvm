// ABOUTME: Memory system for maintaining context during code generation tasks
// ABOUTME: Provides scoped memory storage for types, naming conventions, and generation decisions

package generation

import (
	"fmt"
	"sync"
	"time"
)

// Decision represents a generation decision made during the session
type Decision struct {
	Type      string    `json:"type"` // "type_choice", "naming", "pattern", "refactor"
	Context   string    `json:"context"`
	Choice    string    `json:"choice"`
	Rationale string    `json:"rationale"`
	Timestamp time.Time `json:"timestamp"`
}

// GenerationMemory maintains context within a single generation task
type GenerationMemory struct {
	sessionID      string
	typeChoices    map[string]string // variable/function name -> type choice
	namingPatterns map[string]string // pattern type -> naming convention
	decisions      []Decision        // ordered list of decisions made
	createdAt      time.Time
	lastAccessed   time.Time
	maxSize        int
	mux            sync.RWMutex
}

// MemoryManager manages generation memory sessions
type MemoryManager struct {
	sessions       map[string]*GenerationMemory
	defaultMaxSize int
	cleanupTicker  *time.Ticker
	mux            sync.RWMutex
}

// NewMemoryManager creates a new memory manager with cleanup
func NewMemoryManager(defaultMaxSize int) *MemoryManager {
	mm := &MemoryManager{
		sessions:       make(map[string]*GenerationMemory),
		defaultMaxSize: defaultMaxSize,
		cleanupTicker:  time.NewTicker(5 * time.Minute), // Cleanup every 5 minutes
	}

	// Start cleanup goroutine
	go mm.cleanup()

	return mm
}

// CreateSession creates a new generation memory session
func (mm *MemoryManager) CreateSession(sessionID string) *GenerationMemory {
	mm.mux.Lock()
	defer mm.mux.Unlock()

	// Clean up any existing session with same ID
	if existing, exists := mm.sessions[sessionID]; exists {
		existing.Clear()
	}

	memory := &GenerationMemory{
		sessionID:      sessionID,
		typeChoices:    make(map[string]string),
		namingPatterns: make(map[string]string),
		decisions:      make([]Decision, 0),
		createdAt:      time.Now(),
		lastAccessed:   time.Now(),
		maxSize:        mm.defaultMaxSize,
	}

	mm.sessions[sessionID] = memory
	return memory
}

// GetSession retrieves an existing session or creates a new one
func (mm *MemoryManager) GetSession(sessionID string) *GenerationMemory {
	mm.mux.RLock()
	memory, exists := mm.sessions[sessionID]
	mm.mux.RUnlock()

	if !exists {
		return mm.CreateSession(sessionID)
	}

	memory.mux.Lock()
	memory.lastAccessed = time.Now()
	memory.mux.Unlock()

	return memory
}

// ClearSession removes a session and clears its memory
func (mm *MemoryManager) ClearSession(sessionID string) {
	mm.mux.Lock()
	defer mm.mux.Unlock()

	if memory, exists := mm.sessions[sessionID]; exists {
		memory.Clear()
		delete(mm.sessions, sessionID)
	}
}

// Close stops the cleanup ticker and clears all sessions
func (mm *MemoryManager) Close() {
	mm.cleanupTicker.Stop()

	mm.mux.Lock()
	defer mm.mux.Unlock()

	for _, memory := range mm.sessions {
		memory.Clear()
	}
	mm.sessions = make(map[string]*GenerationMemory)
}

// cleanup removes sessions that haven't been accessed in 30 minutes
func (mm *MemoryManager) cleanup() {
	for range mm.cleanupTicker.C {
		mm.mux.Lock()
		cutoff := time.Now().Add(-30 * time.Minute)

		for sessionID, memory := range mm.sessions {
			memory.mux.RLock()
			lastAccessed := memory.lastAccessed
			memory.mux.RUnlock()

			if lastAccessed.Before(cutoff) {
				memory.Clear()
				delete(mm.sessions, sessionID)
			}
		}
		mm.mux.Unlock()
	}
}

// SetTypeChoice records a type choice for a variable or function
func (gm *GenerationMemory) SetTypeChoice(name, typeStr string) {
	gm.mux.Lock()
	defer gm.mux.Unlock()

	gm.typeChoices[name] = typeStr
	gm.addDecision(Decision{
		Type:      "type_choice",
		Context:   name,
		Choice:    typeStr,
		Timestamp: time.Now(),
	})
	gm.lastAccessed = time.Now()
	gm.enforceSize()
}

// GetTypeChoice retrieves the type choice for a variable or function
func (gm *GenerationMemory) GetTypeChoice(name string) (string, bool) {
	gm.mux.RLock()
	defer gm.mux.RUnlock()

	gm.lastAccessed = time.Now()
	typeStr, exists := gm.typeChoices[name]
	return typeStr, exists
}

// SetNamingPattern records a naming convention for a pattern type
func (gm *GenerationMemory) SetNamingPattern(patternType, convention string) {
	gm.mux.Lock()
	defer gm.mux.Unlock()

	gm.namingPatterns[patternType] = convention
	gm.addDecision(Decision{
		Type:      "naming",
		Context:   patternType,
		Choice:    convention,
		Timestamp: time.Now(),
	})
	gm.lastAccessed = time.Now()
	gm.enforceSize()
}

// GetNamingPattern retrieves the naming convention for a pattern type
func (gm *GenerationMemory) GetNamingPattern(patternType string) (string, bool) {
	gm.mux.RLock()
	defer gm.mux.RUnlock()

	gm.lastAccessed = time.Now()
	convention, exists := gm.namingPatterns[patternType]
	return convention, exists
}

// AddDecision records a generation decision with rationale
func (gm *GenerationMemory) AddDecision(decisionType, context, choice, rationale string) {
	gm.mux.Lock()
	defer gm.mux.Unlock()

	gm.addDecision(Decision{
		Type:      decisionType,
		Context:   context,
		Choice:    choice,
		Rationale: rationale,
		Timestamp: time.Now(),
	})
	gm.lastAccessed = time.Now()
	gm.enforceSize()
}

// GetDecisions returns all decisions made in chronological order
func (gm *GenerationMemory) GetDecisions() []Decision {
	gm.mux.RLock()
	defer gm.mux.RUnlock()

	gm.lastAccessed = time.Now()
	// Return a copy to prevent external modification
	decisions := make([]Decision, len(gm.decisions))
	copy(decisions, gm.decisions)
	return decisions
}

// GetRecentDecisions returns decisions from the last N minutes
func (gm *GenerationMemory) GetRecentDecisions(minutes int) []Decision {
	gm.mux.RLock()
	defer gm.mux.RUnlock()

	gm.lastAccessed = time.Now()
	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute)

	var recent []Decision
	for _, decision := range gm.decisions {
		if decision.Timestamp.After(cutoff) {
			recent = append(recent, decision)
		}
	}

	return recent
}

// GetContext returns a summary of current generation context
func (gm *GenerationMemory) GetContext() map[string]interface{} {
	gm.mux.RLock()
	defer gm.mux.RUnlock()

	gm.lastAccessed = time.Now()

	context := map[string]interface{}{
		"session_id":      gm.sessionID,
		"created_at":      gm.createdAt,
		"last_accessed":   gm.lastAccessed,
		"type_choices":    gm.typeChoices,
		"naming_patterns": gm.namingPatterns,
		"decision_count":  len(gm.decisions),
		"memory_usage":    gm.calculateSize(),
		"max_size":        gm.maxSize,
	}

	return context
}

// Clear removes all data from the memory
func (gm *GenerationMemory) Clear() {
	gm.mux.Lock()
	defer gm.mux.Unlock()

	gm.typeChoices = make(map[string]string)
	gm.namingPatterns = make(map[string]string)
	gm.decisions = make([]Decision, 0)
	gm.lastAccessed = time.Now()
}

// addDecision is an internal method to add decisions (caller must hold lock)
func (gm *GenerationMemory) addDecision(decision Decision) {
	gm.decisions = append(gm.decisions, decision)
}

// enforceSize keeps memory usage within bounds (caller must hold lock)
func (gm *GenerationMemory) enforceSize() {
	currentSize := gm.calculateSize()

	// If we're over the limit, remove the oldest decisions
	for currentSize > gm.maxSize && len(gm.decisions) > 0 {
		// Remove the oldest decision
		gm.decisions = gm.decisions[1:]
		currentSize = gm.calculateSize()
	}

	// Also limit type choices and naming patterns if they get too large
	if len(gm.typeChoices) > gm.maxSize/4 {
		// Keep only the most recent entries by rebuilding the map
		// This is a simple strategy - in production, you might want LRU
		newChoices := make(map[string]string)
		for i := len(gm.decisions) - 1; i >= 0 && len(newChoices) < gm.maxSize/4; i-- {
			if gm.decisions[i].Type == "type_choice" {
				newChoices[gm.decisions[i].Context] = gm.decisions[i].Choice
			}
		}
		gm.typeChoices = newChoices
	}

	if len(gm.namingPatterns) > gm.maxSize/4 {
		newPatterns := make(map[string]string)
		for i := len(gm.decisions) - 1; i >= 0 && len(newPatterns) < gm.maxSize/4; i-- {
			if gm.decisions[i].Type == "naming" {
				newPatterns[gm.decisions[i].Context] = gm.decisions[i].Choice
			}
		}
		gm.namingPatterns = newPatterns
	}
}

// calculateSize estimates memory usage (caller must hold lock)
func (gm *GenerationMemory) calculateSize() int {
	size := len(gm.decisions)
	size += len(gm.typeChoices)
	size += len(gm.namingPatterns)
	return size
}

// SessionStats provides statistics about a memory session
func (gm *GenerationMemory) SessionStats() map[string]interface{} {
	gm.mux.RLock()
	defer gm.mux.RUnlock()

	duration := time.Since(gm.createdAt)
	idleTime := time.Since(gm.lastAccessed)

	stats := map[string]interface{}{
		"session_id":         gm.sessionID,
		"duration_minutes":   duration.Minutes(),
		"idle_minutes":       idleTime.Minutes(),
		"decisions_count":    len(gm.decisions),
		"type_choices":       len(gm.typeChoices),
		"naming_patterns":    len(gm.namingPatterns),
		"memory_usage":       gm.calculateSize(),
		"memory_limit":       gm.maxSize,
		"memory_utilization": fmt.Sprintf("%.1f%%", float64(gm.calculateSize())/float64(gm.maxSize)*100),
	}

	return stats
}
