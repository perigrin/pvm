package generation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryManager_CreateSession(t *testing.T) {
	mm := NewMemoryManager(50)
	defer mm.Close()

	sessionID := "test-session-1"
	memory := mm.CreateSession(sessionID)

	assert.NotNil(t, memory)
	assert.Equal(t, sessionID, memory.sessionID)
	assert.Equal(t, 50, memory.maxSize)
	assert.WithinDuration(t, time.Now(), memory.createdAt, time.Second)
	assert.WithinDuration(t, time.Now(), memory.lastAccessed, time.Second)
	assert.Empty(t, memory.typeChoices)
	assert.Empty(t, memory.namingPatterns)
	assert.Empty(t, memory.decisions)
}

func TestMemoryManager_GetSession(t *testing.T) {
	mm := NewMemoryManager(50)
	defer mm.Close()

	sessionID := "test-session-2"

	// First call should create a new session
	memory1 := mm.GetSession(sessionID)
	assert.NotNil(t, memory1)
	assert.Equal(t, sessionID, memory1.sessionID)

	// Second call should return the same session
	memory2 := mm.GetSession(sessionID)
	assert.Equal(t, memory1, memory2)
	assert.Equal(t, sessionID, memory2.sessionID)
}

func TestMemoryManager_ClearSession(t *testing.T) {
	mm := NewMemoryManager(50)
	defer mm.Close()

	sessionID := "test-session-3"
	memory := mm.CreateSession(sessionID)

	// Add some data
	memory.SetTypeChoice("var1", "Int")
	memory.SetNamingPattern("function", "camelCase")

	assert.Len(t, memory.typeChoices, 1)
	assert.Len(t, memory.namingPatterns, 1)
	assert.Len(t, memory.decisions, 2)

	// Clear the session
	mm.ClearSession(sessionID)

	// Session should be removed
	mm.mux.RLock()
	_, exists := mm.sessions[sessionID]
	mm.mux.RUnlock()
	assert.False(t, exists)
}

func TestGenerationMemory_TypeChoices(t *testing.T) {
	mm := NewMemoryManager(50)
	defer mm.Close()

	memory := mm.CreateSession("test-types")

	// Test setting and getting type choices
	memory.SetTypeChoice("variable1", "Int")
	memory.SetTypeChoice("function1", "String -> Int")

	typeStr, exists := memory.GetTypeChoice("variable1")
	assert.True(t, exists)
	assert.Equal(t, "Int", typeStr)

	typeStr, exists = memory.GetTypeChoice("function1")
	assert.True(t, exists)
	assert.Equal(t, "String -> Int", typeStr)

	// Test non-existent choice
	typeStr, exists = memory.GetTypeChoice("nonexistent")
	assert.False(t, exists)
	assert.Empty(t, typeStr)

	// Verify decisions were recorded
	decisions := memory.GetDecisions()
	assert.Len(t, decisions, 2)
	assert.Equal(t, "type_choice", decisions[0].Type)
	assert.Equal(t, "variable1", decisions[0].Context)
	assert.Equal(t, "Int", decisions[0].Choice)
}

func TestGenerationMemory_NamingPatterns(t *testing.T) {
	mm := NewMemoryManager(50)
	defer mm.Close()

	memory := mm.CreateSession("test-naming")

	// Test setting and getting naming patterns
	memory.SetNamingPattern("function", "snake_case")
	memory.SetNamingPattern("variable", "camelCase")

	pattern, exists := memory.GetNamingPattern("function")
	assert.True(t, exists)
	assert.Equal(t, "snake_case", pattern)

	pattern, exists = memory.GetNamingPattern("variable")
	assert.True(t, exists)
	assert.Equal(t, "camelCase", pattern)

	// Test non-existent pattern
	pattern, exists = memory.GetNamingPattern("nonexistent")
	assert.False(t, exists)
	assert.Empty(t, pattern)

	// Verify decisions were recorded
	decisions := memory.GetDecisions()
	assert.Len(t, decisions, 2)
	assert.Equal(t, "naming", decisions[0].Type)
	assert.Equal(t, "function", decisions[0].Context)
	assert.Equal(t, "snake_case", decisions[0].Choice)
}

