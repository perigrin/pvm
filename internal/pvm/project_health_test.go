// ABOUTME: Tests for project health check functionality
// ABOUTME: Validates health checks, doctor command, and auto-fix capabilities

package pvm

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/project"
)

func TestProjectDoctorCommand(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// Test 1: No project detected
	os.Chdir(tempDir)

	cmd := newProjectDoctorCommand()
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected command to succeed even with no project, got error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No project detected") {
		t.Errorf("Expected output to mention no project detected, got: %s", output)
	}

	// Test 2: Project with missing components
	// Create a basic project structure
	os.WriteFile(filepath.Join(tempDir, ".perl-version"), []byte("5.38.0\n"), 0644)

	buf.Reset()
	cmd = newProjectDoctorCommand()
	cmd.SetArgs([]string{})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Doctor command failed: %v", err)
	}

	output = buf.String()
	if !strings.Contains(output, "Health Check") {
		t.Errorf("Expected health check output, got: %s", output)
	}
}

func TestProjectDoctorJSON(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create a basic project
	os.WriteFile(filepath.Join(tempDir, ".perl-version"), []byte("5.38.0\n"), 0644)
	os.WriteFile(filepath.Join(tempDir, "cpanfile"), []byte("requires 'strict';\n"), 0644)

	cmd := newProjectDoctorCommand()
	cmd.SetArgs([]string{"--json"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Doctor command with JSON failed: %v", err)
	}

	// Validate JSON output
	var health ProjectHealth
	err = json.Unmarshal(buf.Bytes(), &health)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if len(health.Checks) == 0 {
		t.Error("Expected health checks in JSON output")
	}

	if health.OverallStatus == "" {
		t.Error("Expected overall status in JSON output")
	}
}

func TestProjectDoctorAutofix(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	cmd := newProjectDoctorCommand()
	cmd.SetArgs([]string{"--fix"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Doctor command with autofix failed: %v", err)
	}

	// Check if .perl-version was created
	if _, err := os.Stat(filepath.Join(tempDir, ".perl-version")); os.IsNotExist(err) {
		t.Error("Expected .perl-version file to be created by autofix")
	}
}

func TestHealthCheckStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   HealthStatus
		expected string
	}{
		{"healthy", HealthStatusHealthy, "\033[32m"},
		{"warning", HealthStatusWarning, "\033[33m"},
		{"critical", HealthStatusCritical, "\033[31m"},
		{"unknown", HealthStatusUnknown, "\033[37m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := getStatusColor(tt.status)
			if color != tt.expected {
				t.Errorf("Expected color %s for status %s, got %s", tt.expected, tt.status, color)
			}
		})
	}
}

func TestPerformHealthChecks(t *testing.T) {
	// Test with minimal project context
	ctx := &project.ProjectContext{
		IsProject:     true,
		RootDir:       "/tmp/test-project",
		DetectionInfo: "test",
		PerlVersion:   "5.38.0",
		HasCpanfile:   true,
		LocalLibDir:   "/tmp/test-project/lib",
		ConfigFile:    "/tmp/test-project/pvm.toml",
	}

	health := performHealthChecks(ctx, false)

	if health == nil {
		t.Fatal("Expected health result, got nil")
	}

	if len(health.Checks) == 0 {
		t.Error("Expected at least one health check")
	}

	if health.OverallStatus == "" {
		t.Error("Expected overall status to be set")
	}

	if health.Summary == "" {
		t.Error("Expected summary to be set")
	}

	if health.CheckedAt.IsZero() {
		t.Error("Expected CheckedAt timestamp to be set")
	}
}

func TestCheckProjectDetection(t *testing.T) {
	// Test project detected
	ctx := &project.ProjectContext{
		IsProject:     true,
		RootDir:       "/test/project",
		DetectionInfo: ".perl-version",
	}

	check := checkProjectDetection(ctx)
	if check.Status != HealthStatusHealthy {
		t.Errorf("Expected healthy status for detected project, got %s", check.Status)
	}

	// Test no project
	ctx.IsProject = false
	check = checkProjectDetection(ctx)
	if check.Status != HealthStatusCritical {
		t.Errorf("Expected critical status for no project, got %s", check.Status)
	}
}

