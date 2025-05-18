package cli

import (
	"os"
	"testing"
)

func TestDetectInvocation(t *testing.T) {
	// Save original args
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	
	// Create a test case with a direct invocation
	os.Args = []string{"/path/to/pvm"}
	info := DetectInvocation()
	
	if info.Component != ComponentPVM {
		t.Errorf("Expected component PVM, got %s", info.Component)
	}
	
	if info.Type != InvocationDirect {
		t.Errorf("Expected direct invocation, got %s", info.Type)
	}
	
	// Create a test case with an unknown binary name
	os.Args = []string{"/path/to/unknown"}
	info = DetectInvocation()
	
	if info.Component != ComponentPVM {
		t.Errorf("Expected fallback to PVM, got %s", info.Component)
	}
	
	if info.Type != InvocationFallback {
		t.Errorf("Expected fallback invocation, got %s", info.Type)
	}
	
	if info.Detected {
		t.Error("Expected detected flag to be false for unknown binary")
	}
}

func TestDebugInfo(t *testing.T) {
	// This function is mainly used for user output, so we just 
	// ensure it doesn't panic when called
	
	// Temporarily redirect stdout to avoid cluttering test output
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()
	
	os.Stdout, _ = os.Open(os.DevNull)
	
	// This should not panic
	PrintDebugInfo()
}

func TestGetWorkingDir(t *testing.T) {
	dir := getWorkingDir()
	
	// We expect this to return the current working directory
	// which should not be empty
	if dir == "" {
		t.Error("Expected non-empty working directory")
	}
	
	// Make sure it's not the error message
	if len(dir) >= 6 && dir[:6] == "Error:" {
		t.Errorf("Got error when getting working directory: %s", dir)
	}
}