func TestGenerationMemory_Decisions(t *testing.T) {
	mm := NewMemoryManager(50)
	defer mm.Close()

	memory := mm.CreateSession("test-decisions")

	// Add custom decisions
	memory.AddDecision("pattern", "error_handling", "use_exceptions", "Better error propagation")
	memory.AddDecision("refactor", "function_split", "extract_helper", "Reduce complexity")

	decisions := memory.GetDecisions()
	assert.Len(t, decisions, 2)

	assert.Equal(t, "pattern", decisions[0].Type)
	assert.Equal(t, "error_handling", decisions[0].Context)
	assert.Equal(t, "use_exceptions", decisions[0].Choice)
	assert.Equal(t, "Better error propagation", decisions[0].Rationale)
	assert.WithinDuration(t, time.Now(), decisions[0].Timestamp, time.Second)

	assert.Equal(t, "refactor", decisions[1].Type)
	assert.Equal(t, "function_split", decisions[1].Context)
	assert.Equal(t, "extract_helper", decisions[1].Choice)
	assert.Equal(t, "Reduce complexity", decisions[1].Rationale)
}

func TestGenerationMemory_RecentDecisions(t *testing.T) {
	mm := NewMemoryManager(50)
	defer mm.Close()

	memory := mm.CreateSession("test-recent")

	// Add decisions with different timestamps
	memory.AddDecision("old", "context1", "choice1", "reason1")

	// Manually set an old timestamp for the first decision
	memory.mux.Lock()
	memory.decisions[0].Timestamp = time.Now().Add(-10 * time.Minute)
	memory.mux.Unlock()

	memory.AddDecision("recent", "context2", "choice2", "reason2")

	// Get decisions from last 5 minutes
	recentDecisions := memory.GetRecentDecisions(5)
	assert.Len(t, recentDecisions, 1)
	assert.Equal(t, "recent", recentDecisions[0].Type)

	// Get decisions from last 15 minutes
	allRecentDecisions := memory.GetRecentDecisions(15)
	assert.Len(t, allRecentDecisions, 2)
}

func TestGenerationMemory_Context(t *testing.T) {
	mm := NewMemoryManager(50)
	defer mm.Close()

	sessionID := "test-context"
	memory := mm.CreateSession(sessionID)

	// Add some data
	memory.SetTypeChoice("var1", "Int")
	memory.SetNamingPattern("func", "snake_case")
	memory.AddDecision("test", "context", "choice", "reason")

	context := memory.GetContext()

	assert.Equal(t, sessionID, context["session_id"])
	assert.NotNil(t, context["created_at"])
	assert.NotNil(t, context["last_accessed"])
	assert.Equal(t, map[string]string{"var1": "Int"}, context["type_choices"])
	assert.Equal(t, map[string]string{"func": "snake_case"}, context["naming_patterns"])
	assert.Equal(t, 3, context["decision_count"])
	assert.Equal(t, 50, context["max_size"])
	assert.NotZero(t, context["memory_usage"])
}

func TestGenerationMemory_SessionStats(t *testing.T) {
	mm := NewMemoryManager(50)
	defer mm.Close()

	sessionID := "test-stats"
	memory := mm.CreateSession(sessionID)

	// Add some data
	memory.SetTypeChoice("var1", "Int")
	memory.SetTypeChoice("var2", "String")
	memory.SetNamingPattern("func", "camelCase")
	memory.AddDecision("test", "context", "choice", "reason")

	stats := memory.SessionStats()

	assert.Equal(t, sessionID, stats["session_id"])
	assert.IsType(t, float64(0), stats["duration_minutes"])
	assert.IsType(t, float64(0), stats["idle_minutes"])
	assert.Equal(t, 4, stats["decisions_count"]) // 2 type + 1 naming + 1 custom
	assert.Equal(t, 2, stats["type_choices"])
	assert.Equal(t, 1, stats["naming_patterns"])
	assert.Equal(t, 50, stats["memory_limit"])
	assert.NotZero(t, stats["memory_usage"])
	assert.Contains(t, stats["memory_utilization"], "%")
}

func TestGenerationMemory_Clear(t *testing.T) {
	mm := NewMemoryManager(50)
	defer mm.Close()

	memory := mm.CreateSession("test-clear")

	// Add data
	memory.SetTypeChoice("var1", "Int")
	memory.SetNamingPattern("func", "snake_case")
	memory.AddDecision("test", "context", "choice", "reason")

	assert.Len(t, memory.typeChoices, 1)
	assert.Len(t, memory.namingPatterns, 1)
	assert.Len(t, memory.decisions, 3)

	// Clear memory
	memory.Clear()

	assert.Empty(t, memory.typeChoices)
	assert.Empty(t, memory.namingPatterns)
	assert.Empty(t, memory.decisions)
}