func TestCheckPerlVersion(t *testing.T) {
	ctx := &project.ProjectContext{
		IsProject:   true,
		RootDir:     "/test/project",
		PerlVersion: "",
	}

	// Test missing .perl-version
	checks := checkPerlVersion(ctx, false)
	if len(checks) == 0 {
		t.Error("Expected at least one check for Perl version")
	}

	if checks[0].Status != HealthStatusWarning {
		t.Errorf("Expected warning for missing .perl-version, got %s", checks[0].Status)
	}

	// Test with Perl version specified
	ctx.PerlVersion = "5.38.0"
	checks = checkPerlVersion(ctx, false)

	// Should have checks for Perl installation or version consistency
	found := false
	for _, check := range checks {
		if strings.Contains(check.Name, "Consistency") || strings.Contains(check.Name, "Installation") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected version consistency or installation check when Perl version is specified")
	}
}

func TestCheckDependencies(t *testing.T) {
	ctx := &project.ProjectContext{
		IsProject:   true,
		RootDir:     "/test/project",
		HasCpanfile: false,
		LocalLibDir: "/test/project/lib",
	}

	// Test missing cpanfile
	checks := checkDependencies(ctx, false)
	if len(checks) == 0 {
		t.Error("Expected at least one dependency check")
	}

	if checks[0].Status != HealthStatusWarning {
		t.Errorf("Expected warning for missing cpanfile, got %s", checks[0].Status)
	}

	// Test with cpanfile
	ctx.HasCpanfile = true
	checks = checkDependencies(ctx, false)

	// Should check local lib existence
	found := false
	for _, check := range checks {
		if strings.Contains(check.Name, "Local Library") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected local library check when cpanfile exists")
	}
}

func TestCheckBuildSystem(t *testing.T) {
	ctx := &project.ProjectContext{
		IsProject: true,
		RootDir:   "/test/project",
	}

	checks := checkBuildSystem(ctx)
	if len(checks) == 0 {
		t.Error("Expected build system checks")
	}

	// Should check for build artifacts and PSC
	foundBuild := false
	foundPSC := false
	for _, check := range checks {
		if strings.Contains(check.Name, "Build Artifacts") {
			foundBuild = true
		}
		if strings.Contains(check.Name, "PSC") {
			foundPSC = true
		}
	}

	if !foundBuild {
		t.Error("Expected build artifacts check")
	}
	if !foundPSC {
		t.Error("Expected PSC check")
	}
}

func TestCheckConfiguration(t *testing.T) {
	ctx := &project.ProjectContext{
		IsProject:  true,
		RootDir:    "/test/project",
		ConfigFile: "",
	}

	// Test missing config
	checks := checkConfiguration(ctx, false)
	if len(checks) == 0 {
		t.Error("Expected configuration checks")
	}

	if checks[0].Status != HealthStatusWarning {
		t.Errorf("Expected warning for missing config, got %s", checks[0].Status)
	}

	// Test with config file
	ctx.ConfigFile = "/test/project/pvm.toml"
	checks = checkConfiguration(ctx, false)
	if checks[0].Status != HealthStatusHealthy {
		t.Errorf("Expected healthy status with config file, got %s", checks[0].Status)
	}
}

func TestCheckDevelopmentEnvironment(t *testing.T) {
	ctx := &project.ProjectContext{
		IsProject: true,
		RootDir:   "/test/project",
	}

	checks := checkDevelopmentEnvironment(ctx)
	if len(checks) == 0 {
		t.Error("Expected development environment checks")
	}

	// Should check for Git and .gitignore
	foundGit := false
	foundGitIgnore := false
	for _, check := range checks {
		if strings.Contains(check.Name, "Version Control") {
			foundGit = true
		}
		if strings.Contains(check.Name, "Git Ignore") {
			foundGitIgnore = true
		}
	}

	if !foundGit {
		t.Error("Expected Git repository check")
	}
	if !foundGitIgnore {
		t.Error("Expected .gitignore check")
	}
}

