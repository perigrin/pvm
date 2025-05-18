package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogLevels(t *testing.T) {
	// Create a buffer to capture log output
	buf := new(bytes.Buffer)
	
	// Create a new logger
	logger := NewLogger(LevelDebug, buf, "TEST")
	
	// Test different log levels
	logger.Debugf("Debug message")
	logger.Infof("Info message")
	logger.Warningf("Warning message")
	logger.Errorf("Error message")
	
	// Check the output
	output := buf.String()
	
	// Verify each level of message was logged
	if !strings.Contains(output, "[DEBUG]") {
		t.Error("Expected debug message to be logged")
	}
	
	if !strings.Contains(output, "[INFO]") {
		t.Error("Expected info message to be logged")
	}
	
	if !strings.Contains(output, "[WARNING]") {
		t.Error("Expected warning message to be logged")
	}
	
	if !strings.Contains(output, "[ERROR]") {
		t.Error("Expected error message to be logged")
	}
	
	// Verify component is included
	if !strings.Contains(output, "[TEST]") {
		t.Error("Expected component to be included in log output")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	// Create a buffer to capture log output
	buf := new(bytes.Buffer)
	
	// Create a new logger with INFO level
	logger := NewLogger(LevelInfo, buf, "TEST")
	
	// Log messages at different levels
	logger.Debugf("Debug message") // Should not be logged
	logger.Infof("Info message")   // Should be logged
	
	// Check the output
	output := buf.String()
	
	// Verify debug message was not logged
	if strings.Contains(output, "Debug message") {
		t.Error("Expected debug message to be filtered out")
	}
	
	// Verify info message was logged
	if !strings.Contains(output, "Info message") {
		t.Error("Expected info message to be logged")
	}
}

func TestLoggerSetLevel(t *testing.T) {
	// Create a buffer to capture log output
	buf := new(bytes.Buffer)
	
	// Create a new logger with INFO level
	logger := NewLogger(LevelInfo, buf, "TEST")
	
	// Log a debug message (should not appear)
	logger.Debugf("Debug message 1")
	
	// Change level to DEBUG
	logger.SetLevel(LevelDebug)
	
	// Log another debug message (should appear)
	logger.Debugf("Debug message 2")
	
	// Check the output
	output := buf.String()
	
	// Verify first debug message was not logged
	if strings.Contains(output, "Debug message 1") {
		t.Error("Expected first debug message to be filtered out")
	}
	
	// Verify second debug message was logged
	if !strings.Contains(output, "Debug message 2") {
		t.Error("Expected second debug message to be logged")
	}
}

func TestQuietMode(t *testing.T) {
	// Create a buffer to capture log output
	buf := new(bytes.Buffer)
	
	// Create a new logger with INFO level
	logger := NewLogger(LevelInfo, buf, "TEST")
	
	// Log a message
	logger.Infof("Test message 1")
	
	// Enable quiet mode
	logger.SetQuiet(true)
	
	// Log another message
	logger.Infof("Test message 2")
	
	// Check the output
	output := buf.String()
	
	// Verify first message was logged
	if !strings.Contains(output, "Test message 1") {
		t.Error("Expected first message to be logged")
	}
	
	// Verify second message was not logged
	if strings.Contains(output, "Test message 2") {
		t.Error("Expected second message to be filtered out in quiet mode")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"debug", LevelDebug, false},
		{"DEBUG", LevelDebug, false},
		{"info", LevelInfo, false},
		{"INFO", LevelInfo, false},
		{"warning", LevelWarning, false},
		{"WARNING", LevelWarning, false},
		{"warn", LevelWarning, false},
		{"WARN", LevelWarning, false},
		{"error", LevelError, false},
		{"ERROR", LevelError, false},
		{"fatal", LevelFatal, false},
		{"FATAL", LevelFatal, false},
		{"unknown", LevelInfo, true},
	}
	
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			level, err := ParseLevel(test.input)
			
			if test.hasError && err == nil {
				t.Error("Expected error but got nil")
			}
			
			if !test.hasError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			if level != test.expected {
				t.Errorf("Expected level %d, got %d", test.expected, level)
			}
		})
	}
}