func TestGenerationMemory_SizeEnforcement(t *testing.T) {
	mm := NewMemoryManager(5) // Very small limit for testing
	defer mm.Close()

	memory := mm.CreateSession("test-size")

	// Add more decisions than the limit
	for i := 0; i < 10; i++ {
		memory.AddDecision("test", "context", "choice", "reason")
	}

	// Should not exceed the limit
	decisions := memory.GetDecisions()
	assert.LessOrEqual(t, len(decisions), 5)
	assert.LessOrEqual(t, memory.calculateSize(), 5)
}

func TestMemoryManager_Cleanup(t *testing.T) {
	mm := NewMemoryManager(50)
	defer mm.Close()

	// Create a session
	memory := mm.CreateSession("test-cleanup")

	// Manually set old last accessed time
	memory.mux.Lock()
	memory.lastAccessed = time.Now().Add(-45 * time.Minute)
	memory.mux.Unlock()

	// Verify session exists
	mm.mux.RLock()
	_, exists := mm.sessions["test-cleanup"]
	mm.mux.RUnlock()
	assert.True(t, exists)

	// Trigger cleanup manually (simulate timer tick)
	mm.mux.Lock()
	cutoff := time.Now().Add(-30 * time.Minute)
	for sessionID, mem := range mm.sessions {
		mem.mux.RLock()
		lastAccessed := mem.lastAccessed
		mem.mux.RUnlock()

		if lastAccessed.Before(cutoff) {
			mem.Clear()
			delete(mm.sessions, sessionID)
		}
	}
	mm.mux.Unlock()

	// Session should be cleaned up
	mm.mux.RLock()
	_, exists = mm.sessions["test-cleanup"]
	mm.mux.RUnlock()
	assert.False(t, exists)
}

func TestGenerationMemory_ConcurrentAccess(t *testing.T) {
	mm := NewMemoryManager(100)
	defer mm.Close()

	memory := mm.CreateSession("test-concurrent")

	// Test concurrent reads and writes
	done := make(chan bool, 2)

	// Writer goroutine
	go func() {
		for i := 0; i < 50; i++ {
			memory.SetTypeChoice("var", "Int")
			memory.SetNamingPattern("func", "camelCase")
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 50; i++ {
			memory.GetTypeChoice("var")
			memory.GetNamingPattern("func")
			memory.GetDecisions()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Verify data integrity
	typeChoice, exists := memory.GetTypeChoice("var")
	assert.True(t, exists)
	assert.Equal(t, "Int", typeChoice)

	pattern, exists := memory.GetNamingPattern("func")
	assert.True(t, exists)
	assert.Equal(t, "camelCase", pattern)
}

func TestMemoryManager_MultipleSessionIsolation(t *testing.T) {
	mm := NewMemoryManager(50)
	defer mm.Close()

	// Create multiple sessions
	mem1 := mm.CreateSession("session-1")
	mem2 := mm.CreateSession("session-2")

	// Add different data to each session
	mem1.SetTypeChoice("var1", "Int")
	mem1.SetNamingPattern("func1", "snake_case")

	mem2.SetTypeChoice("var2", "String")
	mem2.SetNamingPattern("func2", "camelCase")

	// Verify isolation
	typeChoice, exists := mem1.GetTypeChoice("var1")
	assert.True(t, exists)
	assert.Equal(t, "Int", typeChoice)

	_, exists = mem1.GetTypeChoice("var2")
	assert.False(t, exists)

	typeChoice, exists = mem2.GetTypeChoice("var2")
	assert.True(t, exists)
	assert.Equal(t, "String", typeChoice)

	_, exists = mem2.GetTypeChoice("var1")
	assert.False(t, exists)

	// Verify decision isolation
	decisions1 := mem1.GetDecisions()
	decisions2 := mem2.GetDecisions()

	assert.Len(t, decisions1, 2) // type + naming
	assert.Len(t, decisions2, 2) // type + naming

	// Decisions should be different
	assert.Equal(t, "var1", decisions1[0].Context)
	assert.Equal(t, "var2", decisions2[0].Context)
}