func TestEnhancedProjectStatus(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create a project with all components
	os.WriteFile(filepath.Join(tempDir, ".perl-version"), []byte("5.38.0\n"), 0644)
	os.WriteFile(filepath.Join(tempDir, "cpanfile"), []byte("requires 'strict';\n"), 0644)
	os.WriteFile(filepath.Join(tempDir, "pvm.toml"), []byte("[project]\nname = \"test\"\n"), 0644)
	os.MkdirAll(filepath.Join(tempDir, "lib"), 0755)

	cmd := newProjectStatusCommand()
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Status command failed: %v", err)
	}

	output := buf.String()
	// Check for project status content - the exact text may be styled with ANSI codes
	if !strings.Contains(output, "Project Root:") {
		t.Errorf("Expected project status output with 'Project Root:', got: %q", output)
	}
}

func TestProjectStatusJSON(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create a basic project
	os.WriteFile(filepath.Join(tempDir, ".perl-version"), []byte("5.38.0\n"), 0644)

	cmd := newProjectStatusCommand()
	cmd.SetArgs([]string{"--json"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Status command with JSON failed: %v", err)
	}

	// Validate JSON output
	var status map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &status)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if _, exists := status["is_project"]; !exists {
		t.Error("Expected is_project in JSON output")
	}

	if _, exists := status["timestamp"]; !exists {
		t.Error("Expected timestamp in JSON output")
	}
}

func TestHealthCheckStructures(t *testing.T) {
	// Test HealthCheck creation
	check := HealthCheck{
		Name:       "Test Check",
		Status:     HealthStatusHealthy,
		Message:    "All good",
		Details:    "No issues found",
		Suggestion: "Keep up the good work",
		CheckedAt:  time.Now(),
	}

	if check.Name != "Test Check" {
		t.Error("HealthCheck name not set correctly")
	}

	if check.Status != HealthStatusHealthy {
		t.Error("HealthCheck status not set correctly")
	}

	// Test ProjectHealth creation
	health := ProjectHealth{
		OverallStatus: HealthStatusWarning,
		Checks:        []HealthCheck{check},
		Summary:       "1 warning found",
		NextSteps:     []string{"Fix the warning"},
		CheckedAt:     time.Now(),
	}

	if health.OverallStatus != HealthStatusWarning {
		t.Error("ProjectHealth overall status not set correctly")
	}

	if len(health.Checks) != 1 {
		t.Error("ProjectHealth checks not set correctly")
	}

	if len(health.NextSteps) != 1 {
		t.Error("ProjectHealth next steps not set correctly")
	}
}

func TestHelperFunctions(t *testing.T) {
	tempDir := t.TempDir()

	// Test createPerlVersionFile
	err := createPerlVersionFile(tempDir)
	if err != nil {
		t.Fatalf("Failed to create .perl-version file: %v", err)
	}

	versionPath := filepath.Join(tempDir, ".perl-version")
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		t.Error("Expected .perl-version file to be created")
	}

	// Test countInstalledModules
	libDir := filepath.Join(tempDir, "lib", "perl5")
	os.MkdirAll(libDir, 0755)

	// Create some fake .pm files
	os.WriteFile(filepath.Join(libDir, "Test.pm"), []byte("package Test;"), 0644)
	os.MkdirAll(filepath.Join(libDir, "Test"), 0755)
	os.WriteFile(filepath.Join(libDir, "Test", "More.pm"), []byte("package Test::More;"), 0644)

	count := countInstalledModules(filepath.Join(tempDir, "lib"))
	if count != 2 {
		t.Errorf("Expected 2 installed modules, got %d", count)
	}

	// Test createDefaultConfig
	err = createDefaultConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to create default config: %v", err)
	}

	configPath := filepath.Join(tempDir, "pvm.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected pvm.toml to be created")
	}
}
