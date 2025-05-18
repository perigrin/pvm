package version

import (
	"testing"
)

func TestGetVersion(t *testing.T) {
	if GetVersion() != Version {
		t.Errorf("GetVersion() = %s, want %s", GetVersion(), Version)
	}
}

func TestComponentVersion(t *testing.T) {
	component := "PVM"
	expected := "PVM " + Version
	if got := ComponentVersion(component); got != expected {
		t.Errorf("ComponentVersion(%s) = %s, want %s", component, got, expected)
	}
}
