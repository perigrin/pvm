// ABOUTME: Tests for dependency management types
// ABOUTME: Validates type definitions and methods for dependency management

package dependencies

import (
	"testing"
	"time"
)

func TestConflictSeverity_String(t *testing.T) {
	tests := []struct {
		severity ConflictSeverity
		expected string
	}{
		{ConflictInfo, "info"},
		{ConflictWarning, "warning"},
		{ConflictError, "error"},
		{ConflictFatal, "fatal"},
		{ConflictSeverity(999), "unknown"},
	}

	for _, test := range tests {
		result := test.severity.String()
		if result != test.expected {
			t.Errorf("ConflictSeverity(%d).String() = %q, expected %q",
				test.severity, result, test.expected)
		}
	}
}

func TestResolutionAction_String(t *testing.T) {
	tests := []struct {
		action   ResolutionAction
		expected string
	}{
		{ActionUpgrade, "upgrade"},
		{ActionDowngrade, "downgrade"},
		{ActionRemove, "remove"},
		{ActionIgnore, "ignore"},
		{ActionManualResolve, "manual_resolve"},
		{ResolutionAction(999), "unknown"},
	}

	for _, test := range tests {
		result := test.action.String()
		if result != test.expected {
			t.Errorf("ResolutionAction(%d).String() = %q, expected %q",
				test.action, result, test.expected)
		}
	}
}

func TestValidationSeverity_String(t *testing.T) {
	tests := []struct {
		severity ValidationSeverity
		expected string
	}{
		{ValidationInfo, "info"},
		{ValidationWarning, "warning"},
		{ValidationErrorSeverity, "error"},
		{ValidationFatal, "fatal"},
		{ValidationSeverity(999), "unknown"},
	}

	for _, test := range tests {
		result := test.severity.String()
		if result != test.expected {
			t.Errorf("ValidationSeverity(%d).String() = %q, expected %q",
				test.severity, result, test.expected)
		}
	}
}

func TestRequirement_Basic(t *testing.T) {
	req := Requirement{
		Module:       "DBI",
		Version:      "1.643",
		Phase:        "runtime",
		Relationship: "requires",
		Optional:     false,
	}

	if req.Module != "DBI" {
		t.Errorf("Expected module name 'DBI', got %q", req.Module)
	}
	if req.Version != "1.643" {
		t.Errorf("Expected version '1.643', got %q", req.Version)
	}
	if req.Phase != "runtime" {
		t.Errorf("Expected phase 'runtime', got %q", req.Phase)
	}
}

func TestSnapshot_Basic(t *testing.T) {
	now := time.Now()
	snapshot := Snapshot{
		GeneratedAt: now,
		GeneratedBy: "pvm",
		PerlVersion: "5.36.0",
		Modules:     []*SnapshotModule{},
		Hash:        "abc123",
	}

	if snapshot.GeneratedBy != "pvm" {
		t.Errorf("Expected GeneratedBy 'pvm', got %q", snapshot.GeneratedBy)
	}
	if snapshot.PerlVersion != "5.36.0" {
		t.Errorf("Expected PerlVersion '5.36.0', got %q", snapshot.PerlVersion)
	}
	if len(snapshot.Modules) != 0 {
		t.Errorf("Expected empty modules slice, got %d modules", len(snapshot.Modules))
	}
}

func TestDependencyGraph_Basic(t *testing.T) {
	graph := DependencyGraph{
		Nodes:          make(map[string]*DependencyNode),
		Edges:          []*DependencyEdge{},
		RootModules:    []string{"DBI"},
		ResolutionTime: time.Now(),
	}

	if len(graph.Nodes) != 0 {
		t.Errorf("Expected empty nodes map, got %d nodes", len(graph.Nodes))
	}
	if len(graph.RootModules) != 1 {
		t.Errorf("Expected 1 root module, got %d", len(graph.RootModules))
	}
	if graph.RootModules[0] != "DBI" {
		t.Errorf("Expected root module 'DBI', got %q", graph.RootModules[0])
	}
}

func TestInstallPlan_Basic(t *testing.T) {
	plan := InstallPlan{
		Modules:           []*PlannedInstallation{},
		TotalModules:      0,
		EstimatedDuration: time.Minute * 5,
		Conflicts:         []*Conflict{},
		CreatedAt:         time.Now(),
		Valid:             true,
	}

	if !plan.Valid {
		t.Error("Expected plan to be valid")
	}
	if plan.TotalModules != 0 {
		t.Errorf("Expected 0 total modules, got %d", plan.TotalModules)
	}
	if plan.EstimatedDuration != time.Minute*5 {
		t.Errorf("Expected 5 minute duration, got %v", plan.EstimatedDuration)
	}
